package hamt_go

import (
	e "errors"
)

var (
	NilKey   = e.New("nil key parameter")
	NilValue = e.New("nil value parameter")
	ShortKey = e.New("Bytes*Key is too short")
)
