package brain

import (
	"context"
	"fmt"
	"path/filepath"
)

// ─── Engine Lifecycle ───

// NewEngine creates a new Router Learning Engine.
func NewEngine(cfg Config) (*Engine, error) {
	if cfg.MaxLogsPerIngest <= 0 {
		cfg.MaxLogsPerIngest = 1000
	}
	if cfg.MinConfidence <= 0 {
		cfg.MinConfidence = 0.75
	}
	if cfg.LearningRate <= 0 {
		cfg.LearningRate = 0.1
	}
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4
	}

	e := &Engine{
		taskProviderScore: make(map[string]map[string]*ProviderScore),
		providerMetrics:   make(map[string]*ProviderMetrics),
		preferredProvider: make(map[string]string),
		promptPatterns:    make(map[string]string),
		contextThresholds: make(map[string]int),
		maxLogsPerIngest:  cfg.MaxLogsPerIngest,
		minConfidence:     cfg.MinConfidence,
		learningRate:      cfg.LearningRate,
		dataDir:           cfg.DataDir,
		workerCount:       cfg.WorkerCount,
		modelIntelligence: DefaultModelIntelligence(),
	}

	// Load persisted learnings if available
	if cfg.DataDir != "" {
		if err := e.loadLearnings(); err != nil {
			// No persisted learnings yet — not an error
			_ = err
		}
	}

	return e, nil
}

// LogCount returns total ingested logs.
func (e *Engine) LogCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.logCount
}

// GetPreferredProviders returns all learned task→model preferences.
func (e *Engine) GetPreferredProviders() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make(map[string]string, len(e.preferredProvider))
	for k, v := range e.preferredProvider {
		result[k] = v
	}
	return result
}

// GetProviderMetrics returns metrics for all models.
func (e *Engine) GetProviderMetrics() map[string]ProviderMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make(map[string]ProviderMetrics, len(e.providerMetrics))
	for k, v := range e.providerMetrics {
		pm := *v
		result[k] = pm
	}
	return result
}

// Reset clears all learned patterns (keeps config).
func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.taskProviderScore = make(map[string]map[string]*ProviderScore)
	e.providerMetrics = make(map[string]*ProviderMetrics)
	e.preferredProvider = make(map[string]string)
	e.promptPatterns = make(map[string]string)
	e.contextThresholds = make(map[string]int)
	e.logs = nil
	e.logCount = 0
}

// Save persists learnings to disk.
func (e *Engine) Save() error {
	if e.dataDir == "" {
		return fmt.Errorf("no data dir configured")
	}
	return e.saveLearnings()
}

// GetModelIntelligence returns the model intelligence config.
func (e *Engine) GetModelIntelligence() *ModelIntelligence {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.modelIntelligence
}

// SetDataDir updates the data directory and loads any existing learnings.
func (e *Engine) SetDataDir(dir string) error {
	e.mu.Lock()
	e.dataDir = dir
	e.mu.Unlock()

	if dir != "" {
		return e.loadLearnings()
	}
	return nil
}

// ─── Task Classification (public) ───

// GetTaskType classifies a prompt into a task type.
func GetTaskType(prompt string) string {
	return ClassifyTask(prompt)
}

// ─── Brain Feed Integration ───

// ReloadBrainFeed reloads the brain-feed.jsonl file if it exists.
// This should be called when a reload signal is received.
func (e *Engine) ReloadBrainFeed(ctx context.Context) error {
	if e.dataDir == "" {
		return fmt.Errorf("no data dir configured")
	}

	feedPath := filepath.Join(e.dataDir, "brain-feed.jsonl")
	return e.IngestBrainFeed(ctx, feedPath)
}

// InitializeWithBrainFeed creates a new engine and loads brain feed on startup.
func InitializeWithBrainFeed(cfg Config) (*Engine, error) {
	e, err := NewEngine(cfg)
	if err != nil {
		return nil, err
	}

	// Load brain feed on startup if data dir is configured
	if cfg.DataDir != "" {
		ctx := context.Background()
		feedPath := filepath.Join(cfg.DataDir, "brain-feed.jsonl")
		if err := e.IngestBrainFeed(ctx, feedPath); err != nil {
			// Log but don't fail - brain feed might not exist yet
			fmt.Printf("[WARN] Failed to load brain feed: %v\n", err)
		}
	}

	return e, nil
}
