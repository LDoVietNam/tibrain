# Ti Learning Lab - 01_Learning

> **Purpose**: Learning materials, patterns, and repository analysis for Ti ecosystem development
> **Location**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\01_Learning\`

---

## 🎯 Purpose

Directory này chứa:
- **Repository Analysis**: Analysis của các open-source repos để học patterns
- **Learning Paths**: Structured learning paths cho từng tech stack
- **Pattern Capture**: Scripts và templates để capture learnings
- **TiBrain Integration**: Workflow để tự động add learnings vào TiBrain Hub

---

## 📁 Structure

```
01_Learning/
├── add-learning-to-tibrain.ps1      # Script để add learnings vào TiBrain
├── LEARNING_TO_TIBRAIN.md           # Documentation cho TiBrain integration
├── README.md                         # This file
├── budai/                            # Repo analysis examples
│   └── budai_analysis.md
├── CLI/                              # Ti CLI learning materials
└── [other-repos]/                    # Other repo analyses
    └── [repo_analysis.md]
```

---

## 🚀 Quick Start

### 1. Repository Analysis

Khi analyzing một repo mới:

```bash
# Tạo folder cho repo
mkdir Z:\10_WORKPLACE\Ti\Ti-learning-lab\01_Learning\[repo-name]

# Tạo analysis file
touch Z:\10_WORKPLACE\Ti\Ti-learning-lab\01_Learning\[repo-name]\[repo-name]_analysis.md
```

Sử dụng **Analysis Template** từ README gốc.

### 2. Capture Learnings to TiBrain

Sau khi học patterns từ repo:

```powershell
# Start TiBrain Hub
cd Z:\10_WORKPLACE\Ti\apps\router\tibrain
.\tibrain-server.exe --port 1810

# Add learning
cd Z:\10_WORKPLACE\Ti\Ti-learning-lab\01_Learning
.\add-learning-to-tibrain.ps1 `
    -RepoName "[repo-name]" `
    -PatternName "[pattern-name]" `
    -Description "[pattern-description]" `
    -Category "[category]" `
    -Family "[family]" `
    -Tags "[tags]"
```

Xem chi tiết: [LEARNING_TO_TIBRAIN.md](LEARNING_TO_TIBRAIN.md)

---

## 📋 Learning Paths

### Available Learning Paths

| Repo | Tech Stack | Duration | Status |
|------|------------|----------|--------|
| **Ti CLI** | Go, gRPC, Microkernel | 2 weeks | 📝 Planned |
| **Cherry Studio** | Electron, React 19, TypeScript | 2 weeks | 📝 Planned |
| **Donut Browser** | Rust, Tauri, Chromium | 2 weeks | 📝 Planned |
| **Free Claude Code** | Python, FastAPI | 1 week | 📝 Planned |
| **Ti TUI** | Go, Bubble Tea | 1 week | 📝 Planned |
| **MCP Hub** | Go, Python, JSON-RPC | 1 week | 📝 Planned |

### Learning Path Template

```markdown
# [Repo Name] Learning Path

## Week [X]-[Y]: [Topic]

### Study Path
1. Read README.md for features overview
2. Explore architecture
3. Study key patterns
4. Practice with build/test

### Key Learnings
- Pattern 1: Description
- Pattern 2: Description

### Capture to TiBrain
```powershell
.\add-learning-to-tibrain.ps1 `
    -RepoName "[repo-name]" `
    -PatternName "[pattern-name]" `
    -Description "[description]"
```
```

---

## 📊 Analyzed Repositories

|| Repository | Status | Date | Notes |
|---|------------|--------|------|-------|
| 1 | [budai](https://github.com/manhcuongk55/budai) | ✅ Analyzed | 2026-04-30 | Python-based AI ethics system with MCP integration potential |
| 2 | Ti CLI | 📝 Planned | - | Microkernel + Plugin Architecture (Go, gRPC) |
| 3 | Cherry Studio | 📝 Planned | - | Electron + React 19 desktop app |
| 4 | Donut Browser | 📝 Planned | - | Rust + Tauri anti-detect browser |
| 5 | Free Claude Code | 📝 Planned | - | Python FastAPI proxy for Claude Code |
| 6 | Ti TUI | 📝 Planned | - | Go Bubble Tea terminal UI |
| 7 | MCP Hub | 📝 Planned | - | Go + Python MCP servers |

---

## 🔍 Analysis Template

Khi analyzing một repo mới, bao gồm:

### 1. Tech Stack
- Language: [Go/Rust/Python/TypeScript]
- Frameworks: [List]
- Libraries: [List]
- Build System: [Make/Cargo/npm/Go]

### 2. Purpose
- Primary use case
- Target users
- Key features

### 3. Architecture
- System design
- Patterns used
- Component structure

### 4. Strengths
- What we can learn
- Best practices
- Innovative solutions

### 5. Weaknesses
- What to avoid
- Anti-patterns
- Technical debt

### 6. Integration Potential
- How it could work with Ti CLI
- API endpoints
- Database schemas

### 7. Recommendations
- Actionable insights
- Priority patterns to learn
- Implementation suggestions

---

## 🧠 TiBrain Integration

### Workflow

1. **Study Repo**: Learn patterns from repository
2. **Extract Learnings**: Identify key patterns and best practices
3. **Capture to TiBrain**: Use `add-learning-to-tibrain.ps1` script
4. **Search & Recall**: Use MCP tools to search and recall learnings

### Benefits

- **Persistent Knowledge**: Learnings được lưu centralized
- **Searchable**: Có thể search qua MCP tools
- **Cross-Session**: Recall learnings giữa CLI sessions
- **Structured**: Organized với quality scores, tags, categories

### MCP Tools

```bash
# Search learnings
tibrain_search_skills --query "microkernel plugin"

# List all learnings
curl http://localhost:1810/v1/tibrain/tool/list

# Get specific learning
curl http://localhost:1810/v1/tibrain/tool?id=learning-[repo]-[pattern]
```

Xem chi tiết: [LEARNING_TO_TIBRAIN.md](LEARNING_TO_TIBRAIN.md)

---

## 📖 Usage

```bash
# View all analyzed repos
ls Z:\10_WORKPLACE\Ti\Ti-learning-lab\01_Learning\

# View specific analysis
cat Z:\10_WORKPLACE\Ti\Ti-learning-lab\01_Learning\budai\budai_analysis.md

# Add learning to TiBrain
cd Z:\10_WORKPLACE\Ti\Ti-learning-lab\01_Learning
.\add-learning-to-tibrain.ps1 -RepoName "test" -PatternName "test" -Description "test"

# List all learnings in TiBrain
curl http://localhost:1810/v1/tibrain/tool/list
```

---

## 🎯 Convention

- **Folder naming**: `[repo-name]` (lowercase, hyphen-separated)
- **File naming**: `[repo-name]_analysis.md`
- **Learning ID**: `learning-[repo-name]-[pattern-name]` (lowercase, hyphen-separated)
- **Categories**: `cli`, `desktop-app`, `browser`, `api`, `database`, `general`
- **Families**: `microkernel`, `electron`, `rust`, `python`, `go`, `general`

---

## 🔗 Related

- **TiBrain Hub**: `Z:\10_WORKPLACE\Ti\apps\router\tibrain\`
- **MCP Server**: `Z:\10_WORKPLACE\Ti\apps\mcp\server\tibrain-mcp\`
- **Learning Lab Root**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\`
- **Taskboard**: `Z:\10_WORKPLACE\Ti\taskboard\`

---

*Last Updated: 2026-05-04*
