package calc

import (
	"fmt"
	"regexp"
	"strings"
)

// HeapRE is the regular expression for matching heap memory flags.
var HeapRE = regexp.MustCompile(fmt.Sprintf("^-Xmx(%s)$", SizePattern))

// Heap represents the heap memory size.
type Heap Size

func (h Heap) String() string {
	return fmt.Sprintf("-Xmx%s", Size(h))
}

// MatchHeap returns true if the string matches the heap memory flag pattern.
func MatchHeap(s string) bool {
	return HeapRE.MatchString(strings.TrimSpace(s))
}

// ParseHeap parses a string into a Heap object.
func ParseHeap(s string) (*Heap, error) {
	g := HeapRE.FindStringSubmatch(s)
	if g == nil {
		return nil, fmt.Errorf("%s does not match heap pattern %s", s, HeapRE.String())
	}

	z, err := ParseSize(g[1])
	if err != nil {
		return nil, fmt.Errorf("unable to parse heap size\n%w", err)
	}

	h := Heap(z)
	return &h, nil
}
