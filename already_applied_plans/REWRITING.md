# Fix: `unknown type: index` — Module Instantiation Rewriting Incomplete

**Created: 2026-03-22**

## Context

After fixing include resolution and module registry merging, `goivy_check` on ord_live.ivy fails:
```
error: at line 29: unknown type: index
```

The `index` type is defined inside `order.ivy` at line 382-384:
```ivy
global {
    instance index : unbounded_sequence
}
```

When `instance index : unbounded_sequence` is expanded, it should rewrite all occurrences of `this` in the module body to `index`, producing declarations like `type index`, `alias index.t = index`, `action index.next(...)`, etc. But the rewriting is incomplete — specifically, `AliasDecl` (and other declaration types) are not being rewritten.

## Root Cause

### Python's approach: Unified recursive AST rewriter

Python uses `subst_prefix_atoms_ast(decl, subst, pref, ...)` which calls `ast_rewrite(ast, AstRewriteSubstPrefix(...))`. This is a **generic recursive rewriter** that processes ALL AST nodes uniformly:

```python
# ivy_ast.py line 1688
def ast_rewrite(x, rewrite):
    if isinstance(x, str):
        return rewrite.rewrite_name(x)
    if isinstance(x, Atom):
        return rewrite.rewrite_atom(x)
    if isinstance(x, TypeDef):
        res = x.clone(ast_rewrite(x.args, rewrite))
        ...
    if hasattr(x, 'args'):
        return x.clone(ast_rewrite(x.args, rewrite))  # <-- catches ALL node types
    ...
```

The key line `1644-1646`:
```python
def prefix_str(self, name, always):
    if name == 'this' and self.pref:
        return self.pref.rep  # "this" → "index"
```

This transforms `this` → prefix name in EVERY atom in the tree, regardless of what declaration type contains it.

### Go's approach: Incomplete type-switch

Go uses `prefixDeclNames(decl, prefix)` in `parser/decl.go:917` which has explicit `case` statements for each declaration type. It handles:
- ✅ TypeDecl, ConstantDecl, ActionDecl, MixinDecl, ObjectDecl, VariantDecl, DefinitionDecl, InterpretDecl, ConjectureDecl, PropertyDecl, AxiomDecl, InitDecl, InstantiateDecl, ProofDecl, TheoremDecl

It does NOT handle:
- ❌ **AliasDecl** — `alias t = this` stays unchanged, should become `alias index.t = index`
- ❌ RelationDecl, FunctionDecl, DestructorDecl, SchemaDecl, MacroDecl, DerivedDecl, ProgressDecl, DelegateDecl, NativeDecl, and other declaration types

Also, `substituteNamesInDecl()` (parser/parser.go:1482) has the same problem — it only handles a subset of node types.

### Consequence

When `instance index : unbounded_sequence` expands the module body:
1. `type this` → `type index` ✅ (TypeDecl case handles this)
2. `alias t = this` → stays as `alias t = this` ❌ (no AliasDecl case)
3. Later, the compiler registers alias `t → "this"` instead of `t → "index"`
4. Type lookup for `index` fails with "unknown type: index"

## Fix Strategy

The right fix matches the Python approach: implement a **unified recursive AST rewriter** that handles all node types generically, rather than adding more cases to the type-switch.

### Step 1: Implement `substPrefixAtomsAST` — a unified rewriter

**File: `parser/rewrite.go` (NEW)**

Port Python's `ast_rewrite` + `AstRewriteSubstPrefix` as a generic Go function:

```go
// SubstPrefixRewriter implements Python's AstRewriteSubstPrefix.
// It rewrites all atoms in an AST tree by:
// 1. Substituting names via subst map
// 2. Prefixing names that are in the toPref set (or all if toPref is nil)
// 3. Replacing "this" with the prefix name
type SubstPrefixRewriter struct {
    Subst   map[string]string // formal param → actual arg
    Pref    string            // prefix name (e.g., "index")
    ToPref  map[string]bool   // names that should be prefixed (nil = all)
    Static  map[string]bool   // names that are static (type/destructor defined)
}

// RewriteAST recursively rewrites an AST node.
// Matches Python's ast_rewrite (ivy_ast.py:1688-1746).
func (r *SubstPrefixRewriter) RewriteAST(node ast.Node) ast.Node {
    // Generic: clone node with rewritten children
    // For Atom nodes: rewrite name via rewrite_atom logic
    // For Symbol nodes: rewrite name via prefix_str logic
    // For all other nodes: recursively rewrite children
}
```

