package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestRegisterUnifiedPublicTools_RespectsSurfaceLimit(t *testing.T) {
	config := defaultConfig()
	config.MCP.PublicMode = "unified"
	config.MCP.Compatibility = false
	manager := NewMCPServerManager(nil, "http://127.0.0.1:1810", nil, config)

	tools := manager.mcpServer.ListTools(context.Background())
	if len(tools) > 12 {
		t.Fatalf("public unified tool count = %d, want <= 12", len(tools))
	}
}

func TestHandleBrainPreflight_BasicChecks(t *testing.T) {
	hubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/servers":
			_, _ = w.Write([]byte(`{"success":true,"data":{"servers":[{"name":"1mcp-local-bridge","status":"connected","enabled":true,"connected":true,"tool_count":1}]}}`))
		case "/api/v1/tools":
			_, _ = w.Write([]byte(`{"success":true,"data":{"tools":[{"name":"get_allowed_roots","server_name":"1mcp-local-bridge","description":"roots"}]}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer hubServer.Close()

	localAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/health" || r.URL.Path == "/api/health" {
			_, _ = w.Write([]byte(`{"service":"TiBrain","status":"healthy","port":1810}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer localAPI.Close()

	config := defaultConfig()
	config.MCP.PublicMode = "unified"
	manager := NewMCPServerManager(nil, localAPI.URL, NewMCPHubClient(hubServer.URL, "test-key"), config)

	result, err := manager.handleBrainPreflight(context.Background(), testCallToolRequest(map[string]any{
		"target": "filesystem",
		"name":   "get_allowed_roots",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(map[string]any)["text"].(string)
	var payload map[string]any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	checks := payload["checks"].(map[string]any)
	if checks["tibrain_local"] != "pass" || checks["hub_servers"] != "pass" || checks["hub_tools"] != "pass" || checks["target_lookup"] != "pass" {
		t.Fatalf("unexpected checks: %#v", checks)
	}
}

func testCallToolRequest(args map[string]any) mcp.CallToolRequest {
	request := mcp.CallToolRequest{}
	request.Params.Arguments = args
	return request
}