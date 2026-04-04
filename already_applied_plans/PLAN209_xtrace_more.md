# PLAN206: Comprehensive xtracer instrumentation for divergence debugging

**Created:** 2026-04-05 ~00:45 UTC

## Context

PLANs 203-205c made incremental fixes but the divergence at xtrace line 166784 persists. The diff has changed shape (now Eq vs Iff instead of completely different formulas), but we can't pinpoint the root cause without visibility into the intermediate computation steps. We need to instrument every function in the divergence path so the golden test can detect exactly WHERE Go and Python first produce different results.

## Tracing strategy

Add matching xtracer traces to both Go and Python for every function in the `ite_clauses`/`or_clauses`/`elimDeadDefinitions`/`frame_def`/`defToConstraint` call chain. Each trace must:
- Use the SAME format string in Go and Python so the golden test can compare them
- Include HASH canon (s-expression) of results where feasible
- Include key counts and names to detect ordering/content differences

### Go pattern: `if xtracer.Enabled { xtracer.Trace("...") }`
### Python pattern: `if __debug__: xtracer.trace("...")`

## Trace points to add

### 1. `frame_def` / `FrameDefConst` — trace each frame definition

**Go** `actions/transrel.go` `FrameDefConst` (~line 1741):
```go
func FrameDefConst(sym *lg.Const, op func(*lg.Const) *lg.Const) *il.Definition {
    opSym := op(sym)
    lhs := mod.SymInst(opSym)
    rhs := mod.SymInst(sym)
    def := il.NewDefinition(lhs, rhs)
    if xtracer.Enabled {
        xtracer.Trace("transrel.FrameDefConst sym=%v sort=%v def=%v", sym.Name, sym.CSort, def.Canon())
    }
    return def
}
```

**Python** `ivy_transrel.py` `frame_def` (~line 149):
```python
def frame_def(sym,op):
    lhs = sym_inst(op(sym) if op is new else sym)
    rhs = sym_inst(sym if op is new else op(sym))
    dfn = Definition(lhs,rhs)
    if __debug__:
        xtracer.trace("transrel.FrameDefConst sym=%s sort=%s def=%s" % (sym.name, sym.sort, dfn.canon()))
    return dfn
```

### 2. `diff_frame` / `DiffFrameConst` — trace filtered symbols and result

**Go** `actions/transrel.go` `DiffFrameConst` (~line 1722):
Add after building defs list:
```go
if xtracer.Enabled {
    names := make([]string, len(defs))
    for i, d := range defs {
        names[i] = fmt.Sprintf("%v", d.Defines())
    }
    xtracer.Trace("transrel.DiffFrameConst nDefs=%d syms=%v", len(defs), names)
}
```

**Python** `ivy_transrel.py` `diff_frame` (~line 179):
Add before return:
```python
if __debug__:
    names = [str(d.defines()) for d in result.defs]  # result = frame(updated, op)
    xtracer.trace("transrel.DiffFrameConst nDefs=%d syms=%s" % (len(names), names))
```

### 3. `defToConstraint` / `to_constraint` — trace each conversion

**Go** `module/clauses.go` `defToConstraint` (~line 250):
```go
func defToConstraint(d *il.Definition) lg.Expr {
    result := il.DefinitionToConstraint(d)
    if xtracer.Enabled {
        xtracer.Trace("ops.defToConstraint lhs.sort=%v result_type=%T canon=%v",
            d.Lhs.NodeSort(), result, iu.CanonShort(result, 200))
    }
    return result
}
```

**Python** `ivy_logic_utils.py` — need to find `to_constraint` on the Definition class and add trace there. Alternatively, trace in `elim_definitions` when calling `to_constraint`.

### 4. `elimDefinitions` / `elim_definitions` — trace dead syms, input defs, output fmlas

**Go** `module/ops.go` `elimDefinitions` (~line 824):
Add at entry and exit:
```go
if xtracer.Enabled {
    deadNames := make([]string, len(dead))
    for i, k := range dead {
        deadNames[i] = string(k)
    }
    defNames := make([]string, len(clauses.Defs))
    for i, d := range clauses.Defs {
        defNames[i] = fmt.Sprintf("%v", d.Defines())
    }
    xtracer.Trace("ops.elimDefinitions ENTER nDead=%d dead=%v nDefs=%d defs=%v nFmlas=%d",
        len(dead), deadNames, len(clauses.Defs), defNames, len(clauses.Fmlas))
}
// ... existing code ...
if xtracer.Enabled {
    xtracer.Trace("ops.elimDefinitions EXIT nFmlas=%d nDefs=%d", len(fmlas), len(defs))
}
```

**Python** `ivy_logic_utils.py` `elim_definitions` (~line 1310):
```python
if __debug__:
    deadNames = [str(s) for s in dead]
    defNames = [str(d.defines()) for d in clauses.defs]
    xtracer.trace("ops.elimDefinitions ENTER nDead=%d dead=%s nDefs=%d defs=%s nFmlas=%d" %
        (len(dead), deadNames, len(clauses.defs), defNames, len(clauses.fmlas)))
# ... existing code ...
if __debug__:
    xtracer.trace("ops.elimDefinitions EXIT nFmlas=%d nDefs=%d" % (len(fmlas), len(defs)))
```

### 5. `elimDeadDefinitions` / `elim_dead_definitions` — trace captured, dead, toRename

