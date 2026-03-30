# Plan: Add Vocab() Methods to Tactic Types + Investigate Proof Population

**Created**: 2026-03-31 00:30, **Updated**: 2026-03-31 01:45

## Context

The golden test (`make golden`) diverges at line 152556 in the `allSyms_pre_follow` phase:

```
152556  go : XTRACE: isolate.allSyms_pre_follow.sym cfabric.rd_fair
        py : XTRACE: isolate.allSyms_pre_follow.sym cf_pio_live.issued_pio
```

Go is missing `cf_pio_live.issued_pio` from `allSyms`. Traces match through `allSyms_post_action_refs` (n=141), so the divergence is between `post_action_refs` and `pre_follow`.

## Root Cause — TWO ISSUES

### Issue 1: Broken proof extraction (already partially fixed)

Go's proof symbol extraction (`isolate.go:973`) did `pe.Proof.(lg.Expr)` which always failed for Tactic AST nodes. This has been replaced with `ast.VocabNode` in the working tree, but...

### Issue 2: `mod.Proofs` is EMPTY (the real problem)

**Discovered during implementation testing**: `mod.Proofs` has length 0 in the golden test. The vocab infrastructure is correct, but there are no proofs to process. Python's `mod.proofs` is likely NOT empty for this test case, which is why Python adds `cf_pio_live.issued_pio` and Go doesn't.

**Action needed**: Investigate why Go's `mod.Proofs` is empty. The proofs are populated in `compiler/decl.go:1387` and `compiler/ivy_compile.go:500`. Either the compilation path doesn't reach these points, or proofs are cleared somewhere before `isolate.go` runs.

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
// Only yields from *Atom (where Rep is a string).
// Does NOT yield from *App — Python's symbols_ast yields App.rep (Symbol/This
// objects), but these never match the consumer's string membership test
// (x.formula.defines().name in all_names) because Python's Symbol.__eq__ and
// This.__eq__ reject string comparisons. Yielding strings in Go would be a
// behavioral difference.
//
// Both Atom and App (and all other nodes) recurse on Args() children.
func IterSymbolsASTNode(node Node) iter.Seq[string] {
    return func(yield func(string) bool) {
        iterSymbolsASTNodeRec(node, yield)
    }
}

func iterSymbolsASTNodeRec(node Node, yield func(string) bool) bool {
    if node == nil {
        return true
    }
    if atom, ok := node.(*Atom); ok && atom.Rep != "" {
        if !yield(atom.Rep) {
            return false
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

**Why only `*Atom`, not `*App`:**

Python's `symbols_ast` yields from both `isinstance(ast, (App, Atom))`. But:
- `Atom.rep` is a **string** → matches the consumer's string membership test ✓
- `App.rep` is a **Symbol** or **This** object → NEVER matches due to `__eq__` type checking
- `This` IS possible as `App.Rep` (parser creates `App(This())` for property names, ivy_parser.py:992), but `This.__eq__` uses default object identity — it never matches strings
- If Go yielded strings from App.rep, it would match extra entries the consumer ignores in Python, causing a behavioral divergence

No `repName()` helper or `relnamer` interface needed — the Atom-only check is simpler and strictly conformant.

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

Replace broken `lg.Expr` type assertion block. Note: `xtracer.Trace` calls should NOT be wrapped in `if xtracer.Enabled {}` blocks (the Trace function handles this internally).

```go
// Collect names from proofs
// Python: for x in mod.proofs: x[1].vocab(all_names)
allNames := ast.NewVocabNames()
xtracer.Trace("isolate.proofs n=%d", len(mod.Proofs))
for _, pe := range mod.Proofs {
    if pe.Proof != nil {
        xtracer.Trace("isolate.proof type=%T", pe.Proof)
        ast.VocabNode(pe.Proof, allNames)
    }
}
xtracer.Trace("isolate.allNames_from_proofs n=%d", allNames.Len())

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

// Build a plain map for downstream consumers (EraseUnrefed, filter_symbols)
allNamesMap := make(map[string]bool, allNames.Len())
for name, _ := range allNames.All() {
    allNamesMap[name] = true
}
```

Then use `allNamesMap` for the downstream callers:
- `actions.EraseUnrefed(act, allSyms, allNamesMap)` (line ~1033)
- `!allNamesMap[name]` (line ~1295)

### 5. Investigate why `mod.Proofs` is empty

**Status: NEEDED** — The vocab infrastructure is correct, but `mod.Proofs` is empty (n=0) in the golden test. Must investigate:

1. Where Python's `mod.proofs` is populated for this test case
2. Whether Go's compilation path reaches the proof-storing code in `compiler/decl.go:1387` and `compiler/ivy_compile.go:500`
3. Whether proofs are cleared somewhere before `isolate.go` runs (e.g., in `module.Clone()`)
4. Add XTRACE in Python to confirm Python's `mod.proofs` is non-empty for this test

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
| `ast/tactic.go` | Add `VocabNames` type, `NewVocabNames`, `VocabNamesUpdate`, `Vocaber`, `VocabNode`, `IterSymbolsASTNode` (Atom-only yield), and 9 `Vocab()` methods. Remove `repName()` helper. |
| `isolate/isolate.go:971-1004` | Replace broken `lg.Expr` proof loop with `VocabNode` + `VocabNames`; add debug traces (no `if xtracer.Enabled` wrapping); convert to `allNamesMap` for downstream |

No new files — uses existing `ivyutils.InsMap`.

## Implementation Status

- [x] Vocab infrastructure in `ast/tactic.go` (already in working tree — needs `repName` → `Relname()` refactor)
- [x] isolate.go proof loop replacement (already in working tree — needs `if xtracer.Enabled` cleanup)
- [ ] Investigate why `mod.Proofs` is empty — **this is the actual blocker**

## Verification

1. `cd ~/go/src/github.com/glycerine/goivy && go build ./...` — must compile
2. `cd ~/goivy && make test` — full test suite passes
3. `cd ~/goivy && make golden` — divergence at line 152556 should advance once proofs are populated
