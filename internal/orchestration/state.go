package orchestration

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type TaskStatus string

const (
	TaskCreated         TaskStatus = "created"
	TaskPlanning        TaskStatus = "planning"
	TaskWaitingApproval TaskStatus = "waiting_approval"
	TaskReady           TaskStatus = "ready"
	TaskDispatched      TaskStatus = "dispatched"
	TaskRunning         TaskStatus = "running"
	TaskVerifying       TaskStatus = "verifying"
	TaskRetrying        TaskStatus = "retrying"
	TaskRecovering      TaskStatus = "recovering"
	TaskBlocked         TaskStatus = "blocked"
	TaskCompleted       TaskStatus = "completed"
	TaskFailed          TaskStatus = "failed"
	TaskCancelled       TaskStatus = "cancelled"
)

var taskTransitions = map[TaskStatus]map[TaskStatus]struct{}{
	TaskCreated:         set(TaskPlanning, TaskCancelled),
	TaskPlanning:        set(TaskWaitingApproval, TaskReady, TaskBlocked, TaskFailed, TaskCancelled),
	TaskWaitingApproval: set(TaskReady, TaskBlocked, TaskCancelled),
	TaskReady:           set(TaskDispatched, TaskBlocked, TaskCancelled),
	TaskDispatched:      set(TaskRunning, TaskRecovering, TaskFailed, TaskCancelled),
	TaskRunning:         set(TaskVerifying, TaskRetrying, TaskRecovering, TaskFailed, TaskCancelled),
	TaskVerifying:       set(TaskCompleted, TaskRetrying, TaskBlocked, TaskFailed),
	TaskRetrying:        set(TaskDispatched, TaskFailed, TaskCancelled),
	TaskRecovering:      set(TaskRunning, TaskVerifying, TaskRetrying, TaskFailed, TaskCancelled),
	TaskBlocked:         set(TaskPlanning, TaskReady, TaskFailed, TaskCancelled),
	TaskCompleted:       set(),
	TaskFailed:          set(TaskRetrying),
	TaskCancelled:       set(),
}

