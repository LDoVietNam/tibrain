# Nghiên Cứu OAuth Authentication Patterns trong Go

> **Version**: 1.0.0  
> **Last Updated**: 2026-04-28  
> **Language**: Tiếng Việt  
> **Category**: Research  
> **Purpose**: Nghiên cứu patterns và best practices cho OAuth authentication trong Go để implement auth-agent cho Ti router

---

## Tổng Quan

Nghiên cứu này tập trung vào việc tìm hiểu OAuth authentication patterns trong Go để implement **auth-agent** cho Ti router ecosystem với các tính năng:
- OAuth flows (authorization code, device code, PKCE)
- API key management
- Token refresh
- Session management

---

## 1. Internal Research - Ti Router Implementation

### 1.1 Existing Go Implementation

**Location**: `Z:\Ti\router\layers\authentication\`

#### Files Chính:

| File | Mô Tả | Status |
|------|-------|--------|
| `auth.go` | Service interface, Token, Claims | ✅ Interface defined |
| `oauth.go` | OAuthProvider interface, OAuthRegistry, OAuthToken | ✅ Interface defined |
| `session.go` | Session, SessionStore | ✅ Interface defined |
| `health.go` | TokenHealth, HealthChecker | ✅ Interface defined |
| `keys.go` | APIKeyManager, APIKey CRUD | ✅ Implemented |
| `oauth_handlers.go` | OAuth HTTP handlers | ✅ Partially implemented |
| `providers.go` | OAuth provider implementations | ✅ Partially implemented |

#### Key Interfaces:

**Service Interface** (`auth.go`):
```go
type Service interface {
    IssueToken(ctx context.Context, subject string, scopes []string) (Token, error)
    ValidateToken(ctx context.Context, token string) (*Claims, error)
    RevokeToken(ctx context.Context, token string) error
    HealthCheck(ctx context.Context) error
}
```

**OAuthProvider Interface** (`oauth.go`):
```go
type OAuthProvider interface {
    Name() string
    AuthorizeURL(state, redirectURI string) string
    Exchange(ctx context.Context, code, redirectURI string) (*OAuthToken, error)
    Refresh(ctx context.Context, refreshToken string) (*OAuthToken, error)
}
```

**APIKeyManager** (`keys.go`):
- GenerateKey: Tạo API key mới với format `sk-<base64>`
- GetKey: Retrieve API key by ID
- ListKeys: List tất cả keys với masked values
- UpdatePermissions: Update allowed models
- DeleteKey: Xóa API key
- ValidateKey: Validate API key với rate limiting

### 1.2 TypeScript Reference Implementation

**Location**: `Z:\Ti\router\layers\authentication\nextjs\`

#### Structure:

- `auth/` - Login/logout endpoints
- `oauth/` - 34 files: OAuth provider flows (Google, Anthropic, OpenAI, etc.)
- `keys/` - API key CRUD
- `tokens/` - Token CRUD
- `sessions/` - Session management
- `token-health/` - Token health endpoints
- `lib/tokenHealthCheck.ts` - Token health logic

#### Token Refresh Logic (`tokenRefresh.ts`):

**Key Features**:
- `TOKEN_EXPIRY_BUFFER_MS`: Buffer time trước expiry (proactive refresh)
- `checkAndRefreshToken`: Check và refresh token proactively
- `refreshAccessToken`: Generic refresh function cho multiple providers
- Provider-specific refresh functions:
  - `refreshClaudeOAuthToken`
  - `refreshGoogleToken`
  - `refreshQwenToken`
  - `refreshCodexToken`
  - `refreshIflowToken`
  - `refreshGitHubToken`
  - `refreshCopilotToken`

**Proactive Refresh Pattern**:
```typescript
// Check regular token expiry
if (expiresAt - now < TOKEN_EXPIRY_BUFFER_MS) {
    log.info("TOKEN_REFRESH", "Token expiring soon, refreshing proactively");
    const newCredentials = await getAccessToken(provider, updatedCredentials);
    if (newCredentials && newCredentials.accessToken) {
        await updateProviderCredentials(updatedCredentials.connectionId, newCredentials);
    }
}
```

---

## 2. External Research - Go OAuth Libraries

### 2.1 go-oauth2/oauth2

**GitHub**: https://github.com/go-oauth2/oauth2  
**Stars**: 3.6k  
**Description**: OAuth 2.0 server library for Go

#### Features:
- Dễ sử dụng
- Based on RFC 6749 implementation
- Token storage support TTL
- Support custom expiration time
- Support custom extension fields
- Support custom scope
- Support JWT để generate access tokens

#### Store Implements:
- BuntDB (default)
- Redis
- MongoDB
- MySQL
- PostgreSQL
- DynamoDB
- XORM
- GORM
- Firestore
- Hazelcast

#### Quick Start Example:

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/go-oauth2/oauth2/v4/errors"
    "github.com/go-oauth2/oauth2/v4/manage"
    "github.com/go-oauth2/oauth2/v4/models"
    "github.com/go-oauth2/oauth2/v4/server"
    "github.com/go-oauth2/oauth2/v4/store"
)

func main() {
    manager := manage.NewDefaultManager()
    // Token memory store
    manager.MustTokenStorage(store.NewMemoryTokenStore())
    
    // Client memory store
    clientStore := store.NewClientStore()
    clientStore.Set("000000", &models.Client{
        ID:     "000000",
        Secret: "999999",
        Domain: "http://localhost",
    })
    manager.MapClientStorage(clientStore)
    
    srv := server.NewDefaultServer(manager)
    srv.SetAllowGetAccessRequest(true)
    srv.SetClientInfoHandler(server.ClientFormHandler)
    
    srv.UserAuthorizationHandler = func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
        return "000000", nil
    }
    
    http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
        err := srv.HandleAuthorizeRequest(w, r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
        }
    })
    
    http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
        srv.HandleTokenRequest(w, r)
    })
    
    log.Fatal(http.ListenAndServe(":9096", nil))
}
```

