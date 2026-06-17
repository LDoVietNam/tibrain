# Ti CLI Plugin Architecture

> **Last Updated**: 2026-05-05
> **Version**: 1.0.0
> **Status**: Phase 1 Complete

## Overview

The Ti CLI plugin system provides a modular, extensible architecture for adding functionality to the CLI through plugins. The system supports:

- **Dependency Resolution**: Automatic loading of plugin dependencies
- **Dependency Injection**: Service locator pattern for dependency management
- **Lifecycle Management**: Health checks, graceful shutdown, restart policies
- **Hot Reload**: Configuration changes without restarting the CLI
- **Feature Flags**: Gradual rollout of new features

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                      Plugin Manager                          │
│  - Load/Unload plugins                                       │
│  - Dependency resolution                                     │
│  - Lifecycle management                                      │
│  - Health monitoring                                         │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
        ┌─────▼─────┐  ┌────▼────┐  ┌─────▼──────┐
        │  Registry  │  │  Loader │  │ DI Container│
        │            │  │         │  │             │
        │ - Metadata │  │ - Create│  │ - Services  │
        │ - Deps     │  │ - Load  │  │ - Singleton │
        │ - Version  │  │ - Unload│  │ - Locator   │
        └─────┬─────┘  └────┬────┘  └─────┬──────┘
              │            │              │
              └────────────┴──────────────┘
                             │
                    ┌────────▼────────┐
                    │   Plugins       │
                    │                 │
                    │ - featureflag   │
                    │ - logging       │
                    │ - configwatcher │
                    │ - devin         │
                    │ - native agents │
                    └─────────────────┘
```

### Plugin Interface

All plugins must implement the `Plugin` interface defined in `internal/core/platform.go`:

```go
type Plugin interface {
    // Identity
    Name() string
    Version() string
    Metadata() PluginMetadata

    // Dependencies
    Dependencies() []string

    // Lifecycle
    Initialize(ctx context.Context, config map[string]string) error
    Execute(ctx context.Context, task string, input map[string]interface{}) (map[string]interface{}, error)
    Shutdown(ctx context.Context) error
    HealthCheck(ctx context.Context) error
    State() PluginState
}
```

### Plugin States

Plugins transition through the following states:

```
Stopped → Initializing → Ready → Busy → Ready
                     ↓                     ↓
                   Error ←───────── Shutdown
```

- **Stopped**: Plugin is not loaded
- **Initializing**: Plugin is being initialized
- **Ready**: Plugin is ready to execute tasks
- **Busy**: Plugin is currently executing a task
- **Error**: Plugin encountered an error
- **Shutdown**: Plugin is shutting down

## Dependency Resolution

### Declaration

Plugins declare their dependencies using the `Dependencies()` method:

```go
func (p *Plugin) Dependencies() []string {
    return []string{"featureflag", "logging"}
}
```

### Resolution Process

When loading a plugin with dependencies:

1. Manager queries the registry for dependency information
2. Registry performs topological sort to determine load order
3. Manager loads dependencies recursively
4. Each dependency is registered in the DI container
5. Finally, the target plugin is loaded

### Example

```go
// configwatcher depends on featureflag and logging
// logging depends on featureflag

manager.LoadPlugin(ctx, "configwatcher")
// Automatically loads:
// 1. featureflag (no dependencies)
// 2. logging (depends on featureflag)
// 3. configwatcher (depends on featureflag and logging)
```

### Circular Dependency Detection

The registry detects circular dependencies during resolution and returns an error:

```go
// Plugin A depends on B
// Plugin B depends on A
// Error: circular dependency detected: A → B → A
```

## Dependency Injection

### DI Container

The DI container (`internal/di/container.go`) provides:

- **Service Registration**: Register singleton services
- **Service Retrieval**: Get services by name
- **Service Locator**: Check if a service exists

### Registration

Plugins are automatically registered in the DI container when loaded:

```go
// Manager.LoadPlugin() automatically registers:
diContainer.RegisterSingleton("featureflag", plugin)
```

### Usage

Plugins can retrieve dependencies from the DI container:

```go
func (p *Plugin) Initialize(ctx context.Context, config map[string]interface{}) error {
    // Get feature flag plugin from DI container
    featureFlagPlugin, err := p.diContainer.Get("featureflag")
    if err != nil {
        return err
    }

    // Use the plugin
    result, err := featureFlagPlugin.Execute(ctx, "get_flag", map[string]interface{}{
        "name": "structured_logging",
    })
    // ...
}
```

### Custom Services

Applications can register custom services:

```go
manager.RegisterService("custom-service", customImplementation)
service, err := manager.GetService("custom-service")
```

## Built-in Infrastructure Plugins

### FeatureFlagPlugin

**Location**: `internal/plugins/featureflag/`

**Purpose**: Feature flag management for gradual rollout

**Features**:
- File-based and memory-based storage
- Flag management with rollout percentages
- Default flag configuration

**Dependencies**: None

**Example Usage**:
```go
// Check if a feature is enabled
result, _ := plugin.Execute(ctx, "get_flag", map[string]interface{}{
    "name": "distributed_tracing",
})
enabled := result["enabled"].(bool)

