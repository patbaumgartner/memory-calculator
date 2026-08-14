package calc

import (
	"fmt"
	"strings"
)

const (
	// DirectMemoryFlag is the JVM option controlling maximum direct (off-heap NIO) memory.
	DirectMemoryFlag = "-XX:MaxDirectMemorySize="
	// HeapFlag is the JVM option controlling maximum heap size.
	HeapFlag = "-Xmx"
	// MetaspaceFlag is the JVM option controlling maximum metaspace size.
	MetaspaceFlag = "-XX:MaxMetaspaceSize="
	// ReservedCodeCacheFlag is the JVM option controlling the JIT code cache reservation.
	ReservedCodeCacheFlag = "-XX:ReservedCodeCacheSize="
	// StackFlag is the JVM option controlling per-thread stack size.
	StackFlag = "-Xss"
)

var (
	// DefaultDirectMemory is the default direct memory size (10MB).
	DefaultDirectMemory = DirectMemory{Value: 10 * Mebi, Provenance: Default}
	// DefaultReservedCodeCache is the default reserved code cache size (240MB).
	DefaultReservedCodeCache = ReservedCodeCache{Value: 240 * Mebi, Provenance: Default}
	// DefaultStack is the default stack size (1MB).
	DefaultStack = Stack{Value: Mebi, Provenance: Default}
)

// DirectMemory represents the maximum direct memory size.
type DirectMemory Size

// Heap represents the heap memory size.
type Heap Size

// Metaspace represents the metaspace memory size.
type Metaspace Size

// ReservedCodeCache represents the reserved code cache memory size.
type ReservedCodeCache Size

// Stack represents the thread stack size.
type Stack Size

func (d DirectMemory) String() string      { return DirectMemoryFlag + Size(d).String() }
func (h Heap) String() string              { return HeapFlag + Size(h).String() }
func (m Metaspace) String() string         { return MetaspaceFlag + Size(m).String() }
func (r ReservedCodeCache) String() string { return ReservedCodeCacheFlag + Size(r).String() }
func (s Stack) String() string             { return StackFlag + Size(s).String() }

// MatchDirectMemory reports whether the option sets maximum direct memory.
func MatchDirectMemory(s string) bool { return matchFlag(s, DirectMemoryFlag) }

// MatchHeap reports whether the option sets maximum heap size.
func MatchHeap(s string) bool { return matchFlag(s, HeapFlag) }

// MatchMetaspace reports whether the option sets maximum metaspace size.
func MatchMetaspace(s string) bool { return matchFlag(s, MetaspaceFlag) }

// MatchReservedCodeCache reports whether the option sets the reserved code cache size.
func MatchReservedCodeCache(s string) bool { return matchFlag(s, ReservedCodeCacheFlag) }

// MatchStack reports whether the option sets the thread stack size.
func MatchStack(s string) bool { return matchFlag(s, StackFlag) }

// ParseDirectMemory parses a -XX:MaxDirectMemorySize option.
func ParseDirectMemory(s string) (DirectMemory, error) {
	z, err := parseFlag(s, DirectMemoryFlag)
	return DirectMemory(z), err
}

// ParseHeap parses a -Xmx option.
func ParseHeap(s string) (Heap, error) {
	z, err := parseFlag(s, HeapFlag)
	return Heap(z), err
}

// ParseMetaspace parses a -XX:MaxMetaspaceSize option.
func ParseMetaspace(s string) (Metaspace, error) {
	z, err := parseFlag(s, MetaspaceFlag)
	return Metaspace(z), err
}

// ParseReservedCodeCache parses a -XX:ReservedCodeCacheSize option.
func ParseReservedCodeCache(s string) (ReservedCodeCache, error) {
	z, err := parseFlag(s, ReservedCodeCacheFlag)
	return ReservedCodeCache(z), err
}

// ParseStack parses a -Xss option.
func ParseStack(s string) (Stack, error) {
	z, err := parseFlag(s, StackFlag)
	return Stack(z), err
}

// matchFlag recognizes an option by prefix alone, so that an option carrying an unusable value is
// still identified as user-configured. Recognizing it only when the whole option parses would let
// the calculator treat a malformed `-Xmx` as absent and append a second, conflicting one.
func matchFlag(s, prefix string) bool {
	return strings.HasPrefix(strings.TrimSpace(s), prefix)
}

func parseFlag(s, prefix string) (Size, error) {
	t := strings.TrimSpace(s)

	if !strings.HasPrefix(t, prefix) {
		return Size{}, fmt.Errorf("%q is not a %s option", s, strings.TrimSuffix(prefix, "="))
	}

	z, err := ParseSize(strings.TrimPrefix(t, prefix))
	if err != nil {
		return Size{}, fmt.Errorf("invalid value in JVM option %q\n%w", t, err)
	}
	if z.Value <= 0 {
		return Size{}, fmt.Errorf("JVM option %q must be greater than zero", t)
	}

	return z, nil
}
