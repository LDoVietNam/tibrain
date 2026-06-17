# Semantic Routing Implementation - Ti Router Phase 2

> **Ngày tạo**: 2026-05-04  
> **Mục đích**: Tài liệu implementation cho Semantic Routing trong Ti Router  
> **Trạng thái**: In Progress

---

## 1. Tổng quan Semantic Routing

### 1.1 Định nghĩa
**Semantic Routing** là kỹ thuật phân loại intent của user query dựa trên semantic meaning (nghĩa ngữ nghĩa) thay vì keyword matching đơn thuần, sau đó route đến model/tool phù hợp.

### 1.2 Tại sao cần Semantic Routing?

**Vấn đề với Single Model Approach:**
- Tất cả queries → cùng một model → không tối ưu cost
- Complex tasks cần model mạnh (Opus, GPT-4)
- Simple tasks có thể dùng model rẻ hơn (Haiku, GPT-4o-mini)
- Cost saving: 60-80% khi dùng model phù hợp

**Giải pháp Semantic Routing:**
```
User Query → Intent Classification → Model Selection → LLM Processing
```

---

## 2. Intent Classification Strategies

### 2.1 Strategy 1: Keyword-based (Fast Path)
```go
// Ưu điểm: Nhanh, không cost
// Nhược điểm: Không handle edge cases
func (ic *IntentClassifier) KeywordMatch(query string) *Intent {
    keywords := map[string]Intent{
        "code": IntentCoding,
        "debug": IntentCoding,
        "function": IntentCoding,
        "research": IntentResearch,
        "find": IntentResearch,
        "search": IntentResearch,
        "write": IntentWriting,
        "draft": IntentWriting,
        "chat": IntentChat,
        "hello": IntentChat,
    }
    
    for keyword, intent := range keywords {
        if strings.Contains(strings.ToLower(query), keyword) {
            return &intent
        }
    }
    return nil
}
```

### 2.2 Strategy 2: LLM-based Classification (Flexible)
```go
// Ưu điểm: Flexible, handle edge cases
// Nhược điểm: Higher latency, cost per query
func (ic *IntentClassifier) LLMClassify(query string) Intent {
    prompt := fmt.Sprintf(`
Classify this query into one of these categories:
- CODING: Code generation, debugging, refactoring
- RESEARCH: Information retrieval, analysis, investigation
- WRITING: Content creation, drafting, editing
- CHAT: Casual conversation, greetings

Query: %s

Output only the category name (CODING/RESEARCH/WRITING/CHAT).
`, query)
    
    response := ic.llm.Complete(prompt)
    return parseIntent(response)
}
```

### 2.3 Strategy 3: Hybrid Approach (Recommended)
```go
func (ic *IntentClassifier) Classify(query string) Intent {
    // Fast path: Keyword matching
    if intent := ic.KeywordMatch(query); intent != nil {
        log.Printf("[IntentClassifier] Keyword match: %s", *intent)
        return *intent
    }
    
    // Fallback: LLM classification
    intent := ic.LLMClassify(query)
    log.Printf("[IntentClassifier] LLM classification: %s", intent)
    return intent
}
```

---

## 3. Intent-to-Model Mapping

### 3.1 Model Selection Strategy

```go
type IntentToModelMapping struct {
    Coding   string // "claude-opus" hoặc "gpt-4"
    Research string // "claude-haiku" hoặc "gpt-4o-mini"
    Writing  string // "claude-sonnet" hoặc "gpt-4"
    Chat     string // "claude-haiku" hoặc "gpt-4o-mini"
}

func DefaultModelMapping() IntentToModelMapping {
    return IntentToModelMapping{
        Coding:   "claude-opus",   // Complex reasoning
        Research: "claude-haiku",  // Fast, cost-effective
        Writing:  "claude-sonnet", // Balanced
        Chat:     "claude-haiku",  // Fast, conversational
    }
}
```

### 3.2 Provider Selection

**Ti Router Available Providers:**
- cerebras: Fast inference
- deepseek: Cost-effective
- gemini: Google models
- groq: Fast inference
- openrouter: Multi-model access
- windsurf: Claude models
- xai: Grok models

