# Sequential package audit

Review every package in `go list ./...` order, one at a time. For each package,
inspect production files and existing tests against the original C, review
correctness and compactness, add a red regression before fixing any uncovered
bug, and run meaningful package checks. Commit coherent package changes.
Use `test_vectors` for fixtures. Prior reviews are evidence to inspect, not
proof that this audit has been completed. No parallel package or agent work.

Baseline: 131 packages (128 with tests, three command packages without tests).
The repository-wide uncached tests passed before this audit.

Current status: **2/131 original packages reviewed**. The two driver packages
extracted during the AIG audit bring the current module total to 133; their
source workflows and tests are covered by that audit. Current review: `aig/gia`.

| # | Package | Status | Evidence |
| --- | --- | --- | --- |
| 1 | `.` | reviewed | Root review below; race tests and vet pass. |
| 2 | `aig/aig` | reviewed | All 52 files reconciled; native source matrices, final race suite, vet and module integration gates pass. |
| 3 | `aig/gia` | in progress | Graph/manager, copies, hashing, balancing, windows, bi-decomposition, circuit-cofactoring, BDD collapse, counterexample replay/minimization and chronological/conflict-analysis circuit SAT reviewed; remaining families and final inventory pending. |
| 4 | `aig/gia/mf` | pending | |
| 5 | `aig/gia/satcost` | pending | |
| 6 | `aig/hop` | pending | |
| 7 | `aig/ioa` | pending | |
| 8 | `aig/ivy` | pending | |
| 9 | `aig/miniaig` | pending | |
| 10 | `aig/saig` | pending | |
| 11 | `base/abc` | pending | |
| 12 | `base/abci` | pending | |
| 13 | `base/acb` | pending | |
| 14 | `base/bac` | pending | |
| 15 | `base/cba` | pending | |
| 16 | `base/cmd` | pending | |
| 17 | `base/exor` | pending | |
| 18 | `base/io` | pending | |
| 19 | `base/main` | pending | |
| 20 | `base/pla` | pending | |
| 21 | `base/test` | pending | |
| 22 | `base/ver` | pending | |
| 23 | `base/wlc` | pending | |
| 24 | `base/wln` | pending | |
| 25 | `bdd/bbr` | pending | |
| 26 | `bdd/cas` | pending | |
| 27 | `bdd/cudd` | pending | |
| 28 | `bdd/cudd/core` | pending | |
| 29 | `bdd/dsd` | pending | |
| 30 | `bdd/epd` | pending | |
| 31 | `bdd/extrab` | pending | |
| 32 | `bdd/extrab/cascade` | pending | |
| 33 | `bdd/extrab/primes` | pending | |
| 34 | `bdd/extrab/sets` | pending | |
| 35 | `bdd/llb` | pending | |
| 36 | `bdd/mtr` | pending | |
| 37 | `bdd/reo` | pending | |
| 38 | `bool/bdc` | pending | |
| 39 | `bool/dec` | pending | |
| 40 | `bool/deco` | pending | |
| 41 | `bool/kit` | pending | |
| 42 | `bool/lucky` | pending | |
| 43 | `bool/rpo` | pending | |
| 44 | `bool/rsb` | pending | |
| 45 | `cmd/abc-demo` | pending | |
| 46 | `cmd/abc-starter` | pending | |
| 47 | `integration` | pending | |
| 48 | `map/amap` | pending | |
| 49 | `map/amap/formula` | pending | |
| 50 | `map/cov` | pending | |
| 51 | `map/cov/minimize` | pending | |
| 52 | `map/fpga` | pending | |
| 53 | `map/if` | pending | |
| 54 | `map/mapper` | pending | |
| 55 | `map/mio` | pending | |
| 56 | `map/mpm` | pending | |
| 57 | `map/scl` | pending | |
| 58 | `map/super` | pending | |
| 59 | `misc/avl` | pending | |
| 60 | `misc/bar` | pending | |
| 61 | `misc/bbl` | pending | |
| 62 | `misc/bzlib` | pending | |
| 63 | `misc/espresso` | pending | |
| 64 | `misc/extra` | pending | |
| 65 | `misc/extra/npn` | pending | |
| 66 | `misc/hash` | pending | |
| 67 | `misc/mem` | pending | |
| 68 | `misc/mvc` | pending | |
| 69 | `misc/nm` | pending | |
| 70 | `misc/parse` | pending | |
| 71 | `misc/st` | pending | |
| 72 | `misc/tim` | pending | |
| 73 | `misc/util` | pending | |
| 74 | `misc/vec` | pending | |
| 75 | `misc/zlib` | pending | |
| 76 | `opt/cgt` | pending | |
| 77 | `opt/csw` | pending | |
| 78 | `opt/cut` | pending | |
| 79 | `opt/dar` | pending | |
| 80 | `opt/dar/core` | pending | |
| 81 | `opt/dau` | pending | |
| 82 | `opt/dau/canon` | pending | |
| 83 | `opt/dsc` | pending | |
| 84 | `opt/fret` | pending | |
| 85 | `opt/fsim` | pending | |
| 86 | `opt/fx` | pending | |
| 87 | `opt/fxch` | pending | |
| 88 | `opt/fxu` | pending | |
| 89 | `opt/lpk` | pending | |
| 90 | `opt/mfs` | pending | |
| 91 | `opt/nwk` | pending | |
| 92 | `opt/res` | pending | |
| 93 | `opt/ret` | pending | |
| 94 | `opt/rwr` | pending | |
| 95 | `opt/rwt` | pending | |
| 96 | `opt/sbd` | pending | |
| 97 | `opt/sfm` | pending | |
| 98 | `opt/sim` | pending | |
| 99 | `phys/place` | pending | |
| 100 | `phys/place/cmd/bookshelf` | pending | |
| 101 | `proof/abs` | pending | |
| 102 | `proof/acec` | pending | |
| 103 | `proof/cec` | pending | |
| 104 | `proof/cec/core` | pending | |
| 105 | `proof/dch` | pending | |
| 106 | `proof/fra` | pending | |
| 107 | `proof/fracomb` | pending | |
| 108 | `proof/fraig` | pending | |
| 109 | `proof/framgr` | pending | |
| 110 | `proof/int` | pending | |
| 111 | `proof/int2` | pending | |
| 112 | `proof/live` | pending | |
| 113 | `proof/pdr` | pending | |
| 114 | `proof/ssc` | pending | |
| 115 | `proof/ssw` | pending | |
| 116 | `sat/bmc` | pending | |
| 117 | `sat/bsat` | pending | |
| 118 | `sat/bsat2` | pending | |
| 119 | `sat/cnf` | pending | |
| 120 | `sat/csat` | pending | |
| 121 | `sat/glucose` | pending | |
| 122 | `sat/glucose2` | pending | |
| 123 | `sat/lsat` | pending | |
| 124 | `sat/minisatp` | pending | |
| 125 | `sat/msat` | pending | |
| 126 | `sat/proof` | pending | |
| 127 | `sat/psat` | pending | |
| 128 | `sat/satoko` | pending | |
| 129 | `sat/satoko/utils` | pending | |
| 130 | `sat/satoko/utils/vec` | pending | |
| 131 | `sat/xsat` | pending | |

## 1. Root package

Inspected every production file (`demo.go`, `starter.go`, `template.go`,
`assert.go`) and all three test files against `src/demo.c`, `src/starter.c`,
and `src/template.c`. Reviewed command sequencing and failure exits, framework
ownership, command quoting, shell invocation, worker limits, writer locking,
read-error cleanup, parameter borrowing, and assertion caller locations.

No new correctness bug found. Existing tests exercise default/resyn2 command
order, every default failure position, parser round trips, framework cleanup,
worker concurrency, command limits, read failures, real shell commands and
borrowed template state. `go test -race . -count=1` and `go vet .` pass.

Compactness: the command helper, synchronized writer and bounded worker slots
already consolidate repetitive C control flow. Further consolidation would
obscure ownership or source flow; no speculative rewrite made. Existing
intentional source departures remain explicit in STARTER_TEMPLATE_REVIEW.md
and DEMO_REVIEW.md (including strict CLI integer parsing, wall-time reports,
long lines, single-CPU behavior and cleanup on failures). This review does
not claim integration coverage for every registered synthesis command.

## 2. aig/aig (reviewed)

Inventoried 52 production Go files and the earlier review reports. Historical
reports still mention incomplete conversion and selected-file validation;
they cannot establish completion of this audit. Next review the graph/edge
representation and mutation contracts, then the algorithm files and tests
against src/aig/aig, recording source comparisons and any red regressions.

### aig/aig: representation and mutation review checkpoint

Inspected `graph.go`, `aigObj.go`, `aigInterface.go`, `aigConstruction.go`,
`aigMan.go` (through cleanup/reference operations), `aigFrames.go`,
`aigTruth.go`, `aigTable.go`, and `counterexample.go`; cross-checked terminal
and mutation contracts in `src/aig/aig/aig.h`, `aigObj.c` and `aigMan.c`.
The complete package audit remains in progress, including the other algorithm
files and deeper source comparisons.

Investigated apparently stale terminal indices after interface reordering.
An initial proposed refresh passed local tests but failed the downstream
`TestHideExposePreservesCioIDsAndBadListOrder`. Source inspection confirmed
that register hiding deliberately preserves old CioIds until an explicit reset.
Rejected the proposed behavior change, documented that contract, and added a
regression for preservation followed by explicit index refresh and evaluation.
This is not counted as a discovered production bug. Removed a redundant fanout
rebuild in `Aig_ObjCreateCi`: `NewInput` already performs the rebuild.

The package audit remains incomplete. Earlier full AIG race checks passed. Final sequential tests
(`go test -p 1 ./aig/aig ./aig/saig ./proof/dch -count=1`) pass after the
corrected contract test.

### aig/aig: duplication review checkpoint

Inspected all of `aigDup.go`, focusing source comparisons on simple/partial
copies, output-free and flop-only copies, representative phases, miter creation
and selected sequential outputs in `src/aig/aig/aigDup.c`. Reviewed existing
duplication tests and the shared copy-map/manager initialization operations.

Found output-role metadata incorrectly inherited by `Aig_ManCreateMiter`.
C allocates a fresh manager at `aigDup.c:1049`; the Go `dupStart` copied operand
constraint and assertion counts into the new comparison manager. Added
`TestCreateMiterDoesNotInheritOutputRoles` before changing code; both cases
failed with inherited counts of one. Miter creation now uses the existing
`Aig_ManStartFrom` interface-copy helper without cloning output-role metadata.
The regression also verifies the graph and evaluates both input assignments.
Removed a duplicate assignment of RetimingChanged in `dupMetadata`.

Full AIG race tests pass. The directly affected interpolation caller also passes
`go test -p 1 ./proof/int -count=1`. The full `aig/aig`
audit remains in progress; this checkpoint is not a claim that every
specialized duplication/metadata combination has been exhaustively tested.

### aig/aig: output selection and traversal checks

Source inspection of output-free, flop-only, selected-output, OR-output and
unsolved-output copies found stale assertion counts copied by `dupStart`.
These C constructors start fresh managers and do not copy nAsserts. Eight
new regression cases failed before the fix. The three shared Go construction
paths now clear Assertions; tests check interface counts, register counts,
all input assignments, and unchanged source metadata.

Reviewed `aigDfs.go` and the small representative/choice/data/phase accessors.
Compared ordering, composition and supergate operations against `aigDfs.c`.
Found `Aig_ManVerifyTopoOrder` delegated to Graph.Check, which accepts DAGs
whose object IDs are not topological and does not check timing-box ordering.
Added red tests for a forward buffer reference and an early box-output CI;
both were incorrectly accepted. Restored C's object-order traversal, terminal
index setup, timing-box checks, traversal marks and successful index cleanup.
The implementation retains Graph.Check as the separate graph-validity API.

Full AIG race tests pass with both repairs. Sequential downstream checks for
SSW, FRA and ABCI also pass for the output-selection changes. The
remaining algorithm families still require review; aig/aig is not complete.

### aig/aig: cuts and buffered supergates

Inspected all of `aigCuts.go` and its existing tests. Compared cut cost,
replacement-slot selection, ordered merging, dominance, truth stretching,
support minimization and computation loops with `src/aig/aig/aigCuts.c`.
Existing tests include exhaustive small ordered merges, original-C merge/phase
comparisons, random truth and completeness oracles, constant cuts and failed
merge preservation. Added independent six-input truth validation across all
64 assignments, explicit XORs, inverted buffers and multiword tables; it passes.
No new cut-computation bug found in this review. Existing Go extensions include
explicit zero-support cut validity and buffer/XOR handling.

Found that supergate collection stopped at buffer objects rather than following
the source's Aig_ObjReal_rec calls on both fanins (`aigDfs.c:1092-1093`). Added
positive and negative two-buffer-chain regressions before the repair; both
failed with the buffer retained as a leaf. Collection now uses the existing
phase-aware buffer resolver on each child. Tests check leaf membership and
all eight input assignments without changing the graph's function.

Full AIG race tests pass with the new cut and supergate regressions. The SAIG
caller also passes its full uncached package tests. Remaining families include dominators, factoring,
fanout/timing, memory, partitioning, representatives, retiming, simulation,
canonicalization and visualization; their earlier reviews still require
inspection in this sequential audit.

### aig/aig: storage, fanout and timing review checkpoint

Inspected `aigMem.go`, `aigFanout.go`, `fanout_status.go`, `aigTiming.go` and
`aigCheck.go`, their existing tests, and the corresponding C allocation,
recycling, circular-link, level-update and mark-check routines. Go pools use
owned slices instead of C chunks, so allocation-byte accounting is not a
byte-for-byte reproduction of C arena reservations. Larger stepped entries
are intentionally retained until Stop in C (the recycle free is commented
out); retained Go storage is therefore not treated as a leak to fix.

Added six invalid-recycle cases for small and large stepped allocations:
foreign storage, inconsistent size and double recycling. Tests ensure failed
operations preserve allocated storage and the small entry remains reusable.
They pass without a production change. Fanout tests already cover circular
insert/remove, duplicate pins, growth, deletion holes and unary objects.
Timing tests check forward/reverse levels, corruption and CI arrival levels.

Found `Aig_ManCheckMarkA` inspected deletion holes, unlike C's live-object
iteration in `aigCheck.c:136`. Added a red cleanup regression; it panicked on
an old marked dangling node after removal. The checker now skips deleted
slots, and the test separately confirms it still rejects live marks. Full
uncached AIG race tests pass with this fix and the new pool ownership cases.

Compactness review found no further safe consolidation in these files:
circular per-pin storage remains explicit, allocation ownership has one map
per pool, and timing shares graph recomputation helpers. Go's whole-graph
level recomputation costs more than C's incremental scheduling; replacing it
requires a dedicated performance/correctness change rather than deleting
checks during this audit. The package audit remains in progress.

### aig/aig: representatives and interpolation fidelity gap

Inspected `aigRepr.go` completely and compared representative linking,
transitive lookup, copy precedence, phase conversion, transfer, recursive and
ordered duplication, cleanup, rehashing, choice filtering and mapped classes
against `src/aig/aig/aigRepr.c`. Existing tests cover constant representatives,
chains, cycles, retained latches, transfer and choice-loop rejection. Added
alternating-phase representative chains through both duplication modes,
checking every input assignment and the output graph invariants. Full uncached
AIG race tests pass with the added cases.

Inspected `aigInter.go` and `aigRepar.go` against their complete C entry-point
workflows. **Unresolved fidelity defect:** the Go interpolation and
reparameterization functions substitute Shannon projection/self-comparison for
C's CNF, incremental SAT and proof interpolation. Existing semantic tests
cannot prove an original-algorithm port; these functions prevent completing
the package audit under the user's explicit fidelity requirement.

Repair dependencies identified from current code:

- `sat/bsat` contains the original solver2 and interpolation machinery, including
  partitioned clauses, on-line OR/AND resolution and proof records.
- `sat/cnf` owns the source CNF derivation and imports AIG.
- Five BSAT production files also import AIG (`satSolver2Graph.go`,
  `satAigInterpolation.go`, `satSolver2i.go`, `satInterA.go`, `satInterB.go`).
  Direct imports from AIG back into these owners would create cycles.
- A dependency refactor is required before connecting the original drivers.
  An optional callback with an absent default, another solver, or more
  semantic-only tests would not resolve the required fidelity gap.
- The source reparameterization driver itself contains suspicious indexing:
  its PI creation loop calls Aig_IthVar(pBase,i) while iterating k; its
  interpolant-connection clauses use the original pCnf mapping rather than
  pCnfInter. These require focused source/C-oracle evidence during the port,
  not silent reproduction or an invented replacement algorithm.

This is active required work, not an accepted approximation or a reason to
mark the package complete. Further inspection of the dependency boundary and
red behavioral/source regressions is the next step for this family.

### Native interpolation driver repair: first working integration

Implemented `proof/aiginterp.Aig_ManInter` above AIG/CNF/BSAT to resolve the
import cycle without an optional provider or solver substitution. It follows
`src/aig/aig/aigInter.c` using Cnf_DeriveSimple, lifted B variables, the source
A/B clause-store boundary, optional relation-clause movement, CI equalities,
root marking, BSAT solving and Inta resolution interpolation. Existing owner
APIs were sufficient; no solver implementation was copied.

Added an ACB source-failure regression before switching its caller: an
inconsistent A partition must return nil during clause insertion. The old
Shannon path returned a graph and the test failed. ACB's Gia_ManInterOne now
uses the native driver. The driver tests check all 256 two-input predicate
pairs, source early-failure behavior, graph validity, Craig conditions and
relation-variable ordering. Full new-driver and ACB race tests pass.

This is partial repair: ABCI still calls the legacy AIG implementation, and
InterFast, InterTest and InterRepar still need their original workflows.
The new driver package is supporting implementation work and must also receive
final review/validation; it does not reduce the original 131-package scope.
Required next steps are to move the remaining callers, remove the replacement
algorithms, port the solver2/proof reparameterization driver, and add C-oracle
and broader integration evidence. No optional callback fallback is introduced.

### Native interpolation: ABCI and incremental driver migration

Added a red ABCI regression for the source's inconsistent-partition insertion
failure, then changed both ABCI single-interpolant callers to the native proof
driver. The full ABCI package passes after the switch. Added the incremental
source workflow: mapped CNFs with exposed CO variables, variable lifting,
shared-input equality clauses, simplification, and per-output assumptions on
one persistent BSAT solver. All 256 two-input predicate pairs are checked with
four output roles each (ordinary, complemented, constant one and constant zero).
The driver race tests pass before final migration validation.

Removed the old projection-based Aig_ManInter/Aig_ManInterFast implementations
from aig/aig. Their public entry points now live in proof/aiginterp; all
repository callers were migrated. This is an API ownership relocation needed
to break graph/CNF/solver cycles, not an optional backend. The original semantic
tests moved with the functions. Their relation check now tests A=>I and I=>!B,
not equality to a particular strongest projection; C's proof interpolation
need not choose that strongest interpolant. Inconsistent partitions follow the
verified source failure behavior rather than producing a constant graph.

Aig_ManInterTest and Aig_ManInterRepar still contain replacement algorithms and
remain required repair work. Their presence keeps the AIG audit incomplete.
Full driver, ACB and ABCI race tests pass after migration. The final AIG race
check also covers removal of unused projection helpers.

### Native solver2 self-test repair

Ported Aig_ManInterTest from aigRepar.c into proof/aiginterp: mapped CNF,
partition-A internal variable flags, marked A clauses, lifted B copy, equal
used CIs, XORed COs and the unequal-output assumption. This calls the existing
BSAT solver2 and proof logger. The source's interpolant extraction is commented
out, so no alternate interpolant is synthesized by the self-test. Clause stats
and interpolation timing are reported as in the source.

Added narrow source-style solver2 bridges for allocation, variable/clause
partition marking, statistics and shutdown. A unit test verifies assumption
refutation with an actual resolution core and subsequent incremental SAT.
Found stale models visible after the new variable-growth API; added a red test
before invalidating prior results, matching existing public solver behavior.
Removed the old Shannon self-comparison function from AIG and retained its
separate append test. Sixteen exhaustive truth-function tests now exercise
the actual solver2 self-comparison and verify input graphs are unchanged.

Reparameterization still requires the source solver2 proof interpolant driver;
this audit is not complete. Full uncached BSAT, native driver and AIG race
tests pass with these changes.

### Stored solver2 proof interpolants

Implemented Solver2.ProofInterpolant using satProof.c's proof-root/antecedent
algorithm: collect reachable resolution nodes, form A-root disjunctions over
shared literals (B roots are one), then combine each resolution chain with OR
for A-local pivots and AND otherwise. Local edge maps replace temporary Id
mutation, preserving clauses and proofs for incremental reuse. No projection
or second SAT engine is used. Requested common-variable order, including
unused inputs, is retained.

Tests check an actual four-clause solver2 refutation, every shared assignment,
repeated extraction with unchanged proof core, absence before refutation and
reordered/unused common inputs. Full uncached BSAT race checks pass.
This is a dependency for reparameterization, not completion of that driver.

Source inspection suggests the reparameterization initial SAT formula may
already violate its UNSAT assertion on a nonconstant one-output function.
This remains a hypothesis requiring a C fixture; no replacement constraints
or algorithm have been invented to make it pass. Next verify the initial
formula against original solver2 and clause helpers before porting the driver.

### Reparameterization: original-C initial-stage oracle

Added a compiled-C oracle using the exact clause helpers and initial solver
setup extracted from current `src/aig/aig/aigRepar.c`, linked against original
solver2/SAT/memory/proof/store sources. Its fixture provides the valid free-CO
CNF for an identity function (output v0 equals input v1; v2 is constant one).
Unused interpolation callbacks are link stubs; the SAT implementation and all
constraints under investigation are original C. Core reporting is disabled
only in the fixture process; timeouts remain failures.

The unmodified source prefix aborts in sat_solver2_addclause on `Id != 0`:
aigRepar.c supplies the now-forbidden zero root-clause IDs. A second diagnostic
build replaces only those IDs with increasing positive IDs, leaving the
original constraints unchanged. It returns status=1 (SAT) under the original
first equality assumption, contradicting the subsequent source assertion that
status is l_False. Both outcomes are now regression-tested, and the full native-driver race
suite passes. The production C
and Go algorithms were not changed to manufacture UNSAT.

This verifies two upstream failures before proof extraction, rather than
merely guessing that the indexing looks wrong. It also proves that the
existing Go image-preserving projection does not represent this source
execution. Required follow-up: replace that substitute with a clearly traced
source port preserving demonstrated failure behavior, or repair the original
algorithm only with sufficient specification/evidence. Do not silently invent
new reparameterization constraints. AIG remains in progress.

### Reparameterization source driver replaces the invented projection

Added a Go red regression before changing the old entry point: an identity
circuit returned a graph instead of failing on C's forbidden clause ID zero.
The new driver in proof/aiginterp mechanically retains the solver layout,
partition flags, two CNF copies per side, comparisons, assumptions, proof
interpolation, DAR rewriting, incremental CNF feedback and append steps.
Source-zero IDs use a narrow explicit-ID BSAT bridge; the actual solver guard
rejects them. Later source indexing defects are retained and commented, not
silently corrected or replaced with new mathematics.

Removed the projection-based Aig_ManInterRepar and its image-synthesis helper.
Removed the test that demanded that invented successful algorithm; its source
failure replacement checks the actual solver guard and unchanged input graph,
and is paired with the original-C oracle from the previous checkpoint. Clause
primitive and append tests remain. The entry point moved to proof/aiginterp,
consistent with the other interpolation drivers; no repository callers besides
tests used reparameterization.

The source guard makes later steps unreachable for ordinary nonempty CNFs.
Retaining their mechanical translation is traceability, not evidence that
those steps would work if the guard were bypassed. This routine remains
unusable as in upstream C; no successful port claim is made for its intended
mathematical functionality. A future upstream-algorithm repair requires more
specification, not restoration of the removed substitute. Also flagged for the
remaining partition review: the historical report describes Shannon-based
functional classes where source SAT conflict controls are ignored; those
paths require the same fidelity scrutiny before aig/aig can be completed.

Full uncached BSAT, native-driver and AIG race suites pass for this checkpoint.

### Constructive choice source contract

Reviewed the partition FRAIG/choice driver loops against aigPart.c. Contrary
to the historical PORT_REVIEW description, the current partitioned drivers
already use an explicit FRA backend and forward conflict/level limits. The
saved partition copy maps and representative transfer match the source loop;
the existing backend-forwarding regression covers those controls. Corrected
the obsolete report rather than replacing this working backend.

Constructive choicing still had an invented Shannon equivalence search and
omitted the source's previous-version copy-map validation. Added three tests
for missing, invalid and unreachable-object destination mappings; all failed
before the repair. Restored the source check for every live previous object
and removed the extra equivalence search. Equivalent versions remain a caller
precondition, as in C. Added a three-version regression checking each live copy
map, CO identity, representative links, complemented output and all eight input
assignments. This avoids introducing a separate potentially exponential proof
algorithm into a structural accumulation operation.

Uncached AIG, FRA and ABCI race suites pass. The AIG race suite was rerun after
adding the three-version regression. This checkpoint does not complete the
AIG package audit; the remaining families listed above still require review.

### Register partition and DFS source contracts

Compared aigPartReg.go with aigPartReg.c, including seed/frontier costs,
support collection, partition construction, one-hot projection and traversal.
Found that Aig_ManDfsNodes returned only gates and did not start or mark a
traversal. C returns reachable PIs and gates, visits choice alternatives first,
pre-marks the constant and skips COs. Two new red regressions exposed the
missing inputs, missing choice cones/order and stale traversal generation.
Restored that specialized helper without changing other DFS entry points.

A separate red register-partition test exposed the missing constant/internal
cone marks in Aig_ManRegCreatePart. It now calls the repaired source helper,
replacing its duplicate recursive walk, and constructs the reverse map in C's
DFS order followed by the selected register outputs. Tests check mark scope,
CI/internal reverse mappings and unmapped fake/real COs, alongside existing
truth tables, interface counts and one-hot projection coverage.

Aig_ManRegPartitionLinear had invented a complete partitioning algorithm.
The source only performs a traversal and returns NULL; no production Go
callers exist. A new red test showed invented nonnil results for sizes 1, 2
and 100 (and an extra size-zero panic). Restored the explicit nil result and
source traversal, retained a comment identifying the unfinished upstream
routine, and replaced the old test that demanded the invented behavior.

Full uncached AIG and SSW race suites pass. SSW's existing second cone walk
still warrants review when that package's sequential audit is reached; it
was a workaround for the old helper and has not been counted as audited here.
The general DFS functions and remaining AIG algorithm families remain open.

### Levelization source contract

The next DFS review found that Aig_ManLevelize omitted the constant, inputs
and outputs, although C iterates every live object. It also grew levels past
Aig_ManLevelNum instead of enforcing C's upper-bound assertion. Added and
observed red tests for an empty manager, a manager with an unused CI and CO,
and a dangling gate above the output level. The repaired helper preallocates
the source level range and buckets all live objects in ID order. The test also
checks that deletion holes remain absent. The full uncached AIG race suite
passes; no production Go callers currently use this helper.

The general DFS entry points have further source-contract differences still
to resolve: recursive helpers use local sets instead of traversal marks,
choice traversal order differs, array traversal omits the source's initial
constant, and DfsAll includes unreachable gates unlike the source. ABCI's
roundtrip helpers directly call the recursive API and rely on its current
node-only behavior, so restoring it requires reviewing and adapting those
call sites rather than changing the helper in isolation. These are open
findings, not completed audit work.

### Array and choice DFS source contracts

Restored Aig_ManDfsArray and Aig_ManDfsAll_rec to their source contract.
Array traversal starts a traversal, includes the constant first and respects
the specified root range. The recursive helper respects caller-supplied marks
and marks returned objects, eliminating its old repeated whole-cone scans and
linear result-slice duplicate searches. New red tests exposed the omitted
constant, absent traversal setup and ignored seeds. Empty arrays, repeated
roots, complemented CO drivers and unrelated inputs are covered.

Reviewed choice DFS separately from ordinary DFS: the source deliberately
visits fanins before choice alternatives and marks each logic object after
recursion. Restored its traversal setup, premarked terminals and nonnil
equivalence-table requirement. Red regressions exposed ignored recursive
seeds, duplicate appended cones and the missing equivalence precondition.
The Go cycle guard is retained to reject invalid cyclic graphs safely.

AMAP unconditionally used choice DFS, unlike amapGraph.c. Its caller now
selects ordinary or choice DFS based on allocated equivalence storage, exactly
as C does. The existing ordinary reconstruction test and a new explicit source
choice test cover both branches, copied choice links and all input assignments.
The full uncached AIG and AMAP race suites pass. This is a caller integration
repair, not a completed AMAP package audit.

These repairs close the array/choice findings from the previous checkpoint.
Ordinary DFS, preorder DFS, DfsAll and ABCI's recursive callers remain open;
in particular DfsAll is used by several ports as a topological all-object
iterator, while source DfsAll assumes no dangling gates. Those callers require
source-specific review before changing that contract.

### Ordinary and preorder DFS source contracts

Restored ordinary DFS and preorder DFS, including their recursive APIs, from
aigDfs.c. New red regressions showed incorrect alternative ordering, omitted
constant, ignored caller seeds, missing traversal generations/marks and missing
source completeness assertions. Ordinary DFS visits alternatives before fanins;
preorder appends the current object first but deliberately uses ordinary
postorder for its alternatives. Both wrappers now share source setup/count
checks. Recursive calls collect terminals unless the caller premarked them.
Aig_ManDfsNodes now reuses the restored ordinary helper instead of maintaining
a duplicate walk. Existing array/choice/register-partition regressions pass.

Reviewed ABCI's two direct recursive callers before updating them. Reachable
roundtrips now call the source wrapper; all-node choice roundtrips explicitly
initialize traversal seeds and enumerate live gates, including dangling gates
as the source object loop does. A new red regression exposed missing traversal
initialization; it checks repeated calls and deletion holes after the repair.
This caller adjustment does not constitute a full ABCI audit.

Full uncached AIG, AMAP and ABCI race suites pass. Ordinary/preorder DFS findings
are closed for this checkpoint. DfsAll still needs a distinct all-live-object
iterator for ports that use its current behavior; source DfsAll assumes no
dangling gates. The rest of the AIG algorithm review remains incomplete.

### DfsAll and all-object caller separation

Compared DfsAll and all seventeen production Go call sites with their source
object passes. The C routine traverses CO cones, appends unused CIs afterward
without marking them and asserts that every live object was collected. Go's
routine instead enumerated every live object, including dangling gates, and
left traversal IDs unchanged. New tests failed on ordering, traversal setup
and the missing dangling-gate rejection before restoring the C contract.

The callers in SAIG, SSW, abstraction and cut sweeping implement all-node or
all-object passes; none of their corresponding C files calls Aig_ManDfsAll.
They need the former Go behavior, including support for buffers and rewritten
object order. Moved that unchanged traversal into Graph.TopologicalObjects and
migrated all seventeen calls. It visits every live object in dependency order
without changing traversal IDs or object marks. Direct tests cover dangling
logic, deletion holes, a buffer with a forward object-ID reference, regular
unique results, fanin-before-user ordering, preserved scratch fields and cycle
rejection. This names the Go iteration utility explicitly rather than changing
the source algorithm or dropping objects needed by the caller passes.

Full uncached AIG, SAIG, SSW, abstraction and cut-sweeping race suites pass.
The caller migrations are integration changes, not completed audits of those
four packages. Reverse DFS, cone/support/composition helpers and the remaining
AIG algorithm families still require source review; aig/aig remains in progress.

Integration validation after the DFS checkpoints: an uncached, sequential
`go test -p 1 ./... -count=1 -timeout 180s` completed successfully across the
entire module: 129 packages passed tests and three packages have no tests.
The additional proof/aiginterp package accounts for the difference from the
original 131-package inventory. Passing this integration gate does not mark
the still-pending package reviews complete.

### Cone marking, cleaning and DAG-size contracts

Compared all four cone-marking routines and Aig_DagSize against aigDfs.c.
New red regressions exposed traversal through premarked/unmarked boundaries,
modification of CI marks, traversal through buffers/COs, overcounting below
premarked gates and failure to consume cone marks during DAG sizing. Cleaning
also cleared protected copy entries while leaving the actual auxiliary payload
on newly visited gates. These are observable source-contract errors, not just
differences in private traversal implementation.

The four helpers now share a postorder gate-only traversal, stopping whenever
the current mark already equals the requested state. Mark/count/clean stop at
marked gates; unmark stops at unmarked gates. Cleaning clears both typed-copy
and auxiliary-payload representations of C's pData only on newly visited gates.
DAG sizing follows C's count-and-mark followed by unmark, including consumption
of existing cone marks. The prior Go cycle guard is retained. This also avoids
walking an entire protected subtree before checking its boundary mark.

Added an independent compiled-C oracle that includes the original aigDfs.c
and exercises the five operations on the same seeded cone. Go's returned
counts, marks and both pData representations agree with C. Additional tests
cover shared cones, explicit XOR gates, repeated counting, terminal marks and
buffer/CO boundaries. The full uncached AIG race suite passes, including the
compiled oracle. No production repository callers currently use these five
helpers; wider package audits remain pending. Reverse DFS, support and
composition helpers still need review before this DFS family can be closed.

### Support collection and support-size source contracts

Reviewed all six support helpers against aigDfs.c and inspected their callers.
New red regressions exposed retained output-slice contents in SupportNodes,
missing traversal setup/marks, traversal through seeded internal boundaries,
and missing regular/non-CO root preconditions. SupportSizeTest was a different
calculation entirely: it summed output support sizes instead of counting live
logic gates with support at most sixteen, including dangling gates.

Restored source traversal state and gate-boundary behavior. Recursive helpers
use caller traversal seeds; collection appends to the supplied slice, whereas
SupportNodes clears it, starts a traversal and premarks the constant. Scalar
support counting now directly counts marked CIs without allocating intermediate
whole-cone and support slices. Restored SupportSizeTest's node scan, threshold
and source diagnostics. Collection and counting remain separate as in C.

Updated AIG partition/split/register-component callers, SAIG diagnostics and
FRA output diagnostics to pass regular driver objects, matching their source
call sites. The earlier general traversal test now supplies a regular root
where C requires one. These caller adjustments do not close the wider SAIG or
FRA package audits.

Coverage includes cleared and empty root arrays, CO unwrapping, repeated roots,
seeded gates, preservation of recursive result prefixes, marks rather than
slice membership as the visited criterion, buffers, explicit XOR, constant
collection, repeated wrapper calls and the sixteen/seventeen-input boundary.
A new positive fixture initially assumed insertion-order XOR fanins; corrected
its expectation to the graph's canonical fanin order. The final full uncached
AIG race suite passes. SAIG, FRA, the FRA class manager and SSW race suites also
pass with unchanged production code after that test-only correction.

Reverse DFS and transfer/composition source contracts remain open, along with
the remaining AIG algorithm families. The original package audit scope and
pending package count are unchanged.

### Circular fanout accessor prerequisite for reverse DFS

Reverse-DFS inspection exposed an earlier fanout-review gap in aig.go's two
inline accessor ports. Aig_ObjFanout0Int and Aig_ObjFanoutNext scanned graph
objects by ID and formed a zero-terminated sequence; C reads the maintained
circular per-pin list. The scans lost insertion order after link removal and
reinsertion, and accepted calls without the required fanout allocation.
New regressions failed on the maintained head, circular successor and missing
allocation before the repair. The accessors now directly read the source index
with bounds checks, eliminating whole-graph scans for each link.

Inspected all four production accessor loops against Aig_ObjForEachFanout in
llb1Pivot.c, sswDyn.c, absOldSim.c and pdrTsim.c. Their Go loops incorrectly
terminated on zero; restored the source reference-count bound so circular
lists terminate correctly, including paths containing continue statements.
Their production managers already allocate the index. Two direct low-level
SSW/PDR tests omitted that source precondition; added allocation/cleanup to
those fixtures without weakening their behavioral assertions.

Additional accessor coverage checks unary uses, second-pin encoding, empty
lists and self-linked single-consumer rings. The final AIG, SSW and PDR race
suites pass; LLB and abstraction race suites passed with the same production
changes before those test-only fixture updates. The caller migrations do not
constitute full audits of the four dependent packages.

This closes the inline fanout accessor gap found while starting reverse DFS.
Reverse DFS itself still ignores source traversal marks and rebuilds adjacency;
its timing/DAR callers omit source fanout setup. Those remain the next review
and integration work, along with transfer/composition and the rest of AIG.

### Reverse DFS, reverse timing and DAR fanout setup

Restored reverse DFS from aigDfs.c after repairing its accessor prerequisite.
New red tests exposed incorrect maintained-list order, missing traversal setup
and marks, ignored recursive seeds and acceptance of missing fanout allocation.
The wrapper now premarks COs, visits live gates/buffers in object order and
follows the indexed circular fanouts. Recursive calls honor caller seeds and
mark objects after visiting users; the existing Go cycle guard remains.
Removed the duplicate full-graph adjacency construction and result-list-based
visited state.

Reverse timing now enforces C's fanout allocation and clean-MarkA preconditions,
both exposed by red tests. It takes the global bound from output DRIVER levels,
as C does, rather than CO metadata. A separate red test exposed the same field
error in Aig_ManLevelNum; it now reuses the existing source-correct Aig_ManLevels.
Per-object reverse-level calculation follows indexed fanouts, replacing an
entire-graph scan for every gate. Existing timing tests now provide the source
fanout setup; the behavioral checks remain unchanged.

Inspected darCore.c and darRefact.c before updating their Go callers. Rewrite
now performs source cleanup before fanout/timing setup and honors its Fanout
flag. Refactor establishes fanouts unconditionally, as C does. Both release
owned state on Go cancellation/error paths via defer. New lifecycle tests failed
before these repairs, checking index presence during processing and cleanup on
cancellation. The existing reverse-level cancellation test now also checks for
a leaked fanout index and allocates its own index before restarting timing.

Final uncached AIG, DAR core and DAR facade race suites pass. This is integration
work required by the AIG review, not a completed DAR package audit. Transfer,
composition, incremental timing details and the remaining AIG families still
require review; aig/aig remains in progress.

Integration gate after reverse DFS/timing and DAR setup: the full uncached,
sequential `go test -p 1 ./... -count=1 -timeout 180s` passes all 129 packages
with tests; three packages have no tests. This validates the current integration
state, not completion of the original 131-package audit.

Next checkpoint findings from read-only source inspection: transfer resets
the entire copy map and has stricter variable-count handling than C's bounded
CI loop; its constant fast path also differs. Composition's recursive entry
point delegates to the full wrapper instead of honoring C's seeded marks/copy
state, and the wrapper panics where C returns NULL for an undefined PI index.
These require failing regressions before repair, plus truth-table checks and
care when composition appends to the graph's backing node slice.

### Transfer and composition source contracts

Reviewed the two wrappers and recursive helpers against aigDfs.c. Seven new
regressions failed before repair: constant transfer wiped source copy state;
transfer rejected a variable bound larger than the available CI count; direct
transfer trusted stale unmarked copies and ignored marked boundaries; the
wrapper discarded seeded/unrelated copies; undefined composition variables
panicked instead of returning the NULL equivalent; recursive composition
ignored seed mappings and omitted marks/leaf mappings; composition omitted
its source marks when allocating new graph nodes.

Transfer now preserves the source constant/identity fast paths, maps CIs with
C's bounded loop and retains unrelated copy entries. Its recursive helper
uses MarkA as the visited/cache-validity criterion, updating copies and marks
in postorder; the wrapper unmarks afterward. Composition likewise honors
seeded marks, records visited terminal copies and updates internal copies and
marks before the wrapper clears them. Undefined nonnegative PI indices print
the source diagnostic and return InvalidEdge. Node values are copied before
recursion, and mark writes reacquire nodes by ID after allocation so slice
growth cannot leave updates on an obsolete backing array. Go cycle guards
protect both recursive walks.

Retained the existing Go explicit-XOR transfer extension with a dedicated
truth-table test. The source AND-inverter transfer cases follow C; composition
continues to require an AND-inverter root cone, as the previous Go port did.
This checkpoint does not claim C equivalence for unsupported primitive kinds.

Independent truth-table tests cover all 512 two-input composition combinations
(16 original functions, 16 replacements, two variables), every assignment,
all 16 transferred functions, extra destination CIs, complemented and constant
roots, and post-operation graph/mark consistency. Direct tests also force node
slice growth and verify seeded/unrelated copy preservation. The full uncached
AIG race suite passes, including the existing compiled-C cone oracle. These
four helpers have no production repository callers, so the focused package
verification matches the affected call scope.

The transfer/composition findings from the previous checkpoint are closed for
AND-inverter cones. Incremental timing and the remaining AIG algorithm families
remain open; this does not complete aig/aig or the original package inventory.

### Incremental timing propagation and CO level metadata

Compared both incremental routines with aigTiming.c. Red regressions exposed
whole-graph recomputation that changed unrelated levels, ignored seeded work
marks, updated CO metadata outside the source node-only pass and accepted
unsupported root kinds. Source setup checks were missing from forward update
and level verification. Graph.Check and level verification incorrectly treated
CO cached levels as authoritative, although source propagation leaves them
unchanged. Tests demonstrated that mismatch before changing the checker.

Restored the two source work queues, ordered by old forward/reverse levels.
MarkA prevents duplicate pending entries and is cleared when an entry is
processed. The reverse queue reads each bucket's current length so newly
allocated fanins with old level zero can be scheduled into the same bucket.
Unrelated state and traversal generations remain unchanged. Graph.Check now
ignores CO level metadata while retaining its gate-level checks; explicit
forward verification follows C's logic-node scope and requires fanouts.

The Go DAR drivers genuinely need a complete reverse-level refresh for their
buffer-preserving replacements. Moved that existing behavior into the explicit
Graph.RefreshReverseLevels method and migrated both callers. Source-named
Aig_ManUpdateReverseLevel now requires a real logic node instead of accepting
a constant as a refresh sentinel.

An independent compiled-C fixture links the original timing, fanout and utility
implementations and confirms seeded boundaries, forward increases/decreases,
unchanged CO fields, reverse propagation through a dynamically extended
old-level bucket, unrelated reverse state and cleared work marks. The matching
Go tests pass, as do full uncached AIG, DAR core and DAR facade race suites.
The existing genuine reverse-level verification check is retained: C's
diagnostic predicate mistakenly checks forward levels, while Go already
detects corrupt reverse levels and has a regression for that behavior.

Read-only inspection found two remaining representation issues to address
before closing timing/graph review. C's Aig_ObjLevelNew charges an explicit XOR
an extra level; Go currently uses one level. Also inspect CO reference access:
Go stores an output-anchor reference in Node.Refs, whereas C's accessor reports
actual fanout references. These need their own red tests and caller review.
Other AIG families and the rest of the original package inventory remain open.

### XOR structural timing versus source choice annotations

Source comparison found two deliberately different conventions: the header's
Aig_ObjLevelNew charges an explicit XOR two levels, whereas aigDfs.c's choice
pass increments every logic node by one and resets unvisited nodes to zero.
Three new regressions failed before the repair, exposing one-level structural
XORs, incorrect incremental/recomputed levels and acceptance of an invalid
one-level XOR in ordinary structural validation. The choice regression also
exposed rejection of C's legitimate zeroed dangling-node annotations.

Centralized structural gate cost and used it in node-level calculation, full
recomputation and checking. XOR now costs two levels; AND remains one. Kept the
choice pass's source unit cost. To preserve strong validation under both
conventions, optional packed metadata records structural, choice or reset
annotations using two bits per object. Ordinary graphs allocate no annotation
storage; full structural recomputation releases it. Incremental structural
updates and explicit level writes clear affected annotations. This metadata
supports Go's stronger checker and does not change C's level algorithms.

Added regressions for structural creation/recomputation, input arrival levels,
source unit-cost choice levels, zeroed dangling nodes, corruption detection,
mode transitions, newly appended nodes and annotation storage release. A
compiled oracle includes original aigDfs.c, aigUtil.c and timBox.c and confirms
structural XOR/AND levels 2/3, choice levels 1/2, and arrival-adjusted levels 6/7.

DAR's existing balanced eight-input XOR test assumed three structural levels;
corrected it to C's six levels for three XOR stages while retaining all 256
functional assignments. Final AIG, DAR core and DAR facade race suites pass.
This closes the XOR-cost finding; CO anchor/fanout reference semantics and the
remaining AIG families are still pending.

Full integration validation after the XOR timing repair completed successfully:
`go test -p 1 ./... -count=1 -timeout 180s` passes all 129 packages with tests;
three packages have no tests. No package-audit completion count is inferred
from this regression gate.

Next reference-access review should cover Aig_ObjRefs, Ref, Deref and ClearRef
against Go's internal output-anchor ownership. Read-only caller inspection
found reference mutations in CNF cuts and BBR, plus the previously migrated
reference-bounded fanout loops. The non-integer Aig_ObjFanout0 helper also
still scans by object ID rather than following the maintained head; it has no
production repository callers and needs its own regression before repair.

### Physical CO references and the raw fanout object accessor

Red regressions exposed synthetic manager-ownership references on explicit CO
objects: C reports zero fanouts, while Go reported one, allowed a dereference
of that ownership count, and restored it during reference resets. Removed the
synthetic count from creation, resetting, recounting and validation. Output
ownership remains in the manager's output vector and cleanup follows those
roots. Direct deletion of an attached CO still rejects misuse explicitly;
output-aware deletion removes it normally. Legacy raw output edges retain
their implicit consumer count, with a separate compatibility regression.

The non-integer fanout helper now uses the maintained circular-list head in
constant time rather than scanning objects by ID. Red tests cover reordering
the head, missing fanout allocation and C's unconditional zero-handle
conversion to object zero. A compiled fixture uses the original inline C
reference and fanout accessors to verify the counter sequence 0/1/0/0 and
nonempty/empty handle conversion. The ordinary counter helpers need no special
CO adjustment now that stored counts represent physical fanouts.

Focused regressions, full AIG race tests and the CNF, BBR, PDR and SSW race
suites pass. The full uncached module integration gate also passes:
`go test -p 1 ./... -count=1 -timeout 180s` completes with exit zero,
129 packages tested and three without tests. The log is
`/tmp/abcgo-output-refs-integration.log`. This is an AIG checkpoint, not
completion of its audit.

Read-only dominator review identifies the next source-contract questions:
the Go store keeps private traversal marks/reference counts, whereas C's
drivers and volume analysis mutate graph traversal IDs and marks. Existing
Go tests explicitly expect graph immutability, so their green result does not
prove source fidelity. Compare complete driver state and ordering against C
before blessing that design or replacing its tests. No dominator source
changes have been made in this checkpoint.

### Dominator driver ordering and graph work state

Subsequent source comparison and red tests exposed missing traversal-ID
updates in all three dominator drivers and reordered cuts when a fanout-list
head had been rotated. Restored the source reverse-DFS driver, LI-cone marking
for flop/PI modes, and reference-bounded fanout-list iteration. Removed the
separate stored traversal order and cone-activation pass. Direct drivers now
require fanout allocation, while the three source convenience wrappers own
and release their temporary indexes. Flop TFI rejects complemented roots,
regularizes recursive fanins and uses the shared bounded traversal helper.
Separate red tests cover the setup checks and traversal-generation boundary.

A compiled fixture includes original aigDoms.c and links original DFS,
fanout, memory and utility implementations. Its three-mode comparison checks
the complete ordered dominator vectors, statistics, manager traversal ID and
per-object traversal IDs after rotating a fanout head. Console-only stubs
provide the absent ABC frame/bridge environment, not analysis algorithms.

The source driver dereferences a null dominator vector on a dangling-fanout
fixture. Go previously treated that missing vector as empty and silently
returned incomplete cuts. Added a red regression and an explicit diagnostic
panic; valid-input exhaustive path tests now clean dangling logic first.
Tests that formerly asserted total graph immutability now check structural
and value preservation while allowing the source traversal mutations.

Further red regressions exposed diagnostic routines using private reference
and mark copies, ignoring preexisting graph marks, and failing to leave source
volume traversal state or clear source work marks. Removed those two private
arrays. Volume, LO collection and temporary cone references now use graph
state; FindGood clears graph MarkA flags. Restored the source negative-counter
check in reference recursion after observing its own failing test. The
compiled fixture additionally compares marked-boundary volume, LO marking,
temporary and restored reference counts, final traversal IDs and mark cleanup.
The existing related-node-cut barrier remains a tested Go safety extension;
it prevents double removal of cut inputs and is not attributed to C.

Focused regressions and the full AIG race suite pass. Repository search finds
no production callers of these dominator APIs outside aigDoms.go; no caller
packages needed changes. This checkpoint does not complete AIG review. The
remaining dominator representation/edge-case questions include legacy raw
output boundaries and the widened signature cache. Other AIG families and
the rest of the package inventory remain open.

### Dominator representation cleanup and factoring frontier state

Closed the raw-output question by enforcing C's explicit CO-object domain at
dominator-manager creation. The old synthetic boundaries were not connected
to the source fanout index and could not participate correctly in its driver.
A red regression demonstrates their previous silent acceptance. All public
Go output constructors create explicit COs; only legacy internal raw-edge
states are rejected. Removed the two adjacency snapshots and synthetic-ID
storage. Diagnostics read actual fanins; cofactoring counts follow the actual
reference-bounded fanout list. The path-enumeration test now builds its own
adjacency independently of production storage.

Kept the 64-bit signature cache as a defined, safe set-membership accelerator:
each bit represents an ID modulo 63, so its population count is a lower bound
on exact union size. It can reject only unions already over the limit. Exact
sorted node sets remain authoritative for merges and containment. Added
23,814 bounded-union comparisons over IDs 0,31,32,62,63,126, including signature
collisions and IDs beyond C's int-shift width. These pass with the compiled
driver comparison and full AIG race suite; no source undefined shifts are
used as an oracle for large IDs.

Started a fresh review of aigFact.go against aigFact.c rather than relying on
its historical review report. Red tests exposed support collection ignoring
already-current frontier nodes and reconstruction hiding the source's object
copy results in a local map. Collection now follows graph traversal marks,
removing an extra membership map; reconstruction publishes support, cofactor
and internal-node copies in the graph map while preserving unrelated entries.
A compiled fixture includes original aigFact.c and verifies collected order
and final marks. The full AIG race suite is the checkpoint gate. Other
factoring helpers, support-minimization representation/heuristic differences
and the remaining AIG families are still under review.

### Packed support-minimization truth tables

Replaced the one-bool-per-assignment representation with C-style packed uint32
tables and existing Kit truth primitives. The sixteen-variable boundary test
passed before the refactor and afterward: it checks the exact cofactor list
and reconstruction. Added direct packed-table checks for all 65,536
assignments at that boundary, plus phase-aware AND/XOR/CO truth checks for
two through six variables across the first word boundary. Existing random
care-equivalence, native XOR and progressive-cofactor regressions still pass.

Each sixteen-variable node table now occupies 2048 uint32 words (8 KiB),
rather than 65,536 bools (64 KiB). One reusable cofactor buffer replaces
per-accepted-variable allocations. Care construction, cofactor generation
and equality checks use Kit, following the source implementation more
closely. The regression-backed progressive update deliberately remains:
independently legal cofactors in the original C can jointly change the
function on its care set. Full AIG race tests pass. This compactness change
does not close the remaining factoring-helper/source-state review or the
overall AIG/package inventory audit.

### Factoring direct entry and manager traversal initialization

Direct recursive cone derivation on an unvisited, valid cone panicked when its
copy map had not been allocated by the wrapper. Observed that nil-map panic
with a red test, then initialized typed copy storage in the recursive entry.
A separate red test showed NewGraph started at generation zero, aliasing new
objects' zero traversal marks. The C-named Aig_ManStart already initialized
generation one; moved that initialization into NewGraph so both constructors
share it, removing the redundant assignment from Aig_ManStart. The first
constructor test covered only Aig_ManStart and passed; it was corrected to
cover both constructors and observed red before the shared repair.

Added a compiled fixture using original aigFact.c, object creation, AND,
hashing, memory, fanout and utility routines. It compares direct-recursive
results, copies/traversal marks, wrapper implication substitution and
reconstruction's support/internal copies. Its standalone manager uses real
fanout references and hash insertion; console stubs supply only the absent
frame/bridge environment. No analysis or gate implementation is stubbed.

Ten red entry-contract cases exposed acceptance of complemented roots,
constant overlap/derivation roots and unmarked/nonlogic support-cone entries
where C has assertions. Restored those checks; algebraic factoring now passes
the regular B root into derivation and reapplies the output phase afterward,
with a dedicated complemented-output regression. Final full AIG race tests
pass. The full uncached module gate after the shared constructor and copy-map
repair also passes (129 tested, three without tests, exit zero; log
`/tmp/abcgo-factoring-entry-integration.log`). The subsequent entry checks have
no external production callers and are covered by the final AIG race run.

Source inspection has now covered the factoring file's implication,
overlap/derivation, algebraic driver, packed support simulation,
reconstruction, support/frontier collection, highlighting and example paths.
Native XOR and buffer reconstruction preserve their actual operators rather
than C's AND-only assumptions; empty/inapplicable cones fail safely, and the
previously demonstrated progressive cofactor correctness repair remains.
Typed local simulation storage intentionally avoids publishing C's dangling
truth-table pData pointers after their allocation is freed. These differences
are explicit correctness/storage decisions, not replacement algorithms.

### Ternary simulation source-stream review

Source comparison found Aig_XsimRand2/3 used math/rand/v2, losing ABC's
deterministic shared random stream and its draw/rejection protocol. A red test
checks mixed binary/ternary draws and the following shared word. Restored
Aig_ManRandom draws and zero-rejection for ternary values; an independent
compiled original-C fixture confirms the entire mixed sequence and final word.

A second original-C fixture compares the full mapping and final object
simulation values for eighteen registers across two packed words: holding
zero, toggling, unknown PI-driven transitions, and transitions to constant
one. The targeted fixtures and source-stream regressions pass. Existing
ternary tests cover operators, state packing/ownership, cycles, retirement,
limits, native XOR/buffers and sequential reduction behavior. Remaining
review includes zero-register scratch-state behavior and the reduction
driver's source contracts; no completion claim for ternary simulation or AIG.

### Ternary scratch scope and reduction-driver review

Red regressions exposed two additional source-state mismatches: the
zero-register early return skipped constant/PI initialization and logic
simulation, while sequential simulation overwrote ordinary PO scratch data
that C leaves untouched. Removed the early return and separated logic plus
selected-CO simulation from the existing complete-frame helper. The driver
selects only LIs, following C's transfer scope; the complete-frame helper
still supports the existing explicit-output/XOR/buffer tests. One next-state
snapshot is now reused across rounds rather than allocated each round.

A compiled original-C fixture compares every object value and returned
mapping on zero-register and one-register graphs with seeded PO scratch.
It passes after the repairs. Reviewed the remaining aigTsim.c paths:
ternary operators, state allocation/hash/lookup/insertion, packing, printing,
counting and OR aggregation; bounded simulation/repetition/retirement;
constant mapping and the remap/cleanup/repeat driver. Go maps and immutable
keys replace allocator/hash links; native XOR, buffers and nonnumeric DAG
ordering are explicit operator-preserving extensions. The existing
no-progress check prevents an injected invalid engine looping forever.

Additional tests exercise two successive reductions followed by a nil
mapping, manager disposal/ownership, option forwarding and both assignments
of the remaining register. A separate nil-mapping test proves the unchanged
source manager remains intact. These passed without requiring a production
change. Full AIG, FRA and ABCI race suites pass; the final AIG race run also
includes those newly added driver tests. The log for the caller checks is
`/tmp/abcgo-ternary-scope-review.log`.

This completes the current source comparison of aigTsim.go itself, not the
shared multivalued engine or the complete AIG package. Review of
aigMvSim.go against saigSimMv.c is next, followed by the remaining AIG
families. Original package-audit completion remains 1/131.

### Multivalued simulation source review

Compared all of aigMvSim.go with saigSimMv.c: compact graph/manager creation,
symbolic literal/hash/AND behavior, frame transfer, state hashing and repeat
indices, history classification, oscillator search, retirement/skip masks,
equivalence mapping, diagnostics and symbolic/saturation driver boundaries.
Original unused level/accounting arrays remain omitted; typed values and Go
maps replace C's allocator-backed node and state tables. The simulator is
still the source symbolic algorithm, not a SAT or reachability substitute.

A red test exposed PI symbols allocated ahead of intervening gates. For a
gate followed by a later PI, Go assigned literals 8/6 instead of C's 6/8.
Restored allocation at the object's visit. The complete C fixture now includes
the original saigSimMv.c and memory implementation and compares every frame
value, symbolic count, repeat index, classification, skip mask, direct mapping
and full-driver mapping across symbolic cutover. Only console/bridge hooks
are stubbed; no simulation, hashing, history or mapping routine is replaced.

For compactness, manager setup now computes one dependency order, excluding
LOs from per-frame updates. It preserves numeric object order on source
topological graphs and resolves dependencies for Go's rewritten DAGs. The
frame loop reuses that order and one next-state buffer, eliminating per-frame
traversal flags and recursive walking. A forward-buffer regression passed
before and after the refactor and verifies correct known/undefined values and
unchanged graph fields. The graph/interface must remain fixed during manager
use, consistent with C's compact graph snapshot.

An independent original-C query fixture checks oscillator candidate periods
and state hashes over 4,367 histories: all binary prefixes through length ten,
three-symbol prefixes through length eight, and wider symbolic histories that
exercise uint32 overflow and sixteen-prime wraparound. These pass. The
postprocessor's actual-history classification is retained deliberately: C's
pRegsUndef counters are never populated. A focused output test checks its
defined-history count, equivalence groups, undefined exclusion and lasso marks.

Full AIG, SAIG and ABCI race suites pass after the cached-order refactor (log
`/tmp/abcgo-mv-frame-review.log`). A final AIG race run covers the expanded
hash cases as well. This closes the current multivalued-simulator review,
including its thin SAIG delegation boundary, without counting SAIG as a fully
audited package. AIG's canonicalization, retiming, visualization/property
splitting and other outstanding source families still need review; the full
original 131-package objective remains active.

The final driver coverage also includes a six-register delay chain with no
PIs and saturation at frames two/four. It converges only after aggressive
retirement, so the original-C comparison checks the complete state trace
through both thresholds and the final nil mapping. Undefined values in this
no-PI fixture can arise only from retirement. This test passed without a
production fix. The final full AIG race run includes it.

### Canonicalization and library recording source review

Compared aigCanon.go with the complete aigCanon.c: signature construction,
bytewise comparison, uniqueness, phase/permutation transformations and
diagnostics; truth hashing, ownership, insertion and resizing; manager
construction/disposal; BDC-to-AIG copying, DSD prime selection, statistics,
recording verification and library writing. The signature comparator
deliberately preserves little-endian memcmp of nOnes alone, rather than
numeric ordering or comparing the unused cofactor fields.

An original-C fixture compiles aigCanon.c and Kit truth operations unchanged.
It compares phases, uniqueness flags, permutations, full signatures and
canonical truth for 700 cases: zero-variable constants, every function of
one through three variables, and random normalized tables through twelve
variables, in both return-buffer modes. This baseline passed before and
after replacing duplicate cofactor/phase/swap code with Kit calls. Removed
the unused truth and reverse-permutation buffers (524 bytes per manager).
Verification still checks phase/permutation directly without mutating the
canonical truth back into the original, so no reverse buffer is needed.

Two focused tests were observed red: resizing a one-bin table chose five
rather than C's inclusive next prime three, and the empty-variable signature
diagnostic omitted C's leading zero. Reused util.Abc_PrimeCudd for manager
creation/resizing and restored that diagnostic. The duplicate prime-search
routine is gone. Existing Go safeguards remain intentional: lookup checks
stored variable width, resize reacquires the bucket pointer, and a failed
library write retains recording state for retry. C can read beyond a
different-width truth allocation and use its stale bucket pointer after
resizing. Existing ownership/resize/write-failure tests cover those Go
safeguards; this review does not reintroduce C's unsafe behavior.

Added an independent library semantic round trip: BDC decomposition and
AIGER serialization of constants, identity/inversion, four/six-variable
functions and twelve-variable parity. Evaluation of parsed AIGER literals
checks every one of the 4096 input assignments, including constant-polarity
convention and complemented outputs. Existing tests check fully/partially
DSD counters, prime recording, deduplication, graph consistency and failed
write retry. Full AIG race tests pass after the changes.

This closes the current aigCanon.go review, not its dependency packages or
the complete AIG package. Retiming and the other outstanding AIG families
remain next in the sequential audit; completed package count remains 1/131.

### Edge-retiming manager and driver checkpoint

Compared manager construction, object capacity/role fields, fanin/fanout
connection order, forward/backward movement, autonomous marking,
reconstruction and the main driver against aigRet.c. New red regressions
exposed graph-ID allocation instead of C's constant/PI/PO/LO/LI/logic groups,
missing construction capacities and PO flags, unchecked extra connections,
retained dangling logic, omitted final register sharing and manager cleanup
that retained the borrowed graph and latch objects.

Restored source object grouping and connection order, constructor Num/Temp
capacity checks, and the transition from Num capacity to retiming depth before
queue processing. Removed the two unused duplicate capacity fields. Manager
free now releases all Go-owned references while leaving its borrowed source
graph intact. Reconstruction runs the source cleanup pass, and the driver
finishes with Aig_ManReduceLaches. Depth zero and negative depth now retain
C's initial unary latch-transfer/sharing pass rather than returning early;
the limit applies to subsequently scheduled moves. Trace tests cover all
three depth cases and the regression proves duplicate registers are shared.
The zero-register Go convenience path remains a plain ordered copy.

A compiled fixture includes original aigRet.c and memory routines. It
compares every source-to-retiming object ID, PO flags, capacities, fanin
complements and ordered fanouts, both autonomous-marking passes, maximum
latch count and every edge value after moving two LOs and a gate with both
inputs complemented. Go reconstruction then passes sequential trace checks,
including the resulting ONE initialization's normalization to zero-reset Go
latches. No graph construction or retiming algorithm is stubbed; only
standalone console/bridge hooks are supplied.

A further red metadata regression found one-hot claims about old registers
surviving identity replacement. After actual moves the driver now clears
register-based/mixed claims and retains only groups over unchanged true PIs,
without changing the source annotations. Existing flop-number/representative
invalidation remains. This is a Go metadata correction, not a change to C's
retiming algorithm.

Full AIG and ABCI race suites pass after the manager/driver repair; final AIG
and ABCI race runs also pass after the one-hot fix. Logs are
`/tmp/abcgo-retiming-manager-review.log` and
`/tmp/abcgo-retiming-abci-final.log`. This is not completion of retiming:
packed/overflow latch-storage behavior, queue edge cases and backward VOID
initialization still need detailed source evidence. Also review the frontier
implementation against aigRetF.c and its metadata separately. The edge
manager retains typed local source mapping rather than publishing C's pData
pointers that become dangling when its manager is freed.

### Edge-retiming packed storage checkpoint

Reviewed the compact ten-latch representation, overflow two-bit words, carry
propagation, padding, and transitions in both directions against aigRet.c.
The new compiled `retiming_storage_compare.c` fixture calls the original
storage routines directly. All 3,200 operations match the Go implementation,
including each removed value and the complete contents of three interleaved
edges after every operation. Two deterministic grow/drain cycles cross six
word boundaries and repeatedly return to compact storage; 2,000 seeded mixed
operations exercise insertion/removal at both ends with ZERO, ONE and VOID.
The existing independent slice-oracle and explicit-boundary tests remain.

A new regression failed for both growth and compacting: the Go manager kept
every superseded overflow slice in an unused allocation history. Removed that
history, its cumulative counter and its allocation wrapper. Each edge now
owns only its current buffer. Unlike C's shared offset-based pool, these Go
buffers do not require obsolete allocations to remain live. The regression
keeps both manager and edge alive and checks a weak pointer to a retired
28-byte backing array after GC, avoiding tiny-allocation batching. Both cases
pass across 20 repetitions after the repair. Manager-free ownership checks
now inspect its actual graph, object and mapping references.

Full AIG and ABCI race suites pass after this repair (8.338s and 29.054s),
recorded in `/tmp/abcgo-retiming-storage-review.log`. Separately, the fresh
uncached module check prompted by log.red passed all 129 tested packages
and three packages without tests; its log is
`/tmp/abcgo-log-red-verification.log`.

This completes the current packed-storage review, not the retiming family.
Queue/autonomous corner cases, backward VOID initialization and the separate
frontier implementation still need the source review described above.

### Edge-retiming complete driver and reset checkpoint

Added a compiled comparison that executes the original Rtm_ManRetime,
Rtm_ManToAig, register sharing, graph construction, cleanup and their real C
dependencies. Only standalone console/bridge hooks are supplied. Compared
register counts and 64-frame output traces for 26 graphs, both directions,
and depth limits -1, 0, 1, 2 and 5: all 260 cases match. Directed graphs cover
register chains, complemented inputs, observable autonomous feedback and
unobservable PI-driven state; 20 seeded random graphs add reconvergence,
constants, multiple outputs and feedback. This is 16,640 matched output
frames, not an exhaustive equivalence proof or an internal queue trace.

Inspection of source scheduling confirms its ordered queue and depth handling
in Go. Go filters initially queued autonomous objects when popping instead
of when enqueuing. This does not affect movement order among eligible objects:
forward reachability propagates through all fanouts and backward reachability
through all fanins, so an excluded autonomous object cannot become an eligible
neighbor of a moved object. Existing Go XOR/buffer extensions are outside this
AND-only compiled comparison.

The directed reset example confirms a source limitation rather than a port
bug: backward retiming discards initialization, inserts VOID edge values and
reconstructs them as zero without solving reset constraints. A zero-initialized
register driven by AND(!a,!b) can therefore produce initial output one after a
backward move, where the original output was zero. Added an explicit test and
expanded the public function comment so this is not presented as preservation
of zero-reset traces. No replacement reset algorithm was invented.

Five red regressions exposed incomplete Go clock-domain validation: an empty,
partial, duplicate, negative or out-of-range single-domain list was silently
replaced with membership covering all new registers. The existing single-domain
restriction now requires every old register exactly once when metadata is
present. A positive test accepts reversed full membership, checks rebuilt
membership, verifies source metadata remains unchanged and compares forward
traces. This is a repair to the Go metadata extension; C's retiming algorithm
has not changed. Low-level reconstruction metadata after manual manager edits
and the separate frontier implementation still need review before closing
the retiming family.

Validation after the clock-domain fix: full AIG and ABCI race suites pass
(9.838s and 29.439s), recorded in
`/tmp/abcgo-retiming-driver-review.log`.

### Direct edge-retiming reconstruction metadata

A new red test exposed stale register metadata when callers use the low-level
manager API: after moving two LOs and their AND, Rtm_ManToAig returned one
register with two old flop IDs, a two-register domain and obsolete one-hot
claims. The high-level driver's cleanup did not protect this path.

Moved that cleanup into reconstruction, shared by direct and driver callers.
Original identities are retained only when the reconstructed count matches
the source count and exactly one zero-encoded latch remains on each original
LO edge. The already computed total makes a second scan of all edges
unnecessary. ONE normalization changes register encoding and invalidates old
IDs; VOID, which C reconstructs as zero, has unchanged encoding when the
layout is otherwise identical. Tests cover ZERO/ONE/VOID explicitly.

When identity changes, reconstruction clears old flop-number/representative
and register one-hot claims, retains true-PI-only groups, and rebuilds complete
single-domain membership. Clock-domain validation is shared with the driver.
An unchanged layout preserves its existing metadata, including correspondence;
the original graph's metadata remains untouched. The driver still records
that moves occurred even if a sequence restores the original layout. New
direct-path tests verify metadata and sequential traces both with and without
movement; all 260 complete-driver C comparisons remain green.

Full AIG and ABCI race suites pass (9.944s and 28.641s) in
`/tmp/abcgo-retiming-reconstruction-review.log`. The final AIG race run after
reusing the existing register total and adding initial-encoding cases also
passes (9.994s), in `/tmp/abcgo-retiming-reconstruction-final.log`.

### Frontier retiming source review

Reviewed all of aigRetF.go against aigRetF.c: CO-scoped fixed-point marking,
register links, dynamic object scanning, buffer bypass, complemented-pin reset
normalization, move limits, scratch cleanup, sequential pruning and final
rehashing. The local typed register/domain maps replace C's temporary pNext
links; the existing same-domain restriction and explicit XOR support are Go
extensions. Old registers that survive frontier pruning retain their identities,
so its projected one-hot groups can remain valid; this differs from moving
latches among arbitrary edges in the other retiming implementation.

The new source comparison exposed a port mismatch before repair: Go treated
zero as a no-op and rejected negative limits, while C checks the limit after
the first eligible move. Restored that source behavior by using a minimum
effective limit of one. Updated the old no-op test to exercise an actual
ineligible cross-domain pair rather than asserting the invented zero-limit
behavior. Both the compiled zero-limit regression and a focused negative-limit
regression failed before the change. Existing polarity/XOR/feedback/domain
tests continue to pass.

The C frontier routine's final Aig_ManDupOrdered call crashes even on the
two-register AND example: converting the old gate to a buffer introduces a
forward reference to its new LO, violating that copier's increasing-ID
topology assumption. An AddressSanitizer run traced the null fanin to
Aig_ObjLevelNew through Aig_ObjCreateCo in Aig_ManDupOrdered. The source C
algorithms remain unchanged. The frontier fixture captures the source graph
immediately before this final copy and evaluates it recursively, including
buffer/forward references. It executes real source marking, movement and
sequential cleanup; it does not validate C's broken final copier. Go already
uses its reviewed recursive duplicator, which handles these references.

All 130 frontier cases (26 directed/seeded graphs at limits -1, 0, 1, 2 and 5)
match the captured C register counts and 64-frame traces: 8,320 output frames.
The 260 complete edge-driver cases still pass with the shared recursive C
evaluator. A separate fixture executes original marking twice on each of three
graphs and matches every object's MarkB and TravID plus the manager generation:
observable autonomous feedback, a PI-fed reverse register chain, and no true
PIs. It checks that C clears its temporary latch links; Go preserves unrelated
object payloads. Dangling logic remains outside CO-based marking in both.

Independent sequential tests cover all eight combinations of equal/opposite
LI drivers and the two input complement flags, making new logic simplify to
a literal or constant. Every case moves the gate, preserves zero-reset traces
and leaves no live scratch marks. Constants are handled through regular Go
edges rather than writing through C's potentially complemented pointers.

Full AIG and ABCI race suites pass after the limit repair (11.537s and 28.582s)
in `/tmp/abcgo-frontier-step-review.log`. The added marking and simplified-result
tests have their own focused pass; the final full AIG race suite passes in
11.728s, recorded in `/tmp/abcgo-frontier-final-review.log`.
This closes the current edge/frontier retiming family review. The AIG package
audit is still open, including its property-splitting implementation and the
remaining unreviewed families; completed package count remains 1/131.

### Property splitting: restore the source BDD pipeline

The existing aigSplit.go explicitly replaced CUDD with recursive Shannon
cofactoring into an AIG decision graph. Removed that substitute and restored
the source calls through bdd/cudd/core: Cudd_Init, symmetric automatic-reordering
configuration, recursive cached BDD construction, explicit symmetric sifting,
MSB-first cubes, Cudd_Cofactor, conjunction with each cube, and memoized
BDD-to-MUX reconstruction. Complemented else edges and output roots retain
their phase. Zero cofactors are omitted in source assignment order; remaining
COs, including LIs, are copied afterward, followed by cleanup and register-count
finalization. The existing explicit XOR/buffer support remains a Go extension.
The unrelated interpolation cone copier now has its own private graph-copy
helper instead of calling the misnamed former BDD helper.

Two regressions failed before the repair: constructing inputs without a CI
indexing pass made both selected variables map to the first BDD variable and
panic, and subset selection used math/rand rather than Aig_ManRandom. Variable
mapping now follows original CI order, independently of CioID scratch state.
Subset selection uses the source stream and Vec_PtrRemove semantics (last equal
entry, survivor order retained), without mutating the caller's slice. Typed
local BDD scratch keeps BDD handles separate from graph-copy maps and unrelated
object payloads. The returned BDD root and every cofactor have explicit owned
references, released before quitting the manager in the high-level driver.

The exhaustive ownership test then exposed a dependency bug, isolated by a
separate new red test in bdd/cudd/core: after symmetric reordering of truth
table 5, releasing the sole external root still left one root referenced.
reorderCommit rewired its canonical graph but retained old DAG reference
counts, including on reused internal handles. It now rebuilds counts from
external ownership plus new incoming edges, retaining unchanged ZDD child
ownership and saturated counts, and resets child-release bookkeeping. The
255-function ownership regression passes, as do full core/public CUDD race
suites and a mixed BDD/ZDD release check. This necessary dependency repair is
not a completed audit of the CUDD package.

The new fixture compiles original aigSplit.c, the actual CUDD implementations
and their AIG, extra, ST, MTR, timing and memory dependencies. Only standalone
console/bridge hooks are supplied. All 256 three-input functions and eight
variable subsets (reverse index order) are tested in combinational and
sequential form: 4,096 pipeline cases. Independent assignment checks verify
13,824 cofactor functions across 110,592 evaluations, nonzero output order,
MUX reconstruction, unchanged latch-input functions and zero leaked BDD roots
after releasing each function's results. C performs its own zero-reference
check too. A further 508 complete Aig_ManSplit calls compare random selection
and ordered property/state outputs with C, and independently check that the
disjunction of split outputs equals the original property. This validates
small-function semantics and ownership, not large-instance scaling or every
automatic-reordering trigger in the dependency.

The focused native comparison passes after both repairs. Full CUDD core/public
race suites pass (26.201s/1.045s), recorded in
`/tmp/abcgo-cudd-reorder-ownership.log`. Full AIG, ABCI and proof/aiginterp race
checks pass (20.108s/28.373s/3.698s) in `/tmp/abcgo-split-bdd-review.log`;
the final AIG race run with all expanded cases passes in 24.167s, recorded in
`/tmp/abcgo-split-bdd-final.log`. The fresh uncached module check exits zero:
129 packages tested, three without tests, no failures, in
`/tmp/abcgo-split-bdd-integration.log`.

This closes the current aigSplit.go review. Other AIG families remain in
progress; the audit still has only the root package completed (1/131). Packing,
justification, visualization and printing remain to reconcile with the source
inventory. Initial packing inspection found another math/rand-based initializer
where aigPack.c consumes two Aig_ManRandom words per CI; test that discrepancy
before changing it in the next family review.

### Packing source review

Reviewed aigPack.go against every routine in aigPack.c: allocation, word
helpers, random initialization, simulation, care counting, histogram/candidate
selection, direct/automatic insertion, diagnostics and lifecycle wrappers.
The source uses two Aig_ManRandom draws per CI, high word first, then shifts
left to reserve lane zero. A new regression failed for both uncared and fully
cared words, including the subsequent random-stream position. Removed
math/rand and restored exact source initialization. SetRandom overwrites each
complete input word even when its care mask is populated; the mask itself is
unchanged. The separately named Go SetInputWord helper still preserves packed
cares. Updated the old random-reseeding test that asserted the invented
care-preserving behavior to exercise that explicit manual helper instead.

A compiled fixture includes original aigPack.c and its real AIG random
implementation. It compares all input/care words, every object's signature,
care counts, insertion results, total/skipped counters and ordered candidate
IDs after 1,166 operations (1,167 snapshots including initialization).
Operations include 128 complete assignments exhausting the 63 automatic lanes,
36 direct boundary/empty/duplicate-pattern cases, 1,000 seeded partial cubes
and two random resets after packing. Lanes 0, 1, 31, 32, 62 and 63 are exercised
explicitly. Seven gates cover every complemented AND-pin combination and
reconvergence; outputs include both constants and a
complemented gate. Final histogram and free-tally text match C as well.

The comparison also failed before correcting the source-named PrintCare
wrapper: its C printing is commented out, while Go emitted one line per CI.
Restored that disabled diagnostic; explicit WriteCare remains available as a
Go writer-based extension. Existing empty-graph guards retain finite zero
statistics rather than C's zero-denominator NaNs. They are outside the native
comparison, which has nonzero input/gate counts.

Retained the existing atomic rejection of contradictory duplicate literals:
the source checks its original lane before a separate write loop and can
accept an unsatisfiable cube. Native comparisons use consistent cubes;
independent tests check the Go safeguard and preservation of words/care masks
on rejection. Existing all-64-lane evaluation tests cover explicit XOR,
buffers and rewritten graph IDs, beyond the source's AND/topological-order
simulation. Snapshot ownership and invalid interface/literal/lane guards also
remain covered. Compactness review retained direct slice storage and native
bit operations; unused C cursor/verbosity/repeat fields are already absent.

The focused packing tests and compiled comparison pass. Full AIG and ABCI
race suites pass in 24.361s and 29.114s, recorded in
`/tmp/abcgo-packing-source-review.log`.
This closes the current packing family review, not the AIG package. The
justification/experimental driver and visualization/printing families remain
next; the complete-package count is still 1/131.

### Justification core source checkpoint

Compared aigJust.go with the low-level operators, eight-attempt greedy search
and ternary simulation in aigJust.c. Four new regressions failed before
repairs: the public output wrapper marked CO scratch instead of transparently
delegating to its driver; unsuccessful searches discarded C's last partial
support; justification's duplicate traversal counter bypassed the shared
bounded rollover; and stale/invalid CioIDs could produce a literal naming the
wrong CI while returning success.

Restored source CO delegation and retention of the last failed attempt's
partial support. Failure remains a failed greedy attempt, not an UNSAT proof,
and that partial support is not a satisfying certificate. Removed the private
counter implementation and use Aig_ManIncrementTravId for search and ternary
simulation. Rollover tests now use the shared generation convention, including
the (1<<30)-1 boundary and the old maximum-int case. This preserves the already
reviewed safe Go rollover rather than C's ineffective high-counter pData clear.
When collecting a witness, each CI index must identify that same input in the
current interface. Invalid, negative and out-of-range indices now panic instead
of silently producing a wrong witness. Callers still explicitly initialize CI
scratch indices with Aig_ManSetCioIds; the port does not reindex them implicitly.

The new fixture executes original justification and ternary routines with the
source traversal helper. All 4,310 queries on ten directed/seeded four-input
graphs match results, support literal order, manager generation and every
object's MarkA/MarkB/TravID state. Queries cover all eight heuristics, direct
recursion and the eight-attempt wrapper, failed attempts, output inversion,
reconvergence, and all 81 partial assignments in mixed search/simulation order.
Independent Boolean evaluation checks all assignments consistent with each
successful output witness or known ternary output. Unknown ternary values do
not assert a complete Boolean classification.

Retained the existing Go correction for impossible constant targets: C's
recursive search returns true for constant one regardless of requested value.
The native comparison excludes those invalid source certificates; independent
Go tests require failure for them. Explicit XOR/buffer/complemented-Edge support
and conflicting-partial-assignment guards remain Go extensions covered by the
existing tests. The bounded greedy algorithm has not been replaced by a SAT
solver or a completeness claim.

The focused regressions and native comparison pass. Full AIG and ABCI race
suites pass in 24.255s and 28.353s, in
`/tmp/abcgo-justification-source-review.log`. The expanded full AIG race suite
with the 4,310-query fixture passes in 25.300s, in
`/tmp/abcgo-justification-final-review.log`.

### Justification experiment driver source review

Read aigJust_experiment.go against the original driver: initial packing and
candidate order, opposite-phase target choice, separate target/failure counts,
support-length accumulation, pattern insertion, summary, mark cleanup and final
packing simulation/free. No additional driver bug was found. The Go driver
uses elapsed wall time and guards empty denominators; C uses CPU time and can
print NaNs. These diagnostic differences do not replace the source search.

Extended the native fixture to execute the actual C experiment and packing
implementations on a twelve-input graph with positive/negative AND chains and
a logically impossible reconvergent cone. Initial and final 33-bin histograms,
all summary fields except elapsed time, final pattern tally, manager generation
and every object's final scratch state match C. The impossible cone explicitly
exercises one failed target-one justification. The borrowed Go graph remains
valid and all live scratch flags are cleared. Focused race tests pass for both
the 4,310-query comparison and the complete experiment after this addition.

This closes the current justification core/experiment family review. The AIG
package audit remains open, including visualization/printing and the remaining
source inventory; the completed-package count is still 1/131.

### Visualization source review

Read all of aigShow.c and aigShow.go, including the filename helper used by
the C show wrapper. Three new regressions failed before repair: terminal labels
used CI/CO scratch indices (new inputs all appeared as PI 0) instead of stable
object IDs, unused constant one appeared as an extra node, and Show discarded
the source directory and appended .dot after an existing extension. Labels now
use object IDs, constant one is declared only when referenced, and Show retains
the directory while replacing the final filename extension. filepath.Ext also
avoids mistaking a dotted parent directory for the filename's extension.

Restored the source DOT layout: title and statistics boxes, level scaffolding,
CO ordering/invisible edges, terminal colors and shapes, numeric labels,
internal-node fill highlighting, and bold/dotted signal edges. Shared node/rank
emitters keep the formatting code smaller than the C fprintf sequence. History
choice-edge rendering remains disabled, as in the active C code. A compiled
fixture executes the actual C writer, and eight standard diagrams match byte
for byte across referenced/unreferenced constants, highlights and both history
flag values. The fixture also checks C's post-write marks and output levels.

The Go API retains local highlight state and display ranks rather than mutating
the source graph. A full-node snapshot test covers MarkA/MarkB, traversal IDs,
and a deliberately nonstandard cached CO level. Source C temporarily rewrites
CO levels and clears highlighted marks, which can destroy preexisting scratch
state; Go leaves all of it intact. Named graph labels retain escaped text as a
Go extension. XOR/buffer shapes, explicit zero-level logic, and ranked high
dangling gates remain supported instead of leaving those objects implicit.

Graphviz successfully parses a diagram containing quoted/newline/Unicode graph
names, a buffer, XOR, complemented edges, highlights and a higher dangling
cone. Boundary tests allow 200 logic nodes, reject 201 (including a dangling
gate), preserve existing files on invalid-highlight/size errors and report
directory-write failures. The show wrapper returns the DOT artifact path;
external GUI rendering remains the caller's responsibility. No GUI command is
launched by the library.

Full AIG and ABCI race suites pass in 25.821s and 28.557s, recorded in
`/tmp/abcgo-visualization-source-review.log`. This closes the current
visualization review, not the entire AIG package. Printing and the remaining
source inventory still need verification before marking package 2 complete;
completed-package count remains 1/131.

### Expression and object printing source review

Read the expression, verbose-object/cone/manager, ordinary-object and control
fanout printers against aigUtil.c and aigObj.c. Regressions failed before
repair for root expression parentheses, expanded XOR recognition, distinct
unnamed inputs, verbose traversal scope, ordinary object formatting, null
objects and the control printer's use of the wrong diagnostic routine.
Expression printing now follows source nesting and XOR-before-MUX recognition;
unnamed inputs use current interface positions without changing CioID scratch
state. A further red regression exposed collisions with supplied names. All
supplied names are now reserved before allocating unique fallback names.

Ordinary diagnostics retain complemented-root prefixes, source spacing,
reference counts and choice-class annotations, including constant-one's valid
zero-valued handle. Null prints without accessing a graph. A shared complement
suffix helper removes duplicate formatting logic. Verbose cone/manager
printing uses the source DFS entry points, restoring the constant prelude,
terminal coverage, ordering and traversal marks. Go prints stable CI object IDs
in the manager header instead of C process addresses. Explicit XOR/buffer
diagnostics remain typed Go extensions; physical XOR expressions preserve
repeated operands rather than applying source supergate deduplication, which
would change parity.

The compiled fixture executes the original printers/DFS on all 256
three-input truth functions, both expression/object polarities, complete
manager dumps, cones, null objects and control fanouts: 29,704 queries match,
including traversal generation and every object's traversal mark. Only C's
pointer-valued PI header is normalized; the Go header's ordered IDs are checked
separately. Parsed Eqn expressions independently reproduce all 2,048 truth
assignments. Additional direct tests cover choice annotations, supplied/default
name collisions, and repeated physical XOR parity through complemented outputs.

Full AIG and ABCI race suites pass in 33.424s and 28.894s, recorded in
`/tmp/abcgo-printing-current-review.log`; the subsequently added XOR parity
test passes independently. This closes these printer entry points, not the
complete AIG audit. Full BLIF/Verilog dump writers and the final production-file
inventory still require reconciliation before package 2 can be marked complete.

### BLIF generated-net naming checkpoint

Read the full BLIF/Verilog dump writers against aigUtil.c. Both use local
name maps and deterministic graph order instead of overwriting source graph
scratch fields; typed XORs/buffers and zero-initialized register transitions
remain explicit. Existing Boolean tests cover default-name BLIF truth, but did
not exercise caller-supplied names against internal generated names.

Added an independent-reader regression before repair. Input names n0, n3 and
n4 each caused ParseBLIF to reject the export for a duplicate driver (constant,
internal gate or default output). The dump name allocator now reserves all
supplied input/output names before generating any defaults. A shared unique-name
helper also serves expression printing, removing duplicate suffix logic. This
keeps normal names unchanged while preserving supplied port names when a
collision occurs. C likewise assumes its generated names do not collide; the
Go writer now avoids that invalid-output case.

Six round trips cover constant/gate/output collisions, an output name matching
a default input, and an already reserved numeric suffix. The independent BLIF
reader checks the exported netlist, then its evaluator checks all four input
assignments for each case. Full writer validation remains open, including
sequential export and native-source semantic comparisons; this checkpoint does
not mark the AIG package complete.

After the naming repair, full AIG and ABCI race suites pass in 33.014s and
29.004s (`/tmp/abcgo-dump-names-review.log`). Completed-package count remains
1/131; aig/aig is still the active sequential review.

### Full dump-writer semantic review

Linked the original C BLIF/Verilog writers and DFS into a graph-only fixture;
there are no substitute writer algorithms. Built all 256 three-input Boolean
functions with zero through three registers. Additional direct, complemented
and constant outputs expose every logic node to satisfy C's DFS precondition.
Read both C and Go exports with the production independent BLIF/Verilog readers:
4,096 circuits (1,024 graphs x two implementations x two formats). At every
input/state assignment, all POs and register-input values match the original
Go graph; dimensions, zero initialization and rising-clock-edge declarations
are checked. This proves the one-step transition relation and reset values,
not an independent external HDL simulator's interpretation. The C Verilog
writer emits redundant identical LI assignments; these are harmless in these
readbacks, while Go emits each assignment once.

Additional round trips cover explicit repeated XOR operands, inverted buffers,
zero through two registers and supplied sequential terminal names. Snapshot
checks prove writes preserve node fields (including nonstandard CioIDs/marks),
manager generation, payloads and copy bindings. The basename helper also has
portable mixed-separator/dotted-parent tests. Its local, filename-only extension
handling intentionally avoids the original static-buffer helper's dotted-parent
and mixed-separator quirks. No dump algorithm needed replacement or repair in
this checkpoint beyond the already committed name-collision fix.

Full AIG/ABCI race suites pass in 55.776s/28.930s, recorded in
`/tmp/abcgo-native-dumps-review.log`. Latest dump-only race checks additionally
cover the typed-node, scratch, supplied sequential-name and filename tests
(`/tmp/abcgo-native-dumps-final.log`). Review of the adjacent manager routine
then found a distinct first-PO flip representation bug; its new regression is
red and the repair is the next checkpoint. AIG/package-wide completion remains
unproven; completed-package count is still 1/131.

### First-PO flip source repair

Inspection of Aig_ManFlipFirstPo found that Go complemented the CO handle in
the interface slice, whereas C's Aig_ObjChild0Flip complements the CO's child.
A red structural regression showed terminal identity changed and the child
stayed unchanged. Independent BLIF and Verilog round trips also failed at
assignment zero: exported false while Graph.Evaluate correctly observed true.
The manager routine now flips the explicit CO's first fanin. Legacy implicit
raw outputs retain their previous edge-flip handling. No references, terminal
indices or other cached fields are recomputed: C likewise changes only the pin.

Tests check stable handles, unchanged other output/reference ownership, every
assignment and exact double-flip restoration. A compiled fixture executes the
original C routine for both constant/CI/AND polarities: all sixteen flips match.
Focused race checks pass (`/tmp/abcgo-first-po-flip-final.log`). Full AIG, FRA
and interpolation race checks pass in 56.143s, 12.228s and 11.731s; these latter
two packages are the direct first-PO flip callers. The ABCI integration check
is recorded in the same `/tmp/abcgo-first-po-flip-review.log`.

The adjacent statistics/live-object predicate review has newly red source
regressions and remains the next repair checkpoint. These do not change the
first-PO fix or establish that the whole AIG package is complete.

### Manager statistics and live-object predicate source repair

Red regressions exposed two more source mismatches. Aig_ManPrintStats used
cached CO levels (e.g. 99) instead of actual driver depth (1), always emitted
lat=0 and omitted live XOR/choice/buffer fields. It now delegates depth to the
existing Aig_ManLevels helper and uses the source conditional fields. This
removes the duplicate depth scan and preserves object-type counts and source
formatting. Aig_ManReportImprovement already matches the source count/reduction
formulas, including zero denominator handling; no repair was needed there.

Aig_ManHasNoGaps additionally required an empty deleted-slot map, but C's
predicate compares live object counts. The regression failed after deleting
one dangling AND despite the live-object equation holding. Removed that extra
condition and documented the actual contract. A separate buffer case remains
false, since buffers are excluded from the right side of C's equation. A search
found no production Go callers beyond the entry point itself.

A compiled fixture checks byte-identical statistics and predicate results for
ordinary graphs, a deletion hole, sequential dimensions and explicit XOR,
buffer and a valid equivalent choice alternative; it also has a deliberately
stale CO level. Focused source regressions pass under the race detector
(`/tmp/abcgo-manager-stats-review.log`). Full AIG tests on this state pass in
30.759s, with the ABCI result recorded alongside in
`/tmp/abcgo-manager-current-review.log`. The final equivalent-choice test
fixture refinement passes separately. The preceding output-flip repair also
passes the full ABCI race suite (28.718s).

The remaining manager/cleanup and file-by-file inventory need reconciliation
before closing aig/aig. In particular, recheck cleanup's typed buffer handling
against C rather than assuming the earlier ordinary-AND tests cover it. The
full sequential goal remains active, with 1/131 packages completed.

### Reference-driven manager cleanup source repair

Read Aig_ManCleanup, AntiCleanup, CiCleanup and CoCleanup against aigMan.c,
including recursive pin disconnection/deletion in aigObj.c. The manager cleanup
was an output-reachability sweep instead of C's zero-reference AND/XOR-root
collection and recursive reference release. Two minimal regressions failed:
a standalone buffer plus its gate was removed (reported two instead of zero),
and a dead AND cone containing a buffer reported three instead of two gates.

Restored the source primitive using the existing Aig_ObjDelete_rec helper.
Initial roots are collected before mutation; recursive deletion releases each
physical pin, removes hash entries, and deletes buffers only when encountered
inside a dead gate cone. The return value counts AND/XOR gates. A temporary
reference hold now preserves its node as in C. Fanout rebuilds are batched
around the entire operation instead of rebuilding after each recursive delete.
The source primitive does not regard auxiliary choice links as references and
does not rewrite caller-owned choice tables; callers must rebuild weak entries
before subsequent choice traversal. Directed tests and native choice cases
verify that boundary without inventing a choice-preserving cleanup algorithm.

The compiled fixture uses production C allocation, pin connections, hash tables,
fanouts and recursive deletion. Across 512 graphs and 1,024 cleanup calls,
removed counts, remaining gates/hash entries, every object's liveness/fanins/
references, and circular physical fanout lists match. Cases include ANDs,
physical XORs, complemented pins, standalone/nested buffers, equivalent choice
alternatives, a held reference and release followed by a second cleanup. Graph
invariants pass after releasing the temporary hold. Focused race tests pass in
2.385s (`/tmp/abcgo-cleanup-source-final.log`). Full AIG, ABCI and SAIG race
suites pass in 58.246s/29.134s/5.903s
(`/tmp/abcgo-cleanup-source-review.log`), before the final choice-only fixture
extension; that extension is included in the focused check.

The previous broader sweep remains confined to the existing sequential
register-removal helper. Moved its body into sclCleanupLogic and removed the
unused storage-removal tally while preserving its existing behavior. This is
an explicitly OPEN fidelity issue, not a source-equivalent wrapper: the earlier
sequential buffer test established the Go behavior without comparing that
boundary against C. Next extend the native oracle to SeqCleanup/Basic, checking
terminal disconnection, retained buffers, CioID refresh and the Basic patch's
immediate deletion/return-count boundary. Recheck Aig_ManExtractMiter's source
phase normalization afterward. The package-wide audit remains open at 1/131.

### Local CO patch source repair

While restoring Basic sequential cleanup's call to Aig_ObjPatchFanin0,
inspection found that the patch helper recomputed every graph level/phase and
rebuilt all fanouts. C updates only the target CO's cache and pin. A new red
regression patched one output while another had a valid auxiliary level of 99;
Go incorrectly changed the unrelated output's level to 1.

The helper now removes/adds its exact physical fanout pin, updates only the
patched CO's driver, level and phase, and keeps the existing recursive deletion
of an unreferenced old cone. The regression checks the untouched CO snapshot,
patched cache, active fanout consistency and graph validity. Full AIG, ABCI and
SAIG race checks pass in 59.416s/29.039s/5.858s
(`/tmp/abcgo-seq-cleanup-review.log`). The sequential wrapper repairs and their
expanded state oracle remain a separate open checkpoint; this commit records
only the independently justified local patch repair.

### Sequential cleanup wrapper source repair

The expanded review confirmed that the earlier sequential sweep had not
preserved C's primitive boundaries. Four new regressions failed before repair:
standalone buffers were reclaimed, Basic counted gates that C's patch already
deleted, full cleanup retained old terminal IDs after removing an earlier
register, and Basic left LO pNext/CioId scratch intact. The original C boundary
fixture confirms each expected result. Corrected the earlier standalone-buffer
test, which had encoded the broader Go behavior rather than checking C.

Both wrappers now finish through reference-driven Aig_ManCleanup. Full cleanup
disconnects removed LIs, keeps removed LOs addressable until pin release is
finished, restores current terminal IDs and then retires those LO slots. This
models C's still-addressable arena objects without retaining invalid Go edges.
Basic patches unused LIs through Aig_ObjPatchFanin0, preserving its immediate
old-cone deletion and returning only the final cleanup's count. LO union scratch
is cleared after register reachability, as in C. Removed the duplicate general
reachability sweep. Existing clock/one-hot/flop metadata remapping stays intact.

A native fixture also exposes an upstream corner case: C retains a standalone
buffer even when that buffer references an LO whose vObjs slot was removed.
Go deliberately repairs only these invalid-input buffer cones. Independent
standalone buffers survive, matching the ordinary C boundary. New red/green
coverage checks selective reclamation; shared/nested AND/XOR/buffer cones with
and without fanouts pass Graph.Check and independent sequential trace
comparison. The small dependency memo uses one map and only runs when removed
inputs still have physical consumers. This is a narrow upstream safety repair,
not a substitute cleanup algorithm.

The complete state oracle then failed on a circular fanout ordering difference:
post-deletion rebuilds sorted surviving pins instead of preserving C's removal/
append order after patching a CO. Per-pin disconnection already maintains the
index, so recursive deletion, manager cleanup and Basic no longer perform those
redundant rebuilds. A minimal regression and C fixture confirm the constant
fanout order 12,16,14 (not sorted 12,14,16).

Across 512 generated sequential graphs and 1,024 calls, actual C and Go match
removed-gate counts, register/interface lists, flop-number remapping, explicit
true PI/PO metadata, manager generation, hash counts, live-object references,
fanins, cached levels, traversal IDs, CI/CO IDs and physical fanout cycles.
Cases cover zero through three registers, full/Basic modes, active/inactive
fanout indices, safe standalone buffers and physical XORs. Deliberately stale
terminal IDs, CO levels of 99 and stale true-count metadata check exact scratch
scope. Repeated cleanup checks idempotence without hiding source side effects.
The invalid-input C corner is tested separately rather than requiring Go to
reproduce an inconsistent graph.

Focused race checks pass in 4.937s
(`/tmp/abcgo-seq-cleanup-focused-final.log`); the final small boundary/order
fixture passes in 1.052s. Full AIG and ABCI race suites pass in
60.275s/28.775s; SAIG verification is recorded in
`/tmp/abcgo-seq-cleanup-final-review.log`. This closes the reopened sequential
cleanup wrapper boundaries. Remaining AIG work includes the manager miter's
phase-normalization contract and the final file/function inventory. Completed
package count remains 1/131; the overall sequential audit is active.

### Sequential cleanup repository integration

A fresh `go test -p 1 ./... -count=1 -timeout 180s` after commit ece8b8faa
passes all 129 packages with tests; three command packages have no tests
(`/tmp/abcgo-seq-cleanup-integration.log`). The original 131-package checklist
predates the proof/aiginterp driver migration, which adds one integration
package. These green checks establish integration of the cleanup repairs, not
completion of the remaining package audits. The full SAIG race check above
also passed in 5.901s.

### Miter phase and copied-cone ownership source repair

Independent probes of Aig_ManExtractMiter confirmed three source mismatches.
For regular roots !a&!b and a&b, C normalizes the XOR's all-zero phase, yielding
truth nibble 6; Go emitted nibble 9. C creates fresh zero-level CIs, while Go
reused StartFrom's copied CI levels (3/7 in the regression). Comparing a root
with itself also ran Go-only cleanup, deleting the copied root even though the
source's copy map still named it. Source C retains that one unused copied gate.
All three regression tests, including the compiled-C comparison, failed before
repair (`/tmp/abcgo-miter-red.log`).

The constructor now follows C's local creation flow: fresh interface, fresh
copy bindings, both recursive cone copies, XOR, zero-input phase normalization
and CO creation. It retains copied cones and does not call cleanup. StartFrom
keeps its separate source contract of copying CI levels. The miter's public
comment explains phase normalization and copy lifetime so callers do not assume
it is an unnormalized Boolean mismatch output.

The compiled source comparison checks output truth, retained gate counts,
fresh CI levels and copy validity for opposite-phase and identical roots.
Independent exhaustive checks cover 64 pairs of constants, inputs, each signed
AND phase and a rewriting buffer, evaluating all four assignments (256 checks),
validating copy bindings and checking the resulting graphs. The complete AIG
race suite passes in 61.561s (`/tmp/abcgo-miter-final-review.log`). A source search
finds no production Go caller of ExtractMiter; the original C FRA caller is a
diagnostic dump helper. The repository-wide run above preceded this isolated
miter change, and is not reported as a fresh run of its new code.

The remaining manager inventory includes Start's nonpositive soft-capacity
contract (C accepts negative hints; Go currently panics). Add a red regression
before changing that behavior. aig_.go is only commented empty C scaffolding;
aigTest.go matches the original two-output demonstration, using the current CO
constructor in place of C's obsolete Po spelling. The remaining file/function
inventory still needs completion before closing aig/aig. The sequential goal
remains active, with 1/131 original packages completed.

### Manager start and circular-order inventory checkpoint

A red regression confirmed that Aig_ManStart rejected negative soft capacity
hints, whereas aigMan.c treats every nonpositive hint as unknown. Removed the
extra panic; Go keeps its growable slice policy. Zero, -1, -1000 and minimum
int32 hints now create valid graphs and evaluate the same four assignments.

Read every circular-order routine against aigOrder.c, including growth,
sentinel links, insertion-before-next, current/next cursor removal and Advance's
separate previous/next responsibilities. No production mismatch was found.
Added a 1,000-step independent slice-model test checking the complete order,
reciprocal links, counters and cursor transitions, plus insertion beyond object
ID 4096 and restoration of the empty sentinel. Go safely resets counters when
restarting order storage rather than retaining source Stop's stale counters.
Focused manager/order race checks pass
(`/tmp/abcgo-manager-order-review.log`). No production Go caller currently
uses the order API outside this package's tests.

The inventory also found that Aig_ManSupports still performs per-output DFS
instead of C's reference-counted pooled support propagation, despite the pool
helpers being present. That substitutes an algorithm and changes traversal
scratch. Reopen that specific partition-support entry point next, with a red
scratch regression and the original C propagation as the implementation guide.
The full AIG package remains in progress; completed count is 1/131.

### Restore pooled support propagation and register matrices

The inventory exposed a remaining algorithm substitution in Aig_ManSupports:
Go reran cone DFS for every output rather than executing aigPart.c's pooled,
reference-counted propagation. Red regressions showed traversal generation and
node marks changed; source propagation leaves them alone. A raw-output case
also failed because a CI's output index overwrote its input index in the
terminal union. The native source implementations, not the old Go routine, now
guide this entry point.

Restored one propagation pass through dependency order. CI/constant support
records carry consumer counts; binary gates merge sorted supports; each
physical pin releases its record and the last release recycles it. Output rows
copy their support before recycling and retain their explicit output index.
Go uses local typed records instead of C's temporary pData pointers. Caller
payloads/copies remain owned by the caller, and source-style CioID setup/cleanup
is preserved. Dependency order supports existing forward-reference/buffer/CO
extensions; legacy raw outputs are explicit final consumers and use stable
input positions. No per-output DFS remains in this production path.

Part_ManMergeEntry now merges directly into its fetched record's slice, avoiding
an intermediate allocation. The old per-cone helper is confined to tests as an
independent oracle. RegisterComponents reuses the register dependency matrix
rather than building another support traversal/index map. A further red test
showed Aig_ManSupportsRegisters returned empty dependencies after a previous
whole-support call cleared CioIDs. It now filters the propagated, explicitly
indexed rows as in C, and preserves traversal scratch.

A compiled fixture executes original allocation, propagation, register-matrix
and component routines on 256 graphs with physical AND/XOR nodes, shared and
dangling cones, constants, shuffled CI/CO order and zero through five registers.
All output supports, register dependency rows, component diagnostic text,
manager/node traversal state, physical references and terminal cleanup match.
The test checks decreasing row-size order, then compares rows by their explicit
output indices; C qsort does not specify ordering among equal-size rows, while
Go keeps its deterministic stable tie order.

Additional tests cover transparent unary/CO fanins, a forward buffer reference,
raw output indices, payload preservation, and independent ownership of 100
returned rows sharing a 128-input cone, including a second propagation call
following mutation of an earlier result. Graph invariants pass. Focused final
race/native checks pass in 2.505s
(`/tmp/abcgo-support-propagation-shared-final.log`). The complete AIG race suite
passes in 63.845s before the last helper-only consolidation and wide-row test
addition (`/tmp/abcgo-support-propagation-final.log`). Earlier caller checks
pass FRA/SSW/ABCI in 35.227s/3.176s/28.618s
(`/tmp/abcgo-support-propagation-review.log`); those preceded the final
register-support helper correction, which is covered by the latest native
matrix comparison. These scopes are recorded separately rather than claiming
a fresh whole-repository run after this repair.

This closes the reopened whole-support propagation, register-support and
component-caller contracts. Other partition-driver and remaining file/function
inventory still need final reconciliation. The audit remains on aig/aig,
with 1/131 original packages completed.

### Smart partition scoring and driver boundary review

Read the bitmap helpers, attraction/repulsion selection, smart output/register
wrappers, support-size ordering, compaction and naive grouping against
`aigPart.c`. A new red regression found the size-101 score was wrong:
C's Abc_Base2Log(1) returns 1, so repulsion is 2; Go treated that argument as
zero, making repulsion 1. With one common variable out of ten, Go accepted a
partition at score 100 while C rejects it at 50. Corrected the zero/one special
case without changing the remaining scoring formula.

A native fixture executes the real bitmap/scoring helpers for 8,190 cases:
empty supports, full containment, disjoint/partial overlaps, word boundaries,
size 99/100/101/102 and larger supports, one/two candidates, support caps,
negative/default limits, and exact score-75 acceptance. Another fixture executes
complete C smart partitioning on 25 graphs around size 101 with disabled,
default and explicit compaction limits. Selected partitions and their support
contents match exactly for these cases; output coverage and Go support-vector
lengths are independently checked.

The original compactor shrinks the parts vector twice, leaving extra tail rows
in its returned support vector. The C fixture observes the valid prefix of
support rows just as the original clients do. Go retains the earlier repair
that keeps both returned vectors aligned, as well as the pending-small-before-
large coverage repair. No attempt was made to reproduce those source container
bugs. Naive grouping's quotient/remainder calculation matches the Go grouping;
existing miter tests exercise its output maps and positive-size domain.

Full AIG/FRA/ABCI race suites pass in 66.038s/12.146s/29.273s
(`/tmp/abcgo-partition-scoring-review.log`). The subsequently expanded scalar
case matrix and both native fixtures pass focused race checks in 5.530s
(`/tmp/abcgo-partition-scoring-final.log`). This closes the current scoring and
these smart-driver boundary checks, not every partition-copy contract.

### Remaining typed-XOR partition-copy defect

Source inspection and independent probes expose another real correctness issue
in the existing copy compatibility path. A partitioned miter comparing native
XOR(a,b) with AND(a,b) returns truth zero in both smart and naive modes, instead
of the correct truth nibble 14. The existing TestDupPartNativeXorMatchesCAndCopy
explicitly expects XOR to turn into AND, so its green result is not correctness
evidence. A compiled original-C probe confirms the upstream limitation: with
fCatchExor=0 (expanded AND representation) its miter is truth 14; with
fCatchExor=1 (native EXOR node) it incorrectly becomes truth zero in both modes.
Logs: `/tmp/abcgo-partition-xor-probe.log` and
`/tmp/abcgo-partition-xor-source-probe.log`.

Reopen the typed-XOR copy path next. First add a red functional miter/clone
regression; preserve the AND-only fragment representation expected by the
partition backends while expanding native XORs with their actual Boolean
semantics. Keep the original-C bug explicitly demonstrated rather than
silently claiming source parity proves correctness. Review reverse copy maps
for any expansion nodes and cover direct/inverse duplication, whole-part copies,
partitioned miters and FRA/choice transfer before closing that repair. The
package audit remains active on aig/aig, completed count 1/131.

### Native-XOR partition copy repair

Replaced the test that expected native XOR to become AND with a red truth-table
regression. A second red test showed both smart and naive partitioned miters
incorrectly report XOR(a,b) equivalent to AND(a,b). The compiled original-C
fixture explicitly demonstrates that upstream defect for native EXOR, while
its default expanded-XOR representation yields the correct mismatch truth 14.
Go now expands native XORs using the existing AND-based XOR constructor in
partition-copy paths. Other cloning paths retain native XOR nodes.

The expansion exposed a second bug with a red reverse-copy-map regression:
copying a complemented implementation root discarded its polarity when mapping
back to the original graph. Both whole-part reverse-map writes now preserve
that bit. All eight input/output polarity combinations are checked over all
four assignments, including forward and reverse map semantics. The map contains
only the five original objects; the two new implementation nodes have no
invented source bindings. Existing representative transfer skips absent maps.

Additional tests cover reversed input ordering, direct and inverse copying,
complemented outputs, constants and input aliases. A test-only exhaustive
classifier exercises FRA and choice transfer on native/expanded XOR versions
with an independent AND output. The classifier rejects native XOR fragments
and overwrites its scratch copy map to check driver restoration. Result truth
tables and graph consistency pass. This repairs a source bug rather than
claiming that incorrect native C behavior is the intended Boolean contract.

The sequential audit remains on aig/aig, with 1/131 original packages completed.

Validation: full AIG/FRA/ABCI race suites pass in
68.859s/12.335s/29.071s (`/tmp/abcgo-partition-xor-review.log`). The final
expanded focused race tests, including buffer aliases and exact reverse-map
ownership, pass in 2.316s (`/tmp/abcgo-partition-xor-final.log`). No fresh
whole-repository run is claimed for this isolated repair.

### AIGER snapshot and operation experiment reconciliation

Read the entire AigerModel adapter and compared constant polarity, CI/register
layout, zero initialization and constraint inversion against
`src/aig/ioa/ioaWriteAig.c`. Its AND-only normalization is necessary for typed
XORs, buffers and non-ID-ordered DAGs; it does not replace a synthesis algorithm.
Added a deterministic matrix of 128 mixed four-input graphs, each with a forward
buffer reference and CO-as-fanin, randomized AND/XOR/buffer cones, reordered
interfaces, zero/one/two registers, all applicable constraint counts and deletion
holes. Independent AIGER-literal evaluation checks all 16 assignments, both
before and after binary serialization (4,096 snapshot evaluations). Node state,
interface lists and copy scratch bindings remain unchanged; names, register
counts and constraint counts are checked. No adapter bug found.

Reconciled aigOper.go with every entry point in aigOper.c: elementary-variable
creation, operations, source-default MUX path, majority, balanced multiargument
construction, in-place miter compaction, constant reductions, hash lookup,
EXOR recognition and n-variable constructors. These match their source flows;
existing exhaustive small truth tables and mutation regressions cover them.
Found Aig_MuxTest had instead become a tiny invented consistency check. A new
red regression exposed missing object-count reports and missing test1.blif.
Restored the actual source experiment: 20 inputs, 2,000 sequentially sampled
MUXes, seed 321, cleanup, two object-count reports, BLIF export and manager
release. A private 31-word additive generator reproduces Linux/glibc srand/rand
without changing another package's RNG. Other C runtimes have different random
sequences; this portability distinction is explicit.

A compiled original-C experiment observes the final graph at its export
boundary. All 2,000 output AND structures match after canonicalizing IDs (C
argument evaluation order allocates sibling gates in a different order from
Go). Both object counts are exactly 5,503. The public Go entry point separately
writes a real BLIF in a temporary directory. Removed the old operation test's
incidental experiment invocation; the dedicated artifact test owns this file
side effect. Focused uncached tests pass. Full uncached AIG race tests pass
in 69.915s (`/tmp/abcgo-aiger-mux-audit.log`); `go vet ./aig/aig` also passes.

Inventory reconciliation also identifies aigPartSat.go as still requiring a
complete source-backed review of partition labels, PIO mappings, incremental
CNF/solver feedback and conflict budgets. Earlier representative/FRA review
covered aigPart.go, not this separate SAT driver. Do not mark the package
complete based on those earlier checkpoints. Completed count remains 1/131.

### SAT partition labels and traversal source checkpoint

Read aigPartSat.c in full and began its source-backed reconciliation. Added a
red minimal levelized-label regression: C's reverse level/ID iteration includes
COs, CIs and the constant, so a two-input AND with one output has size-one labels
[4,3,2,1,0]; Go produced all zeros by counting gates alone. The repaired helper
counts every live object in that same order. Its sorted enumeration retains
Go's existing ability to handle dangling gates above the output level range;
C's Levelize asserts outside that domain. Deletion holes remain zero-labelled.

A native fixture executes all four original label generators for 32 cleaned
random AND graphs at four partition sizes (512 result rows). Exact object-ID
maps now match, including holes, constant/PI outputs and empty gate sets.
The fixture cleans dangling gates to meet C DFS completeness assertions.

Two further red cases found generic DFS visited choice alternatives after
ordinary fanins and left traversal state unchanged. Partition DFS now uses the
original setup and the already source-reviewed recursive helpers. Preorder's
choice alternatives still use postorder, as in C. The partition wrapper keeps
Go's typed unary/buffer/dangling-node domain by omitting whole-manager count
assertions; the source-valid AND graph order is unchanged. Choice-first order
and traversal generation are explicitly checked in both modes.

A separate red rollover regression found PartSplitOne incremented its traversal
counter manually, escaping the bounded generation range used elsewhere. It now
calls Aig_ManIncrementTravId, removing duplicated rollover code. The boundary
map, copied-node marks and bounded generation are checked at the wrap threshold.
Focused label/native/SAT/splitting tests pass. Full AIG/ABCI race checks pass
in 70.761s/29.622s (`/tmp/abcgo-sat-partition-labels-final.log`). Final focused
race tests pass in 2.187s (`/tmp/abcgo-sat-partition-focused-final.log`).
The ABCI partition-SAT command is the production caller.

The complete SAT driver review remains open. Source uses cut-mapped CNF
(Cnf_Derive, polarity transformation, lifting) and BSAT, whereas the current
Go file emits per-node Tseitin CNF into SATOKO. Its existing truth-table tests
are correctness evidence for those formulas, not proof of a faithful port of
that source pipeline or conflict-limited behavior. Source also has sparse-label,
constant-output and cumulative-conflict-accounting defects that need explicit
source observations before deciding which behavior to preserve or repair.
Next reconcile boundary maps against native SplitOne, then restore the actual
mapped-CNF/BSAT driver using the higher-layer pattern already applied to
proof/aiginterp if package cycles require it. Do not call the current SAT path
fully source-reviewed based on green functional tests. Completed count 1/131;
aig/aig remains active.

### SAT partition boundary alias repair

A new red typed-buffer regression found PartSplitOne dropped an original
boundary object when a buffer and a later AND simplified to the same copied
node. Copied reference counts no longer identify original cut crossings in
that situation. Count physical crossings in the original selected-node set
instead; retain separate CO terminals for distinct original IDs even when
their copied drivers alias. The complete partitioned solve regression now
returns a valid model through every partition/alignment mode.

A compiled native fixture checks all seven nonempty subsets of a shared
three-gate cone in forward and reverse selection orders (14 cases): exact
PIO vectors, object/fanin IDs, refs and copied restriction marks match the
original C. Thus the ordinary source AND-only boundary contract is preserved.
Full uncached AIG race tests pass in 72.483s
(`/tmp/abcgo-sat-boundary-final.log`).

The next mapped-CNF regression is deliberately red: on a six-input AND chain,
the old per-node helper allocates 15 solver variables instead of Cnf_Derive's
11, with different terminal IDs. This test is pending migration with the
mapped-CNF/BSAT driver. It is not part of the boundary-only checkpoint.

### Restore mapped-CNF and BSAT partition solving

The red six-input layout test exposed 15 per-node SATOKO variables instead of
11 mapped-CNF variables and incorrect terminal mappings. Another red test found
the original synthesis flag simply returned an unsupported-backend error.
Removed the per-node Tseitin/SATOKO implementation. The real driver now lives
in proof/aigpartsat above AIG/CNF/BSAT/DAR, following the interpolation layering
pattern to avoid dependency cycles. ABCI's partitioned-SAT command calls the new
package; graph partitioning, boundary maps and polarity helpers stay in AIG.
Moved its functional tests to the implementation layer using public graph APIs.

The incremental helper now calls Cnf_Derive with free COs, transforms polarity,
lifts into the existing solver namespace, creates the exact mapped variables,
adds mapped clauses and terminal equalities, records the constant binding, then
adds original-output and restriction units. BSAT retains clauses and learned
state between partitions and supplies the model. Source entry synthesis calls
the actual DAR Rwsat pass. Its nConfPart argument remains unused as in C;
the modern options API can explicitly request a per-call cap.

A compiled fixture links the actual source AIG, DAR cuts, mapped-CNF and BSAT
implementations, then invokes the original incremental helper twice. Variable
maps, 11/22-variable lifting, connector clauses and the complete recorded clause
sequence match exactly in both polarity modes. No fake CNF or SAT algorithm is
used in that oracle. Existing exhaustive small/random Boolean and model tests
now exercise this pipeline, alongside synthesis failures/interface validation,
constants, PI outputs, sparse labels and bounded results.

Typed native XORs/buffers normalize through the already reviewed AIGER adapter
before the AND-only mapper. A new red synthesis-callback regression found that
normalization must also occur after a rewrite returning buffers; this is now
covered. The collapsed-buffer boundary case also runs through every solver
partition/alignment mode in the new package.

Retained explicit correctness repairs outside the source's safe domain:
isolated PI/constant outputs are constrained, shared constant boundary CIs are
forced true, different terminal phases connect with the correct sign, total
budgets subtract newly spent rather than cumulative conflicts, zero means
unlimited, and an exhausted budget before the final partition returns unknown
rather than a partial SAT answer. Original C's sparse-label grouping defect
remains repaired in the graph helper. These are visible repairs/extensions,
not claims that a source-equivalent substitute formula is the same algorithm.
Native source demonstrations of remaining driver edge cases/conflict controls
and final file/function reconciliation are still needed before closing AIG.

Uncached migration integration passed AIG/aigpartsat/ABCI in
42.212s/0.385s/14.428s before the final native/callback tests
(`/tmp/abcgo-partsat-migration-integration.log`). The complete new package,
including those tests, passes with the race detector in 8.768s. The subsequent
AIG/ABCI race suites are still running and are not claimed finished here.
The original 131-package audit remains active; this driver migration adds
proof/aigpartsat to the package inventory, just as proof/aiginterp was added.
Only the root package is marked complete.

### Native partition-driver edge-case proofs and integration

The full AIG/aigpartsat/ABCI race suites completed successfully in
72.463s/8.768s/28.722s (`/tmp/abcgo-partsat-migration-race.log`). After adding
source driver edge-case observations, the complete new driver suite passed
race checks in 16.138s (`/tmp/abcgo-partsat-source-final-race.log`). The final
expanded source/clause checks pass race validation in 16.233s
(`/tmp/abcgo-partsat-source-edges-final.log`).

The fixture now executes the actual original partitioned driver with total
budget zero on two incompatible one-gate outputs, (a AND b) and (!a AND b).
C stops after the first partition, returns SAT and supplies an invalid witness.
The Go source entry point treats zero as unlimited, completes the formula and
returns UNSAT. This demonstrates the need for the preserved soundness repair;
matching the erroneous original return value would be a correctness regression.

A second observation executes original PartSplit with a sole gate labelled 7:
C returns one graph but eight output groups, leaving the first group's output
list empty. Go retains one graph and one correctly populated output group.
The source fixture's unreachable DAR rewrite callback is an abort trap because
these driver cases select fSynthesize=0; actual cuts, mapped CNF, solver,
partitioning and model extraction all run from source. The source tests do not
claim to execute native DAR synthesis in that fixture. Go's DAR entry path is
covered by the previously red synthesis test.

Repository-wide `go test -p 1 ./... -run '^$' -count=1` passes for all 133
current packages, including test compilation, after the API migration
(`/tmp/abcgo-partsat-build-all.log`). This is a build check, not a claim that
all package test bodies ran. No experiments directories are in this module.
The original sequential goal remains active with only root marked complete;
AIG still needs final file/function inventory reconciliation and any remaining
source control-flow gaps recorded by that check.

### Final family coverage: frames, windows and standalone truth words

Re-read all of aigFrames.go/aigFrames.c and added a 64-case source matrix:
one through four frames and every combination of init/outs/regs/enlarge.
Two inputs and two interacting registers exercise complemented next-state
logic, reconvergence, ordinary outputs and an unused internal gate. Exact
returned maps (including deleted copies), live object refs/levels/phases,
terminal IDs, count metadata, name/spec and 13,600 exhaustive assignment
results match the compiled original C. Source graph object state is unchanged.
No frame bug found; native XOR/buffer support remains the existing typed
extension, rather than reproducing C's EXOR-as-AND limitation.

Read every window routine against aigWin.c: immediate frontier costs,
high-fanout exception when the frontier cannot grow, level tie-breaking,
stable removal order, marked construction zone and final mark cleanup.
All 512 native cases match exact frontier/visited order and cleared marks:
32 deterministic random AND graphs, four size limits and four fanout limits,
with multiple output consumers and complemented pins. No window repair needed.

Compared standalone truth-word combination against aigTruth.c for four word
counts, all four AND pin phases and source-normalized XOR inputs (20 cases,
120 words). The recursive Go wrapper retains owned truth tables and avoids
source pData scratch aliasing; its error checks and constant/root-phase
handling were read alongside existing truth, cut and factoring tests. No
truth-word bug found in this checkpoint. Focused native frame tests pass in
1.011s; window/truth tests pass in 1.115s. Header/allocation helper regressions
from the subsequent inventory are being handled separately before closure.

### Header records, ghost types and absent children

Re-read the executable aig.h adapters rather than counting the header's
commented declarations as reviewed code. Red tests found CI/CO/buffer ghosts
were silently labelled AND; map each supported source kind explicitly.
Another red test found raw FetchMemory allocated an extra constant, and Type
panicked on a recycled object. A raw C reservation is AIG_OBJ_NONE but still
contributes to ObjNum until recycled. Preserve both facts with an internal
vacant kind, using the existing deletion map only for actual recycling. No
extra per-node or per-manager storage was added.

Native comparison caught the initial manager ghost having the same zero-value
constant confusion. NewGraph, raw allocation and Clean now share explicit
empty-record initialization where appropriate. Constants/CIs have absent
children, distinct from constant-one edge zero. Red tests also found missing
children reported as complemented because InvalidEdge has its low bit set;
source-style complement/fanin accessors now treat the missing edge as untagged.
The source inline helpers confirm that NULL children have no complement bit.
The small existing bit-count helpers were retained: their explicit C flow is
already compact and does not need an unrelated algorithm rewrite.

A compiled original-header fixture compares initial ghost state, all five
supported ghost kinds, sorted/complemented children, terminal child absence,
raw/cleared/recycled object types and exact 4-to-3 object counts. It passes,
as do the direct red/green regressions. Final focused header race checks pass
in 2.107s (`/tmp/abcgo-header-records-final-race.log`). Prior full AIG/partsat/ABCI
race checks passed in 74.627s/16.237s/29.052s before the final initial-ghost and
missing-phase refinements. The current full-module uncached run has already
passed AIG in 45.166s; remaining packages are still running, so no whole-module
success is claimed yet.

### File inventory reconciliation

All 52 current production files are accounted for below. This inventory links
back to the detailed source comparisons above; it is not substitute evidence
for those checks or a declaration that the package is complete.

| Files | Source review and evidence |
| --- | --- |
| `aig.go`, `graph.go` | Header/node/edge construction, lifecycle, hash and mutation checkpoints; native header states above. |
| `aigConstruction.go`, `aigInterface.go` | Construction-only counts, interface permutations, preserved CioID union contracts and metadata tests. |
| `counterexample.go` | State-before-PI witness layout, ownership/dimension guards, sequential/induction witness tests. |
| `choice_access.go`, `phase_access.go`, `object_data_access.go`, `aigReprStorage.go` | Small typed accessors read against corresponding source fields and owning algorithms. |
| `fanout_status.go`, `level_modes.go` | Fanout ownership and structural/choice/reset timing conventions; circular-link and XOR level checkpoints. |
| `aigCheck.go`, `aigObj.go`, `aigMan.go` | Physical refs, lifecycle, local CO patching, predicates, manager start/stats and cleanup source matrices. |
| `aigMem.go`, `aigTable.go` | Owned pools/recycling, retained source large blocks, hash keys/ordering and ghost lookups. |
| `aigDup.go`, `aigDfs.go` | Copy/interface metadata, phase-normalized miters, all source traversal families, transfers/composition and source ordering. |
| `aigMffc.go`, `aigRepr.go` | Reference/frontier state, representative chains/phases, cleanup and choice-loop filtering. |
| `aigCuts.go`, `aigTruth.go`, `aigWin.go` | Cut merging/truth/dominance and window heuristics; new native word/window matrices above. |
| `aigFrames.go` | All 64 source flag/frame cases, exact maps/state and exhaustive truth vectors above. |
| `aigTiming.go`, `aigFanout.go`, `aigOrder.go` | Incremental forward/reverse updates, circular pin links and 1,000-step order model. |
| `aigDoms.go`, `aigFact.go` | Source dominator/factoring flow, IDs/marks/copies, native state and packed support truth checkpoints. |
| `aigTsim.go`, `aigMvSim.go` | Original ternary/multivalue streams, packing/hash/history/state scopes and reduction drivers. |
| `aigCanon.go` | Source signature/permutation/truth/library workflow and 700 native cases. |
| `aigRet.go`, `aigRet_storage.go`, `aigRetF.go` | Source edge/frontier retiming, packed overflow values, reset/reconstruction, metadata and sequential traces. |
| `aigScl.go` | Full/basic sequential cleanup, flop/register maps and exact native reference/fanout state. |
| `aigSplit.go` | Source BDD-cofactoring driver restored; CUDD integration and source safety defects documented. |
| `aigPack.go`, `aigJust.go`, `aigJust_experiment.go` | Source packing/rare-justification flow, random streams, all heuristics and experiment histograms. |
| `aigPart.go`, `aigPartReg.go`, `aigPartSat.go` | Pooled supports, attraction scores, partition/representative maps, typed-XOR repairs and native SAT boundaries/labels. Solver driver moved to `proof/aigpartsat`. |
| `aigInter.go`, `aigRepar.go` | Source clause/append primitives; actual interpolation/reparameterization drivers moved to `proof/aiginterp`, including explicit upstream failures. |
| `aiger_model.go` | Source writer constants/constraints/register layout; mixed-graph exhaustive adapter tests. |
| `aigShow.go` | Source rendering/filename/external-viewer ownership and output/mark tests. |
| `aigOper.go`, `aigTest.go` | All source constructors and original seeded MUX experiment; demo spelling updated to CO, with write-failure regression. |
| `aigUtil.go` | Printing/writers, RNG, lists, phase/reference utilities and witness helpers; **remaining direct scratch-cleanup contract review below**. |
| `assert.go`, `aig_.go` | Assertion caller location/format; source template has no executable body. |

The remaining utility review must explicitly reconcile CleanNext with C's
pNext/CioId union: clearing only the typed next map appears to preserve CioIDs
that C zeroes. Add a red regression before changing it. Also check recursive
CleanData's root assertions and typed traversal extension against the source;
source ternary Xsim is nCuts, not pData, so CleanData must not erase Xsim.
The latter is a confirmed storage distinction, not a bug to invent or fix.
AIG remains in progress, original completion count 1/131.

### Final utility scratch contracts and module gate

The uncached sequential module suite completed with 130 packages passing and
three without test files, exit zero (`/tmp/abcgo-final-aig-integration.log`).
That run covers the constructor/header and demo repairs, not the subsequent
small scratch-cleanup patch described here.

Two new red regressions closed the identified utility gaps. CleanNext now
clears live objects' CioID fields as well as typed next bindings, matching C's
pNext/CioId union. Go's separate structural hash tables remain valid; no
algorithm needs to emulate C's incidental destruction of hash-chain storage.
Recursive CleanData now rejects complemented roots and physical CO roots
before changing payloads, as C asserts. Its existing typed XOR/buffer cone
extension remains explicit; ordinary AND/CI cleanup order/scope is unchanged.
Neither entry point has a production Go caller outside its own declaration.

A compiled utility fixture compares every object's CioID, iData-equivalent
payload and ternary nCuts/Xsim state after next cleanup, global data cleanup
and regular AND-cone data cleanup, including an unrelated cone. All states
match; global data cleanup preserves Xsim. Direct tests additionally preserve
unrelated Go payload/copy bindings and exercise rejection without mutation.
Focused race checks pass in 2.019s (`/tmp/abcgo-utility-cleanup-final.log`).

Read the remaining utility bodies against aigUtil.c: MUX/XOR recognition,
buffer phase resolution, all-zero phase simulation, RNG reset/64-bit ordering,
random-info filling, ordered list union/intersection and witness-session
allocation/simulation/lookup/stop. Existing exhaustive MUX phase/order tests,
RNG golden vectors, list cases, sequential witness checks, invalid-session
checks and writer/state tests cover those contracts. RandomTest2's one-line
file is the actual source behavior: its disabled LFSR assigns one and exits
its do/while immediately. RandomTest1 retains the Go formatter's defined
unsigned decimal output for uint32 draws; C's signed %d with an unsigned
vararg is not a portable signed-stream specification. No new random algorithm
or replacement witness algorithm was added. The timestamp/file writer and
external-viewer portability changes remain those already documented above.

The 52-file inventory is now reconciled. Final full AIG race validation and
vet of AIG plus the two migrated driver packages are in progress before the
status row is advanced. Known upstream unfinished/broken experimental
entry points remain explicit source failures, as recorded in the interpolation
and register-partition checkpoints; review completion must not imply those
upstream routines suddenly have invented successful implementations.


### AIG audit closure and compact node layout

Completed the final source/correctness/readability reconciliation for all
52 production files. The remaining utility gaps above are repaired and
source-checked. Earlier in-progress statements are chronological checkpoints;
this closure supersedes them. All bugs uncovered in this audit have red/green
regressions or an explicit documented upstream-failure contract. No invented
successful algorithm was retained in place of a source driver.

A final storage review grouped Node.Phase with MarkA/MarkB, removing trailing
alignment padding without changing field types, arithmetic, graph algorithms
or adding unsafe production code. On amd64 the measured record shrank from
56 to 48 bytes (eight bytes per node); a temporary read-only Sizeof/Offsetof
probe produced `/tmp/abcgo-node-layout-final.log`. No repository consumer uses
raw Node offsets or serialization. The full module recompiles, including test
literals, so no positional initializer depends on the previous order.

Final checks on this layout:

- Full uncached AIG race suite passes in 75.961s
  (`/tmp/abcgo-aig-layout-final-race.log`). The preceding final utility/header
  state passed the full AIG race suite in 76.344s as well.
- `go vet ./aig/aig ./proof/aiginterp ./proof/aigpartsat` passes.
- All 133 packages and their test files compile with the reordered record
  (`/tmp/abcgo-aig-layout-build-all.log`, exit zero). This is a compile gate,
  not another run of every unrelated test body.
- The earlier full uncached module execution passed 130 tested packages and
  three without tests (`/tmp/abcgo-final-aig-integration.log`). Its scope
  precedes the two unused cleanup-entry repairs and final field reordering;
  those are covered by the final AIG native/race checks and module compile gate.

The source's unfinished linear register partitioner and broken experimental
reparameterization remain explicit, tested upstream limitations rather than
invented replacement algorithms. Pool accounting reflects Go-owned storage;
source pointer/union storage is represented by typed state and stable IDs.
Recorded signed-XOR, forward-DAG/buffer and source-defect repairs remain
intentional correctness extensions. Review completion is not a claim that
these upstream experimental failures have been mathematically solved.

Original package progress is now **2/131**, with `aig/gia` next. The full
sequential goal remains active; none of the other pending packages is counted
as audited merely because its integration tests passed.


## 3. aig/gia (in progress)

The complemented-handle storage review removed the per-object boolean tag:
inverse views use object ID -1, outside the valid nonnegative manager ID
domain. Atomic pairing, stable aliases and regular-object fields are unchanged.
On amd64 this reduces Gia_Obj_t from 64 to 56 bytes, measured with a temporary
Sizeof probe (`/tmp/abcgo-gia-object-size-final.log`). Existing involution,
conditional-complement and concurrent-pairing regressions pass with the full
uncached GIA race suite (15.745s). `go vet ./aig/gia` passes, and all module
packages including tests compile (`/tmp/abcgo-gia-core-layout-build.log`).
These checks cover this representation change, not the remaining GIA audit.

Inventoried 544 production files and 405 test files at the package root;
subpackages mf/satcost remain separate pending audits. Historical GIA review
reports and passing tests are context, not evidence that this sequential audit
has finished. Begin with literal/object encoding, two graph representations,
manager construction/lifecycle, append operations and object import/export.
Then continue through hashing, fanouts/references, special gates and the
remaining source algorithm families. Original completion count remains 2/131.

Read graph.go, literal.go, object.go, graph_objects.go, manager.go,
manager_accessors.go, append_simplified.go and special_gates.go against
src/aig/gia/gia.h and Gia_ManStart in giaMan.c. The positive-capacity start
precondition is real in C (unlike AIG's soft-capacity convention). Pointer
complements are paired Go views; source Gia_NotCond XORs a tag restricted by
call sites to 0/1, so do not replace it with a guessed nonzero test.

A new red adapter test found GraphFromObjects accepts real MUX objects as
ordinary ANDs and silently computes the wrong function. Their two stored data
fanin offsets have ordinary AND ordering; the control and type discriminator
live only in the companion manager table. The adapter already promises to
reject specialized nodes requiring normalization, but checked only real XOR.
Retain a pointer to that table's slice header in the shared object store and
reject real MUX imports explicitly, preserving the documented adapter domain.
This adds one pointer per store/manager, not a flag or pointer per object, and
handles control-table enabling/growth without stale snapshots. Normal manager
MUX simulation/conversion remains supported by its existing metadata-aware
APIs; no fake AND interpretation is retained at this boundary.

The red matrix covers all eight input phase combinations, manager growth and
late control-table activation; a positive case keeps ordinary AND import valid
when an allocated table has no MUX entries. A standalone original-header
fixture confirms classification and encoding plus every three-input assignment
for all eight MUX polarities: IsAnd is true, IsXor false, IsMux true, and the
companion control chooses the actual data pin. Focused adapter/source tests
pass in 0.123s. The complete GIA race suite passes in 15.777s
(`/tmp/abcgo-gia-core-first-race.log`).

This is the first GIA checkpoint, not completion of its core or whole package.
Further review must explicitly distinguish fGiaSimple nodes and buffer
classification, real-gate append preconditions, and source-normalized versus
typed gate semantics. Also flagged for the later duplication-family review:
the existing native choice-miter test expects real XOR/MUX nodes to turn into
stored-fanin ANDs to match C; its green result is not Boolean-correctness
evidence. Check the intended source preconditions and callers before treating
such a compatibility test as a correctness contract.

Continued object/link/append review into network.go, levels.go, references.go,
normalize.go and the dynamic/static fanout infrastructure against gia.h,
giaUtil.c, giaTim.c and giaFanout.c. These families are not yet closed.
The existing opposite-phase repeated-fanin correction in Gia_ObjIsBuf is
necessary: simple-mode x & !x is false, not a buffer. A new red normalization
regression exposed a consequence that remained unfixed: DupNormalize creates
a manager without simple mode and panics when appending that legitimate AND.
Without hashing it now retains the source simple-mode flag, matching the
other Go duplicate entry points. With hashing the destination stays in normal
mode and HashAnd simplifies the repeated pins. This intentionally preserves
Boolean correctness beyond C's buffer misclassification, rather than copying
that upstream error. No alternative normalization algorithm was introduced.

The regression exercises all four repeated-input polarities with/without
hashing, interleaved CI/CO ordering, and every two-input assignment (32 rows).
Before the repair, hash0/signs1 panicked in AppendAnd; all cases pass after it.
The full uncached GIA race suite passes in 15.731s, and direct normalization
consumers base/wlc and map/if pass uncached. Logs:
`/tmp/abcgo-gia-normalize-simple-red.log`,
`/tmp/abcgo-gia-normalize-simple-race.log`, and
`/tmp/abcgo-gia-normalize-consumers.log`.

Next checks include dynamic fanout insertion/removal with repeated pins and
special gates: the source's two-slot edge encoding cannot represent a MUX
control edge, and AddFanout chooses the first matching slot for repeated
fanins. Establish supported caller domains before changing that structure.
Original package audit completion remains 2/131.

Dynamic fanout review found and repaired a repeated-pin ring corruption.
The red regression first panicked while FanoutStart was adding an ordinary
consumer after a simple-mode x & !x: both arcs of that AND had selected pin 0,
overwriting its links. Incremental AppendAnd had the same issue, and also
added two copies of the single logical arc for same-phase repeated pins.
Insertion now selects an unlinked matching pin, removal a linked matching
pin; both use the existing five-word/two-pin circular-list representation.
AppendAnd follows FanoutStart's existing single-arc buffer convention.
No per-node metadata or replacement fanout algorithm was added. The shared
pin selection also rejects duplicate insertion/absent removal rather than
overwriting existing links. Insertion now reuses the existing record accessors
instead of repeating their index formulas in local closures.

The regression covers four pin-phase combinations with fanouts built before
or after the nodes, partial/complete repeated-edge removal, reinsertion, both
link directions and removal to empty. A separate test crosses the 4096-object
initial allocation and removes interior/first/last edges after growth.
A native fixture calls the original giaFanout.c insertion/removal operations
and compares all five record words for 42 snapshots across all six removal
orders of three distinct consumers, followed by reinsertion. It constructs
the initial edges explicitly: source FanoutStart's old missing-child pointer
checks do not match modern GIA terminal encoding. Go's kind-aware initializer
is tested separately, not claimed to be the original C initializer running.
The native fixture needs no substitute graph/list functions or prebuilt ABC
objects. Focused repeated-pin/native checks pass in 0.132s; the first full
uncached GIA race check passes in 15.955s and vet passes. After accessor
cleanup the full uncached GIA race suite passes in 15.926s and vet passes
again (`/tmp/abcgo-gia-fanout-final-race.log` and
`/tmp/abcgo-gia-fanout-final-vet.log`). The original failing regression is
recorded in `/tmp/abcgo-gia-fanout-repeated-red.log`.

No production Go caller explicitly starts or edits dynamic fanout rings;
the package's append hooks operate if a caller has enabled them. Real MUX
control dependencies remain outside the two-pin dynamic representation;
static fanouts/reference counting already include those controls. That source
limitation is not claimed resolved by this repeated-pin repair. GIA remains
in progress; completion is still 2/131 original packages.

The flagged choice-miter behavior is a confirmed Boolean defect, not a valid
test expectation. Changing the existing real-gate test to require XOR and MUX
functions produced a red failure on assignment 0000: all six output polarities
were wrong. Source giaDup.c's recursive copier only handles ordinary ANDs;
DAR callers supply those through Gia_ManFromAig. Go accepts explicitly typed
real gates, but had grouped AND/XOR/MUX/buffer into the same stored-fanin AND
copy. It now dispatches on gate kind: ordinary AND uses HashAnd, XOR/MUX use
the already-ported source HashXor/HashMux AND expansions, and buffers forward
their input. MUX controls are recursively copied as well. The original DFS,
Value cache, shared inputs and interleaved output order are retained.
The destination now uses the original structural-hash allocation/stop lifecycle
instead of the parallel literal-map fallback; no replacement choice algorithm
was invented. The recursive entry point documents the source's active-hash
precondition and has no external Go callers.

The native real-gate fixture now supplies C's own ordinary-AND expansions to
its AND-only copier and compares the resulting miter size and all 16 assignment
rows with Go's real-gate input. Its optional argument retains the original
raw-real-gate source reproducer, whose wrong behavior is not required of Go.
The separate native ordinary-AND conversion oracle still matches exact object
encoding, Value scratch fields, AIG export and truth. Additional analytical
tests cover all 32 input/control/data phase combinations, control-only XOR/buf
cones, constants, output buffers and three-snapshot output interleaving.
These exercise 512 assignments with nine outputs each. The initial red log is
`/tmp/abcgo-gia-choice-real-red.log`; native expansion and focused checks pass
in 1.645s and 1.712s respectively. The full uncached GIA race suite passes
in 15.760s and vet passes; logs are
`/tmp/abcgo-gia-choice-final-race.log` and
`/tmp/abcgo-gia-choice-final-vet.log`.

The sibling HCD copier has the same source-derived stored-fanin assumption;
its production synthesis path also imports ordinary AIGs. It needs a separate
regression/repair before this choice-family review can close. HCD's distinct
dangling-node assertion and register bookkeeping must remain source-faithful.

The HCD follow-up regression also failed: all 32 phase cases left the MUX's
control-only cone unvisited (`/tmp/abcgo-gia-hcd-real-red.log`). Its recursive
C copier is textually identical to the GIA one, so the Go HCD entry point now
delegates to the corrected recursive copier. This removes 12 lines of repeated
logic and ensures both paths normalize real gates consistently. HCD's outer
wrapper still performs its original dangling-node assertion and retains its
original register behavior; it was not replaced by the different GIA wrapper.
The 512-assignment analytical matrix now runs against both entry points.

Expanded the native fixture to call the actual HCD C entry point too. Normal
ABC builds do not include giaHcd.o on this checkout, so the test compiles
giaHcd.c directly into a temporary object and links its real routines with
section garbage collection; no HCD algorithm stubs are used. Both native
miters agree with Go on size and all truth rows when C receives its ordinary
AND expansions. Focused HCD/native regressions pass in 1.841s. The full
uncached GIA race suite passes in 16.109s and vet passes; logs are
`/tmp/abcgo-gia-hcd-final-race.log` and
`/tmp/abcgo-gia-hcd-final-vet.log`.

The remainder of the HCD family (including TFI/choice-link cycle checks) is
still pending inspection; the miter repair alone does not close that review.

Continued the HCD/GIA choice review into hcd_miter.go, hcd_choices.go and
equiv_choices.go against giaHcd.c and giaEquiv.c. Two red regressions exposed
missing MUX-control dependencies in both TFI/cycle queries and bad-choice
pruning. A third case exposed GIA's attempt to descend through the constant
object's nonexistent fanin, which panicked. The TFI regression log is
`/tmp/abcgo-gia-tfi-red.log`; the pruning regression log is
`/tmp/abcgo-gia-choice-remove-red.log`.

The shared DFS now stops at constant leaves, follows real MUX controls in
addition to the stored data pins, handles COs without reading a second fanin,
and skips choice traversal when no next table is installed. It retains the
source's identity/CI checks, visited marks, next-link search and wrapper mark
cleanup. The HCD recursive checker delegates to that implementation instead
of retaining a second copy. Source callers construct ordinary AND choice
graphs, but the Go object API also represents real gates and explicit outputs;
these repairs preserve their actual structural dependencies rather than
requiring the AND-only assumption at this helper boundary.

Bad-choice pruning now marks control fanouts as well, retaining the original
topological marking pass and chain-splicing algorithm without temporary
per-object fanin slices. HCD delegates to the same implementation, eliminating
its duplicate pruning loop. The positive regression verifies that an unused
equivalent alternative survives when the used member is removed. This changes
neither the higher wrappers' register behavior nor equivalence-class policy.

A standalone fixture compiles the real giaEquiv.c, giaHcd.c and giaUtil.c
sources with section collection, requiring no ABC object archive or algorithm
stubs. GIA/HCD agree with C on 1,408 TFI queries through ordinary AND graphs,
zero/one/two-hop equivalent-choice links and NULL roots, including cleared
marks. Eight pruning cases cover retaining alternatives, removing the first
or last member, and removing both; representatives, next links and marks are
compared. Constants and typed-gate extensions are covered separately rather
than claiming that the source's AND-only recursion supports them. An
independent adjacency/BFS test covers another 208 mixed-gate queries, including
XORs, buffers, real MUX controls, output nodes, constant leaves and absent
choice tables, and checks both scratch cleanup and preservation of mark1.

The full uncached GIA race suite passes in 17.088s and vet passes; logs are
`/tmp/abcgo-gia-choice-checks-verified-race.log` and
`/tmp/abcgo-gia-choice-checks-verified-vet.log`.
Choice-class construction/transfer and the remaining HCD synthesis routines
still need review; original package completion remains 2/131.

Equivalence-to-choice copying had the same typed-gate defect as the miter
copiers. A new red matrix showed real XOR outputs being computed as ANDs
(`.../gia_xor_0`: assignment 1 gave false instead of true). Both the GIA and
HCD converters accepted the real gate encoding but copied only its two data
pins with HashAnd. They now share a logical-gate copy helper: recurse through
the correct fanins, use the existing source HashXor/HashMux AND expansions,
or forward a buffer input. Ordinary AND copying and each converter's distinct
representative-conflict, dangling-node and snapshot policies remain intact.
This is a typed-representation correction, not a new equivalence algorithm.

Gate expansion can exceed the original object's count, so destination choice
tables grow after each copied gate and after CO construction. New auxiliary
nodes receive GIA_VOID representatives and zero next links; default zero
representatives would falsely classify them as equivalent to constant zero.
The small three-input MUX regression grows beyond its six source objects and
also checks that conversion invents no equivalence classes. The direct matrix
covers XOR/MUX/buffer input and output phases in 96 cases/768 assignment rows.
Mixed real-XOR/expanded-XOR classes add eight cases/32 assignment rows and
check surviving choice members against their representative phase.

The ordinary-AND native oracle now also calls both actual C equivalence-to-
choice converters on eight phase variants of independent associative AND
implementations. It compares full node kinds/fanin literals, representatives,
next links, counts and 128 assignment rows in total. All match. It reuses the
existing native archive and freshly compiled HCD object; no source routines
were replaced by test implementations. The native check passes in 1.826s.

HCD also omitted C's assertion that a register representative must be constant
zero or another RO. A new red test showed it silently accepting a PI instead;
the precondition is now restored, matching GIA's existing guard. Four valid
register cases (both converters, constant/RO representatives) check all 16
four-frame PI traces, 256 frame evaluations including all RI values. Register
count and the source's zero-initialized representative convention are retained.

Red logs: `/tmp/abcgo-gia-equiv-gates-red.log` and
`/tmp/abcgo-gia-hcd-register-repr-red.log`. Focused logical checks pass; the
full uncached GIA race suite passes in 16.675s and vet passes. Final logs:
`/tmp/abcgo-gia-equiv-copy-final-race.log` and
`/tmp/abcgo-gia-equiv-copy-final-vet.log`. Remaining HCD synthesis and broader
equivalence management still need review; GIA is not counted complete.

Reviewed hcd_synthesis.go against giaHcd.c's parameter defaults, rewrite/
refactor/balance sequences, three-snapshot construction and choice driver.
The algorithm passes and option assignments match the source; verbose output
has fewer intermediate statistics, and Go omits the old debug m3.aig write.
The existing no-synthesis path correctly uses one snapshot rather than C's
three-entry placeholder vector, which would discard outputs or assert on
their count. This review does not claim those C wrapper bugs were reproduced.

Following HCD into giarf_driver.go exposed an actual Boolean failure in
ordinary AND graphs: the source's `level < levels` bound skips the highest
level. A balanced 32-input conjunction was retained as a simulated constant
candidate and rewritten to false with the MiniSat backend. Its complemented
output was wrong as well, despite a structurally valid returned graph. The
red end-to-end case is `/tmp/abcgo-gia-hcd-rare-red.log`; a minimal duplicated-
AND case at level one also showed that no SAT proof ran at all
(`/tmp/abcgo-gia-highest-level-red.log`). The driver now checks the final level.

Resource-limit review found a separate error: a class with unknown SAT results
was marked proved if no query returned a counterexample. The direct red case
reported two unknowns and two proofs; an exhausted-budget end-to-end run still
rewrote the rare conjunction to false even after fixing the level bound.
Unknown class queries now prevent that proof marking. At the final boundary,
only proved, nonfailed equivalences remain eligible for substitution. Filtering
clears rejected representatives in one scan and rebuilds next links once if
needed, rather than repeatedly removing members with quadratic class scans.
This preserves the existing SAT/refinement algorithm and original logic for
unresolved candidates; no alternative solver or heuristic proof was invented.
The strict level bound and unknown/proved branch also appear in the original
giaGiarf.c, so these are documented correctness repairs beyond source behavior,
not claims of native equivalence for those faulty paths.

Proof provenance needs care across levels: a member can be proved only against
a temporary lower-level member while its actual head remains unchecked or
fails a later query. Red tests caught both accepting that unproved head and
selecting an unproved temporary representative. Final filtering requires the
member and its real head to be proved/nonfailed (constant zero is intrinsic);
Gia_ObjTempRepr also requires proof for the selected lower-level candidate.
Known valid equivalences are still retained by the highest-level driver test.
Red logs: `/tmp/abcgo-gia-unknown-proof-red.log`,
`/tmp/abcgo-gia-unproved-head-red.log`, and
`/tmp/abcgo-gia-unproved-temp-red.log`.

The rare-output test now covers both SAT backends, synthesis on/off and an
already-exhausted TAS budget, with 198 assignments and both output polarities.
A four-pigeon/three-hole UNSAT fixture independently exhausts a positive
one-conflict budget on both backends and checks that an unproved constant is
not exposed as a proved equivalence. The previously existing equivalence and
HCD tests remain part of the full package gate. The full uncached GIA race
suite passes in 17.369s and vet passes; final logs are
`/tmp/abcgo-gia-giarf-proof-complete-race.log` and
`/tmp/abcgo-gia-giarf-proof-complete-vet.log`.

Next review includes Giarf simulation/refinement, reduced-level copying and
their typed-gate assumptions, followed by remaining equivalence metadata and
transfer families. Original package audit completion remains 2/131.

Giarf typed-gate review found a stronger failure than an inaccurate simulation:
both simulation entry points and reduced-level copying treated real XOR/MUX
nodes as ANDs. The red end-to-end case put real XOR and ordinary AND on the
same inputs; the driver reported two successful proofs, then returned true
for both outputs at input [true, false]. The AND output should be false. Its SAT
queries had validated the wrong reduced functions, so proof-status filtering
alone could not prevent the incorrect rewrite. The same red file also catches
incorrect recursive/full-word simulation and a wrong MUX root in the reduced
graph (`/tmp/abcgo-gia-giarf-typed-red.log`).

Both simulation routines now share the same logical-fanin word evaluator:
AND, XOR, bitwise MUX selection and buffer forwarding preserve phases and
control dependencies. The reduced-level builder dispatches on gate kind and
uses the already-ported source HashXor/HashMux AND expansions, leaving the
SAT graph in ordinary-AND form. Level filtering, proof-gated temporary
substitution, root collection and source Value mappings remain unchanged.
The typed simulation matrix checks 3,072 node-bit results across two rounds,
all eight input phases and both simulation modes; reduced truth and the full
equivalence/choice path add analytical checks for both SAT backends.

A fresh standalone giaGiarf.c fixture verifies the ordinary-AND domain. It
matches all seven simulation words, split counts, representative/next tables
and head/tail removal results in 32 repeated-round snapshots. No simulation
or class algorithms are stubbed, and no ABC object archive is required. The
fixture assigns CI words explicitly; it does not claim equivalence with C's
Hcd_ManSimulationInit, which writes CI ordinal indexes rather than object IDs.
Go already uses actual CI object IDs and zero-initializes constant storage.

The refinement/removal routines now reuse the manager's existing old/new
scratch vectors, matching C's scratch lifetime, rather than allocating two
new vectors per split/removal. Recursive splitting completes its writes before
the next recursive split reuses the storage; the native snapshots exercise
that recursion and reuse. Removed the unused vClassTemp field/allocation;
no algorithm reads it. Focused typed/native tests pass in 0.185s. The full
uncached GIA race suite passes in 17.391s and vet passes, with logs
`/tmp/abcgo-gia-giarf-typed-final-race.log` and
`/tmp/abcgo-gia-giarf-typed-final-vet.log`. All 133 module packages and their
tests also compile after the private HCD record change
(`/tmp/abcgo-gia-giarf-layout-module-build.log`, exit zero). This compile gate
does not rerun unrelated test bodies.

Remaining review includes broader class rehashing/pattern-packing boundaries,
SAT-manager reuse, equivalence metadata/transfers and the other GIA algorithm
families. Original completion remains 2/131; this does not close the GIA audit.
In particular, giarfSolver currently reconstructs the CEC manager per query
and adds a synthetic AND for two assumptions, although the current CEC core
already exposes CheckConjunction. Check its incremental-query contract before
reusing one manager per reduced level, as C does.

Giarf SAT-lifetime review confirmed that core.Manager supports incremental
CheckLiteral/CheckConjunction queries, per-query conflict limits, validated
witnesses and automatic recycling. giarfSolver now initializes that manager
once per reduced level, reuses it across queries and releases it at close.
Two-literal queries use CheckConjunction directly instead of adding a virtual
AND and snapshotting the whole graph again. TAS ownership is unchanged. This
is a lifetime/compactness improvement to the existing adapter, not a solver
replacement or a claim that its backend matches the original CEC solver.

An exhaustive three-input query sequence checks 840 unary/conjunction queries
across both backends, with constants, complements, repeated assumptions and
alternating SAT/UNSAT results. Every status is compared against all eight
assignments; every returned witness is independently checked for valid CI
ordinals and satisfaction, and UNSAT results must not retain the prior model.
These tests pass before and after the optimization; no artificial failing
layout test was added for a performance-only change. Existing rare-output,
highest-level and exhausted/positive-budget regressions also pass.

The warmed 64-input benchmark, 500 queries in each run, measured 187,731 to
19,063 ns/query, 118,054 to 22,397 bytes/query and 1,159 to 10 allocations/query.
These are local microbenchmark results, not a whole-application speed claim.
Logs are `/tmp/abcgo-gia-giarf-manager-before.log` and
`/tmp/abcgo-gia-giarf-manager-after.log`. The full uncached GIA race suite
passes in 17.715s and vet passes (`/tmp/abcgo-gia-giarf-manager-final-race.log`
and `/tmp/abcgo-gia-giarf-manager-final-vet.log`).

Important remaining source-fidelity work: the current giarfSolver `mini` path
uses proof/cec/core.Manager, whose implementation is Satoko-backed. The
source-shaped Cec_ManSatCheckNode/Two implementation in proof/cec instead
uses cecSourceSAT and the ported BSAT solver, including source frontier,
supergate and MUX CNF construction. Giarf cannot import that parent directly
because proof/cec imports GIA. Reusing the current manager does not resolve
that backend/encoding mismatch. Before closing this family, share the existing
source CNF/BSAT implementation through a dependency-neutral component, keep
the source-shaped parent wrapper thin, and wire Giarf to the shared source
engine without an import cycle or another duplicated/invented solver. Preserve
the existing exact-clause/variable-layout CEC tests and add Giarf source query
comparisons, recycling, budget and witness coverage. This remains pending;
GIA and the full sequential audit are still active at 2/131 completed originals.

Source-BSAT integration, first checkpoint: extracted the existing source CEC
CNF builder into proof/cec/core.SourceCNF, with callbacks for live GIA node
metadata and recognized MUX pins. GIA supplies those callbacks through
SourceCNFGraph; the source-shaped CEC wrapper delegates to the shared builder.
The original frontier order, variable-zero/constant-one-slot convention,
supergate boundaries, polarity handling and MUX clauses are retained. There
is one encoder implementation, not a duplicated GIA solver or replacement
Tseitin encoding. Existing tests read the same state through borrowed
Solver/IDs accessors; their expected values and native records are unchanged.

Focused source/native CEC tests pass in 8.729s. The full uncached race suites
for GIA, CEC and CEC core pass in 17.675s, 41.382s and 1.262s respectively
(`/tmp/abcgo-shared-cec-cnf-race.log`). These adjacent-package checks support
this shared-code extraction and do not count either CEC package as audited.
Giarf still uses its old adapter at this checkpoint. Its new clause-count
regression fails with seven clauses versus the source supergate's four
(`/tmp/abcgo-giarf-source-backend-red.log`); it will accompany the subsequent
BSAT wiring change rather than being included in this green extraction commit.

Source-BSAT integration, second checkpoint: Giarf's selected CEC path now
uses the shared source CNF builder and the ported BSAT solver directly. The
Satoko-backed generic CEC manager is no longer used for these queries. Solver
state remains incremental per reduced level; default 2000-variable/200-call
strict recycling boundaries and disabled polarity flipping match the CEC
parameters used in giaGiarf.c. SAT/UNSAT/unknown statuses and learned negated
assumption clauses follow the source query path. Returned CI assignments are
independently replayed before becoming refinement patterns. TAS is unchanged.
The shared encoder remains the single implementation used by the source-shaped
parent CEC wrapper; this does not add another CNF algorithm or SAT substitute.

The red seven-versus-four clause regression now passes, including the exact
source frontier IDs and omission of the supergate's internal node. A compiled
original-CEC fixture with Giarf's parameters checks ten mixed unary/pair
queries through a supergate and recognized MUX: statuses, active variable
counts, clause counts, call-since-recycle counters and every node's SAT ID
all match. Its native records are collected directly from the source manager;
no SAT/CNF functions are stubbed. The source oracle archive builder is shared
with the existing choice-miter oracle rather than copied into another test.
The native source query check passes in 1.777s
(`/tmp/abcgo-giarf-source-native-verified.log`).

Additional source-boundary coverage runs 202 queries on a 2000-input cone:
recycling occurs only when both strict default limits have been crossed, and
full witnesses remain valid afterward. The 840 cross-query status/witness
matrix, rare outputs, typed gates, highest-level proofs and positive/exhausted
budget cases all pass on the actual source backend. The warmed 500-query
microbenchmark now measures 4,239 ns/query, 1,848 bytes/query and 12 allocations
(`/tmp/abcgo-giarf-source-manager-benchmark.log`), without a claim about complete
application runtime.

Full uncached race suites pass for GIA (19.991s), CEC (41.165s) and CEC core
(1.310s), and vet passes for all three. Logs are
`/tmp/abcgo-giarf-source-backend-final-race.log` and
`/tmp/abcgo-giarf-source-backend-vet.log`. All 133 module packages and their
tests compile (`/tmp/abcgo-shared-source-cnf-module-build.log`, exit zero);
unrelated test bodies were not rerun by that compile gate.

The specific Giarf source-backend mismatch recorded above is resolved. Broader
CEC packages still await their own sequential audits; these adjacent changes
do not count them complete. GIA class rehashing/pattern boundaries, metadata
and transfer algorithms and the remaining algorithm families are still
pending. Original completion remains 2/131 and the full goal stays active.

Giarf pattern/class boundary review: compared packing and refinement with
giaGiarf.c. Added tests for all 32 single-word bit positions, partial literal
counts, repeated and empty cubes, and conflict rejection before any simulation
or presence write. Thirty-two distinct complete five-input assignments verify
that exactly 31 non-phase slots are available, exhaustion leaves data unchanged,
and bit zero remains reserved. Direct pattern insertion checks bit positions
0, 1 and 31 with interleaved CI/internal object IDs and untouched unspecified
inputs/internal nodes.

Class refinement coverage deliberately constructs distinct nonconstant words
with the same initial hash bucket. An independent partition oracle groups
objects by their accumulated normalized simulation history over four rounds;
it checks every representative and next link, complement equivalence,
constant candidates leaving the constant class, recursive splitting, and no
merging when later signatures coincide. All tests pass without production
changes; these checks found no additional bug. The unused related-class
collector's treatment of head roots matches the original source and was not
changed speculatively.

The full uncached GIA race suite passes in 19.853s
(`/tmp/abcgo-giarf-packing-classes-race.log`), and GIA vet passes. GIA remains
under audit; metadata/transfer algorithms and remaining families still need
review. Original package completion remains 2/131.

Mapped hashing review: Gia_ManDupHashMapping copied every internal object as
an AND, silently changing real XOR and MUX functions. Added a failing signed
gate regression before repair (`/tmp/abcgo-mapping-hash-typed-red.log`). The
copy pass now dispatches logical gates through the existing AND/XOR/MUX hash
operations and forwards buffers. The source giaIf.c routine assumes ordinary
AND graphs; that path and its mapping-copy policy remain intact. Real gates
are normalized to ordinary hashed ANDs, using existing primitives rather than
an alternative mapping algorithm.

Coverage checks 64 gate/polarity combinations over all eight input assignments
(512 rows), preserving terminal/register counts, root complement and copied
LUT truth. Another 32 signed XOR/buffer/MUX/AND chains over 16 assignments
(512 rows) check dependent cut truth against actual mapped leaf values and
output equivalence. The existing identical-LUT merge regression also passes.
Mapping/packing transfer was compared with giaIf.c; its existing source-shaped
logic is unchanged and needs additional remapping/dangling-record coverage.

Full uncached race suites pass for GIA (20.235s), map/if (4.478s), and the
base/abci caller (29.464s). The shared log is
`/tmp/abcgo-mapping-hash-review-race.log`. Vet passes for GIA and map/if.
Expanded vet also reports existing base/abci self-assignment, unkeyed literals
and unreachable-code diagnostics (`/tmp/abcgo-mapping-hash-review-vet.log`);
no base/abci files changed in this repair. These diagnostics remain for that
package's sequential audit. Original completion remains 2/131.

Mapping/packing transfer follow-up: added a compiled-original giaIf.c oracle
comparison with reordered CIs, reversed LUT construction order, six retained
LUTs, one dangling LUT, redundant dangling cut leaves, both signs of MUX record
trailers, replacement of existing destination mapping, and packing groups of
sizes one, two and three. Complete destination mapping and packing vectors
match the native source. Independent expectations also check fanin order,
pruned counts, signed trailer remapping, group order, unchanged source records
and independently owned destination storage. This found no additional defect;
production transfer logic is unchanged. Focused native coverage passes in
1.789s (`/tmp/abcgo-mapping-transfer-source.log`). Full uncached GIA race tests
pass in 21.721s (`/tmp/abcgo-mapping-transfer-race.log`). The package remains
under audit; original completion is still 2/131.

Representative metadata compactness: Gia_Rpr_t used a uint32 for each of
four one-bit flags, making every representative record 20 bytes. Store those
flags as uint8 while retaining the existing named fields and uint32 object
ID/sentinel. The record now measures 8 bytes on this host, a 60% reduction
without bit-packing accessors or an algorithm change. AIGER encoding/decoding
uses explicit conversions at its uint32 boundary; its serialized format is
unchanged. The existing native C extension decoder and golden/round-trip tests
pass. New coverage exercises all 16 flag combinations, independent clearing,
representative sentinel preservation, value copying, and adjacent records;
the same test measured 20 bytes before and 8 bytes after the change.

Focused tests pass (`/tmp/abcgo-gia-compact-reprs-focused.log`), as do full
uncached GIA race tests (21.984s,
`/tmp/abcgo-gia-compact-reprs-race.log`) and GIA vet
(`/tmp/abcgo-gia-compact-reprs-vet.log`). All 133 module packages and their
tests compile (`/tmp/abcgo-gia-compact-reprs-module-build.log`); this compile
gate does not run unrelated test bodies. Further GIA algorithm families and
representative/name remapping still need audit. Completion remains 2/131.

Representative/name remapping follow-up: reviewed the shared boundary used by
normalization, marked duplication and CEC graph rebuilding. Added an independent
connectivity-closure oracle over 2,500 object maps across four partitions,
including constants, singletons, dropped class heads, collapsed objects and
reordered IDs. Every destination representative and sorted next link matches;
source metadata remains unchanged. Invalid map lengths/IDs and representative
storage/order leave the destination storage and contents intact. Another 256
input/output name-map combinations cover missing names, omitted entries and
reordering; output-validation failure cannot partially install input names,
and copied/exposed name storage is independent. These checks found no bug in
the existing remapping implementation. The MiniLUT/MiniAIG name mapper was
also compared by inspection with giaMini.c; broader polarity-specific coverage
of that mapper is still pending.

Focused remapping tests pass. The full uncached GIA race suite passes in
23.667s (`/tmp/abcgo-identity-mini-review-race.log`) and GIA vet passes
(`/tmp/abcgo-identity-mini-review-vet.log`). GIA remains under audit; original
completion is still 2/131.

MiniAIG export review: Gia_ManToMiniAig copied real XORs and MUXes as ANDs,
changing exported functions. Added the failing regression before the fix
(`/tmp/abcgo-mini-aig-export-red.log`). The exporter now dispatches logical
gates through MiniAIG's existing AND/XOR/MUX constructors and forwards buffers.
The original C giaMini.c routine assumes ordinary AND graphs; its ordinary
AND path, input/output ordering, source Value updates and register count are
preserved. This uses the existing expansion primitives, not a substitute
export algorithm.

The regression checks 32 signed XOR/buffer/MUX chains on all eight input
assignments (256 rows), comparing all three round-trip outputs against direct
Boolean expectations and checking graph/interface/register validity. A compiled
original-C fixture checks eight signed ordinary AND graphs: complete MiniAIG
fanin arrays and every source object's exported Value match exactly. Focused
coverage passes in 1.820s (`/tmp/abcgo-identity-mini-review-focused.log`). Full
uncached race suites pass for GIA (23.667s) and the base/abci exporter caller
(28.848s), in `/tmp/abcgo-identity-mini-review-race.log`; GIA vet also passes.
GIA export/import and other remaining algorithm families still need review.
Original package completion remains 2/131; the sequential goal stays active.

MiniLUT export/equivalence follow-up: added independent interpretation of
exported MiniLUT records over all assignments at widths 2, 5, 6, 7, 8, 10 and
16, each with uniform versus mixed root-output polarity (268,176 assignment
rows). Tests cover cut-pin reordering beyond six variables, uint32/uint64 truth
boundaries, complemented internal cut leaves, constants, register/interface
counts, and shared root/PI inverters. A compiled giaMini.c fixture receives the
equivalent ordinary AND expansion of the Go real-XOR graph. Digests include
every exported pin and truth word; all 14 cases match original C, including
16-input tables. Focused native coverage passes in 1.859s
(`/tmp/abcgo-mini-lut-wide-native.log`).

MiniAIG reduction/miter coverage checks four signed duplicate-AND networks
with two registers that have identical reset and transition functions. Across
all 16 four-cycle PI traces (256 frames), reduction removes the redundant
register while preserving four POs and the next-state function. The miter
retains two registers and two equivalence assertions, both false on every
reachable frame. Source giaMini.c reduction/miter flow was compared by
inspection; no bug was found in these paths. Full uncached GIA race tests
pass in 26.089s (`/tmp/abcgo-mini-lut-equiv-review-race.log`) and GIA vet passes.
Other MiniAIG helper paths remain under review; original completion is 2/131.

MiniAIG triples generator: found a source-order mismatch with independent
output cones. Go emitted a PO immediately after visiting each output cone,
then appended logic for later outputs; original giaMini.c first computes every
cone and then emits the PO suffix. A red regression catches the intervening
logic node (`/tmp/abcgo-mini-generate-order-red.log`). Moved PO creation to a
final pass, preserving source ID order and existing input/cycle validation.
No truth function changes or replacement generator were introduced.

A compiled original Gia_MiniAigGenerateFromFile fixture checks complete binary
MiniAIG files for independent outputs, shared out-of-order definitions and
repeated fanins. All three now match byte-for-byte. Focused tests pass in
1.735s (`/tmp/abcgo-mini-generate-order-focused.log`), the full uncached GIA
race suite passes in 27.870s (`/tmp/abcgo-mini-generate-order-race.log`), and
GIA vet passes (`/tmp/abcgo-mini-generate-order-vet.log`). GIA's remaining
algorithm families still need audit; original completion remains 2/131.

Mini name/equivalence mapping follow-up: added a 64-case matrix combining
eight AND input-polarity patterns with eight map-literal polarity patterns.
Equivalent three-input cones associate their ANDs differently and use distinct
combined object IDs. The matrix covers constant zero/one, omitted entries,
unmatched classes, objects without a combined copy, duplicate equivalent
members, and CO correspondence. Independent evaluation checks each returned
name and equivalence on all eight input assignments (512 assignment rows).
Explicit expectations enforce the source's last-name preference versus
first-member equivalence preference, with the appropriate complement phases.

A compiled original giaMini.c fixture runs both Gia_ManMapMiniLut2MiniAig
and Gia_ManMapEquivAfterScorr for all 64 cases; complete returned vectors
match. The diagnostic name verifier was also compared with its C source.
No production defect was found in these paths. Focused native tests pass in
1.806s (`/tmp/abcgo-mini-name-equiv-focused.log`), full uncached GIA race tests
pass in 29.634s (`/tmp/abcgo-mini-name-equiv-race.log`), and GIA vet passes.
Remaining GIA algorithm families still await review; original completion is
2/131 and the sequential audit remains active.

Placement metadata compactness: Gia_Plc_t used four uint32 fields for two
15-bit coordinates and two one-bit flags, occupying 16 bytes per object.
Store coordinates as uint16 and flags as uint8, ordered to avoid padding;
the Go record now measures 6 bytes (62.5% smaller). Named fields and their
valid domains are unchanged. Embedding can directly copy its uint16 coordinate
array; placement totals still accumulate in uint32. AIGER encoding explicitly
widens every field before shifting, retaining the source's 32-bit packed wire
format independently of Go's struct layout. Existing overflow rejection is
retained, including the otherwise representable coordinate value 0x8000.

Before/after tests measure 16/6 bytes and check all 36 combinations of
coordinate boundary values 0, 1, 0x7fff and the two flag bits through complete
AIGER round trips. Raw placement bytes and decoded fields match. Separate
tests reject out-of-range coordinates and flags. Existing AIGER native decoder,
normalization and embedding/placement tests pass. Focused tests pass in 0.121s
(`/tmp/abcgo-placement-metadata-focused.log`); full uncached GIA race tests
pass in 29.758s (`/tmp/abcgo-placement-metadata-race.log`), and GIA vet passes.
All 133 module packages/tests compile
(`/tmp/abcgo-placement-metadata-module-build.log`); unrelated test bodies were
not run by that gate. Full placement algorithms and other remaining GIA
families still need review. Original completion remains 2/131.

Embedding/FORCE graph review: both simple builders and the shared supernode
collector followed only the two stored data pins of real MUXes. Their separate
control edge was lost, disconnecting control-only logic from placement. Added
failing embedding and FORCE regressions before fixing
(`/tmp/abcgo-placement-mux-red.log`). A shared placement-fanin helper retains
the original two-pin ordering/multiplicity and adds a real MUX's control pin.
Both simple constructors and clustered cone collection use it. The original
C builders assume ordinary AND graphs; the ordinary path remains intact.

Regressions check simple and clustered embedding/FORCE graphs with a MUX
controlled by an otherwise-unused AND cone. Expected root fanins, reachability
of all four data/control inputs, and valid FORCE refinement permutations now
pass. A compiled original giaForce.c fixture checks an ordinary AND network
with a collapsible supernode over all six orders of three outputs, with and
without clustering. Forward/reverse fanin cross cuts, complete DFS positions,
the selected forward/reverse cut, and all positions after three refinement
iterations match source C in every case. This does not claim native agreement
for the real-MUX omission, which is repaired in Go.

Inspection also found that C's simple FORCE builder sets fCi=0 on CI objects;
Go already correctly labels them as CIs. That existing correction is retained.
Native comparisons cover cross-cut/placement behavior, which is unaffected
because those input objects have no fanins; they do not claim raw flag equality.
FORCE register feedback remains disabled as in the original source, while
embedding retains its source register connections.

Focused source/regression tests pass in 1.899s
(`/tmp/abcgo-placement-control-focused.log`). Full uncached GIA race tests
pass in 31.475s (`/tmp/abcgo-placement-control-race.log`) and GIA vet passes
(`/tmp/abcgo-placement-control-vet.log`). Distance/eigensolver/placement
boundaries and remaining GIA families still need review. Original completion
remains 2/131 and the sequential goal stays active.

Embedding distance/covariance review: discovered that freshly allocated
dimension vectors used zero for unreachable vertices. Covariance already
filters infinity sentinels, but those zeros were centered as real observations,
adding false covariance from unused CIs. Two red regressions demonstrate the
missing sentinel and a diagonal sum of 14 instead of the reachable population's
10 (`/tmp/abcgo-embedding-unreachable-red.log`). Initialize distance storage
to embInfinity before BFS writes reached vertices. This restores the intended
initialization shown commented out in giaEmbed.c; no BFS or covariance
algorithm was replaced. Original C shares the initialization defect.

Regressions cover disconnected output components joined through the embedding
constant, an unused CI, and all 16 covariance cells computed independently from
reachable raw moments. Unreachable dimensions remain sentinel-valued until
centering converts them to zero. A compiled original giaEmbed.c fixture checks
single and paired BFS seeds for every vertex in simple/clustered ordinary AND
graphs with and without register feedback: reached counts, level counts,
farthest-node tie selection and every distance match source. The native
comparison intentionally does not assert the original faulty covariance.

Focused tests pass in 1.873s (`/tmp/abcgo-embedding-distances-focused.log`).
Full uncached GIA race tests pass in 33.447s
(`/tmp/abcgo-embedding-distances-race.log`), and GIA vet passes. Repeated-seed
and zero-norm numerical boundaries are still being examined. Remaining GIA
families and the full sequential goal remain open; original completion is 2/131.

Embedding boundary follow-up: red tests found that repeated BFS seeds counted
four unique reachable vertices as six, and normalizing a zero vector produced
NaNs that propagated through a zero covariance eigensolver and projection
(`/tmp/abcgo-embedding-boundaries-red.log`). Deduplicate the initial frontier
using the existing traversal marks. Preserve zero vectors instead of dividing
by a zero norm. The BFS, power iteration and projection algorithms remain the
source algorithms; no replacement eigensolver or layout policy was introduced.

Additional red tests exposed float32 square underflow, subnormal precision loss
and overflow in normalization: very small vectors remained unnormalized or
changed direction scale, while large finite vectors collapsed to zero. Logs
are `/tmp/abcgo-embedding-normalization-range-red.log` and
`/tmp/abcgo-embedding-normalization-accumulation-red.log`. Keep the source's
float32 products for ordinary vectors; recompute their norm with float64
products when any nonzero term underflows/is subnormal or the sum overflows.
This also handles accumulated subnormal squares whose total is normal-sized.
Only an exactly zero recomputed norm returns without normalization. The
original C shares these duplicate-seed and numerical defects; Go intentionally
corrects those boundaries while retaining ordinary-case arithmetic.

Coverage includes repeated multisource seeds, zero-vector suffix preservation,
zero covariance through finite eigenvectors/projection/placement, six finite
magnitude scales from the smallest float32 to the largest, 2048 accumulated
subnormal squares, and a real constant-output graph through the complete
rank-deficient embedding pipeline. Unique-seed compiled-C BFS comparisons and
existing nonzero matrix tests still pass. Full uncached GIA race tests pass in
33.439s (`/tmp/abcgo-embedding-boundaries-race.log`) and GIA vet passes
(`/tmp/abcgo-embedding-boundaries-vet.log`). Remaining solver/placement
algorithms and other GIA families still need audit; completion remains 2/131.

Embedding projection review: the orthogonalization helper recomputed the
unchanged dot product once per coordinate, making the operation quadratic.
When the destination aliases either input, earlier writes also changed later
dot products and the result was not orthogonal. Red alias tests demonstrate
nonzero residuals for parallel vectors and incorrect general projection
residuals (`/tmp/abcgo-embedding-orthogonal-red.log`). Compute the projection
once before writing components. This makes the helper linear and supports
either full input alias without allocation. The source formula is unchanged;
original C shares the aliasing defect and redundant work. Production power
iteration already uses separate buffers, preserving its arithmetic/results.

A compiled original giaEmbed.c fixture compares all 672 float32 component
bits for 24 disjoint-vector projections; all match. Focused native/alias tests
pass in 1.866s (`/tmp/abcgo-embedding-orthogonal-focused.log`). The 1000-call
microbenchmark changes 32/128/512-element projection times from approximately
569/10518/190414 ns to 33/261/596 ns, with zero allocations in both versions
(`/tmp/abcgo-embedding-orthogonal-before.log`,
`/tmp/abcgo-embedding-orthogonal-after.log`); no whole-program speedup is claimed.
Full uncached GIA race tests pass in 37.243s
(`/tmp/abcgo-embedding-projection-placement-race.log`) and GIA vet passes.
The package remains under audit; original completion is 2/131.

Placement scaling review: zero-valued projection coordinates were skipped
during affine scaling. With both positive and negative coordinates this put
zero below negative values, changing the intended rank order. A constant
nonzero axis also computed zero times infinity, replacing finite solutions
with NaNs. Two red tests demonstrate incorrect placement order and nonfinite
flat axes (`/tmp/abcgo-embedding-placement-order-red.log`). Scale every
coordinate, using a zero scale for flat axes. The source's sorting and rank-grid
placement remain intact. Original C shares both boundary defects; Go corrects
them rather than preserving erroneous output.

New tests check exact coordinates for shuffled negative/zero/positive axes and
finite, valid placements for flat nonzero axes. A compiled original giaEmbed.c
fixture checks ordinary positive axes, repeated coordinate values, both sort
size regimes, simple/clustered graph construction and register feedback.
Every coordinate and HPWL value matches before and after three refinement
steps. Focused tests pass in 1.893s
(`/tmp/abcgo-embedding-placement-focused.log`). The full uncached GIA race
suite passes in 37.243s (`/tmp/abcgo-embedding-projection-placement-race.log`)
and GIA vet passes (`/tmp/abcgo-embedding-projection-placement-vet.log`).

Next inspection target: the embedding constant's reserved fanin slots can
make an isolated constant qualify as a random pivot. That path still needs
a regression and repair. Iterative eigenvector accuracy and remaining GIA
algorithm families also remain under review. Original completion is 2/131.

Embedding pivot review: regressions confirm that the constant's reserved
fanin slots made it eligible for random selection even when isolated from
every output. Some starting states therefore embedded only one vertex instead
of the four output-connected vertices. Wire-only graphs similarly selected
the wrong component; an empty graph never terminated because no vertex passed
the rejection sampler. Red evidence is `/tmp/abcgo-embedding-pivot-red.log`;
the empty-graph regression runs in a bounded child process and captured the
five-second deadline instead of hanging the suite.

Preserve the source rejection sampler but restrict eligible nodes to the
output-connected traversal marks left by component discovery. Mark the
constant when joining multiple components so normal bridge selection remains
eligible. If no eligible internal node exists, seed BFS from the first output,
or from the constant for an empty graph. These are terminating base cases for
the existing distance algorithm, not a replacement layout algorithm. Original
C shares the isolated-constant/empty-graph defects.

Tests cover 128 starting-state/simple-or-clustered combinations with an unused
CI and isolated constant, joined wire-only components, and wire/empty graphs
through dimensions, covariance, eigenvectors, projection and finite placement.
A compiled giaEmbed.c baseline covers 16 ordinary connected cases with varied
random starting states: centered distance dimensions, all covariance entries,
the principal eigenvector and projected coordinates agree within 1e-5. Those
cases retain original pivot sampling behavior; defective boundary cases are
checked against their output-reachability invariants instead of C's failure.

Focused tests pass in 1.907s (`/tmp/abcgo-embedding-pivot-focused.log`), full
uncached GIA race tests pass in 41.107s
(`/tmp/abcgo-embedding-pivot-race.log`), and GIA vet passes
(`/tmp/abcgo-embedding-pivot-vet.log`). Iterative eigenvector accuracy and
remaining GIA algorithm families still need review. Completion remains 2/131.

Embedding eigenvector review: found dominant-direction leakage after covariance
multiplication. The source projected the input iterate away from prior axes,
but multiplication could reintroduce them because those prior eigenvectors
are approximate. A red diagonal-covariance regression gives secondary variance
394658 instead of 1 for diag(1e9, 1, 0.1)
(`/tmp/abcgo-embedding-eigen-red.log`). A compiled original giaEmbed.c witness
independently reproduces secondary variance 394658.026; corrected Go gives
0.999997861 while retaining the original principal eigenvector.

Project the multiplied iterate through the existing orthogonalization helper
before normalization. Also avoid normalizing deflation remnants at the scale
of float32 roundoff: their norm is compared relative to the pre-projection
norm, using dimension times float32 epsilon. A rotated rank-deficient red test
showed a null axis becoming an almost exact duplicate of the principal axis
(`/tmp/abcgo-embedding-eigen-spectra-red.log`); such numerical remnants now
remain zero. Share the robust float64 squared-norm calculation with vector
normalization. This retains source power iteration, random initialization,
100-iteration limit and 0.999 convergence criterion; it does not introduce a
replacement eigensolver. Original C shares the demonstrated deflation defect.

Coverage includes 12 large-gap cases retaining genuine secondary variance
down to 1e-20, twenty rotated known-spectrum cases with repeated/zero eigenvalues
and deficient rank, and a dense 100-dimensional rank-two covariance. Tests
require finite/unit nonzero directions, orthogonality, and the expected
variance spectrum at tolerances consistent with the retained approximate
iteration. Existing zero-covariance and complete rank-deficient graph pipelines
remain green. The original-C principal-vector witness and ordinary connected
dimension/covariance/projection baseline also pass.

Full uncached GIA race tests pass in 42.952s
(`/tmp/abcgo-embedding-eigen-race.log`). Final strengthened spectral assertions
and the additional 100-dimensional case pass the focused race check in 2.983s
(`/tmp/abcgo-embedding-eigen-final-race.log`); final focused native output is
in `/tmp/abcgo-embedding-eigen-final-focused.log`. GIA vet passes
(`/tmp/abcgo-embedding-eigen-vet.log`). Remaining embedding integration/output
helpers and other GIA algorithm families still need review. Original package
completion remains 2/131 and the sequential audit remains active.

Embedding integration/output review: the final placement transfer assumed
GIA object IDs matched embedding vertex order. C explicitly assumes normalized
ordering there, but the Go graph API supports interleaved terminals and logic.
An end-to-end red regression shows transferred coordinates disagreeing with
the plotted source edges for interleaved graphs
(`/tmp/abcgo-embedding-transfer-red.log`). Transfer now follows each GIA
object's Value handle to its embedding vertex and placement index. Normalized
graphs retain the identity mapping; no placement algorithm was changed.

The regression compares two independent outputs of the complete driver:
plotted geometry versus metadata attached to original GIA object IDs. Eight
normalized/interleaved, simple/clustered and unrefined/refined cases check all
logical arcs and register feedback. Tests write plots into temporary working
directories without invoking gnuplot. Focused coverage passes
(`/tmp/abcgo-embedding-transfer-focused.log`).

Compactness: plot preparation now uses a single contiguous bitmap with row
views, matching C's storage arrangement, instead of 500 independently allocated
rows. The raster/run algorithm and sentinel column are unchanged; existing
boundary/duplicate-pixel plot-run coverage passes. A 100-call benchmark with
256 vertices reduces allocations from 503 to 4 per call. Both versions use
about 364 KB and take about 0.48 ms per call on this host; this is a GC-object
reduction, not a claimed byte or whole-program runtime reduction. Logs are
`/tmp/abcgo-embedding-plot-before.log` and
`/tmp/abcgo-embedding-plot-after.log`.

The remaining embedding output/diagnostic routines were compared with
giaEmbed.c; their source reporting and file-format conventions are retained.
Full uncached GIA race tests pass in 43.683s
(`/tmp/abcgo-embedding-output-race.log`), and GIA vet passes
(`/tmp/abcgo-embedding-output-vet.log`). Other GIA algorithm families still
need review. Original package completion remains 2/131; the full sequential
audit stays active.

Cofactor and quantification review: two red regressions exposed MUX composition
failures. Passing the same graph for both branches cleared the CI scratch
mapping before reuse and panicked (`/tmp/abcgo-cofactor-mux-alias-red.log`).
Initialize positive-branch CI mappings from the destination CIs instead.
Distinct branch graphs also inherited negative-branch name/spec metadata,
contrary to C's explicit positive-branch identity choice
(`/tmp/abcgo-quantification-expanded-red.log`). Copy positive identity.
The source C scratch algorithm also assumes distinct branch graphs; the Go
repair handles aliasing without introducing a separate composition algorithm.

Coverage exercises real XOR/MUX gates, buffers, eight polarity combinations,
all input assignments, variable and guarded internal-node cofactors, both
existential entry points, and combinational/sequential interface metadata.
Universal abstraction retains the source convention: combine only output zero,
and use the positive cofactor for auxiliary outputs. The compiled original-C
fixture compares 328 cases and 5,248 input rows against equivalent ordinary-AND
expansions, including register/constraint counts and positive branch identity
(`/tmp/abcgo-quantification-native.log`). Its explicit Exist2 declaration is
needed because that C function is defined but absent from gia.h.

Full uncached race tests pass for GIA (45.637s) and its base/acb caller (5.197s)
(`/tmp/abcgo-quantification-review-race.log`). GIA vet passes
(`/tmp/abcgo-quantification-review-vet.log`). Other duplication algorithms,
including block reduction, still need review. Original completion remains
2/131; the sequential audit remains active.

Block-reduction review: compared Gia_ManDupBlock with giaDup.c. Its input
zeroing, output selection, hashing/cleanup and integer register-count reduction
match the source. No production defect or useful simplification was found.
Added exhaustive independent truth-table coverage for blocks 1/2/3/6, eight
polarity choices, real XOR/MUX gates, buffers, constants, interleaved logic/COs,
and aligned combinational/sequential interfaces. Tests also check identity,
PI/PO/register counts, empty graphs, nonpositive block sizes, and indivisible
CI/CO counts. They do not invent behavior for unaligned register blocks.

A separate compiled-original-C fixture uses equivalent ordinary-AND gates;
all 96 cases and 1,872 assignment rows match. Focused race tests pass in 2.984s
(`/tmp/abcgo-block-review-race.log`); the subsequently added empty/invalid
interface cases pass a separate focused race run in 1.020s. GIA vet passes
(`/tmp/abcgo-block-review-vet.log`). Production code was unchanged after the
preceding full GIA/base-acb race run. Remaining GIA families still require
review; original package completion remains 2/131 and the goal stays active.

Cascade-composition review: a minimal red test passes the same combinational
graph as both stages and exposes an unmapped-fanin panic
(`/tmp/abcgo-compose-alias-red.log`). Clearing stage two's scratch values
also erased stage one's connecting output mappings. Snapshot those literals
through the existing output-copy helper before resetting stage two. Retain
the source hashing, interface checks, identity and cleanup algorithm.

Independent truth tables cover four self-cascade polarity variants and 64
pairs of distinct typed XOR/MUX/buffer graphs, including direct/complemented
edges, interleaved outputs and dangling logic. The latter compare both with
evaluation of the stages in sequence and with compiled original C on equivalent
ordinary-AND graphs: all 1,024 assignment rows match. Tests require first-stage
identity, unchanged interface counts and no dangling output graph. Independent
Go coverage runs even when native-oracle prerequisites are absent.

The initial focused tests pass in 1.856s
(`/tmp/abcgo-cascade-focused.log`). After separating the independent/native
tests and strengthening cleanup checks, the full uncached GIA race suite
passes in 49.906s (`/tmp/abcgo-cascade-review-race.log`); GIA vet passes
(`/tmp/abcgo-cascade-review-vet.log`). Append, shared/sequential composition
and property/constraint families still need expanded review coverage.
Original completion remains 2/131; the sequential audit remains active.

Sequential append review: a new minimal red regression exposed shared scratch
state when Gia_ManDupAppendNew receives the same sequential graph twice.
For PI=0 and first/second register values 1/0, the first copy incorrectly uses
the second register bank (`/tmp/abcgo-append-independent-red.log`). Duplicate
the second source through the existing Gia_ManDup only for this aliased,
sequential case. This gives each bank its own Value mapping without changing
the common distinct-source path or the source composition algorithm. Source C
also reuses one scratch mapping for aliased arguments; the native comparison
below deliberately checks distinct graphs, not agreement with that defect.

Expanded independent transition-table tests cover all eight input/gate
polarity combinations in either graph and zero/one/two registers per graph,
including unequal register counts. They check shared PIs, independent state
banks, PO-before-RI output ordering, and first-source identity. All 576 cases
and 12,544 input/state rows also match compiled original C on equivalent
ordinary-AND graphs. Separate alias coverage checks 24 typed cases and 672
input/state rows, plus the minimal regression. XOR/MUX gates, buffers, direct
complemented outputs and interleaved CO/logic layouts are included. Native
prerequisites do not gate the independent Go tests. Test helper duplication
was removed, and alias cases run once rather than once per unrelated pair.
The initial expanded fixture used the same object as MUX control and an arm
for one-register graphs; it was corrected to respect the real-MUX constructor's
distinct-fanin precondition before comparing results.

Focused tests pass in 1.950s (`/tmp/abcgo-append-review-focused.log`). Full
uncached GIA race and vet verification is recorded below. Other append,
property and constraint routines still need expanded review coverage; original
package completion remains 2/131 and the sequential audit remains active.
Full uncached GIA race tests pass in 52.317s
(`/tmp/abcgo-append-review-race.log`), and GIA vet passes
(`/tmp/abcgo-append-review-vet.log`).

Append/shared-input composition review: compared Gia_ManDup2,
Gia_ManDupAppend and Gia_ManDupAppendShare with giaDup.c. Added independent
truth tables and original-C comparisons across eight polarity choices in
either graph, zero/one/two registers, and three composition modes. All 576
cases and 25,088 input rows match. Tests preserve the precise source contracts:
Dup2/shared append concatenate raw COs and retain the first/destination register
count; ordinary append makes all terminals combinational. They do not silently
reinterpret these routines as the independent-state merge of AppendNew.
Coverage checks destination identity, hash lifetime, already-active/new hash
tables and exact object preservation in unhashed append. An additional 24
cases (640 rows) interleave CIs, logic and COs to catch ordinal/object-ID mixups.

The first native fixture used physical buffer nodes; source HashStart rejects
such nodes and also requires structurally unique ANDs. The corrected C baseline
uses hashed ordinary-AND expansions and transparent buffers, while the Go
fixtures retain real XOR/MUX/buffer nodes. Initial corrected focused tests
pass in 1.985s (`/tmp/abcgo-append-shared-focused.log`); expanded focused race
coverage passes in 4.176s (`/tmp/abcgo-append-shared-race.log`). No production
change was needed for those distinct-source comparisons.

Inspection then identified a shared-self-append termination defect: the CO
loop repeatedly reads the growing source count when destination is source.
A bounded child-process red regression reproduces nontermination without
hanging the main test runner (`/tmp/abcgo-shared-self-append-red.log`). Snapshot
the original CO count before appending. This follows the existing composition
algorithm and adds no source clone. The C routine's source-object iteration
also assumes distinct arguments; native comparisons intentionally avoid that
upstream alias defect. Added self-append checks for a direct complemented wire
and 24 typed combinational/sequential graphs (224 input/state rows), preserving
raw CO ordering and register counts. A child-local GC memory limit and watchdog
bound future regressions; the final watchdog allows two seconds of scheduling
headroom. Full focused tests pass in 1.974s before that headroom adjustment
(`/tmp/abcgo-append-shared-final-focused.log`). Final race/vet results follow.

Property/constraint composition and other GIA families remain open. Original
package completion is still 2/131; the sequential audit remains active.
Final full uncached GIA race tests pass in 70.657s
(`/tmp/abcgo-append-shared-final-race.log`); GIA vet passes
(`/tmp/abcgo-append-shared-final-vet.log`). Raw-MUX arm-phase handling during
unhashed append with an active table remains an inspection follow-up, along
with property/constraint composition; this checkpoint does not close GIA.

Raw MUX phase follow-up: a new red regression confirms that ordinary append
panics when copying a valid real MUX with a complemented high data arm into
its active-hash destination (`/tmp/abcgo-append-mux-phase-red.log`). DFS copying
passed that arm directly to AppendMuxReal, whose source-compatible active-hash
precondition requires a regular then arm. In the active-table case, order the
mapped data arms, invert the control if swapped, and move the high arm's phase
to both data edges and the returned literal. This is the same Boolean phase
identity used by the existing source hash-MUX routine; it performs no hash
lookup, simplification, gate expansion or change in node count. Copying into
an inactive-table destination retains the existing raw phase behavior.

Coverage checks all sixteen data/control/output phase combinations, parent
AND use of the copied MUX and repeated memoized DFS calls. Reversed source-to-
destination leaf mappings exercise the data-arm swap with both active and
inactive tables. Tests require one real MUX and exact source object counts.
Independent truth tables cover 256 append rows and 256 reversed-leaf DFS rows.
A compiled-original-C ordinary-AND expansion matches all sixteen append cases
and their 256 rows. Go real-MUX handling is an extension of that original
AND-only append driver; no unsupported native real-MUX behavior is asserted.
Focused tests pass in 1.858s (`/tmp/abcgo-append-mux-phase-focused.log`).
Final GIA/caller race and GIA vet verification follows below. This closes the
specific phase follow-up from the prior checkpoint; property/constraint and
other GIA families remain open. Original completion remains 2/131.
Full uncached race tests pass for GIA (59.249s) and its base/abci caller
(28.582s), recorded in `/tmp/abcgo-append-mux-phase-race.log`. GIA vet passes
(`/tmp/abcgo-append-mux-phase-vet.log`). The sequential audit stays active.

Property/constraint composition review: compared Gia_ManDupWithNewPo and
Gia_ManDupWithConstraints directly with giaDup.c. No production defect or
useful simplification was found. New independent transition-table coverage
checks that property inputs are current register outputs, not next-state
functions; replacing properties preserves all RIs and drops old constraint
metadata, following C. Constraint conversion checks every PO classification
mask: ordinary outputs first, complemented constraints next, original RIs
last, with the replacement constraint count. Constraint conversion retains
source object counts and both drivers retain identity and normalized layout.

Typed XOR/MUX/buffer graphs include interleaved outputs, three-register
transition relations, two-output properties, constants, and combinational/
all-register interfaces. Across 648 cases and 41,472 input/state rows, both
independent evaluation and the compiled original-C ordinary-AND expansion
agree. Physical buffers are transparent in the C fixture because its
AppendAnd-based constraint driver requires distinct fanin objects; they remain
physical buffers in the Go tests. Additional cases cover empty managers,
constant properties, a property with zero outputs but live transition logic,
and six invalid property/type interfaces. Tests check both insufficient and
excess type counts and reject negative/nonbinary classifications.

The initial focused run passes in 2.056s
(`/tmp/abcgo-property-constraint-focused.log`). Final focused uncached race
coverage, including the added boundary and interface checks, passes in 4.955s
(`/tmp/abcgo-property-constraint-race.log`); GIA vet passes
(`/tmp/abcgo-property-constraint-vet.log`). No production code changed after
the previous full GIA/base-abci race run. This closes expanded review coverage
for the property/constraint functions in dup_compose.go, not the remaining
GIA duplication or other algorithm families. Original completion stays 2/131;
the sequential audit remains active.

Sequential hold/invariant/cone composition review: compared Gia_ManDupSelf,
Gia_ManDupInvMiter and Gia_ManDupAppendCones against giaDup.c. A new minimal
red regression passes the source itself as an additional combinational cone
and exposes an unmapped-fanin panic
(`/tmp/abcgo-append-cones-alias-red.log`). The Go code cleared the cone's Value
scratch and then read CI mappings back from the aliased source. C reads them
from the destination CI list instead. Restore that exact mapping with the
existing Gia_ManCiLit helper, avoiding a source clone or a second algorithm.

Expanded independent transition tables and the compiled original-C equivalent
AND graphs agree for 464 cases and 29,760 input/state rows. They cover all
eight gate polarity combinations in source and invariant/cone, register hold
control, I(state) AND NOT I(nextState), state-only versus complete-CI cone
inputs, zero/one/two appended cones, repeated cone pointers, and truncated
cone lists with a trailing unused nil entry. Unlike the earlier upstream
alias examples, the aliased-source cone cases also match original C directly.
Tests check interface/register/constraint counts, source identity and removal
of dangling logic. Go graphs retain real XOR/MUX gates and physical buffers;
the C baseline uses ordinary-AND expansions and transparent buffers.

Additional tests compare eight-cycle hold traces for 64 sequences and eight
polarity variants (4,096 frames) with the independently stepped source and
verify both constant invariants over all 64 input/state assignments. Focused
coverage initially passes in 2.020s
(`/tmp/abcgo-sequential-composition-focused.log`); trace/constant additions
pass in 0.007s (`/tmp/abcgo-sequential-composition-boundaries.log`). The native
fixture explicitly declares InvMiter because gia.h lacks its prototype.

Full GIA race and BMC caller verification and the final focused check after
simplifying the destination lookup to Gia_ManCiLit are recorded below.
Flop-class extraction, repeated duplication and other GIA families still need
review. Original package completion remains 2/131; the audit remains active.
The full uncached GIA race suite passes in 66.229s
(`/tmp/abcgo-sequential-composition-race.log`); its BMC caller passes race tests
in 25.161s (`/tmp/abcgo-sequential-composition-bmc-race.log`). The final focused
race run after the equivalent CI-literal helper simplification passes all
new cases in 4.530s
(`/tmp/abcgo-sequential-composition-final-focused-race.log`). Final GIA vet
passes (`/tmp/abcgo-sequential-composition-final-vet.log`).

Repeated duplication/flop-class review: a new minimal red test shows
Gia_ManDupTimes changes the input interface of a DFS-ordered graph with an
unused first PI (`/tmp/abcgo-dup-times-order-red.log`). It collected terminals
in physical object order instead of source interface order. Record terminal
mappings from source PI/RO/PO/RI vectors after each copy, preserving their
public order before grouping all primary terminals ahead of register ones.
Physical graph copying and its one-object-per-source-object contract remain
unchanged. Output-vector remapping has a separate regression. A compiled
original-C witness also reproduces the wrong DFS-order output; normal-layout
comparisons remain source-exact rather than endorsing that upstream defect.

Temporary terminal storage now consists of two exactly sized arrays with
PI/RO and PO/RI views. This replaces four growing slices and final
concatenations, mirroring source pre-sizing without extra join allocations.
The 100-call, 32-copy benchmark changes 459 to 447 allocations per call and
about 45.1 KB to 43.3 KB. Logs are `/tmp/abcgo-dup-times-before.log` and
`/tmp/abcgo-dup-times-after.log`. One short timing sample is not a whole-program
speed claim.

Boundary inspection found that both Times and FlopClass lose simple mode,
causing a valid r AND NOT r source node to panic in raw copying. Separate
red tests precede that fix (`/tmp/abcgo-dup-times-simple-red.log` and
`/tmp/abcgo-dup-class-simple-red.log`). Preserve fGiaSimple in Times and in
the shared raw-copy initializer giaDupMappedStart; the same raw-copy contract
also applies to its permutation, constraint, trim and related callers.
Regression coverage explicitly checks times, class, permutation and constraint
copies. No simplification or buffer reinterpretation replaces the retained
opposite-phase AND node.

Independent truth tables exercise 72 repetition cases with one/two/three
copies, zero/one/two registers, and real XOR/MUX/buffer gates, preserving object
counts and CI/CO ordinals. A further 72 DFS-layout cases cover 40,288 rows.
Flop-class extraction covers all 27 assignments of three signed class labels,
four selections including an absent class, and eight gate polarity patterns:
864 cases and 55,296 rows. It preserves stable nonselected/selected register
ordering and correctly exposes nonselected registers as primary terminals.
All 936 normal-layout cases (95,584 rows) match compiled original C using
ordinary-AND, buffer-transparent equivalents. Class extraction itself otherwise
matches the source; its new production change is simple-mode propagation.

Empty managers, constant-only outputs, autonomous registers, no-output graphs
with dangling logic, invalid duplication counts/class vectors and manually
reordered COs have additional coverage. Final focused tests pass in 2.546s
(`/tmp/abcgo-dup-times-class-final-focused.log`); GIA vet passes
(`/tmp/abcgo-dup-times-class-vet.log`). Every current module package, including
tests, compiles (`/tmp/abcgo-dup-times-class-build.log`). These compile checks
do not count other packages as audited. Final GIA/base-abci race results follow.
Other GIA duplication families remain open; original completion remains 2/131.
Full uncached race tests pass for GIA (72.616s) and its base/abci caller
(28.860s), recorded in `/tmp/abcgo-dup-times-class-race.log`. The allocation
benchmark was repeated on the final source in `/tmp/abcgo-dup-times-after.log`.
The sequential audit remains active.

PI/register permutation and gap review: read DupPerm, DupPermFlop,
DupSpreadFlop, DupPermFlopGap, PermuteInputs and DupZero against giaDup.c.
Their valid-input behavior matches original C. Added independent transition
tables for all six three-input/register permutations, all ten placements of
three registers among five slots, all gap/permutation combinations, every
valid regular/extra/pseudo-PI group split, and zero duplication, across eight
gate polarity patterns. Spread coverage includes arbitrary positive and
negative non--1 slot values: those are placeholders, not register indices.
Gap duplication instead uses an actual permutation. Inserted state bits are
ignored by the original logic and receive zero next-state values.

All 744 cases and 155,136 input/state rows match compiled original C on
buffer-transparent ordinary-AND equivalents. The same 744 cases pass with
DFS-ordered Go graphs, giving another 155,136 rows. Tests retain real XOR/MUX
and physical buffers, check interface/gate/object counts, stable register
pairing, and source identity. DupZero deliberately preserves only the source
name, as C does, and omits all internal logic. Direct-wire PI outputs give
additional observable interface-order coverage, avoiding graph symmetries.
Empty permutations, gap-only state with zero original registers, invalid
permutations/gaps/groups, and negative/out-of-range indices have separate cases.

Bounds inspection found signed overflow in PermuteInputs' ppis+extra check.
Two MaxInt32 counts wrap past the guard and enter enormous vector-building
loops. A bounded child-process red test reproduces nontermination
(`/tmp/abcgo-permutation-overflow-red.log`). Validate ppis against n first,
then extra against n-ppis, preserving all valid group behavior while avoiding
addition overflow. This repairs the Go argument guard; invalid arguments have
no supported C baseline. The child-local memory limit/watchdog keeps the red
regression bounded and does not affect the main test process.

The initial focused run passes in 2.939s
(`/tmp/abcgo-permutation-gap-focused.log`). Final expanded focused coverage,
including the repair and signed Spread slots, passes in 2.961s
(`/tmp/abcgo-permutation-gap-final-focused.log`). GIA vet passes
(`/tmp/abcgo-permutation-gap-vet.log`); final full GIA race verification follows.
No source algorithm substitution or useful additional compaction was needed.
Input removal, onset/offset, randomized edge mutation and other GIA families
remain open. Original package completion remains 2/131; the audit stays active.
Full uncached GIA race tests pass in 85.047s
(`/tmp/abcgo-permutation-gap-race.log`).

Input removal/onset-offset review: compared DupLastPis, DupRemovePis and
DupOnsetOffset with giaDup.c and their source callers. LastPis is explicitly
combinational and retains the suffix. C RemovePis keeps the first CI-count
minus removed-count inputs; its actual acb callers are combinational, where
that means removing the PI suffix. The existing Go extension retains register
pairs when removing trailing PIs from a sequential graph; a separate direct
transition test checks it, without claiming a sequential C baseline. No new
extension was introduced in this review.

OnsetOffset pairs every raw CO with its complement, including register-input
terminals, and retains the old register count. Preserve that source convention
rather than inventing a two-machine sequential interpretation. Bmc_FxCompute
uses the paired circuit as combinational CNF over all CIs. Reuse the existing
giaMappedEdge helper instead of duplicating the fanin lookup and phase XOR;
this is the only production change in this checkpoint.

Independent truth tables and compiled original C agree on 136 cases and
3,568 input/state rows: all retained-input counts zero through six, prefix/
suffix removal, eight gate polarities, and onset/offset with zero/three/six
registers. The same cases pass on DFS-ordered Go graphs. Tests preserve real
XOR/MUX/buffer counts and identity, check constant outputs and paired CO order,
and use buffer-transparent ordinary-AND graphs for the native baseline.
Additional cases cover the existing sequential RemovePis behavior, simple-mode
opposite-phase ANDs, negative/excess counts, direct uses of removed inputs,
uses confined to dangling logic, and LastPis' sequential rejection. Dropped
inputs must be unused throughout the raw copied logic; no implicit pruning
algorithm was added to relax that source precondition.

The initial focused tests pass in 1.816s (`/tmp/abcgo-input-dual-focused.log`).
After the helper simplification and added rejection coverage, focused tests
pass in 1.827s (`/tmp/abcgo-input-dual-final-focused.log`). GIA vet passes
(`/tmp/abcgo-input-dual-vet.log`); full GIA/BMC race results follow below.
No correctness defect was found in these routines. Randomized CI-edge
mutation and other GIA duplication families still need review; original
package completion remains 2/131 and the sequential audit stays active.
Full uncached race tests pass for GIA (86.846s) and its BMC caller (25.693s),
recorded in `/tmp/abcgo-input-dual-race.log`.

Randomized CI-edge mutation review: compared Gia_ManDupPiPerm directly with
its giaDup.c experiment and the &permute command caller. It is not a shared PI
bijection: each copied gate edge gets an independent draw, the eligibility
range includes constants and fanins simplified to CIs, and direct CO edges
are not randomized. Both draws are consumed for every ordinary AND, even
when neither copied edge is eligible. The source resets its RNG on entry and
includes register-output CIs in that range. Preserve these counterintuitive
source choices rather than replacing them with a different permutation.

New independent truth tables and compiled original C agree for 100 cases and
1,546 input/state rows. Thirty-two phase combinations and zero/two/four
registers exercise interleaved COs, duplicate ANDs, copied internal fanins,
hashing to a CI/constant, and constant resampling. Comparisons include exact
object counts, complete outputs and the next RNG word. Tests perturb caller
RNG state before mutation and require unchanged identity/interface, stopped
hashing and no dangling logic. Fixed source goldens remain available without
a C compiler, including the transformation of a constant-zero AND into a
nonconstant function and untouched direct CI outputs.

Separate independent Go checks cover all eight polarity variants of the
existing logical-pin XOR/MUX/buffer extension (128 rows) and its exact draw
count. They do not claim native C supports real gates in this AND-only
experiment. A no-CI physical buffer/opposite-phase AND case checks the existing
Go zero-CI fallback and correct copying of literals collapsed by earlier gates;
C would divide by zero for such internal gates. This review introduces no new
extension. No correctness defect was found in the mutation routine.

Compactness: combine copied-literal lookup into the local randomize helper,
removing a per-loop read closure and nested read wrappers without changing
draw order or gate construction. Initial focused tests pass in 1.810s; expanded
independent tables pass in 1.817s; final compact-helper/no-CI coverage passes in
1.784s. Logs are `/tmp/abcgo-random-mutation-focused.log`,
`/tmp/abcgo-random-mutation-final-focused.log`, and
`/tmp/abcgo-random-mutation-compact-focused.log`.

Full uncached race tests pass for GIA (90.558s) and its base/abci caller
(28.710s), recorded in `/tmp/abcgo-random-mutation-race.log`. GIA vet passes
(`/tmp/abcgo-random-mutation-vet.log`). Together with the preceding checkpoints,
this closes the current dup_permutation.go family review, not GIA as a whole.
A core follow-up identified during inspection remains: check raw MUX constant
controls whose canonical control tag becomes zero, the same value used for
no-MUX metadata. Add a red truth-table regression before any constructor
repair. Other duplication/core families remain open; original completion is
still 2/131 and the sequential audit stays active.

Raw-MUX constant-control follow-up: a minimal red truth-table test confirms
that AppendMuxReal(0, highData, lowData) returns false for lowData=true,
highData=false (`/tmp/abcgo-mux-constant-control-red.log`). Sorting the data
pins may also turn an original constant-true control into the same zero tag.
Zero is the companion table's no-MUX marker, so these nodes silently become
ordinary ANDs instead of selecting the lower data input.

After the existing validation and pin ordering, represent canonical control
zero by the existing AppendBuf primitive on the selected edge. This still
allocates exactly one node and preserves every subsequent object ID, without
new type flags or a change in table/wire encoding. All other representable
MUX cases retain their prior representation. Buffer/MUX counts now agree
with the actual stored kind, and complemented selected edges are retained.
The fix is an exact Boolean identity, not a replacement graph algorithm.

The source-header witness reproduces eight legacy failures among 32 raw
control/data-phase cases; the other 24 cases agree unchanged. Go constructor
coverage checks 48 constant/variable, data-order, phase and hash-state cases
(288 assignments) through the original graph, raw clone, both normalization
modes and cleanup. Five invalid-constructor tests confirm that the buffer
case does not bypass storage, fanin-distinctness, fanin-validity or active-hash
phase preconditions and does not allocate on rejection. Focused tests pass
in 1.809s, then 1.825s with those additional rejection checks, in
`/tmp/abcgo-mux-constant-control-focused.log` and
`/tmp/abcgo-mux-constant-control-final-focused.log`.

Full uncached GIA race tests pass in 92.400s
(`/tmp/abcgo-mux-constant-control-race.log`). The complete module test command
`go test -count=1 ./...` passes across all 133 current packages (130 tested,
three without tests), recorded in
`/tmp/abcgo-mux-constant-control-integration.log`. This integration result is
not a code-review completion claim for those other packages. GIA vet passes
(`/tmp/abcgo-mux-constant-control-vet.log`).

A final 64-case DFS integration check maps a source control CI to each
constant, reverses leaf mappings and covers all data phases with and without
hashing. It verifies exact one-node allocation and 256 assignment rows against
the original graph with the control fixed. All new raw-MUX cases, including
that subsequently added check, pass the focused race run in 2.869s
(`/tmp/abcgo-mux-constant-control-final-focused-race.log`); production code was
unchanged after the full-module run. This closes the core follow-up flagged
in the previous checkpoint. Other GIA families remain open; original package
completion stays 2/131 and the sequential audit remains active.

Support-sliced duplication review: two new red regressions expose deviations
from giaDup.c. The Go result retained the original specification string where
C returns name only, and preexisting source fMark0/fMark1 bits remained set
where C clears both scratch marks before/after slicing
(`/tmp/abcgo-sliced-contract-red.log`). Clear the result specification and
restore the source scratch-mark algorithm with deferred cleanup. This also
removes two per-object Boolean arrays; no new selection algorithm is added.

Independent Boolean checks and compiled original C agree across 336 cases
and 5,376 input/state rows: sixteen edge polarities, zero/two/four source
registers and support limits -1 through 5. They compare copied/exposed node
counts, complete positive-cut output ordering, name/spec conventions and
source mark cleanup. Cases include complemented original COs (cut outputs
are regular nodes, following C), CI/constant original outputs, unexposed
copied dangling logic, and excluded parents exposing eligible child cones.
The result remains combinational over every original CI, as in C.

Additional independent Go coverage checks 56 typed XOR/MUX/buffer cases
(3,584 rows), the otherwise easy-to-miss control cone of an excluded real MUX
(eight cases/128 rows), empty graphs, a zero-support constant buffer and a
simple-mode opposite-phase AND. Native comparisons use actual ordinary-AND
source graphs, without asserting C's AND-only slicer supports real gates or
physical buffers. Initial expanded tests pass in 1.824s; final boundary
coverage passes in 1.802s (`/tmp/abcgo-sliced-expanded-focused.log` and
`/tmp/abcgo-sliced-final-focused.log`). Minimal repairs pass in
`/tmp/abcgo-sliced-contract-focused.log`.

Compactness measurements: in the 100-call/256-node benchmark, restoring the
existing mark fields changes 1,038 allocations and 27,408 bytes per call to
1,036 allocations and 26,832 bytes. A final source-style direct CO-fanin
mark replaces the generic allocated fanin-list loop, reducing allocation
calls to 1,035 with the same measured 26,832 bytes. Logs are
`/tmp/abcgo-sliced-storage-before.log`,
`/tmp/abcgo-sliced-storage-after.log` and
`/tmp/abcgo-sliced-storage-final.log`. These are storage/allocation findings;
one short timing sample is not a general runtime claim.

Full uncached race tests pass for GIA (93.854s) and its base/abci caller
(28.607s), in `/tmp/abcgo-sliced-review-race.log`. After the final equivalent
single-fanin shortcut, all new sliced cases pass focused race verification
in 2.996s (`/tmp/abcgo-sliced-final-focused-race.log`), and final GIA vet passes
(`/tmp/abcgo-sliced-final-vet.log`). Levelized, explicit-vector and all-node-
output duplication still need expanded review; the remaining GIA families
are also open. Original package completion remains 2/131 and the sequential
audit remains active.

Levelized duplication and DFS-copy follow-up: three separately red regressions
show lost constraint classification in both full-copy APIs and rejection of a
valid simple-mode opposite-phase AND in DFS copying. Preserve nConstrs in
both results and fGiaSimple in the DFS result. Red logs are
`/tmp/abcgo-levelized-constraints-red.log`,
`/tmp/abcgo-dfs-constraints-red.log` and
`/tmp/abcgo-dfs-simple-red.log`. The compiled original C witness reports
source/levelized/DFS constraint counts `1 0 0`: constraint loss is an inherited
C omission deliberately repaired here because unchanged output ordering must
retain property/constraint roles. This does not change the metadata policy
for filtered/derived-output copy APIs.

Restore the original C stable level-counting algorithm in place of comparison
sorting, using starting offsets that also support level-zero buffers. Ordinary
AND copies exactly match compiled C across 96 cases and 1,536 assignment
rows, including source-to-destination internal literal maps, CI/CO/register
counts and complete output truth tables. Independent Boolean expectations,
exact expected internal ordering and mapped logical pin checks cover real
XOR/MUX/buffer nodes as well. Across ordinary/typed graphs, original/DFS
layout, constraint counts zero/one, 32 edge polarities and zero/two/four
registers, the expanded checks cover 768 cases and 12,288 assignment rows.
Same-level buffer chains, interleaved COs, unused internal gates, identity,
empty/wire-only graphs, constant buffer chains and simple-mode AND storage
are included. Typed cases do not claim native AND-only levelization supports
real gates. The fixture is `aig/gia/test_vectors/levelized.c`.

In the 4,096-node/16-level benchmark (100 calls), comparison sorting measured
1,692,672 ns, 343,296 bytes and 4,396 allocations per call; level counting
measured 820,847 ns, 343,372 bytes and 4,395 allocations. Logs are
`/tmp/abcgo-levelized-before.log` and `/tmp/abcgo-levelized-after.log`.
The algorithm is O(nodes + levels), with a counter per level; this particular
run uses slightly more memory, and neither timing nor footprint is a general
claim for deep graphs. New tests pass focused race verification in 3.119s
(`/tmp/abcgo-levelized-final-focused-race.log`), and GIA vet passes
(`/tmp/abcgo-ordered-copy-review-vet.log`).

Full uncached race tests pass for GIA (95.074s), SSC (7.675s) and ABCI
(28.674s), recorded in `/tmp/abcgo-ordered-copy-review-race.log`. The final
focused race run additionally checks both graph representations against
independent Boolean formulas. This closes levelized ordering and the named
DFS metadata follow-ups; explicit-vector/all-node-output duplication and
other GIA families remain open. Original package completion stays 2/131;
the sequential audit remains active.

Explicit-vector and all-node-output duplication review: inspect giaDup.c's
requested-order contract and the iso_canonical.go caller. Unlike a full copy,
Gia_ManDupFromVecs defines a new interface with the caller's register count;
Gia_ManDupOuts inserts positive internal-node outputs between original POs
and RIs. Both intentionally produce nConstrs zero in C. Do not apply the
previous full-copy metadata repair to these derived-output APIs.

New tests compare 3,936 cases and 62,976 assignment rows with compiled C
(`aig/gia/test_vectors/copy_vectors_outputs.c`): 32 polarities, six CI
permutations, original/level gate orders, complete/reversed/repeated/empty CO
selections and all feasible zero/two/four requested registers, followed by
all-node-output copies with those source register counts. Tests independently
check Boolean formulas, requested CI/gate mappings, normalized layout, exact
object/interface counts, name/spec and reset constraint classification.
The complete matrix is also exercised on real XOR/MUX/buffer graphs and on
DFS layouts: 15,744 cases and 251,904 assignment rows across four variants.
DFS cases account for changed physical order and removal of dangling gates;
non-DFS all-node copies explicitly retain and expose dangling logic.

Additional coverage verifies a two-CI subset with omitted gates, repeated
outputs and requested registers zero/one/two; stale scratch mappings are
cleared. Eighteen malformed selection/register cases must panic, including
missing dependencies, non-topological or repeated gates, duplicate CIs,
wrong object kinds and invalid IDs. Empty graphs, constants, direct wires
and simple-mode opposite-phase ANDs cover boundary storage/semantics.
Native malformed-input behavior is not invoked or asserted. All expanded
checks pass in 2.348s (`/tmp/abcgo-selection-expanded.log`). The broad focused
regex also ran existing explicit-reachability tests; their deliberately
reported counterexample/limit diagnostics do not indicate a test failure.

No new behavior defect was found in these two APIs. For compactness,
giaDupMappedNodes now delegates to the existing giaAppendMappedNode instead
of repeating its typed gate dispatch and mapping validation. This removes
21 net lines while keeping mapped AND/XOR/MUX/buffer handling in one place.
GIA vet passes (`/tmp/abcgo-selection-review-vet.log`).

After consolidating the helper, full uncached race tests pass for GIA
(101.785s) and ABCI (28.961s), recorded in
`/tmp/abcgo-selection-review-race.log`. This closes the explicit-vector and
all-node-output review and completes the reviewed functions in dup_level.go.
Other GIA families remain open; original package completion remains 2/131
and the sequential audit remains active.

Ordered-copy review follow-up: separate red tests confirm lost constraint
classification in reverse DFS, AIGER-order copying and register abstraction,
and valid simple-mode AND rejection in reverse/AIGER/abstraction/group/
selected/presence-vector copying. Red logs are
`/tmp/abcgo-ordered-constraints-red.log` and
`/tmp/abcgo-ordered-simple-red.log`; the latter catches panics per subtest so
all six failing paths are observed before the repair. Use giaDupMappedStart
for common initialization, retaining identity, constant mapping, mux support
and simple construction mode. Preserve nConstrs in reverse/AIGER copies and
abstraction, whose original primary outputs and their roles are unchanged.
Selected-output APIs retain their derived-interface constraint policy.
Compiled original C reports `1 0 0 0 0` for source/DFS/reverse/AIGER/abstraction
constraint counts: these repairs correct inherited metadata omissions.

AIGER copying now follows C's direct topological loop using the existing
mapped-node/output helpers, instead of routing every gate through DFS. It
still retains dangling nodes. The shared initializer and direct loop remove
38 net production lines without changing the traversal algorithms for DFS
or abstraction.

The three full ordered-copy methods are compared exactly against C across
288 cases and 9,216 assignment rows: 32 polarities, zero/two/four registers,
an unused CI appended after all COs, and interleaved gates/outputs. Compare
complete object signatures, fanin literals, terminal ordinals, source Value
mappings, interface counts and output Boolean formulas. The same checks
run with real XOR/MUX/buffer nodes, original/DFS layouts and constraint
counts zero/one: 2,304 cases and 73,728 rows. DFS removes dangling internal
logic while AIGER copying preserves it. CI/CO vectors retain original order,
and AIGER output is normalized. Typed cases use independent Boolean formulas;
they do not claim the C AND-only routines support real gates.

Nine further cases (72 rows) compare representative/next arrays with C,
covering constant candidates, sequential interfaces, reverse traversal that
changes the lowest-ID class head, and independent destination storage.
Boundary cases cover empty/constant-only graphs, unused inputs, constant
buffers and a register-only graph. The fixture is
`aig/gia/test_vectors/ordered_copies.c`. Expanded native/Boolean/class checks
pass in 1.952s (`/tmp/abcgo-ordered-final-focused.log`); the fifteen boundary
cases were added subsequently and receive the final focused race check.
GIA vet passes (`/tmp/abcgo-ordered-review-vet.log`).

Full uncached race tests pass for GIA (104.973s), PDR (15.475s) and ABCI
(28.828s), in `/tmp/abcgo-ordered-review-race.log`. The final focused race
run, including all fifteen subsequently added boundaries, passes in 4.130s
(`/tmp/abcgo-ordered-final-focused-race.log`); production code was unchanged
between those checks. This closes the full DFS/reverse/AIGER copy review.
Abstraction mapping validation and general output-selection behavior still
need expanded review beyond the repaired construction/constraint metadata;
other GIA families remain open. Original package completion stays 2/131 and
the sequential audit remains active.

Register-abstraction mapping review: the Go inverse-map walk accepted any
negative integer as a retained-register marker, while the C precondition and
PDR's producers require exactly -1. New red tests reproduce acceptance of -2
and MinInt32 (`/tmp/abcgo-abs-inverse-red.log`). Reject values below -1 in the
existing inverse-map pass; valid mapping behavior is unchanged. This is a
malformed-input validation repair, not a discovered valid-circuit truth bug.
The existing forward/inverse consistency checks remain in place.

Enumerate all sixteen ordered subsets of three registers (empty, three
singletons, six pairs and six triples). For every subset, check original PIs,
promoted ROs in caller-specified order, retained ROs in original order, all
original POs and exactly the retained RIs. Thirty-two polarity combinations
and all 32 PI/state assignments give 512 native cases and 16,384 assignment
rows exactly matching compiled C. Independent Boolean formulas check every
property and next-state output, including input permutation after promotion.
An RI-only cone is removed iff its register is promoted, and unrelated
dangling logic is always removed. Exact copied gate counts verify pruning;
source CI/RI Value mappings verify retained and discarded interface objects.

Repeat that matrix on real XOR/MUX/buffer representations, original/DFS
layouts and constraint counts zero/one: 4,096 cases and 131,072 rows, checking
name/spec, normalized layout and PI/RO/PO/RI classification. These exhaustive
one-step checks verify the transition relation for every state/input value;
no replacement abstraction algorithm is introduced. The native fixture is
`aig/gia/test_vectors/register_abstraction.c` and uses original AND-only C.

Twelve more malformed mapping cases reject wrong vector lengths, bad or
repeated register IDs, mismatched inverse ordinals and missing entries.
Boundary checks include empty and constant-only graphs, an unused PI and all
sixteen promotions of a three-register graph with no primary inputs or
outputs. Constant-only constraints retain their role. Native assertion
failures are not deliberately launched for malformed input. Expanded checks
pass in 1.964s (`/tmp/abcgo-abs-final-focused.log`); native-only verification
passes in 1.829s (`/tmp/abcgo-abs-native.log`). The implementation already
shares the selected root list between cone copying and output construction,
so no further compactness rewrite is warranted in this routine.

Full uncached race tests pass for GIA (108.369s) and PDR (15.103s), recorded
in `/tmp/abcgo-abs-review-race.log`; GIA vet passes
(`/tmp/abcgo-abs-review-vet.log`). This closes register-abstraction mapping
review. General output selection and other GIA families remain open.
Original package completion stays 2/131; the sequential audit remains active.

Output-selection review: a new red test shows Gia_ManDupSelectedOutputs was
writing source CO Value mappings even though C intentionally leaves all of
them unset (`/tmp/abcgo-output-selection-red.log`). Gia_ManDupOutputVec has
the opposite C postcondition: chosen source COs must map to their copied COs.
Keep the shared selected-cone implementation, leave explicit-list source COs
unset, and perform the small CO-map pass only for the presence-vector API.
This repairs source scratch bookkeeping; no valid-output truth discrepancy
was found. A separate red test exposes the missing C precondition that an
explicit selection cannot contain more entries than the source has outputs
(`/tmp/abcgo-output-selection-count-red.log`). Restore that check while
retaining duplicate/reordered selections within the original bound.

Compiled original C and independent Boolean/dependency expectations agree
across 8,928 cases and 226,368 assignment rows. Coverage includes all 28
contiguous ranges of six COs with zero/two/four source registers, every one
of 64 output masks in forward/reverse explicit order and as a presence
vector, plus repeated and reordered lists. Presence flags alternate positive
and negative nonzero values. Compare exact source CI and CO mappings,
first-discovery input order for ranges, original input order for list/mask
selection, complete output truth and exact copied gate/interface counts.

Range copying keeps only reachable CIs and returns a combinational graph,
even when its selected source COs include RIs. Explicit/presence selection
retains all original CIs, including the unused one created after source COs.
Empty ranges therefore have zero CIs, while empty list/mask selections keep
all five. All derived interfaces reset constraints, following C. Static
logical dependency masks verify shared-cone copying and unused gate pruning
without using the copying algorithm as the expected result.

Repeat the matrix for real XOR/MUX/buffer graphs and DFS layouts: 28,224
cases and 688,512 rows across the four variants. A separate control-only
AND cone driving a real MUX is preserved through every API (24 cases/384
rows). Eighteen empty/unused-input/constant-only selection cases and nine
additional invalid range, ID, mask-size and sequential-interface cases pass.
The fixture is `aig/gia/test_vectors/output_selection.c`. Native-only tests
pass in 2.155s (`/tmp/abcgo-output-selection-native.log`); the final expanded
checks pass in 3.216s (`/tmp/abcgo-output-selection-expanded.log`). No native
assertion is deliberately triggered for invalid arguments. There is no
additional compactness change: sharing cone construction remains appropriate,
but sharing source-CO side effects had caused the regression above.

Full uncached race tests pass for GIA (120.622s), CEC (41.515s) and ABCI
(28.817s), in `/tmp/abcgo-output-selection-review-race.log`; GIA vet passes
(`/tmp/abcgo-output-selection-review-vet.log`). This closes the range,
explicit-list and presence-vector output-selection review. Literal/terminal
remapping helpers and other GIA families remain open; original package
completion stays 2/131 and the sequential audit remains active.

Literal/terminal/equivalence remapping review: inspect Gia_ManDupRemapLiterals,
Gia_ManDupRemapCis/Cos and Gia_ManDupRemapEquiv against giaDup.c and their
WLC/CBA, ordered-copy, sweep-domain and input-unshuffle callers. No production
defect was found. Literal remapping correctly preserves negative placeholders,
emits -1 for either polarity of a dropped object and XORs source/copied edge
phases. Terminal remapping restores vectors and CIO ordinals without changing
the actual edges, source Values or destination identity/register metadata.
Keep the two short terminal loops explicit: their diagnostics distinguish
input and output contract failures, and another generic helper would add
indirection without a useful reduction.

Compiled C agrees across 343 literal-map combinations (3,773 scalar entries),
1,728 independent input/output terminal permutations (27,648 complete
assignment rows) and 128 representative mappings. Literal cases include
constant collapse, complemented copied edges, aliasing destinations, dropped
objects and negative markers -1/-7/MinInt32. Source Value storage is checked
against the original map after remapping. Terminal cases cover all 24x24
permutations with zero/two/four registers and check vector IDs, CIO ordinals,
output truth, unchanged source mappings and unrelated destination metadata.
Repeat the terminal matrix on source DFS layouts for another 27,648 rows.

Representative cases cover two disjoint classes, multiple constant candidates,
rotated/reversed destination IDs, four mapped-literal phase patterns and an
unmapped unclassified object. Expected groups and sorted next chains are
computed independently; destination representative storage must not alias the
source. Five independent Go checks additionally exercise the existing
TransferRepresentativesFrom support for dropped class heads, surviving
singletons, collapsed members/classes and constant collapse. Those robustness
cases do not claim C's class walk supports every dropped/collapsed mapping;
its assertion/invalid indexing is not deliberately invoked.

Boundary/rejection checks cover empty graphs/vectors, bad literal indices,
terminal count mismatch, absent/complemented/out-of-range/nonterminal target
mappings, absent source metadata preserving existing destination classes,
preallocated destination class storage, malformed source representatives and
invalid mapped IDs. The fixture is
`aig/gia/test_vectors/duplication_remapping.c`. Initial native and independent
checks pass in 1.960s (`/tmp/abcgo-remap-native-expanded.log`). This is a
test-only checkpoint; production code is unchanged from the preceding full
GIA/CEC/ABCI race gate. It closes the remaining reviewed helpers and completes
dup_order.go. Other DFS duplication variants and GIA families remain open;
original package completion remains 2/131 and the sequential audit is active.

Final focused uncached race tests pass in 3.835s
(`/tmp/abcgo-remap-final-race.log`); final GIA vet passes
(`/tmp/abcgo-remap-final-vet.log`). No further broad rerun is needed for this
test-only addition after the immediately preceding full caller race gate.

DFS/full-representative-copy review: six new red subtests reproduce rejection
of a valid simple-mode opposite-phase AND in Dfs, Dfs2, OnePo, Skip, Node and
LitArray wrappers. A separate red regression shows Dfs2 drops constraint
classification (`/tmp/abcgo-dfs-metadata-red.log`). Use giaDupMappedStart for
those six constructors and retain nConstrs in the full Dfs2 copy. This removes
35 net production lines while restoring construction metadata. The existing
Dfs constraint and counterexample policies remain unchanged. Original C's
source/Dfs/Dfs2 constraint witness is `1 1 0`, confirming the Dfs2 omission is
inherited; the Go repair retains the roles of the unchanged primary outputs.

Compiled C matches 192 full-copy cases and 6,144 input/state rows: both Dfs
and Dfs2, 32 edge polarities, zero/two/four registers and an unused CI appended
after the COs. Compare complete source Value mappings, interface/object
counts and independent Boolean formulas. Dfs puts CIs first and is normalized;
Dfs2 discovers CIs through DFS and restores their interface order afterward.
Repeat with typed XOR/MUX/buffer graphs, original/reverse-DFS layouts and
constraint counts zero/one: 1,536 cases and 49,152 assignment rows. Unreachable
internal logic is dropped, while every CI and all original CO ordering remain.

Another 192 native cases (3,072 rows) check trusted old-representative
substitution, with and without hashing, across all 32 polarities and three
register counts. Fixtures contain an inverted equivalent, a representative
chain, constant-zero and constant-one representatives, and logic bypassed by
substitution. Real source phases are computed before copying. All cases reduce
to exactly one internal AND, preserve output truth and match every source
Value literal. Repeated recursion must return the memoized literal without
allocating. Existing malformed-table checks reject short arrays, self cycles
and out-of-range representatives; these native invalid-input paths are not run.

Counterexample tests cover 37-frame traces crossing word boundaries with
one/three/five PIs, preserved dimensions/frame/output IDs, independent struct
and data ownership, and the distinguished sentinel. Dfs delegates trace copying
to Abc_CexDup as in C; Dfs2 deliberately leaves the optional counterexample
absent, also as in C. Eight boundary copies cover empty/constant-only graphs,
unused inputs and a register-only graph.

The fixture is `aig/gia/test_vectors/dfs_representatives.c`. An initial fixture
omitted the Dfs2 declaration (it is absent from gia.h), so implicit-int pointer
truncation crashed the oracle. Add the explicit declaration and make implicit
function declarations compile errors in this fixture. Corrected final expanded
checks pass in 1.893s (`/tmp/abcgo-dfs-final-focused.log`); the earlier harness
failure log is `/tmp/abcgo-dfs-native.log`, not evidence of a Go solver defect.
This closes Dfs/Dfs2 full-copy and representative-recursion review. General
OnePo/Skip/Node/LitArray/CiMap behavior remains open beyond the repaired simple
construction mode, and other GIA families remain open. Original package
completion stays 2/131 and the sequential audit remains active.

Full uncached race tests pass for GIA (125.693s), BMC (25.033s), SSC (7.746s)
and ABCI (28.744s), recorded in `/tmp/abcgo-dfs-review-race.log`; final GIA
vet passes (`/tmp/abcgo-dfs-review-vet.log`). No production changes followed
these checks. The remaining DFS variants are the next sequential review target.

One-property/marked-property DFS review: a new red test shows DfsSkip accepts
a marked RI, removes that CO and still sets the original register count. With
two POs and one RI, this silently reclassifies the second PO as next-state
logic (`/tmp/abcgo-dfs-skip-ri-red.log`). Reject marked RIs before clearing
source Value mappings or allocating a copy. This preserves the only coherent
existing sequential contract: filter properties and retain every register and
its corresponding RI. A three-register test rejects each RI position and
verifies failure leaves source Values untouched. This adds a precondition,
not a guessed register-removal/promotion algorithm.

OnePo exactly matches compiled C across 384 cases and 12,288 assignment
rows: all selected property indices with zero/two/four registers, 32 edge
polarities and a trailing unused CI. Compare complete source Value mappings,
object/interface counts and every output. All RIs remain in order after the
chosen PO; all CIs remain, including unused ones. Result identity is retained,
while derived-output constraints and counterexamples remain absent as in C.
Cleanup keeps the resulting CIs/gates/COs normalized. Repeat independently on
real XOR/MUX/buffer graphs and reverse-DFS source layouts for a total of 1,344
OnePo cases/43,008 rows.

DfsSkip independently covers every PO mark mask for those register counts,
polarities, typed graphs and layouts: 8,064 cases/258,048 assignment rows.
Verify exact retained properties followed by unchanged RIs, PI/RO ordering,
cone-node dependency unions, live gate counts and source CI/CO Value mappings.
Internal-node marks do not control filtering, and neither routine changes any
source mark. Both APIs intentionally reset derived-output constraint metadata.
Together the two matrices cover 9,408 cases and 301,056 rows.

The original C DfsSkip passes a CO to Gia_ManDupDfs_rec, whose uncached-node
precondition requires an AND. Its only C caller is commented out in cecCore.c.
Do not execute this broken path or claim a native Skip comparison. The existing
Go implementation uses the CO-aware DFS helper, and the independent complete
truth/dependency checks verify its intended property-filter behavior. The
working OnePo source oracle is `aig/gia/test_vectors/dfs_one_output.c`, with
implicit C function declarations treated as compilation errors.

Boundary checks include constant properties, direct wires with unused inputs,
a sequential constant property with a live RI, removal of every property,
empty and register-only graphs, and invalid property indices. Initial expanded
checks pass in 2.222s (`/tmp/abcgo-dfs-output-focused.log`); final boundary and
RI-position coverage passes in 2.173s
(`/tmp/abcgo-dfs-output-final-focused.log`). No compactness rewrite is needed:
OnePo already shares its root list across cone copying and CO construction,
and Skip's short mark-filter loop now has an explicit state-preservation guard.
Node/literal-array/CiMap behavior remains open beyond prior simple-mode fixes;
other GIA families remain open. Original package completion remains 2/131
and the sequential audit stays active.

Full uncached race tests pass for GIA (131.711s) and ABCI (29.398s), recorded
in `/tmp/abcgo-dfs-output-review-race.log`; GIA vet passes
(`/tmp/abcgo-dfs-output-review-vet.log`). No production changes followed the
checks. OnePo and Skip output/state behavior are now reviewed; node and
literal-array cone extraction are the next sequential targets.

Node/literal-array cone review: a new red ownership regression shows DfsNode
accepts an AND object from another manager and silently returns constant zero
from that object's preexisting Value cache (`/tmp/abcgo-dfs-node-owner-red.log`).
Call Gia_ObjId(src,root) before type checking/clearing source mappings. This
constant-time check enforces the same regular, manager-owned object precondition
used elsewhere in the Go manager API. Foreign/complemented handles are invalid
inputs, not requests for a different graph or output polarity. C's foreign
pointer arithmetic is not invoked as an oracle.

Compiled C matches 3,264 ordinary-AND node/literal cases and 91,008 complete
assignment rows: 32 polarities, zero/two/four source registers, each internal
node as a root, both phases of every constant/input/internal candidate, empty
literal lists, repeated/opposite-phase roots and mixed shared cones. Compare
all source Value mappings, first-discovery inputs for node extraction,
original input order for arrays, exact copied gate counts and output truth.
Node extraction retains only reachable inputs; arrays retain all five original
CIs, including an unused one after the COs. Both return combinational graphs,
retain name/spec and reset derived constraint metadata. Array arguments remain
unchanged, and source CO objects are never copied by these candidate-only APIs.

Independent Boolean formulas and fixed logical dependency/support masks extend
coverage to real XOR/MUX/buffer gates and reverse-DFS layouts, including
explicit extraction of previously dangling logic. The main four-variant matrix
covers 14,496 cases and 395,904 rows. Another 96 cases/1,536 rows compare old
representative substitution with C through the literal-array entry point:
chained inverse equivalents, constants, repeated roots and complemented
requests reduce to one AND while preserving all four original CIs. Source ROs
become ordinary combinational inputs, as in C. Combined native comparison is
3,360 cases/92,544 rows; combined independent matrix coverage is 14,592 cases
and 397,440 rows.

Boundary checks include empty lists/graphs, unused input literals, constant
buffers with zero-input support and both output phases. Root validation rejects
nil/constant/CI/CO/inverse objects without changing source Values. Invalid
literal indices and CO-as-logic literals reject; native invalid-input behavior
is not launched. The fixture is `aig/gia/test_vectors/dfs_literal_cones.c`, with
implicit declarations treated as C compilation errors. Initial matrix checks
pass in 0.488s (`/tmp/abcgo-literal-cone-expanded-go.log`), initial native checks
in 1.935s (`/tmp/abcgo-literal-cone-native.log`), and final old-representative/
boundary coverage in 2.374s (`/tmp/abcgo-literal-cone-final-focused.log`).
No further compactness rewrite is needed in these short routines; they already
share typed recursive copying without an extra traversal or support array.
CiMap behavior and other GIA families remain open. Original package completion
stays 2/131 and the sequential audit remains active.

Full uncached GIA race tests pass in 138.368s
(`/tmp/abcgo-literal-cone-review-race.log`); GIA vet passes
(`/tmp/abcgo-literal-cone-review-vet.log`). No production changes followed
these checks. Node and literal-array extraction review is complete; CI-map
substitution is the next sequential target.

CI-map substitution review checkpoint: three separate red regressions exposed
lost constraint classification on the nil-literal/full-output path, acceptance
of negative markers other than -1, and an explicit selected list too short
for the preserved register count (yielding a negative PO count). Logs are
`/tmp/abcgo-ci-map-red.log` and `/tmp/abcgo-ci-map-short-output-red.log`.
Preserve full-output constraints; reject bad markers and too-short selected
lists. All focused cases now pass in 4.147s
(`/tmp/abcgo-ci-map-final-focused.log`), including 9,432 compiled-C comparisons.
The full GIA/CEC/ABCI race gate is still running. Save this checkpoint with a
red-prefix commit to honor the requested commit interval before that full
package gate is verified. This is not a completed package audit; 2/131 original
packages remain reviewed and GIA remains in progress.

CI-map review evidence: inspect giaDup.c and its ternary/structural-sweep and
top-AND callers. CI maps compose in original CI-vector order: each slot is
allocated even when unused; constants and earlier/self CI literals then replace
its mapped Value. Forward/internal references are unavailable until copying and
must reject. All-output substitution preserves original CO order and register
roles, so it must retain the constraint suffix. Compiled C reports source/result
constraint counts `1 0`, an inherited omission deliberately repaired in Go.
Explicit literal lists define derived outputs and retain nConstrs zero. Both
branches retain the original register count in C; reject lists shorter than
that count rather than inventing register removal or a combinational result.
Gia_ManSetRegNum in both languages merely assigns the count and does not
validate the resulting negative-PO possibility.

Native C agrees across 9,432 cases/301,824 assignment rows: 131 mapping
patterns, eight source polarities, zero/two/four registers and three output
modes. Patterns enumerate the first two CI options, all five-slot chain phases,
all constant assignments and two-variable alias clusters. Modes cover nil/all
COs, repeated/complemented requested properties plus RIs, and RI-only/explicit
empty lists. Compare every source Value mapping and exact hashed graph/object
counts as well as full Boolean output rows. Every original CI slot remains in
order, including an unused late-created one; source CO Values remain unset
because the API copies their drivers. Mapping arrays and selected literal
vectors remain unchanged, and hashing is stopped before return.

Independent formulas repeat the matrix on real XOR/MUX/buffer graphs and
reverse-DFS source layouts: 37,728 cases/1,207,296 rows. The eight polarity
patterns are 0/1/2/4/8/16/21/31, not an asserted exhaustive 32-pattern matrix.
Another 144 cases/2,304 rows check old representative substitution after
identity, inversion/aliasing and constant maps, with nil and explicit output
lists. These checks cover phase-correct representative chains and constant
collapse without claiming arbitrary substitutions preserve the original
sequential behavior: they verify the exact substituted transition relation.

Boundary checks include empty graphs/maps, constant-only constraints, nil
versus nonnil-empty literal lists, and hashing a simple-mode a&!a to false.
The target deliberately starts in ordinary/hash-normalized mode and does not
copy the source raw/simple/MUX construction settings. Reusing the raw-copy
initializer here would obscure that policy. Keep all CI slots rather than
pruning unused state/interface entries. There is no additional compactness
rewrite warranted in this short routine.

Invalid-input checks cover mapping-vector lengths, future/internal/out-of-range
references, invalid markers, bad selected literals, COs used as logic, and both
empty and nonempty selected lists too short for the retained state count.
These invalid C paths are not executed. Final focused verification passes in
4.147s (`/tmp/abcgo-ci-map-final-focused.log`); the native-only run passes in
2.301s (`/tmp/abcgo-ci-map-native.log`). The fixture is
`aig/gia/test_vectors/ci_map.c`, with implicit declarations treated as errors.
The full race gate remains pending at this evidence write.

The pending gate is now green: full uncached race tests pass for GIA
(157.756s), CEC (41.535s) and ABCI (29.327s), recorded in
`/tmp/abcgo-ci-map-review-race.log`; GIA vet passes
(`/tmp/abcgo-ci-map-review-vet.log`). No code changes followed checkpoint
997b88dc8, whose red prefix recorded the then-unverified full gate. That code
is now verified green. CI-map substitution and the reviewed functions in
dup_dfs_extra.go are complete. Representative-class/top-AND duplication and
other GIA families remain open; original completion remains 2/131 and the
sequential audit remains active.

Top-AND/class review checkpoint: separate red regressions expose discarded
false output conjuncts (UNSAT becomes SAT), exponential shared-frontier
expansion and lost constraint roles during class duplication. An initial
printer-table hypothesis was also tested, then withdrawn after checking C's
filtered-new-table behavior (see the follow-up below).
`/tmp/abcgo-topand-review-red.log` records all failures, including the bounded
two-second shared-cone child and both printer-table cases. Repairs preserve
false conjuncts by returning a normalized copy, expand each positive node
once using a temporary state in the existing phase array and preserve
class-copy constraint roles. The provisional printer change in the checkpoint
was subsequently reverted to preserve C behavior.
Minimal regressions pass in 0.016s
(`/tmp/abcgo-topand-repairs-focused.log`). Save these repairs at the requested
commit interval while expanded native/semantic verification remains pending;
original package completion stays 2/131 and the broader GIA audit is active.

Top-AND expanded verification: 640 one-step/full-copy cases and 20,480 complete
input assignments check conjunction satisfiability across ten graph families,
all sixteen input-edge polarities and original/reverse-DFS layouts. Cases
include shared cones, duplicate leaves, direct-CI assignments, complemented
clauses, contradictions, identity constants, false COs, internal false fanins
and real XOR/MUX/buffer leaves. For one-step output models, use copied/forced
CI Values to lift assignments back to the original conjunction and verify every
original output. Original satisfying assignments must also satisfy the reduced
conjunction. This is existential SAT preservation with forced variables, not a
claim that output vectors or model counts remain pointwise identical.

The working ordinary-AND top-AND paths match compiled C exactly across 192
cases/6,144 rows, including complete source Value mappings, output truth,
object/interface counts and the original wrapper's count-based stopping rule.
Original C's false-conjunct witness has zero original models and four reduced
models (`0 4`); Go now keeps that conjunction UNSAT. Cases that would hit C's
constant-frontier assertion are checked independently, not launched natively.

Class duplication additionally covers 96 cases/1,536 rows with phase-correct
representative chains, constant equivalents, zero/two/four registers, hashing,
cleanup and retained constraint roles. The original C Classes entry incorrectly
passes COs to an AND-only DFS helper. Do not claim to run that broken entry:
compare logical results with the actual representative-aware C DFS helper plus
original cleanup, explicitly composed in the fixture. Go already uses that
working helper; this review's production repair is constraint preservation.
Empty/constant-only classes and missing old tables cover boundaries.

Printer hypothesis correction: the C routine filters with pReprsOld and prints
pReprs. Without a stronger specification, changing it to print only old heads
was not justified. Restore the original behavior and table requirements. Tests
now distinguish old head 1 from new head 2 and confirm C/Go print `3->2`, with
missing new storage rejected in Go. The provisional mismatch is recorded in
`/tmp/abcgo-topand-printer-fidelity-red.log`. No printer semantic change remains
relative to the pre-review code.

The fixture is `aig/gia/test_vectors/top_and.c`. Final expanded tests pass in
1.848s (`/tmp/abcgo-topand-final-focused.log`). The 48-level shared-cone test
which previously hit its two-second watchdog now completes. In the 100-call,
12-level benchmark, old frontier expansion used 40,472 bytes/34 allocations
and measured 56,630 ns per call; reusing the phase table's expanded state uses
2,936 bytes/19 allocations and measured 2,498 ns. Logs are
`/tmp/abcgo-topand-frontier-before.log` and
`/tmp/abcgo-topand-frontier-after.log`. These timings describe this benchmark;
the structural improvement is processing positive objects once rather than
unfolding every path. No extra per-object visited array is allocated. Distinct
CI leaves are counted once in verbose diagnostics; the logical leaf set is
unchanged. The full race gate is pending at this write.

The final full gate is green: uncached race tests pass for GIA (161.638s)
and ABCI (28.133s), recorded in `/tmp/abcgo-topand-review-race.log`; GIA vet
passes (`/tmp/abcgo-topand-review-vet.log`). No code changed after checkpoint
f8f532e98. The confirmed repairs are false-conjunct preservation, bounded
shared-frontier expansion and class-copy constraint roles. The printer keeps
C's original filtered-new-table behavior. Top-AND review covers the
combinational domain enforced by its ABC command caller. This completes the
reviewed functions in dup_topand.go; other duplication and GIA families remain
open. Original completion stays 2/131 and the sequential audit remains active.

Cut-replacement checkpoint: new red regressions expose an unfaithful free PI
for an uncut PI-only output, lost constraint roles and rejection of a valid
simple-mode state AND. Logs are `/tmp/abcgo-cut-logic-roles-red.log` and
`/tmp/abcgo-cut-simple-red.log`. Copy uncut PI-only output cones before
allocating ROs, rather than replacing them by new independent inputs; this
retains C's output logic while preserving the RO tail (which the original C
late-input discovery can violate). Use the existing raw-copy initializer and
retain nConstrs. Focused regressions pass (`/tmp/abcgo-cut-repairs-focused.log`).

The flop highlighter now reads logical fanins directly rather than allocating
per-node lists. Its 100-call/258-gate benchmark changes 4,136 bytes and 518
allocations per call to zero bytes/allocations (10,772 ns to 1,575 ns in this
sample). Logs are `/tmp/abcgo-cut-highlight-before.log` and
`/tmp/abcgo-cut-highlight-after.log`. Expanded typed-marker/native/lifted-
behavior checks are still pending; save this checkpoint with a red prefix at
the requested interval. Original completion stays 2/131; GIA remains active.

Expanded cut checks now pass (`/tmp/abcgo-cut-final-focused.log`, 2.011s).
The independent marker/behavior matrix covers 1,536 cases across 128 phase
patterns, three output-cone modes, ordinary AND and real XOR/MUX/buffer
graphs, and original/reverse-DFS layouts. It checks 86,016 abstract input
rows and 49,152 concrete PI/state assignments lifted to cut inputs. Every
lift retains all outputs and next-state bits; this is deliberately a one-way
abstraction contract, not preservation of the joint range of correlated cut
signals. Source mark0 is cleared, independent mark1 survives, register order
and constraints survive, and counterexamples are not copied. Separate tests
cover empty/constant graphs, unused state, no properties, combinational
outputs, and a PI-only MUX control boundary (eight phases, 64 abstract and
64 lifted rows).

The original-C fixture `aig/gia/test_vectors/replace_cut.c` agrees on 128
ordinary-AND cases and 4,096 abstract input rows, including source Value maps
and structural register-dependence marks. This comparison uses output taps
already present at the boundary, where C discovers no late PIs. A separate C
witness returns `1 0 2 1`: late PI discovery puts the original RO outside the
RO tail and an original PI inside it. The Go prepass repairs that inherited
ordering defect while copying the actual uncut logic. Real typed gates are
checked independently; the C routine does not support their representation.
GIA vet passes (`/tmp/abcgo-cut-review-vet.log`); the full GIA/ABCI uncached
race gate is running.

The final uncached race gate passes: GIA 164.221s and ABCI 28.219s
(`/tmp/abcgo-cut-review-race.log`). No production changes followed checkpoint
426813c65; expanded coverage is saved in 6b244efdb. This completes review of
dup_replace_cut.go, including logical register dependence and the cut/state
interface. Other GIA families and the final inventory reconciliation remain
open. Original package completion stays 2/131; the sequential audit is active.

Support-computation review: inspected all six routines in dup_support.go
against giaDup.c, including maximum's minimum value of 1, largest output's
first-tie behavior and descending output order. New independent per-root CI
reachability checks cover 256 graphs (32 deterministic generators, two
physical layouts, zero/two registers, AND and real XOR/MUX/buffer forms).
They check all nodes, reverse and repeated requests, constants, original CI
ordinals including ROs, sorted unique supports, output ranking and source
scratch preservation. Public result vectors are checked for independent
storage. Empty and constant interfaces and invalid requests are covered.
No correctness bug was found in the existing support routines.

The native fixture `aig/gia/test_vectors/supports.c` compares all non-CO node
supports, separate CO supports, maximum size and largest-output selection on
128 ordinary-AND cases. Original C's node routine leaves CO requests empty;
the Go port's existing CO-node support is tested independently, rather than
claimed as a C comparison. C quicksort's incidental order within equal-size
groups is not required in Go: both results must be descending permutations.

Compactness repair: sorted support lists are immutable during propagation.
Read logical fanins directly, share unary/empty-union lists and allocate only
for nonempty binary unions. Public vector entries still copy the lists. The
100-call/256-buffer benchmark changes 48,704 bytes/1,290 allocations per query
to 40,488 bytes/263 allocations, measuring 43,199 ns before and 20,133 ns after
on this host. These are benchmark-specific timings, not a universal speedup.
Focused tests and actual native comparisons pass in 1.912s; logs are
`/tmp/abcgo-support-before.log` and `/tmp/abcgo-support-after.log`. GIA vet
passes (`/tmp/abcgo-support-review-vet.log`). Full GIA/ABCI race tests are
running; original completion remains 2/131 and GIA remains in progress.

The full uncached race gate passes: GIA 167.043s and ABCI 28.740s
(`/tmp/abcgo-support-review-race.log`). Only import grouping changed after
starting that gate. This completes dup_support.go's correctness and
compactness review; no support semantics were changed. Further duplication
routines and the GIA inventory still need review. Original completion is
2/131; the sequential audit remains active.

Constraint-extraction checkpoint: inspection found that the shared AND
collector unfolds every path through a positive conjunction DAG. A new
48-level/three-input regression hit its two-second child watchdog before the
fix (`/tmp/abcgo-constraint-shared-red.log`). Keep a private expanded-node
array and visit each positive AND once. Complemented boundaries are still
recorded before checking expansion, root override is retained, and unique
leaves retain first-encounter order. Source marks and Value scratch remain
unchanged by collection. This removes exponential repeated expansion; leaf
uniqueness still uses the original vector operation, so no claim of linear
total complexity is made. Both extraction and demiter callers now complete
the shared-cone regression. The 100-call/12-level benchmark measures 51,292 ns
before and 331.4 ns after; both use zero heap allocations with this compiler.
These are bounded benchmark results (`/tmp/abcgo-constraint-shared-before.log`
and `/tmp/abcgo-constraint-shared-after.log`).

Expanded checks pass in 1.894s (`/tmp/abcgo-constraint-final-focused.log`).
The extraction matrix covers 896 graphs/28,672 assignments across 32 phase
patterns, seven conjunction families, normal/reverse-DFS layouts and
ordinary-AND/real-XOR/MUX forms. It checks exact ordered leaves, source scratch,
all signed roots and root flags, prepopulated vectors, highest-level/first-tie
selection, identities, hash shutdown and the two-output constraint contract:
`property && !constraint` equals the original output for every assignment.
False/true conjuncts, contradictions, shared subgraphs and signed internal
boundaries are included. Invalid interfaces and nondecomposable boundaries
are tested separately. Demiter compatibility covers AND/OR output polarity,
register interface and next-state preservation; this is not yet a complete
review of the other demiter routines.

The original-C fixture `aig/gia/test_vectors/constraint_extract.c` compares
448 ordinary-AND cases/14,336 input assignments, ordered leaves and result
sizes/roles. The first direct comparison exposed an existing C assumption:
its extraction copier appends CIs in physical object order, which can differ
from CI-vector order after reverse DFS. The first mismatch is preserved in
`/tmp/abcgo-constraint-native.log`. The fixture now reports that explicit
permutation and evaluates C outputs with inputs aligned to original source
ports; the Go path already preserves those ports. This aligned comparison
passes (`/tmp/abcgo-constraint-native-aligned.log`), and normalized cases need
no permutation. Typed nodes and safe rejection of non-AND root domains are
checked independently, not attributed to the C routine. Full race tests and
vet are running. Save this checkpoint with a red prefix until those gates
finish. Original completion stays 2/131; GIA remains active.

The final constraint-extraction gate is green: full uncached race tests pass
for GIA (170.791s) and ABCI (28.338s), recorded in
`/tmp/abcgo-constraint-review-race.log`. GIA vet passes
(`/tmp/abcgo-constraint-review-vet.log`). No code changed after checkpoint
9c2bb0d29. This completes dup_with_constraint.go's collector/extraction
review. The repair preserves the decomposition and leaf order while bounding
repeated expansion of shared ANDs; the existing Go source-port ordering is
retained. Other duplication routines and GIA families remain open. Original
package completion stays 2/131; the sequential audit remains active.

Basic/attribute duplication checkpoint: a new red regression shows that the
source-order basic copier reorders logical CIs and COs on an interleaved DFS
graph. Both basic and attribute copies return `[true true false]` for input
`[true false false]`, where the original returns `[false false false]`
(`/tmp/abcgo-dup-ports-red.log`). Reuse Gia_ManDupRemapCis/Cos after object
copying, before assigning register count. This retains physical object IDs,
logical port identity and the RO/RI tails with four added lines, without a new
copying algorithm. The same repair reaches auxiliary graphs copied by the
attribute wrapper. Focused regressions pass (`/tmp/abcgo-dup-ports-focused.log`).

Expanded basic-copy checks cover 768 graph/API cases and 24,576 input rows:
32 phase patterns, ordinary AND/real XOR/MUX representations, two physical
layouts, zero/two/four registers, basic/attribute entry points. The no-property
four-register cases have zero constraints. Independent Boolean expressions
check all outputs and next-state bits; source Value remains the identity map.
Attribute coverage additionally spans all 384 valid presence combinations
among mapping, packing, timing, auxiliary graph, register classes/inits,
configuration, cell string and AND delay, in four typed/layout variants:
1,536 cases/49,152 main input rows, with auxiliary-graph truth checks too.
Packing is only supplied with mapping. Mutation and stop checks verify owned
vectors, CEX trace, timing terminals/boxes/delays, graph storage and serialized
metadata. Counterexamples obey ABC's existing zero-register-init duplication
policy. Source timing traversal IDs survive while duplicate IDs reset. The
auxiliary graph intentionally receives a basic copy, matching C; its own
attribute vectors are not recursively copied.

The original-C fixture `aig/gia/test_vectors/duplicate_attributes.c` compares
384 normalized ordinary-AND cases/12,288 input rows, mapping/packing/vector
records, delay/cell data, timing boxes and actual PI arrival/PO required times,
and auxiliary-graph shape. It excludes inherited C basic-copy port/constraint
metadata defects: interleaved order and Go's retained constraint/reset/name
metadata are checked independently. The fixture initially used C's bulk
required-time getter, which returned NULL incorrectly for this box layout
and reads arrival rather than required times; the failed run is preserved in
`/tmp/abcgo-attributes-focused.log`. Read actual terminal times directly and
omit irrelevant bulk-array holes. Use an explicit valid indexed delay table.
No TIM implementation changed in this review; its own package audit remains
pending. Native comparisons and the expanded Go checks now pass in 1.954s
(`/tmp/abcgo-attributes-focused-final.log`).

The attribute wrapper's direct mapping/packing copies remain valid because
all object IDs are retained; rebuilding the records is unnecessary here.
Strings are immutable Go values; the timing manager, vectors and auxiliary
graph retain explicit owned copies. Full uncached module race tests and GIA
vet are running because the basic copier is a shared graph boundary. Save
this checkpoint with a red prefix until the integration gate finishes.
Original package completion remains 2/131; GIA review continues.

Final basic/attribute integration is green: the full uncached module race run
passes for all 130 tested packages; three command packages have no tests
(`/tmp/abcgo-attributes-module-race.log`). This includes GIA 176.384s,
ABCI 29.167s and BMC 23.848s. No failure, race warning or extra diagnostic
lines appear in that log. GIA vet passes
(`/tmp/abcgo-attributes-review-vet.log`). No code changed after checkpoint
5adb50cd5. This completes the reviewed functions in dup_basic.go and
dup_attributes.go: logical ports now survive interleaving, and the listed
attribute payloads retain independent ownership. Other duplication/GIA
families and final package inventory reconciliation remain open. Original
completion stays 2/131; the sequential audit remains active.

Marked-duplication checkpoint: red regressions expose reordered logical ports
and rejection of valid retained RO/RI pairs on interleaved DFS sources
(`/tmp/abcgo-marked-ports-red.log`). Copy objects in physical source order as
before, then rebuild retained CI/CO vectors in their original logical order.
Use those same source indices for names, register-pair validation and metadata
selection. This removes physical-order assumptions without changing retained
object IDs or the marked-subgraph algorithm. Separate RO/RI counters are no
longer necessary. Focused regressions pass
(`/tmp/abcgo-marked-ports-focused.log`).

The independent selection/state matrix now passes in 0.267s
(`/tmp/abcgo-marked-matrix.log`): 7,168 cases and 224,576 input rows over four
mixed/all-positive/all-negative phase patterns, two physical layouts,
ordinary-AND/real-XOR/MUX/buffer forms, every valid subset of three register
pairs and their branch inputs/properties, optional constant outputs and an
optional anchor PI. MUX selection retains its anchor when needed. Checks
cover exact outputs/next state, logical port/object maps, source mark cleanup,
independent mark1 preservation, property/bad/constraint counts, names,
register-class/init/reset subsets and owned metadata. Empty retained
interfaces and surplus unused PIs are included. Additional class/choice,
invalid-input and native-C checks are pending, as is the full integration
gate. Save this checkpoint with a red prefix. Original completion stays
2/131; GIA remains in progress.

Expanded marked checks pass in 2.215s
(`/tmp/abcgo-marked-final-focused.log`). The native fixture
`aig/gia/test_vectors/marked_duplicate.c` agrees on 512 normalized
ordinary-AND cases and 23,232 input rows, complete source Value maps, retained
shape/ports and register/constraint counts. It retains the constraint PO:
original C otherwise leaves a stale constraint count. Typed/interleaved cases
and deleted buffers are tested independently; C's marked routine explicitly
forbids deleting buffers. The Go path's existing more general acceptance of a
valid retained subgraph is preserved, rather than treating that C assertion
as an algorithmic requirement.

Eight class/choice removal masks verify retained equivalence-class reheading
and next links, source-table immutability, independent destination storage
and C's immediate-sibling retention rule. This deliberately does not invent
transitive choice-link repairs when the immediate sibling is removed. Tests
also reject missing RO/RI halves, equally sized but mismatched register sets,
retained AND/MUX dependencies on removed inputs and removal of constant zero.
Raw simple-mode contradictory fanins remain supported. These checks found no
further production bug beyond logical port/pair ordering. Full uncached
module race verification and GIA vet are running; marked duplication also
implements combinational/sequential cleanup and representative rebuilding.
Original completion stays 2/131 and the audit remains active.

Final marked-copy integration is green: all 130 tested module packages pass
the uncached race run; three command packages have no tests and no extra
failure/race diagnostic lines appear
(`/tmp/abcgo-marked-module-race.log`). GIA passes in 181.819s, ABCI in
29.396s, CEC in 45.056s and BMC in 24.064s. GIA vet passes
(`/tmp/abcgo-marked-review-vet.log`). Production code is unchanged since
35c418d6c; subsequent changes add the class/choice/invalid-input checks and
native fixture. This completes dup_marked.go's review. Retained logical
interfaces, complete state pairs and their selected metadata now survive
physical interleaving; the original object-copy, class and immediate-choice
rules remain. Other duplication/GIA families and final package inventory
reconciliation are still open. Original package completion remains 2/131;
the sequential audit is active.

Trimming checkpoint: red tests expose removal of an always-different `(0,1)`
dual pair and loss of retained constraint/bad-output roles in both trimming
entry points (`/tmp/abcgo-trim-review-red.log`). The nonzero-pair helper tested
the driver's object ID, discarding constant one as well as constant zero.
Test the full literal so only false is discarded. Count retained output roles
when appending original POs; dummy outputs stay ordinary. Stateful dual tests
also expose a missing bad-output count in the unchanged-output CI-map path
(`/tmp/abcgo-trim-cimap-roles-red.log`). Preserve it alongside nConstrs only
when no explicit output list is supplied. This is a follow-up to the earlier
CI-map audit, which covered constraints but missed the separate Go bad-output
partition. Explicit output selections retain their derived-interface policy.

Focused regressions pass (`/tmp/abcgo-trim-repairs-focused.log`). Reference
counting shared by both trimming routines now reads logical fanins directly,
without temporary per-node lists; skip counting when PI trimming is disabled.
The 100-call/256-buffer benchmark changes 22,944 bytes/528 allocations per
call to 21,904 bytes/269 allocations (20,145 ns to 18,537 ns in this sample).
Logs are `/tmp/abcgo-trim-before.log` and the focused log above. The result is
specific to this benchmark; the retained-node/reference rules are unchanged.
Expanded ordinary/dual/metadata and original-C checks, plus the full
integration gate, are pending. Save this checkpoint with a red prefix.
Original completion remains 2/131; GIA review continues.

Expanded trimming checks now pass: the ordinary/PI-output matrix covers
3,328 cases and 153,600 input/state rows across all 64 polarity patterns,
all input/output trim toggles and constant selectors, two physical layouts
and ordinary/real-XOR/MUX/buffer forms. Exact retained CI/CO maps, property
roles, RO/RI order, independent Boolean output formulas, untouched source
marks/reference cache and graph shape are checked. It includes a control PI
used only by a real MUX after discounting PO references. Results are in
`/tmp/abcgo-trim-matrix.log` (0.172s).

The dual matrix covers 64 cases and 32,768 length-three zero-reset input
sequences (98,304 frame comparisons), including structural equality, constant
inequality, mixed literal phases, typed gates, DFS layouts, unused PIs and a
constant/dynamic register. It verifies retained pair indices, exact retained
outputs/next state and resulting role/register counts. Removed constant
registers are compared only on states reachable from zero reset, matching
ABC's sweep domain. Dummy PI/PO interfaces, no properties, raw simple-mode
contradictory fanins, selector/ownership guards and odd PO counts are covered.
Explicit CI-map output lists are checked to retain their derived role policy.
These tests pass in `/tmp/abcgo-trim-dual-expanded.log` (0.140s).

The native ordinary-trim fixture `aig/gia/test_vectors/trim.c` matches 832
normalized ordinary-AND cases/37,888 input-state rows, retained structure,
ports and non-CO source Value maps. C does not assign source CO Value scratch;
the Go convention is checked independently. Constraint/bad partitions are
also independently checked because original C drops those labels. A separate
C witness returns `0 0 0`: no retained pair and two false outputs after
trimming `(0,1)`, confirming the inherited constant-mismatch defect. This
initial native run passes (`/tmp/abcgo-trim-native.log`, 1.893s). Additional
working-domain native dual comparison and the full integration gate remain
pending. Original completion stays 2/131; GIA remains active.

The final focused trim suite passes in 2.226s
(`/tmp/abcgo-trim-final-focused.log`). The native fixture additionally covers
16 working-domain dual cases/144 input-state rows, including structural equal
pairs, dynamic/constant state, both PI-trim settings and mixed phases. These
compare exact nonzero-pair indices, swept structure, surviving register count
and output/next-state truth. They contain no always-unequal constant pair;
that inherited C defect remains an explicit separate witness. Total native
coverage here is 848 cases/38,032 rows. GIA vet passes
(`/tmp/abcgo-trim-review-vet.log`); full uncached module race tests are running.

Follow-up review focus: check the distinct Go bad-output partition in other
full-output copy paths. Earlier duplication reviews established constraint
preservation but did not, by themselves, prove preservation of that separate
partition. This turn repairs the unchanged-output CI-map path exercised by
trimming's sweep; it does not claim all other duplication/metadata paths have
been reconciled. Package completion remains 2/131 and GIA stays in progress.

Final trimming integration is green: all 130 tested module packages pass
the uncached race run; three command packages have no tests and no extra
failure/race diagnostic lines appear (`/tmp/abcgo-trim-module-race.log`). GIA
passes in 187.584s, ABCI in 29.789s, ABS in 17.895s and BMC in 23.829s. GIA
vet passes (`/tmp/abcgo-trim-review-vet.log`). Production code is unchanged
since 0eac0f4ab; expanded coverage is saved in 3b81760e6 and the final working
native-dual additions. This completes the four original functions in
dup_trim.go and the exercised CI-map bad-output follow-up. The contracts
checked are exact selected outputs/next state and surviving roles, with
zero-reset reachability for swept state; deliberate output deletion is not
claimed to preserve the original property set. Other GIA families, the
bad-partition follow-up above and final inventory reconciliation remain
open. Original completion stays 2/131; the sequential audit remains active.

Full-output bad-partition checkpoint: red regressions show 14 invocation
variants losing two bad-output labels while retaining two constraints
(`/tmp/abcgo-copy-partition-red.log`). AIGER makes the loss observable: the
header changes from one ordinary/two bad/two constraints to three ordinary/
zero bad/two constraints. The affected functions are Normalize (raw/hashed),
Levelized, OrderDfs, OrderDfsReverse, OrderAiger, Dfs, Dfs2, DfsClasses,
Abs (none/each/both promoted registers) and ReplaceCut. These variants keep
the complete source PO list in order, including abstraction/cut variants
whose inputs or remaining state change.

Use one narrow helper to copy both role counters at those explicit sites.
Basic copying and the recently repaired unchanged-output CI-map path use the
same helper without changing their behavior. The shared draft initializer
still starts with no roles, preserving explicit derived-output policies.
Focused regressions, AIGER round trips and the existing CI-map derived-role
control pass (`/tmp/abcgo-copy-partition-focused.log`, 0.005s). No copying,
hashing, abstraction or cone-selection algorithm changed. Expanded partition/
behavior/serialization coverage and full integration verification are pending.
Save this checkpoint with a red prefix. Original completion remains 2/131;
the GIA audit continues.

Expanded partition checks pass in 0.592s
(`/tmp/abcgo-copy-partition-final-focused.log`). The matrix covers all 21
valid ordinary/bad/constraint partitions of five POs, four mixed/literal
phase patterns, ordinary/real-XOR/MUX/buffer forms, two physical layouts and
18 invocation/control variants: 6,048 cases and 193,536 original PI/state
assignments checked against both the copy and its AIGER round trip. Identity
copies must retain the original input vector directly; register promotion
uses its explicit expected permutation, and cut copies use one-way lifted
boundary inputs. The latter is not a joint-range equivalence claim. All
surviving RI values are checked. Raw AIGER header counts are checked separately
from the reader's recovered roles. Empty property lists and six explicit
selection/vector/literal APIs also retain their existing derived-role policy.

AIGER intentionally requires real gates to be lowered first. Connecting that
public lowering path exposed another physical-order assumption: DupNoMuxes
reorders ports on DFS sources. A separate red regression passes for normalized
inputs but fails on the same interleaved circuit
(`/tmp/abcgo-copy-lowering-ports-red.log`). Restore CI/CO vectors through the
existing remapping helpers before metadata copying and cleanup. The direct
regression and expanded original-order matrix now pass. Typed round trips use
that public lowering path; the exporter acceptance rules did not change.
The original failed expanded log (`/tmp/abcgo-copy-partition-expanded.log`)
records the deliberate exporter rejection, not a production exporter defect.

Full uncached module race tests and GIA vet are running. This review covers
output roles in the listed complete-PO copies and logical port retention in
the lowering path; it does not declare all other balancing, quantification,
box-order or rewrite metadata paths reviewed. Save the expanded coverage and
lowering repair with a red prefix until integration verification finishes.
Original completion stays 2/131; GIA remains active.

Final complete-copy partition integration is green: all 130 tested module
packages pass the uncached race run; three command packages have no tests
and no extra failure/race diagnostic lines appear
(`/tmp/abcgo-copy-partition-module-race.log`). GIA passes in 192.397s,
ABCI in 28.991s, CEC in 44.944s and BMC in 23.875s. GIA vet passes
(`/tmp/abcgo-copy-partition-vet.log`). No code changed after 6bfd1f498.
The confirmed repairs are complete-PO role preservation at the explicit
listed sites and logical CI/CO retention during real-gate lowering. The
shared initializer still creates derived interfaces without inherited roles.
This closes this partition/export follow-up; it does not declare every
algorithm in the lowering family reviewed. Continue the separate role audit
in cofactor/quantification, balancing, rehashing and box-order rewrites,
then reconcile the remaining GIA algorithm inventory. Original completion
remains 2/131; the sequential audit stays active.

Cofactor/quantifier bad-partition checkpoint: red regressions show bad-output
labels lost in eight supported operation/state variants: variable cofactors
at both values, Exist2 with/without registers, combinational Univ and the
single-output Exist API (`/tmp/abcgo-quant-partition-red.log`). Reuse the
paired output-role helper at the two sites that already copied nConstrs.
No cofactor/quantifier algorithm or shared constructor changed. The minimal
regressions pass (`/tmp/abcgo-quant-partition-focused.log`).

Expanded checks pass in 0.241s (`/tmp/abcgo-quant-partition-expanded.log`).
There are 3,616 cases/57,856 input-state rows over every valid role partition
for the corresponding one/two/three-PO interface, every PI variable, four
literal-phase patterns, ordinary/real-XOR/MUX/buffer forms and normal/DFS
layouts. Each original assignment is evaluated with the variable forced to
zero/one to form an independent expected result. Variable cofactors use the
selected branch, existential variants OR both branches, and C's Univ rule
ANDs only the first output while auxiliary outputs use the positive branch.
Exist2's RI transformation follows C's component-wise OR contract; no broader
sequential-equivalence claim is made. Every copied/rewritten vector and its
AIGER round trip are checked, including raw header roles, register counts
and source-role immutability.

Seven derived-output controls cover guarded PI/internal-node cofactors at
both values, block extraction, distinct-source MUX reconstruction and aliased
MUX sources. They check exact guarded/selected Boolean values across all 16
inputs and retain the original unclassified-output policy. This distinguishes
the two explicit complete-output role sites from shared constructor behavior.
Full uncached module race tests and GIA vet are running. Save this checkpoint
with a red prefix until integration finishes. Original completion remains
2/131; the GIA audit continues.

Final cofactor/quantifier partition integration is green: all 130 tested
module packages pass the uncached race run; three command packages have no
tests and no extra failure/race diagnostic lines appear
(`/tmp/abcgo-quant-partition-module-race.log`). GIA passes in 195.849s,
ABCI in 29.888s, PDR in 14.769s and BMC in 24.267s. GIA vet passes
(`/tmp/abcgo-quant-partition-vet.log`). No code changed after b8f937c88.
This closes the bad-output role follow-up at the two explicit cofactor/
quantifier sites. Source Boolean rules and derived-operation role policies
remain intact, including C's auxiliary-output universal behavior. Continue
with rehashing, balancing and box-order role paths and the remaining GIA
algorithm inventory. Original completion stays 2/131; the sequential audit
remains active.

Rehash checkpoint: three separate red regressions expose loss of bad-output
roles, reordered logical ports on DFS sources and real XOR/MUX nodes treated
as ordinary ANDs (`/tmp/abcgo-rehash-review-red.log`). Both additional-strash
settings fail; the buffer control already works. The original C routine
covers ordinary AIG objects and likewise copies physical port order. Keep
its ordinary hash/additional-strash path, precreate CIs in logical order,
reuse giaComposeNodes for existing typed hashing/lowering and append COs in
logical order. Preserve paired role counters before cleanup. Source Value
scratch still records pre-cleanup copy literals; final CI slots remain stable.

Focused repairs pass (`/tmp/abcgo-rehash-repairs-focused.log`, 0.005s).
This reuses existing Boolean constructors rather than introducing a new
optimizer, and removes the incomplete gate switch without growing the
function body. Normalized ordinary-AIG source comparison, broader typed/
layout/role tests and integration verification are pending. Save this
checkpoint with a red prefix. Original completion remains 2/131; GIA review
continues.

Expanded rehash checks pass in 1.523s (`/tmp/abcgo-rehash-expanded.log`).
The matrix covers 9,216 cases/589,824 input-state rows: all 64 literal-phase
patterns, six shared/absorbing/contradictory/complemented cone families,
zero/two/four registers, both additional-strash settings, two physical
layouts and ordinary/real-XOR/MUX/buffer forms. Independent Boolean formulas
check every output and next-state bit in the result and its AIGER round trip.
Checks also cover source identity/roles, stable logical CI copy slots,
normalized ordinary-AIG output, cleared destination hash/additional modes,
unchanged source marks/reference cache, and removal of all dangling logic.
Source dangling nodes are deliberately copied before final cleanup, as in C.

The native fixture `aig/gia/test_vectors/rehash.c` agrees on 2,304 normalized
ordinary-AND cases/147,456 input-state rows, retained structure, register count
and every pre-cleanup source Value literal, including dangling objects and
COs (`/tmp/abcgo-rehash-native.log`, 2.238s). Typed/interleaved cases and the
Go bad-output partition are independently checked; they are not claimed as
supported original-C behavior. Native counts intentionally exclude C's lost
constraint metadata. No hash/additional-strash primitive was replaced.

Separate tests retain an already active source hash table and its mode,
including canonical lookup without allocation of another source node. Empty
interfaces, constant-only outputs, state-only interfaces and simple-mode
repeated/opposite fanins preserve truth and serialize correctly; hashed output
uses canonical ordinary nodes. Full uncached module race tests and GIA vet
are running. Save expanded verification with a red prefix pending integration.
Original completion stays 2/131; GIA remains active.

Final rehash integration is green: all 130 tested module packages pass the
uncached race run; three command packages have no tests and no extra
failure/race diagnostic lines appear (`/tmp/abcgo-rehash-module-race.log`).
GIA passes in 208.906s, ABCI in 29.068s, IF in 5.329s and BMC in 23.787s.
GIA vet passes (`/tmp/abcgo-rehash-review-vet.log`). Production code is
unchanged since 4c726df8e; expanded coverage is saved in cde83008b.
This completes Gia_ManRehash's review, not the entire giaHash.go file. Its
normalized ordinary path matches C exactly in the tested families; real-gate
composition and logical interfaces retain the Go graph contract. Output
roles survive cleanup and serialization. Continue with balancing/box-order
metadata paths and the remaining GIA algorithm inventory. Original completion
remains 2/131; the sequential audit is active.

LUT balancing/coarsening follow-up: three new regressions failed before
production edits (`/tmp/abcgo-balance-review-red.log`): balancing discarded
the bad-output count, DFS/interleaved source layouts changed logical port
identity in both coarsening and public balancing, and an unsupported raw XOR
was silently interpreted as an AND by the ordinary-AIG coarsener.

Gia_ManDupMuxes now remaps CI/CO vectors from source logical order before
cleanup and rejects raw XOR alongside its existing raw-MUX precondition.
Both balancing copy stages preserve complete output roles. Focused tests,
including the existing original-ABC balancing oracle, pass. Expanded
parameter/partition/serialization verification is running; full integration
is still pending. This does not finish the balancing algorithm review or
GIA package review. Original completion remains 2/131.

Expanded balancing/coarsening verification passes: 6,384 transformations and
204,288 input/state rows, with direct truth and serialized AIGER truth/role
checks. The matrix covers all 21 valid bad/constraint partitions of five POs,
four phase patterns, ordinary and DFS/interleaved layouts, every supported
LUT size 2..6/cut count 2..8, and coarsening limits 0/2/8. The existing
48-case original-ABC oracle also passes, comparing AND count, structural
level and all 64 truth rows for each result. Focused/expanded tests pass in
1.849s (`/tmp/abcgo-balance-expanded.log`). The matrix graphs are ordinary
AND sources; existing random/native balancing tests exercise recognition
of expanded XOR/MUX patterns. Raw XOR input is deliberately rejected by
the ordinary-AIG coarsener, not claimed as a newly supported operation.

Full uncached module race integration is green: 130 tested packages pass,
three command packages have no tests, and no failure/race diagnostics appear
(`/tmp/abcgo-balance-module-race.log`). GIA passes in 214.597s, ABCI in
29.493s and BMC in 24.065s. GIA vet passes
(`/tmp/abcgo-balance-vet.log`). Production fixes and regression tests are
saved in 6d185e2f1. This completes the identified role/port/precondition
repairs, not the entire low-level LUT balancing review. Original completion
remains 2/131; GIA and the sequential audit remain active.

The remaining LUT-balancing helpers were compared with giaBalLut.c:
cut signatures/subset checks, ordered two-/three-input unions, antichain
insertion and tie order, candidate cut enumeration, gate costs, pair choice,
AND/XOR simplification, bounded supergate collection and recursive balancing.
Ignored complement/function flags correspond to disabled truth-minimization
code in C. The Go unit-cut allocation replaces C's three static scratch
slots, keeping calls isolated without changing the retained cuts.

New native fixture `aig/gia/test_vectors/lut_balance_internal.c` includes the
actual C implementation to expose static helpers; no algorithm is duplicated
in the fixture. `lut_balance_internal_review_test.go` compares exact cut
leaves/signatures/functions/delays, returned costs, sorted candidate literals,
selected pair indices, and all collected supergates for 560 configurations
(16 seeds, K2..6, C2..8), 26,880 gate decisions and 26,880 pair decisions.
Three-input MUX gates use K>=3; K2 exercises AND/XOR as required by the
native gate-fitting contract.

Additional exact native checks cover all 12,870 multisets of up to eight
literals over three variables, both AND/XOR simplification modes (25,740
cases), 13,071 two-cut and 13,869 three-cut merges with deliberately colliding
64-bit signatures, and 16 supergate collection cases crossing the size-50
and fanout-three boundaries. Independent Boolean checks verify every
simplification on all eight assignments (205,920 rows) and each boundary
collection on 128 assignments (2,048 rows). Original-C intermediate decisions
agree; no new production defect or worthwhile simplification was found.
Expanded native coverage passes with the race detector in 4.962s
(`/tmp/abcgo-lut-internal-expanded-race.log`). Full GIA race integration is
still running; vet passes. Save a red checkpoint pending integration.
GIA review remains active; original completion stays 2/131.

Final LUT-internal integration passes: full uncached GIA race tests finish
in 217.048s (`/tmp/abcgo-lut-internal-final-race.log`) and GIA vet passes
(`/tmp/abcgo-lut-internal-final-vet.log`). The first full run used a test
binary built before the three-cut merge expansion while its C fixture had
already been expanded; its sole failure was the resulting extra native
`muxmerge` row, not a production mismatch. The final run held both test and
fixture files unchanged. Do not change native fixture protocols while a
compiled test snapshot is running.

No production changes were needed for this helper review; tests/fixture are
saved in 8cfa09719. Together with the preceding wrapper regressions and
original-ABC structural/truth comparisons, this covers the LUT-balancing
implementation reviewed here. It does not finish GIA or other balancing
families. Continue with `timing_box_order.go`, comparing boxed traversal and
unnormalization with giaTim.c, including output-role retention. Original
completion remains 2/131 and the sequential audit remains active.

Boxed timing/order review: two new regressions fail before their respective
fixes. Gia_ManDupUnnormalize drops the bad-output count despite preserving
all outputs (`/tmp/abcgo-box-order-red.log`). With unreachable source logic,
it also allocates the sibling table at the larger source-object count,
violating the destination's exact-size metadata contract and preventing
AIGER serialization (`/tmp/abcgo-box-order-sibling-red.log`: 19 sibling slots
for 18 retained objects). Copy output roles through giaCopyOutputRoles and
size the sibling table from the retained order, whose entries each produce
one destination object including the existing constant.

Box traversal previously allocated two fanin slices for each ordinary AND.
Direct fanin calls retain the physical data-fanin order, buffer handling and
real-MUX control dependency without allocating slices. A 256-AND chain over
100 benchmark calls improves from 5,280B/514 allocations/19,876ns per call
to 1,184B/2 allocations/5,506ns (`/tmp/abcgo-box-order-bench-before.log`,
`/tmp/abcgo-box-order-optimized.log`). The traversal algorithm is unchanged.

Native fixture `test_vectors/timing_box_order.c` compares the actual giaTim.c
order and unnormalization routines on 1,536 ordinary-AND cases: all 256 phase
patterns, choice siblings on/off, and zero/one/two register pairs. Exact
object order, all source Value copies (including omitted dangling nodes),
destination structure/sibling indices and 49,152 input/state truth rows
agree. Go's bad-output field is independently checked, not represented as
native-C proof. A 304-case typed/plain/choice/property-partition matrix
checks 9,728 rows directly and after AIGER serialization.

Additional checks cover asymmetric box input/output widths; buffer/constant
cones; rejection of future-box data and real-MUX-control dependencies across
repeated calls; all two-box selections/permutations and terminal reductions;
selected extra-AIG output identities; and reduced timing-table ownership.
The expected topological-order diagnostics in rejection tests are intentional.
All focused tests pass in 1.950s (`/tmp/abcgo-box-order-final-focused.log`).
Save a red checkpoint pending full integration. Original completion remains
2/131; GIA and the sequential audit remain active.

Boxed-copy integration is green: full uncached module race tests pass all
130 tested packages; three command packages have no tests and no failure/
race diagnostic lines appear (`/tmp/abcgo-box-order-module-race.log`).
GIA passes in 221.907s, ABCI in 29.173s and BMC in 23.961s. GIA vet passes
(`/tmp/abcgo-box-order-vet.log`). Production fixes and expanded tests are
saved in a44d46122; source/tests/fixture stayed unchanged during integration.
This closes the identified output-role and sibling-sizing bugs, with the
allocation reduction retaining tested ordinary/typed traversal decisions.

Further timing-family review remains: verify the C buffer-specific sibling
traversal rule (C handles a buffer before following choice siblings), then
review timing_unshuffle.go against giaTim.c, especially output-role and real
gate retention. State-coordinate duplication in dup_seq.go also needs its
remaining review; do not infer that prior copying-family repairs cover it.
Original completion remains 2/131; GIA and the sequential audit remain active.

Buffer-specific box traversal follow-up confirms another source discrepancy.
C handles a buffer before considering choice siblings; Go followed a buffer's
sibling. The new minimal red test adds an otherwise unused equivalent buffer
to the resulting order (`/tmp/abcgo-box-buffer-red.log`). The compiled C
oracle also exposes a false topological-order rejection when the equivalent
sibling structurally depends on a future box (`/tmp/abcgo-box-buffer-native-red.log`).
The fixture's two buffer alternatives are exhaustively verified equivalent;
this is not an inconsistent choice annotation.

Only consider choice siblings for non-buffer nodes, preserving C's buffer
rule without allocating fanin slices. Native fixture `test_vectors/box_buffer_order.c`
and `box_buffer_review_test.go` compare exact traversal order on eight cases
and 128 truth rows, covering both input polarities and a sibling with/without
the future-box dependency. Both new regressions pass after the change
(`/tmp/abcgo-box-buffer-green.log`, 1.813s). Save a red checkpoint pending
shared timing-family integration. Original completion remains 2/131.

Boxed input unshuffling review finds two more defects, both covered by new
red regressions before repair (`/tmp/abcgo-unshuffle-red.log`). The complete
output list loses its bad-output count. The old internal-node loop treats
all gates as ordinary ANDs, changing XOR/MUX truth and panicking for buffers
and repeated/opposite-fanin simple-mode ANDs.

Reuse giaDupMappedStart, giaDupMappedNodes and giaDupMappedOutput instead of
the incomplete local copying loops; these retain raw gate kinds and simple
mode while preserving the existing C CI permutation. Copy complete output
roles via giaCopyOutputRoles. This reduces local production code and keeps
the already-reviewed raw-copy rules in one implementation. The source C
routine itself assumes ordinary ANDs; typed/simple cases are checked against
the Go graph's Boolean semantics, not presented as supported native inputs.

`timing_unshuffle_review_test.go` and `test_vectors/timing_unshuffle.c`
compare 432 ordinary native cases/43,904 input-state rows, exact destination
structure, all source copy literals and equivalence representatives/nexts.
PI, register and box-output widths each range independently over 1..3, with
all 16 phase patterns. A separate 5,400-copy matrix covers all five gate
families, valid three-PO bad/constraint partitions and four phase patterns:
548,800 truth rows match before and after serialization under the explicit
C input permutation. This is an assignment-permutation check, not a claim
of sequential equivalence under unchanged register-coordinate labels.
Source informational unshuffle warnings are expected. All focused timing
checks pass in 5.054s (`/tmp/abcgo-unshuffle-expanded.log`). Full module race
and GIA vet integration follow; save a red checkpoint until verified.
Original completion remains 2/131; GIA and the sequential audit stay active.

Buffer traversal and unshuffle integration is green. Full uncached module
race tests pass all 130 tested packages; three command packages have no tests
and no failure/race diagnostic lines appear
(`/tmp/abcgo-unshuffle-module-race.log`). GIA passes in 236.165s, ABCI in
28.999s and BMC in 23.888s. GIA vet passes
(`/tmp/abcgo-unshuffle-vet.log`). Source/tests/native fixtures stayed unchanged
during the integration run. Buffer repair and tests are in bf6ff6f59;
unshuffle repair and expanded tests are in 3a1edbf8e.

This closes the buffer-specific source discrepancy and the unshuffle role/
gate-kind repairs while retaining exact native ordinary-AIG behavior. The
remaining state-coordinate duplication/simulation family in dup_seq.go is
next: inspect physical versus logical port order in Gia_ManDupFlip, preserve
its complemented CI copy-map semantics, and check Cycle/DupCycled against
source and transition-table oracles. Original completion remains 2/131;
GIA and the sequential audit remain active.

State-coordinate duplication/simulation review finds a logical-port defect
in Gia_ManDupFlip. Its physical object loop assigns CI/CO ordinals in DFS
object order instead of source logical vector order. Four new red cases
(mask 0..3) fail, including the zero-flip identity case
(`/tmp/abcgo-state-flip-red.log`). Complemented RO source Value literals are
intentional; the regular-literal CI remapper cannot repair this copy safely.

Create all CIs in logical order, preserving each selected RO's complemented
copy literal; copy internal gates via giaDupMappedNodes; append COs in
logical order, complementing each selected RI. Reuse giaDupMappedStart and
giaCopyOutputRoles. This shortens the local implementation while retaining
raw XOR/MUX/buffers, simple mode and the complete output partition. Ordinary
normalized input retains the original C structure/copy-map behavior.

New tests in dup_seq_review_test.go cover 2,016 phase/type/layout/register/
flip-mask combinations and 64,512 input/state rows, checking the state
coordinate relation, complemented CI copy maps, normalized result, output
roles, AIGER serialization and double-flip restoration. Another 48 cases /
6,144 rows cover 31/32/33/63/64/65 registers, zero/all/alternating/boundary
flip masks, and ordinary/DFS wire graphs across packed-word boundaries.

seq_cycle_review_test.go checks Cycle against a frame-by-frame Boolean
transition evaluator for 20,160 cases, including every object mark, mark1
preservation and the next deterministic random word. It checks DupCycled's
zero-state simulation, dirty-mark cleanup, derived flip semantics and roles
for 1,920 copies/61,440 rows. Another 300 counterexample cases vary register
and PI counts across word boundaries, zero through five frames and DFS
layouts. Their reference retains unpacked PI frames to avoid repeating a
packed-offset bug. Cycle uses existing RO marks and ignores the CEX register
prefix; DupCycled deliberately resets to zero before simulation, as in C.

seq_native_review_test.go and test_vectors/state_cycle.c execute original
C routines on 504 flips, 10,080 simulations and 960 cycled copies. Exact
structure/source Value maps, scratch marks, deterministic random words and
46,848 output truth rows agree (`/tmp/abcgo-state-native.log`, 1.947s).
Typed/DFS behavior and Go-specific output roles are independently checked,
not claimed as supported native behavior. The transition matrix passes in
0.626s (`/tmp/abcgo-seq-transition.log`); wide CEX checks pass in 0.011s
(`/tmp/abcgo-cycle-word-boundaries.log`). No additional Cycle/DupCycled
production bug or useful simplification was found. Save a red checkpoint
pending full module race tests and GIA vet. Original completion stays 2/131;
GIA and the sequential audit remain active.

State-flip/cycle integration is green: full uncached module race tests pass
all 130 tested packages, with three command packages without tests and no
failure/race diagnostic lines (`/tmp/abcgo-state-cycle-module-race.log`).
GIA passes in 241.527s, ABCI in 29.428s and BMC in 24.201s. GIA vet passes
(`/tmp/abcgo-state-cycle-vet.log`). Production repair and expanded regression/
native/transition coverage are saved in f54e87a44. Source/tests/fixture stayed
unchanged during integration.

This completes the reviewed dup_seq.go family: Flip preserves the logical
interface with its intentional signed CI copy map, while Cycle/DupCycled
retain source replay/random-input and scratch-state behavior. Native ordinary
normalized cases agree exactly; typed/interleaved extensions satisfy the
independent coordinate/transition checks. Continue reconciling the remaining
GIA inventory, including giaHash.go functions outside the previously reviewed
Gia_ManRehash path. Original completion stays 2/131; GIA and the full
sequential audit remain active.

Structural-hash follow-up reviews giaHash.go outside Gia_ManRehash. A new
red regression confirms HashProfile omitted the original bucket-occupancy
line (`/tmp/abcgo-hash-profile-red.log`). Restore the bounded 1,000-bucket
chain-length report after the summary. The scalar/reduction rules inspected
so far agree with C, including intentionally different empty-input results:
HashAndMulti returns 0, while HashAndMulti2 returns 1; an odd dual-miter tail
is ignored by the source pair iterator.

New native fixture test_vectors/hash_primitives.c includes the actual C
implementation and compares ordinary/additional/real/simple modes. It checks
58x58 operand pairs per mode, all 10x10x10 primary-literal triples per mode,
and every six-symbol input vector of lengths 0..5 for all three reducers.
Returned literals, object counts, hit/miss counters, vector mutation and
profile output match. Independent 16-row truth tables check AND/OR/XOR,
MUX/majority, AND reductions and paired-output miters; source Simple-mode
construction is retained. Native coverage passes in 2.004s
(`/tmp/abcgo-hash-primitives-native.log`). Save a red checkpoint while resize
allocation/compactness and the remaining static helpers are being checked;
full integration is pending. Original completion remains 2/131, GIA active.

Expanded hashing review now covers the remaining static helpers and lifecycle.
The native protocol compares 13,456 operand pairs, 4,000 control/data triples,
37,324 reducer input vectors (111,972 reducer calls), 4,032 hash keys and all
264 qualifying ordered MUX-recognition pairs. It verifies returned literals,
node counts, hit/miss counters, pointer/literal lookup agreement, destructive
versus non-destructive vectors and complete profile output after stopping,
restarting and explicitly resizing the table. Every recorded Boolean result
is checked on all 16 primary-input assignments. Recognized control/then/else
literals independently reproduce the NOR of the two product terms.

C-style ownership transfer replaces HashResize's unnecessary old-bucket clone:
retain the old slice and zero the destination vector before filling the new
one. Bucket traversal cannot be overwritten; no intermediate copy is needed.
The 100-call 480-gate construction/resize benchmark improves from 55,824B /
520 allocations / 118,298ns per run to 53,616B / 518 allocations / 110,372ns
(`/tmp/abcgo-hash-resize-before.log`, `/tmp/abcgo-hash-optimized.log`). These
are measured whole-workflow figures, not isolated resize timings. Simplify
the recognizer's translated assertion scaffolding and combine repeated
child extraction in AddStrash, retaining all source reduction branches.

Additional lifecycle tests check repeated mixed AND/real-XOR/real-MUX table
rebuilds, same-data MUXes with distinct controls, collision-chain uniqueness
and correct buckets, Go's deliberate buffer exclusion, and empty/stopped
manager behavior. Buffer exclusion is a Go graph extension; the native
HashStart routine does not accept buffer nodes. All focused hashing tests
pass in 2.020s (`/tmp/abcgo-hash-final-focused.log`). Save another red
checkpoint pending full module race tests and GIA vet. Original completion
remains 2/131; GIA and the sequential audit remain active.

Structural-hash integration is green: full uncached module race tests pass
all 130 tested packages, with three command packages without tests and no
failure/race diagnostic lines (`/tmp/abcgo-hash-review-module-race.log`).
GIA passes in 248.426s, ABCI in 29.784s and BMC in 23.891s. GIA vet passes
(`/tmp/abcgo-hash-review-vet.log`). Occupancy reporting and the initial oracle
are saved in ed6c2aebe; resize/compactness and expanded coverage in 9773c2234.
Source/tests/native fixture stayed unchanged during integration.

This closes giaHash.go's remaining hashing/reduction/helper review alongside
the earlier Gia_ManRehash repair. The summary keeps Go's existing full-width
counters rather than C's signed-int diagnostic truncation; the restored
bucket report follows the original bounds/order. Source scalar construction,
additional strashing, vector reducers and MUX-recognition decisions agree
on the native matrices; typed buffer handling has independent Go coverage.
Continue with the other balancing families in giaBalAig.go/giaBalMap.go,
including their shared helpers, timing/arrival interfaces and area divisors.
Original completion remains 2/131; GIA and the full sequential audit stay active.

AIG balancing review exposes a remaining simple-mode classifier gap.
Gia_ObjIsBuf already correctly excludes opposite-phase repeated-fanin ANDs,
but Gia_ObjIsAndReal also excluded them, leaving valid x & !x nodes outside
the ordinary AND category. New classification and depth/area balancing
regressions fail before repair: both public balancing paths panic in supergate
collection (`/tmp/abcgo-balance-simple-class-red.log`). Recognize non-buffer
ANDs with equal fanin object IDs as ordinary ANDs too; real XOR/MUX and true
buffers remain excluded. The shared collector can then simplify x & !x to
false without a new balancing algorithm.

This continues the earlier intentional Go correction of C's simple-mode
buffer misclassification; it does not claim these cases as valid native
balancing inputs. Ordinary source representations are unchanged. New red
regressions and existing balancing/supergate/ABC comparisons pass in 1.098s
(`/tmp/abcgo-balance-simple-class-green.log`). Save a red checkpoint while
adding broader classifier, depth-helper, arrival-level and source comparisons;
area-divisor and window families remain unfinished. Full integration is
pending. Original completion remains 2/131; GIA and the audit remain active.

Expanded depth-balancing coverage passes. balance_depth_native_review_test.go
and test_vectors/balance_depth.c compare the actual Gia_ManBalanceInt on
2,304 cases/147,456 truth rows: six chain/shared/XOR/MUX/buffer/NOR families,
16 phases, both strict settings, integer/float/both arrival vectors and
And2Delay values 0/2/3. Exact returned structure/type counts and pre-cleanup
source Value mappings agree. Integer arrivals take precedence when both
formats exist, matching C; raw XOR inputs are source-canonical.

A 16-case classifier matrix checks ordinary AND versus constant/CI/CO,
buffer, real XOR/MUX and simple equal/opposite phases. A 512-case public
simple-mode depth/area matrix checks 4,096 rows directly and after AIGER
serialization, including DFS layouts, strict flags, coarsening settings and
zero/nonzero area iteration limits. These simple-mode cases independently
check Boolean correctness beyond the known original C buffer-classification
error, rather than claiming native equivalence for that upstream defect.

BalanceGate's unconditional literal clone is redundant. Resolve the sorted
permutation into its own existing storage before clearing the destination
scratch vector; forward copying is safe for borrowed one/two-literal spans.
The source sorting/pair-selection algorithm is unchanged. Borrowed-span
regressions cover both AND/XOR, 1..12 literals, offsets 0..2 and four phase
patterns on all 256 assignments. In 100-call reuse benchmarks, 2 literals
improve 8B/1 allocation/218ns to 0B/0 allocations/150.3ns; 8 literals improve
96B/2 allocations/1843ns to 64B/1 allocation/1759ns; 12 literals improve
240B/4 allocations/3757ns to 192B/3 allocations/4205ns. Allocation reductions
are consistent; timing does not improve in every case
(`/tmp/abcgo-balance-gate-before.log`, `/tmp/abcgo-balance-depth-optimized.log`).
Focused checks including the native oracle pass in 2.288s. Save a red
checkpoint before the remaining helper/normalization-consumer checks and full
integration. Area-divisor and window reviews remain unfinished; original
completion stays 2/131 and GIA remains active.

Depth-helper native coverage also passes: 528 last-pair preparations and
1,056 shared-node selections exercise four level distributions, 2..12 operands
(including the last-eight-node search bound), rotated inputs and both first
operand phases against the actual giaBalAig.c helpers. Another 48 AND/XOR
collector cases vary 8/64 leaves, strict flags, one/two/three uses and leaf
reference counts, with exact native leaf vectors plus 6,144 independent
Boolean samples (`/tmp/abcgo-balance-pair-native.log`, 1.796s). The other
classifier consumer, structural supergate collection, now correctly reduces
the simple-mode contradiction to a singleton false literal too. Its broader
normalization family is not declared reviewed by this check.

Inspection found another incomplete port: verbose Gia_ManBalance replaced
C's four stage-statistics calls with one invented summary. A new red test
finds zero stage reports (`/tmp/abcgo-balance-verbose-red.log`). Restore
Gia_ManPrintStats before copying and after coarsening, balancing and lowering,
at the same ownership-transfer points as C. The regression checks all four
reports, quiet-mode silence and identical Boolean results/structure. This
restores wrapper wiring; the statistics implementation itself still requires
its own remaining review.

All focused classifier/borrowed-span/depth/helper/verbose/original-ABC checks
pass in 5.141s (`/tmp/abcgo-balance-depth-final-focused.log`). Save a red
checkpoint pending full uncached module race tests and GIA vet. This review
covers the depth-balancing portion and shared collectors, not the remaining
area-divisor and window algorithms. Original completion stays 2/131; GIA
and the full sequential audit remain active.

Depth-balancing integration is green: full uncached module race tests pass
all 130 tested packages, with three command packages without tests and no
failure/race diagnostic lines (`/tmp/abcgo-balance-depth-module-race.log`).
GIA passes in 252.379s, ABCI in 29.472s and BMC in 23.332s. GIA vet passes
(`/tmp/abcgo-balance-depth-vet.log`). Source/tests/native fixtures stayed
unchanged during integration. Classifier repair is in 0cd17b660; scratch
copy reduction and expanded depth coverage in fc9279339; source verbose
stage wiring and native helper coverage in 67bde4adc.

This closes the reviewed AIG depth-balancing portion of giaBalAig.go and
shared supergate/selection helpers, including the intentional Go correction
for simple-mode contradictions. Ordinary/canonical native cases agree on
structure, source copy literals and truth; borrowed spans retain their Go
API behavior with fewer allocations. It does not finish the DAM area-balancing
half of giaBalAig.go or giaBalMap.go. Next compare DAM candidate occurrence
lists, scores, queue extraction, node/divisor updates and reconstruction with
C after each extraction, using overlapping AND/XOR supergates and different
arrival/limit settings before moving to mapped windows. Original completion
remains 2/131; GIA and the full sequential audit remain active.

DAM area-balancing review finds a scratch-ownership bug in collection. C
records only the MUXes it marks and clears those entries after traversal;
Go erased mark0 on every graph object. The new red regression loses a
caller's CI mark immediately (`/tmp/abcgo-area-marks-red.log`). Restore the
source vVisit list, reset it before collection, append only newly marked
MUXes, and clear only those recorded marks. Ordinary/CI/CO/unrelated-MUX
marks are preserved; temporary visited-MUX marks are released.

The first regression passes after repair (`/tmp/abcgo-area-marks-green.log`).
Additional checks include a reachable MUX and pre-marked unrelated MUX,
ordinary-node/terminal marks and reuse of the visited list across collection
calls. Save a red checkpoint while building intermediate-state/native tests
for divisor lists, priorities, queue updates and reconstruction. The whole
area-balancing algorithm is not yet reviewed; full integration is pending.
Original completion remains 2/131; GIA and the sequential audit remain active.

New balance_area_native_review_test.go / test_vectors/balance_area_states.c
compare original DAM state after collection and every extraction, not just
final AIG sizes. All 480 cases/500 updates agree exactly on node-set handles
and live literals, divisor keys/payloads/occurrence lists, float32 score bits,
queue membership/winner, level bounds, gain counters, source marks/visited
MUXes and object levels. Reconstruction also matches exact structure, source
copy literals and 30,720 truth rows. Six overlapping AND/XOR/mixed/MUX/buffer/
wide families vary phases, arrival levels and limits 0/1/2/4/8. Source snapshots
compare meaningful vector prefixes, not unrelated backing-capacity padding.
Native coverage passes in 2.015s (`/tmp/abcgo-area-intermediate-native.log`).

Remove redundant temporary copies using the C storage lifetime rules:
collection appends leaf literals once to vSetStore, then reacquires entries
by index across recursive growth; candidate filtering reuses one local slice;
Update borrows its node occurrence list, which is not extended until the
node-update loop finishes. Exact native states/truth stay unchanged. A 100-call
whole construction/collection/extraction benchmark improves from 23,304B /
223 allocations / 43,653ns per run to 23,024B / 209 allocations / 41,265ns
(`/tmp/abcgo-area-extract-before.log`, `/tmp/abcgo-area-optimized.log`).
Focused ownership/native checks pass in 2.042s after the allocation changes.
Save a red checkpoint while completing wrapper/diagnostic and boundary checks;
full integration is pending. Original completion remains 2/131, GIA active.

Public area-boundary regressions find two metadata losses. Mapped derivation
preserves port counts but loses bad/constraint counts and reset values; restore
the complete unchanged interface immediately after converting the derived graph.
Role regression fails in /tmp/abcgo-area-mapped-roles.log, reset regression in
/tmp/abcgo-area-mapped-resets-red.log, both then pass together in
/tmp/abcgo-area-mapped-green.log. This is a Go metadata correctness extension;
the C mapping converter also does not propagate every Go property field.

A larger public interface matrix exposes reset/comment loss in final normalization
on the first unmapped/coarsened case (/tmp/abcgo-area-public-interface.log).
Restore interface metadata at that area-wrapper boundary too; global normalization
is not changed because other callers may reorder state/interface coordinates.
Save a red checkpoint pending focused rerun and full integration. Original
completion remains 2/131, with GIA and the sequential audit active.

Area diagnostics have two red regressions: public flags omit all four source
stage reports, pair and gain summaries, and incorrectly print the extraction
summary for verbose without very-verbose (/tmp/abcgo-area-diagnostics-red.log).
A native diagnostic mode in balance_area_states.c additionally fails on the
invented DAM/queue/divisor text (/tmp/abcgo-area-diagnostics-native-red.log).
Restore the source pair/divisor totals, tried/used/XOR percentages, selection
labels/operators/polarities/weights, final gain/gainX summary and flag wiring.
Keep time.Now/time.Since elapsed timing in Go; native comparison normalizes only
the inherently nondeterministic Time field. Integer diagnostics use Go int
counters rather than intentionally reproducing C signed-overflow behavior.

All 12 AND/XOR/mixed/MUX/buffer/wide phase families' queue/divisor diagnostics
and 48 internal verbose-mode runs match original C exactly after normalizing
time. Public tests verify four stage reports only for verbose, extraction reports
only for very-verbose, quiet silence, and identical truth under all flag pairs.
The 1,152-case mapped/unmapped/simple/coarsened/arrival/limit matrix now retains
roles, reset values, comments and identity, with 73,728 truth rows checked both
directly and after AIGER serialization. Correct the test's initially invalid
reset encoding 2 to a valid 0; this was a fixture error, not a production fix.

Empty/no-output/constant/CI/buffer/one-AND/contradiction cases with no shared
divisors also terminate correctly and print finite DAM percentages. The general
source graph report averages levels over zero COs and prints NaN on zero-output
graphs in both languages; that separate source behavior is not changed here.
The focused run, including the previous depth verbose guard and 480-case native
intermediate states, passes in 4.265s (/tmp/abcgo-area-final-focused.log).
Save a red integration checkpoint and freeze source/tests/fixtures for full
uncached module race tests plus GIA vet. Original completion remains 2/131,
with GIA and the sequential audit active; mapped-window review remains ahead.

Area-balancing integration is green: all 130 tested module packages pass the
uncached race suite, with three command packages without tests and no failure
or data-race diagnostics (/tmp/abcgo-area-module-race.log). GIA passes in
261.030s, ABCI in 28.860s and BMC in 24.308s. GIA vet also passes with an empty
log (/tmp/abcgo-area-vet.log). Source/tests/native fixtures remained frozen
throughout integration; no unrelated parent files were staged.

This closes the reviewed DAM area-divisor collection/extraction/update and
reconstruction portion of giaBalAig.go, plus the tested ordinary/mapped public
interfaces and source diagnostic wiring. Regression-backed mark ownership is
in df4a7fa54; allocation reductions/native intermediate checks in 4b5e175d3;
interface repair in 8ea123291; diagnostics and expanded coverage in a7e857971.
Public interface coverage here is unboxed; do not infer comprehensive boxed
area integration coverage from it. The whole GIA package remains incomplete.
Next review giaBalMap.go window selection/extraction/insertion, comparing
ordinary-AIG native boundary lists and graph copies, then typed/zero-depth
extensions, logical port permutations, property metadata and mapped-level
selection. Original completion remains 2/131; the sequential audit is active.

Window review finds that ExtractWin omits fGiaSimple, so valid repeated-object
ANDs such as x & !x panic while appending the extracted graph. The new minimal
red test reproduces this (/tmp/abcgo-window-simple-red.log); preserve the source
flag on the combinational window and both boundary modes pass afterwards.

New balance_windows.c invokes the original window routines from ABC, not a
replacement algorithm. All 1,024 ordinary-AIG boundary/extraction cases, 512
insertions and 168 mapped/unmapped level/percentage selections agree exactly on
boundary/gate/input IDs, graph structure, source Value/marks/traversal membership,
logical port lists and exhaustive small truth tables. Layouts include interleaved
CI/CO objects and reversed logical port order. Additional Go checks cover 1,920
typed/layout/subset cases with MUX control dependencies, XORs, buffers, constants,
selected CIs, sequential roles and reset metadata; 960 reconstructions retain
truth directly and after AIGER serialization over 30,720 source assignments.
The zero-depth extraction extension intentionally returns regular boundary
objects, not signed selected COs, consistent with the original window contract.

Eliminate temporary per-node fanin slices from highlight/mark loops by reading
the existing Gate representation. The 1,024-node AND benchmark improves from
74,504B / 4,116 allocations / 126,530ns to 41,736B / 19 / 57,916ns; the MUX
benchmark from 74,504B / 4,115 / 166,962ns to 41,736B / 19 / 57,755ns. Native
and typed checks still pass (/tmp/abcgo-window-optimized.log).

Real registered-engine integration adds 48 SOP/DSD window runs. SOP passes;
DSD changes XOR-related outputs to constants, failing the new truth regression
in /tmp/abcgo-window-comprehensive.log. Extraction/insertion and independent
native tests pass, so investigate the registered IF-owned DSD path rather than
substitute another synthesis algorithm. Save a red checkpoint while tracing
this dependency bug; comprehensive integration is pending. Original completion
remains 2/131, with GIA and the sequential audit active.

The dependency bug is in IF's DSD-cache population, not window wiring. C's
If_ObjPerformMappingAnd calls If_DsdManCompute with If_CutTruthWR (raw cached
truth); Go used ifMapTruth64, which calls If_CutTruthW and applies the cut phase.
The DSD cache key is the regular truth ID and If_CutDsdLit applies the phase on
lookup, so storing the phased truth applies it twice and corrupts Boolean logic.
A minimal two-input direct DSD test fails independently of windows (phase0,
assignment0: got true/false/false/true instead of false/true/false/false) in
/tmp/abcgo-dsd-phase-red.log. Restore raw cached truth via the existing
ifTruthRead helper, with a comment naming the C source convention. No synthesis
engine is replaced. The fix and direct tests are checkpointed in the IF-only
commit cffefb8e5.

The direct regression and all 288 small-function cases pass: all 16 two-input
and 256 three-input Boolean functions plus 16 four-input patterns, with 2,368
assignments checked independently against truth bits and source evaluation.
Existing DSD compact-AIG/pin/manager regressions also pass; focused IF takes
0.691s (/tmp/abcgo-dsd-expanded.log). All window checks, including 48 runs through
the actual registered SOP/DSD algorithms and AIGER round trips, pass in 2.181s
(/tmp/abcgo-window-final-focused.log). IF progress is captured in temporary test
files rather than flooding successful verbose test output.

Public wrapper checks also cover mutually exclusive target/percentage arguments,
no-selection fallback without calling a synthesis engine, source warning/count
messages, independent full-interface duplication and selected-output extraction.
Freeze source/tests/fixtures now for the full uncached module race suite plus
GIA/IF vet. Save an integration checkpoint; the entire IF package is not claimed
reviewed merely because its DSD dependency bug was repaired. Original completion
stays 2/131, GIA and the sequential audit active.

Window and DSD-cache integration is green. All 130 tested packages pass the full
uncached module race suite, with three command packages without tests and no
failure/race diagnostic lines (/tmp/abcgo-window-module-race.log). GIA passes in
267.184s, IF in 8.595s, ABCI in 29.477s and BMC in 23.707s. Both GIA and IF vet
pass with an empty log (/tmp/abcgo-window-vet.log). Source/tests/fixtures remained
unchanged throughout the run, and the working tree contains only unrelated
untracked parent files after the review commits.

This closes the reviewed giaBalMap.go window helper/wrapper family and
balance_window_drivers.go wiring: native ordinary-AIG comparison, typed and
zero-depth extensions, physical/logical layout changes, mapped selection,
property/reset preservation, empty/invalid wrapper paths and actual SOP/DSD
engine integration. The discovered IF DSD-cache bug is repaired by restoring
the source raw-truth convention, with direct red evidence and exhaustive small
function coverage. It does not mark the whole IF package reviewed.

Next review giaBidec.go: reconcile cut-truth conversion and mapped resynthesis
with source, extend width/support and scratch-ownership tests, check logical
CI/CO remapping and property boundaries, and measure any avoidable table copies
before changing storage. No bi-decomposition production change has been made
while integration ran. Original completion remains 2/131; GIA and the full
sequential audit stay active.

Bi-decomposition review finds another physical/logical interface mismatch.
Gia_ManPerformBidec appends CIs/COs in object order but omits the remap used by
other full-interface copies. A new minimal red test reverses logical outputs
and fails immediately (got false/true versus true/false) in
/tmp/abcgo-bidec-ports-red.log. Restore both CI and CO remapping before cleanup;
all input/output reversal combinations then pass. This retains Go's logical
interface semantics even though the original C routine assumes object order.

A separate red regression appends an AND after querying cut truth. New vTruths
entries grow with zero, which is a valid cached-leaf index, so the new node
incorrectly returns aaaaaaaa rather than 22222222
(/tmp/abcgo-bidec-growth-red.log). Use Vec_IntGetEntryFull to extend this
-1-sentinel cache safely. The public Go wrapper already initializes the cache
on first use; this makes subsequent graph growth safe without requiring callers
to know scratch lifecycle details. Both fixes pass in 0.006s together
(/tmp/abcgo-bidec-bugs-green.log).

New bidec_truth.c compares the actual C conversion across widths 0..15,
root/edge phases, reversed leaf order, variable/constant roots: all 768 cases
match, with scratch marks checked cleared (/tmp/abcgo-bidec-native.log,1.903s).
Before storage changes, 100-call warmed benchmarks measure width5 at 68B /
17 allocations / 726ns, width10 at 5,504B / 43 / 19,866ns, and width15 at
266,293B / 65 / 1,036,412ns (/tmp/abcgo-bidec-before.log). Repeated fanin/result
copies dominate allocation; elementary variable generation also loops over each
truth bit. Save a red checkpoint before reducing copies while preserving
independently owned public results. Full integration is pending. Original
completion remains 2/131, GIA and sequential audit active.

Cut truth now uses signed-vector offsets internally instead of allocating and
copying each fanin table and each intermediate result. Offsets survive recursive
vector growth; reacquire the backing array only after all recursion/reservation.
Read edge phases while combining words, retaining uncomplemented cached tables.
Copy only at the public uint32 return boundary, preserving independent ownership.
No unsafe alias is added. Elementary variables use five repeated word patterns
and word-index bits for wider inputs, following the source table construction.

All 768 native width cases still agree. Another 80 typed/phase/permutation/width
cases check 277,368 assignments, shared opposite-phase fanouts, MUX controls,
buffers and independently owned results after later queries, result mutation
and scratch writes. Explicit constant fanins and coincident simple edges also
pass. Internal gate cuts stop at the supplied boundary; a mapped 15-input AND
passes all 32,768 assignments. Reversed ports, sequential resets/properties,
identity/name ownership and source mark preservation pass 64 mapped cases,
1,024 truth rows directly and after serialization. Native end-to-end C
resynthesis agrees on sequential interfaces and 2,032 truth rows across 28
cases; this does not claim exact BDC backend topology or complete BDC review.

Warmed queries now allocate just their public result: width5 4B/1 allocation,
width10 128B/1, width15 4,096B/1, versus 68B/17, 5,504B/43, 266,293B/65 before.
The isolated 100-call after benchmark measures 388ns / 3,545ns / 140,932ns
respectively (/tmp/abcgo-bidec-benchmark-after.log). Timing varies: an earlier
focused width5 run was 906.5ns versus the 726ns baseline, while wider cases
improved. Allocation counts are the clear result, not a universal timing claim.

Two additional red diagnostic checks restore the original separate messages
for fewer than two and more than fifteen LUT inputs, and the before/after AND
count when cleanup removes dangling mapped nodes. Red evidence is in
/tmp/abcgo-bidec-limit-red.log and /tmp/abcgo-bidec-cleanup-red.log. Rejections
return nil without changing copy data or initializing scratch; successful
cleanup preserves truth. All focused Bidec checks pass in 3.856s
(/tmp/abcgo-bidec-final-focused.log). Save a red integration checkpoint and
freeze source/tests/native fixtures for full uncached module race tests and
GIA vet. Original completion stays 2/131; GIA and the audit remain active.

Bi-decomposition integration is green. The full uncached module race suite
passes all 130 tested packages, with three command packages without tests and
no failure/race diagnostic lines (/tmp/abcgo-bidec-module-race.log). GIA passes
in 273.560s, ABCI in 29.646s and BMC in 23.512s. GIA vet also passes with an
empty log (/tmp/abcgo-bidec-vet.log). Source/tests/native fixtures stayed frozen
throughout integration. No unrelated parent files were staged.

This closes the reviewed giaBidec.go conversion/resynthesis boundary family:
regression-backed logical-port and growing-cache repairs (ed1d6fca7), safe
scratch offsets and single-copy public results, variable-pattern generation,
source diagnostics, native truth/interface checks and expanded typed/ownership/
width/maximum-support tests (d75e73805). Go deliberately supports raw typed
and constant-fanin cases beyond the ordinary-AND source converter and retains
owned return slices; those extensions are checked against independent Boolean
expectations. Complete BDC backend review and exact native topology remain
outside this GIA-boundary result and pending in the sequential package order.

Next review giaCCof.go circuit cofactoring. A source/test inventory finds no
direct Ccf_Man/Gia_ManCofTest tests. Reconcile incremental CNF, satisfying-model
cofactors, blocking/image enumeration, return conventions, deadlines/conflict
budgets, frame ownership and zero-state property decisions with C, using small
independent exhaustive transition systems and native checks where source
preconditions permit. No circuit-cofactoring production change has been made
during integration. Original completion remains 2/131; GIA and the full
sequential audit remain active.

Circuit-cofactoring review starts with a false-proof regression. A real XOR
property PI xor RO can fail from the zero initial state in frame0, but the
wrapper reports that it holds. The MUX property PI ? 1 : RO has the same
failure. Red evidence: /tmp/abcgo-ccof-xor-red.log and
/tmp/abcgo-ccof-typed-red.log. The cause is shared frame unrolling: it reads
only two raw fanins and reconstructs every internal gate as AND. Preserve
ordinary-AND source behavior, but dispatch the Go typed extensions through
existing AND/XOR/MUX constructors (hashed or non-hashed as selected), and
forward buffers. Both false-proof cases now fail in frame0 as expected;
existing ordinary-unrolling/initialization/lifecycle tests remain green.

Direct incremental CNF checks pass 9,216 fixed-input wrong-gate contradictions
and 1,024 appended XOR checks across phases, typed gates and logical/physical
layouts. Repeated extension retains clause/variable counts. All 256 one-bit
property/transition functions match exhaustive simulation over three frames:
768 incremental image/status checks and 7,168 assignments, including full/empty
images and repeated convergence. Focused evidence is in
/tmp/abcgo-ccof-baseline.log (0.492s). This reviews the dependent unroller's gate
handling, not its separate scheduling/abstraction algorithms.

Save a red checkpoint while extending raw model-cofactor, reset, resource and
native checks. Inspection identifies a further risk: the verdict uses zero
phases even when Go AIGER metadata supplies one or unknown reset states; add a
red regression before any repair. Original completion remains 2/131, GIA and
the sequential audit active; full integration is pending.

A second circuit-cofactoring false-proof regression confirms that latch reset
metadata is ignored by the verdict: the held-register property RO is reported
safe for reset1 and reset-1 (unknown), even though both admit a bad initial
state. Red evidence: /tmp/abcgo-ccof-resets-red.log. Keep C's zero-state phase
behavior for known resets, seed explicit one resets, and for unknown resets
intersect the accumulated bad-state images with the admissible initial-state
cube using the existing BSAT gate encoding in a fresh solver. The enumeration
solver's blocking clauses must not enter this intersection query. Copy its
absolute deadline and per-call conflict budget; Unknown remains undecided.
This is an exact Go metadata extension beyond C's zero-initial-state contract,
not a substitute image or synthesis algorithm.

All 768 complete one-bit property/transition/reset cases agree with independent
exhaustive three-frame simulation, and 144 two-register reset/image intersections
agree for all two-input functions and zero/one/unknown reset cubes even after
contradictory blocking clauses are inserted into the enumeration solver.
Another 32,768 direct model-cofactor checks cover symbolic register inputs,
fixed/flipped frame PIs, raw XOR/MUX/buffer phases, logical layouts, constants,
CI roots and complemented blocking literals. Evidence:
/tmp/abcgo-ccof-models-resets.log and /tmp/abcgo-ccof-reset-intersections.log.

The actual source giaCCof.c is included in circuit_cofactor.c to expose its
private manager. All 192 proper-image enumeration/repeated-convergence cases
agree with C and independent existential state evaluation
(/tmp/abcgo-ccof-native.log,2.157s). Native cases respect C's nonconstant-image
blocking assertion; full/empty images are checked independently in Go rather
than claiming the source supports them without that assertion issue.
Deterministic expired-clock and pigeonhole conflict-budget checks return
Unknown, then resume correctly. Separate reset queries preserve both limits;
stop releases owned frame/unroller state and leaves the source alive
(/tmp/abcgo-ccof-resources.log). Save a red checkpoint before final cross-frame,
diagnostic and integration checks. Original completion remains 2/131; GIA and
the sequential audit remain active.

Cross-frame coverage passes 1,536 ordinary/typed/layout/phase/save/init/hash
configurations and 176,640 assignments against independent sequential
simulation (/tmp/abcgo-ccof-frame-matrix.log). This verifies the unroller gate
repair across fixed and incremental PI layouts, initialization and non-hashed
construction without claiming review of the separate abstraction scheduler.

New red reporting tests show negative attempted-frame counts on immediate
convergence/limits and a verbose image count/format mismatch against C
(/tmp/abcgo-ccof-diagnostics-red.log). Count attempted iterations/frames as
frame+1, restore the source verbose labels and first cumulative AIG/conflict
counts, and report elapsed seconds using Go time. The native selected proper-
image diagnostic now agrees exactly except nondeterministic time. Keep the
existing Go correction that a zero-second budget means unlimited; do not copy
the source wrapper's erroneous nonzero absolute-stop check for nTimeMax==0.

Compactness: remove unused stored frame/time limits and the redundant encoded
object counter; recover the CNF frontier from NumVars before growing the solver,
as C does. Reserve copy scratch on demand rather than 1,000 entries and avoid
clock/statistic work on quiet image iterations. The small-image benchmark uses
5,918,192B/89 allocations versus 5,922,278B/89 before; timing varies from
1,332,761ns before to 1,385,826ns after, so no speedup is claimed there.

Unknown-reset intersection now reuses a distinct owned solver and extends its
CNF as image gates grow, rather than reconstructing a fresh solver each call.
It still receives current absolute deadlines/per-call budgets and never the
enumerator's blocking clauses. Growing-image checks pass. In 100 warmed queries,
a fresh checker uses 5,915,041B/50 allocations/1,325,799ns each; reuse uses
40B/3 allocations/596.8ns (/tmp/abcgo-ccof-after.log). Both owned solvers are
closed at stop. An end-to-end 21-register test verifies that enumeration can
finish while the reset query hits its own conflict budget and yields undecided,
not a safety claim (/tmp/abcgo-ccof-checker-limit.log). The initial budget1
fixture stopped enumeration too early; budget4 reaches the intended reset-query
path, verified by a returned image plus the limit verdict.

Save a red checkpoint pending final focused rerun, full uncached module race
suite and GIA vet. Original completion remains 2/131; GIA and the audit active.

Circuit-cofactoring false-proof/reset/reporting integration is green: all 130
tested packages pass the uncached module race suite, with three command packages
without tests and no failure/race diagnostic lines
(/tmp/abcgo-ccof-module-race.log). GIA passes in 286.324s, ABCI in 29.241s and
BMC in 24.545s. GIA vet passes with an empty log (/tmp/abcgo-ccof-vet.log).
The final focused suite passes in 7.141s (/tmp/abcgo-ccof-final-focused.log).
Source/tests/fixtures stayed frozen through this run.

The review is not closed yet. Read-only inspection during integration found
that incremental unrolling sizes storage as source objects times the maximum
frame bound. A separate /tmp harness (no tested files changed) proves a tiny
safe circuit with bound1<<30 panics on a nil unroller before it can converge
(/tmp/abcgo-ccof-large-bound-harness.log). C starts incremental storage at a
fixed size rather than reserving every possible frame. Next add a red unit
regression and reserve only the current frame in incremental mode while
retaining the requested bound and fixed-mode checks. Original completion
remains 2/131; GIA and the full sequential audit remain active.

The large-bound regression fails with the reproduced nil-unroller panic in
/tmp/abcgo-ccof-large-bound-red.log. Gia_ManUnrollStart now reserves one source
frame when fSaveLastLit selects incremental construction; fixed-mode capacity
and invalid-bound checks remain unchanged, and NFrames still bounds Add calls.
The tiny constant-safe property now converges immediately with bound1<<30.
A native held-register case also accepts that bound and converges. The first
native fixture used a constant root and hit C's separate cofactor-recursion
assertion; switch only that fixture to a supported held-register property,
keeping Go's constant-root regression. No source assertion is disabled.

Another 256 three-register phased/typed/layout/reset systems agree with
independent exhaustive nine-frame state-set simulation, including known,
mixed known/unknown and fully unknown resets. These exercise complete public
checks beyond the earlier one-bit machines. The final focused suite, including
actual native large-bound execution, all cofactoring checks and existing
unrolling guards/lifecycle tests, passes in 9.621s
(/tmp/abcgo-ccof-final-focused.log). A repeated small-image benchmark measures
5,917,980B/89 allocations/1,346,170ns (/tmp/abcgo-ccof-incremental-after.log);
the functional large-bound regression is the important allocation-limit proof,
not a claimed small-case timing improvement.

Save the final red checkpoint and freeze files for another full uncached
module race suite plus GIA vet, since the constructor changed after the prior
integration snapshot. Original completion remains 2/131; GIA and audit active.

Final circuit-cofactoring integration is green after the incremental allocation
repair. All 130 tested module packages pass the full uncached race suite, with
three command packages without tests and no failure/race diagnostic lines
(/tmp/abcgo-ccof-final-module-race.log). GIA passes in 290.209s, ABCI in 29.185s
and BMC in 24.182s. GIA vet passes with an empty log
(/tmp/abcgo-ccof-final-vet.log). Source/tests/native fixtures remained frozen
through the run; only unrelated untracked parent files remain in the worktree.

This closes the reviewed giaCCof.go CNF/model/image/reset/verdict/lifecycle/
resource/reporting family, including the regression-backed shared unroller
gate and incremental-allocation repairs. False-proof fixes are in 924ad6dfe
and 9fa3e9dde; exact initial-state-query reuse, source diagnostic wiring and
compact bookkeeping in f611ab776; large-bound repair and expanded multi-register
coverage in d6e7a387a. It does not finish GiaFrames scheduling/abstraction or
the entire GIA package. Go extensions for typed gates, constant/full images
and nonzero/unknown resets have independent exact checks; native cases respect
C's supported source preconditions rather than disabling its assertions.

Next review giaClp.go BDD collapse and disjoint-support rebuilding. Inspection
finds an explicit Shannon-construction substitute for source ISOP cube-polarity
selection and SOP factoring; restore the actual source workflow rather than
accept Boolean-only equivalence as proof of faithful porting. Audit BDD variable
order, limits/reference ownership and rebuilt interface metadata with red tests,
native comparisons and independent truth checks. No BDD-collapse production
change was made during integration. Original completion remains 2/131; GIA and
the full sequential audit remain active.

BDD-collapse review confirms the explicit Gia_ManRebuildIsop substitute: it
builds Shannon MUXes and ignores the source's positive/negative ISOP covers,
phase selection, 1,000-cube rejection and SOP factorer. A new majority regression
fails because no selected SOP cover is produced
(/tmp/abcgo-collapse-isop-red.log). Restore the source workflow using the
existing CUDD ISOP implementation, exact small cover counts, positive polarity
on ties, source-order negative/positive/absent cube serialization and the
existing Gia_ManFactorNode. Counts above 1,000 are rejected before factoring;
large counts need no lossy integer conversion. Constants retain the supported
Go fast path; source calls this routine only for prime nodes.

The small cube serializer is local because importing base/abc's corresponding
helpers would create a package cycle. It follows their source traversal using
CUDD's paired-cover cofactors. Rebuild now owns reusable SOP/cube scratch.
The first factoring invocation exposed a boundary mistake in this new adapter:
passing the scratch buffer's terminating NUL into Dec_Factor causes its Go SOP
parser to reject a mixed constant/cube row. Pass only the meaningful SOP bytes;
no DEC change is needed. The regression and existing collapse/rebuild and all
256 three-input-function tests pass (/tmp/abcgo-collapse-isop-green.log).

Save a red checkpoint pending exact native cover/AIG comparisons, variable-order
and resource/ownership checks and full integration. Original completion remains
2/131; GIA and the sequential audit stay active.

The exact native ISOP sweep exposed a shared MVC division omission: source
Mvc_CoverDivideInternal sorts both operands in place, whereas Go grouped them
without sorting. Factoring reuses the dividend after division, so this changed
AIG construction order. A minimal operand-order regression fails before the
repair and passes after restoring the two source sorts. MVC and DEC tests pass;
the dependency repair is committed separately as 6e7b7158f. This is a dependency
fix required by the current GIA review, not completion of the MVC package audit.

The new native fixture initially omitted Abc_Start, which initializes the source
DEC/MVC manager; adding source startup/shutdown fixes that fixture crash. A later
rotated-variable discrepancy came from incorrectly inverting Cudd_ShuffleHeap's
input in the fixture. Both source and Go accept the index at each level. Correct
the fixture, not production code. All 6,264 comparisons now match exact selected
SOP bytes, phase, root literal and every AIG node, with an independent truth check
under all small input assignments (/tmp/abcgo-collapse-isop-native.log).

A new sequential-interface regression then demonstrates rebuild discarding
constraints, bad-output roles, reset states and comments
(/tmp/abcgo-collapse-metadata-red.log). Copy unchanged interface metadata using
the existing helper, including independent owned metadata buffers. Add parity
cover boundary checks: 512 cubes accepted with every assignment verified; 1,024
and 2,048 cubes rejected before graph/scratch mutation. Focused verification and
full integration remain pending. Original completion remains 2/131.

The expanded focused collapse suite passes, including exact native factoring,
metadata independence and both cube-limit sides
(/tmp/abcgo-collapse-expanded.log). Save 4c532e6d4 and freeze source/tests/native
fixtures for integration. Full uncached module race tests then pass all 130
tested packages; three command packages have no tests. GIA takes 295.273s, ABCI
28.831s and BMC 24.027s (/tmp/abcgo-collapse-module-race.log). GIA vet passes
with an empty log (/tmp/abcgo-collapse-vet.log). No repository changes occurred
during that run beyond unrelated existing parent files.

This verifies the ISOP-source-workflow restoration, MVC sorting dependency fix,
and interface metadata repair; it does not close the entire collapse family.
Read-only follow-up confirms source prime rebuilding explicitly references and
releases both original and transferred local BDDs, while Go omits that lifecycle.
Review it together with collapse fanout/reference ownership before enabling
source automatic reordering. Source's verbose statistics are conditional, so an
earlier suspicion of unconditional normal-call printing is not a defect.

A separate /tmp probe confirms a resource-limit inconsistency: with limit zero,
a fresh ordinary AND is rejected but native XOR and MUX are accepted, because
they call unlimited BDD operations. All three small probe cases return to zero
external references after releasing outputs. Add a failing regression for typed
allocation limits before repairing that path; this probe is evidence for the
next step, not a completed fix or a substitute for reference-lifecycle coverage.
Also inspect trivial-DSD fallback semantics and extend metadata checks across
logical-interface permutations, names and AIGER round trips. Original completion
remains 2/131; GIA and the sequential audit remain active.

The typed-gate budget regression fails for both native XOR and MUX with a fresh
zero-node budget (/tmp/abcgo-collapse-limits-red.log). Lower typed graphs using
the existing source-faithful XOR/MUX-to-AND routine before collapse; each AND
then uses the same source allocation check. This removes the unlimited typed
BDD paths without introducing a substitute BDD API. Failure/retry/truth and
external-reference checks pass (/tmp/abcgo-collapse-limits-green.log).

A trivial-decomposition regression demonstrates Go skipping source's duplicate
fallback when DSD has no internal nodes (/tmp/abcgo-collapse-trivial-red.log).
Return the duplicate before allocating a replacement manager in that case.
Unused gates in the Go test distinguish the results; the public Go extension
accepts such graphs, whereas native collapse expects a cleaned graph. The
source Rebuild fallback itself unambiguously returns a full duplicate.

The BDD reference review then finds a concrete retention bug: a temporary AND
that simplifies out of a later result retains projection child references after
collapse and output release (/tmp/abcgo-collapse-refs-red.log). The first fixture
used same-object complementary inputs, which the raw GIA append API rejects;
replace it with distinct fanins (a AND b) AND NOT a before recording the actual
retention failure. Restore source live-root ownership and final-fanout releases,
using local fanout counts instead of mutating caller scratch p.Refs. Deferred
cleanup releases live roots on all failures; returned CO references transfer to
the output vector. Go's dangling-gate extension releases unused results too.
Success and failure checks now return projection references to their baseline.

Focused collapse tests pass (/tmp/abcgo-collapse-lifecycle-focused.log), including
192 signed/typed/interface-layout/reset cases, 6,144 graph/AIGER truth rows,
metadata/name independence and reset preservation. The first matrix encoded
unknown reset as 2; correct the test to this package's -1 representation without
changing production reset handling. A new complete native collapse/DSD/rebuild
oracle matches exact graphs and output literals for all 256 three-input functions
with independent exhaustive truth checks
(/tmp/abcgo-collapse-rebuild-native.log). The earlier 6,264 ISOP-specific native
cases also still pass. Save a red checkpoint pending prime-transfer lifecycle,
source reporting review and full integration. Original completion remains 2/131.

Shared/complemented output, buffer and caller-owned-root tests pass after an
actual BDD reorder, including exact external-reference ownership and final
cleanup. Cached typed gates succeed with budget zero after an allocating run;
zero budget means no new nodes, not automatic failure. Prime transfer coverage
passes before the ownership-bracket adjustment, so that adjustment is explicitly
source-fidelity/lifecycle protection, not a newly demonstrated leak fix. Restore
C's reference brackets around both extracted and transferred local functions;
reordered-source success and undersized-target failure paths remain balanced.

The reporting regression fails because Go reports total manager nodes instead
of reachable shared output nodes and uses i0/o0 names instead of source fake
names (/tmp/abcgo-collapse-reporting-red.log). Restore source report labels,
actual shared regular-node counting (including the common constant), fake-name
use, and both graph-statistics calls in Gia_ManCollapseTestTest. The Go CUDD core
has no exposed Cudd_SharingSize, so use a small local source-equivalent traversal;
its peak-count API remains the existing core's approximation, not a new exact
peak tracker. Restore source automatic-reordering enable/disable calls and paired
ZDD-variable setup. The current core enables a policy flag; this does not claim
that its allocation engine now performs source-equivalent automatic reordering.

Add empty/constant interfaces, standalone final-fanout cleanup, generated-name
boundaries, complemented/shared regular-node counts and large-name uniqueness
coverage. Keep the existing readable decimal fake-name suffix past 259 rather
than C's punctuation-producing single-character overflow; this is a documented
Go diagnostic extension, not an algorithm substitution.

An additional native oracle matches 384 exact wide ISOP/SOP/factoring graphs for
4/17/33/65 fanins, signed leaves and forward/reverse variable orders, with all
selected-variable assignments checked independently
(/tmp/abcgo-collapse-wide-native.log). These exercise multiword MVC cubes past
32- and 64-bit literal boundaries. The final focused suite passes in 7.091s
(/tmp/abcgo-collapse-final-focused.log), including 6,264 small ISOP comparisons,
256 complete native collapse/DSD/rebuild graphs, the wide cases and the interface/
resource/reference/reporting checks. Freeze source/tests/fixtures after the next
checkpoint for full uncached module race tests and GIA vet. Original completion
remains 2/131; GIA and the full sequential audit remain active.

Final collapse integration is green for frozen checkpoint 30699a206. All 130
tested module packages pass uncached race tests; three command packages have no
tests, with no failure or race diagnostic lines
(/tmp/abcgo-collapse-final-module-race.log). GIA passes in 297.624s, ABCI in
29.602s and BMC in 23.926s. GIA vet passes with an empty log
(/tmp/abcgo-collapse-final-vet.log). Only unrelated existing parent files remain
untracked. No source, test or native-fixture edits occurred during integration.

This closes the giaClp.go family review: fake names, minimum two-polarity ISOP
selection/serialization/factoring, 1,000-cube limit, prime transfer, OR/XOR/prime
rebuilds, trivial/failure duplication, interface metadata, collapse live-root
fanout/reference ownership, allocation failures/retries, typed lowering,
standalone dereferencing, and public diagnostic/test wrappers. Native exactness
covers 6,904 selected-cover or complete-rebuild graph cases; independent truth,
metadata, AIGER, resource, ownership and boundary checks supplement them. Repairs
are checkpointed in e2df4a4ab, 6e7b7158f, 4c532e6d4, 18600d605 and 30699a206.
Core peak metrics and automatic-reordering machinery retain the documented core
limitations and still require their own sequential CUDD audit. Constants,
dangling gates, typed gates and metadata preservation are tested Go extensions;
none replaces the source ISOP/factoring algorithms.

Next review giaCex.c's replay/value/extension/unrolling/minimization family,
implemented across cex_replay.go, cex_values.go and cex_minimize.go with
counterexample.go interface validation. Read-only inventory during integration
finds per-frame [][]bool probe storage in place of source's packed bit array,
and a SAT care-minimization workflow that omits source's priority-justification
prepass and differs in two-order/final-core/high-effort handling. The priority
prepass exists in sat/bmc, whose dependency on GIA prevents a direct reverse
import; resolve that boundary faithfully rather than accepting a replacement
algorithm. Compare actual replay and minimization behavior with C and independent
state simulation, adding red regressions before demonstrated bug repairs. No
counterexample production or test change has been made yet. This closes neither
the remaining GIA families nor the whole package. Original completion remains
2/131; GIA and the full sequential audit remain active.

Begin giaCex.c review with cex_replay.go/cex_values.go and minimization inventory.
The probe keeps [][]bool and leaves source's pData2 scratch unoccupied. A new
37-frame state/word-boundary regression fails on that ownership discrepancy
(/tmp/abcgo-cex-probe-red.log). Store packed []uint32 in pData2, remove the
separate Boolean-slice storage, and retain a final-frame bound for checked Go
lookups. Commit scratch only after validating the complete witness; stop clears
it. The regression and existing replay/value/minimization tests pass
(/tmp/abcgo-cex-probe-green.log). Retained probe payload is one bit per object
per frame, rather than one Boolean byte plus each frame's slice header.

A new native replay/extension/probe matrix initially differs at the all-object
extension (/tmp/abcgo-cex-native.log from its first red run). Source
Gia_ManCexExtendToIncludeAllObjects copies every object's mark BEFORE current
frame gate evaluation, retaining stale internal values. Go correctly records
current-frame values. Do not introduce that C bug into Go. The oracle now
explicitly labels and models the source legacy-all snapshot; independently
scalar-simulate and verify the Go current-frame extension. Other native checks
remain exact. A first scalar fixture accidentally reused the saved current-frame
slice as next-frame scratch; copy it before reseeding, fixing the test fixture.

All 2,048 one-register native cases now match replay verdicts, dual outputs,
first/final failure selection, state extension, source legacy-all timing and
zero-initialized packed probing, across all sixteen transitions, four property
pairs, varied initial states/input traces and 1/2/5/37 frames
(/tmp/abcgo-cex-native.log). Go all-object extension is checked separately against
the independent current-frame simulator, including nonzero initial states. This
is an explicit source bug exception, not normalization hiding a failed port.
Save a red checkpoint pending broader state/layout/guard coverage, minimizer
workflow repair and full integration. Original completion remains 2/131.

The minimization native regression fails at table 0x07, assignment zero,
low-effort mode: Go selects mask 0xc, source selects 0xa
(/tmp/abcgo-cex-care-red.log). The existing port always used Satoko assumption
minimization, omitting the source priority mask and final-core workflow. Restore
source priority justification, MF CNF and BSAT final-conflict analysis in both
assumption orders; apply LEXUNSAT only for high effort and retain the earlier
mask on ties. Preserve constant-true Go support when negating the output already
contradicts CNF and the priority mask is empty.

Make the existing BMC priority and ternary-verification implementation available
inside GIA so GIA's original caller need not reverse-import sat/bmc. Source
algorithm bodies are initially copied for a buildable checkpoint; the following
BMC commit will delegate to these shared implementations and remove the duplicate
bodies. Preserve its raw typed-gate lowering and diagnostics. The restored main
minimizer matches all 2,032 native three-input care masks in both effort modes,
with every mask independently checked against all possible input completions
(/tmp/abcgo-cex-care-restored.log). Existing multiframe SAT care tests also pass.
This is source workflow restoration, not acceptance of an alternative minimal
implicant implementation. Full integration and the remaining wrapper/trace/
storage checks are still pending; original completion remains 2/131.

BMC now delegates to the shared GIA priority and ternary-verification routines;
its full package tests pass in 8.049s (/tmp/abcgo-cex-bmc-shared.log), and the
BMC-only consolidation is committed as 70bace4de. No duplicate priority algorithm
remains there. Restore the distinct source Gia_ManMinCex diagnostic workflow,
including reused assumption storage, numeric statuses and both final-core and
LEXUNSAT reports; its report regression fails before repair
(/tmp/abcgo-cex-wrapper-red.log) and passes afterward.

A 37-frame care boundary test exposes missing final simplification in the new
CNF adapter. The earlier guess of different output numbering was wrong: source
Cnf_DataWriteIntoSolver calls sat_solver_simplify before adding the output
clause. Restore that step. Both care effort modes, original diagnostic wrapper
and all 2,032 native combinational masks pass (/tmp/abcgo-cex-simplify.log).

Another 1,024 two-register signed/typed/logical-layout/initial-state traces agree
with 37,888 independent scalar state rows, covering simultaneous register
updates, failure truncation, dual outputs, complete extensions and packed probe
word boundaries. Guard/reuse tests pass. A new malformed-witness allocation
regression fails after allocating 5,113,528 bytes before rejecting inconsistent
frame/bit dimensions (/tmp/abcgo-cex-probe-allocation-red.log). Extract the
existing dimension validation and invoke it before probe allocation too; this
repair is awaiting focused rerun.

The extended sequential native minimizer fixture is currently red at a source
assertion adding its negated output clause. Diagnostic preflight isolates
transition table3, property14, two frames, seed0: the source's unrolled output
literal is 1 even though its conservative priority mask is nonempty. Source
asserts on that forced-true case. Preserve Go support with an explicit proof-based
zero-care result and a regression, then compare all supported native cases and
independently check Go's forced-true cases. Do not silently discard this source
boundary. The sequential oracle and full integration remain pending; save this
red checkpoint. Original completion remains 2/131.

The isolated native assertion is a source forced-true boundary, not a numbering
failure: unrolling yields literal1 although conservative priority justification
can retain unnecessary input bits. Add a regression that fails in Go at this
boundary (/tmp/abcgo-cex-forced-true-red.log), then return an empty care mask when
adding NOT(output) contradicts the already-simplified CNF. This is a proof that
no inputs are needed; it preserves Go's previous constant-output support instead
of reproducing C's assertion. The sequential fixture explicitly preflights the
source CNF and labels those cases forced-true; it does not pretend the C API
returned a mask. Go must return zero care bits there and still satisfy independent
transition simulation. Remaining native API masks are compared exactly.

All 504 sequential care checks pass, comprising 456 exact supported native API
masks and two Go effort modes for each of 24 source forced-true boundaries.
Small-trace input completions are exhaustive, and 37-frame completions are
sampled with fixed cared values (/tmp/abcgo-cex-sequential-fixed.log). The
allocation-before-validation regression also passes after early dimension
validation. A priority-report regression fails for omitted Original/order/Final
labels (/tmp/abcgo-cex-priority-report-red.log); restore all source labels.

For compactness, reuse priority/value/flip-flop/mark scratch across the four
priority attempts and remove a forward-propagation Boolean result that is always
true (source propagation returns void). The small four-order benchmark changes
from 592B/14 allocations/4,507ns to 352B/11 allocations/4,501ns
(/tmp/abcgo-cex-priority-before.log and /tmp/abcgo-cex-priority-after.log). Claim
allocation reduction, not a timing improvement. The source algorithms and tie
rules are unchanged.

The final focused counterexample suite passes in 10.245s
(/tmp/abcgo-cex-final-focused.log). It includes 256 signed/typed/logical-layout
unrollings checked over 70,912 exhaustive input histories, alongside 1,024
multi-register replay traces/37,888 scalar state rows, the 2,048 native replay
cases, 2,032 native combinational masks and sequential care checks. Additional
literal-mask tests pass across 32/64-bit boundaries, duplicate/opposite literals,
selected prefixes, empty masks, dimensions and invalid counts/ranges
(/tmp/abcgo-cex-literal-mask.log). Freeze source/tests/native fixtures after the
next checkpoint for full uncached module race tests plus GIA/BMC vet. Original
completion remains 2/131; the GIA package and sequential audit remain active.

Final counterexample integration is green for frozen checkpoint 07beb1d9f.
All 130 tested module packages pass uncached race tests; three command packages
have no tests, with no failure or race diagnostic lines
(/tmp/abcgo-cex-final-module-race.log). GIA passes in 324.020s, ABCI in 29.174s
and BMC in 23.709s. GIA/BMC vet passes with an empty log
(/tmp/abcgo-cex-final-vet.log). Source/tests/native fixtures remained frozen
throughout; only unrelated existing parent files remain untracked.

This closes the giaCex.c replay/value/extension/unrolling/minimization family
review, with the checked Go counterexample interface ownership/dimension/reset
methods also inspected against their existing tests. Source algorithms now use
packed probe ownership, original priority justification, MF CNF/BSAT setup,
final-core analysis in both orders and optional LEXUNSAT; the distinct original
MinCex diagnostic workflow is restored too. BMC delegates to the shared source
priority/ternary routines rather than carrying duplicate bodies. This dependency
consolidation is not completion of the later BMC package audit.

Explicit source exceptions remain documented and tested: Go's all-object trace
records actual current-frame values, while C's routine snapshots stale gate
marks before evaluation; Go returns a proven zero-care mask for forced-true
unrollings where C asserts. Typed gates, reordered interfaces and checked/owned
Go witness storage retain their existing supported extensions. No source SAT or
justification algorithm is replaced by a compatible alternative.

Next review the chronological circuit-SAT engine in giaCSatOld.c/giaCSatOld.go,
then the subsequent CBS/CTAS families. Existing exhaustive tests and the old
REVIEW_circuit_sat_old.md are useful evidence but predate this sequential audit.
Compare actual source decisions, strict backtrack/frontier limits, counters,
partial-CI model order, queue restoration and miter storage/statistics. A first
read-only suspicion of a missing retry was incorrect: that second solve is
inside a commented-out C block. The active source and Go both try once; no retry
repair is warranted. No circuit-SAT production or test edit has been made during
this integration turn. Original completion remains 2/131; remaining GIA families,
final inventory and the full sequential audit remain active.

Begin giaCSatOld.c chronological circuit-SAT review. Compare the actual Cbs0
engine directly through a native private-API fixture, not the newer CBS command.
Random and hard reconvergent/parity graphs yield 30,432 exact source verdict,
per-query/cumulative backtrack/frontier counter and ordered partial-CI model
comparisons across three heuristics and strict resource limits. Outcomes include
8,627 undecided, 20,998 SAT and 807 UNSAT queries, with maximum 54 backtracks and
frontier size eight (/tmp/abcgo-cbs0-native.log). Independently exhaust every
five-input truth table and every completion of each returned partial cube;
assert assignment/queue cleanup after all outcomes. Core solving matches C.

Miter regressions fail because Go counts constant outputs as solver calls and
sends them through limit checks. With negative budget, Go returns undecided for
constants where C still returns their immediate verdicts
(/tmp/abcgo-cbs0-miter-red.log). Source handles constants before calling the
solver. Restore that branch, trivial-counterexample store encoding and source
constant-one diagnostic. Stats also omitted source call percentages, average
conflicts and time fractions, besides counting those constant cases incorrectly
(/tmp/abcgo-cbs0-stats-red.log). Restore source rows and safe zero denominators.
For lifecycle compactness, Stop cancels any active trail then zeroes the whole
manager instead of resetting selected fields individually. All chronological
focused tests pass in 3.793s (/tmp/abcgo-cbs0-focused.log), including exact native
miter statuses/store/diagnostics at negative, zero and ordinary budgets.

Next measure the whole-graph typed-gate scan repeated on each query. Valid GIA
managers already maintain typed-gate counts; preserve typed rejection while
avoiding quadratic validation cost across many small-cone queries. Save a red
checkpoint pending that compactness review, broader guards/queue/miter tests and
full integration. Original completion remains 2/131; GIA and audit stay active.

Avoid the per-query full graph scan in cbs0CheckGraph: valid GIA construction
maintains XOR/MUX counts, so those counters reject unsupported typed graphs in
constant time. Rejection remains checked after construction, including newly
appended typed gates. A leaf-query benchmark changes from 61.97ns (10 gates) /
21,604ns (10,000 gates) to 31.54ns /30.44ns, all with zero allocations
(/tmp/abcgo-cbs0-validation-before.log and /tmp/abcgo-cbs0-validation-after.log).
These measurements precede the final extra queue-head guard comparisons; the
important result is removal of graph-size-dependent validation, not a guaranteed
nanosecond figure. Native core comparisons remain green after this change.

Another 3,600 native public-miter property statuses/stores match exactly across
CI/CO permutations, random and hard parity/reconvergence graphs and five budgets.
Decode every store record and independently exhaust all completions of each
partial cube; require timeout markers, proof validity and assignment cleanup
(/tmp/abcgo-cbs0-hard-miter.log). Native fake-clock statistics match byte-for-byte,
including percentages/averages/time fractions and all zero-denominator cases.
Additional tests cover nested saved frontiers through backing-slice growth,
partial cancellation, active/invalid-heuristic Stop, foreign/nil/CO roots,
buffers/repeated/constant fanins, original allocate-then-attach setup and guarded
restoration/reuse. Go-only fanin extensions have independently checked witnesses.

Final source-precondition comparison finds Go checking queue tails but omitting
heads. A red regression confirms nonzero empty queue heads are accepted instead
of rejected (/tmp/abcgo-cbs0-head-guard-red.log). Restore all four source idle
queue checks. The prior focused suite passes in 7.459s; final focused rerun and
full integration are pending this last guard repair. Save a red checkpoint.
Original completion remains 2/131; GIA and the sequential audit remain active.

The final chronological focused suite passes in 7.486s after restoring the
queue-head guards (/tmp/abcgo-cbs0-final-focused.log). Freeze e40ae07ee for
integration. All 130 tested module packages then pass uncached race tests, with
three command packages without tests and no failure/race diagnostics
(/tmp/abcgo-cbs0-final-module-race.log). GIA passes in 329.803s, ABCI in 29.343s
and BMC in 23.675s. GIA vet passes with an empty log
(/tmp/abcgo-cbs0-final-vet.log). Source/tests/native fixtures remained unchanged
during integration; unrelated existing parent files remain untracked.

This closes the giaCSatOld.c chronological Cbs0 family: parameter/default and
assignment predicates, frontier/trail lifecycle, heuristic choices, propagation,
chronological branching/backtracking, strict limits/counters, ordered partial-CI
models, standalone solves, miter status/store encoding and statistics. Source
private/public native matrices and independent full completions back that review;
Go root/typed/constant/buffer/lifecycle guards have dedicated tests. Source retry
and alternative phase decisions inside C comments are not active algorithms.
Update REVIEW_circuit_sat_old.md to replace its obsolete isolated-build caveat
with this actual full-module evidence. Repairs are in 241da684f/e40ae07ee.

Next review giaCSat.c's CBS conflict-analysis/resolution engine and paired-root
all-object witnesses, then CBS2/CTAS. Read-only comparison finds its miter wrapper
also routing constants through the solver and using reduced custom statistics;
verify those with new red regressions rather than assuming the old-engine fix
covers it. Compare all source idle queue/clause guards and supported-root handling,
conflict reasons/resolution, limits/counters, witness forms and proof-zero output
patching. Source nBTThisNc increments and phase-based decisions are commented out;
Go leaving that counter zero is not itself a defect. No subsequent-engine source
or test edits have been made. Original completion remains 2/131; remaining GIA
families, final inventory and the full sequential audit remain active.

Begin giaCSat.c CBS conflict-analysis/resolution review. A direct source fixture
compares 63,648 single-root, paired-root and single-root/all-object queries across
three heuristics, backtrack/frontier limits, random and hard reconvergent/parity
systems. Exact verdicts/counters/ordered witnesses match, with 21,368 undecided,
39,209 SAT and 3,071 UNSAT outcomes, maximum 27 backtracks and frontier size nine
(/tmp/abcgo-cbs-native.log). Independently check full five-input truth tables,
nonempty consistency of all-object cubes, implication of both queried roots,
and assignment/reason/queue cleanup after every query. Core conflict analysis
already agrees with source; no solver replacement is needed.

Wrapper regressions fail on constant budgets, omitted phase initialization and
reduced statistics (/tmp/abcgo-cbs-wrapper-red.log). Restore source constant
handling before solver calls, source zero-input phase setup and complete call/
conflict/time statistics. Unlike Cbs0, the newer C source comments out its
constant-one message, so do not introduce that message here. Factor identical
source statistics into a shared formatter; both engines' native fake-clock and
zero-denominator tests match. Another 7,200 miter property statuses/witnesses,
phase vectors and proof-zero patches now match actual C across CI/CO permutations.

Source idle guards include propagation/justification heads and both clause
sentinels; Go checked only propagation/justification tails. Root assignment also
accepted foreign objects and COs despite source's candidate precondition.
Regressions fail before repair (/tmp/abcgo-cbs-guards-red.log); restore the source
idle checks and validate both roots before assigning either. Use a fixed two-root
array, retaining the valid nil-second-root all-object convention; the first guard
fixture mistakenly classified that documented valid form as invalid, so correct
the fixture. Candidate ownership checks prevent modifying a foreign graph.

Cbs_ManPrintClauseNew was an alias for the different old report. A regression
fails on both report forms (/tmp/abcgo-cbs-clause-report-red.log); restore signed
object reporting and source spacing. All new core/wrapper/guard/clause regressions
and the chronological statistics test pass in 7.791s
(/tmp/abcgo-cbs-repaired.log). Restore source reusable vTemp mark scratch instead
of allocating a local visited slice on every conflict; benchmark follow-up and
broader lifecycle coverage are pending. Save a red checkpoint before further
work. Original completion remains 2/131; GIA and the full audit remain active.

Source reusable conflict scratch and a fixed two-root array remove all steady-
state allocations in the hard-query benchmark: single UNSAT goes from
2,208B/37 allocations/16,162ns to 0B/0 allocations/8,336ns; paired UNSAT from
1,168B/21 allocations/9,732ns to 0B/0 allocations/4,454ns
(/tmp/abcgo-cbs-scratch-before.log and /tmp/abcgo-cbs-scratch-after.log).
These figures precede the subsequent constant-fanin predicate repair, which
will receive a final benchmark and native rerun.

A new Go raw-gate regression finds an actual false-SAT bug: the accepted ordinary
AND (constant0 AND a) is reported SAT because propagation assigns the constant
node as an unknown variable (/tmp/abcgo-cbs-constant-fanin-red.log). Source's
usual hashed inputs simplify such fanins before this engine; Go accepts raw GIA
gates too. Treat constant fanins as their fixed literal values, and use those
values in justification classification. The regression passes after repair
(/tmp/abcgo-cbs-constant-fanin-green.log). Keep one local fanin pointer per
predicate instead of repeating object lookup.

An expanded deterministic sweep passes 52,992 single/paired/contradictory-root
queries over raw ANDs with constants, buffers, identical/opposite fanins and
all three heuristics. Every verdict, nonempty model consistency, cube implication
and scratch cleanup is independently checked over all three-input assignments
(/tmp/abcgo-cbs-raw-extensions.log). The first generator's repeated-fanin branch
only produced identical fanins because i was always divisible by four; correct
the generator to alternate same/opposite fanins and add random edge phases.
Use a separate root local instead of mutating the range variable.

Save a red checkpoint before direct native reason-clause snapshots, atomic
invalid-pair/lifecycle checks, final source/benchmark reruns and integration.
Original completion remains 2/131; GIA and the full sequential audit stay active.

Direct reason-clause snapshots now match actual C: relative reason offsets,
decision pivot, clause literal order, restored marks, level/reason storage and
both report formats (/tmp/abcgo-cbs-reason-native.log). Add atomic invalid-pair
rejection, rejection after typed graph mutation, nested saved queue frontiers
across backing growth, partial trail cancellation, CI/all-object model formats,
strict conflict limits and active/repeated stop cleanup coverage.

The combined CBS/CBS0 focused suite passes in 15.604s
(/tmp/abcgo-cbs-final-focused.log), including the native matrices after the
constant-fanin repair. Final hard-conflict benchmarks remain allocation-free:
single 8,279ns/0B/0 allocations, paired 4,982ns/0B/0 allocations
(/tmp/abcgo-cbs-scratch-final.log). Save these checks before freezing all source,
tests and fixtures for uncached module race tests and GIA vet. Integration is
pending; original completion remains 2/131 and GIA is still in progress.

CBS conflict-analysis integration is green: uncached full module race tests pass
all 130 tested packages, with three command packages having no tests
(/tmp/abcgo-cbs-final-module-race.log). GIA vet exits successfully with empty
output (/tmp/abcgo-cbs-final-vet.log). Source, tests and native fixtures stayed
frozen throughout integration. REVIEW_circuit_sat.md records this family's
source fidelity, regression coverage, constant-fanin extension and compactness
results. Original completion remains 2/131: GIA's remaining families and final
inventory are still pending.

Next is CBS2 (giaCSat2.c / giaCSat2.go). Read-only inspection and a temporary
/tmp native probe complete 24,384 single-root/all-object source queries across
both propagation modes and resource limits (17,006 SAT, 6,666 undecided, 712
UNSAT; maximum 23 backtracks and frontier eight). These are C-only baseline
results, not yet evidence of Go agreement. Source paired-root queries can assert
when conflict analysis encounters multiple same-level decision pivots; the
probe identifies seed0, fanout0, budget0, frontier0, query16, phase0 as one case.
Preserve that distinction when constructing native comparisons and independently
check Go's supported paired-root extension. Add red tests for any confirmed Go
bugs before repairing them; review source diagnostics, miter constants, watched
clause growth/cleanup and fanout ownership as well as verdicts and cube validity.

### CBS2 native comparisons and first repair checkpoint

The 24,384-query native matrix now agrees with Go on verdicts, search counters
and ordered CI/all-object witnesses across both propagation modes, backtrack/
frontier budgets and hard reconvergent/parity graphs. Independently enumerate
all completing assignments for every cube; verify queue, watch, mark, reason and
fanout cleanup after each query. The matrix includes all three outcomes, up to
23 backtracks and frontier eight (/tmp/abcgo-cbs2-native.log).

Extending the native snapshots to propagation/failure/clause-conflict counters
exposes a real port discrepancy: PropagateOne delegated to PropagateTwo and
incremented its counter, unlike C's inline implication branch. Both the counter
comparison and native fake-clock statistics fail before repair
(/tmp/abcgo-cbs2-counters-red.log). Restore the original inline implication and
use the shared source statistics formatter. Separate red tests find constant
miter outputs incorrectly subject to negative budgets, missing six idle queue
sentinels, invalid second roots changing solver state, and conflated cube/clause
reports (/tmp/abcgo-cbs2-guards-red.log). Restore constant handling before budget
checks, distinct signed reports and source idle checks; validate both root IDs,
types and static graph size before assignment or clearing the old model. Use
maintained typed-gate counts for ordinary-AND validation.

Restore source reusable vTemp mark scratch; read older clause entries by index
while appending instead of copying their slices at every analysis/resolution.
Use a fixed two-root array. The hard UNSAT benchmark drops from 1,400B/61
allocations/11,868ns to 0B/0 allocations/5,163ns for top-down propagation and
from 752B/25 allocations/8,310ns to 0B/0 allocations/4,079ns for fanout propagation
(/tmp/abcgo-cbs2-scratch-before.log, /tmp/abcgo-cbs2-scratch-after.log).

The expanded native miter matrix compares another 3,600 properties, stores,
drivers and unchanged phase vectors across logical CI/CO permutations and five
budgets. Source CBS2 comments out phase initialization, unlike CBS; retain that
distinction. All CBS2 tests pass in 5.836s before a cosmetic fixture cleanup
(/tmp/abcgo-cbs2-miter-native.log). Save an integration-pending red checkpoint.
Next checks: watched clause growth/ownership and reports, lifecycle/nested
queues, raw constant/repeated-fanin inputs and exhaustive paired witnesses;
then focused and full module race/vet integration. CBS2 and GIA remain open;
original completion remains 2/131.
