// Package mcp implements the TiBrain MCP gateway using the mark3labs/mcp-go SDK.
// Transport: Streamable HTTP (preferred) at /mcp, legacy SSE bridge at /mcp/sse.
package mcp

import (
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewMCPServerManager builds the MCP server with the canonical tool set.
// Tools are registered with real implementations; no placeholder/mock success.
func NewMCPServerManager() *MCPServerManager {
	s := server.NewMCPServer(
		"tibrain",
		"2.3.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	m := &MCPServerManager{
		srv:    s,
		stream: server.NewStreamableHTTPServer(s),
		sse:    server.NewSSEServer(s),
	}

	m.registerTools()
	return m
}

// MCPServerManager owns the MCP server and its transports.
type MCPServerManager struct {
	srv    *server.MCPServer
	stream *server.StreamableHTTPServer
	sse    *server.SSEServer
}

// HandleStreamableHTTP serves the canonical /mcp Streamable HTTP endpoint
// (POST/GET/DELETE with Mcp-Session-Id negotiation).
func (m *MCPServerManager) HandleStreamableHTTP(w http.ResponseWriter, r *http.Request) {
	m.stream.ServeHTTP(w, r)
}

// HandleSSE serves the legacy /mcp/sse endpoint (compatibility bridge only).
func (m *MCPServerManager) HandleSSE(w http.ResponseWriter, r *http.Request) {
	m.sse.ServeHTTP(w, r)
}

// HandleMessage serves legacy /mcp/message POST for SSE clients.
func (m *MCPServerManager) HandleMessage(w http.ResponseWriter, r *http.Request) {
	m.sse.ServeHTTP(w, r)
}

// registerTools wires the real, implemented tools. Only tools with a real
// backend implementation are registered; unimplemented ones are NOT added.
func (m *MCPServerManager) registerTools() {
	// Filesystem read-only (real: operates on configured roots only)
	m.srv.AddTool(mcp.NewTool("fs.read_file",
		mcp.WithDescription("Read a file within an allowed root. Read-only, side-effect free."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Absolute path under an allowed root")),
	), m.handleFSReadFile)

	// Health/readiness (real)
	m.srv.AddTool(mcp.NewTool("brain.status",
		mcp.WithDescription("Return TiBrain service status, version and health. Read-only."),
	), m.handleBrainStatus)
}
