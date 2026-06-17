# Tổng Hợp Kiến Trúc Router & Agent Từ GitHub Research

> **Ngày**: 2026-04-29  
> **Người thực hiện**: claude  
> **Mục đích**: Nghiên cứu các repo tương tự trên GitHub để học hỏi và áp dụng cho dự án Ti

---

## 1. pLLM - Go LLM Gateway (⭐ 1k+ stars)

**Repo**: `github.com/andreimerfu/pllm`  
**Ngôn ngữ**: Go  
**Framework**: Echo  
**Tương đồng**: ⭐⭐⭐⭐⭐ (Rất giống Ti Router)

### Kiến trúc

```
User Request → pLLM Router → Dynamic Provider Selection → LLM Provider
                    ↓
            Configuration (YAML)
                    ↓
            Adaptive Routing Algorithm
```

### Tính năng chính

| Tính năng | Mô tả | Áp dụng cho Ti |
|-----------|-------|---------------|
| **OpenAI Compatible API** | `/v1/chat/completions`, `/v1/models` | ✅ Ti đã có |
| **Multi-Provider** | Google, OpenAI, Cohere, Anthropic, Mistral, Azure | ✅ Ti đã có |
| **Dynamic Routing** | Least-latency, cost-optimized, fallback, load-balancing | ⚠️ Ti cần cải thiện |
| **Streaming** | SSE streaming support | ✅ Ti đã có |
| **Non-streaming** | Standard HTTP response | ✅ Ti đã có |
| **Middleware** | Custom headers, retries, timeouts | ⚠️ Cần thêm |
| **YAML Config** | `config.yaml` cho providers | ✅ Ti đã có |
| **Docker** | Containerization support | ❌ Ti chưa có |

### Routing Algorithm của pLLM

```go
// 1. Least Latency - Chọn provider phản hồi nhanh nhất
// 2. Cost Optimized - Chọn provider rẻ nhất cho model
// 3. Fallback - Tự động chuyển khi provider fail
// 4. Load Balancing - Phân phối đều requests
```

**Bài học cho Ti**:
- Cần thêm **Adaptive Routing** thay vì chỉ fallback đơn giản
- Thêm **Latency Tracking** để đo thời gian phản hồi
- Thêm **Cost Tracking** để tối ưu chi phí
- Cân nhắc **Weighted Routing** kết hợp nhiều yếu tố

---

## 2. Agent Orchestrator - Parallel AI Agents (⭐ 3k+ stars)

**Repo**: `github.com/ComposioHQ/agent-orchestrator`  
**Ngôn ngữ**: TypeScript/Node.js  
**Tương đồng**: ⭐⭐⭐⭐ (Orchestration pattern)

### Kiến trúc

```
Dashboard (Web UI)
    ↓
Orchestrator Agent (Master)
    ↓
Worker Agents (Parallel)
    ↓
Git Worktrees (Isolated)
```

### Tính năng chính

| Tính năng | Mô tả | Áp dụng cho Ti |
|-----------|-------|---------------|
| **Parallel Agents** | Nhiều agent chạy song song | ⚠️ Ti chưa có |
| **Git Worktree Isolation** | Mỗi agent có workspace riêng | ⚠️ Có thể áp dụng |
| **Auto CI Fix** | Tự động đọc CI logs và sửa lỗi | ⚠️ Có thể thêm |
| **Auto Review** | Tự động xử lý review comments | ⚠️ Có thể thêm |
| **Reactions** | Trigger actions dựa trên events | ⚠️ Ti chưa có |
| **Plugin System** | 7 plugin slots | ⚠️ Có thể thêm |
| **Dashboard** | Web UI để giám sát | ❌ Ti chưa có |

### Configuration Pattern

```yaml
# agent-orchestrator.yaml
port: 3000
defaults:
  runtime: tmux
  agent: claude-code
  workspace: worktree
  notifiers: [desktop]

projects:
  my-app:
    repo: owner/my-app
    path: ~/my-app
    defaultBranch: main
    
reactions:
  ci-failed:
    auto: true
    action: send-to-agent
    retries: 2
  changes-requested:
    auto: true
    action: send-to-agent
    escalateAfter: 30m
```

**Bài học cho Ti**:
- Thêm **Event-driven Reactions** cho router (CI fail → auto-fix)
- Thêm **Plugin Architecture** cho CLI để mở rộng
- Cân nhắc **Worktree Isolation** cho multi-agent
- Thêm **Dashboard/API** để giám sát agents

---

## 3. MCP Go SDK - Model Context Protocol (⭐ 500+ stars)

**Repo**: `github.com/mark3labs/mcp-go`  
**Ngôn ngữ**: Go  
**Tương đồng**: ⭐⭐⭐⭐⭐ (MCP Integration)

