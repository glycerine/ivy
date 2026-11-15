# Gate-Inverter Graph reader review (in progress)

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
3. [ ] Compare parsed records, category orders, complete constructed graph,
   Boolean truth and public reader reports against actual C. Cover all record
   types, constants/missing pins, dependent LUTs, and shortened/full hex forms.
4. [ ] Preserve original C LUT6 assertion failure, and test six-input behavior
   only with an explicit narrowly checked correction of its k==4 typo.
5. [ ] Expand validation, sparse-ID/whitespace compatibility, resources and
   metadata ownership checks. Distinguish original C defects/preconditions.
6. [ ] Measure compactness and remove duplicated node descriptors/string
   retention while preserving sparse IDs, source order and synthesis.
7. [ ] Reconcile all seven routines, freeze code/tests/fixtures, run uncached
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
(`/tmp/abcgo-gig-minimal-repaired.log`). Full native/parser/constructor and
integration coverage remain pending.

## Native oracle design

The C fixture includes actual giaGig.c. Original mode preserves every source
assertion, including the known line311 LUT6 k==4 typo. Corrected mode intercepts
only that exact file/line/expression and independently requires six stored
fanins for every LUT6; all other assertions remain active. It validates the
number of corrections across private and public reader calls. A source-line
check makes a changed upstream assertion fail the fixture setup explicitly.

The fixture exposes canonical parsed records and category order vectors, then
prints exact object kinds/fanin literals, CI/CO vectors, identity, register
count and truth. The Go comparator will use these source outputs plus an
independent record-level truth oracle. Raw failure cases remain separate from
corrected valid-input comparisons. No production C source is changed.

## Source inventory

| Original C | Go implementation | Review status |
| --- | --- | --- |
| Gls_ManAlloc | gigParse/gigConstruct private records | Native state/lifetime checks pending |
| Gls_ManStop | Go-owned records and file defer | Resource checks pending |
| Gls_ManCount | gigParse and reader count report | Native report repaired; broader counts pending |
| Gls_ManParseOne | gigRef | Literal/ID edge tests pending |
| Gls_ManParse | gigParse | Short-truth and selection regressions repaired |
| Gls_ManConstruct | gigConstruct | Existing sample passes; full graph matrix pending |
| Gia_ManReadGig | Gia_ManReadGig | Native success report passes; complete matrix pending |

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
