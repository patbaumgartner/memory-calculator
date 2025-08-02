# JVM Memory Calculator - API Reference

Complete Go API for calculating optimal JVM memory settings in containerized environments.

## 📚 Quick Navigation

- [🚀 Quick Start](#-quick-start) - Get started immediately
- [📦 Package Overview](#-package-overview) - Architecture and responsibilities  
- [🧮 Core API](#-core-api) - Essential functions and types
- [💾 Memory Management](#-memory-management) - Size handling and operations
- [� Examples](#-examples) - Real-world usage patterns

*For detailed examples and integration patterns, see [USAGE_GUIDE.md](USAGE_GUIDE.md)*

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/patbaumgartner/memory-calculator/internal/calculator"
)

func main() {
    // Create memory calculator
    mc := calculator.Create(false) // false = verbose output
    
    // Execute calculation with automatic memory detection
    result, err := mc.Execute()
    if err != nil {
        log.Fatalf("Calculation failed: %v", err)
    }
    
    // Result contains JAVA_TOOL_OPTIONS
    fmt.Printf("JVM Options: %s\n", result["JAVA_TOOL_OPTIONS"])
}
```

### Advanced Configuration

```go
import (
    "github.com/patbaumgartner/memory-calculator/internal/calc"
    "github.com/patbaumgartner/memory-calculator/internal/memory"
)

// Direct calculator usage with custom configuration
calculator := calc.Calculator{
    TotalMemory:      memory.SizeFromString("4G"),
    ThreadCount:      300,
    LoadedClassCount: 50000,
    HeadRoom:         10, // 10% safety margin
}

regions, err := calculator.Calculate("-XX:MaxMetaspaceSize=512m")
if err != nil {
    log.Fatal(err)
}

// Access individual memory regions
fmt.Printf("Heap: %s\n", regions.Heap)
fmt.Printf("Metaspace: %s\n", regions.Metaspace)
fmt.Printf("Thread Stacks: %s\n", regions.Stack)
```

## 📦 Package Overview

### Architecture

```
memory-calculator/
├── cmd/memory-calculator/       # CLI application entry point
├── internal/                   # Private application packages
│   ├── calc/                  # Core memory calculation algorithms
│   ├── calculator/            # High-level calculator orchestration
│   ├── cgroups/              # Container memory detection (cgroups v1/v2)
│   ├── config/               # Configuration management and validation
│   ├── constants/            # Memory unit constants and defaults
│   ├── count/                # Class counting and JAR analysis
│   ├── display/              # Output formatting and presentation
│   ├── host/                 # Host system memory detection
│   ├── logger/               # Structured logging utilities
│   ├── memory/               # Memory parsing and unit conversion
│   └── parser/               # String parsing utilities
└── pkg/                      # Public packages
    └── errors/               # Structured error types
```

### Package Responsibilities

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `calc` | Core memory calculation algorithms | `Calculator`, `MemoryRegions` |
| `calculator` | High-level orchestration and integration | `MemoryCalculator` |
| `cgroups` | Container memory limit detection | `Detector` |
| `config` | Configuration management and validation | `Config` |
| `constants` | Memory unit constants and defaults | Constants |
| `count` | Class counting and JAR analysis | Class counting functions |
| `display` | Output formatting and presentation | `Formatter` |
| `host` | Host system memory detection | `Detector` |
| `logger` | Structured logging | `Logger` |
| `memory` | Memory parsing and unit conversion | `Size` |
| `parser` | String parsing utilities | Parsing functions |
| `errors` | Structured error handling | `MemoryCalculatorError` |

## 🏗️ Build Variants

The memory calculator supports **two optimized build variants** for different deployment scenarios:

### Standard Build (Full Features)

```bash
go build ./cmd/memory-calculator
# OR
make build
```

**Features:**
- Complete regex-based JVM flag parsing
- Full ZIP/JAR archive processing with dependency scanning
- Advanced class counting with metadata extraction
- Binary size: ~2.4MB
- Dependencies: Complete feature set including `archive/zip`, `regexp`

**Use Cases:**
- Development environments
- Full-featured deployments
- CI/CD pipelines with comprehensive analysis

### Minimal Build (Size Optimized)

```bash
go build -tags minimal ./cmd/memory-calculator
# OR 
make build-minimal
```

**Features:**
- String-based parsing (no regex dependency)
- Size-based class estimation (no ZIP processing)
- Reduced binary size: ~2.2MB (8% smaller)
- Minimal dependencies: eliminates `archive/zip` processing

**Use Cases:**
- Container deployments
- Resource-constrained environments
- Embedded systems and edge computing

### Build Constraint Implementation

```go
//go:build !minimal
// +build !minimal
// Standard implementation - full features
func matchHeap(s string) bool {
    return HeapPattern.MatchString(s)
}

//go:build minimal
// +build minimal  
// Minimal implementation - string matching
func matchHeap(s string) bool {
    return strings.HasPrefix(s, "-Xmx")
}
```

**Important:** Both variants produce **identical output and functionality**, ensuring seamless deployment flexibility.

## 🧮 Core Calculation API

### Calculator Interface

The primary calculation engine implementing sophisticated JVM memory allocation algorithms:

```go
package calc

// Calculator performs JVM memory allocation calculations
type Calculator struct {
    TotalMemory      Size  // Total available memory
    ThreadCount      int   // Number of application threads  
    LoadedClassCount int   // Expected loaded classes
    HeadRoom         int   // Safety margin percentage (0-99)
}

// Calculate performs comprehensive memory allocation
func (c Calculator) Calculate(flags string) (MemoryRegions, error)
```

### Memory Regions

Complete JVM memory allocation specification:

```go
type MemoryRegions struct {
    Heap              *Heap              // Main object storage
    Metaspace         *Metaspace         // Class metadata storage  
    Stack             Stack              // Per-thread stack allocation
    DirectMemory      DirectMemory       // Off-heap NIO memory
    ReservedCodeCache ReservedCodeCache  // JIT compilation cache
    HeadRoom          *HeadRoom          // Safety margin reservation
}

// Convert to JVM command line arguments
func (mr MemoryRegions) ToJVMArgs() []string

// Calculate total allocated memory
func (mr MemoryRegions) TotalSize() Size

// Get allocation summary
func (mr MemoryRegions) Summary() string
```

### Memory Allocation Algorithm

The calculator implements a sophisticated 7-step allocation process:

```go
// 1. Parse existing JVM flags for user overrides
flags, err := parser.ParseJVMFlags(existingFlags)

// 2. Calculate head room reservation (percentage-based)
headRoom := totalMemory * (headRoomPercent / 100)
availableMemory := totalMemory - headRoom

// 3. Allocate thread stack memory (threads × stack size)
stackMemory := threadCount * defaultStackSize // 1MB per thread

// 4. Calculate metaspace requirements
metaspaceMemory := (loadedClasses * classOverhead) + baseMetaspaceSize

// 5. Reserve code cache for JIT compilation
codeCacheMemory := 240 * MB // Optimized for performance

// 6. Reserve direct memory for NIO operations  
directMemory := 10 * MB // Conservative allocation

// 7. Allocate remaining memory to heap
heapMemory := availableMemory - stackMemory - metaspaceMemory - 
              codeCacheMemory - directMemory
```

### Constants and Defaults

```go
const (
    // Memory calculation constants
    ClassSize         = 5_800         // Bytes per loaded class
    ClassOverhead     = 14_000_000    // Base metaspace overhead
    DefaultStackSize  = 1 * MB        // Per-thread stack size
    DefaultCodeCache  = 240 * MB      // JIT compilation cache
    DefaultDirectMem  = 10 * MB       // NIO operations
    
    // Memory unit constants
    KB = 1024
    MB = 1024 * KB  
    GB = 1024 * MB
    TB = 1024 * GB
)
```

## 💾 Memory Management

### Size Type and Operations

Flexible memory size handling with unit conversion:

```go
package memory

// Size represents a memory value with provenance tracking
type Size struct {
    Value      int64      // Memory size in bytes
    Provenance Provenance // Source of this value
}

type Provenance int
const (
    Calculated     Provenance = iota // Calculated by algorithm
    UserConfigured                   // Specified by user
    DefaultValue                     // System default
)

// Creation functions
func SizeFromBytes(bytes int64) Size
func SizeFromString(s string) Size        // Panics on error
func ParseSize(s string) (Size, error)    // Safe parsing

// Conversion methods  
func (s Size) Bytes() int64              // Raw byte value
func (s Size) KB() float64               // Kilobytes  
func (s Size) MB() float64               // Megabytes
func (s Size) GB() float64               // Gigabytes
func (s Size) String() string            // Human-readable (e.g., "2G")
func (s Size) ToJVMArg() string         // JVM format (e.g., "2048m")

// Arithmetic operations
func (s Size) Add(other Size) Size
func (s Size) Sub(other Size) Size  
func (s Size) Mul(factor float64) Size
func (s Size) Div(divisor float64) Size

// Comparison operations
func (s Size) LessThan(other Size) bool
func (s Size) GreaterThan(other Size) bool
func (s Size) Equals(other Size) bool
```

### Memory Unit Parsing

Supports flexible memory unit formats:

```go
// Supported formats
sizes := []string{
    "2G", "2GB", "2g", "2gb",           // Gigabytes
    "512M", "512MB", "512m", "512mb",   // Megabytes  
    "1024K", "1024KB", "1024k",         // Kilobytes
    "2147483648B", "2147483648",        // Bytes
    "1.5G", "512.5M",                   // Decimal values
}

for _, s := range sizes {
    size, err := memory.ParseSize(s)
    if err != nil {
        log.Printf("Failed to parse %s: %v", s, err)
        continue
    }
    fmt.Printf("%s = %d bytes\n", s, size.Bytes())
}
```

---

For complete examples and integration patterns, see the [examples directory](examples/) and [integration tests](integration_test.go).

## 🔗 Related Documentation

- [Architecture Overview](ARCHITECTURE.md) - System design and architecture
- [Project Setup](PROJECT_SETUP.md) - Development and build information
- [Usage Guide](USAGE_GUIDE.md) - Detailed usage examples
- [Test Documentation](TEST_DOCUMENTATION.md) - Testing framework and coverage

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
    Path             string  // Path for JAR scanning and class counting
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
    EnvTotalMemory      = "BPL_JVM_TOTAL_MEMORY"
    EnvThreadCount      = "BPL_JVM_THREAD_COUNT"
    EnvLoadedClassCount = "BPL_JVM_LOADED_CLASS_COUNT"
    EnvHeadRoom         = "BPL_JVM_HEAD_ROOM"
    EnvApplicationPath  = "BPI_APPLICATION_PATH"
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
