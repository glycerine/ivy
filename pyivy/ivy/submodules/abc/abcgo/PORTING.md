# ABC Go port progress

Goal: complete the ABC port in this module using Go and its standard library.
A passing isolated file test is partial progress, not a passing package or a
completed port. Keep original algorithms represented; do not hide unfinished
files with build constraints, rename them away, or replace them with stubs.

## Parallel ownership

Twenty workers own separate areas: misc/{vec,util,mem,hash,avl,extra,nm,st,tim},
aig/{miniaig,hop,aig,gia}, bool/{kit,dec}, bdd/{reo,mtr,epd}, sat/{bsat,satoko}.
The main thread coordinates dependencies, reviews results and runs integration
checks. Workers do not use git. All work stays in the shared current checkout.
As owners finish a first component, assignments extend into adjacent unported
families: misc/extra and Espresso cube/cover integration, SAT/XSAT, CGlucose,
base/main and AIG forward retiming. Ownership and dependency changes are
coordinated in-thread; no branches or worktrees are used.

## Verification

Initial whole-module build output: /tmp/abcgo-baseline.log.
Full completion requires every package to build and meaningful behavior tests
covering the verification algorithms, SAT solving, graph representations and I/O.
Current worker reports already confirm that many generated packages do not parse.
The complete port is therefore still active and incomplete.

Full package sweep on 2026-11-13 (`GOCACHE=/tmp/abcgo-gocache go test -json ./...`):
119 packages passed, 9 failed, and 3 had no test files (131 packages total).
The failing packages were `base/abci`, `base/acb`, `base/bac`, `base/cba`,
`base/wlc`, `base/wln`, `bdd/cudd`, `opt/fret`, and `opt/lpk`. `opt/fret` and
`opt/lpk` were blocked transitively by raw C syntax in `bdd/cudd`; the other
failures had local or dependency parse/type errors. `base/ver` passed an uncached
package test and its box/gate parsing work was committed as `b5c0bf2f8`.
This is a dated checkpoint, not the current status after ongoing edits; rerun
the full sweep before reporting a new package count.

## Verified integration checkpoint

Main-thread package tests pass for misc/mem, misc/avl and misc/hash.
Workers report complete ports for bdd/epd, misc/nm and sat/satoko/utils/vec;
main-thread verification is in progress. Vector scalar values are coordinated
as int32 to preserve C integer semantics; Go slice indices may use int.
Large graph/SAT/BDD packages remain incomplete. Isolated new graph cores must
be integrated into existing APIs before they count as completed packages.

Main-thread verification additionally passes misc/st, misc/bar and bdd/mtr.
Nine foundation/container packages now pass package-level checks. The main
thread is integrating those scoped paths on the current branch. Remaining
packages and full-module tests are still incomplete; no completion claim is made.

## Current full-module audit

Whole-module `go test ./...` still fails. Latest log: /tmp/abcgo-integration-current.log.
Reported passing packages: 10; failing packages: 101.
These counts describe build/test results, not semantic completion.

## Foundation race check

Main-thread `go test -race` passes for the ten integrated packages.
TIM package builds; manager/serialization tests are being expanded before
commit. Original TIM referenced API inventory missing from Go definitions: Tim_ManBoxForEachInput, Tim_ManBoxForEachOutput, Tim_ManForEachBox, Tim_ManForEachCi, Tim_ManForEachCo, Tim_ManForEachPi, Tim_ManForEachPo, Tim_ManForEachTable.
The eight TIM inventory names above are C traversal macros, rather than missing
functions; callers must use Go iteration over the corresponding slices.

## Latest dependency integration

Main-thread tests confirm misc/parse, misc/tim and aig/miniaig pass; their race
checks pass too. Git staging of this next group was rejected by automatic
approval review due to the earlier no-git instruction; clarification is pending.
Porting work continues without repository mutations by workers.

CUDD now has a tested concrete parser adapter, while original CUDD type/API
integration remains incomplete. ABC network core implements the decomposition
network interface but its original package algorithms still require migration.
MVC and DEC are being verified together; DEC additionally needs Extra NPN tables.
Truth-table work is split by non-overlapping source families between three
workers to unblock vector/graph consumers sooner.

Main-thread integration now also passes `go test ./sat/satoko/...`,
`./misc/extra/npn`, `./misc/mvc` and `./bool/dec`. MVC currently reports no tests;
its dedicated algorithm tests are still required before treating it as verified.
Satoko's original heuristic/limit fidelity audit continues despite passing SAT
correctness tests. DEC passes package tests through the new NPN/MVC dependencies.

## BBL and latest module checkpoint

Main-thread BBL package and race tests pass, including bidirectional original-C
binary interoperability checks. Committed independently as 4ec722484, with the
requested single-sentence commit message under 60 characters.
Latest full-module test log reports 18 passing and 94 failing packages; the full
port remains incomplete. Hop implementations now await vector dependency builds.
Truth and ISOP work has additional nonoverlapping workers to unblock util/GIA.

## Solver fidelity and vector dependency checkpoint

Committed tested Satoko root learning/limit corrections independently as
4f40f0e77. Worker reports 900 freshly compiled C versus Go SAT/assumption results
matched; documented legacy conflict-reporting quirks remain intentional API
behavior decisions. Full subtree tests remain green.
Hop now passes the owner's isolated integration of actual converted dependencies;
workspace package build still requires the four remaining vector files. Hash,
Hsh, Mem and Set have separate owners; completed core vectors are shared rather
than copied into consumers. Truth-table cofactor/count families are implemented
and integrated by their foundation owner; other truth/ISOP families continue.

## Cross-package graph/SAT verification

Added integration/aig_sat_test.go exercising public MiniAIG and Satoko APIs.
All 256 three-input functions are synthesized, Tseitin-encoded and checked under
all eight assignments and both requested output values (4096 queries). The
integration package passes. This verifies these connected implementations only;
it does not stand in for the remaining ABC algorithms or full-module build.

## Full vector and Hop integration

Main-thread full-workspace package and race tests pass for misc/vec and aig/hop.
Committed separately as b65aa8965 and b9e00e9fb. Hop now uses the actual shared
vector package, without isolated dependency copies. Remaining AIG algorithms are
split between the core owner and dedicated dominator/MFFC/duplication workers.
Utility ISOP finalization and remaining GIA/CUDD algorithms continue; the whole
ABC port is still incomplete.

Latest module audit: 22 passing packages and 91 failing packages. Log: /tmp/abcgo-integration-current.log. BDC and IOA now have owners to unblock canonical recording. Standard-library zlib port is assigned. Full scope remains incomplete.

## Mandatory design and bug review

User requested a distinct idiomatic/compactness review and bug-review pass for
every package, including prior ports. All twenty workers were notified; fixes and
regressions must precede further integration. See REVIEW.md for required evidence.

## Reviewed Hop integration

Committed reviewed Hop simplification and cleanup invariants as 5a5869431.
Main-thread race checks passed Hop, vectors, BDC, IOA and zlib at that snapshot.
The newer BDC/IOA/zlib packages still await complete two-pass review sign-offs
before their integration commits. Vector owner continues live slice cleanup;
consumer builds must be rerun after those additional edits. AIG output/fanout
representation review found and corrected direct-output reference mismatches;
whole AIG is still incomplete due remaining algorithm dependencies.

## Current reviewed integrations

Integrated individually on the current branch: standard-library zlib, BDC and
SPFD, IOA/AIGER I/O, MiniAIG, AVL, MVC, NM, ST, parser, TIM, vectors and DEC.
Each package passed its focused race tests before its package-scoped commit.
The GIA owner also converted the null-safe `ABC_FREE` expansion directly in
`abcgo/aig/gia`: 373 sites now clear pointer/slice fields to nil, with string
fields cleared to empty strings and a lexical regression that rejects remaining
live C `free` tokens. GIA as a whole still cannot compile because unrelated
generated drafts such as `giaAgi.go` remain unported.

The latest full `go test ./...` still fails across many unfinished packages.
GIA's focused ported-core race suite passes, and its remaining package blocker
has moved from `giaAgi.go` to `giaAig.c.go`. Placement currently fails in the
retained generated test harness, while its extracted database/geometry tests
pass. Full-module verification must be repeated after the active package ports
settle.

Additional reviewed integrations since the earlier checkpoint: SAToko utility
vectors, SAToko utilities, the SAToko solver and LSAT; CUDD's BDD/ADD core and
REO's ADD support; RPO; plus the mem/bar audit fixes. Each was committed in its
own package-scoped change after focused race verification. The root CUDD API,
full AIG, GIA, placement, extra, bsat and other generated package drafts remain
incomplete.

Kit's complete API surface and DEC-backed `deco` inline graph API are now also
reviewed, race-tested and committed independently. The latest additional race
checks covered the AIG GIA ported-core harness, CUDD core/REO/MTR, SAToko and
LSAT, RPO, DEC/DECO, and Kit. Full-module testing remains the release gate; the
package-level commits establish only their individually tested scopes.

## Latest AIG and CUDD progress

The CUDD core now includes reviewed ZDD set operations and exact big-integer
counting; the combined core/REO/MTR race suite passes. In AIG, the focused
converted-file race suite now passes across Graph, cuts, factoring, pack/repr,
sequential cleanup and simulation, frames, splitting, timing and related
algorithms. A Repr bug found by sequential cleanup tests was fixed: child
representatives are applied at fanouts, matching the C algorithm. All four Scl
modes now reduce the fixture from three registers to one while preserving
sequential traces. These are focused component results; the full AIG package
still contains original drafts, now including Part, PartSat, Repar, Ret and
RetF paths.

GIA's explicit-file race suite passes after adding AGI/support and arena-backed
memory APIs; its live-code scan confirms the reviewed free-and-clear sites now
clear fields directly. Full GIA remains blocked by other generated algorithms.

The standard-library `misc/bzlib` port is now integrated, including its pure-Go
encoder and ABC stream adapters; cross-tool bzip2 interoperability, race tests
and vet pass. The base ABC network port has advanced through MV and checker
APIs; its current full-package parser blocker is `abcDfs.go`.

Subsequent integration: `bool/rsb` is ported and race/vet clean, with original-C
search and counterexample oracles; `misc/vec` hash indexes now remain coherent
after explicit mutation and are race-tested. CUDD's ZDD set operations also
have a permanent checksum fixture derived from the original C implementation
(65,536 operation pairs). The expanded AIG/Scl source-file race suite passes
after a Repr fix; all four sweep drivers reduce three registers to one while
preserving sequential traces. The full AIG package now passes race tests and
vet; full-module verification remains outstanding.

`bool/lucky` is now fully ported and committed after race/vet validation; its
fast/simple/swap implementations match the original C outputs, phase flags and
permutations on 550 differential cases.

`sat/msat` is also integrated as an independent solver, with its own activity,
heap, propagation, restart, learning and assumption logic. Its race/vet suites
pass, including 900 original-C differential comparisons; declaration-only C
interfaces with no implementation in the checkout are documented, not stubbed.

## Expanded `ABC_FREE` cleanup

A project-wide cleanup converted 1,181 expanded null-safe `ABC_FREE` ternaries
across 277 `.go` files: pointer and slice targets now clear to `nil`, while
known string fields clear to `""`. The AIG `pMemory` example disappeared as
part of the actual `Part_Man` pool port; the Go manager owns typed pooled
blocks and `Part_ManStop` clears the manager. A follow-up scan found no
remaining expanded null-safe `ABC_FREE` ternary call sites; the sole textual
match is the macro definition in `misc/espresso/util_old.go`. This audit does
not cover raw, unexpanded C `free`/`ABC_FREE` calls in other draft files.
Owners reviewed the cleanup in their converted files, including AIG, GIA,
Hop, Kit, PartSat, BSAT, set/sparse and placement; reports found no type
mismatches after review. GIA's review found string-valued cell/name fields that
had been cleared to `nil`; those were corrected to empty strings and their
nil comparisons were updated. The full package build remains the semantic
validation gate for each area.

The AIG package now passes full `go test -race ./aig/aig` and `go vet
./aig/aig`. A scratch-mark leak found by an independent test was fixed by
validating clock domains before marking reachability. Retiming, storage and
the rest of the converted AIG package were committed as `ab1b2a780`; full
module completion remains outstanding.

The Espresso sparse row/column/matrix family has since been ported, with
12,800 operations checked against fixtures generated from the original C
implementation and 6,000 randomized mutation checks for link and identity
invariants. The shared minimum-cover family reports 1,000 C-oracle matches and
1,000 independent optimality cases. Set/cube geometry and sparse matrix
invariants are now committed as focused tested subsets. Espresso's compile
frontier has moved through `compl.go`, `cubehack.go`, and `cvrin.go`; the latest
remaining syntax blockers are `equiv.go`, `essen.go`, and `gasp.go`, all assigned
for porting.

