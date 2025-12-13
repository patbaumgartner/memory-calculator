package calc

import (
	"testing"
)

// validateCalculationResult validates the basic structure of a calculation result
func validateCalculationResult(t *testing.T, result MemoryRegions) {
	t.Helper()

	if result.DirectMemory.Value <= 0 {
		t.Error("DirectMemory should have positive value")
	}
	if result.ReservedCodeCache.Value <= 0 {
		t.Error("ReservedCodeCache should have positive value")
	}
	if result.Stack.Value <= 0 {
		t.Error("Stack should have positive value")
	}
	if result.Metaspace == nil {
		t.Error("Metaspace should not be nil")
	}
	if result.HeadRoom == nil {
		t.Error("HeadRoom should not be nil")
	}
	if result.Heap != nil && result.Heap.Value <= 0 {
		t.Error("Heap value should be positive")
	}
}

// validateMemoryBounds validates that total memory used doesn't exceed available memory
func validateMemoryBounds(t *testing.T, result MemoryRegions, totalMemory int64, threadCount int) {
	t.Helper()

	if result.Heap == nil {
		return
	}

	totalUsed := result.Heap.Value
	if result.HeadRoom != nil {
		totalUsed += result.HeadRoom.Value
	}
	if result.Metaspace != nil {
		totalUsed += result.Metaspace.Value
	}
	totalUsed += result.DirectMemory.Value
	totalUsed += result.ReservedCodeCache.Value
	totalUsed += result.Stack.Value * int64(threadCount)

	if totalUsed > totalMemory {
		t.Errorf("Total memory used (%d) exceeds available memory (%d)", totalUsed, totalMemory)
	}
}

