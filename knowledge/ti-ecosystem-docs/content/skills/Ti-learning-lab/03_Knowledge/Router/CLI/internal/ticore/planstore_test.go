package ticore

import (
	"path/filepath"
	"testing"
	"time"
)

func TestPlanStoreSaveLoad(t *testing.T) {
	store, err := NewPlanStore(filepath.Join(t.TempDir(), "current-plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	plan := &ApprovedPlan{ID: "p1", Goal: "fix fallback", Project: "ti", TaskType: "bugfix", Summary: "summary", ApprovedAt: time.Now()}
	if err := store.Save(plan); err != nil {
		t.Fatal(err)
	}
	got, err := store.LoadCurrent()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.ID != plan.ID || got.Goal != plan.Goal {
		t.Fatalf("unexpected plan: %+v", got)
	}
	hist, err := store.History(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hist) != 1 {
		t.Fatalf("history=%d want 1", len(hist))
	}
}
