---
description: "Log beads/checkpoints trong Tibrain graph - track decisions và progress"
priority: P0
alwaysApply: true
---

# Beads - Log Checkpoints & Decisions

> **Version**: 1.0.0
> **Date**: 2026-05-04
> **Purpose**: Quy định cách log beads trong Tibrain để track decisions, progress, và handoff
> **Priority**: P0

---

## Beads Là Gì

Bead = 1 unit of work trong Tibrain graph. Beads dùng Dolt-powered database với versioning và cell-level merge.

### Cấu Trúc Bead

Mỗi bead có:
- **ID**: Unique identifier (hash-based, tránh conflict)
- **Title**: Mô tả ngắn
- **Objective**: Mục tiêu cụ thể
- **Status**: `open` → `ready` → `in_progress` → `review` → `closed`
- **Priority**: P0/P1/P2
- **Depends_on**: Beads phụ thuộc
- **Files**: Files liên quan
- **Acceptance**: Criteria hoàn thành
- **Route_hints**: Gợi ý provider/model
- **Commands**: Commands để verify

### Hierarchical IDs (Epics)

Beads hỗ trợ hierarchical IDs cho epics:
- `bd-a3f8` — Epic (task lớn)
- `bd-a3f8.1` — Task trong epic
- `bd-a3f8.1.1` — Sub-task

### Graph Links

Beads có thể liên kết với nhau:
- `relates_to` — Liên quan
- `duplicates` — Trùng lặp
- `supersedes` — Thay thế
- `replies_to` — Phản hồi

## Khi Nào Log Beads

### Bắt Buộc Log
- [ ] Khi bắt đầu task mới (tạo bead cho task)
- [ ] Khi chuyển sang bead mới trong plan
- [ ] Khi phát hiện issue/blocker
- [ ] Khi hoàn thành 1 bead
- [ ] Khi thay đổi quyết định quan trọng (change approach, pivot)

### Nên Log
- [ ] Khi hoàn thành 1 slice trong implementation
- [ ] Khi phát hiện pattern/gotcha đáng nhớ
- [ ] Khi dùng Tibrain skill và kết quả tốt/xấu

## Status Workflow

```
open      → Ready to start
  ↓
ready     → Dependencies satisfied, có thể bắt đầu
  ↓
in_progress → Đang thực hiện
  ↓
review    → Hoàn thành, chờ review
  ↓
closed    → Review xong, accepted

blocked   → Bị chặn, cần help/dependency
  → (sau khi unblocked) → ready
deferred  → Tạm hoãn, không làm ngay
  → (sau khi re-prioritized) → open
```

## Commands

### Liệt Kê Beads Sẵn Sàng

```bash
# Lấy danh sách beads ready (unblocked)
py "Z:\Ti\router\tibrain-cli.py" execute beads.ready

# Lấy với limit
py "Z:\Ti\router\tibrain-cli.py" execute beads.ready --params "{\"limit\": 5}"
```

### Xem Chi Tiết Bead

```bash
py "Z:\Ti\router\tibrain-cli.py" execute beads.get --params "{\"id\": \"<bead-id>\"}"
```

### Cập Nhật Status

```bash
# Mark in_progress
py "Z:\Ti\router\tibrain-cli.py" execute beads.update --params "{\"id\": \"<bead-id>\", \"status\": \"in_progress\", \"comment\": \"Starting exploration phase\"}"

# Mark ready (hoàn thành, chờ next step)
py "Z:\Ti\router\tibrain-cli.py" execute beads.update --params "{\"id\": \"<bead-id>\", \"status\": \"ready\", \"comment\": \"Exploration complete. Found 3 affected files.\"}"

# Mark review
py "Z:\Ti\router\tibrain-cli.py" execute beads.update --params "{\"id\": \"<bead-id>\", \"status\": \"review\", \"comment\": \"Implementation complete. Self-review passed.\"}"

# Mark closed
py "Z:\Ti\router\tibrain-cli.py" execute beads.update --params "{\"id\": \"<bead-id>\", \"status\": \"closed\", \"comment\": \"Merged. Tests pass.\"}"

# Mark blocked
py "Z:\Ti\router\tibrain-cli.py" execute beads.update --params "{\"id\": \"<bead-id>\", \"status\": \"blocked\", \"comment\": \"Blocked: need clarification on API contract\"}"
```

## Comment Format

Comment nên concise, factual, actionable:

```
✓ "Slice 2/3 complete: handler + tests added, 1 edge case remaining"
✓ "Pivoted from Approach A to B: B handles concurrent requests better"
✗ "Done" (không informative)
✗ "Fixed stuff" (không specific)
```

## Handoff Between Agents / Sessions

Khi chuyển agent hoặc session mới:

1. **Log current state** — Cập nhật tất cả beads đang `in_progress`
2. **List ready beads** — `beads.ready` để biết tiếp theo làm gì
3. **Read bead details** — `beads.get` để hiểu context
4. **Continue** — Pick bead và mark `in_progress`

## Anti-Patterns

- ❌ **Không log beads** — Mất track progress, không handoff được
- ❌ **Status không khớp thực tế** — `in_progress` nhưng đã xong từ lâu
- ❌ **Comment vague** — Không giúp agent tiếp theo hiểu context
- ❌ **Quên mark blocked** — Agent khác không biết đang chờ gì
- ❌ **Không update sau pivot** — Decision history bị mất

## Tích Hợp Với Workflow

| Phase | Bead Action |
|-------|-------------|
| **Explore** | Log bead → mark `in_progress` → explore → update findings |
| **Plan** | Tạo beads cho từng task → set dependencies |
| **Implement** | Mark bead `in_progress` → implement slice → update progress |
| **Review** | Mark bead `review` → run checklist → mark `closed` hoặc `open` |
| **Block** | Mark `blocked` + specific reason → hỏi user/agent khác |

## Examples

- ❌ `Bắt đầu task → implement 3 giờ → báo xong`
- ✅ `Bắt đầu task → tạo beads → mark in_progress → log sau mỗi slice → mark ready/review/closed`

## Tibrain Skills

| Skill | Purpose |
|-------|---------|
| `best-context-budget` | Quản lý beads cho context window |
| `best-strategic-compact` | Giữ beads concise, actionable |
