package main

import (
	"strings"
	"testing"
)

func TestDisplayResults(t *testing.T) {
	// Capture output by creating a buffer-like approach
	// Since displayResults prints to stdout, we'll test its components

	testProps := map[string]string{
		"-Xmx":                      "512M",
		"-Xss":                      "1M",
		"-XX:MaxMetaspaceSize":      "128M",
		"-XX:ReservedCodeCacheSize": "240M",
		"-XX:MaxDirectMemorySize":   "10M",
		"JAVA_TOOL_OPTIONS":         "-Xmx512M -Xss1M -XX:MaxMetaspaceSize=128M",
	}

	// Test that displayResults doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayResults panicked: %v", r)
		}
	}()

	displayResults(testProps, 2*1024*1024*1024, "250", "35000", "10")
}

func TestDisplayResultsWithoutJavaToolOptions(t *testing.T) {
	// Test the case where JAVA_TOOL_OPTIONS doesn't exist
	testProps := map[string]string{
		"-Xmx":                      "512M",
		"-Xss":                      "1M",
		"-XX:MaxMetaspaceSize":      "128M",
		"-XX:ReservedCodeCacheSize": "240M",
	}

	// Test that displayResults doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayResults panicked: %v", r)
		}
	}()

	displayResults(testProps, 1024*1024*1024, "300", "40000", "5")
}

func TestDisplayResultsEmpty(t *testing.T) {
	// Test with empty props
	testProps := map[string]string{}

	// Test that displayResults doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayResults panicked: %v", r)
		}
	}()

	displayResults(testProps, 0, "250", "35000", "0")
}

func TestDisplayQuietResults(t *testing.T) {
	// Test quiet output with JAVA_TOOL_OPTIONS
	testProps := map[string]string{
		"JAVA_TOOL_OPTIONS": "-Xmx512M -Xss1M -XX:MaxMetaspaceSize=128M",
	}

	// Test that displayQuietResults doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayQuietResults panicked: %v", r)
		}
	}()

	displayQuietResults(testProps)
}

func TestDisplayQuietResultsWithoutJavaToolOptions(t *testing.T) {
	// Test quiet output without JAVA_TOOL_OPTIONS
	testProps := map[string]string{
		"-Xmx":                      "512M",
		"-Xss":                      "1M",
		"-XX:MaxMetaspaceSize":      "128M",
		"-XX:ReservedCodeCacheSize": "240M",
	}

	// Test that displayQuietResults doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayQuietResults panicked: %v", r)
		}
	}()

	displayQuietResults(testProps)
}

func TestDisplayQuietResultsEmpty(t *testing.T) {
	// Test quiet output with empty props
	testProps := map[string]string{}

	// Test that displayQuietResults doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayQuietResults panicked: %v", r)
		}
	}()

	displayQuietResults(testProps)
}

func TestExtractJVMFlag(t *testing.T) {
	javaToolOptions := "-XX:MaxDirectMemorySize=10M -Xmx324661K -XX:MaxMetaspaceSize=211914K -XX:ReservedCodeCacheSize=240M -Xss1M"

	tests := []struct {
		flag     string
		expected string
	}{
		{"-Xmx", "324661K"},
		{"-Xss", "1M"},
		{"-XX:MaxMetaspaceSize", "211914K"},
		{"-XX:ReservedCodeCacheSize", "240M"},
		{"-XX:MaxDirectMemorySize", "10M"},
		{"-XX:NonExistentFlag", ""},
		{"-Xms", ""}, // Not in the string
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			result := extractJVMFlag(javaToolOptions, tt.flag)
			if result != tt.expected {
				t.Errorf("extractJVMFlag(%q, %q) = %q, expected %q", javaToolOptions, tt.flag, result, tt.expected)
			}
		})
	}
}

