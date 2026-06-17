package plugins

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/ti/cli/internal/agents/devin"
	"github.com/ti/cli/internal/agents/native/logger"
	"github.com/ti/cli/internal/agents/native/permission"
	"github.com/ti/cli/internal/agents/native/status"
	"github.com/ti/cli/internal/agents/native/tmux"
	"github.com/ti/cli/internal/core"
)

// DefaultLoader is the default implementation of PluginLoader
type DefaultLoader struct {
	mu            sync.RWMutex
	loadedPlugins map[string]*LoadedPlugin
}

// LoadedPlugin represents a loaded plugin
type LoadedPlugin struct {
	Name      string
	Metadata  core.PluginMetadata
	Command   *exec.Cmd
	Process   *PluginProcess
	State     core.PluginState
	StartTime time.Time
}

// PluginProcess represents a running plugin process
type PluginProcess struct {
	PID      int
	Command  string
	Args     []string
	ExitChan chan error
}

// NewDefaultLoader creates a new default plugin loader
func NewDefaultLoader() *DefaultLoader {
	return &DefaultLoader{
		loadedPlugins: make(map[string]*LoadedPlugin),
	}
}

// Load loads a plugin based on its metadata
func (dl *DefaultLoader) Load(ctx context.Context, metadata core.PluginMetadata) (core.Plugin, error) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	// Check if plugin is already loaded
	if _, exists := dl.loadedPlugins[metadata.Name]; exists {
		return nil, fmt.Errorf("plugin %s is already loaded", metadata.Name)
	}

	// Create plugin based on type
	plugin, err := dl.createPlugin(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin: %w", err)
	}

	// Store loaded plugin info
	loadedPlugin := &LoadedPlugin{
		Name:      metadata.Name,
		Metadata:  metadata,
		State:     core.PluginStateInitializing,
		StartTime: time.Now(),
	}
	dl.loadedPlugins[metadata.Name] = loadedPlugin

	return plugin, nil
}

// Unload unloads a plugin
func (dl *DefaultLoader) Unload(ctx context.Context, name string) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	loadedPlugin, exists := dl.loadedPlugins[name]
	if !exists {
		return fmt.Errorf("plugin %s is not loaded", name)
	}

	// Stop plugin process if running
	if loadedPlugin.Process != nil {
		if err := dl.stopProcess(ctx, loadedPlugin.Process); err != nil {
			return fmt.Errorf("failed to stop plugin process: %w", err)
		}
	}

	// Remove from loaded plugins
	delete(dl.loadedPlugins, name)

	return nil
}

// GetLoadedPlugin returns a loaded plugin by name
func (dl *DefaultLoader) GetLoadedPlugin(name string) (*LoadedPlugin, error) {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	plugin, exists := dl.loadedPlugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s is not loaded", name)
	}

	return plugin, nil
}

// ListLoadedPlugins returns all loaded plugins
func (dl *DefaultLoader) ListLoadedPlugins() []*LoadedPlugin {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	plugins := make([]*LoadedPlugin, 0, len(dl.loadedPlugins))
	for _, plugin := range dl.loadedPlugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// createPlugin creates a plugin instance based on metadata
func (dl *DefaultLoader) createPlugin(metadata core.PluginMetadata) (core.Plugin, error) {
	// Check plugin type based on name or type
	switch metadata.Name {
	case "devin":
		return devin.NewPlugin(nil)
	case "status":
		return status.NewPlugin()
	case "tmux":
		return tmux.NewPlugin()
	case "logger":
		return logger.NewPlugin()
	case "permission":
		return permission.NewPlugin()
	default:
		// For unknown plugins, return a base plugin
		return &BasePlugin{
			metadata: metadata,
			state:    core.PluginStateReady,
		}, nil
	}
}

// stopProcess stops a plugin process
func (dl *DefaultLoader) stopProcess(ctx context.Context, process *PluginProcess) error {
	// Implement process stopping logic
	// For now, return nil as placeholder
	return nil
}

// BasePlugin is a base implementation of the Plugin interface
type BasePlugin struct {
	metadata core.PluginMetadata
	state    core.PluginState
	config   map[string]string
	mu       sync.RWMutex
}

// Name returns the plugin name
func (bp *BasePlugin) Name() string {
	return bp.metadata.Name
}

// Version returns the plugin version
func (bp *BasePlugin) Version() string {
	return bp.metadata.Version
}

// Metadata returns the plugin metadata
func (bp *BasePlugin) Metadata() core.PluginMetadata {
	return bp.metadata
}

// Initialize initializes the plugin
func (bp *BasePlugin) Initialize(ctx context.Context, config map[string]string) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.config = config
	bp.state = core.PluginStateReady

	return nil
}

// Execute executes a task
func (bp *BasePlugin) Execute(ctx context.Context, task string, input map[string]interface{}) (map[string]interface{}, error) {
	bp.mu.Lock()
	bp.state = core.PluginStateBusy
	bp.mu.Unlock()

	defer func() {
		bp.mu.Lock()
		bp.state = core.PluginStateReady
		bp.mu.Unlock()
	}()

	// Placeholder implementation
	// In the future, this will route to actual plugin implementations
	result := make(map[string]interface{})
	result["task"] = task
	result["status"] = "completed"

	return result, nil
}

// Shutdown shuts down the plugin
func (bp *BasePlugin) Shutdown(ctx context.Context) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.state = core.PluginStateShutdown

	return nil
}

// HealthCheck performs a health check
func (bp *BasePlugin) HealthCheck(ctx context.Context) error {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	if bp.state == core.PluginStateError {
		return fmt.Errorf("plugin is in error state")
	}

	return nil
}

// State returns the current state
func (bp *BasePlugin) State() core.PluginState {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	return bp.state
}

// NativePlugin represents a native Go plugin
type NativePlugin struct {
	*BasePlugin
	executeFunc func(ctx context.Context, task string, input map[string]interface{}) (map[string]interface{}, error)
}

// Execute executes a task using the native plugin's execute function
func (np *NativePlugin) Execute(ctx context.Context, task string, input map[string]interface{}) (map[string]interface{}, error) {
	if np.executeFunc != nil {
		return np.executeFunc(ctx, task, input)
	}

	return np.BasePlugin.Execute(ctx, task, input)
}

// ExternalPlugin represents an external plugin (e.g., Python, gRPC server)
type ExternalPlugin struct {
	*BasePlugin
	command string
	args    []string
}

// Execute executes a task by calling the external plugin
func (ep *ExternalPlugin) Execute(ctx context.Context, task string, input map[string]interface{}) (map[string]interface{}, error) {
	// Placeholder implementation for external plugin execution
	// In the future, this will communicate with external processes via gRPC or stdio

	result := make(map[string]interface{})
	result["task"] = task
	result["status"] = "completed"
	result["plugin_type"] = "external"

	return result, nil
}

// HealthCheck performs a health check on the external plugin
func (ep *ExternalPlugin) HealthCheck(ctx context.Context) error {
	// Check if the external process is still running
	// For now, return nil as placeholder

	return nil
}
