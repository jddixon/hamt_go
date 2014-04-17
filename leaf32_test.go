package hamt_go

// hamt_go/leaf32_test.go

import (
	"fmt"
	xr "github.com/jddixon/xlattice_go/rnglib"
	. "launchpad.net/gocheck"
)

var _ = fmt.Print

func (s *XLSuite) TestLeaf32(c *C) {
	const MIN_KEY_LEN = 8
	rng := xr.MakeSimpleRNG()
	_ = rng

	// a nil argument must cause an error
	p := 0
	gk := make([]byte, MIN_KEY_LEN)
	goodKey, err := NewBytes32Key(gk)
	c.Assert(err, IsNil)

	sk := make([]byte, MIN_KEY_LEN-1)
	_, err = NewBytes32Key(sk)
	// must fail - key is too short
	c.Assert(err, NotNil)

	_, err = NewLeaf32(nil, &p)
	c.Assert(err, NotNil)

	_, err = NewLeaf32(goodKey, nil)
	c.Assert(err, NotNil)

	leaf, err := NewLeaf32(goodKey, &p)
	c.Assert(err, IsNil)
	c.Assert(leaf, NotNil)
	c.Assert(leaf.IsLeaf(), Equals, true)

	// XXX test a Table32, IsLeaf() should return false
	// XXX STUB

}