Latest whole-module snapshot: `GOCACHE=/tmp/abcgo-cache go -C abcgo test ./...`
reports 41 passing and 75 failing package lines, captured in
`/tmp/abcgo-full-current.log`. This snapshot predates the latest commits and
active owner edits. The full module remains unbuildable and incomplete.

## Active integration after the latest snapshot

The parallel owners have advanced several package frontiers since that
whole-module run: GIA truth/DAU canonicalization and MiniAIG balance have
focused race-tested implementations; base/abc has advanced through MV and
checker APIs and is now blocked by `abcDfs.go`; GIA AIGER codecs pass focused
race/C-interoperability review; misc/extra has replaced its central truth/bit
operator header and cube APIs. Espresso's shared cube, header, set, cofactor,
irredundancy and cover helpers are ported; equivalence/essential/gasping and
input/output groups remain underway. SAT/XSAT, base/main, CGlucose and AIG
forward retiming have explicit owners. These component and
focused tests do not replace the whole-module gate.

Since that update, Espresso set/cube geometry has been ported with an original
two-bit encoding and MV masks; review caught and fixed a C-to-Go
operator-precedence hazard in mixed shift/AND expressions. REO's CUDD comparison
matches original-C SIFT and SYMM_SIFT behavior on 512 three-variable cases.

The Espresso set/cube subset passes its isolated original-C-backed race suite.
Sparse/matrix plus minimum-cover explicit-file race tests and vet also pass;
cover-level sparse reduction matches 120 original-C fixtures, with separate
reviews recorded.

`opt/cut`, `opt/dar`, `aig/aig`, `bdd/reo`, and XSAT pass their scoped race
tests and vet, and have commits. `bdd/cudd/core` passes race tests and vet with
original-C ZDD product and reorder fixtures. Ivy's converted graph/manager/cut
groups pass their connected race harness, with `ivyDsd.go` now the next
full-package blocker. The CUDD ISOP/complement core is being added against
original-C fixtures.

## Current package checkpoint (2026-11-12)

The following additional packages passed package-wide race tests and `go vet`
and were committed independently: `base/exor` (`512ac75e7`),
`misc/espresso` (`92546cc57`), `sat/psat` (`1ee0d2fcc`), `opt/dar`
(`b7ce3e477`), and `bdd/dsd` (`8af6eb080`). Their test fixtures now use the
directory name `test_vectors`; do not introduce `testdata` directories in this
module.

The latest whole-module `go test ./...` run is recorded in
`/tmp/abcgo-full-current.log`: 58 package lines passed and 64 failed. This is a
current integration snapshot, not a completion claim. Active package blockers
include GIA, Ivy, base/abc, base/pla, CUDD/Extra BDD, MIO/SCL, misc/extra, DAU,
CEC/DCH, BSAT and several SAT/proof packages. Several package owners have
reported further component-level progress after this run; rerun the full gate
after their changes settle.

The null-safe expanded `ABC_FREE` audit in `aig/aig` is clear: no converted
production Go file in that package contains `free(`, and the earlier
`pMemory` expansion was replaced as part of the typed `Part_Man` pool port.
This does not establish that all unrelated draft packages have had their raw
C-style `free` calls converted.

## Integration checkpoint (2026-11-12, continued)

Since the preceding checkpoint, these tested subsets were committed on the
current branch: CUDD CAS and cascade helpers (`c7a2b2577`), BSAT proof logging
and interpolation (`cbdf670dd`), GIA MF cut/CNF mapping (`d34de997d`), MIO
library and command adapters (`6383ecff3`), Int2 default parameters
(`95983f709`), Ivy graph/mapping/cut APIs (`55b96fba1`), and Satoko conflict
cost corrections (`da41ab7f9`). Each was checked with race tests and vet; the
Int2 commit covers only its explicit-file component because the full package
still has generated syntax/API blockers. `base/pla` subsequently passed its
full package race/vet gates and was committed as `846c02ed8`. CEC SAT sweep and
correspondence (`1f537cb50`) passed native and sequential race tests plus vet;
the verified CUDD numeric bridge (`a18663189`) passed core/CAS/cascade race
tests and root bridge tests/vet.

The latest whole-module test run is `/tmp/abcgo-full-current.log`: 62 package
lines pass and 60 fail. `base/pla`, `map/mio`, `bdd/cas`, and the connected Ivy
and SAT subsets now pass their recorded focused gates. The full Ivy package,
GIA package, BDD Extra root, SCL, and many algorithm packages remain incomplete.
Owners are actively porting those areas, including SAIG simulation/constraints,
Ivy Fraig, SCL sizing and network APIs, DAU, CEC sequential correspondence,
and SAT/BMC.

The fixture-directory convention is now `test_vectors`; a fresh directory
scan should remain part of integration checks to prevent reintroducing
`testdata`.

## Current integration gate (2026-11-12, latest run)

`GOCACHE=/tmp/abcgo-cache go -C abcgo test ./...` completed with 72 package
results passing and 55 failing; full output is in
`/tmp/abcgo-full-latest.log`. This run includes the newly completed `opt/cgt`
and `proof/live` work, but still has broad generated-code frontiers. It is a
module-wide failure, not a completion gate. The normal `map/if` package test
still fails on C-shaped `ifDec07.go`; its selected-file race/vet manifest only
checks the listed native files and is not package-wide evidence. SCL's package
race/vet gate is green after correcting its original-ABC BLIF fixture, and
DAU's plain package test/vet are green after restricting the C oracle comparison
to defined C inputs; DAU's race and independent review were still running at
this checkpoint. Use `test_vectors` for new fixtures.

## Follow-up integration snapshot

The current `go test ./...` result remains 62 package lines passing and 60
failing (`/tmp/abcgo-full-current.log`). The newly added CUDD ADD numeric
oracle was corrected to match the IEEE-enabled original C build; the complete
`bdd/cudd/core` race/vet suite is green again. Additional tested commits after
the preceding entry include SAIG cone/cleanup (`bfb9ba4ca`), SAT-cost
evaluation (`4aa58ea04`), DCH (`d5f507d4e`), resubstitution SAT clauses
(`601fd3fb2`), BMC packed-field/pivot helpers (`24f25f00a`), and BSAT
consumer bridges (`3c92e21e0`). Further root CUDD numeric tests are in
`b8a9cd8c3`. `base/cmd` now selects the exact MF/Satoko backend by default
(`d6f8d2cf7`). The full port remains active; GIA, Ivy, SAIG, base/abc,
`map/if`, SCL, DAU, Extra, FRET, resubstitution and proof/SAT packages still
have visible incomplete algorithm families.

## Current integration gate (2026-11-12)

The newest whole-module run is `/tmp/abcgo-full-current.log`: 66 package lines
pass and 57 fail. `opt/dau`, `map/scl`, `base/pla`, and `sat/bsat2` now compile
and pass their package gates; the count improved from the prior 62/60 snapshot.
New tested commits since that earlier snapshot include IF cut/truth/count
(`2503403fc`), DAU DSD (`b5241d616`), DAU Tree/Merge (`84060ba42`), DAU NPN
(`76d7ae8cc`), Extra Misc transforms (`ee091073b`), and Extra BDD
unate/threshold/primes (`69974e00b`). Ivy Fraig is committed as `4395f534f`.

The remaining package failures are still broad generated frontiers, with GIA,
Ivy, SAIG, ABC, CUDD, Extra BDD, IF, mapping, FRET, resubstitution, proof, CNF
and Glucose2 among the active areas. Fixture directories were scanned and none
are named `testdata`; use `test_vectors` for any new fixtures.

After that snapshot, the ABC DFS/functional-isomorphism group passed its
explicit-file race/vet gate and was committed (`8bf67da52`). The SCL timing
network and incremental-update foundation passed its 22-file manifest race
and vet gates and was committed (`0aaf68811`); SCL still has buffer/upsize
families in progress. SAIG cone, cleanup, structural-constraint, and I/O
components pass explicit-file checks, while its simulation/phase group has
source-backed fixtures; the full SAIG package remains blocked by other drafts.

## Current integration gate (2026-11-12, refreshed)

The next full command, `GOCACHE=/tmp/abcgo-cache go -C abcgo test ./...`,
completed with 73 package results passing and 54 failing; full output is in
`/tmp/abcgo-full-current2.log`. This confirms that `opt/fxu`, `misc/extra`,
`proof/live`, `bdd/extrab/sets`, `map/amap/formula`, and `map/cov/minimize` now
pass their normal package tests. The parent `map/amap` and `map/cov` packages
still fail on generated draft syntax, and the full `map/if` package still fails
on `ifDec07.go`; these leaf successes do not establish parent-package success.
The full `opt/dau` and `map/scl` package tests pass in this run. The subsequent
DAU race/vet reruns and independent source review also passed. `opt/fxu` passes
its full package race/vet gate, and the full `misc/extra` package test now passes
with source-backed Path/Prime coverage. Major remaining frontiers include GIA,
Ivy, SAIG, base/abc, BDD CUDD/Extra/LLB, mapping, many optimization packages,
proof, and SAT/CNF/Glucose2. This is a progress snapshot, not completion
evidence. A fresh scan after the latest additions found no `testdata` paths;
all new fixtures continue to use `test_vectors`.

## Current integration gate (2026-11-12, newest run)

`GOCACHE=/tmp/abcgo-cache go -C abcgo test ./...` completed with 74 package
results passing and 53 failing; full output is in
`/tmp/abcgo-full-current3.log`. The full `sat/cnf` package is now green, and
`misc/extra`, `proof/live`, `opt/dau`, `opt/fxu`, `opt/cgt`, and SCL remain
green in the normal package test. `sat/glucose2` and `map/if` remain red at
their untranslated wrapper/decomposition drafts. The newest changes add
continued native ports for Ivy and resubstitution, but Ivy and `opt/res` are
still blocked by their package/dependency frontiers. This remains a broad
module failure; the pass count is not a completion claim.

## Current integration gate (2026-11-12, latest)

The newest full run, `GOCACHE=/tmp/abcgo-cache go -C abcgo test ./...`, reports
76 package results passing and 51 failing; full output is in
`/tmp/abcgo-full-current6.log`. `sat/cnf` and `misc/extra` both pass their full
package race/vet gates. SCL buffering and both `opt/fxu` and `opt/fxch` pass
their full package race/vet gates. `map/if`, `sat/glucose2`, Ivy, `opt/res`, and
the root `bdd/extrab` package remain blocked by untranslated package or
dependency drafts. The overall port remains incomplete.

The next whole-module run (`/tmp/abcgo-full-current7.log`) reported 74 passing
and 54 failing package results while active edits were underway. The following
run (`/tmp/abcgo-full-current8.log`) reported 76/53 and exposed intermediate
test regressions. The latest run (`/tmp/abcgo-full-current9.log`) reports 77
passing and 51 failing. The source-verified output-order expectations in
`opt/dsc` are corrected; `opt/dsc`, `proof/live`, and `sat/glucose2` pass their
normal package tests. `map/if` remains blocked on DSD/delay/time dependencies;
its scoped component gates do not imply parent-package success.

The next whole-module run (`/tmp/abcgo-full-current10.log`) reports 79 passing
and 48 failing package results. It now includes ordinary package-test success
for `opt/rwt` and `map/super`, in addition to the previously green `opt/dsc`,
`proof/live`, `sat/glucose2`, `sat/cnf`, `misc/extra`, SCL, DAU, FXU/FXCH, and
CGT packages. Race/vet status for the newly green RWT and Super packages is
being confirmed by their owners. The main integration failures remain GIA/Ivy/
SAIG/base/abc, BDD, mapping, optimization, and proof/SAT families.

The follow-up run (`/tmp/abcgo-full-current11.log`) reports 79 passing and 49
failing results. It caught a new `map/super` manager-serialization test failure;
the owner is correcting that regression. The same run confirms `sat/glucose2`,
`opt/dsc`, and `opt/csw` now pass ordinary package tests; their full race/vet
gates are reported by the respective owners. LLB has moved past its matrix and
constraint frontiers and is being ported through the manager/hint layers.

The latest full run (`/tmp/abcgo-full-current13.log`) reports 80 passing and 47
failing package results. The Super serialization regression is fixed; ordinary
package tests pass for `map/super`, `opt/csw`, `opt/rwt`, `opt/dsc`, `proof/live`,
and `sat/glucose2`, as well as the previously green families. The full `map/if`
package still fails on visible `ifDsd.go` C-derived syntax; that file is restored
and remains in the package while its native DSD engine is being implemented.

