# New-choice source-path status

The active `Dar_ManChoiceNew` in `src/opt/dar/darScript.c` creates three
synthesis snapshots, combines their GIA forms with `Gia_ManChoiceMiter`, and
then chooses one of three proof paths. The default converts the interleaved GIA
miter with `Gia_ManToAigSkip(..., 3)` and calls `Dch_ComputeChoices`; the
`fUseGia` and `fUseNew` paths call the corresponding CEC choice engines.

`opt/dar.Dar_ManChoiceNew` now invokes this source path when no engine is
provided. `ChoiceOptions.DchParams` selects the original DCH controls and nil
uses ABC defaults. An explicitly supplied engine remains an opt-in override;
partitioned choices require an explicit FRAIG callback. The distinct
`Dar_ManChoiceNewAig` entry point uses `Dch_DeriveTotalAig`, as its C source
does. That routine is not interchangeable with the GIA miter path: the
`TestChoiceMiterAndTotalAIGHaveDifferentNodeOrder` fixture shows different
intermediate AIG node ordering.

The facade's source path is wired through `proof/cec.DarSourceChoiceEngine`.
The engine creates and stops GIA snapshots and their miter, preserves the AIG
snapshots owned by its caller, and dispatches to the DCH, CEC, or Cec4 source
branch. A compiled original-C miter fixture compares node kinds, literals and
ordering; the lower facade's ABC `&dch`, `-g`, `-x` and `-g -x` option matrix
compares node counts, choice counts and truth tables. These fixtures cover
representative cases, not all circuits or choices.

On success the source-named new-choice entry consumes its input AIG, transfers
its timing manager to the result before recomputing levels, and releases the
intermediate AIGs. On a Go error the input remains available.

## Package boundary

The former import cycle included GIA and CNF as well as direct CEC imports:

```text
opt/dar -> proof/cec -> aig/gia -> opt/dar
opt/dar -> proof/cec -> aig/gia -> sat/cnf -> opt/dar
opt/dar -> proof/cec -> proof/fracomb or proof/framgr -> opt/dar
```

The existing DAR implementation and its internal tests now live in
`opt/dar/core`. GIA, CNF, FRACOMB, FRAMGR and the CEC implementation import
that core; `opt/dar` is a compatibility facade with type aliases and forwarding
functions. The facade can therefore import the CEC integration layer without
a cycle. No algorithm copy, global registration or partitioned fallback is
used by the lower DAR default.

One separate boundary remains: `aig/gia.Gia_ManPerformDch` cannot import the
CEC integration layer because CEC already imports GIA. Its existing default
uses an explicit DCH engine over total AIGs; it is not the same source path as
the lower DAR facade. A future source-faithful GIA entry point needs an
explicit callback supplied from a higher layer or another package split. This
limitation is independent of the fixed lower DAR default.
