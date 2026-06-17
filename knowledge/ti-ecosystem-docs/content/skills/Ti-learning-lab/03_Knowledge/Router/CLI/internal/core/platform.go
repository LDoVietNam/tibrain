package core

import (
	"context"
	"time"
)

// PluginState represents the current state of a plugin
type PluginState int

const (
	PluginStateUnknown PluginState = iota
	PluginStateInitializing
	PluginStateReady
	PluginStateBusy
	PluginStateError
	PluginStateShutdown
	PluginStateStopped
	PluginStateStarting
	PluginStateStopping
	PluginStateRunning
)

func (s PluginState) String() string {
	return [...]string{
		"unknown",
		"initializing",
		"ready",
		"busy",
		"error",
		"shutdown",
		"stopped",
		"starting",
		"stopping",
		"running",
	}[s]
}

// Alias constants for compatibility
const (
	StateUnknown      = PluginStateUnknown
	StateInitializing = PluginStateInitializing
	StateReady        = PluginStateReady
	StateBusy         = PluginStateBusy
	StateError        = PluginStateError
	StateShutdown     = PluginStateShutdown
	StateStopped      = PluginStateStopped
	StateStarting     = PluginStateStarting
	StateStopping     = PluginStateStopping
	StateRunning      = PluginStateRunning
)

// PluginCapability represents a capability that a plugin can provide
type PluginCapability string

const (
	CapabilityStatusDetection  PluginCapability = "status_detection"
	CapabilityTmuxIntegration  PluginCapability = "tmux_integration"
	CapabilityTmuxControl      PluginCapability = "tmux_control"
	CapabilityPermissionCheck  PluginCapability = "permission_check"
	CapabilityPromptInjection  PluginCapability = "prompt_injection"
	CapabilityMCPConfig        PluginCapability = "mcp_config"
	CapabilityOutputExtraction PluginCapability = "output_extraction"
	CapabilityTUIDetection     PluginCapability = "tui_detection"
	CapabilityLogging          PluginCapability = "logging"
)

// PluginMetadata contains metadata about a plugin
type PluginMetadata struct {
	Name         string
	Version      string
	Type         string
	Description  string
	Author       string
	Capabilities []PluginCapability
	Labels       map[string]string
}

// PluginInfo contains runtime information about a plugin
type PluginInfo struct {
	Metadata        PluginMetadata
	State           PluginState
	PID             string
	Uptime          time.Duration
	Statistics      map[string]string
	LastHealthCheck time.Time
	HealthStatus    string
}

// DetailedHealthStatus represents detailed health status
type DetailedHealthStatus struct {
	Healthy bool
	Message string
	Details map[string]string
}

// Plugin represents a plugin that can be loaded and executed
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Metadata returns the plugin metadata
	Metadata() PluginMetadata

	// Initialize initializes the plugin with the given configuration
	Initialize(ctx context.Context, config map[string]string) error

	// Execute executes a task with the given input
	Execute(ctx context.Context, task string, input map[string]interface{}) (map[string]interface{}, error)

	// Shutdown gracefully shuts down the plugin
	Shutdown(ctx context.Context) error

	// HealthCheck performs a health check on the plugin
	HealthCheck(ctx context.Context) error

	// State returns the current state of the plugin
	State() PluginState
}

// Platform represents the microkernel platform that manages plugins
type Platform interface {
	// LoadPlugin loads a plugin by name
	LoadPlugin(ctx context.Context, name string) error

	// UnloadPlugin unloads a plugin by name
	UnloadPlugin(ctx context.Context, name string) error

	// GetPlugin returns a plugin by name
	GetPlugin(name string) (Plugin, error)

	// ListPlugins returns all loaded plugins
	ListPlugins() []PluginInfo

	// ExecutePlugin executes a task on a specific plugin
	ExecutePlugin(ctx context.Context, pluginName, task string, input map[string]interface{}) (map[string]interface{}, error)

	// Shutdown gracefully shuts down the platform
	Shutdown(ctx context.Context) error
}

