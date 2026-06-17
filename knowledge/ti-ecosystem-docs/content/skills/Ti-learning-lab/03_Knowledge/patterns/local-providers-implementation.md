---
tags: ["tibrain", "provider", "documentation", "skill", "go"]
scopes: ["cli", "tibrain"]
last_updated: 2026-05-22
---
# Local Providers Implementation

## Overview

Triển khai local providers (LM Studio, llama.cpp, Ollama) cho Ti Router, cho phép sử dụng các LLM models chạy local trên máy của user.

## Problem Statement

Router chưa có support cho local LLM providers, là lựa chọn cost-effective cho development và privacy-sensitive applications.

## Solution

Tạo 3 local provider packages với OpenAI-compatible API:

### 1. LM Studio Provider

**Location**: `layers/provider/lmstudio/`

**Default BaseURL**: `http://localhost:1234/v1`

**Files Created**:
- `provider.go` - Core provider implementation
- `request.go` - Request conversion (OpenAI-compatible)
- `stream.go` - Stream response handling (SSE)

**Characteristics**:
- No API key required (local)
- Timeout: 300s (local models may be slower)
- Priority: 20 (lower than cloud providers)
- Weight: 1 (fallback priority)
- Cost: 0.0 (free)

### 2. llama.cpp Provider

**Location**: `layers/provider/llamacpp/`

**Default BaseURL**: `http://localhost:8080`

**Files Created**:
- `provider.go` - Core provider implementation
- `request.go` - Request conversion (OpenAI-compatible)
- `stream.go` - Stream response handling (SSE)

**Characteristics**:
- No API key required (local)
- Timeout: 300s (local models may be slower)
- Priority: 21 (lower than cloud providers)
- Weight: 1 (fallback priority)
- Cost: 0.0 (free)

### 3. Ollama Provider

**Location**: `layers/provider/ollama/`

**Default BaseURL**: `http://localhost:11434`

**Files Created**:
- `provider.go` - Core provider implementation
- `request.go` - Request conversion (OpenAI-compatible)
- `stream.go` - Stream response handling (SSE)

**Characteristics**:
- No API key required (local)
- Timeout: 300s (local models may be slower)
- Priority: 22 (lower than cloud providers)
- Weight: 1 (fallback priority)
- Cost: 0.0 (free)
- Default models: llama2, mistral, codellama

## Architecture Pattern

Tất cả local providers sử dụng **BaseProvider Pattern** tương tự như cloud providers:

```go
type LocalProvider struct {
    provider.BaseProvider
    apiKey  string
    baseURL string
    client  *http.Client
    models  []string
}
```

**Lợi ích**:
- Code reuse thông qua BaseProvider
- Consistent interface với cloud providers
- Built-in capabilities management
- Version tracking

## Request Conversion

Tất cả local providers sử dụng OpenAI-compatible format, nên conversion đơn giản:

```go
type LocalRequest struct {
    Model       string         `json:"model"`
    Messages    []Message      `json:"messages"`
    Temperature float64        `json:"temperature,omitempty"`
    MaxTokens   int            `json:"max_tokens,omitempty"`
    TopP        float64        `json:"top_p,omitempty"`
    Stream      bool           `json:"stream,omitempty"`
}
```

**Message Conversion**:
- System prompt → "system" role message
- User/Assistant messages → preserved with roles

## Stream Handling

Sử dụng SSE (Server-Sent Events) cho streaming:

1. Parse SSE chunks từ local provider API
2. Convert sang Anthropic format
3. Forward tới client thông qua onChunk callback
4. Handle [DONE] marker
5. Skip empty lines and comments

## Configuration

### LM Studio
```yaml
lmstudio:
  Name: "lmstudio"
  BaseURL: "http://localhost:1234/v1"
  APIKeyEnv: ""
  Models: ["local-model"]
  Format: "openai"
  TimeoutSec: 300
  Priority: 20
  Weight: 1
  CostPer1K: 0.0
```

