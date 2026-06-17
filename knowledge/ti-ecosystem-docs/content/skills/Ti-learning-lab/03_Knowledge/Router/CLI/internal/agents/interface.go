package agents

import (
	"context"
)

// Agent represents an AI agent that can execute tasks
type Agent interface {
	// Execute executes a task and returns the result
	Execute(ctx context.Context, task Task) (Result, error)

	// GetStatus returns the current status of the agent
	GetStatus(ctx context.Context) (Status, error)

	// HealthCheck performs a health check
	HealthCheck(ctx context.Context) (HealthStatus, error)

	// Shutdown shuts down the agent
	Shutdown(ctx context.Context) error
}

// Task represents a task to be executed by an agent
type Task struct {
	ID      string
	Type    string
	Prompt  string
	Context Context
	Options map[string]interface{}
}

// Context provides context for task execution
type Context struct {
	Workspace   string
	SessionID   string
	Environment map[string]string
	Metadata    map[string]interface{}
}

// Result represents the result of a task execution
type Result struct {
	Success  bool
	Data     interface{}
	Metadata map[string]interface{}
	Error    error
}

// Status represents the current status of an agent
type Status struct {
	State    string
	IsIdle   bool
	IsBusy   bool
	Metadata map[string]string
}

// HealthStatus represents the health status of an agent
type HealthStatus struct {
	Healthy bool
	Message string
	Details map[string]string
}
