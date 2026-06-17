---
tags: ["tibrain", "documentation", "provider", "skill", "pattern-auth-cookie"]
scopes: ["tibrain"]
last_updated: 2026-05-22
---
# Cookie Provider Optimizations

> **Date**: 2026-04-28  
> **Component**: Ti Router - Cookie Provider  
> **Status**: ✅ Implemented & Tested

## Overview

Cookie provider tại `Z:\Ti\router\layers\provider\cookie\` đã được tối ưu hóa với 3 cải tiến chính:

1. **LRU Cache** - Thay thế simple map với LRU cache có size-based eviction
2. **Cookie Auto-Refresh** - Background refresh cho cookies sắp hết hạn
3. **Context Support** - Thêm context.Context cho cancellation support

---

## Optimizations Implemented

### 1. LRU Cache (cache_lru.go)

**Problem**: HTTPClient sử dụng `map[string]*cacheEntry` đơn giản, cache tăng không giới hạn → memory leak.

**Solution**: Implement LRU cache với:
- Entry limit (default: 1000 entries)
- Byte limit (default: 100MB)
- Automatic eviction của least recently used entries
- Thread-safe với mutex

**File**: `Z:\Ti\router\layers\provider\cookie\cache_lru.go`

**Key Features**:
```go
type LRUCache struct {
    maxEntries int      // Maximum number of entries
    maxBytes   int64    // Maximum total bytes
    entries    map[string]*list.Element
    lruList    *list.List  // Doubly-linked list for LRU tracking
    currentBytes int64
}
```

**API**:
- `Get(key)` - Retrieve value, move to front (most recently used)
- `Set(key, value, headers, status, ttl)` - Add/update entry with TTL
- `Delete(key)` - Remove specific entry
- `Clear()` - Remove all entries
- `Len()` - Number of entries
- `Size()` - Total bytes in cache

**Integration**: HTTPClient.go updated để sử dụng LRUCache thay vì simple map.

**Impact**:
- ✅ Prevents memory leaks under heavy usage
- ✅ Better cache hit rate with intelligent eviction
- ✅ Configurable limits per workload

**Tests**: 7 tests trong `cache_lru_test.go` (basic, byte limit, expiry, update, delete, clear, concurrent)

---

### 2. Cookie Auto-Refresh (cookie_refresh.go)

**Problem**: Cookies expire mid-session, không có automatic refresh → service interruption.

**Solution**: Background goroutine refresh cookies trước khi hết hạn.

**File**: `Z:\Ti\router\layers\provider\cookie\cookie_refresh.go`

**Key Features**:
```go
type CookieRefreshService struct {
    mgr         *Manager
    profile     string
    refreshURL  string
    refreshFunc func(*http.Response) ([]*Cookie, error)
    
    checkInterval time.Duration  // Default: 1 minute
    refreshBefore time.Duration  // Default: 5 minutes before expiry
}
```

**API**:
- `NewCookieRefreshService(mgr, profile, refreshURL, refreshFunc)` - Constructor
- `WithCheckInterval(interval)` - Configure check frequency
- `WithRefreshBefore(duration)` - Configure refresh threshold
- `Start()` - Start background goroutine
- `Stop()` - Graceful shutdown
- `IsRunning()` - Check service status

**Workflow**:
1. Check cookies every `checkInterval` (default: 1 minute)
2. If any cookie expires within `refreshBefore` (default: 5 minutes), trigger refresh
3. Hit `refreshURL` with current cookies
4. Extract new cookies using `refreshFunc`
5. Update profile with new cookies

**Impact**:
- ✅ Prevents session interruptions from expired cookies
- ✅ Proactive refresh before expiry
- ✅ Configurable refresh logic per service

**Tests**: 4 tests trong `cookie_refresh_test.go` (start/stop, expiring cookie, fresh cookie, missing profile)

---

### 3. Context Support (cookie.go)

**Problem**: HTTPClient không support context → không thể cancel requests, no timeout control.

**Solution**: Thêm `GetWithContext` và `PostWithContext` methods.

**Changes**:
```go
// New methods with context support
func (hc *HTTPClient) GetWithContext(ctx context.Context, url string) (*http.Response, error)
func (hc *HTTPClient) PostWithContext(ctx context.Context, url, contentType string, body []byte) (*http.Response, error)

