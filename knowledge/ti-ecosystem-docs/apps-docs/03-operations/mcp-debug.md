# Chrome DevTools MCP - TiRoute Debug Guide

## ✅ Setup Complete

Chrome DevTools MCP đã được cấu hình cho TiRoute project.

## 📁 Files Created

1. **`.mcp.json`** - MCP configuration file
2. **`debug-mcp.bat`** - Quick start script

## 🚀 How to Use

### Option 1: Using Qwen Code (Recommended)

Qwen Code tự động đọc `.mcp.json` và kết nối với Chrome DevTools MCP.

1. Mở terminal trong Qwen Code
2. Chạy: `TiRoute-Dev.bat` để start TiRoute
3. MCP server sẽ tự động kết nối khi cần

### Option 2: Using Claude Code

```bash
cd Z:\01_PROJECTS\Spec-Router
claude
```

Claude Code sẽ tự động đọc `.mcp.json` và cung cấp Chrome DevTools tools.

### Option 3: Manual Start

```bash
cd Z:\01_PROJECTS\Spec-Router
debug-mcp.bat
```

## 🔧 Available MCP Tools

Khi MCP server chạy, bạn có thể dùng các tools sau:

### Browser Control
- `browser_navigate` - Mở trang TiRoute
- `browser_screenshot` - Chụp ảnh màn hình
- `browser_click` - Click vào element
- `browser_type` - Gõ text vào input
- `browser_evaluate` - Chạy JavaScript trong console

### Network Debug
- `network_get_requests` - Xem tất cả requests
- `network_get_response` - Xem response chi tiết

### Console Debug
- `console_get_messages` - Đọc console logs
- `console_evaluate` - Chạy lệnh trong console

### Performance
- `performance_start_trace` - Record performance
- `lighthouse_audit` - Chạy Lighthouse audit

## 🎯 Debug TiRoute Workflow

### 1. Start TiRoute
```bash
cd Z:\01_PROJECTS\Spec-Router
TiRoute-Dev.bat
```

### 2. Open Browser via MCP
```
Tool: browser_navigate
URL: http://localhost:3000
```

### 3. Take Screenshot
```
Tool: browser_screenshot
```

### 4. Check Console for Errors
```
Tool: console_get_messages
```

### 5. Inspect Network Requests
```
Tool: network_get_requests
```

## 🔍 Common Debug Scenarios

### Fix UI Issues
1. Navigate to page
2. Take screenshot
3. Check console errors
4. Inspect element styles

### Fix API Issues
1. Make API call
2. Check network tab
3. Inspect request/response
4. Check server logs

### Performance Issues
1. Start performance trace
2. Interact with UI
3. Stop trace and analyze
4. Run Lighthouse audit

## 📝 Notes

- MCP server cần Chrome/Edge cài sẵn
- TiRoute phải đang chạy trước khi debug
- MCP tools chỉ available trong AI agents (Qwen Code, Claude Code)

## 🐛 Troubleshooting

**MCP không kết nối được?**
- Kiểm tra Chrome/Edge đã cài chưa
- Restart MCP server

**TiRoute không chạy?**
- Check `dev-output.log`
- Đảm bảo port 3000 không bị占用

**Lỗi TypeScript?**
- Chạy `npm run build` trong MCP folder
- Hoặc dùng npx trực tiếp
