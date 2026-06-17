package planner

import (
	"fmt"
	"strings"
	"time"

	"github.com/ti/cli/internal/permission"
	"github.com/ti/cli/internal/repoindex"
)

type Plan struct {
	Goal           string            `json:"goal"`
	TaskType       string            `json:"task_type"`
	Project        string            `json:"project"`
	GeneratedAt    time.Time         `json:"generated_at"`
	Phase          PlanningPhase     `json:"phase"`
	Assumptions    []string          `json:"assumptions,omitempty"`
	CandidateFiles []string          `json:"candidate_files,omitempty"`
	Steps          []Step            `json:"steps"`
	Risks          []string          `json:"risks,omitempty"`
	Validation     []string          `json:"validation,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type PlanningPhase string

const (
	PhaseUnderstanding PlanningPhase = "understanding"
	PhaseDesign        PlanningPhase = "design"
	PhaseReview        PlanningPhase = "review"
	PhaseFinalPlan     PlanningPhase = "final_plan"
)

type Step struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Objective        string   `json:"objective"`
	Commands         []string `json:"commands,omitempty"`
	Files            []string `json:"files,omitempty"`
	PermissionImpact string   `json:"permission_impact,omitempty"`
	ExitCriteria     []string `json:"exit_criteria,omitempty"`
}

func Build(goal string, summary *repoindex.Summary, policy *permission.Policy) *Plan {
	taskType := InferTaskType(goal)
	project := "default"
	if summary != nil && summary.Module != "" {
		project = summary.Module
	}
	if summary != nil && project == "default" {
		project = filepathBase(summary.Root)
	}
	plan := &Plan{
		Goal:        goal,
		TaskType:    taskType,
		Project:     project,
		GeneratedAt: time.Now(),
		Assumptions: []string{
			"Planner uses local repository signals and current permission policy.",
			"Final execution should still validate changes with tests or smoke checks.",
		},
		Metadata: map[string]string{
			"planner":         "ti-local",
			"permission_mode": string(defaultMode(policy)),
		},
	}
	relevant := repoindex.SearchRelevantFiles(summary, goal, 8)
	for _, f := range relevant {
		plan.CandidateFiles = append(plan.CandidateFiles, f.Path)
	}
	plan.Steps = []Step{
		step("S1", "Understand scope", fmt.Sprintf("Map the request to likely modules and entry points for %s work.", taskType), []string{"rg -n \"keyword\" .", "ti repo scan"}, plan.CandidateFiles, permissionHint(policy, false), []string{"Relevant modules identified", "Assumptions documented"}),
		step("S2", "Design change", "Pick the smallest safe change set and note dependencies, interfaces, and rollback plan.", []string{"review config and tests", "note expected side effects"}, plan.CandidateFiles, permissionHint(policy, false), []string{"Concrete implementation plan written", "Affected files agreed"}),
		step("S3", "Implement", "Edit targeted files only, keep patch small, and preserve compatibility where possible.", []string{"apply patch", "run focused checks"}, plan.CandidateFiles, permissionHint(policy, true), []string{"Patch compiles or is syntactically valid", "No unrelated files changed"}),
		step("S4", "Validate", "Run focused validation, then broaden to project-level checks if needed.", validationCommands(summary, taskType), nil, permissionHint(policy, true), []string{"At least one targeted validation passes", "Any failures are explained"}),
		step("S5", "Summarize", "Capture what changed, risks, and next actions.", []string{"create change summary", "list follow-up work"}, nil, permissionHint(policy, false), []string{"Summary includes files, risks, and next steps"}),
	}
	plan.Risks = defaultRisks(summary, taskType, policy)
	plan.Validation = validationCommands(summary, taskType)
	return plan
}

func RenderMarkdown(plan *Plan) string {
	if plan == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Ti Plan\n\n")
	fmt.Fprintf(&b, "- Goal: %s\n", plan.Goal)
	fmt.Fprintf(&b, "- Task type: %s\n", plan.TaskType)
	fmt.Fprintf(&b, "- Project: %s\n", plan.Project)
	fmt.Fprintf(&b, "- Generated: %s\n", plan.GeneratedAt.Format(time.RFC3339))
	if len(plan.CandidateFiles) > 0 {
		fmt.Fprintf(&b, "\n## Candidate files\n")
		for _, f := range plan.CandidateFiles {
			fmt.Fprintf(&b, "- `%s`\n", f)
		}
	}
	fmt.Fprintf(&b, "\n## Steps\n")
	for _, s := range plan.Steps {
		fmt.Fprintf(&b, "\n### %s — %s\n", s.ID, s.Title)
		fmt.Fprintf(&b, "%s\n", s.Objective)
		if len(s.Files) > 0 {
			fmt.Fprintf(&b, "- Files: %s\n", strings.Join(s.Files, ", "))
		}
		if s.PermissionImpact != "" {
			fmt.Fprintf(&b, "- Permission: %s\n", s.PermissionImpact)
		}
		if len(s.Commands) > 0 {
			fmt.Fprintf(&b, "- Suggested commands:\n")
			for _, c := range s.Commands {
				fmt.Fprintf(&b, "  - `%s`\n", c)
			}
		}
		if len(s.ExitCriteria) > 0 {
			fmt.Fprintf(&b, "- Exit criteria:\n")
			for _, c := range s.ExitCriteria {
				fmt.Fprintf(&b, "  - %s\n", c)
			}
		}
	}
	if len(plan.Risks) > 0 {
		fmt.Fprintf(&b, "\n## Risks\n")
		for _, r := range plan.Risks {
			fmt.Fprintf(&b, "- %s\n", r)
		}
	}
	if len(plan.Validation) > 0 {
		fmt.Fprintf(&b, "\n## Validation\n")
		for _, v := range plan.Validation {
			fmt.Fprintf(&b, "- `%s`\n", v)
		}
	}
	return b.String()
}

func step(id, title, objective string, commands, files []string, permissionImpact string, exit []string) Step {
	return Step{ID: id, Title: title, Objective: objective, Commands: commands, Files: files, PermissionImpact: permissionImpact, ExitCriteria: exit}
}

// InferTaskType infers the task type from a goal string.
func InferTaskType(goal string) string {
	g := strings.ToLower(goal)
	switch {
	case strings.Contains(g, "fix") || strings.Contains(g, "bug") || strings.Contains(g, "error"):
		return "bugfix"
	case strings.Contains(g, "refactor") || strings.Contains(g, "cleanup"):
		return "refactor"
	case strings.Contains(g, "test"):
		return "test"
	case strings.Contains(g, "doc") || strings.Contains(g, "readme"):
		return "docs"
	default:
		return "feature"
	}
}

func permissionHint(policy *permission.Policy, writes bool) string {
	mode := defaultMode(policy)
	switch mode {
	case permission.ModeDeny:
		return "Current permission mode denies execution; plan only until rules are relaxed."
	case permission.ModeAsk:
		if writes {
			return "Expect approval prompts before write or shell steps."
		}
		return "Read-only steps are safe; write steps will prompt."
	case permission.ModeAuto, permission.ModeSmart:
		if writes {
			return "Read operations can proceed automatically; writes should be reviewed carefully."
		}
		return "Read-oriented discovery is likely to auto-approve."
	case permission.ModeYes:
		return "Execution is permissive; rely on path rules and review discipline."
	default:
		return "Permission behavior unknown; default to cautious execution."
	}
}

func validationCommands(summary *repoindex.Summary, taskType string) []string {
	cmds := []string{}
	if summary != nil && summary.Module != "" {
		cmds = append(cmds, "go test ./...")
	}
	switch taskType {
	case "docs":
		cmds = append(cmds, "spell/style check if available")
	case "test":
		cmds = append(cmds, "run focused test package first")
	default:
		cmds = append(cmds, "run targeted package or integration check")
	}
	return cmds
}

func defaultRisks(summary *repoindex.Summary, taskType string, policy *permission.Policy) []string {
	risks := []string{"Cross-module changes may require broader validation than the planner can infer locally."}
	if summary != nil && summary.TestFileCount == 0 {
		risks = append(risks, "Repository has few or no tests detected; rely on smoke checks and careful review.")
	}
	if defaultMode(policy) == permission.ModeYes {
		risks = append(risks, "Permission mode is very permissive; accidental writes are more likely if commands are executed automatically.")
	}
	if taskType == "feature" {
		risks = append(risks, "Feature work can drift in scope; keep changes tied to candidate files unless evidence expands the scope.")
	}
	return risks
}

func defaultMode(policy *permission.Policy) permission.PermissionMode {
	if policy == nil || !policy.Mode.IsValid() {
		return permission.ModeAsk
	}
	return policy.Mode
}

func filepathBase(path string) string {
	path = strings.TrimRight(path, "/\\")
	idx := strings.LastIndexAny(path, "/\\")
	if idx == -1 {
		return path
	}
	return path[idx+1:]
}
