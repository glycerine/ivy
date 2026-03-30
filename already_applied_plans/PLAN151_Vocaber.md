# Plan: Add Vocab() Methods to Tactic Types for Proof Symbol Extraction

**Created**: 2026-03-31 00:30

## Context

The golden test (`make golden`) diverges at line 152556 in the `allSyms_pre_follow` phase:

```
152555  go : XTRACE: isolate.allSyms_pre_follow.sym cf_pio_l
        py : XTRACE: isolate.allSyms_pre_follow.sym cf_pio_l

152556  go : XTRACE: isolate.allSyms_pre_follow.sym cfabric.rd_fair
        py : XTRACE: isolate.allSyms_pre_follow.sym cf_pio_live.issued_pio
```

Go is missing `cf_pio_live.issued_pio` from `allSyms`. Traces match perfectly through `allSyms_post_action_refs` (n=141), so the divergence is in the proof symbol extraction loop.

## Root Cause

**Go fails to extract symbol names from proof Tactic AST nodes because it tries to cast them to `lg.Expr`, which always fails for Tactic types.**

### Python (correct)

```python
# ivy_isolate.py:1260-1262
all_names = set()
for x in mod.proofs:
    x[1].vocab(all_names)    # x[1] is a Tactic AST node with vocab() method
```

Tactic classes have `vocab(self, names)` methods that call `names.update(symbols_ast(m.args[1]))`. This uses the **AST-level** `symbols_ast` generator defined in `ivy_ast.py:1879` (NOT the logic-level one in `ivy_logic_utils.py:537`):

```python
# ivy_ast.py:1879 — AST-level symbols_ast (generator, yields one at a time)
def symbols_ast(ast):
    if isinstance(ast, (App, Atom)):
        yield ast.rep          # Atom.rep = string; App.rep = Symbol object
    if ast != None and not isinstance(ast, str):
        for arg in ast.args:
            for x in symbols_ast(arg):
                yield x
```

Yields:
- For `ivy_ast.Atom`: the `.rep` **string**
- For `ivy_ast.App`: the `.rep` **Symbol** object (`__hash__=hash(rep)`, `__eq__` checks type+rep)

Consumer: `x.formula.defines().name in all_names` — string membership check. With Python's set semantics, only string entries match (Symbol.__eq__ rejects string comparisons due to type check).

### Go (BUG)

```go
// isolate/isolate.go:973-984
for _, pe := range mod.Proofs {
    if pe.Proof != nil {
        if n, ok := pe.Proof.(lg.Expr); ok {   // ALWAYS FAILS — Tactic ≠ lg.Expr
            ...                                  // never reached
        }
    }
}
```

`ProofEntry.Proof` is `ast.Node`. Tactics implement `ast.Node` but NOT `lg.Expr`. The type assertion always fails → **zero names extracted**.

## Design Decision: iter.Seq + InsMap

**iter.Seq vs just InsMap?** `iter.Seq` is more conformant:

1. Python's `symbols_ast` IS a generator (uses `yield`). Go `iter.Seq` IS an iterator. Direct mapping.
2. The codebase already has `clauseops.IterSymbolsAST` returning `iter.Seq[*lg.Const]` for the logic-level equivalent. The AST-level version should follow the same pattern.
3. Python creates `used_symbols_ast = gen_to_set(symbols_ast)` as a reusable variant, showing the generator is meant to be composable/reusable.
4. Separation of concerns: producer (iterator) is decoupled from consumer (container).

Container: use the existing `ivyutils.InsMap[K, V]` for insertion-ordered storage with structural equality.

## Fix

### 1. Add AST-level `IterSymbolsASTNode` — `iter.Seq[any]`

**File**: `ast/tactic.go`

Port of `ivy_ast.symbols_ast` (line 1879) as a Go iterator. Yields `any` because AST-level reps are heterogeneous (string from `Atom.Rep`, `Node` from `App.Rep`).

