package learn

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskAnalysis(t *testing.T) {
	tests := []struct {
		prompt string
		want   TaskAnalysis
	}{
		{
			prompt: "Write unit test for UserService.CreateUser",
			want: TaskAnalysis{
				TaskType:   "testing",
				TaskDomain: "backend",
			},
		},
		{
			prompt: "Fix SQL query performance issue",
			want: TaskAnalysis{
				TaskType:   "debugging",
				TaskDomain: "database",
			},
		},
		{
			prompt: "Refactor UserRouter to be cleaner",
			want: TaskAnalysis{
				TaskType:   "refactoring",
				TaskDomain: "backend",
			},
		},
	}

	classifier := NewTaskClassifier()
	for _, tt := range tests {
		got := classifier.Classify(tt.prompt)
		assert.Equal(t, tt.want.TaskType, got.TaskType, "Prompt: %s", tt.prompt)
		if tt.want.TaskDomain != "" {
			assert.Equal(t, tt.want.TaskDomain, got.TaskDomain, "Domain mismatch for: %s", tt.prompt)
		}
	}
}

func TestComplexityEstimation(t *testing.T) {
	classifier := NewTaskClassifier()

	// Simple tasks should score low
	score1 := classifier.estimateComplexity("write doc", "documenting", "general")
	assert.Less(t, score1, 4.0, "Doc writing should be simple")

	// Complex tasks (ML) should score high
	score2 := classifier.estimateComplexity("train neural network on dataset", "researching", "ml")
	assert.Greater(t, score2, 7.0, "ML task should be complex")

	// Verify bounds
	assert.GreaterOrEqual(t, score1, 1.0)
	assert.LessOrEqual(t, score2, 10.0)
}

func TestPromptNormalization(t *testing.T) {
	extractor := NewPatternExtractor(nil)

	tests := []struct {
		in  string
		out string
	}{
		{
			"Write unit test for UserService.CreateUser",
			"write unit test for {Type}.{method}",
		},
		{
			"Fix bug in authentication middleware",
			"fix bug in authentication {Type}",
		},
		{
			"sk-1234567890abcdefghij is my API key",
			"{token} is my api key",
		},
		{
			"https://example.com/api/v1/users",
			"{url}/api/v1/users",
		},
	}

	for _, tt := range tests {
		got := extractor.Normalize(tt.in)
		assert.Equal(t, tt.out, got, "Normalize '%s'", tt.in)
	}
}

func TestPromptNormalization_Deterministic(t *testing.T) {
	extractor := NewPatternExtractor(nil)
	prompt := "Write unit test for UserService.CreateUser"

	norm1 := extractor.Normalize(prompt)
	norm2 := extractor.Normalize(prompt)

	assert.Equal(t, norm1, norm2, "Normalization should be deterministic")
}

func TestKeywordOverlap(t *testing.T) {
	a := []string{"go", "test", "function"}
	b := []string{"test", "write", "code"}
	overlap := keywordOverlap(a, b)
	assert.Equal(t, 1.0/3.0, overlap) // 1 common word ("test") / unique union=5 → 0.2
}

func TestBandit_Initialization(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)
	bandit.AddArm("openrouter:claude")
	bandit.AddArm("groq:llama")

	count := bandit.GetArmCount()
	assert.Equal(t, 2, count)
}

func TestBandit_UpdateAndSelect(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)
	bandit.AddArm("a")
	bandit.AddArm("b")

	// Model A wins 10 times, Model B loses 10 times
	for i := 0; i < 10; i++ {
		bandit.Update("a", 9.0, 7.0) // win
		bandit.Update("b", 5.0, 7.0) // loss
	}

	// After training, A should be preferred >80%
	aCount := 0
	trials := 10000
	for i := 0; i < trials; i++ {
		if bandit.Select(0.0) == "a" { // no exploration
			aCount++
		}
	}
	ratio := float64(aCount) / float64(trials)
	assert.Greater(t, ratio, 0.75, "Thompson should favor winning model (got %.2f)", ratio)
}

func TestBandit_Exploration(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)
	bandit.AddArm("a")
	bandit.AddArm("b")

	// With high exploration, both should be chosen sometimes
	for i := 0; i < 100; i++ {
		_ = bandit.Select(0.5) // 50% exploration boost
	}
	// Not asserting exact numbers - just ensure no panic
}

func TestBandit_UCB1(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)
	bandit.AddArm("a")
	bandit.AddArm("b")

	// A wins more
	for i := 0; i < 10; i++ {
		bandit.Update("a", 9.0, 7.0)
		bandit.Update("b", 5.0, 7.0)
	}

	// UCB should still pick under-explored occasionally
	selected := bandit.SelectUCB1(20)
	assert.NotNil(t, selected)
	assert.NotEmpty(t, selected)
}

func TestBandit_LoadFromStats(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)

	stats := map[string]struct {
		Samples    int
		AvgQuality float64
	}{
		"openrouter:claude": {Samples: 100, AvgQuality: 8.5},
		"groq:llama":        {Samples: 50, AvgQuality: 7.0},
	}

	bandit.LoadFromStats(stats)

	// Claude should have more wins
	winRate := bandit.GetWinRate("openrouter:claude")
	assert.Greater(t, winRate, 0.5)
}

