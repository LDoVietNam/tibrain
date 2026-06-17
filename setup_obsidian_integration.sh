#!/bin/bash

# Obsidian Integration Setup Script
# This script sets up the Obsidian ↔ Ti Brain integration for production use

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TIBRAIN_DIR="$SCRIPT_DIR"
VANLUT_PATH="${VAULT_PATH:-$HOME/vaults/tibrain-vault}"

echo "=== Obsidian Integration Setup ==="
echo "Ti Brain Directory: $TIBRAIN_DIR"
echo "Vault Path: $VANLUT_PATH"
echo ""

# Step 1: Verify dependencies
echo "Step 1: Verifying dependencies..."

# Check Node.js
if ! command -v node &> /dev/null; then
    echo "ERROR: Node.js not found. Please install Node.js 22+"
    exit 1
fi

NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 22 ]; then
    echo "ERROR: Node.js version $NODE_VERSION is too old. Please install Node.js 22+"
    exit 1
fi

echo "✓ Node.js version: $(node -v)"

# Check Python
if ! command -v python3 &> /dev/null; then
    echo "ERROR: Python3 not found. Please install Python 3.11+"
    exit 1
fi

echo "✓ Python version: $(python3 --version)"

# Check uv
if ! command -v uv &> /dev/null; then
    echo "WARNING: uv not found. Installing..."
    curl -LsSf https://astral.sh/uv/install.sh | sh
fi

echo "✓ uv version: $(uv --version)"

# Step 2: Build obsidian-mcp-server
echo ""
echo "Step 2: Building obsidian-mcp-server..."
cd "$TIBRAIN_DIR/obsidian-mcp-server"
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies with npm..."
    npm install
fi
if [ ! -d "dist" ]; then
    echo "Building with npm..."
    npm run build
fi
echo "✓ obsidian-mcp-server built successfully"

# Step 3: Setup obsidian-headless
echo ""
echo "Step 3: Setting up obsidian-headless..."
cd "$TIBRAIN_DIR/obsidian-headless"
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    npm install
fi
echo "✓ obsidian-headless dependencies installed"

# Step 4: Create vault directory
echo ""
echo "Step 4: Creating vault directory..."
if [ ! -d "$VANLUT_PATH" ]; then
    mkdir -p "$VANLUT_PATH"
    echo "✓ Created vault directory: $VANLUT_PATH"
else
    echo "✓ Vault directory already exists: $VANLUT_PATH"
fi

# Step 5: Setup environment variables
echo ""
echo "Step 5: Setting up environment variables..."

if [ ! -f "$TIBRAIN_DIR/.env" ]; then
    cat > "$TIBRAIN_DIR/.env" << EOF
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
VAULT_PATH=$VANLUT_PATH
TIBRAIN_PATH=$TIBRAIN_DIR
SYNC_PATTERN=*.md
SYNC_DIRECTION=obsidian-to-tibrain

# Ti Brain Configuration
TIBRAIN_DB_PATH=$TIBRAIN_DIR/tibrain.db
TIBRAIN_KNOWLEDGE_PATH=$TIBRAIN_DIR/knowledge
EOF
    echo "✓ Created .env file"
else
    echo "✓ .env file already exists"
fi

# Step 6: Install Python dependencies
echo ""
echo "Step 6: Installing Python dependencies..."
cd "$TIBRAIN_DIR"
echo "Installing PyYAML and watchdog..."
uv add pyyaml watchdog 2>/dev/null || echo "Dependencies may already be installed"
echo "✓ Python dependencies installed"

# Step 7: Create startup script
echo ""
echo "Step 7: Creating startup script..."

cat > "$TIBRAIN_DIR/start_sync.sh" << 'EOF'
#!/bin/bash
# Start the Obsidian sync service

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Load environment variables
if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Start sync
echo "Starting Obsidian sync service..."
echo "Vault: $VAULT_PATH"
echo "Ti Brain: $TIBRAIN_PATH"
echo ""

uv run --with pyyaml --with watchdog python sync_obsidian.py \
    --direction "$SYNC_DIRECTION" \
    --vault "$VAULT_PATH" \
    --tibrain "$TIBRAIN_PATH" \
    --pattern "$SYNC_PATTERN" \
    --watch
EOF

chmod +x "$TIBRAIN_DIR/start_sync.sh"
echo "✓ Created startup script: $TIBRAIN_DIR/start_sync.sh"

# Step 8: Create MCP server startup script
echo ""
echo "Step 8: Creating MCP server startup script..."

cat > "$TIBRAIN_DIR/start_mcp_server.sh" << 'EOF'
#!/bin/bash
# Start the Obsidian MCP server

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/obsidian-mcp-server"

# Load environment variables
if [ -f "../.env" ]; then
    export $(cat ../.env | grep -v '^#' | xargs)
fi

# Start MCP server
echo "Starting Obsidian MCP server..."
echo "Base URL: $OBSIDIAN_BASE_URL"
echo ""

npm run start:stdio
EOF

chmod +x "$TIBRAIN_DIR/start_mcp_server.sh"
echo "✓ Created MCP server startup script: $TIBRAIN_DIR/start_mcp_server.sh"

# Final summary
echo ""
echo "=== Setup Complete ==="
echo ""
echo "Next steps:"
echo "1. Configure Obsidian Local REST API plugin"
echo "2. Set OBSIDIAN_API_KEY in .env file"
echo "3. Run obsidian-headless authentication:"
echo "   cd $TIBRAIN_DIR/obsidian-headless && node cli.js login"
echo "4. Setup vault sync:"
echo "   cd $TIBRAIN_DIR/obsidian-headless && node cli.js sync-setup --vault 'Your Vault Name' --path $VANLUT_PATH"
echo "5. Start sync service:"
echo "   $TIBRAIN_DIR/start_sync.sh"
echo "6. Start MCP server (optional):"
echo "   $TIBRAIN_DIR/start_mcp_server.sh"
echo ""
echo "For more information, see: $TIBRAIN_DIR/docs/OBSIDIAN_DEPLOYMENT_GUIDE.md"
