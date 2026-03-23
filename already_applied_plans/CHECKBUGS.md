# Code Review: goivy/check vs pyivy/ivy/ivy_check.py

## Context
The Go `check` package is a mechanical port of Python's `ivy_check.py`. This review identifies bugs and divergences from the Python logic that need fixing to bring Go into conformance.

---

## Bug 1: `MatchHandler.Handle` uses renamed sym instead of original sym
**Files:** `/Users/jaten/goivy/check/helpers.go:255-266`
**Python:** `ivy_check.py:338-353`

Python iterates `for sym, renamed_sym in env.items()` where `sym` is the original symbol (dict key) and `renamed_sym` is what it was renamed to (dict value). It checks `is_new(sym)` and `is_skolem(sym)` on the ORIGINAL symbol. Then calls `show_sym(sym, renamed_sym)`.

Go creates `sym := lg.NewSymbol(renamedSym.Name, renamedSym.CSort)` which makes `sym` identical to `renamedSym`. Then checks `is_new(renamedSym.Name)` and `is_skolem(renamedSym)` on the RENAMED symbol. This is wrong — should check the original symbol from the env key.

**Fix:** Reconstruct the original symbol from `symKey` (the map key), not from `renamedSym`. The env map key contains the original symbol's identity.

---

## Bug 2: `MatchHandler.Handle` missing lineno guard and lineno in output
**Files:** `/Users/jaten/goivy/check/helpers.go:243-269`
**Python:** `ivy_check.py:338-353`

Two issues:
1. Python checks `if hasattr(action, 'lineno')` before doing any trace processing. Go processes all actions unconditionally.
2. Python prints `'{}{}'.format(action.lineno, action)` (lineno + action string). Go prints just `fmt.Sprint(action)` without lineno.

**Fix:** Add lineno guard and include lineno in output format.

---

## Bug 3: `PrettyLineno` format divergence
**Files:** `/Users/jaten/goivy/check/helpers.go:32-40`
**Python:** `ivy_check.py:255-256`

Python `pretty_lineno(ast)` returns `str(ast.lineno)` — just the bare number like `"42"`.
Go `PrettyLineno(lf)` returns `fmt.Sprintf("line %d: ", lf.Lineno)` — adds "line " prefix and ": " suffix.

This changes all verification output. Python output: `"        42property_name"`. Go output: `"        line 42: property_name"`.

**Fix:** Change `PrettyLineno` to return `fmt.Sprintf("%d", lf.Lineno)` to match Python format.

---

## Bug 4: `AllAssertLinenos` missing Ranking subaction check
**Files:** `/Users/jaten/goivy/check/isolate_check.go:1022-1033`
**Python:** `ivy_check.py:836-839`

Python checks `isinstance(sub, (act.AssertAction, act.Ranking))` — includes both AssertAction AND Ranking.
Go only checks `sub.(*actions.AssertAction)` — missing Ranking.

**Fix:** Add `*actions.Ranking` check alongside `*actions.AssertAction`.

---

## Bug 5: `AllAssertLinenos` missing checked_assert filtering
**Files:** `/Users/jaten/goivy/check/isolate_check.go:1017-1043`
**Python:** `ivy_check.py:847-852`

Python at the end of `all_assert_linenos()`:
```python
check_lineno = act.checked_assert.get()
if check_lineno:
    if check_lineno in seen:
        return [check_lineno]
    raise iu.IvyError(None,'There is no assertion at the specified line')
```

Go's `AllAssertLinenos` doesn't filter by the checked_assert parameter and doesn't raise an error if the specified line isn't found.

**Fix:** Add checked_assert filtering and error return at the end of `AllAssertLinenos`.

---

## Bug 6: `CheckIsolate` missing `!unprovable` guard on temporal checks
**Files:** `/Users/jaten/goivy/check/isolate_check.go:529-532`
**Python:** `ivy_check.py:718-719`

Python: `if not unprovable: check_temporals()`
Go: always calls `CheckTemporals(mod)` without the unprovable guard.

**Fix:** Wrap the `CheckTemporals` call with `if !mod.Cfg.OnlyCheckUnprovable`.

---

## Bug 7: `CheckModule` missing TrustedIsolateDef check
**Files:** `/Users/jaten/goivy/check/isolate_check.go:783-786`
**Python:** `ivy_check.py:930`

Python: `if len(idef.verified()) == 0 or isinstance(idef, ivy_ast.TrustedIsolateDef): continue`
Go: only checks `len(idef.Verified()) == 0`, missing the TrustedIsolateDef check.

**Fix:** Add `idef.IsTrusted()` (or type assertion to `*ast.TrustedIsolateDef`) check.

---

