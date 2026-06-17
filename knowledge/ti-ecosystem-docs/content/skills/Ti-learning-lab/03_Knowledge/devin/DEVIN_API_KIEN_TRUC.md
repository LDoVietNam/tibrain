# Kiến trúc Devin API - Tổng hợp Kiến thức

> **Ngày**: 2026-04-29
> **Nguồn**: Research repo `revanthpobala/devin-cli` (Python) + Devin Docs
> **Ngôn ngữ**: Tiếng Việt
> **Mục đích**: Hiểu rõ Devin API để port sang Go cho Ti CLI

---

## 1. Devin là gì?

**Devin** (Cognition Labs) là một **AI Software Engineer** tự chủ — không phải LLM API thông thường.

Khác biệt cốt lõi:
- **OpenAI/Claude**: Gọi `/chat/completions` → nhận text response ngay lập tức
- **Devin**: Gọi `/sessions` → Devin tự chạy task (browse web, edit code, run terminal) → trả kết quả sau phút

---

## 2. Authentication

```http
Authorization: Bearer {api_token}
```

- Token format: `cog_...` (service token) hoặc personal token
- **Env var precedence**: `DEVIN_API_TOKEN` > config file
- **Org ID**: `DEVIN_ORG_ID` hoặc trong config

---

## 3. Base URL & API Versions

| Version | Base URL | Trạng thái |
|---------|----------|-----------|
| v1 (Legacy) | `https://api.devin.ai/v1` | Deprecated |
| v3 (Current) | `https://api.devin.ai/v3/organizations/{org_id}` | Active |

**URL Format cho v3:**
```
https://api.devin.ai/v3/organizations/{org_id}/sessions
https://api.devin.ai/v3/organizations/{org_id}/sessions/{session_id}
https://api.devin.ai/v3/organizations/{org_id}/sessions/{session_id}/messages
```

---

## 4. Core API Endpoints

### 4.1 Sessions (Quan trọng nhất)

```http
POST /sessions
Content-Type: application/json

{
  "prompt": "Build a React app with TypeScript",
  "title": "Optional title",
  "advanced_mode": "batch",     // Optional: batch, analyze, create_playbook, ...
  "playbook_id": "...",        // Optional
  "bypass_approval": true,     // Optional: auto-approve actions
  "repos": ["owner/repo"],     // Optional: linked repos
  "knowledge_ids": ["..."],    // Optional
  "tags": ["urgent"],          // Optional
  "max_acu_limit": 100         // Optional: cost control
}
```

**Response:**
```json
{
  "session_id": "sess_abc123",
  "id": "sess_abc123",
  "status": "running",
  "url": "https://app.devin.ai/sessions/sess_abc123"
}
```

### 4.2 Poll Session Status

```http
GET /sessions/{session_id}
```

**Response:**
```json
{
  "session_id": "sess_abc123",
  "status": "running",        // running | completed | failed | cancelled
  "title": "Build React app",
  "created_at": "2026-04-29T10:00:00Z",
  "updated_at": "2026-04-29T10:05:00Z"
}
```

### 4.3 Get Messages (Output)

```http
GET /sessions/{session_id}/messages
```

**Response:**
```json
{
  "messages": [
    {"role": "user", "content": "Build a React app"},
    {"role": "assistant", "content": "I've created a React app with TypeScript..."}
  ]
}
```

### 4.4 Send Follow-up Message

```http
POST /sessions/{session_id}/messages
Content-Type: application/json

{"message": "Add routing with React Router"}
```

### 4.5 Terminate Session

```http
POST /sessions/{session_id}/terminate
```

---

## 5. Session Status Lifecycle

```
CREATED → RUNNING → [COMPLETED | FAILED | CANCELLED]
```

**Polling Strategy:**
- Interval: 5 giây (configurable)
- Timeout: Do caller control qua `context.Context`
- **Devin session có thể mất 1-30 phút** để hoàn thành

---

## 6. Error Handling Patterns

| Status | Ý nghĩa | Xử lý |
|--------|---------|-------|
| 401 | Token invalid/expired | Yêu cầu re-configure |
| 403 | Insufficient permissions | Kiểm tra org_id, token scope |
| 404 | Session not found | Fallback: list sessions + filter |
| 422 | Validation error | Kiểm tra request body |
| 429 | Rate limit | Exponential backoff |
| 5xx | Server error | Retry + log |

**Quan trọng**: `cog_` tokens (service tokens) thường bị 403/404 với `GET /sessions/{id}` trực tiếp.
→ **Fallback**: Dùng `GET /sessions?session_ids=[id]` để lấy từ list.

---

## 7. Các API Modules khác

| Module | Endpoints | Mục đích |
|--------|-----------|----------|
| **Knowledge** | `GET/POST/DELETE /knowledge` | Lưu context dùng lại cho sessions |
| **Playbooks** | `GET/POST/PUT/DELETE /playbooks` | Workflow templates |
| **Secrets** | `GET/POST/DELETE /secrets` | Sensitive data (API keys, tokens) |
| **Schedules** | `GET/POST /schedules` | Lên lịch chạy sessions |
| **Repos** | `GET /repos` | Quản lý GitHub repositories |
| **Attachments** | `GET/POST /attachments` | File uploads |

---

## 8. Key Insights cho Implementation

1. **Session-based, không phải chat-based**: Không có streaming real-time. Phải poll.
2. **Context cancellation quan trọng**: `context.Context` phải truyền xuyên suốt.
3. **Timeout dài**: 120s cho HTTP request, nhưng session polling có thể kéo dài phút.
4. **Env var precedence**: Luôn ưu tiên `DEVIN_API_TOKEN`, `DEVIN_ORG_ID`.
5. **Multi-profile**: Có thể có nhiều workspace/org cùng lúc.

---

## 9. So sánh với OpenAI/Claude SDK

| Feature | OpenAI Go SDK | Devin (cần implement) |
|---------|--------------|---------------------|
| Client init | `openai.NewClient(key)` | `devin.NewClient(opts...)` |
| Request | `CreateChatCompletion()` | `CreateSession()` |
| Response ngay | ✅ Có | ❌ Không (async) |
| Streaming | ✅ SSE | ❌ Không |
| Polling | ❌ Không cần | ✅ Bắt buộc |
| Context | `context.Context` | `context.Context` (timeout control) |
| Error types | `APIError` | `APIError` (custom) |
