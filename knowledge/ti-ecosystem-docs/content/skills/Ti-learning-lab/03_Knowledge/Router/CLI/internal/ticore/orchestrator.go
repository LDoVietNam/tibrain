package ticore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ti/cli/internal/contextpack"
	"github.com/ti/cli/internal/permission"
	"github.com/ti/cli/internal/planner"
	"github.com/ti/cli/internal/promptnorm"
	"github.com/ti/cli/internal/repoindex"
)

type ApprovedPlan struct {
	ID               string         `json:"id"`
	Goal             string         `json:"goal"`
	TaskType         string         `json:"task_type"`
	Project          string         `json:"project"`
	State            PlanState      `json:"state"`
	CurrentStepID    string         `json:"current_step_id,omitempty"`
	CandidateFiles   []string       `json:"candidate_files,omitempty"`
	Validation       []string       `json:"validation,omitempty"`
	Risks            []string       `json:"risks,omitempty"`
	Summary          string         `json:"summary"`
	RouteHints       []string       `json:"route_hints,omitempty"`
	Steps            []ApprovedStep `json:"steps,omitempty"`
	BeadsRecommended bool           `json:"beads_recommended"`
	ApprovedAt       time.Time      `json:"approved_at"`
}

type RuntimeEnvelope struct {
	ApprovedPlan *ApprovedPlan       `json:"approved_plan,omitempty"`
	Context      *contextpack.Bundle `json:"context,omitempty"`
	Prompt       *promptnorm.Result  `json:"prompt,omitempty"`
}

type Orchestrator struct {
	project string
	policy  *permission.Policy
}

func New(project string, policy *permission.Policy) *Orchestrator {
	return &Orchestrator{project: strings.TrimSpace(project), policy: policy}
}

func (o *Orchestrator) PolicyMode() string {
	if o.policy == nil {
		return ""
	}
	return string(o.policy.Mode)
}

func (o *Orchestrator) ApprovePlan(plan *planner.Plan) *ApprovedPlan {
	if plan == nil {
		return nil
	}
	approved := &ApprovedPlan{
		Goal:             plan.Goal,
		TaskType:         plan.TaskType,
		Project:          plan.Project,
		CandidateFiles:   append([]string(nil), plan.CandidateFiles...),
		Validation:       append([]string(nil), plan.Validation...),
		Risks:            append([]string(nil), plan.Risks...),
		Summary:          summarizePlan(plan),
		RouteHints:       []string{"quality-first-routing", "single-executor"},
		Steps:            stepsFromPlanner(plan.Steps),
		BeadsRecommended: len(plan.Steps) >= 5 || len(plan.CandidateFiles) >= 6,
		ApprovedAt:       time.Now(),
		State:            PlanStateApproved,
	}
	approved.ID = hashPlan(approved.Goal, approved.Project, approved.TaskType, approved.Summary)
	return approved
}

func (o *Orchestrator) ApproveCouncil(result *planner.CouncilResult) *ApprovedPlan {
	if result == nil {
		return nil
	}
	if result.Approved != nil {
		approved := &ApprovedPlan{
			Goal:             result.Approved.Goal,
			TaskType:         result.Approved.TaskType,
			Project:          result.Approved.Project,
			CandidateFiles:   append([]string(nil), result.Approved.CandidateFiles...),
			Validation:       append([]string(nil), result.Approved.Validation...),
			Risks:            append([]string(nil), result.Approved.Risks...),
			Summary:          result.Approved.Summary,
			RouteHints:       append([]string(nil), result.Approved.RouteHints...),
			Steps:            stepsFromPlanner(result.Approved.Steps),
			BeadsRecommended: result.Approved.BeadsRecommended,
			ApprovedAt:       result.Approved.ApprovedAt,
			State:            PlanStateApproved,
		}
		approved.ID = hashPlan(approved.Goal, approved.Project, approved.TaskType, approved.Summary)
		return approved
	}
	return o.ApprovePlan(result.Planner)
}

func (o *Orchestrator) PrepareRuntime(goal, phase, route, effort string, approved *ApprovedPlan, summary *repoindex.Summary) *RuntimeEnvelope {
	taskType := inferTaskType(phase, approved)
	project := o.project
	if approved != nil && approved.Project != "" {
		project = approved.Project
	}
	bundle := contextpack.Build(contextpack.Input{Goal: goal, TaskType: taskType, Project: project, CandidateFiles: approvedFiles(approved), Validation: approvedValidation(approved), Summary: summary, MaxFiles: 6})
	transformed := promptnorm.Transform(promptnorm.Request{Goal: goal, TaskType: taskType, Phase: phase, Route: route, Project: project, Effort: effort, ApprovedPlanNote: approvedSummary(approved), Bundle: bundle})
	return &RuntimeEnvelope{ApprovedPlan: approved, Context: bundle, Prompt: transformed}
}

func summarizePlan(plan *planner.Plan) string {
	if len(plan.Steps) == 0 {
		return strings.TrimSpace(plan.Goal)
	}
	titles := make([]string, 0, len(plan.Steps))
	for _, step := range plan.Steps {
		if step.Title != "" {
			titles = append(titles, step.Title)
		}
		if len(titles) == 3 {
			break
		}
	}
	if len(titles) == 0 {
		return strings.TrimSpace(plan.Goal)
	}
	return fmt.Sprintf("%s. Key stages: %s.", strings.TrimSpace(plan.Goal), strings.Join(titles, ", "))
}

func hashPlan(parts ...string) string {
	h := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(h[:8])
}

func inferTaskType(phase string, approved *ApprovedPlan) string {
	if approved != nil && approved.TaskType != "" {
		return approved.TaskType
	}
	phase = strings.TrimSpace(strings.ToLower(phase))
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

func approvedFiles(plan *ApprovedPlan) []string {
	if plan == nil {
		return nil
	}
	return append([]string(nil), plan.CandidateFiles...)
}

func approvedValidation(plan *ApprovedPlan) []string {
	if plan == nil {
		return nil
	}
	return append([]string(nil), plan.Validation...)
}

func approvedSummary(plan *ApprovedPlan) string {
	if plan == nil {
		return ""
	}
	return plan.Summary
}
