package memory

import (
	"fmt"
	"testing"

	"github.com/patbaumgartner/memory-calculator/pkg/errors"
)

func TestParseMemoryString(t *testing.T) {
	parser := CreateParser()

	tests := []struct {
		name      string
		input     string
		expected  int64
		hasError  bool
		errorCode errors.ErrorCode
	}{
		// Bytes
		{"Raw bytes", "2147483648", 2147483648, false, ""},
		{"Zero bytes", "0", 0, false, ""},
		{"Small bytes", "1024", 1024, false, ""},

		// Kilobytes
		{"KB uppercase", "1024KB", 1024 * 1024, false, ""},
		{"K uppercase", "1024K", 1024 * 1024, false, ""},
		{"kb lowercase", "512kb", 512 * 1024, false, ""},
		{"k lowercase", "512k", 512 * 1024, false, ""},

		// Megabytes
		{"MB uppercase", "512MB", 512 * 1024 * 1024, false, ""},
		{"M uppercase", "512M", 512 * 1024 * 1024, false, ""},
		{"mb lowercase", "256mb", 256 * 1024 * 1024, false, ""},
		{"m lowercase", "256m", 256 * 1024 * 1024, false, ""},

		// Gigabytes
		{"GB uppercase", "2GB", 2 * 1024 * 1024 * 1024, false, ""},
		{"G uppercase", "2G", 2 * 1024 * 1024 * 1024, false, ""},
		{"gb lowercase", "4gb", 4 * 1024 * 1024 * 1024, false, ""},
		{"g lowercase", "4g", 4 * 1024 * 1024 * 1024, false, ""},

		// Terabytes
		{"TB uppercase", "1TB", 1024 * 1024 * 1024 * 1024, false, ""},
		{"T uppercase", "1T", 1024 * 1024 * 1024 * 1024, false, ""},

		// Decimal values
		{"Decimal GB", "1.5G", int64(1.5 * 1024 * 1024 * 1024), false, ""},
		{"Decimal MB", "256.5M", int64(256.5 * 1024 * 1024), false, ""},
		{"Decimal KB", "1024.25K", int64(1024.25 * 1024), false, ""},

		// Whitespace handling
		{"Leading space", " 1G", 1024 * 1024 * 1024, false, ""},
		{"Trailing space", "1G ", 1024 * 1024 * 1024, false, ""},
		{"Both spaces", " 1G ", 1024 * 1024 * 1024, false, ""},

		// Edge cases
		{"Just B unit", "1024B", 1024, false, ""},
		{"Large TB", "5T", 5 * 1024 * 1024 * 1024 * 1024, false, ""},

		// Error cases
		{"Empty string", "", 0, true, errors.ErrInvalidMemoryFormat},
		{"Invalid unit", "1X", 0, true, errors.ErrInvalidMemoryFormat},
		{"No number", "GB", 0, true, errors.ErrInvalidMemoryFormat},
		{"Invalid number", "abc", 0, true, errors.ErrInvalidMemoryFormat},
		{"Invalid format", "1.2.3G", 0, true, errors.ErrInvalidMemoryFormat},
		{"Negative number", "-1G", 0, true, errors.ErrInvalidMemoryFormat},
		{"Negative raw bytes", "-1024", 0, true, errors.ErrInvalidMemoryFormat},
		{"Too large", fmt.Sprintf("%dT", MaxMemorySize/TB+1), 0, true, errors.ErrInvalidMemoryFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseMemoryString(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", tt.input)
					return
				}

				if mcErr, ok := err.(*errors.MemoryCalculatorError); ok {
					if mcErr.Code != tt.errorCode {
						t.Errorf("Expected error code %v, got %v", tt.errorCode, mcErr.Code)
					}
				} else {
					t.Errorf("Expected MemoryCalculatorError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("For input %q, expected %d, got %d", tt.input, tt.expected, result)
				}
			}
		})
	}
}

