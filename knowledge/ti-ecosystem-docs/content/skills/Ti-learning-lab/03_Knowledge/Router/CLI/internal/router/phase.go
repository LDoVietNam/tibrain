package router

// Phase routing: maps workflow phases to model selection strategies.
// Each phase has different requirements:
//
//   scan       → large context, cheap model (summarize codebase)
//   plan       → reasoning model (architecture, strategy)
//   spec       → reasoning model (detailed specifications)
//   implement  → code model (actual coding)
//   review     → review model (code review, quality check)
//   summarize  → cheap model (condense, summarize)

// Valid phases.
const (
	PhaseScan      = "scan"
	PhasePlan      = "plan"
	PhaseSpec      = "spec"
	PhaseImplement = "implement"
	PhaseReview    = "review"
	PhaseSummarize = "summarize"
)

// PhaseModel maps each phase to its recommended model.
// These are defaults — can be overridden by config or CLI flags.
var PhaseModel = map[string]string{
	PhaseScan:      "claude-sonnet-4", // large context understanding
	PhasePlan:      "claude-sonnet-4", // reasoning for architecture
	PhaseSpec:      "claude-sonnet-4", // detailed specifications
	PhaseImplement: "claude-sonnet-4", // code generation
	PhaseReview:    "claude-sonnet-4", // code review
	PhaseSummarize: "claude-haiku-4",  // cheap summarization
}

// PhaseBudget maps each phase to its default budget (USD).
var PhaseBudget = map[string]float64{
	PhaseScan:      0.50,
	PhasePlan:      1.00,
	PhaseSpec:      0.50,
	PhaseImplement: 2.00,
	PhaseReview:    0.50,
	PhaseSummarize: 0.10,
}

// ValidPhases returns all valid phase names.
func ValidPhases() []string {
	return []string{PhaseScan, PhasePlan, PhaseSpec, PhaseImplement, PhaseReview, PhaseSummarize}
}

// IsValidPhase checks if a phase string is valid.
func IsValidPhase(phase string) bool {
	_, ok := PhaseModel[phase]
	return ok
}

// SelectModelForPhase returns the recommended model for a phase.
// If config overrides exist, they take precedence.
func SelectModelForPhase(phase string, configModel, fallbackModel string) string {
	// CLI/config override takes highest precedence
	if configModel != "" {
		return configModel
	}

	// Phase-based selection
	if model, ok := PhaseModel[phase]; ok {
		return model
	}

	// Default: use config default or fallback
	if fallbackModel != "" {
		return fallbackModel
	}
	return PhaseModel[PhaseImplement]
}

// SelectBudgetForPhase returns the budget for a phase.
// If maxBudget is set (>0), it takes precedence.
func SelectBudgetForPhase(phase string, maxBudget float64) float64 {
	if maxBudget > 0 {
		return maxBudget
	}
	if budget, ok := PhaseBudget[phase]; ok {
		return budget
	}
	return 1.0 // Default $1 budget
}
