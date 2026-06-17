package management

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// --- Rerank types ---

type RerankRequest struct {
	Model           string        `json:"model"`
	Query           string        `json:"query"`
	Documents       []interface{} `json:"documents"`
	TopN            int           `json:"top_n,omitempty"`
	ReturnDocuments bool          `json:"return_documents,omitempty"`
}

type RerankProvider struct {
	ID         string `json:"id"`
	BaseURL    string `json:"baseUrl"`
	AuthType   string `json:"authType"`
	AuthHeader string `json:"authHeader"`
	Format     string `json:"format,omitempty"`
}

var rerankProviders = map[string]RerankProvider{
	"cohere":    {ID: "cohere", BaseURL: "https://api.cohere.com/v2/rerank", AuthType: "apikey", AuthHeader: "bearer"},
	"together":  {ID: "together", BaseURL: "https://api.together.xyz/v1/rerank", AuthType: "apikey", AuthHeader: "bearer"},
	"nvidia":    {ID: "nvidia", BaseURL: "https://integrate.api.nvidia.com/v1/ranking", AuthType: "apikey", AuthHeader: "bearer", Format: "nvidia"},
	"fireworks": {ID: "fireworks", BaseURL: "https://api.fireworks.ai/inference/v1/rerank", AuthType: "apikey", AuthHeader: "bearer"},
}

func parseRerankModel(modelStr string) (providerID, modelID string) {
	if modelStr == "" {
		return "", ""
	}
	for providerID := range rerankProviders {
		if strings.HasPrefix(modelStr, providerID+"/") {
			return providerID, modelStr[len(providerID)+1:]
		}
	}
	for providerID, cfg := range rerankProviders {
		if providerID == modelStr {
			return providerID, ""
		}
		_ = cfg
	}
	return "", modelStr
}

func (svc *Service) handleRerank(w http.ResponseWriter, r *http.Request) {
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

	var req RerankRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "parse request: "+err.Error(), "invalid_request_error")
		return
	}

	if strings.TrimSpace(req.Model) == "" {
		writeOpenAIError(w, http.StatusBadRequest, "missing model", "invalid_request_error")
		return
	}
	if strings.TrimSpace(req.Query) == "" {
		writeOpenAIError(w, http.StatusBadRequest, "missing query", "invalid_request_error")
		return
	}
	if len(req.Documents) == 0 {
		writeOpenAIError(w, http.StatusBadRequest, "documents must be a non-empty array", "invalid_request_error")
		return
	}

	providerID, modelID := parseRerankModel(req.Model)
	providerCfg, ok := rerankProviders[providerID]
	if !ok {
		writeOpenAIError(w, http.StatusBadRequest, "unknown rerank provider: "+providerID, "invalid_request_error")
		return
	}

	apiKey := svc.cfg.APIKeys[providerID]
	if apiKey == "" {
		apiKey = svc.cfg.APIKeys["cohere"]
	}
	if apiKey == "" {
		writeOpenAIError(w, http.StatusBadRequest, "no API key configured for provider: "+providerID, "invalid_request_error")
		return
	}

	upstreamBody := map[string]interface{}{
		"model":     modelID,
		"query":     req.Query,
		"documents": req.Documents,
		"top_n":     req.TopN,
	}
	if req.TopN <= 0 {
		upstreamBody["top_n"] = len(req.Documents)
	}
	if !req.ReturnDocuments {
		upstreamBody["return_documents"] = false
	} else {
		upstreamBody["return_documents"] = true
	}

	// NVIDIA format transform
	if providerCfg.Format == "nvidia" {
		docs := make([]map[string]string, 0, len(req.Documents))
		for _, d := range req.Documents {
			switch v := d.(type) {
			case string:
				docs = append(docs, map[string]string{"text": v})
			case map[string]interface{}:
				if t, ok := v["text"].(string); ok {
					docs = append(docs, map[string]string{"text": t})
				} else {
					docs = append(docs, map[string]string{"text": ""})
				}
			default:
				docs = append(docs, map[string]string{"text": ""})
			}
		}
		upstreamBody = map[string]interface{}{
			"model":    modelID,
			"query":    map[string]string{"text": req.Query},
			"passages": docs,
			"top_n":    upstreamBody["top_n"],
		}
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

	client := &http.Client{Timeout: 60 * time.Second}
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

	// Transform NVIDIA response back to Cohere format
	if providerCfg.Format == "nvidia" {
		results := make([]map[string]interface{}, 0)
		if rankings, ok := upstreamResp["rankings"].([]interface{}); ok {
			for _, r := range rankings {
				if m, ok := r.(map[string]interface{}); ok {
					score := 0.0
					if s, ok := m["logit"].(float64); ok {
						score = s
					} else if s, ok := m["score"].(float64); ok {
						score = s
					}
					idx := 0
					if i, ok := m["index"].(float64); ok {
						idx = int(i)
					}
					text := ""
					if t, ok := m["text"].(string); ok {
						text = t
					}
					results = append(results, map[string]interface{}{
						"index":           idx,
						"relevance_score": score,
						"document":        map[string]string{"text": text},
					})
				}
			}
		}
		upstreamResp = map[string]interface{}{
			"id":      req.Model,
			"results": results,
			"meta": map[string]interface{}{
				"api_version":  map[string]string{"version": "2"},
				"billed_units": map[string]interface{}{"search_units": 1},
			},
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(upstreamResp)
}
