---
tags: ["tibrain", "caching", "go", "documentation", "ticrew"]
scopes: ["ticrew", "resilience", "cli", "tibrain"]
last_updated: 2026-05-22
---
# Nghiên Cứu Cache Agent Patterns trong Go

> **Version**: 1.0.0  
> **Last Updated**: 2026-04-28  
> **Language**: Tiếng Việt  
> **Category**: Research  
> **Purpose**: Nghiên cứu cache patterns trong Go để implement cache-agent cho Ti router

---

## Tổng Quan

Nghiên cứu này tập trung vào việc tìm hiểu cache patterns trong Go để implement **cache-agent** cho Ti router ecosystem với các tính năng:
- Semantic cache (cache dựa trên nội dung)
- Idempotency layer (tránh duplicate requests)
- LRU cache (Least Recently Used)
- Two-tier cache (memory + persistent)

---

## 1. Internal Research - Ti Router Implementation

### 1.1 Existing Go Implementation

**Location**: `Z:\Ti\router\layers\resilience\` và `Z:\Ti\router\layers\routing\`

#### Files Chính:

| File | Mô Tả | Status |
|------|-------|--------|
| `resilience/cache.go` | Basic cache với SHA-256 key | ✅ Implemented |
| `routing/semantic_cache.go` | LRU semantic cache với token tracking | ✅ Implemented |
| `routing/cache.go` | Routing cache | ✅ Implemented |

#### Basic Cache (`resilience/cache.go`):

**Features**:
- SHA-256 request hashing
- TTL-based expiration
- Thread-safe với sync.RWMutex
- Idempotency store để track in-flight requests

**Key Functions**:
```go
type Cache struct {
    mu    sync.RWMutex
    store map[string]CachedResponse
    ttl   time.Duration
}

func (c *Cache) Get(body map[string]interface{}) *CachedResponse
func (c *Cache) Set(body map[string]interface{}, response []byte)
func (c *Cache) Invalidate(body map[string]interface{})
```

#### Semantic Cache (`routing/semantic_cache.go`):

**Features**:
- LRU (Least Recently Used) eviction policy
- Memory + size limits
- Token tracking (tokens saved)
- SHA-256 signature generation
- Cache statistics (hits, misses, tokens saved)
- Expiration support

**Key Functions**:
```go
type InMemorySemanticCache struct {
    mu     sync.RWMutex
    cache  *memoryCache
    stats  CacheStats
}

func (sc *InMemorySemanticCache) Lookup(ctx context.Context, embedding []float32) (response []byte, similarity float64, ok bool)
func (sc *InMemorySemanticCache) Store(ctx context.Context, embedding []float32, response []byte) error
func (sc *InMemorySemanticCache) StoreWithSignature(signature string, response []byte, tokensSaved int, ttl time.Duration)
func (sc *InMemorySemanticCache) Stats() CacheStats
func (sc *InMemorySemanticCache) InvalidateByModel(model string) int
```

**LRU Implementation**:
```go
type memoryCache struct {
    mu        sync.RWMutex
    items     map[string]*list.Element
    lru       *list.List
    maxSize   int
    maxBytes  int64
    curBytes  int64
}
```

### 1.2 TypeScript Reference Implementation

**Location**: `Z:\Ti\router\layers\routing\nextjs\lib\semanticCache.ts`

**Features**:
- Two-tier cache (memory LRU + SQLite)
- SHA-256 signature generation
- Auto-cleanup timer
- Cache statistics
- Model-based invalidation
- Age-based invalidation

**Key Functions**:
```typescript
export function generateSignature(model, messages, temperature = 0, topP = 1)
export function getCachedResponse(signature)
export function setCachedResponse(signature, model, response, tokensSaved, ttlMs)
export function cleanExpiredEntries()
export function invalidateByModel(model: string)
export function invalidateBySignature(signature: string)
export function invalidateStale(maxAgeMs: number)
export function getCacheStats()
export function isCacheable(body, headers)
```

**Cache Key Generation**:
```typescript
const payload = JSON.stringify({
    model,
    messages: normalizeMessages(messages),
    temperature,
    top_p: topP,
});
return crypto.createHash("sha256").update(payload).digest("hex");
```

**Two-Tier Cache Pattern**:
```typescript
// 1. Check memory cache first
const memResult = getMemoryCache().get(signature);
if (memResult) {
    stats.hits++;
    return memResult.response;
}

