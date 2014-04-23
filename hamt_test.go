package hamt_go

// hamt_go/hamt_test.go

import (
	"bytes"
	"fmt"
	xr "github.com/jddixon/xlattice_go/rnglib"
	. "launchpad.net/gocheck"
	// "sync/atomic"
	//"unsafe"
)

var _ = fmt.Print

// ==================================================================

func (s *XLSuite) TestHAMTCtor(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("TEST_HAMT_CTOR")
	}
	t := uint(0)
	s.doTestHAMTCtor(c, uint(4), t)
	s.doTestHAMTCtor(c, uint(5), t)
	s.doTestHAMTCtor(c, uint(6), t)
}
func (s *XLSuite) doTestHAMTCtor(c *C, w, t uint) {
	h := NewHAMT(w, t)
	c.Assert(h, NotNil)
}

// ==================================================================

func (s *XLSuite) TestDepthZeroHAMT(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("TEST_DEPTH_ZERO_HAMT")
	}
	w := uint(5)
	s.doTestDepthZeroHAMT(c, w, uint(4))
	s.doTestDepthZeroHAMT(c, w, uint(5))
	s.doTestDepthZeroHAMT(c, w, uint(6))
}

func (s *XLSuite) doTestDepthZeroHAMT(c *C, w, t uint) {
	rng := xr.MakeSimpleRNG()

	KEY_LEN := uint(16)
	KEY_COUNT := powerOfTwo(t) // fill all slots

	h := NewHAMT(w, t)
	// DEBUG
	//flag := uint64(1 << t)
	//mask := flag - 1
	// END

	rawKeys, bKeys, hashcodes, values := s.uniqueKeyMaker(
		c, rng, t, KEY_COUNT, KEY_LEN)

	_ = hashcodes // XXX DEBUG

	for i := uint(0); i < KEY_COUNT; i++ {
		key := rawKeys[i]
		bKey := bKeys[i]

		// verify the key is not present
		_, err := h.Find(bKey)
		c.Assert(err, Equals, NotFound)

		// insert the key and value
		// DEBUG
		//ndx := hashcodes[i] & mask
		//fmt.Printf("%2d: ndx 0x%03x, key %s\n", i, ndx, dumpByteSlice(key))
		// END
		err = h.Insert(bKey, values[i])
		c.Assert(err, IsNil)

		// verify that the key is now present
		v, err := h.Find(bKey)
		c.Assert(err, IsNil)
		vBytes := v.(*[]byte)
		c.Assert(bytes.Equal(*vBytes, key), Equals, true)
	}
	// remove each key, then verify that it is in fact gone =========
	for i := uint(0); i < KEY_COUNT; i++ {
		key := rawKeys[i]

		// verify it is present
		//fmt.Printf("%d VERIFYING PRESENT BEFORE DELETE: idx %02x\n", i, idx)
		bKey, err := NewBytesKey(key)
		c.Assert(err, IsNil)
		c.Assert(bKey, NotNil)
		v, err := h.Find(bKey)
		c.Assert(err, IsNil)
		c.Assert(v, NotNil)
		vAsKey := v.(*[]byte)
		c.Assert(bytes.Equal(*vAsKey, key), Equals, true)

		// delete it
		err = h.Delete(bKey)
		c.Assert(err, IsNil)
		v, err = h.Find(bKey)
		c.Assert(err, Equals, NotFound)
		c.Assert(v, IsNil)
	}
} // GEEP

func (s *XLSuite) TestHAMTInsertsOfRandomishValues(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("TEST_HAMT_INSERT_OF_RANDOMISH_VALUES")
	}
	w := uint(5)
	s.doTestHAMTInsertsOfRandomishValues(c, w, uint(4))
	s.doTestHAMTInsertsOfRandomishValues(c, w, uint(5))
	s.doTestHAMTInsertsOfRandomishValues(c, w, uint(6))

}
func (s *XLSuite) doTestHAMTInsertsOfRandomishValues(c *C, w, t uint) {
	KEY_COUNT := uint(32) // 1024)
	KEY_LEN := uint(16)
	var err error

	rng := xr.MakeSimpleRNG()
	h := NewHAMT(w, t)
	c.Assert(h, NotNil)

	// BEGIN KEY_MAKER ==============================================
	keys := make([][]byte, KEY_COUNT)
	bKeys := make([]*BytesKey, KEY_COUNT)
	hashcodes := make([]uint64, KEY_COUNT)
	values := make([]interface{}, KEY_COUNT)
	hcMap := make(map[uint64]bool)

	// Build KEY_COUNT keys of length KEY_LEN, using the hashcode to
	// guarantee uniqueness.
	for i := uint(0); i < KEY_COUNT; i++ {
		var hc uint64
		key := make([]byte, KEY_LEN)
		for {
			rng.NextBytes(key) // fill with quasi-random values
			keys[i] = key

			bKey, err := NewBytesKey(key)
			c.Assert(err, IsNil)
			c.Assert(bKey, NotNil)
			bKeys[i] = bKey

			hc, err = bKey.Hashcode()
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
	// END KEY_MAKER ================================================

	// Insert the KEY_COUNT entries
	for i := uint(0); i < KEY_COUNT; i++ {
		// expect that no entry with this key can be found
		bKey := bKeys[i]
		_, err = h.Find(bKey)
		c.Assert(err, Equals, NotFound)
		//// DEBUG
		//if i == 0 {
		//	fmt.Printf("after insertion key 0: %s\n", dumpByteSlice(bKey.Slice))
		//}
		//// END

		err = h.Insert(bKey, values[i])
		c.Assert(err, IsNil)

		// confirm that the new entry is now present
		_, err = h.Find(bKey)
		c.Assert(err, IsNil)
		// DEBUG
		//if i == 0 {
		//	fmt.Printf("after insertion key 0: %s\n", dumpByteSlice(bKey.Slice))
		//}
		// END
	}
	// SIMPLE SCAN
	for i := uint(0); i < KEY_COUNT; i++ {
		// expect that entry with this key can be found
		bKey := bKeys[i]
		_, err = h.Find(bKey)
		// DEBUG
		if err != nil {
			fmt.Printf("simple scan, slot %4d: ERROR: %s\n", i, err.Error())
		}
		// END
	}
	// END SIMPLE

	// Delete the KEY_COUNT entries
	for i := uint(0); i < KEY_COUNT; i++ {
		bKey := bKeys[i]
		// DEBUG
		//if i == 0 {
		//	fmt.Printf("before deletion key 0: %s\n", dumpByteSlice(bKey.Slice))
		//}
		// fmt.Printf("verifying that key %d is still present\n", i)
		// END
		// confirm again that the entry is present
		_, err = h.Find(bKey)
		c.Assert(err, IsNil) // FAILS XXX

		// delete the entry
		err = h.Delete(bKey)
		c.Assert(err, IsNil)

		// confirm that it is gone
		_, err = h.Find(bKey)
		c.Assert(err, Equals, NotFound)
	}
}
