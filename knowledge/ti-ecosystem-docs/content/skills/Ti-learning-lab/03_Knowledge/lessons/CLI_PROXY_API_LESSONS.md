# CLIProxyAPI-main Deep-Dive Learning Notes

> **Date**: 2026-04-26  
> **Scope**: Deep-dive patterns from `Z:\10_WORKPLACE\CLIProxyAPI-main` to adopt/adapt into `Z:\Ti\router\layers`  
> **Status**: In-depth analysis — implementation roadmap below

---

## 1. Functional Options Pattern (Server Construction)

**Source**: `internal/api/server.go`  
**Pattern**: `ServerOption` closures applied variadically in `NewServer(..., opts ...ServerOption)`. This removes config-struct bloat and lets tests inject middleware or engine tweaks without changing constructor signatures.

```go
type ServerOption func(*serverOptionConfig)
func WithMiddleware(mw gin.HandlerFunc) ServerOption { ... }
func WithEngineConfigurator(fn func(*gin.Engine)) ServerOption { ... }
func WithRequestLoggerFactory(fn func() logging.RequestLogger) ServerOption { ... }
```

**Key insight**: The `serverOptionConfig` is an unexported struct, so external callers cannot construct invalid states. Default values are set inside `NewServer` before options are applied.

**Ti Router Gap**: `http` package server init is imperative and takes a large config struct. Adopt `ServerOption` to make `NewServer` test-friendly and reduce arity.

---

## 2. Base Handler Embed (Delegation Over Inheritance)

**Source**: `sdk/api/handlers/handlers.go` + `sdk/api/handlers/openai/openai_handlers.go`  
**Pattern**: `BaseAPIHandler` owns cross-cutting concerns: client pooling, auth manager, request ID tracing, SSE keep-alive, upstream header forwarding, and common error encoding. Provider-specific handlers embed it and only override JSON translation + routing.

```go
type BaseAPIHandler struct { clientPool *ClientPool; authMgr *auth.Manager; ... }
type OpenAIAPIHandler struct { *BaseAPIHandler }
```

**Key insight**: `BaseAPIHandler` exposes `ExecuteWithAuthManager`, `ExecuteStreamWithAuthManager`, `GetContextWithCancel`, `WriteErrorResponse`, `WriteUpstreamHeaders`. Sub-handlers never touch `net/http` directly for these concerns.

**Ti Router Gap**: `provider` defines a `Provider` interface but HTTP handlers duplicate SSE streaming, error encoding, and auth checks. Create a `BaseHandler` in `http` and embed it into `WindsurfHandler`, `ClaudeHandler`, etc.

---

## 3. Model Registry — TTL Cache + Reference Counted Refresh

**Source**: `internal/registry/model_registry.go`  
**Pattern**: A central `ModelRegistry` holds `[]ModelInfo` with capabilities, cost tiers, context windows. A background `Updater` goroutine refreshes the list from provider APIs at a configurable interval.

- `GetAvailableModels(filter string)` returns filtered metadata.
- `RegisterUpdater(interval)` starts a background ticker.
- Reference-counted TTL: if no caller has fetched models recently, the updater pauses to save requests.

**Key insight**: Model metadata is **static configuration + dynamic enrichment**. Static data (model IDs, context windows) lives in YAML; dynamic data (availability, pricing) is fetched at runtime and cached.

**Ti Router Gap**: `provider.Registry` only tracks active providers. Add a `ModelInfo` slice to each provider and expose `GetAvailableModels("providerName")` so the OpenAI-compatible `/v1/models` endpoint can be served without hard-coding model lists.

---

## 4. Modular Handler Packages — One Package Per API Shape

**Source**: `sdk/api/handlers/{openai,claude,gemini}/`  
**Pattern**: Each directory is a Go package named after the upstream API format (not the provider). This separates:
- **Business logic** (auth, model list, chat) → `provider/*`
- **HTTP translation** (OpenAI-compatible JSON, SSE, error shapes) → `sdk/api/handlers/openai/`
- **Provider-specific translation** (Claude tool_use → OpenAI tool_calls) → `sdk/translator/claude/`

**Key insight**: The `openai` handler package serves OpenAI-compatible clients. It does not care whether the upstream is Gemini, Claude, or OpenAI itself — it only cares about JSON shapes.

**Ti Router Gap**: `provider` and `http` are tightly coupled. Separate into:
- `provider/*` — pure Go interfaces (no gin, no HTTP)
- `http/openai` — OpenAI-compatible HTTP handlers
- `http/internal` — shared SSE, error encoding, middleware

