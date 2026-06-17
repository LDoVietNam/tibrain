# Router-Based Rules Enforcement Architecture

> **Version**: 1.0.0
> **Last Updated**: 2026-05-05
> **Purpose**: Centralized rules enforcement via Ti Router middleware layer

---

## 🎯 Overview

Router đóng vai trò như **middleware layer** để enforce rules centrally cho tất cả CLI/agents/sub-agents. Thay vì mỗi agent phải tự enforce rules locally, Router sẽ:
1. Nhận biết request từ CLI/agent/sub-agent nào
2. Validate request against rules
3. Cache rule validation results
4. Log violations
5. Block/warn/allow dựa trên severity

**Benefits**:
- ✅ Centralized enforcement (không cần duplicate logic)
- ✅ Consistent behavior across all agents
- ✅ Cache optimization (không phải validate lặp lại)
- ✅ Audit trail (tất cả violations được log tại Router)
- ✅ Easy to update rules (update 1 chỗ, áp dụng cho tất cả)

---

## 🏗️ Architecture

### Current Router Layers

```
Request Flow:
HTTP Request
  ↓
Authentication (authentication.Service)
  ↓
Rules Enforcement (NEW - rulesenforcement.Service)
  ↓
Monitoring (monitoring.Monitor)
  ↓
RTK Compression (rtk.RTKCompressor)
  ↓
Prompt Cache (promptcache.Middleware)
  ↓
Episodic Memory (memory.Middleware)
  ↓
Request Optimization (optimization.RequestOptimizer)
  ↓
Semantic Routing (routing.SemanticRoutingMiddleware)
  ↓
Provider Selection (providers.Registry)
  ↓
Rate Limiting (resilience.RateLimiter)
  ↓
Cache (resilience.Cache)
  ↓
Upstream Provider
```

### New Layer: Rules Enforcement

**Location**: `apps/router/layers/rulesenforcement/`

**Components**:
- `service.go` - Main service interface
- `validator.go` - Rule validation logic
- `cache.go` - Rule validation cache
- `detector.go` - Agent/CLI/sub-agent detection
- `middleware.go` - HTTP middleware integration
- `audit.go` - Violation logging

---

## 🔍 Request Identification

### Agent/CLI/Sub-Agent Tagging

**Approach 1: HTTP Headers** (Recommended)

Devin/CLI/Sub-agents gửi headers:

```http
POST /v1/chat/completions
X-Agent-Type: devin
X-Agent-Version: 1.6.0
X-Agent-Session: session-123
X-Task-Type: coding
X-Request-ID: req-123456
```

**Approach 2: API Key with Agent Metadata**

```http
Authorization: Bearer ti_agent_<agent_type>_<api_key>
```

Router parse API key để extract agent type:
- `ti_agent_devin_*` → Devin agent
- `ti_agent_claude_*` → Claude agent
- `ti_agent_codex_*` → Codex agent
- `ti_cli_*` → Ti CLI

**Approach 3: OAuth with Agent Scope**

```
OAuth token với scopes:
- agent:devin
- agent:claude
- agent:codex
- agent:cli
```

### Implementation

**File**: `apps/router/layers/rulesenforcement/detector.go`

