# JAVA_TOOL_OPTIONS Usage Guide

This guide explains how to properly use the memory calculator's output to set `JAVA_TOOL_OPTIONS` for your Java applications.

## Understanding the Problem

The memory calculator is designed to **calculate and output** JVM memory options, but it cannot directly modify your shell's environment variables due to Unix process isolation. This is a security feature - child processes cannot modify their parent's environment.

## Solutions Overview

| Method | Use Case | Complexity | Recommended For |
|--------|----------|------------|-----------------|
| Command Substitution | Interactive shell, scripts | Low | Quick testing, automation |
| Helper Script | Interactive development | Low | Development workflow |
| Docker Integration | Containerized apps | Medium | Production containers |
| Init Containers | Kubernetes | High | Cloud-native deployments |

## Method 1: Command Substitution

### Basic Usage
```bash
# Calculate and set in one command
export JAVA_TOOL_OPTIONS="$(./memory-calculator --total-memory=2G --quiet)"

# With custom application path for better class count estimation
export JAVA_TOOL_OPTIONS="$(./memory-calculator --total-memory=2G --path=/my/app --quiet)"

# Verify it's set
echo $JAVA_TOOL_OPTIONS
# Output: -XX:MaxDirectMemorySize=10M -Xmx494583K -XX:MaxMetaspaceSize=41992K -XX:ReservedCodeCacheSize=240M -Xss1M
```

### With Error Handling
```bash
#!/bin/bash
set -e  # Exit on error

# Calculate JVM options
JVM_OPTS="$(./memory-calculator --total-memory=2G --quiet)"

if [ -z "$JVM_OPTS" ]; then
    echo "Error: Failed to calculate JVM options"
    exit 1
fi

export JAVA_TOOL_OPTIONS="$JVM_OPTS"
echo "Set JVM options: $JAVA_TOOL_OPTIONS"

# Start your application
java -jar myapp.jar
```

### Advanced Script with Fallback
```bash
#!/bin/bash
# startup.sh - Production startup script with fallback

# Try to calculate optimal JVM settings
if command -v ./memory-calculator >/dev/null 2>&1; then
    echo "Calculating optimal JVM settings..."
    JVM_OPTS="$(./memory-calculator --quiet 2>/dev/null)"
    
    if [ -n "$JVM_OPTS" ]; then
        export JAVA_TOOL_OPTIONS="$JVM_OPTS"
        echo "✅ Using calculated JVM options: $JAVA_TOOL_OPTIONS"
    else
        echo "⚠️  Memory calculation failed, using defaults"
    fi
else
    echo "⚠️  Memory calculator not found, using default JVM settings"
fi

# Start application
exec java -jar "$@"
```

## Method 2: Helper Script

### Using the Provided Script
```bash
# Source the helper script with your options
source ./set-java-options.sh --total-memory=2G --thread-count=300

# The script automatically sets JAVA_TOOL_OPTIONS and provides feedback
# ✅ Set JAVA_TOOL_OPTIONS=-XX:MaxDirectMemorySize=10M -Xmx494583K...
```

### Creating Your Own Helper
```bash
#!/bin/bash
# my-java-setup.sh
# Custom helper script for your specific needs

set_java_memory() {
    local memory_size="${1:-1G}"  # Default to 1G if not specified
    local thread_count="${2:-250}" # Default thread count
    
    echo "Setting up JVM for ${memory_size} memory with ${thread_count} threads..."
    
    local opts="$(./memory-calculator --total-memory="$memory_size" --thread-count="$thread_count" --quiet)"
    
    if [ -n "$opts" ]; then
        export JAVA_TOOL_OPTIONS="$opts"
        echo "✅ JVM configured successfully"
        return 0
    else
        echo "❌ Failed to configure JVM"
        return 1
    fi
}

# Usage: source my-java-setup.sh && set_java_memory 2G 300
```

## Method 3: Docker Integration

