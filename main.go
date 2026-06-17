// TiBrain - Central Intelligence Hub for Multi-CLI Coordination
// Provides global handoff tracking, CLI registry, and aggregated context view
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
	_ "modernc.org/sqlite"

	"github.com/ti/router/tibrain/internal/db"
	"github.com/ti/router/tibrain/internal/memory"
	"github.com/ti/router/tibrain/internal/tools"
)

// ─────────────────────────────────────────────────────────────
// Structured Logger
// ─────────────────────────────────────────────────────────────

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

type Logger struct {
	mu     sync.Mutex
	level  LogLevel
	prefix string
}

func NewLogger(level LogLevel, prefix string) *Logger {
	return &Logger{
		level:  level,
		prefix: prefix,
	}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if l == nil {
		log.Printf(format, args...)
		return
	}
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := ""
	switch level {
	case LogLevelDebug:
		levelStr = "DEBUG"
	case LogLevelInfo:
		levelStr = "INFO"
	case LogLevelWarn:
		levelStr = "WARN"
	case LogLevelError:
		levelStr = "ERROR"
	}

	message := fmt.Sprintf(format, args...)
	log.Printf("[%s] [%s] [%s] %s", timestamp, levelStr, l.prefix, message)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LogLevelWarn, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}

var logger *Logger

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Global references for HTTP handlers
var globalGraphStore *Neo4jGraphStore
var globalCognitiveMemory *memory.CognitiveMemoryManager
var globalRetrievalRouter *RetrievalRouter

// ─────────────────────────────────────────────────────────────
// Configuration
// ─────────────────────────────────────────────────────────────

type Config struct {
	Port               int
	DataDir            string
	IndexKnowledge     bool
	IndexOnly          bool
	KnowledgeSources   []string
	CLIRegistry        bool
	HandoffTrack       bool
	SkillSync          bool
	MCPHubEnabled      bool
	MCPHubURL          string
	MCPHubAutoSync     bool
	MCPHubSyncInterval string
	MCPHubRegistryFile string
}

func defaultConfig() *Config {
	wd, _ := os.Getwd()
	tibrainDataDir := filepath.Join(wd, "tibrain_data")
	return &Config{
		Port:               1810,
		DataDir:            tibrainDataDir,
		KnowledgeSources:   []string{},
		CLIRegistry:        true,
		HandoffTrack:       true,
		SkillSync:          false,
		MCPHubEnabled:      false,
		MCPHubURL:          "http://localhost:3000",
		MCPHubAutoSync:     true,
		MCPHubSyncInterval: "5m",
		MCPHubRegistryFile: filepath.Join(tibrainDataDir, "mcp_registry.json"),
	}
}

