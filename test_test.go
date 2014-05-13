package hamt_go

// hamt_go/teset_test.go

import (
	"fmt"
	xr "github.com/jddixon/rnglib_go"
	. "gopkg.in/check.v1"
)

// Generate keyCount raw keys (byte slices with random content) of length
// keyLen; the keys are guaranteed to have random ndx values (hashcode & mask)

func (s *XLSuite) uniqueKeyMaker(c *C, rng *xr.PRNG, w, keyCount, keyLen uint) (
	rawKeys [][]byte, bKeys []*BytesKey, hashcodes []uint64, values []interface{}) {

	maxCount := uint(1 << w)
	if keyCount > maxCount {
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

			hc = bKey.Hashcode()
			ndx := hc & mask
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

// build 2^w keys, each having a unique value in the first w bits
func (s *XLSuite) makePermutedKeys(rng *xr.PRNG, w uint) (
	fields []int, // FOR DEBUGGING ONLY
	keys [][]byte) {

	fieldCount := (1 << w) - 1    // we don't want the zero value
	fields = rng.Perm(fieldCount) // so 2^w distinct values
	for i := 0; i < len(fields); i++ {
		fields[i] += 1
	}
	keyLen := uint((int(w)*fieldCount + 7) / 8) // in bytes, rounded up
	keyCount := uint(fieldCount)
	keys = make([][]byte, keyCount)

	for i := uint(0); i < keyCount; i++ {
		key := make([]byte, keyLen) // all zeroes
		if i != uint(0) {
			copy(key, keys[i-1])
		}
		// OR the field into the appropriate byte(s) of the key
		bitOffset := w * i
		whichByte := bitOffset / uint(8)
		whichBit := bitOffset % uint(8)

		// lower half of the field
		key[whichByte] |= byte(fields[i] << whichBit)
		if whichBit+w >= 8 {
			key[whichByte+1] |= byte(fields[i] >> (8 - whichBit))
		}
		keys[i] = key
	}

	return
}
