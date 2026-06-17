package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ti/cli/internal/beadsgraph"
	"github.com/ti/cli/internal/config"
)

// BeadsTools returns MCP tools for Beads integration.
// Handlers proxy to external bd CLI when available, falling back to local graph.
func BeadsTools() []Tool {
	return []Tool{
		makeTool("beads.ready", "Get ready/unblocked issues from Beads graph", readySchema, handleReady),
		makeTool("beads.get", "Get issue details from Beads graph", getSchema, handleGet),
		makeTool("beads.update", "Update an issue (claim, close, change status)", updateSchema, handleUpdate),
	}
}

func makeTool(name, desc string, schema map[string]any, h HandlerFunc) Tool {
	return Tool{
		Name:        name,
		Description: desc,
		InputSchema: schema,
		Handler:     h,
	}
}

var (
	readySchema = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"limit":  map[string]any{"type": "integer", "description": "Max issues to return"},
			"output": map[string]any{"type": "string", "enum": []string{"human", "json"}},
		},
	}
	getSchema = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{"type": "string", "description": "Issue ID"},
		},
		"required": []string{"id"},
	}
	updateSchema = map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id":      map[string]any{"type": "string"},
			"status":  map[string]any{"type": "string", "enum": []string{"open", "in_progress", "closed", "blocked", "deferred", "ready", "review"}},
			"comment": map[string]any{"type": "string"},
		},
		"required": []string{"id"},
	}
)

// Handler functions
type HandlerFunc func(args map[string]any) (any, error)

func handleReady(args map[string]any) (any, error) {
	output := "human"
	if v, ok := args["output"].(string); ok {
		output = v
	}
	limit := 0
	if v, ok := args["limit"].(float64); ok {
		limit = int(v)
	}

	// Try bd first
	if data, err := runBd("ready", "--json"); err == nil {
		var ready []beadsgraph.Issue
		if err := json.Unmarshal(data, &ready); err != nil {
			return nil, fmt.Errorf("parse bd ready: %w", err)
		}
		if limit > 0 && len(ready) > limit {
			ready = ready[:limit]
		}
		if output == "json" {
			return map[string]any{"ready": ready}, nil
		}
		lines := make([]string, 0, len(ready))
		for _, iss := range ready {
			lines = append(lines, fmt.Sprintf("- %s [%s] %s", iss.ID, iss.Status, iss.Title))
		}
		return map[string]any{"ready": lines}, nil
	}

	// Fallback: local graph
	return readyLocal(output, limit)
}

func handleGet(args map[string]any) (any, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("id required")
	}
	if data, err := runBd("get", id); err == nil {
		var iss beadsgraph.Issue
		if err := json.Unmarshal(data, &iss); err != nil {
			return nil, fmt.Errorf("parse bd get: %w", err)
		}
		return map[string]any{
			"id":          iss.ID,
			"title":       iss.Title,
			"objective":   iss.Objective,
			"status":      string(iss.Status),
			"priority":    iss.Priority,
			"depends_on":  iss.DependsOn,
			"files":       iss.Files,
			"commands":    iss.Commands,
			"acceptance":  iss.Acceptance,
			"route_hints": iss.RouteHints,
		}, nil
	}
	return getLocal(id)
}

func handleUpdate(args map[string]any) (any, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("id required")
	}
	status, _ := args["status"].(string)
	comment, _ := args["comment"].(string)

	bdArgs := []string{"update", id}
	if status != "" {
		bdArgs = append(bdArgs, "--status", status)
	}
	if comment != "" {
		bdArgs = append(bdArgs, "--comment", comment)
	}
	if _, err := runBd(bdArgs...); err == nil {
		return map[string]any{"id": id, "status": "ok"}, nil
	}
	return updateLocal(id, status, comment)
}

// Local fallback implementations
func readyLocal(output string, limit int) (any, error) {
	cfg, _ := config.LoadAuto()
	store, err := beadsgraph.NewStore(beadsGraphStorePath(cfg))
	if err != nil {
		return nil, fmt.Errorf("open beads store: %w", err)
	}
	g, err := store.LoadCurrent()
	if err != nil {
		return nil, fmt.Errorf("load beads graph: %w", err)
	}
	if g == nil {
		return map[string]any{"ready": []any{}}, nil
	}
	ready := beadsgraph.ReadyIssues(g)
	if limit > 0 && len(ready) > limit {
		ready = ready[:limit]
	}
	if output == "json" {
		return map[string]any{"ready": ready}, nil
	}
	lines := make([]string, 0, len(ready))
	for _, iss := range ready {
		lines = append(lines, fmt.Sprintf("- %s [%s] %s", iss.ID, iss.Status, iss.Title))
	}
	return map[string]any{"ready": lines}, nil
}

