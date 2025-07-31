// Package display handles output formatting and result display.
package display

import (
	"fmt"
	"strings"

	"github.com/patbaumgartner/memory-calculator/internal/config"
	"github.com/patbaumgartner/memory-calculator/internal/memory"
)

// Formatter handles output formatting for the memory calculator.
type Formatter struct {
	parser *memory.Parser
}

// NewFormatter creates a new display formatter.
func NewFormatter() *Formatter {
	return &Formatter{
		parser: memory.NewParser(),
	}
}

// DisplayResults shows the calculated JVM settings in a formatted way.
func (f *Formatter) DisplayResults(props map[string]string, totalMemory int64, cfg *config.Config) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("JVM Memory Configuration")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("Total Memory:     %s\n", f.parser.FormatMemory(totalMemory))
	fmt.Printf("Thread Count:     %s\n", cfg.ThreadCount)
	fmt.Printf("Loaded Classes:   %s\n", cfg.LoadedClassCount)
	fmt.Printf("Head Room:        %s%%\n", cfg.HeadRoom)

	fmt.Println("\nCalculated JVM Arguments:")
	fmt.Println(strings.Repeat("-", 30))

	// Extract and display key JVM settings
	f.displayJVMSetting(props, "-Xmx", "Max Heap Size:         ")
	f.displayJVMSetting(props, "-Xss", "Thread Stack Size:     ")
	f.displayJVMSetting(props, "-XX:MaxMetaspaceSize", "Max Metaspace Size:    ")
	f.displayJVMSetting(props, "-XX:ReservedCodeCacheSize", "Code Cache Size:       ")
	f.displayJVMSetting(props, "-XX:MaxDirectMemorySize", "Direct Memory Size:    ")

	fmt.Println("\nComplete JVM Options:")
	fmt.Println(strings.Repeat("-", 30))

	javaToolOptions := f.buildJavaToolOptions(props)
	fmt.Printf("JAVA_TOOL_OPTIONS=%s\n", javaToolOptions)
}

// DisplayQuietResults shows only the JVM parameters without formatting.
func (f *Formatter) DisplayQuietResults(props map[string]string) {
	javaToolOptions := f.buildJavaToolOptions(props)
	fmt.Print(javaToolOptions)
}

// DisplayVersion shows version information.
func (f *Formatter) DisplayVersion(cfg *config.Config) {
	fmt.Printf("JVM Memory Calculator\n")
	fmt.Printf("Version: %s\n", cfg.BuildVersion)
	fmt.Printf("Build Time: %s\n", cfg.BuildTime)
	fmt.Printf("Commit: %s\n", cfg.CommitHash)
	fmt.Printf("Go Version: %s\n", "1.24.5")
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
	fmt.Println("  --loaded-class-count string   JVM loaded class count (default \"35000\")")
	fmt.Println("  --head-room string            JVM head room percentage (default \"0\")")
	fmt.Println("  --quiet                       Only output JVM parameters, no formatting")
	fmt.Println("  --version                     Show version information")
	fmt.Println("  --help                        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  memory-calculator")
	fmt.Println("  memory-calculator --thread-count=300 --head-room=10")
	fmt.Println("  memory-calculator --total-memory=2G")
	fmt.Println("  memory-calculator --total-memory=512M")
	fmt.Println("  memory-calculator --total-memory=2147483648")
	fmt.Println("  memory-calculator --quiet --total-memory=2G  # Only output JVM parameters")
}

// displayJVMSetting extracts and displays a specific JVM setting.
func (f *Formatter) displayJVMSetting(props map[string]string, flag, label string) {
	// First check if it exists as an individual key
	if value, exists := props[flag]; exists {
		fmt.Printf("%s%s\n", label, value)
		return
	}

	// If not found individually, try to extract from JAVA_TOOL_OPTIONS
	if javaToolOptions, exists := props["JAVA_TOOL_OPTIONS"]; exists {
		value := f.extractJVMFlag(javaToolOptions, flag)
		if value != "" {
			fmt.Printf("%s%s\n", label, value)
		}
	}
}

// extractJVMFlag extracts a specific JVM flag value from a JAVA_TOOL_OPTIONS string.
func (f *Formatter) extractJVMFlag(javaToolOptions, flag string) string {
	parts := strings.Fields(javaToolOptions)

	for _, part := range parts {
		if strings.HasPrefix(part, flag) {
			// Handle flags like -Xmx512M or -XX:MaxMetaspaceSize=128M
			if strings.Contains(part, "=") {
				// Format: -XX:MaxMetaspaceSize=128M
				if split := strings.SplitN(part, "=", 2); len(split) == 2 {
					return split[1]
				}
			} else {
				// Format: -Xmx512M
				return strings.TrimPrefix(part, flag)
			}
		}
	}
	return ""
}

// buildJavaToolOptions constructs the JAVA_TOOL_OPTIONS string from properties.
func (f *Formatter) buildJavaToolOptions(props map[string]string) string {
	// Display JAVA_TOOL_OPTIONS if it exists
	if javaToolOptions, exists := props["JAVA_TOOL_OPTIONS"]; exists {
		return javaToolOptions
	}

	// If JAVA_TOOL_OPTIONS doesn't exist, build it from individual flags
	var options []string
	for flag, value := range props {
		if flag != "JAVA_TOOL_OPTIONS" {
			options = append(options, fmt.Sprintf("%s%s", flag, value))
		}
	}

	return strings.Join(options, " ")
}
