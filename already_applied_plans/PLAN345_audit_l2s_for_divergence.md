# Deep Trace: Go vs Python L2S Divergence Audit

**Created:** 2026-04-29 ~06:30 UTC

## Context

The goivy project is a mechanical port of Python Ivy (`~/ivy/pyivy/ivy/ivy/`) into Go (`~/ivy/goivy/`). The Python is the source of truth. ~1009 divergences have been fixed so far. This audit deep-traces the l2s (liveness-to-safety) package and its supporting packages to find remaining behavioral divergences.

The l2s tactic transforms temporal (liveness) properties into safety properties. It's implemented in:
- **Python:** `ivy_l2s.py` (~1700 lines, single file)
- **Go:** `check/l2s.go`, `check/l2s_shared.go`, `check/l2s_auto.go`, `check/l2s_hooks.go`

---

## Findings Summary

After reading ~6000 lines of Go and ~3500 lines of Python across l2s and its supporting packages, here are the divergences found, ranked by severity.

---

## HIGH SEVERITY — Likely to cause wrong verification results

### D1. `SplitReturns` early return on zero returns
- **Python** `ivy_actions.py:1437`: No early return — always runs rename+split logic
- **Go** `actions/action.go:792-795`: `if len(a.ActualReturns) == 0 { return a }` — returns unmodified
- **Impact:** If a CallAction somehow has 0 actual returns but still reaches `split_returns`, Python would process it (creating an empty LocalAction wrapper) while Go skips entirely. This changes the resulting action tree structure.
- **Fix:** Remove the early return.

### D2. `SplitReturns` renaming scope
- **Python** `ivy_actions.py:1440`: `new_returns = [x.rename(rn) for x in actual_returns]` — calls `.rename()` on every return expr (handles Apply, Const, Variable, anything with a rename method)
- **Go** `actions/action.go:811-816`: Only renames `*lg.Const`; non-Const returns are kept unchanged
- **Impact:** If a return expression is an `Apply` (e.g., `f(x)`) or a destructured access, Python would rename it but Go wouldn't. This produces different variable names in the split action, which could cause the solver to see different symbols.
- **Fix:** Port the full `.rename()` dispatch from Python — handle Apply, Variable, and other node types in `SplitReturns`.

### ~~D3. `NormalizeNamedBinders` — Apply node recursion order~~ FALSE ALARM
- **Retracted after re-reading:** Python line 277 normalizes `ast.args` (terms) FIRST, then line 279 normalizes `ast.rep` (func). Go lines 900-903 do the same: terms first, then func. Go even has a comment at line 897 confirming this. The explore agent misread the Python evaluation order. **No divergence.**

---

## MEDIUM SEVERITY — Correct for tested cases, wrong for edge cases

### D4. `SymPlaceholders` return value for 0-arity symbols
- **Python** `ivy_logic_utils.py:863-867`: Returns `[]` (empty list) for constants with empty domain
- **Go** `module/clauses.go:374`: Returns `nil` for 0-arity symbols
- **Impact:** In Go, `nil` and `[]` behave the same for `len()` and `range`, but differ for `== nil` checks and some slice operations.
- **Fix:** Go returning nil is idiomatic and avoids allocation. Instead, audit all callers to ensure they use `len(vs) > 0` rather than `vs == nil` or `vs != nil`. Fix any callers that use nil-checks.

### D5. `GoalPrems` return value for non-SchemaBody
- **Python** `ivy_proof.py:644`: Returns `[]` (empty list) for non-schema goals
- **Go** `proof/goal.go:173`: Returns `nil`
- **Impact:** Similar to D4 — `nil` vs `[]` are interchangeable for most Go operations but could diverge if callers use nil-checks.
- **Fix:** Go returning nil is idiomatic and avoids allocation. Audit callers to ensure they use `len() > 0` checks, not nil-checks.

### D6. `CopyFormalsTo` missing `EnvAction` exclusion
- **Python** `ivy_actions.py:241`: `if not isinstance(res, EnvAction):` — skips copying formal params/returns to EnvAction destinations
- **Go** `actions/helpers.go:231-245`: No `EnvAction` check — always copies
- **Impact:** If the destination action is an EnvAction equivalent, Go would incorrectly set formal params/returns on it. In l2s specifically, the destinations are always `Sequence` (from `prefix_action`/`postfix_action`), so this doesn't fire. But for other callers of `CopyFormalsTo`, it could cause issues.
- **Fix:** Add `if _, isEnv := dst.(*EnvAction); !isEnv { ... }` guard.

