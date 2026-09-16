# Codex Divergence Audit: `ivy_check` vs `goivy_check`

This is the working audit for behavioral divergences between the Python checker
rooted at `~/ivy/pyivy/ivy/ivy/ivy_check.py` and the Go checker rooted at
`~/ivy/goivy/cmd/goivy_check` plus the `goivy` package.

The audit was started after the `make gold-hermes` divergence in liveness-to-
safety was fixed and verified. It is intentionally source-review oriented:
xtrace tells us where the two implementations reached the same named point, but
this file records the code-level correspondences and known behavioral seams.

## Status Legend

- `fixed`: confirmed divergence with a landed Go fix and regression coverage.
- `open`: confirmed or strongly evidenced divergence that still needs a Go fix.
- `watch`: reviewed difference that may be intentional or not on the live
  checker path, but should stay visible.
- `mapped`: Python file has an identified Go home, but no confirmed divergence
  is recorded here.
- `outside checker`: not part of `goivy_check` parity unless a specific mode or
  UI path calls it.

## Verification Baseline

After the latest liveness-to-safety fix:

- `make test` passed.
- `make gold-hermes` passed.
- `log.gold-hermes` reached the end of the Hermes run without Go/Python
  divergence.

The new unit coverage for that fix is in `check_l2s_finite_sort_test.go`.

## Audit Method

1. Treat `ivy_check.py` as the root, then inspect the Python import closure.
2. Map each Python module to the Go file family that owns the same behavior.
3. Use xtrace names as correspondence anchors where available.
4. Record only source-reviewed divergences in the divergence ledger.
5. Mark broad or generated/file-family mappings separately from confirmed bugs.

The Python import closure rooted at `ivy_check.py` currently includes 71
top-level Python modules. The complete top-level Python module inventory is 106
files, excluding subdirectories such as `z3/`, `tests/`, `utils/`, and `ivy2/`.

## Entry Point Map

| Python | Go | Status |
| --- | --- | --- |
| `ivy_check.py` parameters such as `diagnose`, `action`, `mc`, `trace`, `separate`, `unchecked_properties`, `prioritize`, `profile` | `ivycheck_params.go`, `module_config.go`, `cmd/goivy_check/main.go` | mapped |
| `ivy_check.start` | `check.go:Start`, `check.go:startLoaded`, `cmd/goivy_check/main.go` | mapped |
| `ivy_check.check_module` | `check_isolate_check.go:CheckModule` | mapped |
| `ivy_check.check_isolate` | `check_isolate_check.go:CheckIsolate` | mapped |
| `ivy_check.check_subgoals` | `check_isolate_check.go:CheckSubgoals` | mapped |
| `ivy_check.check_fcs_in_state` | `check.go:CheckFcsInStateWithAG`, `check.go:checkFcsTracePath`, `check.go:checkFcsNormalPath` | mapped |
| `ivy_check.check_conjs_in_state` | `check_isolate_check.go:CheckConjsInStateWithAG` on the live path | mapped |
| `ivy_check.check_safety_in_state` | `check_isolate_check.go:CheckSafetyInStateWithAG` on the live path | mapped |
| `ivy_check.check_temporals` | `check.go:CheckTemporals` | mapped |
| `ivy_check.apply_conj_proofs` | `check.go:ApplyConjProofs` | fixed |
| `ivy_check.preprocess_assumed_ignored_properties` | `check.go:PreprocessAssumedIgnoredProperties`, `acl.go` | mapped |
| `ivy_check.mc_tactic`, `vmt_tactic` | `check.go:MCTactic`, `check.go:VMTTactic` | mapped |
| `ivy_check.gui_art` | `check_phase7.go:GuiArt`, webui hook | watch |

## Confirmed Divergences And Fixes

### 1. L2S finite-sort lookup used display strings, not declared sort names

Status: fixed.

Python behavior:

- `ivy_l2s.py:374-381`, `ivy_l2s.py:971-1001`, and
  `ivy_ranking.py:315-316` test liveness variables with
  `var.sort.name not in finite_sorts`.
- For an enumerated sort declared as `type ltask = {ready_finish,o3_finish}`,
  Python uses the declared name `ltask`.

Former Go behavior:

- `check_l2s_auto.go`, `check_l2s.go`, and `check_ranking_tactic.go` used
  `sort.String()` or equivalent display text as the finite-sort key.
- For enumerated sorts this can be `{ready_finish,o3_finish}`, while
  `finiteSorts` is keyed by `ltask`.
- Go therefore emitted spurious `l2s_d(T)` or `l2s_a(T)` guards for finite enum
  task variables. The `gold-hermes` divergence showed this at
  `l2s_needed_when_start`, where Python had an empty `And` consequent and Go
  had `And(l2s_d(T))`.

Fix:

- Added `l2sSortIsFinite` in `check_l2s.go:138`; it checks `SortName(sort)`.
- Replaced the guard generation sites in `check_l2s.go`, `check_l2s_auto.go`,
  and `check_ranking_tactic.go`.
- Added regression coverage:
  - `TestL2SFiniteSortLookupUsesDeclaredEnumName`
  - `TestL2SAuto5EnumWorkNeededDoesNotRequireDOrA`
  - `TestRankingEnumWorkCreatedDoesNotRequireD`

### 2. L2S trace hooks mutated fields that Go did not render

Status: fixed. Source details are also recorded in `audit.md`.

Python behavior:

- `ivy_l2s.py` trace hooks set `hidden_symbols`, `renaming`, `pp`, and loop
  markers on a trace object.
- `ivy_trace.py` consumes those fields while rendering the trace.

Former Go behavior:

- L2S hooks wrote analogous fields onto a `MatchHandler`, but the rendered
  output did not consume them.

Fix:

- Go `Trace` now implements the annotation-handler path used by trace
  reconstruction.
- Trace rendering consumes hidden-symbol predicates, pretty-printers, structural
  renaming, loop markers, detailed/non-detailed mode, and model-derived state
  equations.
- Regression coverage lives in the trace and L2S hook tests named in
  `audit.md`.

### 3. Method subgoal trace hooks discarded the transformed failure object

Status: fixed.

Python behavior:

- In `ivy_check.check_subgoals`, method failures run
  `foo = goal.trace_hook(foo)` and then print or display the returned object.

Former Go behavior:

- Go invoked a trace hook on an empty throwaway handler and continued to report
  the original error object.

Fix:

- Added `TraceFailure` and routed method-failure objects through the actual
  hook result.
- `CheckSubgoals` now uses the transformed trace for `OptTrace` and GUI
  diagnostics.

### 4. `CheckFinalCond` displayed input clauses instead of model valuations

Status: fixed.

Python behavior:

- `ivy_trace.check_vc` obtains a model, builds model clauses, creates a
  model-backed `Trace`, replays action annotations into that trace, and displays
  model-derived valuations.

Former Go behavior:

- The Go CTI/final-condition path constructed graph states from pre/post input
  clauses instead of from the satisfying model.

Fix:

- `CheckVC` now constructs a model-backed `Trace`, replays `MatchAnnotation`,
  and renders states through the same valuation path.
- Relation-minimization objectives are also forwarded to the solver.

### 5. L2S trace renaming used text replacement

Status: fixed.

Python behavior:

- `ivy_l2s.renaming_hook` installs a structural symbol renaming.
- `ivy_logic_utils.rename_ast` renames AST nodes, so overlapping names such as
  `l2s_s_1` and `l2s_s_10` are safe.

Former Go behavior:

- The hook used `strings.ReplaceAll` over already-rendered lines, with Go map
  iteration order controlling replacement order.

Fix:

- Go trace renaming is structural and render-time, not text replacement.
- Regression coverage includes overlapping nonce names.

### 6. Solver state for failing final conditions was popped too early

Status: fixed.

Python behavior:

- `ivy_solver.get_small_model` leaves the failing final condition asserted when
  it breaks out of the checker loop on a SAT failure. The returned model
  satisfies the failed condition.

Former Go behavior:

- `GetSmallModelWithCond` popped the final-condition frame before returning a
  model for the failure.

Fix:

- Go now preserves the failing final condition for model extraction and
  shrinking.

### 7. L2S loop-start hook lost the saved-state index

Status: fixed.

Python behavior:

- `ivy_l2s.trace_hook` scans trace states and marks the predecessor of the state
  containing `l2s_saved = true` as the loop start.

Former Go behavior:

- Go scanned a symbol-indexed equation map and always recorded loop start index
  0.

Fix:

- Go scans ordered trace states and marks the same predecessor state Python
  would mark.

### 8. Trace cloning, hidden-symbol defaults, and model universes diverged

Status: fixed.

Python behavior:

- `ivy_trace.Trace.clone()` preserves model and vocabulary.
- Trace states receive the model universe.
- Returned `Trace` objects have a default false hidden-symbol predicate.

Former Go behavior:

- Subtraces could lose model/vocabulary state, states lacked universe data, and
  hidden-symbol defaults could be nil.

Fix:

- Go trace cloning and state construction now preserve model-backed fields and
  defaults.

### 9. Earlier checker-path regressions covered by `check_regression_test.go`

Status: fixed.

These were already fixed before this audit file was created, but they are part
of the `ivy_check.py` vs `goivy_check` divergence record:

- `IsCheckModUnprovable` now matches Python's
  `lf.unprovable == act.check_unprovable.get()` filter.
