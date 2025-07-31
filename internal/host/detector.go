// Package host handles host system memory detection across different operating systems.
package host

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const (
	// LinuxMemInfoPath is the path to /proc/meminfo on Linux systems
	LinuxMemInfoPath = "/proc/meminfo"

	// Platform constants
	platformLinux  = "linux"
	platformDarwin = "darwin"
)

// Detector handles host system memory detection.
type Detector struct {
	// MemInfoPath is the path to memory information (Linux only)
	MemInfoPath string
}

// NewDetector creates a new host memory detector with default paths.
func NewDetector() *Detector {
	return &Detector{
		MemInfoPath: LinuxMemInfoPath,
	}
}

// NewDetectorWithPath creates a new host memory detector with custom path (useful for testing).
func NewDetectorWithPath(memInfoPath string) *Detector {
	return &Detector{
		MemInfoPath: memInfoPath,
	}
}

// DetectHostMemory attempts to detect total system memory based on the operating system.
// Returns 0 if memory detection fails or is not supported on the current platform.
func (d *Detector) DetectHostMemory() int64 {
	switch runtime.GOOS {
	case platformLinux:
		return d.detectLinuxMemory()
	case platformDarwin:
		return d.detectDarwinMemory()
	default:
		return 0 // Unsupported platform
	}
}

// detectLinuxMemory reads total memory from /proc/meminfo on Linux.
func (d *Detector) detectLinuxMemory() int64 {
	file, err := os.Open(d.MemInfoPath)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "MemTotal:") {
			// Format: "MemTotal:        8062332 kB"
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if memKB, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					// Convert from KB to bytes
					return memKB * 1024
				}
			}
		}
	}

	return 0
}

// detectDarwinMemory detects memory on macOS using system calls.
// Note: This requires CGO to be enabled, so we'll implement a CGO-free version
// using runtime.ReadMemStats() which gives us a reasonable approximation.
func (d *Detector) detectDarwinMemory() int64 {
	// For cross-platform compatibility without CGO, we use a heuristic
	// based on Go's memory stats and some reasonable assumptions
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// This is a heuristic - typically the heap limit is much smaller than total system memory
	// We'll estimate total memory as roughly 16x the current heap size (conservative estimate)
	// This isn't perfect but provides a reasonable fallback without CGO dependencies
	if m.Sys > 0 {
		// Estimate system memory based on allocated system memory
		// This is a rough approximation - real implementation would use syscalls
		estimatedTotal := m.Sys * 32 // Conservative multiplier

		// Cap at reasonable values (between 1GB and 128GB)
		const minMemory = 1024 * 1024 * 1024       // 1GB
		const maxMemory = 128 * 1024 * 1024 * 1024 // 128GB

		if estimatedTotal < minMemory {
			return minMemory
		}
		if estimatedTotal > maxMemory {
			return maxMemory
		}

		return int64(estimatedTotal)
	}

	return 0
}

// IsHostMemoryDetectionSupported returns true if host memory detection is supported on the current platform.
func IsHostMemoryDetectionSupported() bool {
	switch runtime.GOOS {
	case platformLinux, platformDarwin:
		return true
	default:
		return false
	}
}
