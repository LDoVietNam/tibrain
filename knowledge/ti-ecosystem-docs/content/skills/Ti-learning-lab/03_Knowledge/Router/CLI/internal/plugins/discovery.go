package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ti/cli/internal/core"
)

// FileSystemDiscovery discovers plugins from the file system
type FileSystemDiscovery struct {
	pluginDir string
}

// NewFileSystemDiscovery creates a new file system discovery
func NewFileSystemDiscovery(pluginDir string) *FileSystemDiscovery {
	return &FileSystemDiscovery{
		pluginDir: pluginDir,
	}
}

// Discover discovers plugins from the file system
func (fsd *FileSystemDiscovery) Discover(ctx context.Context, pluginDir string) ([]core.PluginMetadata, error) {
	// Use provided pluginDir or default
	if pluginDir == "" {
		pluginDir = fsd.pluginDir
	}

	// Check if plugin directory exists
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin directory does not exist: %s", pluginDir)
	}

	// Scan directory for plugin manifests
	plugins, err := fsd.scanDirectory(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan plugin directory: %w", err)
	}

	return plugins, nil
}

// scanDirectory scans a directory for plugin manifests
func (fsd *FileSystemDiscovery) scanDirectory(dir string) ([]core.PluginMetadata, error) {
	var plugins []core.PluginMetadata

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Look for plugin manifest in subdirectory
			pluginDir := filepath.Join(dir, entry.Name())
			metadata, err := fsd.readPluginManifest(pluginDir)
			if err != nil {
				// Log error but continue with other plugins
				continue
			}

			if metadata != nil {
				plugins = append(plugins, *metadata)
			}
		} else if strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".json") {
			// Read manifest file
			manifestPath := filepath.Join(dir, entry.Name())
			metadata, err := fsd.readPluginManifestFile(manifestPath)
			if err != nil {
				// Log error but continue with other plugins
				continue
			}

			if metadata != nil {
				plugins = append(plugins, *metadata)
			}
		}
	}

	return plugins, nil
}

// readPluginManifest reads a plugin manifest from a directory
func (fsd *FileSystemDiscovery) readPluginManifest(pluginDir string) (*core.PluginMetadata, error) {
	// Look for manifest file (plugin.yaml or plugin.json)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		manifestPath = filepath.Join(pluginDir, "plugin.json")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			// No manifest file found
			return nil, nil
		}
	}

	return fsd.readPluginManifestFile(manifestPath)
}

// readPluginManifestFile reads a plugin manifest from a file
func (fsd *FileSystemDiscovery) readPluginManifestFile(manifestPath string) (*core.PluginMetadata, error) {
	// For now, return a placeholder metadata
	// In the future, this will parse YAML/JSON files

	// Extract plugin name from directory or filename
	name := filepath.Base(filepath.Dir(manifestPath))
	if name == "." {
		name = strings.TrimSuffix(filepath.Base(manifestPath), filepath.Ext(manifestPath))
	}

	metadata := &core.PluginMetadata{
		Name:         name,
		Version:      "1.0.0",
		Description:  "Auto-discovered plugin",
		Author:       "Unknown",
		Capabilities: []core.PluginCapability{},
		Labels:       make(map[string]string),
	}

	return metadata, nil
}

// StaticDiscovery is a discovery mechanism that returns a fixed list of plugins
type StaticDiscovery struct {
	plugins []core.PluginMetadata
}

// NewStaticDiscovery creates a new static discovery
func NewStaticDiscovery(plugins []core.PluginMetadata) *StaticDiscovery {
	return &StaticDiscovery{
		plugins: plugins,
	}
}

// Discover returns the static list of plugins
func (sd *StaticDiscovery) Discover(ctx context.Context, pluginDir string) ([]core.PluginMetadata, error) {
	return sd.plugins, nil
}

// AddPlugin adds a plugin to the static discovery
func (sd *StaticDiscovery) AddPlugin(metadata core.PluginMetadata) {
	sd.plugins = append(sd.plugins, metadata)
}

// RemovePlugin removes a plugin from the static discovery
func (sd *StaticDiscovery) RemovePlugin(name string) {
	for i, plugin := range sd.plugins {
		if plugin.Name == name {
			sd.plugins = append(sd.plugins[:i], sd.plugins[i+1:]...)
			break
		}
	}
}

// CompositeDiscovery combines multiple discovery mechanisms
type CompositeDiscovery struct {
	discoveries []PluginDiscovery
}

// NewCompositeDiscovery creates a new composite discovery
func NewCompositeDiscovery(discoveries ...PluginDiscovery) *CompositeDiscovery {
	return &CompositeDiscovery{
		discoveries: discoveries,
	}
}

// Discover discovers plugins using all registered discovery mechanisms
func (cd *CompositeDiscovery) Discover(ctx context.Context, pluginDir string) ([]core.PluginMetadata, error) {
	var allPlugins []core.PluginMetadata
	seen := make(map[string]bool)

	for _, discovery := range cd.discoveries {
		plugins, err := discovery.Discover(ctx, pluginDir)
		if err != nil {
			// Log error but continue with other discoveries
			continue
		}

		for _, plugin := range plugins {
			if !seen[plugin.Name] {
				allPlugins = append(allPlugins, plugin)
				seen[plugin.Name] = true
			}
		}
	}

	return allPlugins, nil
}

// AddDiscovery adds a discovery mechanism
func (cd *CompositeDiscovery) AddDiscovery(discovery PluginDiscovery) {
	cd.discoveries = append(cd.discoveries, discovery)
}

// DefaultDiscovery returns a default discovery mechanism
func DefaultDiscovery(pluginDir string) PluginDiscovery {
	// Create composite discovery with file system and static plugins
	composite := NewCompositeDiscovery(
		NewFileSystemDiscovery(pluginDir),
		NewStaticDiscovery([]core.PluginMetadata{
			{
				Name:        "devin",
				Version:     "1.0.0",
				Description: "Devin CLI plugin for complex automation",
				Author:      "Devin Team",
				Capabilities: []core.PluginCapability{
					core.CapabilityStatusDetection,
					core.CapabilityTmuxIntegration,
					core.CapabilityPermissionCheck,
					core.CapabilityPromptInjection,
					core.CapabilityMCPConfig,
					core.CapabilityOutputExtraction,
					core.CapabilityTUIDetection,
				},
				Labels: map[string]string{
					"type":     "external",
					"language": "python",
				},
			},
		}),
	)

	return composite
}