#### JWT Access Token Generation:

```go
import (
    "github.com/go-oauth2/oauth2/v4/generates"
    "github.com/dgrijalva/jwt-go"
)

// Use JWT to generate access tokens
manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte("00000000"), jwt.SigningMethodHS512))

// Parse and verify JWT access token
token, err := jwt.ParseWithClaims(access, &generates.JWTAccessClaims{}, func(t *jwt.Token) (interface{}, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("parse error")
    }
    return []byte("00000000"), nil
})

claims, ok := token.Claims.(*generates.JWTAccessClaims)
if !ok || !token.Valid {
    // Handle invalid token
}
```

---

### 2.2 ory/fosite

**GitHub**: https://github.com/ory/fosite  
**Stars**: 2.6k  
**Description**: Security-first OAuth2 & OpenID Connect framework for Go

#### Features:
- **Security-first**: Implements peer-reviewed IETF RFC6749
- **Threat mitigation**: Counterfeits weaknesses covered in RFC6819
- **Database attack scenarios**: Various database attack mitigations
- **OpenID Connect**: Complete implementation with all flows (code, implicit, hybrid)
- **Extensible**: Built simple, powerful and extensible

#### RFCs Implemented:
- OAuth 2.0 Authorization Framework (RFC6749)
- OAuth 2.0 Multiple Response Type Encoding Practices
- OAuth 2.0 Threat Model and Security Considerations (RFC6819)
- Proof Key for Code Exchange by OAuth Public Clients (PKCE)
- OAuth 2.0 for Native Apps
- OpenID Connect Core 1.0
- OAuth 2.0 Pushed Authorization Request

#### Security Features:

**No Cleartext Storage of Credentials**:
- Credentials encrypted at rest
- Never log plaintext

**Encryption of Credentials**:
- Strong encryption algorithms
- Key management

**Use Short Expiration Time**:
- Access tokens: 1-2 hours
- Refresh tokens: 30 days

**Limit Number of Usages or One-Time Usage**:
- One-time authorization codes
- Limited refresh token usage

**Bind Token to Client ID**:
- Tokens bound to specific client
- Prevents token reuse across clients

