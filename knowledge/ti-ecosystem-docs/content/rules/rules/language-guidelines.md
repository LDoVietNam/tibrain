# Language Guidelines

> **Priority**: P1
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\language-guidelines.md`

---

## 🌐 Ngôn Ngữ Giao Tiếp (BẮT BUỘC)

**TẤT CẢ CLI PHẢI giao tiếp bằng tiếng Việt với người dùng.**

- Chỉ sử dụng tiếng Việt trong mọi phản hồi, giải thích, thông báo lỗi
- Thuật ngữ kỹ thuật (API, CLI, YAML, JSON, v.v.) giữ nguyên tiếng Anh
- Không sử dụng tiếng Anh trong câu trả lời trừ khi người dùng yêu cầu cụ thể
- Áp dụng cho tất cả CLI: claude, codex, gemini, qwen, kilo, opencode và các CLI khác trong hệ thống

---

## 📝 Hướng Dẫn Cụ Thể

### Phản Hồi Cho Người Dùng
- Dùng tiếng Việt ngắn gọn, trực tiếp
- Không cần mở đầu/lời kết dài dòng
- Chỉ giải thích khi người dùng yêu cầu

### Thuật Ngữ Kỹ Thuật
- Giữ nguyên: API, CLI, YAML, JSON, HTTP, REST, GraphQL, etc.
- Không dịch các thuật ngữ này sang tiếng Việt

### Error Messages
- Log lỗi bằng tiếng Việt (nếu là thông báo cho user)
- Giữ nguyên error message gốc từ system/code (để debug)

---

## 🔗 References
- **Global AGENTS.md**: `Z:\AGENTS.md` (Section: 🌐 Ngôn Ngữ Giao Tiếp)
- **Agent Guidelines**: `Z:\10_WORKPLACE\Ti\content\agent-guidelines.md`
