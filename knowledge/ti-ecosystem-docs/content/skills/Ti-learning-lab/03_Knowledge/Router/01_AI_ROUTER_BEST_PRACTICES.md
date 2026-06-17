# Best Practices cho AI LLM Router/Gateway

> **Nguồn**: Nghiên cứu từ LiteLLM, pLLM, Portkey-AI/gateway, LLMGateway  
> **Ngày**: 2026-04-29  
> **Agent**: claude  
> **Dự án áp dụng**: Ti Router

---

## 1. Tổng quan các Repo đã nghiên cứu

### 1.1. pLLM (andreimerfu/pllm) - Go
- **Stars**: ~9 | **Ngôn ngữ**: Go (Chi router, GORM, PostgreSQL, Redis)
- **Điểm mạnh**:
  - 100% OpenAI-compatible API
  - Routes (virtual model endpoints) với 4 strategies
  - 3-level failover: instance retry → route model fallback → fallback chain
  - Health tracking với circuit breakers
  - Multi-key load balancing
  - Prometheus metrics + OpenTelemetry tracing

### 1.2. Portkey-AI/gateway
- **Stars**: ~7k+ | **Ngôn ngữ**: TypeScript/Node.js
- **Điểm mạnh**:
  - 200+ LLMs, 50+ AI Guardrails
  - Fallbacks với configurable error triggers
  - Automatic retries (up to 5x) + exponential backoff
  - Load balancing với weights
  - Smart caching (simple + semantic)
  - Usage analytics (request volume, latency, cost, error rates)

### 1.3. LiteLLM
- **Stars**: ~15k+ | **Ngôn ngữ**: Python
- **Điểm mạnh**:
  - Simple-shuffle routing (default, production recommended)
  - Adaptive Router (AI-powered)
  - Request Prioritization / Scheduling
  - Budget Routing
  - Tag-based Routing
  - Health Check Driven Routing
  - Fallback Management Endpoints
  - Provider-specific Wildcard routing

### 1.4. LLMGateway (theopenco)
- **Stars**: ~tens | **Ngôn ngữ**: Không rõ
- **Điểm mạnh**:
  - Unified API interface (OpenAI format)
  - Usage analytics (requests, tokens, response times, costs)
  - Multi-provider support
  - Performance monitoring so sánh models

---

## 2. Kiến trúc chuẩn AI Gateway

```
Request → Auth → Rate Limit → Cache Check → Route Resolution
                                              ↓
                                       ┌──── Strategy ────┐
                                       │ priority         │
                                       │ least-latency    │
                                       │ weighted-rr      │
                                       │ random           │
                                       └──────────────────┘
                                              ↓
                                    Instance Selection
                                              ↓
                                    Health Check + Failover
                                              ↓
                              ┌──────────┐ ┌──────────┐ ┌──────────┐
                              │ OpenAI   │ │Anthropic │ │  Groq    │ ...
                              └──────────┘ └──────────┘ └──────────┘
```

**Tech stack reference**: Go · PostgreSQL (state) · Redis (cache, latency, locks, queues) · Prometheus (metrics) · OpenTelemetry (tracing)

---

## 3. Routing Strategies - Best Practices

### 3.1. Priority-based
- Chọn model theo priority order, fallback xuống model tiếp theo nếu fail
- Phù hợp: Production stability, ưu tiên quality

### 3.2. Least-latency
- Chọn deployment có latency thấp nhất trong thời gian gần đây
- Phù hợp: Real-time applications, chat UI

### 3.3. Weighted Round-Robin (simple-shuffle)
- Phân phối theo trọng số, xáo trộn để tránh burst
- **LiteLLM recommend làm default cho production**
- Phù hợp: High throughput, cost optimization

### 3.4. Random
- Đơn giản, không stateful
- Phù hợp: Dev/test environments

### 3.5. Health Check Driven
- Chỉ route đến healthy providers
- Kết hợp với circuit breaker

### 3.6. Adaptive / AI-powered
- Dùng ML để predict best provider dựa trên history
- Phù hợp: Complex workloads, multi-dimensional optimization

### 3.7. Budget Routing
- Route đến provider rẻ nhất khi trong budget limit
- Phù hợp: Cost-sensitive applications

### 3.8. Tag-based
- Route dựa trên request tags/metadata
- Phù hợp: Multi-tenant, per-team routing

---

## 4. Reliability Patterns

### 4.1. 3-Level Failover (pLLM pattern)
```
Level 1: Instance retry (cùng provider, different endpoint)
Level 2: Route model fallback (switch sang model khác trong cùng route)
Level 3: Fallback chain (global fallback models)
```

### 4.2. Circuit Breakers
- Open → Half-Open → Closed state machine
- Health scores với sliding window
- Auto-recovery sau backoff period

### 4.3. Automatic Retries
- Exponential backoff (Portkey: up to 5 retries)
- Configurable retry conditions (429, 5xx, timeout)
- Không retry 4xx client errors

### 4.4. Timeouts
- Granular timeouts per provider/per model
- Default timeout: 30-120s tùy provider
- Request termination khi vượt timeout

---

## 5. Observability Requirements

### 5.1. Metrics (Prometheus)
- Request rates (RPM, RPS)
- Latency percentiles (p50, p95, p99)
- Token usage (input/output/total)
- Cost per request/provider
- Error rates by status code
- Provider health scores