**Mapping Example:**
```go
func (m *IntentToModelMapping) GetModel(intent Intent, provider string) string {
    switch intent {
    case IntentCoding:
        return provider + "/" + m.Coding
    case IntentResearch:
        return provider + "/" + m.Research
    case IntentWriting:
        return provider + "/" + m.Writing
    case IntentChat:
        return provider + "/" + m.Chat
    default:
        return provider + "/" + m.Chat // Default
    }
}
```

---

## 4. Semantic Routing Middleware

### 4.1 Middleware Architecture

```go
type SemanticRoutingMiddleware struct {
    intentClassifier *IntentClassifier
    modelMapping     *IntentToModelMapping
    defaultModel     string
}

func NewSemanticRoutingMiddleware(
    classifier *IntentClassifier,
    mapping *IntentToModelMapping,
    defaultModel string,
) *SemanticRoutingMiddleware {
    return &SemanticRoutingMiddleware{
        intentClassifier: classifier,
        modelMapping:     mapping,
        defaultModel:     defaultModel,
    }
}

func (m *SemanticRoutingMiddleware) RouteRequest(reqBody map[string]interface{}) (string, error) {
    // Extract query
    query := extractQuery(reqBody)
    
    // Classify intent
    intent := m.intentClassifier.Classify(query)
    
    // Select model
    provider := extractProvider(reqBody)
    selectedModel := m.modelMapping.GetModel(intent, provider)
    
    log.Printf("[SemanticRouting] Intent=%s → Model=%s", intent, selectedModel)
    
    // Update request body with selected model
    reqBody["model"] = selectedModel
    
    return selectedModel, nil
}
```

### 4.2 Integration vào Chat Handler

```go
// Trong cmd/routerd/handlers/chat/handlers.go
func (h *Handler) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
    // ... existing code ...
    
    // Semantic Routing
    if h.semanticRoutingMiddleware != nil {
        selectedModel, err := h.semanticRoutingMiddleware.RouteRequest(reqBody)
        if err != nil {
            log.Printf("[%s] Semantic routing failed: %v", requestID, err)
            // Fallback to default model
        } else {
            modelID = selectedModel
            log.Printf("[%s] Semantic routing selected model: %s", requestID, modelID)
        }
    }
    
    // ... continue with existing code ...
}
```

---

## 5. Configuration

### 5.1 YAML Config Structure

```yaml
# configs/Tiserverrouter.yaml
semantic_routing:
  enabled: true
  strategy: hybrid  # keyword, llm, hybrid
  fallback_model: "claude-3-5-sonnet-20241022"
  
  intent_to_model:
    coding: "claude-opus-4-20250514"
    research: "claude-3-5-haiku-20241022"
    writing: "claude-3-5-sonnet-20241022"
    chat: "claude-3-5-haiku-20241022"
  
  llm_classifier:
    model: "claude-3-5-haiku-20241022"
    provider: "windsurf"
    max_tokens: 50
```

### 5.2 Go Config Struct

```go
// cmd/routerd/appinit/init.go
type SemanticRoutingConfig struct {
    Enabled      bool                   `yaml:"enabled"`
    Strategy     string                 `yaml:"strategy"`
    FallbackModel string                 `yaml:"fallback_model"`
    IntentToModel IntentToModelMapping   `yaml:"intent_to_model"`
    LLMClassifier LLMClassifierConfig   `yaml:"llm_classifier"`
}

type LLMClassifierConfig struct {
    Model      string `yaml:"model"`
    Provider   string `yaml:"provider"`
    MaxTokens  int    `yaml:"max_tokens"`
}
```

---

## 6. HTTP Endpoints

### 6.1 Testing Intent Classification

```go
// cmd/routerd/main.go
mux.HandleFunc("/api/semantic/classify", handleIntentClassification)

func handleIntentClassification(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("query")
    if query == "" {
        http.Error(w, "query parameter required", http.StatusBadRequest)
        return
    }
    
    intent := semanticClassifier.Classify(query)
    model := modelMapping.GetModel(intent, defaultProvider)
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "query":  query,
        "intent": intent,
        "model":  model,
    })
}
```