The newest module run (`/tmp/abcgo-full-current14.log`) remains 80 passing and
47 failing; its failure set is unchanged from current13. `opt/csw` now has a
separate independent readability/bug review and full race/vet/original-C gate;
no blocking review findings were reported. SAIG has advanced through the dual-
rail wrapper and now reaches `saigDup.go`; the full SAIG package remains open.

The latest whole-module run (`/tmp/abcgo-full-current15.log`) improves to 81
passing and 46 failing; `opt/fsim` now passes its normal package test. `map/super`
also passes full race/vet with original-ABC comparisons. The map/if DSD engine
now parses and its reduction loops have native slice forms, but whole-package
testing currently reaches the remaining `ifSat.go` and timing dependencies;
the DSD delay path still needs full recursive-tree semantics, so the DSD family
is not complete.

The latest full-module gate (`/tmp/abcgo-full-current16.log`) reports 82 passing
and 45 failing. The `opt/fsim` package remains green after resolving a transient
duplicate AIG MarkB API, and `sat/csat` now also passes its normal package test.
SAIG has moved beyond the initial wrapper/constraint parser blockers; a helper
is porting the dual-rail miter family, and another is implementing GLA/PBA over
the proof-core-capable `sat/bsat.Solver2` API. Full SAIG and BMC package gates
remain open.

The latest run (`/tmp/abcgo-full-current17.log`) reports 81 passing and 47
failing. `opt/fsim` regressed only in its original-C oracle build because the C
fixture requests `aig/ssw/ssw.h` while the repository contains the real header
at `src/proof/ssw/ssw.h`; the FSIM owner is repairing the test include layout,
not removing the differential test. The normal `map/if` package remains blocked
by its unported SAT family and recursive DSD/timing integration.

The newest module run (`/tmp/abcgo-full-current18.log`) reports 83 passing and
44 failing. `aig/ivy`, `map/super`, `opt/fsim`, and `sat/csat` now pass normal
package tests, joining prior green families. Ivy's race/vet gates pass, but its
sequential rewrite wrapper still lacks the original latch-crossing resynthesis
path, so the package is not considered semantically complete. SAIG, GIA,
base/abc, BDD, IF, and other mapping/optimization/proof/SAT families remain
unbuildable at package scope.

The next whole-module run (`/tmp/abcgo-full-current20.log`) has the same 83/44
result and unchanged package failure set as current19. SBD's window/LUT
components now have focused native tests and the public BSAT polarity bridge;
the package still has additional source families to port. SAIG miter/duplication families have focused race/vet checks,
while full package parsing remains blocked in other SAIG families. `map/if`
DSD and delay/time selected-file race/vet gates now pass, including recursive
tree timing; the ordinary package gate has moved to malformed `ifSat.go` and
remains red. Amap has advanced past native graph, permutation, and rule files
to the remaining match/merge sources. These component checks do not establish
the whole parent packages.

The next run (`/tmp/abcgo-full-current21.log`) remains 83/44 with the same
package failure set. Amap now has native reader, permutation/rule and Liberty
files with scoped tests, but package compilation still reaches `amapMatch.go`;
SBD and BMC have additional native helper coverage but still cannot clear their
package frontiers until their surrounding families and dependencies build.
After that run, `map/if` also advanced: the native partition/factorization
replacement for `ifSat.go` now passes its expanded selected-file race/vet gate;
ordinary package compilation has moved to `ifSelect.go`.

The latest run (`/tmp/abcgo-full-current22.log`) briefly showed 82/45 because
the sequential-cut test requested a cut whose root canonicalized to a latch.
The Ivy owner corrected the fixture to keep a logic root while preserving the
sequential coverage; the test now passes and exercises cross-latch cut
enumeration and rewrite behavior. A new whole-module run is needed to refresh
the aggregate count after this fix and ongoing changes.

The latest gate (`/tmp/abcgo-full-current23.log`) returns to 83 passing and 44
failing package results; the Ivy test failure is gone. The ordinary Ivy package
test passes, while its final race/vet reruns for the latest cross-latch rewrite
are pending. The rest of the package failure set is unchanged.

The newest run (`/tmp/abcgo-full-current24.log`) remains 83/44. Ivy's full
latch-aware sequential rewrite now passes normal, race, and vet gates with an
exhaustive PI/latch-state source differential. The IF package advanced from
`ifSelect.go` to `ifSeq.go` after replacing the selector and SAT choice
constraints; ordinary package compilation still fails in `ifSeq.go`.

The latest full run (`/tmp/abcgo-full-current25.log`) improves to 84 passing and
43 failing package results. `sat/proof` now compiles as a package with no test
files. The major package failures remain GIA/SAIG, base/abc, BDD, mapping,
optimization, and proof/SAT. The current IF ordinary gate still stops at
`ifSeq.go`; DSD/SAT/selection slices are separately tested but do not establish
whole-package completion.

The newest module run (`/tmp/abcgo-full-current26.log`) reports 86 passing and
41 failing package results. `map/amap` now passes full race/vet with native
matching, merging, reading, library, and Liberty paths. `opt/sim` now compiles as
a package with no test files; behavioral coverage and review remain to be added.
Ivy's package-wide race/vet gates remain green; IF still fails at the later
sequential-file frontier.

The next whole-module run (`/tmp/abcgo-full-current27.log`) reports 84 passing,
39 failing, and 3 packages with no test files. It confirms `opt/sim` and
`sat/proof` package gates pass. `map/amap` regressed: the mux mapping test panics
because mapped mux fanins are not visited before their consumers; its owner is
repairing traversal order. `map/if` still fails parsing in `ifTune.go`. GIA and
`base/abc` parser failures continue to block dependent packages, including
SAIG, BMC, `base/io`, CUDD-related consumers, and several optimizers/proof
packages. The run was performed while workers were editing shared files, so
package failures are a moving snapshot, not a stable baseline.

The next module snapshot (`/tmp/abcgo-full-current28.log`) reports 85 passing,
38 failing, and 3 packages with no test files. `map/amap` passes again after
fixing complemented-reference handling and matching the C invocation's
`fUseMuxes=1` setting. Follow-up ordinary, race, and vet gates all pass,
including the MUX2 C parity fixture. `map/fpga` also has an independently
verified ordinary package pass. GIA now passes the previous `giaBalAig.go`
frontier but still fails parsing in `giaBalLut.go`; `base/abc` still fails in
`abcHieGia.go`. `map/if` remains blocked in `ifTune.go`.

Since current27, `map/fpga` completed its full package port with original-C
comparisons, race/vet coverage, and separate readability/bug reviews. Its owner
has moved on to `opt/nwk`, whose native graph core now parses while the
remaining package families are in progress. The `map/if` owner advanced
`ifTune.go` beyond its prior parser errors; remaining raw C syntax begins
around line 701. More source families are split across `base/abc`, `base/io`,
and SAIG to reduce shared bottlenecks without hiding drafts.

The current29 module snapshot (`/tmp/abcgo-full-current29.log`) remains 85
passing, 38 failing, and 3 packages with no test files. Its parser frontiers are
`giaBalLut.go` in GIA, `abcHieGia.go` in `base/abc`, and `ifTune.go` in
`map/if`. IF has moved beyond the earlier line-701 frontier; current errors
include unconverted loop control around lines 824/827 and C syntax from line
987 onward. These are active source files, not excluded drafts.

Current30 (`/tmp/abcgo-full-current30.log`) reports 86 passing, 37 failing,
and 3 packages with no test files. This is a net gain of one passing package
and one fewer failure versus current29. `map/amap` remains green. SAIG has
advanced to `saigIsoFast.go`; the GIA and `base/abc` frontiers remain
`giaBalLut.go` and `abcHieGia.go`. `map/if` now parses `ifTune.go` but its
ordinary package gate is still blocked by missing external type/API wiring,
unexpanded GIA iteration, and compile errors. LLB has passed an isolated
scratch check, but its real package gate remains blocked by GIA and is not
counted as green. The scratch directory was removed after checking.

Current31 (`/tmp/abcgo-full-current31.log`) remains 86 passing, 37 failing,
and 3 packages with no test files. `giaBalLut.go` now parses and the real GIA
and IF gates advance to `giaBalMap.go`; its raw C syntax is the current shared
parser blocker. `base/abc` remains blocked in `abcHieGia.go`. IF's previous
unexpanded GIA loop macros were replaced with native iteration, but the package
cannot reach its next type-check until GIA builds. The completed Amap package
has independent normal/race/vet passes and a separate bug/readability review.
After current31, the LUT-balancing owner completed the full native algorithm;
48 original-ABC combinations matched truth, AIG node count, and depth, with
additional exhaustive/random tests and separate reviews. The ordinary GIA
gate still stops at `giaBalMap.go`. `opt/mfs` passes its ordinary package test;
the Win/Div original-source differential now passes, while package race/vet and
independent review are in progress. The SAIG parser frontier has moved past
`saigIsoFast.go` to `saigRefSat.go`. `abcSop.go` now has a disjoint owner to
unblock `abcNtk` validation; `base/abc` still fails at `abcHieGia.go`.
The AIGER codec/adapter package `aig/ioa` independently passes race tests and
vet, including constant polarity, latch initialization, compression, and
round-trip regressions; `base/io` integration remains blocked by `base/abc`.

Current33 (`/tmp/abcgo-full-current33.log`) reports 86 passing, 37 failing,
and 3 packages with no test files; the temporary LLB scratch package is gone.
GIA has advanced from `giaBalMap.go` to `giaBidec.go` and `giaCCof.go` parser
errors. `base/abc` remains at `abcHieGia.go`, and `base/abci` has raw C syntax
in `abc.go`. `opt/mfs` passes the ordinary package gate, and the independent
review's three behavior findings have been fixed and regression-tested; final
review completion is pending. `map/if` parses its own `ifTune.go` work but
remains blocked before type-check by GIA's current files.
Follow-up reviews fixed two additional issues: LLB group construction now
encodes output-function equivalence (with a regression), and the AIGER CEX
writer now preserves ABC's repeated register-state prefix on later frames.
Their real `bdd/llb` and `base/io` package gates remain blocked by shared
dependencies and were not inferred from focused tests.

Current34 (`/tmp/abcgo-full-current34.log`) reports 87 passing, 36 failing,
and 3 packages with no test files. This gains one passing package versus
current33. The real GIA gate remains in `giaBidec.go`/`giaCCof.go`; `base/abc`
remains in `abcHieGia.go`, and `map/if` is blocked by GIA. `opt/nwk` source
files now all parse but its package gate is blocked by those GIA and SAIG
dependencies. `abcSop.go` has a full native API and a 500-case C differential;
`abcNtk` focused tests now run against it, with one buffer-composition bug under
repair. MFS passes the ordinary package test; its race/vet and separate review
are pending. `aig/ioa` passes race/vet, while integration into `base/io` awaits
the ABC parser frontier.

Current35 and current36 both report 87 passing, 36 failing, and 3 packages
without test files. Current36 advances GIA's ordinary parser frontier to
`giaCSat.go`; SAIG is blocked in `saigRetFwd.go`. `base/abc` remains at
`abcHieGia.go`, and `base/abci` remains raw in `abc.go`. The MFS package now
passes ordinary tests, race tests, and vet after its independent review fixed
source-fidelity issues in BDC WinMax, PowerResub traversal, sequential
probability sampling, and AIG cone import. `map/if` parses its source but is
still blocked before type-checking by GIA. CUDD core race/vet remain green,
while the root CUDD package is still blocked in its I/O files.

After current36, ordinary GIA testing confirms the frontier is now
`giaCSat.go`; GIA remains unbuildable. MFS is fully signed off with normal,
race, vet, original-source comparisons, and independent review. The complete
native `abcNtk` selected-source suite now passes 21 source files against actual
component files, including race, metadata/induction handoff, and C-oracle
cases; the enclosing `base/abc` package still fails in `abcHieGia.go`. The
completed workers have been reassigned to `opt/sfm` and `base/abci/abcSat.go`;
additional native `base/io` formats are in progress.

Current37 (`/tmp/abcgo-full-current37.log`) remains 87 passing, 36 failing,
and 3 packages with no test files. `bdd/extrab` now independently passes
ordinary tests, race tests, and vet after completing its remaining families and
fixing a K-map dimension overflow. It is a full package green. The module
frontiers remain GIA `giaCSat.go`, SAIG `saigRetFwd.go`, `base/abc`
`abcHieGia.go`, and `base/abci/abc.go`.

Current38 and current39 remain 87 passing, 36 failing, and 3 packages with no
test files. `base/abc` has advanced past HieGia to `abcHieNew.go`; GIA is at
`giaCSat2.go`, and SAIG at `saigRetMin.go`. `bdd/extrab` is fully green across
normal, race, and vet gates. MFS and its independent review are also complete
and green; Amap's full gates remain green. `base/abci` still has raw
preprocessor/C syntax in `abc.go`, blocking dependent packages.
An independent `abcgo/base/abc/test_ntk.sh -race -count=1` run passes its
actual-source selection (21 native ABC files). The separate review found and
sent additional metadata fixes; this selected gate does not replace the
ordinary `base/abc` package gate. `map/if` now has no syntax errors of its own,
but cannot compile until GIA's current dependency frontier is cleared.