- The live conjecture/safety checks use AG-aware functions so history, post
  state, and update context are available like Python's `check_conjs_in_state`
  and `check_safety_in_state`.
- `ConvertPostcondsWithUpdate` receives `post.Update` and can rename old-state
  symbols through the update triple, matching `convert_postconds(state, pcs)`.
- Property promotion is followed by `UpdateTheory`, matching
  `im.module.update_theory()`.
- `NewBaseChecker` dualization uses Python's `@` witness prefix rather than
  Go's ordinary double-underscore skolem naming.
- `OnlyCheckUnprovable` guards prevent normal property/init-invariant checks
  in unprovable-only mode.

Coverage:

- `check_regression_test.go` contains the named regression groups.

## Open Or Watch Items

### A. `ApplyConjProofs` can silently pass through proof failures

Status: fixed.

Python behavior:

- `ivy_check.apply_conj_proofs` calls `pc.admit_proposition(lf, proof)` and then
  maps subgoals through `ivy_compiler.theorem_to_property`.
- There is no local try/except fallback; proof application errors escape.

Current Go behavior:

- `check.go:ApplyConjProofs` now returns an error.
- A conjecture with no proof still passes through unchanged, matching Python.
- A conjecture with a proof must either produce converted subgoals or return the
  proof/conversion error.
- Empty subgoal lists remain empty, matching Python's `conjs.extend(subgoals)`.

Former risk:

- A malformed or unsupported proof can be silently ignored by Go where Python
  would stop with an error.

Fix:

- Changed `ApplyConjProofs` to return `error`.
- Changed `CheckIsolate` to propagate that error.
- Removed the fallback-to-original-conjecture path for proof-present failures.
- Added `TestApplyConjProofsPropagatesProofErrorLikePython`.

TDD log:

- Red: the new test initially failed at compile time because `ApplyConjProofs`
  returned no error.
- Green: focused `go test -run 'TestApplyConjProofs(PropagatesProofErrorLikePython)?$' -count=1` passed after the fix.
- Regression gate: `make test` passed after the fix.

### B. CLI `GuiArt` is a hook/stub, not Python's Tk main loop

Status: watch.

Python behavior:

- `ivy_check.gui_art` constructs a Tk UI, adds an analysis graph or trace, runs
  the Tk main loop, and exits.

Current Go behavior:

- `check_phase7.go:GuiArt` prepares comparable data and delegates to
  `mod.Cfg.GuiArtHook` when one is registered.
- With no hook, CLI mode prints a diagnostic summary and returns.

Risk:

- `diagnose=true` without a webui hook is not interactive like Python.

Proposed fix:

- Treat this as accepted for CLI-only `goivy_check`, but keep `--trace` and
  webui hooks as the supported diagnostic paths.
- If full Python diagnostic parity is required, make `diagnose=true` fail with
  an explicit message when no hook is installed instead of pretending a GUI was
  launched.

### C. Parameter parsing accepts multiple `=` characters

Status: fixed.

Python behavior:

- `ivy_init.read_params` uses `str.split(arg, '=')` and calls `usage()` if the
  split has more than two fields.
- Therefore `diagnose=true=false` is rejected before checking starts.

Former Go behavior:

- `compiler_ivyinit.go:ReadParams` uses `strings.SplitN(arg, "=", 2)`.
- `cmd/goivy_check/main.go` has an independent `SplitN` loop.
- Therefore `diagnose=true=false` is accepted with value `true=false`.

Risk:

- Bad command-line parameters that Python rejects can silently reach Go's
  checker with different parameter values or later unknown-value behavior.

Fix:

- Add a shared parser that rejects more than one `=`.
- Use it in `cmd/goivy_check`.
- Make `ReadParams` reject the same shape.
- Add red/green tests for both the shared goivy_check parser and `ReadParams`.

TDD log:

- Red: `TestParseIvyCheckParamsRejectsMultipleEqualsLikePython` initially
  failed at compile time because the shared helper did not exist; the same test
  also captured the intended Python rejection rule.
- Green: focused `go test -run 'Test(ParseIvyCheckParams|ReadParams)RejectsMultipleEqualsLikePython' -count=1` passed after the fix.
- Regression gate: `make test` passed after the fix.

### D. `GuiArt` fixes an upstream Python typo in `default_ui == "art"` mode

Status: watch.

Python behavior:

- In `ivy_check.gui_art`, the `"art"` branch assigns `other_art` but later calls
  `ag.execute(...)`; `ag` is not defined in that function scope.

Current Go behavior:

- `check_phase7.go:GuiArt` uses `otherArt.Execute(...)`, making the branch run.

Risk:

- This is intentionally more usable than Python but not byte-for-byte faithful
  for the rare `"art"` UI mode.

Proposed fix:

- Keep the Go behavior and document it as an intentional upstream-bug repair, or
  gate it behind a compatibility flag if an xtrace comparison ever depends on
  the Python exception.

### E. Convenience wrappers without analysis graph are not Python live paths

Status: watch.

Python behavior:

- `check_conjs_in_state` and `check_safety_in_state` always receive an analysis
  graph and state on the live checker path.

Current Go behavior:

- `CheckConjsInState` and `CheckSafetyInState` exist as convenience wrappers
  and can fall back to direct solver checks without an analysis graph.
- Production isolate checking uses `CheckConjsInStateWithAG` and
  `CheckSafetyInStateWithAG`, which are the Python-equivalent paths.

Risk:

- Tests or future callers may accidentally use the convenience wrapper and miss
  post-state update context, especially for postcondition old-symbol renaming.

Proposed fix:

- Keep wrappers only for tests and simple direct checks.
- Prefer naming or comments that make the AG-aware versions the default for
  checker pipeline work.

### F. Ivy 1.7+ parser ignores top-level `using` imports

Status: fixed.

Python behavior:

- `ivy_parser.p_top_using_symbol` imports the named module for every supported
  parser version, substitutes the module name as a prefix on each imported
  declaration, and declares those prefixed declarations in the current parser
  accumulator.

Former Go behavior:

- The Ivy <=1.6 grammar action calls the shared `parserDeclareUsing` helper.
- The Ivy 1.7+ grammar action only traces `parser.p_top_using_symbol` and does
  not call the importer or declare prefixed declarations.

Risk:

- Ivy 1.7+ files using `using foo` can parse successfully while missing every
  declaration from `foo`, diverging from Python before checking begins.

Fix:

- Added `TestParseV17UsingImportsWithPrefixLikePython`, matching the existing
  v1.6 `using` regression but on the Ivy 1.7 grammar.
- Changed the Ivy 1.7+ grammar action and checked-in generated parser to call
  the shared `parserDeclareUsing` helper.

TDD log:

- Red: the new v1.7 test failed because the importer call list was empty.
- Green: focused `go test -run TestParseV17UsingImportsWithPrefixLikePython -count=1` passed after the fix.
- Focused regression: `go test -run 'TestParseV1[67]UsingImportsWithPrefix' -count=1` passed.
- Regression gate: `make test` passed after the fix.

### G. Lexer accepts malformed uppercase-variable subscripts as variables

Status: fixed.

Python behavior:

- `ivy_lexer.t_VARIABLE` uses the regex
  `[A-Z][_a-zA-Z0-9]*(\[[ab-zA-Z_0-9]*\])*`.
- A bracket suffix is part of a variable token only when it is closed and its
  contents are identifier characters.

Former Go behavior:

- `lexer.go:scanVariable` consumes from `[` through the next `]`, and if no
  `]` exists it consumes through EOF.
- It also accepts punctuation inside the bracket group.

Risk:

- Invalid inputs such as `X[abc` or `X[!]` become single `VARIABLE` tokens in
  Go where Python would tokenize `X` followed by `[` and then either more
  syntax or an illegal character.

Fix:

- Added `TestVariableMalformedSubscriptStopsLikePython`.
- Made `scanVariable` accept only closed Python-shaped bracket suffixes while
  leaving malformed brackets to be tokenized normally.

TDD log:

- Red: the new test failed because `X[abc Y[!]` was emitted as one `VARIABLE`
  token.
- Green: focused `go test -run TestVariableMalformedSubscriptStopsLikePython -count=1` passed after the fix.
- Focused regression: `go test -run 'Test(Variable|Keywords|Version|MultiChar|Iff|Arrow|Dot)' -count=1` passed.
- Regression gate: `make test` passed after the fix.

### H. Lexer accepts Unicode identifier characters that Python rejects

Status: fixed.

Python behavior:

- `ivy_lexer.t_PRESYMBOL` and `ivy_lexer.t_VARIABLE` use explicit ASCII
  identifier ranges: `[_a-z0-9][_a-zA-Z0-9]*` and
  `[A-Z][_a-zA-Z0-9]*...`.
- Non-ASCII letters outside quoted symbols are illegal characters.

Former Go behavior:

- `lexer.go` uses `unicode.IsLower`, `unicode.IsUpper`, `unicode.IsLetter`,
  and `unicode.IsDigit`, so inputs such as `é` or `É` become identifiers.

Risk:

- Go can accept Ivy source with non-ASCII identifiers that Python rejects before
  parsing, shifting failures later or allowing non-portable specs.

Fix:

- Added `TestNonASCIIIdentifierCharactersRejectedLikePython`.
- Replaced Unicode identifier checks with explicit ASCII helpers for unquoted
  symbols and variables.

TDD log:

- Red: the new test failed because `héllo` was emitted as one `SYMBOL` and
  `Éclair` as one `VARIABLE`.
