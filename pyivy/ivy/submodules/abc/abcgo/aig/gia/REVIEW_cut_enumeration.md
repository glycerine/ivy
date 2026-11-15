# Stand-alone priority cuts (giaCut.c)

`giaCut.go` ports sorted cut merging, containment/dominance filtering, source tie
and priority limits, cut signatures, tree-leaf scoring, canonical truth storage,
support minimization and incremental per-node enumeration. This is separate from
the cut-replacement abstraction in dup_replace_cut.go and from mapper-specific
cut engines. Buffers and MUXes retain the source rejection boundary; real XORs
are supported.

Actual-C comparisons cover 1,512 configurations: ordinary AND graphs, real XOR
graphs and larger-ID/capacity-growth graphs, cut sizes 2..8, limits 2/10/50 and
no-truth/unminimized/minimized modes. Graphs, serialized cut order, counters,
overflow counts, canonical function IDs and every truth word match. Independently
validate sorted, unique, non-dominating cuts and 383,656 stored truth functions
over all actual eight-input completions. Unminimized cuts additionally agree on
every formal cut assignment, without using the cut merger or truth-expansion
routines as the reference. The first typed fixture omitted XOR fanins from its
graph dump; correcting that fixture resolves the mismatch without a production
algorithm change.

A separate native primitive matrix checks 73,728 calls to both merge entry
points, 4,096 subset relations and 192 priority-set updates. Cases include
colliding signatures across IDs 0/64, 1/65 and 63/127, zero/full merge bounds,
equal-cost insertion, dominated-cut removal, truncation and candidate pointer
reuse. Independent bit-set unions/subsets verify the results. Source's wide cost
metadata remains representable; no manual packing or narrowed counter is added.

No core enumeration defect was found in these comparisons. Red report tests
expose lost leaf padding/braces, source fields, percentages and timing format;
restore cut and public-driver reports. Empty-interface statistics stay finite
instead of introducing C's undefined zero divisions. Additional tests cover
constructor bounds, buffer/MUX rejection, private references, borrowed graph
marks/Values, preservation of input cuts and stored truth under composition,
every input phase and AND/XOR operation, and eight-leaf formal assignments.

Compactness restores the single-word path and bounded four-word scratch arrays
instead of allocating two input copies and a result per composition. Warm calls
drop from 40B/five allocations to 24B/three at size four, and 160B/five allocations
to 96B/three at size eight. Remaining copies belong to shared truth-variable swaps;
this review does not duplicate that utility implementation or add a cache. Sample
timings are approximately unchanged. Logical limits, private reference counts,
serialized cut lengths and fixed candidate pools serve separate source roles.

The combined cut-focused suite passes in 10.806s, including the earlier cut-
replacement regressions. Full uncached module race tests pass all 130 tested packages
(three command packages have no tests), and GIA vet passes. Logs are
/tmp/abcgo-cut-enumeration-final-module-race.log and
/tmp/abcgo-cut-enumeration-final-vet.log. SEQUENTIAL_AUDIT.md records checkpoints. This note covers stand-alone enumeration; the remaining GIA
families and the 131-package sequential audit stay open.
