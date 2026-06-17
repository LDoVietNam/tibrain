package agent

import (
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"
)

func testAgent(id, name string, expertise, models []string) *Agent {
	defaultModel := ""
	if len(models) > 0 {
		defaultModel = models[0]
	}
	return &Agent{
		ID:           id,
		Name:         name,
		Description:  "test agent",
		Expertise:    expertise,
		Models:       models,
		DefaultModel: defaultModel,
		Config:       map[string]string{"key": "value"},
	}
}

// --- Registration validation ---

func TestRegisterEmptyID(t *testing.T) {
	r := NewRegistry()
	err := r.Register(testAgent("", "test", []string{"coding"}, []string{"gpt-4"}))
	if err != ErrInvalidID {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestRegisterEmptyName(t *testing.T) {
	r := NewRegistry()
	err := r.Register(testAgent("a1", "", []string{"coding"}, []string{"gpt-4"}))
	if err != ErrInvalidName {
		t.Fatalf("expected ErrInvalidName, got %v", err)
	}
}

func TestRegisterNoExpertise(t *testing.T) {
	r := NewRegistry()
	err := r.Register(testAgent("a1", "test", []string{}, []string{"gpt-4"}))
	if err != ErrNoExpertise {
		t.Fatalf("expected ErrNoExpertise, got %v", err)
	}
}

func TestRegisterNoModels(t *testing.T) {
	r := NewRegistry()
	err := r.Register(testAgent("a1", "test", []string{"coding"}, []string{}))
	if err != ErrNoModels {
		t.Fatalf("expected ErrNoModels, got %v", err)
	}
}

func TestRegisterSuccess(t *testing.T) {
	r := NewRegistry()
	a := testAgent("a1", "coder", []string{"coding", "review"}, []string{"gpt-4", "claude-sonnet"})
	if err := r.Register(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := r.Get("a1")
	if !ok {
		t.Fatal("agent not found after register")
	}
	if got.Status != StatusActive {
		t.Errorf("expected status=active, got %q", got.Status)
	}
	if got.LastSeen == 0 {
		t.Error("expected last_seen to be set")
	}
	if got.TaskCount != 0 {
		t.Errorf("expected task_count=0, got %d", got.TaskCount)
	}
	if got.AvgQuality != 0 {
		t.Errorf("expected avg_quality=0, got %f", got.AvgQuality)
	}
}

func TestRegisterDuplicateID(t *testing.T) {
	r := NewRegistry()
	a := testAgent("a1", "coder", []string{"coding"}, []string{"gpt-4"})
	if err := r.Register(a); err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	if err := r.Register(a); err == nil {
		t.Fatal("expected error on duplicate ID, got nil")
	}
}

// --- Deregister ---

func TestDeregister(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testAgent("a1", "coder", []string{"coding"}, []string{"gpt-4"}))

	if err := r.Deregister("a1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := r.Get("a1"); ok {
		t.Fatal("agent still exists after deregister")
	}
}

func TestDeregisterNotFound(t *testing.T) {
	r := NewRegistry()
	if err := r.Deregister("nonexistent"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeregisterEmptyID(t *testing.T) {
	r := NewRegistry()
	if err := r.Deregister(""); err != ErrInvalidID {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

// --- Get and List ---

func TestGetNotFound(t *testing.T) {
	r := NewRegistry()
	if _, ok := r.Get("nonexistent"); ok {
		t.Fatal("expected false for nonexistent agent")
	}
}

func TestListSortedByID(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testAgent("c3", "charlie", []string{"coding"}, []string{"gpt-4"}))
	_ = r.Register(testAgent("a1", "alice", []string{"review"}, []string{"claude"}))
	_ = r.Register(testAgent("b2", "bob", []string{"planning"}, []string{"gpt-4o"}))

	list := r.List()
	if len(list) != 3 {
		t.Fatalf("expected 3 agents, got %d", len(list))
	}
	if !sort.SliceIsSorted(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	}) {
		t.Fatalf("agents not sorted by ID: got %v", list)
	}
}

func TestListEmptyRegistry(t *testing.T) {
	r := NewRegistry()
	if len(r.List()) != 0 {
		t.Fatal("expected empty list")
	}
}

func TestGetReturnsDeepCopy(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testAgent("a1", "coder", []string{"coding"}, []string{"gpt-4"}))

	got, _ := r.Get("a1")
	got.Name = "hacked"
	got.Config["key"] = "hacked"

	orig, _ := r.Get("a1")
	if orig.Name == "hacked" {
		t.Fatal("Get did not return a deep copy — name was mutated")
	}
	if orig.Config["key"] == "hacked" {
		t.Fatal("Get did not return a deep copy — config was mutated")
	}
}

// --- FindByTaskType ---

func TestFindByTaskType(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testAgent("a1", "coder", []string{"coding", "review"}, []string{"gpt-4"}))
	_ = r.Register(testAgent("a2", "planner", []string{"planning"}, []string{"gpt-4o"}))
	_ = r.Register(testAgent("a3", "reviewer", []string{"review", "testing"}, []string{"claude"}))

	results := r.FindByTaskType("review")
	if len(results) != 2 {
		t.Fatalf("expected 2 agents with review expertise, got %d", len(results))
	}
	// Sorted by ID
	if results[0].ID != "a1" || results[1].ID != "a3" {
		t.Errorf("expected [a1, a3], got [%s, %s]", results[0].ID, results[1].ID)
	}
}

func TestFindByTaskTypeCaseInsensitive(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testAgent("a1", "coder", []string{"Coding"}, []string{"gpt-4"}))

	results := r.FindByTaskType("CODING")
	if len(results) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(results))
	}
	if results[0].ID != "a1" {
		t.Errorf("expected a1, got %s", results[0].ID)
	}
}

