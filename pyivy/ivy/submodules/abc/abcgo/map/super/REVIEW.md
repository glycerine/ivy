# Supergate precomputation (map/super)

`superGate.go` is a line-by-line port of `src/map/super/superGate.c`, and
`Super_CommandSupergates` in `super.c.go` of `super.c`. It keeps the C names,
the six nested enumeration loops (including the stale `Area` test of the
six-input case), and every source of ordering:

- the `stmm` table of `Super_ManStart` (`st__ptrhash` of `uTruth[0] ^
  uTruth[1]`, 11 bins, density 5, growth 2.0, new entries at the head of a
  bin, iteration bin by bin) as `stmm_table` in `superGate.go`;
- glibc `qsort` (a stable merge sort) as `sort.SliceStable`;
- `Extra_MmFixed` recycling: a dominated supergate removed by
  `Super_CompareGates` goes on a LIFO free list and the next
  `Super_CreateGateNew` reuses it in place, so the enumeration arrays and
  fanin pointers that still reference it see the new supergate, as in C;
- unmasked truth tables (an inverter over a variable has `uTruth[1] = ~0` for
  fewer than six inputs, which affects the hash key and `Super_WriteCompare`).

The written library is byte-identical to native ABC (header statistics,
padded line count and trailing NUL), as is the network `map` derives from it
(`base/abci/super_map_native_test.go`). As in C, `-o` is parsed but not
passed to `Super_Precompute`.

`superAnd.go` (`super2`) is still a rewrite of `superAnd.c`, not a port.
