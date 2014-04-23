package hamt_go

// hamt_go/hamt.go

// global variables with default values
// XXX DROP THESE
var (
	W          = uint(5)
	MAX_DEPTH  = uint(64 / W)
	LEVEL_MASK = uint64(0x1f) // masks off W bits
)

type HAMT struct {
	root *Root // could be EntryI
}

func NewHAMT(w, t uint) (h *HAMT) {
	root := NewRoot(w, t)
	h = &HAMT{
		root: root,
	}
	return
}

func (h *HAMT) Delete(k KeyI) (err error) {

	var hc uint64
	hc, err = k.Hashcode()
	if err == nil {
		err = h.root.deleteEntry(hc, k)
	}
	return
}

func (h *HAMT) Find(k KeyI) (v interface{}, err error) {

	var hc uint64
	hc, err = k.Hashcode()
	if err == nil {
		v, err = h.root.findEntry(hc, k)
	}
	return
}

func (h *HAMT) Insert(k KeyI, v interface{}) (err error) {

	hc, err := k.Hashcode()
	if err == nil {
		ndx := hc & h.root.mask
		var leaf *Leaf
		leaf, err = NewLeaf(k, v)
		if err == nil {
			var e *Entry
			e, err = NewEntry(byte(ndx), leaf)
			if err == nil {
				_, err = h.root.insertEntry(hc, e)
			}
		}
	}
	return
}
