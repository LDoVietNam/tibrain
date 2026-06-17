# Router Middleware Plugin Architecture

> **Version**: 1.0.0
> **Last Updated**: 2026-05-05
> **Purpose**: Middleware plugin system for Ti Router (extend existing provider plugin system)

---

## 🎯 Overview

Ti Router đã có plugin system cho providers (HashiCorp go-plugin pattern). Chúng ta sẽ **extend** system này để support middleware plugins như Rules Enforcement.

**Benefits of Plugin Approach vs Layer**:
- ✅ **Loosely coupled** - Không hard dependency vào Router core
- ✅ **Easy enable/disable** - Config-based, no code changes
- ✅ **Independent testing** - Test plugin riêng biệt
- ✅ **Hot-reload** - Reload plugin mà không restart Router
- ✅ **Distribution** - Plugin có thể distribute riêng
- ✅ **Versioning** - Mỗi plugin có version riêng
- ✅ **Isolation** - Plugin crash không ảnh hưởng Router

---

## 🏗️ Architecture

### Existing Plugin System (Provider Plugins)

```
Router
  ↓
PluginLoader (layers/provider/plugin/)
  ↓
ProviderPlugin Interface
  ↓
Plugin Process (net/rpc)
  ↓
Provider Implementation
```

### Extended Plugin System (Middleware Plugins)

```
Router
  ↓
PluginLoader (middleware/plugin/) ← NEW
  ↓
MiddlewarePlugin Interface ← NEW
  ↓
Plugin Process (net/rpc)
  ↓
Middleware Implementation
  ↓
Rules Enforcement Plugin ← FIRST PLUGIN
```

### Plugin Types

| Type | Interface | Purpose | Location |
|------|-----------|---------|----------|
| **Provider** | `ProviderPlugin` | AI provider implementations | `layers/provider/plugin/` |
| **Middleware** | `MiddlewarePlugin` | Request/response processing | `layers/middleware/plugin/` ← NEW |

---

## 🔌 Middleware Plugin Interface

**File**: `apps/router/layers/middleware/plugin/interface.go`

```go
// Package plugin implements a dynamic plugin system for middleware.
// Plugins are loaded as separate processes and communicate via net/rpc,
// following the HashiCorp go-plugin pattern.
package plugin

import (
	"context"
	"net/http"
)

// MiddlewareContext holds context about the request/response
type MiddlewareContext struct {
	RequestID   string                 `json:"request_id"`
	Headers     map[string]string      `json:"headers"`
	RequestBody map[string]interface{} `json:"request_body"`
	AgentType   string                 `json:"agent_type"` // devin, claude, codex, cli
	TaskType    string                 `json:"task_type"`
	UserAgent   string                 `json:"user_agent"`
	RemoteAddr  string                 `json:"remote_addr"`
	Timestamp   int64                  `json:"timestamp"`
}

// MiddlewareAction represents an action the plugin wants to take
type MiddlewareAction string

const (
	ActionAllow    MiddlewareAction = "allow"    // Allow request to proceed
	ActionBlock    MiddlewareAction = "block"    // Block request with error
	ActionModify   MiddlewareAction = "modify"   // Modify request/response
	ActionWarn     MiddlewareAction = "warn"     // Allow but log warning
	ActionRedirect MiddlewareAction = "redirect" // Redirect to different endpoint
)

// MiddlewareResult is the result of middleware processing
type MiddlewareResult struct {
	Action    MiddlewareAction       `json:"action"`
	Message   string                 `json:"message,omitempty"`
	Headers   map[string]string      `json:"headers,omitempty"`     // Headers to add/modify
	Body      map[string]interface{} `json:"body,omitempty"`        // Modified body
	ErrorCode int                    `json:"error_code,omitempty"`  // HTTP status code if blocking
	Metadata  map[string]interface{} `json:"metadata,omitempty"`    // Plugin-specific metadata
}

// MiddlewarePhase indicates when the plugin should run
type MiddlewarePhase string

const (
	PhasePreAuth      MiddlewarePhase = "pre_auth"      // Before authentication
	PhasePostAuth     MiddlewarePhase = "post_auth"     // After authentication
	PhasePreRouting   MiddlewarePhase = "pre_routing"   // Before provider selection
	PhasePostRouting  MiddlewarePhase = "post_routing"  // After provider selection
	PhasePreCache     MiddlewarePhase = "pre_cache"     // Before cache check
	PhasePostCache    MiddlewarePhase = "post_cache"    // After cache check
	PhasePreUpstream  MiddlewarePhase = "pre_upstream"  // Before upstream call
	PhasePostUpstream MiddlewarePhase = "post_upstream" // After upstream call
	PhasePreResponse  MiddlewarePhase = "pre_response"  // Before sending response
	PhasePostResponse MiddlewarePhase = "post_response" // After sending response
)

// MiddlewarePlugin is the interface that all middleware plugins must implement.
type MiddlewarePlugin interface {
	// Plugin metadata
	PluginName() string
	PluginVersion() string
	PluginAuthor() string

	// Plugin lifecycle
	InitPlugin(ctx context.Context, config map[string]any) error
	ShutdownPlugin(ctx context.Context) error

	// Middleware processing
	Process(ctx context.Context, phase MiddlewarePhase, ctxData MiddlewareContext) (*MiddlewareResult, error)

	// Plugin capabilities
	Capabilities() MiddlewareCapabilities

	// Health check
	HealthCheck(ctx context.Context) error
}

// MiddlewareCapabilities describes what a plugin can do.
type MiddlewareCapabilities struct {
	Phases        []MiddlewarePhase `json:"phases"`         // Which phases it supports
	CanModify     bool               `json:"can_modify"`     // Can modify request/response
	CanBlock      bool               `json:"can_block"`      // Can block requests
	CanRedirect   bool               `json:"can_redirect"`   // Can redirect requests
	NeedsConfig   bool               `json:"needs_config"`   // Requires configuration
	AsyncProcessing bool             `json:"async_processing"` // Supports async processing
	Features      []string           `json:"features"`       // Additional feature flags
}

// PluginMetadata contains metadata about a loaded middleware plugin.
type MiddlewarePluginMetadata struct {
	Name        string                  `json:"name"`
	Version     string                  `json:"version"`
	Author      string                  `json:"author"`
	Path        string                  `json:"path"`
	Enabled     bool                    `json:"enabled"`
	Config      map[string]any         `json:"config"`
	Capabilities MiddlewareCapabilities `json:"capabilities"`
	LoadedAt    string                  `json:"loaded_at"`
}

// PluginConfig is the configuration for a middleware plugin.
type MiddlewarePluginConfig struct {
	Name     string         `yaml:"name"`     // Plugin name (must match plugin binary name)
	Enabled  bool           `yaml:"enabled"`  // Whether to load this plugin
	Path     string         `yaml:"path"`     // Path to plugin binary (optional)
	Config   map[string]any `yaml:"config"`   // Plugin-specific configuration
	Priority int            `yaml:"priority"` // Execution priority (higher = runs first)
	Phases   []string       `yaml:"phases"`   // Which phases to run at (optional, default all supported)
}
```

---

## 📋 Rules Enforcement Plugin

**Location**: `apps/router/plugins/middleware/rules-enforcement/`

### Plugin Structure

```
apps/router/plugins/middleware/rules-enforcement/
├── main.go           # Plugin entry point
├── plugin.go         # MiddlewarePlugin implementation
├── detector.go       # Agent detection
├── validator.go      # Rule validation
├── cache.go          # Validation cache
├── audit.go          # Violation logging
├── config.go         # Config handling
└── go.mod
```

### Plugin Implementation

**File**: `apps/router/plugins/middleware/rules-enforcement/main.go`

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/go-plugin"
)

// RulesEnforcementPlugin implements the MiddlewarePlugin interface
type RulesEnforcementPlugin struct {
	validator *RuleValidator
	detector  *AgentDetector
	cache     *ValidationCache
	auditor   *ViolationAuditor
	config    map[string]any
}

func (p *RulesEnforcementPlugin) PluginName() string {
	return "rules-enforcement"
}

func (p *RulesEnforcementPlugin) PluginVersion() string {
	return "1.0.0"
}

func (p *RulesEnforcementPlugin) PluginAuthor() string {
	return "Ti Team"
}

func (p *RulesEnforcementPlugin) InitPlugin(ctx context.Context, config map[string]any) error {
	p.config = config

	// Initialize components
	rulesPath := config["rules_path"].(string)
	p.validator = NewRuleValidator(rulesPath)
	p.detector = NewAgentDetector()
	cacheTTL := config["cache_ttl"].(string)
	p.cache = NewValidationCache(cacheTTL)
	p.auditor = NewViolationAuditor()

	// Load rules
	if err := p.validator.LoadRules(ctx); err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	log.Printf("[rules-enforcement] Plugin initialized with %d rules", p.validator.RuleCount())
	return nil
}

