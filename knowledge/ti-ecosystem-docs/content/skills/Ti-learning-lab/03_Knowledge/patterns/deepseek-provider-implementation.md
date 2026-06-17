---
tags: ["tibrain", "provider", "skill", "documentation", "provider-deepseek"]
scopes: ["tibrain"]
last_updated: 2026-05-22
---
# DeepSeek Provider Implementation

## Overview

Triển khai DeepSeek provider cho Ti Router, hỗ trợ các model DeepSeek thông qua OpenAI-compatible API.

## Problem Statement

Router chưa có provider cho DeepSeek API, một trong những LLM providers phổ biến với giá rẻ và hiệu suất tốt.

## Solution

Tạo package `layers/provider/deepseek/` với đầy đủ implementation:

### Files Created

1. **provider.go** - Core provider implementation
   - `DeepSeekProvider` struct với BaseProvider embedding
   - Các methods: Name(), Status(), DefaultModel(), Models(), IsHealthy(), Init()
   - Chat() và ChatStream() methods (delegated to BaseProvider)
   - BaseURL: https://api.deepseek.com
   - Models: deepseek-chat, deepseek-coder

2. **request.go** - Request conversion
   - `convertAnthropicToDeepSeekRequest()` - Convert Anthropic format sang DeepSeek format
   - `handleThinkingTokens()` - Xử lý thinking tokens trong request
   - Mapping: messages, tools, temperature, max_tokens, top_p, stream

3. **stream.go** - Stream response handling
   - `convertDeepSeekToAnthropicStream()` - Convert DeepSeek stream sang Anthropic format
   - `handleThinkingTokensInStream()` - Xử lý thinking tokens trong stream
   - SSE chunk parsing và forwarding

### Files Modified

1. **layers/provider/config.go**
   - Thêm deepseek vào defaultConfigs
   - Config: BaseURL, APIKeyEnv (DEEPSEE_API_KEY), Models, Format (openai), TimeoutSec, Priority, Weight, CostPer1K

2. **layers/provider/plugin_registry.go**
   - Import deepseek package
   - Register factory cho deepseek provider
   - Factory function tạo DeepSeekProvider với API key từ config

## Architecture Pattern

Sử dụng **BaseProvider Pattern** tương tự như NVIDIA NIM provider:

```go
type DeepSeekProvider struct {
    provider.BaseProvider
    apiKey  string
    baseURL string
    client  *http.Client
    models  []string
}
```

**Lợi ích:**
- Code reuse thông qua BaseProvider
- Consistent interface với các providers khác
- Built-in capabilities management
- Version tracking

## Request Conversion

DeepSeek sử dụng OpenAI-compatible format, nên conversion tương đối đơn giản:

```go
type DeepSeekRequest struct {
    Model       string         `json:"model"`
    Messages    []Message      `json:"messages"`
    Tools       []Tool         `json:"tools,omitempty"`
    Temperature float64        `json:"temperature,omitempty"`
    MaxTokens   int            `json:"max_tokens,omitempty"`
    TopP        float64        `json:"top_p,omitempty"`
    Stream      bool           `json:"stream,omitempty"`
}
```

## Stream Handling

Sử dụng SSE (Server-Sent Events) cho streaming:

1. Parse SSE chunks từ DeepSeek API
2. Convert sang Anthropic format
3. Forward tới client thông qua onChunk callback
4. Xử lý thinking tokens nếu có

## Configuration

```yaml
deepseek:
  Name: "deepseek"
  BaseURL: "https://api.deepseek.com"
  APIKeyEnv: "DEEPSEE_API_KEY"
  Models: ["deepseek-chat", "deepseek-coder"]
  Format: "openai"
  TimeoutSec: 120
  Priority: 13
  Weight: 3
  CostPer1K: 0.001
```

## Testing

Do pre-existing import cycle trong layers/provider/cookie, không thể build toàn bộ project. Tuy nhiên:

- Individual layer builds pass (layers/provider/deepseek/)
- Implementation theo pattern đã test với NVIDIA NIM provider
- Code review đảm bảo consistency với existing providers

## Known Issues

**Pre-existing Import Cycle**: layers/provider/cookie → layers/provider
- Đây là issue riêng (TR-000)
- Không liên quan đến TR-014
- Cần fix import cycle trước khi build toàn bộ project

## Lessons Learned

1. **BaseProvider Pattern** rất hiệu quả cho provider implementation
2. OpenAI-compatible format đơn giản hóa conversion logic
3. Factory registration trong plugin_registry.go cần consistency với defaultConfigs
4. Priority và Weight trong config cần cân nhắc dựa trên cost và performance

## Next Steps

- Fix import cycle (TR-000)
- Test DeepSeek provider với API key thực tế
- Add metrics và monitoring cho DeepSeek requests
- Consider thêm DeepSeek reasoning models nếu có

## References

- DeepSeek API: https://api.deepseek.com
- NVIDIA NIM provider implementation (reference pattern)
- OpenAI API specification (for format compatibility)
