package qalog

import "strings"

type Scorecard struct {
	Total      float64            `json:"total"`
	Components map[string]float64 `json:"components,omitempty"`
}

func Score(entry Entry) Scorecard {
	components := map[string]float64{}
	add := func(name string, value float64) {
		if value < 0 {
			value = 0
		}
		components[name] = value
	}

	if entry.Success {
		add("success", 0.32)
	} else {
		add("success", 0.06)
	}
	if strings.TrimSpace(entry.AnswerSummary) != "" {
		summaryLen := len(strings.TrimSpace(entry.AnswerSummary))
		switch {
		case summaryLen >= 180:
			add("answer_substance", 0.14)
		case summaryLen >= 80:
			add("answer_substance", 0.10)
		default:
			add("answer_substance", 0.05)
		}
	}
	switch {
	case entry.TotalTokens == 0:
		add("token_efficiency", 0.04)
	case entry.TotalTokens <= 3000:
		add("token_efficiency", 0.12)
	case entry.TotalTokens <= 8000:
		add("token_efficiency", 0.08)
	default:
		add("token_efficiency", 0.03)
	}
	if entry.ErrorClass == "" {
		add("error_hygiene", 0.08)
	} else {
		add("error_hygiene", 0.01)
	}
	if strings.TrimSpace(entry.Route) != "" {
		add("route_recorded", 0.05)
	}
	if strings.TrimSpace(entry.Model) != "" {
		add("model_recorded", 0.05)
	}
	if strings.TrimSpace(entry.TaskType) != "" {
		add("task_typed", 0.04)
	}
	if strings.TrimSpace(entry.ApprovedPlanID) != "" {
		add("plan_traceability", 0.04)
	}
	if entry.Metadata != nil && strings.TrimSpace(entry.Metadata["phase"]) != "" {
		add("phase_recorded", 0.03)
	}
	if entry.VerifyCommands > 0 {
		if entry.VerifyPassed && entry.VerifyFailures == 0 {
			add("verify_quality", 0.18)
		} else {
			add("verify_quality", 0.02)
		}
		add("verify_traceability", 0.03)
	}

	total := 0.0
	for _, v := range components {
		total += v
	}
	if total > 1.0 {
		total = 1.0
	}
	return Scorecard{Total: total, Components: components}
}

func PreferenceSignal(entry Entry) (quality float64, preferred bool) {
	score := entry.Quality
	if score == 0 {
		score = Score(entry).Total
	}
	preferred = entry.Success
	if entry.VerifyCommands > 0 {
		preferred = entry.VerifyPassed && entry.VerifyFailures == 0 && score >= 0.70
	} else if score < 0.65 {
		preferred = false
	}
	return score, preferred
}
