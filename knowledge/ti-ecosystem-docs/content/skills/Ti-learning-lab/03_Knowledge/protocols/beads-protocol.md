# Beads Logging Protocol

> **Scope**: All projects under `Z:\` workspace  
> **Version**: 1.0  
> **Last Updated**: 2026-04-27

---

## Nguyên tắc

Mỗi **task** = 1 entry beads. Không gộp nhiều task vào 1 entry.

---

## Format Entry

```powershell
# PowerShell template
$timestamp = Get-Date -Format "yyyy-MM-ddTHH:mm:ss.fffZ"
$entry = "## [$timestamp] <Task Name> - <agent> - <status>`nActions: <actions> | Issues: <issues> | Lessons: <lessons> | Verification: <verification> | Files: <files>"
```

```markdown
## [2026-04-27T12:16:00.123Z] Fix Import Alias in openai/handlers.go - claude - DONE
Actions: Changed import path, replaced all provider. → providers. | Issues: E001: undefined: provider → Added import alias | Lessons: Package name must match import path | Verification: go build passes, go test passes | Files: layers/http/openai/handlers.go
```

---

## Quy tắc Chi tiết

### 1. Một Entry = Một Task Duy Nhất

- **Sai**: `[12:15] Session — Health Check + Auto-Fix + Refactor` (gộp 3 task)
- **Đúng**: 3 entry riêng biệt:
  - `[12:15] Health Check All Packages`
  - `[12:16] Fix Import Alias in openai/handlers.go`
  - `[12:17] Refactor ChatStream Error Handling`

### 2. Task Name Ngắn Gọn, Không Chữ "Session"

| Sai | Đúng |
|---|---|
| `[12:15] Session — P1 Wire Complete` | `[12:15] Wire ChunkTranslator into routerd` |
| `[12:15] Afternoon Work — Bug Fixes` | `[12:15] Fix provider import mismatch` |

### 3. Actions Performed Liệt Kê Từng Bước

Phải ghi rõ:
- File nào bị sửa
- Hàm/struct nào đổi
- Dòng code quan trọng (nếu cần)

```markdown
**Actions Performed:**
- `layers/http/openai/handlers.go:15`: Changed import `".../provider"` → `providers ".../provider"`
- `layers/http/openai/handlers.go:50-200`: Replaced all `provider.` → `providers.`
```

### 4. Issues Fixed Mapping Rõ Ràng

```markdown
**Issues Fixed:**
- E001: `undefined: provider` in `layers/http/openai/handlers.go:15`
  → Root cause: Package declared as `providers` but imported as `provider`
  → Fix: Added import alias `providers "github.com/ti/router/layers/provider"`
```

### 5. Verification Chỉ Điền Sau Khi Test Xong

```markdown
**Verification:**
- [x] `go test ./...` passes in `layers/translator`
- [x] `go build .` passes in `layers/http/openai`
- [ ] End-to-end curl test pending (marked pending, not checked)
```

---

## Error Code Reference

| Code | Meaning | Example |
|---|---|---|
| E001 | Import/alias mismatch | `undefined: provider` |
| E002 | Assignment mismatch | `1 variable but returns 2 values` |
| E003 | Type mismatch | `cannot use X as Y` |
| E004 | Missing method in interface | `*T does not implement I` |
| E005 | Package name conflict | `http` shadows `net/http` |
| E006 | Shadowing variable | `http` variable shadows package |
| E007 | Generic build failure | Syntax error, import cycle |
| E008 | Test failure | `go test` FAIL |
| E009 | Environment issue | Symlink/junction module resolution |

---

## File Target cho từng Project

| Project | Beads File Path |
|---|---|
| ti-router | `Z:\Ti\taskboard\docs\beads\beads.md` (system-wide) |
| auto_reg | `Z:\01_PROJECTS\auto_reg\docs\beads.md` |
| ticlaw | `Z:\01_PROJECTS\ticlaw\docs\beads.md` |

**Note:** Ti system hiện tại sử dụng **system-wide beads** tại `Z:\Ti\taskboard\docs\beads\beads.md` cho tất cả projects (ti-router, agent-store, v.v.). |

---

## Ví dụ Hoàn Chỉnh

```markdown
## [2026-04-27T12:16:00.123Z] Fix Import Alias in openai/handlers.go - claude - DONE
Actions: Changed import path from provider → providers alias, replaced all provider. → providers. in lines 14-210 | Issues: E001: undefined: provider → Package name mismatch, added import alias providers | Lessons: Package name must match import path or use alias | Verification: go build passes in layers/http/openai, go test passes in layers/http | Files: layers/http/openai/handlers.go
```

---

## Checklist trước khi Commit

- [ ] Mỗi task có entry riêng (không gộp)
- [ ] Timestamp ISO-8601 đầy đủ với milliseconds (yyyy-MM-ddTHH:mm:ss.fffZ)
- [ ] Format: ## [timestamp] <Task Name> - <agent> - <status>
- [ ] Actions, Issues, Lessons, Verification, Files trên cùng dòng, phân cách bằng |
- [ ] Actions có file:line reference
- [ ] Issues có Error Code (nếu có)
- [ ] Verification có ít nhất 1 check passed
