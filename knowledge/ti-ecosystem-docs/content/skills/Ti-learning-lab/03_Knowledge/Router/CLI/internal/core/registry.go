package core

import (
	"context"
	"fmt"
	"sync"
)

// PluginRegistry manages plugin metadata and dependencies
type PluginRegistry struct {
	mu           sync.RWMutex
	plugins      map[string]*PluginRegistryEntry
	byCapability map[PluginCapability][]string
}

// PluginRegistryEntry contains registry information for a plugin
type PluginRegistryEntry struct {
	Metadata     PluginMetadata
	Dependencies []string
	Version      string
	Enabled      bool
	ConfigSchema map[string]interface{}
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	registry := &PluginRegistry{
		plugins:      make(map[string]*PluginRegistryEntry),
		byCapability: make(map[PluginCapability][]string),
	}

	// Register built-in plugins
	registry.registerBuiltinPlugins()

	return registry
}

// registerBuiltinPlugins registers all built-in plugins
func (r *PluginRegistry) registerBuiltinPlugins() {
	// Devin plugin
	r.Register(PluginMetadata{
		Name:        "devin",
		Version:     "1.0.0",
		Type:        "agent",
		Description: "Devin AI agent plugin",
		Author:      "Ti CLI",
		Capabilities: []PluginCapability{
			CapabilityPromptInjection,
			CapabilityOutputExtraction,
		},
	}, []string{}, nil)

	// Status detection plugin
	r.Register(PluginMetadata{
		Name:        "status",
		Version:     "1.0.0",
		Type:        "native",
		Description: "Native status detection plugin",
		Author:      "Ti CLI",
		Capabilities: []PluginCapability{
			CapabilityStatusDetection,
		},
	}, []string{}, nil)

	// Tmux integration plugin
	r.Register(PluginMetadata{
		Name:        "tmux",
		Version:     "1.0.0",
		Type:        "native",
		Description: "Native tmux integration plugin",
		Author:      "Ti CLI",
		Capabilities: []PluginCapability{
			CapabilityTmuxControl,
		},
	}, []string{}, nil)

	// Logger plugin
	r.Register(PluginMetadata{
		Name:        "logger",
		Version:     "1.0.0",
		Type:        "native",
		Description: "Native structured logging plugin",
		Author:      "Ti CLI",
		Capabilities: []PluginCapability{
			CapabilityLogging,
		},
	}, []string{}, nil)

	// Permission manager plugin
	r.Register(PluginMetadata{
		Name:        "permission",
		Version:     "1.0.0",
		Type:        "native",
		Description: "Native permission management plugin",
		Author:      "Ti CLI",
		Capabilities: []PluginCapability{
			CapabilityPermissionCheck,
		},
	}, []string{}, nil)
}

// Register registers a plugin in the registry
func (r *PluginRegistry) Register(metadata PluginMetadata, dependencies []string, configSchema map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if plugin is already registered
	if _, exists := r.plugins[metadata.Name]; exists {
		return fmt.Errorf("plugin %s is already registered", metadata.Name)
	}

	// Create registry entry
	entry := &PluginRegistryEntry{
		Metadata:     metadata,
		Dependencies: dependencies,
		Version:      metadata.Version,
		Enabled:      true,
		ConfigSchema: configSchema,
	}

	// Add to plugins map
	r.plugins[metadata.Name] = entry

	// Index by capability
	for _, capability := range metadata.Capabilities {
		r.byCapability[capability] = append(r.byCapability[capability], metadata.Name)
	}

	return nil
}

// Unregister removes a plugin from the registry
func (r *PluginRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found in registry", name)
	}

	// Remove from capability index
	for _, capability := range entry.Metadata.Capabilities {
		names := r.byCapability[capability]
		for i, n := range names {
			if n == name {
				r.byCapability[capability] = append(names[:i], names[i+1:]...)
				break
			}
		}
	}

	// Remove from plugins map
	delete(r.plugins, name)

	return nil
}

