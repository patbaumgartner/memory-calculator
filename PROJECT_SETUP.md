# Project Setup Guide

Development environment setup and build information for contributors.

## 🚀 Quick Start

```bash
git clone <repo> && cd memory-calculator
make tools deps test build
```

## 📋 Prerequisites

- **Go 1.21+**, **Make**, **Git**
- **Docker** (optional) - For container testing
- **UPX** (optional) - For compressed builds (`apt install upx-ucl` or `brew install upx`)

## 🛠️ Development Commands

```
memory-calculator/
├── .github/                   # GitHub-specific configuration
│   ├── ISSUE_TEMPLATE/       # Issue templates
│   │   ├── bug_report.yml    # Bug report template
│   │   ├── feature_request.yml # Feature request template
│   │   └── question.yml      # Question template
│   ├── workflows/
│   │   └── build.yml         # GitHub Actions CI/CD pipeline
│   ├── dependabot.yml        # Dependabot configuration
│   └── pull_request_template.md # PR template
├── cmd/                      # Application entry points
│   └── memory-calculator/
│       └── main.go           # Main application entry point
├── internal/                 # Private application packages
│   ├── calc/                # Core calculation with build variants
│   │   ├── calc_standard.go # Standard build implementation
│   │   ├── calc_minimal.go  # Minimal build implementation
│   │   └── build_constraints_test.go # Build constraint tests
│   ├── calculator/          # Calculator orchestration
│   ├── cgroups/             # Container memory detection
│   ├── config/              # Configuration management
│   ├── constants/           # Application constants
│   ├── count/               # Class counting with build variants
│   │   ├── count.go         # Standard implementation
│   │   ├── count_minimal.go # Minimal implementation (size-based)
│   │   └── minimal_build_test.go # Minimal build tests
│   ├── display/             # Output formatting
│   ├── host/                # Host memory detection
│   ├── logger/              # Logging utilities
│   └── memory/              # Memory parsing logic
├── pkg/                     # Public packages
│   └── errors/              # Structured error handling
├── examples/                 # Usage examples and scripts
│   ├── docker-entrypoint.sh # Docker container entry script
│   ├── Dockerfile           # Example Dockerfile
│   ├── kubernetes.yaml      # Kubernetes deployment example
│   ├── README.md            # Examples documentation
│   ├── set-java-options.sh  # Java options configuration script
│   └── simple-startup.sh    # Simple startup script
├── testdata/                # Test data and fixtures
│   └── app/                 # Test application files
│       ├── mock.jar         # Mock JAR file for testing
│       └── test.jar         # Test JAR file
├── coverage/                # Test coverage reports (generated)
├── dist/                   # Build artifacts (generated)
├── .gitignore              # Git ignore patterns
├── .golangci.yml           # Go linter configuration
├── .vscode/                # VS Code settings (optional)
├── API.md                  # API documentation
├── ARCHITECTURE.md         # Architecture documentation
├── BINARY_OPTIMIZATION.md  # Binary optimization guide
├── CHANGELOG.md            # Version changelog
├── CONTRIBUTING.md         # Contribution guidelines
├── Dockerfile              # Container build instructions
├── LICENSE                 # MIT License
├── Makefile                # Build automation
├── PROJECT_SETUP.md        # This file - project setup guide
├── README.md               # Main project documentation
├── SECURITY.md             # Security policy and guidelines
├── TEST_COVERAGE.md        # Test coverage documentation
├── TEST_DOCUMENTATION.md   # Test documentation
├── USAGE_GUIDE.md          # Detailed usage guide
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── integration_test.go     # Integration tests
└── memory-calculator       # Built binary (generated)
```
```

## 🛠️ Development Tools

### Makefile Commands

The project includes a comprehensive Makefile with the following targets:

#### Build Commands
- `make build` - Build binary for current platform
- `make build-all` - Build binaries for all platforms
- `make build-minimal` - Build minimal binary without optional features
- `make build-compressed` - Build ultra-compressed binary (requires UPX)
- `make build-size-comparison` - Compare binary sizes with and without optimization
- `make build-ultimate-comparison` - Compare all build variants (standard, minimal, compressed)

#### Test Commands
- `make test` - Run all tests
- `make integration` - Run integration tests only
- `make test-all` - Run all tests including integration tests
- `make test-coverage` - Run tests with coverage
- `make coverage` - Run tests with coverage (alias for test-coverage)
- `make coverage-html` - Generate HTML coverage report
- `make benchmark` - Run benchmarks
- `make benchmark-compare` - Run benchmarks and save results for comparison

#### Quality Commands
- `make quality` - Run all quality checks (format, lint, security, vulnerabilities)
- `make format` - Format Go code
- `make lint` - Run linter (requires golangci-lint)
- `make security` - Run security checks
- `make vuln-check` - Check for known vulnerabilities

#### Development Commands
- `make deps` - Download dependencies
- `make tools` - Install all required development tools
- `make tools-check` - Check if all tools are available
- `make clean` - Clean build artifacts
- `make dev` - Run in development mode
- `make dev-test` - Run with test parameters
- `make install` - Install binary to GOPATH/bin

#### Docker Commands
- `make docker-build` - Build Docker image
- `make docker-run` - Run Docker container
- `make docker-test` - Test Docker container with memory limit

#### Release Commands
- `make release-check` - Check if ready for release
- `make all` - Build everything (clean, deps, test, build)
- `make help` - Show all available commands

### Build Variants

The project supports **two optimized build variants**:

#### Standard Build
```bash
make build
# OR
go build ./cmd/memory-calculator
```
- **Full features**: Complete regex-based parsing
- **Size**: ~2.4MB (30% optimized from 3.5MB original)
- **Dependencies**: All features included

#### Minimal Build
```bash
make build-minimal
# OR
go build -tags minimal ./cmd/memory-calculator
```
- **Size optimized**: String-based parsing
- **Size**: ~2.2MB (37% optimized from 3.5MB original)
- **Dependencies**: Reduced set, eliminates archive/zip

### Binary Size Optimization

The project uses **aggressive optimization flags** and **build constraints** to produce smaller binaries:

#### Optimization Techniques Used:

1. **Build Constraints**: Conditional compilation with `//go:build` tags
2. **Strip Debug Information**: `-ldflags="-s -w"`
   - `-s`: Removes symbol table and debug info
   - `-w`: Removes DWARF debug information
