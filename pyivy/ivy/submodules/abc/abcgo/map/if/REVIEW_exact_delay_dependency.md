# Exact-delay mapping dependency review (in progress)

This is the optional fUserSesLib cost dependency of the sequential GIA mapping
boundary review. It does not mark IF or ABCI audited. Source IF delegates to
Abc_ExactDelayCost in abcExact.c; Go already contains the SES/BSAT backend.

## Plan

- [x] Replace both unported IF call-site panics with the existing cost backend.
- [x] Preserve source argument order, user flag, pin mutations, failure markers
  and the successful-cost-field quirk; use one shared evaluation helper.
- [x] Register ABCI's implementation without an import cycle and adapt signed
  pin bytes without losing inactive bytes or allocating a conversion slice.
- [x] Add a failing real-backend regression before wiring the call sites.
- [x] Repair compact truth-buffer safety exposed by race/checkptr, preceded by
  a new focused failing memory-safety regression.
- [x] Check AND/XOR phases, arrival profiles, AIG/general gate modes, actual SAT
  calls/cache reuse, and source-unsupported GIA reconstruction.
- [ ] Review complete Abc_ExactDelayCost normalization, bounds, result/cache
  lifecycle and resource-limit cases, including supports3..8/multiword truth.
- [ ] Reconcile the single-output SES search/encoding dependencies used by that
  API before claiming the full optional mode reviewed.
- [ ] Run remaining final GIA mapping-family/module gates.

## Source binding and behavior

In ifMap.c:196..210 and424..438, source initializes local Cost to zero, marks
the cut as user-defined and calls the exact cost routine with truth, leaf count,
integer arrival profile, writable pin bytes and the structural AIG bound.
Successful Cost is not assigned to pCut->Cost. Only Cost==ABC_INFINITY fills
pins with IF_BIG_CHAR and sets IF_COST_MAX/fUseless. Preserve this behavior,
including zero mapped area in the default two-input exact-mode fixtures.

Go previously panicked at both call sites. A new real IF/SES mapping case fails
before repair (/tmp/abcgo-exact-if-mapping-red.log, exit1). A small synchronized
registry snapshots the callback before invocation, matching the existing
recorded-library connection. ABCI registers Abc_ExactDelayCost through a bounded
signed-byte adapter. It preserves inactive bytes, including high-bit values;
no alternate solver or network approximation is introduced by the integration.

GIA reconstruction is a separate source limitation: Gia_ManFromIfAig accepts
the SES flag in its initial assertion but has no SES branch and asserts in its
default case (giaIf.c:959..1000). Keep the corresponding Go rejection and test
it, rather than inventing GIA reconstruction from the new cost registration.

## Compact truth-buffer memory repair

The first full Go-only gate passes, but the focused race gate reports a fatal
checkptr error at Ses_ManAlloc's four-word unsafe view. IF supplies an active
one-word truth table for two-variable cuts. SES's lower-level specification
protocol instead uses four-word slots per output, including inversion/restore
of all four words. Existing allocator tests correctly use padded arrays and
therefore did not expose the cost-boundary mismatch.

Add TestExactDelayCostAcceptsCompactTruthWords with an exactly one-word heap
allocation and both truth polarities. It fails under race/checkptr before the
repair (/tmp/abcgo-exact-compact-truth-red.log, exit1). On a synthesis cache miss,
copy only the active caller words into an owned four-word slot before allocating
SES. Retain the lower-level slot layout and normalization/search algorithm;
caller truth remains unchanged. No checkptr suppression or oversized unsafe
view is used as a workaround. The original four-word allocator normalization/
restore regression remains active.

## Verification and remaining scope

Unit checks cover callback truth/profile/bound arguments, visit-mark cleanup,
reentrant registration, success retaining the previous Cost field, infinity
failure markers, absent backend and unsupported source GIA reconstruction.
ABCI tests include the minimal red regression plus48 AND/XOR phase/timing/gate
mode configurations, actual SAT synthesis and cache hits, selected function
truth, pin depths and source area semantics. Adapter tests cover zero support
and signed/inactive pin bits. The compact-input regression checks both truth
polarities and restoration. These are Go-only fixtures, with no C compiler.

Frozen full Go-only module exits0:130 tested packages pass and3 commands have
no tests. Focused IF race1.033s, ABCI exact/API/store race1.169s, IF/GIA vet exit0.
Logs: /tmp/abcgo-exact-if-reviewed-{module.jsonl,if-race.log,abci-race.log,
vet.log}. All sessions ended. Deeper exact-cost/SES dependencies and final
mapping-family/package audit remain open; original package completion2/131.
