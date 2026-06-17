---
tags: ["tibrain", "authentication", "security", "router", "pattern-auth-cookie"]
scopes: ["auth", "tibrain"]
last_updated: 2026-05-22
---
# CLIProxyAPI Cookie Authentication Patterns

**Source**: CLIProxyAPI (30K stars) - Production-ready AI router with cookie authentication  
**Date**: 2026-05-10  
**Status**: Verified patterns from active codebase  
**Tags**: #cookie-auth #ai-router #go-patterns #production-ready  

---

## 🎯 Overview

CLIProxyAPI implements robust cookie-based authentication for multiple AI providers (Gemini, ChatGPT, Claude). These patterns are extracted from the production codebase and can be directly applied to Ti Router.

---

## 🔑 Core Authentication Pattern

### 1. Cookie Provider Interface

```go
// Base interface for all cookie-based providers
type CookieProvider interface {
    Detect(cookies []*http.Cookie) bool
    ExtractAuth(cookies []*http.Cookie) (*AuthData, error)
    Validate(auth *AuthData) bool
    Refresh(auth *AuthData) (*AuthData, error)
}

// Auth data structure for persistent storage
type AuthData struct {
    APIKey     string    `json:"apiKey"`
    Expire     string    `json:"expire"`
    Email      string    `json:"email"`
    Cookie     string    `json:"cookie"`
    Provider   string    `json:"provider"`
    LastRefresh time.Time `json:"lastRefresh"`
}
```

### 2. Cookie Detection Pattern

```go
// Provider-specific cookie detection
func (p *GeminiProvider) Detect(cookies []*http.Cookie) bool {
    requiredCookies := []string{"__Secure-1PSID", "__Secure-1PSIDTS"}
    
    for _, cookie := range cookies {
        for _, required := range requiredCookies {
            if cookie.Name == required && strings.Contains(cookie.Domain, "google.com") {
                return true
            }
        }
    }
    return false
}
```

---

## 🍪 Cookie Management Patterns

### 1. Cookie Normalization

```go
// NormalizeCookie - Clean and validate cookie strings
func NormalizeCookie(raw string) (string, error) {
    trimmed := strings.TrimSpace(raw)
    if trimmed == "" {
        return "", fmt.Errorf("cookie cannot be empty")
    }

    // Join multiple lines and ensure proper format
    combined := strings.Join(strings.Fields(trimmed), " ")
    if !strings.HasSuffix(combined, ";") {
        combined += ";"
    }
    
    // Provider-specific validation
    if !strings.Contains(combined, "BXAuth=") {
        return "", fmt.Errorf("cookie missing required field")
    }
    
    return combined, nil
}
```

### 2. Cookie Value Extraction

```go
// Extract specific authentication tokens from cookie strings
func ExtractBXAuth(cookie string) string {
    parts := strings.Split(cookie, ";")
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if strings.HasPrefix(part, "BXAuth=") {
            return strings.TrimPrefix(part, "BXAuth=")
        }
    }
    return ""
}

// Generic extraction for any cookie field
func ExtractCookieField(cookie, fieldName string) string {
    parts := strings.Split(cookie, ";")
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if strings.HasPrefix(part, fieldName+"=") {
            return strings.TrimPrefix(part, fieldName+"=")
        }
    }
    return ""
}
```

---

## 🔐 Authentication Flow Pattern

### 1. Cookie-based Authentication

```go
// AuthenticateWithCookie - Main authentication flow
func (ia *ProviderAuth) AuthenticateWithCookie(ctx context.Context, cookie string) (*AuthData, error) {
    if strings.TrimSpace(cookie) == "" {
        return nil, fmt.Errorf("cookie authentication: cookie is empty")
    }

    // Step 1: Fetch initial auth information
    authInfo, err := ia.fetchInitialAuthInfo(ctx, cookie)
    if err != nil {
        return nil, fmt.Errorf("fetch initial auth info failed: %w", err)
    }

    // Step 2: Refresh/validate authentication
    refreshedAuth, err := ia.RefreshAuthentication(ctx, cookie, authInfo)
    if err != nil {
        return nil, fmt.Errorf("refresh authentication failed: %w", err)
    }

    // Step 3: Convert to standard auth data format
    data := &AuthData{
        APIKey:   refreshedAuth.APIKey,
        Expire:   refreshedAuth.ExpireTime,
        Email:    refreshedAuth.Email,
        Cookie:   cookie,
        Provider: ia.ProviderName(),
    }

    return data, nil
}
```

