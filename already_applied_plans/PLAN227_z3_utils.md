# Plan: Faithful Port of Python z3_utils.py to Go

**Created:** 2026-04-07 (afternoon)

## Context

Python Ivy's `z3_utils.py` provides a **simple, bare recursive translator** (`to_z3`) from `logic.py` types to Z3 objects, plus `z3_implies` and `z3_implies_batch` functions. The current Go `ImpliesBatch` at `solver/solver.go:648` incorrectly uses `s.tr.Translate()` — the ivy_solver-style translator which adds LookupNative, SolverName, QuantConstraints, EqFunc, HASH tracing, and other features that `to_z3()` does NOT have. This plan ports `z3_utils.py` faithfully.

### Why z3bridge/ not solver/

`z3_utils.go` lives in `z3bridge/` because:
- z3bridge already imports `logic` (the types to_z3 operates on) — no new dependencies
- z3bridge does NOT import solver — no circular dependency risk
- All 6 packages that import z3bridge (solver, art, interp, alpha, vmt, updr) gain access
- Python's z3_utils.py is a standalone utility, not part of ivy_solver — z3bridge is the natural Go equivalent

### Critical semantic differences between `to_z3()` and `Translator.Translate()`

| Feature | Python `to_z3()` | Go `Translator.Translate()` |
|---------|-------------------|----------------------------|
| LookupNative | NO | YES |
| SolverName | NO (plain name) | YES (may rename) |
| QuantConstraints | NO (bare ForAll/Exists) | YES (nat>=0, range bounds) |
| EqFunc/MyEq | NO (plain z3 ==) | YES |
| EnumEqFunc | NO | YES |
| NumeralFunc | NO | YES |
| HASH tracing | NO | YES |
| Const naming | `name:str(sort)` for ALL Var/Const | Only Variable uses `name:sort` |
| Boolean Apply routing | Same as non-Boolean | Routed to atomToZ3() |
| Iff operator | z3 equality (==) | EqFunc or Ctx.Iff |

## Files to Modify

| File | Action |
|------|--------|
| `z3bridge/z3_utils.go` | **NEW** — Z3Utils struct, ToZ3, Z3Implies, Z3ImpliesBatch |
| `solver/solver.go` | **MODIFY** — Add z3u field, update NewSolver/Close/Clear, rewrite ImpliesBatch |
| `solver/solver_test.go` | **MODIFY** — Update cache access in TestImpliesBatchCache |

## Step 1: Create `z3bridge/z3_utils.go`

### 1A: Z3Utils struct

All Python module-level mutable globals go on this struct (CLAUDE.md rule C).
Since this lives in z3bridge, it uses z3bridge types directly (Sort, Expr, FuncDecl, Z3Context) — no cross-package import needed.

```go
// Z3Utils provides a simple, bare recursive translator from logic types
// to Z3 objects, plus implication checking. Faithfully ports Python
// z3_utils.py. This is independent from Translator (which ports ivy_solver.py).
type Z3Utils struct {
    Ctx              *Z3Context                        // own context, independent from Translator
    toZ3Cache        map[logic.NodeKey]any             // Python _to_z3_cache (stores Sort, Expr, or FuncDecl)
    uninterpSorts    map[logic.NodeKey]Sort             // Python _z3_uninterpreted_sorts
    ImpliesCache     map[[2]logic.NodeKey]bool          // Python _implies_cache
}
```

Z3Utils gets its **own** `Z3Context`, independent from the Translator's context. Python's `z3_utils.py` operates independently of `ivy_solver.py` — same principle.

Read-only state (`_z3_interpreted`, `_z3_operators`, `_z3_quantifiers`) is handled inline via Go type switches — no need for map lookups.

`ImpliesCache` is exported so `solver_test.go` (in package solver) can inspect it in TestImpliesBatchCache.

### 1B: Constructor and lifecycle

- `NewZ3Utils() *Z3Utils` — creates Z3Context, initializes maps
- `(u *Z3Utils) Close() error` — closes Z3Context
- `(u *Z3Utils) Clear()` — resets all caches

### 1C: Port `to_z3(x)` and `_to_z3(x)`

