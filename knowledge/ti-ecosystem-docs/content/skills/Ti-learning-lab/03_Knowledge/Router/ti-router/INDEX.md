# Ti Router Knowledge Base (Tiếng Việt)

> **Folder**: Ti Router Knowledge  
> **Purpose**: Tổng hợp kiến thức về router architecture, patterns, và implementations
> **Language**: Vietnamese (see [`../Router/`](../Router/) for English version)

---

## Cấu Trúc

```
ti-router/
├── kiến-trúc-hệ-thống.md    # Kiến trúc hệ thống Ti Router (Vietnamese)
├── providers/                # Provider configurations
└── INDEX.md                 # This file
```

---

## Tài Liệu Chính

### Kiến Trúc Hệ Thống
- **File**: [`kiến-trúc-hệ-thống.md`](./kiến-trúc-hệ-thống.md)
- **Mô tả**: Tổng quan kiến trúc Ti Router với layers architecture, routing strategies, resilience features
- **Nội dung chính**:
  - Layers Architecture (HTTP, Authentication, Routing, Resilience, Metrics, Provider)
  - Routing Strategies (RoundRobin, LeastLatency, WeightedRR, BudgetBased, TagBased)
  - Resilience Features (Circuit Breaker, Failover, Retry, Rate Limiting)
  - Observability (Prometheus metrics, OpenTelemetry tracing)
  - Performance (Redis caching, SSE streaming, request batching)
  - Security (JWT authentication, TLS encryption, rate limiting)
  - Advanced Features (A/B testing, canary deployment, feature flags, blue-green deployment)

---

## Cross-Reference

- **English version**: [`../Router/`](../Router/) - Router knowledge base in English
- **Router implementation**: `Z:\Ti\router\` - Main router codebase

---

**Last Updated**: 2026-05-05
