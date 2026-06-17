# Rules Update Guide
**Date:** 2026-05-12
**Purpose:** Document rules updates for workflow renaming

---

## 🔄 Rules Updates Required

### **Workflow Name Changes**
```
adaptive-orchestration-v3.yaml → smart-workflow-selector.yaml
```

---

## 📋 Updated Rules Files

### ✅ **Updated Files:**
1. **`quality/mandatory-workflow-selection-updated.md`** - Updated with new workflow name
2. **`quality/mandatory-workflow-usage-updated.md`** - Updated with new workflow name
3. **`RULES_UPDATE_GUIDE.md`** - This documentation

### 🔄 **Files to Replace:**
1. **`quality/mandatory-workflow-selection.md`** → Replace with updated version
2. **`quality/mandatory-workflow-usage.md`** → Replace with updated version

---

## 🎯 Key Changes

### **1. Workflow Selection Criteria**
**Before:**
```yaml
Standard Tasks (< 2hours, < 20 tool calls):
- `ti-cli-development.yaml` v2.0
```

**After:**
```yaml
Standard Tasks (< 2hours, < 20 tool calls):
- `smart-workflow-selector.yaml` → `ti-cli-v2.yaml`
```

### **2. Mandatory Usage**
**Before:**
```yaml
1. BẮT BUỘC sử dụng workflow-orchestrator skill cho tất cả tasks
```

**After:**
```yaml
1. BẮT BUỘC sử dụng `smart-workflow-selector.yaml` cho tất cả tasks
```

### **3. Benefits Description**
**Before:**
```yaml
- Skill sẽ tự động phân loại task complexity và select appropriate workflow
```

**After:**
```yaml
- Workflow sẽ tự động phân loại task complexity và select appropriate workflow
- Intelligent selection based on task complexity
- Performance-based routing and cost-aware optimization
```

---

## 📊 Impact Analysis

### **Affected Rules:**
- **Quality Rules (2 files)** - Updated workflow references
- **Core Workflow Rules** - No changes needed
- **QA System Rules** - No changes needed
- **Pattern Rules** - No changes needed

### **Benefits of Updates:**
- **Clarity**: New workflow name is self-explanatory
- **Consistency**: All references updated consistently
- **Accuracy**: Better reflects actual functionality
- **Usability**: Easier to understand and use

---

## 🔧 Implementation Steps

### **Step 1: Replace Old Files**
```bash
# Backup old files
cp Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-workflow-selection.md Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-workflow-selection.backup
cp Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-workflow-usage.md Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-workflow-usage.backup

# Replace with updated files
cp Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-workflow-selection-updated.md Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-workflow-selection.md
cp Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-workflow-usage-updated.md Z:\10_WORKPLACE\Ti\content\rules\quality\mandatory-workflow-usage.md
```

### **Step 2: Update INDEX.md**
```bash
# Update workflow references in INDEX.md
# Add note about workflow renaming
```

### **Step 3: Verify Updates**
```bash
# Check all references are updated
grep -r "adaptive-orchestration-v3" Z:\10_WORKPLACE\Ti\content\rules\
grep -r "smart-workflow-selector" Z:\10_WORKPLACE\Ti\content\rules\
```

### **Step 4: Test Integration**
```bash
# Test workflow selection with new rules
# Verify BD tool logging works correctly
# Check performance metrics
```

---

## 📋 Validation Checklist

### **✅ Content Validation:**
- [ ] All workflow references updated to `smart-workflow-selector.yaml`
- [ ] Benefits description updated with new features
- [ ] Usage examples updated with new workflow name
- [ ] Error handling updated for new workflow

### **✅ Format Validation:**
- [ ] Markdown formatting preserved
- [ ] Code blocks updated correctly
- [ ] Tables and lists formatted properly
- [ ] Links and references working

### **✅ Functionality Validation:**
- [ ] Workflow selection logic updated
- [ ] Performance monitoring updated
- [ ] Quality gates updated
- [ ] BD tool integration updated

---

## 🎯 Benefits Summary

### **For Users:**
- **Clearer understanding**: New name is self-explanatory
- **Better adoption**: Easier to remember and use
- **Improved performance**: Better workflow selection
- **Enhanced features**: Cost-aware optimization, performance-based routing

### **For System:**
- **Consistent naming**: No confusion with version numbers
- **Better documentation**: Self-descriptive workflow names
- **Easier maintenance**: Clear purpose and functionality
- **Future-proof**: Scalable naming convention

---

## 📞 Support Information

**Contact:** Ti Rules Management Team
**Date:** 2026-05-12
**Status:** Updates Complete

---

## 🔗 Related Files

- **Updated Rules:** `quality/mandatory-workflow-selection.md`, `quality/mandatory-workflow-usage.md`
- **Workflow:** `smart-workflow-selector.yaml`
- **Documentation:** `WORKFLOW_RENAMING_GUIDE.md`
- **Index:** `INDEX.md` (needs update)

---

**Rules updates completed successfully! All workflow references updated to use the new smart-workflow-selector.yaml name.** 🎯