**Automatic Revocation of Derived Tokens**:
- If abuse detected, revoke all derived tokens

**Binding of Refresh Token to Client ID**:
- Refresh tokens bound to specific client
- Prevents refresh token theft

**Refresh Token Rotation**:
- New refresh token issued on each refresh
- Old refresh token invalidated

**Revocation of Refresh Tokens**:
- Manual revocation support
- Automatic revocation on abuse

**Validate Pre-Registered redirect_uri**:
- Only pre-registered URIs allowed
- Prevents open redirect attacks

**Binding of Authorization Code to Client ID**:
- Authorization codes bound to client
- Prevents code interception

**Binding of Authorization Code to redirect_uri**:
- Codes bound to specific redirect URI
- Prevents code injection

**Opaque Access Tokens**:
- Tokens not guessable
- Format: `<key>.<signature>`

**Opaque Refresh Tokens**:
- Refresh tokens not guessable
- Same format as access tokens

**Ensure Confidentiality of Requests**:
- HTTPS required (except localhost)
- TLS validation

**Use of Asymmetric Cryptography**:
- JWT signing with asymmetric keys
- Public/private key pairs

**Enforcing Random States**:
- Random state required
- OpenID Connect nonce required

**Advanced Token Validation**:
- Format: `<key>.<signature>`
- HMAC-SHA256 with global secret

Example token:
```
/tgBeUhWlAT8tM8Bhmnx+Amf8rOYOUhrDi3pGzmjP7c=.BiV/Yhma+5moTP46anxMT6cWW8gz8R5vpC9RbpwSDdM=
```

---

### 2.3 go-pkgz/auth

**GitHub**: https://github.com/go-pkgz/auth  
**Stars**: 1.3k  
**Description**: Authenticator via oauth2, direct, email and telegram

#### Features:
- OAuth2 authentication
- Direct authentication (username/password)
- Email authentication
- Telegram authentication
- JWT tokens
- Middleware support

---

### 2.4 cli/oauth

**GitHub**: https://github.com/cli/oauth  
**Stars**: 521  
**Description**: OAuth Device flow and Web application flow in Go client apps

#### Features:
- OAuth Device flow
- OAuth Web application flow
- Client-side OAuth implementation
- Suitable for CLI applications

---

### 2.5 zitadel/oidc

**GitHub**: https://github.com/zitadel/oidc  
**Stars**: 1.8k  
**Description**: OpenID Connect client and server library for Go

#### Features:
- OpenID Connect client
- OpenID Connect server
- Certified by OpenID Foundation
- Complete OIDC implementation

---

## 3. Best Practices cho OAuth Authentication

### 3.1 OAuth 2.0 Flows

#### Authorization Code Flow (Recommended cho server-side apps)

**Use Cases**:
- Web applications with server-side backend
- Mobile apps with backend
- Desktop apps with backend

**Steps**:
1. User redirected to authorization server
2. User grants permission
3. Authorization server returns authorization code
4. Client exchanges code for access token
5. Client uses access token to access resources

**Security**:
- Code is short-lived and single-use
- Token exchange happens server-to-server
- PKCE recommended cho additional security

#### Device Code Flow (Recommended cho IoT/CLI apps)

**Use Cases**:
- IoT devices
- CLI applications
- Devices without browser

**Steps**:
1. Client requests device code
2. User visits verification URL on device with browser
3. User enters user code
4. Client polls for token
5. Token issued when user completes flow

**Security**:
- User code is short-lived
- Polling interval prevents brute force
- Verification URL is trusted

#### PKCE (Proof Key for Code Exchange)

**Use Cases**:
- Mobile apps
- Single-page apps (SPAs)
- Public clients

**Steps**:
1. Client generates code verifier and code challenge
2. Client sends code challenge in authorization request
3. Authorization server binds code to challenge
4. Client sends code verifier in token request
5. Authorization server verifies challenge

**Security**:
- Prevents authorization code interception
- Code verifier never sent in authorization request
- Strong cryptographic binding

---

### 3.2 Token Refresh Strategies

#### Proactive Refresh (Recommended)

