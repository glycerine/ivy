# Make Actions Implement lg.Expr — Eliminate WrapAction/UnwrapAction

**Created:** 2026-03-27

## Context

Python Ivy has a single type hierarchy: actions ARE AST nodes. `thing()` returns the same base type regardless of whether the result is a formula or an action. Go goivy has three separate hierarchies (`ast.Node`, `lg.Expr`, `actions.Action`) connected by `WrapAction`/`UnwrapAction` bridges. This forces contortions throughout the compiler: wrapping actions to store them as `lg.Expr`, unwrapping them to get actions back, and handling failed clones in `compileGeneric`.

The fix: make all action types implement `lg.Expr` (and `ast.Node`) directly. Actions become first-class `lg.Expr` citizens. `WrapAction`/`UnwrapAction` become unnecessary.

## Architecture

### Self-reference pattern for virtual dispatch

Go's embedding doesn't support virtual dispatch — `ActionBase.Args()` can't call `Sequence.ActionArgs()`. We solve this with a `self_` reference on ActionBase:

```go
type ActionBase struct {
    // ... existing fields ...
    self_ Action  // self-reference for virtual dispatch
}
```

Every constructor and `ActionClone` calls `InitSelf(self)` to set this. Then shared methods on ActionBase can delegate to the concrete type's `ActionArgs()`/`ActionClone()`.

### Method distribution

**On ActionBase (shared, 8 methods):**
| Method | Implementation |
|--------|---------------|
| `Args() []ast.Node` | converts `self_.ActionArgs()` to `[]ast.Node` |
| `Clone([]ast.Node) ast.Node` | converts args, calls `self_.ActionClone()` |
| `GetAstConfig() *ast.AstConfig` | returns `nil` |
| `NodeSort() lg.Sort` | returns `lg.Boolean` (matches current wrapper) |
| `Children() []lg.Expr` | returns `self_.ActionArgs()` |
| `Equal(lg.Expr) bool` | compares `Sexp()` values |
| `Canon() iu.Canonical` | returns `iu.Canonical(self_Sexp())` |
| `InitSelf(Action)` | sets `self_` |

**Per concrete type (1 method each, 30 types):**
| Method | Implementation |
|--------|---------------|
| `Sexp() lg.NodeKey` | type-specific structural key |

`String()` — already defined. `GetLineno()`/`SetLineno()` — already on ActionBase.

### Action interface change

Modify `Action` to embed `ast.Node` and `lg.Expr`:
```go
type Action interface {
    ast.Node
    lg.Expr
    // ... existing Action-specific methods ...
    ActionArgs() []lg.Expr
    ActionClone([]lg.Expr) Action
    // ...
}
```

This makes every `Action` automatically usable wherever `ast.Node` or `lg.Expr` is expected — no wrapping needed.

## Implementation Steps

### Step 1: Add `ActionSort` to logic package

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/logic/sort.go`

Add a sentinel sort for actions (needed since Sort has unexported `sortSeal()`):
```go
// ActionSort is the sort for compiled action nodes.
// Used when actions implement lg.Expr — actions don't have a
// meaningful logic sort, but they need one to satisfy the interface.
var ActionS Sort = &actionSort{}

type actionSort struct{}
func (s *actionSort) sortSeal()              {}
func (s *actionSort) NodeSort() Sort         { return s }
func (s *actionSort) Children() []Expr       { return nil }
func (s *actionSort) Equal(n Expr) bool      { _, ok := n.(*actionSort); return ok }
func (s *actionSort) Sexp() NodeKey          { return "(ActionSort)" }
func (s *actionSort) Args() []ast.Node       { return nil }
func (s *actionSort) Clone([]ast.Node) ast.Node { return s }
func (s *actionSort) GetLineno() ast.Location { return ast.Location{} }
func (s *actionSort) SetLineno(ast.Location)  {}
func (s *actionSort) String() string          { return "ActionSort" }
func (s *actionSort) Canon() iu.Canonical     { return "ActionSort" }
func (s *actionSort) GetAstConfig() *ast.AstConfig { return nil }
```

### Step 2: Add shared methods to ActionBase

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/module/action.go`

Add import for `ivyutils` (for `iu.Canonical`).

