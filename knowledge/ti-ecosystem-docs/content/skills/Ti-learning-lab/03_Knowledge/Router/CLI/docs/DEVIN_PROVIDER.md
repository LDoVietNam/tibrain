# Devin Provider Documentation

> **Provider**: Devin (Cognition AI)
> **Type**: AI Agent / Session-based Autonomous Agent
> **Status**: Beta / Experimental
> **Last Updated**: 2026-05-16 (Updated with GitHub research)

---

## 📋 Overview

Devin là một AI agent tự động (autonomous AI software engineer) được phát triển bởi Cognition Labs. Khác với các LLM providers truyền thống (OpenAI, Anthropic, v.v.), Devin là một agent-based system có khả năng:

- Tự động lập trình và debug code
- Quản lý toàn bộ lifecycle của development tasks
- Tự tạo và chạy tools/scripts
- Làm việc với nhiều files và projects lớn

## 🎯 Key Characteristics

### Agent-based Architecture
- **Session-based**: Mỗi request tạo một session, poll đến khi hoàn thành
- **Autonomous**: Tự quyết định và thực hiện các actions
- **Tool-using**: Tự tạo và sử dụng tools để hoàn thành tasks
- **Multi-step**: Không phải chat completion đơn giản

### Provider Type
- **Không phải LLM provider truyền thống**
- **Agent orchestration layer** trên các LLM models
- **Sử dụng nhiều underlying models** (Claude, GPT, Gemini, v.v.)

## 🔧 Configuration

### Environment Variables
```bash
export DEVIN_API_KEY="your-api-key"
export DEVIN_ORG_ID="your-org-id"
```

### Config File
```yaml
APIKeys:
  devin: "your-api-key"
  devin_org_id: "your-org-id"
```

## 🔐 Authentication

### API Versions & Credentials

**Devin API v3 (Current - Recommended)**
- **Base URL**: `https://api.devin.ai/v3/organizations/*` hoặc `https://api.devin.ai/v3/enterprise/*`
- **Credentials**: Service User API Key với prefix `cog_`
- **Authorization**: `Bearer cog_your_token_here`
- **Features**: RBAC permissions, session attribution, audit trails

**Devin API v1 (Legacy - Deprecated)**
- **Base URL**: `https://api.devin.ai/v1/*`
- **Credentials**: Personal API Key (`apk_user_`) hoặc Service API Key (`apk_`)
- **Status**: Deprecated nhưng vẫn hoạt động
- **Migration**: Cần migrate sang v3 cho RBAC và features mới

**Devin API v2 (Legacy - Enterprise)**
- **Base URL**: `https://api.devin.ai/v2/enterprise/*`
- **Credentials**: Enterprise admin personal API key
- **Status**: Deprecated, dùng cho enterprise management
- **Features**: Cross-org management, analytics, audit logs

### Authentication Methods

**Service User API Keys (Recommended for Automation)**
```bash
# Create service user in Settings > Service Users
# Generate API key with cog_ prefix
curl -X POST "https://api.devin.ai/v3/organizations/$DEVIN_ORG_ID/sessions" \
  -H "Authorization: Bearer $DEVIN_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Create a Python script"}'
```

**Personal Access Tokens (Closed Beta)**
```bash
# For human programmatic access
# Authenticate as your own identity
curl "https://api.devin.ai/v3/organizations/$DEVIN_ORG_ID/sessions" \
  -H "Authorization: Bearer $YOUR_PAT"
```

**Windsurf Authentication**
```bash
# Enterprise users can authenticate via Windsurf
devin auth login
# Select "Log in with Windsurf for Enterprise"
```

### Credentials Format Comparison

| Format | Prefix | API Version | Use Case |
|--------|--------|-------------|----------|
| Service User API Key | `cog_` | v3 | Automation, CI/CD (recommended) |
| Personal Access Token | `cog_` | v3 | Human programmatic access |
| Personal API Key | `apk_user_` | v1, v2 | Legacy user authentication |
| Service API Key | `apk_` | v1 | Legacy organization-scoped |
| Windsurf Service Key | Custom | Windsurf API | Windsurf enterprise analytics |

**Note**: Credentials format hiện tại trong router.env (`user-xxx_org-yyy:zzz`) không phải format chuẩn của Devin API. Format này có thể là:
- Format đặc biệt cho CLI/Windsurf
- Format cũ của v1 API
- Custom format cho internal use

**Recommendation**: Sử dụng v3 API với `cog_` prefix credentials cho production use.

## 📦 Available Models

