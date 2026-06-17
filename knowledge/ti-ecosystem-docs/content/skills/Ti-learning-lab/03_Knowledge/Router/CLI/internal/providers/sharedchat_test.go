package providers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ti/cli/internal/providers"
)

func TestSharedChatProviderName(t *testing.T) {
	cfg := &providers.SharedChatConfig{SessionID: "test-sid"}
	p, err := providers.NewSharedChatProvider(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "sharedchat" {
		t.Fatalf("name = %q, want 'sharedchat'", p.Name())
	}
}

func TestSharedChatProviderDefaults(t *testing.T) {
	cfg := &providers.SharedChatConfig{SessionID: "test-sid"}
	p, err := providers.NewSharedChatProvider(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if p.DefaultModel() != "gpt-4o" {
		t.Fatalf("default model = %q, want 'gpt-4o'", p.DefaultModel())
	}
	if len(p.Models()) != 3 {
		t.Fatalf("models count = %d, want 3", len(p.Models()))
	}
	if !p.IsHealthy() {
		t.Fatal("expected healthy")
	}
	if p.SessionID() != "test-sid" {
		t.Fatalf("session ID = %q, want 'test-sid'", p.SessionID())
	}
}

func TestSharedChatProviderCustomConfig(t *testing.T) {
	cfg := &providers.SharedChatConfig{
		SessionID: "custom-sid", APIPrimary: "https://custom.cc",
		APIFallback: "https://custom.fun", CookieDomain: "custom.domain",
		Models: []string{"custom-model"}, DefaultModel: "custom-model",
	}
	p, err := providers.NewSharedChatProvider(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if p.DefaultModel() != "custom-model" {
		t.Fatalf("default model = %q", p.DefaultModel())
	}
	if len(p.Models()) != 1 || p.Models()[0] != "custom-model" {
		t.Fatalf("models = %v", p.Models())
	}
}

func TestSharedChatProviderConversation(t *testing.T) {
	cfg := &providers.SharedChatConfig{SessionID: "test-sid"}
	p, err := providers.NewSharedChatProvider(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if p.ConversationID() != "" {
		t.Fatal("expected empty conversation ID")
	}
	p.SetConversationID("conv-123")
	if p.ConversationID() != "conv-123" {
		t.Fatalf("conversation ID = %q", p.ConversationID())
	}
	p.ResetConversation()
	if p.ConversationID() != "" {
		t.Fatal("expected empty after reset")
	}
}

func TestSharedChatProviderTokenExchange(t *testing.T) {
	var tokenCalls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tokenCalls, 1)
		if r.URL.Path == "/api/auth/session" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"accessToken": "mock-token-abc", "expiresIn": 600,
			})
		}
	}))
	defer server.Close()

	cfg := &providers.SharedChatConfig{SessionID: "test-sid", APIPrimary: server.URL}
	p, _ := providers.NewSharedChatProvider(cfg)
	_ = p // Token fetched lazily during Chat
}

func TestSharedChatProviderFailover(t *testing.T) {
	var primaryCalls, fallbackCalls int32

	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&primaryCalls, 1)
		if r.URL.Path == "/api/auth/session" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"accessToken": "primary-token", "expiresIn": 600})
			return
		}
		http.Error(w, "primary down", http.StatusServiceUnavailable)
	}))
	defer primary.Close()

	fallback := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&fallbackCalls, 1)
		if r.URL.Path == "/api/auth/session" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"accessToken": "fallback-token", "expiresIn": 600})
			return
		}
		if r.URL.Path == "/backend-api/conversation" {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Write([]byte(`data: {"role":"assistant","content":"hello","done":true}` + "\n\n"))
		}
	}))
	defer fallback.Close()

	cfg := &providers.SharedChatConfig{
		SessionID: "test-sid", APIPrimary: primary.URL, APIFallback: fallback.URL,
	}
	p, err := providers.NewSharedChatProvider(cfg)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := p.Chat(ctx, providers.ChatRequest{
		Messages: []providers.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	if resp.Content != "hello" {
		t.Fatalf("content = %q", resp.Content)
	}
	if atomic.LoadInt32(&primaryCalls) == 0 || atomic.LoadInt32(&fallbackCalls) == 0 {
		t.Fatal("expected both primary and fallback to be called")
	}
}

func TestSharedChatProviderStreaming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/session" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"accessToken": "token", "expiresIn": 600})
			return
		}
		if r.URL.Path == "/backend-api/conversation" {
			w.Header().Set("Content-Type", "text/event-stream")
			for _, c := range []string{
				`data: {"role":"assistant","content":"Hel"}`,
				`data: {"role":"assistant","content":"lo "}`,
				`data: {"role":"assistant","content":"world","done":true}`,
			} {
				w.Write([]byte(c + "\n\n"))
			}
		}
	}))
	defer server.Close()

	cfg := &providers.SharedChatConfig{SessionID: "test-sid", APIPrimary: server.URL}
	p, _ := providers.NewSharedChatProvider(cfg)

	var chunks []string
	ctx := context.Background()
	resp, err := p.ChatStream(ctx, providers.ChatRequest{
		Messages: []providers.Message{{Role: "user", Content: "hi"}},
	}, func(chunk providers.StreamChunk) { chunks = append(chunks, chunk.Content) })
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}
	if len(chunks) < 3 {
		t.Fatalf("got %d chunks", len(chunks))
	}
	if resp.Content != "Hello world" {
		t.Fatalf("content = %q", resp.Content)
	}
}

// Skipped on Windows: httptest.Server.Close blocks on context cancel
func TestSharedChatProviderCancel(t *testing.T) {
	t.Skip("skipping on Windows: httptest.Server.Close blocks on context cancel")
}

func TestSharedChatProviderHealthTracking(t *testing.T) {
	var failCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/session" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"accessToken": "token", "expiresIn": 600})
			return
		}
		if atomic.AddInt32(&failCount, 1) <= 2 {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte(`data: {"content":"ok","done":true}` + "\n\n"))
	}))
	defer server.Close()

	cfg := &providers.SharedChatConfig{SessionID: "test-sid", APIPrimary: server.URL}
	p, _ := providers.NewSharedChatProvider(cfg)
	ctx := context.Background()

	// First 2 calls fail
	p.Chat(ctx, providers.ChatRequest{Messages: []providers.Message{{Role: "user", Content: "hi"}}})
	p.Chat(ctx, providers.ChatRequest{Messages: []providers.Message{{Role: "user", Content: "hi"}}})

	// Third call succeeds
	resp, err := p.Chat(ctx, providers.ChatRequest{Messages: []providers.Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatalf("third call failed: %v", err)
	}
	if resp.Content != "ok" {
		t.Fatalf("content = %q", resp.Content)
	}
	if !p.IsHealthy() {
		t.Fatal("expected healthy after success")
	}
}
