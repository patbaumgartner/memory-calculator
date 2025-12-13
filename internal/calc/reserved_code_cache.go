package calc

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// DefaultReservedCodeCache is the default reserved code cache size (240MB).
	DefaultReservedCodeCache = ReservedCodeCache{Value: 240 * Mebi, Provenance: Default}
	// ReservedCodeCacheRE is the regular expression for matching reserved code cache flags.
	ReservedCodeCacheRE = regexp.MustCompile(fmt.Sprintf("^-XX:ReservedCodeCacheSize=(%s)$", SizePattern))
)

// ReservedCodeCache represents the reserved code cache memory size.
type ReservedCodeCache Size

func (r ReservedCodeCache) String() string {
	return fmt.Sprintf("-XX:ReservedCodeCacheSize=%s", Size(r))
}

// MatchReservedCodeCache returns true if the string matches the reserved code cache flag pattern.
func MatchReservedCodeCache(s string) bool {
	return ReservedCodeCacheRE.MatchString(strings.TrimSpace(s))
}

// ParseReservedCodeCache parses a string into a ReservedCodeCache object.
func ParseReservedCodeCache(s string) (ReservedCodeCache, error) {
	g := ReservedCodeCacheRE.FindStringSubmatch(s)
	if g == nil {
		return ReservedCodeCache{}, fmt.Errorf(
			"%s does not match reserved code cache pattern %s", s, ReservedCodeCacheRE.String())
	}

	z, err := ParseSize(g[1])
	if err != nil {
		return ReservedCodeCache{}, fmt.Errorf("unable to parse reserved code cache size\n%w", err)
	}

	return ReservedCodeCache(z), nil
}
