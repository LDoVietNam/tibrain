---
tags: ["tibrain", "caching", "documentation", "skill", "performance"]
scopes: ["resilience", "tibrain"]
last_updated: 2026-05-22
---
# Caching Pattern

> **Category**: Performance  
> **Language**: Tiếng Việt  
> **Last Updated**: 2026-04-30

---

## Tổng Quan

Caching là kỹ thuật store kết quả của expensive operations để reuse sau này. Giảm latency, reduce load, và improve throughput.

---

## Khi Nào Sử Dụng

- Expensive database queries
- External API calls
- Computationally expensive operations
- Frequently accessed data
- Read-heavy workloads

---

## Cache Types

### 1. In-Memory Cache
- Store trong application memory
- Fastest access (nanoseconds)
- Limited size (RAM)
- Not persistent (lost on restart)

**Use Cases**:
- Hot data (frequently accessed)
- Session data
- Temporary computations

**Tools**: Go map, sync.Map, BigCache

### 2. Redis Cache
- Distributed cache server
- Fast access (microseconds)
- Persistent (optional)
- Scalable horizontally

**Use Cases**:
- Distributed systems
- Shared cache across instances
- Persistent cache
- Large datasets

**Tools**: Redis, Memcached

### 3. CDN Cache
- Content Delivery Network
- Edge caching
- Global distribution
- Static content

**Use Cases**:
- Static assets (images, CSS, JS)
- API responses (GET requests)
- Geographic distribution

**Tools**: Cloudflare CDN, AWS CloudFront, Fastly

### 4. Database Cache
- Query result cache
- Built-in database caching
- Automatic invalidation

**Use Cases**:
- Frequently executed queries
- Read-heavy workloads
- Complex joins

**Tools**: PostgreSQL query cache, MySQL query cache

---

## Cache Invalidation Strategies

### 1. TTL (Time-To-Live)
- Cache entries expire sau fixed time
- Simple, easy implement
- **Problem**: Stale data

**Implementation**:
```go
cache.Set(key, value, 5*time.Minute) // Expire sau 5 phút
```

### 2. Manual Invalidation
- Explicitly invalidate cache khi data changes
- Accurate, no stale data
- **Problem**: Complex, need track all dependencies

**Implementation**:
```go
cache.Delete(key)
cache.InvalidatePattern("user:*") // Invalidate all keys matching pattern
```

### 3. Write-Through
- Write to cache và database simultaneously
- Cache luôn fresh
- **Problem**: Slower writes

**Implementation**:
```go
db.write(data)
cache.set(key, data)
```

### 4. Write-Behind
- Write to cache immediately, database asynchronously
- Fast writes
- **Problem**: Risk of inconsistency

**Implementation**:
```go
cache.set(key, data)
async_queue.push(() => db.write(data))
```

### 5. Write-Around
- Write to database only, skip cache
- Cache populated on read
- **Problem**: Cache miss on first read

**Implementation**:
```go
db.write(data)
// Cache not updated
```

---

## Cache Key Design

### Best Practices

**1. Descriptive Keys**
```
user:123:profile
product:456:details
api:openai:gpt4:cache
```

**2. Hierarchical Keys**
```
user:123:profile
user:123:settings
user:123:permissions
```

**3. Include Versioning**
```
api:v1:user:123:profile
api:v2:user:123:profile
```

**4. Include Context**
```
cache:{provider}:{model}:{prompt_hash}
```

### Key Generation Pattern

```go
key = fmt.Sprintf("%s:%s:%s", prefix, id, version)
key = hashString(provider + model + prompt)
```

---

## Cache Statistics

### Metrics Quan Trọng

**1. Hit Rate**
```
Hit Rate = Cache Hits / Total Requests
```
- Target: > 80%
- Higher = better

**2. Miss Rate**
```
Miss Rate = Cache Misses / Total Requests
```
- Target: < 20%
- Lower = better

**3. Latency**
- Cache hit latency: < 1ms (in-memory), < 10ms (Redis)
- Cache miss latency: Database/API latency

**4. Size**
- Cache size (bytes)
- Number of entries
- Memory usage

---

## Implementation Pattern

### 1. Create Cache

```go
cache := NewRedisCache(redis_client)
```

### 2. Get with Cache-Aside Pattern

```go
value, err := cache.Get(ctx, key)
if err == nil {
    // Cache hit
    return value
}

// Cache miss - fetch from source
value = fetchFromSource(key)
cache.Set(ctx, key, value, ttl)
return value
```

### 3. Set with TTL

```go
cache.Set(ctx, key, value, ttl)
```