### Model List (16 models)
1. **swe-1-6** - FREE trong 3 tháng (200 tok/s) 🔥
2. **swe-1-6-fast** - Paid (950 tok/s)
3. **devin-auto** - Auto selection (default)
4. **swe** - SWE models (coding optimized)
5. **swe-fast** - SWE fast (quick edits)
6. **opus** - Claude Opus (strongest reasoning)
7. **sonnet** - Claude Sonnet (balanced)
8. **gpt** - GPT models
9. **gpt-4** - GPT-4
10. **gpt-4-turbo** - GPT-4 Turbo
11. **codex** - CodeX models
12. **gemini** - Gemini models
13. **gemini-pro** - Gemini Pro
14. **kimi** - Kimi models
15. **glm** - GLM models
16. **glm-4** - GLM-4

### Model Categories

**Anthropic Models:**
- `opus` - Claude Opus (mạnh nhất về reasoning)
- `sonnet` - Claude Sonnet (cân bằng)

**OpenAI Models:**
- `gpt` - GPT models
- `gpt-4` - GPT-4
- `gpt-4-turbo` - GPT-4 Turbo
- `codex` - CodeX models

**Google Models:**
- `gemini` - Gemini models
- `gemini-pro` - Gemini Pro

**Cognition Models (SWE Series):**
- `swe-1-6` - SWE-1.6 (FREE 3 tháng)
- `swe-1-6-fast` - SWE-1.6 Fast
- `swe` - SWE models
- `swe-fast` - SWE fast

**Open Source Models:**
- `kimi` - Kimi models
- `glm` - GLM models
- `glm-4` - GLM-4

## 💰 Pricing

### Plans
- **Free** - Free tier (limited access)
- **Pro** - $20/month với included quota
- **Max** - $200/month với quota lớn hơn
- **Teams** - Usage-based, minimum $80/month
- **Enterprise** - Custom pricing

### Free Tier (SWE-1.6)
- **Model**: `swe-1-6`
- **Duration**: Free trong 3 tháng
- **Speed**: 200 tok/s (via Fireworks)
- **Availability**: Windsurf

### Compute Units (ACUs)
- 15 phút "active Devin work" ≈ 1 ACU
- Usage được tính bằng ACUs hoặc dollars tùy plan

## 🚀 Usage

### Basic Usage
```go
provider := &DevinProvider{}
provider.Init(cfg)

req := ChatRequest{
    Model: "swe-1-6",  // FREE model
    Messages: []Message{
        {Role: "user", Content: "Fix this bug"},
    },
}

resp, err := provider.Chat(ctx, req)
```

### Model Selection
```go
// Auto selection (recommended)
req := ChatRequest{
    Messages: []Message{{Role: "user", Content: "task"}},
}

// Specific model
req := ChatRequest{
    Model: "swe-1-6",  // FREE
    Messages: []Message{{Role: "user", Content: "task"}},
}

// Via options
req := ChatRequest{
    Options: map[string]any{"model": "swe-1-6"},
    Messages: []Message{{Role: "user", Content: "task"}},
}
```

### Model Recommendations
- **Complex refactoring**: `opus` hoặc `gpt`
- **Quick edits/cost-sensitive**: `swe` hoặc `swe-fast`
- **General purpose**: `devin-auto` (tự động chọn)
- **Free usage**: `swe-1-6` (3 tháng free)

## ⚠️ Important Notes

### Authentication Issues
- Credentials format có thể khác nhau giữa CLI và REST API
- CLI credentials có thể KHÔNG work với REST API calls
- Cần verify correct API key type trong Devin dashboard

### Session-based Behavior
- Mỗi request tạo một session
- Poll đến khi session hoàn thành
- Có thể mất từ vài giây đến vài phút
- Timeout mặc định: 30s

### Streaming
- Hỗ trợ streaming nhưng thực chất là fallback
- Devin không support streaming thực sự như các LLM providers

### Status
- Current status: `StatusBeta` / `StatusExperimental`
- Có thể thay đổi trong tương lai

## 🔗 API Endpoints

### Base URL
```
https://api.devin.ai/v3
```

### Key Endpoints
- **Create Session**: `POST /organizations/{org_id}/sessions`
- **Get Session Status**: `GET /organizations/{org_id}/sessions/{session_id}`
- **Get Session Messages**: `GET /organizations/{org_id}/sessions/{session_id}/messages`

## 📚 Documentation

### Official Documentation
- **Official Docs**: https://docs.devin.ai
- **API Reference**: https://docs.devin.ai/api-reference
- **Models Docs**: https://cli.devin.ai/docs/models
- **Blog**: https://cognition.ai/blog
- **Support**: support@devin.ai

### GitHub & Community Resources

