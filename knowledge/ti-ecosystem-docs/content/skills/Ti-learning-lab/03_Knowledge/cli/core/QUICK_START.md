# Ti CLI Quick Start Guide

> **Last Updated**: 2026-05-05
> **Version**: 3.0.0-ti-ecosystem-go1.23

## 🚀 Getting Started

### Installation

```bash
# Build from source
cd apps/cli
go build -o ../../bin/ti .

# Add to PATH (optional)
export PATH=$PATH:/z/10_WORKPLACE/Ti/bin
```

### First Steps

```bash
# Check version
ti --version

# Get help
ti --help

# System health check
ti auto doctor
```

## 🎯 Common Use Cases

### 1. Code Development

```bash
# Fix code issues
ti fix

# Plan implementation
ti plan

# Code review
ti review

# Optimize code
ti optimize
```

### 2. Knowledge Management

```bash
# Sync knowledge to Notion (dry-run)
ti auto knowledge-sync --dry-run

# Sync with AI analysis
ti auto knowledge-sync --ai

# Check sync status
ti auto knowledge-status

# Analyze single file
ti auto knowledge-analyze path/to/file.md
```

### 3. Automation

```bash
# System health check
ti auto doctor

# Create backup
ti auto snapshot

# Safe fix workflow
ti auto fix-safe

# PR review
ti auto pr-review
```

### 4. Agent Management

```bash
# List agents
ti agent list

# Run sub-agent
ti sub-agent devin "fix the bug"

# Autonomous execution
ti autonomous
```

### 5. Sub-Agent Orchestration

Ti CLI can orchestrate external CLI agents (Claude Code, Codex, Gemini, etc.) as sub-agents with powerful parallel execution capabilities. See [SUB_AGENT_GUIDE.md](SUB_AGENT_GUIDE.md) for detailed documentation.

#### Run Single Sub-Agent

```bash
# Run code review via Claude
ti sub-agent run code_reviewer --provider claude --task "review src/auth.go"

# Run refactoring via Codex
ti sub-agent run go-reviewer --provider codex --task "refactor this package" --timeout 120

# With additional context
ti sub-agent run security_reviewer --provider claude \
  --task "check for security issues" \
  --context "This is a payment processing module"
```

#### Parallel Sub-Agent Execution

```bash
# Run review via multiple providers (merge mode)
ti sub-agent parallel code_reviewer \
  --providers claude,codex \
  --task "review src/" \
  --mode merge

# Vote mode (majority consensus)
ti sub-agent parallel code_reviewer \
  --providers claude,codex,gemini \
  --task "..." \
  --mode vote

# Race mode (fastest successful response)
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

#### Agent Profiles

```bash
# List available agent profiles
ti sub-agent list-profiles

# Agent profiles are defined in content/agents/{name}.md
# Example profile structure:
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

#### Multi-Agent Workflows (Ti Claw Plugin)

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

### 6. Skills & Learning

```bash
# List skills
ti skill list

# Search skills
ti skill search "testing"

# Run skill
ti skill run skill-name

# Validate skill
ti skill validate skill-name
```

### 7. MCP Integration

```bash
# List MCP servers
ti mcp list

# Add MCP server
ti mcp add server-name stdio /path/to/server

# Test MCP server
ti mcp test server-name

# Start MCP proxy
ti mcp serve
```

### 8. Context Management

```bash
# Scan project context
ti context scan ./project

# Pack context
ti context pack ./project

# Check context budget
ti context budget
```

### 9. Provider Management

```bash
# List providers
ti provider list

# Query provider
ti provider query openai

# Setup provider
ti auto provider-setup --type openai
```

## 🔧 Configuration

### Basic Config

```bash
# Config location
~/.ti-cli/config.yaml

# Show current config
ti config show

# Validate config
ti config validate
```

### Provider Config

```bash
# Provider config location
configs/providers.yaml

# Sync config with Router
ti auto router sync
```

### Environment Variables

```bash
# Notion API key
export NOTION_API_KEY=your_key

# LLM provider
export LLM_PROVIDER=openai
export LLM_API_KEY=your_key

# BD tool data directory
export TI_DATA_DIR=Z:\03_DATA\ti
```

## 🧪 Testing

### Run Tests

```bash
# Unit tests
cd apps/cli
go test ./...

# E2E tests
cd apps/cli/e2e
go test -v

# Specific test
go test -v -run TestCLIBasicFunctionality
```

### Test Error Handling

```bash
# Test error improvements
ti error-test all

# Test specific error type
ti error-test config
ti error-test provider
```

## 🐛 Troubleshooting

### Common Issues

#### **Router Connection Failed**
```bash
# Check Router status
ti auto router status

# Start Router
cd apps/router
go run ./cmd/routerd --config configs/providers.yaml
```

#### **Plugin Issues**
```bash
# Check plugin compatibility
ti plugin-compat status

# List plugins
ti plugin-compat list

# Migrate plugin
ti plugin-compat migrate plugin-name
```

