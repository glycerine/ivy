# Root MaxMin review

All `extraBddMaxMin.c` APIs are now real forwards to the sole `extrab/sets`
implementation. `Extra_zddMaximal` and its internal forward remain in
`extraBddSet.go`, where the fused set-filter dependency first required them.
No remaining MaxMin drafts or duplicate graph algorithms exist.

Readability review: preserved original API names and retained original source
header; each root wrapper has one purpose. Root MaxMin and Set functions share
the same graph aliases and authoritative engine.

Bug review: all public forwards have a focused regression, including general
non-antichain inputs that expose the original wrappers' missing normalization.
The detailed independent/original-C tests and separate algorithm review are in
`sets/MAXMIN_REVIEW.md`.

Explicit root gate:

```
GOCACHE=/tmp/abcgo-go-cache go -C abcgo test bdd/extrab/extraBdd.go bdd/extrab/extraBddSet.go bdd/extrab/extraBddMaxMin.go bdd/extrab/extraBddMaxMin_test.go -race -count=1
GOCACHE=/tmp/abcgo-go-cache go -C abcgo vet bdd/extrab/extraBdd.go bdd/extrab/extraBddSet.go bdd/extrab/extraBddMaxMin.go bdd/extrab/extraBddMaxMin_test.go
```

All root families are now ported, and normal `go test ./bdd/extrab` passes.
The complete ordinary race/vet gates in `FULL_PACKAGE_REVIEW.md` supersede the
historical explicit-file gate above. No drafts are hidden or skipped.
