package calc

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	DefaultReservedCodeCache = ReservedCodeCache{Value: 240 * Mebi, Provenance: Default}
	ReservedCodeCacheRE      = regexp.MustCompile(fmt.Sprintf("^-XX:ReservedCodeCacheSize=(%s)$", SizePattern))
)

type ReservedCodeCache Size

func (r ReservedCodeCache) String() string {
	return fmt.Sprintf("-XX:ReservedCodeCacheSize=%s", Size(r))
}

func MatchReservedCodeCache(s string) bool {
	return ReservedCodeCacheRE.MatchString(strings.TrimSpace(s))
}

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
