package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestReadCgroupsV1(t *testing.T) {
	// Create a temporary directory for mock cgroups files
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		fileContent   string
		createFile    bool
		expectedValue int64
	}{
		{
			name:          "Valid memory limit",
			fileContent:   "2147483648\n",
			createFile:    true,
			expectedValue: 2147483648,
		},
		{
			name:          "Large unrealistic limit (no limit set)",
			fileContent:   "9223372036854775807\n",
			createFile:    true,
			expectedValue: 0, // Should return 0 for unrealistic limits
		},
		{
			name:          "File doesn't exist",
			createFile:    false,
			expectedValue: 0,
		},
		{
			name:          "Invalid content",
			fileContent:   "invalid\n",
			createFile:    true,
			expectedValue: 0,
		},
		{
			name:          "Empty file",
			fileContent:   "",
			createFile:    true,
			expectedValue: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock cgroups directory structure
			cgroupDir := filepath.Join(tempDir, "sys", "fs", "cgroup", "memory")
			err := os.MkdirAll(cgroupDir, 0o755)
			if err != nil {
				t.Fatalf("Failed to create mock cgroup directory: %v", err)
			}

			filePath := filepath.Join(cgroupDir, "memory.limit_in_bytes")

			if tt.createFile {
				err := os.WriteFile(filePath, []byte(tt.fileContent), 0o644)
				if err != nil {
					t.Fatalf("Failed to create mock cgroup file: %v", err)
				}
			}

			result := readCgroupsV1FromPath(filePath)
			if result != tt.expectedValue {
				t.Errorf("Expected %d, got %d", tt.expectedValue, result)
			}
		})
	}
}

func TestReadCgroupsV2(t *testing.T) {
	// Create a temporary directory for mock cgroups files
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		fileContent   string
		createFile    bool
		expectedValue int64
	}{
		{
			name:          "Valid memory limit",
			fileContent:   "1073741824\n",
			createFile:    true,
			expectedValue: 1073741824,
		},
		{
			name:          "Max value (no limit)",
			fileContent:   "max\n",
			createFile:    true,
			expectedValue: 0, // Should return 0 for "max"
		},
		{
			name:          "File doesn't exist",
			createFile:    false,
			expectedValue: 0,
		},
		{
			name:          "Invalid content",
			fileContent:   "invalid\n",
			createFile:    true,
			expectedValue: 0,
		},
		{
			name:          "Empty file",
			fileContent:   "",
			createFile:    true,
			expectedValue: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock cgroups directory structure
			cgroupDir := filepath.Join(tempDir, "sys", "fs", "cgroup")
			err := os.MkdirAll(cgroupDir, 0o755)
			if err != nil {
				t.Fatalf("Failed to create mock cgroup directory: %v", err)
			}

			filePath := filepath.Join(cgroupDir, "memory.max")

			if tt.createFile {
				err := os.WriteFile(filePath, []byte(tt.fileContent), 0o644)
				if err != nil {
					t.Fatalf("Failed to create mock cgroup file: %v", err)
				}
			}

			result := readCgroupsV2FromPath(filePath)
			if result != tt.expectedValue {
				t.Errorf("Expected %d, got %d", tt.expectedValue, result)
			}
		})
	}
}

// Helper functions for testing with custom paths
func readCgroupsV1FromPath(path string) int64 {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	var line string
	_, err = fmt.Fscanf(file, "%s", &line)
	if err != nil {
		return 0
	}

	memory, err := strconv.ParseInt(strings.TrimSpace(line), 10, 64)
	if err != nil {
		return 0
	}

	// Check if it's a realistic limit (not the "no limit" value)
	if memory < 1024*1024*1024*1024 { // Less than 1TB
		return memory
	}
	return 0
}

func readCgroupsV2FromPath(path string) int64 {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	var line string
	_, err = fmt.Fscanf(file, "%s", &line)
	if err != nil {
		return 0
	}

	line = strings.TrimSpace(line)
	if line == "max" {
		return 0 // No limit set
	}

	if memory, err := strconv.ParseInt(line, 10, 64); err == nil {
		return memory
	}
	return 0
}
