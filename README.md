# JVM Memory Calculator

[![Build Status](https://github.com/patbaumgartner/memory-calculator/actions/workflows/build.yml/badge.svg)](https://github.com/patbaumgartner/memory-calculator/actions/workflows/build.yml)
[![Coverage](https://codecov.io/gh/patbaumgartner/memory-calculator/branch/main/graph/badge.svg)](https://codecov.io/gh/patbaumgartner/memory-calculator)
[![Go Report Card](https://goreportcard.com/badge/github.com/patbaumgartner/memory-calculator)](https://goreportcard.com/report/github.com/patbaumgartner/memory-calculator)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/patbaumgartner/memory-calculator)](https://github.com/patbaumgartner/memory-calculator/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/patbaumgartner/memory-calculator)](https://go.dev/)
[![License](https://img.shields.io/github/license/patbaumgartner/memory-calculator)](LICENSE)

A comprehensive JVM memory calculator compatible with Paketo buildpacks (Temurin, Liberica) that automatically detects container memory limits and calculates optimal JVM memory settings.

## Features

- 🐳 **Container Memory Detection**: Automatically detects memory limits from cgroups v1/v2 with host system fallback
- 📦 **Buildpack Compatibility**: Full integration with Paketo Temurin and Liberica buildpacks
- 🎛️ **Flexible Configuration**: All parameters configurable via command line
- 📏 **Memory Units Support**: Supports B, K, KB, M, MB, G, GB, T, TB with decimal values
- 🤫 **Quiet Mode**: Output only JVM arguments for scripting integration
- 🧪 **Comprehensive Testing**: 77.1% test coverage with unit, integration, and benchmark tests

## Quick Start

### Basic Usage

```bash
# Use automatic memory detection (cgroups + host fallback)
./memory-calculator

# Specify memory and thread count
./memory-calculator --total-memory 2G --thread-count 250

# Quiet mode for scripting
./memory-calculator --total-memory 1G --quiet
```

### Example Output

```
==================================================
JVM Memory Configuration
==================================================
Total Memory:     2.00 GB
Thread Count:     250
Loaded Classes:   35000
Head Room:        0%
Calculated JVM Arguments:
------------------------------
Max Heap Size:         324661K
Thread Stack Size:     1M
Max Metaspace Size:    211914K
Code Cache Size:       240M
Direct Memory Size:    10M
Complete JVM Options:
------------------------------
JAVA_TOOL_OPTIONS=-XX:MaxDirectMemorySize=10M -Xmx324661K -XX:MaxMetaspaceSize=211914K -XX:ReservedCodeCacheSize=240M -Xss1M
```

## Installation

### Download Pre-built Binary

Download the latest release from the [GitHub Releases](https://github.com/patbaumgartner/memory-calculator/releases) page.

### Platform Support

- **Linux**: Full support for cgroups v1/v2 and host detection via `/proc/meminfo`
- **macOS**: Host detection via heuristic methods (no cgroups)  
- **Docker/Containers**: All container runtimes on supported platforms

### Build from Source

```bash
# Clone the repository
git clone https://github.com/patbaumgartner/memory-calculator.git
cd memory-calculator

# Build the executable
go build -o memory-calculator

# Or use make
make build
```

## Development

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

# Run comprehensive quality checks
make quality

# Build for development
make build
```

### Build Commands

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean

# Run tests
make test

# Run tests with coverage
make coverage

# Generate HTML coverage report
make coverage-html

# Run benchmarks
make benchmark

# Install all development tools
make tools

# Check if all tools are available
make tools-check

# Run comprehensive quality checks (format, lint, security, vulnerabilities)
make quality

# Individual quality checks
make format           # Format Go code
make lint             # Run linter
make security         # Run security checks
make vulncheck        # Check for vulnerabilities

# Development and utility
make dev              # Run in development mode with --help
make dev-test         # Run with test parameters
make install          # Install binary to GOPATH/bin
make release-check    # Check if ready for release
make help             # Show all available targets
```

### Testing

The project includes comprehensive test coverage:

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Generate HTML coverage report  
make coverage-html

# Run benchmarks
make benchmark

# Run benchmarks and save results
make benchmark-compare

# Individual test suites (using go test directly)
go test ./internal/memory -v        # Memory parsing tests
go test ./internal/cgroups -v       # Container detection tests
go test ./internal/host -v          # Host memory detection tests
go test ./internal/display -v       # Display formatting tests
go test ./internal/config -v        # Configuration tests
go test ./pkg/errors -v             # Error handling tests
go test -run TestMain -v            # Integration tests only
```

### Test Coverage

Current test coverage: **77.1%** (significantly improved with host detection features)

**Package Coverage Breakdown:**
- `pkg/errors`: **100.0%** - Structured error types with context
- `internal/config`: **100.0%** - Configuration management and validation  
- `internal/display`: **100.0%** - Output formatting and JVM flag extraction
- `internal/memory`: **98.2%** - Memory parsing, formatting, and validation
- `internal/cgroups`: **95.1%** - Container memory detection with host fallback
- `internal/host`: **79.4%** - Cross-platform host memory detection
- `cmd/memory-calculator`: **0.0%** - Main function (tested via integration tests)

**Test Categories:**
- Unit tests for memory parsing and formatting
- Integration tests with binary execution
- Container detection with mock cgroups filesystems  
- Host memory detection across Linux and macOS platforms
- Memory detection fallback priority testing (cgroups → host)
- Comprehensive edge case and error condition testing
- Performance benchmarks for all core components
- Benchmarks for performance validation

## Usage

### Command Line Options

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--total-memory` | Total available memory | Auto-detect from cgroups/host | `2G`, `512M`, `1024MB` |
| `--thread-count` | JVM thread count | `250` | `500` |
| `--loaded-class-count` | Expected loaded classes | `35000` | `40000` |
| `--head-room` | Memory head room percentage | `0` | `10` |
| `--quiet` | Output only JVM options | `false` | - |
| `--help` | Show help message | - | - |

### Memory Units

Supports flexible memory unit specifications:

- **Bytes**: `2147483648` or `2147483648B`
- **Kilobytes**: `1024K`, `1024KB`, `1024kb`
- **Megabytes**: `512M`, `512MB`, `512mb`
- **Gigabytes**: `2G`, `2GB`, `2gb`, `1.5G`
- **Terabytes**: `1T`, `1TB`, `1tb`

### Environment Variables

The calculator respects buildpack environment variables:

- `BPL_JVM_THREAD_COUNT`: Default thread count
- `BPL_JVM_LOADED_CLASS_COUNT`: Default loaded class count
- `BPL_JVM_HEAD_ROOM`: Default head room percentage

### Memory Detection

The memory calculator automatically detects available memory using a prioritized approach:

1. **Container cgroups v2**: `/sys/fs/cgroup/memory.max` (highest priority)
2. **Container cgroups v1**: `/sys/fs/cgroup/memory/memory.limit_in_bytes`
3. **Host system memory**: Platform-specific fallback (lowest priority)

**Platform Support:**
- **Linux**: Reads `/proc/meminfo` for accurate system memory
- **macOS**: Uses heuristic-based detection (CGO-free)
- **Other platforms**: Memory detection not supported

**Detection Priority:**
```bash
# In containers: Uses cgroups limit
docker run --memory=2g my-app
./memory-calculator  # Detects: 2.00 GB

# On host systems: Uses system memory  
./memory-calculator  # Detects: 16.00 GB (example)

# Manual override always takes priority
./memory-calculator --total-memory 4G  # Uses: 4.00 GB
```

### Integration Examples

#### Docker Integration

```dockerfile
FROM paketobuildpacks/builder:base

# Copy memory calculator
COPY memory-calculator /usr/local/bin/

# Use in entrypoint
ENTRYPOINT ["sh", "-c", "export JAVA_TOOL_OPTIONS=$(memory-calculator --quiet) && exec java $@"]
```

#### Shell Script Integration

```bash
#!/bin/bash
# Calculate JVM options
JVM_OPTS=$(./memory-calculator --total-memory 2G --quiet)

# Start Java application
java $JVM_OPTS -jar myapp.jar
```

#### Kubernetes Integration

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    image: myapp:latest
    resources:
      limits:
        memory: "2Gi"
    command: ["sh", "-c"]
    args: ["export JAVA_TOOL_OPTIONS=$(memory-calculator --quiet) && exec java -jar app.jar"]
```

## Architecture

### Memory Calculation Algorithm

1. **Memory Detection**: Automatically detects memory limits from:
   - Container cgroups v2 (`/sys/fs/cgroup/memory.max`)
   - Container cgroups v1 (`/sys/fs/cgroup/memory/memory.limit_in_bytes`)  
   - Host system memory (Linux: `/proc/meminfo`, macOS: heuristic-based)
2. **Memory Allocation**: Distributes memory across JVM components
3. **Heap Calculation**: Calculates max heap with head room
4. **Stack Allocation**: Thread stack size based on thread count
5. **Metaspace Sizing**: Based on expected loaded classes
6. **Code Cache**: Reserved for JIT compilation
7. **Direct Memory**: Off-heap memory allocation

### Component Breakdown

- **Heap Memory**: Primary object storage (largest allocation)
- **Thread Stacks**: Per-thread stack space (thread-count × stack-size)
- **Metaspace**: Class metadata storage (based on loaded classes)
- **Code Cache**: JIT compiled code storage
- **Direct Memory**: Off-heap buffers and NIO

## Contributing

We welcome contributions! Please follow these guidelines:

### Getting Started

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass: `make test`
6. Check code coverage: `make coverage`
7. Commit your changes: `git commit -m 'feat: add amazing feature'`
8. Push to the branch: `git push origin feature/amazing-feature`
9. Open a Pull Request

### Code Style

- Follow Go best practices and conventions
- Use `gofmt` for code formatting
- Add comprehensive tests for new features
- Maintain or improve test coverage
- Include documentation for new features

### Commit Messages

Follow conventional commit format:

- `feat:` new features
- `fix:` bug fixes
- `docs:` documentation changes
- `test:` test additions/changes
- `refactor:` code refactoring
- `style:` code formatting
- `chore:` maintenance tasks

### Testing Requirements

- All new code must include tests
- Tests must pass on all supported Go versions
- Integration tests for CLI functionality
- Benchmark tests for performance-critical code

### Documentation

- Update README.md for new features
- Add inline code documentation
- Include usage examples
- Update command-line help text

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Quick Start for Contributors

1. **Fork** the repository on GitHub
2. **Clone** your fork: `git clone https://github.com/yourusername/memory-calculator.git`
3. **Create a branch**: `git checkout -b feature/amazing-feature`
4. **Make changes** and add tests
5. **Test locally**: `make test && make quality`
6. **Submit a PR** using our pull request template

### GitHub Integration

- **Issues**: Use our issue templates for bugs, features, and questions
- **Pull Requests**: Automated testing and review process
- **Releases**: Fully automated via git tags
- **Actions**: Comprehensive CI/CD pipeline

For detailed guidelines, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Continuous Integration

The project uses **GitHub Actions** for comprehensive automation:

### Automated Testing (Every Push/PR)
- ✅ **Multi-Platform Builds**: Linux & macOS (amd64/arm64)
- ✅ **Test Suite**: Complete test suite with race detection
- ✅ **Coverage Analysis**: Coverage reporting with Codecov integration
- ✅ **Quality Gates**: Linting, security scanning, vulnerability checks
- ✅ **Integration Tests**: End-to-end CLI functionality testing

### Automated Releases (Git Tags)
- 🚀 **Binary Artifacts**: Multi-platform binaries with checksums
- 📝 **Release Notes**: Auto-generated from commit history  
- 🐳 **Docker Images**: Multi-arch containers pushed to registry
- 📦 **Package Distribution**: Ready-to-use downloadable artifacts

### Quality Assurance Pipeline
- **golangci-lint**: Comprehensive code linting
- **gosec**: Security vulnerability scanning
- **govulncheck**: Known vulnerability database checking
- **Dependabot**: Automated dependency updates

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 **Documentation**: Check this README and inline code documentation
- 🐛 **Issues**: Report bugs via [GitHub Issues](https://github.com/patbaumgartner/memory-calculator/issues)
- 💡 **Feature Requests**: Suggest features via [GitHub Issues](https://github.com/patbaumgartner/memory-calculator/issues)
- 📧 **Contact**: Open a discussion for questions

## Acknowledgments

- [Paketo Buildpacks](https://paketo.io/) for the libjvm helper library
- [Temurin](https://adoptium.net/) and [Liberica](https://bell-sw.com/) JDK teams
- Contributors and the Go community

---

**Made with ❤️ for the JVM and container community**
