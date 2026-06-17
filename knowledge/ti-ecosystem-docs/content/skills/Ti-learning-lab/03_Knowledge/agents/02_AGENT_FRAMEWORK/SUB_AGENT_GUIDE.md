# Ti CLI Sub-Agent System Guide

> **Last Updated**: 2026-05-05
> **Version**: 3.0.0-ti-ecosystem-go1.23

## 🎯 Overview

Ti CLI's Sub-Agent System enables orchestration of external CLI agents (Claude Code, Codex, Gemini, Qwen, etc.) as sub-agents with powerful parallel execution capabilities. This is a **unique feature** not available in any other CLI.

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Ti CLI                                  │
│                  (Orchestrator)                             │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
        ┌─────▼─────┐  ┌────▼────┐  ┌─────▼──────┐
        │  Runner   │  │Parallel │  │  Profile   │
        │  (Sync)   │  │ Runner  │  │  System    │
        └─────┬─────┘  └────┬────┘  └─────┬──────┘
              │            │              │
        ┌─────▼─────┐  ┌────▼────┐  ┌─────▼──────┐
        │   CLI     │  │  CLI    │  │  YAML      │
        │ Registry  │  │ Config  │  │  Front-    │
        │           │  │         │  │  matter    │
        └───────────┘  └─────────┘  └────────────┘
```

## 📋 Components

### 1. Runner (`internal/subagent/runner.go`)

Executes external CLI as a sub-agent synchronously (Handoff pattern).

**Features:**
- Spawns external CLI via `exec.CommandContext`
- Passes task as stdin with system prompt
- Captures stdout+stderr
- Supports different CLI flag conventions
- Timeout handling
- CLI Registry integration

**Supported CLIs:**
- Claude Code: `--print` flag
- Codex: `--quiet` flag
- Gemini: `-p` flag
- Generic: positional argument

### 2. Parallel Runner (`internal/subagent/parallel.go`)

Runs multiple sub-agents concurrently with aggregation modes.

**Aggregation Modes:**
- **merge**: Combine all outputs (default)
- **vote**: Majority consensus
- **race**: Return fastest successful response

**Features:**
- Concurrent execution with semaphore limiting
- Live logging callback
- Result aggregation
- Timeout handling
- Concurrency control

### 3. Profile System (`internal/subagent/profile.go`)

Agent profiles defined in Markdown files with YAML front-matter.

**Profile Structure:**
```yaml
---
name: code_reviewer
role: Code review specialist
provider: claude
model: claude-sonnet-4-5
allowed_tools:
  - file_read
  - git_operations
restricted_tools:
  - file_write
max_tokens: 200000
router: true
---

System prompt content here...
```

**Profile Fields:**
- `name`: Agent identifier
- `role`: Agent role description
- `provider`: Default CLI provider
- `model`: Default model
- `allowed_tools`: Permitted tools
- `restricted_tools`: Blocked tools
- `max_tokens`: Token limit
- `router`: Enable Router integration
- `System prompt`: Markdown body below front-matter

**Tool Permission Logic:**
- If `allowed_tools` is empty → all tools permitted
- If `allowed_tools` has values → only those tools permitted
- `restricted_tools` always blocks regardless of `allowed_tools`

### 4. CLI Registry Integration (`internal/cliregistry/`)

Dynamic CLI discovery and configuration.

**Features:**
- CLI registration and discovery
- Executable path resolution
- Environment variable management
- Configuration merging

## 🚀 Usage

### Single Sub-Agent

```bash
# Basic usage
ti sub-agent run code_reviewer --provider claude --task "review src/auth.go"

# With timeout
ti sub-agent run go-reviewer --provider codex \
  --task "refactor this package" \
  --timeout 120

# With additional context
ti sub-agent run security_reviewer --provider claude \
  --task "check for security issues" \
  --context "This is a payment processing module"
```

### Parallel Sub-Agent

```bash
# Merge mode (default)
ti sub-agent parallel code_reviewer \
  --providers claude,codex \
  --task "review src/" \
  --mode merge

