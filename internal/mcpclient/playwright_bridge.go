// Playwright MCP Bridge - Browser automation for ChatGPT web + tunnels
package mcpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// PlaywrightBridge wraps playwright-mcp subprocess for browser automation
type PlaywrightBridge struct {
	client    *client.Client
	connected bool
	workDir   string
	headless  bool
}

// BrowserAction represents an action to perform in browser
type BrowserAction struct {
	Action  string                 `json:"action"`
	Target  string                 `json:"target,omitempty"`
	URL     string                 `json:"url,omitempty"`
	Text    string                 `json:"text,omitempty"`
	Timeout int                    `json:"timeout,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// BrowserResult represents browser automation result
type BrowserResult struct {
	Success    bool   `json:"success"`
	Content    string `json:"content,omitempty"`
	Screenshot string `json:"screenshot,omitempty"`
	Error      string `json:"error,omitempty"`
	Actions    int    `json:"actions_performed"`
}

// NewPlaywrightBridge creates bridge to playwright-mcp
func NewPlaywrightBridge(workDir string, headless bool) *PlaywrightBridge {
	return &PlaywrightBridge{
		workDir:  workDir,
		headless: headless,
	}
}

// Initialize spawns and connects to playwright-mcp
func (p *PlaywrightBridge) Initialize(ctx context.Context) error {
	executable := "npx"
	// Using @playwright/mcp package
	args := []string{"@playwright/mcp@latest"}

	env := append(os.Environ(),
		"NODE_ENV=production",
		fmt.Sprintf("PLAYWRIGHT_HEADLESS=%v", p.headless),
	)

	mcpClient, err := newInitializedStdioClient(ctx, "tibrain-playwright-bridge", "1.0.0", executable, env, args...)
	if err != nil {
		return fmt.Errorf("connect playwright: %w", err)
	}

	p.client = mcpClient
	p.connected = true
	return nil
}

// NavigateToURL opens a URL in browser
func (p *PlaywrightBridge) NavigateToURL(ctx context.Context, url string) (*BrowserResult, error) {
	toolsResult, err := p.client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}

	// Find navigate tool
	var navigateTool string
	for _, t := range toolsResult.Tools {
		if t.Name == "browser_navigate" {
			navigateTool = t.Name
			break
		}
	}

	if navigateTool == "" {
		return nil, fmt.Errorf("browser_navigate tool not found")
	}

	req := newToolRequest(navigateTool, map[string]any{
		"url": url,
	})

	result, err := p.client.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("navigate failed: %w", err)
	}

	return p.parseResult(result)
}

// ClickElement clicks on an element
func (p *PlaywrightBridge) ClickElement(ctx context.Context, selector string) (*BrowserResult, error) {
	req := newToolRequest("browser_click", map[string]any{
		"selector": selector,
	})

	result, err := p.client.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("click failed: %w", err)
	}

	return p.parseResult(result)
}

// ExtractContent extracts text content from page
func (p *PlaywrightBridge) ExtractContent(ctx context.Context, selector string) (string, error) {
	req := newToolRequest("browser_extract", map[string]any{
		"selector": selector,
	})

	result, err := p.client.CallTool(ctx, req)
	if err != nil {
		return "", fmt.Errorf("extract failed: %w", err)
	}

	for _, content := range result.Content {
		if tc, ok := content.(mcp.TextContent); ok {
			return tc.Text, nil
		}
	}

	return "", nil
}

// SearchChatGPT searches ChatGPT web interface
func (p *PlaywrightBridge) SearchChatGPT(ctx context.Context, query string) (*BrowserResult, error) {
	// Navigate to chat.openai.com
	if _, err := p.NavigateToURL(ctx, "https://chat.openai.com"); err != nil {
		return nil, fmt.Errorf("navigate to ChatGPT: %w", err)
	}

	// Wait for page load
	time.Sleep(2 * time.Second)

	// Find chat input and type
	req := newToolRequest("browser_type", map[string]any{
		"selector": "textarea",
		"text":     query,
		"submit":   true,
	})

	result, err := p.client.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("type query: %w", err)
	}

	return p.parseResult(result)
}

// SetupTunnel configures browser for tunnel/proxy
func (p *PlaywrightBridge) SetupTunnel(ctx context.Context, proxyURL string) (*BrowserResult, error) {
	// Set proxy configuration for browser
	req := newToolRequest("browser_configure", map[string]any{
		"proxy": map[string]any{
			"server": proxyURL,
		},
	})

	result, err := p.client.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("configure tunnel: %w", err)
	}

	return p.parseResult(result)
}

// Close terminates the subprocess
func (p *PlaywrightBridge) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

func (p *PlaywrightBridge) parseResult(result *mcp.CallToolResult) (*BrowserResult, error) {
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			// Try parse as JSON
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(textContent.Text), &data); err == nil {
				return &BrowserResult{
					Success: true,
					Content: textContent.Text,
				}, nil
			}
			return &BrowserResult{
				Success: true,
				Content: textContent.Text,
			}, nil
		}
	}
	return &BrowserResult{}, nil
}
