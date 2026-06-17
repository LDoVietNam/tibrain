# Local Intelligence Architecture

## Overview

Local Intelligence là feature cho phép Ti Router trả lời request tại chỗ (locally) sử dụng knowledge từ Ti Brain, thay vì gửi đến external providers. Điều này mang lại:

- **0ms latency** - Không cần gọi external API
- **$0 cost** - Không tốn tiền cho provider calls
- **Privacy** - Dữ liệu không rời khỏi local system
- **Reliability** - Không phụ thuộc vào external service availability

## Architecture

```
User Request
    ↓
Ti Router (:1806)
    ↓
┌─────────────────────────────┐
│ 1. Authentication Check     │
│ 2. Extract Query            │
│ 3. Local Intelligence Check │ ← NEW
│    └─ CanAnswerLocally()?   │
│       ├─ YES → Local Response
│       └─ NO  → Provider Call
└─────────────────────────────┘
    ↓
Ti Brain Server (:1808)
    ↓
Search Brain Knowledge
    ↓
Calculate Confidence Score
    ↓
Return Local Response
```

## Components

### 1. Brain Client (`Z:\Ti\router\layers\brain\client.go`)

HTTP client để giao tiếp với Ti Brain Server:

```go
type Client struct {
    baseURL    string
    httpClient *http.Client
}

// Methods:
- GetContext() → Get full memory stack
- Search(query, limit) → Search brain knowledge
- HealthCheck() → Check brain server health
- ListDrawers() → List all drawers
```

### 2. Local Intelligence (`Z:\Ti\router\layers\brain\local_intelligence.go`)

Core logic cho local intelligence:

```go
type LocalIntelligence struct {
    client *Client
}

// CanAnswerLocally(query, model) → (bool, confidence)
// - Search brain for relevant knowledge
// - Calculate confidence score
// - Return true if confidence >= 0.7

// GenerateLocalResponse(query, model) → LocalResponse
// - Search brain for knowledge
// - Generate answer from drawers
// - Return structured response
```

### 3. Confidence Scoring Algorithm

Confidence score dựa trên 3 factors:

1. **Number of relevant drawers** (30% weight)
   - More drawers = higher confidence
   - Normalized to 0-1

2. **Average importance score** (40% weight)
   - Higher importance = higher confidence
   - Range: 0-1

3. **Category relevance** (30% weight)
   - FAQ, pattern, system categories = boost
   - Binary boost: +0.3 if match

**Formula:**
```
confidence = (countScore * 0.3) + (avgImportance * 0.4) + (categoryBoost * 0.3)
threshold = 0.7
```

### 4. Router Integration (`Z:\Ti\router\cmd\routerd\main.go`)

Local intelligence check trong `handleChatCompletions`:

```go
// Check before provider call
if brain.LocalIntel != nil && !stream {
    query := extractQueryFromMessages(reqBody)
    canAnswer, confidence := brain.LocalIntel.CanAnswerLocally(query, modelID)
    
    if canAnswer {
        localResp := brain.LocalIntel.GenerateLocalResponse(query, modelID)
        // Return local response with X-Response-Source: local_brain
        return
    }
}

// Fall back to provider call
```

## Brain Server Integration

### Endpoints Used

| Endpoint | Purpose | Method |
|----------|---------|--------|
| `/v1/brain/search` | Search knowledge | GET |
| `/v1/brain/context` | Get full memory | GET |
| `/health` | Health check | GET |

### Search Response Format

```json
{
  "l0_identity": "Ti Brain identity...",
  "l1_essential": [...],  // High-priority drawers
  "l2_on_demand": [...],   // Medium-priority drawers
  "l3_deep_search": [...]  // Low-priority drawers
}
```

## Local Response Format

```json
{
  "id": "local-1777329230876715800",
  "object": "chat.completion",
  "created": 1777329230,
  "model": "gemini-2.5-flash",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Based on local knowledge..."
    },
    "finish_reason": "stop"
  }],
  "local_response": {
    "source": "local_brain",
    "confidence": 0.85
  }
}
```

## Configuration

### Router Flags

```bash
-routerd -brain-url http://localhost:1808
```

### Environment Variables

```bash
# Brain Server URL (default: http://localhost:1808)
BRAIN_URL=http://localhost:1808

# Local Intelligence threshold (default: 0.7)
LOCAL_INTEL_THRESHOLD=0.7
```

## Usage Examples

### Example 1: Local Answer Available

```bash
curl -X POST http://localhost:1806/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TI_API_KEY" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "What is Ti Router?"}]
  }'

# Response: Local answer from brain
# Header: X-Response-Source: local_brain
```

### Example 2: No Local Answer

```bash
curl -X POST http://localhost:1806/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TI_API_KEY" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "Latest news?"}]
  }'

# Response: Provider call (OpenRouter/Groq/etc)
# Header: X-Response-Source: provider
```

## Performance Metrics

| Metric | Local Intelligence | Provider Call |
|--------|-------------------|---------------|
| Latency | ~5-10ms | 500-2000ms |
| Cost | $0 | $0.0001-0.01 per 1K tokens |
| Privacy | 100% local | Data sent to provider |
| Reliability | 99.9% (local) | 95-99% (external) |

## Limitations

1. **Knowledge dependency** - Chỉ trả lời được nếu brain có đủ knowledge
2. **Non-streaming only** - Chỉ hoạt động với non-streaming requests
3. **Confidence threshold** - Cần confidence >= 0.7 để dùng local response
4. **Answer quality** - Phụ thuộc vào chất lượng knowledge trong brain

## Future Improvements

1. **Streaming support** - Hỗ trợ streaming cho local responses
2. **ML-based confidence** - Sử dụng ML model để tính confidence chính xác hơn
3. **Hybrid mode** - Kết hợp local + provider cho câu trả lời tốt hơn
4. **Learning feedback** - Học từ user feedback để cải thiện confidence scoring
5. **Context awareness** - Sử dụng context từ conversation history

## Related Files

- `Z:\Ti\router\layers\brain\client.go` - Brain HTTP client
- `Z:\Ti\router\layers\brain\local_intelligence.go` - Local intelligence logic
- `Z:\Ti\router\layers\brain\types.go` - Data structures
- `Z:\Ti\router\cmd\routerd\main.go` - Router integration
- `Z:\Ti\Ti knowledge\brain\server\main.go` - Brain server

## References

- [Ti Brain Server Documentation](../../Ti%20knowledge/brain/README.md)
- [Router Architecture](../README.md)
- [Brain Memory Stack](../../Ti%20knowledge/brain/memory/README.md)
