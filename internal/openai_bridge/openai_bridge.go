package openai_bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ti/router/tibrain/internal/lazyrouter"
)

type ChatCompletionRequest struct {
	Model    string                 `json:"model"`
	Messages []ChatMessage          `json:"messages"`
	Tools    []ToolDefinition       `json:"tools,omitempty"`
	ToolChoice string               `json:"tool_choice,omitempty"`
	Stream   bool                   `json:"stream,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type ToolDefinition struct {
	Type string `json:"type"`
	Function struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Parameters  map[string]interface{} `json:"parameters"`
	} `json:"function"`
}

type ChatCompletionResponse struct {
	ID        string            `json:"id"`
	Object    string            `json:"object"`
	Created   int64             `json:"created"`
	Model     string            `json:"model"`
	Choices   []Choice          `json:"choices"`
	Usage     Usage             `json:"usage"`
	ToolCalls []ToolCall        `json:"tool_calls,omitempty"`
}

type Choice struct {
	Index        int         `json:"index"`
	Message      ResponseMsg `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type ResponseMsg struct {
	Role    string      `json:"role"`
	Content string      `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	} `json:"function"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ModelsResponse struct {
	Data []ModelInfo `json:"data"`
}

type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
}

type OpenAIHandler struct {
	router *lazyrouter.Router
}

func NewOpenAIHandler(router *lazyrouter.Router) *OpenAIHandler {
	return &OpenAIHandler{router: router}
}

func (h *OpenAIHandler) HandleChatCompletion(w http.ResponseWriter, r *http.Request) {
	var req ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	response := h.generateResponse(r.Context(), req)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *OpenAIHandler) generateResponse(ctx context.Context, req ChatCompletionRequest) *ChatCompletionResponse {
	choices := []Choice{
		{
			Index: 0,
			Message: ResponseMsg{
				Role:    "assistant",
				Content: "Tôi đã nhận được yêu cầu. Tools có sẵn: " + strings.Join(h.router.ListTools(), ", "),
			},
			FinishReason: "stop",
		},
	}

	return &ChatCompletionResponse{
		ID:       fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		Object:   "chat.completion",
		Created:  time.Now().Unix(),
		Model:    req.Model,
		Choices:  choices,
		Usage:    Usage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150},
	}
}

func (h *OpenAIHandler) HandleModelsList(w http.ResponseWriter, r *http.Request) {
	models := []ModelInfo{
		{ID: "tibrain-default", Object: "model", Created: time.Now().Unix()},
		{ID: "tibrain-rag", Object: "model", Created: time.Now().Unix()},
		{ID: "tibrain-embeddings", Object: "model", Created: time.Now().Unix()},
	}

	response := ModelsResponse{Data: models}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *OpenAIHandler) HandleToolList(w http.ResponseWriter, r *http.Request) {
	tools := []ToolDefinition{}
	
	for _, name := range h.router.ListTools() {
		tool, ok := h.router.GetTool(name)
		if ok {
			tools = append(tools, ToolDefinition{
				Type: "function",
				Function: struct {
					Name        string                 `json:"name"`
					Description string                 `json:"description"`
					Parameters  map[string]interface{} `json:"parameters"`
				}{
					Name:        tool.Name(),
					Description: tool.Description(),
					Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
				},
			})
		}
	}

	response := map[string]interface{}{
		"object": "list",
		"data":   tools,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *OpenAIHandler) HandleToolStatus(w http.ResponseWriter, r *http.Request) {
	status := h.router.ToolStatus()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *OpenAIHandler) HandleToolCall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ToolName string                 `json:"tool_name"`
		Params   map[string]interface{} `json:"params"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	tool, ok := h.router.GetTool(req.ToolName)
	if !ok {
		http.Error(w, fmt.Sprintf("Tool '%s' not found or not enabled", req.ToolName), http.StatusNotFound)
		return
	}

	result, err := tool.Execute(r.Context(), req.Params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Tool execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tool":    req.ToolName,
		"result":  result,
		"success": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}