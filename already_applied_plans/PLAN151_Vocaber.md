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

Go is missing the symbol `cf_pio_live.issued_pio` from `allSyms`. The trace up to `allSyms_post_action_refs` (n=141) matches perfectly, so the divergence is introduced between `post_action_refs` and `pre_follow` — in the proof symbol extraction loop.

## Root Cause

**Go fails to extract symbol names from proof Tactic AST nodes because it tries to cast them to `lg.Expr`, which always fails for Tactic types.**

### Python (correct)

```python
# ivy_isolate.py:1260-1262
all_names = set()
for x in mod.proofs:
    x[1].vocab(all_names)    # x[1] is a Tactic AST node with vocab() method
```

Python Tactic classes have `vocab(self, names)` methods that recursively extract symbol name strings from the proof tree. The base `Tactic.vocab` is a no-op; subclasses like `TacticWithMatch`, `ProofTactic`, `ComposeTactics`, etc. override it to extract names from match expressions and recurse into sub-proofs.

Inside vocab methods, `names.update(symbols_ast(m.args[1]))` extracts symbol names from AST nodes (Atom.rep strings).

### Go (BUG)

```go
// isolate/isolate.go:973-984
for _, pe := range mod.Proofs {
    if pe.Proof != nil {
        if n, ok := pe.Proof.(lg.Expr); ok {   // ALWAYS FAILS — Tactic ≠ lg.Expr
            proofSyms := make(map[lg.NodeKey]lg.Expr)
            collectUsedSymbolNames(n, proofSyms)   // never reached
            ...
        }
    }
}
```

`ProofEntry.Proof` is typed `ast.Node`. Tactic types implement `ast.Node` but NOT `lg.Expr`. The type assertion `pe.Proof.(lg.Expr)` always fails, so **zero names are extracted from proofs**. This means definition-defined symbols whose names appear only in proofs (like `cf_pio_live.issued_pio`) are never added to `allSyms`.

## Fix

### 1. Add `SymbolNamesASTNode` helper function

**File**: `ast/tactic.go` (add near top, before tactic types)

Port of Python's `symbols_ast` (ivy_logic_utils.py:537) for AST-level nodes. Extracts symbol name strings from AST node trees.

```go
// SymbolNamesASTNode extracts symbol names from an AST node tree into the names map.
// Port of Python symbols_ast (ivy_logic_utils.py:537) at the AST level.
// For Atom nodes, adds Rep (the symbol name string).
// For App nodes, recurses into Rep (which is a Node).
// Recurses on all Args() children.
func SymbolNamesASTNode(node Node, names map[string]bool) {
    if node == nil {
        return
    }
    switch n := node.(type) {
    case *Atom:
        if n.Rep != "" {
            names[n.Rep] = true
        }
    case *App:
        SymbolNamesASTNode(n.Rep, names)
    }
    for _, child := range node.Args() {
        SymbolNamesASTNode(child, names)
    }
}
```

### 2. Add `VocabNode` dispatcher function

**File**: `ast/tactic.go`

```go
// Vocaber is implemented by AST nodes that can extract symbol names from proof trees.
// Port of Python's Tactic.vocab(self, names) method hierarchy.
type Vocaber interface {
    Vocab(names map[string]bool)
}

// VocabNode calls Vocab on the node if it implements Vocaber, otherwise does nothing.
// This matches Python's base Tactic.vocab which is a no-op (pass).
func VocabNode(node Node, names map[string]bool) {
    if v, ok := node.(Vocaber); ok {
        v.Vocab(names)
    }
}
```

### 3. Add `Vocab()` methods to Tactic types

**File**: `ast/tactic.go`

Each method faithfully ports the corresponding Python `vocab()` override:

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

Go implementations:

```go
func (s *SchemaInstantiation) Vocab(names map[string]bool) {
    for _, m := range s.Matches {
        args := m.Args()
        if len(args) >= 2 {
            SymbolNamesASTNode(args[1], names)
        }
    }
}

func (a *AssumeTactic) Vocab(names map[string]bool) {
    for _, m := range a.Matches {
        args := m.Args()
        if len(args) >= 2 {
            SymbolNamesASTNode(args[1], names)
        }
    }
}

func (l *LetTactic) Vocab(names map[string]bool) {
    for _, d := range l.Defs {
        args := d.Args()
        if len(args) >= 2 {
            SymbolNamesASTNode(args[1], names)
        }
    }
}

func (w *WitnessTactic) Vocab(names map[string]bool) {
    for _, m := range w.Witnesses {
        args := m.Args()
        if len(args) >= 2 {
            SymbolNamesASTNode(args[1], names)
        }
    }
}

func (i *IfTactic) Vocab(names map[string]bool) {
    SymbolNamesASTNode(i.Cond, names)
    VocabNode(i.Then, names)
    VocabNode(i.Else, names)
}

func (p *PropertyTactic) Vocab(names map[string]bool) {
    if p.Proof != nil {
        if _, isNone := p.Proof.(*NoneAST); !isNone {
            VocabNode(p.Proof, names)
        }
    }
}

func (t *TacticTactic) Vocab(names map[string]bool) {
    SymbolNamesASTNode(t.Body, names)
    if t.Proof != nil {
        if _, isNone := t.Proof.(*NoneAST); !isNone {
            VocabNode(t.Proof, names)
        }
    }
}

func (p *ProofTactic) Vocab(names map[string]bool) {
    VocabNode(p.Proof, names)
}

func (c *ComposeTactics) Vocab(names map[string]bool) {
    for _, t := range c.Tactics {
        VocabNode(t, names)
    }
}
```

Types that do NOT need Vocab (Python base class has no-op, these don't override):
- Tactic (base), NullTactic, SpoilTactic, FunctionTactic, TacticWith, TacticLets
- AssumeGlobalTactic, UnfoldTactic, ForgetTactic, ShowGoalsTactic, DeferGoalTactic

### 4. Change isolate.go proof loop to use Vocab

**File**: `isolate/isolate.go` (lines 971-985)

Replace the broken `lg.Expr` type assertion with a call to `ast.VocabNode`:

```go
// Collect names from proofs (name-only set, used as fallback in erase_unrefed)
// Python: for x in mod.proofs: x[1].vocab(all_names)
allNames := make(map[string]bool)
for _, pe := range mod.Proofs {
    if pe.Proof != nil {
        ast.VocabNode(pe.Proof, allNames)
    }
}
```

This replaces the 14-line broken block (lines 972-985) with 5 lines that faithfully match Python.

## Files Modified

| File | Changes |
|------|---------|
| `ast/tactic.go` | Add `Vocaber` interface, `VocabNode`, `SymbolNamesASTNode`, and 9 `Vocab()` methods |
| `isolate/isolate.go:971-985` | Replace broken `lg.Expr` proof loop with `ast.VocabNode` call |

## Safety Analysis

- The `allNames` map is ONLY used in the immediately-following loop to check `if allNames[c.Name]` for definition-defined symbols. Adding more names (by correctly extracting them from proofs) can only ADD symbols to `allSyms`, never remove them.
- This is a strict superset fix: Go was extracting zero names before, now it extracts the correct names matching Python.
- No other code paths are affected.

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...` — must compile
2. `cd ~/goivy && make test` — full test suite passes
3. `cd ~/goivy && make golden` — divergence at line 152556 resolved; traces advance further
