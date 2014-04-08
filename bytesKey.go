package hamt_go

// ctries_go/bytesKey.go

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var _ = fmt.Print

type Bytes32Key struct {
	Slice []byte
}

type Bytes64Key struct {
	Slice []byte
}

func NewBytes32Key(b []byte) (k *Bytes32Key, err error) {
	if b == nil {
		err = NilKey
	} else if len(b) < 8 {
		err = ShortKey
	} else {
		k = &Bytes32Key{Slice: b}
	}
	return
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

// convert the first 4 bytes of the key into an unsigned uint32
func (b *Bytes32Key) Hashcode32() (hc uint32, err error) {
	buf := bytes.NewReader(b.Slice)
	err = binary.Read(buf, binary.LittleEndian, &hc)
	if err != nil {
		fmt.Printf("attempt to read key failed: %v\n", err)
	}
	return
}

// convert the first 8 bytes of the key into an unsigned uint64
func (b *Bytes64Key) Hashcode64() (hc uint64, err error) {
	buf := bytes.NewReader(b.Slice)
	err = binary.Read(buf, binary.LittleEndian, &hc)
	if err != nil {
		fmt.Printf("attempt to read key failed: %v\n", err)
	}
	return
}
