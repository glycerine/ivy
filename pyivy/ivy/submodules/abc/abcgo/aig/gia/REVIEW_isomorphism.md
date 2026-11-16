# GIA isomorphism review (in progress)

Scope: giaIso.c, giaIso2.c and giaIso3.c; Go implementations are split among
iso_refine/canonical/reduce/diagnostics, iso2_refine/compare and giaIso3.go.
Review these sequentially within the open GIA package audit. The original
131-package objective remains active, with2/131 packages complete.

## Plan

- [x] Locate original routines, actual Go implementations and existing tests.
- [x] Reconcile first-generation level classes, packed fields, unsigned
  signature arithmetic, sorting and tie-accounting kernels.
- [x] Reconcile full forward/backward refinement and convergence.
- [ ] Review canonical traversal, PI/register permutations, serialized keys,
  output reduction, dual-output and estimate/ownership boundaries.
- [ ] Reconcile second-generation hash/unique labels, cone traversals, sequential
  comparison and quality/skip reduction policies with the original source.
- [x] Review synchronous/reverse signatures, source RNG and giaIso3 drivers.
- [x] Review first-generation permutation/CEX diagnostic drivers.
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


## First-generation partition/tie accounting and completion reports

Forty-two Go-only partition cases independently group keys across empty and
multi-class workspaces,2/3/10/11/13-member classes, equal/distinct/dense/wide
values and deterministic mixed values. Check unsigned key ordering, singleton
labels in class/key order, class boundaries, entry/singleton counters, packed-ID
bijection and untouched signatures/ordinary marks. Uniform classes retain their
payload order, and sentinel words outside classes remain unchanged. Earlier
source tie-payload goldens still pin order within nonuniform tied partitions.

Five tie-policy cases check active highest-level assignment and both experimental
variants, preserving counter accumulation and class removal. The experimental
variants intentionally label only one member of larger classes before dropping
the class, as C does; do not advertise them as complete tie strategies. Variant2
retains Go's accurate selected-class diagnostic rather than C's post-compaction
loop index (which prints the last old index instead). No algorithm/state change
was needed in these kernels. The class printer's source format is also pinned.

Two new verbose regressions fail before restoring the missing full-refinement
completion snapshot and six timing rows. A hand-traced1-CI/1-PO graph expects
iterations0,2 for full refinement versus0,1,2 for forward-only mode, then Start/
Simulate/Refine/Sort/Other/TOTAL percentages. Both quiet-mode paths remain silent.
Use the existing time-percentage printer. Compute total/other durations locally
and remove their redundant manager fields plus the never-used vLevCounts pointer.
Red /tmp/abcgo-iso-classes-red.log; repaired focused log
/tmp/abcgo-iso-classes-green.log.

Frozen full Go-only module passes130 tested packages plus3 commands without
tests; all42 partition,5 tie and2 verbose leaves execute. Focused isomorphism
race1.485s and GIA vet pass; all handles ended. Final logs:
/tmp/abcgo-iso-classes-final-{module.jsonl,race.log,vet.log}. These checks close
the partition/tie kernels and completion reports, not the full convergence,
canonicalization/reduction or first-generation driver review. Original audit2/131.


## Combinational reduction pipelines and restored reports

Add72 Go-only end-to-end cases over first-generation reduction and both
second-generation quality policies, output phases, estimate/actual modes,
requested/omitted metadata and quiet/verbose/very-verbose execution. Known
circuit families specify exact groups/representatives; all64 input assignments
check reduced truth. Pin identity fields, source object counts, owned results,
metadata lifetime after manager shutdown and permutation support/bijection.
First-generation estimate initializes metadata while second-generation estimate
leaves output arguments untouched. The second variant keeps ignoring dual mode,
as C does. These differences are explicit contracts, not unified by the tests.

Two canonical-key cases check all12 cones, constant encodings, exact graph-kind
key equality and cone-local PI permutations. Normalize truth through independently
known supports to verify the returned permutations actually align matching
functions. Six additional pipeline cases distinguish a reconvergent three-AND
graph from isomorphic two-AND chains even though all compute the same three-input
conjunction. This guards against replacing structural reduction with Boolean
minimization. Initial pipeline checks found no grouping/permutation discrepancy.

Five new report regressions fail before restoring omitted source reports:
indexed nontrivial-class listings/headings in both generations, per-candidate
cone statistics in first-generation verbose mode, and100-output progress through
both singleton and non-singleton paths. Reuse existing vector printers/statistics
services and restore source candidate-label spacing. The72-case matrix confirms
restored reporting does not change reduction/metadata behavior. Red log
/tmp/abcgo-iso-reduction-reports-red.log; expanded/repaired focused logs
/tmp/abcgo-iso-pipeline-{reviewed,green,expanded}.log.

