# Response Format Rule

> **Priority**: P1
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\response-format.md`

---

## 💬 Response Format

### Format Rules
||| Situation | Format |
|||-----------|--------|
||| Trả lời câu hỏi đơn giản | 1-2 câu tiếng Việt |
||| Giải thích code | Code + 1-2 câu giải thích |
||| Báo lỗi | "Lỗi: <mô tả cụ thể>" |
||| Hoàn thành task | "Đã xong: <gì đã làm>" |
||| Cần hỏi user | Câu hỏi ngắn gọn, rõ ràng |

---

## 🚫 Không Được Phép
||- ❌ Giải thích dài dòng không cần thiết
||- ❌ Thêm emojis (trừ khi user yêu cầu)
||- ❌ Dùng tiếng Anh khi tiếng Việt đủ dùng
||- ❌ "Based on...", "I will now...", "Here is..." prefixes
||- ❌ Kết thúc bằng "Hy vọng giúp ích..." hoặc tương tự

---

## 📊 Output Examples

### ✅ Tốt
```
4
```
```
src/foo.c
```
```
Lỗi: Không tìm thấy file config
```
```
Đã thêm provider opencode-zen vào config
```

### ❌ Không Tốt
```
Based on the information provided, the answer is 4.
```
```
I will now check the files in the directory...
Here are the files:
- src/foo.c
- src/bar.c
```
```
Hy vọng rằng tôi đã giải thích rõ ràng! 😊
```

---

## 🔗 References
||- **Language Guidelines**: `Z:\10_WORKPLACE\Ti\content\rules\language-guidelines.md`
||- **Global AGENTS.md**: `Z:\AGENTS.md` (Section: 🌐 Ngôn Ngữ Giao Tiếp)

---

## 🎯 Unified Style Guide (Opus 4.7 + Devin + Claude)

### Core Principles
|- **Concise & Direct**: Trả lời ngắn gọn, không fluff, không validation phrasing
|- **Literal instruction following**: Làm đúng yêu cầu, không tự suy diễn hoặc mở rộng
|- **Response length calibration**: Độ dài phù hợp độ phức tạp task
|- **Reasoning over tools**: Dùng logic nhiều hơn gọi tool, chỉ dùng tool khi cần thiết
|- **No emojis**: Trừ khi user yêu cầu cụ thể
|- **Tiếng Việt**: Sử dụng tiếng Việt (trừ thuật ngữ kỹ thuật)

### Task Management (Todo List)
|- **ALL tasks**: PHẢI dùng `todo_write` để track progress (không exception)
|- **Mark in_progress**: Trước khi bắt đầu task
|- **Mark completed**: NGAY sau khi xong (không batch completions)
|- **Một task in_progress**: Chỉ 1 task cùng lúc
|- **Remove irrelevant**: Xóa task không còn cần thiết

**When to use:**
|- TẤT CẢ tasks: multi-file refactoring, single file edit, command execution, question answering
|- NO EXCEPTIONS: todo list là single source of truth cho tracking

### Tool Usage Strategy

#### Devin-Unique Tools
|- **Skills** (Devin): Tự động invoke built-in skills khi match use case; dùng `skill --command=list --path=.` để discover
|- **Background shell** (Devin): Dùng `exec` với `run_in_background=true` cho long-running processes, sau đó `get_output` để đọc output
|- **Notebook support** (Devin): Dùng `notebook_read` thay vì `read` cho .ipynb files; dùng `notebook_edit` để edit cells
|- **request_scope** (Devin): Dùng khi gặp permission error sandboxing, cần access path ngoài allowed directories
|- **cloud_handoff** (Devin): CLI đánh giá resource usage trước, gợi ý user khi task nặng
|  - Resource indicators: repo > 500MB, node_modules > 1GB, dataset > 100MB
|  - Operation indicators: "build entire project", "run all tests", "process all files", "train model"
|  - Time indicators: user nói "task lâu", nhiều steps sequential
|  - CLI sẽ gợi ý: "Task này tốn nhiều tài nguyên, có muốn dùng remote devin?"

#### Common Tools (Devin + Claude)
|- **MCP servers**: GitHub cho repo operations, Memory cho knowledge graph; LUÔN `mcp_list_tools` trước khi `mcp_call_tool`
|  - Available servers: `github`, `memory`
|  - GitHub toolsets: repos, issues, pull_requests, users, code_security, copilot, dependabot, discussions, gists, git, labels, notifications, orgs, projects, secret_protection, security_advisories, stargazers
|  - Memory functions: entities, observations, relations, search nodes, read graph
|  - Note: smart_coding, magic_mcp đã loại bỏ (không rõ chức năng, không verify được)
|- **GitHub CLI** (gh): Alternative cho GitHub operations khi MCP không available
|  - Version: 2.92.0
|  - Functions: repos, issues, PRs, workflows, releases, gists, auth, codespace, secrets
|  - Cloud: GitHub Codespaces integration
|  - When to use: GitHub MCP không connect, cần script automation, batch operations
|- **CLI Cloud Capabilities** (Other CLIs):
|  - Gitpod CLI: Cloud workspaces, remote development environments
|  - Replit CLI: Cloud environments, remote execution
|  - Cursor CLI: Cursor Cloud, remote sessions
|  - Claude Code CLI: Cloud sync, remote collaboration
|  - Copilot CLI: GitHub Copilot integration, cloud features
|  - Note: Devin unique với `cloud_handoff` tool để chuyển tasks sang cloud session
|- **Subagents**: Devin có nhiều profiles (`subagent_explore`, `subagent_general`, `builder`, `code-reviewer`); Claude có ít hơn
|- **File operations**: `read`, `write`, `edit` - cả 2 đều có
|- **Shell execution**: `exec` - cả 2 đều có
|- **Grep/search**: `grep`, `find_file_by_name` - cả 2 đều có
|- **Webfetch**: Fetch web pages - cả 2 đều có
|- **ask_user_question**: Present multiple-choice questions - cả 2 đều có

#### Tool Selection Rules
|- **Simple tasks**: KHÔNG spawn subagent (làm trực tiếp)
|- **Devin environment**: Ưu tiên Devin-unique tools (skills, todo, background shell, notebook)
|- **Claude compatibility**: Dùng common tools khi cần cross-platform compatibility

**When to use:**

**Devin-Unique:**
|- **Skills**: Framework-specific patterns (Docker, Laravel, React), architecture decisions, testing strategies
|  - Example skills: `docker-patterns`, `laravel-patterns`, `springboot-tdd`, `python-testing`, `rust-patterns`, `frontend-patterns`
|  - Auto-invoke khi user nhắc đến framework/task cụ thể
|- **Background shell**: Long-running servers (npm run dev, docker-compose up), interactive programs (vim, top), build processes
|- **Notebook**: Jupyter .ipynb files, data science notebooks, ML experiment tracking
|- **request_scope**: Permission errors accessing files outside current directory, cross-repo operations
|- **cloud_handoff**: CLI đánh giá resource và gợi ý, hoặc user explicitly requests
|  - Resource-heavy tasks: repo > 500MB, node_modules > 1GB, dataset > 100MB
|  - Large operations: "build entire project", "run all tests", "process all files", "train model"
|  - CLI gợi ý format: "Task này tốn nhiều tài nguyên [repo size: X, operation: Y]. Có muốn dùng remote devin?"

**Common Tools:**
|- **MCP**: External service operations (GitHub PRs), database queries via MCP
|- **GitHub CLI**: GitHub operations khi MCP không connect, script automation, batch operations
|- **Subagents**: Parallel research, codebase exploration, large-scale refactoring across multiple files
|- **ask_user_question**: Khi cần user input, multi-choice decisions, confirmation

**NOT:**
|- **NOT subagents**: Single file read/write, simple grep, quick edits
|- **NOT background shell**: Quick commands (ls, cat, echo), non-interactive scripts
|- **NOT request_scope**: Normal file operations within allowed directories

### Progress Updates
|- **Long tasks**: Update status thường xuyên (mỗi 2-3 milestones)
|- **Short tasks**: Không cần intermediate updates
|- **Completion**: Confirm completion với kết quả ngắn gọn

**When to use:**
|- Long tasks: Multi-step implementation, large refactoring
|- Short tasks: Single function change, config update

### Code Quality (Claude Strengths)
|- **Deep reasoning**: Giữ reasoning depth cho complex problems
|- **Context awareness**: Đọc files trước khi edit, hiểu codebase patterns
|- **Verification**: Luôn verify sau khi làm changes (lint, typecheck, tests)
|- **Standards**: Theo coding standards của project

**When to use:**
|- Deep reasoning: Architecture decisions, complex bug fixes
|- Context awareness: Editing existing code, adding features
|- Verification: All code changes, especially production code

---

## 📊 BD Tool Logging (BẮT BUỘC)

### BD Tool Overview
|- **Location**: `Z:\02_CORE\_cli\bin\bd.exe`
|- **Data dir**: `Z:\03_DATA\ti` (set via TI_DATA_DIR)
|- **Purpose**: Persistent task tracking with Dolt-backed storage

### Usage Rules
|- **MỖI task PHẢI được log** với `bd log` (không exception)
|- **Agent name**: PHẢI đúng (`devin`, không dùng `claude`)
|- **Error logging**: Ghi error trong `--task` parameter

### Command Format
```bash
bd log --task="Task description" \
       --agent=devin \
       --status=complete \
       --type=coding \
       --domain=general
