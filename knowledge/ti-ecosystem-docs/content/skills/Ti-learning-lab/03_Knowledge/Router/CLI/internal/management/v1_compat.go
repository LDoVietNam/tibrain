package management

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

func (svc *Service) handleV1Root(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusOK)
	case http.MethodGet:
		svc.handleModels(w, r)
	default:
		writeOpenAIError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request_error")
	}
}

func (svc *Service) handleMessages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusOK)
		return
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeOpenAIError(w, http.StatusBadRequest, "read body: "+err.Error(), "invalid_request_error")
			return
		}
		defer r.Body.Close()

		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			writeOpenAIError(w, http.StatusBadRequest, "parse request: "+err.Error(), "invalid_request_error")
			return
		}
		delete(payload, "reasoning_effort")
		delete(payload, "reasoning")
		delete(payload, "thinking")

		cleaned, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, r.URL.String(), bytes.NewReader(cleaned))
		if err != nil {
			writeOpenAIError(w, http.StatusInternalServerError, "build request: "+err.Error(), "server_error")
			return
		}
		req.Header = r.Header.Clone()
		handleChat(svc).ServeHTTP(w, req)
	default:
		writeOpenAIError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request_error")
	}
}

func (svc *Service) handleMessagesCountTokens(w http.ResponseWriter, r *http.Request) {
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
	var body struct {
		Messages []struct {
			Content interface{} `json:"content"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "invalid JSON body", "invalid_request_error")
		return
	}
	defer r.Body.Close()

	totalChars := 0
	for _, msg := range body.Messages {
		switch content := msg.Content.(type) {
		case string:
			totalChars += len(content)
		case []interface{}:
			for _, part := range content {
				if p, ok := part.(map[string]interface{}); ok {
					if p["type"] == "text" {
						if t, ok := p["text"].(string); ok {
							totalChars += len(t)
						}
					}
				}
			}
		}
	}
	inputTokens := (totalChars + 3) / 4
	writeJSON(w, http.StatusOK, map[string]interface{}{"input_tokens": inputTokens})
}

func (svc *Service) handleV1APIChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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

	modelName := "llama3.2"
	var payload map[string]interface{}
	if json.Unmarshal(body, &payload) == nil {
		if m, ok := payload["model"].(string); ok && strings.TrimSpace(m) != "" {
			modelName = m
		}
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, r.URL.String(), bytes.NewReader(body))
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "build request: "+err.Error(), "server_error")
		return
	}
	req.Header = r.Header.Clone()

	rr := httptest.NewRecorder()
	handleChat(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		for k, vv := range rr.Header() {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(rr.Code)
		_, _ = io.Copy(w, rr.Body)
		return
	}

	var openaiResp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &openaiResp); err != nil {
		for k, vv := range rr.Header() {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(rr.Code)
		_, _ = io.Copy(w, rr.Body)
		return
	}

	content := ""
	if choices, ok := openaiResp["choices"].([]interface{}); ok && len(choices) > 0 {
		if c, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := c["message"].(map[string]interface{}); ok {
				if t, ok := msg["content"].(string); ok {
					content = t
				}
			}
		}
	}

	ollamaResp := map[string]interface{}{
		"model": modelName,
		"message": map[string]interface{}{
			"role":    "assistant",
			"content": content,
		},
		"done": true,
	}
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, http.StatusOK, ollamaResp)
}

func (svc *Service) handleProviderScopedV1(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v1/providers/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		writeOpenAIError(w, http.StatusBadRequest, "invalid provider route", "invalid_request_error")
		return
	}
	provider := strings.TrimSpace(parts[0])
	tail := strings.Join(parts[1:], "/")
	if provider == "" {
		writeOpenAIError(w, http.StatusBadRequest, "missing provider", "invalid_request_error")
		return
	}

	switch tail {
	case "chat/completions":
		svc.handleProviderChatCompletions(provider, w, r)
	case "embeddings":
		svc.handleProviderEmbeddings(provider, w, r)
	case "images/generations":
		svc.handleProviderImageGenerations(provider, w, r)
	default:
		writeOpenAIError(w, http.StatusNotFound, "unknown provider endpoint", "invalid_request_error")
	}
}

func (svc *Service) handleProviderChatCompletions(provider string, w http.ResponseWriter, r *http.Request) {
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

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "invalid JSON body", "invalid_request_error")
		return
	}
	if err := forceProviderModelPrefix(payload, provider, "auto"); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, err.Error(), "invalid_request_error")
		return
	}

	reqBody, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, r.URL.String(), bytes.NewReader(reqBody))
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "build request: "+err.Error(), "server_error")
		return
	}
	req.Header = r.Header.Clone()
	handleChat(svc).ServeHTTP(w, req)
}

func (svc *Service) handleProviderEmbeddings(provider string, w http.ResponseWriter, r *http.Request) {
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

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "invalid JSON body", "invalid_request_error")
		return
	}
	if err := forceProviderModelPrefix(payload, provider, "text-embedding-3-small"); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, err.Error(), "invalid_request_error")
		return
	}

	reqBody, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, r.URL.String(), bytes.NewReader(reqBody))
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "build request: "+err.Error(), "server_error")
		return
	}
	req.Header = r.Header.Clone()
	svc.handleEmbeddings(w, req)
}

func (svc *Service) handleProviderImageGenerations(provider string, w http.ResponseWriter, r *http.Request) {
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

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "invalid JSON body", "invalid_request_error")
		return
	}
	if err := forceProviderModelPrefix(payload, provider, "gpt-image-1"); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, err.Error(), "invalid_request_error")
		return
	}

	reqBody, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, r.URL.String(), bytes.NewReader(reqBody))
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "build request: "+err.Error(), "server_error")
		return
	}
	req.Header = r.Header.Clone()
	svc.handleImageGenerations(w, req)
}

func forceProviderModelPrefix(payload map[string]interface{}, provider string, defaultModel string) error {
	raw, ok := payload["model"].(string)
	if !ok || strings.TrimSpace(raw) == "" {
		payload["model"] = provider + "/" + defaultModel
		return nil
	}
	model := strings.TrimSpace(raw)
	if strings.Contains(model, "/") {
		prefix := strings.SplitN(model, "/", 2)[0]
		if prefix != provider {
			return &routeError{msg: "model \"" + model + "\" does not belong to provider \"" + provider + "\""}
		}
		return nil
	}
	payload["model"] = provider + "/" + model
	return nil
}

type routeError struct{ msg string }

func (e *routeError) Error() string { return e.msg }
