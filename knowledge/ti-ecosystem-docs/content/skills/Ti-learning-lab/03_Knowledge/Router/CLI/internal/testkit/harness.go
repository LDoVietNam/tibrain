package testkit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ti/cli/internal/config"
)

type Harness struct {
	Root    string
	DataDir string
	AuthDir string
}

func New(t *testing.T) *Harness {
	t.Helper()
	root := t.TempDir()
	h := &Harness{
		Root:    root,
		DataDir: filepath.Join(root, ".ti", "data"),
		AuthDir: filepath.Join(root, ".ti", "auth"),
	}
	if err := os.MkdirAll(h.DataDir, 0o755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	if err := os.MkdirAll(h.AuthDir, 0o755); err != nil {
		t.Fatalf("mkdir auth: %v", err)
	}
	return h
}

func (h *Harness) Config() *config.Config {
	return &config.Config{
		DataDir:      h.DataDir,
		AuthDir:      h.AuthDir,
		Provider:     "anthropic",
		Model:        "claude-sonnet-4",
		DefaultPhase: "implement",
		MaxTokens:    8192,
		Host:         "127.0.0.1",
		Port:         1810,
	}
}
