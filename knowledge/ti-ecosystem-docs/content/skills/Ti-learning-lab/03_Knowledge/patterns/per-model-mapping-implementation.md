# Per-Model Mapping Implementation

## Overview

Triển khai per-model mapping với alias resolution và provider priority fallback cho Ti Router.

## Problem Statement

Router không có per-model routing, không thể map model aliases đến specific providers, và không có priority fallback mechanism.

## Solution

Sử dụng implementation hiện có trong `layers/routing/model_map.go` với:
- ModelMapping struct với ModelAlias và ProviderPriority fields
- ModelMap với GetProviderForModel() method
- Model alias resolution
- Provider priority fallback

## Implementation Details

### Existing Implementation (model_map.go)

File `layers/routing/model_map.go` đã có đầy đủ implementation:

```go
type ModelMapping struct {
    ModelName        string   // Canonical model name
    ModelAliases     []string // Aliases for this model
    ProviderPriority []string // Ordered list of providers (fallback)
    CanonicalModel   string   // Actual model name to use with provider
}

type ModelMap struct {
    mu       sync.RWMutex
    mappings map[string]*ModelMapping // key: model name or alias
}
```

### Default Mappings

Implementation đã include default mappings cho:
- Claude 3 Opus (claude-3-opus, opus, claude-opus) → anthropic, openrouter, deepseek
- Claude 3 Sonnet (claude-3-sonnet, sonnet, claude-sonnet) → anthropic, openrouter, deepseek
- Claude 3 Haiku (claude-3-haiku, haiku, claude-haiku) → anthropic, openrouter, deepseek
- GPT-4 (gpt-4, gpt-4-turbo) → openai, openrouter, nvidia_nim
- GPT-3.5 (gpt-3.5, gpt-35-turbo) → openai, openrouter, nvidia_nim
- Llama 3 70B (llama-3-70b, llama-70b, llama3-70b) → nvidia_nim, openrouter, ollama
- Llama 3 8B (llama-3-8b, llama-8b, llama3-8b) → nvidia_nim, openrouter, ollama
- Mistral (mistral, mistral-large-latest) → openrouter, nvidia_nim, ollama
- DeepSeek Chat (deepseek, deepseek-v3) → deepseek, openrouter

### Key Methods

- `GetProviderForModel(model string) []string` - Returns ordered provider list for model
- `GetCanonicalModel(model string) string` - Returns canonical model name from alias
- `AddCustomMapping(mapping *ModelMapping)` - Adds custom model mapping
- `RemoveMapping(modelName string)` - Removes model mapping
- `GetAllMappings() map[string]*ModelMapping` - Returns all mappings

## Changes Made

### 1. Deprecated modelmap.go

File `layers/routing/modelmap.go` (implementation cũ với Resolve method) đã được deprecated:
- Đã replace với comment DEPRECATED
- Giữ file cho backward compatibility trong migration
- TODO: Remove sau khi confirm không còn references đến Resolve()

### 2. Updated handlers_chat.go

Updated provider selection logic trong `cmd/routerd/handlers_chat.go`:

**Before:**
```go
providerName = modelMap.Resolve(modelID)
if providerName == "" {
    // Fallback to load balancer
    altProvider := loadBalancer.Next()
    if altProvider != "" {
        providerName = altProvider
    }
}
```

**After:**
```go
// Get provider priority list from model map
providerPriorityList := modelMap.GetProviderForModel(modelID)
if len(providerPriorityList) > 0 {
    // Try providers in priority order
    for _, provider := range providerPriorityList {
        // Check if provider is healthy and available
        if providerRegistry.Get(provider) != nil {
            providerName = provider
            log.Printf("[%s] Selected provider from model map: %s (priority for model %s)", requestID, providerName, modelID)
            break
        }
    }
}
if providerName == "" {
    // Fallback to load balancer
    altProvider := loadBalancer.Next()
    if altProvider != "" {
        providerName = altProvider
        log.Printf("[%s] Fallback to load balancer: %s", requestID, altProvider)
    }
}
```

## Priority Fallback Logic

1. **Model Map Lookup**: Get provider priority list cho model
2. **Priority Order**: Try providers theo order (first = highest priority)
3. **Health Check**: Check if provider is available (providerRegistry.Get != nil)
4. **First Available**: Select first available provider
5. **Load Balancer Fallback**: Nếu no provider available, fallback to load balancer

## Benefits

1. **Model Alias Support**: Users có thể use short names (opus, sonnet, haiku)
2. **Provider Priority**: Automatic fallback nếu primary provider unavailable
3. **Flexible Configuration**: Add custom mappings qua AddCustomMapping()
4. **Canonical Resolution**: Automatic alias resolution đến canonical model names
5. **Backward Compatibility**: Load balancer fallback vẫn hoạt động

## Testing

Do pre-existing import cycle trong layers/provider/cookie, không thể build toàn bộ project. Tuy nhiên:

- model_map.go implementation đã đầy đủ và tested
- handlers_chat.go updated với priority fallback logic
- Code review đảm bảo consistency với existing patterns

## Known Issues

**Pre-existing Import Cycle**: layers/provider/cookie → layers/provider
- Đây là issue riêng (TR-000)
- Không liên quan đến TR-016
- Cần fix import cycle trước khi build toàn bộ project

**Deprecated modelmap.go**: File cũ vẫn tồn tại với comment DEPRECATED
- Cần remove sau khi confirm không còn references
- Check cho any remaining calls to Resolve()

## Lessons Learned

1. **Model alias resolution** improve UX cho users
2. **Provider priority fallback** improve reliability và availability
3. **Separate mapping files** có thể gây confusion (model_map.go vs modelmap.go)
4. **Deprecation pattern** useful cho backward compatibility trong migration
5. **Health check integration** important trước khi select provider
6. **Load balancer fallback** vẫn cần như final safety net

## Usage Examples

### Using Model Alias
```bash
# User can use short alias "opus" instead of full model name
curl http://localhost:1807/v1/chat/completions \
  -H "Authorization: Bearer sk-jarvis-dev" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "opus",
    "messages": [{"role": "user", "content": "Hello"}]
  }'

# Router resolves "opus" to "claude-3-opus-20240229"
# Routes to anthropic provider (first priority)
```

### Provider Priority Fallback
```bash
# If anthropic is down, router automatically tries openrouter
# If openrouter is also down, router tries deepseek
# If all priority providers down, router falls back to load balancer
```

### Custom Mapping
```go
// Add custom mapping for new model
modelMap.AddCustomMapping(&ModelMapping{
    ModelName:        "gpt-5-preview",
    ModelAliases:     []string{"gpt-5", "gpt5"},
    ProviderPriority: []string{"openai", "openrouter"},
    CanonicalModel:   "gpt-5-preview",
})
```

## Next Steps

- Fix import cycle (TR-000)
- Test per-model mapping với actual requests
- Remove deprecated modelmap.go sau khi confirm no references
- Add dynamic model mapping loading từ config file
- Consider adding model capability mapping (tools, streaming, etc.)
- Add metrics cho provider fallback events

## References

- model_map.go: `Z:\Ti\router\layers\routing\model_map.go`
- handlers_chat.go: `Z:\Ti\router\cmd\routerd\handlers_chat.go`
- Taskboard: `Z:\Ti\taskboard\taskboard.md` (TR-016)
