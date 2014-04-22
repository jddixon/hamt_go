package hamt_go

// hamt_go/table.go

import (
	"bytes"
	"errors"
	"fmt"
)

var _ = fmt.Print

type Table struct {
	depth    uint // only here for use in development and debugging !
	w        uint // non-root tables have 2^n slots
	t        uint // root table have 2^(t+n) slots
	maxDepth uint
	maxSlots uint // maximum slots for table at this depth
	mask     uint64
	indices  []byte // probably only used in development and debugging
	bitmap   uint64
	slots    []*Entry // each nil or a pointer to either a leaf or a table
}

func NewTable(depth, w, t uint) (table *Table, err error) {
	table = new(Table)
	table.depth = depth
	table.w = w
	table.t = t
	var exp uint // power of 2
	if depth == 0 {
		exp = t + w
	} else {
		exp = w
	}
	flag := uint64(1)
	flag <<= exp
	table.mask = flag - 1
	table.maxDepth = (64 / w) // rounds down	XXX NO ALLOWANCE FOR t
	if depth == 0 {
		table.maxSlots = maxSlots(t + w)
	} else {
		table.maxSlots = maxSlots(w)
	}

	// DEBUG
	//fmt.Printf("NewTable: depth %d/%d, w %d, t %d, maxSlots %d\n",
	//	depth, table.maxDepth, w, t, table.maxSlots)
	// END

	err = table.CheckTableDepth(depth)
	if err != nil {
		table = nil
	}
	return
}

// Return a count of leaf nodes in this table.
func (table *Table) GetLeafCount() (count uint) {
	for i := 0; i < len(table.slots); i++ {
		node := table.slots[i].Node
		if node != nil && node.IsLeaf() {
			count++
		}
	}
	return
}

//
func (table *Table) GetTableCount() (count uint) {
	count = 1
	for i := 0; i < len(table.slots); i++ {
		node := table.slots[i].Node
		if node != nil && !node.IsLeaf() {
			tDeeper := node.(*Table)
			count += tDeeper.GetTableCount()
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
func (table *Table) deleteEntry(hc uint64, depth uint, key KeyI) (
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
		// the entry is present; get its position in the slice
		var pos byte
		if mask != 0 {
			pos = byte(BitCount64(table.bitmap & mask))
		}
		entry := table.slots[pos]
		// XXX this MUST exist
		if entry.Node.IsLeaf() {
			// KEYS MUST BE OF THE SAME TYPE
			myLeaf := entry.Node.(*Leaf)
			myKey := myLeaf.Key.(*BytesKey)
			searchKey := key.(*BytesKey)
			if bytes.Equal(searchKey.Slice, myKey.Slice) {
				err = table.removeFromSlices(uint(pos))
				table.bitmap &= ^flag
			} else {
				err = NotFound
			}
		} else {
			// entry is a table, so recurse
			tDeeper := entry.Node.(*Table)
			hc >>= table.w
			depth++
			err = tDeeper.deleteEntry(hc, depth, key)
		}
	}
	return
}

// Enter with hc the hashcode for the key shifted appropriately for the
// current depth.
//
func (table *Table) findEntry(hc uint64, depth uint, key KeyI) (
	value interface{}, err error) {

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
		// the entry is present; get its position in the slice
		var pos byte
		if mask != 0 {
			pos = byte(BitCount64(table.bitmap & mask))
		}
		entry := table.slots[pos]
		// XXX this MUST exist
		if entry.Node.IsLeaf() {
			myLeaf := entry.Node.(*Leaf)
			myKey := myLeaf.Key.(*BytesKey)
			searchKey := key.(*BytesKey)
			if bytes.Equal(searchKey.Slice, myKey.Slice) {
				value = myLeaf.Value
				// fmt.Printf("    FOUND, slot %d\n", i) // DEBUG
			} else {
				//fmt.Printf("    LEAF, NO MATCH\n") // DEBUG
				err = NotFound
			}
		} else {
			// entry is a table, so recurse
			tDeeper := entry.Node.(*Table)
			hc >>= table.w
			depth++
			value, err = tDeeper.findEntry(hc, depth, key)
			// DEBUG
			// if err != nil {
			//	fmt.Printf(
			//	"    findEntry depth %d BACK FROM RECURSION: err %v\n",
			//		depth-1, err)
			// }
			// END
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
func (table *Table) insertEntry(hc uint64, depth uint, entry *Entry) (
	slotNbr uint, err error) {

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
		// DEBUG
		//fmt.Printf("  insert into empty table: depth %d ndx64 %02x\n",
		//	depth, ndx64)
		// END
		table.slots = append(table.slots, entry)
		table.indices = append(table.indices, byte(ndx64))
	} else {
		ndx := byte(ndx64)
		if mask != 0 {
			slotNbr = BitCount64(table.bitmap & mask)
		}
		if table.bitmap&flag != 0 {
			// there is already something at this slotNbrition
			err = table.insertIntoOccupiedSlot(hc, depth, entry, slotNbr, ndx)
		} else if slotNbr == 0 {
			var leftSlots []*Entry
			var leftIndices []byte
			leftSlots = append(leftSlots, entry)
			leftSlots = append(leftSlots, table.slots...)
			table.slots = leftSlots
			leftIndices = append(leftIndices, ndx)
			leftIndices = append(leftIndices, table.indices...)
			table.indices = leftIndices
		} else if slotNbr == sliceSize {
			table.slots = append(table.slots, entry)
			table.indices = append(table.indices, ndx)
		} else {
			var leftSlots []*Entry
			var leftIndices []byte

			leftSlots = append(leftSlots, table.slots[:slotNbr]...)
			leftSlots = append(leftSlots, entry)
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
	entry *Entry, slotNbr uint, ndx byte) (err error) {

	e := table.slots[slotNbr]

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

		depth++
		tableDeeper, err = NewTable(depth, table.w, 0)
		if err == nil {
			newHC >>= table.w // this is hc for the NEW entry

			oldEntry = e
			oldLeaf := e.Node.(*Leaf)
			oldHC, err = oldLeaf.Key.Hashcode()
		}
		if err == nil {
			oldHC >>= depth * table.w
			// indexes for this depth

			// put the existing leaf into the new table
			_, err = tableDeeper.insertEntry(oldHC, depth, oldEntry)
			if err == nil {
				// then put the new entry in the new table
				_, err = tableDeeper.insertEntry(newHC, depth, entry)
				if err == nil {
					// the new table replaces the existing leaf
					var eTab *Entry
					eTab, err = NewEntry(ndx, tableDeeper)
					if err == nil {
						table.slots[slotNbr] = eTab
					}
				}
			}
		}
	} else {
		// otherwise it's a table, so recurse
		tDeeper := e.Node.(*Table)
		newHC >>= table.w
		depth++
		_, err = tDeeper.insertEntry(newHC, depth, entry)

	}
	return
}

func (table *Table) IsLeaf() bool {
	return false
}
