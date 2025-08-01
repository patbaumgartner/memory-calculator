# Examples Directory

This directory contains practical examples of how to integrate the JVM Memory Calculator with different deployment scenarios.

## Files Overview

| File | Description | Use Case |
|------|-------------|----------|
| `set-java-options.sh` | Interactive helper script | Development, testing |
| `simple-startup.sh` | Basic startup script | Development, simple deployments |
| `docker-entrypoint.sh` | Production Docker entrypoint | Container deployments |
| `Dockerfile` | Complete Docker setup | Building container images |
| `kubernetes.yaml` | K8s deployment with init containers | Cloud-native deployments |

## Quick Start

### 1. Interactive Helper Script (Recommended for Development)
```bash
# Source the helper script to set JAVA_TOOL_OPTIONS in your current shell
source examples/set-java-options.sh --total-memory=2G --thread-count=300

# Now your Java applications will use the optimized settings
java -jar myapp.jar
```

### 2. Simple Script Usage
```bash
# Make the script executable
chmod +x examples/simple-startup.sh

# Run it (will calculate JVM options but not start an actual app)
./examples/simple-startup.sh
```

### 3. Docker Usage
```bash
# Build the example Docker image
docker build -f examples/Dockerfile -t memory-calc-example .

# Run with memory limit
docker run --memory=2g memory-calc-example

# Check calculated JVM options
docker run --memory=1g memory-calc-example java-opts
```

### 4. Kubernetes Usage
```bash
# Apply the complete deployment
kubectl apply -f examples/kubernetes.yaml

# Check the logs to see JVM options calculation
kubectl logs -l app=java-app -c jvm-calculator

# Check the application logs
kubectl logs -l app=java-app -c app
```

## Customization Guide

### Modifying for Your Application

1. **Replace placeholders** in the files:
   - `myapp.jar` → your actual JAR file
   - `myapp:latest` → your Docker image
   - Port `8080` → your application port
   - Health check paths → your actual endpoints

2. **Adjust memory parameters**:
   - Change default memory limits in scripts
   - Modify thread counts based on your application
   - Adjust head room percentage if needed

3. **Environment-specific settings**:
   - Update resource requests/limits in Kubernetes
   - Modify Docker memory constraints
   - Set appropriate health check intervals

### Common Patterns

#### Pattern 1: Environment Variable Override
```bash
# Allow environment variables to override defaults
MEMORY_SIZE="${MEMORY_SIZE:-1G}"
THREAD_COUNT="${THREAD_COUNT:-250}"
export JAVA_TOOL_OPTIONS="$(./memory-calculator --total-memory=$MEMORY_SIZE --thread-count=$THREAD_COUNT --quiet)"
```

#### Pattern 2: Configuration File Support
```bash
# Load settings from config file if it exists
if [ -f /etc/jvm-config ]; then
    source /etc/jvm-config
fi
```

#### Pattern 3: Graceful Degradation
```bash
# Fall back gracefully if calculator fails
JVM_OPTS="$(./memory-calculator --quiet 2>/dev/null)" || JVM_OPTS="-Xmx512m"
```

## Testing the Examples

### Prerequisites
```bash
# Build the memory calculator first
make build

# Verify it works
./memory-calculator --total-memory=1G --loaded-class-count=5000 --quiet
```

### Test Simple Script
```bash
chmod +x examples/simple-startup.sh
./examples/simple-startup.sh
```

Expected output:
```
🚀 Starting Java application with optimized memory settings...
🧮 Calculating optimal JVM memory settings...
✅ JVM Options: -XX:MaxDirectMemorySize=10M -Xmx494583K -XX:MaxMetaspaceSize=41992K -XX:ReservedCodeCacheSize=240M -Xss1M
📦 Starting application...
java -jar myapp.jar
```

### Test Docker Entrypoint
```bash
chmod +x examples/docker-entrypoint.sh

# Test health check
./examples/docker-entrypoint.sh health

# Test JVM options display
./examples/docker-entrypoint.sh java-opts
```

## Production Considerations

### Security
- All examples use non-root users where applicable
- Read-only root filesystems in Kubernetes
- Proper security contexts and resource limits

### Monitoring
- Health checks are included in all examples
- Kubernetes examples include HPA configuration
- Docker examples support health check commands

### Error Handling
- Graceful fallbacks when memory calculation fails
- Proper exit codes and error messages
- Logging for troubleshooting

### Performance
- Init containers minimize startup overhead
- Shared volumes for efficient data transfer
- Resource limits prevent resource exhaustion

## Troubleshooting

### Common Issues

**Issue**: "memory-calculator not found"
```bash
# Solution: Build the calculator first
make build
ls -la memory-calculator  # Should exist and be executable
```

**Issue**: "Failed to calculate JVM options"
```bash
# Debug: Run manually with verbose output
./memory-calculator --total-memory=1G --loaded-class-count=5000
```

**Issue**: Container fails to start
```bash
# Check if entrypoint is executable
docker run --entrypoint=ls myapp:latest -la /entrypoint.sh

# Check memory calculator in container
docker run --entrypoint=memory-calculator myapp:latest --help
```

**Issue**: Kubernetes init container fails
```bash
# Check init container logs
kubectl logs <pod-name> -c jvm-calculator

# Check if ConfigMap is mounted correctly
kubectl describe pod <pod-name>
```

### Debug Commands

```bash
# Test memory detection
./memory-calculator  # Shows detected memory and all options

# Test quiet mode
./memory-calculator --quiet  # Just the JVM options

# Test with explicit values
./memory-calculator --total-memory=2G --loaded-class-count=10000 --thread-count=300

# Check current environment
echo "JAVA_TOOL_OPTIONS: $JAVA_TOOL_OPTIONS"
```

## Contributing

To add new examples:

1. Create a new file in this directory
2. Follow the naming convention: `<scenario>-<type>.<ext>`
3. Include comprehensive comments explaining the setup
4. Add error handling and fallback options
5. Update this README with the new example
6. Test the example in the target environment

## Related Documentation

- [Main README](../README.md) - General usage and features
- [Usage Guide](../USAGE_GUIDE.md) - Detailed integration patterns
- [Helper Script](./set-java-options.sh) - Interactive development usage
