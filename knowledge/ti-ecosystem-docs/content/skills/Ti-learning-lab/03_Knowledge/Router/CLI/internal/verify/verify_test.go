package verify

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ti/cli/internal/permission"
)

func TestRunnerRunAndSummarize(t *testing.T) {
	r := NewRunner(t.TempDir(), permission.DefaultPolicy(), 5*time.Second)
	cmd := "printf ok"
	if runtime.GOOS == "windows" {
		cmd = "echo ok"
	}
	results, err := r.Run([]string{cmd})
	if err != nil {
		t.Fatal(err)
	}
	sum := Summarize(results)
	if sum.Commands != 1 || sum.Passed != 1 || sum.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
}

func TestRunnerDeniedByPolicy(t *testing.T) {
	p := permission.DefaultPolicy()
	p.Mode = permission.ModeDeny
	r := NewRunner(t.TempDir(), p, 5*time.Second)
	_, err := r.Run([]string{"printf denied"})
	if err == nil {
		t.Fatal("expected deny error")
	}
}

func TestRenderHumanAndJSON(t *testing.T) {
	report := BuildReport(ReportMeta{
		PlanID: "p1", PlanState: "executing", StepID: "S1", StepTitle: "Patch", NextStepID: "S2", NextStepTitle: "Verify", Note: "verify commands=1 passed=1 failed=0",
	}, Summary{Commands: 1, Passed: 1, Failed: 0}, []Result{{Command: "go test ./...", Passed: true, Output: "ok", Duration: time.Second}}, nil)
	human := RenderHuman(report)
	if !strings.Contains(human, "Next step: S2") {
		t.Fatalf("unexpected human report: %s", human)
	}
	jsonText, err := RenderJSON(report)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(jsonText, `"step_id": "S1"`) {
		t.Fatalf("unexpected json report: %s", jsonText)
	}
}
