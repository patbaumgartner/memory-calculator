package calculator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patbaumgartner/memory-calculator/internal/cgroups"
	"github.com/patbaumgartner/memory-calculator/internal/host"
	"github.com/patbaumgartner/memory-calculator/internal/logger"
)

// calculatorAt builds a MemoryCalculator whose detection reads from a fixture tree, so the whole
// path from a cgroup file to the emitted JVM options can be exercised without a container.
func calculatorAt(t *testing.T, files map[string]string) *MemoryCalculator {
	t.Helper()

	root := t.TempDir()
	for name, contents := range files {
		full := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
			t.Fatalf("MkdirAll(%s): %v", full, err)
		}
		if err := os.WriteFile(full, []byte(contents), 0o600); err != nil {
			t.Fatalf("WriteFile(%s): %v", full, err)
		}
	}

	return &MemoryCalculator{
		Logger: logger.Create(true),
		Detector: &cgroups.Detector{
			V2Root:       filepath.Join(root, "v2"),
			V1Root:       filepath.Join(root, "v1"),
			ProcCgroup:   filepath.Join(root, "proc-cgroup"),
			HostDetector: host.CreateWithPath(filepath.Join(root, "meminfo")),
		},
	}
}

func heapFrom(t *testing.T, options string) string {
	t.Helper()

	for _, field := range strings.Fields(options) {
		if strings.HasPrefix(field, "-Xmx") {
			return field
		}
	}

	t.Fatalf("no -Xmx in %q", options)
	return ""
}

// TestExecuteSizesHeapFromDetectedLimit is the end-to-end guarantee of the tool: the container's
// memory limit, wherever it is expressed, determines the heap.
func TestExecuteSizesHeapFromDetectedLimit(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{
			name:  "cgroup v2 limit sizes the heap",
			files: map[string]string{"v2/memory.max": "536870912\n"},
			want:  "-Xmx175095K",
		},
		{
			name:  "cgroup v1 limit sizes the heap",
			files: map[string]string{"v1/memory.limit_in_bytes": "536870912\n"},
			want:  "-Xmx175095K",
		},
		{
			name: "an ancestor limit sizes the heap when the leaf is unlimited",
			files: map[string]string{
				"proc-cgroup":                   "0::/kubepods/pod123\n",
				"v2/memory.max":                 "max\n",
				"v2/kubepods/memory.max":        "max\n",
				"v2/kubepods/pod123/memory.max": "536870912\n",
			},
			want: "-Xmx175095K",
		},
		{
			name: "host memory is used when no cgroup limit applies",
			files: map[string]string{
				"v2/memory.max": "max\n",
				"meminfo":       "MemTotal: 2000000 kB\nMemAvailable: 524288 kB\n",
			},
			want: "-Xmx175095K",
		},
		{
			name:  "the documented default is used when nothing can be detected",
			files: map[string]string{},
			want:  "-Xmx699383K",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := Input{
				ThreadCount:      50,
				LoadedClassCount: intPtr(5000),
				ApplicationPath:  t.TempDir(),
				JVMClassCount:    1000,
				AdjustmentFactor: 100,
			}

			result, err := calculatorAt(t, tt.files).Execute(input)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if got := heapFrom(t, result.JavaToolOptions); got != tt.want {
				t.Errorf("heap = %s, want %s (from %s)", got, tt.want, result.JavaToolOptions)
			}
		})
	}
}

// TestExecuteIgnoresUnlimitedCgroupValues covers the sentinels that mean "no limit". Reading one as
// a real limit would size the JVM for terabytes and the container would be killed on startup.
func TestExecuteIgnoresUnlimitedCgroupValues(t *testing.T) {
	sentinels := []string{
		"9223372036854771712",
		"9223372036854710272",
		"18446744073709551615",
		"-1",
	}

	for _, sentinel := range sentinels {
		t.Run(sentinel, func(t *testing.T) {
			input := Input{
				ThreadCount:      50,
				LoadedClassCount: intPtr(5000),
				ApplicationPath:  t.TempDir(),
				JVMClassCount:    1000,
				AdjustmentFactor: 100,
			}

			files := map[string]string{
				"v1/memory.limit_in_bytes": sentinel + "\n",
				"meminfo":                  "MemAvailable: 524288 kB\n",
			}

			result, err := calculatorAt(t, files).Execute(input)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if got, want := heapFrom(t, result.JavaToolOptions), "-Xmx175095K"; got != want {
				t.Errorf("heap = %s, want %s: %q was treated as a real limit", got, want, sentinel)
			}
		})
	}
}
