package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type EmbeddingGenerator struct {
	baseURL   string
	model     string
	dimension int
	client    *http.Client
}

func NewEmbeddingGenerator(baseURL, model string, dimension int) *EmbeddingGenerator {
	return NewEmbeddingGeneratorWithTimeout(baseURL, model, dimension, 30*time.Second)
}

func NewEmbeddingGeneratorWithTimeout(baseURL, model string, dimension int, timeout time.Duration) *EmbeddingGenerator {
	return &EmbeddingGenerator{
		baseURL:   baseURL,
		model:     model,
		dimension: dimension,
		client:    &http.Client{Timeout: timeout},
	}
}

// GenerateEmbedding fetches the embedding for a single text.
func (eg *EmbeddingGenerator) GenerateEmbedding(text string) ([]float32, error) {
	resp, err := eg.callEmbeddingAPI([]string{text})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}
	return resp.Data[0].Embedding, nil
}

// GenerateEmbeddings fetches embeddings for many texts in a single API call
// (when supported by the upstream: OpenAI, Ollama, vLLM, etc. all accept
// `input: [text1, text2, ...]` and return embeddings in the same order).
//
// This is the right entry point for any RAG batch — one HTTP round-trip
// instead of N. With the previous loop-based implementation, 1000 chunks
// required 1000 sequential requests; now it's 1.
func (eg *EmbeddingGenerator) GenerateEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	resp, err := eg.callEmbeddingAPI(texts)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) != len(texts) {
		return nil, fmt.Errorf("embedding count mismatch: got %d, want %d", len(resp.Data), len(texts))
	}
	out := make([][]float32, len(texts))
	for i, d := range resp.Data {
		out[i] = d.Embedding
	}
	return out, nil
}

// embeddingAPIResponse is the shared shape of the response from any
// OpenAI-compatible embeddings endpoint.
type embeddingAPIResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (eg *EmbeddingGenerator) callEmbeddingAPI(texts []string) (*embeddingAPIResponse, error) {
	payload := map[string]interface{}{
		"model": eg.model,
		"input": texts,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	resp, err := eg.client.Post(eg.baseURL+"/embeddings", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read embedding response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("embedding API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result embeddingAPIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse embedding response: %w", err)
	}
	return &result, nil
}
