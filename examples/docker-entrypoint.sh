#!/bin/bash
# examples/docker-entrypoint.sh  
# Production-ready Docker entrypoint script

set -e

# Configuration
MEMORY_CALCULATOR="/usr/local/bin/memory-calculator"
DEFAULT_JVM_OPTS="-Xmx512m -XX:MaxMetaspaceSize=128m -Xss1m -XX:ReservedCodeCacheSize=240m"

echo "🐳 Docker container starting..."

# Function to calculate JVM options
calculate_jvm_options() {
    if [ -x "$MEMORY_CALCULATOR" ]; then
        echo "🧮 Calculating JVM options based on container resources..."
        
        # Try to calculate optimal settings
        local calculated_opts
        if calculated_opts="$($MEMORY_CALCULATOR --quiet 2>/dev/null)" && [ -n "$calculated_opts" ]; then
            echo "✅ Using calculated JVM options"
            echo "$calculated_opts"
            return 0
        else
            echo "⚠️  Memory calculation failed, checking container limits..."
            
            # Try with explicit memory limit detection
            local mem_limit
            if mem_limit=$(cat /sys/fs/cgroup/memory.max 2>/dev/null) && [ "$mem_limit" != "max" ]; then
                local mem_gb=$((mem_limit / 1024 / 1024 / 1024))
                echo "📊 Detected ${mem_gb}GB container limit"
                
                calculated_opts="$($MEMORY_CALCULATOR --total-memory=${mem_gb}G --quiet 2>/dev/null)" || true
                if [ -n "$calculated_opts" ]; then
                    echo "✅ Using calculated options for ${mem_gb}GB"
                    echo "$calculated_opts"
                    return 0
                fi
            fi
        fi
    fi
    
    echo "⚠️  Using fallback JVM options"
    echo "$DEFAULT_JVM_OPTS"
}

# Calculate and set JVM options
JVM_OPTS="$(calculate_jvm_options)"
export JAVA_TOOL_OPTIONS="$JVM_OPTS"

echo "🎯 Final JVM Options: $JAVA_TOOL_OPTIONS"
echo ""

# Health check function
health_check() {
    echo "🏥 Container health check:"
    echo "  Memory limit: $(cat /sys/fs/cgroup/memory.max 2>/dev/null || echo 'not detected')"
    echo "  JVM options: $JAVA_TOOL_OPTIONS"
    echo "  Java version: $(java -version 2>&1 | head -1)"
}

# Handle special commands
case "$1" in
    "health")
        health_check
        exit 0
        ;;
    "java-opts")
        echo "$JAVA_TOOL_OPTIONS"
        exit 0
        ;;
esac

# Start the main application
echo "🚀 Starting application: $@"
exec "$@"
