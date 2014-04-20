package hamt_go

// hamt_go/hamt64_test.go

import (
	"bytes"
	"fmt"
	xr "github.com/jddixon/xlattice_go/rnglib"
	. "launchpad.net/gocheck"
	// "sync/atomic"
	//"unsafe"
)

var _ = fmt.Print

func (s *XLSuite) xxxTestHAMT64Ctor(c *C) {
	h64 := NewHAMT64()
	c.Assert(h64, NotNil)
}

// XXX Initially just a copy of the function in table_test.go.

func (s *XLSuite) xxxTestDepthZeroHAMT64(c *C) {

	rng := xr.MakeSimpleRNG()
	perm := rng.Perm(64) // a random permutation of [0..64)

	h64 := NewHAMT64()
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

		// verify the key is not present
		_, err = h64.Find(key64)
		c.Assert(err, Equals, NotFound)

		// insert the key and value
		err = h64.Insert(key64, &key)
		c.Assert(err, IsNil)

		// verify that the key is now present
		v, err := h64.Find(key64)
		c.Assert(err, IsNil)
		vBytes := v.(*[]byte)
		c.Assert(bytes.Equal(*vBytes, key), Equals, true)
	}
	// remove each key, then verify that it is in fact gone =========
	for i := uint(0); i < 64; i++ {
		key := keys[i]

		// verify it is present
		//fmt.Printf("%d VERIFYING PRESENT BEFORE DELETE: idx %02x\n", i, idx)
		key64, err := NewBytes64Key(key)
		c.Assert(err, IsNil)
		c.Assert(key64, NotNil)
		v, err := h64.Find(key64)
		c.Assert(err, IsNil)
		c.Assert(v, NotNil)
		vAsKey := v.(*[]byte)
		c.Assert(bytes.Equal(*vAsKey, key), Equals, true)

		// delete it
		err = h64.Delete(key64)
		c.Assert(err, IsNil)
		v, err = h64.Find(key64)
		c.Assert(err, Equals, NotFound)
		c.Assert(v, IsNil)
	}
}
func (s *XLSuite) xxxTestHAMT64InsertsOfRandomishValues(c *C) {

	const KEY_COUNT = 1024
	var err error

	rng := xr.MakeSimpleRNG()
	h64 := NewHAMT64()
	c.Assert(h64, NotNil)

	keys := make([][]byte, KEY_COUNT)
	key64s := make([]*Bytes64Key, KEY_COUNT)
	hashcodes := make([]uint64, KEY_COUNT)
	values := make([]interface{}, KEY_COUNT)
	hcMap := make(map[uint64]bool)

	// Build KEY_COUNT keys of length 64, using the hashcode to
	// guarantee uniqueness.
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
	// Insert the KEY_COUNT entries
	for i := uint(0); i < KEY_COUNT; i++ {
		// expect that no entry with this key can be found
		key64 := key64s[i]
		_, err = h64.Find(key64)
		c.Assert(err, Equals, NotFound)

		err = h64.Insert(key64, values[i])
		c.Assert(err, IsNil)

		// confirm that the new entry is now present
		_, err = h64.Find(key64)
		c.Assert(err, IsNil)
	}
	// Delete the KEY_COUNT entries
	for i := uint(0); i < KEY_COUNT; i++ {
		key64 := key64s[i]
		// confirm again that the entry is present
		_, err = h64.Find(key64)
		c.Assert(err, IsNil)

		// delete the entry
		err = h64.Delete(key64)
		c.Assert(err, IsNil)

		// confirm that it is gone
		_, err = h64.Find(key64)
		c.Assert(err, Equals, NotFound)
	}
}