Main-thread build-first pass

Work has resumed with only the main thread. Ordinary `go build -p 1 ./...`
currently fails; diagnostics are saved in `/tmp/abcgo-main-build.log`. Shared
blockers include GIA `giaCTas.go`, SAIG `saigRetStep.go`/`saigStrSim.go`, ABC
latch/library code, and the generated `base/abci` command file.

The basic latch routines have moved from `abcLatch.go` into native `latch.go`:
self-feed detection/removal, pipeline insertion, initialization vectors, mux/DC
initialization conversion, one-hot encoding, class-enable restoration, and
counterexample replay. `latch_test.go` checks cycle termination, complemented
pipeline fanouts, initial states, DC mux truth, exhaustive two-register one-hot
transition equivalence, simultaneous counterexample state updates, and restored
complemented enable controls. The actual-source ABC development gate includes
these files; its normal and race tests pass. This is still not a full package
pass: three class-retiming routines remain in `abcLatch.go`, and the ordinary
ABC build also reports `abcLib.go` syntax errors. No package completion or
commit is claimed for this work.

The main-thread `abcLib.go` conversion now includes hierarchical local-AIG
strashing with explicit complemented edges and box input polarity. A focused
race regression checks all XOR input combinations, complemented output, and
polarity-mask copying. The selected ABC race suite passes. Ordinary
`go build -p 1 ./base/abc` still fails in the remaining `abcLatch.go` class
retiming and raw `abcMinBase.go`; this is not a package-wide green gate.

The main-thread ABC port now has native BDD support compaction and duplicate
fanin substitution in `minbase.go`, including complemented duplicate inputs.
Focused actual-source tests pass with race detection. `abcMinBase.go` still
contains its remaining collapse/elimination C routines and preprocessor blocks;
this selected test does not establish a `base/abc` package build.

The main-thread `abcMinBase` work now also includes native fanin-union and
permutation helpers plus BDD node composition/collapse. An exhaustive
three-input regression covers complemented edges before and after collapse;
deleting BDD nodes now releases their owned function references. The selected
ABC race suite passes. `abcMinBase.go` still contains raw C/preprocessor code
for the remaining support and elimination families, so the ordinary package
build is not yet green.

The native minimum-base work now also composes and collapses local Hop logic,
using simultaneous input mapping that retains complemented edges. An exhaustive
three-input before/after regression passes under the selected ABC race suite.
The ordinary package gate remains blocked by the remaining C in
`abcMinBase.go` and by class-retiming code in `abcLatch.go`.

The Hop elimination pass now uses native DFS, occurrence limits, XOR exclusion,
fanin-size checks, and local-function collapse. The selected ABC race suite
passes an exhaustive regression that checks truth preservation and both sides
of the support-size limit. The ordinary `base/abc` build remains blocked by
other `abcMinBase.go` C/preprocessor sections and `abcLatch.go` retiming.

The main-thread BDD network elimination path now uses the native collapse and
local BDD reorder engine, including dangling-node cleanup and fanin-size
limits. An exhaustive complemented-edge regression checks both accepted and
rejected collapse sizes; the selected ABC race suite passes. The remaining C
support/elimination functions and conditional directives still prevent the
ordinary `base/abc` package from building.

`abcMinBase.go` now parses as Go: its BDD support traversal and special
elimination path have been moved to native implementations, and the obsolete
C preprocessor fallback was removed. Selected ABC race tests pass, including
support traversal without node marking and exhaustive special-elimination truth
preservation. The ordinary `base/abc` build now reports syntax errors in
`abcLatch.go` and `abcNames.go`; it has not reached type-checking.

The main-thread names port now covers object name generation/assignment,
prefix/suffix and dummy names, terminal name transfer, fanin/fake-name
collections, and CI/CO name lists. The selected ABC race suite passes,
including unique-name collisions and netlist terminal transfer. The ordinary
ABC build still fails in the remaining `abcNames.go` ordering/ID routines and
`abcLatch.go` retiming; no package-wide build is claimed.

The ABC name-ordering port now sorts PI/PO/latch interfaces by name and can
transfer CI/CO order by matching names, without storing temporary strings in
object pointers. Selected race tests cover latch terminal pairing and reordered
interfaces. The ordinary `base/abc` build remains red in the remaining
`abcNames.go` name-maintenance routines and `abcLatch.go` retiming.

The ABC name-maintenance port now covers dummy/short/clean/moved names,
CI/CO buffer redirection with edge polarity, and the name-ID file lifecycle.
`abcNames.go` now parses as Go. Selected ABC race tests pass, including latch
prefix collisions, complemented buffer redirection, and name-ID file transfer.
The ordinary ABC build now reports `abcLatch.go` and `abcNetlist.go` syntax
errors; package-wide type checking and tests are still pending.

The ABC primary-output buffer operation now uses explicit Go edges and retains
output polarity. Its targeted regression test passes in the selected ABC race
suite. This does not clear the remaining `abcNetlist.go` and `abcLatch.go`
syntax errors; the complete package build has not yet passed.

Ported the simple-output check and the traversal-order repair routine, including
buffer insertion and optional driver duplication. The selected ABC race suite
passes new polarity and shared-driver tests. The full `base/abc` build remains
blocked by untranslated syntax in `abcLatch.go` and `abcNetlist.go`; this is
not evidence that the package builds.

The main simple-output conversion now handles constant drivers, shared
inverters, polarity changes, and output-name conflicts using native edges.
Selected ABC race tests pass truth-preservation and shared-inverter cases.
The complete `base/abc` package remains unbuilt while its other C-derived
files are ported.

Ported logic-network cleanup and retained-node reduction, a prerequisite of
logic-to-netlist conversion. The selected ABC race suite passes a regression
covering a dangling chain and its retained output cone. The netlist round trip
and full `base/abc` build are still pending.

Ported logic/netlist conversion and AIG-to-SOP conversion, and corrected terminal
name copying from netlist nets. The selected ABC race suite passes a
logic/netlist round trip and a complemented-edge AIG truth-table test. The
ordinary `base/abc` build remains red in the untranslated bench/NAND netlist
conversions and latch-retiming code; package-wide type checking is pending.

Ported the bench and NAND AIG-to-SOP conversions and the bench netlist entry
point; `abcNetlist.go` now parses. Selected ABC race tests pass multi-level
truth tables and verify that the NAND conversion does not alter source edges.
The full `base/abc` build now stops on untranslated latch retiming and
preprocessor content in `abcShow.go`; it has not reached type checking.

Ported flop-dependency DOT generation from `abcShow.go`, using explicit latch
support and propagating write errors. A selected ABC race test checks primary
input counts and inter-latch edges. Other visualization routines and the
latch-retiming code remain unfinished, so `base/abc` still does not build.

Ported visualization filename sanitization, configured Graphviz/viewer
launching, and the flop-dependency show entry point. The selected ABC race
suite passes filename and failed-Graphviz input-preservation checks. Remaining
`abcShow.go` visualization and latch retiming are still untranslated; the
whole package remains unbuilt.

Ported network fanin/fanout statistics and AIG choice counting, including the
choice check used by visualization and bench conversion. Selected ABC race
tests pass for node counts and choices. The full package remains blocked by
other untranslated `abcUtil.go`, `abcShow.go`, and latch-retiming code.

Moved the native DOT writer into `base/abc` to remove the import cycle for
network visualization; `base/io` retains forwarding entry points. Ported the
network show routine and verified DOT generation against a copied logic
network in the selected ABC race suite. The `base/io` package cannot yet be
validated as a whole because its `base/abc` dependency still does not build.

Ported fanin-first cone collection for cut visualization, including shared
fanin handling without changing object marks. The selected ABC race suite
passes reconvergent-cone and leaf-inclusion tests. The reconvergence-driven cut
manager remains in untranslated `base/abci/abcReconv.go`, and package-wide
builds remain red.

Ported the one-level reconvergence-driven cut manager and containing-cut
construction into `base/abc`, avoiding caller mark mutation. The selected ABC
race suite passes a reconvergent cut and containing-cone test. TFO candidate
collection and integration with cut visualization remain pending; the whole
package still does not build.

Ported cut TFO candidate collection and connected reconvergence-driven cut
visualization to the native DOT writer. Selected ABC race tests pass candidate
exclusion, level-limit, and highlighted-cut output checks. The full ABC build
still stops on the CUDD visualization preprocessor block and latch retiming.

Added a native BDD/ADD DOT writer to `bdd/cudd/core`, including complemented
edges, quoted labels, and manager ownership validation. The complete core
package passes `go test -race ./bdd/cudd/core`. ABC's BDD visualization entry
points still need integration, and `base/abc` remains unbuilt.

Ported local and global BDD/ADD visualization onto the core DOT writer;
`abcShow.go` now parses. Selected ABC race tests pass local and network BDD
output-name cases in both representations. The ordinary `base/abc` build now
reports untranslated latch-retiming and `abcUtil.go` syntax; it has not yet
reached package-wide type checking.

Ported SOP cube/literal/factoring, BDD/AIG size, and shared-node statistics,
plus network attribute ownership and cleanup. Selected ABC race tests pass
SOP/BDD counts and attribute-manager lifetime checks. The whole `base/abc`
build remains red in `abcUtil.go` clause counting and latch retiming.

Ported BDD ISOP clause counting, mapped gate-area accounting, and twin-gate
lookup. Selected ABC race tests pass XOR/constant clause counts and mapped
barrier accounting. The full `base/abc` build now advances further into
`abcUtil.go` but remains red there and in latch retiming.

Ported network scratch/copy management, mark clearing, and fanin/fanout
queries, preserving complemented copy edges and handling shared hierarchy.
Selected ABC race tests pass copy restoration and output-name/polarity cases.
The full `base/abc` build now advances further into `abcUtil.go`, but remains
blocked by untranslated utilities and latch retiming.

Ported level-ordered object insertion, level comparators, fanout counts, and
object/CI collections. Selected ABC race tests pass for ordering, duplicate
suppression, and counts. The full package build still fails in remaining
`abcUtil.go` ID-reassignment/cross-cut code and latch retiming.

Ported AIG object-ID reassignment with name, attribute, name-ID and SCL mapping
remapping, followed by structural rehashing. The selected ABC race suite passes
an ID-hole regression covering names, attributes and hash reuse. Other
`abcUtil.go` routines and latch retiming still prevent a package build.

Ported copy-map composition and cross-cut traversal with explicit counters,
preserving complemented copy edges. Selected ABC race tests pass for composed
polarity and a shared-node cross-cut. The complete ABC build remains red in
other `abcUtil.go` routines and latch retiming.

Ported cone-comparison reporting, choice-support checks, constraint inversion,
buffer fanout tests and the truth-table command generator. Selected ABC race
tests pass cone-copy preservation and constraint/fanout behavior. The full
ABC build remains blocked by buffering code in `abcUtil.go` and retiming.

Ported direct and reverse level-balancing buffer insertion, including shared
buffer chains and complemented-edge preservation. Selected ABC race tests
pass truth-table and level-alignment checks in both directions. The complete
ABC build still fails in remaining `abcUtil.go` conversion code and retiming.

Ported normalized fanin-delay calculation and SOP/ESOP cube decomposition.
Selected ABC race tests pass two-input delay and complemented XOR truth tables
for OR and XOR cube roots. The full build still stops at reverse-topological
utilities and latch retiming before package type checking.

Ported the reverse-topological index and fanout traversal, with ID
reassignment invalidating the index. Selected ABC race tests pass for a
reconvergent fanout cone. The full package remains red in PLA and other
`abcUtil.go` conversions and latch retiming.

Ported multi-output PLA construction through the native FXU adapter and local
AIG strashing, plus its demonstration BLIF output. Selected ABC race tests
pass all input combinations for the reference PLA. The full `base/abc` build
still fails on SOP splitting and other utilities, plus latch retiming.

Ported SOP splitting, topological-order checking and phase transfer. Selected
ABC race tests pass exhaustive positive/complemented split-cover truth tables
and phase-copy checks. The complete `base/abc` build remains red in remaining
property-monitor utilities and latch retiming.

Ported the single-output property-monitor network derivation and added native
BDD/Hop support to inverter and AND node construction. Selected ABC race tests
pass monitor truth tables for SOP, BDD and local-AIG networks. The subsequent
strashing and sequential latch-sweep monitor wrapper remains untranslated;
the complete ABC package still does not build.