func (p *RulesEnforcementPlugin) ShutdownPlugin(ctx context.Context) error {
	log.Printf("[rules-enforcement] Plugin shutting down")
	return nil
}

func (p *RulesEnforcementPlugin) Process(ctx context.Context, phase plugin.MiddlewarePhase, ctxData plugin.MiddlewareContext) (*plugin.MiddlewareResult, error) {
	// Only run at post-auth phase
	if phase != plugin.PhasePostAuth {
		return &plugin.MiddlewareResult{Action: plugin.ActionAllow}, nil
	}

	// Detect agent
	agentCtx := p.detector.Detect(ctxData)

	// Check cache first
	if cached := p.cache.Get(ctxData.RequestBody, agentCtx); cached != nil {
		return cached, nil
	}

	// Validate rules
	result := p.validator.Validate(ctx, ctxData.RequestBody, agentCtx)

	// Cache result
	p.cache.Set(ctxData.RequestBody, agentCtx, result)

	// Log violations if any
	if !result.Valid {
		for _, violation := range result.Violations {
			p.auditor.LogViolation(ctx, violation)
		}
	}

	// Convert to middleware result
	action := plugin.ActionAllow
	if !result.Valid {
		action = plugin.ActionBlock
	} else if len(result.Warnings) > 0 {
		action = plugin.ActionWarn
	}

	return &plugin.MiddlewareResult{
		Action:  action,
		Message: formatResultMessage(result),
	}, nil
}

func (p *RulesEnforcementPlugin) Capabilities() plugin.MiddlewareCapabilities {
	return plugin.MiddlewareCapabilities{
		Phases:        []plugin.MiddlewarePhase{plugin.PhasePostAuth},
		CanModify:     false,
		CanBlock:      true,
		CanRedirect:   false,
		NeedsConfig:   true,
		AsyncProcessing: false,
		Features:      []string{"rule_validation", "agent_detection", "caching"},
	}
}

func (p *RulesEnforcementPlugin) HealthCheck(ctx context.Context) error {
	return nil
}

func main() {
	// Serve as plugin
	plugin.Serve(&plugin.ServeConfig{
		Plugins: map[string]plugin.Plugin{
			"middleware": &RulesEnforcementPlugin{},
		},
	})
}
```

---

## 🔌 Plugin Integration vào Router

### Extend PluginLoader

**File**: `apps/router/layers/middleware/plugin/loader.go`

```go
package plugin

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-plugin"
)

// MiddlewarePluginLoader is responsible for loading middleware plugins.
type MiddlewarePluginLoader struct {
	pluginDir string
	plugins   map[string]*MiddlewarePluginWrapper
	mu        sync.RWMutex
	clients   map[string]*plugin.Client
}

// NewMiddlewarePluginLoader creates a new middleware plugin loader.
func NewMiddlewarePluginLoader(pluginDir string) *MiddlewarePluginLoader {
	return &MiddlewarePluginLoader{
		pluginDir: pluginDir,
		plugins:   make(map[string]*MiddlewarePluginWrapper),
		clients:   make(map[string]*plugin.Client),
	}
}

// MiddlewarePluginWrapper wraps a loaded middleware plugin.
type MiddlewarePluginWrapper struct {
	Metadata MiddlewarePluginMetadata
	Plugin   MiddlewarePlugin
	Client   *plugin.Client
}

// LoadPlugin loads a middleware plugin.
func (l *MiddlewarePluginLoader) LoadPlugin(ctx context.Context, config MiddlewarePluginConfig) (*MiddlewarePluginWrapper, error) {
	// Similar to provider plugin loader
	// But uses MiddlewarePlugin interface instead of ProviderPlugin
	// ...
}

// ProcessRequest processes a request through all enabled middleware plugins.
func (l *MiddlewarePluginLoader) ProcessRequest(ctx context.Context, phase MiddlewarePhase, ctxData MiddlewareContext) (*MiddlewareResult, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Sort plugins by priority (descending)
	sortedPlugins := l.sortPluginsByPriority()

	// Process through each plugin
	for _, wrapper := range sortedPlugins {
		if !wrapper.Metadata.Enabled {
			continue
		}

		// Check if plugin supports this phase
		if !l.supportsPhase(wrapper.Metadata.Capabilities.Phases, phase) {
			continue
		}

		// Process request
		result, err := wrapper.Plugin.Process(ctx, phase, ctxData)
		if err != nil {
			return nil, fmt.Errorf("plugin %s failed: %w", wrapper.Metadata.Name, err)
		}

		// If plugin blocks request, stop processing
		if result.Action == ActionBlock {
			return result, nil
		}
	}

	// All plugins passed
	return &MiddlewareResult{Action: ActionAllow}, nil
}

// sortPluginsByPriority sorts plugins by priority (descending).
func (l *MiddlewarePluginLoader) sortPluginsByPriority() []*MiddlewarePluginWrapper {
	// Sort implementation
	return nil
}

// supportsPhase checks if a plugin supports a specific phase.
func (l *MiddlewarePluginLoader) supportsPhase(phases []MiddlewarePhase, phase MiddlewarePhase) bool {
	for _, p := range phases {
		if p == phase {
			return true
		}
	}
	return false
}
```

### Integrate vào Handler

**File**: `apps/router/cmd/routerd/handlers/chat/handlers.go`

```go
type Handler struct {
	// Existing dependencies...
	// ...

	// NEW: Middleware plugin loader
	middlewareLoader *middlewareplugin.MiddlewarePluginLoader
}

// SetMiddlewareLoader sets the middleware plugin loader dependency.
func (h *Handler) SetMiddlewareLoader(loader *middlewareplugin.MiddlewarePluginLoader) {
	h.middlewareLoader = loader
}

// HandleChatCompletions handles chat completion requests.
func (h *Handler) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	// Generate request ID
	requestID := "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	// Build middleware context
	ctxData := middlewareplugin.MiddlewareContext{
		RequestID:  requestID,
		Headers:    extractHeaders(r),
		AgentType:  r.Header.Get("X-Agent-Type"),
		TaskType:   r.Header.Get("X-Task-Type"),
		UserAgent:  r.UserAgent(),
		RemoteAddr: r.RemoteAddr,
		Timestamp:  time.Now().Unix(),
	}

	// Process through middleware plugins (pre-auth phase)
	if h.middlewareLoader != nil {
		result, err := h.middlewareLoader.ProcessRequest(
			r.Context(),
			middlewareplugin.PhasePreAuth,
			ctxData,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("middleware error: %v", err), http.StatusInternalServerError)
			return
		}
		if result.Action == middlewareplugin.ActionBlock {
			http.Error(w, result.Message, result.ErrorCode)
			return
		}
	}

	// Authentication
	if !h.authenticateRequest(r, requestID, start) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Process through middleware plugins (post-auth phase)
	if h.middlewareLoader != nil {
		result, err := h.middlewareLoader.ProcessRequest(
			r.Context(),
			middlewareplugin.PhasePostAuth,
			ctxData,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("middleware error: %v", err), http.StatusInternalServerError)
			return
		}
		if result.Action == middlewareplugin.ActionBlock {
			http.Error(w, result.Message, result.ErrorCode)
			return
		}
	}

	// Continue with existing request processing...
	// ...
}
```

---

## 📝 Configuration

### Router Config

**File**: `configs/router.yaml`

```yaml
# Middleware plugins configuration
middleware_plugins:
  enabled: true
  plugin_dir: "apps/router/plugins/middleware"
  
  plugins:
    - name: rules-enforcement
      enabled: true
      priority: 100
      phases:
        - post_auth
      config:
        rules_path: "Z:/10_WORKPLACE/Ti/content/rules"
        cache_ttl: "5m"
        strict_mode: false
        block_critical: true
        warn_non_critical: true
        log_violations: true
        bd_tool_path: "Z:/02_CORE/_cli/bin/bd.exe"
```

### Provider Config (Keep Existing)

```yaml
# Provider plugins configuration (existing)
provider_plugins:
  enabled: true
  plugin_dir: "apps/router/plugins/provider"
  
  plugins:
    - name: anthropic
      enabled: true
      # ...
