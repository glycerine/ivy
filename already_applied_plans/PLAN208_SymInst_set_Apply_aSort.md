# PLAN205c: Fix SymInst Apply sort + elimDefinitions iteration order

**Created:** 2026-04-05 ~00:15 UTC

## Context

PLAN205/205b fixed compose_updates non-determinism and elimDefinitions iteration order. The golden test now shows a NEW diff pattern at line 166784: Go produces `Eq` where Python produces `Iff` for frame definitions of boolean-valued functions. The formula names now MATCH (progress from 205b), but the connective type differs.

## Root Cause: `SymInst` doesn't set Apply.aSort

**Go** (`module/clauses.go:444`):
```go
app := &lg.Apply{Func: sym, Terms: args}  // aSort = nil (zero value)
```

**Python** (`ivy_logic_utils.py:827-830`):
```python
def sym_inst(sym):
    if is_relational(sym):
        return Atom(sym, rel_placeholders(sym))   # Atom sets sort=Boolean
    return sym(*fun_placeholders(sym))              # Apply computes sort from FunctionSort
```

When `defToConstraint` calls `IsIndividual(d.Lhs)`, it checks `d.Lhs.NodeSort() != Boolean`:
- Go: `NodeSort()` returns nil (not set) → nil != Boolean → `IsIndividual=true` → **Eq**
- Python: `NodeSort()` returns BooleanSort (from Apply constructor) → `IsIndividual=false` → **Iff**

## Fix

### Step 1: Change `SymInst` to use `NewApply`

In `module/clauses.go`, replace the raw struct construction with the proper constructor:

```go
func SymInst(sym *lg.Const) lg.Expr {
    phs := SymPlaceholders(sym)
    if len(phs) == 0 {
        return sym
    }
    args := make([]lg.Expr, len(phs))
    for i, v := range phs {
        args[i] = v
    }
    app, err := lg.NewApply(sym, args...)
    if err != nil {
        panic(fmt.Sprintf("SymInst: NewApply failed for %v: %v", sym, err))
    }
    return app
}
```

`NewApply` computes `aSort` from the function's sort:
- For `FunctionSort`: `aSort = fs.Range()` (Boolean for relations)
- For `TopSort` or nil: `aSort = TopS`

### Step 2: Verify + test

```bash
go build ./...
go test ./module/... -count=1
go test ./actions/... -count=1
make golden
```

## Files Modified

- `module/clauses.go` — `SymInst`: use `lg.NewApply` instead of raw `&lg.Apply{}`