// 2. Check SQLite
const row = db.prepare("SELECT response, tokens_saved FROM semantic_cache WHERE signature = ? AND expires_at > datetime('now')").get(signature);
if (row) {
    // Promote to memory cache
    getMemoryCache().set(signature, { response: parsed, tokensSaved: row.tokens_saved });
    return parsed;
}

stats.misses++;
return null;
```

---

## 2. External Research - Go Cache Libraries

### 2.1 patrickmn/go-cache

**GitHub**: https://github.com/patrickmn/go-cache  
**Stars**: 8.8k  
**Description**: In-memory key:value store/cache similar to memcached

#### Features:
- Thread-safe `map[string]interface{}`
- Expiration times
- No serialization needed (in-memory)
- Safe cho multiple goroutines
- Can save/load from file
- Simple API

#### Usage:
```go
import (
    "github.com/patrickmn/go-cache"
    "time"
)

// Create cache with 5min default expiration, 10min cleanup interval
c := cache.New(5*time.Minute, 10*time.Minute)

// Set value with default expiration
c.Set("foo", "bar", cache.DefaultExpiration)

// Set value with no expiration
c.Set("baz", 42, cache.NoExpiration)

// Get value
foo, found := c.Get("foo")
if found {
    fmt.Println(foo)
}

// Delete value
c.Delete("foo")

// Get all items
items := c.Items()
```

#### Pros:
- Simple API
- Thread-safe
- Widely used (8.8k stars)
- Good documentation

#### Cons:
- No LRU eviction
- No size limits
- Basic functionality only

---

### 2.2 dgraph-io/ristretto

**GitHub**: https://github.com/dgraph-io/ristretto  
**Stars**: 6.9k  
**Description**: High performance memory-bound Go cache

#### Features:
- High performance
- Memory-bound
- Cost-based eviction
- TinyLFU admission policy
- Size-based eviction
- TTL support
- Statistics

#### Pros:
- High performance
- Advanced eviction policies
- Memory-bound
- Good for high-throughput

#### Cons:
- More complex API
- Higher memory overhead

---

### 2.3 coocood/freecache

**GitHub**: https://github.com/coocood/freecache  
**Stars**: 5.4k  
**Description**: Cache library for Go with zero GC overhead

#### Features:
- Zero GC overhead
- High performance
- Size-based eviction
- Concurrent access
- No locks for reads

#### Pros:
- Zero GC overhead
- High performance
- Good cho large caches

#### Cons:
- Fixed size allocation
- No TTL support (basic)
- More complex setup

---

### 2.4 eko/gocache

**GitHub**: https://github.com/eko/gocache  
**Stars**: 2.9k  
**Description**: Complete Go cache library with multiple backends

#### Features:
- Multiple backends (memory, Redis, Memcached)
- Load function (cache-aside pattern)
- Metrics (Prometheus)
- Marshaling (JSON, Gob)
- Expiration
- Tags (invalidation groups)

#### Pros:
- Multiple backends
- Cache-aside pattern
- Metrics support
- Flexible

#### Cons:
- More complex
- Requires external dependencies cho Redis/Memcached

---

### 2.5 maypok86/otter

**GitHub**: https://github.com/maypok86/otter  
**Stars**: 2.6k  
**Description**: High performance caching library for Go

#### Features:
- High performance
- W-TinyLFU policy
- Size-based eviction
- TTL support
- Statistics
- Sharded cache

#### Pros:
- High performance
- Advanced eviction
- Good documentation

#### Cons:
- Newer library (less battle-tested)
- More complex API

---

## 3. Best Practices cho Cache Implementation

### 3.1 Cache Key Generation

**Pattern**: Use deterministic hashing

```go
func GenerateCacheKey(model string, messages []Message, temperature float64, topP float64) string {
    payload, _ := json.Marshal(map[string]interface{}{
        "model":       model,
        "messages":    normalizeMessages(messages),
        "temperature": temperature,
        "top_p":       topP,
    })
    h := sha256.New()
    h.Write(payload)
    return fmt.Sprintf("%x", h.Sum(nil))
}

func normalizeMessages(messages []Message) []Message {
    normalized := make([]Message, 0, len(messages))
    for _, m := range messages {
        role := m.Role
        if role == "" {
            role = "user"
        }
        normalized = append(normalized, Message{
            Role:    role,
            Content: m.Content,
        })
    }
    return normalized
}
```

**Best Practices**:
- Normalize input (strip metadata, sort fields)
- Use SHA-256 cho collision resistance
- Include all relevant parameters
- Consistent encoding

---

### 3.2 LRU Eviction Policy

**Pattern**: Use doubly-linked list + hash map

```go
type LRUCache struct {
    capacity int
    cache    map[string]*list.Element
    lru      *list.List
    mutex    sync.RWMutex
}

