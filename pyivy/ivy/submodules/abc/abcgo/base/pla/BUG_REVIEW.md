# Separate bug review

The original C parser, generator and hashing code were compiled without algorithm
changes. Native fixtures match Type, dimensions, every packed word, pairwise
and hashed distance counts. Random generation matches its own Gia_ManRandom
sequence exactly. The experimental fast-extractor native fixture also matches
its complete divisor vector and extracted cube lists.

Review fixes and regressions:

- Original writer emits '?' for FULL but its reader ignores '?'. The Go reader
  accepts it, preserving FDR output codes on read/write/read. Tests cover all four
  types and every literal code at widths1/31/32/33/64/65/129.
- The parser rejects malformed rows, unknown types, negative counts and unsupported
  multi-valued format with errors; comments, tabs, missing/mismatched .p and end
  markers have explicit tests. Reader and writer failures propagate.
- ConvertFromBits formerly treated FULL as a single positive literal. Both
  polarities now preserve the original packed code through conversion.
- The source simple-cover minimizer asserts on a complete square and, with asserts
  disabled, fails to save that cube. The valid four-minterm cube is now stored.
  All256 three-variable truth functions reconstruct exactly from the cover.
- DumpPla skipped empty literal lists, deleting the constant-one cube. The writer
  now retains it; an isolated file roundtrip regression checks the result.
- Native hash distance uses a boolean existence lookup, undercounting duplicate
  rows and potentially tripping its even-count assertion. A red test with
  minterms[0,0,1,3] observed2 hashed pairs versus3 enumerated pairs. Counting every
  matching row fixes multiplicity; the same test is green.
- Failed reads retain the current frame design. End clears its resource, so
  repeated End does not retain or free the same manager twice. Command tests
  cover registration, generation, read/write, stats, replacement and cleanup.

Further tests compare random minterm hashing with pairwise distance enumeration,
check primes, and evaluate300 random fast-extraction covers on every five-variable
assignment. Extraction retains Boolean meaning and does not increase total SOP
plus divisor literal cost. A dedicated case requires profitable extraction.

Passed full-package checks (including the distinct Espresso parser):

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./base/pla ./misc/espresso ./base/main
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./base/pla
```
