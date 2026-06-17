# Phân tích Ưu điểm: Port Devin sang Go cho Ti CLI

> Ngày: 2026-04-29 | Agent: claude
> Mục đích: Trả lời "có thực sự cần thiết port devin cho CLI không ?"

## Trạng thái Hiện tại Ti CLI

```
internal/bridge/
├── devin_wrapper.py      # MOCK (không gọi thật devin_cli)
├── grpc_server.py        # gRPC server Python
├── translator.py         # Context translation
├── lifecycle.py          # Process management
└── requirements.txt      # 6 dependencies Python

internal/agents/devin/
└── (EMPTY - chờ implement)
```

**Python bridge hiện tại là stub/mock** - `_execute_with_devin()` trả về mock data, không gọi API thật.

## Giá trị của Port sang Go Native

### 1. Thay thế Python Bridge (Khắc phục điểm yếu hiện tại)

| Vấn đề Python Bridge | Giải pháp Go Native |
|------------------------|---------------------|
| Cần Python runtime | Single Go binary |
| 6 dependencies (grpcio, psutil, pyyaml...) | 0 external dependency |
| gRPC serialization overhead | Direct HTTP client |
| Mock/stub code | Real API implementation |
| Process management phức tạp | Native goroutines |

### 2. Unified Ti Ecosystem Experience

```
ti ask "build a React app"     → Ti Brain route → Devin (nếu task phù hợp)
ti devin status               → Kiểm tra session status
ti devin resume <id>          → Resume session
ti brain log                  → Xem tất cả sessions từ mọi providers
```

**Giá trị**: Một CLI thống nhất cho tất cả AI agents, không cần chuyển đổi tool.

### 3. Ti Brain Integration

- Devin sessions được log vào `beads.md`
- Ti Brain học được: "Task loại X nên route cho Devin"
- Context từ sessions trước được reuse

### 4. Custom Workflows (Không có ở Official CLI)

```bash
# Workflow: Devin viết code → Claude review → Deploy
ti workflow run code-review-deploy \
  --devin "Implement OAuth2" \
  --claude "Review security" \
  --deploy
```

### 5. Offline/Enterprise Deployment

- Go binary: copy 1 file là chạy
- Python bridge: cần cài Python, pip install, venv

## Chức năng NÊN Port (Theo thứ tự giá trị)

### Phase 1: Core (High Value)

1. **Session Manager**
   - Create session với prompt
   - Poll status cho đến completion
   - Retrieve messages/output
   - List/Resume/Cancel sessions

2. **Context Translator** (từ `translator.py`)
   - TiContext → DevinContext
   - Task type routing (code_edit, analysis, debugging)
   - Tool mapping (file_edit → edit, shell_execute → run_shell_command)

3. **Config Integration**
   - Multi-profile (dev, staging, prod)
   - Env var precedence
   - Integration với Ti CLI config layer

### Phase 2: Advanced (Medium Value)

4. **Knowledge Management**
   - Upload/download knowledge cho reuse

5. **Workspace Integration**
   - `/add-dir` tương đương
   - Sync workspace với Ti project

6. **Playbook/Template**
   - Lưu prompt templates
   - Reusable workflows

### Không cần Port (Low Value / Duplicate)

| Feature | Lý do không port |
|---------|-----------------|
| REPL interactive | Official CLI đã có, dùng `devin` binary |
| Model switching (`/model`) | Official CLI đã có |
| `/accept-edits` mode | Official CLI đã có |
| Real-time stream output | Devin API không hỗ trợ SSE |

## Điểm then chốt: API Access

| Devin Plan | API Access | Port Value |
|-----------|-----------|-----------|
| Core ($20/th) | ❌ Không có | **Thấp** - Chỉ có thể wrap official CLI |
| Team ($500/th) | ✅ Có API | **Cao** - Full programmatic control |
| Enterprise | ✅ Có API | **Cao** - Custom integrations |

## Đề xuất cho User

Nếu bạn có **Core plan** ($20/tháng):
- Không port REST API client (vô dụng - không có API access)
- Thay vào đó: **Wrap official `devin` CLI binary** trong Ti CLI
- Hoặc: Chỉ giữ `DevinProvider` trong Ti Router cho future

Nếu bạn có **Team/Enterprise**:
- Port session management + context translator sang Go
- Bỏ Python bridge
- Giữ integration với Ti Brain

## Quyết định

**Port giá trị nhất**: Session lifecycle (create → poll → result) + Context translation.
**Không port**: REPL, model switching (đã có ở official CLI).
**Yêu cầu tiên quyết**: Phải có Devin API access (Team/Enterprise plan).
