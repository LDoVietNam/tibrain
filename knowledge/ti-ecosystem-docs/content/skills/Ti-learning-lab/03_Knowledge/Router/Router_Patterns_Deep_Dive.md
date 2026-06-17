# Router Patterns Deep Dive - Learnings from 8 Repositories

## Overview
This document summarizes key patterns and learnings from analyzing 8 router repositories and their application to Ti Router.

## Repositories Analyzed
1. WebAI-to-API
2. Claude-API
3. Gemini-API
4. auth2api
5. perplexity-ai
6. sydney.py
7. CLIProxyAPI
8. 9router

---

## 1. WebAI-to-API

### Key Patterns
- **Cross-platform cookie extraction** from multiple browsers (Chrome, Firefox, Edge, Safari)
- **Auto-extraction fallback** when cookies are missing or invalid
- **Browser-specific cookie paths** and formats

### Applied to Ti Router
- Created `browser_cookie_extractor.go` with multi-browser support
- Integrated into `cookie_manager.go` for automatic cookie extraction
- Supports Windows, macOS, Linux platforms

### Implementation Details
```go
// Cross-platform cookie extraction
func ExtractCookies(provider string) (map[string]string, error) {
    browsers := []string{"chrome", "firefox", "edge", "safari"}
    for _, browser := range browsers {
        cookies, err := ExtractFromBrowser(browser, provider)
        if err == nil && len(cookies) > 0 {
            return cookies, nil
        }
    }
    return nil, fmt.Errorf("no cookies found")
}
```

---

## 2. Claude-API

### Key Patterns
- **Browser impersonation headers** to mimic real browser requests
- **Standardized headers** (User-Agent, Accept, etc.)
- **SSE (Server-Sent Events) streaming** for real-time responses
- **Two-stage image upload** (request upload URL, then upload to storage)

### Applied to Ti Router
- Added browser impersonation config to `http_client.go`
- Integrated standardized headers into HTTP retry layer
- Enhanced SSE parser for nested JSON responses

### Implementation Details
```go
// Browser impersonation
type BrowserConfig struct {
    UserAgent string
    Accept    string
    Language  string
    Platform  string
}

// Standardized headers
headers := map[string]string{
    "User-Agent": browserConfig.UserAgent,
    "Accept": "text/event-stream",
    "Accept-Language": browserConfig.Language,
}
```

---

## 3. Gemini-API

### Key Patterns
- **Cookie filtering** to extract only relevant cookies
- **Specific cookie names** for each provider (e.g., `__Secure-1PSID`, `__Secure-1PSIDTS` for Gemini)
- **Auto-extraction** with validation

### Applied to Ti Router
- Implemented `FilterCookies` function in `browser_cookie_extractor.go`
- Integrated filtering into `AutoExtractCookie` in `cookie_manager.go`
- Added provider-specific cookie name mappings

### Implementation Details
```go
// Cookie filtering
func FilterCookies(cookies map[string]string, requiredCookies []string) map[string]string {
    filtered := make(map[string]string)
    for _, name := range requiredCookies {
        if value, ok := cookies[name]; ok {
            filtered[name] = value
        }
    }
    return filtered
}
```

---

## 4. auth2api

### Key Patterns
- **OAuth PKCE flow** (Proof Key for Code Exchange)
- **Code challenge/verifier generation** with SHA256 hashing
- **State parameter** for CSRF protection
- **Token refresh with retry logic**
- **Detection of exhausted refresh tokens**

### Applied to Ti Router
- Created `oauth_pkce.go` with full PKCE implementation
- Supports multiple OAuth providers (Claude, Google, etc.)
- Includes retry mechanism with exponential backoff

### Implementation Details
```go
// PKCE code generation
func GeneratePKCECodes() (*PKCECodes, error) {
    verifier := generateRandomString(128)
    hash := sha256.Sum256([]byte(verifier))
    challenge := base64.RawURLEncoding.EncodeToString(hash[:])
    return &PKCECodes{CodeChallenge: challenge, CodeVerifier: verifier}, nil
}

// Token refresh with retry
func RefreshTokenWithRetry(config *OAuthConfig, refreshToken string, maxRetries int) (*TokenData, error) {
    for attempt := 1; attempt <= maxRetries; attempt++ {
        tokenData, err := RefreshToken(config, refreshToken)
        if err == nil {
            return tokenData, nil
        }
        if isExhaustedError(err) {
            return nil, err // Don't retry exhausted tokens
        }
        time.Sleep(time.Duration(attempt) * time.Second)
    }
}
```

---

## 5. perplexity-ai

### Key Patterns
- **Nested SSE parsing** for complex streaming responses
- **Session initialization** with browser impersonation
- **Two-stage file upload** (S3 direct upload)
- **Nested model preference mapping**

### Applied to Ti Router
- Enhanced `SSEParser` in `http_client.go` to support nested JSON
- Added nested JSON extraction from SSE events
- Supports complex streaming response structures

### Implementation Details
```go
// Nested SSE parsing
func (p *SSEParser) parseNestedJSON(data string) (map[string]interface{}, error) {
    var outer map[string]interface{}
    if err := json.Unmarshal([]byte(data), &outer); err != nil {
        return nil, err
    }
    
    // Extract nested data
    if nested, ok := outer["data"].(map[string]interface{}); ok {
        return nested, nil
    }
    return outer, nil
}
```

---

## 6. sydney.py

### Key Patterns
- **WebSocket streaming** with delimiter-based message splitting
- **Streaming delta pattern** to yield only new tokens incrementally
- **Async context management** for conversation lifecycle
- **Complex nested request building**
- **Conversation signature extraction**

