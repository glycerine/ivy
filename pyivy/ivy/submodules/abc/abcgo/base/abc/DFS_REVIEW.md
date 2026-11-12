# DFS, support, level and isomorphism API group

Scope: original abcDfs.c routines in abcDfs.go and dfs_iso.go, with dfs_test.go.
Network adds integer Temp scratch and an optional CILevel timing callback.

## Readability and compactness pass

Vectors are typed slices. Driver/fanout helpers state the netlist boundary
explicitly, and iterative DFS uses typed {Node,Next} frames instead of packing
integers into pointer slots. Recursive and iterative variants remain separate
actual algorithms. Level-head arrays retain Copy-linked lists for existing
clients. Choice level integers use Temp instead of integer-to-pointer casts.

Functional-isomorphism helpers use the importable AIG Graph as the two-output
manager, avoiding the unfinished GIA command package. Ordered CI renaming and
common-support alignment follow the original. Logic covers convert directly to
AIG; arbitrary Go Boolean functions use exact Shannon expansion. Equivalence is
proved with native Satoko on the output inequality, not assumed from structure.
The returned helper manager is *aig.Graph with explicit Edge values; this is a
Go API representation change from the original Gia_Man_t/literal integers.

## Separate fidelity and bug pass

Reviewed every original traversal group: forward/reverse, supplied roots,
contained TFO, sequential feedback, iterative stacks, hierarchy/all objects,
net ordering, boxes, support, AIG choices/map barriers, levels and cycles.
Netlist reverse DFS retains the source's output-net result convention. Barrier
latch timing transfers and transparent buffer zero delay are retained. CILevel
is the dependency-neutral hook for timing-derived initial levels.

Regressions cover recursive/iterative order and dangling collection, netlists,
latch feedback, support ordering, arrival levels, reverse levels, choice cycles,
traversal-ID rollover, SOP/function equivalence and common-input renaming.
A red contained-TFO test exposed a zero-level barrier being prepended behind
an active traversal cursor; per-level queues now propagate that cone fully.
Another review regression clears stale Temp values before choice-level sorting.

Intentional source bug corrections: no nil prefix in levelized order; collectCIs
actually includes CIs; hierarchy collect-all uses the hierarchy traversal for
remaining objects; whole-network CI support excludes the global constant;
rollover never mistakes cleared traversal marks for completed visits. Choice
chains terminate on cycles instead of looping forever. Single-node support
continues to include a reached AIG constant, matching its original API.

Selected full previously ported ABC core plus DFS tests pass race tests and vet.
Full base/abc remains blocked by other untranslated source files, starting at
abcFanOrder.go. This review does not claim package completion or C differential
coverage for these traversal APIs.
