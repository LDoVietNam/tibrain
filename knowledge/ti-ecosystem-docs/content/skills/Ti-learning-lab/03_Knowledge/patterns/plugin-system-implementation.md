# Plugin System Implementation - Hệ Thống Plugin Động

> **Ngày tạo**: 2026-04-29  
> **Project**: ti-router  
> **Mục tiêu**: Implement plugin system để load providers dynamically mà không cần build lại

---

## Tổng Quan

Plugin system cho phép load providers mới mà không cần build lại router. Plugins được load như separate processes và communicate qua net/rpc (theo pattern của HashiCorp go-plugin).

### Vấn Đề Giải Quyết

**Trước khi implement:**
- ❌ Vẫn cần build lại nếu add provider với code implementation mới (bootstrap.go)
- ❌ Config reload chỉ work cho providers đã được implement
- ❌ Env vars cần được set đúng cho API keys

**Sau khi implement:**
- ✅ Add provider mới chỉ cần:
  1. Build plugin binary riêng
  2. Add config vào providers.yaml
  3. POST /api/config/reload
- ✅ Không cần build lại router
- ✅ Plugins được load dynamically
- ✅ Hot reload support

---

## Architecture

### Components

```
plugin/
├── interface.go      # ProviderPlugin interface definition
├── loader.go         # Plugin discovery và loading
├── constants.go      # Plugin constants và RPC setup
├── rpc.go           # net/rpc server và client implementations
├── registry.go      # Plugin registry
├── config.go        # Plugin config loading
└── adapter.go      # Provider adapter cho integration
```

### Pattern

Sử dụng **HashiCorp go-plugin pattern**:
- Plugins là separate processes
- Communication qua net/rpc (built-in Go)
- Interface-based plugin loading
- Protocol versioning cho compatibility
- Process isolation (plugin crash không crash router)

### Workflow

```
1. Load plugin configs từ providers.yaml
2. Initialize plugin registry
3. Load plugins (launch processes)
4. Discover plugins (convert to providers.Provider)
5. Register với main registry
6. Config reload → Reload plugins
```

---

## Implementation Details

### 1. Plugin Interface (interface.go)

```go
type ProviderPlugin interface {
    // Plugin metadata
    PluginName() string
    PluginVersion() string
    PluginAuthor() string

    // Provider interface
    Name() string
    Status() string
    DefaultModel() string
    Models() []string
    IsHealthy() bool
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error)

    // Plugin lifecycle
    InitPlugin(ctx context.Context, config map[string]any) error
    ShutdownPlugin(ctx context.Context) error

    // Plugin capabilities
    Capabilities() PluginCapabilities
}
```

**Key points:**
- Extends Provider interface với plugin-specific methods
- Lifecycle methods: InitPlugin, ShutdownPlugin
- Capabilities để describe plugin features

### 2. Plugin Loader (loader.go)

```go
type PluginLoader struct {
    pluginDir string
    plugins   map[string]*Plugin
    mu        sync.RWMutex
    clients   map[string]*plugin.Client
}
```

**Features:**
- LoadPlugin() - Load single plugin
- LoadPlugins() - Load multiple plugins
- UnloadPlugin() - Unload single plugin
- ReloadPlugin() - Reload plugin (unload + load)
- GetPlugin() - Get loaded plugin
- ListPlugins() - List all loaded plugins

**Key implementation:**
- Uses hashicorp/go-plugin library
- net/rpc protocol (built-in Go)
- Plugin processes launched via exec.Command
- RPC client để communicate với plugin

### 3. RPC Implementation (rpc.go)

**ProviderRPCServer:**
- Implements net/rpc server methods
- Wraps ProviderPlugin implementation
- Converts requests/responses

**ProviderRPCClient:**
- Implements net/rpc client methods
- Calls plugin methods via RPC
- JSON serialization cho complex types

**Note:** Streaming không được support trong net/rpc, fallback to Chat()

### 4. Plugin Registry (registry.go)

```go
type PluginRegistry struct {
    loader  *PluginLoader
    plugins map[string]*Plugin
    mu      sync.RWMutex
}
```

**Features:**
- LoadPlugins() - Load plugins từ config
- GetPlugin() - Get plugin by name
- ListPlugins() - List all plugins
- ReloadPlugin() - Reload single plugin
- UnloadPlugin() - Unload plugin
- UnloadAll() - Unload all plugins
- HasPlugin() - Check if plugin loaded

### 5. Plugin Config (config.go)