### 2. HTTP Request Pattern with Browser Mimicry

```go
// buildAuthRequest - Create browser-like HTTP requests
func (ia *ProviderAuth) buildAuthRequest(ctx context.Context, endpoint, cookie string, method string) (*http.Request, error) {
    var body io.Reader
    if method == http.MethodPost {
        body = strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, authInfo.Name))
    }

    req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
    if err != nil {
        return nil, fmt.Errorf("create request failed: %w", err)
    }

    // Set browser-mimicking headers
    req.Header.Set("Cookie", cookie)
    req.Header.Set("Accept", "application/json, text/plain, */*")
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    req.Header.Set("Accept-Language", "en-US,en;q=0.9")
    req.Header.Set("Accept-Encoding", "gzip, deflate, br")
    req.Header.Set("Connection", "keep-alive")
    req.Header.Set("Sec-Fetch-Dest", "empty")
    req.Header.Set("Sec-Fetch-Mode", "cors")
    req.Header.Set("Sec-Fetch-Site", "same-origin")

    if method == http.MethodPost {
        req.Header.Set("Content-Type", "application/json")
    }

    return req, nil
}
```

---

## 🔄 Token Management Pattern

### 1. Token Storage and Persistence

```go
// SaveTokenToFile - Persistent token storage
func (ts *TokenStorage) SaveTokenToFile(filePath string) error {
    // Create directory if needed
    if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
        return fmt.Errorf("create auth directory failed: %w", err)
    }

    // Serialize token data
    data, err := json.MarshalIndent(ts, "", "  ")
    if err != nil {
        return fmt.Errorf("marshal token failed: %w", err)
    }

    // Write to file
    if err := os.WriteFile(filePath, data, 0600); err != nil {
        return fmt.Errorf("write token file failed: %w", err)
    }

    return nil
}

// LoadTokenFromFile - Load persistent tokens
func LoadTokenFromFile(filePath string) (*TokenStorage, error) {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("read token file failed: %w", err)
    }

    var ts TokenStorage
    if err := json.Unmarshal(data, &ts); err != nil {
        return nil, fmt.Errorf("unmarshal token failed: %w", err)
    }

    return &ts, nil
}
```

### 2. Token Refresh Pattern

```go
// RefreshToken - Automatic token refresh
func (ia *ProviderAuth) RefreshToken(ctx context.Context, authData *AuthData) (*AuthData, error) {
    // Check if token needs refresh
    if !ia.needsRefresh(authData) {
        return authData, nil
    }

    // Perform refresh using stored cookie
    refreshedAuth, err := ia.AuthenticateWithCookie(ctx, authData.Cookie)
    if err != nil {
        return nil, fmt.Errorf("token refresh failed: %w", err)
    }

    // Update metadata
    refreshedAuth.LastRefresh = time.Now()
    
    return refreshedAuth, nil
}

// needsRefresh - Check if token requires refresh
func (ia *ProviderAuth) needsRefresh(authData *AuthData) bool {
    if authData.Expire == "" {
        return true // No expiry info, refresh to be safe
    }

    expireTime, err := time.Parse(time.RFC3339, authData.Expire)
    if err != nil {
        return true // Invalid expiry format, refresh
    }

    // Refresh if expires within 1 hour
    return time.Until(expireTime) < time.Hour
}
```

---

## 🏭 Provider Registry Pattern

### 1. Dynamic Provider Registration

