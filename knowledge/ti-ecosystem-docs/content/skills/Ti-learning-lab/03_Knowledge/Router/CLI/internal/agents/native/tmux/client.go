package tmux

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Client provides tmux operations
type Client struct {
	mu           sync.RWMutex
	session      string
	socketPath   string
	defaultShell string
	timeout      time.Duration
}

// Config holds tmux client configuration
type Config struct {
	Session      string
	SocketPath   string
	DefaultShell string
	Timeout      time.Duration
}

// CommandResult represents the result of a tmux command
type CommandResult struct {
	Success bool
	Output  string
	Error   string
}

// NewClient creates a new tmux client
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = &Config{
			Session:      "ti-session",
			SocketPath:   "",
			DefaultShell: "bash",
			Timeout:      30 * time.Second,
		}
	}

	return &Client{
		session:      config.Session,
		socketPath:   config.SocketPath,
		defaultShell: config.DefaultShell,
		timeout:      config.Timeout,
	}, nil
}

// WaitForShell waits for shell readiness
func (c *Client) WaitForShell(ctx context.Context, window string, timeout time.Duration) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Check if shell is ready by sending a simple command
			result, err := c.runCommand(ctx, "send-keys", window, "Enter")
			if err == nil && result.Success {
				return nil
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	return fmt.Errorf("shell not ready within timeout")
}

// WaitUntilStatus waits until a specific status is reached
func (c *Client) WaitUntilStatus(ctx context.Context, window string, expectedStatus string, timeout time.Duration) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Capture pane content and check status
			result, err := c.runCommand(ctx, "capture-pane", "-p", "-t", window)
			if err == nil && result.Success {
				if strings.Contains(result.Output, expectedStatus) {
					return nil
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	return fmt.Errorf("status '%s' not reached within timeout", expectedStatus)
}

// SendCommand sends a command to a tmux pane
func (c *Client) SendCommand(ctx context.Context, window, command string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Escape special characters in command
	escapedCommand := strings.ReplaceAll(command, " ", "\\ ")

	result, err := c.runCommand(ctx, "send-keys", "-t", window, escapedCommand, "Enter")
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("command failed: %s", result.Error)
	}

	return nil
}

// CaptureOutput captures output from a tmux pane
func (c *Client) CaptureOutput(ctx context.Context, window string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result, err := c.runCommand(ctx, "capture-pane", "-p", "-t", window)
	if err != nil {
		return "", fmt.Errorf("failed to capture output: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("capture failed: %s", result.Error)
	}

	return result.Output, nil
}

// NewWindow creates a new tmux window
func (c *Client) NewWindow(ctx context.Context, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	result, err := c.runCommand(ctx, "new-window", "-n", name)
	if err != nil {
		return fmt.Errorf("failed to create window: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("window creation failed: %s", result.Error)
	}

	return nil
}

// SplitWindow splits the current window
func (c *Client) SplitWindow(ctx context.Context, window string, vertical bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	args := []string{"split-window", "-t", window}
	if vertical {
		args = append(args, "-v")
	}

	result, err := c.runCommand(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to split window: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("split failed: %s", result.Error)
	}

	return nil
}

// KillPane kills a tmux pane
func (c *Client) KillPane(ctx context.Context, pane string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	result, err := c.runCommand(ctx, "kill-pane", "-t", pane)
	if err != nil {
		return fmt.Errorf("failed to kill pane: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("kill pane failed: %s", result.Error)
	}

	return nil
}

// ListSessions lists all tmux sessions
func (c *Client) ListSessions(ctx context.Context) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result, err := c.runCommand(ctx, "list-sessions")
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("list sessions failed: %s", result.Error)
	}

	// Parse session names
	lines := strings.Split(result.Output, "\n")
	sessions := make([]string, 0)
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				sessions = append(sessions, parts[0])
			}
		}
	}

	return sessions, nil
}

// AttachToSession attaches to a tmux session
func (c *Client) AttachToSession(ctx context.Context, session string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	result, err := c.runCommand(ctx, "attach-session", "-t", session)
	if err != nil {
		return fmt.Errorf("failed to attach to session: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("attach failed: %s", result.Error)
	}

	return nil
}

// runCommand executes a tmux command
func (c *Client) runCommand(ctx context.Context, args ...string) (*CommandResult, error) {
	// Build command with socket path if specified
	cmdArgs := []string{}
	if c.socketPath != "" {
		cmdArgs = append(cmdArgs, "-S", c.socketPath)
	}
	cmdArgs = append(cmdArgs, args...)

	// Create command
	cmd := exec.CommandContext(ctx, "tmux", cmdArgs...)

	// Set timeout context if specified
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	// Execute command
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return &CommandResult{
			Success: false,
			Output:  stdout.String(),
			Error:   stderr.String(),
		}, err
	}

	return &CommandResult{
		Success: true,
		Output:  stdout.String(),
		Error:   stderr.String(),
	}, nil
}

// GetSession returns the current session name
func (c *Client) GetSession() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.session
}

// SetSession sets the session name
func (c *Client) SetSession(session string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session = session
}
