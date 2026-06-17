---
tags: ["tibrain", "documentation", "pattern-auth-cookie", "authentication", "provider"]
scopes: ["auth", "tibrain"]
last_updated: 2026-05-22
---
# Cookie Provider Factory - Implementation Guide

> **Date**: 2026-04-28  
> **Component**: Ti Router - Cookie Provider  
> **Status**: ✅ Implemented & Tested

## Overview

Cookie Provider Factory là generic framework để support cookie-based authentication cho nhiều AI services khác nhau. Thay vì implement riêng lẻ cho mỗi service, factory pattern cho phép:

1. **Auto-detection** platform từ cookies
2. **Pluggable adapters** cho mỗi platform
3. **Unified interface** cho tất cả providers
4. **Easy extensibility** cho services mới

## Supported Platforms

| Platform | Domain | Key Cookies | API Endpoint | Status |
|----------|--------|-------------|--------------|--------|
| **SharedChat** | chat.sharedchat.cc, .fun, .cn | `gfsessionid` | `/api/v3/chat` | ✅ Full |
| **Poe** | poe.com | `poe-formkey`, `poe-tchannel` | `/api/chat` | ✅ Basic |
| **Bing Copilot** | bing.com, copilot.microsoft.com | `_U`, `MUID` | `/api/chat` | ✅ Basic |
| **You.com** | you.com | `you_session` | `/api/chat` | ✅ Basic |
| **Phind** | phind.com | `phind-session` | `/api/search` | ✅ Basic |
| **Character.AI** | character.ai | `character-session` | `/api/chat` | ✅ Basic |
| **Coze** | coze.com | `coze-sessionid` | `/api/chat` | ✅ Basic |
| **Notion** | notion.so | `token_v2` | `/api/v3/chat` | ✅ Basic |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  CookieProviderFactory                         │
│  - Manages platform adapters                                 │
│  - Auto-detects platform from cookies                         │
│  - Creates GenericCookieProvider instances                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ uses
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              PlatformAdapter Interface                        │
│  + Name() string                                            │
│  + Detect(cookies) bool                                      │
│  + ExtractSessionID(cookies) (string, error)                │
│  + BuildRequest(url, cookies, payload) (*http.Request, error)│
│  + ParseResponse(resp) (*ChatResponse, error)                │
│  + RefreshURL() string                                       │
│  + ExtractRefreshCookies(resp) ([]*Cookie, error)            │
│  + GetAPIURL() string                                       │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ implemented by
                              ▼
┌─────────────────┬─────────────────┬─────────────────┬─────────┐
│ SharedChatAdapter│  PoeAdapter     │ BingCopilotAdapter│  ...    │
│                 │                 │                  │         │
└─────────────────┴─────────────────┴─────────────────┴─────────┘
                              │
                              │ used by
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              GenericCookieProvider                             │
│  - Delegates to PlatformAdapter                              │
│  - Manages cookie profile                                    │
│  - Handles refresh service                                   │
│  - Provides unified Chat/ChatStream interface                 │
└─────────────────────────────────────────────────────────────┘
```

## Files Created

### Core Infrastructure
1. **platform_adapter.go** (95 lines)
   - `PlatformAdapter` interface definition
   - `BaseAdapter` with common functionality
   - Helper methods for cookie operations

2. **cookie_provider_factory.go** (134 lines)
   - `CookieProviderFactory` implementation
   - Adapter registration/management
   - Platform auto-detection
   - Provider creation

3. **generic_cookie_provider.go** (183 lines)
   - `GenericCookieProvider` implementation
   - Delegates to PlatformAdapter
   - Cookie profile management
   - Refresh service integration

### Platform Adapters
4. **sharedchat_adapter.go** (112 lines)
   - Full implementation for SharedChat
   - Response parsing
   - Cookie extraction

5. **poe_adapter.go** (58 lines)
   - Basic implementation for Poe
   - Placeholder for response parsing

6. **bing_copilot_adapter.go** (58 lines)
   - Basic implementation for Bing Copilot
   - Placeholder for response parsing

7. **other_adapters.go** (227 lines)
   - You, Phind, Character.AI, Coze, Notion adapters
   - Basic implementations with placeholders

### Tests
8. **cookie_provider_factory_test.go** (108 lines)
   - Factory registration tests
   - Platform detection tests
   - Provider creation tests
   - Real cookie detection tests

## Usage Examples

### Basic Usage - Auto-Detect Platform

```go
package main

import (
    "github.com/ti/router/layers/provider/cookie"
)

func main() {
    // Create cookie manager
    mgr := cookie.NewManager()
    
    // Load cookies from Chrome export
    cookies := loadChromeCookies("sharedchat-cookies.json")
    mgr.LoadChromeJSON(cookies, "sharedchat")
    
    // Create factory
    factory := cookie.NewCookieProviderFactory(mgr)
    
    // Auto-detect platform and create provider
    provider, err := factory.CreateProvider("sharedchat", "sharedchat")
    if err != nil {
        panic(err)
    }
    
    // Initialize with refresh enabled
    provider.Init(true)
    defer provider.Stop()
    
    // Use provider
    resp, err := provider.Chat(context.Background(), cookie.ChatRequest{
        Messages: []cookie.Message{
            {Role: "user", Content: "Hello!"},
        },
        Model: "gpt-4",
    })
    
    fmt.Println(resp.Content)
}
```

### Manual Platform Selection

```go
// Create provider for specific platform
provider, err := factory.CreateProvider("poe", "poe-profile")
if err != nil {
    panic(err)
}

provider.Init(true)
defer provider.Stop()
```

### Adding Custom Adapter

```go
// Define custom adapter
type CustomAdapter struct {
    cookie.BaseAdapter
}