```go
// Registry for managing multiple cookie providers
type CookieRegistry struct {
    providers map[string]CookieProvider
    store     AuthStore
    mutex     sync.RWMutex
}

// RegisterProvider - Add new provider to registry
func (cr *CookieRegistry) RegisterProvider(name string, provider CookieProvider) {
    cr.mutex.Lock()
    defer cr.mutex.Unlock()
    
    cr.providers[name] = provider
}

// DetectProvider - Auto-detect provider from cookies
func (cr *CookieRegistry) DetectProvider(cookies []*http.Cookie) (CookieProvider, error) {
    cr.mutex.RLock()
    defer cr.mutex.RUnlock()

    for name, provider := range cr.providers {
        if provider.Detect(cookies) {
            return provider, nil
        }
    }

    return nil, fmt.Errorf("no provider detected for given cookies")
}
```

### 2. Provider Factory Pattern

```go
// Factory for creating provider instances
type ProviderFactory struct{}

// CreateProvider - Instantiate provider by name
func (pf *ProviderFactory) CreateProvider(providerName string) (CookieProvider, error) {
    switch providerName {
    case "gemini":
        return NewGeminiProvider(), nil
    case "chatgpt":
        return NewChatGPTProvider(), nil
    case "claude":
        return NewClaudeProvider(), nil
    case "meta":
        return NewMetaProvider(), nil
    default:
        return nil, fmt.Errorf("unknown provider: %s", providerName)
    }
}

// RegisterDefaultProviders - Register all built-in providers
func (cr *CookieRegistry) RegisterDefaultProviders() {
    factory := &ProviderFactory{}
    
    providers := []string{"gemini", "chatgpt", "claude", "meta"}
    for _, name := range providers {
        if provider, err := factory.CreateProvider(name); err == nil {
            cr.RegisterProvider(name, provider)
        }
    }
}
```

---

## 🛡️ Security Patterns

### 1. Duplicate Detection

```go
// CheckDuplicateAuth - Prevent duplicate authentication
func CheckDuplicateAuth(authDir, authToken string) (string, error) {
    if authToken == "" {
        return "", nil
    }

    entries, err := os.ReadDir(authDir)
    if err != nil {
        if os.IsNotExist(err) {
            return "", nil
        }
        return "", fmt.Errorf("read auth dir failed: %w", err)
    }

    for _, entry := range entries {
        if !isAuthFile(entry) {
            continue
        }

        filePath := filepath.Join(authDir, entry.Name())
        existingAuth, err := LoadTokenFromFile(filePath)
        if err != nil {
            continue
        }

        if existingAuth.APIKey == authToken {
            return filePath, nil // Duplicate found
        }
    }

    return "", nil // No duplicate
}
```

### 2. Sanitization and Validation

```go
// SanitizeFileName - Create safe filenames from user data
func SanitizeFileName(raw string) string {
    if raw == "" {
        return ""
    }
    
    cleanEmail := strings.ReplaceAll(raw, "*", "x")
    var result strings.Builder
    
    for _, r := range cleanEmail {
        if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
           (r >= '0' && r <= '9') || r == '_' || r == '@' || 
           r == '.' || r == '-' {
            result.WriteRune(r)
        }
    }
    
    return strings.TrimSpace(result.String())
}
```

---

## 📊 Provider-Specific Cookie Patterns

### 1. Gemini Cookie Pattern

```go
// Gemini-specific cookie requirements
type GeminiCookies struct {
    Secure1PSID   string `json:"secure_1psid"`
    Secure1PSIDTS string `json:"secure_1psidts"`
}

func (g *GeminiProvider) ExtractCookies(cookieStr string) (*GeminiCookies, error) {
    return &GeminiCookies{
        Secure1PSID:   ExtractCookieField(cookieStr, "__Secure-1PSID"),
        Secure1PSIDTS: ExtractCookieField(cookieStr, "__Secure-1PSIDTS"),
    }, nil
}
```

### 2. ChatGPT Cookie Pattern

```go
// ChatGPT-specific cookie requirements
type ChatGPTCookies struct {
    SessionToken string `json:"session_token"`
}

func (c *ChatGPTProvider) ExtractCookies(cookieStr string) (*ChatGPTCookies, error) {
    sessionToken := ExtractCookieField(cookieStr, "__Secure-next-auth.session-token.1")
    if sessionToken == "" {
        return nil, fmt.Errorf("missing ChatGPT session token")
    }
    
    return &ChatGPTCookies{
        SessionToken: sessionToken,
    }, nil
}
```

