# Quality Enforcement Rule

> **Priority**: P0
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\quality\enforcement.md`

---

## 🔒 Enforcement Mechanisms (BẮT BUỘC)

### 1. Pre-Commit Gate
Trước khi commit (khi user yêu cầu), PHẢI check:
- [ ] Tests pass (`go test ./...`, `npm test`, `pytest`)
- [ ] Linting pass (`golangci-lint run`, `npm run lint`, etc.)
- [ ] Build thành công (`go build`, `npm run build`)
- [ ] Review checklist completed (5-axis review)
- [ ] BD tool logged completion

**Nếu fail bất kỳ check nào → KHÔNG được commit, phải fix trước**

### 2. Task Completion Gate
Trước khi báo "hoàn thành" cho user, PHẢI:
- [ ] Chạy tests liên quan → pass
- [ ] Verify code compiles/runs → ok
- [ ] Check linting/typecheck → pass
- [ ] Log completion với BD tool
- [ ] Update context files (nếu thay đổi cấu trúc)

**Nếu fail → không báo hoàn thành, tiếp tục fix**

---

## 📊 Quality Metrics (BẮT BUỘC Định Nghĩa)

### Code Coverage (Target: ≥ 80%)
```bash
# Go
go test -cover ./...

# TypeScript
npm test -- --coverage

# Python
pytest --cov=src
```

### Linting (Target: 0 errors, 0 warnings)
```bash
# Go
golangci-lint run ./...

# TypeScript
npm run lint

# Python
pylint src/
flake8 src/
```

### Build Success (Target: 100%)
```bash
# Go
go build ./...

# TypeScript
npm run build

# Rust
cargo build
```

### Review Checklist (Target: 100% items checked)
- 5-axis review hoàn thành đầy đủ
- Verification section completed
- Verdict: Approve

---

## 🚨 Violation Detection & Escalation

### Auto-Detection
Sau mỗi task, agent PHẢI self-check:
1. Có bỏ qua mandatory rule nào không?
2. Có skip review không?
3. Có commit khi tests fail không?

### Violation Logging (Bắt Buộc)
```bash
bd log --task="Quality violation: {violation_type} - {details}" \
       --agent=devin --status=failed --type=violation --domain=<domain>
```

### Escalation Thresholds
| Violation Type | Threshold | Action |
|---------------|-----------|--------|
| Skip mandatory rule | ≥ 1 lần | Log violation, explain to user |
| Skip review (P0) | ≥ 1 lần | Log violation, MUST fix before continue |
| Commit when tests fail | ≥ 1 lần | Rollback, log violation, explain |
| Repeated violations | ≥ 3 lần | Escalate to human review |

---

## 🔗 Integration với Workflows

### Trong reflective-loop-beads.yaml
- Stage: "Quality Validation"
- Checks: coverage, linting, build, review
- Fail → rollback to previous bead

### Trong ti-cli-v2.yaml
- Stage: "Pre-commit Gate"
- Checks: tests, linting, build
- Fail → return to implementation

---

## 📝 Quality Report Template

Sau mỗi task, log vào BD tool:
```bash
bd log --task="Quality Report: coverage={coverage}%, lint={lint_errors}, build={build_status}, review={review_status}" \
       --agent=devin --status=completed --type=metrics --domain=<domain>
```

---

## 🔗 References
- **Review Rule**: `Z:\10_WORKPLACE\Ti\content\rules\core\review.md` (P0)
- **Quality Principles**: `Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-quality-principles.md`
- **Quality Gates**: `Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-quality-gates.md`
- **BD Tool Guide**: `Z:\docs\bd-tool-guide.md`
