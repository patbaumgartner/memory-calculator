package constants

import "testing"

func TestApplicationConstants(t *testing.T) {
	// Test application metadata constants
	if ApplicationName == "" {
		t.Error("ApplicationName should not be empty")
	}

	if DefaultVersion != "dev" {
		t.Errorf("Expected DefaultVersion to be 'dev', got '%s'", DefaultVersion)
	}

	if UnknownValue != "unknown" {
		t.Errorf("Expected UnknownValue to be 'unknown', got '%s'", UnknownValue)
	}
}

func TestDefaultValues(t *testing.T) {
	// Test default configuration values
	if DefaultThreadCount != "250" {
		t.Errorf("Expected DefaultThreadCount to be '250', got '%s'", DefaultThreadCount)
	}

	if DefaultHeadRoom != "0" {
		t.Errorf("Expected DefaultHeadRoom to be '0', got '%s'", DefaultHeadRoom)
	}

	if DefaultApplicationPath != "/app" {
		t.Errorf("Expected DefaultApplicationPath to be '/app', got '%s'", DefaultApplicationPath)
	}
}

func TestEnvironmentVariableNames(t *testing.T) {
	expectedEnvVars := map[string]string{
		"EnvTotalMemory":      "BPL_JVM_TOTAL_MEMORY",
		"EnvThreadCount":      "BPL_JVM_THREAD_COUNT",
		"EnvLoadedClassCount": "BPL_JVM_LOADED_CLASS_COUNT",
		"EnvHeadRoom":         "BPL_JVM_HEAD_ROOM",
		"EnvApplicationPath":  "BPI_APPLICATION_PATH",
		"EnvJVMClassCount":    "BPI_JVM_CLASS_COUNT",
		"EnvQuiet":            "QUIET",
	}

	actualValues := map[string]string{
		"EnvTotalMemory":      EnvTotalMemory,
		"EnvThreadCount":      EnvThreadCount,
		"EnvLoadedClassCount": EnvLoadedClassCount,
		"EnvHeadRoom":         EnvHeadRoom,
		"EnvApplicationPath":  EnvApplicationPath,
		"EnvJVMClassCount":    EnvJVMClassCount,
		"EnvQuiet":            EnvQuiet,
	}

	for name, expected := range expectedEnvVars {
		if actual := actualValues[name]; actual != expected {
			t.Errorf("Expected %s to be '%s', got '%s'", name, expected, actual)
		}
	}
}

func TestSystemPaths(t *testing.T) {
	// Test system path constants
	if DefaultMemoryLimitPathV1 != "/sys/fs/cgroup/memory/memory.limit_in_bytes" {
		t.Errorf("DefaultMemoryLimitPathV1 has unexpected value: %s", DefaultMemoryLimitPathV1)
	}

	if DefaultMemoryLimitPathV2 != "/sys/fs/cgroup/memory.max" {
		t.Errorf("DefaultMemoryLimitPathV2 has unexpected value: %s", DefaultMemoryLimitPathV2)
	}

	if DefaultMemoryInfoPath != "/proc/meminfo" {
		t.Errorf("DefaultMemoryInfoPath has unexpected value: %s", DefaultMemoryInfoPath)
	}
}

func TestMemoryLimits(t *testing.T) {
	// Test memory limit constants
	if MaxJVMSizeGB != 64 {
		t.Errorf("Expected MaxJVMSizeGB to be 64, got %d", MaxJVMSizeGB)
	}

	// Test that MaxRealisticMemoryBytes is a very large number (1 PB)
	expectedMaxRealistic := int64(1 << 50) // 1 PB
	if MaxRealisticMemoryBytes != expectedMaxRealistic {
		t.Errorf("Expected MaxRealisticMemoryBytes to be %d, got %d", expectedMaxRealistic, MaxRealisticMemoryBytes)
	}

	if MinValidMemoryBytes != 1024 {
		t.Errorf("Expected MinValidMemoryBytes to be 1024, got %d", MinValidMemoryBytes)
	}

	// Validate memory limits make sense
	if MinValidMemoryBytes >= MaxRealisticMemoryBytes {
		t.Error("MinValidMemoryBytes should be less than MaxRealisticMemoryBytes")
	}
}

