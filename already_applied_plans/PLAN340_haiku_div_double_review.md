opus' review of haiku ivy/goivy divergences:

┌───────────────────────┬──────────────────────────────┐
│        Verdict        │             DIVs             │
├───────────────────────┼──────────────────────────────┤
│ Real bugs             │ DIV-5, DIV-9, DIV-11, DIV-12 │
├───────────────────────┼──────────────────────────────┤
│ Real but latent/minor │ DIV-2, DIV-8, DIV-10, DIV-14 │
├───────────────────────┼──────────────────────────────┤
│ Cosmetic only         │ DIV-4, DIV-7                 │
├───────────────────────┼──────────────────────────────┤
│ REJECTED (wrong)      │ DIV-3, DIV-6, DIV-13, DIV-15 │
└───────────────────────┴──────────────────────────────┘


haiku produced:

# Conformance Review: Go vs Python Divergences

Created: 2026-04-28 UTC 

---

## Context

This is a read-only audit of known remaining divergences between the Go port (~/ivy/goivy) and the Python source of truth (~/ivy/pyivy/ivy/ivy/). We have already fixed ~1000 divergences. This document catalogs the remaining ones found in this review pass, so they can be fixed systematically.

The Python code is the specification. Every behavioral difference is a bug in the Go port.

---

## Divergences Found

### AREA 1: logicutil / logic packages

---


#### DIV-2: `SubstituteByName` — Lambda incorrectly blocks substitution

**Python** `ivy_logic_utils.py:178`:
```python
if is_quantifier(ast):   # is_quantifier = ForAll | Exists ONLY
    bounds = set(x.name for x in quantifier_vars(ast))
    subs = dict(...)
return ast.clone(substitute_ast(x, subs) for x in ast.args)
```
Lambda is NOT a quantifier. Substitution flows unchanged through Lambda.

**Go** `logicutil/logic_utils.go:649-652`:
```go
case *logic.Lambda:
    newsubs := removeBoundNames(subs, t.Variables)   // WRONG: blocks substitution
    body := substituteByNameRec(t.Body, newsubs)
```
**Fix**: Remove the `case *logic.Lambda:` shadowing block. Lambda should pass `subs` unchanged into its body (matching Python `ivy_logic_utils.substitute_ast`).

---

#### DIV-3: `NormalizeQuantifiers` — Extra distribution logic not in Python

**Python** `logic_util.py:239-271`: Does NOT distribute `ForAll` into `And` or `Exists` into `Or`. Restricts quantifier vars using FVs of the ORIGINAL (un-normalized) body.

**Go** `logicutil/logic_utils.go:1669-1713`: Adds distribution (`ForAll(vars, And(a,b))` → `And(ForAll,ForAll)`), and restricts vars using FVs of the NORMALIZED body.

**Action**: Go `NormalizeQuantifiers` is doing more than the Python equivalent. This may be intentional (optimization), but needs audit vs every Python call site for `normalize_quantifiers`. If any call site relies on the exact structural form, this is a bug.

---

#### DIV-4: `SubstituteApply` — Missing free-variable safety assertion

**Python** `logic_util.py:222-224`:
```python
assert fvr <= fvt, "New free variables!? {}, {}".format(fvr, fvt)
```

**Go** `logicutil/logic_utils.go:1745-1752`: No such assertion.

**Action**: Add the assertion (or equivalent Go panic/error) after applying a substitution function in `substituteApplyRec`.

---

#### DIV-5: `variables_ast` / `FreeVariables` — `Some` binder not correctly handled

**Python** `ivy_logic_utils.py:523-535`: For `Some`, excludes bound params from yielded variables, recurses into formula and arms only.

**Go** `logicutil/logicutil.go:641-678`: `Some` hits `default: t.Children()`, which includes params — params are NOT excluded from the free variable set.

**Action**: Add an explicit `case *ivylogic.Some:` (or equivalent) in `variablesAstRec`, `freeVariablesRec`, `boundVariablesRec`, and all related traversal functions. The bound params must be excluded.

---