---

## 5. Standardised ErrorResponse — Predictable Client UX

**Source**: `sdk/api/handlers/handlers.go`  
**Pattern**: All handlers return the same `ErrorResponse` JSON shape. OpenAI SDKs, curl, and LangChain expect this exact structure.

```go
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}
type ErrorDetail struct {
    Message string `json:"message"`
    Type    string `json:"type"`   // e.g. "invalid_request_error", "authentication_error"
    Code    string `json:"code,omitempty"`
}
```

**Key insight**: `WriteErrorResponse(c *gin.Context, err *interfaces.ErrorMessage)` is a method on `BaseAPIHandler`. It maps internal error types to OpenAI-compatible `Type` strings automatically.

**Ti Router Gap**: Handlers return ad-hoc `map[string]interface{}`. Replace with `ErrorResponse` so clients get predictable errors.

---

## 6. Constants Package + Dot Import

**Source**: `internal/constant/constant.go`  
**Pattern**: All magic strings (provider names, header names, routing phases) live in one package. Handler packages use dot import so constants read like locals.

```go
import . "github.com/router-for-me/CLIProxyAPI/v6/internal/constant"
// Now OpenAI, Claude, Gemini are in scope without prefix
```

**Key insight**: Dot import is safe here because the constant package has no functions — only string constants. It eliminates stutter (`constant.OpenAI` → `OpenAI`).

**Ti Router Gap**: Hard-coded strings like `"claude"`, `"windsurf"`, `"x-ti-request-id"` are scattered. Extract a `layers/constant` package and dot-import in `http` and `provider`.

---

## 7. Unexported Struct Context Keys

**Source**: `sdk/api/handlers/handlers.go`  
**Pattern**: Context keys are unexported empty structs, preventing collision with any string key from another package.

```go
type pinnedAuthContextKey struct{}
func WithPinnedAuthID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, pinnedAuthContextKey{}, id)
}
```

**Key insight**: This is Go's idiomatic way to avoid context key collisions. String keys are an anti-pattern in production code.

**Ti Router Gap**: Audit `context.WithValue(..., "key", value)` calls and replace with struct keys.

---

## 8. YAML Config with Defaults + Validation

**Source**: `internal/config/config.go`  
**Pattern**: Config is a nested struct tagged with `yaml:"key"`. Defaults are set in a `DefaultConfig()` constructor. Validation runs at boot time and panics on missing required fields.

```go
type Config struct {
    Server  ServerConfig  `yaml:"server"`
    Auth    AuthConfig    `yaml:"auth"`
    Logging LoggingConfig `yaml:"logging"`
}
func DefaultConfig() *Config { ... }
```

**Key insight**: Config is loaded once at startup. Hot-reload is not implemented — instead, a `SIGHUP` handler could reload in the future.

**Ti Router Gap**: `config.Config` structure is unclear. Document fields in `docs/03-operations/01-deployment.md`, add `DefaultConfig()`, and validate required fields in `Bootstrap()`.

---

## 9. Auth Abstraction — OAuth + Token Lifecycle

**Source**: `sdk/cliproxy/auth/` + `internal/access/`  
**Pattern**: Two-layer auth:
- `auth.Manager` — OAuth2 flows, token refresh, API key validation, token storage.
- `access.Manager` — Per-request allow/deny: IP allowlist, API key rate-limit, key existence check.

**Key insight**: `auth.Manager` is a long-lived singleton. `access.Manager` is created per-request (or per-middleware invocation). Separation of concerns means you can test rate-limiting without mocking OAuth servers.

**Ti Router Gap**: Auth logic is inline in HTTP handlers. Extract:
- `layers/auth` — token refresh, OAuth flows
- `layers/access` — rate-limit, IP checks

---

## 10. Logging — Interface + Lumberjack Rotation

**Source**: `internal/logging/`  
**Pattern**: `RequestLogger` is an interface. Production uses a file-backed `lumberjack.Logger`; tests inject a `bytes.Buffer` capture.

```go
type RequestLogger interface {
    LogRequest(url, method string, requestHeaders map[string][]string, ...)
    LogStreamingRequest(...) io.WriteCloser
}
```

**Key insight**: `sync.Once` guards global logger setup. `LogFormatter` (custom `logrus.Formatter`) controls field order so logs are grep-friendly.

