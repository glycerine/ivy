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
- [x] Review synchronous/reverse signatures, source RNG and giaIso3 drivers.
- [ ] Review first-generation permutation/CEX diagnostic drivers.
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


## Check signature arithmetic with an independent word model

Add486 Go-only scenarios across18 edge-phase patterns,0/1/2 registers,
three initial-value profiles and three label profiles. An abstract graph model
uses wider accumulators and explicit32-bit masks to check first-generation
forward/backward propagation, iteration-dependent salts and RI/RO feedback.
The same model checks second-generation initialization, weighted fanins,
symmetry/self-feedback biases, and third-generation synchronous snapshots.
Snapshot ownership, uniqueness, unchanged structural fields and20 direct
update-value index/product cases are checked separately. No new arithmetic
bug was found; production code is unchanged in this checkpoint. Tests neither
read C files nor invoke a C compiler.

Frozen full Go-only module:130 tested packages pass,3 commands have no tests;
all486 scenario leaves execute. Focused isomorphism race passes1.146s and GIA
vet is silent. Logs /tmp/abcgo-iso-signatures-final-{module.jsonl,race.log,vet.log};
all process handles are now absent after completion. Canonical traversal,
reduction/comparison and driver/report reconciliation remain open. Audit2/131.


## Canonical traversal against an independent graph specification

Add648 Go-only cases over18 phase patterns,0/1/2 registers, one/two primary
outputs, three small unique-label orders, and requested/omitted PI permutations.
An abstract DAG oracle uses insertion sorting and independent postorder traversal.
It checks exact CI/AND/CO order, PI permutation, RO-to-RI pairing, constant and
dangling cones, superseded traversal marks, and unchanged object fields. Only
multi-PO values may change; the single-PO value must remain untouched, as in C.
These labels stay in C's defined signed-comparison range; this test does not
assert a particular qsort ordering for equal output keys or emulate overflow.
No traversal discrepancy was found, and production remains unchanged.

Frozen full Go-only module passes130 tested packages plus3 commands without
tests; all648 leaf cases execute. Focused isomorphism race1.407s and GIA vet pass.
Logs /tmp/abcgo-iso-canonical-final-{module.jsonl,race.log,vet.log}; all handles
ended. This covers the ordering contract after labels exist, not full refinement
or canonical reduction. Next restore ReportTopmost traversal side effects with
a new failing regression, then continue the remaining family. Audit2/131.


## Restore topmost reporting traversal ownership

A new regression fails on36 graph/class cases before repair. Go's temporary
boolean array reproduces printed topmost representatives but omits C's traversal
epoch increment and current marks. Restore Gia_ManIncrementTravId and source
fanin marking for class ANDs and the RI driver of class ROs; use those marks
when selecting the first unmarked representative in each class. This also
removes the unnecessary scratch array. It does not change printed output.

The Go-only oracle builds class membership from the abstract graph and checks
exact reports, current marks after a premarked old epoch, untouched node values
and structural/ordinary mark fields, and unchanged packed/class storage. Cases
cover0/1/2 registers, three phase patterns and empty/all/AND-only/RO-only classes.
The warmed empty-report check allocates0 before and after on this Go toolchain;
do not claim an allocation reduction from that fixture. The actual new red
failures concern traversal state, not output or allocation.

Red /tmp/abcgo-iso-topmost-red.log contains36 failing leaves; focused repaired
run /tmp/abcgo-iso-topmost-green.log exits0. Frozen full Go-only module passes
130 tested packages plus3 commands without tests, including36 report leaves.
Focused isomorphism race1.445s and GIA vet pass; all handles ended. Final logs:
/tmp/abcgo-iso-topmost-final-{module.jsonl,race.log,vet.log}. Remaining driver
reports (including missing source graph stats), full refinement/reduction and
second-generation comparison still require reconciliation. Audit2/131.


## Consume second-generation class buffers as in C

New ownership regressions fail in10 of16 matrix cases before repair. Existing
Go produces correct class membership but copies accepted buffers and retains
input worklists instead of consuming them. Restore the source contracts:
full-quality checking uses its input vector as a growing worklist, compacts
each class, transfers its buffer to the result and appends rejected members
for later splitting. Skip mode transfers only accepted buffers; rejected
classes remain in the input and their result singletons are copies. Reuse one
reject vector instead of allocating/copying a second queue and per-class vectors.
The existing comparison test now creates a fresh input for its independent
skip-mode run, since the full-quality call consumes its previous input.

Sixteen Go-only scenarios cover four partitions, both policies and both output
phases. Independently specified cone kinds predict exact breadth-first splitting
order, including empty and singleton classes. Exact initial vector capacity
forces worklist relocation on appended splits; pointer checks prove accepted
buffers move without copying, while consumed capacities become zero. Two new
progress regressions also fail before restoring source reports every50 classes.
Red logs /tmp/abcgo-iso2-classes-{red,progress-red}.log; focused repaired log
/tmp/abcgo-iso2-classes-green.log. No source file is a runtime test vector.

Frozen full Go-only module passes130 tested packages plus3 commands without
tests; all16 ownership leaves execute. Focused isomorphism race1.397s and GIA
vet pass; all handles ended. Final logs /tmp/abcgo-iso2-classes-final-
{module.jsonl,race.log,vet.log}. Source comparison/order/cone arithmetic and
end-to-end reduction still need independent reconciliation; this checkpoint
closes class-worklist ownership, not the family. Audit2/131.


## Sequential cone orders and bounded updates