**Pattern**:
```go
const TOKEN_EXPIRY_BUFFER_MS = 5 * 60 * 1000 // 5 minutes

func checkAndRefreshToken(token *OAuthToken) (*OAuthToken, error) {
    expiresAt := token.ExpiresAt
    now := time.Now()
    
    if expiresAt.Sub(now) < TOKEN_EXPIRY_BUFFER_MS {
        // Token expiring soon, refresh proactively
        return refreshAccessToken(token.RefreshToken)
    }
    
    return token, nil
}
```

**Benefits**:
- Prevents token expiry during requests
- Seamless user experience
- No downtime

**Buffer Time**: 5-10 minutes before expiry

#### Reactive Refresh (Fallback)

**Pattern**:
```go
func makeRequestWithToken(url string, token *OAuthToken) (*http.Response, error) {
    resp, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    resp.Header.Set("Authorization", "Bearer " + token.AccessToken)
    
    httpResponse, err := http.DefaultClient.Do(resp)
    if err != nil {
        return nil, err
    }
    
    if httpResponse.StatusCode == 401 {
        // Token expired, refresh and retry
        newToken, err := refreshAccessToken(token.RefreshToken)
        if err != nil {
            return nil, err
        }
        
        // Retry with new token
        resp.Header.Set("Authorization", "Bearer " + newToken.AccessToken)
        return http.DefaultClient.Do(resp)
    }
    
    return httpResponse, nil
}
```

**Benefits**:
- Only refresh when needed
- Reduces unnecessary refresh calls

**Drawbacks**:
- May cause request failures
- User experience impact

---

### 3.3 API Key Management

#### Key Generation

**Pattern**:
```go
func GenerateAPIKey() string {
    keyBytes := make([]byte, 32)
    rand.Read(keyBytes)
    return "sk-" + base64.URLEncoding.EncodeToString(keyBytes)
}
```

**Format**: `sk-<base64-encoded-32-bytes>`

#### Key Storage

**Best Practices**:
- Hash keys before storage (bcrypt/scrypt/argon2)
- Never log plaintext keys
- Use encryption at rest
- Separate database cho credentials

**Pattern**:
```go
type APIKey struct {
    ID           string    `json:"id"`
    KeyHash      string    `json:"key_hash"` // Hashed, not plaintext
    Name         string    `json:"name"`
    MachineID    string    `json:"machine_id"`
    CreatedAt    time.Time `json:"created_at"`
    LastUsed     time.Time `json:"last_used"`
    AllowedModels []string `json:"allowed_models,omitempty"`
    IsActive     bool      `json:"is_active"`
}
```

#### Key Validation

**Pattern**:
```go
func ValidateAPIKey(keyValue string) (*APIKey, error) {
    // Hash the input key
    keyHash := hashKey(keyValue)
    
    // Lookup by hash
    key, err := db.GetAPIKeyByHash(keyHash)
    if err != nil {
        return nil, err
    }
    
    // Check if active
    if !key.IsActive {
        return nil, fmt.Errorf("key is inactive")
    }
    
    // Update last used
    key.LastUsed = time.Now()
    db.UpdateAPIKey(key)
    
    return key, nil
}
```

#### Rate Limiting per Key

**Pattern**:
```go
type RateLimiter struct {
    requests map[string]*TokenBucket
    mutex    sync.Mutex
}

func (r *RateLimiter) Allow(keyID string) bool {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    bucket, exists := r.requests[keyID]
    if !exists {
        bucket = &TokenBucket{
            Capacity: 100, // 100 requests per minute
            Tokens:    100,
            RefillRate: time.Minute,
        }
        r.requests[keyID] = bucket
    }
    
    return bucket.Consume()
}
```

---

### 3.4 Session Management

#### JWT Sessions

**Pattern**:
```go
type SessionClaims struct {
    UserID    string   `json:"user_id"`
    Scopes    []string `json:"scopes"`
    IssuedAt  int64    `json:"iat"`
    ExpiresAt int64    `json:"exp"`
    jwt.RegisteredClaims
}

func CreateSession(userID string, scopes []string) (string, error) {
    claims := SessionClaims{
        UserID:    userID,
        Scopes:    scopes,
        IssuedAt:  time.Now().Unix(),
        ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}
```

