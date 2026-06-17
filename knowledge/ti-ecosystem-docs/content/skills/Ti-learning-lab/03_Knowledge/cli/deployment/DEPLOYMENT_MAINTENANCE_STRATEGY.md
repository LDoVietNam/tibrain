# Ti CLI Deployment & Maintenance Strategy

> **Last Updated**: 2026-05-05
> **Version**: 1.0.0
> **Purpose**: Define deployment strategy for full-feature Ti CLI with maintainable architecture

## 🎯 Objectives

1. **Full Feature Coverage**: Deploy all 50+ features from integration roadmap
2. **Maintainability**: Easy to update, debug, and extend
3. **Modularity**: Features can be enabled/disabled independently
4. **Backward Compatibility**: Existing workflows not broken
5. **Gradual Rollout**: Features can be rolled out incrementally
6. **Cost Control**: Optional features don't add overhead when disabled

---

## 🏗️ Architecture Strategy

### 1. Plugin-Based Modular Architecture

**Principle:** Every feature is a plugin that can be enabled/disabled independently

```
Ti CLI Core (Minimal)
├── Required Plugins (Always On)
│   ├── Router Integration
│   ├── Basic Commands
│   ├── Config Management
│   └── Error Handling
│
├── Optional Plugins (Feature Flags)
│   ├── Security Plugin
│   │   ├── Permission Modes
│   │   ├── Sandbox Modes
│   │   └── Tool Whitelist/Blacklist
│   ├── Session Plugin
│   │   ├── Session Management
│   │   ├── Export/Import
│   │   └── Budget Control
│   ├── Git Plugin
│   │   ├── Git Operations
│   │   ├── Worktree Support
│   │   └── PR Integration
│   ├── Debug Plugin
│   │   ├── Debug Mode
│   │   ├── Health Check
│   │   └── Logging
│   ├── Skills Plugin
│   │   ├── Skills System
│   │   ├── Background Tasks
│   │   └── Quest Orchestrator
│   ├── Memory Plugin
│   │   ├── Persistent Memory
│   │   ├── Capability Inventory
│   │   └── Context Management
│   └── Advanced Plugin
│       ├── Phase Routing
│       ├── Domain Routing
│       ├── Multi-Backend Orchestration
│       └── External CLI Adapter
```

**Benefits:**
- **Zero overhead** when plugin disabled (not loaded)
- **Independent testing** per plugin
- **Easy rollback** (disable problematic plugin)
- **Selective deployment** (enable only needed features)
- **Loose coupling** via dependency injection
- **Version control** per plugin (semver)
- **Hot-swappable** plugins (load/unload at runtime)

### 1.1 Plugin-Based Implementation Strategy

**Principle:** All new features implemented as plugins, not core components

**Why Plugin-Based?**

Ti CLI already has a solid plugin system (`internal/core/platform.go` with Plugin interface). Instead of adding features to core, we extend the plugin system:

```
Core Components (Minimal)
├── Plugin Platform (internal/core/platform.go)
├── Config System (internal/kernel/config.go)
├── Router Integration
└── Basic Commands

New Features as Plugins
├── FeatureFlagPlugin (internal/plugins/featureflag/)
├── StructuredLoggingPlugin (internal/plugins/logging/)
├── HotReloadConfigPlugin (internal/plugins/configwatcher/)
├── SecurityPlugin (internal/plugins/security/)
├── SessionPlugin (internal/plugins/session/)
├── GitPlugin (internal/plugins/git/)
└── ...
```

**Advantages over Core Implementation:**

| Aspect | Core Implementation | Plugin-Based |
|--------|-------------------|---------------|
| **Startup Time** | Slower (all features loaded) | Faster (only enabled plugins) |
| **Memory Usage** | Higher (all features in memory) | Lower (only enabled plugins) |
| **Testing** | Hard (core integration tests) | Easy (isolated plugin tests) |
| **Rollback** | Difficult (need redeploy) | Easy (disable plugin) |
| **Extensibility** | Limited (core changes) | Unlimited (add plugins) |
| **Versioning** | Monolithic (single version) | Per-plugin (semver) |
| **Deployment** | All-or-nothing | Gradual (plugin-by-plugin) |

**Plugin Development Workflow:**