// GetMetadata retrieves metadata for a plugin
func (r *PluginRegistry) GetMetadata(name string) (PluginMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.plugins[name]
	if !exists {
		return PluginMetadata{}, fmt.Errorf("plugin %s not found in registry", name)
	}

	if !entry.Enabled {
		return PluginMetadata{}, fmt.Errorf("plugin %s is disabled", name)
	}

	return entry.Metadata, nil
}

// GetEntry retrieves the full registry entry for a plugin
func (r *PluginRegistry) GetEntry(name string) (*PluginRegistryEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found in registry", name)
	}

	return entry, nil
}

// ListPlugins returns all registered plugins
func (r *PluginRegistry) ListPlugins() []PluginMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata := make([]PluginMetadata, 0, len(r.plugins))
	for _, entry := range r.plugins {
		if entry.Enabled {
			metadata = append(metadata, entry.Metadata)
		}
	}

	return metadata
}

// ListPluginsByCapability returns all plugins that provide a specific capability
func (r *PluginRegistry) ListPluginsByCapability(capability PluginCapability) []PluginMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := r.byCapability[capability]
	metadata := make([]PluginMetadata, 0, len(names))

	for _, name := range names {
		if entry, exists := r.plugins[name]; exists && entry.Enabled {
			metadata = append(metadata, entry.Metadata)
		}
	}

	return metadata
}

// ResolveDependencies resolves dependencies for a plugin
func (r *PluginRegistry) ResolveDependencies(name string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found in registry", name)
	}

	// Check if all dependencies are available
	for _, dep := range entry.Dependencies {
		if _, exists := r.plugins[dep]; !exists {
			return nil, fmt.Errorf("dependency %s not found for plugin %s", dep, name)
		}
	}

	return entry.Dependencies, nil
}

// Enable enables a plugin
func (r *PluginRegistry) Enable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found in registry", name)
	}

	entry.Enabled = true
	return nil
}

// Disable disables a plugin
func (r *PluginRegistry) Disable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found in registry", name)
	}

	entry.Enabled = false
	return nil
}

// ValidateConfig validates a plugin configuration against its schema
func (r *PluginRegistry) ValidateConfig(name string, config map[string]interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found in registry", name)
	}

	// Basic validation - check if required fields are present
	if entry.ConfigSchema != nil {
		for key, schema := range entry.ConfigSchema {
			if schemaMap, ok := schema.(map[string]interface{}); ok {
				if required, ok := schemaMap["required"].(bool); ok && required {
					if _, exists := config[key]; !exists {
						return fmt.Errorf("required config field %s is missing for plugin %s", key, name)
					}
				}
			}
		}
	}

	return nil
}

// GetVersion returns the version of a plugin
func (r *PluginRegistry) GetVersion(name string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.plugins[name]
	if !exists {
		return "", fmt.Errorf("plugin %s not found in registry", name)
	}

	return entry.Version, nil
}

// CheckCompatibility checks if a plugin version is compatible with dependencies
func (r *PluginRegistry) CheckCompatibility(name string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found in registry", name)
	}

	// Check each dependency
	for _, dep := range entry.Dependencies {
		depEntry, exists := r.plugins[dep]
		if !exists {
			return fmt.Errorf("dependency %s not found", dep)
		}

		if !depEntry.Enabled {
			return fmt.Errorf("dependency %s is disabled", dep)
		}

		// Here you could add version compatibility checking
		// For now, we just check if the dependency exists and is enabled
	}

	return nil
}

// PluginLoader is an interface for loading plugins
type PluginLoader interface {
	Load(ctx context.Context, metadata PluginMetadata) (Plugin, error)
}

// PluginRegistry is the interface for plugin registry operations
type PluginRegistryInterface interface {
	Register(metadata PluginMetadata, dependencies []string, configSchema map[string]interface{}) error
	Unregister(name string) error
	GetMetadata(name string) (PluginMetadata, error)
	ListPlugins() []PluginMetadata
	ListPluginsByCapability(capability PluginCapability) []PluginMetadata
	ResolveDependencies(name string) ([]string, error)
	Enable(name string) error
	Disable(name string) error
	ValidateConfig(name string, config map[string]interface{}) error
}