#### DIV-6: `rename_ast` — Not present in Go logicutil

**Python** `ivy_logic_utils.py:196-207`: Renames symbols (not variables) via a `subs` map over `ast.rep`.

**Go**: No `RenameAst` function found in `logicutil/`, `ivylogic/`, or `logicparser/`. Used in Python as `rename_ast`, `rename_clause`, `rename_clauses`.

**Action**: Verify whether the Go code handles renaming in a different location or if this is missing functionality.

---

#### DIV-7: `NormalizeQuantifiers` — Lambda handling differs (Python crashes, Go silently passes)

**Python** `logic_util.py:271`: `assert False, type(t)` — crashes on Lambda.

**Go** `logicutil/logic_utils.go:1716-1717`: Returns Lambda unchanged.

**Action**: Either match Python's crash (panic) on Lambda or confirm this is intentional. If callers never pass Lambda in Python context, the Go silent-pass is functionally equivalent but hides misuse.

---

#### DIV-8: `CloseEPR` — Variable ordering in generated ForAll

**Python** `ivy_logic_utils.py:123-131`: Variables in DFS first-occurrence order.

**Go** `logicutil/logic_utils.go:439-443`: Variables in sorted alphabetical order.

**Action**: For canon comparison, these produce different s-expressions. If callers compare canonical forms, this is a bug. Verify whether any test vectors depend on variable ordering in `CloseEPR`.

---

### AREA 2: actions / compiler / isolate packages

---

#### DIV-9 (CRITICAL): `WhileAction.Expand` — SubgoalAction filtering and AssertToAssume not applied

**Python** `ivy_actions.py:1069-1070`:
```python
assumes = [a.assert_to_assume([AssertAction]) for a in asserts if not isinstance(a, SubgoalAction)]
asserts  = [a for a in asserts if not isinstance(a, AssumeAction)]
```
- `assumes` list: only non-SubgoalAction invariants, converted via `assert_to_assume`
- `asserts` list: filtered to exclude AssumeAction invariants
- SubgoalActions are excluded from BOTH lists

**Go** `actions/update.go:1688-1698`:
```go
for _, inv := range invariants {
    asserts = append(asserts, NewAssertAction(inv))
}
for _, inv := range invariants {
    assumes = append(assumes, NewAssumeAction(inv))
}
```
All invariants unconditionally go into both lists. No `AssertToAssume`, no SubgoalAction filtering, no AssumeAction filtering.

**Critical files**: `actions/update.go:1688-1698`

---

#### DIV-10: `WhileAction.Expand` — lineno not assigned to generated HavocActions

**Python** `ivy_actions.py:1083-1085`:
```python
for h in havocs:
    h.lineno = self.lineno
```

**Go** `actions/update.go:1700-1703`: No lineno assignment on generated havocs.

**Critical files**: `actions/update.go:1700-1703`

---

#### DIV-11 (CRITICAL): `ApplyMixin` — IvyError replaced by silent stderr warning

**Python** `ivy_actions.py:1496-1503`: Parameter count/sort mismatches raise `IvyError` with the mixin name.

**Go** `actions/helpers.go:153-178`: Mismatches emit a warning to stderr and silently return the unmodified action. Mixin wiring failures are invisible.

**Critical files**: `actions/helpers.go:153-178`

---

#### DIV-12 (CRITICAL): `InstantiateAction` missing from `IntUpdate` dispatch

**Python** `ivy_actions.py:804`: `InstantiateAction.int_update` is called via normal Python dispatch.

**Go** `actions/update.go:1176-1244`: Switch has no `case *InstantiateAction:`. Hits `default: return NullUpdate()`. The `IntUpdate` method defined on `*InstantiateAction` in `actions/action.go:2198-2332` is dead code when called through the dispatch switch. Any verification step relying on `IntUpdate` for schema instantiation gets an empty update.

**Critical files**: `actions/update.go:1176-1244`, `actions/action.go:2198-2332`

---

#### DIV-13: `AddMixins` — `DropInvariants` not called in compile mode

**Python** `ivy_isolate.py:68-69`: When `create_imports` is true, calls `res.drop_invariants()`.

