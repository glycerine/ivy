# Audit: Go `ast/rewrite.go` Conformance to Python `ivy_ast.py`

**Created:** 2026-03-25T17:45

## Summary

Compared all 17 branches of Python's `ast_rewrite` (ivy_ast.py:1697-1755) against Go's `AstRewrite` (ast/rewrite.go:507-776), plus all rewriter classes and helper functions.

---

## BUG 1 (Medium): Extra `PrefixStr` on Variable/Atom sort names

**Go** `ast/rewrite.go:531-541` (Variable) and `ast/rewrite.go:586-597` (Atom):
```go
// Variable case (line 538-539):
if sp, ok := rewrite.(*AstRewriteSubstPrefix); ok && newSort != "" {
    newSortStr = sp.PrefixStr(newSortStr, false)  // EXTRA — not in Python
}

// Atom case (line 591-592):
if sp, ok := rewrite.(*AstRewriteSubstPrefix); ok && newSortStr != "" {
    newSortStr = sp.PrefixStr(newSortStr, false)  // EXTRA — not in Python
}
```

**Python** (ivy_ast.py:1704-1705 and 1714):
```python
# Variable:
return x.resort(rewrite_sort(rewrite, x.sort))  # Just rewrite_sort, no extra prefix_str

# Atom:
atom.sort = rewrite_sort(rewrite, x.sort)  # Just rewrite_sort, no extra prefix_str
```

**Impact:** `RewriteSort` already applies `PrefixStr` internally (via `RewriteAtom` → `ParseName.Apply(PrefixStr)`). The extra `PrefixStr` call could double-prefix sort names, especially when `RewriteSort` takes the early-return `BaseNameDiffers` path (subscript substitution changed the base name — Python returns the substituted name as-is, Go then tries to prefix it again).

**Fix:** Remove lines 538-539 and 591-592 (the extra `sp.PrefixStr` calls).

---

## BUG 2 (Low): Missing TypeDef parameter validation

**Go** `ast/rewrite.go:732-735`:
```go
case *TypeDef:
    newArgs := AstRewriteSlice(n.Args(), rewrite)
    return n.Clone(newArgs)
    // Missing: if res.args[0].args: raise error
```

**Python** (ivy_ast.py:1737-1741):
```python
if isinstance(x, TypeDef):
    res = x.clone(ast_rewrite(x.args, rewrite))
    if res.args[0].args:
        raise iu.IvyError(x, 'Types cannot have parameters: {}'.format(x.name))
    return res
```

**Impact:** If a type's name atom gets rewritten to include parameters (unlikely but possible), Python catches this and raises an error. Go silently allows it.

**Fix:** Add parameter validation after Clone.

---

## BUG 3 (Low): `SubstituteConstantsAst2` uses `CopyAttributesAstRef` instead of `CopyAttributesAst`

**Go** `ast/rewrite.go:907-908`:
```go
res := node.Clone(newArgs)
CopyAttributesAstRef(node, res)  // Adds lineno reference info
```

**Python** (ivy_ast.py:1846-1847):
```python
res = ast.clone(new_args)
copy_attributes_ast(ast, res)  # Does NOT add reference info
```

**Impact:** Go adds reference lineno info where Python doesn't. Could cause lineno divergences in deeply nested module instantiation.

**Fix:** Change to `CopyAttributesAst(node, res)` (no ref version). Verify `CopyAttributesAst` exists; if not, create it.

---

## BUG 4 (Low): `SubstituteConstantsAst` routed through `AstRewrite` instead of direct recursion

**Go** `ast/rewrite.go:850-853`:
```go
func SubstituteConstantsAst(node Node, subs map[string]Node) Node {
    rw := NewAstRewriteSubstConstants(subs)
    return AstRewrite(node, rw)  // Full AstRewrite type switch
}
```

**Python** (ivy_ast.py:1811-1824):
```python
def substitute_constants_ast(ast, subs):
    if (isinstance(ast, Atom) or isinstance(ast, App)) and not ast.args:
        return subs.get(ast.rep, ast)
    else:
        if isinstance(ast, str):
            return ast
        new_args = [substitute_constants_ast(x, subs) for x in ast.args]
        res = ast.clone(new_args)
        copy_attributes_ast(ast, res)  # No ref
        return res
```

