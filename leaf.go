package hamt_go

// hamt_go/leaf.go

type Leaf struct {
	Key   KeyI
	Value interface{}
}

func NewLeaf(key KeyI, value interface{}) (leaf *Leaf, err error) {
	if key == nil {
		err = NilKey
	} else if value == nil {
		err = NilValue
	} else {
		leaf = &Leaf{
			Key:   key,
			Value: value,
		}
	}
	return
}

func (leaf *Leaf) IsLeaf() bool { return true }
