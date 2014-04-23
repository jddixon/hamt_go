package hamt_go

// hamt_go/teset_test.go

import (
	"fmt"
	xr "github.com/jddixon/xlattice_go/rnglib"
	. "launchpad.net/gocheck"
)

// Generate keyCount raw keys (byte slices with random content) of length
// keyLen; the keys are guaranteed to have random ndx values (hashcode & mask)

func (s *XLSuite) uniqueKeyMaker(c *C, rng *xr.PRNG, w, keyCount, keyLen uint) (
	rawKeys [][]byte, bKeys []*BytesKey, hashcodes []uint64, values []interface{}) {

	if keyCount > powerOfTwo(w) {
		msg := fmt.Sprintf(
			"too few bits in %d: cannot guarantee uniqueness of %d keys",
			w, keyCount)
		panic(msg)
	}
	flag := uint64(1 << w)
	mask := flag - 1

	rawKeys = make([][]byte, keyCount)
	bKeys = make([]*BytesKey, keyCount)
	hashcodes = make([]uint64, keyCount)
	values = make([]interface{}, keyCount)

	ndxMap := make(map[uint64]bool)

	// Build keyCount rawKeys of length keyLen, using the masked hashcode to
	// guarantee uniqueness.
	for i := uint(0); i < keyCount; i++ {
		var hc uint64
		key := make([]byte, keyLen)
		for {
			rng.NextBytes(key) // fill with quasi-random values
			rawKeys[i] = key

			bKey, err := NewBytesKey(key)
			c.Assert(err, IsNil)
			c.Assert(bKey, NotNil)
			bKeys[i] = bKey

			hc, err = bKey.Hashcode()
			ndx := hc & mask
			c.Assert(err, IsNil)
			_, ok := ndxMap[ndx]
			if !ok {
				ndxMap[ndx] = true
				break
			}
		}
		values[i] = &key
		hashcodes[i] = hc
	}
	return
}