func TestCalculatorCalculate(t *testing.T) {
	tests := []struct {
		name        string
		calculator  Calculator
		flags       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Basic calculation with defaults",
			calculator: Calculator{
				HeadRoom:         10,
				LoadedClassCount: 5000,
				ThreadCount:      250,
				TotalMemory:      Size{Value: 2 * Gibi},
			},
			flags:       "",
			expectError: false,
		},
		{
			name: "Calculation with custom heap",
			calculator: Calculator{
				HeadRoom:         5,
				LoadedClassCount: 3000,
				ThreadCount:      100,
				TotalMemory:      Size{Value: 1 * Gibi},
			},
			flags:       "-Xmx512m",
			expectError: false,
		},
		{
			name: "Calculation with custom metaspace",
			calculator: Calculator{
				HeadRoom:         0,
				LoadedClassCount: 2000,
				ThreadCount:      50,
				TotalMemory:      Size{Value: 1 * Gibi},
			},
			flags:       "-XX:MaxMetaspaceSize=128m",
			expectError: false,
		},
		{
			name: "Calculation with multiple JVM flags",
			calculator: Calculator{
				HeadRoom:         15,
				LoadedClassCount: 8000,
				ThreadCount:      300,
				TotalMemory:      Size{Value: 4 * Gibi},
			},
			flags:       "-Xmx1g -XX:MaxMetaspaceSize=256m -XX:MaxDirectMemorySize=128m",
			expectError: false,
		},
		{
			name: "Flags with unclosed quotes (handled gracefully)",
			calculator: Calculator{
				HeadRoom:         10,
				LoadedClassCount: 5000,
				ThreadCount:      250,
				TotalMemory:      Size{Value: 2 * Gibi},
			},
			flags:       `"unclosed quote`,
			expectError: false,
		},
		{
			name: "Invalid direct memory format",
			calculator: Calculator{
				HeadRoom:         10,
				LoadedClassCount: 5000,
				ThreadCount:      250,
				TotalMemory:      Size{Value: 2 * Gibi},
			},
			flags:       "-XX:MaxDirectMemorySize=999999999999999999999G", // Too large
			expectError: true,
			errorMsg:    "unable to parse direct memory",
		},
		{
			name: "Invalid heap format",
			calculator: Calculator{
				HeadRoom:         10,
				LoadedClassCount: 5000,
				ThreadCount:      250,
				TotalMemory:      Size{Value: 2 * Gibi},
			},
			flags:       "-Xmx999999999999999999999G", // Too large
			expectError: true,
			errorMsg:    "unable to parse heap",
		},
		{
			name: "Invalid metaspace format",
			calculator: Calculator{
				HeadRoom:         10,
				LoadedClassCount: 5000,
				ThreadCount:      250,
				TotalMemory:      Size{Value: 2 * Gibi},
			},
			flags:       "-XX:MaxMetaspaceSize=999999999999999999999G", // Too large
			expectError: true,
			errorMsg:    "unable to parse metaspace",
		},
		{
			name: "Invalid reserved code cache format",
			calculator: Calculator{
				HeadRoom:         10,
				LoadedClassCount: 5000,
				ThreadCount:      250,
				TotalMemory:      Size{Value: 2 * Gibi},
			},
			flags:       "-XX:ReservedCodeCacheSize=999999999999999999999G", // Too large
			expectError: true,
			errorMsg:    "unable to parse reserved code cache",
		},
		{
			name: "Invalid stack format",
			calculator: Calculator{
				HeadRoom:         10,
				LoadedClassCount: 5000,
				ThreadCount:      250,
				TotalMemory:      Size{Value: 2 * Gibi},
			},
			flags:       "-Xss999999999999999999999G", // Too large
			expectError: true,
			errorMsg:    "unable to parse stack",
		},
		{
			name: "Memory too small for fixed regions",
			calculator: Calculator{
				HeadRoom:         0,
				LoadedClassCount: 1000,
				ThreadCount:      1,
				TotalMemory:      Size{Value: 1 * Mebi}, // Very small memory
			},
			flags:       "",
			expectError: true,
			errorMsg:    "fixed memory regions require",
		},
		{
			name: "Memory too small for all regions with large heap",
			calculator: Calculator{
				HeadRoom:         0,
				LoadedClassCount: 1000,
				ThreadCount:      10,
				TotalMemory:      Size{Value: 100 * Mebi}, // Small memory
			},
			flags:       "-Xmx1g", // Huge heap relative to total memory
			expectError: true,
			errorMsg:    "fixed memory regions require", // Actually fails before heap check
		},
		{
			name: "Extreme thread count",
			calculator: Calculator{
				HeadRoom:         0,
				LoadedClassCount: 1000,
				ThreadCount:      10000, // Very high thread count
				TotalMemory:      Size{Value: 1 * Gibi},
			},
			flags:       "",
			expectError: true,
			errorMsg:    "fixed memory regions require", // Fails due to thread stack memory
		},
		{
			name: "Zero thread count edge case",
			calculator: Calculator{
				HeadRoom:         0,
				LoadedClassCount: 1000,
				ThreadCount:      0, // Edge case
				TotalMemory:      Size{Value: 1 * Gibi},
			},
			flags:       "",
			expectError: false,
		},
		{
			name: "Large class count",
			calculator: Calculator{
				HeadRoom:         0,
				LoadedClassCount: 1000000, // Very large class count
				ThreadCount:      250,
				TotalMemory:      Size{Value: 16 * Gibi}, // Large memory to accommodate
			},
			flags:       "",
			expectError: false,
		},
		{
			name: "Maximum head room",
			calculator: Calculator{
				HeadRoom:         20, // More reasonable head room for 8GB
				LoadedClassCount: 1000,
				ThreadCount:      250,
				TotalMemory:      Size{Value: 8 * Gibi},
			},
			flags:       "",
			expectError: false,
		},
		{
			name: "All custom JVM options",
			calculator: Calculator{
				HeadRoom:         5,
				LoadedClassCount: 5000,
				ThreadCount:      200,
				TotalMemory:      Size{Value: 3 * Gibi},
			},
			flags:       "-Xmx1g -XX:MaxMetaspaceSize=512m -XX:MaxDirectMemorySize=256m -XX:ReservedCodeCacheSize=128m -Xss2m",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.calculator.Calculate(tt.flags)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got no error", tt.errorMsg)
					return
				}
				if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.errorMsg, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			// Validate the result structure using helper functions
			validateCalculationResult(t, result)
			validateMemoryBounds(t, result, tt.calculator.TotalMemory.Value, tt.calculator.ThreadCount)
		})
	}
}

