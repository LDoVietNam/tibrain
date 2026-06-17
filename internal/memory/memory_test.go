package memory

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

const memoriesSchema = `
CREATE TABLE memories (
  id TEXT PRIMARY KEY,
  api_key_id TEXT NOT NULL DEFAULT 'local',
  session_id TEXT,
  type TEXT NOT NULL CHECK(type IN ('factual', 'episodic', 'procedural', 'semantic', 'working', 'long_term')),
  key TEXT,
  content TEXT NOT NULL,
  metadata TEXT,
  importance REAL NOT NULL DEFAULT 0.5,
  access_count INTEGER NOT NULL DEFAULT 0,
  last_access TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  expires_at TEXT
);
CREATE INDEX idx_memories_type ON memories(type);
CREATE INDEX idx_memories_expires ON memories(expires_at);
CREATE INDEX idx_memories_importance ON memories(importance);
`

// newTestDB returns an in-memory SQLite database with the memories schema
// already applied. Satisfies the DB interface used by CognitiveMemoryManager.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open :memory: db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(memoriesSchema); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	return db
}

func newTestCM(t *testing.T) *CognitiveMemoryManager {
	t.Helper()
	return NewCognitiveMemoryManager(newTestDB(t), nil)
}

// ---------- Store tests ----------

func TestStoreEpisodicMemory(t *testing.T) {
	cm := newTestCM(t)
	ctx := context.Background()

	id, err := cm.StoreEpisodicMemory(ctx, "user asked about weather", map[string]interface{}{"location": "hanoi"})
	if err != nil {
		t.Fatalf("StoreEpisodicMemory: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty ID")
	}

	entries, err := cm.QueryMemory(ctx, "", MemoryEpisodic, 10)
	if err != nil {
		t.Fatalf("QueryMemory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 episodic, got %d", len(entries))
	}
	if entries[0].ID != id {
		t.Errorf("ID mismatch: got %s want %s", entries[0].ID, id)
	}
	if entries[0].Type != MemoryEpisodic {
		t.Errorf("Type mismatch: got %s want episodic", entries[0].Type)
	}
	if entries[0].Importance != 0.7 {
		t.Errorf("Importance: got %v want 0.7", entries[0].Importance)
	}
	if entries[0].Source != "agent_interaction" {
		t.Errorf("Source: got %s", entries[0].Source)
	}
}

func TestStoreSemanticMemoryWithTags(t *testing.T) {
	cm := newTestCM(t)
	ctx := context.Background()

	_, err := cm.StoreSemanticMemory(ctx, "Go is statically typed", []string{"go", "programming"}, "wikipedia")
	if err != nil {
		t.Fatalf("StoreSemanticMemory: %v", err)
	}

	entries, err := cm.QueryMemory(ctx, "", MemorySemantic, 10)
	if err != nil {
		t.Fatalf("QueryMemory: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 semantic, got %d", len(entries))
	}
	if !stringSliceContains(entries[0].Tags, "go") || !stringSliceContains(entries[0].Tags, "programming") {
		t.Errorf("tags not preserved: %v", entries[0].Tags)
	}
	if entries[0].Source != "wikipedia" {
		t.Errorf("Source: %s", entries[0].Source)
	}
}

func TestStoreProceduralMemory(t *testing.T) {
	cm := newTestCM(t)
	ctx := context.Background()

	_, err := cm.StoreProceduralMemory(ctx, "build then test", map[string]interface{}{"step": 1})
	if err != nil {
		t.Fatalf("StoreProceduralMemory: %v", err)
	}

	entries, err := cm.QueryMemory(ctx, "", MemoryProcedural, 10)
	if err != nil {
		t.Fatalf("QueryMemory: %v", err)
	}
	if len(entries) != 1 || entries[0].Type != MemoryProcedural {
		t.Fatalf("expected 1 procedural, got %d", len(entries))
	}
	if entries[0].Importance != 0.9 {
		t.Errorf("procedural importance: got %v want 0.9", entries[0].Importance)
	}
}

func TestStoreWorkingMemoryRejected(t *testing.T) {
	cm := newTestCM(t)
	_, err := cm.store(context.Background(), MemoryWorking, "", "x", nil, 0.5, nil, "test", 0)
	if err == nil {
		t.Fatal("expected error storing working memory via store()")
	}
}

// ---------- QueryMemory tests (CRITICAL — bug fix) ----------