// Existing methods now delegate to context versions
func (hc *HTTPClient) Get(url string) (*http.Response, error) {
    return hc.GetWithContext(context.Background(), url)
}
```

**Impact**:
- ✅ Request cancellation support
- ✅ Per-request timeout control
- ✅ Better integration with Go's context propagation
- ✅ Backward compatible (existing methods unchanged)

---

## Test Results

**Total Tests**: 63 tests  
**Status**: ✅ All passing

```
=== RUN   TestLRUCacheBasic
--- PASS: TestLRUCacheBasic (0.00s)
=== RUN   TestLRUCacheByteLimit
--- PASS: TestLRUCacheByteLimit (0.00s)
=== RUN   TestLRUCacheExpiry
--- PASS: TestLRUCacheExpiry (0.10s)
=== RUN   TestLRUCacheUpdate
--- PASS: TestLRUCacheUpdate (0.00s)
=== RUN   TestLRUCacheDelete
--- PASS: TestLRUCacheDelete (0.00s)
=== RUN   TestLRUCacheClear
--- PASS: TestLRUCacheClear (0.00s)
=== RUN   TestLRUCacheConcurrent
--- PASS: TestLRUCacheConcurrent (0.00s)
=== RUN   TestCookieRefreshServiceStartStop
--- PASS: TestCookieRefreshServiceStartStop (0.20s)
=== RUN   TestCookieRefreshServiceExpiringCookie
--- PASS: TestCookieRefreshServiceExpiringCookie (0.30s)
=== RUN   TestCookieRefreshServiceFreshCookie
--- PASS: TestCookieRefreshServiceFreshCookie (0.30s)
=== RUN   TestCookieRefreshServiceMissingProfile
--- PASS: TestCookieRefreshServiceMissingProfile (0.20s)
... (53 existing tests)
PASS
ok      github.com/ti/router/layers/provider/cookie    2.092s
```

---

## Files Created/Modified

### New Files
1. `Z:\Ti\router\layers\provider\cookie\cache_lru.go` - LRU cache implementation (145 lines)
2. `Z:\Ti\router\layers\provider\cookie\cache_lru_test.go` - LRU cache tests (126 lines)
3. `Z:\Ti\router\layers\provider\cookie\cookie_refresh.go` - Cookie refresh service (157 lines)
4. `Z:\Ti\router\layers\provider\cookie\cookie_refresh_test.go` - Refresh service tests (134 lines)

### Modified Files
1. `Z:\Ti\router\layers\provider\cookie\cookie.go` - Updated HTTPClient to use LRU cache, added context support

---

## Usage Examples

### Using LRU Cache (Automatic)

LRU cache được sử dụng tự động trong HTTPClient:

```go
mgr := cookie.NewManager()
mgr.LoadChromeJSON(chromeCookies, "notion")

hc := cookie.NewHTTPClient(mgr, "notion")
// Cache automatically uses LRU with:
// - Max 1000 entries
// - Max 100MB total size
// - 5 minute TTL

resp, err := hc.Get("https://api.notion.com/v1/users")
// Response cached automatically for GET requests
```

### Custom Cache Configuration

```go
hc := cookie.NewHTTPClient(mgr, "notion")
// Override default cache limits
hc.cache = cookie.NewLRUCache(500, 50*1024*1024) // 500 entries, 50MB
```

### Using Cookie Auto-Refresh

```go
mgr := cookie.NewManager()
mgr.LoadChromeJSON(chromeCookies, "notion")

// Define refresh logic
refreshFunc := func(resp *http.Response) ([]*cookie.Cookie, error) {
    // Extract cookies from response
    cookies := extractCookiesFromResponse(resp)
    return cookies, nil
}

// Create refresh service
service := cookie.NewCookieRefreshService(
    mgr,
    "notion",
    "https://www.notion.so/",
    refreshFunc,
)

// Configure refresh timing
service.WithCheckInterval(30 * time.Second)  // Check every 30 seconds
service.WithRefreshBefore(10 * time.Minute) // Refresh 10 minutes before expiry

// Start background refresh
service.Start()
defer service.Stop()
```

### Using Context Support

```go
mgr := cookie.NewManager()
hc := cookie.NewHTTPClient(mgr, "notion")

// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := hc.GetWithContext(ctx, "https://api.notion.com/v1/users")
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout
    }
}

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    // Cancel after 2 seconds
    time.Sleep(2 * time.Second)
    cancel()
}()

resp, err := hc.GetWithContext(ctx, "https://api.notion.com/v1/users")
```

---

## Future Optimizations (Not Implemented)

The following optimizations were identified but not implemented in this phase:

### Security Improvements
1. **Cookie Integrity Verification** - Add HMAC-SHA256 signature on cookie sets
2. **Key Rotation** - Add key versioning and rotation support
3. **Memory Encryption** - Keep sensitive cookies encrypted in memory

### Performance Improvements
4. **Cookie Decryption Caching** - Cache decrypted profiles in memory
5. **Connection Pool Tuning** - Add adaptive connection pool scaling
6. **Rate Limiting Precision** - Leaky bucket with nanosecond precision

### Reliability Improvements
7. **Health Monitoring** - Add health checks and metrics
8. **Graceful Degradation** - Fallback mechanisms for critical failures
9. **Structured Logging** - Replace fmt.Fprintf with structured logging

---

## Lessons Learned

1. **LRU Implementation**: Sử dụng `container/list` (doubly-linked list) + map là pattern hiệu quả cho LRU cache
2. **Thread Safety**: Mutex protection cần thiết cho cache operations trong concurrent environment
3. **Background Services**: Cookie refresh service cần graceful shutdown với `stopChan`
4. **Context Propagation**: Adding context support là backward-compatible nếu delegate từ existing methods
5. **Test Coverage**: Concurrent tests cần thiết để verify thread-safety

---

## References

- **LRU Cache Pattern**: https://github.com/golang/groupcache/tree/master/lru
- **Context Best Practices**: https://go.dev/blog/context
- **Cookie Refresh Pattern**: Similar to `TokenRefreshService` in auth-agent
