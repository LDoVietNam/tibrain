package taskinput

import (
	"fmt"
	"strings"

	"github.com/ti/cli/internal/contextpack"
	"github.com/ti/cli/internal/promptnorm"
	"github.com/ti/cli/internal/repoindex"
	"github.com/ti/cli/internal/ticore"
)

type Options struct {
	Goal                 string
	Project              string
	TaskType             string
	Phase                string
	Route                string
	Effort               string
	RequestedModel       string
	SelectionMode        string
	PreferredPromptShape string
	BaseSystem           []string
	ApprovedPlan         *ticore.ApprovedPlan
	Step                 *ticore.ApprovedStep
	Summary              *repoindex.Summary
	MaxFiles             int
}

type Unified struct {
	Goal                 string              `json:"goal"`
	ExecutionGoal        string              `json:"execution_goal,omitempty"`
	Project              string              `json:"project,omitempty"`
	TaskType             string              `json:"task_type,omitempty"`
	Phase                string              `json:"phase,omitempty"`
	Route                string              `json:"route,omitempty"`
	Effort               string              `json:"effort,omitempty"`
	PreferredPromptShape string              `json:"preferred_prompt_shape,omitempty"`
	RequestedModel       string              `json:"requested_model,omitempty"`
	ModelSelectionMode   string              `json:"model_selection_mode,omitempty"`
	ApprovedPlanID       string              `json:"approved_plan_id,omitempty"`
	ApprovedPlanSummary  string              `json:"approved_plan_summary,omitempty"`
	CurrentStepID        string              `json:"current_step_id,omitempty"`
	CurrentStepTitle     string              `json:"current_step_title,omitempty"`
	CandidateFiles       []string            `json:"candidate_files,omitempty"`
	Validation           []string            `json:"validation,omitempty"`
	Bundle               *contextpack.Bundle `json:"context_bundle,omitempty"`
	Prompt               *promptnorm.Result  `json:"prompt,omitempty"`
	Metadata             map[string]string   `json:"metadata,omitempty"`
}

func Build(opts Options) *Unified {
	goal := strings.TrimSpace(opts.Goal)
	if goal == "" {
		goal = inferGoal(opts.ApprovedPlan, opts.Step)
	}
	project := strings.TrimSpace(opts.Project)
	if project == "" && opts.ApprovedPlan != nil {
		project = strings.TrimSpace(opts.ApprovedPlan.Project)
	}
	taskType := normalizeTaskType(opts.TaskType, opts.Phase, opts.ApprovedPlan)
	requestedModel := strings.TrimSpace(opts.RequestedModel)
	selectionMode := strings.TrimSpace(opts.SelectionMode)
	executionGoal := strings.TrimSpace(goal)
	if opts.Step != nil {
		executionGoal = buildStepGoal(opts.ApprovedPlan, opts.Step, goal)
	}
	candidateFiles := mergeFiles(opts.ApprovedPlan, opts.Step)
	validation := approvedValidation(opts.ApprovedPlan)
	bundle := contextpack.Build(contextpack.Input{
		Goal:                 executionGoal,
		TaskType:             taskType,
		Project:              project,
		CandidateFiles:       candidateFiles,
		Validation:           validation,
		Summary:              opts.Summary,
		MaxFiles:             opts.MaxFiles,
		PreferredPromptShape: opts.PreferredPromptShape,
	})
	approvedSummary := ""
	approvedID := ""
	if opts.ApprovedPlan != nil {
		approvedID = opts.ApprovedPlan.ID
		approvedSummary = opts.ApprovedPlan.Summary
	}
	prompt := promptnorm.Transform(promptnorm.Request{
		Goal:                 executionGoal,
		TaskType:             taskType,
		Phase:                strings.TrimSpace(opts.Phase),
		Route:                strings.TrimSpace(opts.Route),
		Project:              project,
		Effort:               strings.TrimSpace(opts.Effort),
		ApprovedPlanNote:     approvedSummary,
		BaseSystem:           append([]string(nil), opts.BaseSystem...),
		Bundle:               bundle,
		PreferredPromptShape: strings.TrimSpace(opts.PreferredPromptShape),
	})
	out := &Unified{
		Goal:                 goal,
		ExecutionGoal:        executionGoal,
		Project:              project,
		TaskType:             taskType,
		Phase:                strings.TrimSpace(opts.Phase),
		Route:                strings.TrimSpace(opts.Route),
		Effort:               strings.TrimSpace(opts.Effort),
		PreferredPromptShape: strings.TrimSpace(opts.PreferredPromptShape),
		ApprovedPlanID:       approvedID,
		ApprovedPlanSummary:  approvedSummary,
		CandidateFiles:       append([]string(nil), candidateFiles...),
		Validation:           append([]string(nil), validation...),
		Bundle:               bundle,
		Prompt:               prompt,
		RequestedModel:       requestedModel,
		ModelSelectionMode:   selectionMode,
		Metadata:             map[string]string{},
	}
	if opts.Step != nil {
		out.CurrentStepID = opts.Step.ID
		out.CurrentStepTitle = opts.Step.Title
		out.Metadata["step_status"] = string(opts.Step.Status)
	}
	if out.Route != "" {
		out.Metadata["route"] = out.Route
	}
	if out.Phase != "" {
		out.Metadata["phase"] = out.Phase
	}
	return out
}

