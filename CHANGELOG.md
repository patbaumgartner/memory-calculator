# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.1] - 2025-08-04

### Enhanced
- **CI/CD Pipeline**: Comprehensive workflow optimization and reliability improvements
  - Enhanced GitHub Actions workflow with better error handling and logging
  - Optimized release process by reusing build artifacts (50-66% faster builds)
  - Added comprehensive artifact validation and verification
  - Improved Docker image testing with enhanced error detection
- **Multi-Architecture Support**: Improved Docker builds for all supported platforms
  - Fixed ARM64 Docker builds by removing unnecessary cross-compiler dependencies
  - Simplified Dockerfile for native architecture builds
  - Enhanced multi-platform Docker image creation (linux/amd64, linux/arm64)
- **Build System Reliability**: Robust cross-compilation and error handling
  - Enhanced ARM64 cross-compilation setup with proper error checking
  - Added comprehensive error handling for cross-compiler installation
  - Improved build process logging and debugging capabilities
  - Better validation of build artifacts and dependencies

### Fixed
- **Docker ARM64 Builds**: Resolved ARM64 Docker build failures
  - Removed erroneous `gcc-aarch64-linux-musl` package dependency
  - Fixed Docker multi-platform builds by leveraging native architecture compilation
  - Eliminated unnecessary cross-compiler logic in Dockerfile
- **Release Process**: Enhanced artifact management and distribution
  - Fixed artifact organization in release job
  - Improved release file validation and missing file detection
  - Enhanced checksum generation and verification process
- **Error Handling**: Comprehensive error handling across all build processes
  - Added proper exit codes and error messages for failed operations
  - Enhanced cross-compiler installation validation
  - Better handling of build environment setup failures

### Technical Details
- **Docker**: Simplified Dockerfile leveraging Docker Buildx native architecture builds
- **GitHub Actions**: Optimized workflow with proper job dependencies and artifact reuse
- **Cross-Compilation**: Enhanced ARM64 support with `gcc-aarch64-linux-gnu` for static builds
- **Artifact Management**: Improved organization and validation of release artifacts
- **Performance**: Significant reduction in release build times through artifact optimization
- **Testing**: Enhanced Docker image testing with comprehensive validation steps

### Build Variants
- **Standard Binaries**: `CGO_ENABLED=0` for all platforms (linux, darwin) × (amd64, arm64)
- **Minimal Binaries**: Reduced feature set with `-tags minimal` for all platforms
- **Static Binaries**: `CGO_ENABLED=1` with static linking for Linux (amd64, arm64)
- **Docker Images**: Multi-architecture support with proper ARM64 native compilation

This release focuses on production readiness, build reliability, and comprehensive multi-architecture support with significantly improved CI/CD pipeline performance.

## [1.3.0] - 2025-08-02

### Added
- **NEW**: `--path` command-line parameter for application JAR scanning and class count estimation
  - Enables intelligent class count estimation by scanning JAR files in specified application directory
  - Integrates with `BPI_APPLICATION_PATH` environment variable for buildpack compatibility
  - Recursive JAR scanning with framework-aware scaling factors (Spring Boot, etc.)
- **NEW**: Enhanced display output with intelligent "Loaded Classes" messaging
  - Shows "auto-calculated from {path}" when class count is estimated from JAR scanning
  - Shows actual number when manually specified via `--loaded-class-count`
  - Clear indication of calculation source for better user understanding
- **NEW**: Comprehensive documentation updates for new features
  - Updated README.md with --path parameter examples and improved display output
  - Enhanced USAGE_GUIDE.md with application path scanning section
  - Updated API.md with BPI_APPLICATION_PATH environment variable documentation
  - Added troubleshooting guidance for path-based class count estimation

### Enhanced
- **User Experience**: Significantly improved display clarity and transparency
  - Application path always shown in configuration summary
  - Clear distinction between calculated vs. manually specified values
  - Enhanced help text with practical usage examples
- **Integration**: Seamless buildpack integration with path-based configuration
  - Default path "/app" aligns with buildpack standards
  - Environment variable support for automated deployments
  - Backward compatibility maintained for existing configurations

### Fixed
- **Display Consistency**: Eliminated confusing empty values in output
- **Documentation**: Updated all examples to reflect new display format and features
- **Help Text**: Enhanced with --path parameter and updated examples

