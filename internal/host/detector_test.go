package host

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func writeMemInfo(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "meminfo")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	return path
}

func TestDetectAvailableMemory(t *testing.T) {
	if runtime.GOOS != platformLinux {
		t.Skipf("host memory detection is Linux-only; running on %s", runtime.GOOS)
	}

	tests := []struct {
		name     string
		meminfo  string
		expected int64
	}{
		{
			name:     "reads MemAvailable rather than MemTotal",
			meminfo:  "MemTotal:       16000000 kB\nMemAvailable:    8000000 kB\n",
			expected: 8000000 * 1024,
		},
		{
			name:     "falls back to MemTotal when MemAvailable is absent",
			meminfo:  "MemTotal:       16000000 kB\nSwapTotal:       1000000 kB\n",
			expected: 16000000 * 1024,
		},
		{
			name:     "handles a value without a unit",
			meminfo:  "MemAvailable:    4096\n",
			expected: 4096,
		},
		{
			name:     "handles an MB unit",
			meminfo:  "MemAvailable:    2048 MB\n",
			expected: 2048 * 1024 * 1024,
		},
		{
			name:     "tolerates surrounding entries",
			meminfo:  "Buffers: 1 kB\nMemAvailable: 512 kB\nCached: 2 kB\n",
			expected: 512 * 1024,
		},
		{
			name:     "returns zero for an empty file",
			meminfo:  "",
			expected: 0,
		},
		{
			name:     "returns zero when the value is not a number",
			meminfo:  "MemAvailable:    lots kB\n",
			expected: 0,
		},
		{
			name:     "returns zero when the value is missing",
			meminfo:  "MemAvailable:\n",
			expected: 0,
		},
		{
			name:     "returns zero for a negative value",
			meminfo:  "MemAvailable:    -1 kB\n",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateWithPath(writeMemInfo(t, tt.meminfo)).DetectAvailableMemory()

			if got != tt.expected {
				t.Errorf("DetectAvailableMemory() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestDetectAvailableMemoryMissingFile(t *testing.T) {
	got := CreateWithPath(filepath.Join(t.TempDir(), "absent")).DetectAvailableMemory()

	if got != 0 {
		t.Errorf("DetectAvailableMemory() = %d, want 0 when /proc/meminfo is unreadable", got)
	}
}

// TestDetectAvailableMemoryIsLinuxOnly locks the decision not to guess on other platforms.
// The previous implementation multiplied Go's runtime heap statistics by 32 and reported that as
// physical memory, which is unrelated to how much memory the machine has.
func TestDetectAvailableMemoryIsLinuxOnly(t *testing.T) {
	if runtime.GOOS == platformLinux {
		t.Skip("platform-specific behaviour only observable off Linux")
	}

	got := CreateWithPath(writeMemInfo(t, "MemAvailable: 8000000 kB\n")).DetectAvailableMemory()

	if got != 0 {
		t.Errorf("DetectAvailableMemory() = %d, want 0 on %s", got, runtime.GOOS)
	}
}

func TestCreateUsesLinuxMemInfoPath(t *testing.T) {
	if got := Create().MemInfoPath; got != LinuxMemInfoPath {
		t.Errorf("Create().MemInfoPath = %q, want %q", got, LinuxMemInfoPath)
	}
}

func TestIsHostMemoryDetectionSupported(t *testing.T) {
	if got, want := IsHostMemoryDetectionSupported(), runtime.GOOS == platformLinux; got != want {
		t.Errorf("IsHostMemoryDetectionSupported() = %t, want %t on %s", got, want, runtime.GOOS)
	}
}