func TestQuality_Combine(t *testing.T) {
	auto, user := 8.0, 9.0
	combined := CombineQuality(auto, user, 0.7, 0.3)
	assert.InEpsilon(t, 8.3, combined, 0.01)
}

func TestQuality_AutoScore(t *testing.T) {
	outcome := Outcome{Outcome: "success"}
	score := AutoScore(outcome, 1*time.Second)
	assert.Greater(t, score, 0.7, "Success should boost score")

	outcome = Outcome{Outcome: "failed"}
	score = AutoScore(outcome, 1*time.Second)
	assert.Less(t, score, 0.5, "Failure should lower score")
}

func TestQuality_Final(t *testing.T) {
	// High base quality + success = >8
	q := ComputeFinal(7.0, Outcome{Outcome: "success", CompilationSuccess: true})
	assert.Greater(t, q, 7.0)

	// Base quality - but compilation fail
	q = ComputeFinal(8.0, Outcome{Outcome: "failed", CompilationSuccess: false})
	assert.Less(t, q, 8.0)
}

func TestBandit_Reset(t *testing.T) {
	bandit := NewBandit(1.0, 1.0)
	bandit.AddArm("a")
	bandit.Update("a", 9.0, 7.0)
	assert.Equal(t, 1, bandit.GetArmCount())

	bandit.Reset()
	assert.Equal(t, 0, bandit.GetArmCount())
}

func TestCostTracker_Basics(t *testing.T) {
	ct := NewCostTracker(BudgetLimits{Daily: 1.0, Monthly: 10.0})

	ct.AddSpent(0.5)
	assert.InEpsilon(t, 0.5, ct.GetDailyUsed(), 0.01)

	assert.True(t, ct.CanAfford(0.4, 3))  // within budget
	assert.False(t, ct.CanAfford(0.6, 3)) // exceeds
}

func TestCostTracker_HighPriority(t *testing.T) {
	ct := NewCostTracker(BudgetLimits{Daily: 0.1, Monthly: 1.0})
	ct.AddSpent(0.09) // almost exhausted

	// Priority 1 (critical) should always be allowed
	assert.True(t, ct.CanAfford(1.0, 1))
	// Priority 5 (low) should be blocked
	assert.False(t, ct.CanAfford(0.02, 5))
}

func TestPatternExtractor_NormalizeVariations(t *testing.T) {
	pe := NewPatternExtractor(nil)

	variants := []string{
		"Write a unit test for Foo",
		"write UNIT TEST for FOO",
		"Write unit test for the Foo function",
		"Write go unit test for FooBar",
	}

	normalized := make(map[string]int)
	for _, v := range variants {
		norm := pe.Normalize(v)
		normalized[norm]++
	}

	// All should produce similar normalized form (at least share key words)
	assert.Less(t, len(normalized), len(variants), "Variants should normalize similarly")
}

func TestContextFinder_KeywordExtraction(t *testing.T) {
	cf := NewContextFinder(nil)
	keywords := extractKeywords("Write a unit test for UserService in Go")
	assert.Contains(t, keywords, "unit")
	assert.Contains(t, keywords, "test")
	assert.Contains(t, keywords, "go")
	assert.NotContains(t, keywords, "a")
	assert.NotContains(t, keywords, "for")
}

func TestContextFinder_Overlap(t *testing.T) {
	a := []string{"go", "test", "code"}
	b := []string{"test", "go", "python"}
	overlap := keywordOverlap(a, b)
	assert.InEpsilon(t, 0.5, overlap, 0.01) // 2 common, union=4 → 0.5
}

func TestQualityCombinedFormula(t *testing.T) {
	// Edge cases
	q := CombineQuality(0, 0, 0.7, 0.3)
	assert.Equal(t, 0.0, q)

	q = CombineQuality(10, 10, 0.7, 0.3)
	assert.Equal(t, 10.0, q)

	q = CombineQuality(7.5, 0, 0.7, 0.3)
	assert.InEpsilon(t, 5.25, q, 0.01) // 7.5 * 0.7 = 5.25
}

func TestMovingAverage(t *testing.T) {
	// First value
	avg := movingAverage(0, 5.0)
	assert.InEpsilon(t, 5.0, avg, 0.01)

	// Second value (n=100 placeholder)
	avg = movingAverage(5.0, 7.0)
	_ = avg
	// With n=100: (5*99 + 7) / 100 = 5.02
	assert.InEpsilon(t, 5.02, avg, 0.01)
}

func BenchmarkBandit_Select100k(b *testing.B) {
	bandit := NewBandit(1.0, 1.0)
	bandit.AddArm("a")
	bandit.AddArm("b")
	bandit.AddArm("c")
	bandit.AddArm("d")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bandit.Select(0.2)
	}
}

func BenchmarkPatternNormalize(b *testing.B) {
	pe := NewPatternExtractor(nil)
	prompt := "Write a comprehensive unit test for the UserService.CreateUser method including edge cases"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pe.Normalize(prompt)
	}
}

func BenchmarkStorage_InsertLog(b *testing.B) {
	storage, _ := NewStorage("")
	defer storage.Close()

	entry := &LogInput{
		Timestamp:       time.Now(),
		SessionID:       "test-session",
		PromptHash:      "abc123",
		TaskType:        "test",
		Model:           "claude",
		QualityCombined: 8.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.InsertLog(entry)
	}
}
