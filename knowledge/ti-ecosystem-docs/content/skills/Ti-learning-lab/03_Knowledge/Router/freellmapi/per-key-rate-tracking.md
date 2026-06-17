# Per-key Rate Tracking (RPM/RPD/TPM/TPD)

> **Feature**: Theo dõi rate limits per-key với sliding window
> **Nguồn cảm hứng**: FreeLLMAPI - `server/src/services/ratelimit.ts`
> **Trạng thái**: ✅ Hoàn thành

## Tổng Quan

Ti Router hiện hỗ trợ per-key rate tracking với 4 metrics:
- **RPM** (Requests Per Minute) - Số requests mỗi phút
- **RPD** (Requests Per Day) - Số requests mỗi ngày
- **TPM** (Tokens Per Minute) - Số tokens mỗi phút
- **TPD** (Tokens Per Day) - Số tokens mỗi ngày

Sử dụng sliding window algorithm để tracking chính xác và tự động cleanup expired entries.

## Tại Cần Per-key Rate Tracking?

### Vấn đề
Khi sử dụng multiple API keys cho cùng một provider:
- Key A: 100 RPM limit
- Key B: 100 RPM limit
- Key C: 100 RPM limit

Nếu không có per-key tracking:
- Tổng requests = 300 RPM (tất cả keys)
- Mỗi key có thể bị rate limit riêng biệt
- Không biết key nào đang接近 limit

### Giải pháp
Per-key rate tracking:
- Track riêng biệt cho từng key
- Sliding window (không fixed reset)
- Cooldown khi gặp 429 errors
- Real-time status reporting

## Kiến Trúc

### File triển khai
- `Z:\Ti\router\layers\authentication\rate_tracker.go` - Core rate tracking logic
- `Z:\Ti\router\layers\authentication\rate_tracker_test.go` - Unit tests

### Components

#### 1. RateLimit
Định nghĩa rate limits:

```go
type RateLimit struct {
    RPM *int // Requests per minute (null = unlimited)
    RPD *int // Requests per day (null = unlimited)
    TPM *int // Tokens per minute (null = unlimited)
    TPD *int // Tokens per day (null = unlimited)
}
```

**Lưu ý**: Sử dụng pointer để hỗ trợ `null` (unlimited).

#### 2. Window
Sliding window cho tracking:

```go
type Window struct {
    requestTimestamps []time.Time
    tokenTimestamps   []tokenTimestamp
    mu                sync.RWMutex
}

type tokenTimestamp struct {
    ts     time.Time
    tokens int
}
```

**Thread-safe**: Mỗi window có mutex riêng để concurrent access.

#### 3. RateTracker
Manager cho rate tracking:

```go
type RateTracker struct {
    windows   map[string]*Window
    cooldowns map[string]time.Time
    mu        sync.RWMutex
}
```

**Key format**: `platform:modelId:keyId:type`

**Constants**:
- `minuteWindow = 1 minute`
- `dayWindow = 24 hours`

#### 4. CanMakeRequest
Kiểm tra nếu request có thể được thực hiện:

```go
func (rt *RateTracker) CanMakeRequest(platform, modelID, keyID string, limits RateLimit) bool
```

**Logic**:
1. Prune timestamps outside window
2. Count current requests
3. Check if < limit
4. Return true/false

**Checks**:
- RPM (nếu != nil)
- RPD (nếu != nil)

#### 5. CanUseTokens
Kiểm tra nếu tokens có thể được sử dụng:

```go
func (rt *RateTracker) CanUseTokens(platform, modelID, keyID string, estimatedTokens int, limits RateLimit) bool
```

**Logic**:
1. Prune token timestamps outside window
2. Sum used tokens
3. Check if used + estimated < limit
4. Return true/false

**Checks**:
- TPM (nếu != nil)
- TPD (nếu != nil)

#### 6. RecordRequest
Ghi nhận request:

```go
func (rt *RateTracker) RecordRequest(platform, modelID, keyID string)
```

**Logic**:
1. Add timestamp to RPM window
2. Add timestamp to RPD window
3. Auto-prune old timestamps

#### 7. RecordTokens
Ghi nhận token usage:

```go
func (rt *RateTracker) RecordTokens(platform, modelID, keyID string, tokens int)
```

**Logic**:
1. Add token timestamp to TPM window
2. Add token timestamp to TPD window
3. Auto-prune old timestamps

#### 8. SetCooldown
Set cooldown khi gặp 429:

```go
func (rt *RateTracker) SetCooldown(platform, modelID, keyID string, duration time.Duration)
```

**Logic**:
1. Tính expiry timestamp
2. Lưu vào cooldowns map
3. Auto-expire khi check

#### 9. IsOnCooldown
Kiểm tra nếu đang cooldown:

```go
func (rt *RateTracker) IsOnCooldown(platform, modelID, keyID string) bool
```

**Logic**:
1. Lookup cooldown expiry
2. Auto-delete nếu expired
3. Return true/false

#### 10. GetRateLimitStatus
Lấy current usage:

```go
func (rt *RateTracker) GetRateLimitStatus(platform, modelID, keyID string, limits RateLimit) map[string]interface{}
```

**Return**:
```json
{
  "rpm": {"used": 5, "limit": 10},
  "rpd": {"used": 50, "limit": 100},
  "tpm": {"used": 500, "limit": 1000},
  "tpd": {"used": 5000, "limit": 10000}
}
```

