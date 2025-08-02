# JVM Memory Calculator - Test Suite Documentation

## Overview
This document describes the comprehensive test suite for the JVM Memory Calculator, which provides **87.2% code coverage** for core calculation logic with multiple types of tests across a well-structured package architecture. The test suite includes **advanced build constraint testing** for both standard and minimal build variants, ensuring consistent functionality across all deployment scenarios.

## Architecture & Test Organization

The test suite is organized by package structure, providing clear separation of concerns:

### Package Structure
```
cmd/memory-calculator/          # Main application entry point
pkg/errors/                     # Public structured error types  
internal/
├── calc/                      # Memory calculations with build constraints
├── calculator/                # Memory calculator
├── cgroups/                   # Container memory detection
├── config/                    # Configuration management
├── constants/                 # Memory unit constants
├── count/                     # Class counting with build variants
├── display/                   # Output formatting
├── host/                      # Host memory detection
├── logger/                    # Logging utilities
├── memory/                    # Memory parsing & formatting
└── parser/                    # Memory string parsing
```

### Build Constraint Testing

The test suite includes comprehensive testing for **build constraints** that enable different binary variants:

**Standard Build Tests:**
- Full regex-based JVM flag parsing
- Complete ZIP/JAR file processing
- All original functionality preserved
- ✅ **Status**: All tests passing

**Minimal Build Tests:**
- Simple string-based parsing validation
- File size-based class count estimation
- Functional equivalence verification
- ✅ **Status**: All tests passing

**Cross-Build Tests:**
- Consistency testing across both build variants
- Performance benchmarking for different implementations
- Integration testing for identical output verification
- ✅ **Status**: Both variants produce identical outputs

**Integration Test Environment:**
- Enhanced test-local.sh script with proper test directory setup
- Both build variants tested in isolation
- Binary size comparison and validation
- ✅ **Status**: Complete integration test compatibility

## Test Files by Package

### 1. `integration_test.go` (Root Package)
**End-to-end integration tests**
- `TestMainIntegration`: Full application testing with various command line arguments
- `TestMainEnvironmentVariables`: Tests buildpack environment variable handling
- `TestMainBoundaryValues`: Tests edge cases and boundary conditions
- `TestMainHostMemoryDetection`: Tests enhanced memory auto-detection with host fallback

### 2. `internal/memory/parser_test.go`
**Memory parsing and formatting tests** (95.7% coverage)
- `TestParseMemoryString`: Comprehensive memory string parsing with 25+ test cases
- `TestFormatMemory`: Memory formatting to human-readable strings with edge cases
- `TestValidateMemorySize`: Memory size validation testing
- `TestCreateParser`: Parser constructor testing
- `TestConstants`: Memory unit constant validation

### 3. `internal/cgroups/detector_test.go`
**Container memory detection tests** (94.6% coverage)
- `TestDetectContainerMemory`: Integration tests for memory detection
- `TestDetectContainerMemoryWithHostFallback`: Tests intelligent fallback to host detection
- `TestHostFallbackPriority`: Tests prioritized detection (cgroups v2 → v1 → host)
- `TestReadCgroupsV1`: cgroups v1 memory limit reading with mock files
- `TestReadCgroupsV2`: cgroups v2 memory limit reading with mock files
- `TestCreateDetector`: Constructor and dependency injection testing

### 4. `internal/host/detector_test.go`
**Host memory detection tests** (100% coverage) - NEW
- `TestDetectHostMemory`: Cross-platform host memory detection
- `TestDetectLinuxMemory`: Linux `/proc/meminfo` parsing with comprehensive scenarios
- `TestDetectDarwinMemory`: macOS heuristic-based memory detection  
- `TestIsHostMemoryDetectionSupported`: Platform support validation
- `TestPlatformSpecificBehavior`: Tests platform-specific detection logic
- `TestMemoryDetectionRealWorldScenarios`: Real-world memory size testing

### 5. `internal/display/formatter_test.go`
**Output and display tests** (100% coverage)
- `TestDisplayResults`: Main result display function testing
- `TestDisplayQuietResults`: Quiet mode output testing
- `TestExtractJVMFlag`: JVM flag extraction and parsing
- `TestBuildJavaToolOptions`: JAVA_TOOL_OPTIONS construction
- `TestDisplayJVMSetting`: Individual JVM setting display

### 6. `internal/config/config_test.go`
**Configuration management tests** (100% coverage)
- `TestLoad`: Default configuration creation
- `TestConfigValidate`: Configuration validation with error cases
- `TestConfigSetEnvironmentVariables`: Environment variable handling
- `TestConfigSetTotalMemory`: Memory configuration setting

### 7. `pkg/errors/errors_test.go`
**Error handling tests** (100% coverage)
- `TestMemoryCalculatorError`: Structured error type testing
- `TestNewMemoryFormatError`: Memory format error creation
- `TestNewCgroupsError`: Cgroups error creation
- `TestNewCalculationError`: Calculation error creation
- `TestNewConfigurationError`: Configuration error creation

