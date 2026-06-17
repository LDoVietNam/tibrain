# Prompt Integration System Design

> **Date**: 2026-05-03
> **Purpose**: Design system to import prompts from content/prompts into Ti CLI modes

## Overview

Prompts in `content/prompts/` use plain markdown format without YAML frontmatter, unlike agents. This requires a different integration approach.

## Key Differences: Agents vs Prompts

| Aspect | Agents | Prompts |
|--------|--------|---------|
| Format | YAML frontmatter + markdown | Plain markdown |
| Metadata | Structured (name, description, priority) | Inferred from filename/content |
| Parsing | YAML parser | Markdown parser |
| Mapping | Agent-type → Mode ID | Filename → Mode ID |
| Validation | Required fields | Content validation |

## Architecture

```
content/prompts/
├── [role].md
├── enhanced-[role].md
└── generic/
    └── generic-[role].md
    ↓
Prompt Parser (plain markdown)
    ↓ (extract title, content)
Prompt Definition
    ↓ (map filename to mode ID)
Mode Storage (.ti/modes/)
    ↓ (save as mode with prompt)
Ti CLI Mode System
```

## Components

### 1. Prompt Parser

**Location**: `apps/cli/internal/promptimport/parser.go`

**Responsibilities**:
- Parse plain markdown files
- Extract title from first heading (# Title)
- Extract entire content as prompt
- Validate content is not empty
- Infer metadata from filename

**Input**:
```markdown
# Developer Prompt

You are a software engineer...

## Core Responsibilities
1. ...
2. ...
```

**Output**:
```go
type Prompt struct {
    Filename    string
    Title       string
    Content     string
    ModeID      string  // Inferred from filename
    Category    string  // Inferred from filename
    Priority    int     // Inferred from filename
}
```

### 2. Prompt Mapper

**Location**: `apps/cli/internal/promptimport/mapper.go`

**Responsibilities**:
- Map filename to mode ID (kebab-case conversion)
- Infer category from filename patterns
- Infer priority from filename patterns
- Convert Prompt to Mode

**Mapping Rules**:
- `mcp-expert.md` → `mode: mcp-expert`, category: integration, priority: 1
- `tibrain-specialist.md` → `mode: tibrain-specialist`, category: ti, priority: 0
- `structure-guardian.md` → `mode: structure-guardian`, category: quality, priority: 1
- `router-prompts.md` → `mode: router`, category: infrastructure, priority: 1
- `generic-*.md` → `mode: generic-*`, category: generic, priority: 3

### 3. Prompt Import Manager

**Location**: `apps/cli/internal/promptimport/manager.go`

**Responsibilities**:
- Scan prompt directory
- Filter by category
- Batch import prompts
- Handle conflicts
- Generate import report

**API**:
```go
type PromptImportManager struct {
    parser  *PromptParser
    mapper  *PromptMapper
    db      *modes.ModeDB
    config  PromptImportConfig
}

type PromptImportConfig struct {
    SourceDir      string
    TargetDir      string
    CategoryFilter []string
    Overwrite      bool
    DryRun         bool
    Verbose        bool
}

type PromptImportReport struct {
    Total      int
    Imported   int
    Skipped   int
    Failed     int
    Errors     []PromptImportError
    Duration   time.Duration
}
```

### 4. CLI Command

**Location**: `apps/cli/cmd/mode.go`

**Command**: `ti mode import-prompts`

**Flags**:
```
--source DIR       Source directory (default: content/prompts)
--target DIR       Target directory (default: .ti/modes)
--category X,Y     Filter by category (comma-separated)
--overwrite        Overwrite existing modes
--dry-run          Preview changes without applying
--verbose          Show detailed output
```

**Examples**:
```bash
# Import all prompts
ti mode import-prompts

# Import only Ti-specific prompts
ti mode import-prompts --category ti,integration,quality

# Import generic prompts
ti mode import-prompts --category generic

# Dry run to preview
ti mode import-prompts --dry-run --verbose
```

## Import Workflow

### Step 1: Scan Source Directory
```go
files, err := filepath.Glob(filepath.Join(sourceDir, "*.md"))
```

### Step 2: Filter Files
- Skip `INDEX.md`, `BEADS_WORKFLOW.md`, `router-prompts.md` (meta files)
- Skip `enhanced-*.md` (already mapped in Phase 1.19)
- Skip `generic/` directory (unless requested)

### Step 3: Parse Each File
```go
prompt, err := parser.Parse(file)
```

### Step 4: Apply Filters
```go
if !matchesCategory(prompt.Category, config.CategoryFilter) {
    report.Skipped++
    continue
}
```

### Step 5: Convert to Mode
```go
mode, err := mapper.PromptToMode(prompt)
```

### Step 6: Check for Conflicts
```go
existing, err := db.GetMode(mode.ID)
if err == nil && existing != nil && !config.Overwrite {
    report.Skipped++
    continue
}
```

### Step 7: Save Mode
```go
if !config.DryRun {
    if err := db.SaveMode(mode); err != nil {
        report.Failed++
        continue
    }
}
```

### Step 8: Generate Report
```go
return &PromptImportReport{
    Total:    len(files),
    Imported: imported,
    Skipped:  skipped,
    Failed:   failed,
    Duration: time.Since(start),
}
```

## Validation Rules

### Content Validation
- Content must not be empty
- Must have at least one heading (# Title)
- Must be valid markdown

### Filename Validation
- Must end with `.md`
- Must not be a meta file (INDEX.md, BEADS_WORKFLOW.md)
- Must match expected pattern

### Mode ID Validation
- Must be unique
- Must be valid kebab-case
- Must not conflict with existing modes

## Error Handling

### Parse Errors
- Invalid markdown
- Empty content
- No heading found

### Mapping Errors
- Invalid filename pattern
- Cannot infer category
- Cannot infer priority

### Storage Errors
- Permission denied
- Disk full
- Database error

## Testing

### Unit Tests
- Test parser with valid/invalid markdown
- Test mapper with various filenames
- Test validator with edge cases

### Integration Tests
- Test full import workflow
- Test with dry-run mode
- Test with category filters
- Test conflict handling

### Manual Tests
- Import all prompts
- Import subset by category
- Handle conflicts
- Verify prompt content in modes

## Implementation Priority

### Phase 1: Core Import (P0)
- Prompt parser (plain markdown)
- Prompt mapper (filename → mode ID)
- Import manager
- CLI command

### Phase 2: Filters (P1)
- Category filtering
- Priority filtering
- Skip enhanced prompts
- Skip meta files

### Phase 3: Advanced (P2)
- Generic prompt support
- Custom mapping table
- Prompt versioning
- Prompt validation

## Future Enhancements

1. **Prompt Versioning**: Track prompt versions for updates
2. **Prompt Dependencies**: Import dependent prompts automatically
3. **Prompt Templates**: Generate prompts from templates
4. **Prompt Validation**: LLM-based prompt quality validation
5. **Prompt A/B Testing**: Compare prompt effectiveness
6. **Prompt Analytics**: Track prompt usage and performance

## Documentation

- User guide: How to import prompts
- Developer guide: How to extend parser
- Prompt format guide: How to write prompts
- Migration guide: From old to new format
