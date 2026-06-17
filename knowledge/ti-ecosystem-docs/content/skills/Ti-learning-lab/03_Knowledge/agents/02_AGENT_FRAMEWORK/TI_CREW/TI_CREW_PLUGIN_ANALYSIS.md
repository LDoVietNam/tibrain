# Ti Crew Plugin Code Analysis

> **Version**: 1.0.0
> **Last Updated**: 2026-05-05
> **Purpose**: Đánh giá code trong `apps/ticrew-plugin` để quyết định nên port hay bỏ

---

## 📊 Tổng Quan

**Location**: `Z:\10_WORKPLACE\Ti\apps\ticrew-plugin\`
**Total Go Files**: 18 files
**Status**: Legacy plugin cho Ti CLI

---

## 🔍 Phân Tích Chi Tiết

### 1. Memory System (`internal/plugins/ticrew/memory/`)

#### Files:
- `bridge.go` (255 lines) - Memory bridge kết nối với Ti's memory system
- `advanced.go` (447 lines) - Advanced memory features

#### Features:
- **ContextManager**: Quản lý conversation context với compression
  - Keep last 50 entries
  - Token counting
  - Context pruning
  - Compression cho long contexts

- **ReviewSystem**: Memory review và reflection
  - Daily reviews
  - Insight extraction
  - Pattern identification
  - Review history

- **MemoryCompression**: Compression logic
  - Compress to 50% of original
  - Context compression (keep header, first/last parts)

#### GoClaw Equivalent:
- ✅ GoClaw có `internal/memory/` với pgvector + knowledge graph
- ✅ GoClaw có context management trong agent loop
- ❌ GoClaw KHÔNG có ReviewSystem (memory reflection)
- ❌ GoClaw KHÔNG có MemoryCompression tùy chỉnh

#### Recommendation: **PARTIAL PORT**
- ✅ Port **ReviewSystem** - useful cho agent learning
- ✅ Port **MemoryCompression** - useful cho cost optimization
- ❌ Bỏ **Bridge** - GoClaw có built-in memory

---

### 2. MCP Bridge (`internal/plugins/ticrew/mcp/bridge.go`)

#### Features:
- Load tools từ Ti's MCP servers
- Execute tools via MCP client
- Tool filtering by category
- Tool statistics
- Refresh tools

#### GoClaw Equivalent:
- ✅ GoClaw có `internal/mcp/` với stdio/SSE/streamable-http support
- ✅ GoClaw có tool registry trong `internal/tools/`
- ✅ GoClaw có built-in MCP bridge

#### Recommendation: **SKIP**
- ❌ GoClaw đã có đầy đủ MCP integration
- ❌ Plugin chỉ wrapper cho Ti's MCP client

---

### 3. Multi-Agent Orchestrator (`internal/plugins/ticrew/multi/orchestrator.go`)

#### Features:
- **AgentRegistry**: Đăng ký agent types với capabilities
- **MessageBus**: Inter-agent communication
- **TaskQueue**: Task distribution với priority
- **AgentInstance**: Running agent instances
- Task execution (notion, github, devin - commented out)

#### GoClaw Equivalent:
- ✅ GoClaw có agent teams với shared task board
- ✅ GoClaw có delegation (sync/async)
- ✅ GoClaw có team mailbox (inter-agent messaging)
- ✅ GoClaw có lane-based scheduler

#### Comparison:
| Feature | Plugin | GoClaw |
|---------|--------|--------|
| Agent Registry | ✅ Custom | ✅ Database-backed |
| Message Bus | ✅ In-memory | ✅ Team mailbox |
| Task Queue | ✅ Priority queue | ✅ Task board (Kanban) |
| Delegation | ❌ Basic | ✅ Sync/async with permissions |
| Scheduling | ❌ Simple | ✅ Lane-based (main/subagent/cron) |

#### Recommendation: **SKIP**
- ❌ GoClaw có agent teams (more advanced)
- ❌ GoClaw có delegation with permissions
- ❌ GoClaw has task board (Kanban UI)

---

### 4. Agent Loop (`internal/plugins/ticrew/agent/loop.go`)

#### Features:
- Think→Act→Observe loop
- Provider integration
- Tool execution

#### GoClaw Equivalent:
- ✅ GoClaw có `internal/agent/` với full agent loop
- ✅ GoClaw có extended thinking
- ✅ GoClaw has auto-summarization at >75% context

#### Recommendation: **SKIP**
- ❌ GoClaw agent loop is production-tested

---

### 5. Scheduler (`internal/plugins/ticrew/scheduler/lanes.go`)

#### Features:
- Lane-based scheduling
- Concurrency control

#### GoClaw Equivalent:
- ✅ GoClaw có `internal/scheduler/` với lane-based concurrency
- ✅ GoClaw supports main/subagent/cron lanes

#### Recommendation: **SKIP**
- ❌ GoClaw scheduler is more advanced

---

### 6. Skills (`internal/plugins/ticrew/skills/skill.go`)

#### Features:
- SKILL.md loader
- Skill search

#### GoClaw Equivalent:
- ✅ GoClaw có `internal/skills/` với BM25 + pgvector hybrid search
- ✅ GoClaw supports SKILL.md

#### Recommendation: **SKIP**
- ❌ GoClaw skills system is more advanced

---

### 7. Notion Agents (`internal/plugins/ticrew/agents/notion/`)

#### Features:
- Notion task management agent
- Simple agent implementation

#### GoClaw Equivalent:
- ❌ GoClaw KHÔNG có Notion-specific agent
- ✅ GoClaw có generic agent system (có thể add Notion tools)

#### Recommendation: **PORT AS TOOL**
- ✅ Port Notion integration as **GoClaw tool**
- ✅ Add Notion tools to GoClaw tool registry
- ❌ Bỏ agent-specific code

---

### 8. Bootstrap (`internal/plugins/ticrew/bootstrap/identity.go`)

#### Features:
- Agent identity management
- System prompts (SOUL.md, IDENTITY.md)

#### GoClaw Equivalent:
- ✅ GoClaw có `internal/bootstrap/` với system prompts
- ✅ GoClaw supports SOUL.md, IDENTITY.md, USER.md

#### Recommendation: **SKIP**
- ❌ GoClaw bootstrap is more comprehensive

---

## 🎯 Tổng Hợp Đề Xuất

### ✅ SHOULD PORT (Giá trị cao)

| File/Feature | Reason | Effort |
|-------------|--------|--------|
| **ReviewSystem** (memory/advanced.go) | Memory reflection - useful cho learning | Medium |
| **MemoryCompression** (memory/advanced.go) | Cost optimization - reduce tokens | Low |
| **Notion Tools** (agents/notion/) | Ti-specific integration - add as GoClaw tool | Medium |

### ❌ SHOULD SKIP (GoClaw đã có)

| File/Feature | Reason |
|-------------|--------|
| **Bridge** (memory/bridge.go) | GoClaw has built-in memory |
| **MCP Bridge** (mcp/bridge.go) | GoClaw has MCP integration |
| **Orchestrator** (multi/orchestrator.go) | GoClaw has agent teams + delegation |
| **Agent Loop** (agent/loop.go) | GoClaw agent loop is production-tested |
| **Scheduler** (scheduler/lanes.go) | GoClaw scheduler is more advanced |
| **Skills** (skills/skill.go) | GoClaw skills system is more advanced |
| **Bootstrap** (bootstrap/identity.go) | GoClaw bootstrap is more comprehensive |

### 📝 Implementation Plan

#### Phase 1: Port ReviewSystem (1 day)
```go
// apps/ticrew/internal/memory/review.go
// Port logic từ ticrew-plugin/memory/advanced.go
// Integrate vào GoClaw's memory system
```

#### Phase 2: Port MemoryCompression (0.5 day)
```go
// apps/ticrew/internal/memory/compression.go
// Port logic từ ticrew-plugin/memory/advanced.go
// Add to agent loop for context compression
```

#### Phase 3: Port Notion Tools (1 day)
```go
// apps/ticrew/internal/tools/notion.go
// Port Notion integration as GoClaw tool
// Register in tool registry
```

**Total Effort**: 2.5 days

---

## 💡 Final Recommendation

**✅ Xóa `apps/ticrew-plugin` sau khi port 3 features trên**

**Rationale**:
- GoClaw đã có 90% features của plugin
- Chỉ 3 features cần port (ReviewSystem, MemoryCompression, Notion Tools)
- Giảm complexity, tránh duplicate code
- Tập trung vào GoClaw production-tested codebase

Bạn muốn tôi:
1. **Port 3 features** (ReviewSystem, MemoryCompression, Notion Tools)?
2. **Xóa toàn bộ plugin** ngay?
3. **Giữ plugin để reference** (đổi tên thành ticrew-plugin-legacy)?
