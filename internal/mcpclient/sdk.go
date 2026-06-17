package mcpclient

import (
	"context"
	"fmt"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func newInitializedStdioClient(ctx context.Context, name, version, command string, env []string, args ...string) (*mcpclient.Client, error) {
	c, err := mcpclient.NewStdioMCPClient(command, env, args...)
	if err != nil {
		return nil, err
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    name,
		Version: version,
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	if _, err := c.Initialize(ctx, initRequest); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("initialize MCP client: %w", err)
	}

	return c, nil
}

func newToolRequest(name string, arguments map[string]any) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	if arguments != nil {
		req.Params.Arguments = arguments
	}
	return req
}
