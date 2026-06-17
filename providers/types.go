package providers

import "context"

// Provider is the interface all AI backends must implement.
type Provider interface {
	Name() string
	DefaultModel() string
	Models() []string
	IsHealthy() bool
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error)
}

// Message represents a single chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the input for a chat request.
type ChatRequest struct {
	Model           string         `json:"model,omitempty"`
	Messages        []Message      `json:"messages"`
	Options         map[string]any `json:"options,omitempty"`
	ReasoningEffort string         `json:"reasoning_effort,omitempty"`
	CacheControl    bool           `json:"cache_control,omitempty"`
	Complexity      float64        `json:"complexity,omitempty"`
	TaskType        string         `json:"task_type,omitempty"`
	MaxTurns        int            `json:"max_turns,omitempty"`
	AllowedTools    []string       `json:"allowed_tools,omitempty"`
}

// ChatResponse is the output from a chat request.
type ChatResponse struct {
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason,omitempty"`
}

// StreamChunk represents a single chunk from a streaming response.
type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}
