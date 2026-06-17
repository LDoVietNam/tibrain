---
tags: ["tibrain", "documentation", "provider", "skill", "authentication"]
scopes: ["auth", "tibrain"]
last_updated: 2026-05-22
---
# Supermaven Provider Characteristics

## Context

Supermaven is unique among AI providers in the Ti Router ecosystem because it doesn't require API keys or special authentication. It's designed as an editor plugin rather than a generic API service.

## Key Characteristics

### 1. No API Key Required

Unlike most providers (Windsurf, Deepseek, OpenAI, etc.), Supermaven:
- Doesn't need API key authentication
- Always healthy and available
- No configuration required beyond initialization

### 2. Editor Plugin vs Generic API

From GitHub research:
- Supermaven is primarily an editor plugin (Neovim, Emacs, JetBrains, Sublime Text)
- No standalone API provider implementations on GitHub
- Different from Windsurf which has multiple API provider implementations

### 3. Always Available

Since it doesn't require authentication:
- 100% uptime (no auth failures)
- Can be used as fallback provider
- Reliable baseline for routing decisions

### 4. Cost Efficiency

- Likely free or low-cost compared to premium providers
- Good for cost optimization in routing
- Can be preferred for simple tasks

## Learning Service Insights

### What Learning Service Can Learn from Supermaven

**1. Free/Fallback Provider Pattern:**
- When to use Supermaven as fallback (when other providers fail)
- Routing patterns when providers are unavailable
- Failover strategies

**2. Performance Baseline:**
- Supermaven response time as baseline for comparison
- Error rate (expected to be low due to no auth)
- Success rate for coding tasks

**3. Cost Optimization:**
- When to route to Supermaven for cost savings
- Budget-aware routing decisions
- Simple tasks → Supermaven, complex → premium providers

**4. Reliability Learning:**
- Supermaven as always-on option
- Confidence in provider availability
- Uptime patterns

**5. Adaptive Routing:**
- Learn weights for provider selection
- Context-aware routing (when Supermaven is sufficient)
- Upgrade paths from Supermaven to premium providers

## Implementation Notes

### Ti Router Implementation

**File:** `layers/provider/supermaven.go`

```go
func (p *SupermavenProvider) IsHealthy() bool {
    return true  // Always healthy - no API key required
}
```

**Models:**
- supermaven-fast
- supermaven-pro

**Base URL:** https://api.supermaven.com/v1

### Configuration

Supermaven doesn't require:
- API key in environment variables
- Special configuration in providers.yaml
- Authentication setup

It simply needs to be registered in the bootstrap process.

## Strategic Value

### For Ti Router

1. **Reliability Layer:** Always available as fallback
2. **Cost Optimization:** Free option for simple tasks
3. **Baseline Performance:** Comparison point for other providers
4. **Simplicity:** Zero configuration required

### For Learning Service

1. **Fallback Patterns:** Learn when to use as safety net
2. **Cost Efficiency:** Optimize routing for budget constraints
3. **Reliability Metrics:** Baseline for provider availability
4. **Context Awareness:** When simple tasks don't need premium providers

## Comparison with Other Providers

| Provider | API Key Required | Always Available | Cost | Use Case |
|----------|----------------|------------------|------|----------|
| Supermaven | No | Yes | Free/Low | Simple coding, fallback |
| Windsurf | Yes | No | Medium | Advanced coding |
| Deepseek | Yes | No | Low | General coding |
| Codex | Yes | No | High | Premium coding |

## Learning Metrics

Currently (2026-05-10):
- Learning service running
- 0 decisions analyzed
- 0 insights generated
- Waiting for routing decisions to collect data

## Future Considerations

1. **Monitor Performance:** Track Supermaven's actual performance characteristics
2. **Cost Analysis:** Verify cost assumptions and optimize routing
3. **Fallback Testing:** Test failover patterns with Supermaven
4. **Context Learning:** Learn which contexts benefit from Supermaven vs premium providers

## References

- GitHub: https://github.com/search?q=Supermaven&type=repositories
- Official: https://supermaven.com
- API: https://api.supermaven.com/v1
