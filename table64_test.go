package hamt_go

// hamt_go/table64_test.go

import (
	"bytes"
	"fmt"
	xr "github.com/jddixon/xlattice_go/rnglib"
	. "launchpad.net/gocheck"
)

var _ = fmt.Print

// xxx comments out the test
func (s *XLSuite) xxxTestTable64Ctor(c *C) {
	rng := xr.MakeSimpleRNG()
	depth := uint(rng.Intn(7)) // max of 6, because 6*5 = 30 < 64
	t64, err := NewTable64(depth)
	c.Assert(err, IsNil)
	c.Assert(t64, NotNil)
	c.Assert(t64.GetDepth(), Equals, depth)

	c.Assert(t64.slots, IsNil)
}

func (s *XLSuite) xxxTestT64DepthZeroInserts(c *C) {

	var (
		bitmap, flag, idx, mask uint64
		pos                     uint
	)
	rng := xr.MakeSimpleRNG()
	perm := rng.Perm(64) // a random permutation of [0..64)
	depth := uint(0)     // COULD VARY DEPTH

	t64, err := NewTable64(depth)
	c.Assert(err, IsNil)
	c.Assert(t64, NotNil)
	c.Assert(t64.GetDepth(), Equals, depth)

	keys := make([][]byte, 64)
	indices := make([]byte, 64)

	for i := uint(0); i < 64; i++ {
		ndx := byte(perm[i])
		indices[i] = ndx

		key := make([]byte, 64)
		key[0] = ndx // all the rest are zeroes
		keys[i] = key

		key64, err := NewBytes64Key(key)
		c.Assert(err, IsNil)
		c.Assert(key64, NotNil)
		hc, err := key64.Hashcode64()
		c.Assert(err, IsNil)
		c.Assert(hc, Equals, uint64(ndx))

		_, err = t64.findEntry(hc, 0, key64)
		c.Assert(err, Equals, NotFound)

		leaf, err := NewLeaf64(key64, &key)
		c.Assert(err, IsNil)
		c.Assert(leaf, NotNil)
		c.Assert(leaf.IsLeaf(), Equals, true)

		e, err := NewEntry64(ndx, leaf)
		c.Assert(err, IsNil)
		c.Assert(e, NotNil)
		c.Assert(e.GetIndex(), Equals, ndx)
		c.Assert(e.Node.IsLeaf(), Equals, true)
		c.Assert(e.GetIndex(), Equals, ndx)

		slotNbr, err := t64.insertEntry(hc, depth, e)
		c.Assert(err, IsNil)
		c.Assert(0 <= slotNbr && slotNbr < 64, Equals, true)

		// insert the value into the hash slice in such a way as
		// to maintain order
		idx = (hc >> (depth * W64)) & LEVEL_MASK64
		c.Assert(idx, Equals, hc) // hc is restricted to that range

		flag = 1 << (idx + 1)
		mask = flag - 1
		pos = BitCount64(bitmap & mask)
		occupied := uint64(1 << idx)
		bitmap |= uint64(occupied)

		// fmt.Printf("%02d: hc %02x, idx %02x, mask 0x%08x, bitmap 0x%08x, pos %02d slotNbr %02d\n\n",
		//	i, hc, idx, mask, bitmap, pos, slotNbr)

		c.Assert(t64.bitmap, Equals, bitmap)

		c.Assert(uint(pos), Equals, slotNbr)

		v, err := t64.findEntry(hc, 0, key64)
		c.Assert(err, IsNil)
		vBytes := v.(*[]byte)
		c.Assert(bytes.Equal(*vBytes, key), Equals, true)

	}
	// verify that the order of entries in the slots is as expected
	c.Assert(len(t64.indices), Equals, 64)
	c.Assert(len(t64.slots), Equals, 64)
	for i := uint(0); i < 64; i++ {
		idx := t64.indices[i]
		entry := t64.slots[i]
		c.Assert(entry.GetIndex(), Equals, idx)
		node := entry.Node
		c.Assert(node.IsLeaf(), Equals, true)
		leaf := node.(*Leaf64)
		key64 := leaf.Key
		hc, err := key64.Hashcode64()
		c.Assert(err, IsNil)
		c.Assert(hc&LEVEL_MASK64, Equals, uint64(idx))
		value := leaf.Value.(*[]byte)

		keyBytes := key64.(*Bytes64Key)
		c.Assert(bytes.Equal((*keyBytes).Slice, *value), Equals, true)
	}
	// remove each key, then verify that it is in fact gone
	c.Assert(len(t64.indices), Equals, 64)
	c.Assert(len(t64.slots), Equals, 64)
	for i := uint(0); i < 64; i++ {
		idx := indices[i]
		key := keys[i]

		// verify it is present -------------------------------------
		//fmt.Printf("%d VERIFYING PRESENT BEFORE DELETE: idx %02x\n", i, idx)
		key64, err := NewBytes64Key(key)
		c.Assert(err, IsNil)
		c.Assert(key64, NotNil)
		hc, err := key64.Hashcode64()
		c.Assert(err, IsNil)
		v, err := t64.findEntry(hc, 0, key64)
		c.Assert(err, IsNil)
		c.Assert(v, NotNil)
		vAsKey := v.(*[]byte)
		c.Assert(bytes.Equal(*vAsKey, key), Equals, true)

		// delete it ------------------------------------------------
		// depth is zero, so hc unshifted
		//fmt.Printf("  %d DELETING: idx %02x\n", i, idx)
		err = t64.deleteEntry(hc, 0, key64)
		c.Assert(err, IsNil)

		// verify that it is gone -----------------------------------
		//fmt.Printf("  %d CHECKING NOT PRESENT AFTER DELETE: idx %02x\n", i, idx)
		v, err = t64.findEntry(hc, 0, key64)

		c.Assert(err, Equals, NotFound)
		c.Assert(v, IsNil)

		_ = idx // DEBUG
	}
}

