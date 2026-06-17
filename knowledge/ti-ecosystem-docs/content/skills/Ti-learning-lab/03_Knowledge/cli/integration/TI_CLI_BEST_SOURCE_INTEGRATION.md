---
tags: ["tibrain", "cli", "testing", "documentation", "skill"]
scopes: ["code", "cli", "tibrain"]
last_updated: 2026-05-22
---
# Ti CLI Integration with best_source (Dynamic Read)

> **Date**: 2026-04-29  
> **Purpose**: Design for Ti CLI to dynamically read skills from Z:\Ti\best_source

---

## Architecture

```
User Command (ti skill list)
    ↓
Ti CLI Core (cmd/skill.go)
    ↓
Config (internal/config/config.go) → SkillsDir field
    ↓
Skill Reader (internal/skills/reader.go) → Read markdown files
    ↓
Z:\Ti\best_source\skills\*.md
```

---

## 1. Config Extension

**File**: `internal/config/config.go`

Add field:
```go
type Config struct {
    // ... existing fields ...
    SkillsDir string `json:"skills_dir,omitempty"`
}
```

Default value in `Default()`:
```go
cfg := &Config{
    // ... existing defaults ...
    SkillsDir: "Z:\\Ti\\best_source\\skills", // Windows path
    // Alternative: use env var or detect from project root
}
```

Fallback logic:
```go
// Priority:
// 1. Config file (skills_dir)
// 2. Environment variable: TI_SKILLS_DIR
// 3. Auto-detect: find nearest .git root + /best_source/skills
// 4. Default: Z:\Ti\best_source\skills
```

---

## 2. Skill Package

**New package**: `internal/skills/`

### File: `internal/skills/skill.go`

```go
package skills

type Skill struct {
    Name        string            `json:"name"`
    Path        string            `json:"path"`
    Type        string            `json:"type"`          // meta-skill, pattern, rubric
    Priority    string            `json:"priority"`      // critical, high, medium, low
    Description string            `json:"description"`
    Tags        []string          `json:"tags,omitempty"`
    Content     string            `json:"content,omitempty"` // Full markdown content
    Metadata    map[string]string `json:"metadata,omitempty"` // YAML frontmatter
}
```

### File: `internal/skills/reader.go`

```go
package skills

import (
    "os"
    "path/filepath"
    "strings"
)

type Reader struct {
    SkillsDir string
}

func NewReader(skillsDir string) *Reader {
    return &Reader{SkillsDir: skillsDir}
}

// List all skills in directory
func (r *Reader) List() ([]Skill, error) {
    entries, err := os.ReadDir(r.SkillsDir)
    if err != nil {
        return nil, err
    }
    
    var skills []Skill
    for _, entry := range entries {
        if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
            continue
        }
        
        path := filepath.Join(r.SkillsDir, entry.Name())
        skill, err := r.parseSkill(path)
        if err != nil {
            continue // Log warning but don't fail
        }
        skills = append(skills, skill)
    }
    return skills, nil
}

// Get single skill by name
func (r *Reader) Get(name string) (*Skill, error) {
    // Try exact match
    path := filepath.Join(r.SkillsDir, name+".md")
    if _, err := os.Stat(path); err == nil {
        return r.parseSkill(path)
    }
    
    // Try fuzzy match
    skills, _ := r.List()
    for _, s := range skills {
        if strings.EqualFold(s.Name, name) {
            return &s, nil
        }
    }
    return nil, fmt.Errorf("skill not found: %s", name)
}

// Search skills by tag or keyword
func (r *Reader) Search(query string) ([]Skill, error) {
    skills, _ := r.List()
    var results []Skill
    for _, s := range skills {
        if strings.Contains(strings.ToLower(s.Name), strings.ToLower(query)) ||
           strings.Contains(strings.ToLower(s.Description), strings.ToLower(query)) {
            results = append(results, s)
        }
    }
    return results, nil
}

func (r *Reader) parseSkill(path string) (Skill, error) {
    // Parse markdown: extract YAML frontmatter + content
    // Implementation: read file, split on "---", parse YAML, store content
}
```

---

## 3. CLI Commands

**File**: `cmd/skill.go` (new)