```

---

## 🚀 Implementation Plan

### Phase 1: Middleware Plugin System (Week 1)
1. ✅ Create `layers/middleware/plugin/` package
2. ✅ Implement `MiddlewarePlugin` interface
3. ✅ Implement `MiddlewarePluginLoader`
4. ✅ Implement plugin discovery and loading
5. ✅ Add unit tests

### Phase 2: Rules Enforcement Plugin (Week 2)
1. ✅ Create `plugins/middleware/rules-enforcement/` plugin
2. ✅ Implement agent detection
3. ✅ Implement rule validation
4. ✅ Implement validation cache
5. ✅ Implement violation logging (BD tool)
6. ✅ Build plugin binary

### Phase 3: Router Integration (Week 3)
1. ✅ Integrate middleware loader into Router handler
2. ✅ Add configuration support
3. ✅ Add middleware processing at appropriate phases
4. ✅ Add monitoring and metrics
5. ✅ Test end-to-end

### Phase 4: Devin Integration (Week 4)
1. ✅ Update Devin skill to send agent headers
2. ✅ Test rules enforcement via Router
3. ✅ Add plugin management UI (optional)
4. ✅ Document usage

---

## 🎯 Benefits vs Layer Approach

| Aspect | Layer Approach | Plugin Approach |
|--------|---------------|-----------------|
| **Coupling** | Tight coupling to Router core | Loosely coupled |
| **Enable/Disable** | Requires code changes | Config-based |
| **Testing** | Requires Router rebuild | Independent testing |
| **Hot-reload** | Requires Router restart | Plugin hot-reload |
| **Distribution** | Part of Router binary | Separate distribution |
| **Versioning** | Router version | Plugin version |
| **Isolation** | Crash affects Router | Isolated process |
| **Complexity** | Simpler initially | More complex initially |

---

## 📚 References

- **Provider Plugin System**: `apps/router/layers/provider/plugin/`
- **HashiCorp go-plugin**: https://github.com/hashicorp/go-plugin
- **Rules Directory**: `content/rules/`
- **BD Tool**: `Z:/02_CORE/_cli/bin/bd.exe`
- **Router Architecture**: `apps/router/layers/`

---

## 🛡️ Error Handling & Fallback Strategy

### Plugin Failure Modes

| Failure Mode | Description | Fallback Strategy |
|--------------|-------------|-------------------|
| **Plugin Crash** | Plugin process dies | Restart plugin + fail closed (block) or fail open (allow) |
| **Plugin Timeout** | Plugin doesn't respond within timeout | Use cached result (if available) or fail closed |
| **Plugin Error** | Plugin returns error | Log error + continue (if non-critical) or block |
| **Plugin Unhealthy** | Health check fails | Disable plugin + alert |
| **Network Failure** | RPC communication fails | Retry + fallback to degraded mode |

### Configuration

```yaml
middleware_plugins:
  error_handling:
    # Plugin failure behavior
    on_plugin_crash: "fail_closed"  # fail_closed | fail_open | retry
    on_plugin_timeout: "use_cache"   # use_cache | fail_closed | fail_open
    on_plugin_error: "continue"      # continue | block | log_only
    
    # Retry configuration
    max_retries: 3
    retry_delay: "1s"
    retry_backoff: "exponential"
    
    # Timeout configuration
    plugin_timeout: "10s"
    health_check_interval: "30s"
    
    # Degraded mode
    enable_degraded_mode: true
    degraded_mode_timeout: "5m"
```

### Implementation

**File**: `apps/router/layers/middleware/plugin/error_handler.go`

```go
package plugin

import (
	"context"
	"fmt"
	"time"
)

// ErrorHandler handles plugin errors and fallbacks.
type ErrorHandler struct {
	config ErrorHandlerConfig
	monitor *MonitoringService
}

type ErrorHandlerConfig struct {
	OnPluginCrash      string        `yaml:"on_plugin_crash"`
	OnPluginTimeout    string        `yaml:"on_plugin_timeout"`
	OnPluginError      string        `yaml:"on_plugin_error"`
	MaxRetries         int           `yaml:"max_retries"`
	RetryDelay         time.Duration `yaml:"retry_delay"`
	RetryBackoff       string        `yaml:"retry_backoff"`
	PluginTimeout      time.Duration `yaml:"plugin_timeout"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
	EnableDegradedMode bool          `yaml:"enable_degraded_mode"`
	DegradedModeTimeout time.Duration `yaml:"degraded_mode_timeout"`
}

// HandlePluginError handles a plugin error according to config.
func (h *ErrorHandler) HandlePluginError(ctx context.Context, pluginName string, err error, phase MiddlewarePhase) (*MiddlewareResult, error) {
	// Log error
	h.monitor.RecordPluginError(pluginName, phase, err)

	// Determine fallback strategy based on error type
	switch {
	case isPluginCrash(err):
		return h.handlePluginCrash(ctx, pluginName, err)
	case isPluginTimeout(err):
		return h.handlePluginTimeout(ctx, pluginName, err)
	default:
		return h.handlePluginError(ctx, pluginName, err)
	}
}

// handlePluginCrash handles a plugin crash.
func (h *ErrorHandler) handlePluginCrash(ctx context.Context, pluginName string, err error) (*MiddlewareResult, error) {
	switch h.config.OnPluginCrash {
	case "fail_closed":
		// Block request for security
		return &MiddlewareResult{
			Action:    ActionBlock,
			Message:   fmt.Sprintf("Plugin %s crashed - request blocked for security", pluginName),
			ErrorCode: 503,
		}, nil
	case "fail_open":
		// Allow request (degraded mode)
		return &MiddlewareResult{
			Action:  ActionAllow,
			Message: fmt.Sprintf("Plugin %s crashed - proceeding in degraded mode", pluginName),
		}, nil
	case "retry":
		// Attempt to restart plugin and retry
		// Implementation depends on plugin loader
		return nil, fmt.Errorf("plugin crashed, retry not implemented yet")
	default:
		return &MiddlewareResult{Action: ActionAllow}, nil
	}
}

// handlePluginTimeout handles a plugin timeout.
func (h *ErrorHandler) handlePluginTimeout(ctx context.Context, pluginName string, err error) (*MiddlewareResult, error) {
	switch h.config.OnPluginTimeout {
	case "use_cache":
		// Try to use cached result
		// Implementation depends on cache
		return &MiddlewareResult{
			Action:  ActionAllow,
			Message: fmt.Sprintf("Plugin %s timed out - using cached result if available", pluginName),
		}, nil
	case "fail_closed":
		return &MiddlewareResult{
			Action:    ActionBlock,
			Message:   fmt.Sprintf("Plugin %s timed out - request blocked", pluginName),
			ErrorCode: 504,
		}, nil
	case "fail_open":
		return &MiddlewareResult{
			Action:  ActionAllow,
			Message: fmt.Sprintf("Plugin %s timed out - proceeding in degraded mode", pluginName),
		}, nil
	default:
		return &MiddlewareResult{Action: ActionAllow}, nil
	}
}

// handlePluginError handles a general plugin error.
func (h *ErrorHandler) handlePluginError(ctx context.Context, pluginName string, err error) (*MiddlewareResult, error) {
	switch h.config.OnPluginError {
	case "continue":
		// Log error but continue
		return &MiddlewareResult{
			Action:  ActionAllow,
			Message: fmt.Sprintf("Plugin %s error: %v - continuing", pluginName, err),
		}, nil
	case "block":
		return &MiddlewareResult{
			Action:    ActionBlock,
			Message:   fmt.Sprintf("Plugin %s error: %v - request blocked", pluginName, err),
			ErrorCode: 500,
		}, nil
	case "log_only":
		// Only log, don't affect result
		return nil, err
	default:
		return nil, err
	}
}

func isPluginCrash(err error) bool {
	// Check if error indicates plugin crash
	return false
}

func isPluginTimeout(err error) bool {
	// Check if error indicates timeout
	return false
}
```

---

## 📊 Monitoring & Metrics

### Metrics to Track

| Metric | Type | Description |
|--------|------|-------------|
| `plugin_latency_ms` | Histogram | Plugin processing latency |
| `plugin_requests_total` | Counter | Total requests to plugin |
| `plugin_errors_total` | Counter | Total plugin errors |
| `plugin_cache_hits_total` | Counter | Cache hit count |
| `plugin_cache_misses_total` | Counter | Cache miss count |
| `plugin_violations_total` | Counter | Rule violations (for rules-enforcement) |
| `plugin_health_status` | Gauge | Plugin health status (1=healthy, 0=unhealthy) |
| `plugin_active_connections` | Gauge | Active plugin connections |

### Prometheus Integration

**File**: `apps/router/layers/middleware/plugin/monitoring.go`

```go
package plugin

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	pluginLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "plugin_latency_ms",
			Help:    "Plugin processing latency in milliseconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"plugin_name", "phase"},
	)

	pluginRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "plugin_requests_total",
			Help: "Total requests to plugin",
		},
		[]string{"plugin_name", "phase", "action"},
	)

	pluginErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "plugin_errors_total",
			Help: "Total plugin errors",
		},
		[]string{"plugin_name", "phase", "error_type"},
	)

	pluginCacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "plugin_cache_hits_total",
			Help: "Total cache hits",
		},
		[]string{"plugin_name"},
	)

	pluginCacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "plugin_cache_misses_total",
			Help: "Total cache misses",
		},
		[]string{"plugin_name"},
	)

	pluginViolations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "plugin_violations_total",
			Help: "Total rule violations",
		},
		[]string{"plugin_name", "rule_name", "severity"},
	)

	pluginHealthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "plugin_health_status",
			Help: "Plugin health status (1=healthy, 0=unhealthy)",
		},
		[]string{"plugin_name"},
	)
)

// MonitoringService handles plugin monitoring.
type MonitoringService struct{}

func (m *MonitoringService) RecordPluginLatency(pluginName string, phase MiddlewarePhase, latencyMs float64) {
	pluginLatency.WithLabelValues(pluginName, string(phase)).Observe(latencyMs)
}

func (m *MonitoringService) RecordPluginRequest(pluginName string, phase MiddlewarePhase, action MiddlewareAction) {
	pluginRequests.WithLabelValues(pluginName, string(phase), string(action)).Inc()
}

func (m *MonitoringService) RecordPluginError(pluginName string, phase MiddlewarePhase, errorType string) {
	pluginErrors.WithLabelValues(pluginName, string(phase), errorType).Inc()
}

func (m *MonitoringService) RecordCacheHit(pluginName string) {
	pluginCacheHits.WithLabelValues(pluginName).Inc()
}

func (m *MonitoringService) RecordCacheMiss(pluginName string) {
	pluginCacheMisses.WithLabelValues(pluginName).Inc()
}

func (m *MonitoringService) RecordViolation(pluginName, ruleName, severity string) {
	pluginViolations.WithLabelValues(pluginName, ruleName, severity).Inc()
}

func (m *MonitoringService) SetPluginHealth(pluginName string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	pluginHealthStatus.WithLabelValues(pluginName).Set(value)
}
```

---

## 💾 Plugin State Persistence

### State Types

| State Type | Storage | Purpose |
|------------|---------|---------|
| **Cache** | In-memory + SQLite | Validation cache |
| **Counters** | In-memory + SQLite | Request counts, violation counts |
| **Configuration** | File + SQLite | Plugin config history |
| **Audit Log** | SQLite + BD Tool | Violation logs |

### Implementation

**File**: `apps/router/layers/middleware/plugin/state.go`

```go
package plugin

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// StateManager manages plugin state persistence.
type StateManager struct {
	db     *sql.DB
	cache  map[string][]byte // In-memory cache
	mu     sync.RWMutex
}

// NewStateManager creates a new state manager.
func NewStateManager(dbPath string) (*StateManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, err
	}

	return &StateManager{
		db:    db,
		cache: make(map[string][]byte),
	}, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS plugin_cache (
			key TEXT PRIMARY KEY,
			value BLOB,
			created_at INTEGER,
			expires_at INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS plugin_counters (
			plugin_name TEXT,
			counter_name TEXT,
			value INTEGER,
			updated_at INTEGER,
			PRIMARY KEY (plugin_name, counter_name)
		)`,
		`CREATE TABLE IF NOT EXISTS plugin_config_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			plugin_name TEXT,
			config BLOB,
			applied_at INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS plugin_audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			plugin_name TEXT,
			event_type TEXT,
			event_data BLOB,
			timestamp INTEGER
		)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

// SaveCache saves a cache entry.
func (s *StateManager) SaveCache(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// In-memory cache
	s.cache[key] = value

	// Persistent cache
	expiresAt := time.Now().Add(ttl).Unix()
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO plugin_cache (key, value, created_at, expires_at) VALUES (?, ?, ?, ?)`,
		key, value, time.Now().Unix(), expiresAt,
	)

	return err
}

