# Phân Tích Backend: OmniRoute Patterns vs Ti Router Hiện Tại

> **Ngày**: 2026-05-04
> **Mục đích**: So sánh backend patterns từ OmniRoute với Ti Router hiện tại để xác định điểm cần áp dụng

---

## Tóm Tắt

Ti Router đã có một số patterns tốt (account fallback, health monitoring, semantic routing), nhưng còn thiếu một số patterns quan trọng từ OmniRoute:
- ❌ Versioned migrations (chỉ có schema list)
- ❌ Domain modules pattern (SQL trực tiếp trong handlers)
- ❌ Executor pattern (provider logic trộn lẫn)
- ❌ Advanced combo routing (chỉ parallel đơn giản)
- ❌ Translator pattern (format conversion)
- ✅ MCP Server (đã có nhưng chưa đầy đủ)
- ✅ Health monitoring (đã có)
- ✅ Account fallback (đã có)

---

## So Sánh Chi Tiết

### 1. Database Layer

| Pattern | OmniRoute | Ti Router | Trạng thái | Ưu tiên |
|---------|-----------|-----------|-----------|---------|
| **Singleton DB** | ✅ `getDbInstance()` | ✅ `db.Get()` | Đã có | - |
| **WAL Journaling** | ✅ Better-sqlite3 | ❌ Không rõ | Cần kiểm tra | P2 |
| **Versioned Migrations** | ✅ 001-021 files | ❌ Schema list | **Cần áp dụng** | **P1** |
| **Migration Tracking** | ✅ `_omniroute_migrations` table | ❌ Không có | **Cần áp dụng** | **P1** |
| **Domain Modules** | ✅ 22 modules | ❌ SQL trong handlers | **Cần áp dụng** | **P1** |
| **Encryption at Rest** | ✅ AES-256-GCM | ❌ Không rõ | Cần kiểm tra | P2 |
| **Prepared Statements** | ✅ | ❌ Cần kiểm tra | Cần áp dụng | P2 |

**Kết luận**: Database layer cần cải thiện đáng kể:
1. **Versioned migrations** - Quan trọng nhất để quản lý schema evolution
2. **Domain modules** - Tách logic DB ra khỏi handlers
3. **Encryption** - Bảo mật credentials

---

### 2. Executor Pattern

| Pattern | OmniRoute | Ti Router | Trạng thái | Ưu tiên |
|---------|-----------|-----------|-----------|---------|
| **BaseExecutor Interface** | ✅ Strategy pattern | ❌ Không có | **Cần áp dụng** | **P1** |
| **Provider-Specific Executors** | ✅ Cursor, Codex, etc. | ❌ Logic trộn lẫn | **Cần áp dụng** | **P1** |
| **Executor Factory** | ✅ Factory pattern | ❌ Không có | **Cần áp dụng** | **P1** |
| **Fallback Mechanism** | ✅ Multiple base URLs | ✅ Account fallback | Đã có | - |
| **Retry Logic** | ✅ Exponential backoff | ✅ Có | Đã có | - |
| **Credential Refresh** | ✅ Auto refresh OAuth | ✅ Có | Đã có | - |
| **Timeout Handling** | ✅ Abort signals | ✅ Context timeout | Đã có | - |

**Kết luận**: Executor pattern cần áp dụng để:
1. Tách provider-specific logic ra khỏi handlers
2. Dễ dàng thêm provider mới
3. Tái sử dụng common logic

---

### 3. Combo Routing Engine

| Pattern | OmniRoute | Ti Router | Trạng thái | Ưu tiên |
|---------|-----------|-----------|-----------|---------|
| **Config-Based Combo** | ✅ DB config | ❌ Hardcoded | **Cần áp dụng** | **P1** |
| **13 Routing Strategies** | ✅ Priority, weighted, etc. | ❌ Chỉ parallel | **Cần áp dụng** | **P1** |
| **Combo Resolution** | ✅ Ordered targets | ❌ Chỉ list | **Cần áp dụng** | **P1** |
| **Circuit Breaker Integration** | ✅ | ✅ Health monitor | Đã có | - |
| **Fallback Chains** | ✅ | ❌ Không rõ | Cần kiểm tra | P2 |
| **Cost-Aware Routing** | ✅ | ❌ Không có | Cần áp dụng | P2 |
| **Latency-Aware Routing** | ✅ | ✅ Least latency | Đã có | - |

