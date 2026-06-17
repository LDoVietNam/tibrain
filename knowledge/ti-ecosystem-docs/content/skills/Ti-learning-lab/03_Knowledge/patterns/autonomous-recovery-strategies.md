---
tags: ["tibrain", "documentation", "skill", "retry", "router"]
scopes: ["resilience", "tibrain"]
last_updated: 2026-05-22
---
# Autonomous Recovery Strategies

> **Version**: 1.0.0  
> **Last Updated**: 2026-04-28  
> **Category**: Resilience  
> **Language**: Tiếng Việt

---

## 📋 Tổng Quan

Autonomous Recovery Strategies là hệ thống phục hồi tự động với các advanced patterns để xử lý failures một cách thông minh và hiệu quả.

## 🎯 Mục Tiêu

1. **Circuit Breaker** - Ngăn chặn cascade failures
2. **Health-Based Routing** - Route đến healthy providers
3. **Graceful Degradation** - Degrade gracefully khi overload
4. **Multi-Strategy Recovery** - Kết hợp nhiều strategies

## 🏗️ Architecture

```
AutonomousRecovery
├── Circuit Breakers
│   ├── Closed State (normal)
│   ├── Open State (failing)
│   └── Half-Open State (testing)
├── Health-Based Routing
│   ├── Provider Health Tracking
│   ├── Health Check Interval
│   └── Routing Table Updates
├── Graceful Degradation
│   ├── Degradation Levels
│   ├── Escalation Logic
│   └── De-escalation Logic
└── Multi-Strategy Recovery
    ├── Strategy Selection
    ├── Sequential Execution
    └── Parallel Execution
```

## 🔌 Circuit Breaker Pattern

### Mục đích

Ngăn chặn cascade failures bằng cách ngắn mạch các components đang fail.

### States

**Closed State** (Normal)
- Tất cả requests đi qua
- Theo dõi failure count
- Nếu failure threshold đạt → Open circuit

**Open State** (Failing)
- Tất cả requests bị reject
- Sau timeout → Half-open state
- Prevent cascade failures

**Half-Open State** (Testing)
- Cho phép một số requests đi qua
- Nếu success threshold đạt → Closed state
- Nếu fail → Open state

### Configuration

```go
circuitBreaker := &CircuitBreaker{
    Name: "router_http",
    State: CircuitStateClosed,
    FailureThreshold: 5,
    SuccessThreshold: 5,
    Timeout: time.Minute,
}
```

### Usage

```go
// Check if request is allowed
if ar.allowRequest(circuitBreaker) {
    // Execute request
    if success {
        ar.recordCircuitBreakerSuccess(circuitBreaker)
    } else {
        ar.recordCircuitBreakerFailure(circuitBreaker)
    }
} else {
    // Circuit is open, reject request
    return ErrCircuitOpen
}
```

## 🏥 Health-Based Routing

### Mục đích

Route requests đến healthy providers dựa trên health metrics.

### Health Metrics

**Success Rate**
- Tính bằng EMA (Exponential Moving Average)
- Alpha = 0.3
- Thresholds: Healthy (>0.95), Degraded (>0.85), Unhealthy (<0.7)

**Average Latency**
- Theo dõi latency trung bình
- Thresholds: Healthy (<1000ms), Degraded (<2000ms), Unhealthy (>2000ms)

**Consecutive Failures**
- Theo dõi số failures liên tiếp
- Threshold: 3 failures → Unhealthy

### Health Status

| Status | Success Rate | Latency | Consecutive Failures |
|--------|-------------|---------|-------------------|
| Healthy | > 0.95 | < 1000ms | 0 |
| Degraded | > 0.85 | < 2000ms | < 3 |
| Unhealthy | < 0.7 | > 2000ms | ≥ 3 |

### Usage

```go
// Update provider health
ar.updateProviderHealth(hbr, provider, success)

// Get healthy provider
healthyProvider := ar.getHealthyProvider(hbr)

// Switch to healthy provider
if healthyProvider != currentProvider {
    context.Provider = healthyProvider
}
```

## 📉 Graceful Degradation

### Mục đích

Degrade functionality gracefully khi system overload thay vì fail hoàn toàn.

### Degradation Levels

**Level 0: Full Functionality**
- Tất cả features enabled
- Trigger: Error rate < 1%, Latency < 1000ms, Success rate > 99%

**Level 1: Reduced Features**
- Non-critical features disabled
- Trigger: Error rate < 5%, Latency < 2000ms, Success rate > 95%

**Level 2: Minimal Mode**
- Chỉ core features enabled
- Trigger: Error rate < 10%, Latency < 5000ms, Success rate > 90%

**Level 3: Fallback Mode**
- Fallback đến default provider
- Trigger: Error rate < 20%, Latency < 10000ms, Success rate > 80%

### Escalation Logic

```go
if errorRate > currentLevel.Trigger.ErrorRate ||
   latency > currentLevel.Trigger.LatencyThreshold ||
   successRate < currentLevel.Trigger.SuccessRate {
    // Escalate to next level
    if currentLevel < maxLevel {
        currentLevel++
    }
}
```

### De-escalation Logic

```go
if errorRate < previousLevel.Trigger.ErrorRate &&
   latency < previousLevel.Trigger.LatencyThreshold &&
   successRate > previousLevel.Trigger.SuccessRate {
    // De-escalate to previous level
    if currentLevel > 0 {
        currentLevel--
    }
}
```