Reconcile both order walkers/wrappers and the bounded-cone walker/wrapper/update
against source: unsigned child-value ordering, physical fanin order on ties,
first-visit deduplication, postorder map positions, appended RI roots, constant
terminals and self-feedback. A separate explicit entry/exit-frame model checks
216 order scenarios over registers, complemented edges, wide/tied values and
empty/repeated/reordered output selections. A trailing invalid PO beyond count
checks prefix bounds. Unvisited map entries remain untouched; old traversal
marks are superseded. Combinational wrappers are checked on register-free graphs.

An iterative bounded first-visit model checks120 cone/update scenarios: depths
0/1/2/3/8,0/1/2 registers, four phase patterns and both candidate orders. Every
non-CO root is checked separately. RI bridges retain depth while AND edges
consume one unit; earlier visits claim shared nodes. Sequential in-place updates
use already updated signatures, including32-bit wraparound. Candidate order,
traversal counts, final marks and unrelated object fields are pinned.

The first model run exposed a fixture assumption, not a production bug: its
abstract AND edges retained call order while C/Go append sorts literals. Correct
the specification to increasing literal order and assert raw encoded edges
match it. Earlier commutative signature checks did not detect that assumption.
No production change was needed. All new fixtures remain Go-only.

Frozen full Go-only module passes130 tested packages plus3 commands without
tests, including all216+120 leaves. Focused isomorphism race1.484s and GIA vet
pass; all handles ended. Logs /tmp/abcgo-iso2-walk-final-
{module.jsonl,race.log,vet.log}; corrected initial run
/tmp/abcgo-iso2-walk-reviewed.log. This reconciles seven source walk/update
routines; exact pair checks, full refinement/reduction and drivers remain open.
Original audit2/131.


## Reverse signatures and source report side effects

Two new report regressions fail before repair. Gia_IsoCompareVecs prints
invented abbreviated graph counts instead of the two source Gia_ManPrintStats
calls. Besides missing names/statistics, that omission leaves reverse levels
in place: the standard statistics routine recomputes forward levels when
existing level storage is present. Restore those two calls. The direct test
pins this level side effect and names/standard headers, then exact signed-value
multiset counts with duplicates and unequal level counts. Collected data keeps
requested order and owns its copy; object lists remain unchanged. A separate
two-graph driver regression reproduces the missing statistics through the wrapper.
Red logs /tmp/abcgo-iso-reverse-{red,two-red}.log; repaired focused logs
/tmp/abcgo-iso-reverse-{green,reviewed}.log.

Twenty-four independent reverse-signature scenarios cover0/1/2 registers, four
phase patterns and zero/near-wrap initial values. Abstract reverse levels and
wide scalar multiply-with-carry arithmetic check every bucket/value, two
accumulating invocations, unchanged ordinary marks/edges, untouched traversal
epochs, reset-sample consumption and the next global RNG sample. The reverse
report driver pins every level/object line. Twelve synchronous-driver scenarios
check exact iteration counts and final values against independent count-stability
iteration; source intentionally stops when distinct counts repeat, not when
signature vectors are identical. Four uniqueness boundary vectors include empty,
all-equal and high-bit values, preserving input order. All fixtures are Go-only.

All12 giaIso3.c entry points now have source behavior reconciliation and
independent arithmetic or driver/ownership checks. Before closing that subfamily,
finish the compactness assessment of its standalone unique-count helper
(source copy/sort versus the current Go map) and its driver dependency.
This does not close the remaining giaIso.c/giaIso2.c routines or full GIA audit.

Frozen full Go-only module passes130 tested packages plus3 commands without
tests, including24 reverse and12 synchronous-driver leaves. Focused isomorphism
race1.490s and GIA vet pass; all handles ended. Final logs:
/tmp/abcgo-iso-reverse-final-{module.jsonl,race.log,vet.log}. Original audit2/131.


## Close giaIso3 standalone count/driver compactness review

Restore the source distinction: Gia_Iso3Unique copies/sorts its snapshot while
Gia_Iso3Test uses hash counting. The standalone helper now clones, sorts and
compacts an int32 slice; its input remains unchanged. Keep the existing typed
hash counter in a private driver helper, specializing C's one-word hash count
and avoiding an unused representative mapping. The driver keeps its previous
complexity and implementation. This is a source/compactness reconciliation,
not a new functional bug; existing boundary/driver regressions stay active.

Twenty-one Go-only cardinality cases check both paths at sizes0/1/2/255/256/257/
8192 with equal,127-way tied and bijective wide values. Three100ms samples of
the unchanged standalone benchmark at8192 entries show152192 B/33 allocations
before and32768 B/1 allocation after. Sorting trades speed for space on distinct
values: sample medians about238us before versus684us after (ties182us versus
264us; all-equal177us versus25us). Do not describe this as a general speedup;
the iterative driver retains hash counting. Logs:
/tmp/abcgo-iso-unique-before-bench.log and /tmp/abcgo-iso-unique-reviewed.log.

Frozen full Go-only module passes130 tested packages plus3 commands without
tests; all21 cardinality leaves execute. Focused isomorphism race1.502s and GIA
vet pass; all handles ended. Final logs /tmp/abcgo-iso-unique-final-
{module.jsonl,race.log,vet.log}. Together with signature-model and reverse/driver
checkpoints, this closes review of all12 giaIso3.c routines. The other56 source
routines in this family, and the larger GIA audit, retain their documented
open checks; original package completion remains2/131.
