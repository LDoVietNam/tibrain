# Migration Guide

## Overview
This guide helps you migrate from the previous version of Ti Router to the new version with integrated patterns from 8 router repositories and 5 specialized agents.

## What's New

### New Patterns Applied
1. **Browser Impersonation** - Standardized headers mimicking real browsers
2. **Nested SSE Parsing** - Support for complex streaming responses
3. **WebSocket Streaming** - Delimiter-based message splitting
4. **OAuth PKCE** - Full PKCE implementation with retry logic
5. **Cookie Extraction** - Cross-platform browser cookie extraction
6. **Cookie Filtering** - Provider-specific cookie filtering
7. **UsageTracker** - Real-time statistics tracking
8. **AuthGuard** - HTTP handler protection with JWT and CLI token

### New Agents
1. **Cloudflare Agent** - Tunnel/Tailscale access control
2. **OAuth Agent** - OAuth PKCE flows for multiple providers
3. **File Upload Agent** - Two-stage upload patterns
4. **Storage Agent** - Multi-storage backend management
5. **Conversation Agent** - Conversation lifecycle management

### New Configuration Files
- `configs/base/cloudflare.yaml`
- `configs/base/oauth.yaml`
- `configs/base/file-upload.yaml`
- `configs/base/storage.yaml`
- `configs/base/conversation.yaml`
- `configs/base/monitoring.yaml`

---

## Pre-Migration Checklist

Before starting the migration, ensure:

- [ ] Backup current Ti Router configuration
- [ ] Backup database (if using PostgreSQL)
- [ ] Backup Git storage (if using Git-based storage)
- [ ] Note down current environment variables
- [ ] Document current provider configurations
- [ ] Test current router is working
- [ ] Schedule maintenance window (recommended: 30 minutes)

---

## Migration Steps

### Step 1: Backup Current Installation

```bash
# Backup configuration
cp -r configs configs.backup

# Backup data
cp -r data data.backup

# Backup database (if using PostgreSQL)
pg_dump $PGSTORE_DSN > backup.sql

# Backup Git storage (if using)
tar -czf git-storage-backup.tar.gz $DATA_DIR/git_storage
```

### Step 2: Pull Latest Code

```bash
cd Z:\10_WORKPLACE\Ti
git pull origin main
```

### Step 3: Install New Dependencies

```bash
cd apps/router
go mod download
go mod tidy
```

### Step 4: Set Up New Configuration Files

```bash
# Copy base configurations
cp configs/base/cloudflare.yaml.example configs/base/cloudflare.yaml
cp configs/base/oauth.yaml.example configs/base/oauth.yaml
cp configs/base/file-upload.yaml.example configs/base/file-upload.yaml
cp configs/base/storage.yaml.example configs/base/storage.yaml
cp configs/base/conversation.yaml.example configs/base/conversation.yaml
cp configs/base/monitoring.yaml.example configs/base/monitoring.yaml
```

### Step 5: Configure Environment Variables

Add new environment variables to your `.env` file:

```bash
# Cloudflare Agent
export CLOUDFLARE_ACCOUNT_ID="your-account-id"
export CLOUDFLARE_API_TOKEN="your-api-token"
export TAILSCALE_API_KEY="your-tailscale-api-key"

# OAuth Agent
export GOOGLE_CLIENT_ID="your-google-client-id"
export GOOGLE_CLIENT_SECRET="your-google-client-secret"
export GEMINI_CLIENT_ID="your-gemini-client-id"
export GEMINI_CLIENT_SECRET="your-gemini-client-secret"
export ENCRYPTION_KEY="your-encryption-key"

# File Upload Agent
export S3_ENDPOINT="https://s3.amazonaws.com"
export S3_ACCESS_KEY="your-access-key"
export S3_SECRET_KEY="your-secret-key"
export S3_BUCKET="your-bucket"
export S3_REGION="us-east-1"

# Storage Agent
export PGSTORE_DSN="postgres://user:password@localhost:5432/ti_router"
export PGSTORE_SCHEMA="public"
export GITSTORE_GIT_URL="https://github.com/user/ti-router-config.git"
export GITSTORE_GIT_USERNAME="your-username"
export GITSTORE_GIT_TOKEN="your-github-token"
export OBJECTSTORE_ENDPOINT="https://s3.amazonaws.com"
export OBJECTSTORE_ACCESS_KEY="your-access-key"
export OBJECTSTORE_SECRET_KEY="your-secret-key"
export OBJECTSTORE_BUCKET="your-bucket"

# Monitoring
export SMTP_SERVER="smtp.gmail.com"
export SMTP_USERNAME="your-email@gmail.com"
export SMTP_PASSWORD="your-app-password"
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
```

### Step 6: Build and Test

```bash
# Build router
cd apps/router
go build -o routerd.exe ./cmd/routerd

# Test build
./routerd.exe --version
```

### Step 7: Initialize New Components

```bash
# Initialize UsageTracker (automatic on first run)
# Initialize AuthGuard (automatic on first run)
# Initialize database schema (if using PostgreSQL)
psql $PGSTORE_DSN -f scripts/init_schema.sql
```

### Step 8: Start Router with New Configuration

```bash
# Start router
./routerd.exe --config configs/base/tiserverrouter.yaml
```

### Step 9: Verify Migration

```bash
# Check health
curl http://localhost:8080/health

# Check agents health
curl http://localhost:8080/v1/agents/health

# Check storage health
curl http://localhost:8080/v1/storage/health

# Check monitoring
curl http://localhost:8080/metrics
```

---