```bash
# 1. Create new plugin
mkdir -p internal/plugins/myfeature
cd internal/plugins/myfeature

# 2. Implement Plugin interface
cat > plugin.go << 'EOF'
package myfeature

import (
    "context"
    "github.com/ti/cli/internal/core"
)

type MyFeaturePlugin struct {
    config Config
    state core.PluginState
}

func (p *MyFeaturePlugin) Name() string {
    return "myfeature"
}

func (p *MyFeaturePlugin) Version() string {
    return "1.0.0"
}

func (p *MyFeaturePlugin) Metadata() core.PluginMetadata {
    return core.PluginMetadata{
        Description: "My feature description",
        Author:      "Ti Team",
        Dependencies: []string{"config"}, // Depends on config plugin
    }
}

func (p *MyFeaturePlugin) Dependencies() []string {
    return []string{"config"} // Required dependencies
}

func (p *MyFeaturePlugin) Initialize(ctx context.Context, config map[string]string) error {
    // Load config
    // Initialize resources
    p.state = core.PluginStateReady
    return nil
}

func (p *MyFeaturePlugin) Execute(ctx context.Context, task string, input map[string]interface{}) (map[string]interface{}, error) {
    // Execute task
    return result, nil
}

func (p *MyFeaturePlugin) Shutdown(ctx context.Context) error {
    // Cleanup resources
    p.state = core.PluginStateShutdown
    return nil
}

func (p *MyFeaturePlugin) HealthCheck(ctx context.Context) error {
    // Check plugin health
    return nil
}

func (p *MyFeaturePlugin) State() core.PluginState {
    return p.state
}
EOF

# 3. Create plugin config
cat > ~/.ti-cli/plugins/myfeature.yaml << 'EOF'
name: myfeature
version: "1.0.0"
enabled: true
dependencies:
  - name: config
    min_version: "1.0.0"
config:
  option1: "value1"
  option2: "value2"
EOF

# 4. Register plugin in CLI
# Add to internal/plugins/registry.go or auto-discovery
```

**Plugin Config Structure:**

```yaml
# ~/.ti-cli/plugins/featureflag.yaml
name: featureflag
version: "1.0.0"
enabled: true
priority: 10  # Load order (higher = earlier)
dependencies:
  - name: config
    min_version: "1.0.0"
config:
  feature_flags:
    permission_modes:
      enabled: true
      rollout_percentage: 100
    sandbox_modes:
      enabled: true
      rollout_percentage: 50
  storage:
    backend: "sqlite"
    path: "~/.ti-cli/feature-flags.db"
```

**Plugin Loading Order:**

1. **Required Plugins** (always on, priority 100)
   - config
   - router
   - error-handling

2. **Optional Plugins** (config-driven, priority 0-99)
   - featureflag (priority 10)
   - logging (priority 20)
   - configwatcher (priority 30)
   - security (priority 40)
   - session (priority 50)
   - git (priority 60)

3. **Dependency Resolution**
   - Load plugins in priority order
   - Check dependencies before loading
   - Fail if dependency not met

**Plugin Lifecycle:**

```
Load → Initialize → Ready → Execute → Shutdown → Unload
  ↓         ↓          ↓       ↓         ↓         ↓
Check   Load Config  Health   Task    Cleanup  Remove
Deps                                 Resources
```

**Plugin Hot-Reload:**

```bash
# Enable plugin without restart
ti plugin enable featureflag

# Disable plugin without restart
ti plugin disable featureflag

# Reload plugin config
ti plugin reload featureflag

# Check plugin status
ti plugin status featureflag
```

### 2. Configuration-Driven Feature Flags

**Principle:** All features controlled via configuration, not code changes

```yaml
# ~/.ti-cli/config.yaml
plugins:
  # Security Plugin
  security:
    enabled: true
    permission_modes:
      - acceptEdits
      - auto
      - bypassPermissions
      - default
      - dontAsk
    sandbox_modes:
      - read-only
      - workspace-write
      - danger-full-access
    tool_whitelist:
      - file_read
      - git_operations
    tool_blacklist:
      - file_write
      - network_operations

  # Session Plugin
  session:
    enabled: true
    persistence: true
    export_formats:
      - json
      - yaml
    budget_control:
      enabled: true
      max_budget_usd: 100.0
      per_session_budget: 10.0

  # Git Plugin
  git:
    enabled: true
    worktree_support: true
    pr_integration: true
    auto_commit: false

  # Debug Plugin
  debug:
    enabled: true
    debug_mode: false
    debug_categories:
      - router
      - provider
      - agent
    debug_file: ""

  # Skills Plugin
  skills:
    enabled: false  # Disabled by default
    bundled_skills:
      - loop
      - qc-helper
      - review
    custom_skills_dir: ~/.ti-cli/skills/

  # Memory Plugin
  memory:
    enabled: true
    persistent: true
    memory_backend: sqlite
    memory_path: ~/.ti-cli/memory.db

  # Advanced Plugin
  advanced:
    enabled: false  # Disabled by default
    phase_routing: false
    domain_routing: false
    multi_backend: false
    external_cli_adapter: false
```

**Benefits:**
- **No code changes** to enable/disable features
- **Per-environment** configuration (dev/staging/prod)
- **User control** over feature set
- **A/B testing** via configuration

### 3. Dependency Injection & Service Locator

**Principle:** Plugins depend on abstractions, not concrete implementations

