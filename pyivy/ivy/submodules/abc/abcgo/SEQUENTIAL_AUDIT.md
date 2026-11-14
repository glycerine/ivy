# Sequential package audit

Review every package in `go list ./...` order, one at a time. For each package,
inspect production files and existing tests against the original C, review
correctness and compactness, add a red regression before fixing any uncovered
bug, and run meaningful package checks. Commit coherent package changes.
Use `test_vectors` for fixtures. Prior reviews are evidence to inspect, not
proof that this audit has been completed. No parallel package or agent work.

Baseline: 131 packages (128 with tests, three command packages without tests).
The repository-wide uncached tests passed before this audit.

| # | Package | Status | Evidence |
| --- | --- | --- | --- |
| 1 | `.` | reviewed | Root review below; race tests and vet pass. |
| 2 | `aig/aig` | in progress | Production-file inventory started. |
| 3 | `aig/gia` | pending | |
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

## 2. aig/aig (in progress)

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
