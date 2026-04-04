# PLAN201: Fix GetModCone natives section + delete dead code

**Created:** 2026-04-05 ~00:15 UTC

## Context

Two issues found during duplication audit D4:

1. **phase7.go `GetModCone`** is dead code — never called from anywhere. Should be deleted.
2. **helpers.go `GetModConeFull`** (the active version, called from `isolate.go:930`) has
   a bug in its natives section: it casts each native to `*ast.LabeledFormula`, but natives
   are `*ast.NativeDef` after compilation, so the cast always fails and no native references
   are followed. Python iterates `n.args[2:]` to find action references.

### Python source of truth (`ivy_isolate.py:1635-1647`)

```python
for n in mod.natives:
    for a in n.args[2:]:
        if isinstance(a, ivy_ast.Atom) and a.rep in mod.actions:
            get_cone(actions, a.rep, cone)
```

Native `args` layout: `[name, code_template, ref1, ref2, ...]`
- `args[0]` = name (compiled via `compile_native_name` → still an Atom)
- `args[1]` = code template (NativeCode)
- `args[2:]` = references (compiled via `compile_native_arg` or `compile_native_symbol`)

Python's `compile_native_arg` returns an `ivy_ast.Atom` (clone) for names not in
`sig.symbols` — these are action references. `isinstance(a, ivy_ast.Atom)` catches these.

### Go compilation of natives (`compiler/phase6.go:1043-1088`)

Go's `CompileNativeDef` wraps all compiled args in `*ast.CompiledNode{Node: lg.Expr}`:
- `compile_native_arg` path → `CompiledNode{Node: *lg.Const}` or `CompiledNode{Node: *lg.Apply}`
- `compile_native_symbol` path → `CompiledNode{Node: *lg.Const}` (symbol from sig)

So Go equivalent of Python's `isinstance(a, ivy_ast.Atom)` is checking for `*ast.CompiledNode`
containing `*lg.Const` (simple name) or `*lg.Apply` (name with args, extract via `.Func`).
Also check `*ast.Atom` directly for the pre-compilation case.

## Fix

### Step 1: Delete dead `GetModCone` from phase7.go

Delete `phase7.go:222-261` (`GetModCone` function). It's never called and duplicates
`GetModConeFull`.

### Step 2: Fix natives section in `GetModConeFull` (helpers.go:753-769)

Replace the current `*ast.LabeledFormula` cast (which always fails for `*ast.NativeDef`)
with proper iteration of `n.Args()[2:]`, matching Python's logic:

```go
// Add actions referenced by natives
// Python: for n in mod.natives: for a in n.args[2:]: if isinstance(a,ivy_ast.Atom) and a.rep in mod.actions
for _, nat := range mod.Natives {
    args := nat.Args()
    for i := 2; i < len(args); i++ {
        name := ""
        if atom, ok := args[i].(*ast.Atom); ok {
            // Pre-compilation: direct Atom (matches Python isinstance check)
            name = atom.Rep
        } else if cn, ok := args[i].(*ast.CompiledNode); ok {
            // Post-compilation: CompiledNode wrapping lg.Expr
            switch e := cn.Node.(type) {
            case *lg.Const:
                name = e.Name
            case *lg.Apply:
                if c, ok := e.Func.(*lg.Const); ok {
                    name = c.Name
                }
            }
        }
        if name != "" {
            if _, exists := actionsMap.Get2(name); exists {
                GetCone(actionsMap, name, cone)
            }
        }
    }
}
```

### Step 3: Verify

```
go build ./...
go test ./isolate/...
go test ./...
```

## Files Modified

- `isolate/phase7.go` — delete `GetModCone` (lines 222-261)
- `isolate/helpers.go` — rewrite natives section of `GetModConeFull` (lines 753-769)
