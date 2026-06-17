---
tags: ["tibrain", "documentation", "skill", "router", "api"]
scopes: ["tibrain"]
last_updated: 2026-05-22
---
# Config Watcher Implementation - Ti Router

> **Ngày tạo**: 2026-04-29  
> **Ngôn ngữ**: Tiếng Việt  
> **Mục đích**: Tài liệu hóa kiến thức về implement config watcher cho auto-reload config

## Tổng quan

Config watcher là feature cho phép Ti Router tự động reload config khi file config thay đổi, không cần restart server. Feature này sử dụng thư viện `fsnotify` để watch file system events.

## Vấn đề ban đầu

Trước khi implement config watcher:
- Khi thay đổi config (providers.yaml, Tiserverrouter.yaml), phải restart router thủ công
- Downtime trong quá trình restart
- Không phù hợp cho production environment

## Giải pháp

### 1. Thư viện fsnotify

Sử dụng `fsnotify` - thư viện cross-platform file watcher cho Go:

```go
import "github.com/fsnotify/fsnotify"
```

**Tại sao chọn fsnotify:**
- Cross-platform (Windows, Linux, macOS)
- API đơn giản, dễ sử dụng
- Active maintained (GitHub: ~7k stars)
- Tích hợp tốt với Go ecosystem

### 2. Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Router (Main)                          │
├─────────────────────────────────────────────────────────────┤
│  Init()                                                      │
│  ├── Load config watcher config from Tiserverrouter.yaml    │
│  ├── Check if config_watcher.enabled = true                 │
│  ├── Create ConfigWatcher with debounce                     │
│  └── Start watching providers.yaml                          │
├─────────────────────────────────────────────────────────────┤
│  reloadConfigCallback()                                      │
│  ├── Reload provider configs from YAML                     │
│  ├── Rebuild provider registry                              │
│  ├── Update routing router                                  │
│  ├── Update rate limiter                                    │
│  └── Update load balancer                                   │
├─────────────────────────────────────────────────────────────┤
│  Shutdown()                                                  │
│  └── Stop config watcher                                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   ConfigWatcher                              │
├─────────────────────────────────────────────────────────────┤
│  - filePath: string (path to config file)                   │
│  - debounce: time.Duration (debounce delay)                 │
│  - callback: func() error (reload callback)                  │
│  - watcher: *fsnotify.Watcher                               │
│  - timer: *time.Timer (debounce timer)                       │
├─────────────────────────────────────────────────────────────┤
│  Start() - Start watching file                              │
│  Stop() - Stop watching file                                │
│  handleEvent() - Handle fsnotify events                     │
│  triggerCallback() - Trigger reload with debounce            │
└─────────────────────────────────────────────────────────────┘
```

### 3. Debounce Logic

**Tại sao cần debounce?**
- File system có thể trigger nhiều events cho một lần save (WRITE, CHMOD, v.v.)
- Debounce đảm bảo chỉ reload một lần sau khi file ổn định
- Tránh reload quá nhiều lần gây performance issue

**Implement:**

```go
func (w *ConfigWatcher) handleEvent(event fsnotify.Event) {
    // Chỉ xử lý WRITE và CREATE events
    if event.Op&fsnotify.Write != fsnotify.Write &&
        event.Op&fsnotify.Create != fsnotify.Create {
        return
    }

    // Reset timer nếu có event mới (debounce)
    if w.timer != nil {
        w.timer.Stop()
    }

    // Set timer mới
    w.timer = time.AfterFunc(w.debounce, w.triggerCallback)
}
```

### 4. Config Structure

Trong `Tiserverrouter.yaml`:

```yaml
server:
  rate-limit:
    default-rpm: 100
    default-burst: 100
config_watcher:
  enabled: true
  debounce: 2s
```

**Config options:**
- `enabled`: Bật/tắt config watcher (default: false)
- `debounce`: Thời gian debounce (default: 2s)

### 5. Reload Callback

Callback được gọi khi config file thay đổi:

```go
func (r *Router) reloadConfigCallback() error {
    // 1. Reload provider configs
    newProviderConfigs, _, err := providers.LoadFromYAML(providerConfigsPath)
    
    // 2. Rebuild registry
    newRegistry, issues := providers.BuildRegistryFromConfig(nil, nil)
    
    // 3. Update global registry
    r.providerRegistry = newRegistry
    
    // 4. Update routing router
    r.routingRouter = routing.NewRouterWithRegistry(r.providerRegistry)
    
    // 5. Update rate limiter
    for name := range newProviderConfigs {
        r.rateLimiter.AddProvider(name, r.rateLimitConfig.DefaultRPM, r.rateLimitConfig.DefaultBurst)
    }
    
    // 6. Update load balancer
    r.loadBalancer = routing.NewLoadBalancer(providerList)
    
    return nil
}
```

## Implementation Steps

### Step 1: Tạo ConfigWatcher

File: `Z:\Ti\router\layers\config\watcher.go`

```go
package config

