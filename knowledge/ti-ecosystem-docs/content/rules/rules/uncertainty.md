# Uncertainty Rule

> **Priority**: P0
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\uncertainty.md`

---

## ❓ Khi Nào Cần Hỏi User

### Thresholds Theo Category

| Category | Hỏi Khi Confidence < | Ví Dụ |
|----------|---------------------|-------|
| **Security** | 90% | Auth methods, encryption, secrets |
| **Architecture** | 70% | Layer choice, pattern selection |
| **Style** | 60% | Indentation, naming conventions |
| **Testing** | 50% | Test framework choice |

### Mọi Trường Hợp Nên Hỏi User
- Decision có thể cause data loss hoặc corruption
- Không có clear pattern trong codebase
- Decision involves user data hoặc authentication
- Confidence < 70% cho non-critical decisions

---

## 🤔 Khi Nào Có Thể Assume

### Được Phép Assume (Không Cần Hỏi)
- Confidence ≥ 80% cho style decisions
- Confidence ≥ 70% cho architecture decisions
- Clear pattern exists trong codebase
- Decision is reversible với rollback
- Project rules provide guidance (AGENTS.md, .editorconfig, linting config)

### Assumption Documentation (Bắt Buộc)
Khi assume, PHẢI document:
1. What was assumed
2. Why (reasoning)
3. Confidence level
4. Context used

---

## 🔗 References
- **Quality Principles**: `Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-quality-principles.md`
- **Decision Making**: `Ti-learning-lab/03_Knowledge/CLI/SUB_AGENT_GUIDE.md`
