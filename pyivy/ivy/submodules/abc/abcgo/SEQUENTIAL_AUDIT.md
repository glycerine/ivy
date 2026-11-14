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
