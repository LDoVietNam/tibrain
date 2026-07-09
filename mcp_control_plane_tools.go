package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func (m *MCPServerManager) registerControlPlaneTools() {
	tools := []struct {
		tool mcp.Tool
		h    func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	}{
		{mcp.NewTool("brain_knowledge_index",
			mcp.WithDescription("Index knowledge sources into TiBrain"),
			mcp.WithString("sources_json", mcp.Description("JSON array of knowledge sources")),
			mcp.WithBoolean("force", mcp.Description("Force reindex")),
			mcp.WithNumber("max_file_bytes", mcp.Description("Max file bytes")),
		), m.handleBrainKnowledgeIndex},
		{mcp.NewTool("brain_register_agent",
			mcp.WithDescription("Register an agent in TiBrain"),
			mcp.WithString("id", mcp.Required(), mcp.Description("Agent id")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Agent name")),
			mcp.WithString("type", mcp.Required(), mcp.Description("Agent type")),
			mcp.WithString("endpoint", mcp.Required(), mcp.Description("Agent endpoint")),
		), m.handleBrainRegisterAgent},
		{mcp.NewTool("brain_orchestrate",
			mcp.WithDescription("Run orchestration query"),
			mcp.WithString("query", mcp.Required(), mcp.Description("Orchestration query")),
			mcp.WithNumber("limit", mcp.Description("Limit")),
		), m.handleBrainOrchestrate},
		{mcp.NewTool("brain_agent_request",
			mcp.WithDescription("Send an agent request through the orchestrator"),
			mcp.WithString("request_json", mcp.Required(), mcp.Description("AgentRequest JSON")),
		), m.handleBrainAgentRequest},
		{mcp.NewTool("brain_agent_memory",
			mcp.WithDescription("Get agent memory"),
			mcp.WithString("agent_id", mcp.Required(), mcp.Description("Agent id")),
			mcp.WithString("type", mcp.Description("Memory type")),
		), m.handleBrainAgentMemory},
		{mcp.NewTool("brain_feedback",
			mcp.WithDescription("Submit RAG feedback"),
			mcp.WithString("feedback_json", mcp.Required(), mcp.Description("Feedback JSON")),
		), m.handleBrainFeedback},
		{mcp.NewTool("brain_ingest",
			mcp.WithDescription("Ingest a knowledge directory into TiBrain"),
			mcp.WithString("directory", mcp.Required(), mcp.Description("Directory path")),
			mcp.WithString("category", mcp.Description("Category")),
		), m.handleBrainIngest},
		{mcp.NewTool("brain_router_brain",
			mcp.WithDescription("Query the router-brain endpoint"),
			mcp.WithString("query", mcp.Description("Query text")),
			mcp.WithNumber("limit", mcp.Description("Limit")),
		), m.handleBrainRouterBrain},
		{mcp.NewTool("brain_ti_brain", mcp.WithDescription("Get merged TiBrain info")), m.handleBrainTiBrain},
		{mcp.NewTool("brain_docs_search",
			mcp.WithDescription("Search MCP docs"),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
			mcp.WithString("library", mcp.Description("Library")),
			mcp.WithString("version", mcp.Description("Version")),
			mcp.WithNumber("max_results", mcp.Description("Maximum results")),
		), m.handleBrainDocsSearch},
		{mcp.NewTool("brain_docs_status", mcp.WithDescription("Get docs bridge status")), m.handleBrainDocsStatus},
		{mcp.NewTool("brain_obsidian_status", mcp.WithDescription("Get Obsidian vault status")), m.handleBrainObsidianStatus},
		{mcp.NewTool("brain_rtk_gain",
			mcp.WithDescription("Run RTK gain analysis"),
			mcp.WithString("input_json", mcp.Required(), mcp.Description("Input JSON")),
		), m.handleBrainRTKGain},
		{mcp.NewTool("brain_rtk_discover",
			mcp.WithDescription("Run RTK discover"),
			mcp.WithString("input_json", mcp.Required(), mcp.Description("Input JSON")),
		), m.handleBrainRTKDiscover},
		{mcp.NewTool("brain_rtk_compress",
			mcp.WithDescription("Run RTK compression"),
			mcp.WithString("input_json", mcp.Required(), mcp.Description("Input JSON")),
		), m.handleBrainRTKCompress},
		{mcp.NewTool("brain_rtk_compress_messages",
			mcp.WithDescription("Compress RTK messages"),
			mcp.WithString("input_json", mcp.Required(), mcp.Description("Input JSON")),
		), m.handleBrainRTKCompressMessages},
		{mcp.NewTool("brain_rtk_log",
			mcp.WithDescription("Log RTK event"),
			mcp.WithString("input_json", mcp.Required(), mcp.Description("Input JSON")),
		), m.handleBrainRTKLog},
		{mcp.NewTool("brain_rtk_rules", mcp.WithDescription("Get or update RTK rules")), m.handleBrainRTKRules},
		{mcp.NewTool("brain_rtk_rules_sync", mcp.WithDescription("Sync RTK rules")), m.handleBrainRTKRulesSync},
		{mcp.NewTool("brain_browser_register",
			mcp.WithDescription("Register a browser-runtime extension"),
			mcp.WithString("input_json", mcp.Required(), mcp.Description("Registration JSON")),
		), m.handleBrainBrowserRegister},
		{mcp.NewTool("brain_browser_tasks",
			mcp.WithDescription("Inspect browser-runtime tasks"),
			mcp.WithString("action", mcp.Description("Action: list, claim, complete, fail")),
			mcp.WithString("input_json", mcp.Description("Optional JSON body")),
		), m.handleBrainBrowserTasks},
		{mcp.NewTool("brain_v1_rag_query",
			mcp.WithDescription("Open-WebUI RAG query"),
			mcp.WithString("query", mcp.Required(), mcp.Description("Query text")),
			mcp.WithString("context", mcp.Description("Context")),
			mcp.WithNumber("limit", mcp.Description("Limit")),
		), m.handleBrainV1RagQuery},
		{mcp.NewTool("brain_v1_knowledge_documents",
			mcp.WithDescription("List knowledge documents"),
			mcp.WithString("category", mcp.Description("Category")),
			mcp.WithString("q", mcp.Description("Search query")),
			mcp.WithString("status", mcp.Description("Status")),
			mcp.WithNumber("limit", mcp.Description("Limit")),
			mcp.WithNumber("offset", mcp.Description("Offset")),
		), m.handleBrainV1KnowledgeDocuments},
		{mcp.NewTool("brain_v1_knowledge_ingest",
			mcp.WithDescription("Ingest knowledge content"),
			mcp.WithString("content", mcp.Description("Content")),
			mcp.WithString("category", mcp.Description("Category")),
			mcp.WithString("source", mcp.Description("Source")),
			mcp.WithString("file_path", mcp.Description("File path")),
			mcp.WithString("repo_path", mcp.Description("Repo path")),
		), m.handleBrainV1KnowledgeIngest},
		{mcp.NewTool("brain_v1_graph_nodes",
			mcp.WithDescription("List graph nodes"),
			mcp.WithString("label", mcp.Description("Node label")),
			mcp.WithNumber("limit", mcp.Description("Limit")),
			mcp.WithNumber("offset", mcp.Description("Offset")),
		), m.handleBrainV1GraphNodes},
		{mcp.NewTool("brain_v1_router_routes",
			mcp.WithDescription("List router routes"),
			mcp.WithString("query", mcp.Description("Query")),
			mcp.WithString("history", mcp.Description("History")),
		), m.handleBrainV1RouterRoutes},
		{mcp.NewTool("brain_v1_memory_store",
			mcp.WithDescription("Store a memory entry"),
			mcp.WithString("key", mcp.Required(), mcp.Description("Memory key")),
			mcp.WithString("type", mcp.Description("Memory type")),
			mcp.WithNumber("ttl", mcp.Description("TTL")),
			mcp.WithString("value_json", mcp.Required(), mcp.Description("JSON value")),
			mcp.WithString("metadata_json", mcp.Description("Metadata JSON")),
		), m.handleBrainV1MemoryStore},
		{mcp.NewTool("brain_v1_analytics",
			mcp.WithDescription("Get analytics snapshot"),
			mcp.WithString("period", mcp.Description("Period")),
			mcp.WithString("metric", mcp.Description("Metric")),
		), m.handleBrainV1Analytics},
	}

	for _, entry := range tools {
		m.mcpServer.AddTool(entry.tool, entry.h)
	}
}

func (m *MCPServerManager) handleBrainKnowledgeIndex(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sources []KnowledgeIndexSource
	if raw := strings.TrimSpace(request.GetString("sources_json", "")); raw != "" {
		if err := json.Unmarshal([]byte(raw), &sources); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid sources_json: %v", err)), nil
		}
	}
	payload := map[string]interface{}{
		"sources":        sources,
		"force":          request.GetBool("force", false),
		"max_file_bytes": int64(request.GetInt("max_file_bytes", 0)),
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/knowledge/index", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRegisterAgent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload := map[string]interface{}{
		"id":       strings.TrimSpace(request.GetString("id", "")),
		"name":     strings.TrimSpace(request.GetString("name", "")),
		"type":     strings.TrimSpace(request.GetString("type", "")),
		"endpoint": strings.TrimSpace(request.GetString("endpoint", "")),
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/agents/register", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainOrchestrate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload := map[string]interface{}{
		"query": strings.TrimSpace(request.GetString("query", "")),
		"limit": request.GetInt("limit", 0),
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/orchestrate", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainAgentRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw := strings.TrimSpace(request.GetString("request_json", ""))
	if raw == "" {
		return mcp.NewToolResultError("request_json is required"), nil
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/agent/request", json.RawMessage(raw))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainAgentMemory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := "?agent_id=" + url.QueryEscape(strings.TrimSpace(request.GetString("agent_id", "")))
	if t := strings.TrimSpace(request.GetString("type", "")); t != "" {
		query += "&type=" + url.QueryEscape(t)
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/api/agent/memory"+query, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainFeedback(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw := strings.TrimSpace(request.GetString("feedback_json", ""))
	if raw == "" {
		return mcp.NewToolResultError("feedback_json is required"), nil
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/rag/feedback", json.RawMessage(raw))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainIngest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload := map[string]interface{}{
		"directory": strings.TrimSpace(request.GetString("directory", "")),
		"category":  strings.TrimSpace(request.GetString("category", "")),
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/rag/ingest", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRouterBrain(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := url.Values{}
	if q := strings.TrimSpace(request.GetString("query", "")); q != "" {
		query.Set("query", q)
	}
	if limit := request.GetInt("limit", 0); limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	path := "/api/brain/router"
	if query.Encode() != "" {
		path += "?" + query.Encode()
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainTiBrain(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/api/brain/ti", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainDocsSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload := map[string]interface{}{
		"query":       strings.TrimSpace(request.GetString("query", "")),
		"library":     strings.TrimSpace(request.GetString("library", "")),
		"version":     strings.TrimSpace(request.GetString("version", "")),
		"max_results": request.GetInt("max_results", 0),
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/docs/search", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainDocsStatus(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/api/docs/status", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainObsidianStatus(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/api/obsidian", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRTKGain(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/rtk/gain", json.RawMessage(strings.TrimSpace(request.GetString("input_json", ""))))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRTKDiscover(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/rtk/discover", json.RawMessage(strings.TrimSpace(request.GetString("input_json", ""))))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRTKCompress(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/rtk/compress", json.RawMessage(strings.TrimSpace(request.GetString("input_json", ""))))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRTKCompressMessages(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/rtk/compress-messages", json.RawMessage(strings.TrimSpace(request.GetString("input_json", ""))))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRTKLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/rtk/log", json.RawMessage(strings.TrimSpace(request.GetString("input_json", ""))))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRTKRules(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method := http.MethodGet
	raw := strings.TrimSpace(request.GetString("input_json", ""))
	if raw != "" {
		method = http.MethodPost
	}
	data, err := m.callLocalJSONBody(ctx, method, "/api/rtk/rules", json.RawMessage(raw))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRTKRulesSync(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/rtk/rules/sync", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainBrowserRegister(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw := strings.TrimSpace(request.GetString("input_json", ""))
	if raw == "" {
		return mcp.NewToolResultError("input_json is required"), nil
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/browser-runtime/register", json.RawMessage(raw))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainBrowserTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := strings.TrimSpace(request.GetString("action", "list"))
	path := "/api/browser-runtime/tasks?action=" + url.QueryEscape(action)
	raw := strings.TrimSpace(request.GetString("input_json", ""))
	var payload any
	if raw != "" {
		payload = json.RawMessage(raw)
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, path, payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainV1RagQuery(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload := map[string]interface{}{
		"query":   strings.TrimSpace(request.GetString("query", "")),
		"context": strings.TrimSpace(request.GetString("context", "")),
		"limit":   request.GetInt("limit", 0),
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/v1/rag/query", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainV1KnowledgeDocuments(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := url.Values{}
	if v := strings.TrimSpace(request.GetString("category", "")); v != "" {
		query.Set("category", v)
	}
	if v := strings.TrimSpace(request.GetString("q", "")); v != "" {
		query.Set("q", v)
	}
	if v := strings.TrimSpace(request.GetString("status", "")); v != "" {
		query.Set("status", v)
	}
	if v := request.GetInt("limit", 0); v > 0 {
		query.Set("limit", strconv.Itoa(v))
	}
	if v := request.GetInt("offset", 0); v > 0 {
		query.Set("offset", strconv.Itoa(v))
	}
	path := "/api/v1/knowledge/documents"
	if query.Encode() != "" {
		path += "?" + query.Encode()
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainV1KnowledgeIngest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload := map[string]interface{}{
		"content":   strings.TrimSpace(request.GetString("content", "")),
		"category":  strings.TrimSpace(request.GetString("category", "")),
		"source":    strings.TrimSpace(request.GetString("source", "")),
		"file_path": strings.TrimSpace(request.GetString("file_path", "")),
		"repo_path": strings.TrimSpace(request.GetString("repo_path", "")),
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/v1/knowledge/ingest", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainV1GraphNodes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := url.Values{}
	if v := strings.TrimSpace(request.GetString("label", "")); v != "" {
		query.Set("label", v)
	}
	if v := request.GetInt("limit", 0); v > 0 {
		query.Set("limit", strconv.Itoa(v))
	}
	if v := request.GetInt("offset", 0); v > 0 {
		query.Set("offset", strconv.Itoa(v))
	}
	path := "/api/v1/graph/nodes"
	if query.Encode() != "" {
		path += "?" + query.Encode()
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainV1RouterRoutes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := url.Values{}
	if v := strings.TrimSpace(request.GetString("query", "")); v != "" {
		query.Set("query", v)
	}
	if v := strings.TrimSpace(request.GetString("history", "")); v != "" {
		query.Set("history", v)
	}
	path := "/api/v1/router/routes"
	if query.Encode() != "" {
		path += "?" + query.Encode()
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainV1MemoryStore(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var value any
	rawValue := strings.TrimSpace(request.GetString("value_json", ""))
	if rawValue == "" {
		return mcp.NewToolResultError("value_json is required"), nil
	}
	if err := json.Unmarshal([]byte(rawValue), &value); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid value_json: %v", err)), nil
	}
	payload := map[string]interface{}{
		"key":      strings.TrimSpace(request.GetString("key", "")),
		"value":    value,
		"type":     strings.TrimSpace(request.GetString("type", "")),
		"ttl":      request.GetInt("ttl", 0),
		"metadata": map[string]interface{}{},
	}
	if rawMeta := strings.TrimSpace(request.GetString("metadata_json", "")); rawMeta != "" {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(rawMeta), &meta); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid metadata_json: %v", err)), nil
		}
		payload["metadata"] = meta
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/v1/memory/store", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainV1Analytics(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := url.Values{}
	if v := strings.TrimSpace(request.GetString("period", "")); v != "" {
		query.Set("period", v)
	}
	if v := strings.TrimSpace(request.GetString("metric", "")); v != "" {
		query.Set("metric", v)
	}
	path := "/api/v1/analytics"
	if query.Encode() != "" {
		path += "?" + query.Encode()
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}
