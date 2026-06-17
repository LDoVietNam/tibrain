---
tags: ["tibrain", "deployment", "documentation", "skill", "monitoring"]
scopes: ["observability", "infrastructure", "tibrain"]
last_updated: 2026-05-22
---
# Canary Deployment Pattern

> **Category**: Deployment & Release  
> **Language**: Tiếng Việt  
> **Last Updated**: 2026-04-30

---

## Tổng Quan

Canary deployment là chiến lược rollout dần dần một phiên bản mới cho một phần nhỏ users, monitor metrics, và nếu tốt thì tăng dần traffic. Nếu có vấn đề, rollback ngay lập tức.

---

## Khi Nào Sử Dụng

- Rollout new features với risk cao
- Test new infrastructure hoặc providers
- Gradual migration từ old system sang new system
- Reduce risk của deployment failures

---

## Cấu Trúc

### Versions (Phiên Bản)

```go
type CanaryVersion struct {
    ID       string
    Name     string
    Weight   float64 // 0.0 đến 1.0
    IsStable bool    // true = stable version
}
```

**Best Practices**:
- Always có một stable version (IsStable = true)
- Canary version bắt đầu với weight nhỏ (5-10%)
- Gradually increase weight nếu metrics good

### Deployment Strategy

```go
type CanaryStrategy struct {
    Type           string  // "gradual", "instant", "time_based"
    Increment      float64 // Tăng weight mỗi bước (e.g., 0.1 = 10%)
    Interval       time.Duration // Thời gian giữa các bước
    SuccessThreshold float64 // Threshold cho success rate (e.g., 0.95 = 95%)
    ErrorThreshold   float64 // Threshold cho error rate (e.g., 0.05 = 5%)
    AutoPromote     bool   // Tự động tăng weight nếu metrics tốt
    AutoRollback    bool   // Tự động rollback nếu metrics bad
}
```

---

## Deployment Flow

### 1. Initial Setup

```
Stable Version: 90% weight
Canary Version: 10% weight
```

### 2. Monitor Metrics

**Metrics quan trọng**:
- Error rate (phải < error_threshold)
- Success rate (phải > success_threshold)
- Latency (không tăng đáng kể)
- User satisfaction (không giảm)

### 3. Gradual Promotion (nếu metrics good)

```
Step 1: Stable 90%, Canary 10%
Step 2: Stable 80%, Canary 20%
Step 3: Stable 70%, Canary 30%
...
Step 9: Stable 10%, Canary 90%
Step 10: Stable 0%, Canary 100% (complete)
```

### 4. Auto-Rollback (nếu metrics bad)

```
Error rate > threshold → Rollback ngay lập tức
→ Reset canary weight về 0%
→ Stable weight về 100%
```

---

## Best Practices

### 1. Start Small
- Initial canary weight: 5-10%
- Monitor cho ít nhất 1-2 hours
- Chỉ tăng weight nếu metrics stable

### 2. Set Appropriate Thresholds
- Error threshold: 1-5% (tùy tolerance)
- Success threshold: 95-99%
- Latency threshold: +10-20% so với stable

### 3. Monitor Continuously
- Real-time metrics monitoring
- Set up alerts cho critical metrics
- Manual review cho important deployments

### 4. Have Rollback Plan
- Always có rollback plan sẵn
- Rollback phải nhanh (< 1 phút)
- Test rollback trước khi production

### 5. Document Everything
- Document deployment process
- Document rollback process
- Document lessons learned

---

## Implementation Pattern

### 1. Create Deployment

```go
versions := []CanaryVersion{
    {ID: "v1", Name: "Version 1.0", Weight: 0.9, IsStable: true},
    {ID: "v2", Name: "Version 2.0", Weight: 0.1, IsStable: false},
}

strategy := CanaryStrategy{
    Type: "gradual",
    Increment: 0.1,
    Interval: 5 * time.Minute,
    SuccessThreshold: 0.95,
    ErrorThreshold: 0.05,
    AutoPromote: true,
    AutoRollback: true,
}

deployment := manager.CreateDeployment("Test Canary", versions, strategy)
```

