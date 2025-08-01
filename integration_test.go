package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainIntegration(t *testing.T) {
	// Build the binary for testing
	binaryPath := filepath.Join(os.TempDir(), "memory-calculator-test")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/memory-calculator")
	cmd.Dir = "./"
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove(binaryPath)

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput []string
		notExpected    []string
	}{
		{
			name: "Help flag",
			args: []string{"--help"},
			expectedOutput: []string{
				"JVM Memory Calculator",
				"Usage:",
				"--total-memory",
				"--thread-count",
				"Examples:",
			},
		},
		{
			name: "Version flag",
			args: []string{"--version"},
			expectedOutput: []string{
				"JVM Memory Calculator",
				"Version:",
				"Build Time:",
				"Commit:",
				"Go Version: 1.24.5",
			},
		},
		{
			name: "Quiet mode with memory",
			args: []string{"--quiet", "--total-memory", "2G"},
			expectedOutput: []string{
				"-X", // Should contain JVM flags
			},
			notExpected: []string{
				"JVM Memory Configuration",
				"Total Memory:",
			},
		},
		{
			name: "Standard output with memory",
			args: []string{"--total-memory", "1G", "--thread-count", "300"},
			expectedOutput: []string{
				"JVM Memory Configuration",
				"Total Memory:",
				"Thread Count:     300",
				"JAVA_TOOL_OPTIONS",
			},
		},
		{
			name: "Invalid memory format",
			args: []string{"--total-memory", "invalid"},
			expectedOutput: []string{
				"JVM Memory Configuration", // Should still show output with detected memory
			},
		},
		{
			name:        "Invalid thread count",
			args:        []string{"--thread-count", "-1"},
			expectError: true,
		},
		{
			name:        "Invalid head room",
			args:        []string{"--head-room", "150"},
			expectError: true,
		},
		{
			name: "Memory units - MB",
			args: []string{"--total-memory", "512MB", "--quiet"},
			expectedOutput: []string{
				"-X", // Should contain JVM flags
			},
		},
		{
			name: "Memory units - KB",
			args: []string{"--total-memory", "2048000KB", "--quiet"},
			expectedOutput: []string{
				"-X", // Should contain JVM flags
			},
		},
		{
			name: "Decimal memory",
			args: []string{"--total-memory", "1.5G", "--quiet"},
			expectedOutput: []string{
				"-X", // Should contain JVM flags
			},
		},
		{
			name: "All parameters",
			args: []string{
				"--total-memory", "8G",
				"--thread-count", "200",
				"--loaded-class-count", "30000",
				"--head-room", "10",
			},
			expectedOutput: []string{
				"JVM Memory Configuration",
				"Thread Count:     200",
				"Loaded Classes:   30000",
				"Head Room:        10%",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			// Set environment variables for testing
			env := append(os.Environ(),
				"BPI_APPLICATION_PATH=.",
			)

			// Use fewer classes for small memory tests
			if strings.Contains(tt.name, "KB") || strings.Contains(tt.name, "MB") {
				env = append(env,
					"BPL_JVM_LOADED_CLASS_COUNT=1000",
					"BPL_JVM_THREAD_COUNT=50",
				)
			} else {
				env = append(env, "BPL_JVM_LOADED_CLASS_COUNT=30000")
			}

			cmd.Env = env
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but command succeeded. Output: %s", outputStr)
				}
				return
			}

			if err != nil {
				t.Errorf("Command failed with error: %v. Output: %s", err, outputStr)
				return
			}

			// Check expected output
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, but got:\n%s", expected, outputStr)
				}
			}

			// Check not expected output
			for _, notExpected := range tt.notExpected {
				if strings.Contains(outputStr, notExpected) {
					t.Errorf("Expected output to NOT contain %q, but got:\n%s", notExpected, outputStr)
				}
			}
		})
	}
}

func TestMainEnvironmentVariables(t *testing.T) {
	// Build the binary for testing
	binaryPath := filepath.Join(os.TempDir(), "memory-calculator-test")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/memory-calculator")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove(binaryPath)

	// Test with environment variables
	cmd = exec.Command(binaryPath, "--total-memory", "4G")
	cmd.Env = append(os.Environ(),
		"BPI_APPLICATION_PATH=.",
		"BPL_JVM_THREAD_COUNT=200",
		"BPL_JVM_LOADED_CLASS_COUNT=30000",
		"BPL_JVM_HEAD_ROOM=10",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v. Output: %s", err, string(output))
	}

	outputStr := string(output)
	expectedParts := []string{
		"Thread Count:     200",
		"Loaded Classes:   30000",
		"Head Room:        10%",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected output to contain %q, but got:\n%s", expected, outputStr)
		}
	}
}

