# Glitch simulation review (in progress)

Scope: all 24 graph/simulation routines in `src/aig/gia/giaGlitch.c`, plus
inline graph accessors. Work is sequential, without agents or branches.
GIG reader review closed at2beee0de9; wider GIA audit remains open (2/131).

## Plan

1. [x] Inspect C and current Go graph/simulation representation and tests.
2. [x] Add a new failing regression before repairing each confirmed defect.
3. [x] Compare graph construction, state transitions, sequential prefixes,
   RNG consumption and reports against actual C, normalizing packed-offset
   handles to Go ordinal handles explicitly.
4. [x] Add independent truth/pulse/sequential oracles, boundary/validation and
   ownership coverage. Characterize source defects separately from Go safety.
5. [x] Measure and simplify redundant storage without algorithm substitution.
6. [ ] Reconcile every routine; freeze code/tests/fixtures for uncached full
   module race testing and GIA vet; record terminal evidence and commit.

## Initial findings to verify

C reuses construction counters through unions as switch/glitch counters.
Finalize zeroes those union fields. Go separated them and only zeroes link
counters, potentially retaining activity across simulations. Test before repair.

Go object printing differs in whitespace, and verbose simulation omits the
source memory field and source count/time formatting. Native comparison should
make any report correction concrete before implementation.

C's standalone Ro/Ri accessors use register-count offsets, unlike its correct
PI/PO-count iteration macros. Retain Go's correct offsets; expose the original
source defect explicitly. C declares three-bit fanin counts and a six-element
sequential scratch array despite accepting <=16; valid native comparisons must
stay within six fanins. Existing Go support beyond that is a safety extension.

## Source inventory

Graph/lifetime: Gli_ManAlloc, Gli_ManStop, Gli_ManPrintObjects,
Gli_ManFinalize, Gli_ObjAddFanin, Gli_ObjAlloc, Gli_ManCreateCi,
Gli_ManCreateCo, Gli_ManCreateNode; inline graph/port accessors.

Truth/activity: Gli_NodeComputeValue, Gli_NodeComputeValue2,
Gli_ObjNumSwitches, Gli_ObjNumGlitches, Gli_ManSetPiRandom,
Gli_ManSetPiFromSaved, Gli_ManSwitching, Gli_ManGlitching, Gli_ManVerify.

Sequential/driver: Gli_ManSimulateSeqNode, Gli_ManUpdateRandomInput,
Gli_ManSimulateSeqPref, Gli_ManSetDataSaved, Gli_ManSetPiRandomSeq,
Gli_ManSwitchesAndGlitches.

Current focused tests cover graph construction, one reconvergent pulse and a
two-input AND word. Broader native/state/RNG/ownership coverage is pending.

## First regression and native checkpoint

New counter regressions both fail before repair (direct finalize retains17/23;
a second zero-pattern driver retains31/31). `/tmp/abcgo-glitch-finalize-red.log`,
handle77463 exit1. Reset separate activity fields alongside construction fields,
matching C's unions. All five original/minimal tests pass0.006s after repair.

The native fixture includes actual giaGlitch.c unchanged. Its first104 complete
state/RNG traces pass: normalized graph identity/links, all phases/marks/counts,
changed/affected/frontier ordering, saved words, combinational and sequential
steps, prefixes and public driver. A separate report regression fails before
fix on Node-line whitespace (`/tmp/abcgo-glitch-native-first.log`,handle51078
exit1). The new verbose report also fails before fix: missing Memory, spacing
and source clock format (`/tmp/abcgo-glitch-verbose-red.log`,handle12734 exit1).
Restore report shape; the memory amount counts Go graph storage, excluding
borrowed truths/work vectors just as C's packed estimate excludes them. The
representation-specific number is intentionally not a C packed-byte claim.

The expanded suite passes9.504s without skips
(`/tmp/abcgo-glitch-expanded-first.log`,handle15694 exit0):104 native state
traces,40 native random-word streams of33 updates with next-RNG checks, both
reports,1,046 scalar/packed truth tables (every function up to three inputs),
independent sequential prefixes and explicit multiword Go safety extensions.

