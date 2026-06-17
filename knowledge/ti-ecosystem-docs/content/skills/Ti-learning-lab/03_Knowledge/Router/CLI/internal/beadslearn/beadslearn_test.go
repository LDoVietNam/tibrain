package beadslearn_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ti/cli/internal/beads"
	"github.com/ti/cli/internal/beadslearn"
	"github.com/ti/cli/internal/memory"
)

func setupLearn(t *testing.T) (*beadslearn.Learn, *memory.Palace, string) {
	t.Helper()
	dir := t.TempDir()

	palace := memory.NewPalace(dir)
	if err := palace.Load(); err != nil {
		t.Fatal(err)
	}

	beadsPath := filepath.Join(dir, "beads.jsonl")
	l := beadslearn.NewLearn(palace, beadsPath)

	t.Cleanup(func() {
		l.Save()
		palace.Close()
	})

	return l, palace, beadsPath
}

func makeEntry(model, taskType, phase, domain, task string, quality float64, success bool) beads.Entry {
	return beads.Entry{
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		Status:            "complete",
		TaskType:          taskType,
		Phase:             phase,
		Domain:            domain,
		Complexity:        0.5,
		Budget:            1.0,
		Task:              task,
		Model:             model,
		Agent:             "coder",
		Reasoning:         "best for task",
		Alternatives:      []string{"other-model"},
		PromptTokens:      500,
		CompletionTokens:  800,
		Quality:           quality,
		LatencyMs:         1500,
		Cost:              0.05,
		Success:           success,
		QualityAcceptable: quality >= 0.7,
		CostAcceptable:    true,
		LatencyAcceptable: true,
		SessionID:         "test-session",
		Project:           "test-project",
	}
}

func TestProcessBEADSEntry(t *testing.T) {
	l, palace, _ := setupLearn(t)

	entry := makeEntry("claude-sonnet-4", "coding", "implement", "auth", "fix auth bug", 0.9, true)
	if err := l.ProcessBEADSEntry(entry); err != nil {
		t.Fatal(err)
	}

	mp := l.GetModelPerformance("claude-sonnet-4", "coding")
	if mp == nil {
		t.Fatal("expected model performance data")
	}
	if mp.TotalTasks != 1 {
		t.Fatalf("expected 1 task, got %d", mp.TotalTasks)
	}
	if mp.SuccessCount != 1 {
		t.Fatalf("expected 1 success, got %d", mp.SuccessCount)
	}
	if mp.TotalQuality != 0.9 {
		t.Fatalf("expected quality 0.9, got %f", mp.TotalQuality)
	}

	drawers := palace.CountDrawers()
	if drawers < 1 {
		t.Fatalf("expected at least 1 drawer, got %d", drawers)
	}
}

func TestProcessRunningEntryIgnored(t *testing.T) {
	l, palace, _ := setupLearn(t)

	initialDrawers := palace.CountDrawers()

	entry := beads.Entry{
		Status: "running",
		Model:  "claude-sonnet-4",
		Task:   "fix auth",
	}
	if err := l.ProcessBEADSEntry(entry); err != nil {
		t.Fatal(err)
	}

	finalDrawers := palace.CountDrawers()
	if finalDrawers != initialDrawers {
		t.Fatalf("expected no new drawers for running task, got %d", finalDrawers-initialDrawers)
	}
}

func TestProcessFailedEntry(t *testing.T) {
	l, palace, _ := setupLearn(t)

	entry := makeEntry("qwen-coder", "coding", "implement", "auth", "fix auth bug", 0.2, false)
	entry.Status = "failed"
	entry.Error = "connection timeout"
	if err := l.ProcessBEADSEntry(entry); err != nil {
		t.Fatal(err)
	}

	mp := l.GetModelPerformance("qwen-coder", "coding")
	if mp == nil {
		t.Fatal("expected model performance data")
	}
	if mp.FailCount != 1 {
		t.Fatalf("expected 1 failure, got %d", mp.FailCount)
	}

	drawers := palace.CountDrawers()
	if drawers < 1 {
		t.Fatalf("expected problem drawer, got %d drawers", drawers)
	}
}

func TestStoreQADrawer(t *testing.T) {
	l, palace, _ := setupLearn(t)

	entry := makeEntry("claude-sonnet-4", "coding", "implement", "auth", "fix auth bug", 0.95, true)
	if err := l.ProcessBEADSEntry(entry); err != nil {
		t.Fatal(err)
	}

	drawers := palace.CountDrawers()
	if drawers < 1 {
		t.Fatalf("expected QA drawer, got %d drawers", drawers)
	}
}

