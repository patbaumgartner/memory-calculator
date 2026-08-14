// Package cgroups determines how much memory the process may actually use, preferring the
// container's cgroup limit and falling back to host memory.
package cgroups

import (
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/patbaumgartner/memory-calculator/internal/host"
)

const (
	// DefaultV2Root is the unified cgroup hierarchy mount point.
	DefaultV2Root = "/sys/fs/cgroup"
	// DefaultV1Root is the cgroup v1 memory controller mount point.
	DefaultV1Root = "/sys/fs/cgroup/memory"
	// DefaultProcCgroup lists the cgroups the current process belongs to.
	DefaultProcCgroup = "/proc/self/cgroup"

	v2LimitFile = "memory.max"
	v1LimitFile = "memory.limit_in_bytes"

	// unlimitedThreshold treats implausibly large limits as "no limit".
	//
	// The kernel reports an unset v1 limit as its page counter maximum, which is int64 max rounded
	// down to a page boundary. That value depends on the page size (4 KiB on x86-64, up to 64 KiB
	// on arm64), so matching a single literal such as 9223372036854771712 misses real systems.
	// No container is limited to more than 4 EiB, so anything at or above that is not a limit.
	unlimitedThreshold = int64(1) << 62
)

// Source identifies where a memory limit was found.
type Source string

const (
	// SourceCgroupV2 indicates the limit came from the unified cgroup hierarchy.
	SourceCgroupV2 Source = "cgroups v2"
	// SourceCgroupV1 indicates the limit came from the cgroup v1 memory controller.
	SourceCgroupV1 Source = "cgroups v1"
	// SourceHost indicates no cgroup limit applied and host memory was used.
	SourceHost Source = "host"
	// SourceNone indicates no memory limit could be determined.
	SourceNone Source = "none"
)

// Detection reports a memory limit together with where it came from.
type Detection struct {
	// Limit is the memory limit in bytes, or 0 when none could be determined.
	Limit int64
	// Source records which mechanism supplied the limit.
	Source Source
}

// Found reports whether a usable limit was determined.
func (d Detection) Found() bool { return d.Limit > 0 && d.Source != SourceNone }

// Detector locates the memory limit that applies to the current process.
type Detector struct {
	// V2Root is the unified cgroup hierarchy mount point.
	V2Root string
	// V1Root is the cgroup v1 memory controller mount point.
	V1Root string
	// ProcCgroup is the file listing this process's cgroup membership.
	ProcCgroup string
	// HostDetector supplies host memory when no cgroup limit applies.
	HostDetector *host.Detector
}

// Create creates a detector using the standard Linux paths.
func Create() *Detector {
	return &Detector{
		V2Root:       DefaultV2Root,
		V1Root:       DefaultV1Root,
		ProcCgroup:   DefaultProcCgroup,
		HostDetector: host.Create(),
	}
}

// Detect returns the memory limit that applies to this process.
//
// cgroup v2 is consulted before v1: on a hybrid system both hierarchies are mounted, but the
// unified hierarchy is the one the container runtime configures.
func (d *Detector) Detect() Detection {
	if limit, ok := d.cgroupLimit(d.V2Root, v2LimitFile, parseV2Limit); ok {
		return Detection{Limit: limit, Source: SourceCgroupV2}
	}

	if limit, ok := d.cgroupLimit(d.V1Root, v1LimitFile, parseLimit); ok {
		return Detection{Limit: limit, Source: SourceCgroupV1}
	}

	if d.HostDetector != nil {
		if limit := d.HostDetector.DetectAvailableMemory(); limit > 0 {
			return Detection{Limit: limit, Source: SourceHost}
		}
	}

	return Detection{Source: SourceNone}
}

// cgroupLimit returns the tightest limit applying to this process.
//
// The limit on the process's own cgroup is not sufficient: a leaf may be unlimited while an
// ancestor caps the whole subtree, which is how Kubernetes constrains a pod. Every level from the
// process's cgroup up to the mount root is checked and the smallest finite limit wins.
func (d *Detector) cgroupLimit(root, file string, parse func(string) (int64, bool)) (int64, bool) {
	if root == "" {
		return 0, false
	}

	limit, found := int64(0), false

	for _, dir := range d.hierarchy(root, file) {
		contents, err := os.ReadFile(path.Join(dir, file)) // #nosec G304 - assembled from cgroup mount paths
		if err != nil {
			continue
		}

		value, ok := parse(strings.TrimSpace(string(contents)))
		if !ok {
			continue
		}

		if !found || value < limit {
			limit, found = value, true
		}
	}

	return limit, found
}

// hierarchy lists the cgroup directories to inspect, from the mount root down to the process's own
// cgroup. When membership cannot be resolved only the root is inspected, which is the correct
// answer inside a cgroup namespace because the container's cgroup is mounted as the root.
func (d *Detector) hierarchy(root, file string) []string {
	dirs := []string{root}

	relative := d.relativeCgroupPath(file)
	for relative != "" && relative != "/" && relative != "." {
		dirs = append(dirs, path.Join(root, relative))
		relative = path.Dir(relative)
	}

	return dirs
}

// relativeCgroupPath reads this process's cgroup path for the relevant hierarchy.
func (d *Detector) relativeCgroupPath(file string) string {
	if d.ProcCgroup == "" {
		return ""
	}

	contents, err := os.ReadFile(d.ProcCgroup) // #nosec G304 - configurable only for tests
	if err != nil {
		return ""
	}

	wantUnified := file == v2LimitFile

	for _, line := range strings.Split(string(contents), "\n") {
		// Format: hierarchy-ID:controller-list:cgroup-path
		fields := strings.SplitN(strings.TrimSpace(line), ":", 3)
		if len(fields) != 3 {
			continue
		}

		unified := fields[0] == "0" && fields[1] == ""
		if unified != wantUnified {
			continue
		}

		if unified || hasMemoryController(fields[1]) {
			return fields[2]
		}
	}

	return ""
}

func hasMemoryController(controllers string) bool {
	for _, controller := range strings.Split(controllers, ",") {
		if controller == "memory" {
			return true
		}
	}

	return false
}

// parseV2Limit reads a cgroup v2 memory.max value, where "max" means no limit.
func parseV2Limit(contents string) (int64, bool) {
	if contents == "max" {
		return 0, false
	}

	return parseLimit(contents)
}

// parseLimit reads a byte count, rejecting values that mean "no limit". An unset cgroup v1 limit is
// reported as a very large number rather than a sentinel word, and may exceed int64 entirely.
func parseLimit(contents string) (int64, bool) {
	value, err := strconv.ParseInt(contents, 10, 64)
	if err != nil {
		return 0, false
	}

	if value <= 0 || value >= unlimitedThreshold {
		return 0, false
	}

	return value, true
}