// LoadCache loads a cache entry.
func (s *StateManager) LoadCache(ctx context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check in-memory first
	if value, ok := s.cache[key]; ok {
		return value, nil
	}

	// Check persistent cache
	var value []byte
	var expiresAt int64
	err := s.db.QueryRowContext(ctx,
		`SELECT value, expires_at FROM plugin_cache WHERE key = ?`,
		key,
	).Scan(&value, &expiresAt)

	if err != nil {
		return nil, err
	}

	// Check expiration
	if time.Now().Unix() > expiresAt {
		return nil, nil
	}

	// Update in-memory cache
	s.cache[key] = value

	return value, nil
}

// IncrementCounter increments a counter.
func (s *StateManager) IncrementCounter(ctx context.Context, pluginName, counterName string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO plugin_counters (plugin_name, counter_name, value, updated_at)
		 VALUES (?, ?, 1, ?)
		 ON CONFLICT(plugin_name, counter_name)
		 DO UPDATE SET value = value + 1, updated_at = ?`,
		pluginName, counterName, time.Now().Unix(), time.Now().Unix(),
	)

	return err
}

// GetCounter gets a counter value.
func (s *StateManager) GetCounter(ctx context.Context, pluginName, counterName string) (int64, error) {
	var value int64
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM plugin_counters WHERE plugin_name = ? AND counter_name = ?`,
		pluginName, counterName,
	).Scan(&value)

	return value, err
}

// SaveConfigHistory saves plugin config history.
func (s *StateManager) SaveConfigHistory(ctx context.Context, pluginName string, config map[string]any) error {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO plugin_config_history (plugin_name, config, applied_at) VALUES (?, ?, ?)`,
		pluginName, configBytes, time.Now().Unix(),
	)

	return err
}

// LogAuditEvent logs an audit event.
func (s *StateManager) LogAuditEvent(ctx context.Context, pluginName, eventType string, eventData map[string]any) error {
	eventBytes, err := json.Marshal(eventData)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO plugin_audit_log (plugin_name, event_type, event_data, timestamp) VALUES (?, ?, ?, ?)`,
		pluginName, eventType, eventBytes, time.Now().Unix(),
	)

	return err
}
```

---

## 🔄 Plugin Hot-Reload Mechanism

### Reload Strategy

1. **Graceful Reload**:
   - Stop accepting new requests to plugin
   - Wait for in-flight requests to complete
   - Unload old plugin
   - Load new plugin
   - Resume accepting requests

2. **Rollback on Failure**:
   - If new plugin fails to load/initialize
   - Reload old plugin version
   - Alert operators

3. **Zero-Downtime**:
   - Use plugin versioning (v1, v2)
   - Load new version alongside old
   - Switch traffic gradually
   - Unload old version after successful migration

### Implementation

**File**: `apps/router/layers/middleware/plugin/reloader.go`

```go
package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Reloader handles plugin hot-reload.
type Reloader struct {
	loader       *MiddlewarePluginLoader
	stateManager *StateManager
	errorHandler *ErrorHandler
	mu           sync.RWMutex
	reloadQueue  chan ReloadRequest
}

type ReloadRequest struct {
	PluginName string
	Config     MiddlewarePluginConfig
	ResultChan chan ReloadResult
}

type ReloadResult struct {
	Success bool
	Error   error
	Version string
}

// NewReloader creates a new plugin reloader.
func NewReloader(loader *MiddlewarePluginLoader, stateManager *StateManager, errorHandler *ErrorHandler) *Reloader {
	r := &Reloader{
		loader:       loader,
		stateManager: stateManager,
		errorHandler: errorHandler,
		reloadQueue:  make(chan ReloadRequest, 10),
	}

	go r.processReloadQueue()

	return r
}

