# Agent Integration Guide

## Overview
This guide provides step-by-step instructions for integrating the 5 specialized agents into Ti Router.

## Prerequisites
- Ti Router installed and running
- Configuration files placed in `configs/base/`
- Required environment variables set
- Access to external services (Cloudflare, S3, etc.)

---

## 1. Cloudflare Agent Integration

### Purpose
Provides tunnel/Tailscale access control, DDoS protection, and CDN management.

### Setup Steps

#### 1.1 Configure Environment Variables
```bash
export CLOUDFLARE_ACCOUNT_ID="your-account-id"
export CLOUDFLARE_API_TOKEN="your-api-token"
export TAILSCALE_API_KEY="your-tailscale-api-key"
export JWT_SECRET="your-jwt-secret"
```

#### 1.2 Create Cloudflare Tunnel
```bash
# Install cloudflared
# Download from: https://github.com/cloudflare/cloudflared/releases

# Create tunnel
cloudflared tunnel create ti-router-tunnel

# Get tunnel token
cloudflared tunnel token ti-router-tunnel
```

#### 1.3 Configure Tailscale ACL
Create `tailscale-acl.json`:
```json
{
  "acls": [
    {
      "action": "accept",
      "src": ["group:admins"],
      "dst": ["tag:router:*"]
    }
  ]
}
```

#### 1.4 Start Cloudflare Tunnel
```bash
cloudflared tunnel run ti-router-tunnel
```

#### 1.5 Verify Integration
```bash
curl https://router.example.com/health
```

### Troubleshooting
- **Tunnel not connecting**: Check API token and account ID
- **Tailscale access denied**: Verify ACL policy
- **JWT authentication failing**: Check JWT_SECRET

---

## 2. OAuth Agent Integration

### Purpose
Manages OAuth PKCE flows for multiple AI providers (Claude, Google, Gemini).

### Setup Steps

#### 2.1 Configure Environment Variables
```bash
export GOOGLE_CLIENT_ID="your-google-client-id"
export GOOGLE_CLIENT_SECRET="your-google-client-secret"
export GEMINI_CLIENT_ID="your-gemini-client-id"
export GEMINI_CLIENT_SECRET="your-gemini-client-secret"
export ENCRYPTION_KEY="your-encryption-key"
```

#### 2.2 Enable OAuth Providers
Edit `configs/base/oauth.yaml`:
```yaml
providers:
  claude:
    enabled: true
  google:
    enabled: true
  gemini:
    enabled: true
```

#### 2.3 Configure OAuth Callbacks
Set up callback URLs:
- Claude: `http://localhost:54545/callback`
- Google: `http://localhost:54545/callback`
- Gemini: `http://localhost:54545/callback`

#### 2.4 Test OAuth Flow
```bash
# Start OAuth flow
curl http://localhost:8080/v1/oauth/authorize?provider=claude

# Exchange code for token
curl -X POST http://localhost:8080/v1/oauth/token \
  -H "Content-Type: application/json" \
  -d '{"code": "your-code", "state": "your-state"}'
```

### Troubleshooting
- **Authorization URL invalid**: Check client_id and redirect_uri
- **Token exchange failing**: Verify client_secret and state parameter
- **Refresh token exhausted**: User must re-authenticate

---

## 3. File Upload Agent Integration

### Purpose
Handles file upload patterns for AI providers with two-stage upload and CDN integration.

### Setup Steps

#### 3.1 Configure S3 Storage
```bash
export S3_ENDPOINT="https://s3.amazonaws.com"
export S3_ACCESS_KEY="your-access-key"
export S3_SECRET_KEY="your-secret-key"
export S3_BUCKET="your-bucket"
export S3_REGION="us-east-1"
```

#### 3.2 Configure Upload Settings
Edit `configs/base/file-upload.yaml`:
```yaml
providers:
  claude:
    enabled: true
    max_size: 52428800  # 50MB
```

#### 3.3 Test File Upload
```bash
# Stage 1: Request upload URL
curl -X POST http://localhost:8080/v1/upload/stage \
  -H "Content-Type: application/json" \
  -d '{"filename": "test.png", "size": 1024}'

# Stage 2: Upload to S3
curl -X POST <upload_url> \
  -F "file=@test.png"

# Stage 3: Get final URL
curl http://localhost:8080/v1/upload/complete/<upload_id>
```

### Troubleshooting
- **Upload URL generation failing**: Check S3 credentials
- **Upload to S3 failing**: Verify bucket permissions
- **File size too large**: Adjust max_size in config

---

## 4. Storage Agent Integration

### Purpose
Manages multi-storage backend (Postgres, Git, Object, Local) with synchronization.

### Setup Steps

