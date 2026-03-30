# Plan: Add Vocab() Methods to Tactic Types for Proof Symbol Extraction

**Created**: 2026-03-31 00:30

## Context

The golden test (`make golden`) diverges at line 152556 in the `allSyms_pre_follow` phase:

```
152556  go : XTRACE: isolate.allSyms_pre_follow.sym cfabric.rd_fair
        py : XTRACE: isolate.allSyms_pre_follow.sym cf_pio_live.issued_pio
```

Go is missing `cf_pio_live.issued_pio` from `allSyms`. Traces match through `allSyms_post_action_refs` (n=141), so the divergence is in the proof symbol extraction loop.

## Root Cause

Go's proof symbol extraction (`isolate.go:973`) does `pe.Proof.(lg.Expr)` which ALWAYS fails for Tactic AST nodes → zero names extracted from proofs.

Python calls `x[1].vocab(all_names)` on each proof Tactic. The `vocab` methods call `names.update(symbols_ast(m.args[1]))` using the **AST-level** `symbols_ast` generator (ivy_ast.py:1879), which yields `ast.rep` from Atom/App nodes.

### What `names` is and what `names.update()` does

`names` is the `all_names = set()` from `ivy_isolate.py:1260`. It's a Python set.

`names.update(iterable)` consumes the `symbols_ast` generator and adds each yielded value to the set.

The call sites (ivy_ast.py):
- Line 782: `TacticWithMatch.vocab` — `names.update(symbols_ast(m.args[1]))` — match RHS expressions
- Line 854: `LetTactic.vocab` — `names.update(symbols_ast(m.args[1]))` — let binding RHS
- Line 863: `WitnessTactic.vocab` — `names.update(symbols_ast(m.args[1]))` — witness values
- Line 877: `IfTactic.vocab` — `names.update(symbols_ast(self.args[0]))` — condition formula
- Line 921: `TacticTactic.vocab` — `names.update(symbols_ast(self.args[1]))` — tactic body

`symbols_ast` (ivy_ast.py:1879) yields:
- `Atom.rep` — a **string** (the symbol name, e.g. `"cf_pio_live.issued_pio"`)
- `App.rep` — a **Symbol** object (`Symbol.rep` is string, `__hash__=hash(rep)`, `__eq__` checks type+rep)

The consumer (`ivy_isolate.py:1265`): `x.formula.defines().name in all_names` checks if a **string** is in the set. Due to Python's `Symbol.__eq__` type check, only string entries (from Atom.rep) ever match. Symbol entries from App.rep are stored but never matched.

### Key insight: We can yield `string` from the Go iterator

Since:
1. `Atom.Rep` is already a `string` in Go
2. `App.Rep` is a `Node`, but in practice always `*ast.Symbol` whose `Rep` field is a `string`
3. The consumer only does string-based membership testing

We can use `iter.Seq[string]` — extracting the name string from both Atom and App reps — and avoid `any`.

## Fix

### 1. Add `IterSymbolsASTNode` — `iter.Seq[string]`

**File**: `ast/tactic.go`

Port of `ivy_ast.symbols_ast` generator (line 1879). Follows the existing `clauseops.IterSymbolsAST` pattern (which returns `iter.Seq[*lg.Const]` for the logic-level equivalent).

```go
// IterSymbolsASTNode yields symbol name strings from an AST node tree.
// Port of Python ivy_ast.symbols_ast (ivy_ast.py:1879) as iter.Seq[string].
//
// For *Atom: yields Rep (string).
// For *App: extracts name string from Rep Node (usually *Symbol.Rep).
// Recurses on all Args() children.
func IterSymbolsASTNode(node Node) iter.Seq[string] {
    return func(yield func(string) bool) {
        iterSymbolsASTNodeRec(node, yield)
    }
}

func iterSymbolsASTNodeRec(node Node, yield func(string) bool) bool {
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
            if name := repName(n.Rep); name != "" {
                if !yield(name) {
                    return false
                }
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

// repName extracts the name string from an App's Rep node.
// Python App.rep is a Symbol with .rep string; Go App.Rep is a Node.
func repName(node Node) string {
    switch n := node.(type) {
    case *Symbol:
        return n.Rep
    case *Atom:
        return n.Rep
    default:
        return node.String()
    }
}
```

### 2. Add `Vocaber` interface and `VocabNode` dispatcher

**File**: `ast/tactic.go`

```go
// VocabNames is an insertion-ordered set of symbol name strings,
// used as the container for Vocab methods.
// Uses existing InsMap from ivyutils, preserving insertion order.
type VocabNames = iu.InsMap[string, bool]

func NewVocabNames() *VocabNames {
    return iu.NewInsMap[string, bool]()
}

// VocabNamesUpdate consumes an iter.Seq[string] and adds to the set.
// Mirrors Python: names.update(symbols_ast(...))
func VocabNamesUpdate(vn *VocabNames, seq iter.Seq[string]) {
    for name := range seq {
        vn.Set(name, true)
    }
}

// Vocaber is implemented by AST nodes that extract symbol names from proof trees.
// Port of Python Tactic.vocab(self, names) method hierarchy (ivy_ast.py).
type Vocaber interface {
    Vocab(names *VocabNames)
}

// VocabNode calls Vocab on the node if it implements Vocaber, otherwise no-op.
// Matches Python's base Tactic.vocab which is pass.
func VocabNode(node Node, names *VocabNames) {
    if v, ok := node.(Vocaber); ok {
        v.Vocab(names)
    }
}
```

