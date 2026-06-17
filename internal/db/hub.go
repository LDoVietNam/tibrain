// Package db provides the central database (Hub) for TiBrain.
// Manages SQLite connections, schema, migrations, and all data operations.
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Hub is the central database connection manager.
type Hub struct {
	db *sql.DB
}

const latestSchemaVersion = 1

type schemaColumn struct {
	Name       string
	Definition string
}

var toolRegistryExtendedColumns = []schemaColumn{
	{Name: "quality_score", Definition: "quality_score INTEGER NOT NULL DEFAULT 80"},
	{Name: "security_score", Definition: "security_score INTEGER NOT NULL DEFAULT 80"},
	{Name: "best_practices_score", Definition: "best_practices_score INTEGER NOT NULL DEFAULT 80"},
	{Name: "source", Definition: "source TEXT NOT NULL DEFAULT 'unknown'"},
	{Name: "family", Definition: "family TEXT NOT NULL DEFAULT ''"},
	{Name: "tags", Definition: "tags TEXT NOT NULL DEFAULT ''"},
	{Name: "version", Definition: "version TEXT NOT NULL DEFAULT ''"},
	{Name: "skill_level", Definition: "skill_level TEXT NOT NULL DEFAULT 'l2'"},
	{Name: "quality_tier", Definition: "quality_tier TEXT NOT NULL DEFAULT 'platinum'"},
	{Name: "security_tier", Definition: "security_tier TEXT NOT NULL DEFAULT 'hardened'"},
	{Name: "security_status", Definition: "security_status TEXT NOT NULL DEFAULT 'passed'"},
	{Name: "validation_status", Definition: "validation_status TEXT NOT NULL DEFAULT 'passed'"},
	{Name: "variant_id", Definition: "variant_id TEXT NOT NULL DEFAULT ''"},
	{Name: "variant_label", Definition: "variant_label TEXT NOT NULL DEFAULT ''"},
	{Name: "source_type", Definition: "source_type TEXT NOT NULL DEFAULT 'community'"},
	{Name: "root_path", Definition: "root_path TEXT NOT NULL DEFAULT ''"},
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

	return &Hub{db: db}, nil
}

// DB returns the underlying sql.DB for direct access.
func (h *Hub) DB() *sql.DB {
	return h.db
}

// Close closes the database connection.
func (h *Hub) Close() error {
	if h.db != nil {
		return h.db.Close()
	}
	return nil
}

