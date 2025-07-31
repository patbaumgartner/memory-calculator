package main

import (
	"testing"
)

// Additional test cases for parseMemoryString beyond what's in main_test.go
func TestParseMemoryStringExtended(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		// Whitespace handling
		{" 1G ", 1024 * 1024 * 1024, false},
		{"  512M  ", 512 * 1024 * 1024, false},

		// Large values
		{"16G", 16 * 1024 * 1024 * 1024, false},
		{"1024M", 1024 * 1024 * 1024, false},

		// More error cases
		{"1.2.3G", 0, true},
		{"abc123", 0, true},
		{"G1", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseMemoryString(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("parseMemoryString(%q) expected error, got result: %d", tt.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("parseMemoryString(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("parseMemoryString(%q) = %d, expected %d", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// Additional test cases for formatMemory beyond what's in main_test.go
func TestFormatMemoryExtended(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{1, "1 B"},
		{1023, "1023 B"},
		{1025, "1 KB"},
		{2048, "2 KB"},
		{1024*1024 - 1, "1024 KB"},
		{1024*1024 + 1, "1 MB"},
		{1536 * 1024, "2 MB"}, // 1.5MB rounds to 2MB
		{1024*1024*1024 - 1, "1024 MB"},
		{2560 * 1024 * 1024, "2.50 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatMemory(tt.input)
			if result != tt.expected {
				t.Errorf("formatMemory(%d) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
