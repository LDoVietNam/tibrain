# Ti Router Optimization Blueprint

> **Version**: 1.0.0  
> **Created**: 2026-05-04  
> **Scope**: Advanced router optimization for AI agent orchestration  
> **Target**: Ti Router (apps/router/layers/)

---

## 🎯 Overview

Document này cung cấp blueprint toàn diện để optimize Ti Router với 12 điểm nâng cấp chính:

1. **Semantic Routing** - Thay thế rule-based routing với semantic understanding
2. **Prompt Caching** - Active caching strategy cho prompts
3. **Adaptive Timeout + Retry** - Exponential backoff với adaptive timeout
4. **Auto Model Selection** - Orchestrator + specialist model routing
5. **3-Tier Memory** - Phân tầng memory tối thiểu 3 lớp
6. **Tool-First Design** - Tool registry + discovery
7. **Dynamic Context Compression** - Tối ưu context window
8. **Personal + Long-term Preference** - Personalization system
9. **Session Continuity** - Maintain session state
10. **Agent Step Tracing** - Trace mọi agent step
11. **Automatic Eval Loop** - Self-evaluation loop
12. **Tool Registry + Discovery** - Dynamic tool discovery

---

## 📊 Current State Analysis

### ✅ Đã có

| Component | Status | Notes |
|-----------|--------|-------|
| Adaptive Routing | ✅ Implemented | Latency tracker, cost tracker, health monitor, circuit breaker |
| Provider Layer | ✅ Implemented | Multiple LLM providers (Anthropic, OpenAI, Groq, etc.) |
| Tag-based Routing | ✅ Implemented | Tag-based routing logic |
| Load Balancing | ✅ Implemented | Weighted routing algorithm |
| Fallback Chains | ✅ Implemented | Provider fallback mechanism |
| Retry Logic | ✅ Implemented | Basic retry in provider layer |
| Circuit Breaker | ✅ Implemented | Health monitor with circuit breaker |
| Rate Limiting | ✅ Implemented | Per-key rate tracking |
| Audit Logging | ✅ Implemented | Audit layer |
| OAuth/Auth | ✅ Implemented | Authentication layer |
| MCP Integration | ✅ Implemented | MCP tool routing (partial) |
| TiBrain Platform | ✅ Designed | Skill platform architecture |

### ❌ Chưa có

| Component | Priority | Impact |
|-----------|----------|--------|
| Semantic Routing | 🔴 High | Tăng accuracy routing từ 70% → 90%+ |
| Prompt Caching | 🔴 High | Giảm cost 30-50% |
| Adaptive Timeout | 🟡 Medium | Tăng reliability |
| Auto Model Selection | 🔴 High | Tối ưu cost/performance |
| 3-Tier Memory | 🔴 High | Tăng context retention |
| Tool-First Design | 🔴 High | Tăng flexibility |
| Dynamic Context Compression | 🟡 Medium | Tăng context window utilization |
| Personal Preferences | 🟢 Low | Tăng UX |
| Session Continuity | 🟡 Medium | Tăng multi-turn capability |
| Agent Step Tracing | 🔴 High | Debugging + observability |
| Automatic Eval Loop | 🔴 High | Self-improvement |
| Tool Registry + Discovery | 🔴 High | Extensibility |

---

## 🏗️ Architecture Design

### 1. Semantic Routing

#### Mục tiêu
Thay thế rule-based routing (if-else, tags) với semantic understanding của user intent.

#### Implementation

```go
// apps/router/layers/routing/semantic_router.go

package routing

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "time"
    
    "github.com/ti/embeddings"
    "github.com/ti/vectorstore"
)

// SemanticRouter uses embeddings for intent classification
type SemanticRouter struct {
    embeddingClient  *embeddings.Client
    vectorStore      *vectorstore.VectorStore
    intentClassifier *IntentClassifier
    routeCache       *RouteCache
}

// Intent represents a classified user intent
type Intent struct {
    Category        string            `json:"category"`        // coding, research, writing, analysis
    Subcategory     string            `json:"subcategory"`     // bug-fix, feature-add, refactor
    Complexity      string            `json:"complexity"`      // simple, standard, complex, critical
    Confidence      float64           `json:"confidence"`      // 0-1
    RequiredTools   []string          `json:"required_tools"`   // git, file-edit, web-search
    ModelPreference string            `json:"model_preference"` // gpt-4, claude-3, llama3
    Metadata        map[string]string `json:"metadata"`
}

// IntentClassifier uses semantic similarity to classify intents
type IntentClassifier struct {
    intentEmbeddings map[string][]float64
    threshold       float64
}

// NewSemanticRouter creates a new semantic router
func NewSemanticRouter(embeddingClient *embeddings.Client, vectorStore *vectorstore.VectorStore) *SemanticRouter {
    return &SemanticRouter{
        embeddingClient:  embeddingClient,
        vectorStore:      vectorStore,
        intentClassifier: NewIntentClassifier(),
        routeCache:       NewRouteCache(1000),
    }
}

// ClassifyIntent classifies user request into intent
func (sr *SemanticRouter) ClassifyIntent(ctx context.Context, request string) (*Intent, error) {
    // Check cache first
    if cached, ok := sr.routeCache.Get(request); ok {
        return cached.(*Intent), nil
    }
    
    // Get embedding for request
    embedding, err := sr.embeddingClient.GetEmbedding(ctx, request)
    if err != nil {
        return nil, fmt.Errorf("failed to get embedding: %w", err)
    }
    
    // Classify intent using semantic similarity
    intent, err := sr.intentClassifier.Classify(embedding)
    if err != nil {
        return nil, fmt.Errorf("failed to classify intent: %w", err)
    }
    
    // Cache result
    sr.routeCache.Set(request, intent, time.Hour)
    
    return intent, nil
}

// SelectRoute selects the best route based on intent
func (sr *SemanticRouter) SelectRoute(ctx context.Context, intent *Intent) (*Route, error) {
    // Build route based on intent
    route := &Route{
        Provider:       selectProvider(intent),
        Model:          selectModel(intent),
        Tools:          intent.RequiredTools,
        Timeout:        selectTimeout(intent.Complexity),
        RetryStrategy:  selectRetryStrategy(intent.Complexity),
        MemoryConfig:   selectMemoryConfig(intent.Complexity),
    }
    
    return route, nil
}

// Route represents a routing decision
type Route struct {
    Provider       string
    Model          string
    Tools          []string
    Timeout        time.Duration
    RetryStrategy  RetryStrategy
    MemoryConfig   MemoryConfig
    CacheStrategy  CacheStrategy
}

// Intent classification logic
func (ic *IntentClassifier) Classify(embedding []float64) (*Intent, error) {
    // Find most similar intent
    bestMatch := ""
    bestScore := 0.0
    
    for intent, intentEmbedding := range ic.intentEmbeddings {
        score := cosineSimilarity(embedding, intentEmbedding)
        if score > bestScore {
            bestScore = score
            bestMatch = intent
        }
    }
    
    if bestScore < ic.threshold {
        return nil, fmt.Errorf("intent confidence too low: %.2f", bestScore)
    }
    
    // Parse intent string to get structured data
    return parseIntent(bestMatch, bestScore)
}

func cosineSimilarity(a, b []float64) float64 {
    // Calculate cosine similarity
    var dotProduct, normA, normB float64
    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    return dotProduct / (sqrt(normA) * sqrt(normB))
}
```

#### Training Data