**Benefits**:
- Stateless
- No database lookup required
- Scalable

**Drawbacks**:
- Cannot revoke easily
- Token size increases with claims

#### Memory Store Sessions

**Pattern**:
```go
type MemorySessionStore struct {
    sessions map[string]*Session
    mutex    sync.RWMutex
}

type Session struct {
    ID        string
    UserID    string
    Scopes    []string
    CreatedAt time.Time
    ExpiresAt time.Time
}

func (s *MemorySessionStore) Create(userID string, scopes []string) (*Session, error) {
    session := &Session{
        ID:        generateID(),
        UserID:    userID,
        Scopes:    scopes,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.sessions[session.ID] = session
    
    return session, nil
}

func (s *MemorySessionStore) Get(id string) (*Session, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    session, exists := s.sessions[id]
    if !exists {
        return nil, fmt.Errorf("session not found")
    }
    
    if time.Now().After(session.ExpiresAt) {
        delete(s.sessions, id)
        return nil, fmt.Errorf("session expired")
    }
    
    return session, nil
}
```

**Benefits**:
- Fast lookup
- Easy to revoke
- Simple implementation

**Drawbacks**:
- Not persistent across restarts
- Memory usage grows with sessions

#### SQLite Sessions

**Pattern**:
```go
type SQLiteSessionStore struct {
    db *sql.DB
}

func (s *SQLiteSessionStore) Create(userID string, scopes []string) (*Session, error) {
    session := &Session{
        ID:        generateID(),
        UserID:    userID,
        Scopes:    scopes,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    
    _, err := s.db.Exec(`
        INSERT INTO sessions (id, user_id, scopes, created_at, expires_at)
        VALUES (?, ?, ?, ?, ?)
    `, session.ID, session.UserID, strings.Join(session.Scopes, ","), session.CreatedAt, session.ExpiresAt)
    
    return session, err
}

func (s *SQLiteSessionStore) Get(id string) (*Session, error) {
    var session Session
    var scopesStr string
    
    err := s.db.QueryRow(`
        SELECT id, user_id, scopes, created_at, expires_at
        FROM sessions WHERE id = ?
    `, id).Scan(&session.ID, &session.UserID, &scopesStr, &session.CreatedAt, &session.ExpiresAt)
    
    if err != nil {
        return nil, err
    }
    
    session.Scopes = strings.Split(scopesStr, ",")
    
    if time.Now().After(session.ExpiresAt) {
        s.Delete(session.ID)
        return nil, fmt.Errorf("session expired")
    }
    
    return &session, nil
}
```

**Benefits**:
- Persistent across restarts
- Scalable
- Easy to backup

**Drawbacks**:
- Slower than memory store
- Requires database maintenance

---

### 3.5 Multi-Tenant Authentication

#### Tenant Isolation

**Pattern**:
```go
type Tenant struct {
    ID       string
    Name     string
    APIClients []APIClient
}

type APIClient struct {
    ID     string
    TenantID string
    Secret string
    Scopes []string
}

func ValidateToken(token string) (*Claims, error) {
    claims, err := parseToken(token)
    if err != nil {
        return nil, err
    }
    
    // Validate tenant exists
    tenant, err := db.GetTenant(claims.TenantID)
    if err != nil {
        return nil, fmt.Errorf("invalid tenant")
    }
    
    // Validate client belongs to tenant
    client, err := db.GetClient(claims.ClientID)
    if err != nil || client.TenantID != tenant.ID {
        return nil, fmt.Errorf("invalid client")
    }
    
    return claims, nil
}
```

#### Per-Tenant Rate Limiting

**Pattern**:
```go
type TenantRateLimiter struct {
    limiters map[string]*RateLimiter
    mutex    sync.Mutex
}

func (t *TenantRateLimiter) Allow(tenantID string) bool {
    t.mutex.Lock()
    defer t.mutex.Unlock()
    
    limiter, exists := t.limiters[tenantID]
    if !exists {
        // Get tenant-specific rate limit
        tenant, err := db.GetTenant(tenantID)
        if err != nil {
            return false
        }
        
        limiter = &RateLimiter{
            Capacity: tenant.RateLimit,
            Tokens:    tenant.RateLimit,
            RefillRate: time.Minute,
        }
        t.limiters[tenantID] = limiter
    }
    
    return limiter.Consume()
}
```

