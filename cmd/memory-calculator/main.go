package main

import (
	"flag"
	"log"
	"os"

	"github.com/paketo-buildpacks/libjvm/helper"
	"github.com/patbaumgartner/memory-calculator/internal/cgroups"
	"github.com/patbaumgartner/memory-calculator/internal/config"
	"github.com/patbaumgartner/memory-calculator/internal/display"
	"github.com/patbaumgartner/memory-calculator/internal/memory"
	"github.com/patbaumgartner/memory-calculator/pkg/errors"
)

// Build information (set by ldflags during build)
var (
	version    = "dev"
	buildTime  = "unknown"
	commitHash = "unknown"
)

func main() {
	cfg := config.DefaultConfig()
	cfg.BuildVersion = version
	cfg.BuildTime = buildTime
	cfg.CommitHash = commitHash

	// Parse command line flags
	flag.StringVar(&cfg.TotalMemory, "total-memory", "", "Total memory (e.g., 2G, 512M, 1024MB, 2147483648)")
	flag.StringVar(&cfg.ThreadCount, "thread-count", cfg.ThreadCount, "JVM thread count")
	flag.StringVar(&cfg.LoadedClassCount, "loaded-class-count", cfg.LoadedClassCount, "JVM loaded class count")
	flag.StringVar(&cfg.HeadRoom, "head-room", cfg.HeadRoom, "JVM head room percentage")
	flag.BoolVar(&cfg.Quiet, "quiet", false, "Only output JVM parameters, no formatting")
	flag.BoolVar(&cfg.Version, "version", false, "Show version information")
	flag.BoolVar(&cfg.Help, "help", false, "Show help")

	flag.Parse()

	formatter := display.NewFormatter()

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
		if !cfg.Quiet {
			log.Printf("Configuration error: %v", err)
		}
		os.Exit(1)
	}

	// Initialize components
	memoryParser := memory.NewParser()
	cgroupsDetector := cgroups.NewDetector()

	// Detect container memory from cgroups
	containerMemory := cgroupsDetector.DetectContainerMemory()

	// Determine final memory to use
	var finalMemory int64
	if cfg.TotalMemory != "" {
		parsed, err := memoryParser.ParseMemoryString(cfg.TotalMemory)
		if err != nil {
			if !cfg.Quiet {
				if mcErr, ok := err.(*errors.MemoryCalculatorError); ok {
					log.Printf("Invalid total-memory value: %v, using detected memory", mcErr)
				} else {
					log.Printf("Invalid total-memory value: %v, using detected memory", err)
				}
			}
			finalMemory = containerMemory
		} else {
			finalMemory = parsed
			if !cfg.Quiet {
				log.Printf("Using specified memory: %s", memoryParser.FormatMemory(finalMemory))
			}
		}
	} else {
		finalMemory = containerMemory
	}

	if !cfg.Quiet {
		if finalMemory > 0 {
			log.Printf("Container memory detected: %s", memoryParser.FormatMemory(finalMemory))
		} else {
			log.Println("No memory limit detected, using system defaults")
		}
	}

	// Set environment variables for memory calculator
	cfg.SetEnvironmentVariables()
	cfg.SetTotalMemory(finalMemory)

	// Execute memory calculator
	mc := helper.MemoryCalculator{}
	props, err := mc.Execute()
	if err != nil {
		mcErr := errors.NewCalculationError("Memory calculation failed", err)
		if !cfg.Quiet {
			log.Printf("Error: %v", mcErr)
		}
		os.Exit(1)
	}

	// Display results
	if cfg.Quiet {
		formatter.DisplayQuietResults(props)
	} else {
		formatter.DisplayResults(props, finalMemory, cfg)
	}
}
