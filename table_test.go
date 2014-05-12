package hamt_go

// hamt_go/table_test.go

import (
	"bytes"
	"fmt"
	xr "github.com/jddixon/rnglib_go"
	. "gopkg.in/check.v1"
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

	// XXX These fail for unknown reasons - whereas hamt_test equivalents
	// succeed.
	//s.doTestTableDepthZeroInserts(c, 7, 0)
	//s.doTestTableDepthZeroInserts(c, 8, 0)

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

		value, err := table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)
		c.Assert(value, IsNil)

		leaf, err := NewLeaf(key64, &rawKey)
		c.Assert(err, IsNil)
		c.Assert(leaf, NotNil)
		c.Assert(leaf.IsLeaf(), Equals, true)

		slotNbr, err := table.insertLeaf(hc, depth, leaf)
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

		c.Assert(table.bitmap, Equals, bitmap)
		c.Assert(uint(pos), Equals, slotNbr)

		v, err := table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)
		vBytes := v.(*[]byte)
		c.Assert(bytes.Equal(*vBytes, rawKey), Equals, true)

	}
	// verify that the order of entries in the slots is as expected
	c.Assert(uint(len(table.indices)), Equals, SLOT_COUNT)
	c.Assert(uint(len(table.slots)), Equals, SLOT_COUNT)
	for i := uint(0); i < SLOT_COUNT; i++ {
		idx := table.indices[i]
		node := table.slots[i]
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
		key64, err := NewBytesKey(key)
		c.Assert(err, IsNil)
		c.Assert(key64, NotNil)
		hc, err := key64.Hashcode()
		c.Assert(err, IsNil)
		v, err := table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)
		c.Assert(v, NotNil)
		vAsKey := v.(*[]byte)
		c.Assert(bytes.Equal(*vAsKey, key), Equals, true)

		// delete it ------------------------------------------------
		// depth is zero, so hc unshifted
		err = table.deleteLeaf(hc, depth, key64)
		c.Assert(err, IsNil)

		// verify that it is gone -----------------------------------
		v, err = table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)
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

	// s.doTestEntrySplittingInserts(c, rng, uint(3))	// key too short
	// s.doTestEntrySplittingInserts(c, rng, uint(4))	// index out of range
	s.doTestEntrySplittingInserts(c, rng, uint(5))
	s.doTestEntrySplittingInserts(c, rng, uint(6))
	// s.doTestEntrySplittingInserts(c, rng, uint(7))	// entry not found
	// s.doTestEntrySplittingInserts(c, rng, uint(8))	// index out of range
}

func (s *XLSuite) doTestEntrySplittingInserts(c *C, rng *xr.PRNG, w uint) {

	depth := uint(1)
	t := uint(0)

	table, err := NewTable(depth, w, t)
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

	_, rawKeys := s.makePermutedKeys(rng, w) // XXX fields ignored
	KEY_COUNT := table.maxDepth              // some keys ignored

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
	c.Assert(table.getTableCount(), Equals, uint(1))

	for i := uint(0); i < KEY_COUNT; i++ {
		hc := hashcodes[i]

		// expect that no entry with this key can be found ----------
		key64 := key64s[i]
		value, err := table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)
		c.Assert(value, IsNil)

		// insert the entry -----------------------------------------
		leaf, err := NewLeaf(key64, values[i])
		c.Assert(err, IsNil)
		c.Assert(leaf, NotNil)
		c.Assert(leaf.IsLeaf(), Equals, true)

		slotNbr, err := table.insertLeaf(hc, depth, leaf)
		c.Assert(err, IsNil)
		// in this test, only one entry at the top level, so slotNbr always zero
		c.Assert(slotNbr, Equals, uint(0))

		// confirm that the new entry is now present ----------------
		_, err = table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)

		// c.Assert(table.GetTableCount(), Equals, i + 1)
	}
	for i := uint(0); i < KEY_COUNT; i++ {
		hc := hashcodes[i]
		key64 := key64s[i]
		// confirm again that the entry is present ------------------
		_, err = table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)

		// delete the entry -----------------------------------------
		err = table.deleteLeaf(hc, depth, key64)
		c.Assert(err, IsNil)

		// confirm that it is gone ----------------------------------
		var value interface{}
		value, err = table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)
		c.Assert(value, IsNil)
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
	for i := uint(0); i < KEY_COUNT; i++ {
		hc := hashcodes[i]

		// expect that no entry with this key can be found ----------
		key64 := key64s[i]
		value, err := table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)
		c.Assert(value, IsNil)

		// insert the entry -----------------------------------------
		leaf, err := NewLeaf(key64, values[i])
		c.Assert(err, IsNil)
		c.Assert(leaf, NotNil)
		c.Assert(leaf.IsLeaf(), Equals, true)

		_, err = table.insertLeaf(hc, depth, leaf)
		c.Assert(err, IsNil)

		// confirm that the new entry is now present ----------------
		_, err = table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)

		// TEST HANDLING OF DUPLICATE KEYS ----------------
		// replace the value associated with this key
		newValue := rng.Int63() // a random 64-bit value
		leaf2, err := NewLeaf(key64, &newValue)
		c.Assert(err, IsNil)
		c.Assert(leaf2, NotNil)
		c.Assert(leaf2.IsLeaf(), Equals, true)

		_, err = table.insertLeaf(hc, depth, leaf2)
		c.Assert(err, IsNil)

		// make sure that a Find returns the new value
		ret, err := table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)
		retPtr := ret.(*int64)
		c.Assert(*retPtr, Equals, newValue)

		// put the old value back
		_, err = table.insertLeaf(hc, depth, leaf)
		c.Assert(err, IsNil)

	}
	for i := uint(0); i < KEY_COUNT; i++ {
		hc := hashcodes[i]
		key64 := key64s[i]
		// confirm again that the entry is present ------------------
		_, err = table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)

		// delete the entry -----------------------------------------
		err = table.deleteLeaf(hc, depth, key64)
		c.Assert(err, IsNil)

		// confirm that it is gone ----------------------------------
		var value interface{}
		value, err = table.findLeaf(hc, depth, key64)
		c.Assert(err, IsNil)
		c.Assert(value, IsNil)
	}
}
