package memory

import "strings"

type PromotionRule struct {
	Kind                string
	MinConfidence       float64
	AllowedSources      []string
	RequireScope        bool
	RequiredPayloadKeys []string
}

var defaultPromotionRules = map[string]PromotionRule{
	"approved_plan": {
		Kind: "approved_plan", MinConfidence: 0.95, AllowedSources: []string{"planner", "ticore"}, RequireScope: true,
		RequiredPayloadKeys: []string{"plan_id", "goal", "task_type", "summary"},
	},
	"project_fact": {
		Kind: "project_fact", MinConfidence: 0.90, AllowedSources: []string{"repoindex", "user", "ticore"}, RequireScope: true,
	},
	"route_policy": {
		Kind: "route_policy", MinConfidence: 0.90, AllowedSources: []string{"routeagent", "ticore"}, RequireScope: true,
	},
	"step_completion": {
		Kind: "step_completion", MinConfidence: 0.95, AllowedSources: []string{"verify", "ticore"}, RequireScope: true,
		RequiredPayloadKeys: []string{"plan_id", "step_id", "step_title", "status"},
	},
	"verified_step": {
		Kind: "verified_step", MinConfidence: 0.96, AllowedSources: []string{"verify"}, RequireScope: true,
		RequiredPayloadKeys: []string{"plan_id", "step_id", "step_title", "verify_commands", "verify_passed"},
	},
}

func (s *GovernedStore) Promote(kind, scope, source string, confidence float64, payload map[string]any) (bool, error) {
	rule, ok := defaultPromotionRules[kind]
	if !ok {
		return false, nil
	}
	if !canPromote(rule, scope, source, confidence, payload) {
		return false, nil
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["promotion_rule"] = kind
	payload["promoted"] = true
	return true, s.append(Item{Layer: LayerCanonical, Kind: kind, Scope: scope, Source: source, Confidence: confidence, Payload: payload})
}

func VerificationConfidence(success bool, commandCount, failedCount int) float64 {
	if !success || failedCount > 0 || commandCount <= 0 {
		return 0.65
	}
	switch {
	case commandCount >= 3:
		return 0.99
	case commandCount == 2:
		return 0.97
	default:
		return 0.96
	}
}

func canPromote(rule PromotionRule, scope, source string, confidence float64, payload map[string]any) bool {
	if confidence < rule.MinConfidence {
		return false
	}
	if rule.RequireScope && strings.TrimSpace(scope) == "" {
		return false
	}
	if len(rule.AllowedSources) > 0 {
		match := false
		for _, allowed := range rule.AllowedSources {
			if strings.EqualFold(strings.TrimSpace(source), strings.TrimSpace(allowed)) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	for _, key := range rule.RequiredPayloadKeys {
		if payload == nil {
			return false
		}
		if _, ok := payload[key]; !ok {
			return false
		}
	}
	if rule.Kind == "verified_step" {
		v, ok := payload["verify_passed"].(bool)
		if !ok || !v {
			return false
		}
	}
	return true
}
