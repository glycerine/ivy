# Plan: Audit Items 10-15 Batch Fix

## Context
Items 10-15 from AUDIT18MARCH.md address missing/divergent behavior in the goivy port vs Python Ivy. Item 11 (TypeCheckContext) is already correctly ported and requires no work. The remaining 5 items involve: Schema instantiation (10), selective assertion checking (12-13), and context manager patterns (14-15).

## Item 11: TypeCheckContext — SKIP (already correct)

---

## Item 13: LabeledFormula/Unprovable in AssumeAction/AssertAction

**Problem:** `ast.LabeledFormula` exists with `Unprovable` field, but the compiler at `compiler/action.go:136-153` strips it when building assert/assume actions, discarding the flag.

**Python behavior:** `AssumeAction.action_update` and `AssertAction.action_update` check `isinstance(fmla, LabeledFormula)`, extract `fmla.unprovable`, and use it for filtering.

### Changes

**File: `actions/action.go`**
- Add `Unprovable bool` field to `AssumeAction` (line 213)
- Add `Unprovable bool` field to `AssertAction` (line 238)
- Propagate in `ActionClone` for both types

**File: `compiler/action.go`**
- At line 136-145 (assert case): after `lf, ok := inner.(*ast.LabeledFormula)`, set `act.Unprovable = lf.Unprovable`
- At line 152-161 (assume case): same pattern
- At line 120-129 (ensure case): same pattern
- At lines 577-584 (`CompileAssertFormula`): check for LabeledFormula, propagate Unprovable
- At lines 588-596 (`CompileAssumeFormula`): same

---

## Item 12: checked_assert / check_unprovable in ActionUpdate

**Problem:** Parameters `CheckUnprovable` (check/check.go) and `CheckedAssert` (mc/phase7.go) exist but `AssertAction.ActionUpdate` ignores them.

**Python behavior** (ivy_actions.py:343-362):
1. If `check_unprovable.get() != unprovable` → skip (return true_clauses, false_clauses)
2. If `checked_assert.get()` set and doesn't match lineno → return formula as-is (not dual) for provable, or skip for unprovable
3. Only matching assertions get `dual_formula`

### Changes

**File: `actions/update.go`**
- Import the check package (or pass params through UpdateContext)
- Rewrite `AssumeAction.ActionUpdate` (line 449-454):
  ```go
  func (a *AssumeAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
      if a.Unprovable {
          return makeUpdate([]*lg.Symbol{}, lg.True, lg.False, EmptyAnnotation{})
      }
      fmla := a.Formula
      fmla = skolemizeFormula(fmla)
      return makeUpdate([]*lg.Symbol{}, fmla, lg.False, EmptyAnnotation{})
  }
  ```
- Rewrite `AssertAction.ActionUpdate` (line 461-467):
  ```go
  func (a *AssertAction) ActionUpdate(ctx *UpdateContext) *transrel.Update {
      fmla := a.Formula
      unprovable := a.Unprovable

      if ctx.CheckUnprovable != unprovable {
          return makeUpdate([]*lg.Symbol{}, lg.True, lg.False, EmptyAnnotation{})
      }
      if ctx.CheckedAssert != "" {
          if ctx.CheckedAssert != a.GetLineno().String() {
              if unprovable {
                  return makeUpdate([]*lg.Symbol{}, lg.True, lg.False, EmptyAnnotation{})
              }
              return makeUpdate([]*lg.Symbol{}, fmla, lg.False, EmptyAnnotation{})
          }
      }
      dual := dualFormula(fmla)
      return makeUpdate([]*lg.Symbol{}, lg.True, dual, EmptyAnnotation{})
  }
  ```
- Add `CheckUnprovable bool` and `CheckedAssert string` fields to `UpdateContext` struct (line 32-45)
- Update callers that create `UpdateContext` to populate these from the global parameters

---

## Item 10: Schema.get_instance() — Reconcile two Schema types

**Problem:** Two incompatible `Schema` types exist:
- `ast/decl.go:464` — AST Schema with `GetInstance()` (full substitution+compilation)
- `actions/action.go:140` — actions Schema with `Instantiate()` that only appends

