package qa_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ti/cli/internal/qa"
)

func newTestStore(t *testing.T) *qa.Store {
	t.Helper()
	dir := t.TempDir()
	cfg := qa.Config{
		Path:       filepath.Join(dir, "test-qa.db"),
		MinQuality: 0.7,
	}
	s, err := qa.NewStore(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func newEntry(question, answer string) qa.Entry {
	return qa.Entry{
		Question:  question,
		Answer:    answer,
		Model:     "claude-sonnet-4-20250514",
		Quality:   0.9,
		TaskType:  "code",
		AgentID:   "coder",
		CreatedAt: time.Now().Unix(),
	}
}

// 1. Test store and retrieve a single pair
func TestStoreAndGet(t *testing.T) {
	s := newTestStore(t)

	entry := newEntry("How do I handle errors in Go?", "Use errors.Is and errors.As")
	if err := s.Store(entry); err != nil {
		t.Fatal(err)
	}

	results, err := s.Search("handle errors in Go", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Question != "How do I handle errors in Go?" {
		t.Fatalf("wrong question: %s", results[0].Question)
	}
	if results[0].Answer != "Use errors.Is and errors.As" {
		t.Fatalf("wrong answer: %s", results[0].Answer)
	}
}

// 2. Test search by question match
func TestSearchByQuestion(t *testing.T) {
	s := newTestStore(t)

	for i := 0; i < 5; i++ {
		e := newEntry("question "+string(rune('a'+i)), "answer "+string(rune('a'+i)))
		if err := s.Store(e); err != nil {
			t.Fatal(err)
		}
	}

	results, err := s.Search("question b", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'question b', got %d", len(results))
	}
}

// 3. Test search by answer match
func TestSearchByAnswer(t *testing.T) {
	s := newTestStore(t)

	if err := s.Store(newEntry("what is context?", "Context is the current state")); err != nil {
		t.Fatal(err)
	}

	results, err := s.Search("current state", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result searching by answer, got %d", len(results))
	}
}

// 4. Test quality filter rejects low-quality pairs
func TestQualityFilterRejectsLowQuality(t *testing.T) {
	s := newTestStore(t)

	lowEntry := qa.Entry{
		Question: "bad question",
		Answer:   "bad answer",
		Quality:  0.3,
	}
	err := s.Store(lowEntry)
	if err == nil {
		t.Fatal("expected error for low quality, got nil")
	}

	n, _ := s.Count()
	if n != 0 {
		t.Fatalf("expected 0 entries, got %d", n)
	}
}

// 5. Test quality filter accepts minimum quality
func TestQualityFilterAcceptsMinimum(t *testing.T) {
	s := newTestStore(t)

	entry := qa.Entry{
		Question: "minimum quality",
		Answer:   "just enough",
		Quality:  0.7,
	}
	if err := s.Store(entry); err != nil {
		t.Fatalf("expected no error for quality=0.7, got %v", err)
	}

	n, _ := s.Count()
	if n != 1 {
		t.Fatalf("expected 1 entry, got %d", n)
	}
}

// 6. Test increment used counter
func TestIncrementUsed(t *testing.T) {
	s := newTestStore(t)

	entry := newEntry("count test", "answer")
	if err := s.Store(entry); err != nil {
		t.Fatal(err)
	}

	results, _ := s.Search("count test", 1)
	if len(results) != 1 {
		t.Fatal("expected 1 result")
	}
	id := results[0].ID

	for i := 0; i < 3; i++ {
		count, err := s.IncrementUsed(id)
		if err != nil {
			t.Fatal(err)
		}
		if count != int64(i+1) {
			t.Fatalf("expected count %d, got %d", i+1, count)
		}
	}
}

// 7. Test delete
func TestDelete(t *testing.T) {
	s := newTestStore(t)

	entry := newEntry("to delete", "gone")
	if err := s.Store(entry); err != nil {
		t.Fatal(err)
	}

	results, _ := s.Search("to delete", 1)
	if len(results) != 1 {
		t.Fatal("expected 1 result before delete")
	}

	if err := s.Delete(results[0].ID); err != nil {
		t.Fatal(err)
	}

	n, _ := s.Count()
	if n != 0 {
		t.Fatalf("expected 0 after delete, got %d", n)
	}
}

// 8. Test count
func TestCount(t *testing.T) {
	s := newTestStore(t)

	for i := 0; i < 10; i++ {
		e := newEntry("count "+string(rune('a'+i)), "answer")
		if err := s.Store(e); err != nil {
			t.Fatal(err)
		}
	}

	n, err := s.Count()
	if err != nil {
		t.Fatal(err)
	}
	if n != 10 {
		t.Fatalf("expected 10, got %d", n)
	}
}

// 9. Test stats
func TestStats(t *testing.T) {
	s := newTestStore(t)

	entries := []qa.Entry{
		{Question: "go errors", Answer: "use errors.Is", Model: "claude-sonnet", Quality: 0.9, TaskType: "code", AgentID: "coder"},
		{Question: "http routing", Answer: "use chi router", Model: "claude-sonnet", Quality: 0.8, TaskType: "code", AgentID: "coder"},
		{Question: "deploy steps", Answer: "build then push", Model: "claude-haiku", Quality: 0.75, TaskType: "devops", AgentID: "ops"},
	}
	for _, e := range entries {
		if err := s.Store(e); err != nil {
			t.Fatal(err)
		}
	}

	stats, err := s.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 3 {
		t.Fatalf("expected total 3, got %d", stats.Total)
	}
	if stats.AvgQuality < 0.8 || stats.AvgQuality > 0.85 {
		t.Fatalf("unexpected avg quality: %f", stats.AvgQuality)
	}
	if stats.ByTaskType["code"] != 2 {
		t.Fatalf("expected 2 code entries, got %d", stats.ByTaskType["code"])
	}
	if stats.ByModel["claude-sonnet"] != 2 {
		t.Fatalf("expected 2 claude-sonnet, got %d", stats.ByModel["claude-sonnet"])
	}
}

// 10. Test empty store stats
func TestStatsEmptyStore(t *testing.T) {
	s := newTestStore(t)

	stats, err := s.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 0 {
		t.Fatalf("expected total 0, got %d", stats.Total)
	}
	if stats.AvgQuality != 0 {
		t.Fatalf("expected avg quality 0, got %f", stats.AvgQuality)
	}
}

// 11. Test search result ordering (quality DESC)
func TestSearchOrderByQuality(t *testing.T) {
	s := newTestStore(t)

	s.Store(qa.Entry{Question: "q1", Answer: "a1", Quality: 0.7, TaskType: "code", AgentID: "coder"})
	s.Store(qa.Entry{Question: "q2", Answer: "a2", Quality: 0.95, TaskType: "code", AgentID: "coder"})
	s.Store(qa.Entry{Question: "q3", Answer: "a3", Quality: 0.85, TaskType: "code", AgentID: "coder"})

	results, err := s.Search("q", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Quality < results[1].Quality || results[1].Quality < results[2].Quality {
		t.Fatalf("not sorted by quality: %v", []float64{results[0].Quality, results[1].Quality, results[2].Quality})
	}
}

// 12. Test default field values
func TestDefaultFields(t *testing.T) {
	s := newTestStore(t)

	entry := qa.Entry{
		Question: "minimal",
		Answer:   "entry",
		Quality:  0.8,
	}
	if err := s.Store(entry); err != nil {
		t.Fatal(err)
	}

	results, _ := s.Search("minimal", 10)
	if len(results) != 1 {
		t.Fatal("expected 1 result")
	}
	e := results[0]
	if e.ID == "" {
		t.Fatal("expected auto-generated ID")
	}
	if len(e.ID) != 32 {
		t.Fatalf("expected 32-char ID, got %d", len(e.ID))
	}
	if e.Model == "" {
		t.Fatal("expected default model")
	}
}

// 13. Test persistence (close and reopen)
func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "persist-qa.db")

	cfg := qa.Config{Path: dbPath, MinQuality: 0.7}
	s1, err := qa.NewStore(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := s1.Store(newEntry("persistent Q", "persistent A")); err != nil {
		t.Fatal(err)
	}
	s1.Close()

	s2, err := qa.NewStore(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	results, err := s2.Search("persistent Q", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result after reopen, got %d", len(results))
	}
}

// 14. Test search limit
func TestSearchLimit(t *testing.T) {
	s := newTestStore(t)

	for i := 0; i < 20; i++ {
		e := newEntry("limited q", "limited a")
		e.Quality = 0.8
		if err := s.Store(e); err != nil {
			t.Fatal(err)
		}
	}

	results, err := s.Search("limited", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}
}

// 15. Test filter by quality range
func TestFilterByQuality(t *testing.T) {
	s := newTestStore(t)

	s.Store(qa.Entry{Question: "low", Answer: "a", Quality: 0.7, TaskType: "code", AgentID: "coder"})
	s.Store(qa.Entry{Question: "mid", Answer: "a", Quality: 0.8, TaskType: "code", AgentID: "coder"})
	s.Store(qa.Entry{Question: "high", Answer: "a", Quality: 0.95, TaskType: "code", AgentID: "coder"})

	results, err := s.FilterByQuality(0.75, 0.9)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result in [0.75, 0.9], got %d", len(results))
	}
	if results[0].Question != "mid" {
		t.Fatalf("expected 'mid', got %s", results[0].Question)
	}
}

// 16. Test filter by agent
func TestFilterByAgent(t *testing.T) {
	s := newTestStore(t)

	s.Store(qa.Entry{Question: "q1", Answer: "a1", Quality: 0.8, TaskType: "code", AgentID: "coder"})
	s.Store(qa.Entry{Question: "q2", Answer: "a2", Quality: 0.8, TaskType: "code", AgentID: "reviewer"})
	s.Store(qa.Entry{Question: "q3", Answer: "a3", Quality: 0.8, TaskType: "code", AgentID: "coder"})

	results, err := s.FilterByAgent("coder")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 coder results, got %d", len(results))
	}
}

// 17. Test batch store
func TestBatchStore(t *testing.T) {
	s := newTestStore(t)

	entries := []qa.Entry{
		{Question: "batch 1", Answer: "a1", Quality: 0.8, TaskType: "code", AgentID: "coder"},
		{Question: "batch 2", Answer: "a2", Quality: 0.9, TaskType: "code", AgentID: "coder"},
		{Question: "batch 3", Answer: "a3", Quality: 0.85, TaskType: "code", AgentID: "coder"},
	}

	if err := s.BatchStore(entries); err != nil {
		t.Fatal(err)
	}

	n, _ := s.Count()
	if n != 3 {
		t.Fatalf("expected 3 entries after batch, got %d", n)
	}
}

// 18. Test batch store rejects low quality
func TestBatchStoreRejectsLowQuality(t *testing.T) {
	s := newTestStore(t)

	entries := []qa.Entry{
		{Question: "good", Answer: "a", Quality: 0.8, TaskType: "code", AgentID: "coder"},
		{Question: "bad", Answer: "a", Quality: 0.2, TaskType: "code", AgentID: "coder"},
	}

	err := s.BatchStore(entries)
	if err == nil {
		t.Fatal("expected error for low quality in batch")
	}

	n, _ := s.Count()
	if n != 0 {
		t.Fatalf("expected 0 entries, got %d", n)
	}
}

// 19. Test database path creation
func TestDatabasePathCreation(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "sub", "dir", "qa.db")

	cfg := qa.Config{Path: dbPath}
	s, err := qa.NewStore(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file not created")
	}
}

// 20. Test top used
func TestTopUsed(t *testing.T) {
	s := newTestStore(t)

	s.Store(qa.Entry{Question: "top1", Answer: "a", Quality: 0.8, TaskType: "code", AgentID: "coder"})
	s.Store(qa.Entry{Question: "top2", Answer: "a", Quality: 0.8, TaskType: "code", AgentID: "coder"})

	results, _ := s.Search("top1", 1)
	for i := 0; i < 10; i++ {
		s.IncrementUsed(results[0].ID)
	}
	results, _ = s.Search("top2", 1)
	for i := 0; i < 5; i++ {
		s.IncrementUsed(results[0].ID)
	}

	top, err := s.TopUsed(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(top) != 2 {
		t.Fatalf("expected 2 top entries, got %d", len(top))
	}
	if top[0].Question != "top1" {
		t.Fatalf("expected top1 first, got %s", top[0].Question)
	}
}
