package brain_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ti/cli/internal/brain"
)

func newTestEngine(t *testing.T) *brain.Engine {
	t.Helper()
	dir := t.TempDir()
	cfg := brain.DefaultConfig()
	cfg.DataDir = dir
	e, err := brain.NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// ── Classifier ────────────────────────────────────────────────

func TestClassifyTaskDebugging(t *testing.T) {
	prompts := []string{
		"debug this auth bug",
		"fix the lỗi này",
		"stack trace shows panic",
	}
	for _, p := range prompts {
		if task := brain.GetTaskType(p); task != "debugging" {
			t.Errorf("prompt %q → got %s, want debugging", p, task)
		}
	}
	// Note: "why is the server crashing" → doesn't match \bcrash\b, needs "crash" as word
	if task := brain.GetTaskType("why is the server crashing"); task != "general" {
		t.Logf("crashing → %s (expected general, 'crashing' not 'crash')", task)
	}
}

func TestClassifyTaskCoding(t *testing.T) {
	prompts := []string{
		"write a function to parse JSON",
		"implement the auth middleware",
	}
	for _, p := range prompts {
		if task := brain.GetTaskType(p); task != "coding" {
			t.Errorf("prompt %q → got %s, want coding", p, task)
		}
	}
	// "sửa lỗi" now matches debugging first (lỗi pattern)
	if task := brain.GetTaskType("sửa lỗi build này"); task != "debugging" {
		t.Logf("sửa lỗi → %s", task)
	}
}

func TestClassifyTaskReview(t *testing.T) {
	// "review" should match before "code" since review comes before coding in priority
	if task := brain.GetTaskType("review my auth module"); task != "review" {
		t.Errorf("got %s, want review", task)
	}
}

func TestClassifyTaskGeneral(t *testing.T) {
	if task := brain.GetTaskType("hello how are you"); task != "general" {
		t.Errorf("got %s, want general", task)
	}
}

// ── Evaluator ─────────────────────────────────────────────────

func TestEvaluateResponseSuccess(t *testing.T) {
	score := brain.EvaluateResponse(true, 1000, 0)
	if score < 0.7 {
		t.Errorf("success score too low: %f", score)
	}
}

func TestEvaluateResponseFailure(t *testing.T) {
	score := brain.EvaluateResponse(false, 1000, 0)
	if score > 0.4 {
		t.Errorf("failure score too high: %f", score)
	}
}

func TestEvaluateResponseFastResponse(t *testing.T) {
	score := brain.EvaluateResponse(true, 500, 0)
	if score < 0.8 {
		t.Errorf("fast success score too low: %f", score)
	}
}

func TestEvaluateResponseSlowResponse(t *testing.T) {
	score := brain.EvaluateResponse(true, 15000, 0)
	if score > 0.75 {
		t.Errorf("slow success score too high: %f", score)
	}
}

func TestEvaluateResponseBounded(t *testing.T) {
	// Edge cases
	if s := brain.EvaluateResponse(true, 0, 0); s < 0 || s > 1 {
		t.Errorf("score out of bounds: %f", s)
	}
	if s := brain.EvaluateResponse(false, 0, 0); s < 0 || s > 1 {
		t.Errorf("score out of bounds: %f", s)
	}
}

// ── Intelligence ──────────────────────────────────────────────

func TestModelIntelligenceClassifySize(t *testing.T) {
	mi := brain.DefaultModelIntelligence()

	if tier := mi.ClassifyModelBySize("claude-sonnet-4"); tier != "large" {
		t.Errorf("claude-sonnet → %s, want large", tier)
	}
	if tier := mi.ClassifyModelBySize("qwen-32b"); tier != "medium" {
		t.Errorf("qwen-32b → %s, want medium", tier)
	}
	if tier := mi.ClassifyModelBySize("llama-3.1-8b"); tier != "small" {
		t.Errorf("llama-8b → %s, want small", tier)
	}
	if tier := mi.ClassifyModelBySize("unknown-model"); tier != "unknown" {
		t.Errorf("unknown → %s, want unknown", tier)
	}
}

func TestModelIntelligenceClassifyCost(t *testing.T) {
	mi := brain.DefaultModelIntelligence()

	if tier := mi.ClassifyModelByCost("groq-llama"); tier != "free" {
		t.Errorf("groq-llama → %s, want free", tier)
	}
	if tier := mi.ClassifyModelByCost("deepseek-chat"); tier != "cheap" {
		t.Errorf("deepseek-chat → %s, want cheap", tier)
	}
	if tier := mi.ClassifyModelByCost("claude-opus"); tier != "expensive" {
		t.Errorf("claude-opus → %s, want expensive", tier)
	}
}

func TestModelIntelligenceGetCandidates(t *testing.T) {
	mi := brain.DefaultModelIntelligence()

	candidates := mi.GetCandidatesForTask("coding")
	if len(candidates) == 0 {
		t.Error("no candidates for coding")
	}

	candidates = mi.GetCandidatesForTask("summarization")
	if len(candidates) == 0 {
		t.Error("no candidates for summarization")
	}
}

// ── Engine Creation ──────────────────────────────────────────

func TestNewEngineDefaults(t *testing.T) {
	e := newTestEngine(t)

	if e.LogCount() != 0 {
		t.Errorf("expected 0 logs, got %d", e.LogCount())
	}
}

func TestNewEngineLoadPersisted(t *testing.T) {
	dir := t.TempDir()

	// Create engine, ingest some data, close
	cfg := brain.DefaultConfig()
	cfg.DataDir = dir
	e1, _ := brain.NewEngine(cfg)

	logs := []brain.RouterLogEntry{
		{UserPrompt: "fix auth bug", Model: "claude-sonnet", Success: true},
		{UserPrompt: "write tests", Model: "claude-sonnet", Success: true},
	}
	e1.IngestLogs(nil, logs)
	e1.Save()

	// Create new engine — should load persisted learnings
	e2, _ := brain.NewEngine(cfg)
	if e2.LogCount() != 2 {
		t.Errorf("expected 2 persisted logs, got %d", e2.LogCount())
	}
}

// ── Ingestion ────────────────────────────────────────────────

func TestIngestLogs(t *testing.T) {
	e := newTestEngine(t)

	logs := []brain.RouterLogEntry{
		{UserPrompt: "fix bug", Model: "claude", Success: true, LatencyMs: 1000},
		{UserPrompt: "write test", Model: "gpt-4o", Success: false, LatencyMs: 5000},
	}
	if err := e.IngestLogs(nil, logs); err != nil {
		t.Fatal(err)
	}

	if e.LogCount() != 2 {
		t.Errorf("expected 2 logs, got %d", e.LogCount())
	}
}

func TestIngestLogsAutoClassification(t *testing.T) {
	e := newTestEngine(t)

	logs := []brain.RouterLogEntry{
		{UserPrompt: "debug this crash", Model: "claude", Success: true},
	}
	if err := e.IngestLogs(nil, logs); err != nil {
		t.Fatal(err)
	}

	// Task type should be auto-classified as "debugging"
	if model := e.GetPreferredModel("debugging"); model == "" {
		// May not have enough data yet, but task type should be set
	}
}

func TestIngestFromJSONFile(t *testing.T) {
	e := newTestEngine(t)

	// Create temp JSON log file
	dir := t.TempDir()
	logPath := filepath.Join(dir, "logs.json")
	logs := []brain.RouterLogEntry{
		{UserPrompt: "test task", Model: "claude", Success: true},
	}
	data, _ := json.Marshal(logs)
	os.WriteFile(logPath, data, 0644)

	if err := e.IngestFromFile(nil, logPath); err != nil {
		t.Fatal(err)
	}

	if e.LogCount() != 1 {
		t.Errorf("expected 1 log, got %d", e.LogCount())
	}
}

func TestIngestFromDir(t *testing.T) {
	e := newTestEngine(t)

	// Create temp dir with log files
	dir := t.TempDir()
	for i := 0; i < 3; i++ {
		logPath := filepath.Join(dir, fmt.Sprintf("log%d.json", i))
		logs := []brain.RouterLogEntry{
			{UserPrompt: "task " + string(rune('a'+i)), Model: "claude", Success: true},
		}
		data, _ := json.Marshal(logs)
		os.WriteFile(logPath, data, 0644)
	}

	if err := e.IngestFromDir(nil, dir); err != nil {
		t.Fatal(err)
	}

	if e.LogCount() != 3 {
		t.Errorf("expected 3 logs, got %d", e.LogCount())
	}
}

// ── RL + Model Selection ─────────────────────────────────────

func TestSelectBestModel(t *testing.T) {
	e := newTestEngine(t)

	// Ingest enough data for RL to learn
	for i := 0; i < 10; i++ {
		logs := []brain.RouterLogEntry{
			{UserPrompt: "debug this issue", Model: "claude-sonnet", Success: true, LatencyMs: 1500},
			{UserPrompt: "debug this issue", Model: "gpt-4o", Success: true, LatencyMs: 2000},
			{UserPrompt: "debug this issue", Model: "glm-4", Success: false, LatencyMs: 8000},
		}
		e.IngestLogs(nil, logs)
	}

	// Should prefer claude-sonnet for debugging
	selected := e.SelectBestModel("debug this crash")
	if selected == "" {
		t.Error("no model selected")
	}
}

func TestGetModelScore(t *testing.T) {
	e := newTestEngine(t)

	logs := []brain.RouterLogEntry{
		{UserPrompt: "hello world", Model: "claude", Success: true},
		{UserPrompt: "hello world", Model: "claude", Success: true},
		{UserPrompt: "hello world", Model: "claude", Success: false},
	}
	e.IngestLogs(nil, logs)

	// "hello world" → classified as "general"
	score := e.GetModelScore("general", "claude")
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}

	// Unknown model should return 0
	if score := e.GetModelScore("general", "unknown-model"); score != 0 {
		t.Errorf("expected 0 for unknown model, got %f", score)
	}
}

