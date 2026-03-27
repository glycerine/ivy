# Fix xtrace divergence at line 124589: Route Sequence and And through real Thing

**Created:** 2026-03-27

## Context

Golden test diverges at xtrace line 124589:
- **Go:** `compiler.ARGSetup.action EXIT name=index.next key=index.next`
- **Python:** `compiler.Thing ENTER type=Sequence`

Go's `CompileActionBody` handles `*ast.Sequence` and `*ast.And` with its own logic, bypassing `Thing`/`CompileNode`/`OtherThing`. Python routes ALL compilation through `thing()`.

## Python's compilation dispatch for these types

**And** — has explicit `.cmpl` (set via `op_pairs` loop at ivy_compiler.py:121,147-148):
- `thing(And)` → `And.cmpl()` → `_make_op_cmpl("And", ivy_logic.And)` → compiles as **logical conjunction**
- Traces: `Thing ENTER type=And` → `CompileNode ENTER type=And` → `CompileNode return case=And` → `CompileAnd ENTER`

**Sequence** — NO explicit `.cmpl`, uses default `other_thing` (ivy_compiler.py:118):
- `thing(Sequence)` → `other_thing()` → `self.clone([a.compile() for a in self.args])`
- Traces: `Thing ENTER type=Sequence` → `CompileNode ENTER type=Sequence` → `CompileNode return case=default type=Sequence` → `OtherThing ENTER type=Sequence`

## Go's current behavior (wrong per Python)

- `*ast.And` in `CompileActionBody`: treats as statement sequence, recursively calls `CompileActionBody` on children. **Wrong** — Python compiles And as conjunction.
- `*ast.Sequence` in `CompileActionBody`: treats as statement sequence with its own logic, bypasses Thing entirely. **Missing traces** — Python routes through thing → other_thing.

## Go's existing infrastructure (already correct)

- `CompileNode` line 196: `case *ast.And` → `compileAnd(n)` — compiles as conjunction. **Matches Python.**
- `CompileNode` default → `OtherThing` → `compileGeneric` — clones with compiled children. **Matches Python's other_thing.**
- `CompileActionBody` default case: calls `Thing(node)`, unwraps action or wraps as AssumeAction. **Works for And.**

## Fix

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/action.go`

### Fix 1: Remove `*ast.And` case from CompileActionBody

Remove the `case *ast.And:` block (lines 180-205). Let And fall to the **default** case which already does:
```go
compiled, err := c.Thing(node)
// ... unwrap or wrap as AssumeAction
```

This routes And through `Thing` → `CompileNode` → `compileAnd` (conjunction), matching Python exactly. The result (conjunction formula) gets wrapped as AssumeAction, which has `SetFormalParams`/`SetFormalReturns`.

### Fix 2: Replace `*ast.Sequence` case to route through Thing

Replace the `case *ast.Sequence:` block (lines 411-435) with:
```go
case *ast.Sequence:
    // Python: Sequence has no .cmpl, uses other_thing (default):
    //   thing() → CompileNode (default) → OtherThing → compileGeneric
    //   compileGeneric: self.clone([a.compile() for a in self.args])
    // Route through REAL Thing for genuine traces and compilation.
    result, err := c.Thing(node)
    if err != nil {
        return nil, err
    }
    // Single child: compileGeneric returns it directly (wrapped action)
    if act := actions.UnwrapAction(result); act != nil {
        return act, nil
    }
    // Multiple children: compileGeneric can't clone ast.Sequence as lg.Expr,
    // so it extracts children and returns lg.And{Terms: [wrappedActions...]}.
    // Convert to actions.Sequence.
    if andExpr, ok := result.(*lg.And); ok {
        seq := actions.NewSequence(andExpr.Terms...)
        seq.SetLineno(node.GetLineno())
        return seq, nil
    }
    // Zero children or unexpected
    return actions.NewSequence(), nil
```

**Why the multi-child handling is needed:** `compileGeneric` calls `Thing` on each Sequence child, gets `lg.Expr` results. It tries `node.Clone(compiled)` but `*ast.Sequence` doesn't implement `lg.Expr`, so clone can't be returned directly. It falls to extracting `lg.Expr` children and combining them as `lg.And`. We then unwrap this back into `actions.NewSequence`.

### Trace flow after fix (for Sequence)

1. `CompileActionBody` → `case *ast.Sequence` → `c.Thing(node)`
2. `Thing` → traces `Thing ENTER type=Sequence` → `CompileNode`
3. `CompileNode` → traces `CompileNode ENTER type=Sequence` → no Sequence case → default → traces `CompileNode return case=default type=Sequence` → `OtherThing`
4. `OtherThing` → traces `OtherThing ENTER type=Sequence` → `compileGeneric`
5. `compileGeneric` → `Thing(child)` for each child → real recursive compilation
6. Returns → traces `OtherThing return type=Sequence sort_infer_root=false`
7. Returns → traces `Thing return type=Sequence`
8. `CompileActionBody` converts result to `actions.Action`

All traces are REAL — emitted by the actual functions being called.

## Key files
- `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/action.go` — CompileActionBody (main changes)
- `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/compiler.go` — CompileNode, compileGeneric (unchanged, already correct)
- `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/phase6.go` — Thing, OtherThing (unchanged, already correct)

## Verification

Run `cd ~/goivy && make golden` and confirm the test advances past line 124589.
Run `go test ./compiler/` to verify no regressions.

---------

All builds and tests pass. Summary of changes: 
   
1. compiler/action.go — CompileActionBody:
- Removed *ast.And case: And now falls to the default case which routes through 
  Thing → CompileNode → compileAnd (logical conjunction), matching Python where
  And.cmpl compiles as conjunction. 
  
- Replaced *ast.Sequence case: Now routes through c.Thing(node) which goes 
  through the real Thing → CompileNode (default) → OtherThing → compileGeneric path.
   Handles the result by unwrapping single actions or extracting multiple children
  from the lg.And that compileGeneric produces.  
  
2. compiler/compiler.go — CompileNode: 
- Added missing action types to the action routing case: *ast.CrashAction,
  *ast.ThunkAction, *ast.LocalAction, *ast.InstantiateDecl. These must 
  be listed so that Thing → CompileNode routes them to CompileActionBody when encountered during compileGeneric's recursive compilation of Sequence children.  
  
  
