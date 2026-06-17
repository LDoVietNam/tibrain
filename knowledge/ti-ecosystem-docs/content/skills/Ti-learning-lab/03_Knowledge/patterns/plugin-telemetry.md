# Plugin-Based Telemetry Capture

> **Version**: 1.0.0  
> **Last Updated**: 2026-04-28  
> **Category**: Extensibility  
> **Language**: Tiếng Việt

---

## 📋 Tổng Quan

Plugin-Based Telemetry Capture là hệ thống telemetry có thể mở rộng theo RuFlo pattern, cho phép dynamic plugin loading và lifecycle management.

## 🎯 Mục Tiêu

1. **Extensibility** - Thêm telemetry plugins mới mà không cần modify core
2. **Dynamic Loading** - Load plugins runtime
3. **Lifecycle Management** - Quản lý plugin lifecycle
4. **Plugin Registry** - Central registry cho plugins

## 🏗️ Architecture

```
PluginTelemetry
├── Plugin Manager
│   ├── Load Plugins
│   ├── Register Plugins
│   └── Unregister Plugins
├── Plugin Registry
│   ├── Built-in Plugins
│   └── Custom Plugins
├── Capture Queue
│   ├── Worker Pool
│   └── Request Processing
└── Plugin Interface
    ├── Initialize
    ├── Capture
    └── Close
```

## 🔌 Plugin Interface

### TelemetryPlugin Interface

```go
type TelemetryPlugin interface {
    Name() string
    Version() string
    Initialize(config map[string]interface{}) error
    Capture(ctx context.Context, data *CaptureData) error
    Close() error
}
```

### Methods

| Method | Description |
|--------|-------------|
| Name() | Trả về plugin name |
| Version() | Trả về plugin version |
| Initialize() | Initialize plugin với config |
| Capture() | Capture telemetry data |
| Close() | Cleanup plugin resources |

## 📦 Built-in Plugins

### FileLogPlugin

Log telemetry data ra file:

```go
plugin := NewFileLogPlugin()
plugin.Initialize(map[string]interface{}{
    "file_path": "telemetry.log",
})
```

### MetricsPlugin

Capture metrics:

```go
plugin := NewMetricsPlugin()
plugin.Initialize(map[string]interface{}{})
```

## 🚀 Plugin Development

### Step 1: Implement Interface

```go
type MyPlugin struct {
    *BasePlugin
    // Custom fields
}

func (mp *MyPlugin) Name() string {
    return "my_plugin"
}

func (mp *MyPlugin) Version() string {
    return "1.0.0"
}

func (mp *MyPlugin) Initialize(config map[string]interface{}) error {
    // Custom initialization
    return nil
}

func (mp *MyPlugin) Capture(ctx context.Context, data *CaptureData) error {
    // Custom capture logic
    return nil
}

func (mp *MyPlugin) Close() error {
    // Custom cleanup
    return nil
}
```

### Step 2: Export Plugin

```go
// Export New function for dynamic loading
func New() (TelemetryPlugin, error) {
    return &MyPlugin{
        BasePlugin: NewBasePlugin("my_plugin", "1.0.0", nil),
    }, nil
}
```

### Step 3: Build Plugin

```bash
# Build as shared library
go build -buildmode=plugin -o my_plugin.so
```

### Step 4: Load Plugin

```go
pt := NewPluginTelemetry("/path/to/plugins", 4)

// Load plugins from directory
err := pt.LoadPlugins(ctx)
if err != nil {
    // Handle error
}
```

## 📝 Plugin Registry

### Register Plugin Factory

```go
RegisterPluginFactory("my_plugin", func() (TelemetryPlugin, error) {
    return NewMyPlugin()
})
```

### Create Plugin from Registry

```go
plugin, err := CreatePlugin("my_plugin")
if err != nil {
    // Handle error
}
```

### Get Registered Plugins

```go
plugins := GetRegisteredPlugins()
```

