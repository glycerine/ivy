# Simulation and dataset utility review

Scope: all 13 active routines in `src/aig/gia/giaGen.c`. Sequential review;
no agents or branches. The wider GIA audit remains open; original completion
is 2/131. The preceding fast-extraction family passed frozen full-module race
and vet at fffb6aa97.

## Plan

1. [x] Inspect all implementations in gen_sim.go, gen_dump.go and gen_read.go.
2. [x] Add new red regressions for actual XOR and MUX word evaluation before
   repairing their shared direct gate helper. Select the actual operation once
   per node; retain the original AND/buffer packed algorithm and edge phases.
3. [x] Check simulation against actual C and independent scalar evaluation,
   including physical random CI order versus logical explicit input order,
   all node values for ordinary graphs, typed gates via C AND expansion,
   interleaved/reordered ports, constants, registers, word widths and RNG state.
4. [x] Compare every binary/PLA dump byte and report against C for positive
   partitions and multiple seeds; independently decode rows and headers.
   Preserve source's unusual binary 64x word consumption and retained state.
5. [x] Review parsing, parameter detection, label comparison and JSON stats
   against C. Test malformed files and safe existing Go extensions separately.
6. [x] Measure and simplify redundant buffering/whole-dataset copies while
   retaining file content, RNG behavior and valid-input contracts.
7. [x] Reconcile all 13 active routines; keep production/tests/fixtures frozen
   through uncached full-module race testing and GIA vet. Inspect terminal
   results and update progress records before closing this source family.

## Evidence so far

`TestGenWordUsesActualXor` (driver and direct gate helper) and
`TestGenWordUsesMuxControl` fail before repair:
`/tmp/abcgo-gen-gates-red.log`. Both incorrectly return 8888...; XOR requires
6666..., MUX requires caca.... The repaired regressions and existing ordinary
simulation test pass (`/tmp/abcgo-gen-gates-repaired.log`).

The actual-C matrix passes all 4,608 records, checking ordinary all-object
payloads, all logical CI/CO words and the next RNG word. Typed C graphs are
explicit ordinary AND expansions of the real Go XOR/MUX graphs; comparison
uses their common interface. All 786,432 packed input assignments are checked
independently at every Go object. Both drivers, 1/2/5 word widths, every chosen
edge/control/output phase, unused inputs, interleaved and reordered ports, and
RO/RI interfaces occur. Input data and Value scratch remain unchanged, and
input/result storage does not alias.
`/tmp/abcgo-gen-native-words-positive.log`, 2.386s, exact handle10292 exit0.

Original C asserts in Vec_WrdEntryP for zero-word random and explicit drivers.
Both failures are characterized in core-disabled child processes. Retain and
independently test the existing safe Go zero-width extension: empty views and
storage, no RNG consumption. The initial zero-width native matrix was an
invalid C precondition, not another implementation bug.

## Source inventory

| C routine | Go file | Status |
| --- | --- | --- |
| Gia_ManObjSim | gen_sim.go | Native all-object views and zero-width extension |
| Gia_ManObjSimPi | gen_sim.go | Native RNG sequence and scratch preservation |
| Gia_ManObjSimPo | gen_sim.go | Native phases/ports and scalar comparison |
| Gia_ManObjSimAnd | gen_sim.go | Red XOR/MUX repairs; native/scalar matrix |
| Gia_ManSimulateWords | gen_sim.go | Native packed/random matrix |
| Gia_ManSimulateWordsInit | gen_sim.go | Native explicit/logical-order/ownership matrix |
| Gia_ManDumpFiles | gen_dump.go | Exact native files/reports/RNG, scalar decoded rows |
| Gia_ManDumpPlaFiles | gen_dump.go | Native three-partition bytes and streamed output |
| Gia_ManSimLogStats | gen_read.go | Native semantic JSON, independent counts, escaped names |
| Gia_ManSimParamRead | gen_read.go | Native dimensions, LF/CRLF/directives, malformed/EOF tests |
| Gia_ManSimFileRead | gen_read.go | Native packing/labels, OR/append, validation before mutation |
| Gia_ManCompareValues | gen_read.go | Native first-logical-CO reports, typed scalar/JSON checks |
| Gia_ManTestOneFile | gen_read.go | Complete native drivers, dimension rejection, typed classification |

## Compatibility and compactness constraints

Simulation is one combinational timeframe; RO/RI ports are additional inputs
and outputs. Do not invent sequential stepping. Source random input generation
follows physical object order; explicit input vectors bind logical CI order.

Binary dump simulates 64 times more words than it exports, affecting both RNG
state between partitions and the retained vSims/NSimWords. Preserve this source
behavior; do not merely truncate unused simulation. Streaming file assembly
and parsing can remove intermediate copies without changing these contracts.

