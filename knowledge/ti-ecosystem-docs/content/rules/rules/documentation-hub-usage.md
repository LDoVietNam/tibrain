---
description: "Quy định sử dụng Documentation Hub cho Ti ecosystem"
priority: P1
alwaysApply: true
---

# Documentation Hub Usage - Sử dụng Documentation Hub

> **Version**: 1.0.0
> **Date**: 2026-05-12
> **Purpose**: Quy định cách sử dụng Documentation Hub tại `Z:\10_WORKPLACE\Ti\docs`
> **Priority**: P1

---

## Nguyên Tắc

**Documentation Hub là nguồn thông tin chính thức cho Ti ecosystem.** Mọi tài liệu, hướng dẫn, và reference phải được centralized tại đây.

## Cấu Trúc Documentation Hub

```
Z:\10_WORKPLACE\Ti\docs\
├── README.md                           # Documentation Hub chính
├── CHANGELOG.md                        # Lịch sử thay đổi
├── COMPONENT_INDEX.md                   # Index tất cả components
├── standardized_rag_system_documentation.md  # RAG System docs
├── architecture/                       # 8 files kiến trúc
├── development/                        # 4 files development
├── phase_docs/                         # Documentation theo phases
│   ├── phase2/                        # 4 files Phase 2
│   ├── phase3/                        # 3 files Phase 3
│   └── phase4/                        # 4 files Phase 4
├── components/                         # Component documentation
├── api/                               # API documentation
├── deployment/                        # Deployment documentation
├── troubleshooting/                    # Troubleshooting guides
└── archived/                          # Archived documentation
```

## Quy Trình Sử Dụng

### 1. Khi Bắt Đầu Task Mới

**BẮT BUỘC:**
1. Đọc `Z:\10_WORKPLACE\Ti\docs\README.md` để hiểu system overview
2. Check `Z:\10_WORKPLACE\Ti\docs\COMPONENT_INDEX.md` để hiểu components liên quan
3. Đọc documentation cho components sẽ làm việc với
4. Check `Z:\10_WORKPLACE\Ti\docs\CHANGELOG.md` để了解 recent changes

**KHÔNG:**
- Bắt đầu task mà không đọc documentation hub
- Bỏ qua component index khi làm việc với components mới
- Không check changelog khi có issues liên quan đến recent changes

### 2. Khi Làm Việc Với Components

**NÊN:**
- Luôn tham khảo documentation tại `Z:\10_WORKPLACE\Ti\docs/components/`
- Check API documentation tại `Z:\10_WORKPLACE\Ti\docs/api/`
- Review architecture docs tại `Z:\10_WORKPLACE\Ti\docs/architecture/`
- Sử dụng troubleshooting guides tại `Z:\10_WORKPLACE\Ti\docs/troubleshooting/`

**KHÔNG:**
- Implement components mà không đọc documentation
- Bỏ qua architecture docs khi thay đổi system structure
- Không tham khảo troubleshooting guides khi gặp issues

### 3. Khi Thêm hoặc Cập nhật Components

**BẮT BUỘC:**
1. Cập nhật `Z:\10_WORKPLACE\Ti\docs\COMPONENT_INDEX.md`
2. Thêm documentation vào appropriate folder
3. Cập nhật `Z:\10_WORKPLACE\Ti\docs\CHANGELOG.md`
4. Update `Z:\10_WORKPLACE\Ti\docs\README.md` nếu cần
5. Test documentation links và navigation

**KHÔNG:**
- Thêm components mà không update documentation hub
- Để documentation hub outdated
- Bỏ qua changelog khi có breaking changes

## Documentation Categories

### 📖 User Documentation
- **Getting Started**: `Z:\10_WORKPLACE\Ti\docs/README.md#quick-start`
- **User Manuals**: Component-specific documentation
- **Tutorials**: Step-by-step guides
- **FAQ**: Troubleshooting section

### 🔧 Developer Documentation
- **API Reference**: `Z:\10_WORKPLACE\Ti\docs/api/`
- **Component Guides**: `Z:\10_WORKPLACE\Ti\docs/components/`
- **Development Guidelines**: `Z:\10_WORKPLACE\Ti\docs/development/`
- **Testing Documentation**: Phase documentation

### 🏗️ Architecture Documentation
- **System Design**: `Z:\10_WORKPLACE\Ti\docs/architecture/`
- **Database Design**: Architecture documentation
- **Integration Patterns**: Phase documentation
- **Performance Optimization**: Component documentation

### 🚀 Operations Documentation
- **Deployment Guides**: `Z:\10_WORKPLACE\Ti\docs/deployment/`
- **Monitoring & Alerting**: Component documentation
- **Troubleshooting**: `Z:\10_WORKPLACE\Ti\docs/troubleshooting/`
- **Maintenance Procedures**: Documentation hub

## Quality Standards

### 📝 Writing Guidelines
- Sử dụng tiếng Việt cho user-facing documentation
- Giữ thuật ngữ kỹ thuật bằng tiếng Anh
- Include practical examples và code snippets
- Provide step-by-step instructions

### 🎨 Formatting Standards
- Sử dụng Markdown cho tất cả documentation
- Include table of contents cho long documents
- Sử dụng code blocks cho examples
- Include diagrams và visual aids khi cần

### 🔗 Link Management
- Sử dụng relative links cho internal documentation
- Test tất cả links trước khi commit
- Update links khi di chuyển files
- Include backlinks cho navigation

## Maintenance Procedures

### 🔄 Regular Updates
- **Weekly**: Check documentation accuracy
- **Monthly**: Review và update outdated content
- **Quarterly**: Comprehensive documentation review
- **As Needed**: Update với feature changes

### 📊 Quality Metrics
- **Coverage**: 95%+ của components phải có documentation
- **Accuracy**: All information phải verified và up-to-date
- **Accessibility**: Easy navigation và search
- **Completeness**: Comprehensive coverage của tất cả topics

### 🚨 Alert Triggers
- Documentation hub outdated > 1 tháng
- Components không có documentation
- Broken links trong documentation
- User complaints về documentation quality

## Integration với Workflow

### Pre-Flight Checklist
- [ ] Đọc relevant documentation từ documentation hub
- [ ] Check component index cho context
- [ ] Review changelog cho recent changes
- [ ] Identify documentation gaps

### Post-Flight Checklist
- [ ] Update documentation hub nếu cần
- [ ] Add new component documentation
- [ ] Update changelog với changes
- [ ] Test documentation navigation

### Quality Gates
- Documentation must exist cho new components
- Documentation must be accurate và up-to-date
- Documentation hub must be updated với changes
- Links must be tested và working

## Troubleshooting Documentation Issues

### Common Issues
1. **Broken Links**: Check file paths và update links
2. **Outdated Content**: Review và update content
3. **Missing Documentation**: Add required documentation
4. **Navigation Issues**: Update structure và links

### Resolution Process
1. Identify issue type và scope
2. Check documentation hub structure
3. Update hoặc add missing content
4. Test navigation và links
5. Update changelog với documentation changes

---

## 📞 Getting Help

### Documentation Issues
- **Broken Links**: Report via GitHub issues
- **Content Issues**: Update directly hoặc report
- **Structure Issues**: Discuss với team
- **Navigation Issues**: Test và fix links

### Contribution Guidelines
- **New Documentation**: Follow formatting standards
- **Updates**: Maintain consistency với existing style
- **Reviews**: Peer review cho important changes
- **Testing**: Test all changes trước commit

---

*Quy định này là bắt buộc cho tất cả agents làm việc với Ti ecosystem. Documentation hub là nguồn thông tin chính thức và phải được maintain và updated.*