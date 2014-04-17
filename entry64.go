package hamt_go

type Entry64I interface {
	GetIndex() byte
	Node64I
}

type Entry64 struct {
	ndx  byte
	Node Node64I
}

func NewEntry64(ndx byte, node Node64I) (e *Entry64, err error) {
	// XXX ndx must be within range

	if err == nil {
		e = &Entry64{
			ndx:  ndx,
			Node: node,
		}
	}
	return
}

func (e64 *Entry64) GetIndex() byte {
	return e64.ndx
}