### llama.cpp
```yaml
llamacpp:
  Name: "llamacpp"
  BaseURL: "http://localhost:8080"
  APIKeyEnv: ""
  Models: ["local-model"]
  Format: "openai"
  TimeoutSec: 300
  Priority: 21
  Weight: 1
  CostPer1K: 0.0
```

### Ollama
```yaml
ollama:
  Name: "ollama"
  BaseURL: "http://localhost:11434"
  APIKeyEnv: ""
  Models: ["llama2", "mistral", "codellama"]
  Format: "openai"
  TimeoutSec: 300
  Priority: 22
  Weight: 1
  CostPer1K: 0.0
```

## Factory Registration

Tất cả local providers được register trong `plugin_registry.go`:

```go
RegisterFactory("lmstudio", func(ctx context.Context, cfg map[string]any) (Provider, error) {
    provider := &lmstudio.LMStudioProvider{}
    configObj := &config.Config{}
    if baseURL, ok := cfg["base_url"].(string); ok {
        configObj.BaseURL = baseURL
    }
    if err := provider.Init(configObj); err != nil {
        return nil, err
    }
    return provider, nil
})
```

**Note**: Local providers không cần API key, chỉ cần BaseURL configuration.

## Testing

Do pre-existing import cycle trong layers/provider/cookie, không thể build toàn bộ project. Tuy nhiên:

- Individual layer builds pass (layers/provider/lmstudio/, llamacpp/, ollama/)
- Implementation theo pattern đã test với DeepSeek và NVIDIA NIM providers
- Code review đảm bảo consistency với existing providers

## Known Issues

**Pre-existing Import Cycle**: layers/provider/cookie → layers/provider
- Đây là issue riêng (TR-000)
- Không liên quan đến TR-015
- Cần fix import cycle trước khi build toàn bộ project

## Lessons Learned

1. **BaseProvider Pattern** rất hiệu quả cho cả cloud và local providers
2. **OpenAI-compatible format** là standard cho hầu hết local LLM servers
3. **Local providers** có timeout dài hơn (300s) vì local models có thể chậm hơn cloud
4. **No API key** cho local providers - chỉ cần BaseURL configuration
5. **Priority thấp hơn** cho local providers - dùng làm fallback hoặc development
6. **Cost 0.0** cho local providers - miễn phí vì chạy local
7. **SSE streaming** pattern consistent với cloud providers

## Usage Examples

### Using LM Studio
```bash
# Start LM Studio server
lmstudio-server --port 1234

# Configure router
curl http://localhost:1807/v1/chat/completions \
  -H "Authorization: Bearer sk-jarvis-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "local-model",
    "messages": [{"role": "user", "content": "Hello"}],
    "provider": "lmstudio"
  }'
```

### Using llama.cpp
```bash
# Start llama.cpp server
./llama-server --port 8080

# Configure router
curl http://localhost:1807/v1/chat/completions \
  -H "Authorization: Bearer sk-jarvis-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "local-model",
    "messages": [{"role": "user", "content": "Hello"}],
    "provider": "llamacpp"
  }'
```

### Using Ollama
```bash
# Start Ollama
ollama serve

# Pull model
ollama pull llama2

# Configure router
curl http://localhost:1807/v1/chat/completions \
  -H "Authorization: Bearer sk-jarvis-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
    "messages": [{"role": "user", "content": "Hello"}],
    "provider": "ollama"
  }'
```

## Next Steps

- Fix import cycle (TR-000)
- Test local providers với actual local servers
- Add health check cho local providers (check if server is running)
- Add model discovery từ /models endpoint
- Consider thêm vLLM provider (local inference server)
- Add metrics cho local provider requests

## References

- LM Studio: https://lmstudio.ai/
- llama.cpp: https://github.com/ggerganov/llama.cpp
- Ollama: https://ollama.com/
- OpenAI API specification (for format compatibility)
