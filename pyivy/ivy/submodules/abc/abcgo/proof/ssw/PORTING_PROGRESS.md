# Sequential sweeping port progress

The package is **not yet buildable**: the original draft retains untranslated C
syntax and pointer representations in most source files. No draft files are
excluded from the build. The whole package still needs translation, followed by
correctness and readability reviews and algorithm-level tests.

The original correspondence and rarity parameter defaults have been translated
into `sswParams.go`, including clearing previously set fields. Their focused
tests pass using a temporary stand-in for the still-unported simulation type;
this is not a full-package test.

The source's 32- and 64-bit bit-matrix transpose operations and rarity word
count are in `sswTranspose.go`. Independent bit-level tests check 100 seeded
matrices at each width and compare the optimized 64-bit operation against a
straightforward implementation. Inspection revealed that the source's slow
demonstration transpose reverses output rows relative to its optimized routine;
the Go reference uses the optimized routine's ordering. Focused verification:

```
GOCACHE=/tmp/abcgo-cache go test -race \
  proof/ssw/sswTranspose.go proof/ssw/sswTranspose_test.go
```

The primary port dependency is representation: ABC's `Aig_Obj_t *` and
complemented pointer arithmetic must consistently become the Go AIG package's
graph-owned `aig.Edge` values across frame maps, simulation, class handling,
and SAT construction. Converting individual pointers mechanically will not
preserve complement and ownership semantics.
