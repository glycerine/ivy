# Code Review: check/ directory — Bugs and Python Non-conformance

## Context
The `check/` directory in goivy ports Python's `ivy_check.py`. This review compares Go against the Python source of truth and identifies bugs, missing logic, and semantic divergences. Issues are ordered by severity.

---

## Critical Bugs

### 1. `CheckConjsInState` / `CheckConjsInStateWithAG` — wrong filtering of conjectures
**Files:** `check.go:649-654`, `isolate_check.go:894-898`
**Python (line 430):** `conjs = [x for x in conjs if is_check_mod_unprovable(x)]`
This filters conjectures by `lf.unprovable == act.check_unprovable.get()`. When `check_unprovable` is false (normal mode), this keeps only non-unprovable conjectures. When `check_unprovable` is true, it keeps only unprovable ones.
**Go:** `if !c.Unprovable { checkable = append(...) }` — hardcoded to only keep non-unprovable. This is wrong when `OnlyCheckUnprovable` is true.
**Fix:** Replace `!c.Unprovable` with `IsCheckModUnprovable(mod.Cfg, c)`.

### 2. `CheckConjsInState` drops ag/post — solver never sees the actual state
**File:** `check.go:673`
**Python (line 439):** `check_fcs_in_state(mod, ag, post, [ConjChecker(c,indent) for c in lcs])`
**Go:** `CheckFcsInState(mod, checkers)` — calls with nil ag and nil post.
**Fix:** Change `CheckConjsInState` to accept and pass ag/post, or add a `CheckConjsInStateWithAG` variant (which already exists in `isolate_check.go` — use it instead).

### 3. `ConvertPostconds` never actually converts — missing state/update parameter
**File:** `check.go:738`
**Python (line 418-426):** `convert_postconds(state, postconds)` — extracts `updated, postcond, pre = state.update` and builds a renaming.
**Go:** `ConvertPostconds(pcs)` always passes nil update, so postconditions are never renamed.
**Fix:** `CheckConjsInStateWithAG` (isolate_check.go:901-903) must pass the post state's update to `ConvertPostcondsWithUpdate`.

### 4. `CheckIsolate` missing `UpdateTheory()` after promoting properties to axioms
**File:** `isolate_check.go:124-129`
**Python (line 564):** `im.module.update_theory()` — rebuilds background theory after adding props to axioms.
**Go:** Just appends to `LabeledAxioms`, never calls `mod.UpdateTheory()`.
**Fix:** Add `mod.UpdateTheory()` after the loop at line 129.

### 5. `DualClauses` uses wrong Skolem prefix
**File:** `check.go:228`
**Python `Checker.__init__` (line 218-220):** `witness=lambda v: lg.Symbol('@' + v.name, v.sort)` — Skolem constants get `@` prefix.
**Go:** `VarToSkolem("__", v)` — uses `__` prefix.
**Fix:** Change the Skolem prefix to `"@"` to match Python.

### 6. `CheckIsolate` guarantee checking uses wrong state (missing `fail_expr`)
**File:** `isolate_check.go:217`
**Python (lines 619-620):**
```python
fail = itp.State(expr = itp.fail_expr(ag.states[0].expr))
check_safety_in_state(mod, ag, fail)
```
**Go:** `CheckSafetyInStateWithAG(mod, ag, ag.States[0], true)` — passes the init state directly, missing `fail_expr` transformation. Same issue at line 387 where post is passed instead of `fail_expr(post.expr)`.
**Fix:** Apply `interp.FailExpr()` to create a fail state before checking safety.

---

## Significant Non-conformance

### 7. `CheckIsolate` missing `!unprovable` guards on several checks
**File:** `isolate_check.go`
- **Line 89:** Property checking missing `!mod.Cfg.OnlyCheckUnprovable` condition (Python line 546: `not unprovable`)
- **Line 189:** Invariant initialization check missing `!unprovable` (Python line 599)
- **Line 212:** Initializer guarantee check missing `!unprovable` (Python line 615)
**Fix:** Add `!mod.Cfg.OnlyCheckUnprovable` to each condition.

### 8. `CheckIsolate` checked_invariants not filtered by `is_check_mod_unprovable`
**File:** `isolate_check.go:140`
**Python (line 571):** `checked_invariants = [x for x in mod.labeled_conjs if is_check_mod_unprovable(x)]`
**Go:** `checkedInvariants := mod.LabeledConjs` — uses all conjectures unfiltered.
**Fix:** Filter using `IsCheckModUnprovable`.