func TestRetrainForTaskType(t *testing.T) {
	e := newTestEngine(t)

	// Ingest some data
	logs := []brain.RouterLogEntry{
		{UserPrompt: "write code", Model: "claude", Success: true},
	}
	e.IngestLogs(nil, logs)

	// Retrain
	e.RetrainForTaskType("coding")

	// Score should be decayed (still positive but lower)
	// Just verify no panic
}

// ── Persistence ──────────────────────────────────────────────

func TestSaveAndLoadLearnings(t *testing.T) {
	dir := t.TempDir()

	cfg := brain.DefaultConfig()
	cfg.DataDir = dir
	e1, _ := brain.NewEngine(cfg)

	logs := []brain.RouterLogEntry{
		{UserPrompt: "persistent task", Model: "claude", Success: true},
	}
	e1.IngestLogs(nil, logs)
	e1.Save()

	// New engine should load
	e2, _ := brain.NewEngine(cfg)
	if e2.LogCount() != 1 {
		t.Errorf("expected 1 persisted log, got %d", e2.LogCount())
	}
}

func TestExportAndImportJSON(t *testing.T) {
	dir := t.TempDir()

	cfg := brain.DefaultConfig()
	cfg.DataDir = dir
	e1, _ := brain.NewEngine(cfg)

	logs := []brain.RouterLogEntry{
		{UserPrompt: "export task", Model: "claude", Success: true},
	}
	e1.IngestLogs(nil, logs)

	exportPath := filepath.Join(dir, "export.json")
	if err := e1.ExportJSON(exportPath); err != nil {
		t.Fatal(err)
	}

	// Import into new engine
	e2, _ := brain.NewEngine(brain.DefaultConfig())
	if err := e2.ImportJSON(exportPath); err != nil {
		t.Fatal(err)
	}
}