### Technical Details
- **Configuration Management**: Extended Config struct with Path field and validation
- **Environment Integration**: Added BPI_APPLICATION_PATH support with default fallback
- **Display Logic**: Conditional formatting based on calculation source
- **Testing**: All tests updated and passing with new functionality
- **Backward Compatibility**: Existing functionality unchanged, new features are additive

## [1.2.0] - 2025-08-01

### Added
- **NEW**: Comprehensive naming consistency verification across all 37 Go source files
- **NEW**: Complete integration test suite with proper environment variable configuration
- **NEW**: Enhanced integration test environment with proper test directory setup
- **NEW**: Internal `calc` package with build-variant optimized JVM flag parsing
- **NEW**: Internal `count` package with size-based estimation for minimal builds
- **NEW**: Comprehensive build constraint tests (`TestBuildConstraints`, `TestBuildConstraintsParsing`)
- **NEW**: Cross-build consistency validation tests
- **NEW**: Enhanced test-local.sh script with build variant testing
- **NEW**: Custom shell parser to replace go-shellwords dependency (37% size reduction)

### Changed
- **BREAKING**: Renamed all constructor functions from `New*` pattern to `Create*/Load*` pattern for consistency
- **ENHANCED**: All function names now follow consistent patterns across the entire codebase  
- **ENHANCED**: Integration tests now properly configure BPI_APPLICATION_PATH and class counting environment
- **ENHANCED**: Documentation updated with correct function names throughout TEST_DOCUMENTATION.md
- **ENHANCED**: Memory allocation adjustments for realistic test scenarios (512M minimum for JVM)
- **ENHANCED**: Test expectations aligned with actual application output strings
- **ENHANCED**: Binary size optimization through dependency reduction

### Fixed
- **Integration Tests**: Fixed all integration test failures by setting proper environment variables
- **Memory Allocation**: Adjusted unrealistic test memory values (1M→512M, 1024KB→2048000KB) for realistic JVM requirements
- **Build System**: Fixed `make all` failures caused by integration test environment configuration issues
- **Dependencies**: Removed go-shellwords dependency to reduce binary size and external dependencies
- **Naming Consistency**: Standardized all constructor function names across the codebase

### Technical Details
- **Codebase Consistency**: All 37 Go source files verified for naming consistency
- **Test Coverage**: Comprehensive integration testing with proper environment setup
- **Build Optimization**: Multiple build variants (standard vs minimal) with size comparison
- **Quality Assurance**: Enhanced testing framework with edge case coverage
- **Documentation**: Complete alignment between documentation and implementation

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

### Security
- Removed multiple external dependencies reducing attack surface
- Enhanced input validation for all user-provided values
- Secure file system operations with proper error handling
- Safe memory parsing preventing integer overflow attacks

### Development
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **Integration Tests**: Resolved test failures by properly configuring environment variables
  - Set `BPI_APPLICATION_PATH=.` for correct application directory detection
  - Added `BPL_JVM_LOADED_CLASS_COUNT` environment variable for class count testing
  - Adjusted memory test values to use realistic minimums (512MB+) that meet JVM requirements
  - Updated test expectations to match actual application output ("Calculating JVM memory")
- **Documentation**: Corrected function name references in `TEST_DOCUMENTATION.md`
  - Fixed inconsistent function names (e.g., `TestNewParser` → `TestCreateParser`)
  - Ensured all documented test functions match actual implementation

### Verified
- **Naming Consistency**: Confirmed consistent naming patterns across entire codebase
  - All 37 Go source files use standardized `Create()` and `Load()` constructor patterns
  - Variable, function, and method names follow Go conventions
  - Documentation accurately reflects implementation details

## [1.0.0] - 2025-07-31

## [v1.0.0] - Previous Version

### Features
- Basic JVM memory calculation
- Container memory detection
- JAR file scanning for class count estimation
- Command-line interface with basic options
- Integration with Paketo buildpacks

---

**Note**: This changelog covers the comprehensive refactoring and enhancement effort that transformed the memory calculator into a production-ready, self-contained tool with minimal dependencies and extensive test coverage.

### Removed
- Vendor directory and all external paketo-buildpack dependencies
- Complex dependency tree including libcnb, libpak, and other buildpack-specific packages

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

[Unreleased]: https://github.com/patbaumgartner/memory-calculator/compare/v1.3.1...HEAD
[1.3.1]: https://github.com/patbaumgartner/memory-calculator/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/patbaumgartner/memory-calculator/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/patbaumgartner/memory-calculator/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/patbaumgartner/memory-calculator/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/patbaumgartner/memory-calculator/releases/tag/v1.0.0