func TestOutputFormatting(t *testing.T) {
	// Test output formatting constants
	if HelpSeparator != "====================" {
		t.Errorf("Expected HelpSeparator to be 20 equal signs, got '%s'", HelpSeparator)
	}

	if len(HelpSeparator) != 20 {
		t.Errorf("Expected HelpSeparator to be 20 characters long, got %d", len(HelpSeparator))
	}

	if IndentationSpaces != "  " {
		t.Errorf("Expected IndentationSpaces to be '  ', got '%s'", IndentationSpaces)
	}

	if len(IndentationSpaces) != 2 {
		t.Errorf("Expected IndentationSpaces to be 2 spaces, got %d characters", len(IndentationSpaces))
	}
}

func TestConstantTypes(t *testing.T) {
	// Test that constants have expected types
	var _ string = ApplicationName
	var _ string = DefaultVersion
	var _ string = UnknownValue
	var _ string = DefaultThreadCount
	var _ string = DefaultHeadRoom
	var _ string = DefaultApplicationPath
	var _ string = EnvTotalMemory
	var _ string = EnvThreadCount
	var _ string = EnvLoadedClassCount
	var _ string = EnvHeadRoom
	var _ string = EnvApplicationPath
	var _ string = EnvJVMClassCount
	var _ string = EnvQuiet
	var _ string = DefaultMemoryLimitPathV1
	var _ string = DefaultMemoryLimitPathV2
	var _ string = DefaultMemoryInfoPath
	var _ int = MaxJVMSizeGB
	var _ int64 = MaxRealisticMemoryBytes
	var _ int64 = MinValidMemoryBytes
	var _ string = HelpSeparator
	var _ string = IndentationSpaces
}

func TestConstantImmutability(t *testing.T) {
	// While Go doesn't enforce true immutability for constants of complex types,
	// we can at least verify that our string constants haven't been accidentally modified

	// Store original values
	origAppName := ApplicationName
	origDefaultVer := DefaultVersion

	// These should remain the same throughout the test
	if ApplicationName != origAppName {
		t.Error("ApplicationName constant should not change during execution")
	}

	if DefaultVersion != origDefaultVer {
		t.Error("DefaultVersion constant should not change during execution")
	}
}

func TestEnvironmentVariableNameConsistency(t *testing.T) {
	// Test that environment variable names follow expected patterns
	bplVars := []string{EnvTotalMemory, EnvThreadCount, EnvLoadedClassCount, EnvHeadRoom}
	bpiVars := []string{EnvApplicationPath, EnvJVMClassCount}

	// BPL variables should start with "BPL_"
	for _, envVar := range bplVars {
		if len(envVar) < 4 || envVar[:4] != "BPL_" {
			t.Errorf("BPL environment variable '%s' should start with 'BPL_'", envVar)
		}
	}

	// BPI variables should start with "BPI_"
	for _, envVar := range bpiVars {
		if len(envVar) < 4 || envVar[:4] != "BPI_" {
			t.Errorf("BPI environment variable '%s' should start with 'BPI_'", envVar)
		}
	}
}

func TestMemoryConstantRelationships(t *testing.T) {
	// Test logical relationships between memory constants

	// Max JVM size in bytes should be MaxJVMSizeGB * 1GB
	expectedMaxJVMBytes := int64(MaxJVMSizeGB) * (1024 * 1024 * 1024)

	// MaxRealisticMemoryBytes should be much larger than reasonable JVM sizes
	if MaxRealisticMemoryBytes <= expectedMaxJVMBytes {
		t.Errorf("MaxRealisticMemoryBytes (%d) should be larger than max JVM size (%d)",
			MaxRealisticMemoryBytes, expectedMaxJVMBytes)
	}

	// MinValidMemoryBytes should be much smaller than max JVM size
	if MinValidMemoryBytes >= expectedMaxJVMBytes {
		t.Errorf("MinValidMemoryBytes (%d) should be much smaller than max JVM size (%d)",
			MinValidMemoryBytes, expectedMaxJVMBytes)
	}
}

func TestPathConstants(t *testing.T) {
	// Test that path constants are absolute paths
	paths := []string{
		DefaultMemoryLimitPathV1,
		DefaultMemoryLimitPathV2,
		DefaultMemoryInfoPath,
		DefaultApplicationPath,
	}

	for _, path := range paths {
		if len(path) == 0 || path[0] != '/' {
			t.Errorf("Path constant '%s' should be an absolute path starting with '/'", path)
		}
	}
}