func getLocal(id string) (any, error) {
	cfg, _ := config.LoadAuto()
	store, err := beadsgraph.NewStore(beadsGraphStorePath(cfg))
	if err != nil {
		return nil, fmt.Errorf("open beads store: %w", err)
	}
	g, err := store.LoadCurrent()
	if err != nil {
		return nil, fmt.Errorf("load beads graph: %w", err)
	}
	if g == nil {
		return nil, fmt.Errorf("no beads graph found")
	}
	for _, iss := range g.Issues {
		if iss.ID == id {
			return map[string]any{
				"id":          iss.ID,
				"title":       iss.Title,
				"objective":   iss.Objective,
				"status":      string(iss.Status),
				"priority":    iss.Priority,
				"depends_on":  iss.DependsOn,
				"files":       iss.Files,
				"commands":    iss.Commands,
				"acceptance":  iss.Acceptance,
				"route_hints": iss.RouteHints,
			}, nil
		}
	}
	return nil, fmt.Errorf("issue %s not found", id)
}

func updateLocal(id, status, comment string) (any, error) {
	cfg, _ := config.LoadAuto()
	store, err := beadsgraph.NewStore(beadsGraphStorePath(cfg))
	if err != nil {
		return nil, fmt.Errorf("open beads store: %w", err)
	}
	g, err := store.LoadCurrent()
	if err != nil {
		return nil, fmt.Errorf("load beads graph: %w", err)
	}
	if g == nil {
		return nil, fmt.Errorf("no beads graph found")
	}
	var updated bool
	for i := range g.Issues {
		if g.Issues[i].ID == id {
			if status != "" {
				g.Issues[i].Status = beadsgraph.Status(status)
			}
			if comment != "" {
				g.Issues[i].SyncSummary = comment
			}
			updated = true
			break
		}
	}
	if !updated {
		return nil, fmt.Errorf("issue %s not found", id)
	}
	if err := store.Save(g); err != nil {
		return nil, fmt.Errorf("save graph: %w", err)
	}
	return map[string]any{"id": id, "status": "ok"}, nil
}

// runBd executes bd with given args, returning stdout.
func runBd(args ...string) ([]byte, error) {
	bdBin, err := resolveBDBinary()
	if err != nil {
		return nil, err
	}
	ctx, cancel := contextWithTimeout()
	defer cancel()
	cmd := exec.CommandContext(ctx, bdBin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("bd %v: %w", args, err)
	}
	return out, nil
}

// resolveBDBinary finds the bd executable.
func resolveBDBinary() (string, error) {
	if env := strings.TrimSpace(os.Getenv("TI_BD_BIN")); env != "" {
		if fileExists(env) {
			return env, nil
		}
		return "", fmt.Errorf("TI_BD_BIN set but not found: %s", env)
	}
	if p, err := exec.LookPath("bd"); err == nil {
		return p, nil
	}
	candidates := []string{
		`Z:\02_CORE\_cli\bin\bd.exe`,
		`Z:\02_CORE\_cli\bd.exe`,
		`Z:\01_PROJECTS\MCP\beads-main\bin\bd.exe`,
	}
	if runtime.GOOS != "windows" {
		candidates = append(candidates,
			filepath.Join(os.Getenv("HOME"), ".local", "bin", "bd"),
			"/usr/local/bin/bd",
		)
	}
	for _, c := range candidates {
		if fileExists(c) {
			return c, nil
		}
	}
	return "", fmt.Errorf("bd binary not found — run 'ti bd doctor'")
}

func fileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// contextWithTimeout returns a context with a reasonable timeout for bd calls.
func contextWithTimeout() (context.Context, func()) {
	// 10-second timeout, cancel func returned
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// beadsGraphStorePath returns path to beads graph JSON.
func beadsGraphStorePath(cfg *config.Config) string {
	base := filepath.Join(os.TempDir(), "ti-beads-graph.json")
	if cfg != nil && cfg.DataDir != "" {
		base = filepath.Join(cfg.DataDir, "beads-graph.json")
	}
	return base
}
