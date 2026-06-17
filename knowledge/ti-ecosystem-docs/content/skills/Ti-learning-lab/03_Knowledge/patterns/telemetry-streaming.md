# Telemetry Streaming

> **Version**: 1.0.0  
> **Last Updated**: 2026-04-28  
> **Category**: Observability  
> **Language**: Tiếng Việt

---

## 📋 Tổng Quan

Telemetry Streaming là hệ thống thu thập và streaming telemetry data real-time theo Google race-condition pattern, cho phép monitoring và debugging hiệu quả.

## 🎯 Mục Tiêu

1. **Real-time streaming** - Streaming telemetry data real-time
2. **Event-based collection** - Thu thập dựa trên events
3. **Low-latency collection** - Thu thập với latency thấp
4. **Scalable architecture** - Architecture có thể scale

## 🏗️ Architecture

```
TelemetryStreaming
├── Streams
│   ├── Component Streams
│   ├── Event Streams
│   └── Metric Streams
├── Event Bus
│   ├── Channels
│   ├── Subscribers
│   └── Publishers
├── Collectors
│   ├── Console Collector
│   ├── Memory Collector
│   └── Custom Collectors
└── Buffer
    ├── Metrics Buffer
    ├── Event Buffer
    └── Flush Interval
```

## 📊 Telemetry Event

### Structure

```go
type TelemetryEvent struct {
    Timestamp   time.Time
    EventType   string
    Component   string
    RequestID   string
    Provider    string
    Model       string
    Metrics     map[string]interface{}
    Metadata    map[string]interface{}
}
```

### Event Types

| Event Type | Description |
|-----------|-------------|
| request_start | Request bắt đầu |
| request_complete | Request hoàn thành |
| request_error | Request error |
| model_switch | Model được switch |
| provider_switch | Provider được switch |
| circuit_breaker_open | Circuit breaker mở |
| circuit_breaker_close | Circuit breaker đóng |
| degradation_escalate | Degradation escalate |
| degradation_deescalate | Degradation de-escalate |

## 🌊 Streams

### Component Streams

Mỗi component có stream riêng:
- Router HTTP
- Brain Server
- Learning System
- Autonomous Recovery

### Event Streams

Stream theo event type:
- Request events
- Error events
- Recovery events
- Performance events

### Metric Streams

Stream theo metric type:
- Latency metrics
- Cost metrics
- Quality metrics
- Token metrics

## 🚌 Event Bus

### Channels

Mỗi component có channel riêng:
- `router`
- `brain_server`
- `learning_system`
- `autonomous_recovery`

### Subscribe

```go
// Subscribe to component events
ch := eventBus.Subscribe("router")

// Receive events
for event := range ch {
    // Process event
}
```

### Publish

```go
// Publish event to component
eventBus.Publish("router", telemetryEvent)
```

## 📦 Collectors

### Console Collector

In events ra console:

```go
collector := &ConsoleCollector{}
ts.AddCollector(collector)
```

### Memory Collector

Lưu events trong memory:

```go
collector := NewMemoryCollector(1000)
ts.AddCollector(collector)

// Get events
events := collector.GetEvents()
```

### Custom Collector

Tạo custom collector:

```go
type CustomCollector struct{}

func (cc *CustomCollector) Collect(ctx context.Context, event *TelemetryEvent) error {
    // Custom collection logic
    return nil
}

func (cc *CustomCollector) Flush(ctx context.Context) error {
    // Custom flush logic
    return nil
}

func (cc *CustomCollector) Close() error {
    // Custom close logic
    return nil
}
```

## 🔄 Workflow

### 1. Create Stream

```go
ts.CreateStream("router_http", "router")
```

### 2. Record Event

```go
event := &TelemetryEvent{
    EventType: "request_start",
    Component: "router",
    RequestID: "req-123",
    Provider: "openai",
    Model: "gpt-4",
    Metrics: map[string]interface{}{
        "latency_ms": 100,
        "tokens": 1000,
    },
}

ts.RecordEvent(ctx, event)
```