// Set a flag
plugin.Execute(ctx, "set_flag", map[string]interface{}{
    "name": "new_feature",
    "enabled": true,
    "rollout_percentage": 50,
})
```

**Configuration**: `~/.ti-cli/plugins/featureflag.yaml`

```yaml
storage:
  type: file
  path: ~/.ti-cli/flags.json

default_flags:
  - name: distributed_tracing
    enabled: false
    description: Enable distributed tracing
    rollout_percentage: 0

  - name: structured_logging
    enabled: true
    description: Enable structured logging
    rollout_percentage: 100

  - name: hot_reload_config
    enabled: false
    description: Enable hot-reload of configuration
    rollout_percentage: 0

  - name: enhanced_metrics
    enabled: false
    description: Enable enhanced metrics collection
    rollout_percentage: 0
```

### StructuredLoggingPlugin

**Location**: `internal/plugins/logging/`

**Purpose**: Structured logging with multiple levels

**Features**:
- 4-level logging (debug, info, warn, error)
- Structured output with fields
- Feature flag integration

**Dependencies**: `["featureflag"]`

**Example Usage**:
```go
// Log messages
plugin.Execute(ctx, "log", map[string]interface{}{
    "level": "info",
    "message": "Plugin initialized",
    "fields": map[string]interface{}{
        "plugin": "my-plugin",
        "version": "1.0.0",
    },
})
```

**Configuration**: `~/.ti-cli/plugins/logging.yaml`

```yaml
level: info
format: json
output: stdout
feature_flag: structured_logging
```

### HotReloadConfigPlugin

**Location**: `internal/plugins/configwatcher/`

**Purpose**: Hot-reload configuration files

**Features**:
- File watching with fsnotify
- Automatic reload on change
- Integration with logging

**Dependencies**: `["featureflag", "logging"]`

**Example Usage**:
```go
// Start watching a config file
plugin.Execute(ctx, "watch", map[string]interface{}{
    "path": "/path/to/config.yaml",
})
```

**Configuration**: `~/.ti-cli/plugins/configwatcher.yaml`

```yaml
watch_paths:
  - ~/.ti-cli/config.yaml
  - ~/.ti-cli/providers.yaml

debounce: 1s
feature_flag: hot_reload_config
```

## Plugin Development Guide

### Creating a New Plugin

1. **Create plugin directory**:
   ```
   internal/plugins/myplugin/
   ```

2. **Implement plugin interface**:
   ```go
   package myplugin

   type Plugin struct {
       name     string
       version  string
       config   *Config
       state    core.PluginState
       metadata core.PluginMetadata
   }

   func NewPlugin(config *Config) (*Plugin, error) {
       return &Plugin{
           name:    "myplugin",
           version: "1.0.0",
           config:  config,
           state:   core.PluginStateStopped,
           metadata: core.PluginMetadata{
               Name:        "myplugin",
               Version:     "1.0.0",
               Type:        "custom",
               Description: "My custom plugin",
               Author:      "Your Name",
           },
       }, nil
   }

   // Implement all interface methods...
   ```

3. **Register in loader**:
   ```go
   // internal/plugins/loader.go
   func (dl *DefaultLoader) createPlugin(metadata core.PluginMetadata) (core.Plugin, error) {
       switch metadata.Name {
       // ...
       case "myplugin":
           return myplugin.NewPlugin(nil)
       // ...
       }
   }
   ```

4. **Register in registry**:
   ```go
   // internal/core/registry.go
   func (r *PluginRegistry) registerBuiltinPlugins() {
       r.Register(core.PluginMetadata{
           Name:        "myplugin",
           Version:     "1.0.0",
           Type:        "custom",
           Description: "My custom plugin",
           Author:      "Your Name",
       }, []string{}, nil)
   }
   ```

5. **Create config file** (optional):
   ```yaml
   # ~/.ti-cli/plugins/myplugin.yaml
   setting1: value1
   setting2: value2
   ```

6. **Write tests**:
   ```go
   package myplugin

   import "testing"

   func TestPluginInitialize(t *testing.T) {
       plugin, err := NewPlugin(nil)
       if err != nil {
           t.Fatalf("failed to create plugin: %v", err)
       }

       err = plugin.Initialize(context.Background(), nil)
       if err != nil {
           t.Errorf("failed to initialize plugin: %v", err)
       }
   }
   ```

### Best Practices

1. **Dependencies**: Declare all dependencies in `Dependencies()` method
2. **Error Handling**: Return detailed errors with context
3. **State Management**: Use mutex for concurrent access
4. **Health Checks**: Implement meaningful health checks
5. **Configuration**: Provide sensible defaults
6. **Testing**: Write comprehensive unit tests
7. **Documentation**: Document all tasks and parameters

### Testing Plugins

```bash
# Run plugin tests
cd apps/cli
go test ./internal/plugins/myplugin/ -v

