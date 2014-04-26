package hamt_go

// hamt_go/leaf_test.go

import (
	"fmt"
	xr "github.com/jddixon/xlattice_go/rnglib"
	. "gopkg.in/check.v1"
	//. "launchpad.net/gocheck"
)

var _ = fmt.Print

func (s *XLSuite) TestLeaf(c *C) {
	const MIN_KEY_LEN = 8
	rng := xr.MakeSimpleRNG()
	_ = rng

	// a nil argument must cause an error
	p := 0
	gk := make([]byte, MIN_KEY_LEN)
	goodKey, err := NewBytesKey(gk)
	c.Assert(err, IsNil)

	sk := make([]byte, MIN_KEY_LEN-1)
	_, err = NewBytesKey(sk)
	// must fail - key is too short
	c.Assert(err, NotNil)

	_, err = NewLeaf(nil, &p)
	c.Assert(err, NotNil)

	_, err = NewLeaf(goodKey, nil)
	c.Assert(err, NotNil)

	leaf, err := NewLeaf(goodKey, &p)
	c.Assert(err, IsNil)
	c.Assert(leaf, NotNil)
	c.Assert(leaf.IsLeaf(), Equals, true)

	// XXX test a Table, IsLeaf() should return false
	// XXX STUB

}
