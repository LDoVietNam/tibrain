package mcp

// Tool defines a single MCP tool with its metadata and handler.
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]any
	Handler     func(args map[string]any) (any, error)
}

// strArg extracts a string argument, returning "" if missing or wrong type.
func strArg(args map[string]any, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// intArg extracts an int argument (also handles float64 from JSON), returning def if missing.
func intArg(args map[string]any, key string, def int) int {
	v, ok := args[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return def
}

// float64Arg extracts a float64 argument, returning def if missing.
func float64Arg(args map[string]any, key string, def float64) float64 {
	v, ok := args[key]
	if !ok {
		return def
	}
	f, ok := v.(float64)
	if !ok {
		return def
	}
	return f
}

// strSliceArg extracts a []string argument from a JSON array.
func strSliceArg(args map[string]any, key string) []string {
	v, ok := args[key]
	if !ok {
		return nil
	}
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// schema helpers

func strProp(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}

func intProp(desc string) map[string]any {
	return map[string]any{"type": "integer", "description": desc}
}

func numProp(desc string) map[string]any {
	return map[string]any{"type": "number", "description": desc}
}

func arrStrProp(desc string) map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": desc}
}

func objSchema(props map[string]any, required []string) map[string]any {
	s := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}
