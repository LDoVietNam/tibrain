---
tags: ["tibrain", "documentation", "provider", "router", "skill"]
scopes: ["tibrain"]
last_updated: 2026-05-22
---
# Router Agent Auto-Update Mechanism

**Last Updated**: 2026-05-08
**Purpose**: Document how Router Agent automatically updates provider implementations

## 🎯 Overview

Router Agent monitors reference router implementations (9Router, LLMCookieBridge, Kilo CLI) and automatically suggests updates to Ti Router's provider code.

## 🔄 Update Flow

```
┌─────────────────────────────────────────────────────────────┐
│  GitHub Action (Daily Check)                                 │
│  - Monitors 9Router commits                                  │
│  - Monitors LLMCookieBridge commits                           │
│  - Monitors Kilo CLI commits                                 │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ Detects update
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Router Agent (Update Check)                                │
│  - Receives webhook/notification                            │
│  - Fetches diff from reference routers                     │
│  - Analyzes changes                                         │
│  - Generates patch suggestions                              │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ Generates suggestions
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Update Review                                              │
│  - Display suggested changes                                │
│  - Request approval from user                              │
│  - Apply changes with approval                              │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ Approved
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Apply Updates                                              │
│  - Apply patch to Ti Router code                           │
│  - Run tests                                               │
│  - Restart router if needed                                │
└─────────────────────────────────────────────────────────────┘
```

## 📋 Implementation

### 1. GitHub Action Monitoring

**File**: `.github/workflows/monitor-router-updates.yml`

**Triggers**:
- Daily schedule (cron: '0 0 * * *')
- Manual workflow dispatch

**Jobs**:
- `monitor-9router` - Check 9Router commits
- `monitor-llmcookiebridge` - Check LLMCookieBridge commits
- `monitor-kilo` - Check Kilo CLI commits
- `notify-router-agent` - Send notification to Router Agent

**Output**:
- Latest commit SHA for each reference router
- List of changed files in provider directories

### 2. Router Agent Update Check

**File**: `../../../../Ti/apps/core/router/cmd/routerd\router_agent.go`