Add `self_` field and helper methods:
```go
type ActionBase struct {
    Loc           ast.Location
    HasLoc        bool
    FormalParams  []*lg.Symbol
    FormalReturns []*lg.Symbol
    Labels        []string
    self_         Action  // self-reference for virtual dispatch
}

// InitSelf sets the self-reference for virtual dispatch.
// Must be called by every constructor and ActionClone.
func (b *ActionBase) InitSelf(a Action) { b.self_ = a }

// --- ast.Node methods (virtual dispatch via self_) ---

func (b *ActionBase) Args() []ast.Node {
    if b.self_ == nil { return nil }
    args := b.self_.ActionArgs()
    nodes := make([]ast.Node, len(args))
    for i, e := range args { nodes[i] = e }
    return nodes
}

func (b *ActionBase) Clone(args []ast.Node) ast.Node {
    if b.self_ == nil { return nil }
    exprs := make([]lg.Expr, len(args))
    for i, a := range args { exprs[i] = a.(lg.Expr) }
    return b.self_.ActionClone(exprs).(ast.Node)
}

func (b *ActionBase) GetAstConfig() *ast.AstConfig { return nil }

// Canon delegates to Sexp() on the concrete type.
func (b *ActionBase) Canon() iu.Canonical {
    if b.self_ == nil { return "" }
    if s, ok := b.self_.(lg.Expr); ok {
        return iu.Canonical(s.Sexp())
    }
    return ""
}

// --- lg.Expr methods ---

func (b *ActionBase) NodeSort() lg.Sort { return lg.ActionS }

func (b *ActionBase) Children() []lg.Expr {
    if b.self_ == nil { return nil }
    return b.self_.ActionArgs()
}

func (b *ActionBase) Equal(other lg.Expr) bool {
    if b.self_ == nil { return false }
    if s, ok := b.self_.(lg.Expr); ok {
        return s.Sexp() == other.Sexp()
    }
    return false
}
```

### Step 3: Add `Sexp()` to each concrete action type

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/actions/action.go` (and `actions/phase3.go`)

Add `Sexp()` to all 30 concrete action types. Pattern:

```go
func (s *Sequence) Sexp() lg.NodeKey {
    parts := make([]string, len(s.Children))
    for i, c := range s.Children { parts[i] = string(c.Sexp()) }
    return lg.NodeKey("(Sequence children:[" + strings.Join(parts, " ") + "])")
}

func (a *AssumeAction) Sexp() lg.NodeKey {
    return lg.NodeKey(fmt.Sprintf("(AssumeAction formula:%v)", a.Formula.Sexp()))
}

func (a *AssignAction) Sexp() lg.NodeKey {
    return lg.NodeKey(fmt.Sprintf("(AssignAction lhs:%v rhs:%v)", a.LHS.Sexp(), a.RHS.Sexp()))
}
// ... etc for all 30 types
```

### Step 4: Add `InitSelf` calls to all constructors and ActionClone methods

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/actions/action.go` (and `actions/phase3.go`)

For every `New*Action()` constructor, add `result.InitSelf(result)` before return:
```go
func NewSequence(args ...lg.Expr) *Sequence {
    s := &Sequence{Children: copyNodes(args)}
    s.InitSelf(s)  // ADD THIS
    return s
}
```

For every `ActionClone()`, add `InitSelf` on the new instance:
```go
func (s *Sequence) ActionClone(args []lg.Expr) Action {
    r := &Sequence{ActionBase: s.ActionBase, Children: copyNodes(args)}
    r.InitSelf(r)  // ADD THIS
    return r
}
```

**30 constructors + 30 ActionClone methods = 60 `InitSelf` additions.**

