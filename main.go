package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/paketo-buildpacks/libjvm/helper"
)

// Build information (set by ldflags during build)
var (
	version    = "dev"
	buildTime  = "unknown"
	commitHash = "unknown"
)

func main() {
	// Command line flags
	var (
		threadCount      = flag.String("thread-count", "250", "JVM thread count")
		loadedClassCount = flag.String("loaded-class-count", "35000", "JVM loaded class count")
		headRoom         = flag.String("head-room", "0", "JVM head room percentage")
		totalMemory      = flag.String("total-memory", "", "Total memory (e.g., 2G, 512M, 1024MB, 2147483648)")
		quiet            = flag.Bool("quiet", false, "Only output JVM parameters, no formatting")
		versionFlag      = flag.Bool("version", false, "Show version information")
		help             = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	if *versionFlag {
		fmt.Printf("JVM Memory Calculator\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Commit: %s\n", commitHash)
		fmt.Printf("Go Version: %s\n", "1.24.5")
		return
	}

	if *help {
		fmt.Println("JVM Memory Calculator")
		fmt.Println("====================")
		fmt.Printf("Version: %s\n", version)
		fmt.Println()
		fmt.Println("Calculates JVM memory settings based on container memory limits.")
		fmt.Println("Automatically detects memory from cgroups v1/v2.")
		fmt.Println()
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  memory-calculator")
		fmt.Println("  memory-calculator --thread-count=300 --head-room=10")
		fmt.Println("  memory-calculator --total-memory=2G")
		fmt.Println("  memory-calculator --total-memory=512M")
		fmt.Println("  memory-calculator --total-memory=2147483648")
		fmt.Println("  memory-calculator --quiet --total-memory=2G  # Only output JVM parameters")
		return
	}

	// Detect container memory from cgroups
	containerMemory := detectContainerMemory()

	// Determine final memory to use
	var finalMemory int64
	if *totalMemory != "" {
		if parsed, err := parseMemoryString(*totalMemory); err == nil {
			finalMemory = parsed
			if !*quiet {
				fmt.Printf("Using specified memory: %s\n", formatMemory(finalMemory))
			}
		} else {
			if !*quiet {
				log.Printf("Invalid total-memory value: %s, using detected memory", *totalMemory)
			}
			finalMemory = containerMemory
		}
	} else {
		finalMemory = containerMemory
	}

	if !*quiet {
		if finalMemory > 0 {
			fmt.Printf("Container memory detected: %s\n", formatMemory(finalMemory))
		} else {
			fmt.Println("No memory limit detected, using system defaults")
		}
	}

	// Set environment variables for memory calculator
	os.Setenv("BPL_JVM_THREAD_COUNT", *threadCount)
	os.Setenv("BPL_JVM_LOADED_CLASS_COUNT", *loadedClassCount)
	os.Setenv("BPL_JVM_HEAD_ROOM", *headRoom)

	if finalMemory > 0 {
		os.Setenv("BPL_JVM_TOTAL_MEMORY", fmt.Sprintf("%d", finalMemory))
	}

	// Execute memory calculator
	mc := helper.MemoryCalculator{}
	props, err := mc.Execute()
	if err != nil {
		log.Fatalf("Memory calculation failed: %v", err)
	}

	// Display results
	if *quiet {
		displayQuietResults(props)
	} else {
		displayResults(props, finalMemory, *threadCount, *loadedClassCount, *headRoom)
	}
}

// detectContainerMemory attempts to read memory limit from cgroups
func detectContainerMemory() int64 {
	// Try cgroups v2 first
	if memory := readCgroupsV2(); memory > 0 {
		return memory
	}

	// Fall back to cgroups v1
	if memory := readCgroupsV1(); memory > 0 {
		return memory
	}

	return 0
}

// readCgroupsV2 reads memory limit from cgroups v2
func readCgroupsV2() int64 {
	file, err := os.Open("/sys/fs/cgroup/memory.max")
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "max" {
			return 0 // No limit set
		}

		if memory, err := strconv.ParseInt(line, 10, 64); err == nil {
			return memory
		}
	}
	return 0
}

// readCgroupsV1 reads memory limit from cgroups v1
func readCgroupsV1() int64 {
	file, err := os.Open("/sys/fs/cgroup/memory/memory.limit_in_bytes")
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if memory, err := strconv.ParseInt(line, 10, 64); err == nil {
			// Check if it's a realistic limit (not the "no limit" value)
			if memory < 1024*1024*1024*1024 { // Less than 1TB
				return memory
			}
		}
	}
	return 0
}

