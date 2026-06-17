# Project Context - Ti-Learning-Lab

> **Context về dự án Ti-Learning-Lab cho agents**
> **Last Updated**: 2026-05-06

---

## 🎯 Mục Tiêu Dự Án

**Ti-Learning-Lab** là learning & development hub - nơi lưu trữ kiến thức, code samples, projects và tài liệu học tập.

### Goal
**Xây dựng hệ thống learning pipeline có cấu trúc để chuyển đổi kiến thức thụ động thành năng lực thực tế.**

### Pipeline
```
Learning → Knowledge → HandsOn → Production
```

### Nguyên Tắc
1. **Actionable Knowledge** - Mọi tài liệu phải có thể áp dụng vào thực tế
2. **Verified Knowledge** - Kiến thức phải được test trong real projects
3. **Connected Knowledge** - Các stages phải có liên kết rõ ràng
4. **Quality over Quantity** - Chỉ giữ lại tài liệu có giá trị thực sự

---

## 📁 Cấu Trúc Project

```
Ti-learning-lab/
├── 01_Learning/              # Learning materials (tutorials, guides, patterns để HỌC)
│   ├── BrowserAutomation/    # Browser automation learning
│   ├── CLI/                  # CLI learning
│   ├── Configs/              # Configuration learning
│   ├── Core/                 # Core learning
│   ├── MCPHub/               # MCP Hub learning
│   ├── TiBrain/              # TiBrain learning
│   └── [project-name]/       # Individual project learning
├── 03_Knowledge/             # Reference materials (API docs, architecture, technical specs)
│   ├── 00_META/              # Metadata (AUDIT_REPORT, CROSS_REFERENCE_EXAMPLE)
│   ├── agents/               # Agent framework
│   ├── cli/                  # CLI patterns
│   ├── patterns/             # Software patterns
│   ├── lessons/              # Lessons learned
│   └── [category]/           # Other categories
├── 05_Taskboard/             # Task tracking (beads.md)
│   ├── beads.md             # Beads log
│   ├── storage/             # Schema and metadata files
│   └── tasks/               # Task definitions
├── 06_HandsOn/              # Thực hành (POC, prototype, spike)
│   ├── codex-autoresearch/  # Codex auto-research project
│   └── [project-name]/      # Individual projects
├── 07_Repositories/          # Repository references (git repos)
│   ├── storage/             # Zip archives
│   ├── undetectable-fingerprint-browser/ # Browser fingerprint research
│   └── [repo-name]/         # Individual repos
├── 08_Archives/              # Archived projects
├── templates/               # Templates cho stages
│   ├── 01_Learning_Template.md
│   ├── 06_HandsOn_Template.md
│   └── README.md
├── WORKFLOW.md              # Workflow guide
├── SUBFOLDER_AUDIT_REPORT.md # Subfolder audit report
├── AGENTS.md                # Agent rules (BẮT BUỘC ĐỌC)
├── README.md                # Documentation
├── MANIFEST.md              # Project manifest
└── .devin/                   # Devin config
```

---

## 🛠️ Workflow Chuẩn

### 1. Learning Phase
```
Đọc tutorials/guides → Notes → 01_Learning/
```
- Sử dụng template: templates/01_Learning_Template.md
- Focus trên understanding, không phải implementation

### 2. Knowledge Phase
```
Tổng hợp từ Learning → Patterns/Runbooks → 03_Knowledge/
```
- Tạo reference materials
- Tạo cross-references
- Thêm metadata (verified, last_used, status)

### 3. HandsOn Phase
```
Áp dụng Knowledge → POC/Prototype → 06_HandsOn/
```
- Sử dụng template: templates/06_HandsOn_Template.md
- Test solution trước khi merge vào production

### 4. Production Phase
```
HandsOn validated → Merge vào production project
```
- Move đến Z:\10_WORKPLACE\Ti\ hoặc Z:\01_PROJECTS\
- Update cross-references

---

## 📝 Quy Tắc Cho Agents

### BẮT BUỘC Khi Bắt Đầu Task
1. Đọc AGENTS.md (project rules)
2. Đọc README.md (project overview)
3. Đọc WORKFLOW.md (workflow guide)
4. Đọc MANIFEST.md (project manifest)
5. Check SUBFOLDER_AUDIT_REPORT.md nếu cần cleanup

### Khi Làm Việc Với Docs
1. Sử dụng templates từ templates/
2. Tạo cross-references giữa stages
3. Thêm metadata (verified, last_used, status)
4. Update MANIFEST.md sau khi hoàn thành
5. Log task với bd tool

### Khi Cleanup Structure
1. Tham khảo SUBFOLDER_AUDIT_REPORT.md
2. Xóa folders trống
3. Move files đến đúng chỗ
4. Update README.md và MANIFEST.md

---

## 📊 Success Metrics

- **Application Rate**: >50% (docs được apply vào projects)
- **Verification Rate**: >80% (docs có verification status)
- **Connection Rate**: >70% (docs có cross-references)
- **Deprecation Rate**: >10% (docs obsolete được deprecate)

---

## 🔗 Liên Kết

### Với Ti Projects
```
Ti-learning-lab → Learning → Z:\10_WORKPLACE\Ti\
               → Knowledge → Z:\10_WORKPLACE\Ti\
               → HandsOn → Z:\10_WORKPLACE\Ti\
```

### Với Production
```
Ti-learning-lab → Validated POC → Z:\01_PROJECTS\
```

---

## 📞 Support

**Questions về Ti-Learning-Lab**:
- Xem AGENTS.md để hiểu rules
- Xem README.md để hiểu structure
- Xem WORKFLOW.md để hiểu workflow
- Xem MANIFEST.md để hiểu status

---

*Last Updated: 2026-05-06*
