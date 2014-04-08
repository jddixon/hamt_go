package hamt_go

// hamt_go/table_test.go

import (
	"bytes"
	// "encoding/binary"
	"fmt"
	xr "github.com/jddixon/xlattice_go/rnglib"
	. "launchpad.net/gocheck"
	// "strings"
	// "sync/atomic"
	//"unsafe"
)

var _ = fmt.Print

// XXX WARNING: this tests the algorithm, but does NOT test the
// implementation of the algorithm in table.go
//
// XXX OBSOLETE XXX
//func (s *XLSuite) insertHash(c *C, slice *[]byte, value byte) (where uint) {
//
//	curSize := uint(len(*slice))
//	c.Assert(where <= curSize, Equals, true)
//
//	if curSize == 0 {
//		*slice = append(*slice, value)
//	} else {
//		mySlice := *slice
//		inserted := false
//		var i uint
//		var curValue, nextValue byte
//		for i = 0; i < curSize-1; i++ {
//			curValue = mySlice[i]
//			if curValue < value {
//				nextValue = mySlice[i+1]
//				if nextValue < value {
//					// fmt.Printf("continuing: %02x after %02x, after %02x\n",
//					//	value, curValue, nextValue)
//					continue
//				}
//				c.Assert(value < nextValue, Equals, true)
//				where = i + 1
//				//fmt.Printf("A: inserting %02x after %02x, before %02x, at %d\n",
//				//	value, curValue, nextValue, where)
//				// do the insertion
//				var left []byte
//				if where > 0 {
//					left = append(left, mySlice[0:where]...)
//				}
//				right := mySlice[where:]
//				//fmt.Printf("%s + %02x + %s => ",
//				//	s.dumpSlice(&left),
//				//	value,
//				//	s.dumpSlice(&right))
//				left = append(left, value)
//				left = append(left, right...)
//
//				//fmt.Printf("%s\n", s.dumpSlice(&left))
//				*slice = left
//				inserted = true
//				break
//			} else {
//				c.Assert(value < curValue, Equals, true)
//				where = i
//				//fmt.Printf("B: inserting %02x before %02x at %d\n",
//				//	value, curValue, where)
//				// do the insertion
//				var left []byte
//				if where > 0 {
//					left = append(left, mySlice[0:where]...)
//				}
//				right := mySlice[where:]
//				// fmt.Printf("%s + %02x + %s\n",
//				//	s.dumpSlice(&left), value, s.dumpSlice(&right))
//				left = append(left, value)
//				left = append(left, right...)
//				*slice = left
//				inserted = true
//				break
//
//			}
//		}
//		if !inserted {
//			c.Assert(uint(i), Equals, curSize-1)
//			c.Assert(i, Equals, curSize-1)
//			curValue = (*slice)[i]
//			if curValue < value {
//				//fmt.Printf("C: appending %02x after %02x\n", value, curValue)
//				*slice = append(*slice, value)
//				where = curSize
//			} else {
//				left := (*slice)[0:i]
//				left = append(left, value)
//				left = append(left, curValue)
//				*slice = left
//				where = curSize - 1
//				//fmt.Printf("D: prepended %02x before %02x at %d\n",
//				//	value, curValue, where)
//			}
//		}
//	}
//	newSize := uint(len(*slice))
//	c.Assert(newSize, Equals, curSize+1)
//	//fmt.Printf("  inserted 0x%02x at %d/%d\n", value, where, newSize)
//	// fmt.Printf("%s\n", s.dumpSlice(slice))
//	return
//} // GEEP
//
//// Verify that the insert algorithm works.  We insert a key, requiring
//// that it be inserted in a point which leaves the table inserted.  We
//// confirm that the position computed by the bit-counting algorithm
//// return the same index.
//
//// XXX OBSOLETE XXX
//
//func (s *XLSuite) TestInsertAlgo(c *C) {
//
//	var (
//		bitmap, flag, hc, idx, mask uint32
//		depth, pos, where             uint
//	)
//	rng := xr.MakeSimpleRNG()
//	perm := rng.Perm(32) // a random permutation of [0..32)
//	var slice []byte
//
//	for i := byte(0); i < 32; i++ {
//		hc = uint32(perm[i])
//		// insert the value into the hash slice in such a way as
//		// to maintain order
//		idx = (hc >> depth) & 0x1f
//		c.Assert(idx, Equals, hc) // hc is restricted to that range
//		where = s.insertHash(c, &slice, byte(idx))
//		flag = 1 << (idx + 1)
//		mask = flag - 1
//		pos = BitCount32(bitmap & mask)
//		occupied := uint32(1 << idx)
//		bitmap |= uint32(occupied)
//
//		//fmt.Printf("%02d: hc %02x, idx %02x, mask 0x%08x, bitmap 0x%08x, pos %02d where %02d\n\n",
//		//	i, hc, idx, mask, bitmap, pos, where)
//		c.Assert(uint(pos), Equals, where)
//	}
//} // GEEPGEEP

func (s *XLSuite) TestTable32Ctor(c *C) {
	rng := xr.MakeSimpleRNG()
	depth := uint(rng.Intn(7)) // max of 6, because 6*5 = 30 < 32
	t32, err := NewTable32(depth)
	c.Assert(err, IsNil)
	c.Assert(t32, NotNil)
	c.Assert(t32.GetDepth(), Equals, depth)

	c.Assert(t32.slots, IsNil)
}

func (s *XLSuite) TestDepthZeroInserts(c *C) {

	var (
		bitmap, flag, idx, mask uint32
		pos                     uint
	)
	rng := xr.MakeSimpleRNG()
	perm := rng.Perm(32) // a random permutation of [0..32)
	depth := uint(0)     // XXX COULD VARY DEPTH

	t32, err := NewTable32(depth)
	c.Assert(err, IsNil)
	c.Assert(t32, NotNil)
	c.Assert(t32.GetDepth(), Equals, depth)

	for i := uint(0); i < 32; i++ {
		ndx := byte(perm[i])
		key := make([]byte, 32)
		key[0] = ndx // all the rest are zeroes
		key32, err := NewBytes32Key(key)
		c.Assert(err, IsNil)
		c.Assert(key32, NotNil)
		hc, err := key32.Hashcode32()
		c.Assert(err, IsNil)
		c.Assert(hc, Equals, uint32(ndx))

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
		idx = (hc >> depth) & 0x1f
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
}