func loadConfig() *Config {
	config := defaultConfig()

	configFile := filepath.Join(getTiBrainDir(), "config.yaml")
	if data, err := os.ReadFile(configFile); err == nil {
		var yamlConfig struct {
			TiBrain struct {
				Port    int    `yaml:"port"`
				DataDir string `yaml:"data_dir"`
			} `yaml:"tibrain"`
			MCPHub struct {
				Enabled      bool   `yaml:"enabled"`
				HubURL       string `yaml:"hub_url"`
				AutoSync     bool   `yaml:"auto_sync"`
				SyncInterval string `yaml:"sync_interval"`
				RegistryFile string `yaml:"registry_file"`
			} `yaml:"mcp_hub"`
			CLIRegistry struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"cli_registry"`
			HandoffTracking struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"handoff_tracking"`
			SkillSync struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"skill_sync"`
		}

		if err := yaml.Unmarshal(data, &yamlConfig); err == nil {
			if yamlConfig.TiBrain.Port > 0 {
				config.Port = yamlConfig.TiBrain.Port
			}
			if yamlConfig.TiBrain.DataDir != "" {
				config.DataDir = yamlConfig.TiBrain.DataDir
			}
			config.MCPHubEnabled = yamlConfig.MCPHub.Enabled
			if yamlConfig.MCPHub.HubURL != "" {
				config.MCPHubURL = yamlConfig.MCPHub.HubURL
			}
			config.MCPHubAutoSync = yamlConfig.MCPHub.AutoSync
			if yamlConfig.MCPHub.SyncInterval != "" {
				config.MCPHubSyncInterval = yamlConfig.MCPHub.SyncInterval
			}
			if yamlConfig.MCPHub.RegistryFile != "" {
				config.MCPHubRegistryFile = yamlConfig.MCPHub.RegistryFile
			}
			config.CLIRegistry = yamlConfig.CLIRegistry.Enabled
			config.HandoffTrack = yamlConfig.HandoffTracking.Enabled
			config.SkillSync = yamlConfig.SkillSync.Enabled
		}
	}

	if os.Getenv("TIBRAIN_INDEX_KNOWLEDGE") == "1" || strings.EqualFold(os.Getenv("TIBRAIN_INDEX_KNOWLEDGE"), "true") {
		config.IndexKnowledge = true
	}

	return config
}

func getTiBrainDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "Z:\\Ti\\router\\tibrain"
}

// ─────────────────────────────────────────────────────────────
// Database (Central Hub)
// ─────────────────────────────────────────────────────────────

type Hub struct {
	db          *sql.DB
	asyncWriter *AsyncWriter
}

func NewHub(dataDir string) (*Hub, error) {
	// Use the internal/db package for database management
	internalHub, err := db.NewHub(dataDir)
	if err != nil {
		return nil, fmt.Errorf("create internal hub: %w", err)
	}

	if err := updateIntegrationSchema(internalHub.DB()); err != nil {
		internalHub.Close()
		return nil, fmt.Errorf("init integration schema: %w", err)
	}
	if err := initRAGSchema(internalHub.DB()); err != nil {
		internalHub.Close()
		return nil, fmt.Errorf("init rag schema: %w", err)
	}

	asyncWriter := NewAsyncWriter(internalHub.DB(), 5000)

	return &Hub{
		db:          internalHub.DB(),
		asyncWriter: asyncWriter,
	}, nil
}

func (h *Hub) Close() error {
	logger.Info("closing Hub database...")
	if h.asyncWriter != nil {
		logger.Info("flushing and closing AsyncWriter queue...")
		h.asyncWriter.Close()
	}
	return h.db.Close()
}

// memoryGraphAdapter wraps *Neo4jGraphStore so it satisfies memory.GraphStore.
// memory.GraphStore.AddNode takes `any` (to avoid a hard dep on its GraphNode
// type); the real Neo4j method takes a concrete GraphNode.
type memoryGraphAdapter struct {
	gs *Neo4jGraphStore
}

func (a memoryGraphAdapter) AddNode(ctx context.Context, node any) (string, error) {
	gn, ok := node.(GraphNode)
	if !ok {
		return "", fmt.Errorf("memoryGraphAdapter: expected GraphNode, got %T", node)
	}
	return a.gs.AddNode(ctx, gn)
}

// BeginTx starts a new SQL transaction. Used by packages that need a
// memory.DB (e.g. internal/memory) without depending on the full Hub type.
func (h *Hub) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return h.db.BeginTx(ctx, opts)
}

// ExecContext delegates to h.db so *Hub satisfies memory.DB.
func (h *Hub) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return h.db.ExecContext(ctx, query, args...)
}

// QueryContext delegates to h.db so *Hub satisfies memory.DB.
func (h *Hub) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return h.db.QueryContext(ctx, query, args...)
}

// QueryRowContext delegates to h.db so *Hub satisfies memory.DB.
func (h *Hub) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return h.db.QueryRowContext(ctx, query, args...)
}

// PrepareContext delegates to h.db so *Hub satisfies memory.DB.
func (h *Hub) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return h.db.PrepareContext(ctx, query)
}

// Query performs a RAG search against the hub database
func (h *Hub) Query(ctx context.Context, query string, tier string, topK int, threshold float64) (*RAGResponse, error) {
	start := time.Now()

	rows, err := h.db.QueryContext(ctx, `
		SELECT id, title, content, path, category, COALESCE(tags, '[]') as tags,
			   COALESCE(metadata, '{}') as metadata
		FROM rag_documents
		WHERE status = 'active'
		  AND (content LIKE '%' || ? || '%' OR title LIKE '%' || ? || '%')
		ORDER BY updated_at DESC
		LIMIT ?`, query, query, topK)
	if err != nil {
		return nil, fmt.Errorf("query documents: %w", err)
	}
	defer rows.Close()

	var results []*RAGResult
	for rows.Next() {
		var id, title, content, path, category, tagsStr, metadataStr string
		if err := rows.Scan(&id, &title, &content, &path, &category, &tagsStr, &metadataStr); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}

		var tags []string
		json.Unmarshal([]byte(tagsStr), &tags)

		results = append(results, &RAGResult{
			DocumentID: id,
			FilePath:   path,
			Title:      title,
			Content:    content,
			Score:      1.0,
			Relevance:  1.0,
			Category:   category,
			Tags:       tags,
		})
	}

	elapsed := time.Since(start)
	return &RAGResponse{
		Results:      results,
		Query:        query,
		ResponseTime: elapsed,
		Confidence:   0.5,
		Tier:         tier,
		Metadata: &QueryMetadata{
			TotalResults: len(results),
			SearchTime:   elapsed,
		},
	}, nil
}

// ─────────────────────────────────────────────────────────────
// CLI Registry Operations
// ─────────────────────────────────────────────────────────────

type CLIInfo struct {
	CLIID         string `json:"cli_id"`
	CLIName       string `json:"cli_name"`
	BrainURL      string `json:"brain_url"`
	LastHeartbeat int64  `json:"last_heartbeat"`
	Status        string `json:"status"`
	CreatedAt     int64  `json:"created_at"`
}

func (h *Hub) RegisterCLI(cliID, cliName, brainURL string) error {
	timestamp := time.Now().Unix()

	_, err := h.db.Exec(`
		INSERT OR REPLACE INTO cli_registry (cli_id, cli_name, brain_url, last_heartbeat, status, created_at)
		VALUES (?, ?, ?, ?, 'active', ?)
	`, cliID, cliName, brainURL, timestamp, timestamp)

	if err != nil {
		return fmt.Errorf("register cli: %w", err)
	}

	logger.Info("CLI registered: %s (%s)", cliID, cliName)
	return nil
}

func (h *Hub) UnregisterCLI(cliID string) error {
	_, err := h.db.Exec(`
		UPDATE cli_registry SET status = 'inactive' WHERE cli_id = ?
	`, cliID)

	if err != nil {
		return fmt.Errorf("unregister cli: %w", err)
	}

	logger.Info("CLI unregistered: %s", cliID)
	return nil
}

func (h *Hub) UpdateHeartbeat(cliID string) error {
	timestamp := time.Now().Unix()

	_, err := h.db.Exec(`
		UPDATE cli_registry SET last_heartbeat = ? WHERE cli_id = ?
	`, timestamp, cliID)

	if err != nil {
		return fmt.Errorf("update heartbeat: %w", err)
	}

	return nil
}

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

type GlobalHandoff struct {
	ID        string            `json:"id"`
	FromCLI   string            `json:"from_cli"`
	ToCLI     string            `json:"to_cli"`
	Context   string            `json:"context"`
	Output    string            `json:"output"`
	Timestamp int64             `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

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
		logger.Error("Insert global handoff failed: %v", err)
		return "", fmt.Errorf("insert global handoff: %w", err)
	}

	logger.Info("Global handoff created: %s → %s", fromCLI, toCLI)
	return id, nil
}

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

	if err != nil {
		return fmt.Errorf("register mcp: %w", err)
	}

	logger.Info("MCP registered: %s (%s)", id, name)
	return nil
}

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

	if err != nil {
		return fmt.Errorf("register mcp hub server: %w", err)
	}

	logger.Info("MCP Hub server registered: %s (%s)", id, name)
	return nil
}

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

