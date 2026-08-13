package calc

import (
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
