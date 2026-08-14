// Package main provides the memory-calculator CLI tool for calculating optimal JVM memory settings.
//
// The memory calculator is designed for containerized Java applications and provides automatic
// memory detection from cgroups v1/v2 with intelligent host system fallback. It calculates
// optimal JVM memory allocation including heap, metaspace, thread stacks, code cache, and
// direct memory based on available system resources.
//
// Features:
//   - Smart container detection (cgroups v1/v2 + host fallback)
//   - Paketo buildpack integration (Temurin, Liberica)
//   - Flexible memory unit support (B, K, KB, M, MB, G, GB, T, TB)
//   - Quiet mode for scripting and automation
//   - Comprehensive error handling and validation
//
// Basic usage:
//
//	memory-calculator --total-memory 2G --thread-count 300
//	memory-calculator --quiet  # outputs only JVM arguments
//
// The calculator automatically detects available memory using this priority:
//  1. Container cgroups v2: /sys/fs/cgroup/memory.max
//  2. Container cgroups v1: /sys/fs/cgroup/memory/memory.limit_in_bytes
//  3. Host system memory: platform-specific detection
//
// Memory allocation algorithm:
//  1. Head room reservation (configurable percentage)
//  2. Thread stacks (threads × 1MB each)
//  3. Metaspace (loaded classes × 8KB each)
//  4. Code cache (240MB for JIT compilation)
//  5. Direct memory (10MB for NIO operations)
//  6. Heap memory (remaining available memory)
//
// Environment integration:
//   - JAVA_TOOL_OPTIONS: automatically configured with calculated JVM arguments
//   - Paketo buildpack variables: BPL_JVM_THREAD_COUNT, BPL_JVM_HEAD_ROOM, etc.
//   - Container orchestration: Docker, Kubernetes, Cloud Foundry
//
// For detailed documentation and examples, see: https://github.com/patbaumgartner/memory-calculator
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/patbaumgartner/memory-calculator/internal/calculator"
	"github.com/patbaumgartner/memory-calculator/internal/config"
	"github.com/patbaumgartner/memory-calculator/internal/display"
	"github.com/patbaumgartner/memory-calculator/pkg/errors"
)

// Build information (set by ldflags during build)
var (
	version    = "dev"
	buildTime  = "unknown"
	commitHash = "unknown"
)

func main() {
	cfg := config.Load()
	cfg.BuildVersion = version
	cfg.BuildTime = buildTime
	cfg.CommitHash = commitHash

	// Parse command line flags
	flag.StringVar(&cfg.TotalMemory, "total-memory", cfg.TotalMemory,
		"Total memory (e.g., 2G, 512M, 1024MB, 2147483648)")
	flag.StringVar(&cfg.ThreadCount, "thread-count", cfg.ThreadCount, "JVM thread count")
	flag.StringVar(&cfg.LoadedClassCount, "loaded-class-count", cfg.LoadedClassCount, "JVM loaded class count")
	flag.StringVar(&cfg.HeadRoom, "head-room", cfg.HeadRoom, "JVM head room percentage")
	flag.StringVar(&cfg.Path, "path", cfg.Path, "Application path for JAR scanning and class counting")
	flag.BoolVar(&cfg.Quiet, "quiet", false, "Only output JVM parameters, no formatting")
	flag.BoolVar(&cfg.Version, "version", false, "Show version information")
	flag.BoolVar(&cfg.Help, "help", false, "Show help")

	flag.Parse()

	formatter := display.CreateFormatter()

	if cfg.Version {
		formatter.DisplayVersion(cfg)
		return
	}

	if cfg.Help {
		formatter.DisplayHelp(cfg)
		return
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fail(err)
	}

	// Execute memory calculator
	mc := calculator.Create(cfg.Quiet)
	result, err := mc.Execute(cfg.Input())
	if err != nil {
		fail(errors.NewCalculationError("memory calculation failed", err))
	}

	// Display results
	displayResults(formatter, result, cfg)
}

// fail reports an error and exits non-zero.
//
// Diagnostics always go to stderr, including under --quiet: that flag keeps stdout clean for
// `$(memory-calculator --quiet)` capture, and silently exiting non-zero would leave a caller with
// empty JVM options and no explanation.
func fail(err error) {
	fmt.Fprintf(os.Stderr, "memory-calculator: %v\n", err)
	os.Exit(1)
}

// displayResults displays the calculation results based on quiet flag
func displayResults(formatter *display.Formatter, result calculator.Result, cfg *config.Config) {
	if cfg.Quiet {
		formatter.DisplayQuietResults(result)
	} else {
		formatter.DisplayResults(result, cfg)
	}
}
