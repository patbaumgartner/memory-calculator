// Package constants provides application-wide constants for the memory calculator.
package constants

const (
	// Application metadata

	// ApplicationName is the name of the application.
	ApplicationName = "JVM Memory Calculator"
	// DefaultVersion is the default version string.
	DefaultVersion = "dev"
	// UnknownValue indicates an unknown value.
	UnknownValue = "unknown"

	// Default values

	// DefaultThreadCount is the default thread count.
	DefaultThreadCount = "250"
	// DefaultHeadRoom is the default headroom configuration string.
	DefaultHeadRoom = "0"
	// DefaultApplicationPath is the default application path.
	DefaultApplicationPath = "/app"

	// Environment variable names

	// EnvTotalMemory is the environment variable for total memory.
	EnvTotalMemory = "BPL_JVM_TOTAL_MEMORY"
	// EnvThreadCount is the environment variable for thread count.
	EnvThreadCount = "BPL_JVM_THREAD_COUNT"
	// EnvLoadedClassCount is the environment variable for loaded class count.
	EnvLoadedClassCount = "BPL_JVM_LOADED_CLASS_COUNT"
	// EnvHeadRoom is the environment variable for head room.
	EnvHeadRoom = "BPL_JVM_HEAD_ROOM"
	// EnvApplicationPath is the environment variable for application path.
	EnvApplicationPath = "BPI_APPLICATION_PATH"
	// EnvJVMClassCount is the environment variable for JVM class count.
	EnvJVMClassCount = "BPI_JVM_CLASS_COUNT"
	// EnvQuiet is the environment variable for quiet mode.
	EnvQuiet = "QUIET"

	// System paths

	// DefaultMemoryLimitPathV1 is the cgroup v1 memory limit path.
	DefaultMemoryLimitPathV1 = "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	// DefaultMemoryLimitPathV2 is the path to the cgroup v2 memory limit file.
	DefaultMemoryLimitPathV2 = "/sys/fs/cgroup/memory.max"
	// DefaultMemoryInfoPath is the path to /proc/meminfo.
	DefaultMemoryInfoPath = "/proc/meminfo"

	// Memory limits and validation

	// MaxJVMSizeGB is the maximum JVM size in GB.
	MaxJVMSizeGB = 64
	// MaxRealisticMemoryBytes is the maximum realistic memory in bytes.
	MaxRealisticMemoryBytes = 1 << 50 // 1 PB - unrealistically large for containers
	// MinValidMemoryBytes is the minimum valid memory in bytes.
	MinValidMemoryBytes = 1024 // 1KB minimum

	// Output formatting

	// HelpSeparator is the separator line for help output.
	HelpSeparator = "===================="
	// IndentationSpaces is the indentation string for help output.
	IndentationSpaces = "  "
)
