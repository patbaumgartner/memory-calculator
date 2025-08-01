// Package memory provides comprehensive memory size handling, parsing, and formatting utilities.
//
// This package offers flexible memory size parsing with support for decimal values and multiple
// unit formats, efficient memory size conversion and formatting, and JVM-compatible memory
// argument generation. It serves as the foundation for all memory-related operations in the
// memory calculator.
//
// Key Features:
//   - Flexible unit parsing: B, K, KB, M, MB, G, GB, T, TB (case-insensitive)
//   - Decimal value support: 1.5G, 2.25GB, 512.5M with proper precision handling
//   - Binary-based calculations: All units use powers of 1024 (not 1000)
//   - JVM compatibility: Generates memory arguments compatible with all JVM versions
//   - Human-readable formatting: Automatically selects appropriate units for display
//   - Validation and error handling: Comprehensive input validation with detailed error messages
//
// Memory Size Calculation:
//   - Bytes (B): Base unit, direct byte values
//   - Kilobytes (K/KB): 1024 bytes
//   - Megabytes (M/MB): 1024² bytes (1,048,576 bytes)
//   - Gigabytes (G/GB): 1024³ bytes (1,073,741,824 bytes)
//   - Terabytes (T/TB): 1024⁴ bytes (1,099,511,627,776 bytes)
//
// Supported Input Formats:
//   - Numeric only: "1073741824" (interpreted as bytes)
//   - With units: "1G", "1GB", "1.5g", "2.25GB", "512m"
//   - Case-insensitive: "1g", "1GB", "1Gb" all equivalent
//   - Decimal precision: Up to 2 decimal places for fractional values
//
// Usage Examples:
//
//	// Parse memory strings
//	size, err := ParseSize("2G")          // 2,147,483,648 bytes
//	size, err := ParseSize("1.5GB")       // 1,610,612,736 bytes
//	size, err := ParseSize("512M")        // 536,870,912 bytes
//
//	// Create from bytes
//	size := SizeFromBytes(1073741824)     // 1GB
//
//	// Format for display
//	fmt.Println(size.String())            // "1.00 GB"
//
//	// Generate JVM arguments
//	fmt.Println(size.ToJVMArg())          // "1048576K"
//
// The Size type provides thread-safe operations and immutable semantics, making it suitable
// for concurrent use across multiple goroutines without synchronization requirements.
package memory

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/patbaumgartner/memory-calculator/pkg/errors"
)

const (
	// Binary-based memory size constants using powers of 1024
	// These constants follow the traditional binary interpretation used by most
	// operating systems and JVMs, where each unit is 1024 times the previous unit.

	// KB represents one kilobyte in binary notation (1024 bytes)
	KB = 1024

	// MB represents one megabyte in binary notation (1,048,576 bytes)
	MB = KB * 1024

	// GB represents one gigabyte in binary notation (1,073,741,824 bytes)
	GB = MB * 1024
	TB = GB * 1024

	// Maximum supported memory size (1PB to prevent overflow)
	MaxMemorySize = 1024 * TB
)

// Parser handles memory string parsing and formatting.
type Parser struct{}

// CreateParser creates a new memory parser.
func CreateParser() *Parser {
	return &Parser{}
}

// ParseMemoryString parses memory strings with units (e.g., "2G", "512M", "1024MB") to bytes.
// Supported units: B, K, KB, M, MB, G, GB, T, TB (case insensitive).
// Decimal values are supported (e.g., "1.5G").
// Returns the memory in bytes and an error if the format is invalid.
func (p *Parser) ParseMemoryString(memStr string) (int64, error) {
	if memStr == "" {
		return 0, errors.NewMemoryFormatError(memStr, fmt.Errorf("empty memory string"))
	}

	memStr = strings.TrimSpace(strings.ToUpper(memStr))

	// Handle plain number (bytes)
	if bytes, err := p.parseAsBytes(memStr); err == nil {
		return bytes, nil
	}

	// Parse with unit
	return p.parseWithUnit(memStr)
}

// parseAsBytes attempts to parse a string as plain bytes (no unit)
func (p *Parser) parseAsBytes(memStr string) (int64, error) {
	num, err := strconv.ParseInt(memStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return p.validateAndReturnBytes(num, memStr)
}

// parseWithUnit parses memory string with unit suffix
func (p *Parser) parseWithUnit(memStr string) (int64, error) {
	numStr, unit := p.extractNumberAndUnit(memStr)

	if numStr == "" {
		return 0, errors.NewMemoryFormatError(memStr, fmt.Errorf("no numeric value found"))
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, errors.NewMemoryFormatError(memStr, fmt.Errorf("invalid numeric value: %s", numStr))
	}

	if num < 0 {
		return 0, errors.NewMemoryFormatError(memStr, fmt.Errorf("negative memory size not allowed"))
	}

	bytes := p.convertToBytes(num, unit)
	if bytes < 0 {
		return 0, errors.NewMemoryFormatError(memStr, fmt.Errorf("unsupported unit: %s", unit))
	}

	return p.validateAndReturnBytes(bytes, memStr)
}

// extractNumberAndUnit separates numeric part from unit part
func (p *Parser) extractNumberAndUnit(memStr string) (string, string) {
	var numStr string
	var unit string

	for i, r := range memStr {
		if (r >= '0' && r <= '9') || r == '.' {
			numStr += string(r)
		} else {
			unit = memStr[i:]
			break
		}
	}

	return numStr, unit
}

// convertToBytes converts number with unit to bytes, returns -1 for invalid unit
func (p *Parser) convertToBytes(num float64, unit string) int64 {
	switch unit {
	case "B", "":
		return int64(num)
	case "K", "KB":
		return int64(num * KB)
	case "M", "MB":
		return int64(num * MB)
	case "G", "GB":
		return int64(num * GB)
	case "T", "TB":
		return int64(num * TB)
	default:
		return -1 // Invalid unit
	}
}

// validateAndReturnBytes validates size limits and returns bytes
func (p *Parser) validateAndReturnBytes(bytes int64, originalStr string) (int64, error) {
	if bytes < 0 {
		return 0, errors.NewMemoryFormatError(originalStr, fmt.Errorf("negative memory size not allowed"))
	}
	if bytes > MaxMemorySize {
		return 0, errors.NewMemoryFormatError(originalStr, fmt.Errorf("memory size exceeds maximum supported size"))
	}
	return bytes, nil
}

// FormatMemory formats bytes to human readable format.
// Returns "Unknown" for zero or negative values.
func (p *Parser) FormatMemory(bytes int64) string {
	if bytes <= 0 {
		return "Unknown"
	}

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

// ValidateMemorySize checks if a memory size is within acceptable bounds.
func (p *Parser) ValidateMemorySize(bytes int64) error {
	if bytes < 0 {
		return errors.NewMemoryFormatError(fmt.Sprintf("%d", bytes), fmt.Errorf("negative memory size not allowed"))
	}
	if bytes > MaxMemorySize {
		return errors.NewMemoryFormatError(fmt.Sprintf("%d", bytes), fmt.Errorf("memory size exceeds maximum supported size"))
	}
	return nil
}