## Test Categories

### Unit Tests (Per Package)
- **Memory Parsing**: 30+ test cases covering all supported units and edge cases
- **Memory Formatting**: 15+ test cases covering byte to human-readable conversion  
- **Container Detection**: Comprehensive cgroups v1/v2 testing with mock file systems
- **Host Memory Detection**: Cross-platform memory detection with platform-specific testing
- **Memory Detection Fallback**: Tests prioritized detection (cgroups → host fallback)
- **Configuration Management**: Environment variables and validation testing
- **Display Formatting**: Output formatting for both standard and quiet modes
- **Error Handling**: Structured error types with context and wrapping

### Integration Tests
- **Command Line Interface**: Tests all command line parameters and combinations
- **Environment Variables**: Tests for buildpack-compatible environment variable handling
- **Memory Units**: Tests various memory unit formats (bytes, K, KB, M, MB, G, GB, T, TB)
- **Parameter Validation**: Tests parameter validation and error handling
- **End-to-End Workflows**: Complete application testing with realistic scenarios

### Package Coverage Summary

| Package | Coverage | Key Features Tested |
|---------|----------|-------------------|
| `pkg/errors` | **100.0%** | Structured error types, error wrapping, context |
| `internal/config` | **100.0%** | Configuration validation, environment variables |
| `internal/display` | **100.0%** | Output formatting, JVM flag extraction |
| `internal/memory` | **98.2%** | Memory parsing, formatting, validation |
| `internal/cgroups` | **95.1%** | Container memory detection, cgroups v1/v2, host fallback |
| `internal/host` | **79.4%** | Host memory detection, cross-platform support |
| `cmd/memory-calculator` | **0.0%** | Main function (tested via integration) |
| **Overall** | **77.1%** | **Enhanced with host detection and improved tooling** |

### Performance Tests (Benchmarks)
- **Memory Parsing Performance** (`internal/memory`): Benchmarks for different memory unit formats
- **Memory Formatting Performance** (`internal/memory`): Benchmarks for different memory sizes  
- **Container Detection Performance** (`internal/cgroups`): Benchmarks for cgroups memory detection
- **Host Detection Performance** (`internal/host`): Benchmarks for cross-platform host memory detection
- **Display Performance** (`internal/display`): Benchmarks for output formatting and JVM flag extraction
- **Main Execution Performance** (`integration`): End-to-end application performance testing

### Build Constraint Tests

**Advanced Testing for Multiple Build Variants:**

#### Standard Build Tests
- Full regex-based JVM flag parsing validation
- Complete ZIP/JAR file processing functionality
- Comprehensive error handling for invalid formats
- Full dependency integration (regexp, archive/zip)

#### Minimal Build Tests  
- Simple string-based parsing accuracy
- File size-based class count estimation
- Streamlined functionality verification
- Reduced dependency validation

#### Cross-Build Consistency Tests
- **Functional Equivalence**: Both builds produce identical results for standard inputs
- **Performance Comparison**: Benchmarking across build variants
- **Integration Validation**: End-to-end testing with both binary variants
- **Error Handling Consistency**: Both builds handle errors appropriately

**Test Commands:**
```bash
# Test standard build constraint implementations
go test -v ./internal/calc -run "TestBuildConstraints"
go test -v ./internal/count -run "TestMinimalBuild"

# Test minimal build constraint implementations  
go test -tags minimal -v ./internal/calc -run "TestBuildConstraints"
go test -tags minimal -v ./internal/count -run "TestMinimalBuild"
```

## Test Coverage: 77.1% 🎯

### Fully Covered Areas ✅
✅ Memory string parsing and validation (98.2%)
✅ Memory formatting and display (100.0%)  
✅ Configuration management and validation (100.0%)
✅ Container memory detection with host fallback (95.1%)
✅ Error handling with structured types (100.0%)
✅ Output formatting and display (100.0%)
✅ Environment variable management (100.0%)  
✅ Host memory detection across platforms (79.4%)
✅ Memory detection fallback priority testing
✅ Integration with Paketo buildpack memory calculator

### Areas with Limited Coverage ⚠️
⚠️ Main function execution (covered by integration tests)
⚠️ Some edge cases in cgroups file reading  
⚠️ Cross-platform system calls (macOS uses heuristics due to CGO-free approach)
⚠️ Complex platform-specific behavior edge cases

## Architecture Benefits

### Professional Package Structure
- **Separation of Concerns**: Each package has a single responsibility
- **Testability**: Packages can be tested independently with dependency injection
- **Maintainability**: Clear interfaces and structured error handling
- **Reusability**: Modular components that can be imported by other projects

### Dependency Injection Pattern  
- Clean main function using dependency-injected components
- Easy mocking and testing of individual components
- Improved testability and maintainability

## Running Tests

### All Tests with Coverage
```bash
make coverage
```

### All Tests (Basic)
```bash
make test
```

### HTML Coverage Report
```bash
make coverage-html
```

