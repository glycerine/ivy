# ABCI integration review

This review completes the plan in [PORTING_PLAN.md](PORTING_PLAN.md).
It covers the full-package integration and new behavioral tests, rather than
claiming every command/option has been exhaustively verified.

## Fidelity and API ownership

Checked repaired paths against `src/base/abci`, including command option
blocks, DAR graph conversion, mapped LUT reconstruction, quantification,
strashing, DSD, resubstitution, IF mapping and Ivy proving. Owner APIs now
carry native Go edges, slices and typed parameter fields. Existing SAT,
mapper, optimizer and verification engines remain responsible for their
algorithms. Source stubs and source assertions remain identifiable as such.
No new algorithm substitute or build exclusion was used.

The shared network SAT implementation was moved to `base/abc` because both
ABCI and retiming need it; wrappers preserve ABCI's entry points. The move
eliminates the dependency cycle without introducing callbacks or another
solver. Full owner and dependent tests pass.

## Correctness review

Reviewed code separately from compilation for edge polarity, output/latch
ordering, model transfer, copy-field ownership, function representation,
truth widths, metadata and error propagation. Behavioral tests caught errors
that earlier helper-only tests missed: IF double complementation, resubstitution
replacement-edge panics, mapped-GIA function types, inactive phase mapping,
and invalid/deleted Ivy results. All identified failures were fixed before
integration.

Two source defects were deliberately corrected, with tests: reachability must
retain an all-one image before its early exit, and Ivy proving must return a
newly strashed network even when an early SAT check finishes immediately.
These are documented departures from defective C paths, not approximations.

## Test changes and false positives

Corrected old expectations only after checking source or API contracts:
retiming includes latch and BI/BO timing edges; small truth tables use their
actual word count; arrival times are normalized; absent MV onsets remain
unfilled; command failures include the `Error:` prefix. Network fixtures now
provide required interface names and initialized command frames. Tests were
not disabled to obtain a green build.

ECO loads the repository's real `abc.rc` and supplies required libraries. Its
outer return value alone was insufficient: the regression now also rejects
internal command errors. ESOP write-failure testing verifies both failure
status and diagnostic text.

## Compactness and readability

Used existing owner operations and small boundary adapters rather than
shadow solver/network representations. Quantification scratch state is local
and typed. Resubstitution no longer saves and restores function/latch Data
around simulation. AIG copying uses explicit edges consistently; C-visible
parameter fields have one owner. The original command structure remains
recognizable, although its size still warrants incremental cleanup driven by
tests rather than a wholesale rewrite.

## Validation and limits

Full normal, race, strict checkptr and `abc_use_cudd` package tests pass, along
with the affected dependency tests listed in the plan. Exhaustive small
truth tables and counterexample replay provide independent correctness
checks. Overall statement coverage is about 37.5%; specialized command paths,
large-design behavior and resource-limit combinations still need additional
validation. The passing suite is a usable integration baseline, not a proof
that every ABC algorithm or configuration is bug-free.
