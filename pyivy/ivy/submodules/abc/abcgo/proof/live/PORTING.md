# Native liveness package

The complete `proof/live` package builds using Go and existing native `aig`/`bsat` packages. It implements the source's Biere liveness-to-safety constructions, constrained k-liveness absorber sequence, monotonicity and two-level disjunctive discovery, arena/barrier logic, combinational monitor generation, and Boolean/GF formula processing.

## API and representation

`Network` supplies an `aig.Graph`, CI/CO names (including latch interfaces), and optional LTL properties. `Frame` is the small native command adapter: current network, formula database, writers, and optional safety checker. It replaces dependence on the unfinished C command-frame/network implementation, rather than importing those draft packages. Source command names remain available; slice lengths replace C argc/pointer traversal. The source `nck` command reads its combination width from standard input (or `Frame.In`) and preserves the C command return values.

The main APIs are `LivenessToSafety`, `ParseFormula`, `CheckSafety`, `KLive`, `FindMonotone`, `FindDisjunctiveMonotone`, and streaming `ForEachCombination`. AIG edge zero is true; complemented edge one is false. Graph registers are zero-initialized, with true PIs/POs preceding latch interfaces.

`TransformOptions` selects full, safety-only, liveness-only, or one-step construction, optional shadow-state abstraction, restartable SAVE, and Boolean combinations of GF formulas. One-step mode checks self-loops only. The C LTL command routes one-step mode to the ordinary liveness construction and ignores LTL formulas. Partial shadow sets are abstractions: they may introduce spurious counterexamples. General full-shadow mode records a saved state and event history to detect recurrent counterexamples.

The C-shaped constructors and signal collectors take names from their `Network`
argument and graph objects from their explicit AIG argument, as in the source;
they do not replace or mutate the `Network.Graph` field.

The formula syntax retains source prefix notation (`&`, `+`, `!`, `:G`, `:F`, `:X`, `:U`). Native `>` additionally implements the source enum's implication operation. Parsing recognizes all these temporal operators; circuit compilation explicitly accepts only Boolean combinations of GF with combinational operands, matching the source supported fragment. Boolean names refer to named true POs, with `true`/`false` constants. The direct Go parser reports trailing operands and malformed formulas as errors. The C-shaped LTL command and wrapper skip invalid formulas when at least one valid formula remains, and report an error when none do (C aborts the process in that case).

## Proof backend boundary

The original package orchestrates PDR. The default `CheckSafety` now calls the native `proof/pdr` implementation on a copied graph containing only the requested property and all register transitions. It validates the PDR witness dimensions and initial state, replays it to return input and state traces, checks that the selected output fails, and rejects traces violating AIG constraint outputs. For constrained graphs, an internal sticky invalid-history latch masks paths after the first constraint violation; constraint-free PDR is otherwise unchanged. Constraint outputs are not accepted as property selections. The liveness-to-safety, k-liveness, monotonicity, and combination monitor builders preserve native constraints as trailing outputs without negating them or selecting them as hints. PDR conflict limits and deadlines return unknown when exhausted. PDR has no equivalent state-count limit, so a nonzero `MaxStates` returns unknown with an error.

`SafetyChecker` allows another proof engine to be supplied. PDR runs are not established safe concurrently because its invariant frame is global; callers must serialize them, including against direct `proof/pdr` users. `KLive` only reports a liveness/stabilization proof when a finite absorber depth is proved safe. A counterexample at one finite depth is retained as `LastSafety` and is not mislabeled a liveness disproof. Preprocessing-only discovery similarly does not report a property proof.

Monotonicity discovery retains only proved candidates. Unknown results never become barriers. Disjunctive discovery preserves the source's level-one and level-two heuristic; it does not claim to enumerate every monotone disjunction.

## Source adaptations and fixes

