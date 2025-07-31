// Package config handles configuration management for the memory calculator.
package config

import (
	"os"
	"strconv"

	"github.com/patbaumgartner/memory-calculator/pkg/errors"
)

// Config holds all configuration parameters for the memory calculator.
type Config struct {
	// Memory configuration
	TotalMemory      string
	ThreadCount      string
	LoadedClassCount string
	HeadRoom         string

	// Output configuration
	Quiet   bool
	Version bool
	Help    bool

	// Build information
	BuildVersion string
	BuildTime    string
	CommitHash   string
}

// DefaultConfig returns a configuration with default values.
func DefaultConfig() *Config {
	return &Config{
		ThreadCount:      getEnvOrDefault("BPL_JVM_THREAD_COUNT", "250"),
		LoadedClassCount: getEnvOrDefault("BPL_JVM_LOADED_CLASS_COUNT", "35000"),
		HeadRoom:         getEnvOrDefault("BPL_JVM_HEAD_ROOM", "0"),
		BuildVersion:     "dev",
		BuildTime:        "unknown",
		CommitHash:       "unknown",
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// Validate thread count
	if threadCount, err := strconv.Atoi(c.ThreadCount); err != nil || threadCount < 1 {
		return errors.NewConfigurationError("thread-count", c.ThreadCount, "must be a positive integer")
	}

	// Validate loaded class count
	if classCount, err := strconv.Atoi(c.LoadedClassCount); err != nil || classCount < 1 {
		return errors.NewConfigurationError("loaded-class-count", c.LoadedClassCount, "must be a positive integer")
	}

	// Validate head room
	if headRoom, err := strconv.Atoi(c.HeadRoom); err != nil || headRoom < 0 || headRoom > 100 {
		return errors.NewConfigurationError("head-room", c.HeadRoom, "must be an integer between 0 and 100")
	}

	return nil
}

// SetEnvironmentVariables sets buildpack environment variables from the config.
func (c *Config) SetEnvironmentVariables() {
	_ = os.Setenv("BPL_JVM_THREAD_COUNT", c.ThreadCount)
	_ = os.Setenv("BPL_JVM_LOADED_CLASS_COUNT", c.LoadedClassCount)
	_ = os.Setenv("BPL_JVM_HEAD_ROOM", c.HeadRoom)
}

// SetTotalMemory sets the total memory environment variable if memory is specified.
func (c *Config) SetTotalMemory(totalMemory int64) {
	if totalMemory > 0 {
		_ = os.Setenv("BPL_JVM_TOTAL_MEMORY", strconv.FormatInt(totalMemory, 10))
	}
}

// getEnvOrDefault returns the environment variable value or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
