package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	_ "modernc.org/sqlite"

	"github.com/ti/router/tibrain/internal/mcp"
	"github.com/ti/router/tibrain/providers/notion"
	"github.com/ti/router/tibrain/pkg/automation"
)

// Config represents the application configuration
type Config struct {
	Tibrain struct {
		Port    int    `yaml:"port"`
		DataDir string `yaml:"data_dir"`
		APIKeys []string `yaml:"api_keys"`
	} `yaml:"tibrain"`
}

// loadConfig loads configuration from config.yaml
func loadConfig() (*Config, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// checkAPIKey validates the provided API key against the configured keys
func checkAPIKey(key string, validKeys []string) bool {
	for _, validKey := range validKeys {
		if key == validKey {
			return true
		}
	}
	return false
}

// authenticateMiddleware is a middleware that checks for API key authentication
func authenticateMiddleware(next http.Handler, validKeys []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Also check query parameter as fallback
			apiKey = r.URL.Query().Get("api_key")
		}

		if !checkAPIKey(apiKey, validKeys) {
			http.Error(w, "Unauthorized: Invalid or missing API key", http.StatusUnauthorized)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Printf("Warning: Failed to load config: %v. Using defaults.", err)
		// Use default config if file not found
		config = &Config{
			Tibrain: struct {
				Port    int    `yaml:"port"`
				DataDir string `yaml:"data_dir"`
				APIKeys []string `yaml:"api_keys"`
			}{Port: 1810, DataDir: "data", APIKeys: []string{"test-key-123", "dev-key-456"}},
		}
	}

	// Initialize database
	db, err := sql.Open("sqlite", "./data/tibrain.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Run migrations
	runMigrations(db)

	// Create hub
	hub := &Hub{
		db: db,
	}

	// MCP server (SSE + Streamable HTTP) for upstream gateway integration
	mcpMgr := mcp.NewMCPServerManager()
	http.HandleFunc("/sse", mcpMgr.HandleSSE)
	http.HandleFunc("/mcp", mcpMgr.HandleMessage)

	// Register routes
	http.HandleFunc("/health", hub.handleHealth)
	http.HandleFunc("/ready", hub.handleReady)
	http.HandleFunc("/mcp/proxy/call", hub.handleMCPProxyCall)
	http.HandleFunc("/mcp/servers", hub.handleListMCPServers)
	http.HandleFunc("/mcp/tools", hub.handleListMCPTools)
	http.HandleFunc("/tool/execute", hub.handleExecuteTool)
	http.HandleFunc("/api/v1/prompt/preflight", hub.handlePromptPreflight)
	http.HandleFunc("/api/v1/prompt/feedback", hub.handlePromptFeedback)
	http.HandleFunc("/api/v1/prompt/catalog/version", hub.handlePromptCatalogVersion)
	http.HandleFunc("/api/v1/feedback/model-stats", hub.handleModelStatsFeedback)
	// Apply authentication middleware to notion-sync endpoint
	http.Handle("/notion-sync", authenticateMiddleware(http.HandlerFunc(hub.handleNotionSync), config.Tibrain.APIKeys))
	// Automation endpoints (protected)
	http.Handle("POST /v1/automation/runs", authenticateMiddleware(http.HandlerFunc(automation.CreateRunHandler), config.Tibrain.APIKeys))
	http.Handle("GET /v1/automation/runs/", authenticateMiddleware(http.HandlerFunc(automation.GetRunHandler), config.Tibrain.APIKeys))

	http.HandleFunc("/api/v1/mcp/enable", hub.handleEnableMCP)
	http.HandleFunc("/api/v1/mcp/disable", hub.handleDisableMCP)

	// OpenAI API Bridge endpoints
	http.HandleFunc("/v1/chat/completions", hub.handleOpenAIChatCompletions)
	http.HandleFunc("/v1/models", hub.handleOpenAIModels)
	http.HandleFunc("/v1/tools/status", hub.handleToolStatus)

	// Start server
	port := getEnvOrDefault("PORT", fmt.Sprintf("%d", config.Tibrain.Port))
	addr := fmt.Sprintf(":%s", port)
	log.Printf("TiBrain starting on %s", addr)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

// Hub represents the TiBrain hub
type Hub struct {
	db *sql.DB
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runMigrations(db *sql.DB) {
	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS cli_registry (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		version TEXT,
		registered_at INTEGER
	);
	CREATE TABLE IF NOT EXISTS global_handoffs (
		id TEXT PRIMARY KEY,
		from_agent TEXT,
		to_agent TEXT,
		task TEXT,
		created_at INTEGER
	);
	CREATE TABLE IF NOT EXISTS mcp_registry (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT,
		config TEXT,
		registered_at INTEGER
	);
	CREATE TABLE IF NOT EXISTS tool_registry (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		server TEXT,
		description TEXT,
		registered_at INTEGER
	);
	CREATE TABLE IF NOT EXISTS model_performance_stats (
		id TEXT PRIMARY KEY,
		model TEXT,
		task_type TEXT,
		total_tasks INTEGER,
		total_quality REAL,
		updated_at INTEGER
	);
	CREATE TABLE IF NOT EXISTS tool_usage_log (
		id TEXT PRIMARY KEY,
		tool TEXT,
		agent TEXT,
		success BOOLEAN,
		error_message TEXT,
		timestamp INTEGER
	);
	`
	db.Exec(schema)
}

func (h *Hub) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}

func (h *Hub) handleReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ready": true,
	})
}

func (h *Hub) handleMCPProxyCall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"result":  "tool called",
	})
}

func (h *Hub) handleListMCPServers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"servers": []map[string]interface{}{
			{"name": "github-mcp", "type": "stdio"},
			{"name": "obsidian-mcp-server", "type": "stdio"},
			{"name": "chrome-devtools-mcp", "type": "stdio"},
		},
	})
}

func (h *Hub) handleListMCPTools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": []map[string]interface{}{
			{"name": "list_repositories", "server": "github-mcp"},
			{"name": "search_notes", "server": "obsidian-mcp-server"},
			{"name": "navigate", "server": "chrome-devtools-mcp"},
		},
	})
}

func (h *Hub) handleExecuteTool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"result":  "tool executed",
	})
}

func (h *Hub) handlePromptPreflight(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Intent string `json:"intent"`
		Domain string `json:"domain"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"envelope": map[string]interface{}{
			"id":          "default",
			"name":        "default-prompt",
			"description": "Default prompt envelope",
			"domain":      req.Domain,
			"intent":      []string{req.Intent},
		},
	})
}

func (h *Hub) handlePromptFeedback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Hub) handlePromptCatalogVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version": "1.0.0",
		"etag":    "w/\"prompt-catalog-v1\"",
		"count":   10,
	})
}

