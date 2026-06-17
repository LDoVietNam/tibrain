# Task Classification System

> **Category**: Quality Rules  
> **Priority**: P0 (BẮT BUỘC)  
> **Last Updated**: 2026-06-05  

---

## BẮT BUỘC Task Classification

**Mục đích**: Devin PHẢI classify task trước khi bắt đầu để chọn workflow và skills phù hợp

---

## Task Types

| Type | Description | Workflow | Auto-Skills | Priority |
|------|-------------|----------|-------------|----------|
| **Bug Fix** | Fix reported bug | reflective-loop-beads | tdd-workflow, testing | High |
| **Feature** | Implement new feature | reflective-loop-beads | tdd-workflow, backend-patterns | High |
| **Refactor** | Improve existing code | reflective-loop | coding-standards | Medium |
| **Documentation** | Write/update docs | ti-cli-development | article-writing | Low |
| **Performance** | Optimize performance | reflective-loop-beads | benchmark, systematic-optimization | High |
| **Security** | Security audit/fix | reflective-loop-beads | security-review | Critical |
| **Testing** | Write tests | ti-cli-development | tdd-workflow, e2e-testing | Medium |
| **Deployment** | Deploy to production | reflective-loop-beads | deployment-patterns, safety-guard | Critical |
| **Research** | Research problem | qa-research-protocol | deep-research | Medium |
| **Analysis** | Analyze codebase | ti-cli-development | repo-analysis, repo-scan | Low |

---

## Classification Criteria

### Bug Fix
**Keywords:** fix, bug, error, issue, broken, crash, fail, exception
**Output:** Code change only
**Example:** "Fix memory leak in router service"

### Feature
**Keywords:** add, implement, create, new, build, develop
**Output:** New functionality
**Example:** "Implement user authentication"

### Refactor
**Keywords:** refactor, improve, optimize, clean, restructure, reorganize
**Output:** Better code structure
**Example:** "Refactor database layer to use repository pattern"

### Documentation
**Keywords:** doc, readme, guide, tutorial, explain, document
**Output:** Documentation files
**Example:** "Write API documentation"

### Performance
**Keywords:** slow, performance, optimize, speed, latency, throughput
**Output:** Faster code
**Example:** "Optimize database query performance"

### Security
**Keywords:** security, auth, vulnerability, exploit, secure, protect
**Output:** Security fixes
**Example:** "Fix SQL injection vulnerability"

### Testing
**Keywords:** test, spec, coverage, e2e, unit test, integration test
**Output:** Test files
**Example:** "Write unit tests for user service"

### Deployment
**Keywords:** deploy, release, production, staging, deploy
**Output:** Deployment config
**Example:** "Deploy router service to production"

### Research
**Keywords:** research, investigate, explore, find, analyze
**Output:** Research report
**Example:** "Research best practices for Go MCP servers"

### Analysis
**Keywords:** analyze, review, audit, assess, evaluate
**Output:** Analysis report
**Example:** "Analyze codebase architecture"

---

## Auto-Classification Logic

```python
def classify_task(task_description):
    """
    Auto-classify task based on keywords in description
    """
    keywords = task_description.lower()
    
    # Priority order (critical first)
    if any(k in keywords for k in ['security', 'auth', 'vulnerability', 'exploit']):
        return 'security'
    elif any(k in keywords for k in ['deploy', 'release', 'production', 'staging']):
        return 'deployment'
    elif any(k in keywords for k in ['fix', 'bug', 'error', 'issue', 'broken', 'crash', 'fail', 'exception']):
        return 'bug-fix'
    elif any(k in keywords for k in ['add', 'implement', 'create', 'new', 'build', 'develop']):
        return 'feature'
    elif any(k in keywords for k in ['refactor', 'improve', 'optimize', 'clean', 'restructure']):
        return 'refactor'
    elif any(k in keywords for k in ['doc', 'readme', 'guide', 'tutorial', 'explain', 'document']):
        return 'documentation'
    elif any(k in keywords for k in ['slow', 'performance', 'speed', 'latency', 'throughput']):
        return 'performance'
    elif any(k in keywords for k in ['test', 'spec', 'coverage', 'e2e', 'unit test', 'integration test']):
        return 'testing'
    elif any(k in keywords for k in ['research', 'investigate', 'explore', 'find', 'analyze', 'review', 'audit']):
        return 'research'
    
    # Default
    return 'standard'
```

---

## Workflow Mapping