---

## 4. Recommendations cho Ti Auth-Agent

### 4.1 Library Selection

**Recommendation**: Sử dụng **ory/fosite** cho OAuth server implementation

**Reasons**:
- Security-first với complete RFC implementation
- Extensible architecture
- Strong security features (token rotation, binding, revocation)
- Active maintenance (2.6k stars)
- OpenID Connect certified

**Alternative**: **go-oauth2/oauth2** cho simpler implementation

**Reasons**:
- Easier to use
- Good documentation
- Multiple store implementations
- JWT support built-in

### 4.2 Architecture Recommendations

#### Layer Structure:

```
auth-agent/
├── oauth/
│   ├── provider.go           // OAuthProvider interface
│   ├── registry.go           // OAuthRegistry
│   ├── providers/
│   │   ├── google.go         // Google OAuth
│   │   ├── anthropic.go      // Anthropic OAuth
│   │   ├── openai.go         // OpenAI OAuth
│   │   └── github.go         // GitHub OAuth
│   ├── flows/
│   │   ├── authorization_code.go  // Authorization code flow
│   │   ├── device_code.go        // Device code flow
│   │   └── pkce.go               // PKCE implementation
│   └── token/
│       ├── refresh.go        // Token refresh logic
│       └── validator.go      // Token validation
├── apikey/
│   ├── manager.go            // APIKeyManager
│   ├── generator.go          // Key generation
│   ├── validator.go          // Key validation
│   └── ratelimit.go          // Rate limiting per key
├── session/
│   ├── store.go              // SessionStore interface
│   ├── jwt.go                // JWT session implementation
│   ├── memory.go             // Memory store implementation
│   └── sqlite.go             // SQLite store implementation
├── middleware/
│   ├── auth.go               // Authentication middleware
│   ├── oauth.go              // OAuth middleware
│   └── ratelimit.go          // Rate limiting middleware
└── health/
    ├── checker.go            // Health checker
    └── monitor.go            // Background health monitor
```

### 4.3 Implementation Priorities

**Priority 1 (Critical)**:
1. OAuth provider flows (Google, Anthropic, OpenAI)
2. Token refresh with proactive strategy
3. API key CRUD operations
4. Basic session management

**Priority 2 (Important)**:
1. Device code flow (GitHub, Qwen)
2. PKCE implementation
3. Rate limiting per API key
4. Token health monitoring

**Priority 3 (Enhancement)**:
1. Multi-tenant support
2. Advanced session management
3. Token rotation
4. Audit logging

### 4.4 Security Recommendations

**Must Implement**:
- Hash API keys before storage (bcrypt/scrypt/argon2)
- Use HTTPS cho all OAuth endpoints
- Validate redirect URIs
- Implement PKCE cho public clients
- Use short-lived access tokens (1-2 hours)
- Implement token refresh with buffer time
- Bind tokens to client ID
- Enforce random states

**Should Implement**:
- Token rotation
- Refresh token revocation
- Rate limiting per API key
- Audit logging
- Session revocation
- Multi-factor authentication (optional)

**Nice to Have**:
- Biometric authentication
- Hardware token support
- Advanced threat detection
- Machine learning-based anomaly detection

### 4.5 Performance Recommendations

**Caching**:
- Cache validated tokens (short TTL)
- Cache provider configurations
- Cache rate limit buckets

**Database**:
- Use indexes cho token lookups
- Partition sessions by date
- Archive expired sessions

**Concurrency**:
- Use sync.RWMutex cho session store
- Use goroutine pool cho token refresh
- Implement connection pooling cho database

---

## 5. Code Examples

### 5.1 OAuth Provider Implementation (Google)