```go
// Plugin Interface
type Plugin interface {
    Name() string
    Version() string
    Dependencies() []string
    Initialize(ctx PluginContext) error
    Execute(ctx Context, args map[string]any) error
    Shutdown() error
}

// Plugin Context (DI Container)
type PluginContext struct {
    Router RouterService
    Config Config
    Logger Logger
    Store Storage
    // ... other services
}

// Service Locator
type ServiceLocator struct {
    services map[string]interface{}
    mutex sync.RWMutex
}

func (sl *ServiceLocator) Get(name string) (interface{}, error) {
    sl.mutex.RLock()
    defer sl.mutex.RUnlock()
    
    service, exists := sl.services[name]
    if !exists {
        return nil, fmt.Errorf("service %s not found", name)
    }
    return service, nil
}

func (sl *ServiceLocator) Register(name string, service interface{}) {
    sl.mutex.Lock()
    defer sl.mutex.Unlock()
    sl.services[name] = service
}
```

**Benefits:**
- **Loose coupling** between plugins
- **Easy testing** (mock dependencies)
- **Flexible composition**
- **Runtime service discovery**

### 4. Layered Configuration (7-Level Precedence)

**Principle:** Configuration cascades from global to local with override capability

```
1. Remote Config (cloud) - Enterprise defaults
2. Global Config (~/.ti-cli/config.yaml) - User defaults
3. Custom Config (~/.ti-cli/custom.yaml) - User overrides
4. Project Config (.ti/config.yaml) - Project-specific
5. .opencode Config (.opencode/config.json) - OpenCode integration
6. Inline Config (command-line flags) - Runtime overrides
7. Managed Config (environment variables) - Deployment overrides
```

**Implementation:**
```go
type ConfigLoader struct {
    loaders []ConfigLoader
}

type ConfigLoader interface {
    Load() (map[string]interface{}, error)
    Priority() int
}

func (cl *ConfigLoader) Load() (map[string]interface{}, error) {
    config := make(map[string]interface{})
    
    // Load in priority order (lowest to highest)
    for _, loader := range cl.loaders {
        layerConfig, err := loader.Load()
        if err != nil {
            continue // Skip failed loaders
        }
        merge(config, layerConfig) // Higher priority overrides
    }
    
    return config, nil
}
```

**Benefits:**
- **Flexible deployment** (dev/staging/prod configs)
- **Enterprise control** (remote config)
- **User customization** (local overrides)
- **Project isolation** (.ti/config.yaml)

---

## 🚀 Deployment Strategy

### 1. Feature Flag Gates

**Principle:** Every new feature behind a feature flag

```yaml
# ~/.ti-cli/feature-flags.yaml
feature_flags:
  # Phase 1 Features
  permission_modes:
    enabled: true
    rollout_percentage: 100
    allowed_users: []
    allowed_teams: []
  
  sandbox_modes:
    enabled: true
    rollout_percentage: 50  # Gradual rollout
    allowed_users: ["admin@example.com"]
    allowed_teams: ["security-team"]
  
  # Phase 2 Features
  session_management:
    enabled: false
    rollout_percentage: 0
    allowed_users: []
    allowed_teams: []
  
  budget_control:
    enabled: false
    rollout_percentage: 0
    allowed_users: []
    allowed_teams: []
```

**Implementation:**
```go
type FeatureFlagService struct {
    flags map[string]FeatureFlag
    mutex sync.RWMutex
}

type FeatureFlag struct {
    Enabled bool
    RolloutPercentage int
    AllowedUsers []string
    AllowedTeams []string
}

func (ffs *FeatureFlagService) IsEnabled(feature string, user User) bool {
    ffs.mutex.RLock()
    defer ffs.mutex.RUnlock()
    
    flag, exists := ffs.flags[feature]
    if !exists {
        return false
    }
    
    // Check if enabled
    if !flag.Enabled {
        return false
    }
    
    // Check user whitelist
    if len(flag.AllowedUsers) > 0 {
        for _, allowedUser := range flag.AllowedUsers {
            if user.Email == allowedUser {
                return true
            }
        }
        return false
    }
    
    // Check team whitelist
    if len(flag.AllowedTeams) > 0 {
        for _, team := range user.Teams {
            for _, allowedTeam := range flag.AllowedTeams {
                if team == allowedTeam {
                    return true
                }
            }
        }
        return false
    }
    
    // Check rollout percentage
    if flag.RolloutPercentage < 100 {
        hash := sha256.Sum256([]byte(user.ID + feature))
        percentage := int(hash[0]) % 100
        return percentage < flag.RolloutPercentage
    }
    
    return true
}
```

**Benefits:**
- **Gradual rollout** (0% → 50% → 100%)
- **User/team whitelisting** (beta testing)
- **Instant rollback** (disable flag)
- **A/B testing** (different configs per user)

### 2. Versioned Plugin System

**Principle:** Plugins have semantic versioning with backward compatibility

```yaml
# ~/.ti-cli/plugins/registry.yaml
plugins:
  security:
    version: "1.0.0"
    min_core_version: "3.0.0"
    dependencies:
      - name: config
        min_version: "1.0.0"
    checksum: "sha256:..."
  
  session:
    version: "1.2.0"
    min_core_version: "3.0.0"
    dependencies:
      - name: config
        min_version: "1.0.0"
      - name: storage
        min_version: "1.0.0"
    checksum: "sha256:..."
  
  git:
    version: "0.9.0"
    min_core_version: "3.0.0"
    dependencies:
      - name: config
        min_version: "1.0.0"
    checksum: "sha256:..."
```

