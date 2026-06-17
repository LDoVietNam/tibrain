---
tags: ["tibrain", "router", "documentation", "javascript", "skill"]
scopes: ["cli", "tibrain"]
last_updated: 2026-05-22
---
# 9Router Deep Dive: Proxy Layer & Usage Tracking

> Files: `src/lib/network/connectionProxy.js`, `outboundProxy.js`, `initOutboundProxy.js`, `src/lib/usageDb.js`
> Date: 2026-04-28

---

## 1. Proxy Layer

### 1.1 Connection Proxy (`connectionProxy.js`)

Per-connection proxy config. 3 sources (fallback):

| Priority | Source | Type |
|----------|--------|------|
| 1 | `proxyPoolId` in connection | Pool config (HTTP or Vercel) |
| 2 | Legacy `providerSpecificData.connectionProxyUrl` | Legacy HTTP proxy |
| 3 | None | Direct |

**Vercel relay** (type="vercel"): Không dùng HTTP_PROXY env, mà **rewrite base URL** thành Vercel endpoint. Proxy ở app-level thay vì network-level.

```javascript
if (proxyPool.type === "vercel") {
  return {
    source: "vercel",
    vercelRelayUrl: proxyUrl,  // Rewrite URL
    connectionProxyEnabled: false,  // Không dùng HTTP_PROXY
  };
}
```

### 1.2 Outbound Proxy (`outboundProxy.js`)

Global proxy cho tất cả outbound connections. Set env vars:
- `HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY`
- `NO_PROXY` (bypass list)

**Quan trọng**: Dùng `NINE_ROUTER_PROXY_MANAGED` flag để theo dõi. Khi disable → **chỉ xóa env vars do 9router tạo**, không xóa env vars người dùng set trước đó.

```javascript
if (!enabled && process.env.NINE_ROUTER_PROXY_MANAGED === "1") {
  delete process.env.HTTP_PROXY;  // Chỉ xóa nếu mình tạo
}
```

### 1.3 Init (`initOutboundProxy.js`)

Lazy init on startup: `ensureOutboundProxyInitialized()` → read settings → apply env.

---

## 2. Usage Tracking (`usageDb.js`)

### 2.1 Architecture

- **DB**: LowDB JSON (`usage.json`), separate từ main `db.json`
- **Cloud mode**: In-memory only (skip saving)
- **Migration**: One-time migrate `history[]` → `dailySummary{}`

### 2.2 Pending Request Tracking (In-Memory)

```javascript
global._pendingRequests = { byModel: {}, byAccount: {} };
```

| Metric | Key | Reset |
|--------|-----|-------|
| byModel | `${model} (${provider})` | START/END calls |
| byAccount | `${connectionId} → ${modelKey}` | START/END calls |

**Safety timer**: Nếu END không được gọi (client disconnect, crash), auto-clear sau **60 giây**.

```javascript
const timer = setTimeout(() => {
  pendingRequests.byModel[modelKey] = 0;
  pendingRequests.byAccount[connectionId][modelKey] = 0;
}, 60000);
```

### 2.3 Stats Emitter (`EventEmitter`)

```javascript
global._statsEmitter = new EventEmitter();
global._statsEmitter.setMaxListeners(50);
```

Emit events:
- `"pending"` — Khi START/END/error

Dashboard subscribe để real-time update.

### 2.4 Error Provider Tracking

```javascript
global._lastErrorProvider = { provider: "", ts: 0 };
// Auto-clear sau 10 giây
```

Dùng cho UI edge coloring (highlight provider đang lỗi).

### 2.5 Active Requests Query

```javascript
getActiveRequests() => {
  activeRequests: [{ model, provider, account, count }],
  recentRequests: [{ timestamp, model, provider, tokens, status }],
  errorProvider: ""
}
```

**Recent requests dedup**: Key = `${model}|${provider}|${promptTokens}|${completionTokens}|${minute}`. Tránh duplicate entries cùng phút.

### 2.6 Save Request Usage

```javascript
saveRequestUsage(entry) {
  entry.timestamp = new Date().toISOString();
  db.data.history.push(entry);
  db.data.totalRequestsLifetime++;
  aggregateToDailySummary(dailySummary, entry);
}
```

---

## 3. Lessons cho Ti

### Proxy
1. **Vercel relay pattern**: Proxy ở app-level (rewrite URL) thay vì network-level (HTTP_PROXY env) — hữu ích cho serverless/edge.
2. **Managed env flag**: Đánh dấu env vars mình tạo → chỉ xóa của mình, không touch user config.

### Usage Tracking
1. **In-memory pending + safety timer**: Theo dõi real-time active requests, auto-recover nếu crash.
2. **EventEmitter for dashboard**: Real-time updates không cần polling.
3. **Dedup recent requests**: Tránh spam history khi retry nhiều lần.
4. **Separate usage DB**: Không mix với config DB → giảm lock contention.
