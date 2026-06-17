# Tool Search Implementation Plan

> **Ngày tạo:** 2026-04-30
> **Task:** Implement Tool Search Support cho Ti Router
> **Status:** IN PROGRESS

---

## Overview

Tool Search support cho phép Claude sử dụng `tool_reference` blocks để reference tools thay vì gửi full schema. Router cần:
1. Parse `tool_reference` blocks từ Claude API response
2. Tool registry để map tool ID → full schema
3. Forward `tool_reference` blocks đến providers với full schema
4. Tool search endpoint cho dynamic tool discovery

---

## Requirements từ AGENTS.md

### Phase 1: Tool Reference Parser (2-4 hours)
- Parse Claude API response
- Extract `tool_reference` blocks
- Validate format

### Phase 2: Tool Registry (1-2 hours)
- Implement tool registry
- Map tool ID → full schema
- Thread-safe operations

### Phase 3: Forward Logic (2-4 hours)
- Replace `tool_reference` với full schema
- Forward đến downstream providers
- Handle errors gracefully

### Phase 4: Tool Search Endpoint (2-4 hours)
- Implement search endpoint
- Support fuzzy matching
- Return ranked results

**Total: 7-14 hours**

**Agent Owner**: routing-agent hoặc provider-agent

---

## Implementation Plan

### Phase 1: Tool Reference Parser

**Location:** `Z:\Ti\router\layers\tool\parser.go`

**Tasks:**
1. Define `ToolReference` struct
2. Implement parser để extract tool_reference blocks từ API response
3. Validate format
4. Write tests

**Code Structure:**
```go
type ToolReference struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    Version string `json:"version,omitempty"`
}

type ToolReferenceParser struct{}

func (p *ToolReferenceParser) Parse(response []byte) ([]ToolReference, error)
func (p *ToolReferenceParser) Validate(ref ToolReference) error
```

---

### Phase 2: Tool Registry

**Location:** `Z:\Ti\router\layers\tool\registry.go`

**Tasks:**
1. Define `ToolSchema` struct
2. Implement registry với CRUD operations
3. Thread-safe với sync.RWMutex
4. Write tests

**Code Structure:**
```go
type ToolSchema struct {
    ID      string          `json:"id"`
    Name    string          `json:"name"`
    Version string          `json:"version"`
    Schema  json.RawMessage `json:"schema"`
    Provider string          `json:"provider"`
}

type ToolRegistry struct {
    tools map[string]*ToolSchema
    mu    sync.RWMutex
}

func (r *ToolRegistry) Register(schema *ToolSchema) error
func (r *ToolRegistry) Get(id string) (*ToolSchema, bool)
func (r *ToolRegistry) List() []*ToolSchema
func (r *ToolRegistry) Search(query string) []*ToolSchema
```

---

### Phase 3: Forward Logic

**Location:** `Z:\Ti\router\layers\tool\forward.go`

**Tasks:**
1. Implement logic để replace tool_reference với full schema
2. Forward đến downstream providers
3. Error handling
4. Write tests

**Code Structure:**
```go
type ToolForwarder struct {
    registry *ToolRegistry
}

func (f *ToolForwarder) ExpandTools(request []byte) ([]byte, error)
func (f *ToolForwarder) ForwardToProvider(request []byte, provider string) ([]byte, error)
```

---

### Phase 4: Tool Search Endpoint

**Location:** `Z:\Ti\router\layers\tool\search.go`

**Tasks:**
1. Implement search endpoint
2. Fuzzy matching algorithm
3. Ranking logic
4. Write tests

**Code Structure:**
```go
type ToolSearcher struct {
    registry *ToolRegistry
}

func (s *ToolSearcher) Search(query string) ([]*ToolSchema, error)
func (s *ToolSearcher) FuzzyMatch(query string, schema *ToolSchema) float64
```

---

## Integration Points

**Router Integration:**
- Wire ToolRegistry vào router initialization
- Add ToolReferenceParser vào request pipeline
- Add ToolForwarder vào provider call logic
- Add ToolSearcher vào admin endpoints

**HTTP Endpoints:**
- `POST /admin/tools/register` - Register tool schema
- `GET /admin/tools/list` - List all tools
- `GET /admin/tools/search?q={query}` - Search tools
- `DELETE /admin/tools/{id}` - Unregister tool

---

## Testing Strategy

### Unit Tests
- ToolReferenceParser tests
- ToolRegistry tests (CRUD, thread-safety)
- ToolForwarder tests
- ToolSearcher tests (fuzzy matching)

### Integration Tests
- Full flow: Claude API → tool_reference → expand → provider
- Tool search endpoint tests
- Error handling tests

---

## Next Steps

1. ✅ Create implementation plan
2. ⏳ Implement Phase 1: Tool Reference Parser
3. ⏳ Implement Phase 2: Tool Registry
4. ⏳ Implement Phase 3: Forward Logic
5. ⏳ Implement Phase 4: Tool Search Endpoint
6. ⏳ Integration testing
7. ⏳ Update AGENTS.md

---

## References

- Claude API documentation cho tool_reference format
- OpenAI Tools API specification
- Existing tool implementations trong codebase