type CacheEntry struct {
    key   string
    value interface{}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    if elem, ok := c.cache[key]; ok {
        c.lru.MoveToFront(elem)
        return elem.Value.(*CacheEntry).value, true
    }
    return nil, false
}

func (c *LRUCache) Set(key string, value interface{}) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    if elem, ok := c.cache[key]; ok {
        c.lru.MoveToFront(elem)
        elem.Value.(*CacheEntry).value = value
        return
    }
    
    elem := c.lru.PushFront(&CacheEntry{key: key, value: value})
    c.cache[key] = elem
    
    if c.lru.Len() > c.capacity {
        c.evict()
    }
}

func (c *LRUCache) evict() {
    elem := c.lru.Back()
    if elem != nil {
        c.lru.Remove(elem)
        delete(c.cache, elem.Value.(*CacheEntry).key)
    }
}
```

**Best Practices**:
- Use doubly-linked list cho O(1) move to front
- Use hash map cho O(1) lookup
- Thread-safe với mutex
- Evict when over capacity

---

### 3.3 Two-Tier Cache (Memory + Persistent)

**Pattern**: Cache-aside with promotion

```go
type TwoTierCache struct {
    memory    *LRUCache
    persistent *SQLiteCache
}

func (c *TwoTierCache) Get(key string) (interface{}, bool) {
    // 1. Check memory cache
    if value, found := c.memory.Get(key); found {
        return value, true
    }
    
    // 2. Check persistent cache
    if value, found := c.persistent.Get(key); found {
        // Promote to memory cache
        c.memory.Set(key, value)
        return value, true
    }
    
    return nil, false
}

func (c *TwoTierCache) Set(key string, value interface{}, ttl time.Duration) {
    // Set in both tiers
    c.memory.Set(key, value)
    c.persistent.Set(key, value, ttl)
}
```

**Best Practices**:
- Memory cache: fast, volatile
- Persistent cache: slower, durable
- Promote from persistent to memory on hit
- Write-through (set in both)
- Async write cho performance

---

### 3.4 Idempotency Layer

**Pattern**: Track in-flight requests

```go
type IdempotencyStore struct {
    mutex    sync.Mutex
    inFlight map[string]bool
}

func (s *IdempotencyStore) Acquire(key string) bool {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    if s.inFlight[key] {
        return false // Already in flight
    }
    
    s.inFlight[key] = true
    return true // Caller should proceed
}

func (s *IdempotencyStore) Release(key string) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    delete(s.inFlight, key)
}

// Usage
func (c *Cache) GetOrCompute(key string, compute func() (interface{}, error)) (interface{}, error) {
    // Check cache
    if value, found := c.Get(key); found {
        return value, nil
    }
    
    // Check idempotency
    if !c.idempotency.Acquire(key) {
        // Wait for other goroutine
        time.Sleep(100 * time.Millisecond)
        return c.Get(key)
    }
    defer c.idempotency.Release(key)
    
    // Compute value
    value, err := compute()
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    c.Set(key, value)
    return value, nil
}
```

**Best Practices**:
- Track in-flight requests by key
- Use mutex cho thread-safety
- Release when done
- Wait/retry pattern cho concurrent requests

---

### 3.5 Cache Statistics

**Pattern**: Track hits, misses, tokens saved

```go
type CacheStats struct {
    Hits        int64
    Misses      int64
    TokensSaved int64
}

type Cache struct {
    stats CacheStats
    mutex sync.RWMutex
}

func (c *Cache) recordHit(tokensSaved int) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    c.stats.Hits++
    c.stats.TokensSaved += int64(tokensSaved)
}

func (c *Cache) recordMiss() {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    c.stats.Misses++
}

func (c *Cache) Stats() CacheStats {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    return c.stats
}

func (c *Cache) HitRate() float64 {
    stats := c.Stats()
    total := stats.Hits + stats.Misses
    if total == 0 {
        return 0.0
    }
    return float64(stats.Hits) / float64(total) * 100.0
}
```

**Best Practices**:
- Thread-safe statistics
- Track hits/misses
- Track tokens saved (cho LLM cache)
- Calculate hit rate
- Export metrics (Prometheus)

---

### 3.6 Cache Invalidation

**Pattern**: Multiple invalidation strategies

```go
type CacheInvalidator struct {
    cache *Cache
}

