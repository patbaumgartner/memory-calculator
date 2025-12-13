//go:build !minimal

// Package calc provides memory calculation logic.
package calc

// Build tag wrappers for standard build (use existing functions)
func matchDirectMemory(s string) bool {
	return MatchDirectMemory(s)
}

func matchHeap(s string) bool {
	return MatchHeap(s)
}

func matchMetaspace(s string) bool {
	return MatchMetaspace(s)
}

func matchReservedCodeCache(s string) bool {
	return MatchReservedCodeCache(s)
}

func matchStack(s string) bool {
	return MatchStack(s)
}

func parseDirectMemory(s string) (DirectMemory, error) {
	return ParseDirectMemory(s)
}

func parseHeap(s string) (Heap, error) {
	h, err := ParseHeap(s)
	if err != nil {
		return Heap{}, err
	}
	return *h, nil
}

func parseMetaspace(s string) (Metaspace, error) {
	m, err := ParseMetaspace(s)
	if err != nil {
		return Metaspace{}, err
	}
	return *m, nil
}

func parseReservedCodeCache(s string) (ReservedCodeCache, error) {
	return ParseReservedCodeCache(s)
}

func parseStack(s string) (Stack, error) {
	return ParseStack(s)
}
