package providers

import (
	"context"
	"fmt"
	"os"
)

// OpenRouterProvider implements Provider for OpenRouter API (unified access to many models).
type OpenRouterProvider struct {
	client       *HTTPClient
	defaultModel string
	models       []string
}

// NewOpenRouterProvider creates an OpenRouter provider from API key.
func NewOpenRouterProvider(apiKey string) *OpenRouterProvider {
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}
	return &OpenRouterProvider{
		client:       NewHTTPClient("https://openrouter.ai/api/v1", apiKey, 0),
		defaultModel: "openai/gpt-4o-mini",
		models: []string{
			"openai/gpt-4o", "openai/gpt-4o-mini", "openai/o3-mini",
			"anthropic/claude-3.5-sonnet", "anthropic/claude-3.5-haiku",
			"google/gemini-2.0-flash", "google/gemini-2.0-pro",
			"deepseek/deepseek-chat", "deepseek/deepseek-r1",
			"x-ai/grok-2", "meta-llama/llama-3.3-70b",
		},
	}
}

func (p *OpenRouterProvider) Name() string         { return "openrouter" }
func (p *OpenRouterProvider) DefaultModel() string { return p.defaultModel }
func (p *OpenRouterProvider) Models() []string     { return append([]string(nil), p.models...) }

func (p *OpenRouterProvider) IsHealthy() bool {
	return p.client.APIKey != ""
}

func (p *OpenRouterProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	body, err := p.client.PostJSON("/chat/completions", buildOpenAIRequest(req))
	if err != nil {
		return nil, fmt.Errorf("openrouter chat: %w", err)
	}
	return parseOpenAIResponse(body)
}

func (p *OpenRouterProvider) ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	resp, err := p.Chat(ctx, req)
	if err == nil && onChunk != nil {
		onChunk(StreamChunk{Content: resp.Content, Done: true})
	}
	return resp, err
}

// detectOpenRouter checks if OpenRouter API key is available.
func detectOpenRouter() Provider {
	if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
		return NewOpenRouterProvider(key)
	}
	return nil
}
