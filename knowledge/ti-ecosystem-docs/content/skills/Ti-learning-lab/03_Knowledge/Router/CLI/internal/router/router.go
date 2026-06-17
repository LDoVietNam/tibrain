package router

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Model struct {
	ID      string `json:"id"`
	OwnedBy string `json:"owned_by"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type ChatChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	TotalTime        float64 `json:"total_time"`
}

type ChatResponse struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   Usage        `json:"usage"`
}

type StreamChunk struct {
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason"`
}

type PingResult struct {
	Model      string        `json:"model"`
	Rounds     int           `json:"rounds"`
	AvgLatency time.Duration `json:"avg_latency"`
	MinLatency time.Duration `json:"min_latency"`
	MaxLatency time.Duration `json:"max_latency"`
}

type AuthFileInfo struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type RoutingStrategy string

const (
	RoundRobin RoutingStrategy = "round-robin"
	FillFirst  RoutingStrategy = "fill-first"
)

type APIError struct {
	Method     string
	URL        string
	StatusCode int
	Body       string
	ModelName  string
}

func (e *APIError) Error() string {
	if e.ModelName != "" {
		return fmt.Sprintf("API error %d (model=%q, %s %s): %s", e.StatusCode, e.ModelName, e.Method, e.URL, e.Body)
	}
	return fmt.Sprintf("API error %d (%s %s): %s", e.StatusCode, e.Method, e.URL, e.Body)
}

type Client struct {
	mu            sync.RWMutex
	baseURL       string
	apiKey        string
	managementKey string
	managementURL string
	httpClient    *http.Client
	retryMax      int
	retryBackoff  time.Duration
	userAgent     string
	tiMeta        map[string]string
}

type ClientOption func(*Client)

func WithAPIKey(key string) ClientOption        { return func(c *Client) { c.apiKey = key } }
func WithManagementKey(key string) ClientOption { return func(c *Client) { c.managementKey = key } }
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) { c.httpClient = hc }
}
func WithRetryPolicy(max int, backoff time.Duration) ClientOption {
	return func(c *Client) {
		if max >= 0 {
			c.retryMax = max
		}
		if backoff > 0 {
			c.retryBackoff = backoff
		}
	}
}
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if timeout > 0 && c.httpClient != nil {
			c.httpClient.Timeout = timeout
		}
	}
}
func WithUserAgent(ua string) ClientOption { return func(c *Client) { c.userAgent = ua } }

func New(baseURL string, opts ...ClientOption) *Client {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:1810"
	}
	c := &Client{
		baseURL:       strings.TrimSuffix(baseURL, "/"),
		managementURL: strings.TrimSuffix(baseURL, "/"),
		httpClient:    &http.Client{Timeout: 300 * time.Second},
		retryMax:      2,
		retryBackoff:  350 * time.Millisecond,
		userAgent:     "TiCLI/0.2",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) SetBaseURL(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.baseURL = strings.TrimSuffix(url, "/")
	if c.managementURL == "" {
		c.managementURL = c.baseURL
	}
}

func (c *Client) SetManagementURL(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.managementURL = strings.TrimSuffix(url, "/")
}

