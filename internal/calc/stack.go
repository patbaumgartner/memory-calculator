package calc

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// DefaultStack is the default stack size (1MB).
	DefaultStack = Stack{Value: Mebi, Provenance: Default}
	// StackRE is the regular expression for matching stack size flags.
	StackRE = regexp.MustCompile(fmt.Sprintf("^-Xss(%s)$", SizePattern))
)

// Stack represents the thread stack size.
type Stack Size

func (s Stack) String() string {
	return fmt.Sprintf("-Xss%s", Size(s))
}

// MatchStack returns true if the string matches the stack size flag pattern.
func MatchStack(s string) bool {
	return StackRE.MatchString(strings.TrimSpace(s))
}

// ParseStack parses a string into a Stack object.
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
