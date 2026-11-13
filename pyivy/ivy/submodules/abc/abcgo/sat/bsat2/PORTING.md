# MiniSAT 2.2 (`bsat2`) port

The complete Go package compiles without excluded files, cgo, unsafe, generated-C fallbacks, or replacement SAT algorithms. Source files remain under their original Go filenames and retain their source licenses.

## Implemented milestones

- Owning generic vectors, circular queues, indexed mutable-priority heaps, MiniSAT quicksort/selection sort, separate-chaining hash maps with original prime growth, and stable-offset region allocation using the source 13/8 capacity sequence.
- Compact literals and three-valued booleans, clause abstraction/subsumption/strengthening, logical word-offset clause references, lazy occurrence deletion, allocation waste accounting, and reference-preserving relocation.
- The actual MiniSAT core: blocker-aware two-watch propagation, phase saving, float activity/rescaling, first-UIP analysis, all three minimization modes, assumption conflict tracing, clause reduction, Luby/geometric restarts, budgets, atomic interruption, simplification, and compacting garbage collection.
- The actual simplifying solver: occurrence counts and elimination heap, touched-clause collection, backward subsumption/self-subsuming resolution, asymmetric branching, implication checks, bounded resolution-based variable elimination, variable substitution, frozen assumptions, and reverse-order eliminated-model reconstruction. Simplification metadata participates in core garbage relocation.
- Plain/gzip DIMACS parsing with line diagnostics, DIMACS emission, typed/range-checked options, SAT/preprocessing command workflows, and the ABC SAT/UNSAT convention wrapper backed by the actual solver. The source `Abc_CallMiniSat22` placeholder now performs solving.

## Public interface and adaptations

`NewSolver()` constructs a core solver; `NewSimpSolver()` constructs the preprocessing solver. Both expose `NewVariable`, `AddClause(...Lit) bool`, `Solve(...Lit) LBool`, `SolveLimited`, `Model`, and `Conflict`. Use `MakeLit(variable, negative)`, `Lit.Not`, `Lit.Var`, and `Lit.Sign`. MiniSAT truth values retain the source convention: `True=0`, `False=1`, `Undef=2`; source's alternate undefined bit pattern is normalized to the same semantic value.

`Solve` disables conflict/propagation budgets as the source bool overload does, but returns `LBool` to distinguish interruption from UNSAT. `SolveLimited` honors budgets; `Interrupt`/`ClearInterrupt` are atomic. Other mutation requires one goroutine per solver. A simplifying solver rejects clauses/assumptions mentioning eliminated variables; freeze variables intended for later incremental use using `SetFrozen`.

`ParseDIMACS`, `WriteDIMACS`, `RunSAT`, and `RunSimp` use readers/writers and returned errors instead of `exit`, raw FILE pointers, and global signal handlers. CLI options include core activity/restart controls and preprocessing controls; unknown platform resource-limit switches are errors, not silently ignored. Process-wide C memory/CPU limit installation and x87 mode changes are not applied by this library. Runtime diagnostics use portable Go runtime-accounted CPU seconds and sampled heap/peak-heap MiB rather than pretending these are OS resident-set measurements. Fixed-width Go types replace the historical compiler-detection macros.

The package uses only the Go standard library. It does not depend on `bsat`, SAToko, or another SAT implementation.

## Validation

`test_vectors/` contains satisfiable, unsatisfiable, and elimination/model-reconstruction DIMACS inputs. They run through plain/gzip parsing, DIMACS round trips, and both command workflows. No `experiments/` tests are run.

- 500 deterministic core formulas are compared with exhaustive enumeration, including assumptions and re-solving without assumptions. Tests vary minimization, phase saving, random decisions, and one-conflict restarts; returned models are checked against original clauses.
- 500 deterministic preprocessing formulas are compared with enumeration, including asymmetric branching, redundancy checks, frozen assumptions, forced garbage collection, and reconstructed models.
- Focused tests cover heap priority changes, vector ownership/reference clearing, allocator relocation, map collision/growth/move behavior, circular queue wrap/growth, original sorting, truth tables, frozen variables, substitution, interruption/budgets, malformed inputs, and option bounds.

## Separate review passes

**Compactness / idiomatic Go / readability:** retained distinct core and preprocessing algorithms, used explicit typed references instead of C overlays, factored propagation and conflict methods by responsibility, retained source algorithm names where useful, and kept ownership transfers explicit. Standard-library readers/writers replace duplicated platform I/O code. Generic containers implement the actual source growth/order algorithms.

**Bug review:** checked source watch/blocker invariants, full reason relocation, occurrence/queue/dummy relocation, minimization seen-bit cleanup, learned-binary retention, assumption isolation, elimination reconstruction polarity/order, queue singleton growth, hash collisions, and three-valued truth semantics. Added a red regression for stale models after an UNSAT preprocessing failure; `SolveSimplified` now clears previous model/conflict results before preprocessing. Source `Substitute` omitted model reconstruction records; this port records the equivalence so substituted original variables receive correct model values. DIMACS headers are validated and truncated clauses rejected instead of accepting malformed input.

Final validation passed: `GOCACHE=/tmp/abcgo-cache go test ./sat/bsat2` (0.058 s), `go test -race ./sat/bsat ./sat/bsat2` (BSAT2 1.237 s), and `go vet ./sat/bsat ./sat/bsat2`. The remaining source-preserving portability comments in `pstdint.go` are license/history text, not retained preprocessor code.

Follow-up review rejects assumptions on eliminated variables even when
preprocessing is disabled. Added exhaustive incremental-query and conflict-core
oracles, post-elimination model checks, and a learned-clause reduction test
covering locked reasons, binary clauses, and garbage collection. These finite
tests do not prove every heuristic or resource-limit interaction.
