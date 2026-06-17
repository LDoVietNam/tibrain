# Episodic Memory & Semantic Routing - Nghiên Cứu & Best Practices

> **Ngày tạo**: 2026-05-04  
> **Nguồn**: Phân tích codebase Ti Router + Kiến thức ngành  
> **Mục đích**: Tài liệu tham khảo cho Phase 2 & 3 của Ti Router Optimization

---

## 1. Tổng quan về Episodic Memory trong LLM Applications

### 1.1 Định nghĩa
**Episodic Memory** là hệ thống lưu trữ ký ức theo từng "episode" hoặc sự kiện cụ thể, tương tự cách con người nhớ lại các trải nghiệm chi tiết theo thời gian.

Trong bối cảnh LLM:
- **Working Memory**: Context hiện tại của conversation (5-10 turns gần nhất)
- **Episodic Memory**: Tóm tắt các turns cũ đã được compress
- **Semantic Memory**: Kiến thức dài hạn, độc lập với session cụ thể

### 1.2 Tại sao cần Episodic Memory?

**Vấn đề Context Window Limitation:**
- LLM có giới hạn context window (4k-200k tokens tùy model)
- Long conversations vượt quá limit → mất context cũ
- Cost tăng tuyến tính với số tokens

**Giải pháp Episodic Memory:**
- Compress old turns thành summaries
- Giữ raw context cho recent turns
- Retrieve relevant memory khi cần

---

## 2. Kiến trúc 3-Layer Memory System

### 2.1 Layer 1: Working Memory (Bộ nhớ làm việc)
```
Mục đích: Lưu trữ context hiện tại
- 5-10 turns gần nhất (configurable)
- Raw context (không compress)
- Token count tracking
- Auto-cleanup khi vượt quá limit
```

**Implementation trong Ti Router:**
```go
// layers/memory/episodic.go
type EpisodicMemoryEntry struct {
    SessionID  string
    TurnID     int
    Type       MemoryType // "working"
    RawContext string
    TokenCount int
    CreatedAt  time.Time
}

// GetRecentWorkingMemory(sessionID, limit)
// StoreWorkingMemory(sessionID, turnID, rawContext, tokenCount)
```

### 2.2 Layer 2: Episodic Memory (Bộ nhớ sự kiện)
```
Mục đích: Lưu trữ summaries của các turns cũ
- Compressed summaries (từ Haiku/GPT-4o-mini)
- Metadata: turn ID, token count, timestamp
- Queryable by session
- TTL-based expiration
```

**Best Practices:**
- Use smaller models for summarization (cost-effective)
- Summary length: 50-100 tokens per turn
- Keep metadata for relevance scoring
- Implement semantic search with embeddings

### 2.3 Layer 3: Semantic Memory (Bộ nhớ ngữ nghĩa)
```
Mục đích: Kiến thức dài hạn, cross-session
- Domain knowledge
- User preferences
- Patterns learned over time
- Vector embeddings cho semantic search
```

**Implementation Ideas:**
```go
// StoreSemanticMemory(summary, metadata)
// GetSemanticMemory(limit)
// Future: Add vector embeddings for semantic search
```

---

## 3. Semantic Routing - Phân loại Intent

### 3.1 Định nghĩa
**Semantic Routing** là kỹ thuật phân loại intent của user query và route đến model/tool phù hợp dựa trên semantic meaning, không phải keyword matching.

### 3.2 Các approach phổ biến

**Approach 1: Rule-based + Keywords**
```python
if "code" in query or "debug" in query:
    route_to("coder_agent")
elif "research" in query or "find" in query:
    route_to("researcher_agent")
```
- ❌ Không scalable
- ❌ Không handle edge cases
- ✅ Simple, fast

**Approach 2: Intent Classification with Embeddings**
```python
# Train classifier on labeled data
query_embedding = embed(query)
intent = classifier.predict(query_embedding)
route_to(intent_to_model[intent])
```
- ✅ Semantic understanding
- ✅ Scalable
- ⚠️ Cần training data
- ⚠️ Model maintenance overhead

**Approach 3: LLM-based Classification**
```python
# Use LLM to classify intent
response = llm.predict(f"""
Classify this query: {query}
Categories: coding, research, writing, chat
Output: category only
""")
route_to(response.strip())
```
- ✅ No training needed
- ✅ Flexible, easy to update
- ⚠️ Higher latency
- ⚠️ Cost per query

### 3.3 Best Practices cho Semantic Routing

**Model Selection per Intent:**
```
Coding tasks → Claude Opus/GPT-4 (complex reasoning)
Research tasks → Claude Haiku/GPT-4o-mini (fast, cost-effective)
Writing tasks → Claude Sonnet/GPT-4 (balanced)
Chat tasks → Claude Haiku/GPT-4o-mini (fast, conversational)
```

**Hybrid Approach:**
```
1. Fast path: Keyword matching cho common intents
2. Fallback: LLM classification cho edge cases
3. Cache: Cache classification results cho similar queries
```

---

## 4. Context Compression Techniques

### 4.1 Token-based Compression
```
Trigger: Khi context > 80% context window
Action: 
  - Compress oldest 20% turns
  - Keep recent 80% raw
  - Append compressed summary
```

### 4.2 Time-based Compression
```
Trigger: Khi turn > threshold (e.g., 50 turns)
Action:
  - Compress turns older than 24h
  - Keep recent 24h raw
```

### 4.3 Semantic Compression
```
Trigger: Khi similarity score < threshold
Action:
  - Group similar turns
  - Compress each group into single summary
  - Preserve diverse perspectives
```

