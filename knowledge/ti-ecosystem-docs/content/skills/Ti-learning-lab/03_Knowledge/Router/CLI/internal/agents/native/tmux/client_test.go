package tmux

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		Session:      "test-session",
		DefaultShell: "bash",
		Timeout:      30 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.GetSession() != "test-session" {
		t.Errorf("Expected session 'test-session', got '%s'", client.GetSession())
	}
}

func TestNewClientDefaultConfig(t *testing.T) {
	client, err := NewClient(nil)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.GetSession() != "ti-session" {
		t.Errorf("Expected default session 'ti-session', got '%s'", client.GetSession())
	}
}

func TestSetSession(t *testing.T) {
	client, err := NewClient(nil)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	client.SetSession("new-session")

	if client.GetSession() != "new-session" {
		t.Errorf("Expected session 'new-session', got '%s'", client.GetSession())
	}
}

func TestListSessions(t *testing.T) {
	client, err := NewClient(nil)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// This will fail if tmux is not installed or no sessions exist
	// For testing purposes, we expect it to not crash
	sessions, err := client.ListSessions(context.Background())

	// It's okay if tmux is not available
	if err != nil {
		t.Logf("tmux not available (expected in test environment): %v", err)
		return
	}

	// If tmux is available, sessions should be a list (possibly empty)
	if sessions == nil {
		t.Error("Expected sessions to be a list, got nil")
	}
}

func TestSendCommand(t *testing.T) {
	client, err := NewClient(nil)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// This will fail if tmux is not installed
	// For testing purposes, we expect it to not crash
	err = client.SendCommand(context.Background(), "test-window:0.0", "echo test")

	// It's okay if tmux is not available
	if err != nil {
		t.Logf("tmux not available (expected in test environment): %v", err)
	}
}

func TestCaptureOutput(t *testing.T) {
	client, err := NewClient(nil)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// This will fail if tmux is not installed
	output, err := client.CaptureOutput(context.Background(), "test-window:0.0")

	// It's okay if tmux is not available
	if err != nil {
		t.Logf("tmux not available (expected in test environment): %v", err)
		return
	}

	if output == "" {
		// Empty output is possible
		t.Log("Captured empty output")
	}
}

func TestWaitForShellTimeout(t *testing.T) {
	client, err := NewClient(nil)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Test with a very short timeout - should timeout
	err = client.WaitForShell(context.Background(), "nonexistent-window", 10*time.Millisecond)

	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestWaitUntilStatusTimeout(t *testing.T) {
	client, err := NewClient(nil)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Test with a very short timeout - should timeout
	err = client.WaitUntilStatus(context.Background(), "nonexistent-window", "ready", 10*time.Millisecond)

	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestCommandResult(t *testing.T) {
	result := &CommandResult{
		Success: true,
		Output:  "test output",
		Error:   "",
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Output != "test output" {
		t.Errorf("Expected output 'test output', got '%s'", result.Output)
	}
}
