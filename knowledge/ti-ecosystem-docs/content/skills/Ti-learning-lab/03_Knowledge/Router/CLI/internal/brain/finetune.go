package brain

// ─── Fine-Tuning System ───

// fineTuneSystem updates prompts and configs based on learned patterns.
func (e *Engine) fineTuneSystem() {
	patterns := e.analyzeTaskPatterns()

	// Acquire write lock for promptPatterns modifications
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, pattern := range patterns {
		switch pattern.TaskType {
		case "coding":
			e.adjustCodingPrompts(pattern)
		case "debugging":
			e.adjustDebuggingPrompts(pattern)
		case "writing":
			e.adjustWritingPrompts(pattern)
		case "explaining":
			e.adjustExplainingPrompts(pattern)
		}
	}
}

// adjustCodingPrompts fine-tunes prompt guidance for coding tasks.
func (e *Engine) adjustCodingPrompts(pattern TaskPattern) {
	if pattern.SuccessRate < 0.80 {
		e.promptPatterns["coding"] = `For coding tasks, provide:
1. Detailed debugging steps for each issue
2. Code snippets with explanations
3. Edge cases and error handling
4. Best practices and optimizations
Always include the full corrected code.`
	}
}

// adjustDebuggingPrompts fine-tunes prompt guidance for debugging tasks.
func (e *Engine) adjustDebuggingPrompts(pattern TaskPattern) {
	if pattern.SuccessRate < 0.80 {
		e.promptPatterns["debugging"] = `For debugging tasks, provide:
1. Root cause analysis
2. Step-by-step reproduction
3. Fix with explanation
4. Prevention tips
Always verify the fix works.`
	}
}

// adjustWritingPrompts fine-tunes prompt guidance for writing tasks.
func (e *Engine) adjustWritingPrompts(pattern TaskPattern) {
	if pattern.SuccessRate < 0.80 {
		e.promptPatterns["writing"] = `For writing tasks, provide:
1. Clear, concise language
2. Proper structure and formatting
3. Appropriate tone for the audience
4. Proofread content`
	}
}

// adjustExplainingPrompts fine-tunes prompt guidance for explaining tasks.
func (e *Engine) adjustExplainingPrompts(pattern TaskPattern) {
	if pattern.SuccessRate < 0.80 {
		e.promptPatterns["explaining"] = `For explaining tasks, provide:
1. Simple, clear explanations first
2. Technical details for depth
3. Examples and analogies
4. Common misconceptions to avoid`
	}
}

// analyzeTaskPatterns extracts patterns from learned data.
func (e *Engine) analyzeTaskPatterns() []TaskPattern {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var patterns []TaskPattern

	for taskType, models := range e.taskProviderScore {
		var totalCalls, successCalls int
		var totalScore float64
		var bestModel string
		var bestScore float64

		for model, ps := range models {
			totalCalls += ps.TotalCalls
			successCalls += ps.SuccessCalls
			totalScore += ps.AvgScore * float64(ps.TotalCalls)

			if ps.AvgScore > bestScore {
				bestScore = ps.AvgScore
				bestModel = model
			}
		}

		if totalCalls > 0 {
			// Get avg latency and cost from provider metrics
			var avgLatency, avgCost float64
			for _, pm := range e.providerMetrics {
				if count, ok := pm.TaskBreakdown[taskType]; ok && count > 0 {
					avgLatency += pm.AvgLatencyMs
					avgCost += pm.TotalCost
				}
			}
			_ = avgLatency
			_ = avgCost

			patterns = append(patterns, TaskPattern{
				TaskType:     taskType,
				SuccessRate:  float64(successCalls) / float64(totalCalls),
				AvgScore:     totalScore / float64(totalCalls),
				AvgLatencyMs: avgLatency,
				AvgCost:      avgCost,
				SampleCount:  totalCalls,
				BestModel:    bestModel,
			})
		}
	}

	return patterns
}

// GetPromptPattern returns the tuned prompt for a task type.
// Returns empty string if no tuning learned yet.
func (e *Engine) GetPromptPattern(taskType string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.promptPatterns[taskType]
}

// GetTaskPatterns returns all learned task patterns.
func (e *Engine) GetTaskPatterns() []TaskPattern {
	return e.analyzeTaskPatterns()
}
