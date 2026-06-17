package plugins

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ti/cli/internal/core"
)

// Manager manages plugins and their lifecycle
type Manager struct {
	mu               sync.RWMutex
	platform         core.Platform
	registry         *core.PluginRegistry
	lifecycleManager *core.LifecycleManager
	discovery        PluginDiscovery
	config           *Config
}

// Config contains configuration for the plugin manager
type Config struct {
	PluginDir           string
	AutoLoad            bool
	HealthCheckInterval int
	HealthCheckTimeout  int
	MaxRestarts         int
	RestartDelay        int
}

// NewManager creates a new plugin manager
func NewManager(config *Config) *Manager {
	// Create registry
	registry := core.NewPluginRegistry()

	// Create lifecycle manager with default policy
	policy := core.DefaultRestartPolicy()
	policy.MaxRestarts = config.MaxRestarts
	if config.RestartDelay > 0 {
		policy.RestartDelay = durationFromSeconds(config.RestartDelay)
	}

	lifecycleManager := core.NewLifecycleManager(
		durationFromSeconds(config.HealthCheckInterval),
		durationFromSeconds(config.HealthCheckTimeout),
		policy,
	)

	// Create platform
	platform := core.NewDefaultPlatform(nil, registry)

	return &Manager{
		platform:         platform,
		registry:         registry,
		lifecycleManager: lifecycleManager,
		config:           config,
	}
}

// SetDiscovery sets the plugin discovery mechanism
func (m *Manager) SetDiscovery(discovery PluginDiscovery) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.discovery = discovery
}

// SetLoader sets the plugin loader
func (m *Manager) SetLoader(loader core.PluginLoader) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if platform, ok := m.platform.(*core.DefaultPlatform); ok {
		platform.SetLoader(loader)
	}
}

// LoadPlugin loads a plugin by name
func (m *Manager) LoadPlugin(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check dependencies
	deps, err := m.registry.ResolveDependencies(name)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Load dependencies first
	for _, dep := range deps {
		if err := m.platform.LoadPlugin(ctx, dep); err != nil {
			return fmt.Errorf("failed to load dependency %s: %w", dep, err)
		}
	}

	// Load plugin
	if err := m.platform.LoadPlugin(ctx, name); err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	// Get plugin and start lifecycle
	plugin, err := m.platform.GetPlugin(name)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// Start plugin lifecycle
	if err := m.lifecycleManager.StartPlugin(ctx, name, plugin, nil); err != nil {
		return fmt.Errorf("failed to start plugin lifecycle: %w", err)
	}

	return nil
}

// UnloadPlugin unloads a plugin by name
func (m *Manager) UnloadPlugin(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get plugin
	plugin, err := m.platform.GetPlugin(name)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	// Stop plugin lifecycle
	if err := m.lifecycleManager.StopPlugin(ctx, name, plugin); err != nil {
		return fmt.Errorf("failed to stop plugin lifecycle: %w", err)
	}

	// Unload plugin
	if err := m.platform.UnloadPlugin(ctx, name); err != nil {
		return fmt.Errorf("failed to unload plugin: %w", err)
	}

	return nil
}

// ExecutePlugin executes a task on a specific plugin
func (m *Manager) ExecutePlugin(ctx context.Context, pluginName, task string, input map[string]interface{}) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.platform.ExecutePlugin(ctx, pluginName, task, input)
}

// GetPlugin returns a plugin by name
func (m *Manager) GetPlugin(name string) (core.Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.platform.GetPlugin(name)
}

// ListPlugins returns all loaded plugins
func (m *Manager) ListPlugins() []core.PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.platform.ListPlugins()
}

// DiscoverPlugins discovers available plugins
func (m *Manager) DiscoverPlugins(ctx context.Context) ([]core.PluginMetadata, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.discovery == nil {
		return nil, fmt.Errorf("plugin discovery not configured")
	}

	return m.discovery.Discover(ctx, m.config.PluginDir)
}

// RegisterPlugin registers a plugin in the registry
func (m *Manager) RegisterPlugin(metadata core.PluginMetadata, dependencies []string, configSchema map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.registry.Register(metadata, dependencies, configSchema)
}

// UnregisterPlugin removes a plugin from the registry
func (m *Manager) UnregisterPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.registry.Unregister(name)
}

// Start starts the plugin manager
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Discover plugins if auto-load is enabled
	if m.config.AutoLoad {
		metadataList, err := m.DiscoverPlugins(ctx)
		if err != nil {
			return fmt.Errorf("failed to discover plugins: %w", err)
		}

		// Register discovered plugins
		for _, metadata := range metadataList {
			if err := m.RegisterPlugin(metadata, nil, nil); err != nil {
				// Log error but continue with other plugins
				continue
			}
		}

		// Load all registered plugins
		for _, metadata := range metadataList {
			if err := m.LoadPlugin(ctx, metadata.Name); err != nil {
				// Log error but continue with other plugins
				continue
			}
		}
	}

	// Start health checks
	go m.lifecycleManager.StartHealthChecks(ctx, m.platform)

	return nil
}

// Stop stops the plugin manager
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Shutdown platform
	return m.platform.Shutdown(ctx)
}

// GetRegistry returns the plugin registry
func (m *Manager) GetRegistry() *core.PluginRegistry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.registry
}

// GetLifecycleManager returns the lifecycle manager
func (m *Manager) GetLifecycleManager() *core.LifecycleManager {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.lifecycleManager
}

// Helper function to convert seconds to duration
func durationFromSeconds(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}

// PluginDiscovery is an interface for discovering plugins
type PluginDiscovery interface {
	Discover(ctx context.Context, pluginDir string) ([]core.PluginMetadata, error)
}
