# Plan: Fix `vSort:S` vs `vSort:mem_type` Divergence in Module Instantiation

**Created:** 2026-03-25T14:30

## Context

The golden test (`cd ~/goivy && make golden`) passes 26188 lines then diverges. Go produces `vSort:S` where Python produces `vSort:mem_type` for Variables inside module instantiation bodies.

**Golden test divergence** in `dramc.step_rd` action (from `instantiate dramc(M:mem_type) : dramc_mod(M)`):
```
-  vSort:S)         # Go
+  vSort:mem_type)   # Python
```

Some Variables in the same declaration are correctly rewritten (action name, top-level assignments) while others keep `vSort:S` (inside `someMin.fmla`). Both are Variable `"M"`.

---

## Root Cause: `DistinctVariableRenaming` only maps clashing variables

**Python** (`ivy_ast.py:1856-1858`):
```python
def distinct_variable_renaming(vars1, vars2):
    rn = UniqueRenamer('', (v.rep for v in vars2))
    return dict((v.rep, rename_variable(v, rn(v.rep))) for v in vars1)
```
Maps **ALL** vars1 variables. Non-clashing names get identity mappings but preserve the original Variable's sort.

**Go** (`ast/rewrite.go:921-940`):
```go
for _, v := range vars1 {
    if used[v.Rep] {   // ← BUG: only maps CLASHING variables
        ...
        result[v.Rep] = nv
    }
}
```

**Impact chain in `inst_mod`** (`lalr_full/inst_mod.go:248-254`):

For `instantiate dramc(M:mem_type) : dramc_mod(M)`:
- `dpref = Atom("dramc", [Variable("M", "mem_type")])` — pref has correct sort annotation
- `dvsubst = {"m": Variable("M", "S")}` — actual arg `M` parsed with default sort "S"
- `map1 = DistinctVariableRenaming(UsedVariablesAst(dpref), UsedVariablesAst(decl))`

**Python:** `map1["M"]` always exists (identity mapping, sort=mem_type from dpref) → `vvsubst["m"]` = Variable("M", **mem_type**) ✓

**Go:** `map1` is empty (no clash) → `buildVVSubst` falls back to `dvsubst["m"]` = Variable("M", **S**) ✗

When `SubstituteConstantsAst2` replaces `m` atoms, Go substitutes Variable("M", "S") instead of Variable("M", "mem_type").

**Why some Variables are correct:** The ActionDef name `(atom rep:"dramc.step_rd" terms:[(variable rep:"M" vSort:mem_type)])` comes from the **pref** atom directly (composed via `ComposeAtoms`), not from SubstituteConstantsAst2. Similarly, top-level `dramc.rd_fair` gets its Variable from a different code path. Only Variables created by substituting the module-body's `m` → `M` have the wrong sort.

---

## Implementation

### Step 1: Fix `ast/rewrite.go:DistinctVariableRenaming`

**File:** `ast/rewrite.go:921-940`

Change to map ALL vars1 variables (matching Python `UniqueRenamer` which always returns a mapping):

```go
func DistinctVariableRenaming(vars1, vars2 []*Variable) map[string]Node {
    used := make(map[string]bool)
    for _, v := range vars2 {
        used[v.Rep] = true
    }
    result := make(map[string]Node)
    for _, v := range vars1 {
        newName := v.Rep
        if used[newName] {
            for used[newName] {
                newName = newName + "'"
            }
        }
        used[newName] = true
        nv := &Variable{Rep: newName, VSort: v.VSort}
        nv.Cfg = v.Cfg
        result[v.Rep] = nv
    }
    return result
}
```

### Step 2: Fix `clauseops/batch16.go:DistinctVariableRenaming` (same bug)

**File:** `clauseops/batch16.go:148-173`

Apply the same fix: map ALL vars1 variables, not just clashing ones. This is the logic-layer version of the same function.

### Step 3: Clean up debug instrumentation

Remove temporary debug prints from prior session:
- `ast/decl.go` — LabeledFormula debug stderr prints / xtrace (if any remain)
- `lalr_full/inst_mod.go` — commented-out `vv()` calls
- `~/pyivy/ivy/ivy/ivy_ast.py` — temporary xtrace in `__init__`/`clone` (if any remain)
- `~/pyivy/ivy/ivy/ivy_parser.py` — temporary xtrace in inst_mod (if any remain)

### Step 4: Verify

```bash
cd ~/goivy && make golden     # should pass beyond line 26188
go test ./ast/                # no regressions
go test ./compiler/           # all previously-fixed tests still pass
go test ./clauseops/          # no regressions from Step 2
go test ./lalr_full/          # golden test passes further
```

---

## Files to Modify

| File | Change |
|------|--------|
| `ast/rewrite.go` | Fix `DistinctVariableRenaming` to map ALL vars1 (line 921-940) |
| `clauseops/batch16.go` | Fix `DistinctVariableRenaming` same pattern (line 148-173) |
| `ast/decl.go` | Remove debug prints (if any remain from prior session) |
| `lalr_full/inst_mod.go` | Remove debug prints (if any remain from prior session) |
| Python files | Remove temporary xtrace instrumentation |