func TestFormatMemory(t *testing.T) {
	parser := CreateParser()

	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"Zero bytes", 0, "Unknown"},
		{"Negative bytes", -1, "Unknown"},
		{"Small bytes", 512, "512 B"},
		{"Kilobytes", 2048, "2 KB"},
		{"Megabytes", 1024 * 1024, "1 MB"},
		{"Large megabytes", 512 * 1024 * 1024, "512 MB"},
		{"Gigabytes", 2 * 1024 * 1024 * 1024, "2.00 GB"},
		{"Decimal gigabytes", int64(1.5 * 1024 * 1024 * 1024), "1.50 GB"},
		{"Large gigabytes", 8 * 1024 * 1024 * 1024, "8.00 GB"},
		{"Terabytes", 1024 * 1024 * 1024 * 1024, "1024.00 GB"},
		{"Boundary MB to GB", 1024 * 1024 * 1024, "1.00 GB"},
		{"Boundary KB to MB", 1024 * 1024, "1 MB"},
		{"Boundary B to KB", 1024, "1 KB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.FormatMemory(tt.input)
			if result != tt.expected {
				t.Errorf("For input %d, expected %q, got %q", tt.input, tt.expected, result)
			}
		})
	}
}

func TestValidateMemorySize(t *testing.T) {
	parser := CreateParser()

	tests := []struct {
		name      string
		input     int64
		expectErr bool
	}{
		{"Valid small size", 1024, false},
		{"Valid large size", 1024 * 1024 * 1024, false},
		{"Zero size", 0, false},
		{"Maximum size", MaxMemorySize, false},
		{"Negative size", -1, true},
		{"Over maximum size", MaxMemorySize + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateMemorySize(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error for input %d, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %d: %v", tt.input, err)
				}
			}
		})
	}
}

func TestCreateParser(t *testing.T) {
	parser := CreateParser()
	if parser == nil {
		t.Error("CreateParser() returned nil")
	}
}

// Benchmark tests
func BenchmarkParseMemoryString(b *testing.B) {
	parser := CreateParser()
	inputs := []string{"1G", "512M", "2048K", "1073741824"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		_, _ = parser.ParseMemoryString(input)
	}
}

func BenchmarkFormatMemory(b *testing.B) {
	parser := CreateParser()
	sizes := []int64{1024, 1024 * 1024, 1024 * 1024 * 1024, 2 * 1024 * 1024 * 1024}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		_ = parser.FormatMemory(size)
	}
}

func TestConstants(t *testing.T) {
	if KB != 1024 {
		t.Errorf("Expected KB=1024, got %d", KB)
	}

	if MB != 1024*1024 {
		t.Errorf("Expected MB=1048576, got %d", MB)
	}

	if GB != 1024*1024*1024 {
		t.Errorf("Expected GB=1073741824, got %d", GB)
	}

	if TB != 1024*1024*1024*1024 {
		t.Errorf("Expected TB=1099511627776, got %d", TB)
	}

	if MaxMemorySize != 1024*TB {
		t.Errorf("Expected MaxMemorySize=1125899906842624, got %d", MaxMemorySize)
	}
}

// Property-based testing for memory parsing
func TestParseMemoryStringProperty(t *testing.T) {
	parser := CreateParser()

	// Test that parsing and formatting a valid memory string is consistent
	testCases := []string{"1G", "2G", "512M", "1024M", "2048K"}

	for _, input := range testCases {
		t.Run("Property_"+input, func(t *testing.T) {
			parsed, err := parser.ParseMemoryString(input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", input, err)
			}

			formatted := parser.FormatMemory(parsed)
			if formatted == "Unknown" {
				t.Errorf("Formatted result should not be 'Unknown' for valid input %s", input)
			}

			// Parse again to ensure consistency
			reparsed, err := parser.ParseMemoryString(input)
			if err != nil {
				t.Fatalf("Failed to reparse %s: %v", input, err)
			}

			if parsed != reparsed {
				t.Errorf("Parsing %s is not consistent: %d != %d", input, parsed, reparsed)
			}
		})
	}
}
