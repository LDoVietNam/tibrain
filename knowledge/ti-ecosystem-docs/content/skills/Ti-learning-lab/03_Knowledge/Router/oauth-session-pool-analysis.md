# OAuth Session Pool & MCP Tool Routing - Knowledge Analysis

> **Date**: 2026-04-28  
> **Purpose**: Tổng hợp kiến thức đã học về OAuth session pool, usage tracking, MCP integration  
> **Status**: Research Phase

---

## 📚 Kiến Thức Đã Học (Ti-learning-lab)

### 1. Account Selection & Fallback (9Router)
**File**: `Z:\Ti\Ti-learning-lab\03_Knowledge\Router\9router\01-account-selection-fallback.md`

**Key Features**:
- ✅ Multi-account per provider (connections array)
- ✅ Account selection strategies (fill-first, round-robin, sticky round-robin)
- ✅ Model-level locking (`modelLock_${model}`)
- ✅ Exponential backoff với config-driven error rules
- ✅ Precise quota reset (resetsAtMs từ upstream)
- ✅ Mutex bảo vệ chọn account (Promise chain)
- ✅ Virtual no-auth connection cho free providers

**Lessons**:
1. Mutex là bắt buộc nếu có multi-account + round-robin (race condition)
2. Model-level lock > account-level lock (account có thể rate limit trên model expensive nhưng vẫn chạy model cheap)
3. resetsAtMs từ upstream quý hơn exponential backoff
4. Sticky round-robin (dùng N lần rồi đổi) tốt hơn pure round-robin cho streaming
5. 503 + Retry-After là response đúng khi all models/accounts unavailable

---

### 2. OAuth Token Refresh (9Router)
**File**: `Z:\Ti\Ti-learning-lab\03_Knowledge\Router\9router\03-oauth-token-refresh.md`

**Key Features**:
- ✅ PKCE Authorization Code Flow (Claude, iFlow, Antigravity, Gemini, Codex, Cursor, Cline, KiloCode, GitLab)
- ✅ Standard OAuth2 với client_secret (Gemini, Antigravity)
- ✅ Device Code Flow (GitHub Copilot, Qwen)
- ✅ Token refresh service (5 phút buffer)
- ✅ Per-provider refresh specialization
- ✅ JWT decode từ access_token để lấy email/identity

**Lessons**:
1. PKCE là chuẩn cho CLI tools (không cần client_secret)
2. Device Code Flow cho headless/server không có browser
3. State validation bắt buộc (chống CSRF)
4. Per-provider refresh specialization (mỗi provider có API format khác nhau)
5. Giữ refreshToken cũ nếu provider không trả về mới
6. JWT decode từ access_token để lấy email/identity

---

### 3. Per-Key Rate Tracking (FreeLLMAPI → Ti Router)
**File**: `Z:\Ti\Ti-learning-lab\03_Knowledge\Router\freellmapi\per-key-rate-tracking.md`

**Status**: ✅ Đã implement trong Ti Router

**Key Features**:
- ✅ Sliding window algorithm
- ✅ 4 metrics: RPM, RPD, TPM, TPD
- ✅ Per-key tracking (platform:modelId:keyId:type)
- ✅ Cooldown khi gặp 429
- ✅ Real-time status reporting
- ✅ Thread-safe với sync.RWMutex

**Implementation**:
- `Z:\Ti\router\layers\authentication\rate_tracker.go`
- `Z:\Ti\router\layers\authentication\rate_tracker_test.go`

---

### 4. Sticky Sessions (FreeLLMAPI → Ti Router)
**File**: `Z:\Ti\Ti-learning-lab\03_Knowledge\Router\freellmapi\sticky-sessions.md`

**Status**: ✅ Đã implement trong Ti Router

**Key Features**:
- ✅ Session key based on first user message (SHA-256 hash)
- ✅ TTL 30 phút
- ✅ Max 500 entries (auto-cleanup)
- ✅ Multi-turn detection (has assistant messages)
- ✅ Thread-safe với sync.RWMutex

**Implementation**:
- `Z:\Ti\router\layers\authentication\sticky_session.go`
- `Z:\Ti\router\layers\authentication\sticky_session_test.go`

---

### 5. Enhanced Analytics (FreeLLMAPI → Ti Router)
**File**: `Z:\Ti\Ti-learning-lab\03_Knowledge\Router\freellmapi\enhanced-analytics.md`

**Status**: ⚠️ Đã implement core, cần DB migration

**Key Features**:
- ✅ Summary stats (total requests, success rate, tokens, latency, cost)
- ✅ Stats by model
- ✅ Stats by platform
- ✅ Timeline data (hourly/daily)
- ✅ Error distribution
- ✅ Recent errors (last 50)
- ✅ Error categorization
- ✅ Cost estimation

**Implementation**:
- `Z:\Ti\router\layers\authentication\analytics.go`
- `Z:\Ti\router\layers\authentication\analytics_test.go`

**TODO**: DB migration để thêm `status`, `latency_ms`, `error` fields vào `usage_logs` table

