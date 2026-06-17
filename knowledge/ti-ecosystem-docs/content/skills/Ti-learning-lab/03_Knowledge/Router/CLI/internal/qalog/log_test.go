package qalog

import (
	"path/filepath"
	"testing"
)

func TestStoreAppendStatsAndLatestExecution(t *testing.T) {
	s, err := NewStore(filepath.Join(t.TempDir(), "qa.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	_ = s.Append(Entry{Project: "ti", TaskType: "code", ApprovedPlanID: "p1", Route: "cliproxyapi", PromptShape: "structured-execution", Success: true, TotalTokens: 42, Metadata: map[string]string{"phase": "implement", "step_id": "S1"}})
	_ = s.Append(Entry{Project: "ti", TaskType: "plan", Route: "sharedchat", Success: false})
	st, err := s.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if st.Total != 2 || st.Successes != 1 || st.Failures != 1 {
		t.Fatalf("unexpected stats: %+v", st)
	}
	if st.ByRoute["cliproxyapi"] != 1 || st.ByTaskType["code"] != 1 {
		t.Fatalf("unexpected breakdown: %+v", st)
	}
	if st.ByPromptShape["structured-execution"] != 1 {
		t.Fatalf("expected prompt shape stats: %+v", st.ByPromptShape)
	}
	latest, err := s.LatestExecutionForStep("p1", "S1")
	if err != nil {
		t.Fatal(err)
	}
	if latest == nil || latest.Route != "cliproxyapi" {
		t.Fatalf("latest=%+v want cliproxyapi", latest)
	}
}
