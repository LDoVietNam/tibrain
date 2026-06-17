---
description: "Review code 5 trục - kết nối Tibrain code-review skills và beads status"
priority: P0
alwaysApply: true
---

# Review - Đánh Giá Chất Lượng (BẮT BUỘC)

> **Version**: 2.0.0
> **Date**: 2026-05-05
> **Purpose**: Quy trình self-review trước khi báo hoàn thành
> **Priority**: P0 (BẮT BUỘC - không được bỏ qua)

---

## Khi Nào Review

- [ ] Sau khi hoàn thành mỗi bead
- [ ] Trước khi báo task hoàn thành
- [ ] Khi code do agent khác/model khác tạo ra
- [ ] Sau bug fix (review cả fix và regression test)

## The Five-Axis Review

Mỗi thay đổi được review qua 5 trục:

### 1. Correctness
- [ ] Change đúng với spec/task requirements
- [ ] Edge cases được handle
- [ ] Error paths được handle
- [ ] Tests cover change adequately

### 2. Readability & Simplicity
- [ ] Names rõ ràng, consistent
- [ ] Logic straightforward, không unnecessarily complex
- [ ] Không premature abstraction

### 3. Architecture
- [ ] Follow existing patterns trong codebase
- [ ] Không unnecessary coupling hoặc dependencies
- [ ] Abstraction level phù hợp

### 4. Security
- [ ] No secrets trong code
- [ ] Input validated at boundaries
- [ ] No injection vulnerabilities
- [ ] Auth checks đầy đủ

### 5. Performance
- [ ] No N+1 patterns
- [ ] No unbounded operations
- [ ] Pagination cho list endpoints

## Review Checklist

```markdown
## Review: [Bead/Task title]

### Context
- [ ] Tôi hiểu change này làm gì và tại sao

### Correctness
- [ ] Change khớp spec
- [ ] Edge cases handled
- [ ] Error paths handled
- [ ] Tests cover adequately

### Readability
- [ ] Names clear và consistent
- [ ] Logic straightforward
- [ ] No unnecessary complexity

### Architecture
- [ ] Follows existing patterns
- [ ] No unnecessary coupling
- [ ] Appropriate abstraction

### Security
- [ ] No secrets in code
- [ ] Input validated
- [ ] No injection risks
- [ ] Auth checks in place

### Performance
- [ ] No N+1 patterns
- [ ] No unbounded operations
- [ ] Pagination if applicable

### Verification
- [ ] Tests pass
- [ ] Build succeeds
- [ ] Manual verification done (if applicable)

### Verdict
- [ ] **Approve** — Ready
- [ ] **Fix required** — Issues must be addressed
```

## Tibrain Integration

### Code Review Skills

```bash
# Lấy security review checklist
py "Z:\Ti\router\tibrain-cli.py" execute best-security-review --params "{\"task\": \"Review auth changes\"}"

# Lấy performance review patterns
py "Z:\Ti\router\tibrain-cli.py" execute best-strategic-compact --params "{\"task\": \"Review for over-engineering\"}"
```

### Model Selection

```bash
# Chọn model cho review (thường cần model có reasoning tốt)
py "Z:\Ti\router\tibrain-cli.py" execute router.select_auto --params "{\"task_type\": \"review\"}"
```

## Dead Code Hygiene

- Remove unused imports, variables, functions
- Không để code commented-out
- Không để debug logs trong production code

## Common Rationalizations (Cẩn Thận)

| Lý do | Thực tế |
|-------|---------|
| "Nó chạy là đủ" | Code chạy nhưng unreadable, insecure, architecturally wrong = tech debt |
| "Tôi viết nên tôi biết đúng" | Authors blind với assumptions của chính mình |
| "Sẽ dọn sau" | Later never comes. Review là quality gate |
| "AI-generated nên OK" | AI code cần scrutiny nhiều hơn, không ít |
| "Tests pass nên OK" | Tests cần nhưng không đủ. Không bắt architecture, security, readability |

## Red Flags

- Merge without review
- Review chỉ check tests pass
- "LGTM" without evidence
- Security-sensitive changes không security review
- PR quá lớn để review properly → split

## Anti-Patterns

- ❌ **Không review code của chính mình** — Self-review bắt buộc
- ❌ **Review qua loa** — Checkbox mentality, không read code
- ❌ **Bỏ qua red flags** — "Tạm được" với issues nghiêm trọng
- ❌ **Review quá muộn** — Sau khi đã merge hoặc deploy

## Kết Nối Beads

```bash
# Mark bead review sau khi review xong
py "Z:\Ti\router\tibrain-cli.py" execute beads.update --params "{\"id\": \"<bead-id>\", \"status\": \"review\", \"comment\": \"5-axis review complete: 1 issue found and fixed\"}"
```