**Go** `isolate/isolate.go:113-120`: The base `AddMixins` never calls `DropInvariants`. Only `AddMixinsExt` in `isolate/helpers.go:52-54` does this. Verify all call sites use the right variant.

**Critical files**: `isolate/isolate.go:113-120`, `isolate/helpers.go:52-54`

---

#### DIV-14: `AssertAction.ActionUpdate` — `checked_assert` comparison ignores file

**Python** `ivy_actions.py:382-389`: Compares full `Location` object (file + line).

**Go** `actions/update.go:374-389`: Compares only `fmt.Sprintf("%d", lineno.Line)` — file component ignored. Two different-file asserts at the same line number are treated as equal.

**Critical files**: `actions/update.go:374-389`

---

#### DIV-15: `HavocAction.ActionUpdate` — node-type branch test replaced by sort test

**Python** `ivy_actions.py:723-729`: Branch on `is_atom(lhs)` (AST node type); reconstructs with `type(lhs)(new_n, *vs)` preserving the original node type.

**Go** `actions/update.go:876-908`: Branch on `il.IsBoolean(sym) || il.IsRelationalSort(...)` (sort test); always reconstructs as `Apply` node regardless of original type.

**Critical files**: `actions/update.go:876-908`

---

## Prioritization

| Priority | DIV # | Description |
|----------|--------|-------------|
| P0 | DIV-12 | InstantiateAction dead code in IntUpdate dispatch |
| P0 | DIV-9  | WhileAction.Expand: SubgoalAction/AssumeAction filtering missing |
| P0 | DIV-11 | ApplyMixin: IvyError → silent skip |
| P1 | DIV-2  | SubstituteByName Lambda shadowing (wrong) |
| P1 | DIV-5  | Some binder not excluded from FreeVariables traversals |
| P1 | DIV-15 | HavocAction: node-type branch vs sort test |
| P1 | DIV-14 | checked_assert file+line vs line-only |
| P1 | DIV-13 | AddMixins missing DropInvariants |
| P2 | DIV-4  | SubstituteApply missing safety assertion |
| P2 | DIV-10 | WhileAction.Expand: lineno not set on havocs |
| P2 | DIV-6  | rename_ast missing from Go logicutil |
| P3 | DIV-3  | NormalizeQuantifiers: extra distribution (may be intentional) |
| P3 | DIV-7  | NormalizeQuantifiers: Lambda crash vs silent pass |
| P3 | DIV-8  | CloseEPR: variable ordering difference |

---

## Verification

For each fix:
1. Run `cd ~/ivy/goivy && make test` — DO NOT use `go test ./...`
2. Check that XTRACE golden logs remain unchanged
3. Specifically run the action update and logic substitution test vectors

---

## Critical Files to Modify (When Implementing)

- `actions/update.go` — DIV-9, DIV-10, DIV-12, DIV-14, DIV-15
- `actions/helpers.go` — DIV-11
- `logicutil/logic_utils.go` — DIV-2, DIV-3, DIV-4, DIV-6, DIV-7, DIV-8
- `logicutil/logicutil.go` — DIV-5
- `isolate/isolate.go` — DIV-13

---
then opus 4.6 reviewd:

DIV-2: SubstituteByName Lambda — CONFIRMED but likely latent. Go removes Lambda's bound names
 from subs; Python does not (Lambda is not in is_quantifier). Real divergence, but only
triggers if SubstituteByName processes a Lambda whose variables collide with the substitution
 map.

DIV-3: NormalizeQuantifiers distribution — REJECTED. The agent confirmed both Go and Python
perform the same ForAll(vars, And(a,b)) → And(ForAll,ForAll) distribution. The original claim
 was simply wrong.

DIV-4: SubstituteApply assertion — CONFIRMED, cosmetic. Python has assert fvr <= fvt; Go
doesn't. This is a debug-only assertion (assert is stripped with -O). No behavioral
difference — just a missing safety net.

