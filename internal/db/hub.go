// Package db provides the central database (Hub) for TiBrain.
// Manages SQLite connections, schema, migrations, and all data operations.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

// Version of the schema. Increment when making changes.
var latestSchemaVersion = 1

// Hub is the central database connection manager.
type Hub struct {
	db          *sql.DB
	mu          sync.RWMutex
}

// NewHub creates and initializes a new Hub with the given data directory.
func NewHub(dataDir string) (*Hub, error) {
	dbPath := filepath.Join(dataDir, "tibrain.db")
	isFreshDB := false

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	if info, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			isFreshDB = true
		} else {
			return nil, fmt.Errorf("stat database: %w", err)
		}
	} else if info.Size() == 0 {
		isFreshDB = true
	}

	dsn := dbPath + "?_pragma=journal_mode(WAL)" +
		"&_pragma=synchronous(NORMAL)" +
		"&_pragma=temp_store(MEMORY)" +
		"&_pragma=foreign_keys(ON)" +
		"&_pragma=busy_timeout(1500)" +
		"&_pragma=cache_size(-20000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := InitDB(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("init database: %w", err)
	}
	if isFreshDB {
		if err := setSchemaVersion(db, latestSchemaVersion); err != nil {
			db.Close()
			return nil, fmt.Errorf("mark fresh schema version: %w", err)
		}
	}

	if err := MigrateDatabase(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	// Note: Integration schema and RAG schema are initialized separately
	// by the main package to avoid circular dependencies.
	// Call db.UpdateIntegrationSchema(db) and db.InitRAGSchema(db) from main.

	h := &Hub{
		db: db,
	}
	return h, nil
}

// DB returns the underlying sql.DB for direct access.
func (h *Hub) DB() *sql.DB {
	return h.db
}

// Close closes the database connection.
func (h *Hub) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.db != nil {
		return h.db.Close()
	}
	return nil
}

// InitDB creates the base schema.
func InitDB(db *sql.DB) error {
	schema := `
		-- CLI Registry
		CREATE TABLE IF NOT EXISTS cli_registry (
			cli_id TEXT PRIMARY KEY,
			cli_name TEXT NOT NULL,
			brain_url TEXT NOT NULL,
			last_heartbeat INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_cli_registry_status ON cli_registry(status);

		-- Global Handoff Tracking
		CREATE TABLE IF NOT EXISTS global_handoffs (
			id TEXT PRIMARY KEY,
			from_cli TEXT NOT NULL,
			to_cli TEXT NOT NULL,
			context TEXT,
			output TEXT,
			timestamp INTEGER NOT NULL,
			metadata TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_global_handoffs_from ON global_handoffs(from_cli);
		CREATE INDEX IF NOT EXISTS idx_global_handoffs_to ON global_handoffs(to_cli);
		CREATE INDEX IF NOT EXISTS idx_global_handoffs_timestamp ON global_handoffs(timestamp);

		-- MCP Registry - Centralized MCP server management
		CREATE TABLE IF NOT EXISTS mcp_registry (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			command TEXT,
			args TEXT,
			env TEXT,
			description TEXT,
			enabled INTEGER NOT NULL DEFAULT 1,
			cli_id TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_mcp_registry_cli ON mcp_registry(cli_id);
		CREATE INDEX IF NOT EXISTS idx_mcp_registry_enabled ON mcp_registry(enabled);

		-- Tool Registry - Tools for agents (Devin, Claude, etc.)
		CREATE TABLE IF NOT EXISTS tool_registry (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			parameters TEXT,
			handler TEXT,
			category TEXT,
			permissions TEXT,
			enabled INTEGER NOT NULL DEFAULT 1,
			quality_score INTEGER NOT NULL DEFAULT 80,
			security_score INTEGER NOT NULL DEFAULT 80,
			best_practices_score INTEGER NOT NULL DEFAULT 80,
			source TEXT NOT NULL DEFAULT 'unknown',
			family TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '',
			version TEXT NOT NULL DEFAULT '',
			skill_level TEXT NOT NULL DEFAULT 'l2',
			quality_tier TEXT NOT NULL DEFAULT 'platinum',
			security_tier TEXT NOT NULL DEFAULT 'hardened',
			security_status TEXT NOT NULL DEFAULT 'passed',
			validation_status TEXT NOT NULL DEFAULT 'passed',
			variant_id TEXT NOT NULL DEFAULT '',
			variant_label TEXT NOT NULL DEFAULT '',
			source_type TEXT NOT NULL DEFAULT 'community',
			root_path TEXT NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_tool_registry_category ON tool_registry(category);
		CREATE INDEX IF NOT EXISTS idx_tool_registry_enabled ON tool_registry(enabled);

		-- Tool Usage Log - Track tool execution
		CREATE TABLE IF NOT EXISTS tool_usage_log (
			id TEXT PRIMARY KEY,
			tool_name TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			parameters TEXT,
			result TEXT,
			success INTEGER NOT NULL DEFAULT 1,
			error_message TEXT,
			timestamp INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_tool_usage_tool ON tool_usage_log(tool_name);
		CREATE INDEX IF NOT EXISTS idx_tool_usage_agent ON tool_usage_log(agent_id);
		CREATE INDEX IF NOT EXISTS idx_tool_usage_timestamp ON tool_usage_log(timestamp);

		-- Memories (for cognitive memory)
		CREATE TABLE IF NOT EXISTS memories (
			id TEXT PRIMARY KEY,
			api_key_id TEXT NOT NULL DEFAULT 'local',
			session_id TEXT,
			type TEXT NOT NULL CHECK(type IN ('factual', 'episodic', 'procedural', 'semantic', 'working', 'long_term')),
			key TEXT,
			content TEXT NOT NULL,
			metadata TEXT,
			importance REAL NOT NULL DEFAULT 0.5,
			access_count INTEGER NOT NULL DEFAULT 0,
			last_access TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now')),
			expires_at TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
		CREATE INDEX IF NOT EXISTS idx_memories_expires ON memories(expires_at);
		CREATE INDEX IF NOT EXISTS idx_memories_importance ON memories(importance);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	return nil
}

// MigrateDatabase handles schema migrations.
func MigrateDatabase(db *sql.DB) error {
	var schemaVersion int
	if err := db.QueryRow("PRAGMA user_version").Scan(&schemaVersion); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if schemaVersion >= latestSchemaVersion {
		return nil
	}
	if err := setSchemaVersion(db, latestSchemaVersion); err != nil {
		return fmt.Errorf("set schema version: %w", err)
	}
	return nil
}

func setSchemaVersion(db *sql.DB, version int) error {
	if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", version)); err != nil {
		return fmt.Errorf("apply schema version %d: %w", version, err)
	}
	return nil
}