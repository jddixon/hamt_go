package hamt_go

// hamt_go/key64I.go

// A Key64 is anything that returns an unsigned 64-bit value.
type Key64I interface {
	Hashcode64() (uint64, error)
}
