package brain

import "strings"

// ModelIntelligence holds model classification and routing rules.
type ModelIntelligence struct {
	// Size tiers — route based on task complexity
	TinyModels   []string `json:"tiny_models"`   // 1B-9B: quick tasks
	SmallModels  []string `json:"small_models"`  // 8B-7x7B: moderate tasks
	MediumModels []string `json:"medium_models"` // 32B-70B: complex reasoning
	LargeModels  []string `json:"large_models"`  // 100B+: coding-plus, deep analysis

	// Cost tiers — route based on budget
	FreeModels  []string `json:"free_models"`
	CheapModels []string `json:"cheap_models"`

	// Task routing rules (Vietnamese)
	RoutingRules string `json:"routing_rules"`
}

// DefaultModelIntelligence returns sensible defaults.
func DefaultModelIntelligence() *ModelIntelligence {
	return &ModelIntelligence{
		TinyModels:   []string{"gemma-2b", "phi-2", "qwen-1.8b"},
		SmallModels:  []string{"llama-3.1-8b", "gemma-7b", "qwen-7b", "mistral-7b"},
		MediumModels: []string{"llama-3.3-70b", "qwen-32b", "mixtral-8x7b", "gemma-27b"},
		LargeModels:  []string{"claude-sonnet", "claude-opus", "gpt-4o", "kimi-k2", "deepseek-v3"},

		FreeModels:  []string{"groq-llama", "groq-gemma", "groq-mixtral", "cloudflare"},
		CheapModels: []string{"qwen", "deepseek-chat", "gemini-flash"},

		RoutingRules: "1. Luôn thử FREE models trước\n" +
			"2. Coding phức tạp → large models (sonnet, opus, gpt-4o)\n" +
			"3. Reasoning đơn giản → medium models đủ\n" +
			"4. Dịch/tóm tắt → small models cũng OK\n" +
			"5. Không dùng large models cho task đơn giản — tốn token vô ích",
	}
}

// ClassifyModelBySize returns the size tier for a model name.
func (mi *ModelIntelligence) ClassifyModelBySize(model string) string {
	lower := strings.ToLower(model)

	for _, m := range mi.LargeModels {
		if strings.Contains(lower, strings.ToLower(m)) {
			return "large"
		}
	}
	for _, m := range mi.MediumModels {
		if strings.Contains(lower, strings.ToLower(m)) {
			return "medium"
		}
	}
	for _, m := range mi.SmallModels {
		if strings.Contains(lower, strings.ToLower(m)) {
			return "small"
		}
	}
	for _, m := range mi.TinyModels {
		if strings.Contains(lower, strings.ToLower(m)) {
			return "tiny"
		}
	}
	return "unknown"
}

// ClassifyModelByCost returns the cost tier for a model name.
func (mi *ModelIntelligence) ClassifyModelByCost(model string) string {
	lower := strings.ToLower(model)

	for _, m := range mi.FreeModels {
		if strings.Contains(lower, strings.ToLower(m)) {
			return "free"
		}
	}
	for _, m := range mi.CheapModels {
		if strings.Contains(lower, strings.ToLower(m)) {
			return "cheap"
		}
	}
	return "expensive"
}

// GetCandidatesForTask returns models suitable for a task type.
func (mi *ModelIntelligence) GetCandidatesForTask(taskType string) []string {
	switch taskType {
	case "coding", "debugging", "refactoring", "analysis":
		return append(mi.LargeModels, mi.MediumModels...)
	case "reasoning", "explaining":
		return append(mi.LargeModels, mi.MediumModels...)
	case "review", "spec", "planning":
		return append(mi.LargeModels, mi.MediumModels...)
	case "writing", "translation", "summarization":
		return append(mi.MediumModels, mi.SmallModels...)
	default:
		return append(mi.MediumModels, mi.SmallModels...)
	}
}

// GetBestFreeModel returns the best free model for a task.
func (mi *ModelIntelligence) GetBestFreeModel(taskType string) string {
	candidates := mi.GetCandidatesForTask(taskType)
	for _, c := range candidates {
		for _, f := range mi.FreeModels {
			if strings.Contains(strings.ToLower(c), strings.ToLower(f)) {
				return c
			}
		}
	}
	if len(mi.FreeModels) > 0 {
		return mi.FreeModels[0]
	}
	return ""
}
