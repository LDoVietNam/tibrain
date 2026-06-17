# Best Source Integration vào Ti CLI

> **Source**: `Z:\Ti\best_source\`
> **Discovered**: 30 agents + 60 commands + 130+ skills
> **Quality**: Production-ready, detailed, structured

---

## 🎯 KEY DISCOVERY

User đã có **best_source** - bộ sưu tập đã được curated kỹ với chất lượng cao hơn aitmpl 400+ agents (vì đã filter):

### Numbers:

| Type | Count | Quality |
|------|-------|---------|
| **Agents** | 30 | ⭐⭐⭐⭐⭐ Detailed YAML + instructions |
| **Commands** | 60 | ⭐⭐⭐⭐⭐ Comprehensive workflows |
| **Skills** | 130+ | ⭐⭐⭐⭐⭐ Production patterns |

### Key Finding: PERFECT MATCH với Ti's Architecture

**Best source đã có những thứ Ti cần**:
- ✅ `architect` - Design Ti architecture
- ✅ `planner` - Planning (= Ti's planner)
- ✅ `chief-of-staff` - Orchestration
- ✅ `loop-operator` - Loop management
- ✅ `go-reviewer`, `go-build-resolver` - Go specifics (Ti là Go!)
- ✅ `harness-optimizer` - Testing
- ✅ `security-reviewer` - Security
- ✅ `tdd-guide` - TDD

**Commands match Ti's needs**:
- ✅ `model-route` → Ti Router functionality
- ✅ `orchestrate` → Ti Brain orchestration
- ✅ `multi-execute, multi-plan, multi-workflow` → Parallel agents
- ✅ `instinct-export/import/status` → Memory system
- ✅ `prompt-optimize` → Custom prompt injection
- ✅ `context-budget` → MemPalace optimization
- ✅ `learn, learn-eval, evolve` → RL learning
- ✅ `verification-loop` → Verify

**Skills match Ti's intelligence**:
- ✅ `agentic-engineering` - Agent engineering
- ✅ `agent-eval, agent-harness-construction` - Agent eval
- ✅ `autonomous-loops, continuous-agent-loop` - Continuous learning
- ✅ `continuous-learning, continuous-learning-v2` - Learning
- ✅ `compaction-gate, strategic-compact` - Context compression
- ✅ `context-budget, token-budget-advisor` - Token management
- ✅ `cost-aware-llm-pipeline` - Cost optimization
- ✅ `golang-patterns, golang-testing` - Go specifics
- ✅ `mcp-server-patterns` - MCP server!
- ✅ `prompt-optimizer` - Prompt optimization
- ✅ `safety-guard` - Safety
- ✅ `search-first, iterative-retrieval` - Search patterns
- ✅ `verification-loop` - Verification

---

## 🚀 NEW STRATEGY: Use Best Source

### ❌ OLD STRATEGY:
Install agents từ aitmpl.com (npx claude-code-templates...)

### ✅ NEW STRATEGY:
**Symlink/copy best_source vào Ti CLI architecture**:

```
Z:\Ti\best_source\           # Source of truth
    ├── agents/              # 30 production agents
    ├── commands/            # 60 commands
    └── skills/              # 130+ skills

Z:\Ti\CLI\internal\agents\external\
    ├── agents/ → symlink to best_source/agents/
    ├── commands/ → symlink to best_source/commands/
    └── skills/ → symlink to best_source/skills/

Each project:
    .claude/agents/           → use best_source agents
    .claude/commands/         → use best_source commands
    .claude/skills/           → use best_source skills
```

### Why this is better:

1. **Already curated** - User đã filter quality cao
2. **Go-focused** - Có `go-reviewer`, `go-build-resolver` (Ti là Go)
3. **Single source of truth** - Update 1 nơi, propagate everywhere
4. **No external dependencies** - Không cần npx, no internet
5. **Production-tested** - đã verify

---

## 📊 MAPPING: Best Source → Ti CLI Architecture

### Layer 1: Ti CLI Wrapper

**Use these agents để build Wrapper**:

| Best Source Agent | Use In | Purpose |
|-------------------|--------|---------|
| `architect` | Phase 1-2 | Design Ti Hybrid Architecture |
| `planner` | Phase 1-3 | Plan implementation phases |
| `chief-of-staff` | Phase 3 | Orchestrate Wrapper logic |
| `go-reviewer` | All phases | Review Go code |
| `go-build-resolver` | All phases | Fix Go build errors |
| `tdd-guide` | All phases | TDD workflow |

### Layer 2: Ti Router Brain Endpoints

**Use these agents để build endpoints**:

| Best Source Agent | Use In | Purpose |
|-------------------|--------|---------|
| `architect` | Phase 2 | Design API endpoints |
| `go-reviewer` | Phase 2 | Review Go handler code |
| `security-reviewer` | Phase 2 | Audit auth + secrets |
| `performance-optimizer` | Phase 2+ | Optimize endpoint performance |
| `database-reviewer` | Phase 2 | Review SQLite schema |

**Use these commands**:

| Command | Purpose |
|---------|---------|
| `/orchestrate feature` | Workflow: planner → tdd → reviewer → security |
| `/model-route` | Direct route logic (= Ti Router functionality) |
| `/go-build` | Build Go endpoints |
| `/go-review` | Review Go code |
| `/go-test` | Test endpoints |

### Layer 3: Ti Brain Core

**Use these skills để enhance Brain**:

| Skill | Integrate Into |
|-------|----------------|
| `continuous-learning-v2` | `internal/learn/` |
| `compaction-gate` | `internal/memory/` (compression) |
| `context-budget` | `internal/contextpack/` |
| `token-budget-advisor` | `internal/memory/` |
| `prompt-optimizer` | NEW: prompt injection logic |
| `agentic-engineering` | Overall architecture |
| `mcp-server-patterns` | `internal/mcp/` (Brain MCP server) |

---

## 🎯 RECOMMENDED WORKFLOW

### Step 1: Setup Symlinks (Today)

```powershell
# Create symlinks from project to best_source
cd Z:\Ti\Ti-learning-lab
mkdir -p .claude

