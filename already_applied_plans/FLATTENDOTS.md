# Fix: `unknown symbol: evs` — Dotted Names Not Flattened During Parsing

**Created: 2026-03-22**

## Context

After fixing module instantiation rewriting, `goivy_check` fails at line 1665 of ord_live.ivy:
```
error: at line 1665: compiling arg 0 of *ast.Dot: compiling arg 1 of *ast.Dot: unknown symbol: evs
```
The expression `ref.evs(T).req` is being compiled. The compiler splits it into `Dot(Dot(ref, evs(T)), req)` and tries to resolve each part, but `evs` is not a standalone symbol.

## Root Cause: Go Creates Dot Nodes Where Python Flattens Names

**Python** (ivy_logic_parser.py line 151-154):
```python
def p_term_dot_appelem(p):
    'term : term DOT appelem'
    if isinstance(p[1], (Atom, App)):
        p[0] = compose_atoms(p[1], p[3])
```

Python **flattens** dotted names during parsing using `compose_atoms`. The expression `ref.evs(T).req` becomes:
1. `ref` → `Atom("ref", [])`
2. `ref.evs(T)` → `compose_atoms(Atom("ref"), App("evs",[T]))` → `App("ref.evs", [T])`
3. `ref.evs(T).req` → `compose_atoms(App("ref.evs",[T]), Atom("req"))` → `Atom("ref.evs.req", [T])`

Result: a single `Atom("ref.evs.req", [T])` — no Dot nodes at all.

**Go** creates nested `Dot` nodes: `Dot(Dot(Symbol("ref"), Atom("evs",[T])), Symbol("req"))`

The compiler's `compileFieldReferenceRec` expects either flat qualified names in `sig.Symbols` or sort-based destructor resolution. It cannot handle nested Dot nodes the way Python handles flat atoms.

## Fix

### Change Go's expression parser to use `compose_atoms` for dotted names

**File: `parser/expr.go`**

The Go parser handles the DOT operator as an infix expression (in `parseInfix`). Currently it creates a `Dot` node. Change it to call `compose_atoms` (matching Python's `compose_atoms(p[1], p[3])`), which flattens `Atom("ref",[]) DOT Atom("evs",[T])` into `Atom("ref.evs", [T])`.

Python's `compose_atoms` (ivy_ast.py):
```python
def compose_atoms(pr, name):
    """Compose a prefix atom and a name atom into one qualified atom.
    E.g., Atom("ref",[]) + Atom("evs",[T]) → Atom("ref.evs",[T])
    """
    name = name.rename(compose_names(pr.rep, name.rep))
    name.args = pr.args + name.args
    return name
```

The logic is:
1. New name = `compose_names(left.rep, right.rep)` (joins with ".")
2. New args = `left.args + right.args` (concatenates argument lists)
3. Returns a new Atom with the composed name and combined args

### Implementation

In `parser/expr.go`, find where DOT is handled as an infix operator. Currently:
```go
case lexer.DOT:
    right := p.parseExpr(precDot)
    return &ast.Dot{Left: left, Right: right}
```

Change to:
```go
case lexer.DOT:
    right := p.parseExpr(precDot)
    // Python: compose_atoms(left, right) — flatten dotted names
    // If both sides are Atom/Symbol, compose into single qualified name
    return composeAtomsExpr(left, right)
```

Where `composeAtomsExpr` matches Python's `compose_atoms`:
```go
func composeAtomsExpr(left, right ast.Node) ast.Node {
    // Get left name and args
    leftName, leftArgs := extractNameAndArgs(left)
    rightName, rightArgs := extractNameAndArgs(right)

    if leftName != "" && rightName != "" {
        // Compose: "ref" + "evs" → "ref.evs", args combined
        composedName := ivyutils.ComposeNames(leftName, rightName)
        allArgs := append(leftArgs, rightArgs...)
        return ast.NewAtom(composedName, allArgs...)
    }
    // Fallback: keep as Dot for non-Atom/non-Symbol cases (e.g., MethodCall)
    return &ast.Dot{Left: left, Right: right}
}
```

Python only composes when `isinstance(p[1], (Atom, App))` — if the left side is a Variable or other expression type, it creates a `MethodCall` instead.

### Also handle in the Python v1.6 path

Python has TWO dot handling paths (version-gated):
- v1.0-1.6 (line 127-133): `aterm : term DOT aterm` → `compose_atoms(p[1], p[3])`
- v1.7+ (line 151-163): `term : term DOT appelem` → `compose_atoms(p[1], p[3])` or `MethodCall`

Both do the same flattening. The Go parser should do it regardless of version.

### Handle `Old` expressions

Python line 156-160:
```python
elif isinstance(p[1], Old):
    t = compose_atoms(p[1].args[0], p[3])
    p[0] = Old(t)  # rewrap in Old
```

If the left side is `old X.Y`, Python composes `X.Y` inside the Old wrapper.

## Files to Modify

| File | Change |
|------|--------|
| `parser/expr.go` | Change DOT handling from creating `ast.Dot` to `composeAtomsExpr()` |

## Verification

1. `go build ./...` compiles
2. `go test ./...` passes
3. Parse `ref.evs(T).req` → should produce `Atom("ref.evs.req", [T])` not `Dot(Dot(...))`
4. `./goivy_check isolate=cf_live ord_live.ivy` gets past "unknown symbol: evs"