**Implementation:**
```go
type PluginRegistry struct {
    plugins map[string]PluginMetadata
    mutex sync.RWMutex
}

type PluginMetadata struct {
    Name string
    Version string
    MinCoreVersion string
    Dependencies []Dependency
    Checksum string
    Path string
}

type Dependency struct {
    Name string
    MinVersion string
}

func (pr *PluginRegistry) LoadPlugin(name string) error {
    pr.mutex.Lock()
    defer pr.mutex.Unlock()
    
    metadata, exists := pr.plugins[name]
    if !exists {
        return fmt.Errorf("plugin %s not found", name)
    }
    
    // Check core version compatibility
    if !versionCompatible(coreVersion, metadata.MinCoreVersion) {
        return fmt.Errorf("plugin %s requires core version %s, current is %s",
            name, metadata.MinCoreVersion, coreVersion)
    }
    
    // Check dependencies
    for _, dep := range metadata.Dependencies {
        depPlugin, exists := pr.plugins[dep.Name]
        if !exists {
            return fmt.Errorf("plugin %s depends on %s which is not loaded",
                name, dep.Name)
        }
        if !versionCompatible(depPlugin.Version, dep.MinVersion) {
            return fmt.Errorf("plugin %s depends on %s version %s, current is %s",
                name, dep.Name, dep.MinVersion, dep.Plugin.Version)
        }
    }
    
    // Load plugin
    plugin, err := loadPlugin(metadata.Path)
    if err != nil {
        return err
    }
    
    // Verify checksum
    if !verifyChecksum(metadata.Path, metadata.Checksum) {
        return fmt.Errorf("plugin %s checksum verification failed", name)
    }
    
    return nil
}
```

**Benefits:**
- **Version compatibility** (semver)
- **Dependency resolution**
- **Checksum verification** (security)
- **Independent updates**

### 3. Hot-Reloadable Configuration

**Principle:** Configuration changes take effect without restart

```go
type ConfigWatcher struct {
    configPath string
    config *Config
    mutex sync.RWMutex
    watcher *fsnotify.Watcher
    callbacks []ConfigChangeCallback
}

type ConfigChangeCallback func(old Config, new Config)

func (cw *ConfigWatcher) Watch() error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    cw.watcher = watcher
    
    err = watcher.Add(cw.configPath)
    if err != nil {
        return err
    }
    
    go cw.watchLoop()
    return nil
}

func (cw *ConfigWatcher) watchLoop() {
    for {
        select {
        case event, ok := <-cw.watcher.Events:
            if !ok {
                return
            }
            if event.Op&fsnotify.Write == fsnotify.Write {
                cw.reloadConfig()
            }
        case err, ok := <-cw.watcher.Errors:
            if !ok {
                return
            }
            log.Printf("Config watcher error: %v", err)
        }
    }
}

func (cw *ConfigWatcher) reloadConfig() {
    cw.mutex.Lock()
    defer cw.mutex.Unlock()
    
    oldConfig := cw.config
    newConfig, err := loadConfig(cw.configPath)
    if err != nil {
        log.Printf("Failed to reload config: %v", err)
        return
    }
    
    // Notify callbacks
    for _, callback := range cw.callbacks {
        callback(oldConfig, newConfig)
    }
    
    cw.config = newConfig
}
```

**Benefits:**
- **Zero downtime** config changes
- **Dynamic feature toggling**
- **Real-time debugging**
- **User customization**

### 4. Health Check & Diagnostics

**Principle:** Comprehensive health checks for all components

```go
type HealthChecker struct {
    checks []HealthCheck
}

type HealthCheck interface {
    Name() string
    Check() HealthStatus
}

type HealthStatus struct {
    Status string // "healthy", "degraded", "unhealthy"
    Message string
    Details map[string]interface{}
}

type RouterHealthCheck struct {
    router RouterService
}

func (rhc *RouterHealthCheck) Name() string {
    return "router"
}

func (rhc *RouterHealthCheck) Check() HealthStatus {
    err := rhc.router.Health()
    if err != nil {
        return HealthStatus{
            Status: "unhealthy",
            Message: err.Error(),
            Details: map[string]interface{}{
                "endpoint": rhc.router.Endpoint(),
            },
        }
    }
    
    return HealthStatus{
        Status: "healthy",
        Message: "Router is healthy",
        Details: map[string]interface{}{
            "endpoint": rhc.router.Endpoint(),
            "providers": rhc.router.ProviderCount(),
        },
    }
}

type PluginHealthCheck struct {
    registry *PluginRegistry
}

func (phc *PluginHealthCheck) Name() string {
    return "plugins"
}

func (phc *PluginHealthCheck) Check() HealthStatus {
    plugins := phc.registry.List()
    unhealthy := []string{}
    
    for _, plugin := range plugins {
        if !plugin.IsHealthy() {
            unhealthy = append(unhealthy, plugin.Name())
        }
    }
    
    if len(unhealthy) > 0 {
        return HealthStatus{
            Status: "degraded",
            Message: fmt.Sprintf("Unhealthy plugins: %v", unhealthy),
            Details: map[string]interface{}{
                "total": len(plugins),
                "unhealthy": len(unhealthy),
                "unhealthy_plugins": unhealthy,
            },
        }
    }
    
    return HealthStatus{
        Status: "healthy",
        Message: "All plugins healthy",
        Details: map[string]interface{}{
            "total": len(plugins),
        },
    }
}
```

