package router_test

import (
	"testing"

	"github.com/ti/cli/internal/router"
)

func TestValidPhases(t *testing.T) {
	phases := router.ValidPhases()
	if len(phases) != 6 {
		t.Errorf("expected 6 phases, got %d", len(phases))
	}
}

func TestIsValidPhase(t *testing.T) {
	valid := []string{"scan", "plan", "spec", "implement", "review", "summarize"}
	for _, p := range valid {
		if !router.IsValidPhase(p) {
			t.Errorf("expected %q to be valid", p)
		}
	}

	invalid := []string{"", "code", "test", "deploy", "SCAN", "Implement"}
	for _, p := range invalid {
		if router.IsValidPhase(p) {
			t.Errorf("expected %q to be invalid", p)
		}
	}
}

func TestSelectModelForPhase(t *testing.T) {
	tests := []struct {
		phase       string
		configModel string
		fallback    string
		want        string
	}{
		{"scan", "", "", "claude-sonnet-4"},
		{"plan", "", "", "claude-sonnet-4"},
		{"spec", "", "", "claude-sonnet-4"},
		{"implement", "", "", "claude-sonnet-4"},
		{"review", "", "", "claude-sonnet-4"},
		{"summarize", "", "", "claude-haiku-4"},

		// Config override
		{"implement", "gpt-4o", "", "gpt-4o"},

		// Config > phase > fallback (phase wins over fallback)
		{"scan", "gpt-4o", "qwen-32b", "gpt-4o"},
		{"scan", "", "qwen-32b", "claude-sonnet-4"}, // phase model wins

		// Unknown phase → fallback
		{"unknown", "", "my-fallback", "my-fallback"},
		{"unknown", "", "", "claude-sonnet-4"}, // default fallback
	}

	for _, tt := range tests {
		got := router.SelectModelForPhase(tt.phase, tt.configModel, tt.fallback)
		if got != tt.want {
			t.Errorf("SelectModelForPhase(%q, %q, %q) = %q, want %q",
				tt.phase, tt.configModel, tt.fallback, got, tt.want)
		}
	}
}

func TestSelectBudgetForPhase(t *testing.T) {
	tests := []struct {
		phase     string
		maxBudget float64
		want      float64
	}{
		{"scan", 0, 0.50},
		{"plan", 0, 1.00},
		{"spec", 0, 0.50},
		{"implement", 0, 2.00},
		{"review", 0, 0.50},
		{"summarize", 0, 0.10},

		// maxBudget override
		{"implement", 0.50, 0.50},
		{"scan", 5.0, 5.0},

		// Unknown phase → default $1
		{"unknown", 0, 1.0},
	}

	for _, tt := range tests {
		got := router.SelectBudgetForPhase(tt.phase, tt.maxBudget)
		if got != tt.want {
			t.Errorf("SelectBudgetForPhase(%q, %.2f) = %.2f, want %.2f",
				tt.phase, tt.maxBudget, got, tt.want)
		}
	}
}
