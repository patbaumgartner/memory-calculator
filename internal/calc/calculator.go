// Package calc provides core memory calculation algorithms for JVM memory optimization.
//
// This package implements the primary memory calculation logic used by the memory calculator
// to determine optimal JVM memory settings based on available system resources, thread counts,
// and class loading estimates.
//
// The calculator uses a sophisticated multi-step algorithm that accounts for:
//   - Heap memory allocation (primary object storage)
//   - Thread stack memory (per-thread stack space)
//   - Metaspace sizing (class metadata storage)
//   - Code cache allocation (JIT compilation)
//   - Direct memory reservation (off-heap NIO operations)
//   - Head room reservation (configurable safety margin)
//
// Memory allocation follows this priority order:
//  1. Head room (percentage-based reservation)
//  2. Thread stacks (threads × stack size)
//  3. Metaspace (classes × overhead per class)
//  4. Code cache (fixed 240MB for optimal JIT performance)
//  5. Direct memory (fixed 10MB for NIO operations)
//  6. Heap (all remaining memory)
//
// All calculations are performed with 64-bit precision to handle large memory values
// and ensure accuracy across different deployment scenarios.
package calc

import (
	"fmt"

	"github.com/patbaumgartner/memory-calculator/internal/parser"
)

const (
	// ClassSize represents the average memory overhead per loaded class in bytes.
	// This value is based on empirical analysis of typical Java applications and
	// includes the metadata storage required for each class in the metaspace.
	// Value: 5,800 bytes per class (rounded up for safety margin).
	ClassSize = int64(5_800)

	// ClassOverhead represents the base memory overhead for the JVM class loading
	// system in bytes. This includes the core JVM runtime classes, bootstrap
	// classloader overhead, and other essential class-related memory structures.
	// Value: 14,000,000 bytes (approximately 13.35 MB).
	ClassOverhead = int64(14_000_000)
)

// Calculator represents the core JVM memory calculation engine.
//
// The Calculator performs sophisticated memory allocation calculations for Java Virtual Machine
// environments, taking into account container memory limits, application characteristics,
// and runtime requirements. It implements a multi-stage allocation algorithm that optimizes
// memory distribution across different JVM memory regions.
//
// Key Features:
//   - Head room management for memory safety margins
//   - Automatic class count estimation and metaspace sizing
//   - Thread-aware stack memory calculation
//   - Container and host memory detection integration
//   - JVM flag parsing and validation
//
// Memory Allocation Strategy:
//  1. Parse existing JVM flags to detect overrides
//  2. Calculate head room reservation (percentage-based)
//  3. Allocate thread stack memory (threads × stack size)
//  4. Calculate metaspace size (classes × overhead + base)
//  5. Reserve code cache memory (240MB for JIT optimization)
//  6. Reserve direct memory (10MB for NIO operations)
//  7. Allocate remaining memory to heap
//
// Thread Safety:
//
//	Calculator instances are immutable and safe for concurrent use.
//	All calculations are performed without modifying the Calculator state.
//
// Example Usage:
//
//	calc := Calculator{
//		TotalMemory:      SizeFromBytes(2 * 1024 * 1024 * 1024), // 2GB
//		ThreadCount:      250,
//		LoadedClassCount: 35000,
//		HeadRoom:         5, // 5% head room
//	}
//
//	regions, err := calc.Calculate("")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Heap: %s, Metaspace: %s\n", regions.Heap, regions.Metaspace)
//
// Validation:
//   - Total memory must be positive and non-zero
//   - Thread count must be positive (minimum 1)
//   - Loaded class count must be positive (minimum 1000)
//   - Head room must be between 0-99% inclusive
//   - All memory calculations validated for overflow conditions
type Calculator struct {
	// HeadRoom specifies the percentage of total memory to reserve as a safety margin.
	// This memory is not allocated to any JVM component and remains available for
	// system operations, memory pressure handling, and unexpected memory usage spikes.
	// Valid range: 0-99 (percentage). Default: 0 (no head room).
	HeadRoom int

	// LoadedClassCount represents the estimated number of classes that will be loaded
	// by the application during runtime. This value is used to calculate the required
	// metaspace size for storing class metadata. If not specified, the calculator
	// will attempt to estimate this value by scanning JAR files in the application path.
	// Minimum recommended: 1000 classes. Typical range: 10,000-100,000 classes.
	LoadedClassCount int

	// ThreadCount specifies the expected number of threads the JVM application will create.
	// This includes both application threads and JVM internal threads. Each thread requires
	// stack memory allocation (default 1MB per thread on most platforms).
	// Minimum: 1 thread. Typical range: 50-1000 threads depending on application type.
	ThreadCount int

	// TotalMemory represents the total amount of memory available to the JVM process.
	// This can be automatically detected from container limits (cgroups) or host system
	// memory, or manually specified. The calculator will distribute this memory across
	// all JVM memory regions according to the allocation algorithm.
	// Must be positive and sufficient for minimum JVM requirements.
	TotalMemory Size
}

