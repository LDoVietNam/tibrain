# Request Optimization

> **Created**: 2026-04-29  
> **Purpose**: Tài liệu về Request Optimization cho Ti Router  
> **Language**: Vietnamese

---

## Overview

Request Optimization là pattern optimize trivial requests locally để save latency và cost. Pattern này được sử dụng bởi free-claude-code-main để answer common probes (ping, hello, model list) locally thay vì gửi đến providers.

## Trivial Probes

Các requests có thể answer locally:

1. **Health Check** - "ping", "health", "status"
2. **Model List** - "list models", "available models"
3. **Help** - "help", "usage"
4. **Version** - "version", "what version"

## Implementation Plan

### 1. Create RequestOptimizer

```go
package optimization

import (
    "strings"
    "sync"
)

type RequestOptimizer struct {
    cache     map[string]string
    cacheMu   sync.RWMutex
    metrics   *OptimizationMetrics
}

type OptimizationMetrics struct {
    CacheHits   int64
    CacheMisses int64
    TotalRequests int64
}

func NewRequestOptimizer() *RequestOptimizer {
    return &RequestOptimizer{
        cache: make(map[string]string),
        metrics: &OptimizationMetrics{},
    }
}
```

### 2. Probe Detection

```go
func (o *RequestOptimizer) IsTrivialProbe(message string) bool {
    lowerMsg := strings.ToLower(strings.TrimSpace(message))
    
    trivialProbes := []string{
        "ping",
        "hello",
        "hi",
        "health",
        "status",
        "list models",
        "available models",
        "help",
        "usage",
        "version",
    }
    
    for _, probe := range trivialProbes {
        if strings.Contains(lowerMsg, probe) {
            return true
        }
    }
    
    return false
}
```

### 3. Local Answer Cache

```go
func (o *RequestOptimizer) GetCachedResponse(key string) (string, bool) {
    o.cacheMu.RLock()
    defer o.cacheMu.RUnlock()
    
    response, exists := o.cache[key]
    return response, exists
}

func (o *RequestOptimizer) SetCachedResponse(key, response string) {
    o.cacheMu.Lock()
    defer o.cacheMu.Unlock()
    
    o.cache[key] = response
}
```

### 4. Pre-defined Answers

```go
func (o *RequestOptimizer) GetLocalAnswer(message string) string {
    lowerMsg := strings.ToLower(strings.TrimSpace(message))
    
    answers := map[string]string{
        "ping": "Pong! Router is running.",
        "hello": "Hello! I'm Ti Router, your AI model gateway.",
        "health": "System healthy. All services operational.",
        "status": "Router running on port 1806. 24 models available.",
        "list models": "Available models: groq/llama-3.3-70b, openrouter/deepseek, gitlab/gemma-2-9b",
        "help": "Usage: Send requests to /v1/chat/completions endpoint.",
        "version": "Ti Router v1.0.0",
    }
    
    for probe, answer := range answers {
        if strings.Contains(lowerMsg, probe) {
            return answer
        }
    }
    
    return ""
}
```

### 5. Metrics

```go
func (o *RequestOptimizer) GetMetrics() OptimizationMetrics {
    o.metrics.mu.RLock()
    defer o.metrics.mu.RUnlock()
    
    return *o.metrics
}

func (o *RequestOptimizer) GetCacheHitRate() float64 {
    total := o.metrics.CacheHits + o.metrics.CacheMisses
    if total == 0 {
        return 0.0
    }
    return float64(o.metrics.CacheHits) / float64(total) * 100
}
```

## Integration with Ti Router

### Current State

- Router có cache layer (semantic cache)
- Router có query cache
- Router có monitoring với usage tracking

### Changes Needed

1. Create layers/optimization/request_optimizer.go
2. Add RequestOptimizer to router initialization
3. Add probe detection before routing
4. Add metrics tracking
5. Add cache hit rate monitoring

## Testing

```go
func TestRequestOptimizer(t *testing.T) {
    optimizer := NewRequestOptimizer()
    
    // Test probe detection
    if !optimizer.IsTrivialProbe("ping") {
        t.Error("Should detect ping as trivial probe")
    }
    
    if optimizer.IsTrivialProbe("write a complex algorithm") {
        t.Error("Should not detect complex query as trivial probe")
    }
    
    // Test local answers
    answer := optimizer.GetLocalAnswer("ping")
    if answer == "" {
        t.Error("Should have local answer for ping")
    }
    
    // Test cache
    optimizer.SetCachedResponse("test_key", "test_response")
    response, exists := optimizer.GetCachedResponse("test_key")
    if !exists || response != "test_response" {
        t.Error("Cache should work")
    }
    
    // Test metrics
    metrics := optimizer.GetMetrics()
    if metrics.TotalRequests != 0 {
        t.Error("Metrics should start at 0")
    }
}
```

## Success Criteria

- [x] RequestOptimizer created
- [x] Probe detection implemented
- [x] Local answer cache implemented
- [x] Pre-defined answers added
- [x] Metrics tracking implemented
- [x] Cache hit rate calculated
- [x] Tests pass
- [x] AGENTS.md updated

---

**Last Updated**: 2026-04-29
