// Package beads implements the BEADS logging system.
// BEADS = Basic Event And Decision Structure
// Logs ALL activity: prompt → response → quality score → learning feed.
//
// BEADS is the nervous system of Ti — without it, Brain is blind.
//
// Schema v2 — ML-grade learning dataset with:
//   - Features: phase, domain, complexity, budget (input context)
//   - Decision: model, reasoning, alternatives (routing choice)
//   - Outcome: quality, cost, latency, success, acceptance flags (reward signals)
//   - Version: schema version for forward/backward compatibility
package beads

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// SchemaVersion is the current BEADS entry schema version.
// Incremented when fields are added/changed in a way that breaks old readers.
const SchemaVersion = 2

// Entry represents a single BEADS log entry.
// Designed as a machine learning dataset:
//   Features (input context) → Decision (routing action) → Outcome (reward signal)
type Entry struct {
	// ── Schema ──────────────────────────────────────────────
	SchemaVersion int `json:"schema_version"` // 1 = legacy, 2 = ML-grade

	// ── Common ──────────────────────────────────────────────
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"` // "running", "complete", "failed", "timeout"

	// ── Features (input context — BEFORE routing) ──────────
	TaskType   string  `json:"task_type"`  // "coding", "review", "plan", "spec", "summarize"
	Phase      string  `json:"phase"`      // "scan", "plan", "spec", "implement", "review", "summarize"
	Domain     string  `json:"domain"`     // "auth", "config", "router", "cli"...
	Complexity float64 `json:"complexity"` // 0-1, estimated task complexity
	Budget     float64 `json:"budget"`     // max cost in USD the router was allowed to spend
	Task       string  `json:"task"`       // human-readable task description

	// ── Decision (routing action — WHY this model?) ────────
	Model        string   `json:"model"`        // selected model
	Agent        string   `json:"agent"`        // which agent handled the task
	Reasoning    string   `json:"reasoning"`    // why this model was chosen
	Alternatives []string `json:"alternatives"` // other models that were considered

	// ── Outcome (reward signal — AFTER execution) ──────────
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	Quality          float64 `json:"quality"` // 0-1, from eval or user feedback
	LatencyMs        int64   `json:"latency_ms"`
	Cost             float64 `json:"cost"`
	Success          bool    `json:"success"`

	// ── Acceptance flags (RL reward shaping) ───────────────
	QualityAcceptable bool `json:"quality_acceptable"` // quality met threshold
	CostAcceptable    bool `json:"cost_acceptable"`    // cost stayed within budget
	LatencyAcceptable bool `json:"latency_acceptable"` // latency was acceptable

	// ── Context (correlation) ──────────────────────────────
	SessionID string `json:"session_id"`
	Project   string `json:"project"`

	// ── Error ──────────────────────────────────────────────
	Error string `json:"error,omitempty"`
}

// ──────────────────────────────────────────────────────────────
// Logger — writes BEADS entries to JSONL file + callback
// ──────────────────────────────────────────────────────────────

// Logger writes BEADS entries to JSONL file + optional callback for Memory.
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	path     string
	callback func(entry Entry) // Called on each log entry (e.g., feed to Memory)
}

// Config for BEADS logger.
type Config struct {
	Path     string      // JSONL file path (default: ~/.ti/data/beads.jsonl)
	Callback func(Entry) // Optional callback for real-time processing
}

// NewLogger creates a BEADS logger.
func NewLogger(cfg Config) (*Logger, error) {
	if cfg.Path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		cfg.Path = filepath.Join(home, ".ti", "data", "beads.jsonl")
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(cfg.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:     f,
		path:     cfg.Path,
		callback: cfg.Callback,
	}, nil
}

// Log writes a BEADS entry (thread-safe, append-only JSONL).
func (l *Logger) Log(entry Entry) error {
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if entry.Status == "" {
		entry.Status = "running" // Default: task in progress
	}
	if entry.SchemaVersion == 0 {
		entry.SchemaVersion = SchemaVersion
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal BEADS entry: %w", err)
	}

	l.mu.Lock()
	_, err = l.file.Write(append(data, '\n'))
	if err != nil {
		l.mu.Unlock()
		return err
	}
	l.file.Sync() // Flush to disk immediately
	l.mu.Unlock()

	// Callback for real-time processing (e.g., feed to Memory/Brain)
	if l.callback != nil {
		go l.callback(entry)
	}

	return nil
}

// Update updates an existing BEADS entry by appending a new line with updated status.
// Used for logging task completion after initial "running" entry.
func (l *Logger) Update(entry Entry) error {
	entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	return l.Log(entry)
}

// Close the logger.
func (l *Logger) Close() error {
	return l.file.Close()
}

