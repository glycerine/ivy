# Goivy Type Flow Guide

This document describes how data flows through the goivy compiler,
the interfaces and concrete types at each stage, and how the Go
implementation relates to the Python Ivy source of truth.

## 1. The Two Type Universes

Goivy has two distinct type hierarchies that mirror Python Ivy's
two module namespaces:

### ast.Node — The Parse-Time Universe

Package `ast/` defines the **parsed AST**. Every node implements
`ast.Node`:

```go
// ast/ast.go
type Node interface {
    Args() []Node
    Clone(args []Node) Node
    GetLineno() Location
    SetLineno(Location)
    String() string
    Canon() iu.Canonical
    GetAstConfig() *AstConfig
}
```

Examples: `*ast.Atom`, `*ast.And`, `*ast.Variable`, `*ast.Forall`,
`*ast.LabeledFormula`, `*ast.AssertAction`, `*ast.Definition`.

These correspond to Python's `ivy_ast` module classes. They carry
source location, string representations, and child nodes, but they
have **no sort information** — an `ast.Atom("x")` doesn't know
whether `x` is a boolean, an integer, or a function.

### logic.Expr — The Compiled-Logic Universe

Package `logic/` defines the **sorted logic IR**. Every node
implements `logic.Expr`:

```go
// logic/node.go
type Expr interface {
    ast.Node         // every Expr IS an ast.Node
    NodeSort() Sort  // the resolved sort (Boolean, FunctionSort, etc.)
    Children() []Expr
    Equal(Expr) bool
    Sexp() NodeKey   // structural identity for hashing/comparison
}
```

Examples: `*logic.Apply`, `*logic.Symbol`, `*logic.Variable`,
`*logic.And`, `*logic.Definition`, `*logic.ForAll`.

These correspond to Python's `ivy_logic` module classes. They carry
**fully resolved sort information** — a `logic.Symbol` knows its
`FunctionSort`, a `logic.And` knows it returns `Boolean`.

The critical relationship: **`logic.Expr` embeds `ast.Node`**.
This means every `logic.Expr` value can be stored in an `ast.Node`
slot without any wrapping or conversion. The reverse is not true —
an `ast.Atom` is NOT a `logic.Expr`.

### Why Two Universes?

Python Ivy has the same split: `ivy_ast.And` (parsed) vs
`ivy_logic.And` (sorted). Because Python is duck-typed, both
classes have `.args` and `.clone()`, and code freely mixes them.
Go's interfaces make the boundary explicit. This is a feature:
the compiler enforces at build time that you don't accidentally
pass an unsorted AST node to a function expecting sorted logic.


## 2. The Compilation Pipeline

The compiler transforms `ast.Node` trees into `logic.Expr` trees.
The entry points and key functions are:

### Thing — The Universal Compiler Dispatcher

```go
// compiler/phase6.go
func (c *Compiler) Thing(node ast.Node) (lg.Expr, error)
```

This is the Go equivalent of Python's `thing(self)`:

```python
# ivy_compiler.py:86-92
def thing(self):
    with ASTContext(self):
        result = self.cmpl()
        return result

ivy_ast.AST.compile = thing  # monkey-patched onto all AST classes
```

In Python, `thing` calls `self.cmpl()` which dispatches via method
resolution — each AST class has its own `.cmpl` override. In Go,
`Thing` calls `CompileNode` which uses a type-switch.

### CompileNode — The Type Switch

```go
// compiler/compiler.go
func (c *Compiler) CompileNode(node ast.Node) (lg.Expr, error)
```

This is where AST types map to their compiled logic equivalents:

| AST type (input)       | Logic type (output)      | Python equivalent         |
|------------------------|--------------------------|---------------------------|
| `*ast.And`             | `*logic.And`             | `ivy_ast.And` → `ivy_logic.And` |
| `*ast.Or`              | `*logic.Or`              | `ivy_ast.Or` → `ivy_logic.Or`   |
| `*ast.Atom`            | `*logic.Apply`           | `ivy_ast.Atom` → `ivy_logic.Apply` |
| `*ast.Variable`        | `*logic.Variable`        | `ivy_ast.Variable` → `ivy_logic.Variable` |
| `*ast.Forall`          | `*logic.ForAll`          | `ivy_ast.Forall` → `ivy_logic.ForAll` |
| `*ast.Definition`      | `*logic.Definition`      | `ivy_ast.Definition` → `ivy_logic.Definition` |
| `*ast.LabeledFormula`  | `*ast.LabeledFormula` *  | `ivy_ast.LabeledFormula` → `ivy_ast.LabeledFormula` * |
| `*ast.AssertAction`    | `*actions.AssertAction`  | `ivy_actions.AssertAction` → `ivy_actions.AssertAction` |