```yaml
task_classification:
  security:
    workflow: reflective-loop-beads.yaml
    priority: critical
    auto_skills:
      mandatory:
        - security-review
        - security-scan
      optional:
        - backend-patterns
        
  deployment:
    workflow: reflective-loop-beads.yaml
    priority: critical
    auto_skills:
      mandatory:
        - deployment-patterns
        - safety-guard
      optional:
        - monitoring
        
  bug-fix:
    workflow: reflective-loop-beads.yaml
    priority: high
    auto_skills:
      mandatory:
        - tdd-workflow
      optional:
        - python-testing
        - debugging-patterns
        - error-handling
        
  feature:
    workflow: reflective-loop-beads.yaml
    priority: high
    auto_skills:
      mandatory:
        - tdd-workflow
        - backend-patterns
      optional:
        - api-design
        - database-migrations
        - frontend-patterns
        
  performance:
    workflow: reflective-loop-beads.yaml
    priority: high
    auto_skills:
      mandatory:
        - benchmark
      optional:
        - systematic-optimization
        - strategic-compact
        
  testing:
    workflow: ti-cli-development.yaml
    priority: medium
    auto_skills:
      mandatory:
        - tdd-workflow
        - e2e-testing
      optional:
        - testing-patterns
        
  refactor:
    workflow: reflective-loop.yaml
    priority: medium
    auto_skills:
      mandatory:
        - coding-standards
      optional:
        - strategic-compact
        
  research:
    workflow: qa-research-protocol.yaml
    priority: medium
    auto_skills:
      mandatory:
        - deep-research
      optional:
        - exa-search
        
  documentation:
    workflow: ti-cli-development.yaml
    priority: low
    auto_skills:
      mandatory:
        - article-writing
      optional:
        - markdown-patterns
        
  analysis:
    workflow: ti-cli-development.yaml
    priority: low
    auto_skills:
      mandatory:
        - repo-analysis
        - repo-scan
      optional:
        - component-tree
        
  standard:
    workflow: ti-cli-development.yaml
    priority: medium
    auto_skills:
      mandatory: []
      optional:
        - coding-standards
```

---

## Language-Specific Skill Mapping

```yaml
language_skills:
  python:
    testing: python-testing
    patterns: python-patterns
    tdd: python-tdd
    security: python-security
    
  go:
    testing: go-testing
    patterns: golang-patterns
    tdd: golang-testing
    security: golang-security
    
  rust:
    testing: rust-testing
    patterns: rust-patterns
    tdd: rust-testing
    security: rust-patterns
    
  kotlin:
    testing: kotlin-testing
    patterns: kotlin-patterns
    tdd: kotlin-testing
    security: kotlin-patterns
    
  typescript:
    testing: e2e-testing
    patterns: frontend-patterns
    tdd: tdd-workflow
    security: frontend-patterns
    
  java:
    testing: springboot-tdd
    patterns: springboot-patterns
    security: springboot-security
```

---

## Enforcement

### BẮT BUỘC

1. **Classify task** trước khi bắt đầu
2. **Select workflow** theo classification
3. **Load auto-skills** theo task type + language
4. **Log classification** với BD tool

### Violation Detection

- Skip classification → Log violation
- Wrong workflow selection → Log violation
- Skip mandatory skills → Log violation

### Violation Logging

```bash
bd log --task="Classification violation: {violation_type} - {details}" \
       --agent=devin \
       --status=failed \
       --type=violation \
       --domain=general
```

### Escalation

- Repeated violations (≥3 times) → Escalate to human review
- Critical violations (skip security/deployment classification) → Immediate escalation

---

## Usage Example

```markdown
## Task: Fix memory leak in router service

### Step 1: Classification
Task description: "Fix memory leak in router service"
Keywords: fix, leak
Classification: bug-fix

### Step 2: Workflow Selection
Workflow: reflective-loop-beads.yaml
Priority: high

### Step 3: Auto-Skill Invocation
Language: Go
Auto-skills:
  - tdd-workflow (mandatory)
  - go-testing (language-specific)
  - debugging-patterns (optional)

### Step 4: Execute
Follow reflective-loop-beads.yaml workflow
Apply mandatory rules from content/rules/core/
Log decisions with beads

### Step 5: Completion
Post-flight checklist
Quality metrics logging
```

---

## Reference

- **Quality Principles**: content/rules/quality/mandatory-quality-principles.md
- **Quality Gates**: content/rules/quality/mandatory-quality-gates.md
- **Workflows**: content/workflows/INDEX.md
- **Skills**: content/skills/
- **Beads Protocol**: Z:\Ti\taskboard\beads.md
