package taskinput

import (
	"strings"
	"testing"

	"github.com/ti/cli/internal/repoindex"
	"github.com/ti/cli/internal/ticore"
)

func TestBuildUsesStepAndPreferredShape(t *testing.T) {
	plan := &ticore.ApprovedPlan{
		ID:             "plan-1",
		Goal:           "implement routing",
		TaskType:       "code",
		Project:        "ti",
		Summary:        "Implement routing improvements",
		CandidateFiles: []string{"main.go", "internal/router/router.go"},
		Validation:     []string{"go test ./..."},
		Steps: []ticore.ApprovedStep{{
			ID: "s1", Title: "Update router", Objective: "Improve route selection", Files: []string{"internal/router/router.go"}, Commands: []string{"go test ./internal/router"}, ExitCriteria: []string{"router tests pass"}, Status: ticore.StepPending,
		}},
	}
	step := &plan.Steps[0]
	summary := &repoindex.Summary{Module: "ti", EntryPoints: []string{"main.go"}, TopDirectories: []string{"internal/router", "internal/ticore"}, TestFileCount: 10}
	built := Build(Options{Phase: "implement", Route: "cliproxyapi", RequestedModel: "claude-sonnet-4", SelectionMode: "preferred", ApprovedPlan: plan, Step: step, Summary: summary, PreferredPromptShape: "structured-execution"})
	if built.TaskType != "code" {
		t.Fatalf("task type = %q", built.TaskType)
	}
	if built.CurrentStepID != "s1" {
		t.Fatalf("step id = %q", built.CurrentStepID)
	}
	if built.Bundle == nil || built.Prompt == nil {
		t.Fatal("expected bundle and prompt")
	}
	if built.Prompt.PromptShape != "structured-execution" {
		t.Fatalf("prompt shape = %q", built.Prompt.PromptShape)
	}
	if !strings.Contains(built.ExecutionGoal, "Execute approved Ti plan step s1") {
		t.Fatalf("execution goal missing step context: %q", built.ExecutionGoal)
	}
	if built.RequestedModel != "claude-sonnet-4" || built.ModelSelectionMode != "preferred" {
		t.Fatalf("model fields not carried through: %+v", built)
	}
}

func TestBuildFallsBackToGoal(t *testing.T) {
	built := Build(Options{Goal: "say hello", Phase: "design", Route: "sharedchat"})
	if built.Goal != "say hello" {
		t.Fatalf("goal = %q", built.Goal)
	}
	if built.TaskType != "plan" {
		t.Fatalf("task type = %q", built.TaskType)
	}
	if built.Prompt == nil || built.Prompt.SystemPrompt == "" || built.Prompt.UserPrompt == "" {
		t.Fatal("missing transformed prompt")
	}
}
