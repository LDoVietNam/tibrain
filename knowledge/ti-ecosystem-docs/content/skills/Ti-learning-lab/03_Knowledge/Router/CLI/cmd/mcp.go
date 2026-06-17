package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/ti/cli/internal/agent"
	"github.com/ti/cli/internal/mcp"
	"github.com/ti/cli/internal/session"
	"github.com/ti/cli/internal/tools"
)

var (
	mcpTransport string
	mcpPort      string
	mcpDBPath    string
)

var mcpCmd = &cobra.Command{Use: "mcp", Short: "Run Ti as an MCP server", Long: "Expose Ti session, agent, and local tool capabilities through a Model Context Protocol server."}

var mcpListToolsCmd = &cobra.Command{Use: "tools", Short: "List MCP tools Ti exposes", RunE: func(cmd *cobra.Command, args []string) error {
	serverTools, cleanup, err := buildMCPTools()
	if err != nil { return err }
	defer cleanup()
	for _, t := range serverTools { fmt.Printf("%-22s %s\n", t.Name, t.Description) }
	return nil
}}

var mcpServeCmd = &cobra.Command{Use: "serve", Short: "Start MCP server", RunE: func(cmd *cobra.Command, args []string) error {
	serverTools, cleanup, err := buildMCPTools()
	if err != nil { return err }
	defer cleanup()
	server := mcp.New(serverTools)
	switch mcpTransport {
	case "stdio":
		return server.RunStdio()
	case "http", "sse":
		return server.RunHTTP(mcpPort)
	default:
		return fmt.Errorf("transport must be stdio, http, or sse")
	}
}}

func buildMCPTools() ([]mcp.Tool, func(), error) {
	store, err := session.NewStore(mcpDBPath)
	if err != nil { return nil, func(){}, err }
	cleanup := func(){ _ = store.Close() }
	registry := agent.NewRegistry()
	_ = registry.Register(&agent.Agent{ID: "planner", Name: "Planner", Description: "Read-only planning agent", Expertise: []string{"planning", "review"}, Models: []string{"default"}, DefaultModel: "default"})
	_ = registry.Register(&agent.Agent{ID: "coder", Name: "Coder", Description: "Implementation agent", Expertise: []string{"coding", "debugging"}, Models: []string{"default"}, DefaultModel: "default"})
	out := []mcp.Tool{}
	out = append(out, mcp.SessionTools(store)...)
	out = append(out, mcp.AgentTools(registry)...)
	out = append(out, tiToolMCPBridge()...)
	return out, cleanup, nil
}

func tiToolMCPBridge() []mcp.Tool {
	return []mcp.Tool{
		{Name: "ti.tool.list", Description: "List Ti local agent tools.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{}}, Handler: func(args map[string]any) (any, error) { return tools.AllDefs(), nil }},
		{Name: "ti.tool.run", Description: "Run a Ti local agent tool by name.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"name": map[string]any{"type": "string"}, "input": map[string]any{"type": "object"}}, "required": []string{"name"}}, Handler: func(args map[string]any) (any, error) {
			name, _ := args["name"].(string)
			if name == "" { return nil, fmt.Errorf("name is required") }
			input, _ := args["input"].(map[string]any)
			if input == nil { input = map[string]any{} }
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			return tools.Dispatch(ctx, name, input)
		}},
		{Name: "ti.session.create", Description: "Create a Ti session.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"title": map[string]any{"type": "string"}, "provider": map[string]any{"type": "string"}, "model": map[string]any{"type": "string"}}}, Handler: func(args map[string]any) (any, error) {
			store, err := session.NewStore(mcpDBPath); if err != nil { return nil, err }
			defer store.Close()
			s := &session.Session{Title: stringArg(args, "title"), Provider: stringArg(args, "provider"), Model: stringArg(args, "model")}
			if err := store.Create(s); err != nil { return nil, err }
			data, _ := json.Marshal(s)
			return string(data), nil
		}},
	}
}

func stringArg(args map[string]any, key string) string { v, _ := args[key].(string); return v }

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.PersistentFlags().StringVar(&mcpTransport, "transport", "stdio", "transport: stdio, http, or sse")
	mcpCmd.PersistentFlags().StringVar(&mcpPort, "port", "8765", "HTTP/SSE port")
	mcpCmd.PersistentFlags().StringVar(&mcpDBPath, "db", "", "session database path")
	mcpCmd.AddCommand(mcpListToolsCmd, mcpServeCmd)
}