#### 4.1 Configure PostgreSQL
```bash
export PGSTORE_DSN="postgres://user:password@localhost:5432/ti_router"
export PGSTORE_SCHEMA="public"
```

#### 4.2 Configure Git Storage
```bash
export GITSTORE_GIT_URL="https://github.com/user/ti-router-config.git"
export GITSTORE_GIT_USERNAME="your-username"
export GITSTORE_GIT_TOKEN="your-github-token"
export GITSTORE_GIT_BRANCH="main"
```

#### 4.3 Configure Object Storage
```bash
export OBJECTSTORE_ENDPOINT="https://s3.amazonaws.com"
export OBJECTSTORE_ACCESS_KEY="your-access-key"
export OBJECTSTORE_SECRET_KEY="your-secret-key"
export OBJECTSTORE_BUCKET="your-bucket"
```

#### 4.4 Initialize Storage
```bash
# Create database schema
psql $PGSTORE_DSN -f scripts/init_schema.sql

# Initialize Git repository
git clone $GITSTORE_GIT_URL $DATA_DIR/git_storage

# Test storage backends
curl http://localhost:8080/v1/storage/test
```

### Troubleshooting
- **Postgres connection failing**: Check DSN and database permissions
- **Git push failing**: Verify token and branch permissions
- **Object storage failing**: Check credentials and bucket access

---

## 5. Conversation Agent Integration

### Purpose
Manages conversation lifecycle, context tracking, and session management.

### Setup Steps

#### 5.1 Configure Conversation Settings
Edit `configs/base/conversation.yaml`:
```yaml
providers:
  claude:
    enabled: true
    max_context_tokens: 200000

context_management:
  enable_optimization: true
  summarize_old_messages: true
```

#### 5.2 Test Conversation Management
```bash
# Create conversation
curl -X POST http://localhost:8080/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{"provider": "claude", "model": "claude-3-5-sonnet"}'

# Add message
curl -X POST http://localhost:8080/v1/conversations/<id>/messages \
  -H "Content-Type: application/json" \
  -d '{"role": "user", "content": "Hello"}'

# Get conversation
curl http://localhost:8080/v1/conversations/<id>
```

### Troubleshooting
- **Context window overflow**: Reduce max_context_tokens or enable summarization
- **Session restoration failing**: Check storage backend connectivity
- **Conversation not persisting**: Verify storage configuration

---

## Go-Only Components Integration

### UsageTracker Integration

#### Setup
UsageTracker is automatically integrated into LearningService. No additional setup required.

#### Monitor Statistics
```bash
# Get provider statistics
curl http://localhost:8080/v1/stats/providers

# Get model statistics
curl http://localhost:8080/v1/stats/models

# Get all statistics
curl http://localhost:8080/v1/stats/all
```

### AuthGuard Integration

#### Setup
AuthGuard is automatically integrated into Router. Configure in code:

```go
// In router.go
authGuard := auth.NewAuthGuard()
authGuard.SetRequireLogin(true)
authGuard.AddAlwaysProtected("/api/shutdown")
authGuard.AddProtectedAPIPath("/api/keys")
```

#### Test Authentication
```bash
# Without authentication (should fail)
curl http://localhost:8080/api/keys

# With CLI token
curl -H "x-ti-cli-token: <token>" http://localhost:8080/api/keys

# With JWT token
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/keys
```

---

## Verification Checklist

After integrating all agents, verify:

- [ ] Cloudflare tunnel is running and accessible
- [ ] OAuth flows work for enabled providers
- [ ] File uploads work with S3 backend
- [ ] Storage backends are synchronized
- [ ] Conversations are persisted and restored
- [ ] UsageTracker is collecting statistics
- [ ] AuthGuard is protecting sensitive endpoints
- [ ] All configuration files are valid
- [ ] Environment variables are set correctly
- [ ] No build errors
- [ ] Router starts successfully

---

## Monitoring and Maintenance

### Health Checks
```bash
# Agent health
curl http://localhost:8080/v1/agents/health

# Storage health
curl http://localhost:8080/v1/storage/health

# OAuth health
curl http://localhost:8080/v1/oauth/health
```

### Log Locations
- Cloudflare: `logs/cloudflare.log`
- OAuth: `logs/oauth.log`
- File Upload: `logs/file-upload.log`
- Storage: `logs/storage.log`
- Conversation: `logs/conversation.log`

### Backup Strategy
- PostgreSQL: Daily backups
- Git: Automatic commits
- Object Storage: Versioning enabled
- Local: Weekly sync to cloud

---

## Rollback Plan

If issues occur:

1. Disable problematic agent in config
2. Restart router
3. Restore from backup if needed
4. Check logs for errors
5. Consult troubleshooting section

---

## References
- Agent definitions: `content/agents/agents/`
- Configuration files: `configs/base/`
- Documentation: `docs/03-operations/`
