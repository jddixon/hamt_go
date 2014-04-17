package hamt_go

// ctries_go/bytes64Key.go

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var _ = fmt.Print

type Bytes64Key struct {
	Slice []byte
}

func NewBytes64Key(b []byte) (k *Bytes64Key, err error) {
	if b == nil {
		err = NilKey
	} else if len(b) < 16 {
		err = ShortKey
	} else {
		k = &Bytes64Key{Slice: b}
	}
	return
}

// KeyI interface ///////////////////////////////////////////////////

// convert the first 8 bytes of the key into an unsigned uint64
func (b *Bytes64Key) Hashcode64() (hc uint64, err error) {
	buf := bytes.NewReader(b.Slice)
	err = binary.Read(buf, binary.LittleEndian, &hc)
	if err != nil {
		fmt.Printf("attempt to read key failed: %v\n", err)
	}
	return
}
