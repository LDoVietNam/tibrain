package planner

import (
	"strings"
	"testing"

	"github.com/ti/cli/internal/permission"
	"github.com/ti/cli/internal/repoindex"
)

func TestRunCouncilApprovesPlan(t *testing.T) {
	policy := &permission.Policy{Mode: permission.ModeAsk}
	summary := &repoindex.Summary{Module: "ti", Files: []repoindex.FileInfo{{Path: "internal/router/router.go", Kind: "go"}, {Path: "main.go", Kind: "go"}}, EntryPoints: []string{"main.go"}}
	result := RunCouncil("fix router retry bug", summary, policy)
	if result == nil || result.Approved == nil {
		t.Fatalf("expected approved plan")
	}
	if result.Approved.Summary == "" {
		t.Fatalf("expected summary")
	}
	rendered := RenderApprovedMarkdown(result)
	if !strings.Contains(rendered, "# Ti Approved Plan") {
		t.Fatalf("unexpected markdown: %s", rendered)
	}
}
