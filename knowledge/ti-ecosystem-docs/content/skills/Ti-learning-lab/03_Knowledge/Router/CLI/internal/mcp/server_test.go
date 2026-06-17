package mcp_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ti/cli/internal/mcp"
)

func newTestServer() *mcp.Server {
	tools := []mcp.Tool{
		{
			Name:        "test.echo",
			Description: "Echo the input back",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"msg": map[string]any{"type": "string"},
				},
			},
			Handler: func(args map[string]any) (any, error) {
				msg, _ := args["msg"].(string)
				return "echo: " + msg, nil
			},
		},
	}
	return mcp.New(tools)
}

func rpc(t *testing.T, srv *mcp.Server, method string, params any) map[string]any {
	t.Helper()
	var rawParams json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("marshal params: %v", err)
		}
		rawParams = b
	}
	req := mcp.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  rawParams,
	}
	in, _ := json.Marshal(req)
	var out bytes.Buffer
	srv.HandleRequest(bytes.NewReader(in), &out)
	var resp map[string]any
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v\nraw: %s", err, out.String())
	}
	return resp
}

func TestInitialize(t *testing.T) {
	srv := newTestServer()
	resp := rpc(t, srv, "initialize", nil)
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("result not a map: %v", resp["result"])
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("wrong protocolVersion: %v", result["protocolVersion"])
	}
	info, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatalf("serverInfo not a map")
	}
	if info["name"] != "ti-mcp" {
		t.Errorf("wrong server name: %v", info["name"])
	}
}

func TestToolsList(t *testing.T) {
	srv := newTestServer()
	resp := rpc(t, srv, "tools/list", nil)
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("result not a map")
	}
	tools, ok := result["tools"].([]any)
	if !ok {
		t.Fatalf("tools not a slice")
	}
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	tool := tools[0].(map[string]any)
	if tool["name"] != "test.echo" {
		t.Errorf("wrong tool name: %v", tool["name"])
	}
}

func TestToolsCall(t *testing.T) {
	srv := newTestServer()
	resp := rpc(t, srv, "tools/call", map[string]any{
		"name":      "test.echo",
		"arguments": map[string]any{"msg": "hello"},
	})
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %v", resp["error"])
	}
	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("result not a map")
	}
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("content missing or empty")
	}
	item := content[0].(map[string]any)
	text, _ := item["text"].(string)
	if !strings.Contains(text, "echo: hello") {
		t.Errorf("unexpected text: %q", text)
	}
}

func TestToolsCallNotFound(t *testing.T) {
	srv := newTestServer()
	resp := rpc(t, srv, "tools/call", map[string]any{
		"name":      "nonexistent",
		"arguments": map[string]any{},
	})
	if resp["error"] == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestUnknownMethod(t *testing.T) {
	srv := newTestServer()
	resp := rpc(t, srv, "unknown/method", nil)
	if resp["error"] == nil {
		t.Fatal("expected error for unknown method")
	}
}
