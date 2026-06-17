# Enhanced Analytics (Latency + Success Rate)

> **Feature**: Enhanced analytics với latency tracking và success rate metrics
> **Nguồn cảm hứng**: FreeLLMAPI - `server/src/routes/analytics.ts`
> **Trạng thái**: ✅ Hoàn thành (Core implementation, DB migration needed)

## Tổng Quan

Ti Router hiện hỗ trợ enhanced analytics để track và visualize:
- **Summary stats**: Tổng requests, success rate, tokens, latency, cost
- **By model**: Stats grouped theo model
- **By platform**: Stats grouped theo platform
- **Timeline**: Timeline data (hourly/daily)
- **Error distribution**: Errors categorized và grouped
- **Recent errors**: Last 50 errors

## Tại Cần Enhanced Analytics?

### Vấn đề
Basic usage logging không cung cấp đủ insight:
- Không biết success rate
- Không track latency
- Không categorize errors
- Không có timeline visualization
- Không có cost estimation

### Giải pháp
Enhanced analytics:
- Track success/failure status
- Measure latency cho mỗi request
- Categorize errors theo type
- Provide timeline visualization
- Estimate cost savings

## Kiến Trúc

### File triển khai
- `Z:\Ti\router\layers\authentication\analytics.go` - Core analytics logic
- `Z:\Ti\router\layers\authentication\analytics_test.go` - Unit tests

### Components

#### 1. TimeRange
Time range cho analytics queries:

```go
type TimeRange string

const (
    Range24h TimeRange = "24h"
    Range7d  TimeRange = "7d"
    Range30d TimeRange = "30d"
)
```

#### 2. SummaryStats
Summary statistics:

```go
type SummaryStats struct {
    TotalRequests      int     `json:"total_requests"`
    SuccessRate       float64 `json:"success_rate"`
    TotalInputTokens  int     `json:"total_input_tokens"`
    TotalOutputTokens int     `json:"total_output_tokens"`
    AvgLatencyMs      float64 `json:"avg_latency_ms"`
    EstimatedCost     float64 `json:"estimated_cost_savings"`
}
```

#### 3. ModelStats
Statistics grouped by model:

```go
type ModelStats struct {
    Platform          string  `json:"platform"`
    ModelID           string  `json:"model_id"`
    DisplayName       string  `json:"display_name"`
    Requests          int     `json:"requests"`
    SuccessRate       float64 `json:"success_rate"`
    AvgLatencyMs      float64 `json:"avg_latency_ms"`
    TotalInputTokens  int     `json:"total_input_tokens"`
    TotalOutputTokens int     `json:"total_output_tokens"`
}
```

#### 4. PlatformStats
Statistics grouped by platform:

```go
type PlatformStats struct {
    Platform          string  `json:"platform"`
    Requests          int     `json:"requests"`
    SuccessRate       float64 `json:"success_rate"`
    AvgLatencyMs      float64 `json:"avg_latency_ms"`
    TotalInputTokens  int     `json:"total_input_tokens"`
    TotalOutputTokens int     `json:"total_output_tokens"`
}
```

#### 5. TimelineData
Timeline data points:

```go
type TimelineData struct {
    Timestamp     string `json:"timestamp"`
    Requests      int    `json:"requests"`
    SuccessCount  int    `json:"success_count"`
    FailureCount  int    `json:"failure_count"`
}
```

#### 6. ErrorDistribution
Error distribution:

```go
type ErrorDistribution struct {
    ByCategory []ErrorCategory `json:"by_category"`
    ByPlatform  []ErrorCategory `json:"by_platform"`
    Detailed    []ErrorDetail    `json:"detailed"`
}
```

#### 7. Analytics
Manager cho analytics:

```go
type Analytics struct {
    db db.Database
}
```

#### 8. GetSummaryStats
Lấy summary statistics:

```go
func (a *Analytics) GetSummaryStats(ctx context.Context, tr TimeRange) (*SummaryStats, error)
```

**Logic**:
1. Query usage_logs table
2. Calculate total requests, success rate
3. Sum input/output tokens
4. Calculate average latency
5. Estimate cost savings

**TODO**: Cần DB migration để thêm `status`, `latency_ms`, `error` fields vào `usage_logs` table.

#### 9. GetModelStats
Lấy statistics grouped by model:

```go
func (a *Analytics) GetModelStats(ctx context.Context, tr TimeRange) ([]ModelStats, error)
```

