package commitflow

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ti/cli/internal/permission"
)

type FileStatus struct {
	Code string `json:"code"`
	Path string `json:"path"`
}

type Summary struct {
	Files []FileStatus `json:"files"`
	Count int          `json:"count"`
}

type CommitReport struct {
	Message  string   `json:"message"`
	DryRun   bool     `json:"dry_run"`
	Commands []string `json:"commands,omitempty"`
}

type Runner struct {
	Policy  *permission.Policy
	WorkDir string
	Timeout time.Duration
}

func NewRunner(workDir string, policy *permission.Policy) *Runner {
	return &Runner{WorkDir: strings.TrimSpace(workDir), Policy: policy, Timeout: 2 * time.Minute}
}

func ParsePorcelain(output string) Summary {
	out := Summary{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r\n")
		if strings.TrimSpace(line) == "" || len(line) < 4 {
			continue
		}
		fs := FileStatus{Code: strings.TrimSpace(line[:2]), Path: strings.TrimSpace(line[3:])}
		out.Files = append(out.Files, fs)
		out.Count++
	}
	return out
}

func GenerateMessage(stepTitle, taskType string, summary Summary) string {
	scope := strings.TrimSpace(stepTitle)
	if scope == "" {
		scope = strings.TrimSpace(taskType)
	}
	if scope == "" {
		scope = "update"
	}
	return fmt.Sprintf("ti: %s (%d files)", strings.ToLower(scope), summary.Count)
}

func (r *Runner) Status() (Summary, error) {
	out, err := r.runShell("git status --porcelain")
	if err != nil {
		return Summary{}, err
	}
	return ParsePorcelain(out), nil
}

func (r *Runner) Commit(message string, dryRun bool) (*CommitReport, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, fmt.Errorf("commit message is required")
	}
	cmds := []string{"git add -A", fmt.Sprintf("git commit -m %q", message)}
	if dryRun {
		return &CommitReport{Message: message, DryRun: true, Commands: cmds}, nil
	}
	for _, cmd := range cmds {
		if err := r.check(cmd); err != nil {
			return nil, err
		}
		if _, err := r.runShell(cmd); err != nil {
			return nil, err
		}
	}
	return &CommitReport{Message: message, DryRun: false, Commands: cmds}, nil
}

func (r *Runner) check(command string) error {
	if r.Policy == nil {
		return nil
	}
	decision := r.Policy.Evaluate("bash", "", command)
	if decision.Behavior == permission.BehaviorDeny {
		return fmt.Errorf("commit denied by policy: %s", decision.Reason)
	}
	return nil
}

func (r *Runner) runShell(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-lc", command)
	}
	if r.WorkDir != "" {
		cmd.Dir = r.WorkDir
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}
