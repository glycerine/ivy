# ZDD reordering correctness review

## Ported paths

- Adjacent swapping now operates on the two affected ZDD levels and retains
  existing node handles. Exhaustive three-variable truth tables check one swap
  against the represented set family and validate the resulting unique table.
- `Cudd_zddShuffleHeap` follows the C `zddShuffle` order: for each requested
  level, it moves that variable upward by adjacent swaps. Tests check nonidentity
  permutations, extra buffer elements, invalid permutations, counts, and handles.
- SIFT records each adjacent swap's heap size and reverses the recorded path to
  the closest best position. SIFT_CONVERGE repeats until the heap size stops
  decreasing. Boundary tests cover minsize, zero swap limits, and the reorder
  counter.
- RANDOM and RANDOM_PIVOT use CUDD's portable two-stream shuffled generator and
  `zddSwapAny`'s interleaved adjacent-swap route. Tests reseed and check family
  preservation and order validity.
- ZDD-to-BDD and BDD-to-ZDD alignment follow the source multiplicity-block
  mapping. Hook tests cover pre/post ordering and pre-hook cancellation.

## Unfinished source behavior

The core does not yet port SYMM_SIFT, SYMM_SIFT_CONV, GROUP_SIFT, LINEAR, or
LINEAR_CONVERGE, and does not apply ZDD group-tree bounds. `zddSymmCheck` is
ported and tested, but it is not integrated into a symmetric sifting driver.
The manager lacks CUDD's unique-table slot/refcount representation and garbage
collection hooks, so source-equivalent dead-node collection, table shrinking,
and swap failure rollback are also not established. The root `bdd/cudd` package
still has unrelated C-derived files that do not parse; package-level build
verification is currently blocked at `cuddLinear.go`.

Core verification command used:

```sh
GO111MODULE=off GOPATH=/tmp/abcgo-gopath go test github.com/glycerine/abcgo/bdd/cudd/core
```
