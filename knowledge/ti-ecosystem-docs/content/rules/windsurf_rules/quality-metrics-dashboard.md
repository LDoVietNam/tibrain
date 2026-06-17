# Quality Metrics Dashboard

> **Category**: Quality Rules  
> **Priority**: P1 (HIGH)  
> **Last Updated**: 2026-06-05  

---

## BẮT BUỘC Quality Metrics Tracking

**Mục đích**: Track và report quality metrics qua thời gian để identify trends và improvement opportunities

---

## Metrics Categories

### 1. Task Completion Metrics

| Metric | Description | Target | Calculation |
|--------|-------------|--------|-------------|
| **Task Success Rate** | % tasks completed successfully | ≥95% | (successful tasks / total tasks) × 100 |
| **Task Completion Time** | Average time to complete task | ≤2 hours | sum(completion times) / count(tasks) |
| **Task Classification Accuracy** | % correct classifications | ≥90% | (correct classifications / total classifications) × 100 |
| **Pre-Flight Compliance** | % tasks with complete pre-flight | ≥95% | (complete pre-flights / total tasks) × 100 |
| **Post-Flight Compliance** | % tasks with complete post-flight | ≥95% | (complete post-flights / total tasks) × 100 |

### 2. Code Quality Metrics

| Metric | Description | Target | Calculation |
|--------|-------------|--------|-------------|
| **Test Coverage** | % code covered by tests | ≥80% | (lines covered / total lines) × 100 |
| **Build Success Rate** | % builds that succeed | ≥95% | (successful builds / total builds) × 100 |
| **Linting Compliance** | % files passing linting | ≥95% | (passing files / total files) × 100 |
| **Type Check Success** | % type checks that pass | ≥95% | (passing type checks / total type checks) × 100 |
| **Security Scan Pass Rate** | % security scans passing | ≥98% | (passing scans / total scans) × 100 |

### 3. Workflow Metrics

| Metric | Description | Target | Calculation |
|--------|-------------|--------|-------------|
| **Workflow Selection Accuracy** | % correct workflow selections | ≥90% | (correct workflows / total selections) × 100 |
| **Workflow Compliance** | % tasks following workflow steps | ≥95% | (compliant tasks / total tasks) × 100 |
| **Skill Invocation Rate** | % tasks with correct skill invocation | ≥90% | (correct skill invocations / total invocations) × 100 |
| **Beads Logging Rate** | % tasks with beads logged | ≥95% | (tasks with beads / total tasks) × 100 |

### 4. Quality Gate Metrics

| Metric | Description | Target | Calculation |
|--------|-------------|--------|-------------|
| **Gate Pass Rate** | % quality gates passing | ≥95% | (passing gates / total gates) × 100 |
| **Gate Failure Recovery** | % failed gates recovered | ≥80% | (recovered failures / total failures) × 100 |
| **Gate Escalation Rate** | % critical escalations | ≤5% | (escalations / total tasks) × 100 |

### 5. Violation Metrics

| Metric | Description | Target | Calculation |
|--------|-------------|--------|-------------|
| **Pre-Flight Violation Rate** | % tasks with pre-flight violations | ≤5% | (violations / total tasks) × 100 |
| **Post-Flight Violation Rate** | % tasks with post-flight violations | ≤5% | (violations / total tasks) × 100 |
| **Repeated Violation Rate** | % repeated violations | ≤2% | (repeated violations / total violations) × 100 |
| **Critical Violation Rate** | % critical violations | ≤1% | (critical violations / total violations) × 100 |

---

## Metrics Collection

### Automatic Collection (via BD Tool)

```bash
# Log task completion with metrics
bd log --task="Task completed" \
       --agent=devin \
       --status=complete \
       --type=coding \
       --domain=general \
       --metrics='{"duration": "1.5h", "test_coverage": "85%", "build_success": true}'
```

### Manual Collection (via Post-Flight Checklist)

Post-flight checklist captures:
- Test coverage percentage
- Build success/failure
- Linting status
- Type check status
- Security scan status

### Weekly Summary

```bash
# Generate weekly metrics summary
bd summary --period=week --format=json
```

---

## Dashboard Structure

### Daily View

```yaml
date: 2026-06-05
task_completion:
  total_tasks: 10
  successful: 9
  failed: 1
  success_rate: 90%
  avg_completion_time: 1.8h

code_quality:
  avg_test_coverage: 82%
  build_success_rate: 95%
  linting_compliance: 93%
  type_check_success: 97%
  security_scan_pass_rate: 99%

workflow_metrics:
  workflow_selection_accuracy: 92%
  workflow_compliance: 96%
  skill_invocation_rate: 88%
  beads_logging_rate: 100%

quality_gates:
  gate_pass_rate: 94%
  gate_failure_recovery: 75%
  gate_escalation_rate: 4%

violations:
  pre_flight_violation_rate: 6%
  post_flight_violation_rate: 4%
  repeated_violation_rate: 1%
  critical_violation_rate: 0%

trends:
  task_completion: ↗ improving
  code_quality: ↗ improving
  workflow_metrics: → stable
  quality_gates: ↗ improving
  violations: ↘ decreasing
```

