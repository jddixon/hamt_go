package main

// hamt_go/cmd/concurProfileHAMT.go

/////////////////////////////////////////////////////////////////////
// Run this with something similar to
//   go build
//   ./concurProfileHAMT
/////////////////////////////////////////////////////////////////////

import (
	"errors"
	"flag"
	"fmt"
	gh "github.com/jddixon/hamt_go"
	xr "github.com/jddixon/rnglib_go"
	"os"
	"time"
)

var _ = errors.New

func Usage() {
	fmt.Printf("Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Printf("where the options are:\n")
	flag.PrintDefaults()
}

const ()

var (
	justShow      = flag.Bool("j", false, "display option settings and exit")
	showTimestamp = flag.Bool("t", false, "output UTC timestamp")
	showVersion   = flag.Bool("V", false, "output package version info")
	testing       = flag.Bool("T", false, "test run")
	verbose       = flag.Bool("v", false, "be talkative")
)

// -- utilities -----------------------------------------------------

// Create N random-ish K-byte values.  These are to be used as BytesKeys,
// so the first 64 bits must represent a unique value.

func makeSomeUniqueKeys(N, K uint) (rawKeys [][]byte, bKeys []*gh.BytesKey) {

	rng := xr.MakeSimpleRNG()
	rawKeys = make([][]byte, N)
	bKeys = make([]*gh.BytesKey, N)
	keyMap := make(map[uint64]bool)

	for i := uint(0); i < N; i++ {
		var bKey *gh.BytesKey
		key := make([]byte, K)
		for {
			rng.NextBytes(key)
			bKey, _ = gh.NewBytesKey(key)
			hc := bKey.Hashcode()
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

// Insert N items into the HAMT, then find each J times.  This is
// a very crude test!
//
// (a) inserts should be done with a write lock (doesn't matter now
//     because insertions are single-threaded
// (b) finds should be done with a read lock
// (c) inserts and finds should be overlapped
//
func doBenchmark(w, t uint, J, N uint) (deltaT time.Duration) {
	// build an array of N random-ish K-byte rawKeys
	K := uint(16)
	rawKeys, bKeys := makeSomeUniqueKeys(N, K)

	// set up HAMT, ignoring any errors
	m, _ := gh.NewHAMT(w, t)
	done := make([]chan bool, J)
	for i := uint(0); i < J; i++ {
		done[i] = make(chan bool)
	}

	t0 := time.Now()
	for i := uint(0); i < N; i++ {
		_ = m.Insert(bKeys[i], &rawKeys[i])
	}

	// Verify several times that the rawKeys are present in the map.
	for j := uint(0); j < J; j++ {
		go func(j uint) {
			for i := uint(0); i < N; i++ {
				value, err := m.Find(bKeys[i])
				// DEBUG
				if err != nil {
					fmt.Printf("error finding key %d\n", i, err.Error())
				}
				if value == nil {
					fmt.Printf("cannot find key %d\n", i)
				}
				// END
				//val := value.(*[]byte)	// NOT USED
				_ = value
			}
			done[j] <- true
		}(j)
	}
	for j := uint(0); j < J; j++ {
		<-done[j]
	}
	t1 := time.Now()
	deltaT = t1.Sub(t0)
	return
}

// MAIN /////////////////////////////////////////////////////////////
func main() {
	var (
		err error
	)

	flag.Usage = Usage
	flag.Parse()

	// FIXUPS ///////////////////////////////////////////////////////
	if !*justShow {
		// XXX STUB
	}
	if *testing {
	}
	// SANITY CHECKS ////////////////////////////////////////////////
	if err == nil {
	}
	// DISPLAY OPTIONS //////////////////////////////////////////////
	if err == nil && *verbose || *justShow {
		fmt.Printf("justShow    	= %v\n", *justShow)
		fmt.Printf("showTimestamp   = %v\n", *showTimestamp)
		fmt.Printf("showVersion 	= %v\n", *showVersion)
		fmt.Printf("testing     	= %v\n", *testing)
		fmt.Printf("verbose     	= %v\n", *verbose)
	}
	// DO IT ////////////////////////////////////////////////////////
	if err == nil && !*justShow {
		w := uint(6)
		n := uint(19)
		t := n - 2
		j := uint(8) // degree of concurrency
		N := uint(1 << n)
		if err == nil {
			deltaT := doBenchmark(w, t, j, N).Nanoseconds() // int64
			opCount := int64((1 + j) * N)
			fmt.Printf("run time for %d rawKeys: %5.2f sec; ",
				N, float64(deltaT)/1000000000.0)
			fmt.Printf("%d finds/insert; ", j)
			fmt.Printf("%d ns/op\n", deltaT/opCount)
		}
	}

	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(-1)
	}

}