### 4.4 Implementation trong Ti Router

**Current State (Phase 1):**
```go
// layers/memory/episodic.go
// Infrastructure đã có:
- SQLite storage
- 3-layer schema
- TTL-based cleanup
// Cần thêm (Phase 2):
- Auto-summarization with Haiku
- Compression triggers
- Vector embeddings cho semantic search
```

---

## 5. Integration Patterns

### 5.1 Memory Injection Flow
```
User Request
    ↓
Extract/Generate Session ID
    ↓
MemoryMiddleware.InjectContext()
    ↓
  ├─ GetRecentWorkingMemory (10 turns)
  ├─ GetEpisodicMemory (5 summaries)
  └─ GetSemanticMemory (20 entries)
    ↓
Inject into System Message
    ↓
LLM Processing
    ↓
Response
    ↓
StoreWorkingMemory(current turn)
[Future] SummarizeTurn(old turns)
```

### 5.2 Error Handling
```go
// Graceful degradation
if memory injection fails:
    log error
    continue without memory
    // Không block request
```

### 5.3 Performance Considerations
```
- Async memory storage (non-blocking)
- Batch memory cleanup (hourly)
- Cache memory queries (session-level)
- Index database (session_id, type, expires_at)
```

---

## 6. References & Best Practices from Industry

### 6.1 LangChain Memory Patterns
```
- ConversationBufferMemory: Keep all messages
- ConversationBufferWindowMemory: Keep last k messages
- ConversationSummaryMemory: Summarize old messages
- ConversationKGMemory: Knowledge graph-based
- VectorStoreMemory: Embedding-based retrieval
```

### 6.2 MemGPT Architecture
```
- Hierarchical memory: Context, Recall, Archive
- Memory operations: Read, Write, Search
- Memory types: Episodic, Semantic, Procedural
- LLM-controlled memory management
```

### 6.3 AutoGPT Memory
```
- Vector database for embeddings
- Similarity search for relevant context
- Time-based decay for old memories
- Importance scoring for retention
```

---

## 7. Recommendations cho Ti Router Phase 2

### 7.1 Priority 1: Auto-summarization
```go
// Implement after response
func (m *Middleware) PostProcessResponse(sessionID, turnID, response) {
    // 1. Store working memory
    m.StoreWorkingMemory(sessionID, turnID, rawContext, tokenCount)
    
    // 2. Check if need compression
    if m.ShouldCompress(sessionID) {
        // 3. Summarize with Haiku
        summary := m.SummarizeWithHaiku(oldContext)
        m.SummarizeTurn(sessionID, oldTurnID, summary)
    }
}
```

### 7.2 Priority 2: Semantic Routing
```go
// Implement intent classifier
type IntentClassifier struct {
    // Option 1: Keyword-based (fast)
    // Option 2: Embedding-based (accurate)
    // Option 3: LLM-based (flexible)
}

func (ic *IntentClassifier) Classify(query string) Intent {
    // Hybrid approach
    if keyword_match := ic.KeywordMatch(query); keyword_match != nil {
        return keyword_match
    }
    return ic.LLMClassify(query)
}
```

### 7.3 Priority 3: Vector Embeddings
```go
// Add to semantic memory
type SemanticMemory struct {
    embeddings *VectorStore // e.g., pgvector, chroma
}

func (sm *SemanticMemory) StoreWithEmbedding(summary string) {
    embedding := sm.embed(summary)
    sm.vectorStore.Insert(embedding, summary)
}

func (sm *SemanticMemory) SemanticSearch(query string, limit int) []string {
    queryEmbedding := sm.embed(query)
    return sm.vectorStore.Search(queryEmbedding, limit)
}
```

---

## 8. Lessons Learned từ Phase 1

### 8.1 Implementation Issues
```
✅ Fixed: MemoryEntry naming conflict với tiered_memory.go
   Solution: Renamed to EpisodicMemoryEntry

✅ Fixed: Duplicate memory import in main.go
   Solution: Removed duplicate import

✅ Fixed: Missing 'range' keyword in loop
   Solution: Added 'range' to for loop
```

### 8.2 Architecture Decisions
```
✅ Chọn SQLite thay vì in-memory
   - Persistent across restarts
   - Scalable to large datasets
   - Simple setup (no external dependencies)

✅ 3-layer architecture
   - Clear separation of concerns
   - Flexible configuration
   - Easy to extend

✅ HTTP endpoints cho management
   - Easy debugging
   - Manual intervention capability
   - Monitoring & observability
```

### 8.3 Future Improvements
```
⏳ Add vector embeddings cho semantic search
⏳ Implement auto-summarization with Haiku
⏳ Add compression triggers (token/time-based)
⏳ Implement semantic routing classifier
⏳ Add memory importance scoring
⏳ Implement cross-session knowledge transfer
```

---

## 9. Next Steps

### Phase 2 Tasks:
1. [ ] Implement auto-summarization with Haiku
2. [ ] Add compression triggers (token/time-based)
3. [ ] Build intent classifier
4. [ ] Implement intent-to-model mapping
5. [ ] Add semantic routing middleware
6. [ ] Test semantic routing accuracy

### Phase 3 Tasks:
1. [ ] Design orchestrator agent architecture
2. [ ] Implement specialist agents (coder, researcher, planner)
3. [ ] Add task delegation logic
4. [ ] Implement session persistence
5. [ ] Add comprehensive observability
6. [ ] Build eval framework

---

**Document Status**: Draft v1.0  
**Last Updated**: 2026-05-04  
**Next Review**: After Phase 2 completion
