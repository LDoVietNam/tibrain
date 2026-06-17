package brain

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ─── Ingestion ───

// IngestFromFile reads router logs from a JSON/JSONL file.
func (e *Engine) IngestFromFile(ctx context.Context, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read log file %s: %w", filePath, err)
	}

	var logs []RouterLogEntry
	if err := json.Unmarshal(data, &logs); err != nil {
		// Try JSON lines format
		return e.ingestFromJSONLines(filePath)
	}

	return e.IngestLogs(ctx, logs)
}

// ingestFromJSONLines reads logs in JSON lines format.
func (e *Engine) ingestFromJSONLines(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var logs []RouterLogEntry
	count := 0

	for decoder.More() && count < e.maxLogsPerIngest {
		var log RouterLogEntry
		if err := decoder.Decode(&log); err != nil {
			break
		}
		logs = append(logs, log)
		count++
	}

	return e.IngestLogs(context.Background(), logs)
}

// IngestFromDir reads all log files from a directory.
func (e *Engine) IngestFromDir(ctx context.Context, dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("read log directory: %w", err)
	}

	ingested := 0
	for _, entry := range entries {
		if ingested >= e.maxLogsPerIngest {
			break
		}
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".json") && !strings.HasSuffix(entry.Name(), ".jsonl")) {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		if err := e.IngestFromFile(ctx, filePath); err != nil {
			continue
		}
		ingested++
	}

	return nil
}

// IngestBrainFeed loads and processes the brain-feed.jsonl file generated from beads.
// This is called on startup and when a reload signal is received.
func (e *Engine) IngestBrainFeed(ctx context.Context, feedPath string) error {
	if _, err := os.Stat(feedPath); os.IsNotExist(err) {
		// Feed doesn't exist yet - not an error
		return nil
	}

	// Check reload flag - if present and newer than feed, skip ingest
	reloadFlag := filepath.Join(filepath.Dir(feedPath), ".reload")
	if flagInfo, err := os.Stat(reloadFlag); err == nil {
		if feedInfo, err := os.Stat(feedPath); err == nil {
			if flagInfo.ModTime().After(feedInfo.ModTime()) {
				// Flag is newer than feed, means reload already processed
				os.Remove(reloadFlag)
				return nil
			}
		}
	}

	err := e.ingestFromJSONLines(feedPath)
	if err != nil {
		return fmt.Errorf("ingest brain feed: %w", err)
	}

	// Clear reload flag if present
	os.Remove(reloadFlag)

	return nil
}

// IngestLogs analyzes logs in parallel and extracts learnings.
func (e *Engine) IngestLogs(ctx context.Context, logs []RouterLogEntry) error {
	if len(logs) == 0 {
		return nil
	}

	// Limit batch size
	if len(logs) > e.maxLogsPerIngest {
		logs = logs[:e.maxLogsPerIngest]
	}

	// Classify + score each log
	for i := range logs {
		if logs[i].TaskType == "" {
			logs[i].TaskType = ClassifyTask(logs[i].UserPrompt)
		}
		if logs[i].Score == 0 {
			logs[i].Score = EvaluateResponse(logs[i].Success, logs[i].LatencyMs, 0)
		}
	}

	// Process logs in parallel
	e.processLogsInParallel(logs)

	e.mu.Lock()
	e.logs = append(e.logs, logs...)
	e.logCount += len(logs)
	e.lastIngestAt = time.Now()
	e.mu.Unlock()

	// Auto-save
	if e.dataDir != "" {
		e.mu.RLock()
		_ = e.saveLearningsLocked()
		e.mu.RUnlock()
	}

	// Apply RL + fine-tuning (these acquire their own lock)
	e.applyReinforcementLearning()
	e.fineTuneSystem()

	return nil
}

// processLogsInParallel processes logs concurrently using worker pool.
func (e *Engine) processLogsInParallel(logs []RouterLogEntry) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, e.workerCount)

	for _, log := range logs {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(log RouterLogEntry) {
			defer wg.Done()
			defer func() { <-semaphore }()
			e.processLog(log)
		}(log)
	}

	wg.Wait()
}

// processLog analyzes a single log entry and updates learned patterns.
func (e *Engine) processLog(log RouterLogEntry) {
	taskType := log.TaskType
	// taskType is pre-classified by IngestLogs before dispatching to workers

	e.mu.Lock()
	defer e.mu.Unlock()

	// Update provider (model) scores by task type
	if log.Model != "" && taskType != "" {
		if e.taskProviderScore[taskType] == nil {
			e.taskProviderScore[taskType] = make(map[string]*ProviderScore)
		}

		ps, exists := e.taskProviderScore[taskType][log.Model]
		if !exists {
			ps = &ProviderScore{}
			e.taskProviderScore[taskType][log.Model] = ps
		}

		ps.TotalCalls++
		if log.Success {
			ps.SuccessCalls++
		} else {
			ps.FailCalls++
		}

		// Running average for score
		n := float64(ps.TotalCalls)
		ps.AvgScore = (ps.AvgScore*(n-1) + log.Score) / n
		ps.LastUpdated = time.Now()
	}

	// Update provider metrics
	if log.Model != "" {
		pm, exists := e.providerMetrics[log.Model]
		if !exists {
			pm = &ProviderMetrics{
				Name:          log.Model,
				TaskBreakdown: make(map[string]int),
			}
			e.providerMetrics[log.Model] = pm
		}

		pm.TotalCalls++
		pm.TaskBreakdown[taskType]++
		pm.LastUsed = time.Now()

		// Update averages
		n := float64(pm.TotalCalls)
		pm.SuccessRate = (pm.SuccessRate*(n-1) + map[bool]float64{true: 1.0, false: 0.0}[log.Success]) / n
		pm.AvgLatencyMs = (pm.AvgLatencyMs*(n-1) + float64(log.LatencyMs)) / n
		pm.AvgScore = (pm.AvgScore*(n-1) + log.Score) / n
		pm.TotalCost += log.Cost
	}

	// Track context thresholds (success → turns = tool calls + 1)
	if taskType != "" && log.Success {
		current := e.contextThresholds[taskType]
		turns := 1 // Default 1 turn for simple tasks
		if current == 0 {
			e.contextThresholds[taskType] = turns
		} else {
			e.contextThresholds[taskType] = (current + turns) / 2
		}
	}
}
