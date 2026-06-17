package verify

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ti/cli/internal/permission"
)

type Result struct {
	Command      string        `json:"command"`
	Passed       bool          `json:"passed"`
	Output       string        `json:"output,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
	Duration     time.Duration `json:"duration"`
}

type Summary struct {
	Commands int `json:"commands"`
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
}

type Runner struct {
	Policy  *permission.Policy
	WorkDir string
	Timeout time.Duration
}

func NewRunner(workDir string, policy *permission.Policy, timeout time.Duration) *Runner {
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	return &Runner{Policy: policy, WorkDir: strings.TrimSpace(workDir), Timeout: timeout}
}

func (r *Runner) Run(commands []string) ([]Result, error) {
	commands = sanitizeCommands(commands)
	if len(commands) == 0 {
		return nil, fmt.Errorf("no verify commands available")
	}
	results := make([]Result, 0, len(commands))
	for _, cmdStr := range commands {
		start := time.Now()
		decision := permission.DefaultPolicy().Evaluate("bash", "", cmdStr)
		if r.Policy != nil {
			decision = r.Policy.Evaluate("bash", "", cmdStr)
		}
		if decision.Behavior == permission.BehaviorDeny {
			results = append(results, Result{Command: cmdStr, Passed: false, ErrorMessage: decision.Reason, Duration: time.Since(start)})
			return results, fmt.Errorf("verification denied by policy: %s", decision.Reason)
		}
		ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
		cmd := shellCommand(ctx, cmdStr)
		if r.WorkDir != "" {
			cmd.Dir = r.WorkDir
		}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		cancel()
		out := strings.TrimSpace(strings.TrimSpace(stdout.String() + "\n" + stderr.String()))
		res := Result{Command: cmdStr, Passed: err == nil, Output: out, Duration: time.Since(start)}
		if err != nil {
			res.ErrorMessage = err.Error()
			results = append(results, res)
			return results, fmt.Errorf("verify command failed: %s", cmdStr)
		}
		results = append(results, res)
	}
	return results, nil
}

func Summarize(results []Result) Summary {
	s := Summary{Commands: len(results)}
	for _, r := range results {
		if r.Passed {
			s.Passed++
		} else {
			s.Failed++
		}
	}
	return s
}

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-lc", command)
}

func sanitizeCommands(commands []string) []string {
	out := make([]string, 0, len(commands))
	seen := map[string]bool{}
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" || seen[cmd] {
			continue
		}
		seen[cmd] = true
		out = append(out, cmd)
	}
	return out
}
