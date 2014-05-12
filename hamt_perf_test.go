package hamt_go

// hamt_go/hamt_perf_test.go

/////////////////////////////////////////////////////////////////////
// THIS NEEDS TO BE RUN WITH
//   go test -gocheck.b
/////////////////////////////////////////////////////////////////////

import (
	"bytes"
	"fmt"
	xr "github.com/jddixon/rnglib_go"
	. "gopkg.in/check.v1"
	"time"
)

var _ = fmt.Print

// -- utilities -----------------------------------------------------

// Create N random-ish K-byte values.  These are to be used as BytesKeys,
// so the first 64 bits must represent a unique value.

func makeSomeUniqueKeys(N, K int) (rawKeys [][]byte, bKeys []*BytesKey) {

	rng := xr.MakeSimpleRNG()
	rawKeys = make([][]byte, N)
	bKeys = make([]*BytesKey, N)
	keyMap := make(map[uint64]bool)

	for i := 0; i < N; i++ {
		var bKey *BytesKey
		key := make([]byte, K)
		for {
			rng.NextBytes(key)
			bKey, _ = NewBytesKey(key)
			hc, _ := bKey.Hashcode()
			_, ok := keyMap[hc]
			if !ok { // value is not in the map
				keyMap[hc] = true
				break
			}
		}
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

// 2^6 is 64, so if we are using a 64-bit integer as a bit map into
// the slots, this is as far as we can go.
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
	rawKeys, bKeys := makeSomeUniqueKeys(N, K)
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
			fmt.Printf("error finding key %d\n", i, err.Error())
		}
		if value == nil {
			fmt.Printf("cannot find key %d\n", i)
		}
		// END
		c.Assert(err, IsNil)
		c.Assert(value, NotNil)
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
