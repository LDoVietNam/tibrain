package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient is a simple reusable HTTP client for AI provider APIs.
type HTTPClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Headers    map[string]string
}

// NewHTTPClient creates a new HTTP client for an AI provider.
func NewHTTPClient(baseURL, apiKey string, timeout time.Duration) *HTTPClient {
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	return &HTTPClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// SetAuthHeader sets the Authorization header. Override for provider-specific auth.
func (c *HTTPClient) SetAuthHeader(req *http.Request) {
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
}

// PostJSON sends a POST request with JSON body and returns the response body.
func (c *HTTPClient) PostJSON(path string, body any) ([]byte, error) {
	url := c.BaseURL + path
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	c.SetAuthHeader(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// openAIChatRequest is the standard OpenAI-compatible chat request format.
type openAIChatRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Stream   bool            `json:"stream,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *Usage `json:"usage"`
}

// buildOpenAIRequest converts our ChatRequest to OpenAI format.
func buildOpenAIRequest(req ChatRequest) openAIChatRequest {
	msgs := make([]openAIMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = openAIMessage{Role: m.Role, Content: m.Content}
	}
	model := req.Model
	if model == "" {
		model = "gpt-4o-mini"
	}
	return openAIChatRequest{
		Model:    model,
		Messages: msgs,
		Stream:   false,
	}
}

// parseOpenAIResponse parses an OpenAI-compatible response.
func parseOpenAIResponse(body []byte) (*ChatResponse, error) {
	var resp openAIChatResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}
	return &ChatResponse{
		Content:      resp.Choices[0].Message.Content,
		FinishReason: resp.Choices[0].FinishReason,
		Usage:        resp.Usage,
	}, nil
}
