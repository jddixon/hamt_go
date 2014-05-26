package hamt_go

// hamt_go/root.go

import (
	"bytes"
	"fmt"
)

var _ = fmt.Print

type Root struct {
	w             uint // non-root tables have 2^w slots
	t             uint // root table has 2^t slots
	maxTableDepth uint // max depth of descendent Tables
	slotCount     uint // number of slots in the root table
	mask          uint64
	slots         []HTNodeI // each nil or a pointer to either a leaf or a table
}

func NewRoot(w, t uint) (root *Root, err error) {
	if w > 6 {
		err = MaxTableSizeExceeded
	} else if t > 64 { // very generous!
		err = MaxRootTableSizeExceeded
	} else {
		flag := uint64(1)
		flag <<= t
		count := uint(1 << t) // number of slots
		root = &Root{
			w: w,
			t: t,
			// The maximum possible depth for any table below the root, (64 - t)/w.
			// There are 64 bits available for keys, the root table uses t, each
			// successive Table uses w more bits.  The root table (of type Root)
			// is at depth 0;  all Tables at at depth >= 1.
			maxTableDepth: (64 - t) / w,
			slotCount:     count,
			mask:          flag - 1,
			slots:         make([]HTNodeI, count),
		}
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

	hc := key.Hashcode()
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
			if 1 > root.maxTableDepth {
				err = NotFound
			} else {
				tDeeper := node.(*Table)
				hc >>= root.t
				err = tDeeper.deleteLeaf(hc, 1, key)
			}
		}
	}
	return
}

// Given a properly shifted hashCode and the full key for any entry,
// return the value associated with the key, nil if there is no such
// value, or any error encountered.
func (root *Root) findLeaf(key KeyI) (value interface{}, err error) {

	hc := key.Hashcode()
	ndx := hc & root.mask
	p := &root.slots
	if (*p)[ndx] != nil {
		// the entry is present
		node := (*p)[ndx]
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
			if 1 <= root.maxTableDepth {
				// entry is a table, so recurse
				tDeeper := node.(*Table)
				hc >>= root.t
				value, err = tDeeper.findLeaf(hc, 1, key)
			}
		}
	}
	return
}

func (root *Root) insertLeaf(leaf *Leaf) (err error) {

	newHC := leaf.Key.Hashcode()
	slotNbr := uint(newHC & root.mask)

	p := &root.slots
	if (*p)[slotNbr] == nil {
		(*p)[slotNbr] = leaf
	} else {
		// there is already something in this slot
		node := (*p)[slotNbr]
		if node.IsLeaf() {
			// if it's a leaf, we replace the value iff the keys match
			oldLeaf := node.(*Leaf)
			curKey := oldLeaf.Key.(*BytesKey)
			newKey := leaf.Key.(*BytesKey)
			if bytes.Equal(curKey.Slice, newKey.Slice) {
				// the keys match, so we replace the value
				oldLeaf.Value = leaf.Value
			} else {
				// keys differ, so we need to replace the leaf with a table
				// Create a new Table containing the existing leaf
				var tableDeeper *Table
				tableDeeper, err = NewTableWithLeaf(1, root, oldLeaf)
				if err == nil {
					if 1 > root.maxTableDepth {
						err = MaxTableDepthExceeded
					} else {
						newHC >>= root.t // this is hc for the NEW entry
						// then put the new entry in the new table
						err = tableDeeper.insertLeaf(newHC, 1, leaf)
						if err == nil {
							// the new table replaces the existing leaf
							(*p)[slotNbr] = tableDeeper
						}
					}
				}
			}
		} else {
			if 1 > root.maxTableDepth {
				err = MaxTableDepthExceeded
			} else {
				// otherwise it's a table, so recurse
				tDeeper := node.(*Table)
				newHC >>= root.t
				err = tDeeper.insertLeaf(newHC, 1, leaf)
			}
		}
	}
	return
}

// SUPERFLUOUS??
//func (root *Root) IsLeaf() bool {
//	return false
//}