```go
// IterSymbolsASTNode yields symbol reps from an AST node tree.
// Port of Python ivy_ast.symbols_ast (ivy_ast.py:1879) as an iter.Seq.
//
// Yields:
//   - For *Atom: the Rep string
//   - For *App: the Rep Node (preserving type identity like Python's Symbol)
// Then recurses on all Args() children.
func IterSymbolsASTNode(node Node) iter.Seq[any] {
    return func(yield func(any) bool) {
        iterSymbolsASTNodeRec(node, yield)
    }
}

func iterSymbolsASTNodeRec(node Node, yield func(any) bool) bool {
    if node == nil {
        return true
    }
    switch n := node.(type) {
    case *Atom:
        if n.Rep != "" {
            if !yield(n.Rep) {
                return false
            }
        }
    case *App:
        if n.Rep != nil {
            if !yield(n.Rep) {
                return false
            }
        }
    }
    for _, child := range node.Args() {
        if !iterSymbolsASTNodeRec(child, yield) {
            return false
        }
    }
    return true
}
```

### 2. Add `Vocaber` interface and `VocabNode` dispatcher

**File**: `ast/tactic.go`

```go
// VocabSet is the container type for Vocab methods.
// Uses InsMap[string, any] to preserve insertion order with structural equality.
// Key: canonical name string. Value: the original yielded object.
type VocabSet = iu.InsMap[string, any]

// NewVocabSet creates an empty VocabSet.
func NewVocabSet() *VocabSet {
    return iu.NewInsMap[string, any]()
}

// VocabSetUpdate adds all values from an iter.Seq[any] to the VocabSet.
// Mirrors Python: names.update(symbols_ast(...))
// Extracts a string key from each value for structural equality:
//   - string: the string itself
//   - Node with Canon(): the canonical form
func VocabSetUpdate(vs *VocabSet, seq iter.Seq[any]) {
    for v := range seq {
        switch s := v.(type) {
        case string:
            vs.Set(s, v)
        case Node:
            vs.Set(string(s.Canon()), v)
        default:
            vs.Set(fmt.Sprint(v), v)
        }
    }
}

// Vocaber is implemented by AST nodes that extract symbol reps from proof trees.
// Port of Python's Tactic.vocab(self, names) method hierarchy.
type Vocaber interface {
    Vocab(names *VocabSet)
}

// VocabNode calls Vocab on the node if it implements Vocaber, otherwise no-op.
// Matches Python's base Tactic.vocab which is pass.
func VocabNode(node Node, names *VocabSet) {
    if v, ok := node.(Vocaber); ok {
        v.Vocab(names)
    }
}
```

### 3. Add `Vocab()` methods to Tactic types

**File**: `ast/tactic.go`

Each method faithfully ports the corresponding Python `vocab()` override (ivy_ast.py lines 760-950).

| Go Type | Python Type | Python vocab() behavior |
|---------|------------|------------------------|
| `SchemaInstantiation` | `TacticWithMatch` | `for m in self.match(): names.update(symbols_ast(m.args[1]))` |
| `AssumeTactic` | `AssumeTactic(TacticWithMatch)` | Same as TacticWithMatch |
| `LetTactic` | `LetTactic` | `for m in self.args: names.update(symbols_ast(m.args[1]))` |
| `WitnessTactic` | `WitnessTactic` | `for m in self.args: names.update(symbols_ast(m.args[1]))` |
| `IfTactic` | `IfTactic` | `symbols_ast(cond) + recurse on then/else` |
| `PropertyTactic` | `PropertyTactic` | `recurse on proof if not NoneAST` |
| `TacticTactic` | `TacticTactic` | `symbols_ast(body) + recurse on proof if not NoneAST` |
| `ProofTactic` | `ProofTactic` | `recurse on proof (args[1])` |
| `ComposeTactics` | `ComposeTactics` | `recurse on all tactics` |

