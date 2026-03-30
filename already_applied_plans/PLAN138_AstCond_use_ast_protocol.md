# Plan: Fix IfAction SomeCondition vs AST Some/SomeMin/SomeMax Divergence

**Created**: 2026-03-30 04:45

## Context

Golden test diverges at line 149327:
```
go : XTRACE: actions.substitute_constants_action ENTER type=SomeCondition nargs=0
py : XTRACE: actions.substitute_constants_action ENTER type=SomeMin nargs=3
```

Same pattern as the CallAction/Atom fix: Go stores a compiled wrapper (`SomeCondition`, `Args()→nil`) where Python stores the cloned AST node (`SomeMin`, 3 children). During `substitute_constants_action` recursion, Go sees 0 children and skips, Python sees 3 and recurses.

**Python** (ivy_compiler.py:747): `args = [self.args[0].clone(sargs), ...]` — clones the original AST Some/SomeMin/SomeMax with compiled args. The type is preserved.

**Go** (compiler/action.go:1291): Creates `&actions.SomeCondition{...}` — a new compiled wrapper type that hides children.

## Fix: Same pattern as CallAction — add `AstCond` field

### Step 1: Add `AstCond` field to IfAction

**File**: `actions/action.go` (~line 358)

```go
type IfAction struct {
    ActionBase
    Cond     lg.Expr   // compiled condition (SomeCondition for existentials, lg.Expr otherwise)
    AstCond  ast.Node  // AST condition for tree-walking (Some/SomeMin/SomeMax when present)
    ThenBody lg.Expr
    ElseBody lg.Expr
}
```

### Step 2: Change `IfAction.Args()` to return AstCond when set

**File**: `actions/action_expr.go` (~line 291)

```go
func (a *IfAction) Args() []ast.Node {
    first := ast.Node(a.Cond)
    if a.AstCond != nil {
        first = a.AstCond
    }
    if a.ElseBody != nil {
        return []ast.Node{first, a.ThenBody, a.ElseBody}
    }
    return []ast.Node{first, a.ThenBody}
}
```

### Step 3: Change `IfAction.Clone()` to handle AST Some nodes

**File**: `actions/action_expr.go` (~line 292-294)

```go
func (a *IfAction) Clone(args []ast.Node) ast.Node {
    var cond lg.Expr
    var astCond ast.Node

    switch s := args[0].(type) {
    case *ast.Some, *ast.SomeMin, *ast.SomeMax:
        astCond = s
        cond = someCondFromAST(s)
    default:
        cond = args[0].(lg.Expr)
        astCond = a.AstCond
    }

    thenBody := args[1].(lg.Expr)
    var elseBody lg.Expr
    if len(args) >= 3 {
        elseBody = args[2].(lg.Expr)
    }

    var res *IfAction
    if elseBody != nil {
        res = NewIfAction(cond, thenBody, elseBody)
    } else {
        res = NewIfAction(cond, thenBody)
    }
    res.AstCond = astCond
    res.ActionBase = a.ActionBase
    return res
}
```

### Step 4: Add `someCondFromAST` helper

**File**: `actions/action.go` (near SomeCondition definition)

Reconstructs a `SomeCondition` from a cloned AST Some/SomeMin/SomeMax node.

```go
func someCondFromAST(node ast.Node) *SomeCondition {
    switch s := node.(type) {
    case *ast.Some:
        params := make([]*lg.Const, len(s.Params))
        for i, p := range s.Params {
            params[i] = p.(*lg.Const)
        }
        return &SomeCondition{Params: params, Fmla: s.Fmla.(lg.Expr), Kind: "some"}
    case *ast.SomeMin:
        params := make([]*lg.Const, len(s.Params))
        for i, p := range s.Params {
            params[i] = p.(*lg.Const)
        }
        return &SomeCondition{Params: params, Fmla: s.Fmla.(lg.Expr), Kind: "some_min", Index: s.Index.(lg.Expr)}
    case *ast.SomeMax:
        params := make([]*lg.Const, len(s.Params))
        for i, p := range s.Params {
            params[i] = p.(*lg.Const)
        }
        return &SomeCondition{Params: params, Fmla: s.Fmla.(lg.Expr), Kind: "some_max", Index: s.Index.(lg.Expr)}
    }
    return nil
}
```

### Step 5: Set `AstCond` in compiler

**File**: `compiler/action.go` (~line 1290-1322, in `compileIfSome`)

After creating the SomeCondition (line 1291-1296), clone the original AST node with compiled args, matching Python's `self.args[0].clone(sargs)`:

```go
// 5. Build SomeCondition (compiled form)
someCond := &actions.SomeCondition{
    Params: compiledParams,
    Fmla:   sfmla,
    Kind:   kind,
    Index:  index,
}

// 5b. Clone original AST node with compiled args (matches Python: self.args[0].clone(sargs))
sargs := make([]ast.Node, 0, len(compiledParams)+2)
for _, p := range compiledParams {
    sargs = append(sargs, p)
}
sargs = append(sargs, sfmla)
if index != nil {
    sargs = append(sargs, index)
}
astCond := condNode.Clone(sargs)
```

Then after creating the IfAction (lines 1313-1321):
```go
res.AstCond = astCond
```

### Step 6: Preserve AstCond in ActionClone

**File**: `actions/action.go`, IfAction.ActionClone (~line 380)

Currently:
```go
func (a *IfAction) ActionClone(args []lg.Expr) Action {
    if len(args) >= 3 {
        r := NewIfAction(args[0], args[1], args[2])
        ...
    }
    r := NewIfAction(args[0], args[1])
    ...
}
```

Add `r.AstCond = a.AstCond` after each creation (for the ActionClone path used by Sexp/Canon).

## Files to modify

1. **`actions/action.go`** — Add `AstCond` field to IfAction; add `someCondFromAST` helper; preserve AstCond in ActionClone
2. **`actions/action_expr.go`** — Change `IfAction.Args()` and `IfAction.Clone()`
3. **`compiler/action.go`** — Clone AST node and set `res.AstCond` in `compileIfSome`

## What stays unchanged

- `Cond lg.Expr` field — kept for internal use by `Subactions`, `subactionsSome`, `action_update`
- `SomeCondition` type — still needed for compiled-level processing
- `ActionArgs()` — returns `[Cond, ThenBody, ElseBody?]` for Sexp/Canon consumers
- `WhileAction` — also has `Cond` but doesn't use `SomeCondition` as condition type (separate investigation if needed)

## Verification

`go build ./...` then `cd ~/goivy && make golden`