### Step 5: Modify Action interface to embed ast.Node and lg.Expr

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/module/action.go`

```go
type Action interface {
    ast.Node   // Args, Clone, GetLineno, SetLineno, String, Canon, GetAstConfig
    lg.Expr    // NodeSort, Children, Equal, Sexp (subsumes ast.Node)

    // Action-specific methods:
    ActionClone(args []lg.Expr) Action
    ActionArgs() []lg.Expr
    IterCalls() []string
    IterSubactions() []Action
    GetFormalParams() []*lg.Symbol
    GetFormalReturns() []*lg.Symbol
    SetFormalParams([]*lg.Symbol)
    SetFormalReturns([]*lg.Symbol)
    Name() string
    Decompose() [][]Action
}
```

Note: since `lg.Expr` embeds `ast.Node`, we only need `lg.Expr`. But listing both makes intent clear. The compiler will deduplicate.

### Step 6: Eliminate WrapAction — actions ARE lg.Expr

**All files using WrapAction (~50 locations):**

Replace `WrapAction(action)` with just `action` everywhere. Since `Action` now implements `lg.Expr`, actions can be stored directly in `[]lg.Expr` fields.

Example transformations:
```go
// BEFORE:
nodes[i] = actions.WrapAction(s)
// AFTER:
nodes[i] = s

// BEFORE:
seq := actions.NewSequence(actions.WrapAction(assertAct))
// AFTER:
seq := actions.NewSequence(assertAct)

// BEFORE:
ifAction := NewIfAction(cond, WrapAction(thenSeq), WrapAction(elseSeq))
// AFTER:
ifAction := NewIfAction(cond, thenSeq, elseSeq)
```

**Files affected** (in rough order of importance):
1. `actions/action.go` — internal uses (~10)
2. `actions/helpers.go` — 6 uses
3. `actions/match.go` — 13 uses
4. `actions/update.go` — ~8 uses
5. `actions/transforms.go` — ~5 uses
6. `compiler/action.go` — ~20 uses
7. `compiler/phase6.go` — 2 uses
8. `compiler/ivy_compile.go` — uses
9. `isolate/strip.go` — 6 uses
10. `check/check.go` — 1 use
11. `bmc/bmc.go` — 1 use
12. `tactics/tactics.go` — 1 use
13. `l2s/shared.go`, `l2s/l2s.go` — 2 uses each
14. `temporal/temporal.go` — 1 use
15. Test files: `actions/audit51_test.go`, `actions/audit53_test.go`, `actions/actions_test.go`, `check/check_test.go`, `check/regression_test.go`, `leangen/leangen_test.go`

### Step 7: Eliminate UnwrapAction — type-assert directly

**All files using UnwrapAction (~40 locations):**

Replace `UnwrapAction(expr)` with `expr.(Action)` or a type switch:
```go
// BEFORE:
if act := actions.UnwrapAction(result); act != nil {
    return act, nil
}
// AFTER:
if act, ok := result.(actions.Action); ok {
    return act, nil
}

// BEFORE (unwrapToAction helper):
func unwrapToAction(n lg.Expr) Action {
    if act, ok := n.(Action); ok { return act }
    if w, ok := n.(*ActionNodeWrapper); ok { return w.Action }
    return nil
}
// AFTER:
func unwrapToAction(n lg.Expr) Action {
    if act, ok := n.(Action); ok { return act }
    return nil
}
```

### Step 8: Simplify compiler — remove compileGeneric contortion

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/action.go`

The `*ast.Sequence` case in `CompileActionBody` no longer needs the `lg.And` extraction hack:
```go
case *ast.Sequence:
    result, err := c.Thing(node)
    if err != nil { return nil, err }
    if act, ok := result.(actions.Action); ok {
        return act, nil  // Direct — actions ARE lg.Expr now
    }
    return actions.NewSequence(), nil
```

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/compiler/compiler.go`

`CompileNode` action cases no longer need to wrap:
```go
case *ast.AssignAction, ...:
    act, err := c.CompileActionBody(node)
    if err != nil { return nil, err }
    return act, nil  // act implements lg.Expr directly
```

`compileGeneric` should work naturally — `node.Clone(compiled)` for `*ast.Sequence` returns `*ast.Sequence` (still doesn't implement `lg.Expr`). But the children (compiled via `Thing`) are now `lg.Expr` values that might be actions. The fallback extraction path creates `lg.And` or returns a single expr. For Sequence, this expr IS the action — `UnwrapAction` (now just type assertion) extracts it.

Actually, `compileGeneric` still has the clone issue for `*ast.Sequence` since it's an ast type, not an actions type. But the `*ast.Sequence` case in `CompileActionBody` handles this before `compileGeneric` is reached... wait, no, it calls `Thing` which calls `CompileNode` which falls to OtherThing which calls `compileGeneric`. The issue persists for `ast.Sequence`.

The improvement is: when `compileGeneric` extracts children and they're actions (now `lg.Expr`), the single-child case returns the action directly. The multi-child case returns `lg.And` — but now `CompileActionBody` can type-assert children from `lg.And` to `Action` without unwrapping.

### Step 9: Remove ActionNodeWrapper (final cleanup)

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/module/action.go`

