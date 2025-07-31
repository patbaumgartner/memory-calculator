// Package memory handles memory parsing, formatting, and validation.
package memory

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/patbaumgartner/memory-calculator/pkg/errors"
)

const (
	// Memory size constants in bytes
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024

	// Maximum supported memory size (1PB to prevent overflow)
	MaxMemorySize = 1024 * TB
)

// Parser handles memory string parsing and formatting.
type Parser struct{}

// NewParser creates a new memory parser.
func NewParser() *Parser {
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
