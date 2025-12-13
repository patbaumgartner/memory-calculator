package calc

// HeadRoom represents the memory headroom.
type HeadRoom Size

func (h HeadRoom) String() string {
	return Size(h).String()
}
