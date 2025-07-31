# JVM Memory Calculator

A comprehensive JVM memory calculator compatible with Paketo buildpacks (Temurin, Liberica) that automatically detects container memory limits and calculates optimal JVM memory settings.

## Features

- 🐳 **Container Memory Detection**: Automatically detects memory limits from cgroups v1/v2
- 📦 **Buildpack Compatibility**: Full integration with Paketo Temurin and Liberica buildpacks
- 🎛️ **Flexible Configuration**: All parameters configurable via command line
- 📏 **Memory Units Support**: Supports B, K, KB, M, MB, G, GB, T, TB with decimal values
- 🤫 **Quiet Mode**: Output only JVM arguments for scripting integration
- 🧪 **Comprehensive Testing**: 53%+ test coverage with unit, integration, and benchmark tests

## Quick Start

### Basic Usage

```bash
# Use automatic container memory detection
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

**Supported Platforms:**
- Linux (amd64, arm64)
- macOS (amd64, arm64/Apple Silicon)
- Windows support is limited due to container-specific dependencies

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

# Install dependencies
go mod download

# Run tests
go test -v

# Run with coverage
go test -cover

# Build for development
go build -o memory-calculator
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
make test-coverage

# Generate coverage report
make coverage-html
```

### Testing

The project includes comprehensive test coverage:

```bash
# Run all tests
go test -v

# Run tests with coverage
go test -v -cover

# Run specific test suites
go test -run TestMemory -v
go test -run TestIntegration -v
go test -run TestCgroups -v

# Generate detailed coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Test Coverage

Current test coverage: **53.5%**

- Unit tests for memory parsing and formatting
- Integration tests with binary execution
- cgroups mocking and edge case testing
- Benchmarks for performance validation

## Usage

### Command Line Options

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--total-memory` | Total available memory | Auto-detect from cgroups | `2G`, `512M`, `1024MB` |
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

1. **Container Detection**: Reads memory limits from cgroups v1/v2
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
6. Check code coverage: `make test-coverage`
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

## Continuous Integration

The project uses GitHub Actions for:

- **Build Verification**: Multi-platform builds (Linux, macOS)
- **Test Execution**: Comprehensive test suite execution
- **Coverage Reporting**: Code coverage analysis
- **Release Automation**: Automatic binary releases
- **Artifact Publishing**: Downloadable binaries for each platform

### Build Matrix

- **Operating Systems**: Linux, macOS
- **Architectures**: amd64, arm64
- **Go Versions**: 1.24.5+

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
