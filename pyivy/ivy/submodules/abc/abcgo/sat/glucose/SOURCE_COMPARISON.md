# Legacy Glucose source comparison

The legacy package is an independent Go implementation and does not import or
alias sat/glucose2. Compared src/sat/glucose with src/sat/glucose2 before adapting
the already-tested common algorithms.

After normalizing include-directory and namespace spelling, Alg, Alloc,
BoundedQueue, Constants, Dimacs, Heap, IntTypes, Map, Options.h/Options.cpp,
pstdint, Queue, Sort, System.h/System.cpp, Vec and XAlloc are identical.
ParseUtils differs only in its include-guard spelling. SimpSolver.cpp differs
only in the Glucose2 addition initializing the parsing flag; SimpSolver.h adds
the circuit feature include. SolverTypes adds circuit declarations/friend access,
with the same literal, LBool, clause and allocator algorithms.

Glucose.cpp / Solver.h have meaningful differences which are retained:

- No circuit justification state, propagation or interpreted gate reasons.
  The source's three circuit hook functions are intentionally empty
  (Solver.h:466-468). The compatibility jftr value remains writable without
  installing Glucose2 behavior.
- No Glucose2 exact conflict-budget check before conflict analysis. Legacy checks
  resource limits after a search call returns. Even an already-zero budget or
  interruption can return a trivial SAT result.
- A SAT model copies every assignment, retaining undefined values for unused
  nondecision variables. Glucose2 instead initializes unvisited model entries
  to true.

The checked-in native_legacy_differences.cpp demonstrates the latter two:
zeroBudget=0, interruptedTrivial=0, nonDecisionModel=2 in raw lbool encoding.
Tests require precisely these results.

The ABC bridge uses the committed aig/aig.Graph boundary, as requested by the
parent integration owner. It implements the original SAT encodings, recursive
assumption minimization, prime-cube enumeration, SAT-based projection,
cofactor-OR quantification and equivalence/overlap queries. Graph's constant one
convention is explicitly encoded by a positive unit clause at variable zero.
Tseitin clauses replace the unavailable GIA/MF eight-input LUT mapper; answers,
models, SOPs and projections are checked semantically. Structural hashing is
provided by Graph, so the old hash flag does not choose a second allocator.
No unfinished GIA, CNF or base/cmd package is imported. Command registration uses
base/main.Frame directly; AIGProvider permits another manager to supply a graph.

Native oracle reproduction from repository root:

```
g++ -O2 -ffunction-sections -fdata-sections -DABC_USE_STDINT_H -I src \
  abcgo/sat/glucose/test_vectors/native_status_oracle.cpp \
  src/sat/glucose/Glucose.cpp src/sat/glucose/Options.cpp \
  src/sat/glucose/System.cpp -Wl,--gc-sections -o /tmp/glucose-native-status
/tmp/glucose-native-status
```

Result:4376511358172924066 over 500 formulas x five assumption requests.
For native_simp_oracle.cpp add src/sat/glucose/SimpSolver.cpp to that command;
result:818442487605194937 over 1,500 preprocessing requests. For
native_legacy_differences.cpp use the first command's three source files.
