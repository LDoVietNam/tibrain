package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// CodebuffProvider implements Provider for Codebuff API.
type CodebuffProvider struct {
	apiKeys      []string
	keyIndex     int
	defaultModel string
	models       []string
	httpClient   *http.Client
	cwd          string
}

// codebuffRequest is the payload sent to Codebuff API.
type codebuffRequest struct {
	Agent   string            `json:"agent"`
	Prompt  string            `json:"prompt"`
	Context map[string]string `json:"context,omitempty"`
}

// codebuffResponse is the response from Codebuff API.
type codebuffResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

const codebuffAPIURL = "https://www.codebuff.com/api/run"

// NewCodebuffProvider creates a Codebuff provider from API key with multi-key support.
func NewCodebuffProvider(apiKey string) *CodebuffProvider {
	keys := collectCodebuffKeys(apiKey)
	cwd, _ := os.Getwd()
	return &CodebuffProvider{
		apiKeys:      keys,
		keyIndex:     0,
		defaultModel: "codebuff/base@latest",
		models:       []string{"codebuff/base@latest", "codebuff/base"},
		httpClient:   &http.Client{Timeout: 120 * time.Second},
		cwd:          cwd,
	}
}

// collectCodebuffKeys collects all available CODEBUFF_API_KEY variants from env.
func collectCodebuffKeys(primary string) []string {
	seen := map[string]bool{}
	var keys []string

	addKey := func(k string) {
		if k != "" && !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}

	addKey(primary)
	addKey(os.Getenv("CODEBUFF_API_KEY"))
	addKey(os.Getenv("CODEBUFF_API_KEY_2"))
	addKey(os.Getenv("CODEBUFF_API_KEY_3"))
	addKey(os.Getenv("CODEBUFF_API_KEY_4"))
	addKey(os.Getenv("CODEBUFF_API_KEY_5"))

	return keys
}

func (p *CodebuffProvider) currentKey() string {
	if len(p.apiKeys) == 0 {
		return ""
	}
	return p.apiKeys[p.keyIndex%len(p.apiKeys)]
}

func (p *CodebuffProvider) Name() string         { return "codebuff" }
func (p *CodebuffProvider) DefaultModel() string { return p.defaultModel }
func (p *CodebuffProvider) Models() []string     { return append([]string(nil), p.models...) }

func (p *CodebuffProvider) IsHealthy() bool {
	return len(p.apiKeys) > 0
}

func (p *CodebuffProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if len(p.apiKeys) == 0 {
		return nil, fmt.Errorf("codebuff: API key not set (set CODEBUFF_API_KEY env var)")
	}

	// Read local files from context
	contextMap := make(map[string]string)
	if targetFiles, ok := req.Options["files"].([]string); ok {
		for _, file := range targetFiles {
			fullPath := filepath.Join(p.cwd, file)
			content, err := os.ReadFile(fullPath)
			if err == nil {
				contextMap[file] = string(content)
			}
		}
	}

	// Build prompt from last user message
	prompt := ""
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			prompt = req.Messages[i].Content
			break
		}
	}
	if prompt == "" && len(req.Messages) > 0 {
		prompt = req.Messages[len(req.Messages)-1].Content
	}

	payload := codebuffRequest{
		Agent:   req.Model,
		Prompt:  prompt,
		Context: contextMap,
	}
	if payload.Agent == "" {
		payload.Agent = p.defaultModel
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("codebuff: marshal request: %w", err)
	}

	// Try keys with fallback
	var lastErr error
	startIdx := rand.Intn(len(p.apiKeys))
	for i := 0; i < len(p.apiKeys); i++ {
		idx := (startIdx + i) % len(p.apiKeys)
		key := p.apiKeys[idx]

		httpReq, err := http.NewRequestWithContext(ctx, "POST", codebuffAPIURL, bytes.NewReader(jsonPayload))
		if err != nil {
			lastErr = fmt.Errorf("codebuff: create request: %w", err)
			continue
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+key)

		resp, err := p.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("codebuff: http request (key %d): %w", idx+1, err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("codebuff: read response: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusUnauthorized {
			lastErr = fmt.Errorf("codebuff: http %d (key %d)", resp.StatusCode, idx+1)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("codebuff: http %d: %s", resp.StatusCode, string(body))
		}

		var responseData codebuffResponse
		if err := json.Unmarshal(body, &responseData); err != nil {
			return nil, fmt.Errorf("codebuff: unmarshal response: %w", err)
		}

		if responseData.Error != "" {
			lastErr = fmt.Errorf("codebuff: backend error (key %d): %s", idx+1, responseData.Error)
			continue
		}

		p.keyIndex = idx
		return &ChatResponse{
			Content:      responseData.Output,
			FinishReason: "stop",
		}, nil
	}

	return nil, lastErr
}

func (p *CodebuffProvider) ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	resp, err := p.Chat(ctx, req)
	if err == nil && onChunk != nil {
		onChunk(StreamChunk{Content: resp.Content, Done: true})
	}
	return resp, err
}

// DetectCodebuff checks if any CODEBUFF_API_KEY environment variable is set,
// or falls back to reading credentials from local config.
func DetectCodebuff() Provider {
	keys := collectCodebuffKeys("")
	if len(keys) > 0 {
		return NewCodebuffProvider(keys[0])
	}

	// Fallback to local credentials file
	creds, err := GetUserCredentials()
	if err == nil && creds != nil && creds.AuthToken != "" {
		return NewCodebuffProvider(creds.AuthToken)
	}
	return nil
}
