package hamt_go

// hamt_go/table_test.go

import (
	"bytes"
	"fmt"
	xr "github.com/jddixon/xlattice_go/rnglib"
	. "launchpad.net/gocheck"
)

var _ = fmt.Print

func (s *XLSuite) TestTableCtor(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("TEST_TABLE_CTOR")
	}
	s.doTestTableCtor(c, uint(4))
	s.doTestTableCtor(c, uint(5))
	s.doTestTableCtor(c, uint(6))
}
func (s *XLSuite) doTestTableCtor(c *C, w uint) {

	rng := xr.MakeSimpleRNG()
	depth := 1 + uint(rng.Intn(7))
	t := uint(0)
	table, err := NewTable(depth, w, t)
	c.Assert(err, IsNil)
	c.Assert(table, NotNil)
	c.Assert(table.GetDepth(), Equals, depth)

	c.Assert(table.slots, IsNil)
}

// ------------------------------------------------------------------
// XXX This is now somewhat nonsensical: the root table is of a
// different type.

func (s *XLSuite) TestTableDepthZeroInserts(c *C) {

	if VERBOSITY > 0 {
		fmt.Println("TEST_TABLE_DEPTH_ZERO_INSERTS")
	}
	s.doTestTableDepthZeroInserts(c, 4, 0)
	s.doTestTableDepthZeroInserts(c, 5, 0)
	s.doTestTableDepthZeroInserts(c, 6, 0)

}
func (s *XLSuite) doTestTableDepthZeroInserts(c *C, w, t uint) {
	var (
		bitmap, flag, idx, mask uint64
		pos                     uint
	)
	depth := uint(1)

	table, err := NewTable(depth, w, t)
	c.Assert(err, IsNil)
	c.Assert(table, NotNil)
	c.Assert(table.GetDepth(), Equals, depth)

	c.Assert(table.w, Equals, w)
	c.Assert(table.t, Equals, t)
	flag = uint64(1)
	flag <<= (t + w)
	expectedMask := flag - 1
	c.Assert(table.mask, Equals, expectedMask)

	SLOT_COUNT := table.maxSlots

	// DEBUG
	//fmt.Printf("doTest: w = %d, t = %d, mask = 0x%x, maxSlots = %d\n",
	//	w, t, table.mask, SLOT_COUNT)
	// END

	rng := xr.MakeSimpleRNG()
	perm := rng.Perm(int(SLOT_COUNT)) // a random permutation of [0..SLOT_COUNT)
	rawKeys := make([][]byte, SLOT_COUNT)
	indices := make([]byte, SLOT_COUNT)

	for i := uint(0); i < SLOT_COUNT; i++ {
		ndx := byte(perm[i])
		indices[i] = ndx

		rawKey := make([]byte, SLOT_COUNT)
		rawKey[0] = ndx // all the rest are zeroes
		rawKeys[i] = rawKey

		key64, err := NewBytesKey(rawKey)
		c.Assert(err, IsNil)
		c.Assert(key64, NotNil)
		hc, err := key64.Hashcode()
		c.Assert(err, IsNil)
		c.Assert(hc, Equals, uint64(ndx))

		_, err = table.findEntry(hc, depth, key64)
		c.Assert(err, Equals, NotFound)

		leaf, err := NewLeaf(key64, &rawKey)
		c.Assert(err, IsNil)
		c.Assert(leaf, NotNil)
		c.Assert(leaf.IsLeaf(), Equals, true)

		e, err := NewEntry(ndx, leaf)
		c.Assert(err, IsNil)
		c.Assert(e, NotNil)
		c.Assert(e.GetIndex(), Equals, ndx)
		c.Assert(e.Node.IsLeaf(), Equals, true)
		c.Assert(e.GetIndex(), Equals, ndx)

		slotNbr, err := table.insertEntry(hc, depth, e)
		c.Assert(err, IsNil)
		c.Assert(0 <= slotNbr && slotNbr < SLOT_COUNT, Equals, true)

		// insert the value into the hash slice in such a way as
		// to maintain order
		idx = hc & table.mask
		flag = 1 << idx
		mask = flag - 1
		pos = BitCount64(bitmap & mask)
		occupied := uint64(1 << idx)
		bitmap |= occupied

		// DEBUG
		//fmt.Printf("%02d: hc %02x, idx %02x, mask 0x%08x, bitmap 0x%08x, pos %02d slotNbr %02d\n\n",
		//	i, hc, idx, mask, bitmap, pos, slotNbr)
		// END

		c.Assert(table.bitmap, Equals, bitmap)
		c.Assert(uint(pos), Equals, slotNbr)

		v, err := table.findEntry(hc, depth, key64)
		c.Assert(err, IsNil)
		vBytes := v.(*[]byte)
		c.Assert(bytes.Equal(*vBytes, rawKey), Equals, true)

	}
	// verify that the order of entries in the slots is as expected
	c.Assert(uint(len(table.indices)), Equals, SLOT_COUNT)
	c.Assert(uint(len(table.slots)), Equals, SLOT_COUNT)
	for i := uint(0); i < SLOT_COUNT; i++ {
		idx := table.indices[i]
		entry := table.slots[i]
		c.Assert(entry.GetIndex(), Equals, idx)
		node := entry.Node
		c.Assert(node.IsLeaf(), Equals, true)
		leaf := node.(*Leaf)
		key64 := leaf.Key
		hc, err := key64.Hashcode()
		c.Assert(err, IsNil)
		c.Assert(hc&table.mask, Equals, uint64(idx))
		value := leaf.Value.(*[]byte)

		keyBytes := key64.(*BytesKey)
		c.Assert(bytes.Equal((*keyBytes).Slice, *value), Equals, true)
	}
	// remove each key, then verify that it is in fact gone
	c.Assert(uint(len(table.indices)), Equals, SLOT_COUNT)
	c.Assert(uint(len(table.slots)), Equals, SLOT_COUNT)
	for i := uint(0); i < SLOT_COUNT; i++ {
		idx := indices[i]
		key := rawKeys[i]

		// verify it is present -------------------------------------
		//fmt.Printf("%d VERIFYING PRESENT BEFORE DELETE: idx %02x\n", i, idx)
		key64, err := NewBytesKey(key)
		c.Assert(err, IsNil)
		c.Assert(key64, NotNil)
		hc, err := key64.Hashcode()
		c.Assert(err, IsNil)
		v, err := table.findEntry(hc, depth, key64)
		c.Assert(err, IsNil)
		c.Assert(v, NotNil)
		vAsKey := v.(*[]byte)
		c.Assert(bytes.Equal(*vAsKey, key), Equals, true)

		// delete it ------------------------------------------------
		// depth is zero, so hc unshifted
		//fmt.Printf("  %d DELETING: idx %02x\n", i, idx)
		err = table.deleteEntry(hc, depth, key64)
		c.Assert(err, IsNil)

		// verify that it is gone -----------------------------------
		//fmt.Printf("  %d CHECKING NOT PRESENT AFTER DELETE: idx %02x\n", i, idx)
		v, err = table.findEntry(hc, depth, key64)

		c.Assert(err, Equals, NotFound)
		c.Assert(v, IsNil)

		_ = idx // DEBUG
	}
}

