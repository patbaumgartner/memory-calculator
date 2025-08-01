package calculator

import (
	"os"
	"strings"
	"testing"
)

func TestExecuteWithDefaultValues(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("BPL_JVM_TOTAL_MEMORY")
	os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")
	os.Unsetenv("BPL_JVM_THREAD_COUNT")

	// Create temporary directory for BPI_APPLICATION_PATH
	tempDir, err := os.MkdirTemp("", "memory-calculator-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set app path to temp dir instead of default /app
	os.Setenv("BPI_APPLICATION_PATH", tempDir)
	defer os.Unsetenv("BPI_APPLICATION_PATH")

	mc := Create(true) // quiet mode
	result, err := mc.Execute()
	if err != nil {
		t.Fatal(err)
	}

	// Should return JVM options with defaults
	javaOptions, exists := result["JAVA_TOOL_OPTIONS"]
	if !exists || len(javaOptions) == 0 {
		t.Error("Expected JAVA_TOOL_OPTIONS to be returned")
	}

	// Should contain memory settings
	if !strings.Contains(javaOptions, "-Xmx") {
		t.Error("Expected -Xmx option in result")
	}
}

func TestExecuteWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("BPL_JVM_TOTAL_MEMORY", "1G")
	os.Setenv("BPL_JVM_LOADED_CLASS_COUNT", "5000")
	os.Setenv("BPL_JVM_THREAD_COUNT", "300")
	defer func() {
		os.Unsetenv("BPL_JVM_TOTAL_MEMORY")
		os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")
		os.Unsetenv("BPL_JVM_THREAD_COUNT")
	}()

	mc := Create(true) // quiet mode
	result, err := mc.Execute()
	if err != nil {
		t.Fatal(err)
	}

	// Should return JVM options
	javaOptions, exists := result["JAVA_TOOL_OPTIONS"]
	if !exists || len(javaOptions) == 0 {
		t.Error("Expected JAVA_TOOL_OPTIONS to be returned")
	}
}

func TestExecuteWithClassCounting(t *testing.T) {
	// Clear environment to enable class counting
	os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")

	// Create temporary directory with mock class file
	tempDir, err := os.MkdirTemp("", "memory-calculator-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock class file
	classFile, err := os.Create(tempDir + "/Test.class")
	if err != nil {
		t.Fatal(err)
	}
	classFile.Close()

	// Set app path to our temp dir
	os.Setenv("BPI_APPLICATION_PATH", tempDir)
	defer os.Unsetenv("BPI_APPLICATION_PATH")

	mc := Create(true) // quiet mode
	result, err := mc.Execute()
	if err != nil {
		t.Fatal(err)
	}

	// Should return JVM options
	javaOptions, exists := result["JAVA_TOOL_OPTIONS"]
	if !exists || len(javaOptions) == 0 {
		t.Error("Expected JAVA_TOOL_OPTIONS to be returned")
	}
}

func TestParseMemoryString(t *testing.T) {
	mc := Create(true)

	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"1G", 1073741824, false},
		{"512M", 536870912, false},
		{"1024K", 1048576, false},
		{"2048", 2048, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		result, err := mc.parseMemoryString(test.input)

		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %s", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("Input %s: expected %d, got %d", test.input, test.expected, result)
			}
		}
	}
}

func TestCountAgentClasses(t *testing.T) {
	mc := Create(true)

	// Test with no agent options
	count, err := mc.CountAgentClasses("")
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("Expected 0 agent classes, got %d", count)
	}

	// Test with invalid JAVA_TOOL_OPTIONS (should handle gracefully)
	_, err = mc.CountAgentClasses("-javaagent:/nonexistent/agent.jar")
	if err != nil {
		// This is expected since the jar doesn't exist, but should not panic
		t.Logf("Expected error for non-existent agent jar: %v", err)
	}
}