```go
package providers

import (
    "context"
    "fmt"
    "net/url"
    
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

type GoogleProvider struct {
    config *oauth2.Config
}

func NewGoogleProvider(clientID, clientSecret, redirectURI string) *GoogleProvider {
    config := &oauth2.Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        RedirectURL:  redirectURI,
        Scopes:       []string{"openid", "email", "profile"},
        Endpoint:     google.Endpoint,
    }
    
    return &GoogleProvider{config: config}
}

func (p *GoogleProvider) Name() string {
    return "google"
}

func (p *GoogleProvider) AuthorizeURL(state string) string {
    return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *GoogleProvider) Exchange(ctx context.Context, code string) (*OAuthToken, error) {
    token, err := p.config.Exchange(ctx, code)
    if err != nil {
        return nil, fmt.Errorf("failed to exchange token: %w", err)
    }
    
    return &OAuthToken{
        AccessToken:  token.AccessToken,
        RefreshToken: token.RefreshToken,
        TokenType:    token.TokenType,
        ExpiresIn:    int(token.Expiry.Seconds()),
    }, nil
}

func (p *GoogleProvider) Refresh(ctx context.Context, refreshToken string) (*OAuthToken, error) {
    token := &oauth2.Token{
        RefreshToken: refreshToken,
    }
    
    newToken, err := p.config.TokenSource(ctx, token).Token()
    if err != nil {
        return nil, fmt.Errorf("failed to refresh token: %w", err)
    }
    
    return &OAuthToken{
        AccessToken:  newToken.AccessToken,
        RefreshToken: newToken.RefreshToken,
        TokenType:    newToken.TokenType,
        ExpiresIn:    int(newToken.Expiry.Seconds()),
    }, nil
}
```

### 5.2 Token Refresh with Proactive Strategy

```go
package token

import (
    "context"
    "log"
    "sync"
    "time"
)

const TOKEN_EXPIRY_BUFFER = 5 * time.Minute

type TokenManager struct {
    tokens      map[string]*OAuthToken
    providers   map[string]OAuthProvider
    mutex       sync.RWMutex
    refreshChan chan string
}

func (tm *TokenManager) StartRefreshWorker(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            tm.checkAndRefreshAll()
        case tokenID := <-tm.refreshChan:
            tm.refreshToken(tokenID)
        }
    }
}

func (tm *TokenManager) checkAndRefreshAll() {
    tm.mutex.RLock()
    defer tm.mutex.RUnlock()
    
    for tokenID, token := range tm.tokens {
        if tm.shouldRefresh(token) {
            go tm.refreshToken(tokenID)
        }
    }
}

func (tm *TokenManager) shouldRefresh(token *OAuthToken) bool {
    expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
    return time.Until(expiresAt) < TOKEN_EXPIRY_BUFFER
}

func (tm *TokenManager) refreshToken(tokenID string) {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    
    token, exists := tm.tokens[tokenID]
    if !exists {
        return
    }
    
    provider, exists := tm.providers[token.Provider]
    if !exists {
        log.Printf("Provider not found: %s", token.Provider)
        return
    }
    
    newToken, err := provider.Refresh(context.Background(), token.RefreshToken)
    if err != nil {
        log.Printf("Failed to refresh token %s: %v", tokenID, err)
        return
    }
    
    tm.tokens[tokenID] = newToken
    log.Printf("Token %s refreshed successfully", tokenID)
}
```

### 5.3 API Key Management with Hashing

