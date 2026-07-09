package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type execJobStatus string

const (
	execJobRunning   execJobStatus = "running"
	execJobCompleted execJobStatus = "completed"
	execJobFailed    execJobStatus = "failed"
	execJobCanceled  execJobStatus = "canceled"
	execJobTimedOut  execJobStatus = "timed_out"
)

type execJob struct {
	ID         string
	Command    string
	WorkingDir string
	StartedAt  time.Time
	FinishedAt time.Time
	Status     execJobStatus
	PID        int
	ExitCode   int
	Stdout     string
	Stderr     string
	DurationMs int64
	Error      string

	cancel context.CancelFunc
	done   chan struct{}

	mu sync.RWMutex
}

type execJobStore struct {
	mu   sync.Mutex
	next int64
	jobs map[string]*execJob
}

var globalExecJobStore = &execJobStore{
	jobs: map[string]*execJob{},
}

var defaultAllowedRoots = []string{
	getTiBrainDir(),
	`Z:\02_CORE\_cli\.config`,
	`Z:\02_CORE\skills`,
	`Z:\01_PROJECTS\apps\extension`,
}

var configuredAllowedRootsMu sync.RWMutex
var configuredAllowedRoots []string

func setExecutionPlaneAllowedRoots(roots []string) {
	configuredAllowedRootsMu.Lock()
	defer configuredAllowedRootsMu.Unlock()
	configuredAllowedRoots = normalizeAllowedRoots(roots)
}

func (s *execJobStore) newID() string {
	id := atomic.AddInt64(&s.next, 1)
	return fmt.Sprintf("exec-%d-%d", time.Now().UnixNano(), id)
}

func (s *execJobStore) put(job *execJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
}

func (s *execJobStore) get(id string) (*execJob, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	return job, ok
}

func (s *execJobStore) snapshot(id string) (map[string]interface{}, bool) {
	job, ok := s.get(id)
	if !ok {
		return nil, false
	}
	return job.snapshot(), true
}

