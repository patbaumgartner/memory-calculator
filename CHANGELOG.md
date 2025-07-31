# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2025-07-31

### Added
- **Host Memory Detection**: Added cross-platform host memory detection as fallback when cgroups are not available
  - Linux: Reads `/proc/meminfo` for accurate system memory detection
  - macOS: Heuristic-based detection without CGO dependencies (Windows support removed)
  - Prioritized detection: cgroups v2 → cgroups v1 → host system memory
- **Enhanced Memory Detection**: Updated memory detection algorithm with intelligent fallback mechanism
- **Cross-Platform Testing**: Comprehensive test suite covering all supported platforms and edge cases
- **Enhanced Build System**: Improved Makefile with smart Go tool path resolution
  - Auto-detection of Go installation paths (GOBIN/GOPATH)
  - Auto-installation of missing development tools
  - New `make quality` target for comprehensive code quality checks
  - New `make tools` and `make tools-check` targets for development tool management
- **Documentation**: Updated README and test documentation with host detection details

### Changed
- **Memory Detection Logic**: Enhanced cgroups detector to include host system fallback
- **Platform Support**: Removed Windows support - now supports Linux and macOS only
- **Code Quality**: Improved code quality by extracting platform strings into constants
- **Test Coverage**: Maintained 77.1% coverage with new host detection tests
- **Error Messages**: Updated log messages to reflect enhanced detection capabilities
- **Build Tools**: Enhanced Makefile to properly handle Go module installation paths

### Fixed
- **Makefile Tool Path Resolution**: Fixed issue where Go-installed tools weren't found in PATH
- **Linter Issues**: Resolved golangci-lint warnings about repeated string literals
- **Platform Constants**: Centralized platform names to improve maintainability

## [1.0.0] - 2025-07-31

### Added
- **Core Features**
  - JVM memory calculation engine using Paketo buildpack libjvm helper
  - Automatic container memory detection via cgroups v1/v2
  - Command-line interface with comprehensive flag support
  - Flexible memory unit parsing (B, K, KB, M, MB, G, GB, T, TB)
  - Quiet mode for scripting integration (`--quiet` flag)
  - Version information system with build-time injection

- **Platform Support**
  - Linux support (amd64, arm64)
  - macOS support (amd64, arm64/Apple Silicon)
  - Docker containerization with multi-architecture support

- **Build System**
  - Professional Makefile with comprehensive build targets
  - Cross-platform build automation
  - Version injection via ldflags
  - Clean artifact management

- **CI/CD Pipeline**
  - GitHub Actions workflow for automated testing
  - Multi-platform build matrix (Linux, macOS)
  - Automated release creation on git tag push
  - Artifact upload for easy distribution
  - Docker image building and publishing support
  - Dependabot configuration for dependency updates

- **Testing Framework**
  - Comprehensive test suite with 75.2% code coverage (significantly improved)
  - Unit tests for all core functions
  - Integration tests with binary execution
  - Benchmark tests for performance validation
  - Mock cgroups testing for container scenarios
  - Edge case testing for robustness

- **Documentation**
  - Comprehensive README with usage examples
  - Detailed contribution guidelines (CONTRIBUTING.md)
  - Technical project setup documentation
  - Test framework documentation
  - Security policy (SECURITY.md)
  - Professional issue and PR templates

- **Container Integration**
  - Docker support with optimized multi-stage builds
  - Non-root user execution for security
  - Integration examples for Docker, Kubernetes, shell scripts
  - Buildpack environment variable support

- **Memory Calculation Features**
  - Heap memory calculation with configurable head room
  - Thread stack sizing based on thread count
  - Metaspace allocation based on loaded class count  
  - Code cache reservation for JIT compilation
  - Direct memory allocation for off-heap usage
  - Professional output formatting with detailed breakdown

### Technical Details
- **Language**: Go 1.24.5
- **Dependencies**: Minimal dependency footprint with only libjvm helper
- **Architecture**: Clean, testable code structure
- **Performance**: Optimized for fast execution in container environments
- **Security**: Input validation, secure defaults, minimal attack surface

### Compatibility
- **Buildpacks**: Full compatibility with Paketo Temurin and Liberica buildpacks
- **Containers**: Works with Docker, Podman, and Kubernetes
- **JVM**: Generates standard JVM memory flags compatible with all major JVM implementations
- **Environments**: Development, staging, and production ready

[Unreleased]: https://github.com/patbaumgartner/memory-calculator/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/patbaumgartner/memory-calculator/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/patbaumgartner/memory-calculator/releases/tag/v1.0.0
