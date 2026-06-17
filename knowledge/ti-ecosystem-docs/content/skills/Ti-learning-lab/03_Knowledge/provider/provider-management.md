---
tags: ["tibrain", "provider", "authentication", "documentation", "skill"]
scopes: ["auth", "tibrain"]
last_updated: 2026-05-22
---
# Provider Management Knowledge

**Date:** 2026-04-28
**Topic:** IDE AI Provider Integration & Management

---

## Core Concepts

### 1. IDE AI Providers as Router Providers

**Pattern:** IDE AI (Cursor, Windsurf, Codex, Kiro) đều có LLM API endpoints có thể tích hợp làm provider cho router

**Why:** Mở rộng provider ecosystem cho router, không chỉ giới hạn ở OpenAI, Anthropic, Gemini

**Integration Pattern:**
- OAuth flow với PKCE (Proof Key for Code Exchange)
- State verification để prevent CSRF
- Callback server để handle OAuth callback
- Token exchange với authorization code
- Session management với HttpOnly cookies

---

## Architecture Components

### 1. IDE Authenticator Interface

**File:** `ide_auth.go`

**Purpose:** Interface chuẩn cho tất cả IDE authenticators

**Methods:**
```go
type IDEAuthenticator interface {
    GetName() string
    GetAuthURL() (string, string, error) // URL, state, error
    ExchangeToken(code string) (*TokenResponse, error)
    RefreshToken() (*TokenResponse, error)
    GetRefreshLeadDuration() time.Duration
}
```

**Implementations:**
- `cursor_auth.go` - Cursor OAuth
- `windsurf_auth.go` - Windsurf OAuth
- `codex_auth.go` - Codex OAuth
- `kiro_auth.go` - Kiro OAuth

---

### 2. OAuth Flow Pattern

**Steps:**
```
1. Generate PKCE codes (code_verifier, code_challenge)
2. Generate random state
3. Start callback server (localhost:8080)
4. Open browser with auth URL
5. User approves → callback
6. Verify state (prevent CSRF)
7. Exchange code for tokens
8. Stop callback server
```

**Security:**
- PKCE prevents authorization code interception
- State verification prevents CSRF
- HttpOnly cookies prevent XSS
- HMAC signing prevents token tampering

---

### 3. Provider Registry Pattern

**File:** `provider_registry.go`

**Purpose:** Dynamic provider loading/unloading without code changes

**Features:**
- Config-based registration (JSON)
- Health check tự động (5 min interval)
- Provider status tracking (healthy/unhealthy/unknown)
- Latency measurement
- Config save/load persistence

**Config Structure:**
```json
{
  "id": "cursor",
  "name": "Cursor",
  "type": "oauth",
  "base_url": "https://api.cursor.sh",
  "enabled": true,
  "health_check": true,
  "models": ["gpt-4", "gpt-3.5-turbo"]
}
```

---

### 4. Unified API Endpoint

**File:** `unified_api.go`

**Purpose:** Single endpoint cho tất cả providers

**Endpoints:**
- `/api/unified/chat` - Chat completion (POST)
- `/api/unified/models` - Models list (GET)
- `/api/unified/health` - Health check (GET)

**Benefits:**
- Client không cần biết provider-specific endpoints
- Easy provider switching
- Unified error handling

---

### 5. Token Refresh Automation

**File:** `token_refresh.go`

**Purpose:** Auto-refresh tokens trước khi expire

**Pattern:**
```
Get RefreshLeadDuration from each authenticator
Background worker check interval (1 hour)
Refresh callback per provider
Manual refresh trigger available
```

**Why:** Prevent token expiration, maintain continuous service

---

### 6. Provider Fallback Mechanism

**File:** `provider_fallback.go`

**Purpose:** Auto-switch khi provider fail

**Strategies:**
- **Sequential:** Try providers one by one
- **Parallel:** Try all simultaneously, return first success
- **Random:** Random provider selection
- **Health-aware:** Skip unhealthy providers

**Pattern:**
```
Preferred provider → Check health → Execute → Fail → Next provider → ...
```

---

### 7. Management UI

**Files:**
- `data/html/provider_management.html` - HTML UI
- `handlers_provider_ui.go` - UI handlers

**Features:**
- Dashboard statistics (Total, Healthy, Unhealthy)
- Provider cards with status
- Register/Health Check/Remove actions
- Add provider modal
- Activity log with auto-refresh
- Config save/load

**UI Endpoints:**
- `/api/providers/ui` - Serve UI
- `/api/providers/list` - List providers
- `/api/providers/add` - Add provider
- `/api/providers/remove` - Remove provider
- `/api/providers/config/save` - Save config
- `/api/providers/config/load` - Load config

---

## Best Practices

### 1. Modular Design
- Separate files cho mỗi component
- Interface-based design
- Global managers (globalIDEAuthManager, globalProviderRegistry)

### 2. Error Handling
- Always verify state trong OAuth callback
- Timeout contexts cho network requests
- Graceful degradation khi provider unavailable

### 3. Security
- PKCE cho OAuth flows
- HttpOnly cookies cho sessions
- HMAC signing cho session tokens
- State verification prevent CSRF

### 4. Performance
- Background workers cho health checks
- Parallel fallback execution
- Caching provider configs
- Lazy initialization

### 5. Observability
- Latency tracking
- Status monitoring
- Activity logging
- Health check metrics

---

## Common Pitfalls

### 1. Unused Imports
- Remove unused crypto imports khi dùng shared helpers
- Use shared functions (generateRandomStringShared, sha256HashShared)

### 2. Type Mismatches
- sha256Hash trả về []byte nhưng buildAuthURL cần string
- Fix: Base64 encode hash result

### 3. Missing Dependencies
- os import cho file operations
- log import cho logging
- context import cho timeout handling

---

## Environment Variables Required

```
CURSOR_CLIENT_ID
CURSOR_CLIENT_SECRET
WINDSURF_CLIENT_ID
WINDSURF_CLIENT_SECRET
CODEX_CLIENT_ID
CODEX_CLIENT_SECRET
KIRO_CLIENT_ID
KIRO_CLIENT_SECRET
```

---

## Integration with Auto-Register

**Future Enhancement:** Integrate với auto-register repos
- cursor-auto-register
- windsurf-auto-register
- kiro-auto-register

**Pattern:**
```
Auto-register → Extract API key → Add to provider registry → Test health check
```

---

## Key Files Reference

| File | Purpose |
|------|---------|
| `ide_auth.go` | IDE authenticator interface |
| `cursor_auth.go` | Cursor OAuth authenticator |
| `windsurf_auth.go` | Windsurf OAuth authenticator |
| `codex_auth.go` | Codex OAuth authenticator |
| `kiro_auth.go` | Kiro OAuth authenticator |
| `oauth_server.go` | OAuth callback server helper |
| `oauth_flow.go` | OAuth flow manager |
| `session_manager.go` | Session management |
| `auth_middleware.go` | Authentication middleware |
| `provider_registry.go` | Dynamic provider loading |
| `unified_api.go` | Unified API endpoint |
| `token_refresh.go` | Token refresh automation |
| `provider_fallback.go` | Fallback mechanism |
| `data/html/provider_management.html` | Management UI |
| `handlers_provider_ui.go` | UI handlers |

---

## Testing Checklist

- [ ] OAuth flow test cho mỗi provider
- [ ] Health check tự động
- [ ] Token refresh tự động
- [ ] Fallback mechanism
- [ ] UI functionality
- [ ] Config save/load
- [ ] Concurrent requests
- [ ] Error handling