# Vote mode
ti sub-agent parallel code_reviewer \
  --providers claude,codex,gemini \
  --task "..." \
  --mode vote

# Race mode
ti sub-agent parallel code_reviewer \
  --providers claude,codex \
  --task "quick fix" \
  --mode race

# Control concurrency
ti sub-agent parallel code_reviewer \
  --providers claude,codex,gemini,qwen \
  --task "..." \
  --max 5 \
  --timeout 600
```

### List Profiles

```bash
# List all available profiles
ti sub-agent list-profiles
```

## 📊 Aggregation Modes

### Merge Mode

Combines all outputs with agent labels.

**Output Format:**
```markdown
## [claude] (2500ms)
Review output from Claude...

## [codex] (1800ms)
Review output from Codex...

## [gemini] (3200ms)
Review output from Gemini...
```

**Use Cases:**
- Getting multiple perspectives
- Comprehensive analysis
- Comparison of outputs

### Vote Mode

Majority consensus based on first-line answers.

**Output Format:**
```markdown
Vote result (2/3): Yes, the code is secure
```

**Use Cases:**
- Binary decisions (yes/no)
- Consensus building
- Reducing bias

### Race Mode

Returns fastest successful response.

**Output Format:**
```markdown
## [codex] (1200ms)
Fastest successful response...
```

**Use Cases:**
- Time-critical tasks
- Quick prototyping
- Latency optimization

## 🎨 Profile Examples

### Code Reviewer Profile

```yaml
---
name: code_reviewer
role: Code review specialist
provider: claude
model: claude-sonnet-4-5
allowed_tools:
  - file_read
  - git_operations
  - git_diff
restricted_tools:
  - file_write
  - git_commit
max_tokens: 200000
router: true
---

You are a code review specialist. Focus on:
- Security vulnerabilities
- Performance issues
- Code quality
- Best practices
- Edge cases

Provide constructive feedback with specific suggestions.
```

### Security Reviewer Profile

```yaml
---
name: security_reviewer
role: Security audit specialist
provider: claude
model: claude-opus-4-5
allowed_tools:
  - file_read
  - git_operations
restricted_tools:
  - file_write
  - network_operations
max_tokens: 500000
router: true
---

You are a security audit specialist. Focus on:
- OWASP Top 10
- Authentication/authorization
- Data encryption
- Input validation
- Dependency vulnerabilities

Flag critical issues with severity levels.
```

### Performance Reviewer Profile

```yaml
---
name: performance_reviewer
role: Performance optimization specialist
provider: codex
model: gpt-4o
allowed_tools:
  - file_read
  - git_operations
restricted_tools:
  - file_write
max_tokens: 100000
router: false
---

You are a performance optimization specialist. Focus on:
- Algorithm complexity
- Memory usage
- I/O operations
- Caching strategies
- Concurrency patterns

Provide measurable improvements with benchmarks.
```

## 🔧 Configuration

### Profile Location

Default profile directory: `content/agents/`

Can be overridden via environment variable:
```bash
export TI_AGENTS_DIR=/path/to/agents
```

### CLI Host Configuration

CLI host configuration in `~/.ti-cli/cli-host.yaml`:

```yaml
cli-host:
  agents:
    claude:
      name: claude
      executable: /path/to/claude
      env_vars:
        ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY}
    codex:
      name: codex
      executable: /path/to/codex
      env_vars:
        OPENAI_API_KEY: ${OPENAI_API_KEY}
```

## 🌐 Multi-Agent Orchestrator (Ti Claw Plugin)

For advanced multi-agent workflows, use the Ti Claw multi-agent orchestrator:

```bash
# Create agent instance
ti ticlaw multi-agent create --type notion --id notion_agent_1

# Start agent
ti ticlaw multi-agent start --id notion_agent_1

# Show system status
ti ticlaw multi-agent status

# Create workflow
ti ticlaw multi-agent workflow

