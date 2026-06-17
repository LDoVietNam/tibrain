package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ti/router/tibrain/providers"
)

// SubagentState tracks the execution state of an individual worker agent in the swarm.
type SubagentState struct {
	ID        string
	Task      string
	Status    string // "pending", "running", "done", "error"
	ModelUsed string
	Result    string
	Error     error
}

// PlannerResult holds the sub-tasks returned by the Planner Agent.
type PlannerResult struct {
	Tasks []string `json:"tasks"`
}

// SwarmMessages defined for BubbleTea coordination
type SwarmPlanMsg struct {
	Tasks []string
	Error error
}

type SubagentStatusMsg struct {
	ID           string
	Status       string // "running", "done", "error"
	ModelUsed    string
	Result       string
	Error        error
	FallbackLogs []string
}

type SwarmAggregateMsg struct {
	Result string
	Error  error
}

// PlanSwarmTask asks the Planner Agent to break down a complex task.
func PlanSwarmTask(ctx context.Context, cb *providers.CodebuffProvider, pool *FailoverModelPool, query string) SwarmPlanMsg {
	systemPrompt := `You are the Master Planner Agent. Your job is to break down a complex user request into a list of independent, self-contained sub-tasks that can be executed in parallel by separate AI workers.
You MUST output your response ONLY as a JSON object containing a "tasks" array of strings. Do not include any markdown styling, backticks, or explanation.
Example Output:
{"tasks": ["Analyze requirement X", "Write code for Y", "Document Z"]}
`
	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("Break down this task: %s", query)},
	}

	resp, _, err := pool.ChatWithFallback(ctx, cb, RolePlanner, messages, nil)
	if err != nil {
		return SwarmPlanMsg{Error: err}
	}

	// Parse JSON
	// Clean markdown block wrappers if model added them
	cleanResp := strings.TrimSpace(resp)
	if strings.HasPrefix(cleanResp, "```json") {
		cleanResp = strings.TrimPrefix(cleanResp, "```json")
		cleanResp = strings.TrimSuffix(cleanResp, "```")
	} else if strings.HasPrefix(cleanResp, "```") {
		cleanResp = strings.TrimPrefix(cleanResp, "```")
		cleanResp = strings.TrimSuffix(cleanResp, "```")
	}
	cleanResp = strings.TrimSpace(cleanResp)

	var result PlannerResult
	if err := json.Unmarshal([]byte(cleanResp), &result); err != nil {
		// Fallback: split by lines if JSON parsing fails
		lines := strings.Split(resp, "\n")
		var tasks []string
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" && !strings.HasPrefix(l, "{") && !strings.HasPrefix(l, "}") && !strings.HasPrefix(l, "[") {
				tasks = append(tasks, l)
			}
		}
		if len(tasks) == 0 {
			return SwarmPlanMsg{Error: fmt.Errorf("failed to parse planner JSON: %v. Raw response: %s", err, resp)}
		}
		return SwarmPlanMsg{Tasks: tasks}
	}

	return SwarmPlanMsg{Tasks: result.Tasks}
}

// RunWorkerTask executes a single sub-task using a Worker Agent.
func RunWorkerTask(ctx context.Context, cb *providers.CodebuffProvider, pool *FailoverModelPool, id string, task string) SubagentStatusMsg {
	var logs []string
	onFallback := func(failedModel, nextModel string, reason error) {
		logs = append(logs, fmt.Sprintf("⚠️ Worker %s: model %s failed, falling back to %s", id, failedModel, nextModel))
	}

	systemPrompt := `You are a Worker Agent in a Swarm. Execute the assigned sub-task thoroughly and return your results clearly. Be precise and detail-oriented.`
	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: task},
	}

	resp, modelUsed, err := pool.ChatWithFallback(ctx, cb, RoleWorker, messages, onFallback)
	if err != nil {
		return SubagentStatusMsg{
			ID:           id,
			Status:       "error",
			Error:        err,
			FallbackLogs: logs,
		}
	}

	return SubagentStatusMsg{
		ID:           id,
		Status:       "done",
		ModelUsed:    modelUsed,
		Result:       resp,
		FallbackLogs: logs,
	}
}

// AggregateSwarmResults combines worker outputs into a single final report.
func AggregateSwarmResults(ctx context.Context, cb *providers.CodebuffProvider, pool *FailoverModelPool, originalQuery string, subagents []SubagentState) SwarmAggregateMsg {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Original Request: %s\n\n", originalQuery))
	builder.WriteString("Here are the results from all specialized worker agents:\n\n")

	for _, sa := range subagents {
		builder.WriteString(fmt.Sprintf("=== Agent: %s (Model: %s) ===\n", sa.Task, sa.ModelUsed))
		if sa.Status == "error" {
			builder.WriteString(fmt.Sprintf("ERROR: %v\n\n", sa.Error))
		} else {
			builder.WriteString(sa.Result + "\n\n")
		}
	}

	systemPrompt := `You are the Master Aggregator Agent. Your job is to synthesize all worker results into a single clean, coherent, unified response that fully answers the original user request. Do not just paste the outputs; summarize them beautifully and professionally.`
	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: builder.String()},
	}

	resp, _, err := pool.ChatWithFallback(ctx, cb, RoleReviewer, messages, nil)
	if err != nil {
		return SwarmAggregateMsg{Error: err}
	}

	return SwarmAggregateMsg{Result: resp}
}
