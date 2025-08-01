// Package calc provides core memory calculation functionality including size handling,
// memory region allocation, and JVM memory optimization algorithms.
package calc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	// Binary-based memory unit constants following traditional computing conventions
	// where each unit represents powers of 1024 bytes (not 1000).

	// Kibi represents one kibibyte (1024 bytes)
	Kibi = int64(1_024)

	// Mebi represents one mebibyte (1,048,576 bytes)
	Mebi = 1_024 * Kibi

	// Gibi represents one gibibyte (1,073,741,824 bytes)
	Gibi = 1_024 * Mebi

	// Tebi represents one tebibyte (1,099,511,627,776 bytes)
	Tebi = 1_024 * Gibi

	// SizePattern defines the regular expression pattern for parsing memory size strings.
	// Supports numeric values followed by optional unit suffixes (k, m, g, t) in both
	// upper and lower case. Examples: "1024", "512m", "2G", "1.5t"
	SizePattern = "([\\d]+)([kmgtKMGT]?)"
)

// SizeRE is the compiled regular expression for parsing memory size strings
var SizeRE = regexp.MustCompile(fmt.Sprintf("^%s$", SizePattern))

// Provenance indicates the source or origin of a memory size value, providing
// context for how the value was determined and whether it can be overridden.
type Provenance uint8

const (
	// Unknown indicates the provenance of the size value is not known or not tracked
	Unknown Provenance = iota

	// Default indicates the size value comes from system defaults or built-in values
	Default

	// UserConfigured indicates the size value was explicitly set by user configuration
	// such as command-line flags, environment variables, or configuration files
	UserConfigured

	// Calculated indicates the size value was computed by the memory calculator
	// based on available resources and allocation algorithms
	Calculated
)

// Size represents a memory size value with provenance tracking and unit conversion capabilities.
//
// The Size type encapsulates both the numeric memory value and metadata about how that
// value was determined. This allows the memory calculator to make intelligent decisions
// about whether values can be overridden and how to handle conflicts between different
// configuration sources.
//
// Key Features:
//   - Value storage: 64-bit signed integer for memory sizes up to 8 exabytes
//   - Provenance tracking: Origin of the value for configuration precedence
//   - Unit conversion: Automatic parsing and formatting of memory units
//   - JVM compatibility: Generation of JVM-compatible memory arguments
//   - Validation: Range checking and overflow protection
//
// Thread Safety:
//
//	Size instances are immutable after creation and safe for concurrent use.
//	All operations return new Size instances rather than modifying existing ones.
//
// Example Usage:
//
//	// Create from parsed string
//	size, err := NewSizeFromString("2G")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Check provenance
//	if size.Provenance == UserConfigured {
//		fmt.Println("User specified memory size")
//	}
//
//	// Convert to bytes
//	bytes := size.Value
//
//	// Format for display
//	display := size.String()  // "2.00 GB"
//
// Memory Size Limits:
//   - Minimum: 1 byte (though practical minimums are much higher)
//   - Maximum: 2^63-1 bytes (approximately 8 exabytes)
//   - Practical maximum: Limited by available system memory
//
// Precision and Rounding:
//   - All calculations use integer arithmetic to avoid floating-point errors
//   - Fractional units are converted to bytes with truncation (not rounding)
//   - Display formatting uses appropriate precision for the unit magnitude
type Size struct {
	// Value stores the memory size in bytes as a 64-bit signed integer.
	// This provides sufficient range for all practical memory sizes while
	// maintaining precision and avoiding floating-point arithmetic issues.
	Value int64

	// Provenance indicates how this size value was determined, providing context
	// for configuration precedence and override behavior. This allows the
	// calculator to make intelligent decisions about whether values should be
	// preserved or can be overridden by other configuration sources.
	Provenance Provenance
}

// ParseSize parses a memory size in bytes from the given string. Size may include a K, M, G, or T suffix which
// indicates kibibytes, mebibytes, gibibytes or tebibytes respectively.
func ParseSize(s string) (Size, error) {
	t := strings.TrimSpace(s)

	if !SizeRE.MatchString(t) {
		return Size{}, fmt.Errorf("memory size %q does not match pattern %q", t, SizeRE.String())
	}

	groups := SizeRE.FindStringSubmatch(t)
	size, err := strconv.ParseInt(groups[1], 10, 64)
	if err != nil {
		return Size{}, fmt.Errorf("memory size %q is not an integer", groups[1])
	}

	switch strings.ToLower(groups[2]) {
	case "k":
		size *= Kibi
	case "m":
		size *= Mebi
	case "g":
		size *= Gibi
	case "t":
		size *= Tebi
	}

	return Size{Value: size}, nil
}

func (s Size) String() string {
	b := s.Value / Kibi

	if b == 0 {
		return "0"
	}

	if b%Gibi == 0 {
		return fmt.Sprintf("%dT", b/Gibi)
	}

	if b%Mebi == 0 {
		return fmt.Sprintf("%dG", b/Mebi)
	}

	if b%Kibi == 0 {
		return fmt.Sprintf("%dM", b/Kibi)
	}

	return fmt.Sprintf("%dK", b)
}

// ParseUnit parses a unit string and returns the number of bytes in the given unit. It assumes all units are binary
// units.
func ParseUnit(u string) (int64, error) {
	switch strings.TrimSpace(u) {
	case "kB", "KB", "KiB":
		return Kibi, nil
	case "MB", "MiB":
		return Mebi, nil
	case "GB", "GiB":
		return Gibi, nil
	case "TB", "TiB":
		return Tebi, nil
	case "B", "":
		return int64(1), nil
	default:
		return 0, fmt.Errorf("unrecognized unit %q", u)
	}
}
