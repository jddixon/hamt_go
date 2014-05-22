package hamt_go

// hamt_go/table.go

import (
	"bytes"
	"errors"
	"fmt"
)

var _ = fmt.Print

// This is a non-root table; depth is guaranteed never to be zero.  We
// use a uint64 as a bitmap, with a bit being set representing the fact
// that a slot is in use, so there may not be more than 64 slots, so
// w may not exceed 6 (2^6==64).
type Table struct {
	w      uint // non-root tables have 2^w slots
	t      uint // root table has 2^t slots
	mask   uint64
	bitmap uint64
	slots  []HTNodeI // each nil or a pointer to either a leaf or a table
	root   *Root     // pointer to the fixed-size root table
}

// Debugging / sanity check
func CheckTableParam(depth uint, root *Root) (w, t uint, err error) {
	if root == nil {
		err = NilRoot
	} else {
		w = root.w
		t = root.t
		if w > 6 {
			err = MaxTableSizeExceeded
		} else if t+(depth-1)*w > 64 {
			err = MaxTableDepthExceeded
		}
	}
	return
}

func NewTable(depth uint, root *Root) (table *Table, err error) {
	w, t, err := CheckTableParam(depth, root)
	if err == nil {
		table = new(Table)
		table.w = w
		table.t = t
		table.root = root
		flag := uint64(1 << w)
		table.mask = flag - 1
	}
	return
}

// Create a new table and insert a first Leaf into it.
//
func NewTableWithLeaf(depth uint, root *Root, firstLeaf *Leaf) (
	table *Table, err error) {

	w, t, err := CheckTableParam(depth, root)
	if err == nil {
		tbl := new(Table)
		tbl.w = w
		tbl.t = t
		tbl.root = root
		wFlag := uint64(1 << w)
		tbl.mask = wFlag - 1
		shiftCount := t + (depth-1)*w
		hc := firstLeaf.Key.Hashcode() >> shiftCount
		ndx := hc & tbl.mask
		flag := uint64(1 << ndx)
		tbl.slots = []HTNodeI{firstLeaf}
		tbl.bitmap = flag
		if err == nil {
			table = tbl
		}
	}
	return
}

func (table *Table) GetRoot() *Root {
	return table.root
}

// Return the maximum number of slots in the table, 2^w.
func (table *Table) MaxSlots() uint {
	return 1 << table.w
}

// Return a count of leaf nodes in this table.
func (table *Table) getLeafCount() (count uint) {
	for i := 0; i < len(table.slots); i++ {
		node := table.slots[i]
		if node != nil {
			if node.IsLeaf() {
				count++
			} else {
				tDeeper := node.(*Table)
				count += tDeeper.getLeafCount()
			}
		}
	}
	return
}

//
func (table *Table) getTableCount() (count uint) {
	count = 1
	for i := 0; i < len(table.slots); i++ {
		node := table.slots[i]
		if node != nil && !node.IsLeaf() {
			tDeeper := node.(*Table)
			count += tDeeper.getTableCount()
		}
	}
	return
}

//func (table *Table) GetDepth() uint {
//	return uint(table.depth)
//}

// XXX Should modify to remove empty table recursively where this
// is the last entry in the table
func (table *Table) removeFromSlices(offset uint) (err error) {
	curSize := uint(len(table.slots))
	if curSize == 0 {
		err = DeleteFromEmptyTable
	} else if offset >= curSize {
		err = errors.New(fmt.Sprintf(
			"InternalError: delete offset %d but table size %d\n",
			offset, curSize))
	} else if curSize == 1 {
		// XXX LEAVES EMPTY TABLE, WHICH WILL HARM PERFORMANCE
		table.slots = table.slots[0:0]
	} else if offset == 0 {
		table.slots = table.slots[1:]
	} else if offset == curSize-1 {
		table.slots = table.slots[0:offset]
	} else {
		// Get rid of expensive appends.
		// shorterSlots := table.slots[0:offset]
		// shorterSlots = append(shorterSlots, table.slots[offset+1:]...)
		shorterSlots := make([]HTNodeI, curSize-1)
		copy(shorterSlots[0:offset], table.slots[0:offset])
		copy(shorterSlots[offset:], table.slots[offset+1:])
		table.slots = shorterSlots
	}
	return
}

// Enter with hc the hashcode for the key shifted appropriately for the
// current depth, so that the first w bits of the shifted hashcode can
// be used as the index of the leaf in the table.
//
// The caller guarantees that depth <= Root.maxDepth.
func (table *Table) deleteLeaf(hc uint64, depth uint, key KeyI) (
	err error) {

	if len(table.slots) == 0 {
		err = NotFound
	} else {
		ndx := hc & table.mask
		flag := uint64(1 << ndx)
		mask := flag - 1
		if table.bitmap&flag == 0 {
			err = NotFound
		} else {
			// the node is present; get its position in the slice
			var slotNbr uint
			if mask != 0 {
				slotNbr = BitCount64(table.bitmap & mask)
			}
			node := table.slots[slotNbr]
			if node.IsLeaf() {
				myLeaf := node.(*Leaf)
				myKey := myLeaf.Key.(*BytesKey)
				searchKey := key.(*BytesKey)
				if bytes.Equal(searchKey.Slice, myKey.Slice) {
					err = table.removeFromSlices(slotNbr)
					table.bitmap &= ^flag
				} else {
					err = NotFound
				}
			} else {
				// node is a table, so recurse
				depth++
				if depth > table.root.maxTableDepth {
					err = NotFound
				} else {
					tDeeper := node.(*Table)
					hc >>= table.w
					err = tDeeper.deleteLeaf(hc, depth, key)
				}
			}
		}
	}
	return
}