**Benefits:**
- **Component visibility**
- **Proactive monitoring**
- **Quick diagnosis**
- **Automated rollback triggers**

---

## 🔄 Deployment Pipeline

### 1. CI/CD Pipeline (Plugin-Based)

```yaml
# .github/workflows/deploy.yml
name: Deploy Ti CLI

on:
  push:
    tags:
      - 'v*'
    branches:
      - main
  workflow_dispatch:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Run core tests
        run: |
          cd apps/cli
          go test ./internal/core/... -coverprofile=core-coverage.txt
          go test ./internal/kernel/... -coverprofile=kernel-coverage.txt
      - name: Run plugin tests
        run: |
          cd apps/cli
          go test ./internal/plugins/... -coverprofile=plugins-coverage.txt
      - name: Run E2E tests
        run: |
          cd apps/cli/e2e
          go test ./...
      - name: Run plugin integration tests
        run: |
          cd apps/cli
          go test ./internal/integration/... -coverprofile=integration-coverage.txt

  build:
    needs: test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        os: [linux, windows, darwin]
        arch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Build CLI
        run: |
          cd apps/cli
          GOOS=${{ matrix.os }} GOARCH=${{ matrix.arch }} go build -o ti-${{ matrix.os }}-${{ matrix.arch }} .
      - name: Build plugins
        run: |
          cd apps/cli
          for plugin in internal/plugins/*/; do
            if [ -f "$plugin/plugin.go" ]; then
              go build -o plugins/$(basename $plugin).so -buildmode=plugin $plugin
            fi
          done
      - name: Upload CLI artifact
        uses: actions/upload-artifact@v3
        with:
          name: ti-${{ matrix.os }}-${{ matrix.arch }}
          path: apps/cli/ti-${{ matrix.os }}-${{ matrix.arch }}
      - name: Upload plugins artifact
        uses: actions/upload-artifact@v3
        with:
          name: plugins-${{ matrix.os }}-${{ matrix.arch }}
          path: apps/cli/plugins/

  plugin-tests:
    needs: test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        plugin: [featureflag, logging, configwatcher, security, session, git]
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Test plugin loading
        run: |
          cd apps/cli
          go test ./internal/plugins/${{ matrix.plugin }}/... -v
      - name: Test plugin dependencies
        run: |
          cd apps/cli
          go test ./internal/plugins/... -run TestDependencies -v
      - name: Test plugin hot-reload
        run: |
          cd apps/cli
          go test ./internal/plugins/${{ matrix.plugin }}/... -run TestHotReload -v

  security-scan:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run security scan on core
        uses: securego/gosec@master
        with:
          args: './apps/cli/internal/core/...'
      - name: Run security scan on plugins
        uses: securego/gosec@master
        with:
          args: './apps/cli/internal/plugins/...'
      - name: Run dependency scan
        run: |
          cd apps/cli
          go list -json -m all | nancy sleuth

  deploy:
    needs: [build, security-scan, plugin-tests]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Download CLI artifacts
        uses: actions/download-artifact@v3
      - name: Download plugin artifacts
        uses: actions/download-artifact@v3
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            ti-linux-amd64
            ti-linux-arm64
            ti-windows-amd64
            ti-windows-arm64
            ti-darwin-amd64
            ti-darwin-arm64
            plugins-*/
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  deploy-registry:
    needs: deploy
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy plugin registry
        run: |
          # Deploy plugin registry to S3/CDN
          aws s3 sync ~/.ti-cli/plugins/ s3://ti-cli-plugins/
          # Invalidate CDN cache
          aws cloudfront create-invalidation --distribution-id ${{ secrets.CLOUDFRONT_ID }} --paths "/*"
      - name: Update plugin index
        run: |
          # Generate plugin index
          ti plugin generate-index > plugin-index.json
          aws s3 cp plugin-index.json s3://ti-cli-plugins/plugin-index.json
```

### 2. Canary Deployment

