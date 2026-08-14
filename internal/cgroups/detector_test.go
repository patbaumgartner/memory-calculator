package cgroups

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/patbaumgartner/memory-calculator/internal/host"
)

// fixture builds a cgroup tree on disk. Keys are paths relative to the temp root.
func fixture(t *testing.T, files map[string]string) string {
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

	return root
}

// detector wires a Detector at the fixture root with host detection disabled unless given.
func detector(root string, hostDetector *host.Detector) *Detector {
	return &Detector{
		V2Root:       filepath.Join(root, "v2"),
		V1Root:       filepath.Join(root, "v1"),
		ProcCgroup:   filepath.Join(root, "proc-cgroup"),
		HostDetector: hostDetector,
	}
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		wantLimit  int64
		wantSource Source
	}{
		{
			name:       "cgroup v2 limit",
			files:      map[string]string{"v2/memory.max": "536870912\n"},
			wantLimit:  536870912,
			wantSource: SourceCgroupV2,
		},
		{
			name:       "cgroup v2 unlimited falls through",
			files:      map[string]string{"v2/memory.max": "max\n"},
			wantSource: SourceNone,
		},
		{
			name:       "cgroup v1 limit",
			files:      map[string]string{"v1/memory.limit_in_bytes": "268435456\n"},
			wantLimit:  268435456,
			wantSource: SourceCgroupV1,
		},
		{
			name: "cgroup v2 wins over v1 on a hybrid system",
			files: map[string]string{
				"v2/memory.max":            "536870912\n",
				"v1/memory.limit_in_bytes": "268435456\n",
			},
			wantLimit:  536870912,
			wantSource: SourceCgroupV2,
		},
		{
			name:       "cgroup v1 page-counter maximum means unlimited",
			files:      map[string]string{"v1/memory.limit_in_bytes": "9223372036854771712\n"},
			wantSource: SourceNone,
		},
		{
			name:       "cgroup v1 64KiB-page maximum means unlimited",
			files:      map[string]string{"v1/memory.limit_in_bytes": "9223372036854710272\n"},
			wantSource: SourceNone,
		},
		{
			name:       "cgroup v1 unsigned overflow means unlimited",
			files:      map[string]string{"v1/memory.limit_in_bytes": "18446744073709551615\n"},
			wantSource: SourceNone,
		},
		{
			name:       "cgroup v1 negative one means unlimited",
			files:      map[string]string{"v1/memory.limit_in_bytes": "-1\n"},
			wantSource: SourceNone,
		},
		{
			name:       "malformed limit is ignored",
			files:      map[string]string{"v2/memory.max": "not-a-number\n"},
			wantSource: SourceNone,
		},
		{
			name:       "empty limit file is ignored",
			files:      map[string]string{"v2/memory.max": ""},
			wantSource: SourceNone,
		},
		{
			name:       "missing files yield no detection",
			files:      map[string]string{},
			wantSource: SourceNone,
		},
		{
			name:       "a limit above 1TiB is still a real limit",
			files:      map[string]string{"v2/memory.max": "2199023255552\n"},
			wantLimit:  2199023255552,
			wantSource: SourceCgroupV2,
		},
		{
			name: "an ancestor limit constrains an unlimited leaf",
			files: map[string]string{
				"proc-cgroup":                                "0::/kubepods/pod123/container456\n",
				"v2/memory.max":                              "max\n",
				"v2/kubepods/memory.max":                     "max\n",
				"v2/kubepods/pod123/memory.max":              "268435456\n",
				"v2/kubepods/pod123/container456/memory.max": "max\n",
			},
			wantLimit:  268435456,
			wantSource: SourceCgroupV2,
		},
		{
			name: "the tightest limit in the hierarchy wins",
			files: map[string]string{
				"proc-cgroup":                                "0::/kubepods/pod123/container456\n",
				"v2/memory.max":                              "max\n",
				"v2/kubepods/memory.max":                     "8589934592\n",
				"v2/kubepods/pod123/memory.max":              "1073741824\n",
				"v2/kubepods/pod123/container456/memory.max": "536870912\n",
			},
			wantLimit:  536870912,
			wantSource: SourceCgroupV2,
		},
		{
			name: "a leaf tighter than its ancestors wins",
			files: map[string]string{
				"proc-cgroup":                   "0::/kubepods/pod123\n",
				"v2/memory.max":                 "max\n",
				"v2/kubepods/memory.max":        "8589934592\n",
				"v2/kubepods/pod123/memory.max": "134217728\n",
			},
			wantLimit:  134217728,
			wantSource: SourceCgroupV2,
		},
		{
			name: "cgroup v1 membership is read from the memory controller line",
			files: map[string]string{
				"proc-cgroup":                         "12:pids:/other\n4:cpu,memory:/docker/abc\n3:cpuset:/wrong\n",
				"v1/memory.limit_in_bytes":            "9223372036854771712\n",
				"v1/docker/abc/memory.limit_in_bytes": "67108864\n",
			},
			wantLimit:  67108864,
			wantSource: SourceCgroupV1,
		},
		{
			name: "a unified line is not used for cgroup v1 lookups",
			files: map[string]string{
				"proc-cgroup":                         "0::/docker/abc\n",
				"v1/memory.limit_in_bytes":            "9223372036854771712\n",
				"v1/docker/abc/memory.limit_in_bytes": "67108864\n",
			},
			wantSource: SourceNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector(fixture(t, tt.files), nil).Detect()

			if got.Source != tt.wantSource {
				t.Errorf("Source = %q, want %q", got.Source, tt.wantSource)
			}
			if got.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", got.Limit, tt.wantLimit)
			}
			if got.Found() != (tt.wantLimit > 0) {
				t.Errorf("Found() = %t, want %t", got.Found(), tt.wantLimit > 0)
			}
		})
	}
}

func TestDetectFallsBackToHost(t *testing.T) {
	root := fixture(t, map[string]string{"v2/memory.max": "max\n"})
	meminfo := filepath.Join(root, "meminfo")
	if err := os.WriteFile(meminfo, []byte("MemTotal: 16000000 kB\nMemAvailable: 8000000 kB\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got := detector(root, host.CreateWithPath(meminfo)).Detect()

	if got.Source != SourceHost {
		t.Errorf("Source = %q, want %q", got.Source, SourceHost)
	}
	if want := int64(8000000 * 1024); got.Limit != want {
		t.Errorf("Limit = %d, want %d (MemAvailable, not MemTotal)", got.Limit, want)
	}
}

func TestDetectPrefersCgroupOverHost(t *testing.T) {
	root := fixture(t, map[string]string{"v2/memory.max": "268435456\n"})
	meminfo := filepath.Join(root, "meminfo")
	if err := os.WriteFile(meminfo, []byte("MemAvailable: 8000000 kB\n"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got := detector(root, host.CreateWithPath(meminfo)).Detect()

	if got.Source != SourceCgroupV2 || got.Limit != 268435456 {
		t.Errorf("Detect() = %+v, want the cgroup limit rather than host memory", got)
	}
}

func TestCreateUsesStandardPaths(t *testing.T) {
	d := Create()

	if d.V2Root != DefaultV2Root || d.V1Root != DefaultV1Root || d.ProcCgroup != DefaultProcCgroup {
		t.Errorf("Create() = %+v, want the documented default paths", d)
	}
	if d.HostDetector == nil {
		t.Error("Create() left HostDetector nil, so host fallback would never run")
	}
}
