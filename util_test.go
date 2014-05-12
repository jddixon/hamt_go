package hamt_go

// hamt_go/util_test.go

import (
	"code.google.com/p/intmath/intgr"
	"fmt"
	xr "github.com/jddixon/rnglib_go"
	. "gopkg.in/check.v1"
)

var _ = fmt.Print

func slowBitCount32(n uint32) uint {
	var count uint32
	for i := 0; i < 32; i++ {
		count += n & 1
		n >>= 1
	}
	return uint(count)
}

func slowBitCount64(n uint64) uint {
	var count uint64
	for i := 0; i < 64; i++ {
		count += n & 1
		n >>= 1
	}
	return uint(count)
}

func (s *XLSuite) TestSWAR32(c *C) {

	rng := xr.MakeSimpleRNG()
	for i := 0; i < 8; i++ {
		n := uint32(rng.Int63())
		slowCount := slowBitCount32(n)
		intgrCount := uint(intgr.BitCount(int(n)))
		swar32Count := BitCount32(n)
		c.Assert(swar32Count, Equals, intgrCount)
		c.Assert(swar32Count, Equals, slowCount)
	}
}

func (s *XLSuite) TestSWAR64(c *C) {

	rng := xr.MakeSimpleRNG()
	for i := 0; i < 8; i++ {
		n1 := uint64(rng.Int63())
		n2 := uint64(rng.Int63())
		n := (n1 << 32) ^ n2 // we want a full 64 random bits
		slowCount := slowBitCount64(n)
		intgrCount := uint(intgr.BitCount(int(n)))
		swar64Count := BitCount64(n)
		c.Assert(swar64Count, Equals, intgrCount)
		c.Assert(swar64Count, Equals, slowCount)
	}
}

func (s *XLSuite) TestMakingPermutedKeys(c *C) {

	rng := xr.MakeSimpleRNG()
	var w uint
	for w = uint(4); w < uint(7); w++ {
		fields, keys := s.makePermutedKeys(rng, w)
		flag := uint64(1 << w)
		mask := flag - 1
		maxDepth := 64 / w // rounding down
		fieldCount := uint(len(fields))
		// we are relying on Hashcode(), which has only 64 bits
		if maxDepth > fieldCount {
			maxDepth = fieldCount
		}
		for i := uint(0); i < maxDepth; i++ {
			bKey, err := NewBytesKey(keys[i])
			c.Assert(err, IsNil)
			hc, err := bKey.Hashcode()
			c.Assert(err, IsNil)
			for j := uint(0); j <= i; j++ {
				ndx := hc & mask
				if uint(ndx) != uint(fields[j]) {
					fmt.Printf(
						"GLITCH: w %d keyset %d field[%2d] %02x ndx %02x\n",
						w, i, j, fields[j], ndx)
				}
				c.Assert(uint(ndx), Equals, uint(fields[j]))
				hc >>= w
			}
		}
	}
}
