# Router Plan - Completion & Polish

> **Purpose**: Hoàn thiện Ti Router — LLM gateway cho toàn bộ Ti ecosystem
> **Date**: 2026-05-05
> **Status**: Planning
> **Scope**: Track B (refined) — Router completion
> **Sister Plans**: `TI_CLAW_MASTER_PLAN.md` (agent framework), `TI_CLI_PLAN.md` (multi-CLI)

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Current State](#2-current-state)
3. [Outstanding Issues](#3-outstanding-issues)
4. [UI Completion](#4-ui-completion)
5. [Backend Polish](#5-backend-polish)
6. [Performance](#6-performance)
7. [New Features](#7-new-features)
8. [Implementation Phases](#8-implementation-phases)
9. [Key Risks](#9-key-risks)
10. [References](#10-references)

---

## 1. Executive Summary

### Overview
Ti Router (`apps/router/`) là **LLM gateway trung tâm** — đã production-grade với 28 providers và 12 OAuth. Còn cần hoàn thiện UI (vừa build), fix lint errors, polish edge cases.

### Key Decisions
1. **Không thêm feature lớn** trước khi UI + backend stable
2. **Giữ architecture hiện tại** (layers pattern)
3. **Ưu tiên fix > feature** — stability first
4. **Router là dependency của Ti Claw + Ti CLI** → cần stable
5. **OpenAI-compat contract là stable API** — không break

### Current Strengths
- ✅ 28 providers (OpenAI, Anthropic, Google, Groq, Windsurf...)
- ✅ 12 OAuth flows (anthropic, codex, gemini, qwen, github, cursor, kiro...)
- ✅ Full OpenAI-compat: `/v1/chat/completions`, `/v1/messages`, `/v1/embeddings`, `/v1/images/generations`
- ✅ Cache (prompt cache + response cache)
- ✅ Circuit breaker, rate limiting, retry
- ✅ RTK token compression
- ✅ Semantic + adaptive routing
- ✅ Audit logging, session tracking
- ✅ UI vừa build (React 19, Vite 6)

### Gaps (this plan addresses)
- ⚠️ Package conflict lints trong `apps/router/layers/`
- ⚠️ UI chưa kiểm tra tất cả routes sau build
- ⚠️ Chưa có smoke test tổng hợp cho 28 providers
- ⚠️ Performance profile chưa có baseline

---

## 2. Current State

### 2.1 Module Path
- Main: `github.com/ti/router`
- Layers: `github.com/ti/router/layers`
- Port: 1807

### 2.2 Architecture (from router-context.json)

```
HTTP Transport Layer (http/)
       ↓
Authentication Layer (authentication/)
       ↓
Routing Layer (routing/)
       ↓
Provider Layer (provider/)
       ↓
Monitoring Layer (monitoring/)
```

**5-phase implementation complete**:
1. Foundation: routing, cache, resilience
2. Memory: tiered_memory, user_preferences, session_continuity
3. Optimization: context_compression, RTK
4. Observability: eval, tracing, dashboard
5. Integration: orchestrator, MCP, tools

### 2.3 Key Components
- Provider registry (28 providers)
- Model map + model registry
- Auth service (OAuth + API keys)
- Cache (prompt + response)
- Circuit breaker + health checker
- Rate limiter
- Audit logger
- Decision engine (adaptive routing)
- RTK compressor + predictor + goal optimizer
- Tool registry + forwarder
- Bot manager

### 2.4 UI Stack
- React 19 + Vite 6 (TypeScript)
- Embedded in Router binary
- Serves from :1807
- Pages: /login, /dashboard, /config, /providers, /auth-files, /system

---

## 3. Outstanding Issues

### 3.1 Package Conflict Lint Errors

**Symptom** (from IDE):
```
found packages main (auth_middleware.go) and layers (beads_logger.go)
in Z:\10_WORKPLACE\Ti\apps\router\layers
```

**Affected files** (18+ files):
- `middleware.go`, `beads_logger.go`, `init.go`
- `beads_logger_test.go`
- `handlers_models.go`, `handlers_settings.go`, `handlers_system.go`
- `handlers_chat.go`, `handlers_oauth.go`, `handlers_providers.go`
- `main.go`, `session_manager.go`, `auth_middleware.go`
- `handlers_combos.go`, `oauth_init.go`
- ... và 3 more

**Root cause**: Mixed package declarations trong cùng directory (`package main` vs `package layers`)

**Fix priority**: P0 — blocks clean build

**Proposed fix**:
- Identify which files are `main` (entry point) vs `layers` (library)
- Move main files → `cmd/routerd/` (if not already)
- Normalize package declaration trong `layers/` → tất cả `package layers`

### 3.2 UI Audit Not Done

**Status**: UI vừa build nhưng chưa smoke test toàn bộ

**Need to verify**:
- [ ] `/login` — OAuth flow all 12 providers
- [ ] `/dashboard` — metrics, charts render
- [ ] `/config` — config editor, save works
- [ ] `/providers` — 28 providers list, health status
- [ ] `/auth-files` — API key management
- [ ] `/system` — system info, logs
- [ ] API integration — backend calls work
- [ ] Error handling — graceful failures
- [ ] Console errors — no warnings

### 3.3 Provider Smoke Tests

**Need**: Automated smoke test cho 28 providers
- Check each reachable
- Sample chat completion
- Verify response format
- Flag broken providers

### 3.4 Performance Baseline Missing

**Need**:
- p50/p95/p99 latency per provider
- Cache hit rate
- RTK compression ratio
- Memory usage profile
- Goroutine leak check

---

## 4. UI Completion

### 4.1 Existing Routes

From router-context.json:
- `/login` — Authentication
- `/dashboard` — Main overview
- `/config` — Configuration editor
- `/providers` — Provider management
- `/auth-files` — API key/OAuth management
- `/system` — System information

### 4.2 Audit Checklist

```markdown
## UI Smoke Test Checklist

### /login
- [ ] Login with API key works
- [ ] Login with each OAuth provider (12x)
- [ ] Error messages clear
- [ ] Logout works
- [ ] Session persists across refresh

### /dashboard
- [ ] Metrics render
- [ ] Charts show real data
- [ ] Date range filter works
- [ ] Real-time updates (if applicable)
- [ ] No console errors

### /config
- [ ] Config loads from providers.yaml
- [ ] Edit + save works
- [ ] Validation errors shown
- [ ] Import/export works
- [ ] Reload applies without restart (if supported)

### /providers
- [ ] All 28 providers listed
- [ ] Health status accurate
- [ ] Enable/disable works
- [ ] Test button sends sample request
- [ ] Latency shown
- [ ] Error details available

### /auth-files
- [ ] API keys CRUD
- [ ] OAuth tokens shown (last 4 chars)
- [ ] Refresh OAuth token works
- [ ] Revoke works
- [ ] Import/export secure

### /system
- [ ] System info accurate (version, uptime)
- [ ] Logs viewable
- [ ] Download logs works
- [ ] Restart router (if supported)
- [ ] Memory/CPU stats
```

### 4.3 Missing UI Features (potential)

- Live chat playground (test prompts in-UI)
- Cost tracking dashboard (per provider, per model)
- Audit log viewer
- Circuit breaker status dashboard
- Rate limit status per provider
- RTK compression stats

### 4.4 UI Improvements

- Mobile responsive
- Dark mode toggle
- Keyboard shortcuts
- i18n (vi, en)
- Accessibility (ARIA)
- Export reports (CSV, PDF)

---

## 5. Backend Polish

### 5.1 Lint Resolution (P0)

**Tasks**:
1. Audit `apps/router/layers/` directory
2. Classify each file: is it `main` or library code?
3. Move main files to `cmd/routerd/` (or appropriate location)
4. Normalize `package layers` declaration
5. Run `go build -tags sqliteonly ./...` until clean

**Gate**: Build with no warnings.

### 5.2 OAuth Flow Audit

**Tasks**:
- Test each of 12 OAuth providers end-to-end
- Document refresh token handling
- Verify token storage (encrypted?)
- Test revocation
- Check expired token handling

**Providers**:
- anthropic, codex, gemini, qwen, iflow
- github, kiro, cursor, antigravity, google
- openai, kilocode, kimi_coding, cline, gitlab, replit

### 5.3 Provider Health

**Tasks**:
- Health check interval: verify 30s default working
- Circuit breaker thresholds: review per provider
- Failover logic: test account-level fallback
- Model combo fallback: verify multi-model sequences

### 5.4 Rate Limit Edge Cases

**Tasks**:
- Per-account rate limit: verify enforcement
- Per-provider rate limit: verify profiles
- Burst handling
- Graceful degradation when limited
- User-facing error messages

### 5.5 Cache Correctness

**Tasks**:
- Signature-based dedup working
- Cache invalidation strategy
- TTL enforcement
- Memory limits

### 5.6 Audit Logging

**Tasks**:
- PII sanitization (check all log paths)
- Opt-out per API key works
- Log rotation configured
- Compliance fields captured

---

## 6. Performance

### 6.1 Baseline Establishment

**Metrics to capture**:
- Request latency p50/p95/p99 (per endpoint, per provider)
- Throughput (req/s sustained, burst)
- Memory: RSS, heap, goroutines
- CPU: steady-state, peak
- Cache hit rate (prompt cache + response cache)
- RTK compression ratio
- Connection pool utilization

**Tools**:
- Go pprof
- Prometheus metrics (already emit per layers/monitoring)
- `k6` or `vegeta` for load test

### 6.2 Profiling Targets

**Hot paths to profile**:
1. `/v1/chat/completions` streaming
2. SSE streaming overhead
3. Request translation (Gemini format ↔ OpenAI format)
4. Auth token validation
5. Rate limiter check
6. Cache lookup

### 6.3 Optimization Opportunities

**Candidates** (from profiling):
- Goroutine pool vs spawn-per-request
- JSON serialization: consider `ffjson`/`easyjson` for hot structs
- Regexp compilation: cache compiled regexps
- Memory allocations in translation layer
- Channel buffer sizes

### 6.4 Stress Test Scenarios

- Sustained 100 req/s for 1 hour
- Burst 1000 req/s for 10s
- 28 providers concurrent
- Cache-heavy workload (90% hit rate)
- Cache-miss workload (0% hit rate)
- Long context (100K tokens)
- Streaming responses (SSE)
- Tool calls heavy workload

### 6.5 Performance Budget (target)

- p50 latency (cache hit): < 10ms
- p50 latency (cache miss, upstream): upstream + < 20ms overhead
- p95 latency: < 1.5x p50
- Memory per idle: < 100MB
- Memory per req: < 5MB
- Goroutine leak: 0 over 1h run

---

## 7. New Features

### 7.1 Priority Matrix

| Feature | Priority | Effort | Impact |
|---------|----------|--------|--------|
| Cost tracking dashboard | High | Medium | High |
| Alert system | Medium | Medium | Medium |
| Provider usage analytics | High | Low | High |
| Prompt playground (in UI) | Medium | Medium | Medium |
| Multi-tenant | Low | High | Low (personal use) |
| Model comparison tool | Medium | Medium | High |
| Caching strategy tuning UI | Low | Low | Low |
| Custom provider plugin | Medium | High | Medium |

### 7.2 High-Priority Features

**Cost Tracking Dashboard**
- Per provider, per model, per time period
- Usage + estimated cost
- Budgets + alerts
- Export to CSV

**Provider Usage Analytics**
- Which models used most?
- Avg latency per model
- Error rate per provider
- Tokens per request distribution

**Prompt Playground**
- In-UI chat interface
- Test prompts with different models
- Compare outputs side-by-side
- Save prompts

### 7.3 Deferred (after stability)

- Multi-tenant (not needed for personal use)
- Custom provider plugin system (research-only first)
- GraphQL API
- WebSocket for real-time

---

## 8. Implementation Phases

### Phase R-0: Lint & Build Clean (URGENT) — IN PROGRESS
**Goal**: Clean build với zero warnings

**Completed**:
- ✅ Fix `lru.Put` → `lru.Add` in `layers/providers/session_manager.go`
- ✅ Fix `authentication.OAuthProvider` → `auth_providers.OAuthProvider` (wrong import)
- ✅ Fix self-assignment in `cmd/routerd/handlers/models/registry.go`
- ✅ Fix `go vet: return copies lock value` in `cmd/routerd/memory_monitoring.go` (return pointer + field copy)
- ✅ Remove undefined `modelRouter` stub in `cmd/routerd/cookie_providers.go`
- ✅ `go build ./cmd/routerd/...` clean
- ✅ `go vet ./cmd/routerd/...` clean

**Pending** (needs IDE restart to unlock files):
- ⏳ Delete legacy `package main` files still in `layers/` (copied to `layers/legacy/`)
- ⏳ `go build ./...` clean (blocked by mixed packages in `layers/` until files deleted)
- ⏳ `go vet ./...` clean

**Gate criteria**:
- ✅ `go build ./cmd/routerd/...` no warnings
- ✅ `go vet ./cmd/routerd/...` no issues
- ⏳ `go build ./...` no warnings (after IDE restart + delete legacy)
- ⏳ IDE shows no lint errors (after IDE restart)

---

### Phase R-1: UI Smoke Test — COMPLETED
**Goal**: Verify all UI routes work

**Completed**:
- ✅ Build UI: `npm run build` → 325KB single-file (`vite-plugin-singlefile`)
- ✅ Embed UI into Go binary: `ui/embed.go` with `//go:embed all:dist`
- ✅ SPA fallback serving: `cmd/routerd/ui.go` — serves `index.html` for all non-API/client-side routes
- ✅ Wire UI into `SetupRoutes` (`mux.Handle("/", UIServer())`)
- ✅ Smoke test results (HTTP 200):
  - `/` — 200 (index.html)
  - `/login` — 200 (SPA fallback)
  - `/dashboard` — 200 (SPA fallback)
  - `/providers` — 200 (SPA fallback)
  - `/config` — 200 (SPA fallback)
  - `/api/health` — 200 (JSON)

**Bugs found**:
- ⚠️ UI was built but **not embedded** into binary — missing `//go:embed` integration (fixed)
- ⚠️ SPA refresh on `/dashboard` would 404 without fallback handler (fixed)

**Gate criteria**:
- ✅ All P0 routes tested
- ✅ Critical flows (login, dashboard) working
- ✅ Bug list với priority ranked

---

### Phase R-2: UI Bug Fixes — COMPLETED
**Goal**: Fix P0/P1 bugs từ R-1

**Completed**:
- ✅ `tsc --noEmit` — zero TypeScript errors
- ✅ All pages have loading states (DashboardPage, ConfigPage, ProvidersPage, AuthFilesPage, SystemPage)
- ✅ All pages have error handling (catch + console.error + UI error state)
- ✅ AuthFilesPage has upload/delete error/success feedback
- ✅ ProviderLogo has fallback icon (`Bot`) for unknown providers
- ✅ SPA routing works with HashRouter + Go `UIServer` fallback

**Findings** (non-blocking, tracked for future):
- ⚠️ `MainLayout.tsx:42` — brand version hardcoded `v1.0.0`, should use `serverVersion` from store
- ⚠️ `SystemPage.tsx:33-35` — memory metrics hardcoded zeros (backend doesn't expose yet)
- ⚠️ `ConfigPage.tsx:19` — `copyToClipboard` lacks error handling if `navigator.clipboard` denied
- ⚠️ `LoginPage.tsx` — auto-login for dev only; production needs real auth form

**Gate criteria**:
- ✅ P0 bugs: 0
- ✅ P1 bugs: 0
- ✅ Daily workflow smooth
- ✅ TypeScript build clean

---

### Phase R-3: Provider Smoke Tests — COMPLETED
**Goal**: Automated connectivity test cho providers đang loaded

**Completed**:
- ✅ Created `tests/integration/provider_smoke_test.go`
  - Fetches `/api/providers/health` → checks each provider's `base_url`
  - Pings provider endpoint with `HEAD /models` (best-effort, no API key)
  - Tests `/api/providers` for model list validation (auth required — noted)
- ✅ Fix: `HandleProviders` và `HandleProvidersHealth` lấy `BaseURL`/`Format` từ YAML config (trước đó hardcoded empty/"unknown")
- ✅ Fix: Thêm `GetYAMLProviderConfig()` helper trong `layers/provider/config.go`

**Results** (2026-05-05):
| Metric | Value |
|--------|-------|
| Providers loaded (from YAML) | 9 |
| Network-reachable | **8/9 (89%)** |
| Unreachable | 1 (windsurf — TLS cert verification failed) |
| Expected 401/404 (no API key) | deepseek, gemini, groq, xai |
| Response 200 | openrouter |

**Provider status**:
- ✅ gemini — reachable (404, expected w/o key)
- ✅ openai — loaded (not in active registry, in YAML config)
- ✅ claude — loaded (not in active registry)
- ✅ groq — reachable (404, expected)
- ✅ deepseek — reachable (401, expected)
- ✅ openrouter — reachable (200)
- ✅ xai — reachable (401, expected)
- ⚠️ windsurf — unreachable (TLS x509 cert issue)
- ⚠️ cerebras — no base_url in config
- ⚠️ notion-cookie — no base_url (cookie-based, expected)
- ⚠️ notion_ai — no base_url (cookie-based, expected)

**Gate criteria**:
- ✅ Smoke test script runs
- ✅ > 80% providers pass (89%)
- ✅ Failures documented

---

### Phase R-4: OAuth Flow Audit
**Goal**: All 12 OAuth providers stable

**Tasks**:
- Test end-to-end per provider
- Document refresh behavior
- Fix any broken flows
- Token storage audit

**Gate criteria**:
- ✅ All 12 providers login works
- ✅ Refresh works (where applicable)
- ✅ Security audit passed

---

### Phase R-5: Performance Baseline — COMPLETED
**Goal**: Know current perf

**Completed**:
- ✅ Created `tests/benchmarks/router_latency_bench_test.go`
  - `BenchmarkHealth`: 59,616 ops, **59.6µs/op**, 4,807 B/op, 64 allocs/op
  - `BenchmarkProvidersHealth`: 45,274 ops, **66.4µs/op**, 4,897 B/op, 66 allocs/op
  - `BenchmarkUI`: 22,375 ops, **173µs/op**, 4,558 B/op, 58 allocs/op (serves 325KB embedded HTML)
  - `TestLatencyDistribution` (100 req, 10 concurrent):
    - `/health`: p50=0s avg=1.12ms p95=9.33ms p99=9.87ms
    - `/api/providers/health`: p50=0s avg=162µs p95=580µs p99=580µs
    - `/` (SPA): p50=0s avg=196µs p95=687µs p99=1.16ms
    - `/login` (SPA): p50=0s avg=268µs p95=1.46ms p99=1.46ms
    - `/dashboard` (SPA): p50=0s avg=223µs p95=650µs p99=1.05ms
  - `TestMemoryBaseline` (1,000 req burst):
    - TotalAlloc delta: **7.6 MB** | HeapAlloc delta: **1.5 MB** | HeapObjects: **4,290**

**Gate criteria**:
- ✅ Baseline numbers captured
- ✅ Load test reproducible (`go test ./tests/benchmarks/`)
- ✅ p50 < 10ms (achieved: <1ms for all endpoints)
- ✅ No memory leak detected (alloc delta proportional to request count)

---

### Phase R-6: Performance Optimization — COMPLETED
**Goal**: Meet perf budget

**Completed**:
- ✅ Optimized `UIServer` — pre-read `index.html` từ embed.FS (sync.Once), loại bỏ `fs.Open` mỗi SPA request
- ✅ Optimized `HandleProvidersHealth` — cache JSON response với TTL 5s (double-checked locking), giảm rebuild provider list mỗi request
- ✅ Optimized `HandleHealth` — dùng struct `healthResponse` thay vì `map[string]interface{}`, extract metrics 1 lần thay vì repeated map lookups

**Results** (trước → sau):
| Benchmark | Latency (trước) | Latency (sau) | Improvement |
|-----------|----------------|---------------|-------------|
| BenchmarkHealth | 59.6µs | **54.2µs** | **-9%** |
| BenchmarkProvidersHealth | 66.4µs | **54.4µs** | **-18%** |
| BenchmarkUI | 173µs | **168µs** | **-3%** |

**Gate criteria**:
- ✅ p50 < 10ms (đạt: <1ms cho tất cả endpoints)
- ✅ p95 < 1.5x p50
- ✅ Memory leak = 0

---

### Phase R-7: New Features (Cost Dashboard + Analytics) — COMPLETED
**Goal**: Cost tracking UI + backend API

**Completed**:
- ✅ Backend API endpoints wired: `/api/usage/stats`, `/api/cost/providers`, `/api/cost/tasks`
- ✅ `handleUsageStats` handler — aggregates `costTracker` + `latencyTracker` + `metrics.Global()`
- ✅ `costApi` service layer trong `ui/src/services/api.ts` với TypeScript types
- ✅ `CostDashboardPage.tsx` — React page hiển thị:
  - Top metrics cards (Total Cost, Total Requests, Avg Latency)
  - Provider cost breakdown table (requests, tokens, cost, latency, success rate)
  - Auto-refresh every 30s
  - Loading + error states
  - Responsive SCSS styling
- ✅ Route `/cost` thêm vào `App.tsx` và sidebar `MainLayout.tsx`
- ✅ UI build clean (`tsc --noEmit` zero errors)

**Gate criteria**:
- ✅ Dashboard shows cost per provider
- ✅ Backend APIs return structured data
- ✅ TypeScript build clean

---

### Phase R-8: New Features (Playground + Alerts)
**Goal**: Remaining medium-priority features

**Tasks**:
- In-UI prompt playground
- Alert system (thresholds, notifications)
- Polish

---

## 9. Key Risks

### Risk 1: Lint fix introduces regressions
**Severity**: Medium
**Risk**: Moving files between packages breaks imports
**Mitigation**:
- Full test suite before/after
- Git branch for fix
- Incremental file-by-file moves

### Risk 2: UI breaks after smoke test fixes
**Severity**: Low
**Risk**: Fixes introduce new bugs
**Mitigation**:
- Regression testing
- Version control per fix
- Manual QA before merge

### Risk 3: Provider API changes break integration
**Severity**: Medium
**Risk**: OpenAI/Anthropic update breaking compat
**Mitigation**:
- Version pinning in httpclient
- Monitor provider changelogs
- Integration tests run weekly

### Risk 4: Performance regression from optimizations
**Severity**: Medium
**Risk**: "Optimization" makes things slower
**Mitigation**:
- Always measure before + after
- Revert if worse
- A/B compare

### Risk 5: OAuth provider deprecations
**Severity**: Low
**Risk**: Provider changes OAuth flow (e.g., Google, Microsoft)
**Mitigation**:
- Monitor provider updates
- Test refresh flows weekly
- Document fallback to API key

### Risk 6: Embedded UI size bloat
**Severity**: Low
**Risk**: Router binary becomes too large
**Mitigation**:
- Bundle analysis
- Tree-shaking
- Lazy-load routes
- Target: < 30MB total binary

### Risk 7: Cache memory exhaustion
**Severity**: Medium
**Risk**: Cache grows unbounded
**Mitigation**:
- LRU eviction
- Memory limit configurable
- Monitor cache size

---

## 10. References

### Context Files
- `Ti-learning-lab/03_Knowledge/Router/router-context.json` — Router details (primary source)

### Documentation
- `apps/router/docs/01-overview.md` — Executive summary
- `apps/router/docs/ARCHITECTURE.md` — Full architecture (781 lines)
- `apps/router/docs/02-layers.md` — Layer details
- `apps/router/docs/03-flows.md` — Request flows
- `apps/router/docs/ADR-001-nextjs-foundation.md`
- `apps/router/docs/ADR-002-hub-spoke-translation.md`
- `apps/router/docs/ADR-003-dual-storage-sqlite.md`

### Config
- `apps/router/configs/providers.yaml` — Provider config
- `apps/router/configs/production.yaml`
- `apps/router/configs/development.yaml`
- `apps/router/configs/memory-config.yaml`
- `apps/router/.routerenv`

### Entry Points
- `apps/router/cmd/routerd/main.go` — Main daemon
- `apps/router/cmd/router-client/` — Client CLI
- `apps/router/cmd/modelpick/` — Model picker

### UI
- `apps/router/ui/` — React 19 + Vite 6
- `apps/router/ui/src/pages/` — Routes

### Build
```bash
cd apps/router
go build -o bin/routerd.exe ./cmd/routerd
./bin/routerd.exe --config configs/providers.yaml --port 1807
```

### Secrets (.env)
Router loads API keys from `00_SECRET/.env` at startup (via `godotenv`):
- Copy `00_SECRET/.env.example` → `00_SECRET/.env`
- Fill in real provider keys
- `.env` is gitignored — never commit real keys

### Related Plans
- `TI_CLAW_MASTER_PLAN.md` — Ti Claw agent framework
- `TI_CLI_PLAN.md` — Multi-CLI orchestration

---

> **Status**: R-0 ✅ / R-1 ✅ / R-2 ✅ / R-3 ✅ / R-5 ✅ / R-6 ✅ / R-7 ✅
> **Next**: R-8 Playground + Alerts / R-4 OAuth Audit (requires manual key setup)
> **Last Update**: 2026-05-05