**API Endpoint**:
```go
// HTTP endpoint for update check
func (ra *RouterAgent) HandleUpdateCheck(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        Provider string `json:"provider"`
        Commit   string `json:"commit"`
        Timestamp string `json:"timestamp"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Trigger update check
    err := ra.CheckProviderUpdates(req.Provider)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Update check triggered",
    })
}
```

### 3. Provider Update Analysis

**Function**: `CheckProviderUpdates(provider string)`

**Steps**:
1. Fetch latest commit from reference router
2. Get diff for provider-specific files
3. Analyze changes (additions, modifications, deletions)
4. Generate patch suggestions
5. Store in memory for review

**Example**:
```go
func (ra *RouterAgent) CheckProviderUpdates(provider string) error {
    // Reference router mappings
    refRouters := map[string]struct{
        github string
        path   string
    }{
        "claude": {
            github: "https://github.com/decolua/9router",
            path:   "src/providers/claude.ts",
        },
        "gemini": {
            github: "https://github.com/tkgo11/LLMCookieBridge",
            path:   "src/llm_cookie_bridge/providers/gemini.py",
        },
    }

    ref, ok := refRouters[provider]
    if !ok {
        return fmt.Errorf("unknown provider: %s", provider)
    }

    // Fetch latest commit
    latestCommit, err := ra.fetchLatestCommit(ref.github, ref.path)
    if err != nil {
        return err
    }

    // Get diff
    diff, err := ra.fetchDiff(ref.github, ref.path, latestCommit)
    if err != nil {
        return err
    }

    // Analyze changes
    changes := ra.analyzeChanges(diff)

    // Generate patch suggestions
    patches := ra.generatePatches(changes, provider)

    // Store in memory
    ra.memory["pending_updates"] = map[string]interface{}{
        "provider": provider,
        "patches":  patches,
        "commit":  latestCommit,
    }

    return nil
}
```

### 4. Patch Generation

**Function**: `generatePatches(changes, provider)`

**Output**:
- Unified diff format
- File paths relative to Ti Router structure
- Change descriptions
- Risk assessment

**Example Output**:
```json
{
  "provider": "claude",
  "patches": [
    {
      "file": "layers/providers/cookie/claude.go",
      "type": "modify",
      "description": "Add cookie refresh mechanism from 9Router",
      "diff": "@@ -15,7 +15,10 @@\n func (cp *ClaudeProvider) Refresh() error {\n-    // TODO: Implement refresh\n+    // Extract session token from cookie\n+    sessionToken := cp.extractSessionToken()\n+    // Validate token\n+    if err := cp.validateToken(sessionToken); err != nil {\n+        return err\n+    }\n+    // Update cookie if needed\n+    return cp.updateCookie()\n }",
      "risk": "low"
    }
  ],
  "commit": "abc123",
  "timestamp": "2026-05-08T00:00:00Z"
}
```

### 5. Update Approval and Application

**API Endpoint**: `/api/router-agent/apply-updates`

**Request**:
```json
{
  "provider": "claude",
  "approved": true,
  "patches": ["patch_id_1", "patch_id_2"]
}
```

**Process**:
1. Validate approval
2. Apply patches to files
3. Run tests
4. If tests pass, restart router
5. Log update

**Implementation**:
```go
func (ra *RouterAgent) ApplyUpdates(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Provider  string   `json:"provider"`
        Approved bool     `json:"approved"`
        Patches  []string `json:"patches"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    if !req.Approved {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "message": "Update rejected by user",
        })
        return
    }

    // Apply patches
    err := ra.applyPatches(req.Provider, req.Patches)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Run tests
    testResult := ra.runTests()
    if !testResult.Success {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "message": "Tests failed",
            "test_output": testResult.Output,
        })
        return
    }

    // Restart router
    err = ra.restartRouter()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Updates applied successfully",
    })
}
```

## 🔧 Configuration

### Reference Router Mappings

**File**: `../../../../Ti/apps/core/router/cmd/routerd\reference_routers.json`

```json
{
  "providers": {
    "claude": {
      "reference": "9router",
      "github": "https://github.com/decolua/9router",
      "path": "src/providers/claude.ts",
      "target": "layers/providers/cookie/claude.go"
    },
    "gemini": {
      "reference": "llmcookiebridge",
      "github": "https://github.com/tkgo11/LLMCookieBridge",
      "path": "src/llm_cookie_bridge/providers/gemini.py",
      "target": "layers/providers/cookie/gemini.go"
    },
    "chatgpt": {
      "reference": "llmcookiebridge",
      "github": "https://github.com/tkgo11/LLMCookieBridge",
      "path": "src/llm_cookie_bridge/providers/chatgpt.py",
      "target": "layers/providers/cookie/chatgpt.go"
    },
    "perplexity": {
      "reference": "llmcookiebridge",
      "github": "https://github.com/tkgo11/LLMCookieBridge",
      "path": "src/llm_cookie_bridge/providers/perplexity.py",
      "target": "layers/providers/cookie/perplexity.go"
    }
  }
}
```

### Update Settings

**Environment Variables**:
```bash
# Enable auto-update checks
ROUTER_AGENT_AUTO_UPDATE=true

# Update check interval (hours)
ROUTER_AGENT_UPDATE_INTERVAL=24

# Require approval for updates
ROUTER_AGENT_REQUIRE_APPROVAL=true

# Auto-apply low-risk updates
ROUTER_AGENT_AUTO_APPLY_LOW_RISK=false
```

## 📊 Monitoring and Logging

### Update History

**File**: `Z:\02_CORE\06_database\router_agent_updates.json`

```json
{
  "updates": [
    {
      "id": "update_001",
      "provider": "claude",
      "reference_commit": "abc123",
      "applied_at": "2026-05-08T00:00:00Z",
      "status": "success",
      "patches_applied": 2
    }
  ]
}
```

### Logging

**Log File**: `Z:\02_CORE\06_database\logs\router_agent_updates.log`

```
2026-05-08 00:00:00 [INFO] Update check triggered for claude
2026-05-08 00:00:05 [INFO] Found 2 patches from 9Router
2026-05-08 00:00:10 [INFO] Generated patch suggestions
2026-05-08 00:01:00 [INFO] User approved updates
2026-05-08 00:01:30 [INFO] Applied patches successfully
2026-05-08 00:01:35 [INFO] Tests passed
2026-05-08 00:01:40 [INFO] Router restarted successfully
```

## 🚨 Safety Mechanisms

### 1. Approval Required

All updates require user approval before application.

### 2. Test Validation

Updates are only applied if tests pass.

### 3. Rollback Support

Failed updates can be rolled back automatically.

### 4. Risk Assessment

Each patch is assessed for risk level (low, medium, high).

### 5. Backup Before Update

Automatic backup before applying updates.

## 🔗 Related Documentation

- [Router Reference](./ROUTER_REFERENCE.md)
- [Cookie Provider Docs](./INDEX.md)
- [GitHub Action Workflow](../../../../Ti/.github/workflows/monitor-router-updates.yml)
- [Router Agent Implementation](../../../../Ti/apps/core/router/cmd/routerd/router_agent.go)

---

**Generated by**: Devin CLI
**Date**: 2026-05-08
