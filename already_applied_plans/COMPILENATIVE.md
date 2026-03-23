# Batch G: Native/compile_native + ARG pass 3

## Context

Two remaining items from SECTION6_PLAN.md Batch G:
- **§6.3 #27**: `DomainSetup.Native` stores raw node without `compile_native_def`. In Python, `native` is only handled in `IvyARGSetup` (pass 3), not `IvyDomainSetup` (pass 1). The Go code handles it in BOTH passes — DomainSetup appends raw, then ARGSetup appends compiled — causing double-append.
- **§6.3 #30**: `ARGSetup` pass 3 has behavioral differences in exports, imports, and progress:
  - **export**: Logs warning instead of raising error via `CheckIsAction`
  - **import_**: No `CheckIsAction` validation at all
  - **progress**: Missing `add_symbol` call to register the progress relation in the signature (Python does `add_symbol(rel.relname, get_relation_sort(...))` before sortifying)
  - **delegate**: Already matches Python (simple append) — no fix needed

## Bugs to Fix

### Bug 1: DomainSetup.Native double-appends raw nodes (§6.3 #27)
- **Python**: Only `IvyARGSetup.native` (line 1445) handles native — calls `compile_native_def()` and appends
- **Go**: `DomainSetup.Native` (decl.go:1013) appends raw node, then `ARGSetup` (ivy_compile.go:524) also appends compiled node → double entries in `mod.Natives`
- **Fix**: Make `DomainSetup.Native` a no-op (return nil, don't append). ARGSetup already compiles correctly.

### Bug 2: CompileNativeDef arg/symbol decision logic differs (§6.3 #27)
- **Python** (line 789): Decides arg vs symbol based on `fields[i*2].endswith('"')` — uses the code template to determine which compilation path
- **Go** (phase6.go:930-939): Tries `CompileNativeArg` first, falls back to `CompileNativeSymbol` on error
- **Fix**: Port the Python field-based decision logic. Split code template by backticks, check if `fields[i*2]` ends with `"` to decide symbol vs arg.

### Bug 3: Export missing CheckIsAction error (§6.3 #30)
- **Python** (line 1436): `check_is_action(self.mod, exp, exp.exported())` — raises IvyError
- **Go** (ivy_compile.go:497-498): Only logs warning
- **Fix**: Return error from `CheckIsAction` instead of warning

### Bug 4: Import missing CheckIsAction validation (§6.3 #30)
- **Python** (line 1439): `check_is_action(self.mod, imp, imp.imported())`
- **Go** (ivy_compile.go:507-508): No validation at all
- **Fix**: Add `CheckIsAction` call for imports. Need to extract imported name from `ImportDef.Imported` field.

### Bug 5: Progress missing add_symbol (§6.3 #30)
- **Python** (line 1197-1202, in `IvyDomainSetup`): Before sortifying, calls `add_symbol(rel.relname, get_relation_sort(sig, rel.args, df.args[1]))` to register the progress relation symbol in the signature
- **Go** (decl.go:1102-1111): Only sortifies and appends — skips the critical `add_symbol` step
- **Fix**: In `DomainSetup.Progress`, extract `rel = df.args[0]`, compute relation sort via `c.GetRelationSort()`, then call `sig.AddSymbol()` before sortifying

## Files to Modify

1. **`~/goivy/compiler/decl.go`** — Make `DomainSetup.Native` a no-op; fix `DomainSetup.Progress` to add symbol
2. **`~/goivy/compiler/ivy_compile.go`** — Fix export/import handlers in ARGSetup.ProcessDecls
3. **`~/goivy/compiler/phase6.go`** — Fix `CompileNativeDef` field-based arg/symbol decision
4. **`~/goivy/compiler/batch_g_test.go`** — New test file

## Existing Utilities to Reuse

| Utility | Location | Purpose |
|---------|----------|---------|
| `CheckIsAction()` | `compiler/phase6.go:1652` | Validate name is an action |
| `c.GetRelationSort()` | `compiler/phase6.go:363` | Compute relation sort from args |
| `c.SortifyWithInference()` | `compiler/phase6.go:377` | Sort inference on AST |
| `sig.AddSymbol()` | `ivylogic/sig.go:94` | Register symbol in signature |
| `c.CompileNativeArg()` | `compiler/phase6.go:717` | Compile native arg |
| `c.CompileNativeSymbol()` | `compiler/phase6.go:774` | Compile native symbol |
| `c.CompileNativeName()` | `compiler/phase6.go:872` | Compile native name |
| `ast.NativeCode.Code` | `ast/decl.go:1220` | The code template string |
| `ExportDef.Exported()` | `ast/decl.go:1095` | Get exported action name |

## Plan

### Step 1: Fix DomainSetup.Native → no-op (`compiler/decl.go:1013`)

```go
func (d *DomainSetup) Native(node ast.Node) error {
	// Python only handles native in IvyARGSetup (pass 3), not IvyDomainSetup.
	// ARGSetup.ProcessDecls already compiles via CompileNativeDef and appends.
	return nil
}
```

### Step 2: Fix CompileNativeDef field-based logic (`compiler/phase6.go:930`)

Replace the try-arg-then-fallback-to-symbol logic with the Python field-based decision:

```go
// Extract code template and split by backticks
var fields []string
if len(args) > 1 {
    if nc, ok := args[1].(*ast.NativeCode); ok {
        fields = strings.Split(nc.Code, "`")
    }
}
for i := 2; i < len(args); i++ {
    fieldIdx := (i - 2) * 2
    useSymbol := fieldIdx < len(fields) && strings.HasSuffix(fields[fieldIdx], "\"")
    if useSymbol {
        compiled, err := c.CompileNativeSymbol(args[i])
        if err != nil {
            newArgs[i] = args[i]
            continue
        }
        newArgs[i] = &ast.CompiledNode{Node: compiled}
    } else {
        compiled, err := c.CompileNativeArg(args[i])
        if err != nil {
            newArgs[i] = args[i]
            continue
        }
        newArgs[i] = &ast.CompiledNode{Node: compiled}
    }
}
```

Need to add `"strings"` import to phase6.go if not present.

### Step 3: Fix export handler in ARGSetup (`compiler/ivy_compile.go:490`)

Change warning to error:
```go
case *ast.ExportDecl:
    for _, arg := range n.DeclArgs {
        if expDef, ok := arg.(*ast.ExportDef); ok {
            name := expDef.Exported()
            if err := CheckIsAction(mod, name); err != nil {
                return err
            }
            mod.Exports = append(mod.Exports, expDef)
        }
    }
