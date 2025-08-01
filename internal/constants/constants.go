// Package constants provides application-wide constants for the memory calculator.
package constants

const (
	// Application metadata
	ApplicationName = "JVM Memory Calculator"
	DefaultVersion  = "dev"
	UnknownValue    = "unknown"

	// Default values
	DefaultThreadCount     = "250"
	DefaultHeadRoom        = "0"
	DefaultApplicationPath = "/app"

	// Environment variable names
	EnvTotalMemory      = "BPL_JVM_TOTAL_MEMORY"
	EnvThreadCount      = "BPL_JVM_THREAD_COUNT"
	EnvLoadedClassCount = "BPL_JVM_LOADED_CLASS_COUNT"
	EnvHeadRoom         = "BPL_JVM_HEAD_ROOM"
	EnvApplicationPath  = "BPI_APPLICATION_PATH"
	EnvJVMClassCount    = "BPI_JVM_CLASS_COUNT"
	EnvQuiet            = "QUIET"

	// System paths
	DefaultMemoryLimitPathV1 = "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	DefaultMemoryLimitPathV2 = "/sys/fs/cgroup/memory.max"
	DefaultMemoryInfoPath    = "/proc/meminfo"

	// Memory limits and validation
	MaxJVMSizeGB            = 64
	MaxRealisticMemoryBytes = 1 << 50 // 1 PB - unrealistically large for containers
	MinValidMemoryBytes     = 1024    // 1KB minimum

	// Output formatting
	HelpSeparator     = "===================="
	IndentationSpaces = "  "
)
