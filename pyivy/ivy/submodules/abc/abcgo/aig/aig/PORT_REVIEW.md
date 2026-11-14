# AIG port review

The full package remains under conversion. These reviews apply to the converted
core, not to the remaining generated algorithm files.

## Design and readability review

Reviewed graph.go, aig.go, aigMan.go, aigDfs.go, aigOper.go, aigObj.go,
aigCheck.go, aigTruth.go, aigTable.go, aigMem.go and aiger_model.go separately
from correctness review.

* Stable Edge IDs replace tagged pointers; Graph owns node payloads and deletion
  holes. Manager-aware accessors replace unsafe object pointer access.
* Output creation has one implementation: Graph.AddOutput creates an explicit
  Output node and Aig_ObjCreateCo delegates to it. This makes output references
  enumerable by the fanout package.
* Shared helpers centralize fanin arity, output-driver unwrapping, level/phase
  recomputation, and fanout refresh after mutations. Multi-step rewrites defer
  rebuilding until the graph is connected again.
* Removed the obsolete ordered-only evaluation helper and corrected the Graph
  ownership/order documentation. Graph evaluation supports rewritten DAG order.
* AIGER conversion is an isolated provider adapter. It normalizes XOR/buffer
  structure into a fresh AND graph without modifying the source.
* Memory APIs use owned byte slices and explicit pool accounting, with no C
  allocator or unsafe pointer arithmetic.

## Separate bug review

Findings fixed with regression coverage:

* Direct output references had no enumerable consumer. All outputs are explicit
  nodes now; fanout and object mutation suites exercise the invariant.
* Fanout scans visited deleted objects and both slots of unary nodes. Header
  scans now use live objects and actual node arity.
* Replacement could rebuild fanout metadata while the old node was temporarily
  disconnected. Mutation batching and per-pin removal fix this.
* Evaluation assumed topological node IDs after rewriting. Recursive evaluation
  now detects cycles and handles buffers, explicit outputs, and explicit XORs.
* Explicit XOR cleanup left stale hash entries; truth computation rejected XORs.
  Both paths are covered by TestExplicitXorTruthAndDanglingCleanup.
* CO cleanup checked the output edge rather than its driver, retaining constant
  zero outputs. TestExplicitConstantOutputCleanup covers explicit output nodes.
* Live object counts included deletion holes. The manager cleanup test checks
  live counts and enumeration while retaining stable IDs.
* Reverse DFS omitted dangling nodes. It now traverses actual consumer edges
  over all internal nodes, with a dangling-node regression.

Additional tests exhaust Boolean constructions, manager duplication/miter,
register cleanup, composition and transfer assignments, support/cut collection,
phase corruption, pool ownership/reuse, and AIGER binary roundtrip. Full-package
build/review remains a required gate when the remaining files are converted.

## Dominator worker review

Design: typed cuts and slices, cached 63-bit signatures with authoritative node
lists, generic sorted union instead of unrolled C cases, local analysis marks and
references, and traversal order derived from the graph rather than numeric IDs.
Separate bug review fixed ancestor-related cut double-dereference using cut
barriers, refreshed stale signatures, and excluded dangling output-unreachable
cones. Worker validation includes 2048 original-C merge comparisons and a
minimal-path-cut oracle on 40 random graphs, plus selected-core race tests.

## MFFC worker review

Readability review consolidated duplicated reference/label traversals. Separate
bug review rejected negative reference counts before mutation and verified
cone-local rollback, avoiding a whole-manager ref snapshot. Worker tests compare
MFFC counts/support to an independent randomized fanout-closure oracle and check
rollback size and retained references.

## Justification and window review

Justification keeps ABC's bounded greedy search rather than claiming UNSAT on a
failed attempt. Ternary operators are centralized and support explicit XOR and
output nodes. Separate bug review corrected false witnesses for constant targets,
handled traversal-ID wrap, and validated partial support assignments. Window cut
expansion uses typed frontier/visited slices and one shared node-arity helper;
marks are cleared by defer. Tests cover reconvergence, truth of the resulting
cut, high-fanout limits and frontier size bounds. Packing has since been
implemented and source-audited: initialization uses the original two-word AIG
random stream, and 1,166 compiled-C operations match words, cares, signatures,
insertion results, candidates, counters and final statistics. The source-named
care diagnostic is disabled as in C; explicit WriteCare remains available.
The sequential justification review now matches 4,310 compiled-C queries,
including support order and every object's scratch state, and the complete
packing/justification experiment. It restored transparent CO handling and failed
attempt support, shared traversal rollover, and validated witness CI indices.
The experimental comparison excludes elapsed time and includes successful
targets, an impossible cone, both histograms and final scratch cleanup.

