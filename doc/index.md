# hamt_go

Go implementation of the HAMT data structure.

This code is not thread-safe.  That is, using code must provide any
necessary locking.

## Project Status

The code works and is reasonably well-tested.  A test run which includes
32K inserts, 32K deletes, and at least 128K finds takes just over a second.`
