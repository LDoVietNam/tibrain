// MCP Client wrapper for integrating documentation server
package mcpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// DocumentationClient wraps MCP client for docs-mcp-server
type DocumentationClient struct {
	client     *client.Client
	connected  bool
	mu         sync.Mutex
	executable string
	args       []string
}

// NewDocumentationClient creates a new docs MCP client
func NewDocumentationClient(executable string, args ...string) *DocumentationClient {
	return &DocumentationClient{
		executable: executable,
		args:       args,
	}
}

// Connect initializes the MCP client connection
func (c *DocumentationClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	mcpClient, err := newInitializedStdioClient(ctx, "tibrain-docs-client", "1.0.0", c.executable, os.Environ(), c.args...)
	if err != nil {
		return fmt.Errorf("connect to docs server: %w", err)
	}

	c.client = mcpClient
	c.connected = true
	return nil
}

// Disconnect closes the MCP client connection
func (c *DocumentationClient) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected || c.client == nil {
		return nil
	}

	if err := c.client.Close(); err != nil {
		return fmt.Errorf("close docs client: %w", err)
	}

	c.connected = false
	return nil
}

// SearchDocumentation calls the docs-mcp-server search tool
func (c *DocumentationClient) SearchDocumentation(ctx context.Context, query string, maxResults int) ([]DocumentationResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		if err := c.Connect(ctx); err != nil {
			return nil, err
		}
	}

	req := newToolRequest("search_docs", map[string]any{
		"query":       query,
		"max_results": maxResults,
	})

	result, err := c.client.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call search_docs tool: %w", err)
	}

	var results []DocumentationResult
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			var docs []DocumentationResult
			if err := json.Unmarshal([]byte(textContent.Text), &docs); err == nil {
				results = append(results, docs...)
			}
		}
	}

	return results, nil
}

// DocumentationResult represents a documentation search result
type DocumentationResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Source  string `json:"source"`
}

// FetchDocumentationContent fetches full content from a documentation URL
func (c *DocumentationClient) FetchDocumentationContent(ctx context.Context, url string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		if err := c.Connect(ctx); err != nil {
			return "", err
		}
	}

	req := newToolRequest("fetch_docs", map[string]any{
		"url": url,
	})

	result, err := c.client.CallTool(ctx, req)
	if err != nil {
		return "", fmt.Errorf("call fetch_docs tool: %w", err)
	}

	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			return textContent.Text, nil
		}
	}

	return "", fmt.Errorf("no content returned from fetch_docs")
}

// ListTools returns available tools from the docs server
func (c *DocumentationClient) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		if err := c.Connect(ctx); err != nil {
			return nil, err
		}
	}

	result, err := c.client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}
	return result.Tools, nil
}

// Ensure tool exists in list
func (c *DocumentationClient) HasTool(toolName string) bool {
	tools, err := c.ListTools(context.Background())
	if err != nil {
		return false
	}

	for _, tool := range tools {
		if tool.Name == toolName {
			return true
		}
	}
	return false
}

// ReadAllContent reads all documentation from a source
func (c *DocumentationClient) ReadAllContent(ctx context.Context, source string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		if err := c.Connect(ctx); err != nil {
			return "", err
		}
	}

	req := newToolRequest("read_all_docs", map[string]any{
		"source": source,
	})

	result, err := c.client.CallTool(ctx, req)
	if err != nil {
		return "", fmt.Errorf("call read_all_docs: %w", err)
	}

	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			return textContent.Text, nil
		}
	}

	return "", nil
}
