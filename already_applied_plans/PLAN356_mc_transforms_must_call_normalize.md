# Fix missing normalize() in InstantiateAxioms (xtrace 227836)

**Created:** 2026-04-30 ~07:30 UTC

## Context

The rfn golden test diverges at xtrace 227836 (`mc.ToAiger postAxiomInst`):
- **Go:** `nAxioms=810`
- **Python:** `nAxioms=1613`
- **S-expression diff:** Go emits `(Eq t1:(Apply ...))` where Python emits `(And terms:[])` (True) — Python normalizes tautological equalities, Go does not.

Root cause: Go's `InstantiateAxioms` (`mc/transforms.go:646`) does not call `normalize()` after substituting variables into axiom formulas. Python does (`ivy_mc.py:751`).

## Fix (2 files, ~4 changed lines)

All required helper functions already exist in `mc/phase7.go:63-125` (`normalize`, `cloneNormal`, `termOrd`). No new functions needed.

### 1. `mc/transforms.go` — Add `iuCfg` parameter and `normalize()` call

**Line 607** — add `iuCfg` parameter to `InstantiateAxioms`:
```go
// before:
func InstantiateAxioms(mod *module.Module, stVars []string, trans *module.Clauses, invariant lg.Expr, sortConstants map[string][]*lg.Const, funs *iu.InsMap[string, *lg.Const]) []lg.Expr {

// after:
func InstantiateAxioms(mod *module.Module, stVars []string, trans *module.Clauses, invariant lg.Expr, sortConstants map[string][]*lg.Const, funs *iu.InsMap[string, *lg.Const], iuCfg *iu.IvyUtilsConfig) []lg.Expr {
```

**Line 646** — call `normalize()` after substitution (mirrors Python `ivy_mc.py:751`):
```go
// before:
inst := lu.SubstituteByName(te.axiom.Formula.(lg.Expr), mp)

// after:
inst := normalize(lu.SubstituteByName(te.axiom.Formula.(lg.Expr), mp), iuCfg)
```

### 2. `mc/toaiger.go` — Pass `iuCfg` at call site

**Line 303** — thread `mod.Cfg.IuCfg` (already available; used at line 295):
```go
// before:
axs := InstantiateAxioms(mod, stVarNameList, trans, invariant, sortConstants, funs)

// after:
axs := InstantiateAxioms(mod, stVarNameList, trans, invariant, sortConstants, funs, mod.Cfg.IuCfg)
```

### 3. Diagnostic xtraces for axiom count investigation

The normalize fix resolves the confirmed structural divergence. The count difference (810 vs 1613) may have a separate cause. Add diagnostic xtraces inside `InstantiateAxioms` plus matching ones in Python `instantiate_axioms` to narrow it down:

**Go `mc/transforms.go`** — after trigger collection (after line 632):
```go
xtracer.Trace("mc.InstantiateAxioms nAxioms=%d nTriggers=%d", len(axioms), len(triggers))
```

**Go `mc/transforms.go`** — before return (line 667):
```go
xtracer.Trace("mc.InstantiateAxioms result nUnique=%d", len(instList))
```

**Python `ivy_mc.py`** — matching xtraces at corresponding locations:
```python
# after line 712:
if __debug__: xtracer.trace("mc.InstantiateAxioms nAxioms=%d nTriggers=%d" % (len(axioms), len(triggers)))

# after line 758:
if __debug__: xtracer.trace("mc.InstantiateAxioms result nUnique=%d" % len(inst_list))
```

## Existing infrastructure (read-only, no changes needed)

| File | Functions | Purpose |
|---|---|---|
| `mc/phase7.go:63` | `normalize(expr, iuCfg)` | Recursive normalization with macro expansion |
| `mc/phase7.go:78` | `cloneNormal(expr, args)` | Eq(X,X)→And(), operand ordering |
| `mc/phase7.go:94` | `termOrd(x, y)` | Total ordering on terms |
| `ivylogic/classify_ext.go:345` | `IsMacro`, `ExpandMacro` | Macro detection/expansion |
| `ivylogic/util.go:13,193` | `CloneNode`, `NodeArgs` | AST cloning helpers |

## Verification

1. `cd ~/ivy/goivy && make test` — compilation + unit tests
2. Run the rfn golden test to check if xtrace 227836 converges
3. If count still differs, compare the new diagnostic xtraces (nAxioms, nTriggers) between Go and Python to identify the remaining divergence source
