# Modal GLM-5 Setup - Configuration Notes

## Summary
- **Date**: 2026-04-17
- **Goal**: Connect OpenCode/Kilo to Modal GLM-5.1 endpoint (direct, no router)
- **Status**: ✅ Complete (except API keys expired)

---

## Architecture Decision

**Direct connection** (no Ti Router):
```
OpenCode/Kilo → Modal API (https://api.us-west-2.modal.direct/v1)
```

NOT via Ti Backend/Router because:
- Ti Backend's BeeknoeeProvider hardcodes Claude models only
- Modal endpoint is OpenAI-compatible, direct is simpler

---

## Files Modified

### 1. OpenCode Config
**File:** `Z:\02_CORE\_cli\opencode.json`

```json
{
  "provider": {
    "modal": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Modal Direct",
      "options": {
        "baseURL": "https://api.us-west-2.modal.direct/v1",
        "apiKey": "{env:MODAL_API_KEY}"
      },
      "models": {
        "zai-org/GLM-5-FP8": { "name": "GLM-5 FP8" },
        "zai-org/GLM-5-FP8-2": { "name": "GLM-5 FP8-2" },
        "zai-org/GLM-5.1-FP8": { "name": "GLM-5.1 FP8" }
      }
    }
  },
  "model": "modal/zai-org/GLM-5.1-FP8"
}
```

### 2. Kilo Config
**File:** `Z:\02_CORE\_cli\.kilo\kilo.json`

```json
{
  "provider": {
    "modal": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Modal GLM-5.1",
      "options": {
        "baseURL": "https://api.us-west-2.modal.direct/v1",
        "apiKey": "{env:MODAL_API_KEY}"
      },
      "models": {
        "zai-org/GLM-5-FP8": { "name": "GLM-5 FP8" },
        "zai-org/GLM-5-FP8-2": { "name": "GLM-5 FP8-2" },
        "zai-org/GLM-5.1-FP8": { "name": "GLM-5.1 FP8" }
      }
    }
  },
  "model": "modal/zai-org/GLM-5.1-FP8"
}
```

### 3. Secrets Storage
**File:** `Z:\00_SECRET\.routerenv`

```bash
MODAL_API_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxx
MODAL_API_KEY_2=yyyyyyyyyyyyyyyyyyyyyyyyyyy
TI_ROUTER_API_KEY=sk-trepremium-dev
```

### 4. Cline Config (Updated)
**File:** `C:\Users\MIN\.cline\data\settings\providers.json`

Added `modal` provider with direct Modal endpoint.

### 5. Aider Config
**File:** `Z:\02_CORE\_cli\.env`

```bash
MODAL_API_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

---

## Environment Variables Loading

**Script:** `Z:\02_CORE\_cli\load-env.ps1`

Loads secrets from `Z:\00_SECRET\.routerenv` into PowerShell environment.

---

## Current Status

| Tool | Provider | Connection | Model | Status |
|------|----------|------------|-------|--------|
| OpenCode | modal | Direct Modal ✅ | GLM-5.1 FP8 | ⚠️ Key expired |
| Kilo | modal | Direct Modal ✅ | GLM-5.1 FP8 | ⚠️ Key expired |
| Cline | modal | Direct Modal ✅ | GLM-5.1 FP8 | ⚠️ Key expired |
| Aider | modal | Direct Modal ✅ | GLM-5.1 FP8 | ⚠️ Key expired |

---

## Required Actions

### 1. Get New Modal API Keys
- Visit: https://modal.com/glm-5-endpoint
- Create 2 new credentials
- Update `Z:\00_SECRET\.routerenv`:
  ```
  MODAL_API_KEY=new_key_1
  MODAL_API_KEY_2=new_key_2
  ```

### 2. Reload Environment
```powershell
powershell -File "Z:\02_CORE\_cli\load-env.ps1"
```

### 3. Verify Connection
```powershell
powershell -File "Z:\02_CORE\_cli\test-modal.ps1"
```

---

## Ti Backend (1806) - Not Used for Modal

**Why not use Ti Backend for Modal:**
- Ti Backend's `BeeknoeeProvider` (alias for Modal) has hardcoded Claude models
- Requires code changes in `modal.go` to support GLM-5 models
- Current config uses Ti Backend only for Claude via OpenRouter

**Ti Backend config** (`Z:\01_PROJECTS\Ti-backend-ui-synced\ti.json`):
```json
{
  "port": 1806,
  "model": "qwenvsclaude-3.6",
  "provider": "openrouter",
  "api_keys": { ... }
}
```

---

## Notes

- OpenCode/Kilo support custom OpenAI-compatible providers via `@ai-sdk/openai-compatible`
- Environment variable substitution: `{env:MODAL_API_KEY}`
- Models must be defined in provider config for client UI
- Modal endpoint: `https://api.us-west-2.modal.direct/v1` (OpenAI-compatible)
- All CLI tools now configured for direct Modal connection (no router)

---

## Testing Commands

```powershell
# Load env vars
.\load-env.ps1

# Test Modal API
.\test-modal.ps1

# Run OpenCode
cd Z:\02_CORE\_cli && opencode

# Run Kilo
cd Z:\02_CORE\_cli && kilo

# Run Cline (VS Code extension) - uses provider config

# Run Aider
cd Z:\02_CORE\_cli && aider --model modal/zai-org/GLM-5.1-FP8 --openai-compatible-base-url https://api.us-west-2.modal.direct/v1
```
