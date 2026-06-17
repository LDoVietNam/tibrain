#!/bin/bash

# Ti Brain Indexing Monitor Script
# Purpose: Real-time monitoring of indexing operations

echo "=== 🚀 Ti Brain Indexing Monitor ==="
echo "Started at: $(date)"
echo ""

# Check if indexing is running
check_indexing_status() {
    echo "🔍 Checking indexing status..."
    
    # Check for active indexing processes
    if pgrep -f "kb-builder.exe" > /dev/null; then
        echo "✅ Indexing is running"
        PID=$(pgrep -f "kb-builder.exe")
        echo "📊 Process ID: $PID"
        
        # Get process details
        ps -p $PID -o pid,ppid,cmd,%mem,%cpu,etime --no-headers
    else
        echo "❌ No indexing process found"
        return 1
    fi
}

# Monitor system resources
monitor_system_resources() {
    echo ""
    echo "💻 System Resources:"
    
    # CPU usage
    echo "🔥 CPU Usage: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)%"
    
    # Memory usage
    MEMORY_INFO=$(free -h | grep "Mem:")
    TOTAL_MEM=$(echo $MEMORY_INFO | awk '{print $2}')
    USED_MEM=$(echo $MEMORY_INFO | awk '{print $3}')
    USED_PERCENT=$(free | grep "Mem:" | awk '{printf "%.1f", $3/$2 * 100.0}')
    echo "💾 Memory: $USED_MEM/$TOTAL_MEM ($USED_PERCENT%)"
    
    # Disk usage
    DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}')
    echo "💿 Disk Usage: $DISK_USAGE"
    
    # Load average
    LOAD_AVG=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | tr -d ',')
    echo "⚡ Load Average: $LOAD_AVG"
}

# Monitor indexing progress
monitor_indexing_progress() {
    echo ""
    echo "📈 Indexing Progress:"
    
    # Check knowledge base size
    if [ -d "knowledge/ti-ecosystem-docs" ]; then
        FILE_COUNT=$(find knowledge/ti-ecosystem-docs -type f | wc -l)
        echo "📄 Files indexed: $FILE_COUNT"
        
        # Calculate directory size
        DIR_SIZE=$(du -sh knowledge/ti-ecosystem-docs | cut -f1)
        echo "💾 Knowledge base size: $DIR_SIZE"
    fi
    
    # Check database size
    if [ -f "memory/hub.db" ]; then
        DB_SIZE=$(du -sh memory/hub.db | cut -f1)
        echo "🗄️ Database size: $DB_SIZE"
    fi
}

# Monitor performance metrics
monitor_performance() {
    echo ""
    echo "⚡ Performance Metrics:"
    
    # Check recent log entries for performance
    if [ -f "logs/tibrain.log" ]; then
        echo "📊 Recent performance logs:"
        tail -10 logs/tibrain.log | grep -E "(performance|speed|time|ms)" | tail -5
    fi
}

# Generate alerts
check_alerts() {
    echo ""
    echo "🚨 Alert Check:"
    
    # Memory alert
    MEMORY_USAGE=$(free | grep "Mem:" | awk '{printf "%.0f", $3/$2 * 100.0}')
    if [ "$MEMORY_USAGE" -gt 85 ]; then
        echo "⚠️  High memory usage: $MEMORY_USAGE%"
    fi
    
    # CPU alert
    CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1 | cut -d'.' -f1)
    if [ "$CPU_USAGE" -gt 90 ]; then
        echo "⚠️  High CPU usage: $CPU_USAGE%"
    fi
    
    # Disk space alert
    DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | cut -d'%' -f1)
    if [ "$DISK_USAGE" -gt 90 ]; then
        echo "⚠️  Low disk space: $DISK_USAGE%"
    fi
}

# Main monitoring loop
main_monitor() {
    while true; do
        clear
        echo "=== 🚀 Ti Brain Indexing Monitor ==="
        echo "Last updated: $(date)"
        echo ""
        
        check_indexing_status
        if [ $? -eq 0 ]; then
            monitor_system_resources
            monitor_indexing_progress
            monitor_performance
            check_alerts
        fi
        
        echo ""
        echo "🔄 Refreshing in 30 seconds... (Press Ctrl+C to stop)"
        sleep 30
    done
}

# Run monitoring
if [ "$1" == "--once" ]; then
    check_indexing_status
    monitor_system_resources
    monitor_indexing_progress
    monitor_performance
    check_alerts
else
    main_monitor
fi