# Mandatory Quality Principles

> **Category**: Quality Rules
> **Priority**: P0 (BẮT BUỘC)
> **Last Updated**: 2026-05-05

---

## BẮT BUỘC Quality Principles

### 1. Context Collection (BẮT BUỘC)

**Khi**: Trước khi làm bất kỳ decision

**Yêu cầu**:
1. **BẮT BUỘC thu thập context** trước khi làm decision
2. Scan project structure (files, directories)
3. Extract code patterns (imports, naming conventions, code style)
4. Load project rules (AGENTS.md, .editorconfig, linting config)
5. Check git history (last 10 commits)
6. Analyze dependencies (go.mod, package.json, etc.)

**Output**: `.devin/context.json` với:
- project_structure
- patterns (imports, naming, style)
- rules (agents_md, editorconfig, linting)
- history (recent_commits)
- dependencies

---

### 2. Decision Heuristics (BẮT BUỘC)

**Khi**: Khi làm bất kỳ decision (library, pattern, architecture, etc.)

**Yêu cầu**:
1. **BẮT BUỘC apply Decision Heuristics** (Assume First, Ask Second)
2. Score confidence (0-100%) cho mỗi decision
3. Use decision thresholds:

| Category | Assume Threshold | Ask Threshold | Examples |
|----------|-----------------|---------------|----------|
| Style | 80% | 60% | Indentation, naming, imports |
| Architecture | 70% | 50% | Layer choice, pattern selection |
| Security | 90% | 70% | Auth methods, encryption |
| Performance | 75% | 55% | Caching, async patterns |
| Testing | 70% | 50% | Test framework choice |

**Decision Logic**:
- Confidence ≥ threshold → Assume without asking
- Confidence < threshold → Ask user
- Critical decisions → Ask user (regardless of confidence)

---

### 3. Assumption Documentation (BẮT BUỘC)

**Khi**: Khi making assumption (giả định hợp lý)

**Yêu cầu**:
1. **BẮT BUỘC document assumptions** với confidence score
2. Document:
   - What was assumed
   - Why (reasoning)
   - Confidence level (0-100%)
   - Context used (patterns, rules, history)

**Format**:
```
Assumption: {assumption_description}
Reason: {reasoning}
Confidence: {confidence}%
Context: {context_source}
```

---

### 4. Assumption Validation (BẮT BUỘC)

**Khi**: Sau khi execute assumption

**Yêu cầu**:
1. **BẮT BUỘC validate assumptions** sau execution
2. Kiểm tra test results, build results, execution output
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

### 5. Confidence Metrics Logging (BẮT BUỘC)

**Khi**: Sau khi hoàn thành task

**Yêu cầu**:
1. **BẮT BUỘC log confidence metrics** với BD tool
2. Log metrics:
   - Average confidence score
   - Low confidence decisions count (< 70%)
   - High confidence decisions count (≥ 90%)
   - Assumption success rate
   - Assumption failure count
   - Rollback count

**Log Command**:
```bash
"Z:/02_CORE/_cli/bin/bd.exe" log \
  --task="Confidence Metrics: avg={avg_confidence}%, low={low_count}, high={high_count}, success_rate={success_rate}%, failures={failure_count}, rollbacks={rollback_count}" \
  --agent=workflow-orchestrator \
  --status=complete \
  --type=metrics \
  --domain=cli
```

---

## Enforcement

**Violation Detection**:
- Skip Context Collection → Log violation
- Skip Decision Heuristics → Log violation
- Skip Assumption Documentation → Log violation
- Skip Assumption Validation → Log violation
- Skip Confidence Metrics Logging → Log violation

**Violation Logging**:
```bash
"Z:/02_CORE/_cli/bin/bd.exe" log \
  --task="Quality violation: {violation_type} - {details}" \
  --agent=workflow-orchestrator \
  --status=failed \
  --type=violation \
  --domain=cli
```

**Escalation**:
- Repeated violations (≥3 times) → Escalate to human review
- Critical violations (skip security principles) → Immediate escalation

---

## Reference

- **Quality Decision Making Principles**: AGENTS.md → Quality Decision Making Principles section
- **SUB_AGENT_GUIDE.md**: Ti-learning-lab/03_Knowledge/CLI/SUB_AGENT_GUIDE.md → Quality Enhancement section
- **Workflow Quality Features**: reflective-loop.yaml v2.0, ti-cli-development.yaml v2.0
