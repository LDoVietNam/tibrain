# Ti CLI Codebase Analysis - Migration Assessment (Plugin-Based)

> **Last Updated**: 2026-05-05
> **Version**: 2.0.0 (Plugin-Based)
> **Purpose**: Assess current codebase vs plugin-based deployment strategy requirements

---

## 📊 Current State Assessment

### ✅ What Already Exists (Excellent Foundation)

#### 1. Plugin System
**Location:** `internal/core/platform.go`
```go
type Plugin interface {
    Name() string
    Version() string
    Metadata() PluginMetadata
    Initialize(ctx context.Context, config map[string]string) error
    Execute(ctx context.Context, task string, input map[string]interface{}) (map[string]interface{}, error)
    Shutdown(ctx context.Context) error
    HealthCheck(ctx context.Context) error
    State() PluginState
}

type Platform interface {
    LoadPlugin(ctx context.Context, name string) error
    UnloadPlugin(ctx context.Context, name string) error
    GetPlugin(name string) (Plugin, error)
    ListPlugins() []PluginInfo
    ExecutePlugin(ctx context.Context, pluginName, task string, input map[string]interface{}) (map[string]interface{}, error)
    Shutdown(ctx context.Context) error
}
```

**Status:** ✅ **EXISTING & EXCELLENT**
- Full Plugin interface with lifecycle methods
- Platform for plugin management
- Health check support
- State management (PluginState enum with 10 states)

**Gap:** Missing Dependencies() method for dependency resolution, plugin registry with metadata, version compatibility checks

---

#### 2. Configuration System
**Location:** `internal/kernel/config.go`
```go
// Precedence (lowest to highest):
// 1. Defaults (built-in)
// 2. Global (~/.config/ti/config.json)
// 3. Project (./ti.json in CWD)
// 4. Portable (./.ti/config.json in CWD)
// 5. Custom (TI_CONFIG env → single file)
// 6. Environment variables
// 7. Inline flags (applied by callers at runtime)
```