func (h *Hub) UpdateMCPHubServerSyncTime(id string) error {
	timestamp := time.Now().Unix()

	_, err := h.db.Exec(`
		UPDATE mcp_hub_registry SET synced_at = ?, updated_at = ? WHERE id = ?
	`, timestamp, timestamp, id)

	if err != nil {
		return fmt.Errorf("update sync time: %w", err)
	}

	return nil
}

// SyncMCPTools syncs tools from all enabled MCP servers to Tool Registry
func (h *Hub) SyncMCPTools() error {
	// Get all enabled MCP servers
	mcps, err := h.ListMCPs("")
	if err != nil {
		return fmt.Errorf("list mcps: %w", err)
	}

	for _, mcp := range mcps {
		if !mcp.Enabled {
			continue
		}

		logger.Info("Syncing tools from MCP: %s (%s)", mcp.ID, mcp.Name)

		// TODO: Integrate with MCP Hub instead of individual MCP clients
		// For now, skip MCP sync
		logger.Warn("MCP sync disabled - use MCP Hub instead")
		continue
	}

	return nil
}

// CallMCPTool calls a tool on an MCP server
// TODO: Integrate with MCP Hub instead of individual MCP clients
func (h *Hub) CallMCPTool(toolID string, params map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("MCP tool calling disabled")
}

// ─────────────────────────────────────────────────────────────
// Tool Registry Operations
// ─────────────────────────────────────────────────────────────

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

	if err != nil {
		return fmt.Errorf("register tool: %w", err)
	}

	logger.Info("Tool registered: %s (%s) - Quality: %d, Security: %d, Source: %s, Tier: %s", tool.ID, tool.Name, tool.QualityScore, tool.SecurityScore, tool.Source, tool.QualityTier)
	return nil
}

// RegisterTool implements tools.ToolRegistrar for use with internal/tools package.
func (h *Hub) RegisterToolByFields(id, name, description, parameters, handler, category, permissions string, enabled bool) error {
	return h.RegisterTool(Tool{
		ID:                 id,
		Name:               name,
		Description:        description,
		Parameters:         parameters,
		Handler:            handler,
		Category:           category,
		Permissions:        permissions,
		Enabled:            enabled,
		QualityScore:       80,
		SecurityScore:      80,
		BestPracticesScore: 80,
		Source:             "best-source",
		SkillLevel:         "l2",
		QualityTier:        "platinum",
		SecurityTier:       "hardened",
		SecurityStatus:     "passed",
		ValidationStatus:   "passed",
		SourceType:         "community",
	})
}

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
		var enabledInt int
		var bestPracticesScore int
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

func (h *Hub) GetTool(id string) (*Tool, error) {
	query := `SELECT id, name, description, parameters, handler, category, permissions, enabled, quality_score, security_score, best_practices_score, source, family, tags, version, skill_level, quality_tier, security_tier, security_status, validation_status, variant_id, variant_label, source_type, root_path, created_at, updated_at
		FROM tool_registry WHERE id = ?`

	row := h.db.QueryRow(query, id)

	var t Tool
	var enabledInt int
	var bestPracticesScore int
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

	if err != nil {
		return fmt.Errorf("log tool usage: %w", err)
	}

	logger.Info("Tool usage logged: %s by %s", toolName, agentID)
	return nil
}

