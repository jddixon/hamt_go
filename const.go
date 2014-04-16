package hamt_go

const (
	VERSION      = "0.1.1"
	VERSION_DATE = "2014-04-12"
)

const (
	W32          = uint(5) // log base 2 of number of entries in a table
	MAX_DEPTH32  = (32 / W32)
	LEVEL_MASK32 = 0x1f // masks off W bits
)