Ported AIG mux/XOR structural recognition, mux control/data extraction,
statistics and fanin/fanout collection. The selected ABC race suite passes
mux truth-table and XOR recognition checks. Full ABC package compilation is
still blocked by unfinished synthesis and monitor/retiming dependencies.

Ported mapped-gate operator classification and the simple-output diagnostic.
Selected ABC race tests pass mapped buffer and AND classification. The full
ABC build remains blocked by synthesis/GIA code and latch retiming.

Started porting `giaCTas.go` by translating SAT-pattern packing to Go slices.
Focused GIA race tests pass conflict atomicity, compatible merges and bit-slot
selection. This does not build the GIA package: its circuit SAT solver and
other C-derived files remain untranslated.

Ported the GIA circuit-SAT queue and offset-addressed clause storage to Go
slices. Focused race tests pass queue snapshot/restore, clause termination,
handle stability across growth, and reset behavior. The remaining circuit-SAT
solver still uses untranslated C and the full GIA package remains unbuilt.

Ported the circuit-SAT manager's defaults, allocation, assignments, literal
interpretation, model collection and reason/handle accessors onto native queue
and clause storage. Focused GIA race tests pass for these interfaces. Solver
propagation, search and miter drivers remain untranslated, and the full GIA
package does not build.

Ported circuit-SAT decision heuristics, assignment trails, typed clause/node
reasons and backtracking. Focused GIA race tests pass heuristic selection and
trail/decision-level restoration. Propagation and search remain untranslated;
the full `aig/gia` package is still unbuilt.

Ported circuit-SAT conflict derivation, resolution, watched-clause creation
and watched-literal propagation. Focused GIA race tests pass watch movement,
unit implication, clause reasons and a decision-reason conflict. Graph
propagation and solver search remain untranslated; GIA still does not build.

Ported circuit-SAT graph and justification propagation. Focused GIA race tests
pass true-AND fanin implication, complemented edges, and false-AND
justification resolution. Recursive search and miter drivers remain
untranslated; the full `aig/gia` package still does not build.

Ported recursive circuit-SAT search, assumptions, model extraction and solver
reset. Focused GIA race tests solve all 256 three-input truth tables and their
complements, validate witnesses, and reject contradictory assumptions. Miter
drivers and other C-derived GIA files remain unfinished; the package is not
buildable yet.

Ported the per-output circuit-SAT miter driver, status bytes, counterexample
store and statistics. Focused GIA race tests pass constant, satisfiable and
nontrivial unsatisfiable outputs and validate stored witnesses. The second
pattern-packing miter driver and other untranslated GIA files still prevent a
full package build.

Ported the pattern-packing miter driver and clause diagnostics; `giaCTas.go`
now parses. Focused GIA race tests pass SAT pattern packing and proved-root
marking. The ordinary GIA build now advances to untranslated `giaCex.go` and
has not reached package-wide type checking.

Counterexample replay progress: ported GIA's final-frame verification,
failed-output search, and first-failure selection to the Go graph evaluator.
Focused race-tested coverage checks sequential register propagation, dual
outputs, and zero-initialized first-failure search. The complete `aig/gia`
package still does not build: `giaCex.go` retains unfinished C-style
counterexample value and minimization routines, and additional files may
also fail once those are ported. The focused test is not a package build.

GIA counterexample follow-up: ported per-frame object probing, counterexample
expansion, sequential unrolling, and SAT-based care masks. The all-object
expansion now records the computed values of the current frame; the C source
records old internal/output marks before computing that frame. SAT care masks
are minimized by deletion in one or both orders rather than using the original
C solver's core and non-SAT prepass. Focused race tests include exhaustive
checks that selected care bits force the output in a small sequential graph.
The full `aig/gia` build now reaches `giaClp.go`, whose conditional CUDD/DSD
collapse implementation is still C; further package errors may remain.

GIA BDD collapse follow-up: `giaClp.go` now constructs BDDs using
`bdd/cudd/core`, decomposes through `bdd/dsd`, and rebuilds a GIA including
register outputs. Prime components are rebuilt using memoized BDD Shannon
expansion rather than the C implementation's ISOP/SOP factoring and 1,000
cube cutoff; this preserves logic but may change optimization results and
resource use. Focused race tests cover all 256 three-input Boolean functions,
multiple outputs, complements, constants, and sequential interface retention.
The complete `aig/gia` build now fails in `giaCof.go`, which still contains
C pointer-arithmetic-based cofactor network code. The legacy `bdd/cudd`
package independently fails to parse; `bdd/dsd` and `bdd/extrab` tests pass.

GIA cofactor follow-up: replaced the packed C cofactor network with explicit
Go object links, ported cone/support traversals, fanout and MFFC analysis,
cofactor reconstruction, and repeated cofactoring. Focused race tests exercise
connectivity, MFFC reference restoration, cofactor simplification, and
exhaustive small-graph equivalence. The Go fanout histogram prints exact counts
rather than the C diagnostic's decade buckets. The full `aig/gia` build now
reaches unfinished C-derived code in `giaCone.go`; the focused suite is not a
package build and further errors remain likely.

GIA cone follow-up: ported output-group merging, sequential cone and flop
counts, pivot signatures, output signature grouping, and largest-partition
reconstruction with trimmed inputs and retained registers. Focused race tests
cover sequential register closure, grouping, output equivalence, and selecting
the largest partition. The two exploratory partition entry points that return
nil in the C source retain that behavior. The full `aig/gia` build now reports
untranslated cut code in `giaCut.go`; additional errors may remain.

GIA cut follow-up: ported priority-cut merging, containment and dominance,
truth-table expansion and minimization, incremental storage, and cut-manager
statistics to Go vectors. Focused race tests exercise cut ordering, storage
ownership, truth evaluation at cut sizes 4, 6 and 8, and removal of irrelevant
truth variables. The complete `aig/gia` build now reports remaining C-derived
syntax in `giaDeep.go` and `giaDup.go`; the focused suite is not a whole-package
build.

GIA duplication/synthesis follow-up: ported DFS ordering in both directions,
AIGER ordering, equivalence and literal remapping, output selection, and flop
abstraction. Ported the randomized synthesis driver against the Go frame and
command interfaces; it now restores batch mode on command failure and applies
the requested node-count goal (the C local minimum variable was never updated).
Focused race tests cover interface ordering, real XOR/MUX nodes, abstraction,
command execution, and failure handling. `aig/gia` still does not build: much
of `giaDup.go` remains untranslated, and later errors have not been ruled out.

GIA duplication follow-up: ported PI/flop permutations, flop gaps, zero and
onset/offset duplicates, copied mapping/timing attributes, and shared or
separate graph composition. Go rejects removing an input that still drives
copied logic; the PI-removal routine preserves register inputs instead of
removing the last CIs. Focused race tests exercise permutation equivalence,
register-transition ordering, independent metadata, and graph composition.
The whole `aig/gia` build still fails on untranslated portions of `giaDup.go`.

GIA duplication follow-up: ported choice-link DFS copying, invariant miters,
additional sequential cones, register self-loops, flop-class extraction,
parallel copies, randomized CI-edge mutation, and PI cofactor/quantification
operations. Focused race tests check preserved links and exhaustive small-graph
semantics including sequential interfaces. The choice-copy link uses the mapped
object ID directly; the C expression shifts the mapped literal twice. The full
`aig/gia` build still fails in untranslated parts of `giaDup.go`.

GIA DFS/cone follow-up: ported old-representative-aware DFS duplication,
selected-output and literal copies, cone extraction and reinsertion, internal
cofactoring, and block reduction. Old representatives now use a Go slice.
Focused race tests verify representative reuse, sequential output selection,
cone scratch cleanup and reinsertion, and small exhaustive cofactor behavior.
The full `aig/gia` build still fails in remaining `giaDup.go` code.

GIA miter follow-up: ported combinational and sequential miters, implication
and dual-output variants, paired-output conversions, AND/OR output reduction,
composition of a graph on top of another, property replacement, CI mapping,
and dual-side ordering. Focused race tests cover exhaustive small combinational
and sequential cases. The full `aig/gia` build remains blocked by untranslated
`giaDup.go` routines, including output trimming that relies on the unfinished
sequential structural sweep.

GIA sweep/trimming follow-up: ported ternary reachable-state analysis,
constant/equivalent-register substitution, iterative sequential structural
sweeping, and output trimming including dual-output pairs. Focused race tests
cover reset-trace preservation, unknown and toggling states, removed registers,
constant outputs, and dual pairs. The ternary simulation uses explicit Go
state slices rather than the source's packed frontier representation; the
remaining lower-level `giaTsim.go` routines still need porting. The whole
`aig/gia` build continues to fail in untranslated `giaDup.go` routines.

GIA duplication follow-up: ported top-level AND decomposition, class-aware DFS,
zero/unknown register initialization and its two-copy miter, constraint-output
reordering, selected sequential cones, limited-depth cones, and one-hot violation
generation. Focused race tests check multi-frame reset behavior, cone support,
constraint order, and exhaustive small one-hot assignments. The full `aig/gia`
build remains blocked by remaining `giaDup.go` code.

GIA duplication/support follow-up: ported one-hot constraint attachment with
register inputs kept last, levelized and explicit-vector copying, support-size
slicing and collection, constraint extraction, and structural output grouping.
Focused race tests cover sequential interface order, small-graph semantics,
support lists, and output classes. The full `aig/gia` build still fails in
untranslated demiter and related routines in `giaDup.go`.

GIA demiter follow-up: ported decomposition, output splitting, cone marking,
sequential half-graphs, and cross-part equivalence merging. The correspondence
division driver moved to `proof/cec` to avoid a GIA/CEC import cycle; callers
in still-untranslated packages must use the CEC entry point when ported. Its
focused correspondence integration test and the GIA focused race suite pass.
`giaDup.go` no longer contains untranslated routines. Edge assignment storage
and packing conversion are also ported, but the full `aig/gia` build now fails
in the remaining `giaEdge.go` delay and edge-optimization routines.

GIA edge-delay follow-up: ported mapped-LUT delay evaluation, forward and
reverse edge assignment, and the timing-box ordering used by delay evaluation.
Focused race tests cover edge assignment and timing-box arrivals. The full
`aig/gia` package still fails to build: `giaEdge.go` contains untranslated
window evaluation and edge optimization code. Focused tests do not establish
that the whole package builds.

GIA edge-optimization follow-up: ported mapping swaps, critical-path analysis,
and iterative edge assignment. Focused race tests exercise critical-level
collection, mapping restoration, and optimization of a mapped two-LUT chain.
The full `aig/gia` build now reaches untranslated `giaEmbed.go`; this does not
establish that later files or the package as a whole build.

GIA embedding follow-up: ported the matrix allocation and vector operations to
Go slices, with a focused numerical test; the existing focused GIA race suite
passes. `giaEmbed.go` still requires its packed, pointer-relative graph and the
remaining placement routines to be ported, and the full `aig/gia` package does
not yet build.

GIA embedding graph follow-up: replaced the packed relative-pointer object
representation with explicit Go adjacency lists, and ported the simple graph
constructor, accessors, and traversal marks. Focused race tests cover both
combinational links and register feedback. The full `aig/gia` build remains
blocked by untranslated supernode construction and subsequent embedding
algorithms in `giaEmbed.go`; the optimized constructor is not yet ported.

GIA embedding supernode follow-up: ported special reference counting,
supernode collection and optimized construction, along with breadth-first
distances and output-connected component discovery. Focused race tests cover
collapsed cones and distances; the full GIA package still does not build
because later routines in `giaEmbed.go` remain untranslated.

GIA embedding solver follow-up: ported distance dimensions, covariance,
eigenvector iteration, solution projection, placement derivation, wire-length
evaluation, and placement refinement. Focused race tests cover covariance,
projection, and placement. The whole GIA package still fails to build in the
remaining untranslated diagnostic and output code in `giaEmbed.go`; focused
tests do not establish whole-package correctness.

GIA embedding output follow-up: ported graph diagnostics, distance reporting,
Gnuplot output, and the top-level placement driver. Focused GIA race tests pass,
including a plot-data test. The full GIA build now reports untranslated C in
`giaEnable.go`; it has not passed a whole-package build or test.

GIA enable follow-up: ported sequential signal detection, initialized frame
unrolling, signal transfer and the alternate enable removal routine. The
alternate removal keeps the original register driver if its enable cannot be
recognized. Focused race tests cover shared reset detection and three-frame
initialization. The full GIA package still fails in the other enable-removal
routine and untranslated `giaEquiv.go`.

GIA enable/equivalence follow-up: ported the remaining enable-removal routine,
including register classes, and original-ID equivalence mapping and reduction.
Focused race tests check removal semantics and transitive equivalence pairs.
The full GIA build now fails in the many remaining untranslated routines in
`giaEquiv.go`; the package has not passed a whole-package build or test.