```yaml
# Canary deployment strategy
canary:
  stage: 1 (10%):  # Week 1
    users: ["beta-testers"]
    features: ["permission_modes", "debug_mode"]
    monitoring: ["error_rate", "latency", "user_feedback"]
    
  stage 2 (25%):  # Week 2
    users: ["beta-testers", "early-adopters"]
    features: ["permission_modes", "debug_mode", "sandbox_modes"]
    monitoring: ["error_rate", "latency", "user_feedback", "cost"]
    
  stage 3 (50%):  # Week 3
    users: ["beta-testers", "early-adopters", "general-users"]
    features: ["permission_modes", "debug_mode", "sandbox_modes", "session_management"]
    monitoring: ["error_rate", "latency", "user_feedback", "cost", "usage"]
    
  stage 4 (100%): # Week 4
    users: ["all"]
    features: ["all"]
    monitoring: ["all_metrics"]
```

### 3. Rollback Strategy

```yaml
rollback_triggers:
  error_rate: > 5% for 5 minutes
  latency_p99: > 2x baseline for 10 minutes
  user_feedback: negative sentiment > 30%
  cost: > 2x budget for 1 hour

rollback_actions:
  immediate:
    - Disable feature flag
    - Revert to previous version
    - Notify team
    
  automated:
    - Health check failure
    - Plugin crash
    - Config validation error
    
  manual:
    - User reported critical bug
    - Security vulnerability
    - Performance degradation
```

---

## 📊 Monitoring & Observability

### 1. Metrics Collection

```go
type MetricsCollector struct {
    metrics map[string]Metric
    mutex sync.RWMutex
}

type Metric struct {
    Name string
    Type string // "counter", "gauge", "histogram"
    Value float64
    Labels map[string]string
    Timestamp time.Time
}

func (mc *MetricsCollector) Increment(name string, labels map[string]string) {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()
    
    key := metricKey(name, labels)
    metric, exists := mc.metrics[key]
    if !exists {
        metric = Metric{
            Name: name,
            Type: "counter",
            Value: 0,
            Labels: labels,
            Timestamp: time.Now(),
        }
    }
    
    metric.Value += 1
    mc.metrics[key] = metric
}

func (mc *MetricsCollector) Observe(name string, value float64, labels map[string]string) {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()
    
    key := metricKey(name, labels)
    metric, exists := mc.metrics[key]
    if !exists {
        metric = Metric{
            Name: name,
            Type: "histogram",
            Value: value,
            Labels: labels,
            Timestamp: time.Now(),
        }
    }
    
    mc.metrics[key] = metric
}
```

**Key Metrics:**
- **Plugin health**: Up/down status, error rate
- **Feature usage**: Calls per feature, success rate
- **Performance**: Latency p50/p95/p99, throughput
- **Cost**: Cost per feature, total cost
- **User feedback**: Satisfaction score, bug reports

### 2. Logging Strategy

```go
type Logger struct {
    level LogLevel
    outputs []LogOutput
    structured bool
}

type LogLevel int

const (
    DEBUG LogLevel = iota
    INFO
    WARN
    ERROR
)

type LogOutput interface {
    Write(entry LogEntry) error
}

type LogEntry struct {
    Timestamp time.Time
    Level LogLevel
    Message string
    Fields map[string]interface{}
    Plugin string
    User string
    SessionID string
}

type StructuredLogger struct {
    logger Logger
}

func (sl *StructuredLogger) WithFields(fields map[string]interface{}) *StructuredLogger {
    // Return a new logger with additional fields
    return &StructuredLogger{
        logger: Logger{
            Fields: fields,
        },
    }
}

func (sl *StructuredLogger) Info(message string) {
    sl.logger.Log(INFO, message)
}

func (sl *StructuredLogger) Error(message string, err error) {
    sl.logger.Log(ERROR, message, map[string]interface{}{
        "error": err.Error(),
        "stack": debug.Stack(),
    })
}
```

**Log Levels:**
- **DEBUG**: Detailed diagnostic information
- **INFO**: General informational messages
- **WARN**: Warning messages for potential issues
- **ERROR**: Error messages with stack traces

**Log Outputs:**
- **Console**: Colored output for local development
- **File**: Rotating log files for persistence
- **Remote**: Centralized logging (ELK, Loki)
- **Audit**: Security audit trail

### 3. Distributed Tracing

```go
type Tracer struct {
    serviceName string
    exporter TraceExporter
}

type TraceExporter interface {
    Export(span Span) error
}

type Span struct {
    TraceID string
    SpanID string
    ParentSpanID string
    Operation string
    StartTime time.Time
    EndTime time.Time
    Tags map[string]string
    Logs []LogEntry
}

func (t *Tracer) StartSpan(operation string) Span {
    span := Span{
        TraceID: generateTraceID(),
        SpanID: generateSpanID(),
        Operation: operation,
        StartTime: time.Now(),
        Tags: map[string]string{
            "service": t.serviceName,
        },
    }
    return span
}

func (t *Tracer) FinishSpan(span Span) {
    span.EndTime = time.Now()
    t.exporter.Export(span)
}
```

**Trace Spans:**
- **Plugin initialization**: Load time, dependency resolution
- **Feature execution**: Duration, success/failure
- **Router calls**: Provider selection, latency
- **External CLI calls**: Command execution, output

---

## 🧪 Testing Strategy

