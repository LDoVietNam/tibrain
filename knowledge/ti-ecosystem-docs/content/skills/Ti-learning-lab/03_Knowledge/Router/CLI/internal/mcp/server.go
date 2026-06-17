// Package mcp implements a Model Context Protocol (MCP) server for Ti.
// It exposes Ti's memory, session, agent, and router capabilities as MCP tools
// that can be called by Kiro, Cursor, Claude Desktop, and other MCP clients.
//
// Supports three transports:
//   - stdio (default): reads JSON-RPC from stdin, writes to stdout
//   - HTTP: listens on --http-port, accepts POST /  requests
//   - SSE: listens on --http-port, GET /sse opens stream, POST /message sends requests
package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/google/uuid"
)

const protocolVersion = "2024-11-05"
const serverName = "ti-mcp"
const serverVersion = "0.1.0"

// Request is a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

// RPCError is a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// sseSession holds the response writer and a channel for pushing messages to a connected SSE client.
type sseSession struct {
	ch   chan string
	done chan struct{}
}

// Server holds the tool registry and handles MCP requests.
type Server struct {
	tools    map[string]Tool
	mu       sync.RWMutex
	sessions map[string]*sseSession
}

// New creates a new MCP server with the given tools registered.
func New(tools []Tool) *Server {
	s := &Server{
		tools:    make(map[string]Tool, len(tools)),
		sessions: make(map[string]*sseSession),
	}
	for _, t := range tools {
		s.tools[t.Name] = t
	}
	return s
}

// HandleRequest decodes one JSON-RPC request from r, handles it, and writes
// the JSON-encoded response to w. Used by tests and HTTP handler.
func (s *Server) HandleRequest(r io.Reader, w io.Writer) {
	var req Request
	if err := json.NewDecoder(r).Decode(&req); err != nil {
		resp := errResp(nil, -32700, "parse error: "+err.Error())
		json.NewEncoder(w).Encode(resp)
		return
	}
	resp := s.handle(req)
	json.NewEncoder(w).Encode(resp)
}

// RunStdio runs the MCP server in stdio mode (reads from stdin, writes to stdout).
func (s *Server) RunStdio() error {
	dec := json.NewDecoder(os.Stdin)
	enc := json.NewEncoder(os.Stdout)
	for {
		var req Request
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		resp := s.handle(req)
		if err := enc.Encode(resp); err != nil {
			return err
		}
		os.Stdout.Sync()
	}
}

// RunHTTP runs the MCP server in HTTP+SSE mode on the given port.
//
// Routes:
//
//	GET  /        → health check
//	POST /        → plain JSON-RPC (for non-SSE clients)
//	GET  /sse     → open SSE stream; sends endpoint event with POST URL
//	POST /message → receive JSON-RPC from SSE client, push response over stream
func (s *Server) RunHTTP(port string) error {
	mux := http.NewServeMux()

	// Health / plain JSON-RPC
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":"ok","server":"ti-mcp"}`)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		resp := s.handle(req)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// SSE stream endpoint — client connects here first
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		sessionID := uuid.New().String()
		sess := &sseSession{
			ch:   make(chan string, 16),
			done: make(chan struct{}),
		}
		s.mu.Lock()
		s.sessions[sessionID] = sess
		s.mu.Unlock()

		defer func() {
			s.mu.Lock()
			delete(s.sessions, sessionID)
			s.mu.Unlock()
			close(sess.done)
		}()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Build absolute message URL — ChatGPT web requires a full URL, not a relative path.
		scheme := "https"
		if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") == "" {
			scheme = "http"
		}
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		}
		host := r.Host
		if fwdHost := r.Header.Get("X-Forwarded-Host"); fwdHost != "" {
			host = fwdHost
		}
		messageURL := fmt.Sprintf("%s://%s/message?sessionId=%s", scheme, host, sessionID)

		// Send the endpoint event so the client knows where to POST
		fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", messageURL)
		flusher.Flush()

		// Stream responses until client disconnects
		ctx := r.Context()
		for {
			select {
			case msg := <-sess.ch:
				fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
				flusher.Flush()
			case <-ctx.Done():
				return
			}
		}
	})

	// Message endpoint — client POSTs JSON-RPC here, response goes back over SSE
	mux.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionID := r.URL.Query().Get("sessionId")
		s.mu.RLock()
		sess, ok := s.sessions[sessionID]
		s.mu.RUnlock()
		if !ok {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		resp := s.handle(req)
		b, _ := json.Marshal(resp)

		select {
		case sess.ch <- string(b):
		case <-sess.done:
			http.Error(w, "session closed", http.StatusGone)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"accepted":true}`)
	})

	fmt.Fprintf(os.Stderr, "ti-mcp SSE server listening on :%s\n  SSE:  http://localhost:%s/sse\n  POST: http://localhost:%s/message\n", port, port, port)
	return http.ListenAndServe(":"+port, mux)
}

func (s *Server) handle(req Request) Response {
	switch req.Method {
	case "initialize":
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": protocolVersion,
				"capabilities":    map[string]any{"tools": map[string]any{}},
				"serverInfo":      map[string]any{"name": serverName, "version": serverVersion},
			},
		}
	case "tools/list":
		list := make([]map[string]any, 0, len(s.tools))
		for _, t := range s.tools {
			list = append(list, map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"inputSchema": t.InputSchema,
			})
		}
		return Response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"tools": list}}
	case "tools/call":
		var p struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return errResp(req.ID, -32602, "invalid params")
		}
		t, ok := s.tools[p.Name]
		if !ok {
			return errResp(req.ID, -32601, "tool not found: "+p.Name)
		}
		result, err := t.Handler(p.Arguments)
		if err != nil {
			return errResp(req.ID, -32603, err.Error())
		}
		text := toText(result)
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]any{"content": []map[string]any{{"type": "text", "text": text}}},
		}
	default:
		return errResp(req.ID, -32601, "method not found: "+req.Method)
	}
}

func errResp(id any, code int, msg string) Response {
	return Response{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: msg}}
}

func toText(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
