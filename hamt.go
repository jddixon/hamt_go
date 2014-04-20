package hamt_go

// hamt_go/hamt.go

type HAMT struct {
	root *Table // could be EntryI
}

func NewHAMT() (h *HAMT) {
	table, _ := NewTable(0)
	h = &HAMT{
		root: table,
	}
	return
}

func (h *HAMT) Delete(k KeyI) (err error) {

	var hc uint64
	hc, err = k.Hashcode()
	if err == nil {
		// this is depth zero, so hc is not shifted
		err = h.root.deleteEntry(hc, 0, k)
	}
	return
}

func (h *HAMT) Find(k KeyI) (v interface{}, err error) {

	var hc uint64
	hc, err = k.Hashcode()
	if err == nil {
		// this is depth zero, so hc is not shifted
		v, err = h.root.findEntry(hc, 0, k)
	}
	return
}

func (h *HAMT) Insert(k KeyI, v interface{}) (err error) {

	hc, err := k.Hashcode()
	if err == nil {
		ndx := byte(hc & LEVEL_MASK)
		var leaf *Leaf
		leaf, err = NewLeaf(k, v)
		if err == nil {
			var e *Entry
			e, err = NewEntry(ndx, leaf)
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