- Green: focused `go test -run TestNonASCIIIdentifierCharactersRejectedLikePython -count=1` passed after the fix.
- Focused regression: `go test -run 'Test(Symbol|Variable|Underscore|NonASCII|Keywords|Version|MultiChar|Iff|Arrow|Dot)' -count=1` passed.
- Regression gate: `make test` passed after the fix.

### I. Parser suppresses `include`/`using` importer errors

Status: fixed.

Python behavior:

- `ivy_parser.p_top_include_symbol` and `ivy_parser.p_top_using_symbol` call
  `importer(name)` directly.
- If the importer raises an `IvyError` for a missing or invalid module, parsing
  aborts.

Former Go behavior:

- `parserDeclareInclude` and `parserDeclareUsing` trace importer errors but do
  not add a parse error or return the error to the parser.
- A source file can therefore parse successfully with missing included/imported
  declarations.

Risk:

- `include missing` or `using missing` can silently drop dependencies in Go
  where Python stops before checking.

Fix:

- Added `TestParserIncludePropagatesImporterErrorLikePython`.
- Added `TestParserUsingPropagatesImporterErrorLikePython`.
- Converted importer errors into parser errors on the active accumulator so
  `Parse` returns a non-nil error.

TDD log:

- Red: both new tests failed because `Parse` returned success after tracing the
  importer error.
- Green: focused `go test -run 'TestParser(Include|Using)PropagatesImporterErrorLikePython' -count=1` passed after the fix.
- Focused regression: `go test -run 'TestParser(NestedIncludeSeesParentIncludedStack|IncludePropagatesImporterErrorLikePython|UsingPropagatesImporterErrorLikePython)|TestParseV1[67]UsingImportsWithPrefix' -count=1` passed.
- Regression gate: `make test` passed after the fix.

### J. `TheoremToProperty` panics on undisclosed definitional subgoals

Status: fixed.

Python behavior:

- `ivy_compiler.theorem_to_property` raises `IvyError(prop, "definitional
  subgoal must be discharged")` when a schema conclusion is a definition.
- Callers on the checker path surface that as a normal checker error.

Current Go behavior:

- `compiler_ivy_compile.go:TheoremToProperty` panics with the same message.
- Checker callsites that convert proof subgoals can therefore crash instead of
  returning an error like Python.

Risk:

- A proof that leaves a definitional subgoal undisclosed can terminate
  `goivy_check` with a panic rather than a controlled diagnostic.

Fix:

- Added `TheoremToPropertyChecked`, which returns a normal error for
  undisclosed definitional subgoals.
- Kept `TheoremToProperty` as a compatibility wrapper that preserves the
  existing panic behavior for direct callers.
- Routed checker/proof-subgoal conversion paths through the checked helper:
  `ApplyConjProofs`, non-temporal isolate subgoal checking, and
  `CheckProperties` subgoal conversion now surface the Python-style error.

TDD log:

- Red: `TestTheoremToPropertyCheckedDefinitionConclusionReturnsErrorLikePython`
  initially failed to compile because `TheoremToPropertyChecked` did not exist.
- Green: focused `go test -run TestTheoremToPropertyCheckedDefinitionConclusionReturnsErrorLikePython -count=1` passed after the helper and callsite changes.
- Focused regression: `go test -run 'TestTheoremToProperty|TestApplyConjProofs|TestCheckProperties' -count=1` passed.
- Regression gate: `make test` passed after the fix.

### K. Assert-action proof failures are converted into assumptions

Status: fixed.

Python behavior:

- `ivy_compiler.apply_assert_proof` calls `prover.get_subgoals(goal, pf)`.
- It does not catch proof errors; failures escape from
  `apply_assert_proofs`/`check_properties`.

Former Go behavior:

- `compiler_phase6.go:applyAssertProofActionWithProof` catches
  `GetSubgoals` errors and returns an `AssumeAction`.
- It also catches theorem-to-property conversion errors after proof subgoal
  generation and returns an `AssumeAction`.

Risk:

- A malformed assertion proof can become an assumption in Go, making the
  checked action weaker where Python would reject the proof.

Fix:

- `applyAssertProofActionWithProof` now returns `(ActionsAction, error)` and
  propagates `GetSubgoals` and theorem-to-property conversion errors.
- `ApplyAssertProofsWithProver` threads that error through the recursive action
  rewrite and returns it to `CheckProperties`.
- `ApplyAssertProofWith` now returns an error too; the L2S generated-proof path
  propagates it from the enclosing tactic.
- No-proof and non-verifying behavior is unchanged.

TDD log:

- Red: `TestApplyAssertProofsWithProverPropagatesProofErrorLikePython`
  initially failed because `ApplyAssertProofsWithProver` returned nil after the
  mock prover failed.
- Green: focused `go test -run TestApplyAssertProofsWithProverPropagatesProofErrorLikePython -count=1` passed after the fix.
- Focused regression:
  `go test -run 'TestApplyAssertProofsWithProver|TestCompileAssertFormula_WithProof' -count=1`
  and `go test -run 'TestL2S|TestRanking' -count=1` passed.
- Regression gate: `make test` passed after the fix.

### L. Named-property specialization looks for `forall` instead of existential bodies

Status: fixed.

Python behavior:

- `ivy_compiler.check_properties.named_trans` first calls
  `ivy_logic.drop_universals(prop.formula)`.
- Valid named declarations are existential facts. After `drop_universals`,
  Python reads the existential's `variables` and `body`, substitutes the first
  existential variable with the named term, and drops any remaining leading
  universals.

Former Go behavior:

- `compiler_phase6.go:namedTrans` only specializes when `prop.Formula` is
  directly a `*ForAll`.
- For the normal named-fact shape `And(exists X. p(X))`, Go leaves the named
  copy unchanged instead of adding Python's `p(a)` specialized property.

Risk:

- Named facts can fail to produce the same specialized property that Python
  adds to `mod.labeled_props` and `mod.subgoals`, changing later proof context.

Fix:

- `namedTrans` now casts to logic `Expr`, applies `IvyDropUniversals`, and
  specializes the first variable of the resulting `LogicExists` body.
- Non-logic and non-existential shapes still pass through unchanged.

TDD log:

- Red: the first focused test showed the old behavior left a single-element
  wrapper as `LogicAnd`. A Python oracle check then refined the valid named
  fact shape to `And(exists X. p(X))`, for which Python produces `p(a)`.
- Green: focused `go test -run TestCheckPropertiesNamedTransDropsOneElementAndLikePython -count=1` passed after the fix.
- Focused regression:
  `go test -run 'TestCheckProperties|TestTheoremToProperty|TestApplyAssertProofsWithProver' -count=1`
  passed.
- Regression gate: `make test` passed after the fix.

### M. `DomainSetup.named` ignores free-parameter validation and stores bare symbols

Status: fixed.

Python behavior:

- `ivy_compiler.IvyDomainSetup.named` drops leading universals from the last
  fact and requires an existential with one witness variable.
- It builds `vmap` from free variables in the existential condition.
- Each named-declaration argument is checked against `vmap`; missing free
  variables raise "`<var> must be a parameter of <name>`".
- When parameters are present, `mod.named` stores `sym(*targs)`, not the bare
  symbol.

Former Go behavior:

- `compiler_decl.go:DomainSetup.Named` compiles the left-hand parameters only
  for their sorts.
- It does not reject missing free parameters.
- It always appends `NamedEntry{Name: sym}` even when Python would append
  `sym(*targs)`.

Risk:

- Named existential witnesses lose their dependency on free variables and can
  be treated as constants, changing both named-property specialization and
  generated named updates.

Fix:

- Ported Python's `vmap`/`used`/`targs` logic into `DomainSetup.Named`.
- Free variables in the existential condition must now be present as named
  parameters.
- `mod.Named` now stores `sym(*targs)` when parameters are present.

TDD log:

- Red:
  `go test -run 'TestDomainSetupNamed(StoresAppliedSymbolForParametersLikePython|RejectsMissingFreeParameterLikePython)' -count=1`
  failed because Go stored a bare `*Const` and accepted a missing free
  parameter.
- Green: the same focused command passed after the fix.
- Focused regression:
  `go test -run 'TestDomainSetupNamed|TestDeclInterp|TestCompile.*Named|TestCheckPropertiesNamed|TestCheckDefinitions_Named' -count=1`
  passed.
- Regression gate: `make test` passed after the fix.

### N. `CheckDefinitions` skips applied named entries

Status: fixed.

Python behavior:

- `ivy_compiler.check_definitions` iterates `for ldf, term in mod.named` and
  calls `checkdef(term.rep, ldf)`.
- For an applied named term `f(Y)`, Python `Apply.rep` is the function symbol
  `f`, so the named witness reserves/checks `f` for redefinitions and
  interpreted-symbol errors.

Former Go behavior:

- `compiler_ivy_compile.go:CheckDefinitions` only checks `mod.Named` entries
  whose `Name` is directly a `*Const`.
- After item M, parameterized named declarations store `*Apply`, so their
  underlying symbol is skipped by definition checking.

Risk:

- A named witness function can collide with a definition or interpreted symbol
  without Go reporting the Python error.

Fix:

- Added `namedEntryRepConst` so `CheckDefinitions` checks both bare `Const`
  named entries and applied named entries by their underlying function symbol.

TDD log:

- Red:
  `go test -run TestCheckDefinitions_AppliedNamedRedefinitionErrorLikePython -count=1`
  failed because Go returned nil for a colliding `f(Y)` named entry.
