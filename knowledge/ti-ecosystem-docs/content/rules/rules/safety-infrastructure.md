# Safety Infrastructure Rule

> **Priority**: P0
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\safety-infrastructure.md`

---

## 🛡️ Safety Infrastructure (Bắt Buộc)

### Backup Before Changes
- TẤT CẢ changes đến core systems PHẢI backup trước
- Core systems: knowledge_base, taskboard, 00_SECRET, AGENTS.md
- Backup format: `<filename>.backup-<YYYYMMDD>`

### Rollback Plan
- Luôn có kế hoạch rollback trước khi implement
- Test rollback procedure nếu changes lớn
- Document rollback steps trong beads log

---

## 🔐 Core Systems Protection

### Z:\00_SECRET\
- KHÔNG bao giờ commit lên git
- Backup trước khi sửa
- Verify permissions sau khi thay đổi

### Z:\02_CORE\_cli\
- Hard links: AGENTS.md, beads.md, etc.
- Backup trước khi sửa
- Test trên môi trường khác trước khi apply

### Knowledge Base / Taskboard
- Export trước khi modify
- Chỉnh sửa qua tools (không edit raw files nếu có thể)
- Verify integrity sau khi sửa

---

## 🚨 Emergency Procedures

### Khi Gặp Lỗi Nghiêm Trọng
1. Stop immediate changes
2. Restore from backup
3. Log incident với BD tool
4. Báo cáo user
5. Investigate root cause

---

## 🔗 References
- **Safety Rules**: `Z:\docs\safety-rules.md`
- **Secrets Management**: `Z:\docs\secrets-management.md`
- **BD Tool Guide**: `Z:\docs\bd-tool-guide.md`
