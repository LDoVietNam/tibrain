package ticore

import "testing"

func TestApprovedPlanBlocksNextActUntilVerify(t *testing.T) {
	plan := &ApprovedPlan{
		ID:    "p1",
		Goal:  "fix fallback",
		State: PlanStateApproved,
		Steps: []ApprovedStep{
			{ID: "S1", Title: "Inspect", Status: StepPending},
			{ID: "S2", Title: "Patch", Status: StepPending},
		},
	}
	step := plan.SelectNextStep("")
	if step == nil || step.ID != "S1" {
		t.Fatalf("got step %+v want S1", step)
	}
	if !plan.MarkStepStarted("S1", "begin") {
		t.Fatal("expected step start")
	}
	if !plan.MarkStepImplemented("S1", "done") {
		t.Fatal("expected implemented")
	}
	if plan.PendingVerificationStep() == nil || plan.PendingVerificationStep().ID != "S1" {
		t.Fatalf("expected pending verification on S1")
	}
	if next := plan.SelectNextStep(""); next != nil {
		t.Fatalf("expected no next step before verify, got %+v", next)
	}
}

func TestApprovedPlanAdvancesAfterCompletion(t *testing.T) {
	plan := &ApprovedPlan{
		ID:    "p1",
		Goal:  "fix fallback",
		State: PlanStateApproved,
		Steps: []ApprovedStep{
			{ID: "S1", Title: "Inspect", Status: StepPending},
			{ID: "S2", Title: "Patch", Status: StepPending},
		},
	}
	if !plan.MarkStepStarted("S1", "begin") {
		t.Fatal("expected step start")
	}
	if !plan.MarkStepImplemented("S1", "done") {
		t.Fatal("expected implemented")
	}
	if !plan.MarkStepVerified("S1", "verified") {
		t.Fatal("expected verified")
	}
	if !plan.MarkStepCompleted("S1", "completed") {
		t.Fatal("expected completed")
	}
	if plan.State != PlanStateExecuting {
		t.Fatalf("state=%s want %s", plan.State, PlanStateExecuting)
	}
	if plan.CurrentStepID != "S2" {
		t.Fatalf("current=%s want S2", plan.CurrentStepID)
	}
	next := plan.SelectNextStep("")
	if next == nil || next.ID != "S2" {
		t.Fatalf("next=%+v want S2", next)
	}
}