// ─────────────────────────────────────────────────────────────
// Schema Initialization
// ─────────────────────────────────────────────────────────────

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

	-- Skill Registry (Optional)
	CREATE TABLE IF NOT EXISTS skill_registry (
		id TEXT PRIMARY KEY,
		cli_id TEXT NOT NULL,
		name TEXT NOT NULL,
		content TEXT,
		category TEXT,
		synced_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_skill_registry_cli ON skill_registry(cli_id);
	CREATE INDEX IF NOT EXISTS idx_skill_registry_category ON skill_registry(category);

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

	-- MCP Hub Registry - Integration with external MCP Hub
	CREATE TABLE IF NOT EXISTS mcp_hub_registry (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		endpoint TEXT NOT NULL,
		transport TEXT NOT NULL DEFAULT 'stdio',
		tools TEXT,
		resources TEXT,
		description TEXT,
		enabled INTEGER NOT NULL DEFAULT 1,
		synced_at INTEGER,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_mcp_hub_registry_enabled ON mcp_hub_registry(enabled);
	CREATE INDEX IF NOT EXISTS idx_mcp_hub_registry_synced ON mcp_hub_registry(synced_at);

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

	-- Model Performance Stats (BEADS LEARN integration)
	CREATE TABLE IF NOT EXISTS model_performance_stats (
		id TEXT PRIMARY KEY,
		model TEXT NOT NULL,
		task_type TEXT NOT NULL,
		phase TEXT,
		total_tasks INTEGER NOT NULL DEFAULT 0,
		success_count INTEGER NOT NULL DEFAULT 0,
		fail_count INTEGER NOT NULL DEFAULT 0,
		total_quality REAL NOT NULL DEFAULT 0,
		total_cost REAL NOT NULL DEFAULT 0,
		total_latency INTEGER NOT NULL DEFAULT 0,
		last_used TEXT,
		quality_history TEXT,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_model_stats_model ON model_performance_stats(model);
	CREATE INDEX IF NOT EXISTS idx_model_stats_task_type ON model_performance_stats(task_type);
	CREATE INDEX IF NOT EXISTS idx_model_stats_updated ON model_performance_stats(updated_at);

	-- Cognitive Memory (mirror of migrations/015 + 034)
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
	if err := migrateToolRegistrySchema(db); err != nil {
		return err
	}
	if err := setSchemaVersion(db, latestSchemaVersion); err != nil {
		return fmt.Errorf("set schema version: %w", err)
	}
	return nil
}

func migrateToolRegistrySchema(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info('tool_registry')")
	if err != nil {
		return fmt.Errorf("inspect tool_registry schema: %w", err)
	}
	defer rows.Close()

	existingColumns := make(map[string]struct{}, len(toolRegistryExtendedColumns))
	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal sql.NullString
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &primaryKey); err != nil {
			return fmt.Errorf("scan tool_registry schema: %w", err)
		}
		existingColumns[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate tool_registry schema: %w", err)
	}

	for _, column := range toolRegistryExtendedColumns {
		if _, exists := existingColumns[column.Name]; exists {
			continue
		}
		if _, err := db.Exec("ALTER TABLE tool_registry ADD COLUMN " + column.Definition); err != nil {
			return fmt.Errorf("add %s column: %w", column.Name, err)
		}
	}
	return nil
}

func setSchemaVersion(db *sql.DB, version int) error {
	if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", version)); err != nil {
		return fmt.Errorf("apply schema version %d: %w", version, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────
// CLI Registry Operations
// ─────────────────────────────────────────────────────────────

// CLIInfo represents a registered CLI.
type CLIInfo struct {
	CLIID         string `json:"cli_id"`
	CLIName       string `json:"cli_name"`
	BrainURL      string `json:"brain_url"`
	LastHeartbeat int64  `json:"last_heartbeat"`
	Status        string `json:"status"`
	CreatedAt     int64  `json:"created_at"`
}

// RegisterCLI registers a new CLI.
func (h *Hub) RegisterCLI(cliID, cliName, brainURL string) error {
	timestamp := time.Now().Unix()
	_, err := h.db.Exec(`
		INSERT OR REPLACE INTO cli_registry (cli_id, cli_name, brain_url, last_heartbeat, status, created_at)
		VALUES (?, ?, ?, ?, 'active', ?)
	`, cliID, cliName, brainURL, timestamp, timestamp)
	if err != nil {
		return fmt.Errorf("register cli: %w", err)
	}
	return nil
}

// UnregisterCLI marks a CLI as inactive.
func (h *Hub) UnregisterCLI(cliID string) error {
	_, err := h.db.Exec(`UPDATE cli_registry SET status = 'inactive' WHERE cli_id = ?`, cliID)
	return fmt.Errorf("unregister cli: %w", err)
}

// UpdateHeartbeat updates the heartbeat timestamp for a CLI.
func (h *Hub) UpdateHeartbeat(cliID string) error {
	timestamp := time.Now().Unix()
	_, err := h.db.Exec(`UPDATE cli_registry SET last_heartbeat = ? WHERE cli_id = ?`, timestamp, cliID)
	return fmt.Errorf("update heartbeat: %w", err)
}

// ListCLIs returns all active CLIs.
func (h *Hub) ListCLIs() ([]CLIInfo, error) {
	rows, err := h.db.Query(`
		SELECT cli_id, cli_name, brain_url, last_heartbeat, status, created_at
		FROM cli_registry WHERE status = 'active' ORDER BY created_at
	`)
	if err != nil {
		return nil, fmt.Errorf("list clis: %w", err)
	}
	defer rows.Close()

	var clis []CLIInfo
	for rows.Next() {
		var c CLIInfo
		if err := rows.Scan(&c.CLIID, &c.CLIName, &c.BrainURL, &c.LastHeartbeat, &c.Status, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan cli: %w", err)
		}
		clis = append(clis, c)
	}
	return clis, nil
}

// ─────────────────────────────────────────────────────────────
// Global Handoff Operations
// ─────────────────────────────────────────────────────────────

// GlobalHandoff represents a handoff between CLIs.
type GlobalHandoff struct {
	ID        string            `json:"id"`
	FromCLI   string            `json:"from_cli"`
	ToCLI     string            `json:"to_cli"`
	Context   string            `json:"context"`
	Output    string            `json:"output"`
	Timestamp int64             `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// CreateGlobalHandoff creates a new handoff record.
func (h *Hub) CreateGlobalHandoff(fromCLI, toCLI, context, output string, metadata map[string]string) (string, error) {
	timestamp := time.Now().Unix()
	id := fmt.Sprintf("global-handoff-%s-%s-%d", fromCLI, toCLI, timestamp)

	var metadataStr string
	if metadata != nil {
		if data, err := json.Marshal(metadata); err == nil {
			metadataStr = string(data)
		}
	}

	_, err := h.db.Exec(`
		INSERT INTO global_handoffs (id, from_cli, to_cli, context, output, timestamp, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, fromCLI, toCLI, context, output, timestamp, metadataStr)
	if err != nil {
		return "", fmt.Errorf("insert global handoff: %w", err)
	}
	return id, nil
}

// RecallGlobalHandoff returns the most recent handoff between two CLIs.
func (h *Hub) RecallGlobalHandoff(fromCLI, toCLI string) (*GlobalHandoff, error) {
	query := `SELECT id, from_cli, to_cli, context, output, timestamp, metadata
		FROM global_handoffs WHERE from_cli = ? AND to_cli = ? ORDER BY timestamp DESC LIMIT 1`

	row := h.db.QueryRow(query, fromCLI, toCLI)
	var g GlobalHandoff
	var metadataStr string
	if err := row.Scan(&g.ID, &g.FromCLI, &g.ToCLI, &g.Context, &g.Output, &g.Timestamp, &metadataStr); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("global handoff not found")
		}
		return nil, fmt.Errorf("query global handoff: %w", err)
	}
	if metadataStr != "" {
		json.Unmarshal([]byte(metadataStr), &g.Metadata)
	}
	return &g, nil
}

// ListHandoffs returns handoffs for a given CLI.
func (h *Hub) ListHandoffs(cliID string) ([]GlobalHandoff, error) {
	query := `SELECT id, from_cli, to_cli, context, output, timestamp, metadata
		FROM global_handoffs WHERE from_cli = ? OR to_cli = ? ORDER BY timestamp DESC LIMIT 50`

	rows, err := h.db.Query(query, cliID, cliID)
	if err != nil {
		return nil, fmt.Errorf("list handoffs: %w", err)
	}
	defer rows.Close()

	var handoffs []GlobalHandoff
	for rows.Next() {
		var g GlobalHandoff
		var metadataStr string
		if err := rows.Scan(&g.ID, &g.FromCLI, &g.ToCLI, &g.Context, &g.Output, &g.Timestamp, &metadataStr); err != nil {
			return nil, fmt.Errorf("scan handoff: %w", err)
		}
		if metadataStr != "" {
			json.Unmarshal([]byte(metadataStr), &g.Metadata)
		}
		handoffs = append(handoffs, g)
	}
	return handoffs, nil
}

// ─────────────────────────────────────────────────────────────
// MCP Registry Operations
// ─────────────────────────────────────────────────────────────

// MCPServer represents a registered MCP server.
type MCPServer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Command     string `json:"command"`
	Args        string `json:"args"`
	Env         string `json:"env"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	CLIID       string `json:"cli_id"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// RegisterMCP registers a new MCP server.
func (h *Hub) RegisterMCP(id, name, mcpType, command, args, env, description string, enabled bool, cliID string) error {
	timestamp := time.Now().Unix()
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	_, err := h.db.Exec(`
		INSERT OR REPLACE INTO mcp_registry (id, name, type, command, args, env, description, enabled, cli_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, name, mcpType, command, args, env, description, enabledInt, cliID, timestamp, timestamp)
	return fmt.Errorf("register mcp: %w", err)
}

// ListMCPs returns enabled MCP servers.
func (h *Hub) ListMCPs(cliID string) ([]MCPServer, error) {
	query := `SELECT id, name, type, command, args, env, description, enabled, cli_id, created_at, updated_at
		FROM mcp_registry WHERE enabled = 1`
	args := []interface{}{}
	if cliID != "" {
		query += " AND cli_id = ?"
		args = append(args, cliID)
	}
	query += " ORDER BY created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list mcps: %w", err)
	}
	defer rows.Close()

	var mcps []MCPServer
	for rows.Next() {
		var m MCPServer
		var enabledInt int
		if err := rows.Scan(&m.ID, &m.Name, &m.Type, &m.Command, &m.Args, &m.Env, &m.Description, &enabledInt, &m.CLIID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan mcp: %w", err)
		}
		m.Enabled = enabledInt == 1
		mcps = append(mcps, m)
	}
	return mcps, nil
}

// GetMCP returns a specific MCP server.
func (h *Hub) GetMCP(id string) (*MCPServer, error) {
	query := `SELECT id, name, type, command, args, env, description, enabled, cli_id, created_at, updated_at
		FROM mcp_registry WHERE id = ?`

	row := h.db.QueryRow(query, id)
	var m MCPServer
	var enabledInt int
	if err := row.Scan(&m.ID, &m.Name, &m.Type, &m.Command, &m.Args, &m.Env, &m.Description, &enabledInt, &m.CLIID, &m.CreatedAt, &m.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mcp not found")
		}
		return nil, fmt.Errorf("query mcp: %w", err)
	}
	m.Enabled = enabledInt == 1
	return &m, nil
}

// ─────────────────────────────────────────────────────────────
// MCP Hub Registry Operations
// ─────────────────────────────────────────────────────────────

// MCPHubServer represents an MCP Hub server.
type MCPHubServer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Endpoint    string `json:"endpoint"`
	Transport   string `json:"transport"`
	Tools       string `json:"tools"`
	Resources   string `json:"resources"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	SyncedAt    int64  `json:"synced_at"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// RegisterMCPHubServer registers a new MCP Hub server.
func (h *Hub) RegisterMCPHubServer(id, name, endpoint, transport, tools, resources, description string, enabled bool) error {
	timestamp := time.Now().Unix()
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	_, err := h.db.Exec(`
		INSERT OR REPLACE INTO mcp_hub_registry (id, name, endpoint, transport, tools, resources, description, enabled, synced_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, name, endpoint, transport, tools, resources, description, enabledInt, timestamp, timestamp, timestamp)
	return fmt.Errorf("register mcp hub server: %w", err)
}

// ListMCPHubServers returns enabled MCP Hub servers.
func (h *Hub) ListMCPHubServers() ([]MCPHubServer, error) {
	query := `SELECT id, name, endpoint, transport, tools, resources, description, enabled, synced_at, created_at, updated_at
		FROM mcp_hub_registry WHERE enabled = 1 ORDER BY created_at DESC`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list mcp hub servers: %w", err)
	}
	defer rows.Close()

	var servers []MCPHubServer
	for rows.Next() {
		var s MCPHubServer
		var enabledInt int
		if err := rows.Scan(&s.ID, &s.Name, &s.Endpoint, &s.Transport, &s.Tools, &s.Resources, &s.Description, &enabledInt, &s.SyncedAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan mcp hub server: %w", err)
		}
		s.Enabled = enabledInt == 1
		servers = append(servers, s)
	}
	return servers, nil
}

// GetMCPHubServer returns a specific MCP Hub server.
func (h *Hub) GetMCPHubServer(id string) (*MCPHubServer, error) {
	query := `SELECT id, name, endpoint, transport, tools, resources, description, enabled, synced_at, created_at, updated_at
		FROM mcp_hub_registry WHERE id = ?`

	row := h.db.QueryRow(query, id)
	var s MCPHubServer
	var enabledInt int
	if err := row.Scan(&s.ID, &s.Name, &s.Endpoint, &s.Transport, &s.Tools, &s.Resources, &s.Description, &enabledInt, &s.SyncedAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mcp hub server not found")
		}
		return nil, fmt.Errorf("query mcp hub server: %w", err)
	}
	s.Enabled = enabledInt == 1
	return &s, nil
}

