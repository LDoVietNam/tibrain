package planner

import (
	"testing"

	"github.com/ti/cli/internal/permission"
	"github.com/ti/cli/internal/repoindex"
)

func TestBuildPlan(t *testing.T) {
	summary := &repoindex.Summary{Root: "/tmp/demo", Module: "example.com/demo", TestFileCount: 2, Files: []repoindex.FileInfo{{Path: "internal/router/router.go"}}}
	policy := permission.DefaultPolicy()
	plan := Build("fix router fallback", summary, policy)
	if plan == nil || len(plan.Steps) == 0 {
		t.Fatal("expected steps")
	}
	if plan.TaskType != "bugfix" {
		t.Fatalf("task type=%q", plan.TaskType)
	}
}
