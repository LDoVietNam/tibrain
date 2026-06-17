package opencode

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const pidFile = ".ti/opencode.pid"

// Sidecar manages a background `opencode serve` process.
type Sidecar struct {
	cfg Config
}

// NewSidecar creates a Sidecar from config.
func NewSidecar(cfg Config) *Sidecar {
	return &Sidecar{cfg: cfg}
}

// Start launches `opencode serve --port <port>` in the background.
// Writes PID to .ti/opencode.pid.
func (s *Sidecar) Start() error {
	if s.IsRunning() {
		return fmt.Errorf("opencode sidecar already running on port %d", s.cfg.Port)
	}

	if err := os.MkdirAll(filepath.Dir(pidFile), 0o755); err != nil {
		return fmt.Errorf("mkdir .ti: %w", err)
	}

	args := []string{"serve", "--port", strconv.Itoa(s.cfg.Port)}
	cmd := exec.Command(s.cfg.Path, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start opencode serve: %w", err)
	}

	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0o644); err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("write pid file: %w", err)
	}

	// Wait briefly for server to be ready.
	url := s.URL() + "/health"
	for i := 0; i < 20; i++ {
		time.Sleep(200 * time.Millisecond)
		resp, err := http.Get(url) //nolint:noctx
		if err == nil {
			resp.Body.Close()
			return nil
		}
	}
	return nil // server may still be starting; not fatal
}

// Stop kills the sidecar process identified by .ti/opencode.pid.
func (s *Sidecar) Stop() error {
	pid, err := s.readPID()
	if err != nil {
		return fmt.Errorf("no sidecar pid found: %w", err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		_ = os.Remove(pidFile)
		return fmt.Errorf("process %d not found: %w", pid, err)
	}

	if runtime.GOOS == "windows" {
		kill := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F")
		_ = kill.Run()
	} else {
		_ = proc.Kill()
	}

	_ = os.Remove(pidFile)
	return nil
}

// IsRunning returns true if the sidecar process is alive and the HTTP server responds.
func (s *Sidecar) IsRunning() bool {
	pid, err := s.readPID()
	if err != nil {
		return false
	}
	if !pidAlive(pid) {
		_ = os.Remove(pidFile)
		return false
	}
	// Quick HTTP ping
	resp, err := http.Get(s.URL() + "/health") //nolint:noctx
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

// URL returns the base URL of the sidecar server.
func (s *Sidecar) URL() string {
	return fmt.Sprintf("http://localhost:%d", s.cfg.Port)
}

// PID returns the sidecar PID, or 0 if not running.
func (s *Sidecar) PID() int {
	pid, _ := s.readPID()
	return pid
}

func (s *Sidecar) readPID() (int, error) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

// pidAlive checks if a process with the given PID is running.
func pidAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		// On Windows FindProcess always succeeds; use tasklist to verify.
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), strconv.Itoa(pid))
	}
	// Unix: send signal 0 to check existence.
	err = proc.Signal(os.Signal(nil))
	return err == nil
}
