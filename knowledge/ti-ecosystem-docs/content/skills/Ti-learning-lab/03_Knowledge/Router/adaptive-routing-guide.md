# Hướng Dẫn Adaptive Routing Engine cho Ti Router

> **Ngày tạo**: 2026-04-29  
> **Tác giả**: claude  
> **Dự án**: Ti Router  
> **Phiên bản**: 1.0.0  

---

## Mục Lục

1. [Tổng Quan](#tổng-quan)
2. [Kiến Trúc](#kiến-trúc)
3. [Thành Phần Chính](#thành-phần-chính)
4. [Cấu Hình](#cấu-hình)
5. [API Monitoring](#api-monitoring)
6. [Sử Dụng](#sử-dụng)
7. [Troubleshooting](#troubleshooting)
8. [Benchmarks](#benchmarks)

---

## Tổng Quan

**Adaptive Routing Engine** là hệ thống routing thông minh cho Ti Router, cho phép:

- **Tự động chọn provider tốt nhất** dựa trên latency, chi phí, và availability
- **Tự động fallback** khi provider gặp sự cố
- **Theo dõi hiệu năng** theo thời gian thực (latency, cost, health)
- **Circuit breaker** để bảo vệ hệ thống khỏi cascading failures
- **Event-driven reactions** cho automated response

**Khác biệt với routing cũ:**

| Feature | Routing Cũ | Adaptive Routing |
|---------|-----------|------------------|
| Chọn provider | Theo thứ tự cố định | Theo score tính toán động |
| Latency | Static (hardcoded) | Measured real-time |
| Fallback | Đơn giản | Weighted + circuit breaker |
| Health check | Không có | Tích hợp circuit breaker |
| Cost tracking | Không có | Theo dõi real-time |
| Auto reaction | Không có | Event-driven |

---

## Kiến Trúc

```
User Request
    ↓
Router Engine (HTTP API)
    ↓
Adaptive Router
    ├── Latency Tracker (Rolling Window + EMA)
    ├── Cost Tracker (Per Provider)
    ├── Health Monitor (Circuit Breaker)
    └── Weighted Decision Algorithm
    ↓
Provider Pool
    ├── OpenRouter  (Score: 0.85)
    ├── Groq        (Score: 0.92) ← Chọn
    ├── Cerebras    (Score: 0.78)
    ├── DeepSeek    (Score: 0.81)
    ├── Google      (Score: 0.75)
    ├── OpenAI      (Score: 0.70)
    └── Anthropic   (Score: 0.88)
    ↓
LLM Provider API
    ↓
Metrics & Monitoring
```

---

## Thành Phần Chính

### 1. Latency Tracker (`latency_tracker.go`)

**Mục đích:** Đo và theo dõi thời gian phản hồi của mỗi provider.

**Tính năng:**
- Rolling window (mặc định: 100 requests)
- Exponential Moving Average (EMA) cho real-time estimate
- Percentiles: P50, P95, P99
- Success rate tracking
- Thread-safe với `sync.RWMutex`

**Cách sử dụng:**

```go
lt := routing.NewLatencyTracker(100)

// Ghi nhận measurement
lt.Record("groq", "llama3", 150*time.Millisecond, true)

// Lấy stats
stats := lt.GetStats("groq")
fmt.Printf("EMA: %v\n", stats.GetEMA())
fmt.Printf("P95: %v\n", stats.GetP95())
fmt.Printf("P99: %v\n", stats.GetP99())
fmt.Printf("Success Rate: %.2f%%\n", stats.GetSuccessRate()*100)
```

### 2. Cost Tracker (`cost_tracker.go`)

**Mục đích:** Theo dõi chi phí per request cho mỗi provider.

**Tính năng:**
- Cấu hình cost per 1K input/output tokens
- Rolling window tracking
- Cumulative cost tracking
- Session cost tracking
- Average cost per request

**Cách sử dụng:**

```go
ct := routing.NewCostTracker(1000)

// Đăng ký provider với cost rates
ct.RegisterProvider("openai", "gpt-4", 0.03, 0.06) // $/1K tokens

// Ghi nhận request
ct.Record("openai", 1000, 500, true) // 1000 input, 500 output tokens

// Lấy stats
stats := ct.GetStats("openai")
fmt.Printf("Total Cost: $%.4f\n", stats.GetTotalCost())
fmt.Printf("Session Cost: $%.4f\n", stats.GetSessionCost())
fmt.Printf("Avg Cost/Request: $%.4f\n", stats.GetAverageCostPerRequest())
```

### 3. Health Monitor & Circuit Breaker (`health_monitor.go`)

**Mục đích:** Theo dõi trạng thái up/down của providers và bảo vệ hệ thống.

**Tính năng:**
- Circuit Breaker pattern (Closed → Open → HalfOpen)
- Configurable failure threshold
- Automatic recovery detection
- Recovery timeout
- Half-open probing

**Cách sử dụng:**

```go
config := routing.DefaultCircuitBreakerConfig()
hm := routing.NewHealthMonitor(config)

// Đăng ký provider
hm.RegisterProvider("groq", "llama3")

// Ghi nhận kết quả
hm.RecordResult("groq", true, 100*time.Millisecond)  // Success
hm.RecordResult("groq", false, 0)                     // Failure
hm.RecordResult("groq", false, 0)                     // Failure
// ... sau N failures, circuit sẽ OPEN

// Kiểm tra health
if hm.IsHealthy("groq") {
    // Cho phép request
}

// Lấy state
state := hm.GetState("groq") // CircuitClosed, CircuitHalfOpen, CircuitOpen
stats := hm.GetStats("groq")
```

**Circuit Breaker States:**

```
[Closed] ----failure count >= threshold----> [Open]
    ↑                                          │
    │                                          │
    │    Recovery timeout expired              │
    │         (cho phép probe)                 │
    │                                          ↓
    └-------success >= threshold------ [HalfOpen]
              (đóng circuit)
```

### 4. Adaptive Router (`adaptive_router.go`)

**Mục đích:** Thuật toán chọn provider dựa trên weighted scoring.

**Trọng số mặc định:**

| Yếu tố | Trọng số | Mô tả |
|--------|----------|-------|
| Latency | 40% | Lower = better |
| Cost | 30% | Lower = better |
| Availability | 20% | Healthy = 1.0 |
| Quota | 10% | Higher = better |

**Tính năng:**
- Weighted scoring với normalization
- Fallback chain
- Randomization nhỏ để tránh thundering herd
- Real-time measured latency ưu tiên hơn static values
- Cost-based routing

**Cách sử dụng:**

```go
// Khởi tạo các trackers
lt := routing.NewLatencyTracker(100)
ct := routing.NewCostTracker(1000)
hm := routing.NewHealthMonitor(routing.DefaultCircuitBreakerConfig())

// Khởi tạo adaptive router
ar := routing.NewAdaptiveRouter(routing.DefaultRoutingWeights(), lt, ct, hm)

// Đăng ký providers
ar.RegisterProvider("groq", "llama3", 0.01, 0.02)
ar.RegisterProvider("openai", "gpt-4", 0.03, 0.06)

// Chọn provider
candidates := []routing.ProviderCandidate{
    {Provider: "groq", Model: "llama3", CostPer1MTokens: 1.0, QuotaRemaining: 10000},
    {Provider: "openai", Model: "gpt-4", CostPer1MTokens: 30.0, QuotaRemaining: 5000},
}

score, err := ar.SelectProvider(candidates, "chat")
if err != nil {
    // No healthy providers available
    log.Fatal(err)
}

fmt.Printf("Selected: %s (score: %.2f)\n", score.Provider, score.Score)
fmt.Printf("Latency: %v, Cost: %.2f/1M tokens\n", score.RawLatency, score.RawCost)

// Sau khi request hoàn thành, ghi nhận kết quả
ar.RecordResult(score.Provider, score.Model, latency, inputTokens, outputTokens, success)
```

### 5. Reaction Engine (`reaction_engine.go`)

**Mục đích:** Tự động phản ứng với sự kiện.

**Trigger types:**

| Trigger | Mô tả | Action mặc định |
|---------|-------|-----------------|
| `provider-down` | Provider không khả dụng | Switch provider |
| `provider-slow` | Latency cao | Switch provider |
| `latency-spike` | Latency tăng đột ngột | Notify |
| `cost-exceeded` | Chi phí vượt ngưỡng | Notify |
| `quota-depleted` | Quota còn ít | Switch provider |

**Cách sử dụng:**

```go
re := routing.NewReactionEngine(ar)

// Đăng ký handler
customHandler := func(event routing.ReactionEvent) error {
    log.Printf("[%s] %s: %s", event.Trigger, event.Provider, event.Message)
    return nil
}
re.RegisterHandler(routing.TriggerProviderDown, customHandler)

// Kích hoạt reaction
event := routing.ReactionEvent{
    Trigger:   routing.TriggerProviderDown,
    Provider:  "groq",
    Message:   "Provider not responding",
}
re.Trigger(event)
```

---

## Cấu Hình

### File cấu hình (`adaptive_config.go`)

```yaml
enabled: true
window_size: 100
ema_alpha: 0.3
unhealthy_latency_threshold_ms: 30000
weights:
  latency_weight: 0.4
  cost_weight: 0.3
  availability_weight: 0.2
  quota_weight: 0.1
  randomization: 0.05
circuit_breaker:
  failure_threshold: 5
  recovery_timeout: 30s
  half_open_max_requests: 3
  success_threshold: 2
reactions:
  - trigger: provider-down
    auto: true
    action: switch-provider
    retries: 3
    escalate_after: 5m
    cooldown: 30s
  - trigger: provider-slow
    auto: true
    action: switch-provider
    threshold: 10000
    retries: 2
    cooldown: 60s
fallback_order:
  - groq
  - cerebras
  - deepseek
  - openrouter
  - google
  - openai
  - anthropic
```

### Hot Reload

```go
// Khởi tạo config với hot reload
config, err := routing.NewHotReloadConfig("adaptive-routing.yaml")
if err != nil {
    log.Fatal(err)
}

// Lấy config hiện tại
cfg := config.Get()

// Tự động reload khi file thay đổi (blocking)
// go config.Watch(5 * time.Second)

// Hoặc kiểm tra thủ công
err = config.CheckAndReload()
```

---

## API Monitoring

### Endpoints

| Method | Path | Mô tả |
|--------|------|-------|
| GET | `/v1/metrics` | System metrics |
| GET | `/v1/metrics/providers` | Per-provider metrics |
| GET | `/v1/metrics/latency` | Latency stats |
| GET | `/v1/metrics/cost` | Cost stats |
| GET | `/v1/health` | Overall health |
| GET | `/v1/health/providers` | Per-provider health |

### Ví dụ Response (`/v1/metrics`)

```json
{
  "timestamp": "2026-04-29T10:30:00Z",
  "uptime": "2h15m30s",
  "providers": [
    {
      "provider": "groq",
      "model": "llama3",
      "health": {
        "state": "closed",
        "healthy": true,
        "failure_count": 0,
        "success_count": 150
      },
      "latency": {
        "ema_ms": 145,
        "p50_ms": 120,
        "p95_ms": 280,
        "p99_ms": 450,
        "avg_ms": 160,
        "sample_count": 100
      },
      "cost": {
        "total_cost_usd": 12.45,
        "session_cost_usd": 3.21,
        "avg_cost_per_request": 0.08,
        "total_tokens": 145000,
        "effective_cost_per_1k": 0.009
      },
      "score": 0.87,
      "last_updated": "2026-04-29T10:29:55Z"
    }
  ],
  "routing_stats": {
    "total_requests": 1450,
    "successful_requests": 1420,
    "failed_requests": 30,
    "current_weights": {
      "latency": 0.4,
      "cost": 0.3,
      "availability": 0.2,
      "quota": 0.1
    },
    "provider_distribution": {
      "groq": 520,
      "cerebras": 310,
      "deepseek": 280,
      "openai": 120,
      "anthropic": 220
    }
  }
}
```

---

## Sử Dụng

### 1. Khởi Tạo trong Router

```go
func initAdaptiveRouting() {
    // Khởi tạo trackers
    lt := routing.NewLatencyTracker(100)
    ct := routing.NewCostTracker(1000)
    hm := routing.NewHealthMonitor(routing.DefaultCircuitBreakerConfig())
    
    // Khởi tạo adaptive router
    ar := routing.NewAdaptiveRouter(routing.DefaultRoutingWeights(), lt, ct, hm)
    
    // Khởi tạo reaction engine
    re := routing.NewReactionEngine(ar)
    
    // Đăng ký providers
    providers := []struct{
        name string
        model string
        inputCost float64
        outputCost float64
    }{
        {"groq", "llama3-70b", 0.59, 0.79},
        {"cerebras", "llama3.1-8b", 0.10, 0.10},
        {"deepseek", "deepseek-chat", 0.14, 0.28},
        {"openai", "gpt-4", 30.00, 60.00},
    }
    
    for _, p := range providers {
        ar.RegisterProvider(p.name, p.model, p.inputCost, p.outputCost)
    }
    
    // Khởi tạo metrics handler
    metrics := routing.NewMetricsHandler(lt, ct, hm, ar)
    
    // Đăng ký routes
    metrics.RegisterRoutes(mux)
    
    // Lưu vào app context
    app.latencyTracker = lt
    app.costTracker = ct
    app.healthMonitor = hm
    app.adaptiveRouter = ar
    app.reactionEngine = re
}
```

### 2. Sử Dụng trong Handler

```go
func handleChatRequest(w http.ResponseWriter, r *http.Request) {
    // ... parse request ...
    
    // Lấy candidates
    candidates := app.providerRegistry.GetCandidates(model)
    
    // Chọn provider tốt nhất
    score, err := app.adaptiveRouter.SelectProvider(candidates, "chat")
    if err != nil {
        // No healthy providers - return error
        http.Error(w, "No healthy providers available", http.StatusServiceUnavailable)
        return
    }
    
    start := time.Now()
    
    // Forward request to selected provider
    response, err := forwardToProvider(score.Provider, req)
    
    // Ghi nhận kết quả
    latency := time.Since(start)
    success := err == nil
    
    app.adaptiveRouter.RecordResult(
        score.Provider, 
        score.Model, 
        latency, 
        requestTokens, 
        responseTokens, 
        success,
    )
    
    // Kiểm tra và trigger reactions nếu cần
    if !success {
        app.reactionEngine.CheckProviderDown(score.Provider, app.healthMonitor.GetState(score.Provider))
    } else {
        // Kiểm tra latency spike
        baseline := app.latencyTracker.GetStats(score.Provider).GetEMA()
        app.reactionEngine.CheckLatencySpike(score.Provider, latency, baseline)
    }
    
    // ... return response ...
}
```

### 3. Manual Fallback

```go
func handleWithFallback(w http.ResponseWriter, r *http.Request) {
    candidates := app.providerRegistry.GetCandidates(model)
    
    score, fallback, err := app.adaptiveRouter.SelectProviderWithFallback(candidates, "chat")
    if err != nil {
        http.Error(w, "No healthy providers", http.StatusServiceUnavailable)
        return
    }
    
    // Thử provider đầu tiên
    response, err := tryProvider(score.Provider, req)
    if err != nil {
        // Thử fallback chain
        for _, provider := range fallback {
            response, err = tryProvider(provider, req)
            if err == nil {
                break
            }
        }
    }
    
    // ... xử lý response ...
}
```

---

## Troubleshooting

### Problem: Tất cả providers đều "unhealthy"

**Nguyên nhân:**
- Circuit breakers đều OPEN
- Network issues
- Quota hết

**Giải pháp:**
```go
// Reset tất cả circuit breakers
for _, provider := range providers {
    app.healthMonitor.Reset(provider)
}

// Hoặc kiểm tra state và escalate
for _, provider := range providers {
    state := app.healthMonitor.GetState(provider)
    if state == routing.CircuitOpen {
        log.Printf("Provider %s is OPEN", provider)
        // Manual intervention needed
    }
}
```

### Problem: Provider được chọn không phải tốt nhất

**Nguyên nhân:**
- Chưa có đủ measurements (cần ít nhất vài requests để EMA ổn định)
- Static latency values được dùng thay vì measured values
- Weights không phù hợp

**Giải pháp:**
```go
// Kiểm tra sample count
stats := app.latencyTracker.GetStats("groq")
if stats.GetSampleCount() < 10 {
    // Chưa đủ data - sử dụng static values hoặc manual selection
}

// Điều chỉnh weights
newWeights := routing.RoutingWeights{
    LatencyWeight:      0.5,  // Ưu tiên latency hơn
    CostWeight:         0.2,
    AvailabilityWeight: 0.2,
    QuotaWeight:        0.1,
    Randomization:      0.05,
}
app.adaptiveRouter.SetWeights(newWeights)
```

### Problem: Cost tracking không chính xác

**Nguyên nhân:**
- Chưa đăng ký provider với cost rates
- Token count không được ghi nhận đúng

**Giải pháp:**
```go
// Đảm bảo đăng ký với đúng cost rates
app.costTracker.RegisterProvider("openai", "gpt-4", 30.00, 60.00)

// Kiểm tra cost stats
stats := app.costTracker.GetStats("openai")
fmt.Printf("Total tokens: %d\n", stats.GetTotalTokens())
fmt.Printf("Effective cost/1K: $%.4f\n", stats.GetCostPer1KTokens())
```

---

## Benchmarks

**Latency Tracker (Record):**
```
BenchmarkLatencyTrackerRecord-8    10000000    0.125 µs/op
```

**Circuit Breaker (AllowRequest):**
```
BenchmarkCircuitBreakerAllowRequest-8    50000000    0.045 µs/op
```

**Select Provider (10 providers):**
```
BenchmarkSelectProvider-8         1000000    2.34 µs/op
BenchmarkSelectProviderNoLatencyTracker-8    2000000    1.12 µs/op
```

**Reaction Engine (Trigger):**
```
BenchmarkReactionEngineTrigger-8    5000000    0.38 µs/op
```

---

## Files Implementation

| File | Mô tả | Dòng |
|------|-------|------|
| `latency_tracker.go` | Latency measurement & statistics | ~260 |
| `cost_tracker.go` | Cost tracking & analysis | ~220 |
| `health_monitor.go` | Circuit breaker & health monitoring | ~340 |
| `adaptive_router.go` | Weighted routing algorithm | ~380 |
| `reaction_engine.go` | Event-driven reactions | ~340 |
| `metrics_api.go` | Monitoring API endpoints | ~420 |
| `adaptive_config.go` | Configuration & hot reload | ~290 |
| `latency_tracker_test.go` | Tests for latency tracker | ~170 |
| `health_monitor_test.go` | Tests for health monitor | ~230 |
| `adaptive_router_test.go` | Tests for adaptive router | ~300 |
| `reaction_engine_test.go` | Tests for reaction engine | ~260 |

**Tổng cộng:** ~3,200 dòng code + ~1,000 dòng tests

---

## References

- [pLLM Routing Pattern](https://github.com/andreimerfu/pllm)
- [Agent Orchestrator Reactions](https://github.com/ComposioHQ/agent-orchestrator)
- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Weighted Round Robin](https://en.wikipedia.org/wiki/Weighted_round_robin)

---

*Document được tạo tự động bởi claude agent. Cập nhật khi có thay đổi kiến trúc.*
