# FIXME.md — Incomplete Ports from Python to Go

This file lists every function that was simplified, stubbed, or only partially
ported during the Phase 3–7 bulk port. Each entry includes the Go file, the
corresponding Python function, and a description of what is missing.

---

## 1. `DestrAsgnVal` — actions/phase3.go

**Python:** `ivy_actions.py:destr_asgn_val` (lines 428–454)

The Go version handles the recursive destructor-chain structure and basic case
but omits three critical pieces:

- **Skolem nondeterministic value:** Python creates `nondet = mut_n.suffix("_nd").skolem()`
  and then `mk_assign_clauses(mut_n, nondet(*sym_placeholders(mut_n)))`. The Go version
  does not generate the skolem symbol or the assignment clauses; it returns
  `co.TrueClauses(nil)` instead of the real `new_clauses`.
- **Destructor frame conditions:** The Python loop
  `for destr in ivy_module.module.sort_destructors[mut.sort.name]` appends
  `eq_atom(destr(*a1), destr(*a2))` for every other destructor of the same sort.
  The Go version omits this loop entirely.
- **Equality guard formulas:** The Python builds `eqs` from non-variable arguments
  and appends `Or(And(*eqs), equiv_ast(dlhs, drhs))`. The Go version does not
  build these guard formulas.

To finish: port the `sym_placeholders`, `mk_assign_clauses`, `equiv_ast` helpers
and implement all three missing pieces faithfully.

---

## 2. `AssignRefs` — actions/phase3.go

**Python:** `ivy_actions.py:assign_refs` (lines 457–467)

The Go version walks the destructor chain and collects symbols, but uses a
generic `collectSymbols` walk instead of the Python's `symbols_ast` which
returns the set of all constant/function symbols (not variables). The Python
version explicitly calls `refs.add(n.rep)` for the destructor itself and
`refs.update(symbols_ast(a))` for the remaining arguments.

To finish: use `co.UsedSymbolsAST` (the Go equivalent of `symbols_ast`) instead
of `collectSymbols` to match the Python semantics exactly. Also ensure the
destructor symbol itself (`n.rep` / `app.Func`) is added to the refs set.

---

## 3. `MakeFieldUpdate` — actions/phase3.go

**Python:** `ivy_actions.py:make_field_update` (lines 680–685)

The Go version validates the field sort but does not compute the action update.
The Python creates `aa = AssignAction(f(l,v), r(v))` and then calls
`aa.action_update(domain, pvars)` to produce the transition relation triple
`(modified, clauses, pre)`.

To finish: after constructing the `AssignAction`, call the Go equivalent of
`action_update` (which is `actions.IntUpdate` via the `UpdateContext`). Return
the resulting `*transrel.Update` instead of just `error`.

---

## 4. `ExtractPrePostModel` — transrel/phase4.go

**Python:** `ivy_transrel.py:extract_pre_post_model` (lines 455–462)

The Go version filters formulas by checking whether their symbol names match an
ignore predicate. The Python version calls `clauses_model_to_clauses(clauses,
ignore=ignore, model=model, numerals=use_numerals())` which actually queries
the Z3 model to extract concrete values for each symbol, building new
ground-term equalities.

To finish: integrate with `solver.Solver` to call the Go equivalent of
`clauses_model_to_clauses` (which is `solver.ClausesModelToClauses` or similar).
The `model` parameter (a Z3 model handle) must be threaded through to extract
concrete values. The `numerals` flag controls whether integer values are
displayed as numerals or named constants.

---

## 5. `SmallModelClauses` — transrel/phase4.go

**Python:** `ivy_transrel.py:small_model_clauses` (lines 569–571)

The Go version calls `slv.GetSmallModel(cls, nil, nil)` ignoring the
`final_cond` and `shrink` parameters. The Python passes
`ivy_logic.uninterpreted_sorts()` as the sorts to minimize and passes
`final_cond` and `shrink` through to `get_small_model`.

To finish: pass the module's uninterpreted sorts (from `il.Sig.Sorts`) as the
`sortsToMinimize` argument. Thread `finalCond` through to
`solver.GetSmallModelWithCond`. Thread `shrink` through (the solver already
has a `shrink` parameter in `GetSmallModelWithCond`).

---

## 6. `GetCore` — interp/phase4.go

**Python:** `ivy_interp.py:get_core` (lines 178–186)

The Go version returns the entire `state.Clauses` as the "core" if implication
holds. The Python calls `unsat_core(clauses1, clauses2)` which invokes Z3's
unsat-core extraction to return a minimal subset of `clauses1` that is
inconsistent with `clauses2`.