// UpdateMCPHubServerSyncTime updates the sync timestamp.
func (h *Hub) UpdateMCPHubServerSyncTime(id string) error {
	timestamp := time.Now().Unix()
	_, err := h.db.Exec(`UPDATE mcp_hub_registry SET synced_at = ?, updated_at = ? WHERE id = ?`, timestamp, timestamp, id)
	return fmt.Errorf("update sync time: %w", err)
}

// SyncMCPTools syncs tools from MCP servers (currently disabled).
func (h *Hub) SyncMCPTools() error {
	mcps, err := h.ListMCPs("")
	if err != nil {
		return fmt.Errorf("list mcps: %w", err)
	}
	for _, mcp := range mcps {
		if !mcp.Enabled {
			continue
		}
		// TODO: Integrate with MCP Hub instead of individual MCP clients
		continue
	}
	return nil
}

// CallMCPTool calls a tool on an MCP server (currently disabled).
func (h *Hub) CallMCPTool(toolID string, params map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("MCP tool calling disabled")
}

// ─────────────────────────────────────────────────────────────
// Tool Registry Operations
// ─────────────────────────────────────────────────────────────

// Tool represents a registered tool.
type Tool struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	Parameters         string `json:"parameters"`
	Handler            string `json:"handler"`
	Category           string `json:"category"`
	Permissions        string `json:"permissions"`
	Enabled            bool   `json:"enabled"`
	QualityScore       int    `json:"quality_score"`
	SecurityScore      int    `json:"security_score"`
	BestPracticesScore int    `json:"best_practices_score"`
	Source             string `json:"source"`
	Family             string `json:"family"`
	Tags               string `json:"tags"`
	Version            string `json:"version"`
	SkillLevel         string `json:"skill_level"`
	QualityTier        string `json:"quality_tier"`
	SecurityTier       string `json:"security_tier"`
	SecurityStatus     string `json:"security_status"`
	ValidationStatus   string `json:"validation_status"`
	VariantID          string `json:"variant_id"`
	VariantLabel       string `json:"variant_label"`
	SourceType         string `json:"source_type"`
	RootPath           string `json:"root_path"`
	CreatedAt          int64  `json:"created_at"`
	UpdatedAt          int64  `json:"updated_at"`
}

