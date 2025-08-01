package calc

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	DefaultDirectMemory = DirectMemory{Value: 10 * Mebi, Provenance: Default}
	DirectMemoryRE      = regexp.MustCompile(fmt.Sprintf("^-XX:MaxDirectMemorySize=(%s)$", SizePattern))
)

type DirectMemory Size

func (d DirectMemory) String() string {
	return fmt.Sprintf("-XX:MaxDirectMemorySize=%s", Size(d))
}

func MatchDirectMemory(s string) bool {
	return DirectMemoryRE.MatchString(strings.TrimSpace(s))
}

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