**Status:** ✅ **EXISTING & SUPERIOR**
- 7-level config precedence (better than OpenCode's 7)
- Validation with LoadReport
- Hash-based change detection
- Multiple config sources

**Gap:** Missing plugin-specific config, hot-reload, feature flag config

---

#### 3. Testing Infrastructure
**Count:** 100+ test files
**Examples:**
- `e2e/e2e_test.go` - E2E test suite
- `internal/*/ *_test.go` - Unit tests
- `internal/integration/*_test.go` - Integration tests

**Status:** ✅ **EXISTING & COMPREHENSIVE**
- Unit tests
- Integration tests
- E2E tests
- Test harness (`internal/testkit/harness_test.go`)

**Gap:** Missing feature flag tests, plugin compatibility tests, canary deployment tests

---

#### 4. Observability
**Location:** `internal/observability/observability.go`
```go
type EventLogger struct {
    file      *os.File
    mu        sync.Mutex
    sessionID string
    filePath  string
}

type Event struct {
    ID        string
    Type      EventType
    SessionID string
    Timestamp time.Time
    Data      map[string]interface{}
}
```

**Status:** ✅ **EXISTING & FUNCTIONAL**
- Event logging to JSONL
- Multiple event types (message, tool_call, error, checkpoint, metric)
- Session tracking

**Gap:** Missing structured logging, log levels, remote logging, distributed tracing

---

#### 5. Metrics
**Location:** `internal/metrics/metrics.go`
```go
type Metrics struct {
    // Uptime
    StartTime time.Time
    
    // Heartbeats
    HeartbeatsSentTotal   int64
    HeartbeatsSentSuccess int64
    HeartbeatsSentFailed  int64
    
    // Connections
    NATSConnected      bool
    RouterConnected    bool
    FallbackMode       bool
    
    // Mode usage
    ModeUsageTotal map[string]int64
    
    // Sessions
    ActiveSessions int64
    TotalSessions   int64
}
```

**Status:** ✅ **EXISTING & BASIC**
- Basic metrics collection
- Global metrics instance
- HTTP metrics endpoint

**Gap:** Missing histogram metrics, labels, plugin-specific metrics, cost metrics

---

### ❌ What's Missing (Needs Implementation)

#### 1. Feature Flag System
**Status:** ❌ **MISSING**
- No feature flag service
- No rollout percentage logic
- No user/team whitelisting
- No A/B testing support

**Effort:** 2-3 weeks

---

#### 2. Plugin Dependency Management
**Status:** ❌ **MISSING**
- No dependency resolution
- No version compatibility checks
- No plugin registry with metadata
- No checksum verification

**Effort:** 2-3 weeks

---

#### 3. Hot-Reload Configuration
**Status:** ❌ **MISSING**
- No config file watcher
- No hot-reload capability
- No config change callbacks
- No dynamic feature toggling

**Effort:** 1-2 weeks

---

#### 4. Structured Logging
**Status:** ❌ **MISSING**
- No log levels (DEBUG, INFO, WARN, ERROR)
- No structured log fields
- No multiple log outputs (console, file, remote)
- No log aggregation

**Effort:** 2-3 weeks

---

#### 5. Distributed Tracing
**Status:** ❌ **MISSING**
- No trace ID generation
- No span management
- No trace export
- No trace context propagation

**Effort:** 3-4 weeks

---

#### 6. CI/CD Pipeline
**Status:** ❌ **MISSING**
- No GitHub workflows
- No automated testing
- No multi-platform builds
- No security scanning
- No automated deployment

**Effort:** 2-3 weeks

---

#### 7. Health Check System
**Status:** ⚠️ **PARTIAL**
- Plugin health check exists
- Router health check exists
- No comprehensive health check command
- No health check aggregation

**Effort:** 1 week

---

#### 8. Service Locator / DI Container
**Status:** ❌ **MISSING**
- No service locator pattern
- No dependency injection container
- Plugins depend on concrete implementations
- Tight coupling between components

**Effort:** 2-3 weeks

---

## 🔄 Migration Strategy (Plugin-Based)

### Phase 1: Plugin Foundation (Weeks 1-4) - Low Risk

**Goal:** Enhance plugin system and implement foundational features as plugins

**Tasks:**
1. **Plugin Interface Enhancement** (Week 1)
   - Add Dependencies() method to Plugin interface
   - Add plugin metadata fields (author, description, capabilities)
   - Add plugin version compatibility checks
   - **Risk:** Low - interface extension is backward compatible

2. **FeatureFlagPlugin Implementation** (Week 1-2)
   - Create `internal/plugins/featureflag/` package
   - Implement FeatureFlagPlugin with Plugin interface
   - Add feature flag config file (~/.ti-cli/plugins/featureflag.yaml)
   - Implement rollout logic (percentage, user/team whitelisting)
   - **Risk:** Low - new plugin, optional feature

3. **StructuredLoggingPlugin Implementation** (Week 2-3)
   - Create `internal/plugins/logging/` package
   - Implement StructuredLoggingPlugin with Plugin interface
   - Add log levels (DEBUG, INFO, WARN, ERROR)
   - Add structured log fields support
   - Add multiple log outputs (console, file, remote)
   - **Risk:** Low - new plugin, optional feature

4. **HotReloadConfigPlugin Implementation** (Week 3-4)
   - Create `internal/plugins/configwatcher/` package
   - Implement HotReloadConfigPlugin with Plugin interface
   - Add fsnotify dependency
   - Implement config file watching
   - Add config change callbacks
   - **Risk:** Low - new plugin, optional feature

5. **Plugin Config Files** (Week 4)
   - Create ~/.ti-cli/plugins/ directory structure
   - Add plugin config files (featureflag.yaml, logging.yaml, configwatcher.yaml)
   - Add plugin registry (plugin-index.json)
   - Implement plugin auto-discovery
   - **Risk:** Low - configuration only

**Deliverables:**
- Extended Plugin interface with Dependencies() method
- `internal/plugins/featureflag/plugin.go`
- `internal/plugins/logging/plugin.go`
- `internal/plugins/configwatcher/plugin.go`
- Plugin config files in ~/.ti-cli/plugins/
- Plugin registry system

**Testing:**
- Plugin interface backward compatibility tests
- FeatureFlagPlugin unit tests
- StructuredLoggingPlugin unit tests
- HotReloadConfigPlugin unit tests
- Plugin loading/unloading tests
- Plugin dependency resolution tests

---

### Phase 2: Dependency Injection & Plugin Refactoring (Weeks 5-8) - Medium Risk

**Goal:** Implement DI container and refactor existing plugins to use it

**Tasks:**
1. **Service Locator / DI Container** (Week 5-6)
   - Create `internal/di/` package
   - Implement ServiceLocator interface
   - Implement DI container with service registration
   - Add service dependency resolution
   - **Risk:** Medium - new infrastructure

2. **Plugin Registry Enhancement** (Week 6-7)
   - Implement plugin registry with metadata
   - Add plugin version management (semver)
   - Add plugin dependency resolution
   - Add plugin checksum verification
   - Add plugin auto-discovery from ~/.ti-cli/plugins/
   - **Risk:** Medium - new infrastructure

3. **Refactor Existing Plugins** (Week 7-8)
   - Refactor existing plugins to use DI container
   - Update plugin configs to use new format
   - Test plugin compatibility with DI
   - Add plugin dependency declarations
   - **Risk:** Medium - refactoring existing code

**Deliverables:**
- `internal/di/locator.go`
- `internal/di/container.go`
- `internal/plugins/registry.go`
- Updated all existing plugins
- Plugin registry with metadata

**Testing:**
- DI container tests
- Plugin registry tests
- Plugin dependency resolution tests
- Existing plugin compatibility tests

---

### Phase 3: Observability Plugins (Weeks 9-12) - Low Risk

**Goal:** Implement observability features as plugins

**Tasks:**
1. **DistributedTracingPlugin Implementation** (Week 9-10)
   - Create `internal/plugins/tracing/` package
   - Implement DistributedTracingPlugin with Plugin interface
   - Add trace ID generation
   - Add span management
   - Add trace export (Jaeger/Zipkin compatible)
   - Add plugin config (~/.ti-cli/plugins/tracing.yaml)
   - **Risk:** Low - new plugin, optional feature

2. **EnhancedMetricsPlugin Implementation** (Week 10-11)
   - Create `internal/plugins/metrics/` package
   - Implement EnhancedMetricsPlugin with Plugin interface
   - Add histogram metrics
   - Add labels support
   - Add plugin-specific metrics
   - Add cost metrics
   - Add plugin config (~/.ti-cli/plugins/metrics.yaml)
   - **Risk:** Low - new plugin, optional feature

3. **LogAggregationPlugin Implementation** (Week 11-12)
   - Create `internal/plugins/logagg/` package
   - Implement LogAggregationPlugin with Plugin interface
   - Add remote log output (ELK/Loki compatible)
   - Add log aggregation
   - Add log filtering
   - Add log retention policy
   - Add plugin config (~/.ti-cli/plugins/logagg.yaml)
   - **Risk:** Low - new plugin, optional feature

**Deliverables:**
- `internal/plugins/tracing/plugin.go`
- `internal/plugins/metrics/plugin.go`
- `internal/plugins/logagg/plugin.go`
- Plugin config files for all 3 plugins

**Testing:**
- Tracing plugin unit tests
- Metrics plugin unit tests
- Log aggregation plugin unit tests
- Integration tests for all 3 plugins

---

### Phase 4: CI/CD & Plugin Deployment (Weeks 13-16) - Medium Risk

**Goal:** Add CI/CD pipeline for plugin-based deployment

**Tasks:**
1. **GitHub Workflows** (Week 13-14)
   - Create `.github/workflows/test.yml` with plugin-specific tests
   - Create `.github/workflows/build.yml` with plugin builds
   - Create `.github/workflows/deploy.yml` with plugin registry deployment
   - Add multi-platform builds for CLI and plugins
   - **Risk:** Medium - new infrastructure

2. **Security Scanning** (Week 14-15)
   - Add gosec scanning for core and plugins
   - Add dependency scanning
   - Add vulnerability scanning
   - Add security audit for plugins
   - **Risk:** Low - non-blocking checks

3. **Plugin Canary Deployment** (Week 15-16)
   - Implement plugin canary deployment logic
   - Add plugin rollout percentage logic
   - Add automated rollback triggers for plugins
   - Add plugin deployment monitoring
   - Add plugin registry CDN deployment
   - **Risk:** Medium - deployment infrastructure

**Deliverables:**
- `.github/workflows/` directory
- Security scanning configuration
- Plugin canary deployment scripts
- Plugin registry CDN setup

**Testing:**
- CI/CD pipeline tests
- Plugin deployment tests
- Rollback tests for plugins

---

## 📊 Change Impact Analysis (Plugin-Based)

### Code Changes Required

| Component | Current State | Target State | Change Type | Effort |
|-----------|---------------|--------------|-------------|--------|
| **Plugin System** | Basic interface | Full-featured with Dependencies() | Extension | 20% |
| **Plugin Registry** | Missing | Full registry with metadata | New | 100% |
| **DI Container** | Missing | ServiceLocator + DI | New | 100% |
| **FeatureFlagPlugin** | Missing | New plugin | New | 100% |
| **StructuredLoggingPlugin** | Missing | New plugin | New | 100% |
| **HotReloadConfigPlugin** | Missing | New plugin | New | 100% |
| **DistributedTracingPlugin** | Missing | New plugin | New | 100% |
| **EnhancedMetricsPlugin** | Missing | New plugin | New | 100% |
| **LogAggregationPlugin** | Missing | New plugin | New | 100% |
| **Config System** | 7-level precedence | + plugin config support | Extension | 10% |
| **Existing Plugins** | Hard-coded dependencies | DI-based | Refactor | 30% |
| **Testing** | 100+ tests | + plugin tests | Extension | 20% |
| **CI/CD** | Missing | Plugin-based pipeline | New | 100% |

**Overall Code Change:** ~25% of codebase needs modification (lower than core implementation due to plugin approach)

### Breaking Changes

**Zero Breaking Changes with Plugin-Based Approach:**
1. Plugin interface extension (Dependencies() method) - backward compatible (default returns nil)
2. Config system extension - backward compatible
3. All plugins are opt-in via plugin config files
4. All other changes are additive

**Migration Path:**
- Existing plugins continue to work without changes
- New features are opt-in via plugin config files (~/.ti-cli/plugins/*.yaml)
- Config changes are backward compatible
- No database schema changes required
- Plugins can be disabled/rolled back independently

---

## 🎯 Risk Assessment (Plugin-Based)

### Low Risk (Phase 1 - Plugin Foundation)
- FeatureFlagPlugin (new plugin)
- StructuredLoggingPlugin (new plugin)
- HotReloadConfigPlugin (new plugin)
- Plugin interface extension (Dependencies() method)

**Mitigation:**
- New plugins don't affect existing code
- Plugins are opt-in via config files
- Interface extension is backward compatible (default returns nil)
- Can disable any plugin independently

### Medium Risk (Phase 2 - DI & Plugin Refactoring)
- ServiceLocator/DI container (new infrastructure)
- Existing plugin refactoring to use DI
- Plugin registry with metadata and versioning

**Mitigation:**
- DI is optional initially
- Refactoring is incremental per plugin
- Existing plugins continue to work without DI
- Plugin registry is additive

### Low Risk (Phase 3 - Observability Plugins)
- DistributedTracingPlugin (new plugin)
- EnhancedMetricsPlugin (new plugin)
- LogAggregationPlugin (new plugin)

**Mitigation:**
- All are new plugins, opt-in
- No changes to core system
- Can enable/disable independently
- Zero impact if disabled

### Medium Risk (Phase 4 - CI/CD & Plugin Deployment)
- Plugin-specific CI/CD pipeline
- Plugin registry deployment
- Canary deployment for plugins

**Mitigation:**
- CI/CD is non-blocking initially
- Plugin registry deployment is separate from core
- Canary deployment starts at 10% per plugin
- Can rollback individual plugins

---

## 📈 Success Metrics (Plugin-Based)

### Phase 1 Success Criteria (Plugin Foundation)
- [ ] FeatureFlagPlugin operational with config file support
- [ ] StructuredLoggingPlugin with 4 levels working
- [ ] HotReloadConfigPlugin detecting config changes
- [ ] Plugin interface extended with Dependencies() method
- [ ] Plugin config files (~/.ti-cli/plugins/*.yaml) created
- [ ] Zero breaking changes

### Phase 2 Success Criteria (DI & Plugin Refactoring)
- [ ] ServiceLocator/DI container operational
- [ ] Plugin dependency resolution working
- [ ] Plugin registry with metadata and versioning
- [ ] All existing plugins refactored to use DI
- [ ] Plugin loading/unloading independently tested
- [ ] Plugin dependencies resolution tested

### Phase 3 Success Criteria (Observability Plugins)
- [ ] DistributedTracingPlugin operational
- [ ] EnhancedMetricsPlugin with labels working
- [ ] LogAggregationPlugin working
- [ ] All observability plugins can be enabled/disabled independently
- [ ] Zero performance degradation

### Phase 4 Success Criteria (CI/CD & Plugin Deployment)
- [ ] Plugin-specific CI/CD pipeline operational
- [ ] Plugin registry deployment working
- [ ] Security scanning passing for all plugins
- [ ] Canary deployment working for plugins
- [ ] Automated rollback functional per plugin
- [ ] Plugin build and test automation

---

## 💡 Recommendations (Plugin-Based)

### Immediate Actions (This Week)
1. **Start with Phase 1** - Plugin Foundation is low-risk
2. **Create FeatureFlagPlugin first** - enables gradual rollout for all other plugins
3. **Add StructuredLoggingPlugin** - improves observability immediately
4. **Extend Plugin interface with Dependencies() method** - enables dependency resolution

### Short Term (Next Month)
1. **Complete Phase 1** - Plugin Foundation in place
2. **Start Phase 2** - DI container and plugin refactoring
3. **Create plugin config files** - ~/.ti-cli/plugins/*.yaml for opt-in control
4. **Refactor existing plugins incrementally** - one plugin at a time to use DI

### Medium Term (Next Quarter)
1. **Complete Phase 2** - DI and plugin refactoring done
2. **Start Phase 3** - Observability plugins (DistributedTracingPlugin, EnhancedMetricsPlugin, LogAggregationPlugin)
3. **Start Phase 4** - Plugin-specific CI/CD pipeline
4. **Canary deployment for new plugins** - safe rollout per plugin

### Long Term (Next 6 Months)
1. **Complete all phases** - full plugin-based deployment strategy
2. **Optimize plugin performance** - based on metrics from observability plugins
3. **Enhance plugin documentation** - auto-generated docs from plugin registry
4. **Establish plugin maintenance schedule** - regular updates per plugin
5. **Expand plugin ecosystem** - community plugins

---

## 🎯 Conclusion (Plugin-Based)

### Summary
- **Current codebase has good foundation** (basic plugin system, config, testing, observability)
- **25% of codebase needs modification** to achieve full plugin-based deployment strategy (lower than core implementation)
- **Zero breaking changes** - all plugins are opt-in via config files
- **16-week migration timeline** (4 phases × 4 weeks)
- **Low to medium risk** - gradual rollout with independent plugin deployment

### Key Advantages of Plugin-Based Approach
1. **Existing plugin system** - don't need to rebuild from scratch, just enhance
2. **Zero overhead when disabled** - plugins only load when enabled
3. **Independent testing** - each plugin can be tested separately
4. **Easy rollback** - disable individual plugins without affecting others
5. **Selective deployment** - enable different plugins per environment
6. **Loose coupling via DI** - plugins depend on abstractions, not concrete implementations
7. **Version control per plugin** - each plugin has its own version and lifecycle
8. **Hot-swappable** - can enable/disable plugins without restart

### Migration Path
1. **Phase 1 (Plugin Foundation)** - Low risk, high value (FeatureFlagPlugin, StructuredLoggingPlugin, HotReloadConfigPlugin)
2. **Phase 2 (DI & Plugin Refactoring)** - Medium risk, necessary for modularity (ServiceLocator, plugin registry, DI-based plugins)
3. **Phase 3 (Observability Plugins)** - Low risk, high value (DistributedTracingPlugin, EnhancedMetricsPlugin, LogAggregationPlugin)
4. **Phase 4 (CI/CD & Plugin Deployment)** - Medium risk, necessary for deployment (plugin-specific CI/CD, plugin registry deployment)

### Final Recommendation
**Start with Phase 1 immediately** - FeatureFlagPlugin and StructuredLoggingPlugin provide immediate value with minimal risk. The plugin-based approach ensures zero breaking changes, independent testing, and easy rollback. Proceed incrementally through other phases, using plugin config files to control which plugins are enabled in each environment.
