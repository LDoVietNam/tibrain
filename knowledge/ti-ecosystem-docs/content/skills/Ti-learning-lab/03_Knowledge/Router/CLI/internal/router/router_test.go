package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// Helpers
// =============================================================================

func newTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func newTestClient(srv *httptest.Server, opts ...ClientOption) *Client {
	allOpts := []ClientOption{WithAPIKey("test-key")}
	allOpts = append(allOpts, opts...)
	return New(srv.URL, allOpts...)
}

// =============================================================================
// Tests: Client construction
// =============================================================================

func TestNewDefaultBaseURL(t *testing.T) {
	c := New("")
	if c.baseURL != "http://127.0.0.1:1810" {
		t.Fatalf("expected default baseURL, got %q", c.baseURL)
	}
}

func TestNewCustomBaseURL(t *testing.T) {
	c := New("http://localhost:9999")
	if c.baseURL != "http://localhost:9999" {
		t.Fatalf("expected custom baseURL, got %q", c.baseURL)
	}
}

func TestWithOptions(t *testing.T) {
	c := New("http://localhost:9999",
		WithAPIKey("my-key"),
		WithManagementKey("mgmt-key"),
	)
	if c.apiKey != "my-key" {
		t.Fatalf("expected apiKey 'my-key', got %q", c.apiKey)
	}
	if c.managementKey != "mgmt-key" {
		t.Fatalf("expected managementKey 'mgmt-key', got %q", c.managementKey)
	}
}

func TestSetBaseURL(t *testing.T) {
	c := New("")
	c.SetBaseURL("http://new-host:1234")
	if c.baseURL != "http://new-host:1234" {
		t.Fatalf("expected updated baseURL, got %q", c.baseURL)
	}
}

func TestSetManagementURL(t *testing.T) {
	c := New("http://main:1810")
	c.SetManagementURL("http://mgmt:1810")
	if c.managementURL != "http://mgmt:1810" {
		t.Fatalf("expected managementURL 'http://mgmt:1810', got %q", c.managementURL)
	}
}

// =============================================================================
// Tests: GetModels
// =============================================================================

func TestGetModels(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			http.Error(w, "not found", 404)
			return
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			http.Error(w, "unauthorized", 401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{
				{"id": "gpt-4o", "owned_by": "openai"},
				{"id": "claude-sonnet", "owned_by": "anthropic"},
			},
		})
	})
	defer srv.Close()

	c := newTestClient(srv)
	models, err := c.GetModels()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	if models[0].ID != "gpt-4o" {
		t.Fatalf("expected first model 'gpt-4o', got %q", models[0].ID)
	}
	if models[1].OwnedBy != "anthropic" {
		t.Fatalf("expected second owned_by 'anthropic', got %q", models[1].OwnedBy)
	}
}

func TestGetModelsServerError(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", 500)
	})
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.GetModels()
	if err == nil {
		t.Fatal("expected error on server failure")
	}
}

// =============================================================================
// Tests: Chat (non-streaming)
// =============================================================================

func TestChat(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/chat/completions" {
			http.Error(w, "not found", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "chatcmpl-123",
			"model": "gpt-4o",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": "Hello from AI!",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]int{
				"prompt_tokens":     10,
				"completion_tokens": 5,
				"total_tokens":      15,
			},
		})
	})
	defer srv.Close()

	c := newTestClient(srv)
	resp, err := c.Chat("gpt-4o", []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "chatcmpl-123" {
		t.Fatalf("expected id 'chatcmpl-123', got %q", resp.ID)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "Hello from AI!" {
		t.Fatalf("expected content 'Hello from AI!', got %q", resp.Choices[0].Message.Content)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Fatalf("expected 15 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestChatWithOptions(t *testing.T) {
	var receivedBody ChatRequest
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "x",
			"model":   receivedBody.Model,
			"choices": []map[string]interface{}{{"message": map[string]string{"role": "assistant", "content": "ok"}, "finish_reason": "stop"}},
		})
	})
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Chat("gpt-4o", []Message{{Role: "user", Content: "hi"}},
		WithMaxTokens(100),
		WithTemperature(0.7),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedBody.MaxTokens != 100 {
		t.Fatalf("expected max_tokens 100, got %d", receivedBody.MaxTokens)
	}
	if receivedBody.Temperature != 0.7 {
		t.Fatalf("expected temperature 0.7, got %.1f", receivedBody.Temperature)
	}
}

