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

func (eg *EmbeddingGenerator) GenerateEmbedding(text string) ([]float32, error) {
	payload := map[string]interface{}{
		"model": eg.model,
		"input": []string{text},
	}
	body, _ := json.Marshal(payload)

	resp, err := eg.client.Post(eg.baseURL+"/embeddings", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse embedding response: %w", err)
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}
	return result.Data[0].Embedding, nil
}

func (eg *EmbeddingGenerator) GenerateEmbeddings(texts []string) ([][]float32, error) {
	var results [][]float32
	for _, text := range texts {
		emb, err := eg.GenerateEmbedding(text)
		if err != nil {
			return nil, err
		}
		results = append(results, emb)
	}
	return results, nil
}