### 1. Unit Testing

```go
// Example: Permission mode plugin test
func TestPermissionModePlugin_Initialize(t *testing.T) {
    tests := []struct {
        name string
        config PluginConfig
        wantErr bool
    }{
        {
            name: "valid config",
            config: PluginConfig{
                PermissionModes: []string{"acceptEdits", "auto"},
            },
            wantErr: false,
        },
        {
            name: "invalid mode",
            config: PluginConfig{
                PermissionModes: []string{"invalid_mode"},
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            plugin := &PermissionModePlugin{}
            ctx := PluginContext{}
            err := plugin.Initialize(ctx, tt.config)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 2. Integration Testing

```go
func TestSessionPlugin_Integration(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()
    
    plugin := &SessionPlugin{
        Store: db,
    }
    ctx := PluginContext{
        Store: db,
    }
    
    // Test
    err := plugin.Initialize(ctx, PluginConfig{})
    if err != nil {
        t.Fatalf("Initialize() error = %v", err)
    }
    
    // Create session
    session, err := plugin.CreateSession("test-session", map[string]interface{}{
        "task": "test task",
    })
    if err != nil {
        t.Fatalf("CreateSession() error = %v", err)
    }
    
    // Resume session
    resumed, err := plugin.ResumeSession(session.ID)
    if err != nil {
        t.Fatalf("ResumeSession() error = %v", err)
    }
    
    if resumed.ID != session.ID {
        t.Errorf("ResumeSession() ID = %v, want %v", resumed.ID, session.ID)
    }
}
```

### 3. E2E Testing

```go
func TestE2E_FullWorkflow(t *testing.T) {
    // Start Ti CLI with all plugins enabled
    cli := startTestCLI(t, TestCLIConfig{
        Plugins: []string{"security", "session", "git", "debug"},
    })
    defer cli.Stop()
    
    // Test full workflow
    output := cli.Run("ti", "fix", "--file", "test.go")
    if !strings.Contains(output, "Fixed") {
        t.Errorf("Expected 'Fixed' in output, got: %s", output)
    }
    
    // Verify session was created
    sessions := cli.ListSessions()
    if len(sessions) == 0 {
        t.Error("Expected at least one session")
    }
    
    // Verify budget was tracked
    stats := cli.GetStats()
    if stats.TotalCost == 0 {
        t.Error("Expected non-zero cost")
    }
}
```

### 4. Feature Flag Testing

```go
func TestFeatureFlag_Rollout(t *testing.T) {
    service := &FeatureFlagService{
        flags: map[string]FeatureFlag{
            "test_feature": {
                Enabled: true,
                RolloutPercentage: 50,
            },
        },
    }
    
    // Test deterministic rollout based on user ID
    enabledCount := 0
    for i := 0; i < 1000; i++ {
        user := User{
            ID: fmt.Sprintf("user-%d", i),
        }
        if service.IsEnabled("test_feature", user) {
            enabledCount++
        }
    }
    
    // Should be approximately 50%
    ratio := float64(enabledCount) / 1000.0
    if ratio < 0.45 || ratio > 0.55 {
        t.Errorf("Rollout ratio = %v, want ~0.5", ratio)
    }
}
```

---

## 📚 Documentation Strategy

### 1. Auto-Generated Documentation

```go
// Generate plugin documentation from code
func GeneratePluginDocs(plugin Plugin) string {
    doc := fmt.Sprintf("# %s Plugin\n\n", plugin.Name())
    doc += fmt.Sprintf("Version: %s\n\n", plugin.Version())
    doc += fmt.Sprintf("Description: %s\n\n", plugin.Description())
    
    doc += "## Configuration\n\n"
    doc += "```yaml\n"
    doc += plugin.ConfigSchema()
    doc += "```\n\n"
    
    doc += "## Commands\n\n"
    for _, cmd := range plugin.Commands() {
        doc += fmt.Sprintf("### %s\n\n", cmd.Name)
        doc += fmt.Sprintf("%s\n\n", cmd.Description)
        doc += "```bash\n"
        doc += fmt.Sprintf("ti %s\n", cmd.Usage)
        doc += "```\n\n"
    }
    
    return doc
}
```

### 2. Interactive Documentation

```bash
# Generate interactive CLI documentation
ti docs generate --format markdown --output docs/

# Generate API documentation
ti docs api --format openapi --output api/openapi.yaml

# Generate plugin documentation
ti docs plugins --format html --output docs/plugins/
```

### 3. Example Documentation

```markdown
# Example: Using Permission Modes

## Enable Permission Mode

```bash
ti fix --permission-mode auto
```

## Configure Permission Mode

```yaml
# ~/.ti-cli/config.yaml
plugins:
  security:
    enabled: true
    permission_modes:
      - acceptEdits
      - auto
    tool_whitelist:
      - file_read
      - git_operations
```

## Use Cases

### Auto Mode
Automatically accept all edits without confirmation:

```bash
ti fix --permission-mode auto
```

### Default Mode
Ask for confirmation before each edit:

