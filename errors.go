package hamt_go

import (
	e "errors"
)

var (
	DeleteFromEmptyTable = e.New("Internal Error: delete from empty table")
	MaxTableSizeExceeded = e.New("max table size (64) exceeded")
	NilKey               = e.New("nil key parameter")
	NilValue             = e.New("nil value parameter")
	NotFound             = e.New("entry not found")
	ShortKey             = e.New("Bytes*Key is too short")
	ZeroLengthTables     = e.New("Cannot create: zero length tables")
)
