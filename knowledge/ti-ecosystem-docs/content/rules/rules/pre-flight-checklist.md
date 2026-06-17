# Pre-Flight Checklist

> **Category**: Quality Rules  
> **Priority**: P0 (BẮT BUỘC)  
> **Last Updated**: 2026-06-05  

---

## BẮT BUỘC Pre-Flight Checklist

**Mục đích**: BẮT BUỘC hoàn thành checklist này trước khi bắt đầu task để đảm bảo chuẩn bị đầy đủ

---

## Checklist

### 1. Task Understanding (BẮT BUỘC)

- [ ] **Tôi hiểu rõ task requirements**
  - Task description rõ ràng
  - Deliverables đã xác định
  - Acceptance criteria đã biết
  - Deadline đã biết (nếu có)

- [ ] **Tôi biết scope của task**
  - In-scope: những gì cần làm
  - Out-of-scope: những gì KHÔNG cần làm
  - Dependencies đã identify
  - Blockers đã identify

### 2. Context Collection (BẮT BUỘC)

- [ ] **Đã đọc AGENTS.md**
  - Project structure đã hiểu
  - Build commands đã biết
  - Workflow đã hiểu
  - Secrets location đã biết

- [ ] **Đã đọc project rules (content/rules/)**
  - Core workflow rules đã đọc
  - Quality rules đã đọc
  - Task classification đã đọc
  - Pre/post-flight checklist đã đọc

- [ ] **Đã scan project structure**
  - Directories đã mapped
  - Files đã identified
  - Entry points đã biết
  - Dependencies đã understand

- [ ] **Đã extract code patterns**
  - Naming conventions đã note
  - Code style đã understand
  - Import patterns đã note
  - Architecture patterns đã note

- [ ] **Đã check git history**
  - Last 10 commits đã review
  - Recent changes đã understand
  - Branching strategy đã know
  - Merge conflicts đã anticipate

- [ ] **Đã analyze dependencies**
  - go.mod/package.json đã review
  - External dependencies đã list
  - Version conflicts đã check
  - Security vulnerabilities đã check

### 3. Risk Assessment (BẮT BUỘC)

- [ ] **Đã identify potential blockers**
  - Technical blockers đã list
  - Knowledge gaps đã identify
  - Resource constraints đã note
  - External dependencies đã check

- [ ] **Đã plan mitigation strategies**
  - Blocker 1: {mitigation}
  - Blocker 2: {mitigation}
  - Blocker 3: {mitigation}

- [ ] **Đã estimate effort**
  - Estimated time: {time}
  - Complexity level: {low/medium/high}
  - Confidence: {confidence}%

- [ ] **Đã identify dependencies**
  - Task dependencies: {list}
  - Team dependencies: {list}
  - External dependencies: {list}

### 4. Tool/Skill Selection (BẮT BUỘC)

- [ ] **Đã classify task type**
  - Task type: {type}
  - Classification method: {auto/manual}
  - Confidence: {confidence}%

- [ ] **Đã select appropriate workflow**
  - Workflow: {workflow}
  - Reason: {reason}
  - Priority: {priority}

- [ ] **Đã load required skills**
  - Mandatory skills: {list}
  - Language-specific skills: {list}
  - Optional skills: {list}

- [ ] **Đã verify tools availability**
  - Required tools: {list}
  - Tool availability: {status}
  - Alternative tools: {list}

### 5. Quality Planning (BẮT BUỘC)

- [ ] **Đã plan testing strategy**
  - Unit tests: {plan}
  - Integration tests: {plan}
  - E2E tests: {plan}
  - Coverage target: {target}

- [ ] **Đã plan review process**
  - Self-review: {plan}
  - Code review: {plan}
  - Review criteria: {criteria}

- [ ] **Đã plan documentation requirements**
  - Code comments: {plan}
  - README updates: {plan}
  - API docs: {plan}
  - Changelog: {plan}

- [ ] **Đã set quality metrics**
  - Success criteria: {criteria}
  - Performance metrics: {metrics}
  - Security checks: {checks}

---

## Sign-Off

**Agent**: Devin  
**Task**: {task_description}  
**Classification**: {task_type}  
**Workflow**: {workflow}  
**Auto-Skills**: {skills}  
**Estimated Effort**: {effort}  
**Risk Level**: {risk_level}

**Pre-Flight Checklist Score**: {score}/5

**Decision**: ✅ **Ready to proceed** / ❌ **Need clarification**

**If Need Clarification:**
- Unclear points: {list}
- Questions for user: {list}
- Blockers: {list}

---

## Enforcement

### BẮT BUỘC

- **BẮT BUỘC** hoàn thành checklist trước khi bắt đầu task
- **BẮT BUỘC** log sign-off với BD tool
- **BẮT BUỘC** ask clarification nếu có uncertain points

### Violation Detection

- Skip checklist → Log violation
- Incomplete checklist → Log violation
- Proceed without sign-off → Log violation

### Violation Logging

```bash
bd log --task="Pre-flight violation: {violation_type} - {details}" \
       --agent=devin \
       --status=failed \
       --type=violation \
       --domain=general
```

### Escalation

- Repeated violations (≥3 times) → Escalate to human review
- Critical violations (skip context collection) → Immediate escalation

---

## Quick Reference

**Pre-Flight Checklist Summary:**
1. Task Understanding (5 items)
2. Context Collection (6 items)
3. Risk Assessment (4 items)
4. Tool/Skill Selection (4 items)
5. Quality Planning (4 items)

**Total:** 23 items

**Minimum to proceed:** 20/23 items (87%)
**Recommended:** 23/23 items (100%)

---

## Example

```markdown
## Pre-Flight: Implement user authentication

### 1. Task Understanding
✅ Task: Implement user authentication with JWT
✅ Deliverables: Auth service, JWT tokens, login/logout endpoints
✅ Acceptance: Users can login, tokens are validated, logout works
✅ Deadline: None

### 2. Context Collection
✅ AGENTS.md read - CLI structure understood
✅ Rules read - Quality principles understood
✅ Project scanned - apps/auth/ identified
✅ Patterns extracted - Go patterns noted
✅ Git history checked - Last 5 commits reviewed
✅ Dependencies analyzed - go.mod checked

### 3. Risk Assessment
✅ Blockers: None identified
✅ Mitigation: N/A
✅ Effort: 2-3 hours
✅ Dependencies: JWT library needed

### 4. Tool/Skill Selection
✅ Classification: feature
✅ Workflow: reflective-loop-beads.yaml
✅ Skills: tdd-workflow, go-patterns, backend-patterns
✅ Tools: Go compiler, JWT library

### 5. Quality Planning
✅ Testing: Unit tests for auth service, integration tests for endpoints
✅ Review: 5-axis review
✅ Documentation: API docs, README update
✅ Metrics: Test coverage ≥80%, build success

## Sign-Off
Agent: Devin
Task: Implement user authentication
Classification: feature
Workflow: reflective-loop-beads.yaml
Auto-Skills: tdd-workflow, go-patterns, backend-patterns
Estimated Effort: 2-3 hours
Risk Level: Medium

Pre-Flight Checklist Score: 23/23

Decision: ✅ Ready to proceed
```

---

## Reference

- **Task Classification**: content/rules/task-classification.md
- **Quality Principles**: content/rules/quality/mandatory-quality-principles.md
- **Quality Gates**: content/rules/quality/mandatory-quality-gates.md
- **Workflows**: content/workflows/INDEX.md
- **Skills**: content/skills/
- **Beads Protocol**: Z:\Ti\taskboard\beads.md