Frozen full Go-only module passes130 tested packages plus3 commands without
tests; all72 matrix leaves execute. Focused isomorphism race2.138s and GIA vet
pass; all handles ended. Final logs /tmp/abcgo-iso-pipeline-final-
{module.jsonl,race.log,vet.log}. Sequential/dual reduction, exact pair terminal
checks, full convergence/source-driver reconciliation and remaining compactness
checks stay open. Also inspect the newly reused Vec_VecPrintInt dependency's
empty-level source discrepancy with a new red before repair. Original audit2/131.


The vector-printer dependency follow-up is resolved separately: Vec_VecPrintInt
now omits empty levels as C does; Vec_WecPrint still emits them. Four Go-only
goldens and full module/vec race/GIA isomorphism race/vet pass. Details and logs
in misc/vec/REVIEW.md. Neither this dependency checkpoint nor the combinational
matrix closes the outstanding GIA family scope.


## Repair inherited CO verification omissions after real signature collisions

A new direct regression shows the pair checker accepts ordered output tuples
(a,b,a) and(a,b,b). The real preparation/refinement routine gives the two PIs
equal signatures and all COs equal signatures; both traversal lengths match.
Their output sharing cannot be matched by a PI permutation, but the C/Go checker
checks only signatures for CO rows. The new regression fails before repair; a
consistent(a,b,a)->(b,a,b) renaming remains accepted after repair.

A separate18-AND static Go fixture produces an actual level-seven signature of
zero, using the source's full initialization, weighted propagation and equal-child
bias. Duplicate the DAG so candidates stay tied and uniquification adds no salts;
only the first copy drives two opposite-phase outputs. The fixture asserts zero
root/CO signatures and one candidate class after the real Gia_Iso2ManPerform.
Both second-generation reduction policies incorrectly merge those outputs before
repair and delete the complemented output. First-generation reduction is a
passing control. An early fixture draft omitted the equal-child bias and failed
its own construction precondition; that discarded draft is not bug evidence.
The corrected fixture's red log reaches the actual pair/reduction failures.

The graph computes a positive conjunction of three support inputs and its
complement; their satisfying-assignment counts differ, proving no PI permutation
can equate them. Static fixture data and all64 assignments over six total PIs
(including the unused clone inputs) require only Go. No C source/test compiler
is involved. Red logs /tmp/abcgo-iso-pair-{collision,mapping}-red.log; focused
repaired isomorphism suite /tmp/abcgo-iso-pair-collision-green.log.

Add the missing CO branch to Gia_Iso2ManCheckIsoPair: require a CO counterpart,
matching output polarity and matching fanin positions in the existing maps.
Keep source signature refinement, traversal and class-quality policies intact.
This deliberately repairs an inherited C correctness bug; retaining the omission
would silently discard real circuit outputs. It is not a replacement GI solver
or Boolean minimization strategy. Other terminal-role and sequential-contract
checks remain for the continuing pair/driver audit.

Frozen full Go-only module passes130 tested packages plus3 commands without
tests, including the direct phase/mapping regressions and all three pipeline
controls. Focused isomorphism race2.310s and GIA vet pass; all handles ended.
Final logs /tmp/abcgo-iso-pair-final-{module.jsonl,race.log,vet.log}. Original
audit2/131; the broader GIA family is not closed by this repair.


## Sequential reduction against a finite transition system

Twelve Go-only cases exercise first-generation reduction and both second-
generation quality policies, two output phases and estimate/actual modes. Two
isomorphic components each contain the coupled cycle q'=input&&!r, r'=q.
Outputs include both component functions, the first function's complement and
a constant. Actual reduction keeps groups[0,1],[2],[3], two representative-state
registers and both original PIs; estimate retains all four registers/outputs.

An independent boolean transition system enumerates all16 initial register
states and64 three-frame PI sequences for each case:36,864 frames total. Check
every retained output and every next-state bit, not just final-frame truth.
This pins source register discovery/order, cyclic feedback, state pruning,
complement polarity and estimate interfaces. First-generation PI permutations
are cone-local[0] for both matched components; the second variant returns none.
Two additional canonical-key cases retain register-cone identity and distinguish
opposite phases and constant cones; constant PI permutations remain empty.
No new sequential discrepancy was found in these cases.

This test file was frozen during the counterexample-permutation dependency gate:
full Go-only module passes130 tested packages plus3 commands without tests; all
12 transition and2 key leaves execute. Focused isomorphism race1.823s and
util/GIA vet pass; all handles ended. Logs /tmp/abcgo-cex-permute-final-
{module.jsonl,util-race.log,gia-race.log,vet.log}; initial focused run
/tmp/abcgo-iso-sequential-first.log. The accompanying utility repair preserves
nonzero initial counterexample bits; see misc/util/REVIEW.md. Remaining dual,
full-convergence, terminal-role and first-generation CEX-driver checks stay open.
Original audit2/131.


## First-generation permutation/CEX diagnostic drivers