To finish: call `solver.Solver.UnsatCore` (already implemented in
`solver/solver.go:439`) with the conjoined state+axioms clauses and the negated
clause. Return the extracted core subset, or `nil` if the clause is not implied.

---

## 7. `ReverseJoinConcreteClauses` — interp/phase4.go

**Python:** `ivy_interp.py:reverse_join_concrete_clauses` (lines 251–262)

The Go version checks conjunction satisfiability but omits the interpolant path.
The Python, after finding no compatible joined state, computes
`interpolant(pre, clauses, axioms, state.domain.functions)` and raises
`UnsatCoreWithInterpolant` if an interpolant is found. If no interpolant
is found either, it raises `IvyError("decision procedure incompleteness")`.

To finish: after the loop over `joinOf` fails, compute
`pre = or_clauses(*[s.clauses for s in join_of])`, call the interpolant
function (needs `solver.Interpolant`), and return the appropriate error.

---

## 8. `UnderapproximateState` — interp/phase4.go

**Python:** `ivy_interp.py:underapproximate_state` (lines 347–353)

The Go version adds a trivial under-approximation (the conjoined state+axioms
clauses). The Python calls `clauses_model_to_clauses(and_clauses(state.clauses,
axioms), is_skolem, implied)` which extracts a concrete model from Z3 and
builds ground-term clauses from it (filtering out skolem symbols).

To finish: integrate with the solver to call the Go equivalent of
`clauses_model_to_clauses`. Use the `is_skolem` filter (which is
`transrel.IsSkolem` in Go). Pass the `implied` parameter through.

---

## 9. `DecomposeActionApp` — interp/phase4.go

**Python:** `ivy_interp.py:decompose_action_app` (lines 477–516)

The Go version has the decomposition loop with History/BMC but returns a trivial
`TrueClauses` post-state instead of extracting concrete values from the model.
The Python, after `h.satisfy(bg)` succeeds, unpacks `(universe, path)` and
builds a chain of concrete `State` objects with `state.expr`, `state.update`,
`state.pred`, `state.universe` all set from the model path.

To finish: after `h.Satisfy(bg.ToFormula())` returns non-nil, unpack the result
into universe and path values. For each step in the path, create a `State` with
the correct `Expr` (via `ActionApp`), `Update`, `Pred`, and `Universe` fields.
Return the final state in the chain. This requires `History.Satisfy` to return
structured `(universe, []StateValue)` data.

---

## 10. `EvalAssertRhs` — interp/phase4.go

**Python:** `ivy_interp.py:eval_assert_rhs` (lines 600–605)

The Go version wraps non-RME inputs in a top state without evaluating them.
The Python wraps non-RME `rhs` in `RME(And(), None, rhs)` and then calls
`eval_state(rhs)` inside a temporary `ActionContext(domain)`, which actually
computes the state value from the ensures clause.

To finish: when `rhs` is an `*actions.RME`, create a state from its
`Requires`/`Modifies`/`Ensures` fields. When `rhs` is an `ast.Node`, wrap it
in an RME and call `EvalState` within an `ActionContext`. The
`ActionContext.Enter`/`Exit` pattern needs to be wired to set the global
`actions.Context` variable (matching the Python `with` statement).

---

## 11. `CompileExprVocab` / `CompileExprVocabExt` — proof/phase5_matching.go

**Python:** `ivy_proof.py:compile_expr_vocab` (lines 727–735)

The Go version does basic symbol/variable lookup in the vocab but does not run
the full sort-inference pipeline. The Python uses nested context managers
(`WithSymbols`, `WithSorts`, `top_sort_as_default`, `ASTContext`) and then calls
`il.sort_infer_list([expr.compile()] + vocab.variables)` which runs the
Hindley-Milner-style sort unification engine.

To finish: integrate with `typeinfer` package (already exists in the Go port as
`typeinfer/`). Push the vocab's symbols and sorts onto the signature, compile
the expression via the compiler, run `typeinfer.SortInferList`, and return the
inferred result.

---

## 12. `RemoveVarsMatch` — proof/phase5_matching.go

**Python:** `ivy_proof.py:remove_vars_match` (lines 751–759)

The Go version copies the match map without modification. The Python separates
sort-keyed entries from constant-keyed entries, then calls
`il.rename_vars_no_clash([v for s,v in sympairs], [fmla])` to alpha-rename
the constant-match values so they don't capture variables in `fmla`.

To finish: classify match entries by whether the key `il.is_ui_sort(s)` (sort
match) or `il.is_constant(s)` (symbol match). For symbol matches, call
`il.RenameVarsNoClash` (already in `ivylogic/constructors.go:171`) on the
values with `fmla` as the context.