## Order, timing and visualization reviews

Design review retained the original order-list cursor semantics with typed links
and a snapshot iterator, and centralized forward/reverse timing recomputation.
DOT generation uses a byte buffer, lexical node IDs and local highlight state;
show returns the artifact path so rendering is controlled by the caller.

The sequential visualization audit restored source titles/ranks, terminal
labels/colors, highlight styling and file naming. Eight standard DOT artifacts
match compiled C byte for byte. Go preserves the complete node scratch/cache
snapshot and supports named labels, XOR/buffer shapes and high dangling logic.
Graphviz parsing, the 200/201-node boundary and failure-without-overwrite checks
pass, along with full AIG/ABCI race suites.

Separate bug review and tests verify order cursor repair when removing the
previous/next item, reverse-level corruption detection, nonzero PI arrival-level
preservation, and DOT escaping/complemented edges without changing stored levels
or traversal marks. Cleanup now refreshes fanout metadata after deleting nodes
or terminals. AIGER roundtrip tests additionally compare every assignment's
combinational output and latch next-state values, including constraint polarity.

## Frame and property-splitting reviews

Design review retains source frame flags and frame-major PI ordering with typed
mapping slices; graph recursion supports explicit XOR and rewritten node order.
The sequential audit replaced the earlier Shannon substitute with the source
CUDD pipeline using bdd/cudd/core: canonical BDD construction, symmetric sifting,
MSB-first cofactor cubes and BDD-to-MUX conversion. Subset selection uses the
source AIG random stream and does not depend on terminal scratch IDs. Compiled
original C matches 4,096 pipeline cases and 508 complete splitter calls, including
sequential next-state outputs. BDD ownership checks exposed and repaired stale
reference counts after reordering in the dependency. Detailed evidence and the
remaining package audit scope are in ../../SEQUENTIAL_AUDIT.md.

Separate bug review uses InvalidEdge to distinguish unmapped objects from the
constant-one edge, preserves complemented roots and register next-state outputs,
and fixes initialized expansions that incorrectly advertised register CIs they
did not create. Graph.Check now centrally validates register/constraint counts.
Tests exhaust frame sequences, free-state/enlarge behavior, complemented split
properties, disjoint property reconstruction, semantic-zero cofactors and subset
ownership. Split integration also exposed/fixed the older And-only recursive
manager duplicator when copying explicit XOR, Buffer and Output nodes.

## Interpolation and partition reviews

Interpolation contracts were checked against Cnf_DeriveSimple and
Sto_ManChangeLastClause: ordinary mode asserts every A/B output; relation mode
moves the last A unit to B and shares that unit's regular node variable first.
The earlier Shannon projection backend was rejected by the sequential audit
and removed. Drivers now live in proof/aiginterp and use the original CNF,
BSAT/solver2 and proof-interpolation workflows. Reparameterization preserves
the original zero-clause-ID assertion; a C oracle additionally confirms SAT
after repairing only those IDs. It is an upstream experimental routine that
does not currently provide a usable reparameterization. See SEQUENTIAL_AUDIT.md
for the source evidence, migration and remaining package review scope.

Partition design uses typed pooled blocks/support records, sorted support sets,
bitmap scoring and explicit forward/inverse CI maps. Smart register/output
partitioning shares one implementation. The partitioned choice/FRAIG drivers
call an explicit FRA backend, forwarding conflict and level limits; proof/fra
owns the source-signature entry points to avoid an import cycle. Their saved
partition-to-parent copy maps survive FRA's scratch-map use. The old description
of Shannon functional classification and ignored conflict limits is obsolete.
Constructive choicing follows C's caller-supplied equivalence contract, checks
the previous version's destination copy map and transfers output-driver
representatives without adding a separate equivalence search.