// Insert a series of entries, each of which should replace a leaf with
// a table.

func (s *XLSuite) xxxTestT64EntrySplittingInserts(c *C) {
	rng := xr.MakeSimpleRNG()
	perm := rng.Perm(64) // a random permutation of [0..64)
	depth := uint(0)

	t64, err := NewTable64(depth)
	c.Assert(err, IsNil)
	c.Assert(t64, NotNil)
	c.Assert(t64.GetDepth(), Equals, depth)

	const KEY_COUNT = 8

	keys := make([][]byte, KEY_COUNT)
	key64s := make([]*Bytes64Key, KEY_COUNT)
	hashcodes := make([]uint64, KEY_COUNT)
	values := make([]interface{}, KEY_COUNT)

	// Build KEY_COUNT keys of length 64, each with i+1 non-zero bytes on the
	// left, and zero bytes on the right.  Each key duplicates the
	// previous key except that a non-zero byte is introduced in the
	// i-th position.
	for i := uint(0); i < KEY_COUNT; i++ {
		key := make([]byte, 64)
		var j uint
		if i > 0 {
			lastKey := keys[i-1]
			for j < i {
				key[j] = lastKey[j]
				j++
			}
		}
		key[i] = byte(perm[i] + 1)
		keys[i] = key

		key64, err := NewBytes64Key(key)
		c.Assert(err, IsNil)
		c.Assert(key64, NotNil)
		key64s[i] = key64

		values[i] = &key

		hc, err := key64.Hashcode64()
		c.Assert(err, IsNil)
		hashcodes[i] = hc

	}
	// fmt.Printf("\nINSERTION LOOP\n")
	for i := uint(0); i < KEY_COUNT; i++ {
		//fmt.Printf("\nINSERTING KEY %d: %s\n", i, dumpByteSlice(keys[i]))
		hc := hashcodes[i]
		ndx := byte(hc & LEVEL_MASK64) // depth 0, so no shift

		// expect that no entry with this key can be found ----------
		key64 := key64s[i]
		_, err = t64.findEntry(hc, 0, key64)
		c.Assert(err, Equals, NotFound)

		// insert the entry -----------------------------------------
		leaf, err := NewLeaf64(key64, values[i])
		c.Assert(err, IsNil)
		c.Assert(leaf, NotNil)
		c.Assert(leaf.IsLeaf(), Equals, true)

		e, err := NewEntry64(ndx, leaf)
		c.Assert(err, IsNil)
		c.Assert(e, NotNil)
		c.Assert(e.GetIndex(), Equals, ndx)
		c.Assert(e.Node.IsLeaf(), Equals, true)
		c.Assert(e.GetIndex(), Equals, ndx)

		slotNbr, err := t64.insertEntry(hc, depth, e)
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
		_, err = t64.findEntry(hc, 0, key64)
		c.Assert(err, IsNil) // FAILS XXX
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
		_, err = t64.findEntry(hc, 0, key64)
		c.Assert(err, IsNil)
		//fmt.Printf("    key %2d is present before deletion\n", i) // DEBUG

		// delete the entry -----------------------------------------
		err = t64.deleteEntry(hc, 0, key64)
		c.Assert(err, IsNil)
		//fmt.Printf("    key %2d has been deleted\n", i) // DEBUG

		// confirm that it is gone ----------------------------------
		_, err = t64.findEntry(hc, 0, key64)
		c.Assert(err, Equals, NotFound)
		//fmt.Printf("    key %2d gone after deletion\n\n", i) // DEBUG
	}
}

