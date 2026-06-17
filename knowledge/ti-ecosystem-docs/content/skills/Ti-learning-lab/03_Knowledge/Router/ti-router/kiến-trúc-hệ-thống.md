# Kiến Trúc Hệ Thống Ti Router

> **Project**: Ti Router  
> **Version**: 0.5.0  
> **Last Updated**: 2026-04-30  
> **Language**: Tiếng Việt

---

## Tổng Quan

Ti Router là một API proxy router local-first cho AI, cung cấp endpoint tương thích OpenAI để route requests đến nhiều AI providers với load balancing, failover, và tracking usage.

### Các Thành Phần Chính

1. **Layers Architecture** - Mô hình nhiều lớp (routing, resilience, metrics, authentication, audit)
2. **Routing Strategies** - Chiến lược routing (RoundRobin, LeastLatency, WeightedRR, BudgetBased, TagBased)
3. **Reliability Features** - Tính năng độ tin cậy (failover 3-level, circuit breaker, retry với exponential backoff)
4. **Observability** - Khả năng quan sát (Prometheus metrics, OpenTelemetry tracing, analytics dashboard)
5. **Performance** - Hiệu suất (Redis caching, SSE streaming, request batching, connection pooling)
6. **Security** - Bảo mật (rate limiting, JWT authentication, TLS encryption, audit logging)
7. **Advanced Features** - Tính năng nâng cao (A/B testing, canary deployment, feature flags, blue-green deployment)

---

## Chi Tiết Kiến Trúc

### 1. Layers Architecture

```
┌─────────────────────────────────────────────────────┐
│                    HTTP Layer                         │
│              (net/http, SSE streaming)                │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│                 Authentication Layer                   │
│         (API keys, JWT, OAuth2, session)              │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│                   Routing Layer                        │
│  (Selectors, Strategies, Latency Tracking, Budget)  │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│                 Resilience Layer                      │
│   (Circuit Breaker, Failover, Retry, Rate Limiting)   │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│                   Metrics Layer                         │
│       (Prometheus, Tracing, Analytics, Audit)          │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│                  Provider Layer                        │
│      (OpenAI, Anthropic, Google, xAI, etc.)           │
└─────────────────────────────────────────────────────┘
```

### 2. Routing Layer

**File**: `layers/routing/`

**Các Chiến Lược (Strategies)**:

- **RoundRobin**: Route tuần tự qua các providers
- **LeastLatency**: Route đến provider có độ trễ thấp nhất
- **WeightedRR**: Round Robin với trọng số tùy chỉnh
- **BudgetBased**: Route dựa trên ngân sách token/cost
- **TagBased**: Route dựa trên tags của provider/model

**Selector Factory**:
```go
SelectorFactory → tạo và đăng ký tất cả selectors
- RoundRobinSelector
- LeastLatencySelector (với LatencyTracker)
- WeightedRRSelector
- BudgetBasedSelector (với BudgetManager)
- TagBasedSelector
```

### 3. Resilience Layer

**File**: `layers/resilience/`

**Các Thành Phần**:

- **Circuit Breaker**: Ngăn request đến provider không khỏe
  - Health scoring (0-1): 0.7 * successRate + 0.3 * (1 - errorRate)
  - Adaptive thresholds
  - States: Closed, Open, Half-Open

- **Failover Manager**: 3-level failover
  - Level 1: Instance retry (cùng provider, endpoint khác)
  - Level 2: Route fallback (model khác trong cùng route)
  - Level 3: Global fallback (global fallback models)

- **Retry Logic**: Tự động retry với exponential backoff
  - MaxAttempts: 5 (Portkey pattern)
  - MaxDelay: 30s
  - Retryable status codes: 429, 500, 502, 503, 504

- **Rate Limiting**: Giới hạn rate per API key
  - Sliding window algorithm
  - Configurable limits per API key
  - Redis-based (MVP: in-memory)

### 4. Metrics Layer

**File**: `layers/metrics/`

**Các Metrics**:

- **Request Metrics**: total, success, errors, RPS, RPM
- **Latency Metrics**: avg, P50, P95, P99
- **Token Metrics**: prompt, completion, total
- **Cost Metrics**: total USD, per provider
- **Provider Metrics**: requests per provider, errors per provider, health scores
- **Connection Metrics**: active, peak

**Prometheus Format**: Standard Prometheus metrics cho integration với Grafana/Prometheus

### 5. Performance Layer

**File**: `layers/http/`, `layers/resilience/`

**Các Tối Ưu Hoá**:

- **Redis Caching**: Cache responses cho repeated queries
- **SSE Streaming**: Streaming responses cho long-running requests
- **Request Batching**: Batch multiple concurrent requests
- **Connection Pooling**: HTTP/2, custom dialer với timeout và keep-alive

### 6. Security Layer

**File**: `layers/authentication/`, `layers/security/`, `layers/audit/`

**Các Tính Năng**:

- **JWT Authentication**: HMAC-SHA256 signing, token validation, refresh
- **TLS Encryption**: TLS 1.2-1.3, certificate validation, curve preferences
- **Rate Limiting**: Per API key với sliding window
- **Audit Logging**: Detailed request/response logs với filtering

### 7. Advanced Features Layer