```go
package apikey

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "time"
    
    "golang.org/x/crypto/bcrypt"
)

type APIKeyManager struct {
    keys  map[string]*APIKey
    mutex sync.RWMutex
}

type APIKey struct {
    ID           string    `json:"id"`
    KeyHash      string    `json:"key_hash"`
    Name         string    `json:"name"`
    MachineID    string    `json:"machine_id"`
    CreatedAt    time.Time `json:"created_at"`
    LastUsed     time.Time `json:"last_used"`
    AllowedModels []string `json:"allowed_models,omitempty"`
    IsActive     bool      `json:"is_active"`
}

func (m *APIKeyManager) GenerateKey(name, machineID string) (*APIKey, string, error) {
    // Generate random key
    keyBytes := make([]byte, 32)
    if _, err := rand.Read(keyBytes); err != nil {
        return nil, "", fmt.Errorf("failed to generate key: %w", err)
    }
    keyValue := "sk-" + base64.URLEncoding.EncodeToString(keyBytes)
    
    // Hash the key
    keyHash, err := bcrypt.GenerateFromPassword([]byte(keyValue), bcrypt.DefaultCost)
    if err != nil {
        return nil, "", fmt.Errorf("failed to hash key: %w", err)
    }
    
    // Generate ID
    id := generateID()
    
    apiKey := &APIKey{
        ID:           id,
        KeyHash:      string(keyHash),
        Name:         name,
        MachineID:    machineID,
        CreatedAt:    time.Now(),
        LastUsed:     time.Now(),
        AllowedModels: []string{},
        IsActive:     true,
    }
    
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.keys[id] = apiKey
    
    return apiKey, keyValue, nil
}

func (m *APIKeyManager) ValidateKey(keyValue string) (*APIKey, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    for _, key := range m.keys {
        if !key.IsActive {
            continue
        }
        
        // Compare hash
        err := bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(keyValue))
        if err == nil {
            // Valid key
            key.LastUsed = time.Now()
            return key, nil
        }
    }
    
    return nil, fmt.Errorf("invalid API key")
}
```

### 5.4 Rate Limiting per API Key

```go
package ratelimit

import (
    "sync"
    "time"
)

type TokenBucket struct {
    Capacity   int
    Tokens     int
    RefillRate time.Duration
    LastRefill time.Time
}

type RateLimiter struct {
    buckets map[string]*TokenBucket
    mutex   sync.Mutex
}

func (r *RateLimiter) Allow(keyID string) bool {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    bucket, exists := r.buckets[keyID]
    if !exists {
        bucket = &TokenBucket{
            Capacity:   100, // Default: 100 requests per minute
            Tokens:     100,
            RefillRate: time.Minute,
            LastRefill: time.Now(),
        }
        r.buckets[keyID] = bucket
    }
    
    // Refill tokens
    now := time.Now()
    elapsed := now.Sub(bucket.LastRefill)
    if elapsed >= bucket.RefillRate {
        bucket.Tokens = bucket.Capacity
        bucket.LastRefill = now
    }
    
    // Check if tokens available
    if bucket.Tokens > 0 {
        bucket.Tokens--
        return true
    }
    
    return false
}

func (r *RateLimiter) SetLimit(keyID string, limit int) {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    bucket, exists := r.buckets[keyID]
    if !exists {
        bucket = &TokenBucket{
            Capacity:   limit,
            Tokens:     limit,
            RefillRate: time.Minute,
            LastRefill: time.Now(),
        }
        r.buckets[keyID] = bucket
    } else {
        bucket.Capacity = limit
        bucket.Tokens = limit
    }
}
```

---

## 6. References

### GitHub Repositories:
- **go-oauth2/oauth2**: https://github.com/go-oauth2/oauth2
- **ory/fosite**: https://github.com/ory/fosite
- **go-pkgz/auth**: https://github.com/go-pkgz/auth
- **cli/oauth**: https://github.com/cli/oauth
- **zitadel/oidc**: https://github.com/zitadel/oidc

### Ti Router Files:
- **auth.go**: Z:\Ti\router\layers\authentication\auth.go
- **oauth.go**: Z:\Ti\router\layers\authentication\oauth.go
- **keys.go**: Z:\Ti\router\layers\authentication\keys.go
- **tokenRefresh.ts**: Z:\Ti\router\layers\http\nextjs\sse\services\tokenRefresh.ts

### RFCs:
- **RFC 6749**: OAuth 2.0 Authorization Framework
- **RFC 6819**: OAuth 2.0 Threat Model and Security Considerations
- **RFC 7636**: PKCE (Proof Key for Code Exchange)
- **OpenID Connect Core 1.0**: https://openid.net/specs/openid-connect-core-1_0.html

---

*Last Updated: 2026-04-28*
*Research completed by: Claude*
*Purpose: Implement auth-agent cho Ti router ecosystem*