- Green: the same focused command passed after the fix.
- Focused regression:
  `go test -run 'TestCheckDefinitions|TestDomainSetupNamed|TestCheckPropertiesNamed' -count=1`
  passed.
- Regression gate: `make test` passed after the fix.

### O. `prioritize=` is treated as unset instead of explicitly empty

Status: fixed.

Python behavior:

- `ivy_check.get_prioritized_actions` reads
  `priority_actions = iu.Parameter("prioritize", None)`.
- If the parameter is absent, `priority_actions.get()` is `None` and Python
  returns an empty list.
- If the parameter is explicitly present as `prioritize=`, Python sees the empty
  string, splits it, prefixes the sole element, and returns `["ext:"]`.

Former Go behavior:

- `Config.PriorityActions` is a plain string, so the absent value and explicit
  empty value both become `""`.
- `GetPrioritizedActions` returns nil whenever `PriorityActions == ""`.

Risk:

- Go cannot distinguish an omitted `prioritize` parameter from Python's
  explicitly empty parameter. This can change checker ordering diagnostics and
  any code that tests whether the option was present.

Fix:

- Add a presence bit for `prioritize`, set it from `ApplyIvyCheckParams`, and
  make `GetPrioritizedActions` return `["ext:"]` for explicit `prioritize=`
  while keeping the absent case nil.
- Use the same presence bit for the isolate checker's prioritized-order
  diagnostic, matching Python's `priority_actions.get() != None`.

TDD log:

- Red:
  `go test -run 'Test(GetPrioritizedActionsExplicitEmptyLikePython|ApplyIvyCheckParamsMarksEmptyPrioritizePresentLikePython)' -count=1`
  initially failed at compile time because `Config.PriorityActionsSet` did not
  exist.
- Green: the same focused command passed after adding the presence bit and
  routing `prioritize` parsing/display through it.
- Focused regression:
  `go test -run 'Test(GetPrioritizedActions|ApplyIvyCheckParams|ParseIvyCheckParams|ReadParams|GetCheckedActions)' -count=1`
  passed.
- Regression gate: `make test` passed after the fix.

### P. Existing action and logic utility divergence regressions

Status: fixed.

These divergences were already represented by Go conformance tests before this
audit pass, but were not listed in this audit file. The tests now pass and
therefore serve as regression coverage for fixed behavior:

- `logic_util.py:substitute_ast` does not treat `Lambda` as a quantifier, so
  substitution flows into lambda bodies. Go `SubstituteByName` now matches this
  behavior (`TestDIV2_SubstituteByNameLambda`).
- `logic_util.py:substitute_apply` asserts that a substitution function does
  not introduce new free variables. Go `SubstituteApply` now panics on that
  violation (`TestDIV4_SubstituteApplyMissingAssertion`).
- Python binder traversal treats `Some` as a binder. Go
  `VariablesAstList`/free-variable traversal now excludes variables bound by
  `Some` (`TestDIV5_SomeBinderNotExcluded`).
- `logic_util.py:normalize_quantifiers` rejects unexpected `Lambda` input. Go
  `NormalizeQuantifiers` now panics on the same unsupported shape
  (`TestDIV7_NormalizeQuantifiersLambda`).
- `ivy_actions.py:WhileAction.expand` filters `SubgoalAction` out of generated
  assumptions, propagates while-loop line numbers to generated havocs, and uses
  the `decreases` line number for generated ranking checks. Go `WhileAction`
  expansion now matches these cases (`TestDIV9_WhileExpandSubgoalFiltering`,
  `TestDIV10_WhileExpandHavocLineno`,
  `TestWhileExpandRankingChecksUseDecreasesLineno`).
- `ivy_actions.py:apply_mixin` raises an error on parameter/return count
  mismatches. Go `ApplyMixin` now panics instead of silently returning the
  second action (`TestDIV11_ApplyMixinErrorOnMismatch`).
- Python dynamic dispatch reaches `InstantiateAction.int_update`. Go
  `IntUpdate` now dispatches `*LogicInstantiateAction`
  (`TestDIV12_InstantiateActionDispatch`).
- Python `checked_assert` comparison uses the full `Location`, not line number
  alone. Go `AssertAction.ActionUpdate` now distinguishes same-line assertions
  in different files (`TestDIV14_CheckedAssertIgnoresFile`).

Verification:

- Focused regression:
  `go test -run 'TestDIV|TestWhileExpandRankingChecksUseDecreasesLineno' -count=1`
  passed.
- Regression gate: `make test` passed after item O; these tests are part of
  that full suite.

### Q. Historical fixed divergence ledger imported from `audit.md`

Status: fixed.

The repository already had a detailed divergence ledger in `audit.md`. This
section imports its confirmed fixed items into this audit file so the current
`CODEX_DIVERGENCE_AUDIT.md` is the complete index. The detailed fix notes and
test commands remain in `audit.md`; the full `make test` gate after item O
covered these Go tests.

| `audit.md` item | Fixed divergence |
| --- | --- |
| 1 | L2S trace hooks mutated fields that Go never rendered. |
| 2 | Method-based subgoal failures discarded the transformed trace-hook result. |
| 3 | `CheckFinalCond` displayed input clauses instead of model valuations. |
| 4 | L2S trace renaming used text replacement and corrupted overlapping nonce names. |
| 5 | Go `Trace` was not the annotation handler that Python `Trace` is. |
| 6 | `TraceBase` state construction emitted identity/empty states instead of model-derived states. |
| 7 | `TraceBase.to_lines` missed Python action renaming and line-number rendering. |
| 8 | Non-detailed trace rendering was missing. |
| 9 | State-equation rendering omitted Python's rename/reduce/filter pipeline. |
| 10 | `CheckVC` ignored requested CTI relation minimization. |
| 11 | Failing final conditions were popped before returning the model. |
| 12 | Trace hooks could not model Python's transform-and-return contract. |
| 13 | `l2s_full` loop-start hooks lost the saved-state index. |
| 14 | Trace cloning dropped model/vocabulary state needed for subtraces. |
| 15 | `TraceBase.Handle` did not skip `"nowhere"` internal actions. |
| 16 | Trace states did not receive the model universe. |
| 17 | `CheckVC` returned a trace without Python's default hidden-symbol predicate. |
| 18 | `ValueToStr` omitted Python's array and destructor rendering. |
| 19 | `MakeCheckArt` returned the wrong tuple and did not construct Python's fail state. |
| 20 | `MakeVC` did not execute the action or preserve Python's annotation fixup. |
| 21 | `CheckVC` minimized only formula-mentioned sorts, not all uninterpreted sorts. |
| 22 | `CheckVC` ignored `shrink` and always requested small-model minimization. |
| 23 | `MatchAnnotation` did not recurse through `FailAction` before marking failure. |
| 24 | `MatchAnnotation` omitted Python's extra empty-assume wrapper for existential `if` conditions. |
| 25 | Unlabeled environment choices passed an empty `EnvAction` to the handler. |
| 26 | Trace-matcher while expansion used the `while` line for ranking checks. |
| 27 | `MatchHandler` dropped Python's source-location formatting. |
| 28 | `MatchHandler` silently skipped line-zero actions that Python handles. |
| 29 | `History.SatisfyWithCond` ANDed final conditions where Python ORed them. |
| 30 | `History.SatisfyWithCond` converted reconstructed state clauses through a formula. |
| 31 | `CheckFinalCond` masked missing annotations instead of failing like Python. |
| 32 | `CheckFinalCond` rejected nil final conditions that Python supports. |
| 33 | `History` stored forward renamings by name instead of by symbol identity. |
| 34 | Transition-relation axiom filters used symbol names instead of symbol identity. |
| 35 | `ShowCounterexample` rejected the actual `History.SatisfyWithCond` result type. |
| 36 | CTI relation minimization did not apply the history renaming. |
| 37 | CTI bounded check skipped `AnalysisGraph.add_initial_state`. |
| 38 | BMC/CTI step actions did not use Python's `env_action(None)` shape. |
| 39 | Standalone BMC skipped Python's optional `initialize` action. |
| 40 | Standalone BMC computed assertion-failure clauses but did not put them in the history. |
| 41 | Initializer assertion checking did not reproduce Python's leaked-variable semantics. |
| 42 | `SubgoalAction` fell through to `NullUpdate`. |
| 43 | Compiler property proof errors were logged and ignored. |
| 44 | Action compile failures registered empty fallback actions. |
| 45 | `AnalysisGraph.Unreachable` bypassed Python's module order relation. |
| 46 | Field-action type errors became no-op updates. |
| 47 | Assignment update errors became no-op updates. |
| 48 | Hierarchical assignment expansion did not follow Python runtime key semantics. |
| 49 | `ReachState` dropped Python's tagged disjunct and model-derived under-state. |
| 50 | `ReachStateFromPred` smoothed over Python's undefined-local failure path. |
| 51 | `Diagram` ignored Python's implied/weakening/upward-close parameters. |
| 52 | Action decomposition skipped Python's state-sensitive local/call/while logic. |
| 53 | VMT transition generation used `NullUpdate` for every action. |
| 54 | VMT invariant collection skipped proof tactics. |
| 55 | `CreateConjActions` omitted Python's object-invariant interference check. |
| 56 | Empty tagged disjunctions returned `true` instead of Python's `Or()`. |
| 57 | Isolate import-wrapper creation dropped Python attribute/import/`extra_with` updates. |