func (h *Hub) GetToolUsageLogs(toolName, agentID string, limit int) ([]ToolUsageLog, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `SELECT id, tool_name, agent_id, parameters, result, success, error_message, timestamp
		FROM tool_usage_log`
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
// HTTP Handlers
// ─────────────────────────────────────────────────────────────

type Server struct {
	hub    *Hub
	config *Config
}

func NewServer(hub *Hub, config *Config) *Server {
	return &Server{
		hub:    hub,
		config: config,
	}
}

// CLI Registry Handlers
func (s *Server) handleRegisterCLI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CLIID    string `json:"cli_id"`
		CLIName  string `json:"cli_name"`
		BrainURL string `json:"brain_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.hub.RegisterCLI(req.CLIID, req.CLIName, req.BrainURL); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

func (s *Server) handleUnregisterCLI(w http.ResponseWriter, r *http.Request) {
	cliID := r.URL.Query().Get("cli_id")

	if err := s.hub.UnregisterCLI(cliID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "unregistered"})
}

func (s *Server) handleListCLIs(w http.ResponseWriter, r *http.Request) {
	clis, err := s.hub.ListCLIs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clis)
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	cliID := r.URL.Query().Get("cli_id")

	if err := s.hub.UpdateHeartbeat(cliID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Global Handoff Handlers
func (s *Server) handleCreateGlobalHandoff(w http.ResponseWriter, r *http.Request) {
	var req GlobalHandoff
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.hub.CreateGlobalHandoff(req.FromCLI, req.ToCLI, req.Context, req.Output, req.Metadata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"id": id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleRecallGlobalHandoff(w http.ResponseWriter, r *http.Request) {
	fromCLI := r.URL.Query().Get("from")
	toCLI := r.URL.Query().Get("to")

	handoff, err := s.hub.RecallGlobalHandoff(fromCLI, toCLI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(handoff)
}

func (s *Server) handleListHandoffs(w http.ResponseWriter, r *http.Request) {
	cliID := r.URL.Query().Get("cli")

	handoffs, err := s.hub.ListHandoffs(cliID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(handoffs)
}

// MCP Registry Handlers
func (s *Server) handleRegisterMCP(w http.ResponseWriter, r *http.Request) {
	var req MCPServer
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.hub.RegisterMCP(req.ID, req.Name, req.Type, req.Command, req.Args, req.Env, req.Description, req.Enabled, req.CLIID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

func (s *Server) handleListMCPs(w http.ResponseWriter, r *http.Request) {
	cliID := r.URL.Query().Get("cli_id")

	mcps, err := s.hub.ListMCPs(cliID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mcps)
}

func (s *Server) handleGetMCP(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	mcp, err := s.hub.GetMCP(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mcp)
}

func (s *Server) handleSyncMCPTools(w http.ResponseWriter, r *http.Request) {
	if err := s.hub.SyncMCPTools(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "synced"})
}

// Tool Registry Handlers
func (s *Server) handleRegisterTool(w http.ResponseWriter, r *http.Request) {
	var req Tool
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.hub.RegisterTool(req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	tools, err := s.hub.ListTools(category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tools)
}

func (s *Server) handleGetTool(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	tool, err := s.hub.GetTool(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tool)
}

// Tool Execution Handler
func (s *Server) handleExecuteTool(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string                 `json:"name"`
		Params  map[string]interface{} `json:"params"`
		AgentID string                 `json:"agent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var response map[string]interface{}
	var err error

	// Check if this is an MCP tool
	if strings.HasPrefix(req.Name, "mcp-") {
		// Execute MCP tool
		response, err = s.hub.CallMCPTool(req.Name, req.Params)
		if err != nil {
			response = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
				"tool":    req.Name,
			}
		} else {
			response["success"] = true
			response["tool"] = req.Name
		}
	} else {
		// Delegate to Router for non-MCP tools
		response = map[string]interface{}{
			"success": true,
			"tool":    req.Name,
			"params":  req.Params,
			"message": "Tool execution delegated to Router",
		}
	}

	// Log tool usage
	paramsJSON, _ := json.Marshal(req.Params)
	resultJSON, _ := json.Marshal(response)
	success := response["success"].(bool)
	errorMsg := ""
	if !success {
		if errStr, ok := response["error"].(string); ok {
			errorMsg = errStr
		}
	}
	s.hub.LogToolUsage(req.Name, req.AgentID, string(paramsJSON), string(resultJSON), success, errorMsg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Tool Usage Log Handlers
func (s *Server) handleGetToolUsageLogs(w http.ResponseWriter, r *http.Request) {
	toolName := r.URL.Query().Get("tool_name")
	agentID := r.URL.Query().Get("agent_id")

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	logs, err := s.hub.GetToolUsageLogs(toolName, agentID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// ─────────────────────────────────────────────────────────────
// BEADS LEARN Stats Handlers
// ─────────────────────────────────────────────────────────────

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

// handleUpdateModelStats updates or creates model performance stats
func (s *Server) handleUpdateModelStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var stats ModelPerformanceStats
	if err := json.NewDecoder(r.Body).Decode(&stats); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	now := time.Now().Unix()
	if stats.ID == "" {
		stats.ID = fmt.Sprintf("%s:%s", stats.Model, stats.TaskType)
	}
	if stats.CreatedAt == 0 {
		stats.CreatedAt = now
	}
	stats.UpdatedAt = now

	// Serialize quality_history as JSON
	qualityHistoryJSON, _ := json.Marshal(stats.QualityHistory)

	// Upsert into database
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

	_, err := s.hub.db.Exec(query,
		stats.ID, stats.Model, stats.TaskType, stats.Phase,
		stats.TotalTasks, stats.SuccessCount, stats.FailCount,
		stats.TotalQuality, stats.TotalCost, stats.TotalLatency,
		stats.LastUsed, string(qualityHistoryJSON), stats.CreatedAt, stats.UpdatedAt,
		// ON CONFLICT values
		stats.TotalTasks, stats.SuccessCount, stats.FailCount,
		stats.TotalQuality, stats.TotalCost, stats.TotalLatency,
		stats.LastUsed, string(qualityHistoryJSON), stats.UpdatedAt,
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"id":     stats.ID,
	})
}

// handleGetModelStats retrieves model performance stats
func (s *Server) handleGetModelStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	model := r.URL.Query().Get("model")
	taskType := r.URL.Query().Get("task_type")

	if model == "" || taskType == "" {
		http.Error(w, "model and task_type required", http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("%s:%s", model, taskType)

	query := `SELECT id, model, task_type, phase, total_tasks, success_count, fail_count,
		total_quality, total_cost, total_latency, last_used, quality_history, created_at, updated_at
		FROM model_performance_stats WHERE id = ?`

	var stats ModelPerformanceStats
	var qualityHistoryJSON string
	err := s.hub.db.QueryRow(query, id).Scan(
		&stats.ID, &stats.Model, &stats.TaskType, &stats.Phase,
		&stats.TotalTasks, &stats.SuccessCount, &stats.FailCount,
		&stats.TotalQuality, &stats.TotalCost, &stats.TotalLatency,
		&stats.LastUsed, &qualityHistoryJSON, &stats.CreatedAt, &stats.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Stats not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		}
		return
	}

	json.Unmarshal([]byte(qualityHistoryJSON), &stats.QualityHistory)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleGetBestModel returns the best performing model for a task type
func (s *Server) handleGetBestModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskType := r.URL.Query().Get("task_type")
	if taskType == "" {
		http.Error(w, "task_type required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT model, task_type, total_tasks, success_count, total_quality
		FROM model_performance_stats
		WHERE task_type = ? AND total_tasks >= 3
		ORDER BY (total_quality / total_tasks) DESC
		LIMIT 1
	`

	var model string
	var foundTaskType string
	var totalTasks int
	var successCount int
	var totalQuality float64

	err := s.hub.db.QueryRow(query, taskType).Scan(
		&model, &foundTaskType, &totalTasks, &successCount, &totalQuality,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"model":       nil,
				"task_type":   taskType,
				"avg_quality": 0,
				"total_tasks": 0,
				"message":     "Not enough data points (need at least 3)",
			})
			return
		}
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	avgQuality := totalQuality / float64(totalTasks)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"model":       model,
		"task_type":   taskType,
		"avg_quality": avgQuality,
		"total_tasks": totalTasks,
	})
}

// handleGetQualityTrend returns quality trend for a model
func (s *Server) handleGetQualityTrend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	model := r.URL.Query().Get("model")
	if model == "" {
		http.Error(w, "model required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT quality_history FROM model_performance_stats
		WHERE model = ?
		ORDER BY updated_at DESC
		LIMIT 10
	`

	rows, err := s.hub.db.Query(query, model)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"model":   model,
			"trend":   0,
			"message": "Not enough data points (need at least 5)",
		})
		return
	}

	// Calculate trend (compare first half vs second half)
	mid := len(allQuality) / 2
	firstHalf := average(allQuality[:mid])
	secondHalf := average(allQuality[mid:])
	trend := secondHalf - firstHalf

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"model":  model,
		"trend":  trend,
		"points": len(allQuality),
	})
}

