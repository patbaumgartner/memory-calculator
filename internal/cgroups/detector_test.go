package cgroups

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/patbaumgartner/memory-calculator/pkg/errors"
)

func TestNewDetector(t *testing.T) {
	detector := NewDetector()

	if detector.CgroupsV2Path != "/sys/fs/cgroup/memory.max" {
		t.Errorf("Expected CgroupsV2Path='/sys/fs/cgroup/memory.max', got %s", detector.CgroupsV2Path)
	}

	if detector.CgroupsV1Path != "/sys/fs/cgroup/memory/memory.limit_in_bytes" {
		t.Errorf("Expected CgroupsV1Path='/sys/fs/cgroup/memory/memory.limit_in_bytes', got %s", detector.CgroupsV1Path)
	}
}

func TestNewDetectorWithPaths(t *testing.T) {
	v2Path := "/custom/v2/path"
	v1Path := "/custom/v1/path"
	detector := NewDetectorWithPaths(v2Path, v1Path)

	if detector.CgroupsV2Path != v2Path {
		t.Errorf("Expected CgroupsV2Path='%s', got %s", v2Path, detector.CgroupsV2Path)
	}

	if detector.CgroupsV1Path != v1Path {
		t.Errorf("Expected CgroupsV1Path='%s', got %s", v1Path, detector.CgroupsV1Path)
	}
}

