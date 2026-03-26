# Plan: Extend Canon() comparison system beyond parser into compiler/verification

**Created:** 2026-03-26T23:00

## Context

The Canon() s-expression system currently covers AST nodes (parser output) and Logic IR types (via Sexp()). But the golden test only compares xtracer execution flow — it cannot verify that the **compiled state** matches between Go and Python. Once we move past parser conformance, we need to compare the state that flows from compilation → transition relations → Z3.

If Go and Python submit the same Z3 formulas, the port is correct. If they don't, we need to find where they diverge. Canon() at each pipeline stage gives us that microscope.

## Pipeline stages and Canon coverage

```
AST (parser output)           ← Canon() EXISTS on all types
  ↓ compilation
Module state (Logic IR)       ← Canon() EXISTS on logic types (via Sexp())
                                 but NOT on Clauses, Actions, or Module snapshot
  ↓ action compilation
Transition relations          ← Canon() MISSING (updates, modified symbols)
  ↓ forward image
History (symbolic traces)     ← Canon() MISSING
  ↓ formula translation
Z3 commands                   ← Natural choke point (sexpr() on Z3 objects)
```

## What needs Canon() — prioritized

### Tier 1: Clauses (highest value, smallest effort)

`Clauses` is the universal intermediate representation — axioms, conjectures, initial conditions, and action transitions all flow through it.

**Go**: `clauseops/clauses.go` — `Clauses` struct has `Fmlas []lg.Expr`, `Defs []*lg.Definition`, `Annot interface{}`
**Python**: `ivy_logic_utils.py` — `Clauses` class has `fmlas`, `defs`, `annot`

Both `lg.Expr` and `lg.Definition` already have `Sexp()` methods. Adding `Canon()` to Clauses is straightforward:

```go
// clauseops/clauses.go
func (c *Clauses) Canon() iu.Canonical {
    // fmlas
    var fmlaStrs []string
    for _, f := range c.Fmlas {
        fmlaStrs = append(fmlaStrs, f.Sexp())
    }
    // defs
    var defStrs []string
    for _, d := range c.Defs {
        defStrs = append(defStrs, d.Sexp())
    }
    return iu.Canonical(fmt.Sprintf("(clauses fmlas:[%s] defs:[%s])",
        strings.Join(fmlaStrs, " "), strings.Join(defStrs, " ")))
}
```

Python equivalent in `canon.py` or `canon_ast.py`:
```python
def _clauses_canon(self):
    fmlas = ' '.join(f.sexp() for f in self.fmlas)
    defs = ' '.join(d.sexp() for d in self.defs)
    return '(clauses fmlas:[%s] defs:[%s])' % (fmlas, defs)
Clauses.canon = _clauses_canon
```

### Tier 2: Module snapshot after each compilation pass

After each pass (IvyDomainSetup, IvyConjectureSetup, IvyARGSetup), emit a Canon snapshot of the key module fields. This catches divergences in what axioms/conjectures/actions were registered.

**Key fields to canonicalize:**
- `mod.LabeledAxioms` — already `[]*ast.LabeledFormula` with Canon()
- `mod.LabeledConjs` — same
- `mod.LabeledProps` — same
- `mod.LabeledInits` — same
- `mod.Definitions` — same
- `mod.Sig.Sorts` — map[string]Sort, each Sort has Sexp()
- `mod.Sig.Symbols` — map[string]*Symbol, each has Sexp()
- `mod.Sig.Interp` — map[string]interface{}, needs string conversion

Approach: Add `Module.CanonSnapshot(label string)` that emits a sorted canonical snapshot. Call it after each pass with xtracer so the golden test can compare.

### Tier 3: Action compilation output

Python: `action.update(domain, in_scope)` returns `(modified_symbols, transition_clauses, precondition)`
Go: equivalent in `actions/` package

Each action update contains `Clauses` (Tier 1 covers that). The tuple structure needs a simple Canon wrapper.

### Tier 4: Z3 choke point

Z3's `sexpr()` method returns S-expression strings for any Z3 AST. Both Go (via cgo z3bridge) and Python (via z3 module) can call it. This is the ultimate comparison — if both sides produce the same Z3 sexpr, the port is correct.

**Approach**: Before each `solver.check()`, emit `xtracer.Trace("z3.assert sexpr=%s", assertion.sexpr())`. The golden test then compares Z3 assertions directly.

This requires:
- Go: `z3bridge` already has access to Z3 C API — add `Z3_ast_to_string()` calls
- Python: `z3.ExprRef.sexpr()` already works

## Implementation order

**Phase A** (do first — enables module state comparison):
1. Add `Clauses.Canon()` in Go `clauseops/clauses.go`
2. Add `Clauses.canon()` in Python `canon.py` or `canon_ast.py`
3. Add `Module.CanonSnapshot()` in Go `module/module.go`
4. Add `Module.canon_snapshot()` in Python `ivy_module.py`
5. Emit snapshots after each compilation pass via xtracer

**Phase B** (Z3 choke point — the high-value target):
1. In Go `z3bridge/translate.go`, after building each Z3 assertion, emit its sexpr via xtracer
2. In Python `ivy_solver.py` or `z3_utils.py`, same
3. The golden test then catches Z3-level divergences directly

**Phase C** (action/transition comparison — if needed after A+B):
1. Add Canon to action update tuples
2. Add Canon to History objects

## Files to modify

| File | Changes |
|---|---|
| `clauseops/clauses.go` | Add `Canon()` method |
| `module/module.go` | Add `CanonSnapshot()` method |
| `compiler/ivy_compile.go` | Emit Canon snapshots after each pass |
| `z3bridge/translate.go` | Emit Z3 sexpr via xtracer before solver calls |
| `~/pyivy/ivy/ivy/canon.py` or `canon_ast.py` | Add `Clauses.canon()` |
| `~/pyivy/ivy/ivy/ivy_module.py` | Add `Module.canon_snapshot()` |
| `~/pyivy/ivy/ivy/ivy_solver.py` | Emit Z3 sexpr via xtracer |

## Verification

After each phase:
```bash
go build ./...
XTRACE_OFF=1 go test ./... -count=1 -timeout 120s
cd ~/goivy && make golden
```

The golden test will initially show many new divergences from the Canon snapshots — each one points to a specific compilation mismatch that needs fixing. This is the intended microscope effect.

## Design decisions (resolved)

- Canon snapshots go directly into xtracer output — no separate flag. The golden test is our microscope.
- Z3 sexpr comparison MUST normalize variable names to be numbering-invariant (Z3 may use different internal numbering between Go cgo and Python bindings). Apply a normalization pass that replaces `!k!N` style internal names with deterministic equivalents based on declaration order.
- All phases (A, B, C) will be implemented — not just the Z3 choke point. Each tier provides diagnostic value when the tier below diverges.
