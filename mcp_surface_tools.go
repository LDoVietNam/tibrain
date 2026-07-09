package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func (m *MCPServerManager) registerUnifiedPublicTools() {
	tools := []struct {
		tool mcp.Tool
		h    func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	}{
		{mcp.NewTool("brain_health", mcp.WithDescription("Get TiBrain health status")), m.handleBrainHealth},
		{mcp.NewTool("brain_capabilities", mcp.WithDescription("List logical targets and capabilities")), m.handleBrainCapabilities},
		{mcp.NewTool("brain_preflight",
			mcp.WithDescription("Validate target, tool, and connectivity before execution"),
			mcp.WithString("target", mcp.Required(), mcp.Description("Logical target name (e.g. filesystem, router)")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Tool name to check")),
		), m.handleBrainPreflight},
		{mcp.NewTool("brain_tool_search",
			mcp.WithDescription("Search for internal tools by keyword or target"),
			mcp.WithString("keyword", mcp.Description("Optional search keyword")),
			mcp.WithString("target", mcp.Description("Optional target filter")),
		), m.handleBrainToolSearch},
		{mcp.NewTool("brain_execute",
			mcp.WithDescription("Execute an internal tool on a logical target"),
			mcp.WithString("target", mcp.Required(), mcp.Description("Logical target (e.g. filesystem, router, browser, obsidian)")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Full tool name or unqualified tool name")),
			mcp.WithString("args_json", mcp.Description("Optional JSON arguments object string")),
			mcp.WithString("access", mcp.Description("Optional access override: read, write, destructive")),
		), m.handleBrainExecute},
		{mcp.NewTool("brain_batch_execute",
			mcp.WithDescription("Execute multiple unified tool calls sequentially"),
			mcp.WithString("calls_json", mcp.Required(), mcp.Description("JSON array of UnifiedToolCall objects")),
		), m.handleBrainBatchExecute},
		{mcp.NewTool("brain_transaction_begin",
			mcp.WithDescription("Begin a tracked background execution/shell command"),
			mcp.WithString("command", mcp.Required(), mcp.Description("Command line to execute")),
			mcp.WithString("working_dir", mcp.Description("Optional working directory")),
			mcp.WithNumber("timeout_ms", mcp.Description("Optional timeout in milliseconds")),
		), m.handleTransactionBegin},
		{mcp.NewTool("brain_transaction_status",
			mcp.WithDescription("Check status of a tracked background execution"),
			mcp.WithString("transaction_id", mcp.Required(), mcp.Description("Transaction/execution job ID")),
		), m.handleTransactionStatus},
		{mcp.NewTool("brain_transaction_commit",
			mcp.WithDescription("Wait for tracked execution to finish and get final results"),
			mcp.WithString("transaction_id", mcp.Required(), mcp.Description("Transaction/execution job ID")),
		), m.handleTransactionCommit},
		{mcp.NewTool("brain_transaction_rollback",
			mcp.WithDescription("Cancel and terminate a running background execution"),
			mcp.WithString("transaction_id", mcp.Required(), mcp.Description("Transaction/execution job ID")),
		), m.handleTransactionRollback},
	}

	for _, entry := range tools {
		m.mcpServer.AddTool(entry.tool, entry.h)
	}

	// Register compatibility aliases only when explicitly enabled.
	if m.config != nil && m.config.MCP.Compatibility {
		m.registerCompatibilityAliases()
	}
}

func (m *MCPServerManager) registerCompatibilityAliases() {
	// Register legacy aliases so they still function but are separate from unified list
	aliases := []struct {
		tool mcp.Tool
		h    func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	}{
		{mcp.NewTool("hub_call_tool",
			mcp.WithDescription("[Compatibility Alias] Call one upstream MCP tool"),
			mcp.WithString("server", mcp.Required(), mcp.Description("Upstream server")),
			mcp.WithString("tool", mcp.Required(), mcp.Description("Upstream tool")),
			mcp.WithString("args_json", mcp.Description("Args JSON")),
		), m.handleHubCallTool},
		{mcp.NewTool("hub_batch_call",
			mcp.WithDescription("[Compatibility Alias] Call multiple tools"),
			mcp.WithString("calls_json", mcp.Required(), mcp.Description("JSON array of calls")),
		), m.handleHubBatchCall},
		{mcp.NewTool("brain_status", mcp.WithDescription("[Compatibility Alias] Get brain status")), m.handleBrainStatus},
	}
	for _, entry := range aliases {
		m.mcpServer.AddTool(entry.tool, entry.h)
	}
}

func (m *MCPServerManager) registerHubTools() {
	tools := []struct {
		tool mcp.Tool
		h    func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	}{
		{mcp.NewTool("brain_health", mcp.WithDescription("Get TiBrain health from the live HTTP surface")), m.handleBrainHealth},
		{mcp.NewTool("brain_status", mcp.WithDescription("Get TiBrain status from the live HTTP surface")), m.handleBrainStatus},
		{mcp.NewTool("brain_overview", mcp.WithDescription("Get a merged TiBrain overview")), m.handleBrainOverview},
		{mcp.NewTool("brain_knowledge_status", mcp.WithDescription("Get TiBrain knowledge base status")), m.handleBrainKnowledgeStatus},
		{mcp.NewTool("brain_agents", mcp.WithDescription("List TiBrain agents")), m.handleBrainAgents},
		{mcp.NewTool("brain_clis", mcp.WithDescription("List registered CLIs")), m.handleBrainCLIs},
		{mcp.NewTool("brain_mcp_registry", mcp.WithDescription("List registered MCP servers")), m.handleBrainMCPs},
		{mcp.NewTool("brain_tools_registry", mcp.WithDescription("List registered tools")), m.handleBrainToolsRegistry},
		{mcp.NewTool("brain_runtime_registry", mcp.WithDescription("List runtime registry entries")), m.handleBrainRuntimeRegistry},
		{mcp.NewTool("brain_rag_status", mcp.WithDescription("Get RAG status")), m.handleBrainRAGStatus},
		{mcp.NewTool("brain_rag_query",
			mcp.WithDescription("Run a TiBrain RAG query"),
			mcp.WithString("query", mcp.Required(), mcp.Description("Query text")),
			mcp.WithNumber("limit", mcp.Description("Result limit")),
		), m.handleBrainRAGQuery},
		{mcp.NewTool("brain_adaptive_retrieve",
			mcp.WithDescription("Run adaptive retrieval"),
			mcp.WithString("query", mcp.Required(), mcp.Description("Query text")),
			mcp.WithString("scopes_csv", mcp.Description("Optional comma-separated scopes")),
			mcp.WithString("context_source_json", mcp.Description("Optional JSON context source")),
		), m.handleBrainAdaptiveRetrieve},
		{mcp.NewTool("brain_retrieval_traces",
			mcp.WithDescription("List retrieval traces"),
			mcp.WithNumber("limit", mcp.Description("Trace limit")),
		), m.handleBrainRetrievalTraces},
		{mcp.NewTool("brain_pattern_candidates",
			mcp.WithDescription("List pattern candidates"),
			mcp.WithString("status", mcp.Description("Optional status filter")),
			mcp.WithNumber("limit", mcp.Description("Candidate limit")),
		), m.handleBrainPatternCandidates},
		{mcp.NewTool("brain_promote_pattern",
			mcp.WithDescription("Promote a pattern candidate"),
			mcp.WithString("candidate_id", mcp.Required(), mcp.Description("Pattern candidate id")),
		), m.handleBrainPromotePattern},
		{mcp.NewTool("brain_model_stats", mcp.WithDescription("Get model stats")), m.handleBrainModelStats},
		{mcp.NewTool("brain_best_model",
			mcp.WithDescription("Get best model for a task"),
			mcp.WithString("task", mcp.Description("Optional task name")),
		), m.handleBrainBestModel},
		{mcp.NewTool("brain_all_model_stats", mcp.WithDescription("Get all model stats")), m.handleBrainAllModelStats},
		{mcp.NewTool("brain_tool_usage_logs",
			mcp.WithDescription("Get tool usage logs"),
			mcp.WithString("tool_name", mcp.Description("Optional tool name")),
			mcp.WithString("agent_id", mcp.Description("Optional agent id")),
			mcp.WithNumber("limit", mcp.Description("Log limit")),
		), m.handleBrainToolUsageLogs},
		{mcp.NewTool("hub_servers", mcp.WithDescription("List upstream MCP servers from apps/tibrain/mcp")), m.handleHubServers},
		{mcp.NewTool("hub_server_status",
			mcp.WithDescription("Get a single upstream MCP server status"),
			mcp.WithString("server", mcp.Required(), mcp.Description("Upstream server name")),
		), m.handleHubServerStatus},
		{mcp.NewTool("hub_tools", mcp.WithDescription("List upstream MCP tools from apps/tibrain/mcp")), m.handleHubTools},
		{mcp.NewTool("hub_sync_registry", mcp.WithDescription("Sync upstream MCP tools into TiBrain tool registry")), m.handleHubSyncRegistry},
		{mcp.NewTool("hub_call_tool",
			mcp.WithDescription("Call one upstream MCP tool through apps/tibrain/mcp"),
			mcp.WithString("server", mcp.Required(), mcp.Description("Upstream server name")),
			mcp.WithString("tool", mcp.Required(), mcp.Description("Upstream tool name")),
			mcp.WithString("args_json", mcp.Description("Optional JSON args object")),
		), m.handleHubCallTool},
		{mcp.NewTool("hub_batch_call",
			mcp.WithDescription("Call multiple upstream MCP tools through apps/tibrain/mcp"),
			mcp.WithString("calls_json", mcp.Required(), mcp.Description("JSON array of calls")),
		), m.handleHubBatchCall},
		{mcp.NewTool("brain_mcp_api_call",
			mcp.WithDescription("Call any apps/tibrain/mcp REST endpoint through TiBrain"),
			mcp.WithString("method", mcp.Description("HTTP method: GET, POST, PUT, PATCH, DELETE")),
			mcp.WithString("path", mcp.Required(), mcp.Description("REST path such as /api/v1/status")),
			mcp.WithString("body_json", mcp.Description("Optional JSON request body")),
		), m.handleBrainMCPAPICall},
	}

	for _, entry := range tools {
		m.mcpServer.AddTool(entry.tool, entry.h)
	}
}

func (m *MCPServerManager) callLocalJSON(ctx context.Context, method string, paths []string, payload any) ([]byte, string, error) {
	if m.apiBaseURL == "" {
		return nil, "", fmt.Errorf("api base url is not configured")
	}

	var rawBody []byte
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, "", fmt.Errorf("marshal request: %w", err)
		}
		rawBody = raw
	}

	var lastErr error
	for _, path := range paths {
		var bodyReader io.Reader
		if rawBody != nil {
			bodyReader = bytes.NewReader(rawBody)
		}
		req, err := http.NewRequestWithContext(ctx, method, m.apiBaseURL+path, bodyReader)
		if err != nil {
			return nil, "", fmt.Errorf("create request: %w", err)
		}
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := m.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("status %d from %s: %s", resp.StatusCode, path, strings.TrimSpace(string(data)))
			continue
		}

		return data, path, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no path succeeded")
	}
	return nil, "", lastErr
}