func TestDisplayJVMSetting(t *testing.T) {
	// Test with individual keys
	testProps := map[string]string{
		"-Xmx": "512M",
		"-Xss": "1M",
	}

	// Test that displayJVMSetting doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayJVMSetting panicked: %v", r)
		}
	}()

	displayJVMSetting(testProps, "-Xmx", "Max Heap Size: ")
	displayJVMSetting(testProps, "-Xss", "Thread Stack Size: ")
	displayJVMSetting(testProps, "-XX:NonExistent", "Non Existent: ")
}

func TestDisplayJVMSettingFromJavaToolOptions(t *testing.T) {
	// Test with JAVA_TOOL_OPTIONS
	testProps := map[string]string{
		"JAVA_TOOL_OPTIONS": "-XX:MaxDirectMemorySize=10M -Xmx324661K -XX:MaxMetaspaceSize=211914K -Xss1M",
	}

	// Test that displayJVMSetting doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayJVMSetting panicked: %v", r)
		}
	}()

	displayJVMSetting(testProps, "-Xmx", "Max Heap Size: ")
	displayJVMSetting(testProps, "-XX:MaxMetaspaceSize", "Max Metaspace Size: ")
	displayJVMSetting(testProps, "-XX:NonExistent", "Non Existent: ")
}

func TestStringRepeat(t *testing.T) {
	// Test the strings.Repeat function used in displayResults
	result := strings.Repeat("=", 50)
	expected := "=================================================="
	if result != expected {
		t.Errorf("strings.Repeat('=', 50) = %q, expected %q", result, expected)
	}

	result = strings.Repeat("-", 30)
	expected = "------------------------------"
	if result != expected {
		t.Errorf("strings.Repeat('-', 30) = %q, expected %q", result, expected)
	}
}

// Test edge cases for memory calculation inputs
func TestMemoryCalculationInputs(t *testing.T) {
	tests := []struct {
		name        string
		memory      int64
		threadCount string
		classCount  string
		headRoom    string
		shouldPass  bool
	}{
		{
			name:        "Normal case",
			memory:      2 * 1024 * 1024 * 1024, // 2GB
			threadCount: "250",
			classCount:  "35000",
			headRoom:    "10",
			shouldPass:  true,
		},
		{
			name:        "Zero memory",
			memory:      0,
			threadCount: "250",
			classCount:  "35000",
			headRoom:    "0",
			shouldPass:  true, // Should not crash
		},
		{
			name:        "Negative memory",
			memory:      -1,
			threadCount: "250",
			classCount:  "35000",
			headRoom:    "0",
			shouldPass:  true, // Should not crash
		},
		{
			name:        "Very large memory",
			memory:      1024 * 1024 * 1024 * 1024, // 1TB
			threadCount: "250",
			classCount:  "35000",
			headRoom:    "0",
			shouldPass:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && tt.shouldPass {
					t.Errorf("Test %s panicked when it should pass: %v", tt.name, r)
				}
			}()

			// Test formatMemory with various inputs
			result := formatMemory(tt.memory)
			if tt.memory <= 0 && result != "Unknown" {
				t.Errorf("formatMemory(%d) = %q, expected 'Unknown'", tt.memory, result)
			}
		})
	}
}

// Test various memory format edge cases
func TestMemoryFormatEdgeCases(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{1, "1 B"},
		{1023, "1023 B"},
		{1024, "1 KB"},
		{1025, "1 KB"},
		{2048, "2 KB"},
		{1024*1024 - 1, "1024 KB"},
		{1024 * 1024, "1 MB"},
		{1024*1024 + 1, "1 MB"},
		{1536 * 1024, "2 MB"}, // 1.5MB rounds to 2MB
		{1024*1024*1024 - 1, "1024 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{1536 * 1024 * 1024, "1.50 GB"},
		{2560 * 1024 * 1024, "2.50 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatMemory(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatMemory(%d) = %q, expected %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

// Benchmark the main parsing functions
func BenchmarkDetectContainerMemory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		detectContainerMemory()
	}
}

func BenchmarkReadCgroupsV1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readCgroupsV1()
	}
}

func BenchmarkReadCgroupsV2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readCgroupsV2()
	}
}