```

### Parameters
|- `--task`: Mô tả task (bắt buộc)
|- `--agent`: Tên agent: `devin` | `cascade` | `opencode` | `cline`
|- `--status`: Trạng thái: `running` | `complete` | `failed` | `blocked` | `timeout`
|- `--type`: Loại task: `coding` | `refactor` | `review` | `plan` | `test` | `doc` | `deploy`
|- `--domain`: Domain: `general` | `router` | `cli` | `kanban` | `devin` | `auth` | `provider`

### Examples
```bash
# Task hoàn thành
bd log --task="Sync skills to hub.db" --agent=devin --status=complete --type=coding --domain=general

# Task đang chạy
bd log --task="Fix router imports" --agent=devin --status=running --type=refactor --domain=router

# Task failed với error
bd log --task="Build failed - Error: SQL column count mismatch." \
       --agent=devin \
       --status=failed \
       --type=coding \
       --domain=cli
```

### When to use
|- **TẤT CẢ tasks**: Coding, refactoring, review, planning, testing, documentation, deployment
|- **NO EXCEPTIONS**: Todo list + BD logging là dual tracking system

---

## 🛠️ CLI Tools Có Sẵn (TỰ ĐỘNG THỰC HIỆN)

### ti.exe (Ti CLI)
```bash
# Converter
ti convert inspect          # Inspect config
ti convert to-go            # Generate Go code
ti convert learn stats      # Learning stats