// ──────────────────────────────────────────────────────────────
// ReadEntries — reads BEADS entries from JSONL file
// ──────────────────────────────────────────────────────────────

// ReadEntries reads BEADS entries from the file.
func ReadEntries(path string, filter func(Entry) bool) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	dec := json.NewDecoder(f)
	for {
		var entry Entry
		if err := dec.Decode(&entry); err != nil {
			if err.Error() == "EOF" {
				break
			}
			// Skip malformed lines
			continue
		}
		if filter == nil || filter(entry) {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// ──────────────────────────────────────────────────────────────
// Query API — filter BEADS entries by various criteria
// ──────────────────────────────────────────────────────────────

// QueryFilter holds filter criteria for BEADS queries.
type QueryFilter struct {
	Phase      []string // filter by phase (match any)
	Domain     []string // filter by domain (match any)
	Agent      []string // filter by agent (match any)
	Model      []string // filter by model (match any)
	TaskType   []string // filter by task type (match any)
	Status     []string // filter by status (match any)
	MinQuality float64  // minimum quality (0-1)
	MaxCost    float64  // maximum cost in USD
	DateFrom   string   // ISO 8601 start date (e.g. "2026-04-10T00:00:00Z")
	DateTo     string   // ISO 8601 end date
}

// Query reads BEADS entries and applies filters.
func Query(path string, qf QueryFilter) ([]Entry, error) {
	return ReadEntries(path, func(e Entry) bool {
		if len(qf.Phase) > 0 && !containsStr(qf.Phase, e.Phase) {
			return false
		}
		if len(qf.Domain) > 0 && !containsStr(qf.Domain, e.Domain) {
			return false
		}
		if len(qf.Agent) > 0 && !containsStr(qf.Agent, e.Agent) {
			return false
		}
		if len(qf.Model) > 0 && !containsStr(qf.Model, e.Model) {
			return false
		}
		if len(qf.TaskType) > 0 && !containsStr(qf.TaskType, e.TaskType) {
			return false
		}
		if len(qf.Status) > 0 && !containsStr(qf.Status, e.Status) {
			return false
		}
		if qf.MinQuality > 0 && e.Quality < qf.MinQuality {
			return false
		}
		if qf.MaxCost > 0 && e.Cost > qf.MaxCost {
			return false
		}
		if qf.DateFrom != "" && e.Timestamp < qf.DateFrom {
			return false
		}
		if qf.DateTo != "" && e.Timestamp > qf.DateTo {
			return false
		}
		return true
	})
}

// ──────────────────────────────────────────────────────────────
// Stats — compute statistics from BEADS entries
// ──────────────────────────────────────────────────────────────

// Stats computes statistics from BEADS entries.
type Stats struct {
	TotalTasks        int                   `json:"total_tasks"`
	SuccessRate       float64               `json:"success_rate"`
	AvgQuality        float64               `json:"avg_quality"`
	AvgLatencyMs      float64               `json:"avg_latency_ms"`
	TotalCost         float64               `json:"total_cost"`
	TotalTokens       int                   `json:"total_tokens"`
	QualityAcceptRate float64               `json:"quality_accept_rate"`
	CostAcceptRate    float64               `json:"cost_accept_rate"`
	LatencyAcceptRate float64               `json:"latency_accept_rate"`
	ModelBreakdown    map[string]ModelStats `json:"model_breakdown"`
	PhaseBreakdown    map[string]PhaseStats `json:"phase_breakdown"`
	AgentBreakdown    map[string]AgentStats `json:"agent_breakdown"`
}

// ModelStats holds per-model statistics.
type ModelStats struct {
	Tasks        int     `json:"tasks"`
	Successes    int     `json:"successes"`
	AvgQuality   float64 `json:"avg_quality"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	TotalCost    float64 `json:"total_cost"`
}

// PhaseStats holds per-phase statistics.
type PhaseStats struct {
	Tasks      int     `json:"tasks"`
	Successes  int     `json:"successes"`
	AvgQuality float64 `json:"avg_quality"`
	AvgCost    float64 `json:"avg_cost"`
	TotalCost  float64 `json:"total_cost"`
}

// AgentStats holds per-agent statistics.
type AgentStats struct {
	Tasks        int     `json:"tasks"`
	Successes    int     `json:"successes"`
	AvgQuality   float64 `json:"avg_quality"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	TotalCost    float64 `json:"total_cost"`
}

// ComputeStats reads the BEADS file and computes statistics.
func ComputeStats(path string) (*Stats, error) {
	entries, err := ReadEntries(path, nil)
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		ModelBreakdown: make(map[string]ModelStats),
		PhaseBreakdown: make(map[string]PhaseStats),
		AgentBreakdown: make(map[string]AgentStats),
	}

	for _, e := range entries {
		stats.TotalTasks++
		if e.Success {
			stats.SuccessRate++
		}
		stats.AvgQuality += e.Quality
		stats.AvgLatencyMs += float64(e.LatencyMs)
		stats.TotalCost += e.Cost
		stats.TotalTokens += e.PromptTokens + e.CompletionTokens

		// Acceptance rates (only count if schema v2+)
		if e.SchemaVersion >= 2 {
			if e.QualityAcceptable {
				stats.QualityAcceptRate++
			}
			if e.CostAcceptable {
				stats.CostAcceptRate++
			}
			if e.LatencyAcceptable {
				stats.LatencyAcceptRate++
			}
		}

		// Per-model breakdown
		ms := stats.ModelBreakdown[e.Model]
		ms.Tasks++
		if e.Success {
			ms.Successes++
		}
		ms.AvgQuality += e.Quality
		ms.AvgLatencyMs += float64(e.LatencyMs)
		ms.TotalCost += e.Cost
		stats.ModelBreakdown[e.Model] = ms

		// Per-phase breakdown
		if e.Phase != "" {
			ps := stats.PhaseBreakdown[e.Phase]
			ps.Tasks++
			if e.Success {
				ps.Successes++
			}
			ps.AvgQuality += e.Quality
			ps.AvgCost += e.Cost
			ps.TotalCost += e.Cost
			stats.PhaseBreakdown[e.Phase] = ps
		}

		// Per-agent breakdown
		if e.Agent != "" {
			as := stats.AgentBreakdown[e.Agent]
			as.Tasks++
			if e.Success {
				as.Successes++
			}
			as.AvgQuality += e.Quality
			as.AvgLatencyMs += float64(e.LatencyMs)
			as.TotalCost += e.Cost
			stats.AgentBreakdown[e.Agent] = as
		}
	}

	// Averages
	if stats.TotalTasks > 0 {
		stats.SuccessRate /= float64(stats.TotalTasks)
		stats.AvgQuality /= float64(stats.TotalTasks)
		stats.AvgLatencyMs /= float64(stats.TotalTasks)
	}
	// Acceptance rates: only count v2 entries that have acceptance data
	v2Count := 0
	for _, e := range entries {
		if e.SchemaVersion >= 2 {
			v2Count++
		}
	}
	if v2Count > 0 {
		stats.QualityAcceptRate /= float64(v2Count)
		stats.CostAcceptRate /= float64(v2Count)
		stats.LatencyAcceptRate /= float64(v2Count)
	}
	for model, ms := range stats.ModelBreakdown {
		if ms.Tasks > 0 {
			ms.AvgQuality /= float64(ms.Tasks)
			ms.AvgLatencyMs /= float64(ms.Tasks)
			stats.ModelBreakdown[model] = ms
		}
	}
	for phase, ps := range stats.PhaseBreakdown {
		if ps.Tasks > 0 {
			ps.AvgQuality /= float64(ps.Tasks)
			ps.AvgCost /= float64(ps.Tasks)
			stats.PhaseBreakdown[phase] = ps
		}
	}
	for agent, as := range stats.AgentBreakdown {
		if as.Tasks > 0 {
			as.AvgQuality /= float64(as.Tasks)
			as.AvgLatencyMs /= float64(as.Tasks)
			stats.AgentBreakdown[agent] = as
		}
	}

	return stats, nil
}

// ──────────────────────────────────────────────────────────────
// SQLite Store — materialize BEADS entries for query-based learning
// ──────────────────────────────────────────────────────────────

// Store manages BEADS entries in SQLite for query-based learning.
// JSONL is kept for streaming; SQLite is for analytics and ML training.
type Store struct {
	db *sql.DB
}

// NewStore creates or opens a BEADS SQLite store.
func NewStore(path string) (*Store, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".ti", "data", "beads.db")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.initDB(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close the store.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB returns the underlying *sql.DB for direct access (testing).
func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) initDB() error {
	schema := `
	CREATE TABLE IF NOT EXISTS beads (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		schema_ver  INTEGER NOT NULL DEFAULT 2,
		timestamp   TEXT NOT NULL DEFAULT '',
		status      TEXT NOT NULL DEFAULT 'running',

		-- Features (input context)
		task_type   TEXT NOT NULL DEFAULT '',
		phase       TEXT NOT NULL DEFAULT '',
		domain      TEXT NOT NULL DEFAULT '',
		complexity  REAL NOT NULL DEFAULT 0,
		budget      REAL NOT NULL DEFAULT 0,
		task        TEXT NOT NULL DEFAULT '',

		-- Decision (routing action)
		model       TEXT NOT NULL DEFAULT '',
		agent       TEXT NOT NULL DEFAULT '',
		reasoning   TEXT NOT NULL DEFAULT '',
		alternatives TEXT NOT NULL DEFAULT '[]',

		-- Outcome (reward signal)
		prompt_tokens     INTEGER NOT NULL DEFAULT 0,
		completion_tokens INTEGER NOT NULL DEFAULT 0,
		quality           REAL NOT NULL DEFAULT 0,
		latency_ms        INTEGER NOT NULL DEFAULT 0,
		cost              REAL NOT NULL DEFAULT 0,
		success           INTEGER NOT NULL DEFAULT 0,

		-- Acceptance flags
		quality_acceptable  INTEGER NOT NULL DEFAULT 0,
		cost_acceptable     INTEGER NOT NULL DEFAULT 0,
		latency_acceptable  INTEGER NOT NULL DEFAULT 0,

		-- Context
		session_id  TEXT NOT NULL DEFAULT '',
		project     TEXT NOT NULL DEFAULT '',

		-- Error
		error       TEXT NOT NULL DEFAULT ''
	);

	CREATE INDEX IF NOT EXISTS idx_beads_timestamp ON beads(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_beads_phase ON beads(phase);
	CREATE INDEX IF NOT EXISTS idx_beads_model ON beads(model);
	CREATE INDEX IF NOT EXISTS idx_beads_agent ON beads(agent);
	CREATE INDEX IF NOT EXISTS idx_beads_task_type ON beads(task_type);
	CREATE INDEX IF NOT EXISTS idx_beads_success ON beads(success);
	CREATE INDEX IF NOT EXISTS idx_beads_quality ON beads(quality);
	`
	_, err := s.db.Exec(schema)
	return err
}

// Insert adds a BEADS entry to SQLite.
func (s *Store) Insert(e Entry) error {
	altJSON := "[]"
	if len(e.Alternatives) > 0 {
		data, _ := json.Marshal(e.Alternatives)
		altJSON = string(data)
	}

	successInt := 0
	if e.Success {
		successInt = 1
	}
	qaInt := 0
	if e.QualityAcceptable {
		qaInt = 1
	}
	caInt := 0
	if e.CostAcceptable {
		caInt = 1
	}
	laInt := 0
	if e.LatencyAcceptable {
		laInt = 1
	}

	_, err := s.db.Exec(
		`INSERT INTO beads (
			schema_ver, timestamp, status,
			task_type, phase, domain, complexity, budget, task,
			model, agent, reasoning, alternatives,
			prompt_tokens, completion_tokens, quality, latency_ms, cost, success,
			quality_acceptable, cost_acceptable, latency_acceptable,
			session_id, project, error
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?
		)`,
		e.SchemaVersion, e.Timestamp, e.Status,
		e.TaskType, e.Phase, e.Domain, e.Complexity, e.Budget, e.Task,
		e.Model, e.Agent, e.Reasoning, altJSON,
		e.PromptTokens, e.CompletionTokens, e.Quality, e.LatencyMs, e.Cost, successInt,
		qaInt, caInt, laInt,
		e.SessionID, e.Project, e.Error,
	)
	return err
}

// QueryDB queries BEADS entries from SQLite using SQL.
func (s *Store) QueryDB(query string, args ...interface{}) ([]Entry, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		var altJSON string
		var successInt, qaInt, caInt, laInt int

		if err := rows.Scan(
			&e.SchemaVersion, &e.Timestamp, &e.Status,
			&e.TaskType, &e.Phase, &e.Domain, &e.Complexity, &e.Budget, &e.Task,
			&e.Model, &e.Agent, &e.Reasoning, &altJSON,
			&e.PromptTokens, &e.CompletionTokens, &e.Quality, &e.LatencyMs, &e.Cost, &successInt,
			&qaInt, &caInt, &laInt,
			&e.SessionID, &e.Project, &e.Error,
		); err != nil {
			return nil, err
		}

		e.Success = successInt != 0
		e.QualityAcceptable = qaInt != 0
		e.CostAcceptable = caInt != 0
		e.LatencyAcceptable = laInt != 0

		if altJSON != "" && altJSON != "[]" {
			json.Unmarshal([]byte(altJSON), &e.Alternatives)
		}

		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// Count returns total entries.
func (s *Store) Count() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM beads`).Scan(&n)
	return n, err
}

// Cleanup removes entries older than the given duration.
func (s *Store) Cleanup(age time.Duration) error {
	cutoff := time.Now().Add(-age).UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`DELETE FROM beads WHERE timestamp < ?`, cutoff)
	return err
}

// ──────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}
