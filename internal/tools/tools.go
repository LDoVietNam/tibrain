package tools

import (
	"log"
)

// ToolRegistrar is the minimal surface required to register a tool. The root
// *Hub type implements this interface via its RegisterToolByFields method.
type ToolRegistrar interface {
	RegisterToolByFields(id, name, description, parameters, handler, category, permissions string, enabled bool) error
}

// RegisterDefaultTools registers the built-in TiBrain tools against reg.
// This is a build-enabling scaffold; the tool set is intentionally minimal.
func RegisterDefaultTools(reg ToolRegistrar) error {
	defaults := []struct {
		id          string
		name        string
		description string
		parameters  string
		handler     string
		category    string
		permissions string
	}{
		{
			id:          "ping",
			name:        "Ping",
			description: "Health check that returns pong.",
			parameters:  `{"type":"object","properties":{},"required":[]}`,
			handler:     "local",
			category:    "utility",
			permissions: "public",
		},
		{
			id:          "echo",
			name:        "Echo",
			description: "Echoes the provided message back to the caller.",
			parameters:  `{"type":"object","properties":{"message":{"type":"string"}},"required":["message"]}`,
			handler:     "local",
			category:    "utility",
			permissions: "public",
		},
	}

	for _, t := range defaults {
		if err := reg.RegisterToolByFields(t.id, t.name, t.description, t.parameters, t.handler, t.category, t.permissions, true); err != nil {
			return err
		}
		log.Printf("tools: registered default tool %q", t.id)
	}

	return nil
}
