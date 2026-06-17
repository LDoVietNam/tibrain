package routeagent

import (
	"sort"
)

type PromptPreference struct {
	PromptShape     string  `json:"prompt_shape"`
	SuccessCount    int     `json:"success_count"`
	FailureCount    int     `json:"failure_count"`
	AvgQuality      float64 `json:"avg_quality"`
	VerifySuccesses int     `json:"verify_successes"`
	SuccessRate     float64 `json:"success_rate"`
	VerifyRate      float64 `json:"verify_rate"`
	Score           float64 `json:"score"`
}

type RoutePreferenceDiagnostics struct {
	Project           string             `json:"project"`
	TaskType          string             `json:"task_type"`
	Route             string             `json:"route"`
	Policy            PolicyStats        `json:"policy"`
	PromptPreferences []PromptPreference `json:"prompt_preferences"`
	BestPromptShape   string             `json:"best_prompt_shape,omitempty"`
}

func promptPreferenceScore(succ, fail int, avgQuality float64, verifySucc int) float64 {
	total := succ + fail
	successRate := 0.5
	if total > 0 {
		successRate = float64(succ+1) / float64(total+2)
	}
	verifyRate := 0.0
	if succ > 0 {
		verifyRate = float64(verifySucc) / float64(succ)
	}
	score := 0.45*successRate + 0.35*avgQuality + 0.20*verifyRate
	if score > 1.0 {
		score = 1.0
	}
	return score
}

func (s *Store) PromptPreferences(project, taskType, route string) ([]PromptPreference, error) {
	rows, err := s.db.Query(`SELECT prompt_shape, success_count, failure_count, avg_quality, verify_successes
FROM prompt_policies WHERE project=? AND task_type=? AND route=? ORDER BY updated_at DESC`, project, taskType, route)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PromptPreference
	for rows.Next() {
		var p PromptPreference
		if err := rows.Scan(&p.PromptShape, &p.SuccessCount, &p.FailureCount, &p.AvgQuality, &p.VerifySuccesses); err != nil {
			return nil, err
		}
		total := p.SuccessCount + p.FailureCount
		if total > 0 {
			p.SuccessRate = float64(p.SuccessCount) / float64(total)
		} else {
			p.SuccessRate = 0.5
		}
		if p.SuccessCount > 0 {
			p.VerifyRate = float64(p.VerifySuccesses) / float64(p.SuccessCount)
		}
		p.Score = promptPreferenceScore(p.SuccessCount, p.FailureCount, p.AvgQuality, p.VerifySuccesses)
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out, nil
}

func (s *Store) BestPromptPreference(project, taskType, route string) (*PromptPreference, error) {
	prefs, err := s.PromptPreferences(project, taskType, route)
	if err != nil || len(prefs) == 0 {
		return nil, err
	}
	best := prefs[0]
	if best.Score < 0.55 {
		return nil, nil
	}
	return &best, nil
}

func (s *Store) RoutePreferenceDiagnostics(project, taskType, route string) (*RoutePreferenceDiagnostics, error) {
	policy, err := s.Policy(project, taskType, route)
	if err != nil {
		return nil, err
	}
	prefs, err := s.PromptPreferences(project, taskType, route)
	if err != nil {
		return nil, err
	}
	d := &RoutePreferenceDiagnostics{Project: project, TaskType: taskType, Route: route, Policy: policy, PromptPreferences: prefs}
	if len(prefs) > 0 {
		d.BestPromptShape = prefs[0].PromptShape
	}
	return d, nil
}
