package orchestration

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Store struct {
	db  *sql.DB
	now func() time.Time
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db, now: time.Now}
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("orchestration database is not available")
	}
	const schema = `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		project_id TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL,
		priority INTEGER NOT NULL DEFAULT 50,
		requested_agent TEXT NOT NULL DEFAULT '',
		assigned_agent TEXT NOT NULL DEFAULT '',
		current_run_id TEXT NOT NULL DEFAULT '',
		idempotency_key TEXT NOT NULL DEFAULT '',
		acceptance_criteria TEXT NOT NULL DEFAULT '[]',
		allowed_tools TEXT NOT NULL DEFAULT '[]',
		requires_approval INTEGER NOT NULL DEFAULT 0,
		max_attempts INTEGER NOT NULL DEFAULT 3,
		backoff_seconds INTEGER NOT NULL DEFAULT 5,
		metadata TEXT NOT NULL DEFAULT '{}',
		version INTEGER NOT NULL DEFAULT 1,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_idempotency ON tasks(idempotency_key) WHERE idempotency_key <> '';
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status, updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id, updated_at DESC);

	CREATE TABLE IF NOT EXISTS task_runs (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL,
		attempt INTEGER NOT NULL,
		status TEXT NOT NULL,
		executor TEXT NOT NULL,
		adapter_id TEXT NOT NULL DEFAULT '',
		pid INTEGER NOT NULL DEFAULT 0,
		process_start_time INTEGER,
		workdir TEXT NOT NULL DEFAULT '',
		argv TEXT NOT NULL DEFAULT '[]',
		timeout_seconds INTEGER NOT NULL DEFAULT 900,
		heartbeat_at INTEGER,
		checkpoint_id TEXT NOT NULL DEFAULT '',
		snapshot_id TEXT NOT NULL DEFAULT '',
		log_path TEXT NOT NULL DEFAULT '',
		artifact_ids TEXT NOT NULL DEFAULT '[]',
		verification_profile TEXT NOT NULL DEFAULT '',
		verification_status TEXT NOT NULL DEFAULT 'pending',
		error TEXT NOT NULL DEFAULT '{}',
		version INTEGER NOT NULL DEFAULT 1,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		started_at INTEGER,
		finished_at INTEGER,
		FOREIGN KEY(task_id) REFERENCES tasks(id)
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_task_runs_attempt ON task_runs(task_id, attempt);
	CREATE INDEX IF NOT EXISTS idx_task_runs_status ON task_runs(status, heartbeat_at);
	CREATE INDEX IF NOT EXISTS idx_task_runs_task ON task_runs(task_id, created_at DESC);

	CREATE TABLE IF NOT EXISTS task_events (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		source TEXT NOT NULL,
		subject_id TEXT NOT NULL DEFAULT '',
		correlation_id TEXT NOT NULL DEFAULT '',
		causation_id TEXT NOT NULL DEFAULT '',
		sequence INTEGER NOT NULL,
		occurred_at INTEGER NOT NULL,
		data TEXT NOT NULL DEFAULT '{}'
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_task_events_sequence ON task_events(subject_id, sequence);
	CREATE INDEX IF NOT EXISTS idx_task_events_time ON task_events(occurred_at DESC);
	`
	_, err := s.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("initialize orchestration schema: %w", err)
	}
	return nil
}