**Ti Router Gap**: No structured logging exists. Add `layers/logging` with a `RequestLogger` interface and wire it through `ServerOption`.

---

## 11. Streaming Pattern — Peek First Chunk Before Headers

**Source**: `sdk/api/handlers/openai/openai_handlers.go` `handleStreamingResponse()`  
**Pattern**: SSE responses must set headers **after** the first upstream chunk is received. If the upstream errors immediately, the handler returns a JSON error (not SSE) with the correct HTTP status code.

```go
for {
    select {
    case errMsg := <-errChan:
        h.WriteErrorResponse(c, errMsg) // JSON, not SSE
        return
    case chunk := <-dataChan:
        setSSEHeaders()                 // Now safe to commit
        flusher.Flush()
        h.handleStreamResult(...)       // Continue streaming
        return
    }
}
```

**Key insight**: This prevents "200 OK then JSON error inside SSE" — a common bug that breaks OpenAI SDKs.

**Ti Router Gap**: Current SSE handlers likely set headers immediately. Adopt peek-first-chunk pattern.

---

## 12. Non-Streaming Keep-Alive

**Source**: `sdk/api/handlers/handlers.go` `StartNonStreamingKeepAlive()`  
**Pattern**: Some clients (e.g., curl with `--no-buffer`) time out if no bytes are received within N seconds. `BaseAPIHandler` starts a goroutine that sends periodic whitespace bytes until the real response arrives.

**Ti Router Gap**: Not implemented. Add `StartNonStreamingKeepAlive(c, ctx)` to `BaseHandler` for long upstream requests.

---

## 13. JSON Manipulation — gjson + sjson

**Source**: `sdk/api/handlers/openai/openai_handlers.go`  
**Pattern**: `gjson` parses JSON without unmarshalling into structs. `sjson` modifies JSON bytes in-place. Used for:
- Extracting `model` name from raw request bytes
- Converting OpenAI Responses format → Chat Completions format
- Filtering model metadata fields

**Key insight**: Zero-allocation JSON manipulation is critical for high-throughput proxies. Only unmarshal into structs when full validation is needed.

**Ti Router Gap**: Adopt `gjson`/`sjson` for request shape detection and light transformation; keep struct-based unmarshalling for business logic.

---

## 14. Request Format Conversion — Responses ↔ ChatCompletions

**Source**: `sdk/api/handlers/openai/openai_handlers.go`  
**Pattern**: Some clients send OpenAI Responses-format payloads to `/v1/chat/completions`. The handler detects this (`shouldTreatAsResponsesFormat`) and converts the request before forwarding to upstream.

**Key insight**: Conversion happens **before** routing to the translator. The translator only sees valid Chat Completions shapes.

**Ti Router Gap**: Add `shouldTreatAsResponsesFormat` helper and conversion logic in OpenAI-compatible handlers.

---

## 15. Background Updater Goroutine with Graceful Stop

**Source**: `internal/registry/model_updater.go`  
**Pattern**: Background goroutines use a `stopChan` + `sync.WaitGroup`. On shutdown, `Stop()` closes the channel and waits for `wg.Done()`.

```go
type Updater struct { stopChan chan struct{}; wg sync.WaitGroup }
func (u *Updater) Start() { u.wg.Add(1); go u.loop() }
func (u *Updater) Stop()  { close(u.stopChan); u.wg.Wait() }
```

**Ti Router Gap**: Any background goroutines (health checks, token refresh) should use this pattern for clean shutdown.

---

## Immediate Action Items (Priority Order)

| # | Action | Effort | Impact |
|---|--------|--------|--------|
| 1 | Adopt `ServerOption` functional options in `http` | Medium | Testability |
| 2 | Create `http.BaseHandler` with SSE, error, auth helpers | High | DRY handlers |
| 3 | Extract `layers/constant` package + dot import | Low | Readability |
| 4 | Standardise `ErrorResponse` across all handlers | Low | Client UX |
| 5 | Add `ModelInfo` + TTL registry to `provider` | Medium | Dynamic models |
| 6 | Separate HTTP handlers from provider logic | High | Architecture |
| 7 | Adopt unexported struct context keys | Low | Safety |
| 8 | Add `layers/logging` interface + `lumberjack` | Medium | Observability |
| 9 | Implement peek-first-chunk SSE pattern | Medium | Reliability |
| 10 | Add `auth.Manager` + `access.Manager` split | High | Security |

---

*Next review: after `go build ./...` is green and `provider` package refactor is complete.*