```yaml
plugins:
  - name: example_provider
    enabled: false
    path: plugins/example_provider.exe
    config:
      api_key_env: EXAMPLE_API_KEY
      base_url: https://api.example.com/v1
    priority: 100
```

**Features:**
- LoadPluginConfigsFromFile() - Load từ YAML
- SavePluginConfigsToFile() - Save to YAML
- Resolve relative paths

### 6. Provider Adapter (adapter.go)

```go
type ProviderAdapter struct {
    plugin ProviderPlugin
}
```

**Purpose:** Convert ProviderPlugin → providers.Provider

**Features:**
- Implements providers.Provider interface
- Converts types (plugin.ChatRequest → providers.ChatRequest)
- Handles streaming conversion
- Wraps plugin calls

### 7. Integration với Bootstrap (bootstrap.go)

```go
func loadPluginsFromConfig(cfg *config.Config, reg *Registry, issues *[]BootstrapIssue) error {
    // 1. Load plugin configs từ providers.yaml
    _, pluginConfigs, err := LoadFromYAML("configs/providers.yaml")
    
    // 2. Initialize plugin registry
    pluginRegistry := plugin.NewPluginRegistry("plugins")
    
    // 3. Load plugins
    pluginRegistry.LoadPlugins(ctx, configs)
    
    // 4. Discover plugins
    providerPlugins, err := plugin.DiscoverPlugins(ctx, pluginRegistry)
    
    // 5. Register với main registry
    for _, p := range providerPlugins {
        reg.Register(p)
    }
}
```

**Integration points:**
- BuildRegistryFromConfig() calls loadPluginsFromConfig()
- Config reload calls BuildRegistryFromConfig() → reloads plugins
- Plugins loaded alongside existing providers

---

## Config Reload Integration

### Trước

```go
// handlers_admin.go
newProviderConfigs, err := providers.LoadFromYAML(providerConfigsPath)
newRegistry, issues := providers.BuildRegistryFromConfig(nil, nil)
providerRegistry = newRegistry
```

### Sau

```go
// handlers_admin.go
newProviderConfigs, _, err := providers.LoadFromYAML(providerConfigsPath)  // Now returns plugin configs too
newRegistry, issues := providers.BuildRegistryFromConfig(nil, nil)  // Now loads plugins
providerRegistry = newRegistry
```

**Workflow:**
1. POST /api/config/reload
2. Load providers.yaml (cả providers và plugins)
3. BuildRegistryFromConfig() → loadPluginsFromConfig()
4. Reload plugins (unload old, load new)
5. Update global registry
6. Sync model registry

---

## Usage Example

### 1. Tạo Plugin Binary

```go
// plugins/my_provider/main.go
package main

import (
    "github.com/hashicorp/go-plugin"
    "github.com/ti/router/layers/provider/plugin"
)

type MyProvider struct{}

func (p *MyProvider) PluginName() string { return "my_provider" }
func (p *MyProvider) PluginVersion() string { return "1.0.0" }
func (p *MyProvider) PluginAuthor() string { return "Ti Team" }

// ... implement other ProviderPlugin methods ...

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: plugin.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            plugin.PluginName: &plugin.ProviderPluginImpl{Impl: &MyProvider{}},
        },
    })
}
```

Build:
```bash
cd plugins/my_provider
go build -o my_provider.exe
```

### 2. Add Config

```yaml
# configs/providers.yaml
plugins:
  - name: my_provider
    enabled: true
    path: plugins/my_provider/my_provider.exe
    config:
      api_key_env: MY_PROVIDER_API_KEY
      base_url: https://api.my-provider.com/v1
    priority: 100
```

### 3. Reload Config

```bash
curl -X POST http://localhost:1807/api/config/reload \
  -H "Authorization: Bearer sk-jarvis-dev"
```

### 4. Test

```bash
curl http://localhost:1807/v1/models \
  -H "Authorization: Bearer sk-jarvis-dev"
```

---

## Lessons Learned

### 1. HashiCorp go-plugin Pattern

**Ưu điểm:**
- Battle-tested (used by Terraform, Vault, Nomad)
- Process isolation (plugin crash không crash host)
- Cross-language support (via gRPC)
- Protocol versioning
- Built-in logging và TTY preservation

**Nhược điểm:**
- Performance overhead (IPC communication)
- Complex setup (proto buffers cho gRPC)
- Streaming support limited trong net/rpc

### 2. net/rpc vs gRPC

