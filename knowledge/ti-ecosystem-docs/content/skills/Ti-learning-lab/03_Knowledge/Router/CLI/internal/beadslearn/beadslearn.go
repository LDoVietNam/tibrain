// Package beadslearn connects BEADS logs to Memory scoring.
// When a BEADS entry completes (success/fail), this package:
// 1. Updates drawer importance based on task quality
// 2. Stores QA pairs for learning
// 3. Records model performance stats for Brain routing
package beadslearn

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ti/cli/internal/beads"
	"github.com/ti/cli/internal/memory"
)

// Learn connects BEADS logs to Memory scoring and model performance tracking.
type Learn struct {
	mu           sync.RWMutex
	palace       *memory.Palace
	modelStats   map[string]*ModelPerformance // model → performance history
	beadsPath    string
	savedStatsAt time.Time
}

// ModelPerformance tracks how well a model performs on different task types.
type ModelPerformance struct {
	Model          string    `json:"model"`
	TaskType       string    `json:"task_type"`
	Phase          string    `json:"phase"`
	TotalTasks     int       `json:"total_tasks"`
	SuccessCount   int       `json:"success_count"`
	FailCount      int       `json:"fail_count"`
	TotalQuality   float64   `json:"total_quality"`
	TotalCost      float64   `json:"total_cost"`
	TotalLatency   int64     `json:"total_latency_ms"`
	LastUsed       string    `json:"last_used"`
	QualityHistory []float64 `json:"quality_history"` // last 50 entries
}

// NewLearn creates a BEADS→Memory learning connector.
func NewLearn(palace *memory.Palace, beadsPath string) *Learn {
	l := &Learn{
		palace:     palace,
		modelStats: make(map[string]*ModelPerformance),
		beadsPath:  beadsPath,
	}
	l.loadStats()
	return l
}

// ProcessBEADSEnty is called when a BEADS entry is logged.
// Updates memory scoring and model performance tracking.
func (l *Learn) ProcessBEADSEntry(entry beads.Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Only process completed/failed tasks (not "running")
	if entry.Status != "complete" && entry.Status != "failed" {
		return nil
	}

	// Update model performance stats
	l.updateModelStats(entry)

	// Score relevant drawers in Memory
	l.scoreDrawers(entry)

	// Add QA drawer if task was successful and quality is high
	if entry.Success && entry.Quality >= 0.7 {
		l.storeQADrawer(entry)
	}

	// Save stats periodically (every 10 entries or 5 minutes)
	now := time.Now()
	if now.Sub(l.savedStatsAt) > 5*time.Minute {
		l.saveStats()
		l.savedStatsAt = now
	}

	return nil
}

// updateModelStats tracks how well each model performs per task type.
func (l *Learn) updateModelStats(entry beads.Entry) {
	if entry.Model == "" {
		return
	}

	key := entry.Model + ":" + entry.TaskType
	mp, ok := l.modelStats[key]
	if !ok {
		mp = &ModelPerformance{
			Model:    entry.Model,
			TaskType: entry.TaskType,
			Phase:    entry.Phase,
		}
		l.modelStats[key] = mp
	}

	mp.TotalTasks++
	if entry.Success {
		mp.SuccessCount++
	} else {
		mp.FailCount++
	}
	mp.TotalQuality += entry.Quality
	mp.TotalCost += entry.Cost
	mp.TotalLatency += entry.LatencyMs
	mp.LastUsed = entry.Timestamp

	// Keep last 50 quality entries for trend analysis
	if len(mp.QualityHistory) >= 50 {
		mp.QualityHistory = mp.QualityHistory[1:]
	}
	mp.QualityHistory = append(mp.QualityHistory, entry.Quality)
}

// scoreDrawers updates drawer importance based on task outcome.
// If task succeeded with high quality → boost related drawers.
// If task failed → reduce importance of related drawers.
func (l *Learn) scoreDrawers(entry beads.Entry) {
	if l.palace == nil {
		return
	}

	// Find drawers related to this task's domain/phase
	taxonomy := l.palace.GetTaxonomy()
	matched := false
	for wingName, rooms := range taxonomy {
		for roomName := range rooms {
			// Check if room/domain matches task
			if strings.Contains(strings.ToLower(roomName), strings.ToLower(entry.Domain)) ||
				strings.Contains(strings.ToLower(roomName), strings.ToLower(entry.Phase)) ||
				strings.Contains(strings.ToLower(roomName), strings.ToLower(entry.TaskType)) {

				// Get drawers in this room and update importance
				l.adjustRoomDrawers(wingName, roomName, entry)
				matched = true
			}
		}
	}

	// If no matching room found, add to default wing/room
	if !matched {
		wing := "default"
		if entry.Project != "" {
			wing = entry.Project
		}
		room := entry.Domain
		if room == "" {
			room = entry.TaskType
		}
		if room == "" {
			room = "general"
		}
		l.adjustRoomDrawers(wing, room, entry)
	}
}

