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

	depth uint // only here for use in development and debugging !
}

func NewTable(depth, w, t uint) (table *Table, err error) {
	if w > 6 {
		err = MaxTableSizeExceeded
	} else {
		table = new(Table)
		table.depth = depth
		table.w = w
		table.t = t

		flag := uint64(1)
		flag <<= w
		table.mask = flag - 1
		err = table.CheckTableDepth(depth)
		if err != nil {
			table = nil
		}
	}
	return
}

// Return the maximum possible depth for the table, (64 - t)/w.  There are
// 64 bits available for keys, the root table uses t, each successive
// table uses w more bits.  The root table is at depth 0.
func (table *Table) MaxDepth() uint {
	return (64 - table.t) / table.w
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

func (table *Table) GetDepth() uint {
	return uint(table.depth)
}

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
// current depth.
//
func (table *Table) deleteLeaf(hc uint64, depth uint, key KeyI) (
	err error) {

	var (
		ndx64, flag, mask uint64
	)
	sliceSize := byte(len(table.slots))
	if sliceSize == 0 {
		err = NotFound
	}
	if err == nil {
		ndx64 = hc & table.mask
		flag = uint64(1 << ndx64)
		mask = flag - 1
		if table.bitmap&flag == 0 {
			err = NotFound
		}
	}
	if err == nil {
		// the node is present; get its position in the slice
		var pos byte
		if mask != 0 {
			pos = byte(BitCount64(table.bitmap & mask))
		}
		node := table.slots[pos]
		// XXX this MUST exist
		if node.IsLeaf() {
			// KEYS MUST BE OF THE SAME TYPE
			myLeaf := node.(*Leaf)
			myKey := myLeaf.Key.(*BytesKey)
			searchKey := key.(*BytesKey)
			if bytes.Equal(searchKey.Slice, myKey.Slice) {
				err = table.removeFromSlices(uint(pos))
				table.bitmap &= ^flag
			} else {
				err = NotFound
			}
		} else {
			// node is a table, so recurse
			tDeeper := node.(*Table)
			hc >>= table.w
			depth++
			err = tDeeper.deleteLeaf(hc, depth, key)
		}
	}
	return
}

// Enter with hc the hashcode for the key shifted appropriately for the
// current depth, the depth as a zero-based integer, and the full key.
// Return nil if no matching entry is found or the value associated with
// the matching entry or any error encountered.
//
func (table *Table) findLeaf(hc uint64, depth uint, key KeyI) (
	value interface{}, err error) {

	sliceSize := uint(len(table.slots))
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
			node := table.slots[slotNbr]
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
				tDeeper := node.(*Table)
				hc >>= table.w
				depth++
				value, err = tDeeper.findLeaf(hc, depth, key)
			}
		}
	}
	return
}

func (table *Table) CheckTableDepth(depth uint) (err error) {
	if depth > table.MaxDepth() {
		msg := fmt.Sprintf("Table depth is %d but max depth is %d\n",
			depth, table.MaxDepth())
		err = errors.New(msg)
	}
	return
}

// Enter with hc having been shifted so that the first w bits are ndx.
// 2014-05-13: Performance of this function was considerably improved (runtime
// down 25-50%) by replacing slice appends with slice make/copy sequences.
//
func (table *Table) insertLeaf(hc uint64, depth uint, leaf *Leaf) (
	slotNbr uint, err error) {

	err = table.CheckTableDepth(depth)
	if err != nil {
		return
	}
	var (
		ndx64, flag, mask uint64
	)
	sliceSize := uint(len(table.slots))
	if err == nil {
		ndx64 = hc & table.mask
		flag = uint64(1 << ndx64)
		mask = flag - 1
	}
	if sliceSize == 0 {
		table.slots = []HTNodeI{leaf}
	} else {
		if mask != 0 {
			slotNbr = BitCount64(table.bitmap & mask)
		}
		// Is there is already something at this slotNbr ?
		if table.bitmap&flag != 0 {
			entry := table.slots[slotNbr]

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
						oldHC       uint64
					)
					depth++
					tableDeeper, err = NewTable(depth, table.w, table.t)
					if err == nil {
						hc >>= table.w // this is hashcode for the NEW leaf
						oldLeaf := entry.(*Leaf)
						oldHC = oldLeaf.Key.Hashcode()
						oldHC >>= table.t + (depth-1)*table.w
						// put the existing leaf into the new table
						_, err = tableDeeper.insertLeaf(oldHC, depth, oldLeaf)
						if err == nil {
							// then put the new leaf in the new table
							_, err = tableDeeper.insertLeaf(hc, depth, leaf)
							if err == nil {
								// the new table replaces the existing leaf
								table.slots[slotNbr] = tableDeeper
							}
						}
					}
				}
			} else {
				// otherwise it's a table, so recurse
				tDeeper := entry.(*Table)
				hc >>= table.w
				depth++
				_, err = tDeeper.insertLeaf(hc, depth, leaf)
			}
		} else if slotNbr == 0 {
			leftSlots := make([]HTNodeI, sliceSize+1)
			leftSlots[0] = leaf
			copy(leftSlots[1:], table.slots[:])
			table.slots = leftSlots
		} else if slotNbr == sliceSize {
			table.slots = append(table.slots, leaf)
		} else {
			leftSlots := make([]HTNodeI, sliceSize+1)
			copy(leftSlots[:slotNbr], table.slots[:slotNbr])
			leftSlots[slotNbr] = leaf
			copy(leftSlots[slotNbr+1:], table.slots[slotNbr:])
			table.slots = leftSlots
		}
	}
	if err == nil {
		table.bitmap |= flag
	}
	return
}

func (table *Table) IsLeaf() bool {
	return false
}
