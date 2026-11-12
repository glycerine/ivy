# Ivy port progress

Completed API groups include attributes, central node/manager representation,
header edge/node/init helpers, paged memory, basic manager lifecycle, structural
hashing, object construction/deletion, Boolean operations, canonicalization,
fanout tracking, validation and combinational balancing. These groups are tested
together independently of the remaining drafts. Whole-package compilation is
still blocked by unported replacement/manager/history/DFS/checking and advanced
mapping/fraiging/rewriting/sequential routines. No draft file is hidden or
excluded from the package build.

Attribute integer values are `int32`; pointer attributes are Go `any` values.
Managed fixed-size attributes use zeroed byte slices instead of raw allocator
addresses. Slot cells keep `FetchSpot` addresses stable across table growth.
`SetCleanup` replaces callback fields with typed functions. Managers are not
concurrency-safe; independent managers may be used concurrently.

## Separate review passes

The compactness/readability review eliminated duplicated array size/capacity
fields and C pointer/integer reinterpretation. Table lengths are slice lengths,
unused slots are nil, and integer cells allocate only when needed. The managed
mode and attribute size remain because they describe ownership and entry shape.

The bug review covered original-header defects: growth only doubled once even
when an ID lay far beyond that size; initialization repeatedly wrote the
requested slot instead of clearing all new slots; and managed fetch allocated a
new buffer when one already existed. The Go implementation grows to include the
requested ID, relies on zero-initialized slices and preserves existing buffers.
It also rejects negative IDs and kind mismatches, normalizes typed nil pointers,
and makes stop/cleanup idempotent under callback reentrancy.

Focused validation (does not imply the whole Ivy package is ported):

```
go test attr.go attr_test.go
go test -race attr.go attr_test.go
```

Tests exercise far-ID growth, old slot pointers after growth, integer/pointer
updates, managed buffer initialization/persistence, cleanup ordering and invalid
operations. No draft files are hidden or excluded from the package build.

## Core representation, lifecycle and table review

Regular nodes own state; complemented pointers are cached Go proxies referencing
regular nodes. This retains pointer equality without unsafe pointer tagging.
Integer edges are uint32 literals with a complement bit. Node IDs and traversal
fields retain int32 APIs. Manager vectors are typed slices, with sizes/capacities
derived directly; no local vector wrappers or extra dependencies are needed.

Memory consists of owned Go node pages and a free-slot slice. Fetch/recycle clears
node state while preserving the cached complement identity. Stop releases graph
references and invalidates ownership; foreign, duplicate, live/published, ghost
and constant-node recycling is rejected. There is no address arithmetic or
alignment trick. Constant and ghost identities follow the original manager.

Basic lifecycle routines live in ivyLife.go. Their already-ported duplicate
bodies were removed from ivyMan.go; its other routines remain visible drafts.
StartFrom copies only the constant/input boundary and establishes pEquiv maps,
matching the original function's scope.

The structural table uses a Go map of the complete structural key: both signed
fanin pointers, object type and latch initial state. Built-in hashing replaces
manual open addressing/prime-table allocation. Replacement updates preserve the
new object's identity, and reject absent/duplicate or structurally inconsistent
entries. Resize rebuilds the map; profile reports its entry count rather than
C-specific bucket-chain statistics.

The compactness/readability pass removed duplicated table/vector allocation
metadata and cast-based memory operations. The bug-review pass corrected the
original fanout-polarity helper reading polarity from the driving node instead
of the fanout edge, protected signed edge/leaf shifts, preserved complement proxy
identity through clean/overwrite/recycle, and tightened memory/hash ownership
checks. Tests cover phase/init algebra, bit boundaries, page expansion, manager
isolation, lifecycle, mapping, table key distinctions and identity transfer.

Earlier core/table checks (the current combined suite also needs the construction and balance files listed below):

```
go test ivy.go ivyMem.go ivyLife.go ivyTable.go ivyCore_test.go ivyTable_test.go
go test -race ivy.go ivyMem.go ivyLife.go ivyTable.go ivyCore_test.go ivyTable_test.go
```