### Single-Stage Dockerfile
```dockerfile
FROM bellsoft/liberica-runtime-container:jdk-21-slim-musl

# Copy your application
COPY myapp.jar /app/myapp.jar
COPY memory-calculator /usr/local/bin/

# Create startup script
RUN echo '#!/bin/bash\n\
export JAVA_TOOL_OPTIONS="$(memory-calculator --quiet)"\n\
echo "Using JVM options: $JAVA_TOOL_OPTIONS"\n\
exec java -jar /app/myapp.jar "$@"' > /startup.sh && \
    chmod +x /startup.sh

ENTRYPOINT ["/startup.sh"]
```

### Multi-Stage Build
```dockerfile
# Build stage - compile memory calculator
FROM golang:1.24.5 as calculator-builder
WORKDIR /build
COPY . .
RUN make build-minimal

# Runtime stage
FROM bellsoft/liberica-runtime-container:jdk-21-slim-musl
WORKDIR /app

# Copy calculator and application
COPY --from=calculator-builder /build/memory-calculator /usr/local/bin/
COPY myapp.jar .

# Smart entrypoint script
COPY <<EOF /entrypoint.sh
#!/bin/bash
set -e

# Calculate JVM options based on container resources
export JAVA_TOOL_OPTIONS="\$(memory-calculator --quiet)"
echo "🚀 Starting with JVM options: \$JAVA_TOOL_OPTIONS"

# Execute the main command
exec "\$@"
EOF

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["java", "-jar", "myapp.jar"]
```

### Docker Compose
```yaml
version: '3.8'
services:
  java-app:
    build: .
    environment:
      # These will be picked up by the memory calculator
      - BPL_JVM_THREAD_COUNT=300
      - BPL_JVM_HEAD_ROOM=10
    deploy:
      resources:
        limits:
          memory: 2G
        reservations:
          memory: 1G
    command: >
      sh -c "
        export JAVA_TOOL_OPTIONS=\"$$(memory-calculator --quiet)\" &&
        echo \"Using JVM options: $$JAVA_TOOL_OPTIONS\" &&
        java -jar myapp.jar
      "
```

## Method 4: Kubernetes Integration

### Init Container Pattern
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: java-app
spec:
  template:
    spec:
      initContainers:
      - name: jvm-calculator
        image: myapp:latest
        command: ["/bin/sh"]
        args:
        - -c
        - |
          echo "Calculating JVM options for container..."
          memory-calculator --quiet > /shared/java-opts
          echo "JVM options written to shared volume"
        volumeMounts:
        - name: jvm-config
          mountPath: /shared
        resources:
          requests:
            memory: "64Mi"
            cpu: "10m"
          limits:
            memory: "128Mi"
            cpu: "50m"
      
      containers:
      - name: app
        image: myapp:latest
        command: ["/bin/sh"]
        args:
        - -c
        - |
          export JAVA_TOOL_OPTIONS="$(cat /shared/java-opts)"
          echo "Starting with JVM options: $JAVA_TOOL_OPTIONS"
          exec java -jar app.jar
        volumeMounts:
        - name: jvm-config
          mountPath: /shared
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
      
      volumes:
      - name: jvm-config
        emptyDir: {}
```

### Sidecar Pattern
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: java-app
spec:
  template:
    spec:
      containers:
      - name: jvm-optimizer
        image: myapp:latest
        command: ["/bin/sh"]
        args:
        - -c
        - |
          while true; do
            memory-calculator --quiet > /shared/java-opts
            sleep 300  # Recalculate every 5 minutes
          done
        volumeMounts:
        - name: jvm-config
          mountPath: /shared
        resources:
          requests:
            memory: "32Mi"
            cpu: "5m"
          limits:
            memory: "64Mi"
            cpu: "10m"
      
      - name: app
        image: myapp:latest
        command: ["/bin/sh"]
        args:
        - -c
        - |
          # Initial setup
          if [ -f /shared/java-opts ]; then
            export JAVA_TOOL_OPTIONS="$(cat /shared/java-opts)"
          fi
          
          # Start app with periodic reloading
          java -jar app.jar &
          PID=$!
          
          while kill -0 $PID 2>/dev/null; do
            if [ -f /shared/java-opts ]; then
              NEW_OPTS="$(cat /shared/java-opts)"
              if [ "$NEW_OPTS" != "$JAVA_TOOL_OPTIONS" ]; then
                echo "JVM options changed, restart required"
                # In production, you might trigger a graceful restart here
              fi
            fi
            sleep 60
          done
        volumeMounts:
        - name: jvm-config
          mountPath: /shared
      
      volumes:
      - name: jvm-config
        emptyDir: {}
```

