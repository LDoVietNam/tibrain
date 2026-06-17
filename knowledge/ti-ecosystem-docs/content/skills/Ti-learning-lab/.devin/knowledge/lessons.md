# Lessons Learned - Ti-Learning-Lab

> **Lessons learned từ research và development**

---

## 📚 Lessons từ RTK Research

### Lesson 1: Token Optimization is Critical
**Date**: 2026-04-28
**Source**: RTK analysis

**Lesson**: Giảm token consumption 60-90% có thể đạt được với filtering strategies đơn giản

**Key Insights**:
- Smart filtering: remove noise (comments, whitespace, boilerplate)
- Grouping: aggregate similar items
- Truncation: keep relevant context, cut redundancy
- Deduplication: collapse repeated lines

**Application**:
- Áp dụng cho Ti Router agent tool outputs
- Filter provider responses
- Filter error messages

**Impact**: Giảm cost 60-90% cho tool-heavy workflows

---

### Lesson 2: Command Proxy Architecture is Powerful
**Date**: 2026-04-28
**Source**: RTK architecture

**Lesson**: Command proxy pattern cho phép intercept, process, modify behavior transparently

**Key Insights**:
- Clap CLI routing đến specialized modules
- Filter modules per command type
- Fallback pattern cho compatibility
- SQLite tracking cho analytics

**Application**:
- Áp dụng cho Ti Router agent tool calls
- Intercept tool outputs, filter, return optimized
- Track usage metrics

**Impact**: Flexible, extensible, maintainable

---

### Lesson 3: Rust Performance is Excellent
**Date**: 2026-04-28
**Source**: RTK performance (<10ms startup, <5MB memory)

**Lesson**: Rust có thể đạt hiệu suất cao cho CLI tools

**Key Insights**:
- Startup <10ms
- Memory <5MB
- Single-threaded design
- Zero-cost abstractions

**Application**:
- Có thể viết performance-critical components trong Rust
- Go cũng tốt, nhưng Rust có thể tốt hơn cho một số use cases

**Impact**: Better performance cho critical paths

---

## 📚 Lessons từ 9Router Research

### Lesson 4: Multi-Account Per Provider is Valuable
**Date**: 2026-04-28
**Source**: 9Router account selection research

**Lesson**: Multiple accounts per provider với selection strategies improve reliability và cost

**Key Insights**:
- fill-first: tiêu thụ quota account 1 trước
- round-robin: phân bổ đều
- sticky round-robin: dùng N lần rồi đổi (tốt cho streaming)
- Mutex là bắt buộc để tránh race condition

**Application**:
- Ti Router hiện tại chỉ có 1 key/provider
- Nên thêm multi-account support
- Implement selection strategies

**Impact**: 30-50% cost reduction, better reliability

---

### Lesson 5: Model-Level Locking > Account-Level Locking
**Date**: 2026-04-28
**Source**: 9Router model-level locking

**Lesson**: Lock từng model riêng biệt cho phép granular control

**Key Insights**:
- Account có thể bị lock cho claude-sonnet nhưng vẫn dùng claude-haiku
- `modelLock_${model}` trong DB
- Reset khi model success, giữ nếu model khác vẫn lock

**Application**:
- Ti Router circuit breaker là provider-level
- Nên thêm model-level circuit breaker
- Better resource utilization

**Impact**: Better availability, less waste

---

### Lesson 6: Precise Quota Reset is Better than Exponential Backoff
**Date**: 2026-04-28
**Source**: 9Router resetsAtMs

**Lesson**: Dùng upstream quota reset time (resetsAtMs) thay vì chỉ exponential backoff

**Key Insights**:
- Codex quota reset at specific time
- Dùng resetsAtMs làm cooldown
- Fallback về exponential backoff nếu không có resetsAtMs

**Application**:
- Ti Router chỉ có exponential backoff
- Nên detect và use quota reset time từ upstream
- Hybrid approach: quota reset + exponential backoff

**Impact**: More accurate cooldown, less waste

---

## 📚 Lessons từ AI Agent Router Plan

### Lesson 7: AI Agent Router > Traditional Router for Complex Decisions
**Date**: 2026-04-28
**Source**: AI_AGENT_ROUTER_PLAN.md

**Lesson**: AI agent với think→tool→observe loop có thể đưa ra decisions tốt hơn rule-based

**Key Insights**:
- Context-aware routing (task, budget, latency, quality)
- Adaptive learning từ experience
- Explainable decisions
- Swarm intelligence với specialist agents

**Trade-offs**:
- Higher latency (agent think time)
- More complexity
- Cần memory/state management

**Application**:
- Ti Router hiện tại là rule-based
- Nên consider AI agent router cho complex routing
- Hybrid approach: AI + rule-based fallback

**Impact**: Better decisions, but higher latency

---

### Lesson 8: Semantic Caching > Exact Match Caching
**Date**: 2026-04-28
**Source**: AI_AGENT_ROUTER_PLAN.md

**Lesson**: Semantic caching với vector similarity có thể đạt cache hit rate 60%+

**Key Insights**:
- Exact match: 30% hit rate
- Semantic similarity: 60%+ hit rate
- Cần vector DB (pgvector/qdrant)
- Cần embedding client

**Application**:
- Ti Router có cache nhưng là exact match
- Nên thêm semantic caching
- Vector DB cho embeddings

**Impact**: 50-70% reduction in redundant API calls

---

## 📝 Quy Tắc Thêm Lesson Mới

### Khi học được lesson mới:

1. **Document lesson**:
   ```
   - Date
   - Source (repo/file/research)
   - Lesson statement
   - Key insights (3-5 bullets)
   - Application cho Ti
   - Impact
   ```

2. **Validate lesson**:
   ```
   - Test lesson trong practice
   - Measure impact
   - Confirm/refute lesson
   ```

3. **Update patterns** (nếu applicable):
   ```
   - Extract pattern từ lesson
   - Thêm vào patterns.md
   - Document use case
   ```

---

## 🎯 Priority Lessons để Áp Dụng

### Priority 1 (Immediate):
1. **RTK Integration** - Lesson 1 & 2
2. **Multi-Account Support** - Lesson 4
3. **Model-Level Locking** - Lesson 5

### Priority 2 (Short-term):
4. **Quota Reset Detection** - Lesson 6
5. **Semantic Caching** - Lesson 8

### Priority 3 (Long-term):
6. **AI Agent Router** - Lesson 7
7. **Rust Components** - Lesson 3

---

**Last Updated**: 2026-04-28
