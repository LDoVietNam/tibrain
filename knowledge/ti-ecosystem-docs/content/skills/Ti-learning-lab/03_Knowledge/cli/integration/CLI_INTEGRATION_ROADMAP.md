---
tags: ["tibrain", "skill", "testing", "documentation", "cli"]
scopes: ["code", "cli", "tibrain"]
last_updated: 2026-05-22
---
# Ti CLI Integration Roadmap

> **Last Updated**: 2026-05-05
> **Version**: 1.0.0
> **Purpose**: Define integration roadmap for adopting features from other AI coding CLIs

## 📊 Executive Summary

Based on analysis of 8 major AI coding CLIs (Claude Code, Qwen Code, OpenCode, Codex CLI, Kilo CLI, Gemini CLI, Qoder CLI, Aider), Ti CLI can adopt **50+ features** across **3 priority levels**.

**Current Ti CLI Status:**
- ✅ Multi-provider support (22+ providers via Router)
- ✅ Sub-agent orchestration (unique feature)
- ✅ Multi-agent orchestrator (Ti Claw plugin)
- ✅ MCP integration
- ✅ Plugin system (old + new PACK 1)
- ✅ Knowledge graph memory
- ✅ CLI Registry integration

**Key Gaps:**
- ❌ Permission modes (security)
- ❌ Sandbox modes (security)
- ❌ Session management (resume/fork/export)
- ❌ Budget control (cost management)
- ❌ System prompt override
- ❌ Git integration (commit/diff/review)
- ❌ Debug mode with categories
- ❌ Health check (doctor)
- ❌ Input/output format control (JSON/stream-JSON)

---

## 🎯 Priority Matrix

### P0 — Must Have (Core Identity) - 12 Features

| # | Feature | Source CLI | Complexity | Impact | Est. Effort |
|---|---------|-----------|------------|--------|-------------|
| 1 | Permission modes (4+) | Claude (5), Qwen (4), Codex (4) | Medium | High | 2-3 weeks |
| 2 | Session management (continue/resume/fork) | All 8 CLIs | Medium | High | 2-3 weeks |
| 3 | Model selection + fallback | Claude (`--model` + `--fallback-model`) | Low | High | 1 week |
| 4 | Budget control | Claude (`--max-budget-usd`) | Medium | High | 2 weeks |
| 5 | System prompt override + append | Claude, Qwen | Low | Medium | 1 week |
| 6 | Tool whitelist/blacklist | Claude (`--allowed-tools`, `--disallowed-tools`) | Low | High | 1 week |
| 7 | Git integration (commit/diff/review) | Aider, Claude, OpenCode | Medium | High | 2-3 weeks |
| 8 | Debug mode with categories | Claude (`-d [filter]`) | Medium | Medium | 1-2 weeks |
| 9 | Health check (doctor) | Claude (`doctor`) | Low | High | 1 week |
| 10 | Headless/non-interactive mode | All 8 CLIs | Low | Medium | 1 week |
| 11 | Input/output format control (JSON/stream-JSON) | Claude, Qwen, Gemini | Low | Medium | 1 week |
| 12 | Config precedence (7 levels) | OpenCode | High | High | 3-4 weeks |

**Total P0 Effort: 17-23 weeks**

### P1 — Should Have (Competitive Parity) - 17 Features

