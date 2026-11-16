# GIA isomorphism review (in progress)

Scope: giaIso.c, giaIso2.c and giaIso3.c; Go implementations are split among
iso_refine/canonical/reduce/diagnostics, iso2_refine/compare and giaIso3.go.
Review these sequentially within the open GIA package audit. The original
131-package objective remains active, with2/131 packages complete.

## Plan

- [x] Locate original routines, actual Go implementations and existing tests.
- [ ] Reconcile first-generation level classes, packed fields, unsigned
  signatures, sorting/ties, forward/backward refinement and convergence.
- [ ] Review canonical traversal, PI/register permutations, serialized keys,
  output reduction, dual-output and estimate/ownership boundaries.
- [ ] Reconcile second-generation hash/unique labels, cone traversals, sequential
  comparison and quality/skip reduction policies with the original source.
- [ ] Review synchronous/reverse signatures, source RNG and diagnostic drivers.
- [ ] Add red Go-only regressions before uncovered bug fixes, then independent
  scalar/classification/permutation/sequence tests and compactness checks.
- [ ] Reconcile all source names, freeze edits, run affected race/vet and full
  module Go-only integration before claiming family closure.

Initial findings to investigate: first-generation level allocation only uses
output depth despite including all nodes, making deeper dangling cones unsafe.
The source explicitly resets terminal levels before computing node levels.
Second-generation uniquification currently uses a new map instead of its
allocated reusable hash workspace; examine ordering, mark/scratch ownership and
compactness before deciding whether this is a faithful representation change.
Several verbose reports/diagnostic drivers omit source behavior. Existing simple
tests do not prove this whole family correct; no new C fixtures will be used.


## First level-preparation repair

Inventory totals:33 routines in giaIso.c (including four packed-word accessors),
23 in giaIso2.c and12 in giaIso3.c. Closure requires all68, not just their
currently green simple wrappers. Main refinement/canonicalization and second-/
third-generation runtime reconciliation remain open.

A new depth-five dangling-chain regression fails at the output-sized level
array (/tmp/abcgo-iso-dangling-levels-red.log). C's Gia_IsoPrepare similarly
allocates from output depth while classifying every object, so this is an
explicit repair of inherited out-of-bounds behavior: include all AND levels in
the maximum. Retain all nodes, bucket order and source refinement algorithm;
do not silently prune unused cones. Separately, a new terminal-reset regression
fails with stale negative levels (/tmp/abcgo-iso-terminal-levels-red.log).
Restore C's explicit constant/CI level initialization before computing ANDs.

Twenty-four chain-depth/output/sequential combinations cover depths1/2/5/32,
constant/input/shallow-AND outputs and RI/RO boundaries. Independent stable
buckets check every level, class record and packed object ID; low signature
words and graph node Value fields must stay unchanged. The terminal-reset case
checks the source initialization contract separately. All fixtures are Go-only.

Frozen full Go-only module exits0 (130 tested packages,3 commands without
tests), focused existing/new isomorphism race1.025s and GIA vet exit0. Logs:
/tmp/abcgo-iso-prepare-final-{module.jsonl,race.log,vet.log}. All sessions ended.
Only this preparation checkpoint is closed; the family and full GIA package
remain in progress, original package completion2/131.


## Second-generation uniqueness state and compact owned workspace

