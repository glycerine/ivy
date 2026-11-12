# SAIG structural constraint review

## Ported scope

`saigConstr.go` now exposes native `*aig.Graph`/`aig.Edge` adapters for
`Saig_ManDetectConstr`, `Saig_ManDupUnfoldConstrs`, and
`Saig_ManDupFoldConstrs`. `saigConstrNative.go` holds the actual graph
algorithms. They use per-call edge maps and slices; they do not mutate shared
node payload or depend on `Vec_Ptr_t` storage.

The detector follows the original complemented AND-supergate rule. It finds a
latch output factor, compares its complemented next-state supergate against the
property frontier, removes the latch factor, and returns the remaining
property/constraint factors with their source polarity. Unfolding keeps the C
fallback behavior: if an output is unrecognized or outputs disagree on their
constraint set, it returns an ordinary graph copy. Constant-false properties
remain constant false. Folding creates one extra latch whose next state is the
OR of selected failed constraints and gates every property output by the
healthy-state literal.

## Readability pass

- Replaced C vector manager pairs with a compact `StructuralConstraint` value
  holding typed `[]aig.Edge` lists.
- Replaced pointer tagging with the existing complemented `aig.Edge` bit.
- Kept duplication scratch maps local to one graph transform.
- Preserved the original source-facing function names with signatures that use
  the actual Go graph and ordinary integer slices.
- Made constraint selection deterministic by preserving the caller's index
  order and removing duplicate indices before constructing gates.

## Independent bug pass

- Checked positive and negative edge phases while collecting frontier terms and
  while matching the latch input; tests cover the non-trivial complemented
  constraint case.
- Checked that a candidate latch factor is a real LO, that the next-state cone
  is a complemented AND, that its factors are contained in the property, and
  that the latch factor is removed exactly once.
- Checked fallback behavior exhaustively on every assignment for incompatible
  per-output constraints and for outputs outside the recognized shape.
- Checked folding over every assignment to the original inputs and the added
  latch bit. Duplicate constraint indices are harmless and deterministic;
  invalid indices and reserved property output zero are rejected.
- Checked output ordering and register metadata: properties, constraints, old
  next-state outputs, then the added next-state output.

No remaining findings in this structural slice. The unrelated generated
`saigConstr2.go` functional-constraint routines and other SAIG drafts still
prevent whole-package compilation. The focused suite uses only the actual ported
files and does not claim package-wide verification.

## Verification

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race \
  ./aig/saig/saigConstr.go \
  ./aig/saig/saigConstrNative.go \
  ./aig/saig/saigConstrNative_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet \
  ./aig/saig/saigConstr.go \
  ./aig/saig/saigConstrNative.go \
  ./aig/saig/saigConstrNative_test.go
```

Both commands pass. This component has not yet been compared against a live
original-C executable; its regressions check the factorization and sequential
truth behavior independently.
