package hamt_go

// hamt_go/table_test.go

import (
	"bytes"
	"fmt"
	xr "github.com/jddixon/xlattice_go/rnglib"
	. "launchpad.net/gocheck"
)

var _ = fmt.Print

func (s *XLSuite) TestTable32Ctor(c *C) {
	rng := xr.MakeSimpleRNG()
	depth := uint(rng.Intn(7)) // max of 6, because 6*5 = 30 < 32
	t32, err := NewTable32(depth)
	c.Assert(err, IsNil)
	c.Assert(t32, NotNil)
	c.Assert(t32.GetDepth(), Equals, depth)

	c.Assert(t32.slots, IsNil)
}

// XXX IN EFFECT, COMMENTED OUT
func (s *XLSuite) xxxTestDepthZeroInserts(c *C) {

	var (
		bitmap, flag, idx, mask uint32
		pos                     uint
	)
	rng := xr.MakeSimpleRNG()
	perm := rng.Perm(32) // a random permutation of [0..32)
	depth := uint(0)     // COULD VARY DEPTH

	t32, err := NewTable32(depth)
	c.Assert(err, IsNil)
	c.Assert(t32, NotNil)
	c.Assert(t32.GetDepth(), Equals, depth)

	keys := make([][]byte, 32)
	indices := make([]byte, 32)

	for i := uint(0); i < 32; i++ {
		ndx := byte(perm[i])
		indices[i] = ndx

		key := make([]byte, 32)
		key[0] = ndx // all the rest are zeroes
		keys[i] = key

		key32, err := NewBytes32Key(key)
		c.Assert(err, IsNil)
		c.Assert(key32, NotNil)
		hc, err := key32.Hashcode32()
		c.Assert(err, IsNil)
		c.Assert(hc, Equals, uint32(ndx))

		_, err = t32.findEntry(hc, 0, key32)
		c.Assert(err, Equals, NotFound)

		leaf, err := NewLeaf32(key32, &key)
		c.Assert(err, IsNil)
		c.Assert(leaf, NotNil)
		c.Assert(leaf.IsLeaf(), Equals, true)

		e, err := NewEntry32(ndx, leaf)
		c.Assert(err, IsNil)
		c.Assert(e, NotNil)
		c.Assert(e.GetIndex(), Equals, ndx)
		c.Assert(e.Node.IsLeaf(), Equals, true)
		c.Assert(e.GetIndex(), Equals, ndx)

		slotNbr, err := t32.insertEntry(hc, depth, e)
		c.Assert(err, IsNil)
		c.Assert(0 <= slotNbr && slotNbr < 32, Equals, true)

		// insert the value into the hash slice in such a way as
		// to maintain order
		idx = (hc >> (depth * W32)) & 0x1f
		c.Assert(idx, Equals, hc) // hc is restricted to that range

		flag = 1 << (idx + 1)
		mask = flag - 1
		pos = BitCount32(bitmap & mask)
		occupied := uint32(1 << idx)
		bitmap |= uint32(occupied)

		// fmt.Printf("%02d: hc %02x, idx %02x, mask 0x%08x, bitmap 0x%08x, pos %02d slotNbr %02d\n\n",
		//	i, hc, idx, mask, bitmap, pos, slotNbr)

		c.Assert(t32.bitmap, Equals, bitmap)

		c.Assert(uint(pos), Equals, slotNbr)

		v, err := t32.findEntry(hc, 0, key32)
		c.Assert(err, IsNil)
		vBytes := v.(*[]byte)
		c.Assert(bytes.Equal(*vBytes, key), Equals, true)

	}
	// verify that the order of entries in the slots is as expected
	c.Assert(len(t32.indices), Equals, 32)
	c.Assert(len(t32.slots), Equals, 32)
	for i := uint(0); i < 32; i++ {
		idx := t32.indices[i]
		entry := t32.slots[i]
		c.Assert(entry.GetIndex(), Equals, idx)
		node := entry.Node
		c.Assert(node.IsLeaf(), Equals, true)
		leaf := node.(*Leaf32)
		key32 := leaf.Key
		hc, err := key32.Hashcode32()
		c.Assert(err, IsNil)
		c.Assert(hc&LEVEL_MASK32, Equals, uint32(idx))
		value := leaf.Value.(*[]byte)

		keyBytes := key32.(*Bytes32Key)
		c.Assert(bytes.Equal((*keyBytes).Slice, *value), Equals, true)
	}
	// remove each key, then verify that it is in fact gone
	c.Assert(len(t32.indices), Equals, 32)
	c.Assert(len(t32.slots), Equals, 32)
	for i := uint(0); i < 32; i++ {
		idx := indices[i]
		key := keys[i]

		// verify it is present -------------------------------------
		//fmt.Printf("%d VERIFYING PRESENT BEFORE DELETE: idx %02x\n", i, idx)
		key32, err := NewBytes32Key(key)
		c.Assert(err, IsNil)
		c.Assert(key32, NotNil)
		hc, err := key32.Hashcode32()
		c.Assert(err, IsNil)
		v, err := t32.findEntry(hc, 0, key32)
		c.Assert(err, IsNil)
		c.Assert(v, NotNil)
		vAsKey := v.(*[]byte)
		c.Assert(bytes.Equal(*vAsKey, key), Equals, true)

		// delete it ------------------------------------------------
		// depth is zero, so hc unshifted
		//fmt.Printf("  %d DELETING: idx %02x\n", i, idx)
		err = t32.deleteEntry(hc, 0, key32)
		c.Assert(err, IsNil)

		// verify that it is gone -----------------------------------
		//fmt.Printf("  %d CHECKING NOT PRESENT AFTER DELETE: idx %02x\n", i, idx)
		v, err = t32.findEntry(hc, 0, key32)

		c.Assert(err, Equals, NotFound)
		c.Assert(v, IsNil)

		_ = idx // DEBUG
	}
}