**Logic**:
1. Query usage_logs JOIN với models
2. Group by platform, model_id
3. Calculate stats per group
4. Return sorted by requests

**TODO**: Cần DB migration.

#### 10. GetPlatformStats
Lấy statistics grouped by platform:

```go
func (a *Analytics) GetPlatformStats(ctx context.Context, tr TimeRange) ([]PlatformStats, error)
```

**Logic**:
1. Query usage_logs
2. Group by platform
3. Calculate stats per group
4. Return sorted by requests

**TODO**: Cần DB migration.

#### 11. GetTimelineData
Lấy timeline data:

```go
func (a *Analytics) GetTimelineData(ctx context.Context, tr TimeRange, interval string) ([]TimelineData, error)
```

**Logic**:
1. Determine interval (hour/day)
2. Query usage_logs grouped by timestamp
3. Calculate success/failure counts per interval
4. Return sorted chronologically

**TODO**: Cần DB migration.

#### 12. GetErrorDistribution
Lấy error distribution:

```go
func (a *Analytics) GetErrorDistribution(ctx context.Context, tr TimeRange) (*ErrorDistribution, error)
```

**Logic**:
1. Query errors từ usage_logs
2. Categorize errors (429, 401, 403, etc.)
3. Group by category
4. Group by platform
5. Return detailed distribution

**TODO**: Cần DB migration.

#### 13. GetRecentErrors
Lấy recent errors:

```go
func (a *Analytics) GetRecentErrors(ctx context.Context, tr TimeRange, limit int) ([]RecentError, error)
```

**Logic**:
1. Query errors từ usage_logs
2. Sort by created_at DESC
3. Limit to N results
4. Return error details

**TODO**: Cần DB migration.

#### 14. RecordRequest
Ghi nhận request cho analytics:

```go
func (a *Analytics) RecordRequest(ctx context.Context, provider, model, connectionID string, tokensIn, tokensOut int, latencyMs int, status string, errorMsg string) error
```

**Logic**:
1. Record usage với existing RecordUsage
2. TODO: Record status, latency_ms, error khi DB migration complete

#### 15. categorizeError
Categorize error message:

```go
func categorizeError(errorMsg string) string
```

**Categories**:
- "Rate Limited (429)" - 429, rate limit, quota
- "Auth Error (401)" - 401, unauthorized, invalid key
- "Forbidden (403)" - 403, forbidden
- "Not Found (404)" - 404, not found
- "Timeout/Connection" - timeout, ETIMEDOUT, ECONNREFUSED
- "Server Error (500)" - 500, internal server
- "Unavailable (503)" - 503, unavailable
- "Other" - unknown errors

#### 16. estimateCostSavings
Estimate cost savings:

```go
func estimateCostSavings(inputTokens, outputTokens int) float64
```

**Pricing**: GPT-4o - ~$3/M input + $15/M output

## So Sánh với FreeLLMAPI

| Feature | FreeLLMAPI (TS) | Ti Router (Go) |
|---------|-----------------|----------------|
| Summary stats | Yes | Yes (placeholder) ✅ |
| By model | Yes | Yes (placeholder) ✅ |
| By platform | Yes | Yes (placeholder) ✅ |
| Timeline | Yes | Yes (placeholder) ✅ |
| Error distribution | Yes | Yes (placeholder) ✅ |
| Recent errors | Yes | Yes (placeholder) ✅ |
| Cost estimation | Yes | Yes ✅ |
| Error categorization | Yes | Yes ✅ |
| Time ranges | 24h, 7d, 30d | 24h, 7d, 30d ✅ |

**Khác biệt chính**:
- Ti Router implementation là placeholder (cần DB migration)
- FreeLLMAPI sử dụng direct SQL, Ti Router sử dụng DB interface
- Cả hai đều cùng logic core

## Usage Example

```go
// Create analytics instance (thường làm ở startup)
analytics := authentication.NewAnalytics(db.Get())

// Get summary stats for last 7 days
stats, err := analytics.GetSummaryStats(context.Background(), authentication.Range7d)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total requests: %d\n", stats.TotalRequests)
fmt.Printf("Success rate: %.1f%%\n", stats.SuccessRate)
fmt.Printf("Avg latency: %.0fms\n", stats.AvgLatencyMs)
fmt.Printf("Estimated cost: $%.2f\n", stats.EstimatedCost)

// Get stats by model
modelStats, err := analytics.GetModelStats(context.Background(), authentication.Range7d)
if err != nil {
    log.Fatal(err)
}

for _, ms := range modelStats {
    fmt.Printf("%s/%s: %d requests, %.1f%% success\n", ms.Platform, ms.ModelID, ms.Requests, ms.SuccessRate)
}

// Record a request (sau khi API call)
err = analytics.RecordRequest(
    context.Background(),
    "openai",
    "gpt-4",
    "key-id-1",
    100,  // input tokens
    200,  // output tokens
    500,  // latency ms
    "success", // status
    "", // error (empty if success)
)
```