#### **Config Issues**
```bash
# Validate config
ti doctor

# Show config
ti config show

# Reset config (backup first!)
rm ~/.ti-cli/config.yaml
```

#### **Permission Issues**
```bash
# Check file permissions
ti permission check

# Fix permissions
chmod +x script.sh
```

## 📚 Advanced Usage

### Natural Language Commands

```bash
# The CLI understands natural language
ti "fix the authentication bug in auth.go"

# Plan with context
ti "plan a refactoring of the user module"

# Review with specific criteria
ti "review the PR for security issues"
```

### Multi-Agent Workflows

```bash
# Run multiple agents in parallel
ti autonomous --parallel

# Orchestrate specific agents
ti sub-agent devin "fix bug" &
ti sub-agent codex "add tests" &
wait
```

### Context Budgeting

```bash
# Set token budget
ti context pack ./project --budget 10000

# Check budget usage
ti context budget
```

### Plugin Development

```bash
# Create plugin in .ti/plugins/
mkdir -p .ti/plugins/my-plugin

# Plugin will be auto-discovered
# See plugin documentation for structure
```

## 🔌 Integration Examples

### Notion Integration

```bash
# Setup Notion database (see NOTION_DATABASE_CONFIGURATION.md)

# Sync knowledge
ti auto knowledge-sync --ai

# Check status
ti auto knowledge-status
```

### GitHub Integration

```bash
# PR review
ti auto pr-review --pr 123

# Repo analysis
ti repo analyze owner/repo
```

### Router Integration

```bash
# Check Router health
ti auto router status

# Get Router metrics
ti auto router metrics

# Sync config
ti auto router sync
```

## 🎓 Learning Resources

### Documentation
- **Architecture**: `ARCHITECTURE.md`
- **CLI Guide**: `CLI_GUIDE.md`
- **AGENTS.md**: Project-specific context

### Skills
```bash
# List available skills
ti skill list

# Search for skills
ti skill search "topic"

# Learn skill patterns
ti skill learn
```

### Examples
```bash
# Run example workflows
ti auto fix-safe --dry-run

# Test with sample project
ti context scan ./examples/sample-project
```

## 🚀 Performance Tips

### 1. Use Context Budgeting
```bash
# Limit context size for faster operations
ti context pack ./project --budget 5000
```

### 2. Enable Caching (when available)
```bash
# Context caching reduces I/O
ti context pack ./project --cache
```

### 3. Use Dry-Run First
```bash
# Preview changes before execution
ti auto knowledge-sync --dry-run
ti auto fix-safe --dry-run
```

### 4. Parallel Operations
```bash
# Run agents in parallel when possible
ti autonomous --parallel
```

## 🔒 Security Best Practices

### 1. Secrets Management
```bash
# Use centralized secrets
# Z:\00_SECRET\

# Never hardcode secrets
# Use environment variables
export API_KEY=xxx
```

### 2. Permission Control
```bash
# Use permission sandbox
ti permission check

# Review policies
ti permission list
```

### 3. Validation
```bash
# Always validate before risky operations
ti doctor

# Use dry-run for preview
ti auto snapshot --dry-run
```

## 📈 Monitoring

### Health Checks
```bash
# System health
ti auto doctor

# Router health
ti auto router status

# Plugin health
ti plugin-compat status
```

### Metrics
```bash
# Router metrics
ti auto router metrics

# Context stats
ti context stats
```

### Logging
```bash
# Enable debug logging
ti --log-level debug <command>

# Check logs
# Logs are in standard output/error
```

## 🆘 Getting Help

### Built-in Help
```bash
# General help
ti --help

# Command help
ti <command> --help

# Subcommand help
ti <command> <subcommand> --help
```

### Documentation
```bash
# CLI guide
ti guide

# Architecture docs
# See ARCHITECTURE.md
```

### Community
- GitHub Issues: Report bugs and feature requests
- Documentation: Check `Ti-learning-lab/03_Knowledge/`
- Skills: `ti skill list` for available skills

## 🎯 Next Steps

1. **Explore**: Try basic commands to get familiar
2. **Configure**: Set up your providers and config
3. **Integrate**: Connect to Router and Notion
4. **Automate**: Set up workflows for your use cases
5. **Customize**: Develop skills and plugins

## 💡 Tips

- Use `--dry-run` to preview changes
- Use `--help` to explore commands
- Use `ti doctor` for troubleshooting
- Use natural language for complex requests
- Check `ARCHITECTURE.md` for deep understanding

## 🔄 Updates

### Stay Updated
```bash
# Check version
ti --version

# Pull latest changes
git pull

# Rebuild
cd apps/cli && go build -o ../../bin/ti .
```

### Migration Notes
- Check `CLI_GUIDE.md` for migration from legacy CLI
- Use `ti plugin-compat` to manage plugin migration
- Backup config before updates