**Go** `module/ops.go` `elimDeadDefinitions` (~line 752):
Add after computing captured, dead, toRename:
```go
if xtracer.Enabled {
    xtracer.Trace("ops.elimDeadDefinitions nArgs=%d nDefined=%d nCaptured=%d nDead=%d nToRename=%d",
        len(args), defined.Len(), len(captured), len(dead), len(toRename))
    if len(dead) > 0 {
        deadNames := make([]string, len(dead))
        for i, k := range dead { deadNames[i] = string(k) }
        xtracer.Trace("ops.elimDeadDefinitions dead=%v", deadNames)
    }
}
```

**Python** `ivy_logic_utils.py` `elim_dead_definitions` (~line 1326):
```python
if __debug__:
    xtracer.trace("ops.elimDeadDefinitions nArgs=%d nDefined=%d nCaptured=%d nDead=%d nToRename=%d" %
        (len(args), len(defd), len(captured), len(dead), len(to_rename)))
    if dead:
        xtracer.trace("ops.elimDeadDefinitions dead=%s" % [str(s) for s in dead])
```

### 6. `iteClausesInt` / `ite_clauses_int` — trace args, fmlas, defs

**Go** `module/ops.go` `iteClausesInt` (~line 361):
Add at entry (after elimDeadDefinitions) and exit:
```go
if xtracer.Enabled {
    xtracer.Trace("ops.iteClausesInt ENTER nArgs0Fmlas=%d nArgs0Defs=%d nArgs1Fmlas=%d nArgs1Defs=%d",
        len(args[0].Fmlas), len(args[0].Defs), len(args[1].Fmlas), len(args[1].Defs))
}
// ... existing code ...
// before return:
if xtracer.Enabled {
    result := NewClauses(fmlas, defs, annot)
    xtracer.Trace("ops.iteClausesInt EXIT HASH canon= %s", result.Canon())
}
```

**Python** `ivy_logic_utils.py` `ite_clauses_int` (~line 1368):
```python
if __debug__:
    xtracer.trace("ops.iteClausesInt ENTER nArgs0Fmlas=%d nArgs0Defs=%d nArgs1Fmlas=%d nArgs1Defs=%d" %
        (len(args[0].fmlas), len(args[0].defs), len(args[1].fmlas), len(args[1].defs)))
# ... existing code ...
# before return:
if __debug__:
    xtracer.trace("ops.iteClausesInt EXIT HASH canon= %s" % res.canon())
```

### 7. `orClausesIntWithVs` / `or_clauses_int` — trace args, result

**Go** `module/ops.go` `orClausesIntWithVs` (~line 283):
Add at entry and exit:
```go
if xtracer.Enabled {
    argInfo := make([]string, len(args))
    for i, a := range args {
        argInfo[i] = fmt.Sprintf("(%dF,%dD)", len(a.Fmlas), len(a.Defs))
    }
    xtracer.Trace("ops.orClausesInt ENTER nArgs=%d args=%v", len(args), argInfo)
}
// before return:
if xtracer.Enabled {
    xtracer.Trace("ops.orClausesInt EXIT HASH canon= %s", result.Canon())
}
```

**Python** `ivy_logic_utils.py` `or_clauses_int` (~line 1339):
```python
if __debug__:
    argInfo = ["(%dF,%dD)" % (len(a.fmlas), len(a.defs)) for a in args]
    xtracer.trace("ops.orClausesInt ENTER nArgs=%d args=%s" % (len(args), argInfo))
# before return:
if __debug__:
    xtracer.trace("ops.orClausesInt EXIT HASH canon= %s" % res.canon())
```

### 8. `iteUpdate` / `ite` — trace the ite action composition

**Go** `actions/transrel.go` `iteUpdate` (~line 838):
Add traces showing df12/df21 results and final clauses:
```go
if xtracer.Enabled {
    xtracer.Trace("transrel.iteUpdate ENTER u1.nMod=%d u2.nMod=%d", len(u1.Modified), len(u2.Modified))
    xtracer.Trace("transrel.iteUpdate df12 HASH canon= %s", df12.Canon())
    xtracer.Trace("transrel.iteUpdate df21 HASH canon= %s", df21.Canon())
    xtracer.Trace("transrel.iteUpdate c1 HASH canon= %s", c1.Canon())
    xtracer.Trace("transrel.iteUpdate c2 HASH canon= %s", c2.Canon())
}
```

**Python** `ivy_transrel.py` `ite` (~line 204):
```python
if __debug__:
    xtracer.trace("transrel.iteUpdate ENTER u1.nMod=%d u2.nMod=%d" % (len(u1), len(u2)))
    xtracer.trace("transrel.iteUpdate df12 HASH canon= %s" % df12.canon())
    xtracer.trace("transrel.iteUpdate df21 HASH canon= %s" % df21.canon())
    xtracer.trace("transrel.iteUpdate c1 HASH canon= %s" % c1.canon())
    xtracer.trace("transrel.iteUpdate c2 HASH canon= %s" % c2.canon())
```

## Files to modify

- `module/ops.go` — traces in elimDeadDefinitions, elimDefinitions, orClausesIntWithVs, iteClausesInt
- `module/clauses.go` — trace in defToConstraint
- `actions/transrel.go` — traces in DiffFrameConst, FrameDefConst, iteUpdate
- `~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py` — traces in elim_dead_definitions, elim_definitions, or_clauses_int, ite_clauses_int
- `~/ivy/pyivy/ivy/ivy/ivy_transrel.py` — traces in diff_frame, frame_def, ite

## Verification

```bash
go build ./...
go test ./module/... -count=1
go test ./actions/... -count=1
make golden
```

The golden test will now show the FIRST trace line where Go and Python diverge, pinpointing the exact function and data that differs.
