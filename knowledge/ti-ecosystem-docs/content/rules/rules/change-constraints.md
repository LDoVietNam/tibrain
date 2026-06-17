# Change Constraints Rule

> **Priority**: P0
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\change-constraints.md`

---

## 🔒 Change Constraints (Bắt Buộc)

### Scope Discipline
- Chỉ sửa những gì nằm trong scope của task
- Không "refactor cho vui" khi đang làm task khác
- Nếu thấy issue khác → log vào beads, xử lý sau

### Minimal Changes
- Chỉ thay đổi những dòng code cần thiết
- Không reformat toàn bộ file nếu chỉ sửa 1 function
- Giữ nguyên code style của file đó

---

## 🚧 Safe Changes
### Được Phép
- ✅ Sửa bug trong scope
- ✅ Thêm feature theo yêu cầu
- ✅ Refactor nhỏ liên quan đến task
- ✅ Thêm tests cho code mới

### Không Được Phép
- ❌ Refactor toàn bộ module khi chỉ sửa 1 bug
- ❌ Upgrade dependencies lớn khi không cần thiết
- ❌ Thay đổi API public khi không có lý do chính đáng
- ❌ Xóa code "cũ" mà không có test coverage

---

## 📝 Logging Changes
Sử dụng BD tool để log tất cả changes:
```bash
bd log --task="Task description" --agent=devin --status=completed --type=coding --domain=<domain>
```

---

## 🔗 References
- **Implementation Rule**: `Z:\10_WORKPLACE\Ti\content\rules\core\implementation.md`
- **BD Tool Guide**: `Z:\docs\bd-tool-guide.md`