## 🔄 Capture Workflow

### 1. Create Plugin Telemetry

```go
pt := NewPluginTelemetry("/path/to/plugins", 4)
```

### 2. Load Plugins

```go
err := pt.LoadPlugins(ctx)
if err != nil {
    // Handle error
}
```

### 3. Capture Data

```go
data := &CaptureData{
    Timestamp: time.Now().Unix(),
    EventType: "request_complete",
    Component: "router",
    RequestID: "req-123",
    Provider: "openai",
    Model: "gpt-4",
    Metrics: map[string]interface{}{
        "latency_ms": 100,
        "tokens": 1000,
    },
}

err := pt.Capture(ctx, data)
if err != nil {
    // Handle error
}
```

### 4. Close Plugin Telemetry

```go
err := pt.Close()
if err != nil {
    // Handle error
}
```

## ⚙️ Configuration

### Worker Count

```go
pt := NewPluginTelemetry(pluginDir, workerCount)
```

- `workerCount`: Số worker goroutines cho capture queue
- Tăng worker count cho high throughput

### Plugin Directory

```go
pt := NewPluginTelemetry(pluginDir, workerCount)
```

- `pluginDir`: Directory chứa plugin files (.so)
- Plugins được load từ directory này

### Enable/Disable

```go
pt.Enable()  // Enable plugin telemetry
pt.Disable() // Disable plugin telemetry
```

## 🎓 Best Practices

### Plugin Development

1. **Implement all interface methods** - Đảm bảo tất cả methods được implement
2. **Handle errors gracefully** - Graceful error handling
3. **Cleanup resources in Close()** - Đảm bảo resources được cleanup
4. **Use versioning** - Version plugins để track changes

### Plugin Loading

1. **Validate plugins** - Validate plugin interface trước khi load
2. **Handle load failures** - Graceful handling khi plugin load fail
3. **Use plugin registry** - Register plugins trong registry
4. **Monitor plugin health** - Theo dõi plugin health

### Capture Performance

1. **Use async operations** - Async operations để tránh blocking
2. **Batch capture requests** - Batch requests để giảm overhead
3. **Use worker pool** - Worker pool cho concurrent processing
4. **Monitor queue depth** - Theo dõi queue depth để detect backlog

### Plugin Lifecycle

1. **Initialize properly** - Đảm bảo initialization complete
2. **Handle re-initialization** - Support re-initialization
3. **Graceful shutdown** - Graceful shutdown với timeout
4. **Resource cleanup** - Cleanup tất cả resources

## 🔍 Troubleshooting

### Plugin Loading Issues

**Issue**: Plugin không load được
- **Solution**: Check plugin path, verify plugin interface, check dependencies

**Issue**: Plugin initialization fails
- **Solution**: Check config, validate config schema, handle missing config

**Issue**: Plugin version conflict
- **Solution**: Use versioning, implement version compatibility, deprecate old versions

### Capture Issues

**Issue**: Capture latency quá cao
- **Solution**: Tăng worker count, use async operations, batch requests

**Issue**: Queue backlog
- **Solution**: Tăng worker count, reduce capture frequency, add queue monitoring

**Issue**: Data loss
- **Solution**: Increase queue size, add persistence, implement retry logic

### Plugin Issues

**Issue**: Plugin crash
- **Solution**: Add error handling, implement recovery, monitor plugin health

**Issue**: Plugin memory leak
- **Solution**: Profile memory usage, cleanup resources, implement limits

**Issue**: Plugin deadlock
- **Solution**: Avoid blocking operations, use timeouts, implement watchdog

## 📚 References

- Go Plugins: https://pkg.go.dev/plugin
- Plugin Pattern: https://en.wikipedia.org/wiki/Plugin_(computing)
- RuFlo Patterns: https://github.com/RuFlo/patterns
- Extensibility Patterns: https://martinfowler.com/articles/patterns-of-distributed-systems/

---

*Last Updated: 2026-04-28*
