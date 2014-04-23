package hamt_go

// hamt_go/root.go

import (
	"bytes"
	"fmt"
)

var _ = fmt.Print

type Root struct {
	w        uint // non-root tables have 2^w slots
	t        uint // root table has 2^t slots
	maxSlots uint // maximum slots for the root table
	mask     uint64
	slots    []*Entry // each nil or a pointer to either a leaf or a table
}

func NewRoot(w, t uint) (root *Root) {
	flag := uint64(1)
	flag <<= t
	slocCount := powerOfTwo(t)
	root = &Root{
		w:        w,
		t:        t,
		mask:     flag - 1,
		slots:    make([]*Entry, slocCount),
		maxSlots: slocCount,
	}
	return
}

// Return a count of leaf nodes in the root
func (root *Root) GetLeafCount() (count uint) {
	for i := 0; i < len(root.slots); i++ {
		node := root.slots[i].Node
		if node != nil && node.IsLeaf() {
			count++
		}
	}
	return
}

//
func (root *Root) GetTableCount() (count uint) {
	count = 1 // we include the root in the count
	for i := 0; i < len(root.slots); i++ {
		node := root.slots[i].Node
		if node != nil && !node.IsLeaf() {
			tDeeper := node.(*Table)
			count += tDeeper.GetTableCount()
		}
	}
	return
}

func (root *Root) deleteEntry(hc uint64, key KeyI) (err error) {

	ndx := hc & root.mask
	if root.slots[ndx] == nil {
		err = NotFound
	}
	if err == nil {
		// the entry is present
		entry := root.slots[ndx]
		if entry.Node.IsLeaf() {
			// KEYS MUST BE OF THE SAME TYPE
			myLeaf := entry.Node.(*Leaf)
			myKey := myLeaf.Key.(*BytesKey)
			searchKey := key.(*BytesKey)
			if bytes.Equal(searchKey.Slice, myKey.Slice) {
				root.slots[ndx] = nil
			} else {
				err = NotFound
			}
		} else {
			// entry is a table, so recurse
			tDeeper := entry.Node.(*Table)
			hc >>= root.t
			err = tDeeper.deleteEntry(hc, 1, key)
		}
	}
	return
}

func (root *Root) findEntry(hc uint64, key KeyI) (
	value interface{}, err error) {

	ndx := hc & root.mask
	if root.slots[ndx] == nil {
		err = NotFound
	}
	if err == nil {
		// the entry is present
		entry := root.slots[ndx]
		// XXX this MUST exist
		if entry.Node.IsLeaf() {
			myLeaf := entry.Node.(*Leaf)
			myKey := myLeaf.Key.(*BytesKey)
			searchKey := key.(*BytesKey)
			if bytes.Equal(searchKey.Slice, myKey.Slice) {
				value = myLeaf.Value
			} else {
				err = NotFound
			}
		} else {
			// entry is a table, so recurse
			tDeeper := entry.Node.(*Table)
			hc >>= root.t
			value, err = tDeeper.findEntry(hc, 1, key)
		}
	}
	return
}

// XXX MERGE THESE
func (root *Root) insertEntry(hc uint64, entry *Entry) (
	slotNbr uint, err error) {

	ndx := hc & root.mask

	// DEBUG
	// fmt.Printf("insert root slot %4d (0x%03x)", ndx, ndx)
	// END
	if root.slots[ndx] == nil {
		root.slots[ndx] = entry
	} else {
		// WORKING HERE
		// DEBUG
		// fmt.Printf(" already occupied")
		// END

		// there is already something in this slot
		err = root.insertIntoOccupiedSlot(hc, entry, ndx)
	}
	// DEBUG
	// fmt.Println()
	// END
	return
}

// Insert a new entry into a slot which is already occupied.
//
func (root *Root) insertIntoOccupiedSlot(newHC uint64,
	entry *Entry, ndx uint64) (err error) {

	e := root.slots[ndx]

	if e.Node.IsLeaf() {
		// if it's a leaf, we replace the value iff the keys match

		//////////////////////////////////////
		// XXX NOT CHECKING FOR DUPLICATE KEYS
		//////////////////////////////////////

		var (
			tableDeeper *Table
			oldEntry    *Entry
			oldHC       uint64
		)

		tableDeeper, err = NewTable(1, root.w, root.t)
		if err == nil {
			newHC >>= root.t // this is hc for the NEW entry

			oldEntry = e
			oldLeaf := e.Node.(*Leaf)
			oldHC, err = oldLeaf.Key.Hashcode()
		}
		if err == nil {
			oldHC >>= root.t

			// put the existing leaf into the new table
			_, err = tableDeeper.insertEntry(oldHC, 1, oldEntry)
			if err == nil {
				// then put the new entry in the new table
				_, err = tableDeeper.insertEntry(newHC, 1, entry)
				if err == nil {
					// the new table replaces the existing leaf
					var eTab *Entry
					eTab, err = NewEntry(byte(ndx), tableDeeper)
					if err == nil {
						root.slots[ndx] = eTab
					}
				}
			}
		}
	} else {
		// otherwise it's a table, so recurse
		tDeeper := e.Node.(*Table)
		newHC >>= root.t
		_, err = tDeeper.insertEntry(newHC, 1, entry)

	}
	return
}

// SUPERFLUOUS??
func (root *Root) IsLeaf() bool {
	return false
}
