# Production Setup Guide - Obsidian Integration

This guide provides step-by-step instructions for setting up the Obsidian ↔ Ti Brain integration in a production environment.

## Quick Start (Windows)

1. **Run the setup script:**
   ```powershell
   cd Z:\10_WORKPLACE\Ti\apps\tibrain
   .\setup_obsidian_integration.ps1
   ```

2. **Configure Obsidian Local REST API:**
   - Install Local REST API plugin in Obsidian
   - Generate API key
   - Set port to 27123 (default)

3. **Set API key in .env:**
   ```powershell
   # Edit Z:\10_WORKPLACE\Ti\apps\tibrain\.env
   OBSIDIAN_API_KEY=your-actual-api-key
   ```

4. **Authenticate obsidian-headless:**
   ```powershell
   cd Z:\10_WORKPLACE\Ti\apps\tibrain\obsidian-headless
   node cli.js login
   ```

5. **Setup vault sync:**
   ```powershell
   node cli.js sync-setup --vault "Your Vault Name" --path Z:\10_WORKPLACE\Ti\apps\tibrain\vault
   ```

6. **Start sync service:**
   ```powershell
   cd Z:\10_WORKPLACE\Ti\apps\tibrain
   .\start_sync.ps1
   ```

## Manual Setup

### Prerequisites

- **Node.js** 22+
- **Python** 3.11+
- **uv** (Python package manager)
- **Obsidian** with Local REST API plugin
- **Obsidian Sync** account (for headless sync)

### Directory Structure

```
apps/tibrain/
├── obsidian-mcp-server/          # MCP server (TypeScript)
├── obsidian-headless/           # Headless sync client (Node.js)
├── sync_obsidian.py             # Sync orchestration (Python)
├── TAGS_SIMPLE.md               # Tag taxonomy
├── SCOPES_SIMPLE.md             # Scope taxonomy
├── .env                         # Environment variables
├── start_sync.ps1              # Sync service startup
├── start_mcp_server.ps1        # MCP server startup
└── vault/                       # Obsidian vault (optional)
```

### Environment Variables

Edit `.env` file:

```bash
# Obsidian MCP Server
OBSIDIAN_API_KEY=your-api-key-here
OBSIDIAN_BASE_URL=http://127.0.0.1:27123
OBSIDIAN_VERIFY_SSL=false
OBSIDIAN_REQUEST_TIMEOUT_MS=30000
OBSIDIAN_ENABLE_COMMANDS=false
OBSIDIAN_READ_PATHS=
OBSIDIAN_WRITE_PATHS=
OBSIDIAN_READ_ONLY=false

# Sync Configuration
VAULT_PATH=Z:\10_WORKPLACE\Ti\apps\tibrain\vault
TIBRAIN_PATH=Z:\10_WORKPLACE\Ti\apps\tibrain
SYNC_PATTERN=*.md
SYNC_DIRECTION=obsidian-to-tibrain

# Ti Brain
TIBRAIN_DB_PATH=Z:\10_WORKPLACE\Ti\apps\tibrain\tibrain.db
TIBRAIN_KNOWLEDGE_PATH=Z:\10_WORKPLACE\Ti\apps\tibrain\knowledge
```

### Build Components

#### obsidian-mcp-server

```powershell
cd Z:\10_WORKPLACE\Ti\apps\tibrain\obsidian-mcp-server
npm install
npm run build
```

#### obsidian-headless

```powershell
cd Z:\10_WORKPLACE\Ti\apps\tibrain\obsidian-headless
npm install
```

#### Python Dependencies

```powershell
cd Z:\10_WORKPLACE\Ti\apps\tibrain
uv add pyyaml watchdog
```

### Operations

#### Start Sync Service

```powershell
cd Z:\10_WORKPLACE\Ti\apps\tibrain
.\start_sync.ps1
```

#### Start MCP Server

```powershell
cd Z:\10_WORKPLACE\Ti\apps\tibrain
.\start_mcp_server.ps1
```

#### Manual Sync

```powershell
cd Z:\10_WORKPLACE\Ti\apps\tibrain
uv run --with pyyaml --with watchdog python sync_obsidian.py `
  --direction obsidian-to-tibrain `
  --vault Z:\10_WORKPLACE\Ti\apps\tibrain\vault `
  --tibrain Z:\10_WORKPLACE\Ti\apps\tibrain `
  --pattern "*.md"
```

#### Continuous Sync with File Watching

```powershell
cd Z:\10_WORKPLACE\Ti\apps\tibrain
uv run --with pyyaml --with watchdog python sync_obsidian.py `
  --direction obsidian-to-tibrain `
  --vault Z:\10_WORKPLACE\Ti\apps\tibrain\vault `
  --tibrain Z:\10_WORKPLACE\Ti\apps\tibrain `
  --pattern "*.md" `
  --watch
```

#### obsidian-headless Operations

```powershell
cd Z:\10_WORKPLACE\Ti\apps\tibrain\obsidian-headless

# Login
node cli.js login

# List remote vaults
node cli.js sync-list-remote

# Setup sync
node cli.js sync-setup --vault "Your Vault" --path Z:\10_WORKPLACE\Ti\apps\tibrain\vault

# One-time sync
node cli.js sync --path Z:\10_WORKPLACE\Ti\apps\tibrain\vault

# Continuous sync
node cli.js sync --path Z:\10_WORKPLACE\Ti\apps\tibrain\vault --continuous

