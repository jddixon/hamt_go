package hamt_go

// hamt_go/table.go

import (
	"fmt"
)

var _ = fmt.Print

type Table32 struct {
	bitmap uint32
	slots  []Entry32I // each nil or a pointer to either a leaf or a table
}

func NewTable32() (t32 *Table32) {
	t32 = new(Table32)
	return
}

// hc is the hashcode shifted the number of bits appropriate for the
// depth.  If it is necessary to recurse, the depth is increased by
func (t32 *Table32) Insert(hc uint32, depth uint, k Key32I, v interface{}) (
	err error) {

	// Mask off the first level mask bits, and see what we have there
	where := hc & LEVEL_MASK32

	p := t32.slots[where]
	if p == nil {
		// empty slot, so put the entry there
		var leaf *Leaf32
		leaf, err = NewLeaf32(k, v)
		if err == nil {
			t32.slots[where] = leaf
		}
	} else if p.IsLeaf() {

		// XXX STUB
		fmt.Printf("should be replacing leaf at slot %d\n", where)

	} else {
		// it's not a leaf, so try to insert
		nextT := p.(*Table32)
		hc <<= W32 // shift appropriate number of bits
		err = nextT.Insert(hc, depth+1, k, v)

	}

	// XXX STUB
	return
}

func (t32 *Table32) IsLeaf() bool { return false }

// ==================================================================

type Table64 struct {
	bitmap uint64
	slots  []Entry64I // nil or a key-value pair or a pointer to a table
}

func (t64 *Table64) IsLeaf() bool { return false }
