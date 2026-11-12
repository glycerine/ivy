# `ifDec75` implementation review

## Reviewed scope

Reviewed `ifDec75.go`, the DSD boundset parser in `ifDec75Dsd.go`, and the
focused parser regression tests. The source behavior was compared with
`src/map/if/ifDec75.c`.

## Findings and fixes

- Replaced mechanically translated declarations and C syntax with Go calls to
  the native truth-table, DSD decomposition, and seven-input minimum-base
  helpers.
- Kept the C ABI's `*uint32` truth pointer at the boundary. The adapter views
  it as 64-bit truth words, matching the original C cast. Callers must supply
  the full aligned truth table; this is a pre-existing pointer contract.
- Replaced fixed-size DSD text buffers with allocated byte slices supplied to
  `Dau_DsdDecompose`.
- Made malformed or mismatched DSD delimiters return an error internally;
  exported checks panic with context, matching the source's assertion-based
  failure behavior while avoiding an out-of-bounds walk.
- Removed the unused static diagnostic counter from the C implementation; it
  did not affect filtering results and would have introduced a data race.
- Added regression cases for six/seven/eight-variable boundset masks, nested
  DSD forms, trailing C NUL bytes, and malformed delimiters.

## Verification and remaining integration gate

The focused parser suite passes with:

```sh
go test ./map/if/ifDec75Dsd.go ./map/if/ifDec75Dsd_test.go
```

The normal `go test ./map/if` gate is currently blocked before reaching this
code by syntax errors in the sibling `ifDelay.go` file. This review does not
claim that the whole package builds. Re-run package tests and race/vet checks
after the sibling package work is integrated.
