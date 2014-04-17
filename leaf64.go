package hamt_go

// hamt_go/leaf64.go

type Leaf64 struct {
	Key   Key64I
	Value interface{}
}

func NewLeaf64(key Key64I, value interface{}) (leaf64 *Leaf64, err error) {
	if key == nil {
		err = NilKey
	} else if value == nil {
		err = NilValue
	} else {
		leaf64 = &Leaf64{
			Key:   key,
			Value: value,
		}
	}
	return
}

func (l64 *Leaf64) IsLeaf() bool { return true }
