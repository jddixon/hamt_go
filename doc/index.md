# hamt_go

Go implementation of the HAMT data structure.

A Hash Array Mapped Trie ([HAMT][bagwell2001]) 
provides fast and memory-efficient access to large amounts of data held 
in memory.  Values are stored by *key*.  All HAMT keys are mapped into 
fixed length unsigned integers. In this implementation these are 64 bits
long, Go `uint64`s..  The *value* may be of any type but typically is a 
*pointer* to a data structure of interest.

The HAMT trie is essentially a prefix trie.  At each level
there is a table with `2^w` slots.  The `w` bits are construed as
an index into that table.  The current implementation allows values of
`4, 5, 6, 7,` or `8` for `w`.  Preliminary performance tests indicate 
that `w=5` is optimal.  That is, a table with 32 slots gives the best
performance.  

Whereas a normal hash table would be quite large and might
require periodic expensive resizing, the HAMT data structure is roughly 
as fast as a hash table, but starts small and consumes more memory only 
as needed.

## Limitations

* This code is not thread-safe.  That is, using code must provide any
necessary locking.

* the HAMT algorithm depends upon bit-counting.  On modern Intel and AMD 
processors this 
can be done using a specific machine-language instruction, POPCNT.  The current
implementation of hamt_go emulates this in software using the 
[SWAR][wiki-swar] algorithm,  The emulation code is on the order of ten times
slower than the machine instruction.  
*In practice POPCNT emulation might slow down accesses by something like 10%, 
because the emulation code simply is not run all that often.*

## Project Status

The code works and is reasonably well-tested. 
`Insert`, `Find`, and `Delete` operations, while not yet thoroughly optimized, 
take on the order of 1.5 microseconds each on a lightly-loaded server 
(just under 3us each to insert a million values and verify that the 
value can be found using the key.

## References

[Bagwell, "Ideal Hash Trees"][bagwell2001]  (2001 PDF)

[Wikipeida, "Hash array mapped trie"][wiki-hamt]

[Wikipedia, "SWAR"][wiki-swar]


[bagwell2001]: http://infoscience.epfl.ch/record/64398/files/idealhashtrees.pdf

[wiki-hamt]: http://en.wikipedia.org/wiki/Hash_array_mapped_trie

[wiki-swar]: http://en.wikipedia.org/wiki/SWAR