func TestFindByTaskTypeNoMatch(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testAgent("a1", "coder", []string{"coding"}, []string{"gpt-4"}))

	results := r.FindByTaskType("nonexistent")
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

// --- UpdateStatus ---

func TestUpdateStatus(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testAgent("a1", "coder", []string{"coding"}, []string{"gpt-4"}))

	if err := r.UpdateStatus("a1", StatusInactive); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := r.Get("a1")
	if got.Status != StatusInactive {
		t.Errorf("expected status=inactive, got %q", got.Status)
	}
}

func TestUpdateStatusNotFound(t *testing.T) {
	r := NewRegistry()
	if err := r.UpdateStatus("nonexistent", StatusError); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// --- RecordTask ---

func TestRecordTaskMovingAverage(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testAgent("a1", "coder", []string{"coding"}, []string{"gpt-4"}))

	r.RecordTask("a1", 0.8)
	a, _ := r.Get("a1")
	if a.TaskCount != 1 {
		t.Fatalf("expected task_count=1, got %d", a.TaskCount)
	}
	if a.AvgQuality != 0.8 {
		t.Fatalf("expected avg_quality=0.8, got %f", a.AvgQuality)
	}

	r.RecordTask("a1", 1.0)
	a, _ = r.Get("a1")
	if a.TaskCount != 2 {
		t.Fatalf("expected task_count=2, got %d", a.TaskCount)
	}
	expected := 0.9 // (0.8 + 1.0) / 2
	if a.AvgQuality != expected {
		t.Fatalf("expected avg_quality=0.9, got %f", a.AvgQuality)
	}
}

func TestRecordTaskNotFound(t *testing.T) {
	r := NewRegistry()
	// Should not panic
	r.RecordTask("nonexistent", 0.5)
}

// --- HealthCheck ---

func TestHealthCheckRecent(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testAgent("a1", "coder", []string{"coding"}, []string{"gpt-4"}))
	if !r.HealthCheck("a1") {
		t.Fatal("expected healthy agent")
	}
}

func TestHealthCheckStale(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(testAgent("a1", "coder", []string{"coding"}, []string{"gpt-4"}))

	// Manually set last_seen to 10 minutes ago
	r.mu.Lock()
	r.agents["a1"].LastSeen = time.Now().Add(-10 * time.Minute).Unix()
	r.mu.Unlock()

	if r.HealthCheck("a1") {
		t.Fatal("expected stale agent to be unhealthy")
	}
}

func TestHealthCheckNotFound(t *testing.T) {
	r := NewRegistry()
	if r.HealthCheck("nonexistent") {
		t.Fatal("expected false for nonexistent agent")
	}
}

// --- Count ---

func TestCount(t *testing.T) {
	r := NewRegistry()
	if r.Count() != 0 {
		t.Fatal("expected count=0")
	}
	_ = r.Register(testAgent("a1", "a", []string{"coding"}, []string{"gpt-4"}))
	_ = r.Register(testAgent("a2", "b", []string{"review"}, []string{"gpt-4"}))
	if r.Count() != 2 {
		t.Fatalf("expected count=2, got %d", r.Count())
	}
	_ = r.Deregister("a1")
	if r.Count() != 1 {
		t.Fatalf("expected count=1, got %d", r.Count())
	}
}

// --- Concurrency ---

func TestConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("a%d", i)
		_ = r.Register(testAgent(id, "agent", []string{"coding"}, []string{"gpt-4"}))
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.List()
			_ = r.Count()
			_ = r.FindByTaskType("coding")
			_, _ = r.Get("a0")
		}()
	}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = r.UpdateStatus(fmt.Sprintf("a%d", i), StatusInactive)
		}(i)
	}
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r.RecordTask(fmt.Sprintf("a%d", i), 0.8)
		}(i)
	}
	wg.Wait()

	if r.Count() != 100 {
		t.Fatalf("expected 100 agents, got %d", r.Count())
	}
}