---

## 🆚 Ti Router Hiện Tại vs Kiến Thức Đã Học

| Feature | 9Router (Learned) | Ti Router (Current) | Gap |
|---------|------------------|---------------------|-----|
| **OAuth Flow** | ✅ PKCE, Standard OAuth2, Device Code | ✅ OAuth providers, handlers, refresh | ❌ Device Code Flow chưa có |
| **Multi-account per provider** | ✅ connections[] array | ❌ GetCredentialsByProvider có thể lấy nhiều, nhưng không có selection logic | ❌ Cần selection strategy |
| **Account selection strategy** | ✅ fill-first, round-robin, sticky RR | ❌ Không có | ❌ Cần implement |
| **Model-level locking** | ✅ modelLock_${model} | ❌ Không có | ❌ Cần implement |
| **Exponential backoff** | ✅ Config-driven error rules | ❌ LoadBalancer có round-robin nhưng không có backoff | ❌ Cần implement |
| **Precise quota reset** | ✅ resetsAtMs từ upstream | ❌ Không có | ❌ Cần implement |
| **Mutex chọn account** | ✅ Promise chain | ❌ Không có (single-threaded selection) | ❌ Cần implement |
| **Per-key rate tracking** | N/A (9Router không có) | ✅ RateTracker (RPM/RPD/TPM/TPD) | ✅ Đã có |
| **Sticky sessions** | N/A (9Router không có) | ✅ StickySessionManager | ✅ Đã có |
| **Usage tracking per OAuth session** | ❌ 9Router tracking per connection (account) | ❌ OAuthCredential không có usage fields | ❌ Cần implement |
| **Session pool management** | ✅ connections[] array + selection | ❌ OAuthCredential store có nhiều, nhưng không có pool logic | ❌ Cần implement |

---

## 🎯 MCP Tool Routing Analysis

### Ti Router Hiện Tại
**Files**:
- `Z:\Ti\router\layers\http\mcp\mcp.go` - Basic MCPHandler (placeholder)
- `Z:\Ti\router\cmd\router-mcp\main.go` - Router MCP server (separate service)

**Status**:
- ❌ MCP handler có placeholder implementation
- ❌ Không có MCP client management
- ❌ Không có MCP tool routing thực sự
- ❌ Router-MCP service là separate service, không integrated với OAuth sessions

### Kiến Thức Đã Học
**Không có tài liệu cụ thể về MCP tool routing trong Ti-learning-lab**

---

## 📋 Gap Analysis Summary

### OAuth Session Pool (Cần Implement)

**Missing Features**:
1. ❌ OAuthCredential struct thiếu usage tracking fields
   - Current: `ID, Provider, AccessToken, RefreshToken, ExpiresAt, IsActive, LastHealthCheckAt, HealthCheckInterval`
   - Need: `DailyRequests, MonthlyRequests, DailyTokens, MonthlyTokens, Limits`

2. ❌ Không có OAuth session pool manager
   - Need: SessionPool struct với selection logic (fill-first, round-robin, sticky RR)
   - Need: Mutex bảo vệ selection (sync.Mutex)

3. ❌ Không có session rotation khi hết limit
   - Need: Check usage limits trước khi select session
   - Need: Fallback to next session khi current session hit limit
   - Need: Model-level locking (modelLock_${model})

4. ❌ Không có exponential backoff
   - Need: Config-driven error rules
   - Need: Cooldown khi gặp rate limit
   - Need: Precise quota reset (resetsAtMs từ upstream)

### MCP Tool Routing (Cần Implement)

**Missing Features**:
1. ❌ MCP client management
   - Need: MCP client registry
   - Need: MCP connection pool
   - Need: MCP authentication (OAuth token injection)

2. ❌ MCP tool routing
   - Need: Tool discovery từ MCP servers
   - Need: Tool routing logic (select MCP server cho tool)
   - Need: Tool call execution với OAuth authentication

3. ❌ Integration với OAuth sessions
   - Need: Route MCP tool calls đến appropriate OAuth session
   - Need: Usage tracking per MCP tool
   - Need: Session rotation cho MCP calls

---

## 🔧 Implementation Plan

### Phase 1: OAuth Session Pool (Priority 1)

**1.1 Extend OAuthCredential struct**
```go
type OAuthCredential struct {
    ID                string
    Provider          string
    AccessToken       string
    RefreshToken      string
    ExpiresAt         time.Time
    IsActive          bool
    LastHealthCheckAt time.Time
    HealthCheckInterval time.Duration
    
    // New fields for usage tracking
    DailyRequests     int
    MonthlyRequests   int
    DailyTokens       int
    MonthlyTokens     int
    DailyLimit        *int  // null = unlimited
    MonthlyLimit      *int  // null = unlimited
    TokenLimit        *int  // null = unlimited
    LastResetDate     time.Time
}
```