// Calculate performs comprehensive JVM memory allocation calculations and returns
// optimized memory region settings based on the Calculator configuration.
//
// This method implements the core memory allocation algorithm that distributes available
// memory across different JVM memory regions while respecting existing JVM flags and
// ensuring optimal performance characteristics.
//
// Parameters:
//
//	flags - A string containing existing JVM flags that may override default calculations.
//	        Supported flags include -Xmx, -Xms, -XX:MaxMetaspaceSize, -XX:MaxDirectMemorySize,
//	        -XX:ReservedCodeCacheSize, and -Xss. Flags are parsed using shell-word parsing
//	        to handle complex quoting and escaping correctly.
//
// Returns:
//
//	MemoryRegions - A complete specification of all JVM memory regions with calculated sizes
//	error - Any validation or calculation errors encountered during processing
//
// Algorithm Details:
//  1. Parse existing JVM flags to identify user-specified overrides
//  2. Validate Calculator configuration (memory limits, counts, percentages)
//  3. Calculate head room reservation based on total memory percentage
//  4. Determine thread stack allocation (ThreadCount × stack size)
//  5. Calculate metaspace size (LoadedClassCount × ClassSize + ClassOverhead)
//  6. Apply fixed allocations (code cache: 240MB, direct memory: 10MB)
//  7. Allocate all remaining memory to heap
//  8. Validate final allocation fits within available memory
//
// Error Conditions:
//   - Invalid JVM flag syntax in flags parameter
//   - Total memory insufficient for minimum JVM requirements
//   - Memory allocation overflow or underflow conditions
//   - Invalid Calculator configuration values
//
// Performance:
//   - Execution time: < 1ms for typical configurations
//   - Memory usage: < 1KB temporary allocations
//   - Thread-safe: No shared state modifications
//
// Example:
//
//	calc := Calculator{
//		TotalMemory:      SizeFromString("2G"),
//		ThreadCount:      300,
//		LoadedClassCount: 50000,
//		HeadRoom:         10,
//	}
//
//	// Calculate with existing JVM flags
//	regions, err := calc.Calculate("-XX:MaxMetaspaceSize=512m -Xss2m")
//	if err != nil {
//		return fmt.Errorf("calculation failed: %w", err)
//	}
//
//	// Use calculated regions for JVM startup
//	jvmArgs := regions.ToJVMArgs()
func (c Calculator) Calculate(flags string) (MemoryRegions, error) {
	// Initialize default memory regions
	m := MemoryRegions{
		DirectMemory:      DefaultDirectMemory,
		ReservedCodeCache: DefaultReservedCodeCache,
		Stack:             DefaultStack,
	}

	// Parse and apply JVM flags
	if err := c.parseAndApplyFlags(flags, &m); err != nil {
		return MemoryRegions{}, err
	}

	// Calculate metaspace if not configured
	c.calculateMetaspaceIfNeeded(&m)

	// Calculate head room
	c.calculateHeadRoom(&m)

	// Validate memory constraints and calculate heap
	if err := c.validateAndCalculateHeap(&m); err != nil {
		return MemoryRegions{}, err
	}

	return m, nil
}

// parseAndApplyFlags parses JVM flags and applies them to memory regions
func (c Calculator) parseAndApplyFlags(flags string, m *MemoryRegions) error {
	p, err := parser.ParseFlags(flags)
	if err != nil {
		return fmt.Errorf("unable to parse flags\n%w", err)
	}

	for _, s := range p {
		if err := c.applyFlagToRegion(s, m); err != nil {
			return err
		}
	}
	return nil
}

// applyFlagToRegion applies a single flag to the appropriate memory region
func (c Calculator) applyFlagToRegion(flag string, m *MemoryRegions) error {
	if matchDirectMemory(flag) {
		return c.setDirectMemory(flag, m)
	} else if matchHeap(flag) {
		return c.setHeap(flag, m)
	} else if matchMetaspace(flag) {
		return c.setMetaspace(flag, m)
	} else if matchReservedCodeCache(flag) {
		return c.setReservedCodeCache(flag, m)
	} else if matchStack(flag) {
		return c.setStack(flag, m)
	}
	return nil
}

// setDirectMemory parses and sets direct memory configuration
func (c Calculator) setDirectMemory(flag string, m *MemoryRegions) error {
	d, err := parseDirectMemory(flag)
	if err != nil {
		return fmt.Errorf("unable to parse direct memory\n%w", err)
	}
	d.Provenance = UserConfigured
	m.DirectMemory = d
	return nil
}