## Bug 8: `CheckModule` missing user-specified isolate parameter
**Files:** `/Users/jaten/goivy/check/isolate_check.go:756-770`
**Python:** `ivy_check.py:906-908`

Python reads `ivy_compiler.isolate.get()` to let the user specify a single isolate. If set, only that isolate is checked. Go doesn't support this — it always checks all isolates.

**Fix:** Check `mod.Cfg.Isolate` (or similar field) and if set, use `[]string{isolate}` instead of iterating all.

---

## Bug 9: `CheckModule` missing `complete` logic attribute
**Files:** `/Users/jaten/goivy/check/isolate_check.go:878-885`
**Python:** `ivy_check.py:958-961`

Python in the default (ic) case:
```python
logic = get_isolate_attr(isolate, 'complete', None)
if logic is not None:
    im.module.logics = [logic]
check_isolate()
```

Go's default case doesn't set the logics attribute from the isolate's `complete` attribute.

**Fix:** Add logic attribute lookup and `isoMod.Logics` assignment before calling `CheckIsolate`.

---

## Bug 10: `CheckIsolate` summary mode output lost
**Files:** `/Users/jaten/goivy/check/isolate_check.go:132-165, 242-249`
**Python:** `ivy_check.py:546-606`

Python has separate conditions: the outer `if` prints headers regardless of `check`, and the inner `if check:` controls actual verification. Go merges `check` into the outer `if`, so in summary mode (check=false), headers aren't printed and the `else` branch (which prints properties/schema instances) is never reached.

Affected sections:
- Property checking (lines 546-560 in Python)
- Initialization invariant checking (lines 599-606 in Python)

**Fix:** Separate the `check` condition from the outer `if` to match Python's structure.

---

## Bug 11: `ApplyConjProofs` uses wrong method name
**Files:** `/Users/jaten/goivy/check/check.go:452`
**Python:** `ivy_check.py:461`

Python: `pc.admit_proposition(lf, proof)` — calls `admit_proposition` with 2 args.
Go: `pc.ApplyProof([]*ast.LabeledFormula{astLF}, astProof)` — calls different method name `ApplyProof` with different arg structure.

**Fix:** Change to use `pc.AdmitProposition(astLF, astProof)` to match Python.

---

## Bug 12: `CheckSeparately` lacks tri-state for OptSeparate
**Files:** `/Users/jaten/goivy/check/isolate_check.go:1009-1014`
**Python:** `ivy_check.py:866-869`

Python `opt_separate` is initialized as `BooleanParameter("separate", None)` — three states: None (unset), True, False. Python checks `if opt_separate.get() is not None:` to distinguish "user set it" from "not set". Go uses a plain `bool` which can't distinguish "explicitly set to false" from "not set".

**Fix:** Use `*bool` or a separate `OptSeparateSet bool` field in Config.

---

## Bug 13: `CheckIsolate` printing format for implementations/monitors
**Files:** `/Users/jaten/goivy/check/isolate_check.go:209-211`
**Python:** `ivy_check.py:582`

Python: `"{}implementation of {}".format(pretty_lineno(action), mixee)` — includes line number.
Go: `"implementation of %s"` — missing line number.

**Fix:** Add `prettyActionLineno` to the implementation/monitor print statements.

---

## Bug 14: `CheckIsolate` printing format for initializers
**Files:** `/Users/jaten/goivy/check/isolate_check.go:231-233`
**Python:** `ivy_check.py:596-597`

Python: `"{}{}".format(pretty_lineno(action), actname)` — prints lineno + name.
Go: just prints name.

**Fix:** Add lineno to initializer print output.

---

## Priority Order for Fixes

**High (semantic bugs):**
1. Bug 1: MatchHandler.Handle wrong symbol
2. Bug 4: AllAssertLinenos missing Ranking
3. Bug 5: AllAssertLinenos missing checked_assert filter
4. Bug 6: Missing unprovable guard on temporals
5. Bug 7: Missing TrustedIsolateDef check
6. Bug 11: ApplyConjProofs wrong method

**Medium (behavioral divergence):**
7. Bug 10: Summary mode output lost
8. Bug 2: Handle missing lineno guard/output
9. Bug 3: PrettyLineno format
10. Bug 8: Missing user-specified isolate
11. Bug 9: Missing complete/logics attribute

**Low (cosmetic / edge case):**
12. Bug 12: CheckSeparately tri-state
13. Bug 13: Implementation/monitor lineno
14. Bug 14: Initializer lineno

---

## Verification

After fixes, run:
```bash
cd /Users/jaten/goivy && go build ./check/...
cd /Users/jaten/goivy && go test ./check/... -v
```

Compare output format of any available `.ivy` test files against the Python version's output.
