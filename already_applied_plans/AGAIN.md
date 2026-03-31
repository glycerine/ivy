# Plan: Audit of 14 CRITICAL Fixes → FIX14_AGAIN.md

**Created:** 2026-03-25T21:30

## Context

The 14 CRITICAL fixes just applied need auditing against the Python source. The user specifically called out that omitting version guards is NOT acceptable.

## Divergences Found

### FIX #1 — `p_top_assert_symbol_arrow_assert_rhs` — **VERSION GUARD CRITICAL**

Python guard: `if iu.get_numeric_version() <= [1,6]:` — this rule only fires for v1.6 and below.
Go: Has the rule unconditionally with the body we just added.
**Problem:** Go grammar is v1.7+. The rule body should be EMPTY (or guarded) for v1.7. The previous empty body was correct! Our fix incorrectly added logic that should never execute for v1.7.
**Fix:** Revert to empty body, add comment explaining the v1.6 guard.

### FIX #2 — `p_top_nativequote` — **WRONG TYPE: NewAtom vs NativeCode**

Python: `NativeDef(*([mk_label(None,'native')] + [NativeCode(text)] + bqs))`
Go: `defnArgs := append([]ast.Node{label, acfg(v17lex).NewAtom(text)}, bqs...)`
**Problem:** Go wraps the text in `NewAtom(text)` but Python uses `NativeCode(text)` — a dedicated type with `.code` field. Go has `ast.NativeCode` and `NewNativeCode(code string)`.
**Fix:** Change `acfg(v17lex).NewAtom(text)` → `acfg(v17lex).NewNativeCode(text)`

### FIX #3 — `p_lit_atom` — **OK**

Python: `Literal(1, p[1])` with `p[0].lineno = get_lineno(p,1)`
Go: `NewLiteral(1, $1)` with `SetLineno(nodeLineno($1))`
Correct. `nodeLineno` produces the trace and extracts lineno from the atom node, matching Python's `get_lineno(p,1)` (position 1 = the atom).

### FIX #4 — `p_lit_term_eq_term` — **OK**

Python: `Literal(1, Atom(p[2], [symbol(p[1]), symbol(p[3])]))` + `get_lineno(p,2)`
Go: `NewLiteral(1, NewAtom("=", NewAtom($1.Val), NewAtom($3.Val)))` + `tokLineno(..., $2)`
Correct. Python's `symbol()` creates a Symbol from a string; Go uses `NewAtom`. The Atom with rep "=" and two symbol args matches.

### FIX #5 — `p_lit_term_tildaeq_term` — **OK**

Python: `Literal(0, Atom(...))` — polarity 0 for negation.
Go: `NewLiteral(0, ...)` — matches.

### FIX #5b — `p_lit_tilda_atom` — **OK**

Python: `~p[2]` flips Literal polarity via `__invert__`.
Go: checks if `*ast.Literal`, flips polarity with `1 - lit.Polarity`. Falls back to `NewLiteral(0, $2)` if not a Literal. Correct.

### FIX #6 — `p_fmla_fmla_isa_atype` — **VERSION GUARD OK, LOGIC OK**

Python guard: `if not (iu.get_numeric_version() <= [1,6]):` — v1.7+ only. Go grammar is v1.7+, so correct.
Python: `tp = Atom(p[3], [])` — wraps atype string in Atom with empty args.
Go: `atypeToAtom(acfg(v17lex), $3)` — converts atype Node to Atom. Correct.
Both set 2 linenos. OK.

### FIX #7 — `p_term_namedbinder_vars_dot_term` — **MINOR: NewApp first arg semantics**

Python: `App(x, p[9])` where `x` is the NamedBinder. Python's `App.__init__(funsym, *terms)` sets `self.rep = funsym` (the NamedBinder) and `self.args = terms`.
Go: `NewApp(binder, $9...)` — sets `Rep = binder` and `Terms = $9`. This matches.
**OK.**

### FIX #8 — `p_term_term_colon_term` — **VERSION GUARD OK, LOGIC OK**

