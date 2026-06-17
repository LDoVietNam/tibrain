---
tags: ["tibrain", "documentation", "testing", "cli", "provider-claude"]
scopes: ["code", "providers", "cli", "tibrain"]
last_updated: 2026-05-22
---
# Ti CLI & Claude Code Format Integration

> **Last Updated:** 2026-04-29  
> **Purpose:** Hướng dẫn integrate Claude Code format vào Ti CLI

---

## 📊 Hiện Tại Ti CLI Đã Có

### **Claude Code Format (Đã có ở Z:\Ti\)**

```
Z:\Ti\.claude\
├── CLAUDE.md              # Ti project overview
├── rules/
│   └── global.md         # Global coding rules
├── commands/
│   ├── test.md           # /test command
│   ├── build.md          # /build command
│   └── health.md         # /health command
├── agents/
│   ├── ti-architect.md   # Architecture review agent
│   └── security-guard.md # Secret leakage prevention agent
├── hooks/
│   └── pre-commit.md     # Pre-commit validation
└── skills/
    └── model-router.md   # Free model selection skill
```

**Global context:** `Z:\02_CORE\_cli\.config\claude\CLAUDE.md`

---

### **Ti CLI Mới (Z:\Ti\CLI)**

```
Z:\Ti\CLI\
├── .golangci.yml
├── Makefile
├── README.md
├── build.ps1
├── cmd/                  # CLI entry points
├── internal/             # Core packages
├── pkg/                  # Public packages
└── scripts/              # Build scripts
```

**Status:** Đang trống, đang rebuild với Microkernel + Plugin architecture

---

## 🎯 Cách Integrate Claude Code Format vào Ti CLI

### **Option 1: Sử dụng .claude/ Folder Hiện Tại (RECOMMENDED)**

**Cách làm:**
- Ti CLI reference `.claude/` folder từ `Z:\Ti\.claude\`
- Không cần duplicate
- Cả Claude Code và Ti CLI dùng chung

**Implementation:**
```powershell
# Trong Ti CLI, reference .claude/ folder
# Claude Code tự động đọc .claude/CLAUDE.md
# Ti CLI có thể đọc .claude/ folder cho context
```

**Benefits:**
- Không duplicate data
- Cả 2 IDE dùng chung
- Dễ maintain

---

### **Option 2: Tạo .claude/ Folder Riêng cho Ti CLI**

**Cách làm:**
- Tạo `.claude/` folder trong `Z:\Ti\CLI\`
- Copy/modify từ `Z:\Ti\.claude\`
- Ti CLI có context riêng

**Structure:**
```
Z:\Ti\CLI\.claude\
├── CLAUDE.md              # Ti CLI specific context
├── rules/
│   └── global.md         # Ti CLI coding rules
├── commands/
│   ├── test.md           # Ti CLI test commands
│   ├── build.md          # Ti CLI build commands
│   └── plugin.md         # Ti CLI plugin commands
├── agents/
│   └── cli-architect.md  # CLI architecture review agent
├── hooks/
│   └── pre-commit.md     # Pre-commit validation
└── skills/
    └── plugin-dev.md     # Plugin development skill
```

**Benefits:**
- Context riêng cho Ti CLI
- Không conflict với Ti root
- Dễ customize cho CLI-specific tasks

---

### **Option 3: Sử dụng AGENTS.md (Standard)**

**Cách làm:**
- Tạo `AGENTS.md` trong `Z:\Ti\CLI\`
- Sử dụng standard format (supported bởi Claude Code, Windsurf, v.v.)
- Ti CLI và các IDE khác có thể dùng chung

**Content:**
```markdown
# AGENTS.md - Ti CLI

## Project Overview
Ti CLI with Microkernel + Plugin architecture

## Tech Stack
- Go 1.25
- Cobra (CLI framework)
- gRPC (plugin communication)
- SQLite (data persistence)

## Commands
```bash
# Build
go build -o bin/ti.exe .

# Test
go test ./...

# Plugin
ti plugin list
ti plugin install <name>
```

## Directory Structure
```
cmd/          # CLI entry points
internal/     # Core packages
pkg/          # Public packages
```

## Coding Style
- Use 4 spaces indentation
- Follow Go conventions
- Add godoc comments
- Keep functions under 50 lines

## Skills Location
Z:\Ti\best_source\skills\

## Workflow
1. Check beads (Z:\Ti\taskboard\beads.md)
2. Identify task type
3. Choose location
4. Create/update code
5. Test & validate
6. Update beads
```

**Benefits:**
- Standard format
- Supported bởi nhiều IDE
- Dễ integrate với skills repository

---

## 🎯 Recommendation: Option 3 (AGENTS.md) + Option 1 (Reference .claude/)

**Hybrid approach:**

1. **Tạo AGENTS.md trong Z:\Ti\CLI\** cho Ti CLI specific context
2. **Reference .claude/ từ Z:\Ti\.claude\** cho global context
3. **Sử dụng skills từ Z:\Ti\best_source\skills\**

**Implementation:**
```
Z:\Ti\CLI\
├── AGENTS.md              # Ti CLI specific context
├── .claude/               # Junction to Z:\Ti\.claude\ (optional)
├── cmd/
├── internal/
└── pkg/
```

**Benefits:**
- Ti CLI có context riêng
- Không duplicate global context
- Standard format (AGENTS.md)
- Dễ integrate với skills

---

## 📝 AGENTS.md Template cho Ti CLI

```markdown
# AGENTS.md - Ti CLI

> **Version**: 1.0.0  
> **Last Updated**: 2026-04-29