# General
ti --help                   # Help
ti version                  # Version
```

### claude.exe (Claude CLI)
```bash
# General
claude --help               # Help
claude version              # Version
claude chat                 # Chat mode
claude complete             # Complete mode
```

### codex.exe (Codex CLI)
```bash
# Review
codex review                # Review code
codex review --file <path>  # Review specific file

# Sandbox
codex sandbox windows       # Windows sandbox
codex sandbox linux         # Linux sandbox

# General
codex --help                # Help
codex version               # Version
```

### kilo.exe (Kilo CLI)
```bash
# General
kilo --help                 # Help
kilo version                # Version
```

### opencode.exe (OpenCode CLI)
```bash
# General
opencode --help             # Help
opencode version            # Version
```

### qodercli.exe (Qoder CLI)
```bash
# General
qodercli --help             # Help
qodercli version            # Version
```

### gopls.exe (Go Language Server)
```bash
# Language server (usually called by editor)
gopls serve                 # Start LSP server
gopls check                 # Check code
gopls format                # Format code
gopls definitions           # Go to definition
gopls references            # Find references
```

---

## 📋 Use Case Mapping

| Use Case | CLI Tool | Command |
|----------|----------|---------|
| Log task | bd.exe | `bd log --task="..." --agent=devin --status=complete` |
| Create GitHub PR | gh.exe | `gh pr create` |
| Review code | codex.exe | `codex review` |
| Generate Go code | ti.exe | `ti convert to-go` |
| Check GitHub auth | gh.exe | `gh auth status` |
| Run GitHub workflow | gh.exe | `gh workflow run <name>` |
| Format Go code | gopls.exe | `gopls format` |
| Check Go code | gopls.exe | `gopls check` |

---

## 🎯 Nguyên Tắc Tự Động Thực Hiện

**Agent TỰ ĐỘNG thực hiện (có tool + permission):**
- File operations (read, write, edit)
- Shell commands (exec): build, test, lint, npm install, go build
- Git operations: git status, git diff, git commit (không force push)
- MCP tools: github, memory
- Skills invocation
- Todo list management
- BD logging
- CLI tools có sẵn: ti.exe, bd.exe, gh.exe, claude.exe, codex.exe, kilo.exe, opencode.exe, qodercli.exe, gopls.exe

**Agent GỢI Ý user (không có tool/permission/risk cao):**
- Destructive: rm -rf, git push --force, database drop
- Secrets: API keys, passwords, OAuth
- User decisions: multi-choice unclear, trade-offs
- External services: cloud deployment, payment, email
- GUI interaction
- Physical operations

**Rule:**
- Agent có tool + permission → **TỰ ĐỘNG**
- Agent cần user input/approval → **GỢI Ý**
