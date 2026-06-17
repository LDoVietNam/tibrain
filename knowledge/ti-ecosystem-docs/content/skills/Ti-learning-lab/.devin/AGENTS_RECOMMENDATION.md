# aitmpl.com Agents - Recommendations cho Ti CLI

> **Source**: https://www.aitmpl.com/agents (claude-code-templates)
> **Total**: 400+ Agents, 225+ Commands, 65+ MCPs, 60+ Settings, 45+ Hooks, 2700+ Skills

---

## 📊 OVERVIEW

aitmpl.com cung cấp **3500+ components** sẵn sàng dùng cho Claude Code:

| Type | Count | Use For |
|------|-------|---------|
| 🤖 Agents | 400+ | Specialized AI specialists |
| ⚡ Commands | 225+ | Slash commands |
| 🔌 MCPs | 65+ | External integrations |
| ⚙️ Settings | 60+ | Configurations |
| 🪝 Hooks | 45+ | Automation triggers |
| 🎨 Skills | 2700+ | Reusable capabilities |

---

## 🎯 STRATEGY: 3-TIER USAGE

### Tier 1: BUILD Ti CLI (Immediate use) 🛠️
**Use these agents NOW to build Ti CLI architecture**

### Tier 2: INTEGRATE into Ti CLI 🔌
**Bake these into Ti's plugin system**

### Tier 3: EXPOSE to End Users 📦
**Ti CLI act as registry/router for these**

---

## 🛠️ TIER 1: BUILD Ti CLI (Use Immediately)

### A. Architecture & Planning

**1. `code-architect`** ⭐⭐⭐⭐⭐
- **Purpose**: System architecture design
- **Use for**: Design Ti Hybrid Architecture (3 layers)
- **Why important**: Help validate design, suggest patterns
- **Install**: `npx claude-code-templates@latest --agent development-team/code-architect`

**2. `backend-architect`** ⭐⭐⭐⭐⭐
- **Purpose**: Backend system design
- **Use for**: Design Ti Router brain endpoints
- **Why important**: API design, data flow, scaling
- **Install**: `npx claude-code-templates@latest --agent development-team/backend-architect`

**3. `task-decomposition-expert`** ⭐⭐⭐⭐⭐
- **Purpose**: Split complex tasks into subtasks
- **Use for**: Break down Ti Hybrid Design implementation
- **Why important**: Manage 3 parallel tracks (A/B/C)
- **Install**: `npx claude-code-templates@latest --agent ai-specialists/task-decomposition-expert`

### B. Go Development (Ti CLI is Go)

**4. `backend-developer`** ⭐⭐⭐⭐⭐
- **Purpose**: API and server-side development
- **Use for**: Implement Ti Router brain endpoints
- **Why important**: Go expertise, API patterns
- **Install**: `npx claude-code-templates@latest --agent development-team/backend-developer`

**5. `code-explorer`** ⭐⭐⭐⭐
- **Purpose**: Codebase exploration
- **Use for**: Understand existing Ti CLI 60+ packages
- **Why important**: Quickly map dependencies
- **Install**: `npx claude-code-templates@latest --agent development-team/code-explorer`

### C. AI/LLM Specifics

**6. `prompt-engineer`** ⭐⭐⭐⭐⭐ (CRITICAL!)
- **Purpose**: Prompt design and optimization
- **Use for**: Design **custom prompt injection** logic
- **Why important**: Core của Ti CLI Wrapper - injection quality = output quality
- **Install**: `npx claude-code-templates@latest --agent ai-specialists/prompt-engineer`

**7. `llm-architect`** ⭐⭐⭐⭐
- **Purpose**: LLM system architecture
- **Use for**: Optimize Ti Brain context window management
- **Why important**: 4-layer MemPalace optimization
- **Install**: `npx claude-code-templates@latest --agent ai-specialists/llm-architect`

**8. `model-evaluator`** ⭐⭐⭐⭐
- **Purpose**: Evaluate LLM outputs
- **Use for**: Test Ti Brain quality across agents
- **Why important**: Verify cross-agent learning works
- **Install**: `npx claude-code-templates@latest --agent ai-specialists/model-evaluator`

### D. Security & DevOps

**9. `security-auditor`** ⭐⭐⭐⭐
- **Purpose**: Security review
- **Use for**: Audit Ti Router auth, secrets management
- **Why important**: Z:\00_SECRET\ integration must be secure
- **Install**: `npx claude-code-templates@latest --agent security/security-auditor`

**10. `devops-engineer`** ⭐⭐⭐⭐
- **Purpose**: Infrastructure and deployment
- **Use for**: Deploy Ti Router as service, manage MCP server
- **Why important**: Production deployment
- **Install**: `npx claude-code-templates@latest --agent development-team/devops-engineer`

---

## 🔌 TIER 2: INTEGRATE INTO Ti's Plugin System

These should become **native Ti plugins** (in `internal/plugins/` or `cmd/agents/`):

### A. Code Quality Agents

