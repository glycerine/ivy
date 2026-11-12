# `proof/int` port status

`proof/int` is the McMillan interpolation engine. Its source includes
`intCheck.c`, `intContain.c`, `intCore.c`, `intCtrex.c`, `intDup.c`,
`intFrames.c`, `intInter.c`, `intM114.c`, `intM114p.c`, `intMan.c`, and
`intUtil.c`, with declarations in `int.h` and `intInt.h`.

## Current state

The generated `.go` files are still starting-point drafts, not a compiling Go
package. `go test ./proof/int` currently stops on malformed C-style macro loop
expansions in `intCheck.go` (first error: line 70). `gofmt` independently finds
parse failures in all 14 Go files. Examples include assignment expressions in
`for` conditions, C casts, raw `extern` declarations, and an unconverted
`#ifdef ABC_USE_LIBRARIES` region in `intM114p.go`.

Fixing syntax is only the first integration step. The package also assumes C's
global include namespace: it currently declares no aliases or imports for AIG,
SAIG, vector, CNF, SAT, counterexample, or Fraig APIs. The generated model uses
C-style object pointers and `pData`; the current `aig/aig` API represents
objects as `aig.Edge` values in `aig.Graph`, so a mechanical import alone will
not make these drafts type-correct.

## Dependency and porting order

1. Define the package boundary and import the current `aig.Graph` APIs, with
   explicit typed aliases only where they preserve real semantics. Coordinate
   SAIG, CNF, SAT, counterexample, and Fraig adapters with their package owners.
2. Port the graph-only utilities and transforms (`intDup.c`, `intFrames.c`,
   `intInter.c`) against the current graph representation, replacing `pData`
   side channels with per-operation edge maps.
3. Port the containment check and counterexample extraction against real
   solver/CNF APIs, preserving SAT/UNSAT/unknown states and cleanup ownership.
4. Port manager lifecycle and the interpolation engine (`intMan.c`,
   `intCheck.c`, `intM114.c`), then add source-C and independent proof
   validation before calling the package complete.
5. Treat `intM114p.c` separately: its implementation is conditional on
   `ABC_USE_LIBRARIES` and depends on MiniSat-1.14p. Do not expose a fake solver
   implementation when that dependency is unavailable.

## Verification status

No package tests or vet checks pass yet because parsing fails before type
checking. The first usable gate is package compilation after the shared API
boundary and all C-style loop/cast/preprocessor remnants are removed. The
release gate is `go test -race ./proof/int` plus `go vet ./proof/int`, with
tests for interpolation soundness, sequential traces, and SAT unknown handling.
