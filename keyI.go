package hamt_go

// hamt_go/keyI.go

// A Key32 is anything that returns an unsigned 32-bit value.
type Key32I interface {
	Hashcode32() (uint32, error)
}

// A Key64 is anything that returns an unsigned 64-bit value.
type Key64I interface {
	Hashcode64() (uint64, error)
}