A lifecycle review regression also caught manager-owned references surviving when
StopMemory had been called before ManStop. ManStop now always clears graph/table
references even when the memory pages were already stopped. Combined attribute,
core, lifecycle and table tests pass together under the race detector.

## Balance and adjacent construction APIs

Algebraic balancing is now ported, along with real Boolean constructors,
canonical AND/XOR/latch construction, node creation/connection/deletion,
intrusive fanout tracking, cleanup, buffer resolution, traversal IDs, graph
validation and cycle detection. History creation and representative lookup are
ported; the remaining history algorithms are still drafts. Matching completed
bodies were moved out of draft object/manager/utility/check/DFS/history files;
all unported bodies remain visible. Balancing explicitly requires a valid
combinational graph, as the original algorithm cannot balance latch leaves.

The separate readability pass uses pointer-stable scratch slices for recursive
supergates, derives their sizes from len, shares the actual PI constructor with
StartFrom, and shares cycle-DFS state across roots to keep validation linear.
It retains the original level-order pairing and structural-sharing heuristic.

The separate bug pass fixed native XOR supergate multiplicity: duplicate leaves
cancel, opposite leaves toggle parity, and rebuilding XOR cones must preserve
operand multiplicity. The original generic AND-style deduplication was wrong
for native XORs. Zero/singleton cones are handled directly, and pairing checks
same-regular operands before creating a hash ghost. Object creation also now
uses XOR rather than AND when computing a native XOR's phase.

Regressions compare all 256 assignments of an eight-input skewed AND before and
after balancing and require depth 3. They cover both level-update options,
complemented/constant outputs, contradictory ANDs, native XOR repeated/opposite
operands, Boolean/MUX/majority truth tables, fanout insertion and recursive cone
deletion. Tests use the real constructors and storage, with no test-only stubs.
The focused core, table, operation and balance suite passes; whole-package
compilation still fails in the visible unported ivyCheck.go/ivyCut.go routines
and later drafts. Replacement/buffer propagation, the remaining manager
transformations, full DFS/level maintenance, full history, mapping, cuts,
fraiging, rewriting and sequential optimization remain unfinished.

The full DFS/check helper groups are now also ported. DFS returns typed ID
slices; cone collection uses typed node slices; levelization returns slices per
level. Sequential DFS includes latch-input cones and returns latch IDs through
an optional slice pointer. Required levels preserve the original XOR delay of
two. Choice-aware level computation retains the original representative-ring
semantics; cycle validation includes choice dependencies.

A separate checker review replaced list-length assertions with a safe
bidirectional fanout audit: missing links, incorrect previous pointers and
cyclic lists return failure. A DFS bug regression caught repeated frontier
entries restoring a stale mark; frontier marks are now saved once per regular
node and restored on return. Tests cover DFS order, shared cone collection,
recomputed/required levels, latch boundaries and choice-induced cycles. These
groups pass together with all completed APIs under the race detector.

Manager duplication, sequential reachability cleanup, self-feeding latch checks
and statistics are also ported in ivyManage.go. Duplication maps latch boundaries
before copying combinational cones, then reconnects feedback. It preserves XOR
behavior (the source duplicate routine incorrectly rebuilt every gate as AND).
Sequential cleanup first disconnects the entire unreachable component, allowing
unreachable latch cycles to be recycled safely. Self-feed checks terminate on
cycles that do not include the requested root.

The separate manager bug review distinguishes pEquiv mappings to another manager
from choice links within the same manager. This is necessary because StartFrom
and Dup deliberately leave mapping pointers on the source; DFS/validation must
not treat those mappings as local graph choices. Regression fixtures cover XOR
duplication, two-latch feedback duplication and deletion, and unrelated cycles.
The combined completed-API suite passes under the race detector.

Current remaining manager routines are Frames, PropagateBuffers and MakeSeq.
Object replacement and buffer-fanin repair remain drafts, as do the larger
history/rewriting/fraiging/mapping/sequential algorithms. The whole Ivy package
still does not compile; completed-group tests explicitly select the real ported
files and their tests, without excluding or disguising any draft file.

## Cuts, timeframes and complete manager/object groups

