# Quy tắc đặt file binary cho MCP

## Mục tiêu
Đảm bảo tất cả các file thực thi Go (mcpproxy.exe, mcp‑workspace.exe, social‑mcp.exe, tibrain.exe, extension.exe, …) được đặt ở một vị trí rõ ràng, nhất quán và dễ tham chiếu từ các script, file cấu hình và tài liệu.

## Vị trí chuẩn hoá
```
Z:\01_PROJECTS\apps\tibrain\mcp\bin\<tên‑binary>.exe
```
Ví dụ:
- `Z:\01_PROJECTS\apps\tibrain\mcp\bin\mcpproxy.exe`
- `Z:\01_PROJECTS\apps\tibrain\mcp\bin\mcp‑workspace.exe`
- `Z:\01_PROJECTS\apps\tibrain\mcp\bin\social‑mcp.exe`
- `Z:\01_PROJECTS\apps\tibrain\mcp\bin\tibrain.exe`
- `Z:\01_PROJECTS\apps\tibrain\mcp\bin\extension.exe`

### Lý do chọn vị trí này
- Các script hiện tại (`Bootstrap.ps1`, `Start.ps1`, `Update‑GptWebSetup.ps1`, …) đều tạo đường dẫn tới `<repo>\bin` khi biên dịch (`go build -o bin\…`).
- Giữ nguyên cấu trúc repository: thư mục `bin` nằm ngay bên trong thư mục dự án (`apps/tibrain/mcp`), không làm thay đổi cấu trúc gốc và không影響 tới các submodule khác (tibrain, browser‑runtime‑extension, …).
- Dễ dàng dọn dẹp: chỉ cần xóa toàn bộ thư mục `bin` và rebuild lại – không ảnh hưởng tới file cấu hình, source code hoặc dữ liệu runtime.
- Tránh nhầm lẫn với các thư mục tạm thời như `.runtime\bin`, `.cache\go-mod\…\bin` hoặc `node_modules\…\bin` dùng riêng cho các bản build tạm thời hoặc dependencies.

## Cách tham chiếu từ các script và cấu hình

### Trong PowerShell (ví dụ)
```powershell
$binDir = Join-Path $PSScriptRoot '..\bin'
$binary = Join-Path $binDir 'mcp-workspace.exe'
```

### Trong file cấu hình JSON (mcp_config.template.json)
```json
{
  "mcpServers": [
    {
      "name": "workspace-filesystem",
      "command": "Z:/01_PROJECTS/apps/tibrain/mcp/bin/mcp-workspace.exe",
      "args": [
        "--config=deploy/ti-local/policies/workspace-policy.json",
        "--addr=127.0.0.1:1841"
      ],
      "enabled": true,
      "auto_approve_tool_changes": true
    }
  ]
}
```

## Tùy chọn: vị trí chung cho nhiều dự án
Nếu muốn chia sẻ binary giữa nhiều dự án con (mcp, tibrain, browser‑extension, …), có thể dùng:
```
Z:\01_PROJECTS\apps\bin\<tên‑binary>.exe
```
Các script của từng dự án sẽ cần điều chỉnh đường dẫn để trỏ tới `%REPO_ROOT%\apps\bin`.

## Cách cập nhật khi thay đổi vị trí
1. Cập nhật tất cả các nơi tham chiếu (script, file JSON, tài liệu) tới đường dẫn mới.
2. Chạy lại `Bootstrap.ps1` để tạo mới file cấu hình nếu cần.
3. Build lại các binary và đặt chúng vào vị trí mới theo quy tắc trên.

## Lưu ý
- Không nên đặt binary trực tiếp trong thư mục `.runtime\` vì đây là nơi lưu trữ dữ liệu runtime (logs, data, snapshots, …) và có thể bị xóa trong các thao tác dọn dẹp.
- Đảm bảo rằng người dùng thực thi có quyền truy cập và thực thi file trong thư mục `bin` (thường không yêu cầu quyền administrator trừ khi thư mục nằm trong vị trí hệ thống bảo vệ).

--- 
*Đoc quy tắc này để các agent (tibrain, GPT Web, Claude, …) biết rõ nơi đặt và cách truy cập các binary cần thiết để hoạt động với MCP pipeline.*