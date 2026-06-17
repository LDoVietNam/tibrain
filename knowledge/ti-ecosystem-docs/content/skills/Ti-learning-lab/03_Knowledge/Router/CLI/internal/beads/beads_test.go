package beads_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ti/cli/internal/beads"
)

func newTestLogger(t *testing.T) (*beads.Logger, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "beads.jsonl")
	cfg := beads.Config{Path: path}
	l, err := beads.NewLogger(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	return l, path
}

func newTestStore(t *testing.T) *beads.Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "beads.db")
	s, err := beads.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// ── Schema Version ────────────────────────────────────────────

func TestSchemaVersion(t *testing.T) {
	l, path := newTestLogger(t)

	if err := l.Log(beads.Entry{Task: "test"}); err != nil {
		t.Fatal(err)
	}

	entries, err := beads.ReadEntries(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].SchemaVersion != beads.SchemaVersion {
		t.Fatalf("expected schema version %d, got %d", beads.SchemaVersion, entries[0].SchemaVersion)
	}
}

func TestSchemaVersionExplicit(t *testing.T) {
	l, path := newTestLogger(t)

	// Simulate legacy entry (v1)
	if err := l.Log(beads.Entry{SchemaVersion: 1, Task: "legacy"}); err != nil {
		t.Fatal(err)
	}

	entries, err := beads.ReadEntries(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if entries[0].SchemaVersion != 1 {
		t.Fatalf("expected version 1, got %d", entries[0].SchemaVersion)
	}
}

// ── ML Fields ─────────────────────────────────────────────────

func TestMLEntry(t *testing.T) {
	l, path := newTestLogger(t)

	entry := beads.Entry{
		TaskType:          "coding",
		Phase:             "implement",
		Domain:            "auth",
		Complexity:        0.7,
		Budget:            0.50,
		Task:              "fix auth bug",
		Model:             "claude-sonnet-4",
		Agent:             "ticlaw",
		Reasoning:         "implement phase → code model needed",
		Alternatives:      []string{"gpt-4o (cheaper)", "glm-4.7 (free)"},
		Quality:           0.9,
		PromptTokens:      800,
		CompletionTokens:  400,
		LatencyMs:         2300,
		Cost:              0.03,
		Success:           true,
		QualityAcceptable: true,
		CostAcceptable:    true,
		LatencyAcceptable: true,
		SessionID:         "sess-abc",
		Project:           "ticlaw",
	}
	if err := l.Log(entry); err != nil {
		t.Fatal(err)
	}

	entries, err := beads.ReadEntries(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	e := entries[0]

	if e.Phase != "implement" {
		t.Fatalf("wrong phase: %s", e.Phase)
	}
	if e.Domain != "auth" {
		t.Fatalf("wrong domain: %s", e.Domain)
	}
	if e.Complexity != 0.7 {
		t.Fatalf("wrong complexity: %f", e.Complexity)
	}
	if e.Budget != 0.50 {
		t.Fatalf("wrong budget: %f", e.Budget)
	}
	if e.Reasoning == "" {
		t.Fatal("missing reasoning")
	}
	if len(e.Alternatives) != 2 {
		t.Fatalf("expected 2 alternatives, got %d", len(e.Alternatives))
	}
	if !e.QualityAcceptable {
		t.Fatal("quality should be acceptable")
	}
	if !e.CostAcceptable {
		t.Fatal("cost should be acceptable")
	}
	if !e.LatencyAcceptable {
		t.Fatal("latency should be acceptable")
	}
}

// ── Query API ─────────────────────────────────────────────────

func TestQueryByPhase(t *testing.T) {
	l, path := newTestLogger(t)

	phases := []string{"scan", "plan", "implement"}
	for _, p := range phases {
		if err := l.Log(beads.Entry{Phase: p, Task: p + " task"}); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := beads.Query(path, beads.QueryFilter{Phase: []string{"implement"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Phase != "implement" {
		t.Fatalf("wrong phase: %s", entries[0].Phase)
	}
}

func TestQueryByAgent(t *testing.T) {
	l, path := newTestLogger(t)

	agents := []string{"ticlaw", "codex", "aider"}
	for _, a := range agents {
		if err := l.Log(beads.Entry{Agent: a, Task: a + " task"}); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := beads.Query(path, beads.QueryFilter{Agent: []string{"ticlaw", "codex"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestQueryByModel(t *testing.T) {
	l, path := newTestLogger(t)

	models := []string{"claude-sonnet-4", "gpt-4o", "glm-4.7"}
	for _, m := range models {
		if err := l.Log(beads.Entry{Model: m, Task: m + " task"}); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := beads.Query(path, beads.QueryFilter{Model: []string{"claude-sonnet-4"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestQueryByQuality(t *testing.T) {
	l, path := newTestLogger(t)

	qualities := []float64{0.3, 0.6, 0.9}
	for _, q := range qualities {
		if err := l.Log(beads.Entry{Quality: q, Task: "quality task"}); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := beads.Query(path, beads.QueryFilter{MinQuality: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries with quality >= 0.5, got %d", len(entries))
	}
}

func TestQueryByCost(t *testing.T) {
	l, path := newTestLogger(t)

	costs := []float64{0.01, 0.05, 0.10}
	for _, c := range costs {
		if err := l.Log(beads.Entry{Cost: c, Task: "cost task"}); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := beads.Query(path, beads.QueryFilter{MaxCost: 0.06})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries with cost <= 0.06, got %d", len(entries))
	}
}

func TestQueryCaseInsensitive(t *testing.T) {
	l, path := newTestLogger(t)

	if err := l.Log(beads.Entry{Phase: "Implement", Task: "test"}); err != nil {
		t.Fatal(err)
	}

	entries, err := beads.Query(path, beads.QueryFilter{Phase: []string{"implement"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (case-insensitive), got %d", len(entries))
	}
}

// ── Stats with Phase/Agent breakdown ──────────────────────────

func TestComputeStatsWithPhaseBreakdown(t *testing.T) {
	l, path := newTestLogger(t)

	phases := []beads.Entry{
		{Phase: "implement", Model: "claude", Quality: 1.0, Cost: 0.05, Success: true},
		{Phase: "implement", Model: "gpt", Quality: 0.8, Cost: 0.06, Success: true},
		{Phase: "review", Model: "qwen", Quality: 0.6, Cost: 0.01, Success: false},
	}
	for _, e := range phases {
		if err := l.Log(e); err != nil {
			t.Fatal(err)
		}
	}

	stats, err := beads.ComputeStats(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(stats.PhaseBreakdown) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(stats.PhaseBreakdown))
	}

	impl := stats.PhaseBreakdown["implement"]
	if impl.Tasks != 2 {
		t.Fatalf("expected 2 implement tasks, got %d", impl.Tasks)
	}
	if impl.Successes != 2 {
		t.Fatalf("expected 2 implement successes, got %d", impl.Successes)
	}
	if impl.TotalCost != 0.11 {
		t.Fatalf("expected 0.11 implement cost, got %f", impl.TotalCost)
	}
}

func TestComputeStatsWithAgentBreakdown(t *testing.T) {
	l, path := newTestLogger(t)

	agents := []beads.Entry{
		{Agent: "ticlaw", Model: "claude", Quality: 0.9, LatencyMs: 1000, Cost: 0.05, Success: true},
		{Agent: "ticlaw", Model: "claude", Quality: 0.7, LatencyMs: 1500, Cost: 0.06, Success: true},
		{Agent: "codex", Model: "gpt", Quality: 0.5, LatencyMs: 800, Cost: 0.01, Success: false},
	}
	for _, e := range agents {
		if err := l.Log(e); err != nil {
			t.Fatal(err)
		}
	}

	stats, err := beads.ComputeStats(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(stats.AgentBreakdown) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(stats.AgentBreakdown))
	}

	ticlaw := stats.AgentBreakdown["ticlaw"]
	if ticlaw.Tasks != 2 {
		t.Fatalf("expected 2 ticlaw tasks, got %d", ticlaw.Tasks)
	}
	if ticlaw.AvgQuality != 0.8 {
		t.Fatalf("expected 0.8 ticlaw quality, got %f", ticlaw.AvgQuality)
	}
}

func TestComputeStatsAcceptanceRates(t *testing.T) {
	l, path := newTestLogger(t)

	entries := []beads.Entry{
		{SchemaVersion: 2, QualityAcceptable: true, CostAcceptable: true, LatencyAcceptable: true},
		{SchemaVersion: 2, QualityAcceptable: false, CostAcceptable: true, LatencyAcceptable: false},
		{SchemaVersion: 2, QualityAcceptable: false, CostAcceptable: false, LatencyAcceptable: true},
		{SchemaVersion: 1}, // v1 — no acceptance flags, should be excluded from accept rate
	}
	for _, e := range entries {
		if err := l.Log(e); err != nil {
			t.Fatal(err)
		}
	}

	stats, err := beads.ComputeStats(path)
	if err != nil {
		t.Fatal(err)
	}

	// 4 total tasks, 3 have v2 acceptance flags
	// quality: 1/3, cost: 2/3, latency: 2/3
	if stats.QualityAcceptRate != 1.0/3.0 {
		t.Fatalf("expected 0.33 quality accept rate, got %f", stats.QualityAcceptRate)
	}
	if stats.CostAcceptRate != 2.0/3.0 {
		t.Fatalf("expected 0.67 cost accept rate, got %f", stats.CostAcceptRate)
	}
	if stats.LatencyAcceptRate != 2.0/3.0 {
		t.Fatalf("expected 0.67 latency accept rate, got %f", stats.LatencyAcceptRate)
	}
}

// ── SQLite Store ──────────────────────────────────────────────

func TestSQLiteStoreInsert(t *testing.T) {
	s := newTestStore(t)

	entry := beads.Entry{
		SchemaVersion:     beads.SchemaVersion,
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		Status:            "complete",
		TaskType:          "coding",
		Phase:             "implement",
		Domain:            "auth",
		Complexity:        0.7,
		Budget:            0.50,
		Task:              "fix auth bug",
		Model:             "claude-sonnet-4",
		Agent:             "ticlaw",
		Reasoning:         "implement phase → code model",
		Alternatives:      []string{"gpt-4o", "glm-4.7"},
		PromptTokens:      800,
		CompletionTokens:  400,
		Quality:           0.9,
		LatencyMs:         2300,
		Cost:              0.03,
		Success:           true,
		QualityAcceptable: true,
		CostAcceptable:    true,
		LatencyAcceptable: true,
		SessionID:         "sess-abc",
		Project:           "ticlaw",
	}
	if err := s.Insert(entry); err != nil {
		t.Fatal(err)
	}

	count, err := s.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 entry, got %d", count)
	}

	// Query back
	entries, err := s.QueryDB(`SELECT schema_ver, timestamp, status,
		task_type, phase, domain, complexity, budget, task,
		model, agent, reasoning, alternatives,
		prompt_tokens, completion_tokens, quality, latency_ms, cost, success,
		quality_acceptable, cost_acceptable, latency_acceptable,
		session_id, project, error FROM beads ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	e := entries[0]
	if e.Phase != "implement" {
		t.Fatalf("wrong phase: %s", e.Phase)
	}
	if e.Agent != "ticlaw" {
		t.Fatalf("wrong agent: %s", e.Agent)
	}
	if !e.QualityAcceptable {
		t.Fatal("quality should be acceptable")
	}
	if len(e.Alternatives) != 2 {
		t.Fatalf("expected 2 alternatives, got %d", len(e.Alternatives))
	}
}

func TestSQLiteStoreQuerySQL(t *testing.T) {
	s := newTestStore(t)

	// Insert entries with different phases
	for _, phase := range []string{"scan", "plan", "implement", "review"} {
		if err := s.Insert(beads.Entry{
			Phase: phase,
			Task:  phase + " task",
		}); err != nil {
			t.Fatal(err)
		}
	}

	// Query specific phase
	entries, err := s.QueryDB(
		`SELECT schema_ver, timestamp, status,
			task_type, phase, domain, complexity, budget, task,
			model, agent, reasoning, alternatives,
			prompt_tokens, completion_tokens, quality, latency_ms, cost, success,
			quality_acceptable, cost_acceptable, latency_acceptable,
			session_id, project, error
		FROM beads WHERE phase = ?`,
		"implement",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Phase != "implement" {
		t.Fatalf("wrong phase: %s", entries[0].Phase)
	}
}

func TestSQLiteStoreCleanup(t *testing.T) {
	s := newTestStore(t)

	// Insert with old timestamp
	if err := s.Insert(beads.Entry{
		Timestamp: "2020-01-01T00:00:00Z",
		Task:      "old task",
	}); err != nil {
		t.Fatal(err)
	}

	// Insert with recent timestamp
	if err := s.Insert(beads.Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Task:      "new task",
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.Cleanup(24 * time.Hour); err != nil {
		t.Fatal(err)
	}

	count, err := s.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 entry after cleanup, got %d", count)
	}
}

// ── Original Tests (still pass with new schema) ───────────────

func TestLogAndRead(t *testing.T) {
	l, path := newTestLogger(t)

	entry := beads.Entry{
		Task:             "fix auth bug",
		Agent:            "coder",
		Model:            "claude-sonnet-4",
		PromptTokens:     800,
		CompletionTokens: 1200,
		Quality:          1.0,
		LatencyMs:        2300,
		Cost:             0.05,
		Success:          true,
		SessionID:        "abc123",
		Project:          "my-app",
		TaskType:         "coding",
	}
	if err := l.Log(entry); err != nil {
		t.Fatal(err)
	}

	entries, err := beads.ReadEntries(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Task != "fix auth bug" {
		t.Fatalf("wrong task: %s", entries[0].Task)
	}
	if entries[0].Quality != 1.0 {
		t.Fatalf("wrong quality: %f", entries[0].Quality)
	}
}

func TestAutoTimestamp(t *testing.T) {
	l, path := newTestLogger(t)

	if err := l.Log(beads.Entry{Task: "test"}); err != nil {
		t.Fatal(err)
	}

	entries, err := beads.ReadEntries(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatal(err)
	}
	if entries[0].Timestamp == "" {
		t.Fatal("expected auto-timestamp")
	}
}

func TestMultipleEntries(t *testing.T) {
	l, path := newTestLogger(t)

	for i := 0; i < 10; i++ {
		if err := l.Log(beads.Entry{
			Task:    "task " + string(rune('a'+i)),
			Quality: 0.5 + float64(i)*0.05,
			Success: i%2 == 0,
		}); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := beads.ReadEntries(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 10 {
		t.Fatalf("expected 10 entries, got %d", len(entries))
	}
}

func TestFilter(t *testing.T) {
	l, path := newTestLogger(t)

	agents := []string{"coder", "reviewer", "planner"}
	for _, agent := range agents {
		if err := l.Log(beads.Entry{
			Task:  agent + " task",
			Agent: agent,
		}); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := beads.ReadEntries(path, func(e beads.Entry) bool {
		return e.Agent == "coder"
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 coder entry, got %d", len(entries))
	}
	if entries[0].Agent != "coder" {
		t.Fatalf("wrong agent: %s", entries[0].Agent)
	}
}

func TestComputeStats(t *testing.T) {
	l, path := newTestLogger(t)

	entries := []beads.Entry{
		{Task: "task1", Model: "claude", Quality: 1.0, LatencyMs: 1000, Cost: 0.05, Success: true},
		{Task: "task2", Model: "claude", Quality: 0.8, LatencyMs: 1500, Cost: 0.06, Success: true},
		{Task: "task3", Model: "qwen", Quality: 0.6, LatencyMs: 800, Cost: 0.01, Success: false},
	}
	for _, e := range entries {
		if err := l.Log(e); err != nil {
			t.Fatal(err)
		}
	}

	stats, err := beads.ComputeStats(path)
	if err != nil {
		t.Fatal(err)
	}

	if stats.TotalTasks != 3 {
		t.Fatalf("expected 3 tasks, got %d", stats.TotalTasks)
	}
	if stats.SuccessRate != 2.0/3.0 {
		t.Fatalf("expected 0.67 success rate, got %f", stats.SuccessRate)
	}
	if stats.TotalCost != 0.12 {
		t.Fatalf("expected 0.12 cost, got %f", stats.TotalCost)
	}

	if len(stats.ModelBreakdown) != 2 {
		t.Fatalf("expected 2 models, got %d", len(stats.ModelBreakdown))
	}
	claude := stats.ModelBreakdown["claude"]
	if claude.Tasks != 2 {
		t.Fatalf("expected 2 claude tasks, got %d", claude.Tasks)
	}
	if claude.AvgQuality != 0.9 {
		t.Fatalf("expected 0.9 claude quality, got %f", claude.AvgQuality)
	}
}

func TestComputeStatsEmpty(t *testing.T) {
	l, path := newTestLogger(t)
	l.Close()

	f, _ := os.Create(path)
	f.Close()

	stats, err := beads.ComputeStats(path)
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalTasks != 0 {
		t.Fatalf("expected 0 tasks, got %d", stats.TotalTasks)
	}
}

func TestCallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "beads.jsonl")

	var received []beads.Entry
	cfg := beads.Config{
		Path: path,
		Callback: func(e beads.Entry) {
			received = append(received, e)
		},
	}
	l, err := beads.NewLogger(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	if err := l.Log(beads.Entry{Task: "callback test"}); err != nil {
		t.Fatal(err)
	}

	// Give goroutine time to fire
	time.Sleep(100 * time.Millisecond)

	if len(received) != 1 {
		t.Fatalf("expected 1 callback, got %d", len(received))
	}
	if received[0].Task != "callback test" {
		t.Fatalf("wrong task: %s", received[0].Task)
	}
}

func TestConcurrentLogging(t *testing.T) {
	l, path := newTestLogger(t)

	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(n int) {
			err := l.Log(beads.Entry{
				Task:    "concurrent " + string(rune('a'+n%26)),
				Quality: 0.7,
			})
			done <- err == nil
		}(i)
	}

	for i := 0; i < 100; i++ {
		if !<-done {
			t.Fatal("log failed")
		}
	}

	entries, err := beads.ReadEntries(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 100 {
		t.Fatalf("expected 100 entries, got %d", len(entries))
	}
}

func TestStatsWithSingleModel(t *testing.T) {
	l, path := newTestLogger(t)

	for i := 0; i < 5; i++ {
		if err := l.Log(beads.Entry{
			Model:   "claude",
			Quality: 0.8,
			Success: true,
		}); err != nil {
			t.Fatal(err)
		}
	}

	stats, err := beads.ComputeStats(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(stats.ModelBreakdown) != 1 {
		t.Fatalf("expected 1 model, got %d", len(stats.ModelBreakdown))
	}
	if stats.ModelBreakdown["claude"].Tasks != 5 {
		t.Fatalf("expected 5 claude tasks, got %d", stats.ModelBreakdown["claude"].Tasks)
	}
}
