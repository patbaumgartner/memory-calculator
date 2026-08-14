// Package calc provides core memory calculation functionality including size handling,
// memory region allocation, and JVM memory optimization algorithms.
package calc

import (
	"fmt"
	"math"
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
)

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
// indicates kibibytes, mebibytes, gibibytes or tebibytes respectively. Sizes that would overflow int64 once the
// unit multiplier is applied are rejected rather than wrapping around to a negative value.
//
// This accepts exactly the grammar HotSpot accepts for memory options: unsigned digits and at most one
// unit letter. Signs and decimal points are rejected, so anything this parser accepts is a value the JVM
// will also accept.
func ParseSize(s string) (Size, error) {
	t := strings.TrimSpace(s)

	digits, multiplier := t, int64(1)
	if len(t) > 0 {
		switch t[len(t)-1] {
		case 'k', 'K':
			digits, multiplier = t[:len(t)-1], Kibi
		case 'm', 'M':
			digits, multiplier = t[:len(t)-1], Mebi
		case 'g', 'G':
			digits, multiplier = t[:len(t)-1], Gibi
		case 't', 'T':
			digits, multiplier = t[:len(t)-1], Tebi
		}
	}

	if !isDigits(digits) {
		return Size{}, fmt.Errorf("memory size %q must be digits optionally followed by K, M, G, or T", t)
	}

	size, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return Size{}, fmt.Errorf("memory size %q does not fit in a 64-bit integer", t)
	}

	if size > math.MaxInt64/multiplier {
		return Size{}, fmt.Errorf("memory size %q overflows the maximum supported size of %d bytes", t, int64(math.MaxInt64))
	}

	return Size{Value: size * multiplier}, nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}

	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}

// String renders the size as a JVM memory argument using the largest unit that divides it exactly,
// rounding down so a generated maximum never exceeds its computed budget. Sizes below one kibibyte
// render as a plain byte count, never as "0".
func (s Size) String() string {
	if s.Value < Kibi {
		return strconv.FormatInt(s.Value, 10)
	}

	b := s.Value / Kibi

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