**Kết luận**: Combo routing cần nâng cấp:
1. **Config-based** - Lưu combo config trong DB
2. **Multiple strategies** - Priority, weighted, round-robin
3. **Cost-aware** - Route theo giá rẻ nhất

---

### 4. Translator Pattern

| Pattern | OmniRoute | Ti Router | Trạng thái | Ưu tiên |
|---------|-----------|-----------|-----------|---------|
| **Format Registry** | ✅ OpenAI, Anthropic, Gemini | ❌ Không có | **Cần áp dụng** | **P2** |
| **Request Translation** | ✅ Normalize → Translate | ❌ Không có | **Cần áp dụng** | **P2** |
| **Response Translation** | ✅ Reverse of request | ❌ Không có | **Cần áp dụng** | **P2** |
| **Canonical Format** | ✅ OpenAI | ❌ Không có | **Cần áp dụng** | **P2** |
| **Tool Call Normalization** | ✅ | ❌ Không có | Cần áp dụng | P3 |
| **Role Normalization** | ✅ | ❌ Không có | Cần áp dụng | P3 |

**Kết luận**: Translator pattern quan trọng nếu muốn hỗ trợ nhiều provider formats:
1. OpenAI, Anthropic, Gemini formats khác nhau
2. Cần normalize để xử lý nhất quán
3. Priority thấp nếu chỉ hỗ trợ OpenAI-compatible

---

### 5. MCP Server

| Pattern | OmniRoute | Ti Router | Trạng thái | Ưu tiên |
|---------|-----------|-----------|-----------|---------|
| **Tool Definition** | ✅ 29 tools | ✅ Có nhưng ít | Cần mở rộng | P2 |
| **Zod Validation** | ✅ | ❌ Go validation | Cần áp dụng | P2 |
| **Scope Enforcement** | ✅ 10 scopes | ❌ Không có | **Cần áp dụng** | **P1** |
| **Audit Logging** | ✅ SHA-256 hash | ❌ Không có | **Cần áp dụng** | **P1** |
| **Multiple Transports** | ✅ stdio, SSE, HTTP | ❌ Cần kiểm tra | Cần kiểm tra | P2 |

**Kết luận**: MCP Server cần cải thiện:
1. **Scope enforcement** - Quan trọng cho security
2. **Audit logging** - Quan trọng cho compliance
3. **More tools** - Mở rộng functionality

---

### 6. Services Pattern

| Service | OmniRoute | Ti Router | Trạng thái | Ưu tiên |
|---------|-----------|-----------|-----------|---------|
| **Rate Limit Manager** | ✅ Token bucket | ❌ Không có | **Cần áp dụng** | **P1** |
| **Token Refresh** | ✅ Auto refresh | ✅ Có | Đã có | - |
| **Account Fallback** | ✅ Strategy-based | ✅ Account manager | Đã có | - |
| **Quota Monitoring** | ✅ | ✅ Usage logs | Đã có | - |
| **Session Management** | ✅ | ❌ Không rõ | Cần kiểm tra | P2 |

**Kết luận**: Rate limiting là thiếu sót quan trọng nhất:
1. **Token bucket** - Hiệu quả cho rate limiting
2. Per-provider rate limits
3. Per-user rate limits

---

### 7. API Route Pattern

| Pattern | OmniRoute | Ti Router | Trạng thái | Ưu tiên |
|---------|-----------|-----------|-----------|---------|
| **CORS Preflight** | ✅ | ✅ Có | Đã có | - |
| **Body Validation** | ✅ Zod | ❌ Go validation | Cần áp dụng | P2 |
| **Optional Auth** | ✅ | ✅ Có | Đã có | - |
| **API Key Policy** | ✅ Model permissions | ✅ Có | Đã có | - |
| **Prompt Injection Guard** | ✅ Clone & detect | ❌ Không có | Cần áp dụng | P2 |
| **Handler Delegation** | ✅ | ✅ Có | Đã có | - |

**Kết luận**: API routes khá tốt, chỉ thiếu:
1. **Prompt injection guard** - Security
2. **Structured validation** - Zod equivalent trong Go

---

## Điểm Cần Áp Dụng (Theo Ưu tiên)