**File**: `layers/routing/`

**Các Tính Năng**:

- **A/B Testing Framework**: Test routing strategies với traffic allocation
  - Variants (A, B, ...) với weights
  - Traffic strategies: percentage, user_hash, random
  - Metrics tracking per variant
  - Winner selection dựa trên success rate

- **Canary Deployment**: Gradual rollout với auto-rollback
  - Versions với weights
  - Gradual promotion
  - Auto-rollback khi error rate > threshold
  - Health checks

- **Feature Flags System**: Dynamic configuration without deployment
  - Boolean, percentage, user_list, custom flags
  - Rule evaluation (user_id, percentage, custom)
  - Cache với TTL
  - Export JSON

- **Blue-Green Deployment**: Zero-downtime deployment
  - Blue/Green environments
  - Switch với health check
  - Rollback support
  - Health check config

---

## Flow Request

### 1. Chat Completion Request Flow

```
Client Request
    ↓
Authentication (API Key/JWT)
    ↓
Rate Limiting Check
    ↓
Circuit Breaker Check
    ↓
Routing Strategy Selection
    ↓
Provider Selection (LeastLatency/WeightedRR/etc.)
    ↓
Failover (if needed)
    ↓
Request to Provider
    ↓
Response Processing
    ↓
Metrics Recording
    ↓
Audit Logging
    ↓
Response to Client
```

### 2. A/B Testing Flow

```
Create A/B Test
    ↓
Define Variants (A, B, ...)
    ↓
Set Traffic Allocation
    ↓
Start Test
    ↓
For each request:
    - Select variant based on strategy
    - Route to variant's selector
    - Record metrics per variant
    ↓
Stop Test
    ↓
Select Winner (highest success rate)
    ↓
Promote Winner
```

### 3. Canary Deployment Flow

```
Create Canary Deployment
    ↓
Define Versions (stable, canary)
    ↓
Set Initial Weights (e.g., 90% stable, 10% canary)
    ↓
Start Deployment
    ↓
Monitor Metrics:
    - Error rate
    - Latency
    - Success rate
    ↓
Gradual Promotion (if metrics good)
    ↓
Auto-Rollback (if error rate > threshold)
    ↓
Complete (when canary weight = 100%)
```

---

## Configuration

### Providers Configuration

**File**: `configs/providers.yaml`

```yaml
providers:
  openai:
    name: OpenAI
    base_url: https://api.openai.com/v1
    api_key_env: OPENAI_API_KEY
    timeout_sec: 60
    models:
      - name: gpt-5.4
        timeout_sec: 120
      - name: gpt-4o
        timeout_sec: 60
```

### Server Configuration

**File**: `configs/Tiserverrouter.yaml`

```yaml
server:
  port: 20128
  host: localhost
  require_login: false
```

---

## Testing

### Unit Tests

- Routing strategies: `layers/routing/*_test.go`
- Resilience features: `layers/resilience/*_test.go`
- Metrics: `layers/metrics/*_test.go`
- Authentication: `layers/authentication/*_test.go`
- Advanced features: `layers/routing/*_test.go`

### Integration Tests

**File**: `tests/integration/routing_integration_test.go`

Test integration giữa các components:
- Routing + Selectors
- A/B Testing + Routing
- Canary + Metrics
- Feature Flags + Routing
- Blue-Green + Health Checks
- Circuit Breaker + Routing
- Failover + Retry

### E2E Tests

**File**: `tests/e2e/api_e2e_test.go`

Test complete API flow:
- Start server
- Make real HTTP requests
- Test all endpoints
- Verify responses

---

## Best Practices

### 1. Error Handling

- Use circuit breaker để prevent cascading failures
- Implement 3-level failover cho maximum reliability
- Retry với exponential backoff cho transient errors
- Log tất cả errors với context

### 2. Performance

- Use connection pooling để reduce overhead
- Cache responses cho repeated queries
- Use SSE cho streaming responses
- Batch concurrent requests

### 3. Security

- Validate all inputs
- Use TLS cho encryption in transit
- Implement rate limiting để prevent abuse
- Log tất cả security events

### 4. Observability

- Track metrics cho mọi requests
- Use distributed tracing cho debugging
- Monitor health scores của providers
- Set up alerts cho critical metrics

---

## Deployment

### Local Development

```bash
# Start server
go run cmd/routerd/main.go

# Or use scripts
.\start-simple.bat
```

### Production

- Use TLS encryption
- Configure proper rate limits
- Set up monitoring và alerting
- Use Redis cho distributed caching
- Implement proper backup strategy

---

## Troubleshooting

### Common Issues

1. **Build Errors**: Check Go version (requires 1.26.1)
2. **Provider Errors**: Check API keys, network connectivity
3. **Circuit Breaker Open**: Wait cho recovery hoặc manual reset
4. **Rate Limit Exceeded**: Increase limit hoặc wait
5. **High Latency**: Check provider status, network latency

---

## Next Steps

1. **Integration Testing**: Complete E2E tests với running server
2. **Documentation**: Cập nhật API docs với new endpoints
3. **Monitoring**: Set up Grafana dashboard
4. **Optimization**: Performance tuning cho production
5. **Scaling**: Horizontal scaling với load balancer
