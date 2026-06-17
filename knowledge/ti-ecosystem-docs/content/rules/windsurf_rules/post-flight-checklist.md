# Post-Flight Checklist

> **Category**: Quality Rules  
> **Priority**: P0 (BẮT BUỘC)  
> **Last Updated**: 2026-06-05  

---

## BẮT BUỘC Post-Flight Checklist

**Mục đích**: BẮT BUỘC hoàn thành checklist này sau khi hoàn thành task để đảm bảo chất lượng

---

## Checklist

### 1. Deliverables Verification (BẮT BUỘC)

- [ ] **Tất cả deliverables đã hoàn thành**
  - Core functionality: ✅
  - Additional features: ✅
  - Bug fixes: ✅
  - Performance improvements: ✅

- [ ] **Tất cả acceptance criteria đã meet**
  - Criteria 1: ✅
  - Criteria 2: ✅
  - Criteria 3: ✅

- [ ] **Tất cả files đã tạo/modified**
  - New files: {count} files
  - Modified files: {count} files
  - Deleted files: {count} files

- [ ] **Tất cả code đã reviewed**
  - Self-review completed
  - Code style consistent
  - No obvious bugs
  - No security issues

### 2. Quality Verification (BẮT BUỘC)

- [ ] **Tests đã viết và pass**
  - Unit tests: {count} tests, {coverage}% coverage
  - Integration tests: {count} tests, all pass
  - E2E tests: {count} tests, all pass
  - Test coverage: {coverage}% (target: {target}%)

- [ ] **Code quality checks đã pass**
  - Linting: ✅ pass
  - Type checking: ✅ pass
  - Static analysis: ✅ pass
  - Security scan: ✅ pass

- [ ] **Build đã thành công**
  - Build command: {command}
  - Build time: {time}
  - Build artifacts: {artifacts}
  - Build status: ✅ success

- [ ] **Documentation đã update**
  - README updated: ✅
  - API docs updated: ✅
  - Changelog updated: ✅
  - Comments added: ✅

- [ ] **No TODOs/FIXMEs đã bỏ lại**
  - TODO count: 0
  - FIXME count: 0
  - HACK count: 0

### 3. Functional Verification (BẮT BUỘC)

- [ ] **Core functionality đã test**
  - Function 1: ✅ working
  - Function 2: ✅ working
  - Function 3: ✅ working

- [ ] **Edge cases đã test**
  - Edge case 1: ✅ handled
  - Edge case 2: ✅ handled
  - Edge case 3: ✅ handled

- [ ] **Error handling đã test**
  - Error 1: ✅ handled
  - Error 2: ✅ handled
  - Error 3: ✅ handled

- [ ] **Performance đã verify**
  - Response time: {time} (target: {target})
  - Throughput: {throughput} (target: {target})
  - Memory usage: {memory} (target: {target})

- [ ] **Security đã verify**
  - No SQL injection: ✅
  - No XSS: ✅
  - No CSRF: ✅
  - Secrets not exposed: ✅

### 4. Integration Verification (BẮT BUỘC)

- [ ] **Git status đã check**
  - Uncommitted changes: 0
  - Untracked files: {count} files
  - Staged files: {count} files

- [ ] **Git diff đã review**
  - Changes reviewed: ✅
  - No unintended changes: ✅
  - No sensitive data: ✅

- [ ] **Dependencies đã update**
  - New dependencies: {count}
  - Updated dependencies: {count}
  - Removed dependencies: {count}

- [ ] **Breaking changes đã identify**
  - Breaking changes: {count}
  - Migration notes: {notes}
  - Backward compatibility: {status}

### 5. Documentation Verification (BẮT BUỘC)

- [ ] **Code comments đã add**
  - Complex logic: ✅ commented
  - Public APIs: ✅ documented
  - Edge cases: ✅ documented

- [ ] **README đã update**
  - New features: ✅ documented
  - Usage examples: ✅ added
  - Configuration: ✅ documented

- [ ] **API docs đã update**
  - New endpoints: ✅ documented
  - Request/response: ✅ documented
  - Error codes: ✅ documented

- [ ] **Changelog đã update**
  - Changes: ✅ listed
  - Version: {version}
  - Date: {date}

### 6. Final Sign-Off (BẮT BUỘC)

- [ ] **Pre-flight checklist đã complete**
  - Score: {score}/5
  - All items: ✅

- [ ] **Task classification đã correct**
  - Type: {type}
  - Workflow: {workflow}
  - Skills: {skills}

- [ ] **Workflow đã followed**
  - Steps completed: {count}/{count}
  - All beads logged: ✅

- [ ] **Quality gates đã pass**
  - Gate 1: ✅ pass
  - Gate 2: ✅ pass
  - Gate 3: ✅ pass

- [ ] **Metrics đã log**
  - BD log: ✅
  - Quality metrics: ✅
  - Performance metrics: ✅