DIV-5: variables_ast Some binder — CONFIRMED. Go has no case *ivylogic.Some: in
variablesAstRec. Python's is_binder() includes Some, protecting its bound params. Some is
actively constructed in compiler/ and ivylogic/, so this can trigger in practice.

DIV-6: rename_ast missing — REJECTED. Go has RenameASTByName and RenameClausesByName in
module/astutil.go:226-248. The agent missed these because it searched for the exact Python
name. The functionality exists under a different name.

DIV-7: NormalizeQuantifiers Lambda crash — CONFIRMED, irrelevant. Python crashes (assert
False) on Lambda; Go returns it unchanged. But the Python crash proves Lambda never reaches
NormalizeQuantifiers in practice. Go's silent pass is functionally equivalent to "never
happens."

DIV-8: CloseEPR variable ordering — CONFIRMED, claim is backwards. Go preserves DFS insertion
 order (via Omap). Python uses gen_to_set(variables_ast) which returns an unordered set. The
claim said Go sorts alphabetically — that's wrong. There IS a divergence, but the direction
is opposite.

DIV-9: WhileAction.Expand filtering — CONFIRMED. Python filters SubgoalAction from assumes
and AssumeAction from asserts (lines 1069-1070). Go unconditionally appends all invariants to
 both lists. Real behavioral divergence.

DIV-10: WhileAction.Expand lineno on havocs — CONFIRMED, minor. Python sets h.lineno =
self.lineno on each havoc. Go doesn't. Real but low-impact — lineno affects error reporting,
not computation.

DIV-11: ApplyMixin error handling — CONFIRMED. Python raises IvyError on param count/sort
mismatch. Go prints to stderr and silently returns the unmodified action. A mixin failure
that Python would halt on, Go swallows.

DIV-12: InstantiateAction dispatch — CONFIRMED. The top-level IntUpdate(action, ctx) function
 in update.go:1170 has no case *InstantiateAction:. It hits default: → NullUpdate(). The
method InstantiateAction.IntUpdate(ctx) exists at action.go:2234 and is designed to compile
macro expansions (per ivy_compile.go:265), but is dead code via the switch dispatch.
InstantiateAction IS constructed in compiler/action.go:429.

DIV-13: AddMixins DropInvariants — REJECTED. All production call sites in isolate/isolate.go
(lines 558, 566, 579, 581) use AddMixinsExt, which DOES call DropInvariants. The bare
AddMixins (which doesn't) is unused in production.

DIV-14: checked_assert comparison — CONFIRMED. Python's checked_assert parameter processes
input as Location(file+'.ivy', line) and compares the full location. Go compares only
fmt.Sprintf("%d", lineno.Line). Two asserts at the same line in different files would be
treated as equal in Go but distinct in Python.

DIV-15: HavocAction branch/reconstruction — REJECTED (branch), UNCLEAR (reconstruction). The
branch conditions are semantically equivalent: Python's is_atom(lhs) checks if the
expression's sort is Boolean, Go's IsRelationalSort(sym.CSort) checks if the symbol's range
sort is Boolean — same result. The reconstruction difference (Python preserves type(lhs), Go
always uses Apply) is likely inconsequential since both represent function application.

---
Summary scorecard:

┌───────────────────────┬──────────────────────────────┐
│        Verdict        │             DIVs             │
├───────────────────────┼──────────────────────────────┤
│ Real bugs             │ DIV-5, DIV-9, DIV-11, DIV-12 │
├───────────────────────┼──────────────────────────────┤
│ Real but latent/minor │ DIV-2, DIV-8, DIV-10, DIV-14 │
├───────────────────────┼──────────────────────────────┤
│ Cosmetic only         │ DIV-4, DIV-7                 │
├───────────────────────┼──────────────────────────────┤
│ REJECTED (wrong)      │ DIV-3, DIV-6, DIV-13, DIV-15 │
└───────────────────────┴──────────────────────────────┘

The four I'd prioritize: DIV-12 (dead code in dispatch — easy one-line fix), DIV-9
(WhileAction filtering), DIV-11 (silent mixin errors), and DIV-5 (Some binder leaking
variables).

---------
