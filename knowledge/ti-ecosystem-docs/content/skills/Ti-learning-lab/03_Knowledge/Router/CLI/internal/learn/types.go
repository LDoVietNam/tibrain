// Package learn implements the BEADS v2 learning system.
// This package extends the existing beads logging with:
// - Extended metrics (quality_auto, compilation_success, etc.)
// - Pattern normalization and extraction
// - Thompson Sampling bandit for model selection
// - Cost-aware budget optimization
// - Context injection from similar sessions
package learn

import (
	"time"
)

const (
	// Schema version for this extension
	SchemaVersion = 3

	// Quality thresholds
	QualitySuccessThreshold = 7.0 // quality >= 7 → success
	QualityMinForPattern    = 0.7 // min quality to store pattern
	MinSamplesForPattern    = 10  // min samples before using pattern
)

// TaskAnalysis extracts metadata from a prompt for routing decisions.
type TaskAnalysis struct {
	TaskType        string   `json:"task_type"`        // code, debug, test, refactor, doc, research
	TaskDomain      string   `json:"task_domain"`      // backend, frontend, db, devops, ml
	Complexity      float64  `json:"complexity"`       // 1-10 score
	Keywords        []string `json:"keywords"`         // extracted keywords
	EstimatedTokens int      `json:"estimated_tokens"` // rough token count
	PromptHash      string   `json:"prompt_hash"`      // SHA256 for dedup
	PromptLength    int      `json:"prompt_length"`    // char count
}

// Outcome captures the result of a task execution.
type Outcome struct {
	Outcome            string  `json:"outcome"` // success, partial, failed, error
	Reason             string  `json:"reason"`  // why failed
	CompilationSuccess bool    `json:"compilation_success"`
	TestRunSuccess     bool    `json:"test_run_success"`
	TestPassRate       float64 `json:"test_pass_rate"` // 0-1
	GitCommitted       bool    `json:"git_committed"`
	UserEdited         bool    `json:"user_edited"`
	IterationsNeeded   int     `json:"iterations_needed"`
}

// QualityScores combines multiple quality signals.
type QualityScores struct {
	Auto         float64 `json:"auto"`          // heuristic 0-1
	User         float64 `json:"user"`          // explicit feedback 1-10 (0 if none)
	Combined     float64 `json:"combined"`      // weighted avg
	CompileScore float64 `json:"compile_score"` // from compilation
	TestScore    float64 `json:"test_score"`    // from test results
	Final        float64 `json:"final"`         // final quality 1-10
}

// LogInput is the extended BEADS v2 entry.
type LogInput struct {
	Timestamp    time.Time
	SessionID    string
	PromptHash   string
	PromptLength int
	Prompt       string // raw prompt text for context indexing

	// Task analysis
	TaskType        string
	TaskDomain      string
	ComplexityScore float64
	Keywords        []string

	// Context injection
	ContextSources    []string // ["session", "memory", "file"]
	ContextItemsCount int

	// Provider/Model
	Provider      string
	Model         string
	ModelFamily   string
	RoutingSource string // "default", "beads", "manual", "fallback"

	// Usage
	TokensInput      int
	TokensOutput     int
	LatencyMs        int64
	LatencyBreakdown map[string]int64 // {"network": 200, "generation": 1500}
	CostUSD          float64

	// Quality
	QualityUser     float64 // 1-10 from explicit feedback
	QualityAuto     float64 // heuristic 0-1
	QualityCombined float64 // weighted

	// Implicit signals
	CompilationSuccess bool
	TestRunSuccess     bool
	TestPassRate       float64
	GitCommitted       bool
	UserEdited         bool
	IterationsNeeded   int

	// Outcome
	Outcome       string
	OutcomeReason string

	// Pattern matching (filled async)
	PromptTemplate     string
	MatchedPatternID   int
	RecommendedContext string
	FallbackChain      []string
}

