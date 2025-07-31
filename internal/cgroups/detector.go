// Package cgroups handles container memory detection from cgroups v1 and v2,
// with host system memory detection as fallback.
package cgroups

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/patbaumgartner/memory-calculator/internal/host"
	"github.com/patbaumgartner/memory-calculator/pkg/errors"
)

const (
	// Maximum realistic memory limit (1TB) to filter out "no limit" values
	MaxRealisticMemory = 1024 * 1024 * 1024 * 1024
)

// Detector handles container memory detection from cgroups with host system fallback.
type Detector struct {
	// CgroupsV2Path is the path to cgroups v2 memory.max file
	CgroupsV2Path string
	// CgroupsV1Path is the path to cgroups v1 memory.limit_in_bytes file
	CgroupsV1Path string
	// HostDetector handles host system memory detection as fallback
	HostDetector *host.Detector
}

// NewDetector creates a new cgroups detector with default paths and host fallback.
func NewDetector() *Detector {
	return &Detector{
		CgroupsV2Path: "/sys/fs/cgroup/memory.max",
		CgroupsV1Path: "/sys/fs/cgroup/memory/memory.limit_in_bytes",
		HostDetector:  host.NewDetector(),
	}
}

// NewDetectorWithPaths creates a new cgroups detector with custom paths (useful for testing).
func NewDetectorWithPaths(v2Path, v1Path string) *Detector {
	return &Detector{
		CgroupsV2Path: v2Path,
		CgroupsV1Path: v1Path,
		HostDetector:  host.NewDetector(),
	}
}

// NewDetectorWithPathsAndHost creates a new cgroups detector with custom paths and host detector (useful for testing).
func NewDetectorWithPathsAndHost(v2Path, v1Path string, hostDetector *host.Detector) *Detector {
	return &Detector{
		CgroupsV2Path: v2Path,
		CgroupsV1Path: v1Path,
		HostDetector:  hostDetector,
	}
}

// DetectContainerMemory attempts to read memory limit from cgroups v2 first, then v1,
// and falls back to host system memory detection if cgroups are not available.
// Returns 0 if no memory limit is detected or if an error occurs.
func (d *Detector) DetectContainerMemory() int64 {
	// Try cgroups v2 first
	if memory, err := d.readCgroupsV2(); err == nil && memory > 0 {
		return memory
	}

	// Fall back to cgroups v1
	if memory, err := d.readCgroupsV1(); err == nil && memory > 0 {
		return memory
	}

	// Fall back to host system memory detection
	if d.HostDetector != nil {
		if hostMemory := d.HostDetector.DetectHostMemory(); hostMemory > 0 {
			return hostMemory
		}
	}

	return 0
}

// readCgroupsV2 reads memory limit from cgroups v2.
func (d *Detector) readCgroupsV2() (int64, error) {
	file, err := os.Open(d.CgroupsV2Path)
	if err != nil {
		return 0, errors.NewCgroupsError(d.CgroupsV2Path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0, errors.NewCgroupsError(d.CgroupsV2Path, scanner.Err())
	}

	line := strings.TrimSpace(scanner.Text())
	if line == "max" {
		return 0, nil // No limit set
	}

	memory, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return 0, errors.NewCgroupsError(d.CgroupsV2Path, err)
	}

	if memory > MaxRealisticMemory {
		return 0, nil // Unrealistic limit, treat as no limit
	}

	return memory, nil
}

// readCgroupsV1 reads memory limit from cgroups v1.
func (d *Detector) readCgroupsV1() (int64, error) {
	file, err := os.Open(d.CgroupsV1Path)
	if err != nil {
		return 0, errors.NewCgroupsError(d.CgroupsV1Path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0, errors.NewCgroupsError(d.CgroupsV1Path, scanner.Err())
	}

	line := strings.TrimSpace(scanner.Text())
	memory, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return 0, errors.NewCgroupsError(d.CgroupsV1Path, err)
	}

	// Check if it's a realistic limit (not the "no limit" value)
	if memory > MaxRealisticMemory {
		return 0, nil // Unrealistic limit, treat as no limit
	}

	return memory, nil
}
