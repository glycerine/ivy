# FXU design and compactness review

Completed separate readability/representation review of the complete package.
The port implements single-cube, double-cube and complement-aware extraction
from `src/opt/fxu`, not a replacement generic factoring engine. It keeps source
function names at the compatibility boundary and uses Go slices/maps inside.

Sparse row/column incidence lists share literal objects. Cubes have stable
identities for pair matching; removal is explicit rather than conflating a
zero-literal constant-one cube with a deleted row. Node counts, cube counts,
literal counts and heap capacity derive from slices. Heap indices are genuine
mutable membership indices needed by update/delete, not redundant length state.
`OldNodes` is an immutable boundary needed to interpret the original/new
co-singleton polarity convention; divisor weights/catalogs are algorithm caches.

Intrusive linked lists, allocation pools, hash bucket chains and capacity fields
were replaced with slice ownership, Go allocation and complete-key maps. The
single/double heaps share one generic implementation, preserving strict comparison
and tie behavior. Full candidate rebuilding after each rewrite replaces the
source's distributed incremental link maintenance. This trades incremental
performance and later equal-weight tie order for smaller auditable code while
retaining the actual divisor definitions, occurrence weights, extraction and
complement lookahead algorithms. It is not an exact optimization-trace guarantee
for arbitrary tied multistep inputs. No claim of original large-network throughput
is made; benchmarking is outstanding.

The neutral `FXUNetwork` boundary uses stable object-ID-indexed arrays. Appended
rows are extracted nodes and can have larger IDs than their users. The included
Network validates dependencies and evaluates by DFS. Concrete ABC integration
belongs in its owner package, avoiding an import cycle or unfinished graph import.
SOP normalization is explicit: callers must satisfy the original SCC-free contract.
Source settings LiteralCountMax/CanonicalDivisors are retained and documented as
unused by the original active algorithm, rather than inventing new restrictions.
