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

func (s *XLSuite) TestHAMTCtor(c *C) {
	h32 := NewHAMT()
	c.Assert(h32, NotNil)
}

// XXX Initially just a copy of the function in table_test.go.

func (s *XLSuite) TestDepthZeroHAMT(c *C) {

	rng := xr.MakeSimpleRNG()
	perm := rng.Perm(32) // a random permutation of [0..32)

	h32 := NewHAMT()
	keys := make([][]byte, 32)
	indices := make([]byte, 32)

	for i := uint(0); i < 32; i++ {
		ndx := byte(perm[i])
		indices[i] = ndx

		key := make([]byte, 32)
		key[0] = ndx // all the rest are zeroes
		keys[i] = key

		key64, err := NewBytesKey(key)
		c.Assert(err, IsNil)
		c.Assert(key64, NotNil)

		// verify the key is not present
		_, err = h32.Find(key64)
		c.Assert(err, Equals, NotFound)

		// insert the key and value
		err = h32.Insert(key64, &key)
		c.Assert(err, IsNil)

		// verify that the key is now present
		v, err := h32.Find(key64)
		c.Assert(err, IsNil)
		vBytes := v.(*[]byte)
		c.Assert(bytes.Equal(*vBytes, key), Equals, true)
	}
	// remove each key, then verify that it is in fact gone =========
	for i := uint(0); i < 32; i++ {
		key := keys[i]

		// verify it is present
		//fmt.Printf("%d VERIFYING PRESENT BEFORE DELETE: idx %02x\n", i, idx)
		key64, err := NewBytesKey(key)
		c.Assert(err, IsNil)
		c.Assert(key64, NotNil)
		v, err := h32.Find(key64)
		c.Assert(err, IsNil)
		c.Assert(v, NotNil)
		vAsKey := v.(*[]byte)
		c.Assert(bytes.Equal(*vAsKey, key), Equals, true)

		// delete it
		err = h32.Delete(key64)
		c.Assert(err, IsNil)
		v, err = h32.Find(key64)
		c.Assert(err, Equals, NotFound)
		c.Assert(v, IsNil)
	}
}
func (s *XLSuite) TestHAMTInsertsOfRandomishValues(c *C) {

	const KEY_COUNT = 1024
	var err error

	rng := xr.MakeSimpleRNG()
	h32 := NewHAMT()
	c.Assert(h32, NotNil)

	keys := make([][]byte, KEY_COUNT)
	key64s := make([]*BytesKey, KEY_COUNT)
	hashcodes := make([]uint64, KEY_COUNT)
	values := make([]interface{}, KEY_COUNT)
	hcMap := make(map[uint64]bool)

	// Build KEY_COUNT keys of length 32, using the hashcode to
	// guarantee uniqueness.
	for i := uint(0); i < KEY_COUNT; i++ {
		var hc uint64
		key := make([]byte, 32)
		for {
			rng.NextBytes(key) // fill with quasi-random values
			keys[i] = key

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
	// Insert the KEY_COUNT entries
	for i := uint(0); i < KEY_COUNT; i++ {
		// expect that no entry with this key can be found
		key64 := key64s[i]
		_, err = h32.Find(key64)
		c.Assert(err, Equals, NotFound)

		err = h32.Insert(key64, values[i])
		c.Assert(err, IsNil)

		// confirm that the new entry is now present
		_, err = h32.Find(key64)
		c.Assert(err, IsNil)
	}
	// Delete the KEY_COUNT entries
	for i := uint(0); i < KEY_COUNT; i++ {
		key64 := key64s[i]
		// confirm again that the entry is present
		_, err = h32.Find(key64)
		c.Assert(err, IsNil)

		// delete the entry
		err = h32.Delete(key64)
		c.Assert(err, IsNil)

		// confirm that it is gone
		_, err = h32.Find(key64)
		c.Assert(err, Equals, NotFound)
	}
}
