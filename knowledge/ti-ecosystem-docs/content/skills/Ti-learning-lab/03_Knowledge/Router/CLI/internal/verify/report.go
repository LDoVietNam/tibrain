package verify

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ReportMeta struct {
	PlanID        string `json:"plan_id,omitempty"`
	PlanState     string `json:"plan_state,omitempty"`
	StepID        string `json:"step_id,omitempty"`
	StepTitle     string `json:"step_title,omitempty"`
	NextStepID    string `json:"next_step_id,omitempty"`
	NextStepTitle string `json:"next_step_title,omitempty"`
	Note          string `json:"note,omitempty"`
}

type Report struct {
	Meta      ReportMeta `json:"meta"`
	Summary   Summary    `json:"summary"`
	Results   []Result   `json:"results"`
	Succeeded bool       `json:"succeeded"`
}

func BuildReport(meta ReportMeta, summary Summary, results []Result, runErr error) Report {
	return Report{
		Meta:      meta,
		Summary:   summary,
		Results:   append([]Result(nil), results...),
		Succeeded: runErr == nil,
	}
}

func RenderHuman(report Report) string {
	var b strings.Builder
	status := "FAILED"
	if report.Succeeded {
		status = "PASSED"
	}
	fmt.Fprintf(&b, "ti verify — %s\n", status)
	fmt.Fprintf(&b, "--------------------\n")
	if report.Meta.StepID != "" || report.Meta.StepTitle != "" {
		fmt.Fprintf(&b, "Step: %s", strings.TrimSpace(report.Meta.StepID))
		if report.Meta.StepTitle != "" {
			fmt.Fprintf(&b, " — %s", report.Meta.StepTitle)
		}
		fmt.Fprintf(&b, "\n")
	}
	if report.Meta.PlanID != "" {
		fmt.Fprintf(&b, "Plan: %s", report.Meta.PlanID)
		if report.Meta.PlanState != "" {
			fmt.Fprintf(&b, " (%s)", report.Meta.PlanState)
		}
		fmt.Fprintf(&b, "\n")
	}
	fmt.Fprintf(&b, "Commands: %d | Passed: %d | Failed: %d\n", report.Summary.Commands, report.Summary.Passed, report.Summary.Failed)
	if report.Meta.Note != "" {
		fmt.Fprintf(&b, "Summary: %s\n", strings.TrimSpace(report.Meta.Note))
	}
	if report.Meta.NextStepID != "" {
		fmt.Fprintf(&b, "Next step: %s", report.Meta.NextStepID)
		if report.Meta.NextStepTitle != "" {
			fmt.Fprintf(&b, " — %s", report.Meta.NextStepTitle)
		}
		fmt.Fprintf(&b, "\n")
	}
	fmt.Fprintf(&b, "\nResults:\n")
	for _, res := range report.Results {
		mark := "✅"
		if !res.Passed {
			mark = "❌"
		}
		fmt.Fprintf(&b, "%s %s (%s)\n", mark, res.Command, res.Duration.Round(time.Millisecond))
		if strings.TrimSpace(res.Output) != "" {
			fmt.Fprintf(&b, "    %s\n", indentWrapped(strings.TrimSpace(res.Output), "    "))
		}
		if strings.TrimSpace(res.ErrorMessage) != "" {
			fmt.Fprintf(&b, "    error: %s\n", strings.TrimSpace(res.ErrorMessage))
		}
	}
	return strings.TrimSpace(b.String())
}

func RenderJSON(report Report) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func indentWrapped(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = prefix + strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}
