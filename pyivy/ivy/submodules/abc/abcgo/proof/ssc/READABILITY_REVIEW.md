# Readability and compactness review

Completed separately from the bug review. The source repeats graph rebuilding,
SAT scratch maps, and phase-normalized comparison logic across several passes.
The native port uses one graph builder/copier, one packed simulator, one class
partition representation, and one incremental encoder. Their state is explicit
and independently owned. No unsafe arithmetic, pointer sentinels, global random
state, string-generated code, or function-closure adaptation is used in outputs.
Closures in the library implement bounded local visitors and callbacks.

Reviewed the final control flow against source family boundaries. Entry points
keep their original names; native parameter changes and consolidated low-level
helpers are documented in PORTING.md. Counters are grouped in Statistics rather
than repeated storage fields. SAT/UNSAT/UNKNOWN handling remains explicit.

## Follow-up: class ownership and containment

Reviewed the manager metadata publication and containment solver separately
from the source correctness pass. The class partition now has one publication
helper that updates representative and next links together while preserving
proved flags; the former second publication at sweep exit was removed.
Containment now derives the source CNF and queries one solver directly, so it
no longer creates an unrelated sweeping manager or changes input classes.
The existing sequential containment extension uses a combinational copy with
state inputs left free; the C routine's CNF output-count precondition does not
generally cover that interface.

Class refinement currently reindexes and republishes the entire partition for
each split. This is simple and follows the existing Go representation, but can
cost O(objects) per split. Incremental publication would need separate tests
for constant candidates, class heads and proved flags before changing it.
