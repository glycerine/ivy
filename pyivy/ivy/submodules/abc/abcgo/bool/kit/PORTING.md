# Go computation kit

The package is a pure-Go port of `src/bool/kit`, including the local Cloud BDD
engine. All product `.go` files participate in the normal package build. There
are no cgo dependencies, source-exclusion tags, fallback stubs or alternate test
implementations. All original `Kit_*` function names are present; the C iterator
macros have Go `iter.Seq2` counterparts.

## Representations and API adaptations

Truth tables are `[]uint32`, with bit zero representing the all-zero assignment.
Functions with at most five variables use one word and replicate their truth
pattern through that word, matching ABC. Most dimensions retain `int32` for
readable correspondence with the source. Permutations are mutable `[]byte`.
C buffers become slices, result/scratch vectors become pointers to slices, and
PLA text is an immutable string. `Kit_PlaComplement` returns the changed string.
The old opaque allocation-pool arguments on PLA constructors are compatibility
parameters; Go owns the returned strings.

Graph nodes have stable Go addresses, complemented edges have explicit fields,
and BDD complement pairs replace tagged pointers. The former optional CUDD
adapter functions use the local canonical Cloud manager. The AIG and GIA graph
bridges accept manager interfaces, avoiding an import cycle when AIG algorithms
use Kit. Hop bridging uses the separately ported Go Hop package.

DSD retains ABC's recursive output/input AND, XOR and MUX decomposition. ISOP
retains its three recursive subproblems and polarity comparison. SOP factoring
retains kernel extraction, algebraic division and balanced trivial factoring.
Automatic memory management replaces C scratch-pointer ownership and unions.
Debug helpers use Go file I/O; `Kit_TruthTest` and `Kit_DsdPrecompute4Vars` return
I/O errors. `Kit_DsdVerify` additionally returns its verification result.

The original experimental decomposition-pattern code contained incorrect
cofactor offsets/uniqueness checks, and its shared-variable recursion lacked a
base case and never inspected the input pattern. The Go functions implement the
stated operations explicitly: equal-cofactor labeling and a bounded exhaustive
search for the smallest shared-variable partition whose groups need at most one
bound-function bit. This behavior has direct tests.

## Verification and review

Run from `abcgo`:

```
go test -race ./bool/kit -count=1
go vet ./bool/kit
```

Tests include every three-variable function through ISOP/factoring, DSD,
PLA and the real Hop bridge; random multiword truth/care sets; DSD round trips
through 16 variables; expansion/shrink and partial composition; packed Cloud
programs, canonical BDD operations, support, cubes, bounds and restart; LUT
networks and timing; permutation, phase, quantification and hash primitives;
original API/data inventory; and source-coordinate-independent file handling.
The GIA worker additionally verified the actual `Gia_Man_t` bridge with manager
objects and exhaustive small assignments under the race detector.

Two independent passes were coordinated before edits: `misc_avl` reviewed
compactness/idiomatic representation/readability, and `misc_st` reviewed bugs.
The independent bug findings and red-to-green regressions are recorded in
`BUG_REVIEW.md`. The compactness pass removes redundant slice length/capacity
state, separates the ISOP failure flag from cover length, and uses bounded Go
cache state. Internal review also reproduced and fixed a cofactor-count helper
that accidentally mutated its input and doubled the count, restored the ABC
script dump payload, and restored all pairwise truth-profile metrics.