// adjustRoomDrawers updates drawer importance based on task outcome.
func (l *Learn) adjustRoomDrawers(wing, room string, entry beads.Entry) {
	// We can't directly access drawers without exposing internals,
	// so we add a new "learning" drawer with the outcome
	if entry.Status == "complete" && entry.Quality >= 0.7 {
		d := memory.Drawer{
			Text: fmt.Sprintf("Task '%s' completed with quality %.2f using %s (cost: $%.4f, latency: %dms)",
				entry.Task, entry.Quality, entry.Model, entry.Cost, entry.LatencyMs),
			Importance: entry.Quality,
			Category:   "milestone",
			SourceFile: entry.SessionID,
		}
		l.palace.AddDrawer(wing, room, d)
	} else if entry.Status == "failed" {
		d := memory.Drawer{
			Text: fmt.Sprintf("Task '%s' FAILED with %s: %s",
				entry.Task, entry.Model, entry.Error),
			Importance: 0.3,
			Category:   "problem",
			SourceFile: entry.SessionID,
		}
		l.palace.AddDrawer(wing, room, d)
	}
}

// storeQADrawer creates a memory drawer from a successful high-quality task.
func (l *Learn) storeQADrawer(entry beads.Entry) {
	if l.palace == nil {
		return
	}

	// Determine wing/room from task context
	wing := "default"
	if entry.Project != "" {
		wing = entry.Project
	}
	room := entry.Domain
	if room == "" {
		room = entry.TaskType
	}
	if room == "" {
		room = "general"
	}

	// Create QA drawer with full context
	d := memory.Drawer{
		Text: fmt.Sprintf(
			"Q: %s\nA: Completed successfully\nModel: %s\nPhase: %s\nQuality: %.2f\nCost: $%.4f\nLatency: %dms\nTokens: %d in / %d out",
			entry.Task, entry.Model, entry.Phase,
			entry.Quality, entry.Cost, entry.LatencyMs,
			entry.PromptTokens, entry.CompletionTokens,
		),
		Importance: entry.Quality,
		Category:   "decision",
		SourceFile: entry.SessionID,
	}

	l.palace.AddDrawer(wing, room, d)
}

// GetModelPerformance returns performance stats for a model on a task type.
func (l *Learn) GetModelPerformance(model, taskType string) *ModelPerformance {
	l.mu.RLock()
	defer l.mu.RUnlock()
	key := model + ":" + taskType
	mp, ok := l.modelStats[key]
	if !ok {
		return nil
	}
	// Return copy to avoid mutation
	cp := *mp
	cp.QualityHistory = make([]float64, len(mp.QualityHistory))
	copy(cp.QualityHistory, mp.QualityHistory)
	return &cp
}

// GetBestModel returns the best performing model for a given task type.
// Best = highest average quality with at least 3 data points.
func (l *Learn) GetBestModel(taskType string) (string, float64, int) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var bestModel string
	var bestAvgQuality float64
	var bestTasks int

	for key, mp := range l.modelStats {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) != 2 || parts[1] != taskType {
			continue
		}
		if mp.TotalTasks < 3 {
			continue // Need minimum data points
		}
		avgQuality := mp.TotalQuality / float64(mp.TotalTasks)
		if avgQuality > bestAvgQuality {
			bestAvgQuality = avgQuality
			bestModel = parts[0]
			bestTasks = mp.TotalTasks
		}
	}

	return bestModel, bestAvgQuality, bestTasks
}

// GetAllModelStats returns a copy of all model performance data.
func (l *Learn) GetAllModelStats() map[string]*ModelPerformance {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make(map[string]*ModelPerformance, len(l.modelStats))
	for k, v := range l.modelStats {
		cp := *v
		cp.QualityHistory = make([]float64, len(v.QualityHistory))
		copy(cp.QualityHistory, v.QualityHistory)
		result[k] = &cp
	}
	return result
}

// GetQualityTrend returns the quality trend for a model over recent tasks.
// Returns positive if improving, negative if declining.
func (l *Learn) GetQualityTrend(model string) float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Collect all quality entries for this model
	var allQuality []float64
	for key, mp := range l.modelStats {
		if strings.HasPrefix(key, model+":") {
			allQuality = append(allQuality, mp.QualityHistory...)
		}
	}

	if len(allQuality) < 5 {
		return 0 // Not enough data
	}

	// Compare first half vs second half
	mid := len(allQuality) / 2
	firstHalf := average(allQuality[:mid])
	secondHalf := average(allQuality[mid:])

	return secondHalf - firstHalf
}

// Save persists model stats to disk.
func (l *Learn) Save() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.saveStats()
}

// Load reloads model stats from disk.
func (l *Learn) Load() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.loadStats()
}

// saveStats writes model performance data to disk.
func (l *Learn) saveStats() error {
	statsPath := l.statsFilePath()
	if err := os.MkdirAll(filepath.Dir(statsPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(l.modelStats, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal model stats: %w", err)
	}

	return os.WriteFile(statsPath, data, 0644)
}

// loadStats reads model performance data from disk.
func (l *Learn) loadStats() error {
	statsPath := l.statsFilePath()
	data, err := os.ReadFile(statsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // First run, no stats yet
		}
		return fmt.Errorf("read model stats: %w", err)
	}

	return json.Unmarshal(data, &l.modelStats)
}

// statsFilePath returns the path for model stats file.
func (l *Learn) statsFilePath() string {
	dir := filepath.Dir(l.beadsPath)
	return filepath.Join(dir, "model-stats.json")
}

// average computes the mean of a slice of float64.
func average(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

// GenerateSessionID creates a deterministic session ID from context.
func GenerateSessionID(project, task string) string {
	h := sha256.Sum256([]byte(project + ":" + task))
	return hex.EncodeToString(h[:8])
}
