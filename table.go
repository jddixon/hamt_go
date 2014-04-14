package hamt_go

// hamt_go/table.go

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

var _ = fmt.Print

// DEBUG
func dumpIndices(indices []byte) string {
	var ss []string
	for i := 0; i < len(indices); i++ {
		ss = append(ss, fmt.Sprintf("%02x ", indices[i]))
	}
	return strings.Join(ss, "")
}

// END

type Table32 struct {
	// prefix []byte	// XXX could add pfor debugging
	depth   uint   // only here for use in development and debugging !
	indices []byte // probably only used in development and debugging
	bitmap  uint32
	slots   []*Entry32 // each nil or a pointer to either a leaf or a table
}

func NewTable32(depth uint) (t32 *Table32, err error) {
	// XXX CHECK FOR IMPOSSIBLE DEPTH

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
func (t32 *Table32) deleteEntry(hc uint32, depth uint, key Key32I) (err error) {

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
					hc >>= W32
					depth++
					err = t32.deleteEntry(hc, depth, key)
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
	// curSlotCount := uint(len(t32.slots))

	if curSize == 0 {
		err = NotFound
	} else {
		// ndx is the value of the next W32 key bits
		ndx := byte(hc & LEVEL_MASK32)
		for i := uint(0); i < curSize; i++ {
			curNdx := t32.indices[i]
			// DEBUG
			//fmt.Printf("  findEntry, ndx %2x, slot %2d, slot ndx %02x ",
			//	ndx, i, curNdx)
			// END
			if curNdx < ndx {
				if i < curSize-1 {
					// fmt.Printf("continuing\n")	// DEBUG
					continue
				} else {
					// fmt.Printf("no more slots\n")	// DEBUG
					err = NotFound
				}
			} else if curNdx == ndx {
				// DEBUG
				//fmt.Printf("MATCH: curNdx %02x == ndx %02x\n", curNdx, ndx)
				// END
				entry := t32.slots[i]
				// XXX this MUST exist
				if entry.Node.IsLeaf() {
					myLeaf := entry.Node.(*Leaf32)
					myKey := myLeaf.Key.(*Bytes32Key)
					searchKey := key.(*Bytes32Key)
					if bytes.Equal(searchKey.Slice, myKey.Slice) {
						value = myLeaf.Value
						//fmt.Printf("    FOUND, slot %d\n", i)	// DEBUG
					} else {
						//fmt.Printf("    LEAF, NO MATCH\n")	// DEBUG
						err = NotFound
					}
				} else {
					// entry is a table, so recurse
					hc >>= W32
					depth++
					value, err = t32.findEntry(hc, depth, key)
					//fmt.Printf("    RECURSING\n")	// DEBUG
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

func (t32 *Table32) insertEntry(hc uint32, depth uint, entry *Entry32) (
	slotNbr uint, err error) {

	// ndx is the value of the next W32 key bits
	ndx := byte(hc & LEVEL_MASK32)

	curSize := uint(len(t32.indices))

	// DEBUG
	curSlotCount := uint(len(t32.slots))
	fmt.Printf("Table32.insertEntry: depth %d, index count %d, slot count %d\n",
		depth, curSize, curSlotCount)
	// END

	if curSize == 0 {
		t32.slots = append(t32.slots, entry)
		t32.indices = append(t32.indices, ndx)
	} else {
		inserted := false
		var i uint
		var curNdx, nextNdx byte
		for i = 0; i < curSize-1; i++ {
			curNdx = t32.indices[i]
			if curNdx < ndx {
				nextNdx = t32.indices[i+1]
				if nextNdx < ndx {
					// fmt.Printf("continuing: %02x after %02x, after %02x\n",
					//	ndx, curNdx, nextNdx)
					continue
				}
				slotNbr = i + 1
				// fmt.Printf("A: inserting %02x after %02x, before %02x, at %d\n",
				//	ndx, curNdx, nextNdx, slotNbr)

				// first insert the index ---------------------------
				var leftNdx []byte
				if slotNbr > 0 {
					leftNdx = append(leftNdx, t32.indices[0:slotNbr]...)
				}
				rightNdx := t32.indices[slotNbr:]
				//fmt.Printf("%s + %02x + %s => ",
				//	dumpIndices(&leftNdx),
				//	ndx,
				//	dumpIndices(&rightNdx))
				leftNdx = append(leftNdx, ndx)
				leftNdx = append(leftNdx, rightNdx...)

				//fmt.Printf("%s\n", dumpIndices(&leftNdx))
				t32.indices = leftNdx // FOO

				// then insert the entry ----------------------------
				// WORKING HERE
				var leftSlots []*Entry32
				if slotNbr > 0 {
					// XXX PANICS
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
				// fmt.Printf("B: inserting %02x before %02x at %d\n",
				//	ndx, curNdx, slotNbr)

				// first insert the index ---------------------------
				var leftNdx []byte
				if slotNbr > 0 {
					leftNdx = append(leftNdx, t32.indices[0:slotNbr]...)
				}
				rightNdx := t32.indices[slotNbr:]
				// fmt.Printf("%s + %02x + %s\n",
				//	dumpIndices(leftNdx), ndx, dumpIndices(rightNdx))
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
				// fmt.Printf("C: appending %02x after %02x\n", ndx, curNdx)
				t32.indices = append(t32.indices, ndx)
				slotNbr = curSize
				// insert entry -------------------------------------
				t32.slots = append(t32.slots, entry)

			} else {
				// insert index -------------------------------------
				leftNdx := (t32.indices)[0:i]
				leftNdx = append(leftNdx, ndx)
				leftNdx = append(leftNdx, curNdx)
				t32.indices = leftNdx // FOO
				slotNbr = curSize - 1
				// fmt.Printf("D: prepended %02x before %02x at %d\n",
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
	//newSize := uint(len(t32.indices))
	//fmt.Printf("  inserted 0x%02x at %d/%d\n", ndx, slotNbr, newSize)
	//fmt.Printf("%s\n", dumpIndices(t32.indices))
	return
} // GEEP

// hc is the hashcode shifted the number of bits appropriate for the
// depth.  If it is necessary to recurse, the depth is increased by
//
// XXX OBSOLETE
//
//func (t32 *Table32) Insert(hc uint32, depth uint, k Key32I, v interface{}) (
//	err error) {
//
//	fmt.Printf("XXX CALL TO OBSOLETE FUNCTION t32.Insert\n")
//
//	// Mask off the first level mask bits, and see what we have there
//	where := byte(hc & LEVEL_MASK32)
//
//	p := t32.slots[where]
//	if p == nil {
//		// empty slot, so put the entry there
//		var leaf *Leaf32
//		leaf, err = NewLeaf32(k, v)
//		if err == nil {
//			var e *Entry32
//			e, err = NewEntry32(where, leaf)
//			if err == nil {
//				t32.slots[where] = e
//			}
//		}
//	} else if p.Node.IsLeaf() {
//
//		// XXX STUB
//		fmt.Printf("should be replacing leaf at slot %d\n", where)
//
//	} else {
//		// it's not a leaf, so try to insert
//		node := p.Node
//		nextT := node.(*Table32)
//		hc <<= W32 // shift appropriate number of bits
//		err = nextT.Insert(hc, depth+1, k, v)
//
//	}
//
//	// XXX STUB
//	return
//}
//// END OBSOLETE

func (t32 *Table32) IsLeaf() bool { return false }

// ==================================================================

type Table64 struct {
	bitmap uint64
	slots  []Entry64I // nil or a key-value pair or a pointer to a table
}

func (t64 *Table64) IsLeaf() bool { return false }
