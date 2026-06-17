---
description: "Tạo plan và breakdown task - kết nối Tibrain patterns và beads"
priority: P0
alwaysApply: true
---

# Planning - Lập Kế Hoạch

> **Version**: 1.0.0
> **Date**: 2026-05-04
> **Purpose**: Quy định cách tạo plan, breakdown task, và chọn execution path
> **Priority**: P0

---

## Nguyên Tắc

**Plan trước implement.** Mọi task non-trivial (≥2 files hoặc ≥2 logical steps) cần plan.

## Quy Trình Planning

### 1. Clarify Intent (Uncertainty Rule)

Nếu request mơ hồ, hỏi user trước khi plan:
- "Bạn muốn A hay B?"
- "Scope là X hay Y?"
- "Priority là correctness hay speed?"

### 2. Task Breakdown

Chia task thành beads (units of work) độc lập:

```
Bead 1: [Explore] Khám phá codebase liên quan
Bead 2: [Implement] Thay đổi core logic
Bead 3: [Implement] Update tests
Bead 4: [Validate] Chạy tests và verify
```

**Nguyên tắc chia bead:**
- Mỗi bead là 1 logical concern (1 file type hoặc 1 layer)
- Beads có thể parallel nếu độc lập
- Beads sequential nếu có dependency
- Estimate mỗi bead ≤ 100 lines of code

### 3. Select Tibrain Patterns

Sử dụng Tibrain để chọn patterns phù hợp:

```bash
# Tìm patterns cho domain
py "Z:\Ti\router\tibrain-cli.py" search <domain>

# Ví dụ: API design
py "Z:\Ti\router\tibrain-cli.py" search api
py "Z:\Ti\router\tibrain-cli.py" execute best-api-design --params "{\"task\": \"Design REST endpoint\"}"

# Ví dụ: Database
py "Z:\Ti\router\tibrain-cli.py" search database
py "Z:\Ti\router\tibrain-cli.py" execute best-postgres-patterns --params "{\"task\": \"Add migration\"}"
```

### 4. Chọn Execution Path

| Tình huống | Execution Path |
|------------|---------------|
| Single-track (1 concern, sequential) | Main agent, sequential tools |
| Parallel reads/exploration | Main agent, parallel tool calls |
| 2+ independent tracks | Subagents (2+, never 1) |
| Complex cross-domain | Main agent explore → Subagents implement |

**Router auto-selection (Tibrain):**
```bash
# Chọn provider/model phù hợp task type
py "Z:\Ti\router\tibrain-cli.py" execute router.select_auto --params "{\"task_type\": \"plan\"}"
```

### 5. Define Acceptance Criteria

Mỗi bead cần acceptance criteria rõ ràng:
- [ ] Tests pass
- [ ] Build succeeds
- [ ] Behavior đúng như spec
- [ ] No regression

## Plan Format

```markdown
## Plan: [Task Title]

### Context
[Summary của findings từ exploration]

### Beads
1. **[Bead ID] [Title]** (Priority: P0/P1)
   - Objective: [Mục tiêu cụ thể]
   - Files: [Files liên quan]
   - Acceptance: [Criteria]
   - Dependencies: [Beads phụ thuộc]

2. **[Bead ID] [Title]** ...

### Patterns Used
- [Tibrain pattern name]

### Risks
- [Rủi ro và mitigation]
```

## Tibrain Skills cho Planning

| Skill | Khi nào dùng |
|-------|-------------|
| `best-strategic-compact` | Plan ngắn gọn, không over-engineer |
| `best-context-budget` | Quản lý context window khi plan dài |
| `best-autonomous-loops` | Agent workflow cho complex tasks |
| `best-dmux-workflows` | Decompose và multiplex sub-tasks |

## Anti-Patterns

- ❌ **Plan trong đầu** — Không document, dễ quên scope
- ❌ **Over-engineer** — Plan cho requirements không tồn tại
- ❌ **1 bead khổng lồ** — Không chia nhỏ, dễ fail giữa chừng
- ❌ **Không define acceptance** — Không biết khi nào xong
- ❌ **Không check Tibrain** — Reinvent patterns đã có

## Kết Nối Workflow

1. **Explore** (`exploration.md`) → Thu thập context
2. **Plan** (`planning.md`) → Breakdown + chọn patterns
3. **Implement** (`implementation.md`) → Execute beads
4. **Review** (`review.md`) → Verify acceptance criteria
