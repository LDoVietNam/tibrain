package management

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// --- OpenAI Responses API types ---

type ResponsesRequest struct {
	Model        string        `json:"model"`
	Input        interface{}   `json:"input"`
	Instructions string        `json:"instructions,omitempty"`
	Stream       bool          `json:"stream,omitempty"`
	Tools        []interface{} `json:"tools,omitempty"`
	ToolChoice   interface{}   `json:"tool_choice,omitempty"`
	Temperature  float64       `json:"temperature,omitempty"`
	MaxTokens    int           `json:"max_tokens,omitempty"`
	TopP         float64       `json:"top_p,omitempty"`
}

// ResponsesResponse matches the OpenAI Responses API output format.
type ResponsesResponse struct {
	ID        string                 `json:"id"`
	Object    string                 `json:"object"`
	CreatedAt int64                  `json:"created_at"`
	Model     string                 `json:"model"`
	Output    []interface{}          `json:"output"`
	Usage     map[string]interface{} `json:"usage,omitempty"`
}

// handleResponses handles POST /v1/responses by converting to Chat Completions
// and proxying upstream. This is a simplified port of the TS responsesHandler.ts.
func (svc *Service) handleResponses(w http.ResponseWriter, r *http.Request) {
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

	var req ResponsesRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "parse request: "+err.Error(), "invalid_request_error")
		return
	}

	if strings.TrimSpace(req.Model) == "" {
		writeOpenAIError(w, http.StatusBadRequest, "missing model", "invalid_request_error")
		return
	}

	// Convert Responses API format to Chat Completions format
	messages := convertResponsesInputToMessages(req.Input, req.Instructions)

	chatReq := OpenAIChatRequest{
		Model:    req.Model,
		Messages: messages,
	}
	if req.Temperature != 0 {
		t := req.Temperature
		chatReq.Temperature = &t
	}
	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	}
	chatReq.Stream = req.Stream

	newBody, _ := json.Marshal(chatReq)
	proxyReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, "", bytes.NewReader(newBody))
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "build request: "+err.Error(), "server_error")
		return
	}
	proxyReq.Header = r.Header.Clone()

	// Re-use chat completions logic via internal proxy
	// For simplicity, forward to the same management server's chat endpoint
	upstreamBase := "http://localhost:" + strconv.Itoa(svc.cfg.Port) + "/v1/chat/completions"
	if svc.cfg.Host != "" && svc.cfg.Host != "0.0.0.0" {
		upstreamBase = "http://" + svc.cfg.Host + ":" + strconv.Itoa(svc.cfg.Port) + "/v1/chat/completions"
	}

	upReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamBase, bytes.NewReader(newBody))
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "build upstream request: "+err.Error(), "server_error")
		return
	}
	upReq.Header.Set("Content-Type", "application/json")
	if auth := r.Header.Get("Authorization"); auth != "" {
		upReq.Header.Set("Authorization", auth)
	}

	client := &http.Client{Timeout: 300 * time.Second}
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

	// Convert Chat Completions response back to Responses API format
	var chatResp OpenAIChatResponse
	_ = json.Unmarshal(respBody, &chatResp)

	result := ResponsesResponse{
		ID:        chatResp.ID,
		Object:    "response",
		CreatedAt: chatResp.Created,
		Model:     chatResp.Model,
		Usage: map[string]interface{}{
			"input_tokens":  chatResp.Usage.PromptTokens,
			"output_tokens": chatResp.Usage.CompletionTokens,
			"total_tokens":  chatResp.Usage.TotalTokens,
		},
	}

	if len(chatResp.Choices) > 0 {
		choice := chatResp.Choices[0]
		content := choice.Message.Content
		if content == "" && choice.Delta != nil {
			content = choice.Delta.Content
		}
		result.Output = []interface{}{
			map[string]interface{}{
				"type":    "message",
				"role":    "assistant",
				"content": []map[string]string{{"type": "output_text", "text": content}},
			},
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// convertResponsesInputToMessages converts the Responses API "input" field
// (string or array of message objects) into OpenAI Chat Completions messages.
func convertResponsesInputToMessages(input interface{}, instructions string) []OpenAIMessage {
	var messages []OpenAIMessage
	if instructions != "" {
		messages = append(messages, OpenAIMessage{Role: "system", Content: instructions})
	}

	switch v := input.(type) {
	case string:
		messages = append(messages, OpenAIMessage{Role: "user", Content: v})
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				role, _ := m["role"].(string)
				content, _ := m["content"].(string)
				if role == "" {
					role = "user"
				}
				messages = append(messages, OpenAIMessage{Role: role, Content: content})
			}
		}
	case map[string]interface{}:
		role, _ := v["role"].(string)
		content, _ := v["content"].(string)
		if role == "" {
			role = "user"
		}
		messages = append(messages, OpenAIMessage{Role: role, Content: content})
	}
	return messages
}
