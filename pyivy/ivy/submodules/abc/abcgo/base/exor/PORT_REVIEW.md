# EXORCISM port and review

The package contains the real exclusive sum-of-products minimizer, rather than an XOR gate placeholder. Its seven original algorithm groups are implemented as packed cube primitives, manager-owned storage, ExorLink generation, cover/pair iteration, search, costs, and ESOP serialization. `graph.go` connects the implementation to the committed native AIG API.

## Representation and API

`Manager` replaces the C singleton cover and static iterator/scratch globals. Input cubes retain the original two-bit NEG/POS/ABS value sets; output words hold GF(2) coefficients. Public calls take an explicit manager, native slices and errors instead of process-global state, Vec wrappers, allocation-success integers or FILE pointers. `Exorcism`/`Abc_ExorcismMain` return the minimized manager and optionally write a named file; a blank path performs no filesystem write. Original entry-point names for cube manipulation, pool/queue operations, ExorLink iteration, cost calculation and cover reduction remain available with these adapted signatures. Table initialization is replaced by `math/bits`; obsolete C bit-count lookup initialization is unnecessary.

ExorLink uses the same algebraic identities as C, generating every differing-dimension permutation instead of storing handwritten rule/group tables. An output-mask difference participates as one dimension. Distance two, three and four generate 2, 6 and 24 groups respectively, with matching complete groups verified against the original C implementation.

The deterministic minimizer first cancels equal products and merges distance-one cubes. It then searches distance-two/three/four replacements, normalizes each candidate against the entire cover, and commits strict improvements. Quantum-cost mode prioritizes quantum cost; otherwise cube count and then literals lead. Quality bounds additional neutral exploration; unsuccessful exploration is rolled back. This preserves exact ESOP semantics, while the search order and quality effort differ from the original mutable circular-queue heuristic. No claim of identical final covers or global optimality is made.

AIG export constructs exact AND/XOR networks. Import uses an exact Boolean Möbius transform to obtain positive-polarity Reed-Muller terms, with a caller-specified input limit because truth tables grow exponentially. Sequential graphs are rejected. XNOR and constants naturally use the constant product term. `Abc_ExorcismNtk2Esop` now accepts the AIG graph and explicit truth-table budget, avoiding a dependency on the still-porting base/abc network package.

## Separate design and readability review

The packed representation keeps the source identities recognizable. Explicit manager ownership removes global scratch races. Generic permutations replace the large ExorLink rule tables; one normalization path handles cancellation in direct insertion, optimization and trial rewrites. Pool ownership and pair iterators are explicit. Standard I/O and errors replace C allocation and stream conventions. No external or unsafe dependencies remain. Search complexity remains quadratic in cube pairs with candidate normalization; large ESOPs can require substantial work, just as exhaustive truth-table import requires an explicit budget.

## Separate bug review

Reviewed multi-output coefficients, negative/absent literal XOR identities, packed-word boundaries, duplicate cancellation, distance-one merging, quality rollback/termination, stale pair ownership, ANF constant phase, malformed PLA data and I/O failures.

Three concrete findings were first reproduced by a failing regression, then fixed:

1. Writers accepted partial writes returning no error. All emitted text now checks byte counts and returns `io.ErrShortWrite`.
2. Normalization replaced cube objects while old queue handles remained active. Replaced objects are now invalidated before constructing the normalized cover.
3. PLA input without an explicit ESOP type could silently reinterpret ordinary SOP as XOR. The reader requires `.type esop`.

Caller ESOP rows are copied instead of having their output marker destructively popped. Contradictory literals, wrong dimensions, padding bits, invalid value sets, negative options and truth-table budget violations are rejected.

## Validation

- Every pair of cubes over up to three inputs, all nonempty two-output coefficient masks, and every ExorLink permutation: exhaustive truth comparison.
- All 256 three-input Boolean functions, plus their complements: AIG → ANF ESOP → minimization → AIG truth comparison.
- Eighty seeded random multi-output covers: exact truth preservation, nonworsening cost and fixed-point checks.
- Original-C differential oracle: 300 seeded cases comparing bounded distance, quantum cost and all ExorLink groups, including output-difference identities.
- Packed 16-variable/32-output boundaries, pool/pair iterator behavior, undo, caller ownership, PLA round trips and malformed/short-write regressions.

`GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./base/exor` passes (4.885 seconds). `go vet ./base/exor` passes. The C differential fixture is optional when a C compiler or supported ELF linker is unavailable; the production package remains pure Go.
