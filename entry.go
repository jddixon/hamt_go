package hamt_go

type Entry32I interface {
	GetIndex() byte
	Node32I
}

type Entry32 struct {
	ndx byte
	Node32I
}

func NewEntry32(ndx byte, node Node32I) (e Entry32I, err error) {
	// XXX ndx must be within range

	if err == nil {
		e = &Entry32{
			ndx:     ndx,
			Node32I: node, // is this possible??
		}
	}
	return
}

func (e32 *Entry32) GetIndex() byte {
	return e32.ndx
}

// ==================================================================

type Entry64I interface {
	GetIndex() byte
	IsLeaf() bool
}