The key operations matching Python:
- `prefix_str(name)`: if `name == "this"` → return `pref`; else if `name in toPref` → return `compose(pref, name)`
- `rewrite_atom(atom)`: apply `prefix_str` to atom name, then optionally compose with prefix
- For each node: clone with rewritten children (using `ast.Node.Clone()`)

### Step 2: Replace `prefixDeclNames` and `substituteNamesInDecl`

**File: `parser/decl.go`**

Replace the incomplete `prefixDeclNames` type-switch with a call to the unified rewriter:

```go
func prefixDeclNames(decl ast.Node, prefix string, toPref map[string]bool, static map[string]bool) []ast.Node {
    rw := &SubstPrefixRewriter{
        Pref:   prefix,
        ToPref: toPref,
        Static: static,
    }
    rewritten := rw.RewriteAST(decl)
    return []ast.Node{rewritten}
}
```

**File: `parser/parser.go`**

Replace `substituteNamesInDecl` similarly.

### Step 3: Update `expandInstantiation` to use the rewriter

**File: `parser/parser.go`**

The current `expandInstantiation()` calls:
```go
expanded = substituteNamesInDecl(expanded, subst)
expanded = prefixDeclNames(expanded, prefix)
```

Replace with a single unified rewrite call:
```go
rw := &SubstPrefixRewriter{
    Subst:  subst,
    Pref:   prefix,
    ToPref: toPref,   // built from module.defined
    Static: static,   // built from module.static
}
expanded = rw.RewriteAST(decl)
```

### Step 4: Implement Clone() on all AST nodes

The rewriter needs to clone nodes with modified children. Verify that all AST node types implement `Clone(children []Node) Node`. If not, add the method.

Python's `ast_rewrite` uses `x.clone(ast_rewrite(x.args, rewrite))` — it clones the node with rewritten args. Go AST nodes have `Clone()` methods via the `Base` struct.

## Critical Implementation Details

### The `this` → prefix transformation

Python line 1644-1646:
```python
def prefix_str(self, name, always):
    if name == 'this' and self.pref:
        return self.pref.rep
```

In Go, when we encounter any `Atom` or `Symbol` with `Rep == "this"`, replace it with `prefix`.

### The `toPref` set (what names to prefix)

Python builds `to_pref` from `module.defined` — the set of names defined inside the module. Only names in this set get prefixed. Names NOT in this set (like built-in operators) stay unchanged.

For `unbounded_sequence`, the defined names include: `this`, `t`, `next`, `prev`, `succ`, `is_zero`, etc. The `toPref` set ensures only these get prefixed with `index.`.

### Static names

Python line 133-136:
```python
static = module.static.copy()
for name, dfs in module.defined.items():
    if any((df[1] is TypeDecl) or (df[1] is DestructorDecl) for df in dfs):
        static.add(name)
```

Static names (types and destructors) get a prefix without arguments. Non-static names (actions, functions) get prefix with arguments from the instantiation.

## Files to Modify

| File | Change |
|------|--------|
| `parser/rewrite.go` (NEW) | Implement `SubstPrefixRewriter` and `RewriteAST` |
| `parser/decl.go` | Replace `prefixDeclNames` type-switch with call to unified rewriter |
| `parser/parser.go` | Replace `substituteNamesInDecl` with call to unified rewriter; update `expandInstantiation` |

## Unit Tests

### Test 1: AliasDecl rewriting
```go
func TestRewriteAliasDecl(t *testing.T) {
    // alias t = this → alias index.t = index
    alias := ast.NewAliasDecl(ast.NewDefinition(
        ast.NewAtom("t"), ast.NewAtom("this")))
    rw := &SubstPrefixRewriter{Pref: "index", ToPref: map[string]bool{"t": true, "this": true}}
    result := rw.RewriteAST(alias)
    // Verify t → index.t and this → index
}
```

### Test 2: Full module expansion
```go
func TestExpandUnboundedSequence(t *testing.T) {
    // Parse: module mymod = { type this; alias t = this }
    // Then: instance x : mymod
    // Verify: type x produced, alias x.t = x produced
}
```

### Test 3: End-to-end include order
```go
func TestIncludeOrderIndexType(t *testing.T) {
    // Parse and compile: #lang ivy1.8\ninclude order\ntype foo
    // Verify: index type exists in signature
}
```

## Verification

1. `go build ./...` compiles
2. `go test ./...` passes
3. `./goivy_check isolate=cf_live ord_live.ivy` gets past "unknown type: index"
