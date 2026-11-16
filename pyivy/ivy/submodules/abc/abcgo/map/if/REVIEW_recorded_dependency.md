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
- [x] Review competing delay profiles/area ties and missing-class/small-support
  behavior with independent physical profiles and Boolean reconstruction.
- [x] Check source-supported IF reconstruction widths6..15, sixteen-variable
  LMS storage/profiles, multiword keys and the truth-cache rejection boundary.
- [x] Reconcile C char area storage as unsigned gate counts; cover127/128 and
  the source253-node traversal-volume limit, including cut leaves.
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


## Competing recorded implementations and compact selection

Reconcile If_CutFindBestStruct, If_CutDelayRecCost3 and Abc_RecToGia3 against
abcRec3.c:895..1173: canonicalize, look up adjacent PO ranges, minimize output
arrival, break delay ties by area, retain the first equal delay/area candidate,
remap pin depths to cut order and rebuild the chosen physical cone. Trivial
cuts deliberately cost1 even when their rebuilt graph has no gates. Missing
classes set maximum cost and leave the optional best-PO output unchanged;
trivial cases also leave canonical-output parameters untouched. Preserve those
source policies rather than normalizing them into different behavior.

New Go-only library has16 explicit truth groups with five implementations each.
Independent profiles[222,221,212,122,122] and areas[3,2,2,2,2] check all80
stored records and17 class offsets. Across64 phase/timing configurations, a
stable-sort oracle ranks the manual profiles independently of the production
scan. Check selected PO, delay/area, canonical pin remapping, required-time
propagation for integer arrivals, both raw/hashed reconstruction modes, all8
assignments, two-gate area and clean library marks after repeated rebuilding.

Additional tests cover a missing three-input class in a two-input library and
eight constant/positive/negative single-support cases with support reduction
on/off. Check sentinel pin delays, unchanged unused output arguments, direct
zero-gate rebuilding and all assignments. No C fixture/compiler is used.

Simplify the two delay/selection helpers with scoped locals, shared trivial-cut
handling and removal of the source's permanently disabled diagnostic block.
Keep the original linear scan and tie policy. Inspection during this refactor
catches assertion ordering: invalid holey multi-input support must fail before
cost/pin writes. A new red regression demonstrates the temporary mutation
(/tmp/abcgo-recorded-selection-guard-red.log, exit1); restore the source guard
order before finishing. This was a refactor defect caught before committing,
not a newly discovered defect in the previously committed implementation.

Frozen cases pass0.417s; full Go-only module exits0 (130 tested packages plus3
commands without tests), focused recorded/mapping race8.391s and IF/GIA vet
exits0. Logs: /tmp/abcgo-recorded-selection-final-{cases.log,module.jsonl,
race.log,vet.log}. All sessions ended. Wide/multiword boundaries, exact-SAT
integration and final mapping-family gates remain open; original audit2/131.


## Eliminate recorded truth-stretch temporary copies

Source Abc_TtStretch5 repeats a prefix in place and returns immediately when
widths already agree. Go's lmsStretch5 converted all1024 workspace words into
an8KB uint32 buffer and back on every cut. Four new allocation regressions
fail before repair (/tmp/abcgo-recorded-stretch-red.log, exit1), including the
sixteen-to-sixteen no-op case.

Retain the exact source representation: below six variables repeat the first
native uint32 into the required words; at wider sizes the source blocks align
to whole uint64 words, so reuse the existing Abc_TtStretch6 helper. Do not scan
or rewrite unused workspace suffixes. An independent scalar32-bit specification
checks every153 source/target width pair0..16, with distinct halves and a poisoned
extra word. All outputs match before/after; only redundant work is removed.

Local150ms benchmark samples (indicative, not universal timing claims):

| Widths | Before ns/op | After ns/op | Before B/op | After B/op |
| --- | ---: | ---: | ---: | ---: |
|3 to6|8218|2.892|8192|0|
|6 to16|25796|5046|8192|0|
|7 to16|23671|4791|8192|0|
|16 to16|8477|2.764|8192|0|

All four go from one allocation to zero. Logs:
/tmp/abcgo-recorded-stretch-{before,after}.log. Frozen recorded cases pass0.268s;
full Go-only module exits0 (130 tested packages,3 commands without tests),
focused recorded/mapping race2.607s and IF/GIA vet exit0. Logs:
/tmp/abcgo-recorded-stretch-final-{cases.log,module.jsonl,race.log,vet.log}.
All sessions ended. Full wide signed reconstruction and exact-SAT work remain
open; width-helper coverage alone does not close the dependency or package.


