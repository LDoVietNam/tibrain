---
tags: ["tibrain", "pattern-auth-cookie", "authentication", "documentation", "provider"]
scopes: ["auth", "tibrain"]
last_updated: 2026-05-22
---
# Cookie Provider Factory Design - Multi-Platform Support

> **Date**: 2026-04-28  
> **Component**: Ti Router - Cookie Provider  
> **Status**: Design Phase

## Overview

Thiết kế generic cookie provider factory để support cookie-based authentication cho nhiều AI services khác nhau, thay vì implement riêng lẻ cho mỗi service như hiện tại (chỉ có Notion).

## Problem Statement

Hiện tại:
- Chỉ có `NotionCookieProvider` được implement riêng
- Mỗi service cần implement riêng → code duplication
- Không có generic pattern cho cookie-based auth
- Khó mở rộng cho services mới

## AI Services with Cookie-Based Auth

Dựa trên research và cookies user cung cấp, các services sau sử dụng cookie auth:

| Service | Domain | Key Cookies | API Endpoint | Status |
|---------|--------|-------------|--------------|--------|
| **SharedChat** | chat.sharedchat.cc, .fun, .cn | `gfsessionid` | `/api/v3/chat` | ✅ Detected |
| **Poe** | poe.com | `poe-formkey`, `poe-tchannel` | `/api/chat` | ✅ Detected |
| **Bing Copilot** | bing.com, copilot.microsoft.com | `_U`, `MUID` | `/api/chat` | ✅ Detected |
| **You.com** | you.com | `you_session` | `/api/chat` | ✅ Detected |
| **Phind** | phind.com | `phind-session` | `/api/search` | ✅ Detected |
| **Character.AI** | character.ai | `character-session` | `/api/chat` | ✅ Detected |
| **Coze** | coze.com | `coze-sessionid` | `/api/chat` | ✅ Detected |
| **Notion** | notion.so | `token_v2` | `/api/v3/chat` | ✅ Implemented |

## Architecture Design

### 1. Generic Cookie Provider Interface

```go
// PlatformAdapter defines interface for platform-specific cookie operations
type PlatformAdapter interface {
    // Name returns the platform identifier
    Name() string
    
    // Detect checks if cookies belong to this platform
    Detect(cookies []*Cookie) bool
    
    // ExtractSessionID extracts the session identifier from cookies
    ExtractSessionID(cookies []*Cookie) (string, error)
    
    // BuildRequest builds HTTP request with platform-specific headers
    BuildRequest(url string, cookies []*Cookie, payload []byte) (*http.Request, error)
    
    // ParseResponse parses platform-specific response
    ParseResponse(resp *http.Response) (*ChatResponse, error)
    
    // RefreshURL returns the URL to hit for cookie refresh
    RefreshURL() string
    
    // ExtractRefreshCookies extracts new cookies from refresh response
    ExtractRefreshCookies(resp *http.Response) ([]*Cookie, error)
}
```

### 2. Cookie Provider Factory

```go
// CookieProviderFactory creates cookie providers for different platforms
type CookieProviderFactory struct {
    adapters map[string]PlatformAdapter
    cookieMgr *cookie.Manager
}

func NewCookieProviderFactory(mgr *cookie.Manager) *CookieProviderFactory {
    factory := &CookieProviderFactory{
        adapters: make(map[string]PlatformAdapter),
        cookieMgr: mgr,
    }
    
    // Register built-in adapters
    factory.RegisterAdapter(&SharedChatAdapter{})
    factory.RegisterAdapter(&PoeAdapter{})
    factory.RegisterAdapter(&BingCopilotAdapter{})
    factory.RegisterAdapter(&YouAdapter{})
    factory.RegisterAdapter(&PhindAdapter{})
    factory.RegisterAdapter(&CharacterAIAdapter{})
    factory.RegisterAdapter(&CozeAdapter{})
    factory.RegisterAdapter(&NotionAdapter{})
    
    return factory
}

// RegisterAdapter adds a platform adapter
func (f *CookieProviderFactory) RegisterAdapter(adapter PlatformAdapter) {
    f.adapters[adapter.Name()] = adapter
}

// CreateProvider creates a cookie provider for given cookies
func (f *CookieProviderFactory) CreateProvider(cookies []*Cookie) (Provider, error) {
    // Auto-detect platform
    for _, adapter := range f.adapters {
        if adapter.Detect(cookies) {
            return &GenericCookieProvider{
                adapter:   adapter,
                cookieMgr: f.cookieMgr,
                profile:   f.createProfile(cookies),
            }, nil
        }
    }
    return nil, fmt.Errorf("unknown platform")
}
```

### 3. Generic Cookie Provider

```go
// GenericCookieProvider implements Provider interface using PlatformAdapter
type GenericCookieProvider struct {
    adapter   PlatformAdapter
    cookieMgr *cookie.Manager
    profile   string
    client    *http.Client
    refreshService *cookie.CookieRefreshService
}

func (p *GenericCookieProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    cookies, _ := p.cookieMgr.GetHTTPCookies(p.profile)
    payload := p.buildPayload(req)
    
    httpReq, err := p.adapter.BuildRequest(p.getAPIURL(), cookies, payload)
    if err != nil {
        return nil, err
    }
    
    resp, err := p.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    return p.adapter.ParseResponse(resp)
}

func (p *GenericCookieProvider) Init(cfg *config.Config) error {
    // Initialize cookie manager with cookies from config or file
    // Setup refresh service
    refreshFunc := func(resp *http.Response) ([]*cookie.Cookie, error) {
        return p.adapter.ExtractRefreshCookies(resp)
    }
    
    p.refreshService = cookie.NewCookieRefreshService(
        p.cookieMgr,
        p.profile,
        p.adapter.RefreshURL(),
        refreshFunc,
    )
    
    return p.refreshService.Start()
}
```