### 9. Missing `Ranking` action type checks throughout
**Files:** `helpers.go:84`, `isolate_check.go:204,329`, `isolate_check.go:789`
**Python:** `isinstance(sub, (act.AssertAction, act.Ranking))` — checks for both types.
**Go:** Only checks for `*actions.AssertAction`, missing `*actions.Ranking`.
**Fix:** Add `Ranking` type assertion wherever `AssertAction` is checked.

### 10. `PrettyLineno` format doesn't match Python, IGNORE DO NOTHING HERE.

### 11. `checkFcsNormalPath` reimplements solver instead of using `history.satisfy`
**File:** `check.go:540-623`
**Python (line 413):** `res = history.satisfy(axioms, gmc, filter_fcs(fcs))` — delegates to History's satisfy method which calls `small_model_clauses`.
**Go:** Creates its own Z3 solver with push/pop. This may produce different results because `satisfy` and `small_model_clauses` have additional logic (shrinking, model extraction, etc.).
Also, Go doesn't call `FilterCheckers` in the normal path (Python calls `filter_fcs(fcs)`).
**Fix:** Port `History.Satisfy` method and use it.

### 12. `CheckIsolate` missing fragment checker call
**File:** `isolate_check.go` (after line 54)
**Python (line 521):** `ifc.check_fragment()` — validates the isolate is in a decidable fragment.
**Go:** Skipped entirely.
**Fix:** Add call to `fragment.CheckFragment()` when available.

### 13. `CheckIsolate` missing theory_context wrapping
**File:** `isolate_check.go:30-414`
**Python (line 525):** `with im.module.theory_context():` — wraps the entire body.
**Go:** The default case in `CheckModule` wraps with `TheoryContext()`, but `CheckIsolate` itself doesn't. When called from `CheckSubgoals`, it may not have a theory context.
**Fix:** Add wrapping inside `CheckIsolate`.

### 14. Initializer guarantee loop structure wrong
**File:** `isolate_check.go:199-210`
**Python (lines 608-611):**
```python
guarantees = [sub for sub in action.iter_subactions()
              if isinstance(sub,(act.AssertAction,act.Ranking))
              for action in mod.initializers]
```
Wait — this Python comprehension is actually `for sub in action.iter_subactions() ... for action in mod.initializers` but `action` is rebound in the outer scope. This is a Python bug/quirk where `action` refers to the last value in the loop scope from earlier code. The intent is to iterate over all initializers.
**Go:** Correctly iterates over `mod.Initializers` then sub-actions. But missing the `check_lineno` filter (Python line 612-613).
**Fix:** Add check_lineno filtering to initializer guarantees.

---

## Minor Issues / Code Smells

### 15. `ModuleLFToAstLF` / `AstLFToModuleLF` are no-ops
**File:** `helpers.go:311-347`
Both functions convert `*ast.LabeledFormula` → `*ast.LabeledFormula` (same type). The comments say "converts module LF to ast LF" but they're the same type. These are vestigial from an earlier design where there were two distinct types.
**Fix:** Remove.

### 16. `PrintDots` takes an `int` parameter; Python takes no params
**File:** `phase7.go:51`
**Python (line 201-203):** `def print_dots(): print('...', end=' ')`
**Go:** `PrintDots(n int)` with repeat logic.
**Fix:** Remove the `n` parameter, always print `"... "`.

### 17. `GetCheckedActions` sets `CheckedActionFound` even when action not found
**File:** `check.go:700`
**Python (line 184-187):** Only sets `checked_action_found = True` inside the `if not(cact and cact not in ...)` branch.
**Go:** Sets `mod.Cfg.CheckedActionFound = true` on line 700, before the check for empty. But if `cact` is specified and not in public actions, the function returns nil at line 698 *before* setting the flag. So actually the Go is correct for that case. However, for the case where cact is empty, Go sets the flag then returns all public actions — that matches Python.

### 18. `CheckSubgoals` missing `goal_vocab` / `WithSymbols` / `WithSorts` context
**File:** `isolate_check.go:471-493`
**Python (lines 778-780):**
```python
vocab = ivy_proof.goal_vocab(goal)
with lg.WithSymbols(vocab.symbols):
    with lg.WithSorts(vocab.sorts):
```
**Go:** Doesn't establish the symbol/sort context when checking subgoals.
**Fix:** Add `GoalVocab` call and symbol/sort context when ported.

### 19. `CheckSubgoals` TemporalModels branch incomplete
**File:** `isolate_check.go:441-493`
Missing most of the model field assignments (invars, postconds, calls, binding_map, init, asms, params, updates). Also missing the definition/derived-update handling from premises (Python lines 757-758).

---

## Verification Plan
After fixes:
1. Run `make test` in goivy to ensure compilation
2. Run existing check/ tests: `cd check && make test`
3. Test with a simple Ivy file that has properties, conjectures, and assertions
4. Compare Go output format to Python output on same input file
