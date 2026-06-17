---
tags: ["tibrain", "documentation", "javascript", "router", "skill"]
scopes: ["cli", "tibrain"]
last_updated: 2026-05-22
---
# 9Router Deep Dive: Account Selection & Fallback Engine

> Files: `src/sse/services/auth.js`, `src/sse/handlers/chat.js`, `open-sse/services/accountFallback.js`, `open-sse/services/combo.js`
> Date: 2026-04-28

---

## 1. Tổng quan flow

Một request chat đi qua 3 tầng fallback:

```
Tầng 1: Combo fallback   (combo.js)     — Model A → Model B → Model C
Tầng 2: Account fallback (auth.js)      — Account 1 → Account 2 → Account 3 (cùng provider)
Tầng 3: Error cooldown   (accountFallback.js) — Exponential backoff + model-level locks
```

---

## 2. Tầng 2: Account Selection (auth.js)

### 2.1 Mutex bảo vệ chọn account (dòng 8-30)

```javascript
let selectionMutex = Promise.resolve();

// Mỗi lần chọn account:
const currentMutex = selectionMutex;
let resolveMutex;
selectionMutex = new Promise(resolve => { resolveMutex = resolve; });
try {
  await currentMutex;
  // ... logic chọn account ...
} finally {
  resolveMutex();
}
```

**Tại sao cần**: Nếu 2 request song song cùng gọi `getProviderCredentials()`, không có mutex thì cả 2 có thể chọn cùng 1 account (ví dụ round-robin đang ở account X, cả 2 request đều thấy X là "current" và chọn X → quá tải X).

**Cơ chế**: Promise chain — mỗi request đợi request trước xong mới bắt đầu chọn. Không dùng semaphore, chỉ đơn giản là 1 hàng đợi FIFO.

### 2.2 Virtual connection cho free providers (dòng 35-38)

```javascript
if (FREE_PROVIDERS[providerId]?.noAuth) {
  return { id: "noauth", connectionName: "Public", isActive: true, accessToken: "public" };
}
```

Free providers (không cần OAuth/API key) được inject 1 connection giả. Điều này **đồng nhất hóa** code path — không cần if/else đặc biệt ở các tầng sau. `markAccountUnavailable()` cũng bỏ qua `connectionId === "noauth"`.

### 2.3 Lọc account khả dụng (dòng 48-53)

```javascript
const availableConnections = connections.filter(c => {
  if (excludeSet.has(c.id)) return false;        // Đã thử ở vòng lặp trước
  if (isModelLockActive(c, model)) return false;  // Đang bị lock cho model này
  return true;
});
```

**excludeSet**: Tích lũy qua vòng lặp `while(true)` trong `handleSingleModelChat()`. Account đã fail thì bị exclude, lần sau thử account khác.

**modelLock**: Mỗi account có thể bị lock riêng cho từng model (`modelLock_${model}`). Một account có thể bị lock cho `claude-sonnet` nhưng vẫn khả dụng cho `claude-haiku`.

### 2.4 Chiến lược chọn account (dòng 85-142)

| Strategy | Cách chọn | Mục đích |
|----------|-----------|----------|
| **fill-first** (default) | `availableConnections[0]` (đã sort theo priority từ DB) | Tiêu thụ quota account 1 trước, rồi đến 2, 3... |
| **round-robin** | Sticky: dùng account hiện tại N lần (`stickyLimit`), rồi chuyển LRU | Phân bổ đều, tránh 1 account bị quá tải |

**Round-robin sticky chi tiết**:
```
1. Sort by lastUsedAt desc → tìm "current" (vừa dùng gần nhất)
2. Nếu current.consecutiveUseCount < stickyLimit:
   → Tiếp tục dùng current, increment count
3. Nếu đã đủ stickyLimit:
   → Sort by lastUsedAt asc (LRU), chọn cái lâu nhất chưa dùng
   → Reset count = 1
```

**Per-provider override**: `settings.providerStrategies[providerId].fallbackStrategy` và `.stickyRoundRobinLimit`.

### 2.5 Proxy resolution (dòng 144)

Mỗi account có thể có proxy riêng (`connectionProxyEnabled`, `connectionProxyUrl`). Proxy có 2 loại:
- `http`: dùng `HTTP_PROXY` env
- `vercel`: rewrite base URL thành Vercel relay endpoint

