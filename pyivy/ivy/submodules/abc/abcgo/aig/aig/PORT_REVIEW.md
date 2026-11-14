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
cut, high-fanout limits and frontier size bounds. The Just experimental driver
still awaits the Pack implementation before its integration gate can pass.

## Order, timing and visualization reviews

Design review retained the original order-list cursor semantics with typed links
and a snapshot iterator, and centralized forward/reverse timing recomputation.
DOT generation uses a byte buffer, lexical node IDs and local highlight state;
show returns the artifact path so rendering is controlled by the caller.

Separate bug review and tests verify order cursor repair when removing the
previous/next item, reverse-level corruption detection, nonzero PI arrival-level
preservation, and DOT escaping/complemented edges without changing stored levels
or traversal marks. Cleanup now refreshes fanout metadata after deleting nodes
or terminals. AIGER roundtrip tests additionally compare every assignment's
combinational output and latch next-state values, including constraint polarity.

## Frame and property-splitting reviews

Design review retains source frame flags and frame-major PI ordering with typed
mapping slices; graph recursion supports explicit XOR and rewritten node order.
Property splitting uses an exact Shannon-reduced native decision graph instead
of CUDD allocation/refcount APIs, with shared cofactor and conversion helpers.

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
partitioning shares one implementation; exact functional classes use separate
Shannon scratch graphs, avoiding auxiliary nodes in the source. SAT-conflict
parameters remain source API controls but this backend generates no SAT conflicts;
level filtering is retained. Source-named choice/FRAIG drivers perform actual
partition duplication, equivalence proofs and representative/choice transfer.

Separate bug review checks sorted support/output coverage, inverse maps, phase
handling, pool ownership, compaction with a pending small partition before a
large one, and exact reduction of redundant logic. Independent truth tests cover
partitioned mismatch miters and retained alternative implementations. Choice
level review found the source takes the MAXIMUM alternative depth, not the
minimum: corrected the pass, preserved timing-box dependencies and added maximum
choice-depth/box-boundary regressions.

The package-wide live-free gate lexes every Go file, including unfinished drafts.
It ignores comments/literals and reports source locations of free/ABC_FREE calls.
It failed at all four remaining cleanup expansions before nil-assignment fixes,
then passed with zero live calls. This gate is independent of parser readiness.

### Edge retiming and final package integration

Readability/design review: `aigRet.go` now uses graph-local object/edge references, shared edge objects for each connection, native queue slices, and explicit packed-latch storage in `aigRet_storage.go`. Compact inline state and overflow words preserve the original two-bit initialization representation without pointer tagging or manual allocation. Graph conversion handles complemented pins, explicit XORs, buffers, and feedback through register edges. The original demonstration in `aigTest.go` now uses Edge values and current original API names.

Separate bug review: checked consumer-versus-producer latch ordering, inline/overflow transitions, carry propagation, arbitrary graph IDs, zero-step metadata preservation, initialized forward behavior, and traversal termination at external terminals. Forward retiming computes complemented AND/XOR initialization; backward retiming explicitly retains the original algorithm's unconstrained-initial-state (VOID) contract, and therefore does not promise preservation of a specific initial state. Multiple clock domains are rejected before mutation instead of mixing domains. Actual state movement invalidates obsolete flop-number/representative tracking and sets `RetimingChanged`; single-domain membership is rebuilt. Shared duplication copies this flag.

Validation: randomized 10,000-operation packed-state deque oracle, explicit 10/11 and word-boundary crossings through 100 latches, forward sequential trace comparisons including complemented pins/XOR/feedback, and backward structural checks. Full `GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./aig/aig` passes. The package-wide lexical no-C-free scan is included in that suite, with zero live `free`/`ABC_FREE` calls.

Counterexample ownership integration: Graph now exposes Set/Get/Clear/Take combinational witness methods matching GIA. Design review keeps storage private and uses explicit borrowed/transfer ownership. Separate bug review verified CI-order conversion into ABC's register-first counterexample layout, copied input ownership, and preserving an existing model after invalid setter arguments. Ownership/packing regression and full AIG race suite pass.

Terminal metadata follow-up: source calls to `Aig_ManSetRegNum` in array and node duplication, representative duplication, retiming conversion, and frame expansion finalize the stored true PI/PO counts and CI/CO IDs. Source-derived regressions cover those producers, including both representative traversal modes on a sequential graph. Proof FRA, SSW, CEC, and frame-manager paths consume representative duplicates; FRA frame construction and interpolation array duplication also use the affected producers. FRA counterexample handling reads stored true PI counts, and SSW simulation uses terminal IDs, so preserving this metadata affects proof integration.
