# DCH Go port

`Dch_DeriveTotalAig` and the DCH SAT-sweeping/choice pipeline are active Go
implementations. No generated draft is hidden by build tags or filename changes.
The original algorithm sources remain in `src/proof/dch`.

## Typed entry points

- `DeriveTotalAig([]*aig.Graph) (*aig.Graph, error)` merges snapshot cones, shares
  CIs, retains secondary dangling implementations and uses the first snapshot's
  ordered outputs, validates sequential metadata, and preserves it.
  `Dch_DeriveTotalAig` follows C's CI/CO-only interface checks and returns
  a fresh manager without copying sequential or constraint metadata; it
  asserts via panic on invalid input.
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

Eight rounds of ABC's process-global multiply-with-carry stream produce 32-bit
simulation signatures, normalized by all-zero-input phase. Classes only split;
these signatures propose candidates and never certify equivalence.

The source sweep builds an incremental working FRAIG: proved nodes are
substituted before later cones are constructed, structural matches need no SAT
query, and unknown nodes block their fanouts. Breadth-first supergate and MUX CNF follows Dch_CnfNodeAddToSolver.
As required by the source sweep, inputs must contain AND/inverter logic with
topologically ordered object IDs; native XOR and buffer nodes are rejected.
Representative storage must not already exist on the input graph; the sweep
creates it even when no equivalences are proved. Polarity adjustment
changes SAT variables consistently. The ported BSAT solver proves both miter
directions; only UNSAT in both directions certifies a pair. Each UNSAT
query adds its proved implication to the incremental solver, and a constant-one
representative needs only the first query. SAT counterexamples
refine classes, while budget exhaustion blocks unproved nodes/fanouts. Solver
recycling follows the source's soft variable/count thresholds. There is no input
truth-table bound or small-circuit production fallback.

Counterexamples resimulate the queried cones using SAT values and complete
other selected cones using the continued source random stream. TFO mode
refines candidate classes reached through fanouts; non-TFO mode refines the
queried class or the source's 500-ID constant-candidate window. Choices replace
proved constant/CI aliases, redirect fanins through representatives, preserve
phase and reject referenced alternatives, redundant support (when enabled),
overlapping chains and combined logic/choice cycles. DFS finalizes choice order.

CI/CO order, trailing register interface, constraints, Name/Spec and AIG-style
terminal metadata are preserved. Registers are free SAT inputs, as appropriate
for combinational equivalence of logic for every state; no reachability/induction
claim is made. Mutable graph operations require ordinary external serialization.

## Validation

`go -C abcgo test ./proof/dch ./aig/aig ./sat/bsat -race -count=1` and DCH vet
pass. Separate compactness and correctness reviews are in REVIEW.md. A genuine
original-C oracle covers associative snapshots and redundant-support filtering,
matching exact total/result AND counts, choice counts and output truth tables.
Normal Go tests read the fixture and need no C toolchain or runtime.

## Source API and validation gaps

The C public DCH entry points in `dch.h` have Go counterparts, but the
internal `dchInt.h` interface is not exported function-for-function. Class
preparation/refinement, SAT/CNF, resimulation, manager lifecycle, and sweeping
are represented by typed Go methods and helpers; callers depending on the C
internal entry points and callbacks do not have a directly equivalent API.
`Dar_ManChoiceNew` is an external DAR declaration, not an implementation in
`src/proof/dch`.

The old DCH SAT path passes the conflict budget to `sat_solver_solve` and
passes zero for its other three budgets (`dchSat.c`). There is no DCH-local
wall-clock timeout in that path. Tests exercise conflict exhaustion as unknown,
unlimited conflicts, blocked fanouts, solver recycling boundaries, and SAT
counterexamples in either query direction. They do not establish behavior for
wall-clock limits supplied by surrounding ABC tools.

The source oracle covers selected small graphs; larger real ABC workloads,
additional solver outcomes and performance characteristics have not been
compared exhaustively against C. Statement coverage does not establish source
fidelity or correctness for all AIGs.

`Dch_DeriveTotalAig` uses Go-local copy maps rather than clearing and updating
the source graphs' C `pData` scratch pointers. `Dch_DeriveChoiceAig` currently
preserves more source metadata than C's fresh manager, and its typed API may
return errors for unsupported metadata clone contracts; this remains a source
contract difference requiring resolution.
