# Agent Team - Key Concepts (từ video Claude Code)

## 📖 Tổng Quan
Agent Team = nhiều AI agents làm việc cùng nhau với **session riêng biệt**, được điều phối bởi **Captain Agent**.

## 🏗️ Kiến Trúc Agent Team (từ video)

### 1. Single Agent (Hiện tại - Ticlaw)
```
User → TaskAgent → [Tools] → Output
```
- ✅ Đơn giản, dễ debug
- ❌ Giới hạn ở 1 context window
- ❌ Dễ "ảo giác" khi data quá lớn

### 2. Subagent (Đã có trong Claude Code)
```
User → Main Agent → Subagent(task) → Tóm tắt → Output
```
- ✅ Bảo tồn token của main session
- ❌ Subagent không có session riêng
- ❌ Chỉ báo cáo kết quả, không maintain context

### 3. Agent Team (Mục tiêu - từ video)
```
User → Captain Agent
         ├─ Worker A (session riêng) → focus: left side of codebase
         ├─ Worker B (session riêng) → focus: right side of codebase
         └─ Worker C (session riêng) → focus: top-down view
```
- ✅ Mỗi agent có session riêng → không bị phân tâm
- ✅ Parallel execution với context isolation
- ✅ Captain lập kế hoạch, phân task, tổng hợp kết quả

## 🔄 Coordination Patterns (từ video)

### Pattern 1: Captain-Worker (Chính)
```
User → Captain
         ├─ Plan: chia task thành subtasks
         ├─ Assign: giao việc cho Workers
         ├─ Lock: khóa task tránh race condition
         ├─ Monitor: theo dõi progress
         └─ Merge: tổng hợp kết quả
```

### Pattern 2: Race Condition Prevention
```
Task Pool:
  [Task A] → LOCKED by Worker 1
  [Task B] → AVAILABLE
  [Task C] → LOCKED by Worker 3
```
- Khi worker bắt đầu task → lock ngay
- Tránh 2 workers làm cùng 1 task → lãng phí token

### Pattern 3: Memory Files (Shared Context)
```
shared-memory.md
├── Current context
├── Known issues
├── Solutions found
└── Next steps
```
- Agents không share history → dùng `.md` files để chia sẻ context
- Captain ghi plan vào file → Workers đọc và thực hiện
- Workers ghi kết quả → Captain đọc và tổng hợp

### Pattern 4: Skill Creation
```
Workflow → Captain observes → Extracts pattern → Creates Skill
Skill → Reusable via `/command` → Future tasks
```
- Sau khi hoàn thành workflow → biến thành Skill
- Tái sử dụng nhanh cho tasks tương tự

## 🎯 Áp Dụng Vào Ticlaw

### Hiện tại đã có:
| Component | Status | Gap |
|-----------|--------|-----|
| TaskAgent (single) | ✅ Done | Cần upgrade thành Captain |
| Test Runner tool | ✅ Done | ✅ Ready cho Workers |
| Git tool | ✅ Done | ✅ Ready cho Workers |
| Pre-commit review | ✅ Done | ✅ Ready cho Reviewer Worker |
| Autonomous loop | ✅ Done | Cần upgrade thành Captain loop |
| Session store | ✅ In-memory | Cần per-agent sessions |

### Cần thêm để thành Agent Team:
| Component | Priority | Effort | Video Pattern |
|-----------|----------|--------|---------------|
| Captain Agent | P0 | Medium | Team Lead pattern |
| Worker Agents | P0 | Medium | Session isolation |
| Task Locking | P0 | Low | Race condition check |
| Shared Memory Files | P1 | Low | Memory files pattern |
| Skill Creation | P1 | Medium | Skill extraction |
| Cost Optimization | P2 | Low | Mixed models |
| Dashboard Monitoring | P2 | Medium | TMUX-like view |

## 📊 So Sánh

| Feature | Single Agent | Subagent | Agent Team |
|---------|-------------|----------|------------|
| Context limit | 1 window | 1 main + 1 sub | N windows (isolated) |
| Parallel | ❌ | ❌ | ✅ |
| Session isolation | ❌ | Partial | ✅ Full |
| Race condition | N/A | N/A | ✅ Locked |
| Shared context | N/A | Summary | Memory files |
| Cost | Low | Medium | Optimizable |
| Complexity | Low | Medium | High |

## 🔗 References
- Video: https://www.youtube.com/watch?v=fWUgYTD3Jvs
- Claude Code Agent Team docs
- OpenAI Swarm patterns
- AutoGen multi-agent framework
- Microsoft TaskWeaver