// handleGetAllModelStats returns all model performance stats
func (s *Server) handleGetAllModelStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `
		SELECT id, model, task_type, phase, total_tasks, success_count, fail_count,
		total_quality, total_cost, total_latency, last_used, quality_history, created_at, updated_at
		FROM model_performance_stats
		ORDER BY updated_at DESC
	`

	rows, err := s.hub.db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var allStats []ModelPerformanceStats
	for rows.Next() {
		var stats ModelPerformanceStats
		var qualityHistoryJSON string
		err := rows.Scan(
			&stats.ID, &stats.Model, &stats.TaskType, &stats.Phase,
			&stats.TotalTasks, &stats.SuccessCount, &stats.FailCount,
			&stats.TotalQuality, &stats.TotalCost, &stats.TotalLatency,
			&stats.LastUsed, &qualityHistoryJSON, &stats.CreatedAt, &stats.UpdatedAt,
		)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(qualityHistoryJSON), &stats.QualityHistory)
		allStats = append(allStats, stats)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats": allStats,
		"count": len(allStats),
	})
}

// average computes the mean of a slice of float64
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

// Status Handler
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	clis, _ := s.hub.ListCLIs()

	response := map[string]interface{}{
		"status":          "ok",
		"tibrain":         "TiBrain Central Hub v1.0.0",
		"active_clis":     len(clis),
		"cli_registry":    s.config.CLIRegistry,
		"handoff_track":   s.config.HandoffTrack,
		"skill_sync":      s.config.SkillSync,
		"mcp_hub_enabled": s.config.MCPHubEnabled,
		"mcp_hub_url":     s.config.MCPHubURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Notion Sync Handler
func (s *Server) handleNotionSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Failed to get working directory: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Run sync script
	cmd := exec.Command("python", filepath.Join(wd, "sync_notion.py"))
	cmd.Dir = wd
	output, err := cmd.CombinedOutput()

	response := map[string]interface{}{
		"success": err == nil,
		"output":  string(output),
	}

	if err != nil {
		response["error"] = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ─────────────────────────────────────────────────────────────
// MCP Hub HTTP Handlers
// ─────────────────────────────────────────────────────────────

// handleMCPHubSync syncs MCP servers from MCP Hub
func (s *Server) handleMCPHubSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.config.MCPHubEnabled {
		http.Error(w, "MCP Hub is not enabled", http.StatusServiceUnavailable)
		return
	}

	// Sync tools from local MCP Hub registry
	servers, err := s.hub.ListMCPHubServers()
	if err != nil {
		logger.Error("Failed to list MCP Hub servers: %v", err)
		http.Error(w, fmt.Sprintf("Failed to list servers: %v", err), http.StatusInternalServerError)
		return
	}

	syncedCount := 0
	for _, server := range servers {
		if !server.Enabled {
			continue
		}

		logger.Info("Syncing tools from MCP Hub server: %s", server.Name)

		// Parse tools JSON
		var tools []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal([]byte(server.Tools), &tools); err != nil {
			logger.Warn("Failed to parse tools for %s: %v", server.Name, err)
			continue
		}

		// Register each tool
		for _, tool := range tools {
			toolID := fmt.Sprintf("mcp-hub-%s-%s", server.ID, tool.Name)
			toolReg := Tool{
				ID:                 toolID,
				Name:               tool.Name,
				Description:        tool.Description,
				Parameters:         "{}",
				Handler:            "mcp-hub",
				Category:           "mcp",
				Permissions:        `["devin","claude","cursor"]`,
				Enabled:            true,
				QualityScore:       80,
				SecurityScore:      80,
				BestPracticesScore: 80,
				Source:             "mcp-hub",
				Family:             "",
				Tags:               "",
				Version:            "",
				SkillLevel:         "l2",
				QualityTier:        "platinum",
				SecurityTier:       "hardened",
				SecurityStatus:     "passed",
				ValidationStatus:   "passed",
				VariantID:          "",
				VariantLabel:       "",
				SourceType:         "community",
				RootPath:           "",
			}

			if err := s.hub.RegisterTool(toolReg); err != nil {
				logger.Warn("Failed to register tool %s: %v", tool.Name, err)
			} else {
				syncedCount++
			}
		}

		// Update sync time
		if err := s.hub.UpdateMCPHubServerSyncTime(server.ID); err != nil {
			logger.Warn("Failed to update sync time for %s: %v", server.Name, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "MCP Hub sync completed",
		"synced":  syncedCount,
		"servers": len(servers),
	})
}

// handleMCPHubList lists MCP Hub servers
func (s *Server) handleMCPHubList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	servers, err := s.hub.ListMCPHubServers()
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"servers": servers,
		"count":   len(servers),
	})
}

