// Package providers - agentic mode types and provider stubs.
package providers

import (
	"context"
	"fmt"

	"github.com/ti/cli/internal/config"
)

// AgentResponse is the result of a tool-aware (agentic) chat request.
// It contains the final text content, raw provider-specific content blocks
// (used to round-trip tool_use IDs), tool calls requested by the model,
// and a stop reason indicating whether the turn ended naturally.
type AgentResponse struct {
	Content    string     `json:"content"`
	RawBlocks  []any      `json:"raw_blocks,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	StopReason string     `json:"stop_reason,omitempty"`
	Usage      *Usage     `json:"usage,omitempty"`
}

// ClaudeProvider is the Anthropic Claude provider.
// Stub implementation — initialized via Init() with config.
type ClaudeProvider struct {
	APIKey  string
	Model   string
	BaseURL string
}

// Init configures the provider from config.
func (p *ClaudeProvider) Init(cfg *config.Config) error {
	if p == nil {
		return fmt.Errorf("nil provider")
	}
	if cfg != nil {
		if p.Model == "" {
			p.Model = "claude-sonnet-4-5"
		}
		if p.BaseURL == "" {
			p.BaseURL = "https://api.anthropic.com/v1"
		}
	}
	return nil
}

// Name returns the provider name.
func (p *ClaudeProvider) Name() string { return "anthropic" }

// DefaultModel returns the default model.
func (p *ClaudeProvider) DefaultModel() string {
	if p.Model == "" {
		return "claude-sonnet-4-5"
	}
	return p.Model
}

// Models returns supported models.
func (p *ClaudeProvider) Models() []string {
	return []string{"claude-sonnet-4-5", "claude-haiku-4", "claude-opus-4"}
}

// IsHealthy returns whether the provider is healthy.
func (p *ClaudeProvider) IsHealthy() bool { return p.APIKey != "" }

// Chat sends a chat request and returns the response.
func (p *ClaudeProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	return nil, fmt.Errorf("ClaudeProvider.Chat not implemented")
}

// ChatStream sends a chat request and streams chunks via onChunk.
func (p *ClaudeProvider) ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	return nil, fmt.Errorf("ClaudeProvider.ChatStream not implemented")
}

// ChatWithTools sends a tool-aware chat request, returning AgentResponse
// with any tool calls the model wants to make.
func (p *ClaudeProvider) ChatWithTools(ctx context.Context, req ChatRequest) (*AgentResponse, error) {
	return nil, fmt.Errorf("ClaudeProvider.ChatWithTools not implemented")
}

// AntigravityProvider is the Antigravity (Google) provider.
// Stub implementation.
type AntigravityProvider struct {
	APIKey  string
	Model   string
	BaseURL string
}

// Init configures the provider from config.
func (p *AntigravityProvider) Init(cfg *config.Config) error {
	if p == nil {
		return fmt.Errorf("nil provider")
	}
	return nil
}

// Name returns the provider name.
func (p *AntigravityProvider) Name() string { return "antigravity" }

// DefaultModel returns the default model.
func (p *AntigravityProvider) DefaultModel() string {
	if p.Model == "" {
		return "gemini-2.5-flash"
	}
	return p.Model
}

// Models returns supported models.
func (p *AntigravityProvider) Models() []string {
	return []string{"gemini-2.5-flash", "gemini-2.5-pro"}
}

// IsHealthy returns whether the provider is healthy.
func (p *AntigravityProvider) IsHealthy() bool { return true }

// Chat sends a chat request.
func (p *AntigravityProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	return nil, fmt.Errorf("AntigravityProvider.Chat not implemented")
}

// ChatStream sends a streaming chat request.
func (p *AntigravityProvider) ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	return nil, fmt.Errorf("AntigravityProvider.ChatStream not implemented")
}
