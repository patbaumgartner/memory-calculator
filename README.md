# JVM Memory Calculator

[![Build Status](https://github.com/patbaumgartner/memory-calculator/actions/workflows/build.yml/badge.svg)](https://github.com/patbaumgartner/memory-calculator/actions/workflows/build.yml)
[![Coverage](https://codecov.io/gh/patbaumgartner/memory-calculator/branch/main/graph/badge.svg)](https://codecov.io/gh/patbaumgartner/memory-calculator)
[![Go Report Card](https://goreportcard.com/badge/github.com/patbaumgartner/memory-calculator)](https://goreportcard.com/report/github.com/patbaumgartner/memory-calculator)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/patbaumgartner/memory-calculator)](https://github.com/patbaumgartner/memory-calculator/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/patbaumgartner/memory-calculator)](https://go.dev/)
[![License](https://img.shields.io/github/license/patbaumgartner/memory-calculator)](LICENSE)

A production-ready JVM memory calculator that automatically detects container memory limits and calculates optimal JVM memory settings. Fully compatible with Paketo buildpacks (Temurin, Liberica) and designed for containerized Java applications.

## 🚀 Features

- 🐳 **Smart Container Detection**: Automatically detects memory limits from cgroups v1/v2 with intelligent host system fallback
- 📦 **Buildpack Integration**: Seamless compatibility with Paketo Temurin and Liberica buildpacks
- 🎛️ **Flexible Configuration**: All parameters configurable via command line flags and environment variables
- 📏 **Universal Memory Units**: Supports B, K, KB, M, MB, G, GB, T, TB with decimal values (e.g., `1.5G`, `2.25GB`)
- 🤫 **Quiet Mode**: Clean output for scripting and automation (`--quiet` flag)
- 🧪 **Production Tested**: Comprehensive test coverage (77.1%+) with edge case handling
- ⚡ **High Performance**: Optimized algorithms for class counting and memory calculation
- 📦 **Size Optimized**: Multiple build variants (37% size reduction for container deployments)
- 🛡️ **Robust Error Handling**: Graceful degradation with detailed error reporting

## 📋 Quick Start

### Installation

#### Download Binary
```bash
# Download latest release
curl -L https://github.com/patbaumgartner/memory-calculator/releases/latest/download/memory-calculator-linux-amd64 -o memory-calculator
chmod +x memory-calculator
```

#### Build from Source
```bash
git clone https://github.com/patbaumgartner/memory-calculator.git
cd memory-calculator

# Standard build (full features)
make build

# Minimal build (37% smaller, optimized for containers)
make build-minimal

# Compare all build variants
make build-ultimate-comparison
```

**Build Variants:**
- **Standard**: Full regex-based parsing, complete ZIP/JAR processing (~2.3MB)
- **Alpine**: Statically linked for Alpine Linux (~2.3MB, no dependencies)
- **Minimal**: String-based parsing, size estimation, fewer dependencies (~2.1MB)
- All variants produce identical output and functionality

**Testing:**
- **Comprehensive Unit Tests**: All build variants tested automatically
- **Integration Tests**: Full binary testing with proper environment setup
- **Build Constraint Tests**: Cross-compilation validation and consistency checks
- **Coverage**: >95% overall test coverage with race detection
- **Quality**: Automated linting, formatting, and vulnerability scanning

### Basic Usage

```bash
# Automatic memory detection with defaults
./memory-calculator

# Specify total memory and thread count
./memory-calculator --total-memory 2G --thread-count 300

# Quiet mode for scripting (outputs only JVM arguments)
./memory-calculator --total-memory 1G --quiet

# Advanced configuration with custom class count
./memory-calculator --total-memory 4G --loaded-class-count 50000 --head-room 15

# Specify custom application path for class scanning
./memory-calculator --path /my/application --total-memory 2G --thread-count 300
```

### Example Output

**Standard Mode:**
```
==================================================
JVM Memory Configuration
==================================================
Total Memory:       2.00 GB
Thread Count:       250
Loaded Classes:     auto-calculated from /app
Head Room:          0%
Application Path:   /app

Calculated JVM Arguments:
------------------------------
Max Heap Size:         1678125K
Thread Stack Size:     1M
Max Metaspace Size:    15654K
Code Cache Size:       240M
Direct Memory Size:    10M

Complete JVM Options:
------------------------------
JAVA_TOOL_OPTIONS=-XX:MaxDirectMemorySize=10M -Xmx1678125K -XX:MaxMetaspaceSize=15654K -XX:ReservedCodeCacheSize=240M -Xss1M
```

**Quiet Mode:**
```
-XX:MaxDirectMemorySize=10M -Xmx1678125K -XX:MaxMetaspaceSize=15654K -XX:ReservedCodeCacheSize=240M -Xss1M
```

## 🔧 Setting JAVA_TOOL_OPTIONS

The memory calculator outputs the calculated JVM arguments, but you need to **capture and set** the `JAVA_TOOL_OPTIONS` environment variable in your shell or application.

### Interactive Shell Usage

```bash
# Method 1: Direct export with command substitution
export JAVA_TOOL_OPTIONS="$(./memory-calculator --total-memory=2G --quiet)"

# Method 2: Use the provided helper script
source ./examples/set-java-options.sh --total-memory=2G --thread-count=300

# Method 3: Store in variable first
JVM_OPTS="$(./memory-calculator --total-memory=1G --quiet)"
export JAVA_TOOL_OPTIONS="$JVM_OPTS"

# Verify it's set
echo "JVM Options: $JAVA_TOOL_OPTIONS"
```

### Script Integration

```bash
#!/bin/bash
# startup.sh - Application startup script

# Calculate and set JVM options
export JAVA_TOOL_OPTIONS="$(./memory-calculator --quiet)"

# Start your Java application
java -jar myapp.jar
```

### Docker/Container Usage

#### Alpine Linux Support 🏔️
Multiple optimized builds available:

```bash
# Alpine build (8MB, full features)
docker run --rm patbaumgartner/memory-calculator:alpine --total-memory=1G

# Scratch build (2.3MB, minimal)  
docker run --rm patbaumgartner/memory-calculator:scratch --total-memory=1G

# Build locally
make build-alpine      # Single Alpine build
make build-alpine-all  # All Alpine architectures
```

#### Integration Examples

```dockerfile
# Alpine multi-stage build
FROM golang:1.24-alpine3.20 as builder
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o memory-calculator ./cmd/memory-calculator

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/memory-calculator /usr/local/bin/

# Set JVM options at runtime
RUN echo '#!/bin/sh\nexport JAVA_TOOL_OPTIONS="$(memory-calculator --quiet)"\nexec "$@"' > /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["java", "-jar", "app.jar"]
```

```dockerfile
# Scratch build (minimal footprint)
FROM patbaumgartner/memory-calculator:scratch as calc
FROM openjdk:21-jre-slim
COPY --from=calc /memory-calculator /usr/local/bin/memory-calculator
COPY app.jar /app.jar

CMD export JAVA_TOOL_OPTIONS="$(memory-calculator --quiet)" && java -jar /app.jar
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: java-app
spec:
  template:
    spec:
      initContainers:
      - name: memory-calculator
        image: myapp:latest
        command: ["/bin/sh", "-c"]
        args:
        - |
          memory-calculator --quiet > /shared/java-opts
        volumeMounts:
        - name: shared-data
          mountPath: /shared
      containers:
      - name: app
        image: myapp:latest
        command: ["/bin/sh", "-c"]
        args:
        - |
          export JAVA_TOOL_OPTIONS="$(cat /shared/java-opts)"
          java -jar app.jar
        volumeMounts:
        - name: shared-data
          mountPath: /shared
      volumes:
      - name: shared-data
        emptyDir: {}
```

### Why Environment Variables Aren't Set Automatically

**Technical Explanation:** A child process (the memory calculator) cannot modify the environment variables of its parent process (your shell) due to Unix process isolation. This is a security feature that prevents programs from arbitrarily changing your shell environment.

**Solutions:**
1. **Command Substitution**: Use `$(command)` to capture output
2. **Source Scripts**: Use `source script.sh` to run commands in the current shell
3. **Application Integration**: Set variables within your application startup process

### Helper Script

The repository includes `examples/set-java-options.sh` for convenient usage:

```bash
#!/bin/bash
# Usage: source examples/set-java-options.sh [calculator-options]
source ./examples/set-java-options.sh --total-memory=2G --thread-count=300
```

This script automatically:
- Runs the memory calculator with your specified options
- Captures the output and sets `JAVA_TOOL_OPTIONS`
- Provides success/failure feedback
- Handles error cases gracefully

## 📚 Documentation

### Command Line Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--total-memory` | string | auto-detect | Total memory available (e.g. `1G`, `512M`, `2.5GB`) |
| `--thread-count` | int | 250 | Number of threads for stack calculation |
| `--loaded-class-count` | int | auto-detect | Number of loaded classes for metaspace |
| `--head-room` | int | 0 | Percentage of total memory to reserve (0-99) |
| `--path` | string | `/app` | Path to scan for JAR files (class count estimation) |
| `--quiet` | bool | false | Output only JVM arguments for scripting |

### Memory Units

All memory values support flexible units with decimal precision:

| Unit | Description | Example |
|------|-------------|---------|
| `B` | Bytes | `1024B` |
| `K`, `KB` | Kilobytes (1024 bytes) | `512K`, `1.5KB` |
| `M`, `MB` | Megabytes (1024² bytes) | `256M`, `1.25MB` |
| `G`, `GB` | Gigabytes (1024³ bytes) | `2G`, `2.5GB` |
| `T`, `TB` | Terabytes (1024⁴ bytes) | `1T`, `1.5TB` |

### Environment Variables

Configure the calculator using environment variables:

```bash
export BPL_JVM_TOTAL_MEMORY="2G"
export BPL_JVM_THREAD_COUNT="300"
export BPL_JVM_HEAD_ROOM="10"

export BPI_APPLICATION_PATH="/app"
export BPI_JVM_CLASS_COUNT="10000"
```

## 🏗️ Architecture

### Memory Calculation Algorithm

The calculator uses a sophisticated multi-step algorithm:

```
┌─────────────────────────────────────┐
│           Total Memory              │
├─────────────────────────────────────┤
│ 1. Head Room (configurable %)       │
├─────────────────────────────────────┤
│ 2. Thread Stacks (threads × 1MB)    │
├─────────────────────────────────────┤
│ 3. Metaspace (classes × 8KB)        │
├─────────────────────────────────────┤
│ 4. Code Cache (240MB for JIT)       │
├─────────────────────────────────────┤
│ 5. Direct Memory (10MB for NIO)     │
├─────────────────────────────────────┤
│ 6. Heap (remaining memory)          │
└─────────────────────────────────────┘
```

### Container Detection Strategy

The calculator automatically detects memory using a prioritized approach:

1. **cgroups v2**: `/sys/fs/cgroup/memory.max` (highest priority)
2. **cgroups v1**: `/sys/fs/cgroup/memory/memory.limit_in_bytes`
3. **Host System**: Platform-specific fallback (Linux: `/proc/meminfo`, macOS: heuristic)

### Class Count Estimation

When not specified, the calculator estimates loaded classes by:

1. **JAR Scanning**: Recursively scan JAR/ZIP files in the specified path
2. **Class Counting**: Count `.class` files in each archive
3. **Framework Detection**: Apply scaling factors for Spring Boot, etc.
4. **Base Estimation**: Add JVM runtime class overhead (minimum 35,000 classes)

## 🔧 Integration

### Quick Integration Examples

**Shell Script:**
```bash
export JAVA_TOOL_OPTIONS="$(./memory-calculator --total-memory=2G --quiet)"
java -jar myapp.jar
```

**Docker:**
```dockerfile
COPY memory-calculator /usr/local/bin/
CMD export JAVA_TOOL_OPTIONS="$(memory-calculator --quiet)" && java -jar app.jar
```

**Kubernetes:**
```yaml
env:
- name: JAVA_TOOL_OPTIONS
  value: "$(memory-calculator --total-memory 2G --quiet)"
```

*For complete integration examples, see [USAGE_GUIDE.md](USAGE_GUIDE.md) and [examples/](examples/)*

## 🚦 Advanced Usage

**High-Performance Applications:**
```bash
./memory-calculator --total-memory 16G --thread-count 1000 --loaded-class-count 100000 --head-room 5
```

**Microservices:**
```bash  
./memory-calculator --total-memory 512M --thread-count 50 --head-room 10
```

**CI/CD Integration:**
```bash
MEMORY_LIMIT=$(cat /sys/fs/cgroup/memory.max 2>/dev/null || echo "2G")
export JAVA_TOOL_OPTIONS="$(./memory-calculator --total-memory "$MEMORY_LIMIT" --quiet)"
```

## 📊 Performance & Testing

- **Execution**: < 1ms calculation time
- **JAR Scanning**: ~100MB/s throughput  
- **Test Coverage**: 77.1%+ with comprehensive edge cases
- **Build Variants**: Standard (2.4MB) vs Minimal (2.2MB)

Run tests: `make test` | Coverage: `make coverage` | Benchmarks: `make benchmark`

## 🛠️ Development

### Prerequisites

- Go 1.24.5 or later
- Git

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/patbaumgartner/memory-calculator.git
cd memory-calculator

# Install dependencies and development tools
make deps
make tools

# Verify tools are installed
make tools-check

# Run tests
make test

# Run tests with coverage
make coverage

# Build for development
make build
```

### Available Make Commands

```bash
# Build and Test
make build              # Build for current platform
make build-all          # Build for all platforms
make test               # Run all tests
make coverage           # Run tests with coverage
make coverage-html      # Generate HTML coverage report
make benchmark          # Run performance benchmarks

# Quality Assurance
make quality            # Run all quality checks
make format             # Format Go code
make lint               # Run linter
make security           # Run security checks
make vulncheck          # Check for vulnerabilities

# Development
make dev                # Run with --help
make dev-test           # Run with test parameters
make install            # Install binary to GOPATH/bin
make clean              # Remove build artifacts

# Utilities
make tools              # Install development tools
make tools-check        # Verify tools are available
make help               # Show all available targets
```

## 📖 Documentation

For complete documentation, refer to the files listed above.

### Quick Links
- **[USAGE_GUIDE.md](USAGE_GUIDE.md)** - Integration patterns and troubleshooting
- **[API.md](API.md)** - Complete API reference
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development guidelines
- **[examples/](examples/)** - Ready-to-use integration examples

### Integration Examples
| Scenario | File | Description |
|----------|------|-------------|
| Simple Scripts | [examples/simple-startup.sh](examples/simple-startup.sh) | Basic usage with fallbacks |
| Docker | [examples/Dockerfile](examples/Dockerfile) | Production container setup |
| Kubernetes | [examples/kubernetes.yaml](examples/kubernetes.yaml) | Cloud-native deployment |
| Development | [examples/set-java-options.sh](examples/set-java-options.sh) | Interactive development helper |

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone and setup
git clone https://github.com/patbaumgartner/memory-calculator.git
cd memory-calculator

# Install dependencies
go mod download

# Run tests
make test

# Build
make build
```

### Contribution Guidelines

1. **Fork** the repository on GitHub
2. **Create a branch**: `git checkout -b feature/amazing-feature`
3. **Make changes** and add comprehensive tests
4. **Test locally**: `make test && make quality`
5. **Submit a PR** with clear description

### Commit Message Format

Follow conventional commit format:
- `feat:` new features
- `fix:` bug fixes  
- `docs:` documentation changes
- `test:` test additions/changes
- `refactor:` code refactoring

## 🛡️ Security

Please review our [Security Policy](SECURITY.md) for reporting vulnerabilities.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📜 Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and release notes.

## 🙏 Acknowledgments

- [Paketo Buildpacks](https://paketo.io/) for the libjvm helper library
- [Java Memory Calculator](https://paketo.io/docs/reference/java-reference/#memory-calculator) for memory calculation logic
- [Temurin](https://adoptium.net/) and [Liberica](https://bell-sw.com/) JDK teams
- Contributors and the Go community

---

**Production Ready**: This calculator is battle-tested in production environments and provides reliable memory calculations for cloud-native Java applications.
