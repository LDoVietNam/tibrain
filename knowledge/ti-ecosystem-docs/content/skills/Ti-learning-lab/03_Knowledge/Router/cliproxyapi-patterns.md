# CLIProxyAPI Pattern Analysis

## Repository: CLIProxyAPI
**URL:** https://github.com/router-for-me/CLIProxyAPI
**Location:** Z:\10_WORKPLACE\Ti\Ti-learning-lab\07_Repositories\router\CLIProxyAPI

## Architecture Patterns

### 1. Multi-Storage Backend Pattern
**Pattern:** Pluggable storage backends for configuration and tokens
```go
// Environment-based storage selection
if value, ok := lookupEnv("PGSTORE_DSN", "pgstore_dsn"); ok {
    usePostgresStore = true
    pgStoreDSN = value
}
if value, ok := lookupEnv("GITSTORE_GIT_URL", "gitstore_git_url"); ok {
    useGitStore = true
    gitStoreRemoteURL = value
}
if value, ok := lookupEnv("OBJECTSTORE_ENDPOINT", "objectstore_endpoint"); ok {
    useObjectStore = true
    objectStoreEndpoint = value
}

// Priority: Home > Postgres > Object > Git > File
if strings.TrimSpace(homeAddr) != "" {
    // Load config from home control plane
} else if usePostgresStore {
    // Use postgres-backed store
} else if useObjectStore {
    // Use object store (S3-compatible)
} else if useGitStore {
    // Use git-backed store
} else {
    // Use local file store
}
```

**Key Learnings:**
- Environment-based storage selection
- Multiple backend support (Postgres, Git, S3, File)
- Priority-based fallback chain
- Home control plane integration
- Bootstrap from template if needed

### 2. OAuth Authentication Pattern
**Pattern:** Multiple OAuth providers with callback server
```go
// Command-line flags for different OAuth flows
var login bool
var codexLogin bool
var codexDeviceLogin bool
var claudeLogin bool
var antigravityLogin bool
var kimiLogin bool
var noBrowser bool
var oauthCallbackPort int

// Login options
options := &cmd.LoginOptions{
    NoBrowser:    noBrowser,
    CallbackPort: oauthCallbackPort,
}

// Provider-specific login handlers
if login {
    cmd.DoLogin(cfg, projectID, options)  // Google/Gemini
} else if codexLogin {
    cmd.DoCodexLogin(cfg, options)
} else if claudeLogin {
    cmd.DoClaudeLogin(cfg, options)
} else if kimiLogin {
    cmd.DoKimiLogin(cfg, options)
}
```

**Key Learnings:**
- Multiple OAuth provider support
- Device code flow alternative
- Browser automation option
- Custom callback port
- Provider-specific handlers

### 3. TUI Mode Pattern
**Pattern:** Terminal User Interface with embedded server
```go
if tuiMode {
    if standalone {
        // Start embedded local server
        localMgmtPassword := fmt.Sprintf("tui-%d-%d", os.Getpid(), time.Now().UnixNano())
        if password == "" {
            password = localMgmtPassword
        }
        
        cancel, done := cmd.StartServiceBackground(cfg, configFilePath, password)
        
        // Wait for server to be ready with exponential backoff
        backoff := 100 * time.Millisecond
        for i := 0; i < 30; i++ {
            if _, errGetConfig := client.GetConfig(); errGetConfig == nil {
                ready = true
                break
            }
            time.Sleep(backoff)
            if backoff < time.Second {
                backoff = time.Duration(float64(backoff) * 1.5)
            }
        }
        
        // Run TUI client
        tui.Run(cfg.Port, password, hook, origStdout)
    } else {
        // Pure management client (server must be running)
        tui.Run(cfg.Port, password, nil, os.Stdout)
    }
}
```

**Key Learnings:**
- Standalone mode with embedded server
- Auto-generated management password
- Exponential backoff for readiness check
- Log redirection for TUI
- Client-only mode for remote management