func (s *Store) CreateTask(ctx context.Context, input CreateTaskInput) (Task, error) {
	if err := input.Normalize(); err != nil {
		return Task{}, err
	}
	if strings.TrimSpace(input.IdempotencyKey) != "" {
		if existing, err := s.GetTaskByIdempotencyKey(ctx, input.IdempotencyKey); err == nil {
			return existing, nil
		} else if !errors.Is(err, sql.ErrNoRows) {
			return Task{}, err
		}
	}
	if strings.TrimSpace(input.ID) == "" {
		input.ID = newID("task")
	}
	now := s.now().UTC()
	criteria, err := marshalJSON(input.AcceptanceCriteria, []string{})
	if err != nil {
		return Task{}, fmt.Errorf("encode acceptance criteria: %w", err)
	}
	tools, err := marshalJSON(input.AllowedTools, []string{})
	if err != nil {
		return Task{}, fmt.Errorf("encode allowed tools: %w", err)
	}
	metadata, err := marshalJSON(input.Metadata, map[string]any{})
	if err != nil {
		return Task{}, fmt.Errorf("encode metadata: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Task{}, err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO tasks (
			id, title, description, project_id, status, priority, requested_agent,
			idempotency_key, acceptance_criteria, allowed_tools, requires_approval,
			max_attempts, backoff_seconds, metadata, version, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		input.ID, input.Title, input.Description, input.ProjectID, string(TaskCreated), input.Priority,
		input.RequestedAgent, input.IdempotencyKey, criteria, tools, boolInt(input.RequiresApproval),
		input.RetryPolicy.MaxAttempts, input.RetryPolicy.BackoffSeconds, metadata, unixMillis(now), unixMillis(now),
	)
	if err != nil {
		return Task{}, fmt.Errorf("insert task: %w", err)
	}
	if err := s.insertEventTx(ctx, tx, "task.created", "tibrain", input.ID, input.ID, "", map[string]any{"status": TaskCreated}); err != nil {
		return Task{}, err
	}
	if err := tx.Commit(); err != nil {
		return Task{}, err
	}
	return s.GetTask(ctx, input.ID)
}

func (s *Store) GetTask(ctx context.Context, id string) (Task, error) {
	return scanTask(s.db.QueryRowContext(ctx, taskSelect+" WHERE id = ?", id))
}

func (s *Store) GetTaskByIdempotencyKey(ctx context.Context, key string) (Task, error) {
	return scanTask(s.db.QueryRowContext(ctx, taskSelect+" WHERE idempotency_key = ?", key))
}

func (s *Store) ListTasks(ctx context.Context, status TaskStatus, limit int) ([]Task, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := taskSelect
	args := []any{}
	if status != "" {
		if !IsTaskStatus(status) {
			return nil, fmt.Errorf("unknown task status %q", status)
		}
		query += " WHERE status = ?"
		args = append(args, string(status))
	}
	query += " ORDER BY updated_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Task{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, task)
	}
	return result, rows.Err()
}

func (s *Store) TransitionTask(ctx context.Context, id string, to TaskStatus, expectedVersion *int64, source, reason string, data map[string]any) (Task, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Task{}, err
	}
	defer tx.Rollback()
	current, err := scanTask(tx.QueryRowContext(ctx, taskSelect+" WHERE id = ?", id))
	if err != nil {
		return Task{}, err
	}
	if expectedVersion != nil && current.Version != *expectedVersion {
		return Task{}, fmt.Errorf("task version conflict: expected %d, found %d", *expectedVersion, current.Version)
	}
	if err := ValidateTransition(current.Status, to); err != nil {
		return Task{}, err
	}
	now := s.now().UTC()
	result, err := tx.ExecContext(ctx, `UPDATE tasks SET status = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ?`, string(to), unixMillis(now), id, current.Version)
	if err != nil {
		return Task{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return Task{}, errors.New("task update lost due to concurrent modification")
	}
	if data == nil {
		data = map[string]any{}
	}
	data["from"] = current.Status
	data["to"] = to
	if strings.TrimSpace(reason) != "" {
		data["reason"] = reason
	}
	if strings.TrimSpace(source) == "" {
		source = "tibrain"
	}
	if err := s.insertEventTx(ctx, tx, "task.transitioned", source, id, id, "", data); err != nil {
		return Task{}, err
	}
	if err := tx.Commit(); err != nil {
		return Task{}, err
	}
	return s.GetTask(ctx, id)
}