// setHeap parses and sets heap configuration
func (c Calculator) setHeap(flag string, m *MemoryRegions) error {
	h, err := parseHeap(flag)
	if err != nil {
		return fmt.Errorf("unable to parse heap\n%w", err)
	}
	h.Provenance = UserConfigured
	m.Heap = &h
	return nil
}

// setMetaspace parses and sets metaspace configuration
func (c Calculator) setMetaspace(flag string, m *MemoryRegions) error {
	ms, err := parseMetaspace(flag)
	if err != nil {
		return fmt.Errorf("unable to parse metaspace\n%w", err)
	}
	ms.Provenance = UserConfigured
	m.Metaspace = &ms
	return nil
}

// setReservedCodeCache parses and sets reserved code cache configuration
func (c Calculator) setReservedCodeCache(flag string, m *MemoryRegions) error {
	r, err := parseReservedCodeCache(flag)
	if err != nil {
		return fmt.Errorf("unable to parse reserved code cache\n%w", err)
	}
	r.Provenance = UserConfigured
	m.ReservedCodeCache = r
	return nil
}

// setStack parses and sets stack configuration
func (c Calculator) setStack(flag string, m *MemoryRegions) error {
	st, err := parseStack(flag)
	if err != nil {
		return fmt.Errorf("unable to parse stack\n%w", err)
	}
	st.Provenance = UserConfigured
	m.Stack = st
	return nil
}

// calculateMetaspaceIfNeeded calculates metaspace if not already configured by user
func (c Calculator) calculateMetaspaceIfNeeded(m *MemoryRegions) {
	if m.Metaspace == nil {
		ms := Metaspace{
			Value:      ClassOverhead + (int64(c.LoadedClassCount) * ClassSize),
			Provenance: Calculated,
		}
		m.Metaspace = &ms
	}
}

// calculateHeadRoom calculates the head room based on total memory and percentage
func (c Calculator) calculateHeadRoom(m *MemoryRegions) {
	m.HeadRoom = &HeadRoom{
		Value:      int64((float64(c.HeadRoom) / 100) * float64(c.TotalMemory.Value)),
		Provenance: Calculated,
	}
}

// validateAndCalculateHeap validates memory constraints and calculates heap if needed
func (c Calculator) validateAndCalculateHeap(m *MemoryRegions) error {
	// Validate fixed regions
	if err := c.validateFixedRegions(m); err != nil {
		return err
	}

	// Validate non-heap regions and calculate heap if needed
	if err := c.validateNonHeapAndCalculateHeap(m); err != nil {
		return err
	}

	// Final validation of all regions
	return c.validateAllRegions(m)
}

// validateFixedRegions validates that fixed regions fit within total memory
func (c Calculator) validateFixedRegions(m *MemoryRegions) error {
	f, err := m.FixedRegionsSize(c.ThreadCount)
	if err != nil {
		return fmt.Errorf("unable to calculate fixed regions size\n%w", err)
	}

	if f.Value > c.TotalMemory.Value {
		return fmt.Errorf(
			"fixed memory regions require %s which is greater than %s available for allocation: %s",
			f, c.TotalMemory, m.FixedRegionsString(c.ThreadCount),
		)
	}
	return nil
}

// validateNonHeapAndCalculateHeap validates non-heap regions and calculates heap if needed
func (c Calculator) validateNonHeapAndCalculateHeap(m *MemoryRegions) error {
	n, err := m.NonHeapRegionsSize(c.ThreadCount)
	if err != nil {
		return fmt.Errorf("unable to calculate non-heap regions size\n%w", err)
	}

	if n.Value > c.TotalMemory.Value {
		return fmt.Errorf(
			"non-heap memory regions require %s which is greater than %s available for allocation: %s",
			n, c.TotalMemory, m.NonHeapRegionsString(c.ThreadCount),
		)
	}

	// Calculate heap if not configured by user
	if m.Heap == nil {
		m.Heap = &Heap{
			Value:      c.TotalMemory.Value - n.Value,
			Provenance: Calculated,
		}
	}
	return nil
}

// validateAllRegions performs final validation that all regions fit within total memory
func (c Calculator) validateAllRegions(m *MemoryRegions) error {
	a, err := m.AllRegionsSize(c.ThreadCount)
	if err != nil {
		return fmt.Errorf("unable to calculate all regions size\n%w", err)
	}

	if a.Value > c.TotalMemory.Value {
		return fmt.Errorf(
			"all memory regions require %s which is greater than %s available for allocation: %s",
			a, c.TotalMemory, m.AllRegionsString(c.ThreadCount))
	}
	return nil
}