Separate bug review checks sorted support/output coverage, inverse maps, phase
handling, pool ownership, compaction with a pending small partition before a
large one, and exact reduction of redundant logic. Independent truth tests cover
partitioned mismatch miters and retained alternative implementations. Choice
level review found the source takes the MAXIMUM alternative depth, not the
minimum: corrected the pass, preserved timing-box dependencies and added maximum
choice-depth/box-boundary regressions.

The sequential register-partition review restored Aig_ManDfsNodes' C contract:
include reachable inputs, visit choice alternatives before ordinary fanins,
skip returned COs/constants and leave current traversal marks on the cone.
Register-partition creation now uses that helper instead of duplicating DFS,
and writes reverse mappings in source DFS order followed by selected LOs.
Aig_ManRegPartitionLinear is an unfinished diagnostic in upstream C; its Go
entry point now traverses and returns nil instead of inventing partitions.
Direct regressions failed before each behavioral repair; full AIG and SSW
race suites pass. Other DFS entry points still need their own source review.

The package-wide live-free gate lexes every Go file, including unfinished drafts.
It ignores comments/literals and reports source locations of free/ABC_FREE calls.
It failed at all four remaining cleanup expansions before nil-assignment fixes,
then passed with zero live calls. This gate is independent of parser readiness.

### Edge retiming and integration checkpoints

Readability/design review: `aigRet.go` now uses graph-local object/edge references, shared edge objects for each connection, native queue slices, and explicit packed-latch storage in `aigRet_storage.go`. Compact inline state and overflow words preserve the original two-bit initialization representation without pointer tagging or manual allocation. Graph conversion handles complemented pins, explicit XORs, buffers, and feedback through register edges. The original demonstration in `aigTest.go` now uses Edge values and current original API names.

Separate bug review: checked consumer-versus-producer latch ordering, inline/overflow transitions, carry propagation, arbitrary graph IDs, nonpositive depth limits, initialized forward behavior, and traversal termination at external terminals. Forward retiming computes complemented AND/XOR initialization; backward retiming discards initial values, inserts VOID latches and reconstructs them as zero without solving reset constraints, so it need not preserve zero-reset traces. Multiple clock domains are rejected before mutation instead of mixing domains. Changed register layout or encoding invalidates obsolete flop-number/representative tracking, including direct manager reconstruction; single-domain membership is rebuilt. Unchanged layout and encoding retain correspondence. The driver sets `RetimingChanged` when moves occur. Shared duplication copies this flag.

Validation: randomized 10,000-operation packed-state deque oracle, explicit 10/11 and word-boundary crossings through 100 latches, forward sequential trace comparisons including complemented pins/XOR/feedback, and backward structural checks. Full `GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./aig/aig` passes. The package-wide lexical no-C-free scan is included in that suite, with zero live `free`/`ABC_FREE` calls.

Counterexample ownership integration: Graph now exposes Set/Get/Clear/Take combinational witness methods matching GIA. Design review keeps storage private and uses explicit borrowed/transfer ownership. Separate bug review verified CI-order conversion into ABC's register-first counterexample layout, copied input ownership, and preserving an existing model after invalid setter arguments. Ownership/packing regression and full AIG race suite pass.

Terminal metadata follow-up: source calls to `Aig_ManSetRegNum` in array and node duplication, representative duplication, retiming conversion, and frame expansion finalize the stored true PI/PO counts and CI/CO IDs. Source-derived regressions cover those producers, including both representative traversal modes on a sequential graph. Proof FRA, SSW, CEC, and frame-manager paths consume representative duplicates; FRA frame construction and interpolation array duplication also use the affected producers. FRA counterexample handling reads stored true PI counts, and SSW simulation uses terminal IDs, so preserving this metadata affects proof integration.

### Sequential printing checkpoint

Expression and ordinary/verbose object, cone, manager and control-fanout
printers were checked against the original C implementations. Red regressions
preceded repairs to expression root formatting/XOR recognition, unique fallback
input names, source DFS scope, null diagnostics and source object formatting.
The native fixture matches 29,704 queries across all three-input truth
functions, including complemented objects and traversal scratch state. Stable
Go CI IDs replace C addresses in one header. Independent Eqn evaluation covers
all 2,048 assignments plus supplied-name collisions and repeated physical XOR
parity through complemented outputs. Physical XOR/buffer support remains a Go
extension; XOR operand occurrences must retain their parity.

