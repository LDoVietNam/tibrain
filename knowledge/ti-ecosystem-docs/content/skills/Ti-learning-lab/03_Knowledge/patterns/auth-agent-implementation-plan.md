---
tags: ["tibrain", "authentication", "security", "documentation", "go"]
scopes: ["cli", "auth", "tibrain"]
last_updated: 2026-05-22
---
# Kế Hoạch Triển Khai Auth-Agent cho Ti Router

> **Version**: 1.0.0  
> **Last Updated**: 2026-04-29  
> **Language**: Tiếng Việt  
> **Purpose**: Kế hoạch chi tiết để implement auth-agent cho Ti router ecosystem

---

## Tổng Quan

Auth-agent sẽ được implement để cung cấp:
- OAuth flows (authorization code, device code, PKCE)
- API key management (generation, validation, rate limiting)
- Token refresh (proactive strategy)
- Session management

---

## 1. Research Summary

### 1.1 Internal Research - Existing Ti Implementation

**Files Đã Phân Tích**:
- `Z:\Ti\router\layers\authentication\auth.go` - Service interface, Token, Claims
- `Z:\Ti\router\layers\authentication\oauth.go` - OAuthProvider interface, OAuthRegistry
- `Z:\Ti\router\layers\authentication\keys.go` - APIKeyManager (đã implemented)
- `Z:\Ti\router\layers\http\nextjs\sse\services\tokenRefresh.ts` - Token refresh logic (TypeScript reference)

**Status**:
- ✅ Interfaces đã được defined
- ✅ APIKeyManager đã implemented (keys.go)
- ⏳ OAuth provider implementations partial
- ⏳ Token refresh chưa port sang Go

### 1.2 External Research - GitHub Repos

**Các Repos Đã Nghiên Cứ**:

| Repo | Stars | Features | Sử Dụng |
|------|-------|----------|---------|
| **ory/fosite** | 2.6k | Security-first, complete RFC implementation, OpenID Connect | Framework enterprise-grade |
| **go-oauth2/oauth2** | 3.6k | Simple, widely used, JWT support | Framework đơn giản |
| **oauth2go** | 2 | Small, embeddable, zero dependencies, PKCE support | Embedded server |
| **go-oauth2-example** | 3 | Simple example using golang.org/x/oauth2 | Reference pattern |

### 1.3 Key Patterns Đã Học

**OAuth Provider Pattern**:
```go
type OAuthProvider interface {
    Name() string
    AuthorizeURL(state, redirectURI string) string
    Exchange(ctx context.Context, code, redirectURI string) (*OAuthToken, error)
    Refresh(ctx context.Context, refreshToken string) (*OAuthToken, error)
}
```

**Token Refresh Pattern**:
- Proactive refresh với 5-minute buffer
- Background goroutine để check và refresh
- Update credentials in database sau khi refresh

**API Key Management Pattern**:
- Hash keys với bcrypt/scrypt/argon2
- Rate limiting per key
- Mask keys khi display
- Validate keys với hash comparison

---

## 2. Implementation Plan

### Phase 1: OAuth Provider Implementations (Priority 1)

**Files**: `Z:\Ti\router\layers\authentication\oauth\providers\`

#### 1.1 Google OAuth Provider
- **File**: `google.go`
- **Features**: Authorization code flow, PKCE, token refresh
- **Reference**: TypeScript implementation trong `nextjs/oauth/providers/google.ts`

#### 1.2 Anthropic OAuth Provider
- **File**: `anthropic.go`
- **Features**: Authorization code flow, PKCE, token refresh
- **Reference**: TypeScript implementation trong `nextjs/oauth/providers/claude.ts`

#### 1.3 OpenAI OAuth Provider
- **File**: `openai.go`
- **Features**: Authorization code flow, token refresh
- **Reference**: TypeScript implementation trong `nextjs/oauth/services/openai.ts`

#### 1.4 GitHub OAuth Provider
- **File**: `github.go`
- **Features**: Device code flow, token refresh
- **Reference**: TypeScript implementation trong `nextjs/oauth/providers/github.ts`

#### 1.5 Qwen OAuth Provider
- **File**: `qwen.go`
- **Features**: Device code flow, PKCE, token refresh
- **Reference**: TypeScript implementation trong `nextjs/oauth/providers/qwen.ts`

### Phase 2: Token Refresh Service (Priority 1)

**File**: `Z:\Ti\router\layers\authentication\token\refresh.go`

**Features**:
- Background goroutine để check và refresh tokens
- Proactive refresh với 5-minute buffer
- Provider-specific refresh functions
- Update credentials in database
- Error handling và retry logic

**Pattern**:
```go
type TokenRefreshService struct {
    providers map[string]OAuthProvider
    credentials *CredentialStore
    mutex     sync.RWMutex
    ticker    *time.Ticker
}

func (s *TokenRefreshService) Start(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            s.checkAndRefreshAll()
        }
    }
}

