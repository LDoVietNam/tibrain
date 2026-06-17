package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type RunRequest struct {
	Command      string
	CWD          string
	Backend      string
	Timeout      time.Duration
	AllowNetwork bool
	Policy       *Policy
}

type RunResult struct {
	ID          string
	Backend     string
	Command     string
	CWD         string
	ExitCode    int
	Duration    time.Duration
	Output      string
	OutputBytes int
	TimedOut    bool
	Blocked     bool
}

type Backend interface {
	Name() string
	Available(context.Context) bool
	Run(context.Context, RunRequest) (*RunResult, error)
}

func AvailableBackends(ctx context.Context) []string {
	backends := []Backend{PortableBackend{}, BubblewrapBackend{}, DockerBackend{}}
	names := []string{}
	for _, b := range backends {
		if b.Available(ctx) {
			names = append(names, b.Name())
		}
	}
	return names
}

func ResolveBackend(ctx context.Context, name string, fallback string) string {
	if name == "" || name == BackendAuto {
		if runtime.GOOS == "linux" && (BubblewrapBackend{}).Available(ctx) {
			return BackendBubblewrap
		}
		return BackendPortable
	}
	b := backendByName(name)
	if b != nil && b.Available(ctx) {
		return name
	}
	if fallback != "" && fallback != name {
		fb := backendByName(fallback)
		if fb != nil && fb.Available(ctx) {
			return fallback
		}
	}
	return BackendPortable
}

func Run(ctx context.Context, req RunRequest) (*RunResult, error) {
	p := req.Policy
	if p == nil {
		p = DefaultPolicy()
	}
	p.Normalize()
	if !p.Enabled {
		return nil, fmt.Errorf("sandbox policy is disabled")
	}
	if req.Timeout <= 0 {
		req.Timeout = p.Timeout()
	}
	if req.CWD == "" {
		cwd, _ := os.Getwd()
		req.CWD = cwd
	}
	if abs, err := filepath.Abs(req.CWD); err == nil {
		req.CWD = abs
	}
	if req.Backend == "" {
		req.Backend = p.Backend
	}
	if err := ValidateCommand(req.Command, p); err != nil {
		res := &RunResult{ID: newID(), Backend: req.Backend, Command: req.Command, CWD: req.CWD, Blocked: true, ExitCode: -1}
		_ = WriteAudit(p.AuditFile, AuditEvent{ID: res.ID, Time: time.Now(), Backend: res.Backend, Command: req.Command, CWD: req.CWD, Network: p.Network, Allowed: false, ExitCode: -1, Error: err.Error()})
		return res, err
	}
	backendName := ResolveBackend(ctx, req.Backend, p.Fallback)
	req.Backend = backendName
	req.Policy = p
	req.AllowNetwork = req.AllowNetwork || p.AllowNetwork()
	b := backendByName(backendName)
	if b == nil {
		b = PortableBackend{}
	}
	start := time.Now()
	res, err := b.Run(ctx, req)
	if res == nil {
		res = &RunResult{ID: newID(), Backend: backendName, Command: req.Command, CWD: req.CWD, ExitCode: -1}
	}
	res.Duration = time.Since(start)
	if res.ID == "" {
		res.ID = newID()
	}
	auditErr := ""
	if err != nil {
		auditErr = err.Error()
	}
	_ = WriteAudit(p.AuditFile, AuditEvent{
		ID: res.ID, Time: time.Now(), Backend: res.Backend, Command: req.Command, CWD: req.CWD,
		Network: p.Network, Allowed: err == nil, ExitCode: res.ExitCode, DurationMS: res.Duration.Milliseconds(), OutputBytes: res.OutputBytes, Error: auditErr,
	})
	return res, err
}

func backendByName(name string) Backend {
	switch strings.ToLower(name) {
	case BackendPortable, "":
		return PortableBackend{}
	case BackendBubblewrap, "bwrap":
		return BubblewrapBackend{}
	case BackendDocker:
		return DockerBackend{}
	default:
		return nil
	}
}

func runCommand(ctx context.Context, req RunRequest, argv []string, env []string) (*RunResult, error) {
	if len(argv) == 0 {
		return nil, errors.New("empty argv")
	}
	runCtx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, argv[0], argv[1:]...)
	cmd.Dir = req.CWD
	cmd.Env = env
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	timedOut := runCtx.Err() == context.DeadlineExceeded
	exitCode := 0
	if err != nil {
		exitCode = 1
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		}
	}
	out := buf.String()
	max := req.Policy.MaxOutputBytes
	out = truncate(out, max)
	if timedOut {
		out += fmt.Sprintf("\n[timeout: command exceeded %s]", req.Timeout)
	}
	return &RunResult{ID: newID(), Backend: req.Backend, Command: req.Command, CWD: req.CWD, ExitCode: exitCode, Output: out, OutputBytes: len(out), TimedOut: timedOut}, nil
}

func shellArgv(command string) []string {
	if runtime.GOOS == "windows" {
		return []string{"cmd", "/C", command}
	}
	return []string{"sh", "-c", command}
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	head := max * 6 / 10
	tail := max - head
	return s[:head] + fmt.Sprintf("\n\n[... %d bytes truncated by sandbox ...]\n\n", len(s)-head-tail) + s[len(s)-tail:]
}

func newID() string {
	return fmt.Sprintf("sbox-%d", time.Now().UnixNano())
}
