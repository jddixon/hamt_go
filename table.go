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
	w        uint // non-root tables have 2^w slots
	t        uint // root table has 2^t slots
	maxDepth uint
	mask     uint64
	bitmap   uint64
	slots    []HTNodeI // each nil or a pointer to either a leaf or a table

	indices []byte // slice holding index of each slot in use

	depth    uint // only here for use in development and debugging !
	maxSlots uint // maximum slots for table at this depth
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
		table.maxDepth = (64 / w) // rounds down	XXX WRONG: NO ALLOWANCE FOR t
		table.maxSlots = 1 << w

		err = table.CheckTableDepth(depth)
		if err != nil {
			table = nil
		}
	}
	return
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
	curSize := uint(len(table.indices))
	if curSize == 0 {
		err = DeleteFromEmptyTable
	} else if offset >= curSize {
		err = errors.New(fmt.Sprintf(
			"InternalError: delete offset %d but table size %d\n",
			offset, curSize))
	} else if curSize == 1 {
		// XXX LEAVES EMPTY TABLE, WHICH WILL HARM PERFORMANCE
		table.indices = table.indices[0:0]
		table.slots = table.slots[0:0]
	} else if offset == 0 {
		table.indices = table.indices[1:]
		table.slots = table.slots[1:]
	} else if offset == curSize-1 {
		table.indices = table.indices[0:offset]
		table.slots = table.slots[0:offset]
	} else {
		shorterNdx := table.indices[0:offset]
		shorterNdx = append(shorterNdx, table.indices[offset+1:]...)
		table.indices = shorterNdx
		shorterSlots := table.slots[0:offset]
		shorterSlots = append(shorterSlots, table.slots[offset+1:]...)
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
	sliceSize := byte(len(table.indices))
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

	var (
		ndx, flag, mask uint64
	)
	sliceSize := byte(len(table.indices))
	if sliceSize != 0 {
		ndx = hc & table.mask
		flag = uint64(1 << ndx)
		mask = flag - 1
		if table.bitmap&flag != 0 {
			// the node is present; get its position in the slice
			var pos byte
			if mask != 0 {
				pos = byte(BitCount64(table.bitmap & mask))
			}
			node := table.slots[pos]
			// XXX this MUST exist
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
	if depth > table.maxDepth {
		msg := fmt.Sprintf("Table depth is %d but max depth is %d\n",
			depth, table.maxDepth)
		err = errors.New(msg)
	}
	return
}

// Enter with hc having been shifted so that the first w bits are ndx.
//
func (table *Table) insertLeaf(hc uint64, depth uint, node HTNodeI) (
	slotNbr uint, err error) {

	// DEBUG
	if depth != table.depth {
		fmt.Printf("INTERNAL ERROR: inserting at depth %d but table depth is %d\n",
			depth, table.depth)
	}
	// END
	err = table.CheckTableDepth(depth)
	if err != nil {
		return
	}

	var (
		ndx64, flag, mask uint64
	)
	sliceSize := uint(len(table.indices))
	if err == nil {
		ndx64 = hc & table.mask
		flag = uint64(1 << ndx64)
		mask = flag - 1
	}

	if sliceSize == 0 {
		table.slots = append(table.slots, node)
		table.indices = append(table.indices, byte(ndx64))
	} else {
		ndx := byte(ndx64)
		if mask != 0 {
			slotNbr = BitCount64(table.bitmap & mask)
		}
		if table.bitmap&flag != 0 {
			// there is already something at this slotNbrition
			err = table.insertIntoOccupiedSlot(hc, depth, node, slotNbr, ndx)
		} else if slotNbr == 0 {
			var leftSlots []HTNodeI
			var leftIndices []byte
			leftSlots = append(leftSlots, node)
			leftSlots = append(leftSlots, table.slots...)
			table.slots = leftSlots
			leftIndices = append(leftIndices, ndx)
			leftIndices = append(leftIndices, table.indices...)
			table.indices = leftIndices
		} else if slotNbr == sliceSize {
			table.slots = append(table.slots, node)
			table.indices = append(table.indices, ndx)
		} else {
			var leftSlots []HTNodeI
			var leftIndices []byte

			leftSlots = append(leftSlots, table.slots[:slotNbr]...)
			leftSlots = append(leftSlots, node)
			leftSlots = append(leftSlots, table.slots[slotNbr:]...)
			table.slots = leftSlots

			leftIndices = append(leftIndices, table.indices[:slotNbr]...)
			leftIndices = append(leftIndices, ndx)
			leftIndices = append(leftIndices, table.indices[slotNbr:]...)
			table.indices = leftIndices
		}
	}
	if err == nil {
		table.bitmap |= flag
	}
	return
}

// Insert a new entry into a slot which is already occupied.
//
func (table *Table) insertIntoOccupiedSlot(newHC uint64, depth uint,
	node HTNodeI, slotNbr uint, ndx byte) (err error) {

	e := table.slots[slotNbr]

	if e.IsLeaf() {
		// if it's a leaf, we replace the value iff the keys match
		curLeaf := e.(*Leaf)
		curKey := curLeaf.Key.(*BytesKey)
		nodeAsLeaf := node.(*Leaf)
		newKey := nodeAsLeaf.Key.(*BytesKey)
		if bytes.Equal(curKey.Slice, newKey.Slice) {
			// the keys match, so we replace the value
			newLeaf := node.(*Leaf)
			curLeaf.Value = newLeaf.Value
		} else {
			var (
				tableDeeper *Table
				oldNode     HTNodeI
				oldHC       uint64
			)

			depth++
			tableDeeper, err = NewTable(depth, table.w, table.t)
			if err == nil {
				newHC >>= table.w // this is hc for the NEW node

				oldNode = e
				oldLeaf := e.(*Leaf)
				oldHC, err = oldLeaf.Key.Hashcode()
			}
			if err == nil {
				oldHC >>= table.t + (depth-1)*table.w

				// indexes for this depth

				// put the existing leaf into the new table
				_, err = tableDeeper.insertLeaf(oldHC, depth, oldNode)
				if err == nil {
					// then put the new node in the new table
					_, err = tableDeeper.insertLeaf(newHC, depth, node)
					if err == nil {
						// the new table replaces the existing leaf
						table.slots[slotNbr] = tableDeeper
					}
				}
			}
		}
	} else {
		// otherwise it's a table, so recurse
		tDeeper := e.(*Table)
		newHC >>= table.w
		depth++
		_, err = tDeeper.insertLeaf(newHC, depth, node)

	}
	return
}

func (table *Table) IsLeaf() bool {
	return false
}