| # | Feature | Source CLI | Complexity | Impact | Est. Effort |
|---|---------|-----------|------------|--------|-------------|
| 13 | Sandbox modes | Codex (3 levels), Qwen (Docker) | High | High | 3-4 weeks |
| 14 | Session export/import | OpenCode, Kilo | Medium | Medium | 2 weeks |
| 15 | Stats/usage tracking | OpenCode, Qoder, Aider | Medium | Medium | 2 weeks |
| 16 | Custom commands (YAML) | Qwen (markdown+YAML), Qoder | Low | Medium | 1 week |
| 17 | Skills system | Qwen (loop/qc-helper/review) | Medium | High | 2-3 weeks |
| 18 | Background tasks | Qoder (`/bashes`) | Medium | Medium | 2 weeks |
| 19 | Chrome integration | Claude (`--chrome`) | Medium | Low | 1-2 weeks |
| 20 | Server mode + web UI | OpenCode, Kilo | High | Medium | 3-4 weeks |
| 21 | ACP mode | Qwen, OpenCode, Kilo | Low | Low | 1 week |
| 22 | Git worktree | Claude (`-w/--worktree`) | Low | Medium | 1 week |
| 23 | PR integration | Claude (`--from-pr`), OpenCode, Kilo | Medium | Medium | 2 weeks |
| 24 | Image input | Codex (`-i/--image`), Aider | Medium | Medium | 2 weeks |
| 25 | Vim mode | Qwen, Aider | Low | Low | 1 week |
| 26 | Checkpointing | Qwen, Gemini | Medium | Medium | 2 weeks |
| 27 | Token caching | Gemini | Low | Medium | 1 week |
| 28 | Local provider support | Codex (Ollama/LM Studio) | Medium | Medium | 2 weeks |
| 29 | Effort level | Claude (`--effort` 4 levels) | Low | High | 1 week |

**Total P1 Effort: 29-40 weeks**

### P2 — Nice to Have (Differentiation) - 20 Features

| # | Feature | Source CLI | Complexity | Impact | Est. Effort |
|---|---------|-----------|------------|--------|-------------|
| 30 | Arena mode | Qwen (subagents) | Medium | Medium | 2-3 weeks |
| 31 | MCP inline invocation | Gemini (`@<name>`) | Low | Medium | 1 week |
| 32 | Cron jobs | Qwen (loop skill) | Medium | Low | 2 weeks |
| 33 | Memory system | Qwen, Qoder (`/memory`) | High | High | 3-4 weeks |
| 34 | Quest orchestrator | Qoder (`/quest`) | High | High | 3-4 weeks |
| 35 | Dual-model architecture | Aider (main + weak/editor) | Medium | High | 2-3 weeks |
| 36 | Lint auto-fix | Aider (`/lint`) | Medium | Medium | 2 weeks |
| 37 | Test auto-add | Aider (`/test`) | Medium | Medium | 2 weeks |
| 38 | Multimodal input | Gemini (PDF/image/sketch → app) | High | High | 4-5 weeks |
| 39 | Google Search grounding | Gemini | Medium | Medium | 2-3 weeks |
| 40 | mDNS discovery | OpenCode, Kilo | Medium | Low | 2 weeks |
| 41 | Remote connection | Codex, Kilo | High | Medium | 3-4 weeks |
| 42 | Context copy as markdown | Aider (`/copy-context`) | Low | Low | 1 week |
| 43 | Repository map | Aider (`/map`) | Medium | Medium | 2 weeks |
| 44 | Trusted folders | Gemini | Low | Medium | 1 week |
| 45 | Release tags | Gemini (`@preview/@latest/@nightly`) | Low | Low | 1 week |
| 46 | Voice input | Aider (`/voice`) | Medium | Low | 2 weeks |
| 47 | Tmux integration | Claude (`--tmux`) | Low | Low | 1 week |
| 48 | Portable mode | OpenCode | High | Medium | 3-4 weeks |
| 49 | JSON Schema validation | Claude (`--json-schema`) | Low | Medium | 1 week |
| 50 | Partial messages | Claude, Qwen | Low | Low | 1 week |

**Total P2 Effort: 36-50 weeks**

---

## 🚀 Implementation Phases

### Phase 1: Security Foundation (Weeks 1-8)

**Goal:** Establish production-grade security model

