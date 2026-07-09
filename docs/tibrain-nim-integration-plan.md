# Plan: Claude Code + NVIDIA NIM Integration

## Tình huống hiện tại
- NIM server đang chạy: `http://127.0.0.1:3456` (OpenAI-compatible API)
- Claude Code extension đang hoạt động
- Claude Desktop chỉ hỗ trợ Anthropic API, không thể trực tiếp dùng NIM

## Giải pháp: OmniRoute Router
OmniRoute đã có sẵn trong hệ thống (`z:\02_CORE\_cli\.config\skills\omniroute-router`)

### Kiến trúc đề xuất
```
Claude Code → OmniRoute (localhost:router-port) → NVIDIA NIM (127.0.0.1:3456)
```

## Các bước thực hiện

### Bước 1: Cấu hình OmniRoute Router
```bash
# Tạo config cho NIM provider
claw config set providers.nim-nvidia \
  --type openai \
  --base-url http://127.0.0.1:3456/v1 \
  --api-key dummy-key
```

### Bước 2: Tạo routing rule
```yaml
# Trong OmniRoute config
routing_rules:
  - name: "claude-to-nim"
    match:
      model_pattern: "nvidia_nim/*"  # hoặc model_alias
    route_to: nim-nvidia
```

### Bước 3: Claude Code sử dụng qua router
Trong `~/.config/kilo/kilo.json` hoặc Claude config:
```json
{
  "providers": {
    "default": {
      "type": "openai",
      "base_url": "http://127.0.0.1:OMNIROUTE_PORT/v1",
      "api_key": "router-key"
    }
  }
}
```

### Bước 4: Model selection
Dùng model ID từ NIM:
```
anthropic/nvidia_nim/nvidia/llama-3.1-nemotron-70b-instruct
anthropic/nvidia_nim/nvidia/nemotron-4-340b-instruct
anthropic/nvidia_nim/qwen/qwen3.5-122b-a10b
```

## Model đề xuất dùng
- **Coding**: `anthropic/nvidia_nim/nvidia/llama-3.1-nemotron-70b-instruct`, `anthropic/nvidia_nim/mistralai/codestral-22b-instruct-v0.1`
- **Reasoning**: `anthropic/nvidia_nim/nvidia/nemotron-4-340b-instruct`, `anthropic/nvidia_nim/qwen/qwen3.5-122b-a10b`
- **Nhanh**: `anthropic/nvidia_nim/meta/llama-3.2-1b-instruct`

## Lưu ý
- NIM không phải là Claude model - chỉ là LLM local
- Cần có GPU để chạy NIM với model lớn
- OmniRoute giúp chuyển đổi format API giữa Claude và OpenAI