# Obsidian Integration Setup Script (PowerShell)
# This script sets up the Obsidian ↔ Ti Brain integration for production use

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$TibrainDir = $ScriptDir
$VaultPath = if ($env:VAULT_PATH) { $env:VAULT_PATH } else { "$HOME/vaults/tibrain-vault" }

Write-Host "=== Obsidian Integration Setup ===" -ForegroundColor Cyan
Write-Host "Ti Brain Directory: $TibrainDir"
Write-Host "Vault Path: $VaultPath"
Write-Host ""

# Step 1: Verify dependencies
Write-Host "Step 1: Verifying dependencies..." -ForegroundColor Yellow

# Check Node.js
try {
    $nodeVersion = node -v
    $majorVersion = [int]($nodeVersion -replace 'v(\d+).*', '$1')
    if ($majorVersion -lt 22) {
        Write-Host "ERROR: Node.js version $nodeVersion is too old. Please install Node.js 22+" -ForegroundColor Red
        exit 1
    }
    Write-Host "✓ Node.js version: $nodeVersion" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Node.js not found. Please install Node.js 22+" -ForegroundColor Red
    exit 1
}

# Check Python
try {
    $pythonVersion = python --version
    Write-Host "✓ Python version: $pythonVersion" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Python not found. Please install Python 3.11+" -ForegroundColor Red
    exit 1
}

# Check uv
try {
    $uvVersion = uv --version
    Write-Host "✓ uv version: $uvVersion" -ForegroundColor Green
} catch {
    Write-Host "WARNING: uv not found. Please install from https://github.com/astral-sh/uv" -ForegroundColor Yellow
    # Continue anyway as uv might be optional
}

# Step 2: Build obsidian-mcp-server
Write-Host ""
Write-Host "Step 2: Building obsidian-mcp-server..." -ForegroundColor Yellow
Set-Location "$TibrainDir/obsidian-mcp-server"
if (-not (Test-Path "node_modules")) {
    Write-Host "Installing dependencies with npm..."
    npm install
}
if (-not (Test-Path "dist")) {
    Write-Host "Building with npm..."
    npm run build
}
Write-Host "✓ obsidian-mcp-server built successfully" -ForegroundColor Green

# Step 3: Setup obsidian-headless
Write-Host ""
Write-Host "Step 3: Setting up obsidian-headless..." -ForegroundColor Yellow
Set-Location "$TibrainDir/obsidian-headless"
if (-not (Test-Path "node_modules")) {
    Write-Host "Installing dependencies..."
    npm install
}
Write-Host "✓ obsidian-headless dependencies installed" -ForegroundColor Green

# Step 4: Create vault directory
Write-Host ""
Write-Host "Step 4: Creating vault directory..." -ForegroundColor Yellow
if (-not (Test-Path $VaultPath)) {
    New-Item -ItemType Directory -Path $VaultPath -Force
    Write-Host "✓ Created vault directory: $VaultPath" -ForegroundColor Green
} else {
    Write-Host "✓ Vault directory already exists: $VaultPath" -ForegroundColor Green
}

# Step 5: Setup environment variables
Write-Host ""
Write-Host "Step 5: Setting up environment variables..." -ForegroundColor Yellow
$envFile = "$TibrainDir/.env"
if (-not (Test-Path $envFile)) {
    @"
# Obsidian MCP Server Configuration
OBSIDIAN_API_KEY=your-api-key-here
OBSIDIAN_BASE_URL=http://127.0.0.1:27123
OBSIDIAN_VERIFY_SSL=false
OBSIDIAN_REQUEST_TIMEOUT_MS=30000
OBSIDIAN_ENABLE_COMMANDS=false
OBSIDIAN_READ_PATHS=
OBSIDIAN_WRITE_PATHS=
OBSIDIAN_READ_ONLY=false

# Sync Configuration
VAULT_PATH=$VaultPath
TIBRAIN_PATH=$TibrainDir
SYNC_PATTERN=*.md
SYNC_DIRECTION=obsidian-to-tibrain

# Ti Brain Configuration
TIBRAIN_DB_PATH=$TibrainDir/tibrain.db
TIBRAIN_KNOWLEDGE_PATH=$TibrainDir/knowledge
"@ | Out-File -FilePath $envFile -Encoding utf8
    Write-Host "✓ Created .env file" -ForegroundColor Green
} else {
    Write-Host "✓ .env file already exists" -ForegroundColor Green
}