ivyCut.go is fully ported: sequential reconvergence expansion, Boolean pivot
cuts, cut hashing/derivation/dominance filtering, bounded cut enumeration and
original diagnostic drivers. Typed slices replace temporary Vec containers.
The bounded cut records retain fixed arrays and explicit live counts/limits,
because these describe a bounded record rather than duplicate slice metadata.
Cut enumeration owns its result per call instead of reusing C's static store.
Structural MUX recognition is in ivyMux.go and covers complemented data/control.

The separate cuts bug review fixed marks leaking on a failed Boolean pivot
search, safe handling of buffers/constants and duplicate fanins, and overflow
when deriving a cut larger than its bound. Dominance filtering checks rejection
before deleting dominated records. Tests assert the exact cuts of a three-input
tree, size limits, per-call ownership, latch-depth leaves, reconvergent Boolean
cuts, mark cleanup, buffer/duplicate expansion and every MUX input polarity.

All ivyMan.c routines are now implemented across ivyLife.go, ivyManage.go and
ivyFrames.go. All ivyObj.c routines are implemented across ivyCreate.go and
ivyObj.go. ivyUpdate.go adds the required incremental/reverse/reset-level
algorithms, and history choice creation plus dependency search live beside the
previous history-creation helpers. The remaining unported utility/history
algorithms remain visible in their source-derived files.

The separate manager/object readability pass retained typed slices, explicit
identity transfer and cached complement proxies. The bug pass corrected frame
mapping row overlap (stride is the number of object slots, including ID zero),
XOR reconstruction during unrolling, and sequential conversion's deleted-object
counter. Replacement rejects dependency cycles before mutation and preserves
old node/proxy identity while recycling the replacement slot. Its freeTop option
must be zero: recycling the identity being preserved is invalid; all original
callers use zero. Bounded dependency search tracks remaining depth so a shared
cone first reached on a long path can still be explored via a shorter path.

Regressions exhaust three-frame XOR state transitions with initialized and free
initial state, verify mapping bounds, replacement truth/identity/hash/fanout
links, complemented buffer propagation, rejection of dependency cycles without
mutation, and sequential conversion's visible initial state and all transitions.
The combined completed-API suite, including the previous attributes, graph,
balance, DFS and manager tests, passes under the race detector. Full-package
compilation now first fails in the visible ivyCutTrav.go draft; advanced cut
traversal, history, rewriting, fraiging, mapping and sequential synthesis remain
unfinished. No unported implementation is hidden or replaced with a stub.

## Traversal cut group

ivyCutTrav.go is fully ported: elementary bit cuts, depth-based and size-based
volume traversal, topological collection, wordwise merging and implication,
bounded-leaf root cut collection, and the original diagnostic driver. Temporary
node/frontier vectors and bit-cut rows are typed slices. The original default
volume policy is six levels; its node-limit parameter was unused in C and
remains advisory. Volume2 is the original size-driven alternative.

The separate readability review replaced raw arena pointers with Go-owned word
slices and per-call result ownership, deriving sizes from len. The separate bug
review deduplicated shared nodes during volume collection, handled unary buffers
and constants safely, checked cycles before changing marks, and ensured scratch
traversal IDs/marks are cleared on exit. Bit-cut filtering first checks whether
an existing cut rejects a candidate before compacting the output: the original
interleaved filtering could corrupt existing cuts when a later cut rejected it.

Regression tests compare traversal cuts against expansion cuts, check result
ownership across subsequent calls, shared-cone uniqueness, node-limit volume
collection, cycle rejection without mark/ID mutation, stable arena slices after
growth, rejected-candidate preservation, unary buffers, and a 63-node tree whose
bit cuts cross a 32-bit word boundary. Each returned cut of that tree is checked
as an input separator. The connected completed-API suite passes with the race
detector. Full Ivy compilation now first fails in visible ivyDsd.go drafts; no
unported source has been excluded or replaced with a stub.

## Five-variable DSD group

ivyDsd.go now implements the complete active ivyDsd.c API: truth cofactors and
support detection, recursive AND/XOR decomposition, disjoint-cofactor MUX and
three-input majority recognition, truth reconstruction, textual printing, Ivy
graph construction and the active diagnostics. The source's disabled #if 0
file-driven experiments are not active APIs. Prime functions return failure,
matching the original recognizer's intentionally restricted decomposition basis.

