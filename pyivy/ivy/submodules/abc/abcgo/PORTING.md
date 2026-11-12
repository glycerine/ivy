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
