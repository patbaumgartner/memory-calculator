package calculator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patbaumgartner/memory-calculator/internal/calc"
)

func TestExecuteWithDefaultValues(t *testing.T) {
	// Clear environment variables
	_ = os.Unsetenv("BPL_JVM_TOTAL_MEMORY")
	_ = os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")
	_ = os.Unsetenv("BPL_JVM_THREAD_COUNT")

	// Create temporary directory for BPI_APPLICATION_PATH
	tempDir, err := os.MkdirTemp("", "memory-calculator-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Set app path to temp dir instead of default /app
	_ = os.Setenv("BPI_APPLICATION_PATH", tempDir)
	defer func() { _ = os.Unsetenv("BPI_APPLICATION_PATH") }()

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
	_ = os.Setenv("BPL_JVM_TOTAL_MEMORY", "1G")
	_ = os.Setenv("BPL_JVM_LOADED_CLASS_COUNT", "5000")
	_ = os.Setenv("BPL_JVM_THREAD_COUNT", "300")
	defer func() {
		_ = os.Unsetenv("BPL_JVM_TOTAL_MEMORY")
		_ = os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")
		_ = os.Unsetenv("BPL_JVM_THREAD_COUNT")
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
	_ = os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")

	// Create temporary directory with mock class file
	tempDir, err := os.MkdirTemp("", "memory-calculator-test")
	if err != nil {
		t.Fatal(err)
	}
	//nolint:gosec // Safe in tests
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a mock class file
	//nolint:gosec // Safe in tests
	classFile, err := os.Create(filepath.Join(tempDir, "Test.class"))
	if err != nil {
		t.Fatal(err)
	}
	_ = classFile.Close()

	// Set app path to our temp dir
	_ = os.Setenv("BPI_APPLICATION_PATH", tempDir)
	defer func() { _ = os.Unsetenv("BPI_APPLICATION_PATH") }()

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

func TestParseHeadroomConfig(t *testing.T) {
	mc := Create(true)

	t.Run("Default Headroom", func(t *testing.T) {
		// No environment variables set
		_ = os.Unsetenv("BPL_JVM_HEAD_ROOM")
		_ = os.Unsetenv("BPL_JVM_HEADROOM")

		c := &calc.Calculator{HeadRoom: DefaultHeadroom}
		err := mc.parseHeadroomConfig(c)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if c.HeadRoom != DefaultHeadroom {
			t.Errorf("Expected headroom %d, got %d", DefaultHeadroom, c.HeadRoom)
		}
	})

	t.Run("Standard Headroom Variable", func(t *testing.T) {
		_ = os.Setenv("BPL_JVM_HEAD_ROOM", "5")
		defer func() { _ = os.Unsetenv("BPL_JVM_HEAD_ROOM") }()

		c := &calc.Calculator{HeadRoom: DefaultHeadroom}
		err := mc.parseHeadroomConfig(c)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if c.HeadRoom != 5 {
			t.Errorf("Expected headroom 5, got %d", c.HeadRoom)
		}
	})

	t.Run("Deprecated Headroom Variable", func(t *testing.T) {
		_ = os.Setenv("BPL_JVM_HEADROOM", "10")
		defer func() { _ = os.Unsetenv("BPL_JVM_HEADROOM") }()

		c := &calc.Calculator{HeadRoom: DefaultHeadroom}
		err := mc.parseHeadroomConfig(c)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if c.HeadRoom != 10 {
			t.Errorf("Expected headroom 10, got %d", c.HeadRoom)
		}
	})

	t.Run("Standard Precedence Over Deprecated", func(t *testing.T) {
		_ = os.Setenv("BPL_JVM_HEAD_ROOM", "5")
		_ = os.Setenv("BPL_JVM_HEADROOM", "10")
		defer func() {
			_ = os.Unsetenv("BPL_JVM_HEAD_ROOM")
			_ = os.Unsetenv("BPL_JVM_HEADROOM")
		}()

		c := &calc.Calculator{HeadRoom: DefaultHeadroom}
		err := mc.parseHeadroomConfig(c)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if c.HeadRoom != 5 {
			t.Errorf("Expected headroom 5 (standard), got %d", c.HeadRoom)
		}
	})

	t.Run("Invalid Value", func(t *testing.T) {
		_ = os.Setenv("BPL_JVM_HEAD_ROOM", "invalid")
		defer func() { _ = os.Unsetenv("BPL_JVM_HEAD_ROOM") }()

		c := &calc.Calculator{HeadRoom: DefaultHeadroom}
		err := mc.parseHeadroomConfig(c)
		if err == nil {
			t.Error("Expected error for invalid headroom value, got nil")
		}
	})
}

func TestParseClassCountConfig(t *testing.T) {
	mc := Create(true)

	// Create mock app directory
	tempDir, err := os.MkdirTemp("", "class-count-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Default environment setup
	setupEnv := func() {
		_ = os.Setenv("BPI_APPLICATION_PATH", tempDir)
	}
	cleanupEnv := func() {
		_ = os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")
		_ = os.Unsetenv("BPI_APPLICATION_PATH")
		_ = os.Unsetenv("BPI_JVM_CLASS_COUNT")
		_ = os.Unsetenv("BPI_CLASS_ADJUSTMENT_FACTOR")
		_ = os.Unsetenv("BPI_CLASS_STATIC_ADJUSTMENT")
	}

	t.Run("Direct Override", func(t *testing.T) {
		cleanupEnv()
		_ = os.Setenv("BPL_JVM_LOADED_CLASS_COUNT", "5000")
		defer cleanupEnv()

		c := &calc.Calculator{}
		err := mc.parseClassCountConfig(c, "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if c.LoadedClassCount != 5000 {
			t.Errorf("Expected class count 5000, got %d", c.LoadedClassCount)
		}
	})

	t.Run("Calculation with defaults", func(t *testing.T) {
		setupEnv()
		defer cleanupEnv()

		// Create a class file
		//nolint:gosec // Safe in tests
		classFile, _ := os.Create(filepath.Join(tempDir, "Test.class"))
		_ = classFile.Close()

		// Calculation: (JVM(1000) + App(1) + Agent(0) + Static(0)) * Factor(1.0) * LoadFactor(0.35)
		// = 1001 * 0.35 = 350.35 -> 350

		c := &calc.Calculator{}
		err := mc.parseClassCountConfig(c, "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check range due to float arithmetic potential minor diffs, though here it should be exact
		if c.LoadedClassCount != 350 {
			t.Errorf("Expected class count ~350, got %d", c.LoadedClassCount)
		}
	})

	t.Run("Calculation with Adjustments", func(t *testing.T) {
		setupEnv()
		_ = os.Setenv("BPI_JVM_CLASS_COUNT", "2000")
		_ = os.Setenv("BPI_CLASS_ADJUSTMENT_FACTOR", "150") // 1.5x
		_ = os.Setenv("BPI_CLASS_STATIC_ADJUSTMENT", "100")
		defer cleanupEnv()

		// Create a class file
		//nolint:gosec // Safe in tests
		classFile, _ := os.Create(filepath.Join(tempDir, "Test.class"))
		_ = classFile.Close()

		// Calculation: (JVM(2000) + App(1) + Agent(0) + Static(100)) * Factor(1.5) * LoadFactor(0.35)
		// = 2101 * 1.5 * 0.35 = 3151.5 * 0.35 = 1103.025 -> 1103

		c := &calc.Calculator{}
		err := mc.parseClassCountConfig(c, "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if c.LoadedClassCount != 1103 {
			t.Errorf("Expected class count ~1103, got %d", c.LoadedClassCount)
		}
	})
}
