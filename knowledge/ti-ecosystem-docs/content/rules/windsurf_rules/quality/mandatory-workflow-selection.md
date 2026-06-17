# Mandatory Workflow Selection

> **Category**: Quality Rules
> **Priority**: P0 (BẮT BUỘC)
> **Last Updated**: 2026-05-12

---

## BẮT BUỘC Workflow Selection

### 🎯 RECOMMENDED: Smart Workflow Selector

**Primary Approach**: `smart-workflow-selector.yaml`
- **Description**: Intelligent workflow selection based on task complexity
- **Use Case**: **DEFAULT workflow** cho mọi tasks - tự động phân loại và delegate
- **Benefits**: No overhead for simple tasks, full quality for complex tasks

### Workflow Selection Criteria

**Simple Tasks** (< 15min, < 3 tool calls):
- `smart-workflow-selector.yaml` → Direct execution (no workflow overhead)
- Examples: Fix syntax error, add single function, update documentation
- Log reason: "Smart workflow selector: Direct execution (simple task < 15min, < 3 tool calls)"

**Standard Tasks** (< 2hours, < 20 tool calls):
- `smart-workflow-selector.yaml` → `ti-cli-v2.yaml` (enhanced standard workflow)
- Examples: Implement feature, refactor module, add tests
- Quality features: Context Collection, Decision Heuristics, Assumption Validation

**Complex Tasks** (< 1 day, < 100 tool calls):
- `smart-workflow-selector.yaml` → `reflective-loop-v2.yaml` (enhanced reflective loop)
- Examples: Architecture change, multi-module refactor, complex feature
- Quality features: Context Collection, Decision Heuristics, Confidence-based HITL, Progressive Disclosure

**Critical Tasks** (unlimited, high-stakes):
- `smart-workflow-selector.yaml` → `reflective-loop-beads.yaml` (ultimate workflow)
- Examples: Database migration, security audit, production deployment
- Quality features: All v2.0 features + Beads integration + HITL

### Special Cases

**High-stakes security**:
- → `smart-workflow-selector.yaml` → `reflective-loop-beads.yaml` (regardless of complexity)
- Reason: Security requires full audit trail

**Architecture change**:
- → `smart-workflow-selector.yaml` → `reflective-loop-v2.yaml` (or `reflective-loop-beads.yaml` if critical)
- Reason: Architecture changes need cross-verification

**Database migration**:
- → `smart-workflow-selector.yaml` → `reflective-loop-beads.yaml` (always critical)
- Reason: Database migration is always critical

**Production deployment**:
- → `smart-workflow-selector.yaml` → `reflective-loop-beads.yaml` (always critical)
- Reason: Production deployment requires HITL and full audit trail

### 🔄 Alternative: Direct Workflow Selection

**When to use direct selection**:
- User explicitly requests specific workflow
- Smart workflow selector fails or unavailable
- Known task type with predictable complexity

**Direct Selection Mapping**:
- **Simple**: Direct execution (no workflow)
- **Standard**: `ti-cli-v2.yaml`
- **Complex**: `reflective-loop-v2.yaml`
- **Critical**: `reflective-loop-beads.yaml`

---

## 📋 Implementation Requirements

### 1. BẮT BUỘC Logging
- All workflow selections PHẢI được log với BD tool
- Include reasoning for selection
- Track performance metrics

### 2. Performance Monitoring
- Monitor smart workflow selector accuracy
- Track task completion rates
- Adjust complexity thresholds if needed

### 3. Quality Assurance
- Validate workflow selection matches task complexity
- Ensure quality gates are applied appropriately
- Monitor for workflow selection errors

---

## 🚀 Benefits of Smart Workflow Selector

### **Advantages:**
- **Auto-selection**: No manual decision required
- **Optimal routing**: Always selects best workflow
- **No overhead**: Simple tasks get direct execution
- **Full coverage**: All task types handled
- **Learning**: Improves over time with performance data

### **When to Override:**
- User explicitly requests different workflow
- Task has special requirements not captured by complexity
- Smart selector makes incorrect classification

---

## 📊 Success Metrics

### **Selection Accuracy:**
- Correct workflow selection rate: > 90%
- Task completion rate by workflow: > 95%
- User satisfaction with selection: > 4.0/5.0

### **Performance Metrics:**
- Selection time: < 5 seconds
- Overhead for simple tasks: < 10%
- Quality gate pass rate: > 95%

---

*Last Updated: 2026-05-12*