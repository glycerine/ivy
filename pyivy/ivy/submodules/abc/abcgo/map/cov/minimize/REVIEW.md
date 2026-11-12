# Coverage minimizer review

## Scope

`minimize.go` is a native, value-based exact SOP minimizer intended as the
algorithmic core for replacing the pointer-pool implementation in `map/cov`.
It does not yet wire the result into `Abc_NtkSopEsopCover`; the surrounding
generated `map/cov` package still has parser errors and is not buildable.

## Separate bug and readability review

- Cube matching masks off non-care bits and therefore does not depend on stale
  values in don't-care positions.
- Prime generation starts from the deduplicated union of on-set and don't-care
  minterms; only implicants that cover at least one required on-set point are
  retained as primes.
- Cover search only selects implicants covering an uncovered required minterm.
  It restores both the covered set and selected flags after each recursive
  branch, including when that branch becomes the best cover.
- Results are deterministic: prime order, branch order, and final cube order
  have stable tie breakers.
- Invalid variable counts, out-of-range minterms, oversized on-sets, and an
  excessive prime count return errors instead of overflowing fixed-width masks
  or starting unbounded exact-cover work.

## Verification

This is a normal package test (the subpackage is independently buildable):

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./map/cov/minimize -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./map/cov/minimize
```

Tests exhaustively check all 256 three-variable Boolean functions and cover
don't-care expansion, constants, and input bounds. The parent package command
`go -C abcgo test ./map/cov` remains blocked by generated syntax in
`map/cov/cov.go`, `covBuild.go`, `covCore.go`, `covInt.go`, `covMan.go`,
`covMinEsop.go`, `covMinMan.go`, `covMinSop.go`, `covMinUtil.go`, and
`covTest.go`.
