package learn

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Engine is the main BEADS v2 learning engine.
type Engine struct {
	mu               sync.RWMutex
	config           Config
	storage          *Storage
	bandit           *Bandit
	costOptimizer    *CostTracker
	patternExtractor *PatternExtractor
	contextFinder    *ContextFinder
	cache            *Cache

	// In-memory aggregates (hot reads)
	statsCache   *Stats
	statsCacheAt time.Time
	statsTTL     time.Duration
	lastFlush    time.Time
}

// NewEngine creates a new learning engine.
func NewEngine(config Config, storage *Storage) (*Engine, error) {
	if storage == nil {
		var err error
		storage, err = NewStorage("")
		if err != nil {
			return nil, fmt.Errorf("create storage: %w", err)
		}
	}

	engine := &Engine{
		config:           config,
		storage:          storage,
		bandit:           NewBandit(config.BanditPriorAlpha, config.BanditPriorBeta),
		costOptimizer:    NewCostTracker(BudgetLimits{config.DailyBudgetUSD, config.MonthlyBudgetUSD}),
		patternExtractor: NewPatternExtractor(storage),
		contextFinder:    NewContextFinder(nil), // TODO: inject QA store
		cache:            NewCache(),
		statsCache:       &Stats{},
		statsCacheAt:     time.Time{}, // zero = never cached
		statsTTL:         30 * time.Second,
	}

	// Load persisted state
	if err := engine.load(); err != nil {
		fmt.Printf("Warning: failed to load learning state: %v\n", err)
	}

	return engine, nil
}

