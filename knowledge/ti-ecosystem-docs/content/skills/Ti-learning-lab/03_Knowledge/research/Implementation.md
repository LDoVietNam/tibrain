# Agent Teams Implementation Guide

> **For**: Claude Code v2.1+ với Opus 4.6
> **Complexity**: Intermediate - Requires CLI + TMUX familiarity

---

## Bước 1: Cài Đặt TMUX

```bash
# Linux (WSL)
sudo apt install tmux

# macOS
brew install tmux

# Kiểm tra
tmux -V
```

---

## Bước 2: Cấu Hình settings.json

```json
{
  "experimental": {
    "agent": {
      "team": {
        "enabled": true,
        "tmuxIntegration": true,
        "maxMembers": 4
      }
    }
  },
  "modelConfig": {
    "captain": "claude-opus-4-6-thinking",
    "members": ["claude-sonnet-4-6", "gpt-5.1-codex-mini"]
  }
}
```

---

## Bước 3: Tạo Agent Config

Tạo thư mục `.claude/agents/`:

```bash
mkdir -p .claude/agents
```

### Agent Template

```markdown
---
name: <agent-name>
description: <mô tả ngắn>
tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - Bash(<commands>)
---

## System Prompt
You are a specialized <role>. Your job is to <task description>.

## Guidelines
1. <rule 1>
2. <rule 2>

## Output Format
Return kết quả dạng:
- Status: <pass/fail>
- Summary: <tóm tắt>
- Details: <chi tiết>
```

---

## Bước 4: Chạy Agent Teams

###CLI Commands:

```bash
# List available agents
claude --agents list

# Run với specific agent
claude --agent code-reviewer "Review the auth module"

# Run multiple agents in parallel
claude --agents code-writer,code-reviewer "Refactor the API"
```

### Trong Session:

```bash
# Gọi sub-agent từ main session
/use subagent <agent-name> <task>

# Ví dụ:
/use subagent code-reviewer "Check for SQL injection vulnerabilities"
```

---

## Bước 5: Memory & Context Sharing

### Tạo Memory File

```markdown
# .claude/memory/team-context.md
## Project Context
- Tech stack: Express.js + PostgreSQL
- API: REST on port 3000
- Auth: JWT

## Known Issues
- Bug in auth (fixed by Agent A)
- Memory leak investigated

## Shared Variables
- DB_HOST=localhost
- API_BASE=http://localhost:3000
```

### Trong Captain Prompt:

```
Before starting, read .claude/memory/team-context.md 
for project context and share with all members.
```

---

## Bước 6: Tạo Skills từ Workflow

Sau khi hoàn thành workflow thành công:

```bash
# Yêu cầu Claude tạo skill
/use skill-create "<workflow-name>"

# Ví dụ:
/use skill-create "api-research"
# → Tạo /api-research command để reuse
```

---

## Best Practices

### 1. Clear Role Definition
- Mỗi agent chỉ handle 1 responsibility
- Tránh overlap giữa agents
- Captain nên define "scope" rõ ràng

### 2. Tool Permission Minimal
- Chỉ enable tools cần thiết
- Không give shell access unless required
- Members có thể dùng cheaper models

### 3. Communication Pattern
- Dùng structured output (JSON/markdown)
- Include status + summary + details
- Captain tổng hợp trước khi report

### 4. Error Handling
- Each agent nên have fallback behavior
- Race condition: task auto-lock khi 1 agent start
- Captain nên handle agent failures gracefully

### 5. Resource Cleanup
- Captain tự động close members sau khi done
- Có thể keep some members alive for debugging

---

## Example: Code Review Team

### agents/code-writer.md
```markdown
---
name: Code Writer
description: Writes new features based on requirements
tools:
  - Read
  - Write
  - Edit
  - Glob
  - Bash(pnpm build)
---

You write code following project patterns.
Output format:
- Files changed: [...]
- Tests added: [...]
```

### agents/code-reviewer.md
```markdown
---
name: Code Reviewer  
description: Reviews code for quality and security
tools:
  - Read
  - Glob
  - Grep
  - Bash(pnpm test)
---

You review code for:
1. Security vulnerabilities
2. Code quality
3. Test coverage
Output format:
- Issues found: [...]
- Severity: <high/medium/low>
- Recommendations: [...]
```

### Usage:
```bash
# Sequential: Write then Review
claude --agent code-writer "Add user auth feature" 
# → Output → claude --agent code-reviewer "Review the auth feature"
```

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Agent timeout | Increase timeout trong settings |
| Tool not allowed | Add to tools array in agent config |
| Context overflow | Use summarization hoặc memory files |
| Communication fails | Use structured output format |
| TMUX not found | Cài đặt TMUX trước |

---

## Cleanup & Session End

- **Auto cleanup**: Captain tự động close members khi done
- **Manual keep**: Có thể keep members alive cho debugging
- **Resource release**: All sessions closed sau khi hoàn thành

---

*See also: README.md for overview*