**Terraform Provider for Devin AI**
- **Repository**: https://github.com/hirosi1900day/terraform-provider-devin
- **Description**: Terraform provider để quản lý Devin AI knowledge resources
- **Features**:
  - Create/update/delete knowledge resources
  - Set triggers cho knowledge activation
  - Organize knowledge trong folder hierarchies
  - Import existing knowledge vào Terraform state
- **API Version**: v1 (legacy)
- **Base URL**: `https://api.devin.ai/v1`
- **Authentication**: Bearer token (legacy format)
- **Usage Example**:
  ```hcl
  terraform {
    required_providers {
      devin = {
        source  = "hirosi1900day/devin"
        version = "~> 0.0.6"
      }
    }
  }

  provider "devin" {
    api_key = "your_api_key" # Hoặc DEVIN_API_KEY env var
  }

  resource "devin_knowledge" "example" {
    name                = "Sample Knowledge"
    body                = "This is the content"
    trigger_description = "Trigger conditions"
  }
  ```

**Devin CLI Integrations**
- **AWS CLI Agent Orchestrator**: https://github.com/awslabs/cli-agent-orchestrator/issues/148
  - Devin CLI như built-in provider
  - Support autonomous operation với `--permission-mode dangerous`
  - Support MCP server integration

**OpenCode Integration**
- **Feature Request**: https://github.com/anomalyco/opencode/issues/24072
  - Request để thêm Devin provider vào OpenCode

### API Documentation

**v1 API (Legacy)**
- **Overview**: https://docs.devin.ai/api-reference/v1/overview
- **Usage Examples**: https://docs.devin.ai/api-reference/v1/usage-examples
- **Endpoints**: Sessions, Knowledge, Playbooks, Secrets

**v3 API (Current)**
- **Overview**: https://docs.devin.ai/api-reference/overview
- **Authentication**: https://docs.devin.ai/api-reference/authentication
- **Common Flows**: https://docs.devin.ai/api-reference/common-flows
- **Usage Examples**: https://docs.devin.ai/api-reference/v3/usage-examples

**Enterprise API**
- **Enterprise Quick Start**: https://docs.devin.ai/api-reference/getting-started/enterprise-quickstart
- **Service Keys**: https://docs.windsurf.com/plugins/accounts/api-reference/api-introduction

## 🧪 Testing

### Test Configuration
```bash
# Set environment variables
export DEVIN_API_KEY="your-api-key"
export DEVIN_ORG_ID="your-org-id"

# Build router
cd Z:\10_WORKPLACE\Ti\apps\router
go build ./...
```

### Test Provider
```go
cfg, _ := config.Load("")
// Devin credentials sẽ tự động load từ environment variables

// Provider sẽ được register qua bootstrap
reg, _ := providers.BuildRegistryFromConfig(cfg, nil)
provider, _ := reg.Get("devin")
```

## 🎯 Best Practices

### Model Selection
- Sử dụng `swe-1-6` trong khi còn free (3 tháng)
- Test multiple models để tìm model phù hợp với use case
- Sử dụng `devin-auto` cho general purpose

### Cost Management
- Monitor usage trong free tier
- Upgrade sang paid plan khi cần thường xuyên
- Sử dụng `swe-fast` cho cost-sensitive tasks

### Task Suitability
- **Complex refactoring**: Devin rất mạnh cho multi-file edits
- **Bug fixing**: Tốt cho debugging và testing
- **Architecture changes**: Có thể handle large refactors
- **Simple edits**: Có thể overkill, dùng SWE models thay vì Devin

## 🔧 Implementation Details

### File Locations
- **Provider**: `layers/provider/devin.go`
- **Bootstrap**: `layers/provider/bootstrap.go`
- **Descriptor**: `layers/provider/descriptor.go`
- **Config**: `layers/config/config_methods.go`
- **Credentials Validator**: `layers/provider/credentials_validator.go`
- **Cached Provider**: `layers/provider/cached_provider.go`
- **Retry Provider**: `layers/provider/retry_provider.go`
- **Version Provider**: `layers/provider/version_provider.go`
- **Model Selector**: `layers/provider/model_selector.go`
- **Knowledge Manager**: `layers/provider/knowledge_manager.go`
- **Session Manager**: `layers/provider/session_manager.go`
- **Permission System**: `layers/provider/permission_system.go`
- **Health Checker**: `layers/provider/health_checker.go`

### Key Methods
- `Chat()` - Tạo session, poll, return result
- `ChatStream()` - Fallback đến Chat, emit single chunk
- `Models()` - Return 16 available models
- `DefaultModel()` - Return "devin-auto"

