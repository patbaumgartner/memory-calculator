package calc

import (
	"testing"
)

// TestBuildConstraints tests that our build constraint wrapper functions work correctly
// These tests ensure both standard and minimal builds produce equivalent results
func TestBuildConstraints(t *testing.T) {
	tests := []struct {
		name      string
		flag      string
		wantMatch bool
	}{
		{"Direct Memory valid", "-XX:MaxDirectMemorySize=512M", true},
		{"Direct Memory invalid", "-XX:MaxDirectMemory=512M", false},
		{"Heap valid", "-Xmx2G", true},
		{"Heap invalid", "-XX:MaxHeapSize=2G", false},
		{"Metaspace valid", "-XX:MaxMetaspaceSize=256M", true},
		{"Metaspace invalid", "-XX:Metaspace=256M", false},
		{"Reserved Code Cache valid", "-XX:ReservedCodeCacheSize=128M", true},
		{"Reserved Code Cache invalid", "-XX:CodeCacheSize=128M", false},
		{"Stack valid", "-Xss2M", true},
		{"Stack invalid", "-XX:ThreadStackSize=2M", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test match functions
			switch {
			case matchDirectMemory(tt.flag):
				if !tt.wantMatch || !matchDirectMemory(tt.flag) {
					t.Errorf("matchDirectMemory(%q) = %v, want %v", tt.flag, true, tt.wantMatch)
				}
			case matchHeap(tt.flag):
				if !tt.wantMatch || !matchHeap(tt.flag) {
					t.Errorf("matchHeap(%q) = %v, want %v", tt.flag, true, tt.wantMatch)
				}
			case matchMetaspace(tt.flag):
				if !tt.wantMatch || !matchMetaspace(tt.flag) {
					t.Errorf("matchMetaspace(%q) = %v, want %v", tt.flag, true, tt.wantMatch)
				}
			case matchReservedCodeCache(tt.flag):
				if !tt.wantMatch || !matchReservedCodeCache(tt.flag) {
					t.Errorf("matchReservedCodeCache(%q) = %v, want %v", tt.flag, true, tt.wantMatch)
				}
			case matchStack(tt.flag):
				if !tt.wantMatch || !matchStack(tt.flag) {
					t.Errorf("matchStack(%q) = %v, want %v", tt.flag, true, tt.wantMatch)
				}
			default:
				if tt.wantMatch {
					t.Errorf("No match function returned true for %q, but expected match", tt.flag)
				}
			}
		})
	}
}

// TestBuildConstraintsParsing tests that parsing functions work correctly in both builds
func TestBuildConstraintsParsing(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		wantErr bool
	}{
		{"Direct Memory valid", "-XX:MaxDirectMemorySize=512M", false},
		{"Heap valid", "-Xmx2G", false},
		{"Metaspace valid", "-XX:MaxMetaspaceSize=256M", false},
		{"Reserved Code Cache valid", "-XX:ReservedCodeCacheSize=128M", false},
		{"Stack valid", "-Xss2M", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test parsing functions
			if matchDirectMemory(tt.flag) {
				_, err := parseDirectMemory(tt.flag)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseDirectMemory(%q) error = %v, wantErr %v", tt.flag, err, tt.wantErr)
				}
			}
			if matchHeap(tt.flag) {
				_, err := parseHeap(tt.flag)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseHeap(%q) error = %v, wantErr %v", tt.flag, err, tt.wantErr)
				}
			}
			if matchMetaspace(tt.flag) {
				_, err := parseMetaspace(tt.flag)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseMetaspace(%q) error = %v, wantErr %v", tt.flag, err, tt.wantErr)
				}
			}
			if matchReservedCodeCache(tt.flag) {
				_, err := parseReservedCodeCache(tt.flag)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseReservedCodeCache(%q) error = %v, wantErr %v", tt.flag, err, tt.wantErr)
				}
			}
			if matchStack(tt.flag) {
				_, err := parseStack(tt.flag)
				if (err != nil) != tt.wantErr {
					t.Errorf("parseStack(%q) error = %v, wantErr %v", tt.flag, err, tt.wantErr)
				}
			}
		})
	}
}

// TestCalculatorWithBuildConstraints ensures the Calculator works correctly with both build variants
func TestCalculatorWithBuildConstraints(t *testing.T) {
	calc := Calculator{
		TotalMemory:      Size{Value: 2 * 1024 * 1024 * 1024}, // 2GB
		ThreadCount:      100,
		LoadedClassCount: 10000,
		HeadRoom:         0,
	}

	// Test with common JVM flags
	flags := "-Xmx1G -XX:MaxMetaspaceSize=256M -Xss2M"

	result, err := calc.Calculate(flags)
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	// Verify results are reasonable
	if result.Heap == nil {
		t.Error("Heap should be set")
	}
	if result.Metaspace == nil {
		t.Error("Metaspace should be set")
	}
	if result.Stack.Value <= 0 {
		t.Error("Stack should be positive")
	}

	// Test specific parsed values
	if result.Heap.Value != 1024*1024*1024 { // 1GB
		t.Errorf("Heap = %d, want %d", result.Heap.Value, 1024*1024*1024)
	}
	if result.Metaspace.Value != 256*1024*1024 { // 256MB
		t.Errorf("Metaspace = %d, want %d", result.Metaspace.Value, 256*1024*1024)
	}
	if result.Stack.Value != 2*1024*1024 { // 2MB
		t.Errorf("Stack = %d, want %d", result.Stack.Value, 2*1024*1024)
	}
}

// TestBuildConstraintsConsistency ensures both build variants produce the same results for valid inputs
func TestBuildConstraintsConsistency(t *testing.T) {
	// This test will pass in both standard and minimal builds if they produce consistent results
	testCases := []struct {
		flag string
		size int64
	}{
		{"-Xmx1G", 1024 * 1024 * 1024},
		{"-Xmx512M", 512 * 1024 * 1024},
		{"-XX:MaxMetaspaceSize=128M", 128 * 1024 * 1024},
		{"-XX:MaxDirectMemorySize=64M", 64 * 1024 * 1024},
		{"-Xss1M", 1024 * 1024},
	}

	for _, tc := range testCases {
		t.Run(tc.flag, func(t *testing.T) {
			calc := Calculator{
				TotalMemory:      Size{Value: 4 * 1024 * 1024 * 1024}, // 4GB
				ThreadCount:      100,
				LoadedClassCount: 10000,
				HeadRoom:         0,
			}

			result, err := calc.Calculate(tc.flag)
			if err != nil {
				t.Fatalf("Calculate(%q) error = %v", tc.flag, err)
			}

			// Check that the specific flag was parsed correctly
			var actualSize int64
			switch {
			case matchHeap(tc.flag):
				if result.Heap == nil {
					t.Fatal("Heap should be set")
				}
				actualSize = result.Heap.Value
			case matchMetaspace(tc.flag):
				if result.Metaspace == nil {
					t.Fatal("Metaspace should be set")
				}
				actualSize = result.Metaspace.Value
			case matchDirectMemory(tc.flag):
				actualSize = result.DirectMemory.Value
			case matchStack(tc.flag):
				actualSize = result.Stack.Value
			}

			if actualSize != tc.size {
				t.Errorf("Size for %q = %d, want %d", tc.flag, actualSize, tc.size)
			}
		})
	}
}