// Enter with hc the hashcode for the key shifted appropriately for the
// current depth, the depth as a zero-based integer, and the full key.
// Return nil if no matching entry is found or the value associated with
// the matching entry or any error encountered.
//
// The caller guarantees that depth<=Root.maxDepth.
//
func (table *Table) findLeaf(hc uint64, depth uint, key KeyI) (
	value interface{}, err error) {

	p := &table.slots // 14 of 46 samples; 22 of 41; 17 of 43
	sliceSize := uint(len(*p))
	if sliceSize != 0 {
		ndx := hc & table.mask
		flag := uint64(1 << ndx)
		mask := flag - 1
		if table.bitmap&flag != 0 {
			// the node is present; get its position in the slice
			var slotNbr uint
			if mask != 0 {
				slotNbr = uint(BitCount64(table.bitmap & mask))
			}
			node := (*p)[slotNbr] // 20 of 46; 14 of 41; 23 of 43
			if node.IsLeaf() {
				myLeaf := node.(*Leaf)
				myKey := myLeaf.Key.(*BytesKey)
				searchKey := key.(*BytesKey)
				if bytes.Equal(searchKey.Slice, myKey.Slice) {
					value = myLeaf.Value
				}
				// otherwise the value returned is nil
			} else {
				// node is a table, so recurse
				depth++
				if depth <= table.root.maxTableDepth {
					tDeeper := node.(*Table)
					hc >>= table.w
					value, err = tDeeper.findLeaf(hc, depth, key)
				}
				// otherwise the value returned is nil
			}
		}
	}
	return
}

// Enter with hc having been shifted so that the first w bits are ndx.
// 2014-05-13: Performance of this function was considerably improved (runtime
// down 25-50%) by replacing slice appends with slice make/copy sequences.
//
// The caller guarantees that depth <= Root.maxDepth.
func (table *Table) insertLeaf(hc uint64, depth uint, leaf *Leaf) (err error) {

	var slotNbr uint // whatever is in first line: about 15 of 37
	ndx := hc & table.mask
	flag := uint64(1 << ndx)
	mask := flag - 1
	if mask != 0 {
		slotNbr = BitCount64(table.bitmap & mask)
	}
	p := &table.slots
	sliceSize := uint(len(*p))
	if sliceSize == 0 {
		table.slots = []HTNodeI{leaf}
		table.bitmap |= flag
	} else {
		// Is there is already something at this slotNbr ?
		if table.bitmap&flag != 0 {
			entry := (*p)[slotNbr]

			if entry.IsLeaf() {
				// if it's a leaf, we replace the value iff the keys match
				curLeaf := entry.(*Leaf)
				curKey := curLeaf.Key.(*BytesKey)
				newKey := leaf.Key.(*BytesKey)
				if bytes.Equal(curKey.Slice, newKey.Slice) {
					// the keys match, so we replace the value
					curLeaf.Value = leaf.Value
				} else {
					var (
						tableDeeper *Table
					)
					depth++
					if depth > table.root.maxTableDepth {
						err = MaxTableDepthExceeded
					} else {
						oldLeaf := entry.(*Leaf)
						tableDeeper, err = NewTableWithLeaf(
							depth, table.root, oldLeaf)
						if err == nil {
							hc >>= table.w // this is hashcode for the NEW leaf
							// then put the new leaf in the new table
							err = tableDeeper.insertLeaf(hc, depth, leaf)
							if err == nil {
								// the new table replaces the existing leaf
								(*p)[slotNbr] = tableDeeper
							}
						}
					}
				}
			} else {
				depth++
				if depth > table.root.maxTableDepth {
					err = MaxTableDepthExceeded
				} else {
					// otherwise it's a table, so recurse
					tDeeper := entry.(*Table)
					hc >>= table.w
					err = tDeeper.insertLeaf(hc, depth, leaf)
				}
			}
		} else if slotNbr == 0 {
			leftSlots := make([]HTNodeI, sliceSize+1)
			leftSlots[0] = leaf
			copy(leftSlots[1:], table.slots[:])
			table.slots = leftSlots
			table.bitmap |= flag
		} else if slotNbr == sliceSize {
			table.slots = append(table.slots, leaf)
			table.bitmap |= flag
		} else {
			leftSlots := make([]HTNodeI, sliceSize+1)
			copy(leftSlots[:slotNbr], table.slots[:slotNbr])
			leftSlots[slotNbr] = leaf
			copy(leftSlots[slotNbr+1:], table.slots[slotNbr:])
			table.slots = leftSlots
			table.bitmap |= flag
		}
	}
	return
}

func (table *Table) IsLeaf() bool {
	return false
}