### Bootstrap Integration
```go
// Đã thêm vào bootstrap.go
if p, err := buildDevinProvider(cfg); err == nil && p != nil {
    reg.Register(p)
}
```

### Advanced Patterns Implemented

#### 1. Credentials Format Validation
```go
// layers/provider/credentials_validator.go
validator := NewCredentialValidator()
pattern, err := validator.Validate(apiKey)
// Detects: cog_ (v3), apk_user_ (v1), apk_ (v1), custom formats
```

#### 2. Caching Applied to Providers
```go
// layers/provider/cached_provider.go
cachedProvider := NewCachedProvider(provider, cache, ttl)
// Wraps provider with semantic caching (memory + SQLite)
```

#### 3. Sophisticated Retry Strategy
```go
// layers/provider/retry_provider.go
retryProvider := NewRetryProvider(provider, retryConfig)
// Exponential backoff, jitter, status code-aware retry
```

#### 4. Multiple API Versions Support
```go
// layers/provider/version_provider.go
versionProvider := NewVersionProvider(provider, versionConfig)
// Auto-detect v1/v2/v3 from credentials, graceful fallback
```

#### 5. Model Auto-Selection
```go
// layers/provider/model_selector.go
modelSelector := NewModelSelectorProvider(provider, modelConfig)
// Task-based model selection (refactoring -> opus, quick_edit -> swe-fast)
```

#### 6. Knowledge Management
```go
// layers/provider/knowledge_manager.go
knowledgeManager := NewKnowledgeManager()
// Knowledge resources with triggers, context-aware routing
```

#### 7. Advanced Session Management
```go
// layers/provider/session_manager.go
sessionManager := NewSessionManager()
// Progress tracking, configurable polling, session persistence
```

#### 8. RBAC & Permission System
```go
// layers/provider/permission_system.go
permissionSystem := NewPermissionSystem()
// Role-based access control, granular permissions, audit trails
```

#### 9. Advanced Health Check System
```go
// layers/provider/health_checker.go
healthChecker := NewHealthChecker()
// Continuous health monitoring, status aggregation, proactive alerts
```

## 📊 Performance

### SWE-1.6 Performance
- Cải thiện 11% so với SWE-1.5 trên SWE-Bench Pro
- Tối ưu cho cả intelligence và model UX
- Giảm overthinking và looping behavior
- Tăng parallel tool calls

### Speed
- **SWE-1.6 (free)**: 200 tok/s
- **SWE-1.6-fast**: 950 tok/s (industry-leading)

## 🆓 Free Tier Summary

**Available:** YES
- **Model**: `swe-1-6`
- **Duration**: 3 months
- **Speed**: 200 tok/s
- **Platform**: Windsurf
- **Provider**: Fireworks

**Note:** Free tier có giới hạn và chỉ có trong promotional period.

---

## 📝 Changelog

### 2026-05-16 (Updated - Advanced Patterns Implementation)
- ✅ Thêm Devin provider vào router
- ✅ Cập nhật 16 models (bao gồm SWE-1.6 và SWE-1.6-fast)
- ✅ Thêm environment variable mapping
- ✅ Thêm vào bootstrap process
- ✅ Cấu hình credentials trong `Z:\00_SECRET\router.env`
- ✅ Thêm Authentication section với API versions (v1, v2, v3)
- ✅ Thêm GitHub & Community Resources (terraform-provider-devin, CLI integrations)
- ✅ Thêm credentials format comparison table
- ✅ Thêm API documentation links (v1, v3, Enterprise)
- ✅ **NEW: Implement Credentials Format Validation** - Auto-detect API versions from credentials
- ✅ **NEW: Implement Caching Applied to Providers** - Semantic caching wrapper
- ✅ **NEW: Implement Sophisticated Retry Strategy** - Exponential backoff with jitter
- ✅ **NEW: Implement Multiple API Versions Support** - Graceful version migration
- ✅ **NEW: Implement Model Auto-Selection** - Task-based intelligent model selection
- ✅ **NEW: Implement Knowledge Management** - Context-aware knowledge routing
- ✅ **NEW: Implement Advanced Session Management** - Progress tracking & persistence
- ✅ **NEW: Implement RBAC & Permission System** - Role-based access control
- ✅ **NEW: Implement Advanced Health Check System** - Continuous monitoring

### Model Updates
- Thêm `swe-1-6` (FREE 3 tháng)
- Thêm `swe-1-6-fast` (Paid)
- Cập nhật danh sách models từ 1 → 16 models

---

**Generated**: 2026-05-16
**Router Version**: Current
**Devin API Version**: v3
