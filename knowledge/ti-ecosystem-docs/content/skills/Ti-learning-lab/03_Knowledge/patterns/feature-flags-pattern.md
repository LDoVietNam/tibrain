---
tags: ["tibrain", "documentation", "skill"]
scopes: ["tibrain"]
last_updated: 2026-05-22
---
# Feature Flags Pattern

> **Category**: Configuration & Release  
> **Language**: Tiếng Việt  
> **Last Updated**: 2026-04-30

---

## Tổng Quan

Feature flags là kỹ thuật enable/disable features trong production mà không cần deployment. Cho phép dynamic configuration, gradual rollout, và A/B testing.

---

## Khi Nào Sử Dụng

- Enable/disable features mà không cần deployment
- Gradual rollout cho new features
- A/B testing cho features
- Kill switch cho problematic features
- User-specific features (beta users, premium users)

---

## Cấu Trúc

### Feature Flag

```go
type FeatureFlag struct {
    ID          string
    Name        string
    Description string
    Enabled     bool
    Type        FlagType
    Rules       []FlagRule
    Metadata    map[string]interface{}
}
```

### Flag Types

**1. Boolean (Boolean)**
- Simple on/off
- Dễ implement
- Useful cho kill switches

**2. Percentage (Phần trăm)**
- Enable cho percentage của users
- Useful cho gradual rollout
- Deterministic based on user hash

**3. User List (Danh sách user)**
- Enable cho specific users
- Useful cho beta testing
- Manual control

**4. Custom (Tùy chỉnh)**
- Custom logic
- Flexible nhưng phức tạp
- Cần evaluation engine

### Rules

```go
type FlagRule struct {
    ID        string
    Type      string // "user_id", "percentage", "custom"
    Value     interface{}
    Condition string // "equals", "contains", "greater_than", "less_than"
}
```

---

## Implementation Pattern

### 1. Create Flag

```go
flag := manager.CreateFlag(
    "new_feature",
    "New feature for beta users",
    FlagTypeBoolean,
    true, // enabled
)
```

### 2. Add Rules

```go
manager.AddRule(flag.ID, FlagRule{
    ID:        "rule1",
    Type:      "user_id",
    Value:     "user123",
    Condition: "equals",
})
```

### 3. Check Flag

```go
context := map[string]interface{}{
    "user_id": "user123",
}

enabled, err := manager.IsEnabled(flag.ID, context)
if enabled {
    // Enable feature
} else {
    // Disable feature
}
```

### 4. Update Flag

```go
manager.UpdateFlag(flag.ID, false) // Disable
```

---

## Best Practices

### 1. Naming Convention
- Use descriptive names (e.g., "enable_new_api", "beta_feature_x")
- Use consistent prefix (e.g., "enable_", "beta_", "test_")
- Avoid generic names (e.g., "flag1", "feature2")

### 2. Default Values
- Always have default value
- Default should be safe (usually false)
- Document default behavior

### 3. Cache Evaluation
- Cache evaluation results với TTL
- Reduce load on evaluation engine
- Invalidate cache khi flag changes

### 4. Audit Logging
- Log tất cả flag changes
- Log flag evaluations (optional, cho debugging)
- Track flag usage statistics

### 5. Cleanup
- Remove unused flags
- Archive old flags
- Keep flag count manageable (< 100-200)

---

## Common Mistakes

### 1. Too Many Flags
- **Problem**: Too many flags làm system phức tạp
- **Solution**: Cleanup unused flags, merge related flags

### 2. No Default Values
- **Problem**: Flags không có default values
- **Solution**: Always set default values

### 3. Long-Lived Flags
- **Problem**: Flags tồn tại quá lâu (months/years)
- **Solution**: Convert flags to code, remove flags

### 4. No Audit Trail
- **Problem**: Không track flag changes
- **Solution**: Log tất cả flag changes với timestamp và user

### 5. Inconsistent Naming
- **Problem**: Flag names không consistent
- **Solution**: Use naming convention, enforce it

---

## Advanced Patterns

### 1. Nested Flags
- Flags depend on other flags
- Useful cho complex feature combinations

### 2. Remote Configuration
- Flags stored in remote service (e.g., LaunchDarkly, ConfigCat)
- Real-time updates without deployment
- Useful cho distributed systems

### 3. Environment-Specific Flags
- Different flags cho different environments
- Useful cho staging vs production

### 4. Flag Groups
- Group related flags
- Enable/disable entire groups at once

---

## Tools & Services

### Open Source
- **Unleash**: Open-source feature flag service
- **Flagsmith**: Open-source feature flag platform
- **ConfigCat**: Feature flag service với free tier

### Commercial
- **LaunchDarkly**: Enterprise feature flag platform
- **Optimizely**: Feature flags với A/B testing
- **Split**: Feature flag platform

### Go Libraries
- **github.com/feature-flags**: Feature flag library
- **github.com/unleash**: Unleash client for Go

---

## Comparison: Feature Flags vs Canary Deployment

| Aspect | Feature Flags | Canary Deployment |
|--------|--------------|-------------------|
| Granularity | Per user/request | Per deployment |
| Speed | Instant (no deployment) | Requires deployment |
| Risk | Lower (easy rollback) | Higher (need rollback) |
| Complexity | Higher | Lower |
| Use Case | Dynamic configuration | Infrastructure changes |

---

## Ti Router Implementation

**File**: `layers/routing/feature_flags.go`

**Features**:
- FeatureFlagManager với in-memory storage
- Support cho 4 flag types (boolean, percentage, user_list, custom)
- Rule evaluation (user_id, percentage, custom)
- Cache với TTL (5 minutes)
- Export JSON
- Thread-safe với mutex

**Usage**:
```go
manager := NewFeatureFlagManager()
flag := manager.CreateFlag("test_feature", "Test", FlagTypeBoolean, true)
manager.AddRule(flag.ID, FlagRule{...})
context := map[string]interface{}{"user_id": "user123"}
enabled, err := manager.IsEnabled(flag.ID, context)
manager.UpdateFlag(flag.ID, false)
```

---

## References

- [Feature Flags Best Practices](https://martinfowler.com/articles/feature-toggles.html)
- [LaunchDarkly Documentation](https://docs.launchdarkly.com/)
- [Unleash Open Source](https://github.com/Unleash/unleash)

---

## Next Steps

1. Implement remote configuration (LaunchDarkly/ConfigCat)
2. Add flag groups support
3. Implement nested flags
4. Add flag analytics dashboard
5. Integrate với CI/CD pipeline
