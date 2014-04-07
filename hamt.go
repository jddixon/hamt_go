package hamt_go

type HAMT32 struct {
	root *Table32 // could be Entry32I
}

func NewHAMT32() (h *HAMT32) {
	h = &HAMT32{
		root: NewTable32(0),
	}
	return
}

func (h *HAMT32) Delete(k Key32I) (err error) {

	// XXX STUB

	return
}

func (h *HAMT32) Find(k Key32I) (v interface{}, err error) {

	// XXX STUB

	return
}

func (h *HAMT32) Insert(k Key32I, v interface{}) (err error) {

	hc, err := k.Hashcode32()
	if err == nil {
		err = h.root.Insert(hc, 0, k, v) // depth 0, so hc unshifted
	}
	return
} // FOO

// ==================================================================

type HAMT64 struct {
	root Entry64I
}
