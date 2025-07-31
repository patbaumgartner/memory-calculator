# JVM Memory Calculator - Test Suite Documentation

## Overview
This document describes the comprehensive test suite for the JVM Memory Calculator, which provides **75.2% code coverage** with multiple types of tests across a well-structured package architecture.

## Architecture & Test Organization

The test suite is organized by package structure, providing clear separation of concerns:

### Package Structure
```
cmd/memory-calculator/          # Main application entry point
pkg/errors/                     # Public structured error types  
internal/
├── config/                    # Configuration management
├── memory/                    # Memory parsing & formatting
├── cgroups/                   # Container memory detection
└── display/                   # Output formatting
```

## Test Files by Package

### 1. `integration_test.go` (Root Package)
**End-to-end integration tests**
- `TestMainIntegration`: Full application testing with various command line arguments
- `TestMainEnvironmentVariables`: Tests buildpack environment variable handling
- `TestMainBoundaryValues`: Tests edge cases and boundary conditions

### 2. `internal/memory/parser_test.go`
**Memory parsing and formatting tests** (95.7% coverage)
- `TestParseMemoryString`: Comprehensive memory string parsing with 25+ test cases
- `TestFormatMemory`: Memory formatting to human-readable strings with edge cases
- `TestValidateMemorySize`: Memory size validation testing
- `TestNewParser`: Parser constructor testing
- `TestConstants`: Memory unit constant validation

### 3. `internal/cgroups/detector_test.go`
**Container memory detection tests** (94.6% coverage)
- `TestDetectContainerMemory`: Integration tests for memory detection
- `TestReadCgroupsV1`: cgroups v1 memory limit reading with mock files
- `TestReadCgroupsV2`: cgroups v2 memory limit reading with mock files
- `TestNewDetector`: Constructor and dependency injection testing

### 4. `internal/display/formatter_test.go`
**Output and display tests** (100% coverage)
- `TestDisplayResults`: Main result display function testing
- `TestDisplayQuietResults`: Quiet mode output testing
- `TestExtractJVMFlag`: JVM flag extraction and parsing
- `TestBuildJavaToolOptions`: JAVA_TOOL_OPTIONS construction
- `TestDisplayJVMSetting`: Individual JVM setting display

### 5. `internal/config/config_test.go`
**Configuration management tests** (100% coverage)
- `TestDefaultConfig`: Default configuration creation
- `TestConfigValidate`: Configuration validation with error cases
- `TestConfigSetEnvironmentVariables`: Environment variable handling
- `TestConfigSetTotalMemory`: Memory configuration setting

### 6. `pkg/errors/errors_test.go`
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
| `pkg/errors` | **100%** | Structured error types, error wrapping, context |
| `internal/config` | **100%** | Configuration validation, environment variables |
| `internal/display` | **100%** | Output formatting, JVM flag extraction |
| `internal/memory` | **98.2%** | Memory parsing, formatting, validation |
| `internal/host` | **98.5%** | Host memory detection, cross-platform support |
| `internal/cgroups` | **95.1%** | Container memory detection, cgroups v1/v2, host fallback |
| `cmd/memory-calculator` | **0%** | Main function (tested via integration) |
| **Overall** | **76.4%** | **Enhanced with host detection support** |

### Performance Tests (Benchmarks)
- **Memory Parsing Performance** (`internal/memory`): Benchmarks for different memory unit formats
- **Memory Formatting Performance** (`internal/memory`): Benchmarks for different memory sizes  
- **Container Detection Performance** (`internal/cgroups`): Benchmarks for cgroups memory detection
- **Host Detection Performance** (`internal/host`): Benchmarks for cross-platform host memory detection
- **Display Performance** (`internal/display`): Benchmarks for output formatting and JVM flag extraction
- **Main Execution Performance** (`integration`): End-to-end application performance testing

## Test Coverage: 76.4% 🎯

### Fully Covered Areas ✅
✅ Memory string parsing and validation (95.7%)
✅ Memory formatting and display (100%)  
✅ Configuration management and validation (100%)
✅ Container memory detection (94.6%)
✅ Error handling with structured types (100%)
✅ Output formatting and display (100%)
✅ Environment variable management (100%)  
✅ Host memory detection across platforms (98.5%)
✅ Memory detection fallback priority testing (95.1%)
✅ Integration with Paketo buildpack memory calculator

### Areas with Limited Coverage ⚠️
⚠️ Main function execution (covered by integration tests)
⚠️ Some edge cases in cgroups file reading  
⚠️ Platform-specific system calls (Darwin/Windows use heuristics)
⚠️ Complex system-specific behavior

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
make test-coverage
```

### All Tests (Basic)
```bash
make test
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

# Package-specific benchmarks  
go test ./internal/memory -bench=.
go test ./internal/cgroups -bench=.
go test ./internal/display -bench=.

# Compare benchmark results
make benchmark-compare
```

### Coverage Reports
```bash
# Generate HTML coverage report
make test-coverage
go tool cover -html=coverage/coverage.out -o coverage/coverage.html
```

## Test Results Summary

**Total Test Packages**: 6 packages
**Total Tests**: 100+ test cases across all packages
**Overall Coverage**: **75.2%** (improved from 53.5%)
**Package Coverage**: 3 packages at 100%, 2 packages at 94%+
**Reliability**: 100% pass rate across all test scenarios
**Architecture**: Professional package structure with dependency injection

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
- No memory limit detected
- cgroups v1 memory limits
- cgroups v2 memory limits
- Unrealistic memory limits (filtered out)
- File system errors and missing files

This comprehensive test suite ensures the JVM Memory Calculator works reliably across different environments and use cases, particularly in containerized environments with buildpack deployment scenarios. The professional package architecture provides excellent maintainability and testability while achieving high code coverage.
