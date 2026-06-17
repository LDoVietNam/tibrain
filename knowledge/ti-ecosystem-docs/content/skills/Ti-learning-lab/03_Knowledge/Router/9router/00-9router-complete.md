---
tags: ["tibrain", "router", "authentication", "documentation", "skill"]
scopes: ["auth", "tibrain"]
last_updated: 2026-05-22
---
# 9Router Complete Research Notes

> Repo: `Z:\Ti\Ti-learning-lab\05_Repositories\router\9router`
> Date: 2026-04-28
> Status: Consolidated from 5 deep-dive sections

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Account Selection & Fallback Engine](#2-account-selection--fallback-engine)
3. [Combo System](#3-combo-system)
4. [OAuth & Token Refresh](#4-oauth--token-refresh)
5. [Format Translation & SSE Streaming](#5-format-translation--sse-streaming)
6. [Proxy Layer](#6-proxy-layer)
7. [Usage Tracking](#7-usage-tracking)
8. [Key Lessons for Ti Router](#8-key-lessons-for-ti-router)

---

## 1. Architecture Overview

- **Stack**: Next.js 15 app router, TypeScript/JS transpiled, Node.js runtime
- **Port**: 20128 (dashboard + API same port)
- **Database**: LowDB (JSON file `DATA_DIR/db.json`, NOT SQLite). Separate `usage.json` for tracking
- **Core proxy logic**: Lives in `open-sse` git submodule (separate package). 9router is a **thin orchestration layer**
- **Concurrency**: `proper-lockfile` + custom `LocalMutex` queue. Cloud mode skips file locking

### Data Model (db.json)

```
db.json
├── providerConnections[]   # OAuth/API key accounts per provider
├── providerNodes[]         # Custom OpenAI/Anthropic-compatible endpoints
├── proxyPools[]            # HTTP or Vercel relay proxies
├── combos[]                # Named model fallback chains
├── modelAliases{}          # "alias" → "provider/model"
├── customModels[]          # Custom model definitions
├── apiKeys[]               # API keys for client auth to 9router
├── settings{}              # Global config
└── pricing{}               # Provider pricing cache
```

### Default Settings
- `stickyRoundRobinLimit: 3`
- `comboStrategy: "fallback"`
- `providerStrategies: {}` (per-provider overrides)
- `requireLogin: true`
- `observabilityEnabled: true`

---

## 2. Account Selection & Fallback Engine

### 2.1 Three-Tier Fallback Architecture

```
Tier 1: Combo fallback   (combo.js)     — Model A → Model B → Model C
Tier 2: Account fallback (auth.js)      — Account 1 → Account 2 (same provider)
Tier 3: Error cooldown   (accountFallback.js) — Exponential backoff + model locks
```

### 2.2 Account Selection Mutex (auth.js)

```javascript
let selectionMutex = Promise.resolve();

// Each request:
const currentMutex = selectionMutex;
selectionMutex = new Promise(resolve => { resolveMutex = resolve; });
try {
  await currentMutex;
  // ... select account ...
} finally {
  resolveMutex();
}
```

**Why**: Prevents race condition where 2 parallel requests both pick the same "current" account in round-robin.

### 2.3 Virtual Connections for Free Providers

```javascript
if (FREE_PROVIDERS[providerId]?.noAuth) {
  return { id: "noauth", connectionName: "Public", isActive: true, accessToken: "public" };
}
```

Free providers get a synthetic connection. `markAccountUnavailable()` skips `connectionId === "noauth"`.

### 2.4 Account Filtering

```javascript
const availableConnections = connections.filter(c => {
  if (excludeSet.has(c.id)) return false;        // Already tried in loop
  if (isModelLockActive(c, model)) return false;  // Model-level lock
  return true;
});
```

### 2.5 Selection Strategies

| Strategy | Behavior |
|----------|----------|
| **fill-first** (default) | Priority-sorted, pick first available |
| **round-robin** | Sticky: use current `stickyLimit` times (default 3), then switch to LRU |

**Per-provider override**: `settings.providerStrategies[providerId].fallbackStrategy`

### 2.6 Sticky Round-Robin Detail

```
1. Sort by lastUsedAt desc → find "current" candidate
2. If current.consecutiveUseCount < stickyLimit:
   → Continue using current, increment count
3. If reached limit:
   → Sort by lastUsedAt asc (LRU), pick oldest
   → Reset count = 1
```

### 2.7 markAccountUnavailable() Flow

```
1. If upstream returns resetsAtMs (quota reset time):
   → Use resetsAtMs as cooldown (skip exponential backoff)
   → shouldFallback = true
2. If no resetsAtMs:
   → checkFallbackError(status, errorText, backoffLevel)
   → Get cooldownMs + newBackoffLevel
3. Create model lock: buildModelLockUpdate(model, cooldownMs)
4. Update DB: testStatus="unavailable", lastError, errorCode, lastErrorAt, backoffLevel
5. Return { shouldFallback: true, cooldownMs }
```

### 2.8 clearAccountError() Logic

```
1. Find all keys starting with "modelLock_"
2. Clear: model lock for the success model + all expired locks
3. If NO active locks remain:
   → Reset: testStatus="active", lastError=null, backoffLevel=0
4. If SOME locks still active (other models locked):
   → Keep testStatus="unavailable"
```

**Key insight**: Only reset to active when account is COMPLETELY clean.

### 2.9 Exponential Backoff

```javascript
const level = Math.max(0, backoffLevel - 1);
const cooldown = BACKOFF_CONFIG.base * Math.pow(2, level);
return Math.min(cooldown, BACKOFF_CONFIG.max);  // max 4 min
```

Each fail → `backoffLevel++` → cooldown doubles.

### 2.10 Model-Level Locking

```
connection.modelLock_claude-sonnet-4-20250514 = "2026-04-28T03:00:00Z"
connection.modelLock___all = "2026-04-28T03:00:00Z"  // account-level
```

**isModelLockActive(connection, model)**:
- Check `modelLock_${model}` first
- Fall back to `modelLock___all`
- Compare expiry with `Date.now()`

**Advantage**: Account can be rate-limited on model A but still serve model B.

---

## 3. Combo System

### 3.1 Combo Schema

```javascript
combos: [
  { name: "if/kimi-k2-thinking", models: ["claude/sonnet", "openai/gpt-4o", "gemini/2.5-pro"] }
]
```

### 3.2 Combo Strategies

| Strategy | Behavior |
|----------|----------|
| **fallback** (default) | Try model 0 → 1 → 2 in order |
| **round-robin** | Rotate array per request via `comboRotationState` Map |

### 3.3 handleComboChat Flow

```javascript
for (let i = 0; i < rotatedModels.length; i++) {
  result = await handleSingleModel(body, modelStr);
  if (result.ok) return result;  // Success

  checkFallbackError(status, errorText);
  if (!shouldFallback) return result;  // Fatal error

  // SPECIAL: transient 502/503/504 + short cooldown → WAIT
  if (cooldownMs > 0 && cooldownMs <= 5000 && transientStatus) {
    await sleep(cooldownMs);
  }
  // Continue → try next model
}
```

All failed → **503 + Retry-After** (if earliestRetryAfter available).

### 3.4 Combo Rotation (In-Memory)

```javascript
const comboRotationState = new Map();  // Not persisted

function getRotatedModels(models, comboName, strategy) {
  const currentIndex = comboRotationState.get(comboName) || 0;
  comboRotationState.set(comboName, (currentIndex + 1) % models.length);
  return rotatedModels;
}
```

Reset on combo update/delete: `resetComboRotation(comboName)`.

### 3.5 Comparison: Ti Router vs 9router Combo

| Feature | 9router | Ti Router |
|---------|---------|-----------|
| Strategies | `fallback`, `round-robin` | `priority`, `round-robin`, `random`, `least-used` |
| Weighted random | ❌ | ✅ |
| DAG validation | ❌ | ✅ (prevent circular refs) |
| Auto combo (task-aware) | ❌ | ✅ `auto:<taskType>` |
| Pre-check availability | ❌ | ✅ skip cooldown models |
| Combo model format | String array | `{ model, weight }[]` or string |

**Ti combo is MORE advanced.** 9router only adds: **transient wait** (502/503/504 + cooldown <=5s → wait before trying next model).

---

## 4. OAuth & Token Refresh

### 4.1 PKCE Utilities

```javascript
generatePKCE() => {
  codeVerifier:  base64url(randomBytes(32)),   // 43 chars
  codeChallenge: base64url(sha256(verifier)),  // S256 method
  state:         base64url(randomBytes(32))    // CSRF
}
```

### 4.2 OAuth Flow Types

| Type | Providers | Characteristics |
|------|-----------|----------------|
| **PKCE Authorization Code** | Claude, iFlow, Antigravity, Gemini, Codex, Cursor, Cline, KiloCode, GitLab | No client_secret needed |
| **Standard OAuth2** | Gemini, Antigravity (Google) | Requires `client_secret` |
| **Device Code** | GitHub Copilot, Qwen | Headless, poll with `device_code` |

### 4.3 Device Code Flow Detail (GitHub Copilot)

```
1. getDeviceCode() → { device_code, user_code, verification_uri }
2. Print: "Visit URL, enter code"
3. open(browser) automatically
4. pollAccessToken() → while(true):
   - authorization_pending → continue
   - slow_down → interval += 5000ms
   - expired_token → throw
   - access_denied → throw
   - access_token → success
```

### 4.4 Custom Flows

- **Qoder**: Device Token Flow with `deviceTokenUrl` + `deviceRefreshUrl` + `statusUrl`
- **Kiro**: Custom flow with platform enum (`darwin arm64=2`, `win32=5`)
- **GitHub Copilot**: 2-step (GitHub device code → get Copilot token)

### 4.5 Token Refresh Pattern

```javascript
// Refresh when: expiresAt - now < TOKEN_EXPIRY_BUFFER_MS (5 min)
fetch(refreshUrl, {
  method: "POST",
  body: URLSearchParams({
    grant_type: "refresh_token",
    refresh_token,
    client_id,
    client_secret  // Optional (Claude doesn't need)
  })
})

// Fallback: keep old refreshToken if provider doesn't return new one
return { accessToken, refreshToken: tokens.refresh_token || refreshToken, expiresIn };
```

### 4.6 Per-Provider Refresh Specializations

| Provider | Format | Notes |
|----------|--------|-------|
| Claude | JSON body | No client_secret |
| Google | Form-urlencoded | Requires client_secret |
| Qwen | Form-urlencoded | Standard |
| GitHub | Device flow | No refresh token (re-auth when expired) |
| Copilot | 2-step | GitHub token → Copilot token (separate expiry) |

### 4.7 JWT Account Identification

```javascript
function decodeJwtPayload(jwt) {
  const parts = jwt.split(".");
  const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
  const padded = base64 + "=".repeat((4 - (base64.length % 4)) % 4);
  return JSON.parse(Buffer.from(padded, "base64").toString("utf8"));
}

// Identity priority: email → preferred_username → sub
return payload.email || payload.preferred_username || payload.sub;
```

### 4.8 OAuthService Base Class

4 core methods:
- `buildAuthUrl()` — PKCE params + `state`
- `startAuthFlow()` — Local server, 5-min timeout, poll callback
- `exchangeCode()` — POST token endpoint (JSON or form)
- `authenticate()` — Full flow orchestration

**State validation**: Callback state MUST match generated state (CSRF protection).

---

## 5. Format Translation & SSE Streaming

### 5.1 Format Detection

| Endpoint | Format |
|----------|--------|
| `/v1/responses` | `OPENAI_RESPONSES` |
| `/v1/messages` | `CLAUDE` |
| `/v1/chat/completions` + `body.input[]` | `OPENAI` (Cursor CLI hack) |

**By body**: `body.messages[]` → OpenAI, `body.contents[]` → Gemini, `body.input[]` → OpenAI Responses

### 5.2 Hub-and-Spoke Translation

```
Source Format → OpenAI Format → Target Format
```

```javascript
// Request:
if (sourceFormat !== "openai") {
  result = requestRegistry.get(`${sourceFormat}:openai`)(model, result, stream, credentials);
}
if (targetFormat !== "openai") {
  result = requestRegistry.get(`openai:${targetFormat}`)(model, result, stream, credentials);
}

// Response: reverse direction
```

### 5.3 Pre-Translation Transformations

| Step | Function | Purpose |
|------|----------|---------|
| 1 | `compressMessages()` (RTK) | Compress tool_result before translation |
| 2 | `stripContentTypes()` | Remove image/audio if provider doesn't support |
| 3 | `normalizeThinkingConfig()` | Remove thinking config if last message ≠ user |
| 4 | `ensureToolCallIds()` | Ensure tool_calls have id |
| 5 | `fixMissingToolResponses()` | Insert empty tool_result if missing |

### 5.4 Native Passthrough

```javascript
const clientTool = detectClientTool(headers, body);
const passthrough = isNativePassthrough(clientTool, provider);

if (passthrough) {
  // Same ecosystem — skip ALL translation
  translatedBody = { ...body, model };  // Only swap model + token
}
```

**Reduces latency and avoids translation bugs** when CLI tool and provider match (e.g., Claude Code → Claude API).

### 5.5 Tool Cloaking (Anti-Ban)

**Claude OAuth** (`sk-ant-oat` token):
```javascript
if (provider === "claude" && apiKey?.includes("sk-ant-oat")) {
  const { body: cloakedBody, toolNameMap } = cloakClaudeTools(result);
  result._toolNameMap = toolNameMap;  // For uncloaking response
}
```
- Rename tools with `_cc` suffix
- Prevents Anthropic detection

**Antigravity**:
```javascript
if (provider === "antigravity" && body.userAgent !== "antigravity") {
  const { cloakedBody, toolNameMap } = AntigravityExecutor.cloakTools(result);
}
```
- Rename tools + inject decoys
- Skip if client is native AG

### 5.6 Streaming vs Non-Streaming

```javascript
// Client wants streaming?
const clientRequestedStreaming = body.stream === true || 
  sourceFormat === FORMATS.ANTIGRAVITY || 
  sourceFormat === FORMATS.GEMINI;

// Provider REQUIRES streaming?
const providerRequiresStreaming = provider === "openai" || provider === "codex";

// Default
let stream = providerRequiresStreaming ? true : (body.stream !== false);

// FIX: AI SDK sends Accept: application/json but wants non-streaming
const clientPrefersJson = acceptHeader.includes("application/json");
const clientPrefersSSE = acceptHeader.includes("text/event-stream");
if (clientPrefersJson && !clientPrefersSSE && body.stream !== true) {
  stream = false;
}
```

### 5.7 Provider Thinking Override

```javascript
if (providerThinking?.mode && providerThinking.mode !== "auto") {
  const mode = providerThinking.mode;  // "on", "off", "low", "medium", "high"
  
  if (mode === "on" && !body.thinking) {
    body = { ...body, thinking: { type: "enabled", budget_tokens: 10000 } };
  } else if (mode === "off" && !body.thinking) {
    body = { ...body, thinking: { type: "disabled" } };
  } else if (!body.reasoning_effort) {
    body = { ...body, reasoning_effort: mode };
  }
}
```

User can override thinking config at provider level (dashboard setting), only if client hasn't set.

### 5.8 HTTP Status Constants

```javascript
HTTP_STATUS = {
  BAD_REQUEST: 400, UNAUTHORIZED: 401, PAYMENT_REQUIRED: 402,
  FORBIDDEN: 403, NOT_FOUND: 404, NOT_ACCEPTABLE: 406,
  REQUEST_TIMEOUT: 408, RATE_LIMITED: 429,
  SERVER_ERROR: 500, BAD_GATEWAY: 502, 
  SERVICE_UNAVAILABLE: 503, GATEWAY_TIMEOUT: 504
}
```

**Retry config**: 502→3 retries/3000ms, 503→3 retries/2000ms, 504→2 retries/3000ms. 429→0 retries (handled by backoff/lock).

---

## 6. Proxy Layer

### 6.1 Connection Proxy (Per-Connection)

3 sources (fallback priority):

| Priority | Source | Type |
|----------|--------|------|
| 1 | `proxyPoolId` in connection | Pool (HTTP or Vercel) |
| 2 | Legacy `connectionProxyUrl` | Legacy HTTP proxy |
| 3 | None | Direct |

**Vercel relay** (`type="vercel"`): Rewrite base URL instead of using HTTP_PROXY env. App-level proxy.

```javascript
if (proxyPool.type === "vercel") {
  return {
    source: "vercel",
    vercelRelayUrl: proxyUrl,  // Rewrite URL
    connectionProxyEnabled: false,  // Don't use HTTP_PROXY
  };
}
```

### 6.2 Outbound Proxy (Global)

Sets env vars: `HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY`, `NO_PROXY`.

**Managed env flag** (`NINE_ROUTER_PROXY_MANAGED`): Only clears env vars that 9router created, never touches user-set vars.

```javascript
if (!enabled && process.env.NINE_ROUTER_PROXY_MANAGED === "1") {
  delete process.env.HTTP_PROXY;  // Only if we created it
}
```

### 6.3 Proxy Types

| Type | Mechanism |
|------|-----------|
| `http` | Standard HTTP proxy via env |
| `vercel` | Rewrite base URL to Vercel endpoint |

---

## 7. Usage Tracking

### 7.1 Database

- **Local**: LowDB JSON (`usage.json`)
- **Cloud**: In-memory only
- **Migration**: One-time `history[]` → `dailySummary{}`

### 7.2 Pending Request Tracking (In-Memory)

```javascript
global._pendingRequests = { byModel: {}, byAccount: {} };
```

| Metric | Key | Tracking |
|--------|-----|----------|
| byModel | `${model} (${provider})` | Increment on START, decrement on END |
| byAccount | `${connectionId} → ${modelKey}` | Same |

**Safety timer**: Auto-clear after **60 seconds** if END never called (crash/disconnect).

```javascript
const timer = setTimeout(() => {
  pendingRequests.byModel[modelKey] = 0;
  pendingRequests.byAccount[connectionId][modelKey] = 0;
}, 60000);
```

### 7.3 Stats Emitter

```javascript
global._statsEmitter = new EventEmitter();
global._statsEmitter.setMaxListeners(50);
```

Emits `"pending"` events for real-time dashboard updates.

### 7.4 Error Provider Tracking

```javascript
global._lastErrorProvider = { provider: "", ts: 0 };
// Auto-clear after 10 seconds
```

Used for UI edge coloring (highlight provider currently erroring).

### 7.5 Active Requests Query

```javascript
getActiveRequests() => {
  activeRequests: [{ model, provider, account, count }],
  recentRequests: [{ timestamp, model, provider, tokens, status }],
  errorProvider: ""
}
```

**Recent requests dedup**: Key = `${model}|${provider}|${promptTokens}|${completionTokens}|${minute}`. Prevents duplicate entries from retries.

### 7.6 Save Request Usage

```javascript
saveRequestUsage(entry) {
  entry.timestamp = new Date().toISOString();
  db.data.history.push(entry);
  db.data.totalRequestsLifetime++;
  aggregateToDailySummary(dailySummary, entry);
}
```

---

## 8. Key Lessons for Ti Router

### From Account Selection
1. **Mutex is mandatory** for multi-account + round-robin (Promise chain FIFO)
2. **Model-level lock >> account-level lock** — account can serve model B while locked on model A
3. **`resetsAtMs` from upstream** (quota reset) is more valuable than exponential backoff
4. **Sticky round-robin** (use N times then switch) reduces connection churn for streaming
5. **503 + Retry-After** is correct response when all models/accounts unavailable
6. **Config-driven error rules** (ERROR_RULES array) allow tuning without code deploy

### From Combo System
7. **Ti combo is already more advanced** than 9router (weighted random, DAG validation, auto combo)
8. **One thing to add**: transient wait for 502/503/504 with short cooldown (<=5s)

### From OAuth
9. **PKCE is standard for CLI tools** — no client_secret needed
10. **Device Code Flow** for headless/server without browser (GitHub Copilot)
11. **State validation mandatory** — callback state MUST match
12. **Per-provider refresh specialization** — each has different API format
13. **Keep old refreshToken** if provider doesn't return new one
14. **JWT decode from access_token** for identity (email/username/sub)
15. **GitHub Copilot 2-step**: GitHub device code → Copilot token (separate expiry)

### From Translation
16. **Native passthrough** reduces latency when CLI and provider match ecosystem
17. **Accept header check** — AI SDK sends `Accept: application/json` but uses `/v1/chat/completions`
18. **Provider thinking override** — useful for enterprise deployment
19. **Tool cloaking** — if supporting OAuth/free tier providers with anti-tool policies

### From Proxy
20. **Vercel relay pattern** — app-level proxy (rewrite URL) vs network-level (HTTP_PROXY)
21. **Managed env flag** — only clear env vars you created

### From Usage Tracking
22. **In-memory pending + safety timer** — track real-time active requests, auto-recover on crash
23. **EventEmitter for dashboard** — real-time updates without polling
24. **Dedup recent requests** — prevent history spam from retries
25. **Separate usage DB** — reduces lock contention vs config DB

## 9. Per-Provider Executor Tricks

### Qwen (Device Code Flow)
- **Static fingerprint matching**: Full `X-Stainless-*` headers to mimic Qwen Code CLI
- **Token shard binding**: `resource_url` from OAuth → must use exact host or 401/403
- **Thinking + tool_choice conflict**: Auto-sanitize `tool_choice="required"` → `"auto"` when thinking active
- **Inject empty system message**: `[{role:"system",content:[{type:"text",text:"",cache_control:{type:"ephemeral"}}]}]`

### Grok Web (SSO Cookie)
- **Fake browser fingerprint**: Full Chrome headers (`Sec-Ch-Ua`, `Sec-Fetch-*`, `traceparent`)
- **`x-statsig-id`**: Fake base64 TypeError message to mimic real client
- **Model mode mapping**: `grok-4.1-fast` → `MODEL_MODE_FAST` (internal enum)
- **SSO cookie strip**: Remove `"sso="` prefix if user pasted full cookie value
- **Thinking extraction**: `modelResponse.message` yielded as `reasoning_content` before main content

### Perplexity Web (Session Cache)
- **FNV-1a hash session key**: `history.map(h=>`${h.role}:${h.content}`).join("\n")` hashed for lookup
- **Session continuation**: `backendUuid` from previous response → `last_backend_uuid` for follow-up
- **Response cleaning**: Strip `[<d>]` citations, XML declarations, Grok tags, multi-spaces
- **Tool hints injection**: Convert OpenAI tools list → text instructions ("reference only, cannot invoke")
- **Query length cap**: 96000 chars, slice from end if exceeded

### Kiro (AWS EventStream)
- **Binary EventStream → SSE**: Custom `parseEventFrame()` with `DataView`, prelude+headers+payload+CRC
- **Multi-event finish**: Wait for BOTH `meteringEvent` AND `contextUsageEvent` before emitting final chunk
- **Token estimation fallback**: If no usage events, estimate: output = content/4, input = contextUsagePercentage * 200000 / 100
- **Tool call accumulation**: Track `seenToolIds` Map across multiple `toolUseEvent` frames
- **TransformStream pattern**: Avoid Workers timeout (comment says "instead of ReadableStream.pull()")

### Qoder (HMAC-SHA256)
- **Dynamic signature per request**: `HMAC-SHA256(apiKey, "UserAgent:sessionID:timestamp")`
- **3 required headers**: `session-id`, `x-qoder-timestamp`, `x-qoder-signature`
- **Without signature**: Returns 406 error

### Codex (OpenAI OAuth)
- **Force model**: Always `"o4-mini"` regardless of user request
- **Force max_tokens**: 32000 if not set (Codex requires)
- **Override temperature**: Always 0 (deterministic for coding)
- **Special endpoint**: `/responses` not `/chat/completions`

### Cursor (Authentication Code)
- **Workos OAuth**: `client_id=client_...`, scopes `openid profile email offline_access cursor:user`
- **Token URL**: `https://api2.cursor.com/auth/token` with Basic auth
- **User info**: `https://api2.cursor.com/auth/me` → `sub`, `email`

### Gemini CLI
- **Force endpoint**: Always `/v1beta/openai/chat/completions` (OpenAI-compatible mode)
- **Override model**: Strip `-thinking` suffix, force `gemini-2.5-pro-exp-03-25`
- **Thinking extraction**: `geminiThinking` field mapped to OpenAI `reasoning_content`
- **Usage extraction**: `geminiUsage` → `prompt_tokens`, `completion_tokens`