## Post-Migration Steps

### Step 1: Verify All Providers Work

Test each provider to ensure authentication still works:

```bash
# Test OpenAI
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4", "messages": [{"role": "user", "content": "test"}]}'

# Test Anthropic
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "claude-3-5-sonnet", "messages": [{"role": "user", "content": "test"}]}'

# Test other providers...
```

### Step 2: Test New Features

```bash
# Test OAuth flow
curl http://localhost:8080/v1/oauth/authorize?provider=claude

# Test file upload
curl -X POST http://localhost:8080/v1/upload/stage \
  -H "Content-Type: application/json" \
  -d '{"filename": "test.png", "size": 1024}'

# Test usage statistics
curl http://localhost:8080/v1/stats/providers
curl http://localhost:8080/v1/stats/models

# Test conversation management
curl -X POST http://localhost:8080/v1/conversations \
  -H "Content-Type: application/json" \
  -d '{"provider": "claude", "model": "claude-3-5-sonnet"}'
```

### Step 3: Enable New Agents (Optional)

Edit configuration files to enable agents:

```yaml
# configs/base/cloudflare.yaml
tunnel:
  enabled: true

# configs/base/oauth.yaml
providers:
  claude:
    enabled: true
```

### Step 4: Configure Monitoring

Set up monitoring and alerting:

```bash
# Enable Prometheus metrics
curl http://localhost:8080/metrics

# Configure Grafana dashboard
# Import dashboard from configs/grafana-dashboard.json
```

### Step 5: Update Documentation

Update your internal documentation to reflect:
- New configuration files
- New environment variables
- New API endpoints
- New monitoring setup

---

## Rollback Plan

If migration fails, follow these steps to rollback:

### Immediate Rollback

```bash
# Stop new router
pkill routerd

# Restore old router
cd apps/router
git checkout <previous-commit>
go build -o routerd.exe ./cmd/routerd

# Restore configuration
rm -rf configs
mv configs.backup configs

# Restore data
rm -rf data
mv data.backup data

# Restore database (if needed)
psql $PGSTORE_DSN < backup.sql

# Start old router
./routerd.exe --config configs/base/tiserverrouter.yaml
```

### Partial Rollback

If only specific components fail:

```bash
# Disable problematic agent
# Edit config file and set enabled: false

# Restart router
pkill routerd
./routerd.exe --config configs/base/tiserverrouter.yaml
```

---

## Troubleshooting

### Build Errors

**Error**: `cannot find package`
```bash
# Solution
cd apps/router
go mod download
go mod tidy
```

**Error**: `undefined: UsageTracker`
```bash
# Solution: Ensure you're on the latest commit
git pull origin main
```

### Runtime Errors

**Error**: `failed to initialize UsageTracker`
```bash
# Solution: Check DATA_DIR environment variable
export DATA_DIR="/path/to/data"
```

**Error**: `AuthGuard initialization failed`
```bash
# Solution: Check JWT_SECRET environment variable
export JWT_SECRET="your-secret"
```

**Error**: `OAuth provider not configured`
```bash
# Solution: Check oauth.yaml configuration
# Ensure client_id and client_secret are set
```

### Configuration Errors

**Error**: `failed to load config file`
```bash
# Solution: Validate YAML syntax
python -c "import yaml; yaml.safe_load(open('configs/base/cloudflare.yaml'))"
```

**Error**: `invalid environment variable`
```bash
# Solution: Check environment variables are set
env | grep -E "CLOUDFLARE|OAUTH|S3|PGSTORE"
```

---

## Breaking Changes

### Configuration Changes

- **Old**: Single configuration file `configs/base/tiserverrouter.yaml`
- **New**: Multiple configuration files for each agent

### Environment Variable Changes

- **New**: Required environment variables for new agents
- **Changed**: Some variable names may have changed

### API Changes

- **New**: `/v1/oauth/*` endpoints for OAuth flows
- **New**: `/v1/upload/*` endpoints for file uploads
- **New**: `/v1/conversations/*` endpoints for conversation management
- **New**: `/v1/stats/*` endpoints for usage statistics
- **Protected**: `/api/keys`, `/api/providers`, `/api/config` now require authentication

### Database Schema Changes

- **New**: UsageTracker tables (if using PostgreSQL)
- **New**: Conversation tables (if using PostgreSQL)

---

## Performance Considerations

### Memory Usage

- **Increase**: ~50-100MB additional memory for UsageTracker and AuthGuard
- **Recommendation**: Ensure at least 2GB RAM available

### Disk Usage

- **Increase**: ~10-50MB for new configuration files
- **Increase**: ~100MB-1GB for usage statistics (30-day retention)
- **Recommendation**: Ensure sufficient disk space

### CPU Usage

- **Increase**: Minimal impact for UsageTracker (periodic cleanup)
- **Increase**: Minimal impact for AuthGuard (token validation)
- **Recommendation**: No special requirements

---

## Support

If you encounter issues during migration:

1. Check logs in `logs/` directory
2. Review this migration guide
3. Check integration guide: `docs/03-operations/AGENT_INTEGRATION_GUIDE.md`
4. Check knowledge base: `Ti-learning-lab/03_Knowledge/Router/Router_Patterns_Deep_Dive.md`
5. Open an issue on GitHub

---

## Summary

This migration adds:
- 8 new patterns from router repositories
- 5 specialized agents
- 6 new configuration files
- Enhanced authentication and monitoring
- Real-time usage statistics tracking

Estimated migration time: 30 minutes (including testing)

Recommended maintenance window: 30-60 minutes
