package optimize

import (
	"strings"
	"testing"

	"github.com/ti/cli/internal/ticore"
)

func TestBuildPrompt(t *testing.T) {
	out := BuildPrompt(Context{Plan: &ticore.ApprovedPlan{Goal: "cleanup"}, Step: &ticore.ApprovedStep{ID: "S2", Title: "Validate", Files: []string{"x.go"}}, VerifySummary: "tests passed"})
	if !strings.Contains(out, "tests passed") || !strings.Contains(out, "x.go") {
		t.Fatalf("unexpected prompt: %s", out)
	}
}
