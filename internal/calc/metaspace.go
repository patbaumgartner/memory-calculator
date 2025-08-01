package calc

import (
	"fmt"
	"regexp"
	"strings"
)

var MetaspaceRE = regexp.MustCompile(fmt.Sprintf("^-XX:MaxMetaspaceSize=(%s)$", SizePattern))

type Metaspace Size

func (m Metaspace) String() string {
	return fmt.Sprintf("-XX:MaxMetaspaceSize=%s", Size(m))
}

func MatchMetaspace(s string) bool {
	return MetaspaceRE.MatchString(strings.TrimSpace(s))
}

func ParseMetaspace(s string) (*Metaspace, error) {
	g := MetaspaceRE.FindStringSubmatch(s)
	if g == nil {
		return nil, fmt.Errorf("%s does not match metaspace pattern %s", s, MetaspaceRE.String())
	}

	z, err := ParseSize(g[1])
	if err != nil {
		return nil, fmt.Errorf("unable to parse metaspace size\n%w", err)
	}

	m := Metaspace(z)
	return &m, nil
}