Current Go reader correctly accepts final rows without a newline and safely
escapes JSON names. Original C's parameter counter ignores an unterminated
last line and fprintf does not escape JSON names. Characterize these differences
rather than regress existing safe behavior. Comparison requires at least one
CI in C (division by CI count); Go explicitly rejects the undefined zero-CI
case. Empty Go filename selects the manager name, representing C's NULL option.

## Dataset and classification evidence

All 128 positive-partition dump configurations match every native file byte,
file-list entry, stdout report, retained word count and next RNG word. Both
formats cover multiple seeds/sizes/phases, explicit and manager-name fallback,
logical reordering, registers, and typed gates via C expansions. Independently
decode binary headers/row-major bits and PLA directives/rows against scalar
truth (`/tmp/abcgo-gen-native-dumps-first.log`, 2.225s). Two more native cases
cover no-CI constant outputs. Go's safe empty-dataset extension and negative
count/missing-parent failures are separate from C's undefined zero-width path.

All 1,536 direct/complete classifier configurations match C reports, packed
words, labels and semantic JSON fields. Independent expected label/error/
majority counts and graph statistics cover ordinary AND/XOR/MUX expansions,
phases, reordered logical ports, LF/CRLF, directive lines, multiple word counts
and optional stats (`/tmp/abcgo-gen-native-reader-first.log`, 5.789s).

Additional checks retain source OR/append behavior, support label-only files,
reject mismatched input counts without changing simulation, and preserve caller
outputs/buffers on invalid parameter/file dimensions. EOF without newline and
proper rejection of empty/nonbinary datasets are existing safe Go extensions;
C's differing parameter results are characterized separately. Typed classifier
checks exercise the repaired simulator through file reading and JSON stats,
including quoted/newline names. ABC's actual XOR/MUX depth cost is two AND
levels, as defined in gia.h; the initial test's unit-depth assumption was a
fixture expectation error. Repeated widths exercise grow/shrink/zero/grow and
invalid-operation guards; bounded object views cannot append into a neighbor.

## Compactness results

Parameter scans no longer store every byte row. File packing validates first,
then streams into existing caller storage using one reusable row buffer. This
retains the previous Go validation-before-mutation contract, at the cost of a
second sequential scan for packing. The materializing private helper remains
for callers explicitly requesting rows. PLA writes directly through one file
buffer; binary header and payload use one allocation. Preserve source random
consumption, partition order and retained simulation, including its binary64x
word behavior. File-close errors remain observable.

One 150ms local before/after benchmark run:

| Operation | Before time | After time | Before B/op | After B/op | Before allocs | After allocs |
| --- | --- | --- | --- | --- | --- | --- |
| Parameter 64 rows | 40.313us | 19.759us | 18,424 | 4,264 | 330 | 4 |
| Parameter 4,096 rows | 1,913.712us | 308.799us | 1,046,668 | 4,280 | 20,498 | 4 |
| Parameter 65,536 rows | 40,651.379us | 4,943.472us | 19,069,152 | 4,280 | 327,709 | 4 |
| PLA one word | 237.644us | 257.923us | 16,093 | 14,078 | 33 | 36 |
| PLA 64 words | 2,028.093us | 1,466.545us | 413,957 | 32,216 | 48 | 39 |

Logs: `/tmp/abcgo-gen-before-bench.log` and `...-after-bench.log`.
Tiny PLA output is slightly slower and allocates three more objects. Do not
claim a uniform speedup from short local observations. The native/scalar format
and reader matrices pass after storage simplification. The final combined focused suite and frozen full-module integration pass,
as recorded below.

## Checkpoints

- d73a89c7c: packed actual-gate repair and regressions.
- 7fd447184: native dataset/classifier matrices and baseline benchmarks.

The code checkpoints use red: until full-family integration is verified.
- 9636cb604: streamed dataset storage and compatibility/boundary coverage.

## Focused reconciliation and integration gate

All 19 final focused tests pass without skips in 22.745s
(`/tmp/abcgo-gen-final-complete-focused.log`, exact handle46932 exit0).
This also includes the existing compiled simulation boundary test. The full
13-routine source inventory is reconciled above.

Production/tests/fixtures are frozen at9636cb604 for the uncached full-module
race gate (exact handle16690, `/tmp/abcgo-gen-final-module-race.log`).
GIA vet handle53077 terminates exit0 with an empty log
(`/tmp/abcgo-gen-final-vet.log`). Exact full-module race handle16690 now
terminates exit0: 130 tested packages pass, three command packages have no
tests, and there are zero FAIL/DATA RACE records. GIA 601.497s, ABCI 29.709s,
BMC 24.450s. Code/tests/fixtures remained frozen throughout both gates.

Close the simulation/dataset source family. The wider GIA review and original
131-package goal remain open; completion stays2/131. Next is giaGig.c, with
read-only native evidence of shortened-truth Boolean loss, selection validation
and reporting issues. Preparation is in `/tmp/abcgo-gig-next-review-plan.md`.
