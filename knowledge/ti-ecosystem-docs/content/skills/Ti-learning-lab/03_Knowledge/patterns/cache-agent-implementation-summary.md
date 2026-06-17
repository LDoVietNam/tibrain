---
tags: ["tibrain", "caching", "documentation", "ticrew", "skill"]
scopes: ["ticrew", "resilience", "tibrain"]
last_updated: 2026-05-22
---
# Cache-Agent Implementation Summary (Tiếng Việt)

> **Version**: 1.0.0  
> **Last Updated**: 2026-04-29  
> **Purpose**: Tóm tắt implementation của cache-agent cho Ti router

---

## Tổng Quan

Cache-agent đã được implement với các components sau:

### 1. Basic Cache (Đã tồn tại trong cache.go)

**File**: `Z:\Ti\router\layers\resilience\cache.go`

**Features**:
- Semantic cache với SHA-256 key generation
- TTL-based expiration
- In-memory storage
- Thread-safe với mutex

**Usage**:
```go
cache := NewCache(5 * time.Minute)
cache.Set(requestBody, response)
result := cache.Get(requestBody)
```

### 2. Idempotency Store (Đã tồn tại trong cache.go)

**File**: `Z:\Ti\router\layers\resilience\cache.go`

**Features**:
- Tracks in-flight requests
- Prevents duplicate concurrent upstream calls
- Thread-safe với mutex

**Usage**:
```go
store := NewIdempotencyStore()
if store.Acquire(requestBody) {
    // Proceed with upstream call
    defer store.Release(requestBody)
}
```

### 3. LRU Cache (Mới - cache_lru.go)

**File**: `Z:\Ti\router\layers\resilience\cache_lru.go`

**Features**:
- LRU (Least Recently Used) eviction policy
- Configurable capacity
- Automatic eviction khi capacity reached
- Moves accessed entries to front (mark as recently used)
- Thread-safe với mutex
- Statistics tracking (hit rate, hits, misses)

**Usage**:
```go
// Basic LRU cache
cache := NewLRUCache(1000, 5*time.Minute)
cache.Set(requestBody, response)
result := cache.Get(requestBody)

// With statistics
statsCache := NewStatsLRUCache(1000, 5*time.Minute)
result := statsCache.Get(requestBody)
stats := statsCache.Stats()
fmt.Printf("Hit rate: %.2f%%\n", stats.HitRate*100)
```

**Key Methods**:
- `Get(body)` - Lookup cache entry, marks as recently used
- `Set(body, response)` - Store cache entry, evicts if at capacity
- `Invalidate(body)` - Remove specific entry
- `Clear()` - Remove all entries
- `Size()` - Current number of entries
- `Stats()` - Cache statistics (for StatsLRU)
- `ResetStats()` - Reset hit/miss counters

### 4. Tiered Cache (Mới - cache_tiered.go)

**File**: `Z:\Ti\router\layers\resilience\cache_tiered.go`

**Features**:
- Two-tier cache: memory (LRU) + SQLite backend
- Memory-first lookup for speed
- SQLite fallback for persistence
- Writes to both tiers for consistency
- Configurable sync/async writes
- Automatic cleanup of expired entries
- Database optimization (VACUUM)

**Usage**:
```go
// Create tiered cache
cache, err := NewTieredCache(1000, 5*time.Minute, "cache.db")
if err != nil {
    log.Fatal(err)
}
defer cache.Close()

// Use like regular cache
cache.Set(requestBody, response)
result := cache.Get(requestBody)

// Configure
cache.SetSyncMode(true) // Synchronous writes (safer but slower)

// Maintenance
cache.Cleanup() // Remove expired entries
cache.Vacuum()  // Optimize database

// Stats
memSize := cache.Size()
dbSize, _ := cache.DBSize()
fmt.Printf("Memory: %d, DB: %d\n", memSize, dbSize)
```

**Key Methods**:
- `Get(body)` - Lookup in memory, then SQLite
- `Set(body, response)` - Write to both tiers
- `Invalidate(body)` - Remove from both tiers
- `Clear()` - Clear both tiers
- `Size()` - Memory cache size
- `DBSize()` - SQLite cache size
- `Cleanup()` - Remove expired entries from SQLite
- `SetSyncMode(sync)` - Configure write mode
- `Close()` - Close database connection
- `Vacuum()` - Optimize SQLite database

