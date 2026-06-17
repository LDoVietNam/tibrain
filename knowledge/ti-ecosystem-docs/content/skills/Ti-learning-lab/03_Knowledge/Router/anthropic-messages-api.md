# Anthropic Messages API Support

> **Created**: 2026-04-29  
> **Purpose**: Tài liệu về Anthropic Messages API support cho Ti Router  
> **Language**: Vietnamese

---

## Overview

Anthropic Messages API là protocol riêng của Anthropic cho Claude models. Ti Router hiện chỉ hỗ trợ OpenAI-compatible API. Cần add Anthropic Messages transport để support providers như OpenRouter, DeepSeek, LM Studio, llama.cpp, Ollama.

## Anthropic Messages API Format

### Request Format

```json
{
  "model": "claude-sonnet-4-20250514",
  "max_tokens": 1024,
  "messages": [
    {
      "role": "user",
      "content": "Hello!"
    }
  ],
  "tools": [...],
  "tool_choice": {...},
  "thinking": {
    "type": "enabled",
    "budget_tokens": 10000
  }
}
```

### Response Format

```json
{
  "id": "msg_...",
  "type": "message",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "Hello!"
    }
  ],
  "model": "claude-sonnet-4-20250514",
  "stop_reason": "end_turn",
  "stop_sequence": null,
  "usage": {
    "input_tokens": 10,
    "output_tokens": 6,
    "cache_creation_input_tokens": 0,
    "cache_read_input_tokens": 0
  }
}
```

### Thinking Blocks

```json
{
  "type": "thinking",
  "thinking": "Let me think about this...",
  "signature": "..."
}
```

## Implementation Plan

### 1. Create AnthropicMessagesTransport

```go
package provider

import (
    "encoding/json"
    "fmt"
)

type AnthropicMessagesTransport struct {
    baseURL    string
    apiKey     string
    httpClient *http.Client
}

func NewAnthropicMessagesTransport(baseURL, apiKey string) *AnthropicMessagesTransport {
    return &AnthropicMessagesTransport{
        baseURL:    baseURL,
        apiKey:     apiKey,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

func (t *AnthropicMessagesTransport) SendRequest(req AnthropicMessagesRequest) (*AnthropicMessagesResponse, error) {
    // Implement request sending
}
```

### 2. Request Translation

```go
// OpenAI Request → Anthropic Messages Request
func TranslateOpenAIToAnthropic(openaiReq OpenAIRequest) AnthropicMessagesRequest {
    anthropicReq := AnthropicMessagesRequest{
        Model:     openaiReq.Model,
        MaxTokens: openaiReq.MaxTokens,
        Messages:  translateMessages(openaiReq.Messages),
        Tools:     translateTools(openaiReq.Tools),
    }
    
    // Handle thinking blocks
    if hasThinkingBlock(openaiReq.Messages) {
        anthropicReq.Thinking = &ThinkingConfig{
            Type:         "enabled",
            BudgetTokens: 10000,
        }
    }
    
    return anthropicReq
}
```

### 3. Response Translation

```go
// Anthropic Messages Response → OpenAI Response
func TranslateAnthropicToOpenAI(anthropicResp AnthropicMessagesResponse) OpenAIResponse {
    openaiResp := OpenAIResponse{
        ID:      anthropicResp.ID,
        Object:  "chat.completion",
        Created: time.Now().Unix(),
        Model:   anthropicResp.Model,
        Choices: []Choice{
            {
                Message: Message{
                    Role:    "assistant",
                    Content: translateContent(anthropicResp.Content),
                },
                FinishReason: translateStopReason(anthropicResp.StopReason),
            },
        },
        Usage: Usage{
            PromptTokens:     anthropicResp.Usage.InputTokens,
            CompletionTokens: anthropicResp.Usage.OutputTokens,
            TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
        },
    }
    
    return openaiResp
}
```

### 4. Thinking Block Normalization

```go
func normalizeThinkingBlocks(content []ContentBlock) []ContentBlock {
    var normalized []ContentBlock
    
    for _, block := range content {
        if block.Type == "thinking" {
            // Preserve thinking block
            normalized = append(normalized, block)
        } else if block.Type == "text" {
            // Regular text block
            normalized = append(normalized, block)
        }
    }
    
    return normalized
}
```

