# Notion Sync Setup Guide

Hướng dẫn cấu hình Notion Sync cho TiBrain.

## Prerequisites

1. **Notion Account** với quyền tạo Integration
2. **Obsidian Vault** (hoặc thư mục markdown) để sync
3. Python 3.x đã cài đặt

## Step 1: Tạo Notion Integration

1. Vào [Notion Integrations](https://www.notion.so/my-integrations)
2. Click "New integration"
3. Đặt tên: "TiBrain Sync"
4. Chọn workspace của bạn
5. Copy **Internal Integration Token**

## Step 2: Tạo Notion Database

1. Tạo một page mới trong Notion
2. Thêm một Database inline (Table view)
3. Các properties cần có:
   - **Title** (default)
   - **Tags** (Multi-select)
   - **Scopes** (Multi-select)
   - **Category** (Select)
   - **Tier** (Select: T1, T2, T3)
   - **Priority** (Select: P0, P1, P2)
   - **Last Updated** (Date)
   - **Version** (Text)
   - **Source** (Select)
   - **Sync Status** (Select)
4. Copy **Database ID** từ URL:
   - URL: `https://www.notion.so/workspace/1234567890abcdef?v=...`
   - Database ID: `1234567890abcdef`

## Step 3: Share Database với Integration

1. Mở database trong Notion
2. Click "Share" (góc trên bên phải)
3. Click "Add people"
4. Tìm và chọn integration "TiBrain Sync"
5. Set quyền "Can edit"

## Step 4: Cấu hình TiBrain

Chỉnh sửa file `.env` trong `apps/tibrain/`:

```env
# Notion Configuration
NOTION_TOKEN=secret_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
NOTION_DATABASE_ID=1234567890abcdef

# Sync Configuration
VAULT_PATH=Z:/path/to/your/obsidian/vault
SYNC_PATTERN=*.md
```

## Step 5: Chạy Sync

### Cách 1: API Endpoint

```bash
# Start TiBrain
cd z:/10_WORKPLACE/Ti/apps/tibrain
./tibrain.exe

# Trigger sync (từ terminal khác)
curl -X POST http://localhost:1810/v1/tibrain/sync/notion
```

### Cách 2: Python Script trực tiếp

```bash
cd z:/10_WORKPLACE/Ti/apps/tibrain
python sync_notion.py
```

### Cách 3: Python Script với tham số

```bash
cd z:/10_WORKPLACE/Ti/apps/tibrain
python notion_sync.py \
  --vault "Z:/path/to/vault" \
  --notion-token "secret_xxx" \
  --database-id "1234567890abcdef"
```

## Step 6: Schedule Sync (Tùy chọn)

### Windows Task Scheduler

1. Mở Task Scheduler
2. Create Basic Task
3. Name: "TiBrain Notion Sync"
4. Trigger: Daily (hoặc hourly)
5. Action: Start a program
   - Program: `python`
   - Arguments: `sync_notion.py`
   - Start in: `Z:\10_WORKPLACE\Ti\apps\tibrain`

### Linux/macOS Cron

```bash
# Mỗi giờ
0 * * * * cd /z/10_WORKPLACE/Ti/apps/tibrain && python sync_notion.py
```

## Troubleshooting

### Lỗi "NOTION_TOKEN not configured"

- Kiểm tra file `.env` đã có token chưa
- Đảm bảo token đúng format: `secret_xxxxxxxx`

### Lỗi "Database not found"

- Kiểm tra Database ID đúng chưa
- Đảm bảo đã share database với integration

### Lỗi Python module not found

```bash
pip install pyyaml requests
```

### Lỗi "Access denied" (Windows ACL)

Nếu gặp lỗi ACL khi chạy TiBrain:

```powershell
# Run as Admin
icacls "Z:\10_WORKPLACE\Ti\apps\tibrain" /reset /T
```

## Sync Behavior

- **Incremental sync**: Kiểm tra title để update thay vì create duplicate
- **Tags validation**: Chỉ sync tags hợp lệ theo taxonomy
- **Content conversion**: Markdown → Notion blocks (basic support)

## API Reference

### POST /v1/tibrain/sync/notion

Trigger sync từ TiBrain API.

**Response:**

```json
{
  "success": true,
  "output": "Sync log...",
  "error": null
}
```