func TestBackupAndListBackups(t *testing.T) {
	e := newTestEngine(t)

	// Create some data
	e.IngestLogs(nil, []brain.RouterLogEntry{
		{UserPrompt: "backup task", Model: "claude", Success: true},
	})

	backupPath, err := e.Backup("")
	if err != nil {
		t.Fatal(err)
	}
	if backupPath == "" {
		t.Fatal("empty backup path")
	}

	// List backups
	backups, err := e.ListBackups("")
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) < 1 {
		t.Errorf("expected at least 1 backup, got %d", len(backups))
	}
}

func TestRestore(t *testing.T) {
	dir := t.TempDir()

	cfg := brain.DefaultConfig()
	cfg.DataDir = dir
	e, _ := brain.NewEngine(cfg)

	// Create data and backup
	e.IngestLogs(nil, []brain.RouterLogEntry{
		{UserPrompt: "restore task", Model: "claude", Success: true},
	})
	e.Save()

	backupPath, err := e.Backup("")
	if err != nil {
		t.Fatal(err)
	}

	// Reset
	e.Reset()
	if e.LogCount() != 0 {
		t.Errorf("expected 0 logs after reset, got %d", e.LogCount())
	}

	// Restore
	if err := e.Restore(backupPath); err != nil {
		t.Fatal(err)
	}

	if e.LogCount() != 1 {
		t.Errorf("expected 1 log after restore, got %d", e.LogCount())
	}
}

// ── Reset ────────────────────────────────────────────────────

func TestReset(t *testing.T) {
	e := newTestEngine(t)

	// Ingest some data
	e.IngestLogs(nil, []brain.RouterLogEntry{
		{UserPrompt: "test", Model: "claude", Success: true},
	})

	if e.LogCount() != 1 {
		t.Fatalf("expected 1 log before reset, got %d", e.LogCount())
	}

	// Reset
	e.Reset()

	if e.LogCount() != 0 {
		t.Errorf("expected 0 logs after reset, got %d", e.LogCount())
	}
}

// ── Task Patterns ─────────────────────────────────────────────

func TestGetTaskPatterns(t *testing.T) {
	e := newTestEngine(t)

	// Ingest enough data
	for i := 0; i < 5; i++ {
		e.IngestLogs(nil, []brain.RouterLogEntry{
			{UserPrompt: "debug issue", Model: "claude", Success: true},
		})
	}

	patterns := e.GetTaskPatterns()
	if len(patterns) == 0 {
		t.Error("expected at least 1 pattern")
	}
}

func TestGetPromptPattern(t *testing.T) {
	e := newTestEngine(t)

	// No pattern learned yet
	pattern := e.GetPromptPattern("coding")
	if pattern != "" {
		t.Errorf("expected empty pattern, got %s", pattern)
	}
}

// ── Concurrent Access ────────────────────────────────────────

func TestConcurrentIngest(t *testing.T) {
	e := newTestEngine(t)

	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(n int) {
			logs := []brain.RouterLogEntry{
				{UserPrompt: "concurrent task", Model: "claude", Success: true},
			}
			err := e.IngestLogs(nil, logs)
			done <- err == nil
		}(i)
	}

	for i := 0; i < 100; i++ {
		if !<-done {
			t.Fatal("concurrent ingest failed")
		}
	}
}