// ReloadPlugin initiates a plugin reload.
func (r *Reloader) ReloadPlugin(ctx context.Context, pluginName string, config MiddlewarePluginConfig) error {
	resultChan := make(chan ReloadResult, 1)

	r.reloadQueue <- ReloadRequest{
		PluginName: pluginName,
		Config:     config,
		ResultChan: resultChan,
	}

	select {
	case result := <-resultChan:
		if !result.Success {
			return result.Error
		}
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("reload timeout")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// processReloadQueue processes reload requests.
func (r *Reloader) processReloadQueue() {
	for req := range r.reloadQueue {
		result := r.doReload(req.PluginName, req.Config)
		req.ResultChan <- result
	}
}

// doReload performs the actual reload.
func (r *Reloader) doReload(pluginName string, config MiddlewarePluginConfig) ReloadResult {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Step 1: Get current plugin
	currentPlugin, err := r.loader.GetPlugin(pluginName)
	if err != nil {
		return ReloadResult{Success: false, Error: err}
	}

	// Step 2: Save current version info for rollback
	oldVersion := currentPlugin.Metadata.Version
	oldConfig := currentPlugin.Metadata.Config

	// Step 3: Unload current plugin
	ctx := context.Background()
	if err := r.loader.UnloadPlugin(ctx, pluginName); err != nil {
		return ReloadResult{Success: false, Error: fmt.Errorf("failed to unload plugin: %w", err)}
	}

	// Step 4: Load new plugin
	newPlugin, err := r.loader.LoadPlugin(ctx, config)
	if err != nil {
		// Rollback: reload old version
		rollbackConfig := MiddlewarePluginConfig{
			Name:   pluginName,
			Enabled: true,
			Config: oldConfig,
		}
		if _, rollbackErr := r.loader.LoadPlugin(ctx, rollbackConfig); rollbackErr != nil {
			// Critical: both new and old failed
			return ReloadResult{
				Success: false,
				Error:   fmt.Errorf("failed to load new plugin and rollback failed: %v (original error: %w)", rollbackErr, err),
			}
		}

		return ReloadResult{
			Success: false,
			Error:   fmt.Errorf("failed to load new plugin, rolled back to version %s: %w", oldVersion, err),
		}
	}

	// Step 5: Save config history
	if err := r.stateManager.SaveConfigHistory(ctx, pluginName, config.Config); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: Failed to save config history: %v\n", err)
	}

	return ReloadResult{
		Success: true,
		Version: newPlugin.Metadata.Version,
	}
}
```

---

## 👁️ Rule File Watching

### File Watching Strategy

Use `fsnotify` library to watch rule directory for changes:

- **Debounce**: Wait 1 second after last change before reloading
- **Validation**: Validate rules before applying
- **Rollback**: If validation fails, keep old rules

### Implementation

**File**: `apps/router/plugins/middleware/rules-enforcement/watcher.go`

```go
package main

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// RuleWatcher watches rule directory for changes.
type RuleWatcher struct {
	watcher   *fsnotify.Watcher
	validator *RuleValidator
	rulesPath string
	debounce  time.Duration
	lastEvent time.Time
}

// NewRuleWatcher creates a new rule watcher.
func NewRuleWatcher(validator *RuleValidator, rulesPath string) (*RuleWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Watch rules directory
	if err := watcher.Add(rulesPath); err != nil {
		watcher.Close()
		return nil, err
	}

	// Watch subdirectories
	entries, err := os.ReadDir(rulesPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				subPath := filepath.Join(rulesPath, entry.Name())
				watcher.Add(subPath)
			}
		}
	}

	rw := &RuleWatcher{
		watcher:   watcher,
		validator: validator,
		rulesPath: rulesPath,
		debounce:  1 * time.Second,
	}

	go rw.watch()

	return rw, nil
}

// watch watches for file system events.
func (rw *RuleWatcher) watch() {
	for {
		select {
		case event, ok := <-rw.watcher.Events:
			if !ok {
				return
			}

			// Only process create, write, remove events
			if event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Remove == fsnotify.Remove {
				rw.handleEvent(event)
			}

		case err, ok := <-rw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("[rules-enforcement] Watcher error: %v\n", err)
		}
	}
}

// handleEvent handles a file system event.
func (rw *RuleWatcher) handleEvent(event fsnotify.Event) {
	// Debounce: wait before processing
	now := time.Now()
	if now.Sub(rw.lastEvent) < rw.debounce {
		return
	}
	rw.lastEvent = now

	log.Printf("[rules-enforcement] Detected change: %s\n", event.Name)

	// Reload rules
	ctx := context.Background()
	if err := rw.validator.ReloadRules(ctx); err != nil {
		log.Printf("[rules-enforcement] Failed to reload rules: %v\n", err)
	} else {
		log.Printf("[rules-enforcement] Rules reloaded successfully\n")
	}
}

// Close closes the watcher.
func (rw *RuleWatcher) Close() error {
	return rw.watcher.Close()
}
```

---

## 🔗 Plugin Dependencies

### Dependency Resolution

If plugin A depends on plugin B:

1. **Declaration**: Plugin A declares dependencies in metadata
2. **Validation**: Loader validates dependency graph (no cycles)
3. **Loading Order**: Load dependencies before dependent plugins
4. **Failure Handling**: If dependency fails, dependent plugin fails to load

### Implementation

**File**: `apps/router/layers/middleware/plugin/dependencies.go`

```go
package plugin

import (
	"fmt"
	"sort"
)

// DependencyGraph represents plugin dependencies.
type DependencyGraph struct {
	nodes map[string]*PluginNode
	edges map[string][]string // plugin -> dependencies
}

type PluginNode struct {
	Name         string
	Config       MiddlewarePluginConfig
	Dependencies []string
	Loaded       bool
}

// NewDependencyGraph creates a new dependency graph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*PluginNode),
		edges: make(map[string][]string),
	}
}

// AddPlugin adds a plugin to the graph.
func (g *DependencyGraph) AddPlugin(config MiddlewarePluginConfig, dependencies []string) {
	g.nodes[config.Name] = &PluginNode{
		Name:         config.Name,
		Config:       config,
		Dependencies: dependencies,
		Loaded:       false,
	}
	g.edges[config.Name] = dependencies
}

// Validate validates the dependency graph (no cycles).
func (g *DependencyGraph) Validate() error {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for name := range g.nodes {
		if !visited[name] {
			if err := g.validateNode(name, visited, recursionStack); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateNode validates a node and its dependencies.
func (g *DependencyGraph) validateNode(name string, visited, recursionStack map[string]bool) error {
	visited[name] = true
	recursionStack[name] = true

	for _, dep := range g.edges[name] {
		if !visited[dep] {
			if err := g.validateNode(dep, visited, recursionStack); err != nil {
				return err
			}
		} else if recursionStack[dep] {
			return fmt.Errorf("circular dependency detected: %s -> %s", name, dep)
		}
	}

	recursionStack[name] = false
	return nil
}

// GetLoadOrder returns the order to load plugins (dependencies first).
func (g *DependencyGraph) GetLoadOrder() ([]string, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}

	// Topological sort
	order := make([]string, 0, len(g.nodes))
	visited := make(map[string]bool)

	for name := range g.nodes {
		if !visited[name] {
			g.topologicalSort(name, visited, &order)
		}
	}

	return order, nil
}

// topologicalSort performs topological sort.
func (g *DependencyGraph) topologicalSort(name string, visited map[string]bool, order *[]string) {
	visited[name] = true

	for _, dep := range g.edges[name] {
		if !visited[dep] {
			g.topologicalSort(dep, visited, order)
		}
	}

	*order = append(*order, name)
}
```

---

## 🔒 Plugin Security

### Security Measures

| Measure | Description |
|---------|-------------|
| **Resource Limits** | CPU, memory, file descriptor limits per plugin |
| **Permission Isolation** | Plugin runs with limited permissions |
| **Sandbox** | Plugin runs in isolated environment (chroot, namespace) |
| **Code Signing** | Only load signed plugins |
| **Network Isolation** | Control plugin network access |
| **File System Isolation** | Plugin only access allowed directories |

### Implementation

**File**: `apps/router/layers/middleware/plugin/security.go`

```go
package plugin

import (
	"context"
	"os/exec"
	"syscall"
)

// SecurityConfig holds security configuration.
type SecurityConfig struct {
	CPUQuota        string `yaml:"cpu_quota"`        // e.g., "0.5" for 50% CPU
	MemoryLimit     string `yaml:"memory_limit"`     // e.g., "512M"
	AllowedPaths    []string `yaml:"allowed_paths"`   // Paths plugin can access
	BlockNetwork    bool   `yaml:"block_network"`    // Block network access
	RequireSigned  bool   `yaml:"require_signed"`  // Require code signing
	SandboxEnabled bool   `yaml:"sandbox_enabled"`  // Enable sandbox
}

// SecurityManager applies security constraints to plugin processes.
type SecurityManager struct {
	config SecurityConfig
}

// NewSecurityManager creates a new security manager.
func NewSecurityManager(config SecurityConfig) *SecurityManager {
	return &SecurityManager{
		config: config,
	}
}

// ApplySecurity applies security constraints to a command.
func (s *SecurityManager) ApplySecurity(cmd *exec.Cmd) error {
	// Set resource limits
	if s.config.CPUQuota != "" {
		// Set CPU quota via cgroups
		// Implementation depends on OS
	}

	if s.config.MemoryLimit != "" {
		// Set memory limit via cgroups
		// Implementation depends on OS
	}

	// Set process attributes
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Set UID/GID for privilege dropping
		// Set chroot for filesystem isolation
		// Set namespace for process isolation
	}

	return nil
}

// ValidateSignature validates plugin signature.
func (s *SecurityManager) ValidateSignature(pluginPath string) error {
	if !s.config.RequireSigned {
		return nil
	}

	// Validate signature
	// Implementation depends on signing method

	return nil
}
```

---

## 📦 Plugin Version Compatibility

### Compatibility Matrix

| Router Version | Plugin Version | Compatible |
|----------------|----------------|------------|
| 1.0.0          | 1.0.0          | ✅         |
| 1.0.0          | 1.1.0          | ✅ (forward compatible) |
| 1.1.0          | 1.0.0          | ⚠️ (may work) |
| 1.1.0          | 0.9.0          | ❌ (incompatible) |

### Implementation

**File**: `apps/router/layers/middleware/plugin/version.go`

```go
package plugin

import (
	"fmt"
	"strings"
)

// Version represents a semantic version.
type Version struct {
	Major int
	Minor int
	Patch int
}

// ParseVersion parses a version string.
func ParseVersion(version string) (Version, error) {
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return Version{}, fmt.Errorf("invalid version format: %s", version)
	}

	var v Version
	_, err := fmt.Sscanf(parts[0], "%d", &v.Major)
	if err != nil {
		return Version{}, err
	}

	_, err = fmt.Sscanf(parts[1], "%d", &v.Minor)
	if err != nil {
		return Version{}, err
	}

	_, err = fmt.Sscanf(parts[2], "%d", &v.Patch)
	if err != nil {
		return Version{}, err
	}

	return v, nil
}