func (h *Hub) handleModelStatsFeedback(w http.ResponseWriter, r *http.Request) {
	var stats struct {
		Model    string  `json:"model"`
		TaskType string  `json:"task_type"`
		Rating   float64 `json:"rating"`
	}
	json.NewDecoder(r.Body).Decode(&stats)

	// Update model_performance_stats
	query := `INSERT OR REPLACE INTO model_performance_stats 
		(id, model, task_type, total_tasks, total_quality, updated_at)
		VALUES (?, ?, ?, 1, ?, ?)`
	h.db.Exec(query, fmt.Sprintf("%s:%s", stats.Model, stats.TaskType), stats.Model, stats.TaskType, stats.Rating, time.Now().Unix())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *Hub) handleNotionSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req notion.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create Notion client
	client := notion.NewNotionClient("") // Token will come from request

	// Call sync
	resp, err := client.Sync(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// MCPServer represents a MCP server
type MCPServer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Enabled     bool   `json:"enabled"`
	Command     string `json:"command"`
	Description string `json:"description"`
}

// MCP control handlers

func (h *Hub) handleEnableMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ServerID string `json:"server_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Enabled MCP server: %s", req.ServerID),
	})
}

func (h *Hub) handleDisableMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ServerID string `json:"server_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Disabled MCP server: %s", req.ServerID),
	})
}

// OpenAI API Bridge handlers
func (h *Hub) handleOpenAIChatCompletions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Model    string              `json:"model"`
		Messages []map[string]string `json:"messages"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Lazy load tools on demand
	tools := h.getToolsForQuery(req.Model)

	response := map[string]interface{}{
		"id":                fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		"object":            "chat.completion",
		"created":           time.Now().Unix(),
		"model":             req.Model,
		"choices":           []map[string]interface{}{},
		"usage":             map[string]int{"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0},
		"tools":             tools,
		"lazy_loaded_tools": true,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Hub) handleOpenAIModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"object": "list",
		"data": []map[string]interface{}{
			{
				"id":      "tibrain-mcp-hub",
				"object":  "model",
				"created": time.Now().Unix(),
				"owned_by": "tibrain",
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Hub) getToolsForQuery(model string) []map[string]interface{} {
	// Return lazy-loaded tools
	return []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "read_file",
				"description": "Read file content",
				"parameters":  map[string]interface{}{"type": "object"},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "list_files",
				"description": "List files in directory",
				"parameters":  map[string]interface{}{"type": "object"},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "execute_tool",
				"description": "Execute MCP tool on-demand",
				"parameters":  map[string]interface{}{"type": "object"},
			},
		},
	}
}

func (h *Hub) handleToolStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"tools": map[string]interface{}{
			"lazy_loading": true,
			"on_demand":    true,
		},
	})
}