### D7. `assume_when_axioms` — missing `.set_lineno()` in Python
- **Python** `ivy_l2s.py:1087-1090`: `AssumeAction(...)` — NO `.set_lineno(lineno)` call
- **Go** `check/l2s_shared.go:388-389`: `setLineno(actions.NewAssumeAction(inner), cfg.Lineno)` — HAS lineno
- **Impact:** Go sets lineno on when axioms but Python doesn't. This is a Python omission that Go "fixed" — but since Python is the source of truth, Go technically diverges by adding lineno. This could affect error reporting location but not verification correctness.
- **Status:** Go is arguably better here. Keep Go's behavior.

### D8. `assume_when_axioms` — Go silently skips malformed input
- **Go** `check/l2s_shared.go:380-382`: `if !ok { continue }` — silently skips non-Cond bodies
- **Python**: No type check — would crash on malformed input, surfacing the bug to the user
- **Impact:** Go hides bad user input. If a when binder has a non-Cond body, the user should be told about it, not left wondering why their axiom vanished.
- **Fix:** Replace `continue` with `panic(fmt.Sprintf("assume_when_axioms: when binder %s has non-Cond body type %T", when.Name, when.Body))`. Apply to all three locations where this pattern appears (SharedStep6 when axioms, whenEventsFunc first loop, whenEventsFunc second loop).

---

## LOW SEVERITY — Structural differences unlikely to cause bugs

### D9. `NormalProgram.Postconds` lifecycle
- **Python**: Dynamically added via `hasattr` check (not in `__init__`)
- **Go**: Declared struct field, always present (may be nil)
- **Impact:** Semantic equivalent for l2s (where postconds is nil). Only matters for ranking tactic which is Go-only extension.

### D10. Sort comparison strategy in `BuildAddConstsToD`
- **Python** `ivy_l2s.py:875`: `c.sort == s` — structural equality
- **Go** `check/l2s_shared.go:881`: `sym.CSort.String() == s.String()` — string comparison
- **Impact:** "Equivalent in practice" is a recipe for bugs. `.String()` is the pretty-printer (PrettyFmla) which drops annotations. Use proper structural comparison via NodeKey/Sexp/Canon, consistent with how D12 already correctly uses structural keys.
- **Fix:** Use `lg.SortEqual(sym.CSort, s)` or compare via `sym.CSort.Sexp() == s.Sexp()` (NodeKey-based), not `.String()`.

### D11. `RemoveUnusedDefinitionsGoal` — caller extracts concFmlas
- **Python**: Function internally extracts `conc.model.fmlas + [conc.fmla]` from TemporalModels
- **Go**: Caller passes `concFmlas` as parameter
- **Impact:** Same formulas are used; just different function boundary. Go's l2s.go:863-876 extracts the same formulas.

### D12. Event dedup strategy in `instrStmt` — wrong map key type
- **Python**: Uses Python `set()` (object hash/eq) for `event_props`, `event_whens`, `event_waits`
- **Go**: Uses `map[string]*lg.NamedBinder` with `Sexp()` cast to string as key
- **Impact:** The Sexp comparison is structurally correct, but the key type should be `map[lg.NodeKey]*lg.NamedBinder` not `map[string]*lg.NamedBinder`. `lg.NodeKey` is the project-standard type for structural keys (it IS `[]byte` under the hood, same as Sexp output). Using raw `string` is inconsistent and loses the semantic intent.
- **Fix:** Change `eventProps`, `eventWhens`, `eventWaits` from `map[string]*lg.NamedBinder` to `map[lg.NodeKey]*lg.NamedBinder`. Use `prop.Sexp()` directly as key (already returns `lg.NodeKey`), removing the `string()` cast. File: `check/l2s_shared.go:654-656,698,701,704`

### D13. `symprops`/`symwhens`/`symwaits` key type — wrong map key type
- **Python**: Const objects as defaultdict keys (hash by structure — compares name + sort)
- **Go**: String names as map keys (`map[string][]*lg.NamedBinder`)
- **Impact:** Using bare string names assumes symbol names are unique in the signature, which is true today but fragile. Python's structural equality compares the full Const (name + sort). Go should use `map[lg.NodeKey]` with `Sexp()` keys to match Python's structural comparison, consistent with D12 and the rest of the codebase.
- **Fix:** Change `symprops`, `symwhens`, `symwaits` from `map[string][]*lg.NamedBinder` to `map[lg.NodeKey][]*lg.NamedBinder`. Key by `lg.NodeKey(sym.Sexp())` instead of `c.Name`. Update all lookup sites (instrStmt return monitoring, dependency loop, defaultdict touching). File: `check/l2s_shared.go:416-460,624-690`

---

## CONFIRMED MATCHES — Areas that look identical

The following were verified to be correct ports:

