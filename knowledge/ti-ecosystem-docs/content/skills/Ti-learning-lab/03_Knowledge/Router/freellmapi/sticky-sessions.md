# Sticky Sessions (30 phút)

> **Feature**: Sticky sessions để giữ model consistency trong multi-turn conversations
> **Nguồn cảm hứng**: FreeLLMAPI - `server/src/routes/proxy.ts`
> **Trạng thái**: ✅ Hoàn thành

## Tổng Quan

Ti Router hiện hỗ trợ sticky sessions để đảm bảo rằng cùng một conversation sử dụng cùng một model trong suốt multi-turn dialogue. Điều này giúp tránh hallucinations do model switching giữa các turns.

## Tại Cần Sticky Sessions?

### Vấn đề
Khi routing multi-turn conversations:
- Turn 1: GPT-4 trả lời câu hỏi A
- Turn 2: Claude trả lời câu hỏi B (vì GPT-4 rate limit)
- Turn 3: GPT-4 tiếp tục câu hỏi C

**Kết quả**: Inconsistency trong context, model không hiểu conversation history, gây hallucinations.

### Giải pháp
Sticky sessions:
- Định danh session dựa trên first user message
- Ghi nhớ model đã sử dụng cho session đó
- Ưu tiên dùng cùng model cho các turns tiếp theo
- TTL 30 phút (sau đó session expire)

## Kiến Trúc

### File triển khai
- `Z:\Ti\router\layers\authentication\sticky_session.go` - Core sticky session logic
- `Z:\Ti\router\layers\authentication\sticky_session_test.go` - Unit tests

### Components

#### 1. StickySessionManager
Manager cho sticky sessions:

```go
type StickySessionManager struct {
    sessions map[string]*StickySessionEntry
    mu       sync.RWMutex
}
```

**Constants**:
- `stickySessionTTL = 30 * time.Minute` - TTL cho session
- `maxStickyEntries = 500` - Max entries in memory (auto-cleanup)

#### 2. Session Key Generation
Định danh session dựa trên messages:

```go
func (m *StickySessionManager) GetSessionKey(messages []providers.Message) string
```

**Logic**:
1. Tìm first user message
2. Lấy 100 chars đầu tiên làm prefix
3. Xác định single-turn vs multi-turn (>2 messages)
4. Hash: SHA-256(prefix + ":" + turn_type)
5. Return hex-encoded hash

**Ví dụ**:
```
Messages: [user: "Hello", assistant: "Hi", user: "How are you?"]
Key: SHA-256("Hello:multi") → "a1b2c3d4..."
```

#### 3. GetStickyModel
Lấy sticky model cho session:

```go
func (m *StickySessionManager) GetStickyModel(messages []providers.Message) string
```

**Logic**:
1. Chỉ áp dụng cho multi-turn (có assistant messages)
2. Tính session key
3. Lookup trong map
4. Check TTL (30 phút)
5. Return model ID nếu valid, empty string nếu không

**Return conditions**:
- Empty string nếu:
  - Không có assistant messages (single turn)
  - Session không tồn tại
  - Session đã expire (>30 phút)

#### 4. SetStickyModel
Set sticky model cho session:

```go
func (m *StickySessionManager) SetStickyModel(messages []providers.Message, modelID string)
```

**Logic**:
1. Tính session key
2. Lưu vào map: `{key: {modelID, lastUsed}}`
3. Update lastUsed timestamp
4. Trigger cleanup nếu cần

#### 5. Cleanup
Tự động cleanup expired entries:

```go
func (m *StickySessionManager) cleanup()
```

**Logic**:
1. Remove entries > TTL (30 phút)
2. Enforce size limit (max 500 entries)
3. Remove oldest entries nếu vượt limit
4. Chạy tự động sau mỗi SetStickyModel

#### 6. GetStats
Thống kê sticky session manager:

```go
func (m *StickySessionManager) GetStats() map[string]interface{}
```

**Return**:
```json
{
  "total_entries": 150,
  "expired_entries": 20,
  "active_entries": 130,
  "ttl_minutes": 30,
  "max_entries": 500
}
```

#### 7. Clear
Xóa tất cả sessions (cho testing/admin):

```go
func (m *StickySessionManager) Clear()
```

## So Sánh với FreeLLMAPI

