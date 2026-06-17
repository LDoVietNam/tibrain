# FreeLLMAPI Features Integration Guide

> **Version**: 1.0.0  
> **Date**: 2026-04-28  
> **Status**: Core implementation complete, integration pending

## Tổng Quan

Đã hoàn thành core implementation cho 5 features từ FreeLLMAPI:
1. ✅ Encrypted Key Storage (AES-256-GCM)
2. ✅ Sticky Sessions (30 phút)
3. ✅ Per-key Rate Tracking (RPM/RPD/TPM/TPD)
4. ✅ Health Check Probes
5. ✅ Enhanced Analytics (Latency + Success Rate)

Document này hướng dẫn cách integrate các features này vào Ti Router.

## Files Đã Tạo

### Core Implementation
- `Z:\Ti\router\layers\authentication\encryption.go` - AES-256-GCM encryption
- `Z:\Ti\router\layers\authentication\sticky_session.go` - Sticky session manager
- `Z:\Ti\router\layers\authentication\rate_tracker.go` - Per-key rate tracking
- `Z:\Ti\router\layers\authentication\health_probes.go` - Health check probes
- `Z:\Ti\router\layers\authentication\analytics.go` - Enhanced analytics

### Database
- `Z:\Ti\router\layers\db\migrations.go` - Migration 4: Added latency_ms, error fields to usage_logs
- `Z:\Ti\router\layers\db\database.go` - Added RecordUsageExtended method

### Documentation
- `Z:\Ti\Ti-learning-lab\03_Knowledge\freellmapi\encrypted-key-storage.md`
- `Z:\Ti\Ti-learning-lab\03_Knowledge\freellmapi\sticky-sessions.md`
- `Z:\Ti\Ti-learning-lab\03_Knowledge\freellmapi\per-key-rate-tracking.md`
- `Z:\Ti\Ti-learning-lab\03_Knowledge\freellmapi\health-probes.md`
- `Z:\Ti\Ti-learning-lab\03_Knowledge\freellmapi\enhanced-analytics.md`

## Integration Steps

### 1. DB Migration

Chạy migration để thêm fields mới vào usage_logs table:

```bash
cd Z:\Ti\router
# Migration sẽ tự động chạy khi khởi tạo DB
# Hoặc chạy thủ công nếu cần
```

**Migration 4** đã được thêm:
- `ALTER TABLE usage_logs ADD COLUMN latency_ms INTEGER DEFAULT 0`
- `ALTER TABLE usage_logs ADD COLUMN error TEXT DEFAULT ''`
- Indexes cho latency_ms và status

### 2. Health Check Integration

**File cần update**: `cmd/routerd/main.go`

```go
import (
    "github.com/ti/router/layers/authentication"
    "github.com/ti/router/layers/db"
)

// Trong main function sau khi khởi tạo DB:
dbConn := db.Get()
healthProbe := authentication.NewHealthProbe(dbConn)

// Start health checker
healthProbe.Start()

// Graceful shutdown
defer healthProbe.Stop()
```

### 3. Sticky Sessions Integration

**File cần update**: `layers/internal/router/router.go`

```go
import (
    "github.com/ti/router/layers/authentication"
    "github.com/ti/router/layers/providers"
)

// Tạo global sticky session manager
var stickyManager = authentication.NewStickySessionManager()

// Trong routing logic:
func (r *Router) RouteRequest(req *providers.ChatRequest) (*providers.ChatResponse, error) {
    messages := req.Messages
    
    // Get sticky model nếu có
    preferredModel := stickyManager.GetStickyModel(messages)
    if preferredModel != "" {
        // Sử dụng sticky model
        // ...
    }
    
    // Sau khi request thành công
    stickyManager.SetStickyModel(messages, usedModelID)
    
    // ...
}
```

### 4. Rate Tracking Integration

**File cần update**: `layers/internal/router/router.go`

```go
import (
    "github.com/ti/router/layers/authentication"
)

// Tạo global rate tracker
var rateTracker = authentication.NewRateTracker()

// Define rate limits
var rateLimits = authentication.RateLimit{
    RPM: intPtr(100),  // 100 requests per minute
    RPD: intPtr(1000), // 1000 requests per day
    TPM: intPtr(50000), // 50k tokens per minute
    TPD: intPtr(500000), // 500k tokens per day
}

// Trong routing logic:
func (r *Router) RouteRequest(req *providers.ChatRequest) (*providers.ChatResponse, error) {
    platform := "openai"
    modelID := "gpt-4"
    keyID := "sk-abc123"
    
    // Check rate limits
    if !rateTracker.CanMakeRequest(platform, modelID, keyID, rateLimits) {
        return nil, fmt.Errorf("rate limited")
    }
    
    estimatedTokens := estimateTokens(req.Messages)
    if !rateTracker.CanUseTokens(platform, modelID, keyID, estimatedTokens, rateLimits) {
        return nil, fmt.Errorf("token rate limited")
    }
    
    // Check cooldown
    if rateTracker.IsOnCooldown(platform, modelID, keyID) {
        return nil, fmt.Errorf("on cooldown")
    }
    
    // Make request...
    resp, err := provider.Chat(ctx, req)
    if err != nil {
        // Set cooldown on error
        rateTracker.SetCooldown(platform, modelID, keyID, time.Minute*2)
        return nil, err
    }
    
    // Record usage
    rateTracker.RecordRequest(platform, modelID, keyID)
    rateTracker.RecordTokens(platform, modelID, keyID, resp.Usage.TotalTokens)
    
    return resp, nil
}
```

### 5. Encryption Integration

**File cần update**: `layers/authentication/credential_store.go` hoặc `layers/authentication/keys.go`