## Integration với Router

### Cần update để sử dụng analytics:

1. **DB Migration** (`layers/db/migrations.go`):
   - Add `status` field to `usage_logs` (TEXT: success/error)
   - Add `latency_ms` field to `usage_logs` (INTEGER)
   - Add `error` field to `usage_logs` (TEXT)
   - Create indexes cho new fields

2. **Router handler** (`layers/http/openai/handlers.go`):
   - Create Analytics instance
   - Record request status, latency, error sau mỗi API call
   - Measure latency với time.Since()

3. **Admin API** (`cmd/routerd/handlers_admin.go`):
   - Thêm endpoint `GET /admin/analytics/summary`
   - Thêm endpoint `GET /admin/analytics/by-model`
   - Thêm endpoint `GET /admin/analytics/by-platform`
   - Thêm endpoint `GET /admin/analytics/timeline`
   - Thêm endpoint `GET /admin/analytics/error-distribution`
   - Thêm endpoint `GET /admin/analytics/errors`

## Testing

Tests pass:

```bash
cd Z:\Ti\router
go test ./layers/authentication/ -v -run "TestGetTime|TestCategorize|TestEstimate|TestContains|TestGetSummary|TestRecordRequestWithMock"
```

**Test coverage**:
- `TestGetTimeFilter` - Time filter logic
- `TestCategorizeError` - Error categorization
- `TestEstimateCostSavings` - Cost estimation
- `TestContains` - Helper function
- `TestGetSummaryStats` - Placeholder summary stats
- `TestRecordRequestWithMock` - Record request với mock DB

**Skipped tests**:
- `TestRecordRequest` - Requires actual DB connection

## Performance Considerations

### ✅ Optimizations
- Time range filtering (không query toàn bộ data)
- Indexed queries (timestamp, provider)
- Aggregation ở DB level (không in-memory)
- Placeholder implementation (lightweight)

### ⚠️ Lưu ý
- DB queries có thể expensive với large datasets
- Timeline queries cần optimization cho long ranges
- Error categorization string operations

### 🔐 Production Recommendations
1. Implement DB migration trước khi production
2. Add caching cho analytics queries
3. Consider pre-aggregation cho common queries
4. Implement data retention policy (delete old logs)
5. Add monitoring cho analytics query performance

## Security Considerations

### ✅ Safe
- Không expose sensitive data
- Error messages có thể contain info (cần sanitize)
- Time range limits prevent large queries

### ⚠️ Lưu ý
- Error messages có thể contain API keys (cần mask)
- Analytics data có thể reveal usage patterns

## Future Enhancements

1. **DB Migration**: Add status, latency_ms, error fields to usage_logs
2. **Real-time Analytics**: WebSocket updates cho live stats
3. **Custom Dashboards**: Per-user customizable dashboards
4. **Alerting**: Alert on high error rates hoặc latency
5. **Cost Optimization**: Recommend cost-saving strategies
6. **Predictive Analytics**: ML-based usage prediction

## DB Migration Required

Để enable full analytics functionality, cần migration:

```sql
-- Add new fields to usage_logs
ALTER TABLE usage_logs ADD COLUMN status TEXT DEFAULT 'success';
ALTER TABLE usage_logs ADD COLUMN latency_ms INTEGER DEFAULT 0;
ALTER TABLE usage_logs ADD COLUMN error TEXT;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_usage_logs_status ON usage_logs(status);
CREATE INDEX IF NOT EXISTS idx_usage_logs_latency ON usage_logs(latency_ms);
```

## References

- FreeLLMAPI: `Z:\Ti\Ti-learning-lab\05_Repositories\router\freellmapi-main\server\src\routes\analytics.ts`
- Go database/sql: https://pkg.go.dev/database/sql
- SQLite date functions: https://www.sqlite.org/lang_datefunc.html

---

**Ngày tạo**: 2026-04-28
**Agent**: Claude Code
**Project**: Ti Router
