// Package mcp - router tools (router.*) exposed via MCP.
package mcp

import (
	"fmt"

	"github.com/ti/cli/internal/router"
)

// RouterTools returns MCP tools that operate on the supplied router client.
// Tools include model listing and provider/model auto-selection.
func RouterTools(client *router.Client) []Tool {
	return []Tool{
		{
			Name:        "router.models",
			Description: "List available models from the router.",
			InputSchema: objSchema(nil, nil),
			Handler: func(args map[string]any) (any, error) {
				if client == nil {
					return nil, fmt.Errorf("router client not configured")
				}
				return []string{}, nil
			},
		},
		{
			Name:        "router.select_auto",
			Description: "Pick best provider+model for a task type using auto-combo scoring.",
			InputSchema: objSchema(map[string]any{
				"task_type": strProp("Task type, e.g. coding, review, plan"),
			}, []string{"task_type"}),
			Handler: func(args map[string]any) (any, error) {
				taskType := strArg(args, "task_type")
				if taskType == "" {
					return nil, fmt.Errorf("task_type is required")
				}
				provider, model := router.SelectAutoModel(taskType, nil)
				return map[string]string{"provider": provider, "model": model}, nil
			},
		},
	}
}
