package config

import (
	"os"
	"strings"
	"testing"
)

const customAppPath = "/custom/app"

func TestLoad(t *testing.T) {
	// Clear any existing environment variables
	_ = os.Unsetenv("BPL_JVM_TOTAL_MEMORY")
	_ = os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")
	_ = os.Unsetenv("BPL_JVM_THREAD_COUNT")
	_ = os.Unsetenv("BPL_JVM_HEAD_ROOM")
	_ = os.Unsetenv("BPI_APPLICATION_PATH")

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

	if cfg.Path != "/app" {
		t.Errorf("Expected path '/app', got '%s'", cfg.Path)
	}

	if cfg.BuildVersion != "dev" {
		t.Errorf("Expected build version 'dev', got '%s'", cfg.BuildVersion)
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("BPL_JVM_LOADED_CLASS_COUNT", "15000")
	_ = os.Setenv("BPL_JVM_THREAD_COUNT", "500")
	_ = os.Setenv("BPL_JVM_HEAD_ROOM", "10")
	_ = os.Setenv("BPI_APPLICATION_PATH", customAppPath)

	defer func() {
		_ = os.Unsetenv("BPL_JVM_LOADED_CLASS_COUNT")
		_ = os.Unsetenv("BPL_JVM_THREAD_COUNT")
		_ = os.Unsetenv("BPL_JVM_HEAD_ROOM")
		_ = os.Unsetenv("BPI_APPLICATION_PATH")
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

	if cfg.Path != customAppPath {
		t.Errorf("Expected path '/custom/app', got '%s'", cfg.Path)
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
				Path:             "/app",
			},
			expectError: false,
		},
		{
			name: "Valid config with values",
			config: &Config{
				ThreadCount:      "300",
				LoadedClassCount: "5000",
				HeadRoom:         "5",
				Path:             customAppPath,
			},
			expectError: false,
		},
		{
			name: "Invalid thread count - negative",
			config: &Config{
				ThreadCount:      "-1",
				LoadedClassCount: "1000",
				HeadRoom:         "0",
				Path:             "/app",
			},
			expectError: true,
		},
		{
			name: "Invalid thread count - not a number",
			config: &Config{
				ThreadCount:      "abc",
				LoadedClassCount: "1000",
				HeadRoom:         "0",
				Path:             "/app",
			},
			expectError: true,
		},
		{
			name: "Invalid loaded class count - negative",
			config: &Config{
				ThreadCount:      "250",
				LoadedClassCount: "-1",
				HeadRoom:         "0",
				Path:             "/app",
			},
			expectError: true,
		},
		{
			name: "Invalid head room - negative",
			config: &Config{
				ThreadCount:      "250",
				LoadedClassCount: "1000",
				HeadRoom:         "-1",
				Path:             "/app",
			},
			expectError: true,
		},
		{
			name: "Invalid head room - over 100",
			config: &Config{
				ThreadCount:      "250",
				LoadedClassCount: "1000",
				HeadRoom:         "101",
				Path:             "/app",
			},
			expectError: true,
		},
		{
			name: "Invalid path - empty",
			config: &Config{
				ThreadCount:      "250",
				LoadedClassCount: "1000",
				HeadRoom:         "0",
				Path:             "",
			},
			expectError: true,
		},
		{
			name: "Valid total memory with unit",
			config: &Config{
				TotalMemory: "2G", ThreadCount: "250", HeadRoom: "0", Path: "/app",
			},
			expectError: false,
		},
		{
			name: "Valid total memory as decimal",
			config: &Config{
				TotalMemory: "1.5GB", ThreadCount: "250", HeadRoom: "0", Path: "/app",
			},
			expectError: false,
		},
		{
			name: "Valid total memory as plain bytes",
			config: &Config{
				TotalMemory: "1073741824", ThreadCount: "250", HeadRoom: "0", Path: "/app",
			},
			expectError: false,
		},
		{
			name: "Invalid total memory - not a size",
			config: &Config{
				TotalMemory: "banana", ThreadCount: "250", HeadRoom: "0", Path: "/app",
			},
			expectError: true,
		},
		{
			name: "Invalid total memory - unknown unit",
			config: &Config{
				TotalMemory: "2X", ThreadCount: "250", HeadRoom: "0", Path: "/app",
			},
			expectError: true,
		},
		{
			name: "Invalid total memory - negative",
			config: &Config{
				TotalMemory: "-2G", ThreadCount: "250", HeadRoom: "0", Path: "/app",
			},
			expectError: true,
		},
		{
			name: "Invalid total memory - zero",
			config: &Config{
				TotalMemory: "0", ThreadCount: "250", HeadRoom: "0", Path: "/app",
			},
			expectError: true,
		},
		{
			name: "Valid head room at the 99 percent bound",
			config: &Config{
				ThreadCount: "250", HeadRoom: "99", Path: "/app",
			},
			expectError: false,
		},
		{
			name: "Invalid head room - 100 percent leaves nothing for the JVM",
			config: &Config{
				ThreadCount: "250", HeadRoom: "100", Path: "/app",
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

func TestInputConvertsValidatedConfig(t *testing.T) {
	cfg := &Config{
		TotalMemory:      "2G",
		ThreadCount:      "300",
		LoadedClassCount: "40000",
		HeadRoom:         "15",
		Path:             customAppPath,
		JVMClassCount:    "1500",
		AdjustmentFactor: "125",
		StaticAdjustment: "-50",
		JavaToolOptions:  "-Xss2M",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	input := cfg.Input()
	if input.TotalMemory == nil || *input.TotalMemory != 2*1024*1024*1024 {
		t.Errorf("TotalMemory = %v, want 2G", input.TotalMemory)
	}
	if input.LoadedClassCount == nil || *input.LoadedClassCount != 40000 {
		t.Errorf("LoadedClassCount = %v, want 40000", input.LoadedClassCount)
	}
	if input.ThreadCount != 300 || input.HeadRoom != 15 || input.ApplicationPath != customAppPath {
		t.Errorf("Input = %+v, want converted core values", input)
	}
	if input.JVMClassCount != 1500 || input.AdjustmentFactor != 125 || input.StaticAdjustment != -50 {
		t.Errorf("Input = %+v, want converted class adjustments", input)
	}
	if input.JavaToolOptions != "-Xss2M" {
		t.Errorf("JavaToolOptions = %q, want -Xss2M", input.JavaToolOptions)
	}
}

func TestLoadDeprecatedHeadRoomPrecedence(t *testing.T) {
	t.Setenv("BPL_JVM_HEADROOM", "10")
	t.Setenv("BPL_JVM_HEAD_ROOM", "")
	if got := Load().HeadRoom; got != "10" {
		t.Errorf("deprecated-only HeadRoom = %q, want 10", got)
	}
	if warnings := Load().Warnings; len(warnings) != 1 || !strings.Contains(warnings[0], "deprecated") {
		t.Errorf("deprecated-only Warnings = %v, want one deprecation warning", warnings)
	}

	t.Setenv("BPL_JVM_HEAD_ROOM", "20")
	if got := Load().HeadRoom; got != "20" {
		t.Errorf("new HeadRoom = %q, want 20 to override the deprecated value", got)
	}
	if warnings := Load().Warnings; len(warnings) != 1 || !strings.Contains(warnings[0], "ignored") {
		t.Errorf("both-key Warnings = %v, want one ignored-value warning", warnings)
	}
}

func TestLoadReadsEveryExternalInput(t *testing.T) {
	for key, value := range map[string]string{
		"BPL_JVM_TOTAL_MEMORY":        "2G",
		"BPL_JVM_THREAD_COUNT":        "300",
		"BPL_JVM_LOADED_CLASS_COUNT":  "40000",
		"BPL_JVM_HEAD_ROOM":           "15",
		"BPI_APPLICATION_PATH":        customAppPath,
		"BPI_JVM_CLASS_COUNT":         "1500",
		"BPI_CLASS_ADJUSTMENT_FACTOR": "125",
		"BPI_CLASS_STATIC_ADJUSTMENT": "-50",
		"JAVA_TOOL_OPTIONS":           "-Xss2M",
	} {
		t.Setenv(key, value)
	}

	got := Load()
	if got.TotalMemory != "2G" || got.ThreadCount != "300" || got.LoadedClassCount != "40000" ||
		got.HeadRoom != "15" || got.Path != customAppPath || got.JVMClassCount != "1500" ||
		got.AdjustmentFactor != "125" || got.StaticAdjustment != "-50" || got.JavaToolOptions != "-Xss2M" {
		t.Errorf("Load() = %+v, want all external inputs", got)
	}
}
