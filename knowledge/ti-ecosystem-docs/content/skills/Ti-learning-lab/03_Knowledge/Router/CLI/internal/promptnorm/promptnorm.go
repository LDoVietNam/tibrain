package promptnorm

import (
	"fmt"
	"strings"

	"github.com/ti/cli/internal/contextpack"
	"github.com/ti/cli/internal/normalize"
)

type Request struct {
	Goal                 string
	TaskType             string
	Phase                string
	Route                string
	Project              string
	Effort               string
	ApprovedPlanNote     string
	BaseSystem           []string
	Bundle               *contextpack.Bundle
	PreferredPromptShape string
}

type Result struct {
	SystemPrompt string  `json:"system_prompt"`
	UserPrompt   string  `json:"user_prompt"`
	PromptShape  string  `json:"prompt_shape"`
	Savings      float64 `json:"savings"`
}

func Transform(req Request) *Result {
	route := strings.TrimSpace(strings.ToLower(req.Route))
	if route == "" {
		route = "cliproxyapi"
	}
	normalizer := normalize.NewNormalizer().WithMaxPromptLen(12000).WithMaxLineLen(800)
	userGoal := strings.TrimSpace(req.Goal)
	if userGoal == "" {
		userGoal = "Proceed with the current approved task."
	}
	goalNorm := normalizer.Normalize(userGoal)
	ctxText := ""
	if req.Bundle != nil {
		ctxText = req.Bundle.Render()
	}
	ctxNorm := normalizer.Normalize(ctxText)
	shape := choosePromptShape(route, req.TaskType, req.PreferredPromptShape)
	sys := buildSystemPrompt(route, req, shape)
	user := buildUserPrompt(goalNorm.Normalized, ctxNorm.Normalized, req, shape)
	return &Result{SystemPrompt: sys, UserPrompt: user, PromptShape: shape, Savings: goalNorm.Savings + ctxNorm.Savings}
}

func buildSystemPrompt(route string, req Request, shape string) string {
	base := []string{}
	for _, s := range req.BaseSystem {
		s = strings.TrimSpace(s)
		if s != "" {
			base = append(base, s)
		}
	}
	switch route {
	case "sharedchat":
		base = append(base,
			"You are the execution layer behind Ti. Keep responses direct, practical, and implementation-focused.",
			"Preserve user intent exactly, but avoid repeating unnecessary context.",
			"Prefer concise final answers unless detail is needed to complete the task.",
		)
	default:
		base = append(base,
			"You are the execution layer behind Ti using the CLIProxyAPI route.",
			"Produce structured, implementation-ready responses with clear next actions.",
			"Use the provided context bundle instead of asking to rescan the repository unless required.",
		)
	}
	if shape != "" {
		base = append(base, fmt.Sprintf("Preferred prompt shape from learned history: %s.", shape))
	}
	if req.TaskType != "" {
		base = append(base, fmt.Sprintf("Task type: %s.", req.TaskType))
	}
	if req.Phase != "" {
		base = append(base, fmt.Sprintf("Workflow phase: %s.", req.Phase))
	}
	if req.Effort != "" {
		base = append(base, fmt.Sprintf("Reasoning effort target: %s.", req.Effort))
	}
	return strings.Join(base, "\n")
}

func buildUserPrompt(goal, context string, req Request, shape string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Goal:\n%s\n", strings.TrimSpace(goal))
	if req.Project != "" {
		fmt.Fprintf(&b, "\nProject: %s\n", req.Project)
	}
	if req.ApprovedPlanNote != "" {
		fmt.Fprintf(&b, "\nApproved plan:\n%s\n", strings.TrimSpace(req.ApprovedPlanNote))
	}
	if shape != "" {
		fmt.Fprintf(&b, "\nPrompt shape target: %s\n", shape)
	}
	if strings.TrimSpace(context) != "" {
		fmt.Fprintf(&b, "\nContext bundle:\n%s\n", strings.TrimSpace(context))
	}
	fmt.Fprintf(&b, "\nReturn only what is useful for this step. Avoid repeating unchanged context.\n")
	return strings.TrimSpace(b.String())
}

func choosePromptShape(route, taskType, preferred string) string {
	preferred = strings.TrimSpace(strings.ToLower(preferred))
	if preferred != "" {
		return preferred
	}
	return promptShape(route, taskType)
}

func promptShape(route, taskType string) string {
	taskType = strings.TrimSpace(strings.ToLower(taskType))
	switch route {
	case "sharedchat":
		if taskType == "plan" {
			return "concise-plan"
		}
		return "concise-execution"
	default:
		if taskType == "plan" {
			return "structured-plan"
		}
		return "structured-execution"
	}
}
