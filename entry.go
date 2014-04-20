package hamt_go

type EntryI interface {
	GetIndex() byte
	HTNodeI
}

type Entry struct {
	ndx  byte
	Node HTNodeI
}

func NewEntry(ndx byte, node HTNodeI) (e *Entry, err error) {
	// XXX ndx must be within range

	if err == nil {
		e = &Entry{
			ndx:  ndx,
			Node: node,
		}
	}
	return
}

func (e *Entry) GetIndex() byte {
	return e.ndx
}
