# DAU independent review

## Scope

This was a separate correctness and readability review of the recently
completed DAU implementation slice: `dau.h.go`, `dauInt.go`, `dauCore.go`,
`dau.c.go`, `dauCanon.go`, `dauCount.go`, `truth_counts.go`,
`hierarchical.go`, `adaptive.go`, `dauArray.go`, `dauNonDsd.go`, and the
associated canonicalization leaf additions.

## Correctness review

Compared the array DSD routines with `src/opt/dau/dauArray.c`, the canonical
and hierarchical cache logic with `src/opt/dau/dauCanon.c`, and cofactor/count
operations with `src/opt/dau/dauCount.c`. Reviewed the original-C oracle tests
and focused regressions for the DAU slice. No additional concrete correctness
defects were found.

The implementation's intentional differences from the source are covered by
regressions and/or original-C comparisons: quick-count overflow and recursive
cofactor accumulation, complement propagation when flattening nested XORs,
and the uppercase set parser's correction of C undefined behavior.

The C-shaped API assumes valid buffer sizes, valid dimensions, and
sentinel-terminated DSD arrays. The Go code checks or panics on invalid inputs
in several such cases; these assumptions were not treated as defects because
they match the intended source API contract.

## Readability and compactness review

The new algorithm families use Go slices and maps for owned storage, call-local
scratch state, and small helpers for truth-table operations. C bitfields and
manager storage are represented with explicit Go fields. The logic remains
close to the source's operation order where that order determines phase,
permutation, or decomposition results. No change was recommended during this
pass.

## Verification

Independently run from the repository root:

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test ./opt/dau
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./opt/dau ./opt/dau/canon
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./opt/dau ./opt/dau/canon
```

All three gates passed. This review covers the DAU implementation slice and
does not claim that every package in `abcgo/` is complete.