### 4. Invalidate

```go
cache.Delete(ctx, key)
cache.InvalidatePattern(ctx, "user:*")
```

### 5. Get Statistics

```go
stats := cache.GetStats()
hit_rate := stats.Hits / (stats.Hits + stats.Misses)
```

---

## Best Practices

### 1. Cache-Aside Pattern
- Check cache first
- If miss, fetch from source
- Update cache
- Most common pattern

### 2. Appropriate TTL
- Short TTL cho frequently changing data (seconds/minutes)
- Long TTL cho rarely changing data (hours/days)
- Never use infinite TTL

### 3. Cache Warming
- Pre-populate cache với hot data
- Load cache on startup
- Background refresh

### 4. Monitoring
- Track hit rate, miss rate, latency
- Set up alerts cho low hit rate
- Monitor cache size

### 5. Eviction Policy
- LRU (Least Recently Used) - common
- LFU (Least Frequently Used) - cho stable workloads
- TTL-based - simple
- Size-based - limit memory usage

---

## Common Mistakes

### 1. Caching Wrong Data
- **Problem**: Cache data không nên cache (user-specific, real-time)
- **Solution**: Chỉ cache appropriate data

### 2. No Invalidation
- **Problem**: Cache không invalidate khi data changes
- **Solution**: Implement proper invalidation strategy

### 3. Too Long TTL
- **Problem**: TTL quá dài, stale data
- **Solution**: Use appropriate TTL based on data freshness requirements

### 4. Cache Stampede
- **Problem**: Many requests miss cache simultaneously, cause thundering herd
- **Solution**: Use single-flight, cache warming

### 5. No Monitoring
- **Problem**: Không track cache metrics
- **Solution**: Monitor hit rate, miss rate, latency

---

## Advanced Patterns

### 1. Multi-Level Cache
- L1: In-memory cache (fastest)
- L2: Redis cache (fast)
- L3: Database/API (slowest)
- Check L1 → L2 → L3 sequentially

### 2. Cache Warming
- Pre-populate cache với hot data
- Load cache on startup
- Background refresh

### 3. Cache Partitioning
- Partition cache by user/tenant
- Reduce cache contention
- Improve scalability

### 4. Cache Compression
- Compress cache entries cho large data
- Reduce memory usage
- Trade-off: CPU vs memory

---

## Tools & Libraries

### Go Libraries
- **github.com/go-redis/redis**: Redis client cho Go
- **github.com/allegro/bigcache**: High-performance in-memory cache
- **github.com/patrickmn/go-cache**: Simple in-memory cache

### Cache Servers
- **Redis**: In-memory data structure store
- **Memcached**: Distributed memory object caching system

### CDN Services
- **Cloudflare CDN**: Global CDN
- **AWS CloudFront**: AWS CDN
- **Fastly**: Edge cloud platform

---

## Comparison: In-Memory vs Redis

| Aspect | In-Memory | Redis |
|--------|-----------|-------|
| Speed | Fastest (nanoseconds) | Fast (microseconds) |
| Persistence | No | Yes (optional) |
| Scalability | Single instance | Distributed |
| Size | Limited by RAM | Large (disk-backed) |
| Complexity | Simple | More complex |
| Use Case | Single-server, hot data | Distributed, shared cache |

---

## Ti Router Implementation

**File**: `layers/resilience/redis_cache.go`

**Features**:
- RedisCache struct với in-memory storage (MVP)
- Get(), Set(), Delete() methods với context support
- InvalidatePattern() method cho pattern-based invalidation
- CacheWithStats wrapper với hits/misses tracking
- GetStats() method với hit rate calculation
- CacheKey struct với Provider, Model, Prompt, Params
- GenerateKey() method cho cache key generation
- Thread-safe với mutex

**Usage**:
```go
cache := NewRedisCache()
value, err := cache.Get(ctx, key)
cache.Set(ctx, key, value, ttl)
cache.Delete(ctx, key)
cache.InvalidatePattern(ctx, "user:*")
stats := cache.GetStats()
```

---

## References

- [Redis Caching Best Practices](https://redis.io/docs/manual/patterns/)
- [Cache Aside Pattern](https://docs.aws.amazon.com/elasticache/latest/memcached/GettingStarted.Strategies.html)
- [Google Cloud Cache Best Practices](https://cloud.google.com/architecture/in-memory-data-caching)

---

## Next Steps

1. Implement actual Redis integration (MVP: in-memory)
2. Add multi-level cache
3. Implement cache warming
4. Add cache partitioning
5. Implement cache compression
