package hamt_go

import (
	// . "launchpad.net/gocheck"
	. "gopkg.in/check.v1"
	"testing"
)

// IF USING gocheck, need a file like this in each package=directory.

func Test(t *testing.T) { TestingT(t) }

type XLSuite struct{}

var _ = Suite(&XLSuite{})

const VERBOSITY = 1