The separate readability pass replaces C unions/bitfields with explicit 32-bit
packing and a small bounded record, preserving original packed tree words.
Temporary trees/frontiers use slices and output accepts io.Writer. There is no
static shared decomposition tree; the diagnostic counter is atomic. Graph
construction uses the actual typed Ivy constructors.

The separate bug review adds packed-field bounds and checks every child index
is smaller than its parent before truth computation, printing or construction,
so malformed cyclic trees cannot recurse forever. It preserves the source's
frontier behavior: sequential leaf encodings are changed to their object IDs;
this constructor does not synthesize extra latch stages from leaf-depth bits.
The fixed five-variable/32-bit truth-table scope is explicit.

Exhaustive tests cover every four-variable function (65,536 truth tables).
A differential test compiles the actual source-C decomposition section with a
minimal vector shim and compares recognition plus every packed tree record for
all four-variable functions and 4,096 seeded five-variable functions. Additional
tests exercise 10,000 random five-variable truth tables, reconstruct every
recognized three-variable function in the actual Ivy manager and compare all
32 assignments, check packing/majority printing, and reject cyclic trees. The
connected completed-API suite passes both race testing and go vet. Whole Ivy
compilation now first fails in the visible ivyFastMap.go draft.

## Fast K-LUT mapping

ivyFastMap.go is now fully ported. This source file implements technology
mapping, rather than structural hash lookup: delay-oriented support selection,
sorted bounded support union, mapped-area collection, required times, exact
reference/dereference area, three area-recovery passes, frontier compaction and
the original alternative recovery helpers. Mapping requires a valid buffer-free
combinational graph and LUT sizes 2..32. Structural lookup remains ivyTable.go.

The separate readability pass replaced variable-length raw memory records with
typed support slices, deriving counts from len. Manager-owned support records
and levelized LUT slices replace pointer arithmetic and shared static temporary
buffers. Mapping duration is stored per mapping manager. There are no global
mutable mapping scratch buffers or duplicated byte-allocation metadata.

The separate bug pass made failed/aliased support merges transactional, kept
recovered support leaves sorted for subsequent merges, and stopped using the
constant's support record as scratch storage. Mapping walks actual topology,
including dangling nodes, so identity-preserving rewrites that create references
to larger IDs work correctly. Foreign data in the manager's pData slot is
rejected. Reference/dereference mismatches are checked; stop is idempotent.
The original recovery2 alternative intentionally omits timing rejection; the
active Perform sequence uses the timing-constrained main recovery routine.

Tests perform 10,000 randomized support unions against independent sets, with
failed aliased merges preserving the original. One hundred seeded random DAGs
verify support bounds/order, separator correctness, unchanged canonical hash
lookup, LUT delays and independently recomputed mapped references; exact-area
queries must restore every reference count. Tree fixtures test LUT sizes 2..8,
delay/area recovery and release; a replacement fixture tests non-ID-topological
mapping. A source-C differential test compiles the actual original Node/Merge
functions and compares every delay and support on 200 seeded random AIGs.
The connected suite passes race testing and vet. Whole-package compilation now
first fails in visible ivyFraig.go drafts, and the remaining fraiging/history/
rewriting/sequential synthesis scope is still unfinished.

## Functional reduction (FRAIG)

The complete active ivyFraig.c family is now ported across ivyFraigSim.go,
ivyFraigSat.go, ivyFraigDriver.go and ivyFraigCone.go. The source-name file points
to these implementations. This includes normalized word simulation, candidate
classes and refinement, distinguishing-pattern scoring, supergate/MUX CNF,
incremental implication checks, native XOR CNF, SAT counterexamples, simulation
resaturation, sweep/reduction, miter proving and the iterative resource-limited
prove driver. The optional CUDD-backed equivalence entry point uses a real local
reduced ordered BDD with unique/apply tables, removing the external C dependency.
Fresh cone extraction returns the existing aig.Graph type and preserves phase.

