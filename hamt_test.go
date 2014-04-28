package hamt_go

// hamt_go/hamt_test.go

import (
	"bytes"
	"fmt"
	xr "github.com/jddixon/xlattice_go/rnglib"
	. "gopkg.in/check.v1"
	//. "launchpad.net/gocheck"
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

// 'aaa' in effect comments out the test
func (s *XLSuite) aaaTestDepthZeroHAMT(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("TEST_DEPTH_ZERO_HAMT")
	}
	w := uint(5)
	s.doTestDepthZeroHAMT(c, w, uint(4))
	s.doTestDepthZeroHAMT(c, w, uint(5))
	s.doTestDepthZeroHAMT(c, w, uint(6))
	s.doTestDepthZeroHAMT(c, w, uint(7))
	s.doTestDepthZeroHAMT(c, w, uint(8))
}

func (s *XLSuite) doTestDepthZeroHAMT(c *C, w, t uint) {
	rng := xr.MakeSimpleRNG()

	KEY_LEN := uint(16)
	KEY_COUNT := uint(1 << t) // fill all slots

	h := NewHAMT(w, t)
	rawKeys, bKeys, hashcodes, values := s.uniqueKeyMaker(
		c, rng, t, KEY_COUNT, KEY_LEN)

	_ = hashcodes // XXX DEBUG

	c.Assert(h.GetLeafCount(), Equals, uint(0))

	for i := uint(0); i < KEY_COUNT; i++ {
		key := rawKeys[i]
		bKey := bKeys[i]

		// verify the key is not present
		_, err := h.Find(bKey)
		c.Assert(err, Equals, NotFound)

		// insert the key and value
		err = h.Insert(bKey, values[i])
		c.Assert(err, IsNil)
		c.Assert(h.GetLeafCount(), Equals, i+1)

		// verify that the key is now present
		v, err := h.Find(bKey)
		c.Assert(err, IsNil)
		vBytes := v.(*[]byte)
		c.Assert(bytes.Equal(*vBytes, key), Equals, true)

		// replace the value associated with this key
		newValue := rng.Int63() // a random 64-bit value
		err = h.Insert(bKey, &newValue)
		c.Assert(err, IsNil)

		// verify that Find returns the new value
		ret, err := h.Find(bKey)
		c.Assert(err, IsNil)
		returned := *(ret.(*int64))
		c.Assert(returned, Equals, newValue)

		// put the old value back
		err = h.Insert(bKey, values[i])
		c.Assert(err, IsNil)

		// verify that Find now returns the old value
		v, err = h.Find(bKey)
		c.Assert(err, IsNil)
		vBytes = v.(*[]byte)
		c.Assert(bytes.Equal(*vBytes, key), Equals, true)

	}

	c.Assert(h.GetLeafCount(), Equals, KEY_COUNT)

	// remove each key, then verify that it is in fact gone =========
	for i := uint(0); i < KEY_COUNT; i++ {
		key := rawKeys[i]

		// verify it is present
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
		c.Assert(h.GetLeafCount(), Equals, KEY_COUNT-(i+1))
		v, err = h.Find(bKey)
		c.Assert(err, Equals, NotFound)
		c.Assert(v, IsNil)
	}
}

