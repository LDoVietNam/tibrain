package brain

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestIngestBrainFeed(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "brain_feed_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test brain-feed.jsonl
	feedPath := filepath.Join(tmpDir, "brain-feed.jsonl")
	feedContent := `{"timestamp":"2026-04-27T13:31:00.0000000+07:00","user_prompt":"Test task","response":"Test response","task_type":"coding","model":"claude-sonnet-4","agent":"claude","phase":"execution","domain":"test","success":true,"score":0.9,"latency_ms":100,"cost":0.0,"budget":0.0,"session_id":"test-session"}
{"timestamp":"2026-04-27T13:32:00.0000000+07:00","user_prompt":"Debug task","response":"Debug response","task_type":"debugging","model":"claude-sonnet-4","agent":"claude","phase":"execution","domain":"test","success":true,"score":0.85,"latency_ms":150,"cost":0.0,"budget":0.0,"session_id":"test-session-2"}
`
	if err := os.WriteFile(feedPath, []byte(feedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create engine with test config
	cfg := Config{
		DataDir:          tmpDir,
		MaxLogsPerIngest: 1000,
		MinConfidence:    0.75,
		LearningRate:     0.1,
		WorkerCount:      4,
	}

	e, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Ingest brain feed
	ctx := context.Background()
	if err := e.IngestBrainFeed(ctx, feedPath); err != nil {
		t.Fatal(err)
	}

	// Verify logs were ingested
	if e.LogCount() != 2 {
		t.Errorf("Expected 2 logs, got %d", e.LogCount())
	}

	// Verify task types were learned
	preferredCoding := e.GetPreferredModel("coding")
	if preferredCoding == "" {
		t.Error("Expected preferred model for coding to be set")
	}

	preferredDebugging := e.GetPreferredModel("debugging")
	if preferredDebugging == "" {
		t.Error("Expected preferred model for debugging to be set")
	}
}

func TestInitializeWithBrainFeed(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "brain_feed_init_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test brain-feed.jsonl
	feedPath := filepath.Join(tmpDir, "brain-feed.jsonl")
	feedContent := `{"timestamp":"2026-04-27T13:31:00.0000000+07:00","user_prompt":"Test task","response":"Test response","task_type":"coding","model":"claude-sonnet-4","agent":"claude","phase":"execution","domain":"test","success":true,"score":0.9,"latency_ms":100,"cost":0.0,"budget":0.0,"session_id":"test-session"}
`
	if err := os.WriteFile(feedPath, []byte(feedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize with brain feed
	cfg := Config{
		DataDir:          tmpDir,
		MaxLogsPerIngest: 1000,
		MinConfidence:    0.75,
		LearningRate:     0.1,
		WorkerCount:      4,
	}

	e, err := InitializeWithBrainFeed(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Verify logs were ingested on startup
	if e.LogCount() != 1 {
		t.Errorf("Expected 1 log after initialization, got %d", e.LogCount())
	}
}

func TestReloadBrainFeed(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "brain_feed_reload_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create engine
	cfg := Config{
		DataDir:          tmpDir,
		MaxLogsPerIngest: 1000,
		MinConfidence:    0.75,
		LearningRate:     0.1,
		WorkerCount:      4,
	}

	e, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Initially no logs
	if e.LogCount() != 0 {
		t.Errorf("Expected 0 logs initially, got %d", e.LogCount())
	}

	// Create test brain-feed.jsonl
	feedPath := filepath.Join(tmpDir, "brain-feed.jsonl")
	feedContent := `{"timestamp":"2026-04-27T13:31:00.0000000+07:00","user_prompt":"Test task","response":"Test response","task_type":"coding","model":"claude-sonnet-4","agent":"claude","phase":"execution","domain":"test","success":true,"score":0.9,"latency_ms":100,"cost":0.0,"budget":0.0,"session_id":"test-session"}
`
	if err := os.WriteFile(feedPath, []byte(feedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Reload brain feed
	ctx := context.Background()
	if err := e.ReloadBrainFeed(ctx); err != nil {
		t.Fatal(err)
	}

	// Verify logs were ingested
	if e.LogCount() != 1 {
		t.Errorf("Expected 1 log after reload, got %d", e.LogCount())
	}
}
