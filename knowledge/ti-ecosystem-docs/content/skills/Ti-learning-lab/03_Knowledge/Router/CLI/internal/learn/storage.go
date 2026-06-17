package learn

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Storage wraps SQLite database for BEADS v2 learning data.
type Storage struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

// NewStorage creates or opens the learning database.
func NewStorage(path string) (*Storage, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		path = filepath.Join(home, ".ti", "data", "learn.db")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	// WAL mode for concurrent reads/writes
	dbPath := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Connection pool tuning
	db.SetMaxOpenConns(5) // Allow 5 concurrent connections
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(1 * time.Hour)

	s := &Storage{db: db, path: path}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

// Path returns the database file path.
func (s *Storage) Path() string {
	return s.path
}

// DB returns the underlying *sql.DB (for advanced queries).
func (s *Storage) DB() *sql.DB {
	return s.db
}

// migrate runs schema migrations.
func (s *Storage) migrate() error {
	schema := `
	-- BEADS v2 extended logs (new columns on existing beads table)
	-- NOTE: We extend the existing beads table, not replace it

	-- New table: prompt_patterns (extracted patterns)
	CREATE TABLE IF NOT EXISTS prompt_patterns (
		id               INTEGER PRIMARY KEY AUTOINCREMENT,
		pattern_hash     TEXT NOT NULL UNIQUE,
		pattern_template TEXT NOT NULL,
		task_type        TEXT NOT NULL DEFAULT 'general',
		task_domain      TEXT NOT NULL DEFAULT 'general',
		success_count    INTEGER NOT NULL DEFAULT 0,
		failure_count    INTEGER NOT NULL DEFAULT 0,
		avg_quality      REAL NOT NULL DEFAULT 0.0,
		best_model       TEXT DEFAULT '',
		first_seen       DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used        DATETIME DEFAULT CURRENT_TIMESTAMP,
		confidence       REAL DEFAULT 0.0,  -- Bayesian confidence (0-1)
		sample_count     INTEGER AS (success_count + failure_count) VIRTUAL,
		success_rate     REAL AS (CAST(success_count AS REAL) / (success_count + failure_count)) VIRTUAL
	);

	-- New table: model_performance (aggregated stats)
	CREATE TABLE IF NOT EXISTS model_performance (
		id               INTEGER PRIMARY KEY AUTOINCREMENT,
		provider         TEXT NOT NULL,
		model            TEXT NOT NULL,
		task_type        TEXT NOT NULL DEFAULT 'all',
		task_domain      TEXT NOT NULL DEFAULT 'all',
		sample_count     INTEGER NOT NULL DEFAULT 0,
		sample_count_7d  INTEGER NOT NULL DEFAULT 0,
		sample_count_30d INTEGER NOT NULL DEFAULT 0,
		avg_quality      REAL NOT NULL DEFAULT 0.0,
		avg_quality_7d   REAL NOT NULL DEFAULT 0.0,
		avg_quality_30d  REAL NOT NULL DEFAULT 0.0,
		success_rate     REAL NOT NULL DEFAULT 0.0,
		success_rate_7d  REAL NOT NULL DEFAULT 0.0,
		success_rate_30d REAL NOT NULL DEFAULT 0.0,
		avg_cost         REAL NOT NULL DEFAULT 0.0,
		total_cost       REAL NOT NULL DEFAULT 0.0,
		avg_latency_ms   REAL NOT NULL DEFAULT 0.0,
		quality_per_dollar REAL DEFAULT 0.0,
		quality_per_second REAL DEFAULT 0.0,
		routing_weight   REAL DEFAULT 1.0,
		exploration_bonus REAL DEFAULT 0.0,
		last_updated     DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(provider, model, task_type, task_domain)
	);

	-- New table: learning_config (user preferences)
	CREATE TABLE IF NOT EXISTS learning_config (
		key           TEXT PRIMARY KEY,
		value         TEXT NOT NULL,
		value_type    TEXT NOT NULL DEFAULT 'string',
		description   TEXT,
		default_value TEXT,
		updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- New table: budget_tracker (daily/monthly spending)
	CREATE TABLE IF NOT EXISTS budget_tracker (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		date         DATE NOT NULL DEFAULT CURRENT_DATE,
		daily_used   REAL NOT NULL DEFAULT 0.0,
		monthly_used REAL NOT NULL DEFAULT 0.0,
		UNIQUE(date)
	);

	-- Indexes for fast queries
	CREATE INDEX IF NOT EXISTS idx_prompt_task ON prompt_patterns(task_type, task_domain);
	CREATE INDEX IF NOT EXISTS idx_prompt_quality ON prompt_patterns(avg_quality DESC);
	CREATE INDEX IF NOT EXISTS idx_pattern_hash ON prompt_patterns(pattern_hash);

	CREATE INDEX IF NOT EXISTS idx_model_provider ON model_performance(provider, model);
	CREATE INDEX IF NOT EXISTS idx_model_task ON model_performance(task_type);
	CREATE INDEX IF NOT EXISTS idx_model_weight ON model_performance(routing_weight DESC);

	CREATE INDEX IF NOT EXISTS idx_config_key ON learning_config(key);

	-- Seed default config
	INSERT OR IGNORE INTO learning_config (key, value, value_type, description, default_value) VALUES
		('mode', 'off', 'string', 'Learning mode: off, shadow, hybrid, auto', 'off'),
		('hybrid_ratio', '20', 'int', 'Percentage of requests for learned routing (0-100)', '20'),
		('daily_budget_usd', '1.00', 'float', 'Daily spending limit in USD', '1.00'),
		('monthly_budget_usd', '30.00', 'float', 'Monthly spending limit in USD', '30.00'),
		('success_threshold', '7.0', 'float', 'Quality threshold for success (1-10)', '7.0'),
		('exploration_rate', '0.2', 'float', 'Bandit exploration rate (0-1)', '0.2'),
		('min_samples', '10', 'int', 'Min samples before recommending model', '10'),
		('enable_context_injection', 'true', 'bool', 'Auto-inject context from similar sessions', 'true'),
		('enable_cost_optimization', 'true', 'bool', 'Use cost-aware routing', 'true');

	-- Trigger: Update model_performance quality_per_dollar on change
	CREATE TRIGGER IF NOT EXISTS trg_update_quality_per_dollar
	AFTER UPDATE ON model_performance
	FOR EACH ROW
	BEGIN
		UPDATE model_performance
		SET quality_per_dollar = CASE
			WHEN NEW.avg_cost > 0 THEN NEW.avg_quality / NEW.avg_cost
			ELSE 0
		END,
		quality_per_second = CASE
			WHEN NEW.avg_latency_ms > 0 THEN NEW.avg_quality / (NEW.avg_latency_ms / 1000)
			ELSE 0
		END,
		last_updated = CURRENT_TIMESTAMP
		WHERE id = NEW.id;
	END;
	`

	_, err := s.db.Exec(schema)
	return err
}

// InsertLog adds a BEADS v2 log entry.
func (s *Storage) InsertLog(e *LogInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Convert LogInput to flat row
	keywordsJSON, _ := json.Marshal(e.Keywords)
	contextSourcesJSON, _ := json.Marshal(e.ContextSources)
	fallbackChainJSON, _ := json.Marshal(e.FallbackChain)

	_, err := s.db.Exec(`
		INSERT INTO beads_logs (
			timestamp, session_id, prompt_hash, prompt_length,
			task_type, task_domain, complexity_score, keywords,
			context_sources, context_items_count,
			provider, model, model_family, routing_source,
			tokens_input, tokens_output, latency_ms, cost_usd,
			quality_user, quality_auto, quality_combined,
			compilation_success, test_run_success, test_pass_rate,
			git_committed, user_edited, iterations_needed,
			outcome, outcome_reason,
			prompt_template, matched_pattern_id, recommended_context, fallback_chain
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		e.Timestamp, e.SessionID, e.PromptHash, e.PromptLength,
		e.TaskType, e.TaskDomain, e.ComplexityScore, string(keywordsJSON),
		string(contextSourcesJSON), e.ContextItemsCount,
		e.Provider, e.Model, e.ModelFamily, e.RoutingSource,
		e.TokensInput, e.TokensOutput, e.LatencyMs, e.CostUSD,
		e.QualityUser, e.QualityAuto, e.QualityCombined,
		e.CompilationSuccess, e.TestRunSuccess, e.TestPassRate,
		e.GitCommitted, e.UserEdited, e.IterationsNeeded,
		e.Outcome, e.OutcomeReason,
		e.PromptTemplate, e.MatchedPatternID, e.RecommendedContext, string(fallbackChainJSON),
	)

	return err
}

// GetBySession retrieves logs for a session.
func (s *Storage) GetBySession(sessionID string) ([]*LogInput, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT timestamp, session_id, prompt_hash, prompt_length,
		       task_type, task_domain, complexity_score, keywords,
		       context_sources, context_items_count,
		       provider, model, model_family, routing_source,
		       tokens_input, tokens_output, latency_ms, cost_usd,
		       quality_user, quality_auto, quality_combined,
		       compilation_success, test_run_success, test_pass_rate,
		       git_committed, user_edited, iterations_needed,
		       outcome, outcome_reason,
		       prompt_template, matched_pattern_id, recommended_context, fallback_chain
		FROM beads_logs
		WHERE session_id = ?
		ORDER BY timestamp DESC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*LogInput
	for rows.Next() {
		var l LogInput
		var keywordsJSON, contextSourcesJSON, fallbackChainJSON string

		err := rows.Scan(
			&l.Timestamp, &l.SessionID, &l.PromptHash, &l.PromptLength,
			&l.TaskType, &l.TaskDomain, &l.ComplexityScore, &keywordsJSON,
			&contextSourcesJSON, &l.ContextItemsCount,
			&l.Provider, &l.Model, &l.ModelFamily, &l.RoutingSource,
			&l.TokensInput, &l.TokensOutput, &l.LatencyMs, &l.CostUSD,
			&l.QualityUser, &l.QualityAuto, &l.QualityCombined,
			&l.CompilationSuccess, &l.TestRunSuccess, &l.TestPassRate,
			&l.GitCommitted, &l.UserEdited, &l.IterationsNeeded,
			&l.Outcome, &l.OutcomeReason,
			&l.PromptTemplate, &l.MatchedPatternID, &l.RecommendedContext, &fallbackChainJSON,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(keywordsJSON), &l.Keywords)
		json.Unmarshal([]byte(contextSourcesJSON), &l.ContextSources)
		json.Unmarshal([]byte(fallbackChainJSON), &l.FallbackChain)

		logs = append(logs, &l)
	}
	return logs, rows.Err()
}

// GetStats returns aggregated statistics.
func (s *Storage) GetStats(limit int) (*Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &Stats{}

	// Total samples
	err := s.db.QueryRow(`SELECT COUNT(*) FROM beads_logs`).Scan(&stats.TotalSamples)
	if err != nil {
		return nil, err
	}
	if stats.TotalSamples == 0 {
		return stats, nil
	}

	// Avg quality
	err = s.db.QueryRow(`
		SELECT AVG(COALESCE(quality_user, quality_combined))
		FROM beads_logs
	`).Scan(&stats.AvgQuality)
	if err != nil {
		return nil, err
	}

	// Success rate (quality >= 7 OR outcome == 'success')
	err = s.db.QueryRow(`
		SELECT AVG(CASE
			WHEN quality_combined >= 7 OR outcome = 'success' THEN 1.0
			ELSE 0.0
		END)
		FROM beads_logs
	`).Scan(&stats.SuccessRate)
	if err != nil {
		return nil, err
	}

	// Total cost
	err = s.db.QueryRow(`SELECT SUM(cost_usd) FROM beads_logs`).Scan(&stats.TotalCostUSD)
	if err != nil {
		return nil, err
	}

	// Top models (by quality)
	rows, err := s.db.Query(`
		SELECT model, provider, AVG(quality_combined) as avg_q, COUNT(*) as cnt, SUM(cost_usd) as cost
		FROM beads_logs
		WHERE quality_combined > 0
		GROUP BY provider, model
		ORDER BY avg_q DESC, cnt DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ms ModelStat
		err := rows.Scan(&ms.Model, &ms.Provider, &ms.Quality, &ms.Samples, &ms.Successes)
		if err != nil {
			return nil, err
		}
		stats.TopModels = append(stats.TopModels, ms)
	}

	return stats, nil
}

// PruneOld removes logs older than retention days.
func (s *Storage) PruneOld(retentionDays int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -retentionDays).Format("2006-01-02")
	_, err := s.db.Exec(`DELETE FROM beads_logs WHERE date(timestamp) < ?`, cutoff)
	return err
}

// GetDBSize returns the database file size in bytes.
func (s *Storage) GetDBSize() (int64, error) {
	info, err := os.Stat(s.path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// Close closes the storage.
func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}