```go
package rulesenforcement

import (
	"net/http"
	"strings"
)

// AgentType identifies the type of agent making the request
type AgentType string

const (
	AgentTypeDevin   AgentType = "devin"
	AgentTypeClaude  AgentType = "claude"
	AgentTypeCodex  AgentType = "codex"
	AgentTypeQwen   AgentType = "qwen"
	AgentTypeKilo   AgentType = "kilo"
	AgentTypeCLI    AgentType = "cli"
	AgentTypeUnknown AgentType = "unknown"
)

// AgentContext holds agent identification information
type AgentContext struct {
	Type        AgentType `json:"type"`
	Version     string    `json:"version"`
	SessionID   string    `json:"session_id"`
	TaskType    string    `json:"task_type"`
	RequestID   string    `json:"request_id"`
	UserAgent   string    `json:"user_agent"`
	RemoteAddr  string    `json:"remote_addr"`
}

// Detector identifies the agent type from request
type Detector struct {
	// Optional: API key patterns for different agents
	apiKeyPatterns map[AgentType]string
}

// NewDetector creates a new agent detector
func NewDetector() *Detector {
	return &Detector{
		apiKeyPatterns: map[AgentType]string{
			AgentTypeDevin:  "ti_agent_devin_",
			AgentTypeClaude: "ti_agent_claude_",
			AgentTypeCodex:  "ti_agent_codex_",
			AgentTypeQwen:   "ti_agent_qwen_",
			AgentTypeKilo:   "ti_agent_kilo_",
			AgentTypeCLI:    "ti_cli_",
		},
	}
}

// Detect identifies the agent from HTTP request
func (d *Detector) Detect(r *http.Request) AgentContext {
	ctx := AgentContext{
		Type:       AgentTypeUnknown,
		UserAgent:  r.UserAgent(),
		RemoteAddr: r.RemoteAddr,
	}

	// Method 1: Check X-Agent-Type header
	if agentType := r.Header.Get("X-Agent-Type"); agentType != "" {
		ctx.Type = AgentType(agentType)
		ctx.Version = r.Header.Get("X-Agent-Version")
		ctx.SessionID = r.Header.Get("X-Agent-Session")
		ctx.TaskType = r.Header.Get("X-Task-Type")
		ctx.RequestID = r.Header.Get("X-Request-ID")
		return ctx
	}

	// Method 2: Parse Authorization header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		for agentType, pattern := range d.apiKeyPatterns {
			if strings.HasPrefix(token, pattern) {
				ctx.Type = agentType
				break
			}
		}
	}

	// Method 3: Check User-Agent
	userAgent := r.UserAgent()
	if strings.Contains(userAgent, "Devin") {
		ctx.Type = AgentTypeDevin
	} else if strings.Contains(userAgent, "Claude") {
		ctx.Type = AgentTypeClaude
	} else if strings.Contains(userAgent, "Codex") {
		ctx.Type = AgentTypeCodex
	} else if strings.Contains(userAgent, "Ti-CLI") {
		ctx.Type = AgentTypeCLI
	}

	return ctx
}

// IsKnownAgent checks if the agent type is known
func (d *Detector) IsKnownAgent(agentType AgentType) bool {
	switch agentType {
	case AgentTypeDevin, AgentTypeClaude, AgentTypeCodex, AgentTypeQwen, AgentTypeKilo, AgentTypeCLI:
		return true
	default:
		return false
	}
}
```

---

## 📋 Rule Validation

### Rule Types

**Core Rules** (from `content/rules/core/`):
- Exploration rules
- Planning rules
- Implementation rules
- Review rules
- Beads rules

**Quality Rules** (from `content/rules/quality/`):
- Mandatory Workflow Usage
- Mandatory Quality Principles
- Mandatory Workflow Selection
- Mandatory Quality Gates

### Validation Logic

**File**: `apps/router/layers/rulesenforcement/validator.go`

```go
package rulesenforcement

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Rule represents a single rule
type Rule struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Priority    string   `json:"priority"`
	Description string   `json:"description"`
	Checkpoints []string `json:"checkpoints"`
}

// Violation represents a rule violation
type Violation struct {
	RuleName    string `json:"rule_name"`
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	Context     string `json:"context"`
	AgentType   string `json:"agent_type"`
	RequestID   string `json:"request_id"`
	Timestamp   int64  `json:"timestamp"`
}

// ValidationResult holds the result of rule validation
type ValidationResult struct {
	Valid      bool         `json:"valid"`
	Violations []Violation  `json:"violations"`
	Warnings   []Violation  `json:"warnings"`
	Confidence float64      `json:"confidence"`
}

// Validator validates requests against rules
type Validator struct {
	rulesPath string
	rules     map[string][]Rule
}

// NewValidator creates a new rule validator
func NewValidator(rulesPath string) *Validator {
	return &Validator{
		rulesPath: rulesPath,
		rules:     make(map[string][]Rule),
	}
}

// LoadRules loads rules from the rules directory
func (v *Validator) LoadRules(ctx context.Context) error {
	categories := []string{"core", "quality"}

	for _, category := range categories {
		categoryPath := filepath.Join(v.rulesPath, category)
		entries, err := os.ReadDir(categoryPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			ruleFile := filepath.Join(categoryPath, entry.Name())
			rule, err := v.parseRuleFile(ruleFile)
			if err != nil {
				continue
			}

			v.rules[category] = append(v.rules[category], rule)
		}
	}

	return nil
}

// parseRuleFile parses a rule markdown file
func (v *Validator) parseRuleFile(path string) (Rule, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Rule{}, err
	}

	// Parse markdown frontmatter and content
	// Extract rule name, category, priority, checkpoints
	// This is simplified - use proper markdown parser in production

	return Rule{
		Name:        filepath.Base(path),
		Category:    "quality",
		Priority:    "P0",
		Description: string(content),
		Checkpoints: []string{},
	}, nil
}

// Validate validates a request against rules
func (v *Validator) Validate(ctx context.Context, req map[string]interface{}, agentCtx AgentContext) ValidationResult {
	result := ValidationResult{
		Valid:      true,
		Violations: []Violation{},
		Warnings:   []Violation{},
		Confidence: 1.0,
	}

	// Check quality rules
	for _, rule := range v.rules["quality"] {
		if violation := v.checkRule(rule, req, agentCtx); violation != nil {
			if rule.Priority == "P0" {
				result.Violations = append(result.Violations, *violation)
				result.Valid = false
			} else {
				result.Warnings = append(result.Warnings, *violation)
			}
		}
	}

	// Check core rules
	for _, rule := range v.rules["core"] {
		if violation := v.checkRule(rule, req, agentCtx); violation != nil {
			if rule.Priority == "P0" {
				result.Violations = append(result.Violations, *violation)
				result.Valid = false
			} else {
				result.Warnings = append(result.Warnings, *violation)
			}
		}
	}

	return result
}

// checkRule checks a single rule
func (v *Validator) checkRule(rule Rule, req map[string]interface{}, agentCtx AgentContext) *Violation {
	// This is simplified - implement actual rule checking logic
	// For example:
	// - Check if workflow-orchestrator skill was invoked
	// - Check if context collection was done
	// - Check if assumptions were documented
	// - Check if workflow selection was logged

	return nil
}
```

