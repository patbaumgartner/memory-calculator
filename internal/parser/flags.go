// Package parser provides utilities for parsing flags and options.
package parser

import (
	"strings"
	"unicode"
)

// ParseFlags parses JVM flags from a string, handling basic quoting and escaping
// This replaces the go-shellwords dependency with a simpler, more focused implementation
func ParseFlags(input string) ([]string, error) {
	if input == "" {
		return nil, nil
	}

	var result []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune
	var escaped bool

	for i, r := range input {
		switch {
		case escaped:
			// Previous character was escape, add this character literally
			current.WriteRune(r)
			escaped = false

		case r == '\\':
			// Escape character
			escaped = true

		case !inQuotes && (r == '"' || r == '\''):
			// Start of quoted section
			inQuotes = true
			quoteChar = r

		case inQuotes && r == quoteChar:
			// End of quoted section - add even if empty
			result = append(result, current.String())
			current.Reset()
			inQuotes = false
			quoteChar = 0

		case !inQuotes && unicode.IsSpace(r):
			// Space outside quotes - end current argument
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}

		default:
			// Regular character
			current.WriteRune(r)
		}

		// Handle end of string
		if i == len(input)-1 {
			if current.Len() > 0 {
				result = append(result, current.String())
			}
		}
	}

	return result, nil
}