### R. Interactive UPDR tactic is not yet a complete Python `tactics.py` port

Status: watch.

Python behavior:

- `tactics.py:UPDR.apply` drives the interactive analysis graph UPDR loop using
  `check_cover`, `arg_add_action_node`, `push_diagram`,
  `refine_or_reverse`, and propagation via `recalculate_facts`.

Current Go behavior:

- `tactics.go:UPDR.Apply` ports the loop structure, but
  `TestUPDR_InductiveWithInitializer` still skips on errors from incomplete
  ART-level helper behavior.

Risk:

- This is not a default `ivy_check.py` proof-script tactic path; `ivy_check.py`
  imports `ivy_tactics.py`, not the interactive `tactics.py` UPDR driver.
- If a UI/interactive path invokes Go UPDR, it can fail where Python may
  continue.

Proposed fix:

- Treat this under the UI/interactive tactics audit, not as a blocking
  `goivy_check` parity bug. When that surface is prioritized, convert the skip
  into a red test, complete the missing ART helper behavior, and require the
  UPDR test to pass.

### S. `goivy_check` boolean parameters accept values Python rejects

Status: fixed.

Python behavior:

- `ivy_utils.BooleanParameter` accepts only the literal strings `"true"` and
  `"false"`.
- `ivy_init.read_params` reports an `IvyError` for values such as
  `diagnose=yes` or `diagnose=1`.

Former Go behavior:

- `parseIvyCheckBool` accepts `"1"` and `"yes"` as true.
- Any unrecognized string silently becomes false.

Risk:

- Bad command-line values that Python rejects can silently change checker
  configuration in Go.

Fix:

- Make the `goivy_check` parameter applier validate booleans with the same
  `"true"`/`"false"` rule as Python and return an error on anything else.
- Propagate that validation error from every boolean key handled by
  `ApplyIvyCheckParams`.

TDD log:

- Red:
  `go test -run TestApplyIvyCheckParamsRejectsBadBooleanLikePython -count=1`
  failed because Go accepted `diagnose=yes`.
- Green:
  `go test -run 'TestApplyIvyCheckParamsRejectsBadBooleanLikePython|TestApplyIvyCheckParamsMarksEmptyPrioritizePresentLikePython' -count=1`
  passed after the fix.
- Focused regression:
  `go test -run 'Test(ParseIvyCheckParams|ApplyIvyCheckParams|ReadParams|BooleanParameter|DiagnoseParameter|CoverageParameter|OptTrustedParameter|OptMCParameter|OptTraceParameter|OptIvyStatsParameter|NoCheckGuaranteesParameter|ProfilingParameter)' -count=1`
  passed.
- Regression gate: `make test` passed after the fix.

### T. `parser=` checker parameter panics instead of reporting a parameter error

Status: fixed.

Python behavior:

- No `parser` parameter is registered on the `ivy_check.py` path.
- `ivy_init.read_params` therefore reports an undefined-parameter `IvyError`.

Former Go behavior:

- `ApplyIvyCheckParams` has a special `parser` case that panics with
  `"parser is no longer a choice; we only have the one now."`.

Risk:

- A bad command-line parameter can crash Go instead of producing the Python
  style parameter diagnostic.

Fix:

- Return an error for `parser` like any other unsupported goivy_check
  parameter.

TDD log:

- Red:
  `go test -run TestApplyIvyCheckParamsParserParameterReturnsErrorLikePython -count=1`
  failed because `ApplyIvyCheckParams` panicked for `parser=lalr`.
- Green:
  `go test -run 'TestApplyIvyCheckParams(ParserParameterReturnsErrorLikePython|RejectsBadBooleanLikePython|MarksEmptyPrioritizePresentLikePython)|TestParseIvyCheckParams|TestReadParams' -count=1`
  passed after the fix.
- Regression gate: `make test` passed after the fix.

### U. `close_unmatched` quantifies inside `TemporalModels` instead of around it

Status: fixed.

Python behavior:

- `ivy_proof.close_unmatched` takes the raw `goal_conc(goal)`, finds unmatched
  free variables in that raw conclusion, and wraps the raw conclusion with
  `il.ForAll([v], conc)`.
- When the raw conclusion is `ivy_ast.TemporalModels`, the resulting shape is a
  quantifier whose body is the original `TemporalModels` node.

Former Go behavior:

- `proof_phase5_goals.go:CloseUnmatched` unwraps `TemporalModels` with
  `ConcAsExpr`, quantifies the inner formula, and then rewraps the quantifier
  inside `TemporalModels`.

Risk:

- Schema instantiation proof steps that close unmatched variables over temporal
  goals can produce a different proof goal shape than Python, changing later
  tactic matching and xtrace/canonical output.

Fix:

- Use the AST-level `Forall` wrapper when the raw conclusion is not a plain
  logic expression, so the quantifier wraps `TemporalModels` exactly as Python's
  duck-typed proof AST does.

TDD log:

- Red:
  `go test -run TestCloseUnmatchedWrapsTemporalModelsConclusionLikePython -count=1`
  failed because Go returned `TemporalModels(... ForAll ...)`.
- Green:
  the same focused command passed after `CloseUnmatched` switched raw AST
  conclusions to `AstConfig.NewForall`.
- Focused regression:
  `go test -run 'Test(CloseUnmatched|Goal|AssumeTactic|Tempind|IfTactic|WrapImplies)' -count=1`
  passed.
- Regression gate: `make test` passed after the fix.

### V. `prioritize` action splitting trims whitespace that Python preserves

Status: fixed.

Python behavior:

- `ivy_check.get_prioritized_actions` evaluates
  `list(map(lambda x: 'ext:'+x, pas.split(',')))`.
- Split components are not stripped, so `prioritize=send, recv` becomes
  `["ext: recv", "ext:send"]` after sorting.

Former Go behavior:

- `check.go:GetPrioritizedActions` applies `strings.TrimSpace` to each split
  component, turning the same input into `["ext:recv", "ext:send"]`.

Risk:

- A misspelled or space-padded priority list can affect Go's action ordering
  where Python would leave the padded name unmatched against public actions.

Fix:

- Remove trimming from `GetPrioritizedActions` and preserve the raw split text.

TDD log:

- Red:
  `go test -run TestGetPrioritizedActionsPreservesWhitespaceLikePython -count=1`
  failed because Go returned `[ext:recv ext:send]`.
- Green:
  `go test -run 'TestGetPrioritizedActions|TestApplyIvyCheckParamsMarksEmptyPrioritizePresentLikePython' -count=1`
  passed after removing `strings.TrimSpace`.
- Regression gate: `make test` passed after the fix.

### W. Selective assertion CLI parameter is `assert`, not `checked_assert`

Status: fixed.

Python behavior:

- `ivy_actions.py` registers the selective assertion location parameter as
  `iu.Parameter("assert", "", ...)`.
- The parameter checker requires exactly one colon, then `p_c_a` converts
  `file:line` into `iu.Location(file + ".ivy", int(line))`.
- Consequently `assert=foo:17` is accepted and represented as `foo.ivy:17`,
  while `checked_assert=foo:17` is an unknown command parameter.

Former Go behavior:

- `ivycheck_params.go:ApplyIvyCheckParams` accepts `checked_assert` and stores
  the raw string in `cfg.CheckLineno`.
- It does not accept Python's public `assert` parameter.

Risk:

- `goivy_check assert=none:0 ...` fails before checking where Python uses it
  to request the `NOT CHECKED` sentinel.
- `goivy_check checked_assert=none:0 ...` is accepted by Go but never becomes
  Python's `none.ivy:0` sentinel.

Fix:

- Added coverage for `assert=file:line` location normalization, malformed
  `assert` rejection, and `checked_assert` rejection.
- Replaced the Go-only `checked_assert` command parameter with the Python
  `assert` parameter and Python's `file + ".ivy"` conversion rule.

TDD log:

- Red:
  `go test -run 'TestApplyIvyCheckParams(AssertParameterNormalizesLocationLikePython|RejectsBadAssertLocationLikePython|CheckedAssertParameterReturnsErrorLikePython)' -count=1`
  failed because Go rejected `assert=foo:17` and accepted
  `checked_assert=foo:17`.
- Green:
  `go test -run 'TestApplyIvyCheckParams(AssertParameterNormalizesLocationLikePython|RejectsBadAssertLocationLikePython|CheckedAssertParameterReturnsErrorLikePython|ParserParameterReturnsErrorLikePython|RejectsBadBooleanLikePython|MarksEmptyPrioritizePresentLikePython)|TestParseIvyCheckParams|TestReadParams' -count=1`
  passed after the fix.
- Regression gate: `make test` passed after the fix.

### X. `goivy_check` misses parameters registered by imported Python modules

Status: fixed.

Python behavior:

- `ivy_check.py` imports modules that register additional global
  `ivy_utils.Parameter` objects before `ivy_init.read_params` runs.
- These include solver parameters (`seed`, `incremental`, `show_vcs`),
  trace/UI parameters (`detailed`, `ui`, `mode`, `use_numerals`, `new_ui`,
  `catch`, `debug`), isolate parameters (`coi`, `filter_symbols`,
  `create_imports`, `interference`, `ext`, and others), model-checking
  parameters (`fullqi`), compiler options (`mutax`), and liveness debug
  switches (`l2s_debug`, `ranking_debug`, `abs_init`).

Former Go behavior:

