package ticore

import (
	"strings"
	"testing"

	"github.com/ti/cli/internal/planner"
	"github.com/ti/cli/internal/repoindex"
)

func TestApprovePlanAndPrepareRuntime(t *testing.T) {
	plan := &planner.Plan{Goal: "fix router fallback", TaskType: "code", Project: "ti", CandidateFiles: []string{"main.go", "internal/router/router.go"}, Validation: []string{"go test ./..."}, Steps: []planner.Step{{ID: "S1", Title: "Inspect"}, {ID: "S2", Title: "Patch"}}}
	orch := New("ti", nil)
	approved := orch.ApprovePlan(plan)
	if approved == nil || approved.ID == "" {
		t.Fatalf("expected approved plan id")
	}
	env := orch.PrepareRuntime("fix router fallback", "implement", "cliproxyapi", "medium", approved, &repoindex.Summary{Module: "ti", EntryPoints: []string{"main.go"}})
	if env.Prompt == nil || env.Context == nil {
		t.Fatalf("expected prompt and context")
	}
	if !strings.Contains(env.Prompt.UserPrompt, "fix router fallback") {
		t.Fatalf("user prompt missing goal")
	}
}
