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
	slotCount uint // number of slots in the root table
	mask      uint64
	slots     []HTNodeI // each nil or a pointer to either a leaf or a table
}

func NewRoot(w, t uint) (root *Root) {
	flag := uint64(1)
	flag <<= t
	count := uint(1 << t) // number of slots
	root = &Root{
		w:         w,
		t:         t,
		mask:      flag - 1,
		slots:     make([]HTNodeI, count),
		slotCount: count,
	}
	return
}

// Return a count of leaf nodes in the root
func (root *Root) getLeafCount() (count uint) {
	if root.slots != nil {
		for i := uint(0); i < root.slotCount; i++ {
			if root.slots[i] != nil {
				node := root.slots[i]
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
				node := root.slots[i]
				if node != nil && !node.IsLeaf() {
					tDeeper := node.(*Table)
					count += tDeeper.getTableCount()
				}
			}
		}
	}
	return
}

func (root *Root) deleteLeaf(key KeyI) (err error) {

	hc, err := key.Hashcode()
	if err == nil {
		ndx := hc & root.mask
		if root.slots[ndx] == nil {
			err = NotFound
		}
		if err == nil {
			// the entry is present
			node := root.slots[ndx]
			if node.IsLeaf() {
				// KEYS MUST BE OF THE SAME TYPE
				myLeaf := node.(*Leaf)
				myKey := myLeaf.Key.(*BytesKey)
				searchKey := key.(*BytesKey)
				if bytes.Equal(searchKey.Slice, myKey.Slice) {
					root.slots[ndx] = nil
				} else {
					err = NotFound
				}
			} else {
				// entry is a table, so recurse
				tDeeper := node.(*Table)
				hc >>= root.t
				err = tDeeper.deleteEntry(hc, 1, key)
			}
		}
	}
	return
}

// Given a properly shifted hashCode and the full key for any entry,
// return the value associated with the key, nil if there is no such
// value, or any error encountered.
func (root *Root) findLeaf(key KeyI) (value interface{}, err error) {

	hc, err := key.Hashcode()
	if err == nil {
		ndx := hc & root.mask
		if root.slots[ndx] != nil {
			// the entry is present
			node := root.slots[ndx]
			if node.IsLeaf() {
				myLeaf := node.(*Leaf)
				myKey := myLeaf.Key.(*BytesKey)
				searchKey := key.(*BytesKey)
				if bytes.Equal(searchKey.Slice, myKey.Slice) {
					value = myLeaf.Value
				} else {
					value = nil
				}
			} else {
				// entry is a table, so recurse
				tDeeper := node.(*Table)
				hc >>= root.t
				value, err = tDeeper.findEntry(hc, 1, key)
			}
		}
	}
	return
}

func (root *Root) insertLeaf(leaf *Leaf) (slotNbr uint, err error) {

	newHC, err := leaf.Key.Hashcode()
	if err == nil {
		ndx := newHC & root.mask

		if root.slots[ndx] == nil {
			root.slots[ndx] = leaf
		} else {
			// there is already something in this slot
			node := root.slots[ndx]

			if node.IsLeaf() {
				// if it's a leaf, we replace the value iff the keys match
				curLeaf := node.(*Leaf)
				curKey := curLeaf.Key.(*BytesKey)
				newKey := leaf.Key.(*BytesKey)
				if bytes.Equal(curKey.Slice, newKey.Slice) {
					// the keys match, so we replace the value
					curLeaf.Value = leaf.Value
				} else {
					var newEntry *Entry

					// keys differ, so we need to replace the leaf with a table
					var (
						tableDeeper *Table
						oldEntry    *Entry
						oldHC       uint64
					)
					// XXX (byte)newHC serves no purpose
					newEntry, err = NewEntry(byte(newHC), leaf)
					if err == nil {
						tableDeeper, err = NewTable(1, root.w, root.t)
					}
					if err == nil {
						newHC >>= root.t // this is hc for the NEW entry

						oldLeaf := node.(*Leaf)
						oldHC, err = oldLeaf.Key.Hashcode()
						// XXX byte(oldHC) serves no purpose
						if err == nil {
							oldEntry, err = NewEntry(byte(oldHC), oldLeaf)
						}
					}
					if err == nil {
						oldHC >>= root.t

						// XXX THE SLOT NBRS ARE FOR DEBUGGING
						//var slotNbrOE, slotNbrNE uint

						// put the existing leaf into the new table
						_, err = tableDeeper.insertEntry(oldHC, 1, oldEntry)
						if err == nil {
							// then put the new entry in the new table
							_, err = tableDeeper.insertEntry(newHC, 1, newEntry)
							if err == nil {
								// the new table replaces the existing leaf
								root.slots[ndx] = tableDeeper
							}
						}
						// DEBUG
						//fmt.Printf("root table slot %d (0x%x): replaced entry with table, OE %d (0x%x), NE %d (0x%x)\n",
						//	ndx, ndx, slotNbrOE, slotNbrOE, slotNbrNE, slotNbrNE)
						// END
					}
				}
			} else {
				// otherwise it's a table, so recurse
				newEntry, err := NewEntry(byte(newHC), leaf)
				if err == nil {
					tDeeper := node.(*Table)
					newHC >>= root.t
					_, err = tDeeper.insertEntry(newHC, 1, newEntry)
				}

			}
		}
	}
	return
}

// SUPERFLUOUS??
//func (root *Root) IsLeaf() bool {
//	return false
//}