// ------------------------------------------------------------------

// Insert a series of entries, each of which should replace a leaf with
// a table.

func (s *XLSuite) TestEntrySplittingInserts(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("TEST_TABLE_ENTRY_SPLITTING_INSERTS")
	}
	rng := xr.MakeSimpleRNG()

	s.doTestEntrySplittingInserts(c, rng, uint(5))
}

func (s *XLSuite) doTestEntrySplittingInserts(c *C, rng *xr.PRNG, w uint) {

	depth := uint(1)
	t := uint(0)

	table, err := NewTable(depth, w, t)

	// DEBUG
	var depthNTables []*Table
	tSoFar := uint(0)
	// END

	c.Assert(err, IsNil)
	c.Assert(table, NotNil)
	c.Assert(table.GetDepth(), Equals, depth)

	c.Assert(table.w, Equals, w)
	c.Assert(table.t, Equals, t)

	flag := uint64(1)
	flag <<= (t + w)
	expectedMask := flag - 1
	c.Assert(table.mask, Equals, expectedMask)
	c.Assert(table.maxDepth, Equals, 64/w)

	rawKeys := s.makePermutedKeys(rng, w)
	KEY_COUNT := uint(len(rawKeys))

	//fmt.Printf("KEY_COUNT = %d\n", KEY_COUNT) // DEBUG

	key64s := make([]*BytesKey, KEY_COUNT)
	hashcodes := make([]uint64, KEY_COUNT)
	values := make([]interface{}, KEY_COUNT)

	for i := uint(0); i < KEY_COUNT; i++ {
		key := rawKeys[i]

		key64, err := NewBytesKey(key)
		c.Assert(err, IsNil)
		c.Assert(key64, NotNil)
		key64s[i] = key64

		values[i] = &key

		hc, err := key64.Hashcode()
		c.Assert(err, IsNil)
		hashcodes[i] = hc

	}
	c.Assert(table.GetTableCount(), Equals, uint(1))

	// fmt.Printf("\nINSERTION LOOP\n")
	for i := uint(0); i < KEY_COUNT; i++ {
		//fmt.Printf("\nINSERTING KEY %d: %s\n", i, dumpByteSlice(rawKeys[i]))
		hc := hashcodes[i]
		ndx := byte(hc & table.mask) // depth 0, so no shift

		// expect that no entry with this key can be found ----------
		key64 := key64s[i]
		_, err = table.findEntry(hc, depth, key64)
		c.Assert(err, Equals, NotFound)

		// insert the entry -----------------------------------------
		leaf, err := NewLeaf(key64, values[i])
		c.Assert(err, IsNil)
		c.Assert(leaf, NotNil)
		c.Assert(leaf.IsLeaf(), Equals, true)

		e, err := NewEntry(ndx, leaf)
		c.Assert(err, IsNil)
		c.Assert(e, NotNil)
		c.Assert(e.GetIndex(), Equals, ndx)
		c.Assert(e.Node.IsLeaf(), Equals, true)
		c.Assert(e.GetIndex(), Equals, ndx)

		slotNbr, err := table.insertEntry(hc, depth, e)
		c.Assert(err, IsNil)
		// in this test, only one entry at the top level, so slotNbr always zero
		c.Assert(slotNbr, Equals, uint(0))

		// DEBUG
		//fmt.Printf("inserted i = %2d, hc 0x%x, ndx 0x%02x; slotNbr => %d\n",
		//	i, hc, ndx, slotNbr)
		// END

		// confirm that the new entry is now present ----------------
		// DEBUG
		//fmt.Printf("--- verifying new entry is present after insertion -----\n")
		// END
		_, err = table.findEntry(hc, depth, key64)
		c.Assert(err, IsNil)

		// DEBUG depthN -- WORKING HERE
		tCount := table.GetTableCount()
		fmt.Printf("  insertion %2d count %2d: %s\n",
			i, tCount, dumpByteSlice(rawKeys[i]))

		if tSoFar <= i {
			var dTable *Table
			if i == 0 {
				dTable = table
				depthNTables = append(depthNTables, dTable) // just a pointer
				tSoFar++
				fmt.Printf("  dTable %2d, depth %2d: %s\n",
					tSoFar,
					dTable.depth,
					dumpByteSlice(dTable.indices))
			} else {
				for tSoFar < tCount {
					var slot *Entry
					dTable = depthNTables[len(depthNTables)-1]
					slotCount := len(dTable.slots)
					if slotCount == 1 {
						slot = dTable.slots[0]
					} else {
						slot = dTable.slots[1]
					}
					c.Assert(slot.Node.IsLeaf(), Equals, false)
					dTable = slot.Node.(*Table)

					depthNTables = append(depthNTables, dTable)
					tSoFar++
					fmt.Printf("    dTable %2d: %s\n",
						tSoFar, dumpByteSlice(dTable.indices))
				}
			}
		}
		// END

		// c.Assert(table.GetTableCount(), Equals, i + 1)	// FAILS XXX
	}
	// DEBUG
	fmt.Println("DUMP OF DEPTH-N TABLES")
	for i := uint(0); i < uint(len(depthNTables)); i++ {
		lineNo := fmt.Sprintf("%2d ", i)
		tDump := dumpTable(lineNo, depthNTables[i])
		fmt.Println(tDump)
	}
	// END
	//fmt.Println("\nDELETION LOOP") // DEBUG
	for i := uint(0); i < KEY_COUNT; i++ {
		hc := hashcodes[i]

		// DEBUG
		//if err != nil {
		//	fmt.Printf("  deleting key %2d, hc 0x%x\n", i, hc)
		//}
		// END

		key64 := key64s[i]
		// confirm again that the entry is present ------------------
		_, err = table.findEntry(hc, depth, key64)
		c.Assert(err, IsNil)
		//fmt.Printf("    key %2d is present before deletion\n", i) // DEBUG

		// delete the entry -----------------------------------------
		err = table.deleteEntry(hc, depth, key64)
		c.Assert(err, IsNil)
		//fmt.Printf("    key %2d has been deleted\n", i) // DEBUG

		// confirm that it is gone ----------------------------------
		_, err = table.findEntry(hc, depth, key64)
		c.Assert(err, Equals, NotFound)
		//fmt.Printf("    key %2d gone after deletion\n\n", i) // DEBUG
	}
}