---

## Architecture

```
resilience/
├── cache.go              # Basic cache + idempotency store (đã tồn tại)
├── cache_lru.go          # LRU cache with statistics (MỚI)
└── cache_tiered.go       # Two-tier cache (memory + SQLite) (MỚI)
```

---

## Key Patterns

### 1. Semantic Cache Pattern

```go
// SHA-256 key generation from request body
func hashRequest(body map[string]interface{}) string {
    b, _ := json.Marshal(body)
    return fmt.Sprintf("%x", sha256.Sum256(b))
}
```

### 2. LRU Eviction Pattern

- Use `container/list` for O(1) eviction
- Move accessed items to front
- Evict from back when at capacity

### 3. Two-Tier Cache Pattern

- Memory cache for hot data (fast)
- SQLite for persistence (survives restarts)
- Write-through to both tiers
- Read-through from memory first, then SQLite

---

## Performance Considerations

### LRU Cache
- **Get**: O(1) - hash lookup + list move
- **Set**: O(1) - hash insert + list push
- **Eviction**: O(1) - list remove from back
- **Memory**: O(capacity) - fixed size

### Tiered Cache
- **Memory Get**: O(1) - same as LRU
- **SQLite Get**: O(log n) - indexed lookup
- **Memory Set**: O(1) - same as LRU
- **SQLite Set**: O(log n) - indexed insert (async by default)
- **Memory**: O(memoryCapacity) - fixed size
- **Disk**: O(unbounded) - limited by disk space

---

## Testing

Build test:
```bash
cd /z/Ti/router/layers/resilience
go build -o /dev/null .
```

Result: ✅ Build succeeded

---

## Configuration Recommendations

### Development
- Use basic cache or LRU cache
- In-memory only (no SQLite)
- Capacity: 100-1000 entries
- TTL: 5-10 minutes

### Production
- Use tiered cache
- Memory capacity: 1000-10000 entries
- SQLite for persistence
- TTL: 10-60 minutes
- Async writes for performance
- Periodic cleanup (every hour)

### High Traffic
- Use tiered cache with larger memory
- Memory capacity: 10000+ entries
- Sync writes for consistency
- Shorter TTL (5-10 minutes)
- Frequent cleanup (every 15-30 minutes)

---

## Integration with Router

The cache can be integrated into the router's request handling:

```go
// In router handler
cache := NewTieredCache(1000, 5*time.Minute, "cache.db")

// Check cache first
if cached := cache.Get(requestBody); cached != nil {
    return cached.Body // Return cached response
}

// Acquire idempotency lock
if !idempotencyStore.Acquire(requestBody) {
    return nil // Another goroutine is handling this
}
defer idempotencyStore.Release(requestBody)

// Make upstream call
response := callUpstream(requestBody)

// Cache the response
cache.Set(requestBody, response)

return response
```

---

## Next Steps

1. **Metrics Integration** - Add Prometheus metrics for cache hit rate, size, etc.
2. **Cache Warming** - Pre-populate cache with common requests
3. **Cache Invalidation** - Add invalidation by pattern or prefix
4. **Distributed Cache** - Consider Redis for multi-instance deployments
5. **Compression** - Compress cached responses to save memory/disk
6. **TTL Per Key** - Support different TTLs for different request types

---

## Lessons Learned

1. **LRU Implementation** - Using `container/list` provides O(1) operations for both access and eviction
2. **Two-Tier Trade-offs** - Async writes improve performance but risk data loss on crash
3. **SQLite Optimization** - Regular VACUUM needed to prevent database bloat
4. **Thread Safety** - Mutex is essential for concurrent access to shared cache
5. **Statistics Value** - Hit rate monitoring helps optimize cache capacity and TTL

---

## Comparison with Research

### Research Findings
- **patrickmn/go-cache** (8.8k stars): Simple in-memory cache with expiration
- **ristretto** (6.9k stars): High-performance cache with admission policies

### Our Implementation
- **Basic Cache**: Similar to go-cache, simpler interface
- **LRU Cache**: Custom LRU with statistics (not in go-cache)
- **Tiered Cache**: Two-tier design (not in ristretto or go-cache)
- **Idempotency**: Built-in idempotency layer (unique to our implementation)

---

*Last Updated: 2026-04-29*  
*Created by: Claude*  
*Purpose: Cache-agent implementation summary*