func set(values ...TaskStatus) map[TaskStatus]struct{} {
	result := make(map[TaskStatus]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func IsTaskStatus(value TaskStatus) bool {
	_, ok := taskTransitions[value]
	return ok
}

func CanTransition(from, to TaskStatus) bool {
	allowed, ok := taskTransitions[from]
	if !ok {
		return false
	}
	_, ok = allowed[to]
	return ok
}

func ValidateTransition(from, to TaskStatus) error {
	if !IsTaskStatus(from) {
		return fmt.Errorf("unknown source task status %q", from)
	}
	if !IsTaskStatus(to) {
		return fmt.Errorf("unknown target task status %q", to)
	}
	if !CanTransition(from, to) {
		return fmt.Errorf("task transition %s -> %s is not allowed", from, to)
	}
	return nil
}

type RunStatus string

const (
	RunQueued     RunStatus = "queued"
	RunStarting   RunStatus = "starting"
	RunRunning    RunStatus = "running"
	RunVerifying  RunStatus = "verifying"
	RunRecovering RunStatus = "recovering"
	RunSucceeded  RunStatus = "succeeded"
	RunFailed     RunStatus = "failed"
	RunCancelled  RunStatus = "cancelled"
	RunTimedOut   RunStatus = "timed_out"
)

var runTransitions = map[RunStatus]map[RunStatus]struct{}{
	RunQueued:     runSet(RunStarting, RunCancelled),
	RunStarting:   runSet(RunRunning, RunRecovering, RunFailed, RunCancelled, RunTimedOut),
	RunRunning:    runSet(RunVerifying, RunRecovering, RunFailed, RunCancelled, RunTimedOut),
	RunVerifying:  runSet(RunSucceeded, RunRecovering, RunFailed, RunCancelled, RunTimedOut),
	RunRecovering: runSet(RunStarting, RunRunning, RunVerifying, RunFailed, RunCancelled, RunTimedOut),
	RunSucceeded:  runSet(),
	RunFailed:     runSet(),
	RunCancelled:  runSet(),
	RunTimedOut:   runSet(),
}

func runSet(values ...RunStatus) map[RunStatus]struct{} {
	result := make(map[RunStatus]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func ValidateRunTransition(from, to RunStatus) error {
	allowed, ok := runTransitions[from]
	if !ok {
		return fmt.Errorf("unknown source run status %q", from)
	}
	if _, ok := runTransitions[to]; !ok {
		return fmt.Errorf("unknown target run status %q", to)
	}
	if _, ok := allowed[to]; !ok {
		return fmt.Errorf("run transition %s -> %s is not allowed", from, to)
	}
	return nil
}

type RetryPolicy struct {
	MaxAttempts    int `json:"max_attempts"`
	BackoffSeconds int `json:"backoff_seconds"`
}

type Task struct {
	ID                 string         `json:"id"`
	Title              string         `json:"title"`
	Description        string         `json:"description,omitempty"`
	ProjectID          string         `json:"project_id,omitempty"`
	Status             TaskStatus     `json:"status"`
	Priority           int            `json:"priority"`
	RequestedAgent     string         `json:"requested_agent,omitempty"`
	AssignedAgent      string         `json:"assigned_agent,omitempty"`
	CurrentRunID       string         `json:"current_run_id,omitempty"`
	IdempotencyKey     string         `json:"idempotency_key,omitempty"`
	AcceptanceCriteria []string       `json:"acceptance_criteria,omitempty"`
	AllowedTools       []string       `json:"allowed_tools,omitempty"`
	RequiresApproval   bool           `json:"requires_approval"`
	RetryPolicy        RetryPolicy    `json:"retry_policy"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	Version            int64          `json:"version"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type CreateTaskInput struct {
	ID                 string         `json:"id,omitempty"`
	Title              string         `json:"title"`
	Description        string         `json:"description,omitempty"`
	ProjectID          string         `json:"project_id,omitempty"`
	Priority           int            `json:"priority,omitempty"`
	RequestedAgent     string         `json:"requested_agent,omitempty"`
	IdempotencyKey     string         `json:"idempotency_key,omitempty"`
	AcceptanceCriteria []string       `json:"acceptance_criteria,omitempty"`
	AllowedTools       []string       `json:"allowed_tools,omitempty"`
	RequiresApproval   bool           `json:"requires_approval,omitempty"`
	RetryPolicy        RetryPolicy    `json:"retry_policy,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

func (input *CreateTaskInput) Normalize() error {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return errors.New("title is required")
	}
	if input.Priority == 0 {
		input.Priority = 50
	}
	if input.Priority < 0 || input.Priority > 100 {
		return errors.New("priority must be between 0 and 100")
	}
	if input.RetryPolicy.MaxAttempts == 0 {
		input.RetryPolicy.MaxAttempts = 3
	}
	if input.RetryPolicy.MaxAttempts < 1 || input.RetryPolicy.MaxAttempts > 10 {
		return errors.New("retry_policy.max_attempts must be between 1 and 10")
	}
	if input.RetryPolicy.BackoffSeconds < 0 {
		return errors.New("retry_policy.backoff_seconds cannot be negative")
	}
	return nil
}

type Run struct {
	ID                  string         `json:"id"`
	TaskID              string         `json:"task_id"`
	Attempt             int            `json:"attempt"`
	Status              RunStatus      `json:"status"`
	Executor            string         `json:"executor"`
	AdapterID           string         `json:"adapter_id,omitempty"`
	PID                 int            `json:"pid,omitempty"`
	ProcessStartTime    *time.Time     `json:"process_start_time,omitempty"`
	WorkDir             string         `json:"workdir,omitempty"`
	Argv                []string       `json:"argv,omitempty"`
	TimeoutSeconds      int            `json:"timeout_seconds"`
	HeartbeatAt         *time.Time     `json:"heartbeat_at,omitempty"`
	CheckpointID        string         `json:"checkpoint_id,omitempty"`
	SnapshotID          string         `json:"snapshot_id,omitempty"`
	LogPath             string         `json:"log_path,omitempty"`
	ArtifactIDs         []string       `json:"artifact_ids,omitempty"`
	VerificationProfile string         `json:"verification_profile,omitempty"`
	VerificationStatus  string         `json:"verification_status,omitempty"`
	Error               map[string]any `json:"error,omitempty"`
	Version             int64          `json:"version"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	StartedAt           *time.Time     `json:"started_at,omitempty"`
	FinishedAt          *time.Time     `json:"finished_at,omitempty"`
}

type CreateRunInput struct {
	TaskID              string   `json:"task_id"`
	Executor            string   `json:"executor"`
	AdapterID           string   `json:"adapter_id,omitempty"`
	WorkDir             string   `json:"workdir,omitempty"`
	Argv                []string `json:"argv,omitempty"`
	TimeoutSeconds      int      `json:"timeout_seconds,omitempty"`
	VerificationProfile string   `json:"verification_profile,omitempty"`
}

type Event struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Source        string         `json:"source"`
	SubjectID     string         `json:"subject_id,omitempty"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	CausationID   string         `json:"causation_id,omitempty"`
	Sequence      int64          `json:"sequence"`
	OccurredAt    time.Time      `json:"occurred_at"`
	Data          map[string]any `json:"data"`
}
