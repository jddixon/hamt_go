package hamt_go

// hamt_go/bytesKey.go

import (
	"fmt"
)

var _ = fmt.Print

type BytesKey struct {
	Slice []byte
}

func NewBytesKey(b []byte) (k *BytesKey, err error) {
	if b == nil {
		err = NilKey
	} else if len(b) < 8 {
		err = ShortKey
	} else {
		k = &BytesKey{Slice: b}
	}
	return
}

// KeyI interface ///////////////////////////////////////////////////

// Convert the first 8 bytes of the key into an unsigned uint64.
// We are guaranteed that len(b.Slice) is >= 8, so error return is unneeded.
func (b *BytesKey) Hashcode() (hc uint64) {
	//buf := bytes.NewReader(b.Slice)
	// err = binary.Read(buf, binary.LittleEndian, &hc
	// XXX Calculating this here makes the code run about 10% faster)
	s := b.Slice
	hc = uint64(s[0]) +
		 uint64(s[1]) <<  8 +
		 uint64(s[2]) << 16 +
		 uint64(s[3]) << 24 +
		 uint64(s[4]) << 32 +
		 uint64(s[5]) << 40 +
		 uint64(s[6]) << 48 +
		 uint64(s[7]) << 56 
		 
	return
}
