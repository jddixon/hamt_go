package main

// hamt_go/cmd/profileHAMT.go

/////////////////////////////////////////////////////////////////////
// Run this with something similar to 
//   time profileHAMT -c cpu.prof -m mem.prof
// Then process results with 
//   go tool pprof profileHAMT cpu.prof mem.prof
/////////////////////////////////////////////////////////////////////

import (
	//"encoding/hex"
	"errors"
	"flag"
	"fmt"
	gh "github.com/jddixon/hamt_go"
	xr "github.com/jddixon/rnglib_go"
	//"io/ioutil"
	"os"
	//"path"
	"runtime/pprof"
	//"strings"
)

var _ = errors.New

func Usage() {
	fmt.Printf("Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Printf("where the options are:\n")
	flag.PrintDefaults()
}

const ()

var (
	// these need to be referenced as pointers
	cpuProf		  = flag.String("c", "", "cpuprofile file name")
	memProf		  = flag.String("m", "", "memprofile file name")
	
	justShow      = flag.Bool("j", false, "display option settings and exit")
	showTimestamp = flag.Bool("t", false, "output UTC timestamp")
	showVersion   = flag.Bool("V", false, "output package version info")
	testing       = flag.Bool("T", false, "test run")
	usingSHA1     = flag.Bool("1", false, "test run")
	verbose       = flag.Bool("v", false, "be talkative")
)

// -- utilities -----------------------------------------------------

// Create N random-ish K-byte values.  These are to be used as BytesKeys,
// so the first 64 bits must represent a unique value.

func makeSomeUniqueKeys(N, K int) (rawKeys [][]byte, bKeys []*gh.BytesKey) {

	rng := xr.MakeSimpleRNG()
	rawKeys = make([][]byte, N)
	bKeys = make([]*gh.BytesKey, N)
	keyMap := make(map[uint64]bool)

	for i := 0; i < N; i++ {
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

func  doBenchmark(w, t uint, N int, cpuProfFile *os.File) {
	// build an array of N random-ish K-byte rawKeys
	K := 16
	// t0 := time.Now()
	rawKeys, bKeys := makeSomeUniqueKeys(N, K)
	//t1 := time.Now()
	//deltaT := t1.Sub(t0)
	//fmt.Printf("setup time for %d %d-byte rawKeys: %v\n", N, K, deltaT)

	pprof.StartCPUProfile(cpuProfFile)
	m := gh.NewHAMT(w, t)

	for i := 0; i < N; i++ {
		_ = m.Insert(bKeys[i], &rawKeys[i])
	}

	// verify that the rawKeys are present in the map
	//for i := 0; i < N; i++ {
	//	value, err := m.Find(bKeys[i])
	//	// DEBUG
	//	if err != nil {
	//		fmt.Printf("error finding key %d\n", i, err.Error())
	//	}
	//	if value == nil {
	//		fmt.Printf("cannot find key %d\n", i)
	//	}
	//	// END
	//	val := value.(*[]byte)

	//} // GEEP
}
// MAIN /////////////////////////////////////////////////////////////
func main() {
	var (
		cpuProfFile, memProfFile *os.File
		err             error
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
		fmt.Printf("cpuProf    	= %v\n", *cpuProf)
		fmt.Printf("memProf    	= %v\n", *memProf)
		fmt.Printf("justShow    	= %v\n", *justShow)
		fmt.Printf("showTimestamp   = %v\n", *showTimestamp)
		fmt.Printf("showVersion 	= %v\n", *showVersion)
		fmt.Printf("testing     	= %v\n", *testing)
		fmt.Printf("usingSHA1       = %v\n", *usingSHA1)
		fmt.Printf("verbose     	= %v\n", *verbose)
	}
	// DO IT ////////////////////////////////////////////////////////
	if err == nil && !*justShow {
		cpuProfFile, err = os.Create(*cpuProf)
		if err == nil {
			defer cpuProfFile.Close()
			memProfFile, err = os.Create(*memProf)
			if err == nil {
				defer memProfFile.Close()
			}
		}
		if err == nil {
			//         w , t , N
			doBenchmark(6, 16, 1000000, cpuProfFile) 
		}
		if err == nil {
			if memProfFile != nil {
				err = pprof.WriteHeapProfile(memProfFile)
			}
			pprof.StopCPUProfile()
		}
	}

	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(-1)
	}

}
