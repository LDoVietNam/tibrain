# Per-Model Routing (Opus/Sonnet/Haiku) + Amp Agent Modes

> **Created**: 2026-04-29  
> **Updated**: 2026-05-20 (Added Amp Agent Modes integration)
> **Purpose**: Tài liệu về per-model routing cho Ti Router với Amp-style Task-Based Model Routing
> **Language**: Vietnamese

---

## Overview

Per-model routing là pattern route requests theo model tiers (Opus/Sonnet/Haiku) và Amp Agent Modes (Smart/Deep/Rush). Pattern này được sử dụng bởi free-claude-code-main và Amp.com để optimize cost và performance.

## Tiers

| Tier | Description | Use Case | Cost | Amp Mode Mapping |
|------|-------------|----------|------|-----------------|
| **Opus** | High-end model | Complex reasoning, code generation | High | Smart/Deep |
| **Sonnet** | Mid-tier model | General purpose, balanced cost/performance | Medium | Smart |
| **Haiku** | Low-tier model | Simple tasks, fast response | Low | Rush |

## Amp Agent Modes

Reference: https://ampcode.com/models

| Mode | Description | Use Case | Model Tier |
|------|-------------|----------|------------|
| **Smart** | SOTA models without constraints for maximum capability | Complex tasks, general purpose | Opus/Sonnet |
| **Deep** | Extended thinking with GPT-5.5 for complex reasoning | Reasoning-heavy tasks, security-heavy tasks | Opus |
| **Rush** | Faster, cheaper, and less capable | Small, well-defined tasks | Haiku |

### Agent Mode Mapping Logic

```go
// Map task complexity to Amp Agent Mode
func MapToAgentMode(complexityScore float64, securityIndicators int) string {
    if complexityScore <= 0.3 {
        return "rush"  // Simple tasks → Haiku
    } else if complexityScore <= 0.6 {
        return "smart"  // Standard tasks → Sonnet
    } else if securityIndicators > 0 {
        return "deep"  // Security-heavy tasks → Opus
    } else {
        return "smart"  // Complex tasks → Opus/Sonnet
    }
}
```

## Implementation Plan

### 1. Add Tier Field to ModelMetadata

```go
type ModelMetadata struct {
    ID           string   `json:"id"`
    Provider     string   `json:"provider"`
    Tier         string   `json:"tier"` // "opus", "sonnet", "haiku"
    Capabilities []string `json:"capabilities"`
    CostPer1K    float64  `json:"cost_per_1k"`
    MaxTokens    int      `json:"max_tokens"`
    Status       string   `json:"status"`
    ContextSize  int      `json:"context_size"`
}
```

### 2. Add Config Fields

**providers.yaml:**
```yaml
model_tiers:
  opus: "groq/llama-3.3-70b-versatile"
  sonnet: "openrouter/deepseek/deepseek-r1"
  haiku: "gitlab/gemma-2-9b-it"
```

**Tiserverrouter.yaml:**
```yaml
model_routing:
  tier_based: true
  default_tier: "sonnet"
  fallback_tier: "haiku"
```

### 3. Update ModelRegistry

```go
// GetByTier returns all models for a specific tier
func (r *ModelRegistry) GetByTier(tier string) []ModelMetadata {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    var models []ModelMetadata
    for _, metadata := range r.models {
        if metadata.Tier == tier {
            models = append(models, metadata)
        }
    }
    return models
}

// ResolveTier resolves a Claude model name to tier
func (r *ModelRegistry) ResolveTier(modelName string) string {
    // claude-opus-4.6 -> opus
    // claude-sonnet-4.6 -> sonnet
    // claude-haiku-4.5 -> haiku
    // Default: sonnet
}
```

### 4. Update Selector

```go
// SelectByTier selects model based on tier
func (s *Selector) SelectByTier(tier string) (string, error) {
    models := s.modelRegistry.GetByTier(tier)
    if len(models) == 0 {
        return "", fmt.Errorf("no models found for tier: %s", tier)
    }
    // Select based on load balancing, health, cost
    return s.selectBestModel(models)
}
```

## Use Cases

### 1. Cost Optimization

- Route simple tasks to Haiku (low cost)
- Route complex tasks to Opus (high quality)
- Route general tasks to Sonnet (balanced)

### 2. Performance Optimization

- Fast response for simple queries (Haiku)
- High quality for complex queries (Opus)
- Balanced for general queries (Sonnet)

### 3. Load Balancing

- Distribute load across tiers
- Avoid overloading high-cost models
- Scale based on request complexity

## Testing

```go
func TestPerModelRouting(t *testing.T) {
    registry := NewModelRegistry()
    
    // Register models with tiers
    registry.Register(ModelMetadata{
        ID: "groq/llama-3.3-70b-versatile",
        Tier: "opus",
    })
    registry.Register(ModelMetadata{
        ID: "openrouter/deepseek/deepseek-r1",
        Tier: "sonnet",
    })
    registry.Register(ModelMetadata{
        ID: "gitlab/gemma-2-9b-it",
        Tier: "haiku",
    })
    
    // Test tier resolution
    opusModels := registry.GetByTier("opus")
    assert.Equal(t, 1, len(opusModels))
    
    // Test tier selection
    selector := NewSelector(registry)
    model, err := selector.SelectByTier("opus")
    assert.NoError(t, err)
    assert.Equal(t, "groq/llama-3.3-70b-versatile", model)
}
```

## Integration with Ti Router

### Current State

- Router có ModelRegistry với ModelMetadata
- Router có Selector để chọn model
- Router có LoadBalancer để distribute load

### Changes Needed

1. Add Tier field to ModelMetadata
2. Add GetByTier method to ModelRegistry
3. Add ResolveTier method to ModelRegistry
4. Add SelectByTier method to Selector
5. Add tier-based routing logic to LoadBalancer
6. Add config fields to providers.yaml and Tiserverrouter.yaml

## Success Criteria

- [x] Tier field added to ModelMetadata
- [x] GetByTier method implemented
- [x] ResolveTier method implemented
- [x] SelectByTier method implemented
- [x] Config fields added
- [x] Tests pass
- [x] AGENTS.md updated

---

**Last Updated**: 2026-04-29
