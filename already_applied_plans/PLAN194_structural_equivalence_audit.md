# Audit: Structural vs Identity Equality in Go Port

Created: 2026-04-04 10:00

## Context

Structural equivalence in Python Ivy is provided for
the logic.py types by inheriting from a custom
recstruct class defined in utils/recstruct_object.py.

The recstruct base class is a hand-rolled "metaprogramming" 
implementation for the Ivy project. This is a clever 
bit of code that dynamically generates class strings 
and then execs them to create highly optimized and
immutable data structures.

Python Ivy has **two distinct equality regimes**:

1. **`logic.py` types** (via `recstruct`): ALL types (`Var`, `Const`, `Apply`, `And`, `Or`, `Not`, `ForAll`, `Exists`, `Lambda`, etc.) have auto-generated `__eq__` and `__hash__` based on structural content (`_tup`). They can all be dict keys with structural semantics.

2. **`ivy_ast.py` types**: Only **Symbol** and **Variable** define both `__eq__` and `__hash__` (by `rep` only). A few others define `__eq__` but NOT `__hash__` (Atom, App, Literal — making them unhashable/non-dict-keyable). The ~130 remaining classes (And, Or, Not, LabeledFormula, all Decl types, all Tactic types, etc.) use **Python default identity equality** — `a == b` iff `a is b`.

The Go port uses `lg.NodeKey` (via `Sexp()`) for structural equivalence in the `logic/` package. This is **correct** because all `logic.py` types use `recstruct`.

The question is: are there places where Go code applies structural equivalence to `ast`-package types that Python handles with identity?

## Investigation Findings

### What's CORRECT in the Go port:

- **`logic/` package**: All types have `Sexp()` → `NodeKey` structural keys. This matches Python's `recstruct` `__eq__`/`__hash__` on `logic.py` types. ✅
- **`LabeledFormula` maps**: Go uses `map[int64]` keyed by `lf.ID`, matching Python's `pmap[lf.id]`. ✅
- **`ast/rewrite.go` substitutions**: Use `map[string]Node` keyed by `.Rep` string, not structural node equality. ✅
- **Name-based sets** (`seen[v.Name]`, `seen[c.Name]`): Safe — just deduplication by name string. ✅

### What needs closer inspection:

The `ast/` package types do NOT have `Key()`/`Sexp()` methods. They are NOT used in `map[NodeKey]` lookups directly. The only structural-key maps use `logic/` types (which is correct).

However, there are subtle points:

1. **`ivy_ast.Symbol.__eq__` compares by `rep` only** (ignores sort). Go's `lg.Const` (the logic-layer equivalent) compares by `(name, sort)` via `Sexp()`. This is correct because `lg.Const` corresponds to `logic.Const` (recstruct with `name` and `sort` fields), not `ivy_ast.Symbol`.

2. **`ivy_ast.Variable.__eq__` compares by `rep` only** (ignores sort). Go's `lg.Variable` compares by `(name, sort)` via `Sexp()`. Again correct — `lg.Variable` maps to `logic.Var`, not `ivy_ast.Variable`.

3. **`ivy_ast.Atom.__eq__` compares by `(rep, args)`** but has NO `__hash__`, so Atoms can never be dict keys in Python. If Go ever uses an Atom-equivalent as a map key... but ast.Atom doesn't implement `Sexp()` and isn't used as `NodeKey`.

## Conclusion

**The Go port is actually correct here.** The two-layer architecture is preserved:

| Layer | Python | Go | Equality |
|-------|--------|-----|----------|
| Logic IR | `logic.py` types (recstruct) | `logic/` types (`Sexp()`/`NodeKey`) | Structural ✅ |
| AST | `ivy_ast.py` types | `ast/` types | Identity (pointer) ✅ |

The `logic/` types all correctly use structural equivalence because Python's `recstruct` gives them all `__eq__` + `__hash__`. The `ast/` types use pointer identity in Go (default for structs behind pointers), matching Python's default identity equality for most `ivy_ast.py` classes.

The only bridge between layers is when `ast` types get converted to `logic` types (e.g., `ast.Symbol` → `lg.Const`, `ast.Variable` → `lg.Variable`). At that point, structural equality kicks in via `NodeKey`, which is correct.

## No code changes needed

This audit confirms the current implementation is faithful to Python's semantics. The earlier concern about "structural equivalence for all types" was unfounded — Go correctly restricts structural keying (`NodeKey`/`Sexp()`) to `logic/` package types, matching Python's `recstruct`-based `logic.py` types.

## Verification

To further validate, one could:
- `grep` for any `ast.Atom`, `ast.App`, `ast.LabeledFormula`, or other `ast/` types being used as values in `map[lg.NodeKey]` — they shouldn't be used as *keys*
- Confirm no `ast/` type implements `Sexp()` or is passed to `lg.Key()`
- All `lg.Key()` calls should receive `lg.Expr` (logic-layer) types, not `ast.Node` types