func startTrackedShellCommand(command, workingDir string, timeoutMs int) (*execJob, error) {
	if strings.TrimSpace(command) == "" {
		return nil, fmt.Errorf("command is required")
	}

	runCtx := context.Background()
	cancel := func() {}
	if timeoutMs > 0 {
		var timeoutCancel context.CancelFunc
		runCtx, timeoutCancel = context.WithTimeout(runCtx, time.Duration(timeoutMs)*time.Millisecond)
		cancel = timeoutCancel
	} else {
		runCtx, cancel = context.WithCancel(runCtx)
	}

	cmd := exec.CommandContext(runCtx, resolveShellCommand("cmd.exe"), "/C", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	job := &execJob{
		ID:         globalExecJobStore.newID(),
		Command:    command,
		WorkingDir: workingDir,
		StartedAt:  time.Now().UTC(),
		Status:     execJobRunning,
		done:       make(chan struct{}),
		cancel:     cancel,
		ExitCode:   -1,
	}
	globalExecJobStore.put(job)

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	go func() {
		defer close(job.done)

		start := time.Now()
		job.mu.Lock()
		job.PID = 0
		job.mu.Unlock()

		if err := cmd.Start(); err != nil {
			job.finish(execJobFailed, -1, stdoutBuf.String(), stderrBuf.String(), time.Since(start), err.Error())
			return
		}

		job.mu.Lock()
		if cmd.Process != nil {
			job.PID = cmd.Process.Pid
		}
		job.mu.Unlock()

		waitErr := cmd.Wait()
		exitCode := 0
		status := execJobCompleted
		if waitErr != nil {
			exitCode = exitCodeFromErr(waitErr)
			switch runCtx.Err() {
			case context.Canceled:
				status = execJobCanceled
			case context.DeadlineExceeded:
				status = execJobTimedOut
			default:
				status = execJobFailed
			}
		}

		job.finish(status, exitCode, stdoutBuf.String(), stderrBuf.String(), time.Since(start), execErrorString(waitErr))
	}()

	return job, nil
}

func resolveShellCommand(command string) string {
	if strings.EqualFold(command, "cmd.exe") {
		if windir := strings.TrimSpace(os.Getenv("WINDIR")); windir != "" {
			candidate := filepath.Join(windir, "System32", "cmd.exe")
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}
	return command
}

func waitForTrackedJob(ctx context.Context, job *execJob) (map[string]interface{}, error) {
	select {
	case <-job.done:
		return job.snapshot(), nil
	case <-ctx.Done():
		return job.snapshot(), ctx.Err()
	}
}

func cancelTrackedJob(job *execJob) map[string]interface{} {
	job.mu.Lock()
	cancel := job.cancel
	job.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	select {
	case <-job.done:
	case <-time.After(2 * time.Second):
	}
	return job.snapshot()
}

func (j *execJob) finish(status execJobStatus, exitCode int, stdout, stderr string, duration time.Duration, errMsg string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.Status != execJobRunning {
		return
	}
	j.Status = status
	j.ExitCode = exitCode
	j.Stdout = stdout
	j.Stderr = stderr
	j.DurationMs = duration.Milliseconds()
	j.FinishedAt = time.Now().UTC()
	j.Error = strings.TrimSpace(errMsg)
}

func (j *execJob) snapshot() map[string]interface{} {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return map[string]interface{}{
		"execId":     j.ID,
		"command":    j.Command,
		"workingDir": j.WorkingDir,
		"startedAt":  j.StartedAt.Format(time.RFC3339Nano),
		"finishedAt": j.FinishedAt.Format(time.RFC3339Nano),
		"status":     string(j.Status),
		"running":    j.Status == execJobRunning,
		"pid":        j.PID,
		"exitCode":   j.ExitCode,
		"exit_code":  j.ExitCode,
		"stdout":     j.Stdout,
		"stderr":     j.Stderr,
		"durationMs": j.DurationMs,
		"error":      j.Error,
	}
}

func execErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func allowedToolRoots() []string {
	configuredAllowedRootsMu.RLock()
	roots := append([]string(nil), configuredAllowedRoots...)
	configuredAllowedRootsMu.RUnlock()
	if len(roots) == 0 {
		roots = parseAllowedToolRoots()
	}
	if len(roots) == 0 {
		roots = defaultAllowedRoots
	}
	return normalizeAllowedRoots(roots)
}

func parseAllowedToolRoots() []string {
	raw := strings.TrimSpace(getEnvOrDefault("TIBRAIN_ALLOWED_ROOTS", ""))
	if raw == "" {
		raw = strings.TrimSpace(getEnvOrDefault("MCP_ALLOWED_ROOTS", ""))
	}
	if raw == "" {
		return nil
	}

	return parseAllowedRootsList(raw)
}

func parseAllowedRootsList(raw string) []string {
	fields := strings.FieldsFunc(strings.TrimSpace(raw), func(r rune) bool {
		return r == ';' || r == ',' || r == '\n' || r == '\r'
	})
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if trimmed := strings.TrimSpace(field); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func normalizeAllowedRoots(roots []string) []string {
	normalized := make([]string, 0, len(roots))
	for _, root := range roots {
		if trimmed := strings.TrimSpace(root); trimmed != "" {
			normalized = append(normalized, normalizeToolPath(trimmed))
		}
	}
	return normalized
}

func normalizeToolPath(path string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(path), "/", `\`)
	normalized = strings.TrimRight(normalized, `\`)
	normalized = strings.ToLower(normalized)
	return normalized
}

func validateToolPathAllowed(path string) error {
	normalized := normalizeToolPath(path)
	if normalized == "" {
		return fmt.Errorf("path is required")
	}
	if isDeniedToolPath(path) {
		return fmt.Errorf("access denied for sensitive path %q", path)
	}

	for _, root := range allowedToolRoots() {
		if pathUnderRoot(normalized, root) {
			return nil
		}
	}

	return fmt.Errorf("path %q is outside allowed roots", path)
}

func pathUnderRoot(path, root string) bool {
	root = strings.TrimRight(normalizeToolPath(root), `\`)
	if root == "" {
		return false
	}
	return path == root || strings.HasPrefix(path, root+`\`)
}


func isDeniedToolPath(path string) bool {
	normalized := normalizeToolPath(path)
	base := strings.ToLower(strings.TrimSpace(strings.TrimRight(path, `\\/`)))
	if idx := strings.LastIndexAny(base, `\/`); idx >= 0 {
		base = base[idx+1:]
	}

	if base == ".env" || strings.HasPrefix(base, ".env.") || strings.HasSuffix(base, ".key") || strings.HasSuffix(base, ".pem") {
		return true
	}

	deniedBasenames := map[string]struct{}{
		"cookies":         {},
		"cookie":          {},
		"local state":     {},
		"login data":      {},
		"local storage":   {},
		"session storage": {},
	}
	if _, ok := deniedBasenames[base]; ok {
		return true
	}

	segments := strings.FieldsFunc(normalized, func(r rune) bool {
		return r == '\\' || r == '/'
	})
	for _, segment := range segments {
		seg := strings.TrimSpace(strings.ToLower(segment))
		if seg == "" {
			continue
		}
		if seg == "cookies" || seg == "local state" || seg == "login data" || seg == "browser profiles" || seg == "user data" {
			return true
		}
		if strings.HasPrefix(seg, "profile ") || strings.HasPrefix(seg, "profile_") {
			return true
		}
	}

	return false
}

// ── Unified MCP Tool Dispatcher ─────────────────────────────────────────────

// UnifiedToolCall is the payload for brain_execute / brain_batch_execute.
type UnifiedToolCall struct {
	Target    string                 `json:"target"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Access    string                 `json:"access,omitempty"` // "read"|"write"|"destructive"|"auto"
}

// UnifiedToolResult is the structured response from executeUnifiedTool.
type UnifiedToolResult struct {
	Success   bool        `json:"success"`
	Target    string      `json:"target"`
	Name      string      `json:"name"`
	Executor  string      `json:"executor"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	Code      string      `json:"code,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Retryable bool        `json:"retryable"`
}

// executeUnifiedTool is the central dispatcher for brain_execute.
func (m *MCPServerManager) executeUnifiedTool(ctx context.Context, call UnifiedToolCall) *UnifiedToolResult {
	if strings.TrimSpace(call.Target) == "" {
		return &UnifiedToolResult{
			Success:   false,
			Target:    call.Target,
			Name:      call.Name,
			Error:     "target is required",
			Code:      "MISSING_TARGET",
			Retryable: false,
		}
	}
	if strings.TrimSpace(call.Name) == "" {
		return &UnifiedToolResult{
			Success:   false,
			Target:    call.Target,
			Name:      call.Name,
			Error:     "name is required",
			Code:      "MISSING_NAME",
			Retryable: false,
		}
	}

	resolved, err := globalTargetRegistry.Resolve(call.Target)
	if err != nil {
		return &UnifiedToolResult{
			Success:   false,
			Target:    call.Target,
			Name:      call.Name,
			Error:     err.Error(),
			Code:      "TARGET_NOT_FOUND",
			Retryable: false,
		}
	}

	if m.mcpHubClient == nil {
		return &UnifiedToolResult{
			Success:   false,
			Target:    call.Target,
			Name:      call.Name,
			Executor:  resolved.ServerName,
			Error:     "MCP hub client is not configured",
			Code:      "HUB_NOT_CONFIGURED",
			Retryable: false,
		}
	}

	// Determine access policy
	access := inferMCPToolAccess(call.Name)
	if a := strings.TrimSpace(strings.ToLower(call.Access)); a != "" && a != "auto" {
		access = MCPToolAccess(a)
	}

	result, execErr := m.mcpHubClient.CallToolWithAccess(ctx, resolved.ServerName, call.Name, call.Arguments, access)
	if execErr != nil {
		return classifyExecutionError(call, resolved, execErr)
	}

	return &UnifiedToolResult{
		Success:   true,
		Target:    call.Target,
		Name:      call.Name,
		Executor:  resolved.ServerName,
		Result:    result,
		Retryable: false,
	}
}

// classifyExecutionError maps MCP Hub errors to structured error codes.
func classifyExecutionError(call UnifiedToolCall, target LogicalTarget, err error) *UnifiedToolResult {
	msg := err.Error()
	code := "EXECUTION_ERROR"
	retryable := false

	switch {
	case strings.Contains(msg, "auth failed") || strings.Contains(msg, "401") || strings.Contains(msg, "403"):
		code = "AUTH_FAILED"
		retryable = false
	case strings.Contains(msg, "429") || strings.Contains(msg, "rate limit"):
		code = "RATE_LIMITED"
		retryable = true
	case strings.Contains(msg, "502") || strings.Contains(msg, "503") || strings.Contains(msg, "504"):
		code = "SERVER_ERROR"
		retryable = true
	case strings.Contains(msg, "unknown tool") || strings.Contains(msg, "not found"):
		code = "TOOL_NOT_FOUND"
		retryable = false
	case strings.Contains(msg, "unmarshal") || strings.Contains(msg, "decode"):
		code = "DECODE_ERROR"
		retryable = false
	case strings.Contains(msg, "direct tool calls removed"):
		code = "ACCESS_POLICY_VIOLATION"
		retryable = false
	}

	return &UnifiedToolResult{
		Success:   false,
		Target:    call.Target,
		Name:      call.Name,
		Executor:  target.ServerName,
		Error:     msg,
		Code:      code,
		Retryable: retryable,
	}
}

