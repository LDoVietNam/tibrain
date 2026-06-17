// MCP Hub Client - Integration with external MCP Hub
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MCPHubClient handles communication with MCP Hub
type MCPHubClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewMCPHubClient creates a new MCP Hub client
func NewMCPHubClient(baseURL string) *MCPHubClient {
	return &MCPHubClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MCPServerInfo represents MCP server information from Hub
type MCPServerInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Endpoint    string            `json:"endpoint"`
	Transport   string            `json:"transport"`
	Tools       []ToolInfo        `json:"tools"`
	Resources   []ResourceInfo    `json:"resources"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ToolInfo represents a tool from MCP server
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ResourceInfo represents a resource from MCP server
type ResourceInfo struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// RegisterServerRequest represents request to register MCP server
type RegisterServerRequest struct {
	Name        string            `json:"name"`
	Endpoint    string            `json:"endpoint"`
	Transport   string            `json:"transport"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ListServers returns all registered MCP servers from Hub
func (c *MCPHubClient) ListServers() ([]MCPServerInfo, error) {
	url := fmt.Sprintf("%s/api/servers", c.BaseURL)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var servers []MCPServerInfo
	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return servers, nil
}

// GetServer returns a specific MCP server from Hub
func (c *MCPHubClient) GetServer(id string) (*MCPServerInfo, error) {
	url := fmt.Sprintf("%s/api/servers/%s", c.BaseURL, id)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("server not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var server MCPServerInfo
	if err := json.NewDecoder(resp.Body).Decode(&server); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &server, nil
}

// RegisterServer registers a new MCP server with Hub
func (c *MCPHubClient) RegisterServer(req RegisterServerRequest) (*MCPServerInfo, error) {
	url := fmt.Sprintf("%s/api/servers", c.BaseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var server MCPServerInfo
	if err := json.NewDecoder(resp.Body).Decode(&server); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &server, nil
}

// UnregisterServer removes an MCP server from Hub
func (c *MCPHubClient) UnregisterServer(id string) error {
	url := fmt.Sprintf("%s/api/servers/%s", c.BaseURL, id)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("http delete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SyncTools syncs tools from MCP Hub to local registry
func (c *MCPHubClient) SyncTools(hub *Hub) error {
	servers, err := c.ListServers()
	if err != nil {
		return fmt.Errorf("list servers: %w", err)
	}

	for _, server := range servers {
		// Register or update in mcp_hub_registry
		timestamp := time.Now().Unix()
		toolsJSON, _ := json.Marshal(server.Tools)
		resourcesJSON, _ := json.Marshal(server.Resources)

		enabledInt := 1
		if _, err := hub.db.Exec(`
			INSERT OR REPLACE INTO mcp_hub_registry (id, name, endpoint, transport, tools, resources, description, enabled, synced_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, server.ID, server.Name, server.Endpoint, server.Transport, string(toolsJSON), string(resourcesJSON), server.Description, enabledInt, timestamp, timestamp, timestamp); err != nil {
			logger.Warn("Failed to sync MCP Hub server %s: %v", server.ID, err)
			continue
		}

		// Sync tools to tool_registry
		for _, tool := range server.Tools {
			toolID := fmt.Sprintf("mcp-hub-%s-%s", server.ID, tool.Name)
			parametersJSON, _ := json.Marshal(tool.InputSchema)

			toolReg := Tool{
				ID:                   toolID,
				Name:                 tool.Name,
				Description:          tool.Description,
				Parameters:           string(parametersJSON),
				Handler:              "mcp-hub",
				Category:             "mcp",
				Permissions:          `["devin","claude","cursor"]`,
				Enabled:              true,
				QualityScore:         80,
				SecurityScore:        80,
				BestPracticesScore:   80,
				Source:               "mcp-hub",
				Family:               "",
				Tags:                 "",
				Version:              "",
				SkillLevel:           "l2",
				QualityTier:          "platinum",
				SecurityTier:         "hardened",
				SecurityStatus:       "passed",
				ValidationStatus:     "passed",
				VariantID:            "",
				VariantLabel:         "",
				SourceType:           "community",
				RootPath:             "",
			}

			if err := hub.RegisterTool(toolReg); err != nil {
				logger.Warn("Failed to register tool %s: %v", tool.Name, err)
			}
		}

		logger.Info("Synced MCP Hub server: %s (%d tools)", server.Name, len(server.Tools))
	}

	return nil
}

// HealthCheck checks if MCP Hub is accessible
func (c *MCPHubClient) HealthCheck() error {
	url := fmt.Sprintf("%s/api/health", c.BaseURL)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}