```yaml
# apps/router/layers/routing/intent_training_data.yaml

intents:
  coding:
    bug_fix:
      examples:
        - "fix the bug in the login function"
        - "the authentication is failing"
        - "debug the error in user service"
      required_tools: [git, file-edit, test]
      model_preference: "claude-3-opus"
      complexity: "standard"
    
    feature_add:
      examples:
        - "add a new endpoint for user registration"
        - "implement a new feature for file upload"
        - "create a new component for dashboard"
      required_tools: [git, file-edit, test, docs]
      model_preference: "gpt-4"
      complexity: "standard"
    
    refactor:
      examples:
        - "refactor the authentication module"
        - "optimize the database queries"
        - "improve the code structure"
      required_tools: [git, file-edit, test, review]
      model_preference: "claude-3-opus"
      complexity: "complex"
  
  research:
    web_search:
      examples:
        - "search for information about Rust patterns"
        - "find documentation for React hooks"
        - "research best practices for API design"
      required_tools: [web-search, docs]
      model_preference: "gpt-4"
      complexity: "standard"
    
    analysis:
      examples:
        - "analyze the codebase architecture"
        - "review the security vulnerabilities"
        - "evaluate the performance bottlenecks"
      required_tools: [file-read, analysis, review]
      model_preference: "claude-3-opus"
      complexity: "complex"
  
  writing:
    documentation:
      examples:
        - "write documentation for the API"
        - "create a README for the project"
        - "document the installation process"
      required_tools: [file-write, docs]
      model_preference: "gpt-4"
      complexity: "simple"
    
    content:
      examples:
        - "write a blog post about the new feature"
        - "create a tutorial for using the CLI"
        - "generate marketing content"
      required_tools: [file-write, content]
      model_preference: "gpt-4"
      complexity: "standard"
```

#### Integration with Existing Router

```go
// apps/router/layers/engine/decision.go (modified)

func (e *Engine) MakeRoutingDecision(ctx context.Context, request string) (*Route, error) {
    // Try semantic routing first
    if e.semanticRouter != nil {
        intent, err := e.semanticRouter.ClassifyIntent(ctx, request)
        if err == nil && intent.Confidence > 0.8 {
            route, err := e.semanticRouter.SelectRoute(ctx, intent)
            if err == nil {
                return route, nil
            }
        }
    }
    
    // Fallback to tag-based routing
    return e.tagBasedRouter.SelectRoute(ctx, request)
}
```

---

### 2. Prompt Caching

#### Mục tiêu
Active caching strategy để giảm cost và latency.

#### Implementation

```go
// apps/router/layers/cache/prompt_cache.go

package cache

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/redis/go-redis/v9"
)

// PromptCache implements active prompt caching
type PromptCache struct {
    redis      *redis.Client
    localCache *LRUCache
    stats      *CacheStats
}

// CacheEntry represents a cached prompt
type CacheEntry struct {
    Key           string        `json:"key"`
    Prompt        string        `json:"prompt"`
    Response      string        `json:"response"`
    Tokens        int           `json:"tokens"`
    Cost          float64       `json:"cost"`
    Latency       time.Duration `json:"latency"`
    HitCount      int           `json:"hit_count"`
    CreatedAt     time.Time     `json:"created_at"`
    LastAccessed  time.Time     `json:"last_accessed"`
    TTL           time.Duration `json:"ttl"`
}

// CacheStats tracks cache performance
type CacheStats struct {
    Hits          int64
    Misses        int64
    TotalTokens   int64
    SavedCost     float64
    SavedLatency  time.Duration
}

// NewPromptCache creates a new prompt cache
func NewPromptCache(redisAddr string) *PromptCache {
    return &PromptCache{
        redis:      redis.NewClient(&redis.Options{Addr: redisAddr}),
        localCache: NewLRUCache(1000),
        stats:      &CacheStats{},
    }
}

// Get retrieves cached response for prompt
func (pc *PromptCache) Get(ctx context.Context, prompt string) (*CacheEntry, bool) {
    // Check local cache first (fastest)
    if entry, ok := pc.localCache.Get(prompt); ok {
        cacheEntry := entry.(*CacheEntry)
        cacheEntry.HitCount++
        cacheEntry.LastAccessed = time.Now()
        pc.stats.Hits++
        return cacheEntry, true
    }
    
    // Check Redis cache (fallback)
    key := pc.hashPrompt(prompt)
    data, err := pc.redis.Get(ctx, key).Bytes()
    if err == nil {
        var entry CacheEntry
        if err := json.Unmarshal(data, &entry); err == nil {
            entry.HitCount++
            entry.LastAccessed = time.Now()
            pc.localCache.Set(prompt, &entry, entry.TTL)
            pc.stats.Hits++
            return &entry, true
        }
    }
    
    pc.stats.Misses++
    return nil, false
}

// Set stores prompt and response in cache
func (pc *PromptCache) Set(ctx context.Context, prompt, response string, tokens int, cost float64, latency time.Duration, ttl time.Duration) {
    key := pc.hashPrompt(prompt)
    
    entry := &CacheEntry{
        Key:          key,
        Prompt:       prompt,
        Response:     response,
        Tokens:       tokens,
        Cost:         cost,
        Latency:      latency,
        HitCount:     0,
        CreatedAt:    time.Now(),
        LastAccessed: time.Now(),
        TTL:          ttl,
    }
    
    // Store in local cache
    pc.localCache.Set(prompt, entry, ttl)
    
    // Store in Redis (async)
    go func() {
        data, _ := json.Marshal(entry)
        pc.redis.Set(context.Background(), key, data, ttl)
    }()
}

// hashPrompt creates a hash key for the prompt
func (pc *PromptCache) hashPrompt(prompt string) string {
    hash := sha256.Sum256([]byte(prompt))
    return hex.EncodeToString(hash[:])
}

// GetStats returns cache statistics
func (pc *PromptCache) GetStats() *CacheStats {
    return pc.stats
}

// GetHitRate returns cache hit rate
func (pc *PromptCache) GetHitRate() float64 {
    total := pc.stats.Hits + pc.stats.Misses
    if total == 0 {
        return 0
    }
    return float64(pc.stats.Hits) / float64(total)
}
```

#### Caching Strategy

```go
// apps/router/layers/cache/cache_strategy.go

package cache

import (
    "context"
    "time"
)

// CacheStrategy defines caching behavior
type CacheStrategy struct {
    EnableSystemPrompt  bool          // Cache system prompts
    EnableUserPrompt    bool          // Cache user prompts
    EnableConversation  bool          // Cache conversation history
    EnableToolResults   bool          // Cache tool results
    SystemPromptTTL     time.Duration // TTL for system prompts (24h)
    UserPromptTTL       time.Duration // TTL for user prompts (1h)
    ConversationTTL     time.Duration // TTL for conversations (6h)
    ToolResultTTL       time.Duration // TTL for tool results (30min)
    MaxCacheSize        int           // Max cache entries
    CompressionEnabled  bool          // Enable compression
}

// DefaultCacheStrategy returns default caching strategy
func DefaultCacheStrategy() *CacheStrategy {
    return &CacheStrategy{
        EnableSystemPrompt:  true,
        EnableUserPrompt:    true,
        EnableConversation:  false, // Disabled by default (too dynamic)
        EnableToolResults:   true,
        SystemPromptTTL:     24 * time.Hour,
        UserPromptTTL:       1 * time.Hour,
        ConversationTTL:     6 * time.Hour,
        ToolResultTTL:       30 * time.Minute,
        MaxCacheSize:        10000,
        CompressionEnabled:  true,
    }
}

// AggressiveCacheStrategy returns aggressive caching for cost optimization
func AggressiveCacheStrategy() *CacheStrategy {
    return &CacheStrategy{
        EnableSystemPrompt:  true,
        EnableUserPrompt:    true,
        EnableConversation:  true,
        EnableToolResults:   true,
        SystemPromptTTL:     48 * time.Hour,
        UserPromptTTL:       6 * time.Hour,
        ConversationTTL:     12 * time.Hour,
        ToolResultTTL:       2 * time.Hour,
        MaxCacheSize:        50000,
        CompressionEnabled:  true,
    }
}
```

---

### 3. Adaptive Timeout + Retry with Backoff

#### Mục tiêu
Adaptive timeout với exponential backoff để tăng reliability.

#### Implementation

