package management

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// --- OpenAI-compatible image generation types ---

type ImageGenerationRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	Style          string `json:"style,omitempty"`
}

type ImageGenerationResponse struct {
	Created int64       `json:"created"`
	Data    []ImageItem `json:"data"`
}

type ImageItem struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// --- Image provider registry ---

type ImageProvider struct {
	ID         string `json:"id"`
	BaseURL    string `json:"baseUrl"`
	AuthType   string `json:"authType"`
	AuthHeader string `json:"authHeader"`
}

var imageProviders = map[string]ImageProvider{
	"openai":     {ID: "openai", BaseURL: "https://api.openai.com/v1/images/generations", AuthType: "apikey", AuthHeader: "bearer"},
	"hyperbolic": {ID: "hyperbolic", BaseURL: "https://api.hyperbolic.xyz/v1/image/generation", AuthType: "apikey", AuthHeader: "bearer"},
	"nebius":     {ID: "nebius", BaseURL: "https://api.studio.nebius.ai/v1/images/generations", AuthType: "apikey", AuthHeader: "bearer"},
	"together":   {ID: "together", BaseURL: "https://api.together.xyz/v1/images/generations", AuthType: "apikey", AuthHeader: "bearer"},
}

func parseImageModel(modelStr string) (providerID, modelID string) {
	if modelStr == "" {
		return "", ""
	}
	slashIdx := strings.Index(modelStr, "/")
	if slashIdx > 0 {
		prefix := modelStr[:slashIdx]
		if _, ok := imageProviders[prefix]; ok {
			return prefix, modelStr[slashIdx+1:]
		}
	}
	return "openai", modelStr
}

func (svc *Service) handleImageGenerations(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		writeOpenAIError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request_error")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "read body: "+err.Error(), "invalid_request_error")
		return
	}
	defer r.Body.Close()

	var req ImageGenerationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "parse request: "+err.Error(), "invalid_request_error")
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		writeOpenAIError(w, http.StatusBadRequest, "missing prompt", "invalid_request_error")
		return
	}

	providerID, modelID := parseImageModel(req.Model)
	providerCfg, ok := imageProviders[providerID]
	if !ok {
		writeOpenAIError(w, http.StatusBadRequest, "unknown image provider: "+providerID, "invalid_request_error")
		return
	}

	apiKey := svc.cfg.APIKeys[providerID]
	if apiKey == "" {
		apiKey = svc.cfg.APIKeys["openai"]
	}
	if apiKey == "" {
		writeOpenAIError(w, http.StatusBadRequest, "no API key configured for provider: "+providerID, "invalid_request_error")
		return
	}

	upstreamBody := map[string]interface{}{
		"model":  modelID,
		"prompt": req.Prompt,
	}
	if req.N > 0 {
		upstreamBody["n"] = req.N
	}
	if req.Size != "" {
		upstreamBody["size"] = req.Size
	}
	if req.Quality != "" {
		upstreamBody["quality"] = req.Quality
	}
	if req.ResponseFormat != "" {
		upstreamBody["response_format"] = req.ResponseFormat
	}
	if req.Style != "" {
		upstreamBody["style"] = req.Style
	}
	payload, _ := json.Marshal(upstreamBody)

	upReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, providerCfg.BaseURL, bytes.NewReader(payload))
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "build request: "+err.Error(), "server_error")
		return
	}
	upReq.Header.Set("Content-Type", "application/json")
	if providerCfg.AuthHeader == "bearer" {
		upReq.Header.Set("Authorization", "Bearer "+apiKey)
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

	var upstreamResp map[string]interface{}
	_ = json.Unmarshal(respBody, &upstreamResp)

	// Normalize to OpenAI format
	result := ImageGenerationResponse{
		Created: time.Now().Unix(),
		Data:    []ImageItem{},
	}
	if data, ok := upstreamResp["data"].([]interface{}); ok {
		for _, item := range data {
			if m, ok := item.(map[string]interface{}); ok {
				item := ImageItem{}
				if v, ok := m["url"].(string); ok {
					item.URL = v
				}
				if v, ok := m["b64_json"].(string); ok {
					item.B64JSON = v
				}
				if v, ok := m["revised_prompt"].(string); ok {
					item.RevisedPrompt = v
				}
				result.Data = append(result.Data, item)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