func (s *XLSuite) TestHAMTInsertsOfRandomishValues(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("TEST_HAMT_INSERT_OF_RANDOMISH_VALUES")
	}
	w := uint(5)
	s.doTestHAMTInsertsOfRandomishValues(c, w, uint(4))
	s.doTestHAMTInsertsOfRandomishValues(c, w, uint(5))
	s.doTestHAMTInsertsOfRandomishValues(c, w, uint(6))
	s.doTestHAMTInsertsOfRandomishValues(c, w, uint(7))
	s.doTestHAMTInsertsOfRandomishValues(c, w, uint(8))

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
	c.Assert(h.GetLeafCount(), Equals, uint(0))
	for i := uint(0); i < KEY_COUNT; i++ {
		// expect that no entry with this key can be found
		bKey := bKeys[i]
		_, err = h.Find(bKey)
		c.Assert(err, Equals, NotFound)

		err = h.Insert(bKey, values[i])
		c.Assert(err, IsNil)
		c.Assert(h.GetLeafCount(), Equals, i+1)

		key := keys[i]
		// verify that the key is now present
		v, err := h.Find(bKey)
		c.Assert(err, IsNil)
		vBytes := v.(*[]byte)
		c.Assert(bytes.Equal(*vBytes, key), Equals, true)

		// replace the value associated with this key
		newValue := rng.Int63() // a random 64-bit value
		err = h.Insert(bKey, &newValue)
		c.Assert(err, IsNil)

		// verify that Find returns the new value
		ret, err := h.Find(bKey)
		c.Assert(err, IsNil)
		returned := *(ret.(*int64))
		c.Assert(returned, Equals, newValue)

		// put the old value back
		err = h.Insert(bKey, values[i])
		c.Assert(err, IsNil)

		// verify that Find now returns the old value
		v, err = h.Find(bKey)
		c.Assert(err, IsNil)
		vBytes = v.(*[]byte)
		c.Assert(bytes.Equal(*vBytes, key), Equals, true)
	}
	// SIMPLE SCAN
	for i := uint(0); i < KEY_COUNT; i++ {
		// expect that entry with this key can be found
		bKey := bKeys[i]
		_, err = h.Find(bKey)
	}
	// END SIMPLE

	// Delete the KEY_COUNT entries
	c.Assert(h.GetLeafCount(), Equals, KEY_COUNT)
	for i := uint(0); i < KEY_COUNT; i++ {
		bKey := bKeys[i]
		// confirm again that the entry is present
		_, err = h.Find(bKey)
		c.Assert(err, IsNil)

		// delete the entry
		err = h.Delete(bKey)
		c.Assert(err, IsNil)
		c.Assert(h.GetLeafCount(), Equals, KEY_COUNT-(i+1))

		// confirm that it is gone
		_, err = h.Find(bKey)
		c.Assert(err, Equals, NotFound)
	}
}

// Insert a series of entries, each of which should replace a leaf with
// a table.

func (s *XLSuite) TestHamtEntrySplittingInserts(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("TEST_HAMT_ENTRY_SPLITTING_INSERTS")
	}
	rng := xr.MakeSimpleRNG()

	s.doTestHamtEntrySplittingInserts(c, rng, 5, 5)
}

func (s *XLSuite) doTestHamtEntrySplittingInserts(c *C, rng *xr.PRNG,
	t, w uint) {

	var err error

	h := NewHAMT(w, t)
	c.Assert(h, NotNil)
	c.Assert(h.GetW(), Equals, w)
	c.Assert(h.GetT(), Equals, t)

	_, rawKeys := s.makePermutedKeys(rng, w) // XXX fields ignored
	KEY_COUNT := 64 / w                      // some keys may be ignored

	bKeys := make([]*BytesKey, KEY_COUNT)
	hashcodes := make([]uint64, KEY_COUNT)
	values := make([]interface{}, KEY_COUNT)

	for i := uint(0); i < KEY_COUNT; i++ {
		key := rawKeys[i]

		bKey, err := NewBytesKey(key)
		c.Assert(err, IsNil)
		c.Assert(bKey, NotNil)
		bKeys[i] = bKey

		values[i] = &key

		hc, err := bKey.Hashcode()
		c.Assert(err, IsNil)
		hashcodes[i] = hc

	}
	tCount := h.GetTableCount()
	c.Assert(tCount, Equals, uint(1))
	c.Assert(h.GetLeafCount(), Equals, uint(0))

	// insertion loop -----------------------------------------------
	for i := uint(0); i < KEY_COUNT; i++ {
		// expect that no entry with this key can be found
		bKey := bKeys[i]
		_, err = h.Find(bKey)
		c.Assert(err, Equals, NotFound)

		// insert the entry -------------------------------
		err = h.Insert(bKey, values[i])
		c.Assert(err, IsNil)
		c.Assert(h.GetLeafCount(), Equals, i+1)

		// confirm the entry is present -------------------
		_, err = h.Find(bKey)
		c.Assert(err, IsNil)

		// all insertions except the first split the entry
		if i != 0 {
			c.Assert(h.GetTableCount(), Equals, tCount+1)
			tCount++
		}
	}
	// deletion loop ------------------------------------------------
	c.Assert(h.GetLeafCount(), Equals, KEY_COUNT)
	for i := uint(0); i < KEY_COUNT; i++ {
		bKey := bKeys[i]

		// confirm again that the entry is present --------
		_, err = h.Find(bKey)
		c.Assert(err, IsNil)

		// delete the entry -------------------------------
		err = h.Delete(bKey)
		c.Assert(err, IsNil)
		c.Assert(h.GetLeafCount(), Equals, KEY_COUNT-(i+1))

		// confirm that it is gone ------------------------
		_, err = h.Find(bKey)
		c.Assert(err, Equals, NotFound)
	}
}
