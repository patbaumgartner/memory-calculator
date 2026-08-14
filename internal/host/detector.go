// Package host reads how much memory the host system has available.
package host

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const (
	// LinuxMemInfoPath is the path to /proc/meminfo on Linux systems.
	LinuxMemInfoPath = "/proc/meminfo"

	platformLinux = "linux"

	availableKey = "MemAvailable:"
	totalKey     = "MemTotal:"
)

// Detector reads host memory information.
type Detector struct {
	// MemInfoPath is the path to memory information (Linux only).
	MemInfoPath string
}

// Create creates a host memory detector using the standard Linux path.
func Create() *Detector {
	return &Detector{MemInfoPath: LinuxMemInfoPath}
}

// CreateWithPath creates a host memory detector reading from a specific path.
func CreateWithPath(memInfoPath string) *Detector {
	return &Detector{MemInfoPath: memInfoPath}
}

// DetectAvailableMemory reports memory available for a new workload, or 0 when it cannot be
// determined.
//
// This reads MemAvailable rather than MemTotal. Without a cgroup limit the process may share the
// host with everything else running on it, and sizing a heap against total RAM would overcommit a
// busy machine. MemAvailable is the kernel's own estimate of what can be allocated without
// swapping, which is the conservative choice and matches the Paketo calculator this tool mirrors.
//
// Only Linux is supported. Other platforms return 0 so the caller can apply its documented
// default rather than act on a guess.
func (d *Detector) DetectAvailableMemory() int64 {
	if runtime.GOOS != platformLinux {
		return 0
	}

	if available := d.readMemInfo(availableKey); available > 0 {
		return available
	}

	// Kernels before 3.14 do not report MemAvailable.
	return d.readMemInfo(totalKey)
}

// readMemInfo returns the named /proc/meminfo value in bytes.
func (d *Detector) readMemInfo(key string) int64 {
	file, err := os.Open(d.MemInfoPath) // #nosec G304 - configurable only for tests
	if err != nil {
		return 0
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, key) {
			continue
		}

		// Format: "MemAvailable:    8062332 kB"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0
		}

		value, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil || value < 0 {
			return 0
		}

		return value * unitMultiplier(fields)
	}

	return 0
}

func unitMultiplier(fields []string) int64 {
	if len(fields) < 3 {
		return 1
	}

	switch strings.ToLower(fields[2]) {
	case "kb":
		return 1024
	case "mb":
		return 1024 * 1024
	case "gb":
		return 1024 * 1024 * 1024
	default:
		return 1
	}
}

// IsHostMemoryDetectionSupported reports whether host memory can be read on this platform.
func IsHostMemoryDetectionSupported() bool {
	return runtime.GOOS == platformLinux
}
