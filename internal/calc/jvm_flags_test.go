package calc

import (
	"math"
	"testing"
)

func TestFlagMatchersSelectTheRightRegion(t *testing.T) {
	matchers := map[string]func(string) bool{
		"DirectMemory":      MatchDirectMemory,
		"Heap":              MatchHeap,
		"Metaspace":         MatchMetaspace,
		"ReservedCodeCache": MatchReservedCodeCache,
		"Stack":             MatchStack,
	}

	tests := []struct {
		name  string
		flag  string
		want  string
		match bool
	}{
		{"Direct memory", "-XX:MaxDirectMemorySize=512M", "DirectMemory", true},
		{"Heap", "-Xmx2G", "Heap", true},
		{"Metaspace", "-XX:MaxMetaspaceSize=256M", "Metaspace", true},
		{"Reserved code cache", "-XX:ReservedCodeCacheSize=128M", "ReservedCodeCache", true},
		{"Stack", "-Xss2M", "Stack", true},
		{"Unknown direct memory spelling", "-XX:MaxDirectMemory=512M", "", false},
		{"Unknown heap spelling", "-XX:MaxHeapSize=2G", "", false},
		{"Unknown metaspace spelling", "-XX:Metaspace=256M", "", false},
		{"Unknown code cache spelling", "-XX:CodeCacheSize=128M", "", false},
		{"Unknown stack spelling", "-XX:ThreadStackSize=2M", "", false},
		{"Initial heap is not max heap", "-Xms1G", "", false},
		{"Unrelated flag", "-Djava.awt.headless=true", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matched []string
			for name, match := range matchers {
				if match(tt.flag) {
					matched = append(matched, name)
				}
			}

			if !tt.match {
				if len(matched) != 0 {
					t.Errorf("%q matched %v, want no match", tt.flag, matched)
				}
				return
			}

			if len(matched) != 1 || matched[0] != tt.want {
				t.Errorf("%q matched %v, want exactly [%s]", tt.flag, matched, tt.want)
			}
		})
	}
}

func TestCalculateAppliesEachFlagToItsRegion(t *testing.T) {
	tests := []struct {
		flag string
		size int64
		get  func(MemoryRegions) int64
	}{
		{"-Xmx1G", Gibi, func(m MemoryRegions) int64 { return m.Heap.Value }},
		{"-Xmx512M", 512 * Mebi, func(m MemoryRegions) int64 { return m.Heap.Value }},
		{"-XX:MaxMetaspaceSize=128M", 128 * Mebi, func(m MemoryRegions) int64 { return m.Metaspace.Value }},
		{"-XX:MaxDirectMemorySize=64M", 64 * Mebi, func(m MemoryRegions) int64 { return m.DirectMemory.Value }},
		{"-XX:ReservedCodeCacheSize=96M", 96 * Mebi, func(m MemoryRegions) int64 { return m.ReservedCodeCache.Value }},
		{"-Xss1M", Mebi, func(m MemoryRegions) int64 { return m.Stack.Value }},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			c := Calculator{
				TotalMemory:      Size{Value: 4 * Gibi},
				ThreadCount:      100,
				LoadedClassCount: 10000,
			}

			result, err := c.Calculate(tt.flag)
			if err != nil {
				t.Fatalf("Calculate(%q) error = %v", tt.flag, err)
			}

			if got := tt.get(result); got != tt.size {
				t.Errorf("Size for %q = %d, want %d", tt.flag, got, tt.size)
			}
		})
	}
}

// TestUnparseableFlagIsRejected covers options the calculator recognises but cannot read.
// Treating them as absent would leave the user's flag in JAVA_TOOL_OPTIONS and append a second,
// conflicting one, and the JVM honours whichever comes last.
func TestUnparseableFlagIsRejected(t *testing.T) {
	flags := []string{
		"-Xmxbogus",
		"-Xmx1.5G",
		"-Xmx-1G",
		"-Xmx",
		"-Xss2X",
		"-XX:MaxMetaspaceSize=lots",
		"-XX:MaxDirectMemorySize=",
		"-XX:ReservedCodeCacheSize=1.5G",
	}

	for _, flag := range flags {
		t.Run(flag, func(t *testing.T) {
			c := Calculator{
				TotalMemory:      Size{Value: 4 * Gibi},
				ThreadCount:      100,
				LoadedClassCount: 10000,
			}

			if _, err := c.Calculate(flag); err == nil {
				t.Errorf("Calculate(%q) succeeded, want an error", flag)
			}
		})
	}
}

