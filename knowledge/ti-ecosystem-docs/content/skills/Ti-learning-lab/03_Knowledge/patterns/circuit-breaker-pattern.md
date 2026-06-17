---
tags: ["retry", "tibrain", "pattern-retry-circuit-breaker", "documentation", "skill"]
scopes: ["resilience", "tibrain"]
last_updated: 2026-05-22
---
# Circuit Breaker Pattern

> **Category**: Resilience & Fault Tolerance  
> **Language**: Tiếng Việt  
> **Last Updated**: 2026-04-30

---

## Tổng Quan

Circuit breaker là pattern ngăn requests đến một service đang fail. Giống như circuit breaker điện - khi có quá nhiều failures, "circuit" mở để ngăn requests, cho phép service recover.

---

## Khi Nào Sử Dụng

- Call external services/APIs
- Prevent cascading failures
- Protect against overloaded services
- Reduce load cho failing services

---

## Cấu Trúc

### Circuit Breaker States

**1. Closed (Đóng) - Normal State**
- Requests pass through normally
- Track successes và failures
- Nếu failure rate > threshold → Open circuit

**2. Open (Mở) - Failing State**
- Block tất cả requests
- Return error hoặc fallback immediately
- Sau timeout → Half-Open

**3. Half-Open (Nửa Mở) - Recovery State**
- Allow limited requests (e.g., 1 request)
- Nếu success → Closed
- Nếu failure → Open

### Health Score

```go
HealthScore = 0.7 * SuccessRate + 0.3 * (1 - ErrorRate)
```

- Success rate: % of successful requests
- Error rate: % of failed requests
- Health score: 0.0 (bad) đến 1.0 (good)

---

## Implementation Pattern

### 1. Create Circuit Breaker

```go
cb := NewCircuitBreaker(
    "provider_name",
    failureThreshold,    // e.g., 5 failures
    timeout,            // e.g., 30 seconds
)
```

### 2. Record Success/Failure

```go
if request successful:
    cb.RecordSuccess()
else:
    cb.RecordFailure()
```

### 3. Check Before Request

```go
if cb.Allow():
    // Make request
else:
    // Circuit is open, return error or fallback
```

### 4. Get Health Score

```go
health := cb.GetHealthScore()
if health < 0.5:
    // Service is unhealthy
```

---

## Best Practices

### 1. Thresholds
- **Failure threshold**: 3-10 failures (tùy criticality)
- **Timeout**: 30-60 seconds (tùy recovery time)
- **Health threshold**: 0.5-0.7 (50-70%)

### 2. Fallback Strategy
- Always có fallback (cached response, alternative service)
- Log circuit breaker events
- Alert khi circuit opens

### 3. Monitoring
- Track circuit state changes
- Track health scores over time
- Monitor how long circuits stay open

### 4. Adaptive Thresholds
- Adjust thresholds dựa on traffic patterns
- Higher thresholds cho critical services
- Lower thresholds cho non-critical services

### 5. Manual Override
- Allow manual open/close (cho maintenance)
- Reset circuit after maintenance
- Document manual overrides

---

## Common Mistakes

### 1. Too Sensitive
- **Problem**: Threshold quá thấp, circuit opens quá thường xuyên
- **Solution**: Tăng threshold, add hysteresis

### 2. No Fallback
- **Problem**: Không có fallback khi circuit open
- **Solution**: Always implement fallback

### 3. No Monitoring
- **Problem**: Không track circuit state
- **Solution**: Log state changes, set up alerts

### 4. Long Timeout
- **Problem**: Timeout quá dài, service không recover nhanh
- **Solution**: Reduce timeout, implement exponential backoff

### 5. Per-Request Circuit
- **Problem**: Circuit breaker per request thay vì per service
- **Solution**: Circuit breaker per service/endpoint

---

## Advanced Patterns

### 1. Adaptive Circuit Breaker
- Thresholds tự động adjust dựa on metrics
- Learn optimal thresholds từ historical data
- Useful cho variable workloads

### 2. Distributed Circuit Breaker
- Circuit state shared across instances
- Sử dụng Redis hoặc distributed cache
- Useful cho distributed systems

### 3. Multi-Level Circuit Breaker
- Circuit breakers ở multiple levels
- E.g., per-instance, per-region, per-cluster
- Useful cho complex architectures

### 4. Circuit Breaker with Metrics
- Track detailed metrics (latency, error types)
- Use metrics cho threshold tuning
- Integrate với observability platform

---

## Tools & Libraries

### Go Libraries
- **github.com/sony/gobreaker**: Popular circuit breaker library
- **github.com/streadway/amqp**: Circuit breaker cho RabbitMQ
- **github.com/afex/hystrix-go**: Hystrix port cho Go

### Java Libraries
- **Resilience4j**: Circuit breaker cho Java
- **Hystrix**: Netflix circuit breaker (deprecated)

### Python Libraries
- **pybreaker**: Circuit breaker cho Python
- **circuitbreaker**: Simple circuit breaker

---

## Comparison: Circuit Breaker vs Retry

| Aspect | Circuit Breaker | Retry |
|--------|-----------------|-------|
| Purpose | Prevent calls to failing service | Retry failed requests |
| When | When service is consistently failing | On transient failures |
| State | Stateful (Closed/Open/Half-Open) | Stateless |
| Use Case | Protect against cascading failures | Handle temporary failures |

---

## Ti Router Implementation

**File**: `layers/resilience/circuit_breaker.go`

**Features**:
- CircuitBreaker struct với health scoring
- Three states: Closed, Open, Half-Open
- Failure threshold configurable
- Timeout configurable
- RecordSuccess(), RecordFailure(), Allow() methods
- GetHealthScore() method
- Thread-safe với mutex

**Usage**:
```go
cb := NewCircuitBreaker("provider", 5, 30*time.Second)
if cb.Allow():
    // Make request
    if success:
        cb.RecordSuccess()
    else:
        cb.RecordFailure()
else:
    // Circuit open, use fallback
```

---

## References

- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Microsoft Circuit Breaker](https://docs.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker)
- [Resilience4j Documentation](https://resilience4j.readme.io/)

---

## Next Steps

1. Implement adaptive circuit breaker
2. Add distributed circuit breaker state
3. Implement multi-level circuit breaker
4. Add circuit breaker analytics
5. Integrate với Prometheus metrics
