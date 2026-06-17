// Package providers defines the core provider interface and types.
package providers

import "context"

// ToolInputSchema is a JSON Schema object describing a tool's parameters.
type ToolInputSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]ToolPropSchema `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

// ToolPropSchema describes a single property in a tool's input schema.
type ToolPropSchema struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ToolDef is the definition of a tool sent to the AI provider.
type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema ToolInputSchema `json:"input_schema"`
}

// ToolCall is a tool invocation requested by the AI.
type ToolCall struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// ToolResult is the result of executing a tool, sent back to the AI.
type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

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
	ReasoningEffort string         `json:"reasoning_effort,omitempty"` // "low", "medium", "high"
	CacheControl    bool           `json:"cache_control,omitempty"`    // Enable prompt caching
	Complexity      float64        `json:"complexity,omitempty"`       // 0-1, for cost-aware routing
	TaskType        string         `json:"task_type,omitempty"`        // "coding", "review", "plan"...
	MaxTurns        int            `json:"max_turns,omitempty"`        // Limit agentic turns
	AllowedTools    []string       `json:"allowed_tools,omitempty"`    // Restrict tools (save tokens)
	Tools           []ToolDef      `json:"tools,omitempty"`            // Tool definitions for agentic mode
}

// CostTier represents model cost tier.
type CostTier string

const (
	CostTierFree      CostTier = "free"
	CostTierCheap     CostTier = "cheap"
	CostTierMedium    CostTier = "medium"
	CostTierExpensive CostTier = "expensive"
)

// ReasoningEffort represents reasoning effort level.
type ReasoningEffort string

const (
	ReasoningLow    ReasoningEffort = "low"    // Skip deep reasoning, faster, cheaper
	ReasoningMedium ReasoningEffort = "medium" // Default
	ReasoningHigh   ReasoningEffort = "high"   // Full reasoning, slower, expensive
)

// ChatResponse is the output from a chat request.
type ChatResponse struct {
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason,omitempty"`
	Usage        *Usage `json:"usage,omitempty"`
}

// Usage contains token usage information.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a single chunk from a streaming response.
type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}
