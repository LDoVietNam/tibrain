package mcp

import (
	"fmt"

	"github.com/ti/cli/internal/agent"
	"github.com/ti/cli/internal/router"
)

// AgentTools returns the 4 agent.* MCP tools.
func AgentTools(registry *agent.Registry) []Tool {
	return []Tool{
		{
			Name:        "agent.list",
			Description: "List all registered agents with their capabilities and status.",
			InputSchema: objSchema(nil, nil),
			Handler: func(args map[string]any) (any, error) {
				return registry.List(), nil
			},
		},
		{
			Name:        "agent.get",
			Description: "Get a single agent by ID.",
			InputSchema: objSchema(map[string]any{
				"id": strProp("Agent ID"),
			}, []string{"id"}),
			Handler: func(args map[string]any) (any, error) {
				id := strArg(args, "id")
				if id == "" {
					return nil, fmt.Errorf("id is required")
				}
				a, ok := registry.Get(id)
				if !ok {
					return nil, fmt.Errorf("agent not found: %s", id)
				}
				return a, nil
			},
		},
		{
			Name:        "agent.register",
			Description: "Register a new agent in the registry.",
			InputSchema: objSchema(map[string]any{
				"id":          strProp("Unique agent ID"),
				"name":        strProp("Human-readable agent name"),
				"description": strProp("Agent description (optional)"),
				"expertise":   arrStrProp("List of expertise tags (e.g. ['coding', 'review'])"),
				"models":      arrStrProp("List of model IDs this agent uses"),
			}, []string{"id", "name", "expertise", "models"}),
			Handler: func(args map[string]any) (any, error) {
				id := strArg(args, "id")
				name := strArg(args, "name")
				expertise := strSliceArg(args, "expertise")
				models := strSliceArg(args, "models")
				if id == "" || name == "" {
					return nil, fmt.Errorf("id and name are required")
				}
				if len(expertise) == 0 {
					return nil, fmt.Errorf("at least one expertise tag is required")
				}
				if len(models) == 0 {
					return nil, fmt.Errorf("at least one model is required")
				}
				a := &agent.Agent{
					ID:          id,
					Name:        name,
					Description: strArg(args, "description"),
					Expertise:   expertise,
					Models:      models,
				}
				if err := registry.Register(a); err != nil {
					return nil, err
				}
				return map[string]any{"ok": true, "id": id}, nil
			},
		},
		{
			Name:        "agent.select_model",
			Description: "Use autoCombo to select the best model for an agent and task type.",
			InputSchema: objSchema(map[string]any{
				"id":   strProp("Agent ID"),
				"task": strProp("Task type: coding | review | summarize | default"),
			}, []string{"id"}),
			Handler: func(args map[string]any) (any, error) {
				id := strArg(args, "id")
				if id == "" {
					return nil, fmt.Errorf("id is required")
				}
				_, ok := registry.Get(id)
				if !ok {
					return nil, fmt.Errorf("agent not found: %s", id)
				}
				task := strArg(args, "task")
				if task == "" {
					task = "default"
				}
				result := router.AutoComboSelect(task)
				return result, nil
			},
		},
	}
}
