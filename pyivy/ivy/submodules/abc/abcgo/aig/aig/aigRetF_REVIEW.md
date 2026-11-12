# Forward retiming review

Ported aigRetF.c's fixed-point PI reachability and forward retiming frontier to
Graph/Edge. The source manager remains mutable work storage; the returned manager
is structurally rehashed after sequential cleanup. Initialization is all zero,
with gate/output phase chosen to preserve initial output values.

Separate compactness/idiomatic pass: graph-local maps replace pNext pointer links,
Edge carries complement state explicitly, and the reachability/clock-domain
helpers keep the frontier scan short. No node pointer survives appending to the
manager's backing slice. Existing graph construction, buffer skipping, sequential
cleanup and duplication APIs are shared rather than reimplemented.

Separate bug pass and regressions:

- zero steps performs no retiming and preserves meaningful metadata;
- negative steps are rejected;
- explicit XOR gates retain XOR semantics and XOR initialization phase instead
  of incorrectly applying the source's AND-only rule;
- AND combinations with both complemented register inputs retain initial one;
- mixed clock domains are not merged; valid domain membership survives cleanup;
- real moves invalidate FlopNumbers/Reprs and set RetimingChanged; no-op calls
  preserve tracking and clone mutable metadata;
- duplicate-domain validation runs before reachability marking. Its regression
  was observed failing before the ordering fix, then passed;
- scratch marks are cleaned, including through the error-cleanup defer;
- buffer chains, feedback, and growing-manager iteration preserve trace behavior.

Ordered duplication required CI prebinding because a retimed old buffer can point
to a newly appended register input. The duplication owner fixed that shared API
with its own regression; RetF's multi-stage/step-limit test exercises integration.

Validation: the actual complete `aig/aig` package passes ordinary and race tests,
with no file exclusions. RetF tests compare zero-initialized traces across random
PI sequences for AND/XOR polarity combinations, two-stage limits, buffers and
feedback, and verify graph invariants and metadata contracts.