## Wide recorded-library keys, phases and reconstruction limits

Add Go-only canonical-minterm libraries with independent scalar truth and
physical-depth expectations. The canonicalizer supplies the stored key; an
actual chain/balanced GIA realizes it, and scalar evaluation verifies every
rebuilt input assignment. Mixed/small/full supports distinguish cut word count
from library word count, exercising prefix stretching and unused library CIs.

The initial matrix tries sixteen-variable IF truth mapping and is rejected by
the existing guard (/tmp/abcgo-recorded-wide-first.log). Source LMS_VAR_MAX is16,
but source IF_MAX_FUNC_LUTSIZE is15 and its truth-cache arrays have16 slots.
Constructing IF truth mode at16 would index past them. Preserve the supported
source domain and Go's rejection instead of widening arrays or bypassing the
precondition. This was an invalid fixture assumption, not a production defect.

Full paths cover widths6..15 and supports2,3,5,6/full width:372 signed cases,
744 raw/hashed reconstructed graphs and1,063,648 explicit assignments. Check
multiword keys, canonical permutations, pin profiles, source cost/arrival,
IF timing readback and exact physical area. Twelve additional sixteen-variable
LMS-only cases cover small/full supports,1024-word storage, both graph shapes
and source four-bit depth saturation. A16-input chain's physical depth15 becomes
profile14 because C saturates the CI-offset nibble before subtracting at the CO;
characterize that storage behavior without claiming valid16-input IF mapping.

Precondition tests retain rejection of unsupported IF width and verify that
library-limit mismatch or construction-mode use fail before preparing state
or altering cut cost. No production change was needed beyond the separately
committed truth-stretch optimization. Source char area values above127 remain
a separate inspection item: source recording allows up to253 internal nodes,
so signedness needs reconciliation before declaring all runtime cases reviewed.

Frozen recorded cases pass1.708s; full Go-only module exits0 (130 tested packages
plus3 commands without tests), focused recorded/mapping race10.982s and IF/GIA
vet exits0. Logs: /tmp/abcgo-recorded-wide-final-{cases.log,module.jsonl,race.log,
vet.log}. All sessions ended. Byte-area policy, exact-SAT integration and final
mapping-family/package gates remain open; original audit2/131.


## Reconcile recorded area-byte interpretation

Source stores physical AND counts in Vec_Str_t, whose elements are plain char,
then promotes them to int during area tie breaking and to a12-bit unsigned
Cost field. C char signedness and out-of-range conversion are implementation
dependent; repository flags do not force unsigned char. On a signed-char build,
128 may promote to-128 and win a delay tie against127, then become3968 in the
12-bit Cost field. Reproducing that platform quirk would break nonnegative
gate-count ordering. Retain Go's existing unsigned byte interpretation, matching
unsigned-char builds and the stated vAreas gate-count meaning. Document this
portable numeric interpretation at Lms_GiaAreas; no production comparison or
selection algorithm is replaced, and no new numeric policy is introduced here.
This explicitly does not claim bit-for-bit equivalence to signed-char C builds
for counts above127.

Correct the earlier volume description: If_CutTraverse in ifUtil.c:565..581
pushes cut leaves before internal nodes. abcRec3.c:752 rejects traversal volume
greater than253, not253 internal ANDs. A two-input cut can have at most251
internal nodes under that check. New valid raw trees cover126,127,128,129,250,
251 with equal exact depth8, below nibble saturation. Sixty-four stored area/
profile records and36 signed/timing cases check both cross-boundary ties and
sole128/251 implementations. Seventy-two raw/hashed rebuilds verify all4
assignments and expected physical sharing. The first test incorrectly expected
hashing alone to minimize every redundant tree to one gate; correct the fixture
expectation to require actual sharing and preserved truth, without changing
source structural hashing into Boolean minimization.

Frozen recorded cases pass1.735s; full Go-only module exits0 (130 tested packages,
3 commands without tests), focused recorded/mapping race11.160s and IF/GIA vet
exit0. Logs: /tmp/abcgo-recorded-area-final-{cases.log,module.jsonl,race.log,
vet.log}. All sessions ended. All recorded runtime-dependency plan rows are
reviewed; exact-SAT integration is the separate remaining optional cost mode.
Final mapping-family/package gates and original audit remain open2/131.