### P1 (Cao) - Cần áp dụng ngay

1. **Versioned Migrations**
   - Hiện tại: Schema list trong code
   - Cần: File migrations với version tracking
   - Lợi ích: Quản lý schema evolution, rollback, idempotent

2. **Domain Modules Pattern**
   - Hiện tại: SQL trực tiếp trong handlers
   - Cần: Tách ra thành modules riêng (`pkg/db/`)
   - Lợi ích: Tái sử dụng, test dễ hơn, separation of concerns

3. **Executor Pattern**
   - Hiện tại: Provider logic trộn lẫn
   - Cần: BaseExecutor + provider-specific executors
   - Lợi ích: Dễ thêm provider, tái sử dụng logic

4. **Config-Based Combo Routing**
   - Hiện tại: Hardcoded
   - Cần: Lưu combo config trong DB
   - Lợi ích: Dynamic routing, không cần deploy để thay đổi

5. **Multiple Routing Strategies**
   - Hiện tại: Chỉ parallel
   - Cần: Priority, weighted, round-robin
   - Lợi ích: Flexibility, cost optimization

6. **Scope Enforcement (MCP)**
   - Hiện tại: Không có
   - Cần: Fine-grained scopes cho tools
   - Lợi ích: Security, access control

7. **Audit Logging (MCP)**
   - Hiện tại: Không có
   - Cần: SHA-256 hash input
   - Lợi ích: Compliance, debugging

8. **Rate Limit Manager**
   - Hiện tại: Không có
   - Cần: Token bucket algorithm
   - Lợi ích: Prevent abuse, fair usage

### P2 (Trung bình) - Cần áp dụng sau

9. **Encryption at Rest**
   - Lợi ích: Security cho credentials

10. **Translator Pattern**
    - Lợi ích: Hỗ trợ nhiều provider formats

11. **More MCP Tools**
    - Lợi ích: Mở rộng functionality

12. **WAL Journaling**
    - Lợi ích: Concurrent reads

13. **Prompt Injection Guard**
    - Lợi ích: Security

### P3 (Thấp) - Có thể bỏ qua

14. **Tool Call Normalization**
15. **Role Normalization**
16. **All 13 Routing Strategies** (chỉ cần 3-5 phổ biến)

---

## Kế Hoạch Áp Dụng Đề Xuất

### Phase 1: Database Layer (1-2 ngày)
- [ ] Implement versioned migrations
- [ ] Add migration tracking table
- [ ] Create domain modules pattern
- [ ] Move SQL logic to domain modules

### Phase 2: Executor Pattern (2-3 ngày)
- [ ] Create BaseExecutor interface
- [ ] Implement provider-specific executors
- [ ] Add executor factory
- [ ] Refactor handlers to use executors

### Phase 3: Combo Routing (2-3 ngày)
- [ ] Create combo config table
- [ ] Implement combo resolution logic
- [ ] Add routing strategies (priority, weighted, round-robin)
- [ ] Refactor combo routing to use config

### Phase 4: MCP Server (1-2 ngày)
- [ ] Add scope enforcement
- [ ] Implement audit logging
- [ ] Add more tools (health, quota, cost)

### Phase 5: Rate Limiting (1 ngày)
- [ ] Implement token bucket algorithm
- [ ] Add per-provider rate limits
- [ ] Add per-user rate limits

**Tổng thời gian**: ~7-11 ngày

---

## Kết Luận

Ti Router đã có architecture tốt (layered, health monitoring, account fallback), nhưng còn thiếu một số patterns quan trọng từ OmniRoute:

**Ưu tiên cao nhất**:
1. Versioned migrations - Quan trọng cho database management
2. Domain modules - Tách logic DB
3. Executor pattern - Tách provider logic
4. Config-based combo routing - Dynamic routing
5. Rate limiting - Prevent abuse
6. MCP scope enforcement - Security

**Ưu tiên thấp hơn**:
- Translator pattern (nếu chỉ hỗ trợ OpenAI-compatible)
- Encryption (nếu credentials không nhạy cảm)
- Full 13 routing strategies (chỉ cần 3-5 phổ biến)

Áp dụng các patterns này sẽ giúp Ti Router:
- Dễ maintain hơn
- Dễ mở rộng hơn
- Bảo mật hơn
- Performance tốt hơn