GIA equivalence inspection follow-up: ported topology validation, class and
literal counts, class reporting, and minimum-level choice selection. Focused
race tests cover class counts, topology rejection, and choice selection. The
full GIA build remains blocked by substantial untranslated `giaEquiv.go` code.

GIA equivalence reduction follow-up: ported representative-based and
minimum-level reduction, sibling-to-class conversion, phase handling, and
dual-output coloring. Focused race tests exercise both reductions over all
inputs of a small graph. The full GIA package still fails in remaining
`giaEquiv.go` routines; package-wide builds and tests remain outstanding.

GIA equivalence remapping follow-up: ported class remapping through cleanup and
DFS duplication, output-pair fixing, and speculative reduction with trace and
comparison outputs. Focused race tests verify class preservation and an
exhaustive false-equivalence example. The whole GIA build still fails in other
untranslated `giaEquiv.go` routines.

GIA initialized speculation follow-up: ported counterexample-initialized
speculative unrolling and frame selection, plus equivalence transfer after
object reordering. Focused race tests cover nonzero initial state, frame-copy
cleanup, and equivalence preservation. The full GIA build still fails in other
untranslated `giaEquiv.go` code.

GIA choice conversion follow-up: ported equivalence-to-choice conversion,
fanin-cycle checks, bad-choice removal, and choice counts. Focused race tests
exercise a structurally distinct equivalent choice across its truth table.
The full GIA build still fails in other untranslated `giaEquiv.go` routines.

GIA equivalence filtering follow-up: ported class transformation, proof-result
marking, and filtering by disproved speculative outputs. Focused race tests
cover removing disproved and shared-color candidates. The full GIA package
continues to fail in remaining untranslated `giaEquiv.go` routines.

GIA equivalence follow-up: ported latch-based filtering, representative
improvement, and combinational speculative comparison. Focused race tests
exercise latch-only classes and comparison truth tables. The full GIA build
still fails in other untranslated `giaEquiv.go` routines.

GIA part-filtering follow-up: ported the two AIGER-based equivalence filters
and preserved input-object mappings through miter cleanup and normalization.
A focused race test writes and reloads two sequential AIGER graphs and checks
cross-part class retention. The full GIA build still fails in remaining
`giaEquiv.go` and `giaEra.go` code.

GIA independent-choice follow-up: ported independent equivalence-class
marking and transfer verification. Focused race tests distinguish independent
alternatives from ones depending on the class head. The source solver-choice
helper itself is unimplemented in C and needs a considered Go implementation;
other `giaEquiv.go` routines and `giaEra.go` still block the whole GIA build.

GIA/CEC original-ID sweeping follow-up: moved the CEC-dependent sweeping
entry point into `proof/cec` to avoid a GIA import cycle; proved pairs are
collected from SAT-sweep representatives before original-ID reduction. The
focused CEC correspondence race suite passes, including a new integration test.
The full GIA build still fails in remaining `giaEquiv.go` and `giaEra.go` code;
callers of this entry point must use the CEC package when they are ported.

GIA explicit reachability follow-up: ported the state manager, hash table,
parallel input simulation, transition exploration, trace collection, and STG
writing. Focused race tests cover a reachable failure, full traversal, and the
state limit. The full GIA build now reports remaining `giaEquiv.go` routines
and untranslated `giaEra2.go`; package-wide builds and tests are outstanding.

GIA cube reachability follow-up: ported packed state-cube relations, minterm
expansion, packed pointer conversion, and CI fanout-cone collection. Minterm
expansion iterates only over collected dash variables, correcting the C loop's
out-of-range indexing. Focused race tests cover these operations. The remaining
cube manager and exploration code in `giaEra2.go` and `giaEquiv.go` still block
the full GIA build.

GIA cube manager follow-up: replaced page arithmetic with Go-managed state and
node slices retaining packed numeric handles, and ported allocation, recycling,
list containment, compression, and rebalancing. Focused race tests cover initial
states, predecessor links, minterm counts, and list containment. The full GIA
build remains blocked by untranslated tree exploration in `giaEra2.go` and
remaining `giaEquiv.go` routines.

GIA cube-tree follow-up: ported tree/list insertion, overlap collection,
containment checking, rebalancing, and output-support helpers. Focused race
tests include all 64 six-register assignments and duplicate rejection. The
full GIA build remains blocked by untranslated reachability exploration in
`giaEra2.go` and remaining `giaEquiv.go` routines.

GIA cube reachability driver follow-up: ported symbolic successor splitting,
SAT-based transition reconstruction, and the exploration driver. Focused race
tests cover correlated inputs, a verified sequential counterexample, complete
traversal, and the state limit. The full GIA build now reports remaining
`giaEquiv.go` and untranslated `giaEsop.go`; package-wide verification remains
outstanding.

GIA SAT-choice follow-up: translated the remaining Cec4 choice reconstruction
wrapper and verified the no-class duplication path in the focused race suite.
The upstream `Cec4_ManSatSolverAnd_rec` is itself an unconditional-zero stub;
the port preserves that incomplete behavior, so the independent-choice path
is not considered functional. `Gia_CommandSpecI` remains untranslated and
the full GIA package remains unbuildable.

CEC speculative refinement follow-up: moved `Gia_CommandSpecI` orchestration
to `proof/cec`, with an injected BMC callback to avoid a GIA/CEC/BMC import
cycle. Focused CEC correspondence race tests cover missing/no candidates and
invocation of BMC for a nontrivial candidate. The real BMC callback and the
`base/abci` caller still need wiring when those packages are ported; witness
resimulation through this driver is not yet tested. Full GIA build no longer
reports syntax errors in `giaEquiv.go`; it now stops in `giaIf.go`.

GIA ESOP follow-up: ported cover hashing, minimization, multiplication,
complementation, extraction, and conversion back to GIA. Focused race tests
compare exhaustive inputs across 50 deterministic small graphs. The full GIA
build now reports remaining `giaEquiv.go` and untranslated `giaExist.go`;
whole-package builds and tests remain outstanding.

GIA quantification follow-up: ported support tracking, cone collection and
copying, and both existential-quantification paths. Focused race tests cover
support expansion past 64 inputs, a quantified multiplexer, and constant/input
edge cases. The full GIA build now reports remaining `giaEquiv.go` and
untranslated `giaFalse.go`; whole-package verification remains outstanding.

GIA false-path follow-up: ported critical-path collection, SAT-based path
checking, graph rebuilding, and iterative/batch drivers. Focused race tests
cover an unsensitizable path, a sensitizable path, and output preservation.
The full GIA build now reports remaining `giaEquiv.go` and untranslated
`giaFanout.go`; whole-package verification remains outstanding.

GIA fanout/FORCE follow-up: ported static mapped-LUT fanout construction and
started the FORCE placement port with Go-owned graph nodes and simple/clustered
constructors. Focused race tests cover mapping fanouts and both constructors.
The full GIA build still fails in FORCE placement routines and remaining
`giaEquiv.go` code.

GIA FORCE placement follow-up: ported cross-cut analysis, output ordering,
DFS placement, refinement, and the associated experiment routines. Corrected
the simple graph constructor to mark primary inputs as inputs. The focused
port test suite passes with the race detector, including placement permutation
and cross-cut checks. The full GIA package still does not build: the latest
build reports syntax errors in `giaEquiv.go` and `giaFront.go`, and other
untranslated files may be exposed after those errors are resolved.

GIA frontier follow-up: translated frontier-slot allocation, cross-cut
measurement, graph conversion, and conversion back to ordinary fanin
differences. Focused race tests verify round trips with complemented edges,
multiple outputs, a constant output, and a dangling input. The full GIA
build still fails in `giaEquiv.go` and `giaFx.go`; the focused tests do not
establish that the package builds.

GIA extraction follow-up: translated cube dependency ordering, including the
unchanged-network case and first-cube/count lookup. The focused race suite
passes with a dependency-order regression test. The extraction driver and
other routines in `giaFx.go` are still C drafts; the full GIA build remains
blocked there and in `giaEquiv.go`.

GIA extraction SOP follow-up: translated direct PLA-to-AIG conversion and
factored decomposition-graph conversion using the Go `kit` and `dec`
packages. Focused race tests exhaustively evaluate small positive and
complemented covers. The full GIA build still fails in the remaining
`giaFx.go` and `giaEquiv.go` drafts.

GIA extraction retrieval follow-up: translated mapped-LUT truth collection,
node numbering, and conversion of truth tables to extraction cubes and
polarity flags. A focused race test reconstructs the extracted network for
both LUT fanin orders and all input assignments. The full GIA build still
fails in the remaining extraction insertion/driver and equivalence drafts.

GIA extraction insertion follow-up: rebuilt graphs from extracted cube
covers and added a focused exhaustive-input race test. This currently factors
all covers through `dec`; the original uses a separate DSD construction for
up to six inputs, so structural/optimization parity remains outstanding.
Removed duplicate translated retrieval drafts. The full GIA build now reports
the extraction driver, `giaEquiv.go`, and `giaGen.go` as syntax blockers.

GIA extraction DSD follow-up: insertion now uses the existing Go DSD
decomposer for covers with at most six inputs and imports its graph into the
GIA structural hash; larger covers still use SOP factoring as in the source.
The focused race suite checks all 256 three-input truth tables, and the full
`opt/dau` tests pass. Full GIA compilation remains blocked by the extraction
driver, `giaEquiv.go`, and `giaGen.go`; exact structural parity is unverified.

Classical fast-extract follow-up: started a standalone `opt/fx` Go package
with cube-free divisor construction and pivot selection from `abcFx.c`.
`go test -race ./opt/fx` passes. This does not yet implement the divisor
catalog, selection, or updates, and the GIA extraction driver is still a
draft; the new package is not a replacement for the complete algorithm.

Classical fast-extract divisor follow-up: ported four-literal mux/XOR
normalization and the divisor cube-literal update helpers. Race tests check
normalization against exhaustive input assignments and cover cube updates.
The catalog, priority selection, and full extraction update loop remain to be
ported before GIA can use this package.

Classical fast-extract catalog follow-up: implemented initial single- and
two-cube candidate collection, source-style weights and level tie-breakers,
literal-count/canonical filters, and containment identification. The full
`opt/fx` race tests pass. Candidate selection and iterative graph updates are
still outstanding, and GIA remains unbuildable.

Classical fast-extract pair matching follow-up: ported owner-grouped matching
of divisor cube pairs, including complemented mux divisors. Corrected cube
literal removal to preserve the owner entry when it equals a removed literal.
The complete `opt/fx` race tests pass; iterative extraction remains incomplete.

Classical fast-extract update follow-up: added a transactional divisor rewrite
for single-cube and paired-cube occurrences, including complementary-literal
simplification and new divisor covers. `opt/fx` race tests pass for these
cases. Overlapping occurrences currently return an error, and the iterative
priority/update driver is not connected; GIA is still unbuildable.

Classical fast-extract iteration follow-up: added an initial iterative driver
that selects positive-weight candidates, applies rewrites, and recomputes
containment. `opt/fx` race tests pass for shared-term extraction and limits.
It rebuilds the catalog rather than incrementally updating ABC's queue, so
selection/level parity and overlapping occurrences still require work before
connecting it to the GIA driver. The full GIA package remains unbuildable.

Classical fast-extract containment follow-up: remove duplicate and absorbed
cubes before cataloging, avoiding malformed one-sided divisors for covers
with containment. A deterministic test checks semantic preservation across
1,000 generated covers and all 16 input assignments; `opt/fx` race tests
pass. This preprocessing is currently quadratic, and selection parity and
overlapping occurrences still need validation before GIA integration.

GIA extraction integration follow-up: connected the Go cube extractor to
the GIA retrieval/rebuild driver; focused race tests exercise the complete
driver for both LUT orders. Containment comparisons are now limited to cubes
of the same owner. The extractor still rebuilds candidate priorities and can
reject overlapping occurrences, so ABC behavior/performance parity remains
outstanding. The full GIA build now reports `giaEquiv.go` and `giaGen.go`
syntax errors; it has not yet reached complete type checking.

GIA generator follow-up: translated packed random and supplied-input word
simulation, including complemented edges and output propagation. Focused
race tests pass with multiple words and complemented fanins. The full GIA
build remains blocked by `giaEquiv.go` and the file I/O routines in
`giaGen.go`; whole-package type checking remains outstanding.

GIA generator input follow-up: translated PLA-pattern reading, simulation
comparison, and JSON statistics output. A focused race test reads 64 patterns,
simulates an AND, and checks the resulting statistics. Full GIA compilation
still fails in `giaEquiv.go` and the remaining generator output routines.

