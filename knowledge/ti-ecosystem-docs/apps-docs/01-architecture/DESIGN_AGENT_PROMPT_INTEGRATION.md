# Agent and Prompt Integration System

## Overview

The Ti CLI now supports importing agents and prompts from markdown files in the `content/` directory and converting them to modes. This system allows for easy management and distribution of specialized AI agent behaviors.

## Architecture

### Components

1. **Agent Import System** (`internal/agentimport/`)
   - `parser.go` - Parses YAML frontmatter from agent markdown files
   - `mapper.go` - Converts agent definitions to mode structures
   - `manager.go` - Manages the import process with database integration

2. **Prompt Import System** (`internal/promptimport/`)
   - `parser.go` - Parses plain markdown files (no YAML frontmatter)
   - `mapper.go` - Converts prompt definitions to mode structures
   - `manager.go` - Manages the import process with database integration

3. **Mode Storage** (`internal/modes/`)
   - `schema.go` - Database schema and CRUD operations
   - `manager.go` - High-level mode management
   - `types.go` - Mode data structures

## Agent Import

### Agent File Format

Agent files use YAML frontmatter to define agent metadata:

```markdown
---
name: Architect
description: Software architecture specialist
category: architecture
priority: 0
model: claude-3-5-sonnet
tools: [read, write, exec]
permissions:
  mode: normal
  allow_shell: true
  allow_network: false
  allow_file_read: true
  allow_file_write: true
  allow_system: false
---

# Architect

You are a software architecture specialist...
```

### Import Command

```bash
# Import agents from content/agents
ti agent import content/agents

# Import with specific options
ti agent import content/agents \
  --category architecture \
  --priority 0 \
  --overwrite \
  --verbose

# List importable agents
ti agent list-importable
```

### Agent Categories

Agents are categorized by priority:
- **P0** (High): Ti-specific core agents (tibrain, infrastructure)
- **P1** (Medium): Domain-specific agents (architecture, security, performance)
- **P2** (Normal): General-purpose agents
- **P3** (Low): Legacy or deprecated agents

### Agent Categories

- **ti**: Ti-specific agents (tibrain-specialist, router)
- **architecture**: Architecture and design (architect, system-designer)
- **security**: Security specialists (security-auditor, pentester)
- **performance**: Performance optimization (perf-engineer, db-tuner)
- **quality**: Quality assurance (qa-engineer, tester)
- **devops**: DevOps and infrastructure (devops-engineer, sre)
- **generic**: Generic purpose agents (coder, debugger, reviewer)

## Prompt Import

### Prompt File Format

Prompt files use plain markdown (no YAML frontmatter):

```markdown
# MCP Expert

You are an MCP (Model Context Protocol) integration specialist...

## Features
- MCP server development
- MCP client integration
- MCP tool implementation
```

### Import Command

```bash
# Import prompts from content/prompts
ti mode import-prompts content/prompts

# Import with options
ti mode import-prompts content/prompts \
  --use-mapping-table \
  --dry-run \
  --verbose
```

### Prompt Parsing

The parser extracts:
- **Title**: First heading (#) in the file
- **Mode ID**: Filename converted to kebab-case
- **Category**: Inferred from filename prefix
- **Priority**: Inferred from filename prefix

### Filename Conventions

- `mcp-*.md` → Category: integration, Priority: P1
- `tibrain-*.md` → Category: ti, Priority: P0
- `structure-*.md` → Category: quality, Priority: P1
- `router-*.md` → Category: infrastructure, Priority: P1
- `generic-*.md` → Category: generic, Priority: P3
- `enhanced-*.md` → Category: enhanced, Priority: P2

### Mapping Table

A predefined mapping table can be used instead of inference:

```go
{
  "mcp-expert.md": {
    "mode_id": "mcp-expert",
    "category": "integration",
    "priority": "1"
  },
  "tibrain-specialist.md": {
    "mode_id": "tibrain-specialist",
    "category": "ti",
    "priority": "0"
  }
}
```

## Mode Management

### List Modes

```bash
ti mode list
```

### Show Mode Details

```bash
ti mode show architect
```

### Switch to Mode

```bash
ti mode architect
```

### Create Custom Mode

```bash
ti mode custom create my-mode \
  --prompt "You are a specialist..." \
  --model claude-3-5-sonnet
```

### Edit Custom Mode

```bash
ti mode custom edit my-mode \
  --prompt "Updated prompt..."
```

### Delete Custom Mode

```bash
ti mode custom delete my-mode
```

## Database Storage

Modes are stored in SQLite database (`~/.ti/data/modes.db`):

```sql
CREATE TABLE modes (
  id TEXT PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  description TEXT,
  prompt TEXT,
  tools TEXT,
  model TEXT,
  permissions TEXT,
  context_strategy TEXT,
  skills TEXT,
  metadata TEXT,
  created_at TEXT,
  updated_at TEXT,
  is_built_in INTEGER
);
```

## API Integration

### ModeDB Interface

```go
type ModeDB struct {
    db          *sql.DB
    modeCache   map[string]*Mode
    modeCacheMu sync.RWMutex
    cacheExpiry time.Duration
}

// CRUD Operations
func (m *ModeDB) SaveMode(mode *Mode) error
func (m *ModeDB) GetMode(name string) (*Mode, error)
func (m *ModeDB) ListModes() ([]*Mode, error)
func (m *ModeDB) DeleteMode(name string) error
```

### Manager Interface

```go
type Manager struct {
    db      *ModeDB
    dataDir string
}

// High-level operations
func (m *Manager) CreateMode(mode *Mode) error
func (m *Manager) UpdateMode(mode *Mode) error
func (m *Manager) GetMode(name string) (*Mode, error)
func (m *Manager) ListModes() ([]*Mode, error)
func (m *Manager) SetActiveMode(name string, config interface{}) error
func (m *Manager) GetActiveMode() (*Mode, error)
```

## Testing

### Unit Tests

```bash
# Test agent import
go test ./internal/agentimport/...

# Test prompt import
go test ./internal/promptimport/...

# Test modes
go test ./internal/modes/...
```

### Integration Tests

```bash
# Test end-to-end import
ti agent import content/agents --dry-run --verbose
ti mode import-prompts content/prompts --dry-run --verbose
```

## Best Practices

1. **Use YAML Frontmatter for Agents**: Define all metadata in the frontmatter for consistency
2. **Use Plain Markdown for Prompts**: Keep prompts simple without frontmatter
3. **Follow Filename Conventions**: Use descriptive prefixes for automatic categorization
4. **Test Before Import**: Always use `--dry-run` to preview changes
5. **Version Control**: Keep agent/prompt files in git for version tracking
6. **Documentation**: Include usage examples in agent/prompt descriptions

## Troubleshooting

### Import Fails

- Check file format (YAML frontmatter for agents, plain markdown for prompts)
- Verify filename conventions
- Use `--verbose` flag for detailed error messages
- Use `--dry-run` to preview without making changes

### Mode Not Found

- ModeDB uses `name` (not `id`) for lookups
- Check if mode name matches exactly (case-sensitive)
- Use `ti mode list` to see all available modes

### Database Issues

- Database location: `~/.ti/data/modes.db`
- Backup before bulk imports
- Use SQLite tools to inspect database if needed

## Future Enhancements

- [ ] Add mode versioning and migration support
- [ ] Support for prompt templates with variables
- [ ] Mode dependencies and composition
- [ ] Mode sharing and discovery
- [ ] Mode performance metrics
- [ ] Mode A/B testing
