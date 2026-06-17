# Auto-Skill Invocation System

> **Category**: Quality Rules  
> **Priority**: P1 (HIGH)  
> **Last Updated**: 2026-06-05  

---

## BẮT BUỘC Auto-Skill Invocation

**Mục đích**: Tự động load skills phù hợp dựa trên task type và language để đảm bảo quality

---

## Skill Invocation Logic

### Step 1: Classify Task

```python
def classify_task(task_description):
    """
    Classify task based on keywords in description
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

### Step 2: Detect Language

```python
def detect_language(project_path):
    """
    Detect primary language based on project files
    """
    # Check for Go
    if glob.glob('**/*.go', root_dir=project_path, recursive=True):
        return 'go'
    
    # Check for Python
    if glob.glob('**/*.py', root_dir=project_path, recursive=True):
        return 'python'
    
    # Check for Rust
    if glob.glob('**/*.rs', root_dir=project_path, recursive=True):
        return 'rust'
    
    # Check for Kotlin
    if glob.glob('**/*.kt', root_dir=project_path, recursive=True):
        return 'kotlin'
    
    # Check for TypeScript/JavaScript
    if glob.glob('**/*.ts', root_dir=project_path, recursive=True) or glob.glob('**/*.js', root_dir=project_path, recursive=True):
        return 'typescript'
    
    # Check for Java
    if glob.glob('**/*.java', root_dir=project_path, recursive=True):
        return 'java'
    
    # Default
    return 'unknown'
```

### Step 3: Select Workflow

```python
def select_workflow(task_type):
    """
    Select workflow based on task type
    """
    workflow_map = {
        'security': 'reflective-loop-beads.yaml',
        'deployment': 'reflective-loop-beads.yaml',
        'bug-fix': 'reflective-loop-beads.yaml',
        'feature': 'reflective-loop-beads.yaml',
        'performance': 'reflective-loop-beads.yaml',
        'testing': 'ti-cli-development.yaml',
        'refactor': 'reflective-loop.yaml',
        'research': 'qa-research-protocol.yaml',
        'documentation': 'ti-cli-development.yaml',
        'analysis': 'ti-cli-development.yaml',
        'standard': 'ti-cli-development.yaml'
    }
    
    return workflow_map.get(task_type, 'ti-cli-development.yaml')
```

### Step 4: Load Skills

```python
def load_skills(task_type, language):
    """
    Load skills based on task type and language
    """
    # Task-specific skills (mandatory)
    mandatory_skills = get_mandatory_skills(task_type)
    
    # Language-specific skills
    language_skills = get_language_skills(language)
    
    # Optional skills
    optional_skills = get_optional_skills(task_type, language)
    
    return {
        'mandatory': mandatory_skills,
        'language': language_skills,
        'optional': optional_skills
    }
```

---

## Skill Mapping Tables

### Task Type → Mandatory Skills

```yaml
task_mandatory_skills:
  security:
    - security-review
    - security-scan
    
  deployment:
    - deployment-patterns
    - safety-guard
    
  bug-fix:
    - tdd-workflow
    
  feature:
    - tdd-workflow
    - backend-patterns
    
  performance:
    - benchmark
    
  testing:
    - tdd-workflow
    - e2e-testing
    
  refactor:
    - coding-standards
    
  research:
    - deep-research
    
  documentation:
    - article-writing
    
  analysis:
    - repo-analysis
    - repo-scan
    
  standard: []
```

### Language → Language-Specific Skills

```yaml
language_skills:
  go:
    testing: golang-testing
    patterns: golang-patterns
    tdd: golang-testing
    
  python:
    testing: python-testing
    patterns: python-patterns
    tdd: python-tdd
    security: python-security
    
  rust:
    testing: rust-testing
    patterns: rust-patterns
    tdd: rust-testing
    
  kotlin:
    testing: kotlin-testing
    patterns: kotlin-patterns
    tdd: kotlin-testing
    
  typescript:
    testing: e2e-testing
    patterns: frontend-patterns
    tdd: tdd-workflow
    
  java:
    testing: springboot-tdd
    patterns: springboot-patterns
    security: springboot-security
```

### Task Type + Language → Optional Skills

```yaml
optional_skills:
  security:
    go: [backend-patterns]
    python: [python-patterns]
    typescript: [frontend-patterns]
    
  deployment:
    go: [monitoring]
    python: [monitoring]
    
  bug-fix:
    go: [golang-testing, debugging-patterns]
    python: [python-testing, debugging-patterns]
    
  feature:
    go: [api-design, database-migrations]
    python: [api-design, database-migrations]
    typescript: [frontend-patterns, api-design]
    
  performance:
    go: [systematic-optimization]
    python: [systematic-optimization]
    typescript: [systematic-optimization]
    
  testing:
    go: [testing-patterns]
    python: [testing-patterns]
    typescript: [testing-patterns]
    
  refactor:
    go: [strategic-compact]
    python: [strategic-compact]
    
  research:
    any: [exa-search]
    
  documentation:
    any: [markdown-patterns]
    
  analysis:
    any: [component-tree]