The `*` on LabeledFormula is significant — see Section 4.

In Python, this mapping happens via `op_pairs` plus per-class
`.cmpl` overrides (ivy_compiler.py:120-148). The `op_pairs`
list directly maps ast classes to logic constructors:

```python
op_pairs = [
    (ivy_ast.And, ivy_logic.And),
    (ivy_ast.Or, ivy_logic.Or),
    (ivy_ast.Definition, ivy_logic.Definition),
    ...
]
```

### SortifyWithInference — Compile + Sort-Infer

```go
// compiler/compiler.go
func (c *Compiler) SortifyWithInference(node ast.Node) (lg.Expr, error)
```

Python equivalent: `sortify_with_inference(ast)` (ivy_compiler.py:555).

This is the most common entry point for compiling a formula. It:

1. Sets `TopSort` as the default sort context
2. Calls `Thing(node)` to compile `ast.Node` → `logic.Expr`
3. Calls `SortInfer(result)` to resolve any remaining `TopSort` placeholders
4. Returns the fully sorted `logic.Expr`

Most formula compilation flows through this function. When you see
`SortifyWithInference` in the code, think: "take this raw AST and
produce a fully typed logic expression."


## 3. The Action Layer

Actions (`assert`, `assume`, `assign`, `call`, `if`, `while`, etc.)
are imperative commands that modify state. They exist in both the
AST layer and the compiled actions layer:

### AST Actions vs Compiled Actions

| AST type (ast/)             | Compiled type (actions/)       |
|-----------------------------|--------------------------------|
| `*ast.AssertAction`         | `*actions.AssertAction`        |
| `*ast.AssumeAction`         | `*actions.AssumeAction`        |
| `*ast.AssignAction`         | `*actions.AssignAction`        |
| `*ast.CallAction`           | `*actions.CallAction`          |
| `*ast.IfAction`             | `*actions.IfAction`            |
| `*ast.Sequence`             | `*actions.Sequence`            |

AST actions store their children as `ast.Node`. Compiled actions
store their children as `logic.Expr`. For example:

```go
// ast/ — parsed, untyped
type AssertAction struct {
    Base
    Args []Node  // args[0] is the formula to assert (ast.Node)
}

// actions/ — compiled, typed
type AssertAction struct {
    ActionBase
    Formula    lg.Expr              // the compiled logic formula
    LF         *ast.LabeledFormula  // metadata container (see Section 4)
    Proof      lg.Expr
    Unprovable bool
}
```

The compiled `actions.AssertAction` stores its formula as `lg.Expr`
rather than `ast.Node`, guaranteeing that anyone who accesses
`.Formula` gets a fully sorted expression.

### CompileActionBody

```go
// compiler/action.go
func (c *Compiler) CompileActionBody(node ast.Node) (actions.Action, error)
```

This compiles AST action nodes into their `actions/` equivalents.
It dispatches on the concrete AST action type and calls specialized
compilers (e.g., `CompileAssertFormula`, `CompileAssign`).

Note: compiled actions implement `logic.Expr` (via their
`ActionBase`), so they can appear anywhere a `logic.Expr` is
expected. This matches Python where `ivy_actions.Action` inherits
from `ivy_ast.AST` and participates in the same `thing`/`compile`
dispatch.


## 4. The LabeledFormula Pattern

`LabeledFormula` is the most architecturally interesting type in
Ivy because it straddles the AST and logic universes.

### What Is It?

A LabeledFormula pairs a logical formula with metadata:

```go
// ast/decl_ast.go
type LabeledFormula struct {
    Base
    Label        Node   // source label like [my_property], may be nil
    Formula      Node   // the actual logical formula
    ID           int64  // unique identifier for proof tracking
    Temporal     *bool
    Explicit     bool
    IsDefinition bool
    Assumed      bool
    Unprovable   bool
    Annot        interface{}
}
```

Python equivalent: `ivy_ast.LabeledFormula` (ivy_ast.py:622).

### Structural vs Semantic Accessors

LabeledFormula exposes three ways to access its contents, serving
different purposes:

