# NIM Integration Handoff - 2026-07-09

## Trạng thái
- ✅ NIM server đang chạy: http://127.0.0.1:3456/v1 (OpenAI-compatible)
- ✅ NVIDIA NIM đã thêm vào LOCAL_PROVIDERS
- ❌ OmniRoute chưa chạy (cần start)

## Cấu hình Claude Code CLI

Claude Code CLI dùng Anthropic API. Để dùng NIM qua OmniRoute:

### 1. Cấu hình OmniRoute Provider

Tạo provider connection qua API hoặc Dashboard:
- **Provider**: `openai-compatible` (hoặc tạo custom `anthropic-compatible-cc-nim`)
- **Base URL**: `http://127.0.0.1:3456/v1`
- **API Key**: để trống (NIM local không cần key)

### 2. Claude Code CLI Environment

```bash
# Windows PowerShell
$env:ANTHROPIC_BASE_URL="http://localhost:20128/v1"  
$env:ANTHROPIC_API_KEY="sk-no-key-required"

# Hoặc dùng provider-specific endpoint
$env:ANTHROPIC_BASE_URL="http://localhost:20128/api/v1/providers/nvidia-nim/chat/completions"
```

### 3. Claude Code Extension (VS Code)

File: `C:\Users\MIN\AppData\Roaming\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "omniroute": {
      "command": "npx",
      "args": ["mcp-remote", "http://localhost:20128/mcp"]
    }
  }
}
```

## Cấu hình OmniRoute

### Start Router
```bash
cd Z:\01_PROJECTS\apps\Tirouter\OmniRoute-main
npm run dev
# Hoặc build rồi chạy:
npm run build
node bin/omniroute.mjs
```

### Test Connection
```bash
curl http://localhost:20128/v1/models
curl http://localhost:20128/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-no-key-required" \
  -d '{"model":"nvlm","messages":[{"role":"user","content":"Hello"}]}'
```

## Files Modified
- `src/shared/constants/providers/local.ts` - Thêm nvidia-nim provider
- `src/shared/constants/providers.ts` - Thêm nvidia-nim vào SELF_HOSTED_CHAT_PROVIDER_IDS

## Lưu ý
- OmniRoute cần quyền truy cập localhost (mặc định đã bật)
- Claude Code CLI/extension không hỗ trợ trực tiếp OpenAI endpoint, cần qua OmniRoute translation layer
- NIM model names: llama-3.1-nemotron-70b-instruct, nemotron-4-340b-instruct, codestral-22b