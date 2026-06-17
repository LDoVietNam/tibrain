package management

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// --- OpenAI Audio Transcriptions types ---

type AudioTranscriptionRequest struct {
	Model          string  `json:"model"`
	Language       string  `json:"language,omitempty"`
	Prompt         string  `json:"prompt,omitempty"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Temperature    float64 `json:"temperature,omitempty"`
}

// handleAudioTranscriptions proxies POST /v1/audio/transcriptions to upstream.
func (svc *Service) handleAudioTranscriptions(w http.ResponseWriter, r *http.Request) {
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

	// Parse multipart form for file upload
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "parse form: "+err.Error(), "invalid_request_error")
		return
	}

	model := r.FormValue("model")
	if strings.TrimSpace(model) == "" {
		writeOpenAIError(w, http.StatusBadRequest, "missing model", "invalid_request_error")
		return
	}

	providerID := "openai"
	providerCfg := struct {
		BaseURL    string
		AuthHeader string
	}{
		BaseURL:    "https://api.openai.com/v1/audio/transcriptions",
		AuthHeader: "bearer",
	}

	// Allow provider override via model prefix
	slashIdx := strings.Index(model, "/")
	if slashIdx > 0 {
		prefix := model[:slashIdx]
		if prefix == "groq" || prefix == "openai" {
			providerID = prefix
			model = model[slashIdx+1:]
		}
	}
	if providerID == "groq" {
		providerCfg.BaseURL = "https://api.groq.com/openai/v1/audio/transcriptions"
	}

	apiKey := svc.cfg.APIKeys[providerID]
	if apiKey == "" {
		apiKey = svc.cfg.APIKeys["openai"]
	}
	if apiKey == "" {
		writeOpenAIError(w, http.StatusBadRequest, "no API key configured for provider: "+providerID, "invalid_request_error")
		return
	}

	// Rebuild multipart request for upstream
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("model", model)
	if lang := r.FormValue("language"); lang != "" {
		_ = mw.WriteField("language", lang)
	}
	if prompt := r.FormValue("prompt"); prompt != "" {
		_ = mw.WriteField("prompt", prompt)
	}
	if format := r.FormValue("response_format"); format != "" {
		_ = mw.WriteField("response_format", format)
	}
	if temp := r.FormValue("temperature"); temp != "" {
		_ = mw.WriteField("temperature", temp)
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "missing file: "+err.Error(), "invalid_request_error")
		return
	}
	defer file.Close()
	fw, err := mw.CreateFormFile("file", header.Filename)
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "create form file: "+err.Error(), "server_error")
		return
	}
	_, _ = io.Copy(fw, file)
	_ = mw.Close()

	upReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, providerCfg.BaseURL, &buf)
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "build request: "+err.Error(), "server_error")
		return
	}
	upReq.Header.Set("Content-Type", mw.FormDataContentType())
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

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respBody)
}