#### 11. Clear
Xóa tất cả data (cho testing):

```go
func (rt *RateTracker) Clear()
```

#### 12. GetStats
Thống kê rate tracker:

```go
func (rt *RateTracker) GetStats() map[string]interface{}
```

**Return**:
```json
{
  "total_windows": 8,
  "total_cooldowns": 2,
  "active_cooldowns": 1
}
```

## So Sánh với FreeLLMAPI

| Feature | FreeLLMAPI (TS) | Ti Router (Go) |
|---------|-----------------|----------------|
| Sliding window | Yes | Yes ✅ |
| RPM/RPD/TPM/TPD | Yes | Yes ✅ |
| Cooldown | Yes | Yes ✅ |
| Thread-safe | No (JS single-threaded) | Yes (sync.RWMutex) ✅ |
| Key format | platform:modelId:keyId:type | Same ✅ |
| Auto-prune | Yes | Yes ✅ |
| Unlimited support | null | *int (nil) ✅ |

**Khác biệt chính**:
- Ti Router thread-safe với mutex (Go concurrent)
- Cả hai đều cùng logic sliding window

## Usage Example

```go
// Create tracker (thường làm ở startup)
rateTracker := authentication.NewRateTracker()

// Define limits
limits := authentication.RateLimit{
    RPM: intPtr(10),
    RPD: intPtr(100),
    TPM: intPtr(1000),
    TPD: intPtr(10000),
}

platform := "openai"
modelID := "gpt-4"
keyID := "sk-abc123"

// Check if request can be made
if !rateTracker.CanMakeRequest(platform, modelID, keyID, limits) {
    return errors.New("Rate limited")
}

// Check if tokens can be used
estimatedTokens := 500
if !rateTracker.CanUseTokens(platform, modelID, keyID, estimatedTokens, limits) {
    return errors.New("Token rate limited")
}

// Check cooldown
if rateTracker.IsOnCooldown(platform, modelID, keyID) {
    return errors.New("On cooldown")
}

// Make request...
// Record usage
rateTracker.RecordRequest(platform, modelID, keyID)
rateTracker.RecordTokens(platform, modelID, keyID, actualTokens)

// Get status
status := rateTracker.GetRateLimitStatus(platform, modelID, keyID, limits)
// status["rpm"] = {"used": 1, "limit": 10}
```

## Integration với Router

### Cần update để sử dụng rate tracking:

1. **Router handler** (`layers/http/openai/handlers.go`):
   - Tạo global RateTracker
   - Check limits trước khi routing
   - Record usage sau khi request thành công
   - Set cooldown khi gặp 429

2. **Provider selection** (`layers/internal/router/router.go`):
   - Chấp nhận rate limits parameter
   - Skip keys on cooldown

3. **Admin API** (`cmd/routerd/handlers_admin.go`):
   - Thêm endpoint `/admin/rate-limits/status`
   - Thêm endpoint `/admin/rate-limits/clear`

## Testing

Tất cả tests pass:

```bash
cd Z:\Ti\router
go test ./layers/authentication/ -v -run "TestCan|TestSliding|TestCooldown|TestGetRate|TestUnlimited|TestRateTracker"
```

**Test coverage**:
- `TestCanMakeRequest` - RPM limit check
- `TestCanMakeRequestRPD` - RPD limit check
- `TestCanUseTokens` - TPM limit check
- `TestCanUseTokensTPD` - TPD limit check
- `TestSlidingWindowRPM` - Sliding window expiration (60s)
- `TestSlidingWindowTPM` - Sliding window expiration (60s)
- `TestCooldown` - Cooldown functionality
- `TestGetRateLimitStatus` - Status reporting
- `TestUnlimited` - nil limits (unlimited)
- `TestRateTrackerClear` - Manual clear
- `TestRateTrackerGetStats` - Statistics

## Performance Considerations

### ✅ Optimizations
- Sliding window (O(1) prune on each check)
- Per-window mutex (concurrent reads)
- Lazy cleanup (chỉ khi cần)
- In-memory map (fast lookup)

### ⚠️ Lưu ý
- Memory growth với nhiều keys
- Prune operation trên mỗi check
- Timestamp comparison overhead

### 🔐 Production Recommendations
1. Monitor memory usage
2. Implement periodic cleanup goroutine
3. Consider persistent storage cho cross-instance
4. Log rate limit hits cho analytics

## Security Considerations

### ✅ Safe
- Không lưu sensitive data
- Auto-expire timestamps
- Thread-safe operations

### ⚠️ Lưu ý
- In-memory only (restart = lost data)
- Không persist cooldowns

## Future Enhancements

1. **Persistent Storage**: Lưu rate limits trong DB
2. **Distributed Tracking**: Redis cho multi-instance
3. **Dynamic Limits**: Per-user hoặc per-model limits
4. **Burst Allowance**: Short burst capability
5. **Priority Queuing**: Queue requests khi rate limited
6. **Predictive Limits**: ML-based limit prediction

## References

- FreeLLMAPI: `Z:\Ti\Ti-learning-lab\05_Repositories\router\freellmapi-main\server\src\services\ratelimit.ts`
- Go sync package: https://pkg.go.dev/sync
- Rate limiting algorithms: https://en.wikipedia.org/wiki/Rate_limiting

---

**Ngày tạo**: 2026-04-28
**Agent**: Claude Code
**Project**: Ti Router