### 5. Tool Call Normalization

```go
func normalizeToolCalls(content []ContentBlock) []ContentBlock {
    var normalized []ContentBlock
    
    for _, block := range content {
        if block.Type == "tool_use" {
            // Anthropic tool_use → OpenAI tool_calls
            normalized = append(normalized, ContentBlock{
                Type: "tool_calls",
                ToolCalls: []ToolCall{
                    {
                        ID:   block.ID,
                        Type: "function",
                        Function: Function{
                            Name:      block.Name,
                            Arguments: block.Input,
                        },
                    },
                },
            })
        } else {
            normalized = append(normalized, block)
        }
    }
    
    return normalized
}
```

### 6. Token Usage Metadata

```go
func normalizeTokenUsage(anthropicUsage AnthropicUsage) OpenAIUsage {
    return OpenAIUsage{
        PromptTokens:     anthropicUsage.InputTokens,
        CompletionTokens: anthropicUsage.OutputTokens,
        TotalTokens:      anthropicUsage.InputTokens + anthropicUsage.OutputTokens,
        // Cache tokens (if available)
        CacheReadTokens: anthropicUsage.CacheReadInputTokens,
        CacheWriteTokens: anthropicUsage.CacheCreationInputTokens,
    }
}
```

## Provider Support

### OpenRouter

```yaml
openrouter:
  name: openrouter
  base_url: https://openrouter.ai/api/v1
  api_key_env: OPENROUTER_API_KEY
  format: anthropic_messages
  models:
    - anthropic/claude-sonnet-4
    - anthropic/claude-opus-4
```

### DeepSeek

```yaml
deepseek:
  name: deepseek
  base_url: https://api.deepseek.com/v1
  api_key_env: DEEPSEEK_API_KEY
  format: anthropic_messages
  models:
    - deepseek-chat
    - deepseek-reasoner
```

### LM Studio / Ollama

```yaml
lm_studio:
  name: lm_studio
  base_url: http://localhost:1234/v1
  format: anthropic_messages
  models:
    - local-model

ollama:
  name: ollama
  base_url: http://localhost:11434/v1
  format: anthropic_messages
  models:
    - llama3
```

## Testing

```go
func TestAnthropicMessagesTransport(t *testing.T) {
    transport := NewAnthropicMessagesTransport("https://api.anthropic.com/v1", "test-key")
    
    // Test request translation
    openaiReq := OpenAIRequest{
        Model: "claude-sonnet-4",
        Messages: []Message{
            {Role: "user", Content: "Hello!"},
        },
    }
    
    anthropicReq := TranslateOpenAIToAnthropic(openaiReq)
    assert.Equal(t, "claude-sonnet-4", anthropicReq.Model)
    assert.Equal(t, 1, len(anthropicReq.Messages))
    
    // Test response translation
    anthropicResp := AnthropicMessagesResponse{
        ID: "msg_123",
        Content: []ContentBlock{
            {Type: "text", Text: "Hello!"},
        },
        Usage: AnthropicUsage{
            InputTokens:  10,
            OutputTokens: 6,
        },
    }
    
    openaiResp := TranslateAnthropicToOpenAI(anthropicResp)
    assert.Equal(t, "msg_123", openaiResp.ID)
    assert.Equal(t, 16, openaiResp.Usage.TotalTokens)
}
```

## Integration with Ti Router

### Current State

- Router có OpenAI-compatible transport
- Router có translator layer cho format detection
- Router có provider registry cho multiple providers

### Changes Needed

1. Create AnthropicMessagesTransport in layers/provider/
2. Add format detection for Anthropic Messages
3. Add request/response translators
4. Register Anthropic Messages providers
5. Add config to providers.yaml

## Success Criteria

- [x] AnthropicMessagesTransport created
- [x] Request translation implemented
- [x] Response translation implemented
- [x] Thinking blocks normalized
- [x] Tool calls normalized
- [x] Token usage metadata normalized
- [x] Providers registered (OpenRouter, DeepSeek, LM Studio, Ollama)
- [x] Tests pass
- [x] AGENTS.md updated

---

**Last Updated**: 2026-04-29
