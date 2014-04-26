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
	//. "launchpad.net/gocheck"
	. "gopkg.in/check.v1"
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

func (s *XLSuite) BenchmarkHAMT_3(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("\nBenchmarkHAMT_3")
	}
	s.doBenchmark(c, uint(3), 18)
}
func (s *XLSuite) BenchmarkHAMT_4(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("\nBenchmarkHAMT_4")
	}
	s.doBenchmark(c, uint(4), 18)
}
func (s *XLSuite) BenchmarkHAMT_5(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("\nBenchmarkHAMT_5")
	}
	s.doBenchmark(c, uint(5), 18)
}
func (s *XLSuite) BenchmarkHAMT_6(c *C) {
	if VERBOSITY > 0 {
		fmt.Println("\nBenchmarkHAMT_6")
	}
	s.doBenchmark(c, uint(6), 18)
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
	tableCount := m.GetTableCount()
	slotsPerChild := uint(1 << m.root.w)
	// the slots in the root table plus those in child tables
	slotCount := m.root.slotCount + slotsPerChild*(tableCount-1)
	byteCount := slotCount * 8 // assume 64-bit architecture
	megabytes := float32(byteCount) / (1000 * 1000)
	fmt.Printf("%6d tables, %8d slots, %5.2f megabytes\n",
		tableCount, slotCount, megabytes)

	// SAY AGAIN ???
	fmt.Printf("t %2d, slotsInRoot %7d, w %d, slotsPerChild %2d\n",
		t, m.root.slotCount,
		w, slotsPerChild)
}
