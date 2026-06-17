package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// RouterProvider connects to the Ti Router's OpenAI-compatible API.
type RouterProvider struct {
	baseURL      string
	apiKey       string
	httpClient   *http.Client
	defaultModel string
	models       []string
}

// NewRouterProvider creates a provider that connects to Ti Router.
func NewRouterProvider(baseURL, apiKey string) *RouterProvider {
	if baseURL == "" {
		baseURL = os.Getenv("TI_ROUTER_URL")
		if baseURL == "" {
			baseURL = "http://localhost:1806"
		}
	}
	if apiKey == "" {
		apiKey = os.Getenv("TI_ROUTER_API_KEY")
	}
	return &RouterProvider{
		baseURL:      baseURL,
		apiKey:       apiKey,
		httpClient:   &http.Client{Timeout: 120 * time.Second},
		defaultModel: "auto",
		models: []string{
			"auto", "gpt-4o", "gpt-4o-mini", "claude-3-5-sonnet",
			"combo-fast", "combo-quality", "combo-cheap",
		},
	}
}

func (p *RouterProvider) Name() string         { return "router" }
func (p *RouterProvider) DefaultModel() string { return p.defaultModel }
func (p *RouterProvider) Models() []string     { return append([]string(nil), p.models...) }

func (p *RouterProvider) IsHealthy() bool {
	resp, err := p.httpClient.Get(p.baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (p *RouterProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	body, err := p.postJSON(ctx, "/v1/chat/completions", buildOpenAIRequest(req))
	if err != nil {
		return nil, fmt.Errorf("router chat: %w", err)
	}
	return parseOpenAIResponse(body)
}

func (p *RouterProvider) ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	resp, err := p.Chat(ctx, req)
	if err == nil && onChunk != nil {
		onChunk(StreamChunk{Content: resp.Content, Done: true})
	}
	return resp, err
}

func (p *RouterProvider) postJSON(ctx context.Context, path string, payload any) ([]byte, error) {
	url := p.baseURL + path
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("router http %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// detectRouter checks if Ti Router is available.
func detectRouter() Provider {
	url := os.Getenv("TI_ROUTER_URL")
	if url == "" {
		url = "http://localhost:1806"
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url + "/health")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return NewRouterProvider(url, "")
	}
	return nil
}
