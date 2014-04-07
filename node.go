package hamt_go

// hamt_go/node.go

type Node32I interface {
	IsLeaf() bool
}

type Node64I interface {
	IsLeaf() bool
}