import (
    "log"
    "time"
    
    "github.com/fsnotify/fsnotify"
)

type ConfigWatcher struct {
    filePath string
    debounce time.Duration
    callback func() error
    watcher  *fsnotify.Watcher
    timer    *time.Timer
}

func NewConfigWatcher(filePath string, debounce time.Duration, callback func() error) (*ConfigWatcher, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }
    
    return &ConfigWatcher{
        filePath: filePath,
        debounce: debounce,
        callback: callback,
        watcher:  watcher,
    }, nil
}

func (w *ConfigWatcher) Start() error {
    if err := w.watcher.Add(w.filePath); err != nil {
        return err
    }
    
    go w.watch()
    return nil
}

func (w *ConfigWatcher) Stop() {
    w.watcher.Close()
    if w.timer != nil {
        w.timer.Stop()
    }
}

func (w *ConfigWatcher) watch() {
    for {
        select {
        case event, ok := <-w.watcher.Events:
            if !ok {
                return
            }
            w.handleEvent(event)
        case err, ok := <-w.watcher.Errors:
            if !ok {
                return
            }
            log.Printf("[ConfigWatcher] Error: %v", err)
        }
    }
}

func (w *ConfigWatcher) handleEvent(event fsnotify.Event) {
    if event.Op&fsnotify.Write != fsnotify.Write &&
        event.Op&fsnotify.Create != fsnotify.Create {
        return
    }
    
    if w.timer != nil {
        w.timer.Stop()
    }
    
    w.timer = time.AfterFunc(w.debounce, w.triggerCallback)
}

func (w *ConfigWatcher) triggerCallback() {
    log.Printf("[ConfigWatcher] Config file changed, reloading...")
    if err := w.callback(); err != nil {
        log.Printf("[ConfigWatcher] Reload failed: %v", err)
    } else {
        log.Printf("[ConfigWatcher] Config reloaded successfully")
    }
}
```

### Step 2: Integrate vào Router

File: `Z:\Ti\router\cmd\routerd\router.go`

**Add field vào Router struct:**

```go
type Router struct {
    // ... existing fields
    configWatcher *configwatcher.ConfigWatcher
}
```

**Add initialization trong Init():**

```go
func (r *Router) Init(configPath string) error {
    // ... existing initialization
    
    // Initialize config watcher (optional - only if enabled in config)
    if err := r.initConfigWatcher(providerConfigsPath); err != nil {
        log.Printf("[Router] Failed to initialize config watcher: %v (continuing without auto-reload)", err)
    }
    
    return nil
}
```

**Add initConfigWatcher method:**

```go
func (r *Router) initConfigWatcher(configPath string) error {
    // Read config to check if config watcher is enabled
    data, err := os.ReadFile("configs/Tiserverrouter.yaml")
    if err != nil {
        return fmt.Errorf("read config file: %w", err)
    }
    
    var config struct {
        ConfigWatcher struct {
            Enabled  bool          `yaml:"enabled"`
            Debounce time.Duration `yaml:"debounce"`
        } `yaml:"config_watcher"`
    }
    if err := yaml.Unmarshal(data, &config); err != nil {
        return fmt.Errorf("parse config: %w", err)
    }
    
    // If not enabled, return without error
    if !config.ConfigWatcher.Enabled {
        log.Printf("[Router] Config watcher disabled")
        return nil
    }
    
    // Set default debounce if not specified
    debounce := config.ConfigWatcher.Debounce
    if debounce == 0 {
        debounce = 2 * time.Second // Default 2 seconds
    }
    
    // Create config watcher
    r.configWatcher, err = configwatcher.NewConfigWatcher(configPath, debounce, r.reloadConfigCallback)
    if err != nil {
        return fmt.Errorf("create config watcher: %w", err)
    }
    
    // Start watching
    if err := r.configWatcher.Start(); err != nil {
        return fmt.Errorf("start config watcher: %w", err)
    }
    
    log.Printf("[Router] Config watcher enabled (debounce: %v)", debounce)
    return nil
}
```

**Add reload callback:**

```go
func (r *Router) reloadConfigCallback() error {
    // Reload provider configs
    providerConfigsPath := "configs/providers.yaml"
    newProviderConfigs, _, err := providers.LoadFromYAML(providerConfigsPath)
    if err != nil {
        return fmt.Errorf("failed to load providers.yaml: %w", err)
    }
    
    // Rebuild registry with new configs
    newRegistry, issues := providers.BuildRegistryFromConfig(nil, nil)
    for _, issue := range issues {
        log.Printf("[Config Reload] Provider %s: %s", issue.Provider, issue.Message)
    }
    
    // Update global registry
    r.providerRegistry = newRegistry
    
    // Update provider configs
    r.providerConfigs = newProviderConfigs
    
    // Update routing router with new registry
    r.routingRouter = routing.NewRouterWithRegistry(r.providerRegistry)
    
    // Update rate limiter with new configs
    for name := range newProviderConfigs {
        r.rateLimiter.AddProvider(name, r.rateLimitConfig.DefaultRPM, r.rateLimitConfig.DefaultBurst)
    }
    
    // Update load balancer with new provider list
    providerList := make([]string, 0, len(newProviderConfigs))
    for name := range newProviderConfigs {
        providerList = append(providerList, name)
    }
    r.loadBalancer = routing.NewLoadBalancer(providerList)
    
    log.Printf("[Config Reload] Configuration reloaded successfully. Providers: %d", len(newProviderConfigs))
    return nil
}
```

**Add cleanup trong Shutdown():**

```go
func (r *Router) Shutdown(ctx context.Context) error {
    // ... existing cleanup
    
    // Stop config watcher if exists
    if r.configWatcher != nil {
        log.Println("[Router] Stopping config watcher...")
        r.configWatcher.Stop()
        log.Println("[Router] Config watcher stopped")
    }
    
    // ... rest of shutdown
}
```

### Step 3: Update Config File

File: `Z:\Ti\router\configs\Tiserverrouter.yaml`

```yaml
server:
  rate-limit:
    default-rpm: 100
    default-burst: 100
