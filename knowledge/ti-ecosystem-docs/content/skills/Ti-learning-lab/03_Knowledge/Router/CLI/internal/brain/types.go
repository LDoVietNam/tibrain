// Package brain provides Ti's Router Learning Engine.
// Learns from BEADS logs to improve routing decisions.
package brain

import (
	"sync"
	"time"
)

// ─── Ingestion Types ───

// RouterLogEntry represents a single router log entry from BEADS.
type RouterLogEntry struct {
	Timestamp  string  `json:"timestamp"`
	UserPrompt string  `json:"user_prompt"`
	Response   string  `json:"response"`
	TaskType   string  `json:"task_type"`
	Model      string  `json:"model"`
	Agent      string  `json:"agent"`
	Phase      string  `json:"phase"`
	Domain     string  `json:"domain"`
	Success    bool    `json:"success"`
	Score      float64 `json:"score"`
	LatencyMs  int64   `json:"latency_ms"`
	Cost       float64 `json:"cost"`
	Budget     float64 `json:"budget"`
	SessionID  string  `json:"session_id"`
}

// ─── Performance Tracking ───

// ProviderScore tracks per-provider performance per task type.
type ProviderScore struct {
	TotalCalls   int       `json:"total_calls"`
	SuccessCalls int       `json:"success_calls"`
	FailCalls    int       `json:"fail_calls"`
	AvgScore     float64   `json:"avg_score"`
	LastUpdated  time.Time `json:"last_updated"`
}

// ProviderMetrics tracks overall provider performance.
type ProviderMetrics struct {
	Name          string         `json:"name"`
	TotalCalls    int            `json:"total_calls"`
	SuccessRate   float64        `json:"success_rate"`
	AvgLatencyMs  float64        `json:"avg_latency_ms"`
	AvgScore      float64        `json:"avg_score"`
	TotalCost     float64        `json:"total_cost"`
	TaskBreakdown map[string]int `json:"task_breakdown"` // task_type → count
	LastUsed      time.Time      `json:"last_used"`
}

// TaskPattern represents an extracted pattern from logs.
type TaskPattern struct {
	TaskType     string  `json:"task_type"`
	SuccessRate  float64 `json:"success_rate"`
	AvgScore     float64 `json:"avg_score"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	AvgCost      float64 `json:"avg_cost"`
	SampleCount  int     `json:"sample_count"`
	BestProvider string  `json:"best_provider"`
	BestModel    string  `json:"best_model"`
}

// ─── Engine State ───

// Engine is the Router Learning Engine.
type Engine struct {
	mu sync.RWMutex

	// Learned patterns
	taskProviderScore map[string]map[string]*ProviderScore // task_type → provider → score
	providerMetrics   map[string]*ProviderMetrics          // provider → metrics
	preferredProvider map[string]string                    // task_type → preferred provider
	promptPatterns    map[string]string                    // task_type → tuned prompt
	contextThresholds map[string]int                       // task_type → optimal turns

	// Ingestion state
	logs         []RouterLogEntry
	logCount     int
	lastIngestAt time.Time

	// Config
	maxLogsPerIngest int
	minConfidence    float64
	learningRate     float64
	dataDir          string
	workerCount      int

	// Model intelligence
	modelIntelligence *ModelIntelligence
}

// Config holds learning engine configuration.
type Config struct {
	MaxLogsPerIngest int
	MinConfidence    float64
	LearningRate     float64
	DataDir          string
	WorkerCount      int
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxLogsPerIngest: 1000,
		MinConfidence:    0.75,
		LearningRate:     0.1,
		DataDir:          "",
		WorkerCount:      4,
	}
}
