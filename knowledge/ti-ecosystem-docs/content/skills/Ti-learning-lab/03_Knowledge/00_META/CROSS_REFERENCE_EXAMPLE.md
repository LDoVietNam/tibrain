# Cross-Reference Example

> **Purpose**: Example showing how to create cross-references between stages

---

## 🎯 Why Cross-References?

Cross-references help:
- Navigate between related docs
- Trace knowledge flow: Learning → Research → Planning → Taskboard → HandsOn
- Understand context: "Why was this doc created?"
- Update related docs when knowledge changes

---

## 📋 Cross-Reference Template

### For Learning Docs

```markdown
## 🔗 Related Resources

### Research
- [02_Research/ProjectName/analysis.md](../../02_Research/ProjectName/analysis.md) - Analysis of [topic]

### Planning
- [04_Planning/ProjectName/enhancement-plan.md](../../04_Planning/ProjectName/enhancement-plan.md) - Plan for [feature]

### Knowledge
- [03_Knowledge/patterns/pattern-name.md](../patterns/pattern-name.md) - Pattern for [topic]

### Taskboard
- [05_Taskboard/beads.md#task-id](../../05_Taskboard/beads.md#task-id) - Task: [task description]

### HandsOn
- [06_HandsOn/project-name/](../../06_HandsOn/project-name/) - POC for [topic]
```

### For Research Docs

```markdown
## 🔗 Related Resources

### Learning
- [01_Learning/ProjectName/topic.md](../../01_Learning/ProjectName/topic.md) - Learning materials for [topic]

### Planning
- [04_Planning/ProjectName/enhancement-plan.md](../../04_Planning/ProjectName/enhancement-plan.md) - Plan based on this research

### Knowledge
- [03_Knowledge/patterns/pattern-name.md](../patterns/pattern-name.md) - Pattern validated by this research

### Taskboard
- [05_Taskboard/beads.md#task-id](../../05_Taskboard/beads.md#task-id) - Task: [task description]

### HandsOn
- [06_HandsOn/project-name/](../../06_HandsOn/project-name/) - POC validating research findings
```

### For Planning Docs

```markdown
## 🔗 Related Resources

### Learning
- [01_Learning/ProjectName/topic.md](../../01_Learning/ProjectName/topic.md) - Learning materials for [topic]

### Research
- [02_Research/ProjectName/analysis.md](../../02_Research/ProjectName/analysis.md) - Research informing this plan

### Knowledge
- [03_Knowledge/patterns/pattern-name.md](../patterns/pattern-name.md) - Patterns used in this plan

### Taskboard
- [05_Taskboard/beads.md#task-id](../../05_Taskboard/beads.md#task-id) - Tasks from this plan

### HandsOn
- [06_HandsOn/project-name/](../../06_HandsOn/project-name/) - Implementation of this plan
```

### For Knowledge Docs

```markdown
## 🔗 Related Resources

### Learning
- [01_Learning/ProjectName/topic.md](../../01_Learning/ProjectName/topic.md) - Learning materials for [topic]

### Research
- [02_Research/ProjectName/analysis.md](../../02_Research/ProjectName/analysis.md) - Research validating this knowledge

### Planning
- [04_Planning/ProjectName/enhancement-plan.md](../../04_Planning/ProjectName/enhancement-plan.md) - Plan using this knowledge

### Taskboard
- [05_Taskboard/beads.md#task-id](../../05_Taskboard/beads.md#task-id) - Task applying this knowledge

### HandsOn
- [06_HandsOn/project-name/](../../06_HandsOn/project-name/) - POC testing this knowledge
```

---

## 📝 Full Example: CLI Plugin Architecture

### 01_Learning/cli/plugin-architecture.md