---

## 💾 Cache Layer

### Rule Validation Cache

**File**: `apps/router/layers/rulesenforcement/cache.go`

```go
package rulesenforcement

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CachedValidation stores a cached validation result
type CachedValidation struct {
	Result     ValidationResult `json:"result"`
	AgentType  AgentType        `json:"agent_type"`
	CreatedAt  time.Time        `json:"created_at"`
	TTL        time.Duration    `json:"ttl"`
}

// IsExpired checks if the cached entry has expired
func (c *CachedValidation) IsExpired() bool {
	return time.Since(c.CreatedAt) > c.TTL
}

// ValidationCache caches rule validation results
type ValidationCache struct {
	mu    sync.RWMutex
	store map[string]CachedValidation
	ttl   time.Duration
}

// NewValidationCache creates a new validation cache
func NewValidationCache(ttl time.Duration) *ValidationCache {
	return &ValidationCache{
		store: make(map[string]CachedValidation),
		ttl:   ttl,
	}
}

// Get retrieves a cached validation result
func (vc *ValidationCache) Get(req map[string]interface{}, agentCtx AgentContext) *ValidationResult {
	key := vc.cacheKey(req, agentCtx)
	vc.mu.RLock()
	entry, ok := vc.store[key]
	vc.mu.RUnlock()
	if !ok || entry.IsExpired() {
		return nil
	}
	return &entry.Result
}

// Set stores a validation result in cache
func (vc *ValidationCache) Set(req map[string]interface{}, agentCtx AgentContext, result ValidationResult) {
	key := vc.cacheKey(req, agentCtx)
	vc.mu.Lock()
	vc.store[key] = CachedValidation{
		Result:    result,
		AgentType: agentCtx.Type,
		CreatedAt: time.Now(),
		TTL:       vc.ttl,
	}
	vc.mu.Unlock()
}

// Invalidate removes a cached entry
func (vc *ValidationCache) Invalidate(req map[string]interface{}, agentCtx AgentContext) {
	key := vc.cacheKey(req, agentCtx)
	vc.mu.Lock()
	delete(vc.store, key)
	vc.mu.Unlock()
}

// cacheKey generates a cache key from request and agent context
func (vc *ValidationCache) cacheKey(req map[string]interface{}, agentCtx AgentContext) string {
	data := map[string]interface{}{
		"request":   req,
		"agent":     agentCtx.Type,
		"task_type": agentCtx.TaskType,
	}
	b, _ := json.Marshal(data)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}
```

---

## 🔗 Devin Integration

### Approach 1: HTTP API (Recommended)

Router exposes HTTP API cho Devin:

```http
POST /api/rules/validate
Content-Type: application/json
X-Agent-Type: devin
X-Agent-Version: 1.6.0

{
  "request": {...},
  "task_type": "coding",
  "action": "edit_file"
}
```

Response:
```json
{
  "valid": true,
  "violations": [],
  "warnings": [],
  "confidence": 1.0
}
```

### Approach 2: MCP Server

Router exposes MCP server:

```go
// apps/mcp/server/rules-validator/main.go
package main

func main() {
	server := mcp.NewServer("rules-validator")
	
	server.AddTool(mcp.Tool{
		Name: "validate_action",
		Description: "Validate action against rules",
		InputSchema: schema,
		Handler: validateActionHandler,
	})
	
	server.ServeStdio()
}
```

### Approach 3: Devin Skill với Router API

**File**: `.devin/skills/router-rules-checker/SKILL.md`