- `ivycheck_params.go:ApplyIvyCheckParams` recognizes only the direct
  `ivy_check.py` options plus a few already-audited additions.
- Many Python-accepted command parameters therefore fail as unknown before the
  Go checker starts, even though Go already has corresponding config fields for
  most of them.

Risk:

- Python command lines that tune solver behavior, isolate construction,
  diagnostic trace rendering, UI mode, mutable-axiom checking, or liveness
  debug output cannot be reproduced with `goivy_check`.

Fix:

- Added TDD coverage for the imported parameter surface and Python's `seed`
  integer / `mode` enumeration validation.
- Extended `ApplyIvyCheckParams` to map those keys onto the existing Go
  `Config`, `SolverOptions`, `IsolateConfig`, and `IvyUtilsConfig` fields,
  adding only the missing `catch` and UI `mode` storage fields needed to
  preserve Python parameters.

TDD log:

- Red:
  `go test -run 'TestApplyIvyCheckParams(ImportedPythonParameterSurface|RejectsBadSeedLikePython|RejectsBadModeLikePython)' -count=1`
  failed at compile time because Go had no config storage for Python's
  `catch` or `mode` parameters; without those fields the remaining parameters
  would also be rejected as unknown.
- Green:
  the same focused command passed after adding the config fields and parameter
  mappings.
- Focused regression:
  `go test -run 'TestApplyIvyCheckParams|TestParseIvyCheckParams|TestReadParams' -count=1`
  passed.
- Regression gate: `make test` passed after the fix.

### Y. `complete` checker parameter accepts invalid logic names

Status: fixed.

Python behavior:

- `ivy_module.py:param_logic` registers `complete` with
  `check=lambda ls: all(s in il.logics for s in ls.split(','))`.
- Valid names are `epr`, `qf`, and `fo`; empty strings and unknown names are
  parameter errors.

Former Go behavior:

- `ivycheck_params.go:ApplyIvyCheckParams` stores `complete` unchanged in
  `cfg.CompleteLogic`.
- Invalid values such as `complete=bogus` or `complete=` therefore reach later
  checker code where Python would stop during parameter parsing.

Risk:

- Go can run with an impossible logic-selection value and silently change
  fragment/completeness behavior instead of reporting the command-line error.

Fix:

- Added TDD coverage for valid comma-separated logic lists and Python-style
  rejection of empty or unknown logic names.
- Validated the Go `complete` value against the existing `KnownLogics` table
  before storing it.

TDD log:

- Red:
  `go test -run TestApplyIvyCheckParamsValidatesCompleteLogicsLikePython -count=1`
  failed because Go accepted `complete=`.
- Green:
  `go test -run 'TestApplyIvyCheckParams(ValidatesCompleteLogicsLikePython|ImportedPythonParameterSurface|RejectsBadSeedLikePython|RejectsBadModeLikePython)|TestParseIvyCheckParams|TestReadParams' -count=1`
  passed after the validation helper was added.
- Regression gate: `make test` passed after the fix.

### Z. `seed` checker parameter is parsed but not applied to Z3

Status: fixed.

Python behavior:

- `ivy_solver.py:opt_seed` has `process=int` and a callback that calls
  `z3.set_param('smt.random_seed', seed)` whenever the user supplies `seed=...`.
- The default value does not invoke the callback; only an explicit parameter
  write does.

Former Go behavior:

- After item X, `ApplyIvyCheckParams` stores the integer in
  `cfg.SolverOpts.Seed`.
- `z3bridge_solver.go:applyZ3SolverOptions` only forwards `smt.macro_finder`
  to Z3, so the parsed seed never affects solver construction.

Risk:

- Runs that rely on `seed=...` for reproducible Z3 search order can diverge
  between Python and Go even with the same command line.

Fix:

- Added TDD coverage for an explicit seed producing `smt.random_seed` in the
  solver option map and for the default seed remaining absent.
- Tracked whether `seed` was explicitly set, then made solver construction apply
  `smt.random_seed` when the seed is explicit or programmatically non-zero.

TDD log:

- Red:
  `go test -run 'Test(SolverOptionParamValuesIncludesExplicitSeedLikePython|ApplyIvyCheckParamsImportedPythonParameterSurface)' -count=1`
  failed at compile time because Go had no explicit-seed bit and no inspectable
  solver option map.
- Green:
  the same focused command passed after adding `SolverOptions.SeedSet`,
  setting it from `seed=`, and routing solver construction through
  `solverOptionParamValues`.
- Focused regression:
  `go test -run 'Test(Z3BridgeDefaultOptions|Z3BridgeNewWithOptions|SolverOptionParamValuesIncludesExplicitSeedLikePython|ApplyIvyCheckParams|ParseIvyCheckParams|ReadParams)' -count=1`
  passed.
- Regression gate: `make test` passed after the fix.

### AA. `show_vcs` checker parameter is parsed but ignored

Status: fixed.

Python behavior:

- `ivy_solver.py:get_small_model` prints `definitions:` and `axioms:` before
  solving when `show_vcs=true`.
- For final conditions, it also prints `assume: ...` for assumed conditions
  and `assert: ...` for checked conditions before handing them to Z3.

Former Go behavior:

- After item X, `show_vcs` sets `cfg.SolverOpts.ShowVCs`.
- `z3bridge_solver_model.go:GetSmallModelWithCond` never reads the option, so
  no VC display occurs.

Risk:

- `goivy_check show_vcs=true ...` omits the diagnostics Python users rely on
  when comparing generated verification conditions.

Fix:

- Added TDD coverage that captures `GetSmallModelWithCond` stdout and expects
  Python's stable `definitions:`, `axioms:`, `assume:`, and `assert:` labels.
- Printed the same labels from the Go solver path when `SolverOptions.ShowVCs`
  is enabled.

TDD log:

- Red:
  `go test -run TestGetSmallModelShowVCsPrintsFinalConditionsLikePython -count=1`
  failed because the output had no `definitions:` label.
- Green:
  the same focused command passed after adding `showVCsBase` and
  `showVCsFinalCond` to the solver path.
- Focused regression:
  `go test -run 'Test(Z3BridgeDefaultOptions|Z3BridgeNewWithOptions|SolverOptionParamValuesIncludesExplicitSeedLikePython|GetSmallModelShowVCsPrintsFinalConditionsLikePython|Z3BridgeFailingFinalCondModelKeepsConditionLikePython|GetSmallModelFinalCondAssumeOrderMatchesPython|ApplyIvyCheckParams|ParseIvyCheckParams|ReadParams)' -count=1`
  passed.
- Regression gate: `make test` passed after the fix.

## File Correspondence Inventory

This table accounts for every top-level Python `.py` file under
`pyivy/ivy/ivy`. "Go home" names the primary Go file family, not every test.

