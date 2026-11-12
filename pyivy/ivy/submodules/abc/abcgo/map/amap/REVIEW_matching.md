# Matching review

Scope: `amapMatch.go` and `amapMatch_native_test.go`.

The separate readability pass retains the original match entry points and flow/exact-area selection order, while factoring set-permutation/polarity traversal into one helper. Cuts use owned slices; best cuts are deep copies independent of temporary linked chains. Temporary exact-area references store typed counter pointers instead of C pointer arithmetic. Original verbose-only debugging branches were replaced with a concise per-round report.

The separate bug review checked reference-before-recursion ordering, opposite-polarity inverter accounting, shared logic, temporary counter rollback, the original eight-level recursion bound, epsilon comparisons, and estimated fanout normalization. Regressions exercise both phases of one shared output driver, complemented leaves, repeated references to the same counter, deep bounded/full recovery, fractional delays, independent best-cut storage, flow metrics, and source comparator priorities. Invalid match/permutation state fails explicitly instead of silently producing a mapping.

Two deliberate fixes are recorded: delay maxima use floating-point comparisons (the original `Abc_MaxInt` truncates fractional delays), and complemented fanin wrappers are regularized before accessing node metadata. Native zero-input matches avoid a divide-by-zero fanout average. Original unit-gate delay and eight-level approximate exact-area behavior are preserved; this is not Liberty pin-delay timing.

Full-package gates passed, including the real library/rule/merge/matching sources and both library suites:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -count=1 ./map/amap
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./map/amap
```

No source files are excluded or hidden to run these gates.