### 5.2. Distributed Tracing (OpenTelemetry)
- Trace ID across request lifecycle
- Span per provider call
- Tag routing decisions

### 5.3. Analytics
- Traffic distribution per route
- Cost breakdown per provider/model
- Performance comparison dashboard
- Budget consumption tracking

---

## 6. Security & Access Control

### 6.1. Authentication
- JWT với role-based access
- API Key management với per-key permissions
- Virtual keys (Portkey pattern)

### 6.2. Rate Limiting
- Per-user, per-model, per-endpoint
- Token-based rate limits (TPM)
- Request-based rate limits (RPM)

### 6.3. Budget Management
- User-level spending limits
- Team-level budgets
- Real-time cost tracking
- Alert khi approaching limit

### 6.4. Guardrails
- Input/output validation
- Content filtering
- PII detection
- Custom rules

---

## 7. Performance Optimization

### 7.1. Caching
- **Simple caching**: Exact match cache
- **Semantic caching**: Similar prompt cache (vector similarity)
- TTL-based expiration
- Cache hit rate monitoring

### 7.2. Connection Pooling
- HTTP keep-alive đến providers
- Connection pool size tuning
- Idle timeout management

### 7.3. Streaming
- Pass-through streaming responses
- No buffering cho SSE/WebSocket

### 7.4. Request Batching
- Batch multiple requests cho cùng provider
- Reduce overhead

---

## 8. Gap Analysis: Ti Router vs Best Practices

### 8.1. Đã có ✅
- Config-driven providers (YAML)
- Basic routing (priority, weight)
- Health checker
- Rate limiter
- Circuit breaker
- Load balancer
- Audit logger
- Model registry
- Decision engine
- RTK (Real-Time Knowledge)
- Brain (memory/embedded)
- Bot platforms
- MCP
- OAuth
- API key management
- Tool Search

### 8.2. Còn thiếu ❌
- **Routing Strategies**: least-latency, weighted-rr (chỉ có priority/weight), adaptive, budget-based, tag-based
- **3-Level Failover**: chỉ có 1-level retry
- **Caching**: không có caching layer
- **Metrics/Observability**: không có Prometheus integration
- **Distributed Tracing**: không có OpenTelemetry
- **Analytics Dashboard**: không có UI
- **Guardrails**: không có input/output validation
- **Budget Management**: không có spending limits
- **Semantic Caching**: không có vector similarity
- **Request Prioritization**: không có scheduler/queue
- **Virtual Keys**: không có key abstraction
- **Per-key Rate Limiting**: chỉ có global rate limit

---

## 9. Priority Roadmap cho Ti Router

### Phase 1: Foundation Fixes (WEEK 1)
- Fix build errors ✅ (đã xong)
- Secrets management (load từ Z:\00_SECRET\.routerenv) ✅ (đã xong)
- Config validation
- E2E tests cơ bản

### Phase 2: Routing Enhancement (WEEK 2)
- Implement least-latency routing
- Implement weighted-round-robin (simple-shuffle)
- Implement budget-based routing
- Implement tag-based routing
- Route configuration API (dynamic routes)

### Phase 3: Reliability & Resilience (WEEK 3)
- 3-level failover implementation
- Circuit breaker enhancement (health scoring)
- Automatic retries với exponential backoff
- Granular timeouts per provider

### Phase 4: Observability (WEEK 4)
- Prometheus metrics integration
- OpenTelemetry tracing
- Analytics dashboard (UI)
- Real-time monitoring

### Phase 5: Performance (WEEK 5)
- Simple caching layer (Redis)
- Connection pooling optimization
- Streaming improvements
- Request batching

### Phase 6: Security (WEEK 6)
- Per-key rate limiting
- Budget management
- Guardrails integration
- Virtual keys

### Phase 7: Advanced Features (WEEK 7-8)
- Semantic caching (vector similarity)
- Request prioritization / scheduler
- Adaptive routing (ML-based)
- Multi-tenant isolation

---

## 10. Lessons Learned

### 10.1. Architecture
- Go là ngôn ngữ tốt cho high-performance gateway (no GIL)
- Chi router + GORM + PostgreSQL + Redis là stack chuẩn
- Redis không chỉ cache mà còn latency tracking, locks, queues

### 10.2. Routing
- Simple-shuffle (weighted-rr) là default tốt nhất cho production
- Least-latency cần tracking latency history
- Priority routing đơn giản nhưng không tối ưu cost

### 10.3. Reliability
- 3-level failover critical cho production
- Circuit breaker không chỉ open/close mà cần health scoring
- Retries phải có exponential backoff

### 10.4. Observability
- Metrics phải có: request rate, latency, tokens, cost, errors
- Tracing cần trace ID + spans cho mỗi provider call
- Analytics dashboard cần traffic + cost breakdown

### 10.5. Security
- Per-key permissions critical cho multi-tenant
- Budget tracking real-time để tránh overspending
- Guardrails cho input/output validation

---

## 11. References

- pLLM: https://github.com/andreimerfu/pllm
- Portkey-AI/gateway: https://github.com/Portkey-AI/gateway
- LiteLLM: https://github.com/BerriAI/litellm
- LLMGateway: https://github.com/theopenco/llmgateway
- AWS AI Gateway Reference: https://aws.amazon.com/blogs/machine-learning/streamline-ai-operations-with-the-multi-provider-generative-ai-gateway-reference-architecture/
