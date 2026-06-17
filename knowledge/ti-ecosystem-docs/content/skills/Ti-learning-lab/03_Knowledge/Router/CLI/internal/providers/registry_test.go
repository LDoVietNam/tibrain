package providers

import (
	"context"
	"errors"
	"testing"
)

type stubProvider struct {
	name, model string
	models      []string
	healthy     bool
	err         error
}

func (s stubProvider) Name() string         { return s.name }
func (s stubProvider) DefaultModel() string { return s.model }
func (s stubProvider) Models() []string     { return s.models }
func (s stubProvider) IsHealthy() bool      { return s.healthy }
func (s stubProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &ChatResponse{Content: s.name}, nil
}
func (s stubProvider) ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	return s.Chat(ctx, req)
}

func TestRegistrySelectPreferredHealthy(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(stubProvider{name: "a", model: "m1", models: []string{"m1"}, healthy: false})
	_ = r.Register(stubProvider{name: "b", model: "m2", models: []string{"m2"}, healthy: true})
	p, err := r.Select(SelectionCriteria{Preferred: []string{"a", "b"}, RequireHealthy: true})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "b" {
		t.Fatalf("got %s want b", p.Name())
	}
}

func TestRegistryChatWithFallback(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(stubProvider{name: "broken", model: "m1", models: []string{"m1"}, healthy: true, err: errors.New("boom")})
	_ = r.Register(stubProvider{name: "ok", model: "m2", models: []string{"m2"}, healthy: true})
	resp, name, err := r.ChatWithFallback(context.Background(), ChatRequest{}, "broken", "ok")
	if err != nil {
		t.Fatal(err)
	}
	if name != "ok" {
		t.Fatalf("provider=%s want ok", name)
	}
	if resp.Content != "ok" {
		t.Fatalf("content=%q", resp.Content)
	}
}
