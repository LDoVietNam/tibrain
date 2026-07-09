# Ti Console

`Ti Console` là TUI/control surface đọc trạng thái của nền tảng Ti, được thiết kế theo mô hình vận hành của AI DevKit nhưng giữ TiBrain làm control plane và source of truth.

## Phạm vi MVP

Phiên bản đầu tiên chỉ đọc dữ liệu và không thực hiện thao tác phá hủy hoặc thay đổi cấu hình. Console hiện hỗ trợ:

- Kiểm tra sức khỏe TiBrain, Beads, 1MCP và các Router candidate.
- Làm mới tự động theo chu kỳ.
- Hiển thị trạng thái `healthy`, `degraded`, `down`, `unknown`.
- Hiển thị latency, health endpoint thực tế và thông tin phản hồi ngắn.
- Xuất một snapshot JSON để tích hợp với script hoặc automation.
- Không coi một Router là active chỉ vì port xuất hiện trong cấu hình.

## Kiến trúc

```text
Ti Console renderer
        │
        ▼
Backend interface
        │
        ▼
Ti HTTP backend adapter
        │
        ├── TiBrain
        ├── Beads
        ├── 1MCP
        └── Router candidates
```

UI không đọc trực tiếp database, file secret hoặc runtime state. Mọi trạng thái đều đến từ live HTTP probe. Ranh giới `Backend` cho phép thay renderer hiện tại bằng Agent Console của AI DevKit trong giai đoạn tiếp theo mà không đưa business logic vào TUI.

## Build

```powershell
cd Z:\01_PROJECTS\apps\tibrain
powershell -ExecutionPolicy Bypass -File .\scripts\build-ti-console.ps1
```

Binary đầu ra:

```text
Z:\01_PROJECTS\apps\tibrain\bin\ti-console.exe
```

## Chạy

```powershell
.\ti-console.cmd
```

Hoặc:

```powershell
.\bin\ti-console.exe --refresh 3s
```

Các lệnh trong TUI:

```text
1  Dashboard
2  Services
3  Help
r  Làm mới ngay
q  Thoát
```

Nhập lệnh rồi nhấn `Enter`. Dashboard vẫn tự động làm mới trong lúc chờ input.

## Snapshot JSON

```powershell
.\bin\ti-console.exe --once
```

## Biến môi trường

```text
TIBRAIN_URL      mặc định http://127.0.0.1:1810
TI_BEADS_URL     mặc định http://127.0.0.1:1811
TI_MCP_URL       mặc định http://127.0.0.1:8080
TI_ROUTER_URLS   danh sách URL Router candidate, phân tách bằng dấu phẩy
```

Ví dụ:

```powershell
$env:TI_ROUTER_URLS = "http://127.0.0.1:1809,http://127.0.0.1:1817"
.\ti-console.cmd
```

## Ràng buộc an toàn

- Console không đánh dấu service active từ static config.
- Không có action start, stop, kill, delete hoặc sửa config trong MVP.
- Không đọc hoặc hiển thị secret.
- Mọi action ghi trong các giai đoạn sau phải đi qua Ti Control API, policy check, idempotency key và audit event.

## Giai đoạn tiếp theo

1. Thêm tab Agents và Tasks từ TiBrain/Beads API.
2. Thêm SSE event stream để giảm polling.
3. Tạo adapter TypeScript cho Agent Console của AI DevKit.
4. Chỉ bật các action ghi sau khi Ti Control API có policy, audit và rollback contract.
