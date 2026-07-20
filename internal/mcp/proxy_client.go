// Package mcp provides MCP proxy client stub for demonstration
package mcp

import (
	"context"
	"encoding/json"
)

// ProxyCallRequest represents MCP tool call request
type ProxyCallRequest struct {
	ServerName string                 `json:"server_name"`
	ToolName   string                 `json:"tool_name"`
	Arguments  map[string]interface{} `json:"arguments"`
}

// ProxyCallResponse represents MCP tool call response
type ProxyCallResponse struct {
	Content  interface{} `json:"content"`
	Error    string      `json:"error,omitempty"`
	RawJSON  string      `json:"raw_json,omitempty"`
}

// ListMCPServers lists registered MCP servers
func (pe *PolicyEngine) ListMCPServers() []string {
	return []string{"github-mcp", "obsidian-mcp", "chrome-devtools-mcp"}
}

// ListMCPTools lists tools for a server
func (pe *PolicyEngine) ListMCPTools(serverName string) []string {
	switch serverName {
	case "github-mcp":
		return []string{"list_repositories", "create_issue", "list_issues"}
	case "obsidian-mcp":
		return []string{"search_notes", "read_note", "create_note"}
	case "chrome-devtools-mcp":
		return []string{"navigate", "screenshot", "evaluate_js"}
	}
	return []string{}
}

// CallMCPTool demonstrates calling MCP tool via proxy
func (pe *PolicyEngine) CallMCPTool(ctx context.Context, req *ProxyCallRequest) (*ProxyCallResponse, error) {
	// Demo: github-mcp list_repositories
	if req.ServerName == "github-mcp" && req.ToolName == "list_repositories" {
		// In real implementation, this would connect to github-mcp via stdio
		return &ProxyCallResponse{
			Content: []string{"tibrain", "tirouter", "Tiiextension"},
		}, nil
	}
	
	return &ProxyCallResponse{Error: "tool not found"}, nil
}

// Example usage:
// curl -X POST http://localhost:1810/mcp/call \
//   -H "Content-Type: application/json" \
//   -d '{"server_name":"github-mcp","tool_name":"list_repositories"}'
func (pe *PolicyEngine) CallMCPToolJSON(jsonReq string) (string, error) {
	var req ProxyCallRequest
	if err := json.Unmarshal([]byte(jsonReq), &req); err != nil {
		return "", err
	}
	resp, err := pe.CallMCPTool(context.Background(), &req)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(resp)
	return string(data), nil
}