# Step 6: Install Python dependencies
Write-Host ""
Write-Host "Step 6: Installing Python dependencies..." -ForegroundColor Yellow
Set-Location $TibrainDir
Write-Host "Installing PyYAML and watchdog..."
try {
    uv add pyyaml watchdog 2>$null
    Write-Host "✓ Python dependencies installed" -ForegroundColor Green
} catch {
    Write-Host "WARNING: Failed to install Python dependencies (may already be installed)" -ForegroundColor Yellow
}

# Step 7: Create startup script
Write-Host ""
Write-Host "Step 7: Creating startup script..." -ForegroundColor Yellow
$startSyncScript = "$TibrainDir/start_sync.ps1"
@'
# Start the Obsidian sync service

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

# Load environment variables
if (Test-Path ".env") {
    Get-Content .env | Where-Object { $_ -notmatch '^#' -and $_ -match '=' } | ForEach-Object {
        $key, $value = $_.split('=', 2)
        [Environment]::SetEnvironmentVariable($key.Trim(), $value.Trim())
    }
}

# Start sync
Write-Host "Starting Obsidian sync service..."
Write-Host "Vault: $env:VAULT_PATH"
Write-Host "Ti Brain: $env:TIBRAIN_PATH"
Write-Host ""

$syncDirection = if ($env:SYNC_DIRECTION) { $env:SYNC_DIRECTION } else { "obsidian-to-tibrain" }
$vaultPath = if ($env:VAULT_PATH) { $env:VAULT_PATH } else { "$HOME/vaults/tibrain-vault" }
$tibrainPath = if ($env:TIBRAIN_PATH) { $env:TIBRAIN_PATH } else { $ScriptDir }
$syncPattern = if ($env:SYNC_PATTERN) { $env:SYNC_PATTERN } else { "*.md" }

& uv run --with pyyaml --with watchdog python sync_obsidian.py `
    --direction $syncDirection `
    --vault $vaultPath `
    --tibrain $tibrainPath `
    --pattern $syncPattern `
    --watch
'@ | Out-File -FilePath $startSyncScript -Encoding utf8
Write-Host "✓ Created startup script: $startSyncScript" -ForegroundColor Green

# Step 8: Create MCP server startup script
Write-Host ""
Write-Host "Step 8: Creating MCP server startup script..." -ForegroundColor Yellow
$startMcpScript = "$TibrainDir/start_mcp_server.ps1"
@'
# Start the Obsidian MCP server

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location "$ScriptDir/obsidian-mcp-server"

# Load environment variables
$envFile = "../.env"
if (Test-Path $envFile) {
    Get-Content $envFile | Where-Object { $_ -notmatch '^#' -and $_ -match '=' } | ForEach-Object {
        $key, $value = $_.split('=', 2)
        [Environment]::SetEnvironmentVariable($key.Trim(), $value.Trim())
    }
}

# Start MCP server
Write-Host "Starting Obsidian MCP server..."
Write-Host "Base URL: $env:OBSIDIAN_BASE_URL"
Write-Host ""

npm run start:stdio
'@ | Out-File -FilePath $startMcpScript -Encoding utf8
Write-Host "✓ Created MCP server startup script: $startMcpScript" -ForegroundColor Green

# Final summary
Write-Host ""
Write-Host "=== Setup Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Configure Obsidian Local REST API plugin"
Write-Host "2. Set OBSIDIAN_API_KEY in .env file"
Write-Host "3. Run obsidian-headless authentication:"
Write-Host "   cd $TibrainDir/obsidian-headless; node cli.js login"
Write-Host "4. Setup vault sync:"
Write-Host "   cd $TibrainDir/obsidian-headless; node cli.js sync-setup -vault 'Your Vault Name' -path $VaultPath"
Write-Host "5. Start sync service:"
Write-Host "   $startSyncScript"
Write-Host "6. Start MCP server (optional):"
Write-Host "   $startMcpScript"
Write-Host ""
Write-Host "For more information, see: $TibrainDir/docs/OBSIDIAN_DEPLOYMENT_GUIDE.md" -ForegroundColor Cyan
