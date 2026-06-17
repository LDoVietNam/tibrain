package management

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// --- OpenAI-compatible embedding types ---

type EmbeddingRequest struct {
	Model          string      `json:"model"`
	Input          interface{} `json:"input"`
	Dimensions     int         `json:"dimensions,omitempty"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
}

type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingItem `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

type EmbeddingItem struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// --- Embedding provider registry ---

type EmbeddingProvider struct {
	ID         string           `json:"id"`
	BaseURL    string           `json:"baseUrl"`
	AuthType   string           `json:"authType"`
	AuthHeader string           `json:"authHeader"`
	Models     []EmbeddingModel `json:"models"`
}

type EmbeddingModel struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Dimensions int    `json:"dimensions"`
}

var embeddingProviders = map[string]EmbeddingProvider{
	"nebius": {
		ID: "nebius", BaseURL: "https://api.tokenfactory.nebius.com/v1/embeddings",
		AuthType: "apikey", AuthHeader: "bearer",
		Models: []EmbeddingModel{{ID: "Qwen/Qwen3-Embedding-8B", Name: "Qwen3 Embedding 8B", Dimensions: 4096}},
	},
	"openai": {
		ID: "openai", BaseURL: "https://api.openai.com/v1/embeddings",
		AuthType: "apikey", AuthHeader: "bearer",
		Models: []EmbeddingModel{
			{ID: "text-embedding-3-small", Name: "Text Embedding 3 Small", Dimensions: 1536},
			{ID: "text-embedding-3-large", Name: "Text Embedding 3 Large", Dimensions: 3072},
			{ID: "text-embedding-ada-002", Name: "Text Embedding Ada 002", Dimensions: 1536},
		},
	},
	"mistral": {
		ID: "mistral", BaseURL: "https://api.mistral.ai/v1/embeddings",
		AuthType: "apikey", AuthHeader: "bearer",
		Models: []EmbeddingModel{{ID: "mistral-embed", Name: "Mistral Embed", Dimensions: 1024}},
	},
	"together": {
		ID: "together", BaseURL: "https://api.together.xyz/v1/embeddings",
		AuthType: "apikey", AuthHeader: "bearer",
		Models: []EmbeddingModel{
			{ID: "BAAI/bge-large-en-v1.5", Name: "BGE Large EN v1.5", Dimensions: 1024},
			{ID: "togethercomputer/m2-bert-80M-8k-retrieval", Name: "M2 BERT 80M 8K", Dimensions: 768},
		},
	},
	"fireworks": {
		ID: "fireworks", BaseURL: "https://api.fireworks.ai/inference/v1/embeddings",
		AuthType: "apikey", AuthHeader: "bearer",
		Models: []EmbeddingModel{{ID: "nomic-ai/nomic-embed-text-v1.5", Name: "Nomic Embed Text v1.5", Dimensions: 768}},
	},
	"nvidia": {
		ID: "nvidia", BaseURL: "https://integrate.api.nvidia.com/v1/embeddings",
		AuthType: "apikey", AuthHeader: "bearer",
		Models: []EmbeddingModel{{ID: "nvidia/nv-embedqa-e5-v5", Name: "NV EmbedQA E5 v5", Dimensions: 1024}},
	},
}

// parseEmbeddingModel parses "provider/model" or bare model ID.
func parseEmbeddingModel(modelStr string) (providerID, modelID string) {
	if modelStr == "" {
		return "", ""
	}
	slashIdx := strings.Index(modelStr, "/")
	if slashIdx > 0 {
		prefix := modelStr[:slashIdx]
		if _, ok := embeddingProviders[prefix]; ok {
			return prefix, modelStr[slashIdx+1:]
		}
		// Handle nested IDs like nebius/Qwen/Qwen3-Embedding-8B
		for providerID := range embeddingProviders {
			if strings.HasPrefix(modelStr, providerID+"/") {
				return providerID, modelStr[len(providerID)+1:]
			}
		}
		return prefix, modelStr[slashIdx+1:]
	}
	// Bare model ID — search all providers
	for providerID, cfg := range embeddingProviders {
		for _, m := range cfg.Models {
			if m.ID == modelStr {
				return providerID, modelStr
			}
		}
	}
	return "", modelStr
}