// IsCompatible checks if plugin version is compatible with router version.
func (v *Version) IsCompatible(routerVersion Version) bool {
	// Major version must match
	if v.Major != routerVersion.Major {
		return false
	}

	// Plugin minor version can be higher (forward compatible)
	// Plugin minor version can be lower (may work but not guaranteed)
	return true
}

// String returns string representation.
func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}
```

---

## 🗑️ Cache Invalidation Strategy

### Invalidation Triggers

| Trigger | Strategy |
|---------|----------|
| **TTL Expiration** | Automatic |
| **Rule File Change** | Invalidate all cache entries |
| **Manual Invalidation** | API endpoint |
| **Plugin Reload** | Invalidate all cache entries |
| **Memory Pressure** | LRU eviction |

### Implementation

**File**: `apps/router/layers/middleware/plugin/cache_invalidation.go`

```go
package plugin

import (
	"context"
	"time"
)

// CacheInvalidator handles cache invalidation.
type CacheInvalidator struct {
	cache *ValidationCache
}

// NewCacheInvalidator creates a new cache invalidator.
func NewCacheInvalidator(cache *ValidationCache) *CacheInvalidator {
	return &CacheInvalidator{
		cache: cache,
	}
}

// InvalidateAll invalidates all cache entries.
func (ci *CacheInvalidator) InvalidateAll(ctx context.Context) error {
	// Clear in-memory cache
	ci.cache.mu.Lock()
	ci.cache.store = make(map[string]CachedResponse)
	ci.cache.mu.Unlock()

	// Clear persistent cache (if any)
	// Implementation depends on state manager

	return nil
}

// InvalidateByKey invalidates cache entry by key.
func (ci *CacheInvalidator) InvalidateByKey(ctx context.Context, key string) error {
	ci.cache.mu.Lock()
	delete(ci.cache.store, key)
	ci.cache.mu.Unlock()

	return nil
}

// InvalidateByAgentType invalidates cache entries for a specific agent type.
func (ci *CacheInvalidator) InvalidateByAgentType(ctx context.Context, agentType string) error {
	ci.cache.mu.Lock()
	defer ci.cache.mu.Unlock()

	for key, entry := range ci.cache.store {
		if entry.AgentType == agentType {
			delete(ci.cache.store, key)
		}
	}

	return nil
}
```

---

## 🔄 Plugin Rollback

### Rollback Triggers

- Plugin fails to load after update
- Plugin health check fails after update
- Error rate increases after update
- Manual rollback request

### Implementation

**File**: `apps/router/layers/middleware/plugin/rollback.go`

```go
package plugin

import (
	"context"
	"fmt"
)

// RollbackManager handles plugin rollback.
type RollbackManager struct {
	stateManager *StateManager
	loader       *MiddlewarePluginLoader
}

// NewRollbackManager creates a new rollback manager.
func NewRollbackManager(stateManager *StateManager, loader *MiddlewarePluginLoader) *RollbackManager {
	return &RollbackManager{
		stateManager: stateManager,
		loader:       loader,
	}
}

// RollbackToVersion rolls back a plugin to a specific version.
func (rm *RollbackManager) RollbackToVersion(ctx context.Context, pluginName, version string) error {
	// Get config history
	config, err := rm.getConfigAtVersion(ctx, pluginName, version)
	if err != nil {
		return fmt.Errorf("failed to get config for version %s: %w", version, err)
	}

	// Reload plugin with old config
	if err := rm.loader.ReloadPlugin(ctx, pluginName, *config); err != nil {
		return fmt.Errorf("failed to rollback to version %s: %w", version, err)
	}

	return nil
}

// RollbackToPrevious rolls back to the previous version.
func (rm *RollbackManager) RollbackToPrevious(ctx context.Context, pluginName string) error {
	// Get previous version from config history
	// Implementation depends on state manager

	return nil
}

// getConfigAtVersion gets plugin config at a specific version.
func (rm *RollbackManager) getConfigAtVersion(ctx context.Context, pluginName, version string) (*MiddlewarePluginConfig, error) {
	// Query config history from state manager
	// Implementation depends on state manager

	return nil, nil
}
```

---

## 📝 Updated Configuration

**File**: `configs/router.yaml`

```yaml
# Middleware plugins configuration
middleware_plugins:
  enabled: true
  plugin_dir: "apps/router/plugins/middleware"
  
  # Error handling
  error_handling:
    on_plugin_crash: "fail_closed"
    on_plugin_timeout: "use_cache"
    on_plugin_error: "continue"
    max_retries: 3
    retry_delay: "1s"
    retry_backoff: "exponential"
    plugin_timeout: "10s"
    health_check_interval: "30s"
    enable_degraded_mode: true
    degraded_mode_timeout: "5m"
  
  # Security
  security:
    cpu_quota: "0.5"
    memory_limit: "512M"
    block_network: false
    require_signed: false
    sandbox_enabled: false
  
  # State persistence
  state:
    enabled: true
    db_path: "Z:/03_DATA/ti/plugin_state.db"
    backup_interval: "1h"
  
  # Monitoring
  monitoring:
    enabled: true
    prometheus_port: 9091
  
  plugins:
    - name: rules-enforcement
      enabled: true
      priority: 100
      phases:
        - post_auth
      dependencies: [] # No dependencies
      config:
        rules_path: "Z:/10_WORKPLACE/Ti/content/rules"
        cache_ttl: "5m"
        strict_mode: false
        block_critical: true
        warn_non_critical: true
        log_violations: true
        bd_tool_path: "Z:/02_CORE/_cli/bin/bd.exe"
        watch_rules: true # Enable file watching
```

---

## 🚀 Updated Implementation Plan

### Phase 1: Middleware Plugin System (Week 1)
1. ✅ Create `layers/middleware/plugin/` package
2. ✅ Implement `MiddlewarePlugin` interface
3. ✅ Implement `MiddlewarePluginLoader`
4. ✅ Implement plugin discovery and loading
5. ✅ Add error handling & fallback
6. ✅ Add monitoring & metrics
7. ✅ Add unit tests

### Phase 2: Rules Enforcement Plugin (Week 2)
1. ✅ Create `plugins/middleware/rules-enforcement/` plugin
2. ✅ Implement agent detection
3. ✅ Implement rule validation
4. ✅ Implement validation cache
5. ✅ Implement violation logging (BD tool)
6. ✅ Implement rule file watching
7. ✅ Build plugin binary

### Phase 3: Router Integration (Week 3)
1. ✅ Integrate middleware loader into Router handler
2. ✅ Add configuration support
3. ✅ Add middleware processing at appropriate phases
4. ✅ Add state persistence
5. ✅ Add hot-reload mechanism
6. ✅ Add rollback mechanism
7. ✅ Test end-to-end

### Phase 4: Advanced Features (Week 4)
1. ✅ Implement plugin dependencies
2. ✅ Implement plugin security
3. ✅ Implement version compatibility checks
4. ✅ Implement cache invalidation
5. ✅ Add plugin management UI (optional)
6. ✅ Document usage

### Phase 5: Devin Integration (Week 5)
1. ✅ Update Devin skill to send agent headers
2. ✅ Test rules enforcement via Router
3. ✅ Test error scenarios
4. ✅ Test hot-reload
5. ✅ Test rollback
6. ✅ Performance testing

---

## ✅ Plugin Validation

### Validation Checklist

Validate plugin before loading:

| Check | Description | Fail Action |
|-------|-------------|-------------|
| **Binary Exists** | Plugin binary exists and is executable | Block load |
| **Signature Valid** | Plugin signature is valid (if required) | Block load |
| **Interface Implements** | Plugin implements MiddlewarePlugin interface | Block load |
| **Dependencies Met** | All dependencies are available | Block load |
| **Version Compatible** | Plugin version compatible with Router version | Warn or block |
| **Config Valid** | Plugin configuration is valid | Block load |
| **Resources Available** | Required resources (CPU, memory) available | Warn or block |
| **Security Check** | Plugin passes security checks | Block load |

### Implementation

**File**: `apps/router/layers/middleware/plugin/validator.go`

```go
package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// PluginValidator validates plugins before loading.
type PluginValidator struct {
	securityManager *SecurityManager
	routerVersion   Version
	config          ValidationConfig
}