Reconcile Gia_IsoTestOld, Gia_IsoTestGenPerm and Gia_IsoTest with the source.
The old driver does full unique labeling, so duplicate POs still report separate
singletons; four output-count/verbosity combinations preserve that source policy.
The permutation generator keeps C's full-range swap loop and tests sizes0/1/2/3/
31/33/257 for bijection. Go's portable RNG replaces platform-dependent libc rand;
no matching C seed stream or modulo bias is claimed. The generator is diagnostic,
not part of canonical refinement.

Six deterministic structural PI permutations exercise the actual duplication/
combination/reduction/canonical-map/CEX-remapping pipeline. A uniquely positioned
asymmetric three-input register transition permits independent raw PI-bit checks.
Enumerate both initial register states and all traces of lengths1/2/3:7008
witnesses. Check every remapped PI record, unchanged initial state, source packet
ownership and scalar-predicted validity in both original and permuted circuits.
This includes the counterexample utility's separately committed prefix repair.

Eight new red report cases precede restoring Vec_IntPrint in the CEX driver.
Zero/one-PI, one/two-frame and both initial-state cases check source formatting
and valid/invalid messages. Two further red cases show inherited C/Go panic for
unused global PIs: canonical orders are cone-local, but the CEX helper requires
a full global PI order. Driver-local isoCexPIOrder projects canonical indices
through sequential support in source CIO order, then pairs unused PIs in index
order. It keeps library metadata cone-local and retains the full-support fast
path. This repairs the index-domain boundary without replacing the isomorphism
algorithm or assuming a particular generated random permutation.

Six additional deterministic permutation cases cover non-prefix support{0,2},
checking768 two-frame witnesses and unchanged local metadata. Completely unused
PI interfaces also preserve CEX validity. All fixture construction/data are Go;
no C source or compiler is a runtime dependency. Red logs
/tmp/abcgo-iso-driver-{red,unused-red}.log; repaired focused logs
/tmp/abcgo-iso-driver-{green,reviewed}.log.

Frozen full Go-only module passes130 tested packages plus3 commands without
tests; all6 full-support,6 partial-support,8 report and2 unused-driver leaves
execute. Focused isomorphism race1.937s and GIA vet pass; all handles ended.
Final logs /tmp/abcgo-iso-driver-final-{module.jsonl,race.log,vet.log}. This closes
the three diagnostic source routines. Full canonicalization/convergence, dual
reduction, remaining terminal-role checks and dormant mux-signature recognition
still require reconciliation before closing the68-routine family. Audit2/131.


## Full canonicalization, complete labels and final stability round

Add108 Go-only canonicalization cases over0/1/2 registers,18 phase patterns and
one/two POs. Independently compute reachable gates from the abstract graph while
keeping all CIs/COs. Verify the source copy-literal mapping is a bijection onto
the new graph, preserves every selected node kind and phased edge, preserves
RO/RI pairing and identity strings, removes only unreachable gates and leaves
source structural/ordinary mark fields unchanged. Exhaustively evaluate all1008
CI assignments across these fixtures, comparing every retained node to the
abstract boolean model under the derived input permutation. This checks graph
and transition preservation, not universal completeness of canonical labeling
for arbitrary graph-isomorphism instances.

Another108 cases run full refinement twice with different scratch poisons.
Every object, including dangling gates, receives exactly one label in1..N; the
constant retains label1, output classes follow source singleton semantics, and
labels do not depend on incoming Value fields. Twelve XOR/MUX recognition cases
check source salt indices, complemented input patterns,32-bit wraparound and
untouched fields. The recognition helper remains dormant in the main algorithm,
as it is in C; no new invocation is introduced.

A new hand-traced regression fails before repair: the asymmetric two-PI AND
resolves its last class during backward refinement, and Go skips C's final
stability round. Its old iteration trace0..5,7 must be0..9,11. Restore that
confirmation round even after classes become empty. Keep the explicit iteration
guard: source C's redundant outer loop can hang when the cap is reached; Go
retains bounded simulation and then strictly reduces unresolved entries through
tie assignment. The10000-pass cap is reviewed by inspection, not forced by these
small fixtures. Red /tmp/abcgo-iso-convergence-red.log; repaired focused suite
/tmp/abcgo-iso-full-canonical-green.log.

Frozen full Go-only module passes130 tested packages plus3 commands without
tests; all108 canonicalization,108 label and12 salt leaves execute. Focused
isomorphism race2.018s and GIA vet pass; all handles ended. Final logs
/tmp/abcgo-iso-full-canonical-final-{module.jsonl,race.log,vet.log}. Remaining
dual-output reduction, terminal-role/hash-filter checks and final68-routine
inventory reconciliation stay open. Original package completion remains2/131.

## Requested pause checkpoint

Current canonicalization/convergence work is verified and committed before the
user-requested repository-maintenance pause. No dual-output implementation or
test changes have started. Resume at dual-output reduction and its metadata/
preprocessing boundaries, then the remaining pair-role checks and source-name
inventory. The full GIA and131-package objectives remain incomplete.