```bash
ti fix --permission-mode default
```

### Bypass Mode
Skip all permission checks (dangerous):

```bash
ti fix --permission-mode bypassPermissions
```
```

---

## 🎯 Maintenance Strategy

### 1. Regular Maintenance Tasks

**Daily:**
- Monitor error rates and latency
- Review user feedback
- Check feature flag status

**Weekly:**
- Review plugin health
- Update dependencies
- Review cost metrics
- Generate usage reports

**Monthly:**
- Security audit
- Performance review
- Plugin compatibility check
- Documentation update

**Quarterly:**
- Architecture review
- Deprecation planning
- Feature roadmap update
- User survey

### 2. Plugin Lifecycle Management

```yaml
# Plugin lifecycle stages
stages:
  experimental:
    description: "Early development, not for production"
    duration: "1-2 months"
    criteria:
      - Basic functionality
      - Unit tests
      - Documentation
  
  beta:
    description: "Limited rollout for testing"
    duration: "2-3 months"
    criteria:
      - Integration tests
      - E2E tests
      - Performance benchmarks
      - Security review
  
  stable:
    description: "Production-ready"
    duration: "6-12 months"
    criteria:
      - 99.9% uptime
      - < 0.1% error rate
      - User satisfaction > 4.5/5
      - Cost within budget
  
  deprecated:
    description: "Scheduled for removal"
    duration: "3-6 months"
    criteria:
      - Migration path available
      - User notification sent
      - Documentation updated
  
  retired:
    description: "Removed from distribution"
    action: "Delete from registry"
```

### 3. Dependency Management

```go
type DependencyManager struct {
    dependencies map[string]Dependency
    mutex sync.RWMutex
}

type Dependency struct {
    Name string
    Version string
    Type string // "go", "plugin", "external"
    Required bool
    UpdatePolicy string // "auto", "manual", "security-only"
}

func (dm *DependencyManager) CheckUpdates() []Update {
    dm.mutex.RLock()
    defer dm.mutex.RUnlock()
    
    updates := []Update{}
    for name, dep := range dm.dependencies {
        if dep.UpdatePolicy == "manual" {
            continue
        }
        
        latest, err := getLatestVersion(name)
        if err != nil {
            continue
        }
        
        if version.Compare(latest, dep.Version, ">") {
            updates = append(updates, Update{
                Name: name,
                Current: dep.Version,
                Latest: latest,
                Type: dep.UpdatePolicy,
            })
        }
    }
    
    return updates
}
```

### 4. Automated Dependency Updates

```yaml
# .github/workflows/dependencies.yml
name: Update Dependencies

on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly
  workflow_dispatch:

jobs:
  check-updates:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Check Go module updates
        run: |
          go list -u -m all
          go get -u ./...
          go mod tidy
      - name: Check plugin updates
        run: |
          ti plugin check-updates
      - name: Create PR
        uses: peter-evans/create-pull-request@v4
        with:
          title: "Weekly dependency updates"
          body: "Automated dependency updates"
          branch: "deps/update"
```

---

## 📈 Success Metrics

### Deployment Metrics
- **Deployment success rate**: > 99%
- **Rollback rate**: < 1%
- **Deployment time**: < 5 minutes
- **Downtime**: < 1 minute per deployment

### Feature Metrics
- **Feature adoption rate**: > 80% for P0 features
- **Feature success rate**: > 95%
- **Feature latency**: < 2s for 95th percentile
- **Feature cost**: Within budget

### Plugin Metrics
- **Plugin health**: > 99% uptime
- **Plugin load time**: < 100ms
- **Plugin memory**: < 50MB per plugin
- **Plugin error rate**: < 0.1%

### User Metrics
- **User satisfaction**: > 4.5/5
- **Bug report rate**: < 1% of users
- **Support ticket rate**: < 0.5% of users
- **Churn rate**: < 5% monthly

---

## 🎯 Summary

**Architecture Principles:**
1. **Plugin-based modularity** - Every feature is a plugin
2. **Configuration-driven** - Features controlled via config
3. **Dependency injection** - Loose coupling between components
4. **Feature flags** - Gradual rollout capability
5. **Hot-reload** - Zero downtime config changes
6. **Health checks** - Comprehensive monitoring
7. **Automated testing** - Unit, integration, E2E
8. **Versioned plugins** - Semver with compatibility
9. **Observability** - Metrics, logging, tracing
10. **Documentation** - Auto-generated and interactive

**Deployment Strategy:**
- **CI/CD pipeline** with automated testing
- **Canary deployment** with gradual rollout
- **Rollback strategy** with automated triggers
- **Monitoring** with comprehensive metrics
- **Maintenance** with regular tasks and lifecycle management

**Expected Outcome:**
- **Full feature coverage** (50+ features)
- **Easy maintenance** (modular, testable, documented)
- **High reliability** (99.9% uptime, < 1% rollback)
- **User satisfaction** (> 4.5/5)
- **Cost control** (within budget, predictable)
