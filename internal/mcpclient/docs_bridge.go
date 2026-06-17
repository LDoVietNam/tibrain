// Docs MCP Bridge - Integrates docs-mcp-server into TiBrain
package mcpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// DocsBridge wraps docs-mcp-server subprocess
type DocsBridge struct {
	client    *client.Client
	connected bool
	workDir   string
}

// NewDocsBridge creates a bridge to docs-mcp-server
func NewDocsBridge(workDir string) *DocsBridge {
	return &DocsBridge{
		workDir: workDir,
	}
}

// Initialize spawns and connects to docs-mcp-server
func (d *DocsBridge) Initialize(ctx context.Context) error {
	// Path to docs-mcp-server binary (built from mcps/docs)
	executable := fmt.Sprintf("%s/mcps/docs/dist/index.js", d.workDir)

	// Check if built
	if _, err := os.Stat(executable); os.IsNotExist(err) {
		// Try building
		return fmt.Errorf("docs-mcp-server not built. Run: npm run build in mcps/docs")
	}

	mcpClient, err := newInitializedStdioClient(ctx, "tibrain-docs-bridge", "1.0.0", "node", append(os.Environ(), "NODE_ENV=production"), executable)
	if err != nil {
		return fmt.Errorf("connect to docs: %w", err)
	}

	d.client = mcpClient
	d.connected = true
	return nil
}

// Search calls docs-mcp-server search tool
func (d *DocsBridge) Search(ctx context.Context, library, query string, limit int) ([]map[string]interface{}, error) {
	if !d.connected {
		if err := d.Initialize(ctx); err != nil {
			return nil, err
		}
	}

	req := newToolRequest("search", map[string]any{
		"library": library,
		"query":   query,
		"limit":   limit,
	})

	result, err := d.client.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var results []map[string]interface{}
	for _, content := range result.Content {
		if tc, ok := content.(mcp.TextContent); ok {
			var parsed []map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Text), &parsed); err == nil {
				results = append(results, parsed...)
			} else {
				// If not JSON array, wrap as single result
				results = append(results, map[string]interface{}{
					"content": tc.Text,
				})
			}
		}
	}

	return results, nil
}

// ListLibraries returns available documentation libraries
func (d *DocsBridge) ListLibraries(ctx context.Context) ([]map[string]interface{}, error) {
	if !d.connected {
		if err := d.Initialize(ctx); err != nil {
			return nil, err
		}
	}

	req := newToolRequest("list_libraries", nil)

	result, err := d.client.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("list libraries failed: %w", err)
	}

	var libs []map[string]interface{}
	for _, content := range result.Content {
		if tc, ok := content.(mcp.TextContent); ok {
			json.Unmarshal([]byte(tc.Text), &libs)
		}
	}

	return libs, nil
}

// Close terminates the subprocess
func (d *DocsBridge) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

// HealthCheck verifies the bridge is working
func (d *DocsBridge) HealthCheck() error {
	if !d.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

// GetVersion returns the docs server version
func (d *DocsBridge) GetVersion() string {
	return "docs-mcp-server/2.4.0"
}

// BuildScript returns the npm build command
func (d *DocsBridge) BuildScript() string {
	return "npm run build"
}

// InitScript returns the init command
func (d *DocsBridge) InitScript() string {
	return "npm install"
}
