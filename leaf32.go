package hamt_go

// hamt_go/leaf32.go

type Leaf32 struct {
	Key   Key32I
	Value interface{}
}

func NewLeaf32(key Key32I, value interface{}) (leaf32 *Leaf32, err error) {
	if key == nil {
		err = NilKey
	} else if value == nil {
		err = NilValue
	} else {
		leaf32 = &Leaf32{
			Key:   key,
			Value: value,
		}
	}
	return
}

func (l32 *Leaf32) IsLeaf() bool { return true }

//type Leaf64 struct {
//	Key   Key64I
//	Value interface{}
//}
//
//func (l64 *Leaf64) IsLeaf() bool { return true }

//func NewLeaf64(key Key64I, value interface{}) (leaf64 *Leaf64, err error) {
//	if key == nil {
//		err = NilKey
//	} else if value == nil {
//		err = NilValue
//	} else {
//		leaf64 = &Leaf64{
//			Key:   key,
//			Value: value,
//		}
//	}
//	return
//}
//
