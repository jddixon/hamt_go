# hamt_go

Go implementation of the HAMT data structure.

A Hash Array Mapped Trie ([HAMT][bagwell2002]) 
provides fast and memory-efficient access to large amounts of data held 
in memory.  In this 
particular implementation there are 32- and 64-bit versions
of the code, HAMT32 and HAMT64 respectively.  HAMT associates keys and values.  The **key** is 
something that produces a 32- or 64-bit unsigned integer, as appropriate.
The value is completely arbitrary, but typically is a pointer to a data
structure of interest.

HAMT keys have a `Hashcode` function that produces a fixed length integer
that is treated as a sequence of bits.  This is used to navigate and 
modify the in-memory trie structure.  At each level in the trie there is
a table of `2^5` or `2^6` slots.  Each slot contains either a nil pointer
or a pointer to a value or a pointer to a lower-level table.  So a HAMT32
table is rougly `32 * 4 = 128` bytes in size, and a HAMT64 table about
`64 * 8 = 512` bytes.  Whereas a normal hash table would be quite large and
require periodic expensive resizing, the HAMT data structure is roughly 
as fast as a hash table, but starts small and consumes more memory only 
as needed.

## Limitations

* This code is not thread-safe.  That is, using code must provide any
necessary locking.

* HAMT depends upon bit-counting.  On modern Intel and AMD processors this 
can be done using a specific machine-language instruction, POPCNT.  The current
implementation of hamt_go emulates this in software using the 
[SWAR][wiki-swar] algorithm,  The emulation code is on the order of ten times
slower than the machine instruction.  

In practice POPCNT emulation might slow down accesses by something like 10%, 
because the emulation code simply is not run all that often.

## Project Status

Both 32- and 64-bit versions of the code work and are reasonably well-tested. 
Insert, find, and delete operations, while not yet thoroughly optimized, 
take on the order of 10 microseconds each on a lightly-loaded server.

## References

[Bagwell, "Ideal Hash Trees"][bagwell2002]  (2002 PDF)

[Wikipeida, "Hash array mapped trie"][wiki-hamt]

[Wikipedia, "SWAR"][wiki-swar]


[bagwell2002]: http://infoscience.epfl.ch/record/64398/files/idealhashtrees.pdf

[wiki-hamt]: http://en.wikipedia.org/wiki/Hash_array_mapped_trie

[wiki-swar]: http://en.wikipedia.org/wiki/SWAR
