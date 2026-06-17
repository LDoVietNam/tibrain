---
description: "Triển khai code theo slices - kết nối Tibrain patterns và incremental workflow"
priority: P0
alwaysApply: true
---

# Implementation - Triển Khai

> **Version**: 1.0.0
> **Date**: 2026-05-04
> **Purpose**: Quy định cách implement code theo incremental slices, reuse patterns
> **Priority**: P0

---

## Nguyên Tắc

**Implement theo slices — mỗi slice là 1 increment hoàn chỉnh, testable, compilable.**

## The Increment Cycle

```
Implement ──→ Test ──→ Verify ──→ Commit ──→ Next Slice
     ▲                                     │
     └─────────────────────────────────────┘
```

### 1. Implement

**Simplicity First:** Trước khi code, hỏi: "What is the simplest thing that could work?"

- Implement naive, obviously-correct version trước
- Chỉ optimize sau khi correctness proven by tests
- Mỗi slice ≤ 100 lines (ngưỡng cảnh báo)

**Scope Discipline:**
- Chỉ touch những gì task yêu cầu
- Không "clean up" code adjacent (note lại, đừng sửa)
- Không add features không trong spec

**Reuse Patterns:**
```bash
# Lấy patterns phù hợp từ Tibrain
py "Z:\Ti\router\tibrain-cli.py" search <language>-patterns
py "Z:\Ti\router\tibrain-cli.py" execute best-python-patterns --params "{\"task\": \"Implement API endpoint\"}"
```

### 2. Test

- Viết/run tests cho slice hiện tại
- Đảm bảo existing tests vẫn pass
- Coverage cho new code paths

### 3. Verify

- Build/compile thành công
- Manual check nếu cần
- No new warnings/errors

### 4. Commit

- Commit message rõ ràng: `scope: what changed and why`
- Mỗi commit = 1 logical change
- Không mix concerns trong 1 commit

### 5. Next Slice

Carry forward, không restart.

## Implementation Rules

| Rule | Mô tả |
|------|-------|
| **One Thing at a Time** | Mỗi slice chỉ thay đổi 1 concern |
| **Keep It Compilable** | Sau mỗi slice, project build và tests pass |
| **Safe Defaults** | New code default conservative (disabled, opt-in) |
| **Rollback-Friendly** | Mỗi slice có thể revert độc lập |
| **Feature Flags** | Incomplete features flag để merge safely |

## Code Quality Checks

```
SIMPLICITY CHECK:
✗ Generic EventBus with middleware pipeline for one notification
✓ Simple function call

✗ Abstract factory pattern for two similar components
✓ Two straightforward components

✗ Config-driven form builder for three forms
✓ Three form components
```

## Tibrain Integration

### During Implementation

```bash
# Get patterns for current language
py "Z:\Ti\router\tibrain-cli.py" execute best-<language>-patterns --params "{\"task\": \"<specific task>\"}"

# Get testing patterns
py "Z:\Ti\router\tibrain-cli.py" execute best-<language>-testing --params "{\"task\": \"Write unit tests\"}"

# Check performance patterns nếu cần
py "Z:\Ti\router\tibrain-cli.py" execute best-strategic-compact --params "{\"task\": \"Optimize hot path\"}"
```

### Update Beads Status

```bash
# Mark bead in_progress khi bắt đầu slice
py "Z:\Ti\router\tibrain-cli.py" execute beads.update --params "{\"id\": \"<bead-id>\", \"status\": \"in_progress\", \"comment\": \"Implementing slice N\"}"

# Mark bead ready khi slice hoàn thành
py "Z:\Ti\router\tibrain-cli.py" execute beads.update --params "{\"id\": \"<bead-id>\", \"status\": \"ready\", \"comment\": \"Slice N complete, tests pass\"}"
```

## Slicing Strategies

| Strategy | Khi nào dùng | Ví dụ |
|----------|-------------|-------|
| **Vertical Slice** (Preferred) | Feature mới | API + handler + DB + tests trong 1 slice |
| **Contract-First** | Interface/API changes | Định nghĩa interface → Implement consumer |
| **Risk-First** | Complex/uncertain tasks | Làm phần rủi ro cao nhất trước |

## Anti-Patterns

- ❌ **Implement cả feature trong 1 pass** — Không testable, dễ break
- ❌ **Refactor + feature cùng lúc** — Khó debug khi fail
- ❌ **Bỏ qua existing patterns** — Reinvent trong codebase
- ❌ **Không test giữa chừng** — Phát hiện bug muộn, cost cao
- ❌ **Commit không descriptive** — "fix stuff", "update"

## Tibrain Skills

| Skill | Purpose |
|-------|---------|
| `best-<language>-patterns` | Language-specific patterns |
| `best-tdd-workflow` | Test-first implementation |
| `best-strategic-compact` | Avoid over-engineering |
| `best-context-budget` | Manage context during long sessions |
