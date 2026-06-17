# Evidence Rule

> **Priority**: P0
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\evidence.md`

---

## 🔍 Evidence-Based Decisions

### Bắt Buộc Thu Thập Context Trước Khi Quyết Định
1. ✅ Scan project structure
2. ✅ Extract code patterns (imports, naming, style)
3. ✅ Load project rules (AGENTS.md, .editorconfig, linting config)
4. ✅ Check recent git history (last 10 commits)
5. ✅ Analyze dependencies

---

## 📊 Evidence Sources

### Code Patterns
- Sử dụng Glob/Grep để tìm patterns
- Check existing implementations trước khi tạo mới
- Học từ codebase conventions

### Project Rules
- Đọc AGENTS.md của project
- Đọc .editorconfig, linting configs
- Tham khảo docs/architecture

### Git History
- `git log --oneline -10` để xem recent changes
- Hiểu lý do tại sao code được viết như vậy

---

## 🚫 Không Được Phép
- ❌ Quyết định dựa trên giả định không có cơ sở
- ❌ Copy-paste từ internet mà không adapt to project conventions
- ❌ Ignore existing patterns trong codebase

---

## 🔗 References
- **Context Collection**: `Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-quality-principles.md`
- **Exploration Rule**: `Z:\10_WORKPLACE\Ti\content\rules\core\exploration.md`
