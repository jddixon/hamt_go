package hamt_go

// hamt_go/bytesKey_test.go

import (
	// "encoding/binary"
	"fmt"
	xr "github.com/jddixon/xlattice_go/rnglib"
	. "launchpad.net/gocheck"
)

var _ = fmt.Print

func (s *XLSuite) TestBytes32Key(c *C) {
	rng := xr.MakeSimpleRNG()

	for i := uint(0); i < 8; i++ {
		// length shouldn't matter, so long as it's > 8
		length := 8 + rng.Intn(32)
		data := make([]byte, length)
		rng.NextBytes(data)

		var expected uint32
		for j := uint(0); j < 8; j++ {
			// need to convert data[j] before shifting
			expected += uint32(data[j]) << (8 * j)
		}

		key, err := NewBytes32Key(data)
		c.Assert(err, IsNil)
		c.Assert(key, NotNil)
		hc, err := key.Hashcode32()
		c.Assert(err, IsNil)
		c.Assert(hc, Equals, expected)
	}

}

func (s *XLSuite) TestBytes64Key(c *C) {
	rng := xr.MakeSimpleRNG()

	for i := uint(0); i < 16; i++ {
		// length shouldn't matter, so long as it's > 16
		length := 16 + rng.Int63n(64)
		data := make([]byte, length)
		rng.NextBytes(data)

		var expected uint64
		for j := uint(0); j < 16; j++ {
			// need to convert data[j] before shifting
			expected += uint64(data[j]) << (8 * j)
		}

		key, err := NewBytes64Key(data)
		c.Assert(err, IsNil)
		c.Assert(key, NotNil)
		hc, err := key.Hashcode64()
		c.Assert(err, IsNil)
		c.Assert(hc, Equals, expected)
	}

}
