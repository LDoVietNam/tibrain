# Agent Team - Implementation Plan cho Ticlaw
*(Dựa trên video: Làm chủ Agent Team trong Claude Code)*

## 🎯 Mục Tiêu
Nâng cấp Ticlaw từ Single Agent → Agent Team để:
- Mỗi agent có session riêng → không bị "ảo giác" khi context lớn
- Captain phân task, Workers thực hiện song song
- Task locking tránh race condition
- Shared memory files để agents trao đổi context

## 📋 Roadmap

### Phase 1: Captain Agent + Registry (Bead 1-4)
**Mục tiêu:** Captain nhận task, lập kế hoạch, phân việc

| Bead | Task | Test |
|------|------|------|
| 1 | Create `internal/agent/captain.go` - Captain Agent | Unit tests |
| 2 | Create `internal/agent/registry.go` - Worker registry | Unit tests |
| 3 | Create `internal/agent/worker.go` - Worker Agent types | Unit tests |
| 4 | Create `internal/agent/task_lock.go` - Race condition prevention | Unit tests |

**Files cần tạo:**
- `internal/agent/captain.go`
- `internal/agent/registry.go`
- `internal/agent/worker.go`
- `internal/agent/task_lock.go`
- `internal/agent/captain_test.go`
- `internal/agent/registry_test.go`
- `internal/agent/task_lock_test.go`

### Phase 2: Shared Memory Files (Bead 5-7)
**Mục tiêu:** Agents chia sẻ context qua .md files

| Bead | Task | Test |
|------|------|------|
| 5 | Create `internal/agent/memory_file.go` - Shared memory | Unit tests |
| 6 | Captain writes plan → Workers read | Integration tests |
| 7 | Workers write results → Captain reads | Integration tests |

**Files cần tạo:**
- `internal/agent/memory_file.go`
- `internal/agent/memory_file_test.go`

### Phase 3: Worker Execution (Bead 8-10)
**Mục tiêu:** Workers thực hiện task với session riêng

| Bead | Task | Test |
|------|------|------|
| 8 | Worker session isolation | Unit tests |
| 9 | Parallel worker execution | Integration tests |
| 10 | Result aggregation | Unit tests |

**Files cần tạo:**
- `internal/agent/worker_executor.go`
- `internal/agent/worker_executor_test.go`

### Phase 4: Skill Creation (Bead 11-12)
**Mục tiêu:** Biến workflow thành reusable skills

| Bead | Task | Test |
|------|------|------|
| 11 | Workflow observation & extraction | Unit tests |
| 12 | Skill registration & reuse | Unit tests |

**Files cần tạo:**
- `internal/agent/skill_extractor.go`
- `internal/agent/skill_registry.go`
- `internal/agent/skill_extractor_test.go`

## 📊 Timeline Ước Lượng

| Phase | Beads | Thời gian |
|-------|-------|-----------|
| 1. Captain + Registry + Locking | 4 | 2-3 ngày |
| 2. Shared Memory Files | 3 | 1-2 ngày |
| 3. Worker Execution | 3 | 2-3 ngày |
| 4. Skill Creation | 2 | 1-2 ngày |
| **Total** | **12 beads** | **6-10 ngày** |

## 🔗 Dependencies

```
Phase 1 (Captain + Registry + Locking)
    ↓
Phase 2 (Shared Memory Files) ← cần Captain
    ↓
Phase 3 (Worker Execution) ← cần Registry + Memory Files
    ↓
Phase 4 (Skill Creation) ← cần Worker Execution
```

## ✅ Definition of Done (per Bead)
- [ ] Code implemented
- [ ] Sub-agent viết tests
- [ ] Tests PASS (100%)
- [ ] `go build ./cmd` PASS
- [ ] CHANGELOG.md updated
- [ ] Taskboard updated

## 📝 Notes
- Tuân thủ Beads Pattern nghiêm ngặt
- Mỗi bead = 1 commit
- Không code bead tiếp nếu bead trước chưa PASS tests