### 3. Add `Vocab()` methods to Tactic types

**File**: `ast/tactic.go`

Each faithfully ports the corresponding Python `vocab()` (ivy_ast.py lines 780-950):

```go
// --- TacticWithMatch pattern: SchemaInstantiation, AssumeTactic ---
// Python: for m in self.match(): names.update(symbols_ast(m.args[1]))

func (s *SchemaInstantiation) Vocab(names *VocabNames) {
    for _, m := range s.Matches {
        args := m.Args()
        if len(args) >= 2 {
            VocabNamesUpdate(names, IterSymbolsASTNode(args[1]))
        }
    }
}

func (a *AssumeTactic) Vocab(names *VocabNames) {
    for _, m := range a.Matches {
        args := m.Args()
        if len(args) >= 2 {
            VocabNamesUpdate(names, IterSymbolsASTNode(args[1]))
        }
    }
}

// --- LetTactic ---
// Python: for m in self.args: names.update(symbols_ast(m.args[1]))

func (l *LetTactic) Vocab(names *VocabNames) {
    for _, d := range l.Defs {
        args := d.Args()
        if len(args) >= 2 {
            VocabNamesUpdate(names, IterSymbolsASTNode(args[1]))
        }
    }
}

// --- WitnessTactic ---
// Python: for m in self.args: names.update(symbols_ast(m.args[1]))

func (w *WitnessTactic) Vocab(names *VocabNames) {
    for _, m := range w.Witnesses {
        args := m.Args()
        if len(args) >= 2 {
            VocabNamesUpdate(names, IterSymbolsASTNode(args[1]))
        }
    }
}

// --- IfTactic ---
// Python: names.update(symbols_ast(self.args[0]))
//         for arg in self.args[1:]: arg.vocab(names)

func (i *IfTactic) Vocab(names *VocabNames) {
    VocabNamesUpdate(names, IterSymbolsASTNode(i.Cond))
    VocabNode(i.Then, names)
    VocabNode(i.Else, names)
}

// --- PropertyTactic ---
// Python: if not isinstance(self.args[2], NoneAST): self.args[2].vocab(names)

func (p *PropertyTactic) Vocab(names *VocabNames) {
    if p.Proof != nil {
        if _, isNone := p.Proof.(*NoneAST); !isNone {
            VocabNode(p.Proof, names)
        }
    }
}

// --- TacticTactic ---
// Python: names.update(symbols_ast(self.args[1]))
//         if not isinstance(self.args[2], NoneAST): self.args[2].vocab(names)

func (t *TacticTactic) Vocab(names *VocabNames) {
    VocabNamesUpdate(names, IterSymbolsASTNode(t.Body))
    if t.Proof != nil {
        if _, isNone := t.Proof.(*NoneAST); !isNone {
            VocabNode(t.Proof, names)
        }
    }
}

// --- ProofTactic ---
// Python: self.args[1].vocab(names)

func (p *ProofTactic) Vocab(names *VocabNames) {
    VocabNode(p.Proof, names)
}

// --- ComposeTactics ---
// Python: for arg in self.args: arg.vocab(names)

func (c *ComposeTactics) Vocab(names *VocabNames) {
    for _, t := range c.Tactics {
        VocabNode(t, names)
    }
}
```

Types with no-op vocab (no Python override — base `Tactic.vocab` is `pass`):
- Tactic, NullTactic, SpoilTactic, FunctionTactic, TacticWith, TacticLets
- AssumeGlobalTactic, UnfoldTactic, ForgetTactic, ShowGoalsTactic, DeferGoalTactic

### 4. Change isolate.go proof loop

**File**: `isolate/isolate.go` (lines 971-1004)

Replace broken `lg.Expr` type assertion block:

```go
// Collect names from proofs
// Python: for x in mod.proofs: x[1].vocab(all_names)
allNames := ast.NewVocabNames()
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
            if _, found := allNames.Get2(c.Name); found {
                allSyms[actions.ConstSymKey(c)] = c
            }
        }
    }
}
```

## Two Different `symbols_ast` Functions

Python has TWO separate `symbols_ast`:

1. **`ivy_ast.py:1879`** — AST-level, used by `vocab()` methods
   - Checks `isinstance(ast, (App, Atom))` (AST types)
   - Go equivalent: `IterSymbolsASTNode` → `iter.Seq[string]` ← **this plan**

2. **`ivy_logic_utils.py:537`** — logic-level, used in symbol collection
   - Checks `is_app(ast)` (logic IR types)
   - Go equivalent: `clauseops.IterSymbolsAST` → `iter.Seq[*lg.Const]` ← **already exists**

## Files Modified

| File | Changes |
|------|---------|
| `ast/tactic.go` | Add `VocabNames` type, `NewVocabNames`, `VocabNamesUpdate`, `Vocaber`, `VocabNode`, `IterSymbolsASTNode`, `repName`, and 9 `Vocab()` methods |
| `isolate/isolate.go:971-1004` | Replace broken `lg.Expr` proof loop with `VocabNode` + `VocabNames` |

No new files — uses existing `ivyutils.InsMap`.

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...` — must compile
2. `cd ~/goivy && make test` — full test suite passes
3. `cd ~/goivy && make golden` — divergence at line 152556 resolved; traces advance further
