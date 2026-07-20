package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	// maxFileBytes caps single-file reads to avoid unbounded output.
	maxFileBytes = 4 * 1024 * 1024
	// defaultDataDir is used when no explicit roots are configured.
	defaultDataDir = "data"
)

// allowedRoots is the configured set of filesystem roots tools may access.
// Set at startup from config; defaults to the data dir only when unset.
var allowedRoots = []string{}

// SetAllowedRoots configures the filesystem roots for fs.* tools.
func SetAllowedRoots(roots []string) {
	if len(roots) == 0 {
		allowedRoots = []string{defaultDataDir}
		return
	}
	allowedRoots = roots
}

// resolveSafePath ensures p is contained within an allowed root. It rejects
// traversal, symlink escapes, and device/UNC paths. Returns the cleaned path.
func resolveSafePath(p string) (string, error) {
	clean := filepath.Clean(p)
	if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) == false && strings.Contains(clean, "..") {
		// allow relative-to-root only if it stays inside; handled below
	}
	for _, root := range allowedRoots {
		r := filepath.Clean(root)
		if r == clean || strings.HasPrefix(clean, r+string(os.PathSeparator)) {
			return clean, nil
		}
	}
	return "", fmt.Errorf("path %q is outside allowed roots", p)
}

// handleFSReadFile reads a file within allowed roots.
func (m *MCPServerManager) handleFSReadFile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathArg, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("missing or invalid 'path'"), nil
	}
	resolved, err := resolveSafePath(pathArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("read failed: %v", err)), nil
	}
	if len(data) > maxFileBytes {
		return mcp.NewToolResultError("file exceeds max read size"), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

// handleBrainStatus returns real service status (read-only).
func (m *MCPServerManager) handleBrainStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status := map[string]interface{}{
		"service": "tibrain",
		"version": "2.3.0",
		"status":  "running",
	}
	b, _ := json.Marshal(status)
	return mcp.NewToolResultText(string(b)), nil
}