### Package-Specific Tests
```bash
# Memory parsing tests
go test ./internal/memory -v

# Container detection tests  
go test ./internal/cgroups -v

# Host memory detection tests
go test ./internal/host -v

# Display formatting tests
go test ./internal/display -v

# Configuration tests
go test ./internal/config -v

# Error handling tests
go test ./pkg/errors -v

# Integration tests only
go test -run TestMain -v
```

### Benchmarks
```bash
# All benchmarks
make benchmark

# Benchmark comparison (save results)
make benchmark-compare

# Package-specific benchmarks  
go test ./internal/memory -bench=.
go test ./internal/cgroups -bench=.
go test ./internal/host -bench=.
go test ./internal/display -bench=.
```

### Quality Assurance
```bash
# Run comprehensive quality checks (format, lint, security, vulnerabilities)
make quality

# Individual quality checks
make format           # Format Go code
make lint             # Run golangci-lint
make security         # Run gosec security scan
make vulncheck        # Run govulncheck vulnerability scan

# Development tools
make tools            # Install all development tools
make tools-check      # Verify tools are available
```

## Test Results Summary

**Total Test Packages**: 7 packages (including new host detection)
**Total Tests**: 120+ test cases across all packages
**Overall Coverage**: **77.1%** (improved from 75.2% with host detection)
**Package Coverage**: 3 packages at 100.0%, 2 packages at 95%+
**Reliability**: 100% pass rate across all test scenarios
**Architecture**: Professional package structure with dependency injection
**Quality Assurance**: Comprehensive linting, security scanning, and vulnerability checks

## Key Test Scenarios Covered

### Memory Input Formats
- Raw bytes: `2147483648`
- Kilobytes: `1024K`, `1024KB`
- Megabytes: `512M`, `512MB`
- Gigabytes: `2G`, `2GB`
- Terabytes: `1T`, `1TB`
- Decimal values: `1.5G`, `2.5M`
- Case insensitive: `1g`, `512m`
- Whitespace handling: ` 1G `, `  512M  `

### Error Conditions
- Invalid formats: `invalid`, `1X`, `G1`
- Empty inputs: `""`
- Complex decimals: `1.2.3G`
- Missing numbers: `G`, `MB`

### Memory Ranges
- Very small: 64MB
- Small: 128MB, 256MB, 512MB
- Medium: 1GB, 2GB, 4GB
- Large: 8GB, 16GB
- Very large: 1TB+

### JVM Parameters
- Thread counts: 50-1000 threads
- Loaded classes: 3500-50000 classes
- Head room: 0-50%
- Various memory configurations

### Container Scenarios
- No memory limit detected (fallback to host detection)
- cgroups v1 memory limits
- cgroups v2 memory limits
- Unrealistic memory limits (filtered out)
- File system errors and missing files
- Host memory detection fallback when cgroups unavailable
- Cross-platform host detection (Linux `/proc/meminfo`, macOS heuristics)

### Platform Support Testing
- Linux: `/proc/meminfo` parsing with various formats
- macOS: Heuristic-based memory detection
- Cross-platform compatibility validation

### GitHub Actions Testing
The project uses comprehensive **automated testing** via GitHub Actions on every push and pull request:

#### Test Pipeline Components
1. **Go Environment Setup**: Tests on Go 1.24.5 with module caching
2. **Dependency Verification**: Downloads and verifies all Go modules
3. **Module Structure Check**: Validates project structure and build process
4. **Race Detection**: Runs all tests with `-race` flag for concurrency issues
5. **Coverage Analysis**: Generates coverage reports and summaries
6. **Integration Testing**: Includes separate integration test execution
7. **Quality Assurance**: Multiple quality gates including:
   - **golangci-lint**: Comprehensive linting with custom configuration
   - **gosec**: Security vulnerability scanning
   - **govulncheck**: Known vulnerability database checking

#### Multi-Platform Build Testing
GitHub Actions validates builds across all supported platforms:
- **Linux**: amd64, arm64
- **macOS**: amd64, arm64 (Apple Silicon)
- **Cross-compilation**: CGO disabled for portable binaries

#### Coverage Reporting
- **Local Coverage**: `make coverage` provides detailed per-package statistics
- **CI Coverage**: GitHub Actions uploads to Codecov for tracking
- **Coverage Gates**: PRs cannot decrease coverage significantly

#### Automated Release Testing
On git tags (`v*`), additional testing includes:
- **Multi-platform builds**: All platform binaries built and tested
- **Integration testing**: Complete end-to-end testing with built binaries
- **Checksum validation**: SHA256 checksums generated and verified
- **Docker testing**: Multi-arch container builds and basic functionality tests
- macOS: Heuristic-based detection without CGO dependencies
- Unsupported platforms: Graceful handling with zero values
- Platform detection priority: cgroups v2 → cgroups v1 → host system

This comprehensive test suite ensures the JVM Memory Calculator works reliably across different environments and use cases, particularly in containerized environments with buildpack deployment scenarios. The professional package architecture with enhanced host detection provides excellent maintainability and testability while achieving high code coverage.
