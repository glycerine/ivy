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

Normal `go test ./bdd/extrab` still reports visible unported syntax in
`extraBddAuto.go:190` and later. No files are hidden or skipped to change that
status. Image, auto, symmetry, time, K-map, and remaining misc/cascade-root work
are separate package-frontier tasks.