Delete or deprecate:
- `ActionNodeWrapper` struct
- `WrapAction()` function
- `UnwrapAction()` function
- `ToAction()` function

**File:** `/Users/jaten/go/src/github.com/glycerine/goivy/actions/action.go`

Remove type aliases:
```go
// DELETE:
type ActionNodeWrapper = module.ActionNodeWrapper
var (
    WrapAction   = module.WrapAction
    UnwrapAction = module.UnwrapAction
)
```

## Complete list of 30 action types needing Sexp() + InitSelf

| # | Type | File | Children pattern |
|---|------|------|-----------------|
| 1 | Sequence | action.go | `Children []lg.Expr` |
| 2 | AssumeAction | action.go | `Formula lg.Expr` |
| 3 | AssertAction | action.go | `Formula, Proof lg.Expr` |
| 4 | RequiresAction | action.go | embeds AssertAction |
| 5 | EnsuresAction | action.go | embeds AssertAction |
| 6 | AssignAction | action.go | `LHS, RHS lg.Expr` |
| 7 | HavocAction | action.go | `Target lg.Expr` |
| 8 | SetAction | action.go | `Lit lg.Expr` |
| 9 | IfAction | action.go | `Cond, ThenBody, ElseBody lg.Expr` |
| 10 | WhileAction | action.go | `Cond, Body lg.Expr; Invariants []lg.Expr` |
| 11 | ChoiceAction | action.go | `Branches []lg.Expr` |
| 12 | CallAction | action.go | `Callee lg.Expr; ActualReturns []lg.Expr` |
| 13 | LocalAction | action.go | `Locals []lg.Expr; Body lg.Expr` |
| 14 | LetAction | action.go | `Bindings []lg.Expr; Body lg.Expr` |
| 15 | BindOldsAction | action.go | `Inner lg.Expr` |
| 16 | NativeAction | action.go | `Code lg.Expr; Params []lg.Expr` |
| 17 | CrashAction | action.go | `Target lg.Expr` |
| 18 | ThunkAction | action.go | `Children []lg.Expr` |
| 19 | EnvAction | action.go | embeds ChoiceAction |
| 20 | ReturnAction | action.go | no children |
| 21 | IgnoreAction | action.go | no children |
| 22 | SubgoalAction | action.go | embeds AssertAction |
| 23 | AssignFieldAction | action.go | `Field, Obj, Value lg.Expr` |
| 24 | NullFieldAction | action.go | `Field, Obj lg.Expr` |
| 25 | CopyFieldAction | action.go | `Dst, Field, Src, SrcField lg.Expr` |
| 26 | Ranking | action.go | `Relation lg.Expr; RArgs []lg.Expr` |
| 27 | PatternBasedUpdate | action.go | no lg.Expr children |
| 28 | NamedUpdate | action.go | `Body lg.Expr` |
| 29 | InstantiateAction | action.go | `Inst lg.Expr` |
| 30 | DebugAction | phase3.go | `DebugExpr lg.Expr; WithExprs []lg.Expr` |

## Verification

After each step, run:
```bash
go build ./...                    # compilation check
go test ./actions/ ./compiler/    # unit tests
cd ~/goivy && make golden         # golden test (after all steps)
```

## Risk mitigation

- **Forgotten InitSelf**: Would cause nil pointer in Args()/Clone()/Children(). Caught immediately by any test that touches the action.
- **Circular dispatch**: `ActionBase.Clone()` calls `self_.ActionClone()`. ActionClone creates a new action and calls InitSelf. No circularity.
- **Embedding hierarchy**: RequiresAction embeds AssertAction which embeds ActionBase. InitSelf on ActionBase is inherited — calling `r.InitSelf(r)` in RequiresAction.ActionClone sets the base's `self_` to the RequiresAction instance. Virtual dispatch works correctly.
- **Serialization**: The `self_` field should be tagged `json:"-"` to avoid serialization issues.
