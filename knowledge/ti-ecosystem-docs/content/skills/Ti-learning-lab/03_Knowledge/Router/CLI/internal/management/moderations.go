package management

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// --- OpenAI-compatible moderation types ---

type ModerationRequest struct {
	Model string      `json:"model,omitempty"`
	Input interface{} `json:"input"`
}

type ModerationResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Results []ModerationResult `json:"results"`
}

type ModerationResult struct {
	Flagged        bool               `json:"flagged"`
	Categories     map[string]bool    `json:"categories"`
	CategoryScores map[string]float64 `json:"category_scores"`
}

// --- Moderation provider registry ---

type ModerationProvider struct {
	ID      string `json:"id"`
	BaseURL string `json:"baseUrl"`
}

var moderationProviders = map[string]ModerationProvider{
	"openai": {ID: "openai", BaseURL: "https://api.openai.com/v1/moderations"},
}

func parseModerationModel(modelStr string) (providerID, modelID string) {
	if modelStr == "" {
		return "openai", "omni-moderation-latest"
	}
	slashIdx := strings.Index(modelStr, "/")
	if slashIdx > 0 {
		prefix := modelStr[:slashIdx]
		if _, ok := moderationProviders[prefix]; ok {
			return prefix, modelStr[slashIdx+1:]
		}
	}
	return "openai", modelStr
}

func (svc *Service) handleModerations(w http.ResponseWriter, r *http.Request) {
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

	var req ModerationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "parse request: "+err.Error(), "invalid_request_error")
		return
	}

	if req.Input == nil {
		writeOpenAIError(w, http.StatusBadRequest, "missing input", "invalid_request_error")
		return
	}

	providerID, modelID := parseModerationModel(req.Model)
	providerCfg, ok := moderationProviders[providerID]
	if !ok {
		writeOpenAIError(w, http.StatusBadRequest, "unknown moderation provider: "+providerID, "invalid_request_error")
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
		"input": req.Input,
	}
	if modelID != "" {
		upstreamBody["model"] = modelID
	}
	payload, _ := json.Marshal(upstreamBody)

	upReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, providerCfg.BaseURL, bytes.NewReader(payload))
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "build request: "+err.Error(), "server_error")
		return
	}
	upReq.Header.Set("Content-Type", "application/json")
	upReq.Header.Set("Authorization", "Bearer "+apiKey)

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

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respBody)
}