3. **Reproducible Builds**: `-trimpath`
   - Removes file system paths from the executable
   - Ensures consistent builds across environments
4. **Force Clean Rebuilds**: `-a`
   - Forces rebuilding of all packages
   - Ensures optimal linking

#### Size Comparison:
```bash
make build-size-comparison
# Output example:
# Unoptimized: 3.5M
# Optimized:   2.4M  
# Savings:     30.0%
```

#### Ultra Compression (Optional):
For even smaller binaries, install UPX and use:
```bash
# Ubuntu/Debian
sudo apt install upx-ucl

# macOS  
brew install upx

# Build with UPX compression
make build-compressed
```

**Note**: UPX compression trades startup time for smaller file size.

### GitHub Actions
Comprehensive CI/CD pipeline that automatically:

**On Every Push/PR:**
- ✅ **Tests**: Runs complete test suite with race detection
- ✅ **Coverage**: Generates coverage reports (uploads to Codecov)
- ✅ **Quality**: Runs golangci-lint with custom configuration
- ✅ **Security**: Performs gosec security scanning
- ✅ **Vulnerabilities**: Checks for known vulnerabilities with govulncheck
- ✅ **Cross-Platform Builds**: Builds for all supported platforms

**On Git Tags (v*):**
- 🚀 **Automated Releases**: Creates GitHub releases with binaries
- 📦 **Multi-Platform Artifacts**: Builds and uploads platform-specific binaries
- 🔐 **Checksums**: Generates SHA256 checksums for all artifacts
- 📝 **Release Notes**: Auto-generates release notes from commits