```go
// apps/router/layers/resilience/adaptive_retry.go

package resilience

import (
    "context"
    "math"
    "time"
)

// AdaptiveRetryConfig configures adaptive retry behavior
type AdaptiveRetryConfig struct {
    MaxRetries          int           // Maximum retry attempts
    InitialTimeout      time.Duration // Initial timeout
    MaxTimeout         time.Duration // Maximum timeout
    BackoffMultiplier  float64       // Backoff multiplier
    JitterEnabled      bool          // Enable jitter
    JitterAmount       float64       // Jitter amount (0-1)
    AdaptiveTimeout    bool          // Enable adaptive timeout
    TimeoutMultiplier  float64       // Timeout multiplier for adaptive
    RetryableErrors    []string      // Error types to retry
}

// DefaultAdaptiveRetryConfig returns default config
func DefaultAdaptiveRetryConfig() *AdaptiveRetryConfig {
    return &AdaptiveRetryConfig{
        MaxRetries:         3,
        InitialTimeout:     10 * time.Second,
        MaxTimeout:        60 * time.Second,
        BackoffMultiplier:  2.0,
        JitterEnabled:      true,
        JitterAmount:       0.1,
        AdaptiveTimeout:   true,
        TimeoutMultiplier:  1.5,
        RetryableErrors:    []string{"timeout", "rate_limit", "server_error"},
    }
}

// AdaptiveRetry implements adaptive retry with backoff
type AdaptiveRetry struct {
    config      *AdaptiveRetryConfig
    latencyTracker *LatencyTracker
}

// NewAdaptiveRetry creates a new adaptive retry
func NewAdaptiveRetry(config *AdaptiveRetryConfig, latencyTracker *LatencyTracker) *AdaptiveRetry {
    return &AdaptiveRetry{
        config:         config,
        latencyTracker: latencyTracker,
    }
}

// Execute executes operation with adaptive retry
func (ar *AdaptiveRetry) Execute(ctx context.Context, op func(context.Context) error) error {
    var lastErr error
    timeout := ar.config.InitialTimeout
    
    for attempt := 0; attempt <= ar.config.MaxRetries; attempt++ {
        // Calculate timeout for this attempt
        if ar.config.AdaptiveTimeout && attempt > 0 {
            // Adjust timeout based on historical latency
            avgLatency := ar.latencyTracker.GetAverageLatency()
            if avgLatency > 0 {
                timeout = time.Duration(float64(avgLatency) * ar.config.TimeoutMultiplier)
            }
        }
        
        // Cap timeout at max
        if timeout > ar.config.MaxTimeout {
            timeout = ar.config.MaxTimeout
        }
        
        // Create context with timeout
        attemptCtx, cancel := context.WithTimeout(ctx, timeout)
        
        // Execute operation
        start := time.Now()
        err := op(attemptCtx)
        latency := time.Since(start)
        cancel()
        
        // Record latency
        ar.latencyTracker.Record(latency, err == nil)
        
        // If success, return
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // Check if error is retryable
        if !ar.isRetryable(err) {
            return err
        }
        
        // If last attempt, return error
        if attempt == ar.config.MaxRetries {
            return lastErr
        }
        
        // Calculate backoff
        backoff := ar.calculateBackoff(attempt)
        
        // Add jitter if enabled
        if ar.config.JitterEnabled {
            backoff = ar.addJitter(backoff)
        }
        
        // Wait before retry
        select {
        case <-time.After(backoff):
            continue
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    return lastErr
}

// calculateBackoff calculates exponential backoff
func (ar *AdaptiveRetry) calculateBackoff(attempt int) time.Duration {
    backoff := float64(ar.config.InitialTimeout) * math.Pow(ar.config.BackoffMultiplier, float64(attempt))
    return time.Duration(backoff)
}

// addJitter adds random jitter to backoff
func (ar *AdaptiveRetry) addJitter(backoff time.Duration) time.Duration {
    jitter := backoff * time.Duration(ar.config.JitterAmount)
    return backoff + time.Duration(rand.Float64()*float64(jitter))
}

// isRetryable checks if error is retryable
func (ar *AdaptiveRetry) isRetryable(err error) bool {
    errMsg := err.Error()
    for _, retryable := range ar.config.RetriableErrors {
        if contains(errMsg, retryable) {
            return true
        }
    }
    return false
}
```

---

### 4. Auto Model Selection (Orchestrator + Specialist)

#### Mục tiêu
Tự động chọn model phù hợp: orchestrator cho planning, specialist cho execution.

#### Implementation

```go
// apps/router/layers/routing/model_selector.go

package routing

import (
    "context"
    "fmt"
)

// ModelSelector implements automatic model selection
type ModelSelector struct {
    orchestratorModels []ModelConfig
    specialistModels   map[string][]ModelConfig
    costTracker        *CostTracker
    latencyTracker     *LatencyTracker
}

// ModelConfig represents a model configuration
type ModelConfig struct {
    ID              string
    Provider        string
    Name            string
    Type            ModelType // orchestrator, specialist
    Capabilities    []string
    CostPer1KTokens CostInfo
    QualityScore    float64
    Speed           string
    ContextWindow   int
}

// ModelType represents model type
type ModelType string

const (
    ModelTypeOrchestrator ModelType = "orchestrator"
    ModelTypeSpecialist   ModelType = "specialist"
)

// NewModelSelector creates a new model selector
func NewModelSelector() *ModelSelector {
    ms := &ModelSelector{
        orchestratorModels: []ModelConfig{
            {
                ID:           "anthropic:claude-3-opus",
                Provider:     "anthropic",
                Name:         "claude-3-opus",
                Type:         ModelTypeOrchestrator,
                Capabilities: []string{"planning", "reasoning", "coding"},
                CostPer1KTokens: CostInfo{Input: 0.015, Output: 0.075},
                QualityScore: 9.5,
                Speed:        "medium",
                ContextWindow: 200000,
            },
            {
                ID:           "openai:gpt-4",
                Provider:     "openai",
                Name:         "gpt-4",
                Type:         ModelTypeOrchestrator,
                Capabilities: []string{"planning", "reasoning", "coding"},
                CostPer1KTokens: CostInfo{Input: 0.03, Output: 0.06},
                QualityScore: 9.0,
                Speed:        "medium",
                ContextWindow: 128000,
            },
        },
        specialistModels: map[string][]ModelConfig{
            "coding": {
                {
                    ID:           "anthropic:claude-3-sonnet",
                    Provider:     "anthropic",
                    Name:         "claude-3-sonnet",
                    Type:         ModelTypeSpecialist,
                    Capabilities: []string{"coding", "debugging"},
                    CostPer1KTokens: CostInfo{Input: 0.003, Output: 0.015},
                    QualityScore: 8.5,
                    Speed:        "fast",
                    ContextWindow: 200000,
                },
                {
                    ID:           "groq:llama3-70b",
                    Provider:     "groq",
                    Name:         "llama3-70b",
                    Type:         ModelTypeSpecialist,
                    Capabilities: []string{"coding", "fast-execution"},
                    CostPer1KTokens: CostInfo{Input: 0.00059, Output: 0.00079},
                    QualityScore: 7.5,
                    Speed:        "very-fast",
                    ContextWindow: 8192,
                },
            },
            "research": {
                {
                    ID:           "openai:gpt-4-turbo",
                    Provider:     "openai",
                    Name:         "gpt-4-turbo",
                    Type:         ModelTypeSpecialist,
                    Capabilities: []string{"research", "analysis"},
                    CostPer1KTokens: CostInfo{Input: 0.01, Output: 0.03},
                    QualityScore: 8.5,
                    Speed:        "fast",
                    ContextWindow: 128000,
                },
            },
            "writing": {
                {
                    ID:           "anthropic:claude-3-haiku",
                    Provider:     "anthropic",
                    Name:         "claude-3-haiku",
                    Type:         ModelTypeSpecialist,
                    Capabilities: []string{"writing", "summarization"},
                    CostPer1KTokens: CostInfo{Input: 0.00025, Output: 0.00125},
                    QualityScore: 7.0,
                    Speed:        "very-fast",
                    ContextWindow: 200000,
                },
            },
        },
    }
    
    return ms
}

// SelectOrchestrator selects the best orchestrator model
func (ms *ModelSelector) SelectOrchestrator(ctx context.Context, intent *Intent) (*ModelConfig, error) {
    // Filter orchestrator models by capability
    candidates := ms.filterByCapability(ms.orchestratorModels, intent.RequiredTools)
    
    if len(candidates) == 0 {
        return nil, fmt.Errorf("no orchestrator models available")
    }
    
    // Score and rank candidates
    scored := ms.scoreModels(candidates, intent)
    
    // Select best model
    return scored[0], nil
}

// SelectSpecialist selects the best specialist model
func (ms *ModelSelector) SelectSpecialist(ctx context.Context, intent *Intent) (*ModelConfig, error) {
    // Determine specialist type from intent
    specialistType := ms.determineSpecialistType(intent)
    
    // Get specialist models for type
    candidates, ok := ms.specialistModels[specialistType]
    if !ok || len(candidates) == 0 {
        // Fallback to orchestrator
        return ms.SelectOrchestrator(ctx, intent)
    }
    
    // Score and rank candidates
    scored := ms.scoreModels(candidates, intent)
    
    // Select best model
    return scored[0], nil
}

// scoreModels scores models based on multiple factors
func (ms *ModelSelector) scoreModels(models []ModelConfig, intent *Intent) []*ModelConfig {
    scored := make([]*ModelConfig, len(models))
    
    for i, model := range models {
        score := 0.0
        
        // Quality score (40%)
        score += model.QualityScore * 0.4
        
        // Cost efficiency (30%)
        costScore := 1.0 / (model.CostPer1KTokens.Input + model.CostPer1KTokens.Output)
        score += costScore * 0.3
        
        // Speed (20%)
        speedScore := ms.getSpeedScore(model.Speed)
        score += speedScore * 0.2
        
        // Context window (10%)
        contextScore := float64(model.ContextWindow) / 200000.0
        score += contextScore * 0.1
        
        scored[i] = &model
        scored[i].QualityScore = score
    }
    
    // Sort by score descending
    sort.Slice(scored, func(i, j int) bool {
        return scored[i].QualityScore > scored[j].QualityScore
    })
    
    return scored
}

// determineSpecialistType determines specialist type from intent
func (ms *ModelSelector) determineSpecialistType(intent *Intent) string {
    category := intent.Category
    
    switch category {
    case "coding":
        return "coding"
    case "research":
        return "research"
    case "writing":
        return "writing"
    default:
        return "coding" // Default
    }
}

// getSpeedScore converts speed string to numeric score
func (ms *ModelSelector) getSpeedScore(speed string) float64 {
    switch speed {
    case "very-fast":
        return 1.0
    case "fast":
        return 0.8
    case "medium":
        return 0.5
    case "slow":
        return 0.3
    default:
        return 0.5
    }
}
```

