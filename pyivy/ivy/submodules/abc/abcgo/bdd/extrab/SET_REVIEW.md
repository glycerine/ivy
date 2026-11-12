# Root set-family review

`extraBddSet.go` now forwards the full `extraBddSet.c` API to the authoritative
`extrab/sets` package. The required `Extra_zddMaximal` helper and live internal
entrypoints forward there too. Its previous two draft bodies in
`extraBddMaxMin.go` were replaced with location comments; the remaining MaxMin family is now ported with its own review. No unrelated source family was removed or hidden.

Readability pass: short named forwards preserve source API names; predicate
results use Go bool; shared graph types remain core aliases. No graph algorithm
is duplicated in the root package.

Bug pass: `extraBddSet_test.go` checks every public forwarder, equality/predicate
behavior, and the fused maximal helper. Exhaustive inclusion, empty-family,
ownership, variable-order and original-C differential tests are in the leaf;
its separate review is `sets/REVIEW.md`.

Explicit root gate:

```
GOCACHE=/tmp/abcgo-go-cache go -C abcgo test bdd/extrab/extraBdd.go bdd/extrab/extraBddSet.go bdd/extrab/extraBddSet_test.go -race -count=1
GOCACHE=/tmp/abcgo-go-cache go -C abcgo vet bdd/extrab/extraBdd.go bdd/extrab/extraBddSet.go bdd/extrab/extraBddSet_test.go
```

Full root extrab still has other visible drafts; this is a completed-family gate,
not a claim of full-package success.