// ------------------------------------------------------------------

// Insert a series of randomly selected entries, some of which may replace
// a leaf with a table.  Run time with 1K entries 1 to 2 sec.  With 2K
// entries 3.3 to 3.4 sec.  8K entries, 13 sec.  32K entries, 50 sec.
// Without the debug statements, 32K entries, about 1.2 sec, 32us/entry.

func (s *XLSuite) TestTableInsertsOfRandomishValues(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("TEST_TABLE_INSERTS_OF_RANDOMISH_VALUES")
	}
	rng := xr.MakeSimpleRNG()
	depth := uint(1)
	w := uint(5)
	t := uint(0)

	table, err := NewTable(depth, w, t)
	c.Assert(err, IsNil)
	c.Assert(table, NotNil)
	c.Assert(table.GetDepth(), Equals, depth)

	const KEY_COUNT = 16 * 1024

	rawKeys := make([][]byte, KEY_COUNT)
	key64s := make([]*BytesKey, KEY_COUNT)
	hashcodes := make([]uint64, KEY_COUNT)
	values := make([]interface{}, KEY_COUNT)
	hcMap := make(map[uint64]bool)

	// Build KEY_COUNT rawKeys of length 32.
	for i := uint(0); i < KEY_COUNT; i++ {
		var hc uint64
		key := make([]byte, 32)
		for {
			rng.NextBytes(key) // fill with quasi-random values
			rawKeys[i] = key

			key64, err := NewBytesKey(key)
			c.Assert(err, IsNil)
			c.Assert(key64, NotNil)
			key64s[i] = key64

			hc, err = key64.Hashcode()
			c.Assert(err, IsNil)
			_, ok := hcMap[hc]
			if !ok {
				hcMap[hc] = true
				break
			}
		}
		values[i] = &key
		hashcodes[i] = hc

	}
	//fmt.Printf("\nINSERTION LOOP\n")
	for i := uint(0); i < KEY_COUNT; i++ {
		//fmt.Printf("\nINSERTING KEY %d: %s\n", i, dumpByteSlice(rawKeys[i]))
		hc := hashcodes[i]
		ndx := byte(hc & table.mask) // depth 0, so no shift

		// expect that no entry with this key can be found ----------
		key64 := key64s[i]
		_, err = table.findEntry(hc, depth, key64)
		c.Assert(err, Equals, NotFound)

		// insert the entry -----------------------------------------
		leaf, err := NewLeaf(key64, values[i])
		c.Assert(err, IsNil)
		c.Assert(leaf, NotNil)
		c.Assert(leaf.IsLeaf(), Equals, true)

		e, err := NewEntry(ndx, leaf)
		c.Assert(err, IsNil)
		c.Assert(e, NotNil)
		c.Assert(e.GetIndex(), Equals, ndx)
		c.Assert(e.Node.IsLeaf(), Equals, true)
		c.Assert(e.GetIndex(), Equals, ndx)

		slotNbr, err := table.insertEntry(hc, depth, e)
		_ = slotNbr // DEBUG
		c.Assert(err, IsNil)

		// DEBUG
		//fmt.Printf("inserted i = %2d, hc 0x%x, ndx 0x%02x; slotNbr => %d\n",
		//	i, hc, ndx, slotNbr)
		// END

		// confirm that the new entry is now present ----------------
		// DEBUG
		//fmt.Printf("--- verifying new entry is present after insertion -----\n")
		// END
		_, err = table.findEntry(hc, depth, key64)
		c.Assert(err, IsNil)
	}
	//fmt.Println("\nDELETION LOOP") // DEBUG
	for i := uint(0); i < KEY_COUNT; i++ {
		hc := hashcodes[i]

		// DEBUG
		//if err != nil {
		//	fmt.Printf("  deleting key %2d, hc 0x%x\n", i, hc)
		//}
		// END

		key64 := key64s[i]
		// confirm again that the entry is present ------------------
		_, err = table.findEntry(hc, depth, key64)
		c.Assert(err, IsNil)
		//fmt.Printf("    key %2d is present before deletion\n", i) // DEBUG

		// delete the entry -----------------------------------------
		err = table.deleteEntry(hc, depth, key64)
		c.Assert(err, IsNil)
		//fmt.Printf("    key %2d has been deleted\n", i) // DEBUG

		// confirm that it is gone ----------------------------------
		_, err = table.findEntry(hc, depth, key64)
		c.Assert(err, Equals, NotFound)
		//fmt.Printf("    key %2d gone after deletion\n\n", i) // DEBUG
	}
}