func TestQueryMemoryWithTextFilter(t *testing.T) {
	cm := newTestCM(t)
	ctx := context.Background()

	for _, content := range []string{
		"Go is statically typed",
		"Python is dynamically typed",
		"Rust has a borrow checker",
	} {
		if _, err := cm.StoreSemanticMemory(ctx, content, nil, ""); err != nil {
			t.Fatalf("StoreSemanticMemory(%q): %v", content, err)
		}
	}

	// Filter for "typed" — should match Go and Python, not Rust.
	entries, err := cm.QueryMemory(ctx, "typed", MemorySemantic, 10)
	if err != nil {
		t.Fatalf("QueryMemory: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 results matching 'typed', got %d", len(entries))
	}
	for _, e := range entries {
		if !strings.Contains(strings.ToLower(e.Content), "typed") {
			t.Errorf("result doesn't contain query: %s", e.Content)
		}
	}

	// Filter for "borrow" — should match only Rust.
	entries, err = cm.QueryMemory(ctx, "borrow", MemorySemantic, 10)
	if err != nil {
		t.Fatalf("QueryMemory: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 result matching 'borrow', got %d", len(entries))
	}
	if len(entries) > 0 && !strings.Contains(entries[0].Content, "Rust") {
		t.Errorf("expected Rust, got %s", entries[0].Content)
	}
}

func TestQueryMemoryEmptyFilterReturnsTopByImportance(t *testing.T) {
	cm := newTestCM(t)
	ctx := context.Background()

	cm.StoreSemanticMemory(ctx, "low importance", nil, "")
	cm.StoreEpisodicMemory(ctx, "episodic only", nil)

	entries, err := cm.QueryMemory(ctx, "", MemorySemantic, 10)
	if err != nil {
		t.Fatalf("QueryMemory: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected only the semantic one, got %d", len(entries))
	}
}

func TestQueryMemoryExcludesExpired(t *testing.T) {
	cm := newTestCM(t)
	ctx := context.Background()

	// Short TTL — 50ms.
	_, err := cm.StoreMemoryWithTTL(ctx, MemorySemantic, "ephemeral note", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("StoreMemoryWithTTL: %v", err)
	}
	// Long-lived.
	cm.StoreSemanticMemory(ctx, "permanent fact", nil, "")

	// Immediately, both visible.
	entries, _ := cm.QueryMemory(ctx, "", MemorySemantic, 10)
	if len(entries) != 2 {
		t.Errorf("expected 2 before expiry, got %d", len(entries))
	}

	// After 100ms, only permanent remains.
	time.Sleep(100 * time.Millisecond)
	entries, _ = cm.QueryMemory(ctx, "", MemorySemantic, 10)
	if len(entries) != 1 {
		t.Errorf("expected 1 after expiry, got %d", len(entries))
	}
	if entries[0].Content != "permanent fact" {
		t.Errorf("wrong survivor: %s", entries[0].Content)
	}
}

// ---------- StrengthenMemory tests ----------

func TestStrengthenMemory(t *testing.T) {
	cm := newTestCM(t)
	ctx := context.Background()

	id, _ := cm.StoreEpisodicMemory(ctx, "strengthen me", nil)

	if err := cm.StrengthenMemory(ctx, id); err != nil {
		t.Fatalf("StrengthenMemory: %v", err)
	}
	if err := cm.StrengthenMemory(ctx, id); err != nil {
		t.Fatalf("StrengthenMemory 2nd: %v", err)
	}

	entries, _ := cm.QueryMemory(ctx, "", MemoryEpisodic, 10)
	if entries[0].AccessCount != 3 {
		t.Errorf("AccessCount: got %d want 3 (1 initial + 2 strengthens)", entries[0].AccessCount)
	}
}

func TestStrengthenMemoryNotFound(t *testing.T) {
	cm := newTestCM(t)
	err := cm.StrengthenMemory(context.Background(), "nonexistent-id")
	if err != ErrMemoryNotFound {
		t.Errorf("expected ErrMemoryNotFound, got %v", err)
	}
}

// ---------- ConsolidateMemory tests ----------

func TestConsolidateMemoryPromotesHighAccess(t *testing.T) {
	cm := newTestCM(t)
	ctx := context.Background()

	id1, _ := cm.StoreEpisodicMemory(ctx, "frequently accessed", nil)
	cm.StoreEpisodicMemory(ctx, "barely touched", nil)

	// Strengthen id1 past threshold of 3.
	for i := 0; i < 4; i++ {
		cm.StrengthenMemory(ctx, id1)
	}

	promoted, err := cm.ConsolidateMemory(ctx, 3)
	if err != nil {
		t.Fatalf("ConsolidateMemory: %v", err)
	}
	if promoted != 1 {
		t.Errorf("expected 1 promoted, got %d", promoted)
	}

	// id1 should now be long_term.
	entries, _ := cm.QueryMemory(ctx, "", MemoryLongTerm, 10)
	if len(entries) != 1 || entries[0].ID != id1 {
		t.Errorf("id1 not promoted: got %+v", entries)
	}
	// The other should remain episodic.
	entries, _ = cm.QueryMemory(ctx, "", MemoryEpisodic, 10)
	if len(entries) != 1 || entries[0].ID == id1 {
		t.Errorf("barely touched was wrongly affected: %+v", entries)
	}
}