Full AIG and ABCI race suites pass (33.424s/28.894s), with the later XOR parity
regression passing separately. This checkpoint does not close the package-wide
audit: full BLIF/Verilog dump writers and the remaining file inventory still
need source reconciliation. See SEQUENTIAL_AUDIT.md for detailed evidence.

BLIF naming follow-up: a new independent-reader regression failed for supplied
port names that duplicated generated constant/gate/output nets. Generated names
now reserve all supplied input/output names first and use the same unique-name
helper as expression printing. Six parsed round trips, each evaluated at all
four input assignments, cover both input/output naming and reserved suffixes.
Full dump-writer source validation remains in progress.

Full dump-writer source checkpoint: 4,096 C/Go BLIF/Verilog readbacks cover all
256 three-input functions with zero through three registers. All PO/next-state
values, zero reset, interface/register dimensions and rising clock declarations
match; explicit XOR/buffer and named sequential cases also pass. Writes preserve
graph scratch and payloads. Full AIG/ABCI race checks pass (55.776s/28.930s),
followed by focused race checks for the latest additions. This closes the dump
writer comparison; the adjacent manager first-PO flip has a newly found source
representation mismatch that still needs its separate repair checkpoint.

First-PO flip follow-up: repaired Aig_ManFlipFirstPo to complement the explicit
CO's child, as C does, preserving terminal identity. Before repair both the
structural regression and independent BLIF/Verilog readbacks failed. Sixteen
compiled-C flips, all input assignments, reference ownership and double-flip
restoration now pass; full AIG/FRA/interpolation race checks pass. Adjacent
statistics/live-object predicate regressions are the next open checkpoint.

Manager diagnostics follow-up: red regressions exposed stale CO depths,
missing conditional type counts, and an extra deleted-slot restriction in
Aig_ManHasNoGaps. Statistics now reuse Aig_ManLevels and match source fields;
the predicate now uses C's live-object equation. A compiled-C fixture matches
ordinary/deleted/sequential/typed-choice diagnostic cases byte for byte, and
focused race tests pass. Full AIG tests pass on this state; ABCI verification
is logged in /tmp/abcgo-manager-current-review.log. Manager cleanup's typed
buffer behavior and final file inventory remain open audit work.

Manager cleanup source repair: two red regressions exposed output-reachability
semantics and buffer-inclusive return counts in Aig_ManCleanup. It now uses C's
zero-reference AND/XOR roots and the existing recursive pin-release helper,
counting gates and batching fanout rebuilds. A native fixture matches 1,024
cleanup calls across 512 graphs, including typed buffers/XORs, weak choices,
reference holds, hash state and physical fanouts. Full AIG/ABCI/SAIG race suites
and the expanded focused race test pass.

Sequential cleanup's former general sweep is now local to sclCleanupLogic.
Its typed-buffer, terminal-removal and Basic patch/delete boundaries remain an
open source-fidelity review. The earlier sequential tests are not proof of
matching those C boundaries; no claim of package completion is made here.

CO patch review: a red regression showed Aig_ObjPatchFanin0 changed an unrelated
CO's auxiliary level. It now updates only the target CO and its physical fanout
pin, matching aigObj.c, while retaining recursive old-cone deletion. Full AIG,
ABCI and SAIG race checks pass. Sequential wrapper/state comparisons are still
in progress as a separate checkpoint.

Sequential cleanup boundary repair: four red regressions exposed buffer
liveness, Basic's patch/delete return count, terminal-ID refresh and LO union
scratch mismatches. Both wrappers now use source reference-driven cleanup;
the general reachability sweep is removed. Native state comparison matches
1,024 calls across 512 graphs, including metadata, live levels/traversal IDs,
references, hash state and circular fanout order. Red native ordering evidence
also justified removing redundant post-deletion rebuilds.

An original C boundary fixture confirms its retained-buffer reference to a
removed LO. Go reclaims only those invalid-input buffer cones, preserving
independent buffers. Shared/nested cone tests check graph validity and sequential
traces with and without fanout indices. Full AIG/ABCI race and focused source
checks pass; SAIG status is in /tmp/abcgo-seq-cleanup-final-review.log. These
checks close the reopened wrapper boundary review, not the entire AIG audit.