**1.2 Create SessionPool manager**
```go
type SessionPool struct {
    credentials *InMemoryCredentialStore
    rateTracker *RateTracker
    mu          sync.Mutex
}

type SelectionStrategy string

const (
    FillFirst    SelectionStrategy = "fill-first"
    RoundRobin   SelectionStrategy = "round-robin"
    StickyRoundRobin SelectionStrategy = "sticky-round-robin"
)

func (sp *SessionPool) SelectSession(ctx context.Context, provider string, model string, strategy SelectionStrategy) (*OAuthCredential, error)
func (sp *SessionPool) RecordUsage(ctx context.Context, credentialID string, requests int, tokens int)
func (sp *SessionPool) MarkUnavailable(ctx context.Context, credentialID string, model string, cooldown time.Duration)
func (sp *SessionPool) ClearError(ctx context.Context, credentialID string, model string)
```

**1.3 Model-level locking**
```go
type ModelLock struct {
    CredentialID string
    Model        string
    ExpiresAt    time.Time
}

func (sp *SessionPool) IsModelLocked(credentialID string, model string) bool
func (sp *SessionPool) SetModelLock(credentialID string, model string, cooldown time.Duration)
func (sp *SessionPool) ClearModelLock(credentialID string, model string)
```

**1.4 Exponential backoff**
```go
type ErrorRule struct {
    ErrorText    string
    StatusCode  int
    CooldownMs   int
    BackoffLevel int
}

func CheckFallbackError(statusCode int, errorText string, backoffLevel int) (shouldFallback bool, cooldownMs time.Duration, newBackoffLevel int)
```

**1.5 Integration với router**
- Update `cmd/routerd/main.go` để init SessionPool
- Update `layers/http/openai/handlers.go` để sử dụng SessionPool
- Update `layers/internal/router/router.go` để integrate session selection

### Phase 2: MCP Tool Routing (Priority 2)

**2.1 MCP client management**
```go
type MCPClient struct {
    ID          string
    Name        string
    Endpoint    string
    AuthToken   string // OAuth token
    Tools       []MCPTool
    IsActive    bool
    LastHealthCheckAt time.Time
}

type MCPClientManager struct {
    clients map[string]*MCPClient
    mu      sync.RWMutex
}

func (m *MCPClientManager) RegisterClient(client *MCPClient) error
func (m *MCPClientManager) GetClient(id string) (*MCPClient, error)
func (m *MCPClientManager) DiscoverTools(clientID string) ([]MCPTool, error)
```

**2.2 MCP tool routing**
```go
type MCPToolRouter struct {
    clientManager *MCPClientManager
    sessionPool   *SessionPool
}

func (r *MCPToolRouter) RouteToolCall(toolName string, provider string) (*MCPClient, *OAuthCredential, error)
func (r *MCPToolRouter) ExecuteToolCall(client *MCPClient, credential *OAuthCredential, toolName string, args interface{}) (interface{}, error)
```

**2.3 Integration với OAuth sessions**
- Route MCP tool calls đến appropriate OAuth session
- Usage tracking per MCP tool
- Session rotation cho MCP calls

---

## 📊 Cost Comparison (4 Options)

| Option | Description | Cost (Year 1) | Complexity |
|--------|-------------|---------------|------------|
| **Option 1** | Add to Ti Router | $584-880 | Medium |
| **Option 2** | Replace Router | $1,168-1,760 | High |
| **Option 3** | Separate Service | $880-1,320 | High |
| **Option 4** | Plugin System | $440-660 | Medium-High |

**Recommendation**: Option 1 (Add to Ti Router) - lowest cost, leverage existing OAuth infrastructure

---

## 📝 References

### 9Router (JavaScript)
- `Z:\Ti\Ti-learning-lab\05_Repositories\router\9router\`
- Account selection: `src/sse/services/auth.js`
- Token refresh: `open-sse/services/tokenRefresh.js`
- Account fallback: `open-sse/services/accountFallback.js`
- Combo fallback: `open-sse/services/combo.js`

### FreeLLMAPI (TypeScript)
- `Z:\Ti\Ti-learning-lab\05_Repositories\router\freellmapi-main\`
- Rate limiting: `server/src/services/ratelimit.ts`
- Sticky sessions: `server/src/routes/proxy.ts`
- Analytics: `server/src/routes/analytics.ts`

### Ti Router (Go)
- `Z:\Ti\router\`
- OAuth: `layers/authentication/`
- Rate tracking: `layers/authentication/rate_tracker.go`
- Sticky sessions: `layers/authentication/sticky_session.go`
- Analytics: `layers/authentication/analytics.go`
- MCP: `layers/http/mcp/mcp.go`

---

## ✅ Next Steps

1. **Phase 1.1**: Extend OAuthCredential struct
2. **Phase 1.2**: Create SessionPool manager
3. **Phase 1.3**: Implement model-level locking
4. **Phase 1.4**: Implement exponential backoff
5. **Phase 1.5**: Integration với router
6. **Phase 2.1**: MCP client management
7. **Phase 2.2**: MCP tool routing
8. **Phase 2.3**: Integration với OAuth sessions

---

**Ngày tạo**: 2026-04-28  
**Agent**: Claude Code  
**Project**: Ti Router
