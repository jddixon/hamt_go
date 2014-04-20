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
	maxSlots uint // maximum slots for table at this depth
	mask     uint64
	indices  []byte // probably only used in development and debugging
	bitmap   uint64
	slots    []*Entry // each nil or a pointer to either a leaf or a table
}

func NewTable(depth, w, t uint) (table *Table, err error) {
	err = CheckTableDepth(depth)
	if err == nil {
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

	curSize := uint(len(table.indices))
	// curSlotCount := uint(len(table.slots))

	if curSize == 0 {
		err = NotFound
	} else {

		// ndx is the value of the next W key bits
		ndx := byte(hc & table.mask)
		for i := uint(0); i < curSize; i++ {
			curNdx := table.indices[i]
			if curNdx < ndx {
				continue
			} else if curNdx == ndx {
				entry := table.slots[i]
				// XXX this MUST exist
				if entry.Node.IsLeaf() {
					// KEYS MUST BE OF THE SAME TYPE
					myLeaf := entry.Node.(*Leaf)
					myKey := myLeaf.Key.(*BytesKey)
					searchKey := key.(*BytesKey)
					if bytes.Equal(searchKey.Slice, myKey.Slice) {
						err = table.removeFromSlices(i)
					} else {
						err = NotFound
					}
				} else {
					// entry is a table, so recurse
					tDeeper := entry.Node.(*Table)
					hc >>= W
					depth++
					err = tDeeper.deleteEntry(hc, depth, key)
				}
				break
			} else {
				// curNdx > ndx, so it's not there
				err = NotFound
				break
			}
		}
	}
	return
}

// Enter with hc the hashcode for the key shifted appropriately for the
// current depth.
//
// XXX THIS WORKS, but a binary search would be faster.
//
func (table *Table) findEntry(hc uint64, depth uint, key KeyI) (
	value interface{}, err error) {

	curSize := uint(len(table.indices))
	curSlotCount := uint(len(table.slots))

	// DEBUG
	if curSize != curSlotCount {
		fmt.Printf("curSize %d but curSlotCount %d !!\n",
			curSize, curSlotCount)
	}
	//if depth > 0 {
	//fmt.Printf("findEntry: hc %08x, depth %d, curSize %d\n",
	//	hc, depth, curSize)
	//}
	// END

	if curSize == 0 {
		err = NotFound
	} else {
		// ndx is the value of the next W key bits
		ndx := byte(hc & table.mask)
		/////////////////////////////////////////////////////////////////
		// XXX This linear search is VERY expensive in terms of run time.
		/////////////////////////////////////////////////////////////////
		for i := uint(0); i < curSize; i++ {
			curNdx := table.indices[i]
			// DEBUG
			//fmt.Printf(
			//	"  findEntry, depth %d, ndx %2x, slot %2d, slot ndx %02x ",
			//	depth, ndx, i, curNdx)
			// END
			if curNdx < ndx {
				if i < curSize-1 {
					// fmt.Printf("continuing\n") // DEBUG
					continue
				} else {
					// fmt.Printf("no more slots\n") // DEBUG
					err = NotFound
				}
			} else if curNdx == ndx {
				// DEBUG
				// fmt.Printf("MATCH: curNdx %02x == ndx %02x\n", curNdx, ndx)
				// END
				entry := table.slots[i]
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
					hc >>= W
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
				break
			} else {
				// curNdx > ndx, so it's not there
				// DEBUG
				//fmt.Printf("NO MATCH: curNdx %02x > ndx %02x\n", curNdx, ndx)
				// END
				err = NotFound
				break
			}
		}
	}
	return
}

// We need to insert a new entry into a slot which is already occupied.
//
func (table *Table) insertAtMatch(newHC uint64, depth uint, entry *Entry,
	i uint, ndx byte) (err error) {

	// fmt.Printf("insertAtMatch: ndx == %02x\n", ndx) // DEBUG
	e := table.slots[i]

	// LEAF -----------------------------------
	// if it's a leaf, we replace the value only iff the keys match
	if e.Node.IsLeaf() {
		// fmt.Printf("FOUND LEAF\n")

		//////////////////////////////////////
		// XXX NOT CHECKING FOR DUPLICATE KEYS
		//////////////////////////////////////

		var (
			tableDeeper *Table
			oldEntry    *Entry
			oldHC       uint64
		)

		depth++
		// DEBUG
		// fmt.Printf("  CREATING TABLE AT DEPTH %d\n", depth)
		// END
		tableDeeper, err = NewTable(depth, table.w, 0)
		if err == nil {
			newHC >>= W // this is hc for the NEW entry

			oldEntry = e
			oldLeaf := e.Node.(*Leaf)
			oldHC, err = oldLeaf.Key.Hashcode()
		}
		if err == nil {
			oldHC >>= depth * W
			// indexes for this depth

			// put the existing leaf into the new table
			// DEBUG
			// fmt.Printf("    inserting OLD Entry into new table\n")
			// END
			_, err = tableDeeper.insertEntry(oldHC, depth, oldEntry)
			if err == nil {
				// then put the new entry in the new table
				// DEBUG
				//fmt.Printf("    adding NEW entry to new table\n")
				// END
				_, err = tableDeeper.insertEntry(newHC, depth, entry)
				if err == nil {
					// the new table replaces the existing leaf
					var eTab *Entry
					eTab, err = NewEntry(ndx, tableDeeper)
					// DEBUG
					//fmt.Printf(
					//	"    FINISHED TABLE depth %d, ndx %02x\n",
					//	depth, ndx)
					// END
					if err == nil {
						table.slots[i] = eTab
					}
				}
			}
		}

		// TABLE ----------------------------------
		// otherwise it's a table, so recurse
	} else {
		// fmt.Printf("FOUND TABLE, recursing\n")
		tDeeper := e.Node.(*Table)
		newHC >>= W
		depth++
		_, err = tDeeper.insertEntry(newHC, depth, entry)

	}
	return // FOO
}

func CheckTableDepth(depth uint) (err error) {
	if depth > MAX_DEPTH {
		msg := fmt.Sprintf("Table64: depth is %d but MaxDepth is %d\n",
			depth, MAX_DEPTH)
		err = errors.New(msg)
	}
	return
}

// Enter with hc having been shifted to remove preceding ndx, if any
//
func (table *Table) insertEntry(hc uint64, depth uint, entry *Entry) (
	slotNbr uint, err error) {

	err = CheckTableDepth(depth)
	if err != nil {
		return
	}

	// ndx is the value of the next W key bits
	ndx := byte(hc & table.mask)

	curSize := uint(len(table.indices))

	// DEBUG
	//curSlotCount := uint(len(table.slots))
	//fmt.Printf(
	//	"\nTable.insertEntry: depth %d, hc %08x, ndx %02x, index count %d, slot count %d\n",
	//	depth, hc, ndx, curSize, curSlotCount)
	// END

	if curSize == 0 {
		// DEBUG
		//fmt.Printf("  insertion empty table: depth %d ndx %02x\n", depth, ndx)
		// END
		table.slots = append(table.slots, entry)
		table.indices = append(table.indices, ndx)
	} else {
		inserted := false
		var i uint
		var curNdx, nextNdx byte
		for i = 0; i < curSize-1; i++ {
			curNdx = table.indices[i]
			nextNdx = table.indices[i+1]
			// DEBUG
			//fmt.Printf("  insertion: depth %d ndx %02x curNdx %02x\n",
			//	depth, ndx, curNdx)
			// END

			// XXX MESSY LOGIC
			if curNdx == ndx {
				err = table.insertAtMatch(hc, depth, entry, i, ndx)
				if err == nil {
					inserted = true
				}
				break
			} else if nextNdx == ndx {
				err = table.insertAtMatch(hc, depth, entry, i+1, ndx)
				if err == nil {
					inserted = true
				}
				break
			} else if curNdx < ndx {
				if nextNdx < ndx {
					//fmt.Printf("continuing: %02x after %02x, after %02x\n",
					//	ndx, curNdx, nextNdx)
					continue
				}
				slotNbr = i + 1
				//fmt.Printf(
				//	"A: inserting %02x after %02x, before %02x, at %d\n",
				//	ndx, curNdx, nextNdx, slotNbr)

				// first insert the index ---------------------------
				var leftNdx []byte
				if slotNbr > 0 {
					leftNdx = append(leftNdx, table.indices[0:slotNbr]...)
				}
				rightNdx := table.indices[slotNbr:]
				//fmt.Printf("%s + %02x + %s => ",
				//	dumpByteSlice(leftNdx), ndx, dumpByteSlice(rightNdx))
				leftNdx = append(leftNdx, ndx)
				leftNdx = append(leftNdx, rightNdx...)

				//fmt.Printf("%s\n", dumpByteSlice(leftNdx))
				table.indices = leftNdx // FOO

				// then insert the entry ----------------------------
				var leftSlots []*Entry
				if slotNbr > 0 {
					leftSlots = append(leftSlots, table.slots[0:slotNbr]...)
				}
				rightSlots := table.slots[slotNbr:]
				leftSlots = append(leftSlots, entry)
				leftSlots = append(leftSlots, rightSlots...)
				table.slots = leftSlots // FOO

				// done ---------------------------------------------
				inserted = true
				break
			} else {
				slotNbr = i
				//fmt.Printf("B: inserting %02x before %02x at %d\n",
				//	ndx, curNdx, slotNbr)

				// first insert the index ---------------------------
				var leftNdx []byte
				if slotNbr > 0 {
					leftNdx = append(leftNdx, table.indices[0:slotNbr]...)
				}
				rightNdx := table.indices[slotNbr:]
				//fmt.Printf("%s + %02x + %s\n",
				//	dumpByteSlice(leftNdx), ndx, dumpByteSlice(rightNdx))
				leftNdx = append(leftNdx, ndx)
				leftNdx = append(leftNdx, rightNdx...)
				table.indices = leftNdx // FOO

				// then insert the entry ----------------------------
				var leftSlots []*Entry
				if slotNbr > 0 {
					leftSlots = append(leftSlots, table.slots[0:slotNbr]...)
				}
				rightSlots := table.slots[slotNbr:]
				leftSlots = append(leftSlots, entry)
				leftSlots = append(leftSlots, rightSlots...)
				table.slots = leftSlots // FOO

				// done ---------------------------------------------
				inserted = true
				break

			}
		}
		if !inserted {
			curNdx = table.indices[i]
			curEntry := table.slots[i]

			if curNdx < ndx {
				// insert index -------------------------------------
				//fmt.Printf("C: appending %02x after %02x\n", ndx, curNdx)
				table.indices = append(table.indices, ndx)
				slotNbr = curSize
				// insert entry -------------------------------------
				table.slots = append(table.slots, entry)
			} else if curNdx == ndx {
				//fmt.Printf("* inserting at MATCH: curNdx = ndx = %02x\n", ndx)
				err = table.insertAtMatch(hc, depth, entry, i, ndx)
			} else {
				// insert index -------------------------------------
				leftNdx := (table.indices)[0:i]
				leftNdx = append(leftNdx, ndx)
				leftNdx = append(leftNdx, curNdx)
				table.indices = leftNdx // FOO
				slotNbr = curSize - 1
				//fmt.Printf("D: prepended %02x before %02x at %d\n",
				//	ndx, curNdx, slotNbr)

				// insert entry -------------------------------------
				leftSlots := (table.slots)[0:i]
				leftSlots = append(leftSlots, entry)
				leftSlots = append(leftSlots, curEntry)
				table.slots = leftSlots
			}

		}
	}
	// XXX IGNORES POSSIBLE ERRORS XXX
	flag := uint64(1 << ndx)
	table.bitmap |= flag
	// DEBUG
	//newSize := uint(len(table.indices))
	//fmt.Printf("  depth %d: inserted 0x%02x at %d/%d\n",
	//	depth, ndx, slotNbr, newSize)
	//fmt.Printf("  DEPTH %d INDICES: %s\n", depth, dumpByteSlice(table.indices))
	//// END
	return
}

func (table *Table) IsLeaf() bool { return false }
