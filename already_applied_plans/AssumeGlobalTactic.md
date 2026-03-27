# Fix xtrace divergence at line 92279: AssumeTactic in CompileTactic

Created: 2026-03-27

## Context

The golden trace test (`make golden`) diverges at line 92279:
- **Go**: `compiler.Thing ENTER type=AssumeTactic` (starting to compile an AssumeTactic with full Thing/CompileNode/OtherThing traces)
- **Python**: `compiler.OtherThing return type=ComposeTactics sort_infer_root=False` (finishing the outer ComposeTactics)

## Root Cause

In Python, `TacticWithMatch.compile` is overridden at `ivy_compiler.py:1093`:
```python
ivy_ast.TacticWithMatch.compile = lambda self: compile_schema_instantiation(self, last_fmla)
```

`compile_schema_instantiation` simply does `return self` (line 1061 — the rest is dead code).

This override applies to ALL TacticWithMatch subclasses:
- `SchemaInstantiation` — already handled in Go (`CompileTactic` returns `n, nil`)
- **`AssumeTactic`** — NOT handled in Go (falls through to default)
- **`AssumeGlobalTactic`** — NOT handled in Go (falls through to default)

When the outer `ComposeTactics` compiles its children:
- **Python**: `a.compile()` on an AssumeTactic calls `compile_schema_instantiation` → returns self, **no traces**
- **Go**: `c.CompileTactic(a)` on an AssumeTactic hits the default case → emits Thing ENTER, CompileNode ENTER, CompileNode return default, OtherThing ENTER, OtherThing return, Thing return traces — **6 extra trace lines**

This shifts Go's trace output relative to Python's, causing the divergence.

## Fix

Add explicit cases for `*ast.AssumeTactic` and `*ast.AssumeGlobalTactic` in `CompileTactic`, matching the existing `SchemaInstantiation` pattern:

**File**: `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/compiler.go` (around line 1310)

```go
case *ast.SchemaInstantiation:
    // Python: compile_schema_instantiation returns self
    return n, nil

case *ast.AssumeTactic:
    // Python: TacticWithMatch.compile = compile_schema_instantiation → return self
    return n, nil

case *ast.AssumeGlobalTactic:
    // Python: TacticWithMatch.compile = compile_schema_instantiation → return self
    return n, nil
```

## Verification

Run `cd ~/goivy && make golden` and confirm the divergence at line 92279 is resolved (traces should match further or fully).
