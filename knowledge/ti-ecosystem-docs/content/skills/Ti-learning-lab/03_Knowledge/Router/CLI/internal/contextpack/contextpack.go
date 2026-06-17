package contextpack

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ti/cli/internal/repoindex"
)

type Input struct {
	Goal                 string
	TaskType             string
	Project              string
	CandidateFiles       []string
	Validation           []string
	Summary              *repoindex.Summary
	MaxFiles             int
	PreferredPromptShape string
}

type Bundle struct {
	Goal                 string   `json:"goal"`
	TaskType             string   `json:"task_type,omitempty"`
	Project              string   `json:"project,omitempty"`
	CandidateFiles       []string `json:"candidate_files,omitempty"`
	RepoHints            []string `json:"repo_hints,omitempty"`
	Validation           []string `json:"validation,omitempty"`
	PreferredPromptShape string   `json:"preferred_prompt_shape,omitempty"`
}

func Build(input Input) *Bundle {
	b := &Bundle{Goal: strings.TrimSpace(input.Goal), TaskType: strings.TrimSpace(input.TaskType), Project: strings.TrimSpace(input.Project), PreferredPromptShape: strings.TrimSpace(input.PreferredPromptShape)}
	maxFiles := input.MaxFiles
	if maxFiles <= 0 {
		maxFiles = preferredMaxFiles(input.PreferredPromptShape, input.TaskType)
	}
	b.CandidateFiles = dedupe(input.CandidateFiles)
	if len(b.CandidateFiles) == 0 && input.Summary != nil {
		relevant := repoindex.SearchRelevantFiles(input.Summary, input.Goal, maxFiles)
		for _, f := range relevant {
			b.CandidateFiles = append(b.CandidateFiles, f.Path)
		}
	}
	if len(b.CandidateFiles) > maxFiles {
		b.CandidateFiles = b.CandidateFiles[:maxFiles]
	}
	b.Validation = dedupe(input.Validation)
	b.RepoHints = buildRepoHints(input.Summary, input.PreferredPromptShape)
	return b
}

func (b *Bundle) Render() string {
	if b == nil {
		return ""
	}
	var parts []string
	if b.Project != "" {
		parts = append(parts, fmt.Sprintf("Project: %s", b.Project))
	}
	if b.TaskType != "" {
		parts = append(parts, fmt.Sprintf("Task type: %s", b.TaskType))
	}
	if len(b.RepoHints) > 0 {
		parts = append(parts, "Repo hints:")
		for _, h := range b.RepoHints {
			parts = append(parts, "- "+h)
		}
	}
	if len(b.CandidateFiles) > 0 {
		parts = append(parts, "Candidate files:")
		for _, f := range b.CandidateFiles {
			parts = append(parts, "- "+f)
		}
	}
	if b.PreferredPromptShape != "" {
		parts = append(parts, fmt.Sprintf("Learned prompt preference: %s", b.PreferredPromptShape))
	}
	if len(b.Validation) > 0 {
		parts = append(parts, "Validation hints:")
		for _, v := range b.Validation {
			parts = append(parts, "- "+v)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func buildRepoHints(summary *repoindex.Summary, preferredPromptShape string) []string {
	if summary == nil {
		return nil
	}
	hints := []string{}
	if summary.Module != "" {
		hints = append(hints, fmt.Sprintf("Go module `%s`", summary.Module))
	}
	if len(summary.EntryPoints) > 0 {
		hints = append(hints, fmt.Sprintf("Likely entry points: %s", strings.Join(summary.EntryPoints[:min(4, len(summary.EntryPoints))], ", ")))
	}
	maxDirs := 5
	if strings.Contains(strings.ToLower(preferredPromptShape), "concise") {
		maxDirs = 3
	}
	if len(summary.TopDirectories) > 0 {
		hints = append(hints, fmt.Sprintf("Top directories: %s", strings.Join(summary.TopDirectories[:min(maxDirs, len(summary.TopDirectories))], ", ")))
	}
	if summary.TestFileCount > 0 {
		hints = append(hints, fmt.Sprintf("Detected %d test files", summary.TestFileCount))
	}
	return hints
}

func dedupe(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func preferredMaxFiles(promptShape, taskType string) int {
	promptShape = strings.ToLower(strings.TrimSpace(promptShape))
	switch {
	case strings.Contains(promptShape, "concise"):
		return 4
	case strings.Contains(promptShape, "structured"):
		return 7
	}
	taskType = strings.ToLower(strings.TrimSpace(taskType))
	switch taskType {
	case "plan":
		return 5
	case "code", "implement", "review":
		return 6
	default:
		return 6
	}
}