// ---------- ForgetExpired tests ----------

func TestForgetExpired(t *testing.T) {
	cm := newTestCM(t)
	ctx := context.Background()

	cm.StoreMemoryWithTTL(ctx, MemorySemantic, "ttl1", 50*time.Millisecond)
	cm.StoreMemoryWithTTL(ctx, MemorySemantic, "ttl2", 50*time.Millisecond)
	cm.StoreSemanticMemory(ctx, "permanent", nil, "") // no TTL

	time.Sleep(100 * time.Millisecond)
	deleted, err := cm.ForgetExpired(ctx)
	if err != nil {
		t.Fatalf("ForgetExpired: %v", err)
	}
	if deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", deleted)
	}

	entries, _ := cm.QueryMemory(ctx, "", MemorySemantic, 10)
	if len(entries) != 1 || entries[0].Content != "permanent" {
		t.Errorf("unexpected survivors: %+v", entries)
	}
}

// ---------- GetMemoryStats tests ----------

func TestGetMemoryStatsSingleQuery(t *testing.T) {
	cm := newTestCM(t)
	ctx := context.Background()

	cm.StoreEpisodicMemory(ctx, "e1", nil)
	cm.StoreEpisodicMemory(ctx, "e2", nil)
	cm.StoreSemanticMemory(ctx, "s1", nil, "")
	cm.StoreEpisodicMemory(ctx, "e3", nil)

	stats, err := cm.GetMemoryStats(ctx)
	if err != nil {
		t.Fatalf("GetMemoryStats: %v", err)
	}

	if stats["total"].(int) != 4 {
		t.Errorf("total: got %v want 4", stats["total"])
	}
	if stats[string(MemoryEpisodic)].(int) != 3 {
		t.Errorf("episodic: got %v want 3", stats[string(MemoryEpisodic)])
	}
	if stats[string(MemorySemantic)].(int) != 1 {
		t.Errorf("semantic: got %v want 1", stats[string(MemorySemantic)])
	}
	if stats["working_memory_keys"].(int) != 0 {
		t.Errorf("working_memory_keys should be 0 initially, got %v", stats["working_memory_keys"])
	}
}

// ---------- Working memory tests ----------

func TestWorkingMemoryCRUD(t *testing.T) {
	cm := newTestCM(t)

	cm.UpdateWorkingMemory("user_id", "min")
	cm.UpdateWorkingMemory("context", "test")

	if v, ok := cm.GetWorkingMemory("user_id"); !ok || v != "min" {
		t.Errorf("user_id: got %v ok=%v", v, ok)
	}
	if _, ok := cm.GetWorkingMemory("missing"); ok {
		t.Error("missing key should return ok=false")
	}

	cm.ClearWorkingMemory()
	if v, ok := cm.GetWorkingMemory("user_id"); ok {
		t.Errorf("after Clear, key still present: %v", v)
	}
}

// ---------- Convenience methods ----------

func TestGetRecentExperience(t *testing.T) {
	cm := newTestCM(t)
	ctx := context.Background()

	cm.StoreSemanticMemory(ctx, "semantic 1", nil, "") // should not appear
	cm.StoreEpisodicMemory(ctx, "episodic 1", nil)
	cm.StoreEpisodicMemory(ctx, "episodic 2", nil)

	entries, err := cm.GetRecentExperience(ctx, 10)
	if err != nil {
		t.Fatalf("GetRecentExperience: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 episodic, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Type != MemoryEpisodic {
			t.Errorf("non-episodic leaked: %s", e.Type)
		}
	}
}

func TestGetRelevantKnowledgeUsesFilter(t *testing.T) {
	cm := newTestCM(t)
	ctx := context.Background()

	cm.StoreSemanticMemory(ctx, "Go has goroutines", nil, "")
	cm.StoreSemanticMemory(ctx, "Python has asyncio", nil, "")
	cm.StoreSemanticMemory(ctx, "Rust has tokio", nil, "")

	entries, err := cm.GetRelevantKnowledge(ctx, "goroutines", 10)
	if err != nil {
		t.Fatalf("GetRelevantKnowledge: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 match for 'goroutines', got %d", len(entries))
	}
}

// ---------- Helpers ----------

func stringSliceContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
