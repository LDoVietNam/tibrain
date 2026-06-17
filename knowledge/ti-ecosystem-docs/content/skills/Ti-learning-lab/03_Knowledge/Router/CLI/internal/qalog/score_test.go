package qalog

import "testing"

func TestScoreRewardsSuccessfulTraceableEntries(t *testing.T) {
	entry := Entry{
		Project:        "ti",
		TaskType:       "code",
		Goal:           "fix fallback",
		ApprovedPlanID: "p1",
		Route:          "cliproxyapi",
		Model:          "claude-sonnet-4",
		AnswerSummary:  "Implemented the fallback guard, tightened route selection, and updated validation hints for the next verify step.",
		TotalTokens:    1800,
		Success:        true,
		Metadata:       map[string]string{"phase": "implement"},
	}
	score := Score(entry)
	if score.Total <= 0.65 {
		t.Fatalf("score=%f too low", score.Total)
	}
	if score.Components["success"] == 0 {
		t.Fatalf("missing success component: %+v", score.Components)
	}
}

func TestScoreRewardsVerificationPass(t *testing.T) {
	entry := Entry{
		Project:        "ti",
		TaskType:       "code",
		ApprovedPlanID: "p1",
		Route:          "cliproxyapi",
		Model:          "claude-sonnet-4",
		Success:        true,
		VerifyPassed:   true,
		VerifyCommands: 2,
		VerifyFailures: 0,
		Metadata:       map[string]string{"phase": "verify"},
	}
	score := Score(entry)
	if score.Components["verify_quality"] <= 0 {
		t.Fatalf("missing verify quality component: %+v", score.Components)
	}
}
