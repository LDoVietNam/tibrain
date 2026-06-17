package routeagent

import "testing"

func TestPromptPreferenceScore(t *testing.T) {
	s := promptPreferenceScore(8, 2, 0.9, 6)
	if s <= 0.7 {
		t.Fatalf("score too low: %v", s)
	}
}
