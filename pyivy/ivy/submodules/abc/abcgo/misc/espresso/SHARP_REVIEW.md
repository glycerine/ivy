# Sharp/intersection family review

All source sharp.c functions are ported: cover/cube sharp, balanced recursive
sharp, disjoint-sharp in both argument orientations, make_disjoint, primitive
disjoint difference, and cover intersection. Original API names and argument
conventions remain. cv_difference/sccover do not occur in this source family;
no speculative replacement interfaces were introduced.

## Compactness/readability pass

Reviewed separately from correctness. Recursion uses bounded integer ranges and
balanced halves. Cube scratch storage is local, so composing operations cannot
corrupt a caller's borrowed temporary. Intersection retains the source's
500-cube containment batches with one flush helper. Go-managed cover families
retain the existing ownership contract; source manual-free timing globals and
debug-only progress output were unnecessary and omitted. Packed-cube operations
remain proportional to variables/words, without minterm expansion.

## Bug-review pass

Audited zero-size and singleton covers, disjoint/containing/equal cubes, borrowed
input immutability, multi-valued coordinates, recursion endpoints, batch boundary
and final-empty-batch behavior. Midpoint arithmetic avoids source first+last
overflow. Loops explicitly use active num_vars rather than the entire cached
var_mask slice, preserving correctness after cubehack shrinks a 100-variable
cache; a directed shrink regression covers this interaction. Disjoint difference
updates its accumulated mask only when a difference coordinate exists, matching
the original construction. Pairwise disjointness is checked independently.

The source's sf_union contract assumes sorted lists with cardinality metadata;
sharp's primitive callers do not consistently supply that precondition. As a
result equally valid redundant row counts/order can vary between C and Go.
Differential checks compare represented Boolean functions, rather than assuming
an accidental row ordering is required. Direct primitive and cover properties
independently check the intended Boolean difference/intersection/disjointness.

## Validation

* All 3,969 nonempty mixed-valued primitive cube pairs: ordinary/disjoint sharp
  equal A AND NOT B; disjoint rows are pairwise nonintersecting.
* 120 deterministic random cover cases: both differences, intersection and
  disjointization checked at every minterm.
* 120 original-C fixtures (reproducer build_sharp_oracle.py): six cover/cube
  operations compared against original C at every minterm.
* 900 raw intersections exercise multiple 500-row flushes and duplicate
  containment; cached-geometry shrink exercises inactive masks.

Combined real-source geometry/cofactor/complement/irred/essential/equivalence/
sharp and supporting cover/sparse/mincov subset passes tests, race tests and vet.
