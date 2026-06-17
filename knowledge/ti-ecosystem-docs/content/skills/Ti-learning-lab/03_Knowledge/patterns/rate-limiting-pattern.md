# Rate Limiting Pattern

> **Category**: Security & Performance  
> **Language**: Tiếng Việt  
> **Last Updated**: 2026-04-30

---

## Tổng Quan

Rate limiting là kỹ thuật giới hạn số lượng requests mà user hoặc API key có thể thực hiện trong một khoảng thời gian. Giúp prevent abuse, protect resources, và ensure fair usage.

---

## Khi Nào Sử Dụng

- Prevent API abuse (DDoS, spamming)
- Protect resources (database, external APIs)
- Ensure fair usage among users
- Implement tiered pricing (free vs paid tiers)
- Control costs (prevent unexpected bills)

---

## Algorithms

### 1. Fixed Window (Cố định)

**Cách hoạt động**: Reset counter mỗi X giây/phút

**Ưu điểm**:
- Dễ implement
- Dễ hiểu

**Nhược điểm**:
- "Burstiness" problem (nhiều requests ở đầu window)
- Không smooth

**Implementation**:
```go
if current_count < limit:
    current_count++
else:
    reject request
```

### 2. Sliding Window (Trượt) - Khuyến Nghị

**Cách hoạt động**: Chỉ count requests trong sliding window (last X seconds)

**Ưu điểm**:
- Smooth rate limiting
- No burstiness problem
- More accurate

**Nhược điểm**:
- Phức tạp hơn
- Cần storage (Redis)

**Implementation**:
```go
// Remove requests outside window
requests = requests.filter(r => r.timestamp > now - window_duration)
if len(requests) < limit:
    requests.append(now)
else:
    reject request
```

### 3. Token Bucket (Token Bucket)

**Cách hoạt động**: Bucket với tokens, refill rate cố định

**Ưu điểm**:
- Smooth rate limiting
- Allow bursts
- Flexible

**Nhược điểm**:
- Phức tạp
- Cần track bucket state

**Implementation**:
```go
if bucket.tokens > 0:
    bucket.tokens--
    allow request
else:
    reject request
```

### 4. Leaky Bucket

**Cách hoạt động**: Requests đi qua "leaky bucket" với fixed rate

**Ưu điểm**:
- Smooth output rate
- Prevent bursts

**Nhược điểm**:
- Không intuitive
- Harder to configure

---

## Implementation Pattern

### 1. Create Rate Limiter

```go
limiter := NewRateLimiter(
    defaultLimit,  // e.g., 100 requests per minute
    window,        // e.g., 1 minute
)
```

### 2. Check Before Request

```go
if limiter.Allow(apiKey):
    // Process request
else:
    // Reject with 429 Too Many Requests
```

### 3. Set Custom Limit

```go
limiter.SetLimit(apiKey, customLimit)
```

### 4. Get Remaining

```go
remaining := limiter.GetRemaining(apiKey)
```

---

## Best Practices

### 1. Use Sliding Window
- Sliding window là best practice cho production
- More accurate và smooth
- Use Redis cho distributed rate limiting

### 2. Per-Entity Limiting
- Rate limit per API key, user, or IP
- Different limits cho different tiers
- Document limits clearly

### 3. Graceful Degradation
- Return 429 với Retry-After header
- Suggest user upgrade plan
- Provide clear error message

### 4. Monitoring
- Track rate limit hits
- Monitor abuse patterns
- Alert cho suspicious activity

### 5. Configurable Limits
- Allow dynamic limit changes
- Different limits cho different endpoints
- Adjust limits based on load

---

## Common Mistakes

### 1. Too Strict
- **Problem**: Limits quá low, legitimate users bị block
- **Solution**: Set reasonable limits, allow burst

### 2. No Graceful Handling
- **Problem**: Return 429 không có context
- **Solution**: Include Retry-After header, clear error message

### 3. Per-IP Limiting Only
- **Problem**: Rate limit per IP dễ bypass (proxy, VPN)
- **Solution**: Rate limit per API key hoặc user ID

### 4. No Monitoring
- **Problem**: Không track rate limit hits
- **Solution**: Monitor rate limit metrics, alert cho abuse

### 5. Fixed Window
- **Problem**: Fixed window có burstiness problem
- **Solution**: Use sliding window hoặc token bucket

---

## Advanced Patterns

### 1. Distributed Rate Limiting
- Use Redis để share state across instances
- Consistent rate limiting trong distributed system
- Useful cho horizontal scaling

### 2. Tiered Rate Limiting
- Different limits cho different tiers (free, pro, enterprise)
- Higher limits cho higher tiers
- Incentivize upgrades

### 3. Adaptive Rate Limiting
- Adjust limits based on system load
- Increase limits khi load low
- Decrease limits khi load high

### 4. Rate Limiting with Backoff
- Exponential backoff cho rate limit violations
- Gradual recovery
- Useful cho APIs

---

## Tools & Services

### Open Source
- **Redis**: Distributed rate limiting
- **Nginx**: Built-in rate limiting
- **Envoy**: Advanced rate limiting

### Go Libraries
- **github.com/ulule/limiter**: Rate limiting library
- **github.com/justinas/alice**: Rate limiting middleware
- **golang.org/x/time/rate**: Go standard library

### Commercial
- **Cloudflare**: Enterprise rate limiting
- **Akamai**: Advanced rate limiting
- **AWS API Gateway**: Built-in rate limiting

---

## HTTP Headers

### Request Headers
- **X-RateLimit-Limit**: Maximum requests per window
- **X-RateLimit-Remaining**: Remaining requests
- **X-RateLimit-Reset**: Time when limit resets

### Response Headers
- **Retry-After**: Seconds to wait before retry
- **X-RateLimit-Limit**: Same as request
- **X-RateLimit-Remaining**: Same as request

---

## Ti Router Implementation

**File**: `layers/resilience/apikey_ratelimit.go`

**Features**:
- APIKeyRateLimiter với sliding window algorithm
- Per-API key rate limiting
- Configurable limits per API key
- Allow(), SetLimit(), GetRemaining(), Reset() methods
- GetStats() method
- Thread-safe với mutex

**Usage**:
```go
limiter := NewRateLimiter(100, 1*time.Minute)
if limiter.Allow(apiKey):
    // Process request
else:
    // Reject with 429
limiter.SetLimit(apiKey, customLimit)
remaining := limiter.GetRemaining(apiKey)
```

---

## References

- [Rate Limiting Best Practices](https://cloud.google.com/architecture/rate-limiting-strategies-techniques)
- [Redis Rate Limiting](https://redis.io/docs/manual/patterns/distributed-rate-limiting/)
- [Nginx Rate Limiting](https://docs.nginx.com/nginx/admin-guide/security-controls/request-limiting/)

---

## Next Steps

1. Implement distributed rate limiting với Redis
2. Add tiered rate limiting
3. Implement adaptive rate limiting
4. Add rate limiting analytics
5. Integrate với authentication system