# Run manager tests
go test ./internal/plugins/manager_test.go -v
```

## Configuration

### Plugin Directory

Plugins are configured in `~/.ti-cli/plugins/`:

```
~/.ti-cli/plugins/
├── featureflag.yaml
├── logging.yaml
└── configwatcher.yaml
```

### Manager Configuration

```go
config := &plugins.Config{
    AutoLoad:          false,  // Auto-load plugins on startup
    HealthCheckInterval: 30,   // Health check interval (seconds)
    HealthCheckTimeout: 10,    // Health check timeout (seconds)
    MaxRestarts:       3,     // Maximum restart attempts
    RestartDelay:      5,     // Delay between restarts (seconds)
}
```

## Lifecycle Management

### Startup

```go
manager := plugins.NewManager(config)
manager.SetLoader(plugins.NewDefaultLoader())

// Load plugins manually or auto-load
if config.AutoLoad {
    manager.Start(ctx)
}
```

### Shutdown

```go
manager.Stop(ctx)
```

### Health Checks

Health checks run automatically at the configured interval:

```go
// Manual health check
plugin.HealthCheck(ctx)

// Get lifecycle state
lifecycle, _ := manager.lifecycleManager.GetLifecycle("featureflag")
fmt.Println(lifecycle.State)
```

## Error Handling

### Plugin Load Errors

```go
err := manager.LoadPlugin(ctx, "myplugin")
if err != nil {
    // Handle error
    // Common errors:
    // - Plugin not found
    // - Circular dependency
    // - Initialization failed
}
```

### Execution Errors

```go
result, err := manager.ExecutePlugin(ctx, "myplugin", "task", nil)
if err != nil {
    // Handle error
    // Common errors:
    // - Task not found
    // - Invalid input
    // - Plugin not ready
}
```

## Future Enhancements

### Phase 2 (Planned)

- **Plugin Marketplace**: Discover and install plugins from a registry
- **Version Management**: Plugin version constraints and upgrades
- **Sandboxing**: Isolated plugin execution
- **Hot Reload**: Reload plugins without restarting CLI
- **Plugin Events**: Event system for plugin communication

### Phase 3 (Future)

- **Distributed Plugins**: Plugins that run as separate processes
- **Plugin Composition**: Combine multiple plugins into workflows
- **Plugin Metrics**: Performance monitoring for plugins
- **Plugin Debugging**: Enhanced debugging tools

## References

- **Plugin Interface**: `internal/core/platform.go`
- **Plugin Manager**: `internal/plugins/manager.go`
- **Plugin Loader**: `internal/plugins/loader.go`
- **Plugin Registry**: `internal/core/registry.go`
- **DI Container**: `internal/di/container.go`
- **Lifecycle Manager**: `internal/core/lifecycle.go`

## Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - Overall CLI architecture
- [DEPLOYMENT_MAINTENANCE_STRATEGY.md](DEPLOYMENT_MAINTENANCE_STRATEGY.md) - Deployment strategies
- [CLI_GUIDE.md](CLI_GUIDE.md) - CLI usage guide
