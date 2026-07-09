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

func (m *MCPServerManager) registerRegistryTools() {
	tools := []struct {
		tool mcp.Tool
		h    func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	}{
		{mcp.NewTool("brain_list_clis", mcp.WithDescription("List registered CLIs")), m.handleBrainListCLIs},
		{mcp.NewTool("brain_register_cli",
			mcp.WithDescription("Register a CLI"),
			mcp.WithString("cli_id", mcp.Required(), mcp.Description("CLI id")),
			mcp.WithString("cli_name", mcp.Required(), mcp.Description("CLI name")),
			mcp.WithString("brain_url", mcp.Required(), mcp.Description("Brain URL")),
		), m.handleBrainRegisterCLI},
		{mcp.NewTool("brain_unregister_cli",
			mcp.WithDescription("Unregister a CLI"),
			mcp.WithString("cli_id", mcp.Required(), mcp.Description("CLI id")),
		), m.handleBrainUnregisterCLI},
		{mcp.NewTool("brain_heartbeat_cli",
			mcp.WithDescription("Send a heartbeat for a CLI"),
			mcp.WithString("cli_id", mcp.Required(), mcp.Description("CLI id")),
		), m.handleBrainHeartbeatCLI},
		{mcp.NewTool("brain_create_handoff",
			mcp.WithDescription("Create a global handoff"),
			mcp.WithString("from_cli", mcp.Required(), mcp.Description("Source CLI")),
			mcp.WithString("to_cli", mcp.Required(), mcp.Description("Destination CLI")),
			mcp.WithString("context", mcp.Description("Context")),
			mcp.WithString("output", mcp.Description("Output")),
			mcp.WithString("metadata_json", mcp.Description("Metadata JSON")),
		), m.handleBrainCreateHandoff},
		{mcp.NewTool("brain_recall_handoff",
			mcp.WithDescription("Recall a global handoff"),
			mcp.WithString("from", mcp.Required(), mcp.Description("From CLI")),
			mcp.WithString("to", mcp.Required(), mcp.Description("To CLI")),
		), m.handleBrainRecallHandoff},
		{mcp.NewTool("brain_list_handoffs",
			mcp.WithDescription("List global handoffs"),
			mcp.WithString("cli", mcp.Description("Optional CLI id filter")),
		), m.handleBrainListHandoffs},
		{mcp.NewTool("brain_register_mcp",
			mcp.WithDescription("Register an MCP server"),
			mcp.WithString("id", mcp.Required(), mcp.Description("MCP id")),
			mcp.WithString("name", mcp.Required(), mcp.Description("MCP name")),
			mcp.WithString("type", mcp.Required(), mcp.Description("MCP type")),
			mcp.WithString("command", mcp.Required(), mcp.Description("Command")),
			mcp.WithString("args", mcp.Description("Args JSON string")),
			mcp.WithString("env", mcp.Description("Env JSON string")),
			mcp.WithString("description", mcp.Description("Description")),
			mcp.WithBoolean("enabled", mcp.Description("Enabled")),
			mcp.WithString("cli_id", mcp.Description("CLI id")),
		), m.handleBrainRegisterMCP},
		{mcp.NewTool("brain_list_mcp",
			mcp.WithDescription("List MCP registry entries"),
			mcp.WithString("cli_id", mcp.Description("Optional CLI id filter")),
		), m.handleBrainListMCP},
		{mcp.NewTool("brain_get_mcp",
			mcp.WithDescription("Get a single MCP entry"),
			mcp.WithString("id", mcp.Required(), mcp.Description("MCP id")),
		), m.handleBrainGetMCP},
		{mcp.NewTool("brain_sync_mcp_tools", mcp.WithDescription("Sync MCP tools to TiBrain registry")), m.handleBrainSyncMCPTools},
		{mcp.NewTool("brain_register_tool",
			mcp.WithDescription("Register a tool in TiBrain"),
			mcp.WithString("tool_json", mcp.Required(), mcp.Description("Tool JSON")),
		), m.handleBrainRegisterTool},
		{mcp.NewTool("brain_list_tools_registry",
			mcp.WithDescription("List tools in TiBrain registry"),
			mcp.WithString("category", mcp.Description("Optional category filter")),
		), m.handleBrainListToolsRegistry},
		{mcp.NewTool("brain_get_tool",
			mcp.WithDescription("Get a single tool by id"),
			mcp.WithString("id", mcp.Required(), mcp.Description("Tool id")),
		), m.handleBrainGetTool},
		{mcp.NewTool("brain_execute_tool",
			mcp.WithDescription("Execute a TiBrain or MCP tool through the registry"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Tool name")),
			mcp.WithString("params_json", mcp.Description("Params JSON")),
			mcp.WithString("agent_id", mcp.Description("Agent id")),
		), m.handleBrainExecuteTool},
		{mcp.NewTool("brain_update_model_stats",
			mcp.WithDescription("Update model performance stats"),
			mcp.WithString("stats_json", mcp.Required(), mcp.Description("ModelPerformanceStats JSON")),
		), m.handleBrainUpdateModelStats},
		{mcp.NewTool("brain_get_model_stats",
			mcp.WithDescription("Get model performance stats"),
			mcp.WithString("model", mcp.Required(), mcp.Description("Model")),
			mcp.WithString("task_type", mcp.Required(), mcp.Description("Task type")),
		), m.handleBrainGetModelStats},
		{mcp.NewTool("brain_get_best_model",
			mcp.WithDescription("Get the best model for a task type"),
			mcp.WithString("task_type", mcp.Required(), mcp.Description("Task type")),
		), m.handleBrainGetBestModel},
		{mcp.NewTool("brain_get_quality_trend",
			mcp.WithDescription("Get quality trend for a model"),
			mcp.WithString("model", mcp.Required(), mcp.Description("Model")),
		), m.handleBrainGetQualityTrend},
		{mcp.NewTool("brain_get_all_model_stats", mcp.WithDescription("Get all model stats")), m.handleBrainGetAllModelStats},
		{mcp.NewTool("brain_tool_usage_logs",
			mcp.WithDescription("Get tool usage logs"),
			mcp.WithString("tool_name", mcp.Description("Tool name")),
			mcp.WithString("agent_id", mcp.Description("Agent id")),
			mcp.WithNumber("limit", mcp.Description("Limit")),
		), m.handleBrainToolUsageLogsRegistry},
		{mcp.NewTool("brain_mcp_proxy_servers", mcp.WithDescription("List upstream servers from the MCP proxy")), m.handleBrainMCPProxyServers},
		{mcp.NewTool("brain_mcp_proxy_tools", mcp.WithDescription("List upstream tools from the MCP proxy")), m.handleBrainMCPProxyTools},
		{mcp.NewTool("brain_mcp_proxy_call",
			mcp.WithDescription("Call one MCP proxy tool"),
			mcp.WithString("server", mcp.Required(), mcp.Description("Server name")),
			mcp.WithString("tool", mcp.Required(), mcp.Description("Tool name")),
			mcp.WithString("args_json", mcp.Description("Args JSON")),
		), m.handleBrainMCPProxyCall},
		{mcp.NewTool("brain_mcp_proxy_batch",
			mcp.WithDescription("Call many MCP proxy tools"),
			mcp.WithString("calls_json", mcp.Required(), mcp.Description("Calls JSON array")),
		), m.handleBrainMCPProxyBatch},
		{mcp.NewTool("brain_sync_mcp_hub", mcp.WithDescription("Sync MCP proxy tools into TiBrain registry")), m.handleBrainSyncMCPHub},
	}

	for _, entry := range tools {
		m.mcpServer.AddTool(entry.tool, entry.h)
	}
}

func (m *MCPServerManager) handleBrainListCLIs(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/list-clis", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRegisterCLI(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload := map[string]interface{}{
		"cli_id":    strings.TrimSpace(request.GetString("cli_id", "")),
		"cli_name":  strings.TrimSpace(request.GetString("cli_name", "")),
		"brain_url": strings.TrimSpace(request.GetString("brain_url", "")),
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/register-cli", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainUnregisterCLI(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := "/unregister-cli?cli_id=" + url.QueryEscape(strings.TrimSpace(request.GetString("cli_id", "")))
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainHeartbeatCLI(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := "/heartbeat?cli_id=" + url.QueryEscape(strings.TrimSpace(request.GetString("cli_id", "")))
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainCreateHandoff(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload := map[string]interface{}{
		"from_cli": strings.TrimSpace(request.GetString("from_cli", "")),
		"to_cli":   strings.TrimSpace(request.GetString("to_cli", "")),
		"context":  strings.TrimSpace(request.GetString("context", "")),
		"output":   strings.TrimSpace(request.GetString("output", "")),
	}
	if raw := strings.TrimSpace(request.GetString("metadata_json", "")); raw != "" {
		var metadata map[string]string
		if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid metadata_json: %v", err)), nil
		}
		payload["metadata"] = metadata
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/create-handoff", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRecallHandoff(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := "/recall-handoff?from=" + url.QueryEscape(strings.TrimSpace(request.GetString("from", ""))) + "&to=" + url.QueryEscape(strings.TrimSpace(request.GetString("to", "")))
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainListHandoffs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := "/list-handoffs?cli=" + url.QueryEscape(strings.TrimSpace(request.GetString("cli", "")))
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRegisterMCP(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload := map[string]interface{}{
		"id":          strings.TrimSpace(request.GetString("id", "")),
		"name":        strings.TrimSpace(request.GetString("name", "")),
		"type":        strings.TrimSpace(request.GetString("type", "")),
		"command":     strings.TrimSpace(request.GetString("command", "")),
		"args":        strings.TrimSpace(request.GetString("args", "")),
		"env":         strings.TrimSpace(request.GetString("env", "")),
		"description": strings.TrimSpace(request.GetString("description", "")),
		"enabled":     request.GetBool("enabled", true),
		"cli_id":      strings.TrimSpace(request.GetString("cli_id", "")),
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/register-mcp", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainListMCP(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := "/list-mcps?cli_id=" + url.QueryEscape(strings.TrimSpace(request.GetString("cli_id", "")))
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainGetMCP(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := "/get-mcp?id=" + url.QueryEscape(strings.TrimSpace(request.GetString("id", "")))
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainSyncMCPTools(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/sync-mcp-tools", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainRegisterTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw := strings.TrimSpace(request.GetString("tool_json", ""))
	if raw == "" {
		return mcp.NewToolResultError("tool_json is required"), nil
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/register-tool", json.RawMessage(raw))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainListToolsRegistry(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := "/list-tools?category=" + url.QueryEscape(strings.TrimSpace(request.GetString("category", "")))
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainGetTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := "/get-tool?id=" + url.QueryEscape(strings.TrimSpace(request.GetString("id", "")))
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainExecuteTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params map[string]interface{}
	if raw := strings.TrimSpace(request.GetString("params_json", "")); raw != "" {
		if err := json.Unmarshal([]byte(raw), &params); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid params_json: %v", err)), nil
		}
	}
	if params == nil {
		params = map[string]interface{}{}
	}
	payload := map[string]interface{}{
		"name":     strings.TrimSpace(request.GetString("name", "")),
		"params":   params,
		"agent_id": strings.TrimSpace(request.GetString("agent_id", "")),
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/execute-tool", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainUpdateModelStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw := strings.TrimSpace(request.GetString("stats_json", ""))
	if raw == "" {
		return mcp.NewToolResultError("stats_json is required"), nil
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/update-model-stats", json.RawMessage(raw))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainGetModelStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := "/get-model-stats?model=" + url.QueryEscape(strings.TrimSpace(request.GetString("model", ""))) + "&task_type=" + url.QueryEscape(strings.TrimSpace(request.GetString("task_type", "")))
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainGetBestModel(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := "/get-best-model?task_type=" + url.QueryEscape(strings.TrimSpace(request.GetString("task_type", "")))
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainGetQualityTrend(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := "/get-quality-trend?model=" + url.QueryEscape(strings.TrimSpace(request.GetString("model", "")))
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainGetAllModelStats(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/get-all-model-stats", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainToolUsageLogsRegistry(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := url.Values{}
	if v := strings.TrimSpace(request.GetString("tool_name", "")); v != "" {
		query.Set("tool_name", v)
	}
	if v := strings.TrimSpace(request.GetString("agent_id", "")); v != "" {
		query.Set("agent_id", v)
	}
	if v := request.GetInt("limit", 0); v > 0 {
		query.Set("limit", strconv.Itoa(v))
	}
	path := "/tool-usage-logs"
	if query.Encode() != "" {
		path += "?" + query.Encode()
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, path, nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainMCPProxyServers(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/mcp-hub/servers", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainMCPProxyTools(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/mcp-hub/tools", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainMCPProxyCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args map[string]interface{}
	if raw := strings.TrimSpace(request.GetString("args_json", "")); raw != "" {
		if err := json.Unmarshal([]byte(raw), &args); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid args_json: %v", err)), nil
		}
	}
	if args == nil {
		args = map[string]interface{}{}
	}
	payload := map[string]interface{}{
		"server": strings.TrimSpace(request.GetString("server", "")),
		"tool":   strings.TrimSpace(request.GetString("tool", "")),
		"args":   args,
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/mcp-hub/call", payload)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainMCPProxyBatch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	raw := strings.TrimSpace(request.GetString("calls_json", ""))
	if raw == "" {
		return mcp.NewToolResultError("calls_json is required"), nil
	}
	data, err := m.callLocalJSONBody(ctx, http.MethodPost, "/mcp-hub/batch-call", json.RawMessage(raw))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}

func (m *MCPServerManager) handleBrainSyncMCPHub(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := m.callLocalJSONBody(ctx, http.MethodGet, "/sync-mcp-hub", nil)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return m.resultText(data)
}