func (s *Store) CreateRun(ctx context.Context, input CreateRunInput) (Run, error) {
	input.TaskID = strings.TrimSpace(input.TaskID)
	input.Executor = strings.TrimSpace(input.Executor)
	if input.TaskID == "" || input.Executor == "" {
		return Run{}, errors.New("task_id and executor are required")
	}
	if input.TimeoutSeconds <= 0 {
		input.TimeoutSeconds = 900
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Run{}, err
	}
	defer tx.Rollback()
	task, err := scanTask(tx.QueryRowContext(ctx, taskSelect+" WHERE id = ?", input.TaskID))
	if err != nil {
		return Run{}, err
	}
	if task.Status != TaskReady && task.Status != TaskRetrying && task.Status != TaskRecovering && task.Status != TaskDispatched {
		return Run{}, fmt.Errorf("task %s cannot be dispatched from status %s", task.ID, task.Status)
	}
	var attempt int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(attempt), 0) + 1 FROM task_runs WHERE task_id = ?`, task.ID).Scan(&attempt); err != nil {
		return Run{}, err
	}
	argv, err := marshalJSON(input.Argv, []string{})
	if err != nil {
		return Run{}, err
	}
	now := s.now().UTC()
	runID := newID("run")
	_, err = tx.ExecContext(ctx, `
		INSERT INTO task_runs (
			id, task_id, attempt, status, executor, adapter_id, workdir, argv,
			timeout_seconds, verification_profile, verification_status, version, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', 1, ?, ?)`,
		runID, task.ID, attempt, string(RunQueued), input.Executor, input.AdapterID, input.WorkDir, argv,
		input.TimeoutSeconds, input.VerificationProfile, unixMillis(now), unixMillis(now),
	)
	if err != nil {
		return Run{}, fmt.Errorf("insert task run: %w", err)
	}
	if task.Status != TaskDispatched {
		if err := ValidateTransition(task.Status, TaskDispatched); err != nil {
			return Run{}, err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE tasks SET status = ?, current_run_id = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ?`, string(TaskDispatched), runID, unixMillis(now), task.ID, task.Version)
	if err != nil {
		return Run{}, err
	}
	if err := s.insertEventTx(ctx, tx, "run.created", "tibrain", runID, task.ID, "", map[string]any{"run_id": runID, "attempt": attempt, "executor": input.Executor}); err != nil {
		return Run{}, err
	}
	if err := tx.Commit(); err != nil {
		return Run{}, err
	}
	return s.GetRun(ctx, runID)
}

func (s *Store) GetRun(ctx context.Context, id string) (Run, error) {
	return scanRun(s.db.QueryRowContext(ctx, runSelect+" WHERE id = ?", id))
}

func (s *Store) ListRuns(ctx context.Context, taskID string, limit int) ([]Run, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := runSelect
	args := []any{}
	if strings.TrimSpace(taskID) != "" {
		query += " WHERE task_id = ?"
		args = append(args, taskID)
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Run{}
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, run)
	}
	return result, rows.Err()
}

func (s *Store) TransitionRun(ctx context.Context, id string, to RunStatus, expectedVersion *int64, source string, patch map[string]any) (Run, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Run{}, err
	}
	defer tx.Rollback()
	current, err := scanRun(tx.QueryRowContext(ctx, runSelect+" WHERE id = ?", id))
	if err != nil {
		return Run{}, err
	}
	if expectedVersion != nil && current.Version != *expectedVersion {
		return Run{}, fmt.Errorf("run version conflict: expected %d, found %d", *expectedVersion, current.Version)
	}
	if err := ValidateRunTransition(current.Status, to); err != nil {
		return Run{}, err
	}
	now := s.now().UTC()
	startedAt := nullableMillis(current.StartedAt)
	finishedAt := nullableMillis(current.FinishedAt)
	heartbeatAt := nullableMillis(current.HeartbeatAt)
	if to == RunStarting || to == RunRunning {
		if current.StartedAt == nil {
			startedAt = unixMillis(now)
		}
		heartbeatAt = unixMillis(now)
	}
	if to == RunSucceeded || to == RunFailed || to == RunCancelled || to == RunTimedOut {
		finishedAt = unixMillis(now)
	}
	verificationStatus := current.VerificationStatus
	if to == RunVerifying {
		verificationStatus = "running"
	} else if to == RunSucceeded {
		verificationStatus = "passed"
	} else if to == RunFailed && current.Status == RunVerifying {
		verificationStatus = "failed"
	}
	result, err := tx.ExecContext(ctx, `UPDATE task_runs SET status = ?, heartbeat_at = ?, verification_status = ?, started_at = ?, finished_at = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ?`, string(to), heartbeatAt, verificationStatus, startedAt, finishedAt, unixMillis(now), id, current.Version)
	if err != nil {
		return Run{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return Run{}, errors.New("run update lost due to concurrent modification")
	}
	taskTarget := taskStatusForRun(to)
	if taskTarget != "" {
		task, taskErr := scanTask(tx.QueryRowContext(ctx, taskSelect+" WHERE id = ?", current.TaskID))
		if taskErr != nil {
			return Run{}, taskErr
		}
		if task.Status != taskTarget && CanTransition(task.Status, taskTarget) {
			_, err = tx.ExecContext(ctx, `UPDATE tasks SET status = ?, version = version + 1, updated_at = ? WHERE id = ? AND version = ?`, string(taskTarget), unixMillis(now), task.ID, task.Version)
			if err != nil {
				return Run{}, err
			}
		}
	}
	if patch == nil {
		patch = map[string]any{}
	}
	patch["from"] = current.Status
	patch["to"] = to
	if strings.TrimSpace(source) == "" {
		source = "tibrain"
	}
	if err := s.insertEventTx(ctx, tx, "run.transitioned", source, id, current.TaskID, "", patch); err != nil {
		return Run{}, err
	}
	if err := tx.Commit(); err != nil {
		return Run{}, err
	}
	return s.GetRun(ctx, id)
}

func (s *Store) HeartbeatRun(ctx context.Context, id string, pid int, processStartTime *time.Time, logPath string) (Run, error) {
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `UPDATE task_runs SET heartbeat_at = ?, pid = CASE WHEN ? > 0 THEN ? ELSE pid END, process_start_time = COALESCE(?, process_start_time), log_path = CASE WHEN ? <> '' THEN ? ELSE log_path END, version = version + 1, updated_at = ? WHERE id = ?`, unixMillis(now), pid, pid, nullableMillis(processStartTime), logPath, logPath, unixMillis(now), id)
	if err != nil {
		return Run{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return Run{}, sql.ErrNoRows
	}
	return s.GetRun(ctx, id)
}

func (s *Store) RecoverStaleRuns(ctx context.Context, staleBefore time.Time) ([]Run, error) {
	rows, err := s.db.QueryContext(ctx, runSelect+` WHERE status IN (?, ?, ?) AND COALESCE(heartbeat_at, updated_at) < ? ORDER BY updated_at`, string(RunStarting), string(RunRunning), string(RunVerifying), unixMillis(staleBefore.UTC()))
	if err != nil {
		return nil, err
	}
	var stale []Run
	for rows.Next() {
		run, scanErr := scanRun(rows)
		if scanErr != nil {
			rows.Close()
			return nil, scanErr
		}
		stale = append(stale, run)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	recovered := make([]Run, 0, len(stale))
	for _, run := range stale {
		version := run.Version
		updated, transitionErr := s.TransitionRun(ctx, run.ID, RunRecovering, &version, "recovery", map[string]any{"reason": "stale heartbeat", "stale_before": staleBefore.UTC().Format(time.RFC3339)})
		if transitionErr != nil {
			continue
		}
		recovered = append(recovered, updated)
	}
	return recovered, nil
}

func (s *Store) ListEvents(ctx context.Context, subjectID string, limit int) ([]Event, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	query := `SELECT id, type, source, subject_id, correlation_id, causation_id, sequence, occurred_at, data FROM task_events`
	args := []any{}
	if strings.TrimSpace(subjectID) != "" {
		query += " WHERE subject_id = ? OR correlation_id = ?"
		args = append(args, subjectID, subjectID)
	}
	query += " ORDER BY occurred_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Event{}
	for rows.Next() {
		var event Event
		var occurred int64
		var data string
		if err := rows.Scan(&event.ID, &event.Type, &event.Source, &event.SubjectID, &event.CorrelationID, &event.CausationID, &event.Sequence, &occurred, &data); err != nil {
			return nil, err
		}
		event.OccurredAt = fromMillis(occurred)
		_ = json.Unmarshal([]byte(data), &event.Data)
		result = append(result, event)
	}
	return result, rows.Err()
}

func (s *Store) insertEventTx(ctx context.Context, tx *sql.Tx, eventType, source, subjectID, correlationID, causationID string, data map[string]any) error {
	var sequence int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(sequence), 0) + 1 FROM task_events WHERE subject_id = ?`, subjectID).Scan(&sequence); err != nil {
		return err
	}
	encoded, err := marshalJSON(data, map[string]any{})
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO task_events (id, type, source, subject_id, correlation_id, causation_id, sequence, occurred_at, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, newID("event"), eventType, source, subjectID, correlationID, causationID, sequence, unixMillis(s.now().UTC()), encoded)
	return err
}