### 4. Platform Adapter Implementations

#### SharedChat Adapter

```go
type SharedChatAdapter struct{}

func (a *SharedChatAdapter) Name() string { return "sharedchat" }

func (a *SharedChatAdapter) Detect(cookies []*Cookie) bool {
    for _, ck := range cookies {
        if ck.Name == "gfsessionid" && strings.Contains(ck.Domain, "sharedchat") {
            return true
        }
    }
    return false
}

func (a *SharedChatAdapter) ExtractSessionID(cookies []*Cookie) (string, error) {
    for _, ck := range cookies {
        if ck.Name == "gfsessionid" {
            return ck.Value, nil
        }
    }
    return "", fmt.Errorf("gfsessionid not found")
}

func (a *SharedChatAdapter) BuildRequest(url string, cookies []*Cookie, payload []byte) (*http.Request, error) {
    req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
    if err != nil {
        return nil, err
    }
    
    // Add cookies
    for _, ck := range cookies {
        req.AddCookie(&http.Cookie{
            Name:  ck.Name,
            Value: ck.Value,
            Domain: ck.Domain,
            Path: ck.Path,
        })
    }
    
    // SharedChat-specific headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
    
    return req, nil
}

func (a *SharedChatAdapter) ParseResponse(resp *http.Response) (*ChatResponse, error) {
    // Parse SharedChat-specific response format
    // ...
}

func (a *SharedChatAdapter) RefreshURL() string {
    return "https://chat.sharedchat.cc/"
}

func (a *SharedChatAdapter) ExtractRefreshCookies(resp *http.Response) ([]*Cookie, error) {
    // Extract cookies from Set-Cookie headers
    // ...
}
```

#### Poe Adapter

```go
type PoeAdapter struct{}

func (a *PoeAdapter) Name() string { return "poe" }

func (a *PoeAdapter) Detect(cookies []*Cookie) bool {
    for _, ck := range cookies {
        if (ck.Name == "poe-formkey" || ck.Name == "poe-tchannel") && strings.Contains(ck.Domain, "poe.com") {
            return true
        }
    }
    return false
}

// ... similar methods for Poe-specific implementation
```

### 5. Configuration

```yaml
# config/cookie-providers.yaml
cookie_providers:
  sharedchat:
    profile_name: "sharedchat"
    cookie_file: "data/sharedchat-cookies.json"
    api_url: "https://chat.sharedchat.cc/api/v3/chat"
    refresh_enabled: true
    refresh_interval: "1m"
    refresh_before: "5m"
    
  poe:
    profile_name: "poe"
    cookie_file: "data/poe-cookies.json"
    api_url: "https://poe.com/api/chat"
    refresh_enabled: true
    refresh_interval: "2m"
    refresh_before: "10m"
    
  notion:
    profile_name: "notion"
    cookie_file: "data/notion-cookie.json"
    api_url: "https://www.notion.so/api/v3/chat"
    refresh_enabled: false
```

## Implementation Plan

### Phase 1: Core Infrastructure
1. Create `platform_adapter.go` - Interface definition
2. Create `cookie_provider_factory.go` - Factory implementation
3. Create `generic_cookie_provider.go` - Generic provider
4. Update existing `cookie.go` - Integration with factory

### Phase 2: Platform Adapters
1. Implement `sharedchat_adapter.go`
2. Implement `poe_adapter.go`
3. Implement `bing_copilot_adapter.go`
4. Implement `you_adapter.go`
5. Implement `phind_adapter.go`
6. Implement `character_ai_adapter.go`
7. Implement `coze_adapter.go`
8. Refactor `notion_cookie.go` to use adapter pattern

### Phase 3: Configuration & Integration
1. Create cookie provider config loader
2. Update router config to include cookie providers
3. Add CLI commands for cookie management
4. Add cookie import/export utilities

### Phase 4: Testing
1. Unit tests for each adapter
2. Integration tests with real cookies
3. End-to-end tests for chat flow
4. Refresh service tests for each platform

### Phase 5: Documentation
1. Adapter implementation guide
2. Adding new platform guide
3. Configuration reference
4. Usage examples

## Benefits

1. **Extensibility**: Thêm platform mới chỉ cần implement adapter interface
2. **Maintainability**: Platform-specific logic isolated trong adapters
3. **Reusability**: Generic provider logic shared across platforms
4. **Testability**: Mỗi adapter có thể test independently
5. **Flexibility**: Easy to add/remove platforms without affecting others

## Migration Path

1. Keep existing `NotionCookieProvider` for backward compatibility
2. Implement factory alongside existing code
3. Gradually migrate other providers to use factory
4. Deprecate individual provider implementations

## Open Questions

1. **Cookie Storage**: Should cookies be stored per-profile or globally?
2. **Refresh Strategy**: Should refresh be opt-in or opt-out per platform?
3. **Rate Limiting**: Should rate limiting be per-platform or global?
4. **Error Handling**: How to handle platform-specific errors generically?
5. **Configuration**: YAML config vs code registration for adapters?

## References

- OmniRoute: https://github.com/diegosouzapw/OmniRoute
- Existing detector.go: Platform detection logic
- Existing cookie.go: Cookie manager implementation
- Existing notion_cookie.go: Reference implementation
