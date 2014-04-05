package hamt_go

// hamt_go/hamt.go

type Table32 struct {
	bitmap uint32
	slots  []Entry32I // nil or a key-value pair or a pointer to a table
}

func (t32 *Table32) IsLeaf() bool { return false }

type Table64 struct {
	bitmap uint64
	slots  []Entry64I // nil or a key-value pair or a pointer to a table
}

func (t64 *Table64) IsLeaf() bool { return false }
