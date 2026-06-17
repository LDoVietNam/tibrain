package logger

import (
	"context"
	"fmt"

	"github.com/ti/cli/internal/core"
)

// Plugin implements the Plugin interface for structured logging
type Plugin struct {
	name     string
	version  string
	logger   *Logger
	state    core.PluginState
	metadata core.PluginMetadata
}

// NewPlugin creates a new logger plugin
func NewPlugin() (*Plugin, error) {
	logger, err := NewLogger(nil)
	if err != nil {
		return nil, err
	}

	return &Plugin{
		name:    "logger",
		version: "1.0.0",
		logger:  logger,
		state:   core.StateStopped,
		metadata: core.PluginMetadata{
			Name:        "logger",
			Version:     "1.0.0",
			Description: "Native structured logging plugin",
			Author:      "Ti CLI",
			Type:        "native",
			Capabilities: []core.PluginCapability{
				core.CapabilityLogging,
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
	case "log":
		levelStr, _ := input["level"].(string)
		message, _ := input["message"].(string)
		fields, _ := input["fields"].(map[string]interface{})

		var logLevel Level
		switch levelStr {
		case "debug":
			logLevel = LevelDebug
		case "info":
			logLevel = LevelInfo
		case "warn":
			logLevel = LevelWarn
		case "error":
			logLevel = LevelError
		case "fatal":
			logLevel = LevelFatal
		default:
			logLevel = LevelInfo
		}

		p.logger.log(ctx, logLevel, message, fields)

		return map[string]interface{}{
			"success": true,
		}, nil

	case "detect_patterns":
		message, _ := input["message"].(string)

		patterns := p.logger.DetectPatterns(message)

		return map[string]interface{}{
			"success":  true,
			"patterns": patterns,
		}, nil

	default:
		return nil, fmt.Errorf("unknown task: %s", task)
	}
}

func (p *Plugin) Shutdown(ctx context.Context) error {
	p.state = core.StateStopped
	if p.logger != nil {
		p.logger.Close()
	}
	return nil
}

func (p *Plugin) HealthCheck(ctx context.Context) error {
	if p.logger == nil {
		return fmt.Errorf("logger not initialized")
	}
	return nil
}

func (p *Plugin) State() core.PluginState {
	return p.state
}
