# Clock-gating port

`opt/cgt` implements sequential clock gating (not combinational cut refactoring). The original algorithm groups are ported onto native AIG Graph/Edge, slices and the incremental pure-Go Satoko solver. Public entry points retain Cgt names with explicit errors and native arguments. Parameters are exported fields and have the original defaults.

## Design and readability review

Manager-owned frame and partition copy maps replace temporary node payloads. Candidate traversal preserves source postorder and level/useful-node filtering. Partition construction preserves flop-count and node-count batching, adds only care cones whose support is available, and asserts those care outputs before querying candidate AND transition-miter. Successful UNSAT queries add their derived binary implication clause; SAT witnesses are independently resimulated into cached patterns. Conflict-limited unknown results never admit a gate.

ODC mode builds the original two-state future fanout-boundary miter: all other register outputs take their next-state values, the target register is forced to zero/one, and future primary inputs are independent of the current frame. Clock-gate rewriting inserts hold/update MUXes and area mode applies complete-gate substitutions. Name/spec, register/constraint counts, cloned flop tracking, clock domains and cloneable history/timing metadata survive.

Shared graph-clone, fanin/fanout and simulation helpers avoid pointer tagging, source-global scratch and import cycles. The production package imports only native AIG and the SAT leaf implementation. Selection uses deterministic 32-lane/32-frame simulation with local random state; no global RNG is changed.

## Separate bug review

Reviewed AIG-versus-SAT constant polarity, complemented AND/XOR clauses, solver reuse after candidate graph extension, partition-local care support, counterexample filtering, terminal/buffer traversal, future/current input separation, ODC empty boundaries, gate polarity, area substitution and metadata ownership.

A concrete missing-copy failure was reproduced with a red regression: the clock-gate builder's zero map default silently became native AIG constant one. It now checks both source-edge validity and presence of the destination copy before constructing the OR gate.

Additional source issues are handled explicitly:

- Nil parameter pointers consistently use defaults (the C high-level wrapper later dereferenced the original nil pointer).
- Care/input dimension mismatches return an error rather than silently discarding the care graph.
- Partition strides must be positive, preventing nonadvancing partition loops.
- Complete gate selection tracks literal phases. It does not conflate both phases of a regular node or select an incomplete opposite phase for area substitution.
- Empty unobservable fanout boundaries yield false care, supporting dead state rather than hitting the source assertion.
- Every candidate must pass an actual proof; sampling only rejects candidates or chooses among already-proven gates.

## Fidelity and practical limits

Satoko replaces the original MiniSat implementation, so conflict counts and budget-limited accepted sets may differ. The proof query and care/ODC conditions remain exact. Simple selection uses a local seeded Go generator; simulation ranking can differ from SSW's global generator while preserving correctness. The advertised CandidateMax limit is honored. Frame construction retains source copy-map nodes needed by candidate queries; partitions encode their actual queried cones. Public helper arguments adapt original Vec/pointer structures to native slices/maps.

## Validation

- Original-C differential candidate traversal: 100 seeded AND/XOR DAGs, useful masks, level limits and exact candidate postorder.
- Thirty seeded sequential graphs: every accepted gate checked against exhaustive state/input assignments; simple and area rewriting checked for complete transition equality.
- Twenty-five seeded ODC graphs: 60-frame observable trace comparisons with independently evolving state vectors.
- Enable-hold recognition, care-conditioned transition checks, dead-state ODC, nil defaults, partition bounds, SAT witness recording/cleanup, phase-aware area completeness, missing-copy rejection and metadata no-alias checks.

Package gates:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./opt/cgt
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./opt/cgt
```

The optional original-C fixture needs a C compiler and supported ELF linker. Production execution requires no C toolchain.
