package calc

type HeadRoom Size

func (h HeadRoom) String() string {
	return Size(h).String()
}
