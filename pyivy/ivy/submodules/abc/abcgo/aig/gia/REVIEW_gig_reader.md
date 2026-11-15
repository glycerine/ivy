# Gate-Inverter Graph reader review (complete)

Scope: all seven active routines in `src/aig/gia/giaGig.c`. Work is sequential,
without agents or branches. The wider GIA audit and 131-package goal remain
open; original completion is 2/131. The preceding dataset review passed full
module race/vet and closed at555ed782b.

## Plan and progress

1. [x] Inspect reference parsing, typed records, construction and complete
   reader behavior against C. Preserve Kit_TruthToGia, interface grouping and
   original cut semantics for Box/Del/Seq; do not invent state stepping.
2. [x] Add red regressions for abbreviated truth-table Boolean loss, invalid
   unused selection parents, and missing native success reports before fixes.
   Reuse util.Abc_TtReadHex, reject non-Bar/Del selection parents, and restore
   source count order plus the nine-wide elapsed-time field.
3. [x] Compare parsed records, category orders, complete constructed graph,
   Boolean truth and public reader reports against actual C. Cover all record
   types, constants/missing pins, dependent LUTs, and shortened/full hex forms.
4. [x] Preserve original C LUT6 assertion failure, and test six-input behavior
   only with an explicit narrowly checked correction of its k==4 typo.
5. [x] Expand validation, sparse-ID/whitespace compatibility, resources and
   metadata ownership checks. Distinguish original C defects/preconditions.
6. [x] Measure compactness and remove duplicated node descriptors/string
   retention while preserving sparse IDs, source order and synthesis.
7. [x] Reconcile all seven routines, freeze code/tests/fixtures, run uncached
   full-module race and GIA vet, inspect terminal results and update progress.

## Regressions and current evidence

All five shortened-truth tests fail before repair: F=>000F instead ofFFFF,
8=>0008 instead of8888, AB=>00AB instead ofABAB, A=>000A instead ofAAAA,
5=>0005 instead of5555. These are Boolean failures on valid four-input GIG
files. Four unused invalid selection cases (PI/PO/LUT/self parents) also fail
before repair by returning a graph instead of rejection.
`/tmp/abcgo-gig-short-select-red.log`, exact handle11713 exit1.

Read the truth using the existing source-shaped util.Abc_TtReadHex after bounded
hex validation. It stretches abbreviated tables as the C parser does; plain
integer parsing does not. Require selection parents to be Bar or Del even when
the selection is otherwise unused. The separate success-report regression
fails before repair because Go emits nothing where C prints counts/time
(`/tmp/abcgo-gig-report-red.log`, handle52436 exit1).

The repaired minimal suite passes1.850s, including the existing LUT/delay test
(`/tmp/abcgo-gig-minimal-repaired.log`). Native/parser/constructor coverage subsequently passed; frozen full-module
integration is the remaining gate.

## Native oracle design

The C fixture includes actual giaGig.c. Original mode preserves every source
assertion, including the known line311 LUT6 k==4 typo. Corrected mode intercepts
only that exact file/line/expression and independently requires six stored
fanins for every LUT6; all other assertions remain active. It validates the
number of corrections across private and public reader calls. A source-line
check makes a changed upstream assertion fail the fixture setup explicitly.

The fixture exposes canonical parsed records and category order vectors, then
prints exact object kinds/fanin literals, CI/CO vectors, identity, register
count and truth. The Go comparator uses these source outputs plus an
independent record-level truth oracle. Raw failure cases remain separate from
corrected valid-input comparisons. No production C source is changed.

## Source inventory

| Original C | Go implementation | Review status |
| --- | --- | --- |
| Gls_ManAlloc | gigParse/gigConstruct private records | Native state and independent lifetimes checked |
| Gls_ManStop | Go-owned records and file defer | Read errors and owned-state lifetimes checked |
| Gls_ManCount | gigParse and reader count report | Native category/count/report matrix passes |
| Gls_ManParseOne | gigRef | Constant, malformed and sparse reference tests pass |
| Gls_ManParse | gigParse | Native state, validation and red regressions pass |
| Gls_ManConstruct | gigConstruct | 1,162 native graph/truth cases pass |
| Gia_ManReadGig | Gia_ManReadGig | Public/private native matrix and report pass |

Private representation can be idiomatic Go because these Gls routines are not
publicly declared in gia.h. Their parsing/construction algorithms and interface
orders still require source reconciliation, not a substitute synthesis method.

## Compactness baseline

Before descriptor/index changes, short150ms runs show:

| Operation | 64 records | 4,096 records | 65,536 records |
| --- | --- | --- | --- |
| Parsing time | 50.184us | 2.652ms | 65.111ms |
| Parsing B/op | 18,449 | 1,506,948 | 27,210,922 |
| Construction time | 24.246us | 1.814ms | 30.390ms |
| Construction B/op | 23,248 | 1,504,689 | 23,975,993 |