### 2. Start Deployment

```go
manager.StartDeployment(deployment.ID)
```

### 3. Monitor & Promote

```go
// Monitor metrics
for metrics are good:
    manager.PromoteVersion(deployment.ID, "v2", 0.1)
    wait for interval
```

### 4. Or Rollback

```go
if error_rate > threshold:
    manager.RollbackDeployment(deployment.ID)
```

---

## Common Mistakes

### 1. Starting Too Large
- **Problem**: Canary weight quá lớn từ đầu (50%+)
- **Solution**: Start nhỏ (5-10%), increase gradually

### 2. Not Monitoring
- **Problem**: Chạy canary mà không monitor metrics
- **Solution**: Set up real-time monitoring và alerts

### 3. Too Fast Promotion
- **Problem**: Tăng weight quá nhanh (every few minutes)
- **Solution**: Reasonable interval (5-30 minutes)

### 4. No Rollback Plan
- **Problem**: Không có rollback plan
- **Solution**: Always có rollback plan, test nó trước

### 5. Wrong Thresholds
- **Problem**: Thresholds quá strict hoặc quá loose
- **Solution**: Set thresholds dựa trên baseline metrics

---

## Advanced Patterns

### 1. Time-Based Canary
- Tăng weight dựa trên time thay vì metrics
- Useful cho predictable rollout schedules

### 2. User Segment Canary
- Canary cho specific user segments (e.g., beta users)
- Useful cho controlled testing

### 3. Feature Flag Canary
- Sử dụng feature flags để control canary
- Useful cho instant rollback

### 4. Multi-Stage Canary
- Multiple stages: dev → staging → canary → production
- Useful cho complex deployments

---

## Tools & Platforms

### Kubernetes
- **Kubernetes Rollouts**: Native canary support
- **Argo Rollouts**: Advanced canary deployment
- **Flagger**: Progressive delivery cho Kubernetes

### Cloud Platforms
- **AWS**: CodeDeploy với canary
- **Google Cloud**: Cloud Deploy với canary
- **Azure**: Deployment slots

### CI/CD
- **Spinnaker**: Multi-cloud deployment
- **GitLab**: Canary deployments
- **CircleCI**: Canary với workflows

---

## Comparison: Canary vs Blue-Green

| Aspect | Canary | Blue-Green |
|--------|--------|------------|
| Risk | Lower (gradual) | Higher (instant switch) |
| Speed | Slower (gradual) | Faster (instant) |
| Rollback | Easier (gradual) | Harder (need switch back) |
| Complexity | Higher | Lower |
| Use Case | High-risk changes | Low-risk changes |

---

## Ti Router Implementation

**File**: `layers/routing/canary.go`

**Features**:
- CanaryManager với in-memory storage
- Support cho gradual promotion
- Auto-rollback khi error rate > threshold
- Metrics tracking per version
- Thread-safe với mutex

**Usage**:
```go
manager := NewCanaryManager()
deployment := manager.CreateDeployment(...)
manager.StartDeployment(deployment.ID)
version := manager.GetVersion(deployment.ID, userID)
manager.RecordMetric(deployment.ID, version.ID, success, latency)
manager.PromoteVersion(deployment.ID, "v2", 0.1)
manager.RollbackDeployment(deployment.ID)
```

---

## References

- [Kubernetes Canary Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#creating-a-deployment)
- [Argo Rollouts](https://argoproj.github.io/argo-rollouts/)
- [Progressive Delivery](https://www.weave.works/blog/progressive-delivery-explained)

---

## Next Steps

1. Implement time-based canary
2. Add user segment support
3. Implement multi-stage canary
4. Add canary analytics dashboard
5. Integrate với CI/CD pipeline