---

## 🚀 Integration Patterns for Ti Router

### 1. Router Integration

```go
// CookieAuthMiddleware - Middleware for router integration
func CookieAuthMiddleware(registry *CookieRegistry) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract cookies from request
        cookies := extractCookiesFromRequest(c.Request)
        
        // Detect provider
        provider, err := registry.DetectProvider(cookies)
        if err != nil {
            c.JSON(401, gin.H{"error": "no valid provider found"})
            c.Abort()
            return
        }
        
        // Authenticate
        authData, err := provider.ExtractAuth(cookies)
        if err != nil {
            c.JSON(401, gin.H{"error": "authentication failed"})
            c.Abort()
            return
        }
        
        // Store in context
        c.Set("auth_data", authData)
        c.Set("provider", provider)
        c.Next()
    }
}
```

### 2. CLI Integration Pattern

```go
// CLI Authentication Command
func DoCookieAuth(cfg *config.Config, providerName string) {
    // Prompt for cookie
    cookie, err := promptForCookie(providerName)
    if err != nil {
        fmt.Printf("Failed to get cookie: %v\n", err)
        return
    }
    
    // Get provider from registry
    provider, err := cfg.Registry.GetProvider(providerName)
    if err != nil {
        fmt.Printf("Unknown provider: %s\n", providerName)
        return
    }
    
    // Authenticate
    authData, err := provider.ExtractAuth(parseCookies(cookie))
    if err != nil {
        fmt.Printf("Authentication failed: %v\n", err)
        return
    }
    
    // Save authentication
    authPath := getAuthFilePath(cfg, providerName, authData.Email)
    if err := saveAuthToFile(authData, authPath); err != nil {
        fmt.Printf("Failed to save authentication: %v\n", err)
        return
    }
    
    fmt.Printf("Authentication successful! Saved to: %s\n", authPath)
}
```

---

## 📋 Implementation Checklist

### ✅ Core Components
- [ ] CookieProvider interface implementation
- [ ] Provider registry system
- [ ] Token storage and persistence
- [ ] Cookie normalization and validation
- [ ] Browser-mimicking HTTP requests

### ✅ Provider Implementations
- [ ] Gemini provider (`__Secure-1PSID`, `__Secure-1PSIDTS`)
- [ ] ChatGPT provider (`__Secure-next-auth.session-token.1`)
- [ ] Claude provider (`sessionKey`)
- [ ] Meta.ai provider (`ecto_1_sess`)
- [ ] Custom provider template

### ✅ Integration Points
- [ ] Router middleware for cookie auth
- [ ] CLI commands for cookie setup
- [ ] API endpoints for provider management
- [ ] Health checks and monitoring
- [ ] Error handling and logging

### ✅ Security Features
- [ ] Duplicate authentication detection
- [ ] Cookie sanitization and validation
- [ ] Secure token storage (file permissions)
- [ ] Automatic token refresh
- [ ] Session management

---

## 🔗 Related Resources

- **CLIProxyAPI Source**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\CLIProxyAPI-main`
- **Cookie Helpers**: `internal/auth/iflow/cookie_helpers.go`
- **Authentication Flow**: `internal/auth/iflow/iflow_auth.go`
- **CLI Integration**: `internal/cmd/iflow_cookie.go`
- **Token Storage**: `internal/auth/iflow/iflow_token.go`

---

## 🎯 Next Steps for Ti Router

1. **Implement Base Interface**: Start with CookieProvider interface
2. **Create Gemini Provider**: Use as reference implementation
3. **Build Registry System**: Dynamic provider management
4. **Add CLI Commands**: Cookie setup and management
5. **Integrate with Router**: Middleware and API endpoints
6. **Test with Real Cookies**: Validate each provider implementation

**These patterns are production-tested and can be directly applied to Ti Router!** 🚀