// load loads bandit arms and patterns from storage.
func (e *Engine) load() error {
	// Load bandit from model_performance table
	rows, err := e.storage.db.Query(`
		SELECT provider, model, task_type, sample_count, avg_quality
		FROM model_performance
		WHERE sample_count > 0
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	statsMap := make(map[string]struct {
		Samples    int
		AvgQuality float64
	})

	for rows.Next() {
		var provider, model, taskType string
		var samples int
		var avgQuality float64
		if err := rows.Scan(&provider, &model, &taskType, &samples, &avgQuality); err != nil {
			continue
		}
		key := fmt.Sprintf("%s:%s:%s", provider, model, taskType)
		statsMap[key] = struct {
			Samples    int
			AvgQuality float64
		}{samples, avgQuality}
	}

	e.bandit.LoadFromStats(statsMap)

	// Load patterns
	_ = e.patternExtractor.LoadFromDB()

	return nil
}

// AnalyzePrompt analyzes a prompt for routing decisions.
func (e *Engine) AnalyzePrompt(prompt string) TaskAnalysis {
	return NewTaskClassifier().Classify(prompt)
}

// SelectModel chooses the best model for a given task.
func (e *Engine) SelectModel(analysis TaskAnalysis, prompt string) RouteDecision {
	e.mu.RLock()
	mode := e.config.Mode
	e.mu.RUnlock()

	// If disabled → use default (let router decide)
	if mode == "off" {
		return RouteDecision{
			Provider:   "default",
			Reason:     "learning disabled",
			Confidence: 0.0,
		}
	}

	// Check budget first (if cost optimization enabled)
	if e.config.EnableCostOptimization {
		canAfford := e.costOptimizer.CanAfford(0.02, 3) // Assume mid-cost
		if !canAfford {
			// Budget tight → free tier
			return RouteDecision{
				Provider:     "groq",
				Model:        "llama-3.1-70b",
				ExpectedCost: 0.0,
				Reason:       "budget constraints - using free tier",
				Confidence:   0.8,
			}
		}
	}

	// Get top models for task type from cache
	candidates := e.cache.GetTopModels(analysis.TaskType)
	if len(candidates) == 0 {
		candidates = []string{"openrouter:claude-sonnet-4"} // fallback
	}

	// Bandit selection
	var selectedKey string

	switch mode {
	case "shadow":
		// Always use default routing, but log decision
		selectedKey = candidates[0] // Record what WOULD have been chosen
	case "hybrid":
		// X% use bandit, (100-X)% use default
		if rand.Float64() < float64(e.config.HybridRatio)/100.0 {
			selectedKey = e.bandit.Select(e.config.ExplorationRate)
		} else {
			selectedKey = candidates[0]
		}
	case "auto":
		selectedKey = e.bandit.Select(e.config.ExplorationRate)
	default:
		selectedKey = candidates[0]
	}

	// Parse selected key
	parts := strings.SplitN(selectedKey, ":", 2)
	if len(parts) != 2 {
		selectedKey = "openrouter:claude-sonnet-4"
		parts = strings.SplitN(selectedKey, ":", 2)
	}

	provider := parts[0]
	model := parts[1]

	// Estimate cost (lookup from model registry)
	expectedCost := estimateCost(provider, model)
	expectedLatency := estimateLatency(provider, model, analysis.Complexity)

	// Get context to inject
	var contextToInject []string
	if e.config.EnableContextInjection {
		similar := e.contextFinder.FindSimilar(prompt, analysis.TaskType, 3)
		for _, item := range similar {
			ctx := fmt.Sprintf("[Session %s] %s", item.SessionID[:8], item.Content)
			contextToInject = append(contextToInject, ctx)
		}
	}

	return RouteDecision{
		Provider:        provider,
		Model:           model,
		ExpectedCost:    expectedCost,
		ExpectedLatency: expectedLatency,
		Reason:          fmt.Sprintf("Bandit selection: %.1f%% win rate", e.bandit.GetWinRate(selectedKey)*100),
		Confidence:      e.bandit.GetWinRate(selectedKey),
		FallbackChain:   getFallbackChain(provider, model),
		ContextToInject: contextToInject,
	}
}

// RecordOutcome logs the result and updates learning.
func (e *Engine) RecordOutcome(entry *LogInput, outcome Outcome, quality QualityScores) error {
	// Update quality
	entry.QualityUser = quality.User
	entry.QualityAuto = quality.Auto
	entry.QualityCombined = quality.Combined
	entry.Outcome = outcome.Outcome
	entry.OutcomeReason = outcome.Reason
	entry.CompilationSuccess = outcome.CompilationSuccess
	entry.TestRunSuccess = outcome.TestRunSuccess
	entry.TestPassRate = outcome.TestPassRate
	entry.GitCommitted = outcome.GitCommitted
	entry.UserEdited = outcome.UserEdited
	entry.IterationsNeeded = outcome.IterationsNeeded

	// 1. Store raw log (async doesn't block)
	if err := e.storage.InsertLog(entry); err != nil {
		fmt.Printf("Warning: failed to log: %v\n", err)
	}

	// 2. Update bandit
	threshold := e.config.SuccessThreshold
	qualityScore := quality.Final / 10.0 // 0-1
	e.bandit.Update(fmt.Sprintf("%s:%s", entry.Provider, entry.Model), qualityScore, threshold/10.0)

	// 3. Update cost tracker
	e.costOptimizer.AddSpent(entry.CostUSD)

	// 4. Extract pattern
	if e.config.EnablePromptEnhancement {
		e.patternExtractor.Process(entry)
	}

	// 5. Index context (session)
	e.contextFinder.IndexSession(entry.SessionID, entry.TaskType, entry.TaskDomain, entry.Prompt, entry.Model, quality.Final)

	// 6. Invalidate cache
	e.invalidateCache()

	// 7. Persist (async, debounced)
	go e.schedulePersist()

	return nil
}

// GetStats returns aggregated statistics.
func (e *Engine) GetStats() *Stats {
	e.mu.RLock()
	// Check cache
	if !e.statsCacheAt.IsZero() && time.Since(e.statsCacheAt) < e.statsTTL {
		cache := *e.statsCache
		e.mu.RUnlock()
		return &cache
	}
	e.mu.RUnlock()

	// Query storage
	stats, err := e.storage.GetStats(10)
	if err != nil {
		return &Stats{}
	}

	// Augment with runtime stats
	e.mu.RLock()
	stats.BanditArms = e.bandit.GetArmCount()
	stats.BudgetUsedDaily = e.costOptimizer.GetDailyUsed()
	stats.BudgetUsedMonthly = e.costOptimizer.GetMonthlyUsed()
	stats.PatternCount = len(e.patternExtractor.GetAllPatterns())
	e.mu.RUnlock()

	// Update cache
	e.mu.Lock()
	e.statsCache = stats
	e.statsCacheAt = time.Now()
	e.mu.Unlock()

	return stats
}

// GetTopPatterns returns top successful patterns.
func (e *Engine) GetTopPatterns(limit int) []Pattern {
	return e.patternExtractor.GetAllPatterns()[:limit]
}

// GetModelRecommendation returns best model for task type.
func (e *Engine) GetModelRecommendation(taskType string) (string, float64, int) {
	// Query from bandit stats
	stats := e.bandit.GetAllStats()

	var bestModel string
	var bestWinRate float64
	var bestSamples int

	for key, stat := range stats {
		parts := strings.Split(key, ":")
		if len(parts) < 3 {
			continue
		}
		if parts[2] != taskType {
			continue
		}
		total := stat.Wins + stat.Losses
		if total < e.config.MinSamplesForRecommendation {
			continue // Not enough data
		}
		if stat.WinRate > bestWinRate {
			bestWinRate = stat.WinRate
			bestModel = parts[1]
			bestSamples = total
		}
	}

	return bestModel, bestWinRate, bestSamples
}

// GetBanditStats returns bandit statistics.
func (e *Engine) GetBanditStats() map[string]struct {
	Wins    int
	Losses  int
	WinRate float64
} {
	return e.bandit.GetAllStats()
}

// Save persists engine state.
func (e *Engine) Save() error {
	return e.persist()
}

// Reset clears all learning data.
func (e *Engine) Reset(keepConfig bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear bandit
	e.bandit.Reset()

	// Clear patterns map
	e.patternExtractor.mu.Lock()
	e.patternExtractor.patterns = make(map[string]*Pattern)
	e.patternExtractor.mu.Unlock()

	// Clear cache
	*e.cache = *NewCache()

	// Wipe tables (except config if keepConfig)
	if !keepConfig {
		if _, err := e.storage.db.Exec(`DELETE FROM beads_logs`); err != nil {
			return err
		}
		if _, err := e.storage.db.Exec(`DELETE FROM prompt_patterns`); err != nil {
			return err
		}
		if _, err := e.storage.db.Exec(`DELETE FROM model_performance`); err != nil {
			return err
		}
	}

	return nil
}

// Config returns current config.
func (e *Engine) Config() Config {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config
}

// SetConfig updates config.
func (e *Engine) SetConfig(cfg Config) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config = cfg
}

// Enable/Lifecycle
func (e *Engine) IsEnabled() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config.Mode != "off"
}

func (e *Engine) GetMode() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config.Mode
}

func (e *Engine) SetMode(mode string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.Mode = mode
}

func (e *Engine) GetHybridRatio() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config.HybridRatio
}

func (e *Engine) SetHybridRatio(pct int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if pct < 0 {
		pct = 0
	} else if pct > 100 {
		pct = 100
	}
	e.config.HybridRatio = pct
}

// Health checks engine health.
func (e *Engine) Health() HealthStatus {
	h := HealthStatus{
		Status:           "healthy",
		DBSizeBytes:      0,
		BanditArms:       e.bandit.GetArmCount(),
		PatternCacheHits: e.cache.hits,
		PatternCacheMiss: e.cache.misses,
	}

	// Check DB connectivity
	if _, err := e.storage.db.Exec(`SELECT 1`); err != nil {
		h.Status = "degraded"
		h.Warnings = append(h.Warnings, "DB query failed: "+err.Error())
	}

	// DB size
	size, _ := e.storage.GetDBSize()
	h.DBSizeBytes = size

	// Check patterns loaded
	if h.BanditArms == 0 {
		h.Warnings = append(h.Warnings, "No bandit arms loaded (cold start)")
	}

	// Last flush
	h.LastFlush = e.lastFlush

	return h
}

// invalidateCache clears cached stats.
func (e *Engine) invalidateCache() {
	e.mu.Lock()
	e.statsCacheAt = time.Time{} // Force refresh on next GetStats
	e.cache.Invalidate()
	e.mu.Unlock()
}

// schedulePersist debounces persistence to every 5s or 10 ops.
func (e *Engine) schedulePersist() {
	// TODO: Implement debounced persist
	_ = e.persist()
}

// persist writes engine state to storage.
func (e *Engine) persist() error {
	// Currently: bandit arms + patterns already in DB via other paths
	// Just persist config maybe
	return nil
}

// estimateCost approximates cost for a model using typical token count (internal).
func estimateCost(provider, model string) float64 {
	// Simplified rates (USD/1M tokens) — approximate for typical 4k tokens
	rates := map[string]float64{
		"claude-sonnet-4": 3.0,
		"claude-opus-4":   15.0,
		"gpt-4.1":         2.0,
		"gpt-4o":          2.5,
		"llama-3.1-70b":   0.0, // Free via Groq
		"gemini-2.5":      0.0, // Free tier
	}
	typicalTokens := 4000.0
	costPerToken := 0.0
	if v, ok := rates[model]; ok {
		costPerToken = v / 1_000_000.0
	} else {
		costPerToken = 0.001 / 1_000.0 // Default $0.001/1k
	}
	return typicalTokens * costPerToken
}

// CalculateCost computes the actual cost in USD for a model given token counts.
// provider is kept for signature compatibility but not used currently.
func CalculateCost(provider, model string, tokensIn, tokensOut int) float64 {
	// Rates in USD per 1M tokens (input+output combined rate; can split if needed)
	rates := map[string]float64{
		"claude-sonnet-4":   3.00,
		"claude-opus-4":     15.00,
		"gpt-4.1":           2.00,
		"gpt-4o":            2.50,
		"llama-3.1-70b":     0.00,
		"gemini-2.5":        0.00,
		"deepseek-reasoner": 0.0,
	}
	rate, ok := rates[model]
	if !ok {
		rate = 0.001 // default $0.001 per 1M tokens (very cheap placeholder)
	}
	costPerToken := rate / 1_000_000.0
	total := float64(tokensIn+tokensOut) * costPerToken
	return total
}

// estimateLatency approximates response time.
func estimateLatency(provider, model string, complexity float64) int64 {
	latencies := map[string]int64{
		"claude-sonnet-4": 2000,
		"claude-opus-4":   3000,
		"gpt-4.1":         1500,
		"llama-3.1-70b":   500, // Groq fast
		"gemini-2.0":      1200,
	}
	base, ok := latencies[model]
	if !ok {
		base = 2000
	}
	// Complexity penalty
	penalty := int64(complexity * 200)
	return base + penalty
}

// getFallbackChain returns fallback models.
func getFallbackChain(primaryProvider, primaryModel string) []string {
	// Rules per provider
	chains := map[string][]string{
		"openrouter": {"openrouter:claude-sonnet-4", "openrouter:gpt-4.1", "groq:llama-3.1-70b"},
		"groq":       {"groq:llama-3.1-70b", "openrouter:claude-sonnet-4", "deepseek:deepseek-chat"},
		"cerebras":   {"cerebras:llama-3.1-70b", "groq:llama-3.1-70b"},
		"anthropic":  {"anthropic:claude-sonnet-4", "openrouter:claude-sonnet-4"},
		"deepseek":   {"deepseek:deepseek-chat", "groq:llama-3.1-70b"},
	}
	chain, ok := chains[primaryProvider]
	if !ok {
		chain = []string{primaryProvider + ":" + primaryModel}
	}
	return chain
}

// movingAverage updates running average.
func movingAverage(oldAvg, newVal float64) float64 {
	n := 100.0 // Placeholder, actual n tracked separately
	return (oldAvg*(n-1) + newVal) / n
}