// handleMCPHubGet gets a specific MCP Hub server
func (s *Server) handleMCPHubGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	server, err := s.hub.GetMCPHubServer(id)
	if err != nil {
		if err.Error() == "mcp hub server not found" {
			http.Error(w, "Server not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(server)
}

// handleMCPHubRegister registers a new MCP Hub server
func (s *Server) handleMCPHubRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Endpoint    string `json:"endpoint"`
		Transport   string `json:"transport"`
		Tools       string `json:"tools"`
		Resources   string `json:"resources"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == "" || req.Name == "" || req.Endpoint == "" {
		http.Error(w, "id, name, and endpoint required", http.StatusBadRequest)
		return
	}

	if req.Transport == "" {
		req.Transport = "stdio"
	}

	if err := s.hub.RegisterMCPHubServer(req.ID, req.Name, req.Endpoint, req.Transport, req.Tools, req.Resources, req.Description, req.Enabled); err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"id":      req.ID,
		"message": "MCP Hub server registered",
	})
}

// ─────────────────────────────────────────────────────────────
// Main
// ─────────────────────────────────────────────────────────────

// Cognitive Memory Handlers
func (s *Server) handleStoreMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content string                 `json:"content"`
		Context map[string]interface{} `json:"context"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if globalCognitiveMemory == nil {
		http.Error(w, "cognitive memory not initialized", http.StatusServiceUnavailable)
		return
	}

	ctx := context.Background()
	id, err := globalCognitiveMemory.StoreEpisodicMemory(ctx, req.Content, req.Context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id, "status": "stored"})
}

