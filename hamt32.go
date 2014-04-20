package hamt_go

// hamt_go/hamt32.go

type HAMT32 struct {
	root *Table32 // could be Entry32I
}

func NewHAMT32() (h *HAMT32) {
	t32, _ := NewTable32(0)
	h = &HAMT32{
		root: t32,
	}
	return
}

func (h *HAMT32) Delete(k Key64I) (err error) {

	var hc uint64
	hc, err = k.Hashcode64()
	if err == nil {
		// this is depth zero, so hc is not shifted
		err = h.root.deleteEntry(hc, 0, k)
	}
	return
}

func (h *HAMT32) Find(k Key64I) (v interface{}, err error) {

	var hc uint64
	hc, err = k.Hashcode64()
	if err == nil {
		// this is depth zero, so hc is not shifted
		v, err = h.root.findEntry(hc, 0, k)
	}
	return
}

func (h *HAMT32) Insert(k Key64I, v interface{}) (err error) {

	hc, err := k.Hashcode64()
	if err == nil {
		ndx := byte(hc & LEVEL_MASK32)
		var leaf *Leaf64
		leaf, err = NewLeaf64(k, v)
		if err == nil {
			var e *Entry32
			e, err = NewEntry32(ndx, leaf)
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
