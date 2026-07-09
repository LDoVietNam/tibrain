package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ti/router/tibrain/internal/memory"
)

// MCPServerManager manages the embedded MCP server
type MCPServerManager struct {
	mcpServer    *server.MCPServer
	sseServer    *server.SSEServer
	hub          *Hub
	apiBaseURL   string
	mcpHubClient *MCPHubClient
	httpClient   *http.Client
	config       *Config
}

// NewMCPServerManager creates a new MCP server manager
func NewMCPServerManager(hub *Hub, apiBaseURL string, mcpHubClient *MCPHubClient, config *Config) *MCPServerManager {
	name := "TiBrain Embedded MCP Hub"
	if config != nil && config.MCP.PublicName != "" {
		name = config.MCP.PublicName
	}
	s := server.NewMCPServer(
		name,
		"2.0.0",
		server.WithLogging(),
	)

	sse := server.NewSSEServer(s, server.WithStaticBasePath("/mcp"))

	manager := &MCPServerManager{
		mcpServer:    s,
		sseServer:    sse,
		hub:          hub,
		apiBaseURL:   strings.TrimRight(apiBaseURL, "/"),
		mcpHubClient: mcpHubClient,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		config:       config,
	}

	manager.registerTools()
	return manager
}

// registerTools registers all TiBrain functionalities as MCP tools
func (m *MCPServerManager) registerTools() {
	if m.config != nil && m.config.MCP.PublicMode == "unified" {
		m.registerUnifiedPublicTools()
		return
	}
	m.registerBrainTools()
	m.registerControlPlaneTools()
	m.registerRegistryTools()
	m.registerHubTools()
}

func (m *MCPServerManager) registerBrainTools() {
	// 1. Tool: Query Memory
	queryMemoryTool := mcp.NewTool("query_memory",
		mcp.WithDescription("Query the cognitive long term, episodic, semantic, or procedural memory"),
		mcp.WithString("query", mcp.Required(), mcp.Description("The query string to search for in memory")),
		mcp.WithString("type", mcp.Description("Memory type: episodic, semantic, procedural, long_term")),
		mcp.WithNumber("limit", mcp.Description("Maximum entries to retrieve (default: 5)")),
	)
	m.mcpServer.AddTool(queryMemoryTool, m.handleQueryMemory)

	// 2. Tool: Store Episodic Memory
	storeMemoryTool := mcp.NewTool("store_memory",
		mcp.WithDescription("Store a new episodic memory (experience/interaction) into the cognitive memory system"),
		mcp.WithString("content", mcp.Required(), mcp.Description("The knowledge or experience content to store")),
		mcp.WithString("session_id", mcp.Description("Optional session grouping ID")),
	)
	m.mcpServer.AddTool(storeMemoryTool, m.handleStoreMemory)

	// 3. Tool: Chat RAG
	chatRAGTool := mcp.NewTool("chat_rag",
		mcp.WithDescription("Send a chat message to TiBrain with RAG (Retrieval-Augmented Generation) context enabled"),
		mcp.WithString("message", mcp.Required(), mcp.Description("User input message")),
		mcp.WithString("session_id", mcp.Description("Session ID for conversation history tracking")),
	)
	m.mcpServer.AddTool(chatRAGTool, m.handleChatRAG)
}

func (m *MCPServerManager) handleQueryMemory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := request.GetString("query", "")
	memType := request.GetString("type", "")
	limit := request.GetInt("limit", 5)

	if query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	if globalCognitiveMemory == nil {
		return mcp.NewToolResultError("cognitive memory is not initialized"), nil
	}

	entries, err := globalCognitiveMemory.QueryMemory(ctx, query, memory.MemoryType(memType), limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to query memory: %v", err)), nil
	}

	data, _ := json.MarshalIndent(entries, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (m *MCPServerManager) handleStoreMemory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content := request.GetString("content", "")
	sessionID := request.GetString("session_id", "")

	if content == "" {
		return mcp.NewToolResultError("content parameter is required"), nil
	}

	if globalCognitiveMemory == nil {
		return mcp.NewToolResultError("cognitive memory is not initialized"), nil
	}

	memCtx := map[string]interface{}{
		"source": "mcp",
	}
	if sessionID != "" {
		memCtx["session_id"] = sessionID
	}

	id, err := globalCognitiveMemory.StoreEpisodicMemory(ctx, content, memCtx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to store memory: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Episodic memory stored successfully. ID: %s", id)), nil
}

func (m *MCPServerManager) handleChatRAG(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message := request.GetString("message", "")
	sessionID := request.GetString("session_id", "")

	if message == "" {
		return mcp.NewToolResultError("message parameter is required"), nil
	}

	// Recall history
	var conversationHistory []string
	if sessionID != "" && globalCognitiveMemory != nil {
		entries, err := globalCognitiveMemory.GetRecentExperience(ctx, 6)
		if err == nil {
			for _, e := range entries {
				if ctxVal, ok := e.Context["session_id"]; ok && ctxVal == sessionID {
					conversationHistory = append(conversationHistory, fmt.Sprintf("[%s] %s", e.Type, e.Content))
				}
			}
		}
	}

	// Prepare request schema for routing
	var answer string
	var confidence float64
	var routeStr string
	sources := []string{}

	if globalRetrievalRouter != nil {
		decision := globalRetrievalRouter.RouteQuery(ctx, message)
		if decision != nil {
			routeStr = string(decision.Route)
			resp, err := globalRetrievalRouter.ExecuteRoute(ctx, message, decision, nil, nil)
			if err == nil && resp != nil {
				answer = resp.Answer
				confidence = resp.Confidence
				for _, r := range resp.Results {
					sources = append(sources, fmt.Sprintf("- [%s] %s (Score: %.2f)", r.Source, r.Content[:minInt(120, len(r.Content))], r.Score))
				}
			}
		}
	}

	if answer == "" {
		answer = "TiBrain completed RAG retrieval but no generation context was available."
	}

	// Store user query
	if globalCognitiveMemory != nil {
		globalCognitiveMemory.StoreEpisodicMemory(ctx, fmt.Sprintf("user: %s", message), map[string]interface{}{
			"source":     "mcp_chat",
			"session_id": sessionID,
		})
		// Store assistant response
		globalCognitiveMemory.StoreEpisodicMemory(ctx, fmt.Sprintf("assistant: %s", answer), map[string]interface{}{
			"source":     "mcp_chat",
			"session_id": sessionID,
		})
	}

	resultMap := map[string]interface{}{
		"answer":     answer,
		"confidence": confidence,
		"route":      routeStr,
		"sources":    sources,
		"raw_query":  message,
		"timestamp":  time.Now(),
	}

	data, _ := json.MarshalIndent(resultMap, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// HandleSSE SSE event listener
func (m *MCPServerManager) HandleSSE(w http.ResponseWriter, r *http.Request) {
	m.sseServer.SSEHandler().ServeHTTP(w, r)
}

// HandleMessage SSE commands receiver
func (m *MCPServerManager) HandleMessage(w http.ResponseWriter, r *http.Request) {
	m.sseServer.MessageHandler().ServeHTTP(w, r)
}