- **`Args()`** — structural children for generic tree walking.
  Returns `[]Node{lf.Label, lf.Formula}`. This is the `ast.Node`
  interface method. Any code that generically walks all AST nodes
  (e.g., `AstRewrite`, `collectStaleSymbols`) uses this to descend
  into both the label and the formula. It does not interpret the
  formula's internal structure.

- **`GoalConc(lf)`** (`proof/goal.go`) — semantic conclusion
  extraction. If `lf.Formula` is a `*ast.SchemaBody`, returns the
  last element (the conclusion) as `lg.Expr`. Otherwise returns
  `lf.Formula` cast to `lg.Expr`. This is the logical accessor:
  it knows that a SchemaBody's last element is the conclusion.

- **`GoalPrems(lf)`** (`proof/goal.go`) — semantic premise
  extraction. If `lf.Formula` is a `*ast.SchemaBody`, returns all
  elements except the last. Otherwise returns nil.

The distinction matters: `GoalConc` only returns the conclusion,
so code that needs to examine the **entire** formula tree (including
all premises) must use `Args()` to walk the full structure. The fix
to `collectStaleSymbols` in `proof/checker.go` was specifically
about this: it previously used `GoalConc()` and missed symbols in
premises; the fix switched to walking via `Args()`.

### Why It Does Not Implement logic.Expr

LabeledFormula is a **structural container**, not a logic expression.
It carries metadata that the logic layer does not need: proof labels,
temporal flags, the `unprovable` marker, etc. In Python, it inherits
from `ivy_ast.AST`, NOT from `ivy_logic.Formula`. It never appears
in `op_pairs`. It is never passed to `sort_infer`, `substitute`, or
other logic operations directly. All consuming code unwraps it:

```python
# ivy_actions.py — Python pattern
fmla = self.args[0]
if isinstance(fmla, ivy_ast.LabeledFormula):
    unprovable = fmla.unprovable
    fmla = fmla.formula      # <-- unwrap to get the logic expression
type_check(domain, fmla)
```

### Compilation Preserves Its Identity

When compiled, a LabeledFormula stays a LabeledFormula — it just
gets cloned with compiled children:

```python
# ivy_compiler.py:505-511 — Python
def _labeled_formula_cmpl(self):
    return self.clone([
        None if self.label is None
             else self.label.clone([sortify_with_inference(x) for x in self.label.args]),
        self.formula.compile()
             if isinstance(self.formula, SchemaBody)
             else sortify_with_inference(self.formula)
    ])
```

After compilation:
- `lf.Label` — an `ast.Node` with sorted args (or nil)
- `lf.Formula` — a `logic.Expr` (the compiled, sorted formula)

Note that `lf.Formula` is stored as `ast.Node` (the field type on
the struct), but at runtime after compilation it will satisfy
`logic.Expr` because `sortify_with_inference` produces `logic.Expr`
values, and `logic.Expr` embeds `ast.Node`.

### The Two Compilation Paths

Go provides two ways to compile a LabeledFormula:

**`CompileLF`** — returns `*ast.LabeledFormula` (the container):

```go
func (c *Compiler) CompileLF(lf *ast.LabeledFormula) (*ast.LabeledFormula, error)
```

This is the faithful port of `_labeled_formula_cmpl`. It iterates
`label.args`, calls `SortifyWithInference` on each, clones the
label, compiles the formula, then clones the LabeledFormula with
compiled children. The clone emits the `PRESERVE` xtrace and
copies all metadata.

**`ThingLF`** — wraps CompileLF with trace emissions:

```go
func (c *Compiler) ThingLF(lf *ast.LabeledFormula) (*ast.LabeledFormula, error)
```

This is the Go equivalent of calling `thing()` on a LabeledFormula
in Python. It emits the same `Thing ENTER`/`CompileNode` traces
that Python's `thing` function does, then delegates to `CompileLF`.

Use `ThingLF` when you know you have a LabeledFormula and need the
full container back (label, metadata, compiled formula). Use `Thing`
when you need a bare `logic.Expr` and don't care about the
container (CompileNode extracts the inner formula automatically).

### The LF Field on Actions

Python stores a compiled LabeledFormula directly as
`AssertAction.args[0]` thanks to duck typing. Go can't do that
because `AssertAction.Formula` is `logic.Expr` and LabeledFormula
doesn't implement `logic.Expr`. Instead, Go uses two fields:

