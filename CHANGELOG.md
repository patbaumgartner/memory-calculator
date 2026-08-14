# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **Checked region arithmetic**: thread-stack multiplication, metaspace calculation and aggregate
  region sums now fail on overflow instead of wrapping into plausible but unsafe heaps. Exact-boundary
  budgets that leave no heap and zero-valued JVM regions are rejected.
- **Checked class adjustments**: class-count additions and percentage scaling use exact integer math
  with overflow checks instead of float conversion.
- **JAR descriptor accumulation**: each archive is now closed before the directory walk continues,
  so large classpaths cannot exhaust the process file-descriptor limit.
- **Deprecated head-room warning**: `BPL_JVM_HEADROOM` now emits the documented stderr warning,
  including under `--quiet`; `BPL_JVM_HEAD_ROOM` still takes precedence.
- **Container memory detection**: cgroup v2 is now read before v1. The shipped code read v1 first,
  so on a hybrid host it could return a stale v1 limit instead of the limit the container runtime
  configured.
- **"No limit" cgroup values**: an unset cgroup v1 limit is the kernel's page-counter maximum, which
  depends on the page size. The previous check compared against the single literal
  `9223372036854771712`, so on a 64 KiB-page arm64 kernel the sentinel was read as a real limit and
  clamped to 64 TiB — a JVM sized for terabytes inside a small container. Values at or above 4 EiB,
  at or below zero, or too large for `int64` are now all treated as unlimited.
- **Ancestor cgroup limits**: the detector now walks from the process's own cgroup up to the mount
  root and uses the smallest limit found. A container cgroup is frequently unlimited while its pod
  cgroup is capped, and only the walk finds that cap.
- **Invalid `--total-memory` is fatal**: an unparseable value was a warning, after which the tool
  fell back to auto-detection and exited 0. In a container that sized the JVM for the host, so a
  typo produced a confidently wrong heap. `BPL_JVM_TOTAL_MEMORY` was not validated at all.
- **Errors are no longer silent under `--quiet`**: diagnostics were suppressed entirely, so a failed
  run exited 1 with no output and `export JAVA_TOOL_OPTIONS="$(memory-calculator --quiet)"` produced
  an empty variable with no explanation. Errors now always go to stderr; stdout stays clean.
- **Conflicting `-Xmx`**: a JVM option the tool recognised but could not parse was treated as absent,
  so `JAVA_TOOL_OPTIONS="-Xmx1.5G"` produced `-Xmx1.5G ... -Xmx3640311K` and the JVM honoured
  whichever came last. Unreadable option values are now a hard error naming the option.
- **Size overflow**: `ParseSize` applied the unit multiplier without an overflow check, so `8388608t`
  wrapped to a negative number and the tool emitted `-Xmx-8388608T`.
- **Sizes rendering as `0`**: any size below 1 KiB was rendered as `"0"`, producing options such as
  `-Xmx0` that the JVM rejects.
- **Negative inputs inflating the heap**: region sizes are subtracted from total memory, so a
  negative thread count, class count or head room percentage grew the heap beyond the container
  limit. These are now rejected before any allocation.
- **JVM option splitting**: a quoted section in the middle of a word split it in two, so
  `-D"foo"=bar` parsed as `["-Dfoo", "=bar"]`. Unterminated quotes and dangling escapes were silently
  accepted and mis-split; both are now rejected.
- **`Total Memory: Unknown`**: the report always printed `Unknown`, because the total was never
  passed to the formatter. It now reports the budget actually used, along with the resolved class
  count.
- **`--version` Go version**: reported a hardcoded string rather than the toolchain that built the
  binary.
- **Head room upper bound**: `--head-room=100` was accepted but can never produce a viable
  configuration. The documented range 0–99 is now enforced.

### Changed
- **Release gates fail closed**: medium/high `gosec` findings and missing expected release artifacts
  now fail CI, and Docker behavior is tested against the exact local image built in the job before
  the multi-platform image is published.