// DefaultPlatform is the default implementation of the Platform interface
type DefaultPlatform struct {
	plugins  map[string]Plugin
	loader   PluginLoader
	registry *PluginRegistry
}

// NewDefaultPlatform creates a new default platform
func NewDefaultPlatform(loader PluginLoader, registry *PluginRegistry) *DefaultPlatform {
	return &DefaultPlatform{
		plugins:  make(map[string]Plugin),
		loader:   loader,
		registry: registry,
	}
}

// LoadPlugin loads a plugin by name
func (p *DefaultPlatform) LoadPlugin(ctx context.Context, name string) error {
	// Check if plugin is already loaded
	if _, exists := p.plugins[name]; exists {
		return nil
	}

	// Load plugin metadata from registry
	metadata, err := p.registry.GetMetadata(name)
	if err != nil {
		return err
	}

	// Load plugin using loader
	plugin, err := p.loader.Load(ctx, metadata)
	if err != nil {
		return err
	}

	// Initialize plugin
	if err := plugin.Initialize(ctx, nil); err != nil {
		return err
	}

	// Store plugin
	p.plugins[name] = plugin

	return nil
}

// UnloadPlugin unloads a plugin by name
func (p *DefaultPlatform) UnloadPlugin(ctx context.Context, name string) error {
	plugin, exists := p.plugins[name]
	if !exists {
		return nil
	}

	// Shutdown plugin
	if err := plugin.Shutdown(ctx); err != nil {
		return err
	}

	// Remove plugin
	delete(p.plugins, name)

	return nil
}

// GetPlugin returns a plugin by name
func (p *DefaultPlatform) GetPlugin(name string) (Plugin, error) {
	plugin, exists := p.plugins[name]
	if !exists {
		return nil, ErrPluginNotFound
	}
	return plugin, nil
}

// ListPlugins returns all loaded plugins
func (p *DefaultPlatform) ListPlugins() []PluginInfo {
	infos := make([]PluginInfo, 0, len(p.plugins))

	for _, plugin := range p.plugins {
		metadata := plugin.Metadata()
		info := PluginInfo{
			Metadata:   metadata,
			State:      plugin.State(),
			Statistics: make(map[string]string),
		}
		infos = append(infos, info)
	}

	return infos
}

// ExecutePlugin executes a task on a specific plugin
func (p *DefaultPlatform) ExecutePlugin(ctx context.Context, pluginName, task string, input map[string]interface{}) (map[string]interface{}, error) {
	plugin, err := p.GetPlugin(pluginName)
	if err != nil {
		return nil, err
	}

	return plugin.Execute(ctx, task, input)
}

// SetLoader sets the plugin loader
func (p *DefaultPlatform) SetLoader(loader PluginLoader) {
	p.loader = loader
}

// Shutdown gracefully shuts down the platform
func (p *DefaultPlatform) Shutdown(ctx context.Context) error {
	// Shutdown all plugins
	for name := range p.plugins {
		if err := p.UnloadPlugin(ctx, name); err != nil {
			// Log error but continue shutting down other plugins
			continue
		}
	}

	return nil
}

// Errors
var (
	ErrPluginNotFound      = &PluginError{Code: "PLUGIN_NOT_FOUND", Message: "Plugin not found"}
	ErrPluginAlreadyLoaded = &PluginError{Code: "PLUGIN_ALREADY_LOADED", Message: "Plugin is already loaded"}
	ErrPluginNotReady      = &PluginError{Code: "PLUGIN_NOT_READY", Message: "Plugin is not ready"}
)

// PluginError represents an error that occurs during plugin operations
type PluginError struct {
	Code    string
	Message string
	Details string
}

func (e *PluginError) Error() string {
	if e.Details != "" {
		return e.Code + ": " + e.Message + " (" + e.Details + ")"
	}
	return e.Code + ": " + e.Message
}
