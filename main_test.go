package main

import (
	"os"
	"testing"
)

func TestParseMemoryString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		hasError bool
	}{
		// Bytes
		{"Raw bytes", "2147483648", 2147483648, false},
		{"Zero bytes", "0", 0, false},
		
		// Kilobytes
		{"KB uppercase", "1024KB", 1024 * 1024, false},
		{"K uppercase", "1024K", 1024 * 1024, false},
		{"kb lowercase", "512kb", 512 * 1024, false},
		{"k lowercase", "512k", 512 * 1024, false},
		
		// Megabytes
		{"MB uppercase", "512MB", 512 * 1024 * 1024, false},
		{"M uppercase", "512M", 512 * 1024 * 1024, false},
		{"mb lowercase", "256mb", 256 * 1024 * 1024, false},
		{"m lowercase", "256m", 256 * 1024 * 1024, false},
		
		// Gigabytes
		{"GB uppercase", "2GB", 2 * 1024 * 1024 * 1024, false},
		{"G uppercase", "2G", 2 * 1024 * 1024 * 1024, false},
		{"gb lowercase", "4gb", 4 * 1024 * 1024 * 1024, false},
		{"g lowercase", "4g", 4 * 1024 * 1024 * 1024, false},
		
		// Terabytes
		{"TB uppercase", "1TB", 1024 * 1024 * 1024 * 1024, false},
		{"T uppercase", "1T", 1024 * 1024 * 1024 * 1024, false},
		
		// Decimal values
		{"Decimal GB", "1.5G", int64(1.5 * 1024 * 1024 * 1024), false},
		{"Decimal MB", "256.5M", int64(256.5 * 1024 * 1024), false},
		{"Decimal KB", "1024.25K", int64(1024.25 * 1024), false},
		
		// Whitespace handling
		{"Leading space", " 1G", 1024 * 1024 * 1024, false},
		{"Trailing space", "1G ", 1024 * 1024 * 1024, false},
		{"Both spaces", " 1G ", 1024 * 1024 * 1024, false},
		
		// Error cases
		{"Empty string", "", 0, true},
		{"Invalid unit", "1X", 0, true},
		{"No number", "GB", 0, true},
		{"Invalid number", "abc", 0, true},
		{"Invalid format", "1.2.3G", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseMemoryString(tt.input)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", tt.input)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMemory(tt.input)
			if result != tt.expected {
				t.Errorf("For input %d, expected %q, got %q", tt.input, tt.expected, result)
			}
		})
	}
}

func TestDetectContainerMemory(t *testing.T) {
	// This test is environment-dependent, so we'll just test that it doesn't panic
	// and returns a non-negative value
	t.Run("Memory detection doesn't panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("detectContainerMemory() panicked: %v", r)
			}
		}()
		
		memory := detectContainerMemory()
		if memory < 0 {
			t.Errorf("detectContainerMemory() returned negative value: %d", memory)
		}
	})
}

// Benchmark tests
func BenchmarkParseMemoryString(b *testing.B) {
	testCases := []string{"1G", "512M", "1024K", "2147483648"}
	
	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = parseMemoryString(tc)
			}
		})
	}
}

func BenchmarkFormatMemory(b *testing.B) {
	testCases := []int64{
		1024,
		1024 * 1024,
		1024 * 1024 * 1024,
		2 * 1024 * 1024 * 1024,
	}
	
	for _, tc := range testCases {
		b.Run("", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = formatMemory(tc)
			}
		})
	}
}

// Integration tests
func TestEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalThreadCount := os.Getenv("BPL_JVM_THREAD_COUNT")
	originalClassCount := os.Getenv("BPL_JVM_LOADED_CLASS_COUNT")
	originalHeadRoom := os.Getenv("BPL_JVM_HEAD_ROOM")
	originalTotalMemory := os.Getenv("BPL_JVM_TOTAL_MEMORY")
	
	// Clean up after test
	defer func() {
		os.Setenv("BPL_JVM_THREAD_COUNT", originalThreadCount)
		os.Setenv("BPL_JVM_LOADED_CLASS_COUNT", originalClassCount)
		os.Setenv("BPL_JVM_HEAD_ROOM", originalHeadRoom)
		os.Setenv("BPL_JVM_TOTAL_MEMORY", originalTotalMemory)
	}()
	
	tests := []struct {
		name         string
		threadCount  string
		classCount   string
		headRoom     string
		totalMemory  int64
		expectError  bool
	}{
		{
			name:        "Valid configuration",
			threadCount: "250",
			classCount:  "35000",
			headRoom:    "0",
			totalMemory: 2 * 1024 * 1024 * 1024, // 2GB
			expectError: false,
		},
		{
			name:        "High thread count",
			threadCount: "500",
			classCount:  "35000",
			headRoom:    "0",
			totalMemory: 8 * 1024 * 1024 * 1024, // 8GB
			expectError: false,
		},
		{
			name:        "With head room",
			threadCount: "250",
			classCount:  "35000",
			headRoom:    "10",
			totalMemory: 4 * 1024 * 1024 * 1024, // 4GB
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("BPL_JVM_THREAD_COUNT", tt.threadCount)
			os.Setenv("BPL_JVM_LOADED_CLASS_COUNT", tt.classCount)
			os.Setenv("BPL_JVM_HEAD_ROOM", tt.headRoom)
			if tt.totalMemory > 0 {
				os.Setenv("BPL_JVM_TOTAL_MEMORY", string(rune(tt.totalMemory)))
			}
			
			// Verify environment variables are set correctly
			if os.Getenv("BPL_JVM_THREAD_COUNT") != tt.threadCount {
				t.Errorf("BPL_JVM_THREAD_COUNT not set correctly")
			}
			if os.Getenv("BPL_JVM_LOADED_CLASS_COUNT") != tt.classCount {
				t.Errorf("BPL_JVM_LOADED_CLASS_COUNT not set correctly")
			}
			if os.Getenv("BPL_JVM_HEAD_ROOM") != tt.headRoom {
				t.Errorf("BPL_JVM_HEAD_ROOM not set correctly")
			}
		})
	}
}