### 6.2 Updating Model Mappings

```go
mux.HandleFunc("/api/semantic/mappings", handleModelMappings)

func handleModelMappings(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(modelMapping)
}
```

---

## 7. Testing Strategy

### 7.1 Test Cases per Intent

```go
// Test queries cho mỗi intent type
testCases := map[Intent][]string{
    IntentCoding: {
        "Write a function to sort an array",
        "Debug this code: it's throwing null pointer",
        "Refactor this class to use dependency injection",
    },
    IntentResearch: {
        "Find information about quantum computing",
        "Research the history of the Roman Empire",
        "What are the latest developments in AI?",
    },
    IntentWriting: {
        "Write a blog post about climate change",
        "Draft an email to my boss about the project",
        "Create a summary of this document",
    },
    IntentChat: {
        "Hello, how are you?",
        "What's the weather like today?",
        "Tell me a joke",
    },
}
```

### 7.2 Validation Checklist

- [ ] Keyword matching works cho common queries
- [ ] LLM fallback works cho edge cases
- [ ] Model selection correct cho mỗi intent
- [ ] Fallback to default model khi classification fail
- [ ] Logging captures classification decisions
- [ ] HTTP endpoints work cho testing
- [ ] Config-based model mapping works
- [ ] No breaking changes cho existing requests

---

## 8. Performance Considerations

### 8.1 Latency Impact

```
Keyword-only: ~0ms (instant)
LLM classification: ~100-500ms (network roundtrip)
Hybrid: ~0-500ms (depends on query)
```

**Optimization:**
- Cache classification results cho similar queries
- Use fast model (Haiku) cho LLM classification
- Async classification (non-blocking)
- Pre-classify common queries

### 8.2 Cost Impact

```
Keyword-only: $0 per query
LLM classification: ~$0.0001 per query (Haiku)
Hybrid: ~$0-$0.0001 per query (depends)
```

**ROI:**
- Coding queries: Use Opus → 10x cost but 2x quality
- Chat queries: Use Haiku → 10x cost saving, minimal quality loss
- Overall: 40-60% cost saving với semantic routing

---

## 9. Future Enhancements

### 9.1 Embedding-based Classification
```go
// Future: Use vector embeddings cho semantic similarity
func (ic *IntentClassifier) EmbeddingClassify(query string) Intent {
    queryEmbedding := ic.embedder.Embed(query)
    
    // Calculate similarity với intent embeddings
    similarities := map[Intent]float64{
        IntentCoding:   cosineSimilarity(queryEmbedding, ic.codingEmbedding),
        IntentResearch: cosineSimilarity(queryEmbedding, ic.researchEmbedding),
        // ...
    }
    
    return maxSimilarity(similarities)
}
```

### 9.2 Multi-label Classification
```go
// Future: Support multiple intents per query
type MultiIntentClassification struct {
    Primary   Intent
    Secondary []Intent
    Confidence map[Intent]float64
}
```

### 9.3 Context-aware Classification
```go
// Future: Consider conversation history
func (ic *IntentClassifier) ClassifyWithContext(query string, history []Message) Intent {
    context := buildContextString(history)
    combinedQuery := query + "\n" + context
    return ic.Classify(combinedQuery)
}
```

---

## 10. Lessons Learned (dành cho sau implementation)

**Dựa trên Phase 1:**
- Naming conflicts dễ xảy ra với multiple implementations
- Type consistency quan trọng trong Go
- Config-based initialization cho phép flexibility
- HTTP endpoints giúp debugging

**Dự đoán cho Phase 2:**
- LLM classification latency cần được monitored
- Cache strategy quan trọng cho performance
- Fallback logic cần robust
- Logging essential cho debugging classification decisions

---

**Document Status**: Draft v1.0  
**Last Updated**: 2026-05-04  
**Next Review**: Sau khi implementation complete
