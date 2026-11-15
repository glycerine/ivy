# Glitch simulation review (in progress)

Scope: all 24 graph/simulation routines in `src/aig/gia/giaGlitch.c`, plus
inline graph accessors. Work is sequential, without agents or branches.
GIG reader review closed at2beee0de9; wider GIA audit remains open (2/131).

## Plan

1. [x] Inspect C and current Go graph/simulation representation and tests.
2. [ ] Add a new failing regression before repairing each confirmed defect.
3. [ ] Compare graph construction, state transitions, sequential prefixes,
   RNG consumption and reports against actual C, normalizing packed-offset
   handles to Go ordinal handles explicitly.
4. [ ] Add independent truth/pulse/sequential oracles, boundary/validation and
   ownership coverage. Characterize source defects separately from Go safety.
5. [ ] Measure and simplify redundant storage without algorithm substitution.
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