# Stop agent
ti ticlaw multi-agent stop --id notion_agent_1
```

**Built-in Agent Types:**
- **notion**: Task management agent
- **developer**: Software development agent
- **analyst**: Data analysis agent

**Orchestrator Components:**
- Agent Registry: Agent type management
- Message Bus: Pub/sub inter-agent communication
- Task Queue: Priority-based task distribution
- Agent Instances: Running agent lifecycle

## 📈 Performance

### Execution Time

- Single sub-agent: ~2-10s (depends on task)
- Parallel (2 agents): ~2-10s (concurrent)
- Parallel (5 agents): ~2-10s (concurrent, limited by slowest)

### Memory Usage

- Base overhead: ~50MB
- Per sub-agent: ~20-30MB
- Parallel (5 agents): ~150-200MB

### Concurrency

- Default max concurrent: 3
- Configurable via `--max` flag
- Recommended: 3-5 for typical workloads

## 🎯 Use Cases

### 1. Code Review Parallel

Run code review via multiple providers for comprehensive analysis:

```bash
ti sub-agent parallel code_reviewer \
  --providers claude,codex,gemini \
  --task "review src/payment/" \
  --mode merge
```

### 2. Multi-Provider Testing

Test code with multiple providers to ensure compatibility:

```bash
ti sub-agent parallel tester \
  --providers claude,codex,gemini,qwen \
  --task "run tests" \
  --mode vote
```

### 3. Specialized Agents

Create specialized agents for different tasks:

```bash
# Security review
ti sub-agent run security_reviewer --provider claude --task "audit auth/"

# Performance review
ti sub-agent run performance_reviewer --provider codex --task "optimize db/"

# UX review
ti sub-agent run ux_reviewer --provider gemini --task "review ui/"
```

### 4. Cost Optimization

Use race mode to get fastest response from cheapest provider:

```bash
ti sub-agent parallel quick_fix \
  --providers groq,cerebras,deepseek \
  --task "fix typo" \
  --mode race
```

### 5. Agent Collaboration

Use multi-agent orchestrator for complex workflows:

```bash
# Create specialized agents
ti ticlaw multi-agent create --type notion --id notion_1
ti ticlaw multi-agent create --type developer --id dev_1
ti ticlaw multi-agent create --type analyst --id analyst_1

# Start agents
ti ticlaw multi-agent start --id notion_1
ti ticlaw multi-agent start --id dev_1
ti ticlaw multi-agent start --id analyst_1

# Monitor
ti ticlaw multi-agent status
```

## 🔍 Comparison with Other CLIs

| Feature | Ti CLI | Claude Code | Cursor | Windsurf | Qoder |
|---------|--------|-------------|--------|----------|-------|
| **Sub-agent orchestration** | ✅ | ❌ | ❌ | ❌ | ✅ (custom) |
| **Parallel execution** | ✅ (3 modes) | ❌ | ❌ | ❌ | ❌ |
| **Profile system** | ✅ (YAML) | ❌ | ❌ | ❌ | ❌ |
| **Aggregation modes** | ✅ (merge/vote/race) | ❌ | ❌ | ❌ | ❌ |
| **CLI Registry** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Multi-agent orchestrator** | ✅ | ❌ | ❌ | ❌ | ❌ |

## 🛠️ Troubleshooting

### CLI Not Found

```
Error: executable not found for claude (run `ti cli install claude`)
```

**Solution:**
```bash
# Install CLI via CLI Registry
ti cli install claude

# Or set executable path in cli-host.yaml
```

### Profile Not Found

```
Error: subagent: load profile "code_reviewer": no such file
```

**Solution:**
```bash
# List available profiles
ti sub-agent list-profiles

# Create profile in content/agents/code_reviewer.md
```

### Timeout

```
Error: context deadline exceeded
```

**Solution:**
```bash
# Increase timeout
ti sub-agent run profile --provider claude \
  --task "..." \
  --timeout 600
```

### Parallel Execution Fails

```
Error: all agents failed in race
```

**Solution:**
```bash
# Use merge mode to see all errors
ti sub-agent parallel profile \
  --providers claude,codex \
  --task "..." \
  --mode merge
