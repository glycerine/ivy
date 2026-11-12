# Functional constraint port review

## Scope

The native `saigConstr2` slice includes sequential frame unrolling, AIG-to-CNF
encoding, SAT-based implication discovery, two-frame inductive filtering, and
functional constraint unfold/fold graph transforms. The former generated
`saigConstr2.go` is intentionally untouched; it still depends on C-only types
and cannot be the package build gate.

## Separate review findings

- The frame copier maps each next latch output to the prior frame's latch-input
  function. The state-unroll test checks this with a latch whose next state is
  a primary input.
- The CNF encoder handles constant, input, AND, XOR, complemented edge, buffer,
  and output nodes. The test suite exercises complemented AIG edges through the
  transition miter and checks the SAT result.
- Candidate filtering only keeps UNSAT induction checks. SAT and unknown results
  are discarded, matching the source algorithm's conservative behavior.
- Functional detection requires one property output and returns no candidates
  when the property itself is UNSAT. It proves either candidate polarity with
  SAT assumptions before induction filtering.
- Unfolding preserves original outputs and LIs, then appends delayed candidate
  outputs and their pipeline LIs. The regression checks register/output counts.
- Folding ORs the trailing constraint outputs into a sticky failure value and
  gates the property outputs. A truth-table regression checks all input cases.
- Graph snapshots are treated as immutable. Frame maps and SAT state are local
  to each call, avoiding mutation of shared AIG metadata or solver state.

## Verification

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race \
  ./aig/saig/saigConstr.go \
  ./aig/saig/saigConstrNative.go \
  ./aig/saig/saigConstr2Native.go \
  ./aig/saig/saigConstr2Native_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet \
  ./aig/saig/saigConstr.go \
  ./aig/saig/saigConstrNative.go \
  ./aig/saig/saigConstr2Native.go \
  ./aig/saig/saigConstr2Native_test.go
```

## Remaining integration work

The old diagnostic simulation routine and source-spelling wrappers have not
been ported. Full `go test ./aig/saig` remains blocked by unrelated C-like
drafts elsewhere in that package; this explicit-file gate isolates the new
functional-constraint implementation.
