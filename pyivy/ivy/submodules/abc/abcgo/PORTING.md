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
