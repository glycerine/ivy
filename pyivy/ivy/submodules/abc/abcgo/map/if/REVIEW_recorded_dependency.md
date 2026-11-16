# Recorded-library mapping dependency review (in progress)

This is a dependency of the sequential GIA mapping-boundary review, not a
completed audit of IF or ABCI. The source functions are in abcRec3.c and IF;
Go keeps library ownership in base/abci and dispatch in map/if to avoid cycles.

## Plan

- [x] Identify actual source entry points and existing Go implementations.
- [x] Add a failing real-backend mapping regression before replacing IF panics.
- [x] Connect both candidate evaluation and best-cut recomputation to source cost.
- [x] Default reconstruction to Abc_RecToGia3; preserve explicit Go overrides.
- [x] Restore writable pin storage without allocating for ordinary timing reads.
- [x] Exercise real mapping, timing, phases, registers and delay/area rounds.
- [ ] Review richer recorded classes, competing delay profiles/area ties and
  missing-class/small-support behavior independently before dependency closure.
- [ ] Reconcile remaining exact-SAT cost integration separately.

## Findings and repair

IF's two fUserRecLib cost branches unconditionally panic despite an existing
Go If_CutDelayRecCost3 implementation in ABCI. GIA reconstruction supports only
an explicit builder callback, while C directly calls Abc_RecToGia3. A new
Go-only two-input mapping case fails at the candidate-cost panic before repair:
/tmp/abcgo-recorded-mapping-red.log (exit1).

Register the existing cost and reconstruction functions at ABCI initialization.
IF's small synchronized registry removes the import cycle; it does not replace
library selection, canonicalization, delay profiles or GIA rebuilding. The
per-parameter builder remains an override. Snapshot callbacks before invocation
so registration locks are not held across backend code. The library's lifecycle
and state remain owned by ABCI, as in source.

The real backend then exposes missing writable If_CutPerm storage. C allocates
pin bytes with its cut; Go exposed an empty slice unless a different evaluator
had allocated it. A focused new regression fails before the storage repair:
/tmp/abcgo-recorded-pin-storage-red.log (exit1). Allocate pin bytes lazily at
the writable accessor, preserving cut ownership and independent cut copies.
Reuse that accessor for SOP pin writing rather than a duplicate allocator.

Inspection catches a side effect of the accessor change: ordinary timing reads
would allocate needlessly. A new regression confirms one allocation per call
before adjustment (/tmp/abcgo-recorded-ordinary-allocation-red.log, exit1).
Timing borrows the already-computed slice directly, retaining zero allocations
and the existing missing-profile fallback. No source cut algorithm is changed.

## Verification and scope

Unit checks cover writable/stable/independently copied pin storage, exact
callback arguments/results/mutations, missing backend diagnosis, registration
from inside a callback, default reconstruction and explicit-override precedence.
A hot-path check requires ordinary cut delay to remain8 with zero allocations.

ABCI's real six-input library contains every two-input AND input/output phase.
One minimal regression plus768 configurations exercise64 network phases,
three arrival profiles, combinational/sequential interfaces and delay/area modes.
All8 assignments per configuration use explicit Boolean expectations. Check
source gate count/arrival formulas, registers, identity, AIG representation and
arrival metadata ownership. Unsupported larger classes force selection of the
available two-input cuts; richer class-selection behavior remains open above.
An initial matrix attempted ordinary AIG post-hashing, which C also rejects
because these results have no LUT mapping; correct that fixture rather than
invent a hash/mapping policy. An initial unit fixture used an unavailable local
GIA alias; qualify the existing GIA function instead.

Frozen full module with CGO disabled and C compilers absent from PATH exits0:
130 tested packages pass,3 commands have no tests. IF focused race1.024s,
ABCI focused race8.398s, IF/GIA vet exits0. Logs:
/tmp/abcgo-recorded-final-{module.jsonl,if-race.log,abci-race.log,vet.log}.
All sessions ended. No C fixture/compiler dependency or substitute backend added.


## Source boundary repairs during selection review

Three new Go-only regressions fail before repair. If_CutComputeDelay converts
the leaf arrival before adding the integer pin delay, whereas C assigns the
float sum to an integer. Negative fractions crossing zero and float32 rounding
at16777216 expose different results. Restore conversion after float addition.
Evidence: /tmp/abcgo-TestRecordedDelayConvertsAfterFloatAddition-red.log, exit1.

Lms_GiaAreas uses an eager helper for C's conditional expression, evaluating
Lms_ObjArea on terminal/constant drivers and triggering its AND assertion.
A terminal/constant/complemented-output fixture fails before repair; restore
the conditional branch. Check exact areas[0,0,0,0,1,1] and delay profiles
[0,0,0,0,0x11,0x11]. Evidence:
/tmp/abcgo-TestRecordedLibraryTerminalAreaAndDelayProfiles-red.log, exit1.

The same eager helper dereferences a nil GIA when printing a function-only
library. A separate new regression fails at the nil access. Restore the source
branch and replace generic conditional helpers with a guarded percentage
helper, avoiding eager zero divisions and simplifying report expressions.
Check zero-try and nonzero counters, class counts and absence of graph creation.
Evidence: /tmp/abcgo-TestRecordedFunctionOnlyLibraryReport-red.log, exit1.

Frozen full Go-only module exits0 (130 tested packages,3 commands without
tests); recorded boundary/mapping race passes8.105s and IF/GIA vet exits0.
Logs: /tmp/abcgo-recorded-boundaries-final-{module.jsonl,race.log,vet.log}.
All sessions ended. Richer competing profiles/area ties, explicit missing-class
and small-support selection remain open; no dependency/package closure claimed.
