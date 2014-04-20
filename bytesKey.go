package hamt_go

// hamt_go/bytesKey.go

import (
	"bytes"
	"encoding/binary"
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

// convert the first 8 bytes of the key into an unsigned uint64
func (b *BytesKey) Hashcode() (hc uint64, err error) {
	buf := bytes.NewReader(b.Slice)
	err = binary.Read(buf, binary.LittleEndian, &hc)
	// DEBUG
	if err != nil {
		fmt.Printf("attempt to read key failed: %v\n", err)
	}
	// END
	return
}