---

## 3. Tầng 3: Error Handling & Cooldown (accountFallback.js)

### 3.1 Config-driven error rules

`ERROR_RULES` (không thấy định nghĩa trong file này, nhưng `checkFallbackError` match top-to-bottom):

```
1. Text rules: substring match trong error message
2. Status rules: match HTTP status code
3. Default: TRANSIENT_COOLDOWN_MS (fallback anyway)
```

**Kết quả**: `{ shouldFallback: boolean, cooldownMs: number, newBackoffLevel?: number }`

### 3.2 Exponential backoff (dòng 9-12)

```javascript
const level = Math.max(0, backoffLevel - 1);
const cooldown = BACKOFF_CONFIG.base * Math.pow(2, level);
return Math.min(cooldown, BACKOFF_CONFIG.max);  // max 4 min
```

Mỗi lần fail → `backoffLevel++` → cooldown tăng gấp đôi. Max level giới hạn ở `BACKOFF_CONFIG.maxLevel`.

### 3.3 Model-level locking (dòng 106-150)

**Cơ chế lưu trong DB**:
```
connection.modelLock_claude-sonnet-4-20250514 = "2026-04-28T03:00:00Z"
connection.modelLock___all = "2026-04-28T03:00:00Z"  // lock toàn account
```

**`isModelLockActive(connection, model)`**:
- Kiểm tra `modelLock_${model}` trước
- Nếu không có, kiểm tra `modelLock___all` (account-level lock)
- So sánh expiry với `Date.now()`

**Ưu điểm**: Một account bị rate limit trên model A vẫn có thể dùng model B ngay lập tức.

### 3.4 markAccountUnavailable (auth.js dòng 183-221)

```
1. Nếu upstream trả về resetsAtMs (vd: Codex quota reset at 2026-04-28T00:00:00Z):
   → Dùng resetsAtMs làm cooldown, bỏ qua exponential backoff
   → shouldFallback = true
2. Nếu không có resetsAtMs:
   → Gọi checkFallbackError(status, errorText, backoffLevel)
   → Lấy cooldownMs + newBackoffLevel
3. Tạo model lock: buildModelLockUpdate(model, cooldownMs)
4. Update DB: testStatus="unavailable", lastError, errorCode, lastErrorAt, backoffLevel
5. Log: "Account X locked modelLock_Y for Zs [status]"
6. Return { shouldFallback: true, cooldownMs }
```

### 3.5 clearAccountError (auth.js dòng 232-265)

```
1. Tìm tất cả keys bắt đầu bằng modelLock_
2. Clear: modelLock của model vừa success + tất cả modelLock đã expired
3. Nếu sau khi clear, KHÔNG CÒN active lock nào:
   → Reset: testStatus="active", lastError=null, backoffLevel=0
4. Nếu VẪN CÒN active lock (model khác đang lock):
   → Giữ nguyên testStatus="unavailable" (vì model khác vẫn bị lock)
```

**Điểm tinh tế**: Chỉ reset về active khi account HOÀN TOÀN sạch. Nếu account bị lock 2 model, 1 model success không làm account thành active.

---

## 4. Tầng 1: Combo Fallback (combo.js)

### 4.1 Round-robin rotation (dòng 21-40)

```javascript
const comboRotationState = new Map();  // In-memory, không persist

function getRotatedModels(models, comboName, strategy) {
  const currentIndex = comboRotationState.get(comboName) || 0;
  // Rotate array: đưa phần tử từ currentIndex lên đầu
  comboRotationState.set(comboName, (currentIndex + 1) % models.length);
  return rotatedModels;
}
```

Ví dụ: models = [A, B, C]
- Request 1: currentIndex=0 → [A, B, C] → nextIndex=1
- Request 2: currentIndex=1 → [B, C, A] → nextIndex=2
- Request 3: currentIndex=2 → [C, A, B] → nextIndex=0

**Reset**: Khi combo/settings thay đổi → gọi `resetComboRotation(comboName)`.

### 4.2 handleComboChat (dòng 82-172)