---

### 5. 3-Tier Memory Architecture

#### Mục tiêu
Phân tầng memory tối thiểu 3 lớp để tối ưu context retention.

#### Implementation

```go
// apps/router/layers/memory/tiered_memory.go

package memory

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/redis/go-redis/v9"
)

// MemoryTier represents a memory tier
type MemoryTier int

const (
    TierHot   MemoryTier = iota // Fast, small, in-memory
    TierWarm                    // Medium, Redis
    TierCold                    // Slow, large, persistent storage
)

// MemoryEntry represents a memory entry
type MemoryEntry struct {
    Key         string
    Value       interface{}
    CreatedAt   time.Time
    AccessedAt  time.Time
    AccessCount int
    Size        int64
    Tier        MemoryTier
    TTL         time.Duration
}

// TieredMemory implements 3-tier memory architecture
type TieredMemory struct {
    hotCache   *LRUCache      // Tier 1: In-memory LRU cache
    warmCache  *redis.Client  // Tier 2: Redis
    coldStore  *PersistentStore // Tier 3: Persistent storage (database)
    policy     *EvictionPolicy
    stats      *MemoryStats
}

// MemoryStats tracks memory statistics
type MemoryStats struct {
    HotHits       int64
    HotMisses     int64
    WarmHits      int64
    WarmMisses    int64
    ColdHits      int64
    ColdMisses    int64
    TotalSize     int64
    Evictions     int64
    Promotions    int64
    Demotions     int64
}

// NewTieredMemory creates a new tiered memory
func NewTieredMemory(redisAddr string, coldStore *PersistentStore) *TieredMemory {
    return &TieredMemory{
        hotCache:  NewLRUCache(1000),  // 1000 entries
        warmCache: redis.NewClient(&redis.Options{Addr: redisAddr}),
        coldStore: coldStore,
        policy:    NewEvictionPolicy(),
        stats:     &MemoryStats{},
    }
}

// Get retrieves value from memory (tries all tiers)
func (tm *TieredMemory) Get(ctx context.Context, key string) (interface{}, error) {
    // Tier 1: Hot cache (fastest)
    if value, ok := tm.hotCache.Get(key); ok {
        entry := value.(*MemoryEntry)
        entry.AccessedAt = time.Now()
        entry.AccessCount++
        tm.stats.HotHits++
        return entry.Value, nil
    }
    tm.stats.HotMisses++
    
    // Tier 2: Warm cache (Redis)
    data, err := tm.warmCache.Get(ctx, key).Bytes()
    if err == nil {
        var entry MemoryEntry
        if err := json.Unmarshal(data, &entry); err == nil {
            // Promote to hot cache
            entry.AccessedAt = time.Now()
            entry.AccessCount++
            tm.hotCache.Set(key, &entry, entry.TTL)
            tm.stats.WarmHits++
            tm.stats.Promotions++
            return entry.Value, nil
        }
    }
    tm.stats.WarmMisses++
    
    // Tier 3: Cold storage (persistent)
    entry, err := tm.coldStore.Get(ctx, key)
    if err == nil {
        // Promote to warm cache
        entry.Tier = TierWarm
        data, _ := json.Marshal(entry)
        tm.warmCache.Set(ctx, key, data, entry.TTL)
        tm.stats.ColdHits++
        tm.stats.Promotions++
        return entry.Value, nil
    }
    tm.stats.ColdMisses++
    
    return nil, fmt.Errorf("key not found")
}

// Set stores value in memory (determines tier)
func (tm *TieredMemory) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    entry := &MemoryEntry{
        Key:         key,
        Value:       value,
        CreatedAt:   time.Now(),
        AccessedAt:  time.Now(),
        AccessCount: 0,
        Tier:        tm.policy.DetermineInitialTier(value, ttl),
        TTL:         ttl,
    }
    
    // Store in appropriate tier
    switch entry.Tier {
    case TierHot:
        tm.hotCache.Set(key, entry, ttl)
    case TierWarm:
        data, _ := json.Marshal(entry)
        tm.warmCache.Set(ctx, key, data, ttl)
    case TierCold:
        tm.coldStore.Set(ctx, entry)
    }
    
    tm.stats.TotalSize += entry.Size
    
    return nil
}

// EvictionPolicy determines when to evict/promote/demote entries
type EvictionPolicy struct {
    HotMaxSize      int64
    WarmMaxSize     int64
    ColdMaxSize     int64
    HotTTL          time.Duration
    WarmTTL         time.Duration
    AccessThreshold int
}

// NewEvictionPolicy creates a new eviction policy
func NewEvictionPolicy() *EvictionPolicy {
    return &EvictionPolicy{
        HotMaxSize:      100 * 1024 * 1024,  // 100MB
        WarmMaxSize:     1 * 1024 * 1024 * 1024, // 1GB
        ColdMaxSize:     100 * 1024 * 1024 * 1024, // 100GB
        HotTTL:          5 * time.Minute,
        WarmTTL:         1 * time.Hour,
        AccessThreshold: 10,
    }
}

// DetermineInitialTier determines initial tier for entry
func (ep *EvictionPolicy) DetermineInitialTier(value interface{}, ttl time.Duration) MemoryTier {
    size := estimateSize(value)
    
    // Small, short-lived → Hot
    if size < 10*1024 && ttl < ep.HotTTL {
        return TierHot
    }
    
    // Medium, medium-lived → Warm
    if size < 100*1024 && ttl < ep.WarmTTL {
        return TierWarm
    }
    
    // Large, long-lived → Cold
    return TierCold
}

// ShouldEvict checks if entry should be evicted
func (ep *EvictionPolicy) ShouldEvict(entry *MemoryEntry, tier MemoryTier) bool {
    // Check TTL
    if time.Since(entry.CreatedAt) > entry.TTL {
        return true
    }
    
    // Check access count for hot tier
    if tier == TierHot && entry.AccessCount < ep.AccessThreshold {
        return true
    }
    
    return false
}
```

