package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// AnthropicProvider implements Provider for Anthropic Claude API.
type AnthropicProvider struct {
	client       *HTTPClient
	defaultModel string
	models       []string
}

// NewAnthropicProvider creates an Anthropic provider from API key.
func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	return &AnthropicProvider{
		client: &HTTPClient{
			BaseURL: "https://api.anthropic.com/v1",
			APIKey:  apiKey,
			HTTPClient: &http.Client{
				Timeout: 120 * 1000000000, // 120 seconds
			},
			Headers: map[string]string{
				"Content-Type":      "application/json",
				"x-api-key":         apiKey,
				"anthropic-version": "2023-06-01",
			},
		},
		defaultModel: "claude-3-5-sonnet-20241022",
		models: []string{
			"claude-3-5-sonnet-20241022",
			"claude-3-5-haiku-20241022",
			"claude-3-opus-20240229",
			"claude-3-sonnet-20240229",
			"claude-3-haiku-20240307",
		},
	}
}

func (p *AnthropicProvider) Name() string         { return "anthropic" }
func (p *AnthropicProvider) DefaultModel() string { return p.defaultModel }
func (p *AnthropicProvider) Models() []string     { return append([]string(nil), p.models...) }

func (p *AnthropicProvider) IsHealthy() bool {
	return p.client.APIKey != ""
}

// anthropicChatRequest is the Claude API request format.
type anthropicChatRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicChatResponse is the Claude API response format.
type anthropicChatResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	StopReason string `json:"stop_reason"`
}

func (p *AnthropicProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	msgs := make([]anthropicMessage, len(req.Messages))
	for i, m := range req.Messages {
		role := m.Role
		if role == "system" {
			role = "user" // Claude doesn't have system role in messages
		}
		msgs[i] = anthropicMessage{Role: role, Content: m.Content}
	}

	model := req.Model
	if model == "" {
		model = p.defaultModel
	}

	body, err := p.client.PostJSON("/messages", anthropicChatRequest{
		Model:     model,
		Messages:  msgs,
		MaxTokens: 4096,
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic chat: %w", err)
	}

	var resp anthropicChatResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal anthropic response: %w", err)
	}

	var content string
	for _, c := range resp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &ChatResponse{
		Content:      content,
		FinishReason: resp.StopReason,
		Usage: &Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}, nil
}

func (p *AnthropicProvider) ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	resp, err := p.Chat(ctx, req)
	if err == nil && onChunk != nil {
		onChunk(StreamChunk{Content: resp.Content, Done: true})
	}
	return resp, err
}

// detectAnthropic checks if Anthropic API key is available.
func detectAnthropic() Provider {
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return NewAnthropicProvider(key)
	}
	return nil
}
