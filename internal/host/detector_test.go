package host

import (
	"os"
	"runtime"
	"testing"
)

func TestCreateDetector(t *testing.T) {
	detector := Create()
	if detector.MemInfoPath != LinuxMemInfoPath {
		t.Errorf("Expected MemInfoPath to be %s, got %s", LinuxMemInfoPath, detector.MemInfoPath)
	}
}

func TestCreateDetectorWithPath(t *testing.T) {
	customPath := "/custom/path/meminfo"
	detector := CreateWithPath(customPath)
	if detector.MemInfoPath != customPath {
		t.Errorf("Expected MemInfoPath to be %s, got %s", customPath, detector.MemInfoPath)
	}
}

func TestDetectLinuxMemory(t *testing.T) {
	tests := []struct {
		name           string
		memInfoContent string
		expectedMemory int64 // in bytes
		description    string
	}{
		{
			name: "Valid meminfo with 8GB",
			memInfoContent: `MemTotal:        8062332 kB
MemFree:         1234567 kB
MemAvailable:    2345678 kB`,
			expectedMemory: 8062332 * 1024, // Convert KB to bytes
			description:    "Standard meminfo format",
		},
		{
			name: "Valid meminfo with 16GB",
			memInfoContent: `MemTotal:       16124664 kB
MemFree:         2468135 kB
MemAvailable:    4936271 kB
Buffers:          123456 kB`,
			expectedMemory: 16124664 * 1024,
			description:    "Larger memory system",
		},
		{
			name: "Valid meminfo with 4GB",
			memInfoContent: `MemTotal:        4031166 kB
MemFree:          654321 kB`,
			expectedMemory: 4031166 * 1024,
			description:    "Smaller memory system",
		},
		{
			name:           "Minimal valid meminfo",
			memInfoContent: `MemTotal:        1048576 kB`,
			expectedMemory: 1048576 * 1024, // 1GB
			description:    "Only MemTotal line present",
		},
		{
			name: "MemTotal not first line",
			memInfoContent: `MemFree:         1234567 kB
MemAvailable:    2345678 kB
MemTotal:        8062332 kB
Buffers:          123456 kB`,
			expectedMemory: 8062332 * 1024,
			description:    "MemTotal appears later in file",
		},
		{
			name:           "Invalid format - no fields",
			memInfoContent: `MemTotal:`,
			expectedMemory: 0,
			description:    "Missing memory value",
		},
		{
			name: "Invalid format - non-numeric",
			memInfoContent: `MemTotal:        invalid kB
MemFree:         1234567 kB`,
			expectedMemory: 0,
			description:    "Non-numeric memory value",
		},
		{
			name:           "Empty file",
			memInfoContent: ``,
			expectedMemory: 0,
			description:    "Empty meminfo file",
		},
		{
			name: "No MemTotal line",
			memInfoContent: `MemFree:         1234567 kB
MemAvailable:    2345678 kB
Buffers:          123456 kB`,
			expectedMemory: 0,
			description:    "Missing MemTotal line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file with test content
			tmpFile, err := os.CreateTemp("", "meminfo_test_")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.memInfoContent); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			// Test with custom path
			detector := CreateWithPath(tmpFile.Name())
			memory := detector.detectLinuxMemory()

			if memory != tt.expectedMemory {
				t.Errorf("Expected memory %d bytes, got %d bytes", tt.expectedMemory, memory)
			}
		})
	}
}

func TestDetectLinuxMemoryFileNotFound(t *testing.T) {
	detector := CreateWithPath("/nonexistent/path/meminfo")
	memory := detector.detectLinuxMemory()

	if memory != 0 {
		t.Errorf("Expected 0 for non-existent file, got %d", memory)
	}
}

func TestDetectDarwinMemory(t *testing.T) {
	detector := Create()
	memory := detector.detectDarwinMemory()

	// Darwin memory detection should return some positive value or 0
	// Since it's heuristic-based, we just check it's reasonable
	if memory < 0 {
		t.Errorf("Expected non-negative memory value, got %d", memory)
	}

	// If it returns a value, it should be at least 1GB
	if memory > 0 && memory < 1024*1024*1024 {
		t.Errorf("Expected memory to be at least 1GB if detected, got %d", memory)
	}

	// Should not exceed 128GB (our cap)
	if memory > 128*1024*1024*1024 {
		t.Errorf("Expected memory to be at most 128GB, got %d", memory)
	}
}

