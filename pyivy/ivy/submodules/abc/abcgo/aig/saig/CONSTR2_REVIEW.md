# Functional constraint port review

## Scope

The native `saigConstr2` slice includes sequential frame unrolling, AIG-to-CNF
encoding, SAT-based implication discovery, two-frame inductive filtering,
functional constraint unfold/fold graph transforms, source-name wrappers, and
sampled sequential output profiling. The original C-like `saigConstr2.go` body
has been replaced by native wrappers; unsupported pointer/vector APIs are not
retained as faux Go implementations.

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

The broader package still has unrelated C-like drafts that fail parsing and
depend on C-only manager/vector fields. Full `go test ./aig/saig` remains blocked
by those files; this explicit-file gate isolates the functional-constraint
implementation and its wrappers.