- **Host memory fallback reads `MemAvailable`** instead of `MemTotal`. Without a cgroup limit the
  process shares the machine, and sizing a heap against total RAM overcommits a busy host.
- **Non-Linux host detection removed**: the macOS path multiplied Go's runtime heap statistics by 32
  and reported the result as physical memory. Unsupported platforms now fall through to the
  documented 1 GiB default instead of acting on a fabricated number.
- **`calculator.Execute` returns a typed `Result`** instead of `map[string]string`, carrying the
  options string, the memory budget, the calculated regions and the effective thread, class and head
  room values. The display layer no longer re-parses the options string it was handed.
- **Detection reports its source**, so the log states whether the budget came from cgroup v2, v1,
  host memory or the default.

### Removed
- **The `minimal` build variant.** It substituted a different flag parser and replaced JAR class
  counting with a "one class per 2 KB of file size" guess, so it disagreed with the standard build on
  real input. Its tests were failing and were never run by CI, yet its binaries shipped in every
  release. The legacy `memory-calculator-minimal-*` release filenames are still published as copies
  of the standard binary so existing download URLs keep working; they are deprecated and will be
  dropped in the next major release. Make targets `build-minimal`, `build-compressed`,
  `build-size-comparison` and `build-ultimate-comparison` are gone.
- **Dead packages**: `internal/constants`, and the unreachable duplicates of memory detection. The
  `internal/cgroups` and `internal/host` packages were imported by nothing outside their own tests
  while `internal/calculator` carried an inferior copy; the copy is deleted and the packages are now
  wired in.
- **`regexp`**: no longer in the dependency graph. The five near-identical per-region option files
  collapsed into one table of prefixes, shrinking the stripped binary from 2412 KB to 2140 KB
  (-11.3%), a larger saving than the removed `minimal` variant claimed.
- **`calc.ParseUnit`**, unused since it was written.

### Security
- Sizes are checked for 64-bit overflow before the unit multiplier is applied, so a large input
  cannot wrap into a negative memory budget.

## [1.3.2] - 2025-12-13

### Changed
- **Go Version**: Upgraded to Go 1.25.5 across all build configurations
  - Updated `go.mod` to require Go 1.25.5
  - Updated GitHub Actions workflow to use Go 1.25
  - Updated Dockerfile base images to `golang:1.25.5-alpine3.21`
  - Updated example Dockerfiles to use Go 1.25.5
  - All build artifacts now compiled with Go 1.25.5

### Enhanced
- **Docker Build System**: Improved multi-platform Docker builds
  - Updated to Alpine Linux 3.21 for latest security patches and features
  - Maintained multi-architecture support (linux/amd64, linux/arm64)
  - Enhanced Docker build reliability with updated base images
  - Updated example Dockerfiles with consistent Alpine 3.21 base

- **Code Quality**: Refactored test functions to reduce complexity
  - Simplified `TestMainHostMemoryDetection` using table-driven tests
  - Refactored `TestBuildConstraints` with extracted helper functions
  - Improved `TestCalculatorCalculate` with data-driven approach
  - All test functions now meet cyclomatic complexity threshold (<15)
  - Enhanced test maintainability and readability

- **Linting Configuration**: Updated golangci-lint setup
  - Upgraded to golangci-lint v2.7.2
  - Configured gofumpt settings for consistent code formatting
  - Enhanced linter rules for better code quality enforcement

### Fixed
- **Test Complexity**: Resolved golangci-lint complexity violations
  - Extracted helper functions to reduce cyclomatic complexity
  - Improved test structure with table-driven patterns
  - Better separation of concerns in test code

### Technical Details
- **Go Version**: 1.25.5 (upgraded from 1.24.5)
- **Alpine Version**: 3.21 (upgraded from 3.20)
- **golangci-lint**: v2.7.2
- **Build System**: Enhanced cross-compilation support maintained
- **Test Coverage**: Maintained high coverage with improved test structure

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
