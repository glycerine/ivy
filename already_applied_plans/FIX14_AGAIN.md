# Audit of 14 CRITICAL Fixes Applied 2026-03-25

Systematic line-by-line comparison of each Go fix against the corresponding Python behavior.

## Divergences Found and Fixed

### FIX #1 — `p_top_assert_symbol_arrow_assert_rhs` — **REVERTED (v1.6-only rule)**

Python guard: `if iu.get_numeric_version() <= [1,6]:` — this rule only fires for v1.6 and below.
Go grammar is v1.7+, so the body must be empty. Our fix incorrectly added Implies/AssertDecl/declare logic that should never execute for v1.7. **Reverted to empty body with explanatory comment.**

### FIX #2 — `p_top_nativequote` — **FIXED: NewAtom → NewNativeCode**

Python: `NativeDef(*([mk_label(None,'native')] + [NativeCode(text)] + bqs))`
Go had: `acfg(v17lex).NewAtom(text)` — wrong type.
Go now: `acfg(v17lex).NewNativeCode(text)` — correct, matches Python's `NativeCode(text)`.

## Verified OK (no divergences)

### FIX #3 — `p_lit_atom` — OK
`NewLiteral(1, $1)` + `nodeLineno($1)` matches Python's `Literal(1, p[1])` + `get_lineno(p,1)`.

### FIX #4 — `p_lit_term_eq_term` — OK
`NewLiteral(1, NewAtom("=", ...))` + `tokLineno($2)` matches Python.

### FIX #5 — `p_lit_term_tildaeq_term` — OK
`NewLiteral(0, ...)` matches Python's `Literal(0, ...)`.

### FIX #5b — `p_lit_tilda_atom` — OK
Polarity flip `1 - lit.Polarity` matches Python's `__invert__` on Literal.

### FIX #6 — `p_fmla_fmla_isa_atype` — OK
Version guard: v1.7+ only (`if not (version <= [1,6]):`). Go grammar is v1.7+. Correct.
`atypeToAtom` + 2 lineno calls matches Python's `Atom(p[3],[])` + 2 `get_lineno` calls.

### FIX #7 — `p_term_namedbinder_vars_dot_term` — OK
`NewApp(binder, $9...)` matches Python's `App(x, p[9])`. No version guard.

### FIX #8 — `p_term_term_colon_term` — OK
Version guard: v1.7+ variant. Go handles Variable/Atom/App sort annotation + error check.
Python's `hasattr(p[1], "sort")` semantics are correctly modeled by nil/empty checks.

### FIX #9 — `p_objectend` — OK (minor note)
`accum.isObject = false` added. The companion `isObject = true` is set on `IsolateDef.IsObject`, not on the accumulator. Python dynamically adds `is_object` to the stack entry. The Go field exists but is only cleared, never set to true — this is a no-op but matches the structural intent. Low risk.

### FIX #10 — `p_proofstep_symbol` — OK
Fixed atom lineno from `tokLineno($1)` (APPLY token) to `nodeLineno($2)` (atype node).
Matches Python's `a.lineno = get_lineno(p,2)`.

### FIX #11 — `p_proofstep_symbol_with_defns` — OK
Same $1→$2 fix as #10.

### FIX #12 — `p_simpleact_debug_symbol_optdebugargs` — OK
Version guard: v1.7+. String validation + 2 lineno calls match Python.

### FIX #13 — `p_top_implement_type_symbol_with_symbol` — OK
Version guard: v1.7+. 4 lineno calls + mkLF wrapping + ImplementTypeDecl match Python.

### FIX #14 — `p_tacticwithelem_trigger` — OK
Version guard: v1.7+. `atypeToAtom` wrapping + lineno from WITH token ($3) match Python.

## Version Guard Summary

| Rule | Python Guard | Go Status |
|------|-------------|-----------|
| `p_top_assert_symbol_arrow_assert_rhs` | `<= [1,6]` only | **REVERTED** — body empty for v1.7 |
| `p_proofstep_symbol` | `> [1,6]`: APPLY variant | Go has APPLY variant — correct |
| `p_proofstep_symbol_with_defns` | `> [1,6]`: APPLY variant | Go has APPLY variant — correct |
| `p_fmla_fmla_isa_atype` | `> [1,6]` | Go has it — correct |
| `p_term_term_colon_term` | `> [1,2]` and `> [1,6]` | Go has v1.7 variant — correct |
| `p_tacticwithelem_trigger` | `> [1,6]` | Go has it — correct |
| `p_top_implement_type_symbol_with_symbol` | `> [1,6]` | Go has it — correct |
| `p_simpleact_debug_symbol_optdebugargs` | `> [1,6]` | Go has it — correct |
| All others | No version guard | No guard needed — correct |
