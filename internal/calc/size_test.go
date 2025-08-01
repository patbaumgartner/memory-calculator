package calc

import (
	"testing"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"1", 1, false},
		{"1K", Kibi, false},
		{"1k", Kibi, false},
		{"1M", Mebi, false},
		{"1m", Mebi, false},
		{"1G", Gibi, false},
		{"1g", Gibi, false},
		{"1T", Tebi, false},
		{"1t", Tebi, false},
		{"0", 0, false},
		{"", 0, true},
		{"invalid", 0, true},
		{"-1", 0, true},
	}

	for _, test := range tests {
		result, err := ParseSize(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %q, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", test.input, err)
			}
			if result.Value != test.expected {
				t.Errorf("For input %q, expected %d, got %d", test.input, test.expected, result.Value)
			}
		}
	}
}

func TestSizeString(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0"},
		{Kibi, "1K"},
		{Mebi, "1M"},
		{Gibi, "1G"},
		{Tebi, "1T"},
		{2 * Gibi, "2G"},
		{512 * Mebi, "512M"},
		{1536, "1K"}, // 1.5K rounds down to 1K
	}

	for _, test := range tests {
		size := Size{Value: test.size}
		result := size.String()
		if result != test.expected {
			t.Errorf("For size %d, expected %q, got %q", test.size, test.expected, result)
		}
	}
}

func TestParseUnit(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"", 1, false},
		{"B", 1, false},
		{"kB", Kibi, false},
		{"KB", Kibi, false},
		{"KiB", Kibi, false},
		{"MB", Mebi, false},
		{"MiB", Mebi, false},
		{"GB", Gibi, false},
		{"GiB", Gibi, false},
		{"TB", Tebi, false},
		{"TiB", Tebi, false},
		{" kB ", Kibi, false}, // Test trimming
		{"X", 0, true},
		{"invalid", 0, true},
	}

	for _, test := range tests {
		result, err := ParseUnit(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %q, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("For input %q, expected %d, got %d", test.input, test.expected, result)
			}
		}
	}
}
