package hamt_go

// hamt_go/root.go

import (
	"bytes"
	"fmt"
)

var _ = fmt.Print

type Root struct {
	w         uint // non-root tables have 2^w slots
	t         uint // root table has 2^t slots
	slotCount uint // maximum slots for the root table
	mask      uint64
	slots     []*Entry // each nil or a pointer to either a leaf or a table
}

func NewRoot(w, t uint) (root *Root) {
	flag := uint64(1)
	flag <<= t
	count := uint(1 << t) // number of slots
	root = &Root{
		w:         w,
		t:         t,
		mask:      flag - 1,
		slots:     make([]*Entry, count),
		slotCount: count,
	}
	return
}

// Return a count of leaf nodes in the root
func (root *Root) getLeafCount() (count uint) {
	if root.slots != nil {
		for i := uint(0); i < root.slotCount; i++ {
			if root.slots[i] != nil {
				node := root.slots[i].Node
				if node != nil {
					if node.IsLeaf() {
						count++
					} else {
						// recurse
						table := node.(*Table)
						count += table.getLeafCount()
					}
				}
			}
		}
	}
	return
}

// Return a count of tables (including the root) in the HAMT
func (root *Root) getTableCount() (count uint) {
	count = 1 // we include the root in the count
	if root.slots != nil {
		for i := uint(0); i < root.slotCount; i++ {
			if root.slots[i] != nil {
				node := root.slots[i].Node
				if node != nil && !node.IsLeaf() {
					tDeeper := node.(*Table)
					count += tDeeper.getTableCount()
				}
			}
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

func (root *Root) insertEntry(newHC uint64, entry *Entry) (
	slotNbr uint, err error) {

	ndx := newHC & root.mask

	if root.slots[ndx] == nil {
		root.slots[ndx] = entry
	} else {
		// there is already something in this slot
		e := root.slots[ndx]

		if e.Node.IsLeaf() {
			// if it's a leaf, we replace the value iff the keys match
			curLeaf := e.Node.(*Leaf)
			curKey := curLeaf.Key.(*BytesKey)
			entryAsLeaf := entry.Node.(*Leaf)
			newKey := entryAsLeaf.Key.(*BytesKey)
			if bytes.Equal(curKey.Slice, newKey.Slice) {
				// the keys match, so we replace the value
				newLeaf := entry.Node.(*Leaf)
				curLeaf.Value = newLeaf.Value
			} else {
				// keys differ, so we need to replace the leaf with a table
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
			}
		} else {
			// otherwise it's a table, so recurse
			tDeeper := e.Node.(*Table)
			newHC >>= root.t
			_, err = tDeeper.insertEntry(newHC, 1, entry)

		}
	}
	return
}

// SUPERFLUOUS??
func (root *Root) IsLeaf() bool {
	return false
}
