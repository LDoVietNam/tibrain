# Ti Context Dashboard

Local web dashboard để xem tất cả Ti context files.

## 🚀 Start Dashboard

```bash
cd Z:\10_WORKPLACE\Ti\Ti-learning-lab\03_Knowledge\Router\context-dashboard
node server.js
```

Dashboard sẽ chạy tại: **http://localhost:3000**

## 🌐 Mở Dashboard

Mở browser và truy cập: http://localhost:3000

## 📱 Features

- **Sidebar**: Danh sách 8 context files
- **Search**: Tìm kiếm theo tên/description
- **Expandable Sections**: Click để mở/đóng từng section
- **Stats Dashboard**: Số keys, sections, file size
- **Dark Theme**: Giao diện tối dễ nhìn

## 📂 Context Files

1. Ti Router - AI routing gateway
2. Ti CLI - Microkernel + Plugin architecture
3. Ti Automation - Plugin registry system
4. Ti Dashboard - Web dashboard
5. Donut Browser - Anti-detect browser
6. MCP Hub - Central MCP servers
7. Providers - API providers collection
8. Ti TUI - Terminal UI component

## 🛑 Stop Dashboard

Nhấn **Ctrl+C** trong terminal đang chạy server.

## 🔧 Troubleshooting

Nếu dashboard không load:
1. Kiểm tra server đang chạy: http://localhost:3000/api/contexts
2. Kiểm tra browser console (F12) xem có lỗi JavaScript không
3. Nếu port 3000 bị占用, sửa PORT trong `server.js`

## 📝 File Structure

```
context-dashboard/
├── index.html          # Frontend UI
├── server.js           # Node.js HTTP server
├── server.go           # Go HTTP server (backup)
└── server.py           # Python HTTP server (backup)
```