1. **Fair cycle construction** (Python 977-1006, Go 658-717) — relation projection, function projection, sort checks all match
2. **Monitor state machine** (Python 1012-1032, Go 740-768) — waiting→frozen→saved transitions identical  
3. **SharedBuildSaveAndWait** (Python 945-975, Go 301-341) — save_state, done_waiting, reset_w identical
4. **SharedStep6 tableau axioms** — assume_g, assume_init, assume_w all structurally identical
5. **propEventsFunc** — pre/post event generation matches exactly
6. **whenEventsFunc** — l2s_whennext/l2s_whenprev handling matches
7. **waitEventsFunc** — wait event assignment matches
8. **Event concatenation order** — when_pre prepended, when_post/wait_post appended
9. **prefix_action/postfix_action** — same composition semantics
10. **concat_actions** — same single-level Sequence flattening
11. **l2s_auto conditional logic** — all auto2/3/4/5 variant branches match
12. **Desugar $was/$happened** — recursive expansion identical
13. **SharedStep11 named binder replacement** — dedup by Sexp, sort by Sexp, _old_l2s_g consistency
14. **SharedStep12 goal construction** — TemporalModels(model, lg.And()) identical
15. **Idle action composition** — change_monitor_state + assume_g + add_consts_to_d + assert order
16. **L2S init actions** — waiting=true, frozen=false, saved=false + add_consts_to_d + reset_w + assume_g + assume_init + assume(not_lf)
17. **`SymbolsIluAst`** — Go has the unified function at `ivylogic/symbols.go:39`, structurally matching Python's `symbols_ilu_ast`
18. **`RelationSort`** — `il.RelationSort([]lg.Sort{s})` creates `FunctionSort(s, Boolean)`, matching Python's `lg.FunctionSort(sort, lg.Boolean)`
19. **Triple canon sort key** — Go's `l2sGTriple.key()` is byte-identical to Python's `_l2s_g_triple_canon`
20. **defaultdict auto-vivification** — Go correctly touches symprops/symwhens/symwaits maps at l2s_shared.go:682-690

---

## Recommended Investigation Order

### Phase 1: Fix D1-D2 (high severity — SplitReturns)
1. **D1** — Remove the early return in `SplitReturns`. File: `actions/action.go:792-795`
2. **D2** — Port full `.rename()` dispatch to `SplitReturns` so Apply/Variable returns are also renamed, not just Const. File: `actions/action.go:811-816`

### Phase 2: Fix D8, D10, D12, D13 (structural correctness)
3. **D8** — Replace `continue` with panic on non-Cond when binder body. Files: `check/l2s_shared.go:380-382`, `check/l2s_shared.go:540-542`, `check/l2s_shared.go:562-564`
4. **D10** — Replace `sym.CSort.String() == s.String()` with proper structural sort comparison. File: `check/l2s_shared.go:881`
5. **D12** — Change event dedup maps from `map[string]` to `map[lg.NodeKey]`. File: `check/l2s_shared.go:654-656`
6. **D13** — Change symprops/symwhens/symwaits from `map[string]` to `map[lg.NodeKey]`. File: `check/l2s_shared.go:416-460,624-690`

### Phase 3: Fix D4-D6 (caller audits and guards)
7. **D4** — Audit callers of `SymPlaceholders` for nil-vs-empty checks; fix any that use `== nil` instead of `len() > 0`
8. **D5** — Audit callers of `GoalPrems` for nil-vs-empty checks; fix any that use `== nil` instead of `len() > 0`
9. **D6** — Add EnvAction guard to `CopyFormalsTo`. File: `actions/helpers.go:231`

### Phase 4: Run tests
10. Run `cd ~/ivy/goivy && make test` to verify no regressions
11. Compare xtrace output for an l2s test case between Go and Python to verify canon alignment

---

## Verification

After applying fixes:
```bash
cd ~/ivy/goivy && make test
```

For targeted l2s verification, pick an l2s test case from the test suite and compare the full xtrace output between Go and Python, looking for the first divergence point.

---

## Files to Modify

| File | Divergence | Change |
|------|-----------|--------|
| `actions/action.go` | D1, D2 | Fix SplitReturns: remove early return, port full .rename() dispatch for Apply/Variable returns |
| `check/l2s_shared.go` | D8 | Panic on non-Cond when binder body instead of silent `continue` |
| `check/l2s_shared.go` | D10 | Use structural sort comparison (`SortEqual` or Sexp-based) instead of `.String()` |
| `check/l2s_shared.go` | D12 | Change event dedup maps to `map[lg.NodeKey]*lg.NamedBinder` |
| `check/l2s_shared.go` | D13 | Change symprops/symwhens/symwaits to `map[lg.NodeKey][]*lg.NamedBinder` |
| callers of `SymPlaceholders` | D4 | Audit for `== nil` checks, fix to use `len() > 0` |
| callers of `GoalPrems` | D5 | Audit for `== nil` checks, fix to use `len() > 0` |
| `actions/helpers.go` | D6 | Add EnvAction guard to CopyFormalsTo |