```go
type AssertAction struct {
    Formula    lg.Expr              // the unwrapped logic formula (always set)
    LF         *ast.LabeledFormula  // the container with metadata (nil if absent)
    Unprovable bool
    ...
}
```

This is **stronger** than Python: you get compile-time assurance
that `.Formula` is always a sorted logic expression, while `.LF`
gives access to all the metadata when present. The Python
`isinstance(fmla, LabeledFormula)` check becomes `a.LF != nil`
in Go.


## 5. The Definition Type (Both Universes)

`Definition` exists in both `ast/` and `logic/`:

- **`*ast.Definition`** — parsed: `p(X) = expr` before sort resolution
- **`*logic.Definition`** — compiled: `p(X) = expr` with full sorts

This is one of the `op_pairs` mappings (ivy_compiler.py:127):
```python
(ivy_ast.Definition, ivy_logic.Definition)
```

`logic.Definition` implements `logic.Expr` with `NodeSort()` returning
`Boolean`. It has `Lhs` and `Rhs` fields (both `logic.Expr`), a
`Defines()` method that extracts the symbol being defined, and a
`ToConstraint()` method that converts it to a formula for verification.

Do not confuse `logic.Definition` with `ast.LabeledFormula`.
A Definition is a logical equation (`p(X) <-> fmla`). A
LabeledFormula is a metadata wrapper (`[label] formula`). A
LabeledFormula may contain a Definition as its `.Formula`, but
they serve different roles.


## 6. Sort Inference

Sort inference resolves `TopSort` placeholders to concrete sorts.

### Where Sorts Come From

The parser produces AST nodes with optional sort annotations:
`x:nat` becomes an `ast.Variable{Rep:"x", VSort:"nat"}`. Most
nodes start untyped — their sort is `TopSort` (the universal
placeholder meaning "I don't know yet").

### How Resolution Works

`SortifyWithInference` is the main entry point:

1. **Compile**: `Thing(node)` converts `ast.Node` → `logic.Expr`.
   During compilation, each node gets a provisional sort: symbols
   are looked up in `Sig` (the signature), variables get their
   declared sort or `TopSort`, function applications get `TopSort`
   initially.

2. **Infer**: `SortInfer(expr)` (corresponding to Python's
   `sort_infer`) runs unification-based type inference. It walks the
   expression, collects sort constraints, unifies them, and replaces
   all remaining `TopSort` with concrete sorts. If unification
   fails, it reports a type error.

After `SortifyWithInference`, every node in the returned `logic.Expr`
tree has a concrete sort.

### The Signature (Sig)

The signature (`*ivylogic.Sig`) is the type environment — it maps
names to sorts:

```
"node_count" → FunctionSort(index → nat)
"alive"      → FunctionSort(node → Boolean)
"index"      → UninterpretedSort("index")
```

The signature grows as declarations are compiled. When compiling
inside a scope (action body, quantifier), the signature is copied
to add local bindings without polluting the outer scope.


## 7. Python vs Go: Key Variances

### Dispatch Mechanism

Python uses monkey-patching to assign `.cmpl` methods:
```python
ivy_ast.And.cmpl = _make_op_cmpl("And", ivy_logic.And)
ivy_ast.LabeledFormula.cmpl = _labeled_formula_cmpl
ivy_actions.AssertAction.cmpl = compile_assert_action
```

Go uses a type-switch in `CompileNode`. The effect is the same.

### Return Types

Python's `thing()` returns whatever `.cmpl()` returns — could be
`ivy_logic.And`, `ivy_ast.LabeledFormula`, or `ivy_actions.AssertAction`.
No static type checking.

Go's `Thing()` returns `(logic.Expr, error)`. This means
LabeledFormula compilation through `Thing`/`CompileNode` must
extract the inner formula. When you need the full LabeledFormula
back, use `ThingLF` instead.

### Duck Typing vs Explicit Fields

Python stores heterogeneous values in `.args` lists. The same
`args[0]` slot on an AssertAction might hold a LabeledFormula or
a plain formula. Code uses `isinstance` checks at runtime.

Go uses explicit typed fields. AssertAction has both `Formula`
(`logic.Expr`) and `LF` (`*ast.LabeledFormula`). The type system
enforces correct usage at compile time.

### Global State

Python uses module-level globals (`ivy_logic.sig`, `lf_counter`).
Go uses per-session Config structs threaded through callers
(see CLAUDE.md Section C). This enables concurrent Ivy sessions
on multi-core machines.