### Weekly View

```yaml
week: 2026-W23
task_completion:
  total_tasks: 50
  successful: 47
  failed: 3
  success_rate: 94%
  avg_completion_time: 1.7h

code_quality:
  avg_test_coverage: 81%
  build_success_rate: 96%
  linting_compliance: 94%
  type_check_success: 96%
  security_scan_pass_rate: 98%

workflow_metrics:
  workflow_selection_accuracy: 91%
  workflow_compliance: 95%
  skill_invocation_rate: 89%
  beads_logging_rate: 99%

quality_gates:
  gate_pass_rate: 93%
  gate_failure_recovery: 78%
  gate_escalation_rate: 3%

violations:
  pre_flight_violation_rate: 5%
  post_flight_violation_rate: 3%
  repeated_violation_rate: 2%
  critical_violation_rate: 1%

trends:
  task_completion: ↗ improving
  code_quality: ↗ improving
  workflow_metrics: → stable
  quality_gates: ↗ improving
  violations: ↘ decreasing

improvements:
  - Reduced pre-flight violations by 2%
  - Increased test coverage by 3%
  - Improved workflow compliance by 1%

concerns:
  - Skill invocation rate still below target (89% vs 90%)
  - Gate failure recovery below target (78% vs 80%)
```

### Monthly View

```yaml
month: 2026-06
task_completion:
  total_tasks: 200
  successful: 192
  failed: 8
  success_rate: 96%
  avg_completion_time: 1.6h

code_quality:
  avg_test_coverage: 83%
  build_success_rate: 97%
  linting_compliance: 95%
  type_check_success: 97%
  security_scan_pass_rate: 99%

workflow_metrics:
  workflow_selection_accuracy: 93%
  workflow_compliance: 96%
  skill_invocation_rate: 91%
  beads_logging_rate: 100%

quality_gates:
  gate_pass_rate: 95%
  gate_failure_recovery: 82%
  gate_escalation_rate: 2%

violations:
  pre_flight_violation_rate: 3%
  post_flight_violation_rate: 2%
  repeated_violation_rate: 1%
  critical_violation_rate: 0%

trends:
  task_completion: ↗ improving
  code_quality: ↗ improving
  workflow_metrics: ↗ improving
  quality_gates: ↗ improving
  violations: ↘ decreasing

improvements:
  - Increased task success rate by 1%
  - Increased test coverage by 2%
  - Reduced pre-flight violations by 2%
  - Improved skill invocation rate by 2%

concerns:
  - None

next_month_goals:
  - Achieve 97% task success rate
  - Achieve 85% test coverage
  - Achieve 92% skill invocation rate
  - Achieve 96% workflow compliance
```

---

## Alerting

### Alert Thresholds

| Metric | Warning | Critical | Action |
|--------|---------|----------|--------|
| Task Success Rate | <90% | <80% | Escalate to human review |
| Test Coverage | <70% | <50% | Require improvement plan |
| Build Success Rate | <90% | <80% | Investigate build failures |
| Pre-Flight Violation Rate | >10% | >20% | Review training needs |
| Post-Flight Violation Rate | >10% | >20% | Review workflow compliance |
| Critical Violation Rate | >2% | >5% | Immediate escalation |

### Alert Actions

```bash
# Warning alert
bd alert --level=warning \
         --metric="test_coverage" \
         --value="65%" \
         --target="80%" \
         --message="Test coverage below target"

# Critical alert
bd alert --level=critical \
         --metric="task_success_rate" \
         --value="75%" \
         --target="95%" \
         --message="Task success rate critical - escalate to human"
```

---

## Reporting

### Daily Report

```bash
bd report --period=day --format=markdown
```

### Weekly Report

```bash
bd report --period=week --format=markdown
```

### Monthly Report

```bash
bd report --period=month --format=markdown
```

### Custom Report

```bash
bd report --start=2026-06-01 --end=2026-06-30 --format=json
```

---

## Enforcement

### BẮT BUỘC

- **BẮT BUỘC** log metrics với BD tool sau mỗi task
- **BẮT BUỘC** review metrics weekly
- **BẮT BUỘC** address alerts within 24 hours

### Violation Detection

- Skip metrics logging → Log violation
- Ignore alerts → Log violation
- Miss weekly review → Log violation

### Violation Logging

```bash
bd log --task="Metrics violation: {violation_type} - {details}" \
       --agent=devin \
       --status=failed \
       --type=violation \
       --domain=general
```

### Escalation

- Repeated violations (≥3 times) → Escalate to human review
- Critical violations (ignore critical alerts) → Immediate escalation

---

## Reference

- **Task Classification**: content/rules/task-classification.md
- **Pre-Flight Checklist**: content/rules/pre-flight-checklist.md
- **Post-Flight Checklist**: content/rules/post-flight-checklist.md
- **Quality Principles**: content/rules/quality/mandatory-quality-principles.md
- **Quality Gates**: content/rules/quality/mandatory-quality-gates.md
- **Workflows**: content/workflows/INDEX.md
- **Skills**: content/skills/
- **Beads Protocol**: Z:\Ti\taskboard\beads.md