```markdown
# Router Rules Checker

> **Purpose**: Validate actions against rules via Ti Router
> **Auto-invocation**: Before every file edit, command execution

## Usage

```bash
# Validate action via Router API
curl -X POST http://localhost:1807/api/rules/validate \
  -H "X-Agent-Type: devin" \
  -H "Content-Type: application/json" \
  -d '{
    "request": {...},
    "task_type": "coding",
    "action": "edit_file"
  }'
```

## Response Handling

- If `valid: false` → Block action
- If `warnings` present → Warn user
- If `valid: true` → Proceed
```

---

## 📊 Audit Trail

### Violation Logging

**File**: `apps/router/layers/rulesenforcement/audit.go`

```go
package rulesenforcement

import (
	"context"
	"encoding/json"
	"time"
)

// Auditor logs rule violations
type Auditor struct {
	// Can integrate with existing audit.Auditor
	// Or use BD tool for logging
}

// LogViolation logs a rule violation
func (a *Auditor) LogViolation(ctx context.Context, violation Violation) error {
	// Log to audit log
	// Log to BD tool
	// Log to monitoring

	// Example: Log to BD tool
	bdCmd := fmt.Sprintf(
		`"Z:/02_CORE/_cli/bin/bd.exe" log --task="Rule violation: %s - %s" --agent=%s --status=failed --type=violation --domain=cli`,
		violation.RuleName,
		violation.Message,
		violation.AgentType,
	)

	// Execute BD command
	return nil
}

// GetViolationStats gets violation statistics
func (a *Auditor) GetViolationStats(ctx context.Context, agentType string, timeRange time.Duration) (map[string]int, error) {
	// Query audit log for violation statistics
	// Return counts by rule, severity, etc.

	return map[string]int{
		"total":           0,
		"critical":        0,
		"warnings":        0,
		"workflow_skips":  0,
		"context_missing": 0,
	}, nil
}
```

---

## 🚀 Implementation Plan

### Phase 1: Foundation (Week 1)
1. ✅ Create `layers/rulesenforcement/` package
2. ✅ Implement `detector.go` (agent detection)
3. ✅ Implement `cache.go` (validation cache)
4. ✅ Implement basic `validator.go` (rule loading)
5. ✅ Add unit tests

### Phase 2: Integration (Week 2)
1. ✅ Integrate rules enforcement into Router handler
2. ✅ Add HTTP API endpoint `/api/rules/validate`
3. ✅ Implement `audit.go` (violation logging)
4. ✅ Add BD tool integration for logging
5. ✅ Add monitoring metrics

### Phase 3: Devin Integration (Week 3)
1. ✅ Create Devin skill `router-rules-checker`
2. ✅ Add auto-invocation triggers
3. ✅ Test end-to-end flow
4. ✅ Document usage

### Phase 4: Enhancement (Week 4)
1. ✅ Implement rule-specific validation logic
2. ✅ Add confidence scoring
3. ✅ Add progressive disclosure support
4. ✅ Add rollback tracking
5. ✅ Add pattern database update

---

## 📝 Configuration

### Router Config

```yaml
# configs/router.yaml
rules_enforcement:
  enabled: true
  rules_path: "Z:/10_WORKPLACE/Ti/content/rules"
  cache_ttl: 5m
  strict_mode: false
  log_violations: true
  block_critical: true
  warn_non_critical: true

  agent_detection:
    method: "headers" # or "api_key" or "oauth"
    require_known_agent: false

  validation:
    check_core_rules: true
    check_quality_rules: true
    confidence_threshold: 0.7
```

### Devin Config

```json
// .devin/config/settings.json
{
  "rules_enforcement": {
    "enabled": true,
    "router_url": "http://localhost:1807",
    "validate_before_action": true,
    "block_on_violation": true,
    "warn_on_warning": true
  }
}
```

---

## 🎯 Benefits

### Before (Current)
- ❌ Each agent enforces rules locally
- ❌ Duplicate logic across agents
- ❌ Inconsistent behavior
- ❌ Hard to update rules
- ❌ No centralized audit trail

### After (Router-based)
- ✅ Centralized enforcement
- ✅ Consistent behavior
- ✅ Easy to update rules
- ✅ Centralized audit trail
- ✅ Cache optimization
- ✅ Better monitoring

---

## 📚 References

- **Router Architecture**: `apps/router/layers/`
- **Rules Directory**: `content/rules/`
- **AGENTS.md**: `AGENTS.md`
- **BD Tool**: `Z:/02_CORE/_cli/bin/bd.exe`
- **Knowledge Graph Memory**: `content/mcp/KNOWLEDGE_GRAPH_USAGE_GUIDE.md`
