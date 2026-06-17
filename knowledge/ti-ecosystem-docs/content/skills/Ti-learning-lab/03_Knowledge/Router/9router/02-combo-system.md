---
tags: ["tibrain", "documentation", "router", "skill", "provider"]
scopes: ["tibrain"]
last_updated: 2026-05-22
---
# 9Router vs Ti Router: Combo System

## Kết luận trước
Ti Router đã **có combo system tinh vi hơn 9router**, không cần học từ 9router về combo.

## Ti có, 9router không có
| Feature | Ti | 9router |
|---------|----|---------|
| Weighted random | ✅ weight per model | ❌ |
| 4 strategies | ✅ priority, round-robin, random, least-used | ❌ chỉ fallback+round-robin |
| DAG validation | ✅ Prevent circular refs | ❌ |
| Auto combo (task-aware) | ✅ `auto:<taskType>` + brain scoring | ❌ |
| Pre-check availability | ✅ Skip cooldown models before trying | ❌ |
| Telemetry phases | ✅ `telemetry.startPhase()` | ❌ |

## 9router có, Ti nên học
| Feature | 9router | Ti |
|---------|----|----|
| Transient wait | ✅ 502/503/504 + cooldown <= 5s: WAIT rồi thử tiếp | ❌ Chưa có (Ti skip ngay) |
| Combo rotation reset on update | ✅ `resetComboRotation()` khi PUT/DELETE combo | ⚠️ Có thể đã có |

## Chi tiết 9router combo

### Schema đơn giản
```javascript
combos: [{ name: "my-combo", models: ["claude/sonnet", "openai/gpt-4o"] }]
```
Chỉ string array. Không weight, không config.

### 2 strategies
1. **fallback**: Thử model 0 → 1 → 2 theo thứ tự
2. **round-robin**: Rotate array mỗi request. `comboRotationState` Map in-memory.

### handleComboChat flow
```
for each model in rotatedModels:
  result = await handleSingleModel(body, model)
  if result.ok → return (success)
  
  checkFallbackError(status, errorText)
  if !shouldFallback → return error (không thử model tiếp)
  
  // ĐIỂM ĐẶC BIỆT 9router:
  if (502/503/504 && cooldownMs <= 5000) {
    await sleep(cooldownMs);  // CHỜ provider hồi phục
    continue;  // Thử model TIẾP THEO (không phải model hiện tại)
  }
  
  lastError = error;
  continue;  // Thử model tiếp theo
```

All failed → 503 + Retry-After (nếu có earliestRetryAfter).

### CRUD
- POST /api/combos: validate regex name, check duplicate
- PUT/DELETE /api/combos/[id]: + `resetComboRotation()`
- **Không có DAG validation** — có thể tạo circular combo

## Chi tiết Ti Router combo

### Schema nâng cao
```typescript
models: [
  { model: "claude/sonnet", weight: 3 },
  "groq/llama3"  // shorthand, weight defaults to 1
]
```

### 4 strategies (comboResolver.ts)
| Strategy | Cách chọn |
|----------|-----------|
| priority | Luôn model[0] |
| round-robin | Counter++ % length |
| random | Weighted random pick |
| least-used | Min usage count |

### Pre-check availability (chat.ts)
```typescript
const checkModelAvailable = async (modelString) => {
  const modelInfo = await getModelInfo(modelString);
  if (!isModelAvailable(provider, modelInfo.model)) {
    // Model đang cooldown → skip luôn, không waste time
    return false;
  }
  const creds = await getProviderCredentials(provider);
  if (!creds || creds.allRateLimited) return false;
  return true;
};
```

### Auto combo — brain scoring
```typescript
if (modelStr === "auto" || modelStr.startsWith("auto:")) {
  const taskType = modelStr.split(":")[1] || "default";
  const result = selectAutoModel(taskType);  // Brain scoring!
  resolvedModelStr = `${result.provider}/${result.model}`;
}
```

### DAG validation
```typescript
validateComboDAG(name, [...allCombos, tempCombo]);
// Throw nếu: circular reference hoặc vượt max depth
```

## Lesson cho Ti

**Ti combo system đã tốt hơn 9router.** Chỉ cần bổ sung 1 điểm từ 9router:

> **Transient wait**: Khi model trả về 502/503/504 + cooldown ngắn (<=5s), nên **chờ** thay vì skip ngay. Provider có thể đang hồi phục sau overload.

9router implement bằng `await sleep(cooldownMs)` giữa models trong combo. Ti hiện tại skip ngay → waste fallback slot.
