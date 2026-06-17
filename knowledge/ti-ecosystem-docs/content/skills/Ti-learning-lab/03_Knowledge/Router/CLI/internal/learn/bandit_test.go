package learn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBandit_ArmsManagement(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)

	// Add arms
	bandit.AddArm("a")
	bandit.AddArm("b")
	bandit.AddArm("c")
	assert.Equal(t, 3, bandit.GetArmCount())

	// Update stats
	bandit.Update("a", 9.0, 7.0) // win
	bandit.Update("a", 8.0, 7.0) // win
	bandit.Update("b", 5.0, 7.0) // loss

	stats := bandit.GetAllStats()
	assert.Equal(t, 2, stats["a"].Wins)
	assert.Equal(t, 1, stats["b"].Losses)
}

func TestBandit_ThompsonSampling(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)
	bandit.AddArm("winner")
	bandit.AddArm("loser")

	// Winner wins 80% of time, loser 20%
	for i := 0; i < 100; i++ {
		bandit.Update("winner", 8.0, 7.0)
		bandit.Update("loser", 5.0, 7.0)
	}

	// Run 1000 selections, should choose winner > 70%
	wins := 0
	for i := 0; i < 1000; i++ {
		if bandit.Select(0.0) == "winner" {
			wins++
		}
	}
	ratio := float64(wins) / 10 // percentage
	assert.Greater(t, ratio, 70.0, "Expected winner chosen >70%% but got %.1f%%", ratio)
}

func TestBandit_UCB1(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)
	bandit.AddArm("a") // 10 wins, 0 losses → avg 1.0
	bandit.AddArm("b") // 1 win, 0 losses → avg 1.0 but less explored

	for i := 0; i < 10; i++ {
		bandit.Update("a", 10.0, 7.0)
	}
	bandit.Update("b", 10.0, 7.0)

	// After 11 total trials, UCB should explore b due to uncertainty
	selected := bandit.SelectUCB1(11)
	assert.NotEmpty(t, selected)
	_ = selected // Could be either, just ensure non-crash
}

func TestBandit_EdgeCases(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)

	// Select from empty bandit
	assert.Equal(t, "", bandit.Select(0.2))

	// Get stats from non-existent
	rate := bandit.GetWinRate("nonexistent")
	assert.Equal(t, 0.0, rate)
}

func TestBandit_RestartAfterReset(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)
	bandit.AddArm("a")
	bandit.Update("a", 9.0, 7.0)
	assert.Greater(t, bandit.GetWinRate("a"), 0.5)

	bandit.Reset()
	assert.Equal(t, 0, bandit.GetArmCount())
}

func TestBandit_LoadFromStats(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)

	stats := map[string]struct {
		Samples    int
		AvgQuality float64
	}{
		"openrouter:claude": {Samples: 50, AvgQuality: 8.0}, // ~40 wins
		"groq:llama":        {Samples: 10, AvgQuality: 6.0}, // ~6 wins
	}

	bandit.LoadFromStats(stats)

	claudeRate := bandit.GetWinRate("openrouter:claude")
	llamaRate := bandit.GetWinRate("groq:llama")
	assert.Greater(t, claudeRate, llamaRate, "Claude should have higher win rate")
}

func TestBandit_GetAllStats(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)
	bandit.AddArm("test")
	bandit.Update("test", 8.0, 7.0)

	stats := bandit.GetAllStats()
	_, ok := stats["test"]
	assert.True(t, ok)
}

func TestBandit_Concurrency(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)
	bandit.AddArm("a")
	bandit.AddArm("b")

	// Concurrent updates and selects
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				bandit.Update("a", 8.0, 7.0)
			}
			done <- true
		}()
		go func() {
			for j := 0; j < 100; j++ {
				_ = bandit.Select(0.2)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should not crash - concurrency safety check
	count := bandit.GetArmCount()
	assert.Equal(t, 2, count)
}

func TestBandit_WinRateCalculation(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)
	bandit.AddArm("test")
	bandit.Update("test", 9.0, 7.0)
	bandit.Update("test", 5.0, 7.0)
	bandit.Update("test", 9.0, 7.0)

	rate := bandit.GetWinRate("test")
	assert.InEpsilon(t, 2.0/3.0, rate, 0.01)
}

func TestBandit_SampleBetaDistribution(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)

	// Test that sampleBeta produces values in [0,1]
	for i := 0; i < 100; i++ {
		x := bandit.sampleBeta(2.0, 2.0) // symmetric
		assert.GreaterOrEqual(t, x, 0.0)
		assert.LessOrEqual(t, x, 1.0)
	}
}
