# Mandatory Workflow Usage

> **Category**: Quality Rules
> **Priority**: P0 (BẮT BUỘC)
> **Last Updated**: 2026-05-12

---

## BẮT BUỘC Workflow Usage

### 1. BẮT BUỘC Sử Dụng Smart Workflow Selector

**Khi**: Cho tất cả tasks trong Ti workspace

**Yêu cầu**:
1. **BẮT BUỘC sử dụng `smart-workflow-selector.yaml`** cho tất cả tasks
2. Workflow sẽ tự động phân loại task complexity và select appropriate workflow
3. Không skip workflow trừ khi task cực kỳ đơn giản (< 5 phút, < 2 tool calls)
4. Workflow selection PHẢI được log với BD tool

**Exception**:
- Task cực kỳ đơn giản (< 5 phút, < 2 tool calls) → Direct execution với log reason
- User explicitly requests không dùng workflow → Log reason với BD tool

---

### 2. Smart Workflow Selector Benefits

**Tại sao phải dùng `smart-workflow-selector.yaml`:**

#### **🎯 Intelligent Selection**
- **Auto-classification**: Tự động phân loại task complexity
- **Optimal routing**: Luôn chọn workflow phù hợp nhất
- **No overhead**: Task đơn giản → direct execution
- **Full coverage**: Mọi loại task đều được xử lý

#### **📊 Performance Optimization**
- **Complexity-based routing**: Task phức tạp → workflow mạnh hơn
- **Cost-aware**: Xem xét cost limits và budget
- **Quality gates**: Áp dụng quality gates phù hợp
- **Learning**: Cải thiện theo thời gian với performance data

#### **🔄 Adaptive Learning**
- **Performance tracking**: Theo dõi hiệu suất từng workflow
- **Pattern recognition**: Nhận dạng task patterns
- **Optimization suggestions**: Đề xuất cải tiến
- **Continuous improvement**: Cải tiến liên tục

---

### 3. Workflow Selection Process

#### **Step 1: Task Analysis**
```
Input: Task name + description
Process: Complexity analysis
Output: Complexity classification (simple/standard/complex/critical)
```

#### **Step 2: Workflow Selection**
```
Simple → Direct execution
Standard → ti-cli-v2.yaml
Complex → reflective-loop-v2.yaml
Critical → reflective-loop-beads.yaml
```

#### **Step 3: Execution**
```
Execute selected workflow
Monitor performance
Log results with BD tool
```

---

### 4. Usage Examples

#### **Example 1: Simple Task**
```
Task: "Fix typo in README.md"
Smart workflow selector → Direct execution
Log: "Smart workflow selector: Direct execution (simple task)"
```

#### **Example 2: Standard Task**
```
Task: "Implement user authentication feature"
Smart workflow selector → ti-cli-v2.yaml
Log: "Smart workflow selector: ti-cli-v2.yaml (standard task)"
```

#### **Example 3: Complex Task**
```
Task: "Refactor payment system architecture"
Smart workflow selector → reflective-loop-v2.yaml
Log: "Smart workflow selector: reflective-loop-v2.yaml (complex task)"
```

#### **Example 4: Critical Task**
```
Task: "Database migration for production"
Smart workflow selector → reflective-loop-beads.yaml
Log: "Smart workflow selector: reflective-loop-beads.yaml (critical task)"
```

---

### 5. Quality Requirements

#### **🔍 Pre-Execution**
- Validate task description completeness
- Check for special requirements
- Verify resource availability

#### **⚡ Execution**
- Monitor workflow execution
- Track performance metrics
- Log checkpoints with BD tool

#### **✅ Post-Execution**
- Verify task completion
- Validate quality gates
- Log completion with BD tool

---

### 6. Error Handling

#### **🚨 Smart Selector Fails**
```
Fallback: Direct workflow selection
Log: "Smart workflow selector failed - using direct selection"
Notify: User notification required
```

#### **🔄 Incorrect Classification**
```
Detect: Performance metrics indicate wrong workflow
Action: Switch to appropriate workflow
Log: "Workflow correction: {old} → {new}"
Learn: Update classification patterns
```

#### **⚠️ Quality Gates Failed**
```
Detect: Quality gates not passed
Action: Retry with enhanced workflow
Log: "Quality gates failed - escalating workflow"
Notify: User intervention required
```

---

### 7. Performance Monitoring

#### **📊 Metrics to Track**
- **Selection accuracy**: Correct workflow selection rate
- **Task completion rate**: Success rate by workflow
- **Execution time**: Time to complete by workflow
- **Quality score**: Quality metrics by workflow
- **User satisfaction**: User feedback on selection

#### **🎯 Optimization Targets**
- Selection accuracy: > 90%
- Task completion rate: > 95%
- Execution time: Within estimated limits
- Quality score: > 4.0/5.0
- User satisfaction: > 4.0/5.0

---

### 8. Integration with BD Tool

#### **📝 Logging Requirements**
```bash
# Task start
bd.exe log --task="Starting smart workflow selector: {task_name}" --agent=orchestrator --status=running

# Workflow selection
bd.exe log --task="Smart workflow selector: {workflow} ({complexity})" --agent=orchestrator --status=selected

# Task completion
bd.exe log --task="Smart workflow selector completed: {task_name}" --agent=orchestrator --status=completed
```

#### **🔍 Monitoring Requirements**
```bash
# Performance metrics
bd.exe metrics --workflow={workflow} --time={execution_time} --success={success}

# Quality metrics
bd.exe quality --workflow={workflow} --score={quality_score} --gates={gates_passed}
```

---

### 9. Best Practices

#### **✅ DO**
- Always use `smart-workflow-selector.yaml` as default
- Provide clear task descriptions
- Log all workflow selections
- Monitor performance metrics
- Update classification patterns based on results

#### **❌ DON'T**
- Skip workflow without valid reason
- Use manual selection without justification
- Ignore performance metrics
- Forget to log with BD tool
- Override smart selector without good reason

---

### 10. Troubleshooting

#### **Common Issues**

**Issue**: Smart selector takes too long
```
Solution: Check task description clarity
Fallback: Use direct selection with timeout
```

**Issue**: Wrong workflow selected
```
Solution: Review task description complexity
Fallback: Manually select correct workflow
```

**Issue**: Workflow execution fails
```
Solution: Check resource availability
Fallback: Retry with simpler workflow
```

#### **Support Resources**
- **Documentation**: `smart-workflow-selector.yaml`
- **Troubleshooting**: `WORKFLOW_RENAMING_GUIDE.md`
- **Performance**: `quality-metrics-dashboard.md`
- **Support**: Ti Workflow Team

---

## 📋 Compliance Checklist

### Before Starting Task:
- [ ] Task description is clear and complete
- [ ] Smart workflow selector is available
- [ ] BD tool is configured
- [ ] Resource requirements are met

### During Execution:
- [ ] Workflow selection is logged
- [ ] Performance is monitored
- [ ] Checkpoints are logged
- [ ] Quality gates are applied

### After Completion:
- [ ] Task completion is logged
- [ ] Quality metrics are recorded
- [ ] Performance is analyzed
- [ ] Lessons learned are documented

---

*Last Updated: 2026-05-12*