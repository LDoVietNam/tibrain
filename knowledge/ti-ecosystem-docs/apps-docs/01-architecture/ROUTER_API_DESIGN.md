# Ti Router API - /api/v1/models Endpoint Design

## Overview
Design cho `/api/v1/models` endpoint để support dynamic model routing cho adaptive-orchestration-v3.yaml

## Current State
File: `apps/router/client/models.go`

```go
type Model struct {
    ID       string `json:"id"`
    Object   string `json:"object"`
    Created  int64  `json:"created"`
    OwnedBy  string `json:"owned_by"`
}
```

❌ **Insufficient** - Không có metadata cần thiết cho dynamic routing

## Enhanced Model Schema

### Request Format
```
GET /api/v1/models?capabilities={capability}&complexity={complexity}&available={bool}
```

### Query Parameters
- `capabilities` (optional): Filter by capability (coding, reasoning, architecture, review)
- `complexity` (optional): Filter by complexity level (simple, standard, complex, critical)
- `available` (optional): Filter by availability (default: true)
- `provider` (optional): Filter by provider (openai, anthropic, groq, etc.)
- `min_quality` (optional): Minimum quality score (0-10)
- `max_cost` (optional): Maximum cost per 1k tokens

### Response Format
```json
{
  "object": "list",
  "data": [
    {
      "id": "openai:gpt-4o",
      "provider": "openai",
      "name": "gpt-4o",
      "object": "model",
      "created": 1234567890,
      "owned_by": "openai",
      
      // Enhanced metadata
      "capabilities": ["coding", "reasoning", "architecture", "review"],
      "cost_per_1k_tokens": {
        "input": 0.0025,
        "output": 0.01
      },
      "quality_score": 8.5,
      "speed": "medium",
      "context_window": 128000,
      "available": true,
      "latency_ms": 500,
      "max_tokens": 4096,
      
      // Additional metadata
      "description": "GPT-4o: Most capable GPT-4 model",
      "supports_vision": false,
      "supports_functions": true,
      "supports_streaming": true,
      "supports_json": true
    }
  ],
  "total": 15,
  "filtered": 5
}
```

## Enhanced Model Struct

```go
// Model represents an AI model with enhanced metadata
type Model struct {
    // Basic fields (existing)
    ID       string `json:"id"`
    Object   string `json:"object"`
    Created  int64  `json:"created"`
    OwnedBy  string `json:"owned_by"`
    
    // Provider information
    Provider string `json:"provider"`
    Name     string `json:"name"`
    
    // Capabilities
    Capabilities []string `json:"capabilities"`
    
    // Cost information
    CostPer1kTokens CostInfo `json:"cost_per_1k_tokens"`
    
    // Performance metrics
    QualityScore float64 `json:"quality_score"` // 0-10
    Speed        string  `json:"speed"`         // fast, medium, slow
    LatencyMs    int     `json:"latency_ms"`
    
    // Model limits
    ContextWindow int `json:"context_window"`
    MaxTokens     int `json:"max_tokens"`
    
    // Availability
    Available bool `json:"available"`
    
    // Additional features
    Description       string   `json:"description"`
    SupportsVision   bool     `json:"supports_vision"`
    SupportsFunctions bool    `json:"supports_functions"`
    SupportsStreaming  bool    `json:"supports_streaming"`
    SupportsJSON      bool    `json:"supports_json"`
}

// CostInfo represents cost per 1k tokens
type CostInfo struct {
    Input  float64 `json:"input"`
    Output float64 `json:"output"`
}

// ModelsResponse is the enhanced response
type ModelsResponse struct {
    Object  string  `json:"object"`
    Data    []Model `json:"data"`
    Total   int     `json:"total"`
    Filtered int    `json:"filtered"`
}
```

## Model Metadata Storage

### Option 1: Hardcoded in Router
```go
// apps/router/cmd/routerd/handlers/models/models.go
var modelMetadata = map[string]ModelMetadata{
    "openai:gpt-4o": {
        Provider: "openai",
        Name: "gpt-4o",
        Capabilities: []string{"coding", "reasoning", "architecture", "review"},
        CostPer1kTokens: CostInfo{Input: 0.0025, Output: 0.01},
        QualityScore: 8.5,
        Speed: "medium",
        ContextWindow: 128000,
        // ...
    },
    // ... other models
}
```