```go
func (s *SchemaInstantiation) Vocab(names *VocabSet) {
    for _, m := range s.Matches {
        args := m.Args()
        if len(args) >= 2 {
            VocabSetUpdate(names, IterSymbolsASTNode(args[1]))
        }
    }
}

func (a *AssumeTactic) Vocab(names *VocabSet) {
    for _, m := range a.Matches {
        args := m.Args()
        if len(args) >= 2 {
            VocabSetUpdate(names, IterSymbolsASTNode(args[1]))
        }
    }
}

func (l *LetTactic) Vocab(names *VocabSet) {
    for _, d := range l.Defs {
        args := d.Args()
        if len(args) >= 2 {
            VocabSetUpdate(names, IterSymbolsASTNode(args[1]))
        }
    }
}

func (w *WitnessTactic) Vocab(names *VocabSet) {
    for _, m := range w.Witnesses {
        args := m.Args()
        if len(args) >= 2 {
            VocabSetUpdate(names, IterSymbolsASTNode(args[1]))
        }
    }
}

func (i *IfTactic) Vocab(names *VocabSet) {
    VocabSetUpdate(names, IterSymbolsASTNode(i.Cond))
    VocabNode(i.Then, names)
    VocabNode(i.Else, names)
}

func (p *PropertyTactic) Vocab(names *VocabSet) {
    if p.Proof != nil {
        if _, isNone := p.Proof.(*NoneAST); !isNone {
            VocabNode(p.Proof, names)
        }
    }
}

func (t *TacticTactic) Vocab(names *VocabSet) {
    VocabSetUpdate(names, IterSymbolsASTNode(t.Body))
    if t.Proof != nil {
        if _, isNone := t.Proof.(*NoneAST); !isNone {
            VocabNode(t.Proof, names)
        }
    }
}

func (p *ProofTactic) Vocab(names *VocabSet) {
    VocabNode(p.Proof, names)
}

func (c *ComposeTactics) Vocab(names *VocabSet) {
    for _, t := range c.Tactics {
        VocabNode(t, names)
    }
}
```

Types with no-op Vocab (Python base `Tactic.vocab` is `pass`, these don't override):
- Tactic, NullTactic, SpoilTactic, FunctionTactic, TacticWith, TacticLets
- AssumeGlobalTactic, UnfoldTactic, ForgetTactic, ShowGoalsTactic, DeferGoalTactic

### 4. Change isolate.go proof loop to use Vocab

**File**: `isolate/isolate.go` (lines 971-1004)

Replace the broken `lg.Expr` type assertion block:

```go
// Collect names from proofs (name-only set, used to include definition-defined symbols)
// Python: for x in mod.proofs: x[1].vocab(all_names)
allNames := ast.NewVocabSet()
for _, pe := range mod.Proofs {
    if pe.Proof != nil {
        ast.VocabNode(pe.Proof, allNames)
    }
}

// Add definition-defined symbols that are in allNames
// Python: if x.formula.defines().name in all_names: all_syms.add(x.formula.defines())
for _, dfn := range mod.Definitions {
    if dfn.Formula == nil {
        continue
    }
    fmla, ok := dfn.Formula.(lg.Expr)
    if !ok {
        continue
    }
    children := fmla.Children()
    if len(children) >= 1 {
        if c := definedSymbolConst(children[0]); c != nil {
            // Python: x.formula.defines().name in all_names
            // Only string entries in the set match (Python Symbol.__eq__ rejects
            // string comparisons due to type check). String entries come from
            // Atom.Rep, keyed directly by the name string.
            if _, found := allNames.Get2(c.Name); found {
                allSyms[actions.ConstSymKey(c)] = c
            }
        }
    }
}
```

## Two Different `symbols_ast` Functions

Important context: Python has TWO different `symbols_ast`:

1. **`ivy_ast.py:1879`** — AST-level, used by `vocab()` methods
   - Checks `isinstance(ast, (App, Atom))` (AST types)
   - Yields `ast.rep` (string for Atom, Symbol for App)
   - Go equivalent: `IterSymbolsASTNode` → `iter.Seq[any]` ← **this plan**

2. **`ivy_logic_utils.py:537`** — logic-level, used in symbol collection
   - Checks `is_app(ast)` (logic IR types)
   - Yields `ast.rep` (always `*lg.Const`)
   - Go equivalent: `clauseops.IterSymbolsAST` → `iter.Seq[*lg.Const]` ← **already exists**

## Files Modified

| File | Changes |
|------|---------|
| `ast/tactic.go` | Add `VocabSet` type alias, `NewVocabSet`, `VocabSetUpdate`, `Vocaber` interface, `VocabNode`, `IterSymbolsASTNode`, and 9 `Vocab()` methods |
| `isolate/isolate.go:971-1004` | Replace broken `lg.Expr` proof loop with `VocabNode` + `VocabSet` |

No new files — uses existing `ivyutils.InsMap`.

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...` — must compile
2. `cd ~/goivy && make test` — full test suite passes
3. `cd ~/goivy && make golden` — divergence at line 152556 resolved; traces advance further