**11. `code-reviewer`** - Review code changes (potential Hook)
**12. `refactoring-specialist`** - Code refactoring suggestions
**13. `performance-optimizer`** - Performance improvements

### B. Testing Agents

**14. `test-engineer`** - Generate unit tests
**15. `qa-engineer`** - Quality assurance
**16. `e2e-tester`** - End-to-end testing

### C. Documentation Agents

**17. `technical-writer`** - Generate documentation
**18. `api-documentation`** - API docs from code
**19. `readme-generator`** - README files

### D. Database Agents (for Ti Brain SQLite)

**20. `database/sqlite-expert`** (if exists)
**21. `database-migration-specialist`** - Schema migrations

### Integration Pattern:

```go
// Z:\Ti\CLI\internal\agents\external\
package external

type AitmplAgent struct {
    Name     string
    Category string
    FilePath string  // .claude/agents/{name}.md
    Tools    []string
    Model    string  // sonnet/opus/haiku
}

// Ti auto-loads aitmpl agents on startup
func LoadAitmplAgents(claudeDir string) ([]*AitmplAgent, error) {
    // Scan .claude/agents/ directory
    // Parse markdown frontmatter
    // Register as Ti plugins
}

// Ti exposes agents through brain server
// /v1/brain/agents - list all available agents
// /v1/brain/agents/{name} - get agent details
// /v1/brain/agents/{name}/invoke - run agent through Ti
```

---

## 📦 TIER 3: EXPOSE TO END USERS

Ti CLI becomes **Universal Agent Registry**:

```bash
# List all available agents
ti agent list

# Run specific agent
ti agent run frontend-developer "build login page"

# Install new agent
ti agent install api-security-audit

# Search agents
ti agent search "react"

# Update all agents
ti agent update --all
```

### Categories to Expose (User-Facing):

| Category | Examples | Use Case |
|----------|----------|----------|
| **Development** | frontend, backend, fullstack, mobile | Day-to-day coding |
| **Database** | postgres, mysql, mongodb, redis | DB tasks |
| **Security** | api-audit, pen-tester, compliance | Security reviews |
| **DevOps** | k8s, docker, terraform, aws | Deployment |
| **AI/ML** | ml-engineer, data-scientist, nlp | Data tasks |
| **Documentation** | technical-writer, api-docs | Doc generation |
| **Testing** | unit-test, e2e, performance | QA |

---

## 🎯 RECOMMENDED INSTALL ORDER

### Week 1: Foundation (Tier 1 - Build Ti)

```bash
# Architecture & Planning (3 agents)
npx claude-code-templates@latest \
  --agent development-team/code-architect \
  --agent development-team/backend-architect \
  --agent ai-specialists/task-decomposition-expert

# Go Development (2 agents)
npx claude-code-templates@latest \
  --agent development-team/backend-developer \
  --agent development-team/code-explorer

# Prompt & LLM (3 agents - CRITICAL)
npx claude-code-templates@latest \
  --agent ai-specialists/prompt-engineer \
  --agent ai-specialists/llm-architect \
  --agent ai-specialists/model-evaluator
```

### Week 2: Quality (Tier 1 - Verify)

```bash
# Security & DevOps
npx claude-code-templates@latest \
  --agent security/security-auditor \
  --agent development-team/devops-engineer
```

### Week 3+: Integration (Tier 2 - Bake into Ti)

```bash
# Add to Ti plugins
# (after Ti CLI Wrapper Phase 3 done)
```

---

## 🚀 SPECIFIC RECOMMENDATIONS PER PHASE

### Phase 1: AGENTS.md Convergence (1 day)
**Use**: `prompt-engineer` to design AGENTS.md content
- Optimize for LLM consumption
- Multi-level scoping (project/user/system)
- Cross-tool compatibility

### Phase 2: Ti Router Brain Endpoints (3 days)
**Use**: 
- `backend-architect` - design endpoint API
- `backend-developer` - implement Go handlers
- `security-auditor` - review auth

### Phase 3: Ti CLI Wrapper (5-7 days)
**Use**:
- `prompt-engineer` ⭐ - design custom prompt injection
- `llm-architect` - optimize context window
- `task-decomposition-expert` - split wrapper logic

### Phase 4: Cross-Agent Brain Learning
**Use**:
- `model-evaluator` - test learning effectiveness
- `llm-architect` - tune RL parameters
- `prompt-engineer` - improve injection over time

---

## 💡 SPECIAL: COMPLEMENTARY COMPONENTS

### Hooks to Install (45+ available)

**For Ti CLI development**:

```bash
# Pre-commit validation
npx claude-code-templates@latest --hook git/pre-commit-validation

# Post-completion notifications
npx claude-code-templates@latest --hook notifications/post-completion

# Error handling
npx claude-code-templates@latest --hook errors/auto-recovery
```

### Commands to Install (225+ available)

**For Ti CLI testing**:

