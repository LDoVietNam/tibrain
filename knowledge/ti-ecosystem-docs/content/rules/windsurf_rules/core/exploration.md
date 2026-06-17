---
description: "Tìm hiểu codebase context trước khi hành động - kết nối Tibrain skills"
priority: P0
alwaysApply: true
---

# Exploration - Khám Phá Context

> **Version**: 1.0.0
> **Date**: 2026-05-04
> **Purpose**: Quy định cách explore codebase để thu thập context đủ cho quyết định
> **Priority**: P0

---

## Nguyên Tắc

**Không hành động trước khi hiểu.** Explore proportional to risk — thu thập context tỷ lệ thuận với độ phức tạp của task.

## Quy Trình Explore

### 1. Xác Định Entry Points (2-3 files)

```
Z:\10_WORKPLACE\Ti\docs\README.md                 →  Documentation Hub
README.md / package.json / go.mod / Cargo.toml  →  Architecture
main.go / index.ts / app.py                      →  Entry points
cmd/ / internal/ / src/                          →  Module structure
```

**Hành động:**
- Đọc `README.md` để hiểu architecture
- Đọc `go.mod` / `package.json` để hiểu dependencies và module boundaries
- Liệt kê top-level directories (`apps/`, `packages/`, `internal/`, `cmd/`)

### 2. Trace Execution Path

Từ entry point, trace đến code cần thay đổi:
- Search function/struct names liên quan task
- Đọc call sites (nơi gọi) trước implementation (code thực thi)
- Xác định interfaces/contracts giữa các module

### 3. Check Tibrain Skills

Sử dụng Tibrain CLI để lấy patterns phù hợp:

```bash
# Tìm skills liên quan codebase scan
py "Z:\Ti\router\tibrain-cli.py" search repo-scan

# Tìm skills về context management
py "Z:\Ti\router\tibrain-cli.py" search context-budget

# Execute skill để lấy best practices
py "Z:\Ti\router\tibrain-cli.py" execute best-repo-scan --params "{\"task\": \"Explore monorepo structure\"}"
```

### 4. Xác Định Scope & Rủi Ro

| Mức độ | Context cần thu thập |
|--------|---------------------|
| **Trivial** (1 file, 1 hàm) | Đọc file target + 1-2 adjacent files |
| **Small** (1 module) | Đọc module + interface contracts + tests |
| **Medium** (cross-module) | Trace execution path + data flow + config |
| **Large** (architecture) | Read architecture docs + key abstractions + all affected modules |

### 5. Khi Nào Dừng Explore

Dừng khi có thể trả lời:
- [ ] Code cần thay đổi nằm ở đâu?
- [ ] Có interfaces/contracts nào bị ảnh hưởng?
- [ ] Tests nào cần update?
- [ ] Có patterns/conventions nào cần follow?
- [ ] Có dependencies mới cần thêm?

## Tools Sử Dụng

| Tool | Khi nào dùng |
|------|-------------|
| `list_dir` | Khám phá cấu trúc thư mục |
| `read_file` | Đọc file cụ thể (dùng `limit` cho file lớn) |
| `grep_search` | Tìm function/struct/type references |
| `code_search` | Exploration phức tạp, cần hiểu relationships |
| Tibrain CLI | Lấy best practices, patterns cho domain |

## Anti-Patterns

- ❌ **Không đọc README** — "Tôi biết cấu trúc này"
- ❌ **Giả định conventions** — Không verify coding style, naming conventions
- ❌ **Bỏ qua test files** — Tests là documentation sống động nhất
- ❌ **Explore quá sâu** — Dành 80% thời gian đọc code, 20% implement
- ❌ **Không log findings** — Quên những gì đã discover

## Kết Nối Tibrain

- **Tibrain skill**: `best-repo-scan` — Quy trình scan repo hiệu quả
- **Tibrain skill**: `best-context-budget` — Quản lý context window
- **Bead log**: Sau khi explore, log bead với findings summary

## Examples

- ❌ `User: "Fix bug X" → Sửa ngay file đầu tiên tìm thấy`
- ✅ `User: "Fix bug X" → Search references → Đọc entry point → Trace data flow → Xác định root cause → Fix`