// RegisterTool registers a new tool.
func (h *Hub) RegisterTool(tool Tool) error {
	timestamp := time.Now().Unix()
	enabledInt := 0
	if tool.Enabled {
		enabledInt = 1
	}
	_, err := h.db.Exec(`
		INSERT OR REPLACE INTO tool_registry (id, name, description, parameters, handler, category, permissions, enabled, quality_score, security_score, best_practices_score, source, family, tags, version, skill_level, quality_tier, security_tier, security_status, validation_status, variant_id, variant_label, source_type, root_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tool.ID, tool.Name, tool.Description, tool.Parameters, tool.Handler, tool.Category, tool.Permissions, enabledInt, tool.QualityScore, tool.SecurityScore, tool.BestPracticesScore, tool.Source, tool.Family, tool.Tags, tool.Version, tool.SkillLevel, tool.QualityTier, tool.SecurityTier, tool.SecurityStatus, tool.ValidationStatus, tool.VariantID, tool.VariantLabel, tool.SourceType, tool.RootPath, timestamp, timestamp)
	return fmt.Errorf("register tool: %w", err)
}

// ListTools returns enabled tools, optionally filtered by category.
func (h *Hub) ListTools(category string) ([]Tool, error) {
	query := `SELECT id, name, description, parameters, handler, category, permissions, enabled, quality_score, security_score, best_practices_score, source, family, tags, version, skill_level, quality_tier, security_tier, security_status, validation_status, variant_id, variant_label, source_type, root_path, created_at, updated_at
		FROM tool_registry WHERE enabled = 1`
	args := []interface{}{}
	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	query += " ORDER BY created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		var t Tool
		var enabledInt, bestPracticesScore int
		var source, family, tags, version, skillLevel, qualityTier, securityTier, securityStatus, validationStatus, variantID, variantLabel, sourceType, rootPath sql.NullString
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Parameters, &t.Handler, &t.Category, &t.Permissions, &enabledInt, &t.QualityScore, &t.SecurityScore, &bestPracticesScore, &source, &family, &tags, &version, &skillLevel, &qualityTier, &securityTier, &securityStatus, &validationStatus, &variantID, &variantLabel, &sourceType, &rootPath, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan tool: %w", err)
		}
		t.Enabled = enabledInt == 1
		t.BestPracticesScore = bestPracticesScore
		t.Source = source.String
		t.Family = family.String
		t.Tags = tags.String
		t.Version = version.String
		t.SkillLevel = skillLevel.String
		t.QualityTier = qualityTier.String
		t.SecurityTier = securityTier.String
		t.SecurityStatus = securityStatus.String
		t.ValidationStatus = validationStatus.String
		t.VariantID = variantID.String
		t.VariantLabel = variantLabel.String
		t.SourceType = sourceType.String
		t.RootPath = rootPath.String
		tools = append(tools, t)
	}
	return tools, nil
}

// GetTool returns a specific tool.
func (h *Hub) GetTool(id string) (*Tool, error) {
	query := `SELECT id, name, description, parameters, handler, category, permissions, enabled, quality_score, security_score, best_practices_score, source, family, tags, version, skill_level, quality_tier, security_tier, security_status, validation_status, variant_id, variant_label, source_type, root_path, created_at, updated_at
		FROM tool_registry WHERE id = ?`

	row := h.db.QueryRow(query, id)
	var t Tool
	var enabledInt, bestPracticesScore int
	var source, family, tags, version, skillLevel, qualityTier, securityTier, securityStatus, validationStatus, variantID, variantLabel, sourceType, rootPath sql.NullString
	if err := row.Scan(&t.ID, &t.Name, &t.Description, &t.Parameters, &t.Handler, &t.Category, &t.Permissions, &enabledInt, &t.QualityScore, &t.SecurityScore, &bestPracticesScore, &source, &family, &tags, &version, &skillLevel, &qualityTier, &securityTier, &securityStatus, &validationStatus, &variantID, &variantLabel, &sourceType, &rootPath, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tool not found")
		}
		return nil, fmt.Errorf("query tool: %w", err)
	}
	t.Enabled = enabledInt == 1
	t.BestPracticesScore = bestPracticesScore
	t.Source = source.String
	t.Family = family.String
	t.Tags = tags.String
	t.Version = version.String
	t.SkillLevel = skillLevel.String
	t.QualityTier = qualityTier.String
	t.SecurityTier = securityTier.String
	t.SecurityStatus = securityStatus.String
	t.ValidationStatus = validationStatus.String
	t.VariantID = variantID.String
	t.VariantLabel = variantLabel.String
	t.SourceType = sourceType.String
	t.RootPath = rootPath.String
	return &t, nil
}

// ─────────────────────────────────────────────────────────────
// Tool Usage Log Operations
// ─────────────────────────────────────────────────────────────

// ToolUsageLog represents a tool execution log entry.
type ToolUsageLog struct {
	ID           string `json:"id"`
	ToolName     string `json:"tool_name"`
	AgentID      string `json:"agent_id"`
	Parameters   string `json:"parameters"`
	Result       string `json:"result"`
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message,omitempty"`
	Timestamp    int64  `json:"timestamp"`
}

// LogToolUsage logs a tool execution.
func (h *Hub) LogToolUsage(toolName, agentID, parameters, result string, success bool, errorMessage string) error {
	timestamp := time.Now().Unix()
	id := fmt.Sprintf("tool-log-%s-%s-%d", toolName, agentID, timestamp)
	successInt := 0
	if success {
		successInt = 1
	}
	_, err := h.db.Exec(`
		INSERT INTO tool_usage_log (id, tool_name, agent_id, parameters, result, success, error_message, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, id, toolName, agentID, parameters, result, successInt, errorMessage, timestamp)
	return fmt.Errorf("log tool usage: %w", err)
}

// GetToolUsageLogs returns tool usage logs.
func (h *Hub) GetToolUsageLogs(toolName, agentID string, limit int) ([]ToolUsageLog, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, tool_name, agent_id, parameters, result, success, error_message, timestamp FROM tool_usage_log`
	args := []interface{}{}
	if toolName != "" {
		query += " WHERE tool_name = ?"
		args = append(args, toolName)
	}
	if agentID != "" {
		if toolName != "" {
			query += " AND agent_id = ?"
		} else {
			query += " WHERE agent_id = ?"
		}
		args = append(args, agentID)
	}
	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tool usage logs: %w", err)
	}
	defer rows.Close()

	var logs []ToolUsageLog
	for rows.Next() {
		var l ToolUsageLog
		var successInt int
		if err := rows.Scan(&l.ID, &l.ToolName, &l.AgentID, &l.Parameters, &l.Result, &successInt, &l.ErrorMessage, &l.Timestamp); err != nil {
			return nil, fmt.Errorf("scan tool usage log: %w", err)
		}
		l.Success = successInt == 1
		logs = append(logs, l)
	}
	return logs, nil
}

// ─────────────────────────────────────────────────────────────
// Model Performance Stats Operations
// ─────────────────────────────────────────────────────────────

// ModelPerformanceStats represents model performance data.
type ModelPerformanceStats struct {
	ID             string    `json:"id"`
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
	QualityHistory []float64 `json:"quality_history"`
	CreatedAt      int64     `json:"created_at"`
	UpdatedAt      int64     `json:"updated_at"`
}

// UpdateModelStats upserts model performance stats.
func (h *Hub) UpdateModelStats(stats ModelPerformanceStats) error {
	now := time.Now().Unix()
	if stats.ID == "" {
		stats.ID = fmt.Sprintf("%s:%s", stats.Model, stats.TaskType)
	}
	if stats.CreatedAt == 0 {
		stats.CreatedAt = now
	}
	stats.UpdatedAt = now

	qualityHistoryJSON, _ := json.Marshal(stats.QualityHistory)

	query := `
		INSERT INTO model_performance_stats 
		(id, model, task_type, phase, total_tasks, success_count, fail_count, 
		 total_quality, total_cost, total_latency, last_used, quality_history, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			total_tasks = total_tasks + ?,
			success_count = success_count + ?,
			fail_count = fail_count + ?,
			total_quality = total_quality + ?,
			total_cost = total_cost + ?,
			total_latency = total_latency + ?,
			last_used = ?,
			quality_history = ?,
			updated_at = ?
	`

	_, err := h.db.Exec(query,
		stats.ID, stats.Model, stats.TaskType, stats.Phase,
		stats.TotalTasks, stats.SuccessCount, stats.FailCount,
		stats.TotalQuality, stats.TotalCost, stats.TotalLatency,
		stats.LastUsed, string(qualityHistoryJSON), stats.CreatedAt, stats.UpdatedAt,
		stats.TotalTasks, stats.SuccessCount, stats.FailCount,
		stats.TotalQuality, stats.TotalCost, stats.TotalLatency,
		stats.LastUsed, string(qualityHistoryJSON), stats.UpdatedAt,
	)
	return fmt.Errorf("update model stats: %w", err)
}

// GetModelStats retrieves model performance stats.
func (h *Hub) GetModelStats(model, taskType string) (*ModelPerformanceStats, error) {
	id := fmt.Sprintf("%s:%s", model, taskType)
	query := `SELECT id, model, task_type, phase, total_tasks, success_count, fail_count,
		total_quality, total_cost, total_latency, last_used, quality_history, created_at, updated_at
		FROM model_performance_stats WHERE id = ?`

	var stats ModelPerformanceStats
	var qualityHistoryJSON string
	err := h.db.QueryRow(query, id).Scan(
		&stats.ID, &stats.Model, &stats.TaskType, &stats.Phase,
		&stats.TotalTasks, &stats.SuccessCount, &stats.FailCount,
		&stats.TotalQuality, &stats.TotalCost, &stats.TotalLatency,
		&stats.LastUsed, &qualityHistoryJSON, &stats.CreatedAt, &stats.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("stats not found")
		}
		return nil, fmt.Errorf("query model stats: %w", err)
	}
	json.Unmarshal([]byte(qualityHistoryJSON), &stats.QualityHistory)
	return &stats, nil
}