**Docker Support:**
- 🐳 **Multi-Arch Images**: Builds for linux/amd64 and linux/arm64
- 🏷️ **Smart Tagging**: Version tags, latest tag, and branch tags
- 📤 **Registry Push**: Pushes to Docker Hub (when configured)

**Dependency Management:**
- 🔄 **Dependabot**: Weekly automated dependency updates
- 📋 **Go Modules**: Automatic Go dependency updates
- ⚙️ **GitHub Actions**: Keeps workflow actions up-to-date
- 🐳 **Docker**: Updates base Docker images

### GitHub Issue & PR Templates
Structured templates for better collaboration:

- **Bug Reports**: YAML-based form with environment details
- **Feature Requests**: Structured feature proposal template
- **Questions**: Template for asking questions and getting help
- **Pull Requests**: Comprehensive PR checklist and guidelines

### Supported Platforms
- **Linux**: amd64, arm64
- **macOS**: amd64, arm64 (Apple Silicon)
- **Container**: Docker multi-arch support

## 🏗️ Build System

### Local Development
```bash
# Clone and setup
git clone <repo>
cd memory-calculator
make deps

# Development cycle
make test           # Run tests
make build         # Build local binary
./memory-calculator --help

# Cross-platform builds
make build-all     # Build for all platforms
ls dist/           # Check artifacts
```

### Version Information
Build-time variables injected via ldflags:
- `version` - Git tag or "dev"
- `buildTime` - Build timestamp
- `commitHash` - Git commit hash

### Docker Support
```bash
# Build container
docker build -t memory-calculator .

# Run container
docker run --rm memory-calculator --help
```

## 🧪 Testing Framework

### Test Coverage: 77.5%
The codebase has been refactored with a professional package structure providing excellent test coverage:

- **Unit Tests**: Per-package testing with dependency injection
- **Integration Tests**: Full binary execution and end-to-end testing
- **Benchmark Tests**: Performance validation across all packages
- **Mock Tests**: cgroups simulation and file system mocking

### Test Architecture by Package
- `integration_test.go` - End-to-end application testing
- `internal/calc/*_test.go` - Core calculation and build constraint tests
- `internal/calculator/*_test.go` - Calculator orchestration tests
- `internal/cgroups/*_test.go` - Container memory detection tests (94.6% coverage)
- `internal/config/*_test.go` - Configuration management tests (100% coverage)
- `internal/constants/*_test.go` - Constants and validation tests
- `internal/count/*_test.go` - Class counting and minimal build tests
- `internal/display/*_test.go` - Output formatting tests (100% coverage)
- `internal/host/*_test.go` - Host memory detection tests
- `internal/logger/*_test.go` - Logging utilities tests
- `internal/memory/*_test.go` - Memory parsing tests (95.7% coverage)
- `internal/parser/*_test.go` - Flag parsing tests (100% coverage)
- `pkg/errors/*_test.go` - Structured error handling tests (100% coverage)

### Package Coverage Summary
| Package | Coverage | Status |
|---------|----------|--------|
| `pkg/errors` | 100% | ✅ Complete |
| `internal/config` | 100% | ✅ Complete |
| `internal/display` | 100% | ✅ Complete |
| `internal/parser` | 100% | ✅ Complete |
| `internal/memory` | 95.7% | ✅ Excellent |
| `internal/cgroups` | 94.6% | ✅ Excellent |
| **Total Coverage** | **77.5%** | ✅ **Good** |

## 📦 Dependencies

### Direct Dependencies
- The project has no external dependencies beyond the Go standard library.

### Build Dependencies
All transitive dependencies managed automatically by Go modules.

## 🚀 Release Process

### Automated Releases (Recommended)
The project uses **fully automated releases** triggered by Git tags:

