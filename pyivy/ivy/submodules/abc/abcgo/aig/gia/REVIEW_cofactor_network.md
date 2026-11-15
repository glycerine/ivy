# Cofactor connectivity, scoring and reconstruction

Review all 49 routines in `src/aig/gia/giaCof.c`, implemented by
`cof_network.go`, `cof_analysis.go` and `cof_duplicate.go`. Preserve the source
explicit connectivity, traversal tags, MFFC reference sessions, stable fanout
ranking, levelized structural estimate and two-phase cofactor reconstruction.
This family review is part of the ongoing sequential GIA package audit.

## Repairs and regression evidence

| Finding | Repair and red evidence |
| --- | --- |
| Ordinary structural MUX controls were not counted. | Use the shared mux-reference builder; `cof_references_review_test.go` fails before repair (`/tmp/abcgo-cof-refs-red.log`). |
| The fanout accessor ignored temporary dereference/reference sessions. | Read the existing mutable `refs` counter. Physical adjacency remains intact. Direct sessions and full MFFC restoration have separate checks. |
| The shared mux-reference builder classified real XOR roots or branches as structural MUXes. | Reuse the existing actual-Boolean-AND/structural-MUX predicates, generalized from enable detection. The typed regression fails first (`/tmp/abcgo-cof-mux-typed-red.log`); real MUX controls still count. Legacy raw recognition APIs retain their contract. |
| The Go forward sweep changed C's simplification processing order and scores. | Restore C's levelized LIFO worklist. The native raw scratch/count regression fails first (`/tmp/abcgo-cof-worklist-red.log`); all 6,834 records then agree exactly. |
| Reconstruction discarded output roles, resets and comments. | At the unchanged interface boundary, use `giaCopyInterfaceMetadata`. Six typed/reordered raw, clean and repeated variants fail first (`/tmp/abcgo-cof-rebuild-red.log`); repaired tests verify values and independent ownership. |
| Histogram bins, blank columns and source report fields were changed or absent. | Native histogram and driver tests fail first (`/tmp/abcgo-cof-fanio-red.log`, `/tmp/abcgo-cof-driver-red.log`). Restore decimal ranges, formatting, memory and elapsed fields. |

The preceding investigation considered normalizing scratch aliases and changing
C's hash-query boundary. That experiment was discarded: agreement after such a
change does not establish correctness. `CountRemoved` is a discovery-time
structural estimate, not a complete Boolean replacement map. Its original
processing order and raw literals are now compared directly to unmodified C.
The actual rebuilding routines are separately checked for Boolean equivalence.

## Source reconciliation and coverage

`test_vectors/cof_review.c` includes the actual original `giaCof.c`; its default,
rebuild, histogram and driver modes do not modify or stub the algorithm. Native
checks use Linux/GCC and the original ABC objects through the existing source
oracle helper, and skip when those facilities are unavailable.

| Source routines | Evidence |
| --- | --- |
| Handle/classification/level/count/link accessors; constructor, stop and value/tag helpers | 32 ordinary hashed graphs; complete object-ID-normalized fanin/complement/fanout records, mux references, CI/node counts and cleanup. Direct overflow reset coverage verifies stale tags cannot contaminate later queries. |
| Recursive/public TFO, TFI and support | Native singleton and every distinct candidate pair on all 32 graphs; existing direct output/support tests. |
| Dereference, reference and MFFC | Native per-candidate MFFC values, explicit mutable-count sessions, and every reference counter restored after each query across a reconvergent graph. |
| Bounded insertion and high-fanout collection | Exact native ranked lists at every positive limit, plus independent stable-sort expectations at every prefix of 256 candidates for limits 0/1/2/7/64/257. |
| Removal estimate and high-fanout reports | Both constants at every CI/internal candidate; exact discovered scratch records and counts. The composed native driver checks top-three reports and surrounding diagnostics. |
| Fanin/fanout/MFFC histogram and driver | Nineteen native reports spanning decade boundaries, 10,000 fanouts and MFFC ranges. Separate empty/CI-only tests retain finite Go averages. |
| Raw and cleaned cofactor duplication | Every CI/internal pivot on 64 generated graphs, plus typed XOR/MUX/buffer phase cases and reordered ports; exhaustive scalar output evaluation. |
| Variable selection, transfer and repeated cofactoring | Five native thresholds, including none/all candidates and an empty selected set. Direct tests cover complemented aliases, duplicate targets, constant/unreachable sentinels and preserved source signal order. Invalid pivots and the 200-signal bound have explicit tests. |

The rebuild matrix compares 2,330 graphs against unmodified C for object counts,
CI/CO/register dimensions and all 16 assignments. An independent scalar evaluator
checks those 2,330 results against the original graph. Another 704 typed/reordered
raw, cleaned and repeated rebuilds receive exhaustive output checks. Typed
metadata/ownership checks run independently of the native ordinary-AIG oracle.

## Compactness

Replace full-vector copying, stable sorting and rebuilding at every insertion
with one bounded in-place shift. Preserve descending rank and insertion order
for ties. For 4,096 candidates and a top-100 list, three-pass benchmark evidence
changes from 9,649,997 ns/op, 7,637,333 B/op and 16,416 allocs/op to 36,982 ns/op,
zero bytes and zero allocations. Timing is observational; exact order and
contents are asserted independently.

Remove three unused phase/mark words, the redundant object/level counts and the
CI/CO vectors that were filled but never read (also unused in original C).
Object count comes from the object slice and level count from its queue slice.
Keep `nNodes`, which avoids repeatedly scanning classifications. Keep the mutable
reference counter, which has different semantics from adjacency length, and
`iNext`, which links C's level queues. Direct Go links replace C's packed arena;
there is no reason to recreate pointer arithmetic or store duplicate lengths.

## Source boundaries and explicit Go behavior

The source structural estimate assumes an ordinary AND graph and valid nonconstant
CI/internal candidates. Typed primitives remain atomic for that estimate rather
than being incorrectly evaluated as ANDs. Actual cofactor reconstruction handles
real XOR/MUX/buffer logic using the existing Boolean hashing operations. This is
a representation extension of the source two-phase construction.

Opaque source arena handles become logical Go object IDs. `Cof_ObjSize` is one
logical record, not a C word count; raw pointer offsets and arena byte sizes are
not comparable across representations. Memory reporting uses allocated Go object,
link and complement payload (without runtime allocator overhead), and elapsed
time uses wall time rather than the C CPU clock. Only these two numeric driver
fields are normalized in the native comparison.

Preserve finite zero averages where source divides by zero; safe invalid-pivot
results where source reports errors; bounded transfer checks where C assumes
valid mappings; and traversal-ID rollover cleanup rather than signed C overflow.
The 200-signal cap and reverse-topological repeated cofactoring remain intact.
Duplicate/CO/constant-root cone queries outside the source algorithm's candidate
domain are not redefined by this review. Metadata copying intentionally retains
unchanged Go interface semantics that the old C routine does not copy.

## Validation

Final focused checks pass in 9.745s (`/tmp/abcgo-cof-final-focused.log`). The
frozen uncached module race run passes all 130 packages with tests, including
GIA in 481.225s; three command packages have no tests
(`/tmp/abcgo-cof-final-module-race.log`). GIA vet passes with empty output
(`/tmp/abcgo-cof-final-vet.log`). Both exact process handles terminated
successfully before these results were recorded. All 49 source routines in this
family are reconciled; see `../../SEQUENTIAL_AUDIT.md`.
The overall GIA package audit remains incomplete until its remaining families
and final source inventory are reconciled.