```
for each model in rotatedModels:
  result = await handleSingleModel(body, modelStr)

  if result.ok (2xx):
    → Return ngay
  else:
    → Parse error body lấy retryAfter
    → checkFallbackError(status, errorText)
    → if !shouldFallback: return error (không thử model tiếp theo)
    → if transient (502/503/504) + cooldown <= 5s: wait rồi thử tiếp
    → else: log, lưu lastError, continue vòng lặp

if all failed:
  if có earliestRetryAfter → return 503 + Retry-After header
  else → return 503 + "All combo models unavailable"
```

**Status code chính xác**: 503 (Service Unavailable) thay vì 406. 503 mang nghĩa "thử lại sau", client có thể retry.

---

## 5. Tầng 2 chi tiết: handleSingleModelChat (chat.js dòng 118-242)

```
while (true):
  credentials = getProviderCredentials(provider, excludeConnectionIds, model)

  if !credentials:
    if excludeSet rỗng → 404 "No active credentials"
    else → 503 "All accounts unavailable"

  if credentials.allRateLimited:
    → 503 + Retry-After (dùng earliest lock expiry)

  // Có credentials
  refreshed = checkAndRefreshToken(provider, credentials)

  // Fix cold miss cho antigravity/gemini-cli
  if (provider in [antigravity, gemini-cli]) && !refreshed.projectId:
    pid = fetch projectId
    refreshed.projectId = pid
    update DB in background

  result = handleChatCore({
    body, modelInfo, credentials: refreshed, log,
    onCredentialsRefreshed: (newCreds) => update DB   // token rotate
    onRequestSuccess: () => clearAccountError(...)  // unlock nếu success
  })

  if result.success:
    return result.response  // DONE

  // Fail → mark unavailable
  { shouldFallback } = markAccountUnavailable(
    connectionId, result.status, result.error, provider, model, result.resetsAtMs
  )

  if shouldFallback:
    excludeConnectionIds.add(connectionId)  // Thử account khác
    lastError = result.error
    continue  // Vòng lặp mới
  else:
    return result.response  // Không fallback (lỗi nghiêm trọng)
```

---

## 6. So sánh với Ti Router hiện tại

| Feature | 9Router | Ti Router |
|---------|---------|-----------|
| Multi-account per provider | ✅ Có (connections[]) | ❌ Chưa (1 key/provider) |
| Account selection strategy | ✅ fill-first / round-robin / sticky RR | ❌ Chưa |
| Model-level lock | ✅ `modelLock_${model}` | ❌ Chưa |
| Exponential backoff | ✅ Config-driven | ❌ Chưa |
| Precise quota reset | ✅ `resetsAtMs` | ❌ Chưa |
| Combo fallback chain | ✅ Combo + strategy override | ⚠️ Có route agent nhưng chưa combo |
| Mutex chọn account | ✅ Promise chain | ❌ Chưa |
| Virtual no-auth connection | ✅ `noauth` | ❌ Chưa |

---

## 7. Lessons cho Ti

1. **Mutex là bắt buộc** nếu có multi-account + round-robin. Không có mutex → race condition.
2. **Model-level lock >> account-level lock**. Một account có thể rate limit trên model expensive nhưng vẫn chạy model cheap.
3. **resetsAtMs từ upstream** (Codex quota reset) quý hơn exponential backoff. Dùng nếu có.
4. **Sticky round-robin** (dùng N lần rồi đổi) tốt hơn pure round-robin cho streaming — giữ 1 account giảm connection churn.
5. **Combo rotation in-memory** là đủ, không cần persist. State nhỏ, reset khi restart không sao.
6. **503 + Retry-After** là response đúng khi all models/accounts unavailable — client biết khi nào thử lại.
7. **Config-driven error rules** (ERROR_RULES array) cho phép điều chỉnh fallback behavior không cần deploy code.

---

## 8. Câu hỏi mở

- `ERROR_RULES` và `BACKOFF_CONFIG` định nghĩa ở đâu? (có thể `open-sse/config/errorConfig.js`)
- `handleChatCore` bên trong `open-sse` làm gì? (format translation, actual HTTP proxy, SSE streaming)
- `providerStore.js` dùng Zustand — state client-side, không liên quan routing engine
- `localDb.js` dùng LowDB (JSON file) — có lockfile, nhưng vẫn là file-based. Nếu Ti dùng SQLite thì có thể dùng row-level lock thay vì file lock.
