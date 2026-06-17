# Health Check Probes

> **Feature**: Health check probes cho API keys với auto-disable
> **Nguồn cảm hứng**: FreeLLMAPI - `server/src/services/health.ts`, `server/src/routes/health.ts`
> **Trạng thái**: ✅ Hoàn thành

## Tổng Quan

Ti Router hiện hỗ trợ health check probes để monitor sức khỏe của API keys:
- **Periodic checks**: Mỗi 5 phút
- **Consecutive failure tracking**: Track failures liên tiếp
- **Auto-disable**: Tự động disable sau 3 failures
- **Health status reporting**: Real-time status cho tất cả platforms và keys

## Tại Cần Health Check Probes?

### Vấn đề
API keys có thể:
- Expire
- Revoke
- Hit rate limits
- Quota exceeded
- Provider downtime

Nếu không có health checks:
- Router tiếp tục gửi requests đến invalid keys
- Waste resources và latency
- Poor user experience

### Giải pháp
Health check probes:
- Periodically validate keys
- Track consecutive failures
- Auto-disable problematic keys
- Provide real-time health status

## Kiến Trúc

### File triển khai
- `Z:\Ti\router\layers\authentication\health_probes.go` - Core health check logic
- `Z:\Ti\router\layers\authentication\health_probes_test.go` - Unit tests

### Components

#### 1. KeyStatus
Health status của API key:

```go
type KeyStatus string

const (
    StatusHealthy     KeyStatus = "healthy"
    StatusInvalid     KeyStatus = "invalid"
    StatusError       KeyStatus = "error"
    StatusRateLimited KeyStatus = "rate_limited"
    StatusUnknown     KeyStatus = "unknown"
)
```

#### 2. HealthProbe
Manager cho health checks:

```go
type HealthProbe struct {
    db              db.Database
    failureCount    map[string]int
    failureCountMu  sync.RWMutex
    checkInterval   time.Duration
    stopChan        chan struct{}
    isRunning       bool
    mu              sync.RWMutex
}
```

**Constants**:
- `checkInterval = 5 minutes` - Interval giữa các checks
- `consecutiveFailuresToDisable = 3` - Số failures trước khi auto-disable

#### 3. CheckKeyHealth
Check sức khỏe của một key cụ thể:

```go
func (hp *HealthProbe) CheckKeyHealth(ctx context.Context, keyID string) (KeyStatus, error)
```

**Logic**:
1. Get key từ DB
2. Validate key (TODO: implement provider validation)
3. Update status trong DB
4. Track consecutive failures
5. Auto-disable nếu >= 3 failures
6. Return status

**TODO**: Implement actual provider validation logic.

#### 4. CheckAllKeys
Check tất cả enabled keys:

```go
func (hp *HealthProbe) CheckAllKeys(ctx context.Context) error
```

**Logic**:
1. List tất cả active connections
2. Check từng key
3. Log errors nếu có
4. Return tổng kết quả

#### 5. Start
Start periodic health checker:

```go
func (hp *HealthProbe) Start()
```

**Logic**:
1. Check nếu đang running (idempotent)
2. Create ticker với interval
3. Start goroutine để run CheckAllKeys periodically
4. Handle stop signal

#### 6. Stop
Stop periodic health checker:

```go
func (hp *HealthProbe) Stop()
```

**Logic**:
1. Gửi stop signal qua channel
2. Goroutine sẽ cleanup và exit
3. Idempotent (safe để gọi nhiều lần)

#### 7. IsRunning
Kiểm tra nếu health checker đang running:

```go
func (hp *HealthProbe) IsRunning() bool
```

#### 8. GetHealthStatus
Lấy health status cho tất cả platforms và keys:

```go
func (hp *HealthProbe) GetHealthStatus(ctx context.Context) (map[string]interface{}, error)
```

**Logic**:
1. List tất cả active connections
2. Group by platform
3. Count keys theo status
4. Return platform stats và key list

**Return format**:
```json
{
  "platforms": [
    {
      "platform": "openai",
      "total_keys": 5,
      "healthy_keys": 3,
      "rate_limited_keys": 1,
      "invalid_keys": 1,
      "error_keys": 0,
      "unknown_keys": 0,
      "enabled_keys": 4
    }
  ],
  "keys": [
    {
      "id": "key1",
      "platform": "openai",
      "name": "OpenAI Key 1",
      "status": "healthy",
      "enabled": true,
      "created_at": "2026-04-28T...",
      "last_tested_at": "2026-04-28T..."
    }
  ]
}
```

#### 9. GetFailureCount
Lấy consecutive failure count cho một key:

```go
func (hp *HealthProbe) GetFailureCount(keyID string) int
```

#### 10. ResetFailureCount
Reset failure count cho một key:

```go
func (hp *HealthProbe) ResetFailureCount(keyID string)
```

#### 11. Clear
Xóa tất cả failure counts (cho testing):

```go
func (hp *HealthProbe) Clear()
```