// Insert a series of entries, each of which should replace a leaf with
// a table.

func (s *XLSuite) TestEntrySplittingInserts(c *C) {
	rng := xr.MakeSimpleRNG()
	perm := rng.Perm(32) // a random permutation of [0..32)
	depth := uint(0)

	t32, err := NewTable32(depth)
	c.Assert(err, IsNil)
	c.Assert(t32, NotNil)
	c.Assert(t32.GetDepth(), Equals, depth)

	keys := make([][]byte, 32)
	key32s := make([]*Bytes32Key, 32)
	indices := make([]byte, 32)
	hashcodes := make([]uint32, 32)
	values := make([]interface{}, 32)

	// Build 32 keys of length 32, each with i+1 non-zero bytes on the
	// left, and zero bytes on the right.  Each key duplicates the
	// previous key except that a non-zero byte is introduced in the
	// i-th position.
	for i := uint(0); i < 32; i++ {
		key := make([]byte, 32)
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

		key32, err := NewBytes32Key(key)
		c.Assert(err, IsNil)
		c.Assert(key32, NotNil)
		key32s[i] = key32

		values[i] = &key

		hc, err := key32.Hashcode32()
		c.Assert(err, IsNil)
		hashcodes[i] = hc

	}
	fmt.Printf("\nINSERTION LOOP\n")
	for i := uint(0); i < 32; i++ {
		fmt.Printf("\nINSERTING KEY %d: %s\n", i, dumpByteSlice(keys[i]))
		hc := hashcodes[i]
		ndx := byte(hc & 0x1f) // depth 0, so no shift

		// expect that no entry with this key can be found ----------
		key32 := key32s[i]
		_, err = t32.findEntry(hc, 0, key32)
		c.Assert(err, Equals, NotFound)

		// insert the entry -----------------------------------------
		leaf, err := NewLeaf32(key32, values[i])
		c.Assert(err, IsNil)
		c.Assert(leaf, NotNil)
		c.Assert(leaf.IsLeaf(), Equals, true)

		e, err := NewEntry32(ndx, leaf)
		c.Assert(err, IsNil)
		c.Assert(e, NotNil)
		c.Assert(e.GetIndex(), Equals, ndx)
		c.Assert(e.Node.IsLeaf(), Equals, true)
		c.Assert(e.GetIndex(), Equals, ndx)

		slotNbr, err := t32.insertEntry(hc, depth, e)
		c.Assert(err, IsNil)
		// in this test, only one entry at the top level, so slotNbr always zero
		c.Assert(slotNbr, Equals, uint(0))

		// DEBUG
		fmt.Printf("inserted i = %2d, hc 0x%x, ndx 0x%02x; slotNbr => %d\n",
			i, hc, ndx, slotNbr)
		// END

		// confirm that the new entry is now present ----------------
		// DEBUG
		fmt.Printf("--- verifying new entry is present after insertion -----\n")
		// END
		_, err = t32.findEntry(hc, 0, key32)
		c.Assert(err, IsNil) // FAILS XXX
	}
	fmt.Println("\nDELETION LOOP") // DEBUG
	for i := uint(0); i < 32; i++ {
		hc := hashcodes[i]

		// DEBUG
		if err != nil {
			fmt.Printf("  deleting key %2d, hc 0x%x\n", i, hc)
		}
		// END

		key32 := key32s[i]
		// confirm again that the entry is present ------------------
		_, err = t32.findEntry(hc, 0, key32)
		c.Assert(err, IsNil)
		fmt.Printf("    key %2d is present before deletion\n", i) // DEBUG

		// delete the entry -----------------------------------------
		err = t32.deleteEntry(hc, 0, key32)
		c.Assert(err, IsNil)                            // FAILS XXX :-)
		fmt.Printf("    key %2d has been deleted\n", i) // DEBUG

		// confirm that it is gone ----------------------------------
		_, err = t32.findEntry(hc, 0, key32)
		c.Assert(err, Equals, NotFound)
		fmt.Printf("    key %2d gone after deletion\n\n", i) // DEBUG
	}
	_ = indices // DEBUG
}
