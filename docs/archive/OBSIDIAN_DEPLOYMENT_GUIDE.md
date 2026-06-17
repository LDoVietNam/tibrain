# Obsidian Integration Deployment Guide

This guide provides step-by-step instructions for deploying the Obsidian ↔ Ti Brain integration system.

## Prerequisites

- Node.js v22+
- Python 3.11+ with uv (recommended) or pip
- Go 1.20+ (for frontmatter mapping)
- Obsidian desktop app with Local REST API plugin
- Obsidian Sync account (for headless sync)
- Ti Brain knowledge base directory

## Component Status

| Component | Status | Notes |
|-----------|--------|-------|
| obsidian-mcp-server | ✅ Built | TypeScript MCP server for real-time access |
| obsidian-headless | ✅ Installed | CLI tool for scheduled sync automation |
| Frontmatter Mapper (Go) | ✅ Tested | Tag/scope validation and field mapping |
| Sync Script (Python) | ✅ Tested | Orchestration and file watching |
| Documentation | ✅ Complete | Architecture and master plan documents |

## Deployment Steps

### 1. Build and Configure obsidian-mcp-server

The obsidian-mcp-server has been built successfully using npm instead of bun due to binary remapping issues.

**Location:** `apps/tibrain/obsidian-mcp-server/`

**Build Process:**
```bash
cd apps/tibrain/obsidian-mcp-server
npm install  # Use npm instead of bun to avoid binary remapping issues
npm run build
```

**Configuration:**
Set the following environment variables:
```bash
export OBSIDIAN_API_KEY="your-api-key-from-local-rest-api-plugin"
export OBSIDIAN_BASE_URL="http://127.0.0.1:27123"  # Default
export OBSIDIAN_VERIFY_SSL="false"  # For local development
export OBSIDIAN_REQUEST_TIMEOUT_MS="30000"
export OBSIDIAN_ENABLE_COMMANDS="false"  # Set to true to enable command-palette tools
export OBSIDIAN_READ_PATHS=""  # Comma-separated paths (empty = full vault)
export OBSIDIAN_WRITE_PATHS=""  # Comma-separated paths (empty = full vault)
export OBSIDIAN_READ_ONLY="false"  # Set to true for read-only mode
```

**Testing:**
```bash
# Start the MCP server in stdio mode
npm run start:stdio

# Or start in HTTP mode
npm run start:http
```

### 2. Configure obsidian-headless

The obsidian-headless CLI tool has been installed and is ready for configuration.

**Location:** `apps/tibrain/obsidian-headless/`

**Authentication:**
```bash
cd apps/tibrain/obsidian-headless
node cli.js login
```

This will prompt for:
- Email address
- Password
- 2FA code (if enabled)

**List Remote Vaults:**
```bash
node cli.js sync-list-remote
```

**Setup Sync:**
```bash
# Create local vault directory if it doesn't exist
mkdir -p ~/vaults/my-vault

# Setup sync for a specific vault
node cli.js sync-setup --vault "My Vault" --path ~/vaults/my-vault
```

**Run One-Time Sync:**
```bash
node cli.js sync --path ~/vaults/my-vault
```

**Run Continuous Sync:**
```bash
node cli.js sync --path ~/vaults/my-vault --continuous
```

### 3. Deploy Frontmatter Mapper

The Go-based frontmatter mapper has been tested and is ready for use.

**Location:** `apps/tibrain/internal/frontmatter/`

**Testing:**
```bash
cd apps/tibrain
go test ./internal/frontmatter/
```

