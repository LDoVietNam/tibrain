#!/bin/bash
# Sync .devin memory with Ti Brain
# Hybrid approach: local + centralized

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DEVIN_PATH="$PROJECT_ROOT/.devin"
BRAIN_PATH="/z/Ti/brain"
PROJECT="ti-learning-lab"
TIMESTAMP=$(date +%Y-%m-%d_%H-%M-%S)

echo "=========================================="
echo "Ti-Learning-Lab → Ti Brain Sync"
echo "=========================================="
echo "Timestamp: $TIMESTAMP"
echo ""

# Function to push local knowledge to Ti Brain
push_to_brain() {
    echo "📤 Pushing local knowledge to Ti Brain..."
    
    # Create project folder in Ti Brain if not exists
    mkdir -p "$BRAIN_PATH/memory/lessons/$PROJECT"
    
    # Sync knowledge folder
    if [ -d "$DEVIN_PATH/knowledge" ]; then
        rsync -av --delete "$DEVIN_PATH/knowledge/" "$BRAIN_PATH/memory/lessons/$PROJECT/"
        echo "✅ Knowledge synced to Ti Brain"
    else
        echo "⚠️  Knowledge folder not found, skipping"
    fi
}

# Function to pull system-wide patterns from Ti Brain
pull_from_brain() {
    echo "📥 Pulling system-wide patterns from Ti Brain..."
    
    # Create knowledge folder if not exists
    mkdir -p "$DEVIN_PATH/knowledge"
    
    # Pull patterns (read-only reference)
    if [ -d "$BRAIN_PATH/memory/lessons" ]; then
        rsync -av "$BRAIN_PATH/memory/lessons/" "$DEVIN_PATH/knowledge/"
        echo "✅ System-wide patterns pulled from Ti Brain"
    else
        echo "⚠️  Ti Brain lessons folder not found, skipping"
    fi
}

# Function to bi-directional sync
full_sync() {
    echo "🔄 Performing full sync (bi-directional)..."
    
    # Push local changes first
    push_to_brain
    
    # Then pull system-wide patterns
    pull_from_brain
    
    echo "✅ Full sync completed"
}

# Function to show sync status
sync_status() {
    echo "📊 Sync Status:"
    echo "  Local .devin: $DEVIN_PATH"
    echo "  Ti Brain: $BRAIN_PATH"
    echo "  Project: $PROJECT"
    echo ""
    
    if [ -d "$DEVIN_PATH" ]; then
        echo "  Local files:"
        find "$DEVIN_PATH" -type f | wc -l
    else
        echo "  ⚠️  .devin folder not found"
    fi
}

# Main logic
case "${1:-}" in
    push)
        push_to_brain
        ;;
    pull)
        pull_from_brain
        ;;
    full)
        full_sync
        ;;
    status)
        sync_status
        ;;
    *)
        echo "Usage: $0 {push|pull|full|status}"
        echo ""
        echo "Commands:"
        echo "  push    - Push local knowledge to Ti Brain"
        echo "  pull    - Pull system-wide patterns from Ti Brain"
        echo "  full    - Bi-directional sync"
        echo "  status  - Show sync status"
        exit 1
        ;;
esac

echo ""
echo "=========================================="
echo "Sync completed at $TIMESTAMP"
echo "=========================================="
