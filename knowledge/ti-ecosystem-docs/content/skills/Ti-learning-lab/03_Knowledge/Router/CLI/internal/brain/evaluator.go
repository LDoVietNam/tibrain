package brain

import "math"

// EvaluateResponse scores a response based on multiple factors.
// Returns a score in [0, 1].
func EvaluateResponse(success bool, latencyMs int64, quality float64) float64 {
	score := 0.5 // Base score

	// Success factor (most important — 30% weight)
	if success {
		score += 0.3
	} else {
		score -= 0.3
	}

	// Quality factor (15% weight)
	if quality > 0 {
		score += quality * 0.15
	}

	// Latency factor: faster is better (10% weight)
	if latencyMs > 0 {
		if latencyMs < 2000 {
			score += 0.1 // Fast response
		} else if latencyMs > 10000 {
			score -= 0.05 // Slow response
		}
	}

	// Normalize to 0-1 range
	return math.Max(0, math.Min(1, score))
}