---

### 6. Tool-First Design + Tool Registry + Discovery

#### Mục tiêu
Tool-first architecture với dynamic tool registry và discovery.

#### Implementation

```go
// apps/router/layers/tools/tool_registry.go

package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "plugin"
    "reflect"
    "sync"
)

// Tool represents a tool definition
type Tool struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`
    Category    string            `json:"category"`
    Version     string            `json:"version"`
    Author      string            `json:"author"`
    
    // Execution
    Handler     ToolHandler       `json:"-"`
    Schema      ToolSchema        `json:"schema"`
    
    // Metadata
    Capabilities []string         `json:"capabilities"`
    Dependencies []string         `json:"dependencies"`
    Permissions  []string         `json:"permissions"`
    
    // Performance
    Latency      time.Duration    `json:"latency"`
    SuccessRate  float64          `json:"success_rate"`
    UsageCount   int64            `json:"usage_count"`
    
    // Registration
    RegisteredAt time.Time       `json:"registered_at"`
    LastUpdated  time.Time       `json:"last_updated"`
}

// ToolHandler is the function that executes the tool
type ToolHandler func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

// ToolSchema defines the tool's input/output schema
type ToolSchema struct {
    Input  map[string]interface{} `json:"input"`
    Output map[string]interface{} `json:"output"`
}

// ToolRegistry manages tool registration and discovery
type ToolRegistry struct {
    tools     map[string]*Tool
    plugins   map[string]*plugin.Plugin
    discovery *ToolDiscovery
    mu        sync.RWMutex
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
    return &ToolRegistry{
        tools:     make(map[string]*Tool),
        plugins:   make(map[string]*plugin.Plugin),
        discovery: NewToolDiscovery(),
    }
}

// Register registers a tool
func (tr *ToolRegistry) Register(tool *Tool) error {
    tr.mu.Lock()
    defer tr.mu.Unlock()
    
    // Validate tool
    if err := tr.validateTool(tool); err != nil {
        return fmt.Errorf("invalid tool: %w", err)
    }
    
    // Check if tool already exists
    if _, exists := tr.tools[tool.ID]; exists {
        return fmt.Errorf("tool %s already registered", tool.ID)
    }
    
    // Register tool
    tool.RegisteredAt = time.Now()
    tool.LastUpdated = time.Now()
    tr.tools[tool.ID] = tool
    
    return nil
}

// Unregister unregisters a tool
func (tr *ToolRegistry) Unregister(toolID string) error {
    tr.mu.Lock()
    defer tr.mu.Unlock()
    
    if _, exists := tr.tools[toolID]; !exists {
        return fmt.Errorf("tool %s not found", toolID)
    }
    
    delete(tr.tools, toolID)
    return nil
}

// Get retrieves a tool by ID
func (tr *ToolRegistry) Get(toolID string) (*Tool, error) {
    tr.mu.RLock()
    defer tr.mu.RUnlock()
    
    tool, exists := tr.tools[toolID]
    if !exists {
        return nil, fmt.Errorf("tool %s not found", toolID)
    }
    
    return tool, nil
}

// List lists all registered tools
func (tr *ToolRegistry) List() []*Tool {
    tr.mu.RLock()
    defer tr.mu.RUnlock()
    
    tools := make([]*Tool, 0, len(tr.tools))
    for _, tool := range tr.tools {
        tools = append(tools, tool)
    }
    
    return tools
}

// Search searches for tools by query
func (tr *ToolRegistry) Search(query string) []*Tool {
    tr.mu.RLock()
    defer tr.mu.RUnlock()
    
    var results []*Tool
    query = strings.ToLower(query)
    
    for _, tool := range tr.tools {
        if strings.Contains(strings.ToLower(tool.Name), query) ||
           strings.Contains(strings.ToLower(tool.Description), query) ||
           strings.Contains(strings.ToLower(tool.Category), query) {
            results = append(results, tool)
        }
    }
    
    return results
}

// Discover discovers tools from external sources
func (tr *ToolRegistry) Discover(ctx context.Context, source string) ([]*Tool, error) {
    return tr.discovery.Discover(ctx, source)
}

// LoadPlugin loads a tool from a plugin
func (tr *ToolRegistry) LoadPlugin(path string) error {
    tr.mu.Lock()
    defer tr.mu.Unlock()
    
    // Load plugin
    p, err := plugin.Open(path)
    if err != nil {
        return fmt.Errorf("failed to load plugin: %w", err)
    }
    
    // Lookup tool function
    toolFunc, err := p.Lookup("Tool")
    if err != nil {
        return fmt.Errorf("failed to lookup Tool function: %w", err)
    }
    
    // Call tool function
    toolPtr, err := toolFunc.(func() *Tool)()
    if err != nil {
        return fmt.Errorf("failed to call Tool function: %w", err)
    }
    
    // Register tool
    if err := tr.Register(toolPtr); err != nil {
        return err
    }
    
    tr.plugins[path] = p
    return nil
}

// Execute executes a tool
func (tr *ToolRegistry) Execute(ctx context.Context, toolID string, input map[string]interface{}) (map[string]interface{}, error) {
    tool, err := tr.Get(toolID)
    if err != nil {
        return nil, err
    }
    
    // Validate input against schema
    if err := tr.validateInput(tool, input); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }
    
    // Execute tool
    start := time.Now()
    output, err := tool.Handler(ctx, input)
    latency := time.Since(start)
    
    // Update stats
    tool.UsageCount++
    tool.Latency = latency
    if err == nil {
        // Update success rate with exponential moving average
        tool.SuccessRate = 0.9*tool.SuccessRate + 0.1*1.0
    } else {
        tool.SuccessRate = 0.9*tool.SuccessRate + 0.1*0.0
    }
    
    return output, err
}

// validateTool validates a tool definition
func (tr *ToolRegistry) validateTool(tool *Tool) error {
    if tool.ID == "" {
        return fmt.Errorf("tool ID is required")
    }
    if tool.Name == "" {
        return fmt.Errorf("tool name is required")
    }
    if tool.Handler == nil {
        return fmt.Errorf("tool handler is required")
    }
    return nil
}

// validateInput validates input against schema
func (tr *ToolRegistry) validateInput(tool *Tool, input map[string]interface{}) error {
    // TODO: Implement schema validation
    return nil
}

// ToolDiscovery discovers tools from external sources
type ToolDiscovery struct {
    sources map[string]DiscoverySource
}

// DiscoverySource is an interface for tool discovery sources
type DiscoverySource interface {
    Discover(ctx context.Context) ([]*Tool, error)
}

// NewToolDiscovery creates a new tool discovery
func NewToolDiscovery() *ToolDiscovery {
    return &ToolDiscovery{
        sources: make(map[string]DiscoverySource),
    }
}

// RegisterSource registers a discovery source
func (td *ToolDiscovery) RegisterSource(name string, source DiscoverySource) {
    td.sources[name] = source
}

// Discover discovers tools from a source
func (td *ToolDiscovery) Discover(ctx context.Context, source string) ([]*Tool, error) {
    s, exists := td.sources[source]
    if !exists {
        return nil, fmt.Errorf("discovery source %s not found", source)
    }
    
    return s.Discover(ctx)
}
```

---

### 7. Dynamic Context Compression

#### Mục tiêu
Tối ưu context window utilization với dynamic compression.

#### Implementation

```go
// apps/router/layers/context/compression.go

