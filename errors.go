package hamt_go

import (
	e "errors"
)

var (
	DeleteFromEmptyTable     = e.New("Internal Error: delete from empty table")
	MaxTableDepthExceeded    = e.New("max Table depth exceeded")
	MaxTableSizeExceeded     = e.New("max Table size (w=6) exceeded")
	MaxRootTableSizeExceeded = e.New("max Root table size (t=64) exceeded")
	NilKey                   = e.New("nil key parameter")
	NilRoot                  = e.New("nil root parameter")
	NilValue                 = e.New("nil value parameter")
	NotFound                 = e.New("entry not found")
	ShortKey                 = e.New("Bytes*Key is too short")
	ZeroLengthTables         = e.New("Cannot create: zero length tables")
)
