package hamt_go

// hamt_go/hamt_perf_test.go

/////////////////////////////////////////////////////////////////////
// THIS NEEDS TO BE RUN WITH
//   go test -gocheck.b
/////////////////////////////////////////////////////////////////////

import (
	"bytes"
	"fmt"
	xr "github.com/jddixon/xlattice_go/rnglib"
	. "launchpad.net/gocheck"
	"time"
)

var _ = fmt.Print

// -- utilities -----------------------------------------------------

// Create N random-ish K-byte values.

func makeSomeMoreKeys(N, K int) (rawKeys [][]byte, bKeys []*BytesKey) {

	rng := xr.MakeSimpleRNG()
	rawKeys = make([][]byte, N)
	bKeys = make([]*BytesKey, N)

	for i := 0; i < N; i++ {
		key := make([]byte, K)
		rng.NextBytes(key)
		bKey, _ := NewBytesKey(key)

		rawKeys[i] = key
		bKeys[i] = bKey
	}
	return
}

// -- tests proper --------------------------------------------------

func (s *XLSuite) BenchmarkMakeSomeMoreKeys(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("TEST_MAKE_SOME_MORE_KEYS")
	}

	// build an array of N random-ish K-byte rawKeys
	K := 16
	N := c.N
	t0 := time.Now()
	rawKeys, bKeys := makeSomeMoreKeys(N, K)
	t1 := time.Now()
	deltaT := t1.Sub(t0)
	fmt.Printf("setup time for %d %d-byte rawKeys: %v\n", N, K, deltaT)

	// build an IDMap to put them in (ignoring any error)
	m := NewHAMT()

	c.ResetTimer()
	c.StartTimer()
	// my results: 1603 ns/op for a run of 1 million insertions
	for i := 0; i < c.N; i++ {
		_ = m.Insert(bKeys[i], &rawKeys[i])
	}
	c.StopTimer()

	// verify that the rawKeys are present in the map
	for i := 0; i < N; i++ {
		value, err := m.Find(bKeys[i])
		// DEBUG
		if err != nil {
			fmt.Printf("cannot find key %d: %s\n", i, err.Error())
		}
		// END
		c.Assert(err, IsNil)
		val := value.(*[]byte)
		c.Assert(bytes.Equal(*val, rawKeys[i]), Equals, true)

	}
	// A BIT OF A HACK
	fmt.Printf("table count = %d\n", m.root.GetTableCount())
}
