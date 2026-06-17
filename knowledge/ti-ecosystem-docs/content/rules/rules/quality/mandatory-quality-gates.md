# Mandatory Quality Gates

> **Category**: Quality Rules
> **Priority**: P0 (BẮT BUỘC)
> **Last Updated**: 2026-05-05

---

## BẮT BUỘC Quality Gates

### 1. Assumption Verification (BẮT BUỘC)

**Khi**: Sau khi implement assumption

**Yêu cầu**:
1. **BẮT BUỘC verify assumptions** sau implementation
2. Kiểm tra:
   - Test results
   - Build results
   - Execution output
3. Determine if assumption was correct

**If Assumption Correct**:
- Update pattern database (increase confidence +5%)
- Log: "Assumption Validated: {assumption} (confidence: {confidence}%)"

**If Assumption Failed**:
- Rollback action (nếu có thể)
- Update pattern database (decrease confidence -10%)
- Log: "Assumption Failed: {assumption} (confidence: {confidence}%) - Rolling back"
- Ask user for guidance (nếu cần)

---

### 2. Rollback on Failure (BẮT BUỘC)

**Khi**: Khi assumption failed hoặc implementation failed

**Yêu cầu**:
1. **BẮT BUỘC rollback** nếu assumptions wrong hoặc implementation failed
2. Rollback methods:
   - Git revert (git checkout, git reset)
   - File restore (backup)
   - Database rollback (migration rollback)
3. Log rollback action với BD tool

**Log Command**:
```bash
"Z:/02_CORE/_cli/bin/bd.exe" log \
  --task="Rollback executed: {rollback_type} - {reason}" \
  --agent=workflow-orchestrator \
  --status=running \
  --type=rollback \
  --domain=cli
```

---

### 3. Pattern Database Update (BẮT BUỘC)

**Khi**: Sau khi validate assumption

**Yêu cầu**:
1. **BẮT BUỘC update pattern database** (increase/decrease confidence)
2. Update confidence score cho assumption pattern
3. Update frequency:
   - Correct assumption: +5%
   - Failed assumption: -10%

**Pattern Database Location**:
- `.devin/knowledge/patterns.json` (hoặc equivalent)
- Format:
  ```json
  {
    "patterns": [
      {
        "pattern": "use logrus for logging",
        "confidence": 95,
        "usage_count": 15,
        "success_count": 14
      }
    ]
  }
  ```

---

### 4. Lessons Learned Logging (BẮT BUỘC)

**Khi**: Sau khi hoàn thành task

**Yêu cầu**:
1. **BẮT BUỘC log lessons learned** với BD tool
2. Log:
   - What worked well
   - What didn't work
   - Patterns discovered
   - Assumptions validated/failed
   - Recommendations for future

**Log Command**:
```bash
"Z:/02_CORE/_cli/bin/bd.exe" log \
  --task="Lessons Learned: {lessons}" \
  --agent=workflow-orchestrator \
  --status=complete \
  --type=lesson \
  --domain=cli
```

---

## Quality Gate Checklist

### Pre-Implementation
- [ ] Context collected
- [ ] Patterns extracted
- [ ] Rules loaded
- [ ] Decision heuristics applied
- [ ] Assumptions documented

### Post-Implementation
- [ ] Assumptions validated
- [ ] Rollback executed (if failed)
- [ ] Pattern database updated
- [ ] Lessons learned logged
- [ ] Confidence metrics logged

---

## Enforcement

**Violation Detection**:
- Skip assumption verification → Log violation
- Skip rollback on failure → Log violation
- Skip pattern database update → Log violation
- Skip lessons learned logging → Log violation

**Violation Logging**:
```bash
"Z:/02_CORE/_cli/bin/bd.exe" log \
  --task="Quality gate violation: {gate_type} - {details}" \
  --agent=workflow-orchestrator \
  --status=failed \
  --type=violation \
  --domain=cli
```

**Escalation**:
- Repeated violations (≥3 times) → Escalate to human review
- Critical violations (skip rollback on critical failure) → Immediate escalation

---

## Reference

- **Quality Decision Making Principles**: AGENTS.md → Quality Decision Making Principles section
- **SUB_AGENT_GUIDE.md**: Ti-learning-lab/03_Knowledge/CLI/SUB_AGENT_GUIDE.md → Quality Enhancement section
- **Workflow Quality Features**: reflective-loop.yaml v2.0, ti-cli-development.yaml v2.0
