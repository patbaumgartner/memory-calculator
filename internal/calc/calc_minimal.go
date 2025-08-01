//go:build minimal

// Minimal memory calculation without regex dependencies
package calc

import (
	"fmt"
	"strconv"
	"strings"
)

// Build tag wrappers for minimal build
func matchDirectMemory(s string) bool {
	return MatchDirectMemorySimple(s)
}

func matchHeap(s string) bool {
	return MatchHeapSimple(s)
}

func matchMetaspace(s string) bool {
	return MatchMetaspaceSimple(s)
}

func matchReservedCodeCache(s string) bool {
	return MatchReservedCodeCacheSimple(s)
}

func matchStack(s string) bool {
	return MatchStackSimple(s)
}

func parseDirectMemory(s string) (DirectMemory, error) {
	return ParseDirectMemorySimple(s)
}

func parseHeap(s string) (Heap, error) {
	return ParseHeapSimple(s)
}

func parseMetaspace(s string) (Metaspace, error) {
	return ParseMetaspaceSimple(s)
}

func parseReservedCodeCache(s string) (ReservedCodeCache, error) {
	return ParseReservedCodeCacheSimple(s)
}

func parseStack(s string) (Stack, error) {
	return ParseStackSimple(s)
}

// Simplified memory parsing without regex
func ParseSizeSimple(s string) (Size, error) {
	s = strings.TrimSpace(strings.ToUpper(s))

	if len(s) == 0 {
		return Size{}, fmt.Errorf("empty size string")
	}

	// Handle pure numbers (bytes)
	if num, err := strconv.ParseInt(s, 10, 64); err == nil {
		return Size{Value: num, Provenance: UserConfigured}, nil
	}

	// Simple suffix handling
	var multiplier int64 = 1
	var numStr string

	if strings.HasSuffix(s, "G") || strings.HasSuffix(s, "GB") {
		multiplier = Gibi
		numStr = strings.TrimSuffix(strings.TrimSuffix(s, "GB"), "G")
	} else if strings.HasSuffix(s, "M") || strings.HasSuffix(s, "MB") {
		multiplier = Mebi
		numStr = strings.TrimSuffix(strings.TrimSuffix(s, "MB"), "M")
	} else if strings.HasSuffix(s, "K") || strings.HasSuffix(s, "KB") {
		multiplier = Kibi
		numStr = strings.TrimSuffix(strings.TrimSuffix(s, "KB"), "K")
	} else if strings.HasSuffix(s, "B") {
		multiplier = 1
		numStr = strings.TrimSuffix(s, "B")
	} else {
		return Size{}, fmt.Errorf("invalid size format: %s", s)
	}

	if num, err := strconv.ParseFloat(numStr, 64); err == nil {
		return Size{Value: int64(num * float64(multiplier)), Provenance: UserConfigured}, nil
	}

	return Size{}, fmt.Errorf("invalid number in size: %s", s)
}

// Simplified matching without regex
func MatchDirectMemorySimple(s string) bool {
	return strings.HasPrefix(s, "-XX:MaxDirectMemorySize=")
}

func MatchHeapSimple(s string) bool {
	return strings.HasPrefix(s, "-Xmx")
}

func MatchMetaspaceSimple(s string) bool {
	return strings.HasPrefix(s, "-XX:MaxMetaspaceSize=")
}

func MatchReservedCodeCacheSimple(s string) bool {
	return strings.HasPrefix(s, "-XX:ReservedCodeCacheSize=")
}

func MatchStackSimple(s string) bool {
	return strings.HasPrefix(s, "-Xss")
}

func ParseDirectMemorySimple(s string) (DirectMemory, error) {
	if !strings.HasPrefix(s, "-XX:MaxDirectMemorySize=") {
		return DirectMemory{}, fmt.Errorf("invalid direct memory flag: %s", s)
	}

	sizeStr := strings.TrimPrefix(s, "-XX:MaxDirectMemorySize=")
	size, err := ParseSizeSimple(sizeStr)
	if err != nil {
		return DirectMemory{}, err
	}

	return DirectMemory(size), nil
}

func ParseHeapSimple(s string) (Heap, error) {
	if !strings.HasPrefix(s, "-Xmx") {
		return Heap{}, fmt.Errorf("invalid heap flag: %s", s)
	}

	sizeStr := strings.TrimPrefix(s, "-Xmx")
	size, err := ParseSizeSimple(sizeStr)
	if err != nil {
		return Heap{}, err
	}

	return Heap(size), nil
}

func ParseMetaspaceSimple(s string) (Metaspace, error) {
	if !strings.HasPrefix(s, "-XX:MaxMetaspaceSize=") {
		return Metaspace{}, fmt.Errorf("invalid metaspace flag: %s", s)
	}

	sizeStr := strings.TrimPrefix(s, "-XX:MaxMetaspaceSize=")
	size, err := ParseSizeSimple(sizeStr)
	if err != nil {
		return Metaspace{}, err
	}

	return Metaspace(size), nil
}

func ParseReservedCodeCacheSimple(s string) (ReservedCodeCache, error) {
	if !strings.HasPrefix(s, "-XX:ReservedCodeCacheSize=") {
		return ReservedCodeCache{}, fmt.Errorf("invalid reserved code cache flag: %s", s)
	}

	sizeStr := strings.TrimPrefix(s, "-XX:ReservedCodeCacheSize=")
	size, err := ParseSizeSimple(sizeStr)
	if err != nil {
		return ReservedCodeCache{}, err
	}

	return ReservedCodeCache(size), nil
}

func ParseStackSimple(s string) (Stack, error) {
	if !strings.HasPrefix(s, "-Xss") {
		return Stack{}, fmt.Errorf("invalid stack flag: %s", s)
	}

	sizeStr := strings.TrimPrefix(s, "-Xss")
	size, err := ParseSizeSimple(sizeStr)
	if err != nil {
		return Stack{}, err
	}

	return Stack(size), nil
}
