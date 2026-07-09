package orchestration

import "context"

type ModelRequest struct {
	TaskID       string         `json:"task_id"`
	Capability   string         `json:"capability"`
	Prompt       string         `json:"prompt"`
	Context      map[string]any `json:"context,omitempty"`
	PreferredIDs []string       `json:"preferred_ids,omitempty"`
}

type ModelTarget struct {
	ProviderID string         `json:"provider_id"`
	ModelID    string         `json:"model_id"`
	Endpoint   string         `json:"endpoint,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type ModelResult struct {
	Content  string         `json:"content"`
	Usage    map[string]any `json:"usage,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type ModelRouter interface {
	Health(ctx context.Context) error
	SelectModel(ctx context.Context, request ModelRequest) (ModelTarget, error)
	Invoke(ctx context.Context, target ModelTarget, request ModelRequest) (ModelResult, error)
}

type ToolCall struct {
	TaskID    string         `json:"task_id"`
	RunID     string         `json:"run_id"`
	ToolID    string         `json:"tool_id"`
	Arguments map[string]any `json:"arguments"`
}

type ToolResult struct {
	Success  bool           `json:"success"`
	Content  any            `json:"content,omitempty"`
	Error    string         `json:"error,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type ToolGateway interface {
	Health(ctx context.Context) error
	ListTools(ctx context.Context) ([]string, error)
	Call(ctx context.Context, call ToolCall) (ToolResult, error)
}

type ExecutionSpec struct {
	TaskID          string            `json:"task_id"`
	RunID           string            `json:"run_id"`
	IdempotencyKey  string            `json:"idempotency_key"`
	WorkDir         string            `json:"workdir"`
	Argv            []string          `json:"argv"`
	EnvironmentRefs map[string]string `json:"environment_refs,omitempty"`
	TimeoutSeconds  int               `json:"timeout_seconds"`
	Snapshot        bool              `json:"snapshot"`
}

type ExecutionHandle struct {
	RunID     string `json:"run_id"`
	PID       int    `json:"pid,omitempty"`
	LogPath   string `json:"log_path,omitempty"`
	StatePath string `json:"state_path,omitempty"`
}

type Executor interface {
	Start(ctx context.Context, spec ExecutionSpec) (ExecutionHandle, error)
	Status(ctx context.Context, runID string) (RunStatus, error)
	Heartbeat(ctx context.Context, runID string) error
	Cancel(ctx context.Context, runID string) error
	Recover(ctx context.Context, runID string) (ExecutionHandle, error)
}

type VerificationRequest struct {
	TaskID             string   `json:"task_id"`
	RunID              string   `json:"run_id"`
	Workspace          string   `json:"workspace"`
	Profile            string   `json:"profile"`
	AcceptanceCriteria []string `json:"acceptance_criteria,omitempty"`
}

type VerificationResult struct {
	Passed    bool           `json:"passed"`
	Summary   string         `json:"summary"`
	Artifacts []string       `json:"artifacts,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

type Verifier interface {
	Verify(ctx context.Context, request VerificationRequest) (VerificationResult, error)
}
