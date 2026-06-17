package learn

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// Arm represents a model option in the bandit.
type Arm struct {
	Wins     int       `json:"wins"`
	Losses   int       `json:"losses"`
	LastUsed time.Time `json:"last_used"`
}

// Bandit implements Thompson Sampling for multi-armed bandit problems.
// Thompson Sampling is a Bayesian algorithm that balances exploration vs exploitation.
type Bandit struct {
	mu         sync.RWMutex
	arms       map[string]*Arm // "provider:model" → arm
	priorAlpha float64         // Beta prior α (successes)
	priorBeta  float64         // Beta prior β (failures)
	rng        *rand.Rand
}

// NewBandit creates a Thompson Sampler with default priors.
func NewBandit(alpha, beta float64) *Bandit {
	if alpha <= 0 {
		alpha = 1.0 // Uniform prior
	}
	if beta <= 0 {
		beta = 1.0 // Uniform prior
	}

	src := rand.NewSource(time.Now().UnixNano())
	return &Bandit{
		arms:       make(map[string]*Arm),
		priorAlpha: alpha,
		priorBeta:  beta,
		rng:        rand.New(src),
	}
}

// AddArm registers a new model option.
func (b *Bandit) AddArm(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.arms[key]; !exists {
		b.arms[key] = &Arm{
			Wins:     int(b.priorAlpha) - 1, // Convert prior to counts
			Losses:   int(b.priorBeta) - 1,
			LastUsed: time.Time{}, // Zero time
		}
	}
}

// Update records the outcome for a model.
// quality: 0-10 scale, threshold converts to win/loss
func (b *Bandit) Update(key string, quality float64, threshold float64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	arm, exists := b.arms[key]
	if !exists {
		arm = &Arm{
			Wins:   int(b.priorAlpha) - 1,
			Losses: int(b.priorBeta) - 1,
		}
		b.arms[key] = arm
	}

	// Convert quality to binary reward
	win := quality >= threshold
	if win {
		arm.Wins++
	} else {
		arm.Losses++
	}
	arm.LastUsed = time.Now()
}

// Select chooses a model using Thompson Sampling.
// Returns the selected arm key (e.g., "openrouter:claude-sonnet").
func (b *Bandit) Select(explorationRate float64) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.arms) == 0 {
		return ""
	}

	// Thompson Sampling: sample from Beta(α+wins, β+losses) for each arm
	bestScore := -1.0
	selected := ""

	for key, arm := range b.arms {
		// Beta distribution sample (approximate via Gamma)
		alpha := float64(arm.Wins) + b.priorAlpha
		beta := float64(arm.Losses) + b.priorBeta

		// Sample from Beta(α, β) using acceptance-rejection
		score := b.sampleBeta(alpha, beta)

		// Exploration boost: if exploration mode (ε-greedy), add noise
		if rand.Float64() < explorationRate {
			score += rand.Float64() * 0.1 // Small random boost
		}

		if score > bestScore {
			bestScore = score
			selected = key
		}
	}

	return selected
}

// SelectUCB1 chooses using Upper Confidence Bound algorithm.
// Alternative to Thompson Sampling, good for comparison.
func (b *Bandit) SelectUCB1(totalTrials int) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.arms) == 0 {
		return ""
	}

	// UCB1: argmax( avg_reward + sqrt(2*ln(N)/n) )
	bestScore := -1.0
	selected := ""

	for key, arm := range b.arms {
		total := arm.Wins + arm.Losses
		if total == 0 {
			// Untried: give it a try (high priority)
			return key
		}

		avgReward := float64(arm.Wins) / float64(total)
		exploration := math.Sqrt(2.0 * math.Log(float64(totalTrials)) / float64(total))
		ucb := avgReward + exploration

		if ucb > bestScore {
			bestScore = ucb
			selected = key
		}
	}

	return selected
}

// GetArmCount returns number of arms.
func (b *Bandit) GetArmCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.arms)
}

// GetWinRate returns win rate for a model (0-1).
func (b *Bandit) GetWinRate(key string) float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	arm, ok := b.arms[key]
	if !ok || (arm.Wins+arm.Losses) == 0 {
		return 0.0
	}
	return float64(arm.Wins) / float64(arm.Wins+arm.Losses)
}

// GetAllStats returns all arm statistics.
func (b *Bandit) GetAllStats() map[string]struct {
	Wins    int
	Losses  int
	WinRate float64
} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	stats := make(map[string]struct {
		Wins    int
		Losses  int
		WinRate float64
	})

	for key, arm := range b.arms {
		total := arm.Wins + arm.Losses
		wr := 0.0
		if total > 0 {
			wr = float64(arm.Wins) / float64(total)
		}
		stats[key] = struct {
			Wins    int
			Losses  int
			WinRate float64
		}{arm.Wins, arm.Losses, wr}
	}

	return stats
}

// sampleBeta generates a random sample from Beta(α, β) distribution.
// Uses approximation: if α,β > 1, use normal approximation; else use Gamma.
func (b *Bandit) sampleBeta(alpha, beta float64) float64 {
	// For α,β >= 1, use normal approximation (faster)
	if alpha >= 1 && beta >= 1 {
		mean := alpha / (alpha + beta)
		variance := (alpha * beta) / ((alpha + beta) * (alpha + beta) * (alpha + beta + 1))
		stdDev := math.Sqrt(variance)
		// Normal sample
		z := b.rng.NormFloat64()
		x := mean + stdDev*z
		// Clamp to [0,1]
		if x < 0 {
			x = 0
		} else if x > 1 {
			x = 1
		}
		return x
	}

	// General case: Gamma(α, 1) / (Gamma(α,1) + Gamma(β,1))
	ga := b.sampleGamma(alpha, 1)
	gb := b.sampleGamma(beta, 1)
	return ga / (ga + gb)
}

// sampleGamma generates Gamma(shape, scale) random variable.
// Simplified: uses Marsaglia-Tsang method for shape >= 1.
func (b *Bandit) sampleGamma(shape, scale float64) float64 {
	if shape < 1 {
		// For shape < 1, use transformation
		return b.sampleGamma(shape+1, scale) * math.Pow(b.rng.Float64(), 1/shape)
	}

	// Marsaglia-Tsang method for α >= 1
	d := shape - 1/3.0
	c := 1 / math.Sqrt(9*d)

	for {
		x := b.rng.NormFloat64()
		v := math.Pow(1+c*x, 3)
		if v <= 0 {
			continue
		}

		u := b.rng.Float64()
		x2 := x * x
		if u < 1-0.0331*x2*x2 || math.Log(u) < 0.5*x2+d*(1-v+math.Log(v)) {
			return scale * d * v
		}
	}
}

// Reset clears all arms (for testing).
func (b *Bandit) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.arms = make(map[string]*Arm)
}

// LoadFromStats initializes bandit from historical model performance data.
// Converts avg_quality → wins/losses using threshold.
func (b *Bandit) LoadFromStats(stats map[string]struct {
	Samples    int
	AvgQuality float64
}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for key, stat := range stats {
		// Success threshold is 7.0 (quality 1-10 scale)
		winRate := stat.AvgQuality / 10.0 // Convert 1-10 to 0-1

		wins := int(winRate * float64(stat.Samples))
		if wins < 1 {
			wins = 1 // Prior
		}
		losses := stat.Samples - wins
		if losses < 1 {
			losses = 1 // Prior
		}

		b.arms[key] = &Arm{
			Wins:   wins,
			Losses: losses,
		}
	}
}