func (s *Server) handleQueryMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query      string `json:"query"`
		MemoryType string `json:"memory_type"`
		Limit      int    `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if globalCognitiveMemory == nil {
		http.Error(w, "cognitive memory not initialized", http.StatusServiceUnavailable)
		return
	}

	ctx := context.Background()
	memType := memory.MemoryType(req.MemoryType)
	if memType == "" {
		memType = memory.MemorySemantic
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	entries, err := globalCognitiveMemory.QueryMemory(ctx, req.Query, memType, req.Limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": entries,
		"count":   len(entries),
	})
}

func (s *Server) handleMemoryStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if globalCognitiveMemory == nil {
		http.Error(w, "cognitive memory not initialized", http.StatusServiceUnavailable)
		return
	}

	ctx := context.Background()
	stats, err := globalCognitiveMemory.GetMemoryStats(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleRecentExperience(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	if globalCognitiveMemory == nil {
		http.Error(w, "cognitive memory not initialized", http.StatusServiceUnavailable)
		return
	}

	ctx := context.Background()
	entries, err := globalCognitiveMemory.GetRecentExperience(ctx, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": entries,
		"count":   len(entries),
	})
}

// Natural Language Interaction Handlers
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"response": fmt.Sprintf("Echo: %s", req.Message),
		"status":   "success",
	})
}

func (s *Server) handleAgentProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if globalRetrievalRouter == nil {
		http.Error(w, "retrieval router not initialized", http.StatusServiceUnavailable)
		return
	}

	decision := globalRetrievalRouter.RouteQuery(ctx, req.Query)
	response, err := globalRetrievalRouter.ExecuteRoute(ctx, req.Query, decision, nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"response": response,
		"status":   "success",
	})
}

func main() {
	config := loadConfig()

	logger = NewLogger(LogLevelInfo, "TIBRAIN")
	logger.Info("TiBrain starting with data dir: %s", config.DataDir)

	for i, arg := range os.Args {
		if arg == "--port" && i+1 < len(os.Args) {
			if port, err := strconv.Atoi(os.Args[i+1]); err == nil {
				config.Port = port
			}
		}
		if arg == "--index-knowledge" {
			config.IndexKnowledge = true
		}
		if arg == "--index-only" {
			config.IndexKnowledge = true
			config.IndexOnly = true
		}
		if arg == "--index-source" && i+1 < len(os.Args) {
			config.KnowledgeSources = append(config.KnowledgeSources, os.Args[i+1])
		}
	}

	hub, err := NewHub(config.DataDir)
	if err != nil {
		logger.Error("Failed to initialize hub: %v", err)
		os.Exit(1)
	}
	defer hub.Close()

	if config.IndexKnowledge {
		indexer := NewKnowledgeIndexer(hub)
		options := KnowledgeIndexOptions{}
		for _, sourcePath := range config.KnowledgeSources {
			options.Sources = append(options.Sources, KnowledgeIndexSource{
				Path:        sourcePath,
				Category:    categoryFromPath(sourcePath),
				Description: "CLI-provided knowledge source",
			})
		}
		result, err := indexer.Index(options)
		if err != nil {
			logger.Error("Knowledge indexing failed: %v", err)
			os.Exit(1)
		}
		logger.Info("Knowledge indexing complete: indexed=%d skipped=%d errors=%d sources=%d", result.Indexed, result.Skipped, len(result.Errors), result.Sources)
		if len(result.Errors) > 0 {
			for _, indexErr := range result.Errors {
				logger.Warn("Knowledge indexing issue: %s", indexErr)
			}
		}
		if config.IndexOnly {
			return
		}
	}

	// Register default tools
	if err := tools.RegisterDefaultTools(hub); err != nil {
		logger.Error("Failed to register default tools: %v", err)
		// Continue anyway, tools can be registered later
	}

	// Initialize Neo4j Graph Store
	var graphStore *Neo4jGraphStore
	neo4jConfig := Neo4jConfig{
		URI:      getEnvOrDefault("NEO4J_URI", "bolt://localhost:7687"),
		Username: getEnvOrDefault("NEO4J_USER", "neo4j"),
		Password: getEnvOrDefault("NEO4J_PASSWORD", "password"),
		Database: getEnvOrDefault("NEO4J_DATABASE", "neo4j"),
		MaxDepth: 3,
	}

	// Attempt to initialize Neo4j - continue without if not available
	graphStore, err = NewNeo4jGraphStore(neo4jConfig)
	if err != nil {
		logger.Warn("Neo4j not available, continuing without graph store: %v", err)
		graphStore = nil
	} else {
		defer graphStore.Close(context.Background())
		logger.Info("Neo4j graph store initialized successfully")
	}

	// Initialize Cognitive Memory Manager.
	// graphAdapter wraps *Neo4jGraphStore to satisfy memory.GraphStore (whose
	// interface uses `any` for the node argument, to avoid a hard dependency
	// on memory.GraphNode).
	graphAdapter := memoryGraphAdapter{gs: graphStore}
	cognitiveMemory := memory.NewCognitiveMemoryManager(hub, graphAdapter)
	logger.Info("Cognitive memory manager initialized")

	// Initialize Retrieval Router with Graph support
	retrievalRouter := NewRetrievalRouterWithGraph(hub, graphStore)

	// Initialize Ti Agent Orchestrator with all components
	tiAgentOrchestrator := NewTiAgentOrchestrator(hub, retrievalRouter)
	tiAgentOrchestrator.cognitiveMemory = cognitiveMemory // Inject cognitive memory

	// Store references for HTTP handlers
	globalGraphStore = graphStore
	globalCognitiveMemory = cognitiveMemory
	globalRetrievalRouter = retrievalRouter

	server := NewServer(hub, config)

	http.HandleFunc("/v1/tibrain/cli/register", server.handleRegisterCLI)
	http.HandleFunc("/v1/tibrain/cli/unregister", server.handleUnregisterCLI)
	http.HandleFunc("/v1/tibrain/cli/list", server.handleListCLIs)
	http.HandleFunc("/v1/tibrain/cli/heartbeat", server.handleHeartbeat)

	http.HandleFunc("/v1/tibrain/handoff/create", server.handleCreateGlobalHandoff)
	http.HandleFunc("/v1/tibrain/handoff/recall", server.handleRecallGlobalHandoff)
	http.HandleFunc("/v1/tibrain/handoffs", server.handleListHandoffs)

	http.HandleFunc("/v1/tibrain/mcp/register", server.handleRegisterMCP)
	http.HandleFunc("/v1/tibrain/mcp/list", server.handleListMCPs)
	http.HandleFunc("/v1/tibrain/mcp", server.handleGetMCP)
	http.HandleFunc("/v1/tibrain/mcp/sync", server.handleSyncMCPTools)

	// MCP Hub API
	http.HandleFunc("/v1/tibrain/mcp-hub/sync", server.handleMCPHubSync)
	http.HandleFunc("/v1/tibrain/mcp-hub/list", server.handleMCPHubList)
	http.HandleFunc("/v1/tibrain/mcp-hub", server.handleMCPHubGet)
	http.HandleFunc("/v1/tibrain/mcp-hub/register", server.handleMCPHubRegister)

	http.HandleFunc("/v1/tibrain/tool/register", server.handleRegisterTool)
	http.HandleFunc("/v1/tibrain/tool/list", server.handleListTools)
	http.HandleFunc("/v1/tibrain/tool", server.handleGetTool)
	http.HandleFunc("/v1/tibrain/tool/execute", server.handleExecuteTool)
	http.HandleFunc("/v1/tibrain/tool/logs", server.handleGetToolUsageLogs)

	// Cognitive Memory API
	http.HandleFunc("/v1/tibrain/memory/store", server.handleStoreMemory)
	http.HandleFunc("/v1/tibrain/memory/query", server.handleQueryMemory)
	http.HandleFunc("/v1/tibrain/memory/stats", server.handleMemoryStats)
	http.HandleFunc("/v1/tibrain/memory/recent", server.handleRecentExperience)

	// Natural Language Interaction API
	http.HandleFunc("/v1/tibrain/chat", server.handleChat)
	http.HandleFunc("/v1/tibrain/agent/process", server.handleAgentProcess)

	// BEADS LEARN Stats API
	http.HandleFunc("/v1/tibrain/beads/stats", server.handleUpdateModelStats)
	http.HandleFunc("/v1/tibrain/beads/stats/get", server.handleGetModelStats)
	http.HandleFunc("/v1/tibrain/beads/stats/best", server.handleGetBestModel)
	http.HandleFunc("/v1/tibrain/beads/stats/trend", server.handleGetQualityTrend)
	http.HandleFunc("/v1/tibrain/beads/stats/all", server.handleGetAllModelStats)

	http.HandleFunc("/v1/tibrain/status", server.handleStatus)
	http.HandleFunc("/health", server.handleStatus)

	// Notion Sync API
	http.HandleFunc("/v1/tibrain/sync/notion", server.handleNotionSync)

	integrationManager := NewIntegrationManager(hub)
	apiServer := NewAPIServer(hub, integrationManager)
	http.Handle("/api/", apiServer)

	// TODO: Re-enable when RAG and CLI Context handlers are implemented
	// server.setupRAGHandlers()
	// server.setupCLIContextHandlers()

	// TODO: Re-enable when Enhanced REST Server is implemented
	// go func() {
	// 	enhancedServer := NewEnhancedRESTServer(hub)
	// 	logger.Info("Starting Enhanced REST Server...")
	// 	if err := enhancedServer.Start(); err != nil && err != http.ErrServerClosed {
	// 		logger.Error("Enhanced REST Server failed: %v", err)
	// 	}
	// }()

	addr := fmt.Sprintf(":%d", config.Port)
	logger.Info("TiBrain Central Hub starting on port %d", config.Port)
	logger.Info("Data directory: %s", config.DataDir)
	logger.Info("Health check: http://localhost%s/health", addr)
	logger.Info("Status: http://localhost%s/v1/tibrain/status", addr)
	logger.Info("Integration API: http://localhost%s/api/status", addr)

	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("Shutting down gracefully...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error: %v", err)
		}

		logger.Info("Server stopped")
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Failed to start server: %v", err)
		os.Exit(1)
	}

	logger.Info("Shutdown complete")
}
