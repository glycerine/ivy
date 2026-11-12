# Native public choice drivers

The public CEC choice drivers now use the existing native DCH proof/choice engine. `ChoiceSnapshots` returns explicit errors and splits interleaved COs by `index % snapshotCount`, matching `Gia_ManEquivToChoices`. Source-shaped wrappers retain pointer/nil results and expose errors through `Cec_ParChc_t.Error` where that parameter exists. The original new/old CEC internal speculative solvers are not duplicated: both public multi-snapshot entry points select the real DCH backend. The single-circuit choice driver runs DAR's actual three-version synthesis with that backend.

## Readability pass

Parameter translation is centralized. Snapshot extraction uses the shared AIG partial copier and real GIA converters. One typed DCH engine handles both explicit snapshots and DAR synthesis. The dependency direction is CEC -> DAR -> DCH -> AIG/Satoko, with no upward dependency back to CEC/GIA. No global engine registration or hidden fallback was added.

## Separate bug pass

Verified the source stride (`i % nSnapshots == 0`) before defining snapshot layout. Reject zero snapshot counts, indivisible output counts, malformed models and insufficient register outputs. Keep every CI in each snapshot even if its logic cone does not use it. Preserve trailing register interface metadata; the DCH owner added metadata validation/preservation before enabling this path. Validate single-circuit input before conversion. Tests require real choice alternatives, rather than merely successful graph construction, and independently verify the resulting output functions.

## Validation and remaining boundary

`test_native.sh -race` passes tests for explicit associative alternatives, public vector drivers, invalid layouts and the full single-circuit DAR synthesis route. It also exercises actual public `Cec_ManVerifyTwoAigs` witness ownership and `Cec_FraigCombinational` reduction. Separate SAT-sweep tests run the original C ABC binary as described in `sweep_review.md`.

The remaining early `cecChoice.go` speculative-reduction helpers are still drafts and are not called by the new native drivers. Sequential correspondence wrappers in `cecCec.go` await the separately owned correspondence engine. Full parent compilation remains blocked by other unported GIA/CEC files; the temporary integration test leaves all product files visible and unchanged.