package context

import (
    "context"
    "strings"
)

// CompressionStrategy defines compression behavior
type CompressionStrategy struct {
    EnableSummarization    bool
    EnableDeduplication    bool
    EnablePruning         bool
    EnableTokenOptimization bool
    TargetTokenCount      int
    MinRetentionScore     float64
}

// ContextCompressor implements dynamic context compression
type ContextCompressor struct {
    strategy *CompressionStrategy
    summarizer *Summarizer
    deduplicator *Deduplicator
    pruner     *Pruner
}

// NewContextCompressor creates a new context compressor
func NewContextCompressor(strategy *CompressionStrategy) *ContextCompressor {
    return &ContextCompressor{
        strategy:    strategy,
        summarizer:  NewSummarizer(),
        deduplicator: NewDeduplicator(),
        pruner:      NewPruner(),
    }
}

// Compress compresses context to fit within target token count
func (cc *ContextCompressor) Compress(ctx context.Context, messages []Message) ([]Message, error) {
    currentTokens := cc.estimateTokens(messages)
    
    if currentTokens <= cc.strategy.TargetTokenCount {
        return messages, nil
    }
    
    compressed := messages
    
    // Step 1: Deduplication
    if cc.strategy.EnableDeduplication {
        compressed = cc.deduplicator.Deduplicate(compressed)
        currentTokens = cc.estimateTokens(compressed)
    }
    
    // Step 2: Pruning low-value content
    if cc.strategy.EnablePruning && currentTokens > cc.strategy.TargetTokenCount {
        compressed = cc.pruner.Prune(compressed, cc.strategy.MinRetentionScore)
        currentTokens = cc.estimateTokens(compressed)
    }
    
    // Step 3: Summarization
    if cc.strategy.EnableSummarization && currentTokens > cc.strategy.TargetTokenCount {
        compressed, err := cc.summarizer.Summarize(ctx, compressed, currentTokens-cc.strategy.TargetTokenCount)
        if err != nil {
            return nil, err
        }
        currentTokens = cc.estimateTokens(compressed)
    }
    
    // Step 4: Token optimization
    if cc.strategy.EnableTokenOptimization {
        compressed = cc.optimizeTokens(compressed)
    }
    
    return compressed, nil
}

// estimateTokens estimates token count for messages
func (cc *ContextCompressor) estimateTokens(messages []Message) int {
    total := 0
    for _, msg := range messages {
        total += len(msg.Content) / 4 // Rough estimate: 1 token ≈ 4 chars
    }
    return total
}

// optimizeTokens optimizes token usage
func (cc *ContextCompressor) optimizeTokens(messages []Message) []Message {
    optimized := make([]Message, len(messages))
    
    for i, msg := range messages {
        optimized[i] = Message{
            Role:    msg.Role,
            Content: cc.optimizeString(msg.Content),
        }
    }
    
    return optimized
}

// optimizeString optimizes string for token usage
func (cc *ContextCompressor) optimizeString(s string) string {
    // Remove extra whitespace
    s = strings.Join(strings.Fields(s), " ")
    
    // Remove redundant punctuation
    s = strings.ReplaceAll(s, "...", ".")
    
    return s
}
```

---

### 8. Personal + Long-term Preference

#### Mục tiêu
Personalization system để lưu trữ và áp dụng user preferences.

#### Implementation

```go
// apps/router/layers/personalization/preferences.go

package personalization

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/redis/go-redis/v9"
)

// UserPreferences represents user preferences
type UserPreferences struct {
    UserID          string            `json:"user_id"`
    
    // Model preferences
    PreferredModels map[string]string `json:"preferred_models"` // category -> model
    ModelSettings   map[string]interface{} `json:"model_settings"`
    
    // Tool preferences
    FavoriteTools   []string          `json:"favorite_tools"`
    ToolSettings    map[string]interface{} `json:"tool_settings"`
    
    // Behavior preferences
    ResponseStyle   string            `json:"response_style"` // concise, detailed, balanced
    Language        string            `json:"language"`
    TimeZone        string            `json:"time_zone"`
    
    // Learning preferences
    LearningRate    float64           `json:"learning_rate"`
    FeedbackEnabled bool             `json:"feedback_enabled"`
    
    // Metadata
    CreatedAt       time.Time         `json:"created_at"`
    UpdatedAt       time.Time         `json:"updated_at"`
}

// PreferenceManager manages user preferences
type PreferenceManager struct {
    redis *redis.Client
    local *LRUCache
}

// NewPreferenceManager creates a new preference manager
func NewPreferenceManager(redisAddr string) *PreferenceManager {
    return &PreferenceManager{
        redis: redis.NewClient(&redis.Options{Addr: redisAddr}),
        local: NewLRUCache(1000),
    }
}

// Get retrieves user preferences
func (pm *PreferenceManager) Get(ctx context.Context, userID string) (*UserPreferences, error) {
    // Check local cache
    if cached, ok := pm.local.Get(userID); ok {
        return cached.(*UserPreferences), nil
    }
    
    // Check Redis
    key := fmt.Sprintf("preferences:%s", userID)
    data, err := pm.redis.Get(ctx, key).Bytes()
    if err != nil {
        // Return default preferences
        return pm.defaultPreferences(userID), nil
    }
    
    var prefs UserPreferences
    if err := json.Unmarshal(data, &prefs); err != nil {
        return pm.defaultPreferences(userID), nil
    }
    
    // Cache locally
    pm.local.Set(userID, &prefs, time.Hour)
    
    return &prefs, nil
}

// Set saves user preferences
func (pm *PreferenceManager) Set(ctx context.Context, prefs *UserPreferences) error {
    prefs.UpdatedAt = time.Now()
    
    // Save to Redis
    key := fmt.Sprintf("preferences:%s", prefs.UserID)
    data, _ := json.Marshal(prefs)
    if err := pm.redis.Set(ctx, key, data, 0).Err(); err != nil {
        return err
    }
    
    // Update local cache
    pm.local.Set(prefs.UserID, prefs, time.Hour)
    
    return nil
}

// Update updates specific preference fields
func (pm *PreferenceManager) Update(ctx context.Context, userID string, updates map[string]interface{}) error {
    prefs, err := pm.Get(ctx, userID)
    if err != nil {
        return err
    }
    
    // Apply updates
    for key, value := range updates {
        switch key {
        case "preferred_models":
            if v, ok := value.(map[string]string); ok {
                prefs.PreferredModels = v
            }
        case "response_style":
            if v, ok := value.(string); ok {
                prefs.ResponseStyle = v
            }
        // ... other fields
        }
    }
    
    return pm.Set(ctx, prefs)
}

// defaultPreferences returns default preferences for a user
func (pm *PreferenceManager) defaultPreferences(userID string) *UserPreferences {
    return &UserPreferences{
        UserID:          userID,
        PreferredModels: make(map[string]string),
        ModelSettings:   make(map[string]interface{}),
        FavoriteTools:   []string{},
        ToolSettings:    make(map[string]interface{}),
        ResponseStyle:   "balanced",
        Language:        "en",
        TimeZone:        "UTC",
        LearningRate:    0.1,
        FeedbackEnabled: true,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }
}
```

---

### 9. Session Continuity

#### Mục tiêu
Maintain session state across multiple turns.

#### Implementation

```go
// apps/router/layers/session/continuity.go

package session

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/redis/go-redis/v9"
)

// Session represents a user session
type Session struct {
    ID           string            `json:"id"`
    UserID       string            `json:"user_id"`
    
    // Conversation state
    Messages     []Message         `json:"messages"`
    Context      map[string]interface{} `json:"context"`
    
    // Tool state
    ToolStates   map[string]interface{} `json:"tool_states"`
    
    // Metadata
    CreatedAt    time.Time         `json:"created_at"`
    UpdatedAt    time.Time         `json:"updated_at"`
    LastActivity time.Time         `json:"last_activity"`
    TTL          time.Duration     `json:"ttl"`
}