func (c *Client) GetModels() ([]Model, error) {
	resp, err := c.doRequest("GET", "/v1/models", nil, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, readAPIError(resp, "GET", "/v1/models", "")
	}
	var result struct {
		Data []Model `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}
	return result.Data, nil
}

func (c *Client) Chat(model string, messages []Message, opts ...ChatOption) (*ChatResponse, error) {
	req := &ChatRequest{Model: model, Messages: messages}
	for _, opt := range opts {
		opt(req)
	}
	req.Stream = false
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chat request: %w", err)
	}
	resp, err := c.doRequest("POST", "/v1/chat/completions", bytes.NewBuffer(body), false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, readAPIError(resp, "POST", "/v1/chat/completions", model)
	}
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode chat response: %w", err)
	}
	return &chatResp, nil
}

func (c *Client) ChatStream(model string, messages []Message, onChunk func(string), opts ...ChatOption) (*ChatResponse, error) {
	req := &ChatRequest{Model: model, Messages: messages, Stream: true}
	for _, opt := range opts {
		opt(req)
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stream request: %w", err)
	}
	resp, err := c.doRequest("POST", "/v1/chat/completions", bytes.NewBuffer(body), false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, readAPIError(resp, "POST", "/v1/chat/completions", model)
	}
	return parseStreamResponse(resp.Body, model, onChunk)
}

func parseStreamResponse(body io.Reader, model string, onChunk func(string)) (*ChatResponse, error) {
	var fullContent strings.Builder
	var finalID, finalModel, finalRole, finishReason string
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" {
			continue
		}
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			ID      string `json:"id"`
			Model   string `json:"model"`
			Choices []struct {
				Index int `json:"index"`
				Delta struct {
					Content string `json:"content"`
					Role    string `json:"role"`
				} `json:"delta"`
				Message struct {
					Content string `json:"content"`
					Role    string `json:"role"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.ID != "" {
			finalID = chunk.ID
		}
		if chunk.Model != "" {
			finalModel = chunk.Model
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		choice := chunk.Choices[0]
		if choice.Delta.Role != "" {
			finalRole = choice.Delta.Role
		}
		if choice.Message.Role != "" {
			finalRole = choice.Message.Role
		}
		if choice.Delta.Content != "" {
			fullContent.WriteString(choice.Delta.Content)
			if onChunk != nil {
				onChunk(choice.Delta.Content)
			}
		}
		if choice.Message.Content != "" {
			fullContent.WriteString(choice.Message.Content)
			if onChunk != nil {
				onChunk(choice.Message.Content)
			}
		}
		if choice.FinishReason != "" {
			finishReason = choice.FinishReason
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("stream read error: %w", err)
	}
	if fullContent.Len() == 0 && finishReason == "" {
		return nil, fmt.Errorf("stream completed without final response (model=%q)", model)
	}
	if finalModel == "" {
		finalModel = model
	}
	if finalRole == "" {
		finalRole = "assistant"
	}
	return &ChatResponse{ID: finalID, Model: finalModel, Choices: []ChatChoice{{Index: 0, Message: Message{Role: finalRole, Content: fullContent.String()}, FinishReason: finishReason}}}, nil
}

type ChatOption func(*ChatRequest)

func WithMaxTokens(n int) ChatOption       { return func(r *ChatRequest) { r.MaxTokens = n } }
func WithTemperature(t float64) ChatOption { return func(r *ChatRequest) { r.Temperature = t } }

func (c *Client) Health() error {
	resp, err := c.doRequest("GET", "/v1/models", nil, false)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: server returned status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Ping(model string) (*PingResult, error) {
	const rounds = 3
	latencies := make([]time.Duration, 0, rounds)
	for i := 0; i < rounds; i++ {
		start := time.Now()
		_, err := c.Chat(model, []Message{{Role: "user", Content: "hi"}}, WithMaxTokens(5))
		if err != nil {
			return nil, fmt.Errorf("ping failed round %d: %w", i+1, err)
		}
		latencies = append(latencies, time.Since(start))
	}
	var totalNs, minNs, maxNs int64
	for i, l := range latencies {
		ns := l.Nanoseconds()
		totalNs += ns
		if i == 0 || ns < minNs {
			minNs = ns
		}
		if i == 0 || ns > maxNs {
			maxNs = ns
		}
	}
	return &PingResult{Model: model, Rounds: rounds, AvgLatency: time.Duration(totalNs / rounds), MinLatency: time.Duration(minNs), MaxLatency: time.Duration(maxNs)}, nil
}

func (c *Client) GetAuthFiles() ([]AuthFileInfo, error) {
	resp, err := c.doRequest("GET", "/v0/management/auth-files", nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, readAPIError(resp, "GET", "/v0/management/auth-files", "")
	}
	var files []AuthFileInfo
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to decode auth files response: %w", err)
	}
	return files, nil
}

func (c *Client) UploadAuthFile(name string, data []byte) error {
	payload, err := json.Marshal(map[string]any{"name": name, "data": data})
	if err != nil {
		return fmt.Errorf("failed to marshal upload request: %w", err)
	}
	resp, err := c.doRequest("PUT", "/v0/management/auth-files/"+name, bytes.NewBuffer(payload), true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return readAPIError(resp, "PUT", "/v0/management/auth-files/"+name, "")
	}
	return nil
}

func (c *Client) DeleteAuthFile(name string) error {
	resp, err := c.doRequest("DELETE", "/v0/management/auth-files/"+name, nil, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return readAPIError(resp, "DELETE", "/v0/management/auth-files/"+name, "")
	}
	return nil
}

func (c *Client) GetConfig() ([]byte, error) {
	resp, err := c.doRequest("GET", "/v0/management/config", nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, readAPIError(resp, "GET", "/v0/management/config", "")
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) SetRoutingStrategy(strategy RoutingStrategy) error {
	if strategy != RoundRobin && strategy != FillFirst {
		return fmt.Errorf("invalid routing strategy %q (must be %q or %q)", strategy, RoundRobin, FillFirst)
	}
	payload, err := json.Marshal(map[string]any{"routing_strategy": string(strategy)})
	if err != nil {
		return fmt.Errorf("failed to marshal routing strategy request: %w", err)
	}
	resp, err := c.doRequest("PUT", "/v0/management/config", bytes.NewBuffer(payload), true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return readAPIError(resp, "PUT", "/v0/management/config", "")
	}
	return nil
}

func (c *Client) ShareModel(modelID, provider string, metadata map[string]interface{}) error {
	body := map[string]interface{}{"model_id": modelID, "provider": provider}
	if metadata != nil {
		body["metadata"] = metadata
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal share request: %w", err)
	}
	resp, err := c.doRequest("POST", "/v0/management/models/share", bytes.NewBuffer(payload), true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return readAPIError(resp, "POST", "/v0/management/models/share", "")
	}
	return nil
}

func (c *Client) FetchModel(modelID string) (map[string]interface{}, error) {
	resp, err := c.doRequest("GET", "/v0/management/models/fetch/"+modelID, nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, readAPIError(resp, "GET", "/v0/management/models/fetch/"+modelID, "")
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode fetch model response: %w", err)
	}
	return result, nil
}

func (c *Client) ListSharedModels() ([]map[string]interface{}, error) {
	resp, err := c.doRequest("GET", "/v0/management/models/shared", nil, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, readAPIError(resp, "GET", "/v0/management/models/shared", "")
	}
	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode shared models response: %w", err)
	}
	return result, nil
}

func (c *Client) doRequest(method, path string, body io.Reader, management bool) (*http.Response, error) {
	var payload []byte
	if body != nil {
		var err error
		payload, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body %s %s: %w", method, path, err)
		}
	}
	c.mu.RLock()
	baseURL := c.baseURL
	if management && c.managementURL != "" {
		baseURL = c.managementURL
	}
	apiKey, mgmtKey := c.apiKey, c.managementKey
	httpClient, retryMax, retryBackoff, userAgent := c.httpClient, c.retryMax, c.retryBackoff, c.userAgent
	c.mu.RUnlock()
	url := baseURL + path
	attempts := retryMax + 1
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		var reqBody io.Reader
		if payload != nil {
			reqBody = bytes.NewReader(payload)
		}
		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request %s %s: %w", method, path, err)
		}
		req.Header.Set("Content-Type", "application/json")
		if userAgent != "" {
			req.Header.Set("User-Agent", userAgent)
		}
		if management {
			if mgmtKey != "" {
				req.Header.Set("X-Management-Key", mgmtKey)
			}
		} else if apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
		resp, err := httpClient.Do(req)
		if err == nil {
			if !shouldRetryStatus(resp.StatusCode) || attempt == attempts-1 {
				return resp, nil
			}
			// Drain and close body before retry — next attempt gets fresh response
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			lastErr = &APIError{Method: method, URL: path, StatusCode: resp.StatusCode, Body: "transient error"}
		} else {
			lastErr = fmt.Errorf("request failed %s %s: %w", method, path, err)
		}
		if attempt < attempts-1 && retryBackoff > 0 {
			time.Sleep(retryBackoff * time.Duration(attempt+1))
		}
	}
	return nil, lastErr
}

func shouldRetryStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
}

// readAPIError reads error body from response to create structured API error.
// Returns generic error if body is already consumed/closed.
func readAPIError(resp *http.Response, method, path, model string) error {
	if resp == nil || resp.Body == nil {
		return &APIError{Method: method, URL: path, StatusCode: 0, Body: "", ModelName: model}
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{Method: method, URL: path, StatusCode: resp.StatusCode, Body: fmt.Sprintf("(error reading body: %v)", err), ModelName: model}
	}
	return &APIError{Method: method, URL: path, StatusCode: resp.StatusCode, Body: string(respBody), ModelName: model}
}