// TestCalculateRejectsInputsThatInflateHeap covers inputs that make a region negative.
// Region sizes are subtracted from total memory, so a negative one grows the heap past the
// container limit and yields a plausible -Xmx the container cannot honour.
func TestCalculateRejectsInputsThatInflateHeap(t *testing.T) {
	valid := Calculator{
		TotalMemory:      Size{Value: 2 * Gibi},
		ThreadCount:      250,
		LoadedClassCount: 5000,
		HeadRoom:         0,
	}

	if _, err := valid.Calculate(""); err != nil {
		t.Fatalf("baseline Calculate() error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Calculator)
	}{
		{"negative head room", func(c *Calculator) { c.HeadRoom = -100 }},
		{"head room above the maximum", func(c *Calculator) { c.HeadRoom = 100 }},
		{"negative thread count", func(c *Calculator) { c.ThreadCount = -1000 }},
		{"negative loaded class count", func(c *Calculator) { c.LoadedClassCount = -100000 }},
		{"zero total memory", func(c *Calculator) { c.TotalMemory = Size{} }},
		{"negative total memory", func(c *Calculator) { c.TotalMemory = Size{Value: -Gibi} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := valid
			tt.mutate(&c)

			result, err := c.Calculate("")
			if err == nil {
				t.Fatalf("Calculate() succeeded with heap %v, want an error", result.Heap)
			}
		})
	}
}

func TestCalculateRejectsArithmeticOverflow(t *testing.T) {
	tests := []struct {
		name       string
		calculator Calculator
		flags      string
	}{
		{
			name: "thread stack multiplication",
			calculator: Calculator{
				TotalMemory:      Size{Value: Gibi},
				ThreadCount:      math.MaxInt,
				LoadedClassCount: 1000,
			},
		},
		{
			name: "metaspace multiplication",
			calculator: Calculator{
				TotalMemory:      Size{Value: Gibi},
				ThreadCount:      1,
				LoadedClassCount: math.MaxInt,
			},
		},
		{
			name: "user heap addition",
			calculator: Calculator{
				TotalMemory:      Size{Value: Gibi},
				ThreadCount:      1,
				LoadedClassCount: 1000,
			},
			flags: "-Xmx9223372036854775807",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result, err := tt.calculator.Calculate(tt.flags); err == nil {
				t.Fatalf("Calculate() succeeded with %+v, want an overflow error", result)
			}
		})
	}
}

func TestCalculateRejectsZeroHeapBoundary(t *testing.T) {
	calculator := Calculator{
		TotalMemory:      Size{Value: 277_198_376},
		ThreadCount:      1,
		LoadedClassCount: 1000,
	}

	if result, err := calculator.Calculate(""); err == nil {
		t.Fatalf("Calculate() succeeded with heap %v, want an error", result.Heap)
	}
}

func TestCalculateRejectsZeroValuedJVMRegions(t *testing.T) {
	for _, flag := range []string{
		"-Xmx0",
		"-Xss0",
		"-XX:MaxMetaspaceSize=0",
		"-XX:MaxDirectMemorySize=0",
		"-XX:ReservedCodeCacheSize=0",
	} {
		t.Run(flag, func(t *testing.T) {
			calculator := Calculator{
				TotalMemory:      Size{Value: 2 * Gibi},
				ThreadCount:      1,
				LoadedClassCount: 1000,
			}
			if _, err := calculator.Calculate(flag); err == nil {
				t.Errorf("Calculate(%q) succeeded, want an error", flag)
			}
		})
	}
}

// TestCalculateNeverExceedsTotalMemory is the invariant the whole tool rests on: whatever the
// inputs, the regions it reports must fit inside the memory it was given.
func TestCalculateNeverExceedsTotalMemory(t *testing.T) {
	totals := []int64{256 * Mebi, Gibi, 2 * Gibi, 8 * Gibi, 64 * Gibi}
	threads := []int{0, 1, 50, 250, 1000}
	classes := []int{0, 1000, 35000, 200000}
	headRooms := []int{0, 5, 25, 50, 99}

	for _, total := range totals {
		for _, thread := range threads {
			for _, class := range classes {
				for _, headRoom := range headRooms {
					c := Calculator{
						TotalMemory:      Size{Value: total},
						ThreadCount:      thread,
						LoadedClassCount: class,
						HeadRoom:         headRoom,
					}

					result, err := c.Calculate("")
					if err != nil {
						continue // Configuration does not fit; Calculate correctly refused it.
					}

					used, err := result.AllRegionsSize(thread)
					if err != nil {
						t.Fatalf("AllRegionsSize() error = %v", err)
					}

					if used.Value > total {
						t.Errorf("total=%d threads=%d classes=%d headRoom=%d: allocated %d bytes",
							total, thread, class, headRoom, used.Value)
					}

					if result.Heap.Value <= 0 {
						t.Errorf("total=%d threads=%d classes=%d headRoom=%d: heap %d is not positive",
							total, thread, class, headRoom, result.Heap.Value)
					}
				}
			}
		}
	}
}

func TestCalculateHonoursMultipleUserFlags(t *testing.T) {
	c := Calculator{
		TotalMemory:      Size{Value: 2 * Gibi},
		ThreadCount:      100,
		LoadedClassCount: 10000,
	}

	result, err := c.Calculate("-Xmx1G -XX:MaxMetaspaceSize=256M -Xss2M")
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	if result.Heap == nil || result.Heap.Value != Gibi {
		t.Errorf("Heap = %v, want %d", result.Heap, Gibi)
	}
	if result.Metaspace == nil || result.Metaspace.Value != 256*Mebi {
		t.Errorf("Metaspace = %v, want %d", result.Metaspace, 256*Mebi)
	}
	if result.Stack.Value != 2*Mebi {
		t.Errorf("Stack = %d, want %d", result.Stack.Value, 2*Mebi)
	}

	for name, p := range map[string]Provenance{
		"Heap":      result.Heap.Provenance,
		"Metaspace": result.Metaspace.Provenance,
		"Stack":     result.Stack.Provenance,
	} {
		if p != UserConfigured {
			t.Errorf("%s provenance = %v, want UserConfigured", name, p)
		}
	}
}
