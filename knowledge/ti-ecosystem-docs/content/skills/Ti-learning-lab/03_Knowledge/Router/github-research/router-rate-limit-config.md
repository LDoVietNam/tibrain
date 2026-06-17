# Nghiên cứu cấu hình Rate Limit từ GitHub Repos

## Tổng quan

Tài liệu tổng hợp các mẫu cấu hình rate limiting từ các repository Go phổ biến trên GitHub.

## 1. router-for-me/CLIProxyAPI

### Cấu trúc cấu hình
```go
type SettingsConfig struct {
    Limit       int
    RedisEnabled bool
    RedisAddr   string
    RedisPrefix string
}
```

### Đặc điểm
- Sử dụng `ratelimit.Manager` để quản lý rate limiting
- Hỗ trợ cả in-memory và Redis backend
- Cấu hình qua `LoadSettingsConfig()` function
- Tích hợp với database thông qua `gorm.DB`

### Bài học
- Nên tách rate limit config thành struct riêng biệt
- Hỗ trợ Redis cho distributed rate limiting
- Có callback khi rate limit bị detect hoặc reset

## 2. lucabartmann/golang-api-gateway

### YAML Config mẫu
```yaml
rate_limit:
  enabled: true
  default_rps: 100          # requests per window per key
  window_size: 1s           # sliding window size
  key_strategy: ip            # ip | user | api_key
  local_fallback: true        # fall back to in-process token bucket if Redis fails
```

### Đặc điểm
- Sử dụng YAML config đơn giản
- Hỗ trợ multiple key strategies (IP, user, API key)
- Có fallback mechanism khi Redis fail
- Tích hợp với circuit breaker (`cb_failure_ratio`, `cb_timeout`)

### Bài học
- Nên để user cấu hình qua YAML thay vì hardcode
- `key_strategy` giúp linh hoạt cách identify user
- `local_fallback` quan trọng cho high availability

## 3. andreimerfu/pllm

### YAML Config mẫu
```yaml
rate_limit:
  enabled: true
  global_rpm: 1000           # Global requests per minute
  chat_completions_rpm: 500   # Chat completions per minute
  completions_rpm: 500        # Text completions per minute
  embeddings_rpm: 1000        # Embeddings per minute
  burst: 50                    # Allow burst of requests
  cleanup_interval: 5m          # Cleanup interval for expired entries
```

### Đặc điểm
- Rate limit per endpoint (không chỉ global)
- Hỗ trợ burst parameter
- Cleanup interval cho in-memory store
- Tích hợp với router strategy (priority, round-robin, weighted)

### Bài học
- Nên cho phép cấu hình rate limit riêng cho từng endpoint
- Burst parameter giúp xử lý traffic spikes
- Cleanup interval quan trọng để tránh memory leak

## 4. go-chi/httprate

### Code mẫu
```go
r.Use(httprate.LimitByIP(100, time.Minute))
// hoặc
r.Use(httprate.Limit(
    10,
    time.Minute,
    httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
))
```

### Đặc điểm
- Sliding window counter pattern (chính xác hơn token bucket)
- Hỗ trợ Redis backend qua `httprate-redis`
- Có `LimitCounter` interface để customize backend
- Key functions linh hoạt (`KeyByIP`, `KeyByEndpoint`, custom)

### Bài học
- Sliding window chính xác hơn token bucket cho rate limiting
- Interface-based design giúp dễ dàng mở rộng
- Nên có sẵn default backend (in-memory) và optional Redis

## 5. zalando/skipper

### Config struct
```go
type Settings struct {
    FailClosed bool           `yaml:"fail-closed"`
    Type       RatelimitType   `yaml:"type"`
    Lookuper   Lookuper        `yaml:"-"`
    MaxHits    int             `yaml:"max-hits"`
    TimeWindow time.Duration   `yaml:"time-window"`
    CleanInterval time.Duration `yaml:"-"`
    Group      string          `yaml:"group"`
}
```

### Đặc điểm
- Nhiều loại rate limiter: ServiceRatelimit, ClientRatelimit, ClusterClientRatelimit
- Hỗ trợ cluster-wide rate limiting
- Có `Group` để nhóm các rate limiter lại
- `FailClosed` option để quyết định behavior khi backend fail

### Bài học
- Cần consider cluster-wide rate limiting cho microservices
- FailClosed vs FailOpen là important design decision
- Grouping giúp quản lý rate limits phức tạp

## Kết luận và Bài học áp dụng cho Ti Router

### Cấu trúc YAML config đề xuất
```yaml
rate_limit:
  enabled: true
  user_rate_limit_rpm: 100    # Hardcoded value cần replace
  strategy: token_bucket       # hoặc sliding_window
  burst: 10
  key_strategy: ip             # ip | user
  cleanup_interval: 5m
```

### Implementation Plan cho TR-001
1. Add `UserRateLimitRPM` field to config struct
2. Load from `providers.yaml`
3. Replace hardcoded `100` in `cmd/routerd/main.go`
4. Fallback to default (100) nếu không cấu hình

### Best Practices từ GitHub
1. ✅ Luôn để config qua YAML (không hardcode)
2. ✅ Có default values hợp lý
3. ✅ Support multiple key strategies (IP, user)
4. ✅ Cleanup mechanism cho in-memory store
5. ✅ Burst support cho traffic spikes
6. ✅ Fallback mechanism khi backend fail

### References
- https://github.com/router-for-me/CLIProxyAPI
- https://github.com/lucabartmann/golang-api-gateway
- https://github.com/andreimerfu/pllm
- https://github.com/go-chi/httprate
- https://github.com/zalando/skipper

---
*Nghiên cứu bởi: claude*
*Ngày: 2026-04-30*
*Task: TR-001: Remove Hardcoded Rate Limit Values*