func taskStatusForRun(status RunStatus) TaskStatus {
	switch status {
	case RunStarting, RunRunning:
		return TaskRunning
	case RunVerifying:
		return TaskVerifying
	case RunRecovering:
		return TaskRecovering
	case RunSucceeded:
		return TaskCompleted
	case RunFailed, RunTimedOut:
		return TaskFailed
	case RunCancelled:
		return TaskCancelled
	default:
		return ""
	}
}

const taskSelect = `SELECT id, title, description, project_id, status, priority, requested_agent, assigned_agent, current_run_id, idempotency_key, acceptance_criteria, allowed_tools, requires_approval, max_attempts, backoff_seconds, metadata, version, created_at, updated_at FROM tasks`
const runSelect = `SELECT id, task_id, attempt, status, executor, adapter_id, pid, process_start_time, workdir, argv, timeout_seconds, heartbeat_at, checkpoint_id, snapshot_id, log_path, artifact_ids, verification_profile, verification_status, error, version, created_at, updated_at, started_at, finished_at FROM task_runs`

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(row scanner) (Task, error) {
	var task Task
	var status string
	var criteria, tools, metadata string
	var approval int
	var createdAt, updatedAt int64
	err := row.Scan(&task.ID, &task.Title, &task.Description, &task.ProjectID, &status, &task.Priority, &task.RequestedAgent, &task.AssignedAgent, &task.CurrentRunID, &task.IdempotencyKey, &criteria, &tools, &approval, &task.RetryPolicy.MaxAttempts, &task.RetryPolicy.BackoffSeconds, &metadata, &task.Version, &createdAt, &updatedAt)
	if err != nil {
		return Task{}, err
	}
	task.Status = TaskStatus(status)
	task.RequiresApproval = approval != 0
	task.CreatedAt = fromMillis(createdAt)
	task.UpdatedAt = fromMillis(updatedAt)
	_ = json.Unmarshal([]byte(criteria), &task.AcceptanceCriteria)
	_ = json.Unmarshal([]byte(tools), &task.AllowedTools)
	_ = json.Unmarshal([]byte(metadata), &task.Metadata)
	return task, nil
}

