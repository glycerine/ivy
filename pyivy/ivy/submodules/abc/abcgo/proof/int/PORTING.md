# `proof/int` port status

`proof/int` is ABC's McMillan interpolation engine. The Go package is not yet
complete and does not compile as a package: the optional MiniSat 1.14p implementation in `intM114p.go` remains a C-like draft.
Do not treat successful explicit-file tests as package-level verification.

The graph transformations, frame unrolling, counterexample extraction,
containment checks, manager lifecycle, incremental induction checker and
internal BSAT proof/interpolant step have been translated. Focused tests
exercise graph semantics, initial-state and latch behavior, counterexamples,
SAT/UNSAT containment, interpolation proofs, frame unrolling, incremental
induction and restoration of lifted CNF. These focused tests have passed with
the race detector. Focused end-to-end tests also cover a simple safe circuit and a frame-one counterexample; broader validation remains necessary.

The translated top-level algorithm still needs review and end-to-end tests.
It calls the ported `opt/dar` rewrite sequence with a context deadline. The original optional
`intM114p.c` uses MiniSat 1.14p under `ABC_USE_LIBRARIES`; it is not included
in the normal C build and must not be replaced by an unrelated solver while
claiming it is the original implementation.

Completion requires resolving these dependencies, reviewing the top-level
translation against the C source, and passing `go test -race ./proof/int` and
`go vet ./proof/int`, with end-to-end safe, unsafe, and resource-limited cases.
