package mcp

import (
	"fmt"

	"github.com/ti/cli/internal/session"
)

// SessionTools returns the 4 session.* MCP tools.
func SessionTools(store *session.Store) []Tool {
	return []Tool{
		{
			Name:        "session.list",
			Description: "List recent sessions. Returns id, title, model, provider, created_at, token_count.",
			InputSchema: objSchema(map[string]any{
				"limit": intProp("Max sessions to return (default 20)"),
			}, nil),
			Handler: func(args map[string]any) (any, error) {
				limit := intArg(args, "limit", 20)
				sessions, err := store.List(limit)
				if err != nil {
					return nil, err
				}
				out := make([]map[string]any, 0, len(sessions))
				for _, s := range sessions {
					out = append(out, map[string]any{
						"id":          s.ID,
						"title":       s.Title,
						"model":       s.Model,
						"provider":    s.Provider,
						"created_at":  s.CreatedAt,
						"token_count": s.TokenCount,
						"phase":       s.Phase,
					})
				}
				return out, nil
			},
		},
		{
			Name:        "session.get",
			Description: "Get a session by ID including all messages.",
			InputSchema: objSchema(map[string]any{
				"id": strProp("Session ID"),
			}, []string{"id"}),
			Handler: func(args map[string]any) (any, error) {
				id := strArg(args, "id")
				if id == "" {
					return nil, fmt.Errorf("id is required")
				}
				s, err := store.Get(id)
				if err != nil {
					return nil, err
				}
				return s, nil
			},
		},
		{
			Name:        "session.search",
			Description: "Search sessions by text query (searches title, summary, messages).",
			InputSchema: objSchema(map[string]any{
				"query": strProp("Search query"),
				"limit": intProp("Max results (default 10)"),
			}, []string{"query"}),
			Handler: func(args map[string]any) (any, error) {
				query := strArg(args, "query")
				if query == "" {
					return nil, fmt.Errorf("query is required")
				}
				limit := intArg(args, "limit", 10)
				sessions, err := store.Search(query, limit)
				if err != nil {
					return nil, err
				}
				out := make([]map[string]any, 0, len(sessions))
				for _, s := range sessions {
					out = append(out, map[string]any{
						"id":       s.ID,
						"title":    s.Title,
						"model":    s.Model,
						"provider": s.Provider,
						"summary":  s.Summary,
					})
				}
				return out, nil
			},
		},
		{
			Name:        "session.append",
			Description: "Append a message to an existing session.",
			InputSchema: objSchema(map[string]any{
				"id":      strProp("Session ID"),
				"role":    strProp("Message role: user | assistant | system"),
				"content": strProp("Message content"),
			}, []string{"id", "role", "content"}),
			Handler: func(args map[string]any) (any, error) {
				id := strArg(args, "id")
				role := strArg(args, "role")
				content := strArg(args, "content")
				if id == "" || role == "" || content == "" {
					return nil, fmt.Errorf("id, role, and content are required")
				}
				msg := session.Message{Role: role, Content: content}
				if err := store.AppendMessage(id, msg, 0); err != nil {
					return nil, err
				}
				return map[string]any{"ok": true}, nil
			},
		},
	}
}
