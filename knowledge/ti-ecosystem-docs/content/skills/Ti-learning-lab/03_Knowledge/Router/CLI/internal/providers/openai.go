package providers

import (
	"context"
	"fmt"
	"os"
)

// OpenAIProvider implements Provider for OpenAI API.
type OpenAIProvider struct {
	client       *HTTPClient
	defaultModel string
	models       []string
}

// NewOpenAIProvider creates an OpenAI provider from API key.
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &OpenAIProvider{
		client:       NewHTTPClient("https://api.openai.com/v1", apiKey, 0),
		defaultModel: "gpt-4o-mini",
		models: []string{
			"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4",
			"o1-mini", "o1-preview", "o3-mini",
		},
	}
}

func (p *OpenAIProvider) Name() string         { return "openai" }
func (p *OpenAIProvider) DefaultModel() string { return p.defaultModel }
func (p *OpenAIProvider) Models() []string     { return append([]string(nil), p.models...) }

func (p *OpenAIProvider) IsHealthy() bool {
	return p.client.APIKey != ""
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	body, err := p.client.PostJSON("/chat/completions", buildOpenAIRequest(req))
	if err != nil {
		return nil, fmt.Errorf("openai chat: %w", err)
	}
	return parseOpenAIResponse(body)
}

func (p *OpenAIProvider) ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	// For now, delegate to non-streaming
	resp, err := p.Chat(ctx, req)
	if err == nil && onChunk != nil {
		onChunk(StreamChunk{Content: resp.Content, Done: true})
	}
	return resp, err
}

// detectOpenAI checks if OpenAI API key is available.
func detectOpenAI() Provider {
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return NewOpenAIProvider(key)
	}
	return nil
}