func (s *TokenRefreshService) checkAndRefreshAll() {
    // Check all credentials và refresh nếu cần
}
```

### Phase 3: OAuth HTTP Handlers (Priority 2)

**Files**: `Z:\Ti\router\layers\authentication\oauth\handlers\`

#### 3.1 Authorization Endpoint
- **File**: `authorize.go`
- **Endpoint**: `/oauth/authorize`
- **Features**: Generate authorize URL, handle callback, exchange code

#### 3.2 Token Endpoint
- **File**: `token.go`
- **Endpoint**: `/oauth/token`
- **Features**: Exchange code for token, refresh token, client credentials

#### 3.3 Provider Import Endpoints
- **File**: `import.go`
- **Endpoint**: `/oauth/import/{provider}`
- **Features**: Import tokens from provider (Google, Anthropic, etc.)

### Phase 4: Session Management Enhancement (Priority 2)

**File**: `Z:\Ti\router\layers\authentication\session\sqlite.go`

**Features**:
- SQLite session store
- Session expiration
- Session validation
- Session cleanup

### Phase 5: API Key Management Enhancement (Priority 2)

**File**: `Z:\Ti\router\layers\authentication\keys\enhanced.go`

**Features**:
- Enhanced rate limiting
- Per-key model permissions
- Key usage analytics
- Key rotation support

---

## 3. Architecture

```
auth-agent/
├── oauth/
│   ├── provider.go           // OAuthProvider interface
│   ├── registry.go           // OAuthRegistry
│   ├── providers/
│   │   ├── google.go         // Google OAuth implementation
│   │   ├── anthropic.go      // Anthropic OAuth implementation
│   │   ├── openai.go         // OpenAI OAuth implementation
│   │   ├── github.go         // GitHub OAuth implementation
│   │   └── qwen.go           // Qwen OAuth implementation
│   ├── flows/
│   │   ├── authorization_code.go  // Authorization code flow
│   │   ├── device_code.go        // Device code flow
│   │   └── pkce.go               // PKCE implementation
│   └── token/
│       ├── refresh.go        // Token refresh service
│       └── validator.go      // Token validation
├── apikey/
│   ├── manager.go            // APIKeyManager (existing)
│   ├── generator.go          // Key generation
│   ├── validator.go          // Key validation
│   └── ratelimit.go          // Rate limiting per key
├── session/
│   ├── store.go              // SessionStore interface
│   ├── memory.go             // Memory store (existing)
│   └── sqlite.go             // SQLite store (enhanced)
├── handlers/
│   ├── oauth.go              // OAuth HTTP handlers
│   └── keys.go               // API key HTTP handlers
└── middleware/
    ├── auth.go               // Authentication middleware
    └── ratelimit.go          // Rate limiting middleware
```

---

## 4. Implementation Steps

### Step 1: Implement Google OAuth Provider
- File: `Z:\Ti\router\layers\authentication\oauth\providers\google.go`
- Use `golang.org/x/oauth2` library
- Implement authorization code flow with PKCE
- Implement token refresh

### Step 2: Implement Anthropic OAuth Provider
- File: `Z:\Ti\router\layers\authentication\oauth\providers\anthropic.go`
- Use `golang.org/x/oauth2` library
- Implement authorization code flow with PKCE
- Implement token refresh

### Step 3: Implement Token Refresh Service
- File: `Z:\Ti\router\layers\authentication\token\refresh.go`
- Background goroutine với ticker
- Proactive refresh với 5-minute buffer
- Provider-specific refresh functions

### Step 4: Implement OAuth HTTP Handlers
- File: `Z:\Ti\router\layers\authentication\oauth\handlers\authorize.go`
- File: `Z:\Ti\router\layers\authentication\oauth\handlers\token.go`
- Implement `/oauth/authorize` endpoint
- Implement `/oauth/token` endpoint

### Step 5: Enhance Session Management
- File: `Z:\Ti\router\layers\authentication\session\sqlite.go`
- Implement SQLite session store
- Add session expiration
- Add session cleanup

### Step 6: Test với Sub Agent Review
- Sử dụng sub agent để review code
- Fix issues từ review
- Ensure thread-safety
- Test compilation

### Step 7: Documentation
- Document implementation trong tiếng Việt
- Save đến `Z:\Ti\Ti-learning-lab\03_Knowledge\patterns\`
- Include architecture, usage, best practices

---

## 5. Dependencies

### Required Go Libraries:
- `golang.org/x/oauth2` - OAuth2 client
- `golang.org/x/oauth2/clientcredentials` - Client credentials flow
- `golang.org/x/crypto` - Crypto functions cho key hashing
- `github.com/golang-jwt/jwt` - JWT token generation/validation (optional)

### Ti Internal Dependencies:
- `Z:\Ti\router\layers\authentication\keys.go` - APIKeyManager (existing)
- `Z:\Ti\router\layers\authentication\session.go` - SessionStore interface (existing)
- `Z:\Ti\router\data\settings.db` - SQLite database cho credentials

---

## 6. Success Criteria

Auth-agent được coi là **successful** khi:
- ✅ All OAuth providers implemented (Google, Anthropic, OpenAI, GitHub, Qwen)
- ✅ Token refresh service running in background
- ✅ OAuth HTTP handlers working
- ✅ Session management enhanced với SQLite
- ✅ API key management enhanced với rate limiting
- ✅ All code compiles successfully (`go build ./...`)
- ✅ Sub agent review passed
- ✅ Documentation created trong tiếng Việt
- ✅ Beads log updated

---

*Last Updated: 2026-04-29*
*Created by: Claude*
*Purpose: Implementation plan cho auth-agent trong Ti router ecosystem*