```

---

## Skill Invocation Flow

### Flowchart

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Receive Task Description                                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Classify Task Type (auto-classification)                 │
│    - Extract keywords                                        │
│    - Match task type                                         │
│    - Set priority                                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Detect Language (file extension analysis)                │
│    - Scan project files                                      │
│    - Detect primary language                                 │
│    - Set language context                                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Select Workflow (task type → workflow mapping)            │
│    - Load workflow file                                      │
│    - Verify workflow exists                                  │
│    - Set workflow context                                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Load Mandatory Skills (task type → mandatory skills)       │
│    - Load skill files                                        │
│    - Verify skills exist                                     │
│    - Set mandatory skill context                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. Load Language-Specific Skills (language → skills)         │
│    - Load language skill files                               │
│    - Verify skills exist                                     │
│    - Set language skill context                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 7. Load Optional Skills (task type + language → optional)     │
│    - Load optional skill files                               │
│    - Verify skills exist                                     │
│    - Set optional skill context                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 8. Invoke Skills                                             │
│    - Invoke mandatory skills (blocking)                      │
│    - Invoke language skills (blocking)                       │
│    - Invoke optional skills (non-blocking)                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ 9. Log Skill Invocation                                       │
│    - Log task type                                           │
│    - Log language                                            │
│    - Log workflow                                             │
│    - Log skills invoked                                      │
│    - Log invocation status                                   │
└─────────────────────────────────────────────────────────────┘
```

---

## Skill Invocation Example

### Example 1: Bug Fix in Go

```markdown
## Task: Fix memory leak in router service

### Step 1: Classification
Task description: "Fix memory leak in router service"
Keywords: fix, leak
Classification: bug-fix

### Step 2: Language Detection
Project files: apps/router/**/*.go
Language: go

### Step 3: Workflow Selection
Workflow: reflective-loop-beads.yaml

### Step 4: Load Skills
Mandatory skills:
  - tdd-workflow

Language skills:
  - golang-testing
  - golang-patterns

Optional skills:
  - debugging-patterns
  - strategic-compact

### Step 5: Invoke Skills
skill invoke tdd-workflow
skill invoke golang-testing
skill invoke golang-patterns
skill invoke debugging-patterns
skill invoke strategic-compact

### Step 6: Log Invocation
bd log --task="Skill invocation - bug-fix, go, reflective-loop-beads, tdd-workflow, golang-testing, golang-patterns, debugging-patterns, strategic-compact" \
       --agent=devin \
       --status=complete \
       --type=skill-invocation \
       --domain=general
```

### Example 2: Feature in TypeScript

```markdown
## Task: Implement user authentication

### Step 1: Classification
Task description: "Implement user authentication"
Keywords: implement, authentication
Classification: feature

### Step 2: Language Detection
Project files: apps/web/**/*.ts
Language: typescript

### Step 3: Workflow Selection
Workflow: reflective-loop-beads.yaml

### Step 4: Load Skills
Mandatory skills:
  - tdd-workflow
  - backend-patterns

Language skills:
  - e2e-testing
  - frontend-patterns

Optional skills:
  - api-design
  - database-migrations
  - security-review

### Step 5: Invoke Skills
skill invoke tdd-workflow
skill invoke backend-patterns
skill invoke e2e-testing
skill invoke frontend-patterns
skill invoke api-design
skill invoke database-migrations
skill invoke security-review

### Step 6: Log Invocation
bd log --task="Skill invocation - feature, typescript, reflective-loop-beads, tdd-workflow, backend-patterns, e2e-testing, frontend-patterns, api-design, database-migrations, security-review" \
       --agent=devin \
       --status=complete \
       --type=skill-invocation \
       --domain=general
```

### Example 3: Security Audit

```markdown
## Task: Security audit of router service

### Step 1: Classification
Task description: "Security audit of router service"
Keywords: security, audit
Classification: security

### Step 2: Language Detection
Project files: apps/router/**/*.go
Language: go

### Step 3: Workflow Selection
Workflow: reflective-loop-beads.yaml

### Step 4: Load Skills
Mandatory skills:
  - security-review
  - security-scan

Language skills:
  - golang-patterns

Optional skills:
  - backend-patterns

### Step 5: Invoke Skills
skill invoke security-review
skill invoke security-scan
skill invoke golang-patterns
skill invoke backend-patterns

### Step 6: Log Invocation
bd log --task="Skill invocation - security, go, reflective-loop-beads, security-review, security-scan, golang-patterns, backend-patterns" \
       --agent=devin \
       --status=complete \
       --type=skill-invocation \
       --domain=general
```

---

## Enforcement

### BẮT BUỘC

- **BẮT BUỘC** classify task trước khi load skills
- **BẮT BUỘC** load mandatory skills (blocking)
- **BẮT BUỘC** load language-specific skills (blocking)
- **BẮT BUỘC** log skill invocation với BD tool

### Violation Detection

- Skip task classification → Log violation
- Skip mandatory skills → Log violation
- Skip language skills → Log violation
- Skip skill invocation logging → Log violation

### Violation Logging

```bash
bd log --task="Skill invocation violation: {violation_type} - {details}" \
       --agent=devin \
       --status=failed \
       --type=violation \
       --domain=general
```

### Escalation

- Repeated violations (≥3 times) → Escalate to human review
- Critical violations (skip mandatory skills for security/deployment) → Immediate escalation

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