func TestNoQADrawerForLowQuality(t *testing.T) {
	l, palace, _ := setupLearn(t)

	initialDrawers := palace.CountDrawers()

	entry := makeEntry("claude-sonnet-4", "coding", "implement", "auth", "fix auth bug", 0.5, true)
	if err := l.ProcessBEADSEntry(entry); err != nil {
		t.Fatal(err)
	}

	finalDrawers := palace.CountDrawers()
	qaDrawers := finalDrawers - initialDrawers
	if qaDrawers > 1 {
		t.Fatalf("expected max 1 drawer for low quality, got %d", qaDrawers)
	}
}

func TestGetBestModel(t *testing.T) {
	l, _, _ := setupLearn(t)

	for i := 0; i < 5; i++ {
		l.ProcessBEADSEntry(makeEntry("claude", "coding", "implement", "auth", "task", 0.9, true))
		l.ProcessBEADSEntry(makeEntry("qwen", "coding", "implement", "auth", "task", 0.7, true))
		l.ProcessBEADSEntry(makeEntry("gpt", "coding", "implement", "auth", "task", 0.6, true))
	}

	best, quality, tasks := l.GetBestModel("coding")
	if best != "claude" {
		t.Fatalf("expected best model 'claude', got %q", best)
	}
	if quality < 0.85 {
		t.Fatalf("expected quality >= 0.85, got %f", quality)
	}
	if tasks < 5 {
		t.Fatalf("expected >= 5 tasks, got %d", tasks)
	}
}

func TestGetBestModelInsufficientData(t *testing.T) {
	l, _, _ := setupLearn(t)

	l.ProcessBEADSEntry(makeEntry("claude", "coding", "implement", "auth", "task", 0.9, true))

	best, quality, tasks := l.GetBestModel("coding")
	if best != "" || quality != 0 || tasks != 0 {
		t.Fatalf("expected no best model with insufficient data, got %q %.2f %d", best, quality, tasks)
	}
}

func TestGetAllModelStats(t *testing.T) {
	l, _, _ := setupLearn(t)

	l.ProcessBEADSEntry(makeEntry("claude", "coding", "implement", "auth", "task1", 0.9, true))
	l.ProcessBEADSEntry(makeEntry("qwen", "review", "review", "api", "task2", 0.8, true))

	stats := l.GetAllModelStats()
	if len(stats) != 2 {
		t.Fatalf("expected 2 model stats, got %d", len(stats))
	}
}

func TestQualityTrend(t *testing.T) {
	l, _, _ := setupLearn(t)

	for i := 0; i < 10; i++ {
		q := 0.5 + float64(i)*0.05
		l.ProcessBEADSEntry(makeEntry("claude", "coding", "implement", "auth", "task", q, true))
	}

	trend := l.GetQualityTrend("claude")
	if trend <= 0 {
		t.Fatalf("expected positive trend (improving), got %f", trend)
	}
}

func TestQualityTrendDeclining(t *testing.T) {
	l, _, _ := setupLearn(t)

	for i := 0; i < 10; i++ {
		q := 0.95 - float64(i)*0.05
		l.ProcessBEADSEntry(makeEntry("qwen", "coding", "implement", "auth", "task", q, true))
	}

	trend := l.GetQualityTrend("qwen")
	if trend >= 0 {
		t.Fatalf("expected negative trend (declining), got %f", trend)
	}
}

func TestQualityTrendInsufficientData(t *testing.T) {
	l, _, _ := setupLearn(t)

	l.ProcessBEADSEntry(makeEntry("claude", "coding", "implement", "auth", "task", 0.8, true))

	trend := l.GetQualityTrend("claude")
	if trend != 0 {
		t.Fatalf("expected 0 trend with insufficient data, got %f", trend)
	}
}

func TestSaveAndLoad(t *testing.T) {
	l, _, beadsPath := setupLearn(t)

	for i := 0; i < 5; i++ {
		l.ProcessBEADSEntry(makeEntry("claude", "coding", "implement", "auth", "task", 0.9, true))
	}

	if err := l.Save(); err != nil {
		t.Fatal(err)
	}

	dir := filepath.Dir(beadsPath)
	palace2 := memory.NewPalace(dir)
	if err := palace2.Load(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { palace2.Close() })

	l2 := beadslearn.NewLearn(palace2, beadsPath)
	if err := l2.Load(); err != nil {
		t.Fatal(err)
	}

	mp := l2.GetModelPerformance("claude", "coding")
	if mp == nil {
		t.Fatal("expected model performance data after reload")
	}
	if mp.TotalTasks != 5 {
		t.Fatalf("expected 5 tasks after reload, got %d", mp.TotalTasks)
	}
}