```go
import (
    "github.com/ti/router/layers/authentication"
)

// Khởi tạo encryption key (sau khi khởi tạo DB)
err := authentication.InitEncryptionKey(db.Get())
if err != nil {
    log.Fatal(err)
}

// Khi lưu API key:
crypto := authentication.NewCryptoManager()

// Encrypt trước khi lưu
encrypted, err := crypto.Encrypt(apiKey)
if err != nil {
    return err
}

// Lưu encrypted data thay vì plaintext
// key.EncryptedKey = encrypted.Encrypted
// key.IV = encrypted.IV
```

### 6. Analytics Integration

**File cần update**: `layers/http/openai/handlers.go` hoặc router handlers

```go
import (
    "github.com/ti/router/layers/authentication"
    "github.com/ti/router/layers/db"
)

// Tạo analytics instance
analytics := authentication.NewAnalytics(db.Get())

// Trong handler:
func handleChatCompletion(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // ... routing logic ...
    
    // Record với analytics
    latencyMs := int(time.Since(start).Milliseconds())
    status := "success"
    errorMsg := ""
    if err != nil {
        status = "error"
        errorMsg = err.Error()
    }
    
    err = analytics.RecordRequest(
        context.Background(),
        provider,
        modelID,
        connectionID,
        tokensIn,
        tokensOut,
        latencyMs,
        status,
        errorMsg,
    )
    if err != nil {
        log.Printf("Failed to record analytics: %v", err)
    }
    
    // ...
}
```

### 7. Admin API Endpoints

**File cần update**: `cmd/routerd/handlers_admin.go`

```go
import (
    "github.com/ti/router/layers/authentication"
    "github.com/ti/router/layers/db"
)

// Health check endpoints
func handleHealthStatus(w http.ResponseWriter, r *http.Request) {
    healthProbe := getHealthProbe() // Global instance
    status, err := healthProbe.GetHealthStatus(context.Background())
    // Return JSON response
}

func handleHealthCheckKey(w http.ResponseWriter, r *http.Request) {
    // Parse keyID from URL
    keyID := r.URL.Query().Get("key_id")
    
    healthProbe := getHealthProbe()
    status, err := healthProbe.CheckKeyHealth(context.Background(), keyID)
    // Return JSON response
}

// Analytics endpoints
func handleAnalyticsSummary(w http.ResponseWriter, r *http.Request) {
    analytics := getAnalytics() // Global instance
    range := r.URL.Query().Get("range") // 24h, 7d, 30d
    
    var tr authentication.TimeRange
    switch range {
    case "24h":
        tr = authentication.Range24h
    case "30d":
        tr = authentication.Range30d
    default:
        tr = authentication.Range7d
    }
    
    stats, err := analytics.GetSummaryStats(context.Background(), tr)
    // Return JSON response
}

// Rate tracking endpoints
func handleRateLimitStatus(w http.ResponseWriter, r *http.Request) {
    rateTracker := getRateTracker() // Global instance
    
    platform := r.URL.Query().Get("platform")
    modelID := r.URL.Query().Get("model")
    keyID := r.URL.Query().Get("key_id")
    
    limits := authentication.RateLimit{
        RPM: intPtr(100),
        RPD: intPtr(1000),
        TPM: intPtr(50000),
        TPD: intPtr(500000),
    }
    
    status := rateTracker.GetRateLimitStatus(platform, modelID, keyID, limits)
    // Return JSON response
}
```

## Testing

### Test Encryption

```bash
cd Z:\Ti\router
go test ./layers/authentication/ -v -run TestEncrypt
```

### Test Sticky Sessions

```bash
go test ./layers/authentication/ -v -run "TestGetSession|TestGetSticky"
```

### Test Rate Tracking

```bash
go test ./layers/authentication/ -v -run "TestCanMake|TestCanUse|TestSliding"
```

### Test Health Probes

```bash
go test ./authentication -v -run "TestStartStop|TestGetHealth"
```

### Test Analytics

```bash
go test ./layers/authentication/ -v -run "TestGetTime|TestCategorize"
```

### Test All

```bash
go test ./layers/authentication/ -v
```

## Build Verification

```bash
cd Z:\Ti\router
go build ./...
```

## Notes

### DB Migration
- Migration 4 đã được thêm nhưng chưa chạy
- Cần chạy migration khi khởi tạo DB hoặc chạy thủ công
- Migration là reversible (có Down script)

### Provider Validation
- Health check hiện tại sử dụng connection status
- TODO: Implement actual API call để validate key
- Có thể sử dụng Provider.HealthCheck() method

### Analytics Queries
- Analytics methods hiện tại là placeholder
- Cần thêm Query method vào Database interface
- Hoặc implement raw SQL queries

### Global Instances
- Các managers (sticky, rate, health, analytics) nên là global singletons
- Khởi tạo ở startup trong main.go
- Cleanup gracefully ở shutdown

## Next Steps

1. **Run DB migration**: Chạy migration 4 để thêm fields
2. **Add Query method**: Thêm Query method vào Database interface cho analytics
3. **Implement actual provider validation**: Sử dụng Provider.HealthCheck()
4. **Add to router handlers**: Integrate vào routing logic
5. **Add admin endpoints**: Tạo admin API endpoints
6. **Test end-to-end**: Test full integration với actual requests

## References

- Core implementation docs trong `Z:\Ti\Ti-learning-lab\03_Knowledge\freellmapi\`
- FreeLLMAPI reference: `Z:\Ti\Ti-learning-lab\05_Repositories\router\freellmapi-main\`

---

**Ngày tạo**: 2026-04-28
**Agent**: Claude Code
**Project**: Ti Router