### Applied to Ti Router
- Added `WebSocketStreamParser` with delimiter splitting (default `\u0003`)
- Implemented `StreamingDelta` function for incremental token streaming
- Supports WebSocket connection management

### Implementation Details
```go
// WebSocket stream parser
type WebSocketStreamParser struct {
    delimiter string
}

func (p *WebSocketStreamParser) Parse(data string) []string {
    return strings.Split(data, p.delimiter)
}

// Streaming delta
func StreamingDelta(currentResponse, previousResponse string) string {
    if len(currentResponse) > len(previousResponse) {
        return currentResponse[len(previousResponse):]
    }
    return ""
}
```

---

## 7. CLIProxyAPI

### Key Patterns
- **Multi-storage backend** (Postgres, Git, Object storage, local file)
- **Multiple OAuth login flows** (Google, Codex, Claude, Antigravity, Kimi)
- **Usage statistics tracking** for providers and models
- **Home control plane** for remote configuration
- **TUI mode** with standalone embedded server
- **Cloud deploy mode**

### Applied to Ti Router
- Created `usage_tracker.go` for real-time statistics tracking
- Integrated into `LearningService` for adaptive routing
- Supports provider and model-level metrics

### Implementation Details
```go
// Usage tracker
type UsageTracker struct {
    providerStats map[string]*ProviderStats
    modelStats   map[string]*ModelStats
    retention   time.Duration
}

func (ut *UsageTracker) RecordRequest(providerName, modelName string, latencyMs int64, success bool, costUSD float64) {
    // Track provider statistics
    ps := ut.providerStats[providerName]
    ps.RequestCount++
    ps.TotalLatencyMs += latencyMs
    ps.TotalCostUSD += costUSD
    
    // Track model statistics
    ms := ut.modelStats[modelName]
    ms.RequestCount++
    ms.AvgLatencyMs = float64(ms.TotalLatency()) / float64(ms.RequestCount)
}
```

---

## 8. 9router

### Key Patterns
- **JWT-based authentication** with configurable secret
- **CLI token** based on machine ID (consistent machine ID with salt)
- **Path-based protection** with different levels (ALWAYS_PROTECTED, PROTECTED_API_PATHS)
- **Settings-based authentication toggle** (requireLogin)
- **Tunnel/Tailscale access control** based on hostname
- **Middleware pattern** for Next.js routes
- **Cookie-based auth token storage**

### Applied to Ti Router
- Created `auth_guard.go` for HTTP handler protection
- Integrated into `Router` struct and HTTP handlers
- Protects sensitive endpoints (/api/providers, /api/keys, /api/config)

### Implementation Details
```go
// Auth guard
type AuthGuard struct {
    alwaysProtected map[string]bool
    protectedAPIPaths map[string]bool
}

func (ag *AuthGuard) Protect(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path
        if ag.isAlwaysProtected(path) {
            if ag.HasValidCliToken(r) || ag.HasValidToken(r) {
                next(w, r)
                return
            }
            ag.unauthorized(w)
            return
        }
        next(w, r)
    }
}
```

---

## Specialized Agents Registered

Instead of implementing all functionality in Go, 5 specialized agents were registered into the crew:

### 1. Cloudflare Agent
- **Purpose**: Tunnel/Tailscale access control
- **Capabilities**: Cloudflare Tunnel management, Tailscale integration, DDoS protection
- **Benefits**: Leverages Cloudflare's robust infrastructure, reduces custom Go code

### 2. OAuth Agent
- **Purpose**: OAuth PKCE flows for multiple providers
- **Capabilities**: Token management, refresh logic, multi-provider support
- **Benefits**: Centralizes OAuth logic, handles edge cases

### 3. File Upload Agent
- **Purpose**: File upload patterns
- **Capabilities**: Two-stage upload, image processing, security validation
- **Benefits**: Consistent upload experience across providers

### 4. Storage Agent
- **Purpose**: Multi-storage backend management
- **Capabilities**: Postgres, Git, Object storage, local file storage
- **Benefits**: Flexible storage choices, easy migration

### 5. Conversation Agent
- **Purpose**: Conversation management
- **Capabilities**: CRUD operations, context tracking, session management
- **Benefits**: Optimizes context window, enables conversation restoration

---

## Go-Only Improvements

### UsageTracker Integration
- Integrated into `LearningService` for real-time statistics
- Extracts insights from both memory store and UsageTracker
- Periodic cleanup of old statistics (30-day retention)
- Provides high-confidence insights for adaptive routing

### AuthGuard Integration
- Added to `Router` struct
- Protects sensitive HTTP endpoints
- Supports JWT and CLI token authentication
- Path-based protection with configurable levels

---

## Files Created/Modified

### New Files
- `layers/utils/browser_cookie_extractor.go`
- `layers/utils/http_client.go`
- `layers/auth/oauth_pkce.go`
- `layers/auth/auth_guard.go`
- `layers/stats/usage_tracker.go`
- `content/agents/agents/cloudflare-agent.md`
- `content/agents/agents/oauth-agent.md`
- `content/agents/agents/file-upload-agent.md`
- `content/agents/agents/storage-agent.md`
- `content/agents/agents/conversation-agent.md`

### Modified Files
- `layers/manager/cookie_manager.go`
- `layers/resilience/http_retry.go`
- `cmd/routerd/learning_loop.go`
- `cmd/routerd/router.go`
- `cmd/routerd/server.go`

---

## Build Status
All builds successful. No compilation errors.

---

## Next Steps
1. Create configuration files for each agent
2. Create integration guides
3. Set up monitoring/alerting
4. Create test cases
5. Update documentation
6. Create migration guide

---

## References
- Repository clone path: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router`
- Knowledge base path: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\router`
