package hamt_go

// The 32-bit SWAR algorithm.  These are variants of the code in Bagwell's
// "Ideal Hash Trees".  The algorithm seems to have been created by the
// aggregate.org/MAGIC group at the University of Kentucky earlier than
// the fall of 1996.  Illmari Karonen (vyznev.net) explains the algorithm at
// stackoverflow.com/questions/22081738/how-variable-precision-swar-algorithm-workds

const (
	OCTO_FIVES  = 0x55555555
	OCTO_THREES = 0x33333333
	OCTO_ONES   = 0x01010101
	OCTO_FS     = 0x0f0f0f0f

	HEXI_FIVES  = 0x5555555555555555
	HEXI_THREES = 0x3333333333333333
	HEXI_ONES   = 0x0101010101010101
	HEXI_FS     = 0x0f0f0f0f0f0f0f0f
)

func BitCount32(n uint32) uint {
	n = n - ((n >> 1) & OCTO_FIVES)
	n = (n & OCTO_THREES) + ((n >> 2) & OCTO_THREES)
	return uint((((n + (n >> 4)) & OCTO_FS) * OCTO_ONES) >> 24)
}

func BitCount64(n uint64) uint {
	n = n - ((n >> 1) & HEXI_FIVES)
	n = (n & HEXI_THREES) + ((n >> 2) & HEXI_THREES)
	return uint((((n + (n >> 4)) & HEXI_FS) * HEXI_ONES) >> 56)
}
