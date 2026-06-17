package management

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// --- OpenAI-compatible audio speech types ---

type AudioSpeechRequest struct {
	Model string  `json:"model"`
	Input string  `json:"input"`
	Voice string  `json:"voice,omitempty"`
	Speed float64 `json:"speed,omitempty"`
}

// --- Audio provider registry ---

type AudioProvider struct {
	ID         string `json:"id"`
	BaseURL    string `json:"baseUrl"`
	AuthType   string `json:"authType"`
	AuthHeader string `json:"authHeader"`
}

var audioProviders = map[string]AudioProvider{
	"openai":     {ID: "openai", BaseURL: "https://api.openai.com/v1/audio/speech", AuthType: "apikey", AuthHeader: "bearer"},
	"elevenlabs": {ID: "elevenlabs", BaseURL: "https://api.elevenlabs.io/v1/text-to-speech", AuthType: "apikey", AuthHeader: "bearer"},
	"deepgram":   {ID: "deepgram", BaseURL: "https://api.deepgram.com/v1/speak", AuthType: "apikey", AuthHeader: "bearer"},
}

func parseAudioModel(modelStr string) (providerID, modelID string) {
	if modelStr == "" {
		return "", ""
	}
	slashIdx := strings.Index(modelStr, "/")
	if slashIdx > 0 {
		prefix := modelStr[:slashIdx]
		if _, ok := audioProviders[prefix]; ok {
			return prefix, modelStr[slashIdx+1:]
		}
	}
	return "openai", modelStr
}

func (svc *Service) handleAudioSpeech(w http.ResponseWriter, r *http.Request) {
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

	var req AudioSpeechRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "parse request: "+err.Error(), "invalid_request_error")
		return
	}

	if strings.TrimSpace(req.Input) == "" {
		writeOpenAIError(w, http.StatusBadRequest, "missing input", "invalid_request_error")
		return
	}

	providerID, modelID := parseAudioModel(req.Model)
	providerCfg, ok := audioProviders[providerID]
	if !ok {
		writeOpenAIError(w, http.StatusBadRequest, "unknown audio provider: "+providerID, "invalid_request_error")
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

	var targetURL string
	upstreamBody := map[string]interface{}{}

	if providerID == "elevenlabs" {
		voice := req.Voice
		if voice == "" {
			voice = "21m00Tcm4TlvDq8ikWAM"
		}
		targetURL = providerCfg.BaseURL + "/" + voice
		upstreamBody["text"] = req.Input
		upstreamBody["model_id"] = modelID
	} else {
		targetURL = providerCfg.BaseURL
		upstreamBody["model"] = modelID
		upstreamBody["input"] = req.Input
		if req.Voice != "" {
			upstreamBody["voice"] = req.Voice
		}
		if req.Speed != 0 {
			upstreamBody["speed"] = req.Speed
		}
	}

	payload, _ := json.Marshal(upstreamBody)
	upReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, targetURL, bytes.NewReader(payload))
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

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(respBody)
		return
	}

	// Stream binary audio back
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "audio/mpeg"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, resp.Body)
}