func scanRun(row scanner) (Run, error) {
	var run Run
	var status string
	var processStart, heartbeat, started, finished sql.NullInt64
	var argv, artifacts, encodedError string
	var createdAt, updatedAt int64
	err := row.Scan(&run.ID, &run.TaskID, &run.Attempt, &status, &run.Executor, &run.AdapterID, &run.PID, &processStart, &run.WorkDir, &argv, &run.TimeoutSeconds, &heartbeat, &run.CheckpointID, &run.SnapshotID, &run.LogPath, &artifacts, &run.VerificationProfile, &run.VerificationStatus, &encodedError, &run.Version, &createdAt, &updatedAt, &started, &finished)
	if err != nil {
		return Run{}, err
	}
	run.Status = RunStatus(status)
	run.ProcessStartTime = nullTime(processStart)
	run.HeartbeatAt = nullTime(heartbeat)
	run.StartedAt = nullTime(started)
	run.FinishedAt = nullTime(finished)
	run.CreatedAt = fromMillis(createdAt)
	run.UpdatedAt = fromMillis(updatedAt)
	_ = json.Unmarshal([]byte(argv), &run.Argv)
	_ = json.Unmarshal([]byte(artifacts), &run.ArtifactIDs)
	_ = json.Unmarshal([]byte(encodedError), &run.Error)
	return run, nil
}

func newID(prefix string) string {
	buffer := make([]byte, 12)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(buffer)
}

func marshalJSON(value any, fallback any) (string, error) {
	if value == nil {
		value = fallback
	}
	encoded, err := json.Marshal(value)
	return string(encoded), err
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func unixMillis(value time.Time) int64 { return value.UTC().UnixMilli() }
func fromMillis(value int64) time.Time { return time.UnixMilli(value).UTC() }
func nullableMillis(value *time.Time) any {
	if value == nil {
		return nil
	}
	return unixMillis(*value)
}
func nullTime(value sql.NullInt64) *time.Time {
	if !value.Valid {
		return nil
	}
	result := fromMillis(value.Int64)
	return &result
}