// SessionManager manages sessions
type SessionManager struct {
    redis *redis.Client
    local *LRUCache
}

// NewSessionManager creates a new session manager
func NewSessionManager(redisAddr string) *SessionManager {
    return &SessionManager{
        redis: redis.NewClient(&redis.Options{Addr: redisAddr}),
        local: NewLRUCache(100),
    }
}

// Create creates a new session
func (sm *SessionManager) Create(ctx context.Context, userID string, ttl time.Duration) (*Session, error) {
    sessionID := generateSessionID()
    
    session := &Session{
        ID:           sessionID,
        UserID:       userID,
        Messages:     []Message{},
        Context:      make(map[string]interface{}),
        ToolStates:   make(map[string]interface{}),
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
        LastActivity: time.Now(),
        TTL:          ttl,
    }
    
    // Save to Redis
    if err := sm.Save(ctx, session); err != nil {
        return nil, err
    }
    
    return session, nil
}

// Get retrieves a session
func (sm *SessionManager) Get(ctx context.Context, sessionID string) (*Session, error) {
    // Check local cache
    if cached, ok := sm.local.Get(sessionID); ok {
        session := cached.(*Session)
        session.LastActivity = time.Now()
        return session, nil
    }
    
    // Check Redis
    key := fmt.Sprintf("session:%s", sessionID)
    data, err := sm.redis.Get(ctx, key).Bytes()
    if err != nil {
        return nil, fmt.Errorf("session not found")
    }
    
    var session Session
    if err := json.Unmarshal(data, &session); err != nil {
        return nil, fmt.Errorf("failed to unmarshal session")
    }
    
    session.LastActivity = time.Now()
    
    // Cache locally
    sm.local.Set(sessionID, &session, session.TTL)
    
    return &session, nil
}

// Save saves a session
func (sm *SessionManager) Save(ctx context.Context, session *Session) error {
    session.UpdatedAt = time.Now()
    session.LastActivity = time.Now()
    
    // Save to Redis
    key := fmt.Sprintf("session:%s", session.ID)
    data, _ := json.Marshal(session)
    if err := sm.redis.Set(ctx, key, data, session.TTL).Err(); err != nil {
        return err
    }
    
    // Update local cache
    sm.local.Set(session.ID, session, session.TTL)
    
    return nil
}

// AddMessage adds a message to the session
func (sm *SessionManager) AddMessage(ctx context.Context, sessionID string, message Message) error {
    session, err := sm.Get(ctx, sessionID)
    if err != nil {
        return err
    }
    
    session.Messages = append(session.Messages, message)
    
    return sm.Save(ctx, session)
}

// UpdateContext updates session context
func (sm *SessionManager) UpdateContext(ctx context.Context, sessionID string, key string, value interface{}) error {
    session, err := sm.Get(ctx, sessionID)
    if err != nil {
        return err
    }
    
    session.Context[key] = value
    
    return sm.Save(ctx, session)
}

// UpdateToolState updates tool state
func (sm *SessionManager) UpdateToolState(ctx context.Context, sessionID string, toolID string, state interface{}) error {
    session, err := sm.Get(ctx, sessionID)
    if err != nil {
        return err
    }
    
    session.ToolStates[toolID] = state
    
    return sm.Save(ctx, session)
}
```

---

### 10. Agent Step Tracing

#### Mục tiêu
Trace mọi agent step cho debugging và observability.

#### Implementation

```go
// apps/router/layers/tracing/step_tracer.go

package tracing

import (
    "context"
    "encoding/json"
    "time"
)

// StepType represents the type of step
type StepType string

const (
    StepTypeRouting      StepType = "routing"
    StepTypeToolCall     StepType = "tool_call"
    StepTypeLLMRequest   StepType = "llm_request"
    StepTypeMemory       StepType = "memory"
    StepTypeCache        StepType = "cache"
    StepTypeError        StepType = "error"
)

// Step represents an agent execution step
type Step struct {
    ID           string            `json:"id"`
    SessionID    string            `json:"session_id"`
    Type         StepType          `json:"type"`
    Timestamp    time.Time         `json:"timestamp"`
    Duration     time.Duration     `json:"duration"`
    
    // Input/Output
    Input        interface{}       `json:"input"`
    Output       interface{}       `json:"output"`
    
    // Metadata
    Metadata     map[string]string `json:"metadata"`
    ParentID     string            `json:"parent_id,omitempty"`
    Error        string            `json:"error,omitempty"`
    
    // Performance
    TokensUsed   int               `json:"tokens_used,omitempty"`
    Cost         float64           `json:"cost,omitempty"`
}

// StepTracer traces agent execution steps
type StepTracer struct {
    storage    StepStorage
    buffer     []*Step
    bufferSize int
    flushInterval time.Duration
}

// StepStorage is an interface for step storage
type StepStorage interface {
    Store(ctx context.Context, step *Step) error
    Query(ctx context.Context, query StepQuery) ([]*Step, error)
}

// NewStepTracer creates a new step tracer
func NewStepTracer(storage StepStorage) *StepTracer {
    return &StepTracer{
        storage:    storage,
        buffer:     make([]*Step, 0, 100),
        bufferSize: 100,
        flushInterval: 5 * time.Second,
    }
}

// Start starts a new step
func (st *StepTracer) Start(ctx context.Context, sessionID string, stepType StepType, input interface{}) *Step {
    step := &Step{
        ID:        generateStepID(),
        SessionID: sessionID,
        Type:      stepType,
        Timestamp: time.Now(),
        Input:     input,
        Metadata:  make(map[string]string),
    }
    
    return step
}

// End ends a step
func (st *StepTracer) End(ctx context.Context, step *Step, output interface{}, err error) {
    step.Duration = time.Since(step.Timestamp)
    step.Output = output
    
    if err != nil {
        step.Error = err.Error()
    }
    
    // Buffer step
    st.buffer = append(st.buffer, step)
    
    // Flush if buffer is full
    if len(st.buffer) >= st.bufferSize {
        st.flush(ctx)
    }
}

// flush flushes buffered steps to storage
func (st *StepTracer) flush(ctx context.Context) error {
    for _, step := range st.buffer {
        if err := st.storage.Store(ctx, step); err != nil {
            // Log error but continue
            log.Printf("Failed to store step: %v", err)
        }
    }
    
    st.buffer = st.buffer[:0]
    return nil
}

// Query queries steps
func (st *StepTracer) Query(ctx context.Context, query StepQuery) ([]*Step, error) {
    // Flush buffer first
    st.flush(ctx)
    
    return st.storage.Query(ctx, query)
}

// StepQuery represents a step query
type StepQuery struct {
    SessionID string
    Type      StepType
    StartTime time.Time
    EndTime   time.Time
    Limit     int
}
```

---

### 11. Automatic Eval Loop

#### Mục tiêu
Self-evaluation loop để continuous improvement.

#### Implementation

```go
// apps/router/layers/eval/eval_loop.go

package eval

import (
    "context"
    "time"
)

// EvalCriteria represents evaluation criteria
type EvalCriteria struct {
    Accuracy       float64 // Target accuracy
    Latency        time.Duration // Max latency
    Cost           float64 // Max cost
    UserSatisfaction float64 // Target satisfaction
}

// EvalResult represents evaluation result
type EvalResult struct {
    SessionID      string
    Criteria       *EvalCriteria
    Scores         map[string]float64
    Passed         bool
    Feedback       string
    Recommendations []string
    Timestamp      time.Time
}

// EvalLoop implements automatic evaluation loop
type EvalLoop struct {
    evaluator     *Evaluator
    criteria      *EvalCriteria
    interval      time.Duration
    storage       EvalStorage
}

// NewEvalLoop creates a new eval loop
func NewEvalLoop(evaluator *Evaluator, criteria *EvalCriteria, interval time.Duration) *EvalLoop {
    return &EvalLoop{
        evaluator: evaluator,
        criteria:  criteria,
        interval:  interval,
    }
}

