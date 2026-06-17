package repair

import (
	"strings"
	"testing"

	"github.com/ti/cli/internal/ticore"
)

func TestBuildPrompt(t *testing.T) {
	out := BuildPrompt(Context{Plan: &ticore.ApprovedPlan{Goal: "fix verify"}, Step: &ticore.ApprovedStep{ID: "S1", Title: "Implement", Files: []string{"a.go"}}, FailureSummary: "go test failed"})
	if !strings.Contains(out, "go test failed") || !strings.Contains(out, "a.go") {
		t.Fatalf("prompt missing context: %s", out)
	}
}
