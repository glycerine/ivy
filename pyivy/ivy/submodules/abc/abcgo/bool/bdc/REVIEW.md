# BDC integration review

## Compactness, idiomatic Go, and readability audit

Reviewed the core separately from its behavior tests. The port keeps the original
ISF support minimization, strong/weak OR split, complemented AND split, cached
component preference, and Shannon/MUX fallback algorithms. Truth words and nodes
use owned Go slices and objects; complemented references use explicit regular/
negative wrappers instead of pointer tagging. Client copies are opaque `any`
values, so the decomposition package does not depend on AIG graph types.

Removed the redundant cached truth-word count: it now derives from the current
variable count. Node count derives from the node slice. The new-node budget
counter is retained because it has separate algorithmic meaning from total
nodes (inputs/divisors are excluded and a reduced branch adjusts the count).
Table insertion now appends in constant time and lookup scans backward,
preserving the C newest-candidate-first policy without copying whole buckets.
Support checks no longer repeatedly compute complete support masks per variable;
the dedicated predicate checks the relevant word/cofactor pairs directly.

The independent SPFD family and its own review are in `bdcSpfd.go` and
`SPFD_REVIEW.md`, owned by its porting worker.

## Separate bug and semantic review

Reviewed phase handling, regular-node identity, table lookup order, onset/offset
containment, quantification, node-budget failure, divisors, and repeated manager
use. Fixed stale alternate client-copy state: setting an opaque copy clears the
integer representation and setting an integer copy clears the opaque reference,
matching the C union's replacement semantics. Added a regression for both paths.
Added a duplicate-divisor regression to preserve newest-first lookup order after
the bucket compactness improvement.

Compiled unchanged original `bdcCore.c`, `bdcDec.c`, `bdcTable.c`, and `kitTruth.c`
into a temporary native oracle and recorded exact gate counts for all 256
three-variable functions. The permanent Go regression checks those exact counts,
without requiring a C compiler at test time. Gate-graph simulation independently
checks every minterm for all 256 functions, randomized functions and care sets
from 0 to 8 variables, divisor reuse, budget exhaustion, and copy/phase helpers.
The simulator checks topological fanin ordering and computes graph values rather
than trusting cached node truth tables.

Validation: `go -C abcgo test ./bool/bdc ./misc/mvc ./misc/avl ./bool/dec` passes.
