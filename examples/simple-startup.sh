#!/bin/bash
# examples/simple-startup.sh
# Simple startup script that sets JVM options and runs a Java application

set -e  # Exit on any error

echo "🚀 Starting Java application with optimized memory settings..."

# Check if memory calculator is available
if [ ! -x "./memory-calculator" ]; then
    echo "❌ memory-calculator not found or not executable"
    echo "Please build it first: make build"
    exit 1
fi

# Calculate JVM options (fall back to defaults on failure)
echo "🧮 Calculating optimal JVM memory settings..."
JVM_OPTS="$(./memory-calculator --quiet 2>/dev/null)" || {
    echo "⚠️  Auto-calculation failed, using conservative defaults"
    JVM_OPTS="-Xmx512m -XX:MaxMetaspaceSize=128m -Xss1m"
}

# Set the environment variable
export JAVA_TOOL_OPTIONS="$JVM_OPTS"
echo "✅ JVM Options: $JAVA_TOOL_OPTIONS"

# Example: Start your Java application
# Replace this with your actual application
echo "📦 Starting application..."
echo "java -jar myapp.jar"

# Uncomment to actually run your app:
# exec java -jar myapp.jar "$@"