func TestChatModelInError(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model not found", 404)
	})
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Chat("unknown-model", []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error for unknown model")
	}
	if !strings.Contains(err.Error(), "unknown-model") {
		t.Fatalf("expected error to contain model name, got: %v", err)
	}
	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected error to contain status code, got: %v", err)
	}
}

// =============================================================================
// Tests: ChatStream
// =============================================================================

func TestChatStream(t *testing.T) {
	sseData := `data: {"id":"stream-1","model":"gpt-4o","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":""}]}

data: {"id":"stream-1","model":"gpt-4o","choices":[{"index":0,"delta":{"content":" World"},"finish_reason":""}]}

data: {"id":"stream-1","model":"gpt-4o","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":"stop"}]}

data: [DONE]
`
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	})
	defer srv.Close()

	c := newTestClient(srv)

	var chunks []string
	resp, err := c.ChatStream("gpt-4o", []Message{{Role: "user", Content: "hi"}}, func(chunk string) {
		chunks = append(chunks, chunk)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if chunks[0] != "Hello" || chunks[1] != " World" || chunks[2] != "!" {
		t.Fatalf("unexpected chunk content: %v", chunks)
	}
	if resp.Choices[0].Message.Content != "Hello World!" {
		t.Fatalf("expected final content 'Hello World!', got %q", resp.Choices[0].Message.Content)
	}
	if resp.Choices[0].FinishReason != "stop" {
		t.Fatalf("expected finish_reason 'stop', got %q", resp.Choices[0].FinishReason)
	}
}

func TestChatStreamNilOnChunk(t *testing.T) {
	sseData := `data: {"id":"x","model":"m","choices":[{"index":0,"delta":{"content":"hi"},"finish_reason":"stop"}]}

data: [DONE]
`
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, sseData)
	})
	defer srv.Close()

	c := newTestClient(srv)
	resp, err := c.ChatStream("m", []Message{{Role: "user", Content: "hi"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Choices[0].Message.Content != "hi" {
		t.Fatalf("expected content 'hi', got %q", resp.Choices[0].Message.Content)
	}
}

// =============================================================================
// Tests: Health
// =============================================================================

func TestHealthOK(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []string{}})
	})
	defer srv.Close()

	c := newTestClient(srv)
	err := c.Health()
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestHealthFail(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", 503)
	})
	defer srv.Close()

	c := newTestClient(srv)
	err := c.Health()
	if err == nil {
		t.Fatal("expected error on health failure")
	}
}

// =============================================================================
// Tests: Ping
// =============================================================================

func TestPing(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "x",
			"model":   "test",
			"choices": []map[string]interface{}{{"message": map[string]string{"role": "assistant", "content": "hi"}, "finish_reason": "stop"}},
		})
	})
	defer srv.Close()

	c := newTestClient(srv)
	result, err := c.Ping("test-model")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Model != "test-model" {
		t.Fatalf("expected model 'test-model', got %q", result.Model)
	}
	if result.Rounds != 3 {
		t.Fatalf("expected 3 rounds, got %d", result.Rounds)
	}
	// Local mock server can respond in 0ns; just verify no panic
	if result.AvgLatency < 0 {
		t.Fatal("expected non-negative avg latency")
	}
	if result.MinLatency > result.MaxLatency {
		t.Fatalf("invalid latency relationship: min=%v max=%v",
			result.MinLatency, result.MaxLatency)
	}
}

func TestPingFailure(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model not found", 404)
	})
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Ping("bad-model")
	if err == nil {
		t.Fatal("expected error on ping failure")
	}
}

// =============================================================================
// Tests: Auth File Management
// =============================================================================

func TestGetAuthFiles(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v0/management/auth-files" {
			http.Error(w, "not found", 404)
			return
		}
		if key := r.Header.Get("X-Management-Key"); key != "mgmt-key" {
			http.Error(w, "unauthorized", 401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]AuthFileInfo{
			{Name: "openai.json", Size: 128},
			{Name: "anthropic.yaml", Size: 256},
		})
	})
	defer srv.Close()

	c := newTestClient(srv, WithManagementKey("mgmt-key"))
	files, err := c.GetAuthFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 auth files, got %d", len(files))
	}
	if files[0].Name != "openai.json" {
		t.Fatalf("expected first file 'openai.json', got %q", files[0].Name)
	}
}