| Feature | FreeLLMAPI (TS) | Ti Router (Go) |
|---------|-----------------|----------------|
| Session key | First 100 chars + count | First 100 chars + turn type ✅ |
| Hash algorithm | N/A (string concat) | SHA-256 ✅ |
| TTL | 30 minutes | 30 minutes ✅ |
| Max entries | 500 | 500 ✅ |
| Multi-turn detection | Has assistant messages | Has assistant messages ✅ |
| Auto-cleanup | Yes | Yes ✅ |
| Thread-safe | No (JS single-threaded) | Yes (sync.RWMutex) ✅ |

**Khác biệt chính**:
- Ti Router sử dụng SHA-256 hash thay vì string concat (đảm bảo uniqueness)
- Ti Router thread-safe với mutex (Go concurrent)
- Cả hai đều cùng logic core

## Usage Example

```go
// Create manager (thường làm ở startup)
stickyManager := authentication.NewStickySessionManager()

// Trong routing logic:
messages := []providers.Message{
    {Role: "user", Content: "Hello"},
    {Role: "assistant", Content: "Hi there!"},
    {Role: "user", Content: "How are you?"},
}

// Get sticky model (nếu có)
preferredModel := stickyManager.GetStickyModel(messages)
if preferredModel != "" {
    // Sử dụng sticky model
    routeTo(preferredModel)
} else {
    // Routing bình thường
    routeTo(selectBestModel())
}

// Sau khi request thành công
stickyManager.SetStickyModel(messages, usedModelID)
```

## Integration với Router

### Cần update để sử dụng sticky sessions:

1. **Router handler** (`layers/http/openai/handlers.go`):
   - Tạo global StickySessionManager
   - Gọi GetStickyModel trước khi routing
   - Gọi SetStickyModel sau khi request thành công

2. **Provider selection** (`layers/internal/router/router.go`):
   - Chấp nhận preferredModel parameter
   - Ưu tiên preferredModel nếu available

3. **Admin API** (`cmd/routerd/handlers_admin.go`):
   - Thêm endpoint `/admin/sticky-sessions/stats`
   - Thêm endpoint `/admin/sticky-sessions/clear`

## Testing

Tất cả tests pass:

```bash
cd Z:\Ti\router
go test ./layers/authentication/ -v -run "TestGetSession|TestGetSticky|TestCleanup|TestClear|TestGetStats"
```

**Test coverage**:
- `TestGetSessionKey` - Session key generation
- `TestGetSessionKeyLongMessage` - Long message handling
- `TestGetSessionKeyNoUserMessage` - Edge case: no user message
- `TestGetStickyModelMultiTurn` - Multi-turn sticky logic
- `TestGetStickyModelSingleTurn` - Single turn không sticky
- `TestGetStickyModelTTL` - TTL expiration
- `TestCleanup` - Auto-cleanup trên size limit
- `TestClear` - Manual clear
- `TestGetStats` - Statistics

## Performance Considerations

### ✅ Optimizations
- In-memory map (O(1) lookup)
- SHA-256 hash (fast, ~100ns)
- Mutex RWMutex (concurrent reads)
- Lazy cleanup (chỉ khi cần)

### ⚠️ Lưu ý
- Max 500 entries (memory ~50KB)
- Hash computation cho mỗi request (nhỏ)
- TTL check cho mỗi GetStickyModel

### 🔐 Production Recommendations
1. Monitor stats để track usage
2. Adjust maxStickyEntries nếu cần
3. Consider persisting sessions nếu cần cross-instance
4. Log sticky hits/misses cho analytics

## Security Considerations

### ✅ Safe
- Không lưu sensitive data (chỉ hash)
- TTL tự động expire
- Không persist plaintext messages

### ⚠️ Lưu ý
- Session key dựa trên user message (có thể guessable)
- Không dùng cho authentication/chính xác security

## Future Enhancements

1. **Persistent Storage**: Lưu sessions trong DB cho cross-instance
2. **Custom TTL**: Per-user hoặc per-model TTL
3. **Session Metadata**: Track token count, latency per session
4. **Smart Expiry**: Expire sớm nếu user inactive
5. **Session Analytics**: Track session patterns, model preferences

## References

- FreeLLMAPI: `Z:\Ti\Ti-learning-lab\05_Repositories\router\freellmapi-main\server\src\routes\proxy.ts`
- Go crypto package: https://pkg.go.dev/crypto/sha256
- Sync patterns: https://pkg.go.dev/sync

---

**Ngày tạo**: 2026-04-28
**Agent**: Claude Code
**Project**: Ti Router
