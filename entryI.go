package hamt_go

type Entry32I interface {
	IsLeaf() bool
}

type Entry64I interface {
	IsLeaf() bool
}