| Python file | Go home | Status |
| --- | --- | --- |
| `__init__.py` | package initialization is implicit in Go | outside checker |
| `canon.py` | `logic_canon.go`, `module_canon.go`, `ivyutils_canon.go` | mapped |
| `canon_ast.py` | `ast_canon_decl.go`, `logic_canon.go` | mapped |
| `canon_fragment.py` | `fragment_canon.go` | mapped |
| `client_server_example.py` | examples/webui smoke material | outside checker |
| `concept.py` | `webui/webui_concept*.go` | outside checker |
| `concept_alpha.py` | `webui/webui_concept_alpha.go`, `alpha/alpha.go` | outside checker |
| `concept_interactive_session.py` | `webui/webui_concept_session.go`, `webui/webui_session.go` | outside checker |
| `cy_elements.py` | webui JS/Cytoscape rendering, partial Go data structs | outside checker |
| `cy_render.py` | `art_cyrender.go`, `webui/webui_cyrender.go` | outside checker |
| `cy_styles.py` | `webui/webui_cystyles.go` | outside checker |
| `dot_layout.py` | `dotgraph/dot_layout.go` | outside checker |
| `general.py` | no checker-core Go counterpart identified | outside checker |
| `interrupt_context.py` | no checker-core Go counterpart identified | outside checker |
| `iupdr.py` | `iupdr/iupdr.go` | outside checker |
| `ivy.py` | command entrypoints and parser/compiler startup | outside checker |
| `ivy2.py` | legacy Ivy2 staging; no live `goivy_check` counterpart | outside checker |
| `ivy_acl.py` | `acl.go` | mapped |
| `ivy_actions.py` | `actions_action.go`, `actions_update.go`, `actions_match.go`, `actions_transforms.go` | mapped |
| `ivy_alpha.py` | `alpha/alpha.go`, `tactics_ivy_tactics.go` | mapped |
| `ivy_art.py` | `art.go`, `art_phase7.go`, `art_cyrender.go` | mapped |
| `ivy_ast.py` | `ast.go`, `ast_*.go`, `ivylogic_ast_compat.go` | mapped |
| `ivy_auto_inst.py` | `autoinst.go`, `tactics_auto_inst.go` | mapped |
| `ivy_bmc.py` | `bmc.go` | mapped |
| `ivy_check.py` | `check.go`, `check_isolate_check.go`, `check_phase7.go`, `ivycheck_params.go` | mapped |
| `ivy_compiler.py` | `compiler.go`, `compiler_*.go`, `parser_*_helpers.go` | mapped |
| `ivy_compose.py` | `compose/compose.go` | outside checker |
| `ivy_concept_space.py` | `conceptspace/*.go`, `webui/webui_concept_space.go` | outside checker |
| `ivy_congclos.py` | `congclos.go` | mapped |
| `ivy_core.py` | `core/core.go` | mapped |
| `ivy_cpp.py` | `ivy2cpp/*.go` | outside checker |
| `ivy_cpp_types.py` | `ivy2cpp/cpp_types.go`, `ivy2cpp/types.go` | outside checker |
| `ivy_dafny_ast.py` | `dafnygen/*.go` | outside checker |
| `ivy_dafny_compiler.py` | `dafnygen/*.go` | outside checker |
| `ivy_dafny_grammar.py` | `dafnygen/*.go`, generated parser files | outside checker |
| `ivy_dafny_lexer.py` | `dafnygen/*.go`, lexer pieces | outside checker |
| `ivy_dafny_parser.py` | `dafnygen/*.go` | outside checker |
| `ivy_dump.py` | `ivydump/ivydump.go` | outside checker |
| `ivy_ev_parser.py` | `evparser/*.go` | outside checker |
| `ivy_ev_viewer.py` | `webui/webui_ui_evviewer.go` | outside checker |
| `ivy_fragment.py` | `fragment.go`, `fragment_phase7.go`, `fragment_canon.go` | mapped |
| `ivy_graph.py` | `ivyutils_graph.go`, `webui/webui_graph*.go` | outside checker |
| `ivy_graph_ui.py` | `webui/webui_graph_widget.go` | outside checker |
| `ivy_graphviz.py` | graph rendering/export utilities | outside checker |
| `ivy_init.py` | `compiler_ivyinit.go`, `ivyutils_source_file.go` | mapped |
| `ivy_interp.py` | `interp.go`, `interp_phase4.go`, `interp_eval.go` | mapped |
| `ivy_isolate.py` | `isolate.go`, `isolate_*.go` | mapped |
| `ivy_l2s.py` | `check_l2s.go`, `check_l2s_auto.go`, `check_l2s_shared.go`, `check_l2s_hooks.go` | mapped |
| `ivy_launch.py` | `ivylaunch/ivylaunch.go` | outside checker |
| `ivy_lexer.py` | `lexer.go`, `lexer_token.go` | mapped |
| `ivy_libs.py` | `ivylibs/ivylibs.go`; command/tooling, not imported by `ivy_check` | outside checker |
| `ivy_logic.py` | `ivylogic.go`, `ivylogic_*.go`, `logic_*.go` | mapped |
| `ivy_logic_parser.py` | `logicparser.go`, `lalr_logicparser_*.go` | mapped |
| `ivy_logic_parser_gen.py` | generated parser files | mapped |
| `ivy_logic_utils.py` | `logicutil.go`, `logicutil_logic_utils.go`, `module_clauses.go`, `module_subsume.go` | mapped |
| `ivy_lsp.py` | `ivylsp/ivylsp.go` | outside checker |
| `ivy_lsp_client.py` | `ivylsp/ivylsp.go` | outside checker |
| `ivy_mc.py` | `mc_checker.go`, `mc_*.go` | mapped |
| `ivy_module.py` | `module.go`, `module_*.go` | mapped |
| `ivy_parser.py` | `parser_*.go`, grammar files | mapped |
| `ivy_ply_patch.py` | `parser_patch_parser.go` | mapped |
| `ivy_printer.py` | `printer/printer.go` | mapped |
| `ivy_proof.py` | `proof_*.go`, `proof_checker.go` | mapped |
| `ivy_ranking.py` | `check_ranking.go`, `check_ranking_tactic.go` | mapped |
| `ivy_resolution.py` | `resolution.go` | mapped |
| `ivy_shell.py` | `ivyshell/ivyshell.go` | outside checker |
| `ivy_show.py` | `webui/webui_ui_show.go` | outside checker |
| `ivy_smtlib.py` | `z3bridge_solver_smtlib.go` | mapped |
| `ivy_solver.py` | `z3bridge_solver*.go`, `z3bridge_translate*.go` | mapped |
| `ivy_tactics.py` | `tactics.go`, `tactics_ivy_tactics.go`, `proof_tactics.go` | mapped |
| `ivy_temporal.py` | `temporal.go` | mapped |
| `ivy_theory.py` | `theory.go`, `module_theory.go` | mapped |
| `ivy_to_cpp.py` | `ivy2cpp/*.go` | outside checker |
| `ivy_to_lean.py` | `leangen/leangen.go` | outside checker |
| `ivy_to_md.py` | `mdgen/mdgen.go` or no live checker use | outside checker |
| `ivy_trace.py` | `trace.go`, `z3bridge_xtrace.go` | mapped |
| `ivy_transrel.py` | `actions_transrel.go`, `mc_transforms.go` | mapped |
| `ivy_ui.py` | `webui/webui_ui_main.go`, `webaudit.md` | outside checker |
| `ivy_ui_cti.py` | `webui/webui_ui_cti.go`, `trace.go` CTI paths | outside checker |
| `ivy_ui_none.py` | no live checker-core Go counterpart | outside checker |
| `ivy_ui_util.py` | `webui/webui_ui_util.go` | outside checker |
| `ivy_union_find.py` | `unionfind.go` | mapped |
| `ivy_union_find2.py` | `unionfind.go` | mapped |
| `ivy_unitres.py` | `unitres.go`, `z3bridge_solver_unitres_bridge.go` | mapped |
| `ivy_utils.py` | `ivyutils_*.go`, `vprint.go` | mapped |
| `ivy_vmt.py` | `check_vmt.go` | mapped |
| `logic.py` | `logic_*.go`, `ivylogic_*.go` | mapped |
| `logic_sexp.py` | `logic_sexp.go`, `ivylogic_sexp.go` | mapped |
| `logic_util.py` | `logicutil.go`, `logicutil_logic_utils.go` | mapped |
| `proof.py` | `proof_*.go` | mapped |
| `sidecar.py` | Python helper/sidecar, no checker-core Go counterpart | outside checker |
| `tactics.py` | `tactics.go`, `tactics_ivy_tactics.go` | mapped |
| `tactics_api.py` | `tactics.go`, proof tactic registration | mapped |
| `tk_cy.py` | webui/static/browser replacement | outside checker |
| `tk_graph_ui.py` | `webui/webui_graph_widget.go`, `webaudit.md` | outside checker |
| `tk_ui.py` | webui hook/UI replacement | outside checker |
| `token_counter.py` | `lexer_token.go` and parser diagnostics where relevant | mapped |
| `type_inference.py` | `typeinfer_*.go` | mapped |
| `ui_extensions_api.py` | `webui/webui_ext_api.go` | outside checker |
| `widget_analysis_session.py` | `webui/webui_widget_analysis_session.go` | outside checker |
| `widget_cy_graph.py` | `webui/webui_widget_cy_graph.go` | outside checker |
| `widget_dialog.py` | `webui/webui_widget_dialog.go` | outside checker |
| `widget_modal.py` | `webui/webui_widget_modal_messages.go` | outside checker |
| `widget_modal_messages.py` | `webui/webui_widget_modal_messages.go` | outside checker |
| `xtracer.py` | `xtracer/*.go` | mapped |
| `z3_utils.py` | `z3bridge_z3_utils.go` | mapped |

## Python Subdirectory Inventory

These are below `pyivy/ivy/ivy`. They are not top-level modules, but they are
still accounted for here.

