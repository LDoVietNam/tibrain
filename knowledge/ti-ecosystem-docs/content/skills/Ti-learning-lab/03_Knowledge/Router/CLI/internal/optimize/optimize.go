package optimize

import (
	"fmt"
	"strings"

	"github.com/ti/cli/internal/ticore"
)

type Context struct {
	Plan              *ticore.ApprovedPlan
	Step              *ticore.ApprovedStep
	VerifySummary     string
	LastAnswerSummary string
	ChangedFiles      []string
}

func BuildSystemPrompt(ctx Context) string {
	return `You are the Ti optimization pass. Improve code cleanliness, patch quality, readability, and reduce unnecessary context or follow-up work without changing scope. Return JSON patch ops in this exact shape when you can propose a concrete safe optimization: {"reason":"...","ops":[{"path":"...","old":"...","new":"...","replace_all":false}]}. If you cannot safely produce patch ops, explain why briefly.`
}

func BuildPrompt(ctx Context) string {
	var b strings.Builder
	if ctx.Plan != nil {
		fmt.Fprintf(&b, "Plan: %s\n", ctx.Plan.Goal)
	}
	if ctx.Step != nil {
		fmt.Fprintf(&b, "Optimize around step %s — %s\n", ctx.Step.ID, ctx.Step.Title)
		if len(ctx.Step.Files) > 0 {
			fmt.Fprintf(&b, "Focus files: %s\n", strings.Join(ctx.Step.Files, ", "))
		}
	}
	if len(ctx.ChangedFiles) > 0 {
		fmt.Fprintf(&b, "Changed files: %s\n", strings.Join(ctx.ChangedFiles, ", "))
	}
	if strings.TrimSpace(ctx.VerifySummary) != "" {
		fmt.Fprintf(&b, "Verify summary: %s\n", strings.TrimSpace(ctx.VerifySummary))
	}
	if strings.TrimSpace(ctx.LastAnswerSummary) != "" {
		fmt.Fprintf(&b, "Last execution summary: %s\n", strings.TrimSpace(ctx.LastAnswerSummary))
	}
	fmt.Fprintf(&b, "Return concrete optimization suggestions and only recommend changes that preserve the approved plan scope.\n")
	return strings.TrimSpace(b.String())
}