func TestMainBoundaryValues(t *testing.T) {
	// Build the binary for testing
	binaryPath := filepath.Join(os.TempDir(), "memory-calculator-test")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/memory-calculator")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove(binaryPath)

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name: "Minimum valid values",
			args: []string{
				"--total-memory", "512M",
				"--thread-count", "1",
				"--loaded-class-count", "100",
				"--head-room", "0",
			},
		},
		{
			name: "Maximum head room",
			args: []string{
				"--total-memory", "1G",
				"--head-room", "100",
			},
			expectError: true, // 100% head room leaves no memory for JVM
		},
		{
			name: "Large memory value",
			args: []string{
				"--total-memory", "100G",
			},
		},
		{
			name: "Zero thread count",
			args: []string{
				"--thread-count", "0",
			},
			expectError: true,
		},
		{
			name: "Negative head room",
			args: []string{
				"--head-room", "-1",
			},
			expectError: true,
		},
		{
			name: "Head room over 100",
			args: []string{
				"--head-room", "101",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			// Set environment variables for testing
			cmd.Env = append(os.Environ(),
				"BPI_APPLICATION_PATH=.",
				"BPL_JVM_LOADED_CLASS_COUNT=30000",
			)
			output, err := cmd.CombinedOutput()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but command succeeded. Output: %s", string(output))
				}
			} else {
				if err != nil {
					t.Errorf("Command failed with error: %v. Output: %s", err, string(output))
				}
			}
		})
	}
}

func TestMainHostMemoryDetection(t *testing.T) {
	// Build the binary for testing
	binaryPath := filepath.Join(t.TempDir(), "memory-calculator")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/memory-calculator")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove(binaryPath)

	tests := []struct {
		name        string
		args        []string
		description string
	}{
		{
			name:        "Host memory auto-detection",
			args:        []string{},
			description: "Should detect host memory when no memory is specified",
		},
		{
			name:        "Manual memory override",
			args:        []string{"--total-memory", "2G"},
			description: "Should use specified memory over auto-detection",
		},
		{
			name:        "Host detection with quiet mode",
			args:        []string{"--quiet"},
			description: "Should auto-detect host memory in quiet mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Env = append(os.Environ(),
				"BPI_APPLICATION_PATH=.",
				"BPL_JVM_LOADED_CLASS_COUNT=30000",
			)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Command failed: %v. Output: %s", err, output)
				return
			}

			outputStr := string(output)

			// Verify output contains memory detection message (unless quiet mode)
			hasQuiet := false
			for _, arg := range tt.args {
				if arg == "--quiet" {
					hasQuiet = true
					break
				}
			}

			if !hasQuiet {
				// Should contain memory detection message
				if !strings.Contains(outputStr, "Calculating JVM memory") && !strings.Contains(outputStr, "Using specified memory") {
					t.Errorf("Expected memory detection message in output. Got: %s", outputStr)
				}

				// Check if manual override was used
				hasManualMemory := false
				for _, arg := range tt.args {
					if arg == "2G" {
						hasManualMemory = true
						break
					}
				}

				if hasManualMemory {
					// Manual override should show "Using specified memory: 2G"
					if !strings.Contains(outputStr, "Using specified memory: 2G") {
						t.Errorf("Expected manual memory setting (Using specified memory: 2G) in output. Got: %s", outputStr)
					}
				} else {
					// Auto-detection should show some positive memory value
					if !strings.Contains(outputStr, "Calculating JVM memory based on") {
						t.Errorf("Expected auto-detected memory value in output. Got: %s", outputStr)
					}
				}
			}

			// All outputs should contain JVM options
			if !strings.Contains(outputStr, "-Xmx") {
				t.Errorf("Expected JVM max heap option (-Xmx) in output. Got: %s", outputStr)
			}

			if !hasQuiet && !strings.Contains(outputStr, "JAVA_TOOL_OPTIONS") {
				t.Errorf("Expected JAVA_TOOL_OPTIONS in output. Got: %s", outputStr)
			}
		})
	}
}

// Benchmark test for the main application
func BenchmarkMainExecution(b *testing.B) {
	// Build the binary for testing
	binaryPath := filepath.Join(os.TempDir(), "memory-calculator-bench")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/memory-calculator")
	if err := cmd.Run(); err != nil {
		b.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove(binaryPath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binaryPath, "--total-memory", "2G", "--quiet")
		// Set environment variables for testing
		cmd.Env = append(os.Environ(),
			"BPI_APPLICATION_PATH=.",
			"BPL_JVM_LOADED_CLASS_COUNT=30000",
		)
		_, err := cmd.CombinedOutput()
		if err != nil {
			b.Fatalf("Command failed: %v", err)
		}
	}
}
