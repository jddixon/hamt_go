package hamt_go

// hamt_go/hamt.go
import (
	"fmt"
)

var _ = fmt.Print

type HAMT struct {
	root *Root
}

// Create a new HAMT with 2^t slots in its root table and 2^w slots in
// all lower-level tables.  If t equals zero, it defaults to w.  If
// both t and w are zero, it panics.  In lower-level tables, a uint64
// is used as a bitmap, so w may not exceed 6 (because 2^6 == 64).
func NewHAMT(w, t uint) (h *HAMT, err error) {
	if t == 0 && w == 0 {
		err = ZeroLengthTables
	} else {
		if w > 6 {
			err = MaxTableSizeExceeded
		} else {
			if t == 0 {
				t = w
			}
			root := NewRoot(w, t)
			h = &HAMT{
				root: root,
			}
		}
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
func (h *HAMT) Delete(k KeyI) error {
	return h.root.deleteLeaf(k)
}

// If there is an entry with the key k in the HAMT, return the value
// associated with the key.  If there is no such entry, return nil.
func (h *HAMT) Find(k KeyI) (interface{}, error) {
	return h.root.findLeaf(k)
}

// Try to create an Leaf for the key/value pair..  If this succeeds,
// try to insert the Leaf into the root table.
func (h *HAMT) Insert(k KeyI, v interface{}) (err error) {
	leaf, err := NewLeaf(k, v)
	if err == nil {
		_, err = h.root.insertLeaf(leaf)
	}
	return
}
