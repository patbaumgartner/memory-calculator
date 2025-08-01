package calc

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	DefaultStack = Stack{Value: Mebi, Provenance: Default}
	StackRE      = regexp.MustCompile(fmt.Sprintf("^-Xss(%s)$", SizePattern))
)

type Stack Size

func (s Stack) String() string {
	return fmt.Sprintf("-Xss%s", Size(s))
}

func MatchStack(s string) bool {
	return StackRE.MatchString(strings.TrimSpace(s))
}

func ParseStack(s string) (Stack, error) {
	g := StackRE.FindStringSubmatch(s)
	if g == nil {
		return Stack{}, fmt.Errorf("%s does not match stack pattern %s", s, StackRE.String())
	}

	z, err := ParseSize(g[1])
	if err != nil {
		return Stack{}, fmt.Errorf("unable to parse stack size\n%w", err)
	}

	return Stack(z), nil
}
