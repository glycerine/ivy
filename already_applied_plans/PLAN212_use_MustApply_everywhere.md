# PLAN210: Eliminate all raw `&lg.Apply{}` struct literals — use `MustApply` everywhere

**Created:** 2026-04-07 ~01:00 UTC

## Context

Python's `Apply` constructor (`logic.py:147-183`) validates sort/arity at construction time via `_preprocess_`. Bad arity, wrong sorts, or non-function application raise `SortError` — there is NO silent fallback. The only exemption is `TopSort` (polymorphic), which skips all checks.

Go's `NewApply` (`logic/term.go:139-189`) is a faithful port of this validation. But throughout the Go codebase, ~80 call sites bypass `NewApply` by constructing raw `&lg.Apply{Func: ..., Terms: ...}` struct literals. This:

1. **Loses `aSort`** — the unexported cached sort field is nil (zero value), causing downstream bugs like the PLAN209 `Iff`-vs-`Eq` divergence
2. **Hides caller bugs** — several sites use a "try `NewApply`, catch error → fall back to raw struct" anti-pattern that silently swallows sort/arity errors Python would have crashed on

Three critical-path sites were already fixed in PLAN209. This plan fixes ALL remaining sites.

## Approach

### Step 1: Add `MustApply` to `logic/term.go`

```go
// MustApply is like NewApply but panics on error.
// Matches Python's Apply constructor which raises SortError on bad input.
func MustApply(fn Expr, terms ...Expr) *Apply {
    a, err := NewApply(fn, terms...)
    if err != nil {
        panic(fmt.Sprintf("MustApply: %v", err))
    }
    return a
}
```

### Step 2: Replace all raw `&lg.Apply{}` with `lg.MustApply()`

Every `&lg.Apply{Func: F, Terms: T}` becomes `lg.MustApply(F, T...)`.
Every "try NewApply, fallback to raw struct" becomes just `lg.MustApply(...)`.

**Intentional bypass exceptions** (keep raw struct with comment):
- `compiler/solo7_test.go:114` — tests wrong-arity detection, needs raw struct
- `actions/audit53_test.go:344-345` — tests partial application bypass
- Any other test that deliberately constructs an invalid Apply to test error paths

### Step 3: Simplify the three PLAN209 sites

The three sites already fixed in PLAN209 used 4-line `NewApply` + panic blocks. Replace with 1-line `MustApply`.

## File-by-file changes

### `logic/term.go` — ADD `MustApply`
Add after `NewApply` function.

### Production code (non-test files):

| File | Lines | Pattern | Change |
|------|-------|---------|--------|
| `ivylogic/constructors.go` | 25, 48 | Direct struct | `MustApply` |
| `ivylogic/classify_ext.go` | 319, 329, 337, 403, 404, 417, 418 | Direct struct | `MustApply` |
| `ivylogic/util.go` | 310, 311, 342, 343, 643 | Direct struct | `MustApply` |
| `ivylogic/formula.go` | 216 | Direct struct | `MustApply` |
| `compiler/action.go` | 701, 933, 1037 | 933 is fallback pattern | `MustApply` (remove fallback) |
| `actions/action.go` | 792 | Fallback pattern | `MustApply` (remove fallback) |
| `actions/match.go` | 450, 457 | Direct struct | `MustApply` |
| `actions/update.go` | 433, 901, 1590, 1592 | 901 is fallback | `MustApply` |
| `solver/unitres_bridge.go` | 56, 64 | Direct struct | `MustApply` |
| `solver/herbrand.go` | 444, 637, 668 | Direct struct | `MustApply` |
| `solver/z3convert.go` | 146, 370 | Direct struct | `MustApply` |
| `solver/compat.go` | 412, 424, 434 | Direct struct | `MustApply` |
| `solver/solver.go` | 799 | Direct struct | `MustApply` |
| `proof/match.go` | 388, 398 | Direct struct | `MustApply` |
| `proof/tactics.go` | 454 | Direct struct | `MustApply` |
| `alpha/alpha.go` | 863, 983 | Direct struct | `MustApply` |
| `conceptspace/cs_parser.go` | 520 | Direct struct (grammar) | `MustApply` |
| `conceptspace/conceptspace.go` | 275 | Direct struct | `MustApply` |
| `l2s/l2s.go` | 107, 118 | Fallback pattern | `MustApply` (remove fallback) |
| `logicutil/logic_utils.go` | 498, 925, 978, 1029, 1036, 1047, 1129, 1180, 1449, 1631 | Fallback pattern (8 sites) + direct (2) | `MustApply` (remove all fallbacks) |
| `mc/transforms.go` | 223 | Direct struct | `MustApply` |
| `module/module.go` | 794 | Direct struct | `MustApply` |
| `module/theory.go` | 455, 456, 472, 473 | Direct struct | `MustApply` |
| `module/canonize.go` | 163 | Direct struct | `MustApply` |
| `module/astutil.go` | 183 | Already NewApply+panic (PLAN209) | Simplify to `MustApply` |
| `ivylogic/util.go` | 23 | Already NewApply+panic (PLAN209) | Simplify to `MustApply` |
| `actions/transrel.go` | 1329 | Already NewApply+panic (PLAN209) | Simplify to `MustApply` |
| `autoinst/autoinst.go` | 154 | Direct struct | `MustApply` |

### Test code:

| File | Lines | Change |
|------|-------|--------|
| `logic/sexp_cross_test.go` | 59 | `MustApply` |
| `solver/unitres_bridge.go` | 56, 64 | `MustApply` |
| `solver/solver_test.go` | 967, 980, 996 | `MustApply` |
| `solver/solver2_test.go` | 29, 69, 111, 143 | `MustApply` |
| `solver/solver2_fuzz_test.go` | 492 | `MustApply` |
| `solver/herbrand_unitres_test.go` | 21, 105 | `MustApply` |
| `compiler/batch_f_test.go` | 904 | `MustApply` |
| `compiler/solo7_test.go` | 114 | **KEEP RAW** (intentional arity bypass) |
| `ivylogic/varuniq_test.go` | 183 | `MustApply` |
| `ivylogic/polarity_equiv_test.go` | 75 | `MustApply` |
| `ivylogic/symbols_test.go` | 28 | `MustApply` |
| `ivylogic/nodeargs_test.go` | 23, 45, 237 | `MustApply` |
| `ivylogic/new_funcs_test.go` | 257, 271, 286, 293, 307, 316, 365, 427 | `MustApply` |
| `isolate/isolate_symbols_test.go` | 22 | `MustApply` |
| `actions/audit51_test.go` | 484, 509, 523 | `MustApply` |
| `actions/audit53_test.go` | 344, 345 | **KEEP RAW** (intentional bypass) |
| `actions/update_test.go` | 104 | `MustApply` |
| `actions/macro_test.go` | 796 | `MustApply` |
| `actions/stub_fixes_test.go` | 230, 264, 299, 412, 454, 494 | `MustApply` |
| `proof/proof_test.go` | 292, 293, 593, 619, 641 | `MustApply` |
| `module/clauseops_test.go` | 403 | `MustApply` |
| `z3bridge/translate2_test.go` | 109, 156 | `MustApply` (uses `logic.` alias) |
| `interp/solo7_interp_test.go` | 34, 140 | `MustApply` |

## Verification

```bash
cd ~/ivy/goivy && go build ./...
go test ./...
make golden
```

Any panics from `MustApply` reveal real sort/arity bugs that were previously hidden. Fix those callers, don't revert to raw structs.