// Invalidate by key
func (i *CacheInvalidator) InvalidateByKey(key string) {
    i.cache.Delete(key)
}

// Invalidate by model
func (i *CacheInvalidator) InvalidateByModel(model string) int {
    // Iterate and delete all entries for this model
    count := 0
    for key := range i.cache.store {
        if strings.HasPrefix(key, model+":") {
            i.cache.Delete(key)
            count++
        }
    }
    return count
}

// Invalidate by age
func (i *CacheInvalidator) InvalidateByAge(maxAge time.Duration) int {
    count := 0
    now := time.Now()
    
    for key, entry := range i.cache.store {
        if now.Sub(entry.CreatedAt) > maxAge {
            i.cache.Delete(key)
            count++
        }
    }
    return count
}

// Invalidate expired (auto-cleanup)
func (i *CacheInvalidator) CleanExpired() int {
    count := 0
    for key, entry := range i.cache.store {
        if entry.IsExpired() {
            i.cache.Delete(key)
            count++
        }
    }
    return count
}
```

**Best Practices**:
- Multiple invalidation strategies
- Auto-cleanup expired entries
- Model-based invalidation
- Age-based invalidation
- Manual invalidation by key

---

## 4. Recommendations cho Ti Cache-Agent

### 4.1 Library Selection

**Recommendation**: Sử dụng **patrickmn/go-cache** cho simplicity hoặc **ristretto** cho performance

**Reasons cho patrickmn/go-cache**:
- Simple API
- Thread-safe
- Widely used (8.8k stars)
- Good documentation
- Fits current Ti implementation

**Reasons cho ristretto**:
- High performance
- Advanced eviction policies
- Memory-bound
- Better cho high-throughput scenarios

### 4.2 Architecture Recommendations

#### Layer Structure:

```
cache-agent/
├── semantic/
│   ├── cache.go              // Semantic cache interface
│   ├── memory.go             // Memory LRU cache
│   ├── persistent.go         // SQLite persistent cache
│   ├── two_tier.go           // Two-tier cache (memory + persistent)
│   ├── signature.go          // Cache key generation
│   └── invalidator.go        // Cache invalidation
├── idempotency/
│   ├── store.go              // Idempotency store
│   └── middleware.go         // Idempotency middleware
├── basic/
│   ├── cache.go              // Basic cache interface
│   └── lru.go                // LRU cache implementation
├── stats/
│   ├── collector.go          // Statistics collector
│   └── metrics.go            // Prometheus metrics
└── maintenance/
    ├── cleanup.go            // Auto-cleanup
    └── monitor.go            // Background monitor
```

### 4.3 Implementation Priorities

**Priority 1 (Critical)**:
1. Two-tier semantic cache (memory + SQLite)
2. LRU eviction policy
3. Cache key generation (SHA-256)
4. Basic idempotency layer

**Priority 2 (Important)**:
1. Cache statistics (hits, misses, tokens saved)
2. Auto-cleanup expired entries
3. Model-based invalidation
4. Prometheus metrics export

**Priority 3 (Enhancement)**:
1. Advanced eviction policies (LFU, ARC)
2. Cache warming
3. Distributed cache (Redis)
4. Machine learning-based cache prediction

### 4.4 Performance Recommendations

**Cache Sizing**:
- Memory cache: 100-1000 entries, 4-16MB
- Persistent cache: Unlimited (SQLite)
- TTL: 30 minutes - 24 hours (configurable)
- LRU capacity: Based on memory limit

**Concurrency**:
- Use sync.RWMutex cho read-heavy workloads
- Consider sharding cho high concurrency
- Use atomic operations cho statistics

**Persistence**:
- SQLite cho simplicity
- Async writes cho performance
- Batch inserts cho efficiency
- WAL mode cho concurrency

---

## 5. Code Examples

### 5.1 Two-Tier Semantic Cache

```go
package semantic

import (
    "context"
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "sync"
    "time"
    
    "github.com/patrickmn/go-cache"
)

type TwoTierCache struct {
    memory    *cache.Cache
    persistent *SQLiteCache
    mutex     sync.RWMutex
}

func NewTwoTierCache(memoryTTL, cleanupInterval time.Duration) *TwoTierCache {
    return &TwoTierCache{
        memory:    cache.New(memoryTTL, cleanupInterval),
        persistent: NewSQLiteCache(),
    }
}