- The native `LivenessToSafety` convenience API guards LTL violations with fairness assumptions. The source-shaped LTL wrapper and command match the C branch, which omits the guard and can report unfair loops.
- The source one-step constructor emits a safety output only when assertions are present; the native convenience API also emits a constant-false safety output when none are present.
- `l2ssim` remains liveness-only, independently of the one-step safety-output behavior. CLI `-s` ignores safety and `-l` ignores liveness; option order does not change their meaning. Default abstraction shadows the first half of original registers, matching source preparation.
- Safety-name aliases `Assert`/`Assume` are retained alongside `assert_safety`/`assume_safety`.
- Actual hint indexes replace assumptions that hints are contiguous or immediately follow the pending output. Known-hint delay lookup uses its real vector position, correcting the source's negative-index arithmetic.
- Each monotonicity obligation is queried explicitly. This avoids relying on C's ambiguous null counterexample entries or accidentally solving unrelated POs.
- The source-shaped monotone-discovery function computes results and then returns NULL, as in C; the separate native `FindMonotone` API returns its proved set.
- Parser state is local and reentrant. Invalid input returns errors instead of exiting the embedding process or reading before the input buffer.
- Combination generation handles `k=0`, validates dimensions, checks int64 overflow, and streams combinations. The C recursive count lacks the zero-width base case.
- Source allocation/free helpers disappear under Go ownership. Graph interface counts update automatically; Biere compatibility counter helpers compute the source count difference without manual counter repair.
- Source debug prints and unconditional dumps are not part of library computation. Explicit printing helpers still write when called. The unfinished, unused fairness-file reader from C is omitted.

Graph construction uses source-compatible copy bookkeeping on its input graph. Concurrent transformations of the same graph require caller synchronization; independent graphs and the local parser do not share scratch state.

## Verification

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test ./proof/live
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./proof/live
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./proof/live
```

Tests include six real original-ABC Biere/one-step/simulation modes with exhaustive eight-frame traces; 200 unchanged-C parser/fragment vectors; and seven actual C kernel fixtures covering three absorber depths, combination monitors, arena barriers, and both monotonicity constructors (448 exhaustive six-frame traces). The C kernel generator repairs only the source's one-byte display-name allocation overflow before executing its unchanged graph algorithms.

A separate 250-case random oracle compares PDR against exhaustive state/input exploration and checks every returned PDR witness. A further 100 random constrained graphs compare PDR with an independent legal-path reachability oracle and check the legality of returned witnesses. Monitor regressions check constraint placement and polarity through each constructor and compare constrained versus unconstrained k-liveness outcomes. Other regressions cover recurrence/fairness, LTL GF combinations, unknown proof results, monotone discovery, k-depth proof boundaries, aliases/CLI flags, abstraction, deadlines, and combination order.

Embedded fixtures and reproducible generators live in `test_vectors/`; ordinary tests require no C compiler. The optional original-ABC executable comparison skips only when that executable is unavailable.

## C function and command inventory

- `liveness.c` and `liveness_sim.c`: Biere, one-step, abstraction, restartable simulation and LTL graph constructors are implemented through `LivenessToSafety`; output collection and command adapters are present. C network strashing is outside this package: `Network` accepts an AIG directly. Native constraint outputs are handled separately from properties.
- `ltl_parser.c`: parser, syntax checks, signal binding, GF sharing and Boolean circuit compilation are implemented. Frame database copying is represented by `Network.LTLProperties` and `Frame.LTLProperties`; C allocator and global-cursor helpers are unnecessary in Go.
- `kliveness.c`, `kLiveConstraints.c`, `arenaViolation.c`: zero-live, absorber, safety-invariant and arena/barrier constructions and PDR orchestration are implemented. The unused C fairness reader and placeholder monotonicity functions have no effective source behavior; allocation/free helpers disappear with Go ownership.
- `monotone.c`, `disjunctiveMonotone.c`: monitor constructors, PDR candidate checks, vector operations and source two-level discovery are implemented. The C monotone function's final NULL return and source index bugs are corrected as described above.
- `combination.c`: combination counts/enumeration and both monitor constructors are implemented. `nck` reads its width from standard input, as the C command does.

The source command flags are `l2s`/abstraction `-1`, `-s`, `-l`, `-h`; `l2ssim` checks only the first option for `-1` and otherwise uses full simulation; and k-liveness and `nck` accept `-c`, `-m`, `-C`, `-g`, `-h`. The LTL command additionally recognizes C's `-f`, which deliberately aborts in C and is rejected by Go. Combined short flags such as `-1s` are accepted. The Go command adapter does not reproduce C's statistics/debug output or strashing side effects.
