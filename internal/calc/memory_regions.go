package calc

import (
	"fmt"
	"strings"
)

type MemoryRegions struct {
	DirectMemory      DirectMemory
	HeadRoom          *HeadRoom
	Heap              *Heap
	Metaspace         *Metaspace
	ReservedCodeCache ReservedCodeCache
	Stack             Stack
}

func (m MemoryRegions) FixedRegionsSize(threadCount int) (Size, error) {
	if m.Metaspace == nil {
		return Size{}, fmt.Errorf("unable to calculate fixed regions size without metaspace")
	}

	return Size{
		Value: m.DirectMemory.Value + m.Metaspace.Value + m.ReservedCodeCache.Value +
			(m.Stack.Value * int64(threadCount)),
		Provenance: Calculated,
	}, nil
}

func (m MemoryRegions) FixedRegionsString(threadCount int) string {
	var s []string

	s = append(s, m.DirectMemory.String())
	if m.Metaspace != nil {
		s = append(s, m.Metaspace.String())
	}
	s = append(s, m.ReservedCodeCache.String())
	s = append(s, fmt.Sprintf("%s * %d threads", m.Stack.String(), threadCount))

	return strings.Join(s, ", ")
}

func (m MemoryRegions) NonHeapRegionsSize(threadCount int) (Size, error) {
	if m.HeadRoom == nil {
		return Size{}, fmt.Errorf("unable to calculate non-heap regions size without headroom")
	}

	s, err := m.FixedRegionsSize(threadCount)
	if err != nil {
		return Size{}, fmt.Errorf("unable to calculate fixed regions size\n%w", err)
	}

	return Size{
		Value:      m.HeadRoom.Value + s.Value,
		Provenance: Calculated,
	}, nil
}

func (m MemoryRegions) NonHeapRegionsString(threadCount int) string {
	var s []string

	if m.HeadRoom != nil {
		s = append(s, fmt.Sprintf("%s headroom", m.HeadRoom.String()))
	}
	s = append(s, m.FixedRegionsString(threadCount))

	return strings.Join(s, ", ")
}

func (m MemoryRegions) AllRegionsSize(threadCount int) (Size, error) {
	if m.Heap == nil {
		return Size{}, fmt.Errorf("unable to calculate all regions size without heap")
	}

	s, err := m.NonHeapRegionsSize(threadCount)
	if err != nil {
		return Size{}, fmt.Errorf("unable to calculate non-heap regions size\n%w", err)
	}

	return Size{
		Value:      s.Value + m.Heap.Value,
		Provenance: Calculated,
	}, nil
}

func (m MemoryRegions) AllRegionsString(threadCount int) string {
	var s []string

	if m.Heap != nil {
		s = append(s, m.Heap.String())
	}
	s = append(s, m.NonHeapRegionsString(threadCount))

	return strings.Join(s, ", ")
}