A new source-precondition comparison for NaN transition probabilities and
storage measurements are pending. Wider coverage/compactness and frozen full
integration remain open; this is a red review checkpoint, not closure.

## Validation, source defects and final compactness

The new NaN regressions fail before repair for both input-update paths, after
verifying that unmodified C rejects the same input
(`/tmp/abcgo-glitch-nan-red.log`,handle11339 exit1). Use C's strict positive
comparison so NaN is rejected too. Other invalid probabilities, port/bit bounds,
allocation/link errors, incomplete graphs, verification failures and teardown
are covered. Borrowed truth tables retain source ownership semantics, while
managers own independent graph state. Go's dynamic allocation-hint extension
is explicitly checked with zero-size hints and subsequent graph growth.

The original register accessor defect is characterized unchanged: standalone
C Ro/Ri helpers use register-count offsets, while source simulation macros use
PI/PO-count offsets. The fixture reports ro1/3 and ri6/7 on unequal port counts;
Go keeps the correct macro-consistent offsets. Native valid-domain tests stay
within six fanins; independent multiword tests retain Go's existing safe
7/8/12/16-fanin support instead of reproducing C bitfield/stack corruption.

Source reconciliation:

| Routines | Evidence |
| --- | --- |
| Allocation/stop/finalize, object allocation/linking, CI/CO/node creation | Native graph and complete state traces; new red reset tests; invalid links, growth hints and lifetime checks |
| Graph/port/type/link/count accessors | Native normalized structure; direct classification/order/count tests; explicit source Ro/Ri defect |
| PrintObjects | New red native report regression; only packed-offset handles normalized |
| ComputeValue/ComputeValue2, SimulateSeqNode | Native steps;1,046 independent scalar/packed tables, every Boolean function up to three inputs; explicit multiword extension |
| NumSwitches/NumGlitches, Switching/Glitching/Verify | Native counters/frontiers;126 independent reconvergent pulse transitions;256-deep propagation, duplicate and128 fanout branches; validation failures |
| SetPiRandom/SetPiFromSaved/SetPiRandomSeq | Native changed ordering, phases/counters, snapshots and next RNG; source NaN rejection |
| UpdateRandomInput |40 native streams of33 updates, five initial words and eight probabilities including half/adjacent floats/0/1; next RNG |
| SimulateSeqPref/SetDataSaved | Native full-word/snapshot states;104 independent scalar-lane prefix configurations and saved bits0/1/15/31 |
| SwitchesAndGlitches | Native comb/sequential drivers at0/1/31/32/33/65 patterns; reports; red repeated-driver count reset |

Remove redundant construction counters and manager object counts; slice lengths
already provide them. Remove unused packed-buffer/traversal fields; keep expected
fanin/fanout counts because finalization must validate incomplete construction.
Use one bounded link backing array per object, preserving both port orders.
Reuse saved-word storage, as C does. Bound five initial vector reserves by graph
size instead of allocating1,000 entries each for tiny graphs. Field ordering
reduces padding. No event, RNG, truth or sequential algorithm is replaced.

Short local150ms benchmark results (not universal timing claims):

| Chain nodes | Before ns/op | After ns/op | Before B/op | After B/op | Before allocs | After allocs |
| --- | --- | --- | --- | --- | --- | --- |
|64|27,508|20,549|37,761|17,648|401|337|
|4,096|1,614,950|1,676,325|1,096,237|1,030,654|24,593|20,497|
|65,536|28,192,190|27,543,992|17,091,982|16,028,178|393,250|327,697|

Object size falls128 to112bytes. The medium case is slightly slower in this
short run; allocation decreases at every size. Logs:
`/tmp/abcgo-glitch-before-compact-bench.log` and
`/tmp/abcgo-glitch-final-compact-bench.log` (handles92192/98781 exit0).
The precompact suite passes13.258s; the first final-layout suite passes13.472s.
Final extra accessor/deep-fanout coverage and frozen full integration are pending.

Final focused coverage, including the added accessor/growing-hint and deep
fanout tests, passes all19 tests without skips in13.264s
(`/tmp/abcgo-glitch-final-complete-focused.log`,handle22185 exit0).
Only the frozen full-module race/vet integration gate remains for this family.
