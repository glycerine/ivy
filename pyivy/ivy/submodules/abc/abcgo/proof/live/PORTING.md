# Native liveness package

The complete `proof/live` package builds using Go and existing native `aig`/`bsat` packages. It implements the source's Biere liveness-to-safety constructions, constrained k-liveness absorber sequence, monotonicity and two-level disjunctive discovery, arena/barrier logic, combinational monitor generation, and Boolean/GF formula processing.

## API and representation

`Network` supplies an `aig.Graph`, CI/CO names (including latch interfaces), and optional LTL properties. `Frame` is the small native command adapter: current network, formula database, writers, and optional safety checker. It replaces dependence on the unfinished C command-frame/network implementation, rather than importing those draft packages. Source command names remain available; slice lengths replace C argc/pointer traversal.

The main APIs are `LivenessToSafety`, `ParseFormula`, `CheckSafety`, `KLive`, `FindMonotone`, `FindDisjunctiveMonotone`, and streaming `ForEachCombination`. AIG edge zero is true; complemented edge one is false. Graph registers are zero-initialized, with true PIs/POs preceding latch interfaces.

`TransformOptions` selects full, safety-only, liveness-only, or one-step construction, optional shadow-state abstraction, restartable SAVE, and Boolean combinations of GF formulas. One-step mode checks self-loops only. Partial shadow sets are abstractions: they may introduce spurious counterexamples. General full-shadow mode records a saved state and event history to detect recurrent counterexamples.

The formula syntax retains source prefix notation (`&`, `+`, `!`, `:G`, `:F`, `:X`, `:U`). Native `>` additionally implements the source enum's implication operation. Parsing recognizes all these temporal operators; circuit compilation explicitly accepts only Boolean combinations of GF with combinational operands, matching the source supported fragment. Boolean names refer to named true POs, with `true`/`false` constants. Trailing operands and malformed formulas are errors.

## Proof backend boundary

The original package orchestrates PDR. The default `CheckSafety` now calls the native `proof/pdr` implementation on a copied graph containing only the requested property and all register transitions. It validates the PDR witness dimensions and initial state, replays it to return input and state traces, checks that the selected output fails, and rejects traces violating AIG constraint outputs. For constrained graphs, an internal sticky invalid-history latch masks paths after the first constraint violation; the reachable-image backend imposes the same constraints directly on each SAT query. Constraint outputs are not accepted as property selections. The liveness-to-safety, k-liveness, monotonicity, and combination monitor builders preserve native constraints as trailing outputs without negating them or selecting them as hints. PDR conflict limits and deadlines return unknown when exhausted. An explicit `MaxStates` instead selects the existing incremental-CDCL reachable-image checker, since PDR has no equivalent state-count limit; this checker proves safety by exhausting the frontier and reports unknown when the budget is exhausted.

`SafetyChecker` allows another proof engine to be supplied. PDR runs are not established safe concurrently because its invariant frame is global; callers must serialize them, including against direct `proof/pdr` users. `KLive` only reports a liveness/stabilization proof when a finite absorber depth is proved safe. A counterexample at one finite depth is retained as `LastSafety` and is not mislabeled a liveness disproof. Preprocessing-only discovery similarly does not report a property proof.

Monotonicity discovery retains only proved candidates. Unknown results never become barriers. Disjunctive discovery preserves the source's level-one and level-two heuristic; it does not claim to enumerate every monotone disjunction.

## Source adaptations and fixes

- LTL violation outputs respect supplied fairness assumptions. The original C LTL branch omits this guard, unlike its ordinary liveness branch, and can report unfair loops.
- Full one-step mode retains safety assertions. The C branch accidentally tests mode zero while executing mode four and silently omits them. The original-C comparison isolates its shared liveness output for this explicit correction.
- `l2ssim` remains liveness-only, independently of the full-one-step correction. CLI `-s` ignores safety and `-l` ignores liveness; option order does not change their meaning. Default abstraction shadows the first half of original registers, matching source preparation.
- Safety-name aliases `Assert`/`Assume` are retained alongside `assert_safety`/`assume_safety`.
- Actual hint indexes replace assumptions that hints are contiguous or immediately follow the pending output. Known-hint delay lookup uses its real vector position, correcting the source's negative-index arithmetic.
- Each monotonicity obligation is queried explicitly. This avoids relying on C's ambiguous null counterexample entries or accidentally solving unrelated POs.
- The source's monotone-discovery function computes results and then returns NULL; the native API returns its proved set.
- Parser state is local and reentrant. Invalid input returns errors instead of exiting the embedding process or reading before the input buffer.
- Combination generation handles `k=0`, validates dimensions, checks int64 overflow, and streams combinations. The C recursive count lacks the zero-width base case.
- Source allocation/free helpers disappear under Go ownership. Graph interface counts update automatically; Biere compatibility counter helpers compute the source count difference without manual counter repair.
- Source debug prints and unconditional dumps are not part of library computation. Explicit printing helpers still write when called. The unfinished fairness-file reader has a documented native completion: whitespace-separated integer groups, optionally after a colon, one group per line.

Graph construction uses source-compatible copy bookkeeping on its input graph. Concurrent transformations of the same graph require caller synchronization; independent graphs and the local parser do not share scratch state.

## Verification

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test ./proof/live
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./proof/live
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./proof/live
```

Tests include six real original-ABC Biere/one-step/simulation modes with exhaustive eight-frame traces; 200 unchanged-C parser/fragment vectors; and seven actual C kernel fixtures covering three absorber depths, combination monitors, arena barriers, and both monotonicity constructors (448 exhaustive six-frame traces). The C kernel generator repairs only the source's one-byte display-name allocation overflow before executing its unchanged graph algorithms.

A separate 250-case random oracle compares PDR and SAT reachability against exhaustive state/input exploration and checks every returned PDR witness. A further 100 random constrained graphs compare both engines with an independent legal-path reachability oracle and check the legality of returned witnesses. Monitor regressions check constraint placement and polarity through each constructor and compare constrained versus unconstrained k-liveness outcomes. Other regressions cover recurrence/fairness, LTL GF combinations, unknown proof results, monotone discovery, k-depth proof boundaries, aliases/CLI flags, abstraction, deadlines, and combination order.

Embedded fixtures and reproducible generators live in `test_vectors/`; ordinary tests require no C compiler. The optional original-ABC executable comparison skips only when that executable is unavailable.
