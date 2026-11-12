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
and vet pass. The base ABC network port has advanced through its original
`abcAig` manager APIs; the next full-package parser blocker is `abcBarBuf.go`.

Subsequent integration: `bool/rsb` is ported and race/vet clean, with original-C
search and counterexample oracles; `misc/vec` hash indexes now remain coherent
after explicit mutation and are race-tested. CUDD's ZDD set operations also
have a permanent checksum fixture derived from the original C implementation
(65,536 operation pairs). The expanded AIG/Scl source-file race suite passes
after a Repr fix; all four sweep drivers reduce three registers to one while
preserving sequential traces. The full AIG package remains blocked by other
original files, and full-module verification remains outstanding.

`bool/lucky` is now fully ported and committed after race/vet validation; its
fast/simple/swap implementations match the original C outputs, phase flags and
permutations on 550 differential cases.

## AIG free-to-nil audit

The AIG owner scanned every `abcgo/aig/aig/*.go` file and converted all four
remaining live null-safe free-and-clear expansions to nil assignments. The
`pMemory` example disappeared as part of the actual `Part_Man` pool port: the
Go manager now owns typed pooled blocks and `Part_ManStop` clears the manager,
so there is no C-style `pMemory` local to retain. `no_c_free_test.go` lexes every
Go file in the package and fails on live `free`/`ABC_FREE` calls while ignoring
comments and literals. `GOCACHE=/tmp/abcgo-cache go test
abcgo/aig/aig/no_c_free_test.go -count=1` passes.