## 🔄 Multi-Strategy Recovery

### Mục đích

Kết hợp nhiều recovery strategies để tăng success rate.

### Strategy Selection

Dựa trên context:
- Error type
- Component health
- Provider health
- System load

### Execution Modes

**Sequential Execution**
- Thử từng strategy một
- Dừng khi một strategy thành công
- Thứ tự: Retry → Switch → Scale → Restart → Rollback

**Parallel Execution**
- Thử tất cả strategies cùng lúc
- Chấp nhận kết quả đầu tiên
- Tăng latency nhưng tăng success rate

### Usage

```go
// Execute multi-strategy recovery
result, err := ar.executeMultiStrategy(ctx, context)

// Configure execution mode
ar.multiStrategyRecovery.ParallelExecution = false
```

## 📊 Recovery Strategies

### Built-in Strategies

| Strategy | Type | Success Rate | Use Case |
|----------|------|-------------|----------|
| rate_limit_exceeded | Switch | 0.9 | Rate limit errors |
| timeout | Retry | 0.8 | Timeout errors |
| provider_unavailable | Switch | 0.85 | Provider down |
| high_latency | Switch | 0.75 | High latency |
| quality_degradation | Switch | 0.7 | Low quality |
| generic_error | Retry | 0.6 | Unknown errors |
| circuit_breaker | Circuit Breaker | 0.9 | Cascade failures |
| health_based_routing | Health-Based Routing | 0.85 | Provider health |
| graceful_degradation | Graceful Degradation | 0.8 | System overload |
| multi_strategy | Multi-Strategy | 0.95 | Complex failures |

### Custom Strategies

```go
ar.AddRecoveryStrategy("custom_error", &RecoveryStrategy{
    Name: "Custom Error Handling",
    Type: RecoveryTypeCustom,
    Action: "custom_action",
    Parameters: map[string]interface{}{
        "param1": "value1",
    },
    SuccessRate: 0.8,
})
```

## 🚀 Usage

### Basic Usage

```go
// Handle failure
result, err := ar.HandleFailure(ctx, errorType, context)

// Check circuit breaker state
state := ar.GetCircuitBreakerState("router_http")

// Get degradation level
level := ar.GetDegradationLevel()
```

### Configuration

```go
ar := NewAutonomousRecovery()

// Configure retry
ar.SetMaxRetries(3)
ar.SetRetryBackoff(time.Second)

// Configure auto-rollback
ar.SetAutoRollback(true)
```

## 🎓 Best Practices

### Circuit Breaker

1. **Set appropriate thresholds** - Không quá thấp (false positives) hoặc quá cao (late detection)
2. **Use exponential backoff** - Tăng timeout sau mỗi failure
3. **Monitor circuit states** - Theo dõi tần suất open/close circuits
4. **Test half-open logic** - Đảm bảo half-open state hoạt động đúng

### Health-Based Routing

1. **Use EMA for smoothing** - Smooth out fluctuations
2. **Set appropriate thresholds** - Balance sensitivity and stability
3. **Regular health checks** - Check health at regular intervals
4. **Fallback mechanism** - Always have fallback provider

### Graceful Degradation

1. **Define clear levels** - Mỗi level có actions rõ ràng
2. **Auto-escalate/de-escalate** - Tự động điều chỉnh dựa trên metrics
3. **Monitor degradation** - Theo dõi tần suất degradation
4. **Test escalation logic** - Đảm bảo escalation/de-escalation hoạt động đúng

### Multi-Strategy Recovery

1. **Order strategies carefully** - Thử strategies có success rate cao trước
2. **Use parallel for critical paths** - Parallel execution cho critical requests
3. **Monitor strategy performance** - Theo dõi success rate của từng strategy
4. **Fallback to simple** - Fallback to simple strategies nếu advanced fail

## 🔍 Troubleshooting

### Circuit Breaker Issues

**Issue**: Circuit mở quá thường xuyên
- **Solution**: Tăng failure threshold, giảm timeout

**Issue**: Circuit không đóng lại
- **Solution**: Giảm success threshold, tăng timeout

**Issue**: False positives
- **Solution**: Tăng failure threshold, thêm hysteresis

### Health-Based Routing Issues

**Issue**: Provider switching quá thường xuyên
- **Solution**: Tăng thresholds, sử dụng EMA với alpha thấp hơn

**Issue**: Không tìm thấy healthy provider
- **Solution**: Thêm fallback provider, giảm thresholds

**Issue**: Health metrics không chính xác
- **Solution**: Tăng health check interval, thu thập nhiều data hơn

### Graceful Degradation Issues

**Issue**: Degradation quá thường xuyên
- **Solution**: Tăng thresholds, thêm hysteresis

**Issue**: Không de-escalate
- **Solution**: Giảm thresholds, kiểm tra metrics accuracy

**Issue**: Degradation levels không rõ ràng
- **Solution**: Định nghĩa actions rõ ràng cho từng level

## 📚 References

- Circuit Breaker Pattern: https://martinfowler.com/bliki/CircuitBreaker.html
- Health-Based Routing: https://cloud.google.com/architecture/routing
- Graceful Degradation: https://en.wikipedia.org/wiki/Fault_tolerance
- RuFlo Patterns: https://github.com/RuFlo/patterns

---

*Last Updated: 2026-04-28*
