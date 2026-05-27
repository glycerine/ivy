# ivy2go BUG_TODO

Scope: current `ivy2go` as of this audit. This list includes bugs confirmed by
source review and sub-agent review. It intentionally avoids broad rewrite
recommendations unless there is a concrete failure mode.

## P1 - generated programs can be wrong, invalid, or unlaunchable

1. Descriptor binary paths do not match the Go output layout.

   Evidence: `WriteOutput` writes Go packages under `outDir/BaseName`
   (`compile.go:579`) and writes `.dsc` files in `outputBaseDirectory(outDir)`
   (`compile.go:600`). `BuildPlanFor` builds the executable as
   `outDir/BaseName/BaseName` (`build.go:74`), but `descriptorJSON` records
   only `Binary: out.BaseName` (`compile.go:381`). A launcher reading the
   descriptor beside the package directory will look for `./BaseName`, not
   `./BaseName/BaseName`.

2. Descriptors can advertise binaries for outputs that cannot be executables.

   Evidence: descriptors are emitted for `target=repl` and `target=test`
   (`compile.go:85`) regardless of package name. `Generate` silently disables
   `EmitMain` whenever the package is not `main` (`generator.go:96`), and
   `BuildOutput` later rejects non-class, non-`main` packages
   (`build.go:119`). So `target=test package=demo` can produce `demo.dsc` with
   a process entry for a binary that the current build path refuses to create.

3. Ivy module parameters are described but not wired into generated Go code.

   Evidence: descriptors include `mod.Params` (`compile.go:381` and
   `compile.go:398`), and `repl.go` has unused helpers for positional/default
   parameters (`repl.go:194`). But `NewState` takes no parameters
   (`state.go:92`), and generated mains call `NewState()` directly for impl,
   repl, and test targets (`generator.go:296`, `generator.go:309`,
   `generator.go:354`). A parameterized Ivy module cannot receive those
   parameter values through the generated Go executable.

4. Solver-backed action generation does not constrain function-sorted state
   consistently with action preconditions.

   Evidence: hash-thunk/map-backed function state is skipped entirely in
   `emitStateSymbolFacts` (`solver_emit.go:151`). Array-backed function state
   is encoded as synthetic const names like `f(0)` (`solver_emit.go:162`),
   while preconditions are reified as real `Apply` nodes (`action_gen.go:960`).
   Those are not the same solver term shape, so test input generation can see
   function state as unconstrained and fire actions that should be disabled.

5. Solver-backed action generation skips important ivy2cpp clause
   normalizations.

   Evidence: `buildActionGenPlan` explicitly says `TrimClauses`,
   field-reference expansion, defined-parameter extraction, relevant
   definitions, and variant axioms are not ported (`action_gen.go:166`), then
   derives preconditions from raw `ReverseImage` (`action_gen.go:217`). Derived
   predicates, record fields, and variants can therefore appear to the solver
   as free symbols instead of the constraints the action body actually relies
   on.

6. Applying a function-valued parameter or local lowers to a state field.

   Evidence: every non-definition `Apply` falls through to
   `goStorageAccess(name, ...)` (`expr.go:260`), and `goStorageAccess` always
   starts from `s.<Name>` (`expr.go:279`). If an action parameter or local has
   function sort and the body calls it, the generated code refers to a
   nonexistent `State` field instead of the local value.

7. Test action generators call multi-return actions with C++ out-parameter
   style, but Go actions return values normally.

   Evidence: `emitSomeAction` emits Go methods with named return values
   (`action.go:1039`). `emitActionGenExecute` handles multiple returns by
   declaring locals and appending them to the call argument list
   (`action_gen.go:589` through `action_gen.go:601`). A public action with two
   returns under `target=test` will generate a wrong-arity method call.

8. Range cardinality ignores the lower bound.

   Evidence: `goSortCard` returns `hi + 1` for `RangeSort` and
   range-interpreted sorts (`types.go:257` and `types.go:266`). The correct
   cardinality is `hi - lo + 1`. This feeds nondeterministic initialization
   (`nondet.go:45`), array dimensions (`types.go:291`), and action input
   generation (`action_gen.go:513`). For `type idx = {5..7}`, generation can
   choose `0..4`; negative ranges can produce incorrect dimensions and loop
   bounds.

9. Range clamping emits statement text where expression text is required.

   Evidence: `rangeClampExpr` returns an `if ... return ...` statement string
   (`expr.go:540`). `emitCastApply` uses that string directly as an expression
   (`expr.go:473`), and `emitRangeArithApply` constructs `return if ...`
   (`expr.go:532`). Casts or arithmetic whose result sort is a range can
   produce syntactically invalid Go.

10. Wide BV types above 64 bits are internally inconsistent.

    Evidence: `goInterpTypeName` maps all BV widths above 64 to `*big.Int`
    (`go_types.go:161`). But `emitBVNumeral` emits `Uint128FromString(...).MaskTo(...)`
    for widths `65..128` (`bv_expr.go:40`). Solver-derived initial values also
    cast numerals with `fmt.Sprintf("%s(%d)", typeName, n)` (`initial_state.go:244`),
    which becomes invalid `*big.Int(1)` for those sorts. Wide-BV assignments
    and initial states can therefore fail to compile.

11. Plain variant leaf constructors are called with arguments even though they
    take none.

    Evidence: `emitVariantSuperStruct` emits `func NewSuperLeaf() Super` for
    plain leaves (`variant.go:208`). `variantUpcastExpr` always renders
    `NewSuperLeaf(expr)` (`variant.go:81`), and `mkNondetVariantScoped` also
    calls every variant constructor with a temp value (`nondet.go:174`). Plain
    leaf upcasts or nondet initialization generate wrong-arity calls.

12. Default state initialization skips function-sorted state instead of using
    the existing function-aware nondet logic.

    Evidence: `emitOneInitialState` skips function-sorted symbols when there is
    no solver model or the symbol is unconstrained (`initial_state.go:113`).
    `mkNondetSym` has separate logic for bounded function arrays and maps
    (`nondet.go:76`), but `emitDefaultInitialState` only calls
    `emitScalarChoice` (`initial_state.go:139`). Bounded function state is left
    at Go zero values rather than Ivy nondet values.

13. Variant-super state can be left as an invalid zero value.

    Evidence: `emitDefaultInitialState` goes through `emitScalarChoice`
    (`initial_state.go:139`), not `mkNondetValueScoped`, whose variant branch
    exists at `nondet.go:138`. A variant super whose first leaf carries a
    payload can remain `Tag == 0` with a nil payload pointer. Later printing or
    downcasting can panic.

14. Clearing an extensional map-backed relation can make later writes panic.

    Evidence: `emitExtensionalRelationClear` emits `s.Rel = nil`
    (`extensional.go:215`). Ordinary map writes use raw map indexing in LHS
    context (`expr.go:318`). In Go, assigning to a nil map panics, so a clear
    followed by `rel(x) := true` can fail at runtime.

15. Numeric enum values from the initial-state solver can become undefined Go
    identifiers.

    Evidence: numeric enums are represented as `int` and get no Go constants
    (`types.go:443`). `modelValueToGo` handles all `LogicEnumeratedSort`
    values by returning `goExportedName(name)` (`initial_state.go:233`). For a
    numeric enum value such as `0`, that becomes `X0`, which is not declared.

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
