# Plugin Auto-Discovery - Tự Động Phát Hiện Plugins

> **Ngày tạo**: 2026-04-29  
> **Project**: ti-router  
> **Mục tiêu**: Auto-discovery feature để load plugins từ external directory mà không cần config

---

## Tổng Quan

Auto-discovery feature cho phép router tự động scan và load tất cả plugin binaries từ `plugins/` directory mà không cần config trong YAML. Đây là enhancement cho plugin system đã implement trước đó.

### Workflow

```
1. Router khởi động
2. Load providers.yaml
3. Check nếu có plugin configs trong YAML
   - Nếu CÓ: Load plugins từ config (config-based mode)
   - Nếu KHÔNG: Auto-discover từ plugins/ directory (auto-discovery mode)
4. Auto-discovery mode:
   - Scan plugins/ directory cho executables
   - Load tất cả plugin binaries tìm thấy
   - Register với main registry
5. Config reload → Re-run auto-discovery
```

---

## Implementation

### Files Created

**`router/layers/provider/plugin/discovery.go`**

**Components:**

1. **PluginScanner** - Scan directory cho plugin binaries
   ```go
   type PluginScanner struct {
       pluginDir string
       extensions []string // File extensions (.exe on Windows, none on Unix)
   }
   ```

2. **AutoDiscoverPlugins** - Tự động discover và load plugins
   ```go
   func AutoDiscoverPlugins(ctx context.Context, pluginDir string, loader *PluginLoader) ([]*Plugin, error)
   ```

**Key features:**
- Cross-platform support (.exe on Windows, no extension on Unix)
- Automatic extension detection
- Error tolerance (continue loading other plugins if one fails)
- Warning logging cho failed plugins

### Integration

**File:** `router/layers/provider/plugin/bootstrap.go`

**Updated `loadPluginsFromConfig()` function:**

```go
func loadPluginsFromConfig(cfg *config.Config, reg *Registry, issues *[]BootstrapIssue) error {
    // Load plugin configs từ providers.yaml
    _, pluginConfigs, err := LoadFromYAML("configs/providers.yaml")
    
    if len(pluginConfigs) > 0 {
        // Config-based mode: Load plugins từ YAML config
        pluginRegistry.LoadPlugins(ctx, configs)
    } else {
        // Auto-discovery mode: Scan plugins/ directory
        loader := plugin.NewPluginLoader("plugins")
        plugins, err := plugin.AutoDiscoverPlugins(ctx, "plugins", loader)
        
        // Add discovered plugins to registry
        for _, p := range plugins {
            pluginRegistry.loader.plugins[p.Metadata.Name] = p
        }
        
        fmt.Printf("[Plugin System] Auto-discovered %d plugins from plugins/ directory\n", len(plugins))
    }
    
    // Register plugins với main registry
    plugin.DiscoverPlugins(ctx, pluginRegistry)
}
```

---

## Config Modes

### Auto-Discovery Mode (Default)

**YAML:**
```yaml
# configs/providers.yaml
# plugins section commented out hoặc empty
# plugins:
```

**Workflow:**
```bash
# 1. Tạo plugins/ directory
mkdir plugins

# 2. Build plugin binary
cd plugins/my_provider
go build -o my_provider.exe  # Windows
go build -o my_provider       # Linux/Mac

# 3. Start router (tự động discover plugins)
cd router
.\routerd.exe -config configs\Tiserverrouter.yaml

# Output:
# [Plugin System] Auto-discovered 1 plugins from plugins/ directory
```

**Advantages:**
- ✅ Zero config - chỉ cần drop plugin binary vào directory
- ✅ Fast iteration - build plugin, drop, reload
- ✅ Simple workflow - không cần edit YAML
- ✅ Dynamic - tự động detect plugins mới

### Config-Based Mode (Optional)

**YAML:**
```yaml
# configs/providers.yaml
plugins:
  - name: my_provider
    enabled: true
    path: /custom/path/my_provider.exe
    config:
      api_key_env: MY_API_KEY
      base_url: https://api.my-provider.com/v1
    priority: 100
```

**Workflow:**
```bash
# 1. Edit providers.yaml để add plugin config
# 2. Reload config
curl -X POST http://localhost:1807/api/config/reload \
  -H "Authorization: Bearer sk-jarvis-dev"
```

**Advantages:**
- ✅ Flexible - custom paths, configs, priorities
- ✅ Selective - enable/disable specific plugins
- ✅ Advanced - plugin-specific configuration
- ✅ Explicit - rõ ràng哪些 plugins được load

---

## Lessons Learned

1. **Fallback Pattern** - Auto-discovery là fallback khi không có config
2. **Cross-Platform** - Extension detection khác nhau giữa Windows và Unix
3. **Error Tolerance** - Continue loading nếu một plugin fail
4. **Logging** - Warning messages cho failed plugins
5. **Flexibility** - Support cả auto-discovery và config-based modes

---

## Files Modified

**Created:**
- `router/layers/provider/plugin/discovery.go`

**Modified:**
- `router/layers/provider/plugin/bootstrap.go` - Updated loadPluginsFromConfig()
- `router/configs/providers.yaml` - Commented out plugins section (enable auto-discovery)

---

## References

- Previous implementation: `Z:\Ti\Ti-learning-lab\03_Knowledge\patterns\plugin-system-implementation.md`

---

**Status:** ✅ Auto-discovery feature hoàn tất, enable zero-config plugin loading
