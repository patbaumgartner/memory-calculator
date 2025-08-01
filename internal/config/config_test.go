package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Clear any existing environment variables
	os.Unsetenv("BPL_JVM_TOTAL_MEMORY")
	os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")
	os.Unsetenv("BPL_JVM_THREAD_COUNT")
	os.Unsetenv("BPL_JVM_HEAD_ROOM")

	cfg := Load()

	// Test default values
	if cfg.ThreadCount != "250" {
		t.Errorf("Expected thread count '250', got '%s'", cfg.ThreadCount)
	}

	if cfg.LoadedClassCount != "" {
		t.Errorf("Expected empty loaded class count (should be calculated), got '%s'", cfg.LoadedClassCount)
	}

	if cfg.HeadRoom != "0" {
		t.Errorf("Expected head room '0', got '%s'", cfg.HeadRoom)
	}

	if cfg.BuildVersion != "dev" {
		t.Errorf("Expected build version 'dev', got '%s'", cfg.BuildVersion)
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("BPL_JVM_LOADED_CLASS_COUNT", "15000")
	os.Setenv("BPL_JVM_THREAD_COUNT", "500")
	os.Setenv("BPL_JVM_HEAD_ROOM", "10")

	defer func() {
		os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")
		os.Unsetenv("BPL_JVM_THREAD_COUNT")
		os.Unsetenv("BPL_JVM_HEAD_ROOM")
	}()

	cfg := Load()

	if cfg.LoadedClassCount != "15000" {
		t.Errorf("Expected loaded class count '15000', got '%s'", cfg.LoadedClassCount)
	}

	if cfg.ThreadCount != "500" {
		t.Errorf("Expected thread count '500', got '%s'", cfg.ThreadCount)
	}

	if cfg.HeadRoom != "10" {
		t.Errorf("Expected head room '10', got '%s'", cfg.HeadRoom)
	}
}

func TestConfigValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "Valid config with defaults",
			config: &Config{
				ThreadCount:      "250",
				LoadedClassCount: "", // empty is valid
				HeadRoom:         "0",
			},
			expectError: false,
		},
		{
			name: "Valid config with values",
			config: &Config{
				ThreadCount:      "300",
				LoadedClassCount: "5000",
				HeadRoom:         "5",
			},
			expectError: false,
		},
		{
			name: "Invalid thread count - negative",
			config: &Config{
				ThreadCount:      "-1",
				LoadedClassCount: "1000",
				HeadRoom:         "0",
			},
			expectError: true,
		},
		{
			name: "Invalid thread count - not a number",
			config: &Config{
				ThreadCount:      "abc",
				LoadedClassCount: "1000",
				HeadRoom:         "0",
			},
			expectError: true,
		},
		{
			name: "Invalid loaded class count - negative",
			config: &Config{
				ThreadCount:      "250",
				LoadedClassCount: "-1",
				HeadRoom:         "0",
			},
			expectError: true,
		},
		{
			name: "Invalid head room - negative",
			config: &Config{
				ThreadCount:      "250",
				LoadedClassCount: "1000",
				HeadRoom:         "-1",
			},
			expectError: true,
		},
		{
			name: "Invalid head room - over 100",
			config: &Config{
				ThreadCount:      "250",
				LoadedClassCount: "1000",
				HeadRoom:         "101",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestSetEnvironmentVariables(t *testing.T) {
	cfg := &Config{
		ThreadCount:      "300",
		LoadedClassCount: "40000",
		HeadRoom:         "15",
	}

	cfg.SetEnvironmentVariables()

	if os.Getenv("BPL_JVM_THREAD_COUNT") != "300" {
		t.Errorf("Expected BPL_JVM_THREAD_COUNT=300, got %s", os.Getenv("BPL_JVM_THREAD_COUNT"))
	}

	if os.Getenv("BPL_JVM_LOADED_CLASS_COUNT") != "40000" {
		t.Errorf("Expected BPL_JVM_LOADED_CLASS_COUNT=40000, got %s", os.Getenv("BPL_JVM_LOADED_CLASS_COUNT"))
	}

	if os.Getenv("BPL_JVM_HEAD_ROOM") != "15" {
		t.Errorf("Expected BPL_JVM_HEAD_ROOM=15, got %s", os.Getenv("BPL_JVM_HEAD_ROOM"))
	}

	// Clean up
	os.Unsetenv("BPL_JVM_THREAD_COUNT")
	os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")
	os.Unsetenv("BPL_JVM_HEAD_ROOM")
}

func TestSetTotalMemory(t *testing.T) {
	cfg := &Config{}

	// Test with positive memory
	cfg.SetTotalMemory(2147483648) // 2GB
	if os.Getenv("BPL_JVM_TOTAL_MEMORY") != "2147483648" {
		t.Errorf("Expected BPL_JVM_TOTAL_MEMORY=2147483648, got %s", os.Getenv("BPL_JVM_TOTAL_MEMORY"))
	}

	// Clean up
	os.Unsetenv("BPL_JVM_TOTAL_MEMORY")

	// Test with zero memory (should not set env var)
	cfg.SetTotalMemory(0)
	if os.Getenv("BPL_JVM_TOTAL_MEMORY") != "" {
		t.Errorf("Expected BPL_JVM_TOTAL_MEMORY to be unset, got %s", os.Getenv("BPL_JVM_TOTAL_MEMORY"))
	}
}