---

## 13. `TransformDefnMatch` — proof/phase5_matching.go

**Python:** `ivy_proof.py:transform_defn_match` (lines 790–829)

The Go version returns the problem unchanged. The Python (40 lines) does:
1. Extracts `declsym`, `concsym`, `declargs`, `concargs` from both definitions.
2. Builds a variable substitution `vmap` mapping pattern args to instance args.
3. Applies `vmap` to `concrhs`.
4. Builds `dmatch` mapping the schema's defined symbol to the goal's defined
   symbol, plus all matching sorts.
5. Applies `dmatch` to `concrhs` and `freesyms`.
6. Filters out matched constants from `freesyms`.
7. Applies the combined match to the schema via `apply_match_goal`.
8. Returns a new `MatchProblem` with the rewritten RHS-vs-RHS problem.

To finish: implement all 8 steps using the existing `Match`, `ApplyMatch`,
`ApplyMatchFreesyms`, and `ApplyMatchGoalNode` functions.

---

## 14. `ParameterizeSchema` — proof/phase5_matching.go

**Python:** `ivy_proof.py:parameterize_schema` (lines 857–875)

The Go version returns the schema unchanged. The Python:
1. Creates fresh variables `vars` for the given sorts via `make_distinct_vars`.
2. For each `ConstantDecl` premise, creates a new symbol `sym2` with extended
   sort `FuncConstSort(*(sorts + list(sym.sort.dom) + [sym.sort.rng]))`.
3. Builds a match entry `match[sym] = Lambda(vs2, sym2(*(vars+vs2)))`.
4. Replaces the premise with `ConstantDecl(sym2)`.
5. Applies the match to the conclusion.
6. Returns the cloned goal with new premises and transformed conclusion.

To finish: implement the sort extension (`il.FuncConstSort`), the lambda
wrapping, and the match application. This requires `il.FuncConstSort` or
equivalent sort construction.

---

## 15. `CompileMatchFull` — proof/phase5_matching.go

**Python:** `ivy_proof.py:compile_match` (lines 919–945)

The Go version has an empty merge loop and returns nil. The Python:
1. Copies `freesyms` from the problem, optionally adding witness variables.
2. Calls `compile_match_list` to compile all match entries.
3. For each compiled match, calls `compile_one_match(m.lhs(), m.rhs(),
   freesyms, prob.constants)` to produce individual matches.
4. Merges all individual matches via `merge_matches(*matches)`.

To finish: iterate over the compiled match list, call `CompileOneMatch` for
each, collect the results, and call `MergeMatches` to combine them.

---

## 16. `MakeDistinctVars` — proof/phase5_matching.go

**Python:** `ivy_proof.py:make_distinct_vars` (lines 1215–1217)

The Go version creates variables `V0`, `V1`, ... but does not rename them to
avoid clashes with variables already present in the given ASTs. The Python calls
`lu.rename_variables_distinct_asts(vars, asts)` which alpha-renames the
variables so their names are distinct from all variables in `asts`.

To finish: call `co.RenameVariablesDistinctAsts` (already in
`clauseops/batch16.go:181`) after creating the initial variables.

---

## 17. `RenameGoal` — proof/phase5_matching.go

**Python:** `ivy_proof.py:rename_goal` (lines 703–720)

The Go version builds a flat match from the rename map and applies it. The
Python uses a recursive `rec_goal` inner function that:
1. Recursively processes premises via `map(rec_goal, goal_prems(goal))`.
2. Builds a match from `goal_defns(goal)` entries whose names are in `rmap`,
   using `x.rename(lambda n: rmap[x.name])`.
3. Chains the match through `apply_match_sym` to handle transitive renames.
4. Calls `check_alpha_capture(goal, match)`.
5. Applies the match to the goal via `apply_match_goal(match, goal, apply_match_alt)`.
6. Alpha-renames the conclusion via `il.alpha_rename(rmap, goal_conc(goal))`.
7. Renames the goal's own label via `goal.rename(rmap.get(goal.name, goal.name))`.

To finish: implement the recursive `rec_goal` structure. Use `GoalDefns` to
build the per-level match, call `CheckAlphaCapture`, apply via
`ApplyMatchGoalNode`, alpha-rename the conclusion via `il.AlphaRename`, and
rename the label via `LabeledFormula.Rename`.

---

## 18. `CompileThunkAction` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_thunk_action` (lines 683–732, ~50 lines)

