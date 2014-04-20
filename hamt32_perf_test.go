package hamt_go

// hamt_go/hamt32_perf_test.go

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

// Create N random-ish K-byte values.  It takes about 2 us to create
// a value (21.2 ms for 10K values, 2.008s for 1M values)

func makeSomeKey64s(N, K int) (keys [][]byte, key64s []*Bytes64Key) {

	rng := xr.MakeSimpleRNG()
	keys = make([][]byte, N)
	key64s = make([]*Bytes64Key, N)

	for i := 0; i < N; i++ {
		key := make([]byte, K)
		rng.NextBytes(key)
		key64, _ := NewBytes64Key(key)

		keys[i] = key
		key64s[i] = key64
	}
	return
}

// -- tests proper --------------------------------------------------

func (s *XLSuite) BenchmarkMakeSomeKey64s(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("TEST_MAKE_SOME_KEYS")
	}

	// build an array of N random-ish K-byte keys
	K := 16
	N := c.N
	t0 := time.Now()
	keys, key64s := makeSomeKey64s(N, K)
	t1 := time.Now()
	deltaT := t1.Sub(t0)
	fmt.Printf("setup time for %d %d-byte keys: %v\n", N, K, deltaT)

	// build an IDMap to put them in (ignoring any error)
	m := NewHAMT32()

	c.ResetTimer()
	c.StartTimer()
	// my results: 1603 ns/op for a run of 1 million insertions
	for i := 0; i < c.N; i++ {
		_ = m.Insert(key64s[i], &keys[i])
	}
	c.StopTimer()

	// verify that the keys are present in the map
	for i := 0; i < N; i++ {
		value, err := m.Find(key64s[i])
		// DEBUG
		if err != nil {
			fmt.Printf("cannot find key %d: %s\n", i, err.Error())
		}
		// END
		c.Assert(err, IsNil)
		val := value.(*[]byte)
		c.Assert(bytes.Equal(*val, keys[i]), Equals, true)

	}
	// A BIT OF A HACK
	fmt.Printf("table count = %d\n", m.root.GetTableCount())
}
