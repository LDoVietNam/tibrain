// Package tunnel manages cloudflared quick-tunnel processes.
//
// State is persisted to .ti/tunnel.json so `ti tunnel status` works
// across invocations without keeping a parent process alive.
package tunnel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// TunnelEntry represents one running tunnel.
type TunnelEntry struct {
	Name      string    `json:"name"`
	Port      int       `json:"port"`
	PID       int       `json:"pid"`
	URL       string    `json:"url"`
	LogFile   string    `json:"log_file"`
	StartedAt time.Time `json:"started_at"`
}

// state is the on-disk format.
type state struct {
	Tunnels map[string]*TunnelEntry `json:"tunnels"`
}

// Manager manages tunnel lifecycle.
type Manager struct {
	stateFile string
	logDir    string
}

// New creates a Manager rooted at stateDir (e.g. ".ti").
func New(stateDir string) *Manager {
	return &Manager{
		stateFile: filepath.Join(stateDir, "tunnel.json"),
		logDir:    stateDir,
	}
}

// Start spawns a cloudflared quick-tunnel for the given local port.
// It waits up to 15 s for the public URL to appear in the log, then returns it.
func (m *Manager) Start(name string, port int) (*TunnelEntry, error) {
	if name == "" {
		name = fmt.Sprintf("tunnel-%d", port)
	}

	// Check if already running
	st, _ := m.load()
	if e, ok := st.Tunnels[name]; ok && isAlive(e.PID) {
		return e, fmt.Errorf("tunnel %q already running (pid %d, url %s)", name, e.PID, e.URL)
	}

	logFile := filepath.Join(m.logDir, fmt.Sprintf("cloudflared-%s.log", name))
	_ = os.Remove(logFile) // clear old log

	// Build cloudflared command
	args := []string{
		"tunnel",
		"--url", fmt.Sprintf("http://localhost:%d", port),
		"--logfile", logFile,
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// On Windows, start detached so it survives the parent process
		cmd = exec.Command("cloudflared", args...)
		cmd.SysProcAttr = detachedSysProcAttr()
	} else {
		cmd = exec.Command("cloudflared", args...)
		cmd.SysProcAttr = detachedSysProcAttr()
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start cloudflared: %w (is cloudflared installed?)", err)
	}

	pid := cmd.Process.Pid

	// Wait for URL to appear in log file (up to 15 s)
	url, err := waitForURL(logFile, 15*time.Second)
	if err != nil {
		// Kill the process if we couldn't get the URL
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("cloudflared started (pid %d) but URL not found in log: %w", pid, err)
	}

	entry := &TunnelEntry{
		Name:      name,
		Port:      port,
		PID:       pid,
		URL:       url,
		LogFile:   logFile,
		StartedAt: time.Now(),
	}

	st.Tunnels[name] = entry
	if err := m.save(st); err != nil {
		return entry, fmt.Errorf("save state: %w", err)
	}

	return entry, nil
}

// Stop kills the named tunnel process and removes it from state.
func (m *Manager) Stop(name string) error {
	st, err := m.load()
	if err != nil {
		return err
	}
	e, ok := st.Tunnels[name]
	if !ok {
		return fmt.Errorf("tunnel %q not found", name)
	}

	if isAlive(e.PID) {
		proc, err := os.FindProcess(e.PID)
		if err == nil {
			_ = proc.Kill()
		}
	}

	delete(st.Tunnels, name)
	return m.save(st)
}

// StopAll kills all tracked tunnels.
func (m *Manager) StopAll() error {
	st, err := m.load()
	if err != nil {
		return err
	}
	for _, e := range st.Tunnels {
		if isAlive(e.PID) {
			if proc, err := os.FindProcess(e.PID); err == nil {
				_ = proc.Kill()
			}
		}
	}
	st.Tunnels = map[string]*TunnelEntry{}
	return m.save(st)
}

// List returns all tracked tunnels with live/dead status.
func (m *Manager) List() []*TunnelEntry {
	st, _ := m.load()
	out := make([]*TunnelEntry, 0, len(st.Tunnels))
	for _, e := range st.Tunnels {
		out = append(out, e)
	}
	return out
}

// Get returns a single tunnel entry by name.
func (m *Manager) Get(name string) (*TunnelEntry, bool) {
	st, _ := m.load()
	e, ok := st.Tunnels[name]
	return e, ok
}

// IsAlive reports whether the tunnel process is still running.
func (m *Manager) IsAlive(name string) bool {
	e, ok := m.Get(name)
	if !ok {
		return false
	}
	return isAlive(e.PID)
}

// Prune removes dead entries from state.
func (m *Manager) Prune() int {
	st, _ := m.load()
	removed := 0
	for name, e := range st.Tunnels {
		if !isAlive(e.PID) {
			delete(st.Tunnels, name)
			removed++
		}
	}
	_ = m.save(st)
	return removed
}

// ── internal helpers ──────────────────────────────────────────────────────────

var urlRe = regexp.MustCompile(`https://[a-z0-9\-]+\.trycloudflare\.com`)

// waitForURL polls the cloudflared log file until a trycloudflare URL appears.
func waitForURL(logFile string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		url, err := scanLogForURL(logFile)
		if err == nil && url != "" {
			return url, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return "", fmt.Errorf("timeout after %s waiting for tunnel URL", timeout)
}

func scanLogForURL(logFile string) (string, error) {
	f, err := os.Open(logFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if m := urlRe.FindString(line); m != "" {
			return m, nil
		}
	}
	return "", nil
}

func (m *Manager) load() (state, error) {
	st := state{Tunnels: map[string]*TunnelEntry{}}
	data, err := os.ReadFile(m.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return st, nil
		}
		return st, err
	}
	if err := json.Unmarshal(data, &st); err != nil {
		return st, err
	}
	if st.Tunnels == nil {
		st.Tunnels = map[string]*TunnelEntry{}
	}
	return st, nil
}

func (m *Manager) save(st state) error {
	if err := os.MkdirAll(filepath.Dir(m.stateFile), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.stateFile, data, 0o644)
}

// isAlive checks if a process with the given PID is running.
func isAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds; we need to send signal 0.
	// On Windows, FindProcess returns an error if the process doesn't exist.
	if runtime.GOOS == "windows" {
		// Try to open the process — if it fails, it's dead
		return windowsProcessAlive(pid)
	}
	// Unix: signal 0
	err = proc.Signal(os.Signal(nil))
	_ = err
	// A nil signal error means the process exists
	return err == nil
}

// windowsProcessAlive uses tasklist to check if a PID is alive on Windows.
func windowsProcessAlive(pid int) bool {
	out, err := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/NH", "/FO", "CSV").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), strconv.Itoa(pid))
}