// Config holds BEADS learning configuration.
type Config struct {
	Mode                        string  `json:"mode"`                           // off, shadow, hybrid, auto
	HybridRatio                 int     `json:"hybrid_ratio"`                   // % traffic for learned routing (0-100)
	DailyBudgetUSD              float64 `json:"daily_budget_usd"`               // $ per day
	MonthlyBudgetUSD            float64 `json:"monthly_budget_usd"`             // $ per month
	SuccessThreshold            float64 `json:"success_threshold"`              // quality >= this = success (1-10)
	AutoQualityWeight           float64 `json:"auto_quality_weight"`            // weight for auto score (0-1)
	UserQualityWeight           float64 `json:"user_quality_weight"`            // weight for user score (0-1)
	ExplorationRate             float64 `json:"exploration_rate"`               // bandit exploration (0-1)
	UCBConfidence               float64 `json:"ucb_confidence"`                 // UCB confidence multiplier
	MinSamplesForRecommendation int     `json:"min_samples_for_recommendation"` // before trusting model
	PromptStorageLevel          string  `json:"prompt_storage_level"`           // hash, encrypted, plain
	EnableContextInjection      bool    `json:"enable_context_injection"`
	EnablePromptEnhancement     bool    `json:"enable_prompt_enhancement"`
	EnableCostOptimization      bool    `json:"enable_cost_optimization"`
	EnableFallbackChain         bool    `json:"enable_fallback_chain"`
	RetentionRawDays            int     `json:"retention_raw_days"`        // keep raw logs this many days
	RetentionAggregatesDays     int     `json:"retention_aggregates_days"` // keep aggregates this many days
	ShareAggregatedAnalytics    bool    `json:"share_aggregated_analytics"`

	// Advanced: Bandit priors (Beta distribution)
	BanditPriorAlpha float64 `json:"bandit_prior_alpha"` // prior successes
	BanditPriorBeta  float64 `json:"bandit_prior_beta"`  // prior failures
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Mode:                        "shadow", // default to shadow mode (data collection only)
		HybridRatio:                 20,
		DailyBudgetUSD:              1.00,
		MonthlyBudgetUSD:            30.00,
		SuccessThreshold:            7.0,
		AutoQualityWeight:           0.7,
		UserQualityWeight:           0.3,
		ExplorationRate:             0.2,
		UCBConfidence:               2.0,
		MinSamplesForRecommendation: 10,
		PromptStorageLevel:          "hash",
		EnableContextInjection:      true,
		EnablePromptEnhancement:     true,
		EnableCostOptimization:      true,
		EnableFallbackChain:         true,
		RetentionRawDays:            90,
		RetentionAggregatesDays:     365,
		ShareAggregatedAnalytics:    false,
		BanditPriorAlpha:            1.0,
		BanditPriorBeta:             1.0,
	}
}

// Stats holds aggregated statistics.
type Stats struct {
	TotalSamples      int64       `json:"total_samples"`
	AvgQuality        float64     `json:"avg_quality"`
	SuccessRate       float64     `json:"success_rate"`
	TotalCostUSD      float64     `json:"total_cost_usd"`
	CostSavingsUSD    float64     `json:"cost_savings_usd"`
	TopModels         []ModelStat `json:"top_models"`
	PatternCount      int         `json:"pattern_count"`
	BanditArms        int         `json:"bandit_arms"`
	BudgetUsedDaily   float64     `json:"budget_used_daily"`
	BudgetUsedMonthly float64     `json:"budget_used_monthly"`
}

// ModelStat tracks model performance.
type ModelStat struct {
	Model     string  `json:"model"`
	Provider  string  `json:"provider"`
	Quality   float64 `json:"quality"`
	Successes int     `json:"successes"`
	Samples   int     `json:"samples"`
}

// RouteDecision is the output of model selection.
type RouteDecision struct {
	Provider        string   `json:"provider"`
	Model           string   `json:"model"`
	ExpectedCost    float64  `json:"expected_cost"`
	ExpectedLatency int64    `json:"expected_latency_ms"`
	Reason          string   `json:"reason"`
	Confidence      float64  `json:"confidence"` // 0-1
	FallbackChain   []string `json:"fallback_chain"`
	ContextToInject []string `json:"context_to_inject"`
}

// HealthStatus reports engine health.
type HealthStatus struct {
	Status           string    `json:"status"` // healthy, degraded, down
	Uptime           string    `json:"uptime"`
	DBSizeBytes      int64     `json:"db_size_bytes"`
	LastFlush        time.Time `json:"last_flush"`
	BanditArms       int       `json:"bandit_arms"`
	PatternCacheHits int       `json:"pattern_cache_hits"`
	PatternCacheMiss int       `json:"pattern_cache_misses"`
	Warnings         []string  `json:"warnings,omitempty"`
}