func TestDetectHostMemory(t *testing.T) {
	detector := Create()

	// Test based on current OS
	switch runtime.GOOS {
	case "linux":
		// On Linux, try to detect from actual /proc/meminfo if available
		memory := detector.DetectHostMemory()
		if memory < 0 {
			t.Errorf("Expected non-negative memory value on Linux, got %d", memory)
		}
		// If we're actually on Linux and can read /proc/meminfo, we should get a positive value
		if _, err := os.Stat("/proc/meminfo"); err == nil && memory == 0 {
			t.Log("Warning: /proc/meminfo exists but couldn't read memory (might be expected in containers)")
		}

	case "darwin":
		// On macOS, we should get a heuristic value
		memory := detector.DetectHostMemory()
		if memory < 0 {
			t.Errorf("Expected non-negative memory value on %s, got %d", runtime.GOOS, memory)
		}

	default:
		// On unsupported platforms, we should get 0
		memory := detector.DetectHostMemory()
		if memory != 0 {
			t.Errorf("Expected 0 on unsupported platform %s, got %d", runtime.GOOS, memory)
		}
	}
}

func TestDetectHostMemoryWithCustomLinuxFile(t *testing.T) {
	// Skip if not running the full test suite
	if testing.Short() {
		t.Skip("Skipping custom file test in short mode")
	}

	// Test with a custom meminfo file to ensure Linux detection works
	memInfoContent := `MemTotal:        8062332 kB
MemFree:         1234567 kB
MemAvailable:    2345678 kB`

	tmpFile, err := os.CreateTemp("", "meminfo_integration_")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(memInfoContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	detector := CreateWithPath(tmpFile.Name())

	// Force Linux detection by calling detectLinuxMemory directly
	memory := detector.detectLinuxMemory()
	expectedMemory := int64(8062332 * 1024)

	if memory != expectedMemory {
		t.Errorf("Expected memory %d bytes, got %d bytes", expectedMemory, memory)
	}
}

func TestIsHostMemoryDetectionSupported(t *testing.T) {
	supported := IsHostMemoryDetectionSupported()

	switch runtime.GOOS {
	case "linux", "darwin":
		if !supported {
			t.Errorf("Expected host memory detection to be supported on %s", runtime.GOOS)
		}
	default:
		if supported {
			t.Errorf("Expected host memory detection to be unsupported on %s", runtime.GOOS)
		}
	}
}

func TestPlatformSpecificBehavior(t *testing.T) {
	detector := Create()

	t.Run("Current platform detection", func(t *testing.T) {
		memory := detector.DetectHostMemory()
		supported := IsHostMemoryDetectionSupported()

		if supported {
			// On supported platforms, we should get either 0 (detection failed) or a positive value
			if memory < 0 {
				t.Errorf("Expected non-negative memory on supported platform %s, got %d", runtime.GOOS, memory)
			}
		} else {
			// On unsupported platforms, we should always get 0
			if memory != 0 {
				t.Errorf("Expected 0 on unsupported platform %s, got %d", runtime.GOOS, memory)
			}
		}
	})
}

func TestMemoryDetectionRealWorldScenarios(t *testing.T) {
	tests := []struct {
		name           string
		memInfoContent string
		expectedMemory int64
		description    string
	}{
		{
			name: "Small VPS (1GB)",
			memInfoContent: `MemTotal:        1048576 kB
MemFree:          123456 kB
MemAvailable:     234567 kB`,
			expectedMemory: 1048576 * 1024,
			description:    "Typical small VPS configuration",
		},
		{
			name: "Development laptop (16GB)",
			memInfoContent: `MemTotal:       16777216 kB
MemFree:         2097152 kB
MemAvailable:    8388608 kB`,
			expectedMemory: 16777216 * 1024,
			description:    "Typical developer laptop",
		},
		{
			name: "High-end server (64GB)",
			memInfoContent: `MemTotal:       67108864 kB
MemFree:         8388608 kB
MemAvailable:   33554432 kB`,
			expectedMemory: 67108864 * 1024,
			description:    "High-end server configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "meminfo_realworld_")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.memInfoContent); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			detector := CreateWithPath(tmpFile.Name())
			memory := detector.detectLinuxMemory()

			if memory != tt.expectedMemory {
				t.Errorf("Expected memory %d bytes, got %d bytes for %s", tt.expectedMemory, memory, tt.description)
			}
		})
	}
}
