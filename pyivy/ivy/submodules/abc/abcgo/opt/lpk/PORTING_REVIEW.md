# LPK Port Review

## Review status

This is a source comparison against `src/opt/lpk`. The port is incomplete:
`go test ./opt/lpk` reaches LPK but fails to compile. DSD subset enumeration,
MUX analysis, and portions of cut enumeration and mapping have been translated;
they have not yet been exercised as a package.

## Findings to resolve before integration

- The C `Abc_NtkDecFromTruth` calls `Abc_NtkLutmin`; the existing Go version
  skips that step and is not source-equivalent. The CUDD-backed LUT minimizer
  and its `base/abci` dependencies remain untranslated, and the Go package is
  not buildable. Do not treat the current shortcut as a completed port.
- `lpkCore.go` needs the source IF manager configuration and mapping-to-network
  routines, together with exact network traversal and fanout reference handling.
  `lpkCut.go` needs the Kit cloud manager and ABC `fPersist` scratch state.
- Truth-table arguments and scratch workspaces have been changed to Go slices in the main mapping paths. Remaining work is to type-check every call site and verify the 32-bit truth words versus the separate 64-bit `word` utility type.
- Executable Vec iterator macros and the common allocation/copy/zero idioms have been replaced with indexed loops, `make`, `copy`, and `clear`. A follow-up scan should confirm no C-only statements remain.
- Numerous C flag fields are integer values used directly as Go conditions. Convert these to explicit `!= 0` checks while preserving the public C-shaped parameter representation.
- Cut fanin iteration now uses ABC object fanin slices. Cut array access and object scratch-state handling still need a type-checked semantic review, followed by cut enumeration and dominance tests.
- `Lpk_Implement_rec` and decomposition cleanup need lifecycle tests: verify each generated component is created once, leaves are replaced with the intended complemented edge, and failed decomposition removes all temporary functions.
- Current tests cover LUT-size arithmetic and support delay/index extraction only. Add truth-table, DSD/MUX split, cut enumeration, and network replacement tests against small known functions.

## Findings resolved during this pass

- Renamed the DSD translation unit’s private `Lpk_PrintSetOne` and `Lpk_PrintSets` helpers. C `static` functions can share names across files; Go package-level functions cannot.
- Corrected the `word` alias to 64 bits, matching `bool/kit`’s C-compatible definition. Truth-table storage remains `[]uint32`, matching the Kit API.
- `Lpk_NodeHasChanged` now accepts both `int` and `int32` object IDs in the Go vector. The source stores IDs through pointer-sized integer casts; the previous Go-only `int` assertion incorrectly treated the `int32` IDs used by network objects as changed.
- Added focused tests for support minimization (including truth compaction and fanin delay metadata), cofactor support extraction, and cut signature duplicate filtering.

## Current checks

- `gofmt -w opt/lpk/*.go`: passes.
- `GOCACHE=/tmp/abcgo-gocache go test ./opt/lpk`: currently stops before LPK
  type checking because `bdd/cudd/cuddMatMult.go` contains raw C preprocessor
  syntax. The focused LPK regressions cannot run until that dependency compiles.
- Race and vet remain blocked by the package build failure.