# Symlink agents
mklink /J ".claude\agents" "Z:\Ti\best_source\agents"
mklink /J ".claude\commands" "Z:\Ti\best_source\commands"
mklink /J ".claude\skills" "Z:\Ti\best_source\skills"
```

### Step 2: Use Agents for Phase 1 (AGENTS.md Convergence)

**Workflow**:
```
1. /orchestrate refactor "Migrate .devin/ to AGENTS.md format"
   ↓
2. architect agent: Design AGENTS.md structure
   ↓
3. tdd-guide: Plan tests for migration
   ↓
4. code-reviewer: Review migration
```

### Step 3: Use Agents for Phase 2 (Brain Endpoints)

**Workflow**:
```
1. /orchestrate feature "Add Brain endpoints to Ti Router"
   ↓
2. planner: Create implementation plan
   - Use best_source/agents/planner.md format
   - Phase 1: API design
   - Phase 2: Go handlers
   - Phase 3: Tests
   ↓
3. architect: Design endpoint structure
   ↓
4. go-build-resolver: Implement handlers
   ↓
5. go-reviewer: Review Go code
   ↓
6. security-reviewer: Audit auth
   ↓
7. tdd-guide: Add tests
```

### Step 4: Use Agents for Phase 3 (CLI Wrapper)

**Workflow**:
```
1. /orchestrate feature "Build Ti CLI Wrapper với prompt injection"
   ↓
2. planner: Plan wrapper logic
   ↓
3. architect: Design wrapper architecture
   ↓
4. /prompt-optimize: Use prompt-optimizer skill
   ↓
5. go-build-resolver: Implement wrapper
   ↓
6. go-reviewer: Code review
```

---

## 🔌 INTEGRATION INTO Ti's PLUGIN SYSTEM

### Phase 4 Integration Plan:

**1. Create `internal/agents/external/` package**:

```go
// Z:\Ti\CLI\internal\agents\external\loader.go
package external

import (
    "os"
    "path/filepath"
    "gopkg.in/yaml.v2"
)

type ExternalAgent struct {
    Name        string   `yaml:"name"`
    Description string   `yaml:"description"`
    Tools       []string `yaml:"tools"`
    Model       string   `yaml:"model"`
    FilePath    string   // Source markdown file
    Body        string   // Markdown body (instructions)
}

func LoadFromBestSource() ([]*ExternalAgent, error) {
    sourceDir := "Z:\\Ti\\best_source\\agents"
    return loadAgentsFromDir(sourceDir)
}

func LoadFromClaudeAgents(projectDir string) ([]*ExternalAgent, error) {
    claudeDir := filepath.Join(projectDir, ".claude", "agents")
    return loadAgentsFromDir(claudeDir)
}
```

**2. Register agents qua Ti CLI plugin system**:

```go
// Z:\Ti\CLI\cmd\agents\register.go
package agents

func RegisterExternalAgents(registry *plugins.Registry) error {
    agents, err := external.LoadFromBestSource()
    if err != nil {
        return err
    }
    
    for _, agent := range agents {
        plugin := &AgentPlugin{
            Name:        agent.Name,
            Description: agent.Description,
            Tools:       agent.Tools,
            Model:       agent.Model,
            Instructions: agent.Body,
        }
        registry.Register(plugin)
    }
    
    return nil
}
```

**3. Expose qua Ti CLI commands**:

```bash
# List agents
ti agent list                            # All registered agents

# Run agent
ti agent run go-reviewer "review my Go code"

# Use orchestrate
ti orchestrate feature "Add new feature X"

# Use commands
ti command run go-review .
ti command run model-route "complex task"
```

**4. Ti Brain enhances each agent**:

```
User: ti agent run go-reviewer "review code"
   ↓
