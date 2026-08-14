// Package config loads and validates the memory calculator's external configuration.
package config

import (
	"os"
	"strconv"

	"github.com/patbaumgartner/memory-calculator/internal/calc"
	"github.com/patbaumgartner/memory-calculator/internal/calculator"
	"github.com/patbaumgartner/memory-calculator/internal/memory"
	"github.com/patbaumgartner/memory-calculator/pkg/errors"
)

const (
	defaultThreadCount      = "250"
	defaultHeadRoom         = "0"
	defaultApplicationPath  = "/app"
	defaultJVMClassCount    = "1000"
	defaultAdjustment       = "100"
	defaultStaticAdjustment = "0"
)

// Config holds the CLI and environment representation of the calculator configuration.
type Config struct {
	TotalMemory      string
	ThreadCount      string
	LoadedClassCount string
	HeadRoom         string
	Path             string
	JVMClassCount    string
	AdjustmentFactor string
	StaticAdjustment string
	JavaToolOptions  string

	Quiet   bool
	Version bool
	Help    bool

	BuildVersion string
	BuildTime    string
	CommitHash   string
}

// Load reads every supported environment variable once, applying defaults and compatibility
// precedence at the process boundary.
func Load() *Config {
	headRoom := os.Getenv("BPL_JVM_HEAD_ROOM")
	if headRoom == "" {
		headRoom = os.Getenv("BPL_JVM_HEADROOM")
	}
	if headRoom == "" {
		headRoom = defaultHeadRoom
	}

	return &Config{
		TotalMemory:      os.Getenv("BPL_JVM_TOTAL_MEMORY"),
		ThreadCount:      getEnvOrDefault("BPL_JVM_THREAD_COUNT", defaultThreadCount),
		LoadedClassCount: os.Getenv("BPL_JVM_LOADED_CLASS_COUNT"),
		HeadRoom:         headRoom,
		Path:             getEnvOrDefault("BPI_APPLICATION_PATH", defaultApplicationPath),
		JVMClassCount:    getEnvOrDefault("BPI_JVM_CLASS_COUNT", defaultJVMClassCount),
		AdjustmentFactor: getEnvOrDefault("BPI_CLASS_ADJUSTMENT_FACTOR", defaultAdjustment),
		StaticAdjustment: getEnvOrDefault("BPI_CLASS_STATIC_ADJUSTMENT", defaultStaticAdjustment),
		JavaToolOptions:  os.Getenv("JAVA_TOOL_OPTIONS"),
		BuildVersion:     "dev",
		BuildTime:        "unknown",
		CommitHash:       "unknown",
	}
}

// Validate checks every external value before it reaches the calculation domain.
func (c *Config) Validate() error {
	if c.TotalMemory != "" {
		totalMemory, err := memory.CreateParser().ParseMemoryString(c.TotalMemory)
		if err != nil {
			return errors.NewConfigurationError(
				"total-memory", c.TotalMemory, "must be a memory size such as 2G, 512M, or 1073741824")
		}
		if totalMemory <= 0 {
			return errors.NewConfigurationError("total-memory", c.TotalMemory, "must be greater than zero")
		}
	}

	if _, err := positiveInteger("thread-count", c.ThreadCount); err != nil {
		return err
	}

	if c.LoadedClassCount != "" {
		if _, err := positiveInteger("loaded-class-count", c.LoadedClassCount); err != nil {
			return err
		}
	}

	headRoom, err := strconv.Atoi(c.HeadRoom)
	if err != nil || headRoom < 0 || headRoom > calc.MaxHeadRoom {
		return errors.NewConfigurationError("head-room", c.HeadRoom, "must be an integer between 0 and 99")
	}

	if c.Path == "" {
		return errors.NewConfigurationError("path", c.Path, "application path cannot be empty")
	}

	if _, err := nonNegativeInteger("BPI_JVM_CLASS_COUNT", valueOrDefault(c.JVMClassCount, defaultJVMClassCount)); err != nil {
		return err
	}
	if _, err := nonNegativeInteger("BPI_CLASS_ADJUSTMENT_FACTOR",
		valueOrDefault(c.AdjustmentFactor, defaultAdjustment)); err != nil {
		return err
	}
	if _, err := integer("BPI_CLASS_STATIC_ADJUSTMENT",
		valueOrDefault(c.StaticAdjustment, defaultStaticAdjustment)); err != nil {
		return err
	}

	return nil
}

// Input converts validated external configuration into the typed calculator contract.
func (c *Config) Input() calculator.Input {
	threadCount, _ := strconv.Atoi(c.ThreadCount)
	headRoom, _ := strconv.Atoi(c.HeadRoom)
	jvmClassCount, _ := strconv.Atoi(valueOrDefault(c.JVMClassCount, defaultJVMClassCount))
	adjustmentFactor, _ := strconv.Atoi(valueOrDefault(c.AdjustmentFactor, defaultAdjustment))
	staticAdjustment, _ := strconv.Atoi(valueOrDefault(c.StaticAdjustment, defaultStaticAdjustment))

	input := calculator.Input{
		ThreadCount:      threadCount,
		HeadRoom:         headRoom,
		ApplicationPath:  c.Path,
		JavaToolOptions:  c.JavaToolOptions,
		JVMClassCount:    jvmClassCount,
		AdjustmentFactor: adjustmentFactor,
		StaticAdjustment: staticAdjustment,
	}

	if c.TotalMemory != "" {
		value, _ := memory.CreateParser().ParseMemoryString(c.TotalMemory)
		input.TotalMemory = &value
	}

	if c.LoadedClassCount != "" {
		value, _ := strconv.Atoi(c.LoadedClassCount)
		input.LoadedClassCount = &value
	}

	return input
}

func positiveInteger(name, value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, errors.NewConfigurationError(name, value, "must be a positive integer")
	}
	return parsed, nil
}

func nonNegativeInteger(name, value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, errors.NewConfigurationError(name, value, "must be a non-negative integer")
	}
	return parsed, nil
}

func integer(name, value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.NewConfigurationError(name, value, "must be an integer")
	}
	return parsed, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	return valueOrDefault(os.Getenv(key), defaultValue)
}

func valueOrDefault(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}
