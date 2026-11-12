# Dynamic reordering and transfer review

## Compactness, idiom and readability audit

The active implementation is `cuddReorder.go`, with family-preserving transfer in
`cuddTransfer.go`. Root entry points forward to these algorithms rather than
maintaining duplicate implementations. It implements SIFT, SIFT_CONVERGE,
SYMM_SIFT and SYMM_SIFT_CONV, plus ShuffleHeap, profile metadata, counters and
sifting limits. Unsupported heuristic values return failure explicitly.

The original subtable pointer surgery is replaced with canonical scratch
rebuilding followed by one handle-preserving transplant. Each phase is named:
collect live referenced roots, rebuild candidate order, measure live heap, detect
symmetry, sift groups, and commit. Per-pass level indexing avoids repeated linear
lookups during population sorting. Temporary candidate nodes are isolated from
the original manager's unique tables and become ordinary Go garbage.

Referenced roots, projection variables, constants and complementary handles
remain stable. The transplant relies on a canonical invariant, not unsafe casts:
a regular BDD is true at all-ones; changing variable order does not change that
polarity. Unique tables and computed caches are replaced/cleared only at commit.
Like original CUDD reorder preprocessing, unreferenced unreachable BDD/ADD nodes
are collected. Callers must reference roots they retain across dynamic reorder.
ZDD tables/order/universe remain separate and untouched.

Symmetric sifting performs real positive and negative cofactor checks together
with incoming-arc/reference accounting. It ignores isolated projection pins and
keeps discovered groups intact when moving them. Symmetry metadata is exposed
through cloned snapshots. This is not a plain-SIFT alias.

Scope: default full BDD/ADD heap ordering needed by REO is implemented. Other
CUDD heuristics, persistent MTR grouping/bound-variable constraints and automatic
allocation-triggered reordering remain separate unported CUDD facilities; this
work does not manufacture those APIs or claim them complete.

## Separate bug audit and fixes

1. Initially followed the source public comment claiming symmetric success is
   `1 + number of symmetric variables`. Independent C execution showed the
   actual tree dispatcher normalizes the public result to **1**. Fixed public
   return behavior; real discovered groups are inspected separately.
2. Included explicitly referenced constants in the live-root set; otherwise a
   terminal-only manager incorrectly skipped a real reordering due to minsize.
3. Cleared obsolete symmetry metadata after ordinary sifting and nonidentity
   shuffle. Zero variable/swap limits prevent discovery/movement.
4. Verified transplant handles and external references, full truth tables,
   canonical operations after order changes, cache clearing, dead-node collection,
   permutation bijections, convergence, configured limits and invalid inputs.
5. Preserved existing referenced ZDD membership, unique identity, counts,
   permutation/universe and shared One/Zero handles through both methods and
   explicit shuffle.
6. Tested cross-manager BDD/ADD transfers, numeric terminals, permutations and
   inverse round-trips, foreign/closed managers and invalid mappings.

## Independent original-C evidence

`test_vectors/reorder_oracle.c` links actual original ABC CUDD object files. Its
fixture covers every three-variable truth table under both SIFT and SYMM_SIFT:
512 cases. The Go regression matches exact initial/final DAG sizes, public
status, all output truth tables, grouped-variable masks and group counts.
Variable-order ties are not compared: the original uses unspecified qsort ties,
while Go uses stable level order. No C compiler/runtime is needed for Go tests.

Additional deterministic random five-variable referenced families are checked
exhaustively, with BDD and ADD handle preservation and inverse canonicality.
Run `go -C abcgo test ./bdd/cudd/core ./bdd/reo -race -count=1`.

Final source audit additionally found the original initializer explicitly pins
One, Zero and both infinities. Included these and the background in heap-cost
offsets, so maxGrowth pruning uses the same baseline rather than only constants
reachable from user roots. Dead unreferenced numeric terminals are collected,
while referenced terminals/global pins retain identity. Regression checks the
exact five-terminal cost with four distinct manager pins plus one user terminal.
The all-512 original-C size/status/truth/symmetry oracle remains green after
this correction. Reordering time includes initial live-heap preprocessing.
