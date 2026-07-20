// Database schema updates for Ti Brain integration.
package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// updateIntegrationSchema adds integration-related tables.
func updateIntegrationSchema(db *sql.DB) error {
	schema := `
	-- Agent Registry
	CREATE TABLE IF NOT EXISTS agent_registry (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		endpoint TEXT,
		status TEXT NOT NULL DEFAULT 'inactive',
		last_seen INTEGER,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		capabilities TEXT,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_agent_registry_status ON agent_registry(status);
	CREATE INDEX IF NOT EXISTS idx_agent_registry_type ON agent_registry(type);

	-- Knowledge Bases
	CREATE TABLE IF NOT EXISTS rag_knowledge_bases (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		path TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'inactive',
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		doc_count INTEGER DEFAULT 0,
		vector_count INTEGER DEFAULT 0,
		last_sync INTEGER,
		sync_status TEXT DEFAULT 'pending',
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_knowledge_bases_status ON rag_knowledge_bases(status);

	-- Cross-Brain Communication Log
	CREATE TABLE IF NOT EXISTS cross_brain_communication (
		id TEXT PRIMARY KEY,
		from_brain TEXT NOT NULL,
		to_brain TEXT NOT NULL,
		communication_type TEXT NOT NULL,
		payload TEXT,
		response TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		latency_ms INTEGER,
		timestamp INTEGER NOT NULL,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_cross_brain_timestamp ON cross_brain_communication(timestamp);
	CREATE INDEX IF NOT EXISTS idx_cross_brain_from_to ON cross_brain_communication(from_brain, to_brain);

	-- Orchestration Log
	CREATE TABLE IF NOT EXISTS orchestration_log (
		id TEXT PRIMARY KEY,
		query TEXT NOT NULL,
		sources TEXT NOT NULL,
		results TEXT,
		orchestration_time_ms INTEGER,
		status TEXT NOT NULL DEFAULT 'pending',
		timestamp INTEGER NOT NULL,
		user_id TEXT,
		session_id TEXT,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_orchestration_timestamp ON orchestration_log(timestamp);
	CREATE INDEX IF NOT EXISTS idx_orchestration_status ON orchestration_log(status);

	-- Agent Performance Metrics
	CREATE TABLE IF NOT EXISTS agent_performance (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL,
		metric_type TEXT NOT NULL,
		metric_value REAL NOT NULL,
		timestamp INTEGER NOT NULL,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_agent_performance_agent ON agent_performance(agent_id);
	CREATE INDEX IF NOT EXISTS idx_agent_performance_timestamp ON agent_performance(timestamp);

	-- Task Memory
	CREATE TABLE IF NOT EXISTS task_memory (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL,
		task_type TEXT NOT NULL,
		task_description TEXT NOT NULL,
		task_status TEXT NOT NULL DEFAULT 'pending',
		assigned_agent TEXT,
		priority INTEGER DEFAULT 5,
		context TEXT,
		result TEXT,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		completed_at INTEGER,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_task_memory_task_id ON task_memory(task_id);
	CREATE INDEX IF NOT EXISTS idx_task_memory_status ON task_memory(task_status);
	CREATE INDEX IF NOT EXISTS idx_task_memory_agent ON task_memory(assigned_agent);
	CREATE INDEX IF NOT EXISTS idx_task_memory_created ON task_memory(created_at);

	-- Decision Memory
	CREATE TABLE IF NOT EXISTS decision_memory (
		id TEXT PRIMARY KEY,
		decision_id TEXT NOT NULL,
		decision_type TEXT NOT NULL,
		decision_context TEXT NOT NULL,
		decision_outcome TEXT NOT NULL,
		decision_rationale TEXT,
		alternatives_considered TEXT,
		impact_assessment TEXT,
		confidence REAL DEFAULT 0.5,
		decision_maker TEXT NOT NULL,
		related_task_id TEXT,
		timestamp INTEGER NOT NULL,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_decision_memory_decision_id ON decision_memory(decision_id);
	CREATE INDEX IF NOT EXISTS idx_decision_memory_type ON decision_memory(decision_type);
	CREATE INDEX IF NOT EXISTS idx_decision_memory_maker ON decision_memory(decision_maker);
	CREATE INDEX IF NOT EXISTS idx_decision_memory_timestamp ON decision_memory(timestamp);

	-- User Preferences
	CREATE TABLE IF NOT EXISTS user_preferences (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		preference_key TEXT NOT NULL,
		preference_value TEXT NOT NULL,
		preference_type TEXT NOT NULL,
		category TEXT,
		valid_from INTEGER NOT NULL,
		valid_until INTEGER,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_user_preferences_user ON user_preferences(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_preferences_key ON user_preferences(preference_key);
	CREATE INDEX IF NOT EXISTS idx_user_preferences_category ON user_preferences(category);
	CREATE INDEX IF NOT EXISTS idx_user_preferences_valid ON user_preferences(valid_from, valid_until);

	-- Agent Performance (Enhanced)
	CREATE TABLE IF NOT EXISTS agent_performance_enhanced (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL,
		session_id TEXT,
		task_id TEXT,
		performance_metric TEXT NOT NULL,
		metric_value REAL NOT NULL,
		metric_unit TEXT,
		timestamp INTEGER NOT NULL,
		context TEXT,
		benchmark_value REAL,
		performance_rating TEXT,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_agent_perf_enhanced_agent ON agent_performance_enhanced(agent_id);
	CREATE INDEX IF NOT EXISTS idx_agent_perf_enhanced_task ON agent_performance_enhanced(task_id);
	CREATE INDEX IF NOT EXISTS idx_agent_perf_enhanced_timestamp ON agent_performance_enhanced(timestamp);
	CREATE INDEX IF NOT EXISTS idx_agent_perf_enhanced_metric ON agent_performance_enhanced(performance_metric);

	-- Routing Decision Log
	CREATE TABLE IF NOT EXISTS routing_decision_log (
		id TEXT PRIMARY KEY,
		query TEXT NOT NULL,
		route TEXT NOT NULL,
		confidence REAL NOT NULL,
		reasoning TEXT,
		rule_matched TEXT,
		success BOOLEAN NOT NULL DEFAULT 0,
		duration_ms INTEGER,
		result_count INTEGER DEFAULT 0,
		timestamp INTEGER NOT NULL,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_routing_decision_route ON routing_decision_log(route);
	CREATE INDEX IF NOT EXISTS idx_routing_decision_timestamp ON routing_decision_log(timestamp);
	CREATE INDEX IF NOT EXISTS idx_routing_decision_success ON routing_decision_log(success);

	-- RTK Rules for offline agent compression mapping
	CREATE TABLE IF NOT EXISTS rtk_rules (
		id TEXT PRIMARY KEY,
		rule_type TEXT NOT NULL,
		model_pattern TEXT NOT NULL,
		config TEXT,
		priority INTEGER DEFAULT 0,
		enabled INTEGER DEFAULT 1,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_rtk_rules_type ON rtk_rules(rule_type);
	CREATE INDEX IF NOT EXISTS idx_rtk_rules_enabled ON rtk_rules(enabled);

	-- RTK Savings Log for asynchronous token killer savings stats
	CREATE TABLE IF NOT EXISTS rtk_savings_log (
		id TEXT PRIMARY KEY,
		command TEXT,
		input_tokens INTEGER DEFAULT 0,
		output_tokens INTEGER DEFAULT 0,
		saved_tokens INTEGER DEFAULT 0,
		savings_pct REAL DEFAULT 0.0,
		duration_ms INTEGER DEFAULT 0,
		timestamp INTEGER NOT NULL,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_rtk_savings_timestamp ON rtk_savings_log(timestamp);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("create integration schema: %w", err)
	}

	// Seed default RTK rules if empty
	if err := seedDefaultRTKRules(db); err != nil {
		log.Printf("Warning: Failed to seed default RTK rules: %v", err)
	}

	log.Printf("Integration schema updated successfully")
	return nil
}

// seedDefaultRTKRules populates the rtk_rules table with dynamic defaults.
func seedDefaultRTKRules(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM rtk_rules").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already seeded or customized
	}

	rules := []struct {
		ID           string
		RuleType     string
		ModelPattern string
		Config       string
		Priority     int
	}{
		{
			ID:           "rule_gemini_ultra",
			RuleType:     "compression",
			ModelPattern: "gemini*",
			Config:       `{"max_input_chars":200000,"preserve_last_n":20,"preserve_system":true,"compress_code_blocks":false}`,
			Priority:     100,
		},
		{
			ID:           "rule_claude_gpt_large",
			RuleType:     "compression",
			ModelPattern: "claude*",
			Config:       `{"max_input_chars":80000,"preserve_last_n":10,"preserve_system":true,"compress_code_blocks":false}`,
			Priority:     90,
		},
		{
			ID:           "rule_gpt4",
			RuleType:     "compression",
			ModelPattern: "gpt-4*",
			Config:       `{"max_input_chars":80000,"preserve_last_n":10,"preserve_system":true,"compress_code_blocks":false}`,
			Priority:     80,
		},
		{
			ID:           "rule_default_fallback",
			RuleType:     "compression",
			ModelPattern: "*",
			Config:       `{"max_input_chars":4000,"preserve_last_n":2,"preserve_system":true,"compress_code_blocks":true}`,
			Priority:     0,
		},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Unix()
	stmt, err := tx.Prepare("INSERT INTO rtk_rules (id, rule_type, model_pattern, config, priority, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, rule := range rules {
		_, err = stmt.Exec(rule.ID, rule.RuleType, rule.ModelPattern, rule.Config, rule.Priority, now, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