func TestReadCgroupsV2(t *testing.T) {
	// Create temporary test files
	tempDir, err := os.MkdirTemp("", "cgroups_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		fileContent string
		expected    int64
		expectError bool
		errorCode   errors.ErrorCode
	}{
		{
			name:        "Valid memory limit",
			fileContent: "2147483648\n",
			expected:    2147483648,
			expectError: false,
		},
		{
			name:        "No limit (max)",
			fileContent: "max\n",
			expected:    0,
			expectError: false,
		},
		{
			name:        "Large unrealistic limit",
			fileContent: "9223372036854775807\n", // Very large number
			expected:    0,
			expectError: false,
		},
		{
			name:        "Zero limit",
			fileContent: "0\n",
			expected:    0,
			expectError: false,
		},
		{
			name:        "Invalid format",
			fileContent: "invalid\n",
			expected:    0,
			expectError: true,
			errorCode:   errors.ErrCgroupsAccess,
		},
		{
			name:        "Empty file",
			fileContent: "",
			expected:    0,
			expectError: true,
			errorCode:   errors.ErrCgroupsAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, "memory.max")
			if tt.fileContent != "" {
				err := os.WriteFile(testFile, []byte(tt.fileContent), 0o644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			}

			detector := NewDetectorWithPaths(testFile, "")
			result, err := detector.readCgroupsV2()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if mcErr, ok := err.(*errors.MemoryCalculatorError); ok {
					if mcErr.Code != tt.errorCode {
						t.Errorf("Expected error code %v, got %v", tt.errorCode, mcErr.Code)
					}
				} else {
					t.Errorf("Expected MemoryCalculatorError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}

			// Clean up test file
			os.Remove(testFile)
		})
	}
}

func TestReadCgroupsV1(t *testing.T) {
	// Create temporary test files
	tempDir, err := os.MkdirTemp("", "cgroups_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		fileContent string
		expected    int64
		expectError bool
		errorCode   errors.ErrorCode
	}{
		{
			name:        "Valid memory limit",
			fileContent: "2147483648\n",
			expected:    2147483648,
			expectError: false,
		},
		{
			name:        "Large unrealistic limit",
			fileContent: "9223372036854775807\n", // Very large number (no limit)
			expected:    0,
			expectError: false,
		},
		{
			name:        "Zero limit",
			fileContent: "0\n",
			expected:    0,
			expectError: false,
		},
		{
			name:        "Invalid format",
			fileContent: "invalid\n",
			expected:    0,
			expectError: true,
			errorCode:   errors.ErrCgroupsAccess,
		},
		{
			name:        "Empty file",
			fileContent: "",
			expected:    0,
			expectError: true,
			errorCode:   errors.ErrCgroupsAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, "memory.limit_in_bytes")
			if tt.fileContent != "" {
				err := os.WriteFile(testFile, []byte(tt.fileContent), 0o644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			}

			detector := NewDetectorWithPaths("", testFile)
			result, err := detector.readCgroupsV1()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if mcErr, ok := err.(*errors.MemoryCalculatorError); ok {
					if mcErr.Code != tt.errorCode {
						t.Errorf("Expected error code %v, got %v", tt.errorCode, mcErr.Code)
					}
				} else {
					t.Errorf("Expected MemoryCalculatorError, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}

			// Clean up test file
			os.Remove(testFile)
		})
	}
}

func TestDetectContainerMemory(t *testing.T) {
	// Create temporary test files
	tempDir, err := os.MkdirTemp("", "cgroups_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name          string
		v2FileContent string
		v1FileContent string
		createV2File  bool
		createV1File  bool
		expected      int64
	}{
		{
			name:          "V2 available with valid limit",
			v2FileContent: "2147483648\n",
			createV2File:  true,
			expected:      2147483648,
		},
		{
			name:          "V2 available but no limit, V1 has limit",
			v2FileContent: "max\n",
			v1FileContent: "1073741824\n",
			createV2File:  true,
			createV1File:  true,
			expected:      1073741824,
		},
		{
			name:          "Only V1 available",
			v1FileContent: "1073741824\n",
			createV1File:  true,
			expected:      1073741824,
		},
		{
			name:     "No cgroups files available",
			expected: 0,
		},
		{
			name:          "V2 has unrealistic limit, V1 has valid limit",
			v2FileContent: "9223372036854775807\n",
			v1FileContent: "1073741824\n",
			createV2File:  true,
			createV1File:  true,
			expected:      1073741824,
		},
		{
			name:          "Both files have unrealistic limits",
			v2FileContent: "9223372036854775807\n",
			v1FileContent: "9223372036854775807\n",
			createV2File:  true,
			createV1File:  true,
			expected:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v2File := filepath.Join(tempDir, "memory.max")
			v1File := filepath.Join(tempDir, "memory.limit_in_bytes")

			// Create test files if needed
			if tt.createV2File {
				err := os.WriteFile(v2File, []byte(tt.v2FileContent), 0o644)
				if err != nil {
					t.Fatalf("Failed to write V2 test file: %v", err)
				}
			}

			if tt.createV1File {
				err := os.WriteFile(v1File, []byte(tt.v1FileContent), 0o644)
				if err != nil {
					t.Fatalf("Failed to write V1 test file: %v", err)
				}
			}

			detector := NewDetectorWithPaths(v2File, v1File)
			result := detector.DetectContainerMemory()

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}

			// Clean up test files
			os.Remove(v2File)
			os.Remove(v1File)
		})
	}
}

func TestFileNotFound(t *testing.T) {
	detector := NewDetectorWithPaths("/nonexistent/v2/path", "/nonexistent/v1/path")

	// Should return 0 when files don't exist
	result := detector.DetectContainerMemory()
	if result != 0 {
		t.Errorf("Expected 0 for nonexistent files, got %d", result)
	}
}

func TestMaxRealisticMemoryConstant(t *testing.T) {
	expected := int64(1024 * 1024 * 1024 * 1024) // 1TB
	if MaxRealisticMemory != expected {
		t.Errorf("Expected MaxRealisticMemory=%d, got %d", expected, MaxRealisticMemory)
	}
}

// Benchmark tests
func BenchmarkDetectContainerMemory(b *testing.B) {
	// Create temporary test file
	tempDir, err := os.MkdirTemp("", "cgroups_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	v2File := filepath.Join(tempDir, "memory.max")
	err = os.WriteFile(v2File, []byte("2147483648\n"), 0o644)
	if err != nil {
		b.Fatalf("Failed to write test file: %v", err)
	}

	detector := NewDetectorWithPaths(v2File, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.DetectContainerMemory()
	}
}