```

## 🎯 Quality Enhancement for Sub-Agent Execution

### Overview

Sub-agent execution thường gặp vấn đề: hỏi quá nhiều câu hỏi không cần thiết, thiếu context awareness, over-clarify. Quality Enhancement workflow giải quyết这些问题 bằng context-aware decision making.

### Core Principles

1. **Assume First, Ask Second** - Giả định hợp lý, chỉ hỏi khi critical
2. **Context-Driven Decisions** - Dùng project context để inform choices
3. **Progressive Disclosure** - Reveal complexity chỉ khi needed
4. **Pattern Recognition** - Học từ existing codebase conventions
5. **Confidence Scoring** - Chỉ hỏi khi confidence < threshold

### Features

#### 1. Context Collection

Tự động thu thập project context trước khi execution:
- Project structure scan
- Code patterns (imports, naming conventions, code style)
- Project rules (AGENTS.md, .editorconfig, linting config)
- Git history (last 10 commits)
- Dependencies analysis

#### 2. Decision Heuristics Engine

Confidence-based decision matrix:
- **Confidence ≥ 90%** → Assume without asking
- **Confidence 70-90% + non-critical** → Assume + document
- **Critical + Confidence < 90%** → Ask user
- **Non-critical + Confidence < 70%** → Ask user

**Decision Categories:**
| Category | Assume Threshold | Ask Threshold |
|----------|-----------------|---------------|
| Style | 80% | 60% |
| Architecture | 70% | 50% |
| Security | 90% | 70% |
| Performance | 75% | 55% |
| Testing | 70% | 50% |

#### 3. Progressive Disclosure

- Execute với assumptions
- Document assumptions made
- Provide rollback nếu assumption sai
- Update pattern database (confidence adjustment)

### Implementation

#### Context Collector

Location: `internal/subagent/context/collector.go`

```go
type Context struct {
    Structure   *ProjectStructure
    Patterns    *CodePatterns
    Rules       *ProjectRules
    History     *GitHistory
    Dependencies *Dependencies
}

func (c *Collector) Collect() (*Context, error) {
    // Collect structure, patterns, rules, history, dependencies
}
```

#### Decision Engine

Location: `internal/subagent/decision/engine.go`

```go
type DecisionRule struct {
    Condition  func(*Context) bool
    Confidence int
    Assumption string
    Critical   bool
    Category   string
}

func (e *Engine) MakeDecision(ctx *Context, decision string) (string, bool, *Assumption) {
    // Apply decision rules, calculate confidence, return assumption or ask
}
```

### Profile Enhancement

Thêm quality directives vào profile YAML:

```yaml
---
name: code_reviewer
role: Code review specialist

# Quality directives
quality:
  min_confidence: 70
  critical_categories:
    - security
    - data_integrity
  assume_first: true
  document_assumptions: true
  allow_rollback: true

# Context requirements
context:
  collect_project_structure: true
  extract_patterns: true
  load_project_rules: true
  check_git_history: true
  analyze_dependencies: true
---
```

### Example Usage

❌ **BAD:**
```
"Should I use logrus or the standard log package?"
```

✅ **GOOD:**
```
Detected logrus in 15/20 files. Using logrus for logging.
(Assumption: 95% confidence based on import pattern)
```

### Integration with Workflows

Quality features đã được tích hợp vào `reflective-loop-beads.yaml v2.0`:
- Context Collection step
- Decision Heuristics trong Thought step
- Confidence-based HITL
- Progressive Disclosure với assumption validation

### Metrics

Track để validate improvement:
- **Question Rate**: Target 50% reduction
- **Assumption Accuracy**: Target ≥85%
- **Task Completion Time**: Target 30% faster
- **User Satisfaction**: Target ≥4/5 stars

## 📚 References

- **Source Code**: `apps/cli/internal/subagent/`
- **Profiles**: `content/agents/`
- **CLI Registry**: `apps/cli/internal/cliregistry/`
- **Multi-Agent**: `apps/cli/internal/plugins/ticlaw/multi/`

## 🚀 Future Enhancements

- [ ] Web-based agent dashboard
- [ ] Agent marketplace
- [ ] Custom agent builder UI
- [ ] Agent performance analytics
- [ ] Cost tracking per agent
- [ ] Agent collaboration templates