### Clone Semantics

Both Python and Go use `clone(args)` to create modified copies of
nodes. In Python, `clone` returns `type(self)(*args)` — always the
same Python class. In Go, `Clone` returns `ast.Node` (the
interface), and the caller may need a type assertion to get the
concrete type back:

```go
result := lf.Clone([]ast.Node{label, formula}).(*ast.LabeledFormula)
```

This is safe when you control the input type; `Clone` on a
`*ast.LabeledFormula` always returns `*ast.LabeledFormula`.


## 8. Summary: Reading the Type at Each Stage

When reading goivy compiler code, the type tells you where you are:

| Type                     | Stage              | Sorted? | Has metadata? |
|--------------------------|--------------------|---------|---------------|
| `ast.Node`               | Parsed AST         | No      | No            |
| `*ast.Atom`              | Parsed term        | No      | No            |
| `*ast.LabeledFormula`    | Container          | Mixed*  | Yes           |
| `logic.Expr`             | Compiled logic     | Yes     | No            |
| `*logic.Apply`           | Compiled term      | Yes     | No            |
| `*logic.Definition`      | Compiled equation  | Yes     | No            |
| `*actions.AssertAction`  | Compiled action    | Yes     | Via .LF       |
| `*actions.Sequence`      | Compiled block     | Yes     | No            |

\* After compilation, a LabeledFormula's `.Formula` field holds a
`logic.Expr` (stored as `ast.Node`). Before compilation, it holds
an `ast.Node` that is not yet sorted.

The general flow:

```
Source text
  → LALR parser (parser/)
  → ast.Node tree (ast/)
  → Compiler: Thing / CompileNode / SortifyWithInference (compiler/)
  → logic.Expr tree (logic/) + actions.Action tree (actions/)
  → Transition relation / verification (transrel/, solver/)
```

## 9. The Uniform Child Iteration Protocol (Args/Clone)

Two methods on `ast.Node` form a bidirectional decomposition/recomposition
protocol that enables generic tree traversal and rewriting across all
100+ node types in `ast/`, `logic/`, and `actions/`:

```go
Args() []Node            // decompose: return ordered children
Clone(args []Node) Node  // recompose: rebuild with (possibly modified) children
```

**Invariant**: `x.Clone(x.Args())` produces a structurally identical copy
(modulo pointer identity). `Clone` preserves all non-child state: `Base`
fields (lineno, config), quantifier variables, function symbols, sort
fields, etc. Only the child slots change.


### 9.1 Leaf vs Interior Nodes

Leaf nodes return nil from `Args()` and typically return themselves
from `Clone()`:

| Type                | `Args()`   | `Clone()`     |
|---------------------|------------|---------------|
| `*lg.Variable`      | `nil`      | identity      |
| `*lg.Const`         | `nil`      | identity      |
| `*ast.NoneAST`      | `nil`      | identity      |
| `*lg.TopSort`       | `nil`      | identity      |
| `*lg.BooleanSort`   | `nil`      | identity      |

Interior nodes return their typed children cast to `[]ast.Node`:

| Type                    | `Args()` returns                        |
|-------------------------|-----------------------------------------|
| `*lg.And`               | `a.Terms` (cast from `[]Expr`)          |
| `*lg.ForAll`            | `[]Node{f.Body}`                        |
| `*lg.Implies`           | `[]Node{im.T1, im.T2}`                  |
| `*ast.LabeledFormula`   | `[]Node{lf.Label, lf.Formula}`          |
| `*ast.SchemaBody`       | `s.Elems` (premises + conclusion)       |
| `*ast.Atom`             | `a.Terms`                               |
| `*ast.Sequence`         | `s.Stmts`                               |


### 9.2 How Logic Types Bridge the Protocol

Logic types (`logic/` package) store their children as typed `lg.Expr`
fields. The `Args()` and `Clone()` implementations in
`logic/ast_compat.go` bridge between `[]ast.Node` and the typed fields:

```go
// logic/ast_compat.go — example: *lg.And
func (a *And) Args() []ast.Node {
    r := make([]ast.Node, len(a.Terms))
    for i, t := range a.Terms { r[i] = t }
    return r
}
func (a *And) Clone(args []ast.Node) ast.Node {
    terms := make([]Expr, len(args))
    for i, arg := range args { terms[i] = arg.(Expr) }
    return &And{Terms: terms}
}
```

