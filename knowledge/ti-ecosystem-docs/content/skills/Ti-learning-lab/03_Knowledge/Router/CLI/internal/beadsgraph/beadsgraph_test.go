package beadsgraph

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ti/cli/internal/ticore"
)

func samplePlan() *ticore.ApprovedPlan {
	now := time.Now()
	return &ticore.ApprovedPlan{
		ID:            "plan-1",
		Goal:          "implement slice 9",
		Project:       "ti",
		Summary:       "approved slice 9 plan",
		ApprovedAt:    now,
		TaskType:      "feature",
		RouteHints:    []string{"planning-council", "quality-first-routing"},
		State:         ticore.PlanStateApproved,
		CurrentStepID: "s1",
		Steps: []ticore.ApprovedStep{
			{ID: "s1", Title: "Create graph", Objective: "Map plan to beads", ExitCriteria: []string{"graph saved"}, Status: ticore.StepPending},
			{ID: "s2", Title: "Sync graph", Objective: "Update from execution", ExitCriteria: []string{"sync works"}, Status: ticore.StepPending},
		},
	}
}

func TestFromApprovedPlan(t *testing.T) {
	g := FromApprovedPlan(samplePlan())
	if g == nil {
		t.Fatal("nil graph")
	}
	if len(g.Epics) != 1 || len(g.Issues) != 2 {
		t.Fatalf("got %d epics %d issues", len(g.Epics), len(g.Issues))
	}
	if g.Issues[0].Status != Status("ready") {
		t.Fatalf("first issue not ready: %s", g.Issues[0].Status)
	}
	if len(g.Issues[1].DependsOn) != 1 {
		t.Fatalf("second issue should depend on first")
	}
}

func TestReviewAndReady(t *testing.T) {
	g := FromApprovedPlan(samplePlan())
	report := Review(g)
	if report.ReadyCount != 1 {
		t.Fatalf("ready count=%d want 1", report.ReadyCount)
	}
	if len(report.MissingAcceptance) != 0 {
		t.Fatalf("unexpected missing acceptance")
	}
}

func TestSyncWithApprovedPlan(t *testing.T) {
	plan := samplePlan()
	g := FromApprovedPlan(plan)
	plan.Steps[0].Status = ticore.StepCompleted
	plan.Steps[1].Status = ticore.StepPending
	rep := SyncWithApprovedPlan(g, plan)
	if len(rep.CompletedIssues) != 1 {
		t.Fatalf("completed issues=%d", len(rep.CompletedIssues))
	}
	ready := ReadyIssues(g)
	if len(ready) != 1 || ready[0].StepID != "s2" {
		t.Fatalf("unexpected ready issues: %+v", ready)
	}
}

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "beads-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	g := FromApprovedPlan(samplePlan())
	if err := store.Save(g); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadCurrent()
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil || loaded.PlanID != "plan-1" {
		t.Fatalf("unexpected loaded graph: %+v", loaded)
	}
}
