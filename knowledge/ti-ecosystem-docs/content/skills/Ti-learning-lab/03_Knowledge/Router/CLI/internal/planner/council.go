package planner

import (
	"fmt"
	"strings"
	"time"

	"github.com/ti/cli/internal/permission"
	"github.com/ti/cli/internal/repoindex"
)

type CouncilRoleNote struct {
	Role  string   `json:"role"`
	Notes []string `json:"notes,omitempty"`
}

type ApprovedPlan struct {
	Goal             string            `json:"goal"`
	TaskType         string            `json:"task_type"`
	Project          string            `json:"project"`
	ApprovedAt       time.Time         `json:"approved_at"`
	CandidateFiles   []string          `json:"candidate_files,omitempty"`
	Steps            []Step            `json:"steps,omitempty"`
	Risks            []string          `json:"risks,omitempty"`
	Validation       []string          `json:"validation,omitempty"`
	Summary          string            `json:"summary"`
	RouteHints       []string          `json:"route_hints,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	BeadsRecommended bool              `json:"beads_recommended"`
}

type CouncilResult struct {
	Planner  *Plan             `json:"planner"`
	Approved *ApprovedPlan     `json:"approved"`
	Roles    []CouncilRoleNote `json:"roles,omitempty"`
	Phase    PlanningPhase     `json:"phase"`
}

type PlanningFlow struct {
	Goal    string
	Summary *repoindex.Summary
	Policy  *permission.Policy
	State   *Plan
}

func NewPlanningFlow(goal string, summary *repoindex.Summary, policy *permission.Policy) *PlanningFlow {
	return &PlanningFlow{
		Goal:    goal,
		Summary: summary,
		Policy:  policy,
		State: &Plan{
			Goal:        goal,
			GeneratedAt: time.Now(),
			Phase:       PhaseUnderstanding,
		},
	}
}

func (f *PlanningFlow) NextPhase() PlanningPhase {
	switch f.State.Phase {
	case PhaseUnderstanding:
		return PhaseDesign
	case PhaseDesign:
		return PhaseReview
	case PhaseReview:
		return PhaseFinalPlan
	default:
		return PhaseFinalPlan
	}
}

func RunCouncil(goal string, summary *repoindex.Summary, policy *permission.Policy) *CouncilResult {
	plan := Build(goal, summary, policy)
	roles := []CouncilRoleNote{
		{Role: "explorer", Notes: explorerNotes(goal, summary)},
		{Role: "planner", Notes: []string{"Produced a structured execution plan from local repository signals and permission policy."}},
		{Role: "critic", Notes: criticNotes(plan)},
	}
	approved := approve(plan)
	return &CouncilResult{Planner: plan, Approved: approved, Roles: roles}
}

func RenderApprovedMarkdown(result *CouncilResult) string {
	if result == nil || result.Approved == nil {
		return ""
	}
	plan := result.Approved
	var b strings.Builder
	fmt.Fprintf(&b, "# Ti Approved Plan\n\n")
	fmt.Fprintf(&b, "- Goal: %s\n", plan.Goal)
	fmt.Fprintf(&b, "- Task type: %s\n", plan.TaskType)
	fmt.Fprintf(&b, "- Project: %s\n", plan.Project)
	fmt.Fprintf(&b, "- Approved: %s\n", plan.ApprovedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "- Beads recommended: %v\n", plan.BeadsRecommended)
	if len(plan.RouteHints) > 0 {
		fmt.Fprintf(&b, "- Route hints: %s\n", strings.Join(plan.RouteHints, ", "))
	}
	if plan.Summary != "" {
		fmt.Fprintf(&b, "\n## Summary\n%s\n", plan.Summary)
	}
	if len(result.Roles) > 0 {
		fmt.Fprintf(&b, "\n## Council notes\n")
		for _, role := range result.Roles {
			fmt.Fprintf(&b, "\n### %s\n", strings.Title(role.Role))
			for _, note := range role.Notes {
				fmt.Fprintf(&b, "- %s\n", note)
			}
		}
	}
	if len(plan.CandidateFiles) > 0 {
		fmt.Fprintf(&b, "\n## Candidate files\n")
		for _, f := range plan.CandidateFiles {
			fmt.Fprintf(&b, "- `%s`\n", f)
		}
	}
	fmt.Fprintf(&b, "\n## Approved steps\n")
	for _, s := range plan.Steps {
		fmt.Fprintf(&b, "\n### %s — %s\n", s.ID, s.Title)
		fmt.Fprintf(&b, "%s\n", s.Objective)
		if len(s.Files) > 0 {
			fmt.Fprintf(&b, "- Files: %s\n", strings.Join(s.Files, ", "))
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

func explorerNotes(goal string, summary *repoindex.Summary) []string {
	notes := []string{"Start from the smallest believable scope and avoid unnecessary repository-wide context."}
	files := repoindex.SearchRelevantFiles(summary, goal, 5)
	if len(files) > 0 {
		paths := make([]string, 0, len(files))
		for _, f := range files {
			paths = append(paths, f.Path)
		}
		notes = append(notes, "Most relevant files: "+strings.Join(paths, ", "))
	}
	if summary != nil && summary.Module != "" {
		notes = append(notes, fmt.Sprintf("Detected Go module %s.", summary.Module))
	}
	return notes
}

func criticNotes(plan *Plan) []string {
	notes := []string{"Prefer focused validation and avoid broad project-wide retests until targeted checks pass."}
	if len(plan.CandidateFiles) == 0 {
		notes = append(notes, "No candidate files were found; execution should begin with discovery before code changes.")
	}
	if len(plan.Steps) > 4 {
		notes = append(notes, "Plan is broad enough that Beads may help track multi-step execution.")
	}
	return notes
}

func approve(plan *Plan) *ApprovedPlan {
	approved := &ApprovedPlan{Goal: plan.Goal, TaskType: plan.TaskType, Project: plan.Project, ApprovedAt: time.Now(), CandidateFiles: append([]string(nil), plan.CandidateFiles...), Steps: trimSteps(plan.Steps), Risks: dedupeStrings(plan.Risks), Validation: dedupeStrings(plan.Validation), RouteHints: routeHints(plan.TaskType), Metadata: copyMap(plan.Metadata)}
	approved.BeadsRecommended = len(approved.Steps) >= 5 || len(approved.CandidateFiles) >= 6 || approved.TaskType == "feature"
	approved.Summary = buildApprovedSummary(plan)
	return approved
}

func trimSteps(steps []Step) []Step {
	out := make([]Step, 0, len(steps))
	for _, s := range steps {
		s.Commands = dedupeStrings(s.Commands)
		s.Files = dedupeStrings(s.Files)
		s.ExitCriteria = dedupeStrings(s.ExitCriteria)
		out = append(out, s)
	}
	return out
}

func buildApprovedSummary(plan *Plan) string {
	titles := make([]string, 0, len(plan.Steps))
	for _, s := range plan.Steps {
		if s.Title != "" {
			titles = append(titles, strings.ToLower(s.Title))
		}
		if len(titles) == 3 {
			break
		}
	}
	if len(titles) == 0 {
		return strings.TrimSpace(plan.Goal)
	}
	return fmt.Sprintf("Approved plan for %s. Key stages: %s.", strings.TrimSpace(plan.Goal), strings.Join(titles, ", "))
}

func routeHints(taskType string) []string {
	switch taskType {
	case "bugfix", "test", "docs":
		return []string{"single-executor", "quality-first-routing"}
	case "feature", "refactor":
		return []string{"planning-council", "quality-first-routing", "beads-if-scope-expands"}
	default:
		return []string{"single-executor", "quality-first-routing"}
	}
}

func dedupeStrings(values []string) []string {
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
	return out
}

func copyMap(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		out[k] = v
	}
	return out
}