**Public cached wrapper** — `(u *Z3Utils) ToZ3(x logic.Expr) (any, error)`:
- Check `toZ3Cache[x.Sexp()]`; on hit return cached value
- Otherwise call `toZ3Internal(x)`, cache result, return it
- Return type is `any` because Python's cache stores Sort, Expr, and FuncDecl

**Typed convenience wrappers**:
- `(u *Z3Utils) ToZ3Expr(x logic.Expr) (Expr, error)` — calls ToZ3, type-asserts to Expr
- `(u *Z3Utils) toZ3Sort(s logic.Sort) (Sort, error)` — calls ToZ3 (sorts implement Expr), type-asserts to Sort

**Private recursive implementation** — `(u *Z3Utils) toZ3Internal(x logic.Expr) (any, error)`:

Type switch dispatch (faithful to Python `_to_z3` lines 53-103):

1. **`*logic.BooleanSort`** → `u.Ctx.BoolSort()` (returns Sort)
   - Python: `_z3_interpreted[Boolean]` → `z3.BoolSort()`

2. **`*logic.And` with len==0** (logic.True) → `u.Ctx.BoolVal(true)` (returns Expr)
   - Python: `_z3_interpreted[true]` → `z3.BoolVal(True)`

3. **`*logic.Or` with len==0** (logic.False) → `u.Ctx.BoolVal(false)` (returns Expr)
   - Python: `_z3_interpreted[false]` → `z3.BoolVal(False)`

4. **`*logic.UninterpretedSort`** → check `uninterpSorts` cache; if miss, `u.Ctx.UninterpretedSort(name)`, cache it (returns Sort)

5. **`*logic.FunctionSort`** → error (Python: `assert False`)

6. **`*logic.Variable` with `FirstOrderSort(v.VSort)`** →
   `u.Ctx.Const(v.Name + ":" + v.VSort.String(), toZ3Sort(v.VSort))` (returns Expr)
   - Key detail: uses `sort.String()` not `sortDisplayName()` — gives `"S"` for UninterpretedSort, `"Boolean"` for BooleanSort

7. **`*logic.Const` with `FirstOrderSort(c.CSort)`** →
   `u.Ctx.Const(c.Name + ":" + c.CSort.String(), toZ3Sort(c.CSort))` (returns Expr)
   - DIFFERS from Translator which uses just `name` (no `:sort` suffix for Const)

8. **`*logic.Variable` with FunctionSort** → error (Python: `assert type(x) is Const`)

9. **`*logic.Const` with FunctionSort, Arity()==0** (nullary) →
   `u.Ctx.Const(c.Name + ":" + fs.Range().String(), toZ3Sort(fs.Range()))` (returns Expr)
   - Converts 0-ary function to first-order constant

10. **`*logic.Const` with FunctionSort, Arity()>=1** (higher-order) →
    Translate all `fs.Sorts` via `toZ3Sort`, create `u.Ctx.Function(c.Name, domain, range)` (returns FuncDecl)

11. **`*logic.Apply` with len(Terms)==0** → `u.ToZ3(app.Func)` (delegate to func)

12. **`*logic.Apply` with len(Terms)>0** →
    - `funcResult := u.ToZ3(app.Func)` — yields FuncDecl for higher-order, or Expr for first-order
    - Type-assert to FuncDecl: `fd.Apply(translatedArgs...)`
    - Args translated via `ToZ3Expr(t)` for each term

13. **`*logic.Eq`** → `u.Ctx.Eq(ToZ3Expr(T1), ToZ3Expr(T2))` — plain equality, NO EqFunc

14. **`*logic.Not`** → `u.Ctx.Not(ToZ3Expr(Body))`

15. **`*logic.And`** (non-empty) → `u.Ctx.And(translatedTerms...)`

16. **`*logic.Or`** (non-empty) → `u.Ctx.Or(translatedTerms...)`

17. **`*logic.Implies`** → `u.Ctx.Implies(ToZ3Expr(T1), ToZ3Expr(T2))`

18. **`*logic.Iff`** → `u.Ctx.Eq(ToZ3Expr(T1), ToZ3Expr(T2))` — NOTE: Eq not Iff! Python maps Iff to `==`

19. **`*logic.Ite`** → `u.Ctx.Ite(ToZ3Expr(Cond), ToZ3Expr(Then), ToZ3Expr(Else))`

