package calc

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// DefaultDirectMemory is the default direct memory size (10MB).
	DefaultDirectMemory = DirectMemory{Value: 10 * Mebi, Provenance: Default}
	// DirectMemoryRE is the regular expression for matching direct memory flags.
	DirectMemoryRE = regexp.MustCompile(fmt.Sprintf("^-XX:MaxDirectMemorySize=(%s)$", SizePattern))
)

// DirectMemory represents the maximum direct memory size.
type DirectMemory Size

func (d DirectMemory) String() string {
	return fmt.Sprintf("-XX:MaxDirectMemorySize=%s", Size(d))
}

// MatchDirectMemory returns true if the string matches the direct memory flag pattern.
func MatchDirectMemory(s string) bool {
	return DirectMemoryRE.MatchString(strings.TrimSpace(s))
}

// ParseDirectMemory parses a string into a DirectMemory object.
func ParseDirectMemory(s string) (DirectMemory, error) {
	g := DirectMemoryRE.FindStringSubmatch(s)
	if g == nil {
		return DirectMemory{}, fmt.Errorf("%s does not match direct memory pattern %s", s, DirectMemoryRE.String())
	}

	z, err := ParseSize(g[1])
	if err != nil {
		return DirectMemory{}, fmt.Errorf("unable to parse direct memory size\n%w", err)
	}

	return DirectMemory(z), nil
}
