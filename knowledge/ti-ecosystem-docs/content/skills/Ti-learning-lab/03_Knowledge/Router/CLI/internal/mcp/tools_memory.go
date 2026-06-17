package mcp

import (
	"fmt"

	"github.com/ti/cli/internal/memory"
)

// MemoryTools returns the 4 memory.* MCP tools.
// palace must be a loaded *memory.Palace.
func MemoryTools(palace *memory.Palace) []Tool {
	return []Tool{
		{
			Name:        "memory.wakeup",
			Description: "Return AI-ready context string from the memory palace (Layer 0+1). Optionally filter by wing.",
			InputSchema: objSchema(map[string]any{
				"wing": strProp("Wing name to filter (optional, empty = all wings)"),
			}, nil),
			Handler: func(args map[string]any) (any, error) {
				wing := strArg(args, "wing")
				return palace.WakeUp(wing), nil
			},
		},
		{
			Name:        "memory.search",
			Description: "Full-text search across memory drawers. Returns matching drawers with relevance scores.",
			InputSchema: objSchema(map[string]any{
				"query": strProp("Search query"),
				"wings": arrStrProp("Wing names to search (optional, empty = all)"),
				"limit": intProp("Max results (default 10)"),
			}, []string{"query"}),
			Handler: func(args map[string]any) (any, error) {
				query := strArg(args, "query")
				if query == "" {
					return nil, fmt.Errorf("query is required")
				}
				wings := strSliceArg(args, "wings")
				limit := intArg(args, "limit", 10)
				stack := palace.Search(query, wings, limit)
				return stack, nil
			},
		},
		{
			Name:        "memory.add",
			Description: "Add a new drawer (memory entry) to the palace.",
			InputSchema: objSchema(map[string]any{
				"wing":       strProp("Wing name (e.g. 'project', 'personal')"),
				"room":       strProp("Room name within the wing (e.g. 'decisions', 'facts')"),
				"text":       strProp("Content to store"),
				"importance": numProp("Importance score 0.0–1.0 (default 0.5)"),
			}, []string{"wing", "room", "text"}),
			Handler: func(args map[string]any) (any, error) {
				wing := strArg(args, "wing")
				room := strArg(args, "room")
				text := strArg(args, "text")
				if wing == "" || room == "" || text == "" {
					return nil, fmt.Errorf("wing, room, and text are required")
				}
				importance := float64Arg(args, "importance", 0.5)
				d := memory.Drawer{
					Text:       text,
					Importance: importance,
				}
				if err := palace.AddDrawer(wing, room, d); err != nil {
					return nil, err
				}
				return map[string]any{"ok": true, "wing": wing, "room": room}, nil
			},
		},
		{
			Name:        "memory.list_wings",
			Description: "List all wing names in the memory palace.",
			InputSchema: objSchema(nil, nil),
			Handler: func(args map[string]any) (any, error) {
				return palace.ListWings(), nil
			},
		},
	}
}