type ValidationConfig struct {
	RequireSignature    bool   `yaml:"require_signature"`
	StrictVersionCheck  bool   `yaml:"strict_version_check"`
	CheckDependencies   bool   `yaml:"check_dependencies"`
	CheckResources      bool   `yaml:"check_resources"`
	EnableSecurityCheck bool   `yaml:"enable_security_check"`
}

// ValidationResult holds validation results.
type ValidationResult struct {
	Valid  bool
	Errors []string
	Warnings []string
}

// NewPluginValidator creates a new plugin validator.
func NewPluginValidator(securityManager *SecurityManager, routerVersion Version, config ValidationConfig) *PluginValidator {
	return &PluginValidator{
		securityManager: securityManager,
		routerVersion:   routerVersion,
		config:          config,
	}
}

// ValidatePlugin validates a plugin before loading.
func (v *PluginValidator) ValidatePlugin(ctx context.Context, pluginPath string, config MiddlewarePluginConfig) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check 1: Binary exists
	if err := v.checkBinaryExists(pluginPath); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Binary check failed: %v", err))
	}

	// Check 2: Signature valid
	if v.config.RequireSignature {
		if err := v.securityManager.ValidateSignature(pluginPath); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Signature validation failed: %v", err))
		}
	}

	// Check 3: Interface implements (load and check)
	if err := v.checkInterfaceImplements(ctx, pluginPath); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Interface check failed: %v", err))
	}

	// Check 4: Dependencies met
	if v.config.CheckDependencies {
		if deps := v.checkDependencies(config); len(deps) > 0 {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Unmet dependencies: %v", deps))
		}
	}

	// Check 5: Version compatible
	if v.config.StrictVersionCheck {
		if err := v.checkVersionCompatibility(config); err != nil {
			if v.config.StrictVersionCheck {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("Version check failed: %v", err))
			} else {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Version check warning: %v", err))
			}
		}
	}

	// Check 6: Config valid
	if err := v.checkConfigValid(config); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Config validation failed: %v", err))
	}

	// Check 7: Resources available
	if v.config.CheckResources {
		if err := v.checkResourcesAvailable(); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Resource check warning: %v", err))
		}
	}

	// Check 8: Security check
	if v.config.EnableSecurityCheck {
		if err := v.securityCheck(pluginPath); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Security check failed: %v", err))
		}
	}

	return result
}

// checkBinaryExists checks if plugin binary exists and is executable.
func (v *PluginValidator) checkBinaryExists(pluginPath string) error {
	info, err := os.Stat(pluginPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("plugin binary not found: %s", pluginPath)
	}
	if err != nil {
		return fmt.Errorf("failed to stat plugin: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("plugin path is a directory, not a file: %s", pluginPath)
	}
	if runtime.GOOS != "windows" {
		if info.Mode()&0111 == 0 {
			return fmt.Errorf("plugin binary is not executable: %s", pluginPath)
		}
	}
	return nil
}

// checkInterfaceImplements checks if plugin implements MiddlewarePlugin interface.
func (v *PluginValidator) checkInterfaceImplements(ctx context.Context, pluginPath string) error {
	// Load plugin temporarily to check interface
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         PluginMap,
		Cmd:             exec.Command(pluginPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("failed to connect to plugin: %w", err)
	}

	raw, err := rpcClient.Dispense("middleware")
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}

	_, ok := raw.(MiddlewarePlugin)
	client.Kill()

	if !ok {
		return fmt.Errorf("plugin does not implement MiddlewarePlugin interface")
	}

	return nil
}

// checkDependencies checks if plugin dependencies are met.
func (v *PluginValidator) checkDependencies(config MiddlewarePluginConfig) []string {
	// Check if declared dependencies are available
	// Implementation depends on dependency graph
	return []string{}
}

// checkVersionCompatibility checks plugin version compatibility.
func (v *PluginValidator) checkVersionCompatibility(config MiddlewarePluginConfig) error {
	pluginVersion, err := ParseVersion(config.Version)
	if err != nil {
		return fmt.Errorf("invalid plugin version: %w", err)
	}

	if !pluginVersion.IsCompatible(v.routerVersion) {
		return fmt.Errorf("plugin version %s is not compatible with router version %s",
			pluginVersion.String(), v.routerVersion.String())
	}

	return nil
}

// checkConfigValid validates plugin configuration.
func (v *PluginValidator) checkConfigValid(config MiddlewarePluginConfig) error {
	// Validate required fields
	if config.Name == "" {
		return fmt.Errorf("plugin name is required")
	}

	// Validate plugin-specific config
	// Implementation depends on plugin

	return nil
}

// checkResourcesAvailable checks if required resources are available.
func (v *PluginValidator) checkResourcesAvailable() error {
	// Check CPU, memory, disk space
	// Implementation depends on OS

	return nil
}

// securityCheck performs security checks on plugin.
func (v *PluginValidator) securityCheck(pluginPath string) error {
	// Check for malicious code, suspicious patterns
	// Implementation depends on security policy

	return nil
}
```

---

## 🛑 Plugin Graceful Shutdown

### Shutdown Strategy

When Router stops or plugin is unloaded:

1. **Stop accepting new requests** to plugin
2. **Wait for in-flight requests** to complete (with timeout)
3. **Call ShutdownPlugin()** to let plugin cleanup
4. **Kill plugin process**
5. **Wait for process to exit** (with timeout)
6. **Force kill** if timeout exceeded

### Implementation

**File**: `apps/router/layers/middleware/plugin/shutdown.go`

```go
package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ShutdownManager handles graceful plugin shutdown.
type ShutdownManager struct {
	loader         *MiddlewarePluginLoader
	inflightReqs   map[string]context.CancelFunc
	inflightMutex  sync.RWMutex
	shutdownTimeout time.Duration
}

// NewShutdownManager creates a new shutdown manager.
func NewShutdownManager(loader *MiddlewarePluginLoader, shutdownTimeout time.Duration) *ShutdownManager {
	return &ShutdownManager{
		loader:         loader,
		inflightReqs:   make(map[string]context.CancelFunc),
		shutdownTimeout: shutdownTimeout,
	}
}

// RegisterInflightRequest registers an in-flight request.
func (sm *ShutdownManager) RegisterInflightRequest(requestID string, cancel context.CancelFunc) {
	sm.inflightMutex.Lock()
	defer sm.inflightMutex.Unlock()
	sm.inflightReqs[requestID] = cancel
}

// UnregisterInflightRequest unregisters an in-flight request.
func (sm *ShutdownManager) UnregisterInflightRequest(requestID string) {
	sm.inflightMutex.Lock()
	defer sm.inflightMutex.Unlock()
	delete(sm.inflightReqs, requestID)
}

// ShutdownPlugin gracefully shuts down a plugin.
func (sm *ShutdownManager) ShutdownPlugin(ctx context.Context, pluginName string) error {
	// Step 1: Mark plugin as shutting down
	// Implementation: set flag in plugin wrapper

	// Step 2: Wait for in-flight requests
	if err := sm.waitForInflightRequests(ctx, pluginName); err != nil {
		return fmt.Errorf("failed to wait for in-flight requests: %w", err)
	}

	// Step 3: Get plugin
	plugin, err := sm.loader.GetPlugin(pluginName)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	// Step 4: Call ShutdownPlugin
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := plugin.Plugin.ShutdownPlugin(shutdownCtx); err != nil {
		// Log but continue
		fmt.Printf("Warning: Plugin shutdown failed: %v\n", err)
	}

	// Step 5: Kill plugin process
	if err := sm.loader.UnloadPlugin(ctx, pluginName); err != nil {
		return fmt.Errorf("failed to unload plugin: %w", err)
	}

	return nil
}

// ShutdownAll gracefully shuts down all plugins.
func (sm *ShutdownManager) ShutdownAll(ctx context.Context) error {
	plugins := sm.loader.ListPlugins()

	for _, metadata := range plugins {
		if err := sm.ShutdownPlugin(ctx, metadata.Name); err != nil {
			fmt.Printf("Warning: Failed to shutdown plugin %s: %v\n", metadata.Name, err)
		}
	}

	return nil
}

