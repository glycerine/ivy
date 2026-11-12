# CUDD Go port notes

This file records behavior that differs from the corresponding C implementation
in `src/bdd/cudd`. Keep these differences explicit until the native algorithms
are ported and their C-oracle results are tested.

## Decomposition (`cuddDecomp.go`)

`Cudd_bddVarConjDecomp` follows the C identity directly: for a support variable
`x`, it forms `(f OR x)` and `(f OR !x)`, picks the pair with the smallest sum
of DAG sizes, and returns one factor if either factor is constant one. The
disjunctive form complements the input and resulting factors, as the C code
does. Results are semantically exact and are referenced through `Cudd_Ref`.

`Cudd_bddApproxConjDecomp`, `Cudd_bddIterConjDecomp`, and
`Cudd_bddGenConjDecomp` currently use that variable split. The C implementations
are different: the approximate and iterative variants use remap overapproximation,
squeeze, and LI compaction; the generalized variant uses `cuddConjunctsAux` and
bottom-up factor selection. Their exact factor choice and size heuristics are
therefore not preserved yet. The disjunctive forms retain the C complement
duality. The Go APIs return factors through `*[]*DdNode` so the caller owns a
garbage-collected slice instead of a C-allocated pointer array.

## Approximation (`cuddApprox.go`)

Under-approximation preserves the subset guarantee and over-approximation uses
complement duality. When reduction is requested, the Go implementation chooses
a satisfying cube by following the branch with the larger minterm count. It
does not reproduce the C density, safety, remapping, bias quality, or node
replacement heuristics. Biased under-approximation preserves `f` where the bias
is true and approximates only the remainder; this preserves the documented
correctness condition but not C's quality choices.

## Clipping (`cuddClip.go`)

Clipping returns exact `And` and `AndAbstract` results for either direction.
This preserves the requested under/over bound, but does not enforce `maxDepth`
or reproduce the C clipping heuristic.

## Correlation (`cuddBddCorr.go`)

Unweighted and weighted correlation are computed from the exact XOR probability
recurrence; the C hash-table recursion and OOM sentinel behavior are not
preserved. Invalid inputs return `NaN` in Go.

## Cache and checks (`cuddCache.go`, `cuddCheck.go`)

The computed table is a Go map with operation/operand keys and whole-table
flush at the configured hard limit; CUDD's direct-mapped cache collision and
resize heuristics are not preserved. `Cudd_DebugCheck` validates canonical
table keys, child ownership/order, complemented-edge pairs, and permutations;
it does not reproduce CUDD's per-subtable dead-reference accounting.

## Essential variables and clauses (`cuddEssent.go`)

Essential literals are found by testing the implication `f => literal`. The
two-literal clause list enumerates implied one- and two-literal clauses, then
omits pairs subsumed by an implied one-literal clause. The native C code uses
a recursive memoized enumeration and stores clauses in packed arrays; the Go
version uses explicit clause records and slice-based name input/output.

## Exact reordering (`cuddExact.go`)

Exact reordering enumerates every permutation in the target level range and
selects the one with the fewest canonical BDD nodes. To bound factorial work,
ranges wider than eight variables return failure without changing the manager.
The C implementation uses branch-and-bound plus symmetry information and can
search wider ranges; this implementation has the same exact objective within
its supported range but not the same bound or search strategy.

## Genetic reordering (`cuddGenetic.go`)

The genetic operator uses a fixed-seed population and PMX crossover, evaluates
node counts after rebuilding each candidate order, and installs the best order.
It follows the source's population/crossover model but omits its preliminary
sifting, duplicate-count weighting, and source random-number stream.

## Variable groups (`cuddGroup.go`, `cuddZddGroup.go`)

Tree-node creation maps variable indices through the current BDD/ZDD order and
uses the native `mtr` group tree. The source tree-sifting traversal is not yet
ported; its CDD reorder routines do not currently honor nested MTR constraints.

## Export (`cuddExport.go`)

DOT and BLIF use regular node IDs and explicit complemented-edge styles/inverter
signals. Export functions accept Go root/name slices and `io.Writer` rather
than C pointer arrays and `FILE *`. Factored form emits the Shannon expansion.
DaVinci and DDcal currently use readable expression output rather than the
native serializers, so consumers requiring those exact file grammars need a
follow-up port.

## Generalized cofactors (`cuddGenCof.go`)

BDD constrain/restrict use the native core algorithms. The BDD constrain
decomposition and characteristic-to-vector recursion also follow the source
recurrences. ADD constrain/restrict currently return the exact input ADD, which
is correct on the care set but does not minimize nodes. LI compaction currently
uses Restrict, and squeeze returns the lower bound after checking `lower <=
upper`; these preserve the contracts but not the C edge-marking/size heuristics.
# Additional native ports

## Harwell–Boeing matrices (`cuddHarwell.go`)

`Cudd_addHarwell` now accepts an `io.Reader` and returns a `HarwellMatrix` or
an error. It parses the supported real unsymmetric assembled (`RUA`) form,
validates column pointers and row indices, creates arithmetic row/column
indicator literals, and builds the sparse matrix ADD. The older C interface
mutated several pointer-to-pointer arguments; Go returns those values together
in a result struct instead. Sparse right-hand sides are accepted only in the
full (`F`) form. Other matrix forms fail with an explanatory error.

## Variable interaction (`cuddInteract.go`)

The interaction bitset is stored by the core manager. Rebuilding it traverses
the support of each canonical BDD node and marks every pair of variables in
that support. This can conservatively retain extra interactions from
unreferenced canonical nodes, which may make linear reordering do extra work;
it cannot incorrectly skip a dependent swap.

## Generator state (`cuddInt.go`)

The C union used by cube, prime, and node generators is represented by a Go
struct with named variants and slices. This avoids C allocation and pointer
arithmetic while keeping each generator's state separate. The C header's
platform macros and compiler attributes are omitted; generator limits and
semantic constants used by the Go algorithms remain explicit.