The Go version returns a basic `ThunkAction`. The Python:
1. Copies the signature, compiles formals within the copy.
2. Compiles the body via `sortify(self.args[3])`.
3. Collects all `fml:`/`loc:` prefixed symbols referenced in the body.
4. Creates a sub-sort for the thunk and a `$self` parameter.
5. For each captured symbol, creates a destructor with extended sort and
   registers it in `module.destructor_sorts` / `module.sort_destructors`.
6. Builds a substitution mapping captured symbols to `dsym($self)`.
7. Applies the substitution to the body.
8. Registers the thunk's `run` action in the module.
9. Builds the final `LocalAction` wrapping assignment of captures + continuation.

To finish: port all 9 steps. This requires `il.Sig.Copy`, `il.FindSort`,
destructor registration on the module, and `co.SubstituteConstantsAST`.

---

## 19. `InferParameters` — compiler/phase6.go

**Python:** `ivy_compiler.py:infer_parameters` (lines 1578–1617, ~40 lines)

The Go version is a no-op returning nil. The Python:
1. Collects all action declarations and mixin declarations from `decls`.
2. For each mixin, maps mixer → mixee names.
3. For each action with exactly one mixee, compares formal parameter counts
   between the mixer (monitor) and the mixee.
4. If the mixer has fewer formals than required, extends its `formal_params`
   and `formal_returns` with the mixee's extra parameters.
5. Rewrites the mixin body via `ivy_ast.subst_prefix_atoms_ast`.

To finish: implement the three passes (collect, validate, extend). Requires
access to the declaration list's action and mixin entries, and the
`ast.SubstPrefixAtomsAst` rewriter.

---

## 20. `CheckProperties` — compiler/phase6.go

**Python:** `ivy_compiler.py:check_properties` (lines 1972–2011, ~84 lines)

The Go version is a no-op returning nil. The Python:
1. Calls `reorder_props(mod, mod.labeled_props)` to topologically sort properties.
2. Builds maps from property ID → proof and property ID → named instance.
3. Gives empty proofs to schema-bodied theorems without proofs.
4. Creates a `ProofChecker` from axioms, definitions, and schemata.
5. For each non-temporal property with a proof, calls
   `prover.admit_proposition(prop, pmap[prop.id])` to verify the proof.
6. For properties that are not definitions, applies `named_trans` to specialize
   named instances, then updates the prover's axioms and schemata.
7. Collects resulting subgoals.

To finish: integrate with `proof.ProofChecker` (already exists). Wire up the
`admit_proposition` path, the `named_trans` specialization, and the subgoal
collection. This is the largest single function to port.

---

## 21. `ReorderProps` — compiler/phase6.go

**Python:** `ivy_compiler.py:reorder_props` (lines 1833–1856, ~24 lines)

The Go version returns props unchanged. The Python:
1. Separates "spec" properties (those whose name has a `spec` attribute) from
   regular properties.
2. Groups spec properties by parent name.
3. Iterates in reverse, inserting spec properties before their parent.
4. Reverses the result to restore forward order.

To finish: check `mod.Attributes` for `compose_names(prop.name, "spec")`,
implement the parent-child name splitting (already exists as
`iu.ParentChildName` or similar), and perform the insertion/reversal.

---

## 22. `CompileSchemaInstantiation` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_schema_instantiation` (lines 912–941, ~30 lines)

The Go version returns nil. Note: the Python version's first line is
`return self` (early return), making the rest dead code. However, the dead
code shows the intended implementation:
1. Look up the schema by name.
2. Separate sort-matches from symbol-matches in the proof match list.
3. For sort matches, validate they reference real sorts in the signature.
4. For symbol matches, compile LHS within the schema's symbol context and
   RHS within the goal's context, running sort inference on both.
5. Return a cloned node with compiled matches.

Since Python itself returns `self` (identity), the Go version returning
the input node unchanged is actually correct for current behavior. Mark
as low priority.

---

## 23. `ReadModule` / `ImportModule` / `IvyLoadFile` / `IvyFromString` — compiler/phase6.go

**Python:** `ivy_compiler.py:read_module` (lines 2267–2296), `import_module`,
`ivy_load_file`, `ivy_from_string`

All four Go versions return empty modules. The Python versions:
- `read_module`: reads the file, parses the `#lang ivy` header to determine
  version, reloads the parser if the version changed, then calls `parse(s)`.
- `import_module`: looks up a module by name in a cache, or reads it from disk.
- `ivy_load_file`: opens a file and calls `read_module`.
- `ivy_from_string`: creates a `StringIO` and calls `read_module`.

To finish: these require the Ivy parser (in `parser/` or equivalent). Once
the parser is ported, wire these functions to call it. The version-detection
and parser-reload logic can be simplified in Go since Go doesn't have
Python's `importlib.reload` pattern.

