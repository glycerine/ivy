# New-choice source-path status

The active `Dar_ManChoiceNew` in `src/opt/dar/darScript.c` creates three
synthesis snapshots, combines their GIA forms with `Gia_ManChoiceMiter`, and
then chooses one of three proof paths. The default converts the interleaved GIA
miter with `Gia_ManToAigSkip(..., 3)` and calls `Dch_ComputeChoices`; the
`fUseGia` and `fUseNew` paths call the corresponding CEC choice engines.

The current DAR entry point does **not** yet reproduce this path: absent an
explicit injected engine it returns an error. The
existing `dch.Engine` calls `Dch_DeriveTotalAig`, the distinct constructor used
by `Dar_ManChoiceNewAig`, rather than the active new-choice GIA miter path.
`TestChoiceMiterAndTotalAIGHaveDifferentNodeOrder` demonstrates that those
constructors do not even assign the same AIG node order for a synthesis
fixture; equivalent outputs alone do not establish equivalent DCH choices.

Directly importing the existing GIA or CEC packages from DAR creates an import
cycle: GIA's synthesis code imports DAR, and CEC's choice code imports DAR.
`proof/cec.Dar_ManChoiceNew` provides the source-named, typed Go entry point
through a higher layer: it creates the actual GIA miter and selects the DCH,
CEC, or Cec4 path. The lower-level `opt/dar.Dar_ManChoiceNew` still requires an
explicit engine because it cannot import that higher layer without a cycle;
it reports an error if none is provided. Neither path silently falls back to
the partitioned engine. `Dar_ManChoiceNewAig` uses the source's
`Dch_DeriveTotalAig`/DCH path by default. The explicit partitioned engine is a
separate source operation and requires the actual FRAIG callback supplied by
the higher layer.

On success both source-named new-choice wrappers consume the input AIG, like
the C entry point, and transfer its timing manager to the result before
recomputing choice levels. On a Go error the input graph remains available.