The separate readability pass replaced overloaded fanout-pointer storage with
lazy typed FRAIG state per node. Real fanout lists therefore remain usable while
simulation and SAT metadata exist. Simulation data, patterns, SAT frontiers and
class links have explicit types; the manager owns its RNG and scratch storage.
Intrusive class APIs retain their original list behavior, while class grouping
uses Go maps. Temporary CNF literals are slices. The SAT bridge uses the real
ported bsat.Solver, including activity factors, per-call/global budgets and model
values; no solver or equivalence routine is a stub.

The separate bug review corrected phase handling for constant-one and
complemented simulated outputs, so saved models actually satisfy the output.
Startup ownership validation happens before mutation, simulation state is
released, and cyclic class links are rejected. Pair counting now returns the
actual number rather than the source's unconditional zero. Native XORs retain
XOR behavior in simulation, CNF and copying instead of being treated as ANDs.
Constant SAT variable zero is explicitly constrained true. Topological copying
handles rewritten object IDs. The prove driver avoids division by zero on an
empty output boundary; its rewriting block remains inactive as in the active C
source. Original solver global nInsLimit semantics count propagations, while the
outer returned inspection statistic counts watched-clause inspections; these
units are kept distinct through the bsat bridge.

Tests cover all simulation phase opcodes against a scalar reference and 8,192
cases against the unchanged source-C simulation routine (harness in
test_vectors/). Class partition/refinement, pattern neighborhoods and scoring,
repeated ownership rejection, stop/restart and live fanout preservation have
regressions. CNF tests exhaust every three-input MUX polarity and native XOR
assignment; SAT/BDD equivalence and counterexamples are checked against scalar
truth. Deliberately false simulation candidates force SAT-driven resimulation.
Forty seeded random reductions exhaust all 16 input assignments and check BDD
results. Miter SAT/UNSAT, prove-driver ownership, fresh cone SAT and marks are
covered. The original ABC executable's ifraig command processes the
fraig_associativity.blif fixture; reduced node count and every output assignment
match the Go implementation. Connected tests pass under the race detector and
go vet. Whole Ivy compilation now first fails in the visible ivyHaig.go draft.

The connected test/vet scripts now run from abcgo/ so imports resolve through
its module. They include all completed files plus the FRAIG files and tests;
`bash /tmp/test-ivy-core.sh -race` and `bash /tmp/vet-ivy-core.sh` are the exact
current commands. No experiments directory is built or tested.

## HAIG history family

The remaining ivyHaig.c APIs are now ported in ivyHaig.go: history start,
transfer/stop, choice-ring counting, postprocessing, modified three-valued
simulation and diagnostic output. Earlier representative, object creation and
choice insertion helpers remain in ivyHistoryCreate.go. Transfer requires a
fresh input boundary and moves ownership to the destination. Stop clears working
mappings and is idempotent. New terminal/native-XOR construction records valid
history mappings instead of the original helper's unsupported-type assertion.

The separate readability pass uses typed latch-origin slices and local
simulation-value maps. Diagnostic simulation does not overwrite structural Init
or Level fields. It preserves the source's modified rule where unknown dominates
zero in AND simulation, its choice compatibility rule and ten-round trace.
Choice traversal checks malformed/cyclic rings instead of looping indefinitely.

The separate bug review repaired structural hashing when original history
latches are reset to DC. Formerly distinct initial-state keys may collide;
identity-preserving buffers now retain source mappings while sharing the canonical
reset latch. That regression exposed and fixed an earlier object-replacement
validation bug: Co includes latches, but the original replacement algorithm
explicitly permits latch operands and buffers them; only PO/assert operands must
be rejected. Transfer ownership no longer leaves two managers able to stop the
same history. Simulation scratch cannot corrupt graph validation or hash keys.

Tests cover associative choice creation/counting/representatives, live graph
checks after simulation, stop/repeated stop, canonical DC-latch collisions,
transferred ownership, continued terminal/XOR recording, all modified
three-valued AND combinations and malformed rings. A new-latch diagnostic trace
is checked against test_vectors/haig_new_latch.trace, including unknown-to-known
state propagation and preserved latch structural fields. The connected race
suite and vet pass. Whole Ivy compilation now first fails in visible
ivyMulti.go/ivyMulti8.go drafts.