## Method 5: Cloud-Native Integration

### Paketo Buildpacks
```bash
# The memory calculator integrates seamlessly with Paketo buildpacks
# Just set environment variables:

pack build myapp --env BPL_JVM_THREAD_COUNT=300 --env BPL_JVM_HEAD_ROOM=15

# Or in project.toml:
cat > project.toml << EOF
[build]
env = [
  "BPL_JVM_THREAD_COUNT=300",
  "BPL_JVM_HEAD_ROOM=15"
]
EOF
```

### Cloud Foundry
```yaml
# manifest.yml
applications:
- name: java-app
  memory: 2G
  env:
    BPL_JVM_THREAD_COUNT: 300
    BPL_JVM_HEAD_ROOM: 10
  command: |
    export JAVA_TOOL_OPTIONS="$(memory-calculator --quiet)" &&
    java -jar app.jar
```

## Troubleshooting

### Common Issues

**Issue**: `JAVA_TOOL_OPTIONS` is empty
```bash
# Debug: Check what the calculator outputs
./memory-calculator --total-memory=1G --loaded-class-count=5000

# If it fails, provide explicit values:
./memory-calculator --total-memory=1G --loaded-class-count=5000 --thread-count=250
```

**Issue**: "Unable to determine class count"
```bash
# Solution 1: Set explicit application path for automatic scanning
./memory-calculator --path=/path/to/your/app --total-memory=2G

# Solution 2: Set explicit class count
./memory-calculator --loaded-class-count=10000 --total-memory=2G

# Solution 3: Use environment variable
export BPI_APPLICATION_PATH=/path/to/your/app
./memory-calculator --total-memory=2G
```

**Issue**: Script fails in containers
```bash
# Ensure the binary is executable and present
ls -la memory-calculator
chmod +x memory-calculator

# Check container limits are detected
./memory-calculator  # Should show detected memory
```

### Debugging Commands

```bash
# Test memory detection
./memory-calculator --total-memory=2G

# Test quiet mode
./memory-calculator --total-memory=2G --quiet

# Test with custom application path
./memory-calculator --total-memory=2G --path=/my/app

# Test with explicit values
./memory-calculator --total-memory=2G --loaded-class-count=5000 --thread-count=300

# Check current environment
echo "Current JAVA_TOOL_OPTIONS: $JAVA_TOOL_OPTIONS"
env | grep JAVA_TOOL_OPTIONS
```

## Enhanced Features

### Application Path Scanning

The `--path` parameter enables intelligent class count estimation by scanning JAR files in your application directory:

```bash
# Automatic class count estimation from application path
./memory-calculator --path=/opt/myapp --total-memory=4G

# The calculator will:
# 1. Scan all JAR files in /opt/myapp recursively
# 2. Count classes in each JAR
# 3. Apply framework-specific scaling factors (Spring Boot, etc.)
# 4. Display "Loaded Classes: auto-calculated from /opt/myapp"
```

**Benefits:**
- More accurate metaspace allocation
- Framework-aware class counting
- Eliminates need to manually specify class counts
- Clear display of calculation source

### Improved Display Output

The calculator now provides clearer information about calculated values:

**When class count is auto-calculated:**
```
Loaded Classes:   auto-calculated from /opt/myapp
```

**When class count is manually specified:**
```
Loaded Classes:   50000
```

This makes it clear whether values were calculated automatically or provided manually.

## Best Practices

1. **Always use `--quiet` for automation**: Provides clean output suitable for `export`
2. **Handle errors gracefully**: Check exit codes and output validity
3. **Provide fallback values**: Don't depend solely on auto-detection
4. **Test in your target environment**: Container limits may differ from host
5. **Monitor JVM performance**: Adjust parameters based on actual usage
6. **Use init containers in Kubernetes**: Separates concerns and improves reliability
7. **Cache calculated values**: Avoid recalculating on every container restart

## Integration Examples

See the `examples/` directory for complete working examples of each integration method.