func (m *MCPServerManager) callLocalJSONBody(ctx context.Context, method string, path string, payload any) ([]byte, error) {
	data, _, err := m.callLocalJSON(ctx, method, []string{path}, payload)
	return data, err
}

func (m *MCPServerManager) callLocalJSONFallback(ctx context.Context, method string, paths []string, payload any) ([]byte, error) {
	data, _, err := m.callLocalJSON(ctx, method, paths, payload)
	return data, err
}

func (m *MCPServerManager) resultText(data []byte) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(strings.TrimSpace(string(data))), nil
}

func (m *MCPServerManager) handleBrainHealth(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONFallback(ctx, http.MethodGet, []string{"/api/health", "/health"}, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainStatus(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONFallback(ctx, http.MethodGet, []string{"/api/status", "/v1/tibrain/status", "/status"}, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainOverview(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONFallback(ctx, http.MethodGet, []string{"/overview", "/api/status"}, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainKnowledgeStatus(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONFallback(ctx, http.MethodGet, []string{"/api/knowledge/status", "/api/knowledge"}, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainAgents(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/api/agents", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainCLIs(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/list-clis", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainMCPs(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/list-mcps", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainToolsRegistry(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONFallback(ctx, http.MethodGet, []string{"/api/tools", "/list-tools"}, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRuntimeRegistry(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/api/v2/runtime/registry", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRAGStatus(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/api/rag/status", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRAGQuery(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := strings.TrimSpace(request.GetString("query", ""))
	if query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}
	limit := request.GetInt("limit", 5)
	payload := map[string]interface{}{
		"query": query,
		"limit": limit,
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/rag/query", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainAdaptiveRetrieve(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := strings.TrimSpace(request.GetString("query", ""))
	if query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	var scopes []string
	if raw := strings.TrimSpace(request.GetString("scopes_csv", "")); raw != "" {
		for _, item := range strings.Split(raw, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				scopes = append(scopes, item)
			}
		}
	}

	var contextSource any
	if raw := strings.TrimSpace(request.GetString("context_source_json", "")); raw != "" {
		if err := json.Unmarshal([]byte(raw), &contextSource); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid context_source_json: %v", err)), nil
		}
	}

	payload := map[string]interface{}{
		"query": query,
	}
	if len(scopes) > 0 {
		payload["scopes"] = scopes
	}
	if contextSource != nil {
		payload["context_source"] = contextSource
	}

	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/v2/retrieve", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRetrievalTraces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := request.GetInt("limit", 20)
	path := "/api/v2/retrieve/traces?limit=" + strconv.Itoa(limit)
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainPatternCandidates(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status := strings.TrimSpace(request.GetString("status", ""))
	limit := request.GetInt("limit", 20)
	query := url.Values{}
	if status != "" {
		query.Set("status", status)
	}
	query.Set("limit", strconv.Itoa(limit))
	path := "/api/v2/brain/patterns?" + query.Encode()
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainPromotePattern(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	candidateID := strings.TrimSpace(request.GetString("candidate_id", ""))
	if candidateID == "" {
		return mcp.NewToolResultError("candidate_id is required"), nil
	}

	payload := map[string]interface{}{"candidate_id": candidateID}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/api/v2/brain/patterns/promote", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainModelStats(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/get-model-stats", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainBestModel(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	task := strings.TrimSpace(request.GetString("task", ""))
	path := "/get-best-model"
	if task != "" {
		path += "?task=" + url.QueryEscape(task)
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainAllModelStats(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/get-all-model-stats", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainToolUsageLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := url.Values{}
	if toolName := strings.TrimSpace(request.GetString("tool_name", "")); toolName != "" {
		query.Set("tool_name", toolName)
	}
	if agentID := strings.TrimSpace(request.GetString("agent_id", "")); agentID != "" {
		query.Set("agent_id", agentID)
	}
	limit := request.GetInt("limit", 50)
	query.Set("limit", strconv.Itoa(limit))

	path := "/tool-usage-logs?" + query.Encode()
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleHubServers(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if m.mcpHubClient == nil {
		return mcp.NewToolResultError("mcp hub client is not configured"), nil
	}
	servers, err := m.mcpHubClient.ListServers(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(servers, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleHubServerStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if m.mcpHubClient == nil {
		return mcp.NewToolResultError("mcp hub client is not configured"), nil
	}
	serverName := strings.TrimSpace(request.GetString("server", ""))
	if serverName == "" {
		return mcp.NewToolResultError("server parameter is required"), nil
	}
	info, err := m.mcpHubClient.GetServerStatus(ctx, serverName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(info, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleHubTools(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if m.mcpHubClient == nil {
		return mcp.NewToolResultError("mcp hub client is not configured"), nil
	}
	tools, err := m.mcpHubClient.ListTools(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(tools, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleHubSyncRegistry(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if m.mcpHubClient == nil {
		return mcp.NewToolResultError("mcp hub client is not configured"), nil
	}
	tools, err := m.mcpHubClient.ListTools(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := m.mcpHubClient.SyncToRegistry(ctx, m.hub); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("synced %d upstream MCP tools into TiBrain registry", len(tools))), nil
}

func (m *MCPServerManager) handleHubCallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if m.mcpHubClient == nil {
		return mcp.NewToolResultError("mcp hub client is not configured"), nil
	}
	serverName := strings.TrimSpace(request.GetString("server", ""))
	toolName := strings.TrimSpace(request.GetString("tool", ""))
	if serverName == "" || toolName == "" {
		return mcp.NewToolResultError("server and tool parameters are required"), nil
	}

	args := map[string]interface{}{}
	if raw := strings.TrimSpace(request.GetString("args_json", "")); raw != "" {
		if err := json.Unmarshal([]byte(raw), &args); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid args_json: %v", err)), nil
		}
	}

	result, err := m.mcpHubClient.CallTool(ctx, serverName, toolName, args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleHubBatchCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if m.mcpHubClient == nil {
		return mcp.NewToolResultError("mcp hub client is not configured"), nil
	}

	raw := strings.TrimSpace(request.GetString("calls_json", ""))
	if raw == "" {
		return mcp.NewToolResultError("calls_json is required"), nil
	}

	var calls []MCPToolCall
	if err := json.Unmarshal([]byte(raw), &calls); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid calls_json: %v", err)), nil
	}

	results, err := m.mcpHubClient.BatchCallTools(ctx, calls)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(results, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainMCPAPICall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method := strings.ToUpper(strings.TrimSpace(request.GetString("method", http.MethodGet)))
	path := strings.TrimSpace(request.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/mcp") && path != "/health" && path != "/status" {
		return mcp.NewToolResultError("path must target the apps/tibrain/mcp HTTP surface"), nil
	}

	var payload any
	if raw := strings.TrimSpace(request.GetString("body_json", "")); raw != "" {
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid body_json: %v", err)), nil
		}
	}

	switch method {
	case http.MethodGet, http.MethodDelete, http.MethodHead:
		if payload != nil {
			return mcp.NewToolResultError("body_json is not supported for GET/DELETE/HEAD"), nil
		}
	default:
		if payload == nil {
			payload = map[string]interface{}{}
		}
	}

	data, err := m.callLocalJSONBody(ctx, method, path, payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainCapabilities(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targets := globalTargetRegistry.List()
	var totalTools int
	if m.mcpHubClient != nil {
		if tools, err := m.mcpHubClient.ListTools(ctx); err == nil {
			totalTools = len(tools)
		}
	}
	resp := map[string]interface{}{
		"targets":     targets,
		"total_tools": totalTools,
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainPreflight(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetAlias := strings.TrimSpace(request.GetString("target", ""))
	toolName := strings.TrimSpace(request.GetString("name", ""))
	checks := map[string]string{
		"tibrain_local": "fail",
		"hub_servers":    "fail",
		"hub_tools":      "fail",
		"target_lookup":  "fail",
		"filesystem":     "skip",
		"omniroute":      "skip",
		"public_tunnel":  "skip",
		"connector":      "skip",
	}

	if _, err := m.callLocalJSONFallback(ctx, http.MethodGet, []string{"/health", "/api/health"}, nil); err == nil {
		checks["tibrain_local"] = "pass"
	}

	resolved, err := globalTargetRegistry.Resolve(targetAlias)
	if err != nil {
		resp := map[string]interface{}{
			"success": false,
			"checks":  checks,
			"error":   fmt.Sprintf("target resolution failed: %v", err),
		}
		data, _ := json.MarshalIndent(resp, "", "  ")
		return m.resultText(data)
	}
	checks["target_lookup"] = "pass"

	if m.mcpHubClient == nil {
		resp := map[string]interface{}{
			"success": false,
			"checks":  checks,
			"error":   "mcp hub client not configured",
		}
		data, _ := json.MarshalIndent(resp, "", "  ")
		return m.resultText(data)
	}

	if _, err := m.mcpHubClient.ListServers(ctx); err == nil {
		checks["hub_servers"] = "pass"
	}

	tools, err := m.mcpHubClient.ListTools(ctx)
	if err != nil {
		resp := map[string]interface{}{
			"success": false,
			"checks":  checks,
			"error":   fmt.Sprintf("cannot list tools from hub: %v", err),
		}
		data, _ := json.MarshalIndent(resp, "", "  ")
		return m.resultText(data)
	}
	checks["hub_tools"] = "pass"

	found := false
	for _, t := range tools {
		if t.ServerName == resolved.ServerName && (t.Name == toolName || fmt.Sprintf("%s:%s", t.ServerName, t.Name) == toolName) {
			found = true
			break
		}
	}

	if resolved.ServerName == "1mcp-local-bridge" && found {
		checks["filesystem"] = "pass"
	}
	if resolved.ServerName == "omniroute-router" && found {
		checks["omniroute"] = "pass"
	}
	checks["connector"] = "pass"

	resp := map[string]interface{}{
		"success":      checks["tibrain_local"] == "pass" && checks["hub_servers"] == "pass" && checks["hub_tools"] == "pass" && checks["target_lookup"] == "pass" && found,
		"target_alias": targetAlias,
		"server_name":  resolved.ServerName,
		"tool_name":    toolName,
		"resolved":     true,
		"exists":       found,
		"checks":       checks,
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainToolSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	keyword := strings.ToLower(strings.TrimSpace(request.GetString("keyword", "")))
	targetAlias := strings.ToLower(strings.TrimSpace(request.GetString("target", "")))

	if m.mcpHubClient == nil {
		return mcp.NewToolResultError("mcp hub client not configured"), nil
	}

	tools, err := m.mcpHubClient.ListTools(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var targetServer string
	if targetAlias != "" {
		if resolved, err := globalTargetRegistry.Resolve(targetAlias); err == nil {
			targetServer = resolved.ServerName
		}
	}

	var matches []MCPToolInfo
	for _, t := range tools {
		if targetServer != "" && t.ServerName != targetServer {
			continue
		}
		if keyword != "" {
			nameMatch := strings.Contains(strings.ToLower(t.Name), keyword)
			descMatch := strings.Contains(strings.ToLower(t.Description), keyword)
			serverMatch := strings.Contains(strings.ToLower(t.ServerName), keyword)
			if !nameMatch && !descMatch && !serverMatch {
				continue
			}
		}
		matches = append(matches, t)
	}

	data, _ := json.MarshalIndent(matches, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainExecute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	target := strings.TrimSpace(request.GetString("target", ""))
	name := strings.TrimSpace(request.GetString("name", ""))
	argsJSON := strings.TrimSpace(request.GetString("args_json", ""))
	access := strings.TrimSpace(request.GetString("access", ""))

	var args map[string]interface{}
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid args_json: %v", err)), nil
		}
	}

	call := UnifiedToolCall{
		Target:    target,
		Name:      name,
		Arguments: args,
		Access:    access,
	}

	res := m.executeUnifiedTool(ctx, call)
	data, _ := json.MarshalIndent(res, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainBatchExecute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	callsJSON := strings.TrimSpace(request.GetString("calls_json", ""))
	if callsJSON == "" {
		return mcp.NewToolResultError("calls_json is required"), nil
	}

	var calls []UnifiedToolCall
	if err := json.Unmarshal([]byte(callsJSON), &calls); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid calls_json array: %v", err)), nil
	}

	results := make([]*UnifiedToolResult, 0, len(calls))
	for _, call := range calls {
		res := m.executeUnifiedTool(ctx, call)
		results = append(results, res)
	}

	data, _ := json.MarshalIndent(results, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleTransactionBegin(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	command := strings.TrimSpace(request.GetString("command", ""))
	workingDir := strings.TrimSpace(request.GetString("working_dir", ""))
	timeoutMs := request.GetInt("timeout_ms", 0)

	if command == "" {
		return mcp.NewToolResultError("command is required"), nil
	}

	job, err := startTrackedShellCommand(command, workingDir, timeoutMs)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	resp := map[string]interface{}{
		"success":        true,
		"transaction_id": job.ID,
		"status":         "running",
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleTransactionStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	txID := strings.TrimSpace(request.GetString("transaction_id", ""))
	if txID == "" {
		return mcp.NewToolResultError("transaction_id is required"), nil
	}

	snap, ok := globalExecJobStore.snapshot(txID)
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("transaction %q not found", txID)), nil
	}

	data, _ := json.MarshalIndent(snap, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleTransactionCommit(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	txID := strings.TrimSpace(request.GetString("transaction_id", ""))
	if txID == "" {
		return mcp.NewToolResultError("transaction_id is required"), nil
	}

	job, ok := globalExecJobStore.get(txID)
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("transaction %q not found", txID)), nil
	}

	snap, err := waitForTrackedJob(ctx, job)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, _ := json.MarshalIndent(snap, "", "  ")
	return m.resultText(data)
}

func (m *MCPServerManager) handleTransactionRollback(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	txID := strings.TrimSpace(request.GetString("transaction_id", ""))
	if txID == "" {
		return mcp.NewToolResultError("transaction_id is required"), nil
	}

	job, ok := globalExecJobStore.get(txID)
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("transaction %q not found", txID)), nil
	}

	snap := cancelTrackedJob(job)
	data, _ := json.MarshalIndent(snap, "", "  ")
	return m.resultText(data)
}
