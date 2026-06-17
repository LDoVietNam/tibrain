package permission

import (
	"context"
	"fmt"

	"github.com/ti/cli/internal/core"
)

// Plugin implements the Plugin interface for permission management
type Plugin struct {
	name     string
	version  string
	manager  *Manager
	state    core.PluginState
	metadata core.PluginMetadata
}

// NewPlugin creates a new permission manager plugin
func NewPlugin() (*Plugin, error) {
	manager := NewManager(nil)

	return &Plugin{
		name:    "permission",
		version: "1.0.0",
		manager: manager,
		state:   core.StateStopped,
		metadata: core.PluginMetadata{
			Name:        "permission",
			Version:     "1.0.0",
			Description: "Native permission management plugin",
			Author:      "Ti CLI",
			Type:        "native",
			Capabilities: []core.PluginCapability{
				core.CapabilityPermissionCheck,
			},
		},
	}, nil
}

func (p *Plugin) Name() string {
	return p.name
}

func (p *Plugin) Version() string {
	return p.version
}

func (p *Plugin) Metadata() core.PluginMetadata {
	return p.metadata
}

func (p *Plugin) Initialize(ctx context.Context, config map[string]string) error {
	p.state = core.StateReady
	return nil
}

func (p *Plugin) Execute(ctx context.Context, task string, input map[string]interface{}) (map[string]interface{}, error) {
	switch task {
	case "check_permission":
		actionStr, _ := input["action"].(string)
		resource, _ := input["resource"].(string)

		var action Permission
		switch actionStr {
		case "read":
			action = PermissionRead
		case "write":
			action = PermissionWrite
		case "execute":
			action = PermissionExecute
		case "network":
			action = PermissionNetwork
		case "system":
			action = PermissionSystem
		default:
			action = PermissionRead
		}

		result, err := p.manager.Check(ctx, action, resource)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"success": true,
			"allowed": result.Allowed,
			"reason":  result.Reason,
			"mode":    string(result.Mode),
		}, nil

	case "add_rule":
		actionStr, _ := input["action"].(string)
		resource, _ := input["resource"].(string)
		allow, _ := input["allow"].(bool)
		description, _ := input["description"].(string)

		var action Permission
		switch actionStr {
		case "read":
			action = PermissionRead
		case "write":
			action = PermissionWrite
		case "execute":
			action = PermissionExecute
		case "network":
			action = PermissionNetwork
		case "system":
			action = PermissionSystem
		default:
			action = PermissionRead
		}

		rule := PermissionRule{
			Action:      action,
			Resource:    resource,
			Pattern:     "*",
			Allowed:     allow,
			Description: description,
		}

		p.manager.AddRule(rule)

		return map[string]interface{}{
			"success": true,
		}, nil

	case "remove_rule":
		description, _ := input["description"].(string)

		removed := p.manager.RemoveRule(description)

		return map[string]interface{}{
			"success": true,
			"removed": removed,
		}, nil

	case "set_mode":
		modeStr, _ := input["mode"].(string)

		var mode PermissionMode
		switch modeStr {
		case "strict":
			mode = ModeStrict
		case "permissive":
			mode = ModePermissive
		case "balanced":
			mode = ModeBalanced
		default:
			mode = ModeBalanced
		}

		p.manager.SetMode(mode)

		return map[string]interface{}{
			"success": true,
		}, nil

	case "get_rules":
		rules := p.manager.GetRules()

		rulesMap := make([]map[string]interface{}, len(rules))
		for i, rule := range rules {
			rulesMap[i] = map[string]interface{}{
				"action":      string(rule.Action),
				"resource":    rule.Resource,
				"pattern":     rule.Pattern,
				"allowed":     rule.Allowed,
				"description": rule.Description,
			}
		}

		return map[string]interface{}{
			"success": true,
			"rules":   rulesMap,
		}, nil

	default:
		return nil, fmt.Errorf("unknown task: %s", task)
	}
}

func (p *Plugin) Shutdown(ctx context.Context) error {
	p.state = core.StateStopped
	return nil
}

func (p *Plugin) HealthCheck(ctx context.Context) error {
	if p.manager == nil {
		return fmt.Errorf("manager not initialized")
	}
	return nil
}

func (p *Plugin) State() core.PluginState {
	return p.state
}
