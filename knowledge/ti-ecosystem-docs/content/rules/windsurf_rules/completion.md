# Completion Rule

> **Priority**: P0
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\completion.md`

---

## ✅ Task Completion (Bắt Buộc)

### Verification Checklist
Trước khi báo "hoàn thành", PHẢI:
- [ ] Chạy tests liên quan (`go test ./...`, `npm test`, etc.)
- [ ] Verify code compiles/runs không lỗi
- [ ] Check linting/typecheck (nếu có)
- [ ] **TRIỂN KHAI TEST thực tế** - không chỉ verify code
- [ ] **DEMO FUNCTIONALITY** - test user-facing features
- [ ] **VERIFY END-TO-END** - test complete workflow
- [ ] Log completion với BD tool
- [ ] Update beads log (nếu task phức tạp)

### 🚫 KHÔNG BAO GIỜ Báo "Hoàn Thành" Khi:
- ❌ Code chỉ compile nhưng chưa test
- ❌ Chỉ chạy unit tests mà không test integration
- ❌ Chưa verify functionality thực tế
- ❌ Chưa demo user-facing features
- ❌ Chưa test end-to-end workflow

### ✅ BÁO "HOàn Thành" CHỈ KHI:
- ✅ Code chạy thành công không lỗi
- ✅ Tests pass (unit + integration)
- ✅ Functional demo working
- ✅ End-to-end workflow verified
- ✅ User can see/touch the result

### BD Tool Logging (Bắt Buộc)
```bash
# Task hoàn thành thành công
bd log --task="Task name - Completed successfully" \
       --agent=devin --status=completed --type=coding --domain=<domain>

# Task thất bại
bd log --task="Task name - Failed: <error message>" \
       --agent=devin --status=failed --type=coding --domain=<domain>
```

---

## 📝 Post-Completion

### Cập Nhật Context Files (Bắt Buộc)
Nếu code thay đổi cấu trúc project:
1. Update AGENTS.md (nếu thêm packages/thay đổi kiến trúc)
2. Update CODE-MAP.md (nếu docs thiếu)
3. Commit context updates cùng với code changes

### Lessons Learned
- Log vào beads: patterns phát hiện, mistakes tránh được
- Update `mandatory-quality-principles.md` nếu có insights mới
- Share với team nếu là best practice tốt

---

## 🔗 References
- **BD Tool Guide**: `Z:\docs\bd-tool-guide.md`
- **Beads Rule**: `Z:\10_WORKPLACE\Ti\content\rules\core\beads.md`
- **Quality Gates**: `Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-quality-gates.md`
