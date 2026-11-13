# CEC circuit/SAT core

This is the dependency-neutral Go implementation of the model and incremental
SAT proof group used by ABC's CEC engine. It performs real cone CNF construction,
CDCL solving, equivalence checks and counterexample validation. The parent
`proof/cec` package uses this core for combinational proofs. This is a Go SAT
adapter, rather than a line-by-line port of ABC's CEC SAT implementation.

`Network` exposes GIA-compatible `id<<1|complement` literals with zero false,
ordered input/output interfaces, and constant/input/AND/XOR/MUX/buffer/output
kinds. MUX fanins are control, then, else. Validated snapshots reject dangling or
cyclic fanins, inconsistent interfaces, unknown kinds and unlisted inputs.

```go
result, err := core.VerifyTwo(left, right, core.Options{}) // unlimited conflicts
```

`Equivalent=1`, `Different=0`, and `Undecided=-1` retain the original CEC result
convention. SAT witnesses include ordered Boolean inputs and the failing output
index; each is checked by independent circuit evaluation. `VerifyPairs` accepts
consecutive output pairs; `VerifyMiter` proves all outputs zero. For repeated
local checks, `NewManager` exposes `CheckLiteral` and `CheckPair`. `Model`,
`Model.Values` and `Model.Evaluate` support efficient independent resimulation.

Defaults retain the original SAT/CEC conflict limits and recycling thresholds.
Supplying Options{} explicitly selects unlimited conflicts and no recycling.
Deadlines and context cancellation yield undecided results. Registers require
explicit `TreatRegistersAsInputs`: these are combinational checks, not a claim
of sequential reachability equivalence.

## Separate readability/compactness pass

Immutable typed gate snapshots replace GIA pointer dependencies. Go slices carry
all sizes; there are no separate capacity mirrors. A shared interface-position
index aligns inputs across two models. Cone encoding allocates PI SAT variables
only when used, and aliases unary nodes instead of introducing extra variables.
A fixed three-literal workspace avoids per-gate argument allocations. Witness
resimulation shares one object-value array for same-model comparisons. The
backend is the runnable Go Satoko CDCL implementation; no C runtime is used.

## Separate bug review

Regression findings addressed:

- Constant literal zero is valid; absent cache entries use -1.
- Unknown/deleted objects and cyclic or malformed fanins are rejected before
  proof construction, including dangling structures outside output cones.
- Snapshots and result data do not alias mutable caller graph storage.
- MUX/XOR clauses and phase-flipped SAT literals are exhaustively tested.
- SAT witnesses are independently simulated before a Different result is
  returned; unencoded, irrelevant inputs are assigned false.
- Cancelled/expired queries return undecided instead of continuing to encode.
- Cumulative conflict counts use Satoko's AllConflicts, preserving earlier
  queries and recycled solver history rather than its per-solve counter.
- Recycling rebuilds all SAT-literal caches and shared input bindings.

Tests compare 960
random circuit-node equivalence queries against exhaustive assignment enumeration,
and cross-network queries against independent exhaustive truth tables with
different input-use orders and solver recycling. They also
check every gate's CNF under both phase policies, exercise incremental proofs,
recycling/cancellation and model validation, and compare defaults with original
CEC source when that optional reference tree is available.

## Source-named SAT wrapper boundary

The parent package's source-named node-query wrappers use a separate BSAT
backend with CEC frontier allocation and supergate/MUX clauses. This package's
Satoko manager remains a Go adapter, not the backend for those wrappers.
Source differential coverage currently includes a small AND supergate; it does
not establish matching resource limits, solver statistics, or witness selection
on every network.

`CheckPair`, `VerifyTwo`, `VerifyPairs`, `VerifyMiter`, immutable `Model`
snapshots and evaluation, contexts, deadlines, and mutable conflict limits are
Go convenience APIs, not direct C entry points. `CheckLiteral` and
`CheckConjunction` are also independent adapter operations.
