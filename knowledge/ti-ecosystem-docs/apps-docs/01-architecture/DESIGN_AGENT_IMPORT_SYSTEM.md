# Agent Import System Design

> **Date**: 2026-05-03
> **Purpose**: Design system to import agents from `content/agents/` into Ti CLI mode system

## Overview

The agent import system converts markdown-based agent definitions from `content/agents/` into Ti CLI modes. This enables users to easily import and use pre-defined agents as modes.

## Architecture

```
content/agents/
├── [agent].md              # Source agent definitions
    ↓
Agent Import Parser
    ↓ (parses MD + YAML frontmatter)
Mode Definition
    ↓ (converts to Mode struct)
Mode Storage (.ti/modes/)
    ↓ (saves as JSON)
Ti CLI Mode System
```

## Components

### 1. Agent Import Parser

**Location**: `apps/cli/internal/agentimport/parser.go`

**Responsibilities**:
- Parse markdown files with YAML frontmatter
- Extract agent metadata (name, description, priority, etc.)
- Extract agent body (identity, responsibilities, resources)
- Validate required fields
- Convert to Mode struct

**Input**:
```yaml
---
agent-type: tibrain-specialist
name: tibrain-specialist
description: Agent chuyên về TiBrain...
when-to-use: Khi task liên quan đến memory management...
model: gemini-2.5-pro
inherit-tools: true
inherit-mcps: true
color: cyan
priority: P0
---

# TIBRAIN SPECIALIST

> **Agent**: TiBrain Specialist
> **Focus**: Memory management...
> **Context**: Ti AI Agent Platform...
> **Workspace**: Z:\\10_WORKPLACE\\Ti
> **Trigger**: tibrain, handoff, memory...

## Identity
Bạn là chuyên gia về hệ thống memory...

## Core Responsibilities
1. STM
2. LTM
3. Handoff

## Resources
- Server: Z:\\10_WORKPLACE\\Ti\\router\\tibrain\\main.go
```

**Output**:
```go
Mode{
    ID: "tibrain-specialist",
    Name: "TiBrain Specialist",
    Description: "Agent chuyên về TiBrain...",
    Category: "memory",
    Priority: 0,
    SystemPrompt: "You are a specialist in TiBrain...",
    TriggerKeywords: []string{"tibrain", "handoff", "memory"},
    Model: "gemini-2.5-pro",
    InheritTools: true,
    InheritMCPs: true,
    Color: "cyan",
    Metadata: map[string]interface{}{
        "workspace": "Z:\\10_WORKPLACE\\Ti",
        "context": "Ti AI Agent Platform",
    },
}
```

### 2. Mode Mapper

**Location**: `apps/cli/internal/agentimport/mapper.go`

**Responsibilities**:
- Map agent-type to mode ID
- Map priority (P0/P1/P2) to numeric values
- Map color strings to display colors
- Extract trigger keywords
- Categorize agents by domain

**Mapping Rules**:
- `agent-type` → `mode.ID` (kebab-case)
- `priority: P0` → `priority: 0`
- `priority: P1` → `priority: 1`
- `priority: P2` → `priority: 2`
- `trigger` field → split by comma to `triggerKeywords`

### 3. Import Manager

**Location**: `apps/cli/internal/agentimport/manager.go`

**Responsibilities**:
- Scan `content/agents/` directory
- Filter by priority/category
- Batch import agents
- Handle conflicts (overwrite/skip)
- Validate imported modes
- Generate import report

**API**:
```go
type ImportManager struct {
    parser  *Parser
    mapper  *Mapper
    storage ModeStorage
    config  ImportConfig
}

type ImportConfig struct {
    SourceDir      string
    TargetDir      string
    PriorityFilter []int
    CategoryFilter []string
    Overwrite      bool
    DryRun         bool
}

type ImportReport struct {
    Total      int
    Imported   int
    Skipped    int
    Failed     int
    Errors     []ImportError
    Duration   time.Duration
}
```

### 4. CLI Command

**Location**: `apps/cli/cmd/agent.go`

**Command**: `ti agent import`