**Impact:** Go's version runs the full `AstRewrite` type switch, which means:
- Variables get `RewriteSort` applied (Python's version just clones them)
- LabeledFormulas get special label handling (Python's version just clones)
- Sort names could be accidentally substituted if they match constant names

In practice, `AstRewriteSubstConstants.RewriteName` returns name unchanged and `RewriteAtom` only matches argless atoms, so most branches are no-ops. But `RewriteSort` calls `RewriteAtom(Atom(sort))`, which COULD match if a sort name equals a constant name in `subs`.

**Fix:** Implement as direct recursive function matching Python, or verify that the routing through `AstRewrite` produces identical results for all cases. Low priority since it likely hasn't caused visible issues.

---

## CONFIRMED CORRECT (no issues found)

| Go Case | Python Branch | Status |
|---------|---------------|--------|
| `isTacticType` check | `isinstance(x, Tactic)` | ✓ Fixed (all subtypes now matched) |
| `case *Variable` | `isinstance(x, Variable)` | ✓ (except Bug 1) |
| `case *Symbol` | N/A (Go-specific) | ✓ Correctly simulates nullary Atom |
| `case *Atom` | `isinstance(x, Atom)` | ✓ (except Bug 1 sort) |
| `case *App` | `isinstance(x, App)` | ✓ NamedBinder + BaseNameDiffers correct |
| `case *Literal` | `isinstance(x, Literal)` | ✓ |
| `case *Forall` | `isinstance(x, Quantifier)` | ✓ |
| `case *Exists` | `isinstance(x, Quantifier)` | ✓ |
| `case *NamedBinder` | `isinstance(x, NamedBinder)` | ✓ |
| `case *LabeledFormula` | `isinstance(x, LabeledFormula)` | ✓ Local flag correct |
| `case *NativeDef` | `isinstance(x, NativeDef)` | ✓ Same as LabeledFormula |
| `case *SchemaBody` | `isinstance(x, SchemaBody)` | ✓ Local flag correct |
| `case *DebugItem` | `isinstance(x, DebugItem)` | ✓ |
| `default` (AstRewritable) | `hasattr(x, 'rewrite')` | ✓ |
| `default` (Args fallback) | `hasattr(x, 'args')` | ✓ |
| `AstRewriteSubstPrefix` | class AstRewriteSubstPrefix | ✓ All methods match |
| `AstRewriteSubstConstants` | class AstRewriteSubstConstants | ✓ |
| `AstRewriteSubstConstantsParams` | class AstRewriteSubstConstantsParams | ✓ |
| `AstRewritePostfix` | class AstRewritePostfix | ✓ |
| `AstRewriteAddParams` | class AstRewriteAddParams | ✓ |
| `RewriteSort` | `rewrite_sort()` | ✓ |
| `SubstPrefixAtomsAst` | `subst_prefix_atoms_ast()` | ✓ |
| `SubstituteConstantsAst2` | `substitute_constants_ast2()` | ✓ (except Bug 3 ref) |

---

## Recommended Fix Priority

1. **Bug 1** (Extra PrefixStr) — Fix first. Could cause sort name divergences in module instantiation. Remove lines 538-539 and 591-592 from `ast/rewrite.go`.
2. **Bug 2** (TypeDef validation) — Low priority. Add the parameter check.
3. **Bug 3** (CopyAttributesAstRef) — Low priority. Change to non-ref version.
4. **Bug 4** (SubstituteConstantsAst routing) — Lowest priority. Unlikely to cause issues in practice.

---

## Files to Modify

| File | Change |
|------|--------|
| `ast/rewrite.go:538-539` | Remove extra `sp.PrefixStr` on Variable sort |
| `ast/rewrite.go:591-592` | Remove extra `sp.PrefixStr` on Atom sort |
| `ast/rewrite.go:732-735` | Add TypeDef parameter validation |
| `ast/rewrite.go:908` | Change `CopyAttributesAstRef` to `CopyAttributesAst` |