`Args()` allocates a `[]ast.Node` slice and copies typed children into
it. `Clone()` casts each `ast.Node` argument back to `lg.Expr` via type
assertion and constructs a new instance.

Note: `Clone()` on a logic type will **panic** if passed an `ast.Node`
that does not satisfy `lg.Expr`. This is by design — it prevents
injecting unsorted AST nodes into a sorted logic tree.


### 9.3 Key Consumers

**`ast.AstRewrite()`** (`ast/rewrite.go`): The primary generic tree
rewriter. It has specialized cases for common types (`*Variable`,
`*Atom`, `*App`, `*LabeledFormula`, `*SchemaBody`, etc.), but the
default fallback demonstrates the protocol in its purest form:

```go
// ast/rewrite.go — default fallback (line ~771)
default:
    if rw, ok := x.(AstRewritable); ok {
        return rw.Rewrite(rewrite)
    }
    if args := x.Args(); args != nil {
        newArgs := AstRewriteSlice(args, rewrite)
        res := x.Clone(newArgs)
        res.SetLineno(safeLinenoAddRef(x, x.GetLineno()))
        return res
    }
    return x
```

This handles any node type not explicitly cased: decompose via `Args()`,
rewrite children recursively via `AstRewriteSlice`, recompose via
`Clone()`.

**`collectStaleSymbols()`** (`proof/checker.go`): Walks mixed AST/logic
trees to find all constant symbols. Demonstrates the three-tier dispatch
pattern (see 9.4 below).

**Goal manipulation functions** (`proof/goal.go`): `NormalizeGoal`,
`CloneGoal`, `GoalVocab` all use `Args()` and `Clone()` to restructure
goal trees.


### 9.4 The Three-Tier Dispatch Pattern

When walking a mixed AST/logic tree, code often uses three tiers:

1. **Check specific leaf types** (e.g., `*lg.Const`) — handle directly
2. **Check `lg.Expr`** — use an optimized logic-level walker
3. **Fall back to `Args()`** — generic walk for pure AST containers

Canonical example from `proof/checker.go`:

```go
func collectStaleSymbols(n ast.Node, stale map[string]bool) {
    if n == nil { return }
    // Tier 1: specific leaf type
    if c, ok := n.(*lg.Const); ok {
        stale[c.Name] = true
        return
    }
    // Tier 2: any logic expression — use optimized walker
    if expr, ok := n.(lg.Expr); ok {
        for _, c := range clauseops.UsedSymbolsAST(expr) {
            stale[c.Name] = true
        }
        return
    }
    // Tier 3: pure AST container — walk via Args()
    for _, child := range n.Args() {
        collectStaleSymbols(child, stale)
    }
}
```

Why this pattern exists: A `LabeledFormula.Formula` can be a
`*ast.SchemaBody` (pure AST container) holding `lg.Expr` children.
The SchemaBody itself is not an `lg.Expr`, so tier 2 does not match.
It falls through to tier 3, which descends via `Args()` into the
SchemaBody's elements — those elements ARE `lg.Expr` values and get
caught by tier 2 on the next recursive call.


### 9.5 Python Correspondence

Python's equivalent uses `self.args` (a tuple of children) and
`clone(*args)`:

| Python                          | Go                                         |
|---------------------------------|--------------------------------------------|
| `x.args`                        | `x.Args()`                                 |
| `x.clone(*new_args)`            | `x.Clone(newArgs)`                         |
| `hasattr(x, 'args')`            | Always true — `ast.Node` interface          |
| `[ast_rewrite(e, rw) for e in x.args]` | `AstRewriteSlice(x.Args(), rw)` |

The `AstRewrite` default case (line ~771) is a direct port of Python's
`ivy_ast.py` fallback:

```python
# Python: ivy_ast.py ast_rewrite fallback
if hasattr(x, 'args'):
    new = x.clone(*[ast_rewrite(e, rewrite) for e in x.args])
    return new
```

One difference: Python's `args` is typically a tuple (immutable); Go's
`Args()` returns a fresh `[]Node` slice each call.


# Q & A

Q: Does interface actions.Action implement interface logic.Expr ?

A: Yes. From module/action.go:22-23:

~~~
  type Action interface {
        lg.Expr // subsumes ast.Node
        // ...action-specific methods...
  }
~~~

Action embeds lg.Expr directly, so any actions.Action IS an lg.Expr. 
---------------
