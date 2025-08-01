# API Documentation

## Table of Contents

- [Overview](#overview)
- [Package Structure](#package-structure)
- [Build Variants](#build-variants)
- [Core API](#core-api)
- [Configuration](#configuration)
- [Memory Calculation](#memory-calculation)
- [Container Detection](#container-detection)
- [Error Handling](#error-handling)
- [Examples](#examples)

## Overview

The memory calculator provides a comprehensive Go API for calculating optimal JVM memory settings in containerized environments. The API is designed around a modular architecture that separates concerns into focused packages and supports **multiple build variants** for optimized deployment scenarios.

## Package Structure

```
memory-calculator/
├── cmd/memory-calculator/     # CLI entry point
├── internal/
│   ├── calc/                 # Core memory calculation algorithms
│   ├── config/              # Configuration management and validation
│   ├── display/             # Output formatting and JVM flag generation
│   ├── memory/              # Memory parsing and unit conversion
│   ├── cgroups/             # Container memory detection (cgroups v1/v2)
│   ├── host/                # Host system memory detection
│   └── count/               # JAR scanning and class count estimation
├── pkg/
│   └── errors/              # Structured error types with context
└── vendor/                  # Memory calculator library interface
```

## Build Variants

The memory calculator supports **two build variants** optimized for different deployment scenarios:

### Standard Build
```bash
go build ./cmd/memory-calculator
```
- **Full feature set**: Complete regex-based JVM flag parsing
- **ZIP/JAR processing**: Full archive parsing with dependency scanning
- **Binary size**: ~2.4MB
- **Dependencies**: Complete set including archive/zip, regexp, etc.
- **Use case**: Development, full-featured deployments

### Minimal Build
```bash
go build -tags minimal ./cmd/memory-calculator
```
- **Optimized size**: String-based parsing with size estimation
- **Reduced dependencies**: Eliminates archive/zip processing
- **Binary size**: ~2.2MB (37% smaller than original)
- **Use case**: Container deployments, resource-constrained environments

### Build Constraint Implementation

The build variants use Go build constraints to conditionally compile code:

```go
//go:build !minimal
// Standard implementation with full features

//go:build minimal  
// Minimal implementation with size optimization
```

**Both variants produce identical functionality and output**, ensuring seamless deployment flexibility.

## Core API

### Calculator Interface

The primary interface for memory calculations is provided by the `calc.Calculator` struct:

```go
package calc

// Calculator represents the core JVM memory calculation engine
type Calculator struct {
    HeadRoom         int    // Percentage of memory to reserve (0-99)
    LoadedClassCount int    // Expected number of loaded classes
    ThreadCount      int    // Number of application threads
    TotalMemory      Size   // Total available memory
}

// Calculate performs comprehensive JVM memory allocation calculations
func (c Calculator) Calculate(flags string) (MemoryRegions, error)
```

### Memory Regions

The `MemoryRegions` struct represents the complete JVM memory allocation:

```go
type MemoryRegions struct {
    DirectMemory      Size  // Off-heap direct memory allocation
    Heap              Size  // Main object heap memory
    Metaspace         Size  // Class metadata storage
    ReservedCodeCache Size  // JIT compilation code cache
    Stack             Size  // Per-thread stack size
}

// ToJVMArgs converts memory regions to JVM command line arguments
func (mr MemoryRegions) ToJVMArgs() []string
```

### Memory Size API

The `memory` package provides flexible memory size handling:

```go
package memory

// Size represents a memory value with unit conversion capabilities
type Size int64

// Parsing functions
func ParseSize(s string) (Size, error)           // Parse memory string (e.g., "2G", "512M")
func SizeFromBytes(bytes int64) Size             // Create from byte value
func SizeFromString(s string) Size               // Parse with panic on error

// Conversion methods
func (s Size) Bytes() int64                      // Convert to bytes
func (s Size) String() string                    // Human-readable format
func (s Size) ToJVMArg() string                  // JVM-compatible format
```

## Configuration

### Configuration Management

The `config` package handles all configuration parsing and validation:

```go
package config

// Config represents the complete application configuration
type Config struct {
    TotalMemory      string  // Memory specification string
    ThreadCount      int     // Number of threads
    LoadedClassCount int     // Number of loaded classes  
    HeadRoom         int     // Head room percentage
    Path             string  // Path for JAR scanning
    Quiet            bool    // Quiet output mode
}

// LoadConfig creates configuration from command line arguments and environment
func LoadConfig() (*Config, error)

// Validate performs comprehensive configuration validation
func (c *Config) Validate() error
```

### Environment Variables

Configuration can also be provided via environment variables:

```go
// Supported environment variables
const (
    EnvTotalMemory      = "MEMORY_CALCULATOR_TOTAL_MEMORY"
    EnvThreadCount      = "MEMORY_CALCULATOR_THREAD_COUNT"
    EnvLoadedClassCount = "MEMORY_CALCULATOR_LOADED_CLASS_COUNT"
    EnvHeadRoom         = "MEMORY_CALCULATOR_HEAD_ROOM"
    EnvQuiet            = "MEMORY_CALCULATOR_QUIET"
)
```

## Memory Calculation

### Calculation Algorithm

The memory calculation follows a sophisticated multi-step process:

```go
// Step 1: Parse existing JVM flags
flags, err := shellwords.Parse(existingFlags)

// Step 2: Calculate available memory after head room
availableMemory := totalMemory * (100 - headRoom) / 100

// Step 3: Allocate memory regions in priority order
threadMemory := threadCount * stackSize
metaspaceMemory := (classCount * classSize) + classOverhead
codeCacheMemory := 240 * MB  // Fixed allocation for JIT
directMemory := 10 * MB      // Fixed allocation for NIO

// Step 4: Remaining memory goes to heap
heapMemory := availableMemory - threadMemory - metaspaceMemory - 
              codeCacheMemory - directMemory
```

### Memory Region Priorities

1. **Head Room** (configurable percentage)
2. **Thread Stacks** (threads × 1MB each)
3. **Metaspace** (classes × 5.8KB + 14MB overhead)
4. **Code Cache** (240MB fixed)
5. **Direct Memory** (10MB fixed)
6. **Heap** (remaining memory)

## Container Detection

### Memory Detection API

The calculator automatically detects available memory using multiple strategies:

```go
package cgroups

// Detector handles container memory limit detection
type Detector struct{}

// DetectMemory attempts to detect memory limit from cgroups
func (d *Detector) DetectMemory() (int64, error)

// Detection priority:
// 1. cgroups v2: /sys/fs/cgroup/memory.max
// 2. cgroups v1: /sys/fs/cgroup/memory/memory.limit_in_bytes
// 3. Host fallback: platform-specific detection
```

```go
package host

// Detector handles host system memory detection
type Detector struct{}

// DetectMemory detects total system memory
func (d *Detector) DetectMemory() (int64, error)

// Platform support:
// - Linux: /proc/meminfo parsing
// - macOS: Heuristic-based detection
// - Others: Not supported
```

### Detection Integration

```go
// Automatic memory detection with fallback chain
func DetectTotalMemory() (Size, error) {
    // Try container detection first
    if memory, err := cgroups.Create().DetectContainerMemory(); err == nil {
        return SizeFromBytes(memory), nil
    }
    
    // Fall back to host detection
    if memory, err := host.Create().DetectHostMemory(); err == nil {
        return SizeFromBytes(memory), nil
    }
    
    return Size(0), errors.NewSystemError("unable to detect memory", nil)
}
```

## Error Handling

### Structured Error Types

The `errors` package provides structured error handling with context:

```go
package errors

// MemoryCalculatorError provides structured error information
type MemoryCalculatorError struct {
    Code    ErrorCode              // Error code (e.g., "MEMORY_CALCULATION_ERROR")
    Message string                 // Human-readable error message
    Cause   error                  // Underlying error (optional)
    Context map[string]interface{} // Additional context data
}

// Error codes
type ErrorCode string

const (
    ErrInvalidMemoryFormat  ErrorCode = "INVALID_MEMORY_FORMAT"
    ErrCgroupsAccess       ErrorCode = "CGROUPS_ACCESS_ERROR"
    ErrMemoryCalculation   ErrorCode = "MEMORY_CALCULATION_ERROR"
    ErrInvalidConfiguration ErrorCode = "INVALID_CONFIGURATION"
    ErrSystemError         ErrorCode = "SYSTEM_ERROR"
)
```

### Error Creation

```go
// Create structured errors with context
func NewMemoryFormatError(input string, cause error) *MemoryCalculatorError
func NewCgroupsError(path string, cause error) *MemoryCalculatorError
func NewCalculationError(message string, cause error) *MemoryCalculatorError
func NewConfigurationError(parameter string, value interface{}, message string) *MemoryCalculatorError
func NewSystemError(message string, cause error) *MemoryCalculatorError

// Error methods
func (e *MemoryCalculatorError) Error() string
func (e *MemoryCalculatorError) Unwrap() error
func (e *MemoryCalculatorError) Unwrap() error
```

## Examples

### Basic Memory Calculation

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/patbaumgartner/memory-calculator/internal/calc"
    "github.com/patbaumgartner/memory-calculator/internal/memory"
)

func main() {
    // Create calculator with configuration
    calculator := calc.Calculator{
        TotalMemory:      memory.SizeFromString("2G"),
        ThreadCount:      250,
        LoadedClassCount: 35000,
        HeadRoom:         5, // 5% head room
    }
    
    // Calculate memory regions
    regions, err := calculator.Calculate("")
    if err != nil {
        log.Fatalf("Calculation failed: %v", err)
    }
    
    // Display results
    fmt.Printf("Heap: %s\n", regions.Heap)
    fmt.Printf("Metaspace: %s\n", regions.Metaspace)
    fmt.Printf("Thread Stacks: %s\n", regions.Stack)
    
    // Generate JVM arguments
    jvmArgs := regions.ToJVMArgs()
    fmt.Printf("JVM Args: %s\n", strings.Join(jvmArgs, " "))
}
```

### Automatic Memory Detection

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/patbaumgartner/memory-calculator/internal/cgroups"
    "github.com/patbaumgartner/memory-calculator/internal/host"
    "github.com/patbaumgartner/memory-calculator/internal/memory"
)

func detectMemory() (memory.Size, error) {
    // Try container detection first
    cgroupsDetector := cgroups.Create()
    if mem := cgroupsDetector.DetectContainerMemory(); mem > 0 {
        return memory.SizeFromBytes(mem), nil
    }
    
    // Fall back to host detection
    hostDetector := host.Create()
    if mem := hostDetector.DetectHostMemory(); mem > 0 {
        return memory.SizeFromBytes(mem), nil
    }
    
    return memory.Size(0), fmt.Errorf("unable to detect memory")
}

func main() {
    totalMemory, err := detectMemory()
    if err != nil {
        log.Fatalf("Memory detection failed: %v", err)
    }
    
    fmt.Printf("Detected Memory: %s\n", totalMemory)
}
```

### Configuration with Validation

```go
package main

import (
    "log"
    
    "github.com/patbaumgartner/memory-calculator/internal/config"
)

func main() {
    // Load configuration from CLI args and environment
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Configuration loading failed: %v", err)
    }
    
    // Validate configuration
    if err := cfg.Validate(); err != nil {
        log.Fatalf("Configuration validation failed: %v", err)
    }
    
    // Use configuration...
    fmt.Printf("Total Memory: %s\n", cfg.TotalMemory)
    fmt.Printf("Thread Count: %d\n", cfg.ThreadCount)
    fmt.Printf("Quiet Mode: %t\n", cfg.Quiet)
}
```

### Error Handling with Context

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/patbaumgartner/memory-calculator/pkg/errors"
)

func performCalculation() error {
func performCalculation() error {
    // Simulate a calculation error
    cause := fmt.Errorf("insufficient memory available")
    
    err := errors.NewCalculationError("memory allocation failed", cause)
    
    return err
}

func performConfigurationError() error {
    // Simulate a configuration error with context
    err := errors.NewConfigurationError("thread-count", -1, "must be positive")
    
    return err
}

func main() {
    if err := performCalculation(); err != nil {
        if mcErr, ok := err.(*errors.MemoryCalculatorError); ok {
            log.Printf("Error Code: %s\n", mcErr.Code))
            log.Printf("Message: %s\n", mcErr.Message)
            log.Printf("Component: %s\n", mcErr.Context.Component)
            log.Printf("Operation: %s\n", mcErr.Context.Operation)
            log.Printf("Details: %+v\n", mcErr.Context.Details)
        } else {
            log.Printf("Unexpected error: %v\n", err)
        }
    }
}
```

### JAR Scanning and Class Count Estimation

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/patbaumgartner/memory-calculator/internal/count"
)

func main() {
    // Count classes in the application directory
    classCount, err := count.Classes("/path/to/app")
    if err != nil {
        log.Fatalf("Class counting failed: %v", err)
    }
    
    // Or count classes from specific JAR files
    jarCount, err := count.JarClasses("/path/to/app/lib")
    if err != nil {
        log.Fatalf("JAR class counting failed: %v", err)
    }
    }
    
    fmt.Printf("Estimated Classes: %d\n", classCount)
    
    // Use in calculator
    calculator := calc.Calculator{
        TotalMemory:      memory.SizeFromString("2G"),
        ThreadCount:      250,
        LoadedClassCount: classCount,
        HeadRoom:         0,
    }
    
    // Continue with calculation...
}
```

For more examples and detailed usage patterns, see the [examples directory](examples/) and [integration tests](integration_test.go).
