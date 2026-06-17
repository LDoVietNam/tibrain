# Boundaries Rule

> **Priority**: P0
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\boundaries.md`

---

## 🚧 Agent Boundaries

### Không Được Phép
- ❌ Commit code khi không được user yêu cầu cụ thể
- ❌ Push to remote repository (trừ khi user chỉ định)
- ❌ Sửa đổi git config hoặc hooks
- ❌ Xóa files mà không có lý do chính đáng
- ❌ Thêm secrets/API keys vào code
- ❌ Chạy destructive commands (rm -rf, git reset --hard) mà không hỏi user

### Được Phép
- ✅ Chỉnh sửa files theo yêu cầu của user
- ✅ Chạy local build/test để verify changes
- ✅ Tạo new files (theo yêu cầu)
- ✅ Suggest improvements, nhưng phải hỏi trước khi implement lớn

---

## 🔒 Security Boundaries

### Secrets Management
- Tất cả secrets nằm tại `Z:\00_SECRET\`
- KHÔNG bao giờ log/commit secrets
- KHÔNG hardcode API keys trong code
- Sử dụng environment variables hoặc config files (trỏ đến Z:\00_SECRET\)

### File System Boundaries
- Chỉ làm việc trong Z:\ và các subfolders
- Không truy cập outside Z:\ trừ khi user chỉ định rõ
- Backup trước khi sửa files quan trọng

---

## 🔗 References
- **Secrets Management**: `Z:\docs\secrets-management.md`
- **Safety Rules**: `Z:\docs\safety-rules.md`
