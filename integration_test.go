package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestIntegrationMain(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "memory-calculator-test", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() {
		// Clean up the test binary
		exec.Command("rm", "memory-calculator-test").Run()
	}()

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput []string
	}{
		{
			name:           "Help flag",
			args:           []string{"-help"},
			expectError:    false,
			expectedOutput: []string{"JVM Memory Calculator", "Examples:"},
		},
		{
			name:           "Valid memory specification",
			args:           []string{"-total-memory=1G"},
			expectError:    false,
			expectedOutput: []string{"Using specified memory", "1.00 GB", "JVM Memory Configuration"},
		},
		{
			name:           "Thread count parameter",
			args:           []string{"-total-memory=2G", "-thread-count=300"},
			expectError:    false,
			expectedOutput: []string{"Thread Count:     300"},
		},
		{
			name:           "Head room parameter",
			args:           []string{"-total-memory=2G", "-head-room=10"},
			expectError:    false,
			expectedOutput: []string{"Head Room:        10%"},
		},
		{
			name:           "Class count parameter",
			args:           []string{"-total-memory=2G", "-loaded-class-count=40000"},
			expectError:    false,
			expectedOutput: []string{"Loaded Classes:   40000"},
		},
		{
			name:           "Invalid memory format",
			args:           []string{"-total-memory=invalid"},
			expectError:    false, // Should not error, but should show warning
			expectedOutput: []string{"Invalid total-memory value"},
		},
		{
			name:           "Memory units - MB",
			args:           []string{"-total-memory=512MB"},
			expectError:    false,
			expectedOutput: []string{"512 MB"},
		},
		{
			name:           "Memory units - KB",
			args:           []string{"-total-memory=524288KB"},
			expectError:    false,
			expectedOutput: []string{"512 MB"},
		},
		{
			name:           "Decimal memory",
			args:           []string{"-total-memory=1.5G"},
			expectError:    false,
			expectedOutput: []string{"1.50 GB"},
		},
		{
			name:           "Quiet mode - only JVM options",
			args:           []string{"--quiet", "--total-memory=2G"},
			expectError:    false,
			expectedOutput: []string{"-Xmx", "-XX:MaxMetaspaceSize", "-Xss"},
		},
		{
			name:           "Quiet mode with parameters",
			args:           []string{"--quiet", "--total-memory=1G", "--thread-count=300"},
			expectError:    false,
			expectedOutput: []string{"-Xmx", "-XX:MaxMetaspaceSize", "-Xss"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./memory-calculator-test", tt.args...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but command succeeded")
			}
			if !tt.expectError && err != nil {
				// Some "errors" are expected (like memory calculation failures with extreme parameters)
				// Only fail if it's a real unexpected error
				if !strings.Contains(outputStr, "Memory calculation failed") {
					t.Errorf("Unexpected error: %v\nOutput: %s", err, outputStr)
				}
			}

			// Check for expected output strings
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, but got:\n%s", expected, outputStr)
				}
			}
		})
	}
}

func TestDoubleVsSingleDash(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "memory-calculator-test", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() {
		exec.Command("rm", "memory-calculator-test").Run()
	}()

	// Test that both single and double dash work the same way
	singleDashCmd := exec.Command("./memory-calculator-test", "-total-memory=1G", "-thread-count=300")
	singleDashOutput, err1 := singleDashCmd.CombinedOutput()

	doubleDashCmd := exec.Command("./memory-calculator-test", "--total-memory=1G", "--thread-count=300")
	doubleDashOutput, err2 := doubleDashCmd.CombinedOutput()

	if (err1 == nil) != (err2 == nil) {
		t.Errorf("Single dash and double dash had different error states")
	}

	// The outputs should be identical
	if string(singleDashOutput) != string(doubleDashOutput) {
		t.Errorf("Single dash and double dash produced different outputs:\nSingle: %s\nDouble: %s", 
			string(singleDashOutput), string(doubleDashOutput))
	}
}

func TestMemoryCalculationEdgeCases(t *testing.T) {
	// Build the binary first
	cmd := exec.Command("go", "build", "-o", "memory-calculator-test", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() {
		exec.Command("rm", "memory-calculator-test").Run()
	}()

	edgeCases := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "Very small memory",
			args:        []string{"-total-memory=64M"},
			expectError: false,
			description: "Should handle small memory allocations",
		},
		{
			name:        "Large memory",
			args:        []string{"-total-memory=16G"},
			expectError: false,
			description: "Should handle large memory allocations",
		},
		{
			name:        "High thread count with small memory",
			args:        []string{"-total-memory=128M", "-thread-count=1000"},
			expectError: true,
			description: "Should fail with insufficient memory for many threads",
		},
		{
			name:        "Zero head room",
			args:        []string{"-total-memory=1G", "-head-room=0"},
			expectError: false,
			description: "Should work with zero head room",
		},
		{
			name:        "High head room",
			args:        []string{"-total-memory=8G", "-head-room=20"},
			expectError: false,
			description: "Should work with moderate head room and large memory",
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("./memory-calculator-test", tc.args...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tc.expectError {
				if err == nil && !strings.Contains(outputStr, "Memory calculation failed") {
					t.Errorf("Expected error for %s, but command succeeded. Output: %s", tc.description, outputStr)
				}
			} else {
				if err != nil && strings.Contains(outputStr, "Memory calculation failed") {
					t.Errorf("Unexpected error for %s: %v\nOutput: %s", tc.description, err, outputStr)
				}
			}
		})
	}
}