### 3. Auto Flush

```go
// Start auto flush
go ts.StartAutoFlush(ctx)
```

### 4. Manual Flush

```go
// Flush specific stream
stream, _ := ts.GetStream("router_http")
ts.flushStream(ctx, stream)

// Flush all streams
ts.FlushAll(ctx)
```

## ⚙️ Configuration

### Buffer Size

```go
ts := NewTelemetryStreaming(bufferSize, flushInterval)
```

- `bufferSize`: Số events tối đa trong buffer (default: 1000)
- `flushInterval`: Interval giữa các flush (default: 1 minute)

### Flush Strategy

**Auto Flush**
- Flush tự động tại intervals
- Tối ưu cho production

**Manual Flush**
- Flush thủ công khi cần
- Tối ưu cho debugging

**Hybrid**
- Kết hợp auto và manual
- Balance giữa performance và control

## 🚀 Usage

### Basic Setup

```go
// Create telemetry streaming
ts := NewTelemetryStreaming(1000, time.Minute)

// Create streams
ts.CreateStream("router_http", "router")
ts.CreateStream("brain_server", "brain_server")

// Add collectors
ts.AddCollector(&ConsoleCollector{})
ts.AddCollector(NewMemoryCollector(1000))

// Start auto flush
go ts.StartAutoFlush(ctx)
```

### Recording Events

```go
// Record request event
event := &TelemetryEvent{
    EventType: "request_complete",
    Component: "router",
    RequestID: "req-123",
    Provider: "openai",
    Model: "gpt-4",
    Metrics: map[string]interface{}{
        "latency_ms": 100,
        "tokens": 1000,
        "cost": 0.01,
        "quality_score": 0.9,
    },
}

ts.RecordEvent(ctx, event)
```

### Subscribing to Events

```go
// Subscribe to router events
ch := eventBus.Subscribe("router")

// Process events
for event := range ch {
    // Deserialize event
    telemetryEvent, err := FromJSONEvent([]byte(event))
    if err != nil {
        continue
    }
    
    // Process event
    fmt.Printf("Event: %s\n", telemetryEvent.EventType)
}
```

## 🎓 Best Practices

### Event Design

1. **Use structured events** - Events có cấu trúc rõ ràng
2. **Include metadata** - Thêm metadata cho context
3. **Use consistent naming** - Naming convention nhất quán
4. **Add timestamps** - Timestamps cho correlation

### Stream Management

1. **Separate streams by component** - Mỗi component có stream riêng
2. **Set appropriate buffer sizes** - Balance memory và performance
3. **Monitor buffer usage** - Theo dõi buffer overflow
4. **Flush regularly** - Flush thường xuyên để tránh data loss

### Collector Design

1. **Make collectors idempotent** - Collectors nên idempotent
2. **Handle errors gracefully** - Graceful error handling
3. **Support async operations** - Async operations cho performance
4. **Provide metrics** - Metrics cho collector health

### Performance

1. **Use batching** - Batch events để giảm overhead
2. **Use compression** - Compress data cho network transport
3. **Use async operations** - Async operations để tránh blocking
4. **Monitor latency** - Theo dõi latency của collection

## 🔍 Troubleshooting

### Buffer Overflow

**Issue**: Events bị drop do buffer đầy
- **Solution**: Tăng buffer size, giảm flush interval, add more collectors

### High Latency

**Issue**: Collection latency quá cao
- **Solution**: Use async operations, batching, compression

### Data Loss

**Issue**: Events bị mất
- **Solution**: Increase flush frequency, add persistence, use durable storage

### Memory Issues

**Issue**: Memory usage quá cao
- **Solution**: Giảm buffer size, increase flush frequency, use streaming collectors

## 📚 References

- Google Race-Condition Pattern: https://cloud.google.com/architecture/race-conditions
- Telemetry Best Practices: https://docs.microsoft.com/en-us/azure/telemetry
- Event-Driven Architecture: https://martinfowler.com/articles/201701-event-driven.html

---

*Last Updated: 2026-04-28*