| Python file/subtree | Go home | Status |
| --- | --- | --- |
| `tests/test_base.py` | Go package tests and Python oracle tests | outside checker |
| `tests/test_ivy_union_find2.py` | `unionfind.go`, `ivyutils` tests | mapped |
| `ivy2/test1.py` | legacy Ivy2 experiment | outside checker |
| `ivy2/stage2.py` | legacy Ivy2 staging | outside checker |
| `ivy2/stage3.py` | legacy Ivy2 staging | outside checker |
| `ivy2/stage4.py` | legacy Ivy2 staging | outside checker |
| `ivy2/stage5.py` | legacy Ivy2 staging | outside checker |
| `ivy2/stage6.py` | legacy Ivy2 staging | outside checker |
| `ivy2/stage7.py` | legacy Ivy2 staging | outside checker |
| `utils/__init__.py` | implicit Go packages | outside checker |
| `utils/immutables.py` | Go value types and copy-on-write structures where ported | watch |
| `utils/recstruct_object.py` | Go structs/types replace Python record helpers | watch |
| `utils/rectagtuple.py` | Go structs/types replace Python record helpers | watch |
| `utils/try1.py` | experimental utility | outside checker |
| `utils/try2.py` | experimental utility | outside checker |
| `utils/try3.py` | experimental utility | outside checker |
| `utils/try11.py` | experimental utility | outside checker |
| `z3/__init__.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3consts.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3core.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3num.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3poly.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3printer.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3rcf.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3types.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |
| `z3/z3util.py` | Go uses `z3bridge`, `smt`, and `z3vendor` instead | outside checker |

## Python Repository Wrapper Files

These are under `pyivy/` or `pyivy/ivy/` but outside the `ivy` Python package
that feeds `ivy_check`.

| Python-side file | Go home | Status |
| --- | --- | --- |
| `pyivy/requirements.txt` | Go modules in `go.mod` replace Python package deps | outside checker |
| `pyivy/onetime.sh` | local setup helper | outside checker |
| `pyivy/redo.sh` | local setup helper | outside checker |
| `pyivy/LICENSE.original.ivy.txt` | source licensing | outside checker |
| `pyivy/goivy-venv/pyvenv.cfg` | local Python virtualenv metadata | outside checker |
| `pyivy/ivy/setup.py` | Python package setup | outside checker |
| `pyivy/ivy/build_submodules.py` | Python packaging/build helper | outside checker |
| `pyivy/ivy/build_v2_compiler.py` | legacy compiler build helper | outside checker |
| `pyivy/ivy/INSTALL` | Python project documentation | outside checker |
| `pyivy/ivy/PACKAGING.md` | Python project documentation | outside checker |
| `pyivy/ivy/README.md` | Python project documentation | outside checker |
| `pyivy/ivy/Vagrantfile` | Python project/dev environment | outside checker |

## Completion Log

This section tracks the current full-audit pass. A module row above is only a
map; it is not considered complete until either:

- its behavior has a fixed/open ledger entry above, or
- the review log below says no divergence was found for the relevant
  `goivy_check` path.

Completed audit groups:

- Parser and source loading:
  - `ivy_init.py`/`compiler_ivyinit.go`: parameter parsing divergence fixed in
    item C; source-file version and include-dir behavior reviewed with no
    further checker-path divergence found.
  - `ivy_parser.py`/`parser_*.go`: Ivy 1.7+ `using` fixed in item F; importer
    error propagation fixed in item I; include-stack behavior already covered
    by `TestParserNestedIncludeSeesParentIncludedStack`.
  - `ivy_lexer.py`/`lexer.go`: malformed variable subscripts fixed in item G;
    non-ASCII unquoted identifiers fixed in item H; version keyword tables
    reviewed against Python.
  - `ivy_logic_parser.py`/`logicparser.go` and `lalr_logicparser_*.go`:
    operator, binder, temporal, and version-routing behavior is covered by the
    Python-oracle LALR logic parser tests; no additional checker-path
    divergence found in this pass.
  - `ivy_ply_patch.py`: Python PLY runtime workaround; Go uses checked-in
    generated parsers, so no live checker behavior to port.
  - `ivy_libs.py`: not imported by `ivy_check`; accounted for as outside the
    checker path.

Fixed audit items in this pass:

- `ivy_check.apply_conj_proofs` vs `check.go:ApplyConjProofs`: fixed with TDD.
- `ivy_parser.p_top_using_symbol` vs `parser_grammar_v17`: fixed with TDD for
  Ivy 1.7+ `using` imports.
- `ivy_lexer.t_VARIABLE` vs `lexer.go:scanVariable`: fixed with TDD for
  malformed bracket suffixes.
- `ivy_lexer` identifier regexes vs `lexer.go`: fixed with TDD for non-ASCII
  unquoted identifier rejection.
- `ivy_parser` include/using importer calls vs `parser_include_builder.go`:
  fixed with TDD for importer error propagation.
- `ivy_compiler.apply_assert_proof` vs `compiler_phase6.go`: fixed with TDD
  for assert-proof error propagation.
- `ivy_compiler.check_properties.named_trans` vs `compiler_phase6.go`: fixed
  with TDD for named existential specialization.
- `ivy_compiler.IvyDomainSetup.named` vs `compiler_decl.go`: fixed with TDD
  for free-parameter validation and applied named-entry terms.
- `ivy_compiler.check_definitions` vs `compiler_ivy_compile.go`: fixed with
  TDD for applied named-entry redefinition checks.
- `ivy_check.get_prioritized_actions` vs `check.go:GetPrioritizedActions`:
  fixed with TDD for explicit empty `prioritize=` handling.
- Existing `TestDIV*` action/logic utility conformance regressions: recorded
  as fixed audit item P and rechecked with their focused test slice.
- Historical fixed divergences from `audit.md` entries 1-57: imported as
  fixed audit item Q so this file is the complete divergence index.
- `tactics.py:UPDR` vs `tactics.go:UPDR`: recorded as watch item R because it
  is an interactive tactics surface, not the default `ivy_check.py` proof
  tactic path.
- `ivy_utils.BooleanParameter` vs `ivycheck_params.go`: fixed with TDD for
  strict `true`/`false` validation in goivy_check parameters.
- `ivy_init.read_params` undefined `parser` parameter vs
  `ApplyIvyCheckParams`: fixed with TDD so `parser=` returns an error instead
  of panicking.
- `ivy_proof.close_unmatched` vs `proof_phase5_goals.go:CloseUnmatched`:
  fixed with TDD for temporal-model quantifier wrapping.
- `ivy_check.get_prioritized_actions` vs `check.go:GetPrioritizedActions`:
  fixed with TDD for whitespace-preserving action splitting.
- `ivy_actions.checked_assert`/`ivy_init.read_params` vs
  `ivycheck_params.go`: fixed with TDD for the public `assert=file:line`
  parameter name and Python location normalization.
- Imported Python `Parameter` objects from `ivy_solver.py`, `ivy_trace.py`,
  `ivy_utils.py`, `ivy_ui.py`, `ivy_isolate.py`, `ivy_mc.py`,
  `ivy_compiler.py`, `ivy_l2s.py`, `ivy_ranking.py`, and `ivy_art.py` vs
  `ApplyIvyCheckParams`: fixed with TDD for missing accepted parameters.
- `ivy_module.param_logic` vs `ApplyIvyCheckParams`: fixed with TDD for
  `complete=` logic-name validation.
- `ivy_solver.opt_seed` vs `z3bridge_solver.go`: fixed with TDD so explicit
  `seed=` reaches Z3 as `smt.random_seed`.
- `ivy_solver.opt_show_vcs` vs `z3bridge_solver_model.go`: fixed with TDD so
  `show_vcs=true` prints Python's VC labels for base clauses and final
  conditions.

Completed full-audit groups:

- Compiler and module construction:
  `ivy_compiler.py`, `ivy_module.py`, `ivy_ast.py`, `ivy_logic.py`,
  `ivy_logic_utils.py`, `logic.py`, `logic_util.py`, `logic_sexp.py`, and
  `type_inference.py` are covered by fixed items J-N, P, Q, X, and Y plus the
  existing compiler/parser/type-inference conformance tests. No additional
  live `goivy_check` divergence was found after the listed fixes.
- Action/update/interp:
  `ivy_actions.py`, `ivy_transrel.py`, `ivy_interp.py`, and `ivy_art.py` are
  covered by fixed items K, P, Q, W, and X. `AnalysisGraph.MakeConcreteTrace`
  remains aligned with Python's own TODO-return stub; `CheckConstraints` and
  `StratifyGoals` are Go-only interactive placeholders, not Python
  `ivy_check` behavior.
- Proof/tactics/temporal:
  `ivy_proof.py`, `proof.py`, `tactics_api.py`, `ivy_tactics.py`,
  `ivy_temporal.py`, `ivy_l2s.py`, `ivy_ranking.py`, and `ivy_auto_inst.py`
  are covered by fixed items U, the liveness-to-safety baseline fixes, and
  items P/Q. The separate interactive `tactics.py:UPDR` driver is recorded as
  watch item R because it is not the default `ivy_check.py` proof-tactic path.
- Solver/model/fragments:
  `ivy_solver.py`, `z3_utils.py`, `ivy_smtlib.py`, `ivy_fragment.py`,
  `canon_fragment.py`, `ivy_congclos.py`, `ivy_resolution.py`,
  `ivy_unitres.py`, and `ivy_theory.py` are covered by fixed items Q, X, Z,
  and AA plus the solver/model/fragments tests. No additional checker-path
  divergence was found in this pass.
- Isolates/checking/model checking:
  `ivy_isolate.py`, `ivy_mc.py`, `ivy_vmt.py`, `ivy_bmc.py`, `ivy_acl.py`, and
  `ivy_check.py` are covered by fixed items A, C, O, S, T, V, W, X, Y, and
  AA plus existing isolate/BMC/VMT/ACL tests. No remaining open checker-path
  divergence is recorded.
- UI/diagnostic surfaces used by checker failures:
  `ivy_trace.py` is covered by the imported fixed trace ledger item Q and the
  active trace tests. `ivy_ui.py`, `ivy_ui_cti.py`, `ivy_ui_none.py`,
  `ivy_ui_util.py`, `ivy_graph.py`, `ivy_graph_ui.py`, `tk_ui.py`,
  `tk_graph_ui.py`, `tk_cy.py`, `cy_*`, `widget_*`, and
  `ui_extensions_api.py` are mapped to the Go webui/diagnostic hook surface;
  non-CLI interactivity differences are recorded as watch items B, D, and R.
- Codegen/tooling modules reachable from the package:
  `ivy_cpp.py`, `ivy_cpp_types.py`, `ivy_to_cpp.py`, `ivy_to_lean.py`,
  `ivy_to_md.py`, `ivy_dafny_*`, `ivy_dump.py`, `ivy_ev_*`, `ivy_launch.py`,
  `ivy_lsp*`, `ivy_shell.py`, `ivy_show.py`, `iupdr.py`, `sidecar.py`,
  `token_counter.py`, `general.py`, and `interrupt_context.py` are outside
  the live `goivy_check` checker path unless invoked by their own commands or
  UI tools. Their Go homes are listed in the inventory above.
- Subdirectory files, vendored Z3 Python bindings, utility record helpers, and
  repository wrapper files are accounted for in the inventories above. No
  additional live `goivy_check` divergence was found there.

No pending checker-path audit group remains after item AA. Remaining non-fixed
entries are explicitly marked `watch` or `outside checker` in the tables and
ledger above rather than open bugs.
