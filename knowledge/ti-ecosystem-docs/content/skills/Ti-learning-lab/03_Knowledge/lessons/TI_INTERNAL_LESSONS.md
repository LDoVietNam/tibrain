# Ti Internal (Z:\01_PROJECTS\Ti\internal) Deep-Dive Notes

> **Date**: 2026-04-26  
> **Scope**: Patterns from the original `Ti` monorepo internal packages relevant to `Z:\Ti\router\layers`  
> **Goal**: Identify what to port, what to discard, and how `CLIProxyAPI-main` evolved from these roots.

---

## 1. Config — 7-Level Precedence System

**Source**: `internal/config/config.go`  
**Pattern**: Config resolution happens in 7 layers (highest → lowest):

1. `TI_*` env vars (e.g., `TI_AUTH_ROOT`, `TI_DATA_DIR`)
2. `--flag` CLI overrides
3. `config.yaml` explicit values
4. User home dir defaults (`~/.ti/`)
5. Windows-specific fallbacks (`Z:\04_CONFIG\`, `Z:\06_AUTH\`)
6. Hard-coded defaults
7. Zero values (for optional fields)

**Key types**:
- `Config` — top-level struct with `Default()` constructor
- `LoadReport` — list of `ValidationIssue` structs returned during load
- Path helpers: `buildWindowsPath`, `joinPath` — handle `\` vs `/` correctly

**Porting decision**: The layered precedence is overkill for a router-only module. Adopt a simpler YAML-first config (like `CLIProxyAPI-main`) with env overrides, but keep the `Default()` + `ValidationIssue` pattern for boot-time safety.

---

## 2. Auth — `Loader` with Mutex + Multi-Source Loading

**Source**: `internal/auth/auth.go`  
**Pattern**: `Loader` struct holds `sync.RWMutex` over a `map[string]*Entry`. Loads from:
- Environment variables (`ProviderMap` maps `OPENROUTER_API_KEY` → `"openrouter"`)
- Config file (`config.json`)
- Auth files (`auth_dir/*.json`)
- Cookie files
- OAuth tokens

```go
type Entry struct {
    ProviderID string
    AuthType   string   // "api_key", "oauth", "cookie"
    Source     string   // "env", "config", "auth_file"
    Value      string
    Expiry     time.Time
    Status     string   // "valid", "expired", "missing"
}
```

**Key insight**: `sourcePriority` map controls which source wins when multiple exist for the same provider. This avoids silent overwrites.

**Porting decision**: Port the `Loader` pattern into `layers/auth` but simplify to **env-first, config-second, DB-third**. The router layers already use `db.Database` for token storage — unify this with the loader so DB tokens override env vars only when explicitly configured.

---

## 3. Router Client — Functional Options + Retry

**Source**: `internal/router/router.go`  
**Pattern**: `Client` struct wraps `net/http.Client` with:
- `sync.RWMutex` for baseURL/apiKey mutations
- `ClientOption` functional options (`WithAPIKey`, `WithRetryPolicy`, `WithTimeout`)
- Retry logic with exponential backoff (350ms base)
- `tiMeta` map for custom headers

```go
func New(baseURL string, opts ...ClientOption) *Client {
    c := &Client{...defaults...}
    for _, opt := range opts { opt(c) }
    return c
}
```

**Porting decision**: This client is a **CLI consumer**, not a server. The `layers/http` server should not use this client directly. Instead, create a **ProviderClient** interface in `layers/provider` that each provider implements, then let `layers/http` handlers call `provider.Chat()` without knowing HTTP details.

---

## 4. Provider Interface — Rich `ChatRequest` with Tool Support

**Source**: `internal/providers/types.go`  
**Pattern**: `Provider` interface is deliberately narrow but the `ChatRequest` struct is rich:

```go
type Provider interface {
    Name() string
    DefaultModel() string
    Models() []string
    IsHealthy() bool
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error)
}
```

**ChatRequest fields** (beyond OpenAI standard):
- `ReasoningEffort` — "low"/"medium"/"high" (Claude thinking)
- `CacheControl` — prompt caching flag
- `Complexity` — 0-1 float for cost-aware routing
- `TaskType` — "coding", "review", "plan" → phase routing
- `MaxTurns` — agentic turn limit
- `AllowedTools` — restrict which tools the provider may use
- `Tools` — full `[]ToolDef` for agentic mode

**Porting decision**: Port the **entire `Provider` interface and `ChatRequest` struct** into `layers/provider`. The router layers currently have a stripped-down version. Add `ReasoningEffort`, `CacheControl`, `Complexity`, `TaskType`, and tool fields to enable feature parity with the monorepo.

---

## 5. Provider Registry — Simple Mutex Map

**Source**: `internal/providers/registry.go`  
**Pattern**: `Registry` holds `map[string]Provider` under `sync.RWMutex`. Provides:
- `Register`, `Get`, `List`, `Snapshots`
- `Select(criteria SelectionCriteria)` — basic routing by model name or preferred provider list

**Key gap**: No background health checks, no TTL cache, no cost-aware selection.

**Porting decision**: Keep the simple `Registry` as the **authoritative source** but wrap it with a **TTL-cached, health-aware `ModelRegistry`** (like `CLIProxyAPI-main`). This gives us:
- Fast reads (cache)
- Up-to-date availability (background updater)
- Cost-aware routing (Complexity field)

---

## 6. AutoCombo Pool — Static Candidate List

**Source**: `internal/router/autocombo_pool.go`  
**Pattern**: `AutoComboPool()` returns a static `[]autocombo.ProviderCandidate` hard-coded with cost, latency, quota, and tier data.

```go
{
    Provider: "openrouter", Model: "anthropic/claude-sonnet-4-5",
    QuotaRemaining: 80, CircuitState: autocombo.CircuitClosed,
    CostPer1MTokens: 3.0, P95LatencyMs: 1800,
    AccountTier: autocombo.TierPro,
}
```

**Porting decision**: This is **config data masquerading as code**. Move the static pool into `config.yaml` (or a `providers.yaml` file) and load it at boot. Keep `AutoComboPool()` as a thin loader that reads from config + dynamic health data.

---

## 7. Phase Routing — Map-Based Model Selection

**Source**: `internal/router/phase.go`  
**Pattern**: `PhaseModel` and `PhaseBudget` maps route workflow phases → models and budgets.

```go
var PhaseModel = map[string]string{
    PhaseScan:      "claude-sonnet-4",
    PhasePlan:      "claude-sonnet-4",
    PhaseImplement: "claude-sonnet-4",
    PhaseSummarize: "claude-haiku-4",
}
```

**Key insight**: `SelectModelForPhase` checks config override first, then phase map, then fallback. `SelectBudgetForPhase` does the same for budgets.

**Porting decision**: Port into `layers/routing/phase.go` but refactor:
- Use structs instead of maps (type safety, iteration order)
- Add `PhaseConfig` to `config.yaml` so users override without code changes
- Integrate with `ModelRegistry` so unavailable models fall back automatically

---

## 8. Management Service — Credential Retrieval

**Source**: `internal/management/service.go`  
**Pattern**: `ManagementService` holds `*config.Config`, `*providers.Registry`, and auth path helpers. `GetProviderCredentials()` iterates registry providers, calls `GetProviderConnections()` on the DB, and returns active credentials.

**Porting decision**: The management service belongs in a separate binary or package, not in `layers/http`. However, the **credential retrieval logic** (`GetProviderCredentials`) should move into `layers/provider/bootstrap.go` so the HTTP server can serve `/v1/models` with up-to-date provider lists.

---

## 9. Stream Parsing — SSE Line Scanner

**Source**: `internal/router/router.go` `parseStreamResponse()`  
**Pattern**: Manual `bufio.Scanner` over `io.Reader`, buffering up to 2MB per line. Filters `data:` prefixes, skips `[DONE]`, builds `StreamChunk`.

```go
scanner := bufio.NewScanner(body)
scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
```

**Porting decision**: This is **raw SSE parsing**. In `layers/http`, delegate stream parsing to provider-specific packages (e.g., `provider/openai` handles OpenAI SSE, `provider/claude` handles Claude SSE). The HTTP layer should only:
1. Peek the first chunk to decide headers
2. Forward subsequent chunks as-is (or via a translator)

---

## 10. Tool System — JSON Schema + Tool Use/Result

**Source**: `internal/providers/types.go`  
**Pattern**: Full tool system with `ToolDef` (schema + description), `ToolCall` (id, name, input), `ToolResult` (id, content, is_error).

**Porting decision**: Port **entire tool system** into `layers/provider`. Many AI providers (Claude, Gemini, OpenAI) support tools but with different JSON shapes. The `provider` package should define the canonical shape; HTTP handlers translate to/from provider-specific formats.

---

## Monorepo → Layers Mapping

| Monorepo Package | Router Layers Target | Action |
|-----------------|----------------------|--------|
| `internal/config` | `layers/config` | Simplify to YAML + env overrides |
| `internal/auth` | `layers/auth` | Port Loader, merge with DB storage |
| `internal/providers/types` | `layers/provider` | Port full interface + tool system |
| `internal/providers/registry` | `layers/provider` | Add TTL cache + health updater |
| `internal/router/router` | `layers/provider/*` | Split: HTTP → `layers/http`, client → provider |
| `internal/router/phase` | `layers/routing/phase` | Refactor to struct-based config |
| `internal/router/autocombo` | `layers/routing/autocombo` | Load from config instead of hard-code |
| `internal/management` | New binary `cmd/management` | Separate from router core |
| `internal/api/server` | `layers/http` | Adopt `ServerOption` + `BaseHandler` patterns from CLIProxyAPI |

---

## Key Differences: Ti Internal vs CLIProxyAPI-main

| Aspect | Ti Internal | CLIProxyAPI-main | Recommendation for Layers |
|--------|-------------|------------------|---------------------------|
| Config | JSON + 7-level precedence | YAML + simple defaults | YAML + env overrides |
| Auth | File-based Loader | OAuth + token refresh | Merge: Loader + DB + OAuth |
| Provider HTTP | Inline in handlers | `BaseAPIHandler` embed | Adopt CLIProxyAPI pattern |
| Registry | Simple mutex map | TTL + background updater | Adopt CLIProxyAPI pattern |
| Error format | Ad-hoc maps | `ErrorResponse` struct | Adopt `ErrorResponse` |
| Streaming | Manual scanner | Peek-first-chunk | Adopt peek-first pattern |
| Tools | Full schema | Simplified | Port Ti Internal tool system |
| Phase routing | Map-based | Not implemented | Port + struct refactor |

---

*Next step: reconcile these findings with `CLIProxyAPI_LEARNINGS.md` and produce a unified migration plan.*
