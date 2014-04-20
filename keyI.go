package hamt_go

// hamt_go/keyI.go

// A Key is anything that returns an unsigned 64-bit value.
type KeyI interface {
	Hashcode() (uint64, error)
}
