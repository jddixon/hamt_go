package hamt_go

import (
	"fmt"
	"strings"
)

// The 32- and 64-bit versions of the SWAR algorithm.  These are variants of
// the code in Bagwell's "Ideal Hash Trees".  The algorithm seems to have been
// created by the aggregate.org/MAGIC group at the University of Kentucky
// earlier than the fall of 1996.  Illmari Karonen (vyznev.net) explains the
// algorithm at
// stackoverflow.com/questions/22081738/how-variable-precision-swar-algorithm-workds

const (
	OCTO_FIVES  = uint32(0x55555555)
	OCTO_THREES = uint32(0x33333333)
	OCTO_ONES   = uint32(0x01010101)
	OCTO_FS     = uint32(0x0f0f0f0f)

	HEXI_FIVES  = uint64(0x5555555555555555)
	HEXI_THREES = uint64(0x3333333333333333)
	HEXI_ONES   = uint64(0x0101010101010101)
	HEXI_FS     = uint64(0x0f0f0f0f0f0f0f0f)
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

func dumpByteSlice(sl []byte) string {
	var ss []string
	for i := 0; i < len(sl); i++ {
		ss = append(ss, fmt.Sprintf("%02x ", sl[i]))
	}
	return strings.Join(ss, "")
}

//func dumpTable(title string, dTable *Table) string {
//	var ss []string
//	ss = append(ss, fmt.Sprintf("    dTable %s:", title))
//	for j := 0; j < len(dTable.slots); j++ {
//		if dTable.slots[j].IsLeaf() {
//			ss = append(ss, " L")
//		} else {
//			ss = append(ss, " T")
//		}
//		ss = append(ss, fmt.Sprintf("%02x", dTable.indices[j]))
//	}
//	return strings.Join(ss, "")
//}
