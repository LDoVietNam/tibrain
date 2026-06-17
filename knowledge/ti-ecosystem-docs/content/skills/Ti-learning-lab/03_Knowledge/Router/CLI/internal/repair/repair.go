package repair

import (
	"fmt"
	"strings"

	"github.com/ti/cli/internal/ticore"
)

type Context struct {
	Plan           *ticore.ApprovedPlan
	Step           *ticore.ApprovedStep
	FailureSummary string
	VerifyCommands []string
	LastErrorClass string
}

func BuildSystemPrompt(ctx Context) string {
	parts := []string{
		"You are the Ti repair loop.",
		"Focus on the smallest safe correction that addresses the failing verification outcome.",
		"Do not broaden scope or rewrite unrelated modules.",
	}
	if ctx.Step != nil && len(ctx.Step.ExitCriteria) > 0 {
		parts = append(parts, "Respect the current step exit criteria: "+strings.Join(ctx.Step.ExitCriteria, "; "))
	}
	parts = append(parts, `Return JSON patch ops in this exact shape when you can propose a concrete safe edit: {"reason":"...","ops":[{"path":"...","old":"...","new":"...","replace_all":false}]}. If you cannot safely produce patch ops, explain why briefly.`)
	return strings.Join(parts, " ")
}

func BuildPrompt(ctx Context) string {
	var b strings.Builder
	if ctx.Plan != nil {
		fmt.Fprintf(&b, "Plan: %s\n", ctx.Plan.Goal)
	}
	if ctx.Step != nil {
		fmt.Fprintf(&b, "Repair step %s — %s\n", ctx.Step.ID, ctx.Step.Title)
		if ctx.Step.Objective != "" {
			fmt.Fprintf(&b, "Objective: %s\n", ctx.Step.Objective)
		}
		if len(ctx.Step.Files) > 0 {
			fmt.Fprintf(&b, "Files: %s\n", strings.Join(ctx.Step.Files, ", "))
		}
	}
	if strings.TrimSpace(ctx.FailureSummary) != "" {
		fmt.Fprintf(&b, "Verification failure summary: %s\n", strings.TrimSpace(ctx.FailureSummary))
	}
	if len(ctx.VerifyCommands) > 0 {
		fmt.Fprintf(&b, "Verification commands: %s\n", strings.Join(ctx.VerifyCommands, "; "))
	}
	if strings.TrimSpace(ctx.LastErrorClass) != "" {
		fmt.Fprintf(&b, "Error class: %s\n", strings.TrimSpace(ctx.LastErrorClass))
	}
	fmt.Fprintf(&b, "Return a compact repair plan and, when useful, a minimal patch strategy or replace-ops proposal.\n")
	return strings.TrimSpace(b.String())
}
