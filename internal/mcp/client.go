package mcp

// MCPClientConfig holds the configuration needed to create an MCP client.
type MCPClientConfig struct {
	Type     string
	Command  string
	Args     string
	Env      string
}