# Learning to TiBrain Integration

> **Last Updated**: 2026-05-04
> **Purpose**: Tự động capture learnings từ repo study sessions vào TiBrain Hub

---

## 🎯 Overview

Workflow này cho phép **tự động add learnings** từ repo study sessions vào TiBrain Hub, giúp:
- **Persistent Knowledge**: Learnings được lưu trữ centralized
- **Searchable**: Có thể search và recall learnings qua MCP tools
- **Cross-Session**: Learnings có thể được recall giữa các CLI sessions
- **Structured**: Learnings được tổ chức với quality scores, tags, categories

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Learning Session                         │
│  (Studying Repo: Ti CLI, Cherry Studio, Donut Browser)     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ↓ Extract Key Learnings
┌─────────────────────────────────────────────────────────────┐
│         add-learning-to-tibrain.ps1 Script                 │
│  - Convert learning to Tool struct                         │
│  - Call TiBrain Hub API                                    │
│  - Verify registration                                     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ↓ POST /v1/tibrain/tool/register
┌─────────────────────────────────────────────────────────────┐
│              TiBrain Hub (port 1810)                        │
│  - Store in tool_registry table                           │
│  - Searchable via MCP tools                                │
│  - Recallable across CLI sessions                          │
└─────────────────────────────────────────────────────────────┘
```

---

## 🚀 Quick Start

### Step 1: Start TiBrain Hub

```powershell
cd Z:\10_WORKPLACE\Ti\apps\router\tibrain
.\tibrain-server.exe --port 1810
```

### Step 2: Add Learning

```powershell
cd Z:\10_WORKPLACE\Ti\Ti-learning-lab\01_Learning
.\add-learning-to-tibrain.ps1 `
    -RepoName "ti-cli" `
    -PatternName "microkernel-plugin-architecture" `
    -Description "Microkernel pattern with gRPC communication for extensible CLI architecture" `
    -Category "cli" `
    -Family "microkernel" `
    -Tags "architecture,plugin,gRPC"
```

### Step 3: Verify Learning

```powershell
curl http://localhost:1810/v1/tibrain/tool/list
```

---

## 📋 Script Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `-RepoName` | ✅ Yes | - | Name of the repo being studied |
| `-PatternName` | ✅ Yes | - | Name of the pattern/learning |
| `-Description` | ✅ Yes | - | Detailed description of the learning |
| `-Category` | ❌ No | "learning" | Category for the learning |
| `-Family` | ❌ No | "general" | Family/tech stack family |
| `-QualityScore` | ❌ No | 85 | Quality score (0-100) |
| `-SecurityScore` | ❌ No | 80 | Security score (0-100) |
| `-Tags` | ❌ No | "architecture,best-practices" | Comma-separated tags |
| `-TibrainURL` | ❌ No | "http://localhost:1810" | TiBrain Hub URL |

---

## 🎨 Learning Capture Template

Khi học repo mới, extract learnings theo template này:

```markdown
# Learning: [Repo Name]

## Architecture Patterns
- Pattern 1: Description
- Pattern 2: Description

## Best Practices
- Practice 1: Description
- Practice 2: Description

## Code Patterns
- Pattern 1: Code example
- Pattern 2: Code example

## Tech Stack
- Language: [Go/Rust/Python/TypeScript]
- Frameworks: [List]
- Libraries: [List]

## Key Learnings
- Learning 1: Description
- Learning 2: Description

## Integration Points
- How to integrate with other components
- API endpoints
- Database schemas
```

---

## 📝 Example Usage

### Example 1: Ti CLI Architecture

```powershell
.\add-learning-to-tibrain.ps1 `
    -RepoName "ti-cli" `
    -PatternName "microkernel-plugin-architecture" `
    -Description "Microkernel pattern with gRPC communication for extensible CLI architecture" `
    -Category "cli" `
    -Family "microkernel" `
    -Tags "architecture,plugin,gRPC"
```

### Example 2: Cherry Studio Electron

```powershell
.\add-learning-to-tibrain.ps1 `
    -RepoName "cherry-studio" `
    -PatternName "electron-react-architecture" `
    -Description "Electron desktop app with React 19 frontend and Node.js backend, Redux Toolkit for state management" `
    -Category "desktop-app" `
    -Family "electron" `
    -Tags "electron,react,redux"
```

### Example 3: Donut Browser Rust

```powershell
.\add-learning-to-tibrain.ps1 `
    -RepoName "donut-browser" `
    -PatternName "rust-tauri-browser-automation" `
    -Description "Rust + Tauri desktop app for anti-detect browser with fingerprint spoofing and MCP integration" `
    -Category "browser" `
    -Family "rust" `
    -Tags "rust,tauri,browser,fingerprint"
