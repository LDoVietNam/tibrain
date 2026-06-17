# Networking & Proxy Support

> **Source**: Adapted from Cline patterns (TypeScript) to Go
> **Purpose**: Ensure HTTP clients respect proxy environment variables in all environments
> **Last Updated**: 2026-05-09

---

## Overview

To ensure Ti CLI works correctly in all environments (especially corporate proxies), strictly follow these guidelines for all network activity.

**Key Principle**: Every `http.Client` MUST be configured to respect proxy environment variables.

## Environment Variables

Respect these standard proxy environment variables:

| Variable | Purpose |
|----------|---------|
| `HTTP_PROXY` / `http_proxy` | Proxy for HTTP requests |
| `HTTPS_PROXY` / `https_proxy` | Proxy for HTTPS requests |
| `NO_PROXY` / `no_proxy` | Comma-separated list of hosts to bypass proxy |

Priority: Lowercase takes precedence over uppercase (Go's `ProxyFromEnvironment` behavior).

## Go Implementation

### Pattern 1: Use `http.ProxyFromEnvironment`

**DO THIS:**

```go
// Good: Transport respects proxy environment variables
transport := &http.Transport{
    Proxy: http.ProxyFromEnvironment,
    // Other settings...
}

client := &http.Client{
    Transport: transport,
    Timeout:   30 * time.Second,
}
```

**NOT THIS:**

```go
// Bad: Default http.Client without explicit proxy configuration
// While http.DefaultTransport has ProxyFromEnvironment,
// it's better to be explicit and share a configured client
client := &http.Client{
    Timeout: 30 * time.Second,
    // Transport not set - uses http.DefaultTransport
}
```

### Pattern 2: Shared HTTP Client Factory

Create a shared utility function in your project:

```go
package httpclient

import (
    "net/http"
    "time"
)

// Default timeout for HTTP clients
const DefaultTimeout = 30 * time.Second

// New creates a new proxy-aware HTTP client
func New() *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            Proxy:                 http.ProxyFromEnvironment,
            MaxIdleConns:          100,
            IdleConnTimeout:       90 * time.Second,
            TLSHandshakeTimeout:   10 * time.Second,
            ExpectContinueTimeout: 1 * time.Second,
        },
        Timeout: DefaultTimeout,
    }
}

// NewWithTimeout creates a proxy-aware HTTP client with custom timeout
func NewWithTimeout(timeout time.Duration) *http.Client {
    client := New()
    client.Timeout = timeout
    return client
}
```

### Pattern 3: Third-Party API Clients

For SDKs that accept custom HTTP clients:

```go
// Example: OpenAI SDK
import (
    openai "github.com/sashabaranov/go-openai"
    "your/project/httpclient"
)

config := openai.DefaultConfig("api-key")
config.HTTPClient = httpclient.New()  // <--- CRITICAL: Pass proxy-aware client

client := openai.NewClientWithConfig(config)
```

### Pattern 4: Custom Transport Settings

If you need specific transport settings:

```go
// Create transport with custom settings, but keep proxy support
transport := &http.Transport{
    Proxy:               http.ProxyFromEnvironment,  // <--- KEEP THIS
    MaxIdleConns:        50,
    MaxIdleConnsPerHost: 10,
    DisableCompression:  false,
}

client := &http.Client{
    Transport: transport,
    Timeout:   60 * time.Second,
}
```

## Common Patterns in Ti CLI

### Current Issues Found

These files create `http.Client` without explicit proxy configuration:

| File | Location | Status |
|------|----------|--------|
| `temp_mail_verifier.go` | `internal/autoreg/` | Creates `&http.Client{Timeout: ...}` |
| `client.go` | `internal/donutbrowser/` | Creates `&http.Client{Timeout: ...}` |

### Fix Pattern

Apply this fix to all http.Client creations:

```go
// Before (no explicit proxy):
client: &http.Client{
    Timeout: 30 * time.Second,
}

// After (proxy-aware):
client: &http.Client{
    Transport: &http.Transport{
        Proxy: http.ProxyFromEnvironment,
    },
    Timeout: 30 * time.Second,
}
```

## Verification Checklist

When adding a new network call or integration:

1. [ ] Check that `http.Transport` has `Proxy: http.ProxyFromEnvironment`
2. [ ] Or use the shared `httpclient.New()` factory
3. [ ] For third-party SDKs, verify custom HTTP client is passed
4. [ ] Test with proxy environment variables set:
   ```powershell
   $env:HTTP_PROXY = "http://proxy:8080"
   $env:HTTPS_PROXY = "http://proxy:8080"
   # Test your application
   ```

## Testing Proxy Support

To verify proxy support works:

### Method 1: Test with NO_PROXY

```powershell
# Bypass proxy for specific domains
$env:NO_PROXY = "localhost,127.0.0.1,.internal"
```

### Method 2: Inspect Transport

```go
// Debug: Verify proxy is configured
transport, ok := client.Transport.(*http.Transport)
if ok && transport.Proxy != nil {
    fmt.Println("Proxy configuration is enabled")
}
```

## References

- Go `http.ProxyFromEnvironment` documentation: https://pkg.go.dev/net/http#ProxyFromEnvironment
- Go `http.Transport` documentation: https://pkg.go.dev/net/http#Transport
- Environment variables specification: https://about.gitlab.com/blog/2021/01/27/we-need-to-talk-no-proxy/