// getAllEmbeddingModels returns a flat list of all embedding models.
func getAllEmbeddingModels() []map[string]interface{} {
	out := make([]map[string]interface{}, 0)
	for providerID, cfg := range embeddingProviders {
		for _, m := range cfg.Models {
			out = append(out, map[string]interface{}{
				"id":         providerID + "/" + m.ID,
				"object":     "model",
				"created":    time.Now().Unix(),
				"owned_by":   providerID,
				"type":       "embedding",
				"dimensions": m.Dimensions,
			})
		}
	}
	return out
}

func (svc *Service) handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusOK)
		return
	case http.MethodGet:
		svc.listEmbeddingModels(w, r)
	case http.MethodPost:
		svc.createEmbeddings(w, r)
	default:
		writeOpenAIError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request_error")
	}
}

func (svc *Service) listEmbeddingModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   getAllEmbeddingModels(),
	})
}

func (svc *Service) createEmbeddings(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "read body: "+err.Error(), "invalid_request_error")
		return
	}
	defer r.Body.Close()

	var req EmbeddingRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "parse request: "+err.Error(), "invalid_request_error")
		return
	}

	if strings.TrimSpace(req.Model) == "" {
		writeOpenAIError(w, http.StatusBadRequest, "missing model", "invalid_request_error")
		return
	}

	providerID, modelID := parseEmbeddingModel(req.Model)
	if providerID == "" {
		writeOpenAIError(w, http.StatusBadRequest, "invalid embedding model: "+req.Model, "invalid_request_error")
		return
	}

	providerCfg, ok := embeddingProviders[providerID]
	if !ok {
		writeOpenAIError(w, http.StatusBadRequest, "unknown embedding provider: "+providerID, "invalid_request_error")
		return
	}

	// Resolve API key from config
	apiKey := svc.cfg.APIKeys[providerID]
	if apiKey == "" {
		apiKey = svc.cfg.APIKeys["openai"] // fallback
	}
	if apiKey == "" {
		writeOpenAIError(w, http.StatusBadRequest, "no API key configured for provider: "+providerID, "invalid_request_error")
		return
	}

	// Build upstream request
	upstreamBody := map[string]interface{}{
		"model": modelID,
		"input": req.Input,
	}
	if req.Dimensions > 0 {
		upstreamBody["dimensions"] = req.Dimensions
	}
	if req.EncodingFormat != "" {
		upstreamBody["encoding_format"] = req.EncodingFormat
	}
	payload, _ := json.Marshal(upstreamBody)

	// Build headers
	upReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, providerCfg.BaseURL, bytes.NewReader(payload))
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "build request: "+err.Error(), "server_error")
		return
	}
	upReq.Header.Set("Content-Type", "application/json")
	if providerCfg.AuthHeader == "bearer" {
		upReq.Header.Set("Authorization", "Bearer "+apiKey)
	} else if providerCfg.AuthHeader == "x-api-key" {
		upReq.Header.Set("x-api-key", apiKey)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(upReq)
	if err != nil {
		writeOpenAIError(w, http.StatusBadGateway, "upstream error: "+err.Error(), "server_error")
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		writeOpenAIError(w, http.StatusBadGateway, "read upstream: "+err.Error(), "server_error")
		return
	}

	if resp.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(respBody)
		return
	}

	// Normalize response to OpenAI format
	var upstreamResp map[string]interface{}
	_ = json.Unmarshal(respBody, &upstreamResp)

	// Ensure usage exists
	usage, _ := upstreamResp["usage"].(map[string]interface{})
	if usage == nil {
		usage = map[string]interface{}{"prompt_tokens": 0, "total_tokens": 0}
	}

	result := map[string]interface{}{
		"object": "list",
		"data":   upstreamResp["data"],
		"model":  req.Model,
		"usage":  usage,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
