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
| 3 | `aig/gia` | in progress | Inventory and graph/manager representation review begun; first adapter regression below. |
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