func TestGetModelPerformanceNotFound(t *testing.T) {
	l, _, _ := setupLearn(t)

	mp := l.GetModelPerformance("nonexistent", "coding")
	if mp != nil {
		t.Fatal("expected nil for unknown model")
	}
}

func TestProcessEntryWithNoModel(t *testing.T) {
	l, palace, _ := setupLearn(t)

	initialDrawers := palace.CountDrawers()

	entry := beads.Entry{
		Status:  "complete",
		Task:    "some task",
		Quality: 0.8,
		Success: true,
	}
	if err := l.ProcessBEADSEntry(entry); err != nil {
		t.Fatal(err)
	}

	finalDrawers := palace.CountDrawers()
	if finalDrawers < initialDrawers {
		t.Fatal("expected no drawers added for entry without model")
	}
}

func TestProjectAsWing(t *testing.T) {
	l, palace, _ := setupLearn(t)

	entry := makeEntry("claude", "coding", "implement", "auth", "fix auth bug", 0.9, true)
	entry.Project = "my-awesome-project"
	l.ProcessBEADSEntry(entry)

	wings := palace.ListWings()
	found := false
	for _, w := range wings {
		if w == "my-awesome-project" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected wing 'my-awesome-project', got %v", wings)
	}
}

func TestDomainAsRoom(t *testing.T) {
	l, palace, _ := setupLearn(t)

	entry := makeEntry("claude", "coding", "implement", "auth", "fix auth bug", 0.9, true)
	l.ProcessBEADSEntry(entry)

	// Check that domain was used as room in the project wing
	rooms, _ := palace.ListRooms("test-project")
	found := false
	for _, r := range rooms {
		if r == "auth" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected room 'auth' in wing 'test-project', got %v", rooms)
	}
}

func TestConcurrentProcessing(t *testing.T) {
	l, palace, _ := setupLearn(t)

	done := make(chan error, 100)
	for i := 0; i < 100; i++ {
		go func(n int) {
			entry := makeEntry("claude", "coding", "implement", "auth", "task", 0.8+float64(n%20)*0.01, true)
			done <- l.ProcessBEADSEntry(entry)
		}(i)
	}

	for i := 0; i < 100; i++ {
		if err := <-done; err != nil {
			t.Fatalf("concurrent process failed: %v", err)
		}
	}

	drawers := palace.CountDrawers()
	if drawers < 100 {
		t.Fatalf("expected >= 100 drawers, got %d", drawers)
	}
}

func TestGenerateSessionID(t *testing.T) {
	id1 := beadslearn.GenerateSessionID("project-a", "task-b")
	id2 := beadslearn.GenerateSessionID("project-a", "task-b")
	id3 := beadslearn.GenerateSessionID("project-a", "task-c")

	if id1 != id2 {
		t.Fatal("expected same ID for same inputs")
	}
	if id1 == id3 {
		t.Fatal("expected different ID for different inputs")
	}
	if len(id1) != 16 {
		t.Fatalf("expected 16-char ID, got %d", len(id1))
	}
}

func TestSaveStatsCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "deep", "path")
	beadsPath := filepath.Join(subdir, "beads.jsonl")

	palace := memory.NewPalace(dir)
	palace.Load()
	t.Cleanup(func() { palace.Close() })

	l := beadslearn.NewLearn(palace, beadsPath)
	l.ProcessBEADSEntry(makeEntry("claude", "coding", "implement", "auth", "task", 0.9, true))

	if err := l.Save(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(subdir, "model-stats.json")); os.IsNotExist(err) {
		t.Fatal("expected model-stats.json to be created")
	}
}

func TestLoadStatsMissingFile(t *testing.T) {
	dir := t.TempDir()
	beadsPath := filepath.Join(dir, "beads.jsonl")

	palace := memory.NewPalace(dir)
	palace.Load()
	t.Cleanup(func() { palace.Close() })

	l := beadslearn.NewLearn(palace, beadsPath)
	if err := l.Load(); err != nil {
		t.Fatalf("expected no error for missing stats file, got: %v", err)
	}
}
