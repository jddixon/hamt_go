package hamt_go

// hamt_go/table32.go

import (
	"bytes"
	"errors"
	"fmt"
)

var _ = fmt.Print

type Table32 struct {
	// prefix []byte	// XXX could add pfor debugging
	depth   uint   // only here for use in development and debugging !
	indices []byte // probably only used in development and debugging
	bitmap  uint32
	slots   []*Entry32 // each nil or a pointer to either a leaf or a table
}

func NewTable32(depth uint) (t32 *Table32, err error) {
	err = CheckTable32Depth(depth)
	if err == nil {
		t32 = new(Table32)
		t32.depth = depth
	}
	return
}

func (t32 *Table32) GetDepth() uint {
	return uint(t32.depth)
}

func (t32 *Table32) removeFromSlices(offset uint) (err error) {
	curSize := uint(len(t32.indices))
	if curSize == 0 {
		err = DeleteFromEmptyTable
	} else if offset >= curSize {
		err = errors.New(fmt.Sprintf(
			"InternalError: delete offset %d but table size %d\n",
			offset, curSize))
	} else if curSize == 1 {
		// XXX LEAVES EMPTY TABLE, WHICH WILL HARM PERFORMANCE
		t32.indices = t32.indices[0:0]
		t32.slots = t32.slots[0:0]
	} else if offset == 0 {
		t32.indices = t32.indices[1:]
		t32.slots = t32.slots[1:]
	} else if offset == curSize-1 {
		t32.indices = t32.indices[0:offset]
		t32.slots = t32.slots[0:offset]
	} else {
		shorterNdx := t32.indices[0:offset]
		shorterNdx = append(shorterNdx, t32.indices[offset+1:]...)
		t32.indices = shorterNdx
		shorterSlots := t32.slots[0:offset]
		shorterSlots = append(shorterSlots, t32.slots[offset+1:]...)
		t32.slots = shorterSlots
	}
	return
}