**Python:** Single `Schema` class in `ivy_actions.py` where `instantiate()` calls `get_instance()`.

### Changes

**File: `actions/action.go`**
- Remove the duplicate `Schema` struct (lines 137-178). The canonical Schema is `ast.Schema`.
- Or, make `actions.Schema` embed/delegate to `ast.Schema` and have `Instantiate` call `GetInstance`.

**Decision needed:** The `actions.Schema` stores `lg.Expr` types while `ast.Schema` stores `ast.Node`. These are different type hierarchies. The simplest fix: make `actions.Schema.Instantiate` perform substitution+compilation matching Python, rather than removing the type entirely. This avoids a large type-system refactor.

**File: `actions/action.go`**
- Add `GetInstance(params []lg.Expr, toClauses bool) (lg.Expr, error)` method to actions Schema
- Port the substitution logic: build subst map, rewrite AST, compile
- Make `Instantiate` call `GetInstance` instead of raw append

---

## Item 14: ActionContext context manager with global context

**Problem:** Go's `ActionContext` is a bare struct. Python uses `__enter__`/`__exit__` to push/pop a global `context` variable. Code checks `isinstance(context, UnrollContext)` at runtime.

### Changes

**File: `actions/action.go`**
- Add `OldContext` field to `ActionContext` (for save/restore)
- Add global: `var GlobalContext interface{} = &ActionContext{}`
- Add methods:
  ```go
  func (ac *ActionContext) Enter() {
      ac.OldContext = GlobalContext
      GlobalContext = ac
  }
  func (ac *ActionContext) Exit() {
      GlobalContext = ac.OldContext
  }
  ```
- Add `Get(symbol string) Action` method (delegates to `ivy_module.FindAction`)
- Add `NewState` method matching Python
- Add `RunWithActionContext` helper for scoped execution

**File: `actions/phase3.go`**
- Add `OldContext` field + `Enter`/`Exit` methods to `UnrollContext`
- These save/restore `GlobalContext` just like the base class

**File: `actions/update.go`**
- In `WhileAction` int_update equivalent: check `if uc, ok := GlobalContext.(*UnrollContext); ok` for loop unrolling dispatch

---

## Item 15: SymExContext — Rewrite to match Python semantics

**Problem:** Go's `SymExContext` tracks `Updated`/`PathCondition`/`FreshCounter`. Python's just stores `params` and manages global `symex_params`.

### Changes

**File: `actions/extra_actions.go`**
- Add global: `var SymexParams []lg.Expr`
- Replace `SymExContext` struct:
  ```go
  type SymExContext struct {
      Params    []lg.Expr
      OldParams []lg.Expr
  }
  func NewSymExContext(params []lg.Expr) *SymExContext {
      return &SymExContext{Params: params}
  }
  func (ctx *SymExContext) Enter() {
      ctx.OldParams = SymexParams
      SymexParams = ctx.Params
  }
  func (ctx *SymExContext) Exit() {
      SymexParams = ctx.OldParams
  }
  ```
- Remove `Fresh()`, `AddPathCondition()` methods
- Add `RunWithSymExContext` helper
- Update any callers of old `SymExContext` API (search for `NewSymExContext`, `Fresh`, `AddPathCondition`)

---

## Implementation Order

1. **Item 13** first (LabeledFormula propagation) — prerequisite for Item 12
2. **Item 12** next (checked_assert/check_unprovable in ActionUpdate) — depends on Unprovable field
3. **Items 10, 14, 15** in parallel — independent of each other

## Critical Files

| File | Items |
|------|-------|
| `actions/action.go` | 10, 13, 14 |
| `actions/update.go` | 12, 14 |
| `compiler/action.go` | 13 |
| `actions/extra_actions.go` | 15 |
| `actions/phase3.go` | 14 |

## Verification

1. `cd /Users/jaten/go/src/github.com/glycerine/goivy && make test` (uses DYLD_LIBRARY_PATH)
2. Check compilation: `go build ./...`
3. Grep for any remaining references to old `SymExContext` API (Fresh, AddPathCondition, Updated)
4. Grep for callers of `actions.Schema.Instantiate` to ensure they still work after changes
