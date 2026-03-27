Now I have a thorough understanding of both type systems. Let me write up the analysis.

---

## Type System Analysis: Python Ivy vs Go goivy

### Python's Model: Everything is AST

Python Ivy has a single unified hierarchy:

```
ivy_ast.AST (base: .args, .clone(), .compile())
├── Formula nodes: And, Or, Not, Implies, Atom, Variable, ...
│   └── compile() → thing() → cmpl() → ivy_logic.And/Or/... (type conversion)
└── Action (mutable .args)
    ├── Sequence, AssignAction, IfAction, WhileAction, ...
    │   └── compile() → thing() → cmpl() → cloned self with compiled children (SAME type)
    └── formal_params, formal_returns: set dynamically on any action

ivy_logic.AST (separate base, same .compile = thing protocol)
├── And, Or, Symbol, Apply, ...
└── These ARE the compiled formulas — live alongside actions seamlessly
```

The critical design: `thing()` takes any AST node, returns any AST node. There is **no type boundary** between formulas and actions. `compile_action_def` does:

```python
res = sortify(a.args[1])     # returns Sequence (action) or anything
res.formal_params = formals   # duck-types formal_params onto it
return res                    # caller doesn't care if it's action or formula
```

### Go's Model: Three Separate Hierarchies

```
ast.Node (parser AST)
│   Args() []Node, Clone([]Node) Node, Canon(), ...
│
├── ast.Atom, ast.And, ast.Sequence, ast.AssignAction, ...
│   (raw parser output — not "compiled")

lg.Expr (compiled logic)  [embeds ast.Node]
│   + NodeSort() Sort, Children() []Expr, Equal(), Sexp()
│
├── lg.Symbol, lg.Apply, lg.And, lg.Or, ...
│   (compiled formulas — strong typing on sorts)

actions.Action (compiled actions)  [SEPARATE hierarchy]
│   + ActionArgs() []lg.Expr, ActionClone(), GetFormalParams(), ...
│
├── actions.Sequence, actions.IfAction, actions.AssignAction, ...
│   (compiled actions — store children as []lg.Expr)

ActionNodeWrapper (bridge)
│   implements lg.Expr, wraps actions.Action
│   WrapAction() / UnwrapAction() conversion functions
```

### Why Go's Model Creates Contortions

The root cause: `Thing()` returns `lg.Expr`, but compilation can produce `actions.Action`. These are different interfaces with different method sets, connected only by the `ActionNodeWrapper` bridge.

This forces contortions at every boundary:

1. **`CompileNode` action cases** must `WrapAction(act)` to return `lg.Expr`
2. **`CompileActionBody` default case** must `UnwrapAction(result)` to get back to `actions.Action`
3. **`compileGeneric`** clones `ast.Sequence` → gets `*ast.Sequence` which doesn't implement `lg.Expr` → falls to extraction fallback → returns `lg.And` (wrong semantics) → caller must detect and reconstruct
4. **Everywhere** actions are stored: wrapped as `lg.Expr` via `ActionNodeWrapper`, losing type information

### The Key Insight

`actions.Sequence` **already stores children as `[]lg.Expr`**. The `actions` package **already imports `logic`**. There is no circular dependency preventing actions from implementing `lg.Expr`.

The barrier is that actions don't implement `ast.Node` (they have `ActionBase` not `ast.Base`) and don't implement `lg.Expr` (they lack `NodeSort`, `Children`, `Equal`, `Sexp`).

### Recommended Solution: Make Actions Implement `lg.Expr`

Have all action types satisfy both `actions.Action` AND `lg.Expr`. This is the Go equivalent of Python's "everything is AST" — actions become first-class participants in the expression hierarchy.

**What's needed per action type:**

For `ast.Node` compliance:
```go
func (s *Sequence) Args() []ast.Node {
    // Children are lg.Expr which embed ast.Node
    nodes := make([]ast.Node, len(s.Children))
    for i, c := range s.Children { nodes[i] = c }
    return nodes
}
func (s *Sequence) Clone(args []ast.Node) ast.Node {
    exprs := make([]lg.Expr, len(args))
    for i, a := range args { exprs[i] = a.(lg.Expr) }
    return s.ActionClone(exprs).(ast.Node)
}
func (s *Sequence) Canon() iu.Canonical { ... }
func (s *Sequence) GetAstConfig() *ast.AstConfig { return nil }
```

For `lg.Expr` compliance:
```go
var ActionSort lg.Sort = &actionSort{} // sentinel sort for all actions

func (s *Sequence) NodeSort() lg.Sort    { return ActionSort }
func (s *Sequence) Children() []lg.Expr  { return s.Children }
func (s *Sequence) Equal(other lg.Expr) bool { return s.Sexp() == other.Sexp() }
func (s *Sequence) Sexp() lg.NodeKey     { /* structural key */ }
```

**Boilerplate reduction** — add a helper base:
```go
// ActionExpr embeds ActionBase and provides the ast.Node + lg.Expr
// boilerplate so concrete action types only override what differs.
type ActionExpr struct {
    ActionBase
}
func (a *ActionExpr) NodeSort() lg.Sort         { return ActionSort }
func (a *ActionExpr) GetAstConfig() *ast.AstConfig { return nil }
// ... other shared methods
```

### What This Eliminates

With actions implementing `lg.Expr`:

1. **`WrapAction`/`UnwrapAction` become unnecessary** — actions ARE `lg.Expr` directly
2. **`CompileNode` action cases** just return the action (it's already `lg.Expr`)
3. **`compileGeneric`** clones with compiled children → `Clone` returns an action → action implements `lg.Expr` → returned directly. No extraction fallback needed
4. **`CompileActionBody`** can type-assert `result.(actions.Action)` directly — no unwrapping
5. **`ActionNodeWrapper`** becomes obsolete

### Concern: Formula Code Receiving Actions

If actions implement `lg.Expr`, formula-processing code (sort inference, simplification) might receive action nodes. **Python has this same situation** and handles it fine — duck typing means code just deals with what it gets.

In Go, guard with:
```go
if e.NodeSort() == ActionSort {
    // skip formula-specific processing
}
```

Or use a type assertion: `if _, isAction := e.(actions.Action); isAction { ... }`

### Migration Path

This is a significant but bounded change (~15 action types). Recommended approach:

1. Define `ActionSort` in the `logic` package
2. Add `ActionExpr` helper base in `actions` package with shared `ast.Node` + `lg.Expr` methods
3. Migrate action types one at a time (Sequence first as proof of concept)
4. Once all actions implement `lg.Expr`, remove `ActionNodeWrapper` and `WrapAction`/`UnwrapAction`
5. Simplify `compileGeneric`, `CompileNode`, `CompileActionBody` to remove wrapping contortions

### Why NOT to Have `ast.Sequence` Implement `lg.Expr`

`ast.Sequence` is in the `ast` package, which **cannot import `logic`** (would create a circular dependency: `ast` → `logic` → `ast`). So `ast.Sequence` can never implement `lg.Expr`. This is fine — `ast.Sequence` is the *parser* output. The *compiled* result should be `actions.Sequence` (which can implement `lg.Expr`). Python has one Sequence type serving both roles; Go needs two, but the compiled one should be a full `lg.Expr` citizen.
