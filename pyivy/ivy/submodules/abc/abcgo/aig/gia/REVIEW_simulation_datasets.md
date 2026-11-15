# Simulation and dataset utility review (in progress)

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
4. [ ] Compare every binary/PLA dump byte and report against C for positive
   partitions and multiple seeds; independently decode rows and headers.
   Preserve source's unusual binary 64x word consumption and retained state.
5. [ ] Review parsing, parameter detection, label comparison and JSON stats
   against C. Test malformed files and safe existing Go extensions separately.
6. [ ] Measure and simplify redundant buffering/whole-dataset copies while
   retaining file content, RNG behavior and valid-input contracts.
7. [ ] Reconcile the complete source inventory, freeze code/tests/fixtures,
   run uncached full-module race and GIA vet, inspect terminal results and
   update all progress records before closing the family.

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
| Gia_ManDumpFiles | gen_dump.go | Pending native binary matrix |
| Gia_ManDumpPlaFiles | gen_dump.go | Pending native PLA matrix |
| Gia_ManSimLogStats | gen_read.go | Pending JSON/field checks |
| Gia_ManSimParamRead | gen_read.go | Pending parser/error matrix |
| Gia_ManSimFileRead | gen_read.go | Pending packed file/label matrix |
| Gia_ManCompareValues | gen_read.go | Pending report/statistics matrix |
| Gia_ManTestOneFile | gen_read.go | Pending complete driver matrix |

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
