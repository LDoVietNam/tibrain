package tmux

import (
	"context"
	"fmt"
	"time"

	"github.com/ti/cli/internal/core"
)

// Plugin implements the Plugin interface for tmux integration
type Plugin struct {
	name     string
	version  string
	client   *Client
	state    core.PluginState
	metadata core.PluginMetadata
}

// NewPlugin creates a new tmux integration plugin
func NewPlugin() (*Plugin, error) {
	client, err := NewClient(nil)
	if err != nil {
		return nil, err
	}

	return &Plugin{
		name:    "tmux",
		version: "1.0.0",
		client:  client,
		state:   core.StateStopped,
		metadata: core.PluginMetadata{
			Name:        "tmux",
			Version:     "1.0.0",
			Type:        "native",
			Description: "Native tmux integration plugin",
			Author:      "Ti CLI",
			Capabilities: []core.PluginCapability{
				core.CapabilityTmuxControl,
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
	case "send_command":
		window, _ := input["window"].(string)
		command, _ := input["command"].(string)

		err := p.client.SendCommand(ctx, window, command)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"success": true,
		}, nil

	case "capture_output":
		window, _ := input["window"].(string)

		output, err := p.client.CaptureOutput(ctx, window)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"success": true,
			"output":  output,
		}, nil

	case "wait_for_shell":
		window, _ := input["window"].(string)
		timeoutSec, _ := input["timeout"].(int)
		if timeoutSec == 0 {
			timeoutSec = 30
		}

		err := p.client.WaitForShell(ctx, window, time.Duration(timeoutSec)*time.Second)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"success": true,
		}, nil

	case "wait_until_status":
		window, _ := input["window"].(string)
		status, _ := input["status"].(string)
		timeoutSec, _ := input["timeout"].(int)
		if timeoutSec == 0 {
			timeoutSec = 30
		}

		err := p.client.WaitUntilStatus(ctx, window, status, time.Duration(timeoutSec)*time.Second)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"success": true,
		}, nil

	case "new_window":
		name, _ := input["name"].(string)

		err := p.client.NewWindow(ctx, name)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"success": true,
		}, nil

	case "split_window":
		window, _ := input["window"].(string)
		vertical, _ := input["vertical"].(bool)

		err := p.client.SplitWindow(ctx, window, vertical)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"success": true,
		}, nil

	case "list_sessions":
		sessions, err := p.client.ListSessions(ctx)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"success":  true,
			"sessions": sessions,
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
	if p.client == nil {
		return fmt.Errorf("client not initialized")
	}
	return nil
}

func (p *Plugin) State() core.PluginState {
	return p.state
}