### 4. Home Control Plane Pattern
**Pattern:** Remote configuration from control plane
```go
if strings.TrimSpace(homeAddr) != "" {
    configLoadedFromHome = true
    
    host, portStr, errSplit := net.SplitHostPort(strings.TrimSpace(homeAddr))
    port, errPort := strconv.Atoi(strings.TrimSpace(portStr))
    
    homeCfg := config.HomeConfig{
        Enabled:  true,
        Host:     host,
        Port:     port,
        Password: trimmedHomePassword,
    }
    homeClient := home.New(homeCfg)
    
    ctxHome, cancelHome := context.WithTimeout(context.Background(), 30*time.Second)
    raw, errGetConfig := homeClient.GetConfig(ctxHome)
    cancelHome()
    
    parsed, errParseConfig := config.ParseConfigBytes(raw)
    parsed.Home = homeCfg
    parsed.Port = 8317
    cfg = parsed
}
```

**Key Learnings:**
- Remote configuration loading
- Timeout-based requests
- Host:port validation
- Redis AUTH password support
- Default port for home mode

### 5. Token Store Registration Pattern
**Pattern:** Shared token store registration
```go
// Register the shared token store once so all components use the same persistence backend
if usePostgresStore {
    sdkAuth.RegisterTokenStore(pgStoreInst)
} else if useObjectStore {
    sdkAuth.RegisterTokenStore(objectStoreInst)
} else if useGitStore {
    sdkAuth.RegisterTokenStore(gitStoreInst)
} else {
    sdkAuth.RegisterTokenStore(sdkAuth.NewFileTokenStore())
}
```

**Key Learnings:**
- Singleton token store pattern
- Consistent backend across components
- Registration-based architecture
- Default fallback to file store

### 6. Cloud Deploy Mode Pattern
**Pattern:** Standby mode for cloud deployments
```go
deployEnv := os.Getenv("DEPLOY")
if deployEnv == "cloud" {
    isCloudDeploy = true
}

if isCloudDeploy && !configFileExists {
    // No config file available, just wait for shutdown
    cmd.WaitForCloudDeploy()
    return
}
```

**Key Learnings:**
- Environment-based cloud detection
- Standby mode without config
- Shutdown signal waiting
- Config file validation

### 7. Background Service Pattern
**Pattern:** Start service in background with cancellation
```go
cancel, done := cmd.StartServiceBackground(cfg, configFilePath, password)

// ... do other work ...

cancel()
<-done
```

**Key Learnings:**
- Goroutine-based background service
- Cancellation function
- Done channel for completion
- Graceful shutdown

### 8. Auto-Updater Pattern
**Pattern:** Background auto-updaters for models and assets
```go
managementasset.StartAutoUpdater(context.Background(), configFilePath)
misc.StartAntigravityVersionUpdater(context.Background())
if !localModel && !cfg.Home.Enabled {
    registry.StartModelsUpdater(context.Background())
}
```

**Key Learnings:**
- Multiple background updaters
- Context-based lifecycle
- Conditional model updates
- Home mode disables remote updates

### 9. Usage Statistics Pattern
**Pattern:** Configurable usage tracking
```go
redisqueue.SetUsageStatisticsEnabled(cfg.UsageStatisticsEnabled)
redisqueue.SetRetentionSeconds(cfg.RedisUsageQueueRetentionSeconds)
coreauth.SetQuotaCooldownDisabled(cfg.DisableCooling)
```

**Key Learnings:**
- Global configuration setters
- Retention period configuration
- Cooldown disable option
- Usage statistics toggle

## Implementation Recommendations for Ti Router

### 1. Multi-Storage Backend
- Implement pluggable storage backends
- Add environment-based selection
- Support Postgres, Git, S3, File
- Implement priority-based fallback

### 2. OAuth Authentication
- Implement multiple OAuth providers
- Add device code flow support
- Support custom callback ports
- Add browser automation option

### 3. TUI Mode
- Implement terminal UI
- Add embedded server mode
- Implement readiness checks
- Add log redirection

### 4. Home Control Plane
- Implement remote configuration
- Add timeout-based requests
- Support password authentication
- Implement default ports

### 5. Token Store
- Implement singleton token store
- Add registration pattern
- Support multiple backends
- Default to file store

### 6. Cloud Deploy
- Implement standby mode
- Add environment detection
- Validate config files
- Wait for shutdown signals

### 7. Background Services
- Implement goroutine-based services
- Add cancellation support
- Use done channels
- Graceful shutdown

### 8. Auto-Updaters
- Implement background updaters
- Add context lifecycle
- Conditional updates
- Home mode support

### 9. Usage Statistics
- Implement configurable tracking
- Add retention periods
- Support cooldown disable
- Global configuration

## Priority
Medium - Apply these patterns for multi-storage backends and OAuth authentication.