// Enter with hc the hashcode for the key shifted appropriately for the
// current depth.
//
func (t32 *Table32) deleteEntry(hc uint32, depth uint, key Key32I) (
	err error) {

	curSize := uint(len(t32.indices))
	// curSlotCount := uint(len(t32.slots))

	if curSize == 0 {
		err = NotFound
	} else {

		// ndx is the value of the next W32 key bits
		ndx := byte(hc & LEVEL_MASK32)
		for i := uint(0); i < curSize; i++ {
			curNdx := t32.indices[i]
			if curNdx < ndx {
				continue
			} else if curNdx == ndx {
				entry := t32.slots[i]
				// XXX this MUST exist
				if entry.Node.IsLeaf() {
					// KEYS MUST BE OF THE SAME TYPE
					myLeaf := entry.Node.(*Leaf32)
					myKey := myLeaf.Key.(*Bytes32Key)
					searchKey := key.(*Bytes32Key)
					if bytes.Equal(searchKey.Slice, myKey.Slice) {
						err = t32.removeFromSlices(i)
					} else {
						err = NotFound
					}
				} else {
					// entry is a table, so recurse
					tDeeper := entry.Node.(*Table32)
					hc >>= W32
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
func (t32 *Table32) findEntry(hc uint32, depth uint, key Key32I) (
	value interface{}, err error) {

	curSize := uint(len(t32.indices))
	curSlotCount := uint(len(t32.slots))

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
		// ndx is the value of the next W32 key bits
		ndx := byte(hc & LEVEL_MASK32)
		/////////////////////////////////////////////////////////////////
		// XXX This linear search is VERY expensive in terms of run time.
		/////////////////////////////////////////////////////////////////
		for i := uint(0); i < curSize; i++ {
			curNdx := t32.indices[i]
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
				entry := t32.slots[i]
				// XXX this MUST exist
				if entry.Node.IsLeaf() {
					myLeaf := entry.Node.(*Leaf32)
					myKey := myLeaf.Key.(*Bytes32Key)
					searchKey := key.(*Bytes32Key)
					if bytes.Equal(searchKey.Slice, myKey.Slice) {
						value = myLeaf.Value
						// fmt.Printf("    FOUND, slot %d\n", i) // DEBUG
					} else {
						//fmt.Printf("    LEAF, NO MATCH\n") // DEBUG
						err = NotFound
					}
				} else {
					// entry is a table, so recurse
					tDeeper := entry.Node.(*Table32)
					hc >>= W32
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
func (t32 *Table32) insertAtMatch(newHC uint32, depth uint, entry *Entry32,
	i uint, ndx byte) (err error) {

	// fmt.Printf("insertAtMatch: ndx == %02x\n", ndx) // DEBUG
	e := t32.slots[i]

	// LEAF -----------------------------------
	// if it's a leaf, we replace the value only iff the keys match
	if e.Node.IsLeaf() {
		// fmt.Printf("FOUND LEAF\n")

		//////////////////////////////////////
		// XXX NOT CHECKING FOR DUPLICATE KEYS
		//////////////////////////////////////

		var (
			t32Deeper *Table32
			oldEntry  *Entry32
			oldHC     uint32
		)

		depth++
		// DEBUG
		// fmt.Printf("  CREATING TABLE AT DEPTH %d\n", depth)
		// END
		t32Deeper, err = NewTable32(depth)
		if err == nil {
			newHC >>= W32 // this is hc for the NEW entry

			oldEntry = e
			oldLeaf := e.Node.(*Leaf32)
			oldHC, err = oldLeaf.Key.Hashcode32()
		}
		if err == nil {
			oldHC >>= depth * W32
			// indexes for this depth

			// put the existing leaf into the new table
			// DEBUG
			// fmt.Printf("    inserting OLD Entry into new table\n")
			// END
			_, err = t32Deeper.insertEntry(oldHC, depth, oldEntry)
			if err == nil {
				// then put the new entry in the new table
				// DEBUG
				//fmt.Printf("    adding NEW entry to new table\n")
				// END
				_, err = t32Deeper.insertEntry(newHC, depth, entry)
				if err == nil {
					// the new table replaces the existing leaf
					var eTab *Entry32
					eTab, err = NewEntry32(ndx, t32Deeper)
					// DEBUG
					//fmt.Printf(
					//	"    FINISHED TABLE depth %d, ndx %02x\n",
					//	depth, ndx)
					// END
					if err == nil {
						t32.slots[i] = eTab
					}
				}
			}
		}

		// TABLE ----------------------------------
		// otherwise it's a table, so recurse
	} else {
		// fmt.Printf("FOUND TABLE, recursing\n")
		tDeeper := e.Node.(*Table32)
		newHC >>= W32
		depth++
		_, err = tDeeper.insertEntry(newHC, depth, entry)

	}
	return // FOO
}

func CheckTable32Depth(depth uint) (err error) {
	if depth > MAX_DEPTH32 {
		msg := fmt.Sprintf("Table32: depth is %d but MaxDepth is %d\n",
			depth, MAX_DEPTH32)
		err = errors.New(msg)
	}
	return
}

// Enter with hc having been shifted to remove preceding ndx, if any
//
func (t32 *Table32) insertEntry(hc uint32, depth uint, entry *Entry32) (
	slotNbr uint, err error) {

	err = CheckTable32Depth(depth)
	if err != nil {
		return
	}

	// ndx is the value of the next W32 key bits
	ndx := byte(hc & LEVEL_MASK32)

	curSize := uint(len(t32.indices))

	// DEBUG
	//curSlotCount := uint(len(t32.slots))
	//fmt.Printf(
	//	"\nTable32.insertEntry: depth %d, hc %08x, ndx %02x, index count %d, slot count %d\n",
	//	depth, hc, ndx, curSize, curSlotCount)
	// END

	if curSize == 0 {
		// DEBUG
		//fmt.Printf("  insertion empty table: depth %d ndx %02x\n", depth, ndx)
		// END
		t32.slots = append(t32.slots, entry)
		t32.indices = append(t32.indices, ndx)
	} else {
		inserted := false
		var i uint
		var curNdx, nextNdx byte
		for i = 0; i < curSize-1; i++ {
			curNdx = t32.indices[i]
			nextNdx = t32.indices[i+1]
			// DEBUG
			//fmt.Printf("  insertion: depth %d ndx %02x curNdx %02x\n",
			//	depth, ndx, curNdx)
			// END

			// XXX MESSY LOGIC
			if curNdx == ndx {
				err = t32.insertAtMatch(hc, depth, entry, i, ndx)
				if err == nil {
					inserted = true
				}
				break
			} else if nextNdx == ndx {
				err = t32.insertAtMatch(hc, depth, entry, i+1, ndx)
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
					leftNdx = append(leftNdx, t32.indices[0:slotNbr]...)
				}
				rightNdx := t32.indices[slotNbr:]
				//fmt.Printf("%s + %02x + %s => ",
				//	dumpByteSlice(leftNdx), ndx, dumpByteSlice(rightNdx))
				leftNdx = append(leftNdx, ndx)
				leftNdx = append(leftNdx, rightNdx...)

				//fmt.Printf("%s\n", dumpByteSlice(leftNdx))
				t32.indices = leftNdx // FOO

				// then insert the entry ----------------------------
				var leftSlots []*Entry32
				if slotNbr > 0 {
					leftSlots = append(leftSlots, t32.slots[0:slotNbr]...)
				}
				rightSlots := t32.slots[slotNbr:]
				leftSlots = append(leftSlots, entry)
				leftSlots = append(leftSlots, rightSlots...)
				t32.slots = leftSlots // FOO

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
					leftNdx = append(leftNdx, t32.indices[0:slotNbr]...)
				}
				rightNdx := t32.indices[slotNbr:]
				//fmt.Printf("%s + %02x + %s\n",
				//	dumpByteSlice(leftNdx), ndx, dumpByteSlice(rightNdx))
				leftNdx = append(leftNdx, ndx)
				leftNdx = append(leftNdx, rightNdx...)
				t32.indices = leftNdx // FOO

				// then insert the entry ----------------------------
				var leftSlots []*Entry32
				if slotNbr > 0 {
					leftSlots = append(leftSlots, t32.slots[0:slotNbr]...)
				}
				rightSlots := t32.slots[slotNbr:]
				leftSlots = append(leftSlots, entry)
				leftSlots = append(leftSlots, rightSlots...)
				t32.slots = leftSlots // FOO

				// done ---------------------------------------------
				inserted = true
				break

			}
		}
		if !inserted {
			curNdx = t32.indices[i]
			curEntry := t32.slots[i]

			if curNdx < ndx {
				// insert index -------------------------------------
				//fmt.Printf("C: appending %02x after %02x\n", ndx, curNdx)
				t32.indices = append(t32.indices, ndx)
				slotNbr = curSize
				// insert entry -------------------------------------
				t32.slots = append(t32.slots, entry)
			} else if curNdx == ndx {
				//fmt.Printf("* inserting at MATCH: curNdx = ndx = %02x\n", ndx)
				err = t32.insertAtMatch(hc, depth, entry, i, ndx)
			} else {
				// insert index -------------------------------------
				leftNdx := (t32.indices)[0:i]
				leftNdx = append(leftNdx, ndx)
				leftNdx = append(leftNdx, curNdx)
				t32.indices = leftNdx // FOO
				slotNbr = curSize - 1
				//fmt.Printf("D: prepended %02x before %02x at %d\n",
				//	ndx, curNdx, slotNbr)

				// insert entry -------------------------------------
				leftSlots := (t32.slots)[0:i]
				leftSlots = append(leftSlots, entry)
				leftSlots = append(leftSlots, curEntry)
				t32.slots = leftSlots
			}

		}
	}
	// XXX IGNORES POSSIBLE ERRORS XXX
	flag := uint32(1 << ndx)
	t32.bitmap |= flag
	// DEBUG
	//newSize := uint(len(t32.indices))
	//fmt.Printf("  depth %d: inserted 0x%02x at %d/%d\n",
	//	depth, ndx, slotNbr, newSize)
	//fmt.Printf("  DEPTH %d INDICES: %s\n", depth, dumpByteSlice(t32.indices))
	//// END
	return
}

func (t32 *Table32) IsLeaf() bool { return false }

// ==================================================================

type Table64 struct {
	bitmap uint64
	slots  []Entry64I // nil or a key-value pair or a pointer to a table
}

func (t64 *Table64) IsLeaf() bool { return false }