New regressions fail before repair (/tmp/abcgo-iso2-owned-storage-red.log):
Go ignores an initially marked candidate (returns2 singletons instead of C's1),
and warmed64-candidate classification allocates3 times for all ties or18 times
for all singletons. Source's normal verification assumption is clean candidate
marks; its actual enabled code also consumes initially marked candidates into
the tied list and clears their mark0. Preserve that defined state transition,
including marks on noncandidates remaining untouched.

Follow the original ordered operations: remember each signature's first object,
mark duplicate objects, compact ties in place, clear consumed marks, sort unique
objects and assign incremented labels/signature salts. An owned Go map replaces
the C probing table/touched-slot vector; map iteration is never used, so hash
bucket placement cannot alter output order. Clear the map after the lookup pass
and reuse a typed singleton slice. Existing safe signed32 comparator semantics
remain; singleton signatures are distinct, so sort stability is immaterial.
C's subtracting comparator can overflow. Wide-signature tests pin the existing
portable Go ordering; defined-range C comparisons keep the same order.

Remove unused comparison vectors/maps and never-read count/timing fields from
this opaque Go manager. C's comparison entry points allocate their actual local
buffers separately; no runtime comparison state or source search is removed.
The first-occurrence map grows with encountered signatures, while the singleton
buffer starts bounded by candidate count. A separate review confirms all256
prime constants match both original source tables; tests do not read C files.

Sixty-four independent scalar cases cover candidate counts0/1/2/3/8/17/64/257,
zero/dense/unique/wide signatures and initially marked candidates. Quadratic
classification and insertion ordering check exact kept IDs, labels, salted
values, mark0 consumption and untouched mark1. The257-input case covers salt
index wraparound. Warm64-candidate tests now allocate0 for both extreme classes.
Existing output reduction/comparison tests remain active.

The unchanged local100ms/count3 construction benchmark goes from104520..104525
B/op and19 allocations to1232 B/op and7 allocations. Sample medians37507ns
before and1044ns after describe this64-candidate constructor only, not an
end-to-end isomorphism speedup. Logs:
/tmp/abcgo-iso2-manager-before-bench.log,
/tmp/abcgo-iso2-owned-reviewed.log (final benchmark plus existing/new tests).
An intermediate1296 B result is superseded after sizing singleton capacity
from candidates instead of total objects.

Frozen full Go-only module exits0:130 tested packages pass,3 commands have no
tests. JSON confirms64 scalar leaf cases execute; focused isomorphism race1.042s
and GIA vet exit0. Logs /tmp/abcgo-iso2-owned-final-{module.jsonl,race.log,vet.log}.
All handles ended. This closes uniqueness workspace/mark reconciliation only;
remaining refinement, canonical ordering, sequential comparison, drivers and
reports in the68-routine family still require review. Original audit2/131.


Next: check forward/backward and register signature arithmetic, then utility
sort tie ordering and canonical PI/RO/PO traversal before making broader
canonicalization/reduction claims. Keep source diagnostic/report omissions in
the inventory for their own new regressions. All expectations must remain Go-only.


## Restore source sorting payload order before canonical labeling

A new hand-traced11-CI regression fails at the second packed class item:8
instead of source2 (/tmp/abcgo-iso-canonical-ties-red.log). Both implementations
sort the low-word signature keys, but the old middle-pivot utility sorter
reorders equal-key high-word IDs differently. Tie breaking assigns unique labels
in that order, which feeds canonical PI permutations. The regression checks
source class boundaries/items and, after repair, the entire tie-label transfer
and canonical CI/permutation output. Do not normalize this metadata difference
away merely because both outputs remain Boolean-equivalent.

Restore the packed-sort dependency's original last-pivot operations and its
standalone three-way routines. Source QuickSort3's wrapper calls the two-way
implementation, an important naming quirk retained by Go. Eight utility payload
goldens fail before repair (/tmp/abcgo-iso-sort-payload-red.log);576 independent
key/bijection/range cases cover both directions and the10/11-word cutoff. The
source helper dependency review is in misc/util/REVIEW_packed_sort_dependency.md.
It does not mark that full package audited.

Frozen full Go-only module exits0 (130 tested packages,3 commands without tests),
full utility race1.427s, focused isomorphism race1.046s and util/GIA vet exit0.
Logs /tmp/abcgo-iso-sort-final-{module.jsonl,util-race.log,gia-race.log,vet.log}.
All handles ended. Signature propagation, full canonical traversal/reduction,
second-generation comparison and driver/report reconciliation still remain;
this is a dependency correctness checkpoint, not68-routine family closure.
