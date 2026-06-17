package status

import (
	"context"
	"fmt"

	"github.com/ti/cli/internal/core"
)

// Plugin implements the Plugin interface for status detection
type Plugin struct {
	name     string
	version  string
	detector *Detector
	state    core.PluginState
	metadata core.PluginMetadata
}

// NewPlugin creates a new status detection plugin
func NewPlugin() (*Plugin, error) {
	detector := NewDetector(nil)

	return &Plugin{
		name:     "status",
		version:  "1.0.0",
		detector: detector,
		state:    core.StateStopped,
		metadata: core.PluginMetadata{
			Name:        "status",
			Version:     "1.0.0",
			Description: "Native status detection plugin",
			Author:      "Ti CLI",
			Type:        "native",
			Capabilities: []core.PluginCapability{
				core.CapabilityStatusDetection,
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
	output, _ := input["output"].(string)
	sessionID, _ := input["session_id"].(string)

	result, err := p.detector.DetectStatus(ctx, output, sessionID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success":    true,
		"status":     result.Status,
		"completed":  result.IsCompleted,
		"idle":       result.IsIdle,
		"confidence": result.Confidence,
		"metadata":   result.Metadata,
	}, nil
}

func (p *Plugin) Shutdown(ctx context.Context) error {
	p.state = core.StateStopped
	return nil
}

func (p *Plugin) HealthCheck(ctx context.Context) error {
	if p.detector == nil {
		return fmt.Errorf("detector not initialized")
	}
	return nil
}

func (p *Plugin) State() core.PluginState {
	return p.state
}
