// Package display handles output formatting and result display.
package display

import (
	"fmt"
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

// DisplayHelp shows help information.
func (f *Formatter) DisplayHelp(cfg *config.Config) {
	fmt.Println("JVM Memory Calculator")
	fmt.Println("====================")
	fmt.Printf("Version: %s\n", cfg.BuildVersion)
	fmt.Println()
	fmt.Println("Calculates JVM memory settings based on container memory limits.")
	fmt.Println("Automatically detects memory from cgroups v1/v2.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  memory-calculator [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --total-memory string         Total memory (e.g., 2G, 512M, 1024MB)")
	fmt.Println("  --thread-count string         JVM thread count (default \"250\")")
	fmt.Println("  --loaded-class-count string   JVM loaded class count (calculated if not set)")
	fmt.Println("  --head-room string            JVM head room percentage (default \"0\")")
	fmt.Println("  --path string                 Application path for JAR scanning (default \"/app\")")
	fmt.Println("  --quiet                       Only output JVM parameters, no formatting")
	fmt.Println("  --version                     Show version information")
	fmt.Println("  --help                        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  memory-calculator")
	fmt.Println("  memory-calculator --thread-count=300 --head-room=10")
	fmt.Println("  memory-calculator --total-memory=2G")
	fmt.Println("  memory-calculator --total-memory=512M")
	fmt.Println("  memory-calculator --path=/my/app --total-memory=2G")
	fmt.Println("  memory-calculator --quiet --total-memory=2G  # Only output JVM parameters")
}
