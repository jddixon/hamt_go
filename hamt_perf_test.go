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

func (s *XLSuite) BenchmarkHAMT_4(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("\nBenchmarkHAMT_4")
	}
	s.doBenchmark(c, uint(4), 0)
}
func (s *XLSuite) BenchmarkHAMT_5(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("\nBenchmarkHAMT_5")
	}
	s.doBenchmark(c, uint(5), 0)
}
func (s *XLSuite) BenchmarkHAMT_6(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("\nBenchmarkHAMT_6")
	}
	s.doBenchmark(c, uint(6), 0)
}
func (s *XLSuite) BenchmarkHAMT_7(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("\nBenchmarkHAMT_7")
	}
	s.doBenchmark(c, uint(7), 0)
}
func (s *XLSuite) BenchmarkHAMT_8(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("\nBenchmarkHAMT_8")
	}
	s.doBenchmark(c, uint(8), 0)
}
func (s *XLSuite) doBenchmark(c *C, w, t uint) {
	// build an array of N random-ish K-byte rawKeys
	K := 16
	N := c.N
	t0 := time.Now()
	rawKeys, bKeys := makeSomeMoreKeys(N, K)
	t1 := time.Now()
	deltaT := t1.Sub(t0)
	fmt.Printf("setup time for %d %d-byte rawKeys: %v\n", N, K, deltaT)

	m := NewHAMT(w, t)

	c.ResetTimer()
	c.StartTimer()
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
	// rough measure of resource consumption
	tableCount := m.root.GetTableCount()
	slotCount := m.root.maxSlots * tableCount
	byteCount := slotCount * 8
	megabytes := float32(byteCount) / (1024 * 1024)
	fmt.Printf("%6d tables, %7d slots, %3.1f megabytes\n",
		tableCount, slotCount, megabytes)
}
