package devin

import (
	"context"
	"fmt"
	"sync"

	"github.com/ti/cli/internal/core"
)

// Plugin implements the Plugin interface for Devin agent
type Plugin struct {
	name      string
	version   string
	config    *Config
	lifecycle *LifecycleManager
	adapter   *Adapter
	client    *Client
	mu        sync.RWMutex
	state     core.PluginState
	metadata  core.PluginMetadata
}

// NewPlugin creates a new Devin plugin
func NewPlugin(config *Config) (*Plugin, error) {
	if config == nil {
		var err error
		config, err = LoadConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	lifecycle := NewLifecycleManager(config)
	adapter := NewAdapter(config)

	return &Plugin{
		name:      "devin",
		version:   "1.0.0",
		config:    config,
		lifecycle: lifecycle,
		adapter:   adapter,
		state:     core.StateStopped,
		metadata: core.PluginMetadata{
			Name:        "devin",
			Version:     "1.0.0",
			Type:        "agent",
			Description: "Devin AI agent plugin",
			Author:      "Ti CLI",
			Capabilities: []core.PluginCapability{
				core.CapabilityPromptInjection,
				core.CapabilityOutputExtraction,
			},
		},
	}, nil
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return p.name
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return p.version
}

// Metadata returns the plugin metadata
func (p *Plugin) Metadata() core.PluginMetadata {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.metadata
}

// Initialize initializes the plugin
func (p *Plugin) Initialize(ctx context.Context, config map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != core.StateStopped {
		return fmt.Errorf("plugin already initialized, current state: %s", p.state)
	}

	p.state = core.StateInitializing

	// Start the lifecycle manager
	if err := p.lifecycle.Start(ctx); err != nil {
		p.state = core.StateError
		return fmt.Errorf("failed to start lifecycle: %w", err)
	}

	// Get the client from lifecycle
	p.client = p.lifecycle.GetClient()

	p.state = core.StateReady
	return nil
}

// Stop stops the plugin
func (p *Plugin) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != core.StateRunning {
		return fmt.Errorf("plugin not running, current state: %s", p.state)
	}

	p.state = core.StateStopping

	if err := p.lifecycle.Stop(ctx); err != nil {
		p.state = core.StateError
		return fmt.Errorf("failed to stop lifecycle: %w", err)
	}

	p.state = core.StateStopped
	return nil
}

// State returns the current plugin state
func (p *Plugin) State() core.PluginState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// Execute executes a task
func (p *Plugin) Execute(ctx context.Context, task string, input map[string]interface{}) (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.state != core.StateRunning && p.state != core.StateReady {
		return nil, fmt.Errorf("plugin not ready, current state: %s", p.state)
	}

	if p.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Extract task parameters
	taskID, _ := input["task_id"].(string)
	prompt, _ := input["prompt"].(string)
	options, _ := input["options"].(map[string]interface{})

	// Execute with streaming
	responseChan, err := p.client.Execute(ctx, taskID, prompt, options)
	if err != nil {
		return nil, fmt.Errorf("failed to execute: %w", err)
	}

	// Collect responses
	var result string
	var lastError *ExecuteError

	for response := range responseChan {
		switch response.Type {
		case ResponseTypeOutput:
			result += response.Output
		case ResponseTypeError:
			lastError = response.Error
		case ResponseTypeComplete:
			if !response.Success {
				return nil, fmt.Errorf("task failed: %s", response.Result)
			}
			result = response.Result
		}
	}

	if lastError != nil {
		return nil, fmt.Errorf("execution error: %s", lastError.Message)
	}

	return map[string]interface{}{
		"success": true,
		"result":  result,
	}, nil
}

// Shutdown gracefully shuts down the plugin
func (p *Plugin) Shutdown(ctx context.Context) error {
	return p.Stop(ctx)
}

// HealthCheck performs a health check
func (p *Plugin) HealthCheck(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.lifecycle == nil {
		return fmt.Errorf("lifecycle not initialized")
	}

	healthStatus := p.lifecycle.GetHealthStatus(ctx)
	if !healthStatus.Healthy {
		return fmt.Errorf("health check failed: %s", healthStatus.Message)
	}

	return nil
}