// Insert a series of randomly selected entries, some of which may replace
// a leaf with a table.

func (s *XLSuite) xxxTestT64InsertsOfRandomishValues(c *C) {
	rng := xr.MakeSimpleRNG()
	depth := uint(0)

	t64, err := NewTable64(depth)
	c.Assert(err, IsNil)
	c.Assert(t64, NotNil)
	c.Assert(t64.GetDepth(), Equals, depth)

	const KEY_COUNT = 16 * 1024

	keys := make([][]byte, KEY_COUNT)
	key64s := make([]*Bytes64Key, KEY_COUNT)
	hashcodes := make([]uint64, KEY_COUNT)
	values := make([]interface{}, KEY_COUNT)
	hcMap := make(map[uint64]bool)

	// Build KEY_COUNT keys of length 64.
	for i := uint(0); i < KEY_COUNT; i++ {
		var hc uint64
		key := make([]byte, 64)
		for {
			rng.NextBytes(key) // fill with quasi-random values
			keys[i] = key

			key64, err := NewBytes64Key(key)
			c.Assert(err, IsNil)
			c.Assert(key64, NotNil)
			key64s[i] = key64

			hc, err = key64.Hashcode64()
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
		//fmt.Printf("\nINSERTING KEY %d: %s\n", i, dumpByteSlice(keys[i]))
		hc := hashcodes[i]
		ndx := byte(hc & LEVEL_MASK64) // depth 0, so no shift

		// expect that no entry with this key can be found ----------
		key64 := key64s[i]
		_, err = t64.findEntry(hc, 0, key64)
		c.Assert(err, Equals, NotFound)

		// insert the entry -----------------------------------------
		leaf, err := NewLeaf64(key64, values[i])
		c.Assert(err, IsNil)
		c.Assert(leaf, NotNil)
		c.Assert(leaf.IsLeaf(), Equals, true)

		e, err := NewEntry64(ndx, leaf)
		c.Assert(err, IsNil)
		c.Assert(e, NotNil)
		c.Assert(e.GetIndex(), Equals, ndx)
		c.Assert(e.Node.IsLeaf(), Equals, true)
		c.Assert(e.GetIndex(), Equals, ndx)

		slotNbr, err := t64.insertEntry(hc, depth, e)
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
		_, err = t64.findEntry(hc, 0, key64)
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
		_, err = t64.findEntry(hc, 0, key64)
		c.Assert(err, IsNil)
		//fmt.Printf("    key %2d is present before deletion\n", i) // DEBUG

		// delete the entry -----------------------------------------
		err = t64.deleteEntry(hc, 0, key64)
		c.Assert(err, IsNil)
		//fmt.Printf("    key %2d has been deleted\n", i) // DEBUG

		// confirm that it is gone ----------------------------------
		_, err = t64.findEntry(hc, 0, key64)
		c.Assert(err, Equals, NotFound)
		//fmt.Printf("    key %2d gone after deletion\n\n", i) // DEBUG
	}
}
