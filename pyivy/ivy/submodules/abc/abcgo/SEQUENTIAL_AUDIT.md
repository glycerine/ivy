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