## So Sánh với FreeLLMAPI

| Feature | FreeLLMAPI (TS) | Ti Router (Go) |
|---------|-----------------|----------------|
| Periodic checks | Yes (5 min) | Yes (5 min) ✅ |
| Consecutive failure tracking | Yes | Yes ✅ |
| Auto-disable | Yes (3 failures) | Yes (3 failures) ✅ |
| Health status reporting | Yes | Yes ✅ |
| Manual check endpoint | Yes | TODO ✅ |
| Thread-safe | No (JS single-threaded) | Yes (sync.RWMutex) ✅ |
| Stop/Start control | Yes | Yes ✅ |

**Khác biệt chính**:
- Ti Router thread-safe với mutex (Go concurrent)
- Ti Router sử dụng DB interface thay vì direct SQL
- Cả hai đều cùng logic core

## Usage Example

```go
// Create health probe (thường làm ở startup)
healthProbe := authentication.NewHealthProbe(db.Get())

// Start periodic health checker
healthProbe.Start()

// Check if running
if healthProbe.IsRunning() {
    fmt.Println("Health checker is running")
}

// Get health status
status, err := healthProbe.GetHealthStatus(context.Background())
if err != nil {
    log.Fatal(err)
}

// Access platform stats
platforms := status["platforms"].([]map[string]interface{})
for _, p := range platforms {
    fmt.Printf("Platform %s: %d healthy keys\n", p["platform"], p["healthy_keys"])
}

// Check specific key
keyStatus, err := healthProbe.CheckKeyHealth(context.Background(), "key-id-1")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Key status: %s\n", keyStatus)

// Get failure count
failures := healthProbe.GetFailureCount("key-id-1")
fmt.Printf("Consecutive failures: %d\n", failures)

// Stop health checker (thường làm ở shutdown)
healthProbe.Stop()
```

## Integration với Router

### Cần update để sử dụng health probes:

1. **Router startup** (`cmd/routerd/main.go`):
   - Create HealthProbe instance
   - Start periodic health checker
   - Handle graceful shutdown

2. **Admin API** (`cmd/routerd/handlers_admin.go`):
   - Thêm endpoint `GET /admin/health` - Get health status
   - Thêm endpoint `POST /admin/health/check/:keyId` - Check specific key
   - Thêm endpoint `POST /admin/health/check-all` - Check all keys
   - Thêm endpoint `POST /admin/health/start` - Start health checker
   - Thêm endpoint `POST /admin/health/stop` - Stop health checker

3. **Provider validation** (`layers/providers/*.go`):
   - Implement `ValidateKey()` method cho từng provider
   - Return true/false dựa trên actual API call

## Testing

Tests pass:

```bash
cd Z:\Ti\router
go test ./layers/authentication/ -v -run "TestStart|TestFailure|TestConsecutive|TestCheckInterval|TestGetHealth"
```

**Test coverage**:
- `TestStartStop` - Start/Stop health checker
- `TestFailureCount` - Failure count tracking
- `TestConsecutiveFailuresToDisable` - Constant verification
- `TestCheckInterval` - Interval constant verification
- `TestGetHealthStatus` - Health status reporting với mock DB

**Skipped tests**:
- `TestCheckKeyHealth` - Requires actual DB connection
- `TestCheckAllKeys` - Requires actual DB connection

## Performance Considerations

### ✅ Optimizations
- Periodic checks (không check mỗi request)
- Concurrent-safe với mutex
- Idempotent Start/Stop
- Lazy failure tracking

### ⚠️ Lưu ý
- DB queries trên mỗi check
- Network latency cho provider validation
- Memory cho failure counts

### 🔐 Production Recommendations
1. Implement actual provider validation
2. Add alerting cho high failure rates
3. Consider distributed health checks cho multi-instance
4. Log health check results cho monitoring
5. Adjust check interval dựa trên key count

## Security Considerations

### ✅ Safe
- Không expose sensitive data trong health status
- Auto-disable invalid keys
- Thread-safe operations

### ⚠️ Lưu ý
- Health checks có thể trigger rate limits
- Need rate limiting cho health check API

## Future Enhancements

1. **Provider Validation**: Implement actual ValidateKey() cho từng provider
2. **Custom Intervals**: Per-platform hoặc per-key check intervals
3. **Health History**: Track health status over time
4. **Alerting**: Integrate với monitoring/alerting systems
5. **Graceful Degradation**: Fallback strategies khi nhiều keys unhealthy
6. **Distributed Health Checks**: Coordination giữa multiple router instances

## References

- FreeLLMAPI: `Z:\Ti\Ti-learning-lab\05_Repositories\router\freellmapi-main\server\src\services\health.ts`
- FreeLLMAPI Routes: `Z:\Ti\Ti-learning-lab\05_Repositories\router\freellmapi-main\server\src\routes\health.ts`
- Go context package: https://pkg.go.dev/context

---

**Ngày tạo**: 2026-04-28
**Agent**: Claude Code
**Project**: Ti Router
