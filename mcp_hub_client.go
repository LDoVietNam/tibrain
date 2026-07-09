// MCP Hub Client - Manage MCP servers from apps/tibrain/mcp
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// MCPHubClient manages the connection to the MCP proxy (apps/tibrain/mcp).
type MCPHubClient struct {
	hubURL       string
	apiKey       string
	apiKeySource string
	httpClient   *http.Client
}

// MCPServerInfo is a lightweight upstream server view from apps/tibrain/mcp.
type MCPServerInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Enabled   bool   `json:"enabled"`
	Connected bool   `json:"connected"`
	ToolCount int    `json:"tool_count"`
}

// MCPToolInfo is a lightweight tool view from apps/tibrain/mcp.
type MCPToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ServerName  string `json:"server_name"`
}

// NewMCPHubClient creates a client for the MCP proxy.
func NewMCPHubClient(hubURL, apiKey string) *MCPHubClient {
	resolvedKey := strings.TrimSpace(apiKey)
	resolvedSource := "constructor"
	if resolvedKey == "" {
		if value := strings.TrimSpace(os.Getenv("MCP_HUB_API_KEY")); value != "" {
			resolvedKey = value
			resolvedSource = "MCP_HUB_API_KEY"
		}
	}
	if resolvedKey == "" {
		if value := strings.TrimSpace(os.Getenv("MCPPROXY_API_KEY")); value != "" {
			resolvedKey = value
			resolvedSource = "MCPPROXY_API_KEY"
		}
	}
	if resolvedKey == "" {
		if value := resolveMCPHubKeyFromFiles(); value != "" {
			resolvedKey = value
			resolvedSource = "file fallback"
		}
	}
	return &MCPHubClient{
		hubURL:       hubURL,
		apiKey:       resolvedKey,
		apiKeySource: resolvedSource,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

var errMCPHubAPIKeyMissing = errors.New("MCP hub API key is not configured")

func (c *MCPHubClient) requireAPIKey() (string, error) {
	key := strings.TrimSpace(c.apiKey)
	if key == "" {
		return "", fmt.Errorf("%w; set MCP_HUB_API_KEY, MCPPROXY_API_KEY, MCP_HUB_API_KEY_FILE, MCP_HUB_CONFIG, MCPPROXY_CONFIG, or a supported runtime config file", errMCPHubAPIKeyMissing)
	}
	return key, nil
}

func (c *MCPHubClient) applyAPIKey(req *http.Request) error {
	key, err := c.requireAPIKey()
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", key)
	return nil
}

func readMCPHubErrorResponse(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(resp.StatusCode)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("MCP hub auth failed (%d): %s", resp.StatusCode, message)
	}
	return fmt.Errorf("MCP hub error %d: %s", resp.StatusCode, message)
}

// ListServers returns the upstream server list from apps/tibrain/mcp.
func (c *MCPHubClient) ListServers(ctx context.Context) ([]MCPServerInfo, error) {
	if _, err := c.requireAPIKey(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.hubURL+"/api/v1/servers", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.applyAPIKey(req); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call MCP hub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readMCPHubErrorResponse(resp)
	}

	var apiResp struct {
		Success bool `json:"success"`
		Data    struct {
			Servers []MCPServerInfo `json:"servers"`
		} `json:"data"`
		Error     string `json:"error"`
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !apiResp.Success {
		return nil, fmt.Errorf("MCP hub error: %s", apiResp.Error)
	}

	return apiResp.Data.Servers, nil
}

// ListTools returns the global tool list from apps/tibrain/mcp.
func (c *MCPHubClient) ListTools(ctx context.Context) ([]MCPToolInfo, error) {
	if _, err := c.requireAPIKey(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.hubURL+"/api/v1/tools", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.applyAPIKey(req); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call MCP hub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readMCPHubErrorResponse(resp)
	}

	var apiResp struct {
		Success bool `json:"success"`
		Data    struct {
			Tools []MCPToolInfo `json:"tools"`
		} `json:"data"`
		Error     string `json:"error"`
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !apiResp.Success {
		return nil, fmt.Errorf("MCP hub error: %s", apiResp.Error)
	}

	return apiResp.Data.Tools, nil
}

type MCPToolAccess string

const (
	MCPToolAccessRead        MCPToolAccess = "read"
	MCPToolAccessWrite       MCPToolAccess = "write"
	MCPToolAccessDestructive MCPToolAccess = "destructive"
)

func wrapperToolForAccess(access MCPToolAccess) (string, error) {
	switch access {
	case MCPToolAccessRead:
		return "call_tool_read", nil
	case MCPToolAccessWrite:
		return "call_tool_write", nil
	case MCPToolAccessDestructive:
		return "call_tool_destructive", nil
	default:
		return "", fmt.Errorf("unsupported MCP tool access policy %q", access)
	}
}

func qualifyMCPToolName(serverName, toolName string) string {
	serverName = strings.TrimSpace(serverName)
	toolName = strings.TrimSpace(toolName)

	if strings.Contains(toolName, ":") {
		return toolName
	}

	return serverName + ":" + toolName
}

func inferMCPToolAccess(toolName string) MCPToolAccess {
	name := strings.ToLower(strings.TrimSpace(toolName))

	if index := strings.LastIndex(name, ":"); index >= 0 {
		name = name[index+1:]
	}

	destructivePrefixes := []string{
		"delete_",
		"remove_",
		"kill_",
		"cancel_",
		"rollback_",
		"reset_",
		"uninstall_",
		"restore_snapshot",
		"restore_backup",
		"restore_trashed",
		"git_restore",
		"process_kill",
	}

	for _, prefix := range destructivePrefixes {
		if strings.HasPrefix(name, prefix) {
			return MCPToolAccessDestructive
		}
	}

	writePrefixes := []string{
		"write_",
		"patch_",
		"create_",
		"copy_",
		"move_",
		"append_",
		"apply_",
		"update_",
		"set_",
		"sync_",
		"configure_",
		"register_",
		"enable_",
		"disable_",
		"activate_",
		"deactivate_",
		"install_",
		"commit_",
		"ingest_",
		"promote_",
		"trigger_",
	}

	for _, prefix := range writePrefixes {
		if strings.HasPrefix(name, prefix) {
			return MCPToolAccessWrite
		}
	}

	return MCPToolAccessRead
}

// CallTool invokes one upstream tool through apps/tibrain/mcp.
func (c *MCPHubClient) CallTool(
	ctx context.Context,
	serverName string,
	toolName string,
	args map[string]interface{},
) (interface{}, error) {
	return c.CallToolWithAccess(
		ctx,
		serverName,
		toolName,
		args,
		inferMCPToolAccess(toolName),
	)
}

func (c *MCPHubClient) CallToolWithAccess(
	ctx context.Context,
	serverName string,
	toolName string,
	args map[string]interface{},
	access MCPToolAccess,
) (interface{}, error) {
	serverName = strings.TrimSpace(serverName)
	toolName = strings.TrimSpace(toolName)

	if serverName == "" {
		return nil, fmt.Errorf("server name is required")
	}
	if toolName == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	if _, err := c.requireAPIKey(); err != nil {
		return nil, err
	}

	if args == nil {
		args = map[string]interface{}{}
	}

	targetToolName := qualifyMCPToolName(serverName, toolName)

	wrapperTool, err := wrapperToolForAccess(access)
	if err != nil {
		return nil, err
	}

	reqBody := map[string]interface{}{
		"tool_name": wrapperTool,
		"arguments": map[string]interface{}{
			"name":      targetToolName,
			"arguments": args,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("encode MCP tool request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.hubURL+"/api/v1/tools/call",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("create MCP tool request: %w", err)
	}

	if err := c.applyAPIKey(req); err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call MCP tool: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, readMCPHubErrorResponse(resp)
	}

	var apiResp struct {
		Success   bool            `json:"success"`
		Data      json.RawMessage `json:"data"`
		Result    json.RawMessage `json:"result"`
		Error     string          `json:"error"`
		RequestID string          `json:"request_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode MCP hub response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf(
			"MCP hub error: %s; request_id=%s",
			apiResp.Error,
			apiResp.RequestID,
		)
	}

	payload := apiResp.Data
	if len(payload) == 0 || string(payload) == "null" {
		payload = apiResp.Result
	}

	if len(payload) == 0 || string(payload) == "null" {
		return nil, nil
	}

	var result interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, fmt.Errorf("decode MCP tool result: %w", err)
	}

	return result, nil
}

// GetServerStatus looks up a single upstream server from the server list.
func (c *MCPHubClient) GetServerStatus(ctx context.Context, serverName string) (*MCPServerInfo, error) {
	servers, err := c.ListServers(ctx)
	if err != nil {
		return nil, err
	}

	for _, info := range servers {
		if info.Name == serverName || info.ID == serverName {
			copy := info
			return &copy, nil
		}
	}

	return nil, fmt.Errorf("server %q not found", serverName)
}

// SyncToRegistry registers upstream tools into TiBrain's registry.
func (c *MCPHubClient) SyncToRegistry(ctx context.Context, hub *Hub) error {
	tools, err := c.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("list MCP tools: %w", err)
	}

	for _, tool := range tools {
		toolObj := Tool{
			ID:          fmt.Sprintf("mcp-%s-%s", tool.ServerName, tool.Name),
			Name:        fmt.Sprintf("%s:%s", tool.ServerName, tool.Name),
			Description: tool.Description,
			Category:    "mcp-proxy",
			Enabled:     true,
			Source:      fmt.Sprintf("mcp://%s", tool.ServerName),
		}
		if err := hub.RegisterTool(toolObj); err != nil {
			logger.Warn("Failed to register MCP tool %s: %v", tool.Name, err)
		}
	}

	return nil
}

// BatchCallTools calls tools one by one through the proxy.
func (c *MCPHubClient) BatchCallTools(ctx context.Context, calls []MCPToolCall) ([]MCPToolResult, error) {
	if _, err := c.requireAPIKey(); err != nil {
		return nil, err
	}
	results := make([]MCPToolResult, 0, len(calls))
	for i, call := range calls {
		result, err := c.CallTool(ctx, call.Server, call.Tool, call.Args)
		if err != nil {
			results = append(results, MCPToolResult{
				CallIndex: i,
				Success:   false,
				Error:     err.Error(),
			})
			continue
		}

		results = append(results, MCPToolResult{
			CallIndex: i,
			Success:   true,
			Result:    result,
		})
	}

	return results, nil
}

// MCPToolCall stores one proxy tool call.
type MCPToolCall struct {
	Server string                 `json:"server"`
	Tool   string                 `json:"tool"`
	Args   map[string]interface{} `json:"args"`
}

// MCPToolResult stores one proxy tool call result.
type MCPToolResult struct {
	CallIndex int         `json:"call_index"`
	Success   bool        `json:"success"`
	Result    interface{} `json:"result,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}
