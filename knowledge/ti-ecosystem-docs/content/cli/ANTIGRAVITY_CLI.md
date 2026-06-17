---
title: Antigravity CLI Patterns
scope: cli
tags: ["tibrain", "cli", "documentation", "authentication", "provider-antigravity"]
scopes: ["auth", "providers", "cli", "tibrain"]
last_updated: 2026-05-22
---

# Antigravity CLI (AGY CLI) - Patterns & Configuration

Tài liệu này tổng hợp toàn bộ các pattern, thư mục cấu hình và cách sử dụng Antigravity CLI (Agent) dựa trên tài liệu chính thức từ `https://antigravity.google/docs`. Các Agent trong hệ sinh thái TiBrain cần tham chiếu tài liệu này khi được yêu cầu cấu hình hoặc tương tác với Antigravity CLI.

## 1. Cấu hình Thư mục & Settings (Config Folder Patterns)

Hệ thống Antigravity CLI quản lý cấu hình thông qua các file JSON được lưu trữ trong thư mục gốc của người dùng:
- **Đường dẫn thư mục cấu hình:** `~/.gemini/antigravity-cli/`
- **File Settings:** `~/.gemini/antigravity-cli/settings.json`
  - Chứa cấu hình workspace, safety restrictions, editor preferences, visual style, và hiệu suất.
- **File Keybindings:** `~/.gemini/antigravity-cli/keybindings.json`
  - Chứa toàn bộ map phím tắt tùy chỉnh. Nếu file bị lỗi, CLI sẽ fallback về mặc định. Nếu xóa file, keybindings sẽ reset.

*Lưu ý khi Agent cấu hình:*
Bất kỳ thay đổi cấu hình nào cho Antigravity (ví dụ: kết nối Custom Router) đều cần phải tương tác hoặc chỉnh sửa các file `.json` bên trong thư mục `~/.gemini/antigravity-cli/`.

## 2. Authentication & Security

- **Cơ chế mặc định:** CLI sử dụng system keyring để xác thực ngầm.
- **Dự phòng (Fallback):** Mở trình duyệt Google Sign-In cục bộ hoặc in ra URL để xác thực nếu dùng qua SSH.
- **Enterprise / GCP:** Hỗ trợ login qua enterprise credentials bằng cách connect GCP project lúc onboard.
- **Đăng xuất:** Sử dụng lệnh `/logout` để xóa credentials khỏi hệ thống.

## 3. Các Lệnh Điều khiển (Slash Commands)

Đây là các công cụ giao tiếp/điều khiển quan trọng trong TUI (Terminal UI) của Antigravity CLI:
- `/config` hoặc `/settings`: Mở panel quản lý cấu hình (full-screen overlay).
- `/keybindings`: Quản lý phím tắt.
- `/permissions`: Quản lý quyền hệ thống cho các tools.
- `/goal`: Chạy Agent liên tục cho đến khi hoàn thành mục tiêu (không yêu cầu prompt trung gian).
- `/grill-me`: Yêu cầu Agent phỏng vấn/hỏi ngược người dùng để làm rõ requirement trước khi implement.
- `/schedule`: Thiết lập công việc chạy định kỳ hoặc hẹn giờ.
- `/browser`: Ép buộc Agent sử dụng Browser/Chrome DevTools một cách tường minh (yêu cầu cấp quyền Chrome).
- `/rewind` / `/undo`: Lùi lại lịch sử trò chuyện.
- `/fork`: Tách nhánh hội thoại hiện tại ra một workspace mới.
- `/clear`: Xóa prompt và bắt đầu hội thoại mới.
- `/resume`: Liệt kê và tiếp tục các log hội thoại cũ (Khi CLI đóng, nó tự in ra lệnh resume chính xác).

## 4. Các phím tắt (Keybindings) cốt lõi

- **Mở Editor mặc định:** `ctrl+g` (để viết prompt dài).
- **Edit Terminal Command:** `e` (khi Agent đề xuất lệnh bash/powershell).
- **Phê duyệt lệnh Terminal:** `y` (Yes) / `n` (No).
- **Hủy luồng/Đóng panel:** `ctrl+c` hoặc `esc`.
- **Thoát CLI:** `ctrl+d`.
- **Gửi Prompt:** `enter` (Để xuống dòng: `alt+enter` / `shift+enter`).
- **Autocompletion (Đường dẫn file):** Gõ `@` để kích hoạt gợi ý.
- **Thực thi lệnh Shell:** Bắt đầu prompt bằng `!` (ví dụ: `!ls -la`).

## 5. Overrides (CLI Flags)

Cấu hình có thể bị ghi đè tạm thời trong một session cụ thể thông qua cờ (flags) khi khởi động:
- Ví dụ: `--sandbox` hoặc `--dangerously-skip-permissions`.
- TUI sẽ hiển thị thông báo ghi đè (VD: "Sandbox Mode on overridden by --sandbox").