```

---

## 🔍 MCP Integration

Sau khi learnings được captured, agents có thể:

### Search Learnings

```bash
# Via MCP tool
tibrain_search_skills --query "microkernel plugin"
```

### Recall Learnings

```bash
# Get specific learning
curl http://localhost:1810/v1/tibrain/tool?id=learning-ti-cli-microkernel-plugin-architecture
```

### List Learnings by Category

```bash
# List all learnings
curl http://localhost:1810/v1/tibrain/tool/list

# List by category
curl http://localhost:1810/v1/tibrain/tool/list?category=cli
```

---

## 📊 Tool Struct Mapping

Learnings được convert thành Tool struct với các fields:

```go
Tool{
    ID: "learning-[repo-name]-[pattern-name]",
    Name: "[Repo Name] - [Pattern Name]",
    Description: "Detailed description",
    Category: "learning", // hoặc custom category
    Family: "[tech-stack-family]",
    Tags: "architecture,best-practices,pattern",
    QualityScore: 85, // 0-100
    SecurityScore: 80, // 0-100
    BestPracticesScore: 90, // 0-100
    Source: "[repo-name]",
    SkillLevel: "l2", // l1=beginner, l2=intermediate, l3=advanced
    QualityTier: "platinum", // bronze, silver, gold, platinum
    SecurityTier: "hardened", // basic, hardened, critical
}
```

---

## 🔧 Troubleshooting

### "Connection refused"

**Problem**: Không thể kết nối đến TiBrain Hub

**Solution**:
```powershell
# Start TiBrain Hub
cd Z:\10_WORKPLACE\Ti\apps\router\tibrain
.\tibrain-server.exe --port 1810

# Verify it's running
curl http://localhost:1810/health
```

### "Learning registered but not found"

**Problem**: Learning được register nhưng không thể tìm thấy

**Solution**:
```powershell
# List all tools (không filter)
curl http://localhost:1810/v1/tibrain/tool/list

# Search by tool ID
curl http://localhost:1810/v1/tibrain/tool?id=learning-[repo-name]-[pattern-name]
```

### "PowerShell parsing error"

**Problem**: Script bị lỗi parsing

**Solution**:
```powershell
# Sử dụng -Command thay vì -File
powershell -Command ".\add-learning-to-tibrain.ps1 -RepoName 'test' -PatternName 'test' -Description 'test'"
```

---

## 📚 Integration with Learning Path

Workflow này được tích hợp vào learning path của từng repo:

### Ti CLI Learning Path

```markdown
### Week 1-2: Ti CLI Architecture

**Step 3: Capture Learnings to TiBrain**
```powershell
# Sau khi học mỗi pattern, capture learning:
.\add-learning-to-tibrain.ps1 `
    -RepoName "ti-cli" `
    -PatternName "microkernel-plugin-architecture" `
    -Description "Microkernel pattern with gRPC communication for extensible CLI" `
    -Category "cli" `
    -Family "microkernel" `
    -Tags "architecture,plugin,gRPC"
```
```

### Donut Browser Learning Path

```markdown
### Week 5-6: Donut Browser Complete Stack

**Step 3: Capture Learnings to TiBrain**
```powershell
# Sau khi học Rust + Tauri patterns:
.\add-learning-to-tibrain.ps1 `
    -RepoName "donut-browser" `
    -PatternName "rust-tauri-browser-automation" `
    -Description "Rust + Tauri desktop app for anti-detect browser" `
    -Category "browser" `
    -Family "rust" `
    -Tags "rust,tauri,browser,fingerprint"
```
```

---

## 🎯 Best Practices

### 1. Capture Learnings Immediately

Capture learning ngay sau khi học pattern, không đợi cuối session.

### 2. Use Descriptive Pattern Names

Sử dụng tên pattern mô tả và cụ thể:
- ✅ Good: "microkernel-plugin-architecture"
- ❌ Bad: "pattern-1"

### 3. Use Appropriate Categories

Sử dụng category phù hợp để dễ search:
- `cli` cho CLI patterns
- `desktop-app` cho desktop apps
- `browser` cho browser automation
- `api` cho API patterns
- `database` cho database patterns

### 4. Use Relevant Tags

Sử dụng tags để dễ filter:
- `architecture,best-practices` cho general patterns
- `rust,tauri` cho Rust-specific patterns
- `electron,react` cho Electron-specific patterns

### 5. Set Appropriate Scores

Set quality và security scores dựa trên:
- Quality: Mức độ tốt của pattern (0-100)
- Security: Ảnh hưởng security của pattern (0-100)

---

## 🔗 Related

- **TiBrain Hub**: `Z:\10_WORKPLACE\Ti\apps\router\tibrain\`
- **MCP Server**: `Z:\10_WORKPLACE\Ti\apps\mcp\server\tibrain-mcp\`
- **Learning Lab**: `Z:\10_WORKPLACE\Ti\Ti-learning-lab\`
- **Repo Learning Paths**: `Ti-learning-lab/01_Learning/README.md`

---

*Last Updated: 2026-05-04*