// parseMemoryString parses memory strings with units (e.g., "2G", "512M", "1024MB")
func parseMemoryString(memStr string) (int64, error) {
	memStr = strings.TrimSpace(strings.ToUpper(memStr))

	// If it's just a number, treat as bytes
	if num, err := strconv.ParseInt(memStr, 10, 64); err == nil {
		return num, nil
	}

	// Extract number and unit
	var numStr string
	var unit string

	for i, r := range memStr {
		if r >= '0' && r <= '9' || r == '.' {
			numStr += string(r)
		} else {
			unit = memStr[i:]
			break
		}
	}

	if numStr == "" {
		return 0, fmt.Errorf("no numeric value found")
	}

	// Parse the numeric part
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", numStr)
	}

	// Convert based on unit
	switch unit {
	case "K", "KB":
		return int64(num * 1024), nil
	case "M", "MB":
		return int64(num * 1024 * 1024), nil
	case "G", "GB":
		return int64(num * 1024 * 1024 * 1024), nil
	case "T", "TB":
		return int64(num * 1024 * 1024 * 1024 * 1024), nil
	case "":
		// No unit, treat as bytes
		return int64(num), nil
	default:
		return 0, fmt.Errorf("unsupported unit: %s", unit)
	}
}

// formatMemory formats bytes to human readable format
func formatMemory(bytes int64) string {
	if bytes <= 0 {
		return "Unknown"
	}

	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.0f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.0f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// displayResults shows the calculated JVM settings
func displayResults(props map[string]string, totalMemory int64, threadCount, classCount, headRoom string) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("JVM Memory Configuration")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("Total Memory:     %s\n", formatMemory(totalMemory))
	fmt.Printf("Thread Count:     %s\n", threadCount)
	fmt.Printf("Loaded Classes:   %s\n", classCount)
	fmt.Printf("Head Room:        %s%%\n", headRoom)

	fmt.Println("\nCalculated JVM Arguments:")
	fmt.Println(strings.Repeat("-", 30))

	// Extract and display key JVM settings
	// First check if they exist as individual keys, otherwise parse from JAVA_TOOL_OPTIONS
	displayJVMSetting(props, "-Xmx", "Max Heap Size:         ")
	displayJVMSetting(props, "-Xss", "Thread Stack Size:     ")
	displayJVMSetting(props, "-XX:MaxMetaspaceSize", "Max Metaspace Size:    ")
	displayJVMSetting(props, "-XX:ReservedCodeCacheSize", "Code Cache Size:       ")
	displayJVMSetting(props, "-XX:MaxDirectMemorySize", "Direct Memory Size:    ")

	fmt.Println("\nComplete JVM Options:")
	fmt.Println(strings.Repeat("-", 30))

	// Display all properties as they would appear in JAVA_TOOL_OPTIONS
	if javaToolOptions, exists := props["JAVA_TOOL_OPTIONS"]; exists {
		fmt.Printf("JAVA_TOOL_OPTIONS=%s\n", javaToolOptions)
	} else {
		// If JAVA_TOOL_OPTIONS doesn't exist, build it from individual flags
		var options []string
		for flag, value := range props {
			if flag != "JAVA_TOOL_OPTIONS" {
				options = append(options, fmt.Sprintf("%s%s", flag, value))
			}
		}
		if len(options) > 0 {
			fmt.Printf("JAVA_TOOL_OPTIONS=%s\n", strings.Join(options, " "))
		}
	}
}

// displayJVMSetting extracts and displays a specific JVM setting
func displayJVMSetting(props map[string]string, flag, label string) {
	// First check if it exists as an individual key
	if value, exists := props[flag]; exists {
		fmt.Printf("%s%s\n", label, value)
		return
	}

	// If not found individually, try to extract from JAVA_TOOL_OPTIONS
	if javaToolOptions, exists := props["JAVA_TOOL_OPTIONS"]; exists {
		value := extractJVMFlag(javaToolOptions, flag)
		if value != "" {
			fmt.Printf("%s%s\n", label, value)
		}
	}
}

// extractJVMFlag extracts a specific JVM flag value from a JAVA_TOOL_OPTIONS string
func extractJVMFlag(javaToolOptions, flag string) string {
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

// displayQuietResults shows only the JVM parameters without formatting
func displayQuietResults(props map[string]string) {
	// Display JAVA_TOOL_OPTIONS if it exists
	if javaToolOptions, exists := props["JAVA_TOOL_OPTIONS"]; exists {
		fmt.Print(javaToolOptions)
	} else {
		// If JAVA_TOOL_OPTIONS doesn't exist, build it from individual flags
		var options []string
		for flag, value := range props {
			if flag != "JAVA_TOOL_OPTIONS" {
				options = append(options, fmt.Sprintf("%s%s", flag, value))
			}
		}
		if len(options) > 0 {
			fmt.Print(strings.Join(options, " "))
		}
	}
}
