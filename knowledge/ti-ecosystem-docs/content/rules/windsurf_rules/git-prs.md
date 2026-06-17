# Git PRs Rule

> **Priority**: P1
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\git-prs.md`

---

## 📦 Git & Pull Requests

### Commit Messages
- Sử dụng tiếng Việt (trừ thuật ngữ kỹ thuật)
- Format: `<type>: <mô tả ngắn>`
- Types: feat, fix, refactor, docs, test, chore

### Khi Nào Được Commit
- ✅ User yêu cầu cụ thể "commit changes"
- ✅ Hoàn thành task và user approve
- ❌ KHÔNG commit tự động (trừ khi user chỉ định)

### Khi Nào Được Push
- ✅ User yêu cầu cụ thể "push to remote"
- ✅ Sau khi đã commit và verify
- ❌ KHÔNG force push (trừ khi user yêu cầu rõ ràng)

---

## 🔄 Pull Request Process

### Tạo PR
1. Ensure all tests pass
2. Write clear PR description (tiếng Việt)
3. Link to relevant issues (nếu có)
4. Request review từ appropriate people

### PR Description Template
```markdown
## Tóm Tắt
- [ ] Thay đổi chính 1
- [ ] Thay đổi chính 2

## Loại Thay Đổi
- [ ] Feature mới
- [ ] Bug fix
- [ ] Refactor
- [ ] Documentation

## Test Plan
- [ ] Đã chạy tests
- [ ] Đã verify thủ công
```

---

## 🔗 References
- **Git Workflow**: `Z:\docs\continuous-integration.md`
- **Change Constraints**: `Z:\10_WORKPLACE\Ti\content\rules\change-constraints.md`