Ti CLI Wrapper:
  1. Load go-reviewer agent từ best_source
  2. Query Ti Brain: relevant patterns for Go review
  3. Inject context: "Recent Go bugs in this project: ..."
  4. Spawn Claude/Devin with enhanced agent prompt
  5. Capture output
  6. Brain learns: "Go reviewer applied to X, found Y patterns"
   ↓
Output: Enhanced review (better than running go-reviewer alone)
```

---

## 🎯 BUILD ROADMAP với Best Source

### Phase 1 (Week 1): Setup + AGENTS.md
- [x] Discover best_source
- [ ] Create symlinks `.claude/agents/`, `.claude/commands/`, `.claude/skills/`
- [ ] Use `architect` + `planner` agents để design AGENTS.md
- [ ] Migrate `.devin/` → `AGENTS.md`

### Phase 2 (Week 2): Brain Endpoints
- [ ] Use `/orchestrate feature` workflow
- [ ] `planner` agent: Create plan
- [ ] `architect` agent: Design endpoints
- [ ] `go-build-resolver`: Implement
- [ ] `go-reviewer` + `security-reviewer`: Verify

### Phase 3 (Week 3): CLI Wrapper
- [ ] Use `prompt-optimize` skill
- [ ] `architect` agent: Design wrapper
- [ ] `chief-of-staff` agent: Orchestration logic
- [ ] `loop-operator` agent: Loop management

### Phase 4 (Week 4+): Plugin Integration
- [ ] Create `internal/agents/external/` package
- [ ] Auto-load best_source agents
- [ ] Expose qua `ti agent` commands
- [ ] Brain enhances each agent

### Phase 5 (Week 5+): Cross-Agent Learning
- [ ] Use `continuous-learning-v2` skill
- [ ] Use `agent-eval` skill
- [ ] Implement RL across agents
- [ ] Pattern sharing

---

## 💡 ADDITIONAL INSIGHTS

### Skills là TREASURE TROVE

**Top 10 skills phải đọc kỹ**:

1. **`agentic-engineering`** - Engineering patterns cho agents
2. **`continuous-learning-v2`** - RL learning patterns
3. **`compaction-gate`** - Context compression
4. **`token-budget-advisor`** - Token management
5. **`prompt-optimizer`** - Prompt engineering
6. **`mcp-server-patterns`** - MCP server design
7. **`cost-aware-llm-pipeline`** - Cost optimization
8. **`golang-patterns`** - Go best practices
9. **`safety-guard`** - Safety patterns
10. **`enterprise-agent-ops`** - Production ops

### Commands ecosystem rất mạnh

**Workflow commands** đã có:
- `/orchestrate` - Multi-agent orchestration
- `/multi-plan` - Multi-step planning
- `/multi-execute` - Parallel execution
- `/multi-workflow` - Workflow management

**Memory commands**:
- `/save-session, /resume-session, /sessions` - Session management
- `/instinct-export/import/status` - Instincts (long-term memory)
- `/checkpoint` - Checkpoints

**Quality commands**:
- `/quality-gate` - Quality gates
- `/verify` - Verification
- `/eval` - Evaluation

---

## ✅ IMMEDIATE ACTIONS

### 1. Create Symlinks (5 minutes)

```powershell
cd Z:\Ti\Ti-learning-lab
New-Item -ItemType Directory -Path .claude -Force
cmd /c mklink /J ".claude\agents" "Z:\Ti\best_source\agents"
cmd /c mklink /J ".claude\commands" "Z:\Ti\best_source\commands"
cmd /c mklink /J ".claude\skills" "Z:\Ti\best_source\skills"
```

### 2. Verify Access (1 minute)

```bash
ls .claude/agents/    # Should show 30 agents
ls .claude/commands/  # Should show 60 commands
ls .claude/skills/    # Should show 130+ skills
```

### 3. Test với Claude Code (5 minutes)

```bash
# Trong Claude Code:
@architect "Analyze current Ti CLI architecture"
@planner "Plan Phase 1 of AGENTS.md migration"
```

### 4. Read Top Skills (30 minutes)

```bash
# Must-read skills
cat .claude/skills/agentic-engineering/SKILL.md
cat .claude/skills/continuous-learning-v2/SKILL.md
cat .claude/skills/prompt-optimizer/SKILL.md
cat .claude/skills/mcp-server-patterns/SKILL.md
cat .claude/skills/golang-patterns/SKILL.md
```

---

## 🎉 KEY TAKEAWAYS

1. **Don't install aitmpl** - Best source đã đủ và chất lượng hơn
2. **Symlink để reuse** - Single source of truth
3. **Use orchestrate workflow** - Đã có sẵn pattern
4. **Skills là gold** - 130+ patterns sẵn có
5. **Phase 4 integrates** - Best source thành Ti's plugins
6. **Brain enhances** - Mỗi agent thông minh hơn nhờ Ti Brain

**Best source + Ti CLI = Perfect Match** 🎯

---

**Last Updated**: 2026-04-28
**Source**: Z:\Ti\best_source\