```bash
# 1. Ensure everything is committed and pushed
git add -A
git commit -m "feat: prepare for v1.2.0 release"
git push origin main

# 2. Create and push a version tag
git tag v1.2.0
git push origin v1.2.0

# 3. GitHub Actions automatically:
#    - Runs full test suite
#    - Builds binaries for all platforms (Linux/macOS, amd64/arm64)
#    - Creates GitHub release with auto-generated notes
#    - Uploads all artifacts with checksums
#    - Builds and pushes Docker images
```

### Release Artifacts Generated
For each release, the following artifacts are automatically created:

- `memory-calculator-linux-amd64` - Linux x86_64 binary
- `memory-calculator-linux-arm64` - Linux ARM64 binary  
- `memory-calculator-darwin-amd64` - macOS Intel binary
- `memory-calculator-darwin-arm64` - macOS Apple Silicon binary
- `checksums.txt` - SHA256 checksums for all binaries
- **Docker Images**: Multi-arch images pushed to Docker Hub

### Manual Release (Fallback)
If needed, you can create releases manually:

```bash
# Check release readiness
make release-check

# Build all platforms locally
make build-all

# Create release manually via GitHub UI or GitHub CLI
gh release create v1.2.0 dist/* --generate-notes
```

### Release Checklist
Before creating a release:

- [ ] All tests pass (`make test`)
- [ ] Quality checks pass (`make quality`)
- [ ] CHANGELOG.md updated
- [ ] Version number updated in relevant files
- [ ] Working directory is clean (`git status`)
- [ ] All changes pushed to main branch

### Versioning Strategy
The project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality
- **PATCH** version for backwards-compatible bug fixes

Examples:
- `v1.0.0` - Major release
- `v1.1.0` - New features, backwards compatible
- `v1.1.1` - Bug fixes, backwards compatible

## 📋 Quality Assurance

### Code Quality
- Go standard formatting (`gofmt`)
- Comprehensive test coverage
- Conventional commit messages
- Documentation for all public APIs

### CI/CD Pipeline
- ✅ Automated testing on multiple Go versions
- ✅ Cross-platform build verification
- ✅ Code coverage reporting
- ✅ Automated release generation
- ✅ Docker image building

### Security
- Non-root container execution
- Minimal container dependencies
- No external runtime dependencies

## 🌟 Key Features Implemented

1. **Container Memory Detection** - Automatic cgroups v1/v2 detection
2. **Buildpack Compatibility** - Full Paketo buildpack integration
3. **Flexible CLI** - Comprehensive command-line interface
4. **Memory Units** - Support for B, K, KB, M, MB, G, GB, T, TB
5. **Quiet Mode** - Script-friendly output format
6. **Version Information** - Build-time version injection
7. **Cross-Platform** - Linux and macOS support
8. **Container Ready** - Docker support with multi-arch builds

## 📝 Documentation

The project includes comprehensive documentation:

- **README.md** - Complete user and developer guide
- **CONTRIBUTING.md** - Contribution guidelines and workflow
- **PROJECT_SETUP.md** - This file - project setup and development guide
- **TEST_DOCUMENTATION.md** - Test framework documentation
- **TEST_COVERAGE.md** - Test coverage documentation
- **USAGE_GUIDE.md** - Detailed usage guide and examples
- **API.md** - API documentation and reference
- **ARCHITECTURE.md** - Architecture documentation and design decisions
- **BINARY_OPTIMIZATION.md** - Binary optimization guide and techniques
- **SECURITY.md** - Security policy and guidelines
- **CHANGELOG.md** - Version changelog and release notes
- **Inline Documentation** - Godoc-style code comments
- **Usage Examples** - Multiple integration scenarios in `examples/`

## 🎯 Project Status

**✅ Complete and Production Ready**

The JVM Memory Calculator is fully functional with:
- Comprehensive test suite (77.5% coverage with professional package structure)
- Automated CI/CD pipeline
- Cross-platform build support
- Professional documentation
- Container-ready deployment
- Contribution-friendly setup

Ready for:
- Production deployment
- Community contributions
- Further feature development
- Integration into other projects