**Features:**
1. Permission modes (P0 #1)
2. Tool whitelist/blacklist (P0 #6)
3. Sandbox modes (P1 #13)

**Implementation Plan:**
```bash
# Week 1-2: Permission modes
- Implement 5 permission modes: acceptEdits/auto/bypassPermissions/default/dontAsk
- Add --allowed-tools and --disallowed-tools flags
- Integrate with existing permission system

# Week 3-4: Tool whitelist/blacklist
- Implement tool permission checking
- Add profile-based tool restrictions
- Integrate with MCP tool registry

# Week 5-8: Sandbox modes
- Implement 3 sandbox levels: read-only/workspace-write/danger-full-access
- Add Docker sandbox support (Qwen pattern)
- Add sandbox configuration via YAML
```

**Deliverables:**
- `apps/cli/internal/permissions/` package
- `apps/cli/internal/sandbox/` package
- Updated `apps/cli/cmd/root.go` with permission flags
- Documentation in `docs/SECURITY.md`

### Phase 2: Session & Cost Management (Weeks 9-16)

**Goal:** Enable session persistence and cost control

**Features:**
1. Session management (P0 #2)
2. Budget control (P0 #4)
3. Stats/usage tracking (P1 #14)
4. Session export/import (P1 #13)

**Implementation Plan:**
```bash
# Week 9-10: Session management
- Implement continue/resume/fork commands
- Add custom session ID support
- Add session naming

# Week 11-12: Budget control
- Implement --max-budget-usd flag
- Add budget enforcement in Router
- Add cost tracking per session

# Week 13-14: Stats/usage tracking
- Implement stats command
- Add usage tracking per provider
- Add cost analytics

# Week 15-16: Session export/import
- Implement session export (JSON format)
- Implement session import
- Add session sharing support
```

**Deliverables:**
- `apps/cli/internal/session/` package
- `apps/cli/internal/budget/` package
- Updated `apps/cli/cmd/session.go`
- Updated `apps/cli/cmd/stats.go`
- Database schema for session storage

### Phase 3: Git Integration (Weeks 17-24)

**Goal:** Deep Git workflow integration

**Features:**
1. Git integration (P0 #7)
2. Git worktree (P1 #22)
3. PR integration (P1 #23)

**Implementation Plan:**
```bash
# Week 17-19: Git integration
- Implement git commit/diff/review commands
- Add git operations integration
- Add git-aware context packing

# Week 20-21: Git worktree
- Implement --worktree flag
- Add worktree management
- Add parallel development support

# Week 22-24: PR integration
- Implement --from-pr flag
- Add PR fetching and review
- Add GitHub Actions integration
```

**Deliverables:**
- `apps/cli/internal/git/` package
- Updated `apps/cli/cmd/git.go`
- Updated `apps/cli/cmd/pr.go`
- Integration with GitHub API

### Phase 4: Developer Experience (Weeks 25-32)

**Goal:** Improve developer workflow and debugging

**Features:**
1. Debug mode with categories (P0 #8)
2. Health check (P0 #9)
3. System prompt override (P0 #5)
4. Input/output format control (P0 #11)
5. Effort level (P1 #29)

**Implementation Plan:**
```bash
# Week 25-26: Debug mode with categories
- Implement -d [filter] flag
- Add debug file output
- Add debug category filtering

# Week 27: Health check
- Implement doctor command
- Add system diagnostics
- Add dependency checking

# Week 28: System prompt override
- Implement --system-prompt flag
- Implement --append-system-prompt flag
- Add system prompt management

# Week 29-30: Input/output format control
- Implement --input-format flag
- Implement --output-format flag
- Add JSON/stream-JSON support

# Week 31-32: Effort level
- Implement --effort flag (4 levels)
- Add effort-based model selection
- Add cost/quality tradeoff control
```

**Deliverables:**
- Updated `apps/cli/cmd/debug.go`
- Updated `apps/cli/cmd/doctor.go`
- Updated `apps/cli/cmd/root.go` with new flags
- Documentation in `docs/DEBUGGING.md`

### Phase 5: Advanced Features (Weeks 33-52)

**Goal:** Add competitive differentiation features

**Features:**
1. Skills system (P1 #17)
2. Background tasks (P1 #18)
3. Dual-model architecture (P2 #35)
4. Quest orchestrator (P2 #34)
5. Memory system (P2 #33)

**Implementation Plan:**
```bash
# Week 33-35: Skills system
- Implement skills command
- Add bundled skills (loop/qc-helper/review)
- Add custom skill support

# Week 36-37: Background tasks
- Implement /bashes command
- Add task queue management
- Add async execution

# Week 38-40: Dual-model architecture
- Implement main + weak/editor model separation
- Add model routing based on task type
- Add cost optimization

# Week 41-44: Quest orchestrator
- Implement /quest command
- Add workflow orchestration
- Add sub-agent coordination

# Week 45-52: Memory system
- Implement /memory command
- Add persistent memory storage
- Add memory retrieval and context injection
```

**Deliverables:**
- `apps/cli/internal/skills/` package
- `apps/cli/internal/tasks/` package
- `apps/cli/internal/dualmodel/` package
- `apps/cli/internal/quest/` package
- `apps/cli/internal/memory/` package

---

## 🎨 Unique Opportunities (Ti-Only Features)

Based on CLI-FEATURE-MATRIX.md analysis, these are features **NO existing CLI has** but Ti could pioneer:

### 1. Phase Routing (P0 - Unique)

**Description:** Route requests through sequential phases: `scan → plan → spec → implement → review → summarize`

**Benefits:**
- Structured multi-phase workflows
- Different models/tools per phase
- Complex task planning before execution

**Implementation:**
```go
type PhaseRouter struct {
    Phases []Phase
    CurrentPhase int
}

type Phase struct {
    Name string
    Model string
    Tools []string
    Execute func(ctx Context) error
}
```

**Est. Effort:** 4-5 weeks

### 2. Domain Routing (P0 - Unique)

**Description:** Route by domain: `code → best code model`, `review → best review model`, `plan → best planning model`

**Benefits:**
- Optimal model selection per task type
- Domain-specialized routing
- Cost optimization

**Implementation:**
```go
type DomainRouter struct {
    Domains map[string]DomainConfig
}

type DomainConfig struct {
    Model string
    Provider string
    Capabilities []string
}
```

**Est. Effort:** 3-4 weeks

### 3. Multi-Backend Orchestration (P1 - Unique)

**Description:** Orchestrate Claude + Qwen + Codex + Ticlaw in a single workflow

**Benefits:**
- "Best of breed" approach per sub-task
- Cross-provider orchestration
- Provider fallback chains

**Implementation:**
```go
type MultiBackendOrchestrator struct {
    Backends []Backend
    Workflow Workflow
}

type Workflow struct {
    Steps []WorkflowStep
}

type WorkflowStep struct {
    Backend string
    Task string
    DependsOn []string
}
```

**Est. Effort:** 5-6 weeks

### 4. Multi-Source Auth Loader (P1 - Unique)

**Description:** Load credentials from: env vars, token files, cookies, OAuth, external credential loaders

**Benefits:**
- Unified auth abstraction
- Enterprise auth integration
- Flexible credential management

**Implementation:**
```go
type AuthLoader interface {
    Load(key string) (string, error)
}

type MultiSourceAuthLoader struct {
    Loaders []AuthLoader
    Precedence []string // env, file, cookie, oauth, external
}
```

**Est. Effort:** 2-3 weeks

### 5. Capability Inventory (P2 - Unique)

**Description:** Maintain inventory of: CLIs, providers, tools, packages, adapters

**Benefits:**
- Self-aware capability registry
- Intelligent routing based on available capabilities
- Dynamic discovery

**Implementation:**
```go
type CapabilityInventory struct {
    CLIs []CLIInfo
    Providers []ProviderInfo
    Tools []ToolInfo
    Packages []PackageInfo
}

type CLIInfo struct {
    Name string
    Version string
    Capabilities []string
}
```

**Est. Effort:** 3-4 weeks

### 6. External CLI as Backend Adapter (P2 - Unique)

**Description:** Treat any CLI as a backend: `git`, `docker`, `kubectl`, `terraform`

**Benefits:**
- Universal automation layer
- CLI-first architecture
- Extensible backend system

**Implementation:**
```go
type ExternalCLIAdapter struct {
    CLI string
    Args []string
    Env map[string]string
}

func (a *ExternalCLIAdapter) Execute(ctx Context, task string) error {
    cmd := exec.Command(a.CLI, a.Args...)
    cmd.Env = buildEnv(a.Env)
    // Execute and return result
}
```

**Est. Effort:** 4-5 weeks

---

## 📈 Success Metrics

### Phase 1: Security Foundation
- [ ] 5 permission modes implemented and tested
- [ ] 3 sandbox levels operational
- [ ] Zero security vulnerabilities in new code
- [ ] 90%+ test coverage for permission system

### Phase 2: Session & Cost Management
- [ ] Session resume/fork/export working
- [ ] Budget control enforcement active
- [ ] Usage tracking accurate to 99%
- [ ] Cost variance < 5% vs actual

### Phase 3: Git Integration
- [ ] Git commit/diff/review commands functional
- [ ] Worktree management working
- [ ] PR integration with GitHub API
- [ ] 95%+ success rate for git operations

### Phase 4: Developer Experience
- [ ] Debug mode with category filtering
- [ ] Doctor command covers all dependencies
- [ ] System prompt override functional
- [ ] JSON/stream-JSON output validated

### Phase 5: Advanced Features
- [ ] Skills system with 3 bundled skills
- [ ] Background tasks queue operational
- [ ] Dual-model architecture routing
- [ ] Quest orchestrator with 5 workflows

### Unique Features
- [ ] Phase routing with 6 phases
- [ ] Domain routing with 10+ domains
- [ ] Multi-backend orchestration with 4 providers
- [ ] Multi-source auth loader with 5 sources
- [ ] Capability inventory auto-discovery
- [ ] External CLI adapter with 10+ CLIs

---

## 🔄 Rollout Strategy

### Alpha (Weeks 1-8)
- Internal testing only
- Feature flags for all new features
- Limited to core team
- Focus on security foundation

### Beta (Weeks 9-24)
- Invite 10-20 beta users
- Gather feedback on session/cost/git features
- Iterate based on feedback
- Documentation completion

### RC (Weeks 25-40)
- Public beta
- Stability focus
- Performance optimization
- Security audit

### GA (Weeks 41-52)
- General availability
- Marketing launch
- Enterprise support
- Long-term support commitment

---

## 📚 References

- **CLI Feature Matrix**: `CLI UPGRADE/docs/CLI-FEATURE-MATRIX.md`
- **Claude Code**: `Z:\01_PROJECTS\SJ-learning-lab\05_Repositories\claude\claude-code-deobfuscated\`
- **OpenCode**: `Ti-learning-lab/07_Repositories/opencode/opencode-dev/`
- **Sub-Agent Guide**: `Ti-learning-lab/03_Knowledge/CLI/SUB_AGENT_GUIDE.md`
- **Architecture**: `Ti-learning-lab/03_Knowledge/CLI/ARCHITECTURE.md`

---

## 🎯 Summary

**Total Features to Integrate:** 50+ features
**Total Estimated Effort:** 82-113 weeks (1.5-2.5 years)
**Recommended Timeline:** 3-year rollout (1 year P0, 1 year P1, 1 year P2 + unique features)

**Key Differentiators:**
1. Sub-agent orchestration (already implemented)
2. Multi-agent orchestrator (already implemented)
3. Phase routing (unique opportunity)
4. Domain routing (unique opportunity)
5. Multi-backend orchestration (unique opportunity)

**Competitive Position:**
- **Before Integration:** 4.8/5.0 (current score)
- **After P0:** 4.9/5.0 (security + session management)
- **After P1:** 5.0/5.0 (competitive parity)
- **After P2 + Unique:** 5.5/5.0 (market leader)