**Integration:**
The mapper is integrated into the Python sync script and handles:
- Tag normalization (#go → go)
- Scope validation against taxonomy
- Field mapping between formats
- Default value assignment

### 4. Deploy Sync Script

The Python sync script has been tested and is ready for deployment.

**Location:** `apps/tibrain/sync_obsidian.py`

**Dependencies:**
```bash
# Install using uv (recommended)
cd apps/tibrain
uv add pyyaml watchdog

# Or using pip
pip install pyyaml watchdog
```

**Configuration:**
```bash
# Sync Obsidian to Ti Brain
python sync_obsidian.py \
  --direction obsidian-to-tibrain \
  --vault ~/vaults/my-vault \
  --tibrain ~/tibrain \
  --pattern "*.md" \
  --headless-sync

# Continuous sync with file watching
python sync_obsidian.py \
  --direction obsidian-to-tibrain \
  --vault ~/vaults/my-vault \
  --tibrain ~/tibrain \
  --pattern "*.md" \
  --watch
```

**Testing:**
```bash
cd apps/tibrain
uv run --with pyyaml --with watchdog python test_sync.py
```

### 5. Setup Ti Brain Taxonomy

Ensure your Ti Brain knowledge base has the required taxonomy files:

**TAGS.md** - Valid tags for frontmatter validation
**SCOPES.md** - Valid scopes for frontmatter validation

**Example TAGS.md:**
```markdown
- authentication
- authorization
- rate-limiting
- caching
- translation
- retry
- monitoring
- deployment
- troubleshooting
- security
- performance
- testing
- documentation
- go
- typescript
- python
- provider-antigravity
- provider-openai
- provider-claude
- mcp
- router
- cli
- tibrain
- ticrew
```

**Example SCOPES.md:**
```markdown
- auth
- resilience
- integration
- observability
- infrastructure
- providers
- cli
- tibrain
- ticrew
- web
- code
```

### 6. Configure Obsidian Local REST API

1. Install the Local REST API plugin in Obsidian
2. Enable the plugin in Obsidian settings
3. Generate an API key
4. Configure the plugin to listen on the default port (27123)
5. Set the API key in your environment variables

**Plugin Settings:**
- Port: 27123 (default)
- API Key: (generate and set as OBSIDIAN_API_KEY)
- CORS: Enable for development
- SSL: Disable for local development

## Production Deployment

### Environment Variables

Create a `.env` file in `apps/tibrain/`:

```bash
# Obsidian MCP Server
OBSIDIAN_API_KEY=your-production-api-key
OBSIDIAN_BASE_URL=http://127.0.0.1:27123
OBSIDIAN_VERIFY_SSL=false
OBSIDIAN_REQUEST_TIMEOUT_MS=30000
OBSIDIAN_ENABLE_COMMANDS=false
OBSIDIAN_READ_PATHS=
OBSIDIAN_WRITE_PATHS=
OBSIDIAN_READ_ONLY=false

# Sync Configuration
VAULT_PATH=/path/to/obsidian/vault
TIBRAIN_PATH=/path/to/tibrain/knowledge/base
SYNC_PATTERN=*.md
SYNC_DIRECTION=obsidian-to-tibrain
```

### Systemd Service (Linux)

Create a systemd service for continuous sync:

**/etc/systemd/system/obsidian-sync.service:**
```ini
[Unit]
Description=Obsidian to Ti Brain Sync Service
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/apps/tibrain
Environment="PATH=/usr/local/bin:/usr/bin:/bin"
ExecStart=/usr/local/bin/uv run --with pyyaml --with watchdog python sync_obsidian.py \
  --direction obsidian-to-tibrain \
  --vault /path/to/vault \
  --tibrain /path/to/tibrain \
  --pattern "*.md" \
  --watch
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**Enable and start:**
```bash
sudo systemctl enable obsidian-sync
sudo systemctl start obsidian-sync
sudo systemctl status obsidian-sync
```

### Docker Deployment

**Dockerfile for obsidian-mcp-server:**
```dockerfile
FROM node:22-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

ENV OBSIDIAN_BASE_URL=http://host.docker.internal:27123
ENV OBSIDIAN_VERIFY_SSL=false

EXPOSE 3000

CMD ["npm", "run", "start:http"]
```

**docker-compose.yml:**
```yaml
version: '3.8'

services:
  obsidian-mcp-server:
    build: ./obsidian-mcp-server
    ports:
      - "3000:3000"
    environment:
      - OBSIDIAN_API_KEY=${OBSIDIAN_API_KEY}
      - OBSIDIAN_BASE_URL=http://host.docker.internal:27123
      - OBSIDIAN_VERIFY_SSL=false
    extra_hosts:
      - "host.docker.internal:host-gateway"
```

## Troubleshooting

### obsidian-mcp-server Build Fails

**Issue:** Bun binary remapping errors

**Solution:** Use npm instead of bun:
```bash
rm -rf node_modules bun.lockb
npm install
npm run build
```

### obsidian-headless Authentication Fails

**Issue:** "No account logged in" error

**Solution:** Run interactive login:
```bash
node cli.js login
```

Follow the prompts for email, password, and 2FA.

### Sync Script Taxonomy Warnings

**Issue:** "Taxonomy file TAGS.md not found"

**Solution:** Create TAGS.md and SCOPES.md in your Ti Brain directory with the required taxonomy definitions.

### Frontmatter Validation Failures

**Issue:** Tags or scopes marked as invalid

**Solution:** Ensure your tags and scopes match the taxonomy defined in TAGS.md and SCOPES.md. Use the normalize_tag function to format tags correctly.

### MCP Server Connection Refused

**Issue:** Cannot connect to Obsidian Local REST API

**Solution:** 
1. Ensure Obsidian is running
2. Enable the Local REST API plugin
3. Check the port configuration (default 27123)
4. Verify OBSIDIAN_API_KEY is correct

## Monitoring and Maintenance

### Log Locations

- obsidian-mcp-server: Check MCP client logs
- obsidian-headless: Logs to console
- sync_obsidian.py: Python logging output
- Ti Brain: Check frontmatter validation logs

### Health Checks

```bash
# Check obsidian-mcp-server status
curl http://localhost:3000/health

# Check obsidian-headless sync status
node cli.js sync-status --path ~/vaults/my-vault

# Check recent sync activity
tail -f /var/log/obsidian-sync.log
```

### Backup Strategy

1. Regular backups of Obsidian vault (via Obsidian Sync)
2. Version control for Ti Brain knowledge base
3. Backup taxonomy files (TAGS.md, SCOPES.md)
4. Environment variable backups

### Updates

**Update obsidian-mcp-server:**
```bash
cd apps/tibrain/obsidian-mcp-server
git pull
npm install
npm run build
```

**Update obsidian-headless:**
```bash
cd apps/tibrain/obsidian-headless
git pull
npm install
```

**Update sync script:**
```bash
cd apps/tibrain
git pull
uv sync
```

## Next Steps

1. Complete initial sync of existing Obsidian vault
2. Validate frontmatter mapping results
3. Test MCP server integration with your AI tools
4. Set up monitoring and alerting
5. Configure automated backups
6. Document custom tag/scope taxonomy

## Support

For issues or questions:
- Check architecture documentation: `docs/OBSIDIAN_INTEGRATION_ARCHITECTURE.md`
- Check master plan: `docs/OBSIDIAN_INTEGRATION_MASTER_PLAN.md`
- Review executive summary: `docs/OBSIDIAN_SYNC_ARCHITECTURE.md`

## Security Considerations

1. **API Keys**: Never commit OBSIDIAN_API_KEY to version control
2. **Network**: Local REST API should not be exposed to the internet
3. **File Permissions**: Ensure vault directories have appropriate permissions
4. **Validation**: Always validate frontmatter before ingestion
5. **Backup**: Maintain backups before bulk operations

## Performance Tuning

**For large vaults (>10,000 notes):**
- Increase OBSIDIAN_REQUEST_TIMEOUT_MS
- Use batch sync operations
- Consider incremental sync strategies
- Monitor memory usage of sync script

**For high-frequency updates:**
- Use continuous sync mode
- Optimize file watching patterns
- Consider deduplication strategies
- Implement conflict resolution policies