### Option 2: Database Storage
```sql
CREATE TABLE models (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    name TEXT NOT NULL,
    capabilities TEXT, -- JSON array
    cost_input REAL,
    cost_output REAL,
    quality_score REAL,
    speed TEXT,
    context_window INTEGER,
    max_tokens INTEGER,
    available INTEGER DEFAULT 1,
    latency_ms INTEGER,
    description TEXT,
    supports_vision INTEGER DEFAULT 0,
    supports_functions INTEGER DEFAULT 0,
    supports_streaming INTEGER DEFAULT 0,
    supports_json INTEGER DEFAULT 0
);

CREATE INDEX idx_models_provider ON models(provider);
CREATE INDEX idx_models_available ON models(available);
CREATE INDEX idx_models_quality ON models(quality_score);
```

### Option 3: Dynamic Discovery from Provider APIs
- Query provider APIs (OpenAI, Anthropic, Groq) to discover models
- Cache metadata in database
- Periodically refresh

## Implementation Plan

### Phase 1: Design API Schema ✅
- Define enhanced Model struct
- Define request/response format
- Define query parameters

### Phase 2: Implement Handler
- Create `apps/router/cmd/routerd/handlers/models/handlers.go`
- Implement GET /api/v1/models endpoint
- Implement filtering logic
- Implement sorting logic

### Phase 3: Add Model Metadata
- Choose storage option (recommend Option 1 initially, migrate to Option 3 later)
- Add metadata for existing providers (OpenAI, Anthropic, Groq)
- Add metadata for new providers as they're added

### Phase 4: Test API
- Unit tests for filtering logic
- Unit tests for sorting logic
- Integration tests for API endpoint
- Manual testing with curl

### Phase 5: Document API
- Update API documentation
- Add examples
- Add error handling documentation

## Filtering & Sorting Logic

### Filtering
1. Filter by capabilities (if provided)
2. Filter by complexity (map to quality thresholds)
3. Filter by availability (default: true)
4. Filter by provider (if provided)
5. Filter by min_quality (if provided)
6. Filter by max_cost (if provided)

### Complexity to Quality Threshold Mapping
- simple: quality >= 6.0
- standard: quality >= 7.0
- complex: quality >= 8.0
- critical: quality >= 9.0

### Sorting
Default sort by: `(quality_score / cost_per_1k_tokens.input)` descending

Custom sort options (via query param):
- `sort_by=quality` - Sort by quality_score descending
- `sort_by=cost` - Sort by cost ascending
- `sort_by=speed` - Sort by latency_ms ascending
- `sort_by=efficiency` - Sort by (quality_score / cost) descending (default)

## Error Handling

### HTTP Status Codes
- 200 OK - Success
- 400 Bad Request - Invalid query parameters
- 500 Internal Server Error - Server error

### Error Response Format
```json
{
  "error": {
    "message": "Invalid query parameter: capabilities",
    "code": "INVALID_PARAMETER",
    "details": "Valid capabilities: coding, reasoning, architecture, review"
  }
}
```

## Integration with Adaptive Orchestration

### Workflow Integration
adaptive-orchestration-v3.yaml sẽ query API như sau:

```bash
# Query models for coding task with standard complexity
curl "http://localhost:8080/api/v1/models?capabilities=coding&complexity=standard&available=true"

# Response will include models filtered by:
# - capability: coding
# - quality >= 7.0 (standard threshold)
# - available: true
# - sorted by efficiency (quality/cost)
```

## Success Criteria
- ✅ Enhanced Model struct with all required fields
- ✅ GET /api/v1/models endpoint implemented
- ✅ Filtering by capabilities, complexity, availability
- ✅ Sorting by efficiency (quality/cost)
- ✅ Model metadata for OpenAI, Anthropic, Groq
- ✅ Unit tests
- ✅ Integration tests
- ✅ API documentation
- ✅ Error handling
