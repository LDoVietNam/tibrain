// Package autocombo provides cost/latency/quota-aware provider selection.
// Picks best provider+model for a given task using weighted scoring with
// optional epsilon-greedy exploration.
package autocombo

import (
	"math/rand"
	"time"
)

type AccountTier string

const (
	TierFree     AccountTier = "free"
	TierStandard AccountTier = "standard"
	TierPro      AccountTier = "pro"
)

type CircuitState string

const (
	CircuitClosed CircuitState = "closed"
	CircuitOpen   CircuitState = "open"
)

// ProviderCandidate represents a provider+model option for selection.
type ProviderCandidate struct {
	Provider        string       `json:"provider"`
	Model           string       `json:"model"`
	QuotaRemaining  int          `json:"quota_remaining"`
	CircuitState    CircuitState `json:"circuit_state"`
	CostPer1MTokens float64      `json:"cost_per_1m_tokens"`
	P95LatencyMs    int          `json:"p95_latency_ms"`
	LatencyStdDev   int          `json:"latency_std_dev"`
	AccountTier     AccountTier  `json:"account_tier"`
	QuotaResetSecs  int          `json:"quota_reset_secs,omitempty"`
}

// SelectionResult is the chosen provider+model.
type SelectionResult struct {
	Provider string  `json:"provider"`
	Model    string  `json:"model"`
	Score    float64 `json:"score"`
}

// AutoComboConfig configures the selection algorithm.
type AutoComboConfig struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Weights         map[string]float64 `json:"weights"`
	ExplorationRate float64            `json:"exploration_rate"`
}

// DefaultWeights provides reasonable defaults for scoring.
var DefaultWeights = map[string]float64{
	"cost":    0.3,
	"latency": 0.4,
	"quota":   0.2,
	"tier":    0.1,
}

// Select picks the best candidate from the pool using weighted scoring.
// With probability ExplorationRate, picks a random valid candidate (epsilon-greedy).
func Select(config AutoComboConfig, pool []ProviderCandidate, taskType string) SelectionResult {
	if len(pool) == 0 {
		return SelectionResult{}
	}

	var candidates []ProviderCandidate
	for _, c := range pool {
		if c.CircuitState == CircuitOpen || c.QuotaRemaining <= 0 {
			continue
		}
		candidates = append(candidates, c)
	}

	if len(candidates) == 0 {
		return SelectionResult{
			Provider: pool[0].Provider,
			Model:    pool[0].Model,
			Score:    0,
		}
	}

	weights := config.Weights
	if weights == nil {
		weights = DefaultWeights
	}

	var best ProviderCandidate
	bestScore := -1.0

	for _, c := range candidates {
		score := calculateScore(c, weights)
		if score > bestScore {
			bestScore = score
			best = c
		}
	}

	if config.ExplorationRate > 0 && rand.Float64() < config.ExplorationRate {
		best = candidates[rand.Intn(len(candidates))]
	}

	return SelectionResult{
		Provider: best.Provider,
		Model:    best.Model,
		Score:    bestScore,
	}
}

func calculateScore(c ProviderCandidate, weights map[string]float64) float64 {
	costScore := 1.0 / (c.CostPer1MTokens + 0.01)
	latencyScore := 1.0 / (float64(c.P95LatencyMs) + 1.0)
	quotaScore := float64(c.QuotaRemaining) / 100.0

	tierScore := 0.0
	switch c.AccountTier {
	case TierPro:
		tierScore = 1.0
	case TierStandard:
		tierScore = 0.7
	case TierFree:
		tierScore = 0.5
	}

	return costScore*weights["cost"] +
		latencyScore*weights["latency"] +
		quotaScore*weights["quota"] +
		tierScore*weights["tier"]
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