---

## 24. `ApplyAssertProof` / `ApplyAssertProofs` — compiler/phase6.go

**Python:** `ivy_compiler.py:apply_assert_proof` (~19 lines),
`apply_assert_proofs` (~29 lines)

Both Go versions are no-ops. The Python:
- `apply_assert_proof`: walks an action tree to find `AssertAction` nodes,
  looks up their proof in `mod.proofs`, and attaches it.
- `apply_assert_proofs`: iterates over all module actions and calls
  `apply_assert_proof` for each.

To finish: iterate `mod.Actions`, walk each action's subactions, find
`AssertAction` nodes, and look up `mod.Proofs` by the assertion's ID or
line number to attach the proof. Requires the `Proofs` field on Module
to be populated by the compiler.

---

## 25. `Checked` — mc/phase7.go

**Python:** `ivy_mc.py:checked` (lines 1017–1018)

The Go version does a trivial true/false check on the action formula. The
Python checks `ia.checked_assert.value in ["", thing.lineno]`, where
`checked_assert` is a global `Parameter` whose value is either empty (meaning
"check all") or a specific `filename:line` location.

To finish: add a package-level `CheckedAssert` parameter (string). In
`Checked`, compare the action's `GetLineno()` against the parameter value.
If the parameter is empty, return true. Otherwise return true only if the
action's location matches.

---

## 26. `Badwit` — mc/phase7.go

**Python:** `ivy_mc.py:badwit` (lines 1429–1430)

The Go version always returns true. The Python raises
`iu.IvyError(None, 'model checker returned mis-formated witness')`.

To finish: change the return type or behavior to return an error (or panic)
with the message "model checker returned mis-formatted witness". The callers
should handle this error.

---

## 27. `MatchAnnotation` — mc/phase7.go

**Python:** `ivy_mc.py:match_annotation` (lines 946–1005, ~60 lines)

The Go version is an empty stub. The Python is a recursive function that
walks action/annotation pairs to reconstruct execution traces:
1. `RenameAnnotation`: applies the rename map to the environment, recurses.
2. `Sequence`: decomposes via `ComposeAnnotation`, recursing on prefix and
   last element.
3. `IfAction`: evaluates the `IteAnnotation` condition via `handler.eval`,
   recurses into the taken branch.
4. `ChoiceAction`: decomposes `IteAnnotation` via `unite_annot`, evaluates
   each branch condition, recurses into the first true branch.
5. `CallAction`: looks up the callee, wraps in
   `Sequence(IgnoreAction, callee, ReturnAction)`, recurses.
6. Base cases: calls `handler.handle(action, env)` for atomic actions.

To finish: implement the recursive `recur` inner function with all six cases.
The `AnnotationHandler` interface (already defined in `actions/match.go`) has
`Eval`, `Handle`, `DoReturn`, `Fail` methods. Use `actions.UnwrapAction` to
extract actions from nodes. The `MatchAnnotation` function in `actions/match.go`
already has a partial implementation that may need to be reconciled with this one.

---

## 28. `AigerWitnessToIvyTrace` — mc/phase7.go

**Python:** `ivy_mc.py:aiger_witness_to_ivy_trace` (lines 1570–1628, ~60 lines)

The Go version creates an `IvyMCTrace` with empty states. The Python:
1. Reads an AIGER witness file line by line.
2. For each line, splits into `pre, inp, out, post` columns.
3. Steps the AIGER simulator with the input.
4. Calls `match_annotation` with an `AigerMatchHandler` to decode the trace.
5. Advances the simulator, extracts latch values.
6. Decodes state variable values from the latch state.
7. Builds `IvyMCTrace` with concrete state equalities.

To finish: requires the AIGER simulator (`aiger.sub.step`, `aiger.sub.next`,
`aiger.get_sym`, `aiger.get_state`) and the `AigerMatchHandler` class. These
are part of the AIGER infrastructure in `mc/aiger.go`. Wire the witness file
reading, simulator stepping, and state decoding.

---

## 29. `GuiArt` — check/phase7.go (OMIT, not needed with web UI)

**Python:** `ivy_check.py:gui_art`

The Go version returns "GUI not available". The Python launches a Cytoscape
widget for interactive visualization of the analysis graph. This is a GUI
feature that may not be needed in the CLI Go port.

To finish: either implement a text-based visualization (using `art.RenderRg`)
or integrate with a web-based visualizer if the web UI is being ported. Low
priority for CLI-only usage.

---

## 30. `CompileMatchList` — proof/phase5_matching.go

**Python:** `ivy_proof.py:compile_match_list` (lines 883–893)