func (c *TwoTierCache) Get(ctx context.Context, signature string) ([]byte, bool) {
    // 1. Check memory cache
    if value, found := c.memory.Get(signature); found {
        if data, ok := value.([]byte); ok {
            return data, true
        }
    }
    
    // 2. Check persistent cache
    if data, found := c.persistent.Get(signature); found {
        // Promote to memory cache
        c.memory.Set(signature, data, cache.DefaultExpiration)
        return data, true
    }
    
    return nil, false
}

func (c *TwoTierCache) Set(ctx context.Context, signature string, data []byte, ttl time.Duration) error {
    // Set in memory cache
    c.memory.Set(signature, data, ttl)
    
    // Set in persistent cache (async)
    go func() {
        c.persistent.Set(signature, data, ttl)
    }()
    
    return nil
}
```

### 5.2 LRU Cache with Size Limits

```go
package lru

import (
    "container/list"
    "sync"
)

type LRUCache struct {
    capacity  int
    maxBytes  int64
    curBytes  int64
    items     map[string]*list.Element
    lru       *list.List
    mutex     sync.RWMutex
}

type CacheItem struct {
    key   string
    value []byte
}

func NewLRUCache(capacity int, maxBytes int64) *LRUCache {
    return &LRUCache{
        capacity: capacity,
        maxBytes: maxBytes,
        items:    make(map[string]*list.Element),
        lru:      list.New(),
    }
}

func (c *LRUCache) Get(key string) ([]byte, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    if elem, ok := c.items[key]; ok {
        c.mutex.Lock()
        c.lru.MoveToFront(elem)
        c.mutex.Unlock()
        return elem.Value.(*CacheItem).value, true
    }
    
    return nil, false
}

func (c *LRUCache) Set(key string, value []byte) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    // Update existing
    if elem, ok := c.items[key]; ok {
        old := elem.Value.(*CacheItem)
        c.curBytes += int64(len(value)) - int64(len(old.value))
        old.value = value
        c.lru.MoveToFront(elem)
    } else {
        // Add new
        item := &CacheItem{key: key, value: value}
        elem := c.lru.PushFront(item)
        c.items[key] = elem
        c.curBytes += int64(len(value))
    }
    
    // Evict if over limits
    for c.lru.Len() > c.capacity || c.curBytes > c.maxBytes {
        c.evict()
    }
}

func (c *LRUCache) evict() {
    elem := c.lru.Back()
    if elem == nil {
        return
    }
    
    item := c.lru.Remove(elem).(*CacheItem)
    delete(c.items, item.key)
    c.curBytes -= int64(len(item.value))
}
```

### 5.3 Idempotency Middleware

```go
package idempotency

import (
    "crypto/sha256"
    "encoding/json"
    "net/http"
    "sync"
    "time"
)

type IdempotencyMiddleware struct {
    store     *IdempotencyStore
    cache     *Cache
    mutex     sync.Mutex
}

func (m *IdempotencyMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Generate idempotency key from request body
        key := m.generateKey(r)
        
        // Check if request is already in flight
        if !m.store.Acquire(key) {
            // Wait and retry
            time.Sleep(100 * time.Millisecond)
            if value, found := m.cache.Get(key); found {
                w.Write(value.([]byte))
                return
            }
            http.Error(w, "Request in progress", http.StatusConflict)
            return
        }
        defer m.store.Release(key)
        
        // Process request
        next.ServeHTTP(w, r)
    })
}

func (m *IdempotencyMiddleware) generateKey(r *http.Request) string {
    // Read body
    body := readBody(r)
    
    // Generate hash
    h := sha256.New()
    h.Write([]byte(r.Method))
    h.Write([]byte(r.URL.Path))
    h.Write(body)
    return fmt.Sprintf("%x", h.Sum(nil))
}
```

---

## 6. References

### GitHub Repositories:
- **patrickmn/go-cache**: https://github.com/patrickmn/go-cache
- **dgraph-io/ristretto**: https://github.com/dgraph-io/ristretto
- **coocood/freecache**: https://github.com/coocood/freecache
- **eko/gocache**: https://github.com/eko/gocache
- **maypok86/otter**: https://github.com/maypok86/otter

### Ti Router Files:
- **resilience/cache.go**: Z:\Ti\router\layers\resilience\cache.go
- **routing/semantic_cache.go**: Z:\Ti\router\layers\routing\semantic_cache.go
- **semanticCache.ts**: Z:\Ti\router\layers\routing\nextjs\lib\semanticCache.ts

---

*Last Updated: 2026-04-28*
*Research completed by: Claude*
*Purpose: Implement cache-agent cho Ti router ecosystem*
