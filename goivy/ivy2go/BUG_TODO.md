# ivy2go BUG_TODO

Scope: current `ivy2go` as of this audit. This list includes bugs confirmed by
source review and sub-agent review. It intentionally avoids broad rewrite
recommendations unless there is a concrete failure mode.

## Fixed

- Descriptor binary paths now include the Go package directory, matching the
  `WriteOutput`/`BuildPlanFor` layout.
- Descriptor-producing targets now default to `package=main` and reject explicit
  non-main package/class names instead of advertising unbuildable executables.
- Module parameters are now described with defaults and wired through generated
  argv parsing into `NewState`.
- Solver-backed action generation now constrains function-sorted state with the
  same solver term shape used by action preconditions. Array-backed functions
  emit real `Apply` terms per cell, and map/hash-thunk-backed functions emit
  quantified equalities whose RHS combines the thunk fallback with explicit map
  overrides.
- Solver-backed action generation now runs the ivy_to_cpp.py clause
  normalization pipeline: trim reverse-image clauses, expand field references,
  extract destructor-field inputs, compute defined parameters directly, and add
  relevant definitions plus variant axioms before querying the solver.
- Function-valued action parameters and locals now lower through their local Go
  storage rather than through `*State` fields. Map-backed function-valued
  formals/locals use raw map indexes and contribute their tuple key types to
  `types.go`, while true state functions keep the existing getter path.
- Solver-backed test action generators now call Go action methods using normal
  Go return values. Multi-return actions destructure the call result and print
  the first returned value instead of passing return locals as C++-style
  out-parameters.
- Range cardinality now uses the mathematical width `hi - lo + 1`, while
  generated Go arrays keep compact storage by offsetting range-valued indexes
  by the lower bound.
- Range clamping now emits typed expression-shaped Go IIFEs for casts,
  arithmetic, and numerals instead of embedding statement text where an
  expression is required.
- Wide bitvectors now consistently lower widths above 64 bits as masked
  `*big.Int` values. Solver-derived BV initial-state values also parse Z3
  decimal, hex, binary, and `(_ bvN W)` model syntax before emitting width-
  appropriate Go literals.
- Plain variant leaves now use their zero-argument constructors in both
  leaf-to-super upcasts and nondeterministic variant-super construction, while
  payload leaves still pass their payload values.
- Default initial state now uses the same nondeterministic storage initializer
  as locals: bounded function arrays initialize each cell, and variant-super
  state is constructed through a valid leaf constructor instead of staying at
  an invalid zero value. Recursive variant state now terminates by using a
  finite base leaf for recursive fields.
- Extensional hash-thunk relation clears now replace the backing map with a
  fresh map and clear the thunk fallback, so later direct writes cannot panic
  on a nil map.
- Solver-derived initial values for numeric enums now emit numeric literals
  rather than undeclared exported identifiers.

Remaining entries keep their original audit numbering.

## P1 - generated programs can be wrong, invalid, or unlaunchable

## P2 - target and interface behavior diverges from the intended ivy2cpp shape

16. `target=gen` emits the test-loop main.

    Evidence: `emitMain` routes both `test` and `gen` to `emitTestMain`
    (`generator.go:289`). `emitTestMain` runs the randomized test loop and
    prints `test_completed` (`generator.go:346`). ivy2cpp has a separate
    generator path for `gen`, so the current Go target behaves like `test`.

17. Test descriptors advertise parameters the binary ignores.

    Evidence: `test_runs` is parsed into config (`compile.go:139`), and
    descriptors advertise `iters`, `runs`, `seed`, `delay`, `wait`, and
    `modelfile` (`compile.go:388`). The emitted test helper only parses
    `iters` and `seed` (`runtime.go:557`), and `emitTestMain` has only a single
    `iters` loop (`generator.go:408`). `runs`, `delay`, `wait`, and
    `modelfile` are currently descriptor fiction.

18. REPL dispatch exposes non-public actions.

    Evidence: `publicActionNamesSorted` walks `g.Mod.Actions.All()` and filters
    only by prefix (`repl.go:228`). The test action-generator path uses
    `Mod.PublicActions` when available (`action_gen.go:832`). A private helper
    action with an ordinary name can become callable from the Go REPL.

19. REPL silently accepts unsupported structured arguments as zero values.

    Evidence: `emitOneReplArgParser` returns zero value and nil error for
    non-integer, non-bool, non-enum sorts (`repl.go:181`). Actions taking
    struct, variant, or native-typed parameters can be invoked with any token
    and receive a zero value instead of a parse error.

## P3 - narrower invalid-Go emitters

20. `some_min` and `some_max` compare bool index values with `<` or `>`.

    Evidence: `emitIfSomeMinMax` derives `idxType` from the index expression
    (`action.go:730`) and compares `__cur_idx < __best_idx` or `>`
    unconditionally (`action.go:805`). If the minimizing/maximizing index sort
    is bool, this is invalid Go. Existing focused tests exercise a bool-shaped
    `some_min` case.

21. Debug printing assumes loop variables are numeric.

    Evidence: `emitPrintExpr` emits `if X != 0` for every free variable loop
    (`action.go:316`). For bool or enum variables whose Go type is not
    comparable to untyped integer zero, debug-with expressions can generate
    invalid Go.

22. Thunk body rewriting can leave state-function reads as `s.<Field>` inside
    thunk methods.

    Evidence: `emitThunkBody` substitutes loop variables and env symbols with
    fake names, emits the expression, then text-rewrites fake identifiers
    (`thunk.go:250` through `thunk.go:314`). But a captured function symbol
    used as an application is lowered by `emitApply` through `goStorageAccess`
    to `s.<Field>[...]` (`expr.go:279`) before the fake-name text rewrite can
    match it. Unbounded assignments such as `f(X) := g(X)` can emit thunk code
    with an undefined `s` receiver.