The Go version passes match entries through without compiling them via the
vocab contexts. The Python:
1. Gets `left_goal_vocab` and `right_goal_vocab` from the two goals.
2. If `allow_witness`, extends `left_goal_vocab.variables` with the
   conclusion's used variables.
3. For each match entry `d`, compiles `d.lhs()` via `compile_expr_vocab`
   with the left vocab, and `d.rhs()` via `compile_expr_vocab` with the
   right vocab.

To finish: for each `ast.Definition` in the match list, compile LHS and
RHS through `CompileExprVocab` with the respective goal vocabs. This depends
on fixing item 11 (`CompileExprVocab`) first.

---

## 31. `Implies` (non-Clauses branch) — transrel/phase4.go

**Python:** `ivy_transrel.py:implies` (lines 241–259)

The Go version always treats `c2` and `p2` as `*co.Clauses` and calls
`ClausesImplyFormulaCex`. The Python has a branch: if `c2` is not a `Clauses`
object (i.e. it's a raw formula), it checks `is_prenex_universal(c2)` and
calls `clauses_imply_formula_cex` directly on the formula. If `c2` is a
`Clauses`, it checks `c2.is_universal_first_order()` and calls `clauses_imply`.

To finish: add a type-switch or interface check on `c2`. If it's already a
`*co.Clauses`, call `solver.ClausesImply`; if it's a raw `lg.Node`, check
`il.IsPrenexUniversal` and call `ClausesImplyFormulaCex` on the formula
directly. Likewise for `p2`.

---

## 32. `CheckInstantiations` — compiler/phase6.go

**Python:** `ivy_compiler.py:check_instantiations` (lines 1618–1629)

The Go version does a minimal struct-field check on `mod.Instantiations` but
does not iterate the declaration list or validate that each referenced schema
name actually exists in `mod.Schemata`. The Python iterates all declarations,
finds instantiation nodes, and raises `IvyError` if the target schema is not
defined.

To finish: iterate `mod.Instantiations`, extract the schema name from each
entry, look it up in `mod.Schemata`, and return an error if not found.

---

## 33. `CompileTheory` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_theory` (lines ~5 lines)

The Go version is a no-op returning nil. The Python compiles a theory for a
given sort by looking up the theory name (e.g. "int", "nat") in a theory
registry and generating the appropriate axioms (e.g. Peano arithmetic axioms
for integers, ordering axioms for total orders).

To finish: integrate with the `theory/` package which already has theory
definitions. Call the appropriate theory compiler based on the sort's
interpretation name. Register the generated axioms on the module.

---

## 34. `CompileTheories` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_theories` (~13 lines)

The Go version is a no-op. The Python iterates `mod.interps` (sort
interpretations), and for each interpretation calls `compile_theory` to
generate and install the axioms.

To finish: iterate `mod.Interps`, extract the sort name and theory name,
call `CompileTheory` for each. This depends on item 33.

---

## 35. `AddLabelsToProof` — compiler/phase6.go

**Python:** `ivy_compiler.py:add_labels_to_proof` (~10 lines)

The Go version returns the input unchanged. The Python recursively walks
`ComposeTactics` and `IfTactic` nodes, setting `.label` on each sub-tactic
from the parent tactic's label. This is used to propagate proof labels for
error reporting.

To finish: implement a recursive walk over proof AST nodes. For
`ComposeTactics`, propagate the label to all children. For `IfTactic`,
propagate to both branches. Use `SetLineno` or a custom label field.

---

## 36. `CompileDebugAction` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_debug_action` (lines 739–746)

The Go version returns an empty `Sequence` action. The Python compiles the
debug expression and "with" clauses, creating a proper `DebugAction` node
(which is itself a no-op for semantics but preserves debugging metadata for
trace output).

To finish: instead of returning `NewSequence()`, create a
`actions.NewDebugAction(compiledExpr, compiledWithExprs...)` and return it
wrapped. The DebugAction type already exists in `actions/phase3.go`.

---

## 37. `SortInferCovariant` — compiler/phase6.go

**Python:** `ivy_compiler.py:sort_infer_covariant` (lines 214–221)

The Go version calls `c.SortInfer(term)` and then checks if the result sort
matches the target sort, but it never passes the target sort as a hint to the
inference engine. The Python calls `sort_infer(term, sort)` first (with the
hint), and only falls back to `sort_infer(term)` (without hint) if the first
attempt fails.

To finish: add a sort-hint parameter path through `SortInfer`. Call
`SortInfer(term, sort)` first; if that fails or returns a different sort,
fall back to `SortInfer(term)` and then check compatibility.

---

## 38. `SortInferContravariant` — compiler/phase6.go

**Python:** `ivy_compiler.py:sort_infer_contravariant` (lines 223–230)

Same issue as item 37 but with contravariant (inverse) direction. The Go
version does not pass the sort hint.

To finish: same fix as item 37. Additionally, the contravariant version
checks `c.Module.IsVariant(sort, termSort)` (note reversed argument order
compared to covariant).

---

## 39. `PropToDef` — compiler/phase6.go

**Python:** `ivy_compiler.py:prop_to_def` (lines 1828–1829)

The Go version only checks if the node is already an `*lg.Definition`. The
Python additionally handles `LabeledFormula` by extracting the inner formula,
stripping universal quantifiers via `il.drop_universals`, and checking if the
result is a `Definition`.

To finish: if the input is a `*module.LabeledFormula`, extract `.Formula`,
call `il.DropUniversals` to strip leading `ForAll` quantifiers, and check if
the body is a `*il.Definition`.

---

## 40. `CompileCrashAction` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_crash_action` (lines 673–679)

The Go version falls back to `NewHavocAction(nil)` in error cases. The Python
creates a `CrashAction` (not a `HavocAction`) with the compiled atom target.
The Python code is: `return CrashAction(Atom(ivy_logic.find_symbol(args...)))`.

To finish: use `actions.NewCrashAction(target)` instead of
`actions.NewHavocAction(target)`. The `CrashAction` type already exists in
`actions/action.go`.

---

## 41. `MatchHandler.Eval` — mc/phase7.go

**Python:** `ivy_mc.py:MatchHandler.eval` (lines 934–940)

The Go version returns `true` for all non-trivial conditions. The Python
evaluates the condition symbol against the Z3 model: it looks up the symbol
in the model's symbol table and returns the Boolean truth value.

To finish: look up `cond` (the renamed symbol name) in `h.Model` (the Z3
model), extract its Boolean value, and return it. This requires the model
to provide a symbol-to-value lookup interface.

---

## 42. `CloneNormal` — mc/phase7.go

**Python:** `ivy_mc.py:clone_normal` (lines 839–849)

The Go version copies formulas and definitions without normalizing. The Python
normalizes equalities by: (a) removing tautological equalities `x == x`,
(b) reordering equality arguments to a canonical form (e.g. new_ on left),
(c) simplifying trivially-true conjunctions.

To finish: iterate over `clauses.Fmlas`, filter out tautological equalities
(already have `isTautologyEquality` in transrel/phase4.go), and canonicalize
the ordering of `Eq` arguments.

---

## 43. `IvyCompileTheoryFromString` — compiler/phase6.go

**Python:** `ivy_compiler.py` compile-theory-from-string usage

The Go version delegates to `IvyFromString` which itself is a stub (item 23).
Even after item 23 is fixed, this function additionally needs to call
`CompileTheory(decls, sort)` on the parsed declarations with the given sort
and theory name.

To finish: after `IvyFromString` returns a module with parsed declarations,
call `CompileTheory` with the sort parameter. Depends on items 23 and 33.

---

## 44. `CompileSchemaPrem` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_schema_prem` (lines 869–883, ~20 lines)

The Go version handles `ConstantDecl`, `TypeDef`, and `LabeledFormula` cases
but omits the `DerivedDecl` case and the `PropertyDecl` case that the Python
handles. The Python also adds compiled symbols to the schema's temporary
signature scope.

To finish: add cases for derived declarations (compile the definition and
add to the signature) and property declarations (compile the formula). Ensure
each compiled symbol is added to `c.Sig` so subsequent premises can
reference it.

---

## 45. `CompileSchemaBody` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_schema_body` (lines 896–901)

The Go version creates a `LabeledFormula` with the schema body but does not
preserve the original label or copy attributes like `temporal`, `explicit`,
`assumed`. The Python clones the schema with all its metadata intact.

To finish: propagate the original schema's label, temporal flag, and other
attributes to the compiled result. Use the `CloneWithFreshID` pattern
from `ast.LabeledFormula`.

---

## 46. `CompileSchemaConc` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_schema_conc` (lines 889–894)

The Go version handles `Definition` and falls back to `SortifyWithInference`.
The Python additionally handles the case where the conclusion is a
`TemporalModels` node (compiling its inner formula) and sets up the
`WithSymbols` / `WithSorts` context managers for sort inference.

To finish: add a case for temporal-models conclusions. Ensure the sort
inference runs within the schema's symbol context by using
`il.NewWithSymbols(schemaSig, ...)`.

---

## 47. `MatchAnnotationMC` — mc/phase7.go

**Python:** `ivy_mc.py:match_annotation` (lines 946–1005)

The Go version delegates to `actions.MatchAnnotation`. However, the MC-specific
`match_annotation` in Python has additional logic beyond the generic version
in `actions/match.go`: it handles the `ChoiceAction` case by decomposing
`IteAnnotation` via `unite_annot` and evaluating each branch, and it handles
`CallAction` by wrapping the callee in `Sequence(IgnoreAction, callee,
ReturnAction)`.

To finish: verify that `actions.MatchAnnotation` handles all the cases that
the Python MC-specific version does. If not, extend `MatchAnnotationMC` to
handle `ChoiceAction` with `UniteAnnot` and `CallAction` with the
`Sequence(Ignore, callee, Return)` wrapping.

---

## 48. `SetVerifying` — compiler/phase6.go

**Python:** `ivy_compiler.py:set_verifying` uses `iu.BooleanParameter`

The Go version uses a package-level `optionVerifying` bool. The Python version
sets a `BooleanParameter` which is a registered parameter that can be queried
by name from the command line or other modules. Other code may check
`option_verifying.get()`.

To finish: if other Go code needs to query the verifying flag by name (e.g.
the check package), expose it via a getter function or integrate with a
parameter registry. Low priority if only checked in one place.

---

## 49. `CompileIfTactic` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_if_tactic` (lines 970–972)

The Go version compiles the condition and recursively compiles branches, which
is mostly correct. However, it silently swallows errors from branch compilation
(`err` is ignored). The Python propagates compilation errors.

To finish: propagate errors from `c.CompileTactic(ifT.Then)` and
`c.CompileTactic(ifT.Else)` instead of silently falling back.

---

## 50. `CompilePropertyTactic` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_property_tactic` (lines 976–986)

The Go version compiles the name and proof, but the Python additionally:
1. Compiles the property formula (`self.args[0]`) via `sortify`.
2. If the property has a definition, compiles it with `compile_defn`.
3. Recursively compiles the proof tactic.

To finish: compile `pt.Prop` (the property formula) via `c.Sortify` or
`c.CompileNode`. If it's a definition, use `c.CompileDefn`.

---

## 51. `CompileProofTactic` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_proof_tactic` (lines 1000–1001)

The Go version compiles the proof recursively. The Python additionally
compiles the label argument via `sortify(self.args[0])` and handles the
case where the label is a named schema reference.

To finish: compile `pt.TLabel` (the label/schema name) via `c.Sortify`
before recursing into the proof body.

---

## 52. `OtherThing` — compiler/phase6.go

**Python:** `ivy_compiler.py:other_thing` (lines 59–66)

The Go version's `isSortInferRoot` always returns false, so the sort-infer-root
branch is dead code. The Python checks `hasattr(self, 'sort_infer_root')` which
is true for `AssignAction`, `SetAction`, `HavocAction`, `AssumeAction`,
`AssertAction`, etc.

To finish: implement `isSortInferRoot` to return true for the action-like AST
node types that have the `sort_infer_root` attribute in Python. Check
`*ast.AssignAction`, `*ast.SetAction`, `*ast.HavocAction`,
`*ast.AssumeAction`, `*ast.AssertAction`.

---

## 53. `CompileRootArgs` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_root_args` (lines 56–57)

The Go version compiles each arg via `CompileNode`. The Python calls
`compile_root_arg(a)` which, for string-like args, calls `find_symbol(a)` to
resolve the name through the signature before compiling. This is important for
action atoms whose rep is a symbol name.

To finish: for `*ast.Atom` args with no sub-terms (bare names), look up the
name in `c.Sig.Symbols` via `il.FindSymbol` before falling back to
`CompileNode`.

---

## 54. `CompileNativeArg` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_native_arg` (lines 753–759)

The Go version handles variables and atoms but the Python also handles the
case where the argument is an action name (checking `is_action_like(arg)`).
For action names, Python renames the action via `resolve_alias`.

To finish: add a check for whether the atom name is in `mod.Actions`. If so,
resolve the alias and return the action's compiled symbol.

---

## 55. `CompileNativeSymbol` — compiler/phase6.go

**Python:** `ivy_compiler.py:compile_native_symbol` (lines 762–775)

The Go version handles signature symbols, sorts, numerals, and hierarchy
lookups. The Python additionally handles the case where the name is a
destructor (`name in mod.destructor_sorts`) and the case where the symbol
has a polymorphic definition (`entry.union is not None`), creating a
`PolySymsDict` wrapper.

To finish: add a check for `mod.DestructorSorts[name]` and handle the
polymorphic symbol case by constructing the appropriate overloaded symbol.