// waitForInflightRequests waits for in-flight requests to complete.
func (sm *ShutdownManager) waitForInflightRequests(ctx context.Context, pluginName string) error {
	// Wait for all in-flight requests to complete
	// Implementation: poll inflightReqs map

	timeout := time.After(sm.shutdownTimeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// Timeout: cancel remaining in-flight requests
			sm.cancelInflightRequests(pluginName)
			return fmt.Errorf("timeout waiting for in-flight requests")
		case <-ticker.C:
			sm.inflightMutex.RLock()
			count := len(sm.inflightReqs)
			sm.inflightMutex.RUnlock()

			if count == 0 {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// cancelInflightRequests cancels all in-flight requests for a plugin.
func (sm *ShutdownManager) cancelInflightRequests(pluginName string) {
	sm.inflightMutex.Lock()
	defer sm.inflightMutex.Unlock()

	for requestID, cancel := range sm.inflightReqs {
		cancel()
		delete(sm.inflightReqs, requestID)
	}
}
```

---

## 🧹 Plugin Resource Cleanup

### Cleanup Strategy

When plugin is unloaded or crashes:

1. **Close open files** and file descriptors
2. **Close network connections**
3. **Release memory allocations**
4. **Cleanup temporary files**
5. **Flush pending writes**
6. **Release locks/mutexes**

### Implementation

**File**: `apps/router/layers/middleware/plugin/cleanup.go`

```go
package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// CleanupManager handles plugin resource cleanup.
type CleanupManager struct {
	stateManager *StateManager
	tempDir      string
}

// NewCleanupManager creates a new cleanup manager.
func NewCleanupManager(stateManager *StateManager, tempDir string) *CleanupManager {
	return &CleanupManager{
		stateManager: stateManager,
		tempDir:     tempDir,
	}
}

// CleanupPlugin cleans up resources for a plugin.
func (cm *CleanupManager) CleanupPlugin(ctx context.Context, pluginName string) error {
	var errors []error

	// Cleanup 1: Close open files
	if err := cm.closeOpenFiles(ctx, pluginName); err != nil {
		errors = append(errors, fmt.Errorf("failed to close files: %w", err))
	}

	// Cleanup 2: Close network connections
	if err := cm.closeNetworkConnections(ctx, pluginName); err != nil {
		errors = append(errors, fmt.Errorf("failed to close connections: %w", err))
	}

	// Cleanup 3: Cleanup temporary files
	if err := cm.cleanupTempFiles(ctx, pluginName); err != nil {
		errors = append(errors, fmt.Errorf("failed to cleanup temp files: %w", err))
	}

	// Cleanup 4: Flush pending state
	if err := cm.flushPendingState(ctx, pluginName); err != nil {
		errors = append(errors, fmt.Errorf("failed to flush state: %w", err))
	}

	// Cleanup 5: Invalidate cache
	if err := cm.invalidateCache(ctx, pluginName); err != nil {
		errors = append(errors, fmt.Errorf("failed to invalidate cache: %w", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup completed with %d errors: %v", len(errors), errors)
	}

	return nil
}

// closeOpenFiles closes open files for a plugin.
func (cm *CleanupManager) closeOpenFiles(ctx context.Context, pluginName string) error {
	// Track open files per plugin
	// Close all tracked files
	return nil
}

// closeNetworkConnections closes network connections for a plugin.
func (cm *CleanupManager) closeNetworkConnections(ctx context.Context, pluginName string) error {
	// Track network connections per plugin
	// Close all tracked connections
	return nil
}

// cleanupTempFiles cleans up temporary files for a plugin.
func (cm *CleanupManager) cleanupTempFiles(ctx context.Context, pluginName string) error {
	pluginTempDir := filepath.Join(cm.tempDir, pluginName)

	// Remove plugin temp directory
	if err := os.RemoveAll(pluginTempDir); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// flushPendingState flushes pending state for a plugin.
func (cm *CleanupManager) flushPendingState(ctx context.Context, pluginName string) error {
	// Flush any pending writes to state manager
	return nil
}

// invalidateCache invalidates cache for a plugin.
func (cm *CleanupManager) invalidateCache(ctx context.Context, pluginName string) error {
	// Invalidate all cache entries for this plugin
	return nil
}

// CleanupAll cleans up resources for all plugins.
func (cm *CleanupManager) CleanupAll(ctx context.Context) error {
	// Cleanup temp directory
	if err := os.RemoveAll(cm.tempDir); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
```

---

## 📝 Updated Configuration (MVP)

**File**: `configs/router.yaml`

```yaml
# Middleware plugins configuration
middleware_plugins:
  enabled: true
  plugin_dir: "apps/router/plugins/middleware"
  
  # Validation
  validation:
    require_signature: false
    strict_version_check: false
    check_dependencies: true
    check_resources: true
    enable_security_check: false
  
  # Error handling
  error_handling:
    on_plugin_crash: "fail_closed"
    on_plugin_timeout: "use_cache"
    on_plugin_error: "continue"
    max_retries: 3
    retry_delay: "1s"
    retry_backoff: "exponential"
    plugin_timeout: "10s"
    health_check_interval: "30s"
    enable_degraded_mode: true
    degraded_mode_timeout: "5m"
  
  # Shutdown
  shutdown:
    timeout: "30s"
    wait_for_inflight: true
    inflight_timeout: "10s"
  
  # Cleanup
  cleanup:
    temp_dir: "Z:/03_DATA/ti/plugin_temp"
    auto_cleanup: true
    cleanup_on_error: true
  
  # Security
  security:
    cpu_quota: "0.5"
    memory_limit: "512M"
    block_network: false
    require_signed: false
    sandbox_enabled: false
  
  # State persistence
  state:
    enabled: true
    db_path: "Z:/03_DATA/ti/plugin_state.db"
    backup_interval: "1h"
  
  # Monitoring
  monitoring:
    enabled: true
    prometheus_port: 9091
  
  plugins:
    - name: rules-enforcement
      enabled: true
      priority: 100
      phases:
        - post_auth
      dependencies: []
      config:
        rules_path: "Z:/10_WORKPLACE/Ti/content/rules"
        cache_ttl: "5m"
        strict_mode: false
        block_critical: true
        warn_non_critical: true
        log_violations: true
        bd_tool_path: "Z:/02_CORE/_cli/bin/bd.exe"
        watch_rules: true
```

---

## 🚀 Updated Implementation Plan (MVP)

### Phase 1: Middleware Plugin System (Week 1)
1. ✅ Create `layers/middleware/plugin/` package
2. ✅ Implement `MiddlewarePlugin` interface
3. ✅ Implement `MiddlewarePluginLoader`
4. ✅ Implement plugin discovery and loading
5. ✅ Implement **Plugin Validation** ← NEW
6. ✅ Add error handling & fallback
7. ✅ Add monitoring & metrics
8. ✅ Add unit tests

### Phase 2: Rules Enforcement Plugin (Week 2)
1. ✅ Create `plugins/middleware/rules-enforcement/` plugin
2. ✅ Implement agent detection
3. ✅ Implement rule validation
4. ✅ Implement validation cache
5. ✅ Implement violation logging (BD tool)
6. ✅ Implement rule file watching
7. ✅ Build plugin binary

### Phase 3: Router Integration (Week 3)
1. ✅ Integrate middleware loader into Router handler
2. ✅ Add configuration support
3. ✅ Add middleware processing at appropriate phases
4. ✅ Add state persistence
5. ✅ Add hot-reload mechanism
6. ✅ Add rollback mechanism
7. ✅ Implement **Graceful Shutdown** ← NEW
8. ✅ Implement **Resource Cleanup** ← NEW
9. ✅ Test end-to-end

### Phase 4: Devin Integration (Week 4)
1. ✅ Update Devin skill to send agent headers
2. ✅ Test rules enforcement via Router
3. ✅ Test error scenarios
4. ✅ Test hot-reload
5. ✅ Test rollback
6. ✅ Test graceful shutdown
7. ✅ Test resource cleanup
8. ✅ Performance testing

---

## 📚 MVP Components Summary

| Component | Status | Priority |
|-----------|--------|----------|
| Middleware Plugin Interface | ✅ Designed | P0 |
| Middleware PluginLoader | ✅ Designed | P0 |
| Plugin Validation | ✅ Designed | P0 |
| Error Handling & Fallback | ✅ Designed | P0 |
| Monitoring & Metrics | ✅ Designed | P0 |
| State Persistence | ✅ Designed | P0 |
| Hot-Reload | ✅ Designed | P0 |
| Rollback | ✅ Designed | P0 |
| Graceful Shutdown | ✅ Designed | P0 |
| Resource Cleanup | ✅ Designed | P0 |
| Rule File Watching | ✅ Designed | P0 |
| Rules Enforcement Plugin | ✅ Designed | P0 |

**Total MVP Components**: 12 components

**Remaining (Future)**: Testing strategy, distribution, debugging, lifecycle events, RPC detail, documentation, remote discovery, marketplace