**Flags**:
```
--source DIR       Source directory (default: content/agents)
--target DIR       Target directory (default: .ti/modes)
--priority P0,P1   Filter by priority (default: all)
--category X,Y     Filter by category (default: all)
--overwrite        Overwrite existing modes
--dry-run          Preview changes without applying
--verbose          Show detailed output
```

**Examples**:
```bash
# Import all agents
ti agent import

# Import only P0-P1 agents
ti agent import --priority P0,P1

# Import only language-specific agents
ti agent import --category language

# Dry run to preview
ti agent import --dry-run --verbose

# Import from custom directory
ti agent import --source /path/to/agents
```

## Import Workflow

### Step 1: Scan Source Directory
```go
files, err := filepath.Glob(filepath.Join(sourceDir, "*.md"))
```

### Step 2: Parse Each File
```go
for _, file := range files {
    agent, err := parser.Parse(file)
    if err != nil {
        report.Errors = append(report.Errors, err)
        report.Failed++
        continue
    }
```

### Step 3: Apply Filters
```go
if !matchesPriority(agent.Priority, config.PriorityFilter) {
    report.Skipped++
    continue
}

if !matchesCategory(agent.Category, config.CategoryFilter) {
    report.Skipped++
    continue
}
```

### Step 4: Convert to Mode
```go
mode := mapper.AgentToMode(agent)
```

### Step 5: Validate Mode
```go
if err := validator.Validate(mode); err != nil {
    report.Errors = append(report.Errors, err)
    report.Failed++
    continue
}
```

### Step 6: Check for Conflicts
```go
if exists, err := storage.Exists(mode.ID); err == nil && exists {
    if !config.Overwrite {
        report.Skipped++
        continue
    }
}
```

### Step 7: Save Mode
```go
if !config.DryRun {
    if err := storage.Save(mode); err != nil {
        report.Errors = append(report.Errors, err)
        report.Failed++
        continue
    }
}

report.Imported++
```

### Step 8: Generate Report
```go
return &ImportReport{
    Total:    len(files),
    Imported: report.Imported,
    Skipped:  report.Skipped,
    Failed:   report.Failed,
    Errors:   report.Errors,
    Duration: time.Since(start),
}
```

## Validation Rules

### Required Fields
- `agent-type` (string)
- `name` (string)
- `description` (string)

### Optional Fields
- `when-to-use` (string)
- `model` (string)
- `inherit-tools` (bool, default: false)
- `inherit-mcps` (bool, default: false)
- `color` (string)
- `priority` (string, default: P2)

### Validation Checks
1. YAML frontmatter must be valid
2. Required fields must be present
3. Priority must be P0, P1, or P2
4. Color must be valid (gold, cyan, green, purple, orange, blue, teal)
5. Mode ID must be unique
6. System prompt must not be empty

## Error Handling

### Parse Errors
- Invalid YAML syntax
- Missing required fields
- Invalid field values

### Validation Errors
- Invalid priority
- Invalid color
- Duplicate mode ID
- Empty system prompt

### Storage Errors
- Permission denied
- Disk full
- Filesystem error

## Testing

### Unit Tests
- Test parser with valid/invalid input
- Test mapper with various agent types
- Test validator with edge cases

### Integration Tests
- Test full import workflow
- Test with dry-run mode
- Test with overwrite flag
- Test with filters

### Manual Tests
- Import all agents
- Import subset by priority
- Import subset by category
- Handle conflicts

## Performance Considerations

- Parallel parsing for large agent sets
- Caching of parsed agents
- Batch storage operations
- Progress reporting for long imports

## Security Considerations

- Validate file paths (prevent directory traversal)
- Sanitize mode IDs (prevent injection)
- Limit file size (prevent DoS)
- Validate YAML structure (prevent code execution)

## Future Enhancements

1. **Watch Mode**: Auto-import new agents when added
2. **Sync Mode**: Two-way sync between source and target
3. **Versioning**: Track agent versions
4. **Dependencies**: Import dependent agents automatically
5. **Templates**: Generate agent templates from existing modes
6. **Export**: Export modes back to agent format

## Documentation

- User guide: How to import agents
- Developer guide: How to extend parser
- Agent format guide: How to write agents
- Migration guide: From old to new format