**net/rpc (chosen):**
- ✅ Built-in Go (no external dependencies)
- ✅ Simpler setup (no proto buffers)
- ✅ Good enough cho simple use cases
- ❌ No streaming support
- ❌ Cross-language support limited

**gRPC (not chosen):**
- ✅ Streaming support
- ✅ Cross-language support
- ✅ Better performance
- ❌ Requires protoc (not available in environment)
- ❌ Complex setup (proto files, code generation)

### 3. Type Conversion

**Challenge:** Plugin types vs providers types

**Solution:** ProviderAdapter pattern
- Convert plugin.ChatRequest → providers.ChatRequest
- Convert plugin.ChatResponse → providers.ChatResponse
- Handle nested types (Message, ToolDef, Usage, ThinkingBlock)

### 4. Config Integration

**Challenge:** Load plugin configs từ existing YAML

**Solution:** Extend YAMLConfig struct
- Add Plugins field ([]YAMLPluginConfig)
- Update LoadFromYAML() signature
- Update all call sites to handle new return value

### 5. Concurrency Safety

**Pattern:** sync.RWMutex cho shared state
- PluginLoader: protects plugins map
- PluginRegistry: protects plugins map
- Read locks cho Get operations
- Write locks cho Load/Unload operations

---

## Security Considerations

### 1. Process Isolation

**Benefit:** Plugin crash không crash router
**Risk:** Plugin có thể consume excessive resources
**Mitigation:** Resource limits (future enhancement)

### 2. RPC Communication

**Risk:** Man-in-the-middle attacks
**Mitigation:** TLS/mTLS (future enhancement)
**Current:** Local-only communication (acceptable)

### 3. Plugin Path

**Risk:** Path traversal attacks
**Mitigation:** Validate plugin paths
**Current:** Relative paths resolved from config directory

### 4. Config Secrets

**Risk:** Secrets in config files
**Mitigation:** Use env vars (api_key_env)
**Current:** Config supports both direct keys và env vars

---

## Performance Considerations

### 1. IPC Overhead

**Impact:** Each plugin call requires IPC
**Mitigation:** Cache plugin instances
**Current:** Plugin instances cached in registry

### 2. Process Launch

**Impact:** Launching plugin processes takes time
**Mitigation:** Load plugins at startup, not per-request
**Current:** Plugins loaded in BuildRegistryFromConfig()

### 3. Memory Usage

**Impact:** Each plugin process consumes memory
**Mitigation:** Limit number of plugins
**Current:** No limit (future enhancement)

---

## Future Enhancements

### 1. Streaming Support

**Current:** Fallback to Chat() trong net/rpc
**Future:** Switch to gRPC for streaming support

### 2. Resource Limits

**Current:** No resource limits
**Future:** CPU, memory, file descriptor limits per plugin

### 3. Plugin Verification

**Current:** No plugin verification
**Future:** Checksum verification, signature verification

### 4. Plugin Health Checks

**Current:** Basic IsHealthy() check
**Future:** Periodic health checks, auto-restart unhealthy plugins

### 5. Plugin Hot Reload

**Current:** Full reload (unload all, load all)
**Future:** Incremental reload (only reload changed plugins)

### 6. Plugin Discovery

**Current:** Manual config in YAML
**Future:** Auto-discovery from plugins/ directory

---

## Files Modified/Created

### Created
- `router/layers/provider/plugin/interface.go`
- `router/layers/provider/plugin/loader.go`
- `router/layers/provider/plugin/constants.go`
- `router/layers/provider/plugin/rpc.go`
- `router/layers/provider/plugin/registry.go`
- `router/layers/provider/plugin/config.go`
- `router/layers/provider/plugin/adapter.go`
- `router/layers/provider/plugin/proto/provider.proto` (not used - net/rpc instead)

### Modified
- `router/layers/provider/bootstrap.go` - Added loadPluginsFromConfig()
- `router/layers/provider/config.go` - Added Plugins field to YAMLConfig
- `router/configs/providers.yaml` - Added plugins section
- `router/go.mod` - Added hashicorp/go-plugin dependency
- `router/cmd/routerd/handlers_admin.go` - Updated LoadFromYAML call
- `router/cmd/routerd/router.go` - Updated LoadFromYAML call

---

## References

- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin) - Plugin system library
- [go-plugin examples](https://github.com/hashicorp/go-plugin/tree/main/examples) - Example implementations
- [Go net/rpc](https://pkg.go.dev/net/rpc) - Built-in RPC package

---

**Status:** ✅ Implementation hoàn tất, chờ test với plugin binary thực tế