```markdown
# CLI Plugin Architecture

## 🎯 Learning Objectives
- Understand plugin architecture patterns
- Learn dynamic loading mechanisms
- Study plugin lifecycle management

## 📚 Resources
- [Go Plugin Pattern](https://github.com/hashicorp/go-plugin)
- [Plugin Architecture Best Practices](https://medium.com/...)

## 🔗 Related Resources

### Research
- [02_Research/cli/plugin-analysis.md](../../02_Research/cli/plugin-analysis.md) - Analysis of plugin architectures

### Planning
- [04_Planning/cli/plugin-enhancement.md](../../04_Planning/cli/plugin-enhancement.md) - Plan to enhance plugin system

### Knowledge
- [03_Knowledge/cli/core/ARCHITECTURE.md](../cli/core/ARCHITECTURE.md) - Current CLI architecture

### Taskboard
- [05_Taskboard/beads.md#cli-plugin](../../05_Taskboard/beads.md#cli-plugin) - Task: Implement plugin system

### HandsOn
- [06_HandsOn/cli-plugin-poc/](../../06_HandsOn/cli-plugin-poc/) - POC for plugin system
```

### 02_Research/cli/plugin-analysis.md

```markdown
# CLI Plugin Architecture Analysis

## 🎯 Research Objective
Compare plugin architectures for CLI applications

## 🔬 Findings
- HashiCorp go-plugin: Best for Go CLI
- Python entry points: Best for Python CLI
- Node.js plugins: Best for Node CLI

## 💡 Recommendation
Use HashiCorp go-plugin for Ti CLI

## 🔗 Related Resources

### Learning
- [01_Learning/cli/plugin-architecture.md](../../01_Learning/cli/plugin-architecture.md) - Learning materials

### Planning
- [04_Planning/cli/plugin-enhancement.md](../../04_Planning/cli/plugin-enhancement.md) - Plan based on this research

### Knowledge
- [03_Knowledge/cli/core/ARCHITECTURE.md](../cli/core/ARCHITECTURE.md) - Current architecture

### Taskboard
- [05_Taskboard/beads.md#cli-plugin](../../05_Taskboard/beads.md#cli-plugin) - Task: Implement plugin system

### HandsOn
- [06_HandsOn/cli-plugin-poc/](../../06_HandsOn/cli-plugin-poc/) - POC validating research
```

### 04_Planning/cli/plugin-enhancement.md

```markdown
# CLI Plugin Enhancement Plan

## 🎯 Objectives
Implement plugin system for Ti CLI

## 📋 Requirements
- FR1: Support dynamic plugin loading
- FR2: Plugin lifecycle management
- FR3: Plugin discovery mechanism

## 🔗 Related Resources

### Learning
- [01_Learning/cli/plugin-architecture.md](../../01_Learning/cli/plugin-architecture.md) - Learning materials

### Research
- [02_Research/cli/plugin-analysis.md](../../02_Research/cli/plugin-analysis.md) - Research informing this plan

### Knowledge
- [03_Knowledge/cli/core/ARCHITECTURE.md](../cli/core/ARCHITECTURE.md) - Current architecture

### Taskboard
- [05_Taskboard/beads.md#cli-plugin](../../05_Taskboard/beads.md#cli-plugin) - Tasks from this plan

### HandsOn
- [06_HandsOn/cli-plugin-poc/](../../06_HandsOn/cli-plugin-poc/) - Implementation of this plan
```

---

## 🚀 Best Practices

1. **Use Relative Paths**
   - Always use relative paths (`../../02_Research/...`)
   - Avoid absolute paths (breaks when repo moves)

2. **Link Both Ways**
   - If Doc A links to Doc B, Doc B should link back to Doc A
   - Makes navigation easier

3. **Update Regularly**
   - Update cross-references when docs are moved/renamed
   - Remove broken links

4. **Be Specific**
   - Use descriptive link text
   - Include context in link description

5. **Check Links**
   - Verify links work before committing
   - Use markdown linter to check for broken links

---

## 📊 Connection Rate Tracking

**Goal**: >70% of docs have cross-references

**How to Track**:
```bash
# Count docs with cross-references
cd 03_Knowledge
grep -r "## 🔗 Related Resources" --include="*.md" | wc -l

# Count total docs
find . -name "*.md" | wc -l

# Calculate connection rate
# Connection Rate = (Docs with cross-references / Total docs) * 100
```

---

*Last Updated: 2026-05-06*
