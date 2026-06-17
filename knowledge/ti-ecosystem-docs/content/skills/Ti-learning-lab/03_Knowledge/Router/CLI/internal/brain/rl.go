package brain

import (
	"math"
	"time"
)

// ─── Reinforcement Learning ───

// applyReinforcementLearning adjusts model selection based on success rates.
func (e *Engine) applyReinforcementLearning() {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Apply learning rate decay once per cycle, not per task type
	var decayApplied bool

	for taskType, models := range e.taskProviderScore {
		var bestModel string
		var bestScore float64

		for model, ps := range models {
			if ps.TotalCalls < 3 {
				continue // Not enough data
			}

			// Combined score: 60% success rate + 40% avg evaluation score
			successRate := float64(ps.SuccessCalls) / float64(ps.TotalCalls)
			combinedScore := 0.6*successRate + 0.4*ps.AvgScore

			if combinedScore > bestScore {
				bestScore = combinedScore
				bestModel = model
			}
		}

		// Only switch if confidence is high enough
		if bestModel != "" && bestScore >= e.minConfidence {
			oldPreferred := e.preferredProvider[taskType]
			if oldPreferred != bestModel {
				e.preferredProvider[taskType] = bestModel
			}
		}

		if !decayApplied {
			e.learningRate = math.Max(0.01, e.learningRate*0.99)
			decayApplied = true
		}
	}
}

// SelectBestModel chooses the best model for a given prompt.
func (e *Engine) SelectBestModel(prompt string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	taskType := ClassifyTask(prompt)

	// Check RL-learned preference
	if preferred, ok := e.preferredProvider[taskType]; ok {
		// Verify model still has good performance
		if ps, ok := e.taskProviderScore[taskType][preferred]; ok {
			if ps.TotalCalls >= 3 {
				performance := 0.6*(float64(ps.SuccessCalls)/float64(ps.TotalCalls)) + 0.4*ps.AvgScore
				if performance >= 0.75 {
					return preferred
				}
			}
		}
	}

	// Fallback: use model intelligence to get candidates
	if e.modelIntelligence != nil {
		candidates := e.modelIntelligence.GetCandidatesForTask(taskType)
		if len(candidates) > 0 {
			return candidates[0]
		}
	}

	// Default fallback
	return "claude-sonnet-4"
}

// GetPreferredModel returns the preferred model for a task type.
// Returns empty string if no preference learned yet.
func (e *Engine) GetPreferredModel(taskType string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.preferredProvider[taskType]
}

// GetModelScore returns performance score for a model on a task type.
// Returns 0 if no data available.
func (e *Engine) GetModelScore(taskType, model string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ps, ok := e.taskProviderScore[taskType][model]
	if !ok || ps.TotalCalls < 3 {
		return 0
	}
	return 0.6*(float64(ps.SuccessCalls)/float64(ps.TotalCalls)) + 0.4*ps.AvgScore
}

// AnalyzeModelPerformance returns performance metrics for a model.
func (e *Engine) AnalyzeModelPerformance(model string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	pm, ok := e.providerMetrics[model]
	if !ok {
		return 0.75 // Default unknown model score
	}

	return pm.SuccessRate
}

// RetrainForTaskType reduces confidence and clears patterns for a task type.
func (e *Engine) RetrainForTaskType(taskType string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if taskType == "" {
		taskType = "general"
	}

	// Decay scores slightly
	if providers, ok := e.taskProviderScore[taskType]; ok {
		now := time.Now()
		for model := range providers {
			ps := providers[model]
			ps.AvgScore *= 0.9
			ps.LastUpdated = now
		}
	}

	// Clear prompt pattern to force re-learning
	delete(e.promptPatterns, taskType)
}