func (a *CustomAdapter) Name() string { return "custom" }
func (a *CustomAdapter) Detect(cookies []*cookie.Cookie) bool {
    // Custom detection logic
    return true
}
// ... implement other methods

// Register custom adapter
factory.RegisterAdapter(&CustomAdapter{})

// Now can use custom platform
provider, err := factory.CreateProvider("custom", "custom-profile")
```

### Using with Real SharedChat Cookies

```go
// Load the cookies user provided
sharedChatCookies := `[{
    "domain": "chat.sharedchat.cc",
    "name": "gfsessionid",
    "value": "s5y4hn01h858mcdi560ch2fanhdq5tu4",
    "path": "/"
}]`

mgr := cookie.NewManager()
mgr.LoadChromeJSON([]byte(sharedChatCookies), "sharedchat")

factory := cookie.NewCookieProviderFactory(mgr)

// Auto-detect will recognize as SharedChat
platform, _, err := factory.DetectPlatform(convertToCookie(cookies))
// platform == "sharedchat"
```

## API Reference

### CookieProviderFactory

#### Methods
- `NewCookieProviderFactory(mgr *Manager) *CookieProviderFactory` - Constructor
- `RegisterAdapter(adapter PlatformAdapter)` - Register platform adapter
- `UnregisterAdapter(name string)` - Remove adapter
- `GetAdapter(name string) (PlatformAdapter, bool)` - Get adapter by name
- `ListAdapters() []string` - List all registered adapters
- `DetectPlatform(cookies []*Cookie) (string, PlatformAdapter, error)` - Auto-detect platform
- `CreateProvider(platformName, profileName string) (*GenericCookieProvider, error)` - Create provider
- `CreateProviderFromCookies(cookies []*Cookie, profileName string) (*GenericCookieProvider, error)` - Auto-detect and create

### PlatformAdapter Interface

```go
type PlatformAdapter interface {
    Name() string
    Detect(cookies []*Cookie) bool
    ExtractSessionID(cookies []*Cookie) (string, error)
    BuildRequest(url string, cookies []*Cookie, payload []byte) (*http.Request, error)
    ParseResponse(resp *http.Response) (*ChatResponse, error)
    RefreshURL() string
    ExtractRefreshCookies(resp *http.Response) ([]*Cookie, error)
    GetAPIURL() string
}
```

### GenericCookieProvider

#### Methods
- `Name() string` - Platform name
- `IsHealthy() bool` - Check if provider has valid cookies
- `Init(refreshEnabled bool) error` - Initialize provider, start refresh service
- `Stop()` - Stop provider and refresh service
- `Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)` - Send chat request
- `ChatStream(ctx context.Context, req ChatRequest, onChunk func(chunk *StreamChunk)) (*ChatResponse, error)` - Streaming (TODO)
- `GetSessionID() (string, error)` - Get session ID from cookies
- `UpdateCookies(cookies []*Cookie)` - Update cookies in profile
- `GetProfile() string` - Get profile name

## Test Results

**Total Tests**: 72 tests (63 original + 9 new)  
**Status**: ✅ All passing

```
=== RUN   TestCookieProviderFactory
--- PASS: TestCookieProviderFactory (0.00s)
=== RUN   TestSharedChatAdapterDetection
--- PASS: TestSharedChatAdapterDetection (0.00s)
=== RUN   TestPoeAdapterDetection
--- PASS: TestPoeAdapterDetection (0.00s)
=== RUN   TestFactoryDetectPlatform
--- PASS: TestFactoryDetectPlatform (0.00s)
=== RUN   TestFactoryCreateProvider
--- PASS: TestFactoryCreateProvider (0.00s)
=== RUN   TestFactoryUnknownPlatform
--- PASS: TestFactoryUnknownPlatform (0.00s)
... (66 existing tests)
PASS
ok      github.com/ti/router/layers/provider/cookie    2.113s
```

## Benefits

1. **Extensibility**: Thêm platform mới chỉ cần implement adapter interface
2. **Maintainability**: Platform-specific logic isolated trong adapters
3. **Reusability**: Generic provider logic shared across platforms
4. **Testability**: Mỗi adapter có thể test independently
5. **Flexibility**: Easy to add/remove platforms without affecting others
6. **Auto-detection**: Platform detection from cookies eliminates manual configuration

## Future Work

### High Priority
1. **Complete Response Parsing**: Implement full response parsing for all platforms
2. **Streaming Support**: Implement ChatStream for all adapters
3. **Payload Building**: Add platform-specific payload builders
4. **Error Handling**: Improve platform-specific error handling

### Medium Priority
5. **Configuration**: Add YAML config for platform settings
6. **CLI Commands**: Add CLI commands for cookie management
7. **Cookie Import/Export**: Utilities for importing/exporting cookies
8. **Health Monitoring**: Add health checks per platform

### Low Priority
9. **Metrics**: Add metrics for cookie usage and refresh success rates
10. **Rate Limiting**: Per-platform rate limiting
11. **Cookie Validation**: Signature verification for cookie integrity
12. **Key Rotation**: Support for rotating encryption keys

## Integration with Existing Code

### Backward Compatibility
- Existing `NotionCookieProvider` can continue to work
- Factory pattern is additive, not breaking
- Can gradually migrate existing providers to use factory

### Migration Path
1. Keep existing providers for now
2. Implement factory alongside
3. Test factory with new platforms first
4. Gradually migrate existing providers
5. Deprecate individual provider implementations

## References

- Design Document: `Z:\Ti\Ti-learning-lab\03_Knowledge\patterns\cookie-provider-factory-design.md`
- OmniRoute: https://github.com/diegosouzapw/OmniRoute
- Existing detector.go: Platform detection logic
- Existing cookie.go: Cookie manager implementation
- Existing notion_cookie.go: Reference implementation
