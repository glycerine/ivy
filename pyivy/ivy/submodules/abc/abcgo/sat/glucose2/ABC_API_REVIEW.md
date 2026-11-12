# Glucose2 ABC port review

The C++ wrappers are now native Go paths over the real `SimpSolver` and
`aig.Graph` implementations. The port covers the ABC literal bridge, solver
lifecycle and callbacks, variable elimination/freezing, runtime and conflict
limits, assumption minimization, AND/XOR Tseitin clauses, circuit fanin/round
metadata, DIMACS solving, AIG-to-CNF solving, SOP/cube generation,
quantification, overlap checks, and the `&glucose2` command.

The command registers on `abcmain.Frame`, parses `-C`, `-p`, `-v`, and `-h`,
solves a CNF file or the frame's `pGia` graph, reports SAT status, and stores an
AIG model as the frame counterexample. Parameter fields are exported through
`Glucose2Params`; `Glucose2_Pars` remains an alias for source-facing callers.

## Separate bug and readability pass

- Verified the compact `var*2+sign` bridge over all AND phase combinations and
  the XOR truth table, including complemented inputs.
- Verified all 16 two-input AIG truth tables with preprocessing enabled; every
  SAT result's model evaluates the original graph output to true.
- Verified SAT/UNSAT DIMACS file solving and both registered-command paths
  (CNF file and current AIG resource).
- Verified graph encoding installs ordered AND/XOR fanin metadata so optional
  circuit reasoning sees the correct gate kind while Tseitin clauses retain
  ordinary SAT semantics.
- A regression exposed a subtle incremental bug: calling `SolveLimited` with
  its default simplification could eliminate variables between assumption
  queries. `glucose2_solver_solve` now passes `doSimp=false`, matching the C++
  wrapper's explicit `solveLimited(lits, 0)` and preserving those variables.
- Returned models and conflict data are copied into Go-owned slices. No C
  pointer arithmetic or `unsafe` conversion remains in the wrappers.
- Command usage and parser errors go to the frame's error writer; result text
  goes to the frame's output writer.

## Verification

These are ordinary package-level gates; they compile every Go file in
`sat/glucose2`:

```sh
GOCACHE=/tmp/abcgo-cache go test ./sat/glucose2 -count=1
GOCACHE=/tmp/abcgo-cache go test -race ./sat/glucose2 -count=1
GOCACHE=/tmp/abcgo-cache go vet ./sat/glucose2
```

All three pass. The package has no remaining C-shaped Go wrapper blocker.
