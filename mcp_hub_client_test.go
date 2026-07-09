package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQualifyMCPToolName(t *testing.T) {
	if got := qualifyMCPToolName("srv", "tool"); got != "srv:tool" {
		t.Fatalf("qualifyMCPToolName() = %q, want %q", got, "srv:tool")
	}
	if got := qualifyMCPToolName("srv", "other:tool"); got != "other:tool" {
		t.Fatalf("qualified tool name should be preserved, got %q", got)
	}
}

func TestInferMCPToolAccess(t *testing.T) {
	tests := []struct {
		toolName string
		expected MCPToolAccess
	}{
		{"1mcp-local-bridge:get_allowed_roots", MCPToolAccessRead},
		{"1mcp-local-bridge:write_file", MCPToolAccessWrite},
		{"1mcp-local-bridge:delete_path", MCPToolAccessDestructive},
		{"process_kill", MCPToolAccessDestructive},
		{"git_restore", MCPToolAccessDestructive},
		{"sync_obsidian", MCPToolAccessWrite},
		{"read_file", MCPToolAccessRead},
	}

	for _, tc := range tests {
		got := inferMCPToolAccess(tc.toolName)
		if got != tc.expected {
			t.Errorf("inferMCPToolAccess(%q) = %q, expected %q", tc.toolName, got, tc.expected)
		}
	}
}

func TestMCPHubClient_CallToolWithAccess(t *testing.T) {
	t.Run("read wrapper format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/tools/call" {
				t.Errorf("expected path /api/v1/tools/call, got %q", r.URL.Path)
			}
			if r.Header.Get("X-API-Key") != "test-key" {
				t.Errorf("expected X-API-Key test-key, got %q", r.Header.Get("X-API-Key"))
			}

			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)

			toolName := body["tool_name"].(string)
			if toolName != "1mcp-local-bridge:call_tool_read" {
				t.Errorf("expected tool_name '1mcp-local-bridge:call_tool_read', got %q", toolName)
			}

			args := body["arguments"].(map[string]interface{})
			if args["name"] != "1mcp-local-bridge:get_allowed_roots" {
				t.Errorf("expected arguments.name '1mcp-local-bridge:get_allowed_roots', got %q", args["name"])
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success":true,"data":["Z:\\01_PROJECTS"]}`))
		}))
		defer server.Close()

		client := NewMCPHubClient(server.URL, "test-key")
		res, err := client.CallTool(context.Background(), "1mcp-local-bridge", "get_allowed_roots", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr, ok := res.([]interface{})
		if !ok || len(arr) != 1 || arr[0] != "Z:\\01_PROJECTS" {
			t.Errorf("unexpected result type or value: %v", res)
		}
	})

	t.Run("write wrapper format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["tool_name"] != "test-srv:call_tool_write" {
				t.Fatalf("unexpected wrapper name: %v", body["tool_name"])
			}
			args := body["arguments"].(map[string]interface{})
			if args["name"] != "test-srv:write_file" {
				t.Fatalf("unexpected wrapped target: %v", args["name"])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"ok":true}}`))
		}))
		defer server.Close()

		client := NewMCPHubClient(server.URL, "test-key")
		res, err := client.CallTool(context.Background(), "test-srv", "write_file", map[string]interface{}{"path": "a.txt"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		obj := res.(map[string]interface{})
		if obj["ok"] != true {
			t.Fatalf("unexpected response: %#v", res)
		}
	})

	t.Run("destructive wrapper format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["tool_name"] != "test-srv:call_tool_destructive" {
				t.Fatalf("unexpected wrapper name: %v", body["tool_name"])
			}
			args := body["arguments"].(map[string]interface{})
			if args["name"] != "test-srv:delete_path" {
				t.Fatalf("unexpected wrapped target: %v", args["name"])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":true}`))
		}))
		defer server.Close()

		client := NewMCPHubClient(server.URL, "test-key")
		res, err := client.CallTool(context.Background(), "test-srv", "delete_path", map[string]interface{}{"path": "a.txt"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v, ok := res.(bool); !ok || !v {
			t.Fatalf("unexpected response: %#v", res)
		}
	})

	t.Run("array response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success":true,"data":["foo", "bar"]}`))
		}))
		defer server.Close()

		client := NewMCPHubClient(server.URL, "test-key")
		res, err := client.CallTool(context.Background(), "test-srv", "list_something", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr, ok := res.([]interface{})
		if !ok || len(arr) != 2 || arr[0] != "foo" || arr[1] != "bar" {
			t.Errorf("unexpected result: %v", res)
		}
	})

	t.Run("object response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success":true,"data":{"status":"healthy"}}`))
		}))
		defer server.Close()

		client := NewMCPHubClient(server.URL, "test-key")
		res, err := client.CallTool(context.Background(), "test-srv", "health", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		obj, ok := res.(map[string]interface{})
		if !ok || obj["status"] != "healthy" {
			t.Errorf("unexpected result: %v", res)
		}
	})

	t.Run("null response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success":true,"data":null}`))
		}))
		defer server.Close()

		client := NewMCPHubClient(server.URL, "test-key")
		res, err := client.CallTool(context.Background(), "test-srv", "void", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != nil {
			t.Errorf("expected nil result, got %v", res)
		}
	})

	t.Run("string number and boolean responses", func(t *testing.T) {
		cases := []struct {
			name string
			body string
			want interface{}
		}{
			{name: "string", body: `{"success":true,"data":"ok"}`, want: "ok"},
			{name: "number", body: `{"success":true,"data":42}`, want: float64(42)},
			{name: "bool", body: `{"success":true,"data":false}`, want: false},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(tc.body))
				}))
				defer server.Close()

				client := NewMCPHubClient(server.URL, "test-key")
				res, err := client.CallTool(context.Background(), "test-srv", "health", nil)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if res != tc.want {
					t.Fatalf("got %#v, want %#v", res, tc.want)
				}
			})
		}
	})

	t.Run("401 error handling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`Unauthorized`))
		}))
		defer server.Close()

		client := NewMCPHubClient(server.URL, "invalid-key")
		_, err := client.CallTool(context.Background(), "test-srv", "secret", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "MCP hub auth failed (401)") {
			t.Errorf("expected auth error, got %q", err.Error())
		}
	})
}