```

### Step 4: Fix import handler in ARGSetup (`compiler/ivy_compile.go:503`)

Add CheckIsAction validation:
```go
case *ast.ImportDecl:
    for _, arg := range n.DeclArgs {
        if impDef, ok := arg.(*ast.ImportDef); ok {
            if atom, ok := impDef.Imported.(*ast.Atom); ok {
                name := atom.Relname()
                if err := CheckIsAction(mod, name); err != nil {
                    return err
                }
            }
        }
        mod.Imports = append(mod.Imports, arg)
    }
```

### Step 5: Fix DomainSetup.Progress to add symbol (`compiler/decl.go:1102`)

Python's `get_relation_sort(sig, rel.args, df.args[1])` passes the body as `term` parameter.
In Python's `get_arg_sorts`, when `term` is not None, it wraps `args + [term]` into an AST,
sortifies with inference, then takes `args[0:-1]` sorts. The body helps with sort inference
of untyped relation parameters.

Go's `GetArgSorts` doesn't support the `term` parameter. To match Python:
1. Concatenate `relArgs + [body]`
2. Call `GetArgSorts` on the combined list
3. Drop the last sort (the body's sort)
4. Build `RelationSort` from the remaining sorts

```go
func (d *DomainSetup) Progress(node ast.Node) error {
    args := node.Args()
    if len(args) >= 2 {
        rel := args[0]
        body := args[1]
        if atom, ok := rel.(*ast.Atom); ok {
            // Python: add_symbol(rel.relname, get_relation_sort(sig, rel.args, df.args[1]))
            // get_arg_sorts with term: sortify_with_inference(AST(*(args+[term]))).args[0:-1]
            relArgs := atom.Args()
            combined := append(append([]ast.Node{}, relArgs...), body)
            allSorts, err := d.Compiler.GetArgSorts(combined)
            if err == nil && len(allSorts) > 0 {
                paramSorts := allSorts[:len(allSorts)-1] // drop body sort
                relSort := il.RelationSort(paramSorts)
                d.Compiler.Module.Sig.AddSymbol(atom.Relname(), relSort)
            }
        }
    }
    compiled, err := d.Compiler.SortifyWithInference(node)
    if err != nil {
        return err
    }
    d.Compiler.Module.Progress = append(d.Compiler.Module.Progress, compiled)
    return nil
}
```

Note: `GetArgSorts` calls `CompileNode` (not `SortifyWithInference`), so it won't have
the full sort inference context. However, this matches the `term=None` path in Python
for when args already have explicit sorts. The `term` path is used for inferring sorts
from the body — if this is needed, we'd need to extend `GetArgSorts`. For now, this is
a good first approximation; if tests show sort inference failures on progress declarations,
we can add the term-based inference.

### Step 6: Write tests (`compiler/batch_g_test.go`)

1. **TestDomainSetupNative_NoOp** — DomainSetup.Native doesn't append to mod.Natives
2. **TestARGSetupNative_CompilesAndAppends** — ARGSetup compiles native defs
3. **TestCompileNativeDef_FieldBasedDecision** — Verify field-based arg vs symbol logic
4. **TestExport_CheckIsAction_Error** — Export of non-existent action returns error
5. **TestExport_CheckIsAction_Success** — Export of existing action succeeds
6. **TestImport_CheckIsAction_Error** — Import of non-existent action returns error
7. **TestImport_CheckIsAction_Success** — Import of existing action succeeds
8. **TestProgress_AddsSymbol** — Progress handler adds symbol to sig before sortifying
9. **TestProgress_SortifyAndAppend** — Progress result is sortified and appended

### Step 7: Verify

Run `make test` to ensure no regressions.

## Key Design Notes

- DomainSetup.Native becomes a no-op. This prevents raw nodes from being double-appended.
- The ARGSetup NativeDecl handler (ivy_compile.go:524) is already correct — it calls CompileNativeDef.
- `CheckIsAction` already returns `*lg.IvyError` — just need to propagate it instead of logging.
- Delegate handler is already correct (simple append matches Python).
- Python's `progress` is in `IvyDomainSetup` (pass 1), NOT `IvyARGSetup` (pass 3). The Go `DomainSetup.Progress` just needs the `AddSymbol` call added before the existing sortify-and-append.
- The `GetRelationSort` signature difference (Go doesn't take body arg) is acceptable — the body is only used for additional sort inference in complex cases; for standard progress declarations, the relation args suffice.
