# Ti CLI Plan - Multi-CLI Orchestration

> **Purpose**: Hoàn thiện Ti CLI thành **orchestrator cho nhiều CLI agents** (Claude Code, Codex, Gemini, Junie, Kilo, Cline...) với Router làm LLM gateway chung
> **Date**: 2026-05-05
> **Status**: Planning
> **Scope**: Track C (refined) — Ti CLI completion + multi-CLI orchestration
> **Sister Plans**: `TI_CLAW_MASTER_PLAN.md` (agent framework), `ROUTER_PLAN.md` (LLM gateway)

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Current State](#2-current-state)
3. [Vision & Goals](#3-vision--goals)
4. [CLI Registry System](#4-cli-registry-system)
5. [Config Folder Standardization](#5-config-folder-standardization)
6. [Router Integration](#6-router-integration)
7. [Sub-Agent Orchestration](#7-sub-agent-orchestration)
8. [Parallel Execution](#8-parallel-execution)
9. [Claude Code Native Features](#9-claude-code-native-features)
10. [Architecture](#10-architecture)
11. [Implementation Phases](#11-implementation-phases)
12. [Key Risks](#12-key-risks)
13. [References](#13-references)

---

## 1. Executive Summary

### Vision
Ti CLI trở thành **orchestrator trung tâm** quản lý nhiều external CLIs (Claude Code, Codex, Gemini, Cline, Junie, Kilo, OpenCode, Kimi, Cursor...) như sub-agents, kết nối tất cả vào Ti Router để dùng chung pool of AI models.

### Three Pillars

**Pillar 1 — Registry**
- Junie-style registry JSON: liệt kê available CLIs + distribution (binary/npx/uvx)
- Tự động discover, install, update external CLIs
- Metadata: version, license, ACP support, config path

**Pillar 2 — Config Standardization**
- Mỗi CLI có folder config chuẩn: `Z:\02_CORE\_cli\.config\{cli_name}\`
- Base guideline: `AGENTS.md` + overrides riêng: `{CLI}/{CLI}.md`
- Junction từ `Z:/Ti/.config` → `Z:/02_CORE/_cli/.config` (đã có)

**Pillar 3 — Multi-CLI Orchestration**
- Ti CLI làm supervisor, gọi external CLIs như sub-agents
- Tương tự CAO (CLI Agent Orchestrator) pattern
- Parallel execution (goroutines hoặc tmux)
- Shared Router = 1 LLM gateway cho all sub-agents

### Key Decisions
1. **Ti CLI orchestrates external CLIs** (không thay thế chúng)
2. **Registry pattern** tương tự junie, tích hợp vào Ti CLI
3. **Router as single LLM gateway** — all CLIs point tới :1807
4. **Sub-agent via process** (stdin/stdout IPC) — không cần daemon
5. **Parallel bằng Go goroutines** (first) + optional tmux for interactive
6. **Claude Code features** port vài cái làm native: subagents, skills, plan mode
7. **ACP protocol support** (từ junie registry) — standard agent communication

---

## 2. Current State

### 2.1 Ti CLI (`apps/cli/`)
- 74% migrated (43/58 modules)
- Microkernel + Plugin architecture (gRPC)
- Có `agentcore`, `agent`, `routeagent` — foundation cho sub-agent orchestration
- Có `mcp` bridge
- Easy intent commands: `ti fix`, `ti plan`, `ti review`, `ti optimize`...

### 2.2 `Z:\02_CORE\_cli\` Setup
- Home cho 10+ external CLIs (claude, codex, gemini, qwen, kilo, opencode, devin, qwen, junie, cline, qodercli)
- Config folders tại `.config/{cli_name}/`
- `AGENTS.md` chung (base guidelines) + `{CLI}/{CLI}.md` (overrides)
- `index.json` — master list of tools
- `mcp-router-bridge.ps1` — PowerShell MCP bridge tới Router (example)
- Junction: `Z:/Ti/.config` → `Z:/02_CORE/_cli/.config`
- Junction: `Z:/02_CORE/_cli/.config/devin` → real `.devin/`

### 2.3 CAO — CLI Agent Orchestrator (Reference)
- Located: `Z:\02_CORE\_cli\.config\cli-agent-orchestrator\`
- Python tool (AWS Labs open source)
- Multi-agent via tmux sessions
- Providers: kiro_cli, q_cli, claude_code, codex
- Orchestration patterns: Handoff, Assign, Send Message
- FastAPI :9889 + MCP server
- **Có thể dùng làm reference nhưng không chạy trực tiếp** (Ti CLI sẽ native Go)

### 2.4 Junie Registry (Reference)
- `Z:\02_CORE\_cli\.config\junie-main\registry-nightly.json` — ~50 CLI agents với ACP support
- `registry-experimental.json` — experimental versions
- Format: id, name, version, description, repository, license, distribution (binary/npx/uvx), icon
- Distribution types:
  - `binary`: platform-specific archives
  - `npx`: npm package (cần Node.js)
  - `uvx`: Python package (cần uv)

### 2.5 Registered CLIs (current)
| CLI | Exe | Config | Status | Notes |
|-----|-----|--------|--------|-------|
| claude | `claude.exe` | `.config/claude/` | active | Junction từ APPDATA |
| codex | `codex.exe` | `.config/codex/` | template | XDG supported |
| gemini | `gemini.ps1` | `.config/gemini/` | template | PowerShell script |
| kilo | `kilo.exe` | `.config/kilo/` | template | XDG supported |
| opencode | `opencode.exe` | `.config/opencode/` | template | |
| qwen | `qwen.ps1` | `.config/qwen/` | template | PowerShell script |
| devin | `devin` | `.config/devin/` | active | Real junction |
| ti | `ti.exe` | `.config/ti/` | active | Go native |
| qodercli | `qodercli.exe` | — | standalone | No config file |

---

## 3. Vision & Goals

### 3.1 End-State Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                        User                                    │
│  ti "phân tích codebase này và tạo PR fix security issues"     │
└────────────────────┬───────────────────────────────────────────┘
                     │
              ┌──────▼──────┐
              │   Ti CLI    │   ← Supervisor agent (Go)
              │ (ti.exe)    │
              └──────┬──────┘
                     │
       ┌─────────────┼─────────────┐
       │             │             │
  ┌────▼───┐    ┌────▼───┐    ┌───▼────┐
  │Sub-    │    │Sub-    │    │Sub-    │
  │Agent 1 │    │Agent 2 │    │Agent 3 │
  │        │    │        │    │        │
  │claude  │    │codex   │    │gemini  │  ← External CLIs as workers
  │ (code  │    │ (refact│    │ (OCR/  │
  │  rev)  │    │  oring)│    │  vision)│
  └────┬───┘    └────┬───┘    └───┬────┘
       │             │             │
       └─────────────┼─────────────┘
                     │
              ┌──────▼──────┐
              │ Ti Router   │   ← Single LLM gateway
              │  (:1807)    │     28 providers, cache, routing
              └──────┬──────┘
                     │
       ┌─────────────┼─────────────┬────────────┐
       │             │             │            │
   ┌───▼───┐    ┌───▼────┐    ┌───▼────┐   ┌───▼────┐
   │OpenAI │    │Anthropic│   │Google │   │Groq    │
   │       │    │         │   │Gemini │   │        │
   └───────┘    └─────────┘   └───────┘   └────────┘
```

### 3.2 Goals

**G1: CLI Registry** (High)
- Native Go implementation of registry pattern
- Command: `ti registry list`, `ti registry install <cli>`, `ti registry update`
- Support binary/npx/uvx distributions
- Auto-discover installed CLIs

**G2: Config Standardization** (High)
- Chuẩn hóa `.config/{cli_name}/` cho tất cả CLIs
- Templates cho mỗi CLI: `.config/{cli_name}/{CLI}.md`
- Env vars: `TI_CLI_CONFIG_HOME`, `XDG_CONFIG_HOME`
- Config inheritance: default → AGENTS.md → {CLI}/{CLI}.md → env → flags

**G3: Router Integration** (High)
- Mỗi CLI config có env var pointing tới Router: `OPENAI_BASE_URL=http://localhost:1807/v1`
- Setup script: `ti cli setup <name>` — auto-configure để dùng Router
- Verify: `ti cli ping <name>` — check CLI có thể gọi Router

**G4: Sub-Agent Orchestration** (Critical)
- Ti CLI gọi external CLI như sub-agent qua stdin/stdout
- Agent profile: role, allowedTools, context
- Orchestration patterns:
  - **Handoff**: sync, wait for completion
  - **Assign**: async, fire-and-forget
  - **SendMessage**: message to existing agent
- Config tại `content/agents/` hoặc inline YAML

**G5: Parallel Execution** (High)
- Run multiple sub-agents concurrently
- Go goroutines + channels cho IPC
- Optional: tmux-backed sessions for interactive inspection
- Progress UI: live log aggregation

**G6: Claude Code Native Features** (Medium)
- Port Claude Code features làm native Ti:
  - Sub-agents với custom context
  - Skills system
  - Plan mode
  - Session persistence
  - Tool allowlisting
- Goal: Giảm dependency vào Claude Code CLI binary khi không cần

**G7: ACP Protocol Support** (Medium)
- Agent Client Protocol — standard từ junie registry
- Ti CLI có thể connect tới any ACP-compliant CLI
- Uniform communication layer

---

## 4. CLI Registry System

### 4.1 Registry Format

**Inspiration**: `Z:\02_CORE\_cli\.config\junie-main\registry-nightly.json`

**Ti CLI Registry**: `content/registry/cli-registry.json`

```json
{
  "version": "1.0.0",
  "last_updated": "2026-05-05",
  "agents": [
    {
      "id": "claude",
      "name": "Claude Code",
      "version": "latest",
      "description": "Anthropic's Claude Code CLI",
      "website": "https://claude.ai/code",
      "license": "proprietary",
      "category": "code",
      "distribution": {
        "binary": {
          "windows-x86_64": {
            "exe": "C:\\Users\\MIN\\AppData\\Local\\Programs\\claude\\claude.exe"
          }
        },
        "npx": {
          "package": "@anthropic-ai/claude-code@latest"
        }
      },
      "config": {
        "dir": "Z:/02_CORE/_cli/.config/claude",
        "overrides": "CLAUDE.md",
        "router_compatible": true,
        "env_vars": {
          "ANTHROPIC_API_URL": "http://localhost:1807/v1",
          "ANTHROPIC_API_KEY": "$TI_ROUTER_KEY"
        }
      },
      "acp_support": true,
      "sub_agent_mode": "stdio",
      "capabilities": ["code-review", "documentation", "analysis"]
    },
    {
      "id": "codex",
      "name": "Codex CLI",
      "version": "0.12.0",
      "license": "Apache-2.0",
      "category": "code",
      "distribution": {
        "binary": {
          "windows-x86_64": {
            "archive": "https://github.com/zed-industries/codex-acp/releases/download/v0.12.0/codex-acp-0.12.0-x86_64-pc-windows-msvc.zip",
            "cmd": "codex-acp.exe"
          },
          "npx": {
            "package": "@zed-industries/codex-acp@0.12.0"
          }
        }
      },
      "config": {
        "dir": "Z:/02_CORE/_cli/.config/codex",
        "overrides": "CODEX.md",
        "router_compatible": true,
        "env_vars": {
          "OPENAI_BASE_URL": "http://localhost:1807/v1"
        }
      },
      "acp_support": true,
      "capabilities": ["code-generation", "refactoring"]
    }
  ],
  "categories": [
    "code",
    "multimodal",
    "performance",
    "specialized",
    "management"
  ]
}
```

### 4.2 Registry Commands

```bash
# List all registered CLIs
ti registry list
ti registry list --category code
ti registry list --installed

# Show CLI details
ti registry show claude

# Install a CLI
ti registry install codex
ti registry install codex --version 0.12.0
ti registry install codex --method npx    # force npx method

# Update CLI
ti registry update claude
ti registry update --all

# Sync registry từ upstream sources (junie, custom)
ti registry sync
ti registry sync --source junie-nightly

# Add custom CLI
ti registry add ./my-custom-cli.json
```

### 4.3 Registry Data Sources

**Primary**: Local `content/registry/cli-registry.json`

**Secondary (optional sync)**:
- Junie upstream: `https://cdn.agentclientprotocol.com/registry/v1/latest/`
- Custom: local JSON files
- User-maintained

**Sync strategy**:
- On-demand: `ti registry sync`
- Never auto-sync (user controls)
- Merge strategy: upstream metadata + local overrides

### 4.4 Category Taxonomy

Từ `CLI_AGENTS.md`:

| Category | Description | Example CLIs |
|----------|-------------|--------------|
| **code** | Code generation, review, refactoring | claude, codex, cline, aider |
| **multimodal** | Vision, image, OCR | gemini, claude |
| **performance** | Fast, cost-optimized | qwen, deepseek, groq |
| **specialized** | Domain-specific | junie (JetBrains), kilo, opencode |
| **management** | Project orchestration | ti (self) |

---

## 5. Config Folder Standardization

### 5.1 Current Structure

```
Z:\02_CORE\_cli\.config\
├── AGENTS.md               # Base guidelines (source of truth)
├── CLI_AGENTS.md            # CLI catalog docs
├── PATH.md                 # Path references
├── index.json              # Tool registry
├── env.ps1                 # Shared env vars
├── claude\CLAUDE.md        # Claude overrides
├── codex\CODEX.md          # Codex overrides
├── gemini\GEMINI.md
├── qwen\QWEN.md
├── kilo\KILO.md
├── opencode\OPENCODE.md
├── junie\JUNIE.md
├── qodercli\QODERCLI.md
├── ti\
│   ├── TI.md               # Ti overrides (TODO: create)
│   └── config.yaml          # Ti CLI config
├── devin\                  # Junction to real .devin/
├── cline\
├── shared\                 # Shared configs
└── api\                    # API configs
```

### 5.2 Config Inheritance

**Load order (low → high priority)**:
1. Built-in defaults
2. `AGENTS.md` (shared baseline)
3. `{CLI}/{CLI}.md` (per-CLI overrides)
4. Environment variables (`TI_CLI_*`, `ANTHROPIC_API_KEY`, etc.)
5. CLI command-line flags

**Ti CLI responsibility**:
- Load + merge configs cho each sub-agent invocation
- Inject to sub-agent via env vars / flags
- Provide single source of truth

### 5.3 Standardization Tasks

**Task 1**: Tạo template cho mỗi CLI

```markdown
# {CLI_NAME} Configuration Overrides

> Base guidelines: `../AGENTS.md`
> This file: Per-CLI overrides for {CLI_NAME}

## Router Integration
- API Endpoint: http://localhost:1807/v1
- Env var: {CLI_ENV_VAR_NAME}=http://localhost:1807/v1
- Auth: $TI_ROUTER_KEY

## Agent Profile
- Role: [code|review|multimodal|etc.]
- Allowed tools: [read, write, bash, mcp]
- Context size: {max_tokens}

## Custom Behaviors
- ...
```

**Task 2**: Wire config loading trong Ti CLI

```go
// apps/cli/internal/clihost/config.go
type SubAgentConfig struct {
    Name         string            `yaml:"name"`
    Executable   string            `yaml:"executable"`
    ConfigDir    string            `yaml:"config_dir"`
    Overrides    string            `yaml:"overrides"` // {CLI}.md
    EnvVars      map[string]string `yaml:"env_vars"`
    RouterKey    string            `yaml:"router_key"`
    Role         string            `yaml:"role"`
    AllowedTools []string          `yaml:"allowed_tools"`
}

func LoadSubAgentConfig(name string) (*SubAgentConfig, error) {
    // 1. Load base AGENTS.md
    // 2. Apply {CLI}/{CLI}.md overrides
    // 3. Apply env vars
    // 4. Return merged config
}
```

**Task 3**: Setup commands

```bash
# Bootstrap config folder for new CLI
ti cli init <name>           # Creates .config/{name}/{NAME}.md template
ti cli init codex --template code

# Verify config
ti cli validate <name>

# Show effective config
ti cli config <name> --show
```

---

## 6. Router Integration

### 6.1 Goal

Tất cả external CLIs point tới **Ti Router :1807** thay vì direct API — để:
- Share cache across CLIs
- Central rate-limiting
- Cost tracking per CLI
- Fallback / failover
- Single auth management

### 6.2 Configuration per CLI

**Claude Code (claude.exe)**
```bash
# .config/claude/CLAUDE.md specifies:
ANTHROPIC_API_URL=http://localhost:1807/v1
ANTHROPIC_API_KEY=$TI_ROUTER_KEY
```

**Codex CLI (codex.exe)**
```bash
OPENAI_BASE_URL=http://localhost:1807/v1
OPENAI_API_KEY=$TI_ROUTER_KEY
```

**Gemini CLI**
```bash
GOOGLE_API_BASE=http://localhost:1807/v1beta
GOOGLE_API_KEY=$TI_ROUTER_KEY
```

**Generic pattern** (OpenAI-compatible CLIs):
```bash
# Works for most ACP CLIs
OPENAI_BASE_URL=http://localhost:1807/v1
OPENAI_API_KEY=$TI_ROUTER_KEY
```

### 6.3 Router MCP Bridge

**Existing**: `Z:\02_CORE\_cli\mcp-router-bridge.ps1`

Provides Router as MCP server cho CLIs không natively support Router:
- `router_models` — list models
- `router_chat` — chat completion
- `notion_tasks`, `notion_specs`, `notion_context` — Notion integration

**Ti CLI enhancement**:
- Port PowerShell → native Go: `apps/cli/internal/mcp/routerbridge/`
- Register as MCP server in CLI config
- Auto-serve khi CLI startup

### 6.4 Setup Automation

```bash
# Configure a CLI to use Router
ti cli setup claude
# → Updates ANTHROPIC_API_URL in .config/claude/env
# → Verifies Router reachable
# → Test với 1 sample call

# Setup all registered CLIs
ti cli setup --all

# Reset to direct API
ti cli reset claude
```

### 6.5 Router Compatibility Matrix

| CLI | API Format | Router Support | Setup Method |
|-----|-----------|----------------|--------------|
| claude | Anthropic | ✅ `/v1/messages` | Env vars |
| codex | OpenAI | ✅ `/v1/chat/completions` | Env vars |
| gemini | Google | ✅ `/v1beta` | Env vars |
| qwen | OpenAI-compat | ✅ | Env vars |
| cline | OpenAI-compat | ✅ | Config file |
| junie | Proprietary (ACP) | ❓ TBD | MCP bridge |
| kilo | OpenAI-compat | ✅ | Config file |
| opencode | Per-model | ⚠️ Per-provider | Config file |

---

## 7. Sub-Agent Orchestration

### 7.1 Patterns (từ CAO)

**Handoff — Synchronous**
```
User → Ti CLI → launch sub-agent "reviewer"
                ↓
                wait until done
                ↓
Ti CLI ← result from reviewer → User
```

**Assign — Asynchronous**
```
User → Ti CLI → spawn sub-agent "developer" (fire-and-forget)
                ↓
Ti CLI → User (sub-agent continues in background)
```

**Send Message — Inter-agent**
```
Agent A → Ti CLI → send message to Agent B
                   ↓
                   Agent B receives + responds
```

### 7.2 Agent Profile Format

**Location**: `content/agents/{name}.md`

```markdown
---
name: code_reviewer
role: reviewer
provider: claude
model: claude-sonnet-4-5
allowed_tools: [read, grep, analyze]
restricted_tools: [write, bash, delete]
max_tokens: 200000
router: true
---

# Code Reviewer Agent

You are a specialized code reviewer. Your role is to:
- Analyze code for bugs, security issues, performance
- Suggest improvements
- DO NOT modify code directly

## Context
{{context}}

## Task
{{task}}
```

### 7.3 Ti CLI Commands

```bash
# Run sub-agent synchronously
ti agent run code_reviewer --task "review this PR"
ti agent run code_reviewer --context ./src --task "security audit"

# Assign async
ti agent assign developer --task "implement feature X"

# Send message to running agent
ti agent send <session_id> "please also check test coverage"

# List running agents
ti agent list
ti agent list --status running

# Stop agent
ti agent stop <session_id>
ti agent stop --all

# Interactive: launch agent in foreground
ti agent launch code_reviewer --interactive
```

### 7.4 Implementation

**Process-based** (recommended):
```go
// apps/cli/internal/subagent/runner.go
type SubAgentRunner struct {
    Profile    *AgentProfile
    Provider   string      // "claude", "codex", etc.
    Config     *SubAgentConfig
    Session    *Session
}

func (r *SubAgentRunner) Run(ctx context.Context, task string) (*Result, error) {
    // 1. Resolve CLI executable from registry
    exe := r.Config.Executable

    // 2. Build env vars (Router endpoint, API key)
    env := r.buildEnv()

    // 3. Build args (prompt, context, tool restrictions)
    args := r.buildArgs(task)

    // 4. Spawn process
    cmd := exec.CommandContext(ctx, exe, args...)
    cmd.Env = env
    cmd.Stdin = bytes.NewBufferString(r.Profile.SystemPrompt)

    // 5. Capture output
    out, err := cmd.Output()

    // 6. Parse result (agent-specific)
    return r.parseResult(out)
}
```

**ACP-based** (for ACP-compliant CLIs):
```go
// apps/cli/internal/subagent/acp.go
type ACPClient struct {
    conn *acp.Connection
}

func (c *ACPClient) SendTask(task string) (*Result, error) {
    // Use ACP protocol (stdio JSON-RPC)
}
```

### 7.5 Context Preservation

**Challenge**: Sub-agent không biết về main agent's context

**Solutions**:
1. **Pass context explicitly**: Supervisor truyền relevant files, prior decisions
2. **Shared memory**: Knowledge graph + beads log accessible via MCP
3. **Session threads**: Mỗi agent có session ID persist trong SQLite

---

## 8. Parallel Execution

### 8.1 Use Cases

- Review + Implement: reviewer analyze code while developer implements feature
- Multi-model comparison: gửi same task cho claude + codex + gemini, compare outputs
- Fan-out research: 3 agents research different aspects đồng thời

### 8.2 Implementation Options

**Option A: Goroutines (Go native)** ⭐ recommended

```go
// apps/cli/internal/subagent/parallel.go
func RunParallel(ctx context.Context, tasks []SubAgentTask) []*Result {
    results := make([]*Result, len(tasks))
    var wg sync.WaitGroup
    sem := make(chan struct{}, maxConcurrent) // semaphore

    for i, task := range tasks {
        wg.Add(1)
        go func(idx int, t SubAgentTask) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()

            runner := NewSubAgentRunner(t.Profile, t.Config)
            res, _ := runner.Run(ctx, t.Task)
            results[idx] = res
        }(i, task)
    }

    wg.Wait()
    return results
}
```

**Option B: Tmux-backed** (for interactive inspection)

Tương tự CAO:
- Mỗi agent → separate tmux session
- User có thể attach để xem real-time
- Requires tmux installation

**Option C: Hybrid** (default goroutines + optional tmux)

```bash
# Default: goroutines
ti agent parallel reviewer developer tester --task "..."

# Interactive: tmux
ti agent parallel reviewer developer tester --task "..." --interactive
```

### 8.3 Aggregation Patterns

**Merge** — combine all outputs:
```bash
ti agent parallel claude codex gemini --task "review code" --merge
# Output: combined review from 3 models
```

**Vote** — majority wins:
```bash
ti agent parallel claude codex --task "is this bug?" --vote
# Output: consensus answer
```

**Race** — first to complete:
```bash
ti agent parallel claude-fast codex-fast --task "quick fix" --race
# Output: fastest response
```

### 8.4 Progress UI

**Live log**:
```
[14:20:01] [claude   ] Starting review of src/auth.go...
[14:20:03] [codex    ] Generating refactored version...
[14:20:05] [gemini   ] Analyzing image assets...
[14:20:12] [claude   ] Review complete (2 issues found)
[14:20:18] [codex    ] Refactoring done (120 lines changed)
[14:20:25] [gemini   ] OCR complete (15 images processed)

Summary: 3/3 agents completed in 24s
```

Implementation: goroutine-safe log aggregator + stdout rendering.

---

## 9. Claude Code Native Features

### 9.1 Rationale

Claude Code binary có nhiều features hay (subagents, skills, plan mode) nhưng:
- Proprietary, closed-source
- Phụ thuộc vào Anthropic
- Không flexibile bằng custom implementation

→ Port một số features làm native trong Ti CLI, giữ Claude Code làm "one of many" sub-agents.

### 9.2 Features to Port

**F1: Sub-agents** (already in plan)
- Ti có `agentcore`, `agent`, `routeagent` đã sẵn
- Bổ sung: agent profiles, tool restrictions

**F2: Skills** (Ti đã có `content/skills/` 100+)
- Skill definition: SKILL.md format
- Skill loading: `ti load-skill ./path`
- BM25 search cho skills
- Already in roadmap Phase A2 (Ti Claw)

**F3: Plan Mode**
- Read-only analysis, no modifications
- Shows plan, asks confirmation
- Ti CLI đã có `ti plan` command
- Enhancement: integrate with sub-agents

**F4: Session Persistence**
- CLI đã có `session` module (SQLite)
- Enhancement: cross-CLI session sharing

**F5: Tool Allowlisting**
- Per-agent restrictions
- `allowed_tools` / `restricted_tools` trong profile
- Enforcement layer trong Ti CLI

**F6: MCP Server Integration**
- CLI đã có `mcp` module
- Expose Ti CLI features as MCP servers
- Available to external CLIs

**F7: Hooks**
- Pre/post task hooks
- Git commit hooks
- CLI đã có `commitflow`

### 9.3 Native vs External Trade-off

**Native (Ti CLI implements)**:
- ✅ Full control
- ✅ Consistent UX
- ✅ Integrated with ecosystem
- ❌ More maintenance
- ❌ May lag behind external features

**External (spawn Claude Code CLI)**:
- ✅ Always latest features
- ✅ No maintenance burden
- ❌ Binary dependency
- ❌ Less integration

**Hybrid (Ti CLI primary, Claude Code fallback)**:
- For common workflows: Ti native
- For specialty (e.g., Claude's long context): spawn Claude Code as sub-agent
- User choice: `ti plan` (native) vs `ti agent run claude-plan` (external)

---

## 10. Architecture

### 10.1 High-Level Components

```
┌─────────────────────────────────────────────────────────────┐
│                       Ti CLI (ti.exe)                       │
│                                                             │
│  ┌──────────────┐  ┌─────────────┐  ┌──────────────────┐   │
│  │   Registry   │  │ Sub-Agent   │  │ Config Loader    │   │
│  │   Manager    │  │ Orchestrator│  │  (inheritance)   │   │
│  └──────┬───────┘  └──────┬──────┘  └────────┬─────────┘   │
│         │                 │                   │             │
│  ┌──────▼─────────────────▼───────────────────▼──────────┐  │
│  │             Agent Profile / Skill Engine              │  │
│  └────────────────────────┬──────────────────────────────┘  │
│                           │                                 │
│  ┌────────────────────────▼──────────────────────────────┐  │
│  │       Parallel Executor (goroutines + channels)       │  │
│  └────────────────────────┬──────────────────────────────┘  │
│                           │                                 │
│  ┌──────────────┐  ┌──────▼──────┐  ┌──────────────────┐   │
│  │ Process Pool │  │ ACP Client  │  │  MCP Bridge      │   │
│  │ (stdio IPC)  │  │ (JSON-RPC)  │  │  (Router proxy)  │   │
│  └──────┬───────┘  └──────┬──────┘  └────────┬─────────┘   │
└─────────┼─────────────────┼──────────────────┼─────────────┘
          │                 │                  │
     ┌────▼─────┐      ┌────▼──────┐     ┌────▼─────┐
     │ External │      │ ACP CLIs  │     │ Router   │
     │ CLIs     │      │ (junie,   │     │ :1807    │
     │ (claude, │      │ kilo,     │     │          │
     │  codex,  │      │ cursor,   │     │          │
     │  gemini) │      │  ...)     │     │          │
     └──────────┘      └───────────┘     └──────────┘
```

### 10.2 Module Layout (Ti CLI additions)

```
apps/cli/internal/
├── registry/              # NEW — CLI registry
│   ├── manager.go
│   ├── loader.go
│   ├── installer.go
│   └── types.go
├── subagent/              # NEW — Sub-agent orchestration
│   ├── runner.go
│   ├── parallel.go
│   ├── profile.go
│   ├── acp_client.go
│   ├── stdio_client.go
│   └── aggregator.go
├── clihost/               # NEW — CLI host / process management
│   ├── process.go
│   ├── env.go
│   └── config.go
├── agentcore/             # EXISTING — reuse
├── agent/                 # EXISTING — reuse
├── routeagent/            # EXISTING — reuse
└── mcp/                   # EXISTING — reuse, extend for router bridge
    └── routerbridge/      # NEW subpackage
```

### 10.3 Data Stores

- **CLI Registry**: `content/registry/cli-registry.json`
- **Agent Profiles**: `content/agents/*.md`
- **Skills**: `content/skills/*.md`
- **Sessions**: SQLite (CLI's `session` module)
- **Configs**: `Z:\02_CORE\_cli\.config\*`

### 10.4 External Dependencies

- Ti Router :1807 (required)
- External CLI binaries (optional, per use)
- tmux (optional, for interactive mode)
- Node.js / uv (optional, for npx/uvx CLIs)

---

## 11. Implementation Phases

### Phase CLI-0: Audit & Design (Week 1)
**Goal**: Plan chi tiết trước khi code

**Tasks**:
- Audit current CLI: `apps/cli/internal/` — list relevant modules
- Review CAO source code: patterns, edge cases
- Design registry schema (finalize JSON format)
- Design sub-agent interface (Go interfaces)
- Document process stdio IPC contract

**Gate criteria**:
- ✅ `ti-cli-design-doc.md` written
- ✅ Registry schema v1 frozen
- ✅ Sub-agent interface design reviewed

---

### Phase CLI-1: Registry Foundation
**Goal**: `ti registry list/show/install` working

**Tasks**:
- Create `internal/registry/` package
- Implement registry loader (JSON → Go structs)
- Implement `ti registry list/show` commands
- Create initial `content/registry/cli-registry.json` với 3 CLIs (claude, codex, gemini)
- Implement `ti registry install` (npx + binary methods)

**Gate criteria**:
- ✅ `ti registry list` hiển thị 3+ CLIs
- ✅ `ti registry install codex` download được binary
- ✅ Tests pass

---

### Phase CLI-2: Config Standardization
**Goal**: Config loading standardized cho all CLIs

**Tasks**:
- Implement `internal/clihost/config.go` — inheritance loader
- Template generator: `ti cli init <name>`
- Validation: `ti cli validate <name>`
- Create `.config/ti/TI.md` (currently missing)
- Document config precedence

**Gate criteria**:
- ✅ Load 3 CLI configs (claude, codex, ti) thành công
- ✅ Inheritance working: AGENTS.md + {CLI}.md + env
- ✅ `ti cli validate claude` verifies config

---

### Phase CLI-3: Router Integration per CLI
**Goal**: External CLIs point tới Router

**Tasks**:
- Implement `ti cli setup <name>` — configure env vars
- Ensure env vars persist across sessions
- Write verification test: `ti cli ping <name>` — 1 call via Router
- Document manual setup cho CLIs không auto-supported
- Port `mcp-router-bridge.ps1` → Go in `internal/mcp/routerbridge/`

**Gate criteria**:
- ✅ `ti cli setup claude` works end-to-end
- ✅ Claude Code call → Router → Anthropic → response back
- ✅ Cost tracking in Router shows Claude Code usage

---

### Phase CLI-4: Sub-Agent Runner (stdio)
**Goal**: `ti agent run <profile>` spawns external CLI

**Tasks**:
- Implement `internal/subagent/runner.go`
- Implement `internal/subagent/stdio_client.go`
- Agent profile loader (from `content/agents/*.md`)
- Basic profiles: code_reviewer, developer (hoặc reuse existing)
- Tool allowlisting enforcement
- `ti agent run` command

**Gate criteria**:
- ✅ `ti agent run code_reviewer --task "review X"` produces output
- ✅ Session persisted
- ✅ Tool restrictions enforced (reviewer không write được)

---

### Phase CLI-5: Parallel Execution
**Goal**: Multiple sub-agents concurrent

**Tasks**:
- Implement `internal/subagent/parallel.go`
- `ti agent parallel <list>` command
- Aggregators: merge, vote, race
- Live log aggregator
- Progress UI

**Gate criteria**:
- ✅ `ti agent parallel claude codex --task "..."` runs 2 CLIs concurrent
- ✅ Aggregation modes work
- ✅ Resource limits (max concurrent) enforced

---

### Phase CLI-6: ACP Protocol
**Goal**: Support ACP-compliant CLIs (junie, kilo, cursor...)

**Tasks**:
- Implement `internal/subagent/acp_client.go`
- ACP JSON-RPC handlers
- Fallback to stdio for non-ACP CLIs
- Test với junie CLI
- Document ACP capabilities

**Gate criteria**:
- ✅ `ti agent run --provider junie <profile>` works via ACP
- ✅ Full ACP protocol round-trip

---

### Phase CLI-7: Advanced Features
**Goal**: Polish + Claude Code-native features

**Tasks**:
- Port Claude Code sub-agents pattern natively
- Enhance plan mode (cross-CLI support)
- Session sharing across CLIs
- Hooks system (pre/post task)
- Registry sync từ upstream (junie, awesome-lists)

**Gate criteria**:
- ✅ Native plan mode working
- ✅ Sessions shareable giữa claude ↔ codex
- ✅ Registry sync operational

---

### Phase CLI-8: Quality of Life
**Goal**: Daily usability

**Tasks**:
- Interactive REPL: `ti agent repl`
- Shell completion (bash, zsh, powershell)
- Rich progress UI (TUI optional)
- Metrics dashboard (usage per CLI)
- Export/import agent profiles

**Gate criteria**:
- ✅ Daily workflow smooth
- ✅ Metrics visible

---

## 12. Key Risks

### Risk 1: External CLI API changes
**Severity**: Medium
**Risk**: Claude Code / Codex updates break integration
**Mitigation**:
- Version pinning trong registry
- Integration tests per CLI
- Monitor upstream release notes

### Risk 2: Router không compat với all CLI formats
**Severity**: Medium
**Risk**: Một số CLI dùng non-standard API format
**Mitigation**:
- Router's existing 28 providers cover major formats
- MCP bridge cho special cases
- Document compatibility matrix

### Risk 3: Parallel execution resource exhaustion
**Severity**: Low
**Risk**: Nhiều sub-agents → RAM/CPU overload
**Mitigation**:
- Semaphore để limit concurrent
- Default max=3
- Configurable per profile

### Risk 4: Config complexity
**Severity**: Medium
**Risk**: Inheritance chain khó debug
**Mitigation**:
- `ti cli config <name> --show` shows effective config
- Clear error messages
- Documentation

### Risk 5: Process management bugs
**Severity**: Medium
**Risk**: Orphan processes, zombies, stuck IPC
**Mitigation**:
- Context timeout per agent
- Cleanup on shutdown signal
- Process monitoring

### Risk 6: ACP protocol version drift
**Severity**: Low
**Risk**: Junie/Zed ACP version breaks
**Mitigation**:
- Pin ACP protocol version
- Test matrix per ACP version
- Fallback to stdio mode

### Risk 7: Security — sub-agents execute arbitrary commands
**Severity**: High
**Risk**: Malicious prompt → agent runs `rm -rf`
**Mitigation**:
- Tool allowlisting (strict by default)
- Sandboxed working directory
- Plan mode confirmation
- Never `--yolo` by default

---

## 13. References

### Context Files
- `Ti-learning-lab/03_Knowledge/Router/cli-context.json` — CLI codebase details
- `Ti-learning-lab/03_Knowledge/Router/router-context.json` — Router details

### External References
- **CAO (CLI Agent Orchestrator)**: `Z:\02_CORE\_cli\.config\cli-agent-orchestrator\`
  - Python, AWS Labs OSS
  - Pattern reference cho multi-agent orchestration
- **Junie Registry**: `Z:\02_CORE\_cli\.config\junie-main\registry-nightly.json`
  - ~50 ACP-compliant CLIs
  - Distribution format reference
- **Claude Code docs**: https://docs.claude.com/en/docs/claude-code
- **ACP (Agent Client Protocol)**: https://agentclientprotocol.com

### Internal
- `apps/cli/` — Ti CLI codebase
- `apps/router/` — Router backend
- `Z:\02_CORE\_cli\.config\AGENTS.md` — Base agent guidelines
- `Z:\02_CORE\_cli\.config\CLI_AGENTS.md` — CLI catalog
- `Z:\02_CORE\_cli\.config\index.json` — Tool index
- `Z:\02_CORE\_cli\mcp-router-bridge.ps1` — Router MCP bridge (PowerShell)

### Related Plans
- `TI_CLAW_MASTER_PLAN.md` — Ti Claw agent framework (sister plan)
- `ROUTER_PLAN.md` — Router completion (sister plan)

---

**Status**: Planning v1 ✅
**Next**: Phase CLI-0 — Audit & Design
**Last Update**: 2026-05-05