`/tmp/abcgo-gig-before-bench.log`. Local measurements, not a general performance
claim. Planned candidates: ID->index instead of copying a full gigNode in byID;
compact kind tags instead of strings retaining RHS text. Keep safe sparse-ID
maps rather than C's count+100 allocation assumption.

## Full native construction evidence and adjacent adapter fix

A new Kit repeated-leaf regression fails before repair because the GIA adapter
uses raw AppendAnd with hash=false. Source kitHop.c uses AppendAnd2. Change the
adapter to that existing source operation; preserve local simplification,
nonhashed distinct nodes and simple-mode raw append. This is not global hashing
or a substitute synthesis algorithm.

A separate new red regression catches w0/w1 references: source parses them as
reserved constant references, while Go rejected them. Accept nonnegative wire
references, leaving declaration IDs0/1 prohibited by the existing caller guard.

The native matrix now passes1,160 cases, comparing parsed records, category
vectors, exact graph objects/fanins, CI/CO order, identities, registers, full
truth and public reports; independent Boolean checks also pass. It covers
abbreviated/full/0x/upper/lower hex forms, missing/constant/repeated/permuted LUT
pins, dependent LUTs, barriers/selections, Box4/5 fields, delays with zero through
multiple selectors, Seq outputs and mixed cut interfaces. Source CI grouping
and delay/PO/Seq CO grouping are retained, with zero registers/timing manager.

There are840 original-format comparisons (including explicitly corrected
six-input arity) and320 existing Go oversized-LUT4 cases. The latter retain
Go's low16-bit behavior; compare them with explicitly equivalent four-digit C
input, because unnormalized larger LUT4 tables can assert in the source ISOP
engine. The original oversized failure is independently characterized alongside
original LUT6 and selection failures. No production C source is changed.
Source headers are included before installing the fixture's narrow assertion
hook; preprocessed output confirms system assert.h does not overwrite it.

Logs: `/tmp/abcgo-gig-native-correct-assert.log` (5.212s),
`/tmp/abcgo-gig-kit-unhashed-red.log`,
`/tmp/abcgo-gig-constant-refs-red.log`.

## Completed coverage and compactness checkpoint

The expanded focused suite passes without skips in 9.189s, including existing
Kit hashing/bridge tests, 1,162 native comparisons, malformed inputs, sparse
IDs, whitespace compatibility, independent manager lifetimes and unchanged
parsed records. The bool/kit dependency suite passes in 0.326s. Logs:
`/tmp/abcgo-gig-final-complete-focused.log` and
`/tmp/abcgo-gig-kit-dependency.log`.

The final native matrix has 842 original-format cases and 320 explicitly
normalized wide-LUT4 inputs. It now includes w0/w1 references and alternating
LUT6/LUT4 construction with missing pins, checking scratch reset across widths.
All sixteen one-digit truth tables are exhausted for both LUT kinds and each
pin binding. Private allocation/count/parse/construct/stop and public reader
paths are covered by canonical state, exact graph, truth and lifetime checks.

Compact kind tags reduce each descriptor from 64 to 48 bytes. The construction
index stores only kinds, retaining sparse IDs; pin and Kit-cover scratch are
reused as in C, with missing pins explicitly cleared. Short local benchmarks:

| Operation (65,536 records) | Before | After |
| --- | --- | --- |
| Parsing time | 64.274 ms | 53.123 ms |
| Parsing B/op | 27,210,922 | 23,009,248 |
| Construction time | 30.197 ms | 13.066 ms |
| Construction B/op | 23,975,088 | 4,214,296 |

The 64-record parse was slightly slower (49.125 to 51.768 us); this is a local
allocation/throughput measurement, not a universal speed claim. Corrected
benchmark logs: `/tmp/abcgo-gig-before-compact-bench.log` and
`/tmp/abcgo-gig-after-compact-bench.log`. Full frozen integration remains pending;
earlier pending notes above describe checkpoints, not missing focused coverage.

## Frozen integration and conclusion

The tested production, tests and native fixture remained frozen at93bc5ee9d.
The uncached full-module race command completed successfully (handle12730,
exit0):130 tested packages and three command packages without tests, zero
failures/race diagnostics. GIA604.963s, ABCI29.069s, BMC23.681s. GIA vet completed
successfully (handle12790, exit0), with an empty log. Evidence:
`/tmp/abcgo-gig-final-module-race.log`, `/tmp/abcgo-gig-final-vet.log`.

All seven source routines are now reviewed, covered and reconciled; earlier
pending notes record intermediate checkpoints. Keep actual Kit synthesis and
source interface orders, with the explicitly documented source defects and
safe Go extensions. This closes this source family only. GIA and the broader
package audit remain open (original completion2/131); giaGlitch.c is next.