```go
package cmd

var skillCmd = &cobra.Command{
    Use:   "skill",
    Short: "Manage and query skills from best_source",
    Long:  `Read, search, and apply skills from the best_source repository.`,
}

var skillListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all available skills",
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, _ := config.LoadAuto()
        reader := skills.NewReader(cfg.SkillsDir)
        skills, err := reader.List()
        if err != nil {
            return fmt.Errorf("failed to read skills: %w", err)
        }
        
        fmt.Printf("Skills directory: %s\n\n", cfg.SkillsDir)
        for _, s := range skills {
            fmt.Printf("  %-30s [%s] %s\n", s.Name, s.Priority, s.Description)
        }
        return nil
    },
}

var skillShowCmd = &cobra.Command{
    Use:   "show [name]",
    Short: "Display a skill's content",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, _ := config.LoadAuto()
        reader := skills.NewReader(cfg.SkillsDir)
        skill, err := reader.Get(args[0])
        if err != nil {
            return err
        }
        
        fmt.Printf("# %s\n", skill.Name)
        fmt.Printf("Type: %s | Priority: %s\n\n", skill.Type, skill.Priority)
        fmt.Println(skill.Content)
        return nil
    },
}

var skillSearchCmd = &cobra.Command{
    Use:   "search [query]",
    Short: "Search skills by keyword",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, _ := config.LoadAuto()
        reader := skills.NewReader(cfg.SkillsDir)
        results, err := reader.Search(args[0])
        if err != nil {
            return err
        }
        
        fmt.Printf("Found %d skills for '%s':\n\n", len(results), args[0])
        for _, s := range results {
            fmt.Printf("  %-30s [%s] %s\n", s.Name, s.Type, s.Description)
        }
        return nil
    },
}

func init() {
    rootCmd.AddCommand(skillCmd)
    skillCmd.AddCommand(skillListCmd)
    skillCmd.AddCommand(skillShowCmd)
    skillCmd.AddCommand(skillSearchCmd)
}
```

---

## 4. Usage Examples

```powershell
# List all skills
> ti skill list
Skills directory: Z:\Ti\best_source\skills

  react-agent-pattern           [critical] ReAct workflow pattern for AI agents
  ai-response-quality-rubric  [high] 5-star quality rating system
  agent-training-prompt         [high] Training prompt templates
  go-conventions                [medium] Go coding standards
  ...

# Show specific skill
> ti skill show react-agent-pattern
# react-agent-pattern
Type: meta-skill | Priority: critical

## Overview
ReAct (Reason + Act) combined with Chain-of-Thought...
[full content]

# Search skills
> ti skill search "workflow"
Found 3 skills for 'workflow':

  react-agent-pattern           [meta-skill] ReAct workflow pattern
  agent-training-prompt         [meta-skill] Training prompt templates
  ...

# Check if skills directory exists (for debugging)
> ti config get skills_dir
Z:\Ti\best_source\skills
```

---

## 5. Integration with Other Commands

### Plugin commands can reference skills:
```go
// In plugin execute, look up skill first
skill, err := skills.NewReader(cfg.SkillsDir).Get(task)
if err == nil {
    // Apply skill pattern
}
```

### Agent commands can load skills as context:
```go
// Load all critical skills for agent initialization
reader := skills.NewReader(cfg.SkillsDir)
allSkills, _ := reader.List()
for _, s := range allSkills {
    if s.Priority == "critical" {
        agentContext += s.Content + "\n\n"
    }
}
```

---

## 6. Error Handling

| Scenario | Behavior |
|----------|----------|
| SkillsDir not found | Warning: "Skills directory not found: {path}. Set TI_SKILLS_DIR or config skills_dir." |
| Single skill parse fail | Log warning, skip skill, continue listing others |
| No skills found | Return empty list with message |
| Config skills_dir empty | Use default path or env var |

---

## 7. Implementation Order

1. **Phase 1**: Add `SkillsDir` to Config (1 file change)
2. **Phase 2**: Create `internal/skills/` package (2 files)
3. **Phase 3**: Create `cmd/skill.go` commands (1 file)
4. **Phase 4**: Add skill reference to plugin/agent commands
5. **Phase 5**: Add tests

---

## Files to Create/Modify

| File | Action |
|------|--------|
| `internal/config/config.go` | Add `SkillsDir` field + default |
| `internal/skills/skill.go` | Create - Skill struct |
| `internal/skills/reader.go` | Create - Reader implementation |
| `cmd/skill.go` | Create - CLI commands |
| `internal/skills/reader_test.go` | Create - Unit tests |

---

## References

- Config package: `internal/config/config.go`
- Existing path fields: `MemoryPath`, `BEADSPath`, `QAPath`
- best_source location: `Z:\Ti\best_source\skills\`
