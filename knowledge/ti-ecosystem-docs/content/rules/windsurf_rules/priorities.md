# Priorities Rule

> **Priority**: P0
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\priorities.md`

---

## 🎯 Priority Levels

### P0 - Bắt Buộc (Mandatory)
- Phải thực hiện trước mọi thứ khác
- Không được bỏ qua
- Vi phạm sẽ gây lỗi nghiêm trọng

### P1 - Quan Trọng (Important)
- Nên thực hiện khi có thể
- Có thể tạm hoãn nếu có lý do chính đáng
- Ảnh hưởng đến chất lượng nhưng không gây lỗi hệ thống

### P2 - Tùy Chọn (Optional)
- Thực hiện khi có thời gian
- Không ảnh hưởng đến chức năng chính
- Cải thiện trải nghiệm/hiệu suất

---

## 📋 Áp Dụng Priority

### Khi Thực Hiện Task
1. Đọc priority của task/rule
2. Thực hiện P0 trước, sau đó P1, cuối cùng P2
3. Nếu conflict giữa 2 P0 tasks → báo user để quyết định

### Khi Gặp Multiple Issues
- Fix P0 issues trước
- P1 issues có thể fix song song (nếu không dependent)
- Log tất cả P2 issues để xử lý sau

---

## 🔗 References
- **INDEX.md**: `Z:\10_WORKPLACE\Ti\content\rules\INDEX.md`
- **Workflow Selection**: `Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-workflow-selection.md`
