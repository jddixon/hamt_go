package hamt_go

// hamt_go/hamt.go

type HAMT struct {
	root *Root // could be EntryI
}

// Create a new HAMT with 2^t slots in its root table and 2^w slots in
// all lower-level tables.  If t equals zero, it defaults to w.  If
// both t and w are zero, it panics.
func NewHAMT(w, t uint) (h *HAMT) {
	if t == 0 && w == 0 {
		panic("cannot create HAMT with no slots in tables")
	}
	if t == 0 {
		t = w
	}
	root := NewRoot(w, t)
	h = &HAMT{
		root: root,
	}
	return
}

// Return t which determines the size of the root table (2^t).
func (h *HAMT) GetT() uint {
	return h.root.t
}

// Return w which determines the size of lower-level tables (2^w).
func (h *HAMT) GetW() uint {
	return h.root.w
}

// Return the number of leaf nodes in the HAMT.
func (h *HAMT) GetLeafCount() uint {
	return h.root.getLeafCount()
}

// Return the number of tables, including the root table, in the HAMT.
func (h *HAMT) GetTableCount() uint {
	return h.root.getTableCount()
}

// If there is an entry with the key k in the HAMT, remove it.  If
// there is no such entry, return NotFound.
func (h *HAMT) Delete(k KeyI) (err error) {

	var hc uint64
	hc, err = k.Hashcode()
	if err == nil {
		err = h.root.deleteEntry(hc, k)
	}
	return
}

// If there is an entry with the key k in the HAMT, return the value
// associated with the key.  If there is no such entry, return NotFound.
func (h *HAMT) Find(k KeyI) (v interface{}, err error) {

	var hc uint64
	hc, err = k.Hashcode()
	if err == nil {
		v, err = h.root.findEntry(hc, k)
	}
	return
}

// Store the value v in the table keyed by k.  If there is already
// an entry in the trie with this key, then replace the associated
// value.  This is not an error condition.
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