Python guard: v1.7+ variant (under else after `> [1,2]`, then under else after `> [1,6]`). Go grammar is v1.7+, so correct.
Python: `hasattr(p[1], "sort")` checks if sort attribute exists. For Variable (always set in `__init__`), this is True → error on double-annotation. For App (not set in `__init__`), this is False → allows first annotation.
Go: Checks `VSort != ""` for Variable (set at construction → catches double), `ASort != nil` for Atom/App (nil at construction → allows first). Equivalent behavior.
**OK.**

### FIX #9 — `p_objectend` — **POSSIBLE ISSUE: isObject never set to True**

Python: `stack[-1].is_object = False`
Go: `accum.isObject = false`
**Question:** Where is `isObject` set to `true` in Go? In Python, `is_object` is dynamically set to `True` on the accumulator somewhere during object parsing. If Go never sets it to `true`, setting it to `false` is a no-op.
**Verdict:** The field was added but may never be set to true. Low risk — the fix is harmless but incomplete if there's a corresponding `isObject = true` missing elsewhere.

### FIX #10 — `p_proofstep_symbol` — **OK**

Changed `tokLineno(..., $1)` → `nodeLineno($2)` for the atom's lineno.
Python: `a.lineno = get_lineno(p,2)` — position 2 (atype).
Go: `a.SetLineno(nodeLineno($2))` — extracts lineno from atype node. Correct.
`si.SetLineno(tokLineno(..., $1))` stays at $1 (APPLY token). Matches Python `p[0].lineno = get_lineno(p,1)`.

### FIX #11 — `p_proofstep_symbol_with_defns` — **OK**

Same fix as #10. Correct.

### FIX #12 — `p_simpleact_debug_symbol_optdebugargs` — **VERSION GUARD OK, LOGIC OK**

Python guard: v1.7+ (under `else:` after `> [1,6]`). Go grammar is v1.7+. Correct.
Python: `if not p[2].startswith('"'): report_error(...)` — validates debug string.
Go: `if !strings.HasPrefix($2.Val, "\"")` — same check. Correct.
Both set 2 linenos on action atom and DebugAction. OK.

### FIX #13 — `p_top_implement_type_symbol_with_symbol` — **VERSION GUARD OK, LOGIC OK**

Python guard: v1.7+ (under `else:` after `> [1,6]`). Correct.
Python: 4 lineno calls. Go now has 4 tokLineno calls on a1, a2, impl, d. Correct.
Python: `mk_lf(impl)` wraps in LabeledFormula. Go: `mkLF(acfg(v17lex), impl)`. Correct.
Python: `ImplementTypeDecl(mk_lf(impl))`. Go: `NewImplementTypeDecl(mkLF(...))`. Correct.

### FIX #14 — `p_tacticwithelem_trigger` — **VERSION GUARD OK, MINOR TYPE ISSUE**

Python guard: v1.7+ (under `else:` after `> [1,6]`). Correct.
Python: `Trigger(*([Atom(p[2])]+p[4]))` — p[2] is an atype string, wrapped in `Atom()`.
Go: `atypeToAtom(acfg(v17lex), $2)` — converts atype Node to Atom. Correct.
Python: `Trigger(...)` constructor. Go: `&ast.Trigger{Terms: ...}` — struct init, not config constructor.
**Minor:** Should use config constructor if one exists, but Trigger doesn't have one (`NewTrigger` not found). Acceptable.

## Summary of Required Fixes

| # | Fix | Issue | Severity |
|---|-----|-------|----------|
| 1 | `p_top_assert_symbol_arrow_assert_rhs` | **Revert** — v1.6-only rule, body should be empty for v1.7 | CRITICAL |
| 2 | `p_top_nativequote` | Change `NewAtom(text)` → `NewNativeCode(text)` | MEDIUM |
| 9 | `p_objectend` | `isObject` never set to `true` — check if companion setter needed | LOW |
| rest | All other fixes | OK — no divergences found | — |

## Files to Modify

| File | Change |
|------|--------|
| `lalr_full/grammar_v17.y:1660-1667` | Revert fix #1: empty body for v1.6-only rule |
| `lalr_full/grammar_v17.y:1613` | Fix #2: `NewAtom(text)` → `NewNativeCode(text)` |
| `lalr_full/grammar_v17.go` | Regenerate |

## Verification

```bash
cd ~/goivy && go build ./... && make golden
```
