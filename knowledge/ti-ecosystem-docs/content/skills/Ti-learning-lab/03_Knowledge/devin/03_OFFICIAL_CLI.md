# Phân tích: Devin Official CLI

> Ngày: 2026-04-29 | Agent: claude

## Tóm tắt

Devin đã có CLI chính thức (`cli.devin.ai`). Python repo `revanthpobala/devin-cli` chỉ là unofficial API wrapper.

## Official CLI Features

```bash
devin                          # Interactive REPL
devin -- "prompt"              # REPL với prompt
devin -p "prompt"              # Single-turn stdout rồi exit
devin -c                       # Continue session gần nhất
```

### Modes
- `/normal`, `/accept-edits`, `/plan`, `/bypass` (`/yolo`)

### Models
- `/model swe`, `/model opus`, `/model sonnet`, `/model gpt`

### Session & Workspace
- `/ls`, `/resume`, `/continue`, `/rm-session`
- `/workspace`, `/add-dir <path>`

## So sánh

| Feature | Official CLI | Python Unofficial |
|---------|-------------|-------------------|
| REPL | ✅ | ❌ |
| Modes | ✅ 4 modes | ❌ |
| Model switch | ✅ swe/opus/sonnet/gpt | ❌ |
| Workspace | ✅ Native | ❌ |
| Plan required | Core $20/th | Teams+ $500/th |

## Quyết định

KHÔNG cần port Python CLI sang Go. Devin đã có official CLI.