func TestUploadAuthFile(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
	})
	defer srv.Close()

	c := newTestClient(srv)
	err := c.UploadAuthFile("test.json", []byte(`{"key": "value"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAuthFile(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	defer srv.Close()

	c := newTestClient(srv)
	err := c.DeleteAuthFile("old.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAuthFileNotFound(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", 404)
	})
	defer srv.Close()

	c := newTestClient(srv)
	err := c.DeleteAuthFile("missing.json")
	if err == nil {
		t.Fatal("expected error for missing auth file")
	}
}

// =============================================================================
// Tests: Config
// =============================================================================

func TestGetConfig(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"routing": "round-robin"})
	})
	defer srv.Close()

	c := newTestClient(srv)
	data, err := c.GetConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), "round-robin") {
		t.Fatalf("expected config to contain 'round-robin', got: %s", string(data))
	}
}

// =============================================================================
// Tests: Routing Strategy
// =============================================================================

func TestSetRoutingStrategyRoundRobin(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	defer srv.Close()

	c := newTestClient(srv)
	err := c.SetRoutingStrategy(RoundRobin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetRoutingStrategyFillFirst(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	defer srv.Close()

	c := newTestClient(srv)
	err := c.SetRoutingStrategy(FillFirst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetRoutingStrategyInvalid(t *testing.T) {
	c := New("")
	err := c.SetRoutingStrategy("invalid-strategy")
	if err == nil {
		t.Fatal("expected error for invalid strategy")
	}
	if !strings.Contains(err.Error(), "invalid-strategy") {
		t.Fatalf("expected error to contain strategy name, got: %v", err)
	}
}

// =============================================================================
// Tests: Share / Fetch Model
// =============================================================================

func TestShareModel(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
	})
	defer srv.Close()

	c := newTestClient(srv)
	err := c.ShareModel("gpt-4o", "openai", map[string]interface{}{
		"context_size": 128000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchModel(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"model_id": "gpt-4o",
			"provider": "openai",
		})
	})
	defer srv.Close()

	c := newTestClient(srv)
	result, err := c.FetchModel("gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["model_id"] != "gpt-4o" {
		t.Fatalf("expected model_id 'gpt-4o', got %v", result["model_id"])
	}
}

func TestListSharedModels(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"model_id": "gpt-4o", "provider": "openai"},
			{"model_id": "claude", "provider": "anthropic"},
		})
	})
	defer srv.Close()

	c := newTestClient(srv)
	models, err := c.ListSharedModels()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 shared models, got %d", len(models))
	}
}

// =============================================================================
// Tests: APIError
// =============================================================================

func TestAPIErrorWithModel(t *testing.T) {
	err := &APIError{
		Method:     "POST",
		URL:        "/v1/chat/completions",
		StatusCode: 404,
		Body:       "model not found",
		ModelName:  "gpt-4o",
	}
	msg := err.Error()
	if !strings.Contains(msg, "gpt-4o") {
		t.Fatalf("expected error to contain model name, got: %s", msg)
	}
	if !strings.Contains(msg, "404") {
		t.Fatalf("expected error to contain status code, got: %s", msg)
	}
}

func TestAPIErrorWithoutModel(t *testing.T) {
	err := &APIError{
		Method:     "GET",
		URL:        "/v1/models",
		StatusCode: 500,
		Body:       "internal error",
	}
	msg := err.Error()
	if strings.Contains(msg, "model=") {
		t.Fatalf("expected error without model name, got: %s", msg)
	}
}

// =============================================================================
// Tests: Concurrent safety
// =============================================================================

func TestConcurrentRequests(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "x",
			"model":   "m",
			"choices": []map[string]interface{}{{"message": map[string]string{"role": "assistant", "content": "ok"}, "finish_reason": "stop"}},
		})
	})
	defer srv.Close()

	c := newTestClient(srv)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := c.Chat("m", []Message{{Role: "user", Content: "hi"}})
			if err != nil {
				t.Errorf("unexpected error in concurrent request: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestConcurrentSetBaseURL(t *testing.T) {
	c := New("")

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			c.SetBaseURL(fmt.Sprintf("http://host-%d:1810", n))
			c.SetManagementURL(fmt.Sprintf("http://mgmt-%d:1810", n))
			_ = c.Health()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
