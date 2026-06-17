# Router Repositories — Extracted Patterns & Knowledge

> **Source**: `Z:\Ti\Ti-learning-lab\07_Repositories\router\`
> **Date**: 2026-05-06
> **Purpose**: Distilled patterns from 7 router repos → apply to Ti Router Agent & Router

---

## 1. FreeLLMAPI — Fallback Chain + Dynamic Priority + Sticky Sessions

**Repo**: `freellmapi-main/` (TypeScript/Node)
**Key files**: `server/src/services/router.ts`, `ratelimit.ts`, `routes/proxy.ts`, `analytics.ts`

### 1.1 Dynamic Priority with 429 Penalty Decay

Instead of static fallback priority, FreeLLMAPI dynamically adjusts priority based on rate-limit hits:

```
effective_priority = base_priority + rate_limit_penalty
```

**Pattern**:
- On 429 → `penalty += 3` (capped at 10)
- On success → `penalty -= 1` (floor 0)
- Time-based decay: every 2 min → `penalty -= 1`
- Models sink in priority when rate-limited, automatically recover when stable

**Ti Application**: Enhance `StrategySelector` and `RoutingStrategy` to support penalty-weighted priority sorting with time-decay. This is complementary to circuit breaker but finer-grained — circuit breaker is binary (open/closed), penalty is continuous.

### 1.2 Sticky Sessions (Anti-Hallucination)

**Problem**: Model switching mid-conversation causes hallucination because different models have different context understanding.

**Pattern**:
- Hash first user message → map to model_db_id
- Multi-turn requests (has assistant messages) → prefer same model
- TTL: 30 min session stickiness
- Cleanup: evict entries when map > 500

**Ti Application**: Add `StickySessionManager` to Router Agent. For multi-turn conversations, prefer routing to same provider+model. This prevents quality degradation from model switching.

### 1.3 Sliding Window Rate Limiting (4-dimensional)

**Pattern**: Track 4 limits per key simultaneously:
- **RPM** (requests/minute) — sliding window
- **RPD** (requests/day) — sliding window
- **TPM** (tokens/minute) — sliding window with token counts
- **TPD** (tokens/day) — sliding window with token counts

Plus cooldown system: on 429, block specific model+key for configurable duration.

**Ti Application**: The Ti `SessionPool` and rate limiter should track all 4 dimensions per session/key, not just request count.

### 1.4 Round-Robin Key Rotation

**Pattern**: Per `platform:modelId`, maintain a round-robin index across all healthy API keys. Combined with skipKeys set for failed keys in current request.

**Ti Application**: When Ti Router has multiple keys per provider, use round-robin within `SessionPool.SelectSession()`.

### 1.5 Analytics: Error Distribution Categorization

**Pattern**: Auto-categorize errors via SQL CASE:
- `429 / rate limit / quota` → "Rate Limited"
- `401 / unauthorized` → "Auth Error"
- `timeout / ETIMEDOUT` → "Timeout"
- `500 / internal` → "Server Error"
- `503 / unavailable` → "Unavailable"

Plus: by-model, by-platform, timeline (hour/day granularity), cost estimation.

**Ti Application**: Router dashboard analytics should auto-categorize error types for quick diagnosis.

---

## 2. CLIProxyAPI — Go Proxy with Executor + Translator + Thinking Pipeline

**Repo**: `CLIProxyAPI-main/` (Go)
**Key files**: `AGENTS.md`, `docs/sdk-advanced.md`

### 2.1 Executor Pattern (Go)

**Pattern**: `ProviderExecutor` interface with:
- `Identifier() string` — provider key
- `PrepareRequest(req, auth)` — inject credentials
- `Execute(ctx, auth, req, opts)` — sync call
- `ExecuteStream(ctx, auth, req, opts)` — streaming call
- `Refresh(ctx, auth)` — token refresh

Core manager routes to correct executor by provider key. Clean separation of auth + transport + protocol.

**Ti Application**: Router Agent's provider abstraction should follow this executor pattern for clean per-provider logic separation.

### 2.2 Translator Registry (Format Conversion)

**Pattern**: Bidirectional format translation registered at init:
```
Register(sourceFormat, targetFormat, requestTransform, responseTransform)
```

Supports: OpenAI ↔ Gemini ↔ Claude ↔ Codex conversions. Stream and non-stream response transforms are separate.

**Ti Application**: When Ti Router proxies to different providers, use translator pattern to normalize request/response formats.

### 2.3 Thinking/Reasoning Pipeline

**Pattern**: `ApplyThinking()` pipeline:
1. Parse suffix overrides
2. Normalize to canonical `ThinkingConfig`
3. Validate centrally
4. Apply per-provider via `ProviderApplier`

Architecture: "canonical representation → per-provider translation". Never break this flow.

**Ti Application**: When routing thinking/reasoning requests, normalize thinking config before dispatching to provider-specific handlers.

### 2.4 Model Registry + Remote Updater

**Pattern**: Global model registry with `RegisterClient(authID, provider, models)`. Supports `--local-model` flag to disable remote updates. Hot-reload via config watcher.

**Ti Application**: Ti Router model catalog should support dynamic registration and hot-reload.

---

## 3. LLM Interactive Proxy — Python FastAPI with CBOR + Usage DB + Compression

**Repo**: `llm-interactive-proxy/` (Python/FastAPI)
**Key files**: `AGENTS.md`, `docs/database-usage-tracking.md`, `dev/caveman_style_compression.md`

### 3.1 Database-Backed Usage Tracking

**Pattern**: Full SQL-backed usage tracking:
- `usage_records` table with session_id, backend, model, frontend_type, leg, tokens (original + mutated), cost, timestamp
- Indexes on: session_id, backend_type, model, timestamp
- Aggregation service: by model, backend, frontend, leg, time range
- API endpoint: `/api/v1/usage`

**Ti Application**: Router Agent should persist usage records to DB for historical analytics, not just in-memory.

### 3.2 Caveman Compression (Output Token Saving)

**Pattern**: Proxy-side middleware that injects system prompt constraints to force LLM to output shorter responses:
- Modes: `lite`, `full`, `ultra`, `wenyan`
- Injected as non-forwardable message (doesn't pollute conversation history)
- Toggled via CLI command (`!/caveman ultra`)

**Implementation**: RequestTransformPipeline adds `_apply_output_compression_steering` step before dispatch.

**Ti Application**: Router Agent's `RTKCompressor` already compresses input. This pattern adds **output compression** — steering the LLM to be concise. Implement as configurable middleware in proxy pipeline.

### 3.3 Staged Initialization

**Pattern**: Service startup in ordered stages:
1. Infrastructure (DB, config)
2. Services (rate limiter, usage store)
3. Backends (connectors)
4. Controllers (HTTP handlers)

**Ti Application**: Router Agent bootstrap should follow staged init for clean dependency ordering.

### 3.4 Wire Capture (CBOR)

**Pattern**: Binary capture of all traffic in CBOR format for replay and debugging. Paired with inspection script.

**Ti Application**: Consider wire capture for Router debugging/replay scenarios.

---

## 4. Free Claude Code — Request Optimization + Model Routing + Rate Limiting

**Repo**: `free-claude-code-main/` (Python/FastAPI)
**Key files**: `api/optimization_handlers.py`, `api/detection.py`, `api/model_router.py`, `api/services.py`, `providers/rate_limit.py`, `core/rate_limit.py`

### 4.1 5-Category Request Optimization (Expanded Probe Detection)

**Pattern**: Chain of optimization handlers, cheapest first:
1. **Quota Mock** — `max_tokens=1` + "quota" keyword → mock response
2. **Prefix Detection** — `<policy_spec>` + `Command:` → extract prefix locally
3. **Title Skip** — "sentence-case title" system prompt → return "Conversation"
4. **Suggestion Skip** — `[SUGGESTION MODE:` → return empty
5. **Filepath Extraction** — `Command:` + `Output:` + filepath keywords → parse locally

Each handler is feature-flagged independently.

**Ti Application**: Expand Ti's `TrivialProbeDetector` beyond basic probes to include these 5 categories. The prefix detection and filepath extraction are particularly valuable for coding agents.

### 4.2 Per-Model Routing with Thinking Config

**Pattern**: `ModelRouter` resolves Claude model names to provider+model pairs:
```
claude-3-opus → nvidia_nim/glm4.7  (with thinking=true)
claude-3-sonnet → openrouter/gpt-oss (with thinking=false)
```

`ResolvedModel` carries: original_model, provider_id, provider_model, thinking_enabled.

**Ti Application**: Ti's `ModelTierResolver` should also carry thinking/reasoning config per tier, not just tier label.

### 4.3 Dual-Layer Rate Limiting (Proactive + Reactive + Concurrency)

**Pattern**: `GlobalRateLimiter` with 3 layers:
1. **Proactive**: Strict sliding window (N requests in W seconds) — prevents hitting limits
2. **Reactive**: On 429 → block all requests for exponential backoff duration
3. **Concurrency**: Semaphore cap on simultaneous open streams

Plus `execute_with_retry`: exponential backoff with jitter on 429.

Supports **scoped instances**: different limits per provider.

**Ti Application**: Ti Router should implement proactive rate limiting (not just reactive circuit breaker). The 3-layer approach prevents 429s proactively while handling them gracefully when they occur.

### 4.4 Safe Error Handling (User-Facing)

**Pattern**: `get_user_facing_error_message(exc)` transforms internal exceptions to safe public messages. Stack traces only logged when `log_api_error_tracebacks=true`. HTTP status chosen by exception type, not leaked error content.

**Ti Application**: Already have `SafeError` in `router_advanced.go`. Ensure all error paths use it consistently.

---

## 5. ProxyPal — Desktop App with Split Storage + Aggregate Analytics

**Repo**: `proxypal-main/` (Rust/Tauri + React)
**Key files**: `DESIGN_SPLIT_STORAGE.md`

### 5.1 Split Storage Architecture

**Pattern**: Two-file storage:
- `history.json` — 500 most recent requests (for UI display, trimmed)
- `aggregate.json` — cumulative counters, never trimmed:
  - totalRequests, totalSuccess, totalFailure
  - totalTokensIn/Out, totalCostUsd
  - requestsByDay, tokensByDay (time-series)
  - modelStats, providerStats

**Benefits**: Bounded memory (150KB), but never loses aggregate data.

**Ti Application**: Router Agent memory/analytics should separate hot data (recent decisions) from aggregate counters (lifetime stats). The agent's `GetMemory()` already has patterns; add aggregate analytics that survive memory pruning.

### 5.2 Tauri IPC State Management

**Pattern**: Shared state via `AppState` with `Mutex<T>` for interior mutability, `Arc<AtomicBool>` for flags. All IPC types use `serde(rename_all = "camelCase")`.

**Ti Application**: Router dashboard API types should use consistent camelCase for frontend compatibility.

---

## 6. OmniRoute — Already Documented

See existing knowledge files:
- `omniroute-patterns.md` — routing strategies, executor pattern, combo routing
- `omniroute-backend-analysis-ti-router.md` — backend analysis
- `oauth-session-pool-analysis.md` — OAuth session pool

---

## 7. Cross-Repo Pattern Summary — Priority for Ti Router

| # | Pattern | Source | Priority | Status |
|---|---------|--------|----------|--------|
| 1 | **Dynamic Priority + Penalty Decay** | FreeLLMAPI | HIGH | NEW — implement |
| 2 | **Sticky Sessions (anti-hallucination)** | FreeLLMAPI | HIGH | NEW — implement |
| 3 | **5-Category Request Optimization** | Free-Claude-Code | HIGH | PARTIAL — expand TrivialProbeDetector |
| 4 | **3-Layer Rate Limiting** | Free-Claude-Code | HIGH | NEW — add proactive layer |
| 5 | **4D Sliding Window (RPM/RPD/TPM/TPD)** | FreeLLMAPI | MEDIUM | NEW — enhance SessionPool |
| 6 | **Split Storage (hot + aggregate)** | ProxyPal | MEDIUM | NEW — implement |
| 7 | **Output Compression Steering** | LLM-Interactive | MEDIUM | NEW — add to RTK pipeline |
| 8 | **Executor + Translator Pattern** | CLIProxyAPI | LOW | EXISTING — refine |
| 9 | **Usage Tracking DB** | LLM-Interactive | LOW | FUTURE — persist analytics |
| 10 | **Thinking Pipeline Normalization** | CLIProxyAPI | LOW | FUTURE — for reasoning models |

### Immediate Integration Targets (for router_advanced.go + router_integration.go):

1. **DynamicPriorityManager** — penalty decay per provider, time-based recovery
2. **StickySessionManager** — conversation hash → provider affinity with TTL
3. **Expanded TrivialProbeDetector** — 5 categories from free-claude-code
4. **ProactiveRateLimiter** — sliding window pre-429 throttling
5. **SlidingWindowTracker** — 4D (RPM/RPD/TPM/TPD) per session
6. **AggregateAnalytics** — split hot/cold storage for Router Agent memory

---

*Generated from analysis of 7 router repositories. Ready for implementation.*
