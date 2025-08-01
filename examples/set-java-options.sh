#!/bin/bash
# Helper script to set JAVA_TOOL_OPTIONS from memory calculator
# Usage: source set-java-options.sh [calculator-options]
# Example: source set-java-options.sh --total-memory=2G --thread-count=300

# Check if script is being sourced (not executed)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "❌ This script must be sourced, not executed directly."
    echo "Usage: source $0 [options]"
    echo "Example: source $0 --total-memory=2G"
    exit 1
fi

# Get the directory where this script is located (examples directory)
# Handle both sourcing and direct execution
if [[ -n "${BASH_SOURCE[0]}" ]]; then
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
else
    # When sourced, BASH_SOURCE[0] might be empty, so check if we're in examples dir
    if [[ "$(basename "$PWD")" == "examples" ]]; then
        SCRIPT_DIR="$PWD"
    elif [[ -d "examples" ]]; then
        SCRIPT_DIR="$PWD/examples"
    else
        echo "❌ Cannot determine script location. Please run from project root or examples directory."
        return 1
    fi
fi

# Memory calculator is in the parent directory
CALCULATOR="$(dirname "$SCRIPT_DIR")/memory-calculator"

# Check if memory calculator exists
if [ ! -f "$CALCULATOR" ]; then
    echo "❌ Memory calculator not found at: $CALCULATOR"
    echo "Please build the project first: make build"
    return 1
fi

# Check if memory calculator is executable
if [ ! -x "$CALCULATOR" ]; then
    echo "❌ Memory calculator is not executable: $CALCULATOR"
    echo "Run: chmod +x $CALCULATOR"
    return 1
fi

# Run the memory calculator with provided arguments
echo "🧮 Calculating JVM options..."
JAVA_OPTIONS=$("$CALCULATOR" --quiet "$@" 2>/dev/null)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ] && [ -n "$JAVA_OPTIONS" ]; then
    export JAVA_TOOL_OPTIONS="$JAVA_OPTIONS"
    echo "✅ Set JAVA_TOOL_OPTIONS=$JAVA_TOOL_OPTIONS"
    echo ""
    echo "🚀 You can now run your Java application with optimized memory settings!"
    echo "   Example: java -jar myapp.jar"
    echo ""
else
    echo "❌ Failed to calculate JVM options (exit code: $EXIT_CODE)"
    if [ -n "$1" ]; then
        echo "   Arguments used: $@"
        echo "   Try: $CALCULATOR --help"
    else
        echo "   Hint: You may need to specify --total-memory or --loaded-class-count"
        echo "   Example: source $0 --total-memory=1G --loaded-class-count=5000"
    fi
    return 1
fi