Post-cleanup integration: a fresh repository-wide run passes 129 tested packages
and three packages without tests. This result predates the miter repair below.

Miter source repair: red regressions exposed missing zero-input phase
normalization, inherited CI levels and cleanup-invalidated copy bindings.
ExtractMiter now follows C's fresh-interface/copy/XOR/phase/CO sequence and keeps
copied cones allocated. Native truth/count/level/ownership comparison and 64
small root-pair checks pass; the full AIG race suite passes in 61.561s. The
nonpositive Start hint contract and final file/function inventory remain open.

Start/order checkpoint: a red negative-capacity regression justified matching
C's nonpositive soft-limit contract. Circular order matched source inspection;
1,000 independent model steps and a high-ID growth test now cover its cursor,
link and counter invariants. Focused race checks pass. The file inventory has
reopened Aig_ManSupports: its per-output DFS is an algorithm substitution for
C's pooled propagation and still needs repair before package completion.

Pooled-support source repair: Aig_ManSupports' repeated per-output DFS was still
an algorithm substitution. Red scratch/raw-index regressions preceded restoring
C's reference-counted propagation with the existing typed pool. Rows own their
copies; record release/reuse follows physical consumers, and merges fill the
pooled slice directly. Forward/unary/raw-output extensions retain input-position
semantics without overwriting caller payloads. Per-cone DFS is now test-only.

A separate red test exposed stale-CioID dependence in register supports. The
register matrix now filters whole-support rows as in C, and register grouping
shares that matrix. Native comparison on 256 shuffled-interface graphs matches
supports, register rows, diagnostics, traversal/reference state and terminal
cleanup. Wide shared-cone/row-ownership checks pass. Full AIG race checks and
FRA/SSW/ABCI caller checks pass; latest focused native/race checks also cover
the final shared helper. See SEQUENTIAL_AUDIT.md for the precise validation
order and remaining partition/file inventory scope.

Smart partition scoring checkpoint: a red regression exposed the incorrect
Base2Log(1) special case at 101 support variables. Native comparison now matches
8,190 scalar choices and 25 complete smart-partition boundary cases. Full
AIG/FRA/ABCI race checks pass, followed by the expanded focused native/race
checks. Original compaction container bugs remain safely repaired in Go.

Open correctness defect: the source-compatibility partition copier turns native
XOR into AND. Independent Go and original-C probes show a false-equivalence
miter for XOR versus AND when EXOR is represented natively; expanded C XOR is
correct. The existing native-XOR copy test enshrines that error and must be
replaced with an explicit upstream-bug demonstration plus red/green functional
regressions. AND-only backend compatibility and reverse mapping of expansion
nodes need verification in the next repair. Package review is not complete.

Native-XOR partition copying is now repaired: expand to existing AND-only XOR
logic instead of copying as AND. Red clone/miter tests and an original-C fixture
demonstrate the previous false-equivalence bug. A separate red copy-map test
caught lost reverse-map polarity, now retained. Exhaustive signed two-input
cases, forward/inverse/reverse mappings, constants/aliases and FRA/choice scratch
map restoration pass focused checks. Auxiliary expansion nodes deliberately
have no original-object binding. This closes the above XOR defect; the overall
package audit and final function inventory remain incomplete.
Full AIG/FRA/ABCI race checks pass (68.859s/12.335s/29.071s); the final focused
signed-XOR/map/buffer/native-C checks pass with the race detector in 2.316s.

AIGER snapshot adapter reviewed against the original writer's constants,
register placement and inverted constraints. A new 128-graph mixed-representation
matrix checks all four-input assignments before/after binary serialization,
including reordered interfaces, forward references, CO fanins and registers.
No adapter repair needed. Operation reconciliation found Aig_MuxTest was an
invented small check; a red missing-artifact test preceded restoring C's actual
seeded 2,000-MUX experiment, object reports, cleanup and test1.blif export. A
compiled original-C boundary observer confirms identical canonical AND
structures for all outputs and both 5,503-object reports. glibc random sequence
compatibility is explicit; sibling node IDs need not match C evaluation order.
Remaining reconciliation includes the separate aigPartSat.go source contracts.
Full uncached AIG race suite passes in 69.915s, and go vet passes. No new
whole-repository test claim is made for this experiment-only production change.
