# Kiến trúc Devin API

> Ngày: 2026-04-29 | Agent: claude | Nguồn: revanthpobala/devin-cli + Devin Docs

## Tóm tắt

Devin (Cognition Labs) là **AI Software Engineer** tự chủ, không phải LLM API dạng `/chat/completions`.

## Authentication

```http
Authorization: Bearer {api_token}   // prefix cog_
```

**Precedence**: `DEVIN_API_TOKEN` (env) > config file

## Base URL

| Version | URL |
|---------|-----|
| v1 (Legacy) | `https://api.devin.ai/v1` |
| v3 (Current) | `https://api.devin.ai/v3/organizations/{org_id}` |

## Core Endpoints

### Sessions
```http
POST /sessions                    # Tạo session với prompt
GET  /sessions/{id}             # Lấy status
GET  /sessions/{id}/messages      # Lấy output
POST /sessions/{id}/messages      # Gửi follow-up
POST /sessions/{id}/terminate     # Dừng session
```

### Khác
- `/knowledge` - Context tái sử dụng
- `/playbooks` - Workflow templates
- `/secrets` - Sensitive data
- `/schedules` - Lên lịch
- `/repos` - GitHub repos

## Session Lifecycle

```
CREATED → RUNNING → [COMPLETED | FAILED | CANCELLED]
```

**Polling**: 5s interval, context cancellation để control timeout.
**Thời gian**: 1-30 phút tùy task.

## Error Handling

| Status | Ý nghĩa | Xử lý đặc biệt |
|--------|---------|---------------|
| 401 | Token invalid | Yêu cầu re-configure |
| 403 | No permission | Check org_id, token scope |
| 404 | Not found | **Fallback**: list + filter cho cog_ tokens |
| 422 | Validation error | Check request body |
| 429 | Rate limit | Backoff |
| 5xx | Server error | Retry |

## Khác biệt với OpenAI/Claude

| | OpenAI | Devin |
|--|--------|-------|
| Response | Ngay lập tức | Async (poll) |
| Model | User chọn | Devin tự chọn |
| Output | Text | Task result (code, PR, file) |
| Streaming | SSE | Không có |
