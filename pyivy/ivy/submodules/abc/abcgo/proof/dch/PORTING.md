# DCH Go port

`Dch_DeriveTotalAig` and the DCH SAT-sweeping/choice pipeline are active Go
implementations. No generated draft is hidden by build tags or filename changes.
The original algorithm sources remain in `src/proof/dch`.

## Typed entry points

- `DeriveTotalAig([]*aig.Graph) (*aig.Graph, error)` merges snapshot cones, shares
  CIs, retains secondary dangling implementations and uses the first snapshot's
  ordered outputs. The source-compatible `Dch_DeriveTotalAig` asserts via panic.
- `ComputeEquivalences(*aig.Graph, *Params) error` publishes only proved links in
  the input graph's representative map, with phase differences retained.
- `ComputeChoices(*aig.Graph, *Params) (*aig.Graph, error)` derives an acyclic
  graph with alternatives; `ComputeChoicesWithStats` also returns timing/counts.
- `Engine{Params: ...}.ComputeChoices([]*aig.Graph)` satisfies the structural
  `ChoiceEngine` interface used by DAR/CEC, without importing either package.
- `Params` is also named `Dch_Pars_t`; `DefaultParams` and the original defaults
  setter retain eight words, 1000 conflicts, 5000 soft SAT variables, polarity
  adjustment, TFO simulation, 100 calls before recycling and node lookahead 1000.

`Synthesis`, `Power`, `UseGia`, `UseCSat`, `UseNew`, `LightSynth`, `NodesAhead`
and synthesis duration are retained for surrounding synthesis/choice scripts;
the original old DCH backend does not switch algorithms based on these fields.
A nonnil Params value is explicit; use DefaultParams before changing one field.

## Implementation

Eight rounds of ABC's independent multiply-with-carry stream produce 32-bit
simulation signatures, normalized by all-zero-input phase. Classes only split;
these signatures propose candidates and never certify equivalence.

The source sweep builds an incremental working FRAIG: proved nodes are
substituted before later cones are constructed, structural matches need no SAT
query, and unknown nodes block their fanouts. Lazy Tseitin CNF includes
complemented AND/XOR/buffer logic. Polarity adjustment
changes SAT variables consistently. Incremental pure-Go Satoko proves both miter
directions; only UNSAT in both directions certifies a pair. Each UNSAT
query adds its proved implication to the incremental solver, and a constant-one
representative needs only the first query. SAT counterexamples
refine classes, while budget exhaustion blocks unproved nodes/fanouts. Solver
recycling follows the source's soft variable/count thresholds. There is no input
truth-table bound or small-circuit production fallback.

Counterexamples either run full simulation or update changed TFO values through
a topological priority queue, stopping at controlling values. Choices replace
proved constant/CI aliases, redirect fanins through representatives, preserve
phase and reject referenced alternatives, redundant support (when enabled),
overlapping chains and combined logic/choice cycles. DFS finalizes choice order.

CI/CO order, trailing register interface, constraints, Name/Spec and AIG-style
terminal metadata are preserved. Registers are free SAT inputs, as appropriate
for combinational equivalence of logic for every state; no reachability/induction
claim is made. Mutable graph operations require ordinary external serialization.

## Validation

`go -C abcgo test ./proof/dch ./aig/aig ./sat/satoko -race -count=1` and DCH vet
pass. Separate compactness and correctness reviews are in REVIEW.md. A genuine
original-C oracle covers associative snapshots and redundant-support filtering,
matching exact total/result AND counts, choice counts and output truth tables.
Normal Go tests read the fixture and need no C toolchain or runtime.