### Kiến trúc

```
MCP Server (Go)
    ├── Tools (function calling)
    ├── Prompts (templates)
    ├── Resources (data access)
    ├── Roots (workspace scope)
    ├── Sampling (LLM requests)
    └── Completion (auto-complete)
```

### Transport Types

| Transport | Use Case | Ti Status |
|-----------|----------|-----------|
| **stdio** | Local CLI tools | ⚠️ Cần thêm |
| **HTTP/SSE** | Remote servers | ✅ Ti đã có (router) |
| **WebSocket** | Real-time | ❌ Chưa có |

### Tool Registration Pattern

```go
// Đăng ký tool với MCP server
server.AddTool("search_code", func(args map[string]interface{}) (string, error) {
    // Implementation
})
```

**Bài học cho Ti**:
- Thêm **stdio transport** cho MCP để CLI tools có thể gọi
- Thêm **Tool Registry** pattern cho brain/mcp
- Thêm **Prompt Templates** cho reusable prompts

---

## 4. Awesome AI Agents Frameworks (⭐ 11k+ stars)

**Repo**: `github.com/mb-mal/awesome-ai-agents-frameworks`  
**Tương đồng**: ⭐⭐⭐ (Reference)

### Patterns phổ biến

| Pattern | Mô tả | Áp dụng |
|---------|-------|---------|
| **ReAct** | Reasoning + Acting | ✅ Ti đã có trong .windsurfrules |
| **Multi-Agent** | Nhiều agent phối hợp | ⚠️ Cần thêm |
| **RAG** | Retrieval Augmented Generation | ⚠️ Brain cần cải thiện |
| **Self-Evolving** | Tự động cải thiện | ❌ Chưa có |
| **Hierarchical** | Agent cha-con | ⚠️ Có thể áp dụng |

---

## 5. Beads - Task Tracking System (⭐ 100+ stars)

**Repo**: `github.com/castle-x/skills-x/beads`  
**Tương đồng**: ⭐⭐⭐⭐⭐ (Direct match với Ti beads)

### Tính năng

- **Sub-agents**: Chia task cho nhiều agent con
- **Scheduling**: Lên lịch tự động
- **Planning**: Tạo plan tự động
- **CI/CD Integration**: Kết nối với CI pipeline
- **Test Automation**: Tự động chạy tests
- **Logging**: Centralized logging

**Bài học cho Ti**:
- Đã có beads.md → Cần thêm **automation** cho beads
- Thêm **Sub-agent dispatch** trong router
- Thêm **Scheduling** cho recurring tasks

---

## So Sánh Kiến Trúc Hiện Tại vs Best Practices

### Ti Router hiện tại

```
User → HTTP API → Router → Provider
              ↓
         Config (YAML)
              ↓
         Auth (API Key/OAuth)
```

### Kiến trúc lý tưởng (tổng hợp)

```
User → HTTP API → Router Engine → Adaptive Routing → Provider
              ↓                      ↓
         Auth & Rate Limit    Latency/Cost Tracking
              ↓                      ↓
         Config (YAML)        Fallback & Retry
              ↓                      ↓
         Brain/Memory         Metrics & Monitoring
              ↓                      ↓
         MCP Integration      Event-driven Reactions
```

---

## Đề Xuất Cải Thiện cho Ti

### P0 - Critical

1. **Adaptive Routing Engine**
   - Thêm latency tracking
   - Thêm cost tracking
   - Weighted routing algorithm
   - Priority: HIGH

2. **Event-driven Reactions**
   - CI fail → auto-fix
   - Build error → auto-diagnose
   - Priority: HIGH

### P1 - Important

3. **Plugin Architecture cho CLI**
   - 7 plugin slots như Agent Orchestrator
   - Allow custom commands
   - Priority: MEDIUM

4. **MCP stdio Transport**
   - CLI tools gọi MCP server qua stdio
   - Priority: MEDIUM

5. **Dashboard/API Monitoring**
   - Web UI giám sát router
   - Real-time metrics
   - Priority: MEDIUM

### P2 - Nice to Have

6. **Docker Support**
   - Containerize router
   - Priority: LOW

7. **Multi-Agent Worktree Isolation**
   - Mỗi agent có workspace riêng
   - Priority: LOW

---

## References

- [pLLM](https://github.com/andreimerfu/pllm)
- [Agent Orchestrator](https://github.com/ComposioHQ/agent-orchestrator)
- [MCP Go SDK](https://github.com/mark3labs/mcp-go)
- [Awesome AI Agents](https://github.com/mb-mal/awesome-ai-agents-frameworks)
- [Beads](https://github.com/castle-x/skills-x/beads)