GIA generator output follow-up: translated binary training/test dumps and PLA
training/validation/test dumps, retaining the source random-consumption order
and native-endian binary header. Focused race tests validate file headers and
all emitted AND input/output patterns. Full GIA compilation now reports
`giaEquiv.go` and `giaGiarf.go` syntax errors; it has not passed type checking.

GIA equivalence simulation follow-up: translated the Giarf simulation
workspace, class construction and refinement, hashing, and resimulation.
Focused race tests cover complemented input simulation and class splitting.
The full GIA package still fails on `giaEquiv.go` and later `giaGiarf.go`
routines; the rest of the Giarf SAT driver remains a C draft.

GIA Giarf level follow-up: translated reduced-level graph construction,
related class/member collection, counterexample pattern packing, and class
statistics. Focused race tests cover reduced graph roots and conflicting
packed patterns. Full GIA build still fails in `giaEquiv.go` and the Giarf
SAT driver, which remains untranslated.

GIA Giarf SAT follow-up: added a solver adapter for one- and two-literal
conjunctions using TAS or the CEC core, including model conversion to input
literals. Focused race tests exercise SAT, UNSAT, and constant queries with
both backends. The CEC adapter currently creates a solver per query, so it
does not yet preserve the source's incremental-solver performance. The Giarf
equivalence driver and the full GIA build remain unfinished.

GIA Giarf driver follow-up: translated levelized constant/equivalence SAT
checks, counterexample-driven class refinement, and the top-level driver.
Focused race tests prove a duplicate gate with both TAS and CEC backends.
The CEC path still creates a solver per query; broader SAT-driver parity and
performance validation are outstanding. Full GIA build now reports syntax
errors in `giaEquiv.go` and `giaGig.go`, before complete type checking.

GIA GIG reader follow-up: translated gate-inverter graph parsing and AIG
construction for PIs, POs, barriers, selectors, LUTs, boxes, and delays.
Focused race tests verify a LUT/delay graph across all inputs. No real GIG
fixtures were available, so broader format compatibility remains unverified.
The full GIA build now reports `giaEquiv.go` and `giaGlitch.go` syntax errors.

GIA glitch follow-up: replaced the packed C object allocation with stable
Go-owned nodes, fanin/fanout links, and truth-table evaluation. Focused race
tests cover graph construction and an AND truth table. Switching and glitch
simulation remain C drafts, and the full GIA package still fails to build in
`giaGlitch.go` and `giaEquiv.go`.

GIA glitch simulation follow-up: translated switching, event-frontier glitch
propagation, sequential packed simulation, and simulation drivers. Focused
race tests cover a reconvergent hazard with two glitches and zero net switches,
and packed sequential truth evaluation. The full GIA build now reports
`giaEquiv.go` and `giaHcd.go` syntax errors; full-package type checking and
broader sequential validation remain outstanding.

GIA HCD follow-up: translated shared-input choice miter construction,
transitive-fanin checking, and choice-link insertion. Focused race tests
exercise interleaved outputs, complemented outputs, TFI reachability, and
mark cleanup. HCD synthesis and equivalence-to-choice routines remain drafts;
the full GIA package remains unbuildable.

GIA HCD choice follow-up: translated equivalence-to-choice cone copying,
fanout-invalid choice removal, and snapshot reduction. Focused race tests
check both ordinary snapshot behavior and preservation of a structurally
distinct equivalent alternative. HCD synthesis and the broader GIA package
remain unfinished.

GIA mapping choice-level follow-up: translated level computation through
choice siblings and timing boxes, plus packing and LUT profile reporting.
Focused race tests cover sibling and box level propagation. The mapping
adapter and full GIA package remain unfinished and unbuildable.

GIA HCD synthesis follow-up: connected the DAR rewrite, refactor, balance,
snapshot synthesis, and choice conversion flows using the Go AIG adapter.
Focused race tests verify all snapshot truth tables and the complete choice
flow with synthesis enabled and disabled. The disabled path uses one snapshot
(the source initializes a three-element placeholder despite producing one);
the source's diagnostic `m3.aig` write is not reproduced. Full GIA build now
reports syntax errors in `giaEquiv.go` and `giaIf.go`; type checking remains
outstanding.

GIA mapping metrics follow-up: translated LUT counts, fanin counts, mapped
levels, mapped reference counts, and both overlap calculations. Focused race
tests cover a two-level mapped network. Box-aware LUT levels, the mapping
adapter, and the remainder of `giaIf.go` still require porting; full GIA
compilation remains blocked there.

GIA/IF boundary follow-up: moved the GIA-to-IF graph adapter and mapping
defaults into `map/if` to avoid a GIA-to-IF import cycle. An adapter test was
added, but it cannot yet run: even with a temporary focused GIA build, the
full IF package fails compilation in `ifCheck.go`, `ifCom.go`, and duplicate
`boolInt` definitions. The remaining GIA-side IF callers still need to move
or be rewired. Full GIA compilation also still stops in `giaIf.go`.

IF integration build follow-up: fixed duplicate helper names, C-style
boolean/complement expressions in cut checking, LUT library initialization,
sequential-mapping type errors, and several cell-tuning conversions. Also
fixed a shadowed decomposition step and a nil dereference on cell parse
failure. A temporary build using focused GIA sources now advances to further
`ifTune.go` errors; the bridge test and full IF/GIA package builds still do
not run, so these fixes have not yet received package-level verification.

GIA mapped-timing follow-up: translated box-aware LUT level propagation and
the combined LUT/edge/level metrics. A focused race test verifies a mapped
network containing a timing box. Full GIA compilation remains blocked by
other untranslated routines and preprocessor directives in `giaIf.go`.

GIA mapping transfer follow-up: translated LUT mapping verification and
mapping/packing transfer through copied object literals. Focused race tests
check a two-LUT network and packing record after duplication. More complex
renumbering and dangling-LUT cases remain to be tested; the full GIA package
still does not build because most of `giaIf.go` remains untranslated.

GIA mini-AIG reconstruction follow-up: translated compact AND-list expansion,
including constants, variable phases, intermediate references, and optional
structural hashing. Focused race tests exhaustively check a three-input
example in hashed and non-hashed modes. The IF adapter and full GIA build
remain unfinished.

## Main-thread IF integration checkpoint

Work is now sequential on the main thread; the historical parallel ownership
notes above describe earlier work. `map/if` compiles and its full package tests
pass against a temporary GIA core assembled from `aig/gia/test_ported_core.sh`.
This is a dependency-isolated check, not a successful build of the real package.
The ordinary `go build ./map/if` still fails because the real `aig/gia` package
contains untranslated C, first reported in `giaIf.go`. That file also contains
GIA/IF interoperation which must be moved or reorganized without introducing a
Go import cycle. The remaining GIA sources and dependent packages still require
porting and whole-package verification.

The IF work connects its truth-table, vector, SAT, DSD and GIA operations to
existing Go implementations and exposes GIA cell configuration through a
read-only accessor. The configuration-parser test now checks the zero-based
constraint encoding used by the C implementation. The isolated IF test uses
fixture and C-oracle paths from the checkout and does not remove unfinished
production files from the real package.

## GIA mapping progress

Moved GIA mapping-statistics routines to `mapping_stats.go` and the IF choice
reachability checker to `map/if/choice_check.go`; the latter belongs to the IF
package to avoid a GIA-to-IF import cycle. Focused GIA core tests pass, including
new mapping-statistics and structural-MUX coverage, and all IF package tests
pass against the isolated GIA core. The ordinary `go build ./map/if` remains
red: `giaIf.go` still contains untranslated GIA-to-IF reconstruction code.
The existing `If_CutDsdBalanceEval` only computes timing and does not populate
its requested AIG vector; complete that implementation before relying on the
DSD-balanced mapping reconstruction path.

## DSD-balanced reconstruction progress

`If_CutDsdBalanceEval` now constructs the compact AIG as well as estimating
its delay. Associative, mux and prime DSD nodes retain complemented edges and
pin permutations; nested prime SOP synthesis uses the enclosing cut's support
size when numbering AIG nodes. The mini-AIG-to-GIA bridge now lives with IF.
Tests check truth tables for those node types, manager-level cut polarity, and
reconstruction into a GIA. The complete IF tests pass with the isolated GIA
core. This supersedes the earlier note that DSD AIG generation was absent;
the full `aig/gia` and `map/if` packages remain unbuildable until the rest of
`giaIf.go` and the other untranslated GIA files are ported.

## IF-to-GIA reconstruction progress

Moved the truth-table LUT constructor and balanced IF-to-GIA graph rebuild into
`map/if`, preserving the one-way IF-to-GIA dependency. Tests cover mapping
records, constant and leaf shortcuts, complemented leaves, a full DSD-balanced
IF-to-GIA truth-table round trip, and recorded-library callback dispatch.
The recorded-library implementation in `base/abci` is still untranslated and
must be connected through `SetRecordedGiaBuilder` as that package is ported.
The complete IF tests pass against the isolated GIA core; this does not mean the
real GIA or IF package builds, and remaining `giaIf.go` callers must be moved
and connected to these IF-owned functions.

## IF cut and paired-LUT reconstruction

Ported the 4-4 paired-LUT reconstruction, including its packing record, and
local IF-cut to GIA cone reconstruction with choice alternatives. Exhaustive
truth-table tests cover five-, six- and seven-input paired LUTs. Local-cone
tests cover hashed and unhashed reconstruction, failure, constant leaves and
preservation of cut scratch data. The full IF tests pass only against the
isolated GIA core; the real GIA package still contains untranslated mapping
code, including the remaining logic-reconstruction callers.

## Two-LUT selection progress

Ported `Gia_ManFromIfLogicFindLut` into IF and verified both selected and
DSD-searched partitions against the original truth function. Fixed the
factorization output to stretch each LUT truth table to 64 bits as the C code
does. The internal checker used by reconstruction now accepts multiword cut
truth tables through eleven variables, with exhaustive reconstruction tests.
The older pointer-based public `If_ManSatCheckXY` and `If_ManSatCheckXYall`
adapters remain limited to six variables and need slice-based integration as
their other callers are ported. Full IF tests pass against the isolated GIA
core only; the real package still fails on untranslated GIA mapping code.

## Wide decomposition and cofactor mapping

The public two-LUT decomposition APIs now take truth-word slices and search
wide cuts through eleven variables; tests validate every assignment of each
seven- through eleven-input example. The prior note about their six-input
pointer limit is superseded. Ported cofactor-based LUT construction and mux
mapping records, with truth-table tests for constant cofactors and two internal
cofactors. Full IF tests still pass only with the isolated GIA core; remaining
`giaIf.go` callers and other untranslated GIA files prevent the real build.

## LUT structures and mapped hashing

Ported AND/OR bi-decomposition mapping and the LUT-structure dispatcher,
including ordinary, 4-4, 7-5 and six-input-filter modes. Assignment-by-
assignment tests cover the resulting one-, two- and three-LUT graphs and
bi-decomposition output phases. Ported mapped-GIA structural hashing with a
test for merging equivalent LUT roots and preserving output polarity. The
focused GIA core and the IF package with an isolated GIA core pass; these are
not full-package builds. The real GIA-to-IF mapping driver and other source
files remain to be ported and connected.

## Programmable-cell configuration progress

Ported sequential DSD-to-cell tuning and packing of cell truth bits, input
permutations and output polarity. The DSD configuration accessor now returns
actual stored configuration words instead of unrelated truth-table memory;
configuration bit counts follow the C representation. Tests exercise bit
packing across 32- and 64-bit boundaries and match, pack, and reconstruct a
programmable cell into a GIA. The full IF suite passes with the isolated GIA
core. The actual mapping driver, its remaining cell path, and other GIA files
still need porting before the real package can build.

## IF mapped-network driver progress

Ported `Gia_ManFromIfLogic` into IF and added a validated GIA metadata attach
boundary. Tests exercise structural and truth-driven mapped graphs, as well as
programmable-cell configuration generation and reconstruction. Fixed small-LUT
truth-cache initialization that left supports below six with zero word counts.
The complete IF suite passes against the isolated GIA core. The real packages
still cannot build: the GIA mapping entry points and numerous other files remain
untranslated, and the moved driver must be connected to those entry points.

## Mapping entry-point progress

Ported the GIA mapping entry points and SOP/DSD balancing wrappers to IF,
including boxed-graph unnormalization, timing transfer, and optional mapping
hashing. Tests run a small GIA through complete mapping and both balancing
paths and exhaustively check its outputs against the source function. This
also exposed and fixed initialization and shared indexing of the ISOP
complement cache. Full IF tests pass against the isolated GIA core and the
focused GIA core passes. The source's global DSD manager, power switching,
recorded-library and verbose timing trace now have explicit injection points;
their remaining providers must be connected as the owning packages are ported.
The real GIA and IF packages still do not build because other GIA files contain
untranslated C.