```bash
# Generate tests
npx claude-code-templates@latest --command testing/generate-tests

# Optimize performance
npx claude-code-templates@latest --command performance/optimize-bundle

# Security check
npx claude-code-templates@latest --command security/check-security

# Setup CI/CD
npx claude-code-templates@latest --command devops/setup-ci
```

### MCPs to Install (65+ available)

**For Ti CLI infrastructure**:

```bash
# GitHub integration (for Ti repos)
npx claude-code-templates@latest --mcp development/github-integration

# PostgreSQL (if Ti uses Postgres)
npx claude-code-templates@latest --mcp database/postgresql-integration

# Docker (Ti Router deployment)
npx claude-code-templates@latest --mcp devops/docker-integration
```

---

## ⚠️ POTENTIAL CONFLICTS WITH Ti

### Potential Issue 1: Agent Storage Location
- **aitmpl agents**: `.claude/agents/`
- **Ti agents**: `Z:\Ti\CLI\internal\agents\`
- **Solution**: Ti reads `.claude/agents/` and registers them

### Potential Issue 2: Tool Permissions
- aitmpl agents define their own tool permissions
- Ti has its own permission system (`internal/permission/`)
- **Solution**: Ti respects agent's tool list, intersect with Ti's permissions

### Potential Issue 3: Model Selection
- aitmpl agents specify `sonnet/opus/haiku`
- Ti Router has its own model routing
- **Solution**: Ti maps `sonnet→best-router-model`, etc.

### Potential Issue 4: Memory Conflict
- aitmpl agents query `context-manager` for project context
- Ti has `MemPalace` 4-layer memory
- **Solution**: Ti implements `context-manager` interface, return MemPalace data

---

## 🔄 COMPATIBILITY MATRIX

| aitmpl Component | Ti CLI Status | Integration |
|------------------|---------------|-------------|
| Agents (.claude/agents/) | Compatible | Ti reads + enhances |
| Commands (slash commands) | Need wrapper | Ti CLI provides via `ti agent` |
| MCPs | Compatible | Ti Router exposes MCP |
| Hooks | Need integration | Ti's `internal/events` |
| Settings | Compatible | Ti uses `internal/config` |
| Skills (markdown) | Compatible | Same as Ti's `internal/learn` |

---

## 🎯 SUMMARY RECOMMENDATIONS

### MUST INSTALL (Critical for Ti Build) - 8 agents:

1. ⭐⭐⭐⭐⭐ `prompt-engineer` - Custom prompt injection (CRITICAL)
2. ⭐⭐⭐⭐⭐ `code-architect` - Architecture design
3. ⭐⭐⭐⭐⭐ `backend-architect` - API design
4. ⭐⭐⭐⭐⭐ `backend-developer` - Go implementation
5. ⭐⭐⭐⭐⭐ `task-decomposition-expert` - Task breakdown
6. ⭐⭐⭐⭐ `llm-architect` - LLM optimization
7. ⭐⭐⭐⭐ `code-explorer` - Codebase navigation
8. ⭐⭐⭐⭐ `security-auditor` - Security review

### SHOULD INSTALL (Quality & Productivity) - 6 agents:

9. `model-evaluator` - Output quality
10. `devops-engineer` - Deployment
11. `code-reviewer` - Code quality
12. `test-engineer` - Testing
13. `technical-writer` - Documentation
14. `refactoring-specialist` - Code improvements

### COULD INSTALL (Future, when needed):

- 100+ specialized agents based on task
- Database experts (when scaling)
- Frontend agents (if Ti gets web UI)
- Mobile agents (if Ti gets mobile app)

---

## 🔥 KEY INSIGHT

**Ti CLI's killer feature**: Become the **Universal Agent Hub**

```
User → Ti CLI → registers/loads ALL aitmpl agents
                ↓
          Ti Brain Server enhances each agent với:
          - Cross-agent memory
          - RL learning
          - Pattern recognition  
          - Custom prompt injection
                ↓
          Each agent becomes 10x smarter through Ti
```

This is **bigger than just wrapping .devin** - Ti becomes the **brain layer** for **400+ agents** ecosystem!

---

## 📋 NEXT STEPS

### Immediate (Today):

```bash
# Install Tier 1 critical agents
npx claude-code-templates@latest \
  --agent ai-specialists/prompt-engineer \
  --agent development-team/code-architect \
  --agent development-team/backend-architect \
  --agent development-team/backend-developer \
  --agent ai-specialists/task-decomposition-expert \
  --yes
```

### This Week:

1. Use `prompt-engineer` to design Ti Brain prompt injection format
2. Use `backend-architect` to design Ti Router brain endpoints
3. Use `task-decomposition-expert` to split implementation into subtasks
4. Use `code-architect` to validate Hybrid Design

### Next Week:

1. Implement Phase 1 (AGENTS.md migration)
2. Start Phase 2 (Brain endpoints)
3. Start Phase 3 (CLI Wrapper prototype)

---

**Last Updated**: 2026-04-28
**Source**: https://www.aitmpl.com/agents | https://github.com/davila7/claude-code-templates