config_watcher:
  enabled: true
  debounce: 2s
```

### Step 4: Add Dependency

File: `Z:\Ti\router\go.mod`

```go
require (
    // ... existing dependencies
    github.com/fsnotify/fsnotify v1.7.0
)
```

## Testing

### Test 1: Build Success

```bash
cd Z:\Ti\router
go build -o bin/routerd.exe ./cmd/routerd
```

**Kết quả:** ✅ Build thành công

### Test 2: Start Router

```bash
./bin/routerd.exe -config configs/Tiserverrouter.yaml
```

**Kết quả:** ✅ Router start thành công với config watcher enabled

**Log output:**
```
2026/04/29 07:46:47 [ConfigWatcher] Watching config file: configs/providers.yaml
2026/04/29 07:46:47 [Router] Config watcher enabled (debounce: 2s)
```

## Lessons Learned

### 1. Import Cycle Issue

**Vấn đề:** Khi add plugin system, xảy ra import cycle:
- `providers` imports `plugin`
- `plugin` imports `providers` (via adapter)

**Giải pháp:** Tạm thời remove plugin code từ bootstrap.go vì config watcher không cần plugin system. Plugin system có thể add lại sau khi refactor architecture.

### 2. Debounce là quan trọng

**Vấn đề:** File system trigger nhiều events cho một lần save
**Giải pháp:** Implement debounce với timer reset

### 3. Graceful Degradation

**Vấn đề:** Config watcher fail không nên crash router
**Giải pháp:** Log error và continue without auto-reload

### 4. Config Option

**Vấn đề:** Cần có cách enable/disable feature
**Giải pháp:** Add config_watcher section trong Tiserverrouter.yaml

## Future Improvements

1. **Watch Multiple Files:** Hiện tại chỉ watch providers.yaml, có thể extend để watch Tiserverrouter.yaml
2. **Hot Reload Plugins:** Khi plugin system được fix, có thể add hot reload cho plugins
3. **Validation Before Reload:** Validate config trước khi apply, rollback nếu invalid
4. **Metrics:** Add metrics cho config reload events (success/failure/count)
5. **Notification:** Send notification khi config reload (Slack, email, v.v.)

## References

- **fsnotify GitHub:** https://github.com/fsnotify/fsnotify
- **File Watcher Patterns:** https://github.com/gorilla/websocket/blob/master/examples/filewatch
- **Debounce Pattern:** https://github.com/benbjohnson/clock (Timer pattern)

## Related Documents

- [Config Reload Mechanism](./config-reload-mechanism.md)
- [Plugin System Implementation](./plugin-system-implementation.md)
- [Plugin Auto-Discovery](./plugin-auto-discovery.md)
