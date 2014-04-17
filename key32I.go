package hamt_go

// hamt_go/key32I.go

// A Key32 is anything that returns an unsigned 32-bit value.
type Key32I interface {
	Hashcode32() (uint32, error)
}
