package hamt_go

const (
	VERSION      = "0.3.0"
	VERSION_DATE = "2014-04-21"
)

const (
	W32          = uint(5)    // log base 2 of number of entries in a table
	MAX_DEPTH32  = (64 / W32) // MAJOR CHANGE: 32 -> 64
	LEVEL_MASK32 = 0x1f       // masks off W32 bits

	W64          = uint(6) // log base 2 of number of entries in a table
	MAX_DEPTH64  = (64 / W64)
	LEVEL_MASK64 = 0x3f // masks off W64 bits
)