---

## Sign-Off

**Agent**: Devin  
**Task**: {task_description}  
**Classification**: {task_type}  
**Workflow**: {workflow}  
**Duration**: {duration}

**Post-Flight Checklist Score**: {score}/6

**Deliverables Status**: ✅ All complete

**Quality Status**: ✅ All checks pass

**Decision**: ✅ **Task complete** / ❌ **Need rework**

**If Need Rework:**
- Failed items: {list}
- Rework required: {list}
- Estimated rework time: {time}

---

## Enforcement

### BẮT BUỘC

- **BẮT BUỘC** hoàn thành checklist sau khi hoàn thành task
- **BẮT BUỘC** log sign-off với BD tool
- **BẮT BUỘC** rework nếu có failed items
- **BẮT BUỘC** không cho phép mark COMPLETE nếu deliverables FAIL
- **BẮT BUỘC** mark FAILED nếu task không hoàn thành

### Quality Gate Enforcement

**Gate 1: Deliverables Verification (BẮT BUỘC)**
- Core functionality: ✅ Working
- Acceptance criteria: ✅ All met
- Files created/modified: ✅ As expected
- **IF FAIL → Task FAILED, not COMPLETE**

**Gate 2: Quality Verification (BẮT BUỘC)**
- Tests pass: ✅
- Build success: ✅
- Linting pass: ✅
- **IF FAIL → Task FAILED, not COMPLETE**

**Gate 3: Integration Verification (BẮT BUỘC)**
- Git status: ✅ Clean
- No unintended changes: ✅
- **IF FAIL → Task FAILED, not COMPLETE**

### Violation Detection

- Skip checklist → Log violation
- Incomplete checklist → Log violation
- Proceed without sign-off → Log violation
- **Mark complete when deliverables fail → Log violation**
- **Skip deliverables verification → Log violation**

### Violation Logging

```bash
bd log --task="Post-flight violation: {violation_type} - {details}" \
       --agent=devin \
       --status=failed \
       --type=violation \
       --domain=general
```

### Escalation

- Repeated violations (≥3 times) → Escalate to human review
- Critical violations (skip quality verification, mark complete when fail) → Immediate escalation

---

## Quick Reference

**Post-Flight Checklist Summary:**
1. Deliverables Verification (4 items)
2. Quality Verification (5 items)
3. Functional Verification (5 items)
4. Integration Verification (4 items)
5. Documentation Verification (4 items)
6. Final Sign-Off (5 items)

**Total:** 27 items

**Minimum to proceed:** 24/27 items (89%)
**Recommended:** 27/27 items (100%)

---

## Example

```markdown
## Post-Flight: Implement user authentication

### 1. Deliverables Verification
✅ All deliverables complete
✅ All acceptance criteria met
✅ 3 files created, 2 files modified
✅ Code reviewed

### 2. Quality Verification
✅ Tests written and pass (15 tests, 85% coverage)
✅ Code quality checks pass
✅ Build successful (go build, 12s)
✅ Documentation updated
✅ No TODOs/FIXMEs left

### 3. Functional Verification
✅ Core functionality tested (login, logout, token validation)
✅ Edge cases tested (invalid credentials, expired tokens)
✅ Error handling tested (user not found, invalid token)
✅ Performance verified (50ms response time, target 100ms)
✅ Security verified (no injection, no XSS, no CSRF)

### 4. Integration Verification
✅ Git status checked (3 staged files)
✅ Git diff reviewed (no unintended changes)
✅ Dependencies updated (1 new dependency: golang-jwt/jwt)
✅ Breaking changes identified (0)

### 5. Documentation Verification
✅ Code comments added
✅ README updated
✅ API docs updated
✅ Changelog updated

### 6. Final Sign-Off
✅ Pre-flight checklist complete (23/23)
✅ Task classification correct (feature)
✅ Workflow followed (reflective-loop-beads.yaml)
✅ Quality gates pass (all 3 gates)
✅ Metrics logged (BD log, quality metrics)

## Sign-Off
Agent: Devin
Task: Implement user authentication
Classification: feature
Workflow: reflective-loop-beads.yaml
Duration: 2.5 hours

Post-Flight Checklist Score: 27/27

Deliverables Status: ✅ All complete

Quality Status: ✅ All checks pass

Decision: ✅ Task complete
```

---

## Reference

- **Task Classification**: content/rules/task-classification.md
- **Pre-Flight Checklist**: content/rules/pre-flight-checklist.md
- **Quality Principles**: content/rules/quality/mandatory-quality-principles.md
- **Quality Gates**: content/rules/quality/mandatory-quality-gates.md
- **Workflows**: content/workflows/INDEX.md
- **Skills**: content/skills/
- **Beads Protocol**: Z:\Ti\taskboard\beads.md
