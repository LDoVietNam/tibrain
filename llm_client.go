package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LLMClient struct {
	baseURL    string
	model      string
	client     *http.Client
	compressor *LocalRTKCompressor
}

func NewLLMClient(baseURL, model string) *LLMClient {
	return NewLLMClientWithTimeout(baseURL, model, 60*time.Second)
}

func NewLLMClientWithTimeout(baseURL, model string, timeout time.Duration) *LLMClient {
	return &LLMClient{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: timeout},
	}
}

// SetCompressor registers the LocalRTKCompressor for token and context optimization
func (c *LLMClient) SetCompressor(comp *LocalRTKCompressor) {
	c.compressor = comp
}

func (c *LLMClient) GenerateAnswer(ctx context.Context, query, context string) (string, error) {
	// Apply global/comprehensive RTK compression to context if compressor is set
	if c.compressor != nil {
		strategy := c.compressor.StrategyForModel(c.model)
		if comp, err := c.compressor.CompressText(context, strategy.MaxInputChars); err == nil {
			logger.Info("RTK global RAG prompt context compressed: %d -> %d chars using strategy (MaxInputChars=%d)",
				len(context), len(comp), strategy.MaxInputChars)
			context = comp
		}
	}

	payload := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": "Bạn là trợ lý AI cho hệ sinh thái Ti. Dựa trên context được cung cấp, hãy trả lời câu hỏi một cách chính xác và ngắn gọn. Nếu không có thông tin, hãy nói rõ."},
			{"role": "user", "content": fmt.Sprintf("Context:\n%s\n\nCâu hỏi: %s", context, query)},
		},
		"temperature": 0.3,
		"max_tokens":  1024,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse llm response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from llm")
	}
	return result.Choices[0].Message.Content, nil
}
