# AIG PartSat port review

All 11 original `aigPartSat.c` entry points are implemented. SAT return values
remain 0 (SAT), 1 (UNSAT), and -1 (incomplete). The modern `PartitionedSAT` API
also returns configuration/synthesis errors explicitly.

## Readability and compactness pass

- Native slices and maps replace pointer-vector ranges and macro iteration.
- Partition labels are compacted without allocating up to the largest label.
- `NodeVariables` owns shared SAT numbers and their logical phases; phase is not
  implicitly recovered from mutable graph polarity metadata.
- Direct Tseitin clauses use the existing pure-Go Satoko solver, avoiding a
  cycle through the separate AIG-dependent CNF package. This preserves Boolean
  constraints while differing from the C cut-mapped CNF optimization.
- Synthesis is an injected `PartSATSynthesizer` supplied from a higher layer;
  requesting synthesis without an engine is an error, never an ignored flag.
- No process-global solver, opaque C allocator, or redundant node-count state.

## Independent bug review

Fixed and regression-covered:

- Negative-PI outputs in a zero-node partition must connect the regular PI
  value, then apply the output complement separately.
- Sparse original partition labels must map to compact output-group indices.
- Constant and isolated-PI outputs need constraints even without AND nodes.
- Source constant one must be globally constrained before a boundary CI can
  claim its shared variable.
- Existing client marks must neither become accidental constraints nor be
  destroyed by temporary partition annotations.
- Output objects used as fanins must remain transparent to Boolean splitting.
- Zero conflict limits mean unlimited. Cumulative conflict usage is counted
  using deltas, and a SAT prefix never proves the full problem satisfiable.
- Unknown per-part solves may continue with additional partitions when budget
  remains. Final SAT requires a SAT solve of the complete accumulated CNF.
- A synthesizer that mutates its input cannot evade terminal-count validation.
- Duplicate selected nodes and incomplete mappings are rejected explicitly.

## Verification

Focused real AIG core, Dup, Repr, Scl, PartSat sources and tests pass with the
race detector. Differential checks enumerate input assignments for 24 seeded
random graphs under all four partitioners and both polarity modes, and validate
returned SAT models against `Graph.Evaluate`. Further tests cover degenerate
networks, sparse labels, constant boundaries, nonzero input phases, shared marks,
bounded UNSAT answers, invalid options, and injected/missing/failing synthesis.

The actual DAR synthesis adapter is being supplied by the DAR owner; its end-to-
end wiring remains a higher-level integration step. Full AIG package compilation
still reports unrelated draft syntax in `aigRepar.go`; no files are hidden.
