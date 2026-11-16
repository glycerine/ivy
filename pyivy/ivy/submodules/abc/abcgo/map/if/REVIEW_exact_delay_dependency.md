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
- [x] Reconcile network-key normalization and cached result/failure ownership
  through supports3..8, including multiword keys and source first-hit policy.
- [ ] Review remaining cache-miss depth bounds, search/resource-limit behavior
  and reporting, including actual synthesis at supports3..8/multiword truth.
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


## Restore dynamically sized network normalization

Source Abc_ExactNormalizeArrivalTimesForNetwork (abcExact.c:510..560) sizes its
level vector from the network; it has no100-node limit. Go's dynamic-slice port
nevertheless returned early for more than100 gates or100 total inputs/gates.
Seven new boundary cases include99/100/101-node totals and full eight-variable
chains. Four cases fail before repair (/tmp/abcgo-exact-network-size-red.log,
exit1). Remove the artificial limits; retain input-size and forward-fanin checks
before copying levels back. Serialized byte fields bound the allocation size.
A99-gate cached-cost case verifies actual normalized-key reuse, not just the
standalone helper.

Go-only packed networks use the source header/gate/trailer layout. Independent
chain path depths give normalized keys[0,0,1,2,...], allowing exact expectations
without C tools or calling the production normalizer as its own oracle. Thirty-
six cached queries across supports3..8 cover offsets, variable arrival orders,
full-word truth comparisons/hash collisions and unchanged caller truth/arrivals/
inactive pin bytes. Cache hits must perform no SAT calls or extra insertions.
Both proven/limit-marked empty-cache categories retain infinity cost, unchanged
pins and their original insertion counters without retrying synthesis.

## Preserve source cache selection without overstating optimality

Source Ses_StoreGetEntry tests entries in list order after normalizing each
query against that entry's network. Abc_ExactDelayCost accepts the first hit
and ignores AigLevel on that path. It does not choose a fresh optimum for every
new arrival profile. A new explicit example stores two AND3 networks with
pin depths[2,2,1] and[1,2,2]. For arrivals[0,0,9], the newest matching network
returns11 although the older one gives10. Preserve the C cache/search policy;
changing its optimization strategy would be a separate algorithm change.
Correct the Go API comment, which previously claimed an optimal cached result.
This is a source-inherited optimization limitation, not a new Go fallback.

The source GetEntry loop also fails to advance after a synthesized-entry
normalization miss; Go's existing for-loop advances safely. Valid records
created by the same network-normalization routine match their stored key.
Retain safe iteration rather than introduce a source hang for inconsistent
records. No new policy, model approximation or solver substitution is added.

Frozen full Go-only module exits0 (130 tested packages,3 commands without
tests), focused exact/cache/IF/store/solver race1.159s and IF/GIA vet exit0.
Logs: /tmp/abcgo-exact-cache-final-{module.jsonl,race.log,vet.log}. All sessions
ended. Cache-miss depth/resource/reporting and single-output SES dependencies,
plus final GIA family/package gates, remain open; original audit2/131.


## Cache-miss synthesis, bounds and diagnostics

Twenty-four actual fresh synthesis cases cover AND/NAND supports3..8 in both
AIG and general-gate modes. An independent scalar packed-model decoder checks
all assignments, gate dependencies, physical pin paths, overall depth and the
attainable binary full-support lower bounds (n-1 gates, ceil(log2(n)) depth).
Uniform arrival shifts must reuse the cache without new SAT calls. These
fixtures and expected results are entirely Go, including multiword truth.
Five three-input depth hints cover zero/impossible, tight, unknown and loose
bounds. Impossible hints preserve pins and avoid SAT; other hints synthesize
the minimum two-gate, depth-two AND graph.

New failing diagnostic regressions precede repairs to the omitted verbose
cache-miss header, red resource-limit color and two-digit depth erasure. A
four-input parity fixture with conflict budget1 actually reaches the solver
resource limit; no solver stub or external tool is used. A complemented parity
regression also exposes debug logging from the wrong truth buffer: C logs its
normalized specification before cleanup. Log the owned SES specification,
retaining caller truth and the compact-buffer memory repair. Red logs:
/tmp/abcgo-exact-synthesis-report-red.log and
/tmp/abcgo-exact-synthesis-limit-red.log. Search strategy and returned failure
semantics are unchanged.

Frozen full Go-only module exits0 (130 tested packages,3 commands without
tests); focused exact/IF/SES race4.514s and IF/GIA vet exit0. Logs:
/tmp/abcgo-exact-synthesis-final-{module.jsonl,race.log,vet.log}. All sessions
ended. Single-output search/encoding dependencies still require reconciliation
and additional varied-function/arrival cases before closing this optional mode.
The original package audit remains2/131.


## Remove the inherited truth-evaluation scratch overflow

Ses_ManDeriveTruth clears four words per gate but the manager, like C's
abcExact.c:252, reserved only100 words. The26th gate therefore panics in Go
(and overruns memory in C). A new valid26-gate eight-variable packed AND-chain
regression fails before repair at index100; log:
/tmp/abcgo-exact-truth-scratch-red.log. Replace the fixed array with a reusable
slice sized to four words per encoded gate. Keep operation decoding, fanin
ordering, output inversion and borrowed-result lifetime unchanged. This is an
explicit repair of inherited memory unsafety, not an alternate synthesis
algorithm or a new gate-count restriction.

Twelve boundary models cover supports2/6/7/8 and25/26/100 gates, including
multiword truth, complemented outputs, repeated evaluation and reuse for a
shorter model. Scalar AND expectations are generated in Go. Frozen full
Go-only module exits0 (130 tested packages,3 commands without tests), focused
exact/IF/SES race4.838s and IF/GIA vet exit0. Logs:
/tmp/abcgo-exact-scratch-final-{module.jsonl,race.log,vet.log}. All sessions
ended. Remaining varied-function/arrival synthesis and single-output dependency
reconciliation are still open; original audit2/131.
