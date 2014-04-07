package hamt_go

const (
	VERSION      = "0.1.1"
	VERSION_DATE = "2014-04-07"
)

const (
	W32          = uint(5) // log base 2 of number of entries in a table
	LEVEL_MASK32 = 0x1f    // masks off W bits
)
