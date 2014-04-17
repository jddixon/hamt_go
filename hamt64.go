package hamt_go

// hamt_go/hamt64.go

type HAMT64 struct {
	root *Table64 // could be Entry64I
}

func NewHAMT64() (h *HAMT64) {
	t64, _ := NewTable64(0)
	h = &HAMT64{
		root: t64,
	}
	return
}

func (h *HAMT64) Delete(k Key64I) (err error) {

	var hc uint64
	hc, err = k.Hashcode64()
	if err == nil {
		// this is depth zero, so hc is not shifted
		err = h.root.deleteEntry(hc, 0, k)
	}
	return
}

func (h *HAMT64) Find(k Key64I) (v interface{}, err error) {

	var hc uint64
	hc, err = k.Hashcode64()
	if err == nil {
		// this is depth zero, so hc is not shifted
		v, err = h.root.findEntry(hc, 0, k)
	}
	return
}

func (h *HAMT64) Insert(k Key64I, v interface{}) (err error) {

	hc, err := k.Hashcode64()
	if err == nil {
		ndx := byte(hc & LEVEL_MASK64)
		var leaf *Leaf64
		leaf, err = NewLeaf64(k, v)
		if err == nil {
			var e *Entry64
			e, err = NewEntry64(ndx, leaf)
			if err == nil {
				var slotNbr uint
				// depth is 0, so hc unshifted
				slotNbr, err = h.root.insertEntry(hc, 0, e)
				_ = slotNbr
			}
		}
	}
	return
}
