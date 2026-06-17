---
tags: ["tibrain", "skill", "documentation", "javascript", "sse"]
scopes: ["integration", "cli", "tibrain"]
last_updated: 2026-05-22
---
# 9Router Deep Dive: Format Translation & SSE Streaming

> Files: `open-sse/translator/formats.js`, `open-sse/translator/index.js`, `open-sse/handlers/chatCore.js`, `open-sse/config/runtimeConfig.js`
> Date: 2026-04-28

---

## 1. Format Detection

### 1.1 By endpoint (`formats.js`)
```
/v1/responses        → OPENAI_RESPONSES
/v1/messages         → CLAUDE
/v1/chat/completions + body.input[] → OPENAI (Cursor CLI hack)
```

### 1.2 By body (`detectFormat()`)
- `body.messages[]` → OpenAI format
- `body.contents[]` → Gemini format  
- `body.input[]` → OpenAI Responses format

---

## 2. Translation Pipeline (`translator/index.js`)

**Hub-and-spoke pattern**: Tất cả formats đi qua OpenAI làm trung gian.

```
Source Format → OpenAI Format → Target Format
     (request)       (request)       (request)
```

**Request translation**:
```javascript
// Step 1: source → openai (nếu source ≠ openai)
const toOpenAI = requestRegistry.get(`${sourceFormat}:openai`);
if (toOpenAI) result = toOpenAI(model, result, stream, credentials);

// Step 2: openai → target (nếu target ≠ openai)  
const fromOpenAI = requestRegistry.get(`openai:${targetFormat}`);
if (fromOpenAI) result = fromOpenAI(model, result, stream, credentials);
```

**Response translation** (ngược lại):
```javascript
// Step 1: target → openai
// Step 2: openai → source
```

---

## 3. Pre-translation Transformations

| Step | Function | Mô tả |
|------|----------|-------|
| 1 | `compressMessages()` (RTK) | Nén tool_result content trước translation |
| 2 | `stripContentTypes()` | Bỏ image/audio nếu provider không hỗ trợ (opt-in via `strip[]`) |
| 3 | `normalizeThinkingConfig()` | Xóa thinking config nếu last message không phải user |
| 4 | `ensureToolCallIds()` | Đảm bảo tool_calls có id (một số provider require) |
| 5 | `fixMissingToolResponses()` | Insert empty tool_result nếu thiếu |

---

## 4. Native Passthrough (`chatCore.js` dòng 74-93)

```javascript
const clientTool = detectClientTool(clientRawRequest.headers, body);
const passthrough = isNativePassthrough(clientTool, provider);

if (passthrough) {
  // Skip ALL translation — chỉ swap model và Bearer token
  translatedBody = { ...body, model };
} else {
  translatedBody = translateRequest(sourceFormat, targetFormat, ...);
}
```

**Ý nghĩa**: Nếu client (vd: Claude Code) gọi provider (vd: Claude API) — cùng ecosystem — thì không cần dịch format. Chỉ cần:
1. Swap model name
2. Swap Bearer token (OAuth token thay API key)

**Giảm latency** và **tránh bug** từ translation.

---

## 5. Tool Cloaking (Anti-Ban)

### 5.1 Claude OAuth (`sk-ant-oat` token)
```javascript
if (provider === "claude" && apiKey?.includes("sk-ant-oat")) {
  const { body: cloakedBody, toolNameMap } = cloakClaudeTools(result);
  result._toolNameMap = toolNameMap;  // Để uncloak trong response
}
```
- Rename tool names thêm `_cc` suffix
- Chống detection từ Anthropic

### 5.2 Antigravity
```javascript
if (provider === "antigravity" && body.userAgent !== "antigravity") {
  const { cloakedBody, toolNameMap } = AntigravityExecutor.cloakTools(result);
}
```
- Rename tools + inject decoy tools
- Skip nếu client là native AG (không cần cloak)

---

## 6. Streaming vs Non-streaming Logic (`chatCore.js` dòng 56-67)

```javascript
// Client muốn streaming?
const clientRequestedStreaming = body.stream === true || 
  sourceFormat === FORMATS.ANTIGRAVITY || 
  sourceFormat === FORMATS.GEMINI ||
  sourceFormat === FORMATS.GEMINI_CLI;

// Provider BẮT BUỘC streaming?
const providerRequiresStreaming = provider === "openai" || provider === "codex";

// Default: client preference, nhưng provider có thể override
let stream = providerRequiresStreaming ? true : (body.stream !== false);

// FIX: AI SDK gửi Accept: application/json nhưng vẫn muốn non-streaming
const clientPrefersJson = acceptHeader.includes("application/json");
const clientPrefersSSE = acceptHeader.includes("text/event-stream");
if (clientPrefersJson && !clientPrefersSSE && body.stream !== true) {
  stream = false;
}
```

---

## 7. Provider Thinking Override (`chatCore.js` dòng 42-54)

```javascript
if (providerThinking?.mode && providerThinking.mode !== "auto") {
  const mode = providerThinking.mode;  // "on", "off", "none", "low", "medium", "high"
  
  if (mode === "on" && !body.thinking) {
    body = { ...body, thinking: { type: "enabled", budget_tokens: 10000 } };
  } else if (mode === "off" && !body.thinking) {
    body = { ...body, thinking: { type: "disabled" } };
  } else if (!body.reasoning_effort) {
    body = { ...body, reasoning_effort: mode };  // "low", "medium", "high"
  }
}
```

User có thể override thinking config ở provider-level (dashboard setting), chỉ áp dụng nếu client chưa set.

---

## 8. Response Pipeline (`chatCore.js` dòng 217-237)

```
Provider Error? → parseUpstreamError → createErrorResult(resetsAtMs)
                ↓
Provider forced streaming but client wants JSON? → handleForcedSSEToJson()
                ↓
Non-streaming? → handleNonStreamingResponse()
                ↓
Streaming? → handleStreamingResponse()
```

**handleForcedSSEToJson**: Provider (vd OpenAI) luôn trả SSE, nhưng client yêu cầu JSON. 9router buffer toàn bộ SSE → parse chunks → build JSON response.

---

## 9. HTTP Status Constants (`runtimeConfig.js`)

```javascript
HTTP_STATUS = {
  BAD_REQUEST: 400, UNAUTHORIZED: 401, PAYMENT_REQUIRED: 402,
  FORBIDDEN: 403, NOT_FOUND: 404, NOT_ACCEPTABLE: 406,
  REQUEST_TIMEOUT: 408, RATE_LIMITED: 429,
  SERVER_ERROR: 500, BAD_GATEWAY: 502, SERVICE_UNAVAILABLE: 503, GATEWAY_TIMEOUT: 504
}
```

**Retry config**:
- 429: 0 retries (rate limit — backoff/lock đã handle)
- 502: 3 retries, delay 3000ms
- 503: 3 retries, delay 2000ms  
- 504: 2 retries, delay 3000ms

---

## 10. Lessons cho Ti

1. **Native passthrough** quan trọng cho performance — cùng ecosystem thì không dịch.
2. **Accept header check** — AI SDK và một số client gửi `Accept: application/json` nhưng vẫn dùng `/v1/chat/completions`. Phải respect header.
3. **Thinking config injection** — Provider có thể override client config (nếu client chưa set). Useful cho enterprise deployment.
4. **Tool cloaking** — Nếu hỗ trợ OAuth/free tier providers có anti-tool-ban policy, cần implement.
5. **RTK (compress messages)** — Nếu context window lớn, nén tool_result trước translation tiết kiệm tokens.