// Start starts the eval loop
func (el *EvalLoop) Start(ctx context.Context) {
    ticker := time.NewTicker(el.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            el.runEvaluation(ctx)
        case <-ctx.Done():
            return
        }
    }
}

// runEvaluation runs a single evaluation
func (el *EvalLoop) runEvaluation(ctx context.Context) {
    // Get recent sessions
    sessions, err := el.storage.GetRecentSessions(ctx, 100)
    if err != nil {
        log.Printf("Failed to get recent sessions: %v", err)
        return
    }
    
    // Evaluate each session
    for _, session := range sessions {
        result, err := el.evaluator.Evaluate(ctx, session, el.criteria)
        if err != nil {
            log.Printf("Failed to evaluate session %s: %v", session.ID, err)
            continue
        }
        
        // Store result
        if err := el.storage.StoreEvalResult(ctx, result); err != nil {
            log.Printf("Failed to store eval result: %v", err)
        }
        
        // Apply recommendations if failed
        if !result.Passed {
            el.applyRecommendations(ctx, result.Recommendations)
        }
    }
}

// applyRecommendations applies evaluation recommendations
func (el *EvalLoop) applyRecommendations(ctx context.Context, recommendations []string) {
    for _, rec := range recommendations {
        switch rec {
        case "increase_model_quality":
            // Increase model quality for future requests
        case "decrease_latency":
            // Switch to faster model
        case "improve_caching":
            // Improve caching strategy
        // ... other recommendations
        }
    }
}

// Evaluator evaluates sessions
type Evaluator struct {
    tracer *tracing.StepTracer
}

// Evaluate evaluates a session
func (e *Evaluator) Evaluate(ctx context.Context, session *Session, criteria *EvalCriteria) (*EvalResult, error) {
    result := &EvalResult{
        SessionID: session.ID,
        Criteria:  criteria,
        Scores:    make(map[string]float64),
        Timestamp: time.Now(),
    }
    
    // Get steps for session
    steps, err := e.tracer.Query(ctx, tracing.StepQuery{
        SessionID: session.ID,
    })
    if err != nil {
        return nil, err
    }
    
    // Calculate scores
    result.Scores["accuracy"] = e.calculateAccuracy(steps)
    result.Scores["latency"] = e.calculateLatencyScore(steps, criteria.Latency)
    result.Scores["cost"] = e.calculateCostScore(steps, criteria.Cost)
    result.Scores["satisfaction"] = e.calculateSatisfaction(session)
    
    // Determine if passed
    result.Passed = e.checkPassed(result.Scores, criteria)
    
    // Generate recommendations
    if !result.Passed {
        result.Recommendations = e.generateRecommendations(result.Scores, criteria)
    }
    
    return result, nil
}
```

---

### 12. Tool Registry + Discovery (Enhanced)

#### Mục tiêu
Dynamic tool discovery từ multiple sources.

#### Implementation

```go
// apps/router/layers/tools/discovery_sources.go

package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
)

// FileSystemSource discovers tools from filesystem
type FileSystemSource struct {
    basePath string
}

// NewFileSystemSource creates a new filesystem source
func NewFileSystemSource(basePath string) *FileSystemSource {
    return &FileSystemSource{
        basePath: basePath,
    }
}

// Discover discovers tools from filesystem
func (fs *FileSystemSource) Discover(ctx context.Context) ([]*Tool, error) {
    var tools []*Tool
    
    err := filepath.Walk(fs.basePath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if !info.IsDir() && filepath.Ext(path) == ".json" {
            data, err := ioutil.ReadFile(path)
            if err != nil {
                return err
            }
            
            var tool Tool
            if err := json.Unmarshal(data, &tool); err != nil {
                return err
            }
            
            tools = append(tools, &tool)
        }
        
        return nil
    })
    
    return tools, err
}

// HTTPSource discovers tools from HTTP endpoint
type HTTPSource struct {
    endpoint string
    headers  map[string]string
}

// NewHTTPSource creates a new HTTP source
func NewHTTPSource(endpoint string, headers map[string]string) *HTTPSource {
    return &HTTPSource{
        endpoint: endpoint,
        headers:  headers,
    }
}

// Discover discovers tools from HTTP endpoint
func (hs *HTTPSource) Discover(ctx context.Context) ([]*Tool, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", hs.endpoint, nil)
    if err != nil {
        return nil, err
    }
    
    for key, value := range hs.headers {
        req.Header.Set(key, value)
    }
    
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("HTTP request failed: %s", resp.Status)
    }
    
    var tools []*Tool
    if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
        return nil, err
    }
    
    return tools, nil
}

// MCPSource discovers tools from MCP servers
type MCPSource struct {
    mcpClient *MCPClient
}

// NewMCPSource creates a new MCP source
func NewMCPSource(mcpClient *MCPClient) *MCPSource {
    return &MCPSource{
        mcpClient: mcpClient,
    }
}

// Discover discovers tools from MCP servers
func (ms *MCPSource) Discover(ctx context.Context) ([]*Tool, error) {
    // List tools from MCP server
    tools, err := ms.mcpClient.ListTools(ctx)
    if err != nil {
        return nil, err
    }
    
    // Convert MCP tools to internal tool format
    var internalTools []*Tool
    for _, mcpTool := range tools {
        internalTool := &Tool{
            ID:          mcpTool.Name,
            Name:        mcpTool.Name,
            Description: mcpTool.Description,
            Schema: ToolSchema{
                Input:  mcpTool.InputSchema,
                Output: map[string]interface{}{},
            },
            Handler:     ms.mcpToolHandler(mcpTool),
        }
        internalTools = append(internalTools, internalTool)
    }
    
    return internalTools, nil
}

// mcpToolHandler creates a handler for MCP tool
func (ms *MCPSource) mcpToolHandler(mcpTool MCPTool) ToolHandler {
    return func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
        return ms.mcpClient.CallTool(ctx, mcpTool.Name, input)
    }
}
```

---

## 📋 Implementation Phases

### Phase 1: Foundation (Week 1-2)
- [ ] Implement Semantic Routing
- [ ] Implement Prompt Caching
- [ ] Implement Adaptive Timeout + Retry
- [ ] Unit tests for all components

### Phase 2: Intelligence (Week 3-4)
- [ ] Implement Auto Model Selection
- [ ] Implement 3-Tier Memory
- [ ] Implement Tool Registry + Discovery
- [ ] Integration tests

### Phase 3: Optimization (Week 5-6)
- [ ] Implement Dynamic Context Compression
- [ ] Implement Personal + Long-term Preference
- [ ] Implement Session Continuity
- [ ] Performance benchmarks

### Phase 4: Observability (Week 7-8)
- [ ] Implement Agent Step Tracing
- [ ] Implement Automatic Eval Loop
- [ ] Dashboard for monitoring
- [ ] End-to-end tests

### Phase 5: Integration (Week 9-10)
- [ ] Integrate all components with existing router
- [ ] Migration from old routing to new routing
- [ ] Documentation
- [ ] Training

---

## 🎯 Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Routing Accuracy | 90%+ | Semantic routing vs rule-based |
| Cache Hit Rate | 40%+ | Prompt caching efficiency |
| Cost Reduction | 30%+ | Before/after comparison |
| Latency Improvement | 20%+ | Average response time |
| Memory Efficiency | 50%+ | Context window utilization |
| Tool Discovery Time | <1s | Dynamic tool registration |
| Session Continuity | 95%+ | Multi-turn success rate |
| Trace Coverage | 100% | All steps traced |
| Eval Loop Frequency | Hourly | Automatic evaluation |
| Overall Satisfaction | 4.5/5 | User feedback |

---

## 🚀 Next Steps

1. **Review Blueprint** - Get approval from team
2. **Create Sprint Plan** - Break down into sprints
3. **Set Up Infrastructure** - Redis, databases, monitoring
4. **Start Phase 1** - Begin with foundation components
5. **Iterate** - Continuous improvement based on feedback

---

**Status**: Blueprint Complete  
**Next**: Review and approval  
**Author**: Devin AI Agent  
**Date**: 2026-05-04
