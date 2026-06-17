---
tags: ["tibrain", "testing", "documentation", "skill", "provider"]
scopes: ["code", "tibrain"]
last_updated: 2026-05-22
---
# A/B Testing Pattern

> **Category**: Testing & Experimentation  
> **Language**: Tiếng Việt  
> **Last Updated**: 2026-04-30

---

## Tổng Quan

A/B testing là phương pháp thử nghiệm so sánh hai phiên bản (A và B) của một biến để xác định phiên bản nào hoạt động tốt hơn. Trong routing AI, A/B testing được sử dụng để so sánh các chiến lược routing khác nhau.

---

## Khi Nào Sử Dụng

- So sánh hiệu suất giữa các routing strategies khác nhau
- Test new providers trước khi rollout toàn bộ
- So sánh các model versions
- Tối ưu hóa latency, cost, hoặc success rate

---

## Cấu Trúc

### Variants (Biến Thể)

```go
type Variant struct {
    ID     string
    Name   string
    Weight float64 // 0.0 đến 1.0
    Config map[string]interface{}
}
```

**Best Practices**:
- Sử dụng weights cân bằng (0.5/0.5) cho fair comparison
- Sử dụng weights nhỏ cho test mới (e.g., 0.1/0.9)
- Tổng weights phải = 1.0

### Traffic Allocation Strategies

**1. Percentage (Phần trăm)**
- Route dựa trên phần trăm ngẫu nhiên
- Đơn giản, dễ implement
- Không đảm bảo consistency cho cùng user

**2. User Hash (Hash user)**
- Route dựa trên hash của user ID
- Đảm bảo cùng user luôn được route đến cùng variant
- Consistent cho repeated requests
- Khuyến nghị cho production

**3. Random (Ngẫu nhiên)**
- Route ngẫu nhiên thuần túy
- Không phù hợp cho long-running tests

---

## Metrics Tracking

### Metrics Quan Trọng

```go
type VariantMetrics struct {
    Requests       int64
    Successes      int64
    Errors         int64
    AvgLatencyMs   float64
    AvgCost        float64
    UserSatisfaction float64
}
```

### Winner Selection

**Criteria**:
- Success rate cao nhất
- Latency thấp nhất
- Cost thấp nhất
- User satisfaction cao nhất

**Implementation**:
```go
winner = variant với highest success rate
```

---

## Best Practices

### 1. Test Duration
- Minimum: 24 hours
- Recommended: 1-7 days
- Depends on traffic volume

### 2. Sample Size
- Minimum: 1000 requests per variant
- Recommended: 10,000+ requests per variant
- Statistical significance (p-value < 0.05)

### 3. Rollout Strategy
- Start với small percentage (5-10%)
- Monitor metrics closely
- Gradually increase nếu metrics good
- Stop immediately nếu metrics bad

### 4. Data Collection
- Track tất cả metrics (latency, cost, errors)
- Log context (user ID, request ID)
- Export data cho analysis

---

## Implementation Pattern

### 1. Create Test

```go
variants := []Variant{
    {ID: "variant_a", Name: "Strategy A", Weight: 0.5},
    {ID: "variant_b", Name: "Strategy B", Weight: 0.5},
}

test := manager.CreateTest(
    "Test Routing Strategies",
    "Compare RoundRobin vs LeastLatency",
    variants,
    TrafficAllocation{Strategy: "user_hash"},
)
```

### 2. Start Test

```go
manager.StartTest(test.ID)
```

### 3. For Each Request

```go
variant := manager.GetVariant(test.ID, userID)
// Route to variant
manager.RecordMetric(test.ID, variant.ID, success, latency, cost)
```

### 4. Stop & Analyze

```go
manager.StopTest(test.ID)
winner := manager.GetWinner(test.ID)
// Promote winner
```

---

## Common Mistakes

### 1. Insufficient Sample Size
- **Problem**: Test với quá ít requests
- **Solution**: Minimum 1000 requests per variant

### 2. Short Test Duration
- **Problem**: Test chỉ chạy vài giờ
- **Solution**: Minimum 24 hours cho representative data

### 3. Ignoring Statistical Significance
- **Problem**: Kết luận từ small sample
- **Solution**: Calculate p-value, ensure statistical significance

### 4. Not Monitoring Metrics
- **Problem**: Chạy test mà không monitor
- **Solution**: Set up alerts cho critical metrics

### 5. Premature Rollout
- **Problem**: Rollout winner quá nhanh
- **Solution**: Gradual rollout (5% → 10% → 25% → 50% → 100%)

---

## Advanced Patterns

### 1. Multi-Variant Testing
- Test nhiều hơn 2 variants (A, B, C, ...)
- Useful cho comparing multiple strategies

### 2. Sequential Testing
- Test variants sequentially thay vì parallel
- Useful cho limited traffic

### 3. Contextual Testing
- Test cho specific user segments
- E.g., test cho high-value users

### 4. Bayesian A/B Testing
- Sử dụng Bayesian inference thay vì frequentist
- Faster convergence với smaller samples

---

## Tools & Libraries

### Open Source
- **VWO**: A/B testing platform
- **Optimizely**: Experimentation platform
- **PostHog**: Product analytics với A/B testing

### Go Libraries
- **github.com/ab-testing-in-go-sample**: Sample implementation
- **github.com/feature-flags**: Feature flag với A/B testing

---

## References

- [A/B Testing Best Practices](https://optimizely.com/optimization-glossary/ab-testing/)
- [Statistical Significance Calculator](https://www.optimizely.com/sample-size-calculator/)
- [A/B Testing in Go](https://github.com/configcat-labs/ab-testing-in-go-sample)

---

## Ti Router Implementation

**File**: `layers/routing/ab_testing.go`

**Features**:
- ABTestManager với in-memory storage
- Support cho 3 traffic strategies (percentage, user_hash, random)
- Metrics tracking per variant
- Winner selection dựa trên success rate
- Thread-safe với mutex

**Usage**:
```go
manager := NewABTestManager()
test := manager.CreateTest(...)
manager.StartTest(test.ID)
variant := manager.GetVariant(test.ID, userID)
manager.RecordMetric(test.ID, variant.ID, success, latency, cost)
winner := manager.GetWinner(test.ID)
```

---

## Next Steps

1. Implement Bayesian A/B testing cho faster convergence
2. Add statistical significance calculation
3. Export metrics cho external analysis tools
4. Add alerting cho test anomalies
5. Implement multi-variant testing
