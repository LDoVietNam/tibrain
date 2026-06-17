# TiBrain MCP Server cho Windsurf

## Cấu Hình trong Windsurf IDE

### Bước 1: Cấu hình MCP Server

Mở **Windsurf Settings** → **Windsurf Settings** → **Cascade** → **Model Context Protocol Servers**

Thêm server mới:

```json
{
  "mcpServers": {
    "tibrain": {
      "command": "python",
      "args": ["Z:\\10_WORKPLACE\\Ti\\tibrain\\mcp-server.py"]
    }
  }
}
```

Hoặc nếu dùng file config (thường ở `~/.codeium/windsurf/mcp_config.json` hoặc trong `.windsurf/mcp.json` của project):

```json
{
  "servers": [
    {
      "name": "tibrain",
      "transport": {
        "type": "stdio",
        "command": "python",
        "args": ["Z:\\10_WORKPLACE\\Ti\\tibrain\\mcp-server.py"]
      }
    }
  ]
}
```

### Bước 2: Khởi động TiBrain Hub (nếu chưa chạy)

```powershell
cd Z:\10_WORKPLACE\Ti\router\tibrain-server
.\tibrain-server.exe --port 1809
```

### Bước 3: Kiểm tra trong Windsurf

Sau khi add MCP server, Windsurf sẽ tự động load tools. Kiểm tra bằng cách mở Cascade chat và gõ:

```
Dùng tool tibrain_status để kiểm tra
```

## Tools có sẵn

| Tool | Mô tả |
|------|-------|
| `tibrain_status` | Xem trạng thái hub và CLI đang active |
| `tibrain_list_clis` | Liệt kê tất cả CLI đã đăng ký |
| `tibrain_recall_handoff` | Lấy context handoff giữa 2 CLI |
| `tibrain_search_skills` | Tìm kiếm skill/pattern trong registry |
| `tibrain_execute_tool` | Execute tool qua TiBrain |
| `tibrain_log_usage` | Ghi log việc dùng tool |

## Ví dụ sử dụng

### Kiểm tra trạng thái
```
Kiểm tra TiBrain đang chạy không?
```
→ Windsurf tự động gọi `tibrain_status`

### Lấy handoff từ Claude sang Windsurf
```
Lấy context handoff từ claude sang windsurf
```
→ Gọi `tibrain_recall_handoff` với `from_cli=claude`, `to_cli=windsurf`

### Tìm skill về Python
```
Tìm skill best practices cho Python
```
→ Gọi `tibrain_search_skills` với `query=python`

## Troubleshooting

### "Connection refused"
→ Chưa khởi động `tibrain-server.exe --port 1809`

### "Module not found: requests"
```powershell
pip install requests
```

### Windsurf không nhận tools
→ Restart Windsurf IDE sau khi add MCP server

## Files

- **MCP Server**: `Z:\10_WORKPLACE\Ti\tibrain\mcp-server.py`
- **TiBrain Hub**: `Z:\10_WORKPLACE\Ti\router\tibrain-server\`