## Hierarchical packing and frame timing

Ported GIA hierarchical LUT clustering into IF to avoid the GIA/IF import
cycle, including timing, pair/triple choices, and packing selection. Focused
pair and triple tests pass. Moved frame timing helpers into the GIA frame bridge
and verified that timing arrays are copied. The focused GIA core and the full
IF suite against an isolated GIA core pass. The real build now first reports
untranslated `giaIso.go`; `giaIf.go` still retains the original disabled cell
helper, which must be resolved during full integration.

## GIA isomorphism refinement progress

Ported the isomorphism manager, level-based initial classes, signature
simulation in both directions, class refinement, tie breaking and output class
collection. Tests cover packed signature words, class splits, forward/backward
updates and output classification. The focused GIA core and IF tests against
the isolated core pass. Canonical ordering, graph reduction and other routines
in `giaIso.go` are still untranslated, so this is not a full package build.

## Canonical isomorphism and output reduction

Ported canonical PI, internal-node and output ordering, canonical AIGER
signatures, output equivalence refinement and dual-output reduction. Focused
tests compare renamed cones, check PI permutations and constant networks, and
verify regular and dual-output reductions. The complete `giaIso.go` routines
now live in native `iso_*` files. Focused GIA tests and IF tests against the
isolated GIA core pass. The real package remains unbuildable; the next reported
source file is `giaIso2.go`, with further untranslated GIA files beyond it.

## Second GIA isomorphism implementation

Ported `giaIso2.c` into native refinement and cone-comparison files, including
class splitting and both output-reduction modes. Focused tests cover signature
propagation, tied versus unique objects, cone comparisons, class refinement,
and truth preservation of reduced representatives. The focused GIA core and
IF tests against its isolated copy pass. The real build now first reports
untranslated `giaIso3.go` and `giaJf.go`; the full GIA package remains red.

## Third isomorphism pass and LUT cut representation

Ported `giaIso3.c`, including signature propagation and reverse-level
signatures. Added focused tests for propagation and repeatability. Began
porting `giaJf.c`: packed LUT cut headers, signatures, containment, and merge
operations now live in `jf_cut.go`, with tests for header fields, literal
polarity, ordered merges, and LUT limits. The mapper itself remains largely
untranslated. Focused GIA core tests pass; IF tests also pass against an
isolated, selected GIA core, which is not a full package build. The real
`go build -p 1 ./map/if` still fails on untranslated code in `giaJf.go` and
additional GIA drafts remain. The remaining JF call sites must be adapted to
Go slices and connected as the mapper is ported.

Ported JF truth-table-to-CNF clause generation using the Go KIT ISOP
implementation. Tests exhaust all three-input truth tables and exercise a
six-input table, complemented literals, and appending to existing clauses. The
focused GIA core passes; the actual package build remains blocked by the rest
of `giaJf.go` and other untranslated GIA files.

Ported JF CNF creation and reverse-order variable compaction, including the
optional primary-output clause and object-to-clause metadata. The shared CNF
package now has a constructor that initializes both its public representation
and legacy internal fields. Tests cover direct and remapped CNFs and vector
ownership; the focused GIA core and `go test ./sat/cnf` pass. The CNF's old C
manager pointer cannot represent a GIA manager in the current Go AIG type and
is left unset; integration must use the GIA owner separately where needed. The
real GIA and IF builds still fail on remaining untranslated drafts.

Ported JF reference-count initialization and coarsening marks. Tests exercise
both structural MUX and XOR reference discounts with coarsening enabled and
disabled; the focused GIA core passes. `Jf_ManInitRefs` now delegates to this
implementation, while the surrounding JF manager and mapper remain draft C.

Ported JF cut comparisons, containment filtering, size sorting, and ordered
store insertion with dominance pruning. Focused tests exercise delay versus area
priority, the flow tolerance, and pruning in both ordered and variable-based
modes; the focused GIA core passes. JF manager callers still require adaptation
to Go slices, and the actual GIA package remains unbuildable.

Ported recursive JF cut area/reference accounting and MFFC limit checks, with
coverage for shared cones and restoring reference counts after both normal and
early exits. The draft mapper entry points delegate to the new implementation;
they still depend on changing packed-cut accessors to Go slices. The focused
GIA core passes, but this does not establish a full GIA or IF build.

Replaced JF's pointer-cast packed-cut arena with mutable Go-owned cut lists and
converted the best-cut accessor to slices. Selecting a best cut now reorders
the underlying list in place; focused tests cover stable storage across
allocations and cut reordering. Existing mapper bodies still contain C pointer
and loop syntax and need adaptation to the new slices. Focused GIA tests pass;
the actual `map/if` build continues to fail in `giaJf.go`.

Ported the JF manager state, allocation and cleanup, cut accessors, and DSD
class reporting. Lifecycle tests cover ordinary, truth-table and DSD/CNF modes;
the focused GIA core passes. The real package remains blocked on untranslated
cut enumeration and mapping-round code in `giaJf.go` and other drafts.

Moved JF cut assignment, arrival/flow and direct-reference helpers, store
filtering, and recursive area entry points into compiled Go using the new
manager and mutable cut storage. A manager-level test checks cut assignment,
arrival, flow, reference updates, area and MFFC restoration. The focused GIA
core passes; complete GIA and IF builds remain blocked on untranslated mapper
and other draft files.

Ported JF cut enumeration, truth expansion and support minimization, including
structural, truth-table and DSD modes. Focused tests cover each mode,
complemented fanins and constant support reduction; the focused GIA core
passes. Mapping rounds and other GIA drafts remain untranslated, so this is
not yet a full package build.

Ported JF cut-computation and mapping rounds, including mapped-reference and
delay recomputation, flow selection and exact-area refinement. A two-level
network test runs the rounds in structural, truth-table and DSD modes; the
focused GIA core passes. The original mapper explicitly rejects buffers and
the Go path retains that behavior. Result derivation and other GIA draft files
remain unfinished; this is not a full package build.

Ported structural LUT mapping attachment and minimized-cut mapped-GIA
reconstruction, including DSD, truth-table and optional CNF paths. Focused
tests verify mapped truth tables and GIA mapping structure; CNF validation
checks literal ranges. The full focused GIA core passes. Pure-AIG derivation,
mapper entry points, and other GIA draft files remain to be ported.

Ported the JF driver, pure-AIG reconstruction, CNF entry points and remaining
cut-store diagnostics. End-to-end focused tests exercise structural mapping,
truth-table and DSD mapping, all three pure-AIG modes, and CNF generation;
the focused GIA core passes. `giaJf.go` now parses, although full-package
type checking remains blocked by later C drafts. The actual `map/if` build now
first reports untranslated preprocessor syntax in `giaKf.go`.

Began porting `giaKf.c`, the threaded cut mapper. Its cut structure, packed
storage encoding, load/save operations, ordering and basic merge/containment
functions now have Go implementations. Focused tests cover storage round trips,
unit cuts and ordered/unordered merge limits; the focused GIA core passes.
The KF manager, scheduling and mapping logic remain C drafts, and the real
GIA/IF package build still fails on `giaKf.go`.

Ported KF cut-set state, selection, hash helpers, and ordered/unordered merge
and dominance-filter paths. Focused tests cover hash insertion and cleanup,
selection, merging, filtering and cut statistics; the focused GIA core passes.
The hash-optimized merge path, manager and worker scheduling are still drafts,
so the real GIA and IF packages remain unbuildable.

Ported KF's hash-based cut merging and filtering, manager allocation and
cleanup, cut accessors, and mapped-reference/delay calculations. Focused
tests cover hash merging and manager reference lifecycle; the focused GIA core
passes. Worker scheduling, mapping execution and result derivation remain to be
ported, and the actual GIA/IF builds remain blocked by `giaKf.go`.

Ported the KF mapping driver and dependency-aware Go worker scheduling, then
removed the translated `giaKf.go` C draft. End-to-end tests cover ordered,
simple and hash merging plus one and three workers; the worker test passes
under the race detector, and the focused GIA core passes. The actual
`go build -p 1 ./map/if` now first reports untranslated `giaLf.go`; full GIA
package type checking and integration remain unfinished.

Began porting the LF mapper: replaced its variable-length C cut structure with
bounded Go leaf storage and translated merge, dominance, ordering and
best-cut metadata operations. Focused tests cover regular and MUX merges,
cut ordering, best-cut state and dominance pruning; the focused GIA core
passes. LF's manager, cut memory, truth functions and mapping driver remain
untranslated, so full GIA and IF builds are still blocked.

Ported LF's paged, delta-encoded cut memory and recycling, manager state and
basic accessors, and truth-table combination/minimization for AND, XOR and
MUX cuts. Tests cover page reuse, complemented MUX inputs, support reduction,
and truth tables above and below six variables. The focused GIA core passes;
LF enumeration and integration still contain untranslated C, so the actual
GIA and IF packages remain unbuildable.

Ported LF arrival access, best-cut loading, MUX-cut construction, output-driver
analysis, cut costs and reusable cut-set lifecycle. Focused tests cover best
cut restoration, arrival/cost calculations, output polarity counts, cut-set
reference release and reuse; the focused GIA core passes. LF enumeration,
reference propagation and driver logic remain C drafts, so full GIA/IF builds
are still incomplete.

Ported LF candidate timing and flow evaluation, recursive exact-area costing,
initial flow references and cut-set fanout references. Focused tests verify
arrival/required-time behavior, reference restoration and initial fanout counts;
the focused GIA core passes. LF enumeration, map-reference propagation and
result derivation remain unfinished, so actual GIA/IF builds still fail.

Ported LF mapped-reference selection, required-time propagation, reference
blending and recounting, including timing-manager callbacks and optional MUX
selection. A focused ordinary-network test verifies references, required times,
area, edges and delay; the focused GIA core passes. Timing-manager and MUX7
integration still need broader tests. LF enumeration and the driver remain C
drafts, so the actual GIA and IF builds remain incomplete.

Ported LF's main cut-enumeration loop, including previous-round cut reuse,
choice and explicit MUX handling, grouped cuts, timing, truth minimization and
best-cut storage. Focused tests cover two consecutive mapping rounds with and
without truth minimization and verify an explicit MUX truth table. The focused
GIA core passes; manager lifecycle, result derivation and driver code remain
to be ported before a real GIA/IF build can succeed.

Ported LF manager allocation/cleanup, defaults, and iterative mapping rounds,
including timing-manager arrival callbacks. A two-level network passes two
flow rounds and an exact-area round with and without truth minimization; the
focused GIA core passes. Power-aware LF mapping currently requires a
`LfSwitchingProvider` connected to the still-unported GIA switching simulator.
LF result derivation and public driver remain draft; actual GIA/IF builds fail.

Ported LF structural mapping attachment, coarse-network expansion and
truth-minimized mapped-GIA reconstruction, including cofactor-based MUX
splitting. Focused tests check truth preservation and mapping verification for
structural, minimized and coarse modes, including an explicit MUX. The focused
GIA core passes. The cofactor-splitting case still needs a dedicated test;
public LF driver and power simulation remain incomplete, and the actual
GIA/IF builds are not yet green.

Ported LF's public mapping drivers, timing transfer and normalization paths,
and crosscut diagnostic; removed the translated `giaLf.go` draft. Focused
end-to-end tests cover structural, coarsened and truth-minimized driver modes,
and a dedicated five-input cofactor-splitting test verifies all assignments.
The focused GIA core passes. Power mode still needs the switching simulator
provider connected; timing-box and MUX7 driver cases need broader integration
tests. The actual `go build -p 1 ./map/if` now first reports C syntax in
`giaMan.go`; full-package type checking has not yet run.

Started porting the remaining `giaMan.c` manager utilities. Verilog name
handling, usage analysis, name formatting and structural Verilog export now
use Go I/O; manager memory reporting also has a native implementation. A
focused test checks emitted names and complemented-edge wiring, and the focused
GIA core passes. Statistics, DFS timing and other `giaMan.go` functions remain
C drafts, so the actual GIA and IF builds are still blocked.

Ported manager DFS collection, arrival and required propagation, slack
calculation and reporting, including timing-box traversal. Focused tests cover
raw and LUT-mapped networks plus a timing box; the focused GIA core passes.
Other `giaMan.go` reporting functions and later draft files still prevent the
actual GIA/IF package build.

Ported the manager's NPN class report using the Go DAR canonicalization library
and its per-output miter report using the switching-simulation provider. Focused
tests exercise both reports; the focused GIA core passes. The miter probability
report shares LF's still-unconnected switching provider. The main manager
statistics routine and later C drafts still block actual GIA/IF builds.