// GetBestModel returns the best performing model for a task type.
func (h *Hub) GetBestModel(taskType string) (model string, avgQuality float64, totalTasks int, err error) {
	query := `
		SELECT model, task_type, total_tasks, success_count, total_quality
		FROM model_performance_stats
		WHERE task_type = ? AND total_tasks >= 3
		ORDER BY (total_quality / total_tasks) DESC
		LIMIT 1
	`
	var foundTaskType string
	err = h.db.QueryRow(query, taskType).Scan(&model, &foundTaskType, &totalTasks, new(int), &avgQuality)
	if err == sql.ErrNoRows {
		return "", 0, 0, nil
	}
	if err != nil {
		return "", 0, 0, fmt.Errorf("query best model: %w", err)
	}
	avgQuality = avgQuality / float64(totalTasks)
	return
}

// GetAllModelStats returns all model performance stats.
func (h *Hub) GetAllModelStats() ([]ModelPerformanceStats, error) {
	query := `SELECT id, model, task_type, phase, total_tasks, success_count, fail_count,
		total_quality, total_cost, total_latency, last_used, quality_history, created_at, updated_at
		FROM model_performance_stats ORDER BY updated_at DESC`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query all model stats: %w", err)
	}
	defer rows.Close()

	var allStats []ModelPerformanceStats
	for rows.Next() {
		var stats ModelPerformanceStats
		var qualityHistoryJSON string
		if err := rows.Scan(&stats.ID, &stats.Model, &stats.TaskType, &stats.Phase,
			&stats.TotalTasks, &stats.SuccessCount, &stats.FailCount,
			&stats.TotalQuality, &stats.TotalCost, &stats.TotalLatency,
			&stats.LastUsed, &qualityHistoryJSON, &stats.CreatedAt, &stats.UpdatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(qualityHistoryJSON), &stats.QualityHistory)
		allStats = append(allStats, stats)
	}
	return allStats, nil
}

// GetQualityTrend returns quality trend for a model.
func (h *Hub) GetQualityTrend(model string) (trend float64, points int, err error) {
	query := `SELECT quality_history FROM model_performance_stats WHERE model = ? ORDER BY updated_at DESC LIMIT 10`
	rows, err := h.db.Query(query, model)
	if err != nil {
		return 0, 0, fmt.Errorf("query quality trend: %w", err)
	}
	defer rows.Close()

	var allQuality []float64
	for rows.Next() {
		var qualityHistoryJSON string
		if err := rows.Scan(&qualityHistoryJSON); err != nil {
			continue
		}
		var history []float64
		json.Unmarshal([]byte(qualityHistoryJSON), &history)
		allQuality = append(allQuality, history...)
	}

	if len(allQuality) < 5 {
		return 0, len(allQuality), nil
	}

	mid := len(allQuality) / 2
	firstHalf := average(allQuality[:mid])
	secondHalf := average(allQuality[mid:])
	return secondHalf - firstHalf, len(allQuality), nil
}

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
