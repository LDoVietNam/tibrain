package brain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ─── Persistence ───

// learningsData is the persistable state of the engine.
type learningsData struct {
	TaskProviderScore map[string]map[string]*ProviderScore `json:"task_provider_score"`
	ProviderMetrics   map[string]*ProviderMetrics          `json:"provider_metrics"`
	PreferredProvider map[string]string                    `json:"preferred_provider"`
	PromptPatterns    map[string]string                    `json:"prompt_patterns"`
	ContextThresholds map[string]int                       `json:"context_thresholds"`
	LogCount          int                                  `json:"log_count"`
	LastIngestAt      time.Time                            `json:"last_ingest_at"`
}

// saveLearnings persists learnings to disk.
func (e *Engine) saveLearnings() error {
	if e.dataDir == "" {
		return nil
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.saveLearningsLocked()
}

// saveLearningsLocked persists learnings to disk. Caller must hold e.mu.RLock.
func (e *Engine) saveLearningsLocked() error {
	if e.dataDir == "" {
		return nil
	}

	data := &learningsData{
		TaskProviderScore: e.taskProviderScore,
		ProviderMetrics:   e.providerMetrics,
		PreferredProvider: e.preferredProvider,
		PromptPatterns:    e.promptPatterns,
		ContextThresholds: e.contextThresholds,
		LogCount:          e.logCount,
		LastIngestAt:      e.lastIngestAt,
	}

	filePath := filepath.Join(e.dataDir, "learnings.json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal learnings: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("write learnings: %w", err)
	}

	return nil
}

// loadLearnings loads persisted learnings from disk.
func (e *Engine) loadLearnings() error {
	filePath := filepath.Join(e.dataDir, "learnings.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var persisted learningsData
	if err := json.Unmarshal(data, &persisted); err != nil {
		return fmt.Errorf("unmarshal learnings: %w", err)
	}

	e.mu.Lock()
	e.taskProviderScore = persisted.TaskProviderScore
	e.providerMetrics = persisted.ProviderMetrics
	e.preferredProvider = persisted.PreferredProvider
	e.promptPatterns = persisted.PromptPatterns
	e.contextThresholds = persisted.ContextThresholds
	e.logCount = persisted.LogCount
	e.lastIngestAt = persisted.LastIngestAt
	e.mu.Unlock()

	return nil
}

// ─── Brain State (full export/import) ───

// BrainState holds the complete brain state for export/import/backup.
type BrainState struct {
	Version           string             `json:"version"`
	ExportedAt        time.Time          `json:"exported_at"`
	Learnings         *learningsData     `json:"learnings"`
	ModelIntelligence *ModelIntelligence `json:"model_intelligence"`
}

// ExportJSON exports the complete brain state to a JSON file.
func (e *Engine) ExportJSON(filePath string) error {
	if filePath == "" {
		filePath = filepath.Join(e.dataDir, "brain_export.json")
	}

	e.mu.RLock()
	state := &BrainState{
		Version:    "1.0.0",
		ExportedAt: time.Now(),
		Learnings: &learningsData{
			TaskProviderScore: e.taskProviderScore,
			ProviderMetrics:   e.providerMetrics,
			PreferredProvider: e.preferredProvider,
			PromptPatterns:    e.promptPatterns,
			ContextThresholds: e.contextThresholds,
			LogCount:          e.logCount,
			LastIngestAt:      e.lastIngestAt,
		},
		ModelIntelligence: e.modelIntelligence,
	}
	e.mu.RUnlock()

	jsonData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal brain state: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("write brain export: %w", err)
	}

	return nil
}

// ImportJSON imports brain state from a JSON file.
func (e *Engine) ImportJSON(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read brain import file: %w", err)
	}

	var state BrainState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("unmarshal brain state: %w", err)
	}

	e.mu.Lock()
	if state.ModelIntelligence != nil {
		e.modelIntelligence = state.ModelIntelligence
	}
	if state.Learnings != nil {
		e.taskProviderScore = state.Learnings.TaskProviderScore
		e.providerMetrics = state.Learnings.ProviderMetrics
		e.preferredProvider = state.Learnings.PreferredProvider
		e.promptPatterns = state.Learnings.PromptPatterns
		e.contextThresholds = state.Learnings.ContextThresholds
		e.logCount = state.Learnings.LogCount
		e.lastIngestAt = state.Learnings.LastIngestAt
	}
	e.mu.Unlock()

	return nil
}

// Backup creates a timestamped backup of the brain state.
func (e *Engine) Backup(backupDir string) (string, error) {
	if backupDir == "" {
		backupDir = filepath.Join(e.dataDir, "backups")
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	filePath := filepath.Join(backupDir, fmt.Sprintf("brain_backup_%s.json", timestamp))

	// Safety: check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		for i := 1; i < 100; i++ {
			newPath := filepath.Join(backupDir, fmt.Sprintf("brain_backup_%s_%d.json", timestamp, i))
			if _, err := os.Stat(newPath); os.IsNotExist(err) {
				filePath = newPath
				break
			}
		}
	}

	if err := e.ExportJSON(filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

// ListBackups returns all brain backups sorted by name (newest first).
func (e *Engine) ListBackups(backupDir string) ([]string, error) {
	if backupDir == "" {
		backupDir = filepath.Join(e.dataDir, "backups")
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("read backup dir: %w", err)
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "brain_backup_") || !strings.HasSuffix(name, ".json") {
			continue
		}
		backups = append(backups, name)
	}

	// Reverse (newest first — names are timestamped)
	for i, j := 0, len(backups)-1; i < j; i, j = i+1, j-1 {
		backups[i], backups[j] = backups[j], backups[i]
	}

	return backups, nil
}

// Restore restores brain state from a backup file with pre-restore safety.
func (e *Engine) Restore(backupPath string) error {
	// Pre-restore safety: backup current state first
	currentBackup, err := e.Backup("")
	if err != nil {
		return fmt.Errorf("pre-restore backup failed: %w", err)
	}

	// Import the backup
	if err := e.ImportJSON(backupPath); err != nil {
		// Rollback: restore from the safety backup
		_ = e.ImportJSON(filepath.Join(filepath.Dir(currentBackup), filepath.Base(currentBackup)))
		return fmt.Errorf("restore failed, rolled back to safety backup: %w", err)
	}

	return nil
}