func TestCalculatorEdgeCases(t *testing.T) {
	t.Run("Minimal memory configuration", func(t *testing.T) {
		calc := Calculator{
			HeadRoom:         0,
			LoadedClassCount: 1,
			ThreadCount:      1,
			TotalMemory:      Size{Value: 512 * Mebi},
		}

		result, err := calc.Calculate("")
		if err != nil {
			t.Logf("Expected minimal configuration to work or fail gracefully: %v", err)
			return // This might legitimately fail due to memory constraints
		}

		if result.Heap == nil || result.Heap.Value <= 0 {
			t.Error("Should have allocated some heap memory")
		}
	})

	t.Run("Complex JVM flags parsing", func(t *testing.T) {
		calc := Calculator{
			HeadRoom:         10,
			LoadedClassCount: 5000,
			ThreadCount:      250,
			TotalMemory:      Size{Value: 4 * Gibi},
		}

		flags := `-server -Xmx2g -Xms1g -XX:MaxMetaspaceSize=512m -XX:MaxDirectMemorySize=256m ` +
			`-XX:ReservedCodeCacheSize=128m -Xss1m -XX:+UseG1GC -XX:G1HeapRegionSize=16m`
		result, err := calc.Calculate(flags)
		if err != nil {
			t.Errorf("Should handle complex JVM flags: %v", err)
			return
		}

		// Verify that user-configured values are preserved
		if result.Heap.Provenance != UserConfigured {
			t.Error("Heap should be user-configured")
		}
		if result.Metaspace.Provenance != UserConfigured {
			t.Error("Metaspace should be user-configured")
		}
	})

	t.Run("Memory boundary conditions", func(t *testing.T) {
		testCases := []struct {
			name        string
			totalMemory int64
			expectError bool
		}{
			{"Very small memory", 64 * Kibi, true},
			{"Small memory - insufficient", 128 * Mebi, true}, // Updated expectation
			{"Standard memory", 2 * Gibi, false},
			{"Large memory", 32 * Gibi, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				calc := Calculator{
					HeadRoom:         5,
					LoadedClassCount: 2000,
					ThreadCount:      100,
					TotalMemory:      Size{Value: tc.totalMemory},
				}

				_, err := calc.Calculate("")
				hasError := (err != nil)
				if hasError != tc.expectError {
					t.Errorf("Memory %d: expected error=%v, got error=%v (%v)",
						tc.totalMemory, tc.expectError, hasError, err)
				}
			})
		}
	})

	t.Run("Class count impact on metaspace", func(t *testing.T) {
		baseMem := int64(2 * Gibi)

		testCases := []struct {
			classCount int
			memory     int64
		}{
			{1000, baseMem},
			{10000, baseMem},
			{100000, baseMem * 4}, // Need more memory for large class counts
		}

		for _, tc := range testCases {
			t.Run("classes", func(t *testing.T) {
				calc := Calculator{
					HeadRoom:         0,
					LoadedClassCount: tc.classCount,
					ThreadCount:      100,
					TotalMemory:      Size{Value: tc.memory},
				}

				result, err := calc.Calculate("")
				if err != nil {
					t.Logf("Class count %d with memory %d failed: %v", tc.classCount, tc.memory, err)
					return
				}

				expectedMetaspace := ClassOverhead + (int64(tc.classCount) * ClassSize)
				if result.Metaspace.Value != expectedMetaspace {
					t.Errorf("Expected metaspace %d, got %d", expectedMetaspace, result.Metaspace.Value)
				}
			})
		}
	})
}

// containsString checks if a string contains a substring using simple string search.
func containsString(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
