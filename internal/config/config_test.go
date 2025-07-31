package config

import (
	"os"
	"testing"

	"github.com/patbaumgartner/memory-calculator/pkg/errors"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ThreadCount != "250" {
		t.Errorf("Expected ThreadCount 250, got %s", cfg.ThreadCount)
	}

	if cfg.LoadedClassCount != "35000" {
		t.Errorf("Expected LoadedClassCount 35000, got %s", cfg.LoadedClassCount)
	}

	if cfg.HeadRoom != "0" {
		t.Errorf("Expected HeadRoom 0, got %s", cfg.HeadRoom)
	}

	if cfg.BuildVersion != "dev" {
		t.Errorf("Expected BuildVersion dev, got %s", cfg.BuildVersion)
	}
}

func TestDefaultConfigWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("BPL_JVM_THREAD_COUNT", "500")
	os.Setenv("BPL_JVM_LOADED_CLASS_COUNT", "50000")
	os.Setenv("BPL_JVM_HEAD_ROOM", "10")
	defer func() {
		os.Unsetenv("BPL_JVM_THREAD_COUNT")
		os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")
		os.Unsetenv("BPL_JVM_HEAD_ROOM")
	}()

	cfg := DefaultConfig()

	if cfg.ThreadCount != "500" {
		t.Errorf("Expected ThreadCount 500, got %s", cfg.ThreadCount)
	}

	if cfg.LoadedClassCount != "50000" {
		t.Errorf("Expected LoadedClassCount 50000, got %s", cfg.LoadedClassCount)
	}

	if cfg.HeadRoom != "10" {
		t.Errorf("Expected HeadRoom 10, got %s", cfg.HeadRoom)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorCode   errors.ErrorCode
	}{
		{
			name: "Valid config",
			config: &Config{
				ThreadCount:      "250",
				LoadedClassCount: "35000",
				HeadRoom:         "0",
			},
			expectError: false,
		},
		{
			name: "Invalid thread count - negative",
			config: &Config{
				ThreadCount:      "-1",
				LoadedClassCount: "35000",
				HeadRoom:         "0",
			},
			expectError: true,
			errorCode:   errors.ErrInvalidConfiguration,
		},
		{
			name: "Invalid thread count - not a number",
			config: &Config{
				ThreadCount:      "abc",
				LoadedClassCount: "35000",
				HeadRoom:         "0",
			},
			expectError: true,
			errorCode:   errors.ErrInvalidConfiguration,
		},
		{
			name: "Invalid loaded class count - negative",
			config: &Config{
				ThreadCount:      "250",
				LoadedClassCount: "-1",
				HeadRoom:         "0",
			},
			expectError: true,
			errorCode:   errors.ErrInvalidConfiguration,
		},
		{
			name: "Invalid head room - negative",
			config: &Config{
				ThreadCount:      "250",
				LoadedClassCount: "35000",
				HeadRoom:         "-1",
			},
			expectError: true,
			errorCode:   errors.ErrInvalidConfiguration,
		},
		{
			name: "Invalid head room - over 100",
			config: &Config{
				ThreadCount:      "250",
				LoadedClassCount: "35000",
				HeadRoom:         "101",
			},
			expectError: true,
			errorCode:   errors.ErrInvalidConfiguration,
		},
		{
			name: "Valid head room - boundary values",
			config: &Config{
				ThreadCount:      "1",
				LoadedClassCount: "1",
				HeadRoom:         "100",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

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
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestConfigSetEnvironmentVariables(t *testing.T) {
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

func TestConfigSetTotalMemory(t *testing.T) {
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

func TestGetEnvOrDefault(t *testing.T) {
	key := "TEST_ENV_VAR"
	defaultValue := "default"

	// Test when env var is not set
	result := getEnvOrDefault(key, defaultValue)
	if result != defaultValue {
		t.Errorf("Expected %s, got %s", defaultValue, result)
	}

	// Test when env var is set
	os.Setenv(key, "custom_value")
	defer os.Unsetenv(key)

	result = getEnvOrDefault(key, defaultValue)
	if result != "custom_value" {
		t.Errorf("Expected custom_value, got %s", result)
	}

	// Test when env var is set to empty string
	os.Setenv(key, "")
	result = getEnvOrDefault(key, defaultValue)
	if result != defaultValue {
		t.Errorf("Expected %s for empty env var, got %s", defaultValue, result)
	}
}