# Check status
node cli.js sync-status --path Z:\10_WORKPLACE\Ti\apps\tibrain\vault
```

## Configuration

### Obsidian Local REST API

1. Install the Local REST API plugin in Obsidian
2. Enable the plugin in settings
3. Generate an API key
4. Configure:
   - Port: 27123 (default)
   - CORS: Enable for development
   - SSL: Disable for local development

### Tag Taxonomy

Edit `TAGS_SIMPLE.md` to customize valid tags:

```markdown
- authentication
- authorization
- rate-limiting
- caching
- # Add your custom tags here
```

### Scope Taxonomy

Edit `SCOPES_SIMPLE.md` to customize valid scopes:

```markdown
- auth
- resilience
- integration
- # Add your custom scopes here
```

## Testing

### Test Frontmatter Mapping

```powershell
cd Z:\10_WORKPLACE\Ti\apps\tibrain
uv run --with pyyaml --with watchdog python test_sync.py
```

### Test Go Components

```powershell
cd Z:\10_WORKPLACE\i\apps\tibrain
go test ./internal/frontmatter/
```

### Test MCP Server

```powershell
cd Z:\10_WORKPLACE\Ti\apps\tibrain\obsidian-mcp-server
npm run start:stdio
```

## Troubleshooting

### Build Errors

**Issue:** Bun binary remapping errors
**Solution:** Use npm instead of bun
```powershell
cd Z:\10_WORKPLACE\Ti\apps\tibrain\obsidian-mcp-server
rm -r node_modules bun.lockb
npm install
npm run build
```

### Authentication Failures

**Issue:** "No account logged in" error
**Solution:** Run interactive login
```powershell
cd Z:\10_WORKPLACE\Ti\apps\tibrain\obsidian-headless
node cli.js login
```

### Taxonomy Warnings

**Issue:** "Taxonomy file not found" or "Invalid tags"
**Solution:** Ensure TAGS_SIMPLE.md and SCOPES_SIMPLE.md exist in tibrain directory

### Sync Failures

**Issue:** Files not syncing
**Solution:** Check paths in .env file and verify vault directory exists

## Monitoring

### Log Files

- Sync script: Python logging output
- MCP server: Console output
- obsidian-headless: Console output

### Health Checks

```powershell
# Check if sync is running
Get-Process python

# Check if MCP server is running
Get-Process node

# Check recent sync activity
# (View console output or add logging to file)
```

## Security

### API Keys

- Never commit .env file to version control
- Use environment variables for production
- Rotate API keys regularly

### File Permissions

- Ensure vault directories have appropriate permissions
- Limit write access to sync service only

### Network

- Local REST API should not be exposed to internet
- Use firewall to restrict access to localhost

## Backup

### Vault Backup

```powershell
# Backup Obsidian vault
Copy-Item -Recurse Z:\10_WORKPLACE\Ti\apps\tibrain\vault Z:\backups\vault-backup-$(Get-Date -Format "yyyyMMdd")
```

### Ti Brain Backup

```powershell
# Backup Ti Brain knowledge base
Copy-Item -Recurse Z:\10_WORKPLACE\Ti\apps\tibrain\knowledge Z:\backups\tibrain-knowledge-$(Get-Date -Format "yyyyMMdd")
```

### Taxonomy Backup

```powershell
# Backup taxonomy files
Copy-Item Z:\10_WORKPLACE\Ti\apps\tibrain\TAGS_SIMPLE.md Z:\backups\
Copy-Item Z:\10_WORKPLACE\Ti\apps\tibrain\SCOPES_SIMPLE.md Z:\backups\
```

## Maintenance

### Update Components

```powershell
# Update obsidian-mcp-server
cd Z:\10_WORKPLACE\Ti\apps\tibrain\obsidian-mcp-server
git pull
npm install
npm run build

# Update obsidian-headless
cd Z:\10_WORKPLACE\Ti\apps\tibrain\obsidian-headless
git pull
npm install

# Update sync script
cd Z:\10_WORKPLACE\Ti\apps\tibrain
git pull
uv sync
```

### Taxonomy Maintenance

- Review tags quarterly
- Add new tags as needed
- Deprecate unused tags
- Update TAGS_SIMPLE.md and SCOPES_SIMPLE.md

## Performance

### Optimization

For large vaults (>10,000 notes):
- Increase OBSIDIAN_REQUEST_TIMEOUT_MS
- Use batch sync operations
- Consider incremental sync strategies

### Monitoring Metrics

- Sync success rate
- Sync latency
- Error rate
- API response time

## Support

For issues or questions:
- Deployment guide: `docs/OBSIDIAN_DEPLOYMENT_GUIDE.md`
- Architecture: `docs/OBSIDIAN_INTEGRATION_ARCHITECTURE.md`
- Implementation summary: `docs/OBSIDIAN_IMPLEMENTATION_SUMMARY.md`

## Production Checklist

Before deploying to production:

- [ ] All components built successfully
- [ ] Environment variables configured
- [ ] API keys set and tested
- [ ] Taxonomy files customized
- [ ] Obsidian Local REST API configured
- [ ] obsidian-headless authenticated
- [ ] Vault sync tested
- [ ] Backup strategy in place
- [ ] Monitoring configured
- [ ] Security review completed
- [ ] Documentation updated
- [ ] Team trained on operations