## Project Overview

Ti CLI with Microkernel + Plugin architecture for enterprise-grade CLI development.

## Tech Stack

- **Go** 1.25
- **Cobra** - CLI framework
- **gRPC** - Plugin communication
- **SQLite** - Data persistence
- **Protobuf** - Type-safe communication

## Directory Structure

```
Z:\Ti\CLI\
├── cmd/
│   ├── root.go           # Main CLI entry
│   ├── plugin/           # Plugin commands
│   └── beads/            # BEADS commands
├── internal/
│   ├── core/             # Microkernel core
│   │   ├── plugin_manager.go
│   │   ├── plugin_bus.go
│   │   └── plugin_registry.go
│   └── agents/           # Plugin implementations
│       ├── devin/        # Devin plugin
│       └── native/       # Native plugins
├── pkg/
│   ├── beads/            # BEADS types
│   └── brain/            # Intelligence engine
└── scripts/
```

## Commands

```bash
# Build
go build -o bin/ti.exe .

# Test
go test ./...

# Plugin commands
ti plugin list
ti plugin install <name>
ti plugin remove <name>
ti plugin enable <name>
ti plugin disable <name>

# BEADS commands
ti beads graph init
ti beads kanban export
```

## Coding Style

- Use 4 spaces indentation
- Follow Go conventions (gofmt)
- Add godoc comments to exported functions
- Keep functions under 50 lines
- Use meaningful variable names
- Handle errors explicitly

## Plugin Development

**Plugin Interface:**
```go
type Plugin interface {
    Name() string
    Version() string
    Init(ctx context.Context) error
    Execute(ctx context.Context, args []string) error
    Shutdown(ctx context.Context) error
}
```

**Best Practices:**
- Implement Plugin interface
- Handle context cancellation
- Graceful shutdown
- Resource cleanup

## Skills Location

Z:\Ti\best_source\skills\

## Skill Mapping

| Task Type | Skill Location |
|-----------|---------------|
| Go patterns | golang-patterns/ |
| CLI development | TBD (create new) |
| Plugin development | TBD (create new) |
| gRPC | TBD (create new) |

## Workflow

1. Check beads (Z:\Ti\taskboard\beads.md)
2. Identify task type (CLI core, plugin, BEADS)
3. Research skills (Z:\Ti\best_source\skills\)
4. Choose location (cmd/, internal/, pkg/)
5. Create/update code
6. Test & validate (go test ./...)
7. Update beads (Z:\Ti\taskboard\beads.md)

## Global Context

Claude Code global context: `Z:\Ti\.claude\CLAUDE.md`

## Related Resources

- Ti AGENTS.md: `Z:\Ti\AGENTS.md`
- Ti README: `Z:\Ti\README.md`
- Skills: `Z:\Ti\best_source\skills\`
- Beads: `Z:\Ti\taskboard\beads.md`
```

---

## 🔗 Integration với Best Practices

### **1. Beads Protocol**

Ti CLI follow beads protocol:
- Step 0: Check Beads (Z:\Ti\taskboard\beads.md)
- Step 8: Update Beads (Z:\Ti\taskboard\beads.md)

### **2. Skills Integration**

Ti CLI reference skills từ `Z:\Ti\best_source\skills\`:
- Go patterns: `golang-patterns/`
- CLI patterns: (create new folder)
- Plugin patterns: (create new folder)

### **3. Claude Code Integration**

Ti CLI có thể:
- Reference `.claude/` từ `Z:\Ti\.claude\`
- Sử dụng AGENTS.md format
- Integrate với Claude Code workflow

---

## 🎯 Implementation Steps

### **Step 1: Tạo AGENTS.md trong Z:\Ti\CLI\**

```powershell
# Tạo AGENTS.md
New-Item -Path "Z:\Ti\CLI\AGENTS.md" -ItemType File
# Copy template từ trên
```

### **Step 2: Tạo Junction đến .claude/ (Optional)**

```powershell
# Nếu muốn reference global .claude/
cmd /c mklink /J "Z:\Ti\CLI\.claude" "Z:\Ti\.claude"
```

### **Step 3: Update Ti CLI README**

```markdown
## AI Agent Integration

Ti CLI supports AI agent workflows with:
- AGENTS.md for project context
- Skills integration from Z:\Ti\best_source\skills\
- Beads protocol for task tracking
- Claude Code format compatibility
```

### **Step 4: Test Integration**

```powershell
# Test với Claude Code
# Claude Code nên đọc AGENTS.md và .claude/ folder

# Test với Windsurf
# Windsurf nên đọc AGENTS.md
```

---

## 📊 Summary

**Có thể dùng định dạng Claude Code cho Ti CLI:**
1. ✅ Sử dụng AGENTS.md (standard format)
2. ✅ Reference .claude/ folder (optional)
3. ✅ Integrate skills từ Z:\Ti\best_source\skills\
4. ✅ Follow beads protocol
5. ✅ Compatible với Claude Code và Windsurf

**Recommendation:**
- Tạo AGENTS.md trong Z:\Ti\CLI\
- Reference .claude/ từ Z:\Ti\.claude\ (optional)
- Integrate skills từ Z:\Ti\best_source\skills\
- Follow beads protocol

---

## 🔗 Related Documents

- `Z:\Ti\AGENTS.md` - Ti global AGENTS.md
- `Z:\Ti\.claude\CLAUDE.md` - Claude Code context
- `Z:\Ti\CLI\README.md` - Ti CLI README
- `Z:\Ti\best_source\skills\` - Skills repository
- `Z:\Ti\taskboard\beads.md` - Beads log
