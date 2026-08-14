// Package display handles output formatting and result display.
package display

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/patbaumgartner/memory-calculator/internal/calc"
	"github.com/patbaumgartner/memory-calculator/internal/calculator"
	"github.com/patbaumgartner/memory-calculator/internal/config"
	"github.com/patbaumgartner/memory-calculator/internal/memory"
)

// Formatter handles output formatting for the memory calculator.
type Formatter struct {
	parser *memory.Parser
}

// CreateFormatter creates a new display formatter.
func CreateFormatter() *Formatter {
	return &Formatter{
		parser: memory.CreateParser(),
	}
}

// DisplayResults shows the calculated JVM settings in a formatted way.
func (f *Formatter) DisplayResults(result calculator.Result, cfg *config.Config) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("JVM Memory Configuration")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("Total Memory:     %s\n", f.parser.FormatMemory(result.TotalMemory.Value))
	fmt.Printf("Thread Count:     %d\n", result.ThreadCount)
	fmt.Printf("Loaded Classes:   %d\n", result.LoadedClassCount)
	fmt.Printf("Head Room:        %d%%\n", result.HeadRoom)
	fmt.Printf("Application Path: %s\n", cfg.Path)

	fmt.Println("\nCalculated JVM Arguments:")
	fmt.Println(strings.Repeat("-", 30))

	regions := result.Regions
	for _, setting := range []struct{ label, value string }{
		{"Max Heap Size:         ", sizeString(heapValue(regions))},
		{"Thread Stack Size:     ", sizeString(regions.Stack.Value)},
		{"Max Metaspace Size:    ", sizeString(metaspaceValue(regions))},
		{"Code Cache Size:       ", sizeString(regions.ReservedCodeCache.Value)},
		{"Direct Memory Size:    ", sizeString(regions.DirectMemory.Value)},
	} {
		fmt.Printf("%s%s\n", setting.label, setting.value)
	}

	fmt.Println("\nComplete JVM Options:")
	fmt.Println(strings.Repeat("-", 30))
	fmt.Printf("JAVA_TOOL_OPTIONS=%s\n", result.JavaToolOptions)
}

func sizeString(value int64) string {
	return calc.Size{Value: value}.String()
}

func heapValue(regions calc.MemoryRegions) int64 {
	if regions.Heap == nil {
		return 0
	}

	return regions.Heap.Value
}

func metaspaceValue(regions calc.MemoryRegions) int64 {
	if regions.Metaspace == nil {
		return 0
	}

	return regions.Metaspace.Value
}

// DisplayQuietResults shows only the JVM parameters without formatting.
func (f *Formatter) DisplayQuietResults(result calculator.Result) {
	fmt.Print(result.JavaToolOptions)
}

// DisplayVersion shows version information.
func (f *Formatter) DisplayVersion(cfg *config.Config) {
	fmt.Printf("JVM Memory Calculator\n")
	fmt.Printf("Version: %s\n", cfg.BuildVersion)
	fmt.Printf("Build Time: %s\n", cfg.BuildTime)
	fmt.Printf("Commit: %s\n", cfg.CommitHash)
	fmt.Printf("Go Version: %s\n", runtime.Version())
}

// DisplayHelp shows help information on stdout, for an explicit --help.
func (f *Formatter) DisplayHelp(cfg *config.Config) {
	f.WriteHelp(os.Stdout, cfg)
}

// WriteHelp renders the help text to w.
//
// A usage error must render to stderr rather than stdout: callers capture stdout with
// `$(memory-calculator --quiet)`, and help text written there would be substituted into
// JAVA_TOOL_OPTIONS as if it were JVM options.
func (f *Formatter) WriteHelp(w io.Writer, cfg *config.Config) {
	_, _ = fmt.Fprintf(w, helpText, cfg.BuildVersion)
}

const helpText = `JVM Memory Calculator
====================
Version: %s

Calculates JVM memory settings based on container memory limits.
Automatically detects memory from cgroups v1/v2.

Usage:
  memory-calculator [flags]

Flags:
  --total-memory string         Total memory (e.g., 2G, 512M, 1024MB)
  --thread-count string         JVM thread count (default "250")
  --loaded-class-count string   JVM loaded class count (calculated if not set)
  --head-room string            JVM head room percentage (default "0")
  --path string                 Application path for JAR scanning (default "/app")
  --quiet                       Only output JVM parameters, no formatting
  --version                     Show version information
  -h, --help                    Show this help message

Examples:
  memory-calculator
  memory-calculator --thread-count=300 --head-room=10
  memory-calculator --total-memory=2G
  memory-calculator --total-memory=512M
  memory-calculator --path=/my/app --total-memory=2G
  memory-calculator --quiet --total-memory=2G  # Only output JVM parameters
`