20. **`*logic.ForAll`** → if len(Variables)==0: `ToZ3Expr(Body)`, else: `u.Ctx.ForAll(boundVars, body)` — NO QuantConstraints

21. **`*logic.Exists`** → same pattern as ForAll but `u.Ctx.Exists`

22. **default** → error (Python: `assert False, type(x)`)

### 1D: Port `z3_implies` (lines 109-133)

`(u *Z3Utils) Z3Implies(f1, f2 logic.Expr, timeout bool) (bool, error)`:
- Check ImpliesCache; on hit return cached
- Create `u.Ctx.NewSolver()`
- If timeout: `s.SetParam("timeout", "2000")`
- Assert `ToZ3Expr(f1)`
- Assert `ToZ3Expr(Not{Body: f2})`
- Check: Sat→false, Unsat→true, else→error
- Cache and return

### 1E: Port `z3_implies_batch` (lines 136-171)

`(u *Z3Utils) Z3ImpliesBatch(premise logic.Expr, formulas []logic.Expr, timeout bool) ([]bool, error)`:
- Create `u.Ctx.NewSolver()`
- If timeout: `s.SetParam("timeout", "2000")`
- Assert `ToZ3Expr(premise)`
- For each formula f:
  - Check ImpliesCache; on hit, use cached value
  - Else: Push, assert `ToZ3Expr(Not{Body: f})`, Check, Pop
  - Sat→false, Unsat→true, else→error
  - Cache result
- Return results slice

## Step 2: Modify `solver/solver.go`

### 2A: Add `z3u` field to Solver

```go
type Solver struct {
    mu               sync.Mutex
    tr               *z3bridge.Translator
    z3u              *z3bridge.Z3Utils     // NEW — for z3_utils.py operations
    opts             *module.SolverOptions
    sig              *il.Sig
    HandleRangeSorts bool
    // impliesCache REMOVED — now on Z3Utils.ImpliesCache
}
```

### 2B: Update NewSolver

- Add `s.z3u = z3bridge.NewZ3Utils()`
- Remove `impliesCache: make(...)` initialization

### 2C: Update Close

- Add `s.z3u.Close()` alongside `s.tr.Close()`

### 2D: Update Clear

- Add `s.z3u.Clear()` alongside `s.tr.Clear()`
- Remove `s.impliesCache = make(...)` line

### 2E: Rewrite ImpliesBatch as thin wrapper

```go
func (s *Solver) ImpliesBatch(premise lg.Expr, fmlas []lg.Expr, timeout bool) ([]bool, error) {
    return s.z3u.Z3ImpliesBatch(premise, fmlas, timeout)
}
```

Remove old implementation (lines 648-696) entirely.

### 2F: Add Z3Implies wrapper

```go
func (s *Solver) Z3Implies(f1, f2 lg.Expr, timeout bool) (bool, error) {
    return s.z3u.Z3Implies(f1, f2, timeout)
}
```

## Step 3: Update `solver/solver_test.go`

### 3A: Fix TestImpliesBatchCache

Line 2367-2369 accesses `s.impliesCache` directly. Change to `s.z3u.ImpliesCache`:
```go
key := [2]lg.NodeKey{pAndQ.Sexp(), p.Sexp()}
if _, ok := s.z3u.ImpliesCache[key]; !ok {
    t.Error("cache should contain the entry")
}
```

## Verification

1. **Run existing tests**: `go test ./solver/ -run TestImpliesBatch -v` — all 7 ImpliesBatch tests must pass
2. **Run full solver tests**: `go test ./solver/ -v` — no regressions
3. **Run z3bridge tests**: `go test ./z3bridge/ -v` — no regressions
4. **Run full project tests**: `go test ./...` — no regressions
5. **Manual check**: Verify z3_utils.go has 1:1 function correspondence with z3_utils.py:
   - `to_z3` → `ToZ3` / `toZ3Internal`
   - `_to_z3` → `toZ3Internal`
   - `z3_implies` → `Z3Implies`
   - `z3_implies_batch` → `Z3ImpliesBatch`
   - `_to_z3_cache` → `Z3Utils.toZ3Cache`
   - `_implies_cache` → `Z3Utils.ImpliesCache`
   - `_z3_uninterpreted_sorts` → `Z3Utils.uninterpSorts`