func inferGoal(plan *ticore.ApprovedPlan, step *ticore.ApprovedStep) string {
	if step != nil {
		return buildStepGoal(plan, step, "")
	}
	if plan != nil {
		return strings.TrimSpace(plan.Goal)
	}
	return "Proceed with the requested task."
}

func normalizeTaskType(taskType, phase string, plan *ticore.ApprovedPlan) string {
	taskType = strings.TrimSpace(taskType)
	if taskType != "" {
		return taskType
	}
	if plan != nil && strings.TrimSpace(plan.TaskType) != "" {
		return strings.TrimSpace(plan.TaskType)
	}
	phase = strings.ToLower(strings.TrimSpace(phase))
	switch phase {
	case "plan", "design":
		return "plan"
	case "implement", "code", "review", "debug", "fix", "refactor":
		return "code"
	case "":
		return "chat"
	default:
		return phase
	}
}

func mergeFiles(plan *ticore.ApprovedPlan, step *ticore.ApprovedStep) []string {
	seen := map[string]bool{}
	out := []string{}
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			return
		}
		seen[v] = true
		out = append(out, v)
	}
	if step != nil {
		for _, f := range step.Files {
			add(f)
		}
	}
	if plan != nil {
		for _, f := range plan.CandidateFiles {
			add(f)
		}
	}
	return out
}

func approvedValidation(plan *ticore.ApprovedPlan) []string {
	if plan == nil {
		return nil
	}
	return append([]string(nil), plan.Validation...)
}

func buildStepGoal(plan *ticore.ApprovedPlan, step *ticore.ApprovedStep, extra string) string {
	if step == nil {
		if plan != nil {
			return strings.TrimSpace(plan.Goal)
		}
		return strings.TrimSpace(extra)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Execute approved Ti plan step %s — %s.\n", step.ID, step.Title)
	if strings.TrimSpace(step.Objective) != "" {
		fmt.Fprintf(&b, "Objective: %s\n", strings.TrimSpace(step.Objective))
	}
	if len(step.Files) > 0 {
		fmt.Fprintf(&b, "Focus files: %s\n", strings.Join(step.Files, ", "))
	}
	if len(step.Commands) > 0 {
		fmt.Fprintf(&b, "Suggested commands: %s\n", strings.Join(step.Commands, "; "))
	}
	if len(step.ExitCriteria) > 0 {
		fmt.Fprintf(&b, "Exit criteria: %s\n", strings.Join(step.ExitCriteria, "; "))
	}
	if plan != nil && strings.TrimSpace(plan.Summary) != "" {
		fmt.Fprintf(&b, "Plan summary: %s", strings.TrimSpace(plan.Summary))
	}
	if strings.TrimSpace(extra) != "" {
		fmt.Fprintf(&b, "\n\nAdditional operator note: %s", strings.TrimSpace(extra))
	}
	return strings.TrimSpace(b.